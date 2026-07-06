package tex

import (
	"encoding/binary"
	"fmt"
	"image"
)

// GR is a parsed .GR graphics resource (DOORS.GR etc.): a small archive of
// paletted images. Layout, from the file bytes (DOORS.GR):
//
//	byte 0   = 1 (format)
//	uint16   image count
//	uint32[] per-image offsets
//
// and each image:
//
//	byte  format (4 = raw 8-bit palette indices)
//	byte  width, byte height
//	uint16 data size in bytes (= width*height for format 4)
//	data
type GR struct {
	data    []byte
	offsets []uint32
}

// ParseGR decodes the archive header.
func ParseGR(data []byte) (*GR, error) {
	if len(data) < 3 || data[0] != 1 {
		return nil, fmt.Errorf("tex: not a .GR file")
	}
	count := int(binary.LittleEndian.Uint16(data[1:]))
	if 3+count*4 > len(data) {
		return nil, fmt.Errorf("tex: .GR offset table overruns file")
	}
	offs := make([]uint32, count)
	for i := range offs {
		offs[i] = binary.LittleEndian.Uint32(data[3+i*4:])
	}
	return &GR{data: data, offsets: offs}, nil
}

// Count is the number of images.
func (g *GR) Count() int { return len(g.offsets) }

// Image renders image i through the palette. Only the raw 8-bit format (4) is
// decoded — the door textures use it.
func (g *GR) Image(i int, pal Palette) (*image.RGBA, error) {
	if i < 0 || i >= len(g.offsets) {
		return nil, fmt.Errorf("tex: image %d out of range (%d)", i, len(g.offsets))
	}
	off := int(g.offsets[i])
	if off+5 > len(g.data) {
		return nil, fmt.Errorf("tex: image %d header outside file", i)
	}
	format := g.data[off]
	w, h := int(g.data[off+1]), int(g.data[off+2])
	size := int(binary.LittleEndian.Uint16(g.data[off+3:]))
	if format != 4 {
		return nil, fmt.Errorf("tex: image %d format %d not supported (raw=4)", i, format)
	}
	if size < w*h || off+5+w*h > len(g.data) {
		return nil, fmt.Errorf("tex: image %d %dx%d data outside file", i, w, h)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for p := 0; p < w*h; p++ {
		c := pal[g.data[off+5+p]]
		c.A = 255
		img.Pix[p*4+0] = c.R
		img.Pix[p*4+1] = c.G
		img.Pix[p*4+2] = c.B
		img.Pix[p*4+3] = c.A
	}
	return img, nil
}
