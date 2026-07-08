// webexport is Marble Madness's single-call format-2 asset exporter (see
// STANDARDS.md §4 / site/FORMAT2.md). It reconstructs everything from the Amiga
// disk image and writes the common asset tree under the output root:
//
//	manifest.json                 game index: native res, tick rate, levels/views/music/sprites
//	levels/<key>.json             per course (kind "tilemap2d"): the format-1 tilemap body
//	                              (grid with the content-hashed variant atlas, objects, cellAnims,
//	                              tileAnims) inside the format-2 envelope
//	levels/<key>.objects.json     machine-readable object DB for the course
//	levels/<key>.atlas.png        the course's tile atlas (base tiles + band/shimmer variants)
//	slopes/<key>.slope.json       the course's 3-D height field + markers (a manifest.views entry,
//	                              the bespoke "marble-slope" three.js view — NOT linked by a level)
//	sprites/index.json + PNGs     the scenery-overlay sprite pieces the level objects reference
//	music/<key>.mp3               each course theme, rendered by the from-scratch Go music player
//
// The four producers (levels, slopes, sprites, music) are folded into this one command; -only
// gates which stages run. None of them boots an emulator — every stage is a pure decode.
//
// Usage: webexport [-adf Marble_Madness.adf] [-o dir] [-only levels,slopes,sprites,music,all]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/marble-madness-amiga/extract/mlb"
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

// Manifest is the format-2 per-game index (site/FORMAT2.md). A stage gated out via -only leaves
// its section empty; omitempty drops it, so a partial run stays self-consistent.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Levels   []LevelIndex   `json:"levels,omitempty"`
	Views    []ViewIndex    `json:"views,omitempty"`
	Music    []MusicEntry   `json:"music,omitempty"`
	Sprites  string         `json:"sprites,omitempty"`
}

type LevelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section,omitempty"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Atlas   string `json:"atlas,omitempty"`
	Objects string `json:"objects,omitempty"`
}

// ViewIndex is a bespoke-3D-view entry (the escape hatch): the course slopes are
// standalone "marble-slope" views, not referenced by any level.
type ViewIndex struct {
	Name    string `json:"name"`
	Section string `json:"section,omitempty"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
}

type MusicEntry struct {
	Name string `json:"name"`
	File string `json:"file"`
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects every stage;
// otherwise only the named stages (levels,slopes,sprites,music) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		switch s {
		case "":
		case "all":
			out["levels"], out["slopes"], out["sprites"], out["music"] = true, true, true, true
		case "levels", "slopes", "sprites", "music":
			out[s] = true
		default:
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want levels,slopes,sprites,music,all)\n", s)
			os.Exit(2)
		}
	}
	return out
}

func main() {
	adfPath := flag.String("adf", "../Marble_Madness.adf", "input Amiga disk image (.adf)")
	in := flag.String("in", "", "input Amiga disk image (.adf) — alias for -adf")
	outDir := flag.String("o", "../../site/public/marble-madness-amiga", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages to run: levels,slopes,sprites,music,all")
	flag.Parse()
	if *in != "" {
		*adfPath = *in
	}
	if flag.NArg() > 0 {
		*adfPath = flag.Arg(0) // tolerate a positional ADF path
	}
	sel := parseOnly(*only)

	raw, err := os.ReadFile(*adfPath)
	if err != nil {
		fail(err)
	}
	vol, err := adf.Open(raw)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fail(err)
	}

	// case-insensitive filename -> path (disk names mix case: AerSnd, ulttrack, BegTrack).
	paths := map[string]string{}
	if err := vol.Walk(func(e adf.Entry) error {
		if !e.IsDir {
			paths[strings.ToLower(e.Name)] = e.Path
		}
		return nil
	}); err != nil {
		fail(err)
	}

	man := Manifest{
		Format: 2, Game: "marble-madness-amiga", Platform: "Amiga",
		Native: map[string]int{"w": 288, "h": 200}, TickHz: 50,
	}
	if sel["levels"] {
		man.Levels = exportLevels(vol, paths, *outDir)
	}
	if sel["slopes"] {
		man.Views = exportSlopes(vol, paths, *outDir)
	}
	if sel["sprites"] {
		exportSprites(vol, paths, *outDir)
		man.Sprites = "sprites/index.json"
	}
	if sel["music"] {
		man.Music = exportMusic(vol, paths, *outDir)
	}
	if err := writeJSON(filepath.Join(*outDir, "manifest.json"), man); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, %d views, %d music, sprites=%q -> %s\n",
		len(man.Levels), len(man.Views), len(man.Music), man.Sprites, *outDir)
}

// course is one course's decoded assets shared by the levels and sprites stages: the tilemap
// (co), the Track hunk image, the parsed display block, and the band-baker over them.
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

// hashedRef returns "<name>?v=<fnv32 of the file>" — a cache-busting URL that changes whenever
// the file's content does.
func hashedRef(path, name string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := fnv.New32a()
	h.Write(b)
	return fmt.Sprintf("%s?v=%08x", name, h.Sum32()), nil
}

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
