// Package tex decodes Ultima Underworld's wall/floor textures (the .TR files)
// and the VGA palettes (PALS.DAT), deriving both formats from the file bytes.
//
// A .TR file is: byte 0 = 2 (version), byte 1 = the square texture dimension
// (W64.TR = 64, F32.TR = 32), a uint16 texture count, then that many uint32
// offsets, then the textures — each dim*dim palette indices, one byte per pixel.
// PALS.DAT holds 8 palettes of 256 6-bit RGB triples (VGA DAC values, 0-63).
package tex

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// Palette is 256 8-bit RGB entries.
type Palette [256]color.RGBA

// LoadPalette reads palette n (0-7) from PALS.DAT, scaling the 6-bit VGA
// components up to 8-bit. Index 0 is treated as transparent (UW's convention for
// the colour key in sprites; for opaque wall/floor textures it is unused).
func LoadPalette(pals []byte, n int) (Palette, error) {
	var p Palette
	if n < 0 || (n+1)*768 > len(pals) {
		return p, fmt.Errorf("tex: palette %d out of range (%d bytes)", n, len(pals))
	}
	base := n * 768
	for i := 0; i < 256; i++ {
		r := pals[base+i*3+0] << 2
		g := pals[base+i*3+1] << 2
		b := pals[base+i*3+2] << 2
		a := byte(255)
		if i == 0 {
			a = 0
		}
		p[i] = color.RGBA{r, g, b, a}
	}
	return p, nil
}

// TR is a parsed .TR texture bank.
type TR struct {
	Dim     int      // square texture side (64 for walls, 32 for floors)
	data    []byte   // whole file
	offsets []uint32 // per-texture pixel-data offsets
}

// ParseTR decodes the .TR header.
func ParseTR(data []byte) (*TR, error) {
	if len(data) < 4 || data[0] != 2 {
		return nil, fmt.Errorf("tex: not a .TR file")
	}
	dim := int(data[1])
	count := int(binary.LittleEndian.Uint16(data[2:]))
	if 4+count*4 > len(data) {
		return nil, fmt.Errorf("tex: .TR offset table overruns file")
	}
	offs := make([]uint32, count)
	for i := range offs {
		offs[i] = binary.LittleEndian.Uint32(data[4+i*4:])
	}
	return &TR{Dim: dim, data: data, offsets: offs}, nil
}

// Count is the number of textures.
func (t *TR) Count() int { return len(t.offsets) }

// Image renders texture i through the palette to an RGBA image.
func (t *TR) Image(i int, pal Palette) (*image.RGBA, error) {
	if i < 0 || i >= len(t.offsets) {
		return nil, fmt.Errorf("tex: texture %d out of range (%d)", i, len(t.offsets))
	}
	off := int(t.offsets[i])
	n := t.Dim * t.Dim
	if off+n > len(t.data) {
		return nil, fmt.Errorf("tex: texture %d data %#x..%#x outside file", i, off, off+n)
	}
	img := image.NewRGBA(image.Rect(0, 0, t.Dim, t.Dim))
	for p := 0; p < n; p++ {
		c := pal[t.data[off+p]]
		c.A = 255 // wall/floor textures are opaque even at palette index 0
		img.Pix[p*4+0] = c.R
		img.Pix[p*4+1] = c.G
		img.Pix[p*4+2] = c.B
		img.Pix[p*4+3] = c.A
	}
	return img, nil
}
