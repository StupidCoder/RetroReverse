package pspadapter

import (
	"bytes"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
)

// The UMD is not in the repository (copyright); these tests skip without it.
const umd = "../../../games/loco-roco-psp/image/LocoRoco.cso"

func open(t *testing.T) *Adapter {
	t.Helper()
	if _, err := os.Stat(umd); err != nil {
		t.Skip("Loco Roco UMD not present; skipping")
	}
	a, err := New(umd, Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Close() })
	return a
}

// drawnFrame steps until the game renders a scene, the way the debugger does.
func drawnFrame(t *testing.T, a *Adapter, withOverdraw bool) *debug.FrameCapture {
	t.Helper()
	for i := 0; i < 200; i++ {
		fc, err := a.StepFrame(withOverdraw)
		if err != nil {
			t.Fatal(err)
		}
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc
		}
	}
	t.Fatal("no drawn frame within the frame budget")
	return nil
}

func TestCapabilities(t *testing.T) {
	a := open(t)
	caps := debug.Capabilities(a)
	want := []string{
		debug.CapFrames, debug.CapFastStep, debug.CapReplay, debug.CapCode,
		debug.CapBreak, debug.CapDisasm, debug.CapWatch, debug.CapSurfaces,
		debug.CapFiles, debug.CapFileAt, debug.CapStates, debug.CapResume,
		debug.CapRegions, debug.CapProfile,
	}
	have := map[string]bool{}
	for _, c := range caps {
		have[c] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("missing capability %q (have %v)", w, caps)
		}
	}
}

// TestStepFrameCaptures is the basic contract: a frame has a command stream, and every
// pixel of it can say which command drew it.
func TestStepFrameCaptures(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	if fc.Width != 480 || fc.Height != 272 {
		t.Errorf("frame is %dx%d, want 480x272", fc.Width, fc.Height)
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("provenance is %d entries, want %d", len(fc.Prov), fc.Width*fc.Height)
	}

	// Every provenance entry must name a command that exists.
	attributed := 0
	for i, c := range fc.Prov {
		if c < 0 {
			continue
		}
		if int(c) >= len(fc.Commands) {
			t.Fatalf("pixel %d is attributed to command %d, but the frame has only %d", i, c, len(fc.Commands))
		}
		attributed++
	}
	if attributed == 0 {
		t.Error("not one pixel of the frame is attributed to a command")
	}
	t.Logf("%d commands, %d of %d pixels attributed", len(fc.Commands), attributed, len(fc.Prov))

	// The command that drew a pixel should be a draw, not a register write: PRIM is
	// where pixels come from. (A block transfer would be the other honest answer, but
	// it does not go through the fragment path, so it cannot appear here.)
	for _, c := range fc.Prov {
		if c >= 0 {
			if n := fc.Commands[c].Name; n != "PRIM" && n != "BEZIER" && n != "SPLINE" {
				t.Errorf("a pixel is attributed to %q, which does not draw pixels", n)
			}
			break
		}
	}
}

// TestReplayReproducesTheFrame is the gate the whole scrubber rests on. Replaying the
// captured command list from the frame's start snapshot must land on the same picture
// the live machine drew. If it does not, the snapshot is incomplete — and that is
// exactly how a missing GE register file would show itself.
func TestReplayReproducesTheFrame(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	live, err := a.Display()
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Rect != live.Rect {
		t.Fatalf("replay is %v, the live frame is %v", replayed.Rect, live.Rect)
	}
	if !bytes.Equal(replayed.Pix, live.Pix) {
		diff := 0
		for i := range live.Pix {
			if live.Pix[i] != replayed.Pix[i] {
				diff++
			}
		}
		t.Errorf("replaying the whole command list did not reproduce the frame: %d of %d bytes differ",
			diff, len(live.Pix))
	}
}

// TestScrubStopsWhereItIsAsked: a replay to command k must execute exactly k commands.
//
// This is deliberately mechanical rather than pictorial, and the reason is worth
// recording. The obvious test — "the frame looks different a third of the way through
// than at the end" — cannot work on Loco Roco's boot screen. That frame is 353 command
// words but a single PRIM: one full-screen sprite of a logo that does not move, redrawn
// into the same buffer every frame. The buffer therefore ALREADY holds the finished
// picture when the frame starts, and every scrub position looks identical to every
// other — not because the scrubber is broken, but because the game is drawing the same
// thing again. A picture cannot distinguish a replay that stopped from one that did not,
// so the count is what gets asserted.
func TestScrubStopsWhereItIsAsked(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	m, err := a.scratchMachine()
	if err != nil {
		t.Fatal(err)
	}
	s := fc.Start.(snap)
	for _, k := range []int{1, 2, len(fc.Commands) / 3, len(fc.Commands) - 1, len(fc.Commands)} {
		if err := m.LoadState(*s.ms); err != nil {
			t.Fatal(err)
		}
		n := 0
		m.OnGeCmd = func(uint32) { n++ }
		m.RunStopAfterGeCommand(k, 400_000_000)
		m.OnGeCmd = nil
		if n != k {
			t.Errorf("asked to replay %d commands, the GE executed %d", k, n)
		}
	}
}

// TestOverdrawRecordsRejects: the PSP rasteriser throws fragments away (depth, alpha,
// stencil, scissor), and the overdraw pane exists to show the ones that were thrown.
func TestOverdrawRecordsRejects(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, true)
	if fc.Overdraw == nil {
		t.Fatal("overdraw was requested and not recorded")
	}
	writes, rejects := 0, 0
	for _, ws := range fc.Overdraw {
		for _, w := range ws {
			writes++
			if w.Rejected {
				rejects++
			}
		}
	}
	if writes == 0 {
		t.Error("the overdraw record is empty")
	}
	t.Logf("%d pixel writes recorded, %d of them rejected", writes, rejects)
}

// TestProfileBucketsAreDisjoint: the buckets are meant to add up to the total. The GE
// runs inside a syscall on this machine, so a syscall bucket that failed to subtract the
// GE would double-count the whole frame and the remainder would be driven to zero.
func TestProfileBucketsAreDisjoint(t *testing.T) {
	a := open(t)
	drawnFrame(t, a, false)

	p := a.FrameProfile()
	if p.TotalMs <= 0 {
		t.Fatal("the frame profile reports no time at all")
	}
	var sum float64
	for _, b := range p.Buckets {
		if b.Millis < 0 {
			t.Errorf("bucket %q reports negative time (%.3f ms)", b.Name, b.Millis)
		}
		sum += b.Millis
	}
	// The buckets are disjoint leaves plus a derived remainder, so they sum to the
	// total. Allow a little slack for the clock reads themselves.
	if sum > p.TotalMs*1.02+0.05 {
		t.Errorf("buckets sum to %.3f ms but the frame took %.3f ms: they overlap", sum, p.TotalMs)
	}
	for _, b := range p.Buckets {
		t.Logf("  %-26s %7.3f ms  n=%d", b.Name, b.Millis, b.Count)
	}
}

// TestSnapshotIsIndependent: the debugger restores one snapshot again and again while
// the live machine runs on. A snapshot that shared state with the machine would let the
// present rewrite the past, and every replay after the first would be wrong.
func TestSnapshotIsIndependent(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	first, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	// Run the live machine well past the capture, then replay the same capture again.
	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	second, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Pix, second.Pix) {
		t.Error("replaying the same capture twice gave two different pictures: the snapshot is not independent of the live machine")
	}
}
