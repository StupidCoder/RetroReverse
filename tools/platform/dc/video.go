package dc

// video.go renders what the CRTC would scan out: the framebuffer the FB_R_*
// registers describe, read straight from VRAM. No PVR is needed for this —
// a boot that writes pixels through the 32-bit window (or, later, a PVR that
// has rendered a frame into VRAM) becomes visible here.
//
// "Which buffer is the frame?" has a different answer on every platform; on
// the Dreamcast the honest answer for this milestone is the *read* pointer
// (FB_R_SOF1), because that is the picture the TV would show. An unconfigured
// framebuffer is an error, never a black rectangle — "drew nothing" must not
// be mistakable for "drew black".

import (
	"fmt"
	"image"
)

// PVR register offsets video.go interprets (from pvrBase).
const (
	fbRCtrl = 0x44 / 4
	fbRSOF1 = 0x50 / 4
	fbRSize = 0x5C / 4
)

// RenderFB decodes the scanout framebuffer into an image.
func (m *Machine) RenderFB() (*image.RGBA, error) {
	ctrl := m.PVRRegs[fbRCtrl]
	if ctrl&1 == 0 {
		return nil, fmt.Errorf("no framebuffer: FB_R_CTRL enable is clear (nothing has configured video yet)")
	}
	depth := (ctrl >> 2) & 3
	size := m.PVRRegs[fbRSize]
	xUnits := int(size&0x3FF) + 1        // 32-bit units per line
	lines := int((size>>10)&0x3FF) + 1   // lines per field
	modulus := int((size>>20)&0x3FF) - 1 // 32-bit units skipped between lines
	base := m.PVRRegs[fbRSOF1] & 0x00FFFFFF

	bytesPerLine := xUnits * 4
	var bpp int
	switch depth {
	case 0, 1:
		bpp = 2
	case 2:
		bpp = 3
	default:
		bpp = 4
	}
	w := bytesPerLine / bpp
	if w <= 0 || lines <= 0 {
		return nil, fmt.Errorf("degenerate framebuffer: FB_R_SIZE=%08X", size)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, lines))
	off := int(base)
	for y := 0; y < lines; y++ {
		for x := 0; x < w; x++ {
			p := off + x*bpp
			if p+bpp > len(m.VRAM) {
				return nil, fmt.Errorf("framebuffer overruns VRAM at line %d", y)
			}
			// FB_R_SOF is a 32-bit-path address; fetch each byte through the
			// bank interleave.
			b0 := m.VRAM[vram32to64(uint32(p))]
			b1 := m.VRAM[vram32to64(uint32(p+1))]
			var r, g, b uint8
			switch depth {
			case 0: // RGB555, little-endian
				v := uint16(b0) | uint16(b1)<<8
				r = uint8(v>>10&31) << 3
				g = uint8(v>>5&31) << 3
				b = uint8(v&31) << 3
			case 1: // RGB565
				v := uint16(b0) | uint16(b1)<<8
				r = uint8(v>>11&31) << 3
				g = uint8(v>>5&63) << 2
				b = uint8(v&31) << 3
			default: // RGB888 / 0RGB8888, packed
				b, g, r = b0, b1, m.VRAM[vram32to64(uint32(p+2))]
			}
			i := img.PixOffset(x, y)
			img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, 0xFF
		}
		off += bytesPerLine + modulus*4
	}
	return img, nil
}

// RenderDrawTarget decodes the WRITE framebuffer — the buffer renderFrame
// just drew into, addressed by FB_W_SOF1 before the guest flips FB_R_SOF1
// onto it. This is the debugger's frame picture ("render the DRAW TARGET,
// not the scanout"): captured inside OnRender it shows the frame that was
// just built, one flip ahead of what RenderFB scans out. The decode mirrors
// the rasteriser exactly — tightly packed 565 at the FB_R_SIZE dimensions,
// each byte through the bank interleave — so it shows what plot() wrote,
// not an interpretation of it.
func (m *Machine) RenderDrawTarget() (*image.RGBA, error) {
	if m.PVRRegs[0x48/4]&7 != 1 {
		return nil, fmt.Errorf("draw target: FB_W_CTRL packmode %d unimplemented (only 565)", m.PVRRegs[0x48/4]&7)
	}
	size := m.PVRRegs[fbRSize]
	w := (int(size&0x3FF) + 1) * 2
	h := int(size>>10&0x3FF) + 1
	base := m.PVRRegs[0x60/4] & 0x00FFFFFF
	if w <= 0 || h <= 0 || size == 0 {
		return nil, fmt.Errorf("draw target: degenerate FB_R_SIZE=%08X", size)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := base + uint32(y*w+x)*2
			if off+2 > VRAMSize {
				continue // out of VRAM: the rasteriser never wrote it either
			}
			off = vram32to64(off)
			v := uint16(m.VRAM[off]) | uint16(m.VRAM[off+1])<<8
			i := img.PixOffset(x, y)
			img.Pix[i+0] = uint8(v>>11&31) << 3
			img.Pix[i+1] = uint8(v>>5&63) << 2
			img.Pix[i+2] = uint8(v&31) << 3
			img.Pix[i+3] = 0xFF
		}
	}
	return img, nil
}

// RenderVRAM draws an arbitrary VRAM window as RGB565 — the raw viewer for
// debugging what the framebuffer registers have not yet named.
func (m *Machine) RenderVRAM(offset uint32, w, h int) (*image.RGBA, error) {
	if w <= 0 || h <= 0 || int(offset)+w*h*2 > len(m.VRAM) {
		return nil, fmt.Errorf("VRAM window %08X %dx%d out of range", offset, w, h)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := int(offset) + (y*w+x)*2
			v := uint16(m.VRAM[p]) | uint16(m.VRAM[p+1])<<8
			i := img.PixOffset(x, y)
			img.Pix[i+0] = uint8(v>>11&31) << 3
			img.Pix[i+1] = uint8(v>>5&63) << 2
			img.Pix[i+2] = uint8(v&31) << 3
			img.Pix[i+3] = 0xFF
		}
	}
	return img, nil
}
