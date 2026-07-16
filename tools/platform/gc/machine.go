package gc

// machine.go is the GameCube: a Gekko (tools/cpu/gekko) wired to 24 MiB of main memory,
// 16 MiB of auxiliary RAM, the hardware register blocks, and a disc drive — and it boots
// the retail disc and runs it.
//
// The bus is the whole of the machine's structure. The Gekko reads and writes physical
// addresses (it has already translated the virtual ones through its block-address
// registers), and every one of them is either main memory, a hardware register, or
// nothing. The register blocks are small and mostly their own files; what lives here is
// the dispatch, the memory, and the diagnostic that logs — once — any register the model
// does not yet know, so a gap announces itself rather than reading back a plausible zero.

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/gekko"
)

// The physical memory map. Virtual 0x80000000 (cached) and 0xC0000000 (uncached) both
// translate to physical 0, so these are the addresses the bus actually sees.
const (
	RAMSize  = 24 * 1024 * 1024 // "Splash", the 1T-SRAM main memory
	ARAMSize = 16 * 1024 * 1024 // the auxiliary RAM behind the DSP's DMA engine

	// The hardware register blocks all sit in one 64 KiB page at physical 0x0C000000,
	// which software reaches through the uncached virtual alias 0xCC000000.
	hwBase = 0x0C000000
	hwEnd  = 0x0C010000

	regCP  = 0x0C000000 // command processor: the graphics FIFO's control
	regPE  = 0x0C001000 // pixel engine: the end of the graphics pipe, and its interrupts
	regVI  = 0x0C002000 // video interface: the scanout, and the retrace that is the heartbeat
	regPI  = 0x0C003000 // processor interface: the interrupt controller and the FIFO pointers
	regMI  = 0x0C004000 // memory interface: the protection regions
	regDSP = 0x0C005000 // the DSP mailboxes, the AI DMA, and the ARAM DMA all share this block
	regDI  = 0x0C006000 // disc interface
	regSI  = 0x0C006400 // serial interface: the controllers
	regEXI = 0x0C006800 // external interface: the memory card, the RTC and SRAM
	regAIS = 0x0C006C00 // audio interface: the streaming sample clock

	// The write-gather pipe: the Gekko gathers stores to this one address into cache
	// lines and bursts them into the graphics FIFO. Physical 0x0C008000.
	regWGPipe = 0x0C008000
)

// Machine is the GameCube.
type Machine struct {
	RAM  []byte // 24 MiB, physical 0x00000000
	ARAM []byte // 16 MiB, reachable only through the DSP's DMA engine

	CPU *gekko.CPU

	disc    *Disc
	discMD5 string // pins a savestate to the disc it was taken on

	// The device register blocks. Each is small; each is its own file.
	pi     pi
	mi     mi
	vi     vi
	di     di
	si     si
	exi    exi
	ai     ai
	dsp    dsp
	cp     cp
	pe     pe
	gpu    gpu
	wgFIFO wgPipe

	// Diagnostics. Each distinct message is logged once, so a register touched in a loop
	// does not flood the trace.
	Log     []string
	logSeen map[string]bool

	// Instrumentation (opt-in; checked on the hot paths).
	OnStep             func(m *Machine, pc uint32)
	WatchLo, WatchHi   uint32
	OnWrite            func(addr, val, pc uint32)
	RWatchLo, RWatchHi uint32
	OnRead             func(addr, val, pc uint32)
	OnDVDRead          func(discOffset int64, length uint32, memAddr uint32) // every disc read, for the FST log
	OnDisplay          func(m *Machine)                                      // once per video field
	OnFIFO             func(data []byte)                                     // graphics FIFO bytes, for the GPU
	AIDTap             func(block []byte)                                    // every 32-byte audio-DMA block as it drains — the DAC's ear

	// The frame debugger's three hooks. They are the machine's, not the debugger's, because
	// only the machine knows where a command begins, where a pixel is decided, and when a
	// frame is finished — and each is a place the graphics pipe already passes through.
	//
	// OnGXCmd fires once per command the FIFO interpreter consumes, before it executes, so a
	// capture can number the commands in execution order and a stop after command k lands
	// between two of them.
	//
	// OnPixel fires for every pixel the rasteriser decides, drawn or rejected — that is the
	// point of it: a pixel the depth test threw away is provenance without a store, which is
	// exactly what "why is this pixel not what I expect" needs.
	//
	// OnFlip fires when the pixel engine copies the embedded framebuffer out to the external
	// one — the flip, and the only honest frame boundary here. It is NOT the video field: a
	// field is a scanout clock that ticks whether or not the game drew, and the copy that
	// ends a frame usually CLEARS the EFB behind it, so by the time the field boundary comes
	// round the frame's own draw target is already wiped. The hook therefore runs with the
	// EFB still holding the finished frame.
	OnGXCmd func(m *Machine, op uint8, words []uint32)
	OnPixel func(x, y int, ev PixelEvent)
	OnFlip  func(m *Machine)

	StopRequested bool

	// SingleThreaded forces every stage of the machine onto this goroutine. The parallel
	// stages are deterministic by construction — the partition decides the answer, not the
	// scheduler — but a caller that wants to be certain can say so: a bisect against a
	// suspected race, a debugger capture, or a replayer that is already one of eight and
	// should not ask for eight more of its own.
	SingleThreaded bool

	// Instrs counts Gekko instructions retired since the machine was built. It is the
	// emulator's own tally, not guest state: it is not in the savestate, and a restore does
	// not rewind it — a profile counter wants to know how much work this process has done.
	Instrs uint64

	// Profile turns on the per-subsystem frame timing in profile.go. Off, every boundary
	// there costs one predictable branch.
	Profile bool
	prof    profState

	// The command scrubber's counter. Deliberately not machine state and not in the
	// savestate: a replay restores a frame's start snapshot and counts this frame's
	// commands from zero, so carrying a count across a restore would be wrong.
	gxCmdCount  int
	gxStopAfter int  // 0 = run the FIFO normally
	gxStopped   bool // the interpreter has declined the next command and the run is unwinding

	// gxTotalCmds is the lifetime FIFO command count, which the profiler takes deltas of.
	// Distinct from gxCmdCount, which a replay resets.
	gxTotalCmds int

	run    runState
	noSpin bool
}

// PixelEvent is one pixel the rasteriser produced: the colour the TEV made of it, and
// whether it was stored. A rejected pixel (depth or alpha) carries the colour it would have
// had, which is what makes an overdraw history worth reading.
type PixelEvent struct {
	R, G, B, A uint8
	Drawn      bool
}

// NewMachine builds a GameCube around a disc, in the state the IPL leaves it in just
// before it hands control to the disc's apploader. See ipl.go for what "the state the IPL
// leaves it in" means.
func NewMachine(disc *Disc) (*Machine, error) {
	m := &Machine{
		RAM:     make([]byte, RAMSize),
		ARAM:    make([]byte, ARAMSize),
		disc:    disc,
		logSeen: map[string]bool{},
	}
	if disc != nil {
		md5, err := disc.MD5()
		if err != nil {
			return nil, err
		}
		m.discMD5 = md5
	}
	m.CPU = gekko.NewCPU(m)
	m.di.init()
	m.dsp.init()
	m.exi.init()
	m.vi.init()
	// A standard controller is plugged into port 1 (channel 0), the way a console boots with
	// a pad attached. The oracle's -keys injection presses its buttons.
	m.si.connectPad(0)
	return m, nil
}

// Disc is the mounted image, so the oracle can name the files a read lands in.
func (m *Machine) Disc() *Disc { return m.disc }

// --- The bus ------------------------------------------------------------------------
//
// A physical address is main memory, a hardware register, or unmapped. Main memory is the
// common case and is checked first; the register blocks dispatch by their 64 KiB page;
// anything else is logged once and reads back zero, because on this machine an unmapped
// access is a bug in the model and it should be visible rather than silent.
//
// The words go through binary.BigEndian rather than being assembled a byte at a time, which
// is a performance change and not a modelling one: identical bytes in identical order, and
// TestBusEndianMatchesByteAssembly fuzzes it against the assembly it replaced to say so.
// It earns its place because THIS IS THE HOTTEST PATH IN THE MACHINE — Write32 alone was
// 14% of a boot stretch's samples, since the game feeds the graphics FIFO through ordinary
// stores, so every triangle it draws arrives through here. Four bounds checks and six
// shifts become one check and one byte-reversed load on arm64.

func (m *Machine) Read8(a uint32) uint8 {
	if a < RAMSize {
		v := m.RAM[a]
		m.readWatch(a, uint32(v))
		return v
	}
	if a >= hwBase && a < hwEnd {
		return uint8(m.regRead(a, 1))
	}
	m.logf("read8 unmapped 0x%08X", a)
	return 0
}

func (m *Machine) Read16(a uint32) uint16 {
	if a+1 < RAMSize {
		v := binary.BigEndian.Uint16(m.RAM[a : a+2])
		m.readWatch(a, uint32(v))
		return v
	}
	if a >= hwBase && a < hwEnd {
		return uint16(m.regRead(a, 2))
	}
	m.logf("read16 unmapped 0x%08X", a)
	return 0
}

func (m *Machine) Read32(a uint32) uint32 {
	if a+3 < RAMSize {
		v := binary.BigEndian.Uint32(m.RAM[a : a+4])
		m.readWatch(a, v)
		return v
	}
	if a >= hwBase && a < hwEnd {
		return m.regRead(a, 4)
	}
	m.logf("read32 unmapped 0x%08X (PC 0x%08X)", a, m.CPU.PC)
	return 0
}

func (m *Machine) Write8(a uint32, v uint8) {
	if a < RAMSize {
		m.writeWatch(a, uint32(v))
		m.RAM[a] = v
		return
	}
	if a >= regWGPipe && a < regWGPipe+0x20 {
		m.wgFIFO.write8(m, v)
		return
	}
	if a >= hwBase && a < hwEnd {
		m.regWrite(a, uint32(v), 1)
		return
	}
	m.logf("write8 unmapped 0x%08X = 0x%02X", a, v)
}

func (m *Machine) Write16(a uint32, v uint16) {
	if a+1 < RAMSize {
		m.writeWatch(a, uint32(v))
		binary.BigEndian.PutUint16(m.RAM[a:a+2], v)
		return
	}
	if a >= regWGPipe && a < regWGPipe+0x20 {
		m.wgFIFO.write16(m, v)
		return
	}
	if a >= hwBase && a < hwEnd {
		m.regWrite(a, uint32(v), 2)
		return
	}
	m.logf("write16 unmapped 0x%08X = 0x%04X", a, v)
}

func (m *Machine) Write32(a uint32, v uint32) {
	if a+3 < RAMSize {
		m.writeWatch(a, v)
		binary.BigEndian.PutUint32(m.RAM[a:a+4], v)
		return
	}
	if a >= regWGPipe && a < regWGPipe+0x20 {
		m.wgFIFO.write32(m, v)
		return
	}
	if a >= hwBase && a < hwEnd {
		m.regWrite(a, v, 4)
		return
	}
	m.logf("write32 unmapped 0x%08X = 0x%08X (PC 0x%08X)", a, v, m.CPU.PC)
}

// Fetch32 is an instruction fetch. It reads from main memory directly, because code runs
// from memory and a fetch from a register is a program that has already gone wrong.
func (m *Machine) Fetch32(a uint32) uint32 {
	if a+3 < RAMSize {
		return binary.BigEndian.Uint32(m.RAM[a : a+4])
	}
	m.logf("fetch from unmapped 0x%08X", a)
	return 0
}

// regRead and regWrite dispatch a hardware-register access to its block by the page it
// falls in. The blocks answer for themselves; a page with no owner logs once.
func (m *Machine) regRead(a uint32, size int) uint32 {
	off := a & 0xFFFF
	switch {
	case off >= (regCP&0xFFFF) && off < (regPE&0xFFFF):
		return m.cp.read(m, off, size)
	case off >= (regPE&0xFFFF) && off < (regVI&0xFFFF):
		return m.pe.read(m, off, size)
	case off >= (regVI&0xFFFF) && off < (regPI&0xFFFF):
		return m.vi.read(m, off, size)
	case off >= (regPI&0xFFFF) && off < (regMI&0xFFFF):
		return m.pi.read(m, off, size)
	case off >= (regMI&0xFFFF) && off < (regDSP&0xFFFF):
		return m.mi.read(m, off, size)
	case off >= (regDSP&0xFFFF) && off < (regDI&0xFFFF):
		return m.dsp.read(m, off, size)
	case off >= (regDI&0xFFFF) && off < (regSI&0xFFFF):
		return m.di.read(m, off, size)
	case off >= (regSI&0xFFFF) && off < (regEXI&0xFFFF):
		return m.si.read(m, off, size)
	case off >= (regEXI&0xFFFF) && off < (regAIS&0xFFFF):
		return m.exi.read(m, off, size)
	case off >= (regAIS&0xFFFF) && off < 0x7000:
		return m.ai.read(m, off, size)
	}
	m.logf("read%d unmodelled register 0x%08X (PC 0x%08X)", size*8, a, m.CPU.PC)
	return 0
}

func (m *Machine) regWrite(a uint32, v uint32, size int) {
	off := a & 0xFFFF
	switch {
	case off >= (regCP&0xFFFF) && off < (regPE&0xFFFF):
		m.cp.write(m, off, v, size)
	case off >= (regPE&0xFFFF) && off < (regVI&0xFFFF):
		m.pe.write(m, off, v, size)
	case off >= (regVI&0xFFFF) && off < (regPI&0xFFFF):
		m.vi.write(m, off, v, size)
	case off >= (regPI&0xFFFF) && off < (regMI&0xFFFF):
		m.pi.write(m, off, v, size)
	case off >= (regMI&0xFFFF) && off < (regDSP&0xFFFF):
		m.mi.write(m, off, v, size)
	case off >= (regDSP&0xFFFF) && off < (regDI&0xFFFF):
		m.dsp.write(m, off, v, size)
	case off >= (regDI&0xFFFF) && off < (regSI&0xFFFF):
		m.di.write(m, off, v, size)
	case off >= (regSI&0xFFFF) && off < (regEXI&0xFFFF):
		m.si.write(m, off, v, size)
	case off >= (regEXI&0xFFFF) && off < (regAIS&0xFFFF):
		m.exi.write(m, off, v, size)
	case off >= (regAIS&0xFFFF) && off < 0x7000:
		m.ai.write(m, off, v, size)
	default:
		m.logf("write%d unmodelled register 0x%08X = 0x%08X (PC 0x%08X)", size*8, a, v, m.CPU.PC)
	}
}

// --- Direct memory access for the machine's own use ---------------------------------

// dmaToRAM copies bytes into main memory. It strips the address to its physical offset, so
// a caller may hand it a cached or uncached virtual address indifferently, and it masks to
// the RAM size rather than failing, because a device given a wild address is a bug we want
// to see land somewhere inspectable.
func (m *Machine) dmaToRAM(addr uint32, data []byte) {
	addr = phys(addr)
	for i, b := range data {
		if int(addr)+i < len(m.RAM) {
			m.RAM[addr+uint32(i)] = b
		}
	}
}

// ram32 and setRAM32 are big-endian word access to main memory, for the machine's own
// bookkeeping (the IPL globals, the register blocks that keep pointers in RAM). They take
// an address in any of memory's aliases — physical, cached 0x80000000, uncached
// 0xC0000000 — and strip it to the physical offset, so the machine's own reads see the
// same bytes the CPU's translated stores wrote. Getting this wrong reads back zero from an
// address the game populated, which is exactly the bug that made the apploader's function
// pointers appear unwritten.
func phys(a uint32) uint32 { return a & 0x03FFFFFF }

func (m *Machine) ram32(a uint32) uint32 {
	a = phys(a)
	if a+3 >= RAMSize {
		return 0
	}
	return binary.BigEndian.Uint32(m.RAM[a : a+4])
}

func (m *Machine) setRAM32(a, v uint32) {
	a = phys(a)
	if a+3 >= RAMSize {
		return
	}
	binary.BigEndian.PutUint32(m.RAM[a:a+4], v)
}

// --- Watches and the log ------------------------------------------------------------

func (m *Machine) readWatch(a, v uint32) {
	if m.OnRead != nil && a >= m.RWatchLo && a < m.RWatchHi {
		m.OnRead(a, v, m.CPU.PC)
	}
}

func (m *Machine) writeWatch(a, v uint32) {
	if m.OnWrite != nil && a >= m.WatchLo && a < m.WatchHi {
		m.OnWrite(a, v, m.CPU.PC)
	}
}

// logf records a diagnostic once. The GameCube has no operating system to service, so
// every one of these is a piece of hardware the game touched that the model does not yet
// answer for — a work list, printed at the end of a run.
func (m *Machine) logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if m.logSeen[s] {
		return
	}
	m.logSeen[s] = true
	m.Log = append(m.Log, s)
}

// Census returns the unmodelled-hardware log — what the game asked for that the machine
// could not answer. An empty census after a boot is the goal.
func (m *Machine) Census() []string { return m.Log }

// GPUCensus returns how many of each opcode the graphics FIFO has carried — the quick
// answer to "is the game drawing at all, and with what".
func (m *Machine) GPUCensus() [256]uint64 { return m.gpu.Census }
