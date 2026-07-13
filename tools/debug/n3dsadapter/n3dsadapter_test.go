package n3dsadapter_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n3dsadapter"
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
		debug.CapRegions,
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

// TestScreenDecodeMatchesTheTransfer is the cross-check on the whole mapping. The
// debugger decodes a screen straight from the tiled render target (so a replayed
// frame can fill in draw by draw); the machine's own Framebuffer decodes it from the
// linear buffer the DisplayTransfer wrote. Those are two independent paths through
// the crop, the row offset and the rotation, and they must land on the same picture —
// if they disagree, the provenance the debugger maps through the first path is
// pointing at pixels the player never saw.
func TestScreenDecodeMatchesTheTransfer(t *testing.T) {
	a := atScene(t)
	defer a.Close()

	if _, err := a.StepFrame(false); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	m := a.Machine()
	for _, name := range []string{"top", "bottom"} {
		g, ok := m.ScreenGeom(name)
		if !ok {
			t.Fatalf("%s screen has no transfer geometry", name)
		}
		ours := m.ScreenImage(g)
		theirs := m.Framebuffer(name)
		if ours.Rect != theirs.Rect {
			t.Fatalf("%s: decoded %v, the transfer's framebuffer is %v", name, ours.Rect, theirs.Rect)
		}
		diff := 0
		for i := range ours.Pix {
			if ours.Pix[i] != theirs.Pix[i] {
				diff++
			}
		}
		if diff != 0 {
			t.Errorf("%s: %d of %d bytes differ between the render-target decode and the presented framebuffer",
				name, diff, len(ours.Pix))
		}
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
