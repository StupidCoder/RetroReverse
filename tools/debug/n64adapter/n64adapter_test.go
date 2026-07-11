package n64adapter_test

import (
	"image"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n64adapter"
)

// romPath is the Pilotwings 64 image, relative to this package (three levels up
// to the repo root, same as the n64 package's own tests). Game images are not
// committed, so the tests skip when it is absent.
const romPath = "../../../games/pilotwings-64-n64/image/Pilotwings 64 (USA).z64"

func newAdapter(t *testing.T) *n64adapter.Adapter {
	t.Helper()
	if _, err := os.Stat(romPath); os.IsNotExist(err) {
		t.Skip("Pilotwings 64 image not present (game images are not committed)")
	}
	a, err := n64adapter.New(romPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// mustContent advances a field at a time until a frame draws a real scene (many
// RDP commands with provenance), or skips the test.
func mustContent(t *testing.T, a *n64adapter.Adapter, withOverdraw bool) *debug.FrameCapture {
	t.Helper()
	for i := 0; i < 400; i++ {
		fc, err := a.StepFrame(withOverdraw)
		if err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc
		}
	}
	t.Skip("never reached a drawn frame within the field budget")
	return nil
}

func TestAdapterCapturesAndReplaysAFrame(t *testing.T) {
	a := newAdapter(t)
	fc := mustContent(t, a, true)

	// Command stream is contiguous and named.
	for i, c := range fc.Commands {
		if c.Index != i {
			t.Fatalf("command %d has Index %d", i, c.Index)
		}
		if c.Name == "" {
			t.Errorf("command %d (op %#x) has no name", i, c.Op)
		}
	}

	// Provenance is laid out to the draw target and references real commands.
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("Prov has %d entries, want Width*Height = %d", len(fc.Prov), fc.Width*fc.Height)
	}
	px, py, pc := firstDrawnPixel(fc)
	if pc < 0 {
		t.Fatal("no pixel had provenance")
	}
	if pc >= len(fc.Commands) {
		t.Fatalf("pixel (%d,%d) points at command %d, out of %d", px, py, pc, len(fc.Commands))
	}
	if len(fc.Overdraw) == 0 {
		t.Error("withOverdraw was set but no overdraw history was recorded")
	}

	// The command a pixel is attributed to is the one that set its final colour:
	// rendering up to that command must change the pixel (provenance is the last
	// writer, so it does not change again afterwards).
	if pc >= 1 {
		before, err := a.RenderAfter(fc, pc-1)
		if err != nil {
			t.Fatalf("RenderAfter(%d): %v", pc-1, err)
		}
		after, err := a.RenderAfter(fc, pc)
		if err != nil {
			t.Fatalf("RenderAfter(%d): %v", pc, err)
		}
		if sameSize(before, after) && colorAt(before, px, py) == colorAt(after, px, py) {
			t.Errorf("pixel (%d,%d) attributed to command %d did not change between "+
				"RenderAfter(%d) and RenderAfter(%d)", px, py, pc, pc-1, pc)
		}
	}

	// The frame is built progressively: an early point in the command stream and
	// the finished frame differ, and the finished frame is a real scene (a cleared
	// screen holds a handful of colours; a rendered one holds many). Note distinct
	// colour count is NOT monotonic in k — early replays show the previous frame's
	// buffer being cleared and overdrawn — so we assert difference, not "richer".
	early, err := a.RenderAfter(fc, len(fc.Commands)/8)
	if err != nil {
		t.Fatalf("RenderAfter early: %v", err)
	}
	full, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter last: %v", err)
	}
	if equalPix(early, full) {
		t.Error("the frame at 1/8 of its commands is identical to the finished frame")
	}
	if df := distinctColors(full); df < 50 {
		t.Errorf("the finished frame holds only %d colours: nothing was rendered", df)
	} else {
		t.Logf("frame: %d commands, %dx%d; finished frame holds %d colours; "+
			"first drawn pixel (%d,%d) <- command %d %q",
			len(fc.Commands), fc.Width, fc.Height, df, px, py, pc, fc.Commands[pc].Name)
	}
}

func TestAdapterReplayIsDeterministic(t *testing.T) {
	a := newAdapter(t)
	fc := mustContent(t, a, false)
	k := len(fc.Commands) / 2

	first, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	second, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	if !equalPix(first, second) {
		t.Error("RenderAfter is not deterministic for the same command")
	}
}

func TestAdapterSnapshotRestore(t *testing.T) {
	a := newAdapter(t)
	mustContent(t, a, false) // advance into the attract sequence

	s := a.Snapshot()
	before, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if err := a.Restore(s); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	after, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(before.Commands) != len(after.Commands) {
		t.Errorf("frame after a restore differs: %d commands vs %d",
			len(after.Commands), len(before.Commands))
	}
}

// --- helpers ---------------------------------------------------------------

// firstDrawnPixel scans provenance in raster order for the first pixel a command
// wrote, returning its coordinates and command index (pc = -1 if none).
func firstDrawnPixel(fc *debug.FrameCapture) (x, y, pc int) {
	for yy := 0; yy < fc.Height; yy++ {
		for xx := 0; xx < fc.Width; xx++ {
			if c := fc.ProvAt(xx, yy); c >= 0 {
				return xx, yy, c
			}
		}
	}
	return 0, 0, -1
}

func sameSize(a, b *image.RGBA) bool { return a.Bounds() == b.Bounds() }

func colorAt(img *image.RGBA, x, y int) uint32 {
	i := img.PixOffset(x, y)
	return uint32(img.Pix[i])<<24 | uint32(img.Pix[i+1])<<16 | uint32(img.Pix[i+2])<<8 | uint32(img.Pix[i+3])
}

func distinctColors(img *image.RGBA) int {
	seen := map[uint32]struct{}{}
	for i := 0; i+3 < len(img.Pix); i += 4 {
		c := uint32(img.Pix[i])<<24 | uint32(img.Pix[i+1])<<16 | uint32(img.Pix[i+2])<<8 | uint32(img.Pix[i+3])
		seen[c] = struct{}{}
	}
	return len(seen)
}

func equalPix(a, b *image.RGBA) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	if len(a.Pix) != len(b.Pix) {
		return false
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}
