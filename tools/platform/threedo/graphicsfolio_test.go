package threedo

import "testing"

// TestBlendPixelModes pins the PIXC pixel-processor equation against the two
// stock PPMP modes so the derived (first*MF)/DIV + (second>>SF) formula stays
// faithful: NORMAL copies the source, AVERAGE means (source+dest)/2.
func TestBlendPixelModes(t *testing.T) {
	m := NewMachine()
	// A 2x2 bitmap in VRAM is enough; blendPixel only touches one pixel.
	bm := gfxBitmap{buf: vramBase, w: 2, h: 2}

	// RGB555 helpers.
	pix := func(r, g, b uint16) uint16 { return r<<10 | g<<5 | b }
	read := func(x, y int) uint16 {
		a := bm.buf + uint32(y>>1)*uint32(bm.w)*4 + uint32(x)*4 + uint32(y&1)*2
		return uint16(m.Read(a))<<8 | uint16(m.Read(a+1))
	}

	// NORMAL (0x1F40 in both PPMP words): the source replaces the destination.
	src := pix(20, 10, 4)
	m.blendPixel(bm, 0, 0, src, 0x49, 0x1F401F40, 0)
	if got := read(0, 0); got != src {
		t.Fatalf("NORMAL: got %04X want %04X", got, src)
	}

	// AVERAGE (0x1F81): out = source/2 + dest/2, per channel.
	// Seed the destination, then average a source over it.
	dst := pix(8, 30, 2)
	m.blendPixel(bm, 1, 1, dst, 0x49, 0x1F401F40, 0) // lay down dst opaquely
	m.blendPixel(bm, 1, 1, src, 0x49, 0x1F811F81, 0) // average src over dst
	want := pix((20+8)/2, (10+30)/2, (4+2)/2)
	if got := read(1, 1); got != want {
		t.Fatalf("AVERAGE: got %04X want %04X (src %04X dst %04X)", got, want, src, dst)
	}
}

// TestBlendPixelClamps checks that an over-unity multiply saturates each 5-bit
// channel instead of overflowing into the neighbouring channel.
func TestBlendPixelClamps(t *testing.T) {
	m := NewMachine()
	bm := gfxBitmap{buf: vramBase, w: 2, h: 2}
	// MF=8, DIV=8, so out = source; but feed a max-channel source and confirm it
	// stays 0x1F rather than wrapping.
	src := uint16(0x1F)<<10 | uint16(0x1F)<<5 | 0x1F
	m.blendPixel(bm, 0, 0, src, 0x49, 0x1F401F40, 0)
	a := bm.buf
	got := uint16(m.Read(a))<<8 | uint16(m.Read(a+1))
	if got != src {
		t.Fatalf("clamp: got %04X want %04X", got, src)
	}
}
