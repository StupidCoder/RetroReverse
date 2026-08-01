package gbamachine

// The memory-mapped I/O block at 0x04000000. GBA registers are 16-bit; the bus
// layer splits/joins byte and word accesses. m.io holds the last value written to
// every register (the diagnostic file); registers with live behaviour also update
// the machine's typed state. Unmodelled registers are logged, and their reads
// return 0 — not the last write — so a stub can never impersonate ready hardware.

// ioRead16 services a 16-bit register load.
func (m *Machine) ioRead16(a uint32) uint16 {
	reg := a & 0x00FFFFFF &^ 1
	var v uint16
	switch reg {
	case 0x000: // DISPCNT
		v = m.io[reg]
	case 0x004: // DISPSTAT: live flags in bits 0-2, the game's settings above
		v = m.io[reg] & 0xFFF8
		if m.vid.line >= 160 && m.vid.line <= 226 {
			v |= 1
		}
		if m.vid.hblank {
			v |= 2
		}
		if m.vid.line == int(m.io[0x004]>>8) {
			v |= 4
		}
	case 0x006: // VCOUNT
		v = uint16(m.vid.line)
	case 0x088: // SOUNDBIAS: storage is honest enough for a bias
		v = m.io[reg]
	case 0x0B8, 0x0C4, 0x0D0, 0x0DC: // DMAxCNT_L reads back 0
		v = 0
	case 0x0BA, 0x0C6, 0x0D2, 0x0DE: // DMAxCNT_H
		v = m.dma[(reg-0x0BA)/12].ctrl
	case 0x100, 0x104, 0x108, 0x10C: // TMxCNT_L: the live counter
		v = m.timers[(reg-0x100)/4].counter
	case 0x102, 0x106, 0x10A, 0x10E: // TMxCNT_H
		v = m.timers[(reg-0x102)/4].ctrl
	case 0x130: // KEYINPUT: active-low
		v = ^m.keys & 0x3FF
	case 0x132: // KEYCNT
		v = m.io[reg]
	case 0x200:
		v = m.ie
	case 0x202:
		v = m.if_
	case 0x204: // WAITCNT: storage (timing is nominal in this model)
		v = m.io[reg]
	case 0x208:
		if m.ime {
			v = 1
		}
	case 0x300: // POSTFLG
		v = 1
	default:
		switch {
		case reg <= 0x054: // the PPU register file: stored values are the truth
			v = m.io[reg]
		case reg >= 0x060 && reg <= 0x0A6: // sound: stored, and noted once
			m.note("sound register 0x%03X read (sound is not yet modelled)", reg)
			v = m.io[reg]
		case reg >= 0x120 && reg <= 0x15A: // serial: report an idle link
			m.note("serial register 0x%03X read (link port not modelled; reads idle)", reg)
			v = 0
		default:
			m.note("unmodelled I/O register 0x%03X read", reg)
			v = 0
		}
	}
	if m.OnIO != nil {
		m.OnIO(false, reg, v, m.cpu.R[15])
	}
	return v
}

// ioWrite8 stores one byte of a register, preserving its other half.
func (m *Machine) ioWrite8(a uint32, v byte) {
	reg := a &^ 1
	old := m.io[reg&0x00FFFFFF]
	if a&1 == 0 {
		m.ioWrite16(reg, old&0xFF00|uint16(v))
	} else {
		m.ioWrite16(reg, old&0x00FF|uint16(v)<<8)
	}
}

// ioWrite16 services a 16-bit register store.
func (m *Machine) ioWrite16(a uint32, v uint16) {
	reg := a & 0x00FFFFFF &^ 1
	if m.OnIO != nil {
		m.OnIO(true, reg, v, m.cpu.R[15])
	}
	m.io[reg] = v
	switch {
	case reg == 0x028 || reg == 0x02A || reg == 0x02C || reg == 0x02E ||
		reg == 0x038 || reg == 0x03A || reg == 0x03C || reg == 0x03E:
		// BG2X/BG2Y/BG3X/BG3Y (32-bit reference points): a write reloads the
		// affine engine's internal accumulator immediately.
		m.ppu.reloadAffineRef(m, reg)
	case reg == 0x202: // IF: write-1-to-CLEAR (acknowledging an interrupt)
		m.if_ &^= v
	case reg == 0x200:
		m.ie = v
	case reg == 0x208:
		m.ime = v&1 != 0
	case reg >= 0x0B0 && reg <= 0x0DE:
		m.dmaRegWrite(reg, v)
	case reg >= 0x100 && reg <= 0x10E:
		m.timerRegWrite(reg, v)
	case reg == 0x301: // HALTCNT: bit 7 selects Stop; either way, park until an IRQ
		m.waiting, m.waitAny, m.waitMask = true, true, 0
	case reg >= 0x060 && reg <= 0x0A6:
		m.note("sound register 0x%03X written (sound is not yet modelled)", reg)
	case reg > 0x054 && reg < 0x060, reg >= 0x0A8 && reg < 0x0B0,
		reg >= 0x110 && reg < 0x120, reg >= 0x134 && reg < 0x200 && reg != 0x132:
		m.note("unmodelled I/O register 0x%03X written (value 0x%04X)", reg, v)
	}
	// Registers 0x000-0x054 (display), 0x132 (KEYCNT), 0x204 (WAITCNT) live in
	// m.io and are consumed where they act (the PPU reads its file per scanline;
	// KEYCNT is checked when keys change).
}
