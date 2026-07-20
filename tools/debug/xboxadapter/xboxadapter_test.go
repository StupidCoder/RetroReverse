package xboxadapter_test

import (
	"image"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/xboxadapter"
	"retroreverse.com/tools/platform/xbox"
)

// The disc and a savestate that is already at a drawn frame. Game images are not committed
// and work/ is gitignored scratch, so both are required and the tests skip without them.
//
// The state is the title screen — the movie background, the UI sprite layer and the
// blinking PRESS START. Regenerate it (from games/outrun-2006-xbox/extract, ~20 minutes:
// the boot runs the HDD install and decodes a 27.9 MB XMV movie on the interpreted CPU)
// with:
//
//	go run ./cmd/bootoracle -image "../image/OutRun 2006 - Coast 2 Coast (EUR).iso" \
//	    -gpu -steps 2700000000 -savestate ../work/title.state
//
// For a faster fixture the SEGA logo card is ~20 seconds away and drives the same
// pipeline: the same command with -steps 340000000.
const (
	discPath  = "../../../games/outrun-2006-xbox/image/OutRun 2006 - Coast 2 Coast (EUR).iso"
	statePath = "../../../games/outrun-2006-xbox/work/title.state"
	xbePath   = "/default.xbe"
)

func newAdapter(t *testing.T) *xboxadapter.Adapter {
	t.Helper()
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the OutRun 2006 disc is not present (game images are not committed)")
	}
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Skip("the title savestate is not present (work/ is regenerable scratch)")
	}
	a, err := xboxadapter.New(discPath, xbePath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	if err := a.LoadStateFile(statePath); err != nil {
		t.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

// TestCapabilities pins both halves of the contract: what this target backs, and what it
// must NOT claim. CapKeys moved from the second list to the first when the machine grew a
// pad it could honestly deliver a button to — a capability is a promise, and the promise
// is what puts the input panel on screen.
func TestCapabilities(t *testing.T) {
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the OutRun 2006 disc is not present")
	}
	a, err := xboxadapter.New(discPath, xbePath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()

	caps := map[string]bool{}
	for _, c := range debug.Capabilities(a) {
		caps[c] = true
	}
	for _, want := range []string{
		debug.CapFrames, debug.CapFastStep, debug.CapReplay, debug.CapCode,
		debug.CapBreak, debug.CapDisasm, debug.CapWatch, debug.CapSurfaces,
		debug.CapFiles, debug.CapFileAt, debug.CapStates, debug.CapResume,
		debug.CapRegions, debug.CapHalt, debug.CapKeys, debug.CapProfile,
	} {
		if !caps[want] {
			t.Errorf("capability %q is missing", want)
		}
	}
	// CapProfile moved to the wanted list when the machine grew a per-subsystem frame
	// profiler it could honestly back (xbox/profile.go) — the promise that puts the profile
	// panel on screen.
	for _, unwanted := range []string{debug.CapTouch} {
		if caps[unwanted] {
			t.Errorf("capability %q is claimed but not backed", unwanted)
		}
	}
}

// TestProfileReachesTheDebugger: the adapter must hand over a stepped frame's real profile,
// with the rasteriser named as the cost centre and the work counters that let the panel show
// fragments/ms — the whole reason the profiler exists is to guide the performance work.
func TestProfileReachesTheDebugger(t *testing.T) {
	a := newAdapter(t)
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	p := a.FrameProfile()
	if p.TotalMs == 0 || len(p.Buckets) == 0 {
		t.Fatalf("the adapter reported no profile: %+v", p)
	}
	var sum float64
	for _, b := range p.Buckets {
		sum += b.Millis
	}
	if d := sum - p.TotalMs; d > 0.5 || d < -0.5 {
		t.Errorf("buckets sum to %.3f ms but total is %.3f ms — not disjoint", sum, p.TotalMs)
	}
	// The rasteriser is this machine's cost centre; a profile that does not say so is
	// measuring the wrong thing.
	var raster float64
	for _, b := range p.Buckets {
		if b.Name == "rasterise" {
			raster = b.Millis
		}
	}
	if raster <= 0 {
		t.Errorf("the profile does not name the rasteriser as a cost: %+v", p.Buckets)
	}
	// The panel's fragments/ms line reads these exact counter names.
	need := map[string]bool{"fragments drawn": false, "depth-killed": false, "draws": false}
	for _, c := range p.Counters {
		if _, ok := need[c.Name]; ok {
			need[c.Name] = true
		}
	}
	for name, got := range need {
		if !got {
			t.Errorf("the profile is missing counter %q (the panel's rate line needs it)", name)
		}
	}
}

// TestCaptureAndReplay is the end-to-end frame: a capture must hold a real command stream,
// contiguous indices, named commands, sized provenance — and replaying to the last command
// must reproduce a frame with a picture in it.
func TestCaptureAndReplay(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(fc.Commands) < 100 {
		t.Fatalf("a title-screen frame drew only %d commands; expected a real stream", len(fc.Commands))
	}
	for i, c := range fc.Commands {
		if c.Index != i {
			t.Fatalf("command %d reports index %d", i, c.Index)
		}
		if c.Name == "" || c.Name == "UNKNOWN" {
			t.Errorf("command %d (method %04X) is unnamed", i, c.Op)
		}
	}
	if fc.Width <= 0 || fc.Height <= 0 {
		t.Fatalf("capture is %dx%d", fc.Width, fc.Height)
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("Prov is %d entries for a %dx%d frame", len(fc.Prov), fc.Width, fc.Height)
	}
	for i, c := range fc.Prov {
		if c >= int32(len(fc.Commands)) {
			t.Fatalf("Prov[%d] names command %d of %d", i, c, len(fc.Commands))
		}
	}
	// Writers is the point of a register-write machine's scrubber: the frame is tens of
	// thousands of methods and only a few of them draw.
	if len(fc.Writers) == 0 {
		t.Error("no command stored a pixel; the scrubber would have nothing to tour")
	}
	if len(fc.Writers) >= len(fc.Commands) {
		t.Errorf("%d of %d commands wrote pixels; on the NV2A most methods are state-setting",
			len(fc.Writers), len(fc.Commands))
	}

	img, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	if nonBlack(img) < 10 {
		t.Error("replaying the whole frame produced an (almost) black picture")
	}
}

// TestCaptureIsAWholeFrameNotAnOffScreenPass is the regression for the frame boundary, and
// it is the test the obvious boundary passes at the logo card and fails here.
//
// OutRun's title screen renders its movie into an OFF-SCREEN target and only then draws the
// frame, so a capture bounded by the swap-chain buffer moving on (SET_SURFACE_COLOR_OFFSET)
// stops three times per frame — the first time on a buffer nothing has drawn into. What it
// reports is a clean, sensibly-sized, entirely blank frame: no error, no crash, nothing to
// notice. So this asserts what a whole frame must contain — geometry that stored pixels —
// and that the picture is not one flat colour.
func TestCaptureIsAWholeFrameNotAnOffScreenPass(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	draws := 0
	for _, w := range fc.Writers {
		if fc.Commands[w].Name == "SET_BEGIN_END" {
			draws++
		}
	}
	if draws == 0 {
		t.Errorf("the frame stored pixels from %d commands but none of them was geometry: "+
			"this is a clear-only pass, not a presented frame", len(fc.Writers))
	}
	img, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	if uniform(img) {
		t.Error("the finished frame is a single flat colour — the capture ended on a buffer nothing drew into")
	}
}

// TestDisplayIsThePresentedFrameNotTheStaleScanout is the regression for the debugger's
// MAIN VIEW — the picture play mode streams, and the one a user stares at for thousands of
// frames.
//
// The CRTC's programmed scanout is a 320x240 window on a buffer nothing draws into (the
// title registers its display mode once, during loading, and the 640x480 switch it goes on
// to render at never happens in this HLE). Wiring Display() to it renders a blank white
// rectangle while the game draws perfectly — which is not a crash, not an error, and is
// indistinguishable from a broken emulator. Display() is the buffer the title PRESENTED.
func TestDisplayIsThePresentedFrameNotTheStaleScanout(t *testing.T) {
	a := newAdapter(t)
	if err := a.StepFast(); err != nil { // one flip, so something has been presented
		t.Fatalf("StepFast: %v", err)
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if uniform(img) {
		t.Error("Display() is a single flat colour — the main view shows a buffer nothing drew into")
	}
	// It must be the frame's geometry, not the stale loading mode's.
	surfs := map[string]debug.Surface{}
	for _, s := range a.Surfaces() {
		surfs[s.ID] = s
	}
	dt := surfs["drawtarget"]
	if img.Rect.Dx() != dt.W || img.Rect.Dy() != dt.H {
		t.Errorf("Display() is %dx%d but the frame is %dx%d — that is the stale scanout mode",
			img.Rect.Dx(), img.Rect.Dy(), dt.W, dt.H)
	}
	// And the scanout is still offered separately, because the gap it shows is real and
	// should stay visible rather than be quietly replaced by the picture that works.
	if _, ok := surfs["scanout"]; !ok {
		t.Error("the CRTC scanout is no longer offered as a surface; the frontier it shows is now invisible")
	}
}

// TestTheClearReportsItsPixels: the clear is a command that stores pixels, and most of a
// frame's background is its work. A clear that reported nothing would leave provenance
// saying "no command wrote this pixel" across the whole background — which is false, and
// which is exactly the answer to "why is this pixel this colour" for those pixels.
func TestTheClearReportsItsPixels(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	for _, w := range fc.Writers {
		if fc.Commands[w].Name == "CLEAR_SURFACE" {
			return
		}
	}
	t.Error("CLEAR_SURFACE stored no pixels; the frame's background has no provenance")
}

// TestReplayIsDeterministic is the property an aliased snapshot silently breaks: the same
// replay, twice, must be the same bytes.
func TestReplayIsDeterministic(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	k := len(fc.Commands) - 1
	first, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	second, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter (again): %v", err)
	}
	if len(first.Pix) != len(second.Pix) {
		t.Fatalf("replays differ in size: %d vs %d", len(first.Pix), len(second.Pix))
	}
	for i := range first.Pix {
		if first.Pix[i] != second.Pix[i] {
			t.Fatalf("replays differ at byte %d: %d vs %d", i, first.Pix[i], second.Pix[i])
		}
	}
}

// TestLiveMachineSurvivesAReplay: a replay runs on the scratch machine, and must not
// perturb the live one. This is what lets the scrubber be dragged while the game is paused
// mid-frame rather than re-stepped.
func TestLiveMachineSurvivesAReplay(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	before, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	pcBefore := a.CPU().PC
	golden := append([]uint8(nil), before.Pix...)

	if _, err := a.RenderAfter(fc, len(fc.Commands)/2); err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	after, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	for i := range golden {
		if golden[i] != after.Pix[i] {
			t.Fatalf("the live machine's picture changed at byte %d during a replay", i)
		}
	}
	if pc := a.CPU().PC; pc != pcBefore {
		t.Fatalf("the live machine's PC moved during a replay: %08X -> %08X", pcBefore, pc)
	}
}

// TestSnapshotIsIndependentOfTheMachine is the aliasing guard, and it is deliberately not
// "does it resume identically" — a snapshot that shares the machine's RAM slice passes
// every resume test there is and is still not a snapshot. So it mutates in both
// directions: running the machine on must not rewrite the snapshot's memory, and restoring
// must not hand the machine a slice the snapshot still owns.
func TestSnapshotIsIndependentOfTheMachine(t *testing.T) {
	a := newAdapter(t)
	m := a.Machine()

	// A byte the title's own code is using: its stack. If the snapshot aliases RAM, running
	// the machine on will rewrite it under the snapshot.
	const probe = 0x00100000
	snapshot := a.Snapshot()
	m.RAM[probe] = 0xAB
	if err := a.Restore(snapshot); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got := m.RAM[probe]; got == 0xAB {
		t.Fatal("a write after Snapshot survived Restore: the snapshot aliases the machine's RAM")
	}

	// And the other direction: mutate the machine AFTER a restore, then restore the same
	// snapshot again. A snapshot handed out by reference would have been overwritten.
	m.RAM[probe] = 0xCD
	if err := a.Restore(snapshot); err != nil {
		t.Fatalf("Restore (again): %v", err)
	}
	if got := m.RAM[probe]; got == 0xCD {
		t.Fatal("a snapshot could not be restored twice: Restore handed the machine the snapshot's own memory")
	}
}

// TestCaptureIsTheDrawTargetNotTheScanout: the two pictures are different buffers, and the
// capture must be the one the frame was built in.
func TestCaptureIsTheDrawTargetNotTheScanout(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	surfs := map[string]debug.Surface{}
	for _, s := range a.Surfaces() {
		surfs[s.ID] = s
	}
	dt, ok := surfs["drawtarget"]
	if !ok {
		t.Fatal("no drawtarget surface")
	}
	if fc.Width != dt.W || fc.Height != dt.H {
		t.Errorf("capture is %dx%d but the draw target is %dx%d — the capture is not the draw target",
			fc.Width, fc.Height, dt.W, dt.H)
	}
}

// TestTheDiscIsAFilesystem: browse it, and map an offset back to the file that holds it.
func TestTheDiscIsAFilesystem(t *testing.T) {
	a := newAdapter(t)
	root, err := a.ListDir("/")
	if err != nil {
		t.Fatalf("ListDir(/): %v", err)
	}
	if len(root) == 0 {
		t.Fatal("the disc root is empty")
	}
	// default.xbe is the one file every Xbox disc has.
	var probe debug.FileEntry
	for _, e := range root {
		if !e.Dir && e.Size > 0 {
			probe = e
			break
		}
	}
	if probe.Path == "" {
		t.Skip("no file in the disc root")
	}
	got, within, ok := a.FileAt(probe.Offset + 16)
	if !ok {
		t.Fatalf("FileAt(%d) found nothing, but it is 16 bytes into %s", probe.Offset+16, probe.Path)
	}
	if got.Path != probe.Path || within != 16 {
		t.Errorf("FileAt(%d) = %s+%d, want %s+16", probe.Offset+16, got.Path, within, probe.Path)
	}
	data, err := a.ReadFile(probe.Path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", probe.Path, err)
	}
	if int64(len(data)) != probe.Size {
		t.Errorf("ReadFile(%s) returned %d bytes, want %d", probe.Path, len(data), probe.Size)
	}
}

// TestSurfacesOffersNoPhantomTextures: an unbound texture unit's registers read back zero
// and decode as a 1x1 texture at address 0, so a list that offered every unit would be
// mostly phantoms.
func TestSurfacesOffersNoPhantomTextures(t *testing.T) {
	a := newAdapter(t)
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	sawFree := false
	for _, s := range a.Surfaces() {
		if s.ID == "ram" {
			sawFree = true
			if !s.Free {
				t.Error("the ram surface must be Free (the caller aims it)")
			}
			continue
		}
		if s.W <= 1 && s.H <= 1 {
			t.Errorf("surface %q is %dx%d — a phantom", s.ID, s.W, s.H)
		}
	}
	if !sawFree {
		t.Error("no free RAM surface offered")
	}
}

// TestBreakpointStopsBeforeItsInstruction pins the subtlety that a debugger's PC is the
// instruction about to run, not the one just retired: a breakpoint reported one
// instruction late would still look plausible in every panel.
func TestBreakpointStopsBeforeItsInstruction(t *testing.T) {
	a := newAdapter(t)

	// Learn a PC the machine really does reach, by single-stepping and watching. Guessing
	// "the next instruction in memory" would be a test of nothing: the very next one may be
	// a branch, and then the address never comes up at all.
	start := a.Snapshot()
	const steps = 8
	var target uint64
	for i := 0; i < steps; i++ {
		if _, err := a.StepInstr(1); err != nil {
			t.Fatalf("StepInstr: %v", err)
		}
	}
	target = a.CPU().PC

	// Rewind and run at it. The machine must stop PARKED on that instruction, not after it.
	if err := a.Restore(start); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	a.SetBreakpoint(target)
	defer a.ClearBreakpoint(target)

	sr, err := a.Continue(10000)
	if err != nil {
		t.Fatalf("Continue: %v", err)
	}
	if sr.Kind != "breakpoint" {
		t.Fatalf("stopped for %q (%s), want breakpoint at %08X", sr.Kind, sr.Note, target)
	}
	if sr.PC != target {
		t.Errorf("stopped at %08X, want %08X — the breakpoint fired at the wrong instruction",
			sr.PC, target)
	}
	if pc := a.CPU().PC; pc != target {
		t.Errorf("machine is parked at %08X, want %08X", pc, target)
	}
	// And the instruction must still be unexecuted: stepping once from here must run it and
	// land somewhere else. A breakpoint that fired one instruction late would still have
	// reported the right PC above and only shown itself here.
	if _, err := a.StepInstr(1); err != nil {
		t.Fatalf("StepInstr: %v", err)
	}
	if pc := a.CPU().PC; pc == target {
		t.Errorf("still parked at %08X after a step", target)
	}
}

// TestDisasmWalksVariableLength: x86 instructions are not fixed-width, so a walk that
// assumed a stride would desync into garbage after the first long instruction.
func TestDisasmWalksVariableLength(t *testing.T) {
	a := newAdapter(t)
	ins, err := a.Disasm(a.CPU().PC, 8)
	if err != nil {
		t.Fatalf("Disasm: %v", err)
	}
	if len(ins) == 0 {
		t.Fatal("Disasm returned nothing")
	}
	for i, in := range ins {
		if in.Text == "" {
			t.Errorf("instruction %d has no text", i)
		}
		if len(in.Bytes) == 0 {
			t.Errorf("instruction %d has no bytes", i)
		}
		if i > 0 {
			want := ins[i-1].Addr + uint64(len(ins[i-1].Bytes))
			if in.Addr != want {
				t.Errorf("instruction %d is at %08X, but %d ended at %08X", i, in.Addr, i-1, want)
			}
		}
	}
}

// TestResumeArgsNameTheOracleFlags: the resume line has to be one this game's oracle
// actually accepts — -gpu above all, without which the machine stops at the first push
// instead of rendering.
func TestResumeArgsNameTheOracleFlags(t *testing.T) {
	a := newAdapter(t)
	args := a.ResumeArgs("/tmp/x.state")
	joined := ""
	for _, s := range args {
		joined += s + " "
	}
	for _, want := range []string{"-loadstate", "/tmp/x.state", "-gpu", "-image"} {
		if !contains(joined, want) {
			t.Errorf("resume args %v are missing %q", args, want)
		}
	}
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// uniform reports whether every pixel is the same colour — a blank frame, whatever shade.
func uniform(img *image.RGBA) bool {
	if len(img.Pix) < 8 {
		return true
	}
	for i := 4; i+3 < len(img.Pix); i += 4 {
		if img.Pix[i] != img.Pix[0] || img.Pix[i+1] != img.Pix[1] || img.Pix[i+2] != img.Pix[2] {
			return false
		}
	}
	return true
}

// nonBlack counts pixels that are not pure black.
func nonBlack(img *image.RGBA) int {
	n := 0
	for i := 0; i+3 < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 || img.Pix[i+1] != 0 || img.Pix[i+2] != 0 {
			n++
		}
	}
	return n
}

// --- the pad ---

// TestUnmappedKeyIsIgnored: the browser sends every key the user presses, including ones
// that are not buttons. Erroring on them would make the panel's own noise look like a bug.
func TestUnmappedKeyIsIgnored(t *testing.T) {
	a := newAdapter(t)
	if err := a.Key(debug.Key{Name: "F7", Down: true}); err != nil {
		t.Errorf("Key(F7): %v", err)
	}
}

// TestPadStatesEachGetAFrame is the pathological case the queue exists for: a press and
// its release with no frame in between. The title edge-detects presses from consecutive
// polls, so a level that is never live across a poll is a press that never happened —
// both states must therefore reach the machine, on frames of their own.
func TestPadStatesEachGetAFrame(t *testing.T) {
	a := newAdapter(t)
	m := a.Machine()
	if m.OnFlip == nil {
		t.Fatal("the adapter installed no frame hook, so nothing is pacing the pad")
	}

	// Watch the level the pad is actually left holding at each frame boundary, chaining
	// after the adapter's own hook so we see what it set.
	var seen []uint16
	prev := m.OnFlip
	m.OnFlip = func(mm *xbox.Machine) {
		prev(mm)
		seen = append(seen, mm.PadButtons(0))
	}

	a.Key(debug.Key{Name: "Enter", Down: true})  // press start...
	a.Key(debug.Key{Name: "Enter", Down: false}) // ...and release it, with no frame between

	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatalf("StepFast: %v", err)
		}
	}

	startCtl, _ := xbox.PadControlByName("start")
	start := startCtl.Bit
	pressedAt := -1
	for i, mask := range seen {
		if mask&start != 0 {
			pressedAt = i
			break
		}
	}
	if pressedAt < 0 {
		t.Fatalf("no frame saw the press (levels seen: %v)", seen)
	}
	released := false
	for _, mask := range seen[pressedAt+1:] {
		if mask&start == 0 {
			released = true
			break
		}
	}
	if !released {
		t.Errorf("no later frame saw the release (levels seen: %v)", seen)
	}
}

// TestHeldPadButtonStaysHeld: the queue carries changes, so an empty queue must leave the
// level alone. A pacing loop that wrote the queue's head unconditionally would invent a
// release the moment the user stopped typing.
func TestHeldPadButtonStaysHeld(t *testing.T) {
	a := newAdapter(t)
	m := a.Machine()

	a.Key(debug.Key{Name: "Enter", Down: true})
	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatalf("StepFast: %v", err)
		}
	}

	startCtl, _ := xbox.PadControlByName("start")
	start := startCtl.Bit
	if got := m.PadButtons(0); got&start == 0 {
		t.Errorf("pad level = %04X after draining the queue: a held button must stay held", got)
	}
}
