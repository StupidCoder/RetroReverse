package n64

// machine.go is the Nintendo 64 oracle: a VR4300 (tools/cpu/r4300) wired to
// RDRAM, the RCP's memory-mapped registers, the RSP's DMEM/IMEM, and the
// cartridge. It boots the retail ROM and runs it while exposing the tracing and
// profiling instrumentation the other machine models in this repo provide
// (tools/platform/psx, tools/platform/threedo), so we can watch the game produce
// its data.
//
// Nothing here is high-level-emulated. Unlike the PlayStation, whose BIOS calls
// are serviced in Go, and the 3DO, whose Portfolio folios are, the N64's
// operating system (libultra) ships *inside the cartridge* and runs on the CPU
// as ordinary code. Only the hardware needs modelling. The one exception is the
// PIF boot handoff, which is not on the medium at all — see boot.go.
//
// Addresses here are physical. The CPU translates virtual addresses (the KSEG
// segments and the TLB) before touching the bus.

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/r4300"
	"retroreverse.com/tools/cpu/rsp"
)

// The physical memory map.
const (
	rdramSize = 4 * 1024 * 1024 // 4 MiB; the Expansion Pak's second 4 MiB is absent
	rdramEnd  = 0x00800000      // the address space reserved for RDRAM, populated or not

	rdramRegBase = 0x03F00000 // RDRAM device configuration
	rdramRegEnd  = 0x04000000

	spDMEMBase = 0x04000000
	spIMEMBase = 0x04001000
	spMemEnd   = 0x04002000
	spRegBase  = 0x04040000
	spRegEnd   = 0x04100000

	dpRegBase = 0x04100000
	dpRegEnd  = 0x04300000
	miRegBase = 0x04300000
	miRegEnd  = 0x04400000
	viRegBase = 0x04400000
	viRegEnd  = 0x04500000
	aiRegBase = 0x04500000
	aiRegEnd  = 0x04600000
	piRegBase = 0x04600000
	piRegEnd  = 0x04700000
	riRegBase = 0x04700000
	riRegEnd  = 0x04800000
	siRegBase = 0x04800000
	siRegEnd  = 0x04900000

	cartBase = 0x10000000 // PI domain 1: the cartridge ROM
	cartEnd  = 0x1FC00000

	pifBase = 0x1FC00000
	pifEnd  = 0x1FC00800
)

// spMemSize is the size of each of the RSP's two 4 KiB memories.
const spMemSize = 0x1000

// Machine is the N64 oracle.
type Machine struct {
	RDRAM []byte // 4 MiB, physical 0x00000000
	DMEM  []byte // RSP data memory, physical 0x04000000
	IMEM  []byte // RSP instruction memory, physical 0x04001000
	ROM   []byte // the cartridge, big-endian, physical 0x10000000
	PIF   []byte // PIF RAM, the joybus mailbox

	romMD5 string // pins a save-state to the cartridge it was taken on

	CPU *r4300.CPU
	// RSP is created the first time a task starts. It shares the machine's DMEM
	// and IMEM slices, so a DMA into them is immediately visible to microcode.
	RSP *rsp.CPU

	mi   mi  // interrupt aggregation
	pi   pi  // cartridge DMA
	vi   vi  // video, and the retrace interrupt that is the machine's heartbeat
	ai   ai  // audio (deferred; registers only)
	si   si  // serial: the joybus to the controllers and the save EEPROM
	isv  isv // the IS-Viewer development cartridge's print window
	rdp  rdp // the rasteriser's state (rdp.go, rdp_raster.go)
	ri   regFile
	rd   regFile // RDRAM device registers
	sp   regFile
	spPC uint32 // the RSP's program counter, in its own register window
	dp   regFile

	// Controllers holds the four ports. Only port 0 is attached by default; the
	// joybus reports the rest as empty, which is what a real console does.
	Controllers [4]Controller

	// EEPROM is the cartridge's save device, reached over the joybus. Empty means
	// no save device is fitted.
	EEPROM []byte

	run        runState
	noSpin     bool
	rspRunning bool   // guards against a task restarting itself mid-run
	rspSteps   uint64 // RSP instructions executed, across all tasks
	rdpWords   uint64 // RDP command words queued, across all tasks

	// Diagnostics. Each distinct message is logged once, so a register touched in
	// a loop does not flood the trace.
	Log     []string
	logSeen map[string]bool

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi   uint32                      // "who wrote X" window (inclusive lo, exclusive hi)
	OnWrite            func(addr, val, pc uint32)  // called for writes in the watch window
	OnStep             func(m *Machine, pc uint32) // called before each instruction
	RWatchLo, RWatchHi uint32                      // "who read X" window
	OnRead             func(addr, val, pc uint32)  // called for CPU reads in the read-watch window
	// OnDMA is called at the start of each DMA transfer.
	OnDMA func(kind string, dramAddr, cartAddr, length uint32)
	// OnDisplay is called once per video field, after the retrace interrupt is
	// raised — the natural place to capture a frame.
	OnDisplay func(m *Machine)
	// OnRSPTask is called just before a microcode task begins, with its entry
	// point in IMEM.
	OnRSPTask func(m *Machine, pc uint32)
	// OnRDPCmd is called for each decoded RDP command, before it is executed —
	// the counterpart of the PlayStation oracle's OnGP0.
	OnRDPCmd func(m *Machine, op uint32, words []uint64)
	// OnPixel is called for every pixel the triangle rasteriser and the
	// texture-rectangle walker produce, whether or not it reached memory.
	// Pairing it with a command counter kept in OnRDPCmd attributes each
	// pixel to the draw that produced it — the census method every RDP
	// defect so far was found with.
	OnPixel func(x, y uint32, ev PixelEvent)
	// OnPrint is called for each line a program writes to the IS-Viewer. Test
	// ROMs report their results this way.
	OnPrint func(m *Machine, line string)
	// ContPolls counts joybus controller-state reads, and JoybusCmds counts every
	// joybus command by opcode. Instrumentation only: neither is part of the
	// machine state, and neither is saved.
	ContPolls         uint64
	JoybusCmds        map[byte]uint64
	SIWrites, SIReads uint64
	// StopRequested ends the current Run at the next instruction boundary. A
	// hook (say, OnDisplay counting video fields) sets it to stop a run at a
	// condition a step budget cannot express. Not part of the machine state.
	StopRequested bool

	dpWriteHook func(addr, v uint32)
	hookMuted   bool // suppresses hooks during machine-internal reads (DMA, disassembly)
}

// regFile is a device's register block: 32-bit registers, read back as written
// unless the device gives them meaning. Undecoded registers within a device's
// range therefore behave like memory rather than halting the boot, while an
// access to an unmapped *device* is a gap and says so.
type regFile map[uint32]uint32

// NewMachine builds an oracle around a loaded cartridge.
func NewMachine(rom *ROM) *Machine {
	m := &Machine{
		RDRAM:   make([]byte, rdramSize),
		DMEM:    make([]byte, spMemSize),
		IMEM:    make([]byte, spMemSize),
		PIF:     make([]byte, 64),
		ROM:     rom.Data,
		romMD5:  rom.MD5,
		ri:      regFile{},
		rd:      regFile{},
		sp:      regFile{},
		dp:      regFile{},
		EEPROM:  make([]byte, eeprom4KBlocks*eepromBlockSize),
		logSeen: map[string]bool{},
	}
	m.CPU = r4300.NewCPU(m)
	m.pi.init()
	m.vi.init()
	m.ai.init()
	m.si.init()
	m.sp.init(spStatus, spStatusHalt) // the RSP is halted out of reset
	m.Controllers[0].Present = true   // one controller, in port 1
	return m
}

// newBareCPU attaches a CPU to a machine assembled without a cartridge, for the
// rasteriser's unit tests.
func newBareCPU(m *Machine) *r4300.CPU { return r4300.NewCPU(m) }

func (r regFile) init(pairs ...uint32) {
	for i := 0; i+1 < len(pairs); i += 2 {
		r[pairs[i]] = pairs[i+1]
	}
}

// rdramRead and rdramWrite address RDRAM without mirroring.
//
// The N64 has no RAM mirroring: the address space reserves 8 MiB for RDRAM but a
// console without an Expansion Pak populates only the first 4. Reads of the
// unpopulated half return zero and writes vanish. Wrapping the address instead —
// which is the obvious thing to do with a Go slice — makes a DMA from high
// memory quietly copy the low memory over itself. n64-systemtest's own boot code
// relies on the zeroes to clear the RSP's instruction memory.
func (m *Machine) rdramRead(a uint32) byte {
	if int(a) < len(m.RDRAM) {
		return m.RDRAM[a]
	}
	return 0
}

func (m *Machine) rdramWrite(a uint32, v byte) {
	if int(a) < len(m.RDRAM) {
		m.RDRAM[a] = v
	}
}

// note records a diagnostic once.
func (m *Machine) note(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if m.logSeen[s] {
		return
	}
	m.logSeen[s] = true
	m.Log = append(m.Log, s)
}

// pc reports the address of the instruction currently executing, for hook
// attribution.
func (m *Machine) pc() uint32 { return uint32(m.CPU.CurPC()) }

// --- the bus ----------------------------------------------------------------

// backing returns the byte slice that holds addr, if it is plain memory. The
// fast path for RDRAM, the SP memories and the cartridge; MMIO returns nil.
func (m *Machine) backing(addr uint32) ([]byte, uint32) {
	switch {
	case addr < rdramSize:
		return m.RDRAM, addr
	case addr >= spDMEMBase && addr < spIMEMBase:
		return m.DMEM, addr - spDMEMBase
	case addr >= spIMEMBase && addr < spMemEnd:
		return m.IMEM, addr - spIMEMBase
	case inISViewer(addr):
		return nil, 0 // handled by ioRead / ioWrite, ahead of the cartridge
	case addr >= cartBase && addr < cartEnd:
		off := addr - cartBase
		if int(off) < len(m.ROM) {
			return m.ROM, off
		}
	case addr >= pifBase && addr < pifEnd:
		// PIF ROM is not on the cartridge; only its RAM (the last 64 bytes) is
		// modelled, as the joybus mailbox.
		if addr >= pifEnd-64 {
			return m.PIF, addr - (pifEnd - 64)
		}
	}
	return nil, 0
}

// Fetch32 serves an instruction fetch. It deliberately skips the read-watch
// hooks: a watch is asking "what code reads this data", and every instruction
// executed inside the window would otherwise report itself.
func (m *Machine) Fetch32(addr uint32) uint32 {
	if b, off := m.backing(addr &^ 3); b != nil {
		return binary.BigEndian.Uint32(b[off:])
	}
	return m.ioRead(addr &^ 3)
}

// Read32 is the CPU's hot path for every word load.
func (m *Machine) Read32(addr uint32) uint32 {
	addr &^= 3
	var v uint32
	if b, off := m.backing(addr); b != nil {
		v = binary.BigEndian.Uint32(b[off:])
	} else {
		v = m.ioRead(addr)
	}
	if !m.hookMuted && m.OnRead != nil && addr >= m.RWatchLo && addr < m.RWatchHi {
		m.OnRead(addr, v, m.pc())
	}
	return v
}

func (m *Machine) Write32(addr uint32, v uint32) {
	addr &^= 3
	if !m.hookMuted && m.OnWrite != nil && addr >= m.WatchLo && addr < m.WatchHi {
		m.OnWrite(addr, v, m.pc())
	}
	if b, off := m.backing(addr); b != nil {
		// The cartridge is read-only; a write to it is a bug worth hearing about.
		if addr >= cartBase && addr < cartEnd {
			m.note("write 0x%08X to the cartridge at 0x%08X (ignored)", v, addr)
			return
		}
		binary.BigEndian.PutUint32(b[off:], v)
		return
	}
	m.ioWrite(addr, v)
}

// Read and Write handle the sub-word accesses. Memory-mapped registers are
// word-only on this machine, so a byte access to one reads or modifies the
// containing word.
func (m *Machine) Read(addr uint32) byte {
	if b, off := m.backing(addr); b != nil {
		v := b[off]
		if !m.hookMuted && m.OnRead != nil && addr >= m.RWatchLo && addr < m.RWatchHi {
			m.OnRead(addr, uint32(v), m.pc())
		}
		return v
	}
	w := m.Read32(addr &^ 3)
	return byte(w >> (8 * (3 - addr&3)))
}

func (m *Machine) Write(addr uint32, v byte) {
	if !m.hookMuted && m.OnWrite != nil && addr >= m.WatchLo && addr < m.WatchHi {
		m.OnWrite(addr, uint32(v), m.pc())
	}
	if inISViewer(addr) {
		m.isvWriteByte(addr, v) // the print buffer is filled a byte at a time
		return
	}
	if b, off := m.backing(addr); b != nil {
		if addr >= cartBase && addr < cartEnd {
			m.note("write 0x%02X to the cartridge at 0x%08X (ignored)", v, addr)
			return
		}
		b[off] = v
		return
	}
	shift := 8 * (3 - addr&3)
	w := m.Read32(addr&^3)&^(0xFF<<shift) | uint32(v)<<shift
	m.Write32(addr&^3, w)
}

// ioRead dispatches a word read to the device that owns addr.
func (m *Machine) ioRead(addr uint32) uint32 {
	switch {
	case addr < rdramEnd:
		// RDRAM's address space extends to 8 MiB, but without an Expansion Pak
		// the upper half is unpopulated and reads as zero. IPL3 sizes memory by
		// probing exactly here.
		return 0
	case addr >= rdramRegBase && addr < rdramRegEnd:
		return m.rd[addr&0xFF]
	case addr >= spRegBase && addr < spRegEnd:
		return m.spRead(addr)
	case addr >= dpRegBase && addr < dpRegEnd:
		return m.dpRead(addr)
	case addr >= miRegBase && addr < miRegEnd:
		return m.miRead(addr)
	case addr >= viRegBase && addr < viRegEnd:
		return m.viRead(addr)
	case addr >= aiRegBase && addr < aiRegEnd:
		return m.aiRead(addr)
	case addr >= piRegBase && addr < piRegEnd:
		return m.piRead(addr)
	case addr >= riRegBase && addr < riRegEnd:
		return m.ri[addr&0xFF]
	case addr >= siRegBase && addr < siRegEnd:
		return m.siRead(addr)
	case inISViewer(addr):
		return m.isvRead(addr)
	}
	m.note("read from unmapped physical address 0x%08X (returning 0)", addr)
	return 0
}

func (m *Machine) ioWrite(addr uint32, v uint32) {
	switch {
	case addr < rdramEnd:
		return // unpopulated RDRAM: writes vanish
	case addr >= rdramRegBase && addr < rdramRegEnd:
		m.rd[addr&0xFF] = v
	case addr >= spRegBase && addr < spRegEnd:
		m.spWrite(addr, v)
	case addr >= dpRegBase && addr < dpRegEnd:
		m.dpWrite(addr, v)
	case addr >= miRegBase && addr < miRegEnd:
		m.miWrite(addr, v)
	case addr >= viRegBase && addr < viRegEnd:
		m.viWrite(addr, v)
	case addr >= aiRegBase && addr < aiRegEnd:
		m.aiWrite(addr, v)
	case addr >= piRegBase && addr < piRegEnd:
		m.piWrite(addr, v)
	case addr >= riRegBase && addr < riRegEnd:
		m.ri[addr&0xFF] = v
	case addr >= siRegBase && addr < siRegEnd:
		m.siWrite(addr, v)
	case inISViewer(addr):
		m.isvWrite(addr, v)
	default:
		m.note("write 0x%08X to unmapped physical address 0x%08X (ignored)", v, addr)
	}
}

// --- diagnostics ------------------------------------------------------------

// DisasmAt renders the instruction at a virtual address, for the tracer hooks.
func (m *Machine) DisasmAt(vaddr uint64) string {
	p, ok := m.CPU.Translate(vaddr, false)
	if !ok {
		return fmt.Sprintf("$%08X: <untranslatable>", uint32(vaddr))
	}
	muted := m.hookMuted
	m.hookMuted = true
	w := m.Read32(p)
	m.hookMuted = muted
	return r4300.DecodeWord(w, uint32(vaddr)).Text
}

// writePhys32 writes a word to a physical address without firing hooks, for the
// machine's own setup.
func (m *Machine) writePhys32(addr uint32, v uint32) {
	muted := m.hookMuted
	m.hookMuted = true
	m.Write32(addr, v)
	m.hookMuted = muted
}

// ReadVirt reads a word through the CPU's translation, without firing hooks.
func (m *Machine) ReadVirt(vaddr uint64) (uint32, bool) {
	p, ok := m.CPU.Translate(vaddr, false)
	if !ok {
		return 0, false
	}
	muted := m.hookMuted
	m.hookMuted = true
	v := m.Read32(p)
	m.hookMuted = muted
	return v, true
}

// RSPSteps and RDPWords report how much work the coprocessors have been given,
// for the oracle's diagnostics.
func (m *Machine) RSPSteps() uint64 { return m.rspSteps }
func (m *Machine) RDPWords() uint64 { return m.rdpWords }
