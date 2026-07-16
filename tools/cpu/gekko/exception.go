package gekko

// exception.go takes an exception.
//
// PowerPC's mechanism is small and it is worth stating exactly, because a GameCube game
// installs its own handlers and relies on every part of it: the processor saves the
// address to resume at in SRR0 and the machine state in SRR1, clears the bits of MSR
// that would let another exception in, and jumps to a fixed vector. `rfi` reverses it.
// There is no kernel here — those handlers are the game's own code, compiled into its
// executable — so this path is not an emulator convenience, it is a load-bearing part of
// how the program runs.
//
// MSR[IP] selects where the vectors are. Out of reset it is set, and the vectors are at
// 0xFFF00000, which is where the console's boot ROM lives. A GameCube program clears it
// early and moves them to 0x00000000, which is the bottom of main memory — and that is
// why the first three kilobytes of a GameCube's RAM are exception handlers.

// The exception vectors, as offsets from the vector base.
const (
	VecReset        = 0x0100
	VecMachineCheck = 0x0200
	VecDSI          = 0x0300 // a data access that translation refused
	VecISI          = 0x0400 // an instruction fetch that translation refused
	VecExternal     = 0x0500 // the interrupt controller's line
	VecAlignment    = 0x0600
	VecProgram      = 0x0700 // an illegal instruction, a trap, a privilege violation
	VecFPUnavail    = 0x0800 // a floating-point instruction with MSR[FP] clear
	VecDecrementer  = 0x0900 // the decrementer counted past zero
	VecSyscall      = 0x0C00
	VecTrace        = 0x0D00
	VecPerfMon      = 0x0F00
)

// The bits of SRR1 that a program exception uses to say what kind it was.
const (
	SRR1FPEnabled  uint32 = 1 << 20 // a floating-point enabled exception
	SRR1IllegalOp  uint32 = 1 << 19
	SRR1Privileged uint32 = 1 << 18
	SRR1Trap       uint32 = 1 << 17
	SRR1NotNextPC  uint32 = 1 << 16 // SRR0 is this instruction, not the one after it
)

// vectorBase is 0xFFF00000 or 0x00000000, as MSR[IP] says.
func (c *CPU) vectorBase() uint32 {
	if c.MSR&MSRIP != 0 {
		return 0xFFF00000
	}
	return 0
}

// Exception enters a handler. resume is the address to come back to — for most
// exceptions the instruction that caused it (so that it can be retried), for a syscall
// or a decrementer the one after.
//
// The MSR bits cleared here are the whole of the processor's protection: with EE clear
// no external interrupt can preempt the handler, and with IR/DR clear the handler runs
// on physical addresses, so it does not depend on the translation that may have just
// failed.
func (c *CPU) Exception(vec uint32, resume uint32, srr1Extra uint32) {
	c.SRR0 = resume
	// SRR1 keeps the interesting half of MSR, plus whatever the exception wants to say.
	c.SRR1 = (c.MSR & 0x87C0FF73) | srr1Extra

	c.MSR &^= MSREE | MSRPR | MSRFP | MSRFE0 | MSRFE1 | MSRSE | MSRBE | MSRIR | MSRDR | MSRRI
	c.PC = c.vectorBase() + vec
}

// programException is the illegal-instruction and privilege path. SRR0 is this
// instruction, not the next one — the handler is expected to look at it.
func (c *CPU) programException(kind uint32) {
	c.Exception(VecProgram, c.PC, kind|SRR1NotNextPC)
}

// needsFPU reports whether an instruction requires the floating-point unit, and so must
// not run while MSR[FP] is clear. It is every floating-point instruction — arithmetic,
// the FPSCR moves, the paired singles, and the loads and stores, which count because they
// name a floating-point register even though they do no arithmetic.
//
// Opcode 4 is the paired-single unit and opcodes 56/57/60/61 are its quantised loads and
// stores; on a stock PowerPC those encodings are something else, but this is a Gekko.
func needsFPU(w uint32) bool {
	switch opcd(w) {
	case 4, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 59, 60, 61, 63:
		return true
	case 31:
		switch xo10(w) {
		case 535, 567, 599, 631, 663, 695, 727, 759, 983:
			return true // lfsx lfsux lfdx lfdux stfsx stfsux stfdx stfdux stfiwx
		}
	}
	return false
}

// fpUnavailable takes the floating-point unavailable exception if this instruction needs
// the FPU and MSR[FP] says it is off.
//
// This is not a corner of the architecture a GameCube program can do without: an OS with
// more than one thread switches the floating-point registers *lazily*. A thread that has
// not touched the FPU runs with MSR[FP] clear; the first floating-point instruction it
// executes traps here, and the handler saves the registers of whichever thread owned the
// FPU last, loads this thread's, sets MSR[FP] and returns to retry the instruction. The
// scheduler's own context switch never saves an FPR at all — that is the whole point, and
// it is why the game's thread-save path stores the GPRs and the GQRs and nothing else.
//
// So a core that quietly executes floating-point with MSR[FP] clear does not merely skip
// an exception. It removes the only thing that ever swaps the floating-point registers,
// and the FPRs silently become global across every thread on the machine. It looks
// harmless for as long as one thread is doing the arithmetic, which is nearly always,
// and then one preemption lands between a load and the add that consumes it and a vertex
// leaves for the horizon.
//
// SRR0 is the offending instruction, not the one after it: the handler returns to it and
// it runs again, this time with the registers it was written to see.
func (c *CPU) fpUnavailable(w, pc uint32) bool {
	if c.MSR&MSRFP != 0 || !needsFPU(w) {
		return false
	}
	c.Exception(VecFPUnavail, pc, 0)
	return true
}

// checkInterrupt takes the external interrupt or the decrementer if either is pending and
// the machine state register allows it. It runs at the top of every instruction, which is
// what makes the line level-sensitive: the interrupt controller holds it up, and the CPU
// takes the exception as soon as software permits.
func (c *CPU) checkInterrupt() bool {
	if c.MSR&MSREE == 0 {
		return false
	}
	if c.ExtInt {
		c.Exception(VecExternal, c.PC, 0)
		return true
	}
	// The decrementer fires when its top bit goes from 0 to 1 — that is, when the
	// signed count passes below zero.
	if c.DEC&0x80000000 != 0 && c.decArmed {
		c.decArmed = false
		c.Exception(VecDecrementer, c.PC, 0)
		return true
	}
	return false
}

// Interrupt raises or lowers the external interrupt line. It is level-sensitive: the
// machine holds it while any of its interrupt sources is asserted and unmasked, and the
// CPU keeps taking the exception until the machine lowers it.
func (c *CPU) Interrupt(pending bool) { c.ExtInt = pending }
