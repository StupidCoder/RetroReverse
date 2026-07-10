package audio

// vadpcm.go decodes the Nintendo 64's VADPCM: the codec libultra stores samples
// in and the RSP mixes from. A sample is a run of 9-byte frames; each frame
// expands to 16 s16 PCM samples. The frame's first byte carries a 4-bit scale
// (a left shift) and a 4-bit predictor index into the ALADPCMBook; the other 8
// bytes hold 16 signed 4-bit residuals. Reconstruction is a short IIR filter:
// each output is the book's predictor applied to the carried history plus the
// scaled residual, all in Q11 fixed point.
//
// The decoder is verified without any external oracle: a looping sample's
// ALADPCMLoop stores the exact 16-sample decoder state at the loop point, so
// decoding up to that point and matching those 16 values proves the codec bit
// for bit (see vadpcm_test.go).

// DecodeADPCM expands one VADPCM sample (raw ".TBL" bytes) to mono s16 PCM using
// its predictor book. len(data) need not be a whole number of frames; a trailing
// partial frame is ignored, as the hardware does.
func DecodeADPCM(data []byte, book *ADPCMBook) []int16 {
	order := int(book.Order)
	if order < 1 {
		order = 1
	}
	state := make([]int32, order) // state[0] = most recent output sample
	out := make([]int16, 0, len(data)/9*16)
	var frameOut [16]int32
	for f := 0; f+9 <= len(data); f += 9 {
		decodeFrame(data[f:f+9], book, order, state, &frameOut)
		for _, s := range frameOut {
			out = append(out, clamp16(s))
		}
	}
	return out
}

// decodeFrame decodes one 9-byte frame into frameOut and advances state to hold
// the last `order` samples (state[0] newest).
func decodeFrame(frame []byte, book *ADPCMBook, order int, state []int32, frameOut *[16]int32) {
	header := frame[0]
	shift := uint(header >> 4)
	p := int(header & 0x0f)
	if p >= int(book.NPredictors) {
		p = 0
	}
	// 16 signed 4-bit residuals, left-shifted by the frame scale.
	var res [16]int32
	for i := 0; i < 16; i++ {
		b := frame[1+i/2]
		var nib int32
		if i&1 == 0 {
			nib = int32(int8(b) >> 4) // high nibble, arithmetic shift sign-extends
		} else {
			nib = int32(int8(b<<4) >> 4) // low nibble
		}
		res[i] = nib << shift
	}
	// coef[k][i] = Predictors[(p*order + k)*8 + i]. Row 0 propagates the oldest
	// carried sample, row order-1 the newest; that same last row is the impulse
	// response convolved with the residuals within a sub-frame.
	coef := func(k, i int) int64 { return int64(book.Predictors[(p*order+k)*8+i]) }
	// state[k] pairs with coef row k: state[0] is the oldest of the carried
	// samples, state[order-1] the newest.
	for sub := 0; sub < 2; sub++ {
		in := res[sub*8 : sub*8+8]
		for i := 0; i < 8; i++ {
			acc := int64(in[i]) << 11
			for k := 0; k < order; k++ {
				acc += coef(k, i) * int64(state[k])
			}
			for k := 0; k < i; k++ {
				acc += coef(order-1, i-1-k) * int64(in[k])
			}
			frameOut[sub*8+i] = int32(clamp16(int32(acc >> 11)))
		}
		// Carry the sub-frame's last `order` samples, oldest first.
		for k := 0; k < order; k++ {
			state[k] = frameOut[sub*8+(8-order)+k]
		}
	}
}

func clamp16(v int32) int16 {
	if v > 32767 {
		return 32767
	}
	if v < -32768 {
		return -32768
	}
	return int16(v)
}
