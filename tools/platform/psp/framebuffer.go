package psp

// framebuffer.go decodes the PSP display framebuffer out of VRAM to an image. The
// framebuffer address, line stride and pixel format come from sceDisplaySetFrameBuf
// (or the GE's FBP/FBW registers). The PSP pixel formats (PSM) are 16-bit 5650/5551/
// 4444 and 32-bit 8888.

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// PSP display pixel formats.
const (
	psm5650 = 0 // RGB 565
	psm5551 = 1 // RGBA 5551
	psm4444 = 2 // RGBA 4444
	psm8888 = 3 // RGBA 8888
)

const (
	dispW = 480
	dispH = 272
)

// Framebuffer renders the current display framebuffer to an RGBA image. When the
// framebuffer address is unset it falls back to the start of VRAM.
func (m *Machine) Framebuffer() *image.RGBA {
	addr := m.fbAddr
	if addr == 0 {
		addr = vramBase
	}
	stride := m.fbWidth
	if stride == 0 {
		stride = dispW
	}
	img := image.NewRGBA(image.Rect(0, 0, dispW, dispH))
	for y := 0; y < dispH; y++ {
		for x := 0; x < dispW; x++ {
			img.Set(x, y, m.readPixel(addr, stride, uint32(x), uint32(y)))
		}
	}
	return img
}

// readPixel reads and decodes one framebuffer pixel.
func (m *Machine) readPixel(base, stride, x, y uint32) color.RGBA {
	switch m.fbFormat {
	case psm8888:
		a := base + (y*stride+x)*4
		p := m.read32(a)
		return color.RGBA{byte(p), byte(p >> 8), byte(p >> 16), 0xFF}
	default: // 16-bit formats
		a := base + (y*stride+x)*2
		p := uint16(m.Read(a)) | uint16(m.Read(a+1))<<8
		return decode16(p, m.fbFormat)
	}
}

// decode16 expands a 16-bit PSP pixel to RGBA8.
func decode16(p uint16, fmt uint32) color.RGBA {
	ext := func(v, bits uint16) byte {
		v &= (1 << bits) - 1
		return byte((uint32(v) * 255) / ((1 << bits) - 1))
	}
	switch fmt {
	case psm5551:
		return color.RGBA{ext(p, 5), ext(p>>5, 5), ext(p>>10, 5), 0xFF}
	case psm4444:
		return color.RGBA{ext(p, 4), ext(p>>4, 4), ext(p>>8, 4), 0xFF}
	default: // psm5650
		return color.RGBA{ext(p, 5), ext(p>>5, 6), ext(p>>11, 5), 0xFF}
	}
}

// Screenshot writes the display framebuffer to a PNG file.
func (m *Machine) Screenshot(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, m.Framebuffer()); err != nil {
		return err
	}
	return nil
}

// FramebufferInfo describes the current display target for diagnostics.
func (m *Machine) FramebufferInfo() string {
	return fmt.Sprintf("fb=0x%08X stride=%d psm=%d", m.fbAddr, m.fbWidth, m.fbFormat)
}
