package n3ds

import (
	"encoding/binary"
	"fmt"
	"image"
)

// cgfx_texture.go decodes a CGFX texture object (TXOB) to RGBA. Two layers:
// the TXOB header (dimensions, the OpenGL-DMP format/type enum pair, and the
// pointer into the IMAG block), and the PICA pixel layouts underneath — the
// 8×8-pixel Morton-ordered ("Z-order") tiles every uncompressed 3DS texture is
// swizzled into, and ETC1 block compression for the compressed ones.
//
// TXOB layout (offsets from the magic word), verified on all four banner
// textures by dimension×bpp arithmetic (each dataSize field equals
// width×height×bpp/8 exactly):
//
//	+0x08 name ptr        +0x14 height      +0x18 width
//	+0x1C GL format       +0x20 GL type     +0x24 mipmap levels
//	+0x40 data size       +0x44 data ptr    +0x4C bits per pixel
//
// The GL enums are the 3DS's DMP extensions; the pairs seen so far:
//
//	(0x675A, 0)      ETC1 compressed RGB
//	(0x6757, 0x6761) LUMINANCE, UNSIGNED_4BITS — L4, 4-bit grayscale
type TXOB struct {
	Name          string
	Width, Height int
	GLFormat      uint32
	GLType        uint32
	BPP           int
	dataOff       int64
	dataLen       int64
}

// DecodeTexture parses the TXOB a Textures-dictionary entry points at and
// decodes its pixels to an NRGBA image.
func (c *CGFX) DecodeTexture(e CGFXEntry) (*TXOB, *image.NRGBA, error) {
	M := e.Offset + 4
	if string(c.raw[M:M+4]) != "TXOB" {
		return nil, nil, fmt.Errorf("cgfx: texture entry %q is not a TXOB (magic %q)", e.Name, c.raw[M:M+4])
	}
	t := &TXOB{
		Name:     e.Name,
		Height:   int(c.u32(M + 0x14)),
		Width:    int(c.u32(M + 0x18)),
		GLFormat: c.u32(M + 0x1C),
		GLType:   c.u32(M + 0x20),
		BPP:      int(c.u32(M + 0x4C)),
		dataLen:  int64(c.u32(M + 0x40)),
		dataOff:  c.rel(M + 0x44),
	}
	if want := int64(t.Width) * int64(t.Height) * int64(t.BPP) / 8; want != t.dataLen {
		return nil, nil, fmt.Errorf("cgfx: texture %q: %dx%d at %d bpp needs 0x%x bytes but header says 0x%x",
			t.Name, t.Width, t.Height, t.BPP, want, t.dataLen)
	}
	if err := c.check(t.dataOff, t.dataLen, "texture data"); err != nil {
		return nil, nil, err
	}
	img, err := t.decode(c.raw[t.dataOff : t.dataOff+t.dataLen])
	if err != nil {
		return nil, nil, fmt.Errorf("cgfx: texture %q: %w", t.Name, err)
	}
	return t, img, nil
}

func (t *TXOB) decode(data []byte) (*image.NRGBA, error) {
	switch {
	case t.GLFormat == 0x675A: // ETC1
		return decodeETC1(data, t.Width, t.Height), nil
	case t.GLFormat == 0x6757 && t.GLType == 0x6761: // L4
		return decodeTiled4(data, t.Width, t.Height, func(v byte) [4]byte {
			g := v * 17
			return [4]byte{g, g, g, 255}
		}), nil
	case t.GLType == 0x1401 && t.GLFormat == 0x6752: // RGBA8 native
		return decodeTiledN(data, t.Width, t.Height, 4, func(p []byte) [4]byte {
			return [4]byte{p[3], p[2], p[1], p[0]} // stored ABGR
		}), nil
	case t.GLType == 0x8034: // RGBA5551
		return decodeTiledN(data, t.Width, t.Height, 2, func(p []byte) [4]byte {
			v := binary.LittleEndian.Uint16(p)
			r := byte(v>>11) & 31
			g := byte(v>>6) & 31
			b := byte(v>>1) & 31
			a := byte(v&1) * 255
			return [4]byte{r<<3 | r>>2, g<<3 | g>>2, b<<3 | b>>2, a}
		}), nil
	case t.GLType == 0x8363: // RGB565
		return decodeTiledN(data, t.Width, t.Height, 2, func(p []byte) [4]byte {
			v := binary.LittleEndian.Uint16(p)
			r := byte(v>>11) & 31
			g := byte(v>>5) & 63
			b := byte(v) & 31
			return [4]byte{r<<3 | r>>2, g<<2 | g>>4, b<<3 | b>>2, 255}
		}), nil
	case t.GLType == 0x8033: // RGBA4
		return decodeTiledN(data, t.Width, t.Height, 2, func(p []byte) [4]byte {
			v := binary.LittleEndian.Uint16(p)
			return [4]byte{
				byte(v>>12) * 17, byte(v>>8&15) * 17, byte(v>>4&15) * 17, byte(v&15) * 17,
			}
		}), nil
	}
	return nil, fmt.Errorf("unsupported GL format/type pair (0x%04x, 0x%04x)", t.GLFormat, t.GLType)
}

// mortonXY returns the pixel coordinate inside an 8×8 tile for index i in the
// tile's Morton (Z-order) sequence: bits interleave as x0,y0,x1,y1,x2,y2.
func mortonXY(i int) (int, int) {
	x := (i & 1) | (i>>1)&2 | (i>>2)&4
	y := (i>>1)&1 | (i>>2)&2 | (i>>3)&4
	return x, y
}

// decodeTiledN decodes an uncompressed texture of bpp bytes per pixel, laid out
// as raster-ordered 8×8 tiles with Morton order inside each tile.
func decodeTiledN(data []byte, w, h, bpp int, px func([]byte) [4]byte) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	p := 0
	for ty := 0; ty < h; ty += 8 {
		for tx := 0; tx < w; tx += 8 {
			for i := 0; i < 64; i++ {
				x, y := mortonXY(i)
				c := px(data[p : p+bpp])
				p += bpp
				o := img.PixOffset(tx+x, ty+y)
				copy(img.Pix[o:o+4], c[:])
			}
		}
	}
	return img
}

// decodeTiled4 decodes a 4-bit-per-pixel tiled texture (two pixels per byte,
// low nibble first in Morton order).
func decodeTiled4(data []byte, w, h int, px func(byte) [4]byte) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	n := 0 // nibble index
	for ty := 0; ty < h; ty += 8 {
		for tx := 0; tx < w; tx += 8 {
			for i := 0; i < 64; i++ {
				x, y := mortonXY(i)
				b := data[n/2]
				v := b & 15
				if n%2 == 1 {
					v = b >> 4
				}
				n++
				c := px(v)
				o := img.PixOffset(tx+x, ty+y)
				copy(img.Pix[o:o+4], c[:])
			}
		}
	}
	return img
}

// etc1Modifiers is the ETC1 modifier table: per codeword, the small and large
// luminance offsets (the sign comes from the pixel index's MSB).
var etc1Modifiers = [8][2]int32{
	{2, 8}, {5, 17}, {9, 29}, {13, 42}, {18, 60}, {24, 80}, {33, 106}, {47, 183},
}

// decodeETC1 decodes a 3DS ETC1 texture: raster-ordered 8×8 tiles, each holding
// four 4×4 ETC1 blocks in Z-order — (0,0), (4,0), (0,4), (4,4) — and each block
// stored as one little-endian u64 whose big-endian byte view is the standard
// ETC1 block encoding.
func decodeETC1(data []byte, w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	p := 0
	for ty := 0; ty < h; ty += 8 {
		for tx := 0; tx < w; tx += 8 {
			for _, sub := range [4][2]int{{0, 0}, {4, 0}, {0, 4}, {4, 4}} {
				v := binary.LittleEndian.Uint64(data[p:])
				p += 8
				decodeETC1Block(img, tx+sub[0], ty+sub[1], v)
			}
		}
	}
	return img
}

// decodeETC1Block expands one 4×4 ETC1 block at (bx, by). Bit layout per the
// ETC1 specification, taking the u64 as its big-endian byte sequence b0..b7:
// b0..b2 base colours, b3 = table codes | diff | flip, b4..b7 the 2-bit pixel
// indices (MSB half then LSB half, pixels in column-major order).
func decodeETC1Block(img *image.NRGBA, bx, by int, v uint64) {
	b := [8]byte{}
	for i := range b {
		b[i] = byte(v >> (56 - 8*i))
	}
	diff := b[3]&2 != 0
	flip := b[3]&1 != 0
	table0 := int(b[3] >> 5)
	table1 := int(b[3] >> 2 & 7)

	var base [2][3]int32
	if diff {
		for ch := 0; ch < 3; ch++ {
			c5 := int32(b[ch] >> 3)
			d3 := int32(b[ch] & 7)
			if d3 >= 4 {
				d3 -= 8
			}
			c2 := c5 + d3
			base[0][ch] = c5<<3 | c5>>2
			base[1][ch] = c2<<3 | c2>>2
		}
	} else {
		for ch := 0; ch < 3; ch++ {
			c0 := int32(b[ch] >> 4)
			c1 := int32(b[ch] & 15)
			base[0][ch] = c0*16 + c0
			base[1][ch] = c1*16 + c1
		}
	}

	msb := uint32(b[4])<<8 | uint32(b[5])
	lsb := uint32(b[6])<<8 | uint32(b[7])
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			i := uint(x*4 + y) // column-major pixel numbering
			sub := 0
			if (!flip && x >= 2) || (flip && y >= 2) {
				sub = 1
			}
			table := table0
			if sub == 1 {
				table = table1
			}
			mag := etc1Modifiers[table][lsb>>i&1]
			if msb>>i&1 != 0 {
				mag = -mag
			}
			px, py := bx+x, by+y
			if px >= img.Rect.Max.X || py >= img.Rect.Max.Y {
				continue
			}
			o := img.PixOffset(px, py)
			for ch := 0; ch < 3; ch++ {
				c := base[sub][ch] + mag
				if c < 0 {
					c = 0
				}
				if c > 255 {
					c = 255
				}
				img.Pix[o+ch] = uint8(c)
			}
			img.Pix[o+3] = 255
		}
	}
}
