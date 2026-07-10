// webexport builds Pilotwings 64's Studio assets straight from the cartridge:
// it boots the ROM from power-on on the oracle, runs the attract sequence to
// the flyby (~field 1200) and title-card (~field 2150) scenes, snapshots RDRAM
// at each scene's graphics task, walks the Fast3D display lists with the f3d
// package, and writes the curated per-object GLBs plus manifest.json.
//
// Nothing is read from a saved machine state: the ROM is the source of
// record, and the whole export reproduces from it in one run (~3 minutes).
// The walk itself is verified against the oracle's RDP triangle stream by
// cmd/dlverify; the asset payloads trace back to their cart offsets via
// cmd/romtrace (see pilotwings-64-n64.md).
//
// Usage:
//
//	webexport -image ROM -o site/public/pilotwings-64-n64
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/pilotwings-64-n64/extract/f3d"
	"retroreverse.com/tools/platform/n64"
)

// The two attract fields the exports come from. The flyby (gyrocopters over
// the island) starts near field 1190; the title card (logo landed, menu up)
// is stable from ~1720 — 2150 matches the savestate the curation was derived
// on. Field numbering is deterministic: the oracle paces the VI by
// instruction count.
const (
	flybyField = 1196
	titleField = 2150
)

// object is one curated GLB: draw groups of a scene merged relative to the
// first group's matrix. The indices come from dlwalk's group table for these
// exact fields and are recorded in pilotwings-64-n64.md.
type object struct {
	name    string // file basename
	label   string // Studio browse-list label
	scene   int    // index into scenes
	groups  []int
	section string
}

var scenes = []int{flybyField, titleField}

var objects = []object{
	{"island", "The island", 1, seq(6, 16), "Island"},
	{"ocean", "Ocean", 1, []int{0, 1}, "Island"},
	{"sky", "Sky dome", 1, seq(2, 5), "Island"},
	{"gyro-a", "Gyrocopter A", 0, append(seq(17, 29), 44, 45), "Gyrocopters"},
	{"gyro-b", "Gyrocopter B", 0, append(seq(30, 43), 46, 47), "Gyrocopters"},
	{"logo-letters", "PILOTWINGS letters", 1, []int{19}, "Logo"},
	{"logo-wing", "Emblem wing", 1, []int{18}, "Logo"},
	{"logo-feather", "Feather", 1, []int{17}, "Logo"},
	{"logo-6", "The 6", 1, []int{20, 21}, "Logo"},
	{"logo-4", "The 4", 1, []int{22, 23}, "Logo"},
}

func seq(lo, hi int) []int {
	var out []int
	for i := lo; i <= hi; i++ {
		out = append(out, i)
	}
	return out
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
	Name    string `json:"name"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Section string `json:"section,omitempty"`
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "cartridge image")
	out := flag.String("o", "", "output directory (site/public/pilotwings-64-n64)")
	flag.Parse()
	if *image == "" || *out == "" {
		die("-image and -o are required")
	}

	rom, err := n64.Load(*image)
	if err != nil {
		die("%v", err)
	}
	m := n64.NewMachine(rom)
	if err := m.Boot(rom, n64.DefaultBoot()); err != nil {
		die("%v", err)
	}

	// Run the attract from power-on, snapshotting RDRAM and the display-list
	// pointer at the first graphics task at or after each target field.
	snaps := make([][]byte, len(scenes))
	dls := make([]uint32, len(scenes))
	fields := 0
	m.OnDisplay = func(*n64.Machine) { fields++ }
	m.OnRSPTask = func(mm *n64.Machine, pc uint32) {
		be := func(off int) uint32 {
			d := mm.DMEM[0xFC0+off:]
			return uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
		}
		if be(0) != 1 { // not a graphics task
			return
		}
		done := true
		for i, f := range scenes {
			if snaps[i] == nil && fields >= f {
				snaps[i] = append([]byte(nil), mm.RDRAM...)
				dls[i] = be(48) & 0x3FFFFF
				fmt.Fprintf(os.Stderr, "webexport: field %d gfx task, dl=%06X\n", fields, dls[i])
			}
			if snaps[i] == nil {
				done = false
			}
		}
		if done {
			mm.StopRequested = true
		}
	}
	res := m.Run(1 << 33)
	fmt.Fprintf(os.Stderr, "webexport: %s\n", res)
	for i, s := range snaps {
		if s == nil {
			die("scene at field %d never captured", scenes[i])
		}
	}

	walkers := make([]*f3d.Walker, len(scenes))
	for i := range scenes {
		w := f3d.New(snaps[i], false)
		w.Walk(dls[i])
		fmt.Fprintf(os.Stderr, "webexport: field %d: %d triangles in %d groups\n",
			scenes[i], w.NTris, len(w.Order))
		walkers[i] = w
	}

	modelsDir := filepath.Join(*out, "models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		die("%v", err)
	}

	man := Manifest{
		Format:   2,
		Game:     "pilotwings-64-n64",
		Platform: "Nintendo 64",
		Native:   Size{320, 240},
		TickHz:   60,
	}
	for _, o := range objects {
		w := walkers[o.scene]
		groups := w.Ordered()
		var gs []*f3d.Group
		for _, gi := range o.groups {
			if gi < 0 || gi >= len(groups) {
				die("%s: group %d out of range (%d groups at field %d) — the walk changed; re-derive the curation with dlwalk", o.name, gi, len(groups), scenes[o.scene])
			}
			gs = append(gs, groups[gi])
		}
		w.WriteObject(modelsDir, o.name, gs)
		man.Models = append(man.Models, ModelIndex{
			Name: o.label, File: "models/" + o.name + ".glb",
			Kind: "mesh3d", Section: o.section,
		})
	}

	j, _ := json.MarshalIndent(man, "", "  ")
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(j, '\n'), 0o644); err != nil {
		die("%v", err)
	}
	fmt.Printf("wrote %s: %d models\n", filepath.Join(*out, "manifest.json"), len(man.Models))
}
