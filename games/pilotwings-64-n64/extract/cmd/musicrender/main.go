// musicrender renders Pilotwings 64's 31 songs through the tools/platform/n64/audio
// synth to WAV, then MP3 via ffmpeg — the N64 analogue of the SM64DS musicrender.
// It reads the music instrument bank the game DMAs from cart 0x62D460, that bank's
// own VADPCM sample table at 0x6314D0, and the "S1" sequence bank at 0x618B70.
//
//	musicrender -image ROM [-o DIR] [-song N] [-rate HZ] [-loops N] [-max SEC] [-raw]
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"retroreverse.com/tools/platform/n64"
	"retroreverse.com/tools/platform/n64/audio"
)

const (
	seqBankOff  = 0x618B70
	seqBankEnd  = 0x62D460
	musicBank   = 0x62D460
	musicTblOff = 0x6314D0
)

func main() {
	image := flag.String("image", "", "cartridge image")
	outDir := flag.String("o", "../rendered/music", "output directory")
	songN := flag.Int("song", -1, "render one song (-1 = all)")
	rate := flag.Float64("rate", 22050, "output sample rate")
	loops := flag.Int("loops", 2, "stop looping songs after N repeats")
	maxSec := flag.Float64("max", 150, "hard length cap per song (seconds)")
	raw := flag.Bool("raw", false, "also write raw big-endian s16 stereo (for A/B vs the oracle capture)")
	flag.Parse()
	if *image == "" {
		die(fmt.Errorf("-image is required"))
	}
	rom, err := n64.Load(*image)
	if err != nil {
		die(err)
	}
	bank, err := audio.ParseBankFile(rom.Data[musicBank:])
	if err != nil {
		die(fmt.Errorf("music bank: %w", err))
	}
	sb, err := audio.ParseSeqBank(rom.Data[seqBankOff:seqBankEnd])
	if err != nil {
		die(fmt.Errorf("seq bank: %w", err))
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	tbl := rom.Data[musicTblOff:]
	player := audio.NewPlayer(bank.Banks[0], tbl, *rate)

	type track struct {
		Name string `json:"name"`
		File string `json:"file"`
	}
	var tracks []track
	from, to := 0, len(sb.Songs)
	if *songN >= 0 {
		from, to = *songN, *songN+1
	}
	for i := from; i < to; i++ {
		L, R := player.Render(sb.Songs[i], *loops, *maxSec)
		dur := float64(len(L)) / *rate
		stem := fmt.Sprintf("song_%02d", i)
		wav := filepath.Join(*outDir, stem+".wav")
		if err := writeWAV(wav, L, R, int(*rate)); err != nil {
			die(err)
		}
		if *raw {
			writeRaw(filepath.Join(*outDir, stem+".raw"), L, R)
		}
		mp3 := filepath.Join(*outDir, stem+".mp3")
		if err := ffmpeg(wav, mp3); err != nil {
			fmt.Fprintf(os.Stderr, "musicrender: ffmpeg %s: %v\n", stem, err)
		} else {
			os.Remove(wav)
		}
		fmt.Printf("song %2d: %.1fs, %d samples\n", i, dur, len(L))
		tracks = append(tracks, track{Name: fmt.Sprintf("Song %02d", i), File: "music/" + stem + ".mp3"})
	}
	if b, err := json.MarshalIndent(tracks, "", "  "); err == nil {
		os.WriteFile(filepath.Join(*outDir, "tracks.json"), b, 0o644)
	}
}

func writeWAV(path string, L, R []float64, rate int) error {
	n := len(L)
	buf := make([]byte, 44+n*4)
	copy(buf, "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+n*4))
	copy(buf[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], 1)
	binary.LittleEndian.PutUint16(buf[22:], 2)
	binary.LittleEndian.PutUint32(buf[24:], uint32(rate))
	binary.LittleEndian.PutUint32(buf[28:], uint32(rate*4))
	binary.LittleEndian.PutUint16(buf[32:], 4)
	binary.LittleEndian.PutUint16(buf[34:], 16)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(n*4))
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(buf[44+i*4:], uint16(clip(L[i])))
		binary.LittleEndian.PutUint16(buf[44+i*4+2:], uint16(clip(R[i])))
	}
	return os.WriteFile(path, buf, 0o644)
}

// writeRaw writes big-endian s16 stereo, matching the oracle's AI capture format.
func writeRaw(path string, L, R []float64) error {
	buf := make([]byte, len(L)*4)
	for i := range L {
		binary.BigEndian.PutUint16(buf[i*4:], uint16(clip(L[i])))
		binary.BigEndian.PutUint16(buf[i*4+2:], uint16(clip(R[i])))
	}
	return os.WriteFile(path, buf, 0o644)
}

func clip(v float64) int16 {
	s := v * 32767
	if s > 32767 {
		return 32767
	}
	if s < -32768 {
		return -32768
	}
	return int16(s)
}

func ffmpeg(wav, mp3 string) error {
	return exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
		"-c:a", "libmp3lame", "-b:a", "96k", mp3).Run()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "musicrender:", err)
	os.Exit(1)
}
