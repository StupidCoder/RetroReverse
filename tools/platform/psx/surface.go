package psx

// surface.go renders a region of VRAM as an image.
//
// On the PlayStation everything the GPU can see lives in one 1024×512 16-bit frame
// buffer: the displayed picture, the back buffer, the texture pages and the palettes,
// all at once. So "view memory as an image" here means "view a rectangle of VRAM", and
// the interesting part is that a texture page is *not* 16-bit colour — it is 4- or
// 8-bit indices packed into those 16-bit words, looked up in a palette that is itself a
// row of VRAM pixels. Reading it back needs the same decode the rasteriser does, which
// is why this lives here rather than in the debugger.

import (
	"fmt"
	"image"
	"image/color"
)

// RegionFormats are the ways a rectangle of VRAM can be read.
func RegionFormats() []string { return []string{"rgba5551", "clut8", "clut4"} }

// RegionSpec aims RenderRegion at VRAM. X and Y are in VRAM pixels (0..1023, 0..511);
// W and H count *texels*, which for the indexed formats is not the same as VRAM pixels
// — four 4-bit texels live in one 16-bit VRAM word.
type RegionSpec struct {
	X, Y   int
	W, H   int
	Format string

	// PalX, PalY locate the palette: the VRAM pixel where its first entry sits. A CLUT
	// is a horizontal run of 16 (clut4) or 256 (clut8) 16-bit pixels.
	PalX, PalY int
}

// RenderRegion decodes a rectangle of VRAM. Coordinates wrap, as they do for the GPU
// itself, so an aim past the edge shows you the wrap rather than an error.
func (m *Machine) RenderRegion(s RegionSpec) (*image.RGBA, error) {
	if s.W <= 0 || s.H <= 0 {
		return nil, fmt.Errorf("psx: a %dx%d region has nothing to show", s.W, s.H)
	}
	if s.W*s.H > 4<<20 {
		return nil, fmt.Errorf("psx: a %dx%d region is too large to render", s.W, s.H)
	}
	g := m.gpu
	at := func(x, y int) uint16 { return g.vram[(y&(vramH-1))*vramW+(x&(vramW-1))] }
	pal := func(i int) color.RGBA {
		r, gr, b := rgb15(at(s.PalX+i, s.PalY))
		return color.RGBA{r, gr, b, 255}
	}

	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			var c color.RGBA
			switch s.Format {
			case "rgba5551":
				r, gr, b := rgb15(at(s.X+x, s.Y+y))
				c = color.RGBA{r, gr, b, 255}
			case "clut8":
				// Two texels per VRAM word.
				w := at(s.X+x/2, s.Y+y)
				c = pal(int(w>>(uint(x&1)*8)) & 0xFF)
			case "clut4":
				// Four texels per VRAM word.
				w := at(s.X+x/4, s.Y+y)
				c = pal(int(w>>(uint(x&3)*4)) & 0xF)
			default:
				return nil, fmt.Errorf("psx: unknown pixel format %q (have %v)", s.Format, RegionFormats())
			}
			img.Set(x, y, c)
		}
	}
	return img, nil
}

// RenderVRAM renders the whole frame buffer as 16-bit colour — the god's-eye view: the
// displayed picture, the back buffer and every texture page the game has uploaded, in
// the one image.
func (m *Machine) RenderVRAM() *image.RGBA { return m.gpu.readRect(0, 0, vramW, vramH) }
