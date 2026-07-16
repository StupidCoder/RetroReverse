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
// color and/or zeta surfaces with the latched clear values (the arg is the plane
// mask — bits 4..7 select A/R/G/B, bit 0 depth, bit 1 stencil). The clear rect is
// in logical pixels; the surface's anti-aliasing mode scales it to stored samples,
// exactly as it scales draw coordinates (nv2a_raster.go).
func (g *pgraph) clearSurface(mask uint32) {
	m := g.m
	ax, ay := surfaceAAScale(g.Regs[kelvinSurfaceFormat>>2])
	rh := g.Regs[kelvinClearRectH>>2]
	rv := g.Regs[kelvinClearRectV>>2]
	x1, x2 := rh&0xFFFF*uint32(ax), (rh>>16+1)*uint32(ax)-1
	y1, y2 := rv&0xFFFF*uint32(ay), (rv>>16+1)*uint32(ay)-1

	if mask&0xF0 != 0 {
		base := g.Regs[kelvinSurfaceColorOffset>>2]
		pitch := g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
		color := g.Regs[kelvinColorClearValue>>2]
		if phys, mmio, ok := m.translate(base); ok && !mmio {
			for y := y1; y <= y2; y++ {
				row := phys + y*pitch + x1*4
				for x := x1; x <= x2; x++ {
					if row+4 > uint32(len(m.RAM)) {
						break
					}
					m.RAM[row+0] = byte(color)
					m.RAM[row+1] = byte(color >> 8)
					m.RAM[row+2] = byte(color >> 16)
					m.RAM[row+3] = byte(color >> 24)
					row += 4
					// The clear is a command that STORES PIXELS, so it reports them like
					// any draw does. Without this the debugger's provenance says "no
					// command touched this pixel" for every pixel of the frame that only
					// the clear wrote — which is most of a frame's background, and which
					// is false. The clear is very often the answer to "why is this pixel
					// this colour".
					if m.OnPixel != nil {
						m.OnPixel(x, y, PixelEvent{
							Drawn: true,
							R:     byte(color >> 16), G: byte(color >> 8),
							B: byte(color), A: byte(color >> 24),
						})
					}
				}
			}
		}
	}
	// Only the colour plane reports: a depth or stencil clear does write memory, but the
	// provenance it would claim is of a pixel in the picture, and it did not put one there.

	if mask&0x03 != 0 {
		// Z24S8: depth in the high 24 bits, stencil low 8; the latched clear value is
		// already in stored layout. A depth-only or stencil-only clear preserves the
		// other plane's bits.
		base := g.Regs[kelvinSurfaceZetaOffset>>2]
		pitch := g.Regs[kelvinSurfacePitch>>2] >> 16
		clear := g.Regs[kelvinZStencilClearValue>>2]
		keep := uint32(0)
		if mask&1 == 0 {
			keep |= 0xFFFFFF00
		}
		if mask&2 == 0 {
			keep |= 0x000000FF
		}
		if phys, mmio, ok := m.translate(base); ok && !mmio && pitch != 0 {
			for y := y1; y <= y2; y++ {
				row := phys + y*pitch + x1*4
				for x := x1; x <= x2; x++ {
					if row+4 > uint32(len(m.RAM)) {
						break
					}
					old := uint32(m.RAM[row]) | uint32(m.RAM[row+1])<<8 | uint32(m.RAM[row+2])<<16 | uint32(m.RAM[row+3])<<24
					v := clear&^keep | old&keep
					m.RAM[row+0] = byte(v)
					m.RAM[row+1] = byte(v >> 8)
					m.RAM[row+2] = byte(v >> 16)
					m.RAM[row+3] = byte(v >> 24)
					row += 4
				}
			}
		}
	}
}

// presented is the colour surface as it stood at the last FLIP_STALL — the buffer the
// title handed to the screen. recordPresented latches it there, BEFORE the method's own
// latch and before the game re-points the surface at the next buffer of its swap chain.
//
// This is how the machine knows what the TV is showing, and it needs to, because the CRTC
// scanout registers do not: see RenderPresented.
type presentedSurface struct {
	valid  bool
	base   uint32
	pitch  uint32
	w, h   int
	ax, ay int
}

func (g *pgraph) recordPresented() {
	w := int(g.Regs[kelvinSurfaceClipH>>2] >> 16)
	h := int(g.Regs[kelvinSurfaceClipV>>2] >> 16)
	pitch := g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
	base := g.Regs[kelvinSurfaceColorOffset>>2]
	if w <= 0 || h <= 0 || pitch == 0 || base == 0 {
		return
	}
	ax, ay := surfaceAAScale(g.Regs[kelvinSurfaceFormat>>2])
	g.presented = presentedSurface{valid: true, base: base, pitch: pitch, w: w, h: h, ax: ax, ay: ay}
}

// RenderPresented draws what the TV is showing: the buffer the title last presented.
//
// This is the picture, and it is deliberately NOT RenderScanout. On hardware the CRTC scans
// whatever its start register names, and Direct3D moves that register to flip. This machine
// does not model the move — and the reason is worth stating rather than working around,
// because it is a real frontier and not a shortcut:
//
//   - the title registers a scanout through AvSetDisplayMode exactly ONCE per boot, with the
//     LOADING phase's mode (320x240, pitch 1280, at 0174C000), and the 640x480 switch it
//     would make later has never happened in this HLE — it is a known-pending gap;
//   - it does write PCRTC_START (0x600800) once per vertical blank from its own ISR, but the
//     value is 0xFFFFFB00 every single time — `0 - pitch`, a constant, not an address.
//
// So the scanout registers cannot say which buffer is on screen. The flip can: the title
// says "this frame is finished and meant for the screen" with FLIP_STALL, and the colour
// surface at that instant names the buffer it means. Tracking that IS the flip, and it is
// what every emulator does when it knows the present but not the register behind it.
//
// Before the first flip there is nothing presented, and the draw target is the closest true
// answer — what the machine has drawn so far.
func (m *Machine) RenderPresented() (*image.RGBA, error) {
	p := m.pgraph.presented
	if !p.valid {
		return m.RenderDrawTarget()
	}
	return m.renderRawSurface(p.base, p.pitch, p.w, p.h, p.ax, p.ay)
}

// RenderScanout draws the framebuffer the CRTC is programmed to read — the mode the title
// registered through AvSetDisplayMode.
//
// On a machine that tracked the CRTC this would be the picture. Here it is EVIDENCE: this
// title registers its scanout once, during the loading phase, and never switches to the
// 640x480 mode it goes on to render, so what this draws is a 320x240 window on a buffer
// nothing has drawn into for millions of instructions — a blank white rectangle. It is kept,
// and kept separate from RenderPresented, precisely so that gap stays visible instead of
// being quietly papered over by the picture that works.
func (m *Machine) RenderScanout() (*image.RGBA, error) {
	if m.nv.fbAddr == 0 {
		return m.RenderDrawTarget()
	}
	pitch := m.nv.fbPitch
	w := int(pitch / 4)
	// The registered mode does not carry the line count directly; every AV mode the
	// XDK sets is 4:3 (480-line family), so derive the height from the width.
	h := w * 3 / 4
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("nv2a: no display mode programmed (pitch=%d)", pitch)
	}
	return m.renderRawSurface(m.nv.fbAddr, pitch, w, h, 1, 1)
}

// RenderDrawTarget draws the Kelvin color surface — the frame being built — downsampled
// from stored samples to logical pixels when the surface is anti-aliased (a box
// filter over the AA footprint, which is the resolve a filtered flip performs).
func (m *Machine) RenderDrawTarget() (*image.RGBA, error) {
	g := m.pgraph
	w := int(g.Regs[kelvinSurfaceClipH>>2] >> 16)
	h := int(g.Regs[kelvinSurfaceClipV>>2] >> 16)
	pitch := g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
	base := g.Regs[kelvinSurfaceColorOffset>>2]
	ax, ay := surfaceAAScale(g.Regs[kelvinSurfaceFormat>>2])
	if w <= 0 || h <= 0 || pitch == 0 || base == 0 {
		return nil, fmt.Errorf("nv2a: no render surface programmed (w=%d h=%d pitch=%d base=%08X)", w, h, pitch, base)
	}
	return m.renderRawSurface(base, pitch, w, h, ax, ay)
}

// FramePNG / SurfacePNG encode the two pictures above. They are what the oracle's
// -png / -surfpng export, and the md5s the writeup pins are of exactly these bytes.
func (m *Machine) FramePNG() ([]byte, error) { return encodePNG(m.RenderScanout()) }

func (m *Machine) SurfacePNG() ([]byte, error) { return encodePNG(m.RenderDrawTarget()) }

func encodePNG(img *image.RGBA, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderRawSurface reads an A8R8G8B8 surface at base/pitch, averaging ax x ay sample
// blocks into each output pixel.
func (m *Machine) renderRawSurface(base, pitch uint32, w, h, ax, ay int) (*image.RGBA, error) {
	phys, mmio, ok := m.translate(base)
	if !ok || mmio {
		return nil, fmt.Errorf("nv2a: surface at %08X is not RAM", base)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	n := uint32(ax * ay)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, gg, b uint32
			for sy := 0; sy < ay; sy++ {
				row := phys + uint32(y*ay+sy)*pitch
				for sx := 0; sx < ax; sx++ {
					o := row + uint32(x*ax+sx)*4
					if o+4 > uint32(len(m.RAM)) {
						continue
					}
					// A8R8G8B8 little-endian in RAM: B,G,R,A
					b += uint32(m.RAM[o])
					gg += uint32(m.RAM[o+1])
					r += uint32(m.RAM[o+2])
				}
			}
			i := img.PixOffset(x, y)
			img.Pix[i+0] = byte(r / n)
			img.Pix[i+1] = byte(gg / n)
			img.Pix[i+2] = byte(b / n)
			img.Pix[i+3] = 0xFF // the TV ignores alpha
		}
	}
	return img, nil
}
