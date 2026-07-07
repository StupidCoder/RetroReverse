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
	Width, Height int
	BPP           int      // bits per pixel: 1,2,4,6,8,16
	Packed        bool     // packed (RLE) vs unpacked source data
	Coded         bool     // pixels index the PLUT (vs literal RGB555)
	PLUT          []uint16 // RGB555 palette entries (coded cels)
	PDAT          []byte   // pixel source data
}

// bppFromPRE0 maps the 3-bit PRE0 bpp field to bits per pixel.
var bppFromPRE0 = map[uint32]int{1: 1, 2: 2, 3: 4, 4: 6, 5: 8, 6: 16}

const ccbPacked = 0x00000200 // CCB flag bit 9: packed source data

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
			if len(body) < 0x48 {
				return nil, fmt.Errorf("threedo: CCB chunk too small (%d)", len(body))
			}
			// body[0] is ccbversion; fields below are relative to it.
			c.Flags = be32(body[0x04:])
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
	c.BPP = bppFromPRE0[c.PRE0&0x7]
	if c.BPP == 0 {
		return nil, fmt.Errorf("threedo: unknown bpp code %d", c.PRE0&0x7)
	}
	// Prefer the CCB geometry; fall back to the preamble words if it is blank.
	if c.Width == 0 {
		c.Width = int(c.PRE1&0x7FF) + 1
	}
	if c.Height == 0 {
		c.Height = int((c.PRE0>>6)&0x3FF) + 1
	}
	// A cel is coded when its pixels are <=8-bit indices into a PLUT; 16-bit
	// cels carry literal RGB555 and load no PLUT.
	c.Coded = c.BPP <= 8
	return c, nil
}

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

// Image decodes the cel into an RGBA image. Transparent runs (and, for uncoded
// cels, a zero color word) become fully transparent pixels.
func (c *Cel) Image() (*image.RGBA, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 4096 || c.Height > 4096 {
		return nil, fmt.Errorf("threedo: implausible cel size %dx%d", c.Width, c.Height)
	}
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	set := func(x, y int, v uint32) {
		if x < 0 || x >= c.Width {
			return
		}
		var col color.RGBA
		if c.Coded {
			if int(v) < len(c.PLUT) {
				col = rgb555(c.PLUT[v])
			} else if len(c.PLUT) == 0 {
				// No palette: render the index as grayscale so the shape shows.
				g := uint8(v * 255 / uint32((1<<c.BPP)-1))
				col = color.RGBA{g, g, g, 0xFF}
			}
		} else { // uncoded 16bpp literal RGB555
			if uint16(v) == 0 {
				return // treat the zero word as transparent
			}
			col = rgb555(uint16(v))
		}
		img.SetRGBA(x, y, col)
	}

	if c.Packed {
		c.decodePacked(set)
	} else {
		c.decodeUnpacked(set)
	}
	return img, nil
}

// decodePacked walks the per-line RLE stream. Each line begins with a byte
// (bpp<=8) or two bytes (16bpp) giving the offset in words to the next line,
// minus two; the packet bitstream follows, word-aligned per line.
func (c *Cel) decodePacked(set func(x, y int, v uint32)) {
	offBytes := 1
	if c.BPP > 8 {
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

// decodeUnpacked reads fixed-width pixels row by row. The row stride is the
// pixel row rounded up to a whole number of 32-bit words.
func (c *Cel) decodeUnpacked(set func(x, y int, v uint32)) {
	strideBits := ((c.Width*c.BPP + 31) / 32) * 32
	for y := 0; y < c.Height; y++ {
		br := &bitReader{data: c.PDAT, pos: y * strideBits}
		for x := 0; x < c.Width; x++ {
			set(x, y, br.read(c.BPP))
		}
	}
}
