package audio

// synth.go renders a Song through a Bank to stereo PCM: a Type-0 sequence
// interpreter drives a sample-playback voice per note. It is the offline
// equivalent of the game's audio thread plus RSP mixer — the sequencer walks
// each channel's event stream, and each note-on spawns a voice that decodes its
// VADPCM sample (via vadpcm.go), resamples it to the played pitch, shapes it
// with the instrument's ADSR envelope, and mixes it in. Notes carry an explicit
// gate duration in the stream, so there are no separate note-offs.
//
// Fidelity is verified against the game's own render of a song, captured from
// the oracle's Audio Interface (bootoracle -pcmdump).

import "math"

// masterGain keeps summed polyphony inside range; the game's mixer applies a
// comparable headroom scale.
const masterGain = 0.35

// Player renders songs of one bank/sample-table pair at a fixed output rate.
type Player struct {
	bank  *Bank
	tbl   []byte
	rate  float64
	cache map[int32][]int16 // decoded samples, keyed by wavetable base
}

// NewPlayer builds a renderer for a bank and its ".TBL" at the given output rate.
func NewPlayer(bank *Bank, tbl []byte, rate float64) *Player {
	return &Player{bank: bank, tbl: tbl, rate: rate, cache: map[int32][]int16{}}
}

// sample decodes (and caches) a wavetable's PCM.
func (p *Player) sample(w *WaveTable) []int16 {
	if s, ok := p.cache[w.Base]; ok {
		return s
	}
	var s []int16
	if w.Type == waveADPCM && w.Book != nil && int(w.Base)+int(w.Len) <= len(p.tbl) {
		s = DecodeADPCM(p.tbl[w.Base:int(w.Base)+int(w.Len)], w.Book)
	}
	p.cache[w.Base] = s
	return s
}

// track is one channel's playback cursor.
type track struct {
	data    []byte
	pc      int
	status  byte
	wait    int // ticks until the next event fires
	done    bool
	channel int

	program int
	vol     float64 // channel volume (CC 7), 0..1
	expr    float64 // expression (CC 11), 0..1
	pan     float64 // 0..1, 0.5 = centre
	bend    float64 // pitch bend in semitones

	loopPC int  // stream position of the loop start
	looped bool // loop start seen
}

// voice is one sounding note.
type voice struct {
	smpl               []int16
	pos, step          float64
	loopStart, loopEnd int
	hasLoop            bool

	trk      *track // for live channel volume/expression
	gate     int    // ticks until note-off (the note's stored duration)
	released bool

	env     float64 // current envelope amplitude, 0..1
	phase   int     // 0 attack, 1 decay, 2 sustain, 3 release
	atkRate float64 // per-sample env deltas
	decRate float64
	relRate float64
	sustain float64
	peak    float64

	gainL, gainR float64
}

// Render plays a song to stereo float PCM. loops caps how many times a looping
// song repeats before stopping; maxSec is a hard length cap.
func (p *Player) Render(song *Song, loops int, maxSec float64) (L, R []float64) {
	ppqn := float64(song.Division)
	tempo := 500000.0 // µs per quarter, until a tempo meta says otherwise
	maxSamples := int(maxSec * p.rate)

	tracks := make([]*track, 0, 16)
	for ch := 0; ch < 16; ch++ {
		if song.Track[ch] == 0 {
			continue
		}
		t := &track{data: song.Data, pc: song.Track[ch], channel: ch,
			vol: 1, expr: 1, pan: 0.5}
		t.wait = readVLQ(t.data, &t.pc) // first delta
		tracks = append(tracks, t)
	}

	var voices []*voice
	frac := 0.0
	loopCount := 0
	ringOutStart := 0

	for {
		// Fire every track whose wait has elapsed, chaining zero-delta events.
		active := false
		for _, t := range tracks {
			if t.done {
				continue
			}
			active = true
			if t.wait > 0 {
				t.wait--
				continue
			}
			for t.wait == 0 && !t.done {
				if nt := p.step(t, &tempo, &voices); nt {
					loopCount++
				}
			}
		}
		// Count down note gates; expired notes enter release.
		for _, v := range voices {
			if !v.released && v.gate > 0 {
				v.gate--
				if v.gate == 0 {
					v.released = true
					v.phase = 3
				}
			}
		}
		if active {
			ringOutStart = len(L)
		} else if !anyAudible(voices) || len(L) > ringOutStart+int(2*p.rate) {
			// All tracks ended; let notes ring out at most ~2s — looping samples
			// with slow releases would otherwise sound indefinitely.
			break
		}
		if loops > 0 && loopCount >= loops*countLooping(tracks) && countLooping(tracks) > 0 {
			// Enough repeats: let held notes ring out, then stop.
			for _, t := range tracks {
				t.done = true
			}
		}

		// Render the samples that fall in this tick.
		spt := p.rate * (tempo / 1e6) / ppqn
		frac += spt
		n := int(frac)
		frac -= float64(n)
		for i := 0; i < n; i++ {
			l, r := p.mix(voices)
			L = append(L, l)
			R = append(R, r)
		}
		voices = compact(voices)
		if maxSamples > 0 && len(L) >= maxSamples {
			break
		}
	}
	return L, R
}

// step executes one event on a track and reads the following delta into wait.
// It returns true when the track loops (hits an FF 2E end-with-loop).
func (p *Player) step(t *track, tempo *float64, voices *[]*voice) (looped bool) {
	d := t.data
	if t.pc >= len(d) {
		t.done = true
		return
	}
	b := d[t.pc]
	if b == 0xFE {
		// Loop control (exact operands pending the player disasm); consume the
		// event plus two payload bytes and note that this track loops.
		t.pc++
		t.loopPC = t.pc + 2
		t.looped = true
		t.pc += 2
		if t.pc >= len(d) {
			t.done = true
			return
		}
		t.wait = readVLQ(d, &t.pc)
		return
	}
	if b >= 0x80 {
		t.status = b
		t.pc++
	}
	switch {
	case t.status == 0xFF:
		typ := d[t.pc]
		t.pc++
		switch typ {
		case 0x2F: // end of track
			t.done = true
			return
		case 0x2E: // end of track, looping
			t.done = true
			return true
		case 0x51: // tempo, 3 bytes, no length
			*tempo = float64(int(d[t.pc])<<16 | int(d[t.pc+1])<<8 | int(d[t.pc+2]))
			t.pc += 3
		default:
			t.done = true
			return
		}
	default:
		switch t.status & 0xF0 {
		case 0x90: // note on: note, velocity, duration (VLQ)
			note := d[t.pc]
			vel := d[t.pc+1]
			t.pc += 2
			dur := readVLQ(d, &t.pc)
			p.noteOn(t, note, vel, dur, voices)
		case 0x80, 0xA0, 0xE0: // note off / aftertouch / bend: 2 data bytes
			if t.status&0xF0 == 0xE0 {
				bend := (int(d[t.pc]) | int(d[t.pc+1])<<7) - 0x2000
				t.bend = float64(bend) / 8192.0 * 2 // ±2 semitones
			}
			t.pc += 2
		case 0xB0: // control change
			ctrl, val := d[t.pc], d[t.pc+1]
			t.pc += 2
			switch ctrl {
			case 7:
				t.vol = float64(val) / 127
			case 10:
				t.pan = float64(val) / 127
			case 11:
				t.expr = float64(val) / 127
			}
		case 0xC0: // program change
			t.program = int(d[t.pc])
			t.pc++
		case 0xD0: // channel pressure
			t.pc++
		default:
			t.done = true
			return
		}
	}
	if t.pc >= len(d) {
		t.done = true
		return
	}
	t.wait = readVLQ(d, &t.pc)
	return
}

// noteOn spawns a voice for a note, choosing the instrument by program (channel
// 9 uses the bank's percussion program, keyed by note).
func (p *Player) noteOn(t *track, note, vel byte, dur int, voices *[]*voice) {
	var inst *Instrument
	if t.channel == 9 && p.bank.Percussion != nil {
		inst = p.bank.Percussion
	} else if t.program < len(p.bank.Instruments) {
		inst = p.bank.Instruments[t.program]
	}
	if inst == nil {
		return
	}
	snd := inst.SoundFor(note, vel)
	if snd == nil || snd.Wave == nil {
		return
	}
	smpl := p.sample(snd.Wave)
	if len(smpl) == 0 {
		return
	}
	v := &voice{smpl: smpl, gate: dur + 1, trk: t}
	// Pitch: the sample sounds at keyBase when played at the bank rate; shift by
	// the interval to the played note, plus keymap detune and channel bend.
	semis := 0.0
	if km := snd.KeyMap; km != nil {
		semis = float64(int(note)-int(km.KeyBase)) + float64(km.Detune)/100
	}
	semis += t.bend
	v.step = float64(p.bank.SampleRate) / p.rate * math.Pow(2, semis/12)
	if snd.Wave.Loop != nil {
		v.hasLoop = true
		v.loopStart = int(snd.Wave.Loop.Start)
		v.loopEnd = int(snd.Wave.Loop.End)
	}
	// Envelope: linear amplitude ramps from the ALEnvelope times (µs).
	env := snd.Env
	amp := float64(vel) / 127 * float64(inst.Volume) / 127 * float64(snd.Vol) / 127
	if env != nil {
		v.peak = float64(env.AttackVolume) / 127 * amp
		v.sustain = float64(env.DecayVolume) / 127 * amp
		v.atkRate = ratePerSample(v.peak, env.AttackTime, p.rate)
		v.decRate = ratePerSample(v.peak-v.sustain, env.DecayTime, p.rate)
		v.relRate = ratePerSample(v.sustain, env.ReleaseTime, p.rate)
	} else {
		v.peak, v.sustain, amp = amp, amp, amp
		v.atkRate, v.decRate, v.relRate = amp, 0, amp/2205
	}
	// Pan: instrument pan combined with the per-sound pan, equal-power.
	pan := t.pan
	if snd.Pan != 0 {
		pan = float64(snd.Pan) / 127
	}
	v.gainL = math.Cos(pan * math.Pi / 2)
	v.gainR = math.Sin(pan * math.Pi / 2)
	*voices = append(*voices, v)
}

// mix advances every voice one output sample and returns the stereo sum.
func (p *Player) mix(voices []*voice) (l, r float64) {
	for _, v := range voices {
		if v.smpl == nil {
			continue
		}
		// Envelope.
		switch v.phase {
		case 0:
			v.env += v.atkRate
			if v.env >= v.peak {
				v.env, v.phase = v.peak, 1
			}
		case 1:
			v.env -= v.decRate
			if v.env <= v.sustain {
				v.env, v.phase = v.sustain, 2
			}
		case 3:
			v.env -= v.relRate
			if v.env <= 0 {
				v.env, v.smpl = 0, nil
				continue
			}
		}
		// Sample read (nearest-neighbour) with looping.
		idx := int(v.pos)
		if idx >= len(v.smpl) {
			if v.hasLoop && v.loopEnd > v.loopStart {
				v.pos -= float64(v.loopEnd - v.loopStart)
				idx = int(v.pos)
			} else {
				v.smpl = nil
				continue
			}
		}
		if v.hasLoop && idx >= v.loopEnd {
			v.pos -= float64(v.loopEnd - v.loopStart)
			idx = int(v.pos)
		}
		if idx < 0 || idx >= len(v.smpl) {
			continue
		}
		// Live channel volume and expression scale the note; a master gain keeps
		// the summed polyphony from clipping.
		cv := v.trk.vol * v.trk.expr
		s := float64(v.smpl[idx]) / 32768 * v.env * cv * masterGain
		l += s * v.gainL
		r += s * v.gainR
		v.pos += v.step
	}
	return l, r
}

func ratePerSample(delta float64, timeUS int32, rate float64) float64 {
	sec := float64(timeUS) / 1e6
	if sec <= 0 {
		return delta // instantaneous
	}
	return delta / (sec * rate)
}

func anyAudible(voices []*voice) bool {
	for _, v := range voices {
		if v.smpl != nil {
			return true
		}
	}
	return false
}

func compact(voices []*voice) []*voice {
	out := voices[:0]
	for _, v := range voices {
		if v.smpl != nil {
			out = append(out, v)
		}
	}
	return out
}

func countLooping(tracks []*track) int {
	n := 0
	for _, t := range tracks {
		if t.looped {
			n++
		}
	}
	return n
}

// readVLQ reads a MIDI variable-length quantity, advancing *pc.
func readVLQ(d []byte, pc *int) int {
	v := 0
	for *pc < len(d) {
		c := d[*pc]
		*pc++
		v = v<<7 | int(c&0x7f)
		if c&0x80 == 0 {
			break
		}
	}
	return v
}
