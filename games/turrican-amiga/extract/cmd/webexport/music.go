// music.go is the music stage (folded ex-cmd/music): it renders every distinct, audible TFMX
// sub-song of every module — the $1BB00 sound overlay (title / menu / jingles) plus each of the
// five worlds' in-game themes — to music/*.mp3 by RUNNING the real 68000 sound driver in the
// interpreter (tfmx_driver.go / tfmx_m68k*.go) and mixing its Paula output. Turrican's score is
// Chris Hülsbeck's, in his own TFMX format. This is the slow stage (a CPU interpreter per song),
// but it is self-contained: it needs no levels/sprites stage.
//
// Each WAV is rendered, encoded to MP3 (ffmpeg / libmp3lame), and the WAV removed.
package main

import (
	"encoding/binary"
	"fmt"
	"math"

	"retroreverse.com/games/turrican-amiga/extract/decrunch"
	"retroreverse.com/games/turrican-amiga/extract/scene"
	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

const (
	soundOff  = 0x26000 // ADF offset of the packed sound overlay
	soundLen  = 0xC268
	soundBase = 0x1BB00 // its runtime load address
	mdatAddr  = 0x1CFF4 // api_init d0 (song/pattern/macro data)
	smplAddr  = 0x20E90 // api_init d1 (sample data)
)

// tfmxModule is one TFMX song/sample bank: a runtime address for naming, the mdat
// (song/pattern/macro/trackstep tables) and the sample bank it draws from.
type tfmxModule struct {
	label    string // human display label
	addr     int    // mdat runtime address (used in song names)
	smplAddr int    // sample bank runtime address
	mdat     []byte
	smpl     []byte
}

// modulesOf returns every TFMX module: the $1BB00 sound overlay plus each of the 5 worlds'
// in-game themes (select_scene wires a world's music as mdat = block+$10, smpl = block+$0C).
func modulesOf(game *scene.Game, overlay []byte) []tfmxModule {
	mods := []tfmxModule{{
		label:    "Overlay",
		addr:     mdatAddr,
		smplAddr: smplAddr,
		mdat:     overlay[mdatAddr-soundBase : smplAddr-soundBase],
		smpl:     overlay[smplAddr-soundBase:],
	}}
	for w := 0; w < scene.NumWorlds; w++ {
		blk := game.Block(w)
		mAddr := int(binary.BigEndian.Uint32(blk.Data[scene.BlockBase+0x10-blk.Base:]))
		sAddr := int(binary.BigEndian.Uint32(blk.Data[scene.BlockBase+0x0C-blk.Base:]))
		mOff, sOff := mAddr-blk.Base, sAddr-blk.Base
		if mOff < 0 || mOff >= len(blk.Data) || sOff < 0 || sOff >= len(blk.Data) {
			continue
		}
		mods = append(mods, tfmxModule{
			label:    fmt.Sprintf("World %d", w+1),
			addr:     mAddr,
			smplAddr: sAddr,
			mdat:     blk.Data[mOff:],
			smpl:     blk.Data[sOff:],
		})
	}
	return mods
}

// exportMusic renders every distinct audible sub-song of every module to
// music/*.mp3 and registers the music assets. NO oracle beyond the driver
// interpreter. Deterministic order.
func exportMusic(ctx *cli.Context, adf []byte, game *scene.Game) error {
	overlay, err := decrunch.DecrunchBlock(adf[soundOff : soundOff+soundLen])
	if err != nil {
		return err
	}

	const sr = 44100
	const secs = 90
	n := 0
	for _, m := range modulesOf(game, overlay) {
		be16 := func(o int) int { return int(binary.BigEndian.Uint16(m.mdat[o:])) }
		// The 32-slot song table is mostly padding: unused slots point at a single "stop" step
		// repeated 20+ times, and slot 31 is the $1FF terminator. Real sub-songs are the distinct
		// entries; a song's trackstep start identifies it. Count each (start,end) to spot the
		// filler, and keep one entry per distinct start (the widest trackstep range).
		cnt := map[[2]int]int{}
		for i := 0; i < 32; i++ {
			cnt[[2]int{be16(0x100 + i*2), be16(0x140 + i*2)}]++
		}
		best := map[int]int{} // trackstep start -> chosen slot index
		for i := 0; i < 32; i++ {
			s, e := be16(0x100+i*2), be16(0x140+i*2)
			if s > e || s >= 0x100 || e >= 0x100 { // out of range / $1FF terminator
				continue
			}
			if cnt[[2]int{s, e}] >= 8 { // a repeated stop-step, not a song
				continue
			}
			if cur, ok := best[s]; !ok || e-s > be16(0x140+cur*2)-s {
				best[s] = i
			}
		}
		for s := 0; s < 0x100; s++ {
			i, ok := best[s]
			if !ok {
				continue
			}
			// A single-trackstep song (start == end) loops inside its pattern, so the trackstep
			// position never jumps back — cap those at a short length.
			maxSecs := secs
			if s == be16(0x140+i*2) {
				maxSecs = 25
			}
			pcm, _ := renderDriver(overlay, m, i, sr, maxSecs, true)
			if rms(pcm) < 0.004 { // empty stub / silence
				continue
			}
			samples := make([]int16, len(pcm))
			for k, v := range pcm {
				if v > 1 {
					v = 1
				} else if v < -1 {
					v = -1
				}
				samples[k] = int16(v * 32767)
			}
			stem := fmt.Sprintf("mus_%X_%02X", m.addr, s)
			out, err := ctx.Builder.Path("music", stem+".mp3")
			if err != nil {
				return err
			}
			wave := audio.PCM16{Rate: sr, Channels: 2, Samples: samples}
			if err := audio.EncodeMP3(wave, out); err != nil {
				return err
			}
			n++
			ctx.Builder.AddMedia(schema.Asset{
				ID: stem, Category: schema.CategoryMusic,
				Name:     fmt.Sprintf("%s (track $%02X)", m.label, s),
				File:     "music/" + stem + ".mp3",
				Duration: wave.Duration(),
			})
			ctx.Progress("music", n, 0, fmt.Sprintf("%-8s song %2d (track $%02X-$%02X) %.1fs",
				m.label, i, s, be16(0x140+i*2), wave.Duration()))
		}
	}
	return nil
}

func rms(pcm []float32) float64 {
	var sum float64
	for _, s := range pcm {
		sum += float64(s) * float64(s)
	}
	if len(pcm) == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(len(pcm)))
}
