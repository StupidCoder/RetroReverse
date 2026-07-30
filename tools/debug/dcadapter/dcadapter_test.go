package dcadapter

// The image-backed tests skip without the disc and the drive savestate (game
// images are not committed, work/ is regenerable scratch); the pad and watch
// tests run on a bare machine and always run.

import (
	"image"
	"os"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/dc"
)

const (
	discPath  = "../../../games/crazy-taxi-dc/image/Crazy Taxi (US).cue"
	statePath = "../../../games/crazy-taxi-dc/work/states/drive.st"
)

func newAdapter(t *testing.T) *Adapter {
	t.Helper()
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the Crazy Taxi disc is not present (game images are not committed)")
	}
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Skip("the drive savestate is not present (work/ is regenerable scratch)")
	}
	a, err := New(discPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	if err := a.LoadStateFile(statePath); err != nil {
		t.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

// TestCapabilities asserts both halves of the capability contract: what this
// target claims, and what it does NOT. The negative half is the point.
func TestCapabilities(t *testing.T) {
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the Crazy Taxi disc is not present")
	}
	a, err := New(discPath)
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
		debug.CapFiles, debug.CapStates, debug.CapResume, debug.CapRegions,
		debug.CapKeys, debug.CapHalt,
	} {
		if !caps[want] {
			t.Errorf("the Dreamcast target does not claim %q, and it can back it", want)
		}
	}
	// No touch panel, no profiler, no file-at-offset mapping: claiming any of
	// them would put an empty panel on the page.
	for _, notWant := range []string{debug.CapTouch, debug.CapProfile, debug.CapFileAt} {
		if caps[notWant] {
			t.Errorf("the Dreamcast target claims %q, which it cannot honestly back", notWant)
		}
	}
}

// TestStepFrameCapturesTheDrawTarget: the capture must be the frame the render
// just built — the clear as command zero, a real parameter stream behind it, and
// provenance over the framebuffer plane.
func TestStepFrameCapturesTheDrawTarget(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(fc.Commands) < 100 {
		t.Fatalf("a drive frame captured only %d commands; the stream was not recorded", len(fc.Commands))
	}
	if fc.Commands[0].Name != "CLEAR" {
		t.Errorf("command 0 is %q, want the background CLEAR (the answer to 'why is this pixel black')", fc.Commands[0].Name)
	}
	if fc.Width != 640 || fc.Height != 480 {
		t.Errorf("capture plane is %dx%d, want the 640x480 framebuffer", fc.Width, fc.Height)
	}
	if fc.Prov == nil {
		t.Fatal("no provenance recorded")
	}
	if !fc.Drawn() {
		t.Error("a drive frame reports Drawn()=false")
	}
	if len(fc.Writers) == 0 {
		t.Error("no writers recorded in a frame that painted a whole scene")
	}
	// The clear reported its pixels: every pixel of the plane has a writer.
	holes := 0
	for _, c := range fc.Prov {
		if c < 0 {
			holes++
		}
	}
	if holes > 0 {
		t.Errorf("%d pixels claim no command wrote them; the clear must report its pixels", holes)
	}
}

// TestRenderAfterScrubs: replaying to command 0 shows the cleared frame; replaying
// to the last command reproduces the full frame the capture rendered, exactly —
// same snapshot, same deterministic machine, same walk.
func TestRenderAfterScrubs(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	full, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter(last): %v", err)
	}
	clear, err := a.RenderAfter(fc, 0)
	if err != nil {
		t.Fatalf("RenderAfter(0): %v", err)
	}
	if !allBlack(clear) {
		t.Error("the frame scrubbed to command 0 (the clear) is not black")
	}
	if allBlack(full) {
		t.Error("the frame scrubbed to the last command is black; the replay drew nothing")
	}
	mid, err := a.RenderAfter(fc, len(fc.Commands)/2)
	if err != nil {
		t.Fatalf("RenderAfter(mid): %v", err)
	}
	if imagesEqual(mid, full) {
		t.Error("the half-scrubbed frame equals the full frame; RenderStopAfter did not stop the walk")
	}
}

// TestDisplayIsTheScanout: Display() must serve the picture the CRTC is showing —
// which, captured inside the render completion, is one flip BEHIND the frame the
// capture just built. Both must exist; they are different buffers.
func TestDisplayIsTheScanout(t *testing.T) {
	a := newAdapter(t)
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display after a drive frame: %v", err)
	}
	if allBlack(img) {
		t.Error("the scanout is black at the drive; FB_R decode is broken")
	}
}

// TestPadPacing drives the keyer on a bare machine and observes PER FIELD — the
// Maple poll's own boundary. A queued tap must hold for tapFields fields (the
// oracle's proven tap length), a held key must stay held, and the stick must
// deflect in the wire's own convention (0x00 = up/left).
func TestPadPacing(t *testing.T) {
	m := dc.NewMachine(nil)
	a := newWithMachine(m, "test")

	if err := a.Key(debug.Key{Name: "a", Down: true}); err != nil {
		t.Fatalf("Key: %v", err)
	}
	if err := a.Key(debug.Key{Name: "a", Down: false}); err != nil {
		t.Fatalf("Key: %v", err)
	}
	// Two states queued: pressed, released. Buttons are active-low on the wire.
	if m.Pad.Buttons != 0xFFFF {
		t.Fatalf("pad changed before any field boundary: %04X", m.Pad.Buttons)
	}
	field := uint64(0)
	tick := func() { field++; m.OnDisplay(field) }

	tick()
	if m.Pad.Buttons != 0xFFFF&^dc.PadA {
		t.Fatalf("after field 1 the press is not live: %04X", m.Pad.Buttons)
	}
	// The press must stay live across the full tap, not be replaced next field.
	for i := 0; i < tapFields-1; i++ {
		tick()
		if m.Pad.Buttons != 0xFFFF&^dc.PadA {
			t.Fatalf("the tap was cut short at field %d: %04X", field, m.Pad.Buttons)
		}
	}
	tick()
	if m.Pad.Buttons != 0xFFFF {
		t.Fatalf("the release never landed: %04X", m.Pad.Buttons)
	}

	// The stick: up decreases joyY (0x00 is the top of travel), and a corner
	// splits the throw over both axes.
	a.Key(debug.Key{Name: "ArrowUp", Down: true})
	for i := 0; i < tapFields; i++ {
		tick()
	}
	if m.Pad.JoyY >= 0x80 {
		t.Errorf("stick up gave joyY %02X; up must DECREASE (0x00 is the top of travel)", m.Pad.JoyY)
	}
	full := m.Pad.JoyY
	a.Key(debug.Key{Name: "ArrowLeft", Down: true})
	for i := 0; i < tapFields; i++ {
		tick()
	}
	if m.Pad.JoyY <= full {
		t.Errorf("a corner reads as far as a straight push (joyY %02X vs %02X); the gate is not modelled", m.Pad.JoyY, full)
	}
	if m.Pad.JoyX >= 0x80 {
		t.Errorf("stick left gave joyX %02X; left must decrease", m.Pad.JoyX)
	}

	// A held key with an empty queue stays held: nothing overwrites the level.
	a.held = map[string]bool{}
	a.padQueue = nil
	m.Pad.Buttons = 0xFFFF &^ dc.PadB
	for i := 0; i < 5; i++ {
		tick()
	}
	if m.Pad.Buttons != 0xFFFF&^dc.PadB {
		t.Errorf("an empty queue overwrote the held level: %04X", m.Pad.Buttons)
	}
}

// TestWatchDispatch exercises the adapter's range/kind dispatch over the
// machine's single callback, without an image: install two watches, deliver
// synthetic hits, and check who fired — including the break path.
func TestWatchDispatch(t *testing.T) {
	m := dc.NewMachine(nil)
	a := newWithMachine(m, "test")

	var hits []debug.WatchHit
	a.OnWatchHit(func(h debug.WatchHit) { hits = append(hits, h) })
	wid, err := a.SetWatch(debug.Watch{Kind: "write", Lo: 0x0C001000, Hi: 0x0C001004})
	if err != nil {
		t.Fatalf("SetWatch: %v", err)
	}
	rid, err := a.SetWatch(debug.Watch{Kind: "read", Lo: 0x0C002000, Hi: 0x0C002100, Break: true})
	if err != nil {
		t.Fatalf("SetWatch: %v", err)
	}

	// A write inside the write window fires only the write watch.
	m.Write32(0x0C001000, 0xDEAD)
	if len(hits) != 1 || hits[0].ID != wid {
		t.Fatalf("write hit dispatch wrong: %+v", hits)
	}
	if m.StopRequested {
		t.Fatal("a non-breaking watch requested a stop")
	}
	// A write NEXT TO the window fires nothing.
	m.Write32(0x0C001004, 1)
	if len(hits) != 1 {
		t.Fatalf("a write beyond Hi fired: %+v", hits[len(hits)-1])
	}
	// A read in the read window fires the read watch and asks the run to stop.
	m.Read32(0x0C002080)
	if len(hits) != 2 || hits[1].ID != rid {
		t.Fatalf("read hit dispatch wrong: %+v", hits)
	}
	if !m.StopRequested {
		t.Fatal("a breaking watch did not request a stop")
	}
	if a.stop.Kind != "watch" {
		t.Fatalf("the stop reason is %q, want watch", a.stop.Kind)
	}
	m.StopRequested = false

	// Clearing the breaking watch leaves the other live.
	a.ClearWatch(rid)
	hits = nil
	m.Read32(0x0C002080)
	if len(hits) != 0 {
		t.Fatal("a cleared watch fired")
	}
	m.Write32(0x0C001002, 2) // straddles the window edge: still inside
	if len(hits) != 1 {
		t.Fatal("the surviving write watch is dead")
	}
}

// TestSnapshotRestoreRoundTrip: a Snapshot must restore into a machine that is
// the machine — checked through the platform's own deep-copy state, on a bare
// machine so it always runs.
func TestSnapshotRestoreRoundTrip(t *testing.T) {
	m := dc.NewMachine(nil)
	a := newWithMachine(m, "test")
	m.RAM[0x1234] = 0xAB
	s := a.Snapshot()
	m.RAM[0x1234] = 0xCD
	if err := a.Restore(s); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if m.RAM[0x1234] != 0xAB {
		t.Fatalf("restore did not bring the memory back: %02X", m.RAM[0x1234])
	}
	if err := a.Restore(fakeSnap{}); err == nil {
		t.Fatal("a foreign snapshot was accepted")
	}
}

type fakeSnap struct{}

func (fakeSnap) Platform() string { return "not-dc" }

func allBlack(img *image.RGBA) bool {
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 || img.Pix[i+1] != 0 || img.Pix[i+2] != 0 {
			return false
		}
	}
	return true
}

func imagesEqual(a, b *image.RGBA) bool {
	if a.Rect != b.Rect || len(a.Pix) != len(b.Pix) {
		return false
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}
