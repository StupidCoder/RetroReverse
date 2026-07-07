package threedo

import "image"

// CaptureVRAM reads a region of VRAM as a 16-bit RGB555 bitmap — the 3DO
// framebuffer format — into an image. base is the framebuffer address in VRAM
// (the game's first VRAM allocation is the display buffer). This is how the
// oracle "sees" what the game has rendered.
func (m *Machine) CaptureVRAM(base uint32, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := base - vramBase + uint32((y*w+x)*2)
			if int(o)+2 > len(m.vram) {
				continue
			}
			c := uint16(m.vram[o])<<8 | uint16(m.vram[o+1])
			img.SetRGBA(x, y, rgb555(c))
		}
	}
	return img
}

// VRAMNonZero reports how many of the first n bytes of VRAM are non-zero — a
// quick check for whether anything has been rendered.
func (m *Machine) VRAMNonZero(n int) int {
	if n > len(m.vram) {
		n = len(m.vram)
	}
	c := 0
	for i := 0; i < n; i++ {
		if m.vram[i] != 0 {
			c++
		}
	}
	return c
}
