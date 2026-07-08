// scenemap statically decodes and renders the type-2 attract screen of Sonic the
// Hedgehog (Game Gear) — the screen loaded by the routine traced at $0C7A — without
// running the game. Everything here comes from reading that loader: it decompresses
// background tiles from (bank $0C, $171A), then loads the name table from a stored
// RLE map in bank 5 (two layers at $6C6D and $6DC3, the engine's $0502 codec), and
// uses background palette index $16. This tool reproduces exactly those steps to see
// what scenes 9..17 display — a test of how far pure tracing gets us.
//
// Usage: scenemap <rom.gg> <outdir>
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/games/sonic-gg/extract/decomp"
	"retroreverse.com/tools/platform/gamegear"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: scenemap <rom.gg> <outdir>")
		os.Exit(2)
	}
	rom, err := os.ReadFile(os.Args[1])
	chk(err)
	outdir := os.Args[2]
	chk(os.MkdirAll(outdir, 0o755))

	// The two world maps. The engine selects between them by the upcoming-level
	// countdown $D279 (decremented after each level at bank 1 $4700): the map loaders
	// branch on `LD A,($D279) / CP $06 / JR C` ($2006, $466A, $BF0D), so a high count
	// (early game, >=6) shows the WIDE island and a low count (late game, <6) zooms in
	// on the mountain-top goal. Each variant has its own tile set and name-table map.

	// Wide island ($D279 >= 6): base tiles (bank $0C,$0000) + bank-5 map $6F86 then
	// $716A; palette bg $0A / spr $0B.
	renderMap(rom, outdir, "worldmap_wide", 0x0000, 0x0A, 0x0B,
		layer{0x6F86, 0x01E4, 0x00}, layer{0x716A, 0x01C0, 0x00})

	// Zoomed mountain top ($D279 < 6): tiles (bank $0C,$171A) + bank-5 map $6C6D (hi
	// $10) then $6DC3; palette bg $0C / spr $0D.
	renderMap(rom, outdir, "worldmap_zoom", 0x171A, 0x0C, 0x0D,
		layer{0x6C6D, 0x0156, 0x10}, layer{0x6DC3, 0x0198, 0x00})
}

// layer is one $0502 RLE name-table layer: a bank-5 source address, a source-byte
// count, and the high byte ($D20F) applied to every entry.
type layer struct {
	addr, count int
	hi          byte
}

// renderMap decompresses the background tiles from (bank $0C, tileAddr), composes the
// name table by applying the given bank-5 RLE layers in order (later layers overwrite
// from cell 0, as the engine's repeated $0502 calls do), resolves the palette from
// the bank-8 table, and writes <name>.screen.png and <name>.gg.png.
func renderMap(rom []byte, outdir, name string, tileAddr, bg, spr int, layers ...layer) {
	const bank5 = 5 * 0x4000
	tileSrc := decomp.SourceOffset(0x0C, uint16(tileAddr))
	tiles := decomp.Decompress(rom, tileSrc)
	nt := make([]byte, 32*28*2)
	for _, l := range layers {
		copy(nt, decomp.LoadRLE(rom, bank5+(l.addr-0x4000), l.count, l.hi))
	}
	pal := bankPalette(rom, bg, spr)
	full := gamegear.RenderNameTable(nt, tiles, 32, 28, pal)
	writePNG(filepath.Join(outdir, name+".screen.png"), scale(full, 2))
	gg := full.SubImage(image.Rect(48, 24, 48+160, 24+144)).(*image.Paletted)
	writePNG(filepath.Join(outdir, name+".gg.png"), scale(gg, 3))
	fmt.Printf("wrote %s.{screen,gg}.png (tiles file $%05X)\n", name, tileSrc)
}

// bankPalette resolves a 32-colour palette from the bank-8 pointer table at file
// $23400 (entry = *($23400+i*2); data = $23400 + entry): the 16 background colours
// from index bg, the 16 sprite colours from index spr.
func bankPalette(rom []byte, bg, spr int) color.Palette {
	const tab = 0x23400
	read := func(i int) []byte {
		ptr := int(rom[tab+i*2]) | int(rom[tab+i*2+1])<<8
		off := tab + ptr
		fmt.Printf("palette $%02X: entry $%04X -> file $%05X\n", i, ptr, off)
		return rom[off : off+32]
	}
	cram := make([]byte, 64)
	copy(cram, read(bg))
	copy(cram[32:], read(spr))
	return gamegear.Palette(cram)
}

func scale(src *image.Paletted, n int) *image.Paletted {
	b := src.Bounds()
	out := image.NewPaletted(image.Rect(0, 0, b.Dx()*n, b.Dy()*n), src.Palette)
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			ci := src.ColorIndexAt(b.Min.X+x, b.Min.Y+y)
			for dy := 0; dy < n; dy++ {
				for dx := 0; dx < n; dx++ {
					out.SetColorIndex(x*n+dx, y*n+dy, ci)
				}
			}
		}
	}
	return out
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	chk(err)
	defer f.Close()
	chk(png.Encode(f, img))
}

func chk(e error) {
	if e != nil {
		panic(e)
	}
}
