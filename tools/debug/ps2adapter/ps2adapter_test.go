package ps2adapter

import (
	"os"
	"testing"

	"retroreverse.com/tools/debug"
)

// The disc is not in the repository (copyright); these tests skip without it. A
// savestate saved into the game's states dir lets them start mid-title in a second
// instead of after a billion-instruction boot — the same handoff the debugger offers.
const (
	disc  = "../../../games/jak-and-daxter-ps2/image/Jak and Daxter - The Precursor Legacy.iso"
	state = "../../../games/jak-and-daxter-ps2/work/states/title-logo.state"
)

func open(t *testing.T) *Adapter {
	t.Helper()
	if _, err := os.Stat(disc); err != nil {
		t.Skip("Jak and Daxter disc not present; skipping")
	}
	a, err := New(disc)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// atTitle opens the adapter and jumps to the title savestate if it is present, so the
// render tests have a frame to look at without paying for the boot. Without the state
// they run from a cold boot, which has no drawn frame yet — so those assertions are
// gated on the state existing.
func atTitle(t *testing.T) (*Adapter, bool) {
	t.Helper()
	a := open(t)
	if _, err := os.Stat(state); err != nil {
		return a, false
	}
	if err := a.LoadStateFile(state); err != nil {
		t.Fatalf("loading %s: %v", state, err)
	}
	return a, true
}

// TestCapabilities: the PS2 backs the interactive half — play, surfaces, savestates,
// the CPU and memory tools — and says exactly that. It does NOT claim the frame
// scrubber, and the test asserts that absence too: a capability list is what the
// adapter has, not what it wishes for.
func TestCapabilities(t *testing.T) {
	a := open(t)
	caps := map[string]bool{}
	for _, c := range debug.Capabilities(a) {
		caps[c] = true
	}
	for _, want := range []string{
		debug.CapFrames, debug.CapFastStep, debug.CapCode, debug.CapBreak,
		debug.CapDisasm, debug.CapWatch, debug.CapSurfaces, debug.CapStates,
		debug.CapResume, debug.CapRegions, debug.CapHalt,
	} {
		if !caps[want] {
			t.Errorf("the PS2 target does not advertise %q", want)
		}
	}
	// The per-command scrubber (FrameReplayer) is deliberately absent: a GS kick draws
	// many primitives inside one EE instruction, so replaying to "after primitive k"
	// needs a mid-burst halt this pass does not build.
	if caps[debug.CapReplay] {
		t.Error("the PS2 target advertises replay, but it has no frame replayer yet")
	}
}

// TestStepFrameCapturesProvenance: a stepped title field draws primitives and attributes
// pixels to them — the "click a pixel, get the primitive" data the page runs on.
func TestStepFrameCapturesProvenance(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping")
	}
	var fc *debug.FrameCapture
	for i := 0; i < 8; i++ {
		var err error
		fc, err = a.StepFrame(false)
		if err != nil {
			t.Fatal(err)
		}
		if fc.Drawn() {
			break
		}
	}
	if !fc.Drawn() {
		t.Fatal("no drawn title frame within 8 fields")
	}
	if len(fc.Commands) == 0 {
		t.Error("the drawn frame captured no GS primitives")
	}
	if fc.Prov == nil {
		t.Fatal("the drawn frame captured no provenance")
	}
	// Some pixel names a primitive, and that primitive is in range.
	named := 0
	for _, c := range fc.Prov {
		if c >= 0 {
			if int(c) >= len(fc.Commands) {
				t.Fatalf("provenance names command %d, out of %d", c, len(fc.Commands))
			}
			named++
		}
	}
	if named == 0 {
		t.Error("no pixel is attributed to any primitive")
	}
	if len(fc.Writers) == 0 {
		t.Error("no primitive is recorded as having drawn a pixel")
	}
}

// TestPlatformAndTitle: the identity the page shows.
func TestPlatformAndTitle(t *testing.T) {
	a := open(t)
	if a.Platform() != "ps2" {
		t.Errorf("platform = %q, want ps2", a.Platform())
	}
	if a.Title() == "" {
		t.Error("title is empty")
	}
}

// TestPlayShowsAFrame: the core of what the user asked for — StepFast advances a field
// and Display returns the picture the GS is scanning out. From the title state a played
// field is a real, non-black composited frame.
func TestPlayShowsAFrame(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping the drawn-frame assertion")
	}
	for i := 0; i < 4; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if img.Rect.Dx() == 0 || img.Rect.Dy() == 0 {
		t.Fatal("Display returned an empty image")
	}
	nonBlack := 0
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 || img.Pix[i+1] != 0 || img.Pix[i+2] != 0 {
			nonBlack++
		}
	}
	if nonBlack == 0 {
		t.Errorf("the title frame is entirely black (%dx%d)", img.Rect.Dx(), img.Rect.Dy())
	}
}

// TestSnapshotRestore: a snapshot taken, the machine stepped on, then the snapshot put
// back must return the CPU to where it was — the property play/scrub depend on.
func TestSnapshotRestore(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping")
	}
	before := a.CPU().PC
	snap := a.Snapshot()
	if snap.Platform() != "ps2" {
		t.Fatalf("snapshot platform = %q", snap.Platform())
	}
	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.Restore(snap); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got := a.CPU().PC; got != before {
		t.Errorf("PC after restore = %08X, want %08X", got, before)
	}
}

// TestSurfaces: the scanout is a fixed surface with real dimensions once a frame exists,
// and the free GS-buffer surface reads a PSMCT32 rectangle by its word base — the
// display buffers (0xA0000) among them.
func TestSurfaces(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping")
	}
	if err := a.StepFast(); err != nil {
		t.Fatal(err)
	}
	surfs := a.Surfaces()
	if len(surfs) < 2 {
		t.Fatalf("expected at least scanout + buffer surfaces, got %d", len(surfs))
	}
	img, err := a.RenderSurface("scanout", debug.View{})
	if err != nil {
		t.Fatalf("scanout surface: %v", err)
	}
	if img.Rect.Dx() == 0 {
		t.Error("scanout surface is empty")
	}
	// A display buffer, as a free PSMCT32 rectangle.
	buf, err := a.RenderSurface("buffer", debug.View{Addr: 0xA0000, W: 512, H: 224})
	if err != nil {
		t.Fatalf("buffer surface: %v", err)
	}
	if buf.Rect.Dx() != 512 {
		t.Errorf("buffer width = %d, want 512", buf.Rect.Dx())
	}
	if _, err := a.RenderSurface("nope", debug.View{}); err == nil {
		t.Error("an unknown surface id should error")
	}
}

// TestCodeAndDisasm: the CPU pane's needs — a register file, an instruction decode, a
// bounded step.
func TestCodeAndDisasm(t *testing.T) {
	a, _ := atTitle(t)
	regs := a.CPU()
	if len(regs.Names) != 32 || len(regs.Vals) != 32 {
		t.Fatalf("register file has %d names / %d vals, want 32/32", len(regs.Names), len(regs.Vals))
	}
	ins, err := a.Disasm(regs.PC, 4)
	if err != nil {
		t.Fatalf("Disasm: %v", err)
	}
	if len(ins) != 4 || ins[0].Text == "" {
		t.Fatalf("disasm returned %d instrs, first %q", len(ins), text(ins))
	}
	sr, err := a.StepInstr(16)
	if err != nil {
		t.Fatalf("StepInstr: %v", err)
	}
	if sr.Steps == 0 {
		t.Error("StepInstr ran 0 instructions")
	}
}

// TestWatchOneWindowPerKind: the machine has one write window and one read window, so a
// second watch of the same kind is refused rather than silently dropped.
func TestWatchOneWindowPerKind(t *testing.T) {
	a := open(t)
	if _, err := a.SetWatch(debug.Watch{Kind: "write", Lo: 0x100000, Hi: 0x100010}); err != nil {
		t.Fatalf("first write watch: %v", err)
	}
	if _, err := a.SetWatch(debug.Watch{Kind: "write", Lo: 0x200000, Hi: 0x200010}); err == nil {
		t.Error("a second write watch should be refused")
	}
	if _, err := a.SetWatch(debug.Watch{Kind: "read", Lo: 0x100000, Hi: 0x100010}); err != nil {
		t.Errorf("a read watch alongside a write watch should be allowed: %v", err)
	}
}

// TestResumeArgs: the resume line uses the PS2 oracle's own flags, so a state saved in
// the browser reopens in bootoracle.
func TestResumeArgs(t *testing.T) {
	a := open(t)
	args := a.ResumeArgs("/tmp/x.state")
	joined := ""
	for _, s := range args {
		joined += s + " "
	}
	for _, want := range []string{"-image", "-loadstate", "/tmp/x.state"} {
		if !contains(joined, want) {
			t.Errorf("resume args %v miss %q", args, want)
		}
	}
}

func text(ins []debug.Instr) string {
	if len(ins) == 0 {
		return ""
	}
	return ins[0].Text
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
