// Command sprites extracts Turrican's BOB (blitter-object) sprites from the
// resident engine and writes one PNG sheet per sprite, laying out all of its
// animation frames (Part V).
//
//	sprites [-o dir] [Turrican.adf]
//
// A sprite is an animation: a frame table (an array of pointers) whose entries
// are 14-byte BOB descriptors. draw_object_bob ($603A) reads each descriptor:
//
//	+$0 bitmap data ptr   +$4 mask ptr   +$8 dest modulo
//	+$A BLTSIZE = height<<6 | width-in-words   +$C y-adjust   +$D flag
//
// The pixels are 4 bitplanes stored plane-major (plane 0's rows, then plane 1's,
// …), one word narrower than BLTSIZE's width (the cookie-cut shift reads an extra
// word), so a frame is 4 * height * (width-1)*2 bytes, drawn through the 16-colour
// playfield palette (plane 3 doubles as the cookie-cut mask, so opaque pixels use
// colours 8-15; colour 0 is transparent). The resident frame tables are found by
// scanning for runs of pointers that all resolve to a valid descriptor.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"turrican/extract/decrunch"
)

const (
	residentLo = 0x10
	residentHi = 0x1B780
	gfxLo      = 0x10000 // sprite bitmaps live in this resident region
	sheetCols  = 8
	pad        = 2
)

func main() {
	out := flag.String("o", "rendered/sprites", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}

	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fail(err)
	}
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		fail(err)
	}
	img := res.Data // file offset == runtime address for the resident segment

	pal := worldPalette(adf)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	for _, t := range findTables(img) {
		sheet := renderSheet(img, pal, t)
		name := fmt.Sprintf("sprite_%05X.png", t.addr)
		if err := writePNG(filepath.Join(*out, name), sheet); err != nil {
			fail(err)
		}
		fmt.Printf("$%05X: %d frames -> %s\n", t.addr, len(t.frames), name)
	}
}

type table struct {
	addr   int
	frames []frame
}
type frame struct {
	bitmap, h, w int // w = data width in words
}

// findTables scans the resident segment for runs of >=3 pointers that all
// resolve to plausible BOB descriptors.
func findTables(img []byte) []table {
	be32 := func(o int) int { return int(binary.BigEndian.Uint32(img[o:])) }
	be16 := func(o int) int { return int(binary.BigEndian.Uint16(img[o:])) }
	descAt := func(p int) (frame, bool) {
		if p < residentLo || p+14 > residentHi {
			return frame{}, false
		}
		bm := be32(p)
		bs := be16(p + 0xA)
		h, w := bs>>6, bs&0x3F
		if bm < gfxLo || bm >= residentHi || h < 4 || h > 96 || w < 2 || w > 10 {
			return frame{}, false
		}
		return frame{bitmap: bm, h: h, w: w - 1}, true
	}

	var out []table
	for a := residentLo; a < residentHi-4; {
		if f0, ok := descAt(be32(a)); ok {
			frames := []frame{f0}
			j := a + 4
			for j < residentHi-4 {
				f, ok := descAt(be32(j))
				if !ok {
					break
				}
				frames = append(frames, f)
				j += 4
			}
			if len(frames) >= 3 {
				out = append(out, table{addr: a, frames: frames})
				a = j
				continue
			}
		}
		a += 2 // tables are word- but not always long-aligned
	}
	return out
}

// renderSheet lays a sprite's frames out in a grid.
func renderSheet(img []byte, pal color.Palette, t table) *image.Paletted {
	cw, ch := 0, 0
	for _, f := range t.frames {
		if f.w*16 > cw {
			cw = f.w * 16
		}
		if f.h > ch {
			ch = f.h
		}
	}
	cw += pad
	ch += pad
	rows := (len(t.frames) + sheetCols - 1) / sheetCols
	sheet := image.NewPaletted(image.Rect(0, 0, sheetCols*cw, rows*ch), pal)
	for i, f := range t.frames {
		ox, oy := (i%sheetCols)*cw, (i/sheetCols)*ch
		drawBob(sheet, img, f, ox, oy)
	}
	return sheet
}

// drawBob decodes one 4-bitplane plane-major BOB into the sheet at (ox,oy).
func drawBob(dst *image.Paletted, img []byte, f frame, ox, oy int) {
	bpr := f.w * 2
	planeSize := f.h * bpr
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w*16; x++ {
			var v uint8
			for p := 0; p < 4; p++ {
				o := f.bitmap + p*planeSize + y*bpr + x/8
				if o < len(img) && img[o]&(0x80>>(x%8)) != 0 {
					v |= 1 << uint(p)
				}
			}
			if v != 0 { // colour 0 is transparent (leave the palette-0 background)
				dst.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}
}

// worldPalette reads world 0's 16-colour playfield palette (the BOBs use it).
func worldPalette(adf []byte) color.Palette {
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		fail(err)
	}
	const off = 0x46A // level table; world 0
	o := int(binary.BigEndian.Uint32(res.Data[off:]))
	n := int(binary.BigEndian.Uint32(res.Data[off+4:]))
	block, err := decrunch.DecrunchBlock(adf[o : o+n])
	if err != nil {
		fail(err)
	}
	palOff := int(binary.BigEndian.Uint32(block[8:])) - 0x1B980
	pal := color.Palette{color.RGBA{0, 0, 0, 0}} // index 0 transparent
	for i := 1; i < 16; i++ {
		c := binary.BigEndian.Uint16(block[palOff+i*2:])
		pal = append(pal, color.RGBA{
			R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255,
		})
	}
	return pal
}

func mainBlob(adf []byte) []byte {
	const off = 0x2C00
	return adf[off : off+int(binary.BigEndian.Uint32(adf[off:]))]
}
func writePNG(path string, im image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, im)
}
func fail(err error) {
	fmt.Fprintln(os.Stderr, "sprites:", err)
	os.Exit(1)
}
