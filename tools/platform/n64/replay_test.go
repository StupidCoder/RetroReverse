package n64

// replay_test.go covers the pieces the frame debugger's command scrubber is built
// on: stopping an RDP drain precisely after the k-th command, and rendering the
// RDP's draw target (the back buffer) at that moment rather than the VI scanout.

import (
	"bytes"
	"testing"
)

// drawTriOps are the eight triangle opcodes, all of which count as "drawing".
var drawTriOps = map[uint32]bool{
	cmdTriFill: true, cmdTriFillZ: true, cmdTriTex: true, cmdTriTexZ: true,
	cmdTriShade: true, cmdTriShadeZ: true, cmdTriShadeTex: true, cmdTriShadeTexZ: true,
}

// snapshotAtFrameStart boots and runs until the first Sync_Full that followed real
// drawing, then returns an in-memory snapshot taken at that frame boundary — a
// clean start point for replaying the frame that follows.
func snapshotAtFrameStart(t *testing.T, rom *ROM) *MachineState {
	t.Helper()
	m := NewMachine(rom)
	if err := m.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	var snap *MachineState
	drew := false
	m.OnRDPCmd = func(mm *Machine, op uint32, _ []uint64) {
		if drawTriOps[op] {
			drew = true
		}
		if op == cmdSyncFull && drew && snap == nil {
			snap = mm.SnapshotState()
			mm.StopRequested = true
		}
	}
	m.Run(400_000_000)
	if snap == nil {
		t.Fatal("never reached a drawn frame")
	}
	return snap
}

// TestRDPReplayStopsAfterCommand is the core of the scrubber: replaying a frame
// from a snapshot must stop the RDP drain exactly after the k-th command, and the
// draw target rendered at that point must differ from the finished frame.
func TestRDPReplayStopsAfterCommand(t *testing.T) {
	rom := loadTestROM(t)
	snap := snapshotAtFrameStart(t, rom)

	// Full replay: count every RDP command of the next frame, up to its Sync_Full,
	// then render the finished draw target.
	total := 0
	full := NewMachine(rom)
	if err := full.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	if err := full.RestoreState(snap); err != nil {
		t.Fatal(err)
	}
	done := false
	full.OnRDPCmd = func(mm *Machine, op uint32, _ []uint64) {
		if done {
			return
		}
		total++
		if op == cmdSyncFull {
			done = true
			mm.StopRequested = true
		}
	}
	full.Run(400_000_000)
	if total < 10 {
		t.Fatalf("frame held only %d RDP commands; expected a full scene", total)
	}
	fullImg, err := full.RenderColorImage()
	if err != nil {
		t.Fatalf("RenderColorImage (full): %v", err)
	}

	// Partial replay: stop right after command k = total/2.
	k := total / 2
	part := NewMachine(rom)
	if err := part.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	if err := part.RestoreState(snap); err != nil {
		t.Fatal(err)
	}
	seen, drawn, fills := 0, 0, 0
	part.OnPixel = func(_, _ uint32, ev PixelEvent) {
		if ev.Drawn {
			drawn++
		}
	}
	part.OnRDPCmd = func(_ *Machine, op uint32, _ []uint64) {
		seen++
		if op == cmdFillRect {
			fills++
		}
	}
	part.RunStopAfterRDPCommand(k, 400_000_000)

	if seen != k {
		t.Errorf("replay executed %d commands, want exactly %d", seen, k)
	}
	partImg, err := part.RenderColorImage()
	if err != nil {
		t.Fatalf("RenderColorImage (partial): %v", err)
	}
	if partImg.Bounds() != fullImg.Bounds() {
		t.Fatalf("partial/full bounds differ: %v vs %v", partImg.Bounds(), fullImg.Bounds())
	}
	if bytes.Equal(partImg.Pix, fullImg.Pix) {
		t.Error("the frame halfway through its commands is identical to the finished frame — the stop did nothing")
	}
	if drawn == 0 {
		t.Error("no pixels were drawn during the partial replay")
	}
	t.Logf("frame has %d RDP commands (%d Fill_Rectangle in the first half); "+
		"after %d commands, %d pixels had been drawn", total, fills, k, drawn)
}

// TestRDPReplayDeterministic pins the property the scrubber relies on: replaying
// the same frame to the same command twice yields byte-identical draw targets.
func TestRDPReplayDeterministic(t *testing.T) {
	rom := loadTestROM(t)
	snap := snapshotAtFrameStart(t, rom)

	render := func(k int) []byte {
		m := NewMachine(rom)
		if err := m.Boot(rom, DefaultBoot()); err != nil {
			t.Fatal(err)
		}
		if err := m.RestoreState(snap); err != nil {
			t.Fatal(err)
		}
		m.RunStopAfterRDPCommand(k, 400_000_000)
		img, err := m.RenderColorImage()
		if err != nil {
			t.Fatalf("RenderColorImage: %v", err)
		}
		return img.Pix
	}

	if a, b := render(40), render(40); !bytes.Equal(a, b) {
		t.Error("replaying the same frame to the same command produced different pixels")
	}
}
