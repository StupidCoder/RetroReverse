package dsmachine

// debug.go is the DS oracle's debugger surface: the hooks a frame debugger installs
// to watch a frame being built, and the halts it needs to stop the machine in the
// middle of one.
//
// It is the counterpart of the N64's OnRDPCmd/OnPixel and the 3DS's OnPICACmd — the
// same three questions (what did the display processor execute, which command
// produced this pixel, when does a frame end), asked of a machine whose answers live
// somewhere else again.
//
// The DS's answer to the third question is the interesting one. On the N64 and the
// PSX the rasteriser runs as the commands arrive, so "the framebuffer after command
// k" is simply the framebuffer, halted. The DS does not work that way: the geometry
// engine ACCUMULATES polygons all frame and the rasteriser does not run until the
// buffer swap. There is no partial framebuffer to stop at, because there is no
// framebuffer yet.
//
// So the DS's scrubber renders what the geometry submitted SO FAR: replay to command
// k, rasterise the polygon list as it stands, compose the screens. Dragging the
// scrubber is watching a frame's polygons accumulate, which is exactly the question a
// DS frame raises — the geometry is the frame, and the rasteriser is a formality that
// happens once at the end.
//
// Everything here observes; nothing here perturbs. The hooks are nil by default and
// are outside the savestate.

import (
	"fmt"
	"image"
	"image/color"
)

// PixelEvent is one fragment the rasteriser produced, kept or killed.
//
// A depth-killed fragment carries no colour — the DS kills it before the texture is
// even fetched, so there is nothing to report but the rejection. An alpha-killed one
// carries the colour it would have had. A drawn one carries the colour after blending,
// which is what actually reached the colour buffer.
type PixelEvent struct {
	Drawn                bool
	ZReject, AlphaReject bool
	R, G, B, A           uint8 // the DS's six bits per channel, and five of alpha
}

// gxHalt is the sentinel the command scrubber throws to stop the machine exactly
// after command k.
//
// A panic, rather than a flag checked between instructions, because on the DS a
// single instruction can execute a hundred commands. The geometry engine is fed by
// DMA, and one mode-7 burst pushes 112 words into the GXFIFO inside the store that
// triggered it — so a scrubber that could only stop at instruction boundaries would
// overshoot its position by most of a display list, and the picture it drew would not
// be the picture it claimed. The panic unwinds out of the middle of the DMA, which is
// safe precisely because the machine it unwinds is a scratch replay that is about to
// be thrown away.
type gxHalt struct{}

// RunStopAfterGXCommand replays until k geometry commands have executed, then stops.
// It counts from zero, so restoring a frame's start snapshot and calling this with k
// leaves the machine holding the geometry the frame had after command k-1.
//
// It does not rasterise: call ForceRender for that. The two are separate because the
// halt is exact and the render is a choice — the scrubber wants the partial picture,
// but a caller asking "what geometry had arrived by command k" does not.
func (m *Machine) RunStopAfterGXCommand(k int, budget uint64, quantum int) {
	m.gpu3d.limit, m.gpu3d.count = k, 0
	defer func() {
		m.gpu3d.limit, m.gpu3d.count = -1, 0
		if r := recover(); r != nil {
			if _, ours := r.(gxHalt); !ours {
				panic(r)
			}
		}
	}()
	m.Run(budget, quantum, nil)
}

// GXCommandsRun reports how many geometry commands have executed under the current
// limit — the scrubber's progress through a frame.
func (m *Machine) GXCommandsRun() int { return m.gpu3d.count }

// ForceRender rasterises the polygon list as it currently stands and composes both
// screens from it, without waiting for the buffer swap that would normally do it.
//
// This is what makes a DS command scrubber possible at all (see the file comment).
// It is also, deliberately, what a mid-frame screenshot means: the picture the frame
// would show if the game stopped submitting geometry now.
func (m *Machine) ForceRender() {
	m.gpu3d.render(m)
	m.gpu2d.render(m)
}

// StopRequested asks the scheduler to stop at the next instruction boundary. The
// frame debugger sets it from OnFrame to halt at a VBlank.
func (m *Machine) StopRequested() { m.stop = true }

// AddBreakpoint, ClearBreakpoint and ClearBreakpoints hold ARM9 execution
// breakpoints. Stopped reports whether the last run ended on one.
func (m *Machine) AddBreakpoint(pc uint32) {
	if m.bps == nil {
		m.bps = map[uint32]bool{}
	}
	m.bps[pc] = true
}

func (m *Machine) ClearBreakpoint(pc uint32)  { delete(m.bps, pc) }
func (m *Machine) ClearBreakpoints()          { m.bps = nil }
func (m *Machine) Stopped() (bool, uint32)    { return m.stopped, m.stoppedPC }
func (m *Machine) StepInstructions(n int) int { return m.runInstrs(n) }

// MemRegion is one named span of the address space.
type MemRegion struct {
	Name string
	Base uint32
	Size uint32
}

// MemRegions names the DS's memory map, for a debugger's memory pane. Unlike the
// 3DS's — which is sparse and laid out by a loader, and so has to be asked for — a
// DS's map is the machine's, fixed in silicon, and the only thing that moves is where
// the ARM9 put its tightly-coupled memories.
func (m *Machine) MemRegions() []MemRegion {
	r := []MemRegion{
		{"main RAM", mainBase, mainSize},
		{"shared WRAM", swramBase, swramSize},
		{"ARM7 WRAM", wram7Base, wram7Size},
		{"I/O", 0x04000000, 0x00100000},
		{"palette", palBase, palSize},
		{"VRAM (engine A BG)", 0x06000000, 0x80000},
		{"VRAM (engine B BG)", 0x06200000, 0x20000},
		{"VRAM (engine A OBJ)", 0x06400000, 0x40000},
		{"VRAM (engine B OBJ)", 0x06600000, 0x20000},
		{"VRAM (LCDC window)", 0x06800000, 0xA4000},
		{"OAM", oamBase, oamSize},
	}
	if m.ARM9.itcm != nil {
		r = append(r, MemRegion{"ITCM", m.ARM9.itcmBase, uint32(len(m.ARM9.itcm))})
	}
	if m.ARM9.dtcm != nil {
		r = append(r, MemRegion{"DTCM", m.ARM9.dtcmBase, uint32(len(m.ARM9.dtcm))})
	}
	return r
}

// --- surfaces ----------------------------------------------------------------

// TextureFormats are the DS's seven texture formats, indexed by their TEXIMAGE_PARAM
// value — the vocabulary a free surface accepts.
var textureFormats = []string{
	"none", "a3i5", "4-colour", "16-colour", "256-colour", "4x4-compressed", "a5i3", "direct16",
}

// TextureFormats lists the formats RenderTexture accepts.
func TextureFormats() []string { return append([]string(nil), textureFormats...) }

// TextureFormat maps a format name back to its register value.
func TextureFormat(name string) (uint32, bool) {
	for i, f := range textureFormats {
		if f == name {
			return uint32(i), true
		}
	}
	return 0, false
}

// RenderTexture decodes the 3D texture memory as a texture, in any of the DS's
// formats. Point it at the VRAM offset a TEXIMAGE_PARAM names, with the palette base
// a PLTT_BASE names, and you see whether what the texture unit will sample is the
// texture the game meant to upload — which no draw trace and no counter can tell you,
// and which is exactly the question a wrongly-coloured model raises.
//
// Addresses here are offsets into the texture *space* (what the banks mapped with
// MST=3 present), not CPU addresses: that is what the register holds, and translating
// it for the debugger would only invite the reader to look up the wrong number.
func (m *Machine) RenderTexture(offset, palBase, format, w, h uint32) (*image.RGBA, error) {
	if int(format) >= len(textureFormats) {
		return nil, errf("dsmachine: texture format %d is not a DS format", format)
	}
	if w == 0 || h == 0 || w > 1024 || h > 1024 {
		return nil, errf("dsmachine: texture size %dx%d out of range", w, h)
	}
	// A fake polygon whose texture parameters are the ones asked for. Going through the
	// rasteriser's own sampler rather than a second decoder written for the debugger is
	// the point: a debugger that decodes textures its own way can agree with itself and
	// disagree with the machine, and then it is lying about the very thing it exists to
	// show.
	st := texState{
		format: format, base: offset, pal: palBase,
		sizeS: int(w), sizeT: int(h),
		repeatS: true, repeatT: true,
	}
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			r, g, b, a, ok := m.gpu3d.sampleTex(m, &st, x, y)
			c := color.RGBA{}
			if ok {
				// The sampler works in the DS's six bits per channel and five of alpha.
				c = color.RGBA{expand6(r), expand6(g), expand6(b), expand5(a)}
			}
			img.Set(x, y, c)
		}
	}
	return img, nil
}

// RenderDepth decodes the 3D engine's depth buffer as greyscale — near is dark, far
// is light. The instrument for "the geometry is there and nothing survives the depth
// test": a depth buffer that is uniformly one value was never written, and one full of
// noise was written by a pass that thought it owned the memory.
func (m *Machine) RenderDepth() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, rastW, rastH))
	for y := 0; y < rastH; y++ {
		for x := 0; x < rastW; x++ {
			v := byte(m.gpu3d.rast.depth[y*rastW+x] >> 16) // top 8 of the 24-bit depth
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
	return img
}

// VRAMBank exposes one of the nine banks by index (0 = A .. 8 = I), as the bytes
// actually are — whatever the bank is currently mapped as.
func (m *Machine) VRAMBank(i int) []byte {
	if i < 0 || i >= 9 {
		return nil
	}
	return m.vram.bank[i]
}

// VRAMBankInfo reports a bank's name, size and VRAMCNT setting.
func (m *Machine) VRAMBankInfo(i int) (name string, size int, cnt uint8) {
	if i < 0 || i >= 9 {
		return "", 0, 0
	}
	return string(rune('A' + i)), bankSizes[i], m.vram.cnt[i]
}

// errf is fmt.Errorf, kept local so this file's imports stay to what it draws with.
func errf(format string, a ...interface{}) error { return fmt.Errorf(format, a...) }

func expand6(v uint8) uint8 {
	if v > 63 {
		v = 63
	}
	return v<<2 | v>>4
}

func expand5(v uint8) uint8 {
	if v > 31 {
		v = 31
	}
	return v<<3 | v>>2
}

// ThreeDVisible reports, for each pixel of engine A's panel, whether the 3D layer won
// the priority fight there — whether the player is actually looking at the 3D engine's
// work, or at a 2D background sitting on top of it.
//
// A frame debugger's provenance is only honest where this is true. The rasteriser knows
// which command drew each pixel of the 3D plane; it does not know that a background
// later covered half of them, and a debugger that answers "this draw made that pixel"
// for a pixel the player cannot see is giving a confident wrong answer.
func (m *Machine) ThreeDVisible() []bool {
	return m.gpu2d.a.vis3D[:]
}

// EngineAOnTop reports which physical panel engine A — the only engine with a 3D layer —
// is currently driving. POWCNT1 bit 15 decides it and a game may flip it, so "the 3D is
// on the top screen" is a question about a register, not a constant.
func (m *Machine) EngineAOnTop() bool { return m.powcnt&(1<<15) != 0 }

// Draw3DInto renders the 3D engine's own plane into img — what the rasteriser produced,
// before the 2D engine composited it as engine A's BG0.
//
// A pixel the rasteriser never touched is left transparent rather than black, which is
// the distinction the surface exists to make: a black screen is two entirely different
// bugs wearing one face (the GPU drew nothing / the GPU drew and the compositor threw it
// away), and no counter separates them.
func (m *Machine) Draw3DInto(img *image.RGBA) error {
	px := m.gpu2d.threeD
	if len(px) < rastW*rastH {
		return errf("dsmachine: the 3D engine has not drawn a frame")
	}
	for y := 0; y < rastH; y++ {
		for x := 0; x < rastW; x++ {
			v := px[y*rastW+x]
			img.Set(x, y, color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)})
		}
	}
	return nil
}
