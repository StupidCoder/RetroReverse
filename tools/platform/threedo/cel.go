package threedo

// cel.go decodes the 3DO "cel" image format produced by the Madam cel engine.
// A .cel file is a sequence of IFF-like chunks — an 8-byte header (4-char tag +
// big-endian size that includes the header) followed by the payload:
//
//	CCB   the Cel Control Block: flags, geometry, and the two preamble words
//	      PRE0/PRE1 that encode bits-per-pixel and the image dimensions.
//	PLUT  the Pixel Look-Up Table: up to 32 big-endian RGB555 colors that a
//	      "coded" cel's pixel values index into. Absent for "uncoded" cels.
//	PDAT  the pixel data, either unpacked (fixed bits/pixel, row-strided) or
//	      packed (per-line RLE packets). XTRA/ANIM chunks are art-tool metadata.
//
// The pixel encoding is read straight from the CCB, validated against the Need
// for Speed disc: PRE0 bits 2:0 select bpp (3=4bpp, 6=16bpp on this disc),
// bits 15:6 hold height-1, PRE1 bits 10:0 hold width-1, and CCB flag bit 9
// (CCB_PACKED) selects packed vs unpacked source data.

import (
	"fmt"
	"image"
	"image/color"
)

// Cel is a parsed 3DO cel ready to decode into an image.
type Cel struct {
	Flags         uint32
	PRE0, PRE1    uint32
	PIXC          uint32 // the two PPMP pixel-processor words
	Width, Height int
	BPP           int      // bits per pixel: 1,2,4,6,8,16
	Packed        bool     // packed (RLE) vs unpacked source data
	Coded         bool     // pixels index the PLUT (vs literal RGB555)
	PLUT          []uint16 // RGB555 palette entries (coded cels)
	PDAT          []byte   // pixel source data
}

// bppFromPRE0 maps the 3-bit PRE0 bpp field to bits per pixel.
var bppFromPRE0 = map[uint32]int{1: 1, 2: 2, 3: 4, 4: 6, 5: 8, 6: 16}

// CCB flag bit 9: packed source data. (ccbCCBPre and the other CCB flag bits
// are defined in graphicsfolio.go, same package.)
const ccbPacked = 0x00000200

// ParseCel walks the chunk stream and extracts the fields needed to decode.
func ParseCel(data []byte) (*Cel, error) {
	c := &Cel{}
	haveCCB := false
	for p := 0; p+8 <= len(data); {
		tag := string(data[p : p+4])
		size := int(be32(data[p+4:]))
		if size < 8 || p+size > len(data) {
			break // truncated or padding; stop at the last good chunk
		}
		body := data[p+8 : p+size]
		switch tag {
		case "CCB ":
			// A second CCB begins the NEXT cel — in a track packet the cels sit
			// back-to-back, so without this stop the walk marched on and
			// returned the last cel's header over the first cel's data (the
			// exported textures came out as another mip's dimensions: garbled).
			if haveCCB {
				return c.finish()
			}
			if len(body) < 0x48 {
				return nil, fmt.Errorf("threedo: CCB chunk too small (%d)", len(body))
			}
			// body[0] is ccbversion; fields below are relative to it.
			c.Flags = be32(body[0x04:])
			c.PIXC = be32(body[0x34:])
			c.PRE0 = be32(body[0x38:])
			c.PRE1 = be32(body[0x3C:])
			c.Width = int(be32(body[0x40:]))
			c.Height = int(be32(body[0x44:]))
			c.Packed = c.Flags&ccbPacked != 0
			haveCCB = true
		case "PLUT":
			if len(body) < 4 {
				break
			}
			n := int(be32(body[0:]))
			for i := 0; i < n && 4+(i+1)*2 <= len(body); i++ {
				c.PLUT = append(c.PLUT, uint16(body[4+i*2])<<8|uint16(body[4+i*2+1]))
			}
		case "PDAT":
			c.PDAT = body
		}
		p += size
	}
	if !haveCCB {
		return nil, fmt.Errorf("threedo: no CCB chunk (not a cel?)")
	}
	return c.finish()
}

// finish derives the decode parameters once the cel's own chunks are in.
func (c *Cel) finish() (*Cel, error) {
	// Without CCB_CCBPRE the preamble words lead the source data (PRE0 always,
	// PRE1 only for unpacked cels — packed lines carry their own offsets), and
	// the hardware reads them from there. The struct's own PRE fields may hold
	// a stale copy, so the in-stream words are authoritative — and they MUST be
	// consumed: leaving them in place shifts the packed line-offset chain four
	// bytes and desyncs the whole image (the truncated roadside billboard whose
	// first "line offset" was really the preamble byte 0x80).
	if c.Flags&ccbCCBPre == 0 && len(c.PDAT) >= 4 {
		c.PRE0 = be32(c.PDAT)
		c.PDAT = c.PDAT[4:]
		if c.Flags&ccbPacked == 0 && len(c.PDAT) >= 4 {
			c.PRE1 = be32(c.PDAT)
			c.PDAT = c.PDAT[4:]
		}
	}
	c.BPP = bppFromPRE0[c.PRE0&0x7]
	if c.BPP == 0 {
		return nil, fmt.Errorf("threedo: unknown bpp code %d", c.PRE0&0x7)
	}
	// The preamble is the authoritative source of the cel's dimensions — the
	// hardware never reads the C struct's ccb_Width/Height (the rule the
	// software cel engine learned from the in-race "noise columns"). Fall back
	// to the struct fields only when the preamble is degenerate.
	if w := int(c.PRE1&0x7FF) + 1; w > 1 || c.Width == 0 {
		c.Width = w
	}
	if h := int((c.PRE0>>6)&0x3FF) + 1; h > 1 || c.Height == 0 {
		c.Height = h
	}
	// A cel is coded when its pixels are <=8-bit indices into a PLUT; 16-bit
	// cels carry literal RGB555 and load no PLUT.
	c.Coded = c.BPP <= 8
	return c, nil
}

// RGB555 expands a 15-bit 3DO color (bit 15 is the P/W control bit, ignored)
// to an opaque RGBA color.
func RGB555(v uint16) color.RGBA { return rgb555(v) }

// rgb555 expands a 15-bit 3DO color (bit 15 is the P/W control bit, ignored) to
// an opaque RGBA color.
func rgb555(v uint16) color.RGBA {
	r := uint8((v >> 10) & 0x1F)
	g := uint8((v >> 5) & 0x1F)
	b := uint8(v & 0x1F)
	return color.RGBA{r<<3 | r>>2, g<<3 | g>>2, b<<3 | b>>2, 0xFF}
}

// bitReader reads big-endian (MSB-first) bit fields from a byte slice.
type bitReader struct {
	data []byte
	pos  int // bit index
}

func (br *bitReader) read(n int) uint32 {
	var v uint32
	for i := 0; i < n; i++ {
		bytIdx := br.pos >> 3
		if bytIdx >= len(br.data) {
			br.pos++
			v <<= 1
			continue
		}
		bit := 7 - (br.pos & 7)
		v = v<<1 | uint32((br.data[bytIdx]>>uint(bit))&1)
		br.pos++
	}
	return v
}

// Packed-cel packet types (the top 2 bits of each packet's control field).
const (
	packEOL         = 0
	packLiteral     = 1
	packTransparent = 2
	packRepeat      = 3
)

// Image decodes the cel into an RGBA image, the way the cel engine's PDEC and
// PPMP stages resolve it:
//
//   - Transparent runs, and any pixel whose resolved color is RGB black on a
//     cel without CCB_BGND, become fully transparent (the PDEC rule — this is
//     what cuts out lamp posts and building billboards from their black
//     backgrounds; the road cels set BGND, so their black texels stay).
//   - Each pixel's P-bit (bit 5 at 6bpp, the PLUT entry's bit 15 otherwise)
//     selects one of the CCB's two PPMP words, and the destination-independent
//     part of that word — first-source multiply/divide, constant second
//     source, final shift — is applied to the color. The track's road cels
//     carry PIXC 0x1F001F01: P=0 texels are halved (the dark asphalt), P=1
//     texels pass through (the lane lines) — without this the export showed
//     the raw over-bright palette with white speckles the game never draws.
func (c *Cel) Image() (*image.RGBA, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 4096 || c.Height > 4096 {
		return nil, fmt.Errorf("threedo: implausible cel size %dx%d", c.Width, c.Height)
	}
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	bgnd := c.Flags&ccbBGND != 0
	set := func(x, y int, v uint32) {
		if x < 0 || x >= c.Width {
			return
		}
		var raw uint16 // RGB555 with the P-bit in bit 15
		if c.Coded {
			// The PDEC's palette index is the low 5 bits; at 6bpp bit 5 is the
			// P-bit and at 8bpp the top bits are the AMV shade (not applied
			// here) — neither is part of the index (opera_madam.c cp6btag/
			// cp8btag).
			idx := v
			if c.BPP >= 6 {
				idx &= 0x1F
			}
			if int(idx) < len(c.PLUT) {
				raw = c.PLUT[idx]
				if c.BPP == 6 {
					raw = raw&0x7FFF | uint16(v&0x20)<<10 // P from the pixel
				}
			} else if len(c.PLUT) == 0 {
				// No palette: render the index as grayscale so the shape shows.
				g := uint8(v * 255 / uint32((1<<c.BPP)-1))
				img.SetRGBA(x, y, color.RGBA{g, g, g, 0xFF})
				return
			}
		} else { // uncoded 16bpp literal RGB555
			raw = uint16(v)
		}
		if raw&0x7FFF == 0 && !bgnd {
			return // resolved black without CCB_BGND: transparent (PDEC)
		}
		img.SetRGBA(x, y, rgb555(c.ppmp(raw)))
	}

	if c.Packed {
		c.decodePacked(set)
	} else {
		c.decodeUnpacked(set)
	}
	return img, nil
}

// ppmp applies the destination-independent part of the pixel processor to one
// RGB555 pixel (P-bit in bit 15): the P-bit or the CCB's POVER override picks
// one of PIXC's two PPMP words; the word's first-source multiply/divide,
// constant second source and final shift then scale the color (blendPixel in
// graphicsfolio.go is the full engine). Words that need the framebuffer
// (first source = destination, AMV/self-modulate multipliers, second source =
// destination/source) pass the color through unchanged — a flat texture
// export has no destination to blend with.
func (c *Cel) ppmp(pix uint16) uint16 {
	word := c.PIXC & 0xFFFF
	switch c.Flags & ccbPOVER {
	case 0x100: // PMODE_ZERO: force the low word
	case 0x180: // PMODE_ONE: force the high word
		word = c.PIXC >> 16
	default: // the pixel's own P-bit picks the word
		if pix&0x8000 != 0 {
			word = c.PIXC >> 16
		}
	}
	if word == 0 { // no PIXC recorded (e.g. a synthetic cel): pass through
		return pix
	}
	s1 := word&0x8000 != 0
	ms := (word >> 13) & 3
	mxf := (word >> 10) & 7
	dv1 := (word >> 8) & 3
	s2 := (word >> 6) & 3
	avf := (word >> 1) & 0x1F
	dv2 := word & 1
	if s1 || ms != 0 || s2 >= 2 {
		return pix // needs the destination (or a shade value); pass through
	}
	var second uint32
	if s2 == 1 {
		dv3 := uint32(0)
		if c.Flags&0x400 != 0 { // CCB_USEAV: AV carries control signals
			dv3 = (avf >> 3) & 3
			avf &^= 0x1C
		}
		second = avf >> dv3
	}
	sh1 := ((dv1 - 1) & 3) + 1 // PDV
	scale := func(ch uint32) uint16 {
		v := ((ch*(mxf+1))>>sh1 + second) >> dv2
		if v > 31 {
			v = 31
		}
		return uint16(v)
	}
	r := scale((uint32(pix) >> 10) & 0x1F)
	g := scale((uint32(pix) >> 5) & 0x1F)
	b := scale(uint32(pix) & 0x1F)
	return r<<10 | g<<5 | b
}

// decodePacked walks the per-line RLE stream. Each line begins with a byte
// (bpp<=8) or two bytes (16bpp) giving the offset in words to the next line,
// minus two; the packet bitstream follows, word-aligned per line.
func (c *Cel) decodePacked(set func(x, y int, v uint32)) {
	// The per-line word-offset preamble is 1 byte for cels below 8bpp and 2
	// bytes at 8bpp and above (Madam packed-cel format). Using the wrong width
	// desyncs every line: 8bpp packed cels — NFS's sky, scenery and dashboard —
	// decoded to almost nothing before this boundary was corrected.
	offBytes := 1
	if c.BPP >= 8 {
		offBytes = 2
	}
	pos := 0
	for y := 0; y < c.Height; y++ {
		if pos+offBytes > len(c.PDAT) {
			break
		}
		lineStart := pos
		var words int
		if offBytes == 1 {
			words = int(c.PDAT[lineStart]) + 2
		} else {
			words = (int(c.PDAT[lineStart])<<8 | int(c.PDAT[lineStart+1])) + 2
		}
		br := &bitReader{data: c.PDAT, pos: (lineStart + offBytes) * 8}
		for x := 0; x < c.Width; {
			typ := br.read(2)
			if typ == packEOL {
				break
			}
			count := int(br.read(6)) + 1
			switch typ {
			case packLiteral:
				for i := 0; i < count && x < c.Width; i++ {
					set(x, y, br.read(c.BPP))
					x++
				}
			case packTransparent:
				x += count
			case packRepeat:
				v := br.read(c.BPP)
				for i := 0; i < count && x < c.Width; i++ {
					set(x, y, v)
					x++
				}
			}
		}
		next := lineStart + words*4
		if next <= lineStart {
			break
		}
		pos = next
	}
}

// decodeUnpacked reads fixed-width pixels row by row. The row stride comes
// from PRE1's WOFFSET field — (offset+2) words per row, in bits 24-31 below
// 8bpp and bits 16-25 at 8/16bpp (opera_madam.c) — falling back to the
// minimal word-rounded row when the preamble carries no offset.
func (c *Cel) decodeUnpacked(set func(x, y int, v uint32)) {
	var woffset int
	if c.BPP >= 8 {
		woffset = int((c.PRE1 >> 16) & 0x3FF)
	} else {
		woffset = int((c.PRE1 >> 24) & 0xFF)
	}
	strideBits := (woffset + 2) * 32
	if minBits := ((c.Width*c.BPP + 31) / 32) * 32; strideBits < minBits {
		strideBits = minBits
	}
	for y := 0; y < c.Height; y++ {
		br := &bitReader{data: c.PDAT, pos: y * strideBits}
		for x := 0; x < c.Width; x++ {
			set(x, y, br.read(c.BPP))
		}
	}
}
