// Package audio is the one MP3 encoding helper every exporter shares. Game
// audio engines render PCM in Go; this package wraps it in WAV and shells out
// to ffmpeg (libmp3lame) with uniform settings, so every game's music and
// sound effects are encoded identically.
package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// PCM16 is interleaved signed 16-bit PCM.
type PCM16 struct {
	Rate     int // samples per second
	Channels int // 1 or 2
	Samples  []int16
}

// Duration in seconds.
func (p PCM16) Duration() float64 {
	if p.Rate <= 0 || p.Channels <= 0 {
		return 0
	}
	return float64(len(p.Samples)) / float64(p.Rate*p.Channels)
}

// WAV wraps the PCM in a RIFF/WAVE container.
func (p PCM16) WAV() []byte {
	data := make([]byte, len(p.Samples)*2)
	for i, s := range p.Samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}
	var buf bytes.Buffer
	w32 := func(v uint32) { _ = binary.Write(&buf, binary.LittleEndian, v) }
	w16 := func(v uint16) { _ = binary.Write(&buf, binary.LittleEndian, v) }
	buf.WriteString("RIFF")
	w32(uint32(36 + len(data)))
	buf.WriteString("WAVEfmt ")
	w32(16)
	w16(1) // PCM
	w16(uint16(p.Channels))
	w32(uint32(p.Rate))
	w32(uint32(p.Rate * p.Channels * 2))
	w16(uint16(p.Channels * 2))
	w16(16)
	buf.WriteString("data")
	w32(uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

// mp3Args is the uniform encode: VBR quality 2 (~190 kbit/s), plenty for
// retro sources while keeping files small.
var mp3Args = []string{"-codec:a", "libmp3lame", "-q:a", "2"}

// EncodeMP3 encodes PCM straight to outPath (directories are created).
func EncodeMP3(pcm PCM16, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	args := append([]string{"-hide_banner", "-loglevel", "error", "-y", "-f", "wav", "-i", "-"}, mp3Args...)
	cmd := exec.Command("ffmpeg", append(args, outPath)...)
	cmd.Stdin = bytes.NewReader(pcm.WAV())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// EncodeMP3File encodes an existing audio file (e.g. a rendered WAV).
func EncodeMP3File(srcPath, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	args := append([]string{"-hide_banner", "-loglevel", "error", "-y", "-i", srcPath}, mp3Args...)
	cmd := exec.Command("ffmpeg", append(args, outPath)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// Duration probes a media file's duration in seconds via ffprobe.
func Duration(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}
