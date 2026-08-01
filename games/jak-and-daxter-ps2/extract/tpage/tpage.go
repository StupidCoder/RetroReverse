// Package tpage decodes texture-page objects (tpage-NNN) to images.
//
// The layout is read from the engine's own code, not inferred:
//
//   - adgif-shader<-texture! (0x612684) composes TEX0/TEX1 from a `texture`
//     descriptor, naming every field: +0/+2 s16 w/h, +4 mip count, +5 filter
//     (mmag | mmin<<1), +6 psm, +7 L, +8 clut psm, +10/+12/+14/… u16 per-mip
//     destination block, +24 u16 clut block, +26/+27/… u8 per-mip tbw (in
//     64-px units). Blocks are page-relative; at runtime the allocator adds
//     the page's VRAM base and stamps it into the header's segment dst
//     words (observed live: tpage-463 → block 0x2141, and the draw census's
//     TEX0 TBP/CBP values are exactly these fields plus that base).
//   - The texture-page header: P+0 →file-info, P+4 →name, P+8 id, P+12
//     texture count, P+16/+20 word counts, then three segments {ptr,
//     size-in-words, dst}; the descriptor pointer array starts at P+0x7C
//     (entries are #f for punched-out textures). The segments upload back to
//     back — the live dst deltas equal the sizes — so the page's VRAM
//     footprint is their concatenation, addressed through the GS block
//     swizzle for each texture's psm (tools/platform/ps2's own sampler
//     arithmetic, via its TexAddr* exports).
//
// Textures reference the footprint page-relative, so decoding needs no
// running machine: index texels through TexAddrPSMT8/T4, resolve through the
// CSM1 CLUT at the descriptor's clut block, and emit RGBA (PS2 alpha 0x80 =
// opaque, scaled to 8-bit).
package tpage

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"

	"retroreverse.com/tools/platform/ps2"
)

// Texture is one decoded descriptor.
type Texture struct {
	Name    string
	W, H    int
	PSM     int
	ClutPSM int
	Mips    int
	TBP     []uint16 // per-mip destination block, page-relative
	TBW     []byte   // per-mip buffer width in 64-px units
	CBP     uint16   // clut block, page-relative
}

// Page is a decoded texture-page object.
type Page struct {
	Name     string
	ID       int
	Words    uint32 // total VRAM footprint in words
	Textures []Texture

	vram []byte // the page's footprint, byte-addressed from block 0
}

// GS pixel storage formats.
const (
	PSMCT32  = 0x00
	PSMCT24  = 0x01
	PSMCT16  = 0x02
	PSMT8    = 0x13
	PSMT4    = 0x14
	PSMT8H   = 0x1B
	PSMT4HL  = 0x24
	PSMT4HH  = 0x2C
)

func u16at(b []byte, off uint32) uint16 { return binary.LittleEndian.Uint16(b[off:]) }
func u32at(b []byte, off uint32) uint32 { return binary.LittleEndian.Uint32(b[off:]) }

// Load parses a texture-page from its linked object image. Link the object
// at base 0 (goalobj.Link(data, 0, symtab)) so internal pointers are object
// offsets; #f entries then point far outside the object and are skipped.
func Load(obj []byte) (*Page, error) {
	if len(obj) < 0x80 {
		return nil, fmt.Errorf("tpage: %d bytes is too short", len(obj))
	}
	const P = 4 // object pointer: one word past the type slot
	pg := &Page{Name: goalString(obj, u32at(obj, P+4)), ID: int(u32at(obj, P+8))}
	count := int(u32at(obj, P+12))

	type seg struct{ ptr, words uint32 }
	var segs []seg
	for k := uint32(0); k < 3; k++ {
		segs = append(segs, seg{u32at(obj, P+0x18+k*12), u32at(obj, P+0x1C+k*12)})
		pg.Words += segs[k].words
	}
	// Segment 0 uploads as a 128-px-wide PSMCT32 IMAGE transfer at the page
	// base (VERIFIED byte-exact against live VRAM: every CLUT block and
	// far-tier texel checked matches, and four full textures decode
	// pixel-identical through the oracle's sampler). Segments 1-2 (the
	// mid/near mip tiers, streamed in and out at runtime) do NOT tile VRAM
	// cumulatively: live VRAM shows e.g. page-relative block 0x280 holding
	// segment-2 data from word 0x2C0 with a 128-word row stride, so the
	// tier-to-VRAM mapping goes through the engine's tier allocator, not the
	// header dst deltas. OPEN: solve that mapping; until then the model
	// below is segment-0-faithful and best-effort for tiers 1-2 (textures
	// whose mip 0 lives in tier 1/2 may decode with a wrong layout).
	pg.vram = make([]byte, pg.Words*4)
	segBase := uint32(0)
	for _, s := range segs {
		if int64(s.ptr)+int64(s.words)*4 > int64(len(obj)) {
			return nil, fmt.Errorf("tpage %s: segment [0x%x +0x%x words] outside object", pg.Name, s.ptr, s.words)
		}
		for w := uint32(0); w < s.words; w++ {
			a := ps2.TexAddrPSMCT32(segBase/64, 2, w%128, w/128)
			if int64(a)+4 <= int64(len(pg.vram)) {
				binary.LittleEndian.PutUint32(pg.vram[a:], u32at(obj, s.ptr+w*4))
			}
		}
		segBase += s.words
	}

	for i := 0; i < count; i++ {
		d := u32at(obj, P+0x7C+uint32(i)*4)
		if d == 0 || int64(d)+0x40 > int64(len(obj)) {
			continue // #f (a symbol cell address, far beyond the object)
		}
		mips := int(obj[d+4])
		if mips < 1 {
			mips = 1
		}
		t := Texture{
			Name:    goalString(obj, u32at(obj, d+0x24)),
			W:       int(int16(u16at(obj, d))),
			H:       int(int16(u16at(obj, d+2))),
			PSM:     int(obj[d+6]),
			ClutPSM: int(u16at(obj, d+8)),
			Mips:    mips,
			CBP:     u16at(obj, d+24),
		}
		for m := uint32(0); m < uint32(mips); m++ {
			t.TBP = append(t.TBP, u16at(obj, d+10+2*m))
			t.TBW = append(t.TBW, obj[d+26+m])
		}
		pg.Textures = append(pg.Textures, t)
	}
	return pg, nil
}

func goalString(obj []byte, p uint32) string {
	if p == 0 || int64(p)+8 > int64(len(obj)) {
		return ""
	}
	n := u32at(obj, p)
	if n > 256 || int64(p)+4+int64(n) > int64(len(obj)) {
		return ""
	}
	return string(obj[p+4 : p+4+uint32(n)])
}

func (pg *Page) vramByte(a uint32) byte {
	if int64(a) >= int64(len(pg.vram)) {
		return 0
	}
	return pg.vram[a]
}

func (pg *Page) vram32(a uint32) uint32 {
	if int64(a)+4 > int64(len(pg.vram)) {
		return 0
	}
	return binary.LittleEndian.Uint32(pg.vram[a:])
}

// clut loads the texture's CLUT the way TEX0's CLD does: n entries in CSM1
// arrangement at the descriptor's clut block.
func (pg *Page) clut(t *Texture, n uint32) []uint32 {
	out := make([]uint32, n)
	for i := uint32(0); i < n; i++ {
		x, y := ps2.CSM1ClutXY(i, n)
		switch t.ClutPSM {
		case PSMCT16:
			a := ps2.TexAddrPSMCT16(uint32(t.CBP), 1, x, y, false)
			out[i] = expand16(uint32(pg.vramByte(a)) | uint32(pg.vramByte(a+1))<<8)
		default: // PSMCT32
			out[i] = pg.vram32(ps2.TexAddrPSMCT32(uint32(t.CBP), 1, x, y))
		}
	}
	return out
}

// expand16 converts a CT16 texel to 32-bit with the TEXA reset defaults
// (TA0=0x80 for a set alpha bit, 0 clear — the engine's CLUTs carry their
// own alpha in CT32 almost everywhere; CT16 CLUTs use the plain expansion).
func expand16(px uint32) uint32 {
	r := px & 0x1F << 3
	g := px >> 5 & 0x1F << 3
	b := px >> 10 & 0x1F << 3
	a := uint32(0)
	if px&0x8000 != 0 {
		a = 0x80
	}
	return a<<24 | b<<16 | g<<8 | r
}

// Decode renders mip level m of texture t as an RGBA image. PS2 alpha
// (0x80 = opaque) is scaled to 8 bits.
func (pg *Page) Decode(t *Texture, m int) (*image.RGBA, error) {
	if m >= t.Mips {
		return nil, fmt.Errorf("tpage: %s has %d mips", t.Name, t.Mips)
	}
	w, h := t.W>>m, t.H>>m
	if w < 1 || h < 1 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("tpage: %s mip %d has size %dx%d", t.Name, m, w, h)
	}
	bp, bw := uint32(t.TBP[m]), uint32(t.TBW[m])
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	set := func(x, y int, px uint32) {
		a := px >> 24 & 0xFF * 2
		if a > 255 {
			a = 255
		}
		img.SetRGBA(x, y, color.RGBA{uint8(px), uint8(px >> 8), uint8(px >> 16), uint8(a)})
	}
	switch t.PSM {
	case PSMT8:
		cl := pg.clut(t, 256)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				set(x, y, cl[pg.vramByte(ps2.TexAddrPSMT8(bp, bw, uint32(x), uint32(y)))])
			}
		}
	case PSMT4:
		cl := pg.clut(t, 16)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				a, nib := ps2.TexAddrPSMT4(bp, bw, uint32(x), uint32(y))
				idx := uint32(pg.vramByte(a)) >> (4 * nib) & 0xF
				set(x, y, cl[idx])
			}
		}
	case PSMT8H:
		cl := pg.clut(t, 256)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				a := ps2.TexAddrPSMCT32(bp, bw, uint32(x), uint32(y))
				set(x, y, cl[pg.vramByte(a+3)])
			}
		}
	case PSMT4HL, PSMT4HH:
		cl := pg.clut(t, 16)
		shift := uint32(24)
		if t.PSM == PSMT4HH {
			shift = 28
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				px := pg.vram32(ps2.TexAddrPSMCT32(bp, bw, uint32(x), uint32(y)))
				set(x, y, cl[px>>shift&0xF])
			}
		}
	case PSMCT32, PSMCT24:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				px := pg.vram32(ps2.TexAddrPSMCT32(bp, bw, uint32(x), uint32(y)))
				if t.PSM == PSMCT24 {
					px |= 0x80000000
				}
				set(x, y, px)
			}
		}
	case PSMCT16:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				a := ps2.TexAddrPSMCT16(bp, bw, uint32(x), uint32(y), false)
				set(x, y, expand16(uint32(pg.vramByte(a))|uint32(pg.vramByte(a+1))<<8))
			}
		}
	default:
		return nil, fmt.Errorf("tpage: %s has psm 0x%02X (not handled)", t.Name, t.PSM)
	}
	return img, nil
}

// Window renders a 64x4 PSMCT32 raster window at block bp from the page
// footprint — the same view as the oracle's -gsfb, for debugging the
// transfer reconstruction against real VRAM.
func (pg *Page) Window(bp uint32) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 64, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 64; x++ {
			px := pg.vram32(ps2.TexAddrPSMCT32(bp, 1, uint32(x), uint32(y)))
			img.SetRGBA(x, y, color.RGBA{uint8(px), uint8(px >> 8), uint8(px >> 16), 255})
		}
	}
	return img
}
