package n64

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

const testROM = "../../../games/pilotwings-64-n64/image/Pilotwings 64 (USA).z64"

func loadTestROM(t *testing.T) *ROM {
	t.Helper()
	if _, err := os.Stat(testROM); os.IsNotExist(err) {
		t.Skip("Pilotwings 64 image not present (game images are not committed)")
	}
	r, err := Load(testROM)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestBootReachesGameEntry is the Phase 3a gate. A cold boot runs IPL3 out of
// SP DMEM: it initialises RDRAM, sizes it, relocates a loader stub into RDRAM,
// DMAs a megabyte of the cartridge to the address in the header, checksums it
// against a value derived from the CIC seed in $s6, and jumps to the entry point.
//
// Reaching the entry point therefore proves a great deal at once — the 64-bit
// integer core across a 1 MiB checksum sweep, the PI DMA engine, the segment
// map, the branch-likely annulment the stub's polling loops use, and the CIC
// seed. All of it driven by code on the cartridge; only the PIF handoff is ours.
//
// It is not a conformance suite: nothing here exercises the FPU, TLB refill, or
// most of the annulling branches. That is what n64-systemtest is for.
func TestBootReachesGameEntry(t *testing.T) {
	rom := loadTestROM(t)
	m := NewMachine(rom)
	if err := m.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	m.SetBreakpoint(rom.Header.Entry)

	res := m.Run(50_000_000)
	if res.PC != rom.Header.Entry {
		t.Fatalf("IPL3 did not reach the game entry: %s", res)
	}
	if got, want := m.OSMemSize(), uint32(rdramSize); got != want {
		t.Errorf("osMemSize = 0x%08X, want 0x%08X — IPL3's RDRAM sizing disagrees with the model", got, want)
	}
	// A boot that touches an unmapped address has found a gap in the memory map.
	for _, l := range m.Log {
		t.Errorf("machine note during boot: %s", l)
	}
	t.Logf("reached the game entry 0x%08X after %d instructions", res.PC, res.Steps)
}

// TestSaveStateRoundTrip is what keeps the snapshot honest as devices are added.
// A device whose registers were never added to snapshot diverges here, loudly,
// rather than as a confusing render bug three phases later.
//
// It relies on the oracle pacing everything by instruction count rather than
// wall-clock cycles: restoring a snapshot and running N more steps must land on
// exactly the state an uninterrupted run reaches.
func TestSaveStateRoundTrip(t *testing.T) {
	rom := loadTestROM(t)
	const settle = 250_000

	// Reference: boot to the entry point, then run on for a while.
	ref := NewMachine(rom)
	if err := ref.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	ref.SetBreakpoint(rom.Header.Entry)
	if res := ref.Run(50_000_000); res.PC != rom.Header.Entry {
		t.Fatalf("reference boot did not reach the entry: %s", res)
	}

	path := filepath.Join(t.TempDir(), "entry.state")
	if err := ref.SaveState(path); err != nil {
		t.Fatal(err)
	}
	ref.Run(settle)

	// Restore into a fresh machine and run the same distance.
	restored := NewMachine(rom)
	if err := restored.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	if err := restored.LoadState(path); err != nil {
		t.Fatal(err)
	}
	restored.Run(settle)

	if ref.CPU.PC != restored.CPU.PC {
		t.Errorf("PC diverged: reference %016X, restored %016X", ref.CPU.PC, restored.CPU.PC)
	}
	if ref.CPU.Steps != restored.CPU.Steps {
		t.Errorf("Steps diverged: reference %d, restored %d", ref.CPU.Steps, restored.CPU.Steps)
	}
	if ref.CPU.R != restored.CPU.R {
		t.Error("general registers diverged after a restore")
	}
	if ref.CPU.COP0 != restored.CPU.COP0 {
		t.Error("COP0 registers diverged after a restore")
	}
	if got, want := sha256.Sum256(restored.RDRAM), sha256.Sum256(ref.RDRAM); got != want {
		t.Errorf("RDRAM diverged after a restore: %x != %x", got[:8], want[:8])
	}
	if ref.mi != restored.mi {
		t.Errorf("MI diverged: reference %+v, restored %+v", ref.mi, restored.mi)
	}
}

// A snapshot must refuse to load into a different cartridge, since the ROM is
// not carried in it.
func TestSaveStateRejectsAnotherCartridge(t *testing.T) {
	rom := loadTestROM(t)
	m := NewMachine(rom)
	if err := m.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "s.state")
	if err := m.SaveState(path); err != nil {
		t.Fatal(err)
	}

	other := NewMachine(rom)
	other.romMD5 = "0000000000000000000000000000dead"
	if err := other.LoadState(path); err == nil {
		t.Fatal("a snapshot loaded into a machine holding a different cartridge")
	}
}
