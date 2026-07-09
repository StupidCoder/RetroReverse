package r4300

// A VR4300 execution core — the executable counterpart of the decoder,
// mirroring mips.CPU / arm.CPU. It is the CPU of the Nintendo 64.
//
// One hardware pipeline hazard is modelled, because game code depends on it:
// the branch delay slot. The instruction after a branch or jump always executes
// before control transfers, so PC/nextPC are kept as a pair — Step advances PC
// to the delay slot before the branch runs, and the branch only chooses nextPC.
// The branch-likely forms additionally annul (skip) the slot when not taken.
//
// Unlike the R3000A there is no load delay slot: the VR4300 interlocks, so a
// loaded value is visible to the very next instruction and a single register
// file suffices.
//
// Memory goes through the Bus, so the caller (tools/platform/n64) supplies the
// machine model: RDRAM, the RCP's memory-mapped registers, SP DMEM/IMEM and the
// cartridge. Bus takes *physical* addresses; virtual-to-physical translation
// (the segment map and the TLB) happens here, in cop0.go.

import "fmt"

// Bus is the physical address space the CPU drives, big-endian.
//
// Read32/Write32 are not a convenience: the CPU fetches an instruction word for
// every instruction it runs, and composing each fetch from four interface calls
// (as tools/cpu/mips does) costs more than the interpreter itself. The machine
// model serves RDRAM and the SP memories from a backing slice through these,
// falling through to its address-range switch only for memory-mapped registers.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
	Read32(addr uint32) uint32
	Write32(addr uint32, v uint32)
}

// Fetcher lets a machine model distinguish an instruction fetch from a data
// read. A Bus that implements it has Fetch32 called for every instruction word,
// and Read32 only for loads — without which a "who reads this address" watch is
// drowned by the fetch of every instruction that runs inside the window.
type Fetcher interface {
	Fetch32(addr uint32) uint32
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
	cop0LLAddr   = 17
	cop0XContext = 20
	cop0ErrorEPC = 30
)

// Status register bits.
const (
	statusIE  = 1 << 0  // global interrupt enable
	statusEXL = 1 << 1  // exception level: an exception is being handled
	statusERL = 1 << 2  // error level
	statusBEV = 1 << 22 // bootstrap exception vectors (ROM, not RAM)
	statusFR  = 1 << 26 // full 32-register FPU file (vs 16 even-odd pairs)
	statusCU1 = 1 << 29 // coprocessor 1 usable
)

// Cause register bits. IP2 is the RCP's aggregated interrupt line; IP7 is the
// CPU's own Count/Compare timer.
const (
	causeBD  = 1 << 31 // exception taken in a branch delay slot
	causeIP2 = 1 << 10
	causeIP7 = 1 << 15
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
	excFPE  = 0x0F // floating-point exception
)

// Exception vector bases. Which is used depends on the Status BEV bit, and the
// offset on the kind of exception and whether one is already being handled.
const (
	vecRAM = 0x80000000
	vecROM = 0xBFC00200
)

// Exported names for the COP0 registers and Status/Cause bits a machine model
// needs to reach: the boot code sets Status directly, and the run loop reads the
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
	StatusFR  = statusFR
	StatusCU1 = statusCU1

	CauseIP2 = causeIP2
	CauseIP7 = causeIP7
)

// CPU is the VR4300 programmer's-model core.
type CPU struct {
	R      [32]uint64 // general registers; R[0] is hardwired to zero
	HI, LO uint64     // multiply/divide result registers
	PC     uint64     // address of the instruction to fetch next
	nextPC uint64     // address after PC (delay-slot machinery)

	COP0 [32]uint64 // system-control registers (see cop0.go)
	TLB  [TLBSize]TLBEntry

	FGR   [32]uint64 // floating-point register file (see fpu.go)
	FCR31 uint32     // FPU control/status: rounding mode, flags, condition bit

	LLBit bool // set by ll/lld, cleared by an intervening store or eret; sc/scd test it

	Halted     bool
	HaltReason string
	Steps      uint64

	bus   Bus
	fetch func(uint32) uint32 // bus.Fetch32 when the Bus is a Fetcher, else bus.Read32

	curPC        uint64 // address of the instruction currently executing
	delaySlot    bool   // the current instruction is in a branch delay slot
	pendingDelay bool   // the next instruction will be a delay slot
	branchAddr   uint64 // address of the branch whose delay slot we're in

	// countFrac carries the half-rate Count increment between steps: the counter
	// advances once per two CPU cycles, so it ticks on every other instruction.
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

// Reset returns the core to power-on state: PC at the CPU's reset vector in the
// PIF ROM, and the COP0 bits the hardware forces on reset.
//
// The N64 never executes from that vector under this model — the PIF ROM is not
// on the cartridge, so tools/platform/n64 supplies the state IPL2 would have left
// and starts the core at IPL3 instead. Reset still models it, so a core used
// standalone (in tests) begins somewhere defined.
func (c *CPU) Reset() {
	c.R = [32]uint64{}
	c.HI, c.LO = 0, 0
	c.PC, c.nextPC = 0xFFFFFFFFBFC00000, 0xFFFFFFFFBFC00004
	c.COP0 = [32]uint64{}
	// ERL and BEV are set out of reset: the ROM vectors are in use and KSEG0 is
	// unmapped-uncached until the boot code clears ERL.
	c.COP0[cop0Status] = statusERL | statusBEV
	c.COP0[cop0PRId] = 0x00000B22 // VR4300 revision, as the N64 reports
	c.COP0[cop0Config] = 0x7006E463
	c.COP0[cop0Random] = TLBSize - 1
	c.TLB = [TLBSize]TLBEntry{}
	c.FGR = [32]uint64{}
	c.FCR31 = 0
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
// jump to IPL3's entry point.
func (c *CPU) SetPC(pc uint64) {
	c.PC, c.nextPC = pc, pc+4
	c.pendingDelay = false
}

// SetReg writes a general register (R0 stays hardwired to zero).
func (c *CPU) SetReg(i uint32, v uint64) {
	if i != 0 {
		c.R[i] = v
	}
}

// Reg reads a general register.
func (c *CPU) Reg(i uint32) uint64 { return c.R[i] }

// CurPC returns the address of the instruction currently executing (valid inside
// a Bus write, e.g. to attribute "who wrote this address").
func (c *CPU) CurPC() uint64 { return c.curPC }

// set writes register i (R0 stays hardwired to zero).
func (c *CPU) set(i uint32, v uint64) {
	if i != 0 {
		c.R[i] = v
	}
}

// sext32 sign-extends a 32-bit result into a 64-bit register, which is what
// every 32-bit MIPS III operation does. Getting this wrong is invisible until a
// value is used as an address or compared as a 64-bit quantity.
func sext32(v uint32) uint64 { return uint64(int64(int32(v))) }

// --- memory helpers (big-endian, physical addresses) ------------------------

func (c *CPU) read8(a uint32) uint32     { return uint32(c.bus.Read(a)) }
func (c *CPU) write8(a uint32, v uint32) { c.bus.Write(a, byte(v)) }

func (c *CPU) read16(a uint32) uint32 {
	return uint32(c.bus.Read(a))<<8 | uint32(c.bus.Read(a+1))
}
func (c *CPU) write16(a uint32, v uint32) {
	c.bus.Write(a, byte(v>>8))
	c.bus.Write(a+1, byte(v))
}

func (c *CPU) read32(a uint32) uint32     { return c.bus.Read32(a) }
func (c *CPU) write32(a uint32, v uint32) { c.bus.Write32(a, v) }

func (c *CPU) read64(a uint32) uint64 {
	return uint64(c.bus.Read32(a))<<32 | uint64(c.bus.Read32(a+4))
}
func (c *CPU) write64(a uint32, v uint64) {
	c.bus.Write32(a, uint32(v>>32))
	c.bus.Write32(a+4, uint32(v))
}

// --- exceptions -------------------------------------------------------------

// Exception enters the exception handler with the given cause code, saving the
// return address in EPC (pointing at the branch if the fault was in a delay
// slot, with the Cause BD bit set).
//
// Vector selection follows the VR4300: a TLB refill taken with EXL clear enters
// at offset 0x000, everything else at 0x180; the base is 0x80000000 when Status
// BEV is clear and 0xBFC00200 when it is set. EXL is set on entry, so a nested
// fault re-enters at 0x180 rather than the refill vector.
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

// addrError raises an address-error exception for a misaligned access.
func (c *CPU) addrError(code uint32, vaddr uint64) {
	c.COP0[cop0BadVAddr] = vaddr
	c.Exception(code)
}

// --- interrupts -------------------------------------------------------------

// Interrupt updates the external interrupt line (Cause IP2) from the machine's
// masked RCP interrupt status and, if interrupts are enabled and unmasked, takes
// an interrupt exception. It returns whether one was delivered. Call it between
// instructions (before Step): the address saved in EPC is the instruction about
// to run, which resumes cleanly after the handler.
func (c *CPU) Interrupt(pending bool) bool {
	if pending {
		c.COP0[cop0Cause] |= causeIP2
	} else {
		c.COP0[cop0Cause] &^= causeIP2
	}
	return c.checkInterrupt()
}

// checkInterrupt delivers a pending, enabled, unmasked interrupt.
func (c *CPU) checkInterrupt() bool {
	sr := c.COP0[cop0Status]
	// Interrupts are blocked while an exception or error is being handled, and
	// when the global enable is clear.
	if sr&statusIE == 0 || sr&(statusEXL|statusERL) != 0 {
		return false
	}
	// Cause IP (bits 8..15) against Status IM (bits 8..15).
	if c.COP0[cop0Cause]&sr&0x0000FF00 == 0 {
		return false
	}
	if c.pendingDelay { // don't split a branch from its delay slot
		return false
	}
	// EPC must be the next instruction to fetch, not the last one executed, so
	// the handler's eret resumes exactly where control was preempted.
	c.curPC = c.PC
	c.delaySlot = false
	c.Exception(excInt)
	return true
}

// tickCount advances the Count register at half the CPU clock and raises the
// timer interrupt (IP7) when it reaches Compare. Count is 32-bit and wraps; the
// comparison is on equality, so a wrap between two steps must not skip it —
// hence the increment-then-test on the exact value.
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
