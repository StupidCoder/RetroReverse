package dsmachine

// The oracle's window onto what the console would be showing: the two LCD panels,
// as the 2D engines composed them.
//
// Which engine drives which panel is not fixed. POWCNT1 bit 15 swaps them, and a DS
// game uses that freely — so "the top screen" is a question about a register, not
// about an engine, and an oracle that hard-wires engine A to the top screen will
// show a game's two screens the right way up and the wrong way round. The swap is
// resolved in gpu2d.screens(); this file just presents the result.

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// ScreenW and ScreenH are the size of each of the DS's two panels.
const (
	ScreenW = 256
	ScreenH = 192
)

// Screens returns the two panels as images: the top screen and the bottom screen,
// with POWCNT1's engine/panel assignment already applied.
func (m *Machine) Screens() (top, bottom *image.RGBA) {
	t, b := m.gpu2d.screens()
	return toImage(t), toImage(b)
}

func toImage(px []uint32) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	if px == nil {
		return img
	}
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			v := px[y*ScreenW+x]
			img.Set(x, y, color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), 0xFF})
		}
	}
	return img
}

// WriteShot writes both panels as PNGs, named base+"_top.png" and base+"_bottom.png"
// — the same two-file shape the 3DS oracle's -shot uses, because the question a
// screenshot answers on a two-screen console is always about both of them.
func (m *Machine) WriteShot(base string) error {
	top, bottom := m.Screens()
	for _, s := range []struct {
		name string
		img  *image.RGBA
	}{{base + "_top.png", top}, {base + "_bottom.png", bottom}} {
		f, err := os.Create(s.name)
		if err != nil {
			return err
		}
		if err := png.Encode(f, s.img); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Write3DShot writes the 3D engine's own render target as a PNG — what the
// rasteriser produced, before the 2D engine composited it as engine A's BG0.
//
// This is the DS's answer to the 3DS oracle's -rtshot, and it exists for the same
// reason: a black screen is two entirely different bugs wearing the same face. Either
// the 3D engine drew nothing, or it drew the frame and the compositor threw it away.
// No counter distinguishes them; looking at the plane the rasteriser actually wrote
// does, immediately.
//
// Pixels the 3D engine did not draw have zero alpha, and are written as magenta here
// rather than as black, so "nothing was drawn" cannot be mistaken for "black was
// drawn" — which, on a console whose 3D layer clears to black, it otherwise would be.
func (m *Machine) Write3DShot(path string) error {
	px := m.gpu2d.threeD
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			c := color.RGBA{0xFF, 0x00, 0xFF, 0xFF} // untouched by the rasteriser
			if px != nil {
				if v := px[y*ScreenW+x]; uint8(v) != 0 {
					c = color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), 0xFF}
				}
			}
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// EngineStats reports, for each 2D engine, how many of its 49,152 pixels came out
// non-black on the last composed frame. A engine that renders a valid backdrop
// cannot be at zero, so a zero here says the engine did not run, not that the game
// drew nothing.
func (m *Machine) EngineStats() (a, b int) {
	count := func(px []uint32) int {
		n := 0
		for _, v := range px {
			if v&0xFFFFFF00 != 0 {
				n++
			}
		}
		return n
	}
	return count(m.gpu2d.a.out), count(m.gpu2d.b.out)
}
