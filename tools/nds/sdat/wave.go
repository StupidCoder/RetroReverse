package sdat

import "fmt"

// Wave is one decoded SWAV sample: mono PCM at its own rate, with an optional
// loop (sample indices).
type Wave struct {
	Rate      int
	Loop      bool
	LoopStart int
	Samples   []float64 // -1..1
}

// ParseSWAR decodes every wave in a SWAR archive.
func ParseSWAR(data []byte) ([]Wave, error) {
	if len(data) < 0x3C || string(data[:4]) != "SWAR" {
		return nil, fmt.Errorf("sdat: not a SWAR file")
	}
	n := int(le.Uint32(data[0x38:]))
	var out []Wave
	for i := 0; i < n; i++ {
		off := int(le.Uint32(data[0x3C+i*4:]))
		w, err := parseSWAV(data[off:])
		if err != nil {
			return nil, fmt.Errorf("sdat: wave %d: %w", i, err)
		}
		out = append(out, w)
	}
	return out, nil
}

// parseSWAV decodes one wave: {u8 type, u8 loop, u16 rate, u16 time,
// u16 loopStart, u32 loopLen} + sample data. loopStart/loopLen are in 32-bit
// words of encoded data (verified: (loopStart+loopLen)*4 == data size).
func parseSWAV(b []byte) (Wave, error) {
	if len(b) < 12 {
		return Wave{}, fmt.Errorf("short SWAV")
	}
	typ, loop := b[0], b[1] != 0
	rate := int(le.Uint16(b[2:]))
	loopStart := int(le.Uint16(b[6:]))
	loopLen := int(le.Uint32(b[8:]))
	size := (loopStart + loopLen) * 4
	if 12+size > len(b) {
		size = len(b) - 12
	}
	data := b[12 : 12+size]
	w := Wave{Rate: rate, Loop: loop}
	switch typ {
	case 0: // PCM8
		w.Samples = make([]float64, len(data))
		for i, v := range data {
			w.Samples[i] = float64(int8(v)) / 128
		}
		w.LoopStart = loopStart * 4
	case 1: // PCM16
		w.Samples = make([]float64, len(data)/2)
		for i := range w.Samples {
			w.Samples[i] = float64(int16(le.Uint16(data[i*2:]))) / 32768
		}
		w.LoopStart = loopStart * 2
	case 2: // IMA-ADPCM: 4-byte header {u16 initial sample, u16 initial index}
		if len(data) < 4 {
			return w, fmt.Errorf("short ADPCM")
		}
		pred := int(int16(le.Uint16(data)))
		idx := int(le.Uint16(data[2:])) & 0x7F
		w.Samples = decodeIMA(data[4:], pred, idx)
		// loopStart counts encoded words incl. the 4-byte header; 2 samples/byte
		w.LoopStart = (loopStart - 1) * 8
		if w.LoopStart < 0 {
			w.LoopStart = 0
		}
	default:
		return w, fmt.Errorf("unknown SWAV type %d", typ)
	}
	if w.LoopStart >= len(w.Samples) {
		w.Loop = false
	}
	return w, nil
}

var imaIndexDelta = [8]int{-1, -1, -1, -1, 2, 4, 6, 8}

var imaStep = [89]int{
	7, 8, 9, 10, 11, 12, 13, 14, 16, 17, 19, 21, 23, 25, 28, 31, 34, 37, 41, 45,
	50, 55, 60, 66, 73, 80, 88, 97, 107, 118, 130, 143, 157, 173, 190, 209, 230,
	253, 279, 307, 337, 371, 408, 449, 494, 544, 598, 658, 724, 796, 876, 963,
	1060, 1166, 1282, 1411, 1552, 1707, 1878, 2066, 2272, 2499, 2749, 3024, 3327,
	3660, 4026, 4428, 4871, 5358, 5894, 6484, 7132, 7845, 8630, 9493, 10442,
	11487, 12635, 13899, 15289, 16818, 18500, 20350, 22385, 24623, 27086, 29794,
	32767,
}

// decodeIMA expands IMA-ADPCM nibbles (low nibble first).
func decodeIMA(data []byte, pred, idx int) []float64 {
	out := make([]float64, 0, len(data)*2)
	if idx > 88 {
		idx = 88
	}
	step := func(nib int) {
		st := imaStep[idx]
		diff := st >> 3
		if nib&1 != 0 {
			diff += st >> 2
		}
		if nib&2 != 0 {
			diff += st >> 1
		}
		if nib&4 != 0 {
			diff += st
		}
		if nib&8 != 0 {
			pred -= diff
		} else {
			pred += diff
		}
		if pred > 32767 {
			pred = 32767
		}
		if pred < -32768 {
			pred = -32768
		}
		idx += imaIndexDelta[nib&7]
		if idx < 0 {
			idx = 0
		}
		if idx > 88 {
			idx = 88
		}
		out = append(out, float64(pred)/32768)
	}
	for _, b := range data {
		step(int(b & 0xF))
		step(int(b >> 4))
	}
	return out
}
