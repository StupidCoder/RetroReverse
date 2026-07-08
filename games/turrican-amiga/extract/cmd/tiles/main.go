// Command tiles renders a world's tile sheet and palette from Turrican's
// disk-streamed level blocks (Part IV).
//
//	tiles [-world N] [-o out.png] [Turrican.adf]
//
// It decrunches the main part, reads the per-world level table at runtime $46A
// (file offset $46A of the decoded image) to find the world's packed scene block
// on the disk, decodes that block with the same three-pass decoder, then walks
// the scene_manager directory at the block's head:
//
//	header +$00  tile offset table (N longs, byte offsets from the table base)
//	header +$08  palette (16 big-endian 12-bit RGB words)
//
// The tiles are 32x32 pixels, four bitplanes, interleaved per row (4 bytes per
// plane per row = 512 bytes per tile). Pointers inside a block are absolute
// runtime addresses baked for a load at $1B980.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/games/turrican-amiga/extract/decrunch"
)

const (
	blockBase  = 0x1B980 // where a scene block loads at run time
	levelTable = 0x46A   // file offset of the per-world (offset,length) table
	numWorlds  = 5
	tileW      = 32
	tileH      = 32
	tileBytes  = tileH * 4 * (tileW / 8) // 512
	sheetCols  = 16
)

func main() {
	world := flag.Int("world", 0, "world number (0..4)")
	out := flag.String("o", "", "output PNG (default: world<N>_tiles.png)")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}
	if *out == "" {
		*out = fmt.Sprintf("world%d_tiles.png", *world)
	}

	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fail(err)
	}
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		fail(err)
	}
	img := res.Data

	// Per-world table: (ADF offset, packed length).
	t := levelTable + *world*8
	off := int(binary.BigEndian.Uint32(img[t:]))
	length := int(binary.BigEndian.Uint32(img[t+4:]))
	if *world < 0 || *world >= numWorlds || off+length > len(adf) {
		fail(fmt.Errorf("world %d: bad table entry off=$%X len=$%X", *world, off, length))
	}

	block, err := decrunch.DecrunchBlock(adf[off : off+length])
	if err != nil {
		fail(err)
	}

	at := func(addr int) int { return addr - blockBase } // runtime addr -> block offset
	be32 := func(o int) int { return int(binary.BigEndian.Uint32(block[o:])) }

	palOff := at(be32(0x08))                                // header +$08 -> palette
	tableOff := at(be32(0x00))                              // header +$00 -> tile offset table
	tile0 := int(binary.BigEndian.Uint32(block[tableOff:])) // first entry = table size
	nTiles := tile0 / 4

	pal := make(color.Palette, 16)
	for i := range pal {
		c := binary.BigEndian.Uint16(block[palOff+i*2:])
		pal[i] = color.RGBA{
			R: uint8((c>>8)&0xF) * 17,
			G: uint8((c>>4)&0xF) * 17,
			B: uint8(c&0xF) * 17,
			A: 255,
		}
	}

	rows := (nTiles + sheetCols - 1) / sheetCols
	sheet := image.NewPaletted(image.Rect(0, 0, sheetCols*tileW, rows*tileH), pal)
	for n := 0; n < nTiles; n++ {
		tileOff := tableOff + int(binary.BigEndian.Uint32(block[tableOff+n*4:]))
		if tileOff+tileBytes > len(block) {
			break
		}
		ox, oy := (n%sheetCols)*tileW, (n/sheetCols)*tileH
		for y := 0; y < tileH; y++ {
			var planes [4]uint32
			for p := 0; p < 4; p++ {
				planes[p] = binary.BigEndian.Uint32(block[tileOff+(y*4+p)*4:])
			}
			for x := 0; x < tileW; x++ {
				var v uint8
				for p := 0; p < 4; p++ {
					v |= uint8((planes[p]>>(31-uint(x)))&1) << uint(p)
				}
				sheet.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}

	f, err := os.Create(*out)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	if err := png.Encode(f, sheet); err != nil {
		fail(err)
	}
	fmt.Printf("world %d: %d tiles (32x32x4), palette @ block+$%X -> %s\n", *world, nTiles, palOff, *out)
}

// mainBlob slices the crunched main part (disk $2C00, length from its head).
func mainBlob(adf []byte) []byte {
	const off = 0x2C00
	n := int(binary.BigEndian.Uint32(adf[off:]))
	return adf[off : off+n]
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "tiles:", err)
	os.Exit(1)
}
