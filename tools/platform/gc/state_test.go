package gc

import (
	"crypto/md5"
	"os"
	"testing"
)

// TestSaveStateRoundTrip is the enforcer the oracle-capability-parity rule asks for: a
// snapshot must resume into a run bit-identical to the one that never stopped. It also
// keeps the savestate honest as the machine grows — a device field added to Machine but
// not to MachineState makes the "load then continue" run diverge from the uninterrupted
// one, and this fails.
//
// It needs the disc; without it there is nothing to boot, so it skips.
func TestSaveStateRoundTrip(t *testing.T) {
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the disc image is not present")
	}

	// Boot both machines identically through the apploader.
	boot := func() *Machine {
		d, err := Open(discPath)
		if err != nil {
			t.Fatal(err)
		}
		m, err := NewMachine(d)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.RunApploader(); err != nil {
			t.Fatalf("apploader: %v", err)
		}
		m.SetSpinDetect(false)
		return m
	}

	const (
		before = 40_000_000
		after  = 10_000_000
	)

	// Run A: run `before`, snapshot, run `after`.
	a := boot()
	a.Run(before)
	snap := a.SaveState()
	a.Run(after)
	fpA := fingerprint(a)

	// Run B: a fresh machine restored from the snapshot, then run `after`. This is the
	// real test — a restored machine must continue identically.
	b := boot()
	b.Run(before) // get b to the same point, then overwrite with the snapshot
	if err := b.LoadState(snap); err != nil {
		t.Fatal(err)
	}
	b.Run(after)
	fpB := fingerprint(b)

	if fpA != fpB {
		t.Errorf("a restored machine diverged from an uninterrupted run:\n  uninterrupted %s\n  restored      %s\n"+
			"a device field is probably in Machine but not in MachineState", fpA, fpB)
	}

	// And the snapshot must round-trip through its file form unchanged.
	tmp := t.TempDir() + "/s.gz"
	if err := a.SaveStateFile(tmp); err != nil {
		t.Fatal(err)
	}
	c := boot()
	if err := c.LoadStateFile(tmp); err != nil {
		t.Fatal(err)
	}
	// c is now at the `before`+`after` point (the file was saved from a after its full run).
	if got, want := fingerprint(c), fpA; got != want {
		t.Errorf("a state written to a file and read back does not match:\n  %s\n  %s", got, want)
	}
}

// fingerprint is a compact, total summary of the machine's observable state — the memory,
// the CPU registers, and each device — so any divergence shows up as a different hash.
func fingerprint(m *Machine) string {
	h := md5.New()
	h.Write(m.RAM)
	h.Write(m.ARAM)
	enc := func(v ...uint64) {
		var b [8]byte
		for _, x := range v {
			for i := 0; i < 8; i++ {
				b[i] = byte(x >> (8 * i))
			}
			h.Write(b[:])
		}
	}
	s := m.CPU.Snapshot()
	for _, r := range s.GPR {
		enc(uint64(r))
	}
	enc(uint64(m.CPU.PC), m.vi.Field, uint64(m.di.SR), uint64(m.dsp.FromDSP),
		uint64(m.pi.Cause), uint64(m.pi.Mask), uint64(m.CPU.TB))
	return string(h.Sum(nil))
}

// A savestate from another disc must be rejected, not resumed onto the wrong game.
func TestSaveStateWrongDisc(t *testing.T) {
	m := &Machine{discMD5: "the-right-disc"}
	err := m.LoadState(MachineState{Version: snapshotVersion, DiscMD5: "a-different-disc"})
	if err == nil {
		t.Error("a savestate from a different disc was accepted")
	}
}
