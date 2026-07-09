package r4300

// fpu.go is COP1, the VR4300's floating-point unit: 32 registers, IEEE-754
// single and double precision, and integer conversions in both 32- and 64-bit
// widths.
//
// The register file has two shapes, selected by the Status FR bit:
//
//	FR = 0  16 usable double registers, formed by pairing even/odd 32-bit
//	        halves. A double lives in $f(2n) (low half) and $f(2n+1) (high).
//	FR = 1  32 independent 64-bit registers.
//
// libultra sets FR, but the boot code runs before it does, so both are modelled.
// Reading a double from an odd register with FR clear is undefined on hardware;
// here it takes the even register of the pair, which is what compilers assume.
//
// The rounding mode lives in the low two bits of FCR31, and the condition bit
// that c.cond.fmt writes and bc1t/bc1f test is bit 23.

import "math"

// FCR31 bits.
const (
	fcr31Cond = 1 << 23 // the condition flag set by c.cond.fmt
)

// Rounding modes (FCR31 bits 0..1).
const (
	roundNearest = 0
	roundToZero  = 1
	roundCeil    = 2
	roundFloor   = 3
)

// readFGR32 / writeFGR32 access a single-precision value. With FR clear the
// 32-bit registers are the halves of the 16 doubles; with FR set each register
// holds its own single in the low half.
func (c *CPU) readFGR32(i uint32) uint32 {
	if c.COP0[cop0Status]&statusFR != 0 {
		return uint32(c.FGR[i])
	}
	if i&1 != 0 {
		return uint32(c.FGR[i&^1] >> 32)
	}
	return uint32(c.FGR[i])
}

func (c *CPU) writeFGR32(i uint32, v uint32) {
	if c.COP0[cop0Status]&statusFR != 0 {
		c.FGR[i] = uint64(v)
		return
	}
	if i&1 != 0 {
		c.FGR[i&^1] = c.FGR[i&^1]&0x00000000FFFFFFFF | uint64(v)<<32
		return
	}
	c.FGR[i] = c.FGR[i]&0xFFFFFFFF00000000 | uint64(v)
}

// readFGR64 / writeFGR64 access a double-precision value.
func (c *CPU) readFGR64(i uint32) uint64 {
	if c.COP0[cop0Status]&statusFR == 0 {
		i &^= 1
	}
	return c.FGR[i]
}

func (c *CPU) writeFGR64(i uint32, v uint64) {
	if c.COP0[cop0Status]&statusFR == 0 {
		i &^= 1
	}
	c.FGR[i] = v
}

func (c *CPU) fs(i uint32) float32  { return math.Float32frombits(c.readFGR32(i)) }
func (c *CPU) fd(i uint32) float64  { return math.Float64frombits(c.readFGR64(i)) }
func (c *CPU) setS(i uint32, v float32) { c.writeFGR32(i, math.Float32bits(v)) }
func (c *CPU) setD(i uint32, v float64) { c.writeFGR64(i, math.Float64bits(v)) }

// round applies the FCR31 rounding mode to a float, yielding an integral float.
func (c *CPU) round(v float64) float64 {
	switch c.FCR31 & 3 {
	case roundToZero:
		return math.Trunc(v)
	case roundCeil:
		return math.Ceil(v)
	case roundFloor:
		return math.Floor(v)
	}
	return math.RoundToEven(v)
}

// cop1Mem executes the FPU load/store instructions, which move raw bit patterns
// and never round.
func (c *CPU) cop1Mem(op, rs, ft uint32, simm uint64) {
	vaddr := c.R[rs] + simm
	switch op {
	case 0x31: // lwc1
		if vaddr&3 != 0 {
			c.addrError(excAdEL, vaddr)
			return
		}
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		c.writeFGR32(ft, c.read32(p))
	case 0x35: // ldc1
		if vaddr&7 != 0 {
			c.addrError(excAdEL, vaddr)
			return
		}
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		c.writeFGR64(ft, c.read64(p))
	case 0x39: // swc1
		if vaddr&3 != 0 {
			c.addrError(excAdES, vaddr)
			return
		}
		p, ok := c.Translate(vaddr, true)
		if !ok {
			return
		}
		c.write32(p, c.readFGR32(ft))
	case 0x3D: // sdc1
		if vaddr&7 != 0 {
			c.addrError(excAdES, vaddr)
			return
		}
		p, ok := c.Translate(vaddr, true)
		if !ok {
			return
		}
		c.write64(p, c.readFGR64(ft))
	}
}

// cop1 executes an op == 0x11 instruction: the register moves, the conditional
// branches, and the format-dispatched arithmetic.
func (c *CPU) cop1(w, rs, rt, rd, shamt uint32, branchT uint64) {
	switch rs {
	case 0x00: // mfc1
		c.set(rt, sext32(c.readFGR32(rd)))
		return
	case 0x01: // dmfc1
		c.set(rt, c.readFGR64(rd))
		return
	case 0x02: // cfc1
		switch rd {
		case 0:
			c.set(rt, 0x00000A00) // FCR0: the revision register
		case 31:
			c.set(rt, sext32(c.FCR31))
		}
		return
	case 0x04: // mtc1
		c.writeFGR32(rd, uint32(c.R[rt]))
		return
	case 0x05: // dmtc1
		c.writeFGR64(rd, c.R[rt])
		return
	case 0x06: // ctc1
		if rd == 31 {
			c.FCR31 = uint32(c.R[rt])
		}
		return

	case 0x08: // BC1: branch on the condition flag
		cond := c.FCR31&fcr31Cond != 0
		switch rt & 3 {
		case 0: // bc1f
			c.doBranch(!cond, branchT)
		case 1: // bc1t
			c.doBranch(cond, branchT)
		case 2: // bc1fl
			c.doBranchLikely(!cond, branchT)
		case 3: // bc1tl
			c.doBranchLikely(cond, branchT)
		}
		return
	}

	funct := w & 0x3F
	ft, fs, fd := rt, rd, shamt

	switch rs {
	case 0x10: // single precision
		c.cop1Single(w, funct, ft, fs, fd)
	case 0x11: // double precision
		c.cop1Double(w, funct, ft, fs, fd)
	case 0x14: // 32-bit integer source: only the conversions are defined
		switch funct {
		case 0x20: // cvt.s.w
			c.setS(fd, float32(int32(c.readFGR32(fs))))
		case 0x21: // cvt.d.w
			c.setD(fd, float64(int32(c.readFGR32(fs))))
		default:
			c.Halt("unimplemented cop1 w-format funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, uint32(c.curPC))
		}
	case 0x15: // 64-bit integer source
		switch funct {
		case 0x20: // cvt.s.l
			c.setS(fd, float32(int64(c.readFGR64(fs))))
		case 0x21: // cvt.d.l
			c.setD(fd, float64(int64(c.readFGR64(fs))))
		default:
			c.Halt("unimplemented cop1 l-format funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, uint32(c.curPC))
		}
	default:
		c.Halt("unimplemented cop1 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, uint32(c.curPC))
	}
}

func (c *CPU) cop1Single(w, funct, ft, fs, fd uint32) {
	if funct&0x30 == 0x30 { // c.cond.s
		c.compare(float64(c.fs(fs)), float64(c.fs(ft)), funct&0xF)
		return
	}
	switch funct {
	case 0x00:
		c.setS(fd, c.fs(fs)+c.fs(ft))
	case 0x01:
		c.setS(fd, c.fs(fs)-c.fs(ft))
	case 0x02:
		c.setS(fd, c.fs(fs)*c.fs(ft))
	case 0x03:
		c.setS(fd, c.fs(fs)/c.fs(ft))
	case 0x04:
		c.setS(fd, float32(math.Sqrt(float64(c.fs(fs)))))
	case 0x05:
		c.setS(fd, float32(math.Abs(float64(c.fs(fs)))))
	case 0x06:
		c.writeFGR32(fd, c.readFGR32(fs)) // mov.s moves bits, not a value
	case 0x07:
		c.setS(fd, -c.fs(fs))
	case 0x08: // round.l.s
		c.writeFGR64(fd, uint64(int64(c.round(float64(c.fs(fs))))))
	case 0x09: // trunc.l.s
		c.writeFGR64(fd, uint64(int64(math.Trunc(float64(c.fs(fs))))))
	case 0x0A: // ceil.l.s
		c.writeFGR64(fd, uint64(int64(math.Ceil(float64(c.fs(fs))))))
	case 0x0B: // floor.l.s
		c.writeFGR64(fd, uint64(int64(math.Floor(float64(c.fs(fs))))))
	case 0x0C: // round.w.s
		c.writeFGR32(fd, uint32(int32(c.round(float64(c.fs(fs))))))
	case 0x0D: // trunc.w.s
		c.writeFGR32(fd, uint32(int32(math.Trunc(float64(c.fs(fs))))))
	case 0x0E: // ceil.w.s
		c.writeFGR32(fd, uint32(int32(math.Ceil(float64(c.fs(fs))))))
	case 0x0F: // floor.w.s
		c.writeFGR32(fd, uint32(int32(math.Floor(float64(c.fs(fs))))))
	case 0x21: // cvt.d.s
		c.setD(fd, float64(c.fs(fs)))
	case 0x24: // cvt.w.s
		c.writeFGR32(fd, uint32(int32(c.round(float64(c.fs(fs))))))
	case 0x25: // cvt.l.s
		c.writeFGR64(fd, uint64(int64(c.round(float64(c.fs(fs))))))
	default:
		c.Halt("unimplemented cop1 s-format funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, uint32(c.curPC))
	}
}

func (c *CPU) cop1Double(w, funct, ft, fs, fd uint32) {
	if funct&0x30 == 0x30 { // c.cond.d
		c.compare(c.fd(fs), c.fd(ft), funct&0xF)
		return
	}
	switch funct {
	case 0x00:
		c.setD(fd, c.fd(fs)+c.fd(ft))
	case 0x01:
		c.setD(fd, c.fd(fs)-c.fd(ft))
	case 0x02:
		c.setD(fd, c.fd(fs)*c.fd(ft))
	case 0x03:
		c.setD(fd, c.fd(fs)/c.fd(ft))
	case 0x04:
		c.setD(fd, math.Sqrt(c.fd(fs)))
	case 0x05:
		c.setD(fd, math.Abs(c.fd(fs)))
	case 0x06:
		c.writeFGR64(fd, c.readFGR64(fs)) // mov.d
	case 0x07:
		c.setD(fd, -c.fd(fs))
	case 0x08: // round.l.d
		c.writeFGR64(fd, uint64(int64(c.round(c.fd(fs)))))
	case 0x09: // trunc.l.d
		c.writeFGR64(fd, uint64(int64(math.Trunc(c.fd(fs)))))
	case 0x0A: // ceil.l.d
		c.writeFGR64(fd, uint64(int64(math.Ceil(c.fd(fs)))))
	case 0x0B: // floor.l.d
		c.writeFGR64(fd, uint64(int64(math.Floor(c.fd(fs)))))
	case 0x0C: // round.w.d
		c.writeFGR32(fd, uint32(int32(c.round(c.fd(fs)))))
	case 0x0D: // trunc.w.d
		c.writeFGR32(fd, uint32(int32(math.Trunc(c.fd(fs)))))
	case 0x0E: // ceil.w.d
		c.writeFGR32(fd, uint32(int32(math.Ceil(c.fd(fs)))))
	case 0x0F: // floor.w.d
		c.writeFGR32(fd, uint32(int32(math.Floor(c.fd(fs)))))
	case 0x20: // cvt.s.d
		c.setS(fd, float32(c.fd(fs)))
	case 0x24: // cvt.w.d
		c.writeFGR32(fd, uint32(int32(c.round(c.fd(fs)))))
	case 0x25: // cvt.l.d
		c.writeFGR64(fd, uint64(int64(c.round(c.fd(fs)))))
	default:
		c.Halt("unimplemented cop1 d-format funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, uint32(c.curPC))
	}
}

// compare evaluates one of the sixteen FP conditions and writes the FCR31
// condition bit. The condition's low three bits select, in order, whether an
// unordered pair compares true, whether equality does, and whether less-than
// does; the fourth bit only selects whether a NaN signals, which is not
// modelled.
func (c *CPU) compare(a, b float64, cond uint32) {
	unordered := math.IsNaN(a) || math.IsNaN(b)
	var r bool
	if unordered {
		r = cond&1 != 0
	} else {
		r = (cond&4 != 0 && a < b) || (cond&2 != 0 && a == b)
	}
	if r {
		c.FCR31 |= fcr31Cond
	} else {
		c.FCR31 &^= fcr31Cond
	}
}
