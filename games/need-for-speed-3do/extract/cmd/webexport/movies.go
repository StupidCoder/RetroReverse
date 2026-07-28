package main

// movies.go is the Retro-X `videos` stage: the disc's eighteen NAMED movies —
// the boot/logo films, the attract reels, the pull-over films, the victory
// montage and the eight car showcase films — plus the "Unused" group: the 33
// cut commentary reels the movie-tables RE proved unreachable. Each becomes
// videos/<id>.mp4. (The ~115 reachable numbered reels stay unexported until
// the dispatchers' conditions are decoded enough to title them by outcome.)
// The Cinepak video is decoded by our own Go decoder (tools/platform/threedo/
// cvid.go, byte-identical to the reference) and the SDX2 audio track by our
// own DPCM decoder (snds.go, byte-identical to FFmpeg's sdx2_dpcm across
// every movie on the disc); ffmpeg only re-encodes those already-decoded
// frames and samples to H.264 + AAC for the browser (moov up front,
// RETROX.md §9).

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/threedo"
)

// videoEntry is one curated video asset: a .Stream on the disc plus its
// Studio identity.
type videoEntry struct {
	stream, id, name, group, desc string
	related                       []string
}

// videoSet is the curated named-movie list — all eighteen named .Streams on
// the disc (the generated "Unused" group of cut reels follows, cutReelEntries).
// Showcase films cross-link the car model they present via `related`; the
// block is in the game's own car-table order (frontovl's stem table at
// 0xA7B28: supra diablo 911 vette 512tr viper nsx rx7, indexed by car id —
// reordered here to match the manifest's Player cars group).
var videoSet = []videoEntry{
	{"Movies/eac.stream", "intro", "Electronic Arts logo", "Boot films",
		"The Electronic Arts logo film — the first of the disc's 165 streamed Cinepak movies the game arms at boot.", nil},
	{"Movies/pioneer.stream", "pioneer", "Pioneer logo", "Boot films",
		"Pioneer's starfield logo film, on the disc beside the EA logo (not requested in our traced boot).", nil},
	{"Movies/title.stream", "title", "Title film", "Boot films",
		"Road & Track Presents: The Need for Speed — the torn-paper tachometer title montage.", nil},
	{"Movies/attract1.stream", "attract1", "Attract reel 1", "Attract reels",
		"Live-action driving montage — first of the three attract reels.", nil},
	{"Movies/attract2.stream", "attract2", "Attract reel 2", "Attract reels",
		"Live-action driving montage — second of the three attract reels.", nil},
	{"Movies/attract3.stream", "attract3", "Attract reel 3", "Attract reels",
		"Live-action driving montage — third of the three attract reels.", nil},
	{"Movies/cop1.stream", "pullover1", "Pullover 1", "Pursuit",
		"The state trooper's roadside lecture — first of the three pull-over films.", nil},
	{"Movies/cop2.stream", "pullover2", "Pullover 2", "Pursuit",
		"The state trooper's roadside lecture — second of the three pull-over films.", nil},
	{"Movies/cop3.stream", "pullover3", "Pullover 3", "Pursuit",
		"The state trooper's roadside lecture — third of the three pull-over films.", nil},
	{"Movies/win.stream", "win", "Victory film", "Victory",
		"The winner's montage.", nil},
	{"Movies/diablo.stream", "showcase-diablo", "Diablo", "Car showcase",
		"The showcase film for the Lamborghini Diablo VT.", []string{"car-ldiablo"}},
	{"Movies/512tr.stream", "showcase-512tr", "512 TR", "Car showcase",
		"The showcase film for the Ferrari 512 TR.", []string{"car-f512tr"}},
	{"Movies/911.stream", "showcase-911", "911", "Car showcase",
		"The showcase film for the Porsche 911.", []string{"car-p911"}},
	{"Movies/vette.stream", "showcase-zr1", "ZR-1", "Car showcase",
		"The showcase film for the Chevrolet Corvette ZR-1.", []string{"car-czr1"}},
	{"Movies/viper.stream", "showcase-viper", "Viper", "Car showcase",
		"The showcase film for the Dodge Viper RT/10.", []string{"car-dviper"}},
	{"Movies/nsx.stream", "showcase-nsx", "NSX", "Car showcase",
		"The showcase film for the Acura NSX.", []string{"car-ansx"}},
	{"Movies/rx7.stream", "showcase-rx7", "RX-7", "Car showcase",
		"The showcase film for the Mazda RX-7.", []string{"car-mrx7"}},
	{"Movies/supra.stream", "showcase-supra", "Supra", "Car showcase",
		"The showcase film for the Toyota Supra.", []string{"car-tsupra"}},
}

// cutReels are the 33 orphaned commentary reels — on the disc, but no shipped
// code path can build their stem (the movie-tables RE walked every dispatcher:
// 132 reachable references, these 33 files outside them). Families 35-40, 42
// and 51 are whole topics the dispatchers never mention; the rest are takes
// the reachable topics' random ranges skip.
var cutFamilies = map[string]bool{"35": true, "36": true, "37": true, "38": true,
	"39": true, "40": true, "42": true, "51": true}

var cutReels = []string{
	"3.1", "6.1", "6.2", "16.3", "20.1", "31.2", "31.3",
	"35.1", "35.2", "35.3", "36.1", "36.2", "36.3", "37.1", "37.2",
	"38.1", "38.2", "39.1", "39.2", "39.3", "40.1", "40.2", "40.3",
	"42.1", "42.2", "44.3", "51.1", "51.2", "51.3", "52.1",
	"69.4", "70.2", "70.3",
}

// cutReelEntries generates the "Unused" group's videoSet entries.
func cutReelEntries() []videoEntry {
	var out []videoEntry
	for _, stem := range cutReels {
		fam := strings.SplitN(stem, ".", 2)[0]
		desc := "Post-race commentary take " + stem + ", cut: "
		if cutFamilies[fam] {
			desc += "its whole topic family is unreferenced by the shipped dispatchers."
		} else {
			desc += "the take sits outside its topic's random range."
		}
		if stem == "38.1" {
			desc += " The disc's only silent movie."
		}
		out = append(out, videoEntry{
			stream: "Movies/" + stem + ".stream",
			id:     "reel-" + strings.ReplaceAll(stem, ".", "-"),
			name:   stem, group: "Unused", desc: desc,
		})
	}
	return out
}

// showcaseByCar inverts videoSet's related links: player-car object asset id →
// its showcase film's video asset id, so the cars can link back to their films.
func showcaseByCar() map[string]string {
	m := map[string]string{}
	for _, v := range videoSet {
		for _, r := range v.related {
			m[r] = v.id
		}
	}
	return m
}

// exportVideos decodes and registers every movie in videoSet plus the cut
// reels of the Unused group.
func exportVideos(ctx *cli.Context, vol *threedo.Volume) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found on PATH (needed to encode the videos): %w", err)
	}
	set := append(append([]videoEntry{}, videoSet...), cutReelEntries()...)
	for i, v := range set {
		raw, err := vol.ReadFile(v.stream)
		if err != nil {
			return fmt.Errorf("%s: %w", v.stream, err)
		}
		mv, frames, err := decodeMovie(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", v.stream, err)
		}
		snd, err := threedo.DemuxSnds(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", v.stream, err)
		}

		mp4, err := ctx.Builder.Path("videos", v.id+".mp4")
		if err != nil {
			return err
		}
		if err := encodeMP4(mp4, frames, mv, snd); err != nil {
			return fmt.Errorf("%s: %w", v.stream, err)
		}

		duration := float64(len(frames)) / float64(mv.FPS)
		length := fmt.Sprintf("%d frames @ %d fps", len(frames), mv.FPS)
		if mv.HeaderRate != mv.FPS {
			// 150 of the disc's streams declare 30 in the film header; the
			// frame clock (240 Hz ticks) and the audio track both say 15,
			// and 30 would out-run the double-speed drive.
			length += fmt.Sprintf(" (the header claims %d)", mv.HeaderRate)
		}
		stats := map[string]any{
			"Source":    fmt.Sprintf("%s (%.1f MB)", v.stream, float64(len(raw))/(1024*1024)),
			"Codec":     fmt.Sprintf("Cinepak (%q), software-decoded on the ARM60", mv.Codec),
			"Data rate": fmt.Sprintf("%.0f KB/s streamed off the CD", float64(len(raw))/duration/1024),
			"Native":    fmt.Sprintf("%d × %d px", mv.Width, mv.Height),
			"Colors":    "vector-quantised YCbCr 4:2:0, shown as 15-bit RGB555",
			"Length":    length,
		}
		if snd != nil {
			stats["Audio"] = fmt.Sprintf("SDX2 DPCM, %d Hz, %d channels (2:1)", snd.SampleRate, snd.Channels)
		}
		ctx.Builder.AddMedia(schema.Asset{
			ID: v.id, Category: schema.CategoryVideo, Name: v.name, Group: v.group,
			File:    "videos/" + v.id + ".mp4",
			Related: v.related,
			W:       mv.Width, H: mv.Height, FPS: float64(mv.FPS), Duration: duration,
			Description: v.desc,
			Stats:       stats,
		})
		ctx.Progress("videos", i+1, len(set),
			fmt.Sprintf("%s: %dx%d %d frames @%dfps", v.id, mv.Width, mv.Height, len(frames), mv.FPS))
	}
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
