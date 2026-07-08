// music.go is the music stage of webexport: it synthesises Super Mario Land's music ENTIRELY
// FROM THE ROM song data — a Go port of the bank-3 sound engine (Super_Mario_Land.md Part VI),
// NO emulator in the render loop (folded from cmd/musicrom). Each track is the level themes
// (named by the levels that use them) plus the bonus jingle; -only music runs standalone.
//
// Format (all bank-3 addresses):
//   - The song table $673C maps a music id to a song header: master byte, a 16-bit pointer to
//     the DURATION TABLE, then four 16-bit channel-header pointers (square1, square2, wave, noise).
//   - A channel header is an ORDER LIST of 16-bit pattern pointers ending in an $FF entry followed
//     by a 2-byte LOOP TARGET; patterns before the target are a one-shot intro.
//   - A pattern is a byte stream: $9D a b c set voice; $A0-$AF set the note duration to
//     durtable[low nibble]; $00 ends the pattern; note $01 = note-off/rest; any other byte N is a
//     note (pitch = freqtable[$6F70 + N], two bytes/semitone; the noise channel indexes $7002).
//   - The engine ticks at 64 Hz. The decoded notes drive our DMG APU (tools/gameboy/apu.go); each
//     track is intro + two loops with a fade tail, encoded to MP3 by ffmpeg (libmp3lame).
//
// This file uses its own bank-3 globals (mrom + b3/rb/rw) to avoid clashing with the level/sprite
// stages' ROM helpers.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"retroreverse.com/tools/platform/gameboy"
)

const cycPerTick = 65536 // 4194304 Hz / 64 Hz tick

var mrom []byte

func b3(a int) int  { return 3*0x4000 + (a - 0x4000) }
func rb(a int) byte { return mrom[b3(a)] }
func rw(a int) int  { return int(mrom[b3(a)]) | int(mrom[b3(a)+1])<<8 }

type instrument struct{ env, duty, x byte }

// chanEvent is one note on a channel timeline.
type chanEvent struct {
	tick, dur int
	freq      int // GB 11-bit frequency (square/wave) or noise poly byte
	tie, rest bool
	inst      instrument
}

type channel struct {
	events    []chanEvent // the whole channel: intro then one loop, ticks from 0
	loopStart int         // tick where the looping body begins (intro plays once)
	loopLen   int         // length of the loop body in ticks
}

// decodeChannel walks a channel's order list and patterns into note events, stopping at the $FF
// loop marker (one full loop). isNoise selects the noise vs pitch interpretation.
func decodeChannel(hdr, durTbl int, isNoise bool) channel {
	end := hdr
	target := hdr
	for guard := 0; guard < 512; guard++ {
		if rb(end) == 0xFF {
			target = rw(end + 2)
			break
		}
		if rb(end) == 0x00 {
			break
		}
		end += 2
	}

	var all []chanEvent
	var inst instrument
	dur, tick, order, prev, loopTick := 1, 0, hdr, 0, 0
	for order < end {
		if order == target {
			loopTick = tick
		}
		pat := rw(order)
		order += 2
		for g2 := 0; g2 < 4096; g2++ {
			c := rb(pat)
			if c == 0x00 {
				break
			}
			switch {
			case c == 0x9D:
				inst = instrument{rb(pat + 1), rb(pat + 2), rb(pat + 3)}
				pat += 4
			case c >= 0xA0 && c <= 0xAF:
				dur = int(rb(durTbl + int(c&0x0F)))
				pat++
			default:
				ev := chanEvent{tick: tick, dur: dur, inst: inst}
				switch {
				case c == 0x01: // note-off: retrigger with a DAC-off envelope to silence the channel
					ev.freq = prev
					ev.rest = true
				case isNoise:
					ev.freq = int(c)
					prev = ev.freq
				default:
					ev.freq = rw(0x6F70 + int(c))
					prev = ev.freq
				}
				all = append(all, ev)
				tick += dur
				pat++
			}
		}
	}
	return channel{events: all, loopStart: loopTick, loopLen: tick - loopTick}
}

// song decodes a music id into its four channels. A channel whose header pointer is not in bank 3
// ($6000-$7FFF) is unused (e.g. the bonus theme has no wave channel) and stays empty.
func song(id int) [4]channel {
	hdr := rw(0x673C + (id-1)*2)
	durTbl := rw(hdr + 1)
	var chs [4]channel
	for i := 0; i < 4; i++ {
		if p := rw(hdr + 3 + i*2); p >= 0x6000 && p < 0x8000 {
			chs[i] = decodeChannel(p, durTbl, i == 3)
		}
	}
	return chs
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// regBase is the first APU register of each channel (NR_1).
var regBase = [4]uint16{0xFF11, 0xFF16, 0xFF1B, 0xFF20}

// render plays each channel's intro once and then loops its body, for the whole song's intro +
// `loops` loop iterations, and returns peak-normalised PCM with the tail faded out.
func render(chs [4]channel, loops int) []int16 {
	lcm := func(chans []channel) int {
		l := 1
		for _, c := range chans {
			if c.loopLen > 0 {
				l = l / gcd(l, c.loopLen) * c.loopLen
			}
		}
		return l
	}
	songLoop := lcm(chs[:])
	if songLoop > 5000 {
		songLoop = lcm(chs[0:2])
	}
	maxIntro := 0
	for _, c := range chs {
		if c.loopStart > maxIntro {
			maxIntro = c.loopStart
		}
	}
	total := maxIntro + loops*songLoop // ticks to render

	var ev []gameboy.RegWrite
	at := func(tick int, reg uint16, v byte) {
		ev = append(ev, gameboy.RegWrite{Cycle: int64(tick) * cycPerTick, Reg: reg, Val: v})
	}
	at(0, 0xFF26, 0x80) // power on, full panning, master volume, wave DAC on
	at(0, 0xFF25, 0xFF)
	at(0, 0xFF24, 0x77)
	at(0, 0xFF1A, 0x80)

	solo := os.Getenv("MUSSOLO")
	for ci, c := range chs {
		if len(c.events) == 0 || (solo != "" && solo != fmt.Sprint(ci)) {
			continue
		}
		base := regBase[ci]
		for rep := 0; ; rep++ {
			off := rep * c.loopLen
			done := true
			for _, e := range c.events {
				if rep > 0 && e.tick < c.loopStart {
					continue // intro only on the first pass
				}
				tk := e.tick + off
				if tk >= total {
					continue
				}
				done = false
				e := e
				e.tick = tk
				emitNote(&ev, at, ci, base, e)
			}
			if done || c.loopLen == 0 {
				break
			}
		}
	}
	sort.SliceStable(ev, func(i, j int) bool { return ev[i].Cycle < ev[j].Cycle })

	pcm := normalize(gameboy.NewAPU().Render(ev, int64(total)*cycPerTick))
	return fadeOut(pcm, 2.5)
}

// fadeOut applies a linear fade over the final `secs` seconds.
func fadeOut(pcm []int16, secs float64) []int16 {
	n := int(secs * gameboy.APURate)
	if n > len(pcm) {
		n = len(pcm)
	}
	for i := 0; i < n; i++ {
		g := float64(n-i) / float64(n)
		pcm[len(pcm)-n+i] = int16(float64(pcm[len(pcm)-n+i]) * g)
	}
	return pcm
}

// emitNote appends the APU register writes for one note event on channel ci.
func emitNote(ev *[]gameboy.RegWrite, at func(int, uint16, byte), ci int, base uint16, e chanEvent) {
	switch ci {
	case 0, 1: // square — the $9D bytes are (env, _, duty): NRx2=env, NRx1=duty/length
		envv := e.inst.env
		if e.rest {
			envv = 0x01 // DAC off -> silence (note-off)
		}
		at(e.tick, base, e.inst.x) // NRx1 duty/length (the 3rd $9D byte)
		at(e.tick, base+1, envv)   // NRx2 envelope (the 1st $9D byte)
		at(e.tick, base+2, byte(e.freq))
		at(e.tick, base+3, byte(e.freq>>8)|0x80) // trigger
	case 2: // wave — the instrument's first two bytes are a pointer to 16 bytes of wave RAM
		if e.rest {
			at(e.tick, 0xFF1C, 0x00) // NR32 volume 0 -> mute
			return
		}
		wp := int(e.inst.env) | int(e.inst.duty)<<8
		if wp >= 0x6000 && wp < 0x8000 {
			at(e.tick, 0xFF1A, 0x00) // DAC off while loading wave RAM
			for k := 0; k < 16; k++ {
				at(e.tick, 0xFF30+uint16(k), rb(wp+k))
			}
			at(e.tick, 0xFF1A, 0x80) // DAC on
		}
		at(e.tick, 0xFF1C, e.inst.x) // NR32 volume (3rd $9D byte)
		at(e.tick, 0xFF1D, byte(e.freq))
		at(e.tick, 0xFF1E, byte(e.freq>>8)|0x80)
	case 3: // noise — the note byte indexes a 5-byte entry in the $7002 table:
		// [NR42 env, (unused), NR41 length, NR43 poly, NR44]. NR44 $C0 = trigger + length-enable.
		p := 0x7002 + e.freq
		if e.rest {
			at(e.tick, 0xFF21, 0x01) // DAC off
			return
		}
		at(e.tick, 0xFF20, rb(p+2)) // NR41 length
		at(e.tick, 0xFF21, rb(p))   // NR42 envelope
		at(e.tick, 0xFF22, rb(p+3)) // NR43 polynomial counter
		at(e.tick, 0xFF23, rb(p+4)) // NR44 trigger + length-enable ($C0)
	}
}

func normalize(pcm []int16) []int16 {
	peak := 1
	for _, s := range pcm {
		if v := int(s); v > peak {
			peak = v
		} else if -v > peak {
			peak = -v
		}
	}
	g := 29500.0 / float64(peak)
	for i, s := range pcm {
		v := float64(s) * g
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		pcm[i] = int16(v)
	}
	return pcm
}

// musicTracks are the level themes (named by the levels that use them via the ROM's song-select)
// plus the bonus theme; the display names index the manifest music list.
var musicTracks = []struct {
	id   int
	name string // file stem
	disp string // manifest display name
}{
	{0x07, "level-1-1", "Level 1-1"}, // 1-1, 1-2, 3-1
	{0x03, "level-1-3", "Level 1-3"}, // 1-3, 3-2, 3-3
	{0x08, "level-2-1", "Level 2-1"}, // 2-1, 2-2 (Muda)
	{0x06, "level-4-1", "Level 4-1"}, // 4-1, 4-2 (Chai)
	{0x05, "level-2-3", "Level 2-3"}, // 2-3, 4-3 (boss/vehicle stages)
	{0x04, "bonus", "Bonus"},         // pipe bonus rooms
}

// exportMusic renders the level themes + bonus jingle to music/<stem>.mp3 (WAV via ffmpeg, then
// removed) and returns the manifest music index. NO oracle. Deterministic order.
func exportMusic(romBytes []byte, outdir string) []MusicEntry {
	mrom = romBytes
	musicDir := filepath.Join(outdir, "music")
	must(os.MkdirAll(musicDir, 0o755))
	var entries []MusicEntry
	for i, t := range musicTracks {
		pcm := render(song(t.id), 2) // intro + 2 loops
		wav := filepath.Join(musicDir, t.name+".wav")
		writeWAV(wav, pcm)
		mp3 := filepath.Join(musicDir, t.name+".mp3")
		c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "96k", "-ac", "1", mp3)
		must(c.Run())
		os.Remove(wav)
		fi, _ := os.Stat(mp3)
		entries = append(entries, MusicEntry{Name: t.disp, File: "music/" + t.name + ".mp3"})
		fmt.Fprintf(os.Stderr, "[music] %d/%d  %-12s id $%02X -> %s (%d KB, %.1fs)\n",
			i+1, len(musicTracks), t.name, t.id, t.name+".mp3", fi.Size()/1024, float64(len(pcm))/gameboy.APURate)
	}
	fmt.Fprintf(os.Stderr, "[music] done: %d tracks\n", len(entries))
	return entries
}

func writeWAV(path string, pcm []int16) {
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	dataLen := len(pcm) * 2
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataLen))
	f.Write([]byte("WAVEfmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint32(gameboy.APURate))
	binary.Write(f, binary.LittleEndian, uint32(gameboy.APURate*2))
	binary.Write(f, binary.LittleEndian, uint16(2))
	binary.Write(f, binary.LittleEndian, uint16(16))
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataLen))
	binary.Write(f, binary.LittleEndian, pcm)
}
