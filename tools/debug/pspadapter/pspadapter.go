// Package pspadapter implements a debug.Target for the PlayStation Portable oracle.
//
// It is the fourth adapter, and the first whose machine draws a frame into more than
// one buffer as a matter of course. The N64 and the PSX each have one place a frame
// lands; the 3DS has two screens and taught the debugger to keep provenance per render
// target. The PSP needs that lesson and no more: Burnout renders its reflections and
// its downsampled bloom target into buffers the screen never shows, and a frame's
// provenance therefore has to be gathered per target and only then projected onto the
// buffer the game actually put on screen.
//
// Two things about this machine are worth stating plainly, because both are places a
// frame debugger can quietly lie:
//
//   - A "command" here is one GE command word, not one draw. The PSP is a
//     register-write machine like the PICA200, not a packet machine like the RDP, and
//     the writes that configure a draw are as much a part of the frame as the draw is.
//     PRIM is where the pixels happen; the two hundred words above it are why they look
//     like that.
//
//   - The GE runs synchronously inside sceGeListEnQueue, so the whole frame is drawn
//     without the CPU retiring an instruction. That is why the machine stops on a flag
//     the run loop reads (StopRequested) rather than on anything the CPU does, and why
//     a replay halted mid-list leaves a machine the game would not recognise — which is
//     fine for a scratch machine restored before every replay, and fine for nothing
//     else.
package pspadapter

import (
	"crypto/md5"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"retroreverse.com/tools/cpu/allegrex"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/psp"
)

// The debugger opens a PSP game through the registry. The boot executable's path inside
// the UMD is the one thing that is a property of the game rather than the machine, and
// both titles here use the standard one — so it has a default, and a game that differs
// says so in its debug.json rather than in this file.
func init() {
	debug.Register("psp", func(s debug.OpenSpec) (debug.Target, error) {
		var opts Options
		opts.Exe = s.Get("exe")
		return New(s.Image, opts)
	})
}

// runBudget bounds a single StepFrame or replay. A frame is a few million instructions
// and a boot to the first drawn frame is tens of millions, so this leaves headroom while
// still catching a machine that has wedged.
const runBudget = 400_000_000

// defaultExe is where a PSP game's boot executable lives on the UMD.
const defaultExe = "PSP_GAME/SYSDIR/EBOOT.BIN"

// The display is a fixed 480x272; off-screen targets vary, but the frame the player sees
// does not.
const dispW, dispH = 480, 272

// allegrexRegNames are the o32 ABI names of the 32 general registers, in index order.
var allegrexRegNames = []string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

// Options carries what the platform cannot work out for itself.
type Options struct {
	// Exe is the boot executable's path inside the UMD. Empty means the standard one.
	Exe string
}

// Adapter drives a PSP oracle as a debug.Target.
type Adapter struct {
	imagePath string
	exe       string
	hash      string

	live    *psp.Machine
	scratch *psp.Machine // reused across RenderAfter; its state is always disposable

	// The image is reopened per machine: the module loader reads from it, and a scratch
	// machine needs its own.
	img *psp.Image

	// frame is the render target the last capture's frame was drawn into — latched at
	// the flip, and reused by every replay of that capture. Re-deriving it from the
	// replayed machine would let the frame change identity mid-scrub, on exactly the
	// command where the game happens to rebind its buffer.
	frame target

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64
	stop      debug.StopReason
}

// target names one render target: where the GE was drawing, how wide, in what format.
// Provenance is gathered per target because a frame is drawn into several of them.
type target struct {
	addr, stride, format uint32
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

// snap wraps a PSP in-memory savestate as an opaque debug.Snapshot. psp.SaveState deep-
// copies RAM, VRAM, the scratchpad, the kernel objects and the GE register file, so a
// snapshot outlives the machine that made it and can be restored repeatedly — which is
// what replay needs.
type snap struct{ ms *psp.MachineState }

func (snap) Platform() string { return "psp" }

// New opens a UMD and boots its executable, the way the PSP bootoracle does.
func New(imagePath string, opts Options) (*Adapter, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	a := &Adapter{
		imagePath: imagePath,
		exe:       opts.Exe,
		hash:      fmt.Sprintf("%x", md5.Sum(data)),
	}
	if a.exe == "" {
		a.exe = defaultExe
	}
	a.img, err = psp.OpenImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("opening the UMD: %w", err)
	}
	live, err := a.build()
	if err != nil {
		return nil, err
	}
	a.live = live
	a.live.SetProfile(true)
	return a, nil
}

// build makes a machine from the image, exactly as the oracle does. Both the live
// machine and the replay scratch come from here, so a replay starts from the same
// machine the capture was taken on.
func (a *Adapter) build() (*psp.Machine, error) {
	mod, err := a.img.LoadExecutable(a.exe)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", a.exe, err)
	}
	m := psp.NewMachine()
	m.SetImageHash(a.hash)
	m.SetVolume(a.img.Volume)
	if err := m.LoadModule(mod); err != nil {
		return nil, fmt.Errorf("loading the module: %w", err)
	}
	return m, nil
}

func (a *Adapter) Platform() string { return "psp" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the live machine, for the wiring the generic interface does not cover
// (scripted pad input).
func (a *Adapter) Machine() *psp.Machine { return a.live }

func (a *Adapter) Close() error {
	if a.img != nil {
		return a.img.Close()
	}
	return nil
}

func (a *Adapter) Snapshot() debug.Snapshot { ms := a.live.SaveState(); return snap{ms: &ms} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	sn, ok := s.(snap)
	if !ok {
		return fmt.Errorf("pspadapter: cannot restore a %q snapshot", platformOf(s))
	}
	if err := a.live.LoadState(*sn.ms); err != nil {
		return err
	}
	a.frame = target{}
	a.live.SetProfile(true) // re-arm: the first frame after a load is a whole frame
	return nil
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }

func (a *Adapter) LoadStateFile(path string) error {
	if err := a.live.LoadStateFile(path); err != nil {
		return err
	}
	a.frame = target{}
	a.live.SetProfile(true)
	return nil
}

// ---- frames ----

// StepFrame advances to the next flip — sceDisplaySetFrameBuf, the only moment this
// machine says "a frame is finished" — capturing the GE command stream and per-pixel
// provenance along the way.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	ms := a.live.SaveState()
	fc := &debug.FrameCapture{Start: snap{ms: &ms}, Width: dispW, Height: dispH}

	// Provenance is gathered per render target, because a frame is drawn into several
	// of them and only one of them is the picture. Keyed by packed (y<<16|x) while the
	// frame draws, and laid out row-major once we know which target the game showed.
	prov := map[target]map[uint32]int32{}
	var over map[target]map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[target]map[uint32][]debug.PixelWrite{}
	}
	drawn := map[target]int{} // fragments actually stored, per target — the main-pass vote

	cmd := -1            // the command currently drawing; -1 before the first
	var cur target       // the target it is drawing into
	var presented target // what the flip put on screen

	a.live.OnGeCmd = func(w uint32) {
		cmd = len(fc.Commands)
		op := w >> 24
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index:   cmd,
			Name:    psp.GeCmdName(op),
			Op:      op,
			Words:   []uint64{uint64(w)},
			Decoded: fmt.Sprintf("%s = 0x%06X", psp.GeCmdName(op), w&0xFFFFFF),
		})
		// The hook fires before the word executes, so the register file still holds the
		// target this command will draw into — which is what we want to attribute its
		// pixels to.
		if addr, stride, format, ok := a.live.RenderTarget(); ok {
			cur = target{addr, stride, format}
		}
	}
	a.live.OnPixel = func(x, y uint32, ev psp.PixelEvent) {
		if cmd < 0 {
			return
		}
		key := y<<16 | (x & 0xFFFF)
		if ev.Drawn {
			if prov[cur] == nil {
				prov[cur] = map[uint32]int32{}
			}
			prov[cur][key] = int32(cmd)
			drawn[cur]++
		}
		if over != nil {
			if over[cur] == nil {
				over[cur] = map[uint32][]debug.PixelWrite{}
			}
			over[cur][key] = append(over[cur][key], debug.PixelWrite{
				CmdIndex: cmd, R: ev.R, G: ev.G, B: ev.B, A: ev.A,
				Rejected: !ev.Drawn,
			})
		}
	}
	a.live.OnDisplay = func(m *psp.Machine) {
		// The flip names the buffer the player is about to look at. That — not whatever
		// the GE was last bound to — is the frame.
		addr, stride, format := m.Scanout()
		presented = target{addr, stride, format}
		m.StopRequested = true
	}

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later run
	// is not paying for a census nobody is reading.
	a.live.OnGeCmd, a.live.OnPixel, a.live.OnDisplay = nil, nil, nil

	a.frame = a.pickFrame(presented, prov, drawn)
	if a.frame.addr != 0 {
		fc.Prov = flatten(prov[a.frame], dispW, dispH)
		if over != nil {
			fc.Overdraw = flattenOver(over[a.frame], dispW, dispH)
		}
	}
	return fc, nil
}

// pickFrame decides which render target the frame is.
//
// Normally the flip says so, and the flip is right. But a game that draws into a buffer
// and shows it on the NEXT frame — or that has not flipped yet at all, early in a boot —
// would leave the debugger displaying a buffer this frame never touched, and a scrubber
// whose picture never changes. So when the presented buffer received no pixels, fall
// back to the target that took the most: the main pass is the one that did the work.
func (a *Adapter) pickFrame(presented target, prov map[target]map[uint32]int32, drawn map[target]int) target {
	if presented.addr != 0 && drawn[presented] > 0 {
		return presented
	}
	best, bestN := target{}, 0
	for t, n := range drawn {
		if n > bestN {
			best, bestN = t, n
		}
	}
	if bestN > 0 {
		return best
	}
	return presented // nothing was drawn at all; show what is on screen
}

// flatten lays a target's provenance out row-major over the display.
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
	a.live.OnGeCmd, a.live.OnPixel = nil, nil
	a.live.OnDisplay = func(m *psp.Machine) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnDisplay = nil
	a.frame = target{}
	return nil
}

// RenderAfter replays the frame in a scratch machine and returns the render target
// exactly after GE command word k executed.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("pspadapter: capture holds a %q snapshot, not psp", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("pspadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc, err := a.scratchMachine()
	if err != nil {
		return nil, err
	}
	if err := sc.LoadState(*s.ms); err != nil {
		return nil, err
	}
	sc.OnGeCmd, sc.OnPixel, sc.OnDisplay = nil, nil, nil
	sc.RunStopAfterGeCommand(k+1, runBudget) // 1-based count → stop after 0-based index k

	// The target the CAPTURE decided on, not one re-derived from this replay: mid-frame
	// the game has not flipped yet, and asking the half-run machine which buffer it is
	// showing would answer with the previous frame's.
	t := a.frame
	if t.addr == 0 {
		if addr, stride, format, ok := sc.RenderTarget(); ok {
			t = target{addr, stride, format}
		} else {
			return sc.Framebuffer(), nil
		}
	}
	return sc.RenderSurface(t.addr, t.stride, t.format, dispW, dispH)
}

func (a *Adapter) scratchMachine() (*psp.Machine, error) {
	if a.scratch == nil {
		m, err := a.build()
		if err != nil {
			return nil, err
		}
		a.scratch = m
	}
	return a.scratch, nil
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
	vals := make([]uint64, 32)
	for i := 0; i < 32; i++ {
		vals[i] = uint64(c.R[i])
	}
	return debug.CPUReg{
		PC:    uint64(c.PC),
		Names: allegrexRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"hi": uint64(c.HI), "lo": uint64(c.LO),
			"Status": uint64(c.COP0[12]), "Cause": uint64(c.COP0[13]), "EPC": uint64(c.COP0[14]),
			"steps": c.Steps,
		},
	}
}

func (a *Adapter) Display() (*image.RGBA, error) { return a.live.Framebuffer(), nil }

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

// Surfaces. The render target and the scanout are different buffers on this machine and
// the difference is the whole point: the GE draws into the one its FBP names, and the
// screen shows the one sceDisplaySetFrameBuf named. The texture surface is the one that
// earns its keep — point it at the address a texture unit is configured with and you see
// whether what it will sample is the texture the game meant to upload.
//
// The free surfaces address VRAM the way the machine does: View.Addr is a CPU address.
func (a *Adapter) Surfaces() []debug.Surface {
	return []debug.Surface{
		{ID: "scanout", Name: "Display", W: dispW, H: dispH},
		{ID: "drawtarget", Name: "Render target (what the GE is drawing into)", W: dispW, H: dispH},
		{ID: "depth", Name: "Depth buffer", W: dispW, H: dispH},
		{ID: "target", Name: "VRAM as a render target", Free: true, Formats: []string{"rgb565", "rgba5551", "rgba4444", "rgba8888"}},
		{ID: "texture", Name: "VRAM/RAM as a texture", Free: true, Formats: psp.TextureFormats()},
	}
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "scanout":
		return a.live.Framebuffer(), nil

	case "drawtarget":
		addr, stride, format, ok := a.live.RenderTarget()
		if !ok {
			return nil, fmt.Errorf("pspadapter: the GE has not bound a render target yet: %w", debug.ErrUnsupported)
		}
		return a.live.RenderSurface(addr, stride, format, dispW, dispH)

	case "depth":
		addr, stride, ok := a.live.DepthTarget()
		if !ok {
			return nil, fmt.Errorf("pspadapter: the GE has no depth buffer bound: %w", debug.ErrUnsupported)
		}
		img, err := a.live.RenderDepth(addr, stride, dispW, dispH)
		if err != nil {
			return nil, err
		}
		return nrgbaToRGBA(img), nil

	case "target":
		format, ok := displayFormat(v.Format)
		if !ok {
			return nil, fmt.Errorf("pspadapter: %q is not a display pixel format", v.Format)
		}
		return a.live.RenderSurface(v.Addr, uint32(v.Stride), format, v.W, v.H)

	case "texture":
		format, ok := psp.TextureFormat(v.Format)
		if !ok {
			return nil, fmt.Errorf("pspadapter: %q is not a GE texture format", v.Format)
		}
		// Palette != 0 is the caller asking for the swizzled layout; the PSP's swizzle is
		// a property of the texture, not of an address, and there is nowhere else in the
		// generic View to say so.
		img, err := a.live.RenderTexture(v.Addr, format, uint32(v.W), uint32(v.H), v.Palette != 0)
		if err != nil {
			return nil, err
		}
		return nrgbaToRGBA(img), nil
	}
	return nil, fmt.Errorf("pspadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// displayFormat maps a surface format name to the PSP's display pixel format.
func displayFormat(name string) (uint32, bool) {
	switch name {
	case "rgb565":
		return 0, true
	case "rgba5551":
		return 1, true
	case "rgba4444":
		return 2, true
	case "rgba8888":
		return 3, true
	}
	return 0, false
}

func nrgbaToRGBA(in *image.NRGBA) *image.RGBA {
	out := image.NewRGBA(in.Rect)
	for y := in.Rect.Min.Y; y < in.Rect.Max.Y; y++ {
		for x := in.Rect.Min.X; x < in.Rect.Max.X; x++ {
			out.Set(x, y, in.At(x, y))
		}
	}
	return out
}

// ---- the UMD ----

func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	if path == "" {
		path = "/"
	}
	entries, err := a.img.Volume.ReadDir(path)
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

func (a *Adapter) ReadFile(path string) ([]byte, error) { return a.img.Volume.ReadFile(path) }

// FileAt names the file containing a UMD sector — so a read-watch can say which file the
// game is streaming, rather than which block number.
func (a *Adapter) FileAt(sector int64) (debug.FileEntry, int64, bool) {
	var found debug.FileEntry
	var within int64
	ok := false
	a.img.Volume.Walk(func(e psp.Entry) error {
		if e.IsDir || e.Size == 0 {
			return nil
		}
		blocks := int64((e.Size + 2047) / 2048)
		if sector >= int64(e.Block) && sector < int64(e.Block)+blocks {
			found = debug.FileEntry{Path: e.Path, Name: e.Name, Size: int64(e.Size), Offset: int64(e.Block)}
			within = sector - int64(e.Block)
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
	// Sitting on a breakpoint, run one instruction first — otherwise the run stops on
	// the breakpoint it is already standing on and never moves.
	if a.atBreakpoint() {
		if _, err := a.runSlice(1, "steps"); err != nil {
			return debug.StopReason{}, err
		}
	}
	return a.runSlice(budget, "budget")
}

func (a *Adapter) atBreakpoint() bool {
	pc := uint64(a.live.CPU.PC)
	for _, b := range a.bps {
		if b == pc {
			return true
		}
	}
	return false
}

// runSlice runs the machine and works out why it stopped. A breaking watch and a
// breakpoint both report themselves through a.stop, because from Run's point of view a
// hook merely asked to stop.
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
	sr.PC = uint64(a.live.CPU.PC)
	sr.Steps = res.Steps
	return sr, nil
}

// Disasm decodes n instructions at an address.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*4
		var b [4]byte
		for j := 0; j < 4; j++ {
			b[j] = a.live.Read(va + uint32(j))
		}
		in := allegrex.Decode(b[:], va)
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
// enforced from the per-instruction hook — which is installed only while there is a
// breakpoint to check, so a target with none pays nothing.
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
	a.live.OnStep = func(m *psp.Machine, pc uint32) {
		if set[pc] {
			m.StopRequested = true
			a.stop = debug.StopReason{Kind: "breakpoint", PC: uint64(pc)}
		}
	}
}

// ---- watches ----
//
// Like the other machines here, this one has one write window and one read window, so a
// target holds one watch of each kind. A second of the same kind is an error rather than
// a silent replacement.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("pspadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("pspadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

// ResumeArgs is the command line that reopens this UMD at a saved state. Both PSP titles
// have their own bootoracle and both spell the flags the same way, so there is no
// ResumeDirer here: the game's own extract/ is where this runs.
func (a *Adapter) ResumeArgs(statePath string) []string {
	args := []string{"go", "run", "./cmd/bootoracle", "-image", a.imagePath, "-loadstate", statePath}
	if a.exe != defaultExe {
		args = append(args, "-exe", a.exe)
	}
	return args
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
