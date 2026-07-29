package sh4

// cpu.go is the processor's state and the interface it reaches memory through.
//
// Two shapes matter here. The general registers: R0-R7 exist twice, and SR.RB
// selects which bank is live while the CPU is privileged — an interrupt entry
// sets RB, so a handler wakes up on its own registers and the interrupted
// code's R0-R7 are simply the other bank. The CPU keeps the live file in R and
// the dormant one in Rbank, swapping on the SR writes that change the
// selection, so the interpreter always indexes R directly.
//
// And the floating-point registers: two banks of sixteen 32-bit values, held
// as raw bits. FPSCR.FR names one bank FR0-FR15 and the other XF0-XF15
// (frchg is a bit flip, not a copy); FPSCR.PR reinterprets even-odd pairs as
// doubles (DRn = FRn:FRn+1, FRn the high word); FPSCR.SZ makes fmov move the
// pair. Bits, not floats, are stored so a savestate round-trips NaNs exactly.
//
// Addresses are handled the way a Dreamcast game uses them: P0-P3
// (0x00000000-0xDFFFFFFF) drop their top three bits and go to the external
// bus — the MMU stays off in retail software, and enabling it halts the core
// loudly (see onchip.go). P4 never reaches the bus: store queues at
// 0xE0000000-0xE3FFFFFF (sq.go) and the on-chip registers at 0xFC000000+
// (onchip.go) are the CPU's own.

import "fmt"

// Bus is the memory the CPU talks to. The widths are distinct rather than
// synthesized from bytes because the machine's registers care: a Holly
// register read as two halfwords is not the same event as one read as a word,
// and a bus that could not tell the difference would force the platform to
// guess. Addresses are physical: the CPU has already stripped the P1/P2/P3
// mirror bits.
type Bus interface {
	Read8(addr uint32) uint8
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}

// Fetcher lets a machine model tell an instruction fetch from a data read. A
// bus that does not implement it is simply read with Read16.
type Fetcher interface {
	Fetch16(addr uint32) uint16
}

// SR bits.
const (
	SRT     = 1 << 0  // the test/carry bit every compare writes
	SRS     = 1 << 1  // MAC saturation
	SRQ     = 1 << 8  // div1 state
	SRM     = 1 << 9  // div1 state
	SRFD    = 1 << 15 // FPU disable
	SRBL    = 1 << 28 // exceptions/interrupts blocked
	SRRB    = 1 << 29 // register bank select (privileged only)
	SRMD    = 1 << 30 // privileged mode
	srIMASK = 0xF << 4
	srMask  = SRMD | SRRB | SRBL | SRFD | SRM | SRQ | srIMASK | SRS | SRT
)

// FPSCR bits.
const (
	FPSCRFR   = 1 << 21 // register bank select
	FPSCRSZ   = 1 << 20 // fmov moves 64 bits
	FPSCRPR   = 1 << 19 // arithmetic is double-precision
	FPSCRDN   = 1 << 18 // denormals flush to zero
	fpscrMask = 0x003FFFFF
)

// CPU is an SH7091.
type CPU struct {
	R     [16]uint32
	Rbank [8]uint32 // the dormant bank of R0-R7 (see setSR)

	SR   uint32
	GBR  uint32
	VBR  uint32
	SSR  uint32
	SPC  uint32
	SGR  uint32
	DBR  uint32
	MACH uint32
	MACL uint32
	PR   uint32
	PC   uint32

	fpr   [2][16]uint32 // raw bits; FPSCR.FR picks which bank is FR0-15
	FPSCR uint32
	FPUL  uint32

	// The modeled on-chip registers (onchip.go). Everything else written in
	// the P4 register area lands in onchipRaw and is reported by Gaps().
	MMUCR  uint32
	CCR    uint32
	TRA    uint32
	EXPEVT uint32
	INTEVT uint32
	PTEH, PTEL, TTB, TEA uint32
	QACR0, QACR1         uint32
	ICR                  uint32
	IPRA, IPRB, IPRC     uint32
	TMU                  TMUState

	// SQ are the two 32-byte store queues (sq.go).
	SQ [2][8]uint32

	// The external interrupt request: a level (compared against SR.IMASK) and
	// the INTEVT code the handler will read. Level 0 means deasserted. The
	// machine's interrupt controller owns both via SetIRL.
	irlLevel uint32
	irlCode  uint32

	Halted     bool
	HaltReason string
	Steps      uint64

	bus     Bus
	fetcher Fetcher

	// The delay-slot pipeline (see Step): PC/nextPC advance in lockstep and a
	// delayed transfer redirects nextPC, so the slot at PC still executes.
	nextPC       uint32
	curPC        uint32
	delaySlot    bool
	pendingDelay bool

	onchipRaw  map[uint32]uint32
	onchipGaps map[uint32]int
}

// NewCPU makes an SH7091 attached to a bus, in the state a reset leaves it in.
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	if f, ok := bus.(Fetcher); ok {
		c.fetcher = f
	}
	c.Reset()
	return c
}

// Reset puts the processor where the power-on sequence leaves it: privileged,
// interrupts masked and blocked, FPU in round-to-zero flush-denormals mode,
// and the program counter at the reset vector.
func (c *CPU) Reset() {
	*c = CPU{bus: c.bus, fetcher: c.fetcher}
	c.SR = SRMD | SRRB | SRBL | srIMASK // 0x700000F0
	c.FPSCR = 0x00040001                // DN set, rounding to zero
	c.PC = 0xA0000000
	c.nextPC = c.PC + 2
	c.EXPEVT = 0 // power-on reset code
	c.onchipRaw = map[uint32]uint32{}
	c.onchipGaps = map[uint32]int{}
}

// Halt stops the core, recording why. Every gap in the implementation ends
// here rather than in a silently wrong result: an instruction this core does
// not know is a fact about the core, and a fact should be loud.
func (c *CPU) Halt(format string, args ...interface{}) {
	if c.Halted {
		return // keep the first reason: it is the one that explains the rest
	}
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// CurPC is the address of the instruction being executed.
func (c *CPU) CurPC() uint32 { return c.curPC }

// SetPC jumps, resynchronising the pipeline.
func (c *CPU) SetPC(pc uint32) {
	c.PC = pc
	c.nextPC = pc + 2
	c.pendingDelay, c.delaySlot = false, false
}

// Reg reads a general register; SetReg writes one.
func (c *CPU) Reg(i uint32) uint32       { return c.R[i&15] }
func (c *CPU) SetReg(i uint32, v uint32) { c.R[i&15] = v }

// InDelaySlot reports whether the instruction at CurPC is a delay slot — the
// machine's HLE layer needs it to refuse to trap there.
func (c *CPU) InDelaySlot() bool { return c.delaySlot }

// bankSelect is 1 when the second R0-R7 bank is live: privileged with SR.RB
// set. In user mode RB is ignored.
func bankSelect(sr uint32) uint32 {
	if sr&SRMD != 0 && sr&SRRB != 0 {
		return 1
	}
	return 0
}

// SetSR writes the status register, swapping the R0-R7 bank when the
// selection changes so R always holds the live file.
func (c *CPU) SetSR(v uint32) {
	v &= srMask
	if bankSelect(v) != bankSelect(c.SR) {
		for i := 0; i < 8; i++ {
			c.R[i], c.Rbank[i] = c.Rbank[i], c.R[i]
		}
	}
	c.SR = v
}

// SetFPSCR writes the floating-point status register. The FR bank swap is a
// selector change, not a copy: fr() indexes off the bit.
func (c *CPU) SetFPSCR(v uint32) { c.FPSCR = v & fpscrMask }

// T reads the test bit; setT writes it.
func (c *CPU) T() uint32 {
	return c.SR & SRT
}
func (c *CPU) setT(cond bool) {
	if cond {
		c.SR |= SRT
	} else {
		c.SR &^= SRT
	}
}

// SetIRL asserts (level 1-15) or deasserts (level 0) the external interrupt
// request. code is the INTEVT value the handler will find; the machine's
// interrupt controller computes it, because the encoding is the board's
// business, not the core's.
func (c *CPU) SetIRL(level, code uint32) {
	c.irlLevel, c.irlCode = level&15, code
}

// --- Memory access ----------------------------------------------------------
//
// P0-P3 strip to physical and go to the bus; P4 is the CPU's own. The one
// deliberate simplification is named in the package doc: no MMU, no cache —
// P1's cached window and P2's uncached window read the same bytes here.

func (c *CPU) fetchInstr(addr uint32) uint16 {
	if addr >= 0xE0000000 {
		c.Halt("instruction fetch from P4 at %08X", addr)
		return 0x0009 // nop; the halt already stopped the run
	}
	if c.fetcher != nil {
		return c.fetcher.Fetch16(addr & 0x1FFFFFFF)
	}
	return c.bus.Read16(addr & 0x1FFFFFFF)
}

func (c *CPU) read8(addr uint32) uint8 {
	switch {
	case addr < 0xE0000000:
		return c.bus.Read8(addr & 0x1FFFFFFF)
	case addr < 0xE4000000:
		return uint8(c.sqRead32(addr) >> (8 * (addr & 3)))
	case addr >= 0xFC000000:
		return uint8(c.onchipRead(addr, 1))
	}
	c.Halt("read8 from unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	return 0
}

func (c *CPU) read16(addr uint32) uint16 {
	switch {
	case addr < 0xE0000000:
		return c.bus.Read16(addr & 0x1FFFFFFF)
	case addr < 0xE4000000:
		return uint16(c.sqRead32(addr) >> (8 * (addr & 2)))
	case addr >= 0xFC000000:
		return uint16(c.onchipRead(addr, 2))
	}
	c.Halt("read16 from unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	return 0
}

func (c *CPU) read32(addr uint32) uint32 {
	switch {
	case addr < 0xE0000000:
		return c.bus.Read32(addr & 0x1FFFFFFF)
	case addr < 0xE4000000:
		return c.sqRead32(addr)
	case addr >= 0xFC000000:
		return c.onchipRead(addr, 4)
	}
	c.Halt("read32 from unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	return 0
}

func (c *CPU) write8(addr uint32, v uint8) {
	switch {
	case addr < 0xE0000000:
		c.bus.Write8(addr&0x1FFFFFFF, v)
	case addr < 0xE4000000:
		c.sqWrite(addr, 1, uint32(v))
	case addr >= 0xFC000000:
		c.onchipWrite(addr, 1, uint32(v))
	default:
		c.Halt("write8 to unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	}
}

func (c *CPU) write16(addr uint32, v uint16) {
	switch {
	case addr < 0xE0000000:
		c.bus.Write16(addr&0x1FFFFFFF, v)
	case addr < 0xE4000000:
		c.sqWrite(addr, 2, uint32(v))
	case addr >= 0xFC000000:
		c.onchipWrite(addr, 2, uint32(v))
	default:
		c.Halt("write16 to unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	}
}

func (c *CPU) write32(addr uint32, v uint32) {
	switch {
	case addr < 0xE0000000:
		c.bus.Write32(addr&0x1FFFFFFF, v)
	case addr < 0xE4000000:
		c.sqWrite(addr, 4, v)
	case addr >= 0xFC000000:
		c.onchipWrite(addr, 4, v)
	default:
		c.Halt("write32 to unmapped P4 at %08X (PC %08X)", addr, c.curPC)
	}
}
