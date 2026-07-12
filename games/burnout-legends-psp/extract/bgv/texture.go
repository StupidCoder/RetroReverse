package bgv

import (
	"fmt"
	"image"
	"image/color"
)

// Texture region, reached from header +0x60:
//
//	+0x08  offset of the palette, from the region start
//	+0x0C  width          +0x10  height
//	+0x14  bits per pixel (the vehicles use 8: a 256-entry palette)
//	+0x110 the pixel data
//
// The pixels are stored SWIZZLED — the GE's 16-byte by 8-row block order — and
// the palette entries are RGBA8888, exactly as the hardware samples them. The
// vehicle skins are one 256x256 atlas; every mesh of the car indexes into it.
const texDataOffset = 0x110

// Texture decodes the model's texture atlas into an image.
func Texture(data []byte) (*image.RGBA, error) {
	if len(data) < 0x64 {
		return nil, fmt.Errorf("bgv: too small for a texture region")
	}
	region := int(le32(data, 0x60))
	if region <= 0 || region+texDataOffset > len(data) {
		return nil, fmt.Errorf("bgv: bad texture region 0x%X", region)
	}
	var (
		clutOff = region + int(le32(data, region+0x08))
		w       = int(le32(data, region+0x0C))
		h       = int(le32(data, region+0x10))
		bpp     = int(le32(data, region+0x14))
		pixels  = region + texDataOffset
	)
	if bpp != 8 {
		return nil, fmt.Errorf("bgv: unsupported texture depth %d", bpp)
	}
	if w <= 0 || h <= 0 || w > 1024 || h > 1024 {
		return nil, fmt.Errorf("bgv: implausible texture %dx%d", w, h)
	}
	if pixels+w*h > len(data) || clutOff+256*4 > len(data) {
		return nil, fmt.Errorf("bgv: texture data out of range")
	}

	pal := make([]color.RGBA, 256)
	for i := range pal {
		p := clutOff + i*4
		pal[i] = color.RGBA{data[p], data[p+1], data[p+2], data[p+3]}
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, pal[data[pixels+unswizzle(x, y, w)]])
		}
	}
	return img, nil
}

// unswizzle maps a texel position to its offset in the GE's swizzled layout:
// 16-byte wide, 8-row tall blocks stored one after another. At 8 bits per texel
// the byte and texel widths coincide.
func unswizzle(x, y, rowBytes int) int {
	blocksPerRow := rowBytes / 16
	block := (y/8)*blocksPerRow + x/16
	return block*128 + (y%8)*16 + x%16
}
