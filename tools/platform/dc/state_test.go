package dc

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// plantLoop lays a small program at 8C010000 that marches two pointers
// through RAM and VRAM, writing an incrementing counter — enough churn that
// any lost state moves the fingerprint.
func plantLoop(m *Machine) {
	prog := []uint16{
		0xE000, // mov #0, r0
		0xD104, // mov.l @(pool1,pc), r1 -> 8C020000
		0xD204, // mov.l @(pool2,pc), r2 -> A5000000 (VRAM, 32-bit window)
		0x7001, // loop: add #1, r0
		0x2102, // mov.l r0, @r1
		0x2202, // mov.l r0, @r2
		0x7104, // add #4, r1
		0x7204, // add #4, r2
		0xAFF9, // bra loop
		0x0009, // (slot) nop
		0x0009,
	}
	for i, h := range prog {
		binary.LittleEndian.PutUint16(m.RAM[0x10000+i*2:], h)
	}
	binary.LittleEndian.PutUint32(m.RAM[0x10014:], 0x8C020000)
	binary.LittleEndian.PutUint32(m.RAM[0x10018:], 0xA5000000)
	m.CPU.SetPC(0x8C010000)
}

// fingerprint hashes everything a resumed run's future depends on. The
// failure mode it defends against: a device field in Machine but not in
// MachineState.
func fingerprint(m *Machine) [32]byte {
	h := sha256.New()
	h.Write(m.RAM)
	h.Write(m.VRAM)
	h.Write(m.AICARAM)
	s := m.CPU.Snapshot()
	var scalars []uint32
	scalars = append(scalars, s.R[:]...)
	scalars = append(scalars, s.Rbank[:]...)
	scalars = append(scalars, s.SR, s.GBR, s.VBR, s.SSR, s.SPC, s.SGR, s.DBR,
		s.MACH, s.MACL, s.PR, s.PC, s.NextPC, s.FPSCR, s.FPUL)
	for _, bank := range s.FPR {
		scalars = append(scalars, bank[:]...)
	}
	for _, q := range s.SQ {
		scalars = append(scalars, q[:]...)
	}
	scalars = append(scalars, uint32(s.Steps), uint32(s.Steps>>32),
		uint32(m.Instrs), uint32(m.Fields), m.Holly.ISTNRM, m.PVRRegs[fbRCtrl])
	if s.PendingDelay {
		scalars = append(scalars, 1)
	}
	var buf [4]byte
	for _, v := range scalars {
		binary.LittleEndian.PutUint32(buf[:], v)
		h.Write(buf[:])
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func TestSaveStateRoundTrip(t *testing.T) {
	m := NewMachine(nil)
	plantLoop(m)
	if r := m.Run(1000, RunConfig{}); r.Reason != "steps" {
		t.Fatalf("setup run: %v", r)
	}
	s := m.SaveState()
	if r := m.Run(500, RunConfig{}); r.Reason != "steps" {
		t.Fatalf("continue run: %v", r)
	}
	want := fingerprint(m)

	m2 := NewMachine(nil)
	if err := m2.LoadState(s); err != nil {
		t.Fatal(err)
	}
	if r := m2.Run(500, RunConfig{}); r.Reason != "steps" {
		t.Fatalf("restored run: %v", r)
	}
	if got := fingerprint(m2); got != want {
		t.Fatalf("restored run diverged from the continued run — a device field is probably in Machine but not in MachineState")
	}

	// And the file round-trip.
	path := filepath.Join(t.TempDir(), "state.st")
	if err := m2.SaveStateFile(path); err != nil {
		t.Fatal(err)
	}
	m3 := NewMachine(nil)
	if err := m3.LoadStateFile(path); err != nil {
		t.Fatal(err)
	}
	if got := fingerprint(m3); got != want {
		t.Fatalf("file round-trip changed the state")
	}
}

// TestSnapshotIsIndependent is the aliasing guard, both directions. The
// round-trip test cannot catch a shared backing array: a snapshot that
// aliases the live machine still resumes identically.
func TestSnapshotIsIndependent(t *testing.T) {
	m := NewMachine(nil)
	plantLoop(m)
	m.Run(100, RunConfig{})
	s := m.SaveState()

	// Mutating the live machine must not move the snapshot.
	m.RAM[0]++
	m.VRAM[0]++
	m.AICARAM[0]++
	m.Flash[0]++
	if s.RAM[0] == m.RAM[0] || s.VRAM[0] == m.VRAM[0] || s.AICARAM[0] == m.AICARAM[0] || s.Flash[0] == m.Flash[0] {
		t.Fatalf("snapshot aliases the live machine")
	}

	// Mutating a restored machine must not move the state...
	m2 := NewMachine(nil)
	if err := m2.LoadState(s); err != nil {
		t.Fatal(err)
	}
	m2.RAM[1]++
	if s.RAM[1] == m2.RAM[1] {
		t.Fatalf("restore aliased the state's buffers")
	}
	// ...and a second restore from the same state must not see the first's writes.
	m3 := NewMachine(nil)
	if err := m3.LoadState(s); err != nil {
		t.Fatal(err)
	}
	if m3.RAM[1] == m2.RAM[1] {
		t.Fatalf("two restores share buffers")
	}
}

func TestSaveStateWrongDisc(t *testing.T) {
	cue1 := buildSyntheticGD(t)
	cue2 := buildSyntheticGD(t)
	// Make the second disc differ by one byte of file content.
	bin2 := filepath.Join(filepath.Dir(cue2), "synthetic.bin")
	raw, err := os.ReadFile(bin2)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-100] ^= 0xFF
	if err := os.WriteFile(bin2, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	d1, err := OpenDisc(cue1)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := OpenDisc(cue2)
	if err != nil {
		t.Fatal(err)
	}
	m1 := NewMachine(d1)
	plantLoop(m1)
	s := m1.SaveState()

	m2 := NewMachine(d2)
	if err := m2.LoadState(s); err == nil {
		t.Fatalf("a foreign disc's savestate loaded without complaint")
	}
	m3 := NewMachine(d1)
	if err := m3.LoadState(s); err != nil {
		t.Fatalf("the right disc's savestate was rejected: %v", err)
	}
}
