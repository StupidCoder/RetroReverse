// Package ps2adapter implements a debug.Target for the PlayStation 2 oracle on
// tools/platform/ps2 — Jak and Daxter and Ridge Racer V.
//
// The PS2 is an LLE machine like the PSX and the N64: the boot ELF is side-loaded off
// the disc (the address SYSTEM.CNF's BOOT2 names), and the whole render path — VIF,
// VU1, GIF, GS — runs for real. Some games' IOPRP images carry the whole IOP kernel and
// boot with no ROM (Jak); others carry only an update and need a console ROM supplied
// through the "bios" profile key (Ridge Racer V). See New. What this adapter backs first is the half a
// person most wants when they say "let me see it": the picture the GS is scanning out,
// played a field at a time, with the savestates the bootoracle already writes loadable
// straight into it so a session can jump to the frame worth looking at. The CPU and
// memory tools ride alongside.
//
// What it does NOT yet claim is the frame scrubber (FrameStepper/FrameReplayer). That
// needs the one question every adapter in this repo has had to answer differently and
// been wrong about at least once — "which buffer is the frame?" — settled with evidence
// on this title's PATH2-composited display buffers, plus a way to halt the GIF mid-burst
// (a DMA kick runs many GS primitives inside one EE instruction). It is a deliberate
// follow-up, not a stub: the capability is absent, so the page simply has no scrubber
// rather than an empty or lying one.
package ps2adapter

import (
	"crypto/md5"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/ps2"
)

// The debugger opens a PS2 game through the registry. The PS2 boot is self-contained —
// the disc names its own executable and its own IOP module image — so unlike the PSX
// there is no game-specific interrupt handler to supply; the profile is unused today.
func init() {
	debug.Register("ps2", func(s debug.OpenSpec) (debug.Target, error) {
		// The game may name a console ROM in its debug.json profile, resolved next to
		// the image. It is needed when the game's IOPRP image is an update that carries
		// only some of the IOP kernel (Ridge Racer V); a self-contained image (Jak)
		// leaves the key empty and boots with no ROM, exactly as before.
		var bios string
		if b := s.Get("bios"); b != "" {
			bios = filepath.Join(filepath.Dir(s.Image), b)
		}
		return New(s.Image, bios)
	})
}

// runBudget bounds a single field's run. A field is stepsPerVBlank = 1,000,000 EE
// instructions by construction; a boot to a drawn frame is billions, but a savestate
// resume lands mid-title, so this is a per-field ceiling with headroom for a machine
// that has wedged rather than a whole-boot budget.
const runBudget = 20_000_000

// eeRegNames are the MIPS o32 ABI register names, in index order — the EE is a 128-bit
// MIPS III core and its lower 64 bits carry the ABI the game compiles to.
var eeRegNames = []string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

// Adapter drives a PS2 oracle as a debug.Target.
type Adapter struct {
	imagePath string
	biosPath  string
	bios      []byte
	vol       *iso9660.Volume
	exe       *ps2.Executable
	sum       string

	live *ps2.Machine

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64
	stop      debug.StopReason

	// The pad. held is which keys the browser currently has down; padQueue is the pad
	// states waiting for a field to be sampled in. See "the pad" below.
	held     map[string]bool
	padQueue []padState
}

// padState is everything one poll latches: the buttons and the left stick, together.
// Queued whole, for the reason the machine stores them whole — a poll samples both at the
// same instant, so a state pairing a new stick position with last field's buttons is a pad
// that never existed.
type padState struct {
	buttons uint16
	stickX  uint8
	stickY  uint8
}

// padStickDiag is one axis of a stick held to a corner: PadStickFull / sqrt(2), rounded.
// The DualShock's gate is a circle, so a stick pushed diagonally is still the same
// distance from centre — it simply splits that distance over two axes.
const padStickDiag = 0x5A // 0x7F * 0.7071

// The capabilities this target backs. What is absent is absent from the page — there is
// no FrameStepper here yet, so no frame scrubber, rather than an empty one.
var (
	_ debug.Target       = (*Adapter)(nil)
	_ debug.FrameStepper = (*Adapter)(nil)
	_ debug.FastStepper  = (*Adapter)(nil)
	_ debug.CodeStepper  = (*Adapter)(nil)
	_ debug.Breakpointer = (*Adapter)(nil)
	_ debug.Disassembler = (*Adapter)(nil)
	_ debug.Watcher      = (*Adapter)(nil)
	_ debug.Surfacer     = (*Adapter)(nil)
	_ debug.StateFiler   = (*Adapter)(nil)
	_ debug.Resumer      = (*Adapter)(nil)
	_ debug.MemoryMapper = (*Adapter)(nil)
	_ debug.Haltable     = (*Adapter)(nil)
	_ debug.Keyer        = (*Adapter)(nil)
	_ debug.KeyLegender  = (*Adapter)(nil)
	_ debug.Profiler     = (*Adapter)(nil)
)

// snap wraps a PS2 in-memory savestate as an opaque debug.Snapshot.
type snap struct{ ms ps2.MachineState }

func (snap) Platform() string { return "ps2" }

// New opens a disc image and side-loads its boot executable, the way the bootoracle
// does: SYSTEM.CNF names the executable, the ELF is read off the disc, and the IOP is
// booted on the module image the game itself reboots it onto.
// New opens the disc at imagePath. biosPath is optional: the empty string boots with
// no console ROM (a self-contained IOPRP image, like Jak's), and a path supplies rom0
// for a game whose image is only an update (Ridge Racer V).
func New(imagePath, biosPath string) (*Adapter, error) {
	raw, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	sum := fmt.Sprintf("%x", md5.Sum(raw))
	vol, err := iso9660.OpenBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("opening the disc: %w", err)
	}
	exePath, err := bootExe(vol)
	if err != nil {
		return nil, err
	}
	elfRaw, err := vol.ReadFile(exePath)
	if err != nil {
		return nil, fmt.Errorf("reading the boot executable %s: %w", exePath, err)
	}
	exe, err := ps2.LoadELF(elfRaw)
	if err != nil {
		return nil, err
	}
	a := &Adapter{imagePath: imagePath, biosPath: biosPath, vol: vol, exe: exe, sum: sum, held: map[string]bool{}}
	if biosPath != "" {
		if a.bios, err = os.ReadFile(biosPath); err != nil {
			return nil, fmt.Errorf("reading the BIOS %s: %w", biosPath, err)
		}
	}
	if err := a.boot(); err != nil {
		return nil, err
	}
	return a, nil
}

// bootExe reads SYSTEM.CNF's BOOT2 line — the disc's own name for the executable to run.
func bootExe(vol *iso9660.Volume) (string, error) {
	cnf, err := vol.ReadFile("SYSTEM.CNF")
	if err != nil {
		return "", fmt.Errorf("reading SYSTEM.CNF: %w", err)
	}
	for _, line := range strings.Split(string(cnf), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "BOOT2"); ok {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "=")), nil
		}
	}
	return "", fmt.Errorf("SYSTEM.CNF names no BOOT2 executable")
}

// boot builds the machine and brings the IOP up before the EE's first instruction,
// exactly as the bootoracle does. The IOP the game later reboots itself is a second
// boot; this is the power-on one.
func (a *Adapter) boot() error {
	m := ps2.NewMachine()
	m.SetImageHash(a.sum)
	m.SetVolume(a.vol)
	if a.bios != nil {
		m.SetBIOS(a.bios)
	}
	m.LoadExecutable(a.exe)
	if err := m.RebootIOP(); err != nil {
		return fmt.Errorf("bringing the IOP up: %w", err)
	}
	m.SetProfile(true) // the profile panel reads the last stepped field; arm it from the start
	a.live = m
	return nil
}

func (a *Adapter) Platform() string { return "ps2" }
func (a *Adapter) Title() string    { return filepath.Base(a.imagePath) }

// Machine exposes the live machine, for wiring the generic interface does not cover
// (scripted pad input).
func (a *Adapter) Machine() *ps2.Machine { return a.live }

func (a *Adapter) Close() error { a.live = nil; return nil }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{ms: a.live.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ps, ok := s.(snap)
	if !ok {
		return fmt.Errorf("ps2adapter: snapshot is from %q, not ps2", platformOf(s))
	}
	if err := a.live.LoadState(ps.ms); err != nil {
		return err
	}
	a.live.SetProfile(true) // re-arm: the first field after a load is a whole field
	return nil
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }
func (a *Adapter) LoadStateFile(path string) error {
	if err := a.live.LoadStateFile(path); err != nil {
		return err
	}
	a.live.SetProfile(true) // re-arm: the first field after a load is a whole field
	return nil
}

// FrameProfile reports where the last stepped field's time went (ps2/profile.go): the DMA
// transport, the VU1 geometry and the rasteriser timed at their per-primitive/per-kick
// boundaries, the GIF/VIF/DMA decode derived from the transport, and the EE + IOP + devices
// as the remainder they are. No fragment or IOP-instruction bucket — both happen far too many
// times a field to time without the clock becoming the measurement (they are counters).
func (a *Adapter) FrameProfile() debug.FrameProfile {
	p := a.live.FrameProfile()
	out := debug.FrameProfile{TotalMs: p.TotalMs, Drew: p.Drew}
	for _, b := range p.Buckets {
		out.Buckets = append(out.Buckets, debug.ProfileBucket{Name: b.Name, Millis: b.Millis, Count: b.Count})
	}
	for _, c := range p.Counters {
		out.Counters = append(out.Counters, debug.ProfileCounter{Name: c.Name, Value: c.Value})
	}
	return out
}

// StepFast advances one field capturing nothing — how the debugger plays. The vblank is
// the machine's frame boundary; OnVBlank asks the run to stop the instant it fires.
//
// The pad's next state is released BEFORE the field runs, not at the vblank that ends it,
// so it is live across the whole field and is sampled by the one poll inside it whatever
// order the vblank's handlers happen to run in.
func (a *Adapter) StepFast() error {
	a.releasePad()
	a.live.OnVBlank = func(m *ps2.Machine) { m.StopRequested = true }
	a.live.Run(runBudget)
	a.live.OnVBlank = nil
	return nil
}

// StepFrame advances one field, capturing the GS primitive stream and per-pixel
// provenance along the way, so the page can list what the frame drew and answer "which
// primitive drew this pixel?" without a round trip.
//
// THE FRAME IS THE SCANOUT BUFFER. At Jak's title the composition is built with PATH2
// DIRECT straight into the buffer DISPFB scans out, so the picture GSFrame returns and
// the pixels the frame draws are the same plane — provenance laid over the scanout lines
// up. A pixel written to another buffer (the sky column buffer, a back buffer the frame
// is not showing) has nowhere to land on this picture and is dropped, which is honest:
// it is not part of the frame on screen. (If a future scene turns out to double-buffer,
// this is the "which buffer is the frame?" decision to revisit — the write census per
// buffer is right here to make it on evidence.)
//
// There is no per-command scrubber yet (no FrameReplayer): a GS kick draws many
// primitives inside one EE instruction, so replaying to "after primitive k" needs a
// mid-burst halt this pass does not build. Provenance still gives the two questions that
// matter most — click a pixel to get its primitive, select a primitive to see its
// pixels — because both read the capture, not a replay.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	a.releasePad()
	fc := &debug.FrameCapture{Start: snap{ms: a.live.SaveState()}}
	fc.CountWrites()

	// Where the scanout buffer sits, so a pixel write to it can be mapped onto the
	// visible rectangle. Captured at the frame's start; the title does not move DISPFB
	// mid-field.
	scFB, dbx, dby, sw, sh, haveSC := a.live.GSScanout()

	// Provenance is gathered keyed by packed (y<<16|x) in the scanout buffer while the
	// frame draws, then laid out row-major once the size is known.
	prov := map[uint32]int32{}

	a.live.OnGSPrim = func(ptype int, producer string) int {
		idx := len(fc.Commands)
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index:   idx,
			Name:    primName(ptype),
			Op:      uint32(ptype),
			Decoded: producer,
		})
		return idx
	}
	a.live.OnGSPixel = func(cmd int, fbWord uint32, x, y int32) {
		if cmd < 0 || !haveSC || fbWord != scFB {
			return // a buffer that is not the one on screen
		}
		px, py := x-int32(dbx), y-int32(dby)
		if px < 0 || py < 0 || int(px) >= sw || int(py) >= sh {
			return
		}
		fc.MarkWrite(cmd)
		prov[uint32(py)<<16|uint32(px)&0xFFFF] = int32(cmd)
	}
	a.live.OnVBlank = func(m *ps2.Machine) { m.StopRequested = true }

	a.live.Run(runBudget)

	a.live.OnGSPrim = nil
	a.live.OnGSPixel = nil
	a.live.OnVBlank = nil

	// The picture is the scanout, and provenance is laid over it — same plane, same
	// coordinates. Only the size is needed here; the page fetches the image itself.
	_, w, h := a.live.GSFrame()
	fc.Width, fc.Height = w, h
	if w > 0 && h > 0 && len(prov) > 0 {
		fc.Prov = make([]int32, w*h)
		for i := range fc.Prov {
			fc.Prov[i] = -1
		}
		for key, c := range prov {
			x, y := int(key&0xFFFF), int(key>>16)
			if x < w && y < h {
				fc.Prov[y*w+x] = c
			}
		}
	}
	return fc, nil
}

// primName names a GS primitive type for the command list.
func primName(ptype int) string {
	switch ptype & 7 {
	case 0:
		return "point"
	case 1:
		return "line"
	case 2:
		return "linestrip"
	case 3:
		return "triangle"
	case 4:
		return "tristrip"
	case 5:
		return "trifan"
	case 6:
		return "sprite"
	}
	return "prim7"
}

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 32)
	for i := 0; i < 32; i++ {
		vals[i] = c.Reg(uint32(i))
	}
	return debug.CPUReg{
		PC:    c.PC & 0xFFFFFFFF,
		Names: eeRegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"hi": c.HI, "lo": c.LO,
			"Status": c.COP0[12], "Cause": c.COP0[13], "EPC": c.COP0[14],
		},
	}
}

// Display is the frame the GS is scanning out — the DISPFB rectangle, deswizzled through
// the real read path. It is the same picture the bootoracle's -gsframe writes.
func (a *Adapter) Display() (*image.RGBA, error) {
	pix, w, h := a.live.GSFrame()
	if pix == nil || w == 0 || h == 0 {
		return nil, fmt.Errorf("ps2adapter: no displayable frame yet (no DISPFB, or an unread format)")
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	return img, nil
}

func (a *Adapter) ReadMem(addr uint32, n int) []byte { return a.live.ReadMem(addr, n) }

// Surfaces. The scanout is the displayed frame; the free "GS buffer" reads any PSMCT32
// rectangle of GS memory by its word base — the display buffers (0xA0000/0xC0000), the
// sky column buffer (0x7F800), or anything the draw-target census names — so a buffer
// that is being composited but not yet scanned out can still be looked at.
func (a *Adapter) Surfaces() []debug.Surface {
	pix, w, h := a.live.GSFrame()
	sc := debug.Surface{ID: "scanout", Name: "Display (scanout)"}
	if pix != nil {
		sc.W, sc.H = w, h
	}
	return []debug.Surface{
		sc,
		{ID: "buffer", Name: "GS buffer (PSMCT32)", Free: true, Formats: []string{"PSMCT32"}},
	}
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "scanout":
		return a.Display()
	case "buffer":
		bw := uint32(v.W) / 64
		if bw == 0 {
			bw = 1
		}
		h := v.H
		if h <= 0 {
			h = 256
		}
		pix, w := a.live.GSBuffer(v.Addr, bw, h)
		if pix == nil {
			return nil, fmt.Errorf("ps2adapter: nothing at GS word 0x%X", v.Addr)
		}
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		copy(img.Pix, pix)
		return img, nil
	}
	return nil, fmt.Errorf("ps2adapter: no surface %q: %w", id, debug.ErrUnsupported)
}

func (a *Adapter) Regions() []debug.Region {
	return []debug.Region{
		{Name: "RAM", Lo: 0x0000_0000, Hi: 0x0200_0000},
		{Name: "Scratchpad", Lo: 0x7000_0000, Hi: 0x7000_4000},
		{Name: "IO", Lo: 0x1000_0000, Hi: 0x1001_0000},
		{Name: "VU/GS regs", Lo: 0x1100_0000, Hi: 0x1200_1000},
	}
}

// ---- code ----

// Halted answers for the EE core and the machine's own halt (an unmodelled kernel call,
// a jump into unmapped memory). The video clock keeps ticking after either, so at the
// frame level this is the only thing that tells a stopped machine from a busy one.
func (a *Adapter) Halted() (bool, string) {
	if a.live.CPU.Halted {
		return true, a.live.CPU.HaltReason
	}
	if a.live.Halted {
		return true, a.live.HaltReason
	}
	return false, ""
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
	// Sitting on a breakpoint, run one instruction first, or the run stops on the
	// breakpoint it is already standing on and never moves.
	if a.atBreakpoint() {
		if _, err := a.runSlice(1, "steps"); err != nil {
			return debug.StopReason{}, err
		}
	}
	return a.runSlice(budget, "budget")
}

func (a *Adapter) atBreakpoint() bool {
	pc := a.live.CPU.PC & 0xFFFFFFFF
	for _, b := range a.bps {
		if b == pc {
			return true
		}
	}
	return false
}

// runSlice runs the machine and works out why it stopped. Run's own reason names a
// breakpoint; a breaking watch reports itself through a.stop.
func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	res := a.live.Run(budget)
	sr := a.stop
	if sr.Kind == "" {
		switch {
		case strings.HasPrefix(res.Reason, "breakpoint"):
			sr = debug.StopReason{Kind: "breakpoint", Note: res.Reason}
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

// Disasm decodes n instructions at a virtual address, through the machine's own EE
// disassembler (which also names GOAL symbols when they have been loaded).
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	for i := 0; i < n; i++ {
		va := uint32(addr) + uint32(i)*4
		w := a.live.Fetch32(va)
		var b [4]byte
		b[0], b[1], b[2], b[3] = byte(w), byte(w>>8), byte(w>>16), byte(w>>24)
		text := a.live.DisasmAt(va)
		if s := a.live.Sym(va); s != "" {
			text = s + "  " + text
		}
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: b[:], Text: text})
	}
	return out, nil
}

func (a *Adapter) instrAt(pc uint32) string { return a.live.DisasmAt(pc) }

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
// The machine has one write window and one read window, so a target holds one watch of
// each kind; a second of the same kind is an error rather than a silent replacement.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("ps2adapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("ps2adapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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

// ResumeArgs is the command line that reopens this disc at a saved state. The PS2 oracle
// spells its flags -image and -loadstate.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{
		"go", "run", "./cmd/bootoracle",
		"-image", a.imagePath, "-loadstate", statePath,
	}
}

func platformOf(s debug.Snapshot) string {
	if s == nil {
		return "nil"
	}
	return s.Platform()
}

// ---- the pad ----
//
// The DualShock is the Keyer's third platform, after the DOS keyboard and the GameCube's
// controller, and it is the GameCube's shape rather than the keyboard's.
//
// A PC keyboard DELIVERS EVENTS: make and break are each a scancode the guest's ISR
// consumes, so the adapter enqueues both and coalesces nothing. THE DUALSHOCK IS A LEVEL:
// padman polls it, latches whatever the pad reports at that instant, and the game
// edge-detects presses from successive polls. Nothing queues on the guest's side, so
// nothing self-paces, and a browser key that goes down and up between two polls is a press
// the game never sees. The debugger's fields are wall-clock slow, so that is not
// hypothetical.
//
// So the states queue and release ONE PER FIELD. The Xbox adapter's warning is that the
// fastest boundary is the trap — releasing one state per USB report drains the queue inside
// a single game frame — so the boundary here was MEASURED rather than assumed, twice:
//
//   - padman polls the pad EXACTLY ONCE PER VBLANK (59 polls across 60 vblanks, counted at
//     the pad's own poll command). The vblank is not a boundary that runs faster than the
//     poll and drains the queue; it IS the poll.
//   - ONE VBLANK OF PRESS IS ENOUGH, end to end. -pad X@830:1 — a press live for exactly
//     one vblank, and therefore sampled by exactly one poll — dismisses Jak's memory-card
//     prompt, and a run with no press at all leaves the prompt up. That control matters:
//     without it, a prompt that dismissed itself would look like a press that worked.
//
// Both were worth measuring because the recorded lore said otherwise ("presses need ~150
// vblank holds"). They do not. A state released per field is sampled by one poll, and one
// poll is acted on.
func (a *Adapter) Key(k debug.Key) error {
	name := normalizeKey(k.Name)
	_, isButton := ps2.PadButton(name)
	if _, isStick := stickDirs[name]; !isButton && !isStick {
		return nil // an unmapped key is ignored rather than an error: the browser sends everything
	}
	if k.Down {
		a.held[name] = true
	} else {
		delete(a.held, name)
	}

	st := padState{}
	for n := range a.held {
		if b, ok := ps2.PadButton(n); ok {
			st.buttons |= b
		}
	}
	st.stickX, st.stickY = stickFrom(a.held)

	// Only a CHANGE needs a field of its own; a repeat is the same level and would just
	// cost a field.
	if len(a.padQueue) > 0 && a.padQueue[len(a.padQueue)-1] == st {
		return nil
	}
	a.padQueue = append(a.padQueue, st)
	return nil
}

// KeyLegend tells the page what the keys do. W/A/S/Z is a diamond standing for a diamond
// and reads as arbitrary until that is said once, so it is worth the line.
func (a *Adapter) KeyLegend() string {
	return "arrows = left stick, Enter = start, W/A/S/Z = triangle/square/circle/cross"
}

// stickDirs are the four keys that push the LEFT stick, kept apart from the button table
// because they are not buttons: they resolve to a POSITION, not to bits in a mask.
//
// The offsets are in SCREEN terms — down is +y — and that is the pad's convention too
// here: the DualShock reports 0x00 at the top of its travel and 0xFF at the bottom, so up
// DECREASES. (This is where the GameCube's keyer needs a negation and this one does not,
// which is exactly the kind of difference that ships a stick upside down if it is assumed
// rather than read off the wire.)
var stickDirs = map[string]struct{ dx, dy int }{
	"stickup":    {0, -1},
	"stickdown":  {0, +1},
	"stickleft":  {-1, 0},
	"stickright": {+1, 0},
}

// stickFrom resolves the held direction keys to the left stick's two wire bytes.
//
// THE GATE IS A CIRCLE, which is why this is not two independent axes. The stick's shell
// stops it the same distance from centre in every direction, so a stick held to a corner
// reads about 0.7 of full on each axis and CANNOT read full on both. Driving each axis to
// full independently would hand the game a diagonal 1.41x longer than any real pad can
// produce — which it would either clamp back, making the work pointless, or believe. A
// keyboard has no gate, so the gate is modelled here.
func stickFrom(held map[string]bool) (x, y uint8) {
	var dx, dy int
	for n := range held {
		if d, ok := stickDirs[n]; ok {
			dx += d.dx
			dy += d.dy
		}
	}
	// Opposite keys held together cancel, which is what a physical stick does too.
	if dx == 0 && dy == 0 {
		return ps2.PadStickCentre, ps2.PadStickCentre
	}
	mag := ps2.PadStickFull
	if dx != 0 && dy != 0 {
		mag = padStickDiag // a point on the circle; same distance from centre, split over two axes
	}
	clamp := func(d int) uint8 {
		v := ps2.PadStickCentre + d*mag
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

// releasePad hands the machine the next queued pad state, one per field. An empty queue
// leaves the current state alone, which is what keeps a held button held.
//
// It is called from the field-boundary hook rather than installed as one: OnVBlank is how
// both StepFast and StepFrame end their run, so the pad borrows that boundary instead of
// claiming the machine's one vblank hook for itself.
func (a *Adapter) releasePad() {
	if len(a.padQueue) == 0 {
		return
	}
	s := a.padQueue[0]
	a.live.SetPadButtons(s.buttons)
	a.live.SetPadStick(s.stickX, s.stickY, ps2.PadStickCentre, ps2.PadStickCentre)
	a.padQueue = a.padQueue[1:]
}

// normalizeKey folds a browser KeyboardEvent.key value to the names the pad knows.
//
// THE FACE BUTTONS ARE W/A/S/Z BECAUSE THAT DIAMOND IS THE PAD'S DIAMOND. Triangle, square,
// circle and cross sit at the top, left, right and bottom of a diamond on the controller,
// and W, A, S and Z sit at the top, left, right and bottom of one on the keyboard — so the
// mapping is the shape, not a legend to memorise. Reaching for the button that is "up" on
// the pad means reaching for the key that is "up" under your hand.
//
// THE ARROWS DRIVE THE LEFT STICK, NOT THE D-PAD, because that is what it takes to play the
// game: Jak walks on the stick. The d-pad keeps the numpad's own arrows (8/4/6/2, the
// layout it already looks like), so the capability is moved rather than dropped — those
// arrive as the digits whether typed on the numpad or the number row. Enter is Start, which
// is the one mapping every debugger in this repo already shares.
func normalizeKey(name string) string {
	s := strings.ToLower(name)
	switch s {
	case "arrowup":
		return "stickup"
	case "arrowdown":
		return "stickdown"
	case "arrowleft":
		return "stickleft"
	case "arrowright":
		return "stickright"
	case "w":
		return "triangle"
	case "a":
		return "square"
	case "s":
		return "circle"
	case "z":
		return "cross"
	case "8":
		return "up"
	case "2":
		return "down"
	case "4":
		return "left"
	case "6":
		return "right"
	case "enter", "return":
		return "start"
	}
	return s
}
