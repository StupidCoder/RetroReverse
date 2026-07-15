package gcdsp

import "fmt"

// cpu.go is the DSP's state: its register file, its two RAM spaces, and the accessors that
// build the wide accumulators out of the narrow registers they are stored in.
//
// The register file is 32 entries. Most are plain 16-bit registers; the accumulators are
// assembled from several of them. The layout is the hardware's, and the instruction set
// addresses registers by these indices:
//
//	 0- 3  AR0..AR3   address registers (point into data memory)
//	 4- 7  IX0..IX3   index registers (the signed step AR is advanced by)
//	 8-11  WR0..WR3   wrapping registers (the modulo an AR wraps at; 0xFFFF = no wrap)
//	12-15  ST0..ST3   the four hardware stacks: call, data, loop-address, loop-counter
//	16-17  AC0.H AC1.H   accumulator high  (bits 32..39, sign-extended through the word)
//	   18  CONFIG          the config register
//	   19  SR              the status register (flags in the low byte, modes in the high)
//	20-23  PROD.L M1 H M2  the multiplier product, in four pieces
//	24-27  AX0.L AX1.L AX0.H AX1.H   the two 32-bit extended-operand registers
//	28-31  AC0.L AC1.L AC0.M AC1.M   accumulator low and middle words
//
// AC0 and AC1 are 40-bit: high (8 significant bits, sign-extended), middle 16, low 16. AX0 and
// AX1 are 32-bit: high 16, low 16. PROD is 40-bit, held as low/mid1/high/mid2 so the two
// middle contributions a multiply-accumulate produces can be summed on read.
type CPU struct {
	PC  uint16    // program counter, a word address into instruction memory
	Reg [32]uint16 // the register file, indexed as above

	IRAM [0x1000]uint16 // instruction RAM: the microcode runs from here
	DRAM [0x1000]uint16 // data RAM: the ucode's working memory and command buffers

	// The console-resident ROMs. Absent by default: a read halts loudly rather than inventing
	// a value. A host that has them (a later, fuller model) can supply them.
	IROM []uint16 // instruction ROM at 0x8000
	DROM []uint16 // coefficient ROM at 0x1000 in the data space

	bus Bus // hardware registers in the data space (mailboxes, DMA, accelerator)

	Halted bool   // set when the core hits something it cannot model; Reason says what
	Reason string // the halt message

	// The remaining execution state, exported so the default gob encoder carries the whole core
	// in a savestate. Only bus is left out (it is a back-reference to the host, reattached with
	// SetBus after a load).
	Branched    bool          // the instruction just run has already placed PC; do not advance
	InInterrupt bool          // an interrupt handler is running (until RTI)
	Stacks      [4][]uint16   // the four hardware stacks, indexed as ST0..ST3
	Loops       []LoopFrame   // active hardware loops, innermost last
}

// LoopFrame is one running hardware loop: the first and last instruction addresses of the body
// and the number of iterations still to run.
type LoopFrame struct {
	Start uint16
	End   uint16
	Count uint16
}

// Register indices, named so the interpreter and disassembler agree.
const (
	regAR0 = 0
	regIX0 = 4
	regWR0 = 8
	regST0 = 12
	regST1 = 13
	regST2 = 14
	regST3 = 15
	regAC0H = 16
	regAC1H = 17
	regCONFIG = 18
	regSR   = 19
	regPRODL  = 20
	regPRODM1 = 21
	regPRODH  = 22
	regPRODM2 = 23
	regAX0L = 24
	regAX1L = 25
	regAX0H = 26
	regAX1H = 27
	regAC0L = 28
	regAC1L = 29
	regAC0M = 30
	regAC1M = 31
)

// Status-register flag bits (the low byte). The high byte carries the arithmetic mode bits,
// added as the multiply/accumulate ops that consult them are implemented.
const (
	srCarry     = 1 << 0 // CF: carry out of the last add/subtract
	srOverflow  = 1 << 1 // OF: signed overflow
	srZero      = 1 << 2 // ZF: the result was zero
	srSign      = 1 << 3 // SF: the result was negative
	srAboveS32  = 1 << 4 // AS: the accumulator does not fit in 32 bits
	srTopTwo    = 1 << 5 // TT: the top two bits of the accumulator are equal
	srLogicZero = 1 << 6 // OK/LZ: set by the logic ops when the result is zero
	srOverSticky = 1 << 7 // OS: overflow, sticky until cleared

	// The arithmetic-mode bits in the status register's high half, set and cleared by the
	// dedicated mode ops (set40/set16, clr15/set15, m0/m2). Their exact hardware positions
	// matter only to the multiply/accumulate ops that read them; the internal representation
	// here is consistent with those ops.
	srMulSigned = 1 << 13 // set15/clr15: multiplier operands unsigned vs signed
	srMode40    = 1 << 14 // set40/set16: 40-bit accumulator mode vs 16-bit saturation
	srMulShift  = 1 << 15 // m2/m0: multiply result shifted left one, or not
)

// Bus is the hardware the DSP reaches through the high end of its data address space
// (0xFF00..0xFFFF): the mailboxes to and from the CPU, the DMA engine, and the sample
// accelerator. The host implements it; the core calls it for any data access in that range.
type Bus interface {
	HWRead(addr uint16) uint16
	HWWrite(addr uint16, val uint16)
}

// New makes a DSP core reaching hardware through bus. The bus is kept out of the serialized
// state (it points back at the host); a reloaded core is reattached with SetBus.
func New(bus Bus) *CPU {
	return &CPU{bus: bus}
}

// SetBus attaches (or reattaches, after a savestate load) the hardware bus.
func (c *CPU) SetBus(bus Bus) { c.bus = bus }

// Clone deep-copies the core so an in-memory savestate does not share mutable state with the
// live machine. The register file and RAM are arrays, copied by the struct assignment; the
// slices are copied explicitly; the bus is dropped (a loaded core is reattached with SetBus).
func (c *CPU) Clone() *CPU {
	if c == nil {
		return nil
	}
	n := *c
	n.IROM = append([]uint16(nil), c.IROM...)
	n.DROM = append([]uint16(nil), c.DROM...)
	for i := range c.Stacks {
		n.Stacks[i] = append([]uint16(nil), c.Stacks[i]...)
	}
	n.Loops = append([]LoopFrame(nil), c.Loops...)
	n.bus = nil
	return &n
}

// Halt stops the core and records why. Every unmodelled encoding and every access to absent
// memory routes through here, so a run that cannot proceed says exactly what it hit.
func (c *CPU) Halt(format string, args ...any) {
	if !c.Halted {
		c.Halted = true
		c.Reason = fmt.Sprintf(format, args...)
	}
}

// --- accumulator accessors ---------------------------------------------------------------

// ac reads a 40-bit accumulator (0 or 1) as a sign-extended int64. High is 8 significant bits
// sign-extended through the stored word; the value is (high<<32)|(mid<<16)|low.
func (c *CPU) ac(n int) int64 {
	h := int64(int16(c.Reg[regAC0H+n])) // sign-extended high
	m := int64(c.Reg[regAC0M+n])
	l := int64(c.Reg[regAC0L+n])
	return (h << 32) | (m << 16) | l
}

// setAc writes a 40-bit accumulator, keeping only the low 40 bits and splitting them across
// the three registers. The high word is stored sign-extended so a later read sees the sign.
func (c *CPU) setAc(n int, v int64) {
	c.Reg[regAC0L+n] = uint16(v)
	c.Reg[regAC0M+n] = uint16(v >> 16)
	c.Reg[regAC0H+n] = uint16(v >> 32) // low 8 bits significant; sign carried in bit 39
	// Sign-extend the stored high word so int16() of it recovers the sign on read.
	if v&(1<<39) != 0 {
		c.Reg[regAC0H+n] |= 0xFF00
	} else {
		c.Reg[regAC0H+n] &= 0x00FF
	}
}

// ax reads a 32-bit extended register (0 or 1) as a sign-extended int64.
func (c *CPU) ax(n int) int64 {
	h := int64(int16(c.Reg[regAX0H+n]))
	l := int64(c.Reg[regAX0L+n])
	return (h << 16) | l
}

// prod reads the 40-bit product, summing the two middle contributions.
func (c *CPU) prod() int64 {
	l := int64(c.Reg[regPRODL])
	m1 := int64(c.Reg[regPRODM1])
	h := int64(int16(c.Reg[regPRODH]))
	m2 := int64(c.Reg[regPRODM2])
	p := (h << 32) | (m1 << 16) | l
	p += m2 << 16
	return p & 0xFFFFFFFFFF
}

// sr reads and writes the status register conveniently.
func (c *CPU) sr() uint16      { return c.Reg[regSR] }
func (c *CPU) setSR(v uint16)  { c.Reg[regSR] = v }
func (c *CPU) setFlag(bit uint16, on bool) {
	if on {
		c.Reg[regSR] |= bit
	} else {
		c.Reg[regSR] &^= bit
	}
}
