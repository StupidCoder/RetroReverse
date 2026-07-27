// webexport builds Marble Madness's Retro-X game tree from the Amiga disk
// image. Every stage is a pure decode — nothing boots an emulator:
//
//	levels/<key>.json       per course: tilemap (grid with the colour bands
//	                        baked as recoloured variant tiles, the gold
//	                        shimmer as tileAnims, screen swaps as cellAnims)
//	                        and the scenery-overlay placements
//	levels/<key>.atlas.png  the course's tile atlas (base + variant tiles)
//	objects/<id>.json|.png  the scenery-overlay pieces (drawbridge, goal
//	                        flags, the WAVE, pistons, …) as sprite2d objects
//	objects/<key>-slope.glb the course's 3-D height field as ONE GLB: solid
//	                        terrain triangles + the Track-layer markers baked
//	                        as coloured line pins/routes
//	music/<key>.mp3         each course theme, rendered by the from-scratch
//	                        Go music player
//
// Usage (from games/marble-madness-amiga/):
//
//	go run ./extract/cmd/webexport -in Marble_Madness.adf
package main

import (
	"fmt"
	"os"
	"strings"

	"retroreverse.com/games/marble-madness-amiga/extract/mlb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/amiga/hunk"
)

// courses in play order: key (.mlb basename), Track filename, Snd filename, display name.
var courses = []struct{ key, track, snd, name string }{
	{"practy", "PrcTrack", "PrcSnd", "Practice"},
	{"beginr", "BegTrack", "BegSnd", "Beginner"},
	{"interm", "IntTrack", "IntSnd", "Intermediate"},
	{"aerial", "AerTrack", "AerSnd", "Aerial"},
	{"silly", "SilTrack", "SilSnd", "Silly"},
	{"ultima", "UltTrack", "UltSnd", "Ultimate"},
}

func main() {
	cli.Main("marble-madness-amiga", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in Marble_Madness.adf [-o DIR] [-only levels,objects,music]")
	}
	raw, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	vol, err := adf.Open(raw)
	if err != nil {
		return err
	}

	// case-insensitive filename -> path (disk names mix case: AerSnd, ulttrack, BegTrack).
	paths := map[string]string{}
	if err := vol.Walk(func(e adf.Entry) error {
		if !e.IsDir {
			paths[strings.ToLower(e.Name)] = e.Path
		}
		return nil
	}); err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Marble Madness")
	b.SetPlatform("Amiga")
	b.SetYear(1986)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 288, H: 200},
		TickHz: 50,
		Filter: "crt",
	})

	var refs map[string]objRef
	if ctx.Stage("objects") {
		if refs, err = exportObjects(ctx, vol, paths); err != nil {
			return err
		}
		if err := exportSlopes(ctx, vol, paths); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportLevels(ctx, vol, paths, refs); err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportMusic(ctx, vol, paths); err != nil {
			return err
		}
	}
	return nil
}

// objRef points a placement's sprite key ("<course>/<piece>") at its object asset.
type objRef struct{ asset string }

// course is one course's decoded assets shared by the levels and objects
// stages: the tilemap (co), the Track hunk image, the parsed display block,
// and the band-baker over them.
type course struct {
	co   *mlb.Course
	prog *hunk.Program
	fx   *displayFx
	bake *bandBake
}

// loadCourse decodes a course's .mlb tilemap + Track file and builds the colour-band baker.
func loadCourse(vol *adf.Volume, paths map[string]string, key, track string) (course, error) {
	mp, ok := paths[key+".mlb"]
	if !ok {
		return course{}, fmt.Errorf("%s.mlb not found on disk", key)
	}
	d, err := vol.ReadFile(mp)
	if err != nil {
		return course{}, err
	}
	co := mlb.Decode(d)

	tp, ok := paths[strings.ToLower(track)]
	if !ok {
		return course{}, fmt.Errorf("%s not found on disk", track)
	}
	td, err := vol.ReadFile(tp)
	if err != nil {
		return course{}, err
	}
	prog, err := hunk.Load(td, 0)
	if err != nil {
		return course{}, fmt.Errorf("%s: hunk load: %w", track, err)
	}
	fx := parseDisplayFx(prog.Image)
	return course{co: co, prog: prog, fx: fx, bake: newBandBake(co, fx)}, nil
}

// slugify turns a sprite key or name into a stable kebab-case id
// ("practy/a10" -> "practy-a10").
func slugify(s string) string {
	var out []rune
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+32)
		default:
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// chk/fail keep the untouched decode files (overlay.go, bake.go, bands.go) compiling.
func chk(e error) {
	if e != nil {
		fail(e)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
