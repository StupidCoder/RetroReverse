package audio

import "testing"

// TestDecodeZeroPredictor checks the residual path in isolation: with an
// all-zero predictor book the prediction term vanishes, so every output equals
// its sign-extended nibble left-shifted by the frame scale. This pins nibble
// unpacking, sign extension, the scale shift, and the Q11 round-trip without
// needing a real sample. The full IIR path is verified bit-exactly against the
// game's stored ALADPCMLoop states by cmd/audiodump.
func TestDecodeZeroPredictor(t *testing.T) {
	book := &ADPCMBook{Order: 1, NPredictors: 1, Predictors: make([]int16, 8)}
	// header 0x20: scale shift 2, predictor 0. Then 8 data bytes → 16 nibbles.
	frame := []byte{0x20, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	pcm := DecodeADPCM(frame, book)
	if len(pcm) != 16 {
		t.Fatalf("got %d samples, want 16", len(pcm))
	}
	nibbles := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF}
	for i, n := range nibbles {
		if n >= 8 {
			n -= 16 // sign-extend the 4-bit residual
		}
		want := int16(n << 2) // scale shift 2
		if pcm[i] != want {
			t.Errorf("sample %d = %d, want %d", i, pcm[i], want)
		}
	}
}

// TestDecodeStableFeedback checks the order-2 predictor feeds history back
// without runaway: a gentle predictor and zero residuals must decay, never clip.
func TestDecodeStableFeedback(t *testing.T) {
	// Order 2, one predictor. Row 0 (oldest sample) and row 1 (newest) are small
	// so the filter is well inside the unit circle.
	pred := make([]int16, 16)
	for i := range pred[8:] {
		pred[8+i] = 512 // row 1 ≈ 0.25 in Q11 on the newest sample
	}
	book := &ADPCMBook{Order: 2, NPredictors: 1, Predictors: pred}
	// A first frame that seeds some energy, then a silent frame.
	seed := []byte{0x40, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77}
	quiet := []byte{0x00, 0, 0, 0, 0, 0, 0, 0, 0}
	pcm := DecodeADPCM(append(append([]byte{}, seed...), quiet...), book)
	if len(pcm) != 32 {
		t.Fatalf("got %d samples, want 32", len(pcm))
	}
	for i, s := range pcm {
		if s == 32767 || s == -32768 {
			t.Fatalf("sample %d clipped (%d): predictor diverged", i, s)
		}
	}
}
