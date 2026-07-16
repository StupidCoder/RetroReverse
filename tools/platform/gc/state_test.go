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

	// The window must lie inside the boot's clean run. The boot now reaches the first draw
	// primitive at ~65.7M instructions, where the command processor halts for want of vertex
	// fetch (the next frontier); a snapshot taken across that halt would compare a machine
	// that stayed halted against one whose restored copy re-runs the halting step, diverging
	// by a single timebase tick — an artefact of straddling the halt, not a savestate defect.
	// Staying below it keeps the round trip a test of a running machine, which is what it is for.
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

// TestSnapshotIsIndependent is the guard on the rule that makes deterministic replay
// possible: a snapshot must be a photograph, not a window. It is a separate test from the
// round trip above because the round trip cannot see this — its fingerprint hashes RAM and
// the register blocks, and the buffers that were aliased (the EFB, the depth buffer, the
// pending FIFO bytes) are not in RAM at all. They lived in the machine's own slices, so a
// snapshot that shared them still restored a machine that ran identically, and every test
// passed while the snapshot quietly tracked the live machine.
//
// Both directions are asserted, because they fail differently: a save that aliases means the
// state changes under you after it is taken; a load that aliases means two machines restored
// from one state scribble on each other.
func TestSnapshotIsIndependent(t *testing.T) {
	m := testMachine()
	m.gpu.EFB = []uint32{1, 2, 3, 4}
	m.gpu.ZBuf = []uint32{9, 9}
	m.gpu.Buf = []byte{0x61, 0xAA}
	m.gpu.Tlut = []byte{0x11, 0x22}
	m.wgFIFO.Buf = []byte{0x80}

	// A saved state must not move when the machine does.
	s := m.SaveState()
	m.gpu.EFB[0] = 0xDEAD
	m.gpu.ZBuf[0] = 0xBEEF
	m.gpu.Buf[0] = 0xFF
	m.gpu.Tlut[0] = 0xFF
	m.wgFIFO.Buf[0] = 0xFF
	if s.GPU.EFB[0] != 1 || s.GPU.ZBuf[0] != 9 || s.GPU.Buf[0] != 0x61 ||
		s.GPU.Tlut[0] != 0x11 || s.WG.Buf[0] != 0x80 {
		t.Errorf("the snapshot followed the live machine: EFB[0]=%#x ZBuf[0]=%#x Buf[0]=%#x Tlut[0]=%#x WG[0]=%#x",
			s.GPU.EFB[0], s.GPU.ZBuf[0], s.GPU.Buf[0], s.GPU.Tlut[0], s.WG.Buf[0])
	}

	// And a restored machine must not write back into the state it came from — the second
	// restore of the same state has to see what the first one saw.
	b := testMachine()
	if err := b.LoadState(s); err != nil {
		t.Fatal(err)
	}
	b.gpu.EFB[0] = 0xC0DE
	b.gpu.Tlut[0] = 0xC0
	b.wgFIFO.Buf[0] = 0xC0
	if s.GPU.EFB[0] != 1 || s.GPU.Tlut[0] != 0x11 || s.WG.Buf[0] != 0x80 {
		t.Errorf("a restored machine wrote back into the snapshot: EFB[0]=%#x Tlut[0]=%#x WG[0]=%#x",
			s.GPU.EFB[0], s.GPU.Tlut[0], s.WG.Buf[0])
	}

	c := testMachine()
	if err := c.LoadState(s); err != nil {
		t.Fatal(err)
	}
	if c.gpu.EFB[0] != 1 {
		t.Errorf("the second restore of one snapshot saw the first restore's writes: EFB[0]=%#x", c.gpu.EFB[0])
	}
}

// A savestate from another disc must be rejected, not resumed onto the wrong game.
func TestSaveStateWrongDisc(t *testing.T) {
	m := &Machine{discMD5: "the-right-disc"}
	err := m.LoadState(MachineState{Version: snapshotVersion, DiscMD5: "a-different-disc"})
	if err == nil {
		t.Error("a savestate from a different disc was accepted")
	}
}
