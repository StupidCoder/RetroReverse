package r5900

// fpu.go is COP1, the Emotion Engine's floating-point unit.
//
// It is single-precision only — there is no double format at all — and it is
// deliberately *not* IEEE 754. Three departures matter, and code compiled for this
// machine depends on all three:
//
//   - There are no infinities and no NaNs. The exponent field value 255, which
//     IEEE reserves for them, is an ordinary exponent here, so the representable
//     range runs up to about 3.4e38 with a normal mantissa. An operation that
//     would overflow *saturates* to that maximum rather than producing an
//     infinity, and division by zero saturates too.
//   - There are no denormals. A source operand with a zero exponent reads as zero
//     (keeping its sign), and a result that would be denormal is flushed to zero.
//   - Rounding is always toward zero. There is no rounding mode to select.
//
// Because Go's float32 *is* IEEE, none of this can be had by simply using it: an
// overflow would give +Inf where the hardware gives 3.4e38, and every arithmetic
// result would be rounded to nearest where the hardware chops. So values are
// decoded into a float64 (wide enough to hold the whole R5900 range exactly),
// operated on there, and re-encoded with an explicit truncation. fenc is where the
// saturation and the chop live.
//
// The accumulator, ACC, is a fourth single-precision register that the multiply-add
// family reads and writes. adda/suba/mula load it; madd/msub read it and write a
// general FPR; madda/msuba read it and write it back.

import "math"

// fcr31C is the condition bit the c.cond.s comparisons write and bc1t/bc1f test.
const fcr31C = 1 << 23

// fMax is the largest finite magnitude the unit can hold: exponent field 255 with
// a full mantissa. On an IEEE machine this bit pattern would be a NaN.
const fMax = 0x7FFFFFFF

// fdec decodes an R5900 single into a float64, exactly. Zero exponent means zero
// (denormals do not exist), and exponent 255 is an ordinary large value rather
// than an infinity — which is why this cannot be math.Float32frombits.
func fdec(b uint32) float64 {
	exp := int((b >> 23) & 0xFF)
	man := b & 0x7FFFFF
	sign := 1.0
	if b>>31 != 0 {
		sign = -1.0
	}
	if exp == 0 {
		return sign * 0 // a denormal, or a zero, reads as zero
	}
	return sign * math.Ldexp(1+float64(man)/(1<<23), exp-127)
}

// fenc encodes a float64 back into an R5900 single: truncate toward zero, flush a
// denormal result to zero, and saturate an overflow to fMax instead of producing
// an infinity.
func fenc(v float64) uint32 {
	var sign uint32
	if math.Signbit(v) {
		sign = 1 << 31
	}
	if math.IsNaN(v) {
		// A NaN cannot be represented, and cannot be produced by the hardware
		// either. It can only arise here from an operation Go defines and the
		// R5900 does not (0/0 is handled at the call site), so it becomes zero.
		return sign
	}
	a := math.Abs(v)
	if math.IsInf(a, 0) {
		return sign | fMax
	}
	if a == 0 {
		return sign
	}

	frac, exp := math.Frexp(a) // a == frac * 2^exp, frac in [0.5, 1)
	e := exp - 1               // a == (1 + m) * 2^e, with m = frac*2 - 1
	if e > 128 {
		return sign | fMax // overflow saturates; it does not become an infinity
	}
	if e < -127 {
		return sign // underflow flushes to zero; there are no denormals
	}
	man := uint32((frac*2 - 1) * (1 << 23)) // truncation toward zero: the rounding mode
	return sign | uint32(e+127)<<23 | man&0x7FFFFF
}

// fdiv is division with the unit's saturating answer for a zero divisor.
func fdiv(a, b float64) float64 {
	if b == 0 {
		// The result saturates, with the sign the quotient would have had. 0/0
		// saturates positive, matching the sign rule applied to a zero numerator.
		s := 1.0
		if math.Signbit(a) != math.Signbit(b) {
			s = -1.0
		}
		return s * fdec(fMax)
	}
	return a / b
}

func (c *CPU) f(i uint32) float64       { return fdec(c.FPR[i]) }
func (c *CPU) setF(i uint32, v float64) { c.FPR[i] = fenc(v) }

// cop1 executes an op == 0x11 instruction.
func (c *CPU) cop1(w, rs, rt, rd, shamt uint32, branchT uint64) {
	switch rs {
	case 0x00: // mfc1
		c.set(rt, sext32(c.FPR[rd]))
	case 0x02: // cfc1 — only FCR0 (the implementation id) and FCR31 exist
		if rd == 31 {
			c.set(rt, sext32(c.FCR31))
		} else {
			c.set(rt, sext32(0x00002E00)) // FCR0: the unit's revision
		}
	case 0x04: // mtc1
		c.FPR[rd] = uint32(c.R[rt].Lo)
	case 0x06: // ctc1
		if rd == 31 {
			c.FCR31 = uint32(c.R[rt].Lo)
		}

	case 0x08: // bc1f / bc1t / bc1fl / bc1tl
		cond := c.FCR31&fcr31C != 0
		switch rt & 3 {
		case 0:
			c.doBranch(!cond, branchT)
		case 1:
			c.doBranch(cond, branchT)
		case 2:
			c.doBranchLikely(!cond, branchT)
		case 3:
			c.doBranchLikely(cond, branchT)
		}

	case 0x10: // single-precision arithmetic
		c.cop1Single(w, rt, rd, shamt)

	case 0x14: // cvt.s.w — an integer in an FPR, converted to a float
		if w&0x3F == 0x20 {
			c.setF(shamt, float64(int32(c.FPR[rd])))
			return
		}
		c.Exception(excRI)

	default:
		c.Exception(excRI)
	}
}

// cop1Single is the format-0x10 arithmetic: ft, fs and fd name FPRs.
func (c *CPU) cop1Single(w, ft, fs, fd uint32) {
	s, t := c.f(fs), c.f(ft)

	switch w & 0x3F {
	case 0x00: // add.s
		c.setF(fd, s+t)
	case 0x01: // sub.s
		c.setF(fd, s-t)
	case 0x02: // mul.s
		c.setF(fd, s*t)
	case 0x03: // div.s
		c.setF(fd, fdiv(s, t))
	case 0x04: // sqrt.s — of the magnitude: there is no NaN to return
		c.setF(fd, math.Sqrt(math.Abs(t)))
	case 0x05: // abs.s
		c.setF(fd, math.Abs(s))
	case 0x06: // mov.s — a bit copy, not an arithmetic move
		c.FPR[fd] = c.FPR[fs]
	case 0x07: // neg.s — likewise: flip the sign bit and touch nothing else
		c.FPR[fd] = c.FPR[fs] ^ 0x80000000
	case 0x16: // rsqrt.s
		c.setF(fd, fdiv(s, math.Sqrt(math.Abs(t))))

	case 0x18: // adda.s — the accumulator forms
		c.ACC = fenc(s + t)
	case 0x19: // suba.s
		c.ACC = fenc(s - t)
	case 0x1A: // mula.s
		c.ACC = fenc(s * t)
	case 0x1C: // madd.s — fd = ACC + fs*ft, leaving ACC alone
		c.setF(fd, fdec(c.ACC)+s*t)
	case 0x1D: // msub.s
		c.setF(fd, fdec(c.ACC)-s*t)
	case 0x1E: // madda.s — the same, but back into ACC
		c.ACC = fenc(fdec(c.ACC) + s*t)
	case 0x1F: // msuba.s
		c.ACC = fenc(fdec(c.ACC) - s*t)

	case 0x24: // cvt.w.s — truncate toward zero, saturating at the int32 ends
		c.FPR[fd] = cvtW(s)

	case 0x28: // max.s
		c.setF(fd, math.Max(s, t))
	case 0x29: // min.s
		c.setF(fd, math.Min(s, t))

	// The comparisons write the condition bit rather than a register. There is no
	// unordered case to consider, because there are no NaNs.
	case 0x30: // c.f.s — always false
		c.setCond(false)
	case 0x32: // c.eq.s
		c.setCond(s == t)
	case 0x34: // c.lt.s
		c.setCond(s < t)
	case 0x36: // c.le.s
		c.setCond(s <= t)

	default:
		c.Exception(excRI)
	}
}

func (c *CPU) setCond(b bool) {
	if b {
		c.FCR31 |= fcr31C
	} else {
		c.FCR31 &^= fcr31C
	}
}

// cvtW converts a float to a 32-bit integer, truncating toward zero. A magnitude
// past the integer range saturates rather than wrapping.
func cvtW(v float64) uint32 {
	t := math.Trunc(v)
	if t >= 2147483647 {
		return 0x7FFFFFFF
	}
	if t <= -2147483648 {
		return 0x80000000
	}
	return uint32(int32(t))
}

// cop1Mem is lwc1 / swc1. The FPU registers are 32-bit, so there is no ldc1/sdc1.
func (c *CPU) cop1Mem(op, rs, rt uint32, simm uint64) {
	vaddr := c.R[rs].Lo + simm
	if vaddr&3 != 0 {
		code := uint32(excAdEL)
		if op == 0x39 {
			code = excAdES
		}
		c.addrError(code, vaddr)
		return
	}
	switch op {
	case 0x31: // lwc1
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		c.FPR[rt] = c.read32(p)
	case 0x39: // swc1
		p, ok := c.Translate(vaddr, true)
		if !ok {
			return
		}
		c.write32(p, c.FPR[rt])
	}
}
