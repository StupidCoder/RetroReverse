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

// The VR4300's quiet NaN is not the one IEEE-754 suggests and Go produces. Every
// operation that must invent a NaN produces this exact pattern, and a program
// that compares bit patterns — as a conformance suite does — sees the difference.
// fcr31FS is the flush-denorm-to-zero bit, which decides what happens when a
// result is too small to represent.
const fcr31FS = 1 << 24

const (
	qNaN32 = 0x7FBFFFFF
	qNaN64 = 0x7FF7FFFFFFFFFFFF
)

// setSResult and setDResult write a computed value, substituting the hardware's
// quiet NaN for whatever pattern the host produced.
// snanCheckS and snanCheckD raise invalid-operation for a signalling NaN operand
// and write the hardware's quiet NaN. They report whether they handled it.
func (c *CPU) snanCheckS(fd, fs uint32) bool {
	if !isSNaN32(c.readFGR32(fs)) {
		return false
	}
	c.fpApply(fpInvalid)
	c.writeFGR32(fd, qNaN32)
	return true
}

// snanCheckD tests a *double* source. Reading the register as a single as well
// would misread the low half of an ordinary double as a NaN pattern, and every
// such value would be quietly replaced.
func (c *CPU) snanCheckD(fd, fs uint32) bool {
	if !isSNaN64(c.readFGR64(fs)) {
		return false
	}
	c.fpApply(fpInvalid)
	c.writeFGR64(fd, qNaN64)
	return true
}

// snanCheckSToD is cvt.d.s, whose source is a single and whose destination is a
// double. snanCheckDToS is cvt.s.d, the other way round.
//
// The source's format is the instruction's, not the destination's. Reading a
// double's register as a single — or the reverse — misreads half of an ordinary
// value as a NaN pattern, and every such value is then quietly replaced.
func (c *CPU) snanCheckSToD(fd, fs uint32) bool {
	if !isSNaN32(c.readFGR32(fs)) {
		return false
	}
	c.fpApply(fpInvalid)
	c.writeFGR64(fd, qNaN64)
	return true
}

func (c *CPU) snanCheckDToS(fd, fs uint32) bool {
	if !isSNaN64(c.readFGR64(fs)) {
		return false
	}
	c.fpApply(fpInvalid)
	c.writeFGR32(fd, qNaN32)
	return true
}

func (c *CPU) setSResult(i uint32, v float32) {
	if math.IsNaN(float64(v)) {
		c.writeFGR32(i, qNaN32)
		return
	}
	c.setS(i, v)
}

func (c *CPU) setDResult(i uint32, v float64) {
	if math.IsNaN(v) {
		c.writeFGR64(i, qNaN64)
		return
	}
	c.setD(i, v)
}

// FCR31 bits. The register carries three parallel five- or six-bit fields for
// the same five IEEE exceptions: the sticky flags a program polls, the enables
// that decide whether a condition traps, and the causes raised by the operation
// that just ran.
const (
	fcr31Cond        = 1 << 23    // the condition flag set by c.cond.fmt
	fcr31CauseUnimpl = 1 << 17    // an operation the FPU does not implement
	fcr31CauseMask   = 0x0003F000 // bits 12..17: the six cause bits
)

// clearCause empties the cause field, which every arithmetic operation does
// before it runs.
//
// The causes describe the operation that just executed, not the history of the
// program — that is what the sticky flags below them are for. A model that lets
// them accumulate reports an FPU permanently in every error state at once.
func (c *CPU) clearCause() { c.FCR31 &^= fcr31CauseMask }

// fpUnimplemented raises the floating-point exception a program gets for an
// encoding the FPU has no unit for — cvt.s.s, say, which names a conversion from
// a format to itself.
//
// The hardware does not simply ignore these. It traps, unconditionally, whatever
// the enable bits say, and records the reason in FCR31. n64-systemtest executes
// them on purpose and expects to land in its handler; halting here would stop a
// program that real hardware runs.
func (c *CPU) fpUnimplemented() { c.fpApply(fpUnimpl) }

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

func (c *CPU) fs(i uint32) float32      { return math.Float32frombits(c.readFGR32(i)) }
func (c *CPU) fd(i uint32) float64      { return math.Float64frombits(c.readFGR64(i)) }
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
		c.clearCause()
		switch funct {
		case 0x20: // cvt.s.w
			c.setSResult(fd, float32(int32(c.readFGR32(fs))))
		case 0x21: // cvt.d.w
			c.setD(fd, float64(int32(c.readFGR32(fs))))
		default:
			c.fpUnimplemented()
		}
	case 0x15: // 64-bit integer source
		c.clearCause()
		switch funct {
		case 0x20: // cvt.s.l
			c.setS(fd, float32(int64(c.readFGR64(fs))))
		case 0x21: // cvt.d.l
			c.setD(fd, float64(int64(c.readFGR64(fs))))
		default:
			c.fpUnimplemented()
		}
	default:
		c.fpUnimplemented()
	}
}

func (c *CPU) cop1Single(w, funct, ft, fs, fd uint32) {
	// MOV.fmt moves bits and computes nothing, so it leaves the cause field
	// alone. Every other operation reports on itself, and therefore clears the
	// causes of the one before it.
	if funct != 0x06 {
		c.clearCause()
	}
	if funct&0x30 == 0x30 { // c.cond.s
		c.compare(float64(c.fs(fs)), float64(c.fs(ft)), funct&0xF,
			isSNaN32(c.readFGR32(fs)) || isSNaN32(c.readFGR32(ft)))
		return
	}
	switch funct {
	case 0x00, 0x01, 0x02, 0x03:
		// The four arithmetic operations share their exception handling: the
		// exact result decides which conditions the rounding raised.
		snan := isSNaN32(c.readFGR32(fs)) || isSNaN32(c.readFGR32(ft))
		r, cond := fpArith(c.FCR31&3, "+-*/"[funct], float64(c.fs(fs)), float64(c.fs(ft)), true, snan)
		if c.fpApply(cond) {
			return // an enabled condition trapped; the destination is untouched
		}
		c.setSResult(fd, float32(r))
	case 0x04:
		x := float64(c.fs(fs))
		if x < 0 || isSNaN32(c.readFGR32(fs)) {
			c.fpApply(fpInvalid)
			c.setSResult(fd, float32(math.NaN()))
			break
		}
		r := math.Sqrt(x)
		// The residual r*r - x is exact under a fused multiply-add, so it says
		// whether the root was representable.
		if c.fpApply(sqrtConditions(r, x)) {
			break // a trapping operation leaves its destination alone
		}
		c.setSResult(fd, float32(r))
	case 0x05:
		// Absolute value and negation raise invalid on a signalling NaN, and in
		// any case yield the hardware's quiet NaN rather than the operand with
		// its sign bit rewritten.
		if c.snanCheckS(fd, fs) {
			break
		}
		c.setSResult(fd, float32(math.Abs(float64(c.fs(fs)))))
	case 0x06:
		c.writeFGR32(fd, c.readFGR32(fs)) // mov.s moves bits, not a value
	case 0x07:
		if c.snanCheckS(fd, fs) {
			break
		}
		c.setSResult(fd, -c.fs(fs))
	case 0x08: // round.l.s
		c.writeFGR64(fd, uint64(int64(c.round(float64(c.fs(fs))))))
	case 0x09: // trunc.l.s
		c.writeFGR64(fd, uint64(int64(math.Trunc(float64(c.fs(fs))))))
	case 0x0A: // ceil.l.s
		c.writeFGR64(fd, uint64(int64(math.Ceil(float64(c.fs(fs))))))
	case 0x0B: // floor.l.s
		c.writeFGR64(fd, uint64(int64(math.Floor(float64(c.fs(fs))))))
	case 0x0C: // round.w.s
		c.toInt32(fd, float64(c.fs(fs)), math.RoundToEven)
	case 0x0D: // trunc.w.s
		c.toInt32(fd, float64(c.fs(fs)), math.Trunc)
	case 0x0E: // ceil.w.s
		c.toInt32(fd, float64(c.fs(fs)), math.Ceil)
	case 0x0F: // floor.w.s
		c.toInt32(fd, float64(c.fs(fs)), math.Floor)
	case 0x21: // cvt.d.s
		if c.snanCheckSToD(fd, fs) {
			break
		}
		c.setDResult(fd, float64(c.fs(fs)))
	case 0x24: // cvt.w.s
		c.toInt32(fd, float64(c.fs(fs)), c.round)
	case 0x25: // cvt.l.s
		c.toInt64(fd, float64(c.fs(fs)), c.round)
	default:
		c.fpUnimplemented()
	}
}

func (c *CPU) cop1Double(w, funct, ft, fs, fd uint32) {
	if funct != 0x06 { // MOV.d computes nothing; see cop1Single
		c.clearCause()
	}
	if funct&0x30 == 0x30 { // c.cond.d
		c.compare(c.fd(fs), c.fd(ft), funct&0xF,
			isSNaN64(c.readFGR64(fs)) || isSNaN64(c.readFGR64(ft)))
		return
	}
	switch funct {
	case 0x00, 0x01, 0x02, 0x03:
		snan := isSNaN64(c.readFGR64(fs)) || isSNaN64(c.readFGR64(ft))
		r, cond := fpArith(c.FCR31&3, "+-*/"[funct], c.fd(fs), c.fd(ft), false, snan)
		if c.fpApply(cond) {
			return
		}
		c.setDResult(fd, r)
	case 0x04:
		x := c.fd(fs)
		if x < 0 || isSNaN64(c.readFGR64(fs)) {
			c.fpApply(fpInvalid)
			c.setDResult(fd, math.NaN())
			break
		}
		r := math.Sqrt(x)
		if c.fpApply(sqrtConditions(r, x)) {
			break
		}
		c.setDResult(fd, r)
	case 0x05:
		if c.snanCheckD(fd, fs) {
			break
		}
		c.setDResult(fd, math.Abs(c.fd(fs)))
	case 0x06:
		c.writeFGR64(fd, c.readFGR64(fs)) // mov.d
	case 0x07:
		if c.snanCheckD(fd, fs) {
			break
		}
		c.setDResult(fd, -c.fd(fs))
	case 0x08: // round.l.d
		c.writeFGR64(fd, uint64(int64(c.round(c.fd(fs)))))
	case 0x09: // trunc.l.d
		c.writeFGR64(fd, uint64(int64(math.Trunc(c.fd(fs)))))
	case 0x0A: // ceil.l.d
		c.writeFGR64(fd, uint64(int64(math.Ceil(c.fd(fs)))))
	case 0x0B: // floor.l.d
		c.writeFGR64(fd, uint64(int64(math.Floor(c.fd(fs)))))
	case 0x0C: // round.w.d
		c.toInt32(fd, c.fd(fs), math.RoundToEven)
	case 0x0D: // trunc.w.d
		c.toInt32(fd, c.fd(fs), math.Trunc)
	case 0x0E: // ceil.w.d
		c.toInt32(fd, c.fd(fs), math.Ceil)
	case 0x0F: // floor.w.d
		c.toInt32(fd, c.fd(fs), math.Floor)
	case 0x20: // cvt.s.d
		if c.snanCheckDToS(fd, fs) {
			break
		}
		// Narrowing loses bits whenever the double is not representable.
		v := c.fd(fs)
		if float64(float32(v)) != v {
			c.fpApply(fpInexact)
		}
		c.setSResult(fd, float32(v))
	case 0x24: // cvt.w.d
		c.toInt32(fd, c.fd(fs), c.round)
	case 0x25: // cvt.l.d
		c.toInt64(fd, c.fd(fs), c.round)
	default:
		c.fpUnimplemented()
	}
}

// compare evaluates one of the sixteen FP conditions and writes the FCR31
// condition bit.
//
// The condition's low three bits select, in order, whether an unordered pair
// compares true, whether equality does, and whether less-than does. The fourth
// bit picks the predicate's "signalling" half: those eight raise invalid on *any*
// NaN operand, while the other eight raise it only on a signalling one.
func (c *CPU) compare(a, b float64, cond uint32, snanIn bool) {
	unordered := math.IsNaN(a) || math.IsNaN(b)

	if unordered {
		signalling := cond&8 != 0 || snanIn
		if signalling && c.fpApply(fpInvalid) {
			return // the trap leaves the condition bit alone
		}
		if !signalling {
			c.clearCause()
		}
	}

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

// toInt32 and toInt64 convert a float to an integer with the given rounding, and
// report inexact when the value had a fractional part. A trapping conversion
// leaves its destination alone.
func (c *CPU) toInt32(fd uint32, v float64, round func(float64) float64) {
	r := round(v)
	var cond uint32
	if r != v {
		cond |= fpInexact
	}
	if c.fpApply(cond) {
		return
	}
	c.writeFGR32(fd, uint32(int32(r)))
}

func (c *CPU) toInt64(fd uint32, v float64, round func(float64) float64) {
	r := round(v)
	var cond uint32
	if r != v {
		cond |= fpInexact
	}
	if c.fpApply(cond) {
		return
	}
	c.writeFGR64(fd, uint64(int64(r)))
}
