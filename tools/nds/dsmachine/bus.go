package dsmachine

// bus is one core's view of the DS address space: its private TCM/WRAM/BIOS, the
// shared main RAM and WRAM, and the memory-mapped I/O block (IPC, interrupts, and
// the stubs). It implements arm.Bus.
type bus struct{ c *core }

// slot returns the backing byte slice and index for a private/shared RAM address,
// or (nil, 0) when the address is I/O or unmapped.
func (b *bus) slot(a uint32) ([]byte, uint32) {
	c := b.c
	switch {
	case a >= mainBase && a < mainEnd:
		// The ARM9's DTCM window shadows main-RAM addresses for the ARM9 only.
		if c.dtcm != nil && a >= c.dtcmBase && a < c.dtcmBase+uint32(len(c.dtcm)) {
			return c.dtcm, a - c.dtcmBase
		}
		return c.m.ram, a - mainBase
	case a >= swramBase && a < swramEnd: // shared WRAM, mirrored every 0x8000
		return c.m.swram, (a - swramBase) & (swramSize - 1)
	}
	if c.arm9 {
		if a >= c.itcmBase && a < c.itcmBase+uint32(len(c.itcm)) {
			return c.itcm, a - c.itcmBase
		}
	} else {
		if a < uint32(len(c.low)) { // ARM7 BIOS/low vectors
			return c.low, a
		}
		if a >= wram7Base && a < wram7End {
			return c.wram7, a - wram7Base
		}
	}
	return nil, 0
}

func (b *bus) Read(a uint32) byte {
	if a>>24 == 0x04 {
		return byte(b.c.ioRead(a&^3) >> (8 * (a & 3)))
	}
	if s, i := b.slot(a); s != nil {
		return s[i]
	}
	return 0
}

func (b *bus) Write(a uint32, v byte) {
	if a>>24 == 0x04 {
		b.c.ioWrite(a, v)
		return
	}
	if s, i := b.slot(a); s != nil {
		s[i] = v
	}
}

// helpers for word access through a core's bus
func (b *bus) r32(a uint32) uint32 {
	return uint32(b.Read(a)) | uint32(b.Read(a+1))<<8 | uint32(b.Read(a+2))<<16 | uint32(b.Read(a+3))<<24
}
func (b *bus) w32(a, v uint32) {
	for i := uint32(0); i < 4; i++ {
		b.Write(a+i, byte(v>>(8*i)))
	}
}
