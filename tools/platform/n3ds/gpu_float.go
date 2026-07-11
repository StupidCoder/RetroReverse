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
