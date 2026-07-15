// Package dosadapter implements a debug.Target for the DOS/go32 host — the flat
// 32-bit protected-mode machine that runs a DJGPP go32/COFF image (Quake shareware
// in Phase A of the Xbox-oracle bring-up).
//
// It is the first CPU-centric adapter, and it deliberately backs only the CPU and
// memory half of the debugger. A DOS box has no display processor and no push
// buffer, so there is no frame stepper, no command scrubber, no per-pixel
// provenance and no surface list — the things those need do not exist on this
// machine, and a capability it cannot honestly back is one it does not claim. What
// it does have is a real CPU, a flat address space, an x86 disassembler, and a VGA
// mode-13h framebuffer; so it offers stepping, breakpoints, memory watches, a
// disassembly, a memory map, savestates, and its screen — plus keyboard input, the
// first Keyer, so you can drive the game's menus from the browser.
package dosadapter

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/cpu/x86"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/dos"
)

func init() {
	debug.Register("pc", func(s debug.OpenSpec) (debug.Target, error) {
		return New(s.Image, Options{GameDir: s.Get("gamedir")})
	})
}

// runBudget bounds a single StepInstr/Continue slice when the caller gives none. A
// Quake boot is tens of millions of instructions, so this is generous; the runner
// drives Continue in its own 2 M-instruction slices regardless.
const runBudget = 50_000_000

// Options carries what the platform cannot work out for itself.
type Options struct {
	// GameDir is where the program's data files resolve. Empty means the directory
	// the image sits in — which is where a DJGPP game's PAK files live.
	GameDir string
}

// The capabilities this target backs. Absent here — and so absent from the page —
// are frames, replay, surfaces and files: a DOS box has no display processor and no
// filesystem the debugger walks.
var (
	_ debug.Target       = (*Adapter)(nil)
	_ debug.CodeStepper  = (*Adapter)(nil)
	_ debug.Disassembler = (*Adapter)(nil)
	_ debug.Breakpointer = (*Adapter)(nil)
	_ debug.Watcher      = (*Adapter)(nil)
	_ debug.MemoryMapper = (*Adapter)(nil)
	_ debug.StateFiler   = (*Adapter)(nil)
	_ debug.Resumer      = (*Adapter)(nil)
	_ debug.Keyer        = (*Adapter)(nil)
)

// Adapter drives a go32 protected-mode machine as a debug.Target.
type Adapter struct {
	imagePath string
	gameDir   string
	pm        *dos.PM

	bps     map[uint32]bool // breakpoints, by linear PC
	honorBP bool            // whether the current run stops on a breakpoint (off while single-stepping off one)

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)

	stop debug.StopReason // set by a breakpoint or a breaking watch inside a run
}

// snap wraps a PM savestate as an opaque debug.Snapshot. dos.PMState is a deep copy,
// so it outlives the machine that made it and can be restored repeatedly.
type snap struct{ st *dos.PMState }

func (snap) Platform() string { return "pc" }

// New loads a go32 image into a fresh protected-mode machine and wires the debugger's
// per-step hook — which must run PumpInput every instruction so the DJGPP base-address
// invariant holds and injected keys are delivered, exactly as go32run does.
func New(imagePath string, opts Options) (*Adapter, error) {
	dir := opts.GameDir
	if dir == "" {
		dir = filepath.Dir(imagePath)
	}
	pm, err := dos.LoadGo32(imagePath, dir)
	if err != nil {
		return nil, err
	}
	a := &Adapter{imagePath: imagePath, gameDir: dir, pm: pm, bps: map[uint32]bool{}}
	a.pm.CPU.OnStep = func(c *x86.CPU) {
		// Mandatory every step: the base-address invariant and input delivery.
		a.pm.PumpInput(c)
		// A breakpoint stops the run by halting the CPU; runSlice clears that halt and
		// reports it as a stop, so the machine can continue afterward. The check is on
		// the instruction ABOUT to execute (IP), which OnStep runs before, so IP is left
		// resting on the breakpoint rather than one past it.
		if a.honorBP {
			pc := c.SegBase[x86.CS] + c.IP
			if a.bps[pc] {
				a.stop = debug.StopReason{Kind: "breakpoint", PC: uint64(pc)}
				c.Halt("breakpoint at %08X", pc)
			}
		}
	}
	return a, nil
}

func (a *Adapter) Platform() string { return "pc" }
func (a *Adapter) Title() string    { return "Quake" }

func (a *Adapter) Close() error { a.pm = nil; return nil }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{st: a.pm.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ps, ok := s.(snap)
	if !ok {
		return fmt.Errorf("dosadapter: snapshot is from %q, not pc", platformOf(s))
	}
	return a.pm.LoadState(ps.st)
}

func (a *Adapter) SaveStateFile(path string) error { return a.pm.SaveStateFile(path) }
func (a *Adapter) LoadStateFile(path string) error { return a.pm.LoadStateFile(path) }

// CPU is the register file, x86-named. PC is the linear address CS_base+EIP, the
// same address the disassembler and breakpoints use — not bare EIP, since the go32
// segments carry a non-zero descriptor base.
func (a *Adapter) CPU() debug.CPUReg {
	c := a.pm.CPU
	// Presented in the conventional x86 order (EAX EBX ECX EDX ESI EDI EBP ESP), which
	// is not the encoding order the Regs array is in.
	order := []struct {
		name string
		idx  int
	}{
		{"eax", x86.AX}, {"ebx", x86.BX}, {"ecx", x86.CX}, {"edx", x86.DX},
		{"esi", x86.SI}, {"edi", x86.DI}, {"ebp", x86.BP}, {"esp", x86.SP},
	}
	names := make([]string, len(order))
	vals := make([]uint64, len(order))
	for i, r := range order {
		names[i] = r.name
		vals[i] = uint64(c.Regs[r.idx])
	}
	return debug.CPUReg{
		PC:    uint64(c.SegBase[x86.CS] + c.IP),
		Names: names,
		Vals:  vals,
		Extra: map[string]uint64{
			"eflags": uint64(c.EFlags()),
			"cs":     uint64(c.Seg[x86.CS]), "ds": uint64(c.Seg[x86.DS]), "es": uint64(c.Seg[x86.ES]),
			"ss": uint64(c.Seg[x86.SS]), "fs": uint64(c.Seg[x86.FS]), "gs": uint64(c.Seg[x86.GS]),
		},
	}
}

// ReadMem reads flat linear memory. Addresses are linear (CS_base+EIP space), the
// same as everything else this adapter reports.
func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a.pm.Read(addr + uint32(i))
	}
	return out
}

// Display renders the VGA mode-13h framebuffer (256-colour, 320x200 at linear
// 0xA0000) through the DAC palette the game programmed, the same decode go32run's
// -png writer does.
func (a *Adapter) Display() (*image.RGBA, error) {
	const w, h, fb = 320, 200, 0xA0000
	var pal [256]color.RGBA
	for i := 0; i < 256; i++ {
		pal[i] = color.RGBA{a.pm.Pal[i*3] << 2, a.pm.Pal[i*3+1] << 2, a.pm.Pal[i*3+2] << 2, 255}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, pal[a.pm.Read(uint32(fb+y*w+x))])
		}
	}
	return img, nil
}

// ---- code ----

func (a *Adapter) StepInstr(n int) (debug.StopReason, error) {
	if n <= 0 {
		n = 1
	}
	// Single-stepping does not stop on a breakpoint — you are explicitly asking to move,
	// including off a breakpoint you are resting on. Watches still fire (they are bus
	// hooks, not gated here), which is what you want while stepping.
	return a.runSlice(uint64(n), "steps", false), nil
}

func (a *Adapter) Continue(budget uint64) (debug.StopReason, error) {
	if budget == 0 {
		budget = runBudget
	}
	// Resting on a breakpoint, step one instruction past it first — otherwise the run
	// stops on the breakpoint it already stands on and never moves.
	if a.atBreakpoint() {
		a.runSlice(1, "steps", false)
		if a.pm.CPU.Halted {
			return a.stopReason("halted"), nil
		}
	}
	return a.runSlice(budget, "budget", true), nil
}

func (a *Adapter) atBreakpoint() bool {
	c := a.pm.CPU
	return a.bps[c.SegBase[x86.CS]+c.IP]
}

// runSlice runs the CPU and works out why it stopped. A breakpoint or a breaking
// watch halts the CPU with a marker and records the reason in a.stop; runSlice
// clears that artificial halt so the machine can run again. A genuine halt (a fault,
// a program exit) leaves a.stop empty and is reported as halted.
func (a *Adapter) runSlice(budget uint64, exhausted string, honorBP bool) debug.StopReason {
	a.stop = debug.StopReason{}
	a.honorBP = honorBP
	steps := a.pm.CPU.Run(budget)
	a.honorBP = false

	sr := a.stop
	if sr.Kind != "" {
		// A breakpoint or watch: undo the marker halt so the next run continues.
		a.pm.CPU.Halted = false
		a.pm.CPU.HaltReason = ""
	} else if a.pm.CPU.Halted {
		sr = a.stopReason("halted")
	} else {
		sr = debug.StopReason{Kind: exhausted}
	}
	sr.PC = uint64(a.pm.CPU.SegBase[x86.CS] + a.pm.CPU.IP)
	sr.Steps = steps
	return sr
}

func (a *Adapter) stopReason(kind string) debug.StopReason {
	return debug.StopReason{Kind: kind, Note: a.pm.CPU.HaltReason}
}

// Disasm decodes n instructions at a linear address with the 32-bit decoder.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	va := uint32(addr)
	for i := 0; i < n; i++ {
		var b [16]byte
		for j := range b {
			b[j] = a.pm.Read(va + uint32(j))
		}
		in := x86.Decode32(b[:], va)
		sz := in.Len
		if sz <= 0 {
			sz = 1
		}
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: append([]byte(nil), b[:sz]...), Text: in.Text})
		va += uint32(sz)
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

func (a *Adapter) SetBreakpoint(pc uint64)   { a.bps[uint32(pc)] = true }
func (a *Adapter) ClearBreakpoint(pc uint64) { delete(a.bps, uint32(pc)) }
func (a *Adapter) Breakpoints() []uint64 {
	out := make([]uint64, 0, len(a.bps))
	for pc := range a.bps {
		out = append(out, uint64(pc))
	}
	return out
}

// ---- watches ----
//
// The machine has one write window and one read window (dos.PM holds one callback of
// each), so a target holds one watch of each kind; a second of the same kind is an
// error rather than a silent replacement.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("dosadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("dosadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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
	a.pm.SetWriteWatch(0, 0, nil)
	a.pm.SetReadWatch(0, 0, nil)
	for i := range a.watches {
		w := &a.watches[i]
		cb := func(addr, val, pc uint32) {
			w.Hits++
			if a.watchSink != nil {
				a.watchSink(debug.WatchHit{ID: w.ID, Kind: w.Kind, Addr: addr, Val: val, PC: pc, Instr: a.instrAt(pc)})
			}
			if w.Break {
				a.stop = debug.StopReason{Kind: "watch", PC: uint64(pc), Note: fmt.Sprintf("watch %d (%s %08x)", w.ID, w.Kind, addr)}
				a.pm.CPU.Halt("watch %d (%s %08X)", w.ID, w.Kind, addr)
			}
		}
		if w.Kind == "write" {
			a.pm.SetWriteWatch(w.Lo, w.Hi, cb)
		} else {
			a.pm.SetReadWatch(w.Lo, w.Hi, cb)
		}
	}
}

// ---- memory map ----

func (a *Adapter) Regions() []debug.Region {
	var out []debug.Region
	for _, r := range a.pm.MemRegions() {
		out = append(out, debug.Region{Name: r.Name, Lo: r.Lo, Hi: r.Hi})
	}
	return out
}

// ---- keyboard ----
//
// The first Keyer. A browser keydown/keyup arrives as a key name, which maps to a US
// Set-1 scancode, and a press is that make code while a release is make|0x80 — the
// same make/break the machine's scripted-input path uses. Every event is enqueued (no
// coalescing), and the guest delivers them one at a time as its keyboard ISR is ready.

func (a *Adapter) Key(k debug.Key) error {
	sc, ok := dos.Scancode(normalizeKey(k.Name))
	if !ok {
		return nil // an unmapped key (a browser modifier we do not model) is simply ignored
	}
	if !k.Down {
		sc |= 0x80 // break code
	}
	a.pm.EnqueueScancode(sc)
	return nil
}

// normalizeKey folds a browser KeyboardEvent.key value to the token names
// dos.Scancode knows ("ArrowUp" -> "up", "Escape" -> "esc", "A" -> "a", " " ->
// "space"). Anything already a known token passes through.
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
	case "escape":
		return "esc"
	case " ", "spacebar":
		return "space"
	case "control":
		return "ctrl"
	case "delete":
		return "del"
	case "insert":
		return "ins"
	case "pageup":
		return "pgup"
	case "pagedown":
		return "pgdn"
	}
	return s
}

// ---- resume ----

// ResumeArgs is the command line that reruns this image from a savestate. go32run is
// the DOS host's oracle, and it spells the flag -loadstate.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{"go", "run", "./cmd/go32run", "-image", a.imagePath, "-loadstate", statePath}
}

// ResumeDir runs that command from the tools module, where ./cmd/go32run lives —
// derived from the image path (games/<slug>/... sits beside tools/ in the repo).
func (a *Adapter) ResumeDir() string {
	if i := strings.LastIndex(a.imagePath, "/games/"); i >= 0 {
		return filepath.Join(a.imagePath[:i], "tools")
	}
	return ""
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}
