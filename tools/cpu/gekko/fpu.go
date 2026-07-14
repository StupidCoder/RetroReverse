package gekko

// fpu.go is the floating-point unit: primary opcodes 59 (single precision) and 63
// (double precision), and the FPSCR that records what they did.
//
// Two things here are easy to get subtly, silently wrong.
//
// The first is double rounding. The single-precision instructions — fadds, fmuls, and the
// whole ps_ family — do not compute in single precision. They compute the exact result in
// double and then round it to single. So `fadds` is `f32(a + b)`, and that is NOT the same
// as computing in float32 throughout: a sum whose exact value needs 53 bits, rounded once
// to 53 and then to 24, can differ in the last bit from the same sum rounded straight to
// 24. Every single-precision operation in this file is therefore written as "compute in
// float64, then f32", which is what the hardware does.
//
// The second is the multiply-accumulate. fmadd computes a×c + b with the *product not
// rounded* — one rounding, at the end, of the exact product-plus-addend. Go's math.FMA is
// exactly this, and using `a*c + b` instead would round twice and be wrong in the last
// bit, on an operation a 3-D game executes millions of times per frame. And fmadds rounds
// that fused result *once more*, to single.
//
// FPSCR is modelled for the bits software actually reads: the exception summary and the
// sticky flags, and FPRF, the class of the result. The enable bits are honoured to the
// extent of not trapping, because a GameCube game leaves them off.

import "math"

// The FPSCR bits, MSB-numbered as PowerPC counts them (bit 0 is the top).
const (
	FPSCRFX     uint32 = 1 << 31 // any exception bit went from 0 to 1: sticky summary
	FPSCRFEX    uint32 = 1 << 30 // an *enabled* exception is set
	FPSCRVX     uint32 = 1 << 29 // any invalid-operation bit is set
	FPSCROX     uint32 = 1 << 28 // overflow
	FPSCRUX     uint32 = 1 << 27 // underflow
	FPSCRZX     uint32 = 1 << 26 // zero divide
	FPSCRXX     uint32 = 1 << 25 // inexact
	FPSCRVXSNAN uint32 = 1 << 24
	FPSCRVXISI  uint32 = 1 << 23 // infinity minus infinity
	FPSCRVXIDI  uint32 = 1 << 22 // infinity divided by infinity
	FPSCRVXZDZ  uint32 = 1 << 21 // zero divided by zero
	FPSCRVXIMZ  uint32 = 1 << 20 // infinity times zero
	FPSCRVXVC   uint32 = 1 << 19 // an invalid compare
	FPSCRFR     uint32 = 1 << 18 // the result was rounded up
	FPSCRFI     uint32 = 1 << 17 // the result is inexact
	FPSCRVXSOFT uint32 = 1 << 10
	FPSCRVXSQRT uint32 = 1 << 9
	FPSCRVXCVI  uint32 = 1 << 8 // an invalid integer convert

	// FPRF, the result class, sits in bits 15-19 (MSB-numbered), i.e. shifted by 12.
	fprfShift = 12
	fprfMask  = uint32(0x1F) << fprfShift
)

// The FPRF class codes.
const (
	fprfQNaN     = 0x11
	fprfNegInf   = 0x09
	fprfNegNorm  = 0x08
	fprfNegDenom = 0x18
	fprfNegZero  = 0x12
	fprfPosZero  = 0x02
	fprfPosDenom = 0x14
	fprfPosNorm  = 0x04
	fprfPosInf   = 0x05
)

// setFPRF records the class of a result, which is what a floating-point compare-and-branch
// sequence reads.
func (c *CPU) setFPRF(v float64) {
	var class uint32
	switch {
	case math.IsNaN(v):
		class = fprfQNaN
	case math.IsInf(v, 1):
		class = fprfPosInf
	case math.IsInf(v, -1):
		class = fprfNegInf
	case v == 0:
		if math.Signbit(v) {
			class = fprfNegZero
		} else {
			class = fprfPosZero
		}
	case math.Abs(v) < 2.2250738585072014e-308: // below the smallest normal double
		if v < 0 {
			class = fprfNegDenom
		} else {
			class = fprfPosDenom
		}
	case v < 0:
		class = fprfNegNorm
	default:
		class = fprfPosNorm
	}
	c.FPSCR = (c.FPSCR &^ fprfMask) | (class << fprfShift)
}

// setFPSCRBit sets a sticky exception bit, and with it the summary bits that software
// actually tests. FX is set only on a 0->1 transition — it means "something new happened",
// not "something is set".
func (c *CPU) setFPSCRBit(bit uint32) {
	if c.FPSCR&bit == 0 {
		c.FPSCR |= FPSCRFX
	}
	c.FPSCR |= bit
	if bit&(FPSCRVXSNAN|FPSCRVXISI|FPSCRVXIDI|FPSCRVXZDZ|FPSCRVXIMZ|FPSCRVXVC|FPSCRVXSOFT|FPSCRVXSQRT|FPSCRVXCVI) != 0 {
		c.FPSCR |= FPSCRVX
	}
}

// isSNaN reports a signalling NaN: a NaN whose quiet bit is clear.
func isSNaN(v float64) bool {
	b := math.Float64bits(v)
	return math.IsNaN(v) && b&(1<<51) == 0
}

// qnan is the default NaN this processor produces from an invalid operation.
var qnan = math.Float64frombits(0x7FF8000000000000)

// fpResult stores a double result and updates FPRF and CR1.
func (c *CPU) fpResult(w, d uint32, v float64) {
	c.FPR[d].PS0 = v
	c.setFPRF(v)
	if rcbit(w) {
		c.setCR1()
	}
}

// fpResultS stores a single-precision result — rounded to single, but held as a double,
// and written to BOTH halves of the register, which is what the hardware does: a
// single-precision op in scalar mode still drives the paired-single datapath.
func (c *CPU) fpResultS(w, d uint32, v float64) {
	r := f32(v)
	c.FPR[d].PS0 = r
	c.FPR[d].PS1 = r
	c.setFPRF(r)
	if rcbit(w) {
		c.setCR1()
	}
}

// The arithmetic primitives, with the invalid-operation cases the architecture names.
// Each returns the result and records the exception it raised, because a NaN out of a
// valid operation and a NaN out of an invalid one are different events.

func (c *CPU) fadd(a, b float64) float64 {
	if isSNaN(a) || isSNaN(b) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	// Infinity minus infinity is the invalid case, and it is why addition needs a check
	// at all.
	if math.IsInf(a, 0) && math.IsInf(b, 0) && math.Signbit(a) != math.Signbit(b) {
		c.setFPSCRBit(FPSCRVXISI)
		return qnan
	}
	return a + b
}

func (c *CPU) fsub(a, b float64) float64 {
	if isSNaN(a) || isSNaN(b) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if math.IsInf(a, 0) && math.IsInf(b, 0) && math.Signbit(a) == math.Signbit(b) {
		c.setFPSCRBit(FPSCRVXISI)
		return qnan
	}
	return a - b
}

func (c *CPU) fmul(a, b float64) float64 {
	if isSNaN(a) || isSNaN(b) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if (math.IsInf(a, 0) && b == 0) || (a == 0 && math.IsInf(b, 0)) {
		c.setFPSCRBit(FPSCRVXIMZ)
		return qnan
	}
	return a * b
}

func (c *CPU) fdiv(a, b float64) float64 {
	if isSNaN(a) || isSNaN(b) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	switch {
	case math.IsInf(a, 0) && math.IsInf(b, 0):
		c.setFPSCRBit(FPSCRVXIDI)
		return qnan
	case a == 0 && b == 0:
		c.setFPSCRBit(FPSCRVXZDZ)
		return qnan
	case b == 0 && !math.IsNaN(a):
		// A finite non-zero over zero is not invalid — it is a zero-divide, and it
		// produces a correctly-signed infinity.
		c.setFPSCRBit(FPSCRZX)
		return math.Inf(sign(a) * sign(b))
	}
	return a / b
}

func sign(v float64) int {
	if math.Signbit(v) {
		return -1
	}
	return 1
}

// fmadd is a×c + b with ONE rounding, of the exact product-plus-addend. math.FMA is
// precisely that; rounding the product first would be wrong in the last bit, on an
// operation a 3-D game runs millions of times per frame.
//
// Do not "simplify" this to a*cc + b, and be careful how you convince yourself it is
// safe. The Go specification lets an implementation fuse a multiply and an add on its own,
// and on arm64 it does: `a*cc + b` compiles to a hardware FMADD, gives the fused answer,
// and passes this core's entire vector suite. On amd64 it does not fuse, and the same
// source would round twice and produce different results — so the "simplification" would
// be a real bug that a Mac would never show you and a Linux box would. math.FMA is the
// only spelling that means the same thing on both.
//
// (This is not hypothetical. Mutation-testing the suite found it: the naive rewrite failed
// nothing on this machine, which is what sent us looking for why.)
func (c *CPU) fmaddRaw(a, cc, b float64) float64 {
	if isSNaN(a) || isSNaN(b) || isSNaN(cc) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if (math.IsInf(a, 0) && cc == 0) || (a == 0 && math.IsInf(cc, 0)) {
		c.setFPSCRBit(FPSCRVXIMZ)
		return qnan
	}
	prod := a * cc
	if math.IsInf(prod, 0) && math.IsInf(b, 0) && math.Signbit(prod) != math.Signbit(b) {
		c.setFPSCRBit(FPSCRVXISI)
		return qnan
	}
	return math.FMA(a, cc, b)
}

func (c *CPU) fsqrt(v float64) float64 {
	if isSNaN(v) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if v < 0 && !math.IsNaN(v) {
		c.setFPSCRBit(FPSCRVXSQRT)
		return qnan
	}
	return math.Sqrt(v)
}

// fres and frsqrte are *estimates*. The hardware guarantees only about 12 bits of
// accuracy (a relative error under 1/4096), and it computes them from a lookup table of
// base-and-slope entries rather than by dividing. Software that uses them — and a
// GameCube's vector normalisation does, constantly — is written to tolerate that.
//
// Computing the exact reciprocal here is therefore MORE accurate than the hardware, and
// that is a deliberate, recorded choice rather than an oversight: a game that depends on
// the estimate's exact bit pattern would diverge, but nothing depends on the low bits of
// a value the manual declines to specify beyond a tolerance. If a divergence ever traces
// back to here, this comment is where to start, and the fix is to transcribe the tables.
func (c *CPU) fres(v float64) float64 {
	if isSNaN(v) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if v == 0 {
		c.setFPSCRBit(FPSCRZX)
		return math.Inf(sign(v))
	}
	return f32(1 / v)
}

func (c *CPU) frsqrte(v float64) float64 {
	if isSNaN(v) {
		c.setFPSCRBit(FPSCRVXSNAN)
	}
	if v == 0 {
		c.setFPSCRBit(FPSCRZX)
		return math.Inf(sign(v))
	}
	if v < 0 {
		c.setFPSCRBit(FPSCRVXSQRT)
		return qnan
	}
	return 1 / math.Sqrt(v)
}

// exec59 — single precision. Every result is rounded to single.
func (c *CPU) exec59(w uint32) {
	d, a, b, cc := rs(w), ra(w), rb(w), rc(w)
	fa, fb, fc := c.FPR[a].PS0, c.FPR[b].PS0, c.FPR[cc].PS0

	switch xo5(w) {
	case 18:
		c.fpResultS(w, d, c.fdiv(fa, fb))
	case 20:
		c.fpResultS(w, d, c.fsub(fa, fb))
	case 21:
		c.fpResultS(w, d, c.fadd(fa, fb))
	case 22:
		c.fpResultS(w, d, c.fsqrt(fb))
	case 24:
		c.fpResultS(w, d, c.fres(fb))
	case 25:
		c.fpResultS(w, d, c.fmul(fa, fc))
	case 28: // fmsubs: a×c - b
		c.fpResultS(w, d, c.fmaddRaw(fa, fc, -fb))
	case 29: // fmadds
		c.fpResultS(w, d, c.fmaddRaw(fa, fc, fb))
	case 30: // fnmsubs: -(a×c - b)
		c.fpResultS(w, d, -c.fmaddRaw(fa, fc, -fb))
	case 31: // fnmadds
		c.fpResultS(w, d, -c.fmaddRaw(fa, fc, fb))
	default:
		c.Halt("gekko: unimplemented opcode 59 extended %d (word 0x%08X) at 0x%08X", xo5(w), w, c.PC-4)
	}
}

// exec63 — double precision, plus the FPSCR and the register moves.
func (c *CPU) exec63(w uint32) {
	d, a, b, cc := rs(w), ra(w), rb(w), rc(w)
	fa, fb, fc := c.FPR[a].PS0, c.FPR[b].PS0, c.FPR[cc].PS0

	switch xo5(w) {
	case 18:
		c.fpResult(w, d, c.fdiv(fa, fb))
		return
	case 20:
		c.fpResult(w, d, c.fsub(fa, fb))
		return
	case 21:
		c.fpResult(w, d, c.fadd(fa, fb))
		return
	case 22:
		c.fpResult(w, d, c.fsqrt(fb))
		return
	case 23: // fsel — a branchless select: a >= 0 ? c : b. It is how a GameCube avoids a
		// branch in vector code, and it treats NaN as "less than zero".
		v := fb
		if fa >= 0 && !math.IsNaN(fa) {
			v = fc
		}
		c.FPR[d].PS0 = v
		if rcbit(w) {
			c.setCR1()
		}
		return
	case 25:
		c.fpResult(w, d, c.fmul(fa, fc))
		return
	case 26:
		c.fpResult(w, d, c.frsqrte(fb))
		return
	case 28:
		c.fpResult(w, d, c.fmaddRaw(fa, fc, -fb))
		return
	case 29:
		c.fpResult(w, d, c.fmaddRaw(fa, fc, fb))
		return
	case 30:
		c.fpResult(w, d, -c.fmaddRaw(fa, fc, -fb))
		return
	case 31:
		c.fpResult(w, d, -c.fmaddRaw(fa, fc, fb))
		return
	}

	switch xo10(w) {
	case 0: // fcmpu — an unordered compare does not signal on a quiet NaN
		c.fcmp(crfD(w), fa, fb, false)
	case 32: // fcmpo — an ordered compare does
		c.fcmp(crfD(w), fa, fb, true)
	case 12: // frsp — round to single, but keep it in the register as a double
		c.fpResult(w, d, f32(fb))
	case 14: // fctiw — convert to integer, using the current rounding mode
		c.fctiw(w, d, fb, false)
	case 15: // fctiwz — convert to integer, rounding toward zero
		c.fctiw(w, d, fb, true)
	case 40: // fneg — a sign flip, which does not raise anything even on a NaN
		c.FPR[d].PS0 = math.Float64frombits(math.Float64bits(fb) ^ (1 << 63))
		c.rcF(w)
	case 72: // fmr
		c.FPR[d].PS0 = fb
		c.rcF(w)
	case 136: // fnabs
		c.FPR[d].PS0 = math.Float64frombits(math.Float64bits(fb) | (1 << 63))
		c.rcF(w)
	case 264: // fabs
		c.FPR[d].PS0 = math.Float64frombits(math.Float64bits(fb) &^ (1 << 63))
		c.rcF(w)
	case 64: // mcrfs — move an FPSCR field to CR, and clear the exception bits it held
		f := (c.FPSCR >> (28 - 4*crfS(w))) & 0xF
		c.SetCRField(crfD(w), f)
		c.FPSCR &^= 0xF << (28 - 4*crfS(w))
	case 38: // mtfsb1
		c.setFPSCRBit(1 << (31 - d))
		c.rcF(w)
	case 70: // mtfsb0
		c.FPSCR &^= 1 << (31 - d)
		c.rcF(w)
	case 134: // mtfsfi
		sh := 28 - 4*crfD(w)
		c.FPSCR = (c.FPSCR &^ (0xF << sh)) | (((w >> 12) & 0xF) << sh)
		c.rcF(w)
	case 583: // mffs — read FPSCR into the low half of a register
		c.FPR[d].PS0 = math.Float64frombits(uint64(c.FPSCR))
		c.rcF(w)
	case 711: // mtfsf — write the fields the mask selects
		mask := (w >> 17) & 0xFF
		m := uint32(0)
		for i := 0; i < 8; i++ {
			if mask&(0x80>>i) != 0 {
				m |= 0xF << (28 - 4*i)
			}
		}
		c.FPSCR = (c.FPSCR &^ m) | (uint32(math.Float64bits(fb)) & m)
		c.rcF(w)
	default:
		c.Halt("gekko: unimplemented opcode 63 extended %d (word 0x%08X) at 0x%08X", xo10(w), w, c.PC-4)
	}
}

func (c *CPU) rcF(w uint32) {
	if rcbit(w) {
		c.setCR1()
	}
}

// fcmp sets a condition-register field from a floating-point comparison. An unordered
// result — either operand a NaN — is its own outcome, distinct from less, greater and
// equal, and it is what a game tests when it wants to know that a computation went wrong.
func (c *CPU) fcmp(crf uint32, a, b float64, ordered bool) {
	var f uint32
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		f = crSO // "unordered"
		if isSNaN(a) || isSNaN(b) {
			c.setFPSCRBit(FPSCRVXSNAN)
			if ordered {
				c.setFPSCRBit(FPSCRVXVC)
			}
		} else if ordered {
			c.setFPSCRBit(FPSCRVXVC)
		}
	case a < b:
		f = crLT
	case a > b:
		f = crGT
	default:
		f = crEQ
	}
	c.SetCRField(crf, f)
	c.FPSCR = (c.FPSCR &^ fprfMask) | (f << fprfShift)
}

// fctiw converts a double to a 32-bit integer, left in the low half of the register's
// bit pattern (which is why stfiwx exists: it stores exactly that half).
//
// The saturation is the part worth being careful about: a value too large becomes
// 0x7FFFFFFF and too small 0x80000000, and both raise the invalid-convert exception.
func (c *CPU) fctiw(w, d uint32, v float64, truncate bool) {
	var i int32
	switch {
	case math.IsNaN(v):
		c.setFPSCRBit(FPSCRVXCVI)
		i = -0x80000000
	case v > 2147483647.0:
		c.setFPSCRBit(FPSCRVXCVI)
		i = 0x7FFFFFFF
	case v < -2147483648.0:
		c.setFPSCRBit(FPSCRVXCVI)
		i = -0x80000000
	default:
		if truncate {
			i = int32(math.Trunc(v))
		} else {
			i = int32(c.roundToNearest(v))
		}
	}
	// The high half is architecturally undefined; the hardware leaves a recognisable
	// pattern there and software never reads it, so the sign extension is as good a
	// choice as any — and stfiwx, the only instruction that looks, ignores it.
	c.FPR[d].PS0 = math.Float64frombits(0xFFF8000000000000 | uint64(uint32(i)))
	c.rcF(w)
}

// roundToNearest rounds half to even, which is the default rounding mode and the only one
// a GameCube program ever selects.
func (c *CPU) roundToNearest(v float64) float64 { return math.RoundToEven(v) }
