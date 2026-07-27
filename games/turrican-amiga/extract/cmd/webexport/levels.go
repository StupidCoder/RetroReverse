// levels.go is the levels stage: for each world it writes the tile atlas into
// levels/, and each scene becomes a Retro-X tilemap level — grid with the
// hflip flag bit, the per-tile 4x4 solidity collision, the engine's spawn and
// camera view, and every placement resolved to its objects-stage asset at the
// frame its AI handler selected for its orientation.
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/games/turrican-amiga/extract/scene"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

const (
	blockBase  = 0x1B980
	tileSide   = 32
	tileBytes  = tileSide * 4 * (tileSide / 8) // 512
	atlasCols  = 16
	tileGutter = 1                       // extruded border per tile (kills atlas bleed at fractional zoom)
	atlasCell  = tileSide + 2*tileGutter // 34
)

// hflipMask re-encodes the map's flip convention (raw byte >= ntiles = tile-128, h-flipped)
// as an explicit cell flag bit per RETROX.md.
const hflipMask = 0x8000

// exportLevels writes the level documents + per-world atlases. refs may be nil
// (objects stage disabled): then the scenes ship without placements.
func exportLevels(ctx *cli.Context, game *scene.Game, scenes [][]scene.Scene, refs map[string]objRef) error {
	b := ctx.Builder
	n := 0
	for w := 0; w < scene.NumWorlds; w++ {
		block := game.Block(w).Data
		be := func(o int) int { return be32(block, o) }
		at := func(addr int) int { return addr - blockBase }

		pal := readPalette(block, at(be(0x08)))
		tableOff := at(be(0x00))
		nTiles := be32(block, tableOff) / 4

		atlasName := fmt.Sprintf("atlas%d.png", w)
		p, err := b.Path("levels", atlasName)
		if err != nil {
			return err
		}
		if err := writeAtlas(p, block, tableOff, nTiles, pal); err != nil {
			return err
		}

		// Per-tile collision (16 bytes/tile = 4x4 of 8x8-block solidity).
		collBytes, _ := game.TileCollision(w)
		collision := make([]int, len(collBytes))
		for i, cb := range collBytes {
			collision[i] = int(cb)
		}

		for _, sc := range scenes[w] {
			descOff := at(be(0x16 + sc.Index*4))
			mapOff := at(be(descOff + 0x00))
			if sc.Width <= 0 || sc.Height <= 0 || mapOff+sc.Width*sc.Height > len(block) {
				return fmt.Errorf("world %d scene %d: bad map %dx%d", w, sc.Index, sc.Width, sc.Height)
			}
			cells := make([]int, sc.Width*sc.Height)
			for col := 0; col < sc.Width; col++ {
				for row := 0; row < sc.Height; row++ {
					v := int(block[mapOff+col*sc.Height+row]) // col-major -> row-major
					if v >= nTiles {                          // raw flip convention -> explicit flag bit
						v = (v - 128) | hflipMask
					}
					cells[row*sc.Width+col] = v
				}
			}

			doc := &schema.Level{
				Type: schema.LevelTilemap,
				Tilemap: &schema.Tilemap{
					TileSize: tileSide, Width: sc.Width, Height: sc.Height,
					Atlas:     schema.TileAtlas{File: atlasName, Cols: atlasCols, Gutter: tileGutter},
					Cells:     cells,
					HFlipMask: hflipMask,
					View:      &schema.Rect{X: sc.CamX, Y: sc.CamY, W: amigaViewW, H: amigaViewH},
					Spawn:     &schema.Spawn{X: sc.SpawnX, Y: sc.SpawnY},
				},
				Collision: &schema.Collision{
					Kind: "grid", Sub: 4, Solid: collision,
					Legend: map[string]string{"1": "#ff3030", "127": "#33ddff", "128": "#ffe020", "211": "#ff33cc"},
				},
			}

			for i, o := range sc.Objects {
				if o.FT == 0 {
					continue
				}
				key := fmt.Sprintf("w%d/%05X", w, o.FT)
				if o.Resident {
					key = fmt.Sprintf("resident/%05X", o.FT)
				}
				ref, ok := refs[key]
				if !ok {
					continue // no sprite resolved (unhandled type)
				}
				pl := schema.Placement{
					ID:     i,
					Object: ref.asset,
					Anim:   fmt.Sprintf("f%d", o.Frame),
					Pos:    []float64{float64(o.X), float64(o.Y)},
					Props: map[string]any{
						"type":   o.Type,
						"orient": o.Orient,
						"frame":  o.Frame,
					},
				}
				if o.Handler != 0 {
					pl.Props["handler"] = fmt.Sprintf("$%X", o.Handler)
				}
				doc.Placements = append(doc.Placements, pl)
			}

			b.AddLevel(schema.Asset{
				ID:    fmt.Sprintf("world%d-scene%d", w+1, sc.Index+1),
				Name:  fmt.Sprintf("Scene %d", sc.Index+1),
				Group: fmt.Sprintf("World %d", w+1),
			}, doc)
			n++
			ctx.Progress("levels", n, 13, fmt.Sprintf("world %d scene %d: %dx%d, %d objects",
				w+1, sc.Index+1, sc.Width, sc.Height, len(doc.Placements)))
		}
	}
	return nil
}

// writeAtlas packs the world's 32x32 tiles into a grid PNG, each tile in a
// (tileSide+2*tileGutter)-pixel cell whose 1-pixel border duplicates the tile's edge pixels
// (extrusion guard against atlas bleed at fractional zoom).
func writeAtlas(path string, block []byte, tableOff, nTiles int, pal color.Palette) error {
	rows := (nTiles + atlasCols - 1) / atlasCols
	img := image.NewPaletted(image.Rect(0, 0, atlasCols*atlasCell, rows*atlasCell), pal)
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v >= tileSide {
			return tileSide - 1
		}
		return v
	}
	for n := 0; n < nTiles; n++ {
		off := tableOff + int(binary.BigEndian.Uint32(block[tableOff+n*4:]))
		if off+tileBytes > len(block) {
			break
		}
		var t [tileSide][tileSide]uint8
		for y := 0; y < tileSide; y++ {
			var planes [4]uint32
			for p := 0; p < 4; p++ {
				planes[p] = binary.BigEndian.Uint32(block[off+(y*4+p)*4:])
			}
			for x := 0; x < tileSide; x++ {
				var v uint8
				for p := 0; p < 4; p++ {
					v |= uint8((planes[p]>>(31-uint(x)))&1) << uint(p)
				}
				t[y][x] = v
			}
		}
		ox, oy := (n%atlasCols)*atlasCell, (n/atlasCols)*atlasCell
		for y := -tileGutter; y < tileSide+tileGutter; y++ {
			for x := -tileGutter; x < tileSide+tileGutter; x++ {
				img.SetColorIndex(ox+tileGutter+x, oy+tileGutter+y, t[clamp(y)][clamp(x)])
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// readPalette reads a 16-colour Amiga playfield palette (fully opaque, for the tile atlas).
func readPalette(block []byte, off int) color.Palette {
	pal := make(color.Palette, 16)
	for i := range pal {
		c := binary.BigEndian.Uint16(block[off+i*2:])
		pal[i] = color.RGBA{R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255}
	}
	return pal
}
