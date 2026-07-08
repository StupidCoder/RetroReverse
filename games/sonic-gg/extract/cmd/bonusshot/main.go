// bonusshot renders the shared zone-6 (bonus/special-stage) map with its tiles+palette,
// and prints each bonus descriptor's vertical bounds + spawn, to see whether the 8 bonus
// stages (descriptor idx 28-35) are distinct maps or vertical windows of one map.
//
// Usage: bonusshot -in <rom.gg> -o <out.png>
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/games/sonic-gg/extract/decomp"
	"retroreverse.com/tools/platform/gamegear"
)

func w(rom []byte, o int) int { return int(rom[o]) | int(rom[o+1])<<8 }

func main() {
	in := flag.String("in", "", "input Game Gear ROM (.gg)")
	out := flag.String("o", "", "output PNG")
	flag.Parse()
	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: bonusshot -in <rom.gg> -o <out.png>")
		os.Exit(2)
	}
	rom, _ := os.ReadFile(*in)
	const T = 0x15600
	d := T + w(rom, T+28*2) // idx 28 = first bonus
	stride := w(rom, d+1)
	tileFile := 0x30000 + w(rom, d+21)
	mapFile := 0x14000 + w(rom, d+15)
	mapLen := w(rom, d+17)
	blkTable := 0x10000 + w(rom, d+19)
	palIdx := int(rom[d+29])
	off := w(rom, 0x23400+palIdx*2)
	pal := gamegear.Palette(rom[0x23400+off : 0x23400+off+32])

	tiles := decomp.Decompress(rom, tileFile)
	mp := decomp.LoadMapRLE(rom, mapFile, mapLen)
	rows := 4096 / stride
	cols := stride
	fmt.Printf("map %dx%d blocks, %d tiles, pal idx %d\n", cols, rows, len(tiles)/32, palIdx)
	img := image.NewRGBA(image.Rect(0, 0, cols*32, rows*32))
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			blk := int(mp[r*stride+c])
			for ty := 0; ty < 4; ty++ {
				for tx := 0; tx < 4; tx++ {
					tile := int(rom[blkTable+blk*16+ty*4+tx])
					px := gamegear.DecodeTile(tiles[tile*32:])
					for y := 0; y < 8; y++ {
						for x := 0; x < 8; x++ {
							img.Set(c*32+tx*8+x, r*32+ty*8+y, pal[px[y][x]].(color.RGBA))
						}
					}
				}
			}
		}
	}
	f, _ := os.Create(*out)
	png.Encode(f, img)
	f.Close()
	for i := 28; i < 36; i++ {
		dd := T + w(rom, T+i*2)
		fmt.Printf("idx%d: top=%d bottom=%d spawn=(%d,%d) objtbl@%#x nobj=%d\n",
			i, w(rom, dd+9), w(rom, dd+11), rom[dd+13], int(rom[dd+14])-1,
			T+w(rom, dd+30), rom[T+w(rom, dd+30)])
	}
}
