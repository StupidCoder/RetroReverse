package f3d

import (
	"image"
	"image/color"
)

// DecodeTexture reads a group's texture straight out of the RDRAM snapshot,
// using the same texel formats as the oracle's sampler (rdp_texture.go). The
// data in RDRAM is laid out for a dxt=0 Load_Block — pre-swizzled, odd rows
// with their 32-bit word halves exchanged — so the reader undoes the same swap
// the sampler does.
func (w *Walker) DecodeTexture(g *Group) image.Image {
	t := g.Tile
	if t.SH <= t.SL && t.TH <= t.TL {
		return nil
	}
	tw := int(t.SH>>2-t.SL>>2) + 1
	th := int(t.TH>>2-t.TL>>2) + 1
	if tw <= 0 || th <= 0 || tw > 1024 || th > 1024 || t.Img == 0 {
		return nil
	}
	img := image.NewRGBA(image.Rect(0, 0, tw, th))
	rowBytes := int(t.Line) * 8
	texel := func(off int) byte {
		if off < 0 || off >= len(w.RAM) {
			return 0
		}
		return w.RAM[off]
	}
	for y := 0; y < th; y++ {
		row := int(t.Img) + y*rowBytes
		for x := 0; x < tw; x++ {
			var r, gg, b, a uint32
			swz := func(off int) int {
				if y&1 != 0 {
					return off ^ 4
				}
				return off
			}
			switch {
			case t.Fmt == 0 && t.Size == 2: // RGBA16
				v := uint16(texel(swz(row+x*2)))<<8 | uint16(texel(swz(row+x*2)+1))
				r, gg, b = uint32(v>>11&31)<<3, uint32(v>>6&31)<<3, uint32(v>>1&31)<<3
				a = uint32(v&1) * 255
			case t.Fmt == 4 && t.Size == 0: // I4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				i := n * 17
				r, gg, b, a = i, i, i, i
			case t.Fmt == 4 && t.Size == 1: // I8
				i := uint32(texel(swz(row + x)))
				r, gg, b, a = i, i, i, i
			case t.Fmt == 3 && t.Size == 0: // IA4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				i := (n >> 1) * 36
				r, gg, b = i, i, i
				a = uint32(n&1) * 255
			case t.Fmt == 3 && t.Size == 1: // IA8
				v := texel(swz(row + x))
				i := uint32(v>>4) * 17
				r, gg, b, a = i, i, i, uint32(v&15)*17
			case t.Fmt == 3 && t.Size == 2: // IA16
				i := uint32(texel(swz(row + x*2)))
				a = uint32(texel(swz(row+x*2) + 1))
				r, gg, b = i, i, i
			case t.Fmt == 2 && t.Size == 1: // CI8
				idx := uint32(texel(swz(row + x)))
				v := uint16(texel(int(g.TLUT)+int(idx)*2))<<8 | uint16(texel(int(g.TLUT)+int(idx)*2+1))
				r, gg, b = uint32(v>>11&31)<<3, uint32(v>>6&31)<<3, uint32(v>>1&31)<<3
				a = uint32(v&1) * 255
			case t.Fmt == 2 && t.Size == 0: // CI4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				idx := t.Pal<<4 | n
				pv := uint16(texel(int(g.TLUT)+int(idx)*2))<<8 | uint16(texel(int(g.TLUT)+int(idx)*2+1))
				r, gg, b = uint32(pv>>11&31)<<3, uint32(pv>>6&31)<<3, uint32(pv>>1&31)<<3
				a = uint32(pv&1) * 255
			default:
				return nil
			}
			img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(gg), uint8(b), uint8(a)})
		}
	}
	return img
}
