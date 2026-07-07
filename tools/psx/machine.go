package psx

// machine.go is the PlayStation oracle: a MIPS R3000 + GTE (tools/mips) wired to
// the PSX memory map, the hardware I/O registers, and a high-level-emulated BIOS
// (bios.go). It boots a PS-X EXE and runs it while exposing the tracing and
// profiling instrumentation the other machine models in this repo provide
// (tools/dos, tools/nds) so we can watch the game produce its data.
//
// The BIOS is HLE: instead of a firmware image, the A0/B0/C0 call vectors are
// intercepted at their entry addresses and serviced in Go (see bios.go), the way
// tools/dos reimplements INT 21h. The GPU/DMA are stubbed just enough that the
// game's init code does not stall; rendering is a later milestone.

import (
	"fmt"

	"retroreverse.com/tools/mips"
)

const (
	ramSize     = 2 * 1024 * 1024 // 2 MiB main RAM
	scratchSize = 1024            // 1 KiB scratchpad (D-cache)
	scratchBase = 0x1F800000
	ioBase      = 0x1F801000
	ioEnd       = 0x1F803000
	biosBase    = 0x1FC00000
	biosEnd     = 0x1FC80000

	iStat = 0x1F801070
	iMask = 0x1F801074

	// stepsPerVBlank paces the synthetic vertical-blank IRQ. One NTSC field is
	// ~564,480 CPU cycles (33.8688 MHz / 60 Hz); with Step ≈ one cycle this is a
	// close-enough cadence to drive the game's polled I_STAT timing. Only the rate
	// of in-game time depends on it, not the correctness of the mechanism.
	stepsPerVBlank = 564480

	// isrReturn is a sentinel return address for the ISR trampoline: when a
	// vectored interrupt dispatches to a game handler, $ra is set here so the run
	// loop can catch the handler's `jr $ra`, restore context and resume (see
	// bios.go). It sits in the low BIOS-reserved page, which real code never runs.
	isrReturn = 0x000000E0
)

// Machine is the PSX oracle.
type Machine struct {
	ram     []byte
	scratch []byte
	CPU     *mips.CPU
	GTE     *mips.GTE

	io       map[uint32]uint32 // last-written 32-bit I/O registers
	irqStat  uint32
	irqMask  uint32
	gpuFrame uint32 // toggled into GPUSTAT bit 31 so status polls terminate
	timer    uint32 // free-running value returned for timer reads

	// Interrupt delivery (see run.go / bios.go). The PSX raises IRQs into I_STAT;
	// the game reads them either by polling I_STAT directly (Ridge Racer's CD/timing
	// path) or, once it enables interrupts, by taking a vectored interrupt that the
	// BIOS handler dispatches to a handler the game registered with HookEntryInt.
	vblankAcc uint64 // steps since the last synthetic VBlank
	isrChain  uint32 // HookEntryInt argument: &{next, handler, ...}
	isr       isrState

	// BIOS-HLE bookkeeping.
	biosCalls        map[string]int
	tty              []byte // characters written via the BIOS putchar/std_out
	heapPtr, heapEnd uint32 // bump heap for malloc/InitHeap
	nextEvent        uint32 // OpenEvent handle counter

	// Diagnostics.
	Log     []string
	logSeen map[string]bool

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi uint32                        // "who wrote X" window (inclusive lo, exclusive hi)
	OnWrite          func(addr, val, pc uint32)    // called for writes in the watch window
	OnStep           func(m *Machine, pc uint32)   // called before each instruction

	Halted     bool
	HaltReason string
}

// NewMachine builds a reset machine with RAM, scratchpad, CPU and GTE.
func NewMachine() *Machine {
	m := &Machine{
		ram:       make([]byte, ramSize),
		scratch:   make([]byte, scratchSize),
		io:        map[uint32]uint32{},
		biosCalls: map[string]int{},
		logSeen:   map[string]bool{},
	}
	m.CPU = mips.NewCPU(m)
	m.GTE = mips.NewGTE()
	m.CPU.GTE = m.GTE
	return m
}

// LoadEXE copies a parsed PS-X EXE into RAM and seeds the entry state the BIOS
// would hand the program (PC, gp, sp, fp).
func (m *Machine) LoadEXE(e *EXE) {
	base := e.TAddr & 0x1FFFFF // physical RAM offset
	copy(m.ram[base:base+e.TSize], e.Text)
	m.CPU.SetPC(e.PC0)
	m.CPU.SetReg(28, e.GP0)         // $gp
	sp := e.InitialSP()
	m.CPU.SetReg(29, sp)            // $sp
	m.CPU.SetReg(30, sp)            // $fp
	m.CPU.SetReg(4, 1)              // $a0 = argc (BIOS convention)
	m.CPU.SetReg(5, 0)              // $a1 = argv
}

// --- mips.Bus --------------------------------------------------------------

// phys maps a virtual address to a physical one, folding the KSEG0/KSEG1
// mirrors; RAM is additionally mirrored every 2 MiB across the low region.
func phys(addr uint32) uint32 { return addr & 0x1FFFFFFF }

func (m *Machine) Read(addr uint32) byte {
	a := phys(addr)
	switch {
	case a < 0x00800000:
		return m.ram[a&0x1FFFFF]
	case a >= scratchBase && a < scratchBase+scratchSize:
		return m.scratch[a-scratchBase]
	case a >= ioBase && a < ioEnd:
		base := a &^ 3
		return byte(m.ioReadWord(base) >> ((a & 3) * 8))
	case a >= biosBase && a < biosEnd:
		return 0 // no BIOS image under HLE
	case a >= 0x1F000000 && a < scratchBase:
		return 0xFF // expansion region 1: unpopulated
	default:
		m.note(fmt.Sprintf("read unmapped 0x%08X", addr))
		return 0
	}
}

func (m *Machine) Write(addr uint32, v byte) {
	a := phys(addr)
	if addr == 0xFFFE0130 { // cache control register (KSEG2)
		return
	}
	switch {
	case a < 0x00800000:
		off := a & 0x1FFFFF
		if m.OnWrite != nil && a >= phys(m.WatchLo) && a < phys(m.WatchHi) {
			m.OnWrite(a, uint32(v), m.CPU.CurPC())
		}
		m.ram[off] = v
	case a >= scratchBase && a < scratchBase+scratchSize:
		m.scratch[a-scratchBase] = v
	case a >= ioBase && a < ioEnd:
		base := a &^ 3
		shift := (a & 3) * 8
		m.io[base] = (m.io[base] &^ (0xFF << shift)) | uint32(v)<<shift
		if a&3 == 3 { // register completed by this byte (32-bit stores)
			m.ioSideEffect(base, m.io[base])
		}
	case a >= biosBase && a < biosEnd:
		// ROM: ignore writes.
	default:
		m.note(fmt.Sprintf("write unmapped 0x%08X = 0x%02X", addr, v))
	}
}

// ioReadWord returns the current value of a 32-bit hardware register.
func (m *Machine) ioReadWord(base uint32) uint32 {
	switch base {
	case iStat:
		return m.irqStat
	case iMask:
		return m.irqMask
	case 0x1F801814: // GPUSTAT: report ready, toggling bit 31 so status polls end
		m.gpuFrame ^= 0x80000000
		return 0x1C000000 | m.gpuFrame
	case 0x1F801810: // GPUREAD
		return 0
	case 0x1F801100, 0x1F801110, 0x1F801120: // timer current values
		m.timer += 0x100
		return m.timer & 0xFFFF
	default:
		return m.io[base]
	}
}

// ioSideEffect applies the effect of a completed 32-bit register write.
func (m *Machine) ioSideEffect(base, word uint32) {
	switch {
	case base == iStat:
		// Writing acknowledges (clears) the interrupt bits that are zero.
		m.irqStat &= word
	case base == iMask:
		m.irqMask = word
	case base >= 0x1F801080 && base < 0x1F801100:
		// DMA channel registers. On a CHCR (offset +8) write that starts a
		// transfer, immediately mark it complete so the game does not wait.
		if base&0xF == 0x8 && word&0x01000000 != 0 {
			m.io[base] = word &^ 0x01000000 // clear the busy/start bit
			ch := (base - 0x1F801080) >> 4
			m.raiseIRQ(3)            // DMA interrupt line
			m.io[0x1F8010F4] |= 1 << (24 + ch) // DICR channel-done flag
		}
	}
}

// raiseIRQ sets an interrupt-request bit in I_STAT (0..10).
func (m *Machine) raiseIRQ(bit uint) { m.irqStat |= 1 << bit }

// read32 assembles a little-endian word through the normal memory map (used by
// the BIOS-HLE for structures the game hands us, e.g. the HookEntryInt chain).
func (m *Machine) read32(a uint32) uint32 {
	return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 | uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
}

// write32 stores a little-endian word through the normal memory map.
func (m *Machine) write32(a, v uint32) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
	m.Write(a+2, byte(v>>16))
	m.Write(a+3, byte(v>>24))
}

// note logs a distinct diagnostic message once.
func (m *Machine) note(msg string) {
	if m.logSeen[msg] {
		return
	}
	m.logSeen[msg] = true
	m.Log = append(m.Log, msg)
}

// DisasmAt returns the disassembly text of the instruction at pc (for tracing).
func (m *Machine) DisasmAt(pc uint32) string {
	var b [4]byte
	for i := uint32(0); i < 4; i++ {
		b[i] = m.Read(pc + i)
	}
	return mips.Decode(b[:], pc).Text
}

// TTY returns the text the program printed through the BIOS.
func (m *Machine) TTY() string { return string(m.tty) }

// BiosCalls returns the count of each serviced BIOS call, for a run summary.
func (m *Machine) BiosCalls() map[string]int { return m.biosCalls }
