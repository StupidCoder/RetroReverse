package xbox

// nv2a_swizzle_test.go covers the swizzled render-target support: the format-word
// geometry decode (swizzleGeom), the Morton address math the raster writes and the
// texture unit reads (swizzleOffset), and the write/read round-trip that a
// render-to-texture depends on. The mutation checks use a NON-SQUARE grid on purpose:
// a square target hides an x/y transposition, so it could not tell a correct interleave
// from a swapped one.

import (
	"image"
	"testing"
)

func TestSwizzleGeom(t *testing.T) {
	cases := []struct {
		format     uint32
		w, h       int
		wantReason bool
	}{
		{0x07070228, 128, 128, false}, // OutRun's render-to-texture: 1<<7 x 1<<7
		{0x03030228, 8, 8, false},     // a smaller square
		{0x06070228, 128, 64, true},   // non-square (log2 7 x 6) — orientation unverified
		{0x17070228, 128, 128, true},  // bit 28-31 set — unexpected high nibble
		{0x07170228, 128, 128, true},  // bit 20-23 set — unexpected high nibble
	}
	for _, c := range cases {
		w, h, reason := swizzleGeom(c.format)
		if (reason != "") != c.wantReason {
			t.Errorf("swizzleGeom(%08X): reason=%q, wantReason=%v", c.format, reason, c.wantReason)
		}
		if c.wantReason {
			continue
		}
		if w != c.w || h != c.h {
			t.Errorf("swizzleGeom(%08X) = %dx%d, want %dx%d", c.format, w, h, c.w, c.h)
		}
	}
}

func TestSwizzleOffsetKnownValues(t *testing.T) {
	// On a 4x4 grid the Morton offset interleaves x (low bit first) with y:
	// bit0=x0, bit1=y0, bit2=x1, bit3=y1. These are hand-computed.
	want := map[[2]int]int{
		{0, 0}: 0, {1, 0}: 1, {2, 0}: 4, {3, 0}: 5,
		{0, 1}: 2, {1, 1}: 3, {2, 1}: 6, {3, 1}: 7,
		{0, 2}: 8, {1, 2}: 9, {2, 2}: 12, {3, 2}: 13,
		{0, 3}: 10, {1, 3}: 11, {2, 3}: 14, {3, 3}: 15,
	}
	for xy, exp := range want {
		if got := swizzleOffset(xy[0], xy[1], 4, 4); got != exp {
			t.Errorf("swizzleOffset(%d,%d,4,4) = %d, want %d", xy[0], xy[1], got, exp)
		}
	}
}

func TestSwizzleOffsetBijection(t *testing.T) {
	// The offset must be a permutation of [0, w*h): every texel maps to a distinct byte
	// and every byte is used. A collision would corrupt both the write and the depth
	// test; a gap would leave stale pixels. Checked on a non-square grid.
	for _, wh := range [][2]int{{4, 4}, {8, 4}, {4, 8}, {16, 8}, {128, 128}} {
		w, h := wh[0], wh[1]
		seen := make([]bool, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				o := swizzleOffset(x, y, w, h)
				if o < 0 || o >= w*h {
					t.Fatalf("swizzleOffset(%d,%d,%d,%d)=%d out of [0,%d)", x, y, w, h, o, w*h)
				}
				if seen[o] {
					t.Fatalf("swizzleOffset collision at offset %d on %dx%d", o, w, h)
				}
				seen[o] = true
			}
		}
	}
}

func TestSwizzleOffsetTransposeSensitive(t *testing.T) {
	// The mutation the brief demands: on a NON-SQUARE grid, transposing x and y must
	// change the offset. If it did not, our address function could not tell width from
	// height and a wrong orientation would pass silently. (On a square grid this test is
	// vacuous, which is exactly why the surface decode halts on non-square targets.)
	const w, h = 8, 4
	diff := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// The transposed coordinate only exists when it is in range; compare where both are.
			if x < h && y < w {
				if swizzleOffset(x, y, w, h) != swizzleOffset(y, x, w, h) {
					diff = true
				}
			}
		}
	}
	if !diff {
		t.Fatal("swizzleOffset is transpose-invariant on 8x4 — the interleave cannot distinguish x from y")
	}
}

func TestSwizzleOffsetOffByOne(t *testing.T) {
	// An off-by-one in the interleave (shifting one axis's bit) must be observable: the
	// offset for adjacent texels differs by the true Morton step, not a linear one.
	// On a 4x4 grid, (0,0)->0 and (0,1)->2 (y's low bit is bit 1), NOT 4 (a linear row
	// stride would give 4). This pins the interleave against a row-major mistake.
	if swizzleOffset(0, 1, 4, 4) != 2 {
		t.Fatalf("swizzleOffset(0,1,4,4)=%d, want 2 (Morton, not row stride)", swizzleOffset(0, 1, 4, 4))
	}
	if swizzleOffset(0, 1, 4, 4) == 4 {
		t.Fatal("swizzleOffset addressed row-major, not Morton")
	}
}

// TestSwizzledRoundTrip is the render-to-texture property: what the raster writes at a
// swizzled address (colorAddr's swizzleOffset*4) must be exactly what the texture unit
// reads back (decodeSwizzled, the independently-derived texture path). It is run on a
// NON-SQUARE 8x4 so a transposed write would surface as a scrambled read.
func TestSwizzledRoundTrip(t *testing.T) {
	const w, h, bpp = 8, 4, 4
	ram := make([]byte, w*h*bpp)
	// Write a unique BGRA value per texel through the same address math colorAddr uses.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := swizzleOffset(x, y, w, h) * bpp
			// A8R8G8B8 in RAM is B,G,R,A. Encode (x,y) into R and G so a scramble shows.
			ram[o+0] = 0x11    // B
			ram[o+1] = byte(y) // G
			ram[o+2] = byte(x) // R
			ram[o+3] = 0xFF    // A
		}
	}
	// Read it back through the texture decoder.
	img := &texImage{w: w, h: h, pix: make([]byte, w*h*4)}
	decodeSwizzled(img, ram, 0, bpp, func(pix []byte, o int, b []byte) {
		pix[o], pix[o+1], pix[o+2], pix[o+3] = b[2], b[1], b[0], b[3] // R,G,B,A from B,G,R,A
	})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			if img.pix[i+0] != byte(x) || img.pix[i+1] != byte(y) {
				t.Fatalf("round-trip at (%d,%d): read R=%d G=%d, want R=%d G=%d — swizzle write/read disagree",
					x, y, img.pix[i+0], img.pix[i+1], x, y)
			}
		}
	}
}

// TestRenderSwizzledSurface checks the export de-swizzle (renderSwizzledSurface) recovers
// the logical image a swizzled write produced — the dump must not read the target linearly.
func TestRenderSwizzledSurface(t *testing.T) {
	const w, h, bpp = 8, 4, 4
	m := &Machine{RAM: make([]byte, w*h*bpp)}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := swizzleOffset(x, y, w, h) * bpp
			m.RAM[o+0] = byte(x) // B
			m.RAM[o+1] = byte(y) // G
			m.RAM[o+2] = 0x22    // R
			m.RAM[o+3] = 0xFF    // A
		}
	}
	img, err := m.renderSwizzledSurface(0, w, h, w, h)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds() != image.Rect(0, 0, w, h) {
		t.Fatalf("bounds %v", img.Bounds())
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			// renderSwizzledSurface emits R,G,B,A; we wrote B=x, so blue channel = x.
			if img.Pix[i+2] != byte(x) || img.Pix[i+1] != byte(y) {
				t.Fatalf("de-swizzle at (%d,%d): B=%d G=%d, want B=%d G=%d",
					x, y, img.Pix[i+2], img.Pix[i+1], x, y)
			}
		}
	}
}
