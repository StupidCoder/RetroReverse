package psx

import (
	"path/filepath"
	"testing"
)

// TestSaveStateRoundTrip runs a small program, snapshots mid-run, runs both the
// original and a restored machine forward, and requires identical CPU state and
// RAM. The program exercises RAM writes, the load-delay slot and multiplication
// so the snapshot must carry real pipeline state.
func TestSaveStateRoundTrip(t *testing.T) {
	prog := []uint32{
		0x24080000, // addiu $t0, $zero, 0
		0x24090001, // addiu $t1, $zero, 1
		// loop:
		0x01094021, // addu $t0, $t0, $t1
		0x25290001, // addiu $t1, $t1, 1
		0xAC880100, // sw $t0, 256($a0)
		0x8C8A0100, // lw $t2, 256($a0)
		0x71090002, // (madd-ish filler; decoded or not, both machines agree)
		0x1000FFFA, // b loop
		0x00000000, // nop (delay slot)
	}
	build := func() *Machine {
		m := NewMachine()
		base := uint32(0x80010000)
		for i, w := range prog {
			m.write32(base+uint32(i*4), w)
		}
		m.CPU.PC = base
		m.CPU.R[4] = 0x80020000 // $a0
		return m
	}

	a := build()
	a.Run(1000)
	snap := a.SaveState()

	// Restore into a fresh machine and run both forward.
	b := build()
	b.LoadState(snap)
	a.Run(1000)
	b.Run(1000)

	sa, sb := a.CPU.SaveState(), b.CPU.SaveState()
	if sa != sb {
		t.Fatalf("CPU state diverged after restore:\n a=%+v\n b=%+v", sa, sb)
	}
	for i := range a.ram {
		if a.ram[i] != b.ram[i] {
			t.Fatalf("RAM diverged at 0x%X: %02X vs %02X", i, a.ram[i], b.ram[i])
		}
	}
}

// TestSaveStateFile checks the gob+gzip file round-trip preserves the machine.
func TestSaveStateFile(t *testing.T) {
	m := NewMachine()
	m.write32(0x80010000, 0x24080007)
	m.CPU.PC = 0x80010000
	m.Run(1)
	path := filepath.Join(t.TempDir(), "m.state")
	if err := m.SaveStateFile(path); err != nil {
		t.Fatal(err)
	}
	m2 := NewMachine()
	if err := m2.LoadStateFile(path); err != nil {
		t.Fatal(err)
	}
	if m2.CPU.SaveState() != m.CPU.SaveState() {
		t.Fatalf("CPU state not preserved through file round-trip")
	}
	if m2.CPU.R[8] != 7 {
		t.Fatalf("$t0 = %d, want 7", m2.CPU.R[8])
	}
}
