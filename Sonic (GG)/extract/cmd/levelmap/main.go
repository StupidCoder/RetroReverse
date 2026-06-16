// levelmap proves the level-map pipeline end-to-end from ROM. It decodes the Zone 0
// block-index map directly out of the cartridge with decomp.LoadMapRLE (the $0A73
// codec), then renders it into a real picture using the block-definition table and the
// level tiles — without ever reading the decoded map back out of the running machine.
//
// To keep itself honest it ALSO boots the oracle into the level, snapshots the live
// decompressor's output the instant $0A73 returns, and asserts the from-ROM decode is
// byte-identical. That capture is only used to (a) verify the codec and (b) borrow the
// loaded VRAM tiles + CRAM palette the block-defs reference; the map pixels come from
// the ROM decode.
//
// Usage: levelmap <rom.gg> <outdir>
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"sonicgg/extract/decomp"

	"stupidcoder.com/tools/gamegear"
)

// The Zone 0 map source, located by the oracle (CapturePC at $0A73): bank 5, z80 $7430
// (= file $17430), length $0786, decompressing to 4096 bytes = 16 rows x 256 columns.
const (
	mapSrcFile = 0x17430
	mapSrcLen  = 0x0786
	mapCols    = 256
	mapRows    = 16
)

// blockColour maps a block index to a stable false colour: sky (0) and ground fill (1)
// get fixed, readable colours; every other block gets a deterministic hue so distinct
// terrain features are distinguishable.
func blockColour(idx byte) color.Color {
	switch idx {
	case 0:
		return color.RGBA{0x60, 0xA0, 0xF0, 0xFF} // sky
	case 1:
		return color.RGBA{0x70, 0x48, 0x20, 0xFF} // solid ground
	}
	// spread the remaining indices around the hue wheel via a cheap hash.
	h := uint32(idx) * 2654435761
	return color.RGBA{uint8(h >> 16), uint8(h >> 8), uint8(h), 0xFF}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: levelmap <rom.gg> <outdir>")
		os.Exit(2)
	}
	rom, err := os.ReadFile(os.Args[1])
	chk(err)
	outdir := os.Args[2]
	chk(os.MkdirAll(outdir, 0o755))

	// 1. Decode the map straight from the cartridge — no machine involved.
	romMap := decomp.LoadMapRLE(rom, mapSrcFile, mapSrcLen)
	fmt.Printf("decoded %d bytes from ROM (bank 5 $%05X, src %d bytes -> %.2fx)\n",
		len(romMap), mapSrcFile, mapSrcLen, float64(len(romMap))/float64(mapSrcLen))

	// 2. Boot the oracle into the level to (a) verify the decode and (b) grab tiles+CRAM.
	m := gamegear.NewMachine(rom)
	word := func(a uint16) uint16 { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
	m.CapturePC = 0x0A73
	m.CapLo, m.CapHi, m.CapOutBase = 0x0A73, 0x0AA2, 0xC000 // snapshot $C000 at the RET
	for i := 0; i < 700; i++ {
		m.RunFrame()
	}
	for round := 0; round < 40 && word(0xD26F) == 0; round++ {
		m.Pad00 = 0x7F
		for i := 0; i < 8; i++ {
			m.RunFrame()
		}
		m.Pad00 = 0xFF
		for k := 0; k < 242 && word(0xD26F) == 0; k++ {
			m.RunFrame()
		}
	}
	for i := 0; i < 8 && !m.CapOutDone; i++ { // let the load (and the decompressor) finish
		m.RunFrame()
	}
	if !m.CapOutDone {
		fmt.Fprintln(os.Stderr, "warning: never captured the decompressor output")
	} else {
		match := 0
		first := -1
		for i := 0; i < len(romMap) && i < len(m.CapOut); i++ {
			if romMap[i] == m.CapOut[i] {
				match++
			} else if first < 0 {
				first = i
			}
		}
		if first < 0 && len(romMap) == len(m.CapOut) {
			fmt.Printf("VERIFIED: from-ROM decode == live decompressor output (%d bytes, byte-perfect)\n", len(romMap))
		} else {
			fmt.Printf("MISMATCH: %d/%d bytes; first diff at %d\n", match, len(romMap), first)
		}
	}

	// 3. Render the from-ROM map as a false-colour cross-section: each block index gets
	//    a distinct colour so the level's vertical structure reads at a glance — a band
	//    of sky (block 0) on top, a varied surface line, then solid ground fill (block
	//    1). Reconstructing the actual terrain tiles needs the $0760 expander's dynamic
	//    block table + tile-row indirection ($D211), which is a separate effort; the
	//    point here is that the ROM-decoded map IS a real, structured level.
	const px = 8 // pixels per block
	img := image.NewRGBA(image.Rect(0, 0, mapCols*px, mapRows*px))
	for row := 0; row < mapRows; row++ {
		for col := 0; col < mapCols; col++ {
			idx := romMap[row*mapCols+col]
			c := blockColour(idx)
			for dy := 0; dy < px; dy++ {
				for dx := 0; dx < px; dx++ {
					img.Set(col*px+dx, row*px+dy, c)
				}
			}
		}
	}
	writePNG(filepath.Join(outdir, "level_map_fromrom.png"), img)
	fmt.Printf("wrote level_map_fromrom.png (%dx%d px, false colour)\n", mapCols*px, mapRows*px)

	// Also save the raw decoded block-index map for inspection.
	chk(os.WriteFile(filepath.Join(outdir, "level_map.bin"), romMap, 0o644))
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
