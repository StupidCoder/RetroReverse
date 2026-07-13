package n3ds

// debug.go is the 3DS oracle's debugger surface: the hooks a frame debugger
// installs to watch a frame being built, and the renderers it uses to look at
// what the PICA200 actually put in memory.
//
// It is the counterpart of the N64's OnRDPCmd/OnPixel/OnDisplay and the PSX's
// OnGP0 — the same three questions (what did the display processor execute,
// which command produced this pixel, when does a frame end), asked of a machine
// whose answers live in different places: a PICA frame is a register-write
// stream, not a packet list, and it ends at the GSP's buffer swap rather than a
// scanout.
//
// Everything here observes; nothing here perturbs. The hooks are nil by default,
// they are outside the savestate, and the free-surface texture renderer evicts
// whatever it seeds into the texture cache, so looking at a texture cannot
// change which texels a later draw samples.

import (
	"fmt"
	"image"
)

// PixelEvent is one fragment the rasteriser produced, kept or killed.
//
// A depth-killed fragment carries no colour: the PICA kills it before the TEV
// combiner runs, so there is nothing to report but the rejection. An
// alpha-tested one does carry the colour it would have had, and a drawn one
// carries the colour after blending — which is what actually reached memory.
type PixelEvent struct {
	Drawn                bool
	ZReject, AlphaReject bool
	R, G, B, A           uint8
}

func (g *GPU) pixelEvent(x, y uint32, ev PixelEvent) {
	if g.m.OnPixel != nil {
		g.m.OnPixel(x, y, ev)
	}
}

// RunStopAfterPICACommand runs the machine until k command-list register writes
// have executed, then stops — the command scrubber's halt. It counts writes from
// zero, so restoring a frame's start snapshot and calling this with k renders the
// frame exactly as it stood after write k-1.
//
// A frame whose list is shorter than k simply runs out of budget or reaches the
// next frame; the caller renders whatever the target holds either way.
func (m *Machine) RunStopAfterPICACommand(k, budget int) int {
	m.picaLimit, m.picaCount = k, 0
	n := m.Run(budget)
	m.picaLimit, m.picaCount = 0, 0
	return n
}

// PICACommands reports how many command-list writes have executed under the
// current limit — the scrubber's progress through a list.
func (m *Machine) PICACommands() int { return m.picaCount }

// ClearBreakpoints drops every PC breakpoint. AddBreakpoint (run.go) sets them;
// the debugger re-applies the set it owns after a clear.
func (m *Machine) ClearBreakpoints() { m.bps = map[uint32]bool{} }

// Stopped reports whether the last run ended at a breakpoint.
func (m *Machine) Stopped() bool { return m.stopped }

// MemRegion is one mapped span of the address space.
type MemRegion struct {
	Name string
	Base uint32
	Size uint32
}

// MemRegions names the mapped regions, in map order — the memory pane's map. The
// 3DS process address space is sparse and laid out by the loader (code, stack,
// heap, linear, VRAM, the config and TLS pages, DSP RAM), so it has to be asked
// for rather than hard-coded the way a cartridge machine's is.
func (m *Machine) MemRegions() []MemRegion {
	out := make([]MemRegion, 0, len(m.regions))
	for _, r := range m.regions {
		out = append(out, MemRegion{Name: r.name, Base: r.base, Size: uint32(len(r.data))})
	}
	return out
}

// RomFS exposes the cartridge filesystem, for a debugger that browses it. Nil if
// the image has no RomFS.
func (m *Machine) RomFS() *RomFS { return m.romfs }

// ColorTarget is the colour buffer the PICA is currently drawing into: its
// virtual address and dimensions, straight out of the register file. This is the
// GPU's own view of its render target, which is not the same thing as what the
// screen shows — a DisplayTransfer has to move it there, and the gap between the
// two is exactly where a frame goes missing.
func (g *GPU) ColorTarget() (addr, w, h uint32) {
	fb := g.fbstate()
	return fb.colorAddr, fb.width, fb.height
}

// DepthTarget is the depth buffer the PICA is currently drawing into.
func (g *GPU) DepthTarget() (addr, w, h uint32) {
	fb := g.fbstate()
	return fb.depthAddr, fb.width, fb.height
}

// RenderDepth decodes a tiled D24S8 depth buffer as greyscale — near is dark,
// far is light. The instrument for "the geometry is there but nothing survives
// the depth test": a depth buffer that is uniformly one value was never written,
// and one full of noise was written by a pass that thought it owned the memory.
func (m *Machine) RenderDepth(addr, w, h uint32) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := uint32(0); y < h; y++ {
		for x := uint32(0); x < w; x++ {
			p := addr + tiledOffset(x, y, w)
			d := uint32(m.Read(p)) | uint32(m.Read(p+1))<<8 | uint32(m.Read(p+2))<<16
			v := byte(d >> 16) // the top 8 bits of the 24-bit depth
			o := img.PixOffset(int(x), int(y))
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = v, v, v, 255
		}
	}
	return img
}

// TextureFormats are the PICA200 texture formats, indexed by their register
// value — the vocabulary a free surface accepts.
var textureFormats = []string{
	"rgba8", "rgb8", "rgba5551", "rgb565", "rgba4", "la8", "hilo8",
	"l8", "a8", "la4", "l4", "a4", "etc1", "etc1a4",
}

// TextureFormats lists the formats RenderTexture accepts.
func TextureFormats() []string { return append([]string(nil), textureFormats...) }

// TextureFormat maps a format name back to its PICA register value.
func TextureFormat(name string) (uint32, bool) {
	for i, f := range textureFormats {
		if f == name {
			return uint32(i), true
		}
	}
	return 0, false
}

// RenderTexture decodes memory as a texture, in any of the PICA's formats — the
// surface that earns its keep on this platform. Point it at the address a
// texture unit is configured with and you see whether what the unit will sample
// is the texture the game meant to upload, which no counter and no draw trace
// can tell you.
//
// It refuses a format or a size the GPU would not accept rather than halting the
// machine the way a real draw does: a debugger looking at the wrong address must
// not be able to kill the run it is inspecting. It also leaves the texture cache
// exactly as it found it, so what it decodes cannot change what a later draw
// samples.
func (m *Machine) RenderTexture(addr, format, w, h uint32) (*image.NRGBA, error) {
	if int(format) >= len(textureFormats) {
		return nil, fmt.Errorf("n3ds: texture format 0x%X is not a PICA format", format)
	}
	if w == 0 || h == 0 || w > 2048 || h > 2048 {
		return nil, fmt.Errorf("n3ds: texture size %dx%d out of range", w, h)
	}
	g := m.gpu
	k := texKey{addr, format, w, h}
	_, cached := g.texCache[k]
	tex, ok := g.texture(addr, format, w, h)
	if !cached {
		delete(g.texCache, k) // do not seed the cache the game samples from
	}
	if !ok {
		return nil, fmt.Errorf("n3ds: cannot decode a %s texture at 0x%08X", textureFormats[format], addr)
	}
	img := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	copy(img.Pix, tex.pix)
	return img, nil
}
