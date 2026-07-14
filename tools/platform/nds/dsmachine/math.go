package dsmachine

import "math"

// The ARM9's two hardware maths units: a divider and a square-root unit, both
// memory-mapped, both operating on 64-bit values.
//
// These are not a nicety. The ARM946E-S has no divide instruction at all, so
// *every* division a DS game performs goes through this register block — and
// NitroSDK's fixed-point library routes its matrix inversions, its perspective
// divides and its vector normalisation through it. A machine that leaves the
// divider unmodelled does not fail loudly: it returns whatever the register file
// last held, and the game renders a world built on zeroes.

const (
	regDIVCNT      = 0x04000280
	regDIV_NUMER   = 0x04000290 // 64-bit
	regDIV_DENOM   = 0x04000298 // 64-bit
	regDIV_RESULT  = 0x040002A0 // 64-bit
	regDIVREM      = 0x040002A8 // 64-bit
	regSQRTCNT     = 0x040002B0
	regSQRT_RESULT = 0x040002B4 // 32-bit
	regSQRT_PARAM  = 0x040002B8 // 64-bit
)

// divider is the ARM9's division unit. Its mode selects the operand widths:
//
//	0 — 32-bit numerator / 32-bit denominator
//	1 — 64-bit numerator / 32-bit denominator
//	2 — 64-bit numerator / 64-bit denominator
//
// All of them are *signed*, and the result and remainder are 64-bit.
type divider struct {
	cnt    uint32
	numer  uint64
	denom  uint64
	result uint64
	rem    uint64
}

// run recomputes the division. The hardware does this asynchronously and sets a
// busy bit in DIVCNT while it works; we compute it immediately and leave the busy
// bit clear, which is what any polling loop is waiting to see.
func (d *divider) run() {
	mode := d.cnt & 3
	// A zero denominator sets the divide-by-zero flag (DIVCNT bit 14) and leaves
	// defined-but-strange results, which real code does hit — a normalisation of a
	// zero-length vector, say — and then discards. Producing them faithfully keeps
	// such a path on the same branch it takes on hardware.
	d.cnt &^= 1 << 14

	var num, den int64
	switch mode {
	case 0:
		num, den = int64(int32(uint32(d.numer))), int64(int32(uint32(d.denom)))
	case 1:
		num, den = int64(d.numer), int64(int32(uint32(d.denom)))
	default:
		num, den = int64(d.numer), int64(d.denom)
	}
	if den == 0 {
		d.cnt |= 1 << 14
		d.rem = uint64(num)
		// The hardware's result for a zero denominator is ±1 extended to 64 bits
		// (the sign follows the numerator), except in the 32-bit mode where the
		// upper half is the inverted sign.
		if num < 0 {
			d.result = 1
		} else {
			d.result = ^uint64(0)
		}
		if mode == 0 {
			if num < 0 {
				d.result = 0x00000000_00000001
			} else {
				d.result = 0xFFFFFFFF_FFFFFFFF
			}
		}
		return
	}
	// The one case two's-complement division cannot represent.
	if num == math.MinInt64 && den == -1 {
		d.result, d.rem = uint64(num), 0
		return
	}
	d.result = uint64(num / den)
	d.rem = uint64(num % den)
}

// sqrter is the ARM9's square-root unit: an unsigned integer square root of a
// 32-bit (mode 0) or 64-bit (mode 1) parameter, giving a 32-bit result.
type sqrter struct {
	cnt    uint32
	param  uint64
	result uint32
}

func (s *sqrter) run() {
	v := s.param
	if s.cnt&1 == 0 {
		v &= 0xFFFFFFFF
	}
	// Integer square root by bit-by-bit restoration. Done in integers rather than
	// via math.Sqrt on purpose: float64 has 53 bits of mantissa and the parameter
	// has 64, so the float route is off by one for large inputs — and "off by one"
	// in a vector normalisation is a shape that slowly deforms.
	var res, bit uint64 = 0, 1 << 62
	for bit > v {
		bit >>= 2
	}
	for bit != 0 {
		if v >= res+bit {
			v -= res + bit
			res = res>>1 + bit
		} else {
			res >>= 1
		}
		bit >>= 2
	}
	s.result = uint32(res)
}
