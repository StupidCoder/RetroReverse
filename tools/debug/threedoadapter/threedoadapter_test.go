package threedoadapter

import (
	"bytes"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/threedo"
)

// The disc is not in the repository (copyright); these tests skip without it.
const disc = "../../../games/need-for-speed-3do/Need for Speed.bin"

// needForSpeed is what the game's debug.json says, and what the oracle's flags say: the
// AIF to boot, and where the game keeps the field counter the OS would advance.
var opts = Options{Prog: "LaunchMe", VBLMirror: 0x42734, StallTolerance: 8}

func open(t *testing.T) *Adapter {
	t.Helper()
	if _, err := os.Stat(disc); err != nil {
		t.Skip("Need for Speed disc not present; skipping")
	}
	a, err := New(disc, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Close() })
	return a
}

// drawnFrame steps until the game draws a real scene, the way the debugger does.
func drawnFrame(t *testing.T, a *Adapter, withOverdraw bool) *debug.FrameCapture {
	t.Helper()
	for i := 0; i < 400; i++ {
		fc, err := a.StepFrame(withOverdraw)
		if err != nil {
			t.Fatal(err)
		}
		if len(fc.Commands) > 8 && fc.Prov != nil {
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
		debug.CapRegions, debug.CapProfile, debug.CapKeys,
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

// TestStepFrameCaptures: a frame is a list of cels, and every pixel can name the cel that
// drew it.
func TestStepFrameCaptures(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	if fc.Width != 320 || fc.Height != 240 {
		t.Errorf("frame is %dx%d, want 320x240", fc.Width, fc.Height)
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("provenance is %d entries, want %d", len(fc.Prov), fc.Width*fc.Height)
	}
	attributed := 0
	for i, c := range fc.Prov {
		if c < 0 {
			continue
		}
		if int(c) >= len(fc.Commands) {
			t.Fatalf("pixel %d is attributed to cel %d, but the frame drew only %d", i, c, len(fc.Commands))
		}
		attributed++
	}
	if attributed == 0 {
		t.Error("not one pixel of the frame is attributed to a cel")
	}
	t.Logf("%d cels, %d of %d pixels attributed", len(fc.Commands), attributed, len(fc.Prov))
	t.Logf("first cel: %s — %s", fc.Commands[0].Name, fc.Commands[0].Decoded)
}

// TestReplayReproducesTheFrame is the gate the whole scrubber rests on: replaying every
// cel from the frame's start snapshot must land on the picture the live machine drew.
// If it does not, the snapshot is incomplete — and on this machine, where most of the
// state is the Portfolio HLE rather than memory, that is the thing most likely to be
// wrong.
func TestReplayReproducesTheFrame(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	replayed, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	// The live machine's own copy of the same bitmap — the one the capture decided was
	// the frame.
	live, err := a.live.RenderBitmap(a.frame.buf, a.frame.w, a.frame.h)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(replayed.Pix, live.Pix) {
		diff := 0
		for i := range live.Pix {
			if live.Pix[i] != replayed.Pix[i] {
				diff++
			}
		}
		t.Errorf("replaying all %d cels did not reproduce the frame: %d of %d bytes differ",
			len(fc.Commands), diff, len(live.Pix))
	}
}

// TestScrubStopsWhereItIsAsked: a replay to cel k must draw exactly k cels.
func TestScrubStopsWhereItIsAsked(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	m := a.scratchMachine()
	s := fc.Start.(snap)
	for _, k := range []int{1, 2, len(fc.Commands) / 2, len(fc.Commands)} {
		if k <= 0 {
			continue
		}
		if err := m.LoadState(*s.ms); err != nil {
			t.Fatal(err)
		}
		n := 0
		m.OnCel = func(threedo.CelDraw) { n++ }
		m.RunStopAfterCel(k, 400_000_000)
		m.OnCel = nil
		if n != k {
			t.Errorf("asked to replay %d cels, the engine drew %d", k, n)
		}
	}
}

// TestScrubBuildsThePicture: the frame after its first cel must differ from the frame
// after all of them. Unlike Loco Roco's static boot logo, a 3DO frame starts with a
// FlashClear and then paints — so the picture genuinely changes as the chain runs, and
// the scrubber can be checked on the pixels themselves.
func TestScrubBuildsThePicture(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)
	if len(fc.Commands) < 3 {
		t.Skip("frame too short to scrub")
	}

	first, err := a.RenderAfter(fc, 0)
	if err != nil {
		t.Fatal(err)
	}
	last, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first.Pix, last.Pix) {
		t.Error("the frame after one cel looks exactly like the frame after all of them")
	}
}

// TestNoRejectedWrites: the cel engine has no depth or alpha test, so a pixel it reports
// is a pixel it wrote. The overdraw pane must never claim otherwise.
func TestNoRejectedWrites(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, true)
	if fc.Overdraw == nil {
		t.Fatal("overdraw was requested and not recorded")
	}
	writes := 0
	for _, ws := range fc.Overdraw {
		for _, w := range ws {
			writes++
			if w.Rejected {
				t.Fatal("a write was reported as rejected, but this rasteriser rejects nothing")
			}
		}
	}
	if writes == 0 {
		t.Error("the overdraw record is empty")
	}
	t.Logf("%d pixel writes recorded, none rejected (as it should be)", writes)
}

// TestProfileBucketsAreDisjoint: DrawCels is a folio call that draws a whole chain before
// it returns, so a folio bucket that failed to subtract the cel engine would double-count
// the frame.
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
	if sum > p.TotalMs*1.02+0.05 {
		t.Errorf("buckets sum to %.3f ms but the frame took %.3f ms: they overlap", sum, p.TotalMs)
	}
	for _, b := range p.Buckets {
		t.Logf("  %-26s %7.3f ms  n=%d", b.Name, b.Millis, b.Count)
	}
	for _, c := range p.Counters {
		t.Logf("  #%-24s %d", c.Name, c.Value)
	}
}

// TestKeyerQueue exercises the gamepad mapping/queue without booting: keys build
// the right button mask, holding coalesces, and releasePad delivers each distinct
// state once, in order (press → release both reach the guest).
func TestKeyerQueue(t *testing.T) {
	a := &Adapter{held: map[string]bool{}, live: threedo.NewMachine()}

	// Enter maps to Start; pressing it queues exactly the Start mask.
	a.Key(debug.Key{Name: "Enter", Down: true})
	if a.padMask() != threedo.PadStart {
		t.Fatalf("Enter mask = 0x%08X, want PadStart 0x%08X", a.padMask(), uint32(threedo.PadStart))
	}
	// Holding it (a repeat with no key event, then an unmapped key) must not enqueue more.
	a.Key(debug.Key{Name: "Shift", Down: true}) // unmapped → ignored
	if len(a.padQueue) != 1 {
		t.Fatalf("queue after unmapped key = %d, want 1", len(a.padQueue))
	}
	// Add a second button: the combined state is a new distinct entry.
	a.Key(debug.Key{Name: "a", Down: true})
	if got, want := a.padQueue[len(a.padQueue)-1], uint32(threedo.PadStart|threedo.PadA); got != want {
		t.Fatalf("Start+A queued = 0x%08X, want 0x%08X", got, want)
	}
	// Release Enter (A still held) then A: each distinct transition is a queue entry.
	a.Key(debug.Key{Name: "Enter", Down: false}) // now A only
	a.Key(debug.Key{Name: "a", Down: false})     // now released
	if a.padMask() != 0 {
		t.Fatalf("mask after release = 0x%08X, want 0", a.padMask())
	}

	// releasePad delivers every distinct state in order, one per call.
	want := []uint32{
		threedo.PadStart,
		threedo.PadStart | threedo.PadA,
		threedo.PadA,
		0,
	}
	for i, w := range want {
		a.releasePad()
		if a.padLast != w {
			t.Fatalf("release %d: padLast = 0x%08X, want 0x%08X", i, a.padLast, w)
		}
	}
	if len(a.padQueue) != 0 {
		t.Fatalf("queue not drained: %d left", len(a.padQueue))
	}
	// An empty queue delivers nothing and leaves padLast unchanged (guest holds it).
	a.releasePad()
	if a.padLast != 0 {
		t.Fatalf("release on empty queue changed padLast to 0x%08X", a.padLast)
	}
}

// TestNativeMovieFlag checks that the NativeMovie option flips the machine's movie
// path: default keeps the out-of-band HLE decoder (MovieHLE), and setting it hands
// the guest its own player (MovieHLE off, streams enabled). Checked through build()
// so it needs no disc.
func TestNativeMovieFlag(t *testing.T) {
	if _, err := os.Stat(disc); err != nil {
		t.Skip("Need for Speed disc not present; skipping")
	}
	def, err := New(disc, opts) // default: NativeMovie off
	if err != nil {
		t.Fatal(err)
	}
	defer def.Close()
	if m := def.build(); !m.MovieHLE || m.NoStreams {
		t.Fatalf("default: MovieHLE=%v NoStreams=%v, want true/false", m.MovieHLE, m.NoStreams)
	}

	nativeOpts := opts
	nativeOpts.NativeMovie = true
	nat, err := New(disc, nativeOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer nat.Close()
	if m := nat.build(); m.MovieHLE || m.NoStreams {
		t.Fatalf("native: MovieHLE=%v NoStreams=%v, want false/false", m.MovieHLE, m.NoStreams)
	}
}
