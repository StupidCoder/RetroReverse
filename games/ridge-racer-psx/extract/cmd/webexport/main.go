// webexport extracts Ridge Racer's shipped web assets straight from the CD
// image: one textured GLB per selectable car (the car-select models, with
// wheels) and a textured GLB of the whole course, plus manifest.json. All
// geometry and texture data comes from the rr decoders (MAP.RRM, OBJ.RRO,
// IDX.HED, TEX*.TMS) — verified bit-exact against the running game by
// cmd/geomoracle — and the car catalog comes from the executable's own car
// table (0x80056B40) and name list.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
	"retroreverse.com/tools/platform/psx"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(1)
}

// Manifest is the format-2 asset index the Studio loads.
type Manifest struct {
	Format   int          `json:"format"`
	Game     string       `json:"game"`
	Platform string       `json:"platform"`
	Native   Size         `json:"native"`
	TickHz   int          `json:"tickHz"`
	Models   []ModelIndex `json:"models,omitempty"`
}

type Size struct {
	W int `json:"w"`
	H int `json:"h"`
}

type ModelIndex struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Kind        string `json:"kind"`                  // routes to a Studio renderer plugin
	ObjectsFile string `json:"objectsFile,omitempty"` // placement manifest for the object layer
	Fly         bool   `json:"fly,omitempty"`         // present the item with the fly camera
}

// assets bundles everything the stages need: the decoded files and the
// replayed texture VRAM.
type assets struct {
	objs  []rr.Object
	track *rr.Track
	grid  *rr.Grid
	vrams [3]*rr.VRAM // one texture state per scenery set
	exe   []byte      // text segment (0x80010000..)
}

func main() {
	in := flag.String("in", "../Ridge Racer (Track 01).bin", "PlayStation CD image (.bin)")
	out := flag.String("o", "../../site/public/ridge-racer-psx", "output root")
	only := flag.String("only", "all", "stages to run: models,levels or all")
	flag.Parse()

	stages := map[string]bool{}
	for _, s := range strings.Split(*only, ",") {
		s = strings.TrimSpace(s)
		switch s {
		case "models", "levels", "all":
			stages[s] = true
		default:
			fmt.Fprintf(os.Stderr, "webexport: unknown stage %q (want models,levels,all)\n", s)
			os.Exit(2)
		}
	}
	run := func(stage string) bool { return stages["all"] || stages[stage] }

	data, err := os.ReadFile(*in)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	a, err := load(vol)
	if err != nil {
		die("%v", err)
	}

	man := Manifest{
		Format:   2,
		Game:     "ridge-racer-psx",
		Platform: "Sony PlayStation",
		Native:   Size{W: 320, H: 240},
		TickHz:   60,
	}

	if run("models") {
		models, err := exportCars(a, *out)
		if err != nil {
			die("models: %v", err)
		}
		man.Models = append(man.Models, models...)
	}
	if run("levels") {
		m, err := exportTrack(a, *out)
		if err != nil {
			die("levels: %v", err)
		}
		man.Models = append(man.Models, m)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}
	mj, _ := json.MarshalIndent(man, "", "  ")
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(mj, '\n'), 0o644); err != nil {
		die("%v", err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d models\n", len(man.Models))
}

// load reads and decodes every needed file off the disc.
func load(vol *psx.Volume) (*assets, error) {
	a := &assets{}
	objData, err := vol.ReadFile("OBJ.RRO")
	if err != nil {
		return nil, err
	}
	if a.objs, err = rr.ParseRRO(objData); err != nil {
		return nil, err
	}
	mapData, err := vol.ReadFile("MAP.RRM")
	if err != nil {
		return nil, err
	}
	if a.track, err = rr.ParseRRM(mapData); err != nil {
		return nil, err
	}
	idxData, err := vol.ReadFile("IDX.HED")
	if err != nil {
		return nil, err
	}
	if a.grid, err = rr.ParseIDX(idxData); err != nil {
		return nil, err
	}
	var rects [][]rr.Rect
	for _, name := range []string{"TEX4.TMS", "TEX0.TMS", "TEX1.TMS", "TEX2.TMS", "TEX3.TMS"} {
		d, err := vol.ReadFile(name)
		if err != nil {
			return nil, err
		}
		rs, err := rr.ParseTMS(d)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", name, err)
		}
		rects = append(rects, rs)
	}
	// The race texture states: one per scenery set (the quadrant pages are
	// paged by course progress; everything else is identical across sets).
	a.vrams = rr.NewSceneryVRAMs(rects[0], rects[1], rects[2], rects[3], rects[4])
	_, exe, err := vol.BootEXE()
	if err != nil {
		return nil, err
	}
	a.exe = exe.Text
	return a, nil
}
