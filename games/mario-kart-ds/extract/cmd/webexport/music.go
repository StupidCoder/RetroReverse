// music stage: render Mario Kart DS's sequenced music from sound_data.sdat to
// MP3, as cmd/musicrender (SSEQ bytecode → SBNK instruments → SWAR samples via
// tools/nds/sdat, then ffmpeg to MP3).
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/sdat"
)

const rate = 32768 // the DS mixer's output rate

func runMusic(rom *nds.ROM, out string) []manifestMusic {
	data := rom.FileByPath("data/Sound/sound_data.sdat")
	if data == nil {
		die(fmt.Errorf("data/Sound/sound_data.sdat not found in ROM"))
	}
	s, err := sdat.Parse(data)
	if err != nil {
		die(err)
	}
	musicDir := filepath.Join(out, "music")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		die(err)
	}

	total := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID >= 0 {
			total++
		}
	}
	fmt.Fprintf(os.Stderr, "[music] SDAT: %d sequences (%d renderable), %d banks, %d wave archives\n",
		len(s.Seqs), total, len(s.Banks), len(s.Wavearcs))

	var tracks []manifestMusic
	n := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID < 0 {
			continue
		}
		n++
		stem := fmt.Sprintf("seq_%02d", i)
		L, R, err := s.Render(i, rate, 2, 180)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s: %v\n", n, total, stem, err)
			continue
		}
		if len(L) < rate { // sub-second jingles/stingers cut short
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s (skipped, too short)\n", n, total, stem)
			continue
		}
		fadeOut(L, R)
		wav := filepath.Join(musicDir, stem+".wav")
		if err := writeWAV(wav, L, R); err != nil {
			die(err)
		}
		mp3 := filepath.Join(musicDir, stem+".mp3")
		c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "96k", mp3)
		if e := c.Run(); e != nil {
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s: ffmpeg failed: %v (WAV kept)\n", n, total, stem, e)
			continue
		}
		os.Remove(wav)
		secs := len(L) / rate
		tracks = append(tracks, manifestMusic{
			Name: fmt.Sprintf("Sequence %02d · %d:%02d", i, secs/60, secs%60),
			File: "music/" + stem + ".mp3",
		})
		fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s.mp3 (%5.1fs)\n", n, total, stem, float64(len(L))/rate)
	}
	sort.Slice(tracks, func(i, j int) bool { return tracks[i].File < tracks[j].File })
	return tracks
}

// fadeOut applies a 3-second fade at the end (the render stops after the loop
// count, mid-music).
func fadeOut(L, R []float64) {
	n := 3 * rate
	if n > len(L) {
		n = len(L)
	}
	for i := 0; i < n; i++ {
		g := float64(n-i) / float64(n)
		L[len(L)-n+i] *= g
		R[len(R)-n+i] *= g
	}
}

func writeWAV(path string, L, R []float64) error {
	n := len(L)
	buf := make([]byte, 44+n*4)
	copy(buf, "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+n*4))
	copy(buf[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:], 2) // stereo
	binary.LittleEndian.PutUint32(buf[24:], rate)
	binary.LittleEndian.PutUint32(buf[28:], rate*4)
	binary.LittleEndian.PutUint16(buf[32:], 4)
	binary.LittleEndian.PutUint16(buf[34:], 16)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(n*4))
	clip := func(v float64) int16 {
		if v > 1 {
			v = 1
		}
		if v < -1 {
			v = -1
		}
		return int16(v * 32767)
	}
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(buf[44+i*4:], uint16(clip(L[i])))
		binary.LittleEndian.PutUint16(buf[46+i*4:], uint16(clip(R[i])))
	}
	return os.WriteFile(path, buf, 0o644)
}
