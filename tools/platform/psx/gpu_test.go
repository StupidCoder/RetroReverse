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

// TestFeedNodeSkipsDisabledPrimitive checks that a linked-list node whose
// primitive opcode word is zeroed (a culled/disabled slot Ridge Racer leaves in
// its ordering table) is skipped whole: its interior texcoord/CLUT words — whose
// high bytes alias GP0 rectangle/polygon opcodes — must not be executed, and the
// GP0 FIFO must stay synced so the next real primitive still draws.
func TestFeedNodeSkipsDisabledPrimitive(t *testing.T) {
	g := newGPU()
	// A disabled POLY_FT4 slot: opcode+colour word zeroed. Word [2] is a
	// texcoord+CLUT word (0x7B0000FF) whose top byte 0x7B is a 16x16 rectangle
	// opcode; if fed forward it would paint a rectangle at word [3]'s coords.
	g.feedNode([]uint32{
		0x00000000, // opcode+colour = NOP -> whole node skipped
		0x00320032,
		0x7B0000FF, // top byte 0x7B looks like a rect opcode
		0x00320032, // (50,50) -> where the spurious rect would land
		0x00000000, 0x00000000, 0x00000000, 0x00000000, 0x00000000,
	})
	if g.pixel(50, 50) != 0 {
		t.Errorf("disabled node drew a spurious pixel at (50,50): 0x%04X", g.pixel(50, 50))
	}
	if g.need != 0 || g.imgPx != 0 {
		t.Fatalf("FIFO desynced after skipped node: need=%d imgPx=%d", g.need, g.imgPx)
	}
	// A real primitive after the skipped node still draws (FIFO stayed synced):
	// a 4x4 fill-rect (GP0 0x02) at (0,0) in red.
	g.feedNode([]uint32{0x020000FF, 0x00000000, 0x00040004})
	if g.pixel(1, 1) != 0x001F {
		t.Errorf("fill after skipped node = 0x%04X, want 0x001F", g.pixel(1, 1))
	}
}

// TestModulate checks texture-colour modulation: 0x80 is neutral, brighter/
// darker colours scale the texel, and an all-zero colour is treated as raw
// (unmodulated) so surfaces lit by GTE ops we don't model yet aren't blacked out.
func TestModulate(t *testing.T) {
	white := uint16(0x7FFF) // all lanes 0x1F
	if got := modulate(white, 0x80, 0x80, 0x80); got != white {
		t.Errorf("neutral 0x80 modulation = 0x%04X, want 0x%04X (identity)", got, white)
	}
	if got := modulate(white, 0x40, 0x40, 0x40); got != 0x3DEF { // 0x1F*0x40>>7 = 0x0F per lane
		t.Errorf("half 0x40 modulation = 0x%04X, want 0x3DEF", got)
	}
	if got := modulate(white, 0, 0, 0); got != white {
		t.Errorf("zero-colour modulation = 0x%04X, want raw 0x%04X", got, white)
	}
}

// TestGPUImageStoreReadback checks the VRAM→CPU copy path (GP0 0xC0): after a
// 2x2 image is placed in VRAM, a store command must let GPUREAD stream those
// pixels back out, packed two per word, low pixel first. Ridge Racer relies on
// this to pull staged scenery/font textures back out of VRAM into a RAM work
// buffer; without it the buffer stays zero and the scenery renders untextured.
func TestGPUImageStoreReadback(t *testing.T) {
	g := newGPU()
	// Seed a 2x2 block at (5,5): (5,5)=green,(6,5)=cyan,(5,6)=red,(6,6)=blue.
	g.vram[5*vramW+5], g.vram[5*vramW+6] = 0x03E0, 0x7C1F
	g.vram[6*vramW+5], g.vram[6*vramW+6] = 0x001F, 0x7C00
	// VRAM→CPU copy from (5,5), size 2x2: header (3 words), then read 2 words.
	g.gp0(0xC0000000)
	g.gp0(5<<16 | 5)   // src x=5,y=5
	g.gp0(2<<16 | 2)   // w=2,h=2
	if w := g.read(); w != 0x7C1F_03E0 {
		t.Errorf("readback word 0 = 0x%08X, want 0x7C1F03E0", w)
	}
	if w := g.read(); w != 0x7C00_001F {
		t.Errorf("readback word 1 = 0x%08X, want 0x7C00001F", w)
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
