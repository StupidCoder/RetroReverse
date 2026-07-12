package n3ds

import (
	"math"
	"testing"
)

// A PICA200 shader constant is a float24, so a uniform uploaded in the f32 mode
// of register 0x2C0 is quantised on the way into the register file. What the
// format cannot represent — NaN, infinity, anything under its exponent range —
// does not survive the trip. Captain Toad's whole render depends on this: it
// disables stereoscopic 3D by poisoning c4 with the quiet-NaN word 0x7FC00000,
// and its vertex shader skips the stereo block only if that parameter compares
// equal to zero. Keep the NaN and the comparison goes the other way: the stereo
// maths runs and NaN floods every clip position.
func TestUniformFloat24Quantisation(t *testing.T) {
	nan := math.Float32frombits(0x7FC00000)
	if got := toF24(nan); got != 0 {
		t.Errorf("float24 has no NaN encoding: toF24(NaN) = %v, want 0", got)
	}
	if got := toF24(float32(math.Inf(1))); got != 0 {
		t.Errorf("float24 has no infinity: toF24(+Inf) = %v, want 0", got)
	}
	if got := toF24(1e-30); got != 0 {
		t.Errorf("below float24's exponent range: toF24(1e-30) = %v, want 0", got)
	}

	// Ordinary values survive unchanged when they are exactly representable.
	for _, v := range []float32{0, 1, -1, 0.5, 120, -911.5} {
		if got := toF24(v); got != v {
			t.Errorf("toF24(%v) = %v, want it unchanged (exactly representable)", v, got)
		}
	}
	// float24 keeps 16 mantissa bits: the low 7 bits of an f32 mantissa go.
	if got := toF24(math.Float32frombits(0x3F80007F)); got != 1 {
		t.Errorf("toF24 must drop the low 7 mantissa bits: got %v, want 1", got)
	}
	if got := toF24(0.008333334); math.Abs(float64(got)-0.008333334) > 1e-6 {
		t.Errorf("toF24(1/120) = %v, want ~0.0083333 (within float24 precision)", got)
	}
}
