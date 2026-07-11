package n3ds

// gpu_png.go captures the visible frame as a PNG — the same first-frame
// instrument the N64 oracle's vi_png.go provides. The source is the linear
// framebuffer the game's most recent DisplayTransfer produced (per screen).
// A 3DS framebuffer is stored rotated: 240 pixels wide by 400 (top) or 320
// (bottom) lines, scanning the physical panel's columns — the capture rotates
// it back to the natural landscape orientation.

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// Framebuffer decodes the last-presented framebuffer for a screen ("top" or
// "bottom") into a natural-orientation image, or nil if no DisplayTransfer
// has happened yet.
func (m *Machine) Framebuffer(screen string) *image.NRGBA {
	rec := m.lastXferTop
	if screen == "bottom" {
		rec = m.lastXferBottom
	}
	if rec.dst == 0 {
		return nil
	}
	w, h := int(rec.w), int(rec.h)
	img := image.NewNRGBA(image.Rect(0, 0, h, w)) // rotated: fb lines become columns
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := rec.dst + uint32(y*int(rec.stride)+x)*rec.bpp
			var r, g, b byte
			switch rec.format {
			case 0: // RGBA8 bytes [A,B,G,R]
				b, g, r = m.Read(p+1), m.Read(p+2), m.Read(p+3)
			case 1: // RGB8 bytes [B,G,R]
				b, g, r = m.Read(p), m.Read(p+1), m.Read(p+2)
			case 2: // RGB565
				v := uint16(m.Read(p)) | uint16(m.Read(p+1))<<8
				r5, g6, b5 := byte(v>>11)&31, byte(v>>5)&63, byte(v)&31
				r, g, b = r5<<3|r5>>2, g6<<2|g6>>4, b5<<3|b5>>2
			case 3: // RGB5A1
				v := uint16(m.Read(p)) | uint16(m.Read(p+1))<<8
				r5, g5, b5 := byte(v>>11)&31, byte(v>>6)&31, byte(v>>1)&31
				r, g, b = r5<<3|r5>>2, g5<<3|g5>>2, b5<<3|b5>>2
			case 4: // RGBA4
				v := uint16(m.Read(p)) | uint16(m.Read(p+1))<<8
				r, g, b = byte(v>>12)*17, byte(v>>8&15)*17, byte(v>>4&15)*17
			}
			// Framebuffer (x, y) → screen (y, width-1-x): the panel scans the
			// landscape image's columns bottom-up.
			o := img.PixOffset(y, w-1-x)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, g, b, 255
		}
	}
	return img
}

// Screenshot writes both screens' framebuffers as PNGs: base_top.png and (if
// the bottom screen has been presented) base_bottom.png.
func (m *Machine) Screenshot(base string) error {
	wrote := false
	for _, screen := range []string{"top", "bottom"} {
		img := m.Framebuffer(screen)
		if img == nil {
			continue
		}
		f, err := os.Create(fmt.Sprintf("%s_%s.png", base, screen))
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		wrote = true
	}
	if !wrote {
		return fmt.Errorf("n3ds: no framebuffer presented yet (no DisplayTransfer)")
	}
	return nil
}
