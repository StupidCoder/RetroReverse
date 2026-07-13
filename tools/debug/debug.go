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
	return caps
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
}

// ProvAt returns the index of the last command to write pixel (x, y), or -1 if none
// did, the point is out of range, or provenance was not recorded.
func (fc *FrameCapture) ProvAt(x, y int) int {
	if fc.Prov == nil || x < 0 || y < 0 || x >= fc.Width || y >= fc.Height {
		return -1
	}
	return int(fc.Prov[y*fc.Width+x])
}
