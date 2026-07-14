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

// Framebuffer decodes what a screen ("top" or "bottom") is SHOWING, into a
// natural-orientation image, or nil if nothing has been presented yet.
//
// What the screen shows is the GSP's answer, not the last DisplayTransfer's. The two
// differ, and on a stereoscopic console they differ every frame: the top screen is
// rendered and transferred once per EYE, so the last transfer to it is the right eye —
// which with the 3D slider down can be a buffer the game cleared and never drew into. The
// panel scans the left one (Machine.Scanout). Reading the last transfer instead put an
// empty blue buffer on Super Mario 3D Land's top screen.
// PresentedImage decodes the picture ONE PASS produced: the linear framebuffer its own
// DisplayTransfer wrote. On a game that renders every screen through one buffer this is the
// only way to get a pass's picture back after a later pass has reused the render target —
// and it is what the frame the debugger just stepped actually drew, whether or not the LCD
// has swapped to it yet.
func (m *Machine) PresentedImage(g ScreenGeom) *image.NRGBA {
	if g.Dst == 0 || g.DstBPP == 0 || g.W == 0 || g.H == 0 {
		return nil
	}
	return m.decodeFB(g.Dst, g.W, g.H, g.DstStride, g.DstFmt, g.DstBPP)
}

func (m *Machine) Framebuffer(screen string) *image.NRGBA {
	rec := m.lastXferTop
	idx := 0
	if screen == "bottom" {
		rec, idx = m.lastXferBottom, 1
	}
	// The transfer records give the geometry (the panel's size and the framebuffer's
	// stride); the GSP gives the address the panel actually reads.
	if fb := m.Scanout(idx); fb.Valid && fb.AddrLeft != 0 {
		rec.dst = fb.AddrLeft
		if bpp := fbBPP(fb.Format); bpp != 0 {
			rec.bpp = bpp
			rec.format = fb.Format & 7
			if bpp != 0 && fb.Stride != 0 {
				rec.stride = fb.Stride / bpp
			}
		}
		if rec.w == 0 || rec.h == 0 {
			rec.w = 240
			rec.h = 400
			if idx == 1 {
				rec.h = 320
			}
		}
	}
	if rec.dst == 0 {
		return nil
	}
	return m.decodeFB(rec.dst, rec.w, rec.h, rec.stride, rec.format, rec.bpp)
}

// decodeFB turns a linear framebuffer into the landscape picture the panel shows. A 3DS
// framebuffer is stored rotated — 240 pixels wide by 400 (top) or 320 (bottom) lines,
// scanning the physical panel's columns — so this rotates it back.
func (m *Machine) decodeFB(addr, fw, fh, stride, format, bpp uint32) *image.NRGBA {
	w, h := int(fw), int(fh)
	img := image.NewNRGBA(image.Rect(0, 0, h, w)) // rotated: fb lines become columns
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := addr + uint32(y*int(stride)+x)*bpp
			var r, g, b byte
			switch format {
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
			// right-to-left across the landscape image, and the buffer's first line is
			// the screen's first column because the rasteriser anchors the viewport to
			// the buffer's bottom (the PICA's bottom-left origin — gpu_raster.go).
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

	// Dst is the linear framebuffer this transfer wrote, with the format it wrote in. It is
	// what says WHICH picture a pass produced, and the LCD scans out only one of them: the
	// top screen is transferred once per eye, and the eye the panel shows is the GSP's (see
	// Machine.Scanout).
	Dst       uint32
	DstStride uint32 // in pixels
	DstFmt    uint32
	DstBPP    uint32
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
		Dst:    rec.dst, DstStride: rec.stride, DstFmt: rec.format, DstBPP: rec.bpp,
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

// fbBPP is the byte width of a GSP framebuffer format word's pixel format (its low three
// bits, the same codes a DisplayTransfer's output format uses): RGBA8, RGB8, then the
// 16-bit trio. Zero for a format word that names none of them.
func fbBPP(format uint32) uint32 {
	switch format & 7 {
	case 0:
		return 4
	case 1:
		return 3
	case 2, 3, 4:
		return 2
	}
	return 0
}
