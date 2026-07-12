package n3ds

// gpu_float.go: the PICA200's number formats. GPU registers hold floats in a
// 24-bit format (1 sign, 7 exponent bits biased 63, 16 mantissa bits); the
// shader engine and uniforms use ordinary IEEE 754 single precision (the
// game's uniform uploads set the f32 flag). Fixed/default attributes pack four
// float24 values into three words, highest component first.

import "math"

// f24bits converts a 24-bit PICA float (in the low 24 bits) to float32.
func f24bits(v uint32) float32 {
	sign := v >> 23 & 1
	exp := v >> 16 & 0x7F
	man := v & 0xFFFF
	if exp == 0 && man == 0 {
		if sign != 0 {
			return float32(math.Copysign(0, -1))
		}
		return 0
	}
	// Rebias 63 → 127 and widen the mantissa 16 → 23 bits.
	f := sign<<31 | (exp+64)<<23 | man<<7
	return math.Float32frombits(f)
}

// f32bits reinterprets a register word as an IEEE 754 float32.
func f32bits(v uint32) float32 { return math.Float32frombits(v) }

// unpackF24x4 unpacks three words into a float24 vec4. The packing puts w in
// the top 24 bits of the first word and x in the low 24 bits of the last, each
// component's bits flowing into the next word: [w0: w<<8 | z>>16] [w1: z<<16 |
// y>>8] [w2: y<<24 | x].
func unpackF24x4(w0, w1, w2 uint32) [4]float32 {
	return [4]float32{
		f24bits(w2 & 0xFFFFFF),
		f24bits((w1&0xFFFF)<<8 | w2>>24),
		f24bits((w0&0xFF)<<16 | w1>>16),
		f24bits(w0 >> 8),
	}
}

// toF24 quantises an IEEE-754 float32 to what a PICA200 shader constant can
// actually hold. The shader register file is float24 (1 sign, 7 exponent, 16
// mantissa) — a uniform uploaded in "f32 mode" is converted on the way in, and
// float24 has no NaN: the exponent field is too narrow to hold the f32
// all-ones exponent, so a NaN/Inf pattern does not survive the trip.
//
// This is load-bearing, not pedantry. Captain Toad disables stereoscopic 3D by
// poisoning the stereo uniform (c4) with the quiet-NaN word 0x7FC00000 every
// frame, and its vertex shader then guards the stereo block with "skip unless
// this parameter differs from zero". Kept as an f32 NaN the guard's !=
// comparison is true (IEEE says NaN differs from everything), the stereo maths
// runs, and NaN floods the clip position — every triangle in the game is
// rejected and both screens stay black. Quantised, the parameter is an ordinary
// finite number and the shader behaves.
func toF24(v float32) float32 {
	b := math.Float32bits(v)
	sign, exp, man := b>>31, int(b>>23&0xFF), b&0x7FFFFF
	switch {
	case exp == 0xFF: // NaN or Inf: no float24 encoding — the pattern is lost
		exp, man = 0, 0
	case exp == 0: // (sub)normal below float24's range: flushes to zero
		man = 0
	default:
		e := exp - 127 + 63 // rebias 127 → 63
		if e <= 0 {         // underflows float24's exponent → zero
			exp, man = 0, 0
		} else if e >= 0x7F { // overflows → the largest finite float24
			exp, man = 127+63-1, 0x7F0000
		} else {
			exp = e + 64      // back to an f32 exponent
			man &^= 0x7F      // float24 keeps 16 mantissa bits; drop the low 7
		}
	}
	if exp == 0 && man == 0 {
		if sign != 0 {
			return float32(math.Copysign(0, -1))
		}
		return 0
	}
	return math.Float32frombits(sign<<31 | uint32(exp)<<23 | man)
}
