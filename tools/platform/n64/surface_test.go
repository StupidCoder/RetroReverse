package n64

import (
	"image/color"
	"testing"
)

// A machine with just enough RAM to decode from. RenderRegion reads RDRAM directly, so
// nothing has to boot.
func ramMachine(t *testing.T) *Machine {
	t.Helper()
	return &Machine{RDRAM: make([]byte, 0x1000)}
}

func TestRenderRegionRGBA16(t *testing.T) {
	m := ramMachine(t)
	// RGBA16 is 5:5:5:1, big-endian. Red is the top five bits.
	// 0xF801 = 11111 00000 00000 1 → opaque red.
	m.RDRAM[0], m.RDRAM[1] = 0xF8, 0x01
	// 0x07C1 = 00000 11111 00000 1 → opaque green.
	m.RDRAM[2], m.RDRAM[3] = 0x07, 0xC1

	img, err := m.RenderRegion(RegionSpec{Addr: 0, W: 2, H: 1, Format: "rgba16"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("texel 0 = %v, want %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 248, 0, 255}); got != want {
		t.Errorf("texel 1 = %v, want %v", got, want)
	}
}

func TestRenderRegionRGBA32(t *testing.T) {
	m := ramMachine(t)
	copy(m.RDRAM, []byte{0x11, 0x22, 0x33, 0x44})
	img, err := m.RenderRegion(RegionSpec{Addr: 0, W: 1, H: 1, Format: "rgba32"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{0x11, 0x22, 0x33, 0x44}); got != want {
		t.Errorf("texel = %v, want %v", got, want)
	}
}

// TestRenderRegionIndexed: the indexed formats are the ones a texture upload actually
// uses, and they are the ones where a wrong palette address shows you nothing.
func TestRenderRegionIndexed(t *testing.T) {
	m := ramMachine(t)
	// A palette at 0x100: entry 0 = black, entry 1 = opaque red, entry 2 = opaque blue.
	pal := uint32(0x100)
	m.RDRAM[pal+2], m.RDRAM[pal+3] = 0xF8, 0x01 // red
	m.RDRAM[pal+4], m.RDRAM[pal+5] = 0x00, 0x3F // blue: 00000 00000 11111 1

	// CI8: one index per byte.
	m.RDRAM[0], m.RDRAM[1] = 1, 2
	img, err := m.RenderRegion(RegionSpec{Addr: 0, W: 2, H: 1, Format: "ci8", Palette: pal})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("ci8 texel 0 = %v, want red %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 0, 248, 255}); got != want {
		t.Errorf("ci8 texel 1 = %v, want blue %v", got, want)
	}

	// CI4: two indices per byte, the high nibble first.
	m.RDRAM[0] = 0x12
	img, err = m.RenderRegion(RegionSpec{Addr: 0, W: 2, H: 1, Format: "ci4", Palette: pal})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.RGBAAt(0, 0), (color.RGBA{248, 0, 0, 255}); got != want {
		t.Errorf("ci4 texel 0 = %v, want the HIGH nibble's colour %v", got, want)
	}
	if got, want := img.RGBAAt(1, 0), (color.RGBA{0, 0, 248, 255}); got != want {
		t.Errorf("ci4 texel 1 = %v, want the low nibble's colour %v", got, want)
	}
}

func TestRenderRegionStride(t *testing.T) {
	m := ramMachine(t)
	// Two rows of one RGBA16 texel each, 8 bytes apart.
	m.RDRAM[0], m.RDRAM[1] = 0xF8, 0x01 // red
	m.RDRAM[8], m.RDRAM[9] = 0x07, 0xC1 // green
	img, err := m.RenderRegion(RegionSpec{Addr: 0, W: 1, H: 2, Stride: 8, Format: "rgba16"})
	if err != nil {
		t.Fatal(err)
	}
	if got := img.RGBAAt(0, 1); got != (color.RGBA{0, 248, 0, 255}) {
		t.Errorf("row 1 = %v; the stride was not honoured", got)
	}
}

// TestRenderRegionOutOfRange: an address past RDRAM shows rubbish, not an error. A
// debugger aimed at the wrong place should tell you that by what it draws.
func TestRenderRegionOutOfRange(t *testing.T) {
	m := ramMachine(t)
	img, err := m.RenderRegion(RegionSpec{Addr: 0xFFFF_0000, W: 4, H: 4, Format: "rgba16"})
	if err != nil {
		t.Fatalf("an out-of-range address failed instead of reading as zero: %v", err)
	}
	if img.RGBAAt(0, 0).R != 0 {
		t.Error("out-of-range memory did not read as zero")
	}
}

func TestRenderRegionRejectsNonsense(t *testing.T) {
	m := ramMachine(t)
	if _, err := m.RenderRegion(RegionSpec{W: 4, H: 4, Format: "not-a-format"}); err == nil {
		t.Error("an unknown pixel format was accepted")
	}
	if _, err := m.RenderRegion(RegionSpec{W: 0, H: 4, Format: "rgba16"}); err == nil {
		t.Error("a zero-width region was accepted")
	}
}
