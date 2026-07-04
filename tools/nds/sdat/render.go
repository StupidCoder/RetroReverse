package sdat

import (
	"fmt"
	"math"
)

// The renderer plays one SSEQ through its SBNK/SWARs into stereo float PCM,
// following the NitroSDK sound driver's semantics: 48 ticks per quarter note
// (so ticks/second = tempo × 48/60), envelopes stepped at the driver's 192 Hz
// frame rate in decibel units (0 = full, -92544 = silence — the scale is
// dB × 1024: 92544/1024 ≈ the -90.4 dB floor of 16-bit audio), and the DS
// envelope rate laws: attack multiplies the (negative) envelope by table[a]/255
// each frame; decay/release subtract 0x1E00/(126-d) per frame; sustain holds at
// a dB level mapped from its 0-127 value.
type renderer struct {
	sdat   *SDAT
	bank   *Bank
	swars  [4][]Wave
	seq    []byte
	tracks []*track
	voices []*voice

	rate    float64 // output sample rate
	tempo   float64 // BPM (quarter = 48 ticks)
	tickss  float64 // seconds per tick
	seqVol  float64 // INFO volume 0-127
	elapsed float64
	loops   int // backward jumps taken (loop detection)
}

type track struct {
	pc        int
	stack     []int
	loopPC    [4]int
	loopN     [4]int
	loopDepth int
	done      bool
	wait      float64 // seconds until next command

	program   int
	vol       int // 0-127
	expr      int // 0-127
	pan       int // 0-127 (64 = center)
	bend      int // -128..127
	bendRange int
	transpose int
	noteWait  bool
	tie       bool
	modDepth  int
	modSpeed  int
}

type voice struct {
	on      bool
	release bool
	region  *Region
	wave    *Wave
	trk     *track

	pos, step float64 // sample cursor (wave rate → out rate, pitch applied)
	key, vel  int
	remaining float64 // seconds until note-off (-1 = held)

	env     float64 // dB×1024, 0 = full
	attack  float64
	decay   float64
	sustain float64
	relRate float64

	// PSG state
	psgPhase float64
	lfsr     uint16

	pan     int
	modT    float64
}

// driver attack table for values ≥ 109 (below that the rate is 255-a).
var attackTable = [19]int{
	243, 218, 215, 210, 205, 199, 192, 184, 176, 166,
	156, 144, 131, 116, 100, 82, 63, 42, 21,
}

func attackRate(a int) float64 {
	if a >= 109 {
		return float64(attackTable[a-109])
	}
	return float64(255 - a)
}

func decayRate(d int) float64 {
	if d >= 127 {
		return 92544 // instant
	}
	return float64(0x1E00) / float64(126-d)
}

func sustainLevel(s int) float64 {
	if s == 0 {
		return -92544
	}
	// dB×1024 for a linear 0-127 level
	return 1024 * 20 * math.Log10(float64(s)/127)
}

func dbGain(env float64) float64 { return math.Pow(10, env/1024/20) }

// linGain maps a linear 0-127 control to amplitude (the driver folds these
// into the same dB pipeline; multiplying linear gains is equivalent).
func linGain(v int) float64 { return float64(v) / 127 }

// Render plays sequence seqIdx to stereo PCM at rate Hz. It stops at the end of
// the sequence, or after the sequence has looped loops times (backward jumps),
// or at maxSec. Returns left/right channels.
func (s *SDAT) Render(seqIdx int, rate float64, loops int, maxSec float64) ([]float64, []float64, error) {
	if seqIdx < 0 || seqIdx >= len(s.Seqs) || s.Seqs[seqIdx].FileID < 0 {
		return nil, nil, fmt.Errorf("sdat: no sequence %d", seqIdx)
	}
	info := s.Seqs[seqIdx]
	seqFile := s.File(info.FileID)
	if len(seqFile) < 0x1C+4 || string(seqFile[:4]) != "SSEQ" {
		return nil, nil, fmt.Errorf("sdat: sequence %d: bad SSEQ", seqIdx)
	}
	dataOff := int(le.Uint32(seqFile[0x18:]))
	code := seqFile[dataOff:]

	bankInfo := s.Banks[info.Bank]
	bank, err := ParseSBNK(s.File(bankInfo.FileID))
	if err != nil {
		return nil, nil, err
	}
	r := &renderer{
		sdat: s, bank: bank, seq: code, rate: rate,
		tempo: 120, seqVol: linGain(info.Vol),
	}
	for k, wi := range bankInfo.Swars {
		if wi >= 0 && wi < len(s.Wavearcs) {
			waves, err := ParseSWAR(s.File(s.Wavearcs[wi].FileID))
			if err != nil {
				return nil, nil, err
			}
			r.swars[k] = waves
		}
	}
	r.tracks = []*track{newTrack(0)}

	// A sequence may open with 0xFE (track mask) + 0x93 open-track commands.
	var L, R []float64
	frame := 1.0 / 192 // driver envelope frame
	acc := 0.0
	for r.elapsed < maxSec && (r.loops < loops || loops <= 0) {
		if r.stepTracks() {
			break // all tracks ended
		}
		// advance audio to the next tick boundary in 192 Hz envelope frames
		dt := r.ticksSeconds()
		for dt > 0 {
			h := math.Min(dt, frame-acc)
			n := int(math.Round(h * rate))
			l, rr := r.mix(n)
			L = append(L, l...)
			R = append(R, rr...)
			acc += h
			if acc >= frame-1e-9 {
				acc = 0
				r.stepEnvelopes()
			}
			dt -= h
			r.elapsed += h
		}
	}
	// let releases ring out briefly
	for i := 0; i < 192; i++ {
		n := int(math.Round(rate / 192))
		l, rr := r.mix(n)
		L = append(L, l...)
		R = append(R, rr...)
		r.stepEnvelopes()
	}
	return L, R, nil
}

func newTrack(pc int) *track {
	return &track{pc: pc, vol: 100, expr: 127, pan: 64, bendRange: 2, noteWait: true}
}

func (r *renderer) ticksSeconds() float64 { return 60 / (r.tempo * 48) }

// stepTracks runs every track that is due, advancing one tick. Returns true
// when all tracks have ended and no voices remain.
func (r *renderer) stepTracks() bool {
	alive := false
	for _, t := range r.tracks {
		if t.done {
			continue
		}
		alive = true
		t.wait -= 1
		for t.wait <= 0 && !t.done {
			r.exec(t)
		}
	}
	if alive {
		return false
	}
	for _, v := range r.voices {
		if v.on {
			return false
		}
	}
	return true
}

func readVar(code []byte, pc *int) int {
	v := 0
	for {
		b := code[*pc]
		*pc++
		v = v<<7 | int(b&0x7F)
		if b&0x80 == 0 {
			return v
		}
	}
}

func (r *renderer) exec(t *track) {
	code := r.seq
	if t.pc >= len(code) {
		t.done = true
		return
	}
	cmd := code[t.pc]
	t.pc++
	switch {
	case cmd < 0x80: // note
		vel := int(code[t.pc])
		t.pc++
		dur := readVar(code, &t.pc)
		r.noteOn(t, int(cmd)+t.transpose, vel, dur)
		if t.noteWait {
			t.wait += float64(dur)
		}
	case cmd == 0x80: // rest
		t.wait += float64(readVar(code, &t.pc))
	case cmd == 0x81: // program change
		t.program = readVar(code, &t.pc) & 0x7F
	case cmd == 0xFE: // multitrack mask (only valid at start)
		mask := int(le.Uint16(code[t.pc:]))
		t.pc += 2
		_ = mask
	case cmd == 0x93: // open track {u8 id, u24 offset}
		t.pc++
		off := int(code[t.pc]) | int(code[t.pc+1])<<8 | int(code[t.pc+2])<<16
		t.pc += 3
		nt := newTrack(off)
		r.tracks = append(r.tracks, nt)
	case cmd == 0x94: // jump u24
		off := int(code[t.pc]) | int(code[t.pc+1])<<8 | int(code[t.pc+2])<<16
		t.pc += 3
		if off <= t.pc {
			r.loops++ // backward jump = the music's loop point
		}
		t.pc = off
	case cmd == 0x95: // call u24
		off := int(code[t.pc]) | int(code[t.pc+1])<<8 | int(code[t.pc+2])<<16
		t.pc += 3
		t.stack = append(t.stack, t.pc)
		t.pc = off
	case cmd == 0xFD: // return
		if n := len(t.stack); n > 0 {
			t.pc = t.stack[n-1]
			t.stack = t.stack[:n-1]
		} else {
			t.done = true
		}
	case cmd == 0xC0:
		t.pan = int(code[t.pc])
		t.pc++
	case cmd == 0xC1:
		t.vol = int(code[t.pc])
		t.pc++
	case cmd == 0xC2: // master volume
		t.pc++
	case cmd == 0xC3:
		t.transpose = int(int8(code[t.pc]))
		t.pc++
	case cmd == 0xC4:
		t.bend = int(int8(code[t.pc]))
		t.pc++
		r.retune(t)
	case cmd == 0xC5:
		t.bendRange = int(code[t.pc])
		t.pc++
		r.retune(t)
	case cmd == 0xC6: // priority
		t.pc++
	case cmd == 0xC7:
		t.noteWait = code[t.pc] != 0
		t.pc++
	case cmd == 0xC8:
		t.tie = code[t.pc] != 0
		t.pc++
	case cmd == 0xC9: // portamento key
		t.pc++
	case cmd == 0xCA:
		t.modDepth = int(code[t.pc])
		t.pc++
	case cmd == 0xCB:
		t.modSpeed = int(code[t.pc])
		t.pc++
	case cmd == 0xCC, cmd == 0xCD, cmd == 0xCE, cmd == 0xCF: // mod type/range, porta
		t.pc++
	case cmd == 0xD0, cmd == 0xD1, cmd == 0xD2, cmd == 0xD3: // ADSR overrides (rare)
		t.pc++
	case cmd == 0xD4: // loop start (count)
		if t.loopDepth < 4 {
			t.loopN[t.loopDepth] = int(code[t.pc])
			t.loopPC[t.loopDepth] = t.pc + 1
			t.loopDepth++
		}
		t.pc++
	case cmd == 0xD5:
		t.expr = int(code[t.pc])
		t.pc++
	case cmd == 0xD6: // print var
		t.pc++
	case cmd == 0xDC: // mod delay
		t.pc += 2
	case cmd == 0xE0: // mod delay u16 (alias)
		t.pc += 2
	case cmd == 0xE1: // tempo
		r.tempo = float64(le.Uint16(code[t.pc:]))
		t.pc += 2
	case cmd == 0xE3: // sweep pitch
		t.pc += 2
	case cmd == 0xFC: // loop end
		if t.loopDepth > 0 {
			d := t.loopDepth - 1
			if t.loopN[d] == 0 { // infinite: treat as sequence loop
				r.loops++
				t.pc = t.loopPC[d]
			} else if t.loopN[d] > 1 {
				t.loopN[d]--
				t.pc = t.loopPC[d]
			} else {
				t.loopDepth--
			}
		}
	case cmd == 0xFF:
		t.done = true
	case cmd >= 0xA0 && cmd <= 0xBF: // variable/conditional ops (unused in music)
		// 0xA0-0xA2 prefix ops wrap another command; the B0 block takes u8 var + s16
		if cmd >= 0xB0 {
			t.pc += 3
		}
	default:
		// unknown one-byte command: skip
	}
}

func (r *renderer) noteOn(t *track, key, vel, dur int) {
	if t.program >= len(r.bank.Instruments) {
		return
	}
	reg := r.bank.Instruments[t.program].RegionFor(key)
	if reg == nil || reg.Type == InstNone {
		return
	}
	v := &voice{
		on: true, region: reg, trk: t, key: key, vel: vel,
		env: -92544, lfsr: 0x7FFF,
		attack:  attackRate(reg.Attack),
		decay:   decayRate(reg.Decay),
		sustain: sustainLevel(reg.Sustain),
		relRate: decayRate(reg.Release),
		pan:     reg.Pan,
	}
	if reg.Pan == 64 || reg.Pan == 0x40 {
		v.pan = t.pan
	}
	if dur > 0 {
		v.remaining = float64(dur) * r.ticksSeconds()
	} else {
		v.remaining = -1
	}
	if reg.Type == InstPCM {
		waves := r.swars[reg.Swar&3]
		if reg.Swav >= len(waves) {
			return
		}
		v.wave = &waves[reg.Swav]
	}
	r.tune(v)
	// reuse a dead slot
	for i, old := range r.voices {
		if !old.on {
			r.voices[i] = v
			return
		}
	}
	r.voices = append(r.voices, v)
}

// tune sets the playback step from key, base note, bend and wave rate.
func (r *renderer) tune(v *voice) {
	semis := float64(v.key-v.region.BaseNote) +
		float64(v.trk.bend)/64*float64(v.trk.bendRange)/2
	f := math.Pow(2, semis/12)
	switch v.region.Type {
	case InstPCM:
		v.step = float64(v.wave.Rate) * f / r.rate
	case InstPSGPulse, InstPSGNoise:
		// PSG base: key 60 = middle C (261.63 Hz)
		v.step = 261.6256 * f / r.rate
	}
}

func (r *renderer) retune(t *track) {
	for _, v := range r.voices {
		if v.on && v.trk == t {
			r.tune(v)
		}
	}
}

// stepEnvelopes advances every voice's ADSR one 192 Hz driver frame.
func (r *renderer) stepEnvelopes() {
	for _, v := range r.voices {
		if !v.on {
			continue
		}
		if v.release {
			v.env -= v.relRate
			if v.env <= -92544 {
				v.on = false
			}
			continue
		}
		if v.env < 0 && v.attack > 0 { // attack phase: multiply toward 0
			v.env = v.env * v.attack / 255
			if v.env > -1 {
				v.env = 0
			}
			if v.env < -92544+1 && v.attack <= 21 {
				// pathological zero-attack: snap
				v.env = 0
			}
			continue
		}
		if v.env > v.sustain { // decay
			v.env -= v.decay
			if v.env < v.sustain {
				v.env = v.sustain
			}
		}
	}
}

var dutyTable = [8]float64{0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875, 0.5}

// mix renders n output samples of all voices.
func (r *renderer) mix(n int) ([]float64, []float64) {
	L := make([]float64, n)
	R := make([]float64, n)
	dt := 1 / r.rate
	for _, v := range r.voices {
		if !v.on {
			continue
		}
		t := v.trk
		gain := r.seqVol * linGain(t.vol) * linGain(t.expr) * linGain(v.vel) * 0.6
		pan := float64(v.pan) / 127
		gl, gr := math.Sqrt(1-pan), math.Sqrt(pan)
		for i := 0; i < n; i++ {
			if v.remaining >= 0 {
				v.remaining -= dt
				if v.remaining < 0 && !v.release {
					v.release = true
				}
			}
			e := dbGain(v.env)
			var s float64
			switch v.region.Type {
			case InstPCM:
				w := v.wave
				p := int(v.pos)
				if p >= len(w.Samples) {
					if w.Loop {
						span := float64(len(w.Samples) - w.LoopStart)
						for v.pos >= float64(len(w.Samples)) {
							v.pos -= span
						}
						p = int(v.pos)
					} else {
						v.on = false
						break
					}
				}
				s = w.Samples[p]
				v.pos += v.step
			case InstPSGPulse:
				if v.psgPhase-math.Floor(v.psgPhase) < dutyTable[v.region.Swav&7] {
					s = 0.5
				} else {
					s = -0.5
				}
				v.psgPhase += v.step * 4 // pulse steps 4× (empirical octave match)
			case InstPSGNoise:
				if v.lfsr&1 != 0 {
					v.lfsr = v.lfsr>>1 ^ 0x6000
					s = -0.5
				} else {
					v.lfsr >>= 1
					s = 0.5
				}
			}
			L[i] += s * e * gain * gl
			R[i] += s * e * gain * gr
		}
	}
	return L, R
}
