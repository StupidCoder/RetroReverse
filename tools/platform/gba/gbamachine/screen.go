package gbamachine

// The screen instrument: what the composed frame looks like, as a PNG.

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// Screen returns the last composed frame as an image (240x160).
func (m *Machine) Screen() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, screenW, screenH))
	for y := 0; y < screenH; y++ {
		for x := 0; x < screenW; x++ {
			p := m.screen[y*screenW+x]
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(p >> 16), G: uint8(p >> 8), B: uint8(p), A: 0xFF,
			})
		}
	}
	return img
}

// Screenshot writes the last composed frame to a PNG file.
func (m *Machine) Screenshot(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, m.Screen()); err != nil {
		return err
	}
	return f.Close()
}
