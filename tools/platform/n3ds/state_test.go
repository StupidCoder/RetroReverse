package n3ds

import (
	"os"
	"path/filepath"
	"testing"
)

// A savestate must round-trip exactly, and — because the oracle paces by
// instruction count — resuming a snapshot and running N more instructions must
// land on the same state as running N instructions past the save point without
// interruption. That determinism is what makes a snapshot a regression tool, and
// this test is what enforces that every execution-affecting field is serialised.
func TestSaveStateRoundTrip(t *testing.T) {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		t.Skip("Super Mario 3D Land image not present (game images are not committed)")
	}

	m, err := NewMachine(img)
	if err != nil {
		t.Fatal(err)
	}
	// Run a chunk of the boot so the state is non-trivial (heap allocated, VFP
	// registers written, handles created).
	m.Run(150000)

	dir := t.TempDir()
	path := filepath.Join(dir, "snap.gz")
	if err := m.SaveState(path); err != nil {
		t.Fatal(err)
	}

	// Reference: keep running the original another 20000 instructions.
	const more = 20000
	m.Run(more)
	wantPC := m.CPU.R[15]
	wantInstrs := m.CPU.Instrs
	wantVFP := m.CPU.VFP.S
	wantR := m.CPU.R

	// Restore into a fresh machine and run the same amount; it must match.
	m2, err := NewMachine(img)
	if err != nil {
		t.Fatal(err)
	}
	if err := m2.LoadState(path); err != nil {
		t.Fatal(err)
	}
	m2.Run(more)

	if m2.CPU.R[15] != wantPC {
		t.Errorf("PC after resume = 0x%08X, want 0x%08X", m2.CPU.R[15], wantPC)
	}
	if m2.CPU.Instrs != wantInstrs {
		t.Errorf("instruction count after resume = %d, want %d", m2.CPU.Instrs, wantInstrs)
	}
	if m2.CPU.R != wantR {
		t.Errorf("core registers diverged after resume")
	}
	if m2.CPU.VFP.S != wantVFP {
		t.Errorf("VFP registers diverged after resume (VFP state not fully serialised?)")
	}
}

// A snapshot from one title must not restore into another.
func TestLoadStateRejectsWrongProgram(t *testing.T) {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		t.Skip("image not present")
	}
	m, err := NewMachine(img)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "snap.gz")
	if err := m.SaveState(path); err != nil {
		t.Fatal(err)
	}
	m.programID = 0xDEADBEEF // pretend this machine is a different title
	if err := m.LoadState(path); err == nil {
		t.Fatal("expected LoadState to reject a snapshot from a different program ID")
	}
}
