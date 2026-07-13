package n3ds

// dsp_voice.go is the audio half of the DSP model: the 24 voices (sample
// decode, resampling, filters), the three intermediate mixers and the final
// mix. dsp.go is the other half — the dsp::DSP service, the pipes, and the
// frame clock that calls in here once per audio frame.
//
// The firmware's frame is a fixed pipeline, and this file follows it step for
// step:
//
//	for each of the 24 sources:
//	    consume its 192-byte configuration from the region the app just wrote
//	    decode / resample / filter 160 output samples from its buffer queue
//	    publish its 12-byte status back to the other region
//	    add its frame into each of the 3 intermediate mixes, at that mix's gains
//	the mixers: aux-return, aux-send, downmix the 3 quad mixes to stereo
//	write the 160-sample stereo result into the write region
//
// Everything here is 16-bit PCM in, 16-bit PCM out; the intermediate mixes are
// quadraphonic s32 (they are exposed to the app, which may edit them — that is
// what the aux busses are).
//
// Clean-room note: as in dsp.go, the DSP is platform hardware, and the frame
// pipeline, structure layouts and codec follow the platform-level references
// (3dbrew's DSP memory-region documentation and the Citra/Azahar HLE) under the
// user-approved platform exception, reimplemented in Go.

import (
	"fmt"
	"math"
)

const dspSamplesPerFrame = 160

// dspSample is one stereo PCM16 sample. Mono sources decode to both channels.
type dspSample [2]int16

// dspFrame is one audio frame of stereo output — what a source generates and
// the mixer consumes.
type dspFrame [dspSamplesPerFrame]dspSample

// dspQuadFrame is one audio frame on an intermediate mixer: four channels
// (front-left, front-right, rear-left, rear-right) at 32-bit headroom, so
// summing 24 voices cannot clip before the final downmix.
type dspQuadFrame [dspSamplesPerFrame][4]int32

// Sample formats (SourceConfiguration flags1, bits 2-3) and channel count
// (bits 0-1).
const (
	dspFmtPCM8  = 0
	dspFmtPCM16 = 1
	dspFmtADPCM = 2
)

// Interpolation modes (SourceConfiguration::interpolation_mode).
const (
	dspInterpPolyphase = 0
	dspInterpLinear    = 1
	dspInterpNone      = 2
)

// dspSource is one of the 24 voices.
type dspSource struct {
	Enabled   bool
	SyncCount uint16
	Rate      float32 // input samples consumed per output sample
	Interp    uint8
	Format    uint8
	Stereo    bool
	Gain      [3][4]float32 // per intermediate mixer, per channel
	Filters   dspFilters

	AdpcmCoeffs [16]int16 // s5.11, from the ADPCM-coefficient structure
	AdpcmYn1    int16
	AdpcmYn2    int16

	Queue  []dspBuffer // pending buffers, played lowest buffer id first
	CurBuf []dspSample // the decoded current buffer, samples not yet consumed

	CurSample    uint32 // play position within the current buffer, in samples
	CurPhysAddr  uint32
	CurBufferID  uint16
	LastBufferID uint16
	BufferUpdate bool // current_buffer_id changed; reported once, then cleared

	// Resampler carry: the two input samples straddling the frame boundary and
	// the fractional read position (24 fractional bits).
	Xn1  dspSample
	Xn2  dspSample
	FPos uint64

	Frame dspFrame // this frame's output, held for the mixer
}

// dspBuffer is one queued buffer's metadata. The PCM itself stays in the
// application's memory and is decoded on dequeue.
type dspBuffer struct {
	PhysAddr     uint32
	Length       uint32 // IN SAMPLES — a sample is 1, 2 or 4/7 bytes by format
	AdpcmPS      uint8  // predictor (high nibble) and scale (low nibble)
	AdpcmYn      [2]int16
	AdpcmDirty   bool
	IsLooping    bool
	BufferID     uint16
	Stereo       bool
	Format       uint8
	FromQueue    bool // from the buffer queue, not the embedded buffer
	PlayPosition uint32
	HasPlayed    bool
}

// dspFilters is a voice's two optional filters, both normalised and with the
// feedback coefficients pre-negated by the application.
type dspFilters struct {
	SimpleEnabled bool
	BiquadEnabled bool

	// H(z) = b0 / (1 - a1 z^-1), coefficients s1.15 (held wider than the s16 they
	// arrive as: passthrough is b0 = 1.0 = 1<<15, which an s16 cannot hold).
	SB0 int32
	SA1 int32
	SY1 dspSample

	// H(z) = (b0 + b1 z^-1 + b2 z^-2) / (1 - a1 z^-1 - a2 z^-2), s2.14.
	BA1 int16
	BA2 int16
	BB0 int16
	BB1 int16
	BB2 int16
	BX1 dspSample
	BX2 dspSample
	BY1 dspSample
	BY2 dspSample
}

func (f *dspFilters) reset() {
	*f = dspFilters{SB0: 1 << 15, BB0: 1 << 14} // both passthrough
}

// enable resets whichever filter is being turned off, so a re-enable starts
// from a clean history rather than resuming a stale one.
func (f *dspFilters) enable(simple, biquad bool) {
	f.SimpleEnabled, f.BiquadEnabled = simple, biquad
	if !simple {
		f.SB0, f.SA1, f.SY1 = 1<<15, 0, dspSample{}
	}
	if !biquad {
		f.BB0, f.BB1, f.BB2, f.BA1, f.BA2 = 1<<14, 0, 0, 0, 0
		f.BX1, f.BX2, f.BY1, f.BY2 = dspSample{}, dspSample{}, dspSample{}, dspSample{}
	}
}

func (f *dspFilters) processFrame(frame *dspFrame) {
	if !f.SimpleEnabled && !f.BiquadEnabled {
		return
	}
	for i := range frame {
		x := frame[i]
		if f.SimpleEnabled {
			var y dspSample
			for c := 0; c < 2; c++ {
				y[c] = clampS16((f.SB0*int32(x[c]) + f.SA1*int32(f.SY1[c])) >> 15)
			}
			f.SY1, x = y, y
		}
		if f.BiquadEnabled {
			var y dspSample
			for c := 0; c < 2; c++ {
				acc := int32(f.BB0)*int32(x[c]) + int32(f.BB1)*int32(f.BX1[c]) + int32(f.BB2)*int32(f.BX2[c]) +
					int32(f.BA1)*int32(f.BY1[c]) + int32(f.BA2)*int32(f.BY2[c])
				y[c] = clampS16(acc >> 14)
			}
			f.BX2, f.BX1 = f.BX1, x
			f.BY2, f.BY1 = f.BY1, y
			x = y
		}
		frame[i] = x
	}
}

func clampS16(v int32) int16 {
	switch {
	case v > 32767:
		return 32767
	case v < -32768:
		return -32768
	}
	return int16(v)
}

func addClampS16(a, b int16) int16 { return clampS16(int32(a) + int32(b)) }

// --- the per-source frame -----------------------------------------------------

// dspSourceFrame generates source i's 160 output samples: dequeue buffers as
// they run out, resample at the source's rate, then run the filters. A source
// whose queue runs dry reports "finished" — is_enabled goes 0 and the id of the
// buffer that just ended appears in last_buffer_id. That is the DSP's call, not
// the application's: the app's streaming code watches for exactly this to know
// a voice may be recycled.
func (m *Machine) dspSourceFrame(i int) {
	s := &m.dsp.Sources[i]
	s.Frame = dspFrame{}

	if len(s.CurBuf) == 0 {
		if m.dspDequeue(s) {
			return // one frame of silence while the new buffer starts
		}
		s.Enabled = false
		s.BufferUpdate = true
		s.LastBufferID = s.CurBufferID
		s.CurBufferID = 0
		return
	}

	pos := 0
	// A looping buffer whose address never decodes would re-queue forever; the
	// bound turns that into a silent frame instead of a hang.
	for n := 0; pos < dspSamplesPerFrame && n < 64; n++ {
		if len(s.CurBuf) == 0 && !m.dspDequeue(s) {
			break
		}
		s.resample(&pos)
	}
	s.CurSample += uint32(float64(pos) * float64(s.Rate))
	s.Filters.processFrame(&s.Frame)
}

// resample steps over the current buffer at the source's rate, writing output
// samples until the frame is full or the buffer is consumed. Position is fixed
// point with 24 fractional bits, carried across frames (FPos) along with the
// two input samples the next interpolation needs (Xn1/Xn2).
func (s *dspSource) resample(outi *int) {
	const scale = 1 << 24
	if len(s.CurBuf) == 0 {
		return
	}
	rate := float64(s.Rate)
	if !(rate > 0) {
		rate = 1
	}
	// The two carried samples sit in front of the buffer, so an interpolation
	// straddling the frame boundary sees its left neighbours.
	input := make([]dspSample, 0, len(s.CurBuf)+2)
	input = append(input, s.Xn2, s.Xn1)
	input = append(input, s.CurBuf...)

	step := uint64(rate * scale)
	fpos := s.FPos
	idx := 0
	for *outi < dspSamplesPerFrame {
		idx = int(fpos / scale)
		if idx+2 >= len(input) {
			idx = len(input) - 2
			break
		}
		frac := fpos & (scale - 1)
		s.Frame[*outi] = interpolate(s.Interp, frac, input[idx], input[idx+1])
		*outi++
		fpos += step
	}
	s.Xn2, s.Xn1 = input[idx], input[idx+1]
	s.FPos = fpos - uint64(idx)*scale
	s.CurBuf = s.CurBuf[idx:] // the samples stepped over are consumed
}

// interpolate produces one output sample from the two input samples around the
// fractional position. Polyphase (the firmware's default, a windowed-sinc bank)
// is approximated by linear — the same simplification the reference HLE makes;
// it is a fidelity gap, not a protocol one.
func interpolate(mode uint8, frac uint64, x0, x1 dspSample) dspSample {
	if mode == dspInterpNone {
		return x0
	}
	var out dspSample
	for c := 0; c < 2; c++ {
		// Saturated difference, as the firmware does it.
		delta := clampS16(int32(x1[c]) - int32(x0[c]))
		out[c] = int16(int64(x0[c]) + int64(frac)*int64(delta)/(1<<24))
	}
	return out
}

// dspDequeue pops the lowest-id buffer into the current-buffer slot and decodes
// it. Reports whether a buffer was taken.
func (m *Machine) dspDequeue(s *dspSource) bool {
	if len(s.Queue) == 0 {
		return false
	}
	best := 0
	for i := range s.Queue {
		if s.Queue[i].BufferID < s.Queue[best].BufferID {
			best = i
		}
	}
	buf := s.Queue[best]
	s.Queue = append(s.Queue[:best], s.Queue[best+1:]...)

	if m.DSPTrace {
		fmt.Printf("    dsp DEQUEUE src? id=%d addr=%08X len=%d loop=%v fromQueue=%v played=%v (queue left %d, frame %d)\n",
			buf.BufferID, buf.PhysAddr, buf.Length, buf.IsLooping, buf.FromQueue, buf.HasPlayed, len(s.Queue), m.dsp.Ticks)
	}
	if buf.AdpcmDirty {
		s.AdpcmYn1, s.AdpcmYn2 = buf.AdpcmYn[0], buf.AdpcmYn[1]
	}

	// The buffer address is physical, and the DSP's DMA drops its low two bits.
	s.CurBuf = m.dspDecodeBuffer(s, buf)

	// The first playthrough starts at the configured play position; a loop
	// restarts at the beginning.
	s.CurSample = 0
	if !buf.HasPlayed {
		s.CurSample = buf.PlayPosition
	}
	s.CurPhysAddr = buf.PhysAddr
	s.CurBufferID = buf.BufferID
	s.LastBufferID = 0
	s.BufferUpdate = buf.FromQueue && !buf.HasPlayed
	// The resampler's history and fractional position deliberately carry ACROSS
	// buffers: consecutive buffers of a stream are one continuous signal, and
	// restarting the interpolator at each boundary would click.

	if n := int(s.CurSample); n < len(s.CurBuf) {
		s.CurBuf = s.CurBuf[n:]
	} else {
		s.CurBuf = nil
	}

	if buf.IsLooping {
		buf.HasPlayed = true
		s.Queue = append(s.Queue, buf)
	}
	return true
}

// dspDecodeBuffer reads a buffer's PCM out of application memory and returns it
// as stereo samples. Byte↔sample arithmetic is the whole point: the buffer's
// length is in SAMPLES, and a sample is one byte (PCM8), two (PCM16) or four
// sevenths of one (ADPCM: 8 bytes carry 14 samples) — times the channel count.
func (m *Machine) dspDecodeBuffer(s *dspSource, buf dspBuffer) []dspSample {
	if int32(buf.Length) <= 0 {
		return nil
	}
	addr := m.gpuAddrToVirt(buf.PhysAddr &^ 3)
	chans := uint32(1)
	if buf.Stereo {
		chans = 2
	}
	switch buf.Format {
	case dspFmtPCM8:
		data := m.ReadBytes(addr, buf.Length*chans)
		out := make([]dspSample, buf.Length)
		for i := range out {
			if buf.Stereo {
				out[i] = dspSample{pcm8(data[i*2]), pcm8(data[i*2+1])}
			} else {
				v := pcm8(data[i])
				out[i] = dspSample{v, v}
			}
		}
		return out
	case dspFmtPCM16:
		data := m.ReadBytes(addr, buf.Length*chans*2)
		out := make([]dspSample, buf.Length)
		for i := range out {
			if buf.Stereo {
				out[i] = dspSample{pcm16(data[i*4:]), pcm16(data[i*4+2:])}
			} else {
				v := pcm16(data[i*2:])
				out[i] = dspSample{v, v}
			}
		}
		return out
	case dspFmtADPCM:
		frames := (buf.Length + 13) / 14
		return m.decodeADPCM(s, buf, m.ReadBytes(addr, frames*8))
	}
	return nil
}

func pcm8(b byte) int16 { return int16(uint16(b) << 8) }

func pcm16(b []byte) int16 { return int16(uint16(b[0]) | uint16(b[1])<<8) }

// decodeADPCM decodes GC-ADPCM (mono): 8-byte blocks of one header byte and 14
// 4-bit samples. The header carries a scale (low nibble, a power of two) and an
// index (high nibble) into the source's 16 coefficients — two per index, the
// taps of a second-order predictor whose state (y[n-1], y[n-2]) carries across
// blocks and across buffers.
func (m *Machine) decodeADPCM(s *dspSource, buf dspBuffer, data []byte) []dspSample {
	if buf.AdpcmDirty {
		// The header nibbles of the buffer's own predictor/scale are only used by
		// the firmware to prime the first block; the block headers themselves
		// carry it thereafter.
		_ = buf.AdpcmPS
	}
	yn1, yn2 := int32(s.AdpcmYn1), int32(s.AdpcmYn2)
	out := make([]dspSample, 0, buf.Length)
	for blk := 0; blk*8 < len(data) && uint32(len(out)) < buf.Length; blk++ {
		hdr := data[blk*8]
		scale := int32(1) << (hdr & 0xF)
		idx := int((hdr >> 4) & 0x7)
		c1, c2 := int32(s.AdpcmCoeffs[idx*2]), int32(s.AdpcmCoeffs[idx*2+1])

		decode := func(nibble int32) int16 {
			if nibble >= 8 {
				nibble -= 16 // the nibble is a signed 4-bit value
			}
			// Coefficients are s5.11: run the predictor in 11-bit fixed point,
			// round (0x400 = 0.5), and come back to s16.
			v := ((nibble*scale)<<11 + 0x400 + c1*yn1 + c2*yn2) >> 11
			r := clampS16(v)
			yn2, yn1 = yn1, int32(r)
			return r
		}
		for i := 1; i < 8 && uint32(len(out)) < buf.Length; i++ {
			b := data[blk*8+i]
			hi := decode(int32(b >> 4))
			out = append(out, dspSample{hi, hi})
			if uint32(len(out)) >= buf.Length {
				break
			}
			lo := decode(int32(b & 0xF))
			out = append(out, dspSample{lo, lo})
		}
	}
	s.AdpcmYn1, s.AdpcmYn2 = int16(yn1), int16(yn2)
	return out
}

// mixInto adds this source's frame into one intermediate mix at that mix's four
// gains. Stereo → quadraphonic: the front pair takes the source's left/right at
// gains 0/1, the rear pair the same left/right at gains 2/3.
func (s *dspSource) mixInto(dest *dspQuadFrame, mix int) {
	if !s.Enabled {
		return
	}
	g := &s.Gain[mix]
	for i := range dest {
		l, r := float32(s.Frame[i][0]), float32(s.Frame[i][1])
		dest[i][0] += int32(g[0] * l)
		dest[i][1] += int32(g[1] * r)
		dest[i][2] += int32(g[2] * l)
		dest[i][3] += int32(g[3] * r)
	}
}

// --- the mixers ---------------------------------------------------------------

// DspConfiguration field offsets, within the structure at dspOffDSPConfig.
const (
	cfgDirty          = 0x00 // u32
	cfgMasterVolume   = 0x04 // f32 — the volume of intermediate mix 0 at the final mixer
	cfgAuxReturnVol   = 0x08 // f32[2] — of mixes 1 and 2
	cfgOutputFormat   = 0x16 // u16: 0 mono, 1 stereo, 2 surround
	cfgClippingMode   = 0x18 // u16 (the limiter — not modelled; see mixFinal)
	cfgHeadphones     = 0x1A // u16
	cfgAuxBusEnable   = 0x28 // u16[2]
	cfgDirty2         = 0xC0 // u32

	cfgDirtyAuxBus0    = 1 << 8
	cfgDirtyAuxBus1    = 1 << 9
	cfgDirtyMasterVol  = 1 << 16
	cfgDirtyAuxReturn0 = 1 << 24
	cfgDirtyAuxReturn1 = 1 << 25
	cfgDirtyOutFormat  = 1 << 26

	dspOutMono     = 0
	dspOutStereo   = 1
	dspOutSurround = 2
)

// Intermediate mix samples: two quadraphonic s32 frames, channel-major and
// plain little-endian (not the DSP's middle-endian) — the app reads and writes
// these directly.
const dspIntermediateMixSize = 4 * dspSamplesPerFrame * 4 // one mix, in bytes

// dspMixerConfig consumes the DSP configuration from the read region: the three
// intermediate-mixer volumes, the aux busses and the output format.
func (m *Machine) dspMixerConfig(region uint32) {
	cfg := region + dspOffDSPConfig
	dirty := m.ReadWord(cfg + cfgDirty)
	if dirty == 0 {
		return
	}
	d := &m.dsp
	if dirty&cfgDirtyAuxBus0 != 0 {
		d.AuxBusEnable[0] = m.dspRead16(cfg+cfgAuxBusEnable) != 0
	}
	if dirty&cfgDirtyAuxBus1 != 0 {
		d.AuxBusEnable[1] = m.dspRead16(cfg+cfgAuxBusEnable+2) != 0
	}
	if dirty&cfgDirtyMasterVol != 0 {
		d.MixVolume[0] = dspFloat(m.ReadWord(cfg + cfgMasterVolume))
	}
	if dirty&cfgDirtyAuxReturn0 != 0 {
		d.MixVolume[1] = dspFloat(m.ReadWord(cfg + cfgAuxReturnVol))
	}
	if dirty&cfgDirtyAuxReturn1 != 0 {
		d.MixVolume[2] = dspFloat(m.ReadWord(cfg + cfgAuxReturnVol + 4))
	}
	if dirty&cfgDirtyOutFormat != 0 {
		d.OutputFormat = m.dspRead16(cfg + cfgOutputFormat)
	}
	d.Headphones = m.dspRead16(cfg+cfgHeadphones) != 0
	d.ClippingMode = m.dspRead16(cfg + cfgClippingMode)
	m.WriteWord(cfg+cfgDirty, 0)
	m.WriteWord(cfg+cfgDirty2, 0)
}

// dspFloat reads a shared-memory float. These are plain little-endian f32 (the
// middle-endian rule is for the DSP's u32s).
func dspFloat(w uint32) float32 {
	f := math.Float32frombits(w)
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return 0
	}
	return f
}

// dspMix runs the three intermediate mixers and the final mix, exactly as the
// firmware orders it: an enabled aux bus is a round trip THROUGH the app — the
// DSP writes that mix's quad frame into the write region and takes back what
// the app left in the read region (its effects run on the ARM11, one frame
// later). Then the three mixes are downmixed to stereo at their volumes and
// summed.
func (m *Machine) dspMix(read, write uint32, mixes *[3]dspQuadFrame) dspFrame {
	d := &m.dsp
	var bus [3]dspQuadFrame

	// Aux return: what the app produced last frame, for each enabled bus.
	for b := 0; b < 2; b++ {
		if d.AuxBusEnable[b] {
			m.dspReadQuad(read+dspOffIntermediate+uint32(b)*dspIntermediateMixSize, &bus[b+1])
		}
	}
	// Aux send: mix 0 always goes straight to the final mixer; mixes 1 and 2 are
	// handed to the app when their bus is enabled, and pass through when not.
	bus[0] = mixes[0]
	for b := 0; b < 2; b++ {
		if d.AuxBusEnable[b] {
			m.dspWriteQuad(write+dspOffIntermediate+uint32(b)*dspIntermediateMixSize, &mixes[b+1])
		} else {
			bus[b+1] = mixes[b+1]
		}
	}

	var out dspFrame
	for mix := 0; mix < 3; mix++ {
		g := d.MixVolume[mix]
		for i := range out {
			s := &bus[mix][i]
			var l, r int16
			if d.OutputFormat == dspOutMono {
				mono := clampS16(int32((g*float32(s[0]) + g*float32(s[1]) + g*float32(s[2]) + g*float32(s[3])) / 2))
				l, r = mono, mono
			} else {
				// Stereo, and surround downmixed to it (the surround coefficients
				// depend on console state we do not model).
				l = clampS16(int32(g*float32(s[0]) + g*float32(s[2])))
				r = clampS16(int32(g*float32(s[1]) + g*float32(s[3])))
			}
			out[i][0] = addClampS16(out[i][0], l)
			out[i][1] = addClampS16(out[i][1], r)
		}
	}
	// The firmware's limiter/compressor (clipping_mode, and the compressor
	// response table in shared memory) is NOT modelled: the mix above is what a
	// disabled limiter produces. Recorded, not faked.
	return out
}

func (m *Machine) dspReadQuad(base uint32, q *dspQuadFrame) {
	for c := 0; c < 4; c++ {
		for i := 0; i < dspSamplesPerFrame; i++ {
			q[i][c] = int32(m.ReadWord(base + uint32(c*dspSamplesPerFrame+i)*4))
		}
	}
}

func (m *Machine) dspWriteQuad(base uint32, q *dspQuadFrame) {
	for c := 0; c < 4; c++ {
		for i := 0; i < dspSamplesPerFrame; i++ {
			m.WriteWord(base+uint32(c*dspSamplesPerFrame+i)*4, uint32(q[i][c]))
		}
	}
}

// dspWriteFinal publishes the final mix into the write region (interleaved
// stereo PCM16) and, when the oracle is capturing, appends it to the WAV.
func (m *Machine) dspWriteFinal(write uint32, f *dspFrame) {
	for i := range f {
		m.dspWrite16(write+dspOffFinalSamples+uint32(i)*4, uint16(f[i][0]))
		m.dspWrite16(write+dspOffFinalSamples+uint32(i)*4+2, uint16(f[i][1]))
	}
	if m.AudioCapture {
		for i := range f {
			m.AudioPCM = append(m.AudioPCM, f[i][0], f[i][1])
		}
	}
}
