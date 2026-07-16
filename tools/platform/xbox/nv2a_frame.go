package xbox

// nv2a_frame.go exports the rendered frame. The render target's geometry comes from
// the Kelvin surface state the title itself programmed (the survey's device-init
// set): SET_SURFACE_CLIP_HORIZONTAL/VERTICAL (0x200/0x204: x|width<<16, y|height<<16),
// SET_SURFACE_PITCH (0x20C: color in the low word, zeta in the high), and
// SET_SURFACE_COLOR_OFFSET (0x210, a physical address through the base-0 color DMA
// context). OutRun programs 640x480, pitch 2560 — A8R8G8B8 — at 0x0174C000, the same
// address its swap registers with AvSetDisplayMode.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

// Kelvin surface-state registers (byte offsets; the Regs file is indexed >>2).
const (
	kelvinSurfaceClipH       = 0x0200
	kelvinSurfaceClipV       = 0x0204
	kelvinSurfaceFormat      = 0x0208
	kelvinSurfacePitch       = 0x020C
	kelvinSurfaceColorOffset = 0x0210
	kelvinSurfaceZetaOffset  = 0x0214

	kelvinZStencilClearValue = 0x1D8C
	kelvinColorClearValue    = 0x1D90
	kelvinClearSurface       = 0x1D94
	kelvinClearRectH         = 0x1D98 // x1 | x2<<16 (inclusive)
	kelvinClearRectV         = 0x1D9C // y1 | y2<<16 (inclusive)
)

// clearSurface runs the CLEAR_SURFACE trigger: fill the clear rectangle of the
// color surface with the latched clear color (the arg is the plane mask — bits
// 4..7 select A/R/G/B, bits 0/1 depth/stencil). The zeta planes are not modelled
// yet: nothing reads depth back until the rasteriser lands, and an untouched
// zeta surface cannot leak into the exported color frame.
func (g *pgraph) clearSurface(mask uint32) {
	if mask&0xF0 == 0 {
		return // depth/stencil-only clear
	}
	m := g.m
	base := g.Regs[kelvinSurfaceColorOffset>>2]
	pitch := g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
	rh := g.Regs[kelvinClearRectH>>2]
	rv := g.Regs[kelvinClearRectV>>2]
	x1, x2 := rh&0xFFFF, rh>>16
	y1, y2 := rv&0xFFFF, rv>>16
	color := g.Regs[kelvinColorClearValue>>2]

	phys, mmio, ok := m.translate(base)
	if !ok || mmio {
		return
	}
	for y := y1; y <= y2; y++ {
		row := phys + y*pitch + x1*4
		for x := x1; x <= x2; x++ {
			if row+4 > uint32(len(m.RAM)) {
				return
			}
			m.RAM[row+0] = byte(color)
			m.RAM[row+1] = byte(color >> 8)
			m.RAM[row+2] = byte(color >> 16)
			m.RAM[row+3] = byte(color >> 24)
			row += 4
		}
	}
}

// FramePNG encodes the current color surface as a PNG. The surface address comes
// from the display scanout AvSetDisplayMode registered when available (the frame
// the TV would show), else from the Kelvin color offset (the frame being drawn).
func (m *Machine) FramePNG() ([]byte, error) {
	g := m.pgraph
	w := int(g.Regs[kelvinSurfaceClipH>>2] >> 16)
	h := int(g.Regs[kelvinSurfaceClipV>>2] >> 16)
	pitch := g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
	base := m.nv.fbAddr
	if base == 0 {
		base = g.Regs[kelvinSurfaceColorOffset>>2]
	}
	if w <= 0 || h <= 0 || pitch == 0 || base == 0 {
		return nil, fmt.Errorf("nv2a: no render surface programmed (w=%d h=%d pitch=%d base=%08X)", w, h, pitch, base)
	}
	phys, mmio, ok := m.translate(base)
	if !ok || mmio {
		return nil, fmt.Errorf("nv2a: render surface at %08X is not RAM", base)
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := phys + uint32(y)*pitch
		for x := 0; x < w; x++ {
			o := row + uint32(x)*4
			if o+4 > uint32(len(m.RAM)) {
				break
			}
			// A8R8G8B8 little-endian in RAM: B,G,R,A
			i := img.PixOffset(x, y)
			img.Pix[i+0] = m.RAM[o+2]
			img.Pix[i+1] = m.RAM[o+1]
			img.Pix[i+2] = m.RAM[o+0]
			img.Pix[i+3] = 0xFF // the TV ignores alpha
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
