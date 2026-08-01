package gbamachine

// The four DMA channels. This model performs a triggered transfer instantly at
// its trigger boundary (immediate on enable, V-blank and H-blank at the raster
// events) — nominal timing, like everything else here. The sound-FIFO special
// modes are logged until the sound hardware exists; DMA3 is also the EEPROM's
// courier (the bus routes 0x0D-region accesses to the serial device, so an
// EEPROM transfer needs nothing special here).

type dmaChan struct {
	src, dst           uint32 // the register values as written
	latchSrc, latchDst uint32 // the internal pointers a transfer advances
	count              uint16
	ctrl               uint16
}

// dmaRegWrite services a store to the 0x0B0-0x0DE block.
func (m *Machine) dmaRegWrite(reg uint32, v uint16) {
	n := int(reg-0x0B0) / 12
	d := &m.dma[n]
	switch (reg - 0x0B0) % 12 {
	case 0:
		d.src = d.src&0xFFFF0000 | uint32(v)
	case 2:
		d.src = d.src&0xFFFF | uint32(v)<<16
	case 4:
		d.dst = d.dst&0xFFFF0000 | uint32(v)
	case 6:
		d.dst = d.dst&0xFFFF | uint32(v)<<16
	case 8:
		d.count = v
	case 10:
		was := d.ctrl
		d.ctrl = v
		if v&(1<<15) != 0 && was&(1<<15) == 0 {
			// Rising enable edge: latch source/dest into the channel's internal
			// registers and fire if the timing is "immediately".
			d.latchSrc, d.latchDst = d.src, d.dst
			if v>>12&3 == 0 {
				m.dmaRun(n)
			}
		}
	}
}

// dmaTrigger fires every enabled channel whose timing matches (1 = V-blank,
// 2 = H-blank).
func (m *Machine) dmaTrigger(timing uint16) {
	for n := range m.dma {
		d := &m.dma[n]
		if d.ctrl&(1<<15) != 0 && d.ctrl>>12&3 == timing {
			m.dmaRun(n)
		}
	}
}

// dmaRun performs one whole transfer of channel n.
func (m *Machine) dmaRun(n int) {
	d := &m.dma[n]
	b := &bus{m: m}
	count := uint32(d.count)
	max := uint32(0x4000)
	if n == 3 {
		max = 0x10000
	}
	if count == 0 {
		count = max
	}
	word := d.ctrl&(1<<10) != 0
	unit := uint32(2)
	if word {
		unit = 4
	}
	step := func(mode uint16) int32 {
		switch mode {
		case 1:
			return -int32(unit)
		case 2:
			return 0
		default:
			return int32(unit)
		}
	}
	sstep := step(d.ctrl >> 7 & 3)
	dmode := d.ctrl >> 5 & 3
	dstep := step(dmode)

	src, dst := d.latchSrc, d.latchDst
	startDst := dst
	for i := uint32(0); i < count; i++ {
		if word {
			b.Write32(dst&^3, b.Read32(src&^3))
		} else {
			b.Write16(dst&^1, b.Read16(src&^1))
		}
		src = uint32(int32(src) + sstep)
		dst = uint32(int32(dst) + dstep)
	}
	d.latchSrc = src
	if dmode == 3 {
		// Increment+reload: the destination pointer resets every transfer (the
		// FIFO/frame-buffer refill mode).
		d.latchDst = d.dst
	} else {
		d.latchDst = dst
	}

	// A transfer INTO the EEPROM window is what frames a request: the device
	// counts the bits it was handed and only now decides what it was asked
	// (eeprom.go explains why bit-by-bit parsing cannot work).
	if m.eeprom.present && startDst>>24 == 0x0D {
		m.eeprom.endFrame(m)
	}

	if d.ctrl&(1<<14) != 0 {
		m.raise(uint16(irqDMA0) << uint(n))
	}
	if d.ctrl&(1<<9) == 0 || d.ctrl>>12&3 == 0 {
		d.ctrl &^= 1 << 15 // no repeat (or immediate): the channel disarms
	}
}
