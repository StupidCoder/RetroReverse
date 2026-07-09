package n64

// vi_png.go reads the framebuffer the Video Interface is scanning out and writes
// it to a PNG — the oracle's window onto what the console would be showing.
//
// The VI does more than copy: it scales, it applies an anti-aliasing filter and a
// de-dither filter, and it interlaces. None of that is modelled. What comes out
// here is the framebuffer as the RDP left it, at its own resolution, which is
// what a reverse-engineer wants to look at anyway.

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Framebuffer renders the image the VI is currently scanning out. Height is
// derived from the VI's vertical extent, since nothing records it directly.
func (m *Machine) Framebuffer() (*image.RGBA, error) {
	origin := m.Origin()
	width := m.Width()
	if origin == 0 || width == 0 {
		return nil, fmt.Errorf("n64: the VI is not scanning a framebuffer")
	}
	height := m.viHeight()

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	switch m.PixelType() {
	case viTypeRGBA16:
		for y := uint32(0); y < height; y++ {
			for x := uint32(0); x < width; x++ {
				a := origin + (y*width+x)*2
				if int(a)+1 >= len(m.RDRAM) {
					continue
				}
				c := fromRGBA16(uint16(m.RDRAM[a])<<8 | uint16(m.RDRAM[a+1]))
				img.Set(int(x), int(y), color.RGBA{uint8(c.R), uint8(c.G), uint8(c.B), 255})
			}
		}
	case viTypeRGBA32:
		for y := uint32(0); y < height; y++ {
			for x := uint32(0); x < width; x++ {
				a := origin + (y*width+x)*4
				if int(a)+3 >= len(m.RDRAM) {
					continue
				}
				img.Set(int(x), int(y), color.RGBA{m.RDRAM[a], m.RDRAM[a+1], m.RDRAM[a+2], 255})
			}
		}
	default:
		return nil, fmt.Errorf("n64: the VI is blanked (pixel type %d)", m.PixelType())
	}
	return img, nil
}

// viHeight derives the framebuffer's height from the VI's vertical start and end
// registers and its vertical scale, which is how the hardware knows it too.
func (m *Machine) viHeight() uint32 {
	vStart := m.vi.Regs[viVStart]
	start, end := vStart>>16&0x3FF, vStart&0x3FF
	yScale := m.vi.Regs[viYScale] & 0xFFF // 2.10 fixed point
	if end <= start || yScale == 0 {
		return 240
	}
	// The vertical extent is in half-lines.
	h := (end - start) / 2 * yScale >> 10
	if h == 0 || h > 512 {
		return 240
	}
	return h
}

// Screenshot writes the current framebuffer to a PNG file.
func (m *Machine) Screenshot(path string) error {
	img, err := m.Framebuffer()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
