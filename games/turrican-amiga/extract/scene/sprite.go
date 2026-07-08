package scene

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Frame is a decoded BOB animation frame: a 4-bitplane plane-major bitmap, one word
// narrower than BLTSIZE's width (the cookie-cut shift reads an extra word).
type Frame struct {
	Bitmap int  // bitmap data address
	H      int  // height in pixels
	DW     int  // data width in 16-px words
	VFlip  bool // descriptor +$D flag set -> the engine draws it vertically flipped
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
	return NthFrame(sp, ft, 0, gfxLo, gfxHi)
}

// NthFrame decodes BOB descriptor number n (frame index) of the frame table at ft — the
// frame the AI selects via node+$C — validating its bitmap lies in [gfxLo,gfxHi).
func NthFrame(sp Space, ft, n, gfxLo, gfxHi int) (Frame, bool) {
	if ft == 0 || n < 0 || !sp.Has(ft+n*4, 4) {
		return Frame{}, false
	}
	p := sp.be32(ft + n*4)
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
	// Descriptor +$D flag: the engine draws these via the alt BOB path (draw_object_bob_alt
	// $6202), a descending blit from the bitmap's bottom row = a vertical flip. The bitmap
	// itself is the same pixels, so we decode it normally and reverse the rows on draw.
	vflip := sp.Data[p+0xD-sp.Base] != 0
	return Frame{Bitmap: bm, H: h, DW: dw, VFlip: vflip}, true
}

// FrameAt decodes frame n of frame table ft for world w in the given space (resident or
// scene block), picking the right gfx range. Convenience wrapper over NthFrame.
func (g *Game) FrameAt(w, ft, n int, resident bool) (Frame, bool) {
	sp := g.Space(w, resident)
	lo, hi := g.GfxRange(w, resident)
	return NthFrame(sp, ft, n, lo, hi)
}

// DrawFirstFrame decodes one 4-bitplane plane-major BOB into img at (ox,oy); colour
// index 0 is left transparent.
func DrawFirstFrame(img *image.Paletted, sp Space, f Frame, ox, oy int) {
	bpr := f.DW * 2
	planeSize := f.H * bpr
	for y := 0; y < f.H; y++ {
		dy := y
		if f.VFlip { // draw bottom row first (the descending-blit vertical flip)
			dy = f.H - 1 - y
		}
		for x := 0; x < f.DW*16; x++ {
			var v uint8
			for p := 0; p < 4; p++ {
				a := f.Bitmap + p*planeSize + y*bpr + x/8
				if sp.Has(a, 1) && sp.Data[a-sp.Base]&(0x80>>(x%8)) != 0 {
					v |= 1 << uint(p)
				}
			}
			if v != 0 {
				img.SetColorIndex(ox+x, oy+dy, v)
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
