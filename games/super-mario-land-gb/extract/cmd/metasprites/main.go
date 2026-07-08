// metasprites renders Super Mario Land's object metasprites. The object draw at $25B7
// reads a frame id (slot+6) as an index into the pointer table $2FD9 (bank-0 fixed); the
// pointed-at stream is turtle graphics: a byte with bit7 clear moves the cursor (low
// nibble: bit3 up / bit2 down / bit1 left / bit0 right, 8 px each) and sets the OAM attr,
// a byte with bit7 set stamps an 8x16 OBJ sprite (SML runs sprites in 8x16 mode) of that
// tile at the cursor. We decode the stream from ROM and composite each frame with the
// world's OBJ tiles (pixels from a short oracle run, as the map renderer does).
//
//	go run ./cmd/metasprites [-rom PATH] [-world N] [-n FRAMES] [-o FILE]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/tools/platform/gameboy"
	"retroreverse.com/games/super-mario-land-gb/extract/level"
)

func main() {
	rom := flag.String("rom", "../Super Mario Land (World).gb", "ROM path")
	world := flag.Int("world", 1, "world (1-4) for the OBJ tile set")
	n := flag.Int("n", 128, "number of frames to render")
	out := flag.String("o", "../rendered/world1-metasprites.png", "output PNG")
	flag.Parse()

	data, err := os.ReadFile(*rom)
	ck(err)
	vram, obp0 := worldVRAM(data, byte(*world))

	const cell, cols = 32, 16
	rows := (*n + cols - 1) / cols
	img := image.NewRGBA(image.Rect(0, 0, cols*cell, rows*cell))
	pal := gameboy.GreyPalette()
	for f := 0; f < *n; f++ {
		ox := (f%cols)*cell + cell/2
		oy := (f/cols)*cell + 4
		for _, s := range level.DecodeMetasprite(data, f) {
			// 8x8 OBJ tile at the cursor (objects run in 8x8 mode).
			drawTile(img, vram, s.Tile, obp0, ox+s.DX, oy+s.DY, pal)
		}
	}
	f, err := os.Create(*out)
	ck(err)
	defer f.Close()
	ck(png.Encode(f, img))
	fmt.Println("wrote", *out)
}

func drawTile(img *image.RGBA, vram []byte, tile, obp0 byte, x, y int, pal color.Palette) {
	t := gameboy.DecodeTile(vram[int(tile)*16:])
	for py := 0; py < 8; py++ {
		for px := 0; px < 8; px++ {
			v := t[py][px]
			if v == 0 {
				continue // OBJ colour 0 = transparent
			}
			shade := (obp0 >> (2 * v)) & 3
			img.Set(x+px, y+py, pal[shade])
		}
	}
}

func worldVRAM(rom []byte, world byte) (vram []byte, obp0 byte) {
	m := gameboy.NewMachine(rom)
	m.RunFrames(80)
	for i := 0; i < 6; i++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0
	id := byte(world<<4 | 1)
	for i := 0; i < 40; i++ {
		m.Write(0xFFB4, id)
		m.RunFrame()
	}
	return m.VRAM(), m.Read(0xFF48)
}

func ck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "metasprites:", err)
		os.Exit(1)
	}
}
