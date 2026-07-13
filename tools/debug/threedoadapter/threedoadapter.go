// Package threedoadapter implements a debug.Target for the 3DO oracle.
//
// (The package cannot be called 3doadapter: a Go identifier may not begin with a digit,
// which is the same reason the platform package is threedo and the 3DS's is n3ds. The
// platform STRING is "3do", and that is what the registry, the game slug and the page
// all use.)
//
// This is the fifth adapter and the strangest machine of the five. There is no display
// list, no register file, and no GPU to speak of: the 3DO's cel engine is handed a CCB —
// a Cel Control Block — and walks a linked list of them, blending each cel's pixels into
// a bitmap through a per-pixel processor configured by the CCB itself. So:
//
//   - A "command" is one cel. The CCB is the instruction, and the words the debugger
//     shows are the CCB's own, read back out of the game's memory.
//
//   - There is no rejection to report. The cel engine has no depth test and no alpha
//     test, so a pixel it produces is a pixel it wrote — as on the PSX, the overdraw pane
//     will never show a rejected write here, because there are none to show.
//
//   - Provenance is gathered per BITMAP. Need for Speed renders its rear-view mirror into
//     an off-screen buffer and then draws that buffer back as a cel, so a frame genuinely
//     lands in more than one place, and only one of them is the picture.
//
// The frame boundary is DisplayScreen, the folio call that puts a buffer on screen. The
// buffer it names is the FRONT one; the game has been drawing into the other. A debugger
// that showed the front buffer mid-frame would show a picture that never changes.
package threedoadapter

import (
	"crypto/md5"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/arm60"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/threedo"
)

// The debugger opens a 3DO game through the registry. Two things are properties of the
// GAME rather than the machine, and both arrive in the profile (games/<slug>/debug.json)
// rather than as constants here: which AIF on the disc is the program, and where the
// game keeps the field counter the OS's VBL interrupt would advance. This package knows
// about the 3DO, not about any game on it.
func init() {
	debug.Register("3do", func(s debug.OpenSpec) (debug.Target, error) {
		opts := Options{Prog: s.Get("prog")}
		if v := s.Get("vblmirror"); v != "" {
			n, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(v), "0x"), 16, 32)
			if err != nil {
				return nil, fmt.Errorf("3do profile: bad vblmirror %q: %w", v, err)
			}
			opts.VBLMirror = uint32(n)
		}
		if v := s.Get("stall"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("3do profile: bad stall %q: %w", v, err)
			}
			opts.StallTolerance = n
		}
		return New(s.Image, opts)
	})
}

// runBudget bounds a single StepFrame or replay. A frame is a few hundred thousand
// instructions and a boot to the first drawn frame is tens of millions, so this leaves
// headroom while still catching a machine that has wedged.
const runBudget = 400_000_000

// defaultProg is the AIF a 3DO disc boots.
const defaultProg = "LaunchMe"

// arm60RegNames are the ARM register names, in index order.
var arm60RegNames = []string{
	"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7",
	"r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc",
}

// Options carries what the platform cannot work out for itself.
type Options struct {
	// Prog is the AIF program on the disc. Empty means LaunchMe.
	Prog string

	// VBLMirror is the game's own elapsed-field counter, which the OS's VBL interrupt
	// would advance and which the HLE's virtual VBL manager writes instead. The address
	// is the game's, so the game supplies it. Zero means the game registers it itself
	// (the graphics folio's RegisterVBLCounter).
	VBLMirror uint32

	// StallTolerance scales the run loop's deadlock guard. A game whose main loop has
	// settled — which is exactly what a debugger sits and stares at — re-executes only
	// code it has already run, and the guard has to be told that is not a deadlock.
	StallTolerance int
}

// Adapter drives a 3DO oracle as a debug.Target.
type Adapter struct {
	imagePath string
	hash      string
	opts      Options
	vol       *threedo.Volume
	aif       *threedo.AIF

	live    *threedo.Machine
	scratch *threedo.Machine // reused across RenderAfter; its state is always disposable

	// frame is the bitmap the last capture's frame was drawn into, latched at the flip
	// and reused by every replay of that capture. Re-deriving it from the replayed
	// machine would let the frame change identity mid-scrub.
	frame bitmap

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64
	stop      debug.StopReason
}

// bitmap names one drawable buffer. Provenance is gathered per bitmap because a frame is
// drawn into more than one of them.
type bitmap struct {
	buf  uint32
	w, h int
}

// The capabilities this target backs.
var (
	_ debug.Target         = (*Adapter)(nil)
	_ debug.FrameStepper   = (*Adapter)(nil)
	_ debug.FastStepper    = (*Adapter)(nil)
	_ debug.FrameReplayer  = (*Adapter)(nil)
	_ debug.CodeStepper    = (*Adapter)(nil)
	_ debug.Breakpointer   = (*Adapter)(nil)
	_ debug.Disassembler   = (*Adapter)(nil)
	_ debug.Watcher        = (*Adapter)(nil)
	_ debug.Surfacer       = (*Adapter)(nil)
	_ debug.FileLister     = (*Adapter)(nil)
	_ debug.FileAttributer = (*Adapter)(nil)
	_ debug.StateFiler     = (*Adapter)(nil)
	_ debug.Resumer        = (*Adapter)(nil)
	_ debug.MemoryMapper   = (*Adapter)(nil)
	_ debug.Profiler       = (*Adapter)(nil)
)

// snap wraps a 3DO in-memory savestate as an opaque debug.Snapshot. threedo.SaveState
// deep-copies DRAM, VRAM, the heaps and the whole Portfolio HLE — the item table, the
// task list, the open streams — so a snapshot outlives the machine that made it and can
// be restored repeatedly, which is what replay needs.
type snap struct{ ms *threedo.MachineState }

func (snap) Platform() string { return "3do" }

// New opens a disc and loads its boot program, the way the 3DO bootoracle does.
func New(imagePath string, opts Options) (*Adapter, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	vol, err := threedo.Open(data)
	if err != nil {
		return nil, fmt.Errorf("opening the disc: %w", err)
	}
	a := &Adapter{
		imagePath: imagePath,
		hash:      fmt.Sprintf("%x", md5.Sum(data)),
		opts:      opts,
		vol:       vol,
	}
	if a.opts.Prog == "" {
		a.opts.Prog = defaultProg
	}
	if a.opts.StallTolerance == 0 {
		// A debugger sits on a settled frame loop by definition, so the deadlock guard
		// needs more rope here than a one-shot oracle run does.
		a.opts.StallTolerance = 8
	}
	prog, err := vol.ReadFile(a.opts.Prog)
	if err != nil {
		return nil, fmt.Errorf("reading %s off the disc: %w", a.opts.Prog, err)
	}
	a.aif, err = threedo.ParseAIF(prog)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", a.opts.Prog, err)
	}
	a.live = a.build()
	a.live.SetProfile(true)
	return a, nil
}

// build makes a machine from the disc, exactly as the oracle does. Both the live machine
// and the replay scratch come from here, so a replay starts from the same machine the
// capture was taken on.
func (a *Adapter) build() *threedo.Machine {
	m := threedo.NewMachine()
	m.SetImageHash(a.hash)
	m.StallTolerance = a.opts.StallTolerance
	// The movie player needs the audio folio and the DataStreamer, which the HLE does
	// not model yet; letting the game open a .stream takes it into a player that runs on
	// stubbed answers and corrupts itself. Failing them takes the game's own
	// missing-file path, which is the same graceful fallback it ships for a scratched
	// disc — and lands it in the front end, which is what there is to debug.
	m.NoStreams = true
	m.SetVolume(a.vol)
	if a.opts.VBLMirror != 0 {
		m.SetVBLMirror(a.opts.VBLMirror)
	}
	m.LoadAIF(a.aif)
	return m
}

func (a *Adapter) Platform() string { return "3do" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the live machine, for the wiring the generic interface does not cover
// (scripted pad input).
func (a *Adapter) Machine() *threedo.Machine { return a.live }

func (a *Adapter) Close() error { return nil }

func (a *Adapter) Snapshot() debug.Snapshot { ms := a.live.SaveState(); return snap{ms: &ms} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	sn, ok := s.(snap)
	if !ok {
		return fmt.Errorf("threedoadapter: cannot restore a %q snapshot", platformOf(s))
	}
	if err := a.live.LoadState(*sn.ms); err != nil {
		return err
	}
	a.frame = bitmap{}
	a.live.SetProfile(true) // re-arm: the first frame after a load is a whole frame
	return nil
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }

func (a *Adapter) LoadStateFile(path string) error {
	if err := a.live.LoadStateFile(path); err != nil {
		return err
	}
	a.frame = bitmap{}
	a.live.SetProfile(true)
	return nil
}

// ---- frames ----

// StepFrame advances to the next DisplayScreen — the folio call that puts a buffer on
// screen, and this machine's only statement that a frame is finished — capturing the
// cels it drew and per-pixel provenance along the way.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	ms := a.live.SaveState()
	fc := &debug.FrameCapture{Start: snap{ms: &ms}}

	// Provenance is gathered per bitmap, because a frame is drawn into more than one and
	// only one of them is the picture. Keyed by packed (y<<16|x) while the frame draws.
	prov := map[bitmap]map[uint32]int32{}
	var over map[bitmap]map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[bitmap]map[uint32][]debug.PixelWrite{}
	}
	drawn := map[bitmap]int{}

	cmd := -1
	var cur bitmap
	var presented uint32

	a.live.OnCel = func(c threedo.CelDraw) {
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index:   cmd,
			Name:    c.Name(),
			Op:      c.Flags,
			Words:   c.Words(),
			Decoded: c.Decoded(),
		})
		cur = bitmap{c.Bitmap, c.BitmapW, c.BitmapH}
	}
	a.live.OnPixel = func(x, y uint32, ev threedo.PixelEvent) {
		if cmd < 0 {
			return
		}
		key := y<<16 | (x & 0xFFFF)
		if prov[cur] == nil {
			prov[cur] = map[uint32]int32{}
		}
		prov[cur][key] = int32(cmd)
		drawn[cur]++
		if over != nil {
			if over[cur] == nil {
				over[cur] = map[uint32][]debug.PixelWrite{}
			}
			over[cur][key] = append(over[cur][key], debug.PixelWrite{
				CmdIndex: cmd, R: ev.R, G: ev.G, B: ev.B, A: ev.A,
			})
		}
	}
	a.live.OnDisplay = func(m *threedo.Machine, frame uint64, buf uint32) {
		presented = buf
		m.StopRequested = true
	}

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later run
	// is not paying for a census nobody is reading.
	a.live.OnCel, a.live.OnPixel, a.live.OnDisplay = nil, nil, nil

	a.frame = a.pickFrame(presented, drawn)
	fc.Width, fc.Height = a.frame.w, a.frame.h
	if fc.Width > 0 && fc.Height > 0 {
		fc.Prov = flatten(prov[a.frame], fc.Width, fc.Height)
		if over != nil {
			fc.Overdraw = flattenOver(over[a.frame], fc.Width, fc.Height)
		}
	}
	return fc, nil
}

// pickFrame decides which bitmap the frame is.
//
// DisplayScreen names the buffer going ON screen, and with double buffering that is the
// one the game has just finished drawing — so if cels landed in it, it is the frame. But
// a game that draws ahead, or that has not flipped yet, would leave the debugger showing
// a buffer this frame never touched. So when the presented buffer took no pixels, fall
// back to the bitmap that took the most: the main pass is the one that did the work.
func (a *Adapter) pickFrame(presented uint32, drawn map[bitmap]int) bitmap {
	if presented != 0 {
		for bm, n := range drawn {
			if bm.buf == presented && n > 0 {
				return bm
			}
		}
	}
	best, bestN := bitmap{}, 0
	for bm, n := range drawn {
		if n > bestN {
			best, bestN = bm, n
		}
	}
	if bestN > 0 {
		return best
	}
	// Nothing was drawn at all: show whatever is on screen.
	if buf, w, h, ok := a.live.DrawTarget(); ok {
		if presented != 0 {
			return bitmap{presented, w, h}
		}
		return bitmap{buf, w, h}
	}
	return bitmap{}
}

func flatten(pv map[uint32]int32, w, h int) []int32 {
	if len(pv) == 0 {
		return nil
	}
	out := make([]int32, w*h)
	for i := range out {
		out[i] = -1
	}
	for key, c := range pv {
		x, y := int(key&0xFFFF), int(key>>16)
		if x >= 0 && y >= 0 && x < w && y < h {
			out[y*w+x] = c
		}
	}
	return out
}

func flattenOver(ov map[uint32][]debug.PixelWrite, w, h int) map[int][]debug.PixelWrite {
	if len(ov) == 0 {
		return nil
	}
	out := make(map[int][]debug.PixelWrite, len(ov))
	for key, writes := range ov {
		x, y := int(key&0xFFFF), int(key>>16)
		if x >= 0 && y >= 0 && x < w && y < h {
			out[y*w+x] = writes
		}
	}
	return out
}

// StepFast advances one frame capturing nothing — how the debugger fast-forwards.
func (a *Adapter) StepFast() error {
	a.live.OnCel, a.live.OnPixel = nil, nil
	a.live.OnDisplay = func(m *threedo.Machine, frame uint64, buf uint32) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnDisplay = nil
	a.frame = bitmap{}
	return nil
}

// RenderAfter replays the frame in a scratch machine and returns the bitmap exactly after
// cel k was drawn.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("threedoadapter: capture holds a %q snapshot, not 3do", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("threedoadapter: cel %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc := a.scratchMachine()
	if err := sc.LoadState(*s.ms); err != nil {
		return nil, err
	}
	sc.OnCel, sc.OnPixel, sc.OnDisplay = nil, nil, nil
	sc.RunStopAfterCel(k+1, runBudget) // 1-based count → stop after 0-based index k

	// The bitmap the CAPTURE decided on, not one re-derived from this replay: mid-frame
	// the game has not flipped yet, and asking the half-run machine which buffer it is
	// showing would answer with the previous frame's.
	bm := a.frame
	if bm.buf == 0 {
		buf, w, h, ok := sc.DrawTarget()
		if !ok {
			return nil, fmt.Errorf("threedoadapter: the game has not made a screen yet: %w", debug.ErrUnsupported)
		}
		bm = bitmap{buf, w, h}
	}
	return sc.RenderBitmap(bm.buf, bm.w, bm.h)
}

func (a *Adapter) scratchMachine() *threedo.Machine {
	if a.scratch == nil {
		a.scratch = a.build()
	}
	return a.scratch
}

// FrameProfile reports where the last stepped frame's time went.
func (a *Adapter) FrameProfile() debug.FrameProfile {
	p := a.live.FrameProfile()
	out := debug.FrameProfile{TotalMs: p.TotalMs}
	for _, b := range p.Buckets {
		out.Buckets = append(out.Buckets, debug.ProfileBucket{Name: b.Name, Millis: b.Millis, Count: b.Count})
	}
	for _, c := range p.Counters {
		out.Counters = append(out.Counters, debug.ProfileCounter{Name: c.Name, Value: c.Value})
	}
	return out
}

// ---- the machine at rest ----

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 16)
	for i := 0; i < 16; i++ {
		vals[i] = uint64(c.Reg(uint32(i)))
	}
	return debug.CPUReg{
		PC:    uint64(c.Reg(15)),
		Names: arm60RegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"CPSR": uint64(c.CPSR()), "SPSR": uint64(c.SPSR()),
			"mode": uint64(c.Mode), "instrs": c.Instrs,
		},
	}
}

// Display is the buffer on screen — the front one, which DisplayScreen last named.
func (a *Adapter) Display() (*image.RGBA, error) {
	buf := a.live.DisplayBuffer()
	if buf == 0 {
		// Nothing has been shown yet; show the draw target rather than nothing.
		b, w, h, ok := a.live.DrawTarget()
		if !ok {
			return image.NewRGBA(image.Rect(0, 0, 320, 240)), nil
		}
		return a.live.RenderBitmap(b, w, h)
	}
	w, h := a.screenSize()
	return a.live.RenderBitmap(buf, w, h)
}

// screenSize is the displayed bitmap's geometry. The 3DO's is 320x240 and the game's
// bitmaps say so, but a game that made none yet still has to be shown something.
func (a *Adapter) screenSize() (int, int) {
	if _, w, h, ok := a.live.DrawTarget(); ok && w > 0 && h > 0 {
		return w, h
	}
	return 320, 240
}

func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a.live.Read(addr + uint32(i))
	}
	return out
}

func (a *Adapter) Regions() []debug.Region {
	var out []debug.Region
	for _, r := range a.live.MemRegions() {
		out = append(out, debug.Region{Name: r.Name, Lo: r.Base, Hi: r.Base + r.Size})
	}
	return out
}

// Surfaces. The front buffer and the back buffer are different bitmaps and the difference
// is the whole point. "All of VRAM" is worth looking at as one image on this machine
// because everything the cel engine draws into lives there, and a free surface reads any
// rectangle of it as a bitmap — which is how you look at the rear-view mirror's
// off-screen buffer while the race is running.
func (a *Adapter) Surfaces() []debug.Surface {
	w, h := a.screenSize()
	out := []debug.Surface{
		{ID: "scanout", Name: "Display (front buffer)", W: w, H: h},
		{ID: "drawtarget", Name: "Draw target (back buffer)", W: w, H: h},
		{ID: "bitmap", Name: "VRAM as a bitmap", Free: true, Formats: []string{"rgb555"}},
	}
	return out
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "scanout":
		return a.Display()
	case "drawtarget":
		buf, w, h, ok := a.live.DrawTarget()
		if !ok {
			return nil, fmt.Errorf("threedoadapter: the game has not made a screen yet: %w", debug.ErrUnsupported)
		}
		return a.live.RenderBitmap(buf, w, h)
	case "bitmap":
		w, h := v.W, v.H
		if w == 0 || h == 0 {
			w, h = a.screenSize()
		}
		return a.live.RenderBitmap(v.Addr, w, h)
	}
	return nil, fmt.Errorf("threedoadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// ---- the disc ----

func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	if path == "" {
		path = "/"
	}
	entries, err := a.vol.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]debug.FileEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, debug.FileEntry{
			Path: e.Path, Name: e.Name, Dir: e.IsDir,
			Size: int64(e.Size), Offset: int64(e.Block),
		})
	}
	return out, nil
}

func (a *Adapter) ReadFile(path string) ([]byte, error) { return a.vol.ReadFile(path) }

// FileAt names the file containing a disc block — so a read-watch can say which file the
// game is streaming, rather than which block number.
func (a *Adapter) FileAt(block int64) (debug.FileEntry, int64, bool) {
	var found debug.FileEntry
	var within int64
	ok := false
	a.vol.Walk(func(e threedo.Entry) error {
		if e.IsDir || e.Size == 0 {
			return nil
		}
		if block >= int64(e.Block) && block < int64(e.Block)+int64(e.Blocks) {
			found = debug.FileEntry{Path: e.Path, Name: e.Name, Size: int64(e.Size), Offset: int64(e.Block)}
			within = block - int64(e.Block)
			ok = true
		}
		return nil
	})
	return found, within, ok
}

// ---- code ----

func (a *Adapter) StepInstr(n int) (debug.StopReason, error) {
	if n <= 0 {
		n = 1
	}
	return a.runSlice(uint64(n), "steps")
}

func (a *Adapter) Continue(budget uint64) (debug.StopReason, error) {
	if budget == 0 {
		budget = runBudget
	}
	if a.atBreakpoint() {
		if _, err := a.runSlice(1, "steps"); err != nil {
			return debug.StopReason{}, err
		}
	}
	return a.runSlice(budget, "budget")
}

func (a *Adapter) atBreakpoint() bool {
	pc := uint64(a.live.CPU.Reg(15))
	for _, b := range a.bps {
		if b == pc {
			return true
		}
	}
	return false
}

func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	res := a.live.Run(budget)
	sr := a.stop
	if sr.Kind == "" {
		switch {
		case a.live.CPU.Halted:
			sr = debug.StopReason{Kind: "halted", Note: a.live.CPU.HaltReason}
		case a.live.Halted:
			sr = debug.StopReason{Kind: "halted", Note: a.live.HaltReason}
		default:
			sr = debug.StopReason{Kind: exhausted, Note: res.Reason}
		}
	}
	sr.PC = uint64(res.PC)
	sr.Steps = res.Steps
	return sr, nil
}

// Disasm decodes n instructions at an address. The ARM60 is a fixed 4-byte instruction
// machine, and big-endian: the 3DO is the only big-endian target here.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*4
		var b [4]byte
		for j := 0; j < 4; j++ {
			b[j] = a.live.Read(va + uint32(j))
		}
		in := arm60.Decode(b[:], va)
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: b[:], Text: in.Text})
	}
	return out, nil
}

func (a *Adapter) instrAt(pc uint32) string {
	in, err := a.Disasm(uint64(pc), 1)
	if err != nil || len(in) == 0 {
		return ""
	}
	return in[0].Text
}

// Breakpoints. The machine has no breakpoint table of its own, so they are kept here and
// enforced from the per-instruction hook — installed only while there is a breakpoint to
// check, so a target with none pays nothing.
func (a *Adapter) SetBreakpoint(pc uint64) {
	for _, b := range a.bps {
		if b == pc {
			return
		}
	}
	a.bps = append(a.bps, pc)
	a.applyBreakpoints()
}

func (a *Adapter) ClearBreakpoint(pc uint64) {
	kept := a.bps[:0]
	for _, b := range a.bps {
		if b != pc {
			kept = append(kept, b)
		}
	}
	a.bps = kept
	a.applyBreakpoints()
}

func (a *Adapter) Breakpoints() []uint64 { return append([]uint64(nil), a.bps...) }

func (a *Adapter) applyBreakpoints() {
	if len(a.bps) == 0 {
		a.live.OnStep = nil
		return
	}
	set := make(map[uint32]bool, len(a.bps))
	for _, b := range a.bps {
		set[uint32(b)] = true
	}
	a.live.OnStep = func(m *threedo.Machine, pc uint32) {
		if set[pc] {
			m.StopRequested = true
			a.stop = debug.StopReason{Kind: "breakpoint", PC: uint64(pc)}
		}
	}
}

// ---- watches ----

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("threedoadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("threedoadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
		}
	}
	a.nextWatch++
	w.ID = a.nextWatch
	a.watches = append(a.watches, w)
	a.applyWatches()
	return w.ID, nil
}

func (a *Adapter) ClearWatch(id int) {
	for i, w := range a.watches {
		if w.ID == id {
			a.watches = append(a.watches[:i], a.watches[i+1:]...)
			break
		}
	}
	a.applyWatches()
}

func (a *Adapter) Watches() []debug.Watch { return append([]debug.Watch(nil), a.watches...) }

func (a *Adapter) OnWatchHit(sink func(debug.WatchHit)) {
	a.watchSink = sink
	a.applyWatches()
}

func (a *Adapter) applyWatches() {
	a.live.WatchLo, a.live.WatchHi, a.live.OnWrite = 0, 0, nil
	a.live.RWatchLo, a.live.RWatchHi, a.live.OnRead = 0, 0, nil
	for i := range a.watches {
		w := &a.watches[i]
		cb := func(addr, val, pc uint32) {
			w.Hits++
			if a.watchSink != nil {
				a.watchSink(debug.WatchHit{
					ID: w.ID, Kind: w.Kind, Addr: addr, Val: val, PC: pc,
					Instr: a.instrAt(pc),
				})
			}
			if w.Break {
				a.live.StopRequested = true
				a.stop = debug.StopReason{
					Kind: "watch", PC: uint64(pc),
					Note: fmt.Sprintf("watch %d (%s %08x)", w.ID, w.Kind, addr),
				}
			}
		}
		if w.Kind == "write" {
			a.live.WatchLo, a.live.WatchHi, a.live.OnWrite = w.Lo, w.Hi, cb
		} else {
			a.live.RWatchLo, a.live.RWatchHi, a.live.OnRead = w.Lo, w.Hi, cb
		}
	}
}

// ResumeArgs is the command line that reopens this disc at a saved state.
func (a *Adapter) ResumeArgs(statePath string) []string {
	args := []string{"go", "run", "./cmd/bootoracle", "-image", a.imagePath, "-loadstate", statePath}
	if a.opts.Prog != defaultProg {
		args = append(args, "-prog", a.opts.Prog)
	}
	if a.opts.VBLMirror != 0 {
		args = append(args, "-vblmirror", fmt.Sprintf("0x%X", a.opts.VBLMirror))
	}
	return args
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
