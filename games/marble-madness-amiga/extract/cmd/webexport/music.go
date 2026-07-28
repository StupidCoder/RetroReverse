// music.go is the music stage (folded ex-cmd/musicrender): it renders every tune of
// every course *Snd bank with a from-scratch Go reimplementation of Marble Madness's
// music player (Marble_Madness.md Part VI). It parses each bank's directory, walks the
// per-channel note byte-streams of each op0 (music) record's arrangement, synthesises
// each note Amiga-style from the shared h4 waveform using the ProTracker period table,
// mixes the voices and writes a WAV, then encodes to MP3 (ffmpeg / libmp3lame). No
// emulation: the algorithm is reimplemented from the disassembly.
//
// Which record is which (traced in the disassembly):
//   id30  the in-race course theme — the race-start trigger table ($20DC, indexed by
//         g_course) holds 30 for all six courses, so every bank keys its theme to 30.
//   id25  the out-of-time tune — played unconditionally by the game-over sequence
//         ($BBE4, reached from the $B118 handler that then sets the game-over phase).
//         Every bank carries an identical copy so the tune is on hand whichever bank
//         is loaded; it is shipped once (the exporter verifies the copies match).
//   id31/id32  the between-courses score-tally tune — the tally routine ($B9A8) plays
//         the id from the $20F4 table; only AerSnd/UltSnd actually carry the record,
//         so the other courses tally in silence.
// Silly's id33 is a duplicate directory pointer to its id30 and is skipped.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/amiga/hunk"
)

const sndBase = 0x800000
const paulaClock = 3546895.0 // PAL

// sndPeriod is the ProTracker period table for one octave (C..B), the $1F902 table.
var sndPeriod = [12]float64{428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226}

// sndDurTable is the $1FA26 per-"instrument" note-length table (delta units).
var sndDurTable = [16]uint32{65536, 6291456, 3145728, 1572864, 786432, 393216, 196608, 131072,
	4718592, 2359296, 1179648, 589824, 2097152, 1048576, 524288, 262144}

var (
	sndOutRate   = 44100
	sndSynthBase = 1
	sndVolEnv    []envSeg
)

type sndBank struct{ img []byte }

func (s *sndBank) r16(a uint32) uint16 {
	if a < sndBase || int(a-sndBase)+2 > len(s.img) {
		return 0
	}
	return binary.BigEndian.Uint16(s.img[a-sndBase:])
}
func (s *sndBank) r32(a uint32) uint32 {
	if a < sndBase || int(a-sndBase)+4 > len(s.img) {
		return 0
	}
	return binary.BigEndian.Uint32(s.img[a-sndBase:])
}
func (s *sndBank) bytesAt(a uint32, n int) []byte {
	o := int(a - sndBase)
	if a < sndBase || o+n > len(s.img) {
		return nil
	}
	return s.img[o : o+n]
}

// envSeg is one [rate, target] segment of a volume envelope (16.16 fixed).
type envSeg struct{ rate, target int32 }

// voice is one playing note.
type voice struct {
	active bool
	sample []int8  // the looped waveform slice
	pos    float64 // fractional read position
	step   float64 // samples advanced per output sample
	vol    float64 // base level 0..1
	level  float64 // current envelope level 0..1 (updated per frame)

	env    []envSeg
	envVal int64
	envIdx int
}

// envStep advances the envelope one sequencer frame and returns the level 0..1.
func (v *voice) envStep() float64 {
	if v.envIdx < len(v.env) {
		s := v.env[v.envIdx]
		if s.rate != 0 {
			v.envVal += int64(s.rate)
			if (s.rate > 0 && v.envVal >= int64(s.target)) || (s.rate < 0 && v.envVal <= int64(s.target)) {
				v.envVal = int64(s.target)
				v.envIdx++
			}
		}
	}
	if v.envVal < 0 {
		v.envVal = 0
	}
	return float64(v.envVal) / 65536.0
}

func (v *voice) sampleAt() float64 {
	if len(v.sample) == 0 {
		return 0
	}
	idx := int(v.pos) % len(v.sample)
	val := float64(v.sample[idx]) / 128.0
	v.pos += v.step
	return val
}

// chanState is the per-channel sequencer cursor: the 6-byte "events" are the song's
// order table ([repeat:word][pattern:long]); the pattern (a note byte-stream) is played
// `repeat` times before advancing, a repeat of 0 terminates -> loop to 0.
type chanState struct {
	events     uint32
	evIdx      int
	repeat     int
	sub        int
	instr      int
	timer      int64
	sampleBase uint32
	v          voice
	live       bool
}

// bankTune is one op0 (music) directory record of a *Snd bank.
type bankTune struct {
	id   int
	desc uint32
}

// exportMusic renders every tune of every course *Snd bank to music/*.mp3: the six
// in-race themes (id30) in course order, then the shared out-of-time tune (id25,
// shipped once — the six copies are render-compared) and the score-tally tunes
// (id31/id32). NO oracle. Deterministic order.
func exportMusic(ctx *cli.Context, vol *adf.Volume, paths map[string]string) error {
	const maxSecs = 60.0

	type bank struct {
		s     *sndBank
		tunes []bankTune
	}
	banks := make([]bank, len(courses))
	total := 0
	for i, c := range courses {
		sp, ok := paths[strings.ToLower(c.snd)]
		if !ok {
			return fmt.Errorf("%s not found on disk", c.snd)
		}
		data, err := vol.ReadFile(sp)
		if err != nil {
			return err
		}
		s, tunes, err := loadBank(data)
		if err != nil {
			return fmt.Errorf("%s: %w", c.snd, err)
		}
		banks[i] = bank{s, tunes}
		total += len(tunes)
	}

	step := 0
	emit := func(pcm []int16, file string, a schema.Asset) error {
		out, err := ctx.Builder.Path("music", file)
		if err != nil {
			return err
		}
		wave := audio.PCM16{Rate: sndOutRate, Channels: 1, Samples: pcm}
		if err := audio.EncodeMP3(wave, out); err != nil {
			return err
		}
		a.Category = schema.CategoryMusic
		a.File = "music/" + file
		a.Duration = wave.Duration()
		ctx.Builder.AddMedia(a)
		return nil
	}

	// The six in-race themes, in course order.
	for i, c := range courses {
		t, ok := findTune(banks[i].tunes, 30)
		if !ok {
			return fmt.Errorf("%s: no id30 (theme) record", c.snd)
		}
		pcm := renderTune(banks[i].s, t.desc, maxSecs)
		if err := emit(pcm, c.key+".mp3", schema.Asset{
			ID:   c.key + "-theme",
			Name: c.name + " theme",
			Description: "The course's in-race theme. Every sound bank keys its main tune to " +
				"record 30, and the race-start trigger table holds 30 for all six courses.",
		}); err != nil {
			return err
		}
		step++
		ctx.Progress("music", step, total, fmt.Sprintf("%-12s theme (id30)", c.name))
	}

	// The out-of-time tune: id25, an identical copy in every bank, shipped once.
	// Each bank's copy is rendered and compared, so a differing copy would surface
	// as its own asset instead of being silently dropped.
	var first []int16
	for i, c := range courses {
		t, ok := findTune(banks[i].tunes, 25)
		if !ok {
			continue
		}
		pcm := renderTune(banks[i].s, t.desc, maxSecs)
		step++
		switch {
		case first == nil:
			first = pcm
			if err := emit(pcm, "outoftime.mp3", schema.Asset{
				ID:   "outoftime",
				Name: "Out of time",
				Description: "The tune the game-over sequence plays when the clock runs out. " +
					"Every course's sound bank carries an identical copy as record 25, so the tune " +
					"is on hand whichever bank is loaded; it is shipped here once.",
			}); err != nil {
				return err
			}
			ctx.Progress("music", step, total, "Out of time (id25, every bank)")
		case pcmEqual(first, pcm):
			ctx.Progress("music", step, total, fmt.Sprintf("%-12s id25 matches the shared copy", c.name))
		default:
			if err := emit(pcm, c.key+"-25.mp3", schema.Asset{
				ID:          c.key + "-25",
				Name:        c.name + " tune 25",
				Description: "This bank's record 25, which differs from the other banks' shared out-of-time tune.",
			}); err != nil {
				return err
			}
			ctx.Progress("music", step, total, fmt.Sprintf("%-12s id25 DIFFERS from the shared copy", c.name))
		}
	}

	// Whatever else a bank carries: the between-courses score-tally tunes.
	for i, c := range courses {
		for _, t := range banks[i].tunes {
			if t.id == 25 || t.id == 30 {
				continue
			}
			pcm := renderTune(banks[i].s, t.desc, maxSecs)
			name := fmt.Sprintf("%s tune %d", c.name, t.id)
			desc := fmt.Sprintf("An extra tune only this course's bank carries (record %d).", t.id)
			if t.id == 31 || t.id == 32 {
				name = fmt.Sprintf("Score tally (%s bank)", c.name)
				desc = "Played over the between-courses score tally, the time-bonus count-up. " +
					"The tally trigger asks the loaded bank for this tune number; only this course's " +
					"bank carries the record, so the other courses tally in silence."
			}
			if err := emit(pcm, fmt.Sprintf("%s-%d.mp3", c.key, t.id), schema.Asset{
				ID:          fmt.Sprintf("%s-%d", c.key, t.id),
				Name:        name,
				Description: desc,
			}); err != nil {
				return err
			}
			step++
			ctx.Progress("music", step, total, fmt.Sprintf("%-12s id%d", c.name, t.id))
		}
	}
	return nil
}

// loadBank hunk-loads a *Snd file and lists its op0 (music) directory records.
// Records aliasing an already-seen descriptor are dropped (Silly's id33 is a
// duplicate pointer to its id30).
func loadBank(data []byte) (*sndBank, []bankTune, error) {
	prog, err := hunk.Load(data, sndBase)
	if err != nil {
		return nil, nil, err
	}
	s := &sndBank{img: prog.Image}
	var dir uint32
	for _, sg := range prog.Segments {
		if sg.Kind == "DATA" && sg.Size > 0 {
			dir = sg.Base
			break
		}
	}
	cnt := int(s.r16(dir))
	seen := map[uint32]bool{}
	var tunes []bankTune
	for i := 0; i < cnt; i++ {
		rec := dir + 2 + uint32(i)*8
		if s.r16(rec) != 0 {
			continue
		}
		desc := s.r32(rec + 4)
		if desc == 0 || seen[desc] {
			continue
		}
		seen[desc] = true
		tunes = append(tunes, bankTune{id: i, desc: desc})
	}
	return s, tunes, nil
}

func findTune(ts []bankTune, id int) (bankTune, bool) {
	for _, t := range ts {
		if t.id == id {
			return t, true
		}
	}
	return bankTune{}, false
}

func pcmEqual(a, b []int16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// renderTune synthesises one op0 record's song to mono 16-bit PCM at sndOutRate,
// stopping at the song's own end (every channel's order table hits its null-pattern
// terminator) or after maxSecs for the looping scores.
func renderTune(s *sndBank, desc uint32, maxSecs float64) []int16 {
	song := s.r32(desc)        // arrangement
	sub := s.r32(desc + 4)     // instrument bank
	sampBase := s.r32(sub + 4) // h4 waveform base
	sndVolEnv = sndParseEnv(s, s.r32(sub+8))

	var chans []*chanState
	for ch := uint32(0); ch < 8; ch++ {
		evs := s.r32(song + ch*8)
		if evs == 0 {
			continue
		}
		rep := int(int16(s.r16(evs)))
		if rep == 0 {
			rep = 1
		}
		chans = append(chans, &chanState{events: evs, sampleBase: sampBase, repeat: rep, live: true})
	}

	const fps = 50.0
	delta := int64(59419) // $1FA68 runtime value (tempo 99)
	total := int(maxSecs * float64(sndOutRate))
	samplesPerFrame := float64(sndOutRate) / fps
	buf := make([]float64, 0, total)
	nextFrame := 0.0
	for i := 0; i < total; i++ {
		if float64(i) >= nextFrame {
			live := false
			for _, c := range chans {
				tickChannel(s, c, delta)
				c.v.level = c.v.envStep()
				if c.live {
					live = true
				}
			}
			if !live {
				break
			}
			nextFrame += samplesPerFrame
		}
		var mix float64
		for _, c := range chans {
			if c.v.active {
				mix += c.v.sampleAt() * c.v.vol * c.v.level
			}
		}
		buf = append(buf, mix)
	}
	peak := 0.0
	for _, x := range buf {
		if a := math.Abs(x); a > peak {
			peak = a
		}
	}
	if peak > 0 {
		g := 0.92 / peak
		for i := range buf {
			buf[i] *= g
		}
	}
	pcm := make([]int16, len(buf))
	for i, x := range buf {
		if x > 1 {
			x = 1
		}
		if x < -1 {
			x = -1
		}
		pcm[i] = int16(math.Round(x * 32767))
	}
	return pcm
}

// tickChannel advances one channel's sequencer by one frame.
func tickChannel(s *sndBank, c *chanState, delta int64) {
	c.timer -= delta
	if c.timer > 0 {
		return
	}
	for guard := 0; guard < 64; guard++ {
		ev := c.events + uint32(c.evIdx)*6
		ns := s.r32(ev + 2)
		if ns == 0 {
			c.live = false
			c.v.active = false
			return
		}
		b := s.img[ns-sndBase+uint32(c.sub)]
		c.sub++
		if b&0x80 != 0 {
			if b&0x0F == 0 { // $80: end of this pattern instance
				c.sub = 0
				c.repeat--
				if c.repeat > 0 {
					continue
				}
				c.evIdx++
				rep := int(int16(s.r16(c.events + uint32(c.evIdx)*6)))
				if rep == 0 {
					c.evIdx = 0
					rep = int(int16(s.r16(c.events)))
				}
				c.repeat = rep
				continue
			}
			c.instr = int(b & 0x0F)
			continue
		}
		c.timer += int64(sndDurTable[c.instr&0xF])
		if b == 0x7F { // rest / note-off
			c.v.active = false
			return
		}
		triggerNote(s, c, int(b))
		return
	}
}

// triggerNote synthesises a note from the h4 waveform.
func triggerNote(s *sndBank, c *chanState, note int) {
	octave := sndClamp(note/12, 0, 8)
	semi := note % 12
	lenWords := sndSynthBase << (8 - octave)
	lenBytes := lenWords * 2
	if lenBytes < 2 {
		lenBytes = 2
	}
	off := lenBytes
	wav := s.bytesAt(c.sampleBase+uint32(off), lenBytes)
	if wav == nil {
		wav = s.bytesAt(c.sampleBase, lenBytes)
	}
	if wav == nil {
		return
	}
	smp := make([]int8, len(wav))
	for i, bb := range wav {
		smp[i] = int8(bb)
	}
	per := sndPeriod[semi]
	srcRate := paulaClock / per
	c.v = voice{active: true, sample: smp, step: srcRate / float64(sndOutRate), vol: 0.45, env: sndVolEnv}
}

// sndParseEnv reads the engine's volume-envelope segments ([rate:long][target:long] pairs,
// 16.16 fixed) starting at addr, stopping after a rate==0 (sustain) segment.
func sndParseEnv(s *sndBank, addr uint32) []envSeg {
	if addr == 0 {
		return nil
	}
	var segs []envSeg
	for i := 0; i < 16; i++ {
		rate := int32(s.r32(addr + uint32(i*8)))
		target := int32(s.r32(addr + uint32(i*8) + 4))
		segs = append(segs, envSeg{rate, target})
		if rate == 0 {
			break
		}
	}
	return segs
}

func sndClamp(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
