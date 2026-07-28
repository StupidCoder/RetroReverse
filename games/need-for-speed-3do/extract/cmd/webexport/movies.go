package main

// movies.go is the Retro-X `videos` stage: the boot intro FMV
// (Movies/eac.stream — the first movie the game arms at startup, confirmed by
// the bootoracle's MovieHLE open log) becomes videos/intro.mp4. The Cinepak
// video is decoded by our own Go decoder (tools/platform/threedo/cvid.go,
// byte-identical to the reference) and the SDX2 audio track by our own
// DPCM decoder (snds.go, byte-identical to FFmpeg's sdx2_dpcm across every
// movie on the disc); ffmpeg only re-encodes those already-decoded frames and
// samples to H.264 + AAC for the browser (moov up front, RETROX.md §9).

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/threedo"
)

const introStream = "Movies/eac.stream"

// exportIntroVideo decodes the intro stream and registers it as the game's
// `intro` video asset.
func exportIntroVideo(ctx *cli.Context, vol *threedo.Volume) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found on PATH (needed to encode the intro video): %w", err)
	}
	raw, err := vol.ReadFile(introStream)
	if err != nil {
		return err
	}
	mv, frames, err := decodeMovie(raw)
	if err != nil {
		return err
	}
	snd, err := threedo.DemuxSnds(raw)
	if err != nil {
		return err
	}

	mp4, err := ctx.Builder.Path("videos", "intro.mp4")
	if err != nil {
		return err
	}
	if err := encodeMP4(mp4, frames, mv, snd); err != nil {
		return err
	}

	duration := float64(len(frames)) / float64(mv.FPS)
	stats := map[string]any{
		"Source":    fmt.Sprintf("%s (%.1f MB)", introStream, float64(len(raw))/(1024*1024)),
		"Codec":     fmt.Sprintf("Cinepak (%q), software-decoded on the ARM60", mv.Codec),
		"Data rate": fmt.Sprintf("%.0f KB/s streamed off the CD", float64(len(raw))/duration/1024),
		"Native":    fmt.Sprintf("%d × %d px", mv.Width, mv.Height),
		"Colors":    "vector-quantised YCbCr 4:2:0, shown as 15-bit RGB555",
		"Length":    fmt.Sprintf("%d frames @ %d fps", len(frames), mv.FPS),
	}
	if snd != nil {
		stats["Audio"] = fmt.Sprintf("SDX2 DPCM, %d Hz, %d channels (2:1)", snd.SampleRate, snd.Channels)
	}
	ctx.Builder.AddMedia(schema.Asset{
		ID: "intro", Category: schema.CategoryVideo, Name: "Intro",
		File: "videos/intro.mp4",
		W:    mv.Width, H: mv.Height, FPS: float64(mv.FPS), Duration: duration,
		Description: "The Electronic Arts intro movie the game plays at boot — " +
			"the first of the disc's 165 streamed Cinepak movies.",
		Stats: stats,
	})
	ctx.Progress("videos", 1, 1, fmt.Sprintf("intro: %dx%d %d frames @%dfps", mv.Width, mv.Height, len(frames), mv.FPS))
	return nil
}

// decodeMovie demuxes and Cinepak-decodes a stream into a slice of RGBA frames.
func decodeMovie(data []byte) (*threedo.CvidMovie, []*image.RGBA, error) {
	mv, err := threedo.DemuxStream(data)
	if err != nil {
		return nil, nil, err
	}
	if mv.Codec != "cvid" && mv.Codec != "" {
		return nil, nil, fmt.Errorf("unsupported codec %q", mv.Codec)
	}
	dec := threedo.NewCvidDecoder(mv.Width, mv.Height)
	frames := make([]*image.RGBA, 0, len(mv.Frames))
	for _, fr := range mv.Frames {
		dec.DecodeFrame(fr)
		cp := image.NewRGBA(dec.Frame().Rect)
		copy(cp.Pix, dec.Frame().Pix)
		frames = append(frames, cp)
	}
	return mv, frames, nil
}

// encodeMP4 pipes our decoded RGBA frames into ffmpeg (H.264, even dimensions
// for yuv420p, moov atom up front) together with the decoded SDX2 PCM as an
// AAC track when the movie has one.
func encodeMP4(path string, frames []*image.RGBA, mv *threedo.CvidMovie, snd *threedo.SndsTrack) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames")
	}
	args := []string{"-y", "-loglevel", "error",
		"-f", "rawvideo", "-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", mv.Width, mv.Height), "-r", fmt.Sprintf("%d", mv.FPS),
		"-i", "-"}
	if snd != nil {
		pcm := snd.DecodeSDX2()
		buf := make([]byte, len(pcm)*2)
		for i, s := range pcm {
			binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
		}
		tmp, err := os.CreateTemp(filepath.Dir(path), "pcm-*.raw")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.Write(buf); err != nil {
			tmp.Close()
			return err
		}
		tmp.Close()
		args = append(args,
			"-f", "s16le", "-ar", fmt.Sprintf("%d", snd.SampleRate),
			"-ac", fmt.Sprintf("%d", snd.Channels), "-i", tmp.Name(),
			"-c:a", "aac", "-b:a", "160k")
	}
	args = append(args,
		"-c:v", "libx264", "-crf", "17", "-pix_fmt", "yuv420p",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-movflags", "+faststart", path)
	cmd := exec.Command("ffmpeg", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	for _, f := range frames {
		if _, err := stdin.Write(f.Pix); err != nil {
			stdin.Close()
			cmd.Wait()
			return err
		}
	}
	stdin.Close()
	return cmd.Wait()
}
