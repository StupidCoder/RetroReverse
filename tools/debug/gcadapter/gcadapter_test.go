package gcadapter_test

import (
	"os"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/gcadapter"
	"retroreverse.com/tools/platform/gc"
)

// The disc and a savestate that is already at a drawn frame. Game images are not committed,
// and the boot to the first frame is ~1.5 billion instructions — far too long for a test —
// so both are required and the tests skip without them.
const (
	discPath  = "../../../games/luigis-mansion-gc/image/Luigi's Mansion (USA).iso"
	statePath = "../../../games/luigis-mansion-gc/work/fineb.state"
)

func newAdapter(t *testing.T) *gcadapter.Adapter {
	t.Helper()
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the Luigi's Mansion disc is not present (game images are not committed)")
	}
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Skip("the cutscene savestate is not present (work/ is regenerable scratch)")
	}
	a, err := gcadapter.New(discPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	if err := a.LoadStateFile(statePath); err != nil {
		t.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

// TestCapabilities asserts both halves of the capability contract: what this target claims,
// and what it does NOT. The negative half is the point — a capability list that only ever
// grows would let an adapter claim a panel it cannot back, which is the one thing the whole
// capability design exists to prevent.
func TestCapabilities(t *testing.T) {
	if _, err := os.Stat(discPath); os.IsNotExist(err) {
		t.Skip("the Luigi's Mansion disc is not present")
	}
	a, err := gcadapter.New(discPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()

	caps := map[string]bool{}
	for _, c := range debug.Capabilities(a) {
		caps[c] = true
	}
	for _, want := range []string{
		debug.CapFrames, debug.CapFastStep, debug.CapReplay, debug.CapCode, debug.CapBreak,
		debug.CapDisasm, debug.CapWatch, debug.CapSurfaces, debug.CapFiles, debug.CapFileAt,
		debug.CapStates, debug.CapResume, debug.CapRegions, debug.CapKeys,
	} {
		if !caps[want] {
			t.Errorf("the GameCube target does not claim %q, and it can back it", want)
		}
	}
	// The GameCube has no touch panel and no per-subsystem frame timer, so it must not
	// claim either: an empty panel is worse than no panel.
	for _, notWant := range []string{debug.CapTouch, debug.CapProfile} {
		if caps[notWant] {
			t.Errorf("the GameCube target claims %q, which it cannot honestly back", notWant)
		}
	}
}

// TestCaptureIsTheDrawTargetNotTheScanout is the guard on the lesson every adapter in this
// suite has had to learn: the picture a frame capture reports must be what the GPU drew
// into, not what the video hardware is showing.
//
// On this console the two are not even the same shape, which is what makes the test
// worthwhile: Flipper's embedded framebuffer is 640x528, and the copy that ends a frame
// crops it to the 640x480 the video interface scans. So a capture that is 640x480 is a
// capture of the wrong buffer, and it would still have looked like a plausible frame.
func TestCaptureIsTheDrawTargetNotTheScanout(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if fc.Width != 640 || fc.Height != 528 {
		t.Errorf("the capture is %dx%d; the EFB is 640x528 (640x480 would mean it captured the scanout)",
			fc.Width, fc.Height)
	}
	disp, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if disp.Rect.Dy() == fc.Height {
		t.Error("Display() and the capture are the same height; they are different buffers and should differ")
	}
}

// TestCaptureAndReplay walks a real frame end to end: the command stream is contiguous and
// named, provenance points at real commands, and a replay to the end of the stream returns
// the finished frame.
func TestCaptureAndReplay(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(true)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(fc.Commands) < 100 {
		t.Fatalf("the cutscene frame carried only %d commands; the savestate is not where it should be", len(fc.Commands))
	}
	for i, c := range fc.Commands {
		if c.Index != i {
			t.Fatalf("command %d has Index %d", i, c.Index)
		}
		if c.Name == "" || c.Name == "UNKNOWN" {
			t.Errorf("command %d (op %#x) has no name", i, c.Op)
		}
	}
	if fc.Prov == nil {
		t.Fatal("the frame reported no provenance, but the rasteriser reports every pixel")
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("provenance is %d entries for a %dx%d target", len(fc.Prov), fc.Width, fc.Height)
	}
	drawn := 0
	for _, c := range fc.Prov {
		if c >= 0 {
			drawn++
			if int(c) >= len(fc.Commands) {
				t.Fatalf("a pixel names command %d, past the end of a %d-command stream", c, len(fc.Commands))
			}
		}
	}
	if drawn == 0 {
		t.Error("no pixel in the frame has provenance")
	}
	if !fc.Drawn() {
		t.Error("the forest frame does not report itself as drawn")
	}
	if len(fc.Writers) == 0 {
		t.Error("no command was recorded as having written a pixel")
	}

	// The scrubber's final position must show the finished frame. It is the position that
	// would silently be black if the replay ran the frame's closing copy and let it clear
	// the EFB behind itself.
	img, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	if img.Rect.Dx() != fc.Width || img.Rect.Dy() != fc.Height {
		t.Errorf("the replay is %dx%d, the capture %dx%d", img.Rect.Dx(), img.Rect.Dy(), fc.Width, fc.Height)
	}
	nonBlack := 0
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] > 8 || img.Pix[i+1] > 8 || img.Pix[i+2] > 8 {
			nonBlack++
		}
	}
	if nonBlack*100 < fc.Width*fc.Height {
		t.Errorf("the replayed frame is %d/%d non-black pixels — the closing copy probably cleared it",
			nonBlack, fc.Width*fc.Height)
	}
}

// TestReplayIsDeterministic: two replays of one capture must agree. It is the property the
// whole scrubber rests on, and the one an aliased snapshot silently breaks — the second
// replay would start from whatever the first left behind.
func TestReplayIsDeterministic(t *testing.T) {
	a := newAdapter(t)
	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	k := len(fc.Commands) / 2
	first, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	second, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter (again): %v", err)
	}
	for i := range first.Pix {
		if first.Pix[i] != second.Pix[i] {
			t.Fatalf("two replays of command %d differ at byte %d (%d vs %d)", k, i, first.Pix[i], second.Pix[i])
		}
	}
}

// TestLiveMachineSurvivesAReplay: replaying into the scratch machine must not disturb the
// live one. A snapshot that aliased the live machine's framebuffer would fail this.
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
	pre := append([]byte(nil), before.Pix...)
	if _, err := a.RenderAfter(fc, len(fc.Commands)/2); err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	after, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	for i := range pre {
		if pre[i] != after.Pix[i] {
			t.Fatalf("a replay changed the live machine's picture at byte %d", i)
		}
	}
}

// TestTheDiscIsAFilesystem: the FST browses, and an offset maps back to the file that holds
// it — the capability that makes a read-watch on the drive say which of the game's own files
// is streaming.
func TestTheDiscIsAFilesystem(t *testing.T) {
	a := newAdapter(t)
	root, err := a.ListDir("/")
	if err != nil {
		t.Fatalf("ListDir(/): %v", err)
	}
	if len(root) == 0 {
		t.Fatal("the disc root lists nothing")
	}
	// Find a real file anywhere and check it round-trips through FileAt.
	var probe debug.FileEntry
	var walk func(path string, depth int)
	walk = func(path string, depth int) {
		if probe.Path != "" || depth > 3 {
			return
		}
		es, err := a.ListDir(path)
		if err != nil {
			return
		}
		for _, e := range es {
			if !e.Dir && e.Size > 0 {
				probe = e
				return
			}
		}
		for _, e := range es {
			if e.Dir {
				walk(e.Path, depth+1)
			}
		}
	}
	walk("/", 0)
	if probe.Path == "" {
		t.Skip("no file found in the first few levels of the disc")
	}
	got, within, ok := a.FileAt(probe.Offset + 16)
	if !ok {
		t.Fatalf("FileAt(%d) inside %q found nothing", probe.Offset+16, probe.Path)
	}
	if got.Path != probe.Path || within != 16 {
		t.Errorf("FileAt landed on %q+%d, want %q+16", got.Path, within, probe.Path)
	}
	b, err := a.ReadFile(probe.Path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", probe.Path, err)
	}
	if int64(len(b)) != probe.Size {
		t.Errorf("ReadFile(%q) returned %d bytes, want %d", probe.Path, len(b), probe.Size)
	}
}

// TestPadStatesEachGetAField is the guard on the pad's pacing, and the reason the pacing
// exists at all.
//
// The serial interface samples the pad as a LEVEL, once per video field, and the game
// edge-detects presses from those samples. So a browser key that goes down and comes back up
// between two samples is a press the game never sees — and that is not hypothetical, because
// a debugger's fields are wall-clock-slow and a real key tap is quick. Every distinct button
// state must therefore be live across at least one field.
//
// The observation has to be per FIELD, not per StepFast: this game renders at 30 Hz, so two
// fields pass per flip and a single StepFast can drain two queued states — an end-of-step
// sample would miss the press and call correct pacing a bug.
func TestPadStatesEachGetAField(t *testing.T) {
	a := newAdapter(t)
	m := a.Machine()

	// Watch what the pad holds at every field, chained after the adapter's own pacing hook.
	var seen []uint16
	pacing := m.OnDisplay
	if pacing == nil {
		t.Fatal("the adapter installed no field hook, so nothing is pacing the pad")
	}
	m.OnDisplay = func(mm *gc.Machine) {
		pacing(mm)
		seen = append(seen, mm.PadButtons(0))
	}

	// Press and release A with no field in between — the pathological case.
	if err := a.Key(debug.Key{Name: "a", Down: true}); err != nil {
		t.Fatal(err)
	}
	if err := a.Key(debug.Key{Name: "a", Down: false}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatalf("StepFast: %v", err)
		}
	}

	a2 := mustPadButton("a")
	pressed, released := false, false
	for _, mask := range seen {
		if mask&a2 != 0 {
			pressed = true
		} else if pressed {
			released = true // the release must come after a field that saw the press
		}
	}
	if !pressed {
		t.Errorf("A was pressed and released between two field samples and no field ever held it (saw %v)", seen)
	}
	if !released {
		t.Errorf("A never came back up after the press (saw %v)", seen)
	}
}

// A held key stays held: with the queue drained, later fields keep the last state rather
// than dropping the button. It is the other half of the pacing contract — release one state
// per field, but never invent a release.
func TestHeldPadButtonStaysHeld(t *testing.T) {
	a := newAdapter(t)
	m := a.Machine()
	if err := a.Key(debug.Key{Name: "start", Down: true}); err != nil {
		t.Fatal(err)
	}
	start := mustPadButton("start")
	for i := 0; i < 3; i++ {
		if err := a.StepFast(); err != nil {
			t.Fatalf("StepFast: %v", err)
		}
	}
	if m.PadButtons(0)&start == 0 {
		t.Error("a held Start was dropped once the state queue drained")
	}
}

// mustPadButton is the "a" button's bit, or a panic — the vocabulary is a package constant.
func mustPadButton(name string) uint16 {
	b, ok := gc.PadButton(name)
	if !ok {
		panic("gcadapter_test: no pad button " + name)
	}
	return b
}

// TestSurfacesOffersNoPhantomTextures: only the texture maps the game has actually bound
// may appear. An unbound map's registers read back zero and its size fields are biased by
// one, so it decodes as a 1x1 texture at address 0 — a surface that looks real, renders a
// single black pixel, and means nothing.
func TestSurfacesOffersNoPhantomTextures(t *testing.T) {
	a := newAdapter(t)
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	texes := 0
	for _, s := range a.Surfaces() {
		if !strings.HasPrefix(s.ID, "tex") {
			continue
		}
		texes++
		if s.W <= 1 && s.H <= 1 {
			t.Errorf("surface %q is %dx%d — an unbound texture map offered as a real one", s.ID, s.W, s.H)
		}
	}
	if texes == 0 {
		t.Error("a textured frame offered no bound-texture surfaces at all")
	}
	// The two framebuffers must always be there, and must be the two different buffers.
	ids := map[string]debug.Surface{}
	for _, s := range a.Surfaces() {
		ids[s.ID] = s
	}
	if ids["drawtarget"].H != 528 {
		t.Errorf("the drawtarget surface is %dx%d, want the 640x528 EFB", ids["drawtarget"].W, ids["drawtarget"].H)
	}
	if ids["scanout"].H != 480 {
		t.Errorf("the scanout surface is %dx%d, want the 640x480 XFB", ids["scanout"].W, ids["scanout"].H)
	}
	if !ids["ram"].Free {
		t.Error("the RAM surface should be free (the caller aims it)")
	}
}

// An unmapped key is ignored, not an error: the browser sends every keystroke.
func TestUnmappedKeyIsIgnored(t *testing.T) {
	a := newAdapter(t)
	if err := a.Key(debug.Key{Name: "F7", Down: true}); err != nil {
		t.Errorf("an unmapped key returned an error: %v", err)
	}
}
