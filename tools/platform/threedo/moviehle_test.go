package threedo

import "testing"

// TestMovieHLEStep drives the movie player HLE end to end without the disc: it
// arms a hand-built two-frame movie, steps every frame, and confirms each frame
// lands in the display buffer in the hardware's interleaved RGB555 layout (read
// back through CaptureVRAM) at the centred position.
func TestMovieHLEStep(t *testing.T) {
	m := &Machine{vram: make([]byte, vramSize)}
	m.displayBuf = vramBase // land the movie at the base of VRAM

	frame := buildCvidFrame() // 8x8, macroblock (0,0) is luma 10, no chroma
	mv := &CvidMovie{Width: 8, Height: 8, Codec: "cvid", FPS: 15, Frames: [][]byte{frame, frame}}
	m.movieQueue = []*armedMovie{{name: "test", mv: mv}}

	if m.MoviesPending() != 1 {
		t.Fatalf("MoviesPending = %d, want 1", m.MoviesPending())
	}

	stepped := 0
	for {
		buf, name, idx, total, ok := m.StepMovieFrame()
		if !ok {
			break
		}
		if name != "test" || total != 2 || buf != vramBase {
			t.Fatalf("step: name=%q total=%d buf=0x%X idx=%d", name, total, buf, idx)
		}
		stepped++
	}
	if stepped != 2 {
		t.Fatalf("stepped %d frames, want 2", stepped)
	}
	if m.MoviesPending() != 0 {
		t.Fatalf("MoviesPending after playback = %d, want 0", m.MoviesPending())
	}

	// The 8x8 movie is centred in the 320x240 screen: yoff = (240-8)/2 = 116.
	// Its pixel (0,0) is macroblock 0's luma[0] = 10 (grey), read back at (0,116).
	got := m.CaptureVRAM(m.displayBuf, screenW, screenH)
	o := got.PixOffset(0, 116)
	if got.Pix[o]>>3 != 10>>3 {
		t.Fatalf("centred movie pixel r=%d, want ~10", got.Pix[o])
	}
	// A row well outside the 8-pixel-tall movie band must be black (letterbox).
	ob := got.PixOffset(0, 0)
	if got.Pix[ob] != 0 || got.Pix[ob+1] != 0 || got.Pix[ob+2] != 0 {
		t.Fatalf("letterbox pixel not black: (%d,%d,%d)", got.Pix[ob], got.Pix[ob+1], got.Pix[ob+2])
	}
}
