package dsmachine

// Direct memory access. Each core has four channels; each is a source, a
// destination, a word count and a control word that says when it should fire.
//
// DMA is not an optimisation on the DS, it is the transport: a game moves almost
// nothing with the CPU. Textures and tile data reach VRAM by DMA, the geometry
// engine is fed by DMA (start mode 7 tops the GXFIFO up whenever it drains below
// half), and the card's read FIFO is drained by DMA. A machine without DMA does
// not run a DS game slowly — it does not run it at all.

// DMA start modes. The two cores do not share this encoding: the ARM9 has a
// 3-bit mode field (bits 27..29) and the ARM7 a 2-bit one (bits 28..29), so the
// same control word means different things on the two sides.
const (
	dmaImmediate = iota
	dmaVBlank
	dmaHBlank
	dmaDisplaySync
	dmaMainMemDisplay
	dmaCard
	dmaGBACart
	dmaGXFIFO
)

// dmaChan is one channel's programmed state plus the internal address counters it
// keeps across a repeating transfer.
type dmaChan struct {
	src, dst uint32 // as programmed
	ctrl     uint32 // the control half-word (bits 16..31 of DMAxCNT)
	count    uint32 // word count as programmed (0 means "maximum")

	csrc, cdst uint32 // live counters (survive between repeats)
	crem       uint32 // words remaining in the current repeat
	active     bool   // enabled and latched
}

// dmaMode extracts a channel's start mode, which is encoded differently per core.
func (c *core) dmaMode(ch *dmaChan) int {
	if c.arm9 {
		return int(ch.ctrl>>11) & 7 // bits 27..29 of the 32-bit register
	}
	// The ARM7's field is two bits (28..29), and its meanings are not the ARM9's:
	// 0 immediate, 1 vblank, 2 card, 3 GBA cart / wireless.
	switch (ch.ctrl >> 12) & 3 {
	case 0:
		return dmaImmediate
	case 1:
		return dmaVBlank
	case 2:
		return dmaCard
	default:
		return dmaGBACart
	}
}

// maxCount is the largest word count a channel can be given (the count field is
// wider on the ARM9), and is also what a programmed count of zero means.
func (c *core) maxCount() uint32 {
	if c.arm9 {
		return 0x1FFFFF
	}
	return 0xFFFF
}

// writeDMACtrl latches a channel when the game sets its enable bit, and fires it
// immediately if that is its start mode.
func (c *core) writeDMACtrl(n int) {
	ch := &c.dma[n]
	if ch.ctrl&0x8000 == 0 { // bit 31: disabled
		ch.active = false
		return
	}
	if ch.active {
		return // already latched; a rewrite of the same enable is not a re-arm
	}
	ch.active = true
	// The address counters are latched from the programmed registers when the
	// channel is enabled, and then *kept* across repeats — which is the whole point
	// of repeat mode. Only a fresh enable reloads them.
	ch.csrc, ch.cdst = ch.src, ch.dst
	ch.crem = ch.count
	if ch.crem == 0 {
		ch.crem = c.maxCount()
	}
	if c.dmaMode(ch) == dmaImmediate {
		c.runDMAChan(n)
	}
}

// runDMA fires every channel on both cores whose start mode matches the event.
func (m *Machine) runDMA(mode int) {
	for _, c := range m.cores() {
		for n := range c.dma {
			if c.dma[n].active && c.dmaMode(&c.dma[n]) == mode {
				c.runDMAChan(n)
			}
		}
	}
}

// runDMAChan performs one channel's transfer to completion.
//
// This is a *blocking* model: on hardware the DMA steals bus cycles from the CPU
// and the transfer takes real time, but nothing a DS game does can observe the
// difference as long as the CPU is stopped for the duration — which it is, because
// we run the whole transfer inside the instruction that started it.
func (c *core) runDMAChan(n int) {
	ch := &c.dma[n]
	b := &bus{c: c}
	word := ch.ctrl&0x0400 != 0 // bit 26: 32-bit units
	dstMode := (ch.ctrl >> 5) & 3
	srcMode := (ch.ctrl >> 7) & 3
	unit := uint32(2)
	if word {
		unit = 4
	}

	// A GXFIFO-driven channel does not transfer its whole count in one go: it tops
	// the geometry FIFO up in 112-word bursts whenever it has drained below half.
	// Draining it all at once would be wrong in the one way that matters — the FIFO
	// would overflow, and the words that overflow are silently lost.
	todo := ch.crem
	if c.arm9 && c.dmaMode(ch) == dmaGXFIFO {
		if !c.m.gpu3d.fifoBelowHalf() {
			return
		}
		if todo > 112 {
			todo = 112
		}
	}

	for i := uint32(0); i < todo; i++ {
		if word {
			b.w32(ch.cdst, b.r32(ch.csrc))
		} else {
			v := uint32(b.Read(ch.csrc)) | uint32(b.Read(ch.csrc+1))<<8
			b.Write(ch.cdst, byte(v))
			b.Write(ch.cdst+1, byte(v>>8))
		}
		switch srcMode {
		case 0:
			ch.csrc += unit
		case 1:
			ch.csrc -= unit
		}
		switch dstMode {
		case 0, 3:
			ch.cdst += unit
		case 1:
			ch.cdst -= unit
		}
	}
	ch.crem -= todo
	if ch.crem > 0 {
		return // a GXFIFO burst that has more to give: it resumes when the FIFO drains
	}

	if ch.ctrl&0x4000 != 0 { // bit 30: interrupt on completion
		c.raise(irqDMA0 << uint(n))
	}
	repeat := ch.ctrl&0x0200 != 0 // bit 25
	if repeat && c.dmaMode(ch) != dmaImmediate {
		// Re-arm for the next event. Destination mode 3 ("increment/reload") reloads
		// the destination but leaves the source where it stopped — how a channel
		// streams a growing source into a fixed buffer.
		ch.crem = ch.count
		if ch.crem == 0 {
			ch.crem = c.maxCount()
		}
		if dstMode == 3 {
			ch.cdst = ch.dst
		}
		return
	}
	ch.active = false
	ch.ctrl &^= 0x8000 // clear the enable bit the game reads back
}

// gxfifoDrained is called by the 3D engine when its command FIFO falls below half
// full, which is the event a mode-7 channel waits on.
func (m *Machine) gxfifoDrained() {
	c := m.ARM9
	for n := range c.dma {
		if c.dma[n].active && c.dmaMode(&c.dma[n]) == dmaGXFIFO {
			c.runDMAChan(n)
		}
	}
}
