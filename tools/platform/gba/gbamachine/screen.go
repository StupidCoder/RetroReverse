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

// RenderLayers recomposes the CURRENT video state and returns the frame, with
// sprites optionally suppressed. It steps no CPU and consumes no time: it
// re-runs the PPU over the VRAM, palette and registers exactly as they stand,
// which is what makes it usable as a reference for an offline exporter — a
// background-only picture cannot otherwise be obtained from a running game,
// because the game rewrites DISPCNT every frame and would undo a poke.
func (m *Machine) RenderLayers(objects bool) *image.RGBA {
	saved := m.io[0x000]
	if !objects {
		m.io[0x000] = saved &^ (1 << 12)
	}
	savedScreen := m.screen
	for y := 0; y < screenH; y++ {
		m.ppu.renderLine(m, y)
	}
	img := m.Screen()
	m.io[0x000] = saved
	m.screen = savedScreen
	return img
}
