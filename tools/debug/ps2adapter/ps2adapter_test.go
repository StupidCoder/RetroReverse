package ps2adapter

import (
	"os"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/ps2"
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
	a, err := New(disc, "") // Jak's IOPRP image is self-contained; no BIOS needed
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
		debug.CapResume, debug.CapRegions, debug.CapHalt, debug.CapProfile,
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

// TestFrameProfilePopulates: after a stepped title field the profiler reports where the
// field's time went, mapped through debug.FrameProfile — the data the profile panel draws.
// It exercises the whole chain the page's socket path does: the machine times its buckets,
// the adapter maps them, and the field that drew reports Drew and a positive total.
func TestFrameProfilePopulates(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping")
	}
	// Step until a field draws (the title composites over more than one field), so the
	// profile reported is a drawing field with real bucket time, not an idle one.
	var drew bool
	for i := 0; i < 8 && !drew; i++ {
		if _, err := a.StepFrame(false); err != nil {
			t.Fatal(err)
		}
		drew = a.FrameProfile().Drew
	}
	p := a.FrameProfile()
	if p.TotalMs <= 0 {
		t.Fatalf("profile total is %.3f ms after stepping fields; the profiler was not armed", p.TotalMs)
	}
	if len(p.Buckets) == 0 {
		t.Fatal("profile has no buckets")
	}
	if len(p.Counters) == 0 {
		t.Fatal("profile has no counters")
	}
	// The buckets partition the field: they add up to the total.
	var sum float64
	for _, b := range p.Buckets {
		sum += b.Millis
	}
	if d := sum - p.TotalMs; d < -0.5 || d > 0.5 {
		t.Errorf("buckets sum to %.2f ms, total is %.2f ms", sum, p.TotalMs)
	}
	// A drawing field names a rasterise bucket with primitives, and a fragments counter — the
	// two numbers the coming parallelisation is measured by.
	if !drew {
		t.Skip("no drawing field within 8 steps; the profile is honest but has nothing to assert on")
	}
	var haveRaster, haveFrags bool
	for _, b := range p.Buckets {
		if b.Name == "rasterise" && b.Count > 0 {
			haveRaster = true
		}
	}
	for _, c := range p.Counters {
		if c.Name == "fragments drawn" && c.Value > 0 {
			haveFrags = true
		}
	}
	if !haveRaster {
		t.Error("a drawing field reported no rasterise primitives")
	}
	if !haveFrags {
		t.Error("a drawing field reported no fragments drawn")
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

// TestDeinterlaceDisplayMatchesCapture: the de-interlaced Display and the StepFrame capture
// must report the SAME size, or the page's provenance overlay (keyed by the capture's
// width/height) would be laid over a differently-sized picture and every click would miss.
// This is the invariant the woven-frame change rests on.
func TestDeinterlaceDisplayMatchesCapture(t *testing.T) {
	a, haveState := atTitle(t)
	if !haveState {
		t.Skip("no title savestate; skipping")
	}
	if n, d := a.DisplayAspect(); n != 4 || d != 3 {
		t.Errorf("DisplayAspect = %d:%d, want 4:3", n, d)
	}
	var fc *debug.FrameCapture
	for i := 0; i < 8; i++ {
		var err error
		if fc, err = a.StepFrame(false); err != nil {
			t.Fatal(err)
		}
		if fc.Drawn() {
			break
		}
	}
	if !fc.Drawn() {
		t.Fatal("no drawn title frame within 8 fields")
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display after a drawn frame: %v", err)
	}
	if img.Bounds().Dx() != fc.Width || img.Bounds().Dy() != fc.Height {
		t.Errorf("Display is %dx%d but the capture is %dx%d — provenance would be misaligned",
			img.Bounds().Dx(), img.Bounds().Dy(), fc.Width, fc.Height)
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

// promptState is the memory-card prompt: the screen a person actually presses X on, and
// the only place in this boot where one button visibly changes what the machine does.
const promptState = "../../../games/jak-and-daxter-ps2/work/states/fresh800.state"

// TestKeyMapping pins the keys to the buttons. The face buttons are the pad's diamond
// spelled on the keyboard, so getting one wrong swaps two buttons that sit next to each
// other — which looks like a game bug, not a debugger bug, from every angle.
func TestKeyMapping(t *testing.T) {
	for _, c := range []struct{ key, want string }{
		{"w", "triangle"}, {"a", "square"}, {"s", "circle"}, {"z", "cross"},
		{"W", "triangle"}, {"Z", "cross"},
		{"Enter", "start"},
		{"ArrowUp", "stickup"}, {"ArrowDown", "stickdown"},
		{"ArrowLeft", "stickleft"}, {"ArrowRight", "stickright"},
		{"8", "up"}, {"2", "down"}, {"4", "left"}, {"6", "right"},
	} {
		if got := normalizeKey(c.key); got != c.want {
			t.Errorf("normalizeKey(%q) = %q, want %q", c.key, got, c.want)
		}
	}
	// The four face keys must reach four DIFFERENT bits: a table that maps two of them
	// to the same button still passes every "is it a button" check.
	seen := map[uint16]string{}
	for _, k := range []string{"w", "a", "s", "z"} {
		b, ok := ps2.PadButton(normalizeKey(k))
		if !ok {
			t.Fatalf("%q does not resolve to a button", k)
		}
		if prev, dup := seen[b]; dup {
			t.Errorf("%q and %q both map to bit %#04x", prev, k, b)
		}
		seen[b] = k
	}
}

// TestStickGateIsACircle pins the diagonal. Two axes driven independently would put the
// stick 1.41x further from centre than the shell allows — a position no real pad can
// report, which the game is entitled to be confused by.
func TestStickGateIsACircle(t *testing.T) {
	if x, y := stickFrom(map[string]bool{"stickright": true}); x != ps2.PadStickCentre+ps2.PadStickFull || y != ps2.PadStickCentre {
		t.Errorf("right alone = (%#x,%#x), want (%#x,%#x)", x, y, ps2.PadStickCentre+ps2.PadStickFull, ps2.PadStickCentre)
	}
	// Up must DECREASE y: the DualShock reports 0x00 at the top of its travel. A stick
	// that is upside down still moves, which is why this is asserted and not eyeballed.
	if _, y := stickFrom(map[string]bool{"stickup": true}); y >= ps2.PadStickCentre {
		t.Errorf("up gave y=%#x, want less than centre %#x — the stick is upside down", y, ps2.PadStickCentre)
	}
	dx, dy := stickFrom(map[string]bool{"stickup": true, "stickright": true})
	ddx, ddy := int(dx)-ps2.PadStickCentre, int(dy)-ps2.PadStickCentre
	r2 := ddx*ddx + ddy*ddy
	full2 := ps2.PadStickFull * ps2.PadStickFull
	if lo, hi := full2*9/10, full2*11/10; r2 < lo || r2 > hi {
		t.Errorf("diagonal is %d from centre squared, want ~%d (the gate is a circle)", r2, full2)
	}
	// Opposite keys cancel, as a physical stick does.
	if x, y := stickFrom(map[string]bool{"stickleft": true, "stickright": true}); x != ps2.PadStickCentre || y != ps2.PadStickCentre {
		t.Errorf("left+right = (%#x,%#x), want centred", x, y)
	}
}

// TestPadStateQueuesWholeAndDedups pins the two things the queue is for.
func TestPadStateQueuesWholeAndDedups(t *testing.T) {
	a := &Adapter{held: map[string]bool{}}
	a.Key(debug.Key{Name: "z", Down: true})
	if len(a.padQueue) != 1 || a.padQueue[0].buttons == 0 {
		t.Fatalf("press queued %v", a.padQueue)
	}
	// A repeat of the same level must not cost a field.
	a.Key(debug.Key{Name: "z", Down: true})
	if len(a.padQueue) != 1 {
		t.Errorf("a repeat queued a second state: %v", a.padQueue)
	}
	// A stick move while a button is held composes into ONE state, not two.
	a.Key(debug.Key{Name: "ArrowRight", Down: true})
	last := a.padQueue[len(a.padQueue)-1]
	cross, _ := ps2.PadButton("cross")
	if last.buttons&cross == 0 || last.stickX == ps2.PadStickCentre {
		t.Errorf("holding cross while pushing the stick gave %+v, want both in one state", last)
	}
	// An unmapped key is ignored, not an error: the browser sends everything.
	if err := a.Key(debug.Key{Name: "F5", Down: true}); err != nil {
		t.Errorf("unmapped key errored: %v", err)
	}
}

// TestKeyDismissesTheMemoryCardPrompt is the end-to-end one: a key pressed on the page
// reaches the game and the game acts on it.
//
// It is what proves the PACING, which is the whole difficulty of a level-sampled pad. The
// prompt is the honest target because the control is free — this screen sits there
// indefinitely until X is pressed, so "the text went away" cannot be the machine moving on
// by itself. Both halves are asserted: press and it goes, and (TestNoKeyLeavesThePrompt)
// do not press and it stays.
func TestKeyDismissesTheMemoryCardPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a field at a time; -short skips")
	}
	a := open(t)
	if _, err := os.Stat(promptState); err != nil {
		t.Skip("no memory-card prompt savestate; skipping")
	}
	if err := a.LoadStateFile(promptState); err != nil {
		t.Fatal(err)
	}
	before := drawsText(t, a)
	if !before {
		t.Fatal("the prompt state does not draw text; wrong state or the font regressed")
	}
	a.Key(debug.Key{Name: "z", Down: true}) // cross
	a.Key(debug.Key{Name: "z", Down: false})
	for i := 0; i < 240; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	if drawsText(t, a) {
		t.Error("the prompt is still up after pressing cross: the press never reached the game")
	}
}

// TestNoKeyLeavesThePrompt is the control for the test above. Without it, a prompt that
// timed out on its own would read as a press that worked.
func TestNoKeyLeavesThePrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a field at a time; -short skips")
	}
	a := open(t)
	if _, err := os.Stat(promptState); err != nil {
		t.Skip("no memory-card prompt savestate; skipping")
	}
	if err := a.LoadStateFile(promptState); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 240; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatal(err)
		}
	}
	if !drawsText(t, a) {
		t.Error("the prompt went away with no press: the dismissal test proves nothing")
	}
}

// drawsText reports whether the machine is drawing the prompt, by asking whether
// draw-string's glyph transform runs — the prompt redraws its text every frame, and the
// flyby behind it does not draw any until much later. Reading the engine rather than the
// picture keeps this independent of what the title's colours happen to be doing.
func drawsText(t *testing.T, a *Adapter) bool {
	t.Helper()
	const drawStringFTOI = 0x00703570 // draw-string+0x8FC, the per-glyph vftoi4
	hit := false
	m := a.Machine()
	prev := m.OnStep
	m.OnStep = func(mm *ps2.Machine, pc uint32) {
		if pc == drawStringFTOI {
			hit = true
		}
	}
	for i := 0; i < 4 && !hit; i++ {
		a.StepFast()
	}
	m.OnStep = prev
	return hit
}
