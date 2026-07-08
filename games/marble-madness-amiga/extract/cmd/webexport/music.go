// music.go is the music stage (folded ex-cmd/musicrender): it renders every course's
// theme to music/<key>.mp3 with a from-scratch Go reimplementation of Marble Madness's
// music player (Marble_Madness.md Part VI). It parses each course *Snd bank, finds the
// music entry (an op0 directory record), walks the per-channel note byte-streams of the
// h1 arrangement, synthesises each note Amiga-style from the shared h4 waveform using the
// ProTracker period table, mixes the voices and writes a WAV, then encodes to MP3
// (ffmpeg / libmp3lame). No emulation: the algorithm is reimplemented from the disassembly.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// exportMusic renders every course theme to music/<key>.mp3 and returns the manifest music
// index (display name + file). NO oracle. Deterministic order.
func exportMusic(vol *adf.Volume, paths map[string]string, outDir string) []MusicEntry {
	musicDir := filepath.Join(outDir, "music")
	chk(os.MkdirAll(musicDir, 0o755))

	const secs = 60.0
	var entries []MusicEntry
	for idx, c := range courses {
		sp, ok := paths[strings.ToLower(c.snd)]
		if !ok {
			fail(fmt.Errorf("%s not found on disk", c.snd))
		}
		data, err := vol.ReadFile(sp)
		chk(err)
		pcm := renderSnd(data, secs)

		wav := filepath.Join(musicDir, c.key+".wav")
		sndWriteWAV(wav, pcm, sndOutRate)
		mp3 := filepath.Join(musicDir, c.key+".mp3")
		cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "96k", "-ac", "1", mp3)
		chk(cmd.Run())
		os.Remove(wav)

		fi, _ := os.Stat(mp3)
		entries = append(entries, MusicEntry{Name: c.name, File: "music/" + c.key + ".mp3"})
		fmt.Fprintf(os.Stderr, "[music] %d/%d  %-12s %s (%d KB, %.0fs)\n",
			idx+1, len(courses), c.name, c.key+".mp3", fi.Size()/1024, secs)
	}
	fmt.Fprintf(os.Stderr, "[music] done: %d tracks\n", len(entries))
	return entries
}

// renderSnd synthesises a course *Snd bank to mono 16-bit PCM at sndOutRate.
func renderSnd(data []byte, secs float64) []int16 {
	prog, err := hunk.Load(data, sndBase)
	chk(err)
	s := &sndBank{img: prog.Image}

	var dir uint32
	for _, sg := range prog.Segments {
		if sg.Kind == "DATA" && sg.Size > 0 {
			dir = sg.Base
			break
		}
	}
	cnt := int(s.r16(dir))
	pick := -1
	for i := 0; i < cnt; i++ {
		rec := dir + 2 + uint32(i)*8
		if s.r16(rec) == 0 && s.r32(rec+4) != 0 { // op0 record with a descriptor
			pick = i
			break
		}
	}
	if pick < 0 {
		return nil
	}
	desc := s.r32(dir + 2 + uint32(pick)*8 + 4)
	song := s.r32(desc)         // arrangement
	sub := s.r32(desc + 4)      // instrument bank
	sampBase := s.r32(sub + 4)  // h4 waveform base
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
	total := int(secs * float64(sndOutRate))
	samplesPerFrame := float64(sndOutRate) / fps
	buf := make([]float64, total)
	nextFrame := 0.0
	for i := 0; i < total; i++ {
		if float64(i) >= nextFrame {
			for _, c := range chans {
				tickChannel(s, c, delta)
				c.v.level = c.v.envStep()
			}
			nextFrame += samplesPerFrame
		}
		var mix float64
		for _, c := range chans {
			if c.v.active {
				mix += c.v.sampleAt() * c.v.vol * c.v.level
			}
		}
		buf[i] = mix
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
	pcm := make([]int16, total)
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

func sndWriteWAV(path string, pcm []int16, rate int) {
	f, err := os.Create(path)
	chk(err)
	defer f.Close()
	dataLen := len(pcm) * 2
	w := func(v ...interface{}) {
		for _, x := range v {
			binary.Write(f, binary.LittleEndian, x)
		}
	}
	f.WriteString("RIFF")
	w(uint32(36 + dataLen))
	f.WriteString("WAVE")
	f.WriteString("fmt ")
	w(uint32(16), uint16(1), uint16(1), uint32(rate), uint32(rate*2), uint16(2), uint16(16))
	f.WriteString("data")
	w(uint32(dataLen))
	w(pcm)
}
