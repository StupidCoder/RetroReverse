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
		atlasFile := c.key + ".atlas.png"
		if err := gfx.WritePNG(filepath.Join(outDir, atlasFile), co.Atlas(16)); err != nil {
			return err
		}
		file := c.key + ".json"
		if err := writeJSON(filepath.Join(outDir, file), map[string]any{
			"format": 1,
			"name":   c.name,
			"grid": map[string]any{
				"tileSize": 8, "atlas": atlasFile, "atlasCols": 16, "atlasGutter": 1,
				"width": mlb.CourseW, "height": co.H, "cells": co.Cells,
			},
			// Frame the Amiga's on-screen view (288x200 playfield) at the course top.
			"view": map[string]any{"x": (mlb.CourseW*8 - 288) / 2, "y": 0, "w": 288, "h": 200},
		}); err != nil {
			return err
		}
		h := co.H

		// Slope field (the 3-D rolling surface) from the course Track file.
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
		sj := buildSlope(slope.Build(prog.Image))
		sj.Markers = buildMarkers(slope.Markers(prog.Image))
		slopeFile := c.key + ".slope.json"
		if err := writeJSON(filepath.Join(outDir, slopeFile), sj); err != nil {
			return err
		}

		levels = append(levels, metaLevel{Name: c.name, File: file, Atlas: atlasFile, Slope: slopeFile})
		fmt.Printf("%-12s %s  %d×%d tiles, %d tiles; slope %dx%d, h %d..%d\n",
			c.name, file, mlb.CourseW, h, co.NTiles, sj.W, sj.H, sj.Lo, sj.Hi)
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

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
