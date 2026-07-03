// Package ilb decodes Marble Madness .ilb/.vlb sprite banks (Marble_Madness.md
// Part IV §3): the whole file is one ByteRun1/PackBits stream; unpacked, it is
// [count word][count × 20-byte descriptors @ +2][contiguous planar pixel data].
// Each descriptor: +0 type (1 = stored free sprite, row-interleaved planes;
// 0 = composited scenery, sequential plane blocks), +2 width in 16-px words,
// +4 height rows, +6 one-plane byte size, +8 source offset. A cell's plane
// count is its source span (gap to the next cell) divided by the plane size.
package ilb

import (
	"image"
	"image/color"
)

// Cell is one decoded sprite-cell descriptor.
type Cell struct {
	W, H, Planes, CZ, Src, Typ int
}

// unpack decodes PackBits: signed control byte n; 0..127 -> copy n+1 literals;
// -1..-127 -> repeat next byte (1-n) times; -128 -> no-op.
func unpack(src []byte) []byte {
	var out []byte
	for i := 0; i < len(src); {
		n := int8(src[i])
		i++
		switch {
		case n >= 0:
			for k := 0; k <= int(n) && i < len(src); k++ {
				out = append(out, src[i])
				i++
			}
		case n != -128:
			if i >= len(src) {
				return out
			}
			b := src[i]
			i++
			for k := 0; k < 1-int(n); k++ {
				out = append(out, b)
			}
		}
	}
	return out
}

// Decode unpacks a bank file and parses its cell descriptors.
func Decode(file []byte) (buf []byte, cells []Cell) {
	buf = unpack(file)
	wd := func(o int) int {
		if o+1 < len(buf) {
			return int(buf[o])<<8 | int(buf[o+1])
		}
		return 0
	}
	ln := func(o int) int { return wd(o)<<16 | wd(o+2) }
	count := wd(0)
	for i := 0; i < count; i++ {
		o := 2 + i*20
		if o+20 > len(buf) {
			break
		}
		typ, ww, h, cz, src := int(buf[o]), wd(o+2), wd(o+4), wd(o+6), ln(o+8)
		if cz <= 0 || src <= 0 || src >= len(buf) {
			cells = append(cells, Cell{})
			continue
		}
		end := len(buf)
		if i+1 < count {
			if nx := ln(o + 20 + 8); nx > src && nx <= len(buf) {
				end = nx
			}
		}
		planes := (end - src) / cz
		if planes < 1 {
			planes = 1
		} else if planes > 6 {
			planes = 6
		}
		if typ == 1 && h > 1 {
			h-- // type-1 cells carry a trailing sentinel/guard row
		}
		cells = append(cells, Cell{W: ww * 16, H: h, Planes: planes, CZ: cz, Src: src, Typ: typ})
	}
	return buf, cells
}

// RenderRamp draws a type-1 (hardware-sprite) cell with its own 3-colour ramp —
// the three $0RGB words the piece's 16-byte sprite record points at via +$C,
// which the engine loads into the sprite pair's COLOR registers through the
// copper list. Pixel value 0 is transparent (sprite colour 0).
func RenderRamp(buf []byte, c Cell, ramp [3]uint16) *image.RGBA {
	n4 := func(x uint16) uint8 { return uint8(x) * 17 }
	pal := make(color.Palette, 4)
	pal[0] = color.RGBA{}
	for i, w := range ramp {
		pal[i+1] = color.RGBA{n4((w >> 8) & 0xF), n4((w >> 4) & 0xF), n4(w & 0xF), 0xFF}
	}
	return Render(buf, c, pal, 0)
}

// Render draws one cell as an RGBA image with pixel value 0 transparent — the
// engine masks cells through the OR-of-planes silhouette built at load (the
// cookie-cut mask at descriptor +$C), so 0 pixels never reach the screen.
// type-1 sprites add type1Off to non-zero pixels when no per-record ramp is
// supplied (6 = the course object-accent ramp approximation).
func Render(buf []byte, c Cell, pal color.Palette, type1Off int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.W, c.H))
	bpr := c.W / 8
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			var v int
			for p := 0; p < c.Planes; p++ {
				var o int
				if c.Typ == 1 {
					o = c.Src + y*c.Planes*bpr + p*bpr + x/8
				} else {
					o = c.Src + p*c.CZ + y*bpr + x/8
				}
				if o >= 0 && o < len(buf) {
					v |= int(buf[o]>>uint(7-x%8)&1) << uint(p)
				}
			}
			if v == 0 {
				continue // transparent
			}
			if c.Typ == 1 {
				v += type1Off
			}
			if v < len(pal) {
				img.Set(x, y, pal[v])
			}
		}
	}
	return img
}
