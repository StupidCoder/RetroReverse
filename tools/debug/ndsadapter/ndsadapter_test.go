package ndsadapter_test

import (
	"image"
	"os"
	"reflect"
	"sort"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/ndsadapter"
)

// The Super Mario 64 DS image and a savestate at its title screen. Neither is
// committed (game images are copyright, work/ is ignored), so the tests skip when they
// are absent. Booting cold to a drawn frame costs 1.2 billion scheduler steps and a
// scripted stylus tap; the savestate is how the oracle reaches the interesting frame,
// and it is how these tests do too.
const (
	imagePath = "../../../games/super-mario-64-ds/Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
	statePath = "../../../games/super-mario-64-ds/work/states/title.state"
	dtcm      = "023C0000"
)

func newAdapter(t *testing.T) *ndsadapter.Adapter {
	t.Helper()
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("SM64DS image not present (game images are not committed)")
	}
	a, err := ndsadapter.New(imagePath, ndsadapter.Options{DTCM: dtcm})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// atTitle puts the machine on the title screen, where the 3D engine is actually drawing.
func atTitle(t *testing.T) *ndsadapter.Adapter {
	t.Helper()
	a := newAdapter(t)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Skip("no savestate at the title screen (work/ is not committed)")
	}
	if err := a.LoadStateFile(statePath); err != nil {
		t.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

// drawnFrame steps until a frame that actually put a picture on the screen. The first
// step after a restore lands mid-frame and legitimately carries no geometry.
func drawnFrame(t *testing.T, a *ndsadapter.Adapter, overdraw bool) *debug.FrameCapture {
	t.Helper()
	for i := 0; i < 6; i++ {
		fc, err := a.StepFrame(overdraw)
		if err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
		if fc.Drawn() {
			return fc
		}
	}
	t.Fatal("no drawn frame within six steps of the title savestate")
	return nil
}

// TestCapabilities pins the panels this target backs. A capability quietly disappearing
// is a panel quietly disappearing from the page, which is exactly the kind of regression
// nobody notices until they go looking for it.
func TestCapabilities(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	got := debug.Capabilities(a)
	sort.Strings(got)
	want := []string{
		debug.CapBreak, debug.CapCode, debug.CapDisasm, debug.CapFastStep,
		debug.CapFileAt, debug.CapFiles, debug.CapFrames, debug.CapProfile,
		debug.CapRegions, debug.CapReplay, debug.CapResume, debug.CapStates,
		debug.CapSurfaces, debug.CapTouch, debug.CapWatch,
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("capabilities = %v, want %v", got, want)
	}
	if a.Platform() != "ds" {
		t.Errorf("platform = %q, want ds", a.Platform())
	}
}

// TestFrameIsBothScreens: the DS's frame is the console's picture, not one panel. A
// debugger that showed one screen would be showing half a DS game.
func TestFrameIsBothScreens(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	if fc.Width != 256 || fc.Height != 392 {
		t.Errorf("frame = %dx%d, want 256x392 (two 256x192 panels and the bezel)", fc.Width, fc.Height)
	}
	if len(fc.Commands) == 0 {
		t.Fatal("a drawn frame captured no geometry commands")
	}
	if len(fc.Prov) != fc.Width*fc.Height {
		t.Fatalf("provenance is %d entries, want %d", len(fc.Prov), fc.Width*fc.Height)
	}

	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if b := img.Bounds(); b.Dx() != fc.Width || b.Dy() != fc.Height {
		t.Errorf("Display is %v, want the frame's %dx%d", b, fc.Width, fc.Height)
	}
}

// TestProvenanceNamesACommand: a pixel of the 3D layer names the geometry command that
// drew it, and the command it names is a vertex command — the one that closed the
// polygon. Anything else would mean the polygon's command index was stamped from the
// wrong place.
func TestProvenanceNamesACommand(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	named, vtx := 0, 0
	for _, c := range fc.Prov {
		if c < 0 {
			continue
		}
		named++
		if int(c) >= len(fc.Commands) {
			t.Fatalf("provenance names command %d, out of range [0,%d)", c, len(fc.Commands))
		}
		switch fc.Commands[c].Op {
		case 0x23, 0x24, 0x25, 0x26, 0x27, 0x28: // the VTX_* family
			vtx++
		}
	}
	if named == 0 {
		t.Fatal("no pixel of the frame names a command: the 3D layer's provenance never reached the composed frame")
	}
	// Every attributed pixel must come from a vertex command. A polygon is completed by
	// its last vertex, and nothing else can produce a fragment.
	if vtx != named {
		t.Errorf("%d of %d attributed pixels name a non-vertex command", named-vtx, named)
	}
	t.Logf("%d pixels attributed to geometry commands", named)
}

// TestProvenanceOnlyWhereThe3DLayerShows: the 3D plane is 256x192 and lands on whichever
// panel engine A drives. No pixel of the OTHER panel may be attributed — that panel has
// no 3D layer at all, and naming a draw for it would be a confident wrong answer.
func TestProvenanceOnlyWhereThe3DLayerShows(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	aTop := a.Machine().EngineAOnTop()
	for y := 0; y < fc.Height; y++ {
		for x := 0; x < fc.Width; x++ {
			if fc.Prov[y*fc.Width+x] < 0 {
				continue
			}
			onTopPanel := y < 192
			if onTopPanel != aTop {
				t.Fatalf("pixel (%d,%d) is attributed but lies on the panel engine A does not drive (engine A on top = %v)", x, y, aTop)
			}
		}
	}
}

// TestScrubberIsDeterministic: the same command position, replayed twice, must give the
// same picture. Every replay restores the same snapshot into a machine of its own, so if
// this ever fails the machine is carrying state the snapshot does not.
func TestScrubberIsDeterministic(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	k := len(fc.Commands) / 2
	first, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter: %v", err)
	}
	second, err := a.RenderAfter(fc, k)
	if err != nil {
		t.Fatalf("RenderAfter (again): %v", err)
	}
	if !sameImage(first, second) {
		t.Error("the same scrub position rendered differently twice: the replay is not deterministic")
	}
}

// TestScrubberAccumulatesGeometry: what the DS scrubber shows at command k is the frame
// as it would be if the game stopped submitting geometry there — so a later position must
// have at least as much 3D in it as an earlier one. This is the property the whole
// approach rests on, and it is not true of any other adapter here.
func TestScrubberAccumulatesGeometry(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	early, err := a.RenderAfter(fc, len(fc.Commands)/8)
	if err != nil {
		t.Fatalf("RenderAfter (early): %v", err)
	}
	late, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter (late): %v", err)
	}
	if sameImage(early, late) {
		t.Fatal("the frame looks identical at one-eighth of its commands and at all of them: the scrubber is not replaying the geometry")
	}
}

// TestBatchMatchesSerial: the parallel scrubber must agree with the serial one, pixel for
// pixel. It is the whole safety argument for running four machines at once.
func TestBatchMatchesSerial(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, false)

	n := len(fc.Commands)
	ks := []int{n / 5, n / 3, n / 2, (3 * n) / 4, n - 1}
	batch, err := a.RenderAfterBatch(fc, ks)
	if err != nil {
		t.Fatalf("RenderAfterBatch: %v", err)
	}
	if len(batch) != len(ks) {
		t.Fatalf("batch returned %d images for %d positions", len(batch), len(ks))
	}
	for i, k := range ks {
		want, err := a.RenderAfter(fc, k)
		if err != nil {
			t.Fatalf("RenderAfter(%d): %v", k, err)
		}
		if !sameImage(batch[i], want) {
			t.Errorf("batch position %d (command %d) differs from the serial replay", i, k)
		}
	}
}

// TestOverdrawIncludesRejects: the point of the overdraw history is the writes the
// rasteriser produced and then THREW AWAY — a depth or alpha failure is usually the
// answer to "why is this pixel not the colour I expect", and it is an answer the colour
// buffer cannot give, having never held it.
func TestOverdrawIncludesRejects(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	fc := drawnFrame(t, a, true)

	if fc.Overdraw == nil {
		t.Fatal("overdraw was requested and not recorded")
	}
	rejects, total := 0, 0
	for _, ws := range fc.Overdraw {
		for _, w := range ws {
			total++
			if w.Rejected {
				rejects++
			}
		}
	}
	if total == 0 {
		t.Fatal("the overdraw history is empty")
	}
	if rejects == 0 {
		t.Error("no rejected fragment in the whole frame: a depth-tested scene that never fails a depth test is not being tested")
	}
	t.Logf("%d writes, %d of them rejected", total, rejects)
}

// TestBreakpointStops: a breakpoint must stop the machine, and stop it AT the address.
func TestBreakpointStops(t *testing.T) {
	a := atTitle(t)
	defer a.Close()

	pc := a.CPU().PC
	a.SetBreakpoint(pc)
	if got := a.Breakpoints(); len(got) != 1 || got[0] != pc {
		t.Fatalf("breakpoints = %v, want [%08X]", got, pc)
	}
	r, err := a.Continue(2_000_000)
	if err != nil {
		t.Fatalf("Continue: %v", err)
	}
	if r.Kind != "breakpoint" {
		t.Fatalf("stopped for %q, want breakpoint", r.Kind)
	}
	if r.PC != pc {
		t.Errorf("stopped at %08X, want %08X", r.PC, pc)
	}
	a.ClearBreakpoint(pc)
	if len(a.Breakpoints()) != 0 {
		t.Error("the breakpoint survived being cleared")
	}
}

// TestWatchFires: a write watch over main RAM must see writes, and report the
// instruction that made them.
func TestWatchFires(t *testing.T) {
	a := atTitle(t)
	defer a.Close()

	var hits []debug.WatchHit
	a.OnWatchHit(func(h debug.WatchHit) {
		if len(hits) < 8 {
			hits = append(hits, h)
		}
	})
	if _, err := a.SetWatch(debug.Watch{Kind: "write", Lo: 0x02000000, Hi: 0x023FFFFF}); err != nil {
		t.Fatalf("SetWatch: %v", err)
	}
	if _, err := a.StepInstr(20000); err != nil {
		t.Fatalf("StepInstr: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("a write watch over the whole of main RAM saw nothing in 20,000 instructions")
	}
	if hits[0].Instr == "" {
		t.Error("a watch hit reported no instruction: the debugger cannot say what made the write")
	}
	if w := a.Watches(); len(w) != 1 || w[0].Hits == 0 {
		t.Errorf("the watch recorded no hits: %+v", w)
	}
}

// TestFilesystem: a DS cartridge has a real filesystem, and the debugger browses it. The
// FileAt half is what lets a card-read watch say WHICH FILE the game is streaming.
func TestFilesystem(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	root, err := a.ListDir("")
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(root) == 0 {
		t.Fatal("the cartridge root is empty")
	}

	// Find any file, read it, and map its own offset back to itself.
	var file debug.FileEntry
	var find func(dir string, depth int) bool
	find = func(dir string, depth int) bool {
		if depth > 3 {
			return false
		}
		es, err := a.ListDir(dir)
		if err != nil {
			return false
		}
		for _, e := range es {
			if !e.Dir && e.Size > 0 {
				file = e
				return true
			}
		}
		for _, e := range es {
			if e.Dir && find(e.Path, depth+1) {
				return true
			}
		}
		return false
	}
	if !find("", 0) {
		t.Fatal("no file found in the cartridge filesystem")
	}

	b, err := a.ReadFile(file.Path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", file.Path, err)
	}
	if int64(len(b)) != file.Size {
		t.Errorf("ReadFile(%q) gave %d bytes, the directory says %d", file.Path, len(b), file.Size)
	}

	got, within, ok := a.FileAt(file.Offset + 4)
	if !ok {
		t.Fatalf("FileAt(%d) found nothing, but %q lives there", file.Offset+4, file.Path)
	}
	if got.Path != file.Path || within != 4 {
		t.Errorf("FileAt named %q at +%d, want %q at +4", got.Path, within, file.Path)
	}
}

// TestSurfaces: every surface a target advertises must render. A panel that advertises a
// surface it cannot draw is worse than one that does not offer it.
func TestSurfaces(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	drawnFrame(t, a, false)

	for _, s := range a.Surfaces() {
		v := debug.View{}
		if s.Free {
			v = debug.View{W: 64, H: 64, Format: s.Formats[len(s.Formats)-1]} // direct16
		}
		img, err := a.RenderSurface(s.ID, v)
		if err != nil {
			t.Errorf("RenderSurface(%q): %v", s.ID, err)
			continue
		}
		if img.Bounds().Empty() {
			t.Errorf("RenderSurface(%q) returned an empty image", s.ID)
		}
	}
}

// TestRegionsAndProfile: the memory map names real spans, and the profile's buckets are
// disjoint and sum to the total — a bar chart that does not add up is a bug report about
// the profiler.
func TestRegionsAndProfile(t *testing.T) {
	a := atTitle(t)
	defer a.Close()
	drawnFrame(t, a, false)

	regions := a.Regions()
	if len(regions) < 5 {
		t.Fatalf("only %d memory regions: the DS map is bigger than that", len(regions))
	}
	for _, r := range regions {
		if r.Hi <= r.Lo {
			t.Errorf("region %q is empty or inverted: %08X..%08X", r.Name, r.Lo, r.Hi)
		}
	}

	p := a.FrameProfile()
	if len(p.Buckets) == 0 {
		t.Fatal("the profile has no buckets")
	}
	var sum float64
	for _, b := range p.Buckets {
		if b.Millis < 0 {
			t.Errorf("bucket %q has negative time", b.Name)
		}
		sum += b.Millis
	}
	// The CPU bucket is a remainder, so the buckets sum to the total by construction.
	if d := sum - p.TotalMs; d > 0.5 || d < -0.5 {
		t.Errorf("the buckets sum to %.2f ms but the frame took %.2f", sum, p.TotalMs)
	}
}

// TestResumeArgs: the resume line must name this game's oracle and its own flags. The
// oracles do not share a savestate flag, and pretending they do produces a command that
// looks right and does not run.
func TestResumeArgs(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	args := a.ResumeArgs("/tmp/x.state")
	joined := ""
	for _, s := range args {
		joined += s + " "
	}
	for _, want := range []string{"bootoracle", "-loadstate", "/tmp/x.state"} {
		if !contains(joined, want) {
			t.Errorf("resume args %v do not mention %q", args, want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestTouchPanel pins the stylus to the lower panel of the composed frame, and checks that
// a touch on it reaches the machine's digitiser. The rect matters as much as the touch: if
// it is wrong, every click lands somewhere the player never clicked, and the game will
// happily respond to it — a wrong answer that looks like a working feature.
func TestTouchPanel(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	panels := a.TouchPanels()
	if len(panels) != 1 {
		t.Fatalf("touch panels = %d, want 1", len(panels))
	}
	p := panels[0]
	// The composed frame is 256 x 392: the top panel, an 8px bezel, then the touchscreen.
	if p.ID != "bottom" || p.X != 0 || p.Y != 200 || p.W != 256 || p.H != 192 {
		t.Errorf("touch panel = %+v, want {bottom 0 200 256 192}", p)
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if b := img.Bounds(); p.X+p.W > b.Dx() || p.Y+p.H > b.Dy() {
		t.Errorf("touch panel %+v does not fit the composed frame %v", p, b)
	}

	if err := a.Touch(debug.Touch{Panel: "bottom", X: 128, Y: 96, Down: true}); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	if x, y, down := a.Machine().Touch(); x != 128 || y != 96 || !down {
		t.Errorf("stylus = (%d,%d) down=%v, want (128,96) down=true", x, y, down)
	}

	if err := a.Touch(debug.Touch{Panel: "top", X: 1, Y: 1, Down: true}); err == nil {
		t.Error("touching the top panel was accepted; the digitiser is not on it")
	}
}

func sameImage(a, b *image.RGBA) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !a.Bounds().Eq(b.Bounds()) {
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
