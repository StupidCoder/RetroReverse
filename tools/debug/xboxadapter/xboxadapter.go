// Package xboxadapter implements a debug.Target for the original-Xbox oracle.
//
// It owns a live xbox.Machine and, for command replay, a scratch machine it restores into.
// Almost everything the debugger needs is a hook the machine already offers (OnNVMethod,
// OnPixel, OnFlip, the watch windows, StopRequested) or an in-memory savestate; this
// package is the thin translation between those and the platform-agnostic debug surface.
//
// Three things about the Xbox shape the adapter.
//
// THE FRAME IS FLIP_STALL. The suite's recurring question — which buffer is the frame, and
// where does one end? — has a new answer here, and every plausible alternative is wrong. The
// NV2A renders into ordinary RAM (no on-die EFB to be copied out and wiped, as on the
// GameCube; no tiled render target in a private aperture, as on the 3DS), so the draw target
// is readable at any time and nothing erases it. What is not obvious is where a frame ENDS.
// The boundary is the Kelvin method NV097_FLIP_STALL — what Direct3D's Present compiles to,
// and the only thing in the stream that says "this frame is finished and meant for the
// screen". nv2a_kelvin.go records what that ruled out and how each was measured; the one
// worth repeating here is that the swap-chain buffer moving on (SET_SURFACE_COLOR_OFFSET)
// fires exactly once per frame at the logo card, in lockstep with the real boundary — and
// three times per frame at the title screen, which renders its movie into an off-screen
// target first. A capture bounded by it there ends mid-frame on a buffer nothing has drawn
// into and reports a blank white frame. It looks entirely plausible. That is the whole
// lesson: the scene that validates a boundary is not the scene that breaks it.
//
// The picture is captured INSIDE the flip hook, before the method latches, while the colour
// surface still names the buffer the frame was built in.
//
// AND THE SCANOUT IS NOT THE PICTURE. Display() — the debugger's main view, what play mode
// streams — is the buffer the title PRESENTED, not the framebuffer the CRTC is programmed to
// read. Those differ because this machine does not model the CRTC: the title registers a
// scanout through AvSetDisplayMode exactly once, during loading (320x240 at 0174C000), never
// makes the 640x480 switch it goes on to render at (a known-pending frontier), and the
// PCRTC_START writes its vblank ISR does make carry 0xFFFFFB00 — `0 - pitch`, a constant,
// every time. So the scanout registers cannot say what is on screen: rendering them gives a
// blank white rectangle while the game draws perfectly, which is indistinguishable from a
// broken emulator and is what a user staring at play mode actually sees.
//
// The flip can say. The title marks its presents, and the colour surface at that instant
// names the buffer it means — so the machine tracks it (RenderPresented), which is what any
// emulator does when it knows the present but not the register behind it. The programmed
// scanout is still offered as its own surface, because the gap it shows is real and should
// stay visible; the honesty belongs in keeping that surface, not in crippling the main view.
//
// A COMMAND IS ONE PUSH-BUFFER METHOD. The NV2A is a register-write machine like the PICA
// and the GE, not a packet machine like the GameCube's FIFO: a frame is tens of thousands
// of (subchannel, method, argument) writes of which a few hundred draw, and the rest set up
// the state they draw against. So the scrubber leans on FrameCapture.Writers — the commands
// that actually stored pixels — because scrubbing the raw stream is mostly scrubbing through
// state-setting while the picture stands still.
//
// STOPPING MID-LIST IS NOT DECLINING THE NEXT COMMAND. The GameCube's FIFO is drained by a
// Go loop the caller owns, so its adapter simply stops asking for more. The NV2A's pusher
// runs INSIDE the guest's own store to the DMA_PUT register, several frames deep in the
// title's D3D runtime — there is no loop to decline. So the machine arms a countdown that
// the method dispatch trips (xbox.RunStopAfterNVMethod), which stops the pusher mid-buffer
// and the CPU with it, leaving the machine with GET short of PUT. That is a machine mid-
// frame, and it is why replay only ever runs on the scratch machine.
//
// There WAS no Keyer here, and the reason is worth keeping: the Xbox's USB/XID pad was not
// modelled, so the machine had no honest way to accept a button — and a capability is a
// promise the target can back. An input panel that silently swallowed every press would
// have been worse than no panel. The input phase arrived, and the promise is now backed by
// a real OHCI controller with a pad on its root hub, so the panel appears — because the
// page builds itself from what the target says it can do, and nothing in the web UI or the
// wire protocol needed a line of Xbox-specific code to make that happen.
//
// What the pad cost was one small honesty in this file. The adapter used to OWN OnFlip,
// installing and nilling a fresh closure on every StepFrame and StepFast; a hook that the
// pad's pacing chained onto would have been wiped by the first step. The machine has ONE
// frame-boundary hook and the pad has no more claim to own it than a capture does, so New
// installs it once and each step fills in a.flipHook instead. See the pad section below
// for why the frame — not the vblank, and emphatically not the 1 kHz USB frame — is the
// boundary a button has to be paced against.
package xboxadapter

import (
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/x86"
	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/xbox"
)

// The debugger opens an Xbox game through the registry, so nothing has to import this
// package by name — importing it for its side effect is enough.
func init() {
	debug.Register("xbox", func(s debug.OpenSpec) (debug.Target, error) {
		xbePath := s.Get("xbe")
		if xbePath == "" {
			xbePath = "/default.xbe"
		}
		return New(s.Image, xbePath)
	})
}

// runBudget bounds a single StepFrame/replay. OutRun's title screen spends a few million
// instructions per presented frame (the XMV movie decoder runs on the interpreted CPU),
// so this leaves generous headroom while still catching a machine that has wedged. A cold
// boot to the title is ~2.7 billion — far past this — which is why the debugger is meant
// to be pointed at a savestate rather than made to sit through the boot.
const runBudget = 400_000_000

// x86RegNames are the 32-bit general registers in the machine's own index order (which is
// the x86 encoding's: AX CX DX BX SP BP SI DI, not the alphabetical order a listing
// suggests). ESP and EBP are marked because they are what a stack walk reads.
var x86RegNames = []string{"EAX", "ECX", "EDX", "EBX", "ESP/sp", "EBP/fp", "ESI", "EDI"}

// Adapter drives an Xbox oracle as a debug.Target.
type Adapter struct {
	imagePath string
	xbePath   string
	live      *xbox.Machine
	scratch   *xbox.Machine // reused across RenderAfter; its state is always disposable

	watches   []debug.Watch
	nextWatch int
	watchSink func(debug.WatchHit)
	bps       []uint64

	// stop is filled in by a hook that halted the run (a breaking watch), so the reason
	// survives the return from Run, which only reports "stop requested".
	stop debug.StopReason

	// flipHook is the current step's frame-boundary work — StepFrame's picture, or
	// StepFast's stop. It is a field rather than the machine's OnFlip because the pad
	// needs that boundary too, and a step that installed its own hook would wipe the
	// pacing (see padPace). New installs ONE OnFlip that calls both.
	flipHook func(*xbox.Machine)

	// The pad. held is which keys the browser currently has down; padQueue is the pad
	// states the machine has not sampled yet. See Key.
	held     map[string]bool
	padQueue []xbox.PadState
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
	_ debug.Haltable       = (*Adapter)(nil)
	_ debug.Keyer          = (*Adapter)(nil)
)

// ---- the pad ----
//
// The Keyer's third platform, and its second console button. The DOS keyboard DELIVERS
// EVENTS — a make and a break the guest's ISR consumes, so nothing may coalesce and the
// guest paces itself. The GameCube's pad IS A LEVEL the serial interface latches once per
// field. The Xbox's pad is a level too, and the title edge-detects presses from it with
// its own ~prev&cur (0xA51C8) — so a browser key that goes down and up between two polls
// is a press the game never sees, and the debugger's frames are wall-clock slow enough
// that this is not hypothetical.
//
// So the states queue and release one per FRAME, and the frame is the flip. That choice
// is the one non-obvious part, because there are three candidate boundaries here and the
// fastest is the trap:
//
//   - The USB frame (1 kHz) is where reports actually fly, and it is WRONG: releasing one
//     state per report drains the queue in milliseconds, so a press and its release both
//     land inside a single game frame and the edge-detect never sees the press. It is the
//     GameCube's hazard arriving through a faster pipe.
//   - The vblank is a field, and ticks whether or not the title drew.
//   - The flip is what the title's own poll is synchronous with. A state held across one
//     flip rides ~16 USB reports and is sampled by at least one poll.
//
// Which is the GameCube's "one state per field" transposed exactly: there the pad is
// latched per field, here it is polled per frame.

// Key records a browser key event as a pad state. An unmapped key is ignored rather than
// an error: the browser sends everything, including keys that are not buttons.
func (a *Adapter) Key(k debug.Key) error {
	name := normalizeKey(k.Name)
	if _, ok := xbox.PadControlByName(name); !ok {
		return nil
	}
	// The first button is also what plugs the pad in. Attaching at New would connect a
	// controller to every debugger session whether or not anyone meant to drive one —
	// and a machine merely being LOOKED at should not have hardware appear in it.
	//
	// Guarded because the console is the only part of this that needs one: the vocabulary
	// and the queue are the adapter's own, so pad_test.go can drive Key on a checkout with
	// no disc image — which is the difference between the stick mapping being tested on
	// every run and being tested on the one machine that has the game.
	if a.live != nil {
		a.live.AttachPad(0)
	}
	if k.Down {
		a.held[name] = true
	} else {
		delete(a.held, name)
	}
	// The vocabulary and the gate both live platform-side, next to the device: this
	// adapter resolves a keyboard's names and knows nothing about which byte they are.
	st := xbox.PadStateOf(a.held)
	// Only a CHANGE needs a frame of its own; a repeat (the browser's key-repeat) is the
	// same level and would just cost a frame.
	if len(a.padQueue) > 0 && a.padQueue[len(a.padQueue)-1] == st {
		return nil
	}
	a.padQueue = append(a.padQueue, st)
	return nil
}

// padPace releases one queued pad state per frame, from the machine's flip hook. An
// empty queue leaves the current level alone, which is what keeps a held button held —
// the queue carries changes, and never invents a release.
func (a *Adapter) padPace(m *xbox.Machine) {
	if len(a.padQueue) == 0 {
		return
	}
	m.SetPad(0, a.padQueue[0])
	a.padQueue = a.padQueue[1:]
}

// normalizeKey folds a browser KeyboardEvent.key value to the control names
// xbox.PadControlByName knows — the same names the oracle's -keys scripts use.
//
// THE ARROWS ARE THE STICK, not the d-pad, and that is the one mapping here worth arguing
// for. The Xbox's primary directional control is the left stick: it is what the title's own
// menus steer by, and the pad's derivation watched a stick direction and its matching d-pad
// bit produce the SAME FRAME to the MD5 — so on a menu the two are interchangeable and the
// arrows may as well be the one that also works when the game wants an analog direction.
// The GameCube's adapter made exactly this move for exactly this reason.
//
// So the d-pad needs a home of its own, and it gets the GameCube's: the numeric keypad
// laid out as a d-pad, 8/2/4/6 around 5. It is not a fallback — a menu that reads only the
// d-pad and a game that reads only the stick are both reachable from one keyboard, and
// telling them apart is exactly the sort of thing this debugger exists to do.
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
	// a and b are the analog buttons LICENSE SELECT's own footer named, and they arrive
	// under their own letters. Every other letter falls through to a lookup that will not
	// find it, which is the honest outcome: the pad has ten controls with no name yet
	// (usb_xid.go says which and why), and a key that quietly did nothing would be worse
	// than one the vocabulary refuses.
	return s
}

// snap wraps an Xbox in-memory savestate as an opaque debug.Snapshot.
type snap struct{ st *xbox.XboxState }

func (snap) Platform() string { return "xbox" }

// New opens the disc at imagePath, parses the XBE at xbePath within it, and builds a
// machine parked at the title's entry point. It does NOT boot on to a frame: the title
// screen is ~2.7 billion instructions away, so the debugger is meant to be pointed at a
// savestate (LoadStateFile) rather than made to sit through the boot.
func New(imagePath, xbePath string) (*Adapter, error) {
	disc, err := xbox.Open(imagePath)
	if err != nil {
		return nil, err
	}
	xbeBytes, err := disc.ReadFile(xbePath)
	if err != nil {
		disc.Close()
		return nil, fmt.Errorf("read %s: %w", xbePath, err)
	}
	xbe, err := xbox.ParseXBE(xbeBytes)
	if err != nil {
		disc.Close()
		return nil, fmt.Errorf("parse XBE: %w", err)
	}
	live, err := xbox.NewMachine(xbe, disc)
	if err != nil {
		disc.Close()
		return nil, err
	}
	// The GPU is not optional under a debugger: without the pusher running, the machine
	// stops dead at the first push-buffer kick (the Phase-B milestone) and there is no
	// frame to look at.
	live.EnableGPU()
	a := &Adapter{imagePath: imagePath, xbePath: xbePath, live: live, held: map[string]bool{}}
	a.installFlipHook(live)
	return a, nil
}

// installFlipHook gives the live machine its one frame-boundary hook, which releases a
// queued pad state and then runs whatever the current step wants.
//
// The pad goes first: a state set here is live for the frame that is about to run, which
// is the frame the title will poll. Only the LIVE machine gets this — the scratch
// machine (scratchMachine) replays a captured frame, and replay is deterministic
// re-execution, not play.
func (a *Adapter) installFlipHook(m *xbox.Machine) {
	m.OnFlip = func(mm *xbox.Machine) {
		a.padPace(mm)
		if a.flipHook != nil {
			a.flipHook(mm)
		}
	}
}

func (a *Adapter) Platform() string { return "xbox" }

// Title is the name the executable gives itself. An XBE carries its own title in its
// certificate, so unlike a bare cartridge this platform need not fall back to guessing from
// the file name — and the file name here is a dump's, not the game's.
func (a *Adapter) Title() string {
	if a.live != nil && a.live.XBE != nil && a.live.XBE.TitleName != "" {
		return a.live.XBE.TitleName
	}
	return filepath.Base(a.imagePath)
}

// Machine exposes the underlying live machine, for wiring a caller needs that the generic
// interface does not cover.
func (a *Adapter) Machine() *xbox.Machine { return a.live }

// Close drops the machines. The disc file handle goes with them.
func (a *Adapter) Close() error {
	if a.live != nil {
		if d := a.live.Disc; d != nil {
			d.Close()
		}
	}
	if a.scratch != nil {
		if d := a.scratch.Disc; d != nil {
			d.Close()
		}
	}
	a.live, a.scratch = nil, nil
	return nil
}

// LoadStateFile restores a savestate the oracle wrote.
//
// A halted state is cleared on the way in. The oracle's frontier savestates stop on an
// unimplemented kernel ordinal with EIP still on the trap sentinel and nothing mutated, so
// clearing the halt retries the very call that stopped the run — the fix-and-resume
// workflow. Under a debugger that matters more than anywhere: the state worth opening is
// usually the one that stopped.
func (a *Adapter) LoadStateFile(path string) error {
	if err := a.live.LoadStateFile(path); err != nil {
		return err
	}
	a.live.ClearHalt()
	return nil
}

func (a *Adapter) SaveStateFile(path string) error { return a.live.SaveStateFile(path) }

func (a *Adapter) Snapshot() debug.Snapshot { return snap{st: a.live.SaveState()} }

func (a *Adapter) Restore(s debug.Snapshot) error {
	ns, ok := s.(snap)
	if !ok {
		return fmt.Errorf("xboxadapter: snapshot is from %q, not xbox", platformOf(s))
	}
	return a.live.LoadState(ns.st)
}

// Display is what the TV is showing: the buffer the title last presented.
//
// It is deliberately NOT the CRTC's programmed scanout, which on this machine is a blank
// white rectangle — the title registers its display mode once, during loading, and the
// 640x480 switch it later renders at never happens here. That is a real frontier, and the
// scanout is still offered as its own surface so it stays visible; but a Display() that
// returned it would make the debugger's main view white on a game that is drawing perfectly,
// which is not honesty, just a worse picture of the same fact. What the console shows is the
// buffer the title flipped to, and the machine knows exactly which that is (RenderPresented).
func (a *Adapter) Display() (*image.RGBA, error) { return a.live.RenderPresented() }

// ReadMem reads memory at a guest address. It reads RAM only, and out-of-range bytes read
// as zero: the NV2A's register aperture is in the address space too, but reading one has
// side effects — a status register's read can be what a spin loop is waiting on — and a
// debugger's memory pane must never be the thing that changes the machine.
func (a *Adapter) ReadMem(addr uint32, n int) []byte {
	out := make([]byte, n)
	a.live.ReadRAM(addr, out)
	return out
}

// ---- frames ----

// StepFrame advances the live machine to the next swap — the title's AvSetDisplayMode
// registering a new scanout — capturing that frame's push-buffer method stream and
// per-pixel last-writer provenance.
//
// The picture is the draw target: the buffer the frame was built in, which is the one the
// swap is presenting. Nothing has erased it (this console renders into plain RAM), so
// unlike the GameCube's there is no need to catch it mid-copy.
func (a *Adapter) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	fc := &debug.FrameCapture{Start: snap{st: a.live.SaveState()}}
	fc.CountWrites()

	prov := map[uint32]int32{}
	var over map[uint32][]debug.PixelWrite
	if withOverdraw {
		over = map[uint32][]debug.PixelWrite{}
	}
	cmd := -1 // index of the command currently drawing; -1 before the first

	a.live.OnNVMethod = func(_ *xbox.Machine, subchan, method, arg uint32) {
		cmd = len(fc.Commands)
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index:   cmd,
			Name:    xbox.NVMethodName(a.live.NVSubchannelClass(subchan), method),
			Op:      method,
			Words:   []uint64{uint64(arg)},
			Decoded: xbox.NVMethodDecode(a.live.NVSubchannelClass(subchan), method, arg),
		})
	}
	a.live.OnPixel = func(x, y uint32, ev xbox.PixelEvent) {
		if cmd < 0 {
			return
		}
		// Fragments arrive in SAMPLE coordinates; the picture is in logical pixels. On a
		// surface with anti-aliasing off those are the same numbers, which is exactly why
		// this is easy to leave out and impossible to notice — this title's frames have it
		// off, and the census would look flawless right up until the first anti-aliased
		// scene, where three quarters of the frame would fall outside the picture and be
		// dropped by the bounds check below. Several samples resolving to one pixel means
		// the last one wins, which is the same rule the picture's own box filter follows.
		ax, ay := a.live.SurfaceAAScale()
		x, y = x/uint32(ax), y/uint32(ay)
		key := y<<16 | x&0xFFFF
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
	// The frame's picture has to be taken inside the flip: the swap that ends a frame is
	// the title re-pointing the colour surface at the NEXT buffer, so by the time the run
	// loop regains control the register names a different buffer and rendering it would
	// show the frame from two swaps ago. (Same discipline as the GameCube's, for the
	// opposite reason: there the flip wipes the buffer, here it moves the pointer.)
	var shot *image.RGBA
	a.flipHook = func(m *xbox.Machine) {
		if shot == nil {
			shot, _ = m.RenderDrawTarget()
			m.StopRequested = true
		}
	}

	a.live.Run(runBudget)
	// Leave any user watches wired, but drop the per-frame capture hooks so a later run is
	// not paying for a census nobody is reading. OnFlip itself stays: it is the machine's
	// one frame-boundary hook and the pad's pacing lives behind it.
	a.live.OnNVMethod, a.live.OnPixel, a.flipHook = nil, nil, nil

	if shot == nil {
		return fc, nil // no frame completed within the budget: a capture with no picture
	}
	w, h := shot.Bounds().Dx(), shot.Bounds().Dy()
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

// StepFast advances the live machine one swap, capturing nothing: no frame-start snapshot
// (a 64 MiB copy of RAM), no command stream, no per-pixel census. It is how the debugger
// fast-forwards to the part of the game worth looking at.
func (a *Adapter) StepFast() error {
	a.live.OnNVMethod, a.live.OnPixel = nil, nil
	done := false
	a.flipHook = func(m *xbox.Machine) {
		if !done {
			done = true
			m.StopRequested = true
		}
	}
	a.live.Run(runBudget)
	a.flipHook = nil
	return nil
}

// RenderAfter replays fc.Start in the scratch machine and returns the draw target exactly
// after command k executed.
//
// The last command of a frame is the swap that ends it, and that swap re-points the colour
// surface at the next buffer — so a naive replay to the end of the stream would render the
// WRONG buffer, and the scrubber's final position, the one showing the finished frame,
// would show a frame two swaps old. The flip hook catches the picture before the register
// moves, exactly as StepFrame does.
func (a *Adapter) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	s, ok := fc.Start.(snap)
	if !ok {
		return nil, fmt.Errorf("xboxadapter: capture holds a %q snapshot, not xbox", platformOf(fc.Start))
	}
	if k < 0 || k >= len(fc.Commands) {
		return nil, fmt.Errorf("xboxadapter: command %d out of range [0,%d)", k, len(fc.Commands))
	}
	sc, err := a.scratchMachine()
	if err != nil {
		return nil, err
	}
	if err := sc.LoadState(s.st); err != nil {
		return nil, err
	}
	sc.ClearHalt()
	sc.OnNVMethod, sc.OnPixel = nil, nil
	var shot *image.RGBA
	sc.OnFlip = func(m *xbox.Machine) {
		if shot == nil {
			shot, _ = m.RenderDrawTarget()
		}
	}
	sc.RunStopAfterNVMethod(k+1, runBudget) // 1-based count → stop after index k
	sc.OnFlip = nil
	if shot != nil {
		return shot, nil
	}
	return sc.RenderDrawTarget()
}

// ---- CPU ----

func (a *Adapter) CPU() debug.CPUReg {
	c := a.live.CPU
	vals := make([]uint64, 8)
	for i := 0; i < 8; i++ {
		vals[i] = uint64(c.Regs[i])
	}
	flags := func(b bool, v uint64) uint64 {
		if b {
			return v
		}
		return 0
	}
	return debug.CPUReg{
		// Where the machine is PARKED — the next instruction — not the one just retired.
		PC:    uint64(a.live.PC()),
		Names: x86RegNames,
		Vals:  vals,
		Extra: map[string]uint64{
			"EFLAGS": flags(c.CF, 1<<0) | flags(c.PF, 1<<2) | flags(c.AF, 1<<4) |
				flags(c.ZF, 1<<6) | flags(c.SF, 1<<7) | flags(c.TF, 1<<8) |
				flags(c.IF, 1<<9) | flags(c.DF, 1<<10) | flags(c.OF, 1<<11),
			"CS": uint64(c.Seg[x86.CS]), "DS": uint64(c.Seg[x86.DS]),
			"SS": uint64(c.Seg[x86.SS]), "ES": uint64(c.Seg[x86.ES]),
			// FS is the one segment with a base of its own: it points at the KPCR, and
			// every per-thread read (the current thread, the TLS area, GetLastError)
			// resolves through it — so its base is worth seeing, not just its selector.
			"FS": uint64(c.Seg[x86.FS]), "FS.base": uint64(c.SegBase[x86.FS]),
			"GS":    uint64(c.Seg[x86.GS]),
			"Steps": c.Steps,
		},
	}
}

// Halted answers for the core AND for the machine around it. The Xbox is the platform where
// this matters most: an unimplemented xboxkrnl ordinal halts (kernel.go), and that is the
// port's normal frontier — so a halt here is the expected way to meet new work, not a rare
// fault. The NV2A's own clocks do not stop with the CPU, so without this the page would show
// a machine free-running past the very ordinal it is waiting to be told about.
func (a *Adapter) Halted() (bool, string) {
	if a.live.CPU.Halted {
		return true, a.live.CPU.HaltReason
	}
	return a.live.Halted, a.live.HaltReason
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

// runSlice runs the machine and works out why it stopped. Run's own reason covers a
// breakpoint, a halt or the budget; a breaking watch reports itself through a.stop,
// because from Run's point of view a hook merely asked to stop.
func (a *Adapter) runSlice(budget uint64, exhausted string) (debug.StopReason, error) {
	a.stop = debug.StopReason{}
	reason, steps := a.live.Run(budget)
	sr := a.stop
	if sr.Kind == "" {
		switch reason {
		case xbox.StopHalt:
			sr = debug.StopReason{Kind: "halted", Note: a.live.CPU.HaltReason}
		case xbox.StopBreak:
			sr = debug.StopReason{Kind: "breakpoint"}
		default:
			sr = debug.StopReason{Kind: exhausted, Note: reason.String()}
		}
	}
	sr.PC = uint64(a.live.PC())
	sr.Steps = steps
	return sr, nil
}

// Breakpoints. The adapter keeps the set so Breakpoints() can report it without asking the
// machine to sort its map on every poll.

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
	a.live.ClearBreakpoint(uint32(pc))
}

func (a *Adapter) Breakpoints() []uint64 { return append([]uint64(nil), a.bps...) }

// Disasm decodes n instructions at a linear address. x86 instructions are variable-length,
// so this walks by each decode's own length — and a decode that reports no length would
// walk forever, so it stops there rather than looping.
func (a *Adapter) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, 0, n)
	va := uint32(addr)
	for i := 0; i < n; i++ {
		var buf [16]byte
		a.live.ReadCode(va, buf[:])
		in := x86.Decode32(buf[:], va)
		length := in.Len
		if length <= 0 {
			out = append(out, debug.Instr{Addr: uint64(va), Bytes: buf[:1], Text: "(undecodable)"})
			break
		}
		out = append(out, debug.Instr{Addr: uint64(va), Bytes: buf[:length], Text: in.Text})
		va += uint32(length)
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
// watch of each kind at a time. Asking for a second of the same kind is an error rather
// than a silent replacement — quietly dropping a watch the user set is how a debugger lies.
//
// The windows see CPU accesses only. The NV2A's own reads and writes — the pusher fetching
// commands, the rasteriser storing fragments, a texture fetch — go through the machine's
// RAM directly and no watch will see them, which is a real limit of this model and is why a
// watch that never fires is not proof nobody wrote the address. The pixel hook is what sees
// the GPU's writes.

func (a *Adapter) SetWatch(w debug.Watch) (int, error) {
	if w.Kind != "write" && w.Kind != "read" {
		return 0, fmt.Errorf("xboxadapter: watch kind %q is not write or read", w.Kind)
	}
	for _, ex := range a.watches {
		if ex.Kind == w.Kind {
			return 0, fmt.Errorf("xboxadapter: the machine has one %s window, and it is in use (clear watch %d first)", w.Kind, ex.ID)
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
	a.live.SetWriteWatch(0, 0, nil)
	a.live.SetReadWatch(0, 0, nil)
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
			a.live.SetWriteWatch(w.Lo, w.Hi, cb)
		} else {
			a.live.SetReadWatch(w.Lo, w.Hi, cb)
		}
	}
}

// ---- surfaces ----

// Surfaces are the pictures this machine can show. The first two are the pair the package
// comment is about — what Kelvin drew into, and what the CRTC is scanning — and on a
// double-buffered title those are two different addresses, so seeing both is how you tell a
// drawing bug from a swap bug.
func (a *Adapter) Surfaces() []debug.Surface {
	size := func(render func() (*image.RGBA, error)) (int, int) {
		if img, err := render(); err == nil {
			return img.Rect.Dx(), img.Rect.Dy()
		}
		return 0, 0
	}
	pw, ph := size(a.live.RenderPresented)
	dw, dh := size(a.live.RenderDrawTarget)
	sw, sh := size(a.live.RenderScanout)
	out := []debug.Surface{
		{ID: "presented", Name: "Presented frame (what the TV shows)", W: pw, H: ph},
		{ID: "drawtarget", Name: "Kelvin colour surface (being drawn now)", W: dw, H: dh},
		{ID: "scanout", Name: "CRTC scanout as programmed (stale: no 640x480 mode switch)", W: sw, H: sh},
		{ID: "ram", Name: "RAM as a texture", Free: true, Formats: xbox.RegionFormats()},
	}
	// The texture units, each shown at whatever size and format the game currently has
	// bound. A unit has to be ASKED whether it is bound: an untouched one's registers read
	// back zero and decode as a 1x1 texture at address 0 rather than as nothing, so
	// offering those would be three phantom surfaces in the list.
	for u := 0; u < xbox.NVTextureUnits; u++ {
		if !a.live.TextureBound(u) {
			continue
		}
		img, err := a.live.DumpTexture(u)
		if err != nil {
			continue
		}
		out = append(out, debug.Surface{
			ID:   fmt.Sprintf("tex%d", u),
			Name: fmt.Sprintf("Bound texture %d", u),
			W:    img.Rect.Dx(), H: img.Rect.Dy(),
		})
	}
	return out
}

func (a *Adapter) RenderSurface(id string, v debug.View) (*image.RGBA, error) {
	switch id {
	case "presented":
		return a.live.RenderPresented()
	case "scanout":
		return a.live.RenderScanout()
	case "drawtarget":
		return a.live.RenderDrawTarget()
	case "ram":
		return a.live.RenderRegion(xbox.RegionSpec{
			Addr: v.Addr, W: v.W, H: v.H, Stride: v.Stride, Format: v.Format,
		})
	}
	if strings.HasPrefix(id, "tex") {
		if u, err := strconv.Atoi(strings.TrimPrefix(id, "tex")); err == nil && u >= 0 && u < xbox.NVTextureUnits {
			return a.live.DumpTexture(u)
		}
	}
	return nil, fmt.Errorf("xboxadapter: no surface %q: %w", id, debug.ErrUnsupported)
}

// ---- the disc's filesystem ----
//
// The Xbox's XISO is a real filesystem, and this title streams out of it constantly — the
// 27.9 MB title movie above all — so the file panes earn their keep: a read-watch can say
// which of the game's own files the drive is delivering right now.

// ListDir lists one directory of the disc.
func (a *Adapter) ListDir(path string) ([]debug.FileEntry, error) {
	disc := a.live.Disc
	if disc == nil {
		return nil, debug.ErrUnsupported
	}
	entries, err := disc.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]debug.FileEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, fileEntry(e))
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
	disc := a.live.Disc
	if disc == nil {
		return nil, debug.ErrUnsupported
	}
	return disc.ReadFile(path)
}

// FileAt names the file that contains a raw disc offset — what turns a read-watch on the
// drive into "the game is streaming this file right now".
func (a *Adapter) FileAt(offset int64) (debug.FileEntry, int64, bool) {
	disc := a.live.Disc
	if disc == nil {
		return debug.FileEntry{}, 0, false
	}
	e, within, ok := disc.EntryAt(offset)
	if !ok {
		return debug.FileEntry{}, 0, false
	}
	return fileEntry(e), within, true
}

func fileEntry(e xbox.Entry) debug.FileEntry {
	return debug.FileEntry{
		Path: "/" + strings.TrimPrefix(e.Path, "/"), Name: e.Name, Dir: e.IsDir,
		Size: int64(e.Size), Offset: e.Offset(),
	}
}

// ---- memory map ----

// Regions names the parts of the address space the memory pane can show.
//
// The Xbox's 64 MB of RAM appears at four addresses at once — the identity mapping the
// title's image and stacks live at, and the cached/uncached/write-combined kernel windows
// that fold onto the same bytes. They are listed as what they are rather than collapsed to
// one, because the title's own pointers arrive in all four flavours and a pointer's window
// is a fact about it: a D3D surface pointer is 0xF0000000-based precisely so CPU blits
// bypass the cache, and seeing that is how the write-combined window got modelled at all.
func (a *Adapter) Regions() []debug.Region {
	return []debug.Region{
		{Name: "RAM (identity: the title, stacks, heap)", Lo: 0x0000_0000, Hi: 0x0400_0000},
		{Name: "RAM (cached kernel window)", Lo: 0x8000_0000, Hi: 0x8400_0000},
		{Name: "RAM (uncached kernel window)", Lo: 0xB000_0000, Hi: 0xB400_0000},
		{Name: "RAM (write-combined: D3D surface pointers)", Lo: 0xF000_0000, Hi: 0xF400_0000},
	}
}

// ResumeArgs is the command line that resumes this disc from a saved state. OutRun's
// bootoracle spells the flag -loadstate, and -gpu because without the pusher the machine
// stops at the first push-buffer kick instead of rendering.
func (a *Adapter) ResumeArgs(statePath string) []string {
	return []string{
		"go", "run", "./cmd/bootoracle",
		"-image", a.imagePath,
		"-loadstate", statePath,
		"-gpu",
	}
}

// ---- helpers ----

func (a *Adapter) scratchMachine() (*xbox.Machine, error) {
	if a.scratch == nil {
		disc, err := xbox.Open(a.imagePath)
		if err != nil {
			return nil, err
		}
		xbeBytes, err := disc.ReadFile(a.xbePath)
		if err != nil {
			disc.Close()
			return nil, err
		}
		xbe, err := xbox.ParseXBE(xbeBytes)
		if err != nil {
			disc.Close()
			return nil, err
		}
		sc, err := xbox.NewMachine(xbe, disc)
		if err != nil {
			disc.Close()
			return nil, err
		}
		sc.EnableGPU()
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
