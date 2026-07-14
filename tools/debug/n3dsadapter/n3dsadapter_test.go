package n3dsadapter_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n3dsadapter"
	"retroreverse.com/tools/platform/n3ds"
)

// The Captain Toad image and the savestate that reaches its opening stage. Neither
// is committed (game images are copyright, work/ is ignored), so the tests skip
// when they are absent. Booting cold to a drawn frame costs minutes and a scripted
// walk through the StreetPass and file-select screens; the savestate is how the
// oracle gets to the interesting frame, and it is how these tests do too.
const (
	imagePath = "../../../games/captain-toad-treasure-tracker-3ds/image/Captain Toad - Treasure Tracker (Europe) (En,Fr,De,Es,It,Nl).cci"
	statePath = "../../../games/captain-toad-treasure-tracker-3ds/work/toadstereo4.state"
)

func newAdapter(t *testing.T) *n3dsadapter.Adapter {
	t.Helper()
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Captain Toad image not present (game images are not committed)")
	}
	a, err := n3dsadapter.New(imagePath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// atScene puts the machine at the opening stage, where the GPU is actually drawing.
func atScene(t *testing.T) *n3dsadapter.Adapter {
	t.Helper()
	a := newAdapter(t)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Skip("no savestate at the opening stage (work/ is not committed)")
	}
	if err := a.LoadStateFile(statePath); err != nil {
		t.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

func TestCapabilities(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	want := []string{
		debug.CapFrames, debug.CapFastStep, debug.CapReplay, debug.CapCode,
		debug.CapBreak, debug.CapDisasm, debug.CapWatch, debug.CapSurfaces,
		debug.CapFiles, debug.CapFileAt, debug.CapStates, debug.CapResume,
		debug.CapRegions, debug.CapProfile, debug.CapTouch,
	}
	got := map[string]bool{}
	for _, c := range debug.Capabilities(a) {
		got[c] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("target does not advertise %q", w)
		}
	}
	if a.Platform() != "3ds" {
		t.Errorf("Platform() = %q, want 3ds", a.Platform())
	}
}

// TestReplayIsDeterministic is the property the command scrubber is built on, and
// the one that would fail silently if it were broken.
//
// RenderAfter replays a frame from its start snapshot. That is only meaningful if
// the snapshot is a genuine deep copy of the machine — restore it and the frame
// must build itself the same way, write for write and pixel for pixel. A snapshot
// that aliased even one slice of the live machine (a region, a thread context, the
// DSP's sample state) would still restore, still run, and still produce a
// plausible picture — just not the frame that was captured. So this restores the
// capture's own start state, runs the frame again, and demands the two agree
// exactly.
func TestReplayIsDeterministic(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	first, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(first.Commands) == 0 {
		t.Skip("the savestate's next frame draws nothing")
	}

	if err := a.Restore(first.Start); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	second, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame (replayed): %v", err)
	}

	if len(second.Commands) != len(first.Commands) {
		t.Fatalf("replayed frame ran %d commands, the captured one ran %d — the snapshot is not a faithful copy of the machine",
			len(second.Commands), len(first.Commands))
	}
	for i := range first.Commands {
		if !reflect.DeepEqual(first.Commands[i], second.Commands[i]) {
			t.Fatalf("command %d differs on replay: %+v vs %+v", i, first.Commands[i], second.Commands[i])
		}
	}
	if first.Width != second.Width || first.Height != second.Height {
		t.Fatalf("replayed frame drew into a %dx%d target, the captured one into %dx%d",
			second.Width, second.Height, first.Width, first.Height)
	}
	if len(first.Prov) != len(second.Prov) {
		t.Fatalf("provenance planes differ in size: %d vs %d", len(first.Prov), len(second.Prov))
	}
	for i := range first.Prov {
		if first.Prov[i] != second.Prov[i] {
			t.Fatalf("pixel %d was drawn by command %d on the first run and %d on the replay",
				i, first.Prov[i], second.Prov[i])
		}
	}
}

// TestFrameAndDisplayAgree pins the invariant the pixel overlay depends on: the
// picture the debugger shows for a frame and the plane its provenance is recorded
// in must be the same plane. The 3DS is the platform where these come apart — the
// GPU renders into a padded, tiled VRAM buffer and a DisplayTransfer later crops a
// rotated copy of it to the LCD — and if they ever disagree, clicking a pixel
// reports a command that never touched the buffer on screen.
func TestFrameAndDisplayAgree(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if w, h := img.Bounds().Dx(), img.Bounds().Dy(); w != fc.Width || h != fc.Height {
		t.Fatalf("the frame's picture is %dx%d but its provenance plane is %dx%d — the overlay would be misaligned",
			w, h, fc.Width, fc.Height)
	}
}

// TestProvenanceNamesADraw checks that the commands provenance points at are the
// ones that can actually produce a pixel. A register write that merely latches
// state cannot draw anything, so if provenance ever named one, the pixel hook and
// the command hook would be out of step.
func TestProvenanceNamesADraw(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if fc.Prov == nil {
		t.Skip("the savestate's next frame drew no pixels")
	}
	checked := 0
	for _, k := range fc.Prov {
		if k < 0 {
			continue
		}
		name := fc.Commands[k].Name
		if name != "DRAWARRAYS" && name != "DRAWELEMENTS" {
			t.Fatalf("pixel attributed to command %d (%s), which cannot draw", k, name)
		}
		checked++
	}
	if checked == 0 {
		t.Skip("no pixel in the frame has provenance")
	}
	t.Logf("%d pixels attributed across %d commands", checked, len(fc.Commands))
}

// TestScrubProgresses checks that replaying to different points in the command
// stream actually renders different pictures — that the stop-after-k halt is
// honoured rather than always running the list to the end.
func TestScrubProgresses(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(fc.Commands) < 2 {
		t.Skip("nothing to scrub")
	}
	early, err := a.RenderAfter(fc, 0)
	if err != nil {
		t.Fatalf("RenderAfter(0): %v", err)
	}
	late, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter(last): %v", err)
	}
	if early.Bounds() != late.Bounds() {
		t.Fatalf("the scrubber changed the frame's size mid-scrub: %v then %v", early.Bounds(), late.Bounds())
	}
	if string(early.Pix) == string(late.Pix) {
		t.Error("the first and last command render the same picture — the stop-after-k halt is not being honoured")
	}

	// And the replay must be repeatable: the same k twice is the same picture.
	again, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		t.Fatalf("RenderAfter(last) again: %v", err)
	}
	if string(again.Pix) != string(late.Pix) {
		t.Error("replaying the same command twice rendered two different pictures")
	}
}

// TestRomFS browses the cartridge filesystem the way the files panel does.
func TestRomFS(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	root, err := a.ListDir("/")
	if err != nil {
		t.Fatalf("ListDir(/): %v", err)
	}
	if len(root) == 0 {
		t.Fatal("the RomFS root is empty")
	}
	// Find a file anywhere in the tree, read it, and check FileAt names it back.
	var file debug.FileEntry
	var walk func(string, int)
	walk = func(dir string, depth int) {
		if file.Path != "" || depth > 3 {
			return
		}
		entries, err := a.ListDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.Dir && e.Size > 0 {
				file = e
				return
			}
		}
		for _, e := range entries {
			if e.Dir {
				walk(e.Path, depth+1)
			}
		}
	}
	walk("/", 0)
	if file.Path == "" {
		t.Skip("no file found in the first levels of the RomFS")
	}

	data, err := a.ReadFile(file.Path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", file.Path, err)
	}
	if int64(len(data)) != file.Size {
		t.Errorf("%s: read %d bytes, the entry says %d", file.Path, len(data), file.Size)
	}
	got, within, ok := a.FileAt(file.Offset)
	if !ok {
		t.Fatalf("FileAt(%d) found nothing, but %s starts there", file.Offset, file.Path)
	}
	if got.Path != file.Path || within != 0 {
		t.Errorf("FileAt(%d) = %s+%d, want %s+0", file.Offset, got.Path, within, file.Path)
	}
}

// TestDisasm decodes at the entry point: the ARM11 boots in ARM state, so the
// instructions there must decode as ARM, four bytes apiece.
func TestDisasm(t *testing.T) {
	a := newAdapter(t)
	defer a.Close()

	pc := a.CPU().PC
	ins, err := a.Disasm(pc, 4)
	if err != nil {
		t.Fatalf("Disasm: %v", err)
	}
	if len(ins) != 4 {
		t.Fatalf("Disasm returned %d instructions, want 4", len(ins))
	}
	for i, in := range ins {
		if in.Addr != pc+uint64(i)*4 {
			t.Errorf("instruction %d is at %#x, want %#x", i, in.Addr, pc+uint64(i)*4)
		}
		if strings.TrimSpace(in.Text) == "" {
			t.Errorf("instruction %d at %#x disassembled to nothing", i, in.Addr)
		}
	}
}

// TestWatchReportsTheWriter is what a memory watch is for: not that a location
// changed, but which instruction changed it.
func TestWatchReportsTheWriter(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	var hits []debug.WatchHit
	a.OnWatchHit(func(h debug.WatchHit) { hits = append(hits, h) })

	// The GPU's colour buffer: something writes it every frame.
	regions := a.Regions()
	var vram debug.Region
	for _, r := range regions {
		if r.Name == "vram" {
			vram = r
		}
	}
	if vram.Hi == 0 {
		t.Skip("no vram region in the memory map")
	}
	if _, err := a.SetWatch(debug.Watch{Kind: "write", Lo: vram.Lo, Hi: vram.Lo + 0x1000}); err != nil {
		t.Fatalf("SetWatch: %v", err)
	}
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if len(hits) == 0 {
		t.Skip("nothing wrote the first page of VRAM this frame")
	}
	for _, h := range hits {
		if h.Addr < vram.Lo || h.Addr >= vram.Lo+0x1000 {
			t.Fatalf("watch reported %#x, outside the window [%#x,%#x)", h.Addr, vram.Lo, vram.Lo+0x1000)
		}
	}
	// A second watch of the same kind must be refused, not silently swallow the first.
	if _, err := a.SetWatch(debug.Watch{Kind: "write", Lo: 0, Hi: 4}); err == nil {
		t.Error("a second write watch was accepted; the machine has only one write window")
	}
	t.Logf("%d writes into the first page of VRAM, first from pc=%#x (%s)", len(hits), hits[0].PC, hits[0].Instr)
}

// TestFrameIsTheScreenLayout checks the frame is the console's picture: the top
// screen (400x240) above the bottom (320x240), each horizontally centred.
func TestFrameIsTheScreenLayout(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	m := a.Machine()
	top, ok := m.ScreenGeom("top")
	if !ok {
		t.Fatal("the top screen has no transfer geometry")
	}
	bot, ok := m.ScreenGeom("bottom")
	if !ok {
		t.Fatal("the bottom screen has no transfer geometry")
	}
	tw, th := top.Size()
	bw, bh := bot.Size()
	if tw != 400 || th != 240 {
		t.Errorf("top screen is %dx%d, want 400x240", tw, th)
	}
	if bw != 320 || bh != 240 {
		t.Errorf("bottom screen is %dx%d, want 320x240", bw, bh)
	}
	if fc.Width != 400 {
		t.Errorf("composed frame is %d wide, want the wider screen's 400", fc.Width)
	}
	if fc.Height < th+bh {
		t.Errorf("composed frame is %d tall, too short to stack %d over %d", fc.Height, th, bh)
	}
}

// TestTouchPanelIsTheBottomScreen pins where the stylus can reach. Getting the rectangle
// wrong is worse than getting it missing: every click would land somewhere the player
// never clicked, and the game would respond to it perfectly happily.
//
// The bottom screen is 320 wide under a 400-wide top one and the frame centres both, so
// the panel's origin is 40 pixels in — it is emphatically not (0, 240).
func TestTouchPanelIsTheBottomScreen(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	// A cold machine has presented nothing, so there is nowhere to touch, and saying so is
	// the honest answer rather than a rectangle over a picture that is not the touchscreen.
	if cold := newAdapter(t); len(cold.TouchPanels()) != 0 {
		t.Errorf("a machine that has presented no screen offered a touch panel: %+v", cold.TouchPanels())
	} else {
		cold.Close()
	}

	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	panels := a.TouchPanels()
	if len(panels) != 1 {
		t.Fatalf("touch panels = %+v, want the bottom screen alone", panels)
	}
	p := panels[0]
	if p.ID != "bottom" || p.W != 320 || p.H != 240 {
		t.Errorf("touch panel = %+v, want the 320x240 bottom screen", p)
	}
	if p.X != 40 {
		t.Errorf("touch panel starts at x=%d; the 320-wide screen is centred under the 400-wide one, so it is 40", p.X)
	}
	img, err := a.Display()
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if b := img.Bounds(); p.X+p.W > b.Dx() || p.Y+p.H > b.Dy() {
		t.Errorf("touch panel %+v does not fit the composed frame %v", p, b)
	}

	if err := a.Touch(debug.Touch{Panel: "bottom", X: 160, Y: 120, Down: true}); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	if err := a.Touch(debug.Touch{Panel: "top", X: 1, Y: 1, Down: true}); err == nil {
		t.Error("touching the top screen was accepted; the digitiser is not on it")
	}
}

// TestScreenDecodeMatchesTheTransfer is the cross-check on the whole mapping. The debugger
// decodes a screen straight from the tiled render target (so a replayed frame can fill in
// draw by draw); the machine's own transfer wrote it into a linear framebuffer. Those are
// two independent paths through the crop, the row offset and the rotation, and they must
// land on the same picture — if they disagree, the provenance the debugger maps through the
// first path is pointing at pixels the player never saw.
//
// The cross-check is made on the frame's LAST pass, and only on that one, because a 3DS
// frame is a sequence of passes THROUGH ONE RENDER TARGET: SM3DL clears and reuses the same
// buffer for the top screen, the other eye, and the bottom screen. So only the last pass's
// picture is still in the render target when the frame ends. Comparing an earlier pass would
// be comparing a screen against the pixels of the screen that overwrote it — which is exactly
// the bug this pass model exists to fix, and it used to be what this test asserted.
func TestScreenDecodeMatchesTheTransfer(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	m := a.Machine()
	var last n3ds.ScreenGeom
	var lastScreen string
	m.OnPresent = func(screen string, g n3ds.ScreenGeom) { last, lastScreen = g, screen }
	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	m.OnPresent = nil
	if last.Dst == 0 {
		t.Skip("the savestate's next frame presented nothing")
	}

	ours := m.ScreenImage(last)      // the render target, as it stands
	theirs := m.PresentedImage(last) // what this pass's own transfer produced
	if theirs == nil {
		t.Fatalf("%s: the last pass has no presented picture", lastScreen)
	}
	if ours.Rect != theirs.Rect {
		t.Fatalf("%s: decoded %v, the transfer's framebuffer is %v", lastScreen, ours.Rect, theirs.Rect)
	}
	diff := 0
	for i := range ours.Pix {
		if ours.Pix[i] != theirs.Pix[i] {
			diff++
		}
	}
	if diff != 0 {
		t.Errorf("%s (the frame's last pass): %d of %d bytes differ between the render-target decode and the picture its transfer produced",
			lastScreen, diff, len(ours.Pix))
	}
}

// TestTheScreensAreNotTheSamePicture is the bug this pass model was built for.
//
// Both games render every screen through ONE render target — clear, draw the top screen,
// transfer it out, clear, draw the bottom screen into the same buffer, transfer it out. A
// debugger that reads both panels out of the render target at the end of the frame reads
// the same bytes twice, and shows the bottom screen on both panels. On Super Mario 3D Land
// that meant the welcome dialog appeared where the logo should be.
func TestTheScreensAreNotTheSamePicture(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	// Several frames: the first ones after a restore are mid-render, and the LCD is still
	// pointed at a buffer from before the state.
	drawn := 0
	for i := 0; i < 24 && drawn < 4; i++ {
		fc, err := a.StepFrame(false)
		if err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
		if fc.Drawn() {
			drawn++
		}
	}
	img, err := a.Display()
	if err != nil {
		t.Skip("nothing presented yet")
	}

	// The panels sit one above the other, each centred: compare the rows they share.
	const gap = 8
	top, ok := a.Machine().ScreenGeom("top")
	if !ok {
		t.Skip("no top screen")
	}
	_, topH := top.Size()
	same, n := 0, 0
	for y := 0; y+topH+gap < img.Rect.Dy() && y < topH; y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			ti := img.PixOffset(x, y)
			bi := img.PixOffset(x, y+topH+gap)
			n++
			if img.Pix[ti] == img.Pix[bi] && img.Pix[ti+1] == img.Pix[bi+1] && img.Pix[ti+2] == img.Pix[bi+2] {
				same++
			}
		}
	}
	if n == 0 {
		t.Skip("only one screen is presented")
	}
	if same*2 > n {
		t.Errorf("%d of %d pixels are identical on both panels — the two screens are being decoded from the same buffer",
			same, n)
	}
}

// TestProvenanceReachesTheScreen checks that provenance survives the trip into
// screen space: the whole point of composing the frame from the panels is that a
// click on the picture still names the draw that made it, so the top screen must be
// substantially attributed rather than a plane of -1.
func TestProvenanceReachesTheScreen(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	if fc.Prov == nil {
		t.Skip("the savestate's next frame drew no pixels")
	}
	top, ok := a.Machine().ScreenGeom("top")
	if !ok {
		t.Skip("nothing presented")
	}
	tw, th := top.Size()
	x0 := (fc.Width - tw) / 2 // the top screen sits at the top, horizontally centred
	attributed := 0
	for sy := 0; sy < th; sy++ {
		for sx := 0; sx < tw; sx++ {
			if fc.ProvAt(x0+sx, sy) >= 0 {
				attributed++
			}
		}
	}
	total := tw * th
	if attributed*4 < total { // the scene covers the screen; a quarter is a floor, not a target
		t.Errorf("only %d of %d top-screen pixels carry provenance — the mapping into screen space is losing them",
			attributed, total)
	}
	t.Logf("%d of %d top-screen pixels attributed to a draw", attributed, total)
}

// TestFrameProfile checks that a stepped frame reports where its time went, and —
// the part that is easy to get wrong — that the buckets are disjoint: they must
// add up to the total, because the panel draws them as one stacked bar and a bar
// whose segments overlap is a lie told in colour.
func TestFrameProfile(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	// Step until a frame that actually drew: Captain Toad renders in bursts, so the
	// first frame after a restore is quite often an idle one.
	var p debug.FrameProfile
	for i := 0; i < 8; i++ {
		if _, err := a.StepFrame(false); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
		p = a.FrameProfile()
		// Drew is what the debugger folds fields into frames on, so it has to agree with
		// the frame's own draw count — on the idle fields as much as on the drawing ones.
		if p.Drew != drew(p) {
			t.Errorf("field %d: Drew = %v but the frame reports %d draws", i, p.Drew, draws(p))
		}
		if drew(p) {
			break
		}
	}
	if !drew(p) {
		t.Fatal("no frame drew anything in 8 steps; the profile has nothing to say")
	}
	if p.TotalMs <= 0 {
		t.Fatalf("TotalMs = %v, want > 0", p.TotalMs)
	}

	var sum float64
	for _, b := range p.Buckets {
		if b.Millis < 0 {
			t.Errorf("bucket %q reports negative time %v", b.Name, b.Millis)
		}
		sum += b.Millis
	}
	// The buckets are disjoint leaves plus a derived remainder, so they sum to the
	// total by construction. Allow a hair for the float arithmetic.
	if d := sum - p.TotalMs; d > 0.5 || d < -0.5 {
		t.Errorf("buckets sum to %.2f ms but the frame took %.2f ms (%.2f unaccounted) — they are overlapping or a path is unbucketed",
			sum, p.TotalMs, d)
	}
	if len(p.Counters) == 0 {
		t.Error("no counters: milliseconds alone cannot tell a faster rasteriser from a frame that drew less")
	}
	for _, b := range p.Buckets {
		t.Logf("  %-24s %7.1f ms  ×%d", b.Name, b.Millis, b.Count)
	}
	for _, c := range p.Counters {
		t.Logf("  %-24s %d", c.Name, c.Value)
	}
}

// drew reports whether the profiled frame did any drawing.
func drew(p debug.FrameProfile) bool { return draws(p) > 0 }

func draws(p debug.FrameProfile) int {
	for _, c := range p.Counters {
		if c.Name == "draws" {
			return c.Value
		}
	}
	return 0
}

// TestRenderAfterBatchMatchesSerial is the whole correctness claim for the parallel
// scrubber: a batch of replays must produce exactly the images the same replays
// produce one at a time. They run on different machines, so if any state leaked
// between them — a shared scratch machine, a cached texture, a snapshot restored into
// the wrong place — this is what would show it.
func TestRenderAfterBatchMatchesSerial(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	fc, err := a.StepFrame(false)
	if err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	n := len(fc.Commands)
	if n < 8 {
		t.Skip("the frame has too few commands to scrub")
	}
	ks := []int{n / 8, n / 4, n / 2, 3 * n / 4, n - 1, n / 3}

	want := make([]string, len(ks))
	for i, k := range ks {
		img, err := a.RenderAfter(fc, k)
		if err != nil {
			t.Fatalf("RenderAfter(%d): %v", k, err)
		}
		want[i] = string(img.Pix)
	}

	got, err := a.RenderAfterBatch(fc, ks)
	if err != nil {
		t.Fatalf("RenderAfterBatch: %v", err)
	}
	if len(got) != len(ks) {
		t.Fatalf("batch returned %d images for %d keys", len(got), len(ks))
	}
	for i, k := range ks {
		if string(got[i].Pix) != want[i] {
			t.Errorf("command %d: the batched replay differs from the serial one", k)
		}
	}
}

// BenchmarkScrub is what the parallel scrubber is for: dragging over the command
// stream. Serial is one full frame replay per position; batched is several at once.
func BenchmarkScrub(b *testing.B) {
	a := atSceneB(b)
	defer a.Close()
	fc, err := a.StepFrame(false)
	if err != nil {
		b.Fatal(err)
	}
	n := len(fc.Commands)
	if n < 8 {
		b.Skip("nothing to scrub")
	}
	ks := []int{n / 8, n / 4, 3 * n / 8, n / 2, 5 * n / 8, 3 * n / 4, 7 * n / 8, n - 1}

	b.Run("serial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, k := range ks {
				if _, err := a.RenderAfter(fc, k); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("batched", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := a.RenderAfterBatch(fc, ks); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// atSceneB is atScene for a benchmark.
func atSceneB(b *testing.B) *n3dsadapter.Adapter {
	b.Helper()
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		b.Skip("Captain Toad image not present")
	}
	a, err := n3dsadapter.New(imagePath)
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	if err := a.LoadStateFile(statePath); err != nil {
		b.Fatalf("LoadStateFile: %v", err)
	}
	return a
}

// The scrubber walks the commands that WROTE, so what the capture calls a writer had better
// be a command that draws. On this machine a "command" is a single PICA register write, and
// the only ones that put a fragment anywhere are the two draw triggers — so the writer list
// is a small subset of a very large stream, and every member of it is a DRAW.
func TestWritersAreTheDrawCommands(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	var fc *debug.FrameCapture
	for i := 0; i < 8; i++ {
		var err error
		if fc, err = a.StepFrame(false); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
		if fc.Drawn() {
			break
		}
	}
	if !fc.Drawn() {
		t.Fatal("no drawn frame in 8 steps")
	}
	if fc.Writers == nil {
		t.Fatal("a machine that reports every fragment reported no writer list")
	}
	if len(fc.Writers) == 0 {
		t.Fatal("a frame that drew a picture wrote nothing")
	}

	last := -1
	for _, k := range fc.Writers {
		if k <= last {
			t.Fatalf("writers are not in execution order: %d after %d", k, last)
		}
		last = k
		if k < 0 || k >= len(fc.Commands) {
			t.Fatalf("writer %d is not a command of this frame (%d commands)", k, len(fc.Commands))
		}
		if n := fc.Commands[k].Name; n != "DRAWARRAYS" && n != "DRAWELEMENTS" {
			t.Errorf("command %d (%s) is listed as writing pixels, but only a draw can", k, n)
		}
	}

	// The point of the whole exercise: the scrubber's stops are a tiny fraction of the
	// stream it used to walk.
	t.Logf("%d commands, %d of them wrote (%.2f%%)",
		len(fc.Commands), len(fc.Writers), 100*float64(len(fc.Writers))/float64(len(fc.Commands)))
	if len(fc.Writers) >= len(fc.Commands)/10 {
		t.Errorf("%d of %d commands wrote: that is not the machine we think it is", len(fc.Writers), len(fc.Commands))
	}
}
