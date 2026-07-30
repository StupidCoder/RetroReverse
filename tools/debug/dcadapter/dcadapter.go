// Package dcadapter implements a debug.Target for the Sega Dreamcast oracle.
//
// It owns a live dc.Machine and, for command replay, a scratch machine it restores
// into. The machine already offered most of what the debugger needs (watch windows,
// an in-memory savestate that is deep both ways, the field-boundary OnDisplay hook);
// this package adds the thin translation, and the platform grew the render-walk
// hooks (OnPVRClear/OnPVRCmd/OnPVRPixel/OnRender, RenderStopAfter) for it.
//
// THE FRAME IS THE DRAW TARGET, WHICH IS ONE FLIP AHEAD OF THE SCANOUT. The PVR
// renders a whole recorded TA session at STARTRENDER into the buffer FB_W_SOF1
// names; the guest sees the render-done interrupt and flips FB_R_SOF1 onto it. So
// the capture's picture is RenderDrawTarget taken inside OnRender — after the last
// pixel, before the guest can flip — and Display() is RenderFB, the picture the TV
// is showing now. They are the same VRAM addressed through the same 32-bit-path
// interleave, one frame apart; comparing them catches a frame that rendered and was
// never flipped to.
//
// A COMMAND IS ONE TA PARAMETER of the frame's recorded stream — a header, a
// vertex, an END_OF_LIST — plus the background clear as command zero (it writes
// every pixel of the frame, and it is usually the answer to "why is this pixel
// black"). The render walk is a plain Go loop inside one deferred completion, so
// stopping after command k needs no sentinel panic: RenderStopAfter counts, the
// walk breaks, and the partial frame is simply what the write framebuffer holds.
//
// One honest absence: depth-rejected fragments are not reported (pvr.go says why —
// any edit inside the rasteriser's float loop moves the compiler's FMA contraction
// and with it the pinned pictures), so overdraw shows draws and alpha-rejects only.
package dcadapter

import (
	"encoding/binary"
	"fmt"
	"image"
	"path/filepath"

	"retroreverse.com/tools/cpu/sh4"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/dc"
)

// The debugger opens a Dreamcast game through the registry; importing this package
// for its side effect is enough.
func init() {
	debug.Register("dc", func(s debug.OpenSpec) (debug.Target, error) { return New(s.Image) })
}

// runBudget bounds a single StepFrame/StepFast/replay: ~60 fields at the machine's
// 3.33M instructions per field. A game that is rendering kicks STARTRENDER every
// frame, so this is generous headroom; a cold boot is ~3 billion instructions away
// (the warning screen draws at field ~900), which is why the debugger is meant to
// be pointed at a savestate rather than made to sit through the boot.
const runBudget = 200_000_000

// tapFields is how long one queued pad state stays live before the next replaces
// it. The oracle's -keys scripts settled on 3-field taps: the game's driver
// latches the pad once per field and edge-detects across fields, and its mainline
// occasionally overruns a field — a 1-field press can fall into exactly the field
// the mainline skipped and vanish. Three fields is the tap the game provably sees.
const tapFields = 3

// Adapter drives a Dreamcast oracle as a debug.Target.
type Adapter struct {
	imagePath string
	live      *dc.Machine
	scratch   *dc.Machine // reused across RenderAfter; its state is always disposable

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       map[uint32]bool

	// stop is filled in by a hook that halted the run (a breaking watch), so the
	// reason survives the return from Run, which only reports "stop requested".
	stop debug.StopReason

	// The pad. held is which keys the browser currently has down; padQueue is the
	// pad states the machine has not sampled yet; padHold counts down the fields
	// the current state has left before the next queued one replaces it.
	held     map[string]bool
	padQueue []padState
	padHold  int
}

// padState is everything one Maple poll latches: buttons, both analog triggers and
// the stick, together — the game's VBlank ISR reads the whole condition in one
// response, so a state mixing two instants would be a pad that never existed.
// Buttons is a pressed mask (active-high); the wire's active-low convention is
// applied at the machine. Comparable by ==, which is what lets the queue drop a
// repeat.
type padState struct {
	buttons  uint16
	lt, rt   uint8
	joyX     uint8
	joyY     uint8
}

// The capabilities this target backs. Listed explicitly so that dropping one from
// the implementation breaks the build here rather than silently removing a panel.
var (
	_ debug.Target        = (*Adapter)(nil)
	_ debug.FrameStepper  = (*Adapter)(nil)
	_ debug.FastStepper   = (*Adapter)(nil)
	_ debug.FrameReplayer = (*Adapter)(nil)
	_ debug.CodeStepper   = (*Adapter)(nil)
	_ debug.Breakpointer  = (*Adapter)(nil)
	_ debug.Disassembler  = (*Adapter)(nil)
	_ debug.Watcher       = (*Adapter)(nil)
	_ debug.Surfacer      = (*Adapter)(nil)
	_ debug.FileLister    = (*Adapter)(nil)
	_ debug.StateFiler    = (*Adapter)(nil)
	_ debug.Resumer       = (*Adapter)(nil)
	_ debug.MemoryMapper  = (*Adapter)(nil)
	_ debug.Keyer         = (*Adapter)(nil)
	_ debug.KeyLegender   = (*Adapter)(nil)
	_ debug.Haltable      = (*Adapter)(nil)
)

// snap wraps a Dreamcast in-memory savestate as an opaque debug.Snapshot. The
// platform's SaveState/LoadState deep-copy every slice both ways (state_test.go's
// independence test), so restoring one repeatedly is sound — which is what replay
// is.
type snap struct{ ms dc.MachineState }

func (snap) Platform() string { return "dc" }

// New opens the disc at imagePath and boots it — the boot is the BIOS HLE's
// instant handover (vectors planted, binary at 8C010000), so unlike a console
// with an apploader this returns at the game's first instruction immediately.
// The first drawn frame is still ~900 fields away; load a savestate.
func New(imagePath string) (*Adapter, error) {
	disc, err := dc.OpenDisc(imagePath)
	if err != nil {
		return nil, err
	}
	m := dc.NewMachine(disc)
	if err := m.Boot(); err != nil {
		return nil, err
	}
	return newWithMachine(m, imagePath), nil
}

// newWithMachine wraps an already-built machine; the pad tests use it to drive a
// discless machine.
func newWithMachine(m *dc.Machine, imagePath string) *Adapter {
	a := &Adapter{imagePath: imagePath, live: m, held: map[string]bool{}, bps: map[uint32]bool{}}
	a.installPadPacing(m)
	return a
}

func (a *Adapter) Platform() string { return "dc" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the underlying live machine, for wiring the generic interface
// does not cover.
func (a *Adapter) Machine() *dc.Machine { return a.live }

// Close drops the machines. The dc.Machine owns no goroutines and the disc's file
// handle dies with the process's last reference, so this is only the reference drop.
func (a *Adapter) Close() error {
	a.live, a.scratch = nil, nil
	return nil
}

func (a *Adapter) LoadStateFile(path string) error { return a.live.LoadStateFile(path) }
func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ns, ok := s.(snap)
	if !ok {
		return fmt.Errorf("dcadapter: snapshot is from %q, not dc", platformOf(s))
	}
	return a.live.LoadState(ns.ms)
}

// Display is what the CRTC is scanning out: the buffer FB_R_SOF1 names, through
// the 32-bit-path bank interleave. Before the game configures video it errors —
// "no frame yet" is honest during early boot, and this machine's game turns the
// display on within its first seconds.
func (a *Adapter) Display() (*image.RGBA, error) { return a.live.RenderFB() }

// ReadMem reads main memory at a physical address (the 0x0C000000 mirror or a
// 0-based offset both land in RAM). It reads RAM only: the hardware registers are
// in the address space too, but reading one has side effects, and a debugger's
// memory pane must never be the thing that changes the machine.
func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		p := addr + uint32(i)
		if p>>26 == 3 { // 0x0C000000-0x0FFFFFFF: RAM and its mirrors
			out[i] = a.live.RAM[p&(dc.RAMSize-1)]
		} else if p < dc.RAMSize {
			out[i] = a.live.RAM[p]
		}
	}
	return out
}

// ---- frames ----

// StepFrame advances the live machine to the next completed render — the deferred
// STARTRENDER completion that rasterises the closed TA session — capturing the
// frame's parameter stream and per-pixel last-writer provenance. The picture is
// the write framebuffer, taken inside OnRender before the guest can flip.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SaveState()}}
	fc.CountWrites()

	var over map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[uint32][]debug.PixelWrite{}
	}
	cmd := -1 // index of the command currently drawing; -1 before the clear

	a.live.OnPVRClear = func(w, h int) {
		// The clear opens every render; its arrival also sizes the provenance
		// plane — the walk's pixels are in these framebuffer coordinates.
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index: cmd, Name: "CLEAR", Decoded: fmt.Sprintf("background clear %dx%d", w, h),
		})
		if fc.Prov == nil && w > 0 && h > 0 {
			fc.Width, fc.Height = w, h
			fc.Prov = make([]int32, w*h)
			for i := range fc.Prov {
				fc.Prov[i] = -1
			}
		}
	}
	a.live.OnPVRCmd = func(param []byte) {
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, decodeParam(cmd, param))
	}
	a.live.OnPVRPixel = func(x, y int, r, g, b, alpha uint8, drawn bool) {
		if cmd < 0 || x < 0 || y < 0 || x >= fc.Width || y >= fc.Height {
			return
		}
		if drawn {
			fc.MarkWrite(cmd)
			fc.Prov[y*fc.Width+x] = int32(cmd)
		}
		if over != nil {
			key := uint32(y)<<16 | uint32(x)&0xFFFF
			over[key] = append(over[key], debug.PixelWrite{
				CmdIndex: cmd, R: r, G: g, B: b, A: alpha, Rejected: !drawn,
			})
		}
	}

	var shot *image.RGBA
	a.live.OnRender = func() {
		if shot == nil {
			shot, _ = a.live.RenderDrawTarget()
			a.live.StopRequested = true
		}
	}

	a.runLive(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a
	// later run is not paying for a census nobody is reading.
	a.live.OnPVRClear, a.live.OnPVRCmd, a.live.OnPVRPixel, a.live.OnRender = nil, nil, nil, nil

	if shot != nil && fc.Width == 0 {
		fc.Width, fc.Height = shot.Bounds().Dx(), shot.Bounds().Dy()
	}
	if over != nil {
		fc.Overdraw = make(map[int][]debug.PixelWrite, len(over))
		for key, writes := range over {
			if x, y := int(key&0xFFFF), int(key>>16); x < fc.Width && y < fc.Height {
				fc.Overdraw[y*fc.Width+x] = writes
			}
		}
	}
	return fc, nil
}

// StepFast advances the live machine one completed render, capturing nothing: no
// snapshot (a ~26 MiB copy), no command stream, no census.
func (a *Adapter) StepFast() error {
	done := false
	a.live.OnRender = func() {
		if !done {
			done = true
			a.live.StopRequested = true
		}
	}
	a.runLive(runBudget)
	a.live.OnRender = nil
	return nil
}

// RenderAfter replays fc.Start in the scratch machine and returns the draw target
// exactly after command k executed. The render walk is a plain Go loop, so the
// halt is a counter (RenderStopAfter), not a sentinel panic; the partial frame is
// whatever the walk had plotted when it broke.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("dcadapter: capture holds a %q snapshot, not dc", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("dcadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc := a.scratchMachine()
	if err := sc.LoadState(s.ms); err != nil {
		return nil, err
	}
	sc.RenderStopAfter = k + 1 // 1-based count → stop after index k
	var shot *image.RGBA
	sc.OnRender = func() {
		if shot == nil {
			shot, _ = sc.RenderDrawTarget()
			sc.StopRequested = true
		}
	}
	sc.Run(runBudget, dc.RunConfig{NoSpin: true})
	sc.OnRender = nil
	sc.RenderStopAfter = 0
	if shot == nil {
		return nil, fmt.Errorf("dcadapter: the replay reached no render within the budget")
	}
	return shot, nil
}

// ---- CPU ----

var sh4RegNames = func() []string {
	n := make([]string, 16)
	for i := range n {
		n[i] = fmt.Sprintf("r%d", i)
	}
	n[15] = "r15/sp"
	return n
}()

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 16)
	for i := 0; i < 16; i++ {
		vals[i] = uint64(c.R[i])
	}
	return debug.CPUReg{
		PC:    uint64(c.PC),
		Names: sh4RegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"PR": uint64(c.PR), "SR": uint64(c.SR), "GBR": uint64(c.GBR),
			"VBR": uint64(c.VBR), "SSR": uint64(c.SSR), "SPC": uint64(c.SPC),
			"MACH": uint64(c.MACH), "MACL": uint64(c.MACL),
			"FPSCR": uint64(c.FPSCR), "FPUL": uint64(c.FPUL),
		},
	}
}

// Halted answers for the SH-4 core. The machine keeps its field clock ticking
// after a halt — Run returns immediately, but a played target still looks alive —
// so this is what separates a halted core from a busy one at the frame level.
func (a *Adapter) Halted() (bool, string) { return a.live.CPU.Halted, a.live.CPU.HaltReason }

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
	return a.runSlice(budget, "budget")
}

// runLive runs the live machine under the debugger's standing configuration:
// breakpoints armed, spin detection off (a parked machine is a normal state under
// a debugger, not a bug to report).
func (a *Adapter) runLive(budget uint64) dc.Result {
	return a.live.Run(budget, dc.RunConfig{Breakpoints: a.bps, NoSpin: true})
}

// runSlice runs the machine and works out why it stopped. Run's own Reason covers
// a breakpoint or a halt; a breaking watch reports itself through a.stop, because
// from Run's point of view a hook merely asked to stop.
func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	res := a.runLive(budget)
	sr := a.stop
	if sr.Kind == "" {
		switch {
		case res.Reason == "breakpoint":
			sr = debug.StopReason{Kind: "breakpoint"}
		case a.live.CPU.Halted:
			sr = debug.StopReason{Kind: "halted", Note: a.live.CPU.HaltReason}
		default:
			sr = debug.StopReason{Kind: exhausted, Note: res.Reason}
		}
	}
	sr.PC = uint64(a.live.CPU.PC)
	sr.Steps = res.Steps
	return sr, nil
}

func (a *Adapter) SetBreakpoint(pc uint64)   { a.bps[uint32(pc)] = true }
func (a *Adapter) ClearBreakpoint(pc uint64) { delete(a.bps, uint32(pc)) }

func (a *Adapter) Breakpoints() []uint64 {
	out := make([]uint64, 0, len(a.bps))
	for pc := range a.bps {
		out = append(out, uint64(pc))
	}
	return out
}

// Disasm decodes n instructions at a virtual address, reading the words from RAM
// directly — never through the bus, whose register reads have side effects.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*2
		var h uint16
		if p := va & 0x1FFFFFFF; p>>26 == 3 {
			off := p & (dc.RAMSize - 1)
			h = uint16(a.live.RAM[off]) | uint16(a.live.RAM[off+1])<<8
		}
		in := sh4.DecodeHalfword(h, va)
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: []byte{uint8(h), uint8(h >> 8)}, Text: in.Text})
	}
	return out, nil
}

// instrAt is the one-line disassembly at pc, used to say what wrote a watched
// address.
func (a *Adapter) instrAt(pc uint64) string {
	in, err := a.Disasm(pc, 1)
	if err != nil || len(in) == 0 {
		return ""
	}
	return in[0].Text
}

// ---- watches ----
//
// The machine's watch windows are slices of ranges over one callback, so any
// number of watches coexist; the adapter keeps the set and dispatches each hit to
// the watch whose range and kind it matches. Watches see CPU accesses only: a ch2
// DMA landing textures, the AICA's sample fetches, and the render walk itself all
// touch memory directly, so a watch that never fires is not proof nobody wrote
// the address.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("dcadapter: watch kind %q is not write or read", w.Kind)
	}
	if w.Hi <= w.Lo {
		return 0, fmt.Errorf("dcadapter: watch range [%08X,%08X) is empty", w.Lo, w.Hi)
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
	a.live.WatchW, a.live.WatchR, a.live.OnWatch = nil, nil, nil
	if len(a.watches) == 0 {
		return
	}
	for _, w := range a.watches {
		r := dc.WatchRange{Start: w.Lo, Len: w.Hi - w.Lo}
		if w.Kind == "write" {
			a.live.WatchW = append(a.live.WatchW, r)
		} else {
			a.live.WatchR = append(a.live.WatchR, r)
		}
	}
	a.live.OnWatch = func(write bool, addr, v uint32, size int, pc uint32) {
		kind := "read"
		if write {
			kind = "write"
		}
		for i := range a.watches {
			w := &a.watches[i]
			if w.Kind != kind || addr+uint32(size) <= w.Lo || addr >= w.Hi {
				continue
			}
			w.Hits++
			if a.watchSink != nil {
				a.watchSink(debug.WatchHit{
					ID: w.ID, Kind: w.Kind, Addr: addr, Val: v, PC: pc,
					Instr: a.instrAt(uint64(pc)),
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
	}
}

// ---- surfaces ----

// Surfaces are the pictures this machine can show: the pair the package comment is
// about — the buffer being scanned out and the buffer being drawn, one flip apart —
// and raw VRAM aimed by hand. VRAM view addresses are 64-bit-path offsets (the
// layout textures live in); the two framebuffer surfaces do their own 32-bit-path
// interleave, as the hardware does.
func (a *Adapter) Surfaces() []debug.Surface {
	sw, sh := 0, 0
	if img, err := a.live.RenderFB(); err == nil {
		sw, sh = img.Rect.Dx(), img.Rect.Dy()
	}
	dw, dh := 0, 0
	if img, err := a.live.RenderDrawTarget(); err == nil {
		dw, dh = img.Rect.Dx(), img.Rect.Dy()
	}
	return []debug.Surface{
		{ID: "drawtarget", Name: "Write framebuffer (PVR draw target)", W: dw, H: dh},
		{ID: "scanout", Name: "Read framebuffer (CRTC scanout)", W: sw, H: sh},
		{ID: "vram", Name: "VRAM as RGB565 (64-bit-path offsets)", Free: true, Formats: []string{"rgb565"}},
	}
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "scanout":
		return a.live.RenderFB()
	case "drawtarget":
		return a.live.RenderDrawTarget()
	case "vram":
		return a.live.RenderVRAM(v.Addr, v.W, v.H)
	}
	return nil, fmt.Errorf("dcadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// ---- the disc's filesystem ----

// ListDir lists one directory of the GD-ROM's ISO 9660 volume. Crazy Taxi streams
// its assets out of a handful of big containers, so the pane is mostly a map of
// what there is to extract.
func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	if a.live.Disc == nil || a.live.Disc.Vol == nil {
		return nil, debug.ErrUnsupported
	}
	entries, err := a.live.Disc.Vol.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]debug.FileEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, debug.FileEntry{
			Path: e.Path, Name: e.Name, Dir: e.IsDir, Size: int64(e.Size),
			Offset: int64(e.Block) * 2048,
		})
	}
	return out, nil
}

func (a *Adapter) ReadFile(path string) ([]byte, error) {
	if a.live.Disc == nil || a.live.Disc.Vol == nil {
		return nil, debug.ErrUnsupported
	}
	return a.live.Disc.Vol.ReadFile(path)
}

// ---- memory map ----

func (a *Adapter) Regions() []debug.Region {
	return []debug.Region{
		{Name: "BIOS work area (HLE-planted)", Lo: 0x0C00_0000, Hi: 0x0C01_0000},
		{Name: "Main RAM (1ST_READ.BIN at 0C010000)", Lo: 0x0C01_0000, Hi: 0x0D00_0000},
	}
}

// ResumeArgs is the command line that resumes this disc from a saved state, in the
// Crazy Taxi bootoracle's own flag vocabulary; the workspace runs it from the
// game's extract/.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{
		"go", "run", "./cmd/bootoracle",
		"-image", a.imagePath,
		"-loadstate", statePath,
		"-nospin",
	}
}

// ---- the pad ----
//
// The Dreamcast controller is a LEVEL the Maple bus samples once per field (the
// game's VBlank ISR kicks a Maple DMA and the driver latches the condition), and
// the game edge-detects presses across fields — the GameCube's shape, not the DOS
// keyboard's. So browser key events queue whole pad states and the field boundary
// releases them, each held for tapFields fields: long enough that the game's
// occasionally-overrunning mainline provably sees the edge (the oracle's -keys
// scripts learned the same tap length first).

func (a *Adapter) Key(k debug.Key) error {
	name := normalizeKey(k.Name)
	_, isButton := dc.PadButton(name)
	_, isStick := stickDirs[name]
	if !isButton && !isStick && name != "l" && name != "r" {
		return nil // an unmapped key is ignored rather than an error: the browser sends everything
	}
	if k.Down {
		a.held[name] = true
	} else {
		delete(a.held, name)
	}

	st := padState{joyX: 0x80, joyY: 0x80}
	for n := range a.held {
		if b, ok := dc.PadButton(n); ok {
			st.buttons |= b
		}
	}
	if a.held["l"] {
		st.lt = 255
	}
	if a.held["r"] {
		st.rt = 255
	}
	st.joyX, st.joyY = stickFrom(a.held)

	// Only a CHANGE needs fields of its own; a repeat (the browser's key-repeat)
	// is the same level and would just cost latency.
	if len(a.padQueue) > 0 && a.padQueue[len(a.padQueue)-1] == st {
		return nil
	}
	a.padQueue = append(a.padQueue, st)
	return nil
}

// KeyLegend says what the keys do; W-A-S-Z-style shapes need saying out loud.
func (a *Adapter) KeyLegend() string {
	return "arrows = analog stick · Enter = Start · A/B/X/Y = buttons · L/R = triggers · numpad 8/4/6/2 = d-pad"
}

// stickDirs are the four keys that push the analog stick. The offsets are in the
// pad's own wire convention, where 0x00 is full LEFT and full UP (the oracle's
// -keys jx/jy scripts document the same) — so up DECREASES joyY.
var stickDirs = map[string]struct{ dx, dy int }{
	"stickup":    {0, -1},
	"stickdown":  {0, +1},
	"stickleft":  {-1, 0},
	"stickright": {+1, 0},
}

// stickFrom resolves the held direction keys to the stick's two wire bytes. A
// corner splits the throw over both axes (127/√2 ≈ 90): the stick's gate will not
// let both axes read full at once, and a keyboard has no gate, so the gate is
// modelled here — the same reasoning as the GameCube's octagon.
func stickFrom(held map[string]bool) (x, y uint8) {
	var dx, dy int
	for n := range held {
		if d, ok := stickDirs[n]; ok {
			dx += d.dx
			dy += d.dy
		}
	}
	if dx == 0 && dy == 0 {
		return 0x80, 0x80
	}
	mag := 127
	if dx != 0 && dy != 0 {
		mag = 90
	}
	clamp := func(d int) uint8 {
		v := 0x80 + d*mag
		if v < 0 {
			v = 0
		}
		if v > 0xFF {
			v = 0xFF
		}
		return uint8(v)
	}
	return clamp(dx), clamp(dy)
}

// installPadPacing releases one queued pad state per tapFields fields. OnDisplay
// runs at the VBlank-in that begins the field the game's ISR polls in, so a state
// set here is what the next Maple DMA latches. An empty queue leaves the current
// state alone, which is what keeps a held key held.
//
// It chains any hook already installed rather than replacing it: OnDisplay is the
// machine's one field-boundary hook and the pad has no claim to own it.
func (a *Adapter) installPadPacing(m *dc.Machine) {
	prev := m.OnDisplay
	m.OnDisplay = func(field uint64) {
		if prev != nil {
			prev(field)
		}
		if a.padHold > 0 {
			a.padHold--
			return
		}
		if len(a.padQueue) == 0 {
			return
		}
		s := a.padQueue[0]
		a.padQueue = a.padQueue[1:]
		a.padHold = tapFields - 1
		m.Pad.Buttons = 0xFFFF &^ s.buttons // the wire is active-low
		m.Pad.LT, m.Pad.RT = s.lt, s.rt
		m.Pad.JoyX, m.Pad.JoyY = s.joyX, s.joyY
	}
}

// normalizeKey folds a browser KeyboardEvent.key value to the names the pad
// knows — dc.PadButton's names, the same ones the oracle's -keys scripts use,
// plus the stick directions and the l/r triggers. The letter keys map to the
// button of the same name; the arrows drive the STICK because that is what it
// takes to play the game (Crazy Taxi steers on jx), and the d-pad keeps the
// numpad's own arrows.
func normalizeKey(name string) string {
	switch name {
	case "ArrowUp", "arrowup":
		return "stickup"
	case "ArrowDown", "arrowdown":
		return "stickdown"
	case "ArrowLeft", "arrowleft":
		return "stickleft"
	case "ArrowRight", "arrowright":
		return "stickright"
	case "8":
		return "up"
	case "2":
		return "down"
	case "4":
		return "left"
	case "6":
		return "right"
	case "Enter", "enter", "Return", "return":
		return "start"
	}
	if len(name) == 1 {
		c := name[0]
		if c >= 'A' && c <= 'Z' {
			return string(c + 'a' - 'A')
		}
	}
	return name
}

// ---- helpers ----

// decodeParam names one TA parameter for the command list. The type field is the
// control word's top three bits; headers name their list, vertices their strip
// position.
func decodeParam(index int, param []byte) debug.GPUCommand {
	var pcw uint32
	if len(param) >= 4 {
		pcw = binary.LittleEndian.Uint32(param)
	}
	words := make([]uint64, 0, len(param)/4)
	for i := 0; i+4 <= len(param); i += 4 {
		words = append(words, uint64(binary.LittleEndian.Uint32(param[i:])))
	}
	name, decoded := paramName(pcw)
	return debug.GPUCommand{Index: index, Name: name, Op: pcw >> 29, Words: words, Decoded: decoded}
}

var taListNames = [8]string{"opaque", "opaque modifier", "translucent", "translucent modifier", "punch-through", "list 5", "list 6", "list 7"}

func paramName(pcw uint32) (name, decoded string) {
	list := taListNames[pcw>>24&7]
	switch pcw >> 29 {
	case 0:
		return "END_OF_LIST", ""
	case 1:
		return "USER_TILE_CLIP", ""
	case 2:
		return "OBJECT_LIST_SET", ""
	case 4:
		d := list
		if pcw&8 != 0 {
			d += ", textured"
		}
		if pcw&2 != 0 {
			d += ", gouraud"
		}
		switch pcw >> 4 & 3 {
		case 1:
			d += ", floating colour"
		case 2:
			d += ", intensity"
		case 3:
			d += ", intensity (reused)"
		}
		return "POLYGON", d
	case 5:
		return "SPRITE", list
	case 7:
		d := ""
		if pcw>>28&1 != 0 {
			d = "end of strip"
		}
		return "VERTEX", d
	}
	return fmt.Sprintf("PARAM_%d", pcw>>29), ""
}

func (a *Adapter) scratchMachine() *dc.Machine {
	if a.scratch == nil {
		// The scratch machine shares the live one's disc: the Disc is read-only
		// and the runner is single-threaded, so there is nothing to contend on,
		// and sharing keeps LoadState's identity check on the one cached MD5.
		a.scratch = dc.NewMachine(a.live.Disc)
	}
	return a.scratch
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
