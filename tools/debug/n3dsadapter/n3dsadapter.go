// Package n3dsadapter implements a debug.Target for the Nintendo 3DS oracle.
//
// The 3DS is the first target here whose display processor is not a packet
// machine. An N64 frame is a list of RDP commands and a PSX frame is a list of
// GP0 packets, but a PICA200 frame is a stream of register writes: the draw
// happens when the stream writes 0x22E, and everything the draw does — which
// buffers it fetches vertices from, which shader runs, which textures the TEV
// samples, whether it writes colour at all — is state the preceding writes
// latched. So a "command" here is one register write, and the scrubber's job is
// less "watch the triangles land" than "watch the state that produced them".
// That is the right granularity for the question this platform is stuck on:
// Captain Toad rasterises the right number of pixels and shows the wrong ones.
//
// The other structural difference is the frame boundary. The 3DS has no scanout
// the CPU can watch: the GPU renders into a tiled colour buffer, a DisplayTransfer
// copies it into a linear framebuffer, and the GSP points the LCD at that
// framebuffer at the next VBlank. So a frame ends at the VBlank, and what the
// screen shows and what the GPU drew are two different pictures — which is why
// this target offers both as surfaces. The gap between them is where a 3DS frame
// goes missing, and no counter can show it to you.
package n3dsadapter

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/n3ds"
)

// The debugger opens a 3DS game through the registry, so nothing has to import
// this package by name — importing it for its side effect is enough.
func init() {
	debug.Register("3ds", func(s debug.OpenSpec) (debug.Target, error) { return New(s.Image) })
}

// runBudget bounds a single StepFrame or replay. A 3DS video frame is ~4.5M
// instructions of wall-clock pacing, but most of that is idle time the scheduler
// fast-forwards over rather than executes, so a frame's real cost is far lower.
// This leaves generous headroom while still catching a machine that has wedged.
const runBudget = 400_000_000

// The PICA registers with side effects worth naming in a command list: the two
// draw triggers, and the command-buffer jumps that chain one buffer into the next
// (a submitted list is only the head of its chain).
const (
	regDrawArrays = 0x22E
	regDrawElems  = 0x22F
	regCmdBufJump = 0x23C
)

// armRegNames are the ARM register names, in index order.
var armRegNames = []string{
	"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7",
	"r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc",
}

// Adapter drives a 3DS oracle as a debug.Target.
type Adapter struct {
	imagePath string
	img       []byte // kept for the scratch machine the scrubber replays into
	live      *n3ds.Machine
	scratch   *n3ds.Machine // reused across RenderAfter; its state is always disposable

	// mainPass is the colour buffer the last capture's draws landed in — the plane
	// its provenance is recorded in, and so the plane the scrubber must render.
	mainPass target

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64

	// stop is filled in by a hook that halted the run (a breaking watch), so the
	// reason survives the return from Run, which only reports that it stopped.
	stop debug.StopReason
}

// The capabilities this target backs. Listed explicitly so that dropping one from
// the implementation breaks the build here rather than silently removing a panel.
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
)

// snap wraps a 3DS in-memory savestate as an opaque debug.Snapshot.
type snap struct{ ms *n3ds.MachineState }

func (snap) Platform() string { return "3ds" }

// New boots a 3DS oracle on the decrypted CCI at imagePath. The machine is left
// at its entry point; advance it with StepFrame, or — since reaching anything
// worth looking at costs a long boot and, in Captain Toad's case, a scripted walk
// through the StreetPass and file-select screens — jump straight to a prepared
// frame with LoadStateFile.
func New(imagePath string) (*Adapter, error) {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	live, err := n3ds.NewMachine(img)
	if err != nil {
		return nil, err
	}
	return &Adapter{imagePath: imagePath, img: img, live: live}, nil
}

func (a *Adapter) Platform() string { return "3ds" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the live machine, for wiring a caller needs that the generic
// surface does not cover — pad injection (SetKeys), the GX capture log.
func (a *Adapter) Machine() *n3ds.Machine { return a.live }

// Close drops the machines. The 3DS machine holds no OS resources, so this only
// releases the memory for the garbage collector.
func (a *Adapter) Close() error {
	a.live, a.scratch, a.img = nil, nil, nil
	return nil
}

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SnapshotState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ns, ok := s.(snap)
	if !ok {
		return fmt.Errorf("n3dsadapter: snapshot is from %q, not 3ds", platformOf(s))
	}
	a.mainPass = target{} // the restored frame has its own; the previous one's is not it
	return a.live.RestoreState(ns.ms)
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveState(path) }

func (a *Adapter) LoadStateFile(path string) error {
	a.mainPass = target{} // a state load lands in another frame; the old plane means nothing there
	return a.live.LoadState(path)
}

// target identifies one colour buffer a frame drew into.
type target struct{ addr, w, h uint32 }

// StepFrame advances the live machine to the next VBlank, capturing the PICA200
// command-list writes that executed on the way and per-pixel last-writer
// provenance.
//
// A frame can render into more than one colour buffer — a shadow or reflection
// pass into its own target, then the scene into the main one — and provenance is a
// single plane, so it has to be a plane of *one* target. Which one is decided by
// the pixels: the buffer that received the most drawn fragments is the frame's main
// pass, and that is the one reported, in its own coordinates. Pixels drawn into any
// other target are left out rather than folded in, because two targets' coordinates
// mean different things and overlaying them would attribute a pixel to a command
// that never touched the buffer you are looking at.
//
// Choosing by pixel census rather than by "the target bound when the frame ended"
// matters here: Captain Toad's register file, at the VBlank, still points at
// whatever the last command list left behind, which need not be the buffer the
// scene was drawn in.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SnapshotState()}}

	// Provenance is gathered per target, keyed by packed (y<<16|x) while the frame
	// draws, then laid out row-major once the main pass is known.
	prov := map[target]map[uint32]int32{}
	over := map[target]map[uint32][]debug.PixelWrite{}
	drawn := map[target]int{}

	cmd := -1      // index of the command currently executing; -1 before the first
	var cur target // the colour buffer the current draw is writing into

	a.live.OnPICACmd = func(w n3ds.PICAWrite) {
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, picaCommand(cmd, w))
		// A draw trigger fires against the state the preceding writes latched, and
		// this hook runs before the write executes — so the register file already
		// names the buffer the fragments are about to land in.
		if w.Reg == regDrawArrays || w.Reg == regDrawElems {
			addr, tw, th := a.live.GPU().ColorTarget()
			cur = target{addr, tw, th}
		}
	}
	a.live.OnPixel = func(x, y uint32, ev n3ds.PixelEvent) {
		if cmd < 0 {
			return
		}
		key := y<<16 | (x & 0xFFFF)
		if ev.Drawn {
			drawn[cur]++
			p := prov[cur]
			if p == nil {
				p = map[uint32]int32{}
				prov[cur] = p
			}
			p[key] = int32(cmd)
		}
		if withOverdraw {
			o := over[cur]
			if o == nil {
				o = map[uint32][]debug.PixelWrite{}
				over[cur] = o
			}
			o[key] = append(o[key], debug.PixelWrite{
				CmdIndex: cmd, R: ev.R, G: ev.G, B: ev.B, A: ev.A,
				Rejected: !ev.Drawn,
			})
		}
	}
	a.live.OnFrame = func(m *n3ds.Machine) { m.StopRequested = true }

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later
	// run is not paying for a census nobody is reading.
	a.live.OnPICACmd, a.live.OnPixel, a.live.OnFrame = nil, nil, nil

	// The main pass is the target that received the most drawn fragments. With no
	// draws at all, fall back to whatever the register file points at, so the frame
	// still reports the size of the buffer the next draw would land in.
	main, best := target{}, -1
	for t, n := range drawn {
		if n > best {
			main, best = t, n
		}
	}
	if best < 0 {
		addr, tw, th := a.live.GPU().ColorTarget()
		main = target{addr, tw, th}
	}
	a.mainPass = main
	w, h := int(main.w), int(main.h)
	fc.Width, fc.Height = w, h

	if w > 0 && h > 0 && len(prov[main]) > 0 {
		fc.Prov = make([]int32, w*h)
		for i := range fc.Prov {
			fc.Prov[i] = -1
		}
		for key, c := range prov[main] {
			if x, y := int(key&0xFFFF), int(key>>16); x < w && y < h {
				fc.Prov[y*w+x] = c
			}
		}
		if withOverdraw {
			fc.Overdraw = make(map[int][]debug.PixelWrite, len(over[main]))
			for key, writes := range over[main] {
				if x, y := int(key&0xFFFF), int(key>>16); x < w && y < h {
					fc.Overdraw[y*w+x] = writes
				}
			}
		}
	}
	return fc, nil
}

// StepFast advances one frame capturing nothing — no snapshot, no command stream,
// no pixel census. It is how the debugger fast-forwards to the frame worth
// looking at, and on this platform it matters more than most: a 3DS snapshot is a
// gob of every mapped region, and a frame can rasterise millions of fragments.
func (a *Adapter) StepFast() error {
	// Nothing is captured, so there is no main pass any more. Forgetting it matters:
	// a game that double-buffers its render target draws the next frame into the
	// other buffer, and holding on to the last capture's plane would leave the
	// debugger showing a buffer the frame it just ran never wrote.
	a.mainPass = target{}
	a.live.OnPICACmd, a.live.OnPixel = nil, nil
	a.live.OnFrame = func(m *n3ds.Machine) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnFrame = nil
	return nil
}

// RenderAfter replays the frame's start snapshot in the scratch machine, stops the
// command processor after write k, and returns the colour buffer as it stood then.
//
// The picture it returns is the GPU's render target, not the screen: at write k
// the frame has not been display-transferred yet, and on this platform that
// distinction is the whole point.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("n3dsadapter: capture holds a %q snapshot, not 3ds", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("n3dsadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc, err := a.scratchMachine()
	if err != nil {
		return nil, err
	}
	if err := sc.RestoreState(s.ms); err != nil {
		return nil, err
	}
	sc.OnPICACmd, sc.OnPixel, sc.OnFrame = nil, nil, nil
	sc.RunStopAfterPICACommand(k+1, runBudget) // 1-based count → stop after index k

	// Render the plane the capture's provenance lives in, not whatever buffer
	// happens to be bound at command k. Mid-frame the register file passes through
	// other targets (and, between lists, through none at all), and following it
	// would scrub the picture out from under the overlay — the frame would appear
	// to change size and the pixel-to-command mapping would stop meaning anything.
	addr, w, h := a.mainPass.addr, a.mainPass.w, a.mainPass.h
	if w == 0 || h == 0 || int(w) != fc.Width || int(h) != fc.Height {
		addr, w, h = sc.GPU().ColorTarget()
	}
	if w == 0 || h == 0 {
		return nil, fmt.Errorf("n3dsadapter: the GPU has no colour target at command %d", k)
	}
	return toRGBA(sc.RenderTarget(addr, w, h)), nil
}

// picaCommand renders one command-list register write as a debug command. The
// draw triggers and the command-buffer jumps get their names, because those are
// the writes with side effects; everything else is named by the register group it
// belongs to (rasterizer, tev, framebuffer, shader…), which is how you find the
// state a draw was made under by scrubbing backwards from it.
func picaCommand(idx int, w n3ds.PICAWrite) debug.GPUCommand {
	name := n3ds.PICARegGroup(w.Reg)
	switch w.Reg {
	case regDrawArrays:
		name = "DRAWARRAYS"
	case regDrawElems:
		name = "DRAWELEMENTS"
	case regCmdBufJump, regCmdBufJump + 1:
		name = "CMDBUF_JUMP"
	}
	return debug.GPUCommand{
		Index:   idx,
		Name:    name,
		Op:      uint32(w.Reg),
		Words:   []uint64{uint64(w.Value)},
		Decoded: fmt.Sprintf("0x%03X = 0x%08X (mask %X)", w.Reg, w.Value, w.Mask),
	}
}

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 16)
	for i := 0; i < 16; i++ {
		vals[i] = uint64(c.R[i])
	}
	thumb := uint64(0)
	if c.Thumb {
		thumb = 1
	}
	return debug.CPUReg{
		PC:    uint64(c.PC()),
		Names: armRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"cpsr":   uint64(c.CPSR()),
			"mode":   uint64(c.Mode),
			"thumb":  thumb,
			"fpscr":  uint64(c.VFP.FPSCR),
			"instrs": c.Instrs,
			"vblank": a.live.VBlanks(),
			"draws":  uint64(a.live.GPU().Draws),
		},
	}
}

// Display is the frame's picture: the colour buffer the PICA rendered into.
//
// On this platform that is not the same plane as the scanout, and the difference
// is not cosmetic. The GPU draws into a tiled, padded VRAM buffer (Captain Toad's
// is 256×512 at 0x1F000000); a DisplayTransfer later copies a 240×400 crop of it
// into a linear framebuffer, rotated, and the GSP points the LCD at that. The two
// pictures have different sizes, different orientations and different contents —
// the screen is always at least one transfer behind the GPU.
//
// The debugger builds a frame around the plane the frame's commands actually wrote,
// because that is the plane provenance is recorded in: a click on this image maps
// to the command that drew that pixel. Handing back the scanout instead would put
// the overlay in a coordinate system the provenance buffer does not share, and the
// answers would be wrong rather than absent. The scanout is not lost — it is the
// "scanout" surface, and comparing the two is how you catch a frame that the GPU
// drew and the transfer never delivered.
func (a *Adapter) Display() (*image.RGBA, error) {
	// After a capture, this is the plane that capture's provenance is in — the same
	// buffer the scrubber renders — so a click on the picture and the overlay drawn
	// over it are talking about the same pixels. Between captures, it is whatever
	// the register file currently points at.
	addr, w, h := a.mainPass.addr, a.mainPass.w, a.mainPass.h
	if w == 0 || h == 0 {
		addr, w, h = a.live.GPU().ColorTarget()
	}
	if w == 0 || h == 0 {
		// Before the first draw there is no render target; the screen is all there
		// is to show.
		if img := a.live.Framebuffer("top"); img != nil {
			return toRGBA(img), nil
		}
		return nil, fmt.Errorf("n3dsadapter: nothing rendered yet (no draw and no DisplayTransfer)")
	}
	return toRGBA(a.live.RenderTarget(addr, w, h)), nil
}

func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	if n <= 0 {
		return nil
	}
	return a.live.ReadBytes(addr, uint32(n))
}

// Watches.
//
// The bus is byte-granular — the ARM core issues four Read calls for one LDR — so
// a watch on a word reports four hits for a single store, each with the PC that
// made it. The PC is the question a watch is asked to answer, and it is the same
// for all four.
//
// The machine has one write window and one read window, so the target holds one
// watch of each kind. Asking for a second of the same kind is an error rather than
// a silent replacement: quietly dropping a watch the user set is how a debugger
// lies to you.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("n3dsadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("n3dsadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

// Breakpoints. The machine can only clear all of them at once, so the adapter
// keeps the set and re-applies it.

func (a *Adapter) SetBreakpoint(pc uint64) {
	for _, b := range a.bps {
		if b == pc {
			return
		}
	}
	a.bps = append(a.bps, pc)
	a.live.AddBreakpoint(uint32(pc))
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
		a.live.AddBreakpoint(uint32(b))
	}
}

func (a *Adapter) Breakpoints() []uint64 { return append([]uint64(nil), a.bps...) }

func (a *Adapter) StepInstr(n int) (debug.StopReason, error) {
	if n <= 0 {
		n = 1
	}
	return a.runSlice(n, "steps")
}

func (a *Adapter) Continue(budget uint64) (debug.StopReason, error) {
	if budget == 0 || budget > runBudget {
		budget = runBudget
	}
	return a.runSlice(int(budget), "budget")
}

// runSlice runs the machine and works out why it stopped. The machine reports a
// breakpoint through its own stopped flag and a halt through the CPU; a breaking
// watch reports itself through a.stop, because from Run's point of view a hook
// merely asked to stop.
func (a *Adapter) runSlice(budget int, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	steps := a.live.Run(budget)
	sr := a.stop
	if sr.Kind == "" {
		switch {
		case a.live.CPU.Halted:
			sr = debug.StopReason{Kind: "halted", Note: a.live.HaltReason()}
		case a.live.Stopped():
			sr = debug.StopReason{Kind: "breakpoint"}
		default:
			sr = debug.StopReason{Kind: exhausted}
		}
	}
	sr.PC = uint64(a.live.CPU.PC())
	sr.Steps = uint64(steps)
	return sr, nil
}

// Disasm decodes n instructions at a virtual address. The ARM11 has two
// instruction sets and nothing in memory says which one an address holds, so this
// decodes in whichever the core is currently in — correct for the disassembly
// around the PC, which is what the pane is for.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	thumb := a.live.CPU.Thumb
	width := uint64(4)
	if thumb {
		width = 2
	}
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr + uint64(i)*width)
		b := a.live.ReadBytes(va, uint32(width))
		in := arm.DecodeVariant(b, va, thumb, arm.V6K)
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: b, Text: in.Text})
	}
	return out, nil
}

// instrAt is the one-line disassembly at pc, used to say what wrote a watched
// address. An undecodable pc gives "", not an error: a watch report is worth
// having even where we cannot decode.
func (a *Adapter) instrAt(pc uint64) string {
	in, err := a.Disasm(pc, 1)
	if err != nil || len(in) == 0 {
		return ""
	}
	return in[0].Text
}

// Surfaces are the pictures this machine can show, and on the 3DS there are three
// different pictures that all get called "the frame":
//
//   - the two screens, which are the linear framebuffers the LCD is scanning —
//     what the player sees, one DisplayTransfer behind the GPU;
//   - the draw target, which is the tiled colour buffer the PICA is rendering
//     into right now — what the GPU thinks it drew;
//   - the depth buffer, which says whether a pass that produced no visible pixels
//     was nonetheless testing and writing depth over someone else's memory.
//
// And then the free one: any address, read as a texture in any PICA format. Point
// it at the address a texture unit is configured with and you see whether what the
// unit will sample is the texture the game meant to upload — which is the question
// a garbled render eventually comes down to.
func (a *Adapter) Surfaces() []debug.Surface {
	tw, th := 0, 0
	if img := a.live.Framebuffer("top"); img != nil {
		tw, th = img.Rect.Dx(), img.Rect.Dy()
	}
	bw, bh := 0, 0
	if img := a.live.Framebuffer("bottom"); img != nil {
		bw, bh = img.Rect.Dx(), img.Rect.Dy()
	}
	_, cw, ch := a.live.GPU().ColorTarget()
	return []debug.Surface{
		{ID: "scanout", Name: "Top screen", W: tw, H: th},
		{ID: "bottom", Name: "Bottom screen", W: bw, H: bh},
		{ID: "drawtarget", Name: "PICA draw target", W: int(cw), H: int(ch)},
		{ID: "depth", Name: "Depth buffer", W: int(cw), H: int(ch)},
		{ID: "texture", Name: "Memory as a texture", Free: true, Formats: n3ds.TextureFormats()},
	}
}

// RenderSurface draws one. For the free "texture" surface, View.Addr is a virtual
// address and View.Format is a PICA texture format name; Stride and Palette are
// unused, because a PICA texture is tiled and carries no palette.
func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	g := a.live.GPU()
	switch id {
	case "scanout":
		img := a.live.Framebuffer("top")
		if img == nil {
			return nil, fmt.Errorf("n3dsadapter: no frame presented yet (the game has done no DisplayTransfer)")
		}
		return toRGBA(img), nil
	case "bottom":
		img := a.live.Framebuffer("bottom")
		if img == nil {
			return nil, fmt.Errorf("n3dsadapter: the bottom screen has not been presented yet")
		}
		return toRGBA(img), nil
	case "drawtarget":
		addr, w, h := g.ColorTarget()
		if w == 0 || h == 0 {
			return nil, fmt.Errorf("n3dsadapter: the GPU has no colour target yet")
		}
		return toRGBA(a.live.RenderTarget(addr, w, h)), nil
	case "depth":
		addr, w, h := g.DepthTarget()
		if w == 0 || h == 0 {
			return nil, fmt.Errorf("n3dsadapter: the GPU has no depth target yet")
		}
		return toRGBA(a.live.RenderDepth(addr, w, h)), nil
	case "texture":
		f, ok := n3ds.TextureFormat(v.Format)
		if !ok {
			return nil, fmt.Errorf("n3dsadapter: %q is not a PICA texture format: %w", v.Format, debug.ErrUnsupported)
		}
		img, err := a.live.RenderTexture(v.Addr, f, uint32(v.W), uint32(v.H))
		if err != nil {
			return nil, err
		}
		return toRGBA(img), nil
	}
	return nil, fmt.Errorf("n3dsadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// Regions names the mapped address space for the memory pane. Unlike a cartridge
// machine's, a 3DS process's map is laid out by the loader, so it is asked for
// rather than hard-coded.
func (a *Adapter) Regions() []debug.Region {
	var out []debug.Region
	for _, r := range a.live.MemRegions() {
		out = append(out, debug.Region{Name: r.Name, Lo: r.Base, Hi: r.Base + r.Size})
	}
	return out
}

// The cartridge filesystem. Offsets are byte offsets into the RomFS level-3
// image — the same coordinates the game's own file reads use, so a read-watch on
// the RomFS region can be named back to a file.

func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	fs := a.live.RomFS()
	if fs == nil {
		return nil, fmt.Errorf("n3dsadapter: this image has no RomFS: %w", debug.ErrUnsupported)
	}
	if path == "" {
		path = "/"
	}
	prefix := strings.TrimSuffix(path, "/") + "/"

	var out []debug.FileEntry
	seen := map[string]bool{}
	// The RomFS is flattened, so the children of a directory are the entries whose
	// path starts with it and has no further separator.
	for _, d := range fs.Dirs {
		if child, ok := childOf(d, prefix); ok && !seen[child] {
			seen[child] = true
			out = append(out, debug.FileEntry{Path: prefix + child, Name: child, Dir: true})
		}
	}
	for _, f := range fs.Files {
		if child, ok := childOf(f.Path, prefix); ok && !seen[child] {
			seen[child] = true
			out = append(out, debug.FileEntry{
				Path: f.Path, Name: child, Size: f.Size, Offset: fs.L3Offset(f),
			})
		}
	}
	return out, nil
}

// childOf returns the immediate child of prefix that path lies under, if it is an
// immediate child at all.
func childOf(path, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	rest := path[len(prefix):]
	if rest == "" || strings.Contains(rest, "/") {
		return "", false
	}
	return rest, true
}

func (a *Adapter) ReadFile(path string) ([]byte, error) {
	fs := a.live.RomFS()
	if fs == nil {
		return nil, fmt.Errorf("n3dsadapter: this image has no RomFS: %w", debug.ErrUnsupported)
	}
	return fs.File(path)
}

// FileAt names the file containing an offset in the RomFS. The offset is a
// level-3 filesystem offset — the coordinates the game's own reads use, and the
// ones ListDir reports — so a read-watch on the RomFS can say which file the
// machine is streaming rather than which byte.
func (a *Adapter) FileAt(offset int64) (debug.FileEntry, int64, bool) {
	fs := a.live.RomFS()
	if fs == nil {
		return debug.FileEntry{}, 0, false
	}
	f, within, ok := fs.FileAt(offset)
	if !ok {
		return debug.FileEntry{}, 0, false
	}
	return debug.FileEntry{
		Path:   f.Path,
		Name:   f.Path[strings.LastIndex(f.Path, "/")+1:],
		Size:   f.Size,
		Offset: fs.L3Offset(f),
	}, within, true
}

// ResumeArgs is the command line that reopens this image at a saved state. The
// 3DS bootoracle spells the flag -loadstate.
func (a *Adapter) ResumeArgs(statePath string) []string {
	image, _ := filepath.Abs(a.imagePath)
	state, _ := filepath.Abs(statePath)
	return []string{"go", "run", "./cmd/bootoracle", "-image", image, "-loadstate", state}
}

// ResumeDir is where that command runs. The two 3DS titles share one oracle —
// Captain Toad has no extract/ of its own and boots on Super Mario 3D Land's
// bootoracle — so the directory cannot be assumed to be this game's own.
func (a *Adapter) ResumeDir() string {
	// games/<slug>/image/<file>.cci → games/<slug>
	gameDir := filepath.Dir(filepath.Dir(a.imagePath))
	if abs, err := filepath.Abs(gameDir); err == nil {
		gameDir = abs
	}
	if isOracleDir(filepath.Join(gameDir, "extract")) {
		return filepath.Join(gameDir, "extract")
	}
	// Fall back to a sibling 3DS game's oracle.
	siblings, _ := filepath.Glob(filepath.Join(filepath.Dir(gameDir), "*-3ds", "extract"))
	for _, s := range siblings {
		if isOracleDir(s) {
			return s
		}
	}
	return ""
}

func isOracleDir(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "cmd", "bootoracle"))
	return err == nil && st.IsDir()
}

// scratchMachine is the machine RenderAfter replays into, so the live one keeps
// its place while the scrubber runs the frame again.
func (a *Adapter) scratchMachine() (*n3ds.Machine, error) {
	if a.scratch == nil {
		sc, err := n3ds.NewMachine(a.img)
		if err != nil {
			return nil, err
		}
		a.scratch = sc
	}
	return a.scratch, nil
}

// toRGBA converts the platform's NRGBA output to the RGBA the debug surface uses.
// Every image the 3DS renderers produce is fully opaque, so this is a straight
// copy — but it goes through the image package's own conversion rather than
// assuming that, because an alpha the renderer did not intend is a bug worth
// seeing rather than one worth pre-multiplying away silently.
func toRGBA(src *image.NRGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	dst := image.NewRGBA(src.Bounds())
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	return dst
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
