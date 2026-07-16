// Package gcadapter implements a debug.Target for the Nintendo GameCube oracle.
//
// It owns a live gc.Machine and, for command replay, a scratch machine it restores into.
// Almost everything the debugger needs is a hook the machine already offers (OnGXCmd,
// OnPixel, OnFlip, the watch windows, StopRequested) or an in-memory savestate; this
// package is the thin translation between those and the platform-agnostic debug surface.
//
// Two things about the GameCube shape the adapter, and both are the general lessons this
// suite has learned twice already, arriving here in new clothes:
//
// THE FRAME IS THE COPY, NOT THE FIELD. Flipper does not render into main memory. It draws
// into an embedded framebuffer on the die — the EFB — and the pixel engine then copies that
// out to the external framebuffer the video interface scans. The copy is the flip, and it
// is where a frame ends. The video field is NOT: it is a scanout clock that ticks at 60 Hz
// whether or not the game drew anything, and, worse, the copy that ends a frame almost
// always clears the EFB behind itself, so a capture taken at the field boundary would find
// the frame's own draw target already wiped. So StepFrame runs to OnFlip and the picture it
// reports is the EFB. This is "render the DRAW TARGET, not the scanout" for a console where
// the two are not even the same kind of memory: the EFB is 640x528 RGBA on the GPU, the XFB
// is YUY2 in RAM.
//
// A COMMAND IS ONE FIFO COMMAND. The GameCube is a packet machine like the N64 and the PSX,
// not a register-write machine like the PICA or the GE — a draw carries its own vertices in
// the stream. A field of Luigi's Mansion's forest is ~10k commands, most of them draws, so
// the scrubber is a tour of the scene assembling rather than a crawl through state-setting.
// Stopping mid-list needs no sentinel panic, though: the FIFO is drained by an ordinary Go
// loop that the write-gather pipe feeds, so the interpreter can simply decline the next
// command (gc.RunStopAfterGXCommand).
package gcadapter

import (
	"encoding/binary"
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/gekko"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/gc"
)

// The debugger opens a GameCube game through the registry, so nothing has to import this
// package by name — importing it for its side effect is enough.
func init() {
	debug.Register("gc", func(s debug.OpenSpec) (debug.Target, error) { return New(s.Image) })
}

// runBudget bounds a single StepFrame/replay. A video field is ~2.8M instructions and a
// field of the forest cutscene draws ~10k commands, so this leaves generous headroom while
// still catching a machine that has wedged. A cold boot to the title is ~1.5 billion — far
// past this — which is why the debugger starts from a savestate rather than from reset.
const runBudget = 400_000_000

// gekkoRegNames are the PowerPC EABI names of the 32 general registers. They are just r0..r31
// on this ABI, but two carry conventions worth seeing at a glance in the register pane: r1 is
// the stack pointer (the back-chain the backtrace walks) and r13 the small-data base that
// every one of this game's globals is addressed off.
var gekkoRegNames = func() []string {
	n := make([]string, 32)
	for i := range n {
		n[i] = fmt.Sprintf("r%d", i)
	}
	n[1] = "r1/sp"
	n[13] = "r13/sda"
	return n
}()

// Adapter drives a GameCube oracle as a debug.Target.
type Adapter struct {
	imagePath string
	live      *gc.Machine
	scratch   *gc.Machine // reused across RenderAfter; its state is always disposable

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64

	// stop is filled in by a hook that halted the run (a breaking watch), so the reason
	// survives the return from Run, which only reports "stop requested".
	stop debug.StopReason

	// The pad. held is which keys the browser currently has down; padQueue is the button
	// states the machine has not sampled yet. See Key.
	held     map[string]bool
	padQueue []uint16
}

// The capabilities this target backs. Listed explicitly so that dropping one from the
// implementation breaks the build here rather than silently removing a panel.
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
	_ debug.Keyer          = (*Adapter)(nil)
)

// snap wraps a GameCube in-memory savestate as an opaque debug.Snapshot.
type snap struct{ ms gc.MachineState }

func (snap) Platform() string { return "gc" }

// New opens the disc at imagePath and runs its apploader, leaving the machine at the game's
// entry point. It does NOT boot on to a frame: that is ~1.5 billion instructions away, so
// the debugger is meant to be pointed at a savestate (LoadStateFile) rather than made to
// sit through the boot.
func New(imagePath string) (*Adapter, error) {
	disc, err := gc.Open(imagePath)
	if err != nil {
		return nil, err
	}
	live, err := gc.NewMachine(disc)
	if err != nil {
		return nil, err
	}
	if _, err := live.RunApploader(); err != nil {
		return nil, fmt.Errorf("apploader: %w", err)
	}
	// The OS parks in cooperative wait loops that the spin heuristic reads as a wedge, and
	// under a debugger a paused machine is a normal state rather than a bug to report.
	live.SetSpinDetect(false)
	a := &Adapter{imagePath: imagePath, live: live, held: map[string]bool{}}
	a.installPadPacing(live)
	return a, nil
}

func (a *Adapter) Platform() string { return "gc" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the underlying live machine, for wiring a caller needs that the generic
// interface does not cover.
func (a *Adapter) Machine() *gc.Machine { return a.live }

// Close drops the machines. The disc file handle goes with them.
func (a *Adapter) Close() error {
	if a.live != nil {
		if d := a.live.Disc(); d != nil {
			d.Close()
		}
	}
	a.live, a.scratch = nil, nil
	return nil
}

func (a *Adapter) LoadStateFile(path string) error { return a.live.LoadStateFile(path) }
func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ns, ok := s.(snap)
	if !ok {
		return fmt.Errorf("gcadapter: snapshot is from %q, not gc", platformOf(s))
	}
	return a.live.LoadState(ns.ms)
}

// Display is what the video interface is scanning out: the external framebuffer, decoded
// from YUY2. It is deliberately not the same picture StepFrame reports — see the package
// comment — and comparing the two catches a frame the GPU drew that the copy never
// delivered.
func (a *Adapter) Display() (*image.RGBA, error) { return a.live.RenderXFB() }

// ReadMem reads main memory at a physical address. It reads RAM only, and out-of-range
// bytes read as zero: the hardware registers are in the address space too, but reading one
// has side effects — a mailbox read pops the mail — and a debugger's memory pane must never
// be the thing that changes the machine.
func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	ram := a.live.RAM
	for i := 0; i < n; i++ {
		if p := int(addr) + i; p >= 0 && p < len(ram) {
			out[i] = ram[p]
		}
	}
	return out
}

// ---- frames ----

// StepFrame advances the live machine to the next flip — the pixel engine's copy of the
// embedded framebuffer out to the external one — capturing that frame's FIFO command stream
// and per-pixel last-writer provenance.
//
// The picture is the EFB, captured inside the copy before it clears. Everything else here is
// the N64 adapter's shape.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SaveState()}}
	fc.CountWrites()

	prov := map[uint32]int32{}
	var over map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[uint32][]debug.PixelWrite{}
	}
	cmd := -1 // index of the command currently drawing; -1 before the first

	a.live.OnGXCmd = func(_ *gc.Machine, op uint8, words []uint32) {
		cmd = len(fc.Commands)
		w := make([]uint64, len(words))
		for i, v := range words {
			w[i] = uint64(v)
		}
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index:   cmd,
			Name:    gc.GXName(op),
			Op:      uint32(op),
			Words:   w,
			Decoded: decodeCmd(op, words),
		})
	}
	a.live.OnPixel = func(x, y int, ev gc.PixelEvent) {
		if cmd < 0 || x < 0 || y < 0 {
			return
		}
		key := uint32(y)<<16 | uint32(x)&0xFFFF
		if ev.Drawn {
			fc.MarkWrite(cmd)
			prov[key] = int32(cmd)
		}
		if over != nil {
			over[key] = append(over[key], debug.PixelWrite{
				CmdIndex: cmd, R: ev.R, G: ev.G, B: ev.B, A: ev.A,
				Rejected: !ev.Drawn,
			})
		}
	}

	// The frame's picture has to be taken inside the flip: the same copy that delivers the
	// frame clears the buffer it was drawn in, so by the time the run loop regains control
	// the EFB is already blank.
	var shot *image.RGBA
	a.live.OnFlip = func(m *gc.Machine) {
		if shot == nil {
			shot, _ = m.RenderEFB()
			m.StopRequested = true
		}
	}

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later run is
	// not paying for a census nobody is reading.
	a.live.OnGXCmd, a.live.OnPixel, a.live.OnFlip = nil, nil, nil

	w, h := 0, 0
	if shot != nil {
		w, h = shot.Bounds().Dx(), shot.Bounds().Dy()
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

// StepFast advances the live machine one flip, capturing nothing: no frame-start snapshot
// (a 40 MiB copy of RAM and ARAM), no command stream, no per-pixel census. It is how the
// debugger fast-forwards to the part of the game worth looking at.
func (a *Adapter) StepFast() error {
	a.live.OnGXCmd, a.live.OnPixel = nil, nil
	done := false
	a.live.OnFlip = func(m *gc.Machine) {
		if !done {
			done = true
			m.StopRequested = true
		}
	}
	a.live.Run(runBudget)
	a.live.OnFlip = nil
	return nil
}

// RenderAfter replays fc.Start in the scratch machine and returns the draw target exactly
// after command k executed.
//
// The last command of a frame is the copy that ends it, and that copy usually clears the
// EFB on its way out — so a naive replay to the end of the stream renders a blank buffer and
// the scrubber's final position, the one showing the finished frame, would be black. The
// flip hook catches the picture mid-copy, before the clear, exactly as StepFrame does.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("gcadapter: capture holds a %q snapshot, not gc", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("gcadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc, err := a.scratchMachine()
	if err != nil {
		return nil, err
	}
	if err := sc.LoadState(s.ms); err != nil {
		return nil, err
	}
	sc.OnGXCmd, sc.OnPixel = nil, nil
	var shot *image.RGBA
	sc.OnFlip = func(m *gc.Machine) {
		if shot == nil {
			shot, _ = m.RenderEFB()
		}
	}
	sc.RunStopAfterGXCommand(k+1, runBudget) // 1-based count → stop after index k
	sc.OnFlip = nil
	if shot != nil {
		return shot, nil
	}
	return sc.RenderEFB()
}

// ---- CPU ----

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 32)
	for i := 0; i < 32; i++ {
		vals[i] = uint64(c.GPR[i])
	}
	s := c.Snapshot()
	return debug.CPUReg{
		PC:    uint64(c.PC),
		Names: gekkoRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"LR": uint64(c.LR), "CTR": uint64(c.CTR), "CR": uint64(s.CR),
			"XER": uint64(s.XER), "MSR": uint64(c.MSR),
			"SRR0": uint64(s.SRR0), "SRR1": uint64(s.SRR1),
			"TB": c.TB, "DEC": uint64(s.DEC),
		},
	}
}

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
// breakpoint, a halt or a spin; a breaking watch reports itself through a.stop, because from
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
	sr.PC = uint64(a.live.CPU.PC)
	sr.Steps = res.Steps
	return sr, nil
}

// Breakpoints. The machine can only clear all of them at once, so the adapter keeps the set
// and re-applies it.

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

// Disasm decodes n instructions at a virtual address. PowerPC has no delay slots and fixed
// four-byte instructions, so this is simply a walk.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*4
		w := a.live.ReadVirt32(va)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], w)
		in := gekko.DecodeWord(w, va)
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: b[:], Text: in.Text})
	}
	return out, nil
}

// instrAt is the one-line disassembly at pc, used to say what wrote a watched address.
func (a *Adapter) instrAt(pc uint64) string {
	in, err := a.Disasm(pc, 1)
	if err != nil || len(in) == 0 {
		return ""
	}
	return in[0].Text
}

// ---- watches ----
//
// The machine has exactly one write window and one read window, so a target can hold one
// watch of each kind at a time. Asking for a second of the same kind is an error rather than
// a silent replacement — quietly dropping a watch the user set is how a debugger lies.
//
// The windows see CPU stores only. A DMA — the disc landing a file, the DSP fetching a voice
// — writes straight into RAM and no watch will see it, which is a real limit of this model
// and is why a watch that never fires is not proof nobody wrote the address.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("gcadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("gcadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

// ---- surfaces ----

// Surfaces are the pictures this machine can show. The first two are the pair the package
// comment is about — what Flipper drew into, and what the video interface is scanning — and
// on this console they are different memories in different formats, so seeing both is how
// you tell a drawing bug from a copy bug.
func (a *Adapter) Surfaces() []debug.Surface {
	sw, sh := 0, 0
	if img, err := a.live.RenderXFB(); err == nil {
		sw, sh = img.Rect.Dx(), img.Rect.Dy()
	}
	ew, eh := 0, 0
	if img, err := a.live.RenderEFB(); err == nil {
		ew, eh = img.Rect.Dx(), img.Rect.Dy()
	}
	out := []debug.Surface{
		{ID: "drawtarget", Name: "EFB (Flipper draw target)", W: ew, H: eh},
		{ID: "scanout", Name: "XFB (VI scanout)", W: sw, H: sh},
		{ID: "ram", Name: "Main RAM as a texture", Free: true, Formats: gc.RegionFormats()},
	}
	// The eight texture maps, each shown at whatever size and format the game currently has
	// bound. A map has to be ASKED whether it is bound: an untouched one's registers read
	// back zero and the size fields are biased by one, so it decodes as a 1x1 texture rather
	// than as nothing, and offering those would be six phantom surfaces in the list.
	for i := 0; i < 8; i++ {
		if !a.live.TextureBound(i) {
			continue
		}
		img, err := a.live.DumpTexture(i)
		if err != nil {
			continue
		}
		out = append(out, debug.Surface{
			ID:   fmt.Sprintf("tex%d", i),
			Name: fmt.Sprintf("Bound texture %d", i),
			W:    img.Rect.Dx(), H: img.Rect.Dy(),
		})
	}
	return out
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "scanout":
		return a.live.RenderXFB()
	case "drawtarget":
		return a.live.RenderEFB()
	case "ram":
		return a.live.RenderRegion(gc.RegionSpec{
			Addr: v.Addr, W: v.W, H: v.H, Format: v.Format, Palette: v.Palette,
		})
	}
	if strings.HasPrefix(id, "tex") {
		if i, err := strconv.Atoi(strings.TrimPrefix(id, "tex")); err == nil && i >= 0 && i < 8 {
			return a.live.DumpTexture(i)
		}
	}
	return nil, fmt.Errorf("gcadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// ---- the disc's filesystem ----
//
// The GameCube's FST is a real filesystem, and this game streams out of it constantly, so
// the file panes earn their keep here more than on any cartridge platform: a read-watch can
// say which of the game's own files the drive is delivering right now.

// ListDir lists one directory of the disc. The FST is flat — a directory is an entry with a
// span rather than a node with children — so the listing is synthesized from the paths.
func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	fst := a.live.Disc().FST
	if fst == nil {
		return nil, debug.ErrUnsupported
	}
	prefix := path
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if prefix == "//" {
		prefix = "/"
	}
	seen := map[string]bool{}
	var out []debug.FileEntry
	for _, e := range fst.Entries {
		if e.Path == "/" || !strings.HasPrefix(e.Path, prefix) {
			continue
		}
		rest := strings.TrimPrefix(e.Path, prefix)
		if rest == "" {
			continue
		}
		// Only the immediate children: anything with a further slash belongs to a
		// subdirectory, and the subdirectory itself is the entry to list.
		if i := strings.Index(rest, "/"); i >= 0 {
			name := rest[:i]
			if !seen[name] {
				seen[name] = true
				out = append(out, debug.FileEntry{Path: prefix + name, Name: name, Dir: true})
			}
			continue
		}
		if seen[rest] {
			continue
		}
		seen[rest] = true
		out = append(out, debug.FileEntry{
			Path: e.Path, Name: e.Name, Dir: e.Dir, Size: e.Size, Offset: e.Offset,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir != out[j].Dir {
			return out[i].Dir // directories first
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// ReadFile reads a file off the disc by path.
func (a *Adapter) ReadFile(path string) ([]byte, error) {
	disc := a.live.Disc()
	if disc == nil || disc.FST == nil {
		return nil, debug.ErrUnsupported
	}
	f, ok := disc.FST.ByPath(path)
	if !ok {
		return nil, fmt.Errorf("gcadapter: no file %q on the disc: %w", path, debug.ErrUnsupported)
	}
	if f.Dir {
		return nil, fmt.Errorf("gcadapter: %q is a directory: %w", path, debug.ErrUnsupported)
	}
	return disc.Read(f.Offset, int(f.Size))
}

// FileAt names the file that contains a raw disc offset — what turns a read-watch on the
// drive into "the game is streaming this file right now".
func (a *Adapter) FileAt(offset int64) (debug.FileEntry, int64, bool) {
	disc := a.live.Disc()
	if disc == nil || disc.FST == nil {
		return debug.FileEntry{}, 0, false
	}
	f, within, ok := disc.FST.ByOffset(offset)
	if !ok {
		return debug.FileEntry{}, 0, false
	}
	return debug.FileEntry{
		Path: f.Path, Name: f.Name, Dir: f.Dir, Size: f.Size, Offset: f.Offset,
	}, within, true
}

// ---- memory map ----

// Regions names the parts of the address space the memory pane can show. Only main memory
// is listed, because only main memory is what ReadMem serves: ARAM is not in the address
// space at all (it is reachable only through the DSP's DMA engine), and the hardware
// registers are not readable without side effects.
func (a *Adapter) Regions() []debug.Region {
	return []debug.Region{
		{Name: "OS low memory", Lo: 0x0000_0000, Hi: 0x0000_3100},
		{Name: "Main RAM", Lo: 0x0000_3100, Hi: 0x0180_0000},
	}
}

// ResumeArgs is the command line that resumes this disc from a saved state. Luigi's
// Mansion's bootoracle spells the flag -loadstate, and -nospin because the OS's cooperative
// wait loops look like wedges to the spin heuristic.
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
// The GameCube's controller is the Keyer's second platform, and the first where a "key" is a
// console button. It fits the interface — a button has no place on the picture, so it
// carries an identity rather than coordinates — but its mechanics are the opposite of the
// DOS keyboard's, and the difference is the whole of the code below.
//
// A PC keyboard DELIVERS EVENTS: the make and the break are each a scancode the guest's ISR
// consumes, so the adapter's job is to enqueue both and never coalesce. The GameCube's pad
// IS A LEVEL: the serial interface auto-polls it once per video field and latches whatever
// mask is set at that instant, and the game edge-detects presses from the polls. Nothing
// queues on the guest's side, so nothing self-paces.
//
// That means a browser key that goes down and up between two field polls is a press the game
// never sees — and the debugger's fields are wall-clock-slow, so that is not a hypothetical.
// It is exactly the hazard the touch latch was built for, arriving from the other side: there
// the risk was too many states for the frames available, here it is a state with no field to
// be sampled in. So the states are queued and released one per field, which guarantees every
// distinct mask is live across at least one poll. A held key contributes to every state until
// it is released, so a drag-equivalent (holding A while pressing Start) still composes.
func (a *Adapter) Key(k debug.Key) error {
	name := normalizeKey(k.Name)
	if _, ok := gc.PadButton(name); !ok {
		return nil // an unmapped key is ignored rather than an error: the browser sends everything
	}
	if k.Down {
		a.held[name] = true
	} else {
		delete(a.held, name)
	}
	var mask uint16
	for n := range a.held {
		b, _ := gc.PadButton(n)
		mask |= b
	}
	// Only a CHANGE needs a field of its own; a repeat (the browser's key-repeat) is the
	// same level and would just cost a field.
	if len(a.padQueue) > 0 && a.padQueue[len(a.padQueue)-1] == mask {
		return nil
	}
	a.padQueue = append(a.padQueue, mask)
	return nil
}

// installPadPacing releases one queued button state per video field. OnDisplay runs at the
// field boundary just after the serial interface has auto-polled, so a state set here is
// latched by the next poll and is live for exactly one field before the next state replaces
// it. An empty queue leaves the current mask alone, which is what keeps a held button held.
//
// It chains any hook already installed rather than replacing it: OnDisplay is the machine's
// one field-boundary hook and the pad has no claim to own it.
func (a *Adapter) installPadPacing(m *gc.Machine) {
	prev := m.OnDisplay
	m.OnDisplay = func(mm *gc.Machine) {
		if prev != nil {
			prev(mm)
		}
		if len(a.padQueue) == 0 {
			return
		}
		mm.SetPadButtons(0, a.padQueue[0])
		a.padQueue = a.padQueue[1:]
	}
}

// normalizeKey folds a browser KeyboardEvent.key value to the button names gc.PadButton
// knows — the same names the oracle's -keys scripts use. The letter keys map to the button
// of the same name (A is a, Start is Enter), which is the mapping that needs no legend.
func normalizeKey(name string) string {
	s := strings.ToLower(name)
	switch s {
	case "arrowup":
		return "up"
	case "arrowdown":
		return "down"
	case "arrowleft":
		return "left"
	case "arrowright":
		return "right"
	case "enter", "return":
		return "start"
	}
	return s
}

// ---- helpers ----

// decodeCmd is the one-line human decode for a FIFO command: the register a load names, the
// vertex count a draw carries. The debugger shows it beside the raw words.
func decodeCmd(op uint8, words []uint32) string {
	switch {
	case op == 0x08 && len(words) >= 2:
		return fmt.Sprintf("CP[%02X] = %08X", words[0], words[1])
	case op == 0x10 && len(words) >= 1:
		cnt := (words[0]>>16)&0x0F + 1
		return fmt.Sprintf("XF[%04X] x%d", words[0]&0xFFFF, cnt)
	case op == 0x40 && len(words) >= 2:
		return fmt.Sprintf("list at %08X, %d bytes", words[0], words[1])
	case op == 0x61 && len(words) >= 1:
		return fmt.Sprintf("BP[%02X] = %06X", words[0]>>24, words[0]&0x00FFFFFF)
	case op >= 0x80 && len(words) >= 2:
		return fmt.Sprintf("vat %d, %d vertices", words[0], words[1])
	}
	return ""
}

func (a *Adapter) scratchMachine() (*gc.Machine, error) {
	if a.scratch == nil {
		disc, err := gc.Open(a.imagePath)
		if err != nil {
			return nil, err
		}
		sc, err := gc.NewMachine(disc)
		if err != nil {
			return nil, err
		}
		sc.SetSpinDetect(false)
		a.scratch = sc
	}
	return a.scratch, nil
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
