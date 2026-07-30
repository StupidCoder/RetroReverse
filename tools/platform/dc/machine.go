package dc

// machine.go is the Dreamcast outside the SH-4: the memories, the Holly
// system bus with its interrupt fan-in (holly.go), the PVR register file
// video.go scans out of, and the bus dispatch that ties them to the CPU.
//
// The machine implements sh4.Bus over *physical* addresses — the CPU has
// already stripped the P1/P2/P3 mirror bits — and sh4.Fetcher so a fetch
// from a register block is its own loud event rather than a data read.
//
// Every register this file does not model follows one discipline: the access
// is recorded once per address in the gap log and surfaced by Census(), reads
// return zero, and nothing pretends to be ready. A stub that reads "ready" is
// a lie the boot believes until the frame it doesn't.

import (
	"fmt"
	"io"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/cpu/sh4"
)

// Memory geometry.
const (
	RAMSize     = 16 << 20
	VRAMSize    = 8 << 20
	AICARAMSize = 2 << 20
	FlashSize   = 256 << 10

	ramBase   = 0x0C000000 // area 3, mirrored through 0x0FFFFFFF
	vram64    = 0x04000000 // the texture path's interleaved window
	vram32    = 0x05000000 // the linear window the framebuffer lives in
	aicaRAM   = 0x00800000
	flashBase = 0x00200000

	sbBase   = 0x005F6800 // Holly system block
	gdBase   = 0x005F7000 // GD-ROM ATA registers (unmodelled, gap-logged)
	pvrBase  = 0x005F8000 // PVR register file, stored raw
	aicaBase = 0x00700000 // AICA registers (unmodelled, gap-logged)
	taBase   = 0x10000000 // tile-accelerator FIFO window
	taEnd    = 0x14000000
)

// fieldInstructions is one 60 Hz field of the ~200 MHz core at the one-
// instruction-per-clock pacing this model uses. The VBlank heartbeat and the
// spin-detection window derive from it.
const fieldInstructions = 3_333_333

// Machine is a Dreamcast.
type Machine struct {
	RAM     []byte
	VRAM    []byte
	AICARAM []byte
	Flash   []byte

	CPU  *sh4.CPU
	Disc *Disc

	// The AICA's ARM7 (aica.go): held in reset until the game releases it,
	// paced at one step per armDivisor SH-4 instructions.
	ARM        *arm.CPU
	ARMRunning bool
	AICARegs   map[uint32]uint32
	Timers     aicaTimers
	Slots      [64]AICASlot
	armAcc     int

	Holly Holly

	// Ch2 (PVR) DMA registers; the transfer itself is the counted-not-
	// rendered model in holly.go.
	C2DStat, C2DLen uint32

	// Deferred completions (holly.go): a render and a ch2 DMA take time on
	// hardware, and an interrupt delivered inside the very store that starts
	// the operation arrives before the guest has armed its own bookkeeping —
	// the completion is real and still lost. Countdowns in instructions.
	RenderCountdown uint32
	C2DCountdown    uint32
	C2DPendingBits  uint32

	// TAList is the list type the TA currently has open (-1 none), TAVtx
	// the current vertex parameter size; see holly.go's taFeed.
	// TAFrame records the current TA session's polygon-path stream for the
	// rasteriser (pvr.go); TA_LIST_INIT stashes it into TAClosed — the TA is
	// double-buffered, and the session STARTRENDER draws is the one that
	// closed, not the one now recording. The stream arrives two ways the
	// hardware also has: ch2 DMA jobs, and direct FIFO stores (the game's
	// store queues) accumulated a word at a time in TAFifo. TANeed counts the
	// bytes still owed on a 64-byte parameter; TAFifoCountdown paces the
	// list-done interrupts a FIFO-submitted END_OF_LIST earns.
	TAList          int32
	TAVtx           uint32
	TANeed          uint32
	TAFrame         []byte
	TAClosed        []byte
	TAFifo          []byte
	TAFifoCountdown uint32

	// The PVR register file, stored raw at word granularity; video.go
	// interprets the scanout set, everything else round-trips.
	PVRRegs [0x2000 / 4]uint32

	// TAWrites counts tile-accelerator words submitted by either path.
	TAWrites uint64

	// Maple is the controller bus register file; Pad the injectable port-A
	// controller state (maple.go).
	Maple mapleState
	Pad   PadState

	Instrs       uint64 // instructions retired, the run budget's unit
	Fields       uint64 // VBlank heartbeats delivered
	CurLine      uint32 // the scanline the SPG is sweeping (run.go)
	FieldNum     uint32 // interlace field parity, SPG_STATUS bit 10
	instrInField uint64 // instructions accumulated toward the next line

	bios *biosHLE // installed by Boot; nil on a bare machine

	// Audio capture (aica_synth.go): with AudioCapture set, every mixed
	// stereo pair lands in AudioPCM. Pure instrument — the mix runs either
	// way, and the buffer is never savestated or hashed.
	AudioCapture bool
	AudioPCM     []int16

	// Instrumentation, all opt-in and nil-checked on the hot path.
	OnStep    func(pc uint32)
	OnDisplay func(field uint64)
	OnGDRead  func(fad, count, buf uint32) // every HLE'd disc read, with its RAM destination
	// OnC2DTexture sees every ch2 texture-path DMA: RAM source, VRAM
	// destination (64-bit-path offset) and byte length — the texture
	// placement the loader performs.
	OnC2DTexture func(src, dst, length uint32)
	Verbose   io.Writer               // gap-log lines land here when set

	// StopRequested asks the run loop to return before the next instruction.
	// A hook sets it (a breaking watch, a frame captured); Run clears it on
	// the way out, so a stale request cannot stop the next run.
	StopRequested bool

	// Frame-capture hooks (all nil in the oracle path), called from
	// renderFrame — a plain Go call inside the run loop, so no sentinel panic
	// is needed to stop mid-frame. OnPVRClear reports the background clear
	// that opens a render (it writes every pixel of the frame, and it is
	// usually the answer to "why is this pixel black"); OnPVRCmd fires once
	// per TA parameter the walk visits, with the parameter's full bytes;
	// OnPVRPixel reports each fragment, drawn or rejected, in framebuffer
	// pixels — the same plane RenderDrawTarget decodes. OnRender fires after
	// a render completes, picture still in the write framebuffer.
	//
	// RenderStopAfter, when non-zero, stops the render walk after that many
	// commands (the clear is command one) — the scrubber's mid-frame halt.
	OnPVRClear      func(w, h int)
	OnPVRCmd        func(param []byte)
	OnPVRPixel      func(x, y int, r, g, b, a uint8, drawn bool)
	OnRender        func()
	RenderStopAfter int

	// Watches: with OnWatch set, accesses inside the ranges report. Physical
	// addresses, like everything on this bus.
	WatchW, WatchR []WatchRange
	OnWatch        func(write bool, addr, v uint32, size int, pc uint32)

	gaps map[string]int
}

// NewMachine builds a Dreamcast around an opened disc. The disc may be nil
// for a machine that only ever runs planted code (tests do).
func NewMachine(disc *Disc) *Machine {
	m := &Machine{
		RAM:     make([]byte, RAMSize),
		VRAM:    make([]byte, VRAMSize),
		AICARAM: make([]byte, AICARAMSize),
		Flash:   make([]byte, FlashSize),
		Disc:    disc,
		Pad:     PadState{Buttons: 0xFFFF, JoyX: 0x80, JoyY: 0x80},
		TAList:  -1,
		TAVtx:   32,
		gaps:    map[string]int{},
	}
	// Erased flash reads FF — the true state of a console that has never
	// saved settings, not an invention.
	for i := range m.Flash {
		m.Flash[i] = 0xFF
	}
	m.CPU = sh4.NewCPU(m)
	m.ARM = arm.NewCPU(armBus{m})
	m.AICARegs = map[uint32]uint32{}
	return m
}

// logf records a gap once per distinct message, optionally echoing the first
// occurrence to Verbose.
func (m *Machine) logf(format string, args ...interface{}) {
	key := fmt.Sprintf(format, args...)
	m.gaps[key]++
	if m.gaps[key] == 1 && m.Verbose != nil {
		fmt.Fprintln(m.Verbose, "dc:", key)
	}
}

// Census lists everything unmodelled that the run touched — the machine's
// gaps, then the CPU's on-chip ones — most-hit first is not promised; the
// list is the worklist.
func (m *Machine) Census() []string {
	var out []string
	for k, n := range m.gaps {
		out = append(out, fmt.Sprintf("%s x%d", k, n))
	}
	out = append(out, m.CPU.Gaps()...)
	if m.bios != nil {
		out = append(out, m.bios.census()...)
	}
	return out
}

// backing resolves a physical address to plain memory, or nil for register
// space. The VRAM array is kept in the 64-bit path's layout — the one texture
// control words address — so the 64-bit window is direct and the 32-bit
// window shuffles through the banks' word interleave.
func (m *Machine) backing(addr uint32) ([]byte, uint32) {
	switch {
	case addr>>26 == 3: // 0x0C000000-0x0FFFFFFF: RAM and its mirrors
		return m.RAM, addr & (RAMSize - 1)
	case addr >= vram32 && addr < vram32+VRAMSize:
		return m.VRAM, vram32to64(addr - vram32)
	case addr >= vram64 && addr < vram64+VRAMSize:
		return m.VRAM, addr - vram64
	case addr >= aicaRAM && addr < aicaRAM+AICARAMSize:
		return m.AICARAM, addr - aicaRAM
	}
	return nil, 0
}

// vram32to64 maps a 32-bit-path VRAM offset to its byte in the 64-bit-path
// layout the VRAM array uses. The 8MB are two 4MB banks; the 64-bit window
// interleaves them a 32-bit word at a time (bank 0 in each group's low word,
// bank 1 in the high), so a byte's home is its bank word's slot in the
// group. Bytes within a word stay adjacent, which is what lets backing()
// hand out a direct slice for any aligned access.
func vram32to64(off uint32) uint32 {
	bank := off >> 22 & 1
	i := off & 0x3FFFFF
	return i>>2<<3 | bank<<2 | i&3
}

func (m *Machine) Fetch16(addr uint32) uint16 {
	if b, off := m.backing(addr); b != nil {
		return uint16(b[off]) | uint16(b[off+1])<<8
	}
	// Fetching outside plain memory is how the BIOS HLE traps work (boot.go
	// parks the syscall vectors on reserved addresses), and otherwise a bug.
	return m.Read16(addr)
}

// WatchRange is a [Start, Start+Len) physical window.
type WatchRange struct {
	Start, Len uint32
}

func inRanges(rs []WatchRange, addr uint32, size int) bool {
	for _, r := range rs {
		if addr+uint32(size) > r.Start && addr < r.Start+r.Len {
			return true
		}
	}
	return false
}

func (m *Machine) watch(write bool, addr, v uint32, size int) {
	rs := m.WatchR
	if write {
		rs = m.WatchW
	}
	if inRanges(rs, addr, size) {
		m.OnWatch(write, addr, v, size, m.CPU.CurPC())
	}
}

func (m *Machine) Read8(addr uint32) uint8 {
	v := m.read8i(addr)
	if m.OnWatch != nil {
		m.watch(false, addr, uint32(v), 1)
	}
	return v
}

func (m *Machine) read8i(addr uint32) uint8 {
	if b, off := m.backing(addr); b != nil {
		return b[off]
	}
	if addr >= flashBase && addr < flashBase+FlashSize {
		return m.Flash[addr-flashBase]
	}
	return uint8(m.ioRead(addr, 1))
}

func (m *Machine) Read16(addr uint32) uint16 {
	v := m.read16i(addr)
	if m.OnWatch != nil {
		m.watch(false, addr, uint32(v), 2)
	}
	return v
}

func (m *Machine) read16i(addr uint32) uint16 {
	if b, off := m.backing(addr); b != nil {
		return uint16(b[off]) | uint16(b[off+1])<<8
	}
	if addr >= flashBase && addr < flashBase+FlashSize {
		off := addr - flashBase
		return uint16(m.Flash[off]) | uint16(m.Flash[off+1])<<8
	}
	return uint16(m.ioRead(addr, 2))
}

func (m *Machine) Read32(addr uint32) uint32 {
	v := m.read32i(addr)
	if m.OnWatch != nil {
		m.watch(false, addr, v, 4)
	}
	return v
}

func (m *Machine) read32i(addr uint32) uint32 {
	if b, off := m.backing(addr); b != nil {
		return uint32(b[off]) | uint32(b[off+1])<<8 | uint32(b[off+2])<<16 | uint32(b[off+3])<<24
	}
	if addr >= flashBase && addr < flashBase+FlashSize {
		off := addr - flashBase
		return uint32(m.Flash[off]) | uint32(m.Flash[off+1])<<8 | uint32(m.Flash[off+2])<<16 | uint32(m.Flash[off+3])<<24
	}
	return m.ioRead(addr, 4)
}

func (m *Machine) Write8(addr uint32, v uint8) {
	if m.OnWatch != nil {
		m.watch(true, addr, uint32(v), 1)
	}
	if b, off := m.backing(addr); b != nil {
		b[off] = v
		return
	}
	m.ioWrite(addr, 1, uint32(v))
}

func (m *Machine) Write16(addr uint32, v uint16) {
	if m.OnWatch != nil {
		m.watch(true, addr, uint32(v), 2)
	}
	if b, off := m.backing(addr); b != nil {
		b[off], b[off+1] = uint8(v), uint8(v>>8)
		return
	}
	m.ioWrite(addr, 2, uint32(v))
}

func (m *Machine) Write32(addr uint32, v uint32) {
	if m.OnWatch != nil {
		m.watch(true, addr, v, 4)
	}
	if b, off := m.backing(addr); b != nil {
		b[off], b[off+1], b[off+2], b[off+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
		return
	}
	m.ioWrite(addr, 4, v)
}

// ioRead serves the register space.
func (m *Machine) ioRead(addr uint32, size int) uint32 {
	switch {
	case addr >= 0x005F6C00 && addr < 0x005F6D00:
		return m.mapleRead(addr)
	case addr >= sbBase && addr < gdBase:
		return m.hollyRead(addr)
	case addr >= gdBase && addr < gdBase+0x100:
		m.logf("GD-ROM ATA read %08X (unmodelled; the syscall HLE is the supported path)", addr)
		return 0
	case addr >= pvrBase && addr < pvrBase+0x2000:
		return m.pvrRead(addr)
	case addr >= aicaBase && addr < aicaBase+0x10000:
		return m.aicaRead(addr-aicaBase, size, true)
	case addr >= 0x00710000 && addr < 0x00710010:
		return m.rtcRead(addr)
	case addr < 0x00200000:
		m.logf("boot ROM read %08X (no BIOS image; syscalls are HLE'd)", addr)
		return 0
	case addr >= 0x01000000 && addr < 0x04000000:
		// The G2 expansion sockets (modem, broadband adapter): nothing is
		// plugged in, and an empty bus reads all-ones — the same "empty
		// socket" answer the Maple bus gives. Zeros here made Crazy Taxi's
		// peripheral probe half-succeed and hang in the driver.
		m.logf("G2 expansion read %08X (empty socket, reads FF)", addr)
		switch size {
		case 1:
			return 0xFF
		case 2:
			return 0xFFFF
		}
		return 0xFFFFFFFF
	case addr >= taBase && addr < taEnd:
		m.logf("TA FIFO read %08X", addr)
		return 0
	}
	m.logf("read%d unmodelled %08X (PC %08X)", size*8, addr, m.CPU.CurPC())
	return 0
}

func (m *Machine) ioWrite(addr uint32, size int, v uint32) {
	switch {
	case addr >= 0x005F6C00 && addr < 0x005F6D00:
		m.mapleWrite(addr, v)
		return
	case addr >= sbBase && addr < gdBase:
		m.hollyWrite(addr, v)
		return
	case addr >= gdBase && addr < gdBase+0x100:
		m.logf("GD-ROM ATA write %08X (unmodelled; the syscall HLE is the supported path)", addr)
		return
	case addr >= pvrBase && addr < pvrBase+0x2000:
		m.pvrWrite(addr, v)
		return
	case addr >= aicaBase && addr < aicaBase+0x10000:
		m.aicaWrite(addr-aicaBase, size, v, true)
		return
	case addr >= 0x00710000 && addr < 0x00710010:
		m.logf("RTC write %08X = %08X", addr, v)
		return
	case addr >= taBase && addr < taEnd:
		m.TAWrites++
		switch {
		case size != 4:
			m.logf("TA FIFO write size %d unimplemented", size*8)
		case addr < taBase+0x800000: // polygon path: parameters, word by word
			m.taFifoWrite(v)
		case addr < taBase+0x1000000:
			m.logf("TA YUV-converter FIFO write unimplemented")
		default:
			m.logf("TA texture-path FIFO write unimplemented")
		}
		return
	case addr >= flashBase && addr < flashBase+FlashSize:
		m.logf("flash write %08X (read-only stub)", addr)
		return
	}
	m.logf("write%d unmodelled %08X = %08X (PC %08X)", size*8, addr, v, m.CPU.CurPC())
}

// pvrRead serves the PVR register file: the few registers with live meaning
// are computed, the rest round-trip raw so the game's own configuration is
// preserved for video.go and the eventual renderer.
func (m *Machine) pvrRead(addr uint32) uint32 {
	switch addr {
	case pvrBase + 0x00: // PVR core ID
		return 0x17FD11DB
	case pvrBase + 0x04: // revision
		return 0x00000011
	case pvrBase + 0x10C: // SPG_STATUS: scanline + vsync, derived from field phase
		return m.spgStatus()
	}
	return m.PVRRegs[(addr-pvrBase)/4]
}

func (m *Machine) pvrWrite(addr, v uint32) {
	m.PVRRegs[(addr-pvrBase)/4] = v
	// TA_LIST_INIT and soft reset put the TA's current list at OPAQUE, not
	// "none": Crazy Taxi's frame stream opens with a bare END_OF_LIST right
	// after LIST_INIT and its interrupt bookkeeping requires that EOL to
	// close the opaque list — the game was built against hardware, so the
	// hardware resets to list 0.
	if addr == pvrBase+0x144 && v&0x80000000 != 0 {
		m.TAList, m.TAVtx, m.TANeed = 0, 32, 0
		m.TAFifo = m.TAFifo[:0]
		m.TAClosed = append(m.TAClosed[:0], m.TAFrame...)
		m.TAFrame = m.TAFrame[:0]
	}
	if addr == pvrBase+0x08 && v&1 != 0 {
		m.TAList, m.TAVtx, m.TANeed = 0, 32, 0
		m.TAFifo = m.TAFifo[:0]
	}
	if addr == pvrBase+0x14 {
		// STARTRENDER: no rasteriser yet, so nothing is drawn — but the
		// completion is deferred a few scanlines' worth of instructions,
		// because an interrupt raised inside this very store would land
		// before the game's render-kick returns and arms its wait flag.
		m.RenderCountdown = fieldInstructions / 8
	}
}
