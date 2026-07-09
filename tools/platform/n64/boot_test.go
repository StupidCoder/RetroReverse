package n64

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
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

// TestBootReachesFirstRSPTask is the Phase 3b gate. Past the entry point the
// game's own libultra takes over: it creates threads, blocks the main one on a
// retrace message queue, and lets the idle thread spin. Only the VI's retrace
// interrupt restarts it — with no VI the boot looks like a crash in a `b .`
// loop, which is the idle thread doing its job.
//
// With VI, SI and PI modelled, the boot proceeds until it hands the RSP a
// display list. The oracle halts there rather than silently doing nothing, so
// the frontier is always explicit.
func TestBootReachesFirstRSPTask(t *testing.T) {
	rom := loadTestROM(t)
	m := NewMachine(rom)
	if err := m.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}

	fields, dmas := 0, 0
	m.OnDisplay = func(*Machine) { fields++ }
	m.OnDMA = func(string, uint32, uint32, uint32) { dmas++ }

	res := m.Run(80_000_000)
	if !strings.Contains(res.Reason, "unmodelled RSP task") {
		t.Fatalf("expected to halt on the first RSP task, got: %s", res)
	}
	if fields == 0 {
		t.Error("no video field completed: the retrace interrupt never fired")
	}
	if dmas == 0 {
		t.Error("no DMA ran: the game loaded nothing from the cartridge")
	}
	// libultra unmasks every RCP interrupt source once it is up.
	if m.mi.Mask == 0 {
		t.Error("MI_INTR_MASK is empty: libultra never enabled an interrupt")
	}
	// The VI must be scanning out a real framebuffer by now.
	if m.Origin() == 0 || m.Width() == 0 {
		t.Errorf("VI is not scanning a framebuffer: origin=%08X width=%d", m.Origin(), m.Width())
	}
	t.Logf("first RSP task after %d instructions; %d fields, %d DMAs, VI %dx? at 0x%08X, type %d",
		res.Steps, fields, dmas, m.Width(), m.Origin(), m.PixelType())
}

// The joybus must answer an empty port, or a game polling four controllers waits
// forever for the three that are not there.
func TestJoybusReportsAbsentControllers(t *testing.T) {
	rom := loadTestROM(t)
	m := NewMachine(rom)

	// A status command on every controller channel, then the end marker.
	block := []byte{
		0x01, 0x04, jbControllerState, 0xFF, 0xFF, 0xFF, 0xFF,
		0x01, 0x04, jbControllerState, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFE,
	}
	copy(m.PIF, block)
	m.PIF[pifRAMSize-1] = 1 // the execute flag
	m.Controllers[0] = Controller{Present: true, Buttons: BtnStart | BtnA, StickX: 40, StickY: -20}
	m.Controllers[1] = Controller{Present: false}
	m.joybus()

	if got := uint16(m.PIF[3])<<8 | uint16(m.PIF[4]); got != BtnStart|BtnA {
		t.Errorf("port 1 buttons: got %04X want %04X", got, uint16(BtnStart|BtnA))
	}
	if int8(m.PIF[5]) != 40 || int8(m.PIF[6]) != -20 {
		t.Errorf("port 1 stick: got %d,%d want 40,-20", int8(m.PIF[5]), int8(m.PIF[6]))
	}
	// Port 2 is empty: its receive-length byte must carry the no-response flag.
	if m.PIF[8]&0x80 == 0 {
		t.Errorf("port 2 is empty but reported a response (rx byte %02X)", m.PIF[8])
	}
	// The execute flag is consumed, so the read-back transfer does not re-run it.
	if m.PIF[pifRAMSize-1]&1 != 0 {
		t.Error("the joybus execute flag was not cleared")
	}
}

// osEepromProbe issues an Info command on channel 4 and blocks on SI; the save
// device must identify itself.
func TestJoybusEEPROMIdentifies(t *testing.T) {
	rom := loadTestROM(t)
	m := NewMachine(rom)

	// Skip the four controller channels, then Info on the save device.
	block := []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x03, jbInfo, 0xFF, 0xFF, 0xFF, 0xFE}
	copy(m.PIF, block)
	m.PIF[pifRAMSize-1] = 1
	m.joybus()

	if got := uint16(m.PIF[7])<<8 | uint16(m.PIF[8]); got != devEEPROM4K {
		t.Errorf("EEPROM identity: got %04X want %04X (4 kbit)", got, devEEPROM4K)
	}

	// And a write followed by a read must round-trip a block.
	want := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	copy(m.EEPROM[3*eepromBlockSize:], want)
	block = []byte{0x00, 0x00, 0x00, 0x00, 0x02, 0x08, jbEepromRead, 3,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}
	for i := range m.PIF {
		m.PIF[i] = 0
	}
	copy(m.PIF, block)
	m.PIF[pifRAMSize-1] = 1
	m.joybus()
	if got := m.PIF[8:16]; !bytes.Equal(got, want) {
		t.Errorf("EEPROM read of block 3: got %v want %v", got, want)
	}
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
