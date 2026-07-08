package gfx

import "testing"

// TestRenderBitmapMC checks the four multicolor bit-pairs map to the right
// colour sources (background / matrix hi / matrix lo / colour RAM).
func TestRenderBitmapMC(t *testing.T) {
	bitmap := make([]byte, 8000)
	matrix := make([]byte, 1000)
	colorRAM := make([]byte, 1000)

	// Cell 0, top row: one pixel of each bit-pair, 00 01 10 11 = %00011011.
	bitmap[0] = 0x1B
	matrix[0] = 0x57   // 01 -> $5, 10 -> $7
	colorRAM[0] = 0x0A // 11 -> $A
	const bg = 0x06

	img := RenderBitmapMC(bitmap, matrix, colorRAM, bg, 1)
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 200 {
		t.Fatalf("size = %v, want 320x200", img.Bounds())
	}
	// Each colour pixel is two hires pixels wide; sample the left of each.
	want := []struct {
		x   int
		col byte
	}{
		{0, bg}, {2, 0x5}, {4, 0x7}, {6, 0xA},
	}
	for _, w := range want {
		got := img.RGBAAt(w.x, 0)
		if got != Palette[w.col] {
			t.Errorf("pixel x=%d: got %v, want palette[$%X]=%v", w.x, got, w.col, Palette[w.col])
		}
	}
}
