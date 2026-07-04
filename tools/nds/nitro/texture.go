// Package nitro decodes the Nintendo NITRO-System 3D resource formats used by DS
// games — starting with the textures and palettes packed in an NSBTX/`BTX0` file (a
// `TEX0` block). It is graphics-format code, the DS analogue of tools/c64/gfx and
// tools/gameboy's decoders: it turns the on-cartridge bytes into Go images and makes
// no assumptions about the game.
//
// A TEX0 block holds two NITRO resource dictionaries — one naming the textures, one
// naming the palettes — plus the raw texel and palette-colour data they index into.
// The five DS texture formats supported here are the paletted and direct ones; the
// 4x4-block-compressed format (5) is recognised but not yet decoded.
package nitro

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"strings"
)

var le = binary.LittleEndian

// Texture is one decoded, named texture.
type Texture struct {
	Name   string
	Img    *image.NRGBA
	Format int
	Width  int
	Height int
}

// dictEntry is one entry of a NITRO resource dictionary: its name and its per-entry
// data bytes.
type dictEntry struct {
	name string
	data []byte
}

// parseDict decodes a NITRO resource dictionary at off. The layout (verified against
// real NSBTX files): a 4-byte header, a Patricia-tree block of 4+(count+1)*4 bytes
// (only used for name→index lookup, skipped here), a 4-byte data-block header giving
// the per-entry unit size, then count entries, then count 16-byte names at the end.
func parseDict(data []byte, off int) ([]dictEntry, error) {
	if off+4 > len(data) {
		return nil, fmt.Errorf("nitro: dict out of range")
	}
	count := int(data[off+1])
	dictSize := int(le.Uint16(data[off+2:]))
	dataHdr := off + 8 + (count+1)*4
	if dataHdr+4 > len(data) {
		return nil, fmt.Errorf("nitro: dict data header out of range")
	}
	unit := int(le.Uint16(data[dataHdr:]))
	entriesOff := dataHdr + 4
	namesOff := off + dictSize - count*16
	if namesOff+count*16 > len(data) || entriesOff+count*unit > len(data) {
		return nil, fmt.Errorf("nitro: dict entries/names out of range")
	}
	out := make([]dictEntry, count)
	for i := 0; i < count; i++ {
		name := string(data[namesOff+i*16 : namesOff+i*16+16])
		out[i] = dictEntry{
			name: strings.TrimRight(name, "\x00"),
			data: data[entriesOff+i*unit : entriesOff+i*unit+unit],
		}
	}
	return out, nil
}

// DecodeNSBTX decodes every texture in an NSBTX (`BTX0`) file, pairing each texture
// with the palette at the same index.
func DecodeNSBTX(data []byte) ([]Texture, error) {
	if len(data) < 0x14 || string(data[0:4]) != "BTX0" {
		return nil, fmt.Errorf("nitro: not a BTX0 file")
	}
	tex0 := int(le.Uint32(data[0x10:])) // first (only) block offset
	if tex0+0x3C > len(data) || string(data[tex0:tex0+4]) != "TEX0" {
		return nil, fmt.Errorf("nitro: no TEX0 block")
	}
	return decodeTEX0(data, tex0)
}

func decodeTEX0(data []byte, base int) ([]Texture, error) {
	texInfoOff := base + int(le.Uint16(data[base+0x0E:]))
	texDataOff := base + int(le.Uint32(data[base+0x14:]))
	palInfoOff := base + int(le.Uint32(data[base+0x34:]))
	palDataOff := base + int(le.Uint32(data[base+0x38:]))

	texs, err := parseDict(data, texInfoOff)
	if err != nil {
		return nil, err
	}
	pals, err := parseDict(data, palInfoOff)
	if err != nil {
		return nil, err
	}

	out := make([]Texture, 0, len(texs))
	for i, t := range texs {
		param := le.Uint32(t.data)
		addr := int(param&0xFFFF) << 3
		fmtID := int(param>>26) & 7
		w := 8 << (int(param>>20) & 7)
		h := 8 << (int(param>>23) & 7)
		color0 := param&(1<<29) != 0 // colour index 0 is transparent

		// Textures and palettes are named separately and in different orders (a texture
		// "X" uses the palette named "X_pl"), so pair them by name, not index.
		palBase := palDataOff
		if pi := matchPalette(t.name, pals); pi >= 0 {
			po := int(le.Uint32(padded(pals[pi].data)))
			if fmtID == 2 { // 4-colour palettes are addressed in 8-byte units
				palBase += (po & 0x1FFF) << 3
			} else {
				palBase += (po & 0x1FFF) << 4
			}
		} else if i < len(pals) {
			palBase += (int(le.Uint32(padded(pals[i].data))) & 0x1FFF) << 4
		}

		if fmtID == 5 {
			// 4x4-block-compressed: recognised, but its two-slot (texel / palette-index)
			// region layout is not yet reliably reversed, so skip rather than emit noise.
			continue
		}
		img, derr := decodeTexture(data, texDataOff+addr, palBase, fmtID, w, h, color0)
		if derr != nil {
			return out, fmt.Errorf("texture %q: %w", t.name, derr)
		}
		out = append(out, Texture{Name: t.name, Img: img, Format: fmtID, Width: w, Height: h})
	}
	return out, nil
}

// matchPalette finds the palette for a texture by name — "<tex>_pl", exact, or the
// same name with a trailing digit dropped, then a prefix match — returning its index
// in pals, or -1.
func matchPalette(texName string, pals []dictEntry) int {
	stripped := strings.TrimRight(texName, "0123456789")
	for _, cand := range []string{texName + "_pl", texName, stripped + "_pl", stripped + "1_pl"} {
		for i, p := range pals {
			if p.name == cand {
				return i
			}
		}
	}
	for i, p := range pals {
		if strings.HasPrefix(p.name, stripped) {
			return i
		}
	}
	return -1
}

// padded left-justifies a possibly-short dict entry into 4 bytes for a uint32 read.
func padded(b []byte) []byte {
	if len(b) >= 4 {
		return b
	}
	p := make([]byte, 4)
	copy(p, b)
	return p
}

// bgr555 converts a 15-bit BGR colour word to 8-bit-per-channel RGBA (opaque).
func bgr555(v uint16) color.NRGBA {
	exp := func(c uint16) uint8 { c &= 0x1F; return uint8(c<<3 | c>>2) }
	return color.NRGBA{R: exp(v), G: exp(v >> 5), B: exp(v >> 10), A: 0xFF}
}

func pal(data []byte, base, idx int) color.NRGBA {
	o := base + idx*2
	if o+2 > len(data) {
		return color.NRGBA{}
	}
	return bgr555(le.Uint16(data[o:]))
}

// decodeTexture decodes one texture's texels into an NRGBA image.
func decodeTexture(data []byte, texel, palBase, fmtID, w, h int, color0 bool) (*image.NRGBA, error) {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	set := func(x, y int, c color.NRGBA) { img.SetNRGBA(x, y, c) }
	n := w * h
	switch fmtID {
	case 2: // 4-colour (2bpp)
		for i := 0; i < n; i++ {
			b := data[texel+i/4]
			idx := int(b>>uint((i%4)*2)) & 3
			c := pal(data, palBase, idx)
			if idx == 0 && color0 {
				c.A = 0
			}
			set(i%w, i/w, c)
		}
	case 3: // 16-colour (4bpp)
		for i := 0; i < n; i++ {
			b := data[texel+i/2]
			idx := int(b>>uint((i%2)*4)) & 0xF
			c := pal(data, palBase, idx)
			if idx == 0 && color0 {
				c.A = 0
			}
			set(i%w, i/w, c)
		}
	case 4: // 256-colour (8bpp)
		for i := 0; i < n; i++ {
			idx := int(data[texel+i])
			c := pal(data, palBase, idx)
			if idx == 0 && color0 {
				c.A = 0
			}
			set(i%w, i/w, c)
		}
	case 1: // A3I5: 3-bit alpha, 5-bit index
		for i := 0; i < n; i++ {
			b := data[texel+i]
			c := pal(data, palBase, int(b&0x1F))
			a := (b >> 5) & 7
			c.A = uint8(a<<5 | a<<2 | a>>1)
			set(i%w, i/w, c)
		}
	case 6: // A5I3: 5-bit alpha, 3-bit index
		for i := 0; i < n; i++ {
			b := data[texel+i]
			c := pal(data, palBase, int(b&7))
			a := (b >> 3) & 0x1F
			c.A = uint8(a<<3 | a>>2)
			set(i%w, i/w, c)
		}
	case 7: // direct colour (16bpp, bit15 = alpha)
		for i := 0; i < n; i++ {
			v := le.Uint16(data[texel+i*2:])
			c := bgr555(v)
			if v&0x8000 == 0 {
				c.A = 0
			}
			set(i%w, i/w, c)
		}
	case 5:
		return nil, fmt.Errorf("4x4-compressed texture format not yet decoded")
	default:
		return nil, fmt.Errorf("unknown texture format %d", fmtID)
	}
	return img, nil
}
