// Package n64adapter implements a debug.Target for the Nintendo 64 oracle.
//
// It owns a live n64.Machine and, for command replay, a scratch machine it keeps
// around and restores into. Almost everything the debugger needs is already a hook
// on the machine (OnRDPCmd, OnPixel, the watch windows, StopRequested) or an
// in-memory savestate; this package is the thin translation between those and the
// platform-agnostic debug surface.
//
// The N64 is the richest target in the repository: it backs every frame capability
// there is, because its rasteriser reports every pixel and its display processor can
// be halted mid-list. Other platforms will advertise less.
package n64adapter

import (
	"encoding/binary"
	"fmt"
	"image"
	"path/filepath"

	"retroreverse.com/tools/cpu/r4300"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/n64"
)

// runBudget bounds a single StepFrame/replay. A video field is ~750k instructions
// and a from-reset boot to the first frame is tens of millions, so this leaves
// generous headroom while still catching a machine that has wedged.
const runBudget = 400_000_000

// mipsRegNames are the o32 ABI names of the 32 general registers, in index order.
var mipsRegNames = []string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

// Adapter drives an N64 oracle as a debug.Target.
type Adapter struct {
	rom     *n64.ROM
	romPath string
	live    *n64.Machine
	scratch *n64.Machine // reused across RenderAfter; its state is always disposable

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64

	// stop is filled in by a hook that halted the run (a breaking watch), so the
	// reason survives the return from Run, which only reports "stop requested".
	stop debug.StopReason
}

// The capabilities this target backs. Listed explicitly so that dropping one from the
// implementation breaks the build here rather than silently removing a panel.
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

// snap wraps an N64 in-memory savestate as an opaque debug.Snapshot.
type snap struct{ ms *n64.MachineState }

func (snap) Platform() string { return "n64" }

// New boots an N64 oracle on the cartridge at romPath. The machine is left in the
// reset state; advance it with StepFrame, or jump ahead with LoadStateFile.
func New(romPath string) (*Adapter, error) {
	rom, err := n64.Load(romPath)
	if err != nil {
		return nil, err
	}
	live := n64.NewMachine(rom)
	if err := live.Boot(rom, n64.DefaultBoot()); err != nil {
		return nil, err
	}
	return &Adapter{rom: rom, romPath: romPath, live: live}, nil
}

// LoadStateFile restores a savestate file into the live machine, so the debugger can
// start at a prepared frame rather than boot from reset. SaveStateFile writes one in
// the same format the Pilotwings bootoracle's -loadstate reads.
func (a *Adapter) LoadStateFile(path string) error { return a.live.LoadState(path) }
func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveState(path) }

// Machine exposes the underlying live machine, for wiring the GUI needs that the
// generic interface does not cover (controller input, savestate files).
func (a *Adapter) Machine() *n64.Machine { return a.live }

func (a *Adapter) Platform() string { return "n64" }
func (a *Adapter) Title() string    { return filepath.Base(a.romPath) }

// Close drops the machines. The N64 machine holds no OS resources, so this only
// releases the memory for the garbage collector.
func (a *Adapter) Close() error {
	a.live, a.scratch = nil, nil
	return nil
}

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SnapshotState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ns, ok := s.(snap)
	if !ok {
		return fmt.Errorf("n64adapter: snapshot is from %q, not n64", platformOf(s))
	}
	return a.live.RestoreState(ns.ms)
}

// StepFrame advances the live machine to the next video field, capturing that
// field's RDP command stream and per-pixel last-writer provenance.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SnapshotState()}}

	// Provenance is gathered keyed by packed (y<<16|x) while the frame draws, then
	// laid out into a row-major slice once the draw target's size is known. Both x
	// and y fit in 16 bits for any N64 framebuffer.
	prov := map[uint32]int32{}
	var over map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[uint32][]debug.PixelWrite{}
	}
	cmd := -1 // index of the command currently drawing; -1 before the first

	a.live.OnRDPCmd = func(_ *n64.Machine, op uint32, words []uint64) {
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index: cmd,
			Name:  n64.RDPName(op),
			Op:    op,
			Words: append([]uint64(nil), words...),
		})
	}
	a.live.OnPixel = func(x, y uint32, ev n64.PixelEvent) {
		if cmd < 0 {
			return
		}
		key := y<<16 | (x & 0xFFFF)
		if ev.Drawn {
			prov[key] = int32(cmd)
		}
		if over != nil {
			over[key] = append(over[key], debug.PixelWrite{
				CmdIndex: cmd, R: u8(ev.R), G: u8(ev.G), B: u8(ev.B), A: u8(ev.A),
				Rejected: !ev.Drawn,
			})
		}
	}
	a.live.OnDisplay = func(m *n64.Machine) { m.StopRequested = true }

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later
	// run is not paying for a census nobody is reading.
	a.live.OnRDPCmd, a.live.OnPixel, a.live.OnDisplay = nil, nil, nil

	// Size provenance to the finished draw target.
	w, h := int(a.live.Width()), 0
	if img, err := a.live.RenderColorImage(); err == nil {
		w, h = img.Bounds().Dx(), img.Bounds().Dy()
	}
	fc.Width, fc.Height = w, h
	if w > 0 && h > 0 && len(prov) > 0 {
		fc.Prov = make([]int32, w*h)
		for i := range fc.Prov {
			fc.Prov[i] = -1
		}
		for key, c := range prov {
			if x, y := int(key&0xFFFF), int(key>>16); x < w && y < h {
				fc.Prov[y*w+x] = c
			}
		}
		if over != nil {
			fc.Overdraw = make(map[int][]debug.PixelWrite, len(over))
			for key, writes := range over {
				if x, y := int(key&0xFFFF), int(key>>16); x < w && y < h {
					fc.Overdraw[y*w+x] = writes
				}
			}
		}
	}
	return fc, nil
}

// StepFast advances the live machine one video field, capturing nothing: no
// frame-start snapshot, no command stream, no per-pixel census. It exists for
// fast-forwarding — running to the point in the game you actually want to look at —
// where a full StepFrame would pay for a 4 MiB gob snapshot and a census of every
// pixel that nobody is going to read.
//
// It is not part of debug.DebugTarget: a target that cannot offer it simply steps
// normally instead.
func (a *Adapter) StepFast() error {
	a.live.OnRDPCmd, a.live.OnPixel = nil, nil
	a.live.OnDisplay = func(m *n64.Machine) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnDisplay = nil
	return nil
}

// RenderAfter replays fc.Start in the scratch machine and returns the draw target
// exactly after command k executed.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("n64adapter: capture holds a %q snapshot, not n64", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("n64adapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc := a.scratchMachine()
	if err := sc.RestoreState(s.ms); err != nil {
		return nil, err
	}
	sc.OnRDPCmd, sc.OnPixel, sc.OnDisplay = nil, nil, nil
	sc.RunStopAfterRDPCommand(k+1, runBudget) // 1-based count → stop after index k
	return sc.RenderColorImage()
}

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 32)
	for i := 0; i < 32; i++ {
		vals[i] = c.R[i]
	}
	return debug.CPUReg{
		PC:    c.PC,
		Names: mipsRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"hi": c.HI, "lo": c.LO,
			"Status": c.COP0[12], "Cause": c.COP0[13], "EPC": c.COP0[14],
		},
	}
}

func (a *Adapter) Display() (*image.RGBA, error) { return a.live.Framebuffer() }

func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	ram := a.live.RDRAM
	for i := 0; i < n; i++ {
		if p := int(addr) + i; p >= 0 && p < len(ram) {
			out[i] = ram[p]
		}
	}
	return out
}

// Watches.
//
// The machine has exactly one write window and one read window, so a target can hold
// one watch of each kind at a time. Asking for a second of the same kind is an error
// rather than a silent replacement — quietly dropping a watch the user set is how a
// debugger lies to you.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("n64adapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("n64adapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

func (a *Adapter) Watches() []debug.Watch {
	return append([]debug.Watch(nil), a.watches...)
}

func (a *Adapter) OnWatchHit(sink func(debug.WatchHit)) {
	a.watchSink = sink
	a.applyWatches()
}

// applyWatches rewires the machine's windows to the current watch set. A watch marked
// Break also halts the run, which is what turns "report this write" into "stop when
// it happens".
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
		if w.Kind == "write" {
			a.live.WatchLo, a.live.WatchHi, a.live.OnWrite = w.Lo, w.Hi, cb
		} else {
			a.live.RWatchLo, a.live.RWatchHi, a.live.OnRead = w.Lo, w.Hi, cb
		}
	}
}

// Breakpoints. The machine can only clear all of them at once, so the adapter keeps
// the set and re-applies it.

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

// StepInstr runs n instructions and Continue runs until something stops it. Both are
// bounded: they hand control back so the caller can stay responsive and decide
// whether to run another slice.

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

// runSlice runs the machine and works out why it stopped. Run's own Reason covers a
// breakpoint or a halt; a breaking watch reports itself through a.stop, because from
// Run's point of view a hook merely asked to stop.
func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	res := a.live.Run(budget)
	sr := a.stop
	if sr.Kind == "" {
		sr = debug.StopReason{Kind: exhausted, Note: res.Reason}
		if a.live.CPU.Halted {
			sr.Kind, sr.Note = "halted", a.live.CPU.HaltReason
		} else {
			for _, b := range a.bps {
				if uint64(res.PC) == b {
					sr.Kind, sr.Note = "breakpoint", ""
					break
				}
			}
		}
	}
	sr.PC = a.live.CPU.PC
	sr.Steps = res.Steps
	return sr, nil
}

// Disasm decodes n instructions at a virtual address, for the disassembly pane.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := addr + uint64(i)*4
		w, ok := a.live.ReadVirt(va)
		if !ok {
			break
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], w)
		in := r4300.Decode(b[:], uint32(va))
		out = append(out, debug.Instr{Addr: va, Bytes: b[:], Text: in.Text})
	}
	return out, nil
}

// instrAt is the one-line disassembly at pc, used to say what wrote a watched
// address. An unreadable pc gives "", not an error: a watch report is worth having
// even when the PC is somewhere we cannot decode.
func (a *Adapter) instrAt(pc uint64) string {
	in, err := a.Disasm(pc, 1)
	if err != nil || len(in) == 0 {
		return ""
	}
	return in[0].Text
}

// Regions names the N64's address space for the memory pane.
func (a *Adapter) Regions() []debug.Region {
	return []debug.Region{
		{Name: "RDRAM", Lo: 0x0000_0000, Hi: 0x0080_0000},
		{Name: "RSP DMEM", Lo: 0x0400_0000, Hi: 0x0400_1000},
		{Name: "RSP IMEM", Lo: 0x0400_1000, Hi: 0x0400_2000},
		{Name: "Cartridge", Lo: 0x1000_0000, Hi: 0x1FC0_0000},
	}
}

// ResumeArgs is the command line that reopens this ROM at a saved state — the handoff
// to a shell or another session. Pilotwings' bootoracle spells the flag -loadstate.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{
		"go", "run", "./cmd/bootoracle",
		"-image", a.romPath,
		"-loadstate", statePath,
	}
}

func (a *Adapter) scratchMachine() *n64.Machine {
	if a.scratch == nil {
		a.scratch = n64.NewMachine(a.rom)
		_ = a.scratch.Boot(a.rom, n64.DefaultBoot())
	}
	return a.scratch
}

func u8(v uint32) uint8 {
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
