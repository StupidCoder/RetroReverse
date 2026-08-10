// music.go — the soundtrack. The disc carries the whole jukebox as plain
// numbered WMA files under /Sound (01..28, the list the music-select screen
// scrolls through: the OutRun2 remakes, the 1986 originals, the 1989 Turbo
// OutRun tracks, the Euro mixes, the arranged and prototype versions). The
// files are standard WMA v2 (44.1 kHz stereo, 128 kbit/s), so this is a
// container transcode, not a format decode: each track is pulled off the
// disc and re-encoded to MP3 — the Studio's audio format, ffmpeg/libmp3lame
// with the same settings every other game's music uses — with its name taken
// from the disc's own file name.
package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/xbox"
)

func exportMusic(disc *xbox.Image, b *build.Builder) error {
	re := regexp.MustCompile(`^/Sound/(\d{2})_(.+)\.wma$`)
	type track struct{ num, path, name string }
	var tracks []track
	if err := disc.Walk(func(e xbox.Entry) error {
		if e.IsDir {
			return nil
		}
		m := re.FindStringSubmatch(e.Path)
		if m == nil {
			return nil
		}
		// "02_Magical_Sound_Shower" -> "Magical Sound Shower"; Fields
		// collapses the stray double space in "24_Keep_your_Heart_ ARRANGED"
		name := strings.Join(strings.Fields(strings.ReplaceAll(m[2], "_", " ")), " ")
		tracks = append(tracks, track{m[1], e.Path, name})
		return nil
	}); err != nil {
		return err
	}
	sort.Slice(tracks, func(i, j int) bool { return tracks[i].num < tracks[j].num })

	tmp, err := os.MkdirTemp("", "outrun-bgm")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, t := range tracks {
		data, err := disc.ReadFile(t.path)
		if err != nil {
			return err
		}
		src := filepath.Join(tmp, "in.wma")
		if err := os.WriteFile(src, data, 0o644); err != nil {
			return err
		}
		out, err := b.Path("music", "bgm-"+t.num+".mp3")
		if err != nil {
			return err
		}
		if err := audio.EncodeMP3File(src, out); err != nil {
			return err
		}
		dur, err := audio.Duration(out)
		if err != nil {
			return err
		}
		b.AddMedia(schema.Asset{
			ID: "bgm-" + t.num, Category: schema.CategoryMusic,
			Name:     t.name,
			File:     "music/bgm-" + t.num + ".mp3",
			Duration: dur,
			Stats: map[string]any{
				"source":  t.path,
				"on disc": "WMA v2, 44.1 kHz stereo, 128 kbit/s",
			},
		})
	}
	return nil
}
