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
			// Framebuffer (x, y) → screen (y, width-1-x): framebuffer lines run
			// right-to-left across the landscape image, and the buffer's first
			// line is the screen's first column because the rasteriser now
			// anchors the viewport to the buffer's bottom (the PICA's bottom-left
			// origin — see gpu_raster.go). The two flips cancel exactly whenever
			// the render target is no taller than the viewport, which is why
			// Super Mario 3D Land is pixel-identical across this change and only
			// Captain Toad's 256×512 target could reveal the anchor.
			o := img.PixOffset(y, w-1-x)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, g, b, 255
		}
	}
	return img
}

// RenderTarget decodes a tiled RGBA8 colour buffer straight out of memory, at
// the address and dimensions the PICA's own registers name — the GPU's view of
// what it drew, before any DisplayTransfer moves it. This is the instrument for
// "the draws report pixels but the screen is black": it separates *did we
// rasterise it* from *did it reach the panel*, which no counter can.
func (m *Machine) RenderTarget(addr, w, h uint32) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := uint32(0); y < h; y++ {
		for x := uint32(0); x < w; x++ {
			p := addr + tiledOffset(x, y, w)
			b, g, r := m.Read(p+1), m.Read(p+2), m.Read(p+3)
			o := img.PixOffset(int(x), int(y))
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

// ScreenGeom is where one panel's picture comes from: the window of the tiled
// render target the last DisplayTransfer copied out, and how that window becomes
// the landscape image the LCD shows.
//
// It exists so a caller can do two things the linear framebuffer cannot support.
// It can decode the screen from the render target *as it stands right now* —
// mid-frame, before the transfer that would deliver it — which is what lets a
// debugger scrub through a frame and watch the panel fill in. And it can run a
// pixel BACKWARDS through the transform, from the panel the player looks at to
// the render-target pixel the fragment landed on, which is what lets per-pixel
// draw provenance be shown on the screen image rather than only on the raw buffer.
type ScreenGeom struct {
	Src    uint32 // virtual address of the window's first tiled pixel
	SrcW   uint32 // the source buffer's width in pixels — its tiling stride
	W, H   uint32 // the window: W framebuffer columns (240) by H rows (400/320)
	Flip   bool
	Bottom bool
}

// Size is the landscape image's size: the framebuffer's rows become the screen's
// columns, so a 240x400 window is a 400x240 picture.
func (g ScreenGeom) Size() (w, h int) { return int(g.H), int(g.W) }

// Source maps a pixel of the landscape screen image back to the pixel of the
// tiled source buffer it was copied from — the inverse of gxDisplayTransfer's
// copy composed with Framebuffer's rotation, and it must stay the inverse of both.
func (g ScreenGeom) Source(sx, sy int) (x, y uint32, ok bool) {
	iw, ih := g.Size()
	if sx < 0 || sy < 0 || sx >= iw || sy >= ih {
		return 0, 0, false
	}
	// Framebuffer draws fb(x, y) at screen(y, W-1-x): so the screen column is the
	// fb row, and the screen row counts back along the fb column.
	fx := int(g.W) - 1 - sy
	fy := sx
	if g.Flip { // the transfer wrote its rows bottom-up
		fy = int(g.H) - 1 - fy
	}
	if fx < 0 || fy < 0 || fx >= int(g.W) || fy >= int(g.H) {
		return 0, 0, false
	}
	return uint32(fx), uint32(fy), true
}

// ScreenGeom returns the geometry of the last DisplayTransfer for a screen ("top"
// or "bottom"), or false if that screen has never been presented.
func (m *Machine) ScreenGeom(screen string) (ScreenGeom, bool) {
	rec := m.lastXferTop
	if screen == "bottom" {
		rec = m.lastXferBottom
	}
	// srcW is the tiling stride and a zero one cannot be divided by; a record from
	// an old snapshot (written before the source geometry was kept) has none, and
	// says so rather than decoding garbage.
	if rec.dst == 0 || rec.srcW == 0 || rec.w == 0 || rec.h == 0 {
		return ScreenGeom{}, false
	}
	return ScreenGeom{
		Src: rec.src, SrcW: rec.srcW, W: rec.w, H: rec.h, Flip: rec.flip,
		Bottom: screen == "bottom",
	}, true
}

// ScreenImage decodes a screen straight from its tiled source buffer, as that
// buffer stands NOW — not from the linear framebuffer the last DisplayTransfer
// wrote. The difference is the whole point: the linear framebuffer only changes
// when the game presents, so replaying a frame's command list would leave it
// frozen a frame behind, while this shows the panel filling in draw by draw.
func (m *Machine) ScreenImage(g ScreenGeom) *image.NRGBA {
	iw, ih := g.Size()
	img := image.NewNRGBA(image.Rect(0, 0, iw, ih))
	for sy := 0; sy < ih; sy++ {
		for sx := 0; sx < iw; sx++ {
			x, y, ok := g.Source(sx, sy)
			if !ok {
				continue
			}
			p := g.Src + tiledOffset(x, y, g.SrcW)
			b, gr, r := m.Read(p+1), m.Read(p+2), m.Read(p+3)
			o := img.PixOffset(sx, sy)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, gr, b, 255
		}
	}
	return img
}
