package psp

// debug.go is the PSP oracle's debugger surface: the hooks a frame debugger installs
// to watch a frame being built, and the renderers it uses to look at what the GE
// actually put in memory.
//
// It is the counterpart of the N64's OnRDPCmd/OnPixel/OnDisplay, the PSX's OnGP0 and
// the 3DS's OnPICACmd — the same three questions (what did the display processor
// execute, which command produced this pixel, when does a frame end), asked of a
// machine whose answers live in different places again. A GE frame is a display list
// of register writes, executed synchronously at sceGeListEnQueue; it ends not at a
// scanout but at sceDisplaySetFrameBuf, the moment the game says which buffer to show.
//
// The one thing to keep straight here is which buffer is which. The GE draws into the
// target its own FBP register names (geState.fbAddress), and the screen shows the one
// sceDisplaySetFrameBuf named. They are usually different — that is what double
// buffering is — and Burnout renders its reflections and its downsampled bloom target
// into buffers the screen never shows at all. RenderTarget answers the first question,
// Framebuffer the second, and confusing them is how a frame debugger ends up showing a
// picture that never changes.
//
// Everything here observes; nothing here perturbs. The hooks are nil by default and
// are not part of the savestate.

import (
	"fmt"
	"image"
)

// PixelEvent is one fragment the rasteriser produced, kept or killed.
//
// A killed fragment still carries the colour it would have had: "why is this pixel not
// what I expect" is answered by the fragment that was thrown away, not by the one that
// survived. The reject flags name the test that threw it away, in the hardware order
// putPixel runs them — scissor, alpha, stencil, depth, write mask.
type PixelEvent struct {
	Drawn         bool
	ScissorReject bool
	AlphaReject   bool
	StencilReject bool
	ZReject       bool
	MaskReject    bool
	R, G, B, A    byte
}

// Rejected reports whether the fragment was produced and then discarded.
func (e PixelEvent) Rejected() bool { return !e.Drawn }

// pixelEvent tallies one fragment and reports it, if anyone is listening. The tally is
// unconditional: an integer increment beside a fragment that has just been textured and
// blended costs nothing anyone can measure, and it means the profile does not change
// the run it is profiling.
func (m *Machine) pixelEvent(x, y int, ev PixelEvent) {
	switch {
	case ev.Drawn:
		m.geCnt.frags++
	case ev.ZReject:
		m.geCnt.zKilled++
	case ev.AlphaReject:
		m.geCnt.aKilled++
	case ev.StencilReject:
		m.geCnt.sKilled++
	case ev.ScissorReject:
		m.geCnt.scissored++
	}
	if m.OnPixel != nil && x >= 0 && y >= 0 {
		m.OnPixel(uint32(x), uint32(y), ev)
	}
}

// RunStopAfterGeCommand runs the machine until k GE command words have executed, then
// stops — the command scrubber's halt. It counts words from zero, so restoring a
// frame's start snapshot and calling this with k renders the frame exactly as it stood
// after word k-1.
//
// The machine it leaves behind is mid-list: the game submitted a display list that the
// GE abandoned half way through, and the game does not know. That is fine for what this
// is for — a scratch machine, restored from a snapshot before every replay and never
// resumed — and it is not fine for anything else.
func (m *Machine) RunStopAfterGeCommand(k int, budget uint64) Result {
	m.geLimit, m.geCount = k, 0
	r := m.Run(budget)
	m.geLimit, m.geCount = 0, 0
	return r
}

// GeCommands reports how many GE command words have executed under the current limit.
func (m *Machine) GeCommands() int { return m.geCount }

// RenderTarget is the buffer the GE is currently drawing into: its address, line stride
// and pixel format, straight out of the register file. This is the GPU's own view of
// its render target, which is not the same thing as what the screen shows.
//
// ok is false before the first display list, when the GE has no register file yet.
func (m *Machine) RenderTarget() (addr, stride, format uint32, ok bool) {
	if m.geSt == nil || m.geSt.fbStride == 0 {
		return 0, 0, 0, false
	}
	return m.geSt.fbAddress(), m.geSt.fbStride, m.geSt.fbFmt, true
}

// Scanout is the buffer the screen is showing: the one sceDisplaySetFrameBuf named.
// Read it at the flip and it is the frame the player just saw — which is the frame a
// debugger means when it says "this frame", and which is usually NOT the buffer the GE
// is bound to by then.
func (m *Machine) Scanout() (addr, stride, format uint32) {
	addr = m.fbAddr
	if addr == 0 {
		addr = vramBase
	}
	stride = m.fbWidth
	if stride == 0 {
		stride = dispW
	}
	return addr, stride, m.fbFormat
}

// DepthTarget is the depth buffer the GE is currently drawing into.
func (m *Machine) DepthTarget() (addr, stride uint32, ok bool) {
	if m.geSt == nil || m.geSt.zStride == 0 {
		return 0, 0, false
	}
	return m.geSt.zAddress(), m.geSt.zStride, true
}

// RenderSurface decodes a rectangle of VRAM as a render target, in any of the display
// pixel formats. Point it at the address the GE is configured with and you see what the
// GE is actually building, which is the picture a mid-frame replay has to show.
func (m *Machine) RenderSurface(addr, stride, format uint32, w, h int) (*image.RGBA, error) {
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("psp: render target size %dx%d out of range", w, h)
	}
	if format > psm8888 {
		return nil, fmt.Errorf("psp: 0x%X is not a display pixel format", format)
	}
	if stride == 0 {
		stride = dispW
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, m.readPixelFmt(addr, stride, format, uint32(x), uint32(y)))
		}
	}
	return img, nil
}

// RenderDepth decodes the GE's 16-bit depth buffer as greyscale. The instrument for
// "the geometry is there but nothing survives the depth test": a depth buffer that is
// uniformly one value was never written, and one full of noise was written by a pass
// that thought it owned the memory.
//
// Near is light and far is dark, because the PSP's viewport maps near to the LARGER
// depth value (Burnout's projection puts near at 65535) — so the picture reads the way
// the numbers do, rather than the way another console's would.
func (m *Machine) RenderDepth(addr, stride uint32, w, h int) (*image.NRGBA, error) {
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("psp: depth buffer size %dx%d out of range", w, h)
	}
	if stride == 0 {
		stride = dispW
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			d := u16(m, addr+(uint32(y)*stride+uint32(x))*2)
			v := byte(d >> 8)
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = v, v, v, 255
		}
	}
	return img, nil
}

// textureFormats are the GE texture formats the sampler decodes, indexed by their
// register value — the vocabulary a free surface accepts. The compressed formats
// (DXT1/3/5, 8..10) are not modelled by the rasteriser, so they are not offered here:
// a debugger must not claim to decode what the machine itself cannot sample.
var textureFormats = []string{"rgb565", "rgba5551", "rgba4444", "rgba8888", "clut4", "clut8"}

// TextureFormats lists the formats RenderTexture accepts.
func TextureFormats() []string { return append([]string(nil), textureFormats...) }

// TextureFormat maps a format name back to its GE register value.
func TextureFormat(name string) (uint32, bool) {
	for i, f := range textureFormats {
		if f == name {
			return uint32(i), true
		}
	}
	return 0, false
}

// RenderTexture decodes memory as a texture, in any of the GE's formats — the surface
// that earns its keep on this platform. Point it at the address a texture unit is
// configured with and you see whether what the unit will sample is the texture the game
// meant to upload, which no counter and no draw trace can tell you.
//
// It decodes through the machine's own sampler, so what it shows is what a draw would
// read: the same swizzle, the same CLUT. The palette is therefore the one the GE holds
// right now — an indexed texture has no colours of its own, and inventing some would be
// a picture of nothing.
//
// It refuses a format or a size the GE would not accept rather than halting the machine
// the way a real draw does: a debugger looking at the wrong address must not be able to
// kill the run it is inspecting.
func (m *Machine) RenderTexture(addr, format, w, h uint32, swizzle bool) (*image.NRGBA, error) {
	if int(format) >= len(textureFormats) {
		return nil, fmt.Errorf("psp: texture format 0x%X is not a GE format this rasteriser samples", format)
	}
	if w == 0 || h == 0 || w > 2048 || h > 2048 {
		return nil, fmt.Errorf("psp: texture size %dx%d out of range", w, h)
	}
	if format >= 4 && m.geSt == nil {
		return nil, fmt.Errorf("psp: an indexed texture needs a palette, and the GE has not loaded one yet")
	}

	// Sample through a scratch register file, so looking at a texture cannot change
	// which texels a later draw samples. The CLUT is borrowed from the live state
	// because it IS the palette the game uploaded.
	s := &geState{texAddr: addr, texFmt: format, texW: w, texH: h, texStride: w, texSwizzle: swizzle}
	if m.geSt != nil {
		s.clut = m.geSt.clut
		s.clutFmt = m.geSt.clutFmt
	}
	img := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := uint32(0); y < h; y++ {
		for x := uint32(0); x < w; x++ {
			r, g, b, a := m.sampleTexLvl(s, x, y, 0)
			o := img.PixOffset(int(x), int(y))
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = r, g, b, a
		}
	}
	return img, nil
}

// MemRegion is one mapped span of the address space.
type MemRegion struct {
	Name string
	Base uint32
	Size uint32
}

// MemRegions names the mapped regions — the memory pane's map.
func (m *Machine) MemRegions() []MemRegion {
	return []MemRegion{
		{"scratchpad", scratchBase, scratchSize},
		{"vram", vramBase, vramSize},
		{"ram", ramBase, ramSize},
	}
}

// Volume exposes the mounted UMD, for a debugger that browses it. Nil if none.
func (m *Machine) Volume() *Volume { return m.vol }
