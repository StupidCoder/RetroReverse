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
	raw  []byte // the segments' bytes concatenated in file order (the atlas)
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
	// The page's upload primitive is upload-vram-data (0x6157FC): every
	// transfer is a 128-px-wide PSMCT32 IMAGE (BITBLTBUF DBW=2, TRXREG
	// RRW=128) whose source is consumed at 256 bytes per destination block,
	// in block order. Under that model the concatenated segments form the
	// boot-time page image; VERIFIED byte-exact against boot VRAM for
	// segment 0 — the always-resident far tier, which carries every CLUT
	// and every texture's deeper mips — and pixel-exact through the
	// oracle's sampler for full textures whose mip 0 also sits there.
	// The mid/near tiers (segments 1-2) stream through update-vram-pages'
	// 16 KB chunk allocator with a per-chunk residency table (s2+180), so
	// their live VRAM placement is dynamic and their positions inside the
	// file atlas do not follow the header's dst deltas (seg1 is not even a
	// whole number of 512-byte atlas rows). OPEN: reimplement the packer
	// (texture-page-default-allocate) to place tier-1/2 mip 0 data exactly;
	// mip-consistency (the game's mip chain is an exact 2x2 box filter)
	// validates any placement without the oracle, and FitTiers below
	// searches with it.
	pg.vram = make([]byte, pg.Words*4)
	pg.raw = make([]byte, 0, pg.Words*4)
	for _, s := range segs {
		if int64(s.ptr)+int64(s.words)*4 <= int64(len(obj)) {
			pg.raw = append(pg.raw, obj[s.ptr:s.ptr+s.words*4]...)
		}
	}
	w := uint32(0)
	for _, s := range segs {
		if int64(s.ptr)+int64(s.words)*4 > int64(len(obj)) {
			return nil, fmt.Errorf("tpage %s: segment [0x%x +0x%x words] outside object", pg.Name, s.ptr, s.words)
		}
		for i := uint32(0); i < s.words; i++ {
			a := ps2.TexAddrPSMCT32(0, 2, w%128, w/128)
			if int64(a)+4 <= int64(len(pg.vram)) {
				binary.LittleEndian.PutUint32(pg.vram[a:], u32at(obj, s.ptr+i*4))
			}
			w++
		}
	}

	for i := 0; i < count; i++ {
		d := u32at(obj, P+0x7C+uint32(i)*4)
		if d == 0 || int64(d)+0x40 > int64(len(obj)) {
			// #f (a symbol cell address, far beyond the object): keep the
			// slot so texture INDICES stay aligned — ids address the table
			// by position.
			pg.Textures = append(pg.Textures, Texture{})
			continue
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

// --- tier fitting -----------------------------------------------------------
//
// The mid/near tiers stream through a chunk allocator, so a tier texture's
// data does not sit at its descriptor's destination block in the file's
// atlas. But the game's mip chain is an exact 2x2 box filter and every
// texture's deepest mip lives in tier 0 (dense, verified), so the true atlas
// position of each shallower mip is recoverable by search: the position
// whose decode box-downscales onto the verified next-deeper mip.

// FitResult reports one mip's search.
type FitResult struct {
	Searched bool
	OK       bool
	Offset   uint32 // byte offset of the texel data in the raw atlas
	X, Y     int    // position in the 512-byte-wide atlas raster
	Err      float64
}

// boxErr scores how well img box-downscales onto ref (mean abs RGB diff).
func boxErr(img, ref *image.RGBA) float64 {
	w, h := ref.Rect.Dx(), ref.Rect.Dy()
	if img.Rect.Dx() != w*2 || img.Rect.Dy() != h*2 {
		return 1e18
	}
	sum, n := 0.0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b int
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					c := img.RGBAAt(x*2+dx, y*2+dy)
					r += int(c.R)
					g += int(c.G)
					b += int(c.B)
				}
			}
			pc := ref.RGBAAt(x, y)
			sum += abs3(r/4-int(pc.R)) + abs3(g/4-int(pc.G)) + abs3(b/4-int(pc.B))
			n += 3
		}
	}
	return sum / float64(n)
}

// texelAt reads texel (x,y) of a texture's mip as a CLUT index straight from
// the raw atlas at byte position (offset, 512-wide raster).
func (pg *Page) atlasIndex(t *Texture, x, y int) (uint32, bool) {
	if x < 0 || y < 0 {
		return 0, false
	}
	switch t.PSM {
	case PSMT8:
		a := y*512 + x
		if a >= len(pg.raw) {
			return 0, false
		}
		return uint32(pg.raw[a]), true
	case PSMT4:
		a := y*512 + x/2
		if a >= len(pg.raw) {
			return 0, false
		}
		v := uint32(pg.raw[a])
		if x&1 != 0 {
			v >>= 4
		}
		return v & 0xF, true
	}
	return 0, false
}

// FitTiers solves the atlas positions of every mip of t that does not
// already decode consistently. Returns one FitResult per mip level.
func (pg *Page) FitTiers(t *Texture) []FitResult {
	res := make([]FitResult, t.Mips)
	// Reference chain: start from the deepest mip via the block model
	// (tier 0), then walk up, fitting each level against the one below.
	ref, err := pg.Decode(t, t.Mips-1)
	if err != nil {
		return res
	}
	for m := t.Mips - 2; m >= 0; m-- {
		// try the descriptor's own block first (dense tier-0 case)
		if img, err := pg.Decode(t, m); err == nil && boxErr(img, ref) < 1.0 {
			ref = img
			continue
		}
		r := pg.fitOne(t, m, ref)
		res[m] = r
		if !r.OK {
			break // can't anchor shallower mips
		}
		img := pg.decodeAtAtlas(t, m, r.Offset)
		ref = img
	}
	return res
}

func (pg *Page) fitOne(t *Texture, m int, ref *image.RGBA) FitResult {
	w, h := t.W>>m, t.H>>m
	cl := pg.clutFor(t)
	best := FitResult{Searched: true, Err: 1e18}
	rows := len(pg.raw) / 512
	stepX := 32 // observed placements sit on 32-texel columns
	if w >= 512 {
		stepX = 512
	}
	for y := 0; y+h <= rows; y++ {
		for x := 0; x+w <= 512; x += stepX {
			e := pg.tryFit(t, uint32(y), uint32(x), w, h, cl, ref)
			if e < best.Err {
				best.Err = e
				best.X, best.Y = x, y
				best.Offset = uint32(y)*512 + uint32(x)
				if t.PSM == PSMT4 {
					best.Offset = uint32(y)*512 + uint32(x/2)
				}
				if e == 0 {
					best.OK = true
					return best
				}
			}
		}
	}
	best.OK = best.Err < 2.0
	return best
}

// tryFit scores a candidate position: decode w×h from the atlas, box-filter
// 2x2 in RGB, mean abs diff against the reference mip.
func (pg *Page) tryFit(t *Texture, y0, x0 uint32, w, h int, cl []uint32, ref *image.RGBA) float64 {
	sum, n := 0.0, 0
	rw := ref.Rect.Dx()
	// sample a sparse grid first for speed; full check only if promising
	for _, full := range []bool{false, true} {
		step := 4
		if full {
			step = 1
			if sum/float64(max(n, 1)) > 8.0 {
				break
			}
			sum, n = 0, 0
		}
		for ry := 0; ry < h/2; ry += step {
			for rx := 0; rx < w/2; rx += step {
				var r, g, b int
				ok := true
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						idx, in := pg.atlasIndex(t, int(x0)+rx*2+dx, int(y0)+ry*2+dy)
						if !in {
							ok = false
							break
						}
						c := cl[idx]
						r += int(c & 0xFF)
						g += int(c >> 8 & 0xFF)
						b += int(c >> 16 & 0xFF)
					}
				}
				if !ok || ry >= ref.Rect.Dy() || rx >= rw {
					continue
				}
				pc := ref.RGBAAt(rx, ry)
				sum += abs3(r/4-int(pc.R)) + abs3(g/4-int(pc.G)) + abs3(b/4-int(pc.B))
				n += 3
			}
		}
	}
	if n == 0 {
		return 1e18
	}
	return sum / float64(n)
}

func abs3(v int) float64 {
	if v < 0 {
		return float64(-v)
	}
	return float64(v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// decodeAtAtlas decodes mip m of t from a solved atlas position.
func (pg *Page) decodeAtAtlas(t *Texture, m int, off uint32) *image.RGBA {
	w, h := t.W>>m, t.H>>m
	cl := pg.clutFor(t)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	y0 := int(off / 512)
	x0 := int(off % 512)
	if t.PSM == PSMT4 {
		x0 *= 2
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx, _ := pg.atlasIndex(t, x0+x, y0+y)
			c := cl[idx]
			a := c >> 24 & 0xFF * 2
			if a > 255 {
				a = 255
			}
			img.SetRGBA(x, y, color.RGBA{uint8(c), uint8(c >> 8), uint8(c >> 16), uint8(a)})
		}
	}
	return img
}

// clutFor returns the texture's CLUT (from the verified tier-0 region).
func (pg *Page) clutFor(t *Texture) []uint32 {
	n := uint32(256)
	if t.PSM == PSMT4 || t.PSM == PSMT4HL || t.PSM == PSMT4HH {
		n = 16
	}
	return pg.clut(t, n)
}
