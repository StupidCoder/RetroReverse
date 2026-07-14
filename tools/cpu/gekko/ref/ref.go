// Package ref is a second, deliberately slow implementation of the Gekko's arithmetic,
// written to be *independent* of the interpreter in tools/cpu/gekko rather than fast.
//
// It exists to answer a question the interpreter cannot answer about itself. There is no
// published per-instruction conformance suite for PowerPC — nothing like the
// SingleStepTests data that validates the 8088 and the R3000 cores in this repository —
// so the expected results of a test suite have to come from somewhere, and if they come
// from the interpreter then the suite proves only that the interpreter agrees with
// itself.
//
// So this package derives the same results a second way. Where the interpreter computes
// a carry by adding two uint32s and watching them wrap, this computes it by adding two
// math/big integers and asking whether the sum needs a 33rd bit. Where the interpreter
// gets a signed overflow from an XOR of sign bits, this compares a 64-bit exact product
// against its 32-bit truncation. The two derivations share no code and no idiom; where
// they agree, the agreement means something.
//
// The precedent is in-tree, twice over: tools/cpu/r5900/diff_test.go validates a core
// with no published suite against a second implementation, and tools/cpu/r4300's
// fpu_flags.go re-derives each floating-point result in arbitrary precision because
// "detecting inexactness honestly requires knowing the exact result".
//
// This package MUST NOT import tools/cpu/gekko. TestRefIsIndependent enforces it by
// asking the go tool, rather than by trusting this comment.
package ref

import (
	"math"
	"math/big"
)

// --- The integer arithmetic, derived in arbitrary precision -------------------------

var (
	two32 = new(big.Int).Lsh(big.NewInt(1), 32)
	// The signed range a 32-bit result must fall in for there to be no overflow.
	minS32 = big.NewInt(-2147483648)
	maxS32 = big.NewInt(2147483647)
)

// u returns a as an arbitrary-precision unsigned value.
func u(a uint32) *big.Int { return new(big.Int).SetUint64(uint64(a)) }

// s returns a as an arbitrary-precision *signed* value.
func s(a uint32) *big.Int { return big.NewInt(int64(int32(a))) }

// Result is what an arithmetic instruction produces: a value, and the flags the
// architecture says it sets.
type Result struct {
	Value uint32
	CA    bool // the carry out, as XER[CA] holds it
	OV    bool // a signed overflow, as XER[OV] holds it
}

// Add computes a + b + carryIn, and derives both flags from exact arithmetic.
//
// The carry is "did the unsigned sum need a 33rd bit" — asked of a big integer, not
// inferred from a wraparound. The overflow is "did the signed sum leave the 32-bit signed
// range" — asked of the exact signed sum, not inferred from sign bits.
func Add(a, b uint32, carryIn bool) Result {
	ci := uint64(0)
	if carryIn {
		ci = 1
	}

	// Unsigned, exactly.
	sum := new(big.Int).Add(u(a), u(b))
	sum.Add(sum, new(big.Int).SetUint64(ci))
	ca := sum.Cmp(two32) >= 0

	// Signed, exactly.
	ssum := new(big.Int).Add(s(a), s(b))
	ssum.Add(ssum, new(big.Int).SetUint64(ci))
	ov := ssum.Cmp(minS32) < 0 || ssum.Cmp(maxS32) > 0

	// The stored value is the low 32 bits of either — they agree.
	val := uint32(new(big.Int).Mod(sum, two32).Uint64())
	return Result{Value: val, CA: ca, OV: ov}
}

// Sub computes b - a, which PowerPC defines as b + ¬a + 1. Deriving it that way rather
// than as a subtraction is the point: it is how the carry comes out meaning "no borrow",
// which is the opposite of most architectures' flag and the reason this is worth a
// second derivation.
func Sub(a, b uint32, carryIn bool) Result {
	return Add(^a, b, carryIn)
}

// Neg computes -a. The single overflow case is the most negative integer, which has no
// positive counterpart.
func Neg(a uint32) Result {
	n := new(big.Int).Neg(s(a))
	ov := n.Cmp(maxS32) > 0
	val := uint32(new(big.Int).Mod(n, two32).Uint64())
	return Result{Value: val, OV: ov}
}

// MulLow is the low 32 bits of a signed product, and overflow is "the exact product did
// not fit in 32 bits" — computed from the exact product, not from a sign heuristic.
func MulLow(a, b uint32) Result {
	p := new(big.Int).Mul(s(a), s(b))
	ov := p.Cmp(minS32) < 0 || p.Cmp(maxS32) > 0
	val := uint32(new(big.Int).Mod(p, two32).Uint64())
	return Result{Value: val, OV: ov}
}

// MulHighSigned and MulHighUnsigned are the top 32 bits of the 64-bit product.
func MulHighSigned(a, b uint32) uint32 {
	p := new(big.Int).Mul(s(a), s(b))
	p.Mod(p, new(big.Int).Lsh(big.NewInt(1), 64))
	return uint32(new(big.Int).Rsh(p, 32).Uint64())
}

func MulHighUnsigned(a, b uint32) uint32 {
	p := new(big.Int).Mul(u(a), u(b))
	return uint32(new(big.Int).Rsh(p, 32).Uint64())
}

// DivSigned divides. ok is false for the two cases the architecture leaves the *result*
// undefined — a zero divisor, and the most negative integer over minus one — and in both
// it still reports the overflow, because the architecture does define that.
func DivSigned(a, b uint32) (r Result, ok bool) {
	if b == 0 || (a == 0x80000000 && b == 0xFFFFFFFF) {
		return Result{OV: true}, false
	}
	q := new(big.Int).Quo(s(a), s(b)) // Quo truncates toward zero, as PowerPC does
	return Result{Value: uint32(new(big.Int).Mod(q, two32).Uint64())}, true
}

func DivUnsigned(a, b uint32) (r Result, ok bool) {
	if b == 0 {
		return Result{OV: true}, false
	}
	q := new(big.Int).Quo(u(a), u(b))
	return Result{Value: uint32(q.Uint64())}, true
}

// --- Shifts and rotates, derived bit by bit -----------------------------------------
//
// These are done as strings of bits rather than with shift operators, because the whole
// point is to not repeat the interpreter's reasoning. A rotate is a rotation of a
// 32-element array; a mask is the set of positions between two indices, wrapping if the
// first is greater; and the arithmetic shift's carry is the presence of a one among the
// bits that leave.

// bitsOf explodes a word into 32 bits, most significant first — the order PowerPC counts
// them in, so that mb and me index this array directly.
func bitsOf(v uint32) [32]byte {
	var b [32]byte
	for i := 0; i < 32; i++ {
		b[i] = byte((v >> (31 - i)) & 1)
	}
	return b
}

func wordOf(b [32]byte) uint32 {
	var v uint32
	for i := 0; i < 32; i++ {
		v |= uint32(b[i]) << (31 - i)
	}
	return v
}

// RotL rotates left by n, as a rotation of the bit array.
func RotL(v uint32, n uint32) uint32 {
	in := bitsOf(v)
	var out [32]byte
	for i := 0; i < 32; i++ {
		out[i] = in[(uint32(i)+n)%32]
	}
	return wordOf(out)
}

// Mask is PowerPC's MASK(mb, me): the bits from mb to me inclusive, MSB-numbered, and
// WRAPPING when mb > me. The wrap is the case a transcription gets wrong, so it is
// expressed here as "walk from mb to me the long way round" rather than as two shifts
// OR'd together.
func Mask(mb, me uint32) uint32 {
	var b [32]byte
	i := mb
	for {
		b[i] = 1
		if i == me {
			break
		}
		i = (i + 1) % 32
	}
	return wordOf(b)
}

// Rlwinm is rotate-left-then-mask.
func Rlwinm(v, sh, mb, me uint32) uint32 { return RotL(v, sh) & Mask(mb, me) }

// Rlwimi inserts the rotated value under the mask, keeping the rest of the destination.
func Rlwimi(dst, v, sh, mb, me uint32) uint32 {
	m := Mask(mb, me)
	return (RotL(v, sh) & m) | (dst &^ m)
}

// Srawi is the arithmetic right shift, and returns the carry the architecture defines:
// set if and only if the value is negative AND at least one one-bit left the register.
// The bits that leave are enumerated, rather than tested with a mask, so that a shift of
// zero (no bits leave) and a shift past the width (all of them do) are both simply the
// truth about the array.
func Srawi(v uint32, sh uint32) (uint32, bool) {
	in := bitsOf(v)
	negative := in[0] == 1

	if sh >= 32 {
		// The whole word leaves. The result is all sign bits; the carry is set if any
		// one-bit was among the departing ones, which for a negative value it always is.
		lost := false
		for i := 0; i < 32; i++ {
			if in[i] == 1 {
				lost = true
			}
		}
		var out [32]byte
		if negative {
			for i := range out {
				out[i] = 1
			}
		}
		return wordOf(out), negative && lost
	}

	// The bits that fall off the bottom.
	lost := false
	for i := 32 - int(sh); i < 32; i++ {
		if in[i] == 1 {
			lost = true
		}
	}
	// The result: shifted right, with the sign replicated into the top.
	var out [32]byte
	for i := 0; i < 32; i++ {
		src := i - int(sh)
		if src < 0 {
			out[i] = in[0] // the sign bit
		} else {
			out[i] = in[src]
		}
	}
	return wordOf(out), negative && lost
}

// Slw and Srw shift by a SIX-bit amount, so 32 or more clears the register entirely.
func Slw(v, sh uint32) uint32 {
	if sh >= 32 {
		return 0
	}
	in := bitsOf(v)
	var out [32]byte
	for i := 0; i < 32; i++ {
		if src := i + int(sh); src < 32 {
			out[i] = in[src]
		}
	}
	return wordOf(out)
}

func Srw(v, sh uint32) uint32 {
	if sh >= 32 {
		return 0
	}
	in := bitsOf(v)
	var out [32]byte
	for i := 0; i < 32; i++ {
		if src := i - int(sh); src >= 0 {
			out[i] = in[src]
		}
	}
	return wordOf(out)
}

// CountLeadingZeros counts them by looking.
func CountLeadingZeros(v uint32) uint32 {
	b := bitsOf(v)
	for i := 0; i < 32; i++ {
		if b[i] == 1 {
			return uint32(i)
		}
	}
	return 32
}

// --- Compares ------------------------------------------------------------------------

// The condition-register field bits, in PowerPC's order.
const (
	LT = 8
	GT = 4
	EQ = 2
	SO = 1
)

// CompareSigned and CompareUnsigned give the three-way result as a CR field, without the
// summary bit (which the caller ORs in from XER).
func CompareSigned(a, b uint32) uint32 {
	switch s(a).Cmp(s(b)) {
	case -1:
		return LT
	case 1:
		return GT
	}
	return EQ
}

func CompareUnsigned(a, b uint32) uint32 {
	switch u(a).Cmp(u(b)) {
	case -1:
		return LT
	case 1:
		return GT
	}
	return EQ
}

// --- The quantiser -------------------------------------------------------------------
//
// The densest bug farm in the core, so it gets the most independent derivation: the scale
// is applied by explicit multiplication and division by a power of two computed as a
// big.Float, rather than with math.Ldexp, and the saturation limits are written out as
// the literals the formats actually have.

// The quantisation format codes.
const (
	QFloat = 0
	QU8    = 4
	QU16   = 5
	QS8    = 6
	QS16   = 7
)

// Limits gives a format's saturation bounds. A float has none.
func Limits(typ uint32) (lo, hi float64, ok bool) {
	switch typ {
	case QU8:
		return 0, 255, true
	case QU16:
		return 0, 65535, true
	case QS8:
		return -128, 127, true
	case QS16:
		return -32768, 32767, true
	}
	return 0, 0, false
}

// Size is a format's width in bytes.
func Size(typ uint32) int {
	switch typ {
	case QU8, QS8:
		return 1
	case QU16, QS16:
		return 2
	}
	return 4
}

// pow2 computes 2^n exactly, for n in the signed six-bit range the GQR's scale field
// holds. It is built by repeated multiplication of a big.Float so that the exponent is
// not taken on faith from a library call.
func pow2(n int32) float64 {
	f := big.NewFloat(1).SetPrec(200)
	two := big.NewFloat(2).SetPrec(200)
	half := big.NewFloat(0.5).SetPrec(200)
	if n >= 0 {
		for i := int32(0); i < n; i++ {
			f.Mul(f, two)
		}
	} else {
		for i := int32(0); i > n; i-- {
			f.Mul(f, half)
		}
	}
	v, _ := f.Float64()
	return v
}

// Dequantize turns one stored value into a float: the load direction, which DIVIDES by
// the scale (equivalently, multiplies by 2^-scale).
func Dequantize(raw uint32, typ uint32, scale int32) float64 {
	var v float64
	switch typ {
	case QFloat:
		return float64(math.Float32frombits(raw))
	case QU8:
		v = float64(uint8(raw))
	case QU16:
		v = float64(uint16(raw))
	case QS8:
		v = float64(int8(raw))
	case QS16:
		v = float64(int16(raw))
	}
	return v * pow2(-scale)
}

// Quantize turns a float into a stored value: the store direction, which MULTIPLIES by
// the scale, then truncates, then saturates.
func Quantize(v float64, typ uint32, scale int32) uint32 {
	if typ == QFloat {
		return math.Float32bits(float32(v))
	}
	v = v * pow2(scale)
	if math.IsNaN(v) {
		v = 0
	}
	v = math.Trunc(v)

	lo, hi, _ := Limits(typ)
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}

	switch typ {
	case QU8:
		return uint32(uint8(v))
	case QU16:
		return uint32(uint16(v))
	case QS8:
		return uint32(uint8(int8(v)))
	case QS16:
		return uint32(uint16(int16(v)))
	}
	return 0
}

// --- Floating point ------------------------------------------------------------------

// FMA is a×c + b with a single rounding, derived in 200-bit arbitrary precision and then
// rounded once to double — rather than by calling math.FMA, which is what the interpreter
// does. If the two agree, the interpreter really is fusing.
func FMA(a, c, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsNaN(c) ||
		math.IsInf(a, 0) || math.IsInf(b, 0) || math.IsInf(c, 0) {
		// The special cases are not what this derivation is for; the interpreter's own
		// path handles them and the vector suite tests them separately.
		return math.NaN()
	}
	ba := new(big.Float).SetPrec(200).SetFloat64(a)
	bc := new(big.Float).SetPrec(200).SetFloat64(c)
	bb := new(big.Float).SetPrec(200).SetFloat64(b)
	p := new(big.Float).SetPrec(200).Mul(ba, bc) // exact: 53+53 bits fits in 200
	p.Add(p, bb)
	v, _ := p.Float64() // the one rounding
	return v
}

// RoundToSingle rounds a double to single precision. The single-precision instructions
// compute in double and then round, so a test of them is a test of this.
func RoundToSingle(v float64) float64 { return float64(float32(v)) }
