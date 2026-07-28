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

// topics are the reachable commentary families, decoded from frontovl's three
// dispatchers (see the writeup's movie-tables section): the race dispatcher at
// 0x8E704 compares the results block against the DriveHS records and the
// opponent, the season dispatcher at 0x9383C against course-total records and
// target times, and the season-end/pursuit paths at 0x93DB8/0x94214/0x94020
// handle the champion's per-car reel, the closing remarks and the arrest.
// Times are 60 Hz ticks (300 = 5 s), speeds 16.16 m/s (0x594CCC ≈ 200 mph).
// Counter fields whose engine-side writers are not yet traced are named by
// their offset in the results block, honestly.
type topic struct {
	fam   string
	takes []int
	name  string
	group string
	desc  string
}

const (
	grpRace    = "Race commentary"
	grpSeason  = "Season commentary"
	grpPursuit = "Pursuit commentary"
	grpMag     = "Magazine"
)

var topics = []topic{
	{"1", []int{1, 2, 3}, "Record smashed", grpRace,
		"Plays when you beat the course record by more than two seconds."},
	{"2", []int{1, 2, 3}, "New record", grpRace,
		"Plays when you beat the course record; the dispatcher picks take 3 directly for the second record class."},
	{"3", []int{2, 3}, "Speed record", grpRace,
		"Plays when you set a new top-speed record for the course. Take 1 was cut."},
	{"5", []int{1, 2, 3}, "Sunday drive", grpRace,
		"Plays when your average speed stayed under roughly sixty mph."},
	{"6", []int{3}, "Two hundred", grpRace,
		"Plays when your top speed reached ~200 mph (the code compares against 89.3 m/s = 199.8 mph). Takes 1-2 were cut."},
	{"7", []int{1, 2, 3}, "Won going away", grpRace,
		"Plays when you won the head-to-head by more than five seconds."},
	{"8", []int{1, 2}, "Photo finish", grpRace,
		"Plays when you won the head-to-head by less than three seconds."},
	{"11", []int{1, 2, 3}, "Left for dead", grpRace,
		"Plays when you lost by more than fifteen seconds."},
	{"12", []int{1, 2, 3}, "Outdriven", grpRace,
		"Plays on a loss with the top-speed comparison against the opponent in your favour."},
	{"13", []int{1, 2, 3}, "Demolition run", grpRace,
		"Plays when a crash counter (results block +0x24) passes five — the host brings a bat."},
	{"14", []int{1, 2, 3}, "Bent metal", grpRace,
		"Plays when the +0x2C crash counter passes two."},
	{"15", []int{1}, "Scenic route", grpRace,
		"Plays when the +0x3C time counter passes five seconds. Takes 2-3 were cut."},
	{"16", []int{1, 2}, "Threading traffic", grpRace,
		"Plays when both the +0x48 counter and the pace gate read high — city and coastal races only."},
	{"17", []int{1, 2}, "Stuck in traffic", grpRace,
		"Plays when the +0x48 counter is high but the pace gate low — city and coastal races only. Take 3 was cut."},
	{"18", []int{1, 2}, "Open road", grpRace,
		"Plays when both gates read low — city and coastal races only."},
	{"20", []int{2, 3}, "Stock answer", grpRace,
		"The fallback family: 20.2 doubles as the reel that plays when no other condition fires. Take 1 was cut."},
	{"21", []int{1, 2}, "High stakes", grpRace,
		"Expert-flag races with the +0x64 stat above 5,500."},
	{"22", []int{1, 2}, "Small change", grpRace,
		"Expert-flag races with the +0x64 stat at or below 5,500."},
	{"23", []int{1, 2}, "Spin cycle", grpRace,
		"Plays when the +0x38 counter passes three."},
	{"24", []int{1, 2, 3}, "Roof first", grpRace,
		"Plays when the +0x70 counter passes five seconds on a slow run."},
	{"26", []int{1, 2}, "Excuses, excuses", grpRace,
		"Plays when the +0x74 counter passes 75 on a slow run."},
	{"27", []int{1, 2}, "Daredevil", grpRace,
		"Plays when the +0x80 counter passes 75 on a fast run."},
	{"28", []int{1, 2, 3}, "Pushing your luck", grpRace,
		"Plays when the +0x80 counter passes 75 on a slow run."},
	{"30", []int{1, 2, 3}, "Third strike", grpRace,
		"Plays when the +0x7C counter passes two."},
	{"31", []int{1}, "One for the road", grpRace,
		"A late fallback pick in the dispatcher's tail. Takes 2-3 were cut."},
	{"32", []int{1, 2}, "Trading paint", grpRace,
		"Plays when the +0x84 event fired on a fast run."},
	{"33", []int{1, 2, 3}, "Body damage", grpRace,
		"Plays when the +0x84 event fired on a slow run."},
	{"34", []int{1, 2, 3}, "Total wreck", grpRace,
		"Plays when the +0x90 counter passes two — the host delivers it from the ground."},

	{"56", []int{1, 2, 3}, "Target crushed", grpSeason,
		"Plays when you beat the tour's target time by more than five seconds against an opponent."},
	{"57", []int{1, 2, 3}, "Target beaten", grpSeason,
		"Plays when you beat the tour's target time against an opponent."},
	{"59", []int{1, 2}, "Target missed", grpSeason,
		"Plays when you missed the tour's target time."},
	{"60", []int{1, 2}, "Target missed badly", grpSeason,
		"Plays when you missed the tour's target time by more than thirty seconds."},
	{"64", []int{1, 2, 3, 4}, "Course record crushed", grpSeason,
		"Plays when you beat the course-total record by more than five seconds."},
	{"65", []int{1, 2, 3}, "Course record", grpSeason,
		"Plays when you beat the course-total record."},
	{"67", []int{1, 2, 3}, "Off the pace", grpSeason,
		"Plays when you finished the tour slower than the course record."},
	{"68", []int{2, 4}, "Way off the pace", grpSeason,
		"Plays when you finished more than thirty seconds behind the course record — the code's 2+rand(2)*2 stride only ever picks the even takes."},
	{"69", []int{1, 2, 3}, "Tour speed record", grpSeason,
		"Plays when you clearly beat the stored top-speed record. Take 4 was cut."},
	{"70", []int{1}, "Speed record, just", grpSeason,
		"Plays when you edged the stored top-speed record. Takes 2-3 were cut."},
	{"43", []int{2, 3}, "Closing remarks, record set", grpSeason,
		"The season sign-off when the record flag (+0x210) is set. Take 1 was cut."},
	{"44", []int{1, 2}, "Closing remarks", grpSeason,
		"The season sign-off without a new record. Take 3 was cut."},
	{"46", []int{1, 2, 3, 4, 5, 6}, "Victory lap", grpSeason,
		"The champion's reel — which takes can play depends on your car, via the per-car table at 0xA859C."},

	{"52", []int{2, 3, 4, 5, 6, 7}, "Booked", grpPursuit,
		"Plays after the trooper's pull-over film when the pursuit ends in an arrest. Take 1 was cut."},

	{"101", []int{1, 2, 3}, "Magazine intro", grpMag,
		"The host's issue-introduction reels."},
	{"101", []int{4, 5, 6}, "Attract host", grpMag,
		"The attract mode's host segment — built as sprintf(\"101.%d\", rand(3)+4), so only takes 4-6 ever play there."},
}

// topicEntries generates the reachable commentary reels' videoSet entries.
func topicEntries() []videoEntry {
	var out []videoEntry
	for _, t := range topics {
		for _, take := range t.takes {
			stem := fmt.Sprintf("%s.%d", t.fam, take)
			name := t.name
			if len(t.takes) > 1 {
				name = fmt.Sprintf("%s %d", t.name, take)
			}
			out = append(out, videoEntry{
				stream: "Movies/" + stem + ".stream",
				id:     "reel-" + strings.ReplaceAll(stem, ".", "-"),
				name:   name, group: t.group, desc: t.desc,
			})
		}
	}
	return out
}

// topicNameByFam names a cut take's family when the family itself is reachable.
func topicNameByFam(fam string) string {
	for _, t := range topics {
		if t.fam == fam {
			return t.name
		}
	}
	return ""
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
		} else if tn := topicNameByFam(fam); tn != "" {
			desc += "a take of the \"" + tn + "\" topic that its random range never picks."
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
	set := append(append([]videoEntry{}, videoSet...), topicEntries()...)
	set = append(set, cutReelEntries()...)
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
