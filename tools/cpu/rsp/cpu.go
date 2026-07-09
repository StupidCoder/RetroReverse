package rsp

// cpu.go is the RSP's scalar core, and the plumbing the vector unit sits on.
//
// Memory is not a bus. The RSP addresses only its own 4 KiB of data memory, and
// every access wraps within it — a program cannot fault, because there is
// nowhere to fault to. Instruction fetch likewise wraps inside IMEM, so the
// 12-bit program counter is the whole story.
//
// Registers reached through COP0 belong to the machine, not to this package, so
// the caller supplies them through the Regs interface: reading SP_STATUS or
// kicking a DMA is something only tools/platform/n64 can do.

import "fmt"

// Regs is the RSP's view of the memory-mapped registers it reaches through COP0:
// its own SP block (indices 0..7) and the RDP's command queue (8..15).
type Regs interface {
	ReadCop0(reg uint32) uint32
	WriteCop0(reg uint32, v uint32)
}

// CPU is the RSP: a scalar core, a vector unit, and the two memories.
type CPU struct {
	R [32]uint32

	// VU is the vector unit's register file (see vu.go). Each register holds
	// eight 16-bit lanes; lane 0 is the most significant half of the 128 bits.
	V   [32][8]uint16
	Acc [8]uint64 // 48 bits per lane, held in the low bits
	VCO uint16    // carry (low byte) and not-equal (high byte) flags
	VCC uint16    // compare (low byte) and clip (high byte) flags
	VCE uint8     // compare-extension flags

	// The reciprocal unit's staging registers. A 32-bit divide is spread across
	// three instructions, and these carry the halves between them.
	divIn       uint16
	divOut      uint16
	divInLoaded bool

	DMEM []byte
	IMEM []byte

	PC     uint32 // 12-bit, wraps inside IMEM
	nextPC uint32

	Halted     bool
	Broke      bool // BREAK executed: the CPU is told through SP_STATUS
	HaltReason string
	Steps      uint64

	// Unimplemented counts the encodings this core met and did not model, by
	// instruction word.
	//
	// The RSP has no exception mechanism — no interrupts, no faults, nowhere to
	// vector to — so an encoding it does not implement simply does nothing. A
	// core that halted instead would stop microcode that real hardware runs. The
	// gap is recorded rather than swallowed, so a caller can still see it.
	Unimplemented map[uint32]int

	regs Regs

	curPC uint32
}

// NewCPU makes an RSP over the given memories and COP0 register window.
func NewCPU(dmem, imem []byte, regs Regs) *CPU {
	c := &CPU{DMEM: dmem, IMEM: imem, regs: regs}
	c.Reset()
	return c
}

// Reset clears the core and points it at the start of IMEM.
func (c *CPU) Reset() {
	c.R = [32]uint32{}
	c.V = [32][8]uint16{}
	c.Acc = [8]uint64{}
	c.VCO, c.VCC, c.VCE = 0, 0, 0
	c.divIn, c.divOut, c.divInLoaded = 0, 0, false
	c.PC, c.nextPC = 0, 4
	c.Halted, c.Broke, c.HaltReason = true, false, ""
	c.Steps = 0
}

// unimpl records an encoding this core does not model and continues. See the
// Unimplemented field for why this is not a halt.
func (c *CPU) unimpl(w uint32) {
	if c.Unimplemented == nil {
		c.Unimplemented = map[uint32]int{}
	}
	c.Unimplemented[w]++
}

// Halt stops the core, recording why.
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// SetPC points the core at an IMEM offset and clears any pending delay slot.
func (c *CPU) SetPC(pc uint32) {
	c.PC, c.nextPC = pc&0xFFC, (pc+4)&0xFFC
}

// CurPC is the address of the instruction currently executing.
func (c *CPU) CurPC() uint32 { return c.curPC }

// Start clears the halt and begins execution at pc.
func (c *CPU) Start(pc uint32) {
	c.SetPC(pc)
	c.Halted, c.Broke, c.HaltReason = false, false, ""
}

// --- memory (wraps inside DMEM; there is nowhere else to go) ----------------

func (c *CPU) rd8(a uint32) uint32    { return uint32(c.DMEM[a&0xFFF]) }
func (c *CPU) wr8(a uint32, v uint32) { c.DMEM[a&0xFFF] = byte(v) }

func (c *CPU) rd16(a uint32) uint32 {
	return uint32(c.DMEM[a&0xFFF])<<8 | uint32(c.DMEM[(a+1)&0xFFF])
}
func (c *CPU) wr16(a uint32, v uint32) {
	c.DMEM[a&0xFFF] = byte(v >> 8)
	c.DMEM[(a+1)&0xFFF] = byte(v)
}
func (c *CPU) rd32(a uint32) uint32 {
	return uint32(c.DMEM[a&0xFFF])<<24 | uint32(c.DMEM[(a+1)&0xFFF])<<16 |
		uint32(c.DMEM[(a+2)&0xFFF])<<8 | uint32(c.DMEM[(a+3)&0xFFF])
}
func (c *CPU) wr32(a uint32, v uint32) {
	c.DMEM[a&0xFFF] = byte(v >> 24)
	c.DMEM[(a+1)&0xFFF] = byte(v >> 16)
	c.DMEM[(a+2)&0xFFF] = byte(v >> 8)
	c.DMEM[(a+3)&0xFFF] = byte(v)
}

func (c *CPU) set(i, v uint32) {
	if i != 0 {
		c.R[i] = v
	}
}

// --- execution --------------------------------------------------------------

// Step executes one instruction. A halted core does nothing.
func (c *CPU) Step() {
	if c.Halted {
		return
	}
	c.curPC = c.PC

	w := uint32(c.IMEM[c.PC])<<24 | uint32(c.IMEM[c.PC+1])<<16 |
		uint32(c.IMEM[c.PC+2])<<8 | uint32(c.IMEM[c.PC+3])

	c.PC = c.nextPC
	c.nextPC = (c.nextPC + 4) & 0xFFC

	c.execute(w)
	c.Steps++
}

// Run steps until the core halts (a BREAK, or an unmodelled encoding) or the
// budget runs out. It returns the number of instructions executed.
func (c *CPU) Run(maxSteps uint64) uint64 {
	var n uint64
	for n < maxSteps && !c.Halted {
		c.Step()
		n++
	}
	return n
}

// doBranch redirects the instruction after next; the delay slot runs either way
// because PC has already advanced past the branch.
func (c *CPU) doBranch(taken bool, target uint32) {
	if taken {
		c.nextPC = target & 0xFFC
	}
}

func (c *CPU) execute(w uint32) {
	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	imm := w & 0xFFFF
	simm := uint32(int32(int16(imm)))
	branchT := (c.curPC + 4 + simm<<2) & 0xFFC
	jumpT := (w & 0x03FFFFFF << 2) & 0xFFC

	switch op {
	case 0x00:
		c.special(w, rs, rt)
	case 0x01: // REGIMM
		s := int32(c.R[rs])
		switch rt {
		case 0x00:
			c.doBranch(s < 0, branchT)
		case 0x01:
			c.doBranch(s >= 0, branchT)
		case 0x10:
			c.set(31, (c.curPC+8)&0xFFC)
			c.doBranch(s < 0, branchT)
		case 0x11:
			c.set(31, (c.curPC+8)&0xFFC)
			c.doBranch(s >= 0, branchT)
		default:
			c.unimpl(w)
		}
	case 0x02:
		c.doBranch(true, jumpT)
	case 0x03:
		c.set(31, (c.curPC+8)&0xFFC)
		c.doBranch(true, jumpT)
	case 0x04:
		c.doBranch(c.R[rs] == c.R[rt], branchT)
	case 0x05:
		c.doBranch(c.R[rs] != c.R[rt], branchT)
	case 0x06:
		c.doBranch(int32(c.R[rs]) <= 0, branchT)
	case 0x07:
		c.doBranch(int32(c.R[rs]) > 0, branchT)

	// No arithmetic exceptions exist, so addi and add do not trap.
	case 0x08, 0x09:
		c.set(rt, c.R[rs]+simm)
	case 0x0A:
		c.set(rt, b2u(int32(c.R[rs]) < int32(simm)))
	case 0x0B:
		c.set(rt, b2u(c.R[rs] < simm))
	case 0x0C:
		c.set(rt, c.R[rs]&imm)
	case 0x0D:
		c.set(rt, c.R[rs]|imm)
	case 0x0E:
		c.set(rt, c.R[rs]^imm)
	case 0x0F:
		c.set(rt, imm<<16)

	case 0x10: // COP0
		rd := (w >> 11) & 31
		switch rs {
		case 0x00:
			c.set(rt, c.regs.ReadCop0(rd))
		case 0x04:
			c.regs.WriteCop0(rd, c.R[rt])
		default:
			c.unimpl(w)
		}
	case 0x12: // COP2
		c.cop2(w)

	case 0x20: // lb
		c.set(rt, uint32(int32(int8(byte(c.rd8(c.R[rs]+simm))))))
	case 0x24: // lbu
		c.set(rt, c.rd8(c.R[rs]+simm))
	case 0x21: // lh
		c.set(rt, uint32(int32(int16(uint16(c.rd16(c.R[rs]+simm))))))
	case 0x25: // lhu
		c.set(rt, c.rd16(c.R[rs]+simm))
	case 0x23: // lw
		c.set(rt, c.rd32(c.R[rs]+simm))
	case 0x28:
		c.wr8(c.R[rs]+simm, c.R[rt])
	case 0x29:
		c.wr16(c.R[rs]+simm, c.R[rt])
	case 0x2B:
		c.wr32(c.R[rs]+simm, c.R[rt])

	case 0x32: // LWC2
		c.vecLoad(w, rs, rt)
	case 0x3A: // SWC2
		c.vecStore(w, rs, rt)

	default:
		c.unimpl(w)
	}
}

func (c *CPU) special(w, rs, rt uint32) {
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	switch w & 63 {
	case 0x00:
		c.set(rd, c.R[rt]<<shamt)
	case 0x02:
		c.set(rd, c.R[rt]>>shamt)
	case 0x03:
		c.set(rd, uint32(int32(c.R[rt])>>shamt))
	case 0x04:
		c.set(rd, c.R[rt]<<(c.R[rs]&31))
	case 0x06:
		c.set(rd, c.R[rt]>>(c.R[rs]&31))
	case 0x07:
		c.set(rd, uint32(int32(c.R[rt])>>(c.R[rs]&31)))
	case 0x08:
		c.doBranch(true, c.R[rs])
	case 0x09:
		c.set(rd, (c.curPC+8)&0xFFC)
		c.doBranch(true, c.R[rs])
	case 0x0D: // break
		c.Broke = true
		c.Halted = true
		c.HaltReason = "break"
	case 0x20, 0x21:
		c.set(rd, c.R[rs]+c.R[rt])
	case 0x22, 0x23:
		c.set(rd, c.R[rs]-c.R[rt])
	case 0x24:
		c.set(rd, c.R[rs]&c.R[rt])
	case 0x25:
		c.set(rd, c.R[rs]|c.R[rt])
	case 0x26:
		c.set(rd, c.R[rs]^c.R[rt])
	case 0x27:
		c.set(rd, ^(c.R[rs] | c.R[rt]))
	case 0x2A:
		c.set(rd, b2u(int32(c.R[rs]) < int32(c.R[rt])))
	case 0x2B:
		c.set(rd, b2u(c.R[rs] < c.R[rt]))
	default:
		c.unimpl(w)
	}
}

func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
