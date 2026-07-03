// webexport renders the Marble Madness course tilemaps for the companion
// website. For each course it decodes the .mlb (Marble_Madness.md Part IV §3)
// and writes the assembled course as a 1x PNG (the viewer scales it up), plus a
// meta.json listing the levels in play order.
//
// Usage: webexport [-adf disk.adf] [-o dir]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"

	"marblemad/extract/mlb"
	"marblemad/extract/slope"
	"retroreverse.com/tools/amiga/adf"
	"retroreverse.com/tools/amiga/hunk"
	"retroreverse.com/tools/c64/gfx"
)

// courses in play order: key (.mlb basename), Track filename, display name.
var courses = []struct{ key, track, name string }{
	{"practy", "PrcTrack", "Practice"},
	{"beginr", "BegTrack", "Beginner"},
	{"interm", "IntTrack", "Intermediate"},
	{"aerial", "AerTrack", "Aerial"},
	{"silly", "SilTrack", "Silly"},
	{"ultima", "UltTrack", "Ultimate"},
}

type metaLevel struct {
	Name  string `json:"name"`
	File  string `json:"file"`  // per-course tilemap JSON (common format, site/FORMAT.md)
	Atlas string `json:"atlas"` // the course's tile-atlas PNG
	Slope string `json:"slope"` // the 3-D height field (outside the common format)
}

// slopeJSON is the per-course height field for the 3-D view: a dense grid (w×h,
// row-major from x0,y0) where each cell is the tile's height above lo plus one,
// or 0 where there is no rolling surface (a pit). lo/hi give the real range.
type slopeJSON struct {
	X0      int        `json:"x0"`
	Y0      int        `json:"y0"`
	W       int        `json:"w"`
	H       int        `json:"h"`
	Lo      int        `json:"lo"`
	Hi      int        `json:"hi"`
	Heights []int      `json:"heights"`
	Markers markerJSON `json:"markers"`
}

// markerJSON is the course's Track-layer overlays for the 3-D view: single pins
// (Points) and patrol routes (Paths), each carrying an RGB colour, in tile coords.
type markerJSON struct {
	Points []ptJSON   `json:"points"`
	Paths  []pathJSON `json:"paths"`
}
type ptJSON struct {
	X int `json:"x"`
	Y int `json:"y"`
	C int `json:"c"`
}
type pathJSON struct {
	C   int      `json:"c"`
	Pts [][2]int `json:"pts"`
}

// shimmerPeriod is the vblank frames each gold-rotation phase holds: the
// stepper $84DC runs per WORLD frame (from the $AF7A world-update tick, 2
// vblanks each), adding 2 to the sub-counter $68F and stepping the phase $68E
// past 6 — every 4th world frame = 8 vblanks.
const shimmerPeriod = 8

// marker colours, matching the offline *.wire.png overlays.
const (
	colPlacement = 0x46d4ff // cyan
	colOoze      = 0xff9430 // orange
	colDynRegion = 0xffe000 // yellow
	colMarble    = 0xff46c8 // magenta
	colSlinky    = 0x46e05a // green
)

func buildMarkers(m slope.MarkerSet) markerJSON {
	mj := markerJSON{Points: []ptJSON{}, Paths: []pathJSON{}}
	add := func(pts [][2]int, c int) {
		for _, p := range pts {
			mj.Points = append(mj.Points, ptJSON{p[0], p[1], c})
		}
	}
	add(m.Placement, colPlacement)
	add(m.Ooze, colOoze)
	add(m.DynRegion, colDynRegion)
	for _, p := range m.Marbles {
		mj.Paths = append(mj.Paths, pathJSON{colMarble, p})
	}
	for _, p := range m.Slinkies {
		mj.Paths = append(mj.Paths, pathJSON{colSlinky, p})
	}
	return mj
}

func main() {
	adfPath := flag.String("adf", "../Marble_Madness.adf", "disk image")
	outDir := flag.String("o", "../../site/public/marble", "output directory")
	flag.Parse()
	if err := run(*adfPath, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func run(adfPath, outDir string) error {
	raw, err := os.ReadFile(adfPath)
	if err != nil {
		return err
	}
	vol, err := adf.Open(raw)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// case-insensitive filename -> path
	paths := map[string]string{}
	if err := vol.Walk(func(e adf.Entry) error {
		if !e.IsDir {
			paths[strings.ToLower(e.Name)] = e.Path
		}
		return nil
	}); err != nil {
		return err
	}

	spritesDir := filepath.Join(outDir, "sprites")
	if err := os.MkdirAll(spritesDir, 0o755); err != nil {
		return err
	}
	spriteIndex := map[string]any{}

	var levels []metaLevel
	for _, c := range courses {
		p, ok := paths[c.key+".mlb"]
		if !ok {
			return fmt.Errorf("%s.mlb not found on disk", c.key)
		}
		d, err := vol.ReadFile(p)
		if err != nil {
			return err
		}
		// Atlas + tilemap: the course is a row-major grid of 8x8 tile indices drawn
		// from the course's own tile atlas (the same atlas+tilemap form the SML viewer
		// uses), not a pre-composited image.
		co := mlb.Decode(d)

		// Track file: the slope field, markers, and the scenery-overlay pieces.
		tp, ok := paths[strings.ToLower(c.track)]
		if !ok {
			return fmt.Errorf("%s not found on disk", c.track)
		}
		td, err := vol.ReadFile(tp)
		if err != nil {
			return err
		}
		prog, err := hunk.Load(td, 0)
		if err != nil {
			return fmt.Errorf("%s: hunk load: %w", c.track, err)
		}

		// Bake the display block's colour bands (Part IV §6) into the atlas:
		// rows past a band boundary swap to recoloured variant tiles, so the
		// map fades exactly as the game's scroll-following copper list does.
		fx := parseDisplayFx(prog.Image)
		bake := newBandBake(co, fx)

		// The scenery overlays (sprite pieces + their animations) and the
		// screen-swap tile animations, from the Track's dynamic-region scripts.
		// Strips render with the colour-band palette at each piece's row.
		objects, cellAnims, err := exportOverlays(vol, paths, c.key, prog.Image, co, bake.paletteAt, spritesDir, spriteIndex)
		if err != nil {
			return err
		}
		cells := make([]int, mlb.CourseW*co.PlayableH)
		tilesInBand := map[int]map[int]bool{}
		note := func(band, t int) {
			if tilesInBand[band] == nil {
				tilesInBand[band] = map[int]bool{}
			}
			tilesInBand[band][t] = true
		}
		for r := 0; r < co.PlayableH; r++ {
			band := bake.bandAt(r)
			for x := 0; x < mlb.CourseW; x++ {
				t := co.Cells[r*mlb.CourseW+x]
				note(band, t)
				cells[r*mlb.CourseW+x] = bake.tileFor(t, band)
			}
		}
		// swap-variant rows repaint into the finale band: substitute per
		// destination row, and count their tiles toward that band's shimmer.
		for _, ca := range cellAnims {
			ty, tw := ca["ty"].(int), ca["tw"].(int)
			for _, ph := range ca["phases"].([]map[string]any) {
				tiles := ph["tiles"].([]int)
				for i, t := range tiles {
					band := bake.bandAt(ty + i/tw)
					note(band, t)
					tiles[i] = bake.tileFor(t, band)
				}
			}
		}
		tileAnims := bake.shimmerAnims(tilesInBand, shimmerPeriod)

		// The tilemap's cells reference variant tiles appended to the atlas, so
		// the level JSON and the PNG must never mix versions in a client cache:
		// reference the atlas by a content-hashed URL.
		atlasFile := c.key + ".atlas.png"
		if err := gfx.WritePNG(filepath.Join(outDir, atlasFile), co.AtlasVariants(16, bake.ext, bake.varList)); err != nil {
			return err
		}
		atlasRef, err := hashedRef(filepath.Join(outDir, atlasFile), atlasFile)
		if err != nil {
			return err
		}

		// The map shows the PLAYABLE rows (the .mlb header count = the engine's
		// scroll clamp). Data rows beyond it are off-screen variant storage
		// (Ultimate's three hidden final-screen variants, Silly's cropped wall
		// row) — the swap animation replays them from the full cell array.
		level := map[string]any{
			"format": 1,
			"name":   c.name,
			"grid": map[string]any{
				"tileSize": 8, "atlas": atlasRef, "atlasCols": 16, "atlasGutter": 1,
				"width": mlb.CourseW, "height": co.PlayableH,
				"cells": cells,
			},
			"objects": objects,
			// Frame the Amiga's on-screen view (288x200 playfield) at the course top.
			"view": map[string]any{"x": (mlb.CourseW*8 - 288) / 2, "y": 0, "w": 288, "h": 200},
		}
		if len(cellAnims) > 0 {
			level["cellAnims"] = cellAnims
		}
		if len(tileAnims) > 0 {
			level["tileAnims"] = tileAnims
		}
		file := c.key + ".json"
		if err := writeJSON(filepath.Join(outDir, file), level); err != nil {
			return err
		}
		h := co.PlayableH

		sj := buildSlope(slope.Build(prog.Image))
		sj.Markers = buildMarkers(slope.Markers(prog.Image))
		slopeFile := c.key + ".slope.json"
		if err := writeJSON(filepath.Join(outDir, slopeFile), sj); err != nil {
			return err
		}

		levels = append(levels, metaLevel{Name: c.name, File: file, Atlas: atlasRef, Slope: slopeFile})
		fmt.Printf("%-12s %s  %d×%d tiles, %d tiles; slope %dx%d, h %d..%d; %d overlay sprites\n",
			c.name, file, mlb.CourseW, h, co.NTiles, sj.W, sj.H, sj.Lo, sj.Hi, len(objects))
	}

	if err := writeJSON(filepath.Join(spritesDir, "index.json"), spriteIndex); err != nil {
		return err
	}

	return writeJSON(filepath.Join(outDir, "meta.json"), map[string]any{
		"format": 1, "game": "marble",
		"native": map[string]int{"w": 288, "h": 200},
		"tickHz": 50,
		"levels": levels,
	})
}

// buildSlope flattens a slope field into the dense grid the viewer meshes:
// each cell is the tile's height above lo, +1 (so 0 marks a pit / no surface).
func buildSlope(f slope.Field) slopeJSON {
	w, h := f.MaxX-f.MinX+1, f.MaxY-f.MinY+1
	heights := make([]int, w*h)
	for ty := f.MinY; ty <= f.MaxY; ty++ {
		for tx := f.MinX; tx <= f.MaxX; tx++ {
			if hv, ok := f.H[[2]int{tx, ty}]; ok && hv > 8000 {
				heights[(ty-f.MinY)*w+(tx-f.MinX)] = hv - f.Lo + 1
			}
		}
	}
	return slopeJSON{X0: f.MinX, Y0: f.MinY, W: w, H: h, Lo: f.Lo, Hi: f.Hi, Heights: heights}
}

// hashedRef returns "<name>?v=<fnv32 of the file>" — a cache-busting URL that
// changes whenever the file's content does.
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
