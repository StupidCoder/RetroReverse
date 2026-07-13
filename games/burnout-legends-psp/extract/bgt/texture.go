package bgt

import (
	"fmt"
	"image"
	"image/color"
)

// A texture record, reached from the dictionary at header +0x18:
//
//	+0x08  palette offset, from the record
//	+0x0C  width          +0x10  height
//	+0x14  bits per texel — 4 (CLUT4) or 8 (CLUT8)
//	+0x38  the mip chain: four level offsets, 8 bytes apart, from the record
//
// Level 0 always lands at +0x110 and the levels tile exactly up to the palette,
// which holds 2^bpp RGBA8888 entries. The texels are stored SWIZZLED in the GE's
// 16-byte by 8-row block order, exactly as the hardware samples them.
const texHeader = 0x110

// Texture decodes texture i of the dictionary at its top mip level.
func (t *Track) Texture(i int) (*image.RGBA, error) {
	if i < 0 || i >= len(t.Textures) {
		return nil, fmt.Errorf("bgt: no texture %d", i)
	}
	var (
		info = t.Textures[i]
		data = t.data
		rec  = info.Offset
		pix  = rec + int(le32(data, rec+0x38))
	)
	rowBytes := info.W * info.BPP / 8
	if pix < 0 || pix+rowBytes*info.H > len(data) {
		return nil, fmt.Errorf("bgt: texture %d: data out of range", i)
	}

	// The indexed formats carry a palette; the true-colour ones sample directly.
	var pal []color.RGBA
	if info.BPP != 32 {
		clut := rec + int(le32(data, rec+0x08))
		n := 1 << info.BPP // 16 or 256 entries
		if clut < 0 || clut+n*4 > len(data) {
			return nil, fmt.Errorf("bgt: texture %d: palette out of range", i)
		}
		pal = make([]color.RGBA, n)
		for j := range pal {
			p := clut + j*4
			pal[j] = color.RGBA{data[p], data[p+1], data[p+2], data[p+3]}
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, info.W, info.H))
	for y := 0; y < info.H; y++ {
		for x := 0; x < info.W; x++ {
			switch info.BPP {
			case 4: // two texels to a byte, the low nibble first
				b := data[pix+unswizzle(x/2, y, rowBytes)]
				img.SetRGBA(x, y, pal[(b>>(4*(x&1)))&0xF])
			case 8:
				img.SetRGBA(x, y, pal[data[pix+unswizzle(x, y, rowBytes)]])
			default: // 32: RGBA8888
				p := pix + unswizzle(x*4, y, rowBytes)
				img.SetRGBA(x, y, color.RGBA{data[p], data[p+1], data[p+2], data[p+3]})
			}
		}
	}
	return img, nil
}

// unswizzle maps a byte position to its offset in the GE's swizzled layout:
// 16-byte wide, 8-row tall blocks stored one after another. Narrow textures
// (under a block wide) are stored linearly.
func unswizzle(xb, y, rowBytes int) int {
	if rowBytes < 16 {
		return y*rowBytes + xb
	}
	blocksPerRow := rowBytes / 16
	block := (y/8)*blocksPerRow + xb/16
	return block*128 + (y%8)*16 + xb%16
}
