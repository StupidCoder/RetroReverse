package n64

// surface.go renders a region of memory as an image.
//
// This is what a debugger points at the address a DMA just landed on, to see whether
// what arrived there is the texture the game meant to upload. The pixel formats are the
// platform's business — the decoders are right here, next to the rasteriser that uses
// them — so a debugger says "draw me this address as CI8 with the palette over there"
// and does not have to know what CI8 is.

import (
	"fmt"
	"image"
	"image/color"
)

// RegionFormats are the pixel formats RenderRegion understands: the N64's own texture
// formats, in the "type/size" naming the RDP uses.
func RegionFormats() []string {
	return []string{"rgba16", "rgba32", "ia16", "ia8", "ia4", "i8", "i4", "ci8", "ci4"}
}

// RegionSpec aims RenderRegion at memory.
type RegionSpec struct {
	Addr   uint32 // physical RDRAM address of the top-left texel
	W, H   int    // size in texels
	Stride int    // bytes per row; 0 means "tightly packed", W texels wide
	Format string // one of RegionFormats
	// Palette is the physical address of a TLUT — 256 RGBA16 entries — used by the
	// indexed formats (ci8, ci4). For ci4 it is the base of the 16-entry sub-palette.
	Palette uint32
}

// bitsPerTexel is the size of one texel in bits, for the packed formats.
func bitsPerTexel(format string) (int, error) {
	switch format {
	case "rgba32":
		return 32, nil
	case "rgba16", "ia16":
		return 16, nil
	case "ia8", "i8", "ci8":
		return 8, nil
	case "ia4", "i4", "ci4":
		return 4, nil
	}
	return 0, fmt.Errorf("n64: unknown pixel format %q (have %v)", format, RegionFormats())
}

// RenderRegion decodes W×H texels of RDRAM as an image. Addresses outside RDRAM read as
// zero rather than failing: a debugger aimed at a wrong address should show you rubbish,
// which is informative, not an error message.
func (m *Machine) RenderRegion(s RegionSpec) (*image.RGBA, error) {
	bits, err := bitsPerTexel(s.Format)
	if err != nil {
		return nil, err
	}
	if s.W <= 0 || s.H <= 0 {
		return nil, fmt.Errorf("n64: a %dx%d region has nothing to show", s.W, s.H)
	}
	if s.W*s.H > 4<<20 {
		return nil, fmt.Errorf("n64: a %dx%d region is too large to render", s.W, s.H)
	}
	stride := s.Stride
	if stride == 0 {
		stride = (s.W*bits + 7) / 8
	}

	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	for y := 0; y < s.H; y++ {
		row := s.Addr + uint32(y*stride)
		for x := 0; x < s.W; x++ {
			img.Set(x, y, m.texelOf(s, row, x, bits))
		}
	}
	return img, nil
}

// texelOf decodes one texel at column x of the row starting at addr.
func (m *Machine) texelOf(s RegionSpec, addr uint32, x, bits int) color.RGBA {
	switch s.Format {
	case "rgba32":
		a := addr + uint32(x*4)
		return color.RGBA{m.ram(a), m.ram(a + 1), m.ram(a + 2), m.ram(a + 3)}
	case "rgba16":
		c := fromRGBA16(m.ram16(addr + uint32(x*2)))
		return color.RGBA{uint8(c.R), uint8(c.G), uint8(c.B), uint8(c.A)}
	case "ia16":
		// Intensity in the high byte, alpha in the low one.
		v := m.ram16(addr + uint32(x*2))
		i := uint8(v >> 8)
		return color.RGBA{i, i, i, uint8(v)}
	case "ia8":
		// Four bits of intensity, four of alpha.
		v := m.ram(addr + uint32(x))
		i := (v >> 4) * 17 // 0..15 → 0..255
		return color.RGBA{i, i, i, (v & 15) * 17}
	case "ia4":
		// Three bits of intensity, one of alpha.
		v := nibble(m.ram(addr+uint32(x/2)), x)
		i := (v >> 1) * 36 // 0..7 → 0..252
		a := uint8(0)
		if v&1 != 0 {
			a = 255
		}
		return color.RGBA{i, i, i, a}
	case "i8":
		v := m.ram(addr + uint32(x))
		return color.RGBA{v, v, v, v}
	case "i4":
		v := nibble(m.ram(addr+uint32(x/2)), x) * 17
		return color.RGBA{v, v, v, v}
	case "ci8":
		c := fromRGBA16(m.ram16(s.Palette + uint32(m.ram(addr+uint32(x)))*2))
		return color.RGBA{uint8(c.R), uint8(c.G), uint8(c.B), uint8(c.A)}
	case "ci4":
		i := nibble(m.ram(addr+uint32(x/2)), x)
		c := fromRGBA16(m.ram16(s.Palette + uint32(i)*2))
		return color.RGBA{uint8(c.R), uint8(c.G), uint8(c.B), uint8(c.A)}
	}
	return color.RGBA{}
}

// nibble picks the high or low half of a byte, high first — texel 0 is the high nibble.
func nibble(b byte, x int) uint8 {
	if x%2 == 0 {
		return b >> 4
	}
	return b & 15
}

// ram reads a byte of RDRAM, out of range reading as zero.
func (m *Machine) ram(addr uint32) byte {
	if int(addr) < len(m.RDRAM) {
		return m.RDRAM[addr]
	}
	return 0
}

func (m *Machine) ram16(addr uint32) uint16 {
	return uint16(m.ram(addr))<<8 | uint16(m.ram(addr+1))
}

// TMEM is the RDP's 4 KiB of texture memory — where a texture actually lives once the
// game has loaded it, as opposed to the copy in RDRAM it was loaded from. It is
// returned raw: the RDP stores textures interleaved (odd rows word-swapped) and a tile
// descriptor says how to read them back, so rendering it as a flat image would be a
// picture of the interleaving rather than of the texture.
func (m *Machine) TMEM() []byte { return m.rdp.TMem[:] }
