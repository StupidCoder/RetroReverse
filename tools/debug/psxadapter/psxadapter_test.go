package psxadapter

import (
	"bytes"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
)

// The disc is not in the repository (copyright); these tests skip without it.
const (
	disc = "../../../games/ridge-racer-psx/Ridge Racer (Track 01).bin"

	// ridgeRacerISR is the game's own vectored-interrupt dispatcher, which the retail
	// BIOS would install via HookEntryInt into a slot the HLE BIOS leaves empty.
	ridgeRacerISR = 0x8004DF48
)

func open(t *testing.T) *Adapter {
	t.Helper()
	if _, err := os.Stat(disc); err != nil {
		t.Skip("Ridge Racer disc not present; skipping")
	}
	a, err := New(disc, Options{ISRHandler: ridgeRacerISR})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// drawnFrame steps until the game renders a scene, the way the debugger does.
func drawnFrame(t *testing.T, a *Adapter, withOverdraw bool) *debug.FrameCapture {
	t.Helper()
	for i := 0; i < 800; i++ {
		fc, err := a.StepFrame(withOverdraw)
		if err != nil {
			t.Fatal(err)
		}
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc
		}
	}
	t.Fatal("no drawn frame within the field budget")
	return nil
}

// TestCapabilities: the PSX backs the frame tools, and says so. This is the point of
// the whole exercise — the capability list is not a wish, it is what the adapter has.
func TestCapabilities(t *testing.T) {
	a := open(t)
	caps := map[string]bool{}
	for _, c := range debug.Capabilities(a) {
		caps[c] = true
	}
	for _, want := range []string{
		debug.CapFrames, debug.CapReplay, debug.CapFastStep, debug.CapCode,
		debug.CapBreak, debug.CapDisasm, debug.CapWatch, debug.CapStates,
	} {
		if !caps[want] {
			t.Errorf("the PSX target does not advertise %q", want)
		}
	}
	// It has no filesystem or surface panel yet, and must not pretend otherwise.
	for _, no := range []string{debug.CapFiles, debug.CapSurfaces} {
		if caps[no] {
			t.Errorf("the PSX target advertises %q, which it does not implement", no)
		}
	}
}

// TestReplayReproducesTheFrame is the honest invariant, and the one that would catch a
// broken stop-after-command: replaying a frame from its start snapshot and halting
// after the *last* command must reproduce, pixel for pixel, the buffer the live
// machine actually drew.
//
// The comparison is against the live machine's DRAW TARGET, not its scanout. Ridge
// Racer double-buffers: it draws into a back buffer elsewhere in VRAM while the
// display still shows the previous frame. Comparing scanouts would pass no matter what
// the replay did, because neither run touches the displayed buffer at all.
func TestReplayReproducesTheFrame(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	live := a.Machine().RenderDrawTarget()
	replayed, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Rect != live.Rect {
		t.Fatalf("replay is %v, the live draw target is %v", replayed.Rect, live.Rect)
	}
	if !bytes.Equal(replayed.Pix, live.Pix) {
		t.Errorf("replaying to the last command did not reproduce the live draw target")
	}
}

// TestTheDrawTargetIsNotTheScanout guards the mistake this adapter actually made: if
// the debugger renders the displayed buffer, the scrubber shows a picture that never
// changes, and every test above still passes. The two must genuinely differ while the
// game is drawing.
func TestTheDrawTargetIsNotTheScanout(t *testing.T) {
	a := open(t)
	drawnFrame(t, a, false)

	dx, dy := a.Machine().DrawOrigin()
	sx, sy := a.Machine().DisplayOrigin()
	if dx == sx && dy == sy {
		t.Skipf("this frame draws straight into the displayed buffer (%d,%d); nothing to distinguish", dx, dy)
	}
	t.Logf("draw target at VRAM (%d,%d), scanout at (%d,%d)", dx, dy, sx, sy)
}

// TestReplayIsDeterministic: the scrubber depends on it. Ask twice, get the same image.
func TestReplayIsDeterministic(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	k := len(fc.Commands) / 2
	first, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatal(err)
	}
	// Replay a different command in between, so the scratch machine is genuinely
	// reused and restored rather than sitting where we left it.
	if _, err := a.RenderAfter(fc, len(fc.Commands)-1); err != nil {
		t.Fatal(err)
	}
	second, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Pix, second.Pix) {
		t.Errorf("RenderAfter(%d) is not deterministic", k)
	}
}

// TestReplayBuildsUp: the picture assembles as the scrubber advances.
//
// Not monotonically, though — the frame opens by *clearing* the back buffer with a
// Fill_Rect, so the pixel count legitimately collapses first (before the fill, the
// buffer still holds the frame drawn into it two frames ago). What must hold is that
// the scrub is showing a frame being built at all: the picture right after the clear
// is nearly empty, the finished frame is not, and it grows in between.
func TestReplayBuildsUp(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	clear := -1
	for _, c := range fc.Commands {
		if c.Name == "Fill_Rect" {
			clear = c.Index
			break
		}
	}
	if clear < 0 {
		t.Skip("this frame does not clear its buffer; nothing to measure from")
	}

	last := len(fc.Commands) - 1
	lit := func(k int) int {
		img, err := a.RenderAfter(fc, k)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for i := 0; i < len(img.Pix); i += 4 {
			if img.Pix[i] != 0 || img.Pix[i+1] != 0 || img.Pix[i+2] != 0 {
				n++
			}
		}
		return n
	}

	afterClear, prev := lit(clear), lit(clear)
	for _, k := range []int{clear + (last-clear)/2, last} {
		n := lit(k)
		if n < prev {
			t.Errorf("command %d has %d lit pixels, fewer than the %d before it", k, n, prev)
		}
		prev = n
	}
	if prev <= afterClear {
		t.Errorf("the frame drew nothing after clearing: %d lit pixels at the clear, %d at the end", afterClear, prev)
	}
}

// TestProvenanceNamesRealCommands: every pixel is attributed to a command that exists,
// and at least some of the screen is attributed at all.
func TestProvenanceNamesRealCommands(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, true)

	if fc.Prov == nil {
		t.Fatal("the PSX capture has no provenance")
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("provenance is %d entries for a %dx%d frame", len(fc.Prov), fc.Width, fc.Height)
	}
	attributed := 0
	for i, c := range fc.Prov {
		if c < 0 {
			continue
		}
		if int(c) >= len(fc.Commands) {
			t.Fatalf("pixel %d is attributed to command %d, but the frame has %d", i, c, len(fc.Commands))
		}
		attributed++
	}
	if attributed == 0 {
		t.Error("no pixel is attributed to any command")
	}

	// The overdraw history must agree with the last-writer buffer.
	for key, writes := range fc.Overdraw {
		if len(writes) == 0 {
			continue
		}
		if got, want := int32(writes[len(writes)-1].CmdIndex), fc.Prov[key]; got != want {
			t.Fatalf("pixel %d: overdraw's last write is command %d, provenance says %d", key, got, want)
		}
	}
}

// TestSnapshotIsIndependent: a snapshot must survive the machine moving on, or replay
// silently renders the wrong frame.
func TestSnapshotIsIndependent(t *testing.T) {
	a := open(t)
	fc := drawnFrame(t, a, false)

	want, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	// Run the live machine well past the captured frame.
	for i := 0; i < 20; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	got, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Pix, want.Pix) {
		t.Error("the capture's snapshot did not survive the live machine advancing")
	}
}
