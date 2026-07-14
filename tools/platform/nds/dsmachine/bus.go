package dsmachine

// bus is one core's view of the DS address space. The two cores do not see the
// same machine: the ARM9 has tightly-coupled memories, the graphics engines and all
// of VRAM; the ARM7 has its own WRAM, its own BIOS vectors, and a narrow window
// onto whichever VRAM banks have been handed to it. It implements arm.Bus.
type bus struct{ c *core }

// slot returns the backing byte slice and index for an address that lives in plain
// RAM, or (nil, 0) when the address needs special handling (I/O, VRAM) or is
// unmapped.
func (b *bus) slot(a uint32) ([]byte, uint32) {
	c := b.c
	switch {
	case a >= mainBase && a < mainMirrorEnd:
		// The ARM9's DTCM window shadows main-RAM addresses, for the ARM9 only.
		if c.dtcm != nil && a >= c.dtcmBase && a < c.dtcmBase+uint32(len(c.dtcm)) {
			return c.dtcm, a - c.dtcmBase
		}
		// Main RAM is 4 MiB and it MIRRORS every 4 MiB across the whole 0x02xxxxxx
		// region. That is not a curiosity — DS software lives in the top mirror. The
		// boot parameter block, the cartridge header copy and the ARM9/ARM7 handshake
		// words are all addressed as 0x027FFxxx, an alias of 0x023FFxxx, and NitroSDK
		// hard-codes them that way. Map only the first 4 MiB and every one of those
		// reads returns zero: the ARM7 waits for a handshake value that is being
		// written to an address it cannot see, and both cores spin for ever.
		return c.m.ram, (a - mainBase) & (mainSize - 1)
	case a >= swramBase && a < swramEnd:
		return c.wramSlot(a)
	}
	if c.arm9 {
		if a >= c.itcmBase && a < c.itcmBase+uint32(len(c.itcm)) {
			return c.itcm, a - c.itcmBase
		}
		switch a >> 24 {
		case 0x05: // palette RAM, mirrored every 2 KiB
			return c.m.pal, (a - palBase) & (palSize - 1)
		case 0x07: // OAM, mirrored every 2 KiB
			return c.m.oam, (a - oamBase) & (oamSize - 1)
		}
	} else {
		if a < uint32(len(c.low)) { // ARM7 BIOS/low vectors
			return c.low, a
		}
		// 0x03800000 and up is ALWAYS the ARM7's own WRAM, mirrored every 64 KiB.
		if a >= wram7Base && a < 0x04000000 {
			return c.wram7, (a - wram7Base) & (wram7Size - 1)
		}
	}
	return nil, 0
}

// wramSlot resolves an address in 0x03000000..0x037FFFFF — the shared-WRAM window,
// whose contents depend on WRAMCNT.
//
// This is the DS's one genuinely treacherous piece of address decoding, and getting
// it wrong is invisible until it is fatal. The 32 KiB shared block is split between
// the two cores by WRAMCNT (two 16 KiB banks), and the ARM7's half of the window is
// NOT simply "the shared block" — when WRAMCNT hands the ARM7 nothing, which is how
// the machine boots, the ARM7's whole 0x03000000..0x037FFFFF window MIRRORS ITS OWN
// 64 KiB WRAM.
//
// That mirror is load-bearing. SM64DS's ARM7 crt0 relocates itself to 0x037F8000 and
// writes its BIOS IRQ handler pointer to 0x037FFFFC — which on hardware is
// 0x0380FFFC, the handler slot the BIOS reads. Route that window to the shared block
// instead and the pointer lands in a different array from the one the interrupt path
// reads: the ARM7 never takes an interrupt, never runs its IRQ-driven PXI code, and
// the ARM9 waits for a boot message that cannot come. Both cores then spin for ever,
// and nothing about the symptom points here.
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
	c := b.c
	switch a >> 24 {
	case 0x04:
		return byte(c.ioRead(a&^3) >> (8 * (a & 3)))
	case 0x06:
		return c.vramRead(a)
	}
	if s, i := b.slot(a); s != nil {
		return s[i]
	}
	return 0
}

func (b *bus) Write(a uint32, v byte) {
	c := b.c
	if c.m.OnWrite != nil {
		c.m.OnWrite(c.arm9, a, v, c.cpu.R[15])
	}
	switch a >> 24 {
	case 0x04:
		c.ioWrite(a, v)
		return
	case 0x06:
		c.vramWrite(a, v)
		return
	}
	if s, i := b.slot(a); s != nil {
		s[i] = v
	}
}

// vramRead and vramWrite resolve the 0x06000000 window. The LCDC view (0x06800000+)
// reaches a bank *directly*, whatever it is currently mapped as — that is how a game
// uploads a texture and only then points the bank at the texture space. Below it,
// the engine views reach whichever banks are mapped there right now.
func (c *core) vramRead(a uint32) byte {
	v := c.m.vram
	if c.arm9 {
		if data, off, ok := v.lcdcSlot(a); ok {
			return data[off]
		}
	}
	if space, off, ok := v.cpuAccess(c.arm9, a); ok {
		return v.read8(space, off)
	}
	return 0
}

func (c *core) vramWrite(a uint32, b byte) {
	v := c.m.vram
	if c.arm9 {
		if data, off, ok := v.lcdcSlot(a); ok {
			data[off] = b
			return
		}
	}
	if space, off, ok := v.cpuAccess(c.arm9, a); ok {
		v.write8(space, off, b)
	}
}

// The wide access path (arm.BusWide). This is not an optimisation — it is a
// correctness requirement, and the DS is the machine that proves it.
//
// The IPC receive FIFO and the cartridge data port are *read to pop*: reading them
// consumes a word. Let the CPU core compose a 32-bit load out of four byte reads,
// as it does for a bus that offers nothing better, and each of those four reads pops
// the FIFO again — so the ARM7 swallows four IPC messages to receive one, and every
// word the game reads off the cartridge is glued together from the bytes of four
// different words. The game then boots into a machine that cannot load a file, and
// nothing about the symptom points at the bus.
//
// So the wide reads go through a single ioRead. The wide *writes* still walk the
// four bytes, because that is what the register file wants: I/O side effects are
// keyed to the byte that completes a register (io.go), and walking the lanes is what
// makes a 32-bit store to a 16-bit register pair do the right thing on both halves.

func (b *bus) Read16(a uint32) uint16 {
	if a>>24 == 0x04 {
		return uint16(b.c.ioRead(a&^3) >> (8 * (a & 2)))
	}
	return uint16(b.Read(a)) | uint16(b.Read(a+1))<<8
}

func (b *bus) Read32(a uint32) uint32 {
	if a>>24 == 0x04 {
		return b.c.ioRead(a) // one read, one pop
	}
	return uint32(b.Read(a)) | uint32(b.Read(a+1))<<8 | uint32(b.Read(a+2))<<16 | uint32(b.Read(a+3))<<24
}

func (b *bus) Write16(a uint32, v uint16) {
	b.Write(a, byte(v))
	b.Write(a+1, byte(v>>8))
}

func (b *bus) Write32(a, v uint32) {
	for i := uint32(0); i < 4; i++ {
		b.Write(a+i, byte(v>>(8*i)))
	}
}

// helpers for word access from inside the machine (DMA, the BIOS calls, the IRQ
// path) — they take the same wide path the CPU does.
func (b *bus) r32(a uint32) uint32 { return b.Read32(a) }
func (b *bus) w32(a, v uint32)     { b.Write32(a, v) }
