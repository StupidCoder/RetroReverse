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

// Image renders image i (format 4, raw 8-bit — e.g. the door textures) through
// the palette. Palette index 0 is transparent.
func (g *GR) Image(i int, pal Palette) (*image.RGBA, error) {
	return g.Sprite(i, pal, nil)
}

// Sprite renders image i through the palette, decoding both the raw 8-bit
// format (4) and the RLE format (8). Format 8 packs 4-bit codes and needs an
// auxiliary palette: its header names an index into allPals (32 × 16 bytes,
// each mapping a 4-bit code to an 8-bit main-palette index). Pass allPals for
// object sprites (OBJECTS.GR); it is ignored for format-4 images. Index 0 is
// transparent (alpha 0).
func (g *GR) Sprite(i int, pal Palette, allPals []byte) (*image.RGBA, error) {
	if i < 0 || i >= len(g.offsets) {
		return nil, fmt.Errorf("tex: image %d out of range (%d)", i, len(g.offsets))
	}
	off := int(g.offsets[i])
	if off+5 > len(g.data) {
		return nil, fmt.Errorf("tex: image %d header outside file", i)
	}
	format := g.data[off]
	w, h := int(g.data[off+1]), int(g.data[off+2])
	var idx []byte // one 8-bit main-palette index per pixel
	switch format {
	case 4: // raw 8-bit
		if off+5+w*h > len(g.data) {
			return nil, fmt.Errorf("tex: image %d %dx%d data outside file", i, w, h)
		}
		idx = g.data[off+5 : off+5+w*h]
	case 8: // 4-bit RLE + aux palette
		aux := int(g.data[off+3])
		if aux*16+16 > len(allPals) {
			return nil, fmt.Errorf("tex: image %d needs aux palette %d", i, aux)
		}
		ap := allPals[aux*16 : aux*16+16]
		codes := decodeFormat8(g.data[off:], w, h)
		idx = make([]byte, w*h)
		for p, c := range codes {
			if int(c) < 16 {
				idx[p] = ap[c]
			}
		}
	default:
		return nil, fmt.Errorf("tex: image %d format %d not supported (4/8)", i, format)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for p := 0; p < w*h; p++ {
		c := pal[idx[p]]
		if idx[p] == 0 {
			c.A = 0 // transparent colour key
		} else {
			c.A = 255
		}
		img.Pix[p*4+0], img.Pix[p*4+1], img.Pix[p*4+2], img.Pix[p*4+3] = c.R, c.G, c.B, c.A
	}
	return img, nil
}

// decodeFormat8 expands a format-8 frame (header at data[0]) to w*h 4-bit codes.
// The frame is [08][W][H][aux][u16 nibbleCount][packed nibbles]; the nibbles are
// a run/dump RLE (reverse-engineered from the game's decoder at 07F7:018D):
// each operation is a RUN (repeat one code) followed by a DUMP (literal codes);
// a run byte of 2 switches to a "consecutive run-only" mode for a counted number
// of operations; 0 and 1 select extended run encodings; a dump/run count of 0
// escapes to a wider count. See disasm annotations §10g.
func decodeFormat8(frame []byte, w, h int) []byte {
	if len(frame) < 6 {
		return make([]byte, w*h)
	}
	ncount := int(binary.LittleEndian.Uint16(frame[4:]))
	nibs := make([]byte, 0, ncount)
	for bi := 6; bi < len(frame) && len(nibs) < ncount; bi++ {
		nibs = append(nibs, frame[bi]>>4, frame[bi]&0xF)
	}
	if len(nibs) > ncount {
		nibs = nibs[:ncount]
	}
	out := make([]byte, 0, w*h)
	si := 0
	rd := func() int {
		if si >= len(nibs) {
			return 0
		}
		v := nibs[si]
		si++
		return int(v)
	}
	dump := func() {
		d := rd()
		if d == 0 {
			if nn := rd(); nn != 0 {
				d = nn<<4 | rd()
			} else {
				d = rd()<<12 | rd()<<8 | rd()<<4 | rd()
			}
		}
		for k := 0; k < d; k++ {
			out = append(out, byte(rd()))
		}
	}
	var op func(dumpEnabled bool)
	op = func(dumpEnabled bool) {
		r := rd()
		switch {
		case r >= 3:
			px := byte(rd())
			for k := 0; k < r; k++ {
				out = append(out, px)
			}
		case r == 2: // consecutive run-only ops, then one dump
			c := rd()
			if c == 0 {
				if nn := rd(); nn != 0 {
					c = nn<<4 | rd()
				} else {
					c = rd()<<12 | rd()<<8 | rd()<<4 | rd()
				}
			}
			for k := 0; k < c && len(out) < w*h; k++ {
				op(false)
			}
			dump()
			return
		case r == 1: // no run
		default: // r == 0: extended run
			e := rd()
			var c int
			if e != 0 {
				c = e<<4 | rd()
			} else {
				c = rd()<<12 | rd()<<8 | rd()<<4 | rd()
			}
			px := byte(rd())
			for k := 0; k < c; k++ {
				out = append(out, px)
			}
		}
		if dumpEnabled {
			dump()
		}
	}
	for si < len(nibs) && len(out) < w*h {
		op(true)
	}
	for len(out) < w*h {
		out = append(out, 0)
	}
	return out[:w*h]
}
