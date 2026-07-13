package threedo

// debug.go is the 3DO oracle's debugger surface: the hooks a frame debugger installs to
// watch a frame being built, and the renderers it uses to look at what the cel engine
// actually put in memory.
//
// It is the counterpart of the N64's OnRDPCmd/OnPixel/OnDisplay, the PSX's OnGP0, the
// 3DS's OnPICACmd and the PSP's OnGeCmd — the same three questions (what did the display
// processor execute, which command produced this pixel, when does a frame end), asked of
// a machine that answers them in a shape none of the others do.
//
// There is no display list here and no register file. The 3DO's cel engine is handed a
// CCB — a Cel Control Block — and walks a linked list of them, and the CCB *is* the
// command: its flags, its source pointer, its position and its four projection deltas
// are the whole instruction. So a "command" on this machine is one cel, and the words a
// debugger shows are the CCB's own, read back out of the game's memory where the game
// wrote them.
//
// One honest absence, and it is the PSX's too: this rasteriser has no depth test and no
// alpha test. A pixel it reports is a pixel it wrote. A cel's transparent texels are
// decided before a destination pixel exists at all — they are source pixels that produce
// nothing, not fragments that were produced and thrown away — so the overdraw pane will
// never show a "rejected" write here, and claiming otherwise would be an invention.
//
// Everything here observes; nothing here perturbs. The hooks are nil by default and are
// not part of the savestate.

import (
	"fmt"
	"image"
)

// CelDraw is one cel the engine is about to draw — this machine's drawing command.
//
// The CCB fields are the ones the hardware reads (graphicsfolio.go's drawOneCel reads
// exactly these): the flags, the resolved source and palette pointers, the 16.16
// position, and the projection deltas that make a cel a quad rather than a sprite. HDDX
// and HDDY are what shear each row — they are why Need for Speed's road has perspective.
type CelDraw struct {
	Index int    // the cel's position in the frame, counted from the frame's first
	CCB   uint32 // where the control block lives in memory
	Flags uint32
	Src   uint32 // resolved source-data pointer
	PLUT  uint32 // resolved palette pointer (0 if the cel is not coded)

	XPos, YPos            int32 // 16.16 destination position
	HDX, HDY              int32 // 12.20 per-column deltas
	VDX, VDY              int32 // 16.16 per-row deltas
	HDDX, HDDY            int32 // 12.20 second-order (the perspective shear)
	PIXC                  uint32
	PRE0, PRE1            uint32
	BPP, Width, Height    int
	Packed, Coded, LRForm bool
	Bitmap                uint32 // the bitmap buffer it draws into
	BitmapW, BitmapH      int
}

// Name is the cel's kind, for a command list.
func (c CelDraw) Name() string {
	switch {
	case c.LRForm:
		return "CEL LRFORM16"
	case c.Packed:
		return "CEL packed"
	default:
		return "CEL unpacked"
	}
}

// Decoded is a one-line human decode of the cel.
func (c CelDraw) Decoded() string {
	persp := ""
	if c.HDDX != 0 || c.HDDY != 0 || c.VDX != 0 || c.HDY != 0 {
		persp = " persp"
	}
	return fmt.Sprintf("%dbpp %dx%d at (%d,%d)%s pixc=%08X flags=%08X",
		c.BPP, c.Width, c.Height, c.XPos>>16, c.YPos>>16, persp, c.PIXC, c.Flags)
}

// Words are the CCB's own words, in the order the hardware reads them — what a debugger
// shows when it is asked what this command actually said.
func (c CelDraw) Words() []uint64 {
	return []uint64{
		uint64(c.Flags), uint64(c.Src), uint64(c.PLUT),
		uint64(uint32(c.XPos)), uint64(uint32(c.YPos)),
		uint64(uint32(c.HDX)), uint64(uint32(c.HDY)),
		uint64(uint32(c.VDX)), uint64(uint32(c.VDY)),
		uint64(uint32(c.HDDX)), uint64(uint32(c.HDDY)),
		uint64(c.PIXC), uint64(c.PRE0), uint64(c.PRE1),
	}
}

// PixelEvent is one pixel the cel engine wrote. There is no rejection on this machine —
// see the note at the top of the file — so Drawn is always true, and the field is here
// because the debugger's vocabulary has it, not because this rasteriser needs it.
type PixelEvent struct {
	Drawn      bool
	R, G, B, A uint8
}

// celDraw reports a cel about to be drawn, if anyone is listening. It re-reads the CCB
// rather than being handed drawOneCel's locals, so what the debugger shows is what is in
// the game's memory — and it costs nothing when nobody is looking.
func (m *Machine) celDraw(bm gfxBitmap, ccb, flags, src, plutPtr uint32) {
	m.celCnt.cels++
	if m.OnCel == nil {
		return
	}
	c := CelDraw{
		Index: m.celCount - 1, // celCount was incremented by the caller
		CCB:   ccb, Flags: flags, Src: src, PLUT: plutPtr,
		XPos: int32(m.read32(ccb + 0x10)), YPos: int32(m.read32(ccb + 0x14)),
		HDX: int32(m.read32(ccb + 0x18)), HDY: int32(m.read32(ccb + 0x1C)),
		VDX: int32(m.read32(ccb + 0x20)), VDY: int32(m.read32(ccb + 0x24)),
		HDDX: int32(m.read32(ccb + 0x28)), HDDY: int32(m.read32(ccb + 0x2C)),
		PIXC:   m.read32(ccb + 0x30),
		Bitmap: bm.buf, BitmapW: bm.w, BitmapH: bm.h,
	}
	// The preamble is either in the CCB or ahead of the source data — the same choice
	// drawOneCel makes, and the dimensions come from it and nowhere else.
	s := src
	if flags&ccbCCBPre != 0 {
		c.PRE0, c.PRE1 = m.read32(ccb+0x34), m.read32(ccb+0x38)
	} else {
		c.PRE0 = m.read32(s)
		s += 4
		if flags&ccbPacked == 0 {
			c.PRE1 = m.read32(s)
		}
	}
	c.Packed = flags&ccbPacked != 0
	c.BPP = bppFromPRE0[c.PRE0&7]
	c.Coded = c.BPP <= 8
	c.LRForm = c.PRE1&pre1LRForm != 0
	c.Width = int(c.PRE1&0x7FF) + 1
	c.Height = int((c.PRE0>>6)&0x3FF) + 1
	if c.Packed {
		c.Width = bm.w // packed lines are self-delimiting; the width is the clip bound
	}
	m.OnCel(c)
}

// celPixel reports one written pixel, and tallies it.
func (m *Machine) celPixel(bm gfxBitmap, x, y int, pix uint16) {
	m.celCnt.pixels++
	if m.OnPixel == nil || x < 0 || y < 0 {
		return
	}
	r, g, b := chan5(pix)
	// The cel engine's pixels are RGB555; expand to 8 bits the way the framebuffer
	// renderer does, so the colour the debugger shows is the colour on screen.
	m.OnPixel(uint32(x), uint32(y), PixelEvent{
		Drawn: true,
		R:     uint8(r << 3), G: uint8(g << 3), B: uint8(b << 3), A: 0xFF,
	})
}

// RunStopAfterCel runs the machine until k cels have been drawn, then stops — the
// command scrubber's halt. It counts cels from zero, so restoring a frame's start
// snapshot and calling this with k renders the frame exactly as it stood after cel k-1.
//
// The machine it leaves behind is mid-chain: the game asked for a chain of cels and got
// part of one. That is fine for what this is for — a scratch machine, restored from a
// snapshot before every replay and never resumed — and it is not fine for anything else.
func (m *Machine) RunStopAfterCel(k int, budget uint64) Result {
	m.celLimit, m.celCount = k, 0
	r := m.Run(budget)
	m.celLimit, m.celCount = 0, 0
	return r
}

// Cels reports how many cels have been drawn under the current limit.
func (m *Machine) Cels() int { return m.celCount }

// DrawTarget is the bitmap the game is drawing into: the back buffer, which is not the
// one on screen. DisplayBuffer is the front one. The gap between them is exactly where a
// frame goes missing, and a debugger that showed the front buffer mid-frame would show a
// picture that never changes.
//
// ok is false before the game has made a screen.
func (m *Machine) DrawTarget() (buf uint32, w, h int, ok bool) {
	// The draw target is the bitmap that is NOT being displayed. With one screen made,
	// that is simply the bitmap; with two (double buffering) it is the other one.
	for _, bm := range m.bitmaps {
		if bm.buf != 0 && bm.buf != m.displayBuf {
			return bm.buf, bm.w, bm.h, true
		}
	}
	for _, bm := range m.bitmaps {
		if bm.buf != 0 {
			return bm.buf, bm.w, bm.h, true
		}
	}
	return 0, 0, 0, false
}

// Bitmaps lists every bitmap the graphics folio has made — every buffer a frame could be
// drawn into. The debugger needs them all: provenance is gathered per bitmap, because a
// frame is drawn into more than one (Need for Speed renders its rear-view mirror into an
// off-screen buffer and then draws that buffer as a cel).
func (m *Machine) Bitmaps() []MemRegion {
	var out []MemRegion
	for _, bm := range m.bitmaps {
		if bm.buf != 0 {
			out = append(out, MemRegion{Name: fmt.Sprintf("bitmap %08X", bm.buf), Base: bm.buf, Size: uint32(bm.w * bm.h * 2)})
		}
	}
	return out
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
		{"dram", 0, dramSize},
		{"vram", vramBase, vramSize},
		{"madam", madamBase, madamEnd - madamBase},
		{"clio", clioBase, clioEnd - clioBase},
	}
}

// RenderBitmap decodes a bitmap out of VRAM as an image. The 3DO's framebuffer is not
// row-major: it is stored in interleaved line pairs (see pixelAddress), so a naive read
// of the same memory shows the picture combed into stripes.
func (m *Machine) RenderBitmap(buf uint32, w, h int) (*image.RGBA, error) {
	if w <= 0 || h <= 0 || w > 2048 || h > 2048 {
		return nil, fmt.Errorf("threedo: bitmap size %dx%d out of range", w, h)
	}
	return m.CaptureVRAM(buf, w, h), nil
}

// Frames is how many times the game has put a buffer on screen.
func (m *Machine) Frames() uint64 { return m.frame }

// Volume exposes the mounted disc, for a debugger that browses it. Nil if none.
func (m *Machine) Volume() *Volume { return m.vol }

// ClearBreakpoints/Stopped are not here: the 3DO machine has no breakpoint table, and
// the debugger enforces breakpoints from OnStep instead. See threedoadapter.
