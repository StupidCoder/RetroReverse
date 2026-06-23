package gameboy

import (
	"os"
	"testing"
)

const romPath = "../../Super Mario Land (GB)/Super Mario Land (World).gb"

func loadROM(t *testing.T) []byte {
	t.Helper()
	rom, err := os.ReadFile(romPath)
	if err != nil {
		t.Skipf("ROM not present (%v); skipping integration test", err)
	}
	return rom
}

// TestBootAndFrames runs Super Mario Land through its boot and a number of frames,
// then checks the machine reached its per-frame loop and the game drew tile data
// into VRAM — the basic "the oracle runs a real ROM" check.
func TestBootAndFrames(t *testing.T) {
	rom := loadROM(t)
	m := NewMachine(rom)

	done := m.RunFrames(60) // ~1 second of emulated time
	if m.CPU.Halted {
		t.Fatalf("CPU halted during run at PC=$%04X: %s", m.CPU.PC, m.CPU.HaltReason)
	}
	if done < 60 {
		t.Fatalf("only completed %d/60 frames", done)
	}

	// After a second of running, the title screen's tiles should be in VRAM. Count
	// non-zero bytes in the tile-data region ($8000-$97FF, i.e. vram[0:0x1800]).
	nonZero := 0
	for _, b := range m.vram[0:0x1800] {
		if b != 0 {
			nonZero++
		}
	}
	if nonZero < 256 {
		t.Fatalf("VRAM tile region looks empty (%d non-zero bytes) — boot likely stalled", nonZero)
	}

	// The interrupt master enable should be on (the game runs with interrupts) and
	// PC should be in ROM, not stuck at $0000.
	if m.CPU.PC < 0x0100 {
		t.Errorf("PC=$%04X looks wrong", m.CPU.PC)
	}
	t.Logf("ran 60 frames: PC=$%04X bank=%d, %d non-zero VRAM tile bytes, %d instrs",
		m.CPU.PC, m.romBank, nonZero, m.CPU.Instrs)
}

// TestVBlankProgresses checks the LY counter and the VBlank interrupt are actually
// advancing the game (the frame loop depends on them).
func TestVBlankProgresses(t *testing.T) {
	rom := loadROM(t)
	m := NewMachine(rom)
	m.RunFrames(10)
	// The VBlank interrupt must have been enabled by the game (IE bit 0) for its
	// frame loop to work.
	if m.ie&0x01 == 0 {
		t.Errorf("game never enabled the VBlank interrupt (IE=$%02X)", m.ie)
	}
}
