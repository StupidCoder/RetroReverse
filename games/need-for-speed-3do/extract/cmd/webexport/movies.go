package main

// movies.go exports the disc's streamed FMV (Movies/*.stream) as web-playable
// MP4s. The Cinepak ("cvid") video is decoded by our own Go decoder
// (tools/platform/threedo, verified byte-identical to the reference); ffmpeg is
// used only to re-encode our already-decoded RGB frames to H.264 for the
// browser — the same "render ourselves, encode with ffmpeg" split the Turrican
// music export uses. Video only for now (the SDX2 audio track is left for later).
// This stage is opt-in via -movies because it is heavy (many clips) and needs
// ffmpeg on PATH.

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

// exportMovies decodes every Movies/*.stream on the disc and encodes each to
// movies/<name>.mp4, returning browse-list entries (kind "nfs-movie", grouped
// under "Movies") the Studio's movie renderer plays. It returns an error only
// for setup problems (missing ffmpeg); individual movies that fail to decode are
// skipped with a warning so one bad clip doesn't sink the run.
func exportMovies(vol *threedo.Volume, out string) ([]ModelIndex, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found on PATH (needed to encode movies): %w", err)
	}
	moviesDir := filepath.Join(out, "movies")
	if err := os.MkdirAll(moviesDir, 0o755); err != nil {
		return nil, err
	}

	// Collect the movie paths first so ordering is stable.
	var paths []string
	vol.Walk(func(e threedo.Entry) error {
		if !e.IsDir && strings.HasSuffix(strings.ToLower(e.Name), ".stream") {
			paths = append(paths, e.Path)
		}
		return nil
	})
	sort.Strings(paths)

	var entries []ModelIndex
	for _, p := range paths {
		data, err := vol.ReadFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [movie] skip %s: %v\n", p, err)
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		mp4 := filepath.Join(moviesDir, stem+".mp4")
		mv, frames, err := decodeMovie(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [movie] skip %s: %v\n", p, err)
			continue
		}
		if err := encodeMP4(mp4, frames, mv.Width, mv.Height, mv.FPS); err != nil {
			fmt.Fprintf(os.Stderr, "  [movie] encode %s: %v\n", p, err)
			continue
		}
		entries = append(entries, ModelIndex{
			Name:    stem,
			File:    "movies/" + stem + ".mp4",
			Kind:    "nfs-movie", Section: "Movies",
			W: mv.Width, H: mv.Height,
		})
		fmt.Fprintf(os.Stderr, "  [movie] %-16s %dx%d %d frames @%dfps\n", stem, mv.Width, mv.Height, len(frames), mv.FPS)
	}
	return entries, nil
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

// encodeMP4 pipes raw RGBA frames into ffmpeg to produce an H.264 MP4 sized to
// even dimensions (yuv420p requires it) at the movie's frame rate.
func encodeMP4(path string, frames []*image.RGBA, w, h, fps int) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames")
	}
	if fps <= 0 {
		fps = 15
	}
	cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error",
		"-f", "rawvideo", "-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", w, h), "-r", fmt.Sprintf("%d", fps),
		"-i", "-",
		"-c:v", "libx264", "-pix_fmt", "yuv420p",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-movflags", "+faststart", path)
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
