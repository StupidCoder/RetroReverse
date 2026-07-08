// levels.go is the levels stage: each course becomes a format-2 tilemap2d level
// (Marble_Madness.md Part IV §3/§6). The colour bands are baked into the atlas as
// recoloured variant tiles and the gold shimmer as tileAnims, so the viewer needs
// no palette machinery. The scenery-overlay sprite pieces the level references are
// produced by the sprites stage; here we only emit their object placements (keys
// resolve via sprites/index.json). NO slope link — the slopes are standalone views.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/marble-madness-amiga/extract/mlb"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/c64/gfx"
)

// shimmerPeriod is the vblank frames each gold-rotation phase holds: the stepper
// $84DC runs per WORLD frame (2 vblanks each), stepping the phase past 6 — every
// 4th world frame = 8 vblanks.
const shimmerPeriod = 8

// exportLevels writes the levels/ tree (per-course format-2 tilemaps, object DBs and
// deduped-into-one atlas per course) and returns the manifest level index.
func exportLevels(vol *adf.Volume, paths map[string]string, outDir string) []LevelIndex {
	levelsDir := filepath.Join(outDir, "levels")
	chk(os.MkdirAll(levelsDir, 0o755))

	var levels []LevelIndex
	for idx, c := range courses {
		cr, err := loadCourse(vol, paths, c.key, c.track)
		chk(err)
		co, bake := cr.co, cr.bake

		// The scenery overlays give us the level's object placements + the screen-swap
		// tile animations. render=false: the sprite PNGs/index come from the sprites stage.
		objects, cellAnims, err := exportOverlays(vol, paths, c.key, cr.prog.Image, co, bake.paletteAt, "", nil, false)
		chk(err)

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
		// swap-variant rows repaint into the finale band: substitute per destination
		// row, and count their tiles toward that band's shimmer.
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

		// The tilemap's cells reference variant tiles appended to the atlas, so the level
		// JSON and the PNG must never mix versions in a client cache: hash the atlas URL.
		atlasFile := c.key + ".atlas.png"
		chk(gfx.WritePNG(filepath.Join(levelsDir, atlasFile), co.AtlasVariants(16, bake.ext, bake.varList)))
		hashed, err := hashedRef(filepath.Join(levelsDir, atlasFile), atlasFile)
		chk(err)
		atlasRef := "levels/" + hashed // path relative to the asset root, with the cache-buster

		// The map shows the PLAYABLE rows (the .mlb header count = the engine's scroll
		// clamp). Data rows beyond it are off-screen variant storage; the swap animation
		// replays them from the full cell array.
		level := map[string]any{
			"format": 2,
			"name":   c.name,
			"kind":   "tilemap2d",
			"extents": map[string]any{
				"tileSize": 8, "width": mlb.CourseW, "height": co.PlayableH,
			},
			"wrap": "none",
			"grid": map[string]any{
				"tileSize": 8, "atlas": atlasRef, "atlasCols": 16, "atlasGutter": 1,
				"width": mlb.CourseW, "height": co.PlayableH,
				"cells": cells,
			},
			"objects": objects,
			// Frame the Amiga's on-screen view (288x200 playfield) at the course top.
			"view":        map[string]any{"x": (mlb.CourseW*8 - 288) / 2, "y": 0, "w": 288, "h": 200},
			"objectsFile": c.key + ".objects.json",
		}
		if len(cellAnims) > 0 {
			level["cellAnims"] = cellAnims
		}
		if len(tileAnims) > 0 {
			level["tileAnims"] = tileAnims
		}
		chk(writeJSON(filepath.Join(levelsDir, c.key+".json"), level))

		// machine-readable object DB (sibling of the level file)
		db := objectsDB(c.name, objects)
		chk(writeJSON(filepath.Join(levelsDir, c.key+".objects.json"), db))

		levels = append(levels, LevelIndex{
			Name: c.name, Section: "Courses", File: "levels/" + c.key + ".json",
			Kind: "tilemap2d", Atlas: "levels/" + atlasFile, Objects: "levels/" + c.key + ".objects.json",
		})
		fmt.Fprintf(os.Stderr, "[levels] %d/%d  %-12s %s  %d×%d tiles, %d tiles, %d objects\n",
			idx+1, len(courses), c.name, c.key+".json", mlb.CourseW, co.PlayableH, co.NTiles, len(objects))
	}
	fmt.Fprintf(os.Stderr, "[levels] done: %d courses\n", len(levels))
	return levels
}

// objectsDB builds the machine-readable object database from the level's inline overlay
// objects ({type,name,x,y,sprite} maps produced by exportOverlays).
func objectsDB(level string, objects []map[string]any) map[string]any {
	out := make([]map[string]any, 0, len(objects))
	for i, o := range objects {
		e := map[string]any{
			"id":  i,
			"pos": []int{o["x"].(int), o["y"].(int)},
		}
		if t, ok := o["type"]; ok {
			e["type"] = t
		}
		if n, ok := o["name"]; ok {
			e["name"] = n
		}
		if s, ok := o["sprite"]; ok {
			e["props"] = map[string]any{"sprite": s}
		}
		out = append(out, e)
	}
	return map[string]any{"format": 2, "level": level, "objects": out}
}

func chk(e error) {
	if e != nil {
		fail(e)
	}
}
