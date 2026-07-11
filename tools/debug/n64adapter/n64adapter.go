// Package n64adapter implements debug.DebugTarget for the Nintendo 64 oracle.
//
// It owns a live n64.Machine and, for command replay, a scratch machine it keeps
// around and restores into. Almost everything the debugger needs is already a hook
// on the machine (OnRDPCmd, OnPixel, the watch windows, StopRequested) or an
// in-memory savestate; this package is the thin translation between those and the
// platform-agnostic debug surface.
package n64adapter

import (
	"fmt"
	"image"

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

// Adapter drives an N64 oracle as a debug.DebugTarget.
type Adapter struct {
	rom     *n64.ROM
	live    *n64.Machine
	scratch *n64.Machine // reused across RenderAfter; its state is always disposable
}

var _ debug.DebugTarget = (*Adapter)(nil)

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
	return &Adapter{rom: rom, live: live}, nil
}

// LoadStateFile restores a savestate file into the live machine, so the debugger
// can start at a prepared frame rather than boot from reset.
func (a *Adapter) LoadStateFile(path string) error { return a.live.LoadState(path) }

// Machine exposes the underlying live machine, for wiring the GUI needs that the
// generic interface does not cover (controller input, savestate files).
func (a *Adapter) Machine() *n64.Machine { return a.live }

func (a *Adapter) Name() string { return "Nintendo 64" }

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

func (a *Adapter) SetWriteWatch(lo, hi uint32, cb func(addr, val, pc uint32)) {
	a.live.WatchLo, a.live.WatchHi, a.live.OnWrite = lo, hi, cb
}

func (a *Adapter) SetReadWatch(lo, hi uint32, cb func(addr, val, pc uint32)) {
	a.live.RWatchLo, a.live.RWatchHi, a.live.OnRead = lo, hi, cb
}

func (a *Adapter) SetBreakpoint(vaddr uint64) { a.live.SetBreakpoint(uint32(vaddr)) }

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
