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
	tex4x4Off := base + int(le.Uint32(data[base+0x24:])) // 4x4-format texel data
	palIdxOff := base + int(le.Uint32(data[base+0x28:])) // 4x4-format palette-index data
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
		// "X" uses the palette named "X_pl"), so pair them by name, not index. The
		// dictionary's offset word is in 8-byte units (<<3) — pinned by the last
		// palettes of a real set landing exactly inside the palette region.
		palBase := palDataOff
		if pi := matchPalette(t.name, pals); pi >= 0 {
			palBase += (int(le.Uint16(padded(pals[pi].data))) & 0x1FFF) << 3
		} else if i < len(pals) {
			palBase += (int(le.Uint16(padded(pals[i].data))) & 0x1FFF) << 3
		}

		var img *image.NRGBA
		if fmtID == 5 {
			// 4x4-block-compressed: the texture's address indexes the dedicated 4x4 texel
			// region (TEX0+$24); its per-block palette words live at half that offset in
			// the palette-index region (TEX0+$28). Verified: in a real TEX0 the regions
			// tile exactly — texels, 4x4 texels, 4x4 indices (half size), palettes.
			img = decode4x4(data, tex4x4Off+addr, palIdxOff+addr/2, palBase, w, h)
		} else {
			var derr error
			if img, derr = decodeTexture(data, texDataOff+addr, palBase, fmtID, w, h, color0); derr != nil {
				return out, fmt.Errorf("texture %q: %w", t.name, derr)
			}
		}
		out = append(out, Texture{Name: t.name, Img: img, Format: fmtID, Width: w, Height: h})
	}
	return out, nil
}

// matchPalette finds the palette for a texture by name. Real sets are erratic —
// "X"↔"X_pl", but also dropped prefixes ("nr_road5"↔"road5_pl"), moved underscores
// ("nr_dash_02"↔"nr_dash2") and renames ("nr_start_line2"↔"nr_line_pl") — so this
// scores each candidate: exact match, then containment, then a common-suffix /
// common-substring measure. (The authoritative binding is the material block of the
// NSBMD model that consumes the texture; this is the best a texture set alone gives.)
func matchPalette(texName string, pals []dictEntry) int {
	t := paletteCore(texName)
	best, bestScore := -1, 2 // require a minimal similarity
	for i, p := range pals {
		c := paletteCore(p.name)
		var s int
		switch {
		case c == t:
			s = 1000
		case strings.Contains(t, c):
			s = 500 + len(c)
		case strings.Contains(c, t):
			s = 500 + len(t)
		default:
			suf := commonSuffix(strings.TrimRight(t, "0123456789"), strings.TrimRight(c, "0123456789"))
			sub := commonSubstring(t, c)
			s = sub
			if 2*suf > s {
				s = 2 * suf
			}
		}
		if s > bestScore {
			best, bestScore = i, s
		}
	}
	return best
}

// paletteCore normalises a dictionary name for matching: lower-case, "_pl" dropped.
func paletteCore(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, "_pl")
	return strings.TrimRight(s, "_")
}

func commonSuffix(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[len(a)-1-n] == b[len(b)-1-n] {
		n++
	}
	return n
}

// commonSubstring returns the length of the longest common substring.
func commonSubstring(a, b string) int {
	best := 0
	prev := make([]int, len(b)+1)
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > best {
					best = cur[j]
				}
			}
		}
		prev = cur
	}
	return best
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

// decode4x4 decodes a format-5 (4x4-block-compressed) texture: a grid of 4x4-pixel
// blocks, each with 16 two-bit texel indices (one 32-bit word) and a 16-bit palette
// word — bits 0-13 a sub-palette offset (in 4-byte steps, relative to the texture's
// palette), bits 14-15 a mode selecting how the four 2-bit values map to colours
// (two of them may be interpolations, one may be transparent).
func decode4x4(data []byte, texel, palIdx, palBase, w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	bw := w / 4
	for by := 0; by < h/4; by++ {
		for bx := 0; bx < bw; bx++ {
			blk := by*bw + bx
			if texel+blk*4+4 > len(data) || palIdx+blk*2+2 > len(data) {
				continue
			}
			texels := le.Uint32(data[texel+blk*4:])
			info := le.Uint16(data[palIdx+blk*2:])
			cbase := palBase + int(info&0x3FFF)*4
			cols := blockColors(data, cbase, info>>14)
			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					v := (texels >> uint((py*4+px)*2)) & 3
					img.SetNRGBA(bx*4+px, by*4+py, cols[v])
				}
			}
		}
	}
	return img
}

// blockColors builds the four colours a 4x4 block's 2-bit values select, per its mode.
func blockColors(data []byte, cbase int, mode uint16) [4]color.NRGBA {
	c0, c1 := pal(data, cbase, 0), pal(data, cbase, 1)
	switch mode {
	case 0: // c0, c1, c2, transparent
		return [4]color.NRGBA{c0, c1, pal(data, cbase, 2), {}}
	case 1: // c0, c1, (c0+c1)/2, transparent
		return [4]color.NRGBA{c0, c1, mix(c0, c1, 1, 1), {}}
	case 2: // c0, c1, c2, c3
		return [4]color.NRGBA{c0, c1, pal(data, cbase, 2), pal(data, cbase, 3)}
	default: // 3: c0, c1, (5c0+3c1)/8, (3c0+5c1)/8
		return [4]color.NRGBA{c0, c1, mix(c0, c1, 5, 3), mix(c0, c1, 3, 5)}
	}
}

// mix blends two colours with integer weights (result opaque).
func mix(a, b color.NRGBA, wa, wb int) color.NRGBA {
	t := wa + wb
	return color.NRGBA{
		R: uint8((int(a.R)*wa + int(b.R)*wb) / t),
		G: uint8((int(a.G)*wa + int(b.G)*wb) / t),
		B: uint8((int(a.B)*wa + int(b.B)*wb) / t),
		A: 0xFF,
	}
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
