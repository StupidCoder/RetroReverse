package psx

import (
	"image/color"
	"testing"
)

func vramMachine() *Machine {
	m := &Machine{}
	m.gpu = newGPU()
	return m
}

// put writes one VRAM pixel.
func put(m *Machine, x, y int, v uint16) { m.gpu.vram[y*vramW+x] = v }

func TestRenderRegionRGBA5551(t *testing.T) {
	m := vramMachine()
	// The PSX packs 5:5:5 the other way round from the N64: red in the LOW bits.
	put(m, 0, 0, 0x001F) // red
	put(m, 1, 0, 0x03E0) // green

	img, err := m.RenderRegion(RegionSpec{W: 2, H: 1, Format: "rgba5551"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("pixel 0 = %v, want red %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 248, 0, 255}); got != want {
		t.Errorf("pixel 1 = %v, want green %v", got, want)
	}
}

// TestRenderRegionCLUT is the case that matters: a texture page is not colour at all,
// it is indices packed into 16-bit words, looked up in a palette that is itself a row of
// VRAM. Getting the packing order wrong shows you a plausible-looking mess.
func TestRenderRegionCLUT(t *testing.T) {
	m := vramMachine()
	// A palette at (0, 100): entry 1 = red, entry 2 = green, entry 15 = blue.
	put(m, 1, 100, 0x001F)
	put(m, 2, 100, 0x03E0)
	put(m, 15, 100, 0x7C00)

	// clut4: four texels per word, LOW nibble first.
	put(m, 0, 0, 0x0021) // texels 1, 2, 0, 0
	img, err := m.RenderRegion(RegionSpec{
		W: 4, H: 1, Format: "clut4", PalX: 0, PalY: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("clut4 texel 0 = %v, want the LOW nibble's entry (red) %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 248, 0, 255}); got != want {
		t.Errorf("clut4 texel 1 = %v, want green %v", got, want)
	}

	// clut8: two texels per word, low byte first.
	put(m, 0, 0, 0x0F01) // texels 1, 15
	img, err = m.RenderRegion(RegionSpec{
		W: 2, H: 1, Format: "clut8", PalX: 0, PalY: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("clut8 texel 0 = %v, want red %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 0, 248, 255}); got != want {
		t.Errorf("clut8 texel 1 = %v, want blue %v", got, want)
	}
}

func TestRenderVRAMIsTheWholeBuffer(t *testing.T) {
	m := vramMachine()
	img := m.RenderVRAM()
	if img.Rect.Dx() != vramW || img.Rect.Dy() != vramH {
		t.Errorf("VRAM rendered as %v, want %dx%d", img.Rect, vramW, vramH)
	}
}

// TestRenderRegionWraps: VRAM coordinates wrap for the GPU, so they wrap here too — an
// aim past the edge shows the wrap rather than failing.
func TestRenderRegionWraps(t *testing.T) {
	m := vramMachine()
	put(m, 0, 0, 0x001F)
	img, err := m.RenderRegion(RegionSpec{X: vramW, Y: vramH, W: 1, H: 1, Format: "rgba5551"})
	if err != nil {
		t.Fatal(err)
	}
	if got := img.RGBAAt(0, 0); got.R != 248 {
		t.Errorf("a wrapped read gave %v, want the pixel at (0,0)", got)
	}
}

func TestRenderRegionRejectsNonsense(t *testing.T) {
	m := vramMachine()
	if _, err := m.RenderRegion(RegionSpec{W: 4, H: 4, Format: "nope"}); err == nil {
		t.Error("an unknown pixel format was accepted")
	}
}
