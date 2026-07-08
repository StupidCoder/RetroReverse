// music.go folds the pure-ROM music synth (ex-cmd/musicrom) into webexport: it renders the
// 7 zone themes to music/<name>.mp3 straight from the Sonic (GG) sound-driver data (bank 3,
// Sonic.md Part VI) — NO oracle / machine boot.
//
//   - The song-pointer table $4716 maps a song id to a song base; the header is five relative
//     channel offsets. Channels 0-2 are square tones, channel 3 noise.
//   - Each channel is a byte stream decoded by $42F4: a byte < $70 is a NOTE — (octave<<4)|note
//     — followed by a duration; $7F is a rest; $71-$7E pick a voice; >= $80 are commands.
//   - Per frame the channel renders period = (freq>>octave) + detune + vibrato under an ADSR
//     envelope; the result drives the PSG. The loop is the data's own $FF/$88 structure.
//
// To avoid clashing with main.go's ROM helpers (w/rom operate on file offsets there), this
// file uses its own bank-3 globals: mrom + mw/mrb/mfo, and mtick/mgTempo.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	sr     = 44100
	fps    = 60
	perFrm = sr / fps
)

var mrom []byte

// mfo: bank3 z80 address -> ROM file offset (bank 3 = file $C000, mapped into slot 1).
func mfo(z int) int { return 0xC000 + (z - 0x4000) }

func mw(z int) int   { return int(mrom[mfo(z)]) | int(mrom[mfo(z)+1])<<8 }
func mrb(z int) byte { return mrom[mfo(z)] }

// mfreq $44D5: note index -> 10-bit period (highest octave).
func mfreq(n int) int { return mw(0x44D5 + n*2) }

var psgAtten = func() [16]float64 { // 4-bit attenuation -> linear amplitude (2 dB/step)
	var t [16]float64
	a := 1.0
	for i := 0; i < 15; i++ {
		t[i] = a
		a *= 0.79432823
	}
	return t
}()

// channel mirrors the driver's per-channel work area (the fields we need).
type channel struct {
	pos      int  // current data pointer (z80 addr)
	loop     int  // loop-start ($88)
	dur      int  // duration counter (counts down by tick/frame)
	tempo    int  // note-length multiplier ($80; 0 -> global)
	defDur   int  // default duration ($8A)
	baseFreq int  // freqtable[note] for the current note (NOT yet octave-shifted)
	octave   int  // high nibble of the note byte; period = (baseFreq+detune+vib) >> octave
	detune   int  // $84
	vol      int  // base volume 0-15 ($81/$8B/$8C)
	silent   bool // rest / ended
	active   bool
	noise    bool // channel 3
	noiseMd  int  // noise mode ($89)
	tie      bool // $8D: don't retrigger the envelope on the next note
	end      bool
	loops    int   // count of $FF loop jumps executed
	base     int   // song base (IX+41/42), for $87 relative jump targets
	stack    []int // $86/$87 repeat-counter stack

	env      [6]byte // $82 ADSR params
	envLevel int     // 0-255
	envPhase int     // 0 attack,1 decay,2 decay2,3 release

	vib          [5]byte // $83 vibrato params
	vibVal       int     // running vibrato offset (signed)
	vibDelay     int     // frames before vibrato starts
	vibSpeedCtr  int     // counts down to the next vibrato step
	vibSpeed     int     // reload for the speed counter
	vibDepthCtr  int     // steps left before the direction flips
	vibDepthFull int     // reload for the depth counter
	vibStep      int     // signed per-step increment
}

var mtick int   // global per-frame counter decrement ($DC0A), set by $80
var mgTempo int // global note-length multiplier ($DC08); used by channels with no own $80

// decode advances the channel past zero-time commands until it sets a new note/rest duration.
func (c *channel) decode() {
	for c.active && !c.end {
		b := int(mrb(c.pos))
		switch {
		case b < 0x70: // NOTE
			c.octaveNote(b)
			c.pos++
			d := int(mrb(c.pos))
			c.pos++
			c.startNote(d, false)
			return
		case b == 0x7F: // REST
			c.pos++
			d := int(mrb(c.pos))
			c.pos++
			c.startNote(d, true)
			return
		case b < 0x7F: // voice $70-$7E: an instrument from $43CE (8 bytes)
			v := b & 0x0F
			c.pos++
			c.noiseMd = int(mrb(0x43CE + v*8))
			for i := 0; i < 6; i++ {
				c.env[i] = mrb(0x43CE + v*8 + 1 + i)
			}
			d := int(mrb(c.pos))
			c.pos++
			c.startNote(d, false)
			return
		default: // command >= $80
			if !c.command(b) {
				return // $FE end
			}
		}
	}
}

func (c *channel) octaveNote(b int) {
	c.octave = b >> 4
	c.baseFreq = mfreq(b & 0x0F)
}

func (c *channel) startNote(d int, rest bool) {
	if d == 0 {
		d = c.defDur
	}
	t := c.tempo
	if t == 0 { // no per-channel $80: use the global tempo set by the control channel
		t = mgTempo
	}
	c.dur += d * t
	c.silent = rest
	if !c.tie {
		c.envLevel = 0
		c.envPhase = 0
		c.vibDelay = int(c.vib[0])
		c.vibSpeedCtr = int(c.vib[1])
		c.vibSpeed = int(c.vib[1])
		c.vibDepthCtr = int(c.vib[2]) >> 1
		c.vibDepthFull = int(c.vib[2])
		c.vibStep = int(int16(uint16(c.vib[3]) | uint16(c.vib[4])<<8))
		c.vibVal = 0
	}
	c.tie = false
}

// command returns false on $FE (end).
func (c *channel) command(b int) bool {
	c.pos++
	switch b {
	case 0x80: // tempo: word (mult) + tick + 1
		c.tempo = int(mrb(c.pos)) | int(mrb(c.pos+1))<<8
		mgTempo = c.tempo
		mtick = int(mrb(c.pos + 2))
		c.pos += 4
	case 0x81:
		c.vol = int(mrb(c.pos))
		c.pos++
	case 0x82:
		for i := 0; i < 6; i++ {
			c.env[i] = mrb(c.pos + i)
		}
		c.pos += 6
	case 0x83:
		for i := 0; i < 5; i++ {
			c.vib[i] = mrb(c.pos + i)
		}
		c.pos += 5
	case 0x84:
		c.detune = int(int16(uint16(mrb(c.pos)) | uint16(mrb(c.pos+1))<<8))
		c.pos += 2
	case 0x85:
		c.pos++ // skip 1
	case 0x86: // begin a repeat block: push a fresh counter slot
		c.stack = append(c.stack, 0)
	case 0x87: // repeat: $87 <count> <addrLo> <addrHi>
		count := int(mrb(c.pos))
		target := (int(mrb(c.pos+1)) | int(mrb(c.pos+2))<<8) + c.base
		top := len(c.stack) - 1
		if top < 0 {
			c.pos += 3
			break
		}
		if c.stack[top] == 0 {
			c.stack[top] = count - 1
		} else {
			c.stack[top]--
		}
		if c.stack[top] <= 0 {
			c.stack = c.stack[:top]
			c.pos += 3
		} else {
			c.pos = target
		}
	case 0x88:
		c.loop = c.pos // set loop point
	case 0x89:
		c.noiseMd = int(mrb(c.pos))
		c.pos++
	case 0x8A:
		c.defDur = int(mrb(c.pos))
		c.pos++
	case 0x8B:
		if c.vol < 15 {
			c.vol++
		}
	case 0x8C:
		if c.vol > 0 {
			c.vol--
		}
	case 0x8D:
		c.tie = true
	case 0xFE:
		c.end = true
		c.silent = true
		return false
	case 0xFF:
		if c.loop != 0 {
			c.pos = c.loop
			c.loops++
		} else {
			c.end = true
			return false
		}
	default:
		c.end = true
		return false
	}
	return true
}

// envStep advances the ADSR envelope one frame ($43E3 / $4545..$4597).
func (c *channel) envStep() {
	switch c.envPhase {
	case 0: // attack
		c.envLevel += int(c.env[0])
		if c.envLevel >= 0xFF {
			c.envLevel = 0xFF
			c.envPhase = 1
		}
	case 1: // decay to sustain
		c.envLevel -= int(c.env[1])
		if c.envLevel <= int(c.env[2]) {
			c.envLevel = int(c.env[2])
			c.envPhase = 2
		}
	case 2: // decay to sustain2
		c.envLevel -= int(c.env[3])
		if c.envLevel <= int(c.env[4]) {
			c.envLevel = int(c.env[4])
			c.envPhase = 3
		}
	case 3: // release/hold
		c.envLevel -= int(c.env[5])
		if c.envLevel < 0 {
			c.envLevel = 0
		}
	}
}

// vibStepFrame advances the triangle vibrato one frame ($4412-$4459).
func (c *channel) vibStepFrame() int {
	if c.vibDelay > 0 {
		c.vibDelay--
		return c.vibVal
	}
	c.vibSpeedCtr--
	if c.vibSpeedCtr != 0 {
		return c.vibVal
	}
	c.vibSpeedCtr = c.vibSpeed
	c.vibDepthCtr--
	if c.vibDepthCtr == 0 {
		c.vibDepthCtr = c.vibDepthFull
		c.vibStep = -c.vibStep
		return c.vibVal
	}
	c.vibVal += c.vibStep
	return c.vibVal
}

// tickFrame advances the channel one frame; returns (period, amplitude, isNoise).
func (c *channel) tickFrame() (int, float64, bool) {
	if !c.active || c.end {
		return 0, 0, c.noise
	}
	c.dur -= mtick
	if c.dur <= 0 && c.active && !c.end { // one note per frame, not a loop
		c.decode()
	}
	c.envStep()
	vib := c.vibStepFrame()
	per := (c.baseFreq + c.detune + vib) >> c.octave
	if per < 1 {
		per = 1
	}
	vol := 0
	if !c.silent && c.vol > 0 {
		vol = (c.vol * c.envLevel) >> 8
		if vol > 15 {
			vol = 15
		}
	}
	if c.noise { // the noise channel reports its mode byte instead of a tone period
		per = c.noiseMd
	}
	return per, psgAtten[15-vol] * boolf(vol > 0), c.noise
}

func boolf(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

type song struct {
	name string
	id   int // index into the song-pointer table $4716 (the music id, $D2F7)
}

// zoneSongs are the 7 level themes the music stage renders: id = descriptor+36 (zones map to
// 0-5, the special stage to 16). Their names/display names match musicTrack/musicName in main.go.
var zoneSongs = []song{
	{"greenhills", 0}, {"bridge", 1}, {"jungle", 2}, {"labyrinth", 3},
	{"scrapbrain", 4}, {"skybase", 5}, {"special", 16},
}

// songBase resolves a music id to its channel-data base via the song-pointer table $4716.
func songBase(id int) int { return mw(0x4716 + id*2) }

func newChannels(base int) []*channel {
	chs := make([]*channel, 4)
	for i := 0; i < 4; i++ {
		off := mw(base + i*2)
		c := &channel{pos: base + off, base: base, active: true, tempo: 0, defDur: 1, vol: 15, noise: i == 3}
		c.dur = 0
		c.decode() // prime the first event
		chs[i] = c
	}
	return chs
}

// exportMusic renders the 7 zone themes to music/<name>.mp3 (WAV via ffmpeg, then removed)
// and returns the manifest music index (display name + file). NO oracle. Deterministic order.
func exportMusic(romBytes []byte, outdir string) []MusicEntry {
	mrom = romBytes
	musicDir := filepath.Join(outdir, "music")
	chk(os.MkdirAll(musicDir, 0o755))
	var entries []MusicEntry
	for i, s := range zoneSongs {
		pcm, loopFrames := render(songBase(s.id))
		wav := filepath.Join(musicDir, s.name+".wav")
		writeWAV(wav, pcm)
		mp3 := filepath.Join(musicDir, s.name+".mp3")
		c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "64k", "-ac", "1", mp3)
		chk(c.Run())
		os.Remove(wav)
		fi, _ := os.Stat(mp3)
		entries = append(entries, MusicEntry{Name: musicName(s.name), File: "music/" + s.name + ".mp3"})
		fmt.Fprintf(os.Stderr, "[music] %d/%d  %-12s id %2d base $%04X -> %s (%d KB, loop %.1fs)\n",
			i+1, len(zoneSongs), s.name, s.id, songBase(s.id), s.name+".mp3", fi.Size()/1024, float64(loopFrames)/fps)
	}
	fmt.Fprintf(os.Stderr, "[music] done: %d tracks\n", len(entries))
	return entries
}

func render(base int) ([]int16, int) {
	mtick = 1
	mgTempo = 1
	chs := newChannels(base)
	const maxF = 90 * fps
	pos := make([][4]int, 0, maxF)
	pcm := make([]int16, 0, maxF*perFrm)
	var phase [3]float64
	var lfsr uint16 = 0x8000
	var nAcc, nOut float64
	var firstFF [4]int
	for f := 0; f < maxF; f++ {
		var per [4]int
		var amp [4]float64
		for i, c := range chs {
			per[i], amp[i], _ = c.tickFrame()
		}
		var pp [4]int
		for i, c := range chs {
			pp[i] = c.pos
			if c.loops >= 1 && firstFF[i] == 0 {
				firstFF[i] = f
			}
		}
		pos = append(pos, pp)
		for s := 0; s < perFrm; s++ {
			out := 0.0
			for i := 0; i < 3; i++ { // tone channels
				p := float64(per[i])
				if p < 1 {
					p = 1
				}
				phase[i] += PSGClock / (32 * p) / sr
				if phase[i] >= 1 {
					phase[i] -= math.Floor(phase[i])
				}
				if phase[i] < 0.5 {
					out += amp[i]
				} else {
					out -= amp[i]
				}
			}
			// noise channel (3): the SN76489 noise. per[3] carries the mode byte.
			mode := per[3]
			var np float64
			switch mode & 3 {
			case 0:
				np = 16
			case 1:
				np = 32
			case 2:
				np = 64
			default:
				np = float64(per[2])
				if np < 1 {
					np = 16
				}
			}
			nAcc += PSGClock / (32 * np) / sr
			for nAcc >= 1 {
				nAcc--
				var fb uint16
				if mode&4 != 0 { // white noise (tapped bits 0 and 3, SMS/GG)
					fb = (lfsr ^ (lfsr >> 3)) & 1
				} else { // periodic noise
					fb = lfsr & 1
				}
				bit := lfsr & 1
				lfsr = (lfsr >> 1) | (fb << 15)
				if lfsr == 0 {
					lfsr = 0x8000
				}
				nOut = float64(int(bit)*2 - 1)
			}
			out += nOut * amp[3] * 0.6
			pcm = append(pcm, int16(math.Max(-1, math.Min(1, out/4))*30000))
		}
	}
	loopFrames := detectPeriod(pos)
	loop := pcm[len(pcm)-loopFrames*perFrm:]
	// Some songs (the special stage) play once then loop on silence; if the detected loop
	// region is silent, loop the intro instead (up to the last channel's first $FF).
	if silent(loop) {
		introEnd := 0
		for _, ff := range firstFF {
			if ff > introEnd {
				introEnd = ff
			}
		}
		if introEnd > 0 {
			loopFrames = introEnd
			loop = pcm[:introEnd*perFrm]
		}
	}
	pcm = loop
	// short cross-fade at the loop seam (square phase isn't aligned at the cut)
	if k := sr / 80; len(pcm) > 2*k {
		for i := 0; i < k; i++ {
			a := float64(i) / float64(k)
			pcm[len(pcm)-k+i] = int16(float64(pcm[len(pcm)-k+i]) * (1 - a))
			pcm[i] = int16(float64(pcm[i]) * a)
		}
	}
	return pcm, loopFrames
}

// silent reports whether a clip is essentially silent (a song that loops on a rest).
func silent(pcm []int16) bool {
	var peak int16
	for _, s := range pcm {
		if s > peak {
			peak = s
		} else if -s > peak {
			peak = -s
		}
	}
	return peak < 200
}

// detectPeriod finds the song's loop length (frames) from the simulated per-frame channel
// pointers: the musical loop is the longest channel period.
func detectPeriod(pos [][4]int) int {
	n := len(pos)
	best := 0
	for ch := 0; ch < 4; ch++ {
		for p := 60; p <= 4500 && p < n; p++ {
			run := 0
			for i := n - 1; i-p >= 0 && pos[i][ch] == pos[i-p][ch]; i-- {
				run++
			}
			if run >= 900 { // 15 s of exact repeat at the smallest lag = this channel's loop
				if p > best {
					best = p
				}
				break
			}
		}
	}
	if best == 0 {
		best = 30 * fps
	}
	return best
}

// PSGClock is the SN76489 input clock.
const PSGClock = 3579545.0

func writeWAV(path string, pcm []int16) {
	f, _ := os.Create(path)
	defer f.Close()
	dl := len(pcm) * 2
	wr := func(v interface{}) { binary.Write(f, binary.LittleEndian, v) }
	f.Write([]byte("RIFF"))
	wr(uint32(36 + dl))
	f.Write([]byte("WAVEfmt "))
	wr(uint32(16))
	wr(uint16(1))
	wr(uint16(1))
	wr(uint32(sr))
	wr(uint32(sr * 2))
	wr(uint16(2))
	wr(uint16(16))
	f.Write([]byte("data"))
	wr(uint32(dl))
	wr(pcm)
}
