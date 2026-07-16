package gc

// surface.go renders a region of main memory as an image.
//
// This is what a debugger points at the address a DMA just landed on, to see whether what
// arrived there is the texture the game meant to upload — or at the middle of a decompressed
// archive, to find out what is actually in it.
//
// It decodes through the very same per-format texel readers the rasteriser samples with
// (gpu_texture.go's decodeTexel), rather than a second copy written for the debugger. A
// debugger that decodes a texture its own way can tell you the texture is fine while the
// texture unit reads it wrong, which is precisely the bug you opened the debugger to find.

import (
	"fmt"
	"image"
)

// RegionFormats are the pixel formats RenderRegion understands: Flipper's own texture
// formats, named as the GX library names them.
func RegionFormats() []string {
	return []string{"i4", "i8", "ia4", "ia8", "rgb565", "rgb5a3", "rgba8", "cmpr", "c4", "c8", "c14x2"}
}

// regionFormatCode maps a format name to the code the TX_SETIMAGE0 register carries, which
// is what decodeTexel dispatches on.
func regionFormatCode(name string) (int, error) {
	switch name {
	case "i4":
		return texI4, nil
	case "i8":
		return texI8, nil
	case "ia4":
		return texIA4, nil
	case "ia8":
		return texIA8, nil
	case "rgb565":
		return texRGB565, nil
	case "rgb5a3":
		return texRGB5A3, nil
	case "rgba8":
		return texRGBA8, nil
	case "cmpr":
		return texCMPR, nil
	case "c4":
		return texC4, nil
	case "c8":
		return texC8, nil
	case "c14x2":
		return texC14X2, nil
	}
	return 0, fmt.Errorf("gc: unknown pixel format %q (have %v)", name, RegionFormats())
}

// RegionSpec aims RenderRegion at memory.
type RegionSpec struct {
	Addr   uint32 // physical address of the texel data
	W, H   int    // size in texels
	Format string // one of RegionFormats

	// Palette is the physical address of a TLUT for the indexed formats (c4, c8, c14x2) —
	// the palette as it sits in main memory, which is where a debugger can still see it.
	// PaletteFormat is its entry format, named as the GX library names it: "ia8", "rgb565"
	// or "rgb5a3"; empty means rgb5a3, the common one.
	Palette       uint32
	PaletteFormat string
}

// paletteFormatCode maps a TLUT entry format name to the two-bit code the TX_SETTLUT
// register carries.
func paletteFormatCode(name string) (int, error) {
	switch name {
	case "ia8":
		return 0, nil
	case "rgb565":
		return 1, nil
	case "", "rgb5a3":
		return 2, nil
	}
	return 0, fmt.Errorf("gc: unknown palette format %q (want ia8, rgb565 or rgb5a3)", name)
}

// RenderRegion decodes W×H texels of main memory as an image.
//
// Addresses outside RAM read as zero rather than failing: a debugger aimed at the wrong
// address should show you rubbish, which tells you it is the wrong address, instead of an
// error message that tells you nothing.
func (m *Machine) RenderRegion(s RegionSpec) (*image.RGBA, error) {
	code, err := regionFormatCode(s.Format)
	if err != nil {
		return nil, err
	}
	pfmt, err := paletteFormatCode(s.PaletteFormat)
	if err != nil {
		return nil, err
	}
	if s.W <= 0 || s.H <= 0 {
		return nil, fmt.Errorf("gc: a %dx%d region has nothing to show", s.W, s.H)
	}
	if s.W*s.H > 4<<20 {
		return nil, fmt.Errorf("gc: a %dx%d region is too big to draw", s.W, s.H)
	}

	// The decoders take a bound texture's state, so the spec is expressed as one. Clamp
	// addressing keeps a region whose width is not a whole number of tiles from wrapping
	// round to the other edge, which would be a lie about what is in memory.
	tx := texState{
		format: code, width: s.W, height: s.H, base: s.Addr,
		wrapS: 0, wrapT: 0,
		tlutFmt: pfmt, tlutRAM: s.Palette, tlutFromRAM: true,
	}
	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			r, g, b, a := m.gpu.decodeTexel(m, tx, x, y)
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, g, b, a
		}
	}
	return img, nil
}
