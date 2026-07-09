package r4300

// fpu_flags.go implements the five IEEE-754 exception conditions the FPU raises,
// and the three parallel fields of FCR31 that report them.
//
// Those fields are easy to confuse. The *causes* describe the operation that
// just ran and are cleared before each one. The *flags* are sticky: a program
// polls them to learn whether anything went wrong since it last cleared them.
// The *enables* decide whether a cause traps instead of merely being recorded.
// And the rule that ties them together is the one worth stating: when an enabled
// cause traps, the sticky flags are *not* updated, because the handler is
// supposed to see the causes untouched.
//
// Detecting inexactness honestly requires knowing the exact result, which a
// float32 or float64 cannot hold. Each operation is therefore evaluated a second
// time in arbitrary precision, and the rounded answer compared against it. That
// is slow, and it is the only way to be right: a model that guesses "inexact if
// the operands had many significant bits" is wrong in both directions.

import "math"

func isNaN(v float64) bool { return math.IsNaN(v) }
func isInf(v float64) bool { return math.IsInf(v, 0) }
func nan() float64         { return math.NaN() }

func sign(v float64) float64 {
	if math.Signbit(v) {
		return -1
	}
	return 1
}

func inf(s float64) float64 {
	if s < 0 {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

func zero(s float64) float64 {
	if s < 0 {
		return math.Copysign(0, -1)
	}
	return 0
}

// The five conditions, in the bit order FCR31 uses within each of its fields,
// and above them the unimplemented-operation condition, which has a cause bit
// but no sticky flag and no enable — it always traps.
const (
	fpInexact   = 1 << 0
	fpUnderflow = 1 << 1
	fpOverflow  = 1 << 2
	fpDivByZero = 1 << 3
	fpInvalid   = 1 << 4
	fpUnimpl    = 1 << 5

	// fpTiny is not a hardware bit. It marks a result too small to represent, and
	// leaves the decision of what that means to fpApply, which is where the
	// enables and the FS bit can be consulted.
	fpTiny = 1 << 6
)

// isSNaN32 and isSNaN64 report whether a register holds a signalling NaN.
//
// The VR4300 predates IEEE-754-2008 and uses the opposite convention: a NaN
// whose mantissa's top bit is *set* is signalling, and one whose top bit is
// clear is quiet. Go — and every modern language — has it the other way round,
// so math.NaN() is a signalling NaN on this chip, and an operand of it raises
// invalid-operation rather than propagating silently.
//
// They take raw bits, and must. Widening a float32 NaN to a float64 quiets it,
// so a value round-tripped through the host's arithmetic answers the question
// about the host's NaN and not about the one in the register.
func isSNaN32(bits uint32) bool {
	return bits&0x7F800000 == 0x7F800000 && bits&0x007FFFFF != 0 && bits&0x00400000 != 0
}

func isSNaN64(bits uint64) bool {
	return bits&0x7FF0000000000000 == 0x7FF0000000000000 &&
		bits&0x000FFFFFFFFFFFFF != 0 && bits&(1<<51) != 0
}

// The three fields' positions in FCR31.
const (
	fcr31FlagShift   = 2
	fcr31EnableShift = 7
	fcr31CauseShift  = 12
)

// fpApply records a set of conditions and reports whether the operation trapped.
//
// An enabled condition raises the floating-point exception and leaves the sticky
// flags alone. Otherwise the causes are recorded and folded into the flags, and
// the operation's result stands.
func (c *CPU) fpApply(cond uint32) bool {
	enables := (c.FCR31 >> fcr31EnableShift) & 0x1F

	// A result too small to represent has three possible fates, and the chip
	// picks between them here rather than in the arithmetic.
	//
	// If either the underflow or the inexact trap is enabled, the operation is
	// declared unimplemented and handed to software — the hardware will not even
	// try to report an underflow it might have to trap on. Otherwise the FS bit
	// decides: set, the result flushes to a signed zero and underflow and inexact
	// are merely recorded; clear, it is unimplemented again.
	if cond&fpTiny != 0 {
		cond &^= fpTiny
		if enables&(fpUnderflow|fpInexact) != 0 || c.FCR31&fcr31FS == 0 {
			cond = fpUnimpl
		}
	}

	c.FCR31 = (c.FCR31 &^ fcr31CauseMask) | cond<<fcr31CauseShift

	// The unimplemented-operation condition has no enable to consult. It always
	// traps, and it never sets a sticky flag, because there is no flag for it.
	if cond&fpUnimpl != 0 {
		c.Exception(excFPE)
		return true
	}
	if cond&enables != 0 {
		c.Exception(excFPE)
		return true
	}
	c.FCR31 |= cond << fcr31FlagShift
	return false
}

// denormal reports whether v is a non-zero subnormal in the given format.
//
// The VR4300 has no hardware for denormalised numbers, and it treats the two
// places one can appear quite differently.
//
// A denormal *operand* it cannot compute with at all: the operation raises the
// unimplemented-operation exception, and an operating system is expected to
// finish the arithmetic in software.
//
// A denormal *result* it cannot represent either. What that means depends on the
// enables and on the FS bit, and fpApply decides it.
//
// An emulator that quietly computes with denormals is faster than the real chip
// and wrong in every position.
func denormal(v float64, single bool) bool {
	if single {
		return isSubnormal32(float32(v))
	}
	return isSubnormal64(v)
}

// --- exactness, without arbitrary precision ---------------------------------
//
// Knowing whether a result was rounded means knowing the exact result, which the
// target format cannot hold. Arbitrary-precision arithmetic would answer it, and
// would also make every floating-point instruction allocate — unacceptable in a
// core that runs hundreds of millions of them.
//
// Error-free transforms answer it exactly and in a few flops. The residual of a
// rounded operation is itself representable, and a fused multiply-add computes it
// without a second rounding: a*b - r is exact when r is the rounded product. If
// the residual is zero the operation was exact.
//
// Single precision does not escape needing the residual. A product of two floats
// does fit a double exactly — 24 bits by 24 is at most 48 — but a *sum* does
// not: two floats whose exponents differ by more than twenty-nine have an exact
// sum too wide for 53 bits, and the double addition rounds before we can look at
// it. So the double result and its residual together give the exact value, and
// the float is inexact if either the residual is non-zero or narrowing the
// double loses something. Double rounding through a float64 is harmless here
// because 2*24 + 2 <= 53.

// applyRounding moves a nearest-rounded result onto the value the current
// rounding mode names.
//
// dsign is the sign of (exact - rounded): +1 when the true answer lies above the
// rounded one, -1 when below, 0 when the operation was exact. That single bit of
// information is all four modes need, because a nearest-rounded result is never
// more than one representable step from the truth.
func applyRounding(rm uint32, rounded float64, dsign int, single bool) float64 {
	if dsign == 0 || rm == roundNearest {
		return rounded
	}
	up := func() float64 { return next(rounded, math.Inf(1), single) }
	down := func() float64 { return next(rounded, math.Inf(-1), single) }

	switch rm {
	case roundToZero:
		// Step inward only if nearest overshot the magnitude.
		if rounded > 0 && dsign < 0 {
			return down()
		}
		if rounded < 0 && dsign > 0 {
			return up()
		}
	case roundCeil:
		if dsign > 0 {
			return up()
		}
	case roundFloor:
		if dsign < 0 {
			return down()
		}
	}
	return rounded
}

// next is the adjacent representable value in the target format.
func next(v, toward float64, single bool) float64 {
	if single {
		return float64(math.Nextafter32(float32(v), float32(toward)))
	}
	return math.Nextafter(v, toward)
}

func signOf(v float64) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// classifySingle narrows a double result into a float and classifies the
// rounding. residual is the exact error of the double operation, so the pair
// (rounded, residual) names the exact value between them.
func classifySingle(rounded float64, inexact bool) (float64, uint32) {
	var cond uint32
	f := float32(rounded)
	if inexact || float64(f) != rounded {
		cond |= fpInexact
	}
	switch {
	case isInf(float64(f)) && !isInf(rounded):
		cond |= fpOverflow | fpInexact
	case isSubnormal32(f) || (f == 0 && rounded != 0):
		// Too small to represent; fpApply decides what that costs.
		return zero(sign(rounded)), fpUnderflow | fpInexact | fpTiny
	}
	return float64(f), cond
}

func isSubnormal32(f float32) bool {
	return f != 0 && math.Abs(float64(f)) < 1.1754943508222875e-38
}

func isSubnormal64(v float64) bool {
	return v != 0 && math.Abs(v) < 2.2250738585072014e-308
}

// classifyDouble uses the residual of the rounded operation to decide exactness.
func classifyDouble(r float64, inexact bool) (float64, uint32) {
	var cond uint32
	if inexact {
		cond |= fpInexact
	}
	switch {
	case isInf(r):
		cond |= fpOverflow | fpInexact
	case isSubnormal64(r) || (r == 0 && inexact):
		return zero(sign(r)), fpUnderflow | fpInexact | fpTiny
	}
	return r, cond
}

// fpArith evaluates op on a and b, rounds into the target format, and reports the
// conditions that rounding raised. single selects the format; the caller narrows
// the returned value.
func fpArith(rm uint32, op byte, a, b float64, single, snanIn bool) (float64, uint32) {
	// Invalid and divide-by-zero are decided by the operands, before any
	// arithmetic happens.
	if snanIn {
		return nan(), fpInvalid // signalling here means "mantissa top bit set"
	}
	if isNaN(a) || isNaN(b) {
		return nan(), 0 // a quiet NaN propagates without raising anything
	}
	// A denormal operand is not something this FPU can compute with.
	if denormal(a, single) || denormal(b, single) {
		return 0, fpUnimpl
	}
	switch op {
	case '/':
		if b == 0 {
			if a == 0 {
				return nan(), fpInvalid
			}
			return inf(sign(a) * sign(b)), fpDivByZero
		}
		if isInf(a) && isInf(b) {
			return nan(), fpInvalid
		}
	case '*':
		if (isInf(a) && b == 0) || (a == 0 && isInf(b)) {
			return nan(), fpInvalid
		}
	case '+':
		if isInf(a) && isInf(b) && sign(a) != sign(b) {
			return nan(), fpInvalid
		}
	case '-':
		if isInf(a) && isInf(b) && sign(a) == sign(b) {
			return nan(), fpInvalid
		}
	}
	if isInf(a) || isInf(b) {
		return infiniteResult(op, a, b), 0 // exact, whatever the other operand
	}

	if single {
		switch op {
		case '+', '-':
			if op == '-' {
				b = -b
			}
			s := a + b
			err := twoSumResidual(a, b, s)
			v, cond := classifySingle(s, err != 0)
			d := (s - v) + err // exact - the narrowed result
			return applyRounding(rm, v, signOf(d), true), cond
		case '*':
			// A product of two 24-bit significands is at most 48 bits: exact.
			p := a * b
			v, cond := classifySingle(p, false)
			return applyRounding(rm, v, signOf(p-v), true), cond
		}
		// Division: round through a double, then confirm with the residual, which
		// a fused multiply-add computes without a second rounding.
		q := a / b
		res := math.FMA(float64(float32(q)), b, -a)
		v, cond := classifySingle(q, res != 0)
		// exact - q has the sign of -(residual)/b.
		return applyRounding(rm, v, -signOf(res)*signOf(b), true), cond
	}

	switch op {
	case '+', '-':
		if op == '-' {
			b = -b
		}
		r := a + b
		err := twoSumResidual(a, b, r)
		v, cond := classifyDouble(r, err != 0)
		return applyRounding(rm, v, signOf(err), false), cond
	case '*':
		r := a * b
		if isInf(r) {
			return r, fpOverflow | fpInexact
		}
		err := math.FMA(a, b, -r)
		v, cond := classifyDouble(r, err != 0)
		return applyRounding(rm, v, signOf(err), false), cond
	default: // '/'
		r := a / b
		res := math.FMA(r, b, -a)
		v, cond := classifyDouble(r, res != 0)
		return applyRounding(rm, v, -signOf(res)*signOf(b), false), cond
	}
}

// twoSumResidual is the exact error of a rounded sum, by Knuth's algorithm.
func twoSumResidual(a, b, s float64) float64 {
	bb := s - a
	return (a - (s - bb)) + (b - bb)
}

// infiniteResult is the exact answer when at least one operand is infinite and
// the operation is not invalid.
func infiniteResult(op byte, a, b float64) float64 {
	switch op {
	case '+':
		if isInf(a) {
			return a
		}
		return b
	case '-':
		if isInf(a) {
			return a
		}
		return -b
	case '*':
		return inf(sign(a) * sign(b))
	default: // '/'
		if isInf(a) {
			return inf(sign(a) * sign(b))
		}
		return zero(sign(a) * sign(b))
	}
}

// sqrtConditions classifies a square root. The residual r*r - x is exact under a
// fused multiply-add, which is what makes the inexact test cheap and correct.
func sqrtConditions(r, x float64) uint32 {
	if x == 0 || isInf(x) || isNaN(x) {
		return 0
	}
	if math.FMA(r, r, -x) != 0 {
		return fpInexact
	}
	return 0
}
