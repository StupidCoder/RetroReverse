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
	case a >= swramBase && a < swramEnd:
		return c.wramSlot(a)
	}
	if c.arm9 {
		if a >= c.itcmBase && a < c.itcmBase+uint32(len(c.itcm)) {
			return c.itcm, a - c.itcmBase
		}
	} else {
		if a < uint32(len(c.low)) { // ARM7 BIOS/low vectors
			return c.low, a
		}
		// 0x03800000 and up is ALWAYS the ARM7's own WRAM, mirrored every 64 KiB.
		if a >= wram7Base {
			return c.wram7, (a - wram7Base) & (wram7Size - 1)
		}
	}
	return nil, 0
}

// wramSlot resolves an address in 0x03000000..0x037FFFFF — the shared-WRAM window,
// whose contents depend on WRAMCNT.
//
// This is the DS's one genuinely treacherous piece of address decoding, and getting it
// wrong is invisible until it is fatal. The 32 KiB shared block is split between the two
// cores by WRAMCNT (two 16 KiB banks), and the ARM7's half of the window is NOT simply
// "the shared block" — when WRAMCNT hands the ARM7 nothing, which is how the machine
// boots, the ARM7's whole 0x03000000..0x037FFFFF window MIRRORS ITS OWN 64 KiB WRAM.
//
// That mirror is load-bearing. SM64DS's ARM7 crt0 relocates itself to 0x037F8000 and
// writes its BIOS IRQ handler pointer to 0x037FFFFC — which on hardware is 0x0380FFFC,
// the handler slot the BIOS reads. Route that window to the shared block instead and the
// pointer lands in a different array from the one the interrupt path reads: the ARM7
// never takes an interrupt, never runs its IRQ-driven PXI code, and the ARM9 waits for a
// boot message that cannot come. Both cores then spin for ever, and nothing about the
// symptom points here.
func (c *core) wramSlot(a uint32) ([]byte, uint32) {
	off := a - swramBase
	const bank = swramSize / 2 // 16 KiB
	mode := c.m.wramcnt & 3
	if c.arm9 {
		switch mode {
		case 0: // the whole 32 KiB block is the ARM9's
			return c.m.swram, off & (swramSize - 1)
		case 1: // second bank to the ARM9
			return c.m.swram, bank + off&(bank-1)
		case 2: // first bank to the ARM9
			return c.m.swram, off & (bank - 1)
		default: // 3: the ARM9 gets none — reads are undefined, writes are dropped
			return nil, 0
		}
	}
	switch mode {
	case 1: // first bank to the ARM7
		return c.m.swram, off & (bank - 1)
	case 2: // second bank to the ARM7
		return c.m.swram, bank + off&(bank-1)
	case 3: // the whole block to the ARM7
		return c.m.swram, off & (swramSize - 1)
	default: // 0: the ARM7 gets none, and the window mirrors its own WRAM
		return c.wram7, a & (wram7Size - 1)
	}
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
	if b.c.m.OnWrite != nil {
		b.c.m.OnWrite(b.c.arm9, a, v, b.c.cpu.R[15])
	}
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
