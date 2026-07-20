package xbox

import "testing"

// TestImageBlitCopiesRect drives the 2D engine the way OutRun's bloom does — program the
// source/destination surfaces, then write SET_SIZE to trigger the copy — and checks a
// rectangle is moved byte-for-byte, honouring each surface's pitch and origin.
func TestImageBlitCopiesRect(t *testing.T) {
	m := &Machine{RAM: make([]byte, 1<<16)}
	m.pgraph = newPgraph(m)
	g := m.pgraph

	const (
		srcOff, dstOff     = 0x1000, 0x4000
		srcPitch, dstPitch = 64, 128 // deliberately different, so a pitch bug shows
		w, h, bpp          = 8, 4, 4
	)
	// A recognisable pattern in the source rect, and 0xEE guard bytes around the dest.
	for y := 0; y < h; y++ {
		for x := 0; x < w*bpp; x++ {
			m.RAM[srcOff+y*srcPitch+x] = byte((y*w*bpp + x) & 0xFF)
		}
	}
	for i := range m.RAM[dstOff : dstOff+h*dstPitch] {
		m.RAM[dstOff+i] = 0xEE
	}

	// Program NV_CONTEXT_SURFACES_2D, then NV_IMAGE_BLIT.
	g.surf2DMethod(surf2DFormat, 0x0A) // A8R8G8B8 -> 4 bpp
	g.surf2DMethod(surf2DPitch, srcPitch|(dstPitch<<16))
	g.surf2DMethod(surf2DSrcOff, srcOff)
	g.surf2DMethod(surf2DDstOff, dstOff)
	g.blitMethod(blitPointIn, 0)
	g.blitMethod(blitPointOut, 0)
	g.blitMethod(blitSize, uint32(h<<16|w)) // triggers the copy

	// Each destination row must equal the matching source row's w*bpp bytes...
	for y := 0; y < h; y++ {
		for x := 0; x < w*bpp; x++ {
			if got, want := m.RAM[dstOff+y*dstPitch+x], m.RAM[srcOff+y*srcPitch+x]; got != want {
				t.Fatalf("dst row %d byte %d = %02X, want %02X", y, x, got, want)
			}
		}
		// ...and the bytes past the copied width (dst pitch > src width) must be untouched.
		if g := m.RAM[dstOff+y*dstPitch+w*bpp]; g != 0xEE {
			t.Fatalf("blit overran row %d into %02X past the %d-byte width", y, g, w*bpp)
		}
	}
}

func TestBlitBoundsAreSafe(t *testing.T) {
	m := &Machine{RAM: make([]byte, 4096)}
	m.pgraph = newPgraph(m)
	g := m.pgraph
	// A destination that runs off the end of RAM must abort, not panic or corrupt.
	g.surf2DMethod(surf2DFormat, 0x0A)
	g.surf2DMethod(surf2DPitch, 64|(64<<16))
	g.surf2DMethod(surf2DSrcOff, 0)
	g.surf2DMethod(surf2DDstOff, 4000)
	g.blitMethod(blitSize, uint32(100<<16|16)) // 100 rows from offset 4000 overruns 4096
}
