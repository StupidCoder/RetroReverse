// webexport publishes the decoded room to the Studio as a Retro-X game.
//
// The room is decoded from the ROM (the same path roomexport verifies against
// the running game), the two background layers are composited as the hardware
// stacks them, and the result is emitted as a tilemap: an atlas of DISTINCT
// 8x8 cells plus an index grid.
//
// Deduplicating by pixel CONTENT rather than by the game's tile id is
// deliberate. A GBA cell is a tile id plus a palette bank and two flip bits, so
// one id renders as up to 64 different pictures; keying the atlas on the id
// alone would collapse them into one and repaint half the map wrong. Content is
// what the viewer actually draws.
//
//	webexport [-out DIR]
//
// Every AREA becomes one level: the game's own tables say where each of an
// area's rooms sits on a shared canvas, so 419 room maps publish as 79 places
// a player would recognise.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	out := flag.String("out", "../../../site/public/zelda-minish-cap-gba", "Retro-X game directory")
	maxArea := flag.Int("max", 128, "highest area id to consider")
	maxCells := flag.Int("maxcells", 24000, "skip an area whose atlas would exceed this many distinct cells")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}
	levelsDir := filepath.Join(*out, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		die("%v", err)
	}

	type entry struct {
		id, name, group string
		rooms           int
		w, h            int
	}
	var made []entry
	var skipped []string
	for a := 0; a < *maxArea; a++ {
		ar, err := LoadArea(rom, a)
		if err != nil {
			continue
		}
		as, err := buildArea(rom, ar)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("area %d: %v", a, err))
			continue
		}
		img := as.img
		wt, ht := img.Bounds().Dx()/8, img.Bounds().Dy()/8

		// Slice into 8x8 cells, deduplicated by CONTENT (see the note above).
		index := map[[32]byte]int{}
		var cellImgs []*image.RGBA
		cells := make([]int, wt*ht)
		for ty := 0; ty < ht; ty++ {
			for tx := 0; tx < wt; tx++ {
				sub := image.NewRGBA(image.Rect(0, 0, 8, 8))
				draw.Draw(sub, sub.Bounds(), img, image.Pt(tx*8, ty*8), draw.Src)
				k := sha256.Sum256(sub.Pix)
				id, ok := index[k]
				if !ok {
					id = len(cellImgs)
					index[k] = id
					cellImgs = append(cellImgs, sub)
				}
				cells[ty*wt+tx] = id
			}
		}
		if len(cellImgs) > *maxCells {
			skipped = append(skipped, fmt.Sprintf("area %d: %d distinct cells exceeds the %d budget", a, len(cellImgs), *maxCells))
			continue
		}

		const cols = 16
		rows := (len(cellImgs) + cols - 1) / cols
		atlas := image.NewRGBA(image.Rect(0, 0, cols*8, rows*8))
		for i, c := range cellImgs {
			x, y := (i%cols)*8, (i/cols)*8
			draw.Draw(atlas, image.Rect(x, y, x+8, y+8), c, image.Point{}, draw.Src)
		}
		slug := fmt.Sprintf("area-%02d", a)
		writePNG(filepath.Join(levelsDir, slug+"-atlas.png"), atlas)
		writeJSON(filepath.Join(levelsDir, slug+".json"), map[string]any{
			"format": "retro-x", "version": 1, "type": "tilemap",
			// maxZoom is (screenWidth / view.w) * maxNativeFactor, so a view that
			// frames the WHOLE area — which is what makes it open fitted — would
			// cap zoom at "the whole area on screen" if the factor were 1. Scaling
			// the factor by the view's width in native screens cancels the view out
			// and leaves the usual limit: about 6x the console's pixels, the same
			// range a level with a native-sized view gets.
			"camera": map[string]any{"mode": "map2d", "map2d": map[string]any{
				"maxNativeFactor": float64(wt*8) / 240.0,
			}},
			"tilemap": map[string]any{
				"tileSize": 8, "width": wt, "height": ht,
				"atlas": map[string]any{"file": slug + "-atlas.png", "cols": cols},
				"cells": cells,
				// Frame the whole area. A canvas is the bounding box of the
				// area's rooms, so its corner is usually empty space — opening
				// at a native-sized window there shows a black screen.
				"view": map[string]any{"x": 0, "y": 0, "w": wt * 8, "h": ht * 8},
			},
		})
		group := "Overworld"
		if a >= 32 {
			group = "Interiors & dungeons"
		}
		name := fmt.Sprintf("Area %d", a)
		if a == 3 {
			name = "Hyrule Field (area 3)"
		}
		made = append(made, entry{slug, name, group, as.rooms, img.Bounds().Dx(), img.Bounds().Dy()})
	}

	assets := make([]any, 0, len(made))
	for _, e := range made {
		assets = append(assets, map[string]any{
			"id": e.id, "category": "level", "name": e.name, "group": e.group,
			"description": fmt.Sprintf("%d room%s assembled at the positions the game's own area table gives them — %dx%d pixels. Decoded from the cartridge: LZ77 streams, a metatile grid expanded through its 2x2 table, and the area's palette set.",
				e.rooms, plural(e.rooms), e.w, e.h),
			"file": "levels/" + e.id + ".json",
		})
	}
	writeJSON(filepath.Join(*out, "manifest.json"), map[string]any{
		"format": "retro-x", "version": 1,
		"id": "zelda-minish-cap-gba", "title": "The Legend of Zelda: The Minish Cap",
		"platform": "Game Boy Advance", "year": 2004,
		"description": "The 2004 Game Boy Advance Zelda, developed by Capcom/Flagship — internally, as its own build string admits, \"ZELDA 5\". Every map here is decoded straight from the cartridge. A room is not stored as a tilemap at all: it is a grid of metatile ids, each expanding through an 8-byte table into a 2x2 block of tiles, behind the BIOS's LZ77 compression. Rooms are not standalone either — the game's area tables place each one at a pixel position on a shared canvas, so what you see is the whole place assembled, not a screen at a time. The decode is checked against the game itself: the decompressed tables are byte-identical to what the console's own BIOS call produces, and the expanded map reproduces every tile the running game uploaded to video memory.",
		// filter "gba": the AGB's unlit reflective panel (site/src/screenfx.js).
		"display": map[string]any{"native": map[string]any{"w": 240, "h": 160}, "tickHz": 60, "filter": "gba"},
		"docs": []any{
			map[string]any{"id": "metatiles", "title": "A map of blocks", "file": "docs/metatiles.html"},
			map[string]any{"id": "rooms", "title": "What a room is", "file": "docs/rooms.html"},
			map[string]any{"id": "areas", "title": "Areas, and how they join", "file": "docs/areas.html"},
		},
		"assets": assets,
	})
	fmt.Printf("published %d areas\n", len(made))
	for _, s := range skipped {
		fmt.Println("skipped:", s)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		die("%v", err)
	}
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		die("%v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		die("%v", err)
	}
}
