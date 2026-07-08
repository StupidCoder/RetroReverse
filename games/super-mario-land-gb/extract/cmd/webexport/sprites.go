// sprites.go is the sprites stage of webexport: it composites each known object/enemy type's
// metasprite animation poses (from its script timeline, level.TypeTimeline; fallback: the one
// base metasprite) into a per-world object-icon strip PNG, plus Mario's idle sprite, and writes
// sprites/index.json keyed by "w<world>/<type>" (and "w<world>/mario"). The tile pixels come
// from a short oracle run per world (OBJ tiles + sprite palette); the frame ids and metasprite
// layouts are decoded from ROM (level.TypeFrames / level.DecodeMetasprite).
//
// (writeJSON, must are shared with main.go.)
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/super-mario-land-gb/extract/level"
	"retroreverse.com/tools/platform/gameboy"
)

// Object-icon atlas geometry: each known type's metasprite is composited into one objCell-square
// cell (stacked vertically). objOrigin is the cell pixel where the metasprite's cursor origin
// (0,0) sits, so the viewer can line an icon up with a placement. A 40x40 cell with the origin at
// (8,24) fits every base metasprite (tiles span DX -8..+32, DY -24..+16) without clipping.
const objCell = 40

var objOrigin = [2]int{8, 24}

// marioFrame is Mario's idle standing sprite: a 2x2 of OBJ tiles, each {tile, dx, dy} from the
// sprite top-left (read from the running game's OAM at spawn).
var marioFrame = []level.Sprite{
	{Tile: 0x00}, {Tile: 0x01, DX: 8}, {Tile: 0x10, DY: 8}, {Tile: 0x11, DX: 8, DY: 8},
}

// objIcon describes one type's strip in the object-icon atlas: its row, how many pose cells it
// has, and the per-pose durations (60 Hz frames) when it animates.
type objIcon struct {
	row       int
	frames    int
	durations []int
}

// exportSprites writes the sprites/ tree (index.json + worldN-obj.png) from the ROM, one
// icon atlas per world. No maps are read from the oracle — only VRAM pixels + the OBP0 palette.
func exportSprites(rom []byte, outdir string) {
	spritesDir := filepath.Join(outdir, "sprites")
	must(os.MkdirAll(spritesDir, 0o755))

	spriteIndex := map[string]any{}
	for world := 1; world <= 4; world++ {
		vram, obp0 := worldSpriteData(rom, byte(world))
		objAtlas := fmt.Sprintf("sprites/world%d-obj.png", world)
		objTypes, marioIcon := saveObjIcons(rom, filepath.Join(outdir, objAtlas), vram, obp0)

		icon := func(t objIcon) map[string]any {
			e := map[string]any{"src": objAtlas, "anchor": objOrigin}
			frames := make([][4]int, t.frames)
			for f := range frames {
				frames[f] = [4]int{f * objCell, t.row * objCell, objCell, objCell}
			}
			e["frames"] = frames
			if len(t.durations) > 1 {
				e["durations"] = t.durations // 60 Hz frames per pose (script-derived)
			}
			return e
		}
		for t, ic := range objTypes {
			spriteIndex[fmt.Sprintf("w%d/%s", world, t)] = icon(ic)
		}
		spriteIndex[fmt.Sprintf("w%d/mario", world)] = icon(objIcon{row: marioIcon, frames: 1})
		fmt.Fprintf(os.Stderr, "[sprites] world %d  %d object icons + mario -> %s\n", world, len(objTypes), objAtlas)
	}
	writeJSON(filepath.Join(spritesDir, "index.json"), spriteIndex)
	fmt.Fprintf(os.Stderr, "[sprites] done: %d entries + index.json\n", len(spriteIndex))
}

// saveObjIcons composites each known object type's animation poses (from its script timeline,
// level.TypeTimeline; fallback: the single base metasprite) into a strip of objCell cells, one
// row per type, in this world's OBJ tiles + sprite palette, plus one extra row for Mario's idle
// sprite. It returns the type id -> icon map and Mario's row. Types not in level.TypeFrames are
// omitted (the viewer falls back to a marker for them).
func saveObjIcons(rom []byte, path string, vram []byte, obp0 byte) (map[string]objIcon, int) {
	// Stable order so the atlas is deterministic.
	typeFrame := level.TypeFrames(rom)
	var types []int
	for t := range typeFrame {
		types = append(types, int(t))
	}
	sort.Ints(types)

	out := map[string]objIcon{}
	maxFrames := 1
	for i, t := range types {
		ic := objIcon{row: i, frames: 1}
		if tl := level.TypeTimeline(rom, byte(t)); tl != nil {
			ic.frames = len(tl)
			for _, st := range tl {
				ic.durations = append(ic.durations, st.Frames)
			}
		}
		if ic.frames > maxFrames {
			maxFrames = ic.frames
		}
		out[fmt.Sprintf("%d", t)] = ic
	}

	marioIdx := len(types) // Mario gets the last row
	img := image.NewNRGBA(image.Rect(0, 0, objCell*maxFrames, objCell*(len(types)+1)))
	stamp := func(s level.Sprite, ox, oy int) {
		tl := gameboy.DecodeTile(vram[int(s.Tile)*16:]) // 8x8 OBJ tile
		for py := 0; py < 8; py++ {
			for px := 0; px < 8; px++ {
				sx, sy := px, py // the sprite's flip attrs mirror the tile
				if s.XFlip {
					sx = 7 - px
				}
				if s.YFlip {
					sy = 7 - py
				}
				if v := tl[sy][sx]; v != 0 { // OBJ colour 0 = transparent
					g := []uint8{0xff, 0xaa, 0x55, 0x00}[(obp0>>(2*v))&3]
					img.Set(ox+s.DX+px, oy+s.DY+py, color.NRGBA{g, g, g, 0xff})
				}
			}
		}
	}
	pose := func(frame byte, cellX, cellY int) {
		for _, s := range level.DecodeMetasprite(rom, int(frame)) {
			stamp(s, cellX*objCell+objOrigin[0], cellY*objCell+objOrigin[1])
		}
	}
	for i, t := range types {
		if tl := level.TypeTimeline(rom, byte(t)); tl != nil {
			for f, st := range tl {
				pose(st.Frame, f, i)
			}
		} else {
			pose(typeFrame[byte(t)], 0, i)
		}
	}
	for _, s := range marioFrame { // Mario's 2x2, top-left at the cell origin
		stamp(s, objOrigin[0], marioIdx*objCell+objOrigin[1])
	}
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	must(png.Encode(f, img))
	return out, marioIdx
}

// worldSpriteData warps the oracle into `world` and returns its VRAM (OBJ tiles) and OBP0
// sprite palette register — the pixels the icon compositor needs.
func worldSpriteData(rom []byte, world byte) (vram []byte, obp0 byte) {
	m := gameboy.NewMachine(rom)
	m.RunFrames(80)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0
	id := byte(world<<4 | 1)
	for f := 0; f < 40; f++ {
		m.Write(0xFFB4, id)
		m.RunFrame()
	}
	return m.VRAM(), m.Read(0xFF48)
}
