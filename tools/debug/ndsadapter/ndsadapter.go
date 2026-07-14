// Package ndsadapter drives the Nintendo DS oracle (tools/platform/nds/dsmachine) as
// a debug.Target, so the frame debugger can step, scrub and inspect a DS game.
//
// Two things about the DS make its adapter different from every other one here, and
// both are properties of the machine rather than choices.
//
// THE FRAME IS TWO SCREENS. A DS has two panels and they are both the frame — a game
// puts its world on one and its map, its menu, its stylus target on the other, and a
// debugger that showed one of them would be showing half the picture. So the frame is
// the console's own composition: the top panel above the bottom one, with the bezel's
// gap between them, 256 x 392. Which engine drives which panel is not a constant —
// POWCNT1 bit 15 says, and SM64DS flips it — so the composition asks the machine
// rather than assuming.
//
// THE RASTERISER RUNS ONCE, AT THE END. On the N64 and the PSX the display processor
// draws as the commands arrive, so "the framebuffer after command k" is just the
// framebuffer, halted. The DS's geometry engine ACCUMULATES polygons all frame and the
// rasteriser does not run until the buffer swap: there is no partial framebuffer to
// stop at, because there is no framebuffer yet. So the scrubber replays to command k
// and then rasterises the polygon list AS IT STANDS. Dragging it is watching a frame's
// geometry accumulate — which is the right question to ask of a DS frame, because on
// this machine the geometry IS the frame and the rasteriser is a formality.
//
// Provenance follows from that. The rasteriser reports the polygon that produced each
// fragment, and every polygon remembers the geometry command that closed it — so a
// pixel names a command. But it names it only where the 3D layer actually WON the 2D
// compositor's priority fight (dsmachine.ThreeDVisible): a 3D pixel with a background
// on top of it is a pixel the player cannot see, and answering "this draw made that"
// for it would be a confident wrong answer.
package ndsadapter

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/dsmachine"
)

// The debugger opens a DS game through the registry, so nothing has to import this
// package by name — importing it for its side effect is enough.
func init() {
	debug.Register("ds", func(s debug.OpenSpec) (debug.Target, error) {
		return New(s.Image, Options{DTCM: s.Get("dtcm")})
	})
}

// runBudget bounds a single StepFrame or replay. A DS frame is ~1.1M ARM9 instructions
// of scanline pacing; this leaves generous headroom while still catching a machine that
// has wedged.
const runBudget = 40_000_000

// quantum is how many ARM9 instructions run before the ARM7 gets a turn. It matches the
// oracle's default, and it matters that it does: the two cores' interleaving is part of
// the machine's behaviour, so a debugger that scheduled them differently from the oracle
// would be replaying a subtly different machine and could not be compared with it.
const quantum = 64

// screenGap is the dead space between the two panels, standing in for the console's
// hinge. It belongs to no engine and stays black.
const screenGap = 8

// Screen dimensions, and the composed frame they make.
const (
	panelW = dsmachine.ScreenW
	panelH = dsmachine.ScreenH
	frameW = panelW
	frameH = panelH*2 + screenGap
)

var armRegNames = []string{
	"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7",
	"r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc",
}

// Options carry what the platform cannot work out for itself and that is a property of
// the GAME rather than the machine.
type Options struct {
	// DTCM is the ARM9 data-TCM base the game programs through CP15, as a hex string.
	// The machine needs it up front because the BIOS interrupt vectors live at the top
	// of DTCM, and an interrupt delivered to the wrong address is one the game never
	// takes. It comes from the game's debug.json, so no game knowledge lives here.
	DTCM string
}

// Adapter drives a DS oracle as a debug.Target.
type Adapter struct {
	imagePath string
	rom       *nds.ROM
	dtcm      uint32

	live      *dsmachine.Machine
	scratch   *dsmachine.Machine   // reused by RenderAfter; its state is always disposable
	replayers []*dsmachine.Machine // the machines RenderAfterBatch replays on, in parallel

	// aTop is which panel engine A drove when the last capture was composed. The
	// scrubber must lay its picture out the way the CAPTURE was laid out, not the way
	// the replayed machine happens to have its POWCNT at command k — a frame's geometry
	// is the frame's, fixed when it was taken.
	aTop bool

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64

	stop debug.StopReason
}

// The capabilities this target backs. Listed explicitly so that dropping one from the
// implementation breaks the build here rather than silently removing a panel.
var (
	_ debug.Target         = (*Adapter)(nil)
	_ debug.FrameStepper   = (*Adapter)(nil)
	_ debug.FastStepper    = (*Adapter)(nil)
	_ debug.FrameReplayer  = (*Adapter)(nil)
	_ debug.BatchReplayer  = (*Adapter)(nil)
	_ debug.CodeStepper    = (*Adapter)(nil)
	_ debug.Breakpointer   = (*Adapter)(nil)
	_ debug.Disassembler   = (*Adapter)(nil)
	_ debug.Watcher        = (*Adapter)(nil)
	_ debug.Surfacer       = (*Adapter)(nil)
	_ debug.FileLister     = (*Adapter)(nil)
	_ debug.FileAttributer = (*Adapter)(nil)
	_ debug.StateFiler     = (*Adapter)(nil)
	_ debug.Resumer        = (*Adapter)(nil)
	_ debug.ResumeDirer    = (*Adapter)(nil)
	_ debug.MemoryMapper   = (*Adapter)(nil)
	_ debug.Profiler       = (*Adapter)(nil)
)

// snap wraps a DS in-memory savestate as an opaque debug.Snapshot.
type snap struct{ ms *dsmachine.MachineState }

func (snap) Platform() string { return "ds" }

// New boots a DS oracle on the cartridge image at path. The machine is left at its
// entry point; advance it with StepFrame, or — since reaching anything worth looking at
// costs a billion-step boot — jump straight to a prepared frame with LoadStateFile.
func New(path string, opts Options) (*Adapter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rom, err := nds.Open(data)
	if err != nil {
		return nil, err
	}
	dtcm := uint32(0)
	if opts.DTCM != "" {
		if _, err := fmt.Sscanf(strings.TrimPrefix(opts.DTCM, "0x"), "%x", &dtcm); err != nil {
			return nil, fmt.Errorf("ndsadapter: bad dtcm %q: %w", opts.DTCM, err)
		}
	}
	a := &Adapter{imagePath: path, rom: rom, dtcm: dtcm}
	a.live = dsmachine.New(rom, dtcm)
	// Per-subsystem timing, on for the debugger's whole life. It costs a few hundred
	// clock reads a frame — nothing beside the per-pixel hook the debugger already
	// installs — and it is what the profile panel reads.
	a.live.SetProfile(true)
	return a, nil
}

func (a *Adapter) Platform() string { return "ds" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }
func (a *Adapter) Close() error     { return nil }

// Machine exposes the live oracle, for tools built on the adapter.
func (a *Adapter) Machine() *dsmachine.Machine { return a.live }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SnapshotState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	sn, ok := s.(snap)
	if !ok {
		return fmt.Errorf("ndsadapter: cannot restore a %q snapshot into a DS", platformOf(s))
	}
	return a.live.RestoreState(sn.ms)
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveState(path) }

func (a *Adapter) LoadStateFile(path string) error {
	if err := a.live.LoadState(path); err != nil {
		return err
	}
	a.aTop = a.live.EngineAOnTop()
	return nil
}

// --- frames -------------------------------------------------------------------

// StepFrame advances the live machine to the next VBlank, capturing the geometry
// commands that executed on the way and per-pixel provenance.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{
		Start:  snap{ms: a.live.SnapshotState()},
		Width:  frameW,
		Height: frameH,
	}

	// Provenance is gathered in the 3D plane's own coordinates (256x192) while the frame
	// draws, and moved onto the composed frame afterwards — once we know which panel
	// engine A ended up driving, and which of its pixels the 2D compositor let through.
	prov := make([]int32, panelW*panelH)
	for i := range prov {
		prov[i] = -1
	}
	var over map[int][]debug.PixelWrite
	if withOverdraw {
		over = map[int][]debug.PixelWrite{}
	}

	cur := -1 // the command that produced the polygon currently rasterising
	a.live.OnGXCmd(func(cmd uint8, p []uint32) {
		fc.Commands = append(fc.Commands, gxCommand(len(fc.Commands), cmd, p))
	})
	a.live.OnPoly = func(cmdIdx int) { cur = cmdIdx }
	a.live.OnPixel = func(x, y int, ev dsmachine.PixelEvent) {
		if cur < 0 || x < 0 || y < 0 || x >= panelW || y >= panelH {
			return
		}
		i := y*panelW + x
		if ev.Drawn {
			prov[i] = int32(cur)
		}
		if withOverdraw {
			over[i] = append(over[i], debug.PixelWrite{
				CmdIndex: cur,
				R:        expand6(ev.R), G: expand6(ev.G), B: expand6(ev.B), A: expand5(ev.A),
				Rejected: !ev.Drawn,
			})
		}
	}
	// Run to the VBLANK, not to the frame counter's tick.
	//
	// They are not the same instant, and the difference is the whole capture. The frame
	// counter increments at line 0; the 3D rasteriser runs at the vertical blank, line
	// 192, when the buffer swap is consumed. Stop at line 0 and the window contains a
	// frame's geometry commands and none of its pixels — the NEXT step gets the pixels,
	// with no commands to attribute them to. Both halves look plausible on their own,
	// and neither is a frame.
	a.live.OnFrame = func() { a.live.StopRequested() }
	a.live.Run(runBudget, quantum, nil)

	// Drop the capture hooks; leave any user watches wired. A later run must not pay for
	// a census nobody is reading.
	a.live.OnGXCmd(nil)
	a.live.OnPoly, a.live.OnPixel, a.live.OnFrame = nil, nil, nil

	a.aTop = a.live.EngineAOnTop()

	// Move provenance onto the composed frame, but only where the player can actually
	// see the 3D layer (see the package comment). Everything else stays -1: untouched,
	// which is the truth.
	fc.Prov = make([]int32, frameW*frameH)
	for i := range fc.Prov {
		fc.Prov[i] = -1
	}
	if withOverdraw {
		fc.Overdraw = map[int][]debug.PixelWrite{}
	}
	vis := a.live.ThreeDVisible()
	y0 := panelOriginY(a.aTop)
	for y := 0; y < panelH; y++ {
		for x := 0; x < panelW; x++ {
			i := y*panelW + x
			if i >= len(vis) || !vis[i] {
				continue
			}
			di := (y0+y)*frameW + x
			fc.Prov[di] = prov[i]
			if withOverdraw {
				if ws := over[i]; ws != nil {
					fc.Overdraw[di] = ws
				}
			}
		}
	}
	return fc, nil
}

// StepFast advances one frame capturing nothing — no snapshot, no command stream, no
// pixel census. It is how the debugger fast-forwards to the frame worth looking at.
func (a *Adapter) StepFast() error {
	a.live.OnGXCmd(nil)
	a.live.OnPoly, a.live.OnPixel = nil, nil
	a.live.OnFrame = func() { a.live.StopRequested() }
	a.live.Run(runBudget, quantum, nil)
	a.live.OnFrame = nil
	a.aTop = a.live.EngineAOnTop()
	return nil
}

// RenderAfter replays the frame's start snapshot in the scratch machine, stops the
// geometry engine after command k, rasterises the polygons submitted so far, and returns
// the composed screens.
//
// That the rasteriser has to be RUN, rather than merely halted, is the DS's peculiarity
// and is explained in the package comment. The consequence for the reader of the picture
// is worth stating plainly: what the scrubber shows at command k is not "the frame,
// half-drawn" — it is "the frame, if the game stopped submitting geometry here".
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("ndsadapter: capture holds a %q snapshot, not ds", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("ndsadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc, err := a.scratchMachine()
	if err != nil {
		return nil, err
	}
	return a.replayOn(sc, s.ms, k)
}

// replayOn is RenderAfter's body against a caller-supplied machine, so the serial path
// and the parallel one cannot drift apart.
func (a *Adapter) replayOn(sc *dsmachine.Machine, ms *dsmachine.MachineState, k int) (*image.RGBA, error) {
	if err := sc.RestoreState(ms); err != nil {
		return nil, err
	}
	sc.OnGXCmd(nil)
	sc.OnPoly, sc.OnPixel, sc.OnFrame = nil, nil, nil
	sc.RunStopAfterGXCommand(k+1, runBudget, quantum) // 1-based count → stop after index k
	sc.ForceRender()
	// Compose in the layout the CAPTURE was taken in, not the one the replayed machine
	// happens to hold: the game may flip POWCNT mid-frame, and a scrubber whose two
	// panels swapped places halfway through a drag would be unusable.
	return a.compose(sc, a.aTop), nil
}

const maxReplayers = 4

// RenderAfterBatch renders several points of the command stream at once, on independent
// machines.
//
// This is free of the determinism question, and it is worth saying why: every replay
// restores the SAME snapshot into a machine of its OWN and throws it away. The machines
// cannot see each other, there is no shared buffer to race over, and each one's answer
// is exactly the answer it would have given alone.
func (a *Adapter) RenderAfterBatch(fc *debug.FrameCapture, ks []int) ([]*image.RGBA, error) {
	if len(ks) == 0 {
		return nil, nil
	}
	if len(ks) == 1 {
		img, err := a.RenderAfter(fc, ks[0])
		if err != nil {
			return nil, err
		}
		return []*image.RGBA{img}, nil
	}
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("ndsadapter: capture holds a %q snapshot, not ds", platformOf(fc.Start))
	}
	for _, k := range ks {
		if k < 0 || k >= len(fc.Commands) {
			return nil, fmt.Errorf("ndsadapter: command %d out of range [0,%d)", k, len(fc.Commands))
		}
	}

	n := len(ks)
	if n > maxReplayers {
		n = maxReplayers
	}
	for len(a.replayers) < n {
		a.replayers = append(a.replayers, dsmachine.New(a.rom, a.dtcm))
	}

	out := make([]*image.RGBA, len(ks))
	errs := make([]error, n)
	var wg sync.WaitGroup
	for w := 0; w < n; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			sc := a.replayers[w]
			for i := w; i < len(ks); i += n { // each machine takes every n-th key
				img, err := a.replayOn(sc, s.ms, ks[i])
				if err != nil {
					errs[w] = err
					return
				}
				out[i] = img
			}
		}(w)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (a *Adapter) scratchMachine() (*dsmachine.Machine, error) {
	if a.scratch == nil {
		a.scratch = dsmachine.New(a.rom, a.dtcm)
	}
	return a.scratch, nil
}

// --- the console's own screen layout -------------------------------------------

// panelOriginY is the row a panel starts at in the composed frame. Engine A's panel is
// the one with the 3D layer, so this is what moves provenance onto the right screen.
func panelOriginY(aTop bool) int {
	if aTop {
		return 0
	}
	return panelH + screenGap
}

// compose stacks the two panels the way the console does: top above bottom, with the
// bezel's gap between them. Which engine is which is POWCNT1's business, and the machine
// has already applied it — Screens() hands back top and bottom, not A and B.
func (a *Adapter) compose(m *dsmachine.Machine, aTop bool) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, frameW, frameH))
	top, bottom := m.Screens()
	blit := func(src *image.RGBA, y0 int) {
		for y := 0; y < panelH; y++ {
			si := src.PixOffset(0, y)
			di := out.PixOffset(0, y0+y)
			copy(out.Pix[di:di+panelW*4], src.Pix[si:si+panelW*4])
		}
	}
	blit(top, 0)
	blit(bottom, panelH+screenGap)
	// The gap is left as it was allocated: transparent black. It belongs to no engine,
	// and filling it with a colour would invite the reader to think a screen drew it.
	return out
}

func (a *Adapter) Display() (*image.RGBA, error) { return a.compose(a.live, a.aTop), nil }

// --- the stylus ------------------------------------------------------------------
//
// The touchscreen is bonded to the console's LOWER panel, and it stays there whatever
// POWCNT1 says: flipping the engines swaps which picture is drawn on the lower LCD, not
// which LCD the digitiser is glued to. So the touch panel sits at a fixed place in the
// composed frame — unlike provenance, which has to follow engine A around (panelOriginY).

func (a *Adapter) TouchPanels() []debug.TouchPanel {
	return []debug.TouchPanel{{
		ID: "bottom", Name: "Touch screen",
		X: 0, Y: panelH + screenGap, W: panelW, H: panelH,
	}}
}

// Touch puts the stylus down on the lower panel, or lifts it. The panel's pixels are the
// touchscreen's own, and dsmachine.SetTouch runs them back through the firmware
// calibration the machine synthesised — so the game's own calibration arithmetic lands on
// the pixel that was clicked.
func (a *Adapter) Touch(t debug.Touch) error {
	if t.Panel != "bottom" {
		return fmt.Errorf("%w: no touch panel %q", debug.ErrUnsupported, t.Panel)
	}
	a.live.SetTouch(t.X, t.Y, t.Down)
	return nil
}

// --- commands -------------------------------------------------------------------

// gxNames labels the geometry engine's commands, so a command list reads as the
// geometry the game submitted rather than as a column of opcodes.
var gxNames = map[uint8]string{
	0x10: "MTX_MODE", 0x11: "MTX_PUSH", 0x12: "MTX_POP", 0x13: "MTX_STORE",
	0x14: "MTX_RESTORE", 0x15: "MTX_IDENTITY", 0x16: "MTX_LOAD_4x4", 0x17: "MTX_LOAD_4x3",
	0x18: "MTX_MULT_4x4", 0x19: "MTX_MULT_4x3", 0x1A: "MTX_MULT_3x3", 0x1B: "MTX_SCALE",
	0x1C: "MTX_TRANS", 0x20: "COLOR", 0x21: "NORMAL", 0x22: "TEXCOORD",
	0x23: "VTX_16", 0x24: "VTX_10", 0x25: "VTX_XY", 0x26: "VTX_XZ", 0x27: "VTX_YZ",
	0x28: "VTX_DIFF", 0x29: "POLYGON_ATTR", 0x2A: "TEXIMAGE_PARAM", 0x2B: "PLTT_BASE",
	0x30: "DIF_AMB", 0x31: "SPE_EMI", 0x32: "LIGHT_VECTOR", 0x33: "LIGHT_COLOR",
	0x34: "SHININESS", 0x40: "BEGIN_VTXS", 0x41: "END_VTXS", 0x50: "SWAP_BUFFERS",
	0x60: "VIEWPORT", 0x70: "BOX_TEST", 0x71: "POS_TEST", 0x72: "VEC_TEST",
}

func gxCommand(idx int, cmd uint8, p []uint32) debug.GPUCommand {
	name := gxNames[cmd]
	if name == "" {
		name = fmt.Sprintf("GX_%02X", cmd)
	}
	words := make([]uint64, len(p))
	parts := make([]string, len(p))
	for i, v := range p {
		words[i] = uint64(v)
		parts[i] = fmt.Sprintf("%08X", v)
	}
	return debug.GPUCommand{
		Index: idx, Name: name, Op: uint32(cmd), Words: words,
		Decoded: strings.Join(parts, " "),
	}
}

// --- CPU ------------------------------------------------------------------------

// CPU reports the ARM9's register file. The ARM7 is a real second processor and is not
// hidden — but a debugger pane shows one register file, and the ARM9 is the core a DS
// game's logic runs on, so it is the one on show. The ARM7's state is reachable through
// the Extra map rather than pretended away.
func (a *Adapter) CPU() debug.CPUReg {
	r := a.live.Regs(true)
	vals := make([]uint64, 16)
	for i, v := range r {
		vals[i] = uint64(v)
	}
	ie9, if9, ime9 := a.live.IRQState(true)
	ie7, if7, _ := a.live.IRQState(false)
	imeV := uint64(0)
	if ime9 {
		imeV = 1
	}
	return debug.CPUReg{
		PC:    uint64(r[15]),
		Names: armRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"arm7_pc": uint64(a.live.ARM7PC()),
			"IE":      uint64(ie9),
			"IF":      uint64(if9),
			"IME":     imeV,
			"arm7_IE": uint64(ie7),
			"arm7_IF": uint64(if7),
			"frame":   a.live.Frame(),
			"line":    uint64(a.live.Line()),
		},
	}
}

func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	if n <= 0 {
		return nil
	}
	return a.live.Snapshot(true, addr, uint32(n))
}

func (a *Adapter) StepInstr(n int) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	ran := a.live.StepInstructions(n)
	if hit, pc := a.live.Stopped(); hit {
		return debug.StopReason{Kind: "breakpoint", PC: uint64(pc), Steps: uint64(ran)}, nil
	}
	if a.stop.Kind != "" {
		a.stop.Steps = uint64(ran)
		return a.stop, nil
	}
	return debug.StopReason{Kind: "steps", PC: uint64(a.live.ARM9PC()), Steps: uint64(ran)}, nil
}

func (a *Adapter) Continue(budget uint64) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	if budget == 0 {
		budget = runBudget
	}
	before := a.live.Steps
	a.live.Run(budget, quantum, nil)
	steps := a.live.Steps - before
	if hit, pc := a.live.Stopped(); hit {
		return debug.StopReason{Kind: "breakpoint", PC: uint64(pc), Steps: steps}, nil
	}
	if a.stop.Kind != "" {
		a.stop.Steps = steps
		return a.stop, nil
	}
	return debug.StopReason{Kind: "steps", PC: uint64(a.live.ARM9PC()), Steps: steps}, nil
}

func (a *Adapter) SetBreakpoint(pc uint64) {
	a.live.AddBreakpoint(uint32(pc))
	for _, b := range a.bps {
		if b == pc {
			return
		}
	}
	a.bps = append(a.bps, pc)
}

func (a *Adapter) ClearBreakpoint(pc uint64) {
	a.live.ClearBreakpoint(uint32(pc))
	for i, b := range a.bps {
		if b == pc {
			a.bps = append(a.bps[:i], a.bps[i+1:]...)
			return
		}
	}
}

func (a *Adapter) Breakpoints() []uint64 { return append([]uint64(nil), a.bps...) }

// Disasm decodes instructions at addr, in whichever state the ARM9 is actually in.
//
// Asking the core rather than assuming ARM is not a nicety on this machine: NitroSDK's
// BIOS call thunks are Thumb, its libraries are full of Thumb, and a disassembly that
// decoded them as ARM would produce plausible instructions that are not there. It is the
// same mistake that made an ARM7 crash look like a corrupt jump table.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	thumb := a.live.Thumb(true)
	out := make([]debug.Instr, 0, n)
	pc := uint32(addr)
	for i := 0; i < n; i++ {
		code := a.live.Snapshot(true, pc, 4)
		var in arm.Inst
		if thumb {
			in = arm.DecodeThumb(code, pc)
		} else {
			in = arm.DecodeARM(code, pc)
		}
		if in.Len == 0 {
			in.Len = 4
		}
		out = append(out, debug.Instr{Addr: uint64(pc), Bytes: code[:in.Len], Text: in.Text})
		pc += uint32(in.Len)
	}
	return out, nil
}

func (a *Adapter) instrAt(pc uint32) string {
	code := a.live.Snapshot(true, pc, 4)
	if a.live.Thumb(true) {
		return arm.DecodeThumb(code, pc).Text
	}
	return arm.DecodeARM(code, pc).Text
}

// --- watches --------------------------------------------------------------------

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "read" && w.Kind != "write" {
		return 0, fmt.Errorf("ndsadapter: watch kind %q is neither read nor write", w.Kind)
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

// applyWatches installs one hook per direction, or none. The hooks are on the machine's
// hot path — every byte the CPU touches — so when there is nothing to watch there must
// be nothing installed, not a hook that returns early.
func (a *Adapter) applyWatches() {
	var reads, writes bool
	for _, w := range a.watches {
		if w.Kind == "read" {
			reads = true
		} else {
			writes = true
		}
	}
	fire := func(kind string, addr uint32, v byte, pc uint32) {
		for i := range a.watches {
			w := &a.watches[i]
			if w.Kind != kind || addr < w.Lo || addr > w.Hi {
				continue
			}
			w.Hits++
			if w.Break {
				a.live.StopRequested()
				a.stop = debug.StopReason{
					Kind: "watch", PC: uint64(pc),
					Note: fmt.Sprintf("%s watch %d at 0x%08X", kind, w.ID, addr),
				}
			}
			if a.watchSink != nil {
				a.watchSink(debug.WatchHit{
					ID: w.ID, Kind: kind, Addr: addr, Val: uint32(v), PC: pc,
					Instr: a.instrAt(pc),
				})
			}
		}
	}
	if writes {
		a.live.OnWrite = func(arm9 bool, addr uint32, v byte, pc uint32) {
			fire("write", addr, v, pc)
		}
	} else {
		a.live.OnWrite = nil
	}
	if reads {
		a.live.OnRead = func(arm9 bool, addr uint32, v byte, pc uint32) {
			fire("read", addr, v, pc)
		}
	} else {
		a.live.OnRead = nil
	}
}

// --- surfaces --------------------------------------------------------------------

func (a *Adapter) Surfaces() []debug.Surface {
	s := []debug.Surface{
		{ID: "screens", Name: "Both screens", W: frameW, H: frameH},
		{ID: "top", Name: "Top screen", W: panelW, H: panelH},
		{ID: "bottom", Name: "Bottom screen", W: panelW, H: panelH},
		{ID: "3d", Name: "3D render target", W: panelW, H: panelH},
		{ID: "depth", Name: "3D depth buffer", W: panelW, H: panelH},
		{ID: "texture", Name: "Texture memory", Free: true, Formats: dsmachine.TextureFormats()},
	}
	// One entry per VRAM bank, named with what it currently IS. A DS bank is not a fixed
	// kind of memory — the same 128 KiB is a background now and a texture after the next
	// write to VRAMCNT — so a bank listed without its mapping is a bank you will look at
	// and misread.
	for i := 0; i < 9; i++ {
		name, size, cnt := a.live.VRAMBankInfo(i)
		s = append(s, debug.Surface{
			ID:   fmt.Sprintf("vram%d", i),
			Name: fmt.Sprintf("VRAM %s (%d KiB, %s)", name, size/1024, vramUse(cnt)),
			Free: true, Formats: dsmachine.TextureFormats(),
		})
	}
	return s
}

// vramUse decodes a VRAMCNT byte into what the bank currently is. The MST encoding is
// per-bank on the hardware, and spelling out all nine tables here would only duplicate
// vram.go — what the debugger's label needs is the honest short answer.
func vramUse(cnt uint8) string {
	if cnt&0x80 == 0 {
		return "disabled (LCDC only)"
	}
	switch cnt & 7 {
	case 0:
		return "LCDC"
	case 1:
		return "background"
	case 2:
		return "sprites / ARM7"
	case 3:
		return "texture or texture palette"
	case 4:
		return "engine B / extended palette"
	case 5:
		return "OBJ extended palette"
	}
	return "?"
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	top, bottom := a.live.Screens()
	switch id {
	case "screens":
		return a.compose(a.live, a.aTop), nil
	case "top":
		return top, nil
	case "bottom":
		return bottom, nil
	case "3d":
		img := image.NewRGBA(image.Rect(0, 0, panelW, panelH))
		if err := a.live.Draw3DInto(img); err != nil {
			return nil, err
		}
		return img, nil
	case "depth":
		return a.live.RenderDepth(), nil
	case "texture":
		format, ok := dsmachine.TextureFormat(v.Format)
		if !ok {
			return nil, fmt.Errorf("ndsadapter: %q is not a DS texture format", v.Format)
		}
		return a.live.RenderTexture(v.Addr, v.Palette, format, uint32(v.W), uint32(v.H))
	}
	if strings.HasPrefix(id, "vram") {
		var i int
		if _, err := fmt.Sscanf(id, "vram%d", &i); err != nil {
			return nil, fmt.Errorf("ndsadapter: no surface %q", id)
		}
		return a.renderBank(i, v)
	}
	return nil, fmt.Errorf("ndsadapter: no surface %q", id)
}

// renderBank draws a VRAM bank's raw bytes as a texture in the requested format. The
// address is an offset into the bank, not a CPU address: a bank's CPU address depends on
// what it is currently mapped as, and the whole reason to look at a bank directly is to
// see it when it is mapped as nothing.
func (a *Adapter) renderBank(i int, v debug.View) (*image.RGBA, error) {
	bank := a.live.VRAMBank(i)
	if bank == nil {
		return nil, fmt.Errorf("ndsadapter: VRAM bank %d does not exist", i)
	}
	w, h := v.W, v.H
	if w <= 0 || h <= 0 {
		w, h = 256, len(bank)/512 // 16-bit pixels, the natural way to look at raw VRAM
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := int(v.Addr) + (y*w+x)*2
			var c color.RGBA
			if o+1 < len(bank) {
				p := uint16(bank[o]) | uint16(bank[o+1])<<8
				c = color.RGBA{
					expand5(uint8(p & 31)), expand5(uint8((p >> 5) & 31)),
					expand5(uint8((p >> 10) & 31)), 255,
				}
			}
			img.Set(x, y, c)
		}
	}
	return img, nil
}

// --- memory map, files, resume ----------------------------------------------------

func (a *Adapter) Regions() []debug.Region {
	rs := a.live.MemRegions()
	out := make([]debug.Region, 0, len(rs))
	for _, r := range rs {
		out = append(out, debug.Region{Name: r.Name, Lo: r.Base, Hi: r.Base + r.Size - 1})
	}
	return out
}

// ListDir browses the cartridge's own filesystem — the FNT directory tree over the FAT.
// A DS cartridge has one, which is why this adapter implements FileLister where the N64's
// (a flat cartridge) cannot.
func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	path = strings.Trim(path, "/")
	seen := map[string]bool{}
	var out []debug.FileEntry
	for _, f := range a.rom.Files {
		child, ok := childOf(f.Path, path)
		if !ok {
			continue
		}
		if i := strings.IndexByte(child, '/'); i >= 0 {
			dir := child[:i]
			if !seen[dir] {
				seen[dir] = true
				full := dir
				if path != "" {
					full = path + "/" + dir
				}
				out = append(out, debug.FileEntry{Path: full, Name: dir, Dir: true})
			}
			continue
		}
		fat := a.rom.FAT[f.ID]
		out = append(out, debug.FileEntry{
			Path: f.Path, Name: child, Size: int64(fat.Size()), Offset: int64(fat.Start),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir != out[j].Dir {
			return out[i].Dir
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// childOf reports the part of path that lies directly under prefix.
func childOf(path, prefix string) (string, bool) {
	if prefix == "" {
		return path, true
	}
	if !strings.HasPrefix(path, prefix+"/") {
		return "", false
	}
	return path[len(prefix)+1:], true
}

func (a *Adapter) ReadFile(path string) ([]byte, error) {
	b := a.rom.FileByPath(strings.Trim(path, "/"))
	if b == nil {
		return nil, fmt.Errorf("ndsadapter: no file %q in the cartridge", path)
	}
	return b, nil
}

// FileAt maps a raw ROM offset back to the file that contains it — so a card-read watch
// can say which file the game is streaming right now, rather than which byte.
func (a *Adapter) FileAt(offset int64) (debug.FileEntry, int64, bool) {
	for _, f := range a.rom.Files {
		fat := a.rom.FAT[f.ID]
		if offset >= int64(fat.Start) && offset < int64(fat.End) {
			name := f.Path
			if i := strings.LastIndexByte(name, '/'); i >= 0 {
				name = name[i+1:]
			}
			return debug.FileEntry{
				Path: f.Path, Name: name,
				Size: int64(fat.Size()), Offset: int64(fat.Start),
			}, offset - int64(fat.Start), true
		}
	}
	return debug.FileEntry{}, 0, false
}

// ResumeArgs is the command line that resumes this game from a savestate — the DS
// oracle's own flags, not a common vocabulary the oracles do not in fact share.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{
		"go", "run", "./extract/cmd/bootoracle",
		"-image", a.imagePath,
		"-loadstate", statePath,
		"-frames", "1",
	}
}

// ResumeDir is the game directory the resume command runs in: the one holding the
// oracle, found by walking up from the image.
func (a *Adapter) ResumeDir() string {
	dir := filepath.Dir(a.imagePath)
	for i := 0; i < 4; i++ {
		if isOracleDir(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(a.imagePath)
}

func isOracleDir(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, "extract", "cmd", "bootoracle"))
	return err == nil && fi.IsDir()
}

// --- profile -----------------------------------------------------------------------

// FrameProfile reports where the last frame's time went. The CPU's share is a remainder
// rather than a measurement — see dsmachine/profile.go for why that is the honest way to
// present it — and the counters ride alongside because milliseconds alone cannot tell a
// faster rasteriser from a frame that drew less.
func (a *Adapter) FrameProfile() debug.FrameProfile {
	p := a.live.FrameProfile()
	return debug.FrameProfile{
		TotalMs: p.TotalMs,
		Buckets: []debug.ProfileBucket{
			{Name: "ARM9 + ARM7", Millis: p.CPUMs},
			{Name: "geometry", Millis: p.GeometryMs, Count: p.Commands},
			{Name: "rasteriser", Millis: p.RasterMs, Count: p.Polygons},
			{Name: "2D compose", Millis: p.ComposeMs},
			{Name: "DMA", Millis: p.DMAMs, Count: p.DMAXfers},
		},
		Counters: []debug.ProfileCounter{
			{Name: "geometry commands", Value: p.Commands},
			{Name: "polygons", Value: p.Polygons},
			{Name: "fragments", Value: p.Fragments},
			{Name: "DMA transfers", Value: p.DMAXfers},
		},
	}
}

// --- helpers -------------------------------------------------------------------------

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}

// The DS works in six bits of colour and five of alpha; the debugger wants eight of each.
func expand6(v uint8) uint8 {
	if v > 63 {
		v = 63
	}
	return v<<2 | v>>4
}

func expand5(v uint8) uint8 {
	if v > 31 {
		v = 31
	}
	return v<<3 | v>>2
}
