package gc

// gpu_texture.go is the TX unit: it turns a texture coordinate into a colour by reading the
// texel out of main memory and decoding it from the packed, tiled format the hardware stores.
//
// A GameCube texture is not a linear array of pixels. It is broken into small rectangular
// tiles — 4x4, 8x4 or 8x8 texels depending on the format — and the tiles are laid down one
// after another in memory, each tile's texels in row-major order inside it. That tiling is
// what lets the hardware's texture cache fetch a neighbourhood of texels in one burst, and it
// is the first thing a decoder has to undo: the byte that holds texel (x,y) is found by
// working out which tile it falls in, then where in that tile it sits. Every format in this
// file shares that structure and differs only in its tile size and how many bits it spends
// per texel.
//
// The texel data itself stays in main memory; the hardware streams it through the on-chip
// texture cache on demand. Modelling that cache is a performance concern, not a correctness
// one, so this file reads straight from main memory at the base address the game programmed
// (TX_SETIMAGE3) and decodes on the spot. The formats are hardware facts from the Flipper
// datasheet, the same class as the CPU's instruction encodings; the decode is reimplemented
// here rather than copied, and the gate that keeps it honest (gpu_test.go) builds a texture
// with an independent encoder and asserts the TX unit reads back exactly what went in.
//
// The paletted formats (C4/C8/C14X2) index a colour table the hardware DMAs into texture
// memory ahead of the draw. loadTlut follows that DMA (BP 0x64 names the source, BP 0x65
// triggers the load) and tlutColor decodes an entry; the register layouts are pinned from
// the game's own GXInitTlutObj and GXLoadTlut rather than from a wiki. NOTE that no draw
// has yet bound one — the formats are implemented against the library code that programs
// them, not against a frame that uses them, so the TX_SETTLUT register numbers here are the
// least-verified thing in this file.

import (
	"fmt"
	"image"
	"os"
)

// The texture format codes, as they appear in the TX_SETIMAGE0 register's format field.
const (
	texI4     = 0x0
	texI8     = 0x1
	texIA4    = 0x2
	texIA8    = 0x3
	texRGB565 = 0x4
	texRGB5A3 = 0x5
	texRGBA8  = 0x6
	texC4     = 0x8
	texC8     = 0x9
	texC14X2  = 0xA
	texCMPR   = 0xE
)

// texState is one texture map's configuration, read back from the BP registers the game
// programmed (TX_SETMODE0/TX_SETIMAGE0/TX_SETIMAGE3, and TX_SETTLUT for a paletted format).
type texState struct {
	format        int
	width, height int
	base          uint32 // physical address of the texel data in main memory
	wrapS, wrapT  int
	tlutOff       int // offset of this map's palette within the TMEM high bank
	tlutFmt       int // palette entry format: 0 IA8, 1 RGB565, 2 RGB5A3

	// tlutRAM aims the palette at main memory instead of at texture memory, and tlutOff
	// becomes an offset from it. A draw's palette is in TMEM because the game DMA'd it
	// there; a debugger pointing a free surface at an address has only the copy still in
	// RAM, and this is how it says so without a second decoder existing to disagree with
	// this one.
	tlutRAM     uint32
	tlutFromRAM bool
}

// texSetup reads texture map i's configuration. The eight maps are split across two register
// banks: maps 0..3 at 0x80/0x88/0x94, maps 4..7 at the mirrored 0xA0/0xA8/0xB4.
func (g *gpu) texSetup(i int) texState {
	var mode0, image0, image3, settlut uint32
	if i < 4 {
		mode0 = g.BP[0x80+uint32(i)]
		image0 = g.BP[0x88+uint32(i)]
		image3 = g.BP[0x94+uint32(i)]
		settlut = g.BP[0x98+uint32(i)]
	} else {
		mode0 = g.BP[0xA0+uint32(i-4)]
		image0 = g.BP[0xA8+uint32(i-4)]
		image3 = g.BP[0xB4+uint32(i-4)]
		settlut = g.BP[0xB8+uint32(i-4)]
	}
	return texState{
		format:  int((image0 >> 20) & 0xF),
		width:   int(image0&0x3FF) + 1,
		height:  int((image0>>10)&0x3FF) + 1,
		base:    (image3 & 0x00FFFFFF) << 5,
		wrapS:   int(mode0 & 3),
		wrapT:   int((mode0 >> 2) & 3),
		tlutOff: int(settlut&0x3FF) << 9,
		tlutFmt: int((settlut >> 10) & 3),
	}
}

// loadTlut is the BP 0x65 trigger: copy a palette out of main memory into the TMEM high
// bank, where paletted draws will index it. The source address was staged in BP 0x64.
func (g *gpu) loadTlut(m *Machine, data uint32) {
	src := phys((g.BP[0x64] & 0x00FFFFFF) << 5)
	off := int(data&0x3FF) << 9
	n := int((data>>10)&0x7FF) * 32 // 16 entries of 2 bytes per line
	if g.Tlut == nil {
		g.Tlut = make([]byte, 0x80000)
	}
	if off+n > len(g.Tlut) || int(src)+n > len(m.RAM) {
		m.CPU.Halt("TLUT load out of range: src 0x%08X, tmem offset 0x%X, %d bytes", src, off, n)
		return
	}
	copy(g.Tlut[off:off+n], m.RAM[src:int(src)+n])
	if drawTrace {
		fmt.Fprintf(os.Stderr, "TLUT load 0x%08X -> tmem+0x%05X %d bytes\n", src, off, n)
	}
}

// tlutColor decodes palette entry idx of the TLUT at tlutOff. The entry formats are the
// same three the texture formats share: IA8, RGB565, RGB5A3.
func (g *gpu) tlutColor(m *Machine, tx texState, idx int) (r, gg, b, a uint8) {
	var v uint16
	if tx.tlutFromRAM {
		o := phys(tx.tlutRAM + uint32(tx.tlutOff) + uint32(idx)*2)
		if int(o)+1 >= len(m.RAM) {
			return 0, 0, 0, 0
		}
		v = uint16(m.RAM[o])<<8 | uint16(m.RAM[o+1])
	} else {
		o := tx.tlutOff + idx*2
		if g.Tlut == nil || o+1 >= len(g.Tlut) {
			return 0, 0, 0, 0
		}
		v = uint16(g.Tlut[o])<<8 | uint16(g.Tlut[o+1])
	}
	switch tx.tlutFmt {
	case 0: // IA8: intensity in the low byte, alpha in the high — as the IA8 texture format
		i := uint8(v & 0xFF)
		return i, i, i, uint8(v >> 8)
	case 1:
		return decodeRGB565(v)
	case 2:
		return decodeRGB5A3(v)
	default:
		m.CPU.Halt("TLUT: entry format %d is not a hardware format", tx.tlutFmt)
		return 0, 0, 0, 0
	}
}

// sampleTexmap samples a texture map at a normalised coordinate (s,t in [0,1]), returning the
// texel's colour. The coordinate is scaled to the texture's size, wrapped, and point-sampled;
// bilinear filtering is a later refinement and is a named gap here, exact for the 1:1
// full-screen blits the boot renders and only softening minified textures otherwise.
//
// The map's configuration arrives decoded (gpu_tev_state.go) rather than being read back out
// of the BP registers here. It used to be rebuilt per texel, which is per sample of per stage
// of per fragment, and it cannot change while a draw runs.
func (g *gpu) sampleTexmap(m *Machine, tx *texState, s, t float32) (r, gg, b, a uint8) {
	x := wrapCoord(int(s*float32(tx.width)), tx.width, tx.wrapS)
	y := wrapCoord(int(t*float32(tx.height)), tx.height, tx.wrapT)
	return g.decodeTexel(m, *tx, x, y)
}

// wrapCoord brings a texel coordinate into range by the addressing mode: clamp holds it at the
// edge, repeat wraps it around, mirror folds it back and forth.
func wrapCoord(v, size, mode int) int {
	if size <= 0 {
		return 0
	}
	switch mode {
	case 1: // repeat
		v %= size
		if v < 0 {
			v += size
		}
		return v
	case 2: // mirror
		p := size * 2
		v %= p
		if v < 0 {
			v += p
		}
		if v >= size {
			v = p - 1 - v
		}
		return v
	default: // clamp
		if v < 0 {
			return 0
		}
		if v >= size {
			return size - 1
		}
		return v
	}
}

// decodeTexel reads and decodes the single texel (x,y) of a texture. It dispatches on the
// format to that format's tile geometry and unpacking.
func (g *gpu) decodeTexel(m *Machine, tx texState, x, y int) (r, gg, b, a uint8) {
	switch tx.format {
	case texI4:
		v := g.tileNibble(m, tx.base, tx.width, 8, 8, x, y)
		i := expand4(uint16(v))
		return i, i, i, i
	case texI8:
		v := g.tileByte(m, tx.base, tx.width, 8, 4, x, y)
		return v, v, v, v
	case texIA4:
		v := g.tileByte(m, tx.base, tx.width, 8, 4, x, y)
		i := expand4(uint16(v & 0x0F))
		al := expand4(uint16(v >> 4))
		return i, i, i, al
	case texIA8:
		v := g.tileHalf(m, tx.base, tx.width, 4, 4, x, y)
		i := uint8(v & 0xFF)
		al := uint8(v >> 8)
		return i, i, i, al
	case texRGB565:
		v := g.tileHalf(m, tx.base, tx.width, 4, 4, x, y)
		return decodeRGB565(v)
	case texRGB5A3:
		v := g.tileHalf(m, tx.base, tx.width, 4, 4, x, y)
		return decodeRGB5A3(v)
	case texRGBA8:
		return g.decodeRGBA8(m, tx.base, tx.width, x, y)
	case texCMPR:
		return g.decodeCMPR(m, tx.base, tx.width, x, y)
	case texC4: // 8x8 tile of nibble indices into the bound TLUT
		v := g.tileNibble(m, tx.base, tx.width, 8, 8, x, y)
		return g.tlutColor(m, tx, int(v))
	case texC8: // 8x4 tile of byte indices
		v := g.tileByte(m, tx.base, tx.width, 8, 4, x, y)
		return g.tlutColor(m, tx, int(v))
	case texC14X2: // 4x4 tile of halfwords, the low 14 bits the index
		v := g.tileHalf(m, tx.base, tx.width, 4, 4, x, y)
		return g.tlutColor(m, tx, int(v&0x3FFF))
	default:
		m.CPU.Halt("TX: unknown texture format 0x%X", tx.format)
		return 0, 0, 0, 0
	}
}

// tileByteOffset returns the offset, from a texture's base, of the tile that holds texel (x,y),
// plus the texel's index within that tile. bw/bh are the tile's texel dimensions.
func tileByteOffset(width, bw, bh, x, y int) (tileBytesBase, inTile int) {
	tilesPerRow := (width + bw - 1) / bw
	bx, by := x/bw, y/bh
	ix, iy := x%bw, y%bh
	tileIdx := by*tilesPerRow + bx
	texelsPerTile := bw * bh
	return tileIdx * texelsPerTile, iy*bw + ix
}

// tileHalf, tileByte and tileNibble read one 16-, 8- or 4-bit texel from a tiled texture whose
// texels are that many bits wide and laid out row-major within each tile.
func (g *gpu) tileHalf(m *Machine, base uint32, width, bw, bh, x, y int) uint16 {
	tb, in := tileByteOffset(width, bw, bh, x, y)
	off := base + uint32((tb+in)*2)
	return uint16(m.texByte(off))<<8 | uint16(m.texByte(off+1))
}

func (g *gpu) tileByte(m *Machine, base uint32, width, bw, bh, x, y int) uint8 {
	tb, in := tileByteOffset(width, bw, bh, x, y)
	return m.texByte(base + uint32(tb+in))
}

func (g *gpu) tileNibble(m *Machine, base uint32, width, bw, bh, x, y int) uint8 {
	tb, in := tileByteOffset(width, bw, bh, x, y)
	b := m.texByte(base + uint32((tb+in)/2))
	if in&1 == 0 { // the high nibble is the earlier texel
		return b >> 4
	}
	return b & 0x0F
}

// decodeRGBA8 reads a texel of the 32-bit RGBA8 format, whose 4x4 tile is stored as two
// 32-byte halves: the first holds the sixteen texels' alpha/red pairs, the second their
// green/blue pairs.
func (g *gpu) decodeRGBA8(m *Machine, base uint32, width, x, y int) (r, gg, b, a uint8) {
	tb, in := tileByteOffset(width, 4, 4, x, y)
	tileBase := base + uint32(tb*4) // 4 bytes per texel
	a = m.texByte(tileBase + uint32(in*2) + 0)
	r = m.texByte(tileBase + uint32(in*2) + 1)
	gg = m.texByte(tileBase + 32 + uint32(in*2) + 0)
	b = m.texByte(tileBase + 32 + uint32(in*2) + 1)
	return r, gg, b, a
}

// decodeCMPR reads a texel of the S3TC/DXT1-style compressed format. Its 8x8 tile is four 4x4
// sub-blocks of eight bytes each — two big-endian RGB565 endpoint colours followed by four
// bytes of two-bit indices, one row of four texels per byte, the leftmost texel in the high
// bits. When the first endpoint sorts above the second the block is opaque with two
// interpolated colours; otherwise the fourth index is transparent, the format's punch-through
// alpha.
func (g *gpu) decodeCMPR(m *Machine, base uint32, width, x, y int) (r, gg, b, a uint8) {
	// Locate the 8x8 tile, then the 4x4 sub-block within it.
	tilesPerRow := (width + 7) / 8
	bx, by := x/8, y/8
	tileIdx := by*tilesPerRow + bx
	subX, subY := (x%8)/4, (y%8)/4
	subIdx := subY*2 + subX
	blockBase := base + uint32(tileIdx*32+subIdx*8)

	c0 := uint16(m.texByte(blockBase))<<8 | uint16(m.texByte(blockBase+1))
	c1 := uint16(m.texByte(blockBase+2))<<8 | uint16(m.texByte(blockBase+3))
	ix, iy := x%4, y%4
	bits := m.texByte(blockBase + 4 + uint32(iy))
	idx := (bits >> (6 - uint(ix)*2)) & 3

	r0, g0, b0, _ := decodeRGB565(c0)
	r1, g1, b1, _ := decodeRGB565(c1)
	switch idx {
	case 0:
		return r0, g0, b0, 255
	case 1:
		return r1, g1, b1, 255
	case 2:
		if c0 > c1 {
			return lerp2(r0, r1), lerp2(g0, g1), lerp2(b0, b1), 255
		}
		return avg2(r0, r1), avg2(g0, g1), avg2(b0, b1), 255
	default: // idx == 3
		if c0 > c1 {
			return lerp2(r1, r0), lerp2(g1, g0), lerp2(b1, b0), 255
		}
		return 0, 0, 0, 0 // the transparent index
	}
}

// decodeRGB565 and decodeRGB5A3 unpack the two 16-bit colour formats. RGB565 is opaque; RGB5A3
// spends its top bit on a mode: set means five-bit opaque RGB, clear means three-bit alpha over
// four-bit RGB.
func decodeRGB565(v uint16) (r, g, b, a uint8) {
	return expand5(v >> 11), expand6((v >> 5) & 0x3F), expand5(v & 0x1F), 255
}

func decodeRGB5A3(v uint16) (r, g, b, a uint8) {
	if v&0x8000 != 0 {
		return expand5((v >> 10) & 0x1F), expand5((v >> 5) & 0x1F), expand5(v & 0x1F), 255
	}
	return expand4((v >> 8) & 0xF), expand4((v >> 4) & 0xF), expand4(v & 0xF), expand3((v >> 12) & 0x7)
}

func expand3(v uint16) uint8 { return uint8(v<<5 | v<<2 | v>>1) }

// lerp2 and avg2 are the two interpolated endpoints CMPR needs: one-third of the way and the
// midpoint between two channel values.
func lerp2(a, b uint8) uint8 { return uint8((uint16(a)*2 + uint16(b)) / 3) }
func avg2(a, b uint8) uint8  { return uint8((uint16(a) + uint16(b)) / 2) }

// texByte reads one byte of texture data from main memory, masking to the RAM size so a
// texture whose programmed base or size runs past the end of memory reads zero at the overrun
// rather than panicking — a wild texture is a bug to see, not a crash.
func (m *Machine) texByte(addr uint32) uint8 {
	a := phys(addr)
	if int(a) >= len(m.RAM) {
		return 0
	}
	return m.RAM[a]
}

// TextureBound reports whether texture map i has an image programmed into it.
//
// The question has to be asked rather than inferred from the size, because an unbound map's
// registers read back as zeros and the size fields are stored biased by one — so a map the
// game has never touched decodes as a perfectly plausible 1x1 texture at address 0 rather
// than as nothing at all. A caller enumerating the eight maps and trusting the size gets six
// phantom textures.
func (m *Machine) TextureBound(i int) bool {
	return m.gpu.texSetup(i).base != 0
}

// DumpTexture decodes texture map i in full, at the size and format the game currently has it
// bound, into an RGBA image — the direct proof of the texture unit, independent of whatever the
// TEV then does with the samples. It is how the loading image the boot binds is seen even
// though that frame's combiner discards the texture's colour.
func (m *Machine) DumpTexture(i int) (*image.RGBA, error) {
	tx := m.gpu.texSetup(i)
	if tx.width <= 0 || tx.height <= 0 {
		return nil, fmt.Errorf("texture %d has no size (the game has not bound it)", i)
	}
	img := image.NewRGBA(image.Rect(0, 0, tx.width, tx.height))
	for y := 0; y < tx.height; y++ {
		for x := 0; x < tx.width; x++ {
			r, gg, b, a := m.gpu.decodeTexel(m, tx, x, y)
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, gg, b, a
		}
	}
	return img, nil
}
