package dc

// boot.go is the direct boot: what the BIOS and IP.BIN bootstrap would have
// left behind, reduced to the facts a game may rely on. The boot binary named
// by IP.BIN loads at 8C010000 and entry is its first instruction; the syscall
// vectors hold pointers to trap addresses bios.go services; and every other
// byte of the BIOS work area (8C000000-8C00FFFF) is poisoned with a
// recognisable pattern, so a read of a value this file did not plant shows up
// as F00Dxxxx in a register instead of masquerading as a fact. The HLE-
// fictions lesson: a value our own stub invents comes back as an
// "observation" — poison, don't invent.

import "fmt"

// The trap window: PC values the vectors point at, chosen in the boot-ROM
// area so a missed trap fetches from unmapped ROM and halts loudly rather
// than executing poison.
const (
	trapSysinfo  = 0xA0000100
	trapRomfont  = 0xA0000104
	trapFlashrom = 0xA0000108
	trapGdrom    = 0xA000010C
	trapMenu     = 0xA0000110
)

// Boot loads the disc's boot binary and arranges the machine the way the
// BIOS hands over to a game.
func (m *Machine) Boot() error {
	if m.Disc == nil {
		return fmt.Errorf("boot: no disc mounted")
	}
	name := m.Disc.BootFilePath()
	bin, err := m.Disc.Vol.ReadFile(name)
	if err != nil {
		return fmt.Errorf("boot: reading %s: %w", name, err)
	}
	const load = 0x00010000 // 8C010000 in RAM
	if load+len(bin) > len(m.RAM) {
		return fmt.Errorf("boot: %s (%d bytes) does not fit at 8C010000", name, len(bin))
	}

	// Poison the BIOS work area word by word: F00D + the word's own offset.
	for off := 0; off < 0x10000; off += 4 {
		v := 0xF00D0000 | uint32(off)
		m.RAM[off], m.RAM[off+1], m.RAM[off+2], m.RAM[off+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
	}

	// The syscall vectors: each holds a POINTER the game reads and jumps
	// through. These five are the documented interface; nothing else is
	// planted.
	m.putRAM32(0x000000B0, trapSysinfo)
	m.putRAM32(0x000000B4, trapRomfont)
	m.putRAM32(0x000000B8, trapFlashrom)
	m.putRAM32(0x000000BC, trapGdrom)
	m.putRAM32(0x000000E0, trapMenu)

	// 8C000010: the BIOS's uncached return-from-exception stub — rte; nop as
	// real instructions. The evidence is Crazy Taxi's own interrupt epilogue:
	// it restores the whole context, SSR and SPC included, hand-builds the
	// constant AC000010 (the P2 mirror, so the rte runs uncached) and jumps
	// there with nothing left to do but return. Planted as code, not a trap,
	// because it IS code on hardware.
	m.RAM[0x10], m.RAM[0x11] = 0x2B, 0x00 // rte
	m.RAM[0x12], m.RAM[0x13] = 0x09, 0x00 // nop

	// The console's unique ID lives at 8C000068 (the sysinfo ID syscall
	// returns this address); ours is arbitrary but planted, not poison.
	copy(m.RAM[0x68:0x70], []byte{0x52, 0x45, 0x54, 0x52, 0x4F, 0x52, 0x56, 0x00})

	// 8C00FFF0/F4: Crazy Taxi keeps a lazily-initialised work-structure
	// pointer pair here (its own literal pool names both words), and its
	// "already initialised?" test is tst — a fresh boot must read zero. The
	// poison found this contract; these two words are the evidence-driven
	// exception to it.
	m.putRAM32(0xFFF0, 0)
	m.putRAM32(0xFFF4, 0)

	copy(m.RAM[load:], bin)

	m.bios = newBIOS(m)

	c := m.CPU
	c.Reset()
	c.SetSR(0x400000F0) // privileged, bank 0, interrupts masked; the game unmasks
	c.SetFPSCR(0x00040001)
	c.R[15] = 0x8C00F400 // the stack the BIOS hands over
	c.SetPC(0x8C010000)
	return nil
}

func (m *Machine) putRAM32(off int, v uint32) {
	m.RAM[off], m.RAM[off+1], m.RAM[off+2], m.RAM[off+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
}

func (m *Machine) ram32(addr uint32) uint32 {
	off := addr & (RAMSize - 1)
	return uint32(m.RAM[off]) | uint32(m.RAM[off+1])<<8 | uint32(m.RAM[off+2])<<16 | uint32(m.RAM[off+3])<<24
}
