package gbamachine

// The AGB bus. One address decoder for both byte and wide accesses, with the
// quirks that are load-bearing on this machine:
//
//   - EWRAM/IWRAM/palette/VRAM/OAM are all MIRRORED across their 16 MiB windows;
//     games (and the BIOS's own RegisterRamReset) rely on the top-of-IWRAM mirror
//     0x03FFFFxx.
//   - Byte WRITES to palette and BG VRAM do not store a byte: the hardware writes
//     the byte to BOTH halves of the addressed halfword. Byte writes to OBJ VRAM
//     and OAM are IGNORED entirely. (8-bit stores are how a naive memset breaks
//     sprites on real hardware, and an emulator that stores the byte anyway hides
//     exactly that class of game bug — and mis-renders games that exploit it.)
//   - The ROM appears three times (0x08/0x0A/0x0C) with different waitstates —
//     which this nominal-timing model treats identically — and the 0x0D window is
//     where a cart's serial EEPROM answers.
//   - Reads the model cannot honestly serve (BIOS image, unmapped space) are
//     logged and return zero rather than inventing plausible open-bus values.

type bus struct{ m *Machine }

func (b *bus) region(a uint32) (mem []byte, off uint32, ok bool) {
	m := b.m
	switch a >> 24 {
	case 0x02:
		return m.ewram, a & (ewramSize - 1), true
	case 0x03:
		return m.iwram, a & (iwramSize - 1), true
	case 0x05:
		return m.pal, a & (palSize - 1), true
	case 0x06:
		// 96 KiB mirrored in 128 KiB steps; within a step the last 32 KiB block
		// (0x18000..0x1FFFF) mirrors the OBJ half (0x10000..0x17FFF).
		off = a & 0x1FFFF
		if off >= vramSize {
			off -= 0x8000
		}
		return m.vram, off, true
	case 0x07:
		return m.oam, a & (oamSize - 1), true
	}
	return nil, 0, false
}

func (b *bus) romByte(a uint32) (byte, bool) {
	off := a & 0x01FFFFFF
	if int(off) < len(b.m.rom) {
		return b.m.rom[off], true
	}
	return 0, false
}

// Read services a byte load.
func (b *bus) Read(a uint32) byte {
	m := b.m
	if m.OnRead != nil {
		defer func() { m.OnRead(a, 0, m.cpu.R[15]) }()
	}
	switch a >> 24 {
	case 0x00, 0x01:
		if a < biosSize {
			// No BIOS image in this model: SWIs are HLE'd, and a direct read of the
			// BIOS region (checksums, anti-emulator probes) is a fact worth logging.
			m.note("read of BIOS region 0x%08X (no BIOS image; returns 0)", a)
			return 0
		}
	case 0x04:
		v := m.ioRead16(a &^ 1)
		return byte(v >> (8 * (a & 1)))
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C:
		if v, ok := b.romByte(a); ok {
			return v
		}
		// Past the chip: the AGB bus answers with the prefetch latch, which for a
		// linear read is the address's own low bits. Games use this as a cheap
		// "no chip here" marker; model that shape rather than zero.
		return byte((a >> 1) >> (8 * (a & 1)))
	case 0x0D:
		if m.eeprom.present {
			return byte(m.eeprom.read())
		}
		if v, ok := b.romByte(a); ok {
			return v
		}
		return byte((a >> 1) >> (8 * (a & 1)))
	case 0x0E, 0x0F:
		m.note("read of SRAM/Flash region 0x%08X (this cart is EEPROM; returns 0)", a)
		return 0
	}
	if mem, off, ok := b.region(a); ok {
		return mem[off]
	}
	m.note("read of unmapped address 0x%08X", a)
	return 0
}

// Write services a byte store, with the palette/VRAM/OAM byte-store quirks.
func (b *bus) Write(a uint32, v byte) {
	m := b.m
	if m.OnWrite != nil {
		m.OnWrite(a, v, m.cpu.R[15])
	}
	switch a >> 24 {
	case 0x00, 0x01:
		m.note("write to BIOS region 0x%08X (ignored)", a)
		return
	case 0x04:
		m.ioWrite8(a, v)
		return
	case 0x05: // byte store duplicated into the whole halfword
		off := a & (palSize - 1) &^ 1
		m.pal[off], m.pal[off+1] = v, v
		return
	case 0x06:
		off := a & 0x1FFFF
		if off >= vramSize {
			off -= 0x8000
		}
		if off >= m.ppu.objVRAMBase(m) {
			return // byte stores to OBJ tiles are ignored by the hardware
		}
		off &^= 1
		m.vram[off], m.vram[off+1] = v, v
		return
	case 0x07:
		return // byte stores to OAM are ignored
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C:
		m.note("byte write to ROM region 0x%08X (ignored)", a)
		return
	case 0x0D:
		if m.eeprom.present {
			m.eeprom.write(uint16(v))
			return
		}
		m.note("byte write to 0x0D region 0x%08X with no EEPROM (ignored)", a)
		return
	case 0x0E, 0x0F:
		m.note("write to SRAM/Flash region 0x%08X (this cart is EEPROM; ignored)", a)
		return
	}
	if mem, off, ok := b.region(a); ok {
		mem[off] = v
		return
	}
	m.note("write to unmapped address 0x%08X", a)
}

// --- wide accesses (BusWide): EEPROM is read-to-pop, so these are required ---

func (b *bus) Read16(a uint32) uint16 {
	m := b.m
	if m.OnRead != nil {
		m.OnRead(a, 0, m.cpu.R[15])
		m.OnRead(a+1, 0, m.cpu.R[15])
	}
	switch a >> 24 {
	case 0x04:
		return m.ioRead16(a)
	case 0x0D:
		if m.eeprom.present {
			return m.eeprom.read()
		}
	}
	if mem, off, ok := b.region(a); ok {
		return uint16(mem[off]) | uint16(mem[off+1])<<8
	}
	return uint16(b.Read(a)) | uint16(b.Read(a+1))<<8
}

func (b *bus) Write16(a uint32, v uint16) {
	m := b.m
	if m.OnWrite != nil {
		m.OnWrite(a, byte(v), m.cpu.R[15])
		m.OnWrite(a+1, byte(v>>8), m.cpu.R[15])
	}
	switch a >> 24 {
	case 0x04:
		m.ioWrite16(a, v)
		return
	case 0x0D:
		if m.eeprom.present {
			m.eeprom.write(v)
			return
		}
	}
	if mem, off, ok := b.region(a); ok {
		mem[off] = byte(v)
		mem[off+1] = byte(v >> 8)
		return
	}
	b.Write(a, byte(v))
	b.Write(a+1, byte(v>>8))
}

func (b *bus) Read32(a uint32) uint32 {
	return uint32(b.Read16(a)) | uint32(b.Read16(a+2))<<16
}

func (b *bus) Write32(a uint32, v uint32) {
	b.Write16(a, uint16(v))
	b.Write16(a+2, uint16(v>>16))
}

// r32/w32 are the shim's own accessors (the BIOS IRQ dispatch pushes registers).
func (b *bus) r32(a uint32) uint32    { return b.Read32(a) }
func (b *bus) w32(a, v uint32)        { b.Write32(a, v) }
func (b *bus) r16(a uint32) uint16    { return b.Read16(a) }
func (b *bus) w16(a uint32, v uint16) { b.Write16(a, v) }
