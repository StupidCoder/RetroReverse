package assets

// pvrtex.go decodes a PVRT texture file — the self-describing texture
// format the disc's 0GDTEX.PVR (the boot-menu artwork) uses. The pixel
// formats and the twiddle order are the PVR hardware's own, the same ones
// the oracle's rasteriser reads out of texture control words; this decoder
// is deliberately independent of it (it reads a FILE header, not a TCW)
// and is verified by eye on the boot art.

import (
	"encoding/binary"
	"fmt"
	"image"
)

// PVRTex is a parsed PVRT header with its pixel data.
type PVRTex struct {
	PixFmt   uint8 // 0 = ARGB1555, 1 = RGB565, 2 = ARGB4444
	DataType uint8 // 1 = square twiddled, 3 = VQ, 9 = rectangle raster (+2/+4 mip variants)
	W, H     int
	data     []byte
}

// OpenPVR parses a PVRT file. A leading GBIX chunk (a global-index wrapper
// some discs use) is skipped if present; this disc's file has none.
func OpenPVR(data []byte) (*PVRTex, error) {
	if len(data) >= 8 && string(data[:4]) == "GBIX" {
		skip := 8 + int(binary.LittleEndian.Uint32(data[4:]))
		if skip >= len(data) {
			return nil, fmt.Errorf("pvr: GBIX chunk overruns the file")
		}
		data = data[skip:]
	}
	if len(data) < 16 || string(data[:4]) != "PVRT" {
		return nil, fmt.Errorf("pvr: no PVRT magic")
	}
	t := &PVRTex{
		PixFmt:   data[8],
		DataType: data[9],
		W:        int(binary.LittleEndian.Uint16(data[12:])),
		H:        int(binary.LittleEndian.Uint16(data[14:])),
		data:     data[16:],
	}
	if t.W <= 0 || t.H <= 0 || t.W > 1024 || t.H > 1024 {
		return nil, fmt.Errorf("pvr: unlikely size %dx%d", t.W, t.H)
	}
	return t, nil
}

// Decode renders the texture. Formats are added as files on the disc
// exercise them; an unimplemented one is an error, never a black picture.
func (t *PVRTex) Decode() (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, t.W, t.H))
	px16 := func(off int) uint32 {
		if off+2 > len(t.data) {
			return 0
		}
		return uint32(t.data[off]) | uint32(t.data[off+1])<<8
	}
	put := func(x, y int, v uint32) {
		r, g, b, a := unpack16(t.PixFmt, v)
		i := img.PixOffset(x, y)
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, a
	}
	switch t.DataType {
	case 1, 2: // square twiddled (2 = with mips: levels small-first, top last)
		off := 0
		if t.DataType == 2 {
			// The chain stores the levels below the top first — 1+4+…+(N/2)²
			// = (N²-1)/3 texels — plus three dummy texels of padding; the
			// top level follows. The same formula the oracle's rasteriser
			// uses for a mipmapped TCW.
			side := t.W
			off = ((side*side-1)/3 + 3) * 2
		}
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				put(x, y, px16(off+int(twiddle(uint32(x), uint32(y)))*2))
			}
		}
	case 3: // VQ: a 2048-byte codebook of 256 2x2 entries, then twiddled block indices
		if len(t.data) < 2048 {
			return nil, fmt.Errorf("pvr: VQ texture shorter than its codebook")
		}
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				bx, by := x/2, y/2
				iOff := 2048 + int(twiddle(uint32(bx), uint32(by)))
				if iOff >= len(t.data) {
					continue
				}
				entry := int(t.data[iOff])*8 + ((x&1)*2+(y&1))*2
				put(x, y, px16(entry))
			}
		}
	case 9: // rectangle, raster order
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				put(x, y, px16((y*t.W+x)*2))
			}
		}
	default:
		return nil, fmt.Errorf("pvr: data type %#x unimplemented (census it, then add it)", t.DataType)
	}
	return img, nil
}

// Decode16 renders a header-less 16-bpp page (a TEXDC/LANDDC payload) at a
// caller-chosen size and format, raster order — the raw viewer for pages
// whose real layout the game's own parser will eventually name.
func Decode16(data []byte, pixFmt uint8, w, h int) (*image.RGBA, error) {
	if w <= 0 || h <= 0 || w*h*2 > len(data) {
		return nil, fmt.Errorf("raw16: %dx%d needs %d bytes, have %d", w, h, w*h*2, len(data))
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint32(data[(y*w+x)*2]) | uint32(data[(y*w+x)*2+1])<<8
			r, g, b, a := unpack16(pixFmt, v)
			i := img.PixOffset(x, y)
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, a
		}
	}
	return img, nil
}

// unpack16 expands one 16-bit texel. The numbering is the PVR's own:
// 0 = ARGB1555, 1 = RGB565, 2 = ARGB4444.
func unpack16(fmt uint8, v uint32) (r, g, b, a uint8) {
	switch fmt {
	case 0:
		return uint8(v >> 10 & 31 << 3), uint8(v >> 5 & 31 << 3), uint8(v & 31 << 3), uint8(v >> 15 & 1 * 255)
	case 1:
		return uint8(v >> 11 & 31 << 3), uint8(v >> 5 & 63 << 2), uint8(v & 31 << 3), 255
	default:
		return uint8(v >> 8 & 15 * 17), uint8(v >> 4 & 15 * 17), uint8(v & 15 * 17), uint8(v >> 12 & 15 * 17)
	}
}

// twiddle interleaves x into odd bits and y into even bits — the PVR's
// Morton texture order within a square.
func twiddle(x, y uint32) uint32 {
	var out uint32
	for i := uint32(0); i < 16; i++ {
		out |= (y >> i & 1) << (2 * i)
		out |= (x >> i & 1) << (2*i + 1)
	}
	return out
}
