// Package debug defines a platform-agnostic surface a frame-debugger GUI can
// drive: advance a console emulator one video frame at a time, inspect its CPU
// and memory, watch addresses, and — the reason the package exists — capture a
// frame's whole display-processor command stream, replay it command by command,
// and attribute each rendered pixel to the command that drew it.
//
// The interface is deliberately small and honest. Everything here is expressible
// today for the N64 oracle (see tools/debug/n64adapter); PSX, PSP and 3DS adapters
// can follow, and a method a platform cannot provide returns ErrUnsupported rather
// than lying. All calls are synchronous and single-threaded: the GUI owns the
// goroutine that touches a DebugTarget.
package debug

import (
	"errors"
	"image"
)

// ErrUnsupported is returned by a DebugTarget method a particular platform cannot
// implement. Callers should treat it as "this pane is empty here", not a failure.
var ErrUnsupported = errors.New("debug: unsupported on this target")

// Snapshot is an opaque, in-memory machine state, independent of the live machine
// (a deep copy), so a replay can restore it into a scratch target while the live
// one keeps running. Each adapter returns its own concrete type; Platform names
// the platform it came from, so a target can reject a foreign snapshot.
type Snapshot interface {
	Platform() string
}

// CPUReg is a flat register view for the register pane. Names are platform-defined
// so the same pane renders a VR4300, an R3000, an Allegrex, and so on. Names and
// Vals are parallel and equal length.
type CPUReg struct {
	PC    uint64
	Names []string
	Vals  []uint64
	Extra map[string]uint64 // HI/LO, COP0 Status/Cause, and the like
}

// GPUCommand is one decoded display-processor command (an RDP command on N64, a
// GP0 packet on PSX, a GE command on PSP). Words are the raw command words, kept
// so a pane can show the bytes; Decoded is a one-line human decode, or "" if the
// platform offers none.
type GPUCommand struct {
	Index   int
	Name    string
	Op      uint32
	Words   []uint64
	Decoded string
}

// PixelWrite is one attributed write to a pixel. Rejected marks a pixel the
// rasteriser produced but discarded (a depth or alpha failure): provenance without
// a store, which is exactly what a "why is this pixel not what I expect" question
// needs.
type PixelWrite struct {
	CmdIndex   int
	R, G, B, A uint8
	Rejected   bool
}

// FrameCapture is everything recorded for one frame.
type FrameCapture struct {
	// Start is the machine state at the top of the frame, for RenderAfter to
	// replay from.
	Start Snapshot

	// Commands is the frame's display-processor command stream, in execution
	// order; Command i has Index i.
	Commands []GPUCommand

	Width, Height int

	// Prov is the last command to write each pixel, row-major (y*Width+x); -1
	// means no command touched that pixel. len(Prov) == Width*Height, or Prov is
	// nil if the platform offers no per-pixel provenance.
	Prov []int32

	// Overdraw, when requested, is the full ordered write history of each pixel,
	// keyed by y*Width+x. Nil when not recorded.
	Overdraw map[int][]PixelWrite
}

// ProvAt returns the index of the last command to write pixel (x, y), or -1 if
// none did or the point is out of range or provenance was not recorded.
func (fc *FrameCapture) ProvAt(x, y int) int {
	if fc.Prov == nil || x < 0 || y < 0 || x >= fc.Width || y >= fc.Height {
		return -1
	}
	return int(fc.Prov[y*fc.Width+x])
}

// DebugTarget is a console emulator a frame-debugger can drive.
type DebugTarget interface {
	// Name is a human label for the target ("Nintendo 64", a game title).
	Name() string

	// Snapshot captures the machine's whole state; Restore puts one back. Restore
	// rejects a snapshot from another platform.
	Snapshot() Snapshot
	Restore(Snapshot) error

	// StepFrame advances to the next completed frame, capturing that frame's
	// command stream and per-pixel last-writer provenance (and the full overdraw
	// history when withOverdraw is set). The returned capture's Start is a
	// snapshot taken at the top of the frame, so RenderAfter can replay it.
	StepFrame(withOverdraw bool) (*FrameCapture, error)

	// RenderAfter replays fc.Start and returns the draw target exactly as it stood
	// after command k of fc.Commands executed (k is a 0-based index into
	// Commands). It is deterministic: the same fc and k always yield the same
	// image.
	RenderAfter(fc *FrameCapture, k int) (*image.RGBA, error)

	// CPU returns the current register file for the register pane.
	CPU() CPUReg

	// Display is the image currently being scanned out to the screen.
	Display() (*image.RGBA, error)

	// ReadMem reads n bytes of main memory starting at a physical address, for the
	// memory pane. Out-of-range bytes read as zero.
	ReadMem(addr uint32, n int) []byte

	// SetWriteWatch / SetReadWatch install a callback fired when the CPU writes /
	// reads an address in [lo, hi). A nil cb clears the watch.
	SetWriteWatch(lo, hi uint32, cb func(addr, val, pc uint32))
	SetReadWatch(lo, hi uint32, cb func(addr, val, pc uint32))

	// SetBreakpoint stops the next StepFrame when execution reaches vaddr.
	SetBreakpoint(vaddr uint64)
}
