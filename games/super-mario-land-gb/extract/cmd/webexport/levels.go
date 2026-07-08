// levels.go is the levels stage of webexport: it decodes every Super Mario Land level map
// from the cartridge and writes the format-2 level tree — a per-world 256-tile background
// atlas PNG, a per-level tilemap2d envelope (grid + objects + collision), and a sibling
// machine-readable object DB. The maps come straight from the ROM (extract/level); the tile
// graphics come from a short oracle run nudged into each world (pixels only, not the map).
//
// (writeJSON, must, and the tile/atlas consts are shared with main.go.)
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/games/super-mario-land-gb/extract/level"
	"retroreverse.com/tools/platform/gameboy"
)

// isSolid is the tile-id solidity rule (the walkable/blocking BG tiles occupy one id range in
// every world's tile set) — exported as collision data (kind "grid", sub 1).
func isSolid(id int) bool { return id >= 0x60 && id < 0xF0 }

// levelMusicTrack maps a level (world, lv) to the music-stage track stem it plays. Several
// levels share a theme (Super_Mario_Land.md Part VI / cmd/musicrom's track table): 1-1/1-2/3-1
// share the overworld theme, 1-3/3-2/3-3 the cave theme, etc. The stems match the files the
// music stage renders (music/<stem>.mp3), so a level references its track even when -only
// gates the music stage out.
func levelMusicTrack(world, lv int) string {
	switch [2]int{world, lv} {
	case [2]int{1, 1}, [2]int{1, 2}, [2]int{3, 1}:
		return "level-1-1"
	case [2]int{1, 3}, [2]int{3, 2}, [2]int{3, 3}:
		return "level-1-3"
	case [2]int{2, 1}, [2]int{2, 2}:
		return "level-2-1"
	case [2]int{4, 1}, [2]int{4, 2}:
		return "level-4-1"
	case [2]int{2, 3}, [2]int{4, 3}:
		return "level-2-3"
	}
	return ""
}

// exportLevels writes the levels/ tree (per-world atlas, per-level tilemap2d + object DB) and
// returns the manifest level index. This is the original webexport logic wrapped in the
// format-2 envelope; it is one of the two stages that use the machine oracle (VRAM pixels).
func exportLevels(rom []byte, outdir string) []LevelIndex {
	levelsDir := filepath.Join(outdir, "levels")
	must(os.MkdirAll(levelsDir, 0o755))

	// per-tile solidity: one byte per tile id (kind "grid", sub 1)
	solid := make([]int, nTile)
	for t := range solid {
		if isSolid(t) {
			solid[t] = 1
		}
	}
	// which object types have a decoded metasprite icon (world-independent). Objects of a
	// known type get a "w<world>/<type>" sprite ref; the sprites stage renders those PNGs.
	known := level.TypeFrames(rom)

	var levels []LevelIndex
	for world := 1; world <= 4; world++ {
		// Tiles: warp the oracle into this world and read back VRAM + palettes.
		vram, lcdc, bgp := worldTiles(rom, byte(world))
		// The world's animated-tile frames ($23F8 rewrites tile $5D's high plane), decoded
		// from ROM per-world; the atlas paints tile $5D and the appended frame from the decode.
		anim := level.DecodeTileAnim(rom, byte(world<<4|3))
		for lv := 1; anim == nil && lv <= 3; lv++ {
			anim = level.DecodeTileAnim(rom, byte(world<<4|lv))
		}
		atlas := fmt.Sprintf("world%d.png", world)
		saveAtlas(filepath.Join(levelsDir, atlas), vram, lcdc, bgp, anim)
		atlasRef := "levels/" + atlas // path relative to the asset root

		for lv := 1; lv <= 3; lv++ {
			id := byte(world<<4 | lv)
			cols2, _, _ := level.DecodeLevelByID(rom, id)
			w := len(cols2)
			cells := make([]int, w*16) // row-major: cells[r*w + x]
			for x, col := range cols2 {
				for r := 0; r < 16; r++ {
					cells[r*w+x] = int(col[r])
				}
			}
			// Object/enemy placements (decoded from the ROM list), world px: tile origin plus
			// the position byte's 4px-per-unit fine X nudge (oracle-verified by cmd/spawnverify).
			objs := level.DecodeObjectsByID(rom, id)
			inline := make([]map[string]any, len(objs))
			for i, o := range objs {
				e := map[string]any{
					"type": o.Type, "x": o.Col*8 + o.FineX*4, "y": o.Row * 8, "hard": o.Hard,
				}
				if _, ok := known[o.Type]; ok {
					e["sprite"] = fmt.Sprintf("w%d/%d", world, o.Type)
				}
				inline[i] = e
			}

			name := fmt.Sprintf("%d-%d", world, lv)
			stem := "level-" + name
			grid := map[string]any{
				"tileSize": tile, "atlas": atlasRef, "atlasCols": cols, "atlasGutter": gut,
				"width": w, "height": 16, "cells": cells,
			}
			lj := map[string]any{
				"format": 2,
				"name":   name,
				"kind":   "tilemap2d",
				// format-2 extent: size in CELLS (mirror of grid.width/height)
				"extents": map[string]any{"tileSize": tile, "width": w, "height": 16},
				"wrap":    "none",
				"grid":    grid,
				// Frame the first Game Boy screen (the map is 128px tall inside the 144px
				// screen; the 16px HUD sits above the map).
				"view":    map[string]any{"x": 0, "y": -8, "w": 160, "h": 144},
				"objects": inline,
				// Mario's fixed start (his sprite top-left, in map pixels): the engine spawns
				// him at screen (50,134); minus the 16px HUD that is map px.
				"spawn":       map[string]any{"x": 35, "y": 96, "sprite": fmt.Sprintf("w%d/mario", world)},
				"objectsFile": stem + ".objects.json",
				"collision":   map[string]any{"kind": "grid", "sub": 1, "solid": solid},
			}
			if track := levelMusicTrack(world, lv); track != "" {
				lj["music"] = track
			}
			// Tile animation ($23F8): in enabled levels tile $5D alternates between the accent
			// frame (appended as atlas tile 256) and its resting shape, 8 frames per phase.
			if level.DecodeTileAnim(rom, id) != nil {
				lj["tileAnims"] = []map[string]any{{
					"tiles": []int{level.AnimTile}, "frames": [][]int{{nTile}, {level.AnimTile}},
					"periodFrames": level.AnimPeriod,
				}}
			}
			writeJSON(filepath.Join(levelsDir, stem+".json"), lj)

			// machine-readable object DB (sibling of the level file)
			db := map[string]any{"format": 2, "level": name}
			dbObjs := make([]map[string]any, len(objs))
			for i, o := range objs {
				props := map[string]any{"hard": o.Hard}
				if _, ok := known[o.Type]; ok {
					props["sprite"] = fmt.Sprintf("w%d/%d", world, o.Type)
				}
				dbObjs[i] = map[string]any{
					"id": i, "type": int(o.Type),
					"pos": []int{o.Col*8 + o.FineX*4, o.Row * 8}, "props": props,
				}
			}
			db["objects"] = dbObjs
			writeJSON(filepath.Join(levelsDir, stem+".objects.json"), db)

			levels = append(levels, LevelIndex{
				Name: name, Section: fmt.Sprintf("World %d", world),
				File: "levels/" + stem + ".json", Kind: "tilemap2d",
				Atlas: atlasRef, Objects: "levels/" + stem + ".objects.json",
			})
			fmt.Fprintf(os.Stderr, "[levels] %-4s  %3d cols  %s  %d objects\n", name, w, atlas, len(objs))
		}
	}
	fmt.Fprintf(os.Stderr, "[levels] done: %d levels, 4 world atlases\n", len(levels))
	return levels
}

// worldTiles boots the ROM, nudges it into `world` by forcing the level id through the load
// window, and returns the VRAM, LCDC and BG palette register.
func worldTiles(rom []byte, world byte) (vram []byte, lcdc, bgp byte) {
	m := gameboy.NewMachine(rom)
	m.RunFrames(80)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0
	id := byte(world<<4 | 1) // world's first level
	for f := 0; f < 40; f++ {
		m.Write(0xFFB4, id)
		m.RunFrame()
	}
	return m.VRAM(), m.Read(0xFF40), m.Read(0xFF47)
}

// saveAtlas writes the 256 background tiles (indexed by the map's tile value, signed $8800
// addressing) into a 16-wide atlas with a 1px extruded gutter, in the BG palette. When the
// world animates tile $5D (anim != nil), that cell and the appended frame cell (atlas tile
// 256, a 17th row) are painted from the ROM-decoded frames.
func saveAtlas(path string, vram []byte, lcdc, bgp byte, anim *level.TileAnim) {
	total := nTile
	if anim != nil {
		total = nTile + 1
	}
	rows := (total + cols - 1) / cols
	img := image.NewPaletted(image.Rect(0, 0, cols*cell, rows*cell), gameboy.GreyPalette())
	shade := func(v uint8) uint8 { return (bgp >> (2 * v)) & 3 }
	paint := func(n int, t [8][8]uint8) {
		ox, oy := (n%cols)*cell+gut, (n/cols)*cell+gut
		for y := -gut; y < tile+gut; y++ {
			for x := -gut; x < tile+gut; x++ {
				sx, sy := clamp(x), clamp(y) // extrude edges into the gutter
				img.SetColorIndex(ox+x, oy+y, shade(t[sy][sx]))
			}
		}
	}
	for n := 0; n < nTile; n++ {
		paint(n, gameboy.DecodeTile(vram[tileOffset(lcdc, byte(n)):]))
	}
	if anim != nil {
		paint(level.AnimTile, gameboy.DecodeTile(anim.Frames[1][:])) // resting shape
		paint(nTile, gameboy.DecodeTile(anim.Frames[0][:]))          // accent frame
	}
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	must(png.Encode(f, img))
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v >= tile {
		return tile - 1
	}
	return v
}

// tileOffset mirrors the BG tile addressing (LCDC bit 4: $8000 unsigned vs signed $8800).
func tileOffset(lcdc, idx byte) int {
	if lcdc&0x10 != 0 {
		return int(idx) * 16
	}
	return 0x1000 + int(int8(idx))*16
}
