// Package debug defines what a console emulator has to offer a debugger, and it
// does so as a small required core plus a set of optional capabilities.
//
// The split is not decoration. The platforms in this repository genuinely differ:
// only the N64 rasteriser reports every pixel it draws, only the N64 can stop
// mid-way through a display list, the N64 has no filesystem while the 3DS has a
// whole RomFS, and the 3DS has no in-memory snapshot at all. A single fat interface
// would force most adapters to stub out most of it and return "unsupported" — which
// is a lie dressed as an API, and it pushes the question "can this platform actually
// do that?" onto the caller at the worst possible moment.
//
// So: a Target must do a handful of things every emulator here can do, and it
// advertises anything further by implementing an optional interface. Capabilities
// names them, the server sends that list to the browser, and the page shows exactly
// the tools the current target can back — no empty panels, nothing faked.
//
// All calls on a Target are synchronous and single-threaded: exactly one goroutine
// ever touches one (the server's Runner).
package debug

import (
	"errors"
	"image"
)

// ErrUnsupported is returned when a capability exists but a particular call cannot
// be served — a surface a target does not know, a path that is not in its
// filesystem. It is not the way to say "this platform lacks this feature": that is
// said by not implementing the capability at all.
var ErrUnsupported = errors.New("debug: unsupported on this target")

// Target is what every debuggable machine can do.
type Target interface {
	// Platform is the machine ("n64", "psx"); Title is the game.
	Platform() string
	Title() string

	// CPU returns the current register file.
	CPU() CPUReg

	// ReadMem reads n bytes of main memory from a physical address. Out-of-range
	// bytes read as zero.
	ReadMem(addr uint32, n int) []byte

	// Display is the image currently being scanned out to the screen.
	Display() (*image.RGBA, error)

	// Snapshot captures the whole machine state; Restore puts one back. The
	// snapshot must be independent of the live machine (a deep copy), so it can be
	// restored repeatedly — that is what makes deterministic replay possible.
	Snapshot() Snapshot
	Restore(Snapshot) error

	// Close releases the machine. A target is not used again after Close.
	Close() error
}

// The optional capabilities. A target implements what its machine can honestly back.
type (
	// FrameStepper advances a video frame at a time and reports what the display
	// processor did. A platform needs a frame-boundary hook and a command hook to
	// offer this; per-pixel provenance (FrameCapture.Prov) additionally needs a
	// pixel hook, and is allowed to be nil without giving up the capability.
	FrameStepper interface {
		StepFrame(withOverdraw bool) (*FrameCapture, error)
	}

	// FastStepper advances one frame capturing nothing — no snapshot, no command
	// list, no pixel census. It is how the debugger fast-forwards to the part of a
	// game worth looking at. A target without it simply plays at capture speed.
	FastStepper interface {
		StepFast() error
	}

	// FrameReplayer renders the draw target as it stood after command k of a
	// capture, by replaying the frame from its start snapshot. This is the command
	// scrubber. It needs a way to halt the display processor mid-list.
	FrameReplayer interface {
		RenderAfter(fc *FrameCapture, k int) (*image.RGBA, error)
	}

	// BlankReplayer can start a replay from a blank screen instead of from what the
	// previous frame left behind.
	//
	// It exists because a console's screen is not cleared between frames: a scrub to the
	// top of a frame shows the PREVIOUS frame, and what this frame goes on to draw appears
	// over it, which makes it hard to see what this frame is actually responsible for. With
	// this on, anything the frame has not drawn yet reads as black.
	//
	// It is a debugging lie, and a deliberate one: a game that deliberately builds on the
	// previous frame's buffer (a cheap motion blur, an accumulation pass) really does depend
	// on those pixels, and with this on its frame will look wrong. Hence a switch, off by
	// default, rather than a decision.
	BlankReplayer interface {
		SetReplayBlank(on bool)
	}

	// CodeStepper runs the CPU rather than the video hardware. Both calls are
	// bounded: they execute at most budget instructions and return, stopping early
	// on a breakpoint, a halt, or a watch marked to break.
	//
	// There is deliberately no Interrupt method. A "stop the run" flag written from
	// the socket's goroutine would race the emulator on every instruction; instead
	// the caller runs in bounded slices and decides between them whether to go on.
	// That keeps the single-threaded contract intact.
	CodeStepper interface {
		StepInstr(n int) (StopReason, error)
		Continue(budget uint64) (StopReason, error)
	}

	// Haltable reports a machine whose core has stopped executing and will not start
	// again — an emulator refusing to invent behaviour for hardware it does not model,
	// not a game that crashed.
	//
	// It exists because a halt is otherwise INVISIBLE at the frame level, and invisible
	// in a way that impersonates a running machine. The core stops, but the video and
	// audio clocks do not: fields keep retiring, so play mode keeps stepping, the
	// profile keeps reporting, the frame counter keeps climbing, and the screen holds
	// the last picture. That reads as a game stuck in a loop at several hundred frames a
	// second — and it cost a session's debugging on the GameCube, where the true story
	// ("PE copy-to-texture: half-scale box filter not yet implemented") was sitting in a
	// string in memory the whole time while the page showed a silent 400 fps.
	//
	// The reason string is the whole point: a target that answers true says WHY in the
	// same breath, because a halt the page cannot name is barely better than the hang it
	// already looked like. CodeStepper's StopReason already carries this for a stepped
	// run; this is the same fact for a target being played rather than stepped.
	Haltable interface {
		Halted() (halted bool, reason string)
	}

	// Breakpointer holds execution breakpoints.
	Breakpointer interface {
		SetBreakpoint(pc uint64)
		ClearBreakpoint(pc uint64)
		Breakpoints() []uint64
	}

	// Disassembler decodes instructions at an address.
	Disassembler interface {
		Disasm(addr uint64, n int) ([]Instr, error)
	}

	// Watcher reports reads and writes to an address range as they happen. Hits are
	// delivered to the sink installed by OnWatchHit, called from inside the run —
	// so the sink must be cheap and must not touch the target.
	Watcher interface {
		SetWatch(w Watch) (id int, err error)
		ClearWatch(id int)
		Watches() []Watch
		OnWatchHit(sink func(WatchHit))
	}

	// Surfacer renders memory as an image — the scanout, the draw target, raw VRAM,
	// texture memory, or a region the user aims by hand. Pixel formats are the
	// platform's business, so the decoding lives in the platform package and the
	// adapter forwards to it.
	Surfacer interface {
		Surfaces() []Surface
		RenderSurface(id string, v View) (*image.RGBA, error)
	}

	// FileLister browses the game's own filesystem (a disc, a RomFS). Cartridge
	// platforms have none and do not implement this.
	FileLister interface {
		ListDir(path string) ([]FileEntry, error)
		ReadFile(path string) ([]byte, error)
	}

	// FileAttributer maps a raw offset in the image back to the file that contains
	// it — so a read-watch can say which file the machine is streaming right now.
	FileAttributer interface {
		FileAt(offset int64) (entry FileEntry, within int64, ok bool)
	}

	// StateFiler reads and writes the platform's own savestate files — the same
	// format that platform's bootoracle loads, so a state saved here can be handed
	// straight to another tool.
	StateFiler interface {
		SaveStateFile(path string) error
		LoadStateFile(path string) error
	}

	// Resumer returns the command line that resumes this target from a savestate.
	// The oracles do not agree on their flags (-loadstate, -load, -state), so each
	// adapter emits what its own game's oracle actually accepts rather than
	// pretending there is a common vocabulary.
	Resumer interface {
		ResumeArgs(statePath string) []string
	}

	// ResumeDirer says which directory a Resumer's command runs in, when it is not
	// the game's own extract/. Platforms whose titles share one oracle need this:
	// Captain Toad boots on Super Mario 3D Land's bootoracle, so a resume line that
	// assumed its own game directory would cd into a directory that is not there.
	ResumeDirer interface {
		ResumeDir() string
	}

	// MemoryMapper names the regions of the address space, for the memory pane.
	MemoryMapper interface {
		Regions() []Region
	}

	// BatchReplayer renders several points of the command stream at once.
	//
	// Every RenderAfter(k) is an independent replay of the same frame from the same
	// snapshot, on a machine the caller throws away — so they can run at the same
	// time, on as many machines as the target cares to keep, with no determinism
	// question to answer at all. A target that offers this makes the command scrubber
	// stop costing one full replay per position dragged over.
	//
	// The results are returned in the order the keys were given. A target without it
	// is simply asked for one k at a time.
	BatchReplayer interface {
		RenderAfterBatch(fc *FrameCapture, ks []int) ([]*image.RGBA, error)
	}

	// Toucher accepts stylus input.
	//
	// A touch panel is a region of the composed frame — the DS's lower LCD, the 3DS's
	// bottom screen — so a click on the picture the debugger is already showing IS the
	// touch, and there is no second coordinate system to reconcile.
	//
	// Panels are asked for rather than declared once, because on the 3DS the frame's
	// layout is not fixed: it is whatever DisplayTransfers the game has made, and before
	// it has presented the bottom screen there is no panel to touch.
	Toucher interface {
		TouchPanels() []TouchPanel
		Touch(t Touch) error
	}

	// Keyer accepts keyboard input.
	//
	// It is the keyboard counterpart to Toucher: where a touch is a point on a panel of
	// the composed frame, a key is a whole-keyboard event with no place on the picture —
	// so it carries a key identity rather than coordinates. A press is a Key with Down
	// set, a release one with it clear, and BOTH must be delivered: on a machine that
	// tracks make/break scancodes (a PC keyboard) dropping either desyncs the key state,
	// so unlike a touch drag these events never collapse to the latest.
	//
	// It is named generically — Keyer, not "DOSKeyboard" — because it is the first input
	// capability of its shape, and a console's button input should slot in beside it
	// later exactly as this did beside Toucher.
	Keyer interface {
		Key(k Key) error
	}

	// KeyLegender is an optional companion to Keyer: what this target's keys DO, in one
	// short phrase, for the page to show beside the frame.
	//
	// It exists because a key mapping is only self-describing to whoever wrote it. The
	// GameCube's letters are the buttons' own names and need no legend; the PS2's face
	// buttons are W/A/S/Z, a diamond on the keyboard standing for the diamond on the pad,
	// which is obvious once seen and invisible until then. The alternative — a table of
	// platforms in the page — would put the answer somewhere the target cannot correct,
	// and this page's whole shape is that it builds itself from what the target says.
	//
	// A Keyer that does not implement it gets a generic phrase, which is the right default
	// for a keyboard that is simply a keyboard.
	KeyLegender interface {
		KeyLegend() string
	}

	// DisplayAspecter is an optional statement of the picture's intended display shape, for a
	// target whose pixels are not square. The page presents Display()/the frame at this
	// width:height by scaling the display (never by resampling the image), so the pixel grid
	// — and therefore provenance click-mapping — stays exactly aligned.
	//
	// A target that does not implement it is shown pixel-for-pixel (square pixels), which is
	// right for every render-to-square-buffer platform. The PS2 needs it: its DAC stretches
	// whatever framebuffer width it renders (512, 640, …) to fill a 4:3 standard-def screen,
	// so a woven 640×448 or 512×448 frame is not 4:3 on its own.
	DisplayAspecter interface {
		DisplayAspect() (num, den int)
	}

	// Profiler reports where the last stepped frame's time went, by subsystem.
	//
	// The times are the emulator's own, not a sampling profiler's: only a machine
	// knows that this stretch of nanoseconds was its rasteriser and that one was its
	// audio DSP. A target offers this only if it can time its subsystems without
	// distorting them — timing every fragment costs more than the fragment does — so
	// the honest granularity is a draw, a command list, a supervisor call, and the
	// interpreter's share is a remainder rather than a measurement.
	Profiler interface {
		FrameProfile() FrameProfile
	}
)

// Capability names, as sent to the page. A panel declares which one it needs.
const (
	CapFrames   = "frames"   // FrameStepper
	CapFastStep = "faststep" // FastStepper
	CapReplay   = "replay"   // FrameReplayer
	CapCode     = "code"     // CodeStepper
	CapBreak    = "break"    // Breakpointer
	CapDisasm   = "disasm"   // Disassembler
	CapWatch    = "watch"    // Watcher
	CapSurfaces = "surfaces" // Surfacer
	CapFiles    = "files"    // FileLister
	CapFileAt   = "fileat"   // FileAttributer
	CapStates   = "states"   // StateFiler
	CapResume   = "resume"   // Resumer
	CapRegions  = "regions"  // MemoryMapper
	CapProfile  = "profile"  // Profiler
	CapTouch    = "touch"    // Toucher
	CapKeys     = "keys"     // Keyer
	CapHalt     = "halt"     // Haltable
	CapOverdraw = "overdraw" // reported per capture, not per target — see FrameCapture
)

// Capabilities lists what a target can do, by asking it.
func Capabilities(t Target) []string {
	var caps []string
	add := func(ok bool, name string) {
		if ok {
			caps = append(caps, name)
		}
	}
	_, frames := t.(FrameStepper)
	_, fast := t.(FastStepper)
	_, replay := t.(FrameReplayer)
	_, code := t.(CodeStepper)
	_, brk := t.(Breakpointer)
	_, dis := t.(Disassembler)
	_, watch := t.(Watcher)
	_, surf := t.(Surfacer)
	_, files := t.(FileLister)
	_, fileAt := t.(FileAttributer)
	_, states := t.(StateFiler)
	_, resume := t.(Resumer)
	_, regions := t.(MemoryMapper)
	_, prof := t.(Profiler)
	_, touch := t.(Toucher)
	_, keys := t.(Keyer)
	_, halt := t.(Haltable)

	add(frames, CapFrames)
	add(fast, CapFastStep)
	add(replay, CapReplay)
	add(code, CapCode)
	add(brk, CapBreak)
	add(dis, CapDisasm)
	add(watch, CapWatch)
	add(surf, CapSurfaces)
	add(files, CapFiles)
	add(fileAt, CapFileAt)
	add(states, CapStates)
	add(resume, CapResume)
	add(regions, CapRegions)
	add(prof, CapProfile)
	add(touch, CapTouch)
	add(keys, CapKeys)
	add(halt, CapHalt)
	return caps
}

// TouchPanel is a panel the stylus can reach, located in the plane Display() returns —
// so the debugger can hand a click on the frame it is showing straight to the machine.
type TouchPanel struct {
	ID   string // "bottom"
	Name string
	X, Y int // origin within the composed frame
	W, H int // size: the panel's own pixels, which are the touchscreen's own
}

// Touch is one stylus state: where the pen is on a panel, or that it is lifted.
type Touch struct {
	Panel string
	X, Y  int // panel-local pixels
	Down  bool
}

// Key is one keyboard event: a key going down (Down) or coming up. Name is a
// platform-neutral key identity the target maps to its own key coding — the browser
// sends the same names the machine's scripted-input vocabulary uses ("up", "enter",
// "esc", "a"), and the adapter translates. Code carries the raw browser key code
// when a name is not enough, but Name is what targets key off.
type Key struct {
	Name string
	Code int
	Down bool
}

// FrameProfile is where one frame's time went.
//
// Buckets are disjoint and sum to TotalMs, so they can be drawn as one bar. What a
// target cannot attribute it reports as a bucket that says so, rather than
// distributing it and calling the result a measurement.
//
// Counters carry the work the frame did — fragments, draws, instructions — because
// milliseconds alone cannot tell a faster rasteriser from a frame that drew less,
// and that is the mistake this panel exists to prevent.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter

	// Drew says whether this field put a picture on the screen. A console's profiler
	// measures a FIELD — one VBlank — and a field is not a frame: a game rendering at
	// 30 Hz on a 60 Hz console draws in every other one and spends the rest on logic
	// alone. Only the caller can know what to do about that, so the target reports it
	// rather than deciding, and the debugger folds the fields that drew nothing into
	// the next one that did.
	Drew bool
}

// ProfileBucket is one subsystem's share of the frame. Count is how many of the
// thing it did: draws, command lists, cache misses, supervisor calls.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int
}

// ProfileCounter is one tally for the frame.
type ProfileCounter struct {
	Name  string
	Value int
}

// Snapshot is an opaque, in-memory machine state, independent of the live machine,
// so a replay can restore it into a scratch target while the live one keeps running.
// Platform names where it came from, so a target can reject a foreign snapshot.
type Snapshot interface {
	Platform() string
}

// CPUReg is a flat register view. Names are platform-defined, so the same pane
// renders a VR4300, an R3000 and an Allegrex. Names and Vals are parallel.
type CPUReg struct {
	PC    uint64
	Names []string
	Vals  []uint64
	Extra map[string]uint64 // HI/LO, COP0 Status/Cause, and the like
}

// Instr is one disassembled instruction.
type Instr struct {
	Addr  uint64
	Bytes []byte
	Text  string
}

// StopReason says why a run ended.
type StopReason struct {
	Kind  string // "steps", "breakpoint", "watch", "halted", "interrupt", "frame"
	PC    uint64
	Steps uint64
	Note  string // free text: the halt reason, the watch that fired
}

// Watch is a memory watch. Kind is "write" or "read"; a watch with Break set stops
// the run when it fires, rather than only reporting.
type Watch struct {
	ID     int
	Kind   string
	Lo, Hi uint32
	Break  bool
	Hits   uint64
}

// WatchHit is one observed access, reported as an event.
type WatchHit struct {
	ID    int
	Kind  string
	Addr  uint32
	Val   uint32
	PC    uint32
	Instr string // the instruction at PC, when the target can disassemble
}

// Surface is something a target can draw as an image: a named, fixed one (the
// scanout, the draw target, VRAM, texture memory) or a free one the caller aims with
// a View. Formats lists the pixel formats a free surface accepts.
type Surface struct {
	ID      string
	Name    string
	Free    bool // the caller supplies Addr/W/H/Format
	W, H    int  // natural size, for a fixed surface
	Formats []string
}

// View aims a free surface at memory.
//
// Addr is whatever addresses that surface's memory: a physical address in RAM, or —
// where the memory is not byte-addressed at all — the platform's own index into it. The
// PSX's VRAM is a 1024×512 grid of 16-bit pixels, for instance, so its surfaces read
// Addr as y*1024+x. The adapter documents what it means; the debugger just passes it on.
type View struct {
	Addr    uint32
	W, H    int
	Stride  int    // bytes per row; 0 = tightly packed
	Format  string // one of the surface's Formats
	Palette uint32 // where the palette lives, for the indexed formats
}

// FileEntry is one entry in a game's filesystem.
type FileEntry struct {
	Path   string
	Name   string
	Dir    bool
	Size   int64
	Offset int64 // where it lives in the image, for FileAttributer
}

// Region is a named span of the address space.
type Region struct {
	Name   string
	Lo, Hi uint32
}

// GPUCommand is one decoded display-processor command (an RDP command on N64, a GP0
// packet on PSX, a GE command on PSP). Words are the raw command words; Decoded is a
// one-line human decode, or "" if the platform offers none.
type GPUCommand struct {
	Index   int
	Name    string
	Op      uint32
	Words   []uint64
	Decoded string
}

// PixelWrite is one attributed write to a pixel. Rejected marks a pixel the
// rasteriser produced but discarded (a depth or alpha failure): provenance without a
// store, which is exactly what "why is this pixel not what I expect" needs.
type PixelWrite struct {
	CmdIndex   int
	R, G, B, A uint8
	Rejected   bool
}

// FrameCapture is everything recorded for one frame.
type FrameCapture struct {
	// Start is the machine state at the top of the frame, for RenderAfter to replay
	// from.
	Start Snapshot

	// Commands is the frame's display-processor command stream, in execution order;
	// command i has Index i. What counts as one "command" is the platform's own
	// granularity — an RDP command on N64, a GP0 primitive on PSX.
	Commands []GPUCommand

	Width, Height int

	// Prov is the last command to write each pixel, row-major (y*Width+x); -1 means
	// no command touched it. len(Prov) == Width*Height, or Prov is nil on a platform
	// whose rasteriser does not report pixels.
	Prov []int32

	// Overdraw, when requested, is the full ordered write history of each pixel,
	// keyed by y*Width+x. Nil when not recorded.
	Overdraw map[int][]PixelWrite

	// Writers are the commands that actually put pixels into memory, in execution order.
	//
	// Most of a display processor's command stream does not draw. A 3DS frame is a hundred
	// thousand PICA register writes and a few hundred of them are draws; the rest set up the
	// state the draws run against. Scrubbing through all of them is scrubbing through mostly
	// nothing — the picture stands still for thousands of positions and then jumps. Writers
	// is what makes the scrubber a tour of how the frame was BUILT.
	//
	// The test is a write to VRAM, not a write to the picture. A command that draws into an
	// off-screen target, a shadow map, a buffer no screen ever reads, or that uploads a
	// texture into a corner of VRAM, wrote — and you may well be looking at that corner. A
	// fragment the rasteriser produced and then discarded (a depth or alpha failure) did not
	// write, and is not one.
	//
	// Nil means the platform does not report pixels at all, and a reader should fall back to
	// the whole command stream. Empty means it reports them and this frame stored none.
	Writers []int

	// lastWrite makes MarkWrite cheap enough to call per fragment — which is how it is
	// called, several hundred thousand times in a 3DS frame.
	lastWrite int
	wrote     map[int]bool
}

// MarkWrite records that command i stored a pixel in memory. Adapters call it from their
// pixel hook, for stored fragments only.
//
// A command's fragments are contiguous — a draw rasterises before the next command runs —
// so the common case is the same index as last time, and that costs one comparison. The
// order Writers comes out in is therefore execution order, with no sort.
func (fc *FrameCapture) MarkWrite(i int) {
	if i < 0 || i == fc.lastWrite {
		return
	}
	fc.lastWrite = i
	if fc.wrote == nil {
		fc.wrote = map[int]bool{}
	}
	if fc.wrote[i] {
		return // a platform that interleaves two commands' fragments: seen, already listed
	}
	fc.wrote[i] = true
	fc.Writers = append(fc.Writers, i)
}

// CountWrites tells a capture that this platform counts writes, so an empty Writers means
// "nothing drew" rather than "nobody was counting". Adapters that call MarkWrite call this
// once, whether or not the frame drew anything.
func (fc *FrameCapture) CountWrites() {
	fc.lastWrite = -1
	if fc.Writers == nil {
		fc.Writers = []int{}
	}
}

// ProvAt returns the index of the last command to write pixel (x, y), or -1 if none
// did, the point is out of range, or provenance was not recorded.
func (fc *FrameCapture) ProvAt(x, y int) int {
	if fc.Prov == nil || x < 0 || y < 0 || x >= fc.Width || y >= fc.Height {
		return -1
	}
	return int(fc.Prov[y*fc.Width+x])
}

// Drawn reports whether this capture is a frame worth stopping on: one that actually put
// a picture on the screen, rather than a field the game spent setting up.
//
// The obvious test — "did it issue a lot of commands?" — is a packet machine's idea of a
// frame, and it is wrong on a machine whose commands are big. A whole Need for Speed
// frame is THIRTEEN cels, because a 3DO cel is a textured quad and thirteen of them cover
// the screen; a Ridge Racer frame is thousands of GP0 packets, and a 3DS frame is a
// hundred thousand register writes. Counting commands would make the debugger skip past
// every frame the 3DO ever draws.
//
// So the question is asked of the pixels instead: did this frame's commands cover a
// meaningful part of the picture? A platform whose rasteriser reports no pixels at all
// (Prov nil is allowed) has nothing to answer with, and falls back to the command count.
func (fc *FrameCapture) Drawn() bool {
	if len(fc.Commands) == 0 {
		return false
	}
	if fc.Prov == nil {
		return len(fc.Commands) > 100
	}
	n := 0
	for _, c := range fc.Prov {
		if c >= 0 {
			n++
		}
	}
	// One percent of the frame, or a command list long enough to speak for itself.
	return n*100 >= len(fc.Prov) || len(fc.Commands) > 100
}
