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

	Holly Holly

	// The PVR register file, stored raw at word granularity; video.go
	// interprets the scanout set, everything else round-trips.
	PVRRegs [0x2000 / 4]uint32

	// TAWrites counts tile-accelerator FIFO words the machine swallowed —
	// the honest record that geometry was submitted to a renderer that does
	// not exist yet.
	TAWrites uint64

	Instrs       uint64 // instructions retired, the run budget's unit
	Fields       uint64 // VBlank heartbeats delivered
	instrInField uint64

	// Instrumentation, all opt-in and nil-checked on the hot path.
	OnStep    func(pc uint32)
	OnDisplay func(field uint64)
	Verbose   io.Writer // gap-log lines land here when set

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
		gaps:    map[string]int{},
	}
	m.CPU = sh4.NewCPU(m)
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
	return out
}

// backing resolves a physical address to plain memory, or nil for register
// space. The 64-bit VRAM window is served linearly — a once-logged
// simplification that must be revisited with the PVR's texture layout.
func (m *Machine) backing(addr uint32) ([]byte, uint32) {
	switch {
	case addr>>26 == 3: // 0x0C000000-0x0FFFFFFF: RAM and its mirrors
		return m.RAM, addr & (RAMSize - 1)
	case addr >= vram32 && addr < vram32+VRAMSize:
		return m.VRAM, addr - vram32
	case addr >= vram64 && addr < vram64+VRAMSize:
		m.logf("VRAM 64-bit window served linearly")
		return m.VRAM, addr - vram64
	case addr >= aicaRAM && addr < aicaRAM+AICARAMSize:
		return m.AICARAM, addr - aicaRAM
	}
	return nil, 0
}

func (m *Machine) Fetch16(addr uint32) uint16 {
	if b, off := m.backing(addr); b != nil {
		return uint16(b[off]) | uint16(b[off+1])<<8
	}
	// Fetching outside plain memory is how the BIOS HLE traps work (boot.go
	// parks the syscall vectors on reserved addresses), and otherwise a bug.
	return m.Read16(addr)
}

func (m *Machine) Read8(addr uint32) uint8 {
	if b, off := m.backing(addr); b != nil {
		return b[off]
	}
	if addr >= flashBase && addr < flashBase+FlashSize {
		return m.Flash[addr-flashBase]
	}
	return uint8(m.ioRead(addr, 1))
}

func (m *Machine) Read16(addr uint32) uint16 {
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
	if b, off := m.backing(addr); b != nil {
		b[off] = v
		return
	}
	m.ioWrite(addr, 1, uint32(v))
}

func (m *Machine) Write16(addr uint32, v uint16) {
	if b, off := m.backing(addr); b != nil {
		b[off], b[off+1] = uint8(v), uint8(v>>8)
		return
	}
	m.ioWrite(addr, 2, uint32(v))
}

func (m *Machine) Write32(addr uint32, v uint32) {
	if b, off := m.backing(addr); b != nil {
		b[off], b[off+1], b[off+2], b[off+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
		return
	}
	m.ioWrite(addr, 4, v)
}

// ioRead serves the register space.
func (m *Machine) ioRead(addr uint32, size int) uint32 {
	switch {
	case addr >= sbBase && addr < gdBase:
		return m.hollyRead(addr)
	case addr >= gdBase && addr < gdBase+0x100:
		m.logf("GD-ROM ATA read %08X (unmodelled; the syscall HLE is the supported path)", addr)
		return 0
	case addr >= pvrBase && addr < pvrBase+0x2000:
		return m.pvrRead(addr)
	case addr >= aicaBase && addr < aicaBase+0x10000:
		m.logf("AICA register read %08X (ARM7 not modelled)", addr)
		return 0
	case addr < 0x00200000:
		m.logf("boot ROM read %08X (no BIOS image; syscalls are HLE'd)", addr)
		return 0
	case addr >= taBase && addr < taEnd:
		m.logf("TA FIFO read %08X", addr)
		return 0
	}
	m.logf("read%d unmodelled %08X (PC %08X)", size*8, addr, m.CPU.CurPC())
	return 0
}

func (m *Machine) ioWrite(addr uint32, size int, v uint32) {
	switch {
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
		m.logf("AICA register write %08X (ARM7 not modelled)", addr)
		return
	case addr >= taBase && addr < taEnd:
		m.TAWrites++
		if m.TAWrites == 1 {
			m.logf("TA FIFO writes counted, not rendered (the PVR is a later milestone)")
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
}
