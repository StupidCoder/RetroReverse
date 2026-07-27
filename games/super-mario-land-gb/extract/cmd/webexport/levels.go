// levels.go is the levels stage: it decodes every Super Mario Land level map
// from the cartridge into a Retro-X tilemap level — a per-world 256-tile
// background atlas PNG, the grid, Mario's spawn, per-level tile animations,
// the tile-solidity collision grid, and every object placement (resolved to
// the objects stage's assets). The maps come straight from the ROM
// (extract/level); the tile graphics come from a short oracle run nudged into
// each world (pixels only, not the map).
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"retroreverse.com/games/super-mario-land-gb/extract/level"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/gameboy"
)

// isSolid is the tile-id solidity rule (the walkable/blocking BG tiles occupy
// one id range in every world's tile set) — exported as grid collision.
func isSolid(id int) bool { return id >= 0x60 && id < 0xF0 }

// levelMusicTrack maps a level (world, lv) to the music asset it plays.
// Several levels share a theme (Super_Mario_Land.md Part VI): 1-1/1-2/3-1
// share the overworld theme, 1-3/3-2/3-3 the cave theme, etc.
func levelMusicTrack(world, lv int) string {
	switch [2]int{world, lv} {
	case [2]int{1, 1}, [2]int{1, 2}, [2]int{3, 1}:
		return "theme-1-1"
	case [2]int{1, 3}, [2]int{3, 2}, [2]int{3, 3}:
		return "theme-1-3"
	case [2]int{2, 1}, [2]int{2, 2}:
		return "theme-2-1"
	case [2]int{4, 1}, [2]int{4, 2}:
		return "theme-4-1"
	case [2]int{2, 3}, [2]int{4, 3}:
		return "theme-2-3"
	}
	return ""
}

// exportLevels writes the level documents + per-world atlases. refs may be nil
// (objects stage disabled): then the levels ship without placements.
func exportLevels(ctx *cli.Context, rom []byte, refs map[string]objRef) error {
	b := ctx.Builder

	// per-tile solidity: one byte per tile id
	solid := make([]int, nTile)
	for t := range solid {
		if isSolid(t) {
			solid[t] = 1
		}
	}

	n := 0
	for world := 1; world <= 4; world++ {
		// Tiles: warp the oracle into this world and read back VRAM + palettes.
		vram, lcdc, bgp := worldTiles(rom, byte(world))
		// The world's animated-tile frames ($23F8 rewrites tile $5D's high
		// plane), decoded from ROM per-world; the atlas paints tile $5D and
		// the appended frame from the decode.
		anim := level.DecodeTileAnim(rom, byte(world<<4|3))
		for lv := 1; anim == nil && lv <= 3; lv++ {
			anim = level.DecodeTileAnim(rom, byte(world<<4|lv))
		}
		atlas := fmt.Sprintf("world%d.png", world)
		p, err := b.Path("levels", atlas)
		if err != nil {
			return err
		}
		if err := saveAtlas(p, vram, lcdc, bgp, anim); err != nil {
			return err
		}

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

			name := fmt.Sprintf("%d-%d", world, lv)
			doc := &schema.Level{
				Type:  schema.LevelTilemap,
				Music: levelMusicTrack(world, lv),
				Tilemap: &schema.Tilemap{
					TileSize: tile, Width: w, Height: 16,
					Atlas: schema.TileAtlas{File: atlas, Cols: cols, Gutter: gut},
					Cells: cells,
					// Frame the first Game Boy screen (the map is 128px tall
					// inside the 144px screen; the 16px HUD sits above it).
					View: &schema.Rect{X: 0, Y: -8, W: 160, H: 144},
					// Mario's fixed start (his sprite top-left, in map px):
					// the engine spawns him at screen (50,134); minus the HUD.
					Spawn: spawnFor(world, refs),
				},
				Collision: &schema.Collision{Kind: "grid", Sub: 1, Solid: solid},
			}
			// Tile animation ($23F8): in enabled levels tile $5D alternates
			// between the accent frame (appended as atlas tile 256) and its
			// resting shape, 8 frames per phase.
			if level.DecodeTileAnim(rom, id) != nil {
				doc.Tilemap.TileAnims = []schema.TileAnim{{
					Tiles:        []int{level.AnimTile},
					Frames:       [][]int{{nTile}, {level.AnimTile}},
					PeriodFrames: level.AnimPeriod,
				}}
			}

			// Object/enemy placements (decoded from the ROM list), world px:
			// tile origin plus the position byte's 4px-per-unit fine X nudge
			// (oracle-verified by cmd/spawnverify).
			objs := level.DecodeObjectsByID(rom, id)
			for i, o := range objs {
				ref, ok := refs[fmt.Sprintf("w%d/%d", world, o.Type)]
				if !ok {
					ref, ok = refs[fmt.Sprintf("u/%d", o.Type)]
				}
				if !ok {
					continue // objects stage disabled
				}
				doc.Placements = append(doc.Placements, schema.Placement{
					ID:     i,
					Object: ref.asset,
					Pos:    []float64{float64(o.Col*8 + o.FineX*4), float64(o.Row * 8)},
					Hard:   o.Hard,
					Props:  map[string]any{"type": fmt.Sprintf("0x%02X", o.Type)},
				})
			}

			b.AddLevel(schema.Asset{
				ID: "level-" + name, Name: name, Group: fmt.Sprintf("World %d", world),
			}, doc)
			n++
			ctx.Progress("levels", n, 12, fmt.Sprintf("%-4s %3d cols  %s  %d objects", name, w, atlas, len(objs)))
		}
	}
	return nil
}

func spawnFor(world int, refs map[string]objRef) *schema.Spawn {
	sp := &schema.Spawn{X: 35, Y: 96}
	if ref, ok := refs[fmt.Sprintf("w%d/mario", world)]; ok {
		sp.Object = ref.asset
	}
	return sp
}

// worldTiles boots the ROM, nudges it into `world` by forcing the level id
// through the load window, and returns the VRAM, LCDC and BG palette register.
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

// saveAtlas writes the 256 background tiles (indexed by the map's tile value,
// signed $8800 addressing) into a 16-wide atlas with a 1px extruded gutter, in
// the BG palette. When the world animates tile $5D (anim != nil), that cell
// and the appended frame cell (atlas tile 256, a 17th row) are painted from
// the ROM-decoded frames.
func saveAtlas(path string, vram []byte, lcdc, bgp byte, anim *level.TileAnim) error {
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
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
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

// tileOffset mirrors the BG tile addressing (LCDC bit 4: $8000 unsigned vs
// signed $8800).
func tileOffset(lcdc, idx byte) int {
	if lcdc&0x10 != 0 {
		return int(idx) * 16
	}
	return 0x1000 + int(int8(idx))*16
}
