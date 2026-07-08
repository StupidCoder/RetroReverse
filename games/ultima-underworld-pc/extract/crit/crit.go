// Package crit decodes Ultima Underworld's creature graphics — the CRIT/
// CRxxPAGE.Nyy "critter page" files. Each creature type (item id 64-127 →
// creature index = id-64, named in CRIT/ASSOC.ANM) has page files numbered in
// OCTAL: creature N uses CR<octal N>PAGE. A page holds the creature's sprite
// frames (8 directions × animation poses) referenced by an animation/segment
// table.
//
// Page layout (reverse-engineered from the file + the game's decoders):
//
//   - a slot directory and an animation/segment table (frame-index lists,
//     0xff-separated) whose sizes vary per creature;
//   - 32-byte auxiliary palettes ("01 00" + a 30-entry code → 8-bit-index
//     palette; index 0 is transparent);
//   - a frame count byte, a uint16 frame-offset table, then the frames.
//
// Because the header size varies, the frame table is found dynamically (see
// ParsePage). A frame is [W][H][hotX][hotY][format][uint16 count][packed codes]:
// `format` is the per-frame bit depth — 6 → 5-bit codes (game decoder
// 07F7:011C), 8 → 4-bit codes (07F7:00D9) — and `count` codes are unpacked and
// run through UW's shared run/dump RLE (07F7:018D) to W*H codes, each mapped
// through the aux palette. Verified against the game's decoder (CR00 frame 0 =
// the gray-goblin sprite; ~all creatures decode to recognisable sprites).
package crit

import (
	"encoding/binary"
	"fmt"
	"image"

	"retroreverse.com/games/ultima-underworld-pc/extract/tex"
)

// auxPalSize is one auxiliary palette record: 32 entries mapping a 5-bit code
// (0..31) to an 8-bit main-palette index. Its first two bytes are 01,00 (so
// code 1 is transparent); records are detected by that 01,00 signature.
const auxPalSize = 32

// frameHeaderLen is the per-frame header before the 5-bit data:
// W,H,hotX,hotY,auxbyte,uint16 count.
const frameHeaderLen = 7

// Page is a parsed critter page.
type Page struct {
	data     []byte
	offsets  []int    // per-frame file offsets
	aux      [][]byte // 30-entry auxiliary palettes
	segments [][]int  // per-segment frame-index cycle (ff-terminated list)
}

// ParsePage reads a CRxxPAGE.Nyy file. The header (slot directory + segment
// table + aux palettes) varies in size per creature, so the frame table is
// located dynamically: it is a count byte followed (after a pad byte) by a
// uint16 offset table whose first entry points immediately past the table and
// whose entries strictly increase and stay in-file. The aux palettes are the
// 32-byte "01 00…" records immediately before it.
func ParsePage(data []byte) (*Page, error) {
	tbl := -1
	for cand := 0x20; cand+4 < len(data); cand++ {
		n := int(data[cand])
		if n < 2 || n > 250 || cand+2+2*n > len(data) {
			continue
		}
		if int(binary.LittleEndian.Uint16(data[cand+2:])) != cand+2+2*n {
			continue
		}
		ok := true
		offs := make([]int, n)
		for i := 0; i < n; i++ {
			offs[i] = int(binary.LittleEndian.Uint16(data[cand+2+2*i:]))
			if i > 0 && offs[i] <= offs[i-1] {
				ok = false
				break
			}
			if offs[i]+frameHeaderLen > len(data) {
				ok = false
				break
			}
			if w, h := int(data[offs[i]]), int(data[offs[i]+1]); w < 3 || w > 96 || h < 3 || h > 110 {
				ok = false // every frame must be sanely sized
				break
			}
		}
		// Strong check: a frame's 5-bit data length (from its count word) plus the
		// header must match its size (next offset − this), which pins the correct
		// table and rejects look-alikes.
		if ok {
			end := len(data)
			for i := 0; i < n && ok; i++ {
				sz := end
				if i+1 < n {
					sz = offs[i+1]
				}
				sz -= offs[i]
				count := int(binary.LittleEndian.Uint16(data[offs[i]+5:]))
				want := frameHeaderLen + (count*codeBits(data[offs[i]+4])+7)/8
				if want > sz || want < sz-16 {
					ok = false
				}
			}
		}
		if ok {
			tbl = cand
			break
		}
	}
	if tbl < 0 {
		return nil, fmt.Errorf("crit: frame table not found")
	}
	n := int(data[tbl])
	p := &Page{data: data}
	for i := 0; i < n; i++ {
		p.offsets = append(p.offsets, int(binary.LittleEndian.Uint16(data[tbl+2+2*i:])))
	}
	// Aux palettes: 32-byte "01 00…" records ending just before the count byte.
	start := tbl
	for start-auxPalSize >= 0 && data[start-auxPalSize] == 1 && data[start-auxPalSize+1] == 0 {
		start -= auxPalSize
	}
	for off := start; off+auxPalSize <= tbl; off += auxPalSize {
		// The full 32 bytes are the palette (5-bit code → 8-bit main-palette
		// index); entry 1 is 0 (the transparent code) and so are the trailing
		// entries. (An earlier version dropped the leading "01 00" as a header,
		// which lost the transparent mapping and left an opaque background box.)
		p.aux = append(p.aux, data[off:off+auxPalSize])
	}
	p.parseSegments()
	return p, nil
}

// parseSegments reads the animation/segment table. data[1] points at it: after a
// little 0xff padding comes a segment-count byte, then fixed 8-byte records —
// each an 0xff-terminated list of frame indices for one animation cycle. The
// records are grouped by animation, 8 consecutive records per animation (one per
// compass view), so segments 0-7 are the eight directional views of the
// creature's primary (idle/walk) animation: record[0] is the standing frame for
// that view and the whole list is the loop to play. (Verified visually: for the
// gray goblin, segments 0-7's first frames are a clean 8-way rotation — view 0
// the back, view 4 the front — and each list is a gentle in-place idle cycle.)
func (p *Page) parseSegments() {
	if len(p.data) < 2 {
		return
	}
	seg := int(p.data[1])
	for seg < len(p.data) && p.data[seg] == 0xff {
		seg++
	}
	if seg >= len(p.data) {
		return
	}
	nseg := int(p.data[seg])
	seg++
	for s := 0; s < nseg; s++ {
		rec := seg + s*8
		if rec >= len(p.data) {
			break
		}
		var list []int
		for i := 0; i < 8 && rec+i < len(p.data); i++ {
			f := p.data[rec+i]
			if f == 0xff {
				break
			}
			fr := int(f)
			if fr >= len(p.offsets) {
				fr = 0 // out-of-range → fall back to the first frame
			}
			list = append(list, fr)
		}
		if len(list) == 0 {
			list = []int{0}
		}
		p.segments = append(p.segments, list)
	}
}

// ViewFrames returns the frame index for each of the eight compass views of the
// creature standing still: view 0 faces away from the camera (its back), view 4
// faces the camera (its front). They are the first frames of animation segments
// 0-7; a page with fewer segments falls back to frame 0 for the missing views.
// The game selects the view per frame from the object heading and the camera
// bearing (emit path 2DFE:0221, remap table DGROUP:05AC).
func (p *Page) ViewFrames() [8]int {
	var v [8]int
	for i := range v {
		if i < len(p.segments) {
			v[i] = p.segments[i][0]
		}
	}
	return v
}

// ViewCycle returns the frame-index loop for compass view `view` (segment
// `view` of the creature's primary animation) — the sequence to play for an
// idle creature facing that view. Missing views fall back to a single frame 0.
func (p *Page) ViewCycle(view int) []int {
	if view >= 0 && view < len(p.segments) {
		return p.segments[view]
	}
	return []int{0}
}

// NumFrames is the number of sprite frames in the page.
func (p *Page) NumFrames() int { return len(p.offsets) }

// Aux0 returns the page's aux palette (code -> main-palette index), or nil.
func (p *Page) Aux0() []byte {
	if len(p.aux) == 0 {
		return nil
	}
	return p.aux[0]
}

// NumAux is the number of auxiliary (colour-variant) palettes in the page.
func (p *Page) NumAux() int { return len(p.aux) }

// auxFor returns the variant-th aux palette, clamped (variant 0 = the default).
func (p *Page) auxFor(variant int) []byte {
	if len(p.aux) == 0 {
		return nil
	}
	if variant < 0 || variant >= len(p.aux) {
		variant = 0
	}
	return p.aux[variant]
}

// TranslucentIndex is the reserved main-palette index UW treats as an ADDITIVE
// "brighten the background" marker rather than a literal colour. The ethereal
// creatures (ghost/wisp/fire/shadow) draw their glow with it; the game shifts
// the framebuffer pixels there into a brighter range. Frames that use it are
// split (see FrameLayers) so the viewer renders this layer additively and every
// other index — the ghost's black eyes, fire's flames, the shadow's body —
// normally.
const TranslucentIndex = 252

// Translucent reports whether this page's sprites use TranslucentIndex.
func (p *Page) Translucent() bool {
	for _, m := range p.Aux0() {
		if m == TranslucentIndex {
			return true
		}
	}
	return false
}

// FrameLayers decodes frame i into two layers split by palette index: normal
// holds every opaque pixel whose index is NOT TranslucentIndex (drawn normally),
// additive holds the TranslucentIndex pixels in their palette colour (drawn
// additively). additive is nil when the frame has none. Index 0 is transparent
// in both. This matches the game's rule "the berry index is additive, the others
// are not" without any per-pixel heuristics.
func (p *Page) FrameLayers(i, variant int, pal tex.Palette) (normal, additive *image.RGBA, err error) {
	w, h, codes, err := p.FrameCodes(i)
	if err != nil {
		return nil, nil, err
	}
	ap := p.auxFor(variant)
	if len(ap) == 0 {
		return nil, nil, fmt.Errorf("crit: page has no aux palette")
	}
	normal = image.NewRGBA(image.Rect(0, 0, w, h))
	set := func(img *image.RGBA, k int, r, g, b byte) {
		img.Pix[k*4+0], img.Pix[k*4+1], img.Pix[k*4+2], img.Pix[k*4+3] = r, g, b, 255
	}
	for k := 0; k < w*h; k++ {
		var idx byte
		if c := codes[k]; int(c) < len(ap) {
			idx = ap[c]
		}
		switch {
		case idx == TranslucentIndex:
			if additive == nil {
				additive = image.NewRGBA(image.Rect(0, 0, w, h))
			}
			// The game BRIGHTENS the framebuffer pixels here (a hue-neutral shift
			// to a brighter range), so the additive layer is WHITE — not the berry
			// placeholder pal[252]. The viewer scales this via the material colour.
			set(additive, k, 255, 255, 255)
		case idx != 0:
			c := pal[idx]
			set(normal, k, c.R, c.G, c.B)
		}
	}
	return normal, additive, nil
}

// FrameCodes returns frame i's raw per-pixel codes (before the aux/main palette
// lookup) and its dimensions. Code 0 is the sprite's transparent background.
func (p *Page) FrameCodes(i int) (w, h int, codes []byte, err error) {
	if i < 0 || i >= len(p.offsets) {
		return 0, 0, nil, fmt.Errorf("crit: frame %d out of range (%d)", i, len(p.offsets))
	}
	o := p.offsets[i]
	if o+frameHeaderLen > len(p.data) {
		return 0, 0, nil, fmt.Errorf("crit: frame %d header outside page", i)
	}
	w, h = int(p.data[o]), int(p.data[o+1])
	format := p.data[o+4]
	count := int(binary.LittleEndian.Uint16(p.data[o+5:]))
	data := p.data[o+frameHeaderLen:]
	var cs []byte
	switch format {
	case 8:
		cs = make([]byte, 0, count)
		for bi := 0; bi < len(data) && len(cs) < count; bi++ {
			cs = append(cs, data[bi]>>4, data[bi]&0xF)
		}
		if len(cs) > count {
			cs = cs[:count]
		}
	default:
		cs = unpack5(data, count)
	}
	return w, h, tex.RLEExpand(cs, w*h), nil
}

// Frame decodes frame i to an RGBA image (transparent where the code is 0),
// using the page's aux palette (index 0 — the frames all share one page palette)
// mapped through the given main palette.
func (p *Page) Frame(i, variant int, pal tex.Palette) (*image.RGBA, error) {
	if i < 0 || i >= len(p.offsets) {
		return nil, fmt.Errorf("crit: frame %d out of range (%d)", i, len(p.offsets))
	}
	o := p.offsets[i]
	if o+frameHeaderLen > len(p.data) {
		return nil, fmt.Errorf("crit: frame %d header outside page", i)
	}
	w, h := int(p.data[o]), int(p.data[o+1])
	format := p.data[o+4]
	count := int(binary.LittleEndian.Uint16(p.data[o+5:]))
	data := p.data[o+frameHeaderLen:]
	var codes []byte
	switch format {
	case 8: // 4-bit codes
		codes = make([]byte, 0, count)
		for bi := 0; bi < len(data) && len(codes) < count; bi++ {
			codes = append(codes, data[bi]>>4, data[bi]&0xF)
		}
		if len(codes) > count {
			codes = codes[:count]
		}
	default: // format 6: 5-bit codes
		codes = unpack5(data, count)
	}
	px := tex.RLEExpand(codes, w*h)
	ap := p.auxFor(variant)
	if len(ap) == 0 {
		return nil, fmt.Errorf("crit: page has no aux palette")
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for k := 0; k < w*h; k++ {
		var idx byte
		if c := px[k]; int(c) < len(ap) {
			idx = ap[c]
		}
		c := pal[idx]
		if idx == 0 {
			c.A = 0
		} else {
			c.A = 255
		}
		img.Pix[k*4+0], img.Pix[k*4+1], img.Pix[k*4+2], img.Pix[k*4+3] = c.R, c.G, c.B, c.A
	}
	return img, nil
}

// codeBits maps a frame's format byte (header[4]) to its bits per code: format
// 6 = 5-bit (07F7:011C), format 8 = 4-bit (00D9), format 4 = 8-bit raw.
func codeBits(format byte) int {
	switch format {
	case 8:
		return 4
	case 4:
		return 8
	default:
		return 5
	}
}

// unpack5 reads count 5-bit codes MSB-first from a bitstream (the game's
// 07F7:011C unpacker: ROR extracts the top 5 bits of each byte first).
func unpack5(data []byte, count int) []byte {
	out := make([]byte, count)
	bit := 0
	for i := 0; i < count; i++ {
		var v int
		for b := 0; b < 5; b++ {
			by := bit >> 3
			if by >= len(data) {
				break
			}
			v = v<<1 | int(data[by]>>(7-(bit&7))&1)
			bit++
		}
		out[i] = byte(v)
	}
	return out
}
