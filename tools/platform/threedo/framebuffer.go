package threedo

import "image"

// CaptureVRAM reads a frame buffer from VRAM into an image — how the oracle
// "sees" what the game has rendered. base is the buffer address (pass the
// machine's DisplayBuffer for the screen the game is showing). The layout is
// the hardware's interleaved line-pair format: two screen lines share a
// 4-byte-per-column stripe, the even line in the first halfword and the odd
// line in the second, each pixel big-endian RGB555.
func (m *Machine) CaptureVRAM(base uint32, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := base - vramBase + uint32(y>>1)*uint32(w)*4 + uint32(x)*4 + uint32(y&1)*2
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
