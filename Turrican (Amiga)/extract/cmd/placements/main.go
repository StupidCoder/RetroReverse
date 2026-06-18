// Command placements extracts Turrican's object placement lists from the disk —
// where each object is seeded in every scene of every world (Part V) — via the
// shared extract/scene package, which faithfully reproduces the scroll-triggered
// spawner (resident $1710): a 2D bucket grid (descriptor +$24/+$28) of per-column
// 6-byte entry runs (`type.w, x.w, y.w`), the entry type being the high byte of the
// type word. type<3 selects the resident AI handler table $1A60[type-1]; type>=3
// selects the scene descriptor's +$20 table[type-3]. Each handler installs the
// object's sprite (a frame table, via `MOVE.l #ft,$12(a5)`).
//
// Output: one JSON per scene with the resolved objects and a per-type AI summary,
// for a viewer object layer.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"turrican/extract/scene"
)

type aiEntry struct {
	Type     int  `json:"type"`     // entry type (high byte of the type word)
	Handler  int  `json:"handler"`  // resolved AI handler address
	Sprite   int  `json:"sprite"`   // frame table the handler installs (the sprite)
	Resident bool `json:"resident"` // sprite/handler live in resident space
}
type object struct {
	Type     int  `json:"type"`
	X        int  `json:"x"` // tile column
	Y        int  `json:"y"` // tile row
	Sprite   int  `json:"sprite"`
	Resident bool `json:"resident"`
}
type sceneObjects struct {
	World, Scene, Width, Height int
	AI                          []aiEntry `json:"ai"`
	Objects                     []object
}

func main() {
	out := flag.String("o", "rendered/placements", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}
	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fail(err)
	}
	game, err := scene.Load(adf)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	for w := 0; w < scene.NumWorlds; w++ {
		for _, sc := range game.Scenes(w) {
			// Per-type AI summary (distinct types used in the scene).
			aiByType := map[int]aiEntry{}
			var objs []object
			for _, o := range sc.Objects {
				aiByType[o.Type] = aiEntry{Type: o.Type, Handler: o.Handler, Sprite: o.FT, Resident: o.Resident}
				objs = append(objs, object{Type: o.Type, X: o.X / 32, Y: o.Y / 32, Sprite: o.FT, Resident: o.Resident})
			}
			var ai []aiEntry
			for _, e := range aiByType {
				ai = append(ai, e)
			}
			sort.Slice(ai, func(i, j int) bool { return ai[i].Type < ai[j].Type })

			so := sceneObjects{World: w, Scene: sc.Index, Width: sc.Width, Height: sc.Height, AI: ai, Objects: objs}
			name := fmt.Sprintf("world%d_scene%d.json", w, sc.Index)
			b, _ := json.Marshal(so)
			if err := os.WriteFile(filepath.Join(*out, name), b, 0o644); err != nil {
				fail(err)
			}
			fmt.Printf("world %d scene %d: %d objects -> %s\n", w, sc.Index, len(objs), name)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "placements:", err)
	os.Exit(1)
}
