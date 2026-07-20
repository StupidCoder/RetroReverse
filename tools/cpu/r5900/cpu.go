package r5900

// An Emotion Engine execution core — the executable counterpart of the decoder,
// mirroring r4300.CPU / mips.CPU. It is the main CPU of the PlayStation 2.
//
// One hardware pipeline hazard is modelled, because compiled code depends on it:
// the branch delay slot. The instruction after a branch or jump always executes
// before control transfers, so PC/nextPC are kept as a pair — Step advances PC to
// the delay slot before the branch runs, and the branch only chooses nextPC. The
// branch-likely forms additionally annul (skip) the slot when not taken.
//
// There is no load delay slot: the core interlocks, so a loaded value is visible
// to the very next instruction and a single register file suffices.
//
// Memory goes through the Bus, so the caller (tools/platform/ps2) supplies the
// machine model: 32 MiB of main RAM, the scratchpad, and the memory-mapped
// registers of the DMAC, GIF, VIF, GS and the rest. Bus takes *physical*
// addresses; virtual-to-physical translation (the segment map and the TLB)
// happens here, in cop0.go.

import "fmt"

// Bus is the physical address space the CPU drives, little-endian.
//
// Read32/Write32 are not a convenience: the CPU fetches an instruction word for
// every instruction it runs, and composing each fetch from four interface calls
// (as tools/cpu/mips does) costs more than the interpreter itself. The machine
// model serves RAM and the scratchpad from a backing slice through these, falling
// through to its address-range switch only for memory-mapped registers.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
	Read32(addr uint32) uint32
	Write32(addr uint32, v uint32)
}

// Fetcher lets a machine model distinguish an instruction fetch from a data read.
// A Bus that implements it has Fetch32 called for every instruction word, and
// Read32 only for loads — without which a "who reads this address" watch is
// drowned by the fetch of every instruction that runs inside the window.
type Fetcher interface {
	Fetch32(addr uint32) uint32
}

// Coprocessor2 is VU0 as the EE sees it: 32 vector registers of four floats, 16
// integer registers, and the macro-mode instruction space reached by the COP2
// opcode. tools/cpu/vu implements it. A nil COP2 degrades gracefully — reads
// return zero, writes and operations are dropped — so a boot trace does not stall
// before the vector unit exists.
type Coprocessor2 interface {
	// ReadVF/WriteVF address one 32-bit field of a vector register: reg 0..31,
	// field 0..3 (x, y, z, w). qmfc2/qmtc2 move all four.
	ReadVF(reg, field uint32) uint32
	WriteVF(reg, field, v uint32)
	// ReadCtrl/WriteCtrl reach the integer and control registers (cfc2/ctc2).
	ReadCtrl(reg uint32) uint32
	WriteCtrl(reg, v uint32)
	// Macro executes one macro-mode COP2 instruction word.
	Macro(w uint32)
}

// COP0 register indices.
const (
	cop0Index    = 0
	cop0Random   = 1
	cop0EntryLo0 = 2
	cop0EntryLo1 = 3
	cop0Context  = 4
	cop0PageMask = 5
	cop0Wired    = 6
	cop0BadVAddr = 8
	cop0Count    = 9
	cop0EntryHi  = 10
	cop0Compare  = 11
	cop0Status   = 12
	cop0Cause    = 13
	cop0EPC      = 14
	cop0PRId     = 15
	cop0Config   = 16
	cop0ErrorEPC = 30
)

// Status register bits.
const (
	statusIE  = 1 << 0  // global interrupt enable
	statusEXL = 1 << 1  // exception level: an exception is being handled
	statusERL = 1 << 2  // error level
	statusBEV = 1 << 22 // bootstrap exception vectors (ROM, not RAM)
	statusCU0 = 1 << 28 // coprocessor 0 usable
	statusCU1 = 1 << 29 // coprocessor 1 usable (the FPU)
	statusCU2 = 1 << 30 // coprocessor 2 usable (VU0 macro mode)

	// EIE is the EE's second, independent interrupt enable, set and cleared by the
	// ei/di instructions. An interrupt is delivered only when IE *and* EIE are set,
	// which is how the kernel brackets a critical section without touching IE.
	statusEIE = 1 << 16
)

// Cause register bits. The EE aggregates its peripheral interrupts onto two
// lines: INT0 (the INTC — VBlank, the GS, the timers) and INT1 (the DMAC).
const (
	causeBD  = 1 << 31 // exception taken in a branch delay slot
	causeIP2 = 1 << 10 // INT0: the interrupt controller
	causeIP3 = 1 << 11 // INT1: the DMA controller
	causeIP7 = 1 << 15 // the CPU's own Count/Compare timer
)

// Exception codes (Cause bits 2..6).
const (
	excInt  = 0x00 // interrupt
	excMod  = 0x01 // TLB modification (store to a clean page)
	excTLBL = 0x02 // TLB miss, load or fetch
	excTLBS = 0x03 // TLB miss, store
	excAdEL = 0x04 // address error, load/fetch
	excAdES = 0x05 // address error, store
	excSys  = 0x08 // syscall
	excBp   = 0x09 // breakpoint
	excRI   = 0x0A // reserved instruction
	excCpU  = 0x0B // coprocessor unusable
	excOv   = 0x0C // arithmetic overflow
	excTrap = 0x0D // trap instruction
)

// Exception vector bases, selected by the Status BEV bit.
const (
	vecRAM = 0x80000000
	vecROM = 0xBFC00200
)

// Exported names for the COP0 registers and Status/Cause bits a machine model
// needs to reach: the boot code sets Status directly, and the run loop drives the
// interrupt-pending bits.
const (
	Cop0Count    = cop0Count
	Cop0Compare  = cop0Compare
	Cop0Status   = cop0Status
	Cop0Cause    = cop0Cause
	Cop0EPC      = cop0EPC
	Cop0BadVAddr = cop0BadVAddr

	StatusIE  = statusIE
	StatusEXL = statusEXL
	StatusERL = statusERL
	StatusBEV = statusBEV
	StatusEIE = statusEIE
	StatusCU0 = statusCU0
	StatusCU1 = statusCU1
	StatusCU2 = statusCU2

	CauseIP2 = causeIP2
	CauseIP3 = causeIP3
	CauseIP7 = causeIP7
)

// CPU is the Emotion Engine programmer's-model core.
type CPU struct {
	R [32]Quad // general registers, 128-bit; R[0] is hardwired to zero

	// The EE has two multiply-accumulate units. The ordinary mult/div write the
	// first (HI/LO); the MMI "1" forms (mult1, div1, madd1 …) write the second
	// (HI1/LO1), so a compiler can keep two chains in flight.
	HI, LO   uint64
	HI1, LO1 uint64

	// SA is the shift-amount register, written by mtsa/mtsab/mtsah and read by
	// qfsrv — the funnel shift that extracts an unaligned quadword.
	SA uint32

	PC     uint64 // address of the instruction to fetch next
	nextPC uint64 // address after PC (delay-slot machinery)

	COP0 [32]uint64 // system-control registers (see cop0.go)
	TLB  [TLBSize]TLBEntry

	// COP1: 32 single-precision registers held as bit patterns, the accumulator the
	// multiply-add family targets, and the control/status word. See fpu.go.
	FPR   [32]uint32
	ACC   uint32
	FCR31 uint32

	// COP2 is VU0 in macro mode. Optional; nil is tolerated.
	COP2 Coprocessor2

	LLBit bool // set by ll/lld, cleared by an intervening store or eret; sc/scd test it

	Halted     bool
	HaltReason string
	Steps      uint64

	// Syscall is called for the syscall instruction. Returning true means the host
	// handled it and the core should continue; false takes the normal exception.
	// The machine model uses this to HLE the EE kernel, which lives in a BIOS ROM
	// we do not have.
	Syscall func(c *CPU) bool

	bus   Bus
	fetch func(uint32) uint32 // bus.Fetch32 when the Bus is a Fetcher, else bus.Read32

	curPC        uint64 // address of the instruction currently executing
	delaySlot    bool   // the current instruction is in a branch delay slot
	pendingDelay bool   // the next instruction will be a delay slot
	branchAddr   uint64 // address of the branch whose delay slot we're in

	// countFrac carries the half-rate Count increment between steps.
	countFrac uint64
}

// NewCPU makes a core over bus in the reset state.
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus, fetch: bus.Read32}
	if f, ok := bus.(Fetcher); ok {
		c.fetch = f.Fetch32
	}
	c.Reset()
	return c
}

// Reset returns the core to power-on state: PC at the reset vector in the BIOS,
// and the COP0 bits the hardware forces on reset.
//
// The PS2 never executes from that vector under this model — the BIOS ROM is not
// on the disc, so tools/platform/ps2 loads the boot ELF and starts the core at its
// entry point instead. Reset still models it, so a core used standalone (in tests)
// begins somewhere defined.
func (c *CPU) Reset() {
	c.R = [32]Quad{}
	c.HI, c.LO, c.HI1, c.LO1, c.SA = 0, 0, 0, 0, 0
	c.PC, c.nextPC = 0xFFFFFFFFBFC00000, 0xFFFFFFFFBFC00004
	c.COP0 = [32]uint64{}
	// ERL and BEV are set out of reset: the ROM vectors are in use and KSEG0 is
	// unmapped-uncached until the boot code clears ERL.
	c.COP0[cop0Status] = statusERL | statusBEV
	c.COP0[cop0PRId] = 0x00002E20 // the R5900 revision the EE reports
	c.COP0[cop0Config] = 0x00000440
	c.COP0[cop0Random] = TLBSize - 1
	c.TLB = [TLBSize]TLBEntry{}
	c.FPR = [32]uint32{}
	c.ACC, c.FCR31 = 0, 0
	c.LLBit = false
	c.delaySlot, c.pendingDelay = false, false
	c.Halted, c.HaltReason = false, ""
	c.Steps, c.countFrac = 0, 0
}

// Halt stops the core, recording why (unimplemented instruction or fatal fault).
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// SetPC forces the program counter and the following (delay) address, e.g. to
// jump to the boot ELF's entry point.
func (c *CPU) SetPC(pc uint64) {
	c.PC, c.nextPC = pc, pc+4
	c.pendingDelay = false
}

// SetReg writes the low 64 bits of a general register, leaving the high half
// alone (R0 stays hardwired to zero).
func (c *CPU) SetReg(i uint32, v uint64) { c.set(i, v) }

// Reg reads the low 64 bits of a general register — the half ordinary MIPS code
// works in.
func (c *CPU) Reg(i uint32) uint64 { return c.R[i].Lo }

// SetQuad writes a whole 128-bit register.
func (c *CPU) SetQuad(i uint32, q Quad) {
	if i != 0 {
		c.R[i] = q
	}
}

// Quad reads a whole 128-bit register.
func (c *CPU) Quad(i uint32) Quad { return c.R[i] }

// CurPC returns the address of the instruction currently executing (valid inside
// a Bus write, e.g. to attribute "who wrote this address").
func (c *CPU) CurPC() uint64 { return c.curPC }

// NextPC, InDelaySlot, PendingDelay and BranchAddr expose the branch-delay machinery,
// so an idle-loop detector can prove the core has returned to an *identical* state —
// forgetting that a branch was in flight would let two different points in a loop look
// the same. They read pure architectural state and change nothing.
func (c *CPU) NextPC() uint64      { return c.nextPC }
func (c *CPU) InDelaySlot() bool   { return c.delaySlot }
func (c *CPU) PendingDelay() bool  { return c.pendingDelay }
func (c *CPU) BranchAddr() uint64  { return c.branchAddr }

// set writes the low half of register i, preserving the high half. This is what
// every ordinary MIPS instruction does on this core: the upper 64 bits belong to
// the MMI world and are not disturbed by a daddu.
func (c *CPU) set(i uint32, v uint64) {
	if i != 0 {
		c.R[i].Lo = v
	}
}

// setQ writes a whole 128-bit register (the MMI and lq path).
func (c *CPU) setQ(i uint32, lo, hi uint64) {
	if i != 0 {
		c.R[i].Lo, c.R[i].Hi = lo, hi
	}
}

// sext32 sign-extends a 32-bit result into a 64-bit register, which is what every
// 32-bit MIPS operation does. Getting this wrong is invisible until a value is
// used as an address or compared as a 64-bit quantity.
func sext32(v uint32) uint64 { return uint64(int64(int32(v))) }

// --- memory helpers (little-endian, physical addresses) ----------------------

func (c *CPU) read8(a uint32) uint32     { return uint32(c.bus.Read(a)) }
func (c *CPU) write8(a uint32, v uint32) { c.bus.Write(a, byte(v)) }

func (c *CPU) read16(a uint32) uint32 {
	return uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8
}
func (c *CPU) write16(a uint32, v uint32) {
	c.bus.Write(a, byte(v))
	c.bus.Write(a+1, byte(v>>8))
}

func (c *CPU) read32(a uint32) uint32     { return c.bus.Read32(a) }
func (c *CPU) write32(a uint32, v uint32) { c.bus.Write32(a, v) }

// A doubleword is little-endian: the low word is at the lower address.
func (c *CPU) read64(a uint32) uint64 {
	return uint64(c.bus.Read32(a)) | uint64(c.bus.Read32(a+4))<<32
}
func (c *CPU) write64(a uint32, v uint64) {
	c.bus.Write32(a, uint32(v))
	c.bus.Write32(a+4, uint32(v>>32))
}

func (c *CPU) read128(a uint32) Quad {
	return Quad{Lo: c.read64(a), Hi: c.read64(a + 8)}
}
func (c *CPU) write128(a uint32, q Quad) {
	c.write64(a, q.Lo)
	c.write64(a+8, q.Hi)
}

// --- exceptions -------------------------------------------------------------

// Exception enters the exception handler with the given cause code, saving the
// return address in EPC (pointing at the branch if the fault was in a delay slot,
// with the Cause BD bit set).
func (c *CPU) Exception(code uint32) {
	c.exceptionAt(code, false)
}

func (c *CPU) exceptionAt(code uint32, tlbRefill bool) {
	sr := c.COP0[cop0Status]

	// EPC and the BD bit are only written when an exception is not already being
	// handled; otherwise the original EPC must survive.
	if sr&statusEXL == 0 {
		epc := c.curPC
		cause := c.COP0[cop0Cause] &^ 0xFFFFFFFF00000000
		if c.delaySlot {
			epc = c.branchAddr
			cause |= causeBD
		} else {
			cause &^= causeBD
		}
		c.COP0[cop0EPC] = epc
		// Preserve the interrupt-pending bits; replace ExcCode and BD.
		c.COP0[cop0Cause] = (cause &^ 0x0000007C) | uint64(code<<2)
	} else {
		c.COP0[cop0Cause] = (c.COP0[cop0Cause] &^ 0x0000007C) | uint64(code<<2)
	}

	// The coprocessor field names the unit that faulted, and is written on every
	// exception — zero unless a coprocessor was involved. Leaving the previous
	// exception's value in it tells a handler the wrong story.
	c.COP0[cop0Cause] &^= 0x30000000

	base := uint64(vecRAM)
	if sr&statusBEV != 0 {
		base = vecROM
	}
	offset := uint64(0x180)
	if tlbRefill && sr&statusEXL == 0 {
		offset = 0x000
	}

	c.COP0[cop0Status] = sr | statusEXL
	c.LLBit = false
	c.SetPC(sext64(base + offset))
	c.delaySlot = false
}

// sext64 sign-extends a 32-bit virtual address into the 64-bit PC, which is how
// the CPU sees KSEG addresses in 32-bit mode.
func sext64(v uint64) uint64 { return uint64(int64(int32(uint32(v)))) }

// eret returns from an exception: to ErrorEPC when ERL is set, else to EPC,
// clearing the level bit. It also breaks any outstanding load-linked.
func (c *CPU) eret() {
	sr := c.COP0[cop0Status]
	if sr&statusERL != 0 {
		c.SetPC(c.COP0[cop0ErrorEPC])
		c.COP0[cop0Status] = sr &^ statusERL
	} else {
		c.SetPC(c.COP0[cop0EPC])
		c.COP0[cop0Status] = sr &^ statusEXL
	}
	c.LLBit = false
	// ERET has no delay slot: the next instruction fetched is at the target.
	c.pendingDelay = false
}

// coprocessorUnusable raises the exception a program gets for touching a
// coprocessor its Status register has not enabled. Which coprocessor is recorded
// in the Cause register's CE field, and a handler needs it to know what to do.
func (c *CPU) coprocessorUnusable(unit uint32) {
	c.Exception(excCpU)
	c.COP0[cop0Cause] = (c.COP0[cop0Cause] &^ 0x30000000) | uint64(unit)<<28
}

// addrError raises an address-error exception for a misaligned access.
func (c *CPU) addrError(code uint32, vaddr uint64) {
	c.setFaultAddress(vaddr)
	c.Exception(code)
}

// --- interrupts -------------------------------------------------------------

// Interrupt updates the two external interrupt lines — INT0 (the interrupt
// controller) and INT1 (the DMA controller) — from the machine's masked status
// and, if interrupts are enabled and unmasked, takes an interrupt exception. It
// reports whether one was delivered.
//
// Call it between instructions (before Step): the address saved in EPC is the
// instruction about to run, which resumes cleanly after the handler.
func (c *CPU) Interrupt(int0, int1 bool) bool {
	if int0 {
		c.COP0[cop0Cause] |= causeIP2
	} else {
		c.COP0[cop0Cause] &^= causeIP2
	}
	if int1 {
		c.COP0[cop0Cause] |= causeIP3
	} else {
		c.COP0[cop0Cause] &^= causeIP3
	}
	return c.checkInterrupt()
}

// checkInterrupt delivers a pending, enabled, unmasked interrupt.
//
// The EE gates interrupts on *two* enables, not one: the usual Status IE and the
// EE-specific EIE that the ei/di instructions drive. Both must be set. A core that
// honours only IE takes interrupts inside every kernel critical section.
func (c *CPU) checkInterrupt() bool {
	sr := c.COP0[cop0Status]
	if sr&statusIE == 0 || sr&statusEIE == 0 || sr&(statusEXL|statusERL) != 0 {
		return false
	}
	// Cause IP (bits 8..15) against Status IM (bits 8..15).
	if c.COP0[cop0Cause]&sr&0x0000FF00 == 0 {
		return false
	}
	if c.pendingDelay { // don't split a branch from its delay slot
		return false
	}
	// EPC must be the next instruction to fetch, not the last one executed, so the
	// handler's eret resumes exactly where control was preempted.
	c.curPC = c.PC
	c.delaySlot = false
	c.Exception(excInt)
	return true
}

// tickCount advances the Count register at half the CPU clock and raises the timer
// interrupt (IP7) when it reaches Compare. Count is 32-bit and wraps; the
// comparison is on equality, so a wrap between two steps must not skip it — hence
// the increment-then-test on the exact value.
func (c *CPU) tickCount() {
	c.countFrac++
	if c.countFrac&1 != 0 {
		return
	}
	count := uint32(c.COP0[cop0Count]) + 1
	c.COP0[cop0Count] = uint64(count)
	if count == uint32(c.COP0[cop0Compare]) {
		c.COP0[cop0Cause] |= causeIP7
	}
}

// SkipInstructions advances the core's bookkeeping as if n instructions had run,
// WITHOUT running them. It is the primitive an idle-loop fast-forward is built on
// (tools/platform/ps2/idle.go): a loop proven to return to the same state having
// stored nothing can be fast-forwarded a whole number of its periods, and this
// carries the two pieces of core state that advance with the clock rather than with
// the program — the retired-instruction count, and the Count/Compare timer.
//
// It replicates n calls to tickCount in closed form: Count ticks at half the clock,
// so it advances by the number of even values the fraction counter passes through,
// and IP7 is raised if that run of increments steps across Compare — exactly the
// equality tickCount tests, wrap included. The CALLER is responsible for not skipping
// across a Compare that is actually deliverable (see idleDeadline): this sets the
// pending bit as the hardware would, but nothing here delivers the interrupt.
func (c *CPU) SkipInstructions(n uint64) {
	if n == 0 {
		return
	}
	f0 := c.countFrac
	c.countFrac = f0 + n
	// Even fraction values in (f0, f0+n] are the increments Count takes.
	inc := (f0+n)/2 - f0/2
	if inc > 0 {
		count := uint32(c.COP0[cop0Count])
		comp := uint32(c.COP0[cop0Compare])
		// Compare is hit if it lies count+1..count+inc going forward (mod 2^32). A
		// skip long enough to wrap the whole counter necessarily passes it.
		if inc >= 1<<32 {
			c.COP0[cop0Cause] |= causeIP7
		} else if d := uint64(comp - count); d != 0 && d <= inc {
			c.COP0[cop0Cause] |= causeIP7
		}
		c.COP0[cop0Count] = uint64(count + uint32(inc))
	}
	c.Steps += n
}

// Count and Compare read the Count/Compare timer, so an idle skip can size how far it
// may go before the timer would fire.
func (c *CPU) Count() uint32   { return uint32(c.COP0[cop0Count]) }
func (c *CPU) Compare() uint32 { return uint32(c.COP0[cop0Compare]) }

// CountFrac reports the half-rate fraction carry, which fixes the phase of the next
// Count increment: with the carry odd, the very next step increments Count.
func (c *CPU) CountFrac() uint64 { return c.countFrac }

// TimerIRQDeliverable reports whether a Count/Compare (IP7) interrupt would be taken
// right now if it were pending — both enables on, no exception in progress, IM7 set.
// When it is true the timer is a hard deadline for an idle skip; when false, reaching
// Compare only sets a status bit nothing acts on, which SkipInstructions reproduces.
func (c *CPU) TimerIRQDeliverable() bool {
	sr := c.COP0[cop0Status]
	return sr&statusIE != 0 && sr&statusEIE != 0 && sr&(statusEXL|statusERL) == 0 &&
		sr&(1<<15) != 0
}

// InterruptDeliverable reports whether any enabled, unmasked interrupt is pending and
// would be taken at the next instruction boundary. An idle skip refuses to start when
// one is, so the fast-forward never steps over an interrupt the run loop would deliver.
func (c *CPU) InterruptDeliverable() bool {
	sr := c.COP0[cop0Status]
	if sr&statusIE == 0 || sr&statusEIE == 0 || sr&(statusEXL|statusERL) != 0 {
		return false
	}
	return c.COP0[cop0Cause]&sr&0x0000FF00 != 0
}
