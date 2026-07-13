package threedo

// The savestate has to carry the whole machine, and on this platform most of the machine
// is not in memory. The Portfolio OS is high-level-emulated: the item table, the task
// list, the heaps, the open streams and the graphics folio's bitmaps are Go objects on
// the side. A snapshot that forgot any of them would still restore something that runs —
// and then quietly diverge, several thousand instructions later, from the machine it
// claimed to be.
//
// So the test is not "do the fields round-trip" (they would, by construction) but "does
// the restored machine go on to do the same thing". Clone a machine mid-run, run both,
// and demand they stay identical down to the last byte of VRAM. A forgotten field shows
// up as a divergence rather than as a passing test.

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"testing"
)

const nfsDisc = "../../../games/need-for-speed-3do/Need for Speed.bin"

// booted brings a machine up the way the oracle does, and runs it to a point where the
// HLE has real state: tasks spawned, items created, the disc read, cels drawn.
func booted(t *testing.T, steps uint64) *Machine {
	t.Helper()
	data, err := os.ReadFile(nfsDisc)
	if err != nil {
		t.Skip("Need for Speed disc not present; skipping")
	}
	vol, err := Open(data)
	if err != nil {
		t.Fatal(err)
	}
	prog, err := vol.ReadFile("LaunchMe")
	if err != nil {
		t.Fatal(err)
	}
	aif, err := ParseAIF(prog)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine()
	m.SetImageHash(fmt.Sprintf("%x", md5.Sum(data)))
	m.NoStreams = true
	m.StallTolerance = 8
	m.SetVolume(vol)
	m.SetVBLMirror(0x42734) // Need for Speed's field counter (the oracle's -vblmirror)
	m.LoadAIF(aif)
	m.Run(steps)
	return m
}

// TestSaveStateIsComplete is the whole point of the file: a machine restored from a
// snapshot must go on to draw exactly what the machine it was cloned from drew.
func TestSaveStateIsComplete(t *testing.T) {
	m := booted(t, 40_000_000)
	t.Logf("cloning after %d instructions: %d items, %d tasks, %d streams, %d bitmaps",
		m.CPU.Instrs, len(m.items), len(m.tasks), len(m.streams), len(m.bitmaps))

	s := m.SaveState()

	m2 := NewMachine()
	m2.SetImageHash(m.imageHash)
	m2.NoStreams = true
	m2.StallTolerance = 8
	m2.SetVolume(m.vol)
	if err := m2.LoadState(s); err != nil {
		t.Fatal(err)
	}

	// Both machines now run the same number of instructions from the same state.
	const more = 8_000_000
	r1 := m.Run(more)
	r2 := m2.Run(more)

	if r1.Steps != r2.Steps || r1.PC != r2.PC {
		t.Errorf("the restored machine diverged:\n  live:     %d steps, PC %08X, %s\n  restored: %d steps, PC %08X, %s",
			r1.Steps, r1.PC, r1.Reason, r2.Steps, r2.PC, r2.Reason)
	}
	if !bytes.Equal(m.dram, m2.dram) {
		t.Errorf("DRAM differs after the same run: %s", firstDiff(m.dram, m2.dram))
	}
	if !bytes.Equal(m.vram, m2.vram) {
		t.Errorf("VRAM differs after the same run — the restored machine drew a different picture: %s",
			firstDiff(m.vram, m2.vram))
	}
	if m.CPU.Instrs != m2.CPU.Instrs {
		t.Errorf("instruction counts differ: %d vs %d", m.CPU.Instrs, m2.CPU.Instrs)
	}
}

// TestSaveStateFileRoundTrip: the on-disk format carries what the in-memory one does, so
// a state saved in the frame debugger can be handed to the oracle.
func TestSaveStateFileRoundTrip(t *testing.T) {
	m := booted(t, 20_000_000)

	path := t.TempDir() + "/s.state"
	if err := m.SaveStateFile(path); err != nil {
		t.Fatal(err)
	}
	m2 := NewMachine()
	m2.SetImageHash(m.imageHash)
	m2.NoStreams = true
	m2.SetVolume(m.vol)
	if err := m2.LoadStateFile(path); err != nil {
		t.Fatal(err)
	}

	if m2.CPU.Instrs != m.CPU.Instrs || m2.CPU.Reg(15) != m.CPU.Reg(15) {
		t.Errorf("restored PC/steps mismatch: %08X/%d vs %08X/%d",
			m2.CPU.Reg(15), m2.CPU.Instrs, m.CPU.Reg(15), m.CPU.Instrs)
	}
	if !bytes.Equal(m.dram, m2.dram) {
		t.Error("restored DRAM differs")
	}
	if len(m2.items) != len(m.items) || len(m2.tasks) != len(m.tasks) {
		t.Errorf("restored HLE differs: %d items/%d tasks vs %d/%d",
			len(m2.items), len(m2.tasks), len(m.items), len(m.tasks))
	}

	// A state from another disc must be refused, not silently resumed.
	m3 := NewMachine()
	m3.SetImageHash("some-other-disc")
	if err := m3.LoadStateFile(path); err == nil {
		t.Error("load accepted a state from a different disc")
	}
}

// TestSnapshotIsIndependent: the debugger restores one snapshot again and again while the
// live machine runs on. A snapshot sharing a map with the machine would let the present
// rewrite the past.
func TestSnapshotIsIndependent(t *testing.T) {
	m := booted(t, 20_000_000)
	s := m.SaveState()
	nItems, nTasks := len(s.Items), len(s.Tasks)

	m.Run(8_000_000) // the live machine runs on, creating and destroying HLE objects

	if len(s.Items) != nItems || len(s.Tasks) != nTasks {
		t.Error("the snapshot changed while the live machine ran: it is not independent")
	}
	// And it can be restored twice to the same place.
	m2, m3 := NewMachine(), NewMachine()
	for _, mm := range []*Machine{m2, m3} {
		mm.SetImageHash(m.imageHash)
		mm.NoStreams = true
		mm.StallTolerance = 8
		mm.SetVolume(m.vol)
		if err := mm.LoadState(s); err != nil {
			t.Fatal(err)
		}
		mm.Run(2_000_000)
	}
	if !bytes.Equal(m2.vram, m3.vram) {
		t.Error("two machines restored from one snapshot drew different pictures")
	}
}

func firstDiff(a, b []byte) string {
	n := 0
	first := -1
	for i := range a {
		if a[i] != b[i] {
			if first < 0 {
				first = i
			}
			n++
		}
	}
	return fmt.Sprintf("%d of %d bytes differ, first at 0x%X", n, len(a), first)
}
