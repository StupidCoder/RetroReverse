// spritesheet renders Super Mario Land's object/enemy sprite tiles. The object tile
// art is bulk-loaded per world into VRAM $8A00+ (from the $0DE4 source table in the
// world's data bank; World 1 via $0D30/$05D0). We warp the oracle into each world and
// read back the $8000-$8FFF OBJ tile block, decoding it with the sprite palette OBP0 —
// the oracle supplies pixels only, exactly as the level-map renderer does.
//
//	go run ./cmd/spritesheet [-rom PATH] [-o DIR]
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"stupidcoder.com/tools/gameboy"
)

const (
	tile = 8
	cols = 16  // 16 tiles wide
	nTL  = 256 // $8000-$8FFF = 256 OBJ tiles
)

func main() {
	rom := arg("-rom", "../Super Mario Land (World).gb")
	out := arg("-o", "../rendered")
	data, err := os.ReadFile(rom)
	ck(err)
	ck(os.MkdirAll(out, 0o755))

	for world := 1; world <= 4; world++ {
		vram, obp0 := worldVRAM(data, byte(world))
		img := image.NewPaletted(image.Rect(0, 0, cols*tile, (nTL/cols)*tile), gameboy.GreyPalette())
		shade := func(v uint8) uint8 { return (obp0 >> (2 * v)) & 3 }
		for n := 0; n < nTL; n++ {
			t := gameboy.DecodeTile(vram[n*16:]) // OBJ tiles are always unsigned $8000
			ox, oy := (n%cols)*tile, (n/cols)*tile
			for y := 0; y < tile; y++ {
				for x := 0; x < tile; x++ {
					img.SetColorIndex(ox+x, oy+y, shade(t[y][x]))
				}
			}
		}
		path := filepath.Join(out, fmt.Sprintf("world%d-sprites.png", world))
		f, err := os.Create(path)
		ck(err)
		ck(png.Encode(f, img))
		f.Close()
		fmt.Println("wrote", path)
	}
}

// worldVRAM warps the oracle into `world` and returns VRAM + the OBP0 sprite palette.
func worldVRAM(rom []byte, world byte) (vram []byte, obp0 byte) {
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

func arg(name, def string) string {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return def
}

func ck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "spritesheet:", err)
		os.Exit(1)
	}
}
