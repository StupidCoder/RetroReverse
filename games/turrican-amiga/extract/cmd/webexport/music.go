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
	"os"
	"os/exec"
	"path/filepath"

	"retroreverse.com/games/turrican-amiga/extract/decrunch"
	"retroreverse.com/games/turrican-amiga/extract/scene"
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

// exportMusic renders every distinct audible sub-song to music/*.mp3 and returns the manifest
// music index. NO oracle beyond the driver interpreter. Deterministic order.
func exportMusic(adf []byte, game *scene.Game, outDir string) ([]musicEntry, error) {
	overlay, err := decrunch.DecrunchBlock(adf[soundOff : soundOff+soundLen])
	if err != nil {
		return nil, err
	}
	musicDir := filepath.Join(outDir, "music")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		return nil, err
	}

	const sr = 44100
	const secs = 90
	var entries []musicEntry
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
			stem := fmt.Sprintf("mus_%X_%02X", m.addr, s)
			wav := filepath.Join(musicDir, stem+".wav")
			if err := writeWAV(wav, pcm, sr); err != nil {
				return nil, err
			}
			mp3 := filepath.Join(musicDir, stem+".mp3")
			c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
				"-c:a", "libmp3lame", "-b:a", "96k", "-ac", "2", mp3)
			if err := c.Run(); err != nil {
				return nil, fmt.Errorf("ffmpeg %s: %w", stem, err)
			}
			os.Remove(wav)
			fi, _ := os.Stat(mp3)
			entries = append(entries, musicEntry{
				Name: fmt.Sprintf("%s (track $%02X)", m.label, s),
				File: "music/" + stem + ".mp3",
			})
			fmt.Fprintf(os.Stderr, "[music] %-8s song %2d (track $%02X-$%02X) %.1fs -> %s (%d KB)\n",
				m.label, i, s, be16(0x140+i*2), float64(len(pcm)/2)/sr, stem+".mp3", fi.Size()/1024)
		}
	}
	fmt.Fprintf(os.Stderr, "[music] done: %d tracks\n", len(entries))
	return entries, nil
}

// writeWAV writes interleaved stereo float32 [-1,1] as 16-bit PCM WAV.
func writeWAV(path string, pcm []float32, sr int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	n := len(pcm)
	dataLen := n * 2
	hdr := make([]byte, 44)
	copy(hdr[0:], "RIFF")
	binary.LittleEndian.PutUint32(hdr[4:], uint32(36+dataLen))
	copy(hdr[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(hdr[16:], 16)
	binary.LittleEndian.PutUint16(hdr[20:], 1)              // PCM
	binary.LittleEndian.PutUint16(hdr[22:], 2)              // stereo
	binary.LittleEndian.PutUint32(hdr[24:], uint32(sr))     // rate
	binary.LittleEndian.PutUint32(hdr[28:], uint32(sr*2*2)) // byte rate
	binary.LittleEndian.PutUint16(hdr[32:], 4)              // block align
	binary.LittleEndian.PutUint16(hdr[34:], 16)             // bits
	copy(hdr[36:], "data")
	binary.LittleEndian.PutUint32(hdr[40:], uint32(dataLen))
	if _, err := f.Write(hdr); err != nil {
		return err
	}
	buf := make([]byte, dataLen)
	for i, s := range pcm {
		v := int16(s * 32000)
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}
	_, err = f.Write(buf)
	return err
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
