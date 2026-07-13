// Package psxadapter implements a debug.Target for the PlayStation oracle.
//
// It is the second adapter, and it exists to keep the capability design honest. The
// N64 could back every capability there is, which proves nothing; the PSX is a
// different machine and had to grow a frame hook, a pixel hook, an exported
// framebuffer and a stop-after-command sentinel before it could back the frame tools
// at all. What it still cannot do, it does not claim: there is no depth or alpha
// rejection in its rasteriser, so a pixel it reports is a pixel it drew, and the
// overdraw pane never shows a "rejected" write here.
//
// The machine is a BIOS-HLE one: there is no ROM, the boot executable is side-loaded
// from the disc, and a game whose interrupt dispatcher the retail BIOS would have
// installed needs that address supplied (Options.ISRHandler). That is the one piece
// of game-specific knowledge the PSX needs, and it is a caller's to provide — this
// package knows about the PlayStation, not about any game on it.
package psxadapter

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"retroreverse.com/tools/cpu/mips"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/psx"
)

// runBudget bounds a single StepFrame or replay. A field is ~250k instructions
// (stepsPerVBlank), and a boot to the first drawn frame is tens of millions, so this
// leaves headroom while still catching a machine that has wedged.
const runBudget = 400_000_000

// r3000RegNames are the o32 ABI names of the 32 general registers, in index order.
var r3000RegNames = []string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

// Options carries what the platform cannot work out for itself.
type Options struct {
	// ISRHandler is the game's own vectored-interrupt dispatcher. The retail BIOS
	// installs one via HookEntryInt into a slot the HLE BIOS leaves empty, so a game
	// that relies on it needs the traced address supplied or its interrupts never
	// vector. Zero means "use the registered chain handler".
	ISRHandler uint32
}

// Adapter drives a PSX oracle as a debug.Target.
type Adapter struct {
	imagePath string
	disc      *psx.Volume
	exe       *psx.EXE
	opts      Options

	live    *psx.Machine
	scratch *psx.Machine // reused across RenderAfter; its state is always disposable

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64
	stop      debug.StopReason
}

// The capabilities this target backs. The PSX has no filesystem panel gap — it has a
// disc — but it is not a Surfacer yet, and it has no FileLister yet either; those
// arrive with their milestones. What is absent here is absent from the page.
var (
	_ debug.Target        = (*Adapter)(nil)
	_ debug.FrameStepper  = (*Adapter)(nil)
	_ debug.FastStepper   = (*Adapter)(nil)
	_ debug.FrameReplayer = (*Adapter)(nil)
	_ debug.CodeStepper   = (*Adapter)(nil)
	_ debug.Breakpointer  = (*Adapter)(nil)
	_ debug.Disassembler  = (*Adapter)(nil)
	_ debug.Watcher       = (*Adapter)(nil)
	_ debug.StateFiler    = (*Adapter)(nil)
	_ debug.Resumer       = (*Adapter)(nil)
	_ debug.MemoryMapper  = (*Adapter)(nil)
)

// snap wraps a PSX in-memory savestate as an opaque debug.Snapshot. psx.SaveState
// deep-copies RAM, scratchpad, VRAM and the CD buffers, so a snapshot outlives the
// machine that made it and can be restored repeatedly — which is what replay needs.
type snap struct{ ms *psx.MachineState }

func (snap) Platform() string { return "psx" }

// New opens a disc image and side-loads its boot executable, the way the PSX
// bootoracle does.
func New(imagePath string, opts Options) (*Adapter, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	disc, err := psx.Open(data)
	if err != nil {
		return nil, fmt.Errorf("opening the disc: %w", err)
	}
	_, exe, err := disc.BootEXE()
	if err != nil {
		return nil, fmt.Errorf("finding the boot executable (SYSTEM.CNF): %w", err)
	}
	a := &Adapter{imagePath: imagePath, disc: disc, exe: exe, opts: opts}
	a.live = a.build()
	return a, nil
}

// build makes a machine from the disc, exactly as the oracle does. Both the live
// machine and the replay scratch come from here, so a replay starts from the same
// machine the capture was taken on.
func (a *Adapter) build() *psx.Machine {
	m := psx.NewMachine()
	m.SetDisc(a.disc)
	m.ISRHandler = a.opts.ISRHandler
	m.LoadEXE(a.exe)
	return m
}

func (a *Adapter) Platform() string { return "psx" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the live machine, for the wiring the generic interface does not
// cover (scripted pad input).
func (a *Adapter) Machine() *psx.Machine { return a.live }

func (a *Adapter) Close() error {
	a.live, a.scratch = nil, nil
	return nil
}

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ps, ok := s.(snap)
	if !ok {
		return fmt.Errorf("psxadapter: snapshot is from %q, not psx", platformOf(s))
	}
	a.live.LoadState(ps.ms)
	return nil
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }
func (a *Adapter) LoadStateFile(path string) error { return a.live.LoadStateFile(path) }

// StepFrame advances to the next vertical blank — the machine's only frame boundary,
// since the PSX display is a window onto VRAM and nothing "presents" — capturing the
// GP0 command stream and per-pixel provenance along the way.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SaveState()}}

	// Provenance is gathered keyed by packed (y<<16|x) while the frame draws, then
	// laid out row-major once the display size is known.
	prov := map[uint32]int32{}
	var over map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[uint32][]debug.PixelWrite{}
	}
	cmd := -1 // the command currently drawing; -1 before the first

	a.live.OnGP0(func(words []uint32) {
		cmd = len(fc.Commands)
		w64 := make([]uint64, len(words))
		for i, w := range words {
			w64[i] = uint64(w)
		}
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index: cmd,
			Name:  psx.GP0Name(words[0] >> 24),
			Op:    words[0] >> 24,
			Words: w64,
		})
	})
	a.live.OnPixel(func(x, y uint32, ev psx.PixelEvent) {
		if cmd < 0 {
			return
		}
		key := y<<16 | (x & 0xFFFF)
		prov[key] = int32(cmd)
		if over != nil {
			over[key] = append(over[key], debug.PixelWrite{
				CmdIndex: cmd, R: ev.R, G: ev.G, B: ev.B, A: ev.A,
				Rejected: !ev.Drawn,
			})
		}
	})
	a.live.OnDisplay = func(m *psx.Machine) { m.StopRequested = true }

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later
	// run is not paying for a census nobody is reading.
	a.live.OnGP0(nil)
	a.live.OnPixel(nil)
	a.live.OnDisplay = nil

	// Provenance is laid out over the *draw target*, not the scanout. A
	// double-buffering game draws into a back buffer elsewhere in VRAM while the
	// display window still shows the previous frame, so mapping onto the display
	// would throw away every pixel of the frame we just captured — and the scrubber
	// would show a picture that never changes. The drawing offset is the back
	// buffer's origin (see psx.DrawOrigin).
	//
	// A pixel drawn outside that buffer — a texture upload into another part of VRAM
	// — genuinely has nowhere to land on screen, and is dropped here.
	img := a.live.RenderDrawTarget()
	fc.Width, fc.Height = img.Rect.Dx(), img.Rect.Dy()
	dx, dy := a.live.DrawOrigin()
	if fc.Width > 0 && fc.Height > 0 && len(prov) > 0 {
		fc.Prov = make([]int32, fc.Width*fc.Height)
		for i := range fc.Prov {
			fc.Prov[i] = -1
		}
		put := func(key uint32, set func(i int)) {
			x, y := int(key&0xFFFF)-dx, int(key>>16)-dy
			if x >= 0 && y >= 0 && x < fc.Width && y < fc.Height {
				set(y*fc.Width + x)
			}
		}
		for key, c := range prov {
			c := c
			put(key, func(i int) { fc.Prov[i] = c })
		}
		if over != nil {
			fc.Overdraw = make(map[int][]debug.PixelWrite, len(over))
			for key, writes := range over {
				writes := writes
				put(key, func(i int) { fc.Overdraw[i] = writes })
			}
		}
	}
	return fc, nil
}

// StepFast advances one field capturing nothing — how the debugger fast-forwards.
func (a *Adapter) StepFast() error {
	a.live.OnGP0(nil)
	a.live.OnPixel(nil)
	a.live.OnDisplay = func(m *psx.Machine) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnDisplay = nil
	return nil
}

// RenderAfter replays the frame in a scratch machine and returns the display exactly
// after GP0 command k executed.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("psxadapter: capture holds a %q snapshot, not psx", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("psxadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc := a.scratchMachine()
	sc.LoadState(s.ms)
	sc.OnGP0(nil)
	sc.OnPixel(nil)
	sc.OnDisplay = nil
	sc.RunStopAfterGP0Command(k+1, runBudget) // 1-based count → stop after 0-based index k
	// The draw target, not the scanout: mid-frame the game is drawing into the back
	// buffer, and the displayed window is still showing the previous frame.
	return sc.RenderDrawTarget(), nil
}

func (a *Adapter) scratchMachine() *psx.Machine {
	if a.scratch == nil {
		a.scratch = a.build()
	}
	return a.scratch
}

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 32)
	for i := 0; i < 32; i++ {
		vals[i] = uint64(c.R[i])
	}
	return debug.CPUReg{
		PC:    uint64(c.PC),
		Names: r3000RegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"hi": uint64(c.HI), "lo": uint64(c.LO),
			"Status": uint64(c.COP0[12]), "Cause": uint64(c.COP0[13]), "EPC": uint64(c.COP0[14]),
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
	return []debug.Region{
		{Name: "RAM", Lo: 0x8000_0000, Hi: 0x8020_0000},
		{Name: "Scratchpad", Lo: 0x1F80_0000, Hi: 0x1F80_0400},
		{Name: "I/O", Lo: 0x1F80_1000, Hi: 0x1F80_3000},
	}
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

// runSlice runs the machine and works out why it stopped. Run's own reason covers a
// breakpoint or a halt; a breaking watch reports itself through a.stop, because from
// Run's point of view a hook merely asked to stop.
func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	res := a.live.Run(budget)
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

// Disasm decodes n instructions at a virtual address.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*4
		var b [4]byte
		for j := 0; j < 4; j++ {
			b[j] = a.live.Read(va + uint32(j))
		}
		in := mips.Decode(b[:], va)
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

func (a *Adapter) SetBreakpoint(pc uint64) {
	for _, b := range a.bps {
		if b == pc {
			return
		}
	}
	a.bps = append(a.bps, pc)
	a.live.SetBreakpoint(uint32(pc))
}

func (a *Adapter) ClearBreakpoint(pc uint64) {
	kept := a.bps[:0]
	for _, b := range a.bps {
		if b != pc {
			kept = append(kept, b)
		}
	}
	a.bps = kept
	a.live.ClearBreakpoints()
	for _, b := range a.bps {
		a.live.SetBreakpoint(uint32(b))
	}
}

func (a *Adapter) Breakpoints() []uint64 { return append([]uint64(nil), a.bps...) }

// ---- watches ----
//
// Like the N64, the machine has one write window and one read window, so a target
// holds one watch of each kind. A second of the same kind is an error rather than a
// silent replacement.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("psxadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("psxadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

// ResumeArgs is the command line that reopens this disc at a saved state. The PSX
// oracle spells the flags -load and -isr, not -loadstate — the vocabularies never
// agreed, and pretending otherwise would hand over a line that does not run.
func (a *Adapter) ResumeArgs(statePath string) []string {
	args := []string{"go", "run", "./cmd/bootoracle", "-image", a.imagePath, "-load", statePath}
	if a.opts.ISRHandler != 0 {
		args = append(args, "-isr", fmt.Sprintf("%08X", a.opts.ISRHandler))
	}
	return args
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
