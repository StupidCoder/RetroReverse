// Package sid is a reusable MOS 6581 (SID) emulator aimed at rendering C64 music to
// audio. It models the three voices (phase-accumulator oscillators with the
// saw/triangle/pulse/noise waveforms, plus ring-mod and hard-sync), the per-voice ADSR
// envelope generator with the chip's exponential decay/release curve (the reSID rate
// model), and the multimode (low/band/high-pass) filter. It is register-accurate: a
// driver writes the 29 SID registers ($D400-$D418 as offsets 0..24) exactly as the 6502
// would, then clocks the chip and reads samples.
//
// It is not cycle-perfect like reSID — combined waveforms are approximated by ANDing the
// selected waveforms, and the filter is a discrete state-variable model — but it is
// faithful enough to reproduce a tune's pitch, rhythm, envelopes and timbre.
package sid

import (
	"math"
	"os"
	"strconv"
)

// PAL is the PAL C64 system clock in Hz (the SID is clocked from it).
const PAL = 985248.0
const NTSC = 1022727.0

// envelope rate-counter periods (clock cycles per step) for the 16 attack/decay/release
// settings — the canonical 6581 values.
var ratePeriod = [16]uint32{
	9, 32, 63, 95, 149, 220, 267, 313,
	392, 977, 1954, 3126, 3907, 11720, 19532, 31251,
}

type voice struct {
	freq    uint16 // 16-bit frequency
	pw      uint16 // 12-bit pulse width
	ctrl    uint8  // control: gate/sync/ring/test/tri/saw/pulse/noise
	ad, sr  uint8  // attack-decay, sustain-release
	acc     uint32 // 24-bit phase accumulator
	accPrev uint32 // previous accumulator (for sync/MSB-edge detection)
	noise   uint32 // 23-bit noise LFSR
	// envelope generator
	env      uint8 // 0..255
	state    int   // 0 attack, 1 decay/sustain, 2 release
	rateCnt  uint32
	expCnt   uint32
	expPer   uint32
	holdZero bool
}

const (
	stAttack = iota
	stDecay
	stRelease
)

// SID is one 6581 chip.
type SID struct {
	v       [3]voice
	fcLo    uint8
	fcHi    uint8
	res     uint8 // resonance + filter-routing
	mode    uint8 // volume + filter mode bits
	clock   float64
	srate   float64
	cycPerS float64
	sAcc    float64
	// filter state (state-variable)
	lp, bp float64
	// anti-alias decimation: a 4-pole low-pass run at the clock rate before sampling, so
	// the waveforms' high harmonics don't alias down (which would inflate the bright voices).
	aaA                float64
	aa1, aa2, aa3, aa4 float64
	resScale           float64    // resonance->damping scale (tunable)
	wlpA               float64    // per-voice waveform-rounding low-pass coefficient (the 6581 DAC)
	vlp                [3]float64 // per-voice waveform low-pass state
	dcA                float64    // per-voice DC-blocker coefficient (the output AC-coupling)
	vdc                [3]float64 // per-voice DC estimate
	// reSID-style DAC non-linearity tables (the 6581's R-2R ladder is not perfectly
	// linear; high codes compress, which darkens the bright voices). dacWave maps the
	// 12-bit waveform code, dacEnv the 8-bit envelope code, both to a 0..1 analog level.
	dacWave  [4096]float64
	dacEnv   [256]float64
	dacWZero float64 // dacWave at mid-code 2048, the AC centre
	envDAC   bool    // apply DAC non-linearity to the envelope (else linear)
	waveDAC  bool    // apply DAC non-linearity to the waveform (else linear)
	fcScale  float64 // multiplies the filter cutoff curve (tuning)
}

// buildDAC reimplements reSID's build_dac_table: the analog output of an N-bit R-2R
// ladder DAC computed by superposition of each bit's voltage contribution. The 6581's
// ladder has 2R/R = 2.20 (not the ideal 2.0) and no termination resistor, which makes it
// measurably non-linear — the source of the chip's characteristic timbre. Returned values
// are normalised so the full-scale code maps to 1.0.
func buildDAC(bits int, r2rDivR float64, term bool) []float64 {
	inf := math.Inf(1)
	vbit := make([]float64, bits)
	for setBit := 0; setBit < bits; setBit++ {
		Vn := 1.0      // normalised bit voltage
		R := 1.0       // normalised R
		_2R := r2rDivR // 2R
		Rn := inf      // ladder "tail" resistance; INF = missing termination
		if term {
			Rn = _2R
		}
		bit := 0
		for ; bit < setBit; bit++ { // collapse the tail below this bit
			if math.IsInf(Rn, 1) {
				Rn = R + _2R
			} else {
				Rn = R + _2R*Rn/(_2R+Rn)
			}
		}
		if math.IsInf(Rn, 1) { // source transformation for this bit
			Rn = _2R
		} else {
			Rn = _2R * Rn / (_2R + Rn)
			Vn = Vn * Rn / _2R
		}
		for bit++; bit < bits; bit++ { // propagate to the output
			Rn += R
			I := Vn / Rn
			Rn = _2R * Rn / (_2R + Rn)
			Vn = Rn * I
		}
		vbit[setBit] = Vn
	}
	n := 1 << bits
	dac := make([]float64, n)
	var max float64
	for i := 0; i < n; i++ {
		var vsum float64
		for sb := 0; sb < bits; sb++ {
			if i&(1<<sb) != 0 {
				vsum += vbit[sb]
			}
		}
		dac[i] = vsum
		if vsum > max {
			max = vsum
		}
	}
	for i := range dac {
		dac[i] /= max
	}
	return dac
}

// envF reads a float from an env var (for tuning sweeps); 0 if unset/invalid.
func envF(k string) float64 {
	v, _ := strconv.ParseFloat(os.Getenv(k), 64)
	return v
}

// New makes a SID clocked at clock Hz, producing samples at sampleRate Hz.
func New(clock, sampleRate float64) *SID {
	s := &SID{clock: clock, srate: sampleRate, cycPerS: clock / sampleRate}
	// anti-alias low-pass at ~19 kHz (just under the output Nyquist), as a one-pole coefficient
	aaHz := 19000.0
	if v := envF("SID_AA"); v > 0 {
		aaHz = v
	}
	s.resScale = 0.42 // gentle resonance: the 6581 filter peak is weak (verified vs reSID)
	if v := envF("SID_RES"); v > 0 {
		s.resScale = v
	}
	wlpHz := 9000.0 // the 6581's waveform DAC rounds the edges; model as a per-voice low-pass
	if v := envF("SID_WLP"); v > 0 {
		wlpHz = v
	}
	s.wlpA = 1.0 - math.Exp(-2.0*3.14159265358979*wlpHz/clock)
	s.aaA = 1.0 - math.Exp(-2.0*3.14159265358979*aaHz/clock)
	dcHz := 110.0 // per-voice DC blocker / output AC-coupling; tames the bass voice's fundamental
	if v := envF("SID_DC"); v > 0 {
		dcHz = v
	}
	s.dcA = 1.0 - math.Exp(-2.0*3.14159265358979*dcHz/clock)
	// 6581 DAC non-linearity (R-2R ladder, 2R/R = 2.20, no termination).
	const r2r6581 = 2.20
	copy(s.dacWave[:], buildDAC(12, r2r6581, false))
	copy(s.dacEnv[:], buildDAC(8, r2r6581, false))
	s.dacWZero = s.dacWave[2048]
	s.envDAC = envF("SID_ENVDAC") != 0  // default 0 => linear envelope; set 1 to enable
	s.waveDAC = envF("SID_NOWAVEDAC") == 0 // default on; set 1 to bypass waveform DAC
	s.fcScale = 1.0
	if v := envF("SID_FC"); v > 0 {
		s.fcScale = v
	}
	for i := range s.v {
		s.v[i].noise = 0x7FFFF8
		s.v[i].expPer = 1
	}
	return s
}

// Write sets one SID register (reg 0..24 == $D400..$D418).
func (s *SID) Write(reg uint8, val uint8) {
	if reg >= 0x15 {
		switch reg {
		case 0x15:
			s.fcLo = val & 7
		case 0x16:
			s.fcHi = val
		case 0x17:
			s.res = val
		case 0x18:
			s.mode = val
		}
		return
	}
	vi := reg / 7
	if vi > 2 {
		return
	}
	v := &s.v[vi]
	switch reg % 7 {
	case 0:
		v.freq = (v.freq & 0xFF00) | uint16(val)
	case 1:
		v.freq = (v.freq & 0x00FF) | uint16(val)<<8
	case 2:
		v.pw = (v.pw & 0x0F00) | uint16(val)
	case 3:
		v.pw = (v.pw & 0x00FF) | uint16(val&0x0F)<<8
	case 4:
		gateBefore := v.ctrl & 1
		v.ctrl = val
		if val&1 != 0 && gateBefore == 0 {
			v.state = stAttack // gate rising edge -> (re)start attack
			v.holdZero = false
			// reset the rate/exponential counters: the envelope rate changes with the
			// state, and a stale counter left above the new (smaller) period would never
			// match again (the counters only test for equality), freezing the envelope.
			v.rateCnt, v.expCnt = 0, 0
		} else if val&1 == 0 && gateBefore != 0 {
			v.state = stRelease // gate falling edge -> release
			v.rateCnt, v.expCnt = 0, 0
		}
		if val&8 != 0 { // test bit resets the oscillator
			v.acc = 0
		}
	case 5:
		v.ad = val
	case 6:
		v.sr = val
	}
}

// clockEnv advances one voice's envelope generator by one system cycle (reSID model).
func (v *voice) clockEnv() {
	v.rateCnt++
	var rate uint32
	switch v.state {
	case stAttack:
		rate = ratePeriod[v.ad>>4]
	case stDecay:
		rate = ratePeriod[v.ad&0x0F]
	case stRelease:
		rate = ratePeriod[v.sr&0x0F]
	}
	if v.rateCnt != rate {
		return
	}
	v.rateCnt = 0
	if v.state == stAttack {
		v.env++
		if v.env == 0xFF {
			v.state = stDecay
			v.expPer, v.expCnt = 1, 0 // env is at max: start decay on the fast exp period
		}
		return
	}
	// decay or release: exponential
	v.expCnt++
	if v.expCnt != v.expPer {
		return
	}
	v.expCnt = 0
	sustain := (v.sr >> 4) * 0x11
	if v.state == stDecay && v.env == sustain {
		return // hold at the sustain level
	}
	if !v.holdZero {
		v.env--
		if v.env == 0 {
			v.holdZero = true
		}
	}
	switch v.env { // exponential curve: slow the step at these levels
	case 0xFF:
		v.expPer = 1
	case 0x5D:
		v.expPer = 2
	case 0x36:
		v.expPer = 4
	case 0x1A:
		v.expPer = 8
	case 0x0E:
		v.expPer = 16
	case 0x06:
		v.expPer = 30
	case 0x00:
		v.expPer = 1
	}
}

// wave returns the 12-bit waveform output (0..4095) of voice vi this cycle.
func (s *SID) wave(vi int) uint16 {
	v := &s.v[vi]
	out := uint16(0xFFF)
	saw := uint16(v.acc >> 12)
	// triangle: fold the ramp around the MSB; ring-mod XORs with the previous voice's MSB
	msb := v.acc
	if v.ctrl&0x04 != 0 { // ring mod
		msb ^= s.v[(vi+2)%3].acc
	}
	tri := uint16(v.acc>>11) & 0xFFF
	if msb&0x800000 != 0 {
		tri ^= 0xFFF
	}
	pulse := uint16(0)
	if v.ctrl&8 != 0 || uint16(v.acc>>12) >= v.pw {
		pulse = 0xFFF
	}
	noise := uint16((v.noise>>4)&0x800 | (v.noise>>3)&0x400 | (v.noise>>2)&0x200 |
		(v.noise>>1)&0x100 | (v.noise<<1)&0x80 | (v.noise<<2)&0x40 |
		(v.noise<<3)&0x20 | (v.noise<<4)&0x10)
	any := false
	if v.ctrl&0x10 != 0 {
		out &= tri
		any = true
	}
	if v.ctrl&0x20 != 0 {
		out &= saw
		any = true
	}
	if v.ctrl&0x40 != 0 {
		out &= pulse
		any = true
	}
	if v.ctrl&0x80 != 0 {
		out &= noise
		any = true
	}
	if !any {
		return 0
	}
	return out
}

// clock advances the whole chip by one system cycle.
func (s *SID) clockCycle() {
	for i := range s.v {
		v := &s.v[i]
		v.accPrev = v.acc
		if v.ctrl&8 == 0 { // test bit halts the oscillator
			v.acc = (v.acc + uint32(v.freq)) & 0xFFFFFF
		}
		// hard sync: reset this voice when the source voice's MSB rises
		if v.ctrl&0x02 != 0 {
			src := &s.v[(i+2)%3]
			if src.acc&0x800000 != 0 && src.accPrev&0x800000 == 0 {
				v.acc = 0
			}
		}
		// noise LFSR clocks on accumulator bit 19 rising
		if v.acc&0x080000 != 0 && v.accPrev&0x080000 == 0 {
			bit0 := ((v.noise >> 22) ^ (v.noise >> 17)) & 1
			v.noise = ((v.noise << 1) | bit0) & 0x7FFFFF
		}
		v.clockEnv()
	}
}

// cutoff maps the 11-bit cutoff value to a normalised filter coefficient. The 6581's
// cutoff curve is nonlinear; this is a smooth approximation over ~30Hz..12kHz.
func (s *SID) cutoffW() float64 {
	fc := float64(uint16(s.fcHi)<<3 | uint16(s.fcLo)) // 0..2047
	// 6581 cutoff curve: rises quickly off the floor, so a mostly-linear map (gently
	// shaped) rather than a steep quadratic — the latter buries mid-range cutoffs too low.
	n := fc / 2047.0
	hz := (100.0 + (12000.0-100.0)*(0.35*n*n+0.65*n)) * s.fcScale
	w := 2.0 * 3.14159265358979 * hz / s.clock
	if w > 1.0 {
		w = 1.0
	}
	return w
}

// output computes the current mixed output sample as a normalised float (-1..1).
func (s *SID) output() float64 {
	var direct, filtIn float64
	route := s.res & 7
	for i := range s.v {
		// voice 3 can be muted (mode bit 7) when it is not routed to the filter
		if i == 2 && s.mode&0x80 != 0 && route&4 == 0 {
			continue
		}
		// pass the 12-bit waveform code through the non-linear DAC and AC-centre it,
		// then scale by the (also non-linear) envelope DAC — the analog multiply.
		var raw float64
		if s.waveDAC {
			raw = (s.dacWave[s.wave(i)] - s.dacWZero) * 2.0
		} else {
			raw = (float64(s.wave(i)) - 2048.0) / 2048.0
		}
		s.vdc[i] += s.dcA * (raw - s.vdc[i]) // track DC; subtract it (AC-couple the voice)
		raw -= s.vdc[i]
		s.vlp[i] += s.wlpA * (raw - s.vlp[i]) // DAC edge-rounding (per-voice low-pass)
		envLvl := float64(s.v[i].env) / 255.0
		if s.envDAC {
			envLvl = s.dacEnv[s.v[i].env]
		}
		samp := s.vlp[i] * envLvl
		if route&(1<<i) != 0 {
			filtIn += samp
		} else {
			direct += samp
		}
	}
	// state-variable filter. The 6581 filter is weak: even at max resonance the peak is
	// modest, so map resonance to a gentle damping range (Q ~1..3) rather than self-oscillation.
	w := s.cutoffW()
	q := 1.0 - float64(s.res>>4)/15.0*s.resScale
	hp := filtIn - s.lp - q*s.bp
	s.bp += w * hp
	if s.bp > 2 {
		s.bp = 2
	} else if s.bp < -2 {
		s.bp = -2
	}
	s.lp += w * s.bp
	var filtered float64
	if s.mode&0x10 != 0 {
		filtered += s.lp
	}
	if s.mode&0x20 != 0 {
		filtered += s.bp
	}
	if s.mode&0x40 != 0 {
		filtered += hp
	}
	vol := float64(s.mode&0x0F) / 15.0
	return (direct + filtered) / 3.0 * vol
}

// Env returns voice v's current envelope level (0..255) — for diagnostics.
func (s *SID) Env(v int) uint8 { return s.v[v].env }

// Gate reports whether voice v's gate bit is set — for diagnostics.
func (s *SID) Gate(v int) bool { return s.v[v].ctrl&1 != 0 }

// Sample advances the chip by one output-sample period and returns the sample. The SID is
// clocked every system cycle and its output run through the anti-alias low-pass; the sample
// is the filtered value at the decimation instant, so the bright waveforms don't alias.
func (s *SID) Sample() int16 {
	s.sAcc += s.cycPerS
	for s.sAcc >= 1.0 {
		s.clockCycle()
		o := s.output()
		s.aa1 += s.aaA * (o - s.aa1)
		s.aa2 += s.aaA * (s.aa1 - s.aa2)
		s.aa3 += s.aaA * (s.aa2 - s.aa3)
		s.aa4 += s.aaA * (s.aa3 - s.aa4)
		s.sAcc -= 1.0
	}
	out := s.aa4
	if out > 1.0 {
		out = 1.0
	} else if out < -1.0 {
		out = -1.0
	}
	return int16(out * 30000.0)
}
