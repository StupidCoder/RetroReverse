// musicrender renders Super Mario 64 DS's sequenced music from the cartridge's
// sound_data.sdat through the tools/nds/sdat sequencer+synth (SSEQ bytecode →
// SBNK instruments → SWAR samples) to WAV, then MP3 via ffmpeg — the same
// pipeline as Mario Kart DS's musicrender. Unlike MKDS, this SDAT ships WITH
// its SYMB name block, so every sequence carries the game's own symbol
// (NCS_BGM_TITLE, NCS_BGM_SHIRO, …), used for the file and track names.
//
//	musicrender [-sdat FILE] [-o DIR] [-seq N] [-loops N] [-max SEC]
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/nds/sdat"
)

const rate = 32768 // the DS mixer's output rate

func main() {
	sdatPath := flag.String("sdat", "../extracted/files/data/sound_data.sdat", "SDAT file")
	outDir := flag.String("o", "../rendered/music", "output directory")
	seqN := flag.Int("seq", -1, "render one sequence (-1 = all)")
	loops := flag.Int("loops", 2, "stop after N loops of the sequence")
	maxSec := flag.Float64("max", 180, "hard time limit per sequence (seconds)")
	flag.Parse()

	data, err := os.ReadFile(*sdatPath)
	if err != nil {
		die(err)
	}
	s, err := sdat.Parse(data)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	fmt.Printf("SDAT: %d sequences, %d banks, %d wave archives, %d files\n",
		len(s.Seqs), len(s.Banks), len(s.Wavearcs), s.NumFiles())

	from, to := 0, len(s.Seqs)
	if *seqN >= 0 {
		from, to = *seqN, *seqN+1
	}
	type trackEntry struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	var tracks []trackEntry
	for i := from; i < to; i++ {
		if s.Seqs[i].FileID < 0 {
			continue
		}
		stem := stemFor(i, s.Seqs[i].Name)
		L, R, err := s.Render(i, rate, *loops, *maxSec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  seq %02d %s: %v\n", i, stem, err)
			continue
		}
		if len(L) < rate { // sub-second sequences are jingles/stingers cut short
			fmt.Printf("  seq %02d %s: %5.1fs (skipped, too short)\n", i, stem, float64(len(L))/rate)
			continue
		}
		fadeOut(L, R)
		wav := filepath.Join(*outDir, stem+".wav")
		if err := writeWAV(wav, L, R); err != nil {
			die(err)
		}
		mp3 := filepath.Join(*outDir, stem+".mp3")
		c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "96k", mp3)
		if e := c.Run(); e != nil {
			fmt.Fprintf(os.Stderr, "  ffmpeg failed for seq %d: %v (WAV kept)\n", i, e)
			continue
		}
		os.Remove(wav)
		secs := len(L) / rate
		tracks = append(tracks, trackEntry{
			Name: fmt.Sprintf("%s · %d:%02d", labelFor(i, s.Seqs[i].Name), secs/60, secs%60),
			URL:  "public/sm64ds/music/" + stem + ".mp3",
		})
		fmt.Printf("  seq %02d %s: %5.1fs → %s\n", i, stem, float64(len(L))/rate, filepath.Base(mp3))
	}
	if *seqN < 0 {
		buf, _ := json.MarshalIndent(tracks, "", " ")
		if err := os.WriteFile(filepath.Join(*outDir, "tracks.json"), buf, 0o644); err != nil {
			die(err)
		}
		fmt.Printf("  tracks.json: %d tracks\n", len(tracks))
	}
}

// stemFor makes the output filename from the SYMB name (lowercased, NCS_BGM_
// prefix dropped), falling back to the index for nameless archives.
func stemFor(i int, name string) string {
	if name == "" {
		return fmt.Sprintf("seq_%02d", i)
	}
	return strings.ToLower(strings.TrimPrefix(name, "NCS_BGM_"))
}

// labelFor is the human-readable track name: the game's own symbol, with the
// NCS_BGM_ prefix dropped (we deliberately don't invent friendlier titles —
// the symbol is what the cartridge calls the tune).
func labelFor(i int, name string) string {
	if name == "" {
		return fmt.Sprintf("Sequence %02d", i)
	}
	return strings.TrimPrefix(name, "NCS_BGM_")
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

func die(err error) {
	fmt.Fprintln(os.Stderr, "musicrender:", err)
	os.Exit(1)
}
