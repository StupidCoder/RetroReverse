package allegrex

// A MIPS R3000A integer execution core — the executable counterpart of the
// decoder, mirroring arm.CPU / x86.CPU. It is the CPU of the PlayStation.
//
// Two hardware pipeline hazards are modelled because game code (and the
// SingleStepTests vectors) depend on them:
//
//   - Branch delay slot: the instruction after a branch/jump always executes
//     before control transfers. Modelled with a PC/nextPC pair — Step advances
//     PC to the delay slot before the branch runs, and the branch only chooses
//     nextPC, so the slot executes either way.
//   - Load delay slot: a value loaded from memory is not visible to the very
//     next instruction. Modelled with a two-register-file scheme (reads see R,
//     writes go to out, a pending load lands in out at the start of the *next*
//     instruction), which reproduces the R3000 rule that an ALU write in the
//     load's delay slot beats the load.
//
// Memory goes through the Bus, so the caller (tools/psx) supplies the machine
// model: 2 MiB RAM, the scratchpad, the memory-mapped I/O and the BIOS. COP0
// (system control / exceptions) is built in; COP2 (the GTE) is reached through
// an optional Cop2 hook, exactly as arm.CPU routes CP15 through Coproc.
// Unimplemented encodings call Halt with the offending word, so gaps are
// explicit rather than silently wrong.

import "fmt"

// Bus is the flat 32-bit address space the CPU drives (byte-addressed,
// little-endian words composed by the core). The machine model decodes the PSX
// memory map, the segment mirrors (KUSEG/KSEG0/KSEG1) and the I/O.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
}

// COP0 register indices and exception codes.
const (
	cop0BadVaddr = 8
	cop0Status   = 12 // SR
	cop0Cause    = 13
	cop0EPC      = 14
	cop0PRId     = 15

	excInt  = 0x00 // interrupt
	excAdEL = 0x04 // address error, load/fetch
	excAdES = 0x05 // address error, store
	excSys  = 0x08 // syscall
	excBp   = 0x09 // breakpoint
	excRI   = 0x0A // reserved instruction
	excCpU  = 0x0B // coprocessor unusable
	excOv   = 0x0C // arithmetic overflow

	// Exception vectors, selected by the SR BEV bit.
	vecRAM = 0x80000080
	vecROM = 0xBFC00180
)

type loadSlot struct {
	reg uint32 // 0 = no pending load
	val uint32
}

// CPU is the MIPS R3000 programmer's-model core.
type CPU struct {
	R      [32]uint32 // input register file: what instruction reads see (R[0]=0)
	out    [32]uint32 // output register file: where writes go this instruction
	HI, LO uint32     // multiply/divide result registers
	PC     uint32     // address of the instruction to fetch next
	nextPC uint32     // address after PC (delay-slot machinery)
	COP0   [32]uint32 // system-control coprocessor registers

	// COP1 FPU (single-precision): 32 registers stored as float32 bit patterns,
	// plus the one condition bit of FCSR the bc1t/bc1f branches test.
	F   [32]uint32
	FCC bool // FPU condition code (FCSR bit 23)

	// COP2 VFPU: 128 32-bit registers (float bit patterns) addressed as
	// matrices/rows/columns, plus the 16 control registers reachable by mtvc/mfvc
	// (the vpfxs/vpfxt/vpfxd operand-prefix latches, the vcmp condition-code
	// register, and the vrnd state). See vfpu.go.
	V        [128]uint32
	VfpuCtrl [16]uint32

	// OnVFPU optionally handles an unimplemented VFPU compute op instead of halting
	// (a soft no-op probe, or a census of the ops a program reaches).
	OnVFPU func(w, op uint32)

	// Syscall optionally handles the `syscall` instruction (return true if
	// serviced, so the core does not raise the Sys exception). The PSP kernel-HLE
	// resolves each import stub's `syscall` code to a Go handler; code is the
	// instruction's 20-bit code field (bits 6..25).
	Syscall func(c *CPU, code uint32) bool

	Halted     bool
	HaltReason string
	Steps      uint64

	bus Bus

	curPC        uint32   // address of the instruction currently executing
	ld           loadSlot // load pending from the previous instruction
	delaySlot    bool     // the current instruction is in a branch delay slot
	pendingDelay bool     // the next instruction will be a delay slot
	branchAddr   uint32   // address of the branch whose delay slot we're in
	nullifyNext  bool     // a not-taken likely branch nullifies the next instruction
}

// NewCPU makes a core over bus in the reset state (PC at the BIOS ROM vector).
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset returns the core to power-on state: PC at the BIOS reset vector
// (0xBFC00000), all registers clear.
func (c *CPU) Reset() {
	c.R = [32]uint32{}
	c.out = [32]uint32{}
	c.HI, c.LO = 0, 0
	c.PC, c.nextPC = 0xBFC00000, 0xBFC00004
	c.COP0 = [32]uint32{}
	c.COP0[cop0PRId] = 0x00005E00 // Allegrex revision id
	c.VfpuCtrl = [16]uint32{}
	c.VfpuCtrl[vfpuCtlSPfx] = pfxIdentity
	c.VfpuCtrl[vfpuCtlTPfx] = pfxIdentity
	c.ld = loadSlot{}
	c.delaySlot, c.pendingDelay = false, false
	c.Halted, c.HaltReason = false, ""
}

// Halt stops the core, recording why (unimplemented instruction or fatal fault).
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// SetPC forces the program counter and the following (delay) address, e.g. to
// jump to an EXE entry point.
func (c *CPU) SetPC(pc uint32) {
	c.PC, c.nextPC = pc, pc+4
	c.pendingDelay = false
}

// SetReg writes a general register in both register files, so it survives the
// end-of-step commit. Use it to seed state (sp/gp/args) before running.
func (c *CPU) SetReg(i, v uint32) {
	if i != 0 {
		c.R[i], c.out[i] = v, v
	}
}

// Reg reads a general register.
func (c *CPU) Reg(i uint32) uint32 { return c.R[i] }

// CurPC returns the address of the instruction currently executing (valid inside
// a Bus write, e.g. to attribute "who wrote this address").
func (c *CPU) CurPC() uint32 { return c.curPC }

// --- register access -------------------------------------------------------

// set writes register i (R0 stays hardwired to zero) into the output file.
func (c *CPU) set(i, v uint32) {
	if i != 0 {
		c.out[i] = v
	}
}

// --- memory helpers (little-endian) ----------------------------------------

func (c *CPU) read8(a uint32) uint32  { return uint32(c.bus.Read(a)) }
func (c *CPU) write8(a, v uint32)     { c.bus.Write(a, byte(v)) }
func (c *CPU) read16(a uint32) uint32 { return uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8 }
func (c *CPU) write16(a, v uint32) {
	c.bus.Write(a, byte(v))
	c.bus.Write(a+1, byte(v>>8))
}
func (c *CPU) read32(a uint32) uint32 {
	return uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8 | uint32(c.bus.Read(a+2))<<16 | uint32(c.bus.Read(a+3))<<24
}
func (c *CPU) write32(a, v uint32) {
	c.bus.Write(a, byte(v))
	c.bus.Write(a+1, byte(v>>8))
	c.bus.Write(a+2, byte(v>>16))
	c.bus.Write(a+3, byte(v>>24))
}

// --- exceptions ------------------------------------------------------------

// Exception enters the general exception handler with the given cause code,
// saving the return address in EPC (pointing at the branch if the fault was in a
// delay slot, with the CAUSE BD bit set) and shifting the SR interrupt/kernel
// stack. It vectors to RAM (0x80000080) or ROM (0xBFC00180) per the SR BEV bit.
func (c *CPU) Exception(code uint32) {
	epc := c.curPC
	cause := code << 2
	if c.delaySlot {
		epc = c.branchAddr
		cause |= 1 << 31 // BD: exception in branch delay slot
	}
	c.COP0[cop0EPC] = epc
	// Preserve the interrupt-pending bits (8..15); replace ExcCode and BD.
	c.COP0[cop0Cause] = (c.COP0[cop0Cause] & 0x0000FF00) | (cause &^ 0x0000FF00)
	// Push the interrupt-enable / kernel-user stack (low 6 bits) left by two.
	sr := c.COP0[cop0Status]
	c.COP0[cop0Status] = (sr &^ 0x3F) | ((sr << 2) & 0x3F)

	target := uint32(vecRAM)
	if sr&(1<<22) != 0 {
		target = vecROM
	}
	c.SetPC(target)
}

// rfe pops the SR interrupt/kernel stack (the tail of an exception handler).
func (c *CPU) rfe() {
	sr := c.COP0[cop0Status]
	c.COP0[cop0Status] = (sr &^ 0x0F) | ((sr >> 2) & 0x0F)
}

// Interrupt updates the external interrupt line (CAUSE IP2) from the machine's
// masked IRQ status and, if interrupts are enabled and unmasked, takes an
// interrupt exception. It returns whether an interrupt was delivered. Call it
// between instructions (before Step): the return address saved in EPC is the
// instruction about to run (c.PC), which resumes cleanly after the handler.
func (c *CPU) Interrupt(pending bool) bool {
	if pending {
		c.COP0[cop0Cause] |= 1 << 10
	} else {
		c.COP0[cop0Cause] &^= 1 << 10
	}
	sr := c.COP0[cop0Status]
	if sr&1 == 0 { // IEc: interrupts disabled
		return false
	}
	if c.COP0[cop0Cause]&sr&0x0000FF00 == 0 { // masked
		return false
	}
	if c.pendingDelay { // don't split a branch from its delay slot
		return false
	}
	// Retire a load still in its delay slot before vectoring: the pipeline would
	// otherwise land the loaded value in a register during the handler's first
	// instruction, clobbering whatever the handler set up (e.g. a fresh $ra).
	if c.ld.reg != 0 {
		c.R[c.ld.reg] = c.ld.val
		c.out[c.ld.reg] = c.ld.val
		c.ld = loadSlot{}
	}
	// EPC must be the next instruction to fetch, not the last one executed, so the
	// handler's rfe resumes exactly where control was preempted.
	c.curPC = c.PC
	c.delaySlot = false // an async interrupt is not in a delay slot
	c.Exception(excInt)
	return true
}

// addrError raises an address-error exception for a misaligned access.
func (c *CPU) addrError(code, addr uint32) {
	c.COP0[cop0BadVaddr] = addr
	c.Exception(code)
}
