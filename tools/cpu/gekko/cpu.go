package gekko

// cpu.go is the processor's state and the interface it reaches memory through.
//
// The register file is PowerPC's, plus the Gekko's own. What is worth naming here is
// the shape of the floating-point registers, because it is the whole reason this is a
// separate core: each is *two* values, not one. In ordinary floating-point mode a
// register holds one double, in PS0, and PS1 is along for the ride. In paired-single
// mode both halves are live singles and one instruction operates on both. Because the
// two modes share the same registers, PS1 has to be carried and preserved even by code
// that has never heard of it — a plain fadd writes PS0 and must leave PS1 alone.

import (
	"fmt"
	"math"
)

// Bus is the memory the CPU talks to. The widths are distinct rather than synthesized
// from bytes because the machine's registers care: a GameCube hardware register that is
// read as two halfwords is not the same event as one read as a word, and a bus that
// could not tell the difference would force the platform to guess.
type Bus interface {
	Read8(addr uint32) uint8
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}

// Fetcher lets a machine model tell an instruction fetch from a data read. A bus that
// does not implement it is simply read with Read32.
type Fetcher interface {
	Fetch32(addr uint32) uint32
}

// FPR is one floating-point register: two values, always.
//
// PS0 is the architectural double — what fadd writes and what stfd stores. PS1 is the
// second single of a pair, which only the paired-single instructions and the quantised
// loads touch. Every other instruction must leave it alone, and getting that wrong
// produces a core that passes every scalar test and then draws wrong geometry.
type FPR struct {
	PS0, PS1 float64
}

// CPU is a Gekko.
type CPU struct {
	GPR [32]uint32
	FPR [32]FPR

	PC  uint32
	LR  uint32
	CTR uint32
	CR  uint32 // eight independent four-bit fields; CR0 is the top nibble
	XER uint32
	MSR uint32

	FPSCR uint32

	// The special-purpose registers the Gekko adds. GQR selects the format and scale a
	// quantised load or store converts through; HID2 switches the paired-single unit,
	// the locked cache and the write-gather pipe on.
	GQR  [8]uint32
	HID0 uint32
	HID1 uint32
	HID2 uint32
	HID4 uint32
	WPAR uint32
	DMAU uint32
	DMAL uint32
	L2CR uint32

	// The exception and address-translation registers.
	SRR0, SRR1 uint32
	SPRG       [4]uint32
	DSISR, DAR uint32
	SDR1       uint32
	SR         [16]uint32   // segment registers
	IBAT       [4][2]uint32 // [n][0] = upper, [n][1] = lower
	DBAT       [4][2]uint32
	PVR        uint32

	// The time base and the decrementer, both paced by the instruction count. clockFrac
	// carries the core clocks that have not yet added up to a timer tick; decArmed is the
	// edge on which the decrementer's exception fires, so that one underflow raises one
	// exception rather than one per instruction spent below zero.
	TB        uint64
	DEC       uint32
	clockFrac uint32
	decArmed  bool

	// The reservation that lwarx sets and stwcx. tests. It is not a lock: it is one
	// address and one bit, and any store to the line clears it.
	Reserved    bool
	ReserveAddr uint32

	// The locked cache: 16 KiB of the L1 data cache, dropped out of coherency and mapped
	// as a scratchpad. See cache.go.
	LC LockedCache

	Halted     bool
	HaltReason string
	Steps      uint64

	// ExtInt is the external interrupt line, held by the machine's interrupt controller.
	// It is level-sensitive: while it is up and MSR[EE] is set, the CPU takes the
	// exception at the top of every instruction.
	ExtInt bool

	bus     Bus
	fetcher Fetcher

	// SC is called for a `sc` instruction. A GameCube game has no operating system to
	// call, so nothing normally installs this — but the machine may, to trap the
	// apploader's callbacks, and the exception path is taken when it is nil or returns
	// false.
	SC func(c *CPU) bool
}

// NewCPU makes a Gekko attached to a bus, in the state a reset leaves it in.
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	if f, ok := bus.(Fetcher); ok {
		c.fetcher = f
	}
	c.Reset()
	return c
}

// Reset puts the processor where the power-on sequence leaves it: translation off,
// exceptions at the high vectors, and the program counter at the reset vector.
func (c *CPU) Reset() {
	*c = CPU{bus: c.bus, fetcher: c.fetcher, SC: c.SC}
	c.MSR = MSRIP | MSRME // vectors high, machine check enabled; translation and interrupts off
	c.PC = 0xFFF00100     // the system-reset vector, with MSR[IP] set
	// The processor version register identifies a Gekko. Software reads it to decide
	// which errata to work around, so it has to be a real value rather than zero.
	c.PVR = 0x00083214
	c.FPSCR = 0
	c.LC.Reset()
}

// Halt stops the core, recording why. Every gap in the implementation ends here rather
// than in a silently wrong result: an instruction this core does not know is a fact
// about the core, and a fact should be loud.
func (c *CPU) Halt(format string, args ...interface{}) {
	if c.Halted {
		return // keep the first reason: it is the one that explains the rest
	}
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// CurPC is the address of the instruction being executed.
func (c *CPU) CurPC() uint64 { return uint64(c.PC) }

// SetPC jumps.
func (c *CPU) SetPC(pc uint32) { c.PC = pc }

// Reg reads a general register.
func (c *CPU) Reg(i uint32) uint32 { return c.GPR[i&31] }

// SetReg writes one.
func (c *CPU) SetReg(i uint32, v uint32) { c.GPR[i&31] = v }

// --- The condition register -------------------------------------------------------

// CRField reads one of the eight four-bit fields. Field 0 is the top nibble of CR, which
// is the opposite of the intuition a little-endian machine gives you.
func (c *CPU) CRField(n uint32) uint32 {
	return (c.CR >> (28 - 4*(n&7))) & 0xF
}

// SetCRField writes one.
func (c *CPU) SetCRField(n, v uint32) {
	sh := 28 - 4*(n&7)
	c.CR = (c.CR &^ (0xF << sh)) | ((v & 0xF) << sh)
}

// The bits within a condition-register field, in the order PowerPC numbers them.
const (
	crLT = 8
	crGT = 4
	crEQ = 2
	crSO = 1
)

// setCR0 records the sign of a result in field 0, which is what the "." suffix means.
// The summary-overflow bit is copied in from XER, so a compare can see that an earlier
// arithmetic instruction overflowed.
func (c *CPU) setCR0(v uint32) {
	f := uint32(0)
	switch {
	case int32(v) < 0:
		f = crLT
	case int32(v) > 0:
		f = crGT
	default:
		f = crEQ
	}
	if c.XER&XERSO != 0 {
		f |= crSO
	}
	c.SetCRField(0, f)
}

// setCR1 copies the top four bits of FPSCR into field 1, which is what the "." suffix
// means on a floating-point instruction.
func (c *CPU) setCR1() {
	c.SetCRField(1, c.FPSCR>>28)
}

// --- XER --------------------------------------------------------------------------

func (c *CPU) setCA(on bool) {
	if on {
		c.XER |= XERCA
	} else {
		c.XER &^= XERCA
	}
}

func (c *CPU) ca() uint32 {
	if c.XER&XERCA != 0 {
		return 1
	}
	return 0
}

// setOV records an overflow, and sets the sticky summary bit with it. SO is only ever
// cleared by writing XER or by mcrxr — an overflow is remembered until someone looks.
func (c *CPU) setOV(on bool) {
	if on {
		c.XER |= XEROV | XERSO
	} else {
		c.XER &^= XEROV
	}
}

// --- Memory -----------------------------------------------------------------------
//
// Every access goes through address translation first. On a GameCube that means the
// block-address-translation registers, which map the whole of memory in a handful of
// entries; the page table exists in the architecture but a GameCube never installs one.
// See mmu.go.

func (c *CPU) read8(ea uint32) uint8 {
	pa, ok := c.Translate(ea, false, false)
	if !ok {
		return 0
	}
	if c.LC.Contains(pa) {
		return c.LC.Read8(pa)
	}
	return c.bus.Read8(pa)
}

func (c *CPU) read16(ea uint32) uint16 {
	pa, ok := c.Translate(ea, false, false)
	if !ok {
		return 0
	}
	if c.LC.Contains(pa) {
		return c.LC.Read16(pa)
	}
	return c.bus.Read16(pa)
}

func (c *CPU) read32(ea uint32) uint32 {
	pa, ok := c.Translate(ea, false, false)
	if !ok {
		return 0
	}
	if c.LC.Contains(pa) {
		return c.LC.Read32(pa)
	}
	return c.bus.Read32(pa)
}

func (c *CPU) read64(ea uint32) uint64 {
	return uint64(c.read32(ea))<<32 | uint64(c.read32(ea+4))
}

func (c *CPU) write8(ea uint32, v uint8) {
	pa, ok := c.Translate(ea, true, false)
	if !ok {
		return
	}
	if c.LC.Contains(pa) {
		c.LC.Write8(pa, v)
		return
	}
	c.clearReservation(pa)
	c.bus.Write8(pa, v)
}

func (c *CPU) write16(ea uint32, v uint16) {
	pa, ok := c.Translate(ea, true, false)
	if !ok {
		return
	}
	if c.LC.Contains(pa) {
		c.LC.Write16(pa, v)
		return
	}
	c.clearReservation(pa)
	c.bus.Write16(pa, v)
}

func (c *CPU) write32(ea uint32, v uint32) {
	pa, ok := c.Translate(ea, true, false)
	if !ok {
		return
	}
	if c.LC.Contains(pa) {
		c.LC.Write32(pa, v)
		return
	}
	c.clearReservation(pa)
	c.bus.Write32(pa, v)
}

func (c *CPU) write64(ea uint32, v uint64) {
	c.write32(ea, uint32(v>>32))
	c.write32(ea+4, uint32(v))
}

// clearReservation drops a lwarx reservation when anything writes the line it covers.
// Without it a store-conditional could succeed across a write it should have seen.
func (c *CPU) clearReservation(pa uint32) {
	if c.Reserved && pa&^31 == c.ReserveAddr&^31 {
		c.Reserved = false
	}
}

// ReadMem is the machine's window into the CPU's view of memory, translation and locked
// cache included — so a debugger sees what the program sees.
func (c *CPU) ReadMem(ea uint32) uint8 { return c.read8(ea) }

// fetch reads the instruction at pc.
func (c *CPU) fetch(pc uint32) uint32 {
	pa, ok := c.Translate(pc, false, true)
	if !ok {
		return 0
	}
	if c.fetcher != nil {
		return c.fetcher.Fetch32(pa)
	}
	return c.bus.Read32(pa)
}

// --- Floating-point helpers ---------------------------------------------------------

// f32 rounds a double to single precision and back. It is not cosmetic: the
// single-precision instructions genuinely compute in double and then round, and that
// second rounding is observable — a value that is exactly representable after one
// rounding need not be after two.
func f32(v float64) float64 { return float64(float32(v)) }

// psEnabled reports whether the paired-single unit is on. Software must set HID2[PSE]
// before any ps_ instruction, and a core that ignored the bit would happily execute
// paired singles that the real machine would have refused.
func (c *CPU) psEnabled() bool { return c.HID2&HID2PSE != 0 }

// The bit-pattern conversions. They are named f64* rather than bits/fromBits because
// exec31.go needs math/bits, and a helper that shadowed a standard package would be a
// trap for whoever adds the next instruction.
func f64bits(f float64) uint64 { return math.Float64bits(f) }
func f64from(b uint64) float64 { return math.Float64frombits(b) }
func bits64(f float64) uint64  { return math.Float64bits(f) }

// float32bitsOf rounds a double to single and returns its bit pattern — what stfs stores.
func float32bitsOf(v float64) uint32 { return math.Float32bits(float32(v)) }
