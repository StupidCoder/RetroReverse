package dsmachine

// Video timing. The DS's whole machine is paced by the display: the two CPUs, the
// DMA channels, the timers and the 3D engine all take their cadence from the dot
// clock, and every "wait" a DS game performs ultimately waits on a scanline.
//
// The numbers are the hardware's, not a guess:
//
//	a scanline is 355 dots  = 2130 system cycles (6 cycles per dot)
//	a frame    is 263 lines = 560,190 cycles -> 33.513982 MHz / 560190 = 59.83 Hz
//	lines 0..191 are visible; 192..262 are the vertical blank
//	dots 0..255 are visible; the horizontal blank begins at dot 256
//
// The ARM7 runs at the system clock (33.51 MHz) and the ARM9 at twice it, so a
// scanline is 2130 ARM7 cycles and 4260 ARM9 cycles. Our CPU cores count
// instructions, not cycles, so those become the per-line *instruction* budgets:
// one instruction per cycle is the right first approximation for code running out
// of TCM/cache, and DS software is written against the video clock rather than
// against an instruction count, so it tolerates the difference.
const (
	dotsPerLine   = 355
	linesPerFrame = 263
	visibleLines  = 192
	visibleDots   = 256

	cyclesPerLine9 = 4260 // ARM9 instruction budget for one scanline
	cyclesPerLine7 = 2130 // ARM7 instruction budget for one scanline
)

// DISPSTAT bits (per core — each CPU has its own copy of the register).
const (
	dispstatVBlankFlag = 1 << 0
	dispstatHBlankFlag = 1 << 1
	dispstatVMatchFlag = 1 << 2
	dispstatVBlankIRQ  = 1 << 3
	dispstatHBlankIRQ  = 1 << 4
	dispstatVMatchIRQ  = 1 << 5
)

// interrupt-source bits in IE/IF.
const (
	irqVBlank  = 1 << 0
	irqHBlank  = 1 << 1
	irqVMatch  = 1 << 2
	irqTimer0  = 1 << 3
	irqDMA0    = 1 << 8
	irqCard    = 1 << 19 // card transfer complete
	irqIPCSync = 1 << 16
	irqIPCSend = 1 << 17
	irqIPCRecv = 1 << 18
	irqGXFIFO  = 1 << 21
	irqSPI     = 1 << 23
)

// video is the display controller's timing state, shared by both cores (there is
// one display, and both CPUs observe it through their own DISPSTAT/VCOUNT).
type video struct {
	line   int  // 0..262, the current scanline (VCOUNT)
	hblank bool // set once this line's visible dots have been emitted
	frames uint64
}

// vcountMatch returns the scanline a core's DISPSTAT asks to be interrupted on. It is
// a 9-bit value SPLIT ACROSS the register: bits 15..8 hold its low eight bits, and
// bit 7 — nowhere near them — holds its ninth.
//
// Reassembling that ninth bit as `(d >> 7) & 0x100` looks right and is not: it lands
// on bit 15 of DISPSTAT, which is the TOP of the low byte, so any LYC of 128 or more
// silently acquires an extra 256. The ARM7 asks to be woken at line 197 to sample the
// touchscreen; it gets 453, a line the display never reaches, and the interrupt never
// fires. The screen then works perfectly and does not respond to the stylus, and
// nothing about that points at a scanline comparison.
func (c *core) vcountMatch() int {
	d := c.io[regDISPSTAT]
	return int((d>>8)&0xFF | (d&0x80)<<1)
}

// dispstat assembles a core's DISPSTAT: its own IRQ-enable bits (which the game
// wrote) over the display's live flags (which the hardware owns).
func (c *core) dispstat() uint32 {
	v := c.io[regDISPSTAT] & 0xFFB8 // keep the enables and the LYC field
	ln := c.m.vid.line
	// The VBlank flag is set for lines 192..261 — NOT for line 262, the last line
	// of the frame. Software that waits for the flag to *fall* (rather than for the
	// VBlank interrupt) relies on that one-line gap, and a model that leaves the
	// flag set through the wrap gives it no falling edge to see.
	if ln >= visibleLines && ln < linesPerFrame-1 {
		v |= dispstatVBlankFlag
	}
	if c.m.vid.hblank {
		v |= dispstatHBlankFlag
	}
	if ln == c.vcountMatch() {
		v |= dispstatVMatchFlag
	}
	return v
}

// startLine advances the display to the next scanline and raises whatever that
// entails on both cores: the VBlank and VCount-match interrupts, and the DMA
// channels those events start.
func (m *Machine) startLine() {
	v := &m.vid
	v.line++
	if v.line >= linesPerFrame {
		v.line = 0
		v.frames++
	}
	v.hblank = false

	for _, c := range m.cores() {
		if v.line == c.vcountMatch() && c.io[regDISPSTAT]&dispstatVMatchIRQ != 0 {
			c.raise(irqVMatch)
		}
	}
	switch v.line {
	case visibleLines: // entering the vertical blank
		for _, c := range m.cores() {
			if c.io[regDISPSTAT]&dispstatVBlankIRQ != 0 {
				c.raise(irqVBlank)
			}
		}
		m.runDMA(dmaVBlank)
		m.gpu3d.vblank(m) // the 3D engine's buffer swap is consumed here
		m.onFrame()
		m.prof.reset(m) // a new frame's accounting starts here
	case 0: // a new frame's first line: the display-capture/scroll latches reload
		m.gpu2d.beginFrame(m)
	}
}

// hblankNow marks the horizontal blank of the current line: it sets the flag,
// raises the interrupt on whichever cores enabled it, and starts the HBlank DMA
// channels (which only run on visible lines — there is no display fetch to feed
// during the vertical blank).
func (m *Machine) hblankNow() {
	m.vid.hblank = true
	for _, c := range m.cores() {
		if c.io[regDISPSTAT]&dispstatHBlankIRQ != 0 {
			c.raise(irqHBlank)
		}
	}
	if m.vid.line < visibleLines {
		m.runDMA(dmaHBlank)
	}
}

func (m *Machine) cores() [2]*core { return [2]*core{m.ARM9, m.ARM7} }

// raise posts an interrupt source into a core's IF. Whether it is *delivered*
// depends on IE/IME and on whether the core is parked (see deliver).
func (c *core) raise(src uint32) { c.if_ |= src }
