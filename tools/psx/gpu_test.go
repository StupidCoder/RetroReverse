package psx

import "testing"

func (g *gpu) pixel(x, y int) uint16 { return g.vram[y*vramW+x] }

// TestGPUFlatTriangle draws a flat-shaded triangle via GP0 and checks that an
// interior pixel is filled and an exterior pixel is not.
func TestGPUFlatTriangle(t *testing.T) {
	g := newGPU()
	// 0x20 flat triangle, red (0x0000FF); verts (10,10),(100,10),(10,100).
	for _, w := range []uint32{0x200000FF, 0x000A000A, 0x000A0064, 0x0064000A} {
		g.gp0(w)
	}
	if got := g.pixel(20, 20); got != 0x001F { // red -> 15-bit R=0x1F
		t.Errorf("interior pixel = 0x%04X, want 0x001F", got)
	}
	if got := g.pixel(90, 90); got != 0 { // outside the triangle
		t.Errorf("exterior pixel = 0x%04X, want 0", got)
	}
}

// TestGPUGouraudTriangle checks per-vertex colour interpolation.
func TestGPUGouraudTriangle(t *testing.T) {
	g := newGPU()
	// 0x30 gouraud triangle: red, green, blue corners.
	for _, w := range []uint32{
		0x300000FF, 0x000A000A, // c0=red, v0
		0x0000FF00, 0x000A0064, // c1=green, v1
		0x00FF0000, 0x0064000A, // c2=blue, v2
	} {
		g.gp0(w)
	}
	// Near the red vertex the pixel should be red-dominant.
	px := g.pixel(14, 14)
	r := px & 0x1F
	if r < 0x10 {
		t.Errorf("near-red vertex pixel = 0x%04X, expected red-dominant", px)
	}
}

// TestGPUImageLoadAndDisplay uploads a 2x2 image via GP0 0xA0 and reads it back.
func TestGPUImageLoadAndDisplay(t *testing.T) {
	g := newGPU()
	// CPU->VRAM copy to (0,0), size 2x2: header (3 words) then 2 pixel-pair words.
	g.gp0(0xA0000000)          // command
	g.gp0(0x00000000)          // dest x=0,y=0
	g.gp0(0x00020002)          // w=2,h=2
	g.gp0(0x7C1F_03E0)         // px(0,0)=green(0x03E0), px(1,0)=cyan(0x7C1F)
	g.gp0(0x0000_001F)         // px(0,1)=red(0x001F), px(1,1)=black(0)
	if g.pixel(0, 0) != 0x03E0 || g.pixel(1, 0) != 0x7C1F || g.pixel(0, 1) != 0x001F {
		t.Errorf("image load mismatch: %04X %04X %04X",
			g.pixel(0, 0), g.pixel(1, 0), g.pixel(0, 1))
	}
}
