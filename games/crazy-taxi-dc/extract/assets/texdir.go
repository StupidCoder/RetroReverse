// texdir.go reads the per-course texture directories out of 1ST_READ.BIN —
// the static descriptor arrays the game's own loader walks to create every
// course texture. The arrays were found by tracing the create-texture call
// chain up from the record builder (0C15A2A2 ← 0C15994C ← 0C072B10 ←
// 0C079780): 0C072B10 receives a 16-byte descriptor [w u16][h u16][flags
// u32][staging ptr u32][alias u32] and the course loader indexes the array
// by the model blocks' aux texture id. The staging pointer is the load
// address in the 0CB40000 staging region, so (ptr - 0CB40000) is the
// texture's byte offset in the TEXDC file, BAKED at build time — the whole
// directory decodes offline. Every one of the 216 anchors in the empirical
// truth table (aux id → file offset, from a -gd/-c2dlog cold boot) matches
// TEXDC0's array, and each array's last offset + size lands on its TEXDC
// file's size (32-byte-aligned blobs).
package assets

import (
	"encoding/binary"
	"fmt"
	"image"
)

// Texture layout kinds, from the descriptor flags' second byte. The names
// were pinned by correlating descriptors with the loader-built records'
// final TCWs (mip bit 31, VQ bit 30) over a full course: 4 → mip+VQ,
// 3 → VQ, 2 → mip, 1 → plain twiddled; 0xD marks the non-square textures,
// whose TCWs carry neither bit and whose sizes are exactly w*h*2 — raster
// pages, like the LANDDC billboards. Kinds 5 and 7 appear only in the
// non-course (palette-fed car/UI) directories and are not decoded here.
const (
	TexVQMip    = 4
	TexVQ       = 3
	TexMip      = 2
	TexTwiddled = 1
	TexRaster   = 0xD
)

// TexEntry is one texture: its file placement and how to decode it.
type TexEntry struct {
	W, H int
	Fmt  uint8 // pixel format, the PVR's own: 0 ARGB1555, 1 RGB565, 2 ARGB4444
	Kind uint8
	Off  uint32 // byte offset in the TEXDC file
}

// Size is the texture's byte size in the file (32-byte-aligned, the
// alignment the baked offsets prove).
func (e *TexEntry) Size() uint32 {
	w, h := uint32(e.W), uint32(e.H)
	var n uint32
	switch e.Kind {
	case TexVQMip:
		n = 2048 + vqIndexTop(w) + w*h/4
	case TexVQ:
		n = 2048 + w*h/4
	case TexMip:
		n = ((w*w-1)/3+1)*2 + w*h*2
	default: // TexTwiddled, TexRaster
		n = w * h * 2
	}
	return (n + 31) &^ 31
}

// vqIndexTop is the top mip level's offset within a VQ index area: the
// smaller levels come first, one index byte per 2x2 block with a one-byte
// floor (8x8 → 6, 16x16 → 0x16 — the same ladder the oracle's rasteriser
// climbs for a mipmapped VQ TCW).
func vqIndexTop(w uint32) uint32 {
	var off uint32
	for l := uint32(1); l < w; l *= 2 {
		off += max(1, l*l/4)
	}
	return off
}

// TexDir is one course's texture directory.
type TexDir struct {
	Entries []TexEntry
}

// The four arrays inside 1ST_READ.BIN (loaded at 0C010000; file offset =
// address - 0C010000) and their entry counts. Identified by matching each
// array's cumulative size against its TEXDC file's size; TEXDC0's is
// additionally pinned by the 216-anchor truth table.
var texDirs = [4]struct {
	off, n int
}{
	{0x130570, 438}, // 0C140570  TEXDC0
	{0x134CB0, 336}, // 0C144CB0  TEXDC1
	{0x1361C0, 324}, // 0C1461C0  TEXDC2
	{0x137610, 225}, // 0C147610  TEXDC3
}

// OpenTexDir parses course c's texture directory out of 1ST_READ.BIN.
func OpenTexDir(firstRead []byte, course int) (*TexDir, error) {
	if course < 0 || course > 3 {
		return nil, fmt.Errorf("texdir: course %d out of range", course)
	}
	le := binary.LittleEndian
	td := texDirs[course]
	if len(firstRead) < td.off+td.n*16 {
		return nil, fmt.Errorf("texdir: 1ST_READ.BIN too short for course %d (%d bytes)", course, len(firstRead))
	}
	d := &TexDir{}
	for i := 0; i < td.n; i++ {
		o := td.off + i*16
		e := TexEntry{
			W:    int(le.Uint16(firstRead[o:])),
			H:    int(le.Uint16(firstRead[o+2:])),
			Fmt:  firstRead[o+4],
			Kind: firstRead[o+5],
		}
		stg := le.Uint32(firstRead[o+8:])
		if stg < 0x0CB40000 {
			return nil, fmt.Errorf("texdir: entry %d staging %08X below the staging region", i, stg)
		}
		e.Off = stg - 0x0CB40000
		if e.W == 0 || e.H == 0 || e.W > 1024 || e.H > 1024 {
			return nil, fmt.Errorf("texdir: entry %d unlikely size %dx%d", i, e.W, e.H)
		}
		d.Entries = append(d.Entries, e)
	}
	return d, nil
}

// Decode renders entry i out of the TEXDC file's bytes.
func (d *TexDir) Decode(i int, texdc []byte) (*image.RGBA, error) {
	if i < 0 || i >= len(d.Entries) {
		return nil, fmt.Errorf("texdir: no entry %d", i)
	}
	e := &d.Entries[i]
	if int(e.Off)+int(e.Size()) > len(texdc) {
		return nil, fmt.Errorf("texdir: entry %d (%X+%X) past the file's %d bytes", i, e.Off, e.Size(), len(texdc))
	}
	data := texdc[e.Off:]
	img := image.NewRGBA(image.Rect(0, 0, e.W, e.H))
	put := func(x, y int, v uint32) {
		r, g, b, a := unpack16(e.Fmt, v)
		o := img.PixOffset(x, y)
		img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, g, b, a
	}
	px16 := func(off uint32) uint32 {
		return uint32(data[off]) | uint32(data[off+1])<<8
	}
	switch e.Kind {
	case TexTwiddled, TexMip:
		if e.W != e.H {
			return nil, fmt.Errorf("texdir: entry %d twiddled but %dx%d not square", i, e.W, e.H)
		}
		var off uint32
		if e.Kind == TexMip {
			// The FILE's mip chain pads the sub-top levels with ONE dummy
			// texel, not the three the PVR's VRAM layout uses — proven
			// texel-exact against the loader's uploaded copy (every texel of
			// every probed texture matches at +2, none at +6).
			w := uint32(e.W)
			off = ((w*w-1)/3 + 1) * 2
		}
		for y := 0; y < e.H; y++ {
			for x := 0; x < e.W; x++ {
				put(x, y, px16(off+twiddle(uint32(x), uint32(y))*2))
			}
		}
	case TexVQ, TexVQMip:
		if e.W != e.H {
			return nil, fmt.Errorf("texdir: entry %d VQ but %dx%d not square", i, e.W, e.H)
		}
		idx := uint32(2048)
		if e.Kind == TexVQMip {
			idx += vqIndexTop(uint32(e.W))
		}
		for y := 0; y < e.H; y++ {
			for x := 0; x < e.W; x++ {
				entry := uint32(data[idx+twiddle(uint32(x)/2, uint32(y)/2)])
				put(x, y, px16(entry*8+uint32((x&1)*2+(y&1))*2))
			}
		}
	case TexRaster:
		for y := 0; y < e.H; y++ {
			for x := 0; x < e.W; x++ {
				put(x, y, px16(uint32(y*e.W+x)*2))
			}
		}
	default:
		return nil, fmt.Errorf("texdir: entry %d kind %#x unimplemented", i, e.Kind)
	}
	return img, nil
}
