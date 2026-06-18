package scene

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Frame is a decoded BOB animation frame: a 4-bitplane plane-major bitmap, one word
// narrower than BLTSIZE's width (the cookie-cut shift reads an extra word).
type Frame struct {
	Bitmap int // bitmap data address
	H      int // height in pixels
	DW     int // data width in 16-px words
}

// GfxRange returns the [lo,hi) a frame table's bitmaps must fall in for the space a
// sprite lives in (resident vs a scene block).
func (g *Game) GfxRange(w int, resident bool) (lo, hi int) {
	if resident {
		return ResidGfxLo, residHi
	}
	return BlockBase, BlockBase + len(g.blocks[w])
}

// Space picks the space a sprite's data lives in for world w.
func (g *Game) Space(w int, resident bool) Space {
	if resident {
		return g.Resident
	}
	return g.Block(w)
}

// FirstFrame decodes the first BOB descriptor of the frame table at ft, validating
// its bitmap lies in [gfxLo,gfxHi).
func FirstFrame(sp Space, ft, gfxLo, gfxHi int) (Frame, bool) {
	if ft == 0 || !sp.Has(ft, 4) {
		return Frame{}, false
	}
	p := sp.be32(ft)
	if p < sp.Base || !sp.Has(p, 14) {
		return Frame{}, false
	}
	bm := sp.be32(p)
	bs := sp.be16(p + 0xA)
	h, w := bs>>6, bs&0x3F
	dw := w - 1
	if bm < gfxLo || bm >= gfxHi || dw < 1 || h < 1 || h > 96 || !sp.Has(bm, 4*h*dw*2) {
		return Frame{}, false
	}
	return Frame{Bitmap: bm, H: h, DW: dw}, true
}

// DrawFirstFrame decodes one 4-bitplane plane-major BOB into img at (ox,oy); colour
// index 0 is left transparent.
func DrawFirstFrame(img *image.Paletted, sp Space, f Frame, ox, oy int) {
	bpr := f.DW * 2
	planeSize := f.H * bpr
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.DW*16; x++ {
			var v uint8
			for p := 0; p < 4; p++ {
				a := f.Bitmap + p*planeSize + y*bpr + x/8
				if sp.Has(a, 1) && sp.Data[a-sp.Base]&(0x80>>(x%8)) != 0 {
					v |= 1 << uint(p)
				}
			}
			if v != 0 {
				img.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}
}

// WorldPalette reads world w's 16-colour playfield palette. If transparent, index 0
// is made fully transparent (for cookie-cut sprites); otherwise opaque black.
func (g *Game) WorldPalette(w int, transparent bool) color.Palette {
	blk := g.blocks[w]
	off := int(binary.BigEndian.Uint32(blk[0x08:])) - BlockBase
	pal := color.Palette{color.RGBA{0, 0, 0, 255}}
	if transparent {
		pal[0] = color.RGBA{0, 0, 0, 0}
	}
	for i := 1; i < 16; i++ {
		c := binary.BigEndian.Uint16(blk[off+i*2:])
		pal = append(pal, color.RGBA{
			R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255,
		})
	}
	return pal
}
