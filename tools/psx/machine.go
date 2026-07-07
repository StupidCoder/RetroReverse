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
	cdBase = 0x1F801800 // CD-ROM controller (4 byte-wide, index-banked registers)

	// stepsPerVBlank paces the synthetic vertical-blank IRQ. One NTSC field is
	// ~565,000 CPU cycles (33.8688 MHz / 59.94 Hz), but the oracle counts executed
	// instructions, not cycles — and the R3000 averages well over one cycle per
	// instruction (memory access to un-cached RAM dominates). The game's frame
	// waits count instructions against a VBlank-advanced counter, so the cadence
	// must be scaled by that CPI: ~2 gives roughly a real frame's worth of code.
	stepsPerVBlank = 250000

	// isrStackTop is the top of a dedicated interrupt stack in the low kernel RAM
	// (below the game's text at 0x10000), so a dispatched handler's stack frame
	// never overlaps the interrupted code's stack — the real BIOS switches stacks
	// the same way. Kept clear of ~0x8000F000, which the game uses for a stack of
	// its own.
	isrStackTop = 0x8000D000

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
	cd      *cdrom  // CD-ROM controller (0x1F801800-1803)
	gpu     *gpu    // GPU (0x1F801810-1814)
	disc    *Volume // mounted disc image, the CD sector source

	io       map[uint32]uint32 // last-written 32-bit I/O registers
	irqStat  uint32
	irqMask  uint32
	timer    uint32 // free-running value returned for timer reads
	dmaFlags uint32 // DICR per-channel IRQ flags (bits 0..6), write-1-to-clear

	// Interrupt delivery (see run.go / bios.go). The PSX raises IRQs into I_STAT;
	// the game reads them either by polling I_STAT directly (Ridge Racer's CD/timing
	// path) or, once it enables interrupts, by taking a vectored interrupt that the
	// BIOS handler dispatches to a handler the game registered with HookEntryInt.
	vblankAcc uint64 // steps since the last synthetic VBlank
	isrChain  uint32 // HookEntryInt argument: &{next, handler, ...}
	isr       isrState

	// ISRHandler optionally overrides the vectored-interrupt entry point. The
	// retail BIOS ROM, on HookEntryInt, installs a handler stub at the caller's
	// slot; under HLE (no ROM) that slot stays empty, so a caller that has traced
	// the game's own interrupt dispatcher can point delivery straight at it. Zero
	// means "use the registered chain handler" (the default).
	ISRHandler uint32

	// BIOS-HLE bookkeeping.
	biosCalls        map[string]int
	tty              []byte // characters written via the BIOS putchar/std_out
	heapPtr, heapEnd uint32 // bump heap for malloc/InitHeap
	nextEvent        uint32 // OpenEvent handle counter
	randSeed         uint32 // BIOS rand() state

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
	m.cd = newCDROM(m)
	m.gpu = newGPU()
	return m
}

// SetDisc mounts a disc image so the CD-ROM controller can serve sectors.
func (m *Machine) SetDisc(v *Volume) { m.disc = v }

// Screenshot renders the GPU's active display area to a PNG.
func (m *Machine) Screenshot(path string) error { return m.gpu.Screenshot(path) }

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
	// The BIOS leaves the hardware-interrupt mask (SR IM2, bit 10) enabled before
	// handing control to the game; IEc itself is toggled by critical sections.
	// Without IM2 a raised IP2 (I_STAT&I_MASK) would be masked and never vector.
	m.CPU.COP0[12] |= 1 << 10
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
	case a >= cdBase && a <= cdBase+3:
		return m.cd.read(a - cdBase) // byte-addressed, index-banked
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
	case a >= cdBase && a <= cdBase+3:
		m.cd.write(a-cdBase, v) // byte-addressed, index-banked
	case a >= ioBase && a < ioEnd:
		base := a &^ 3
		shift := (a & 3) * 8
		m.io[base] = (m.io[base] &^ (0xFF << shift)) | uint32(v)<<shift
		// Fire the register side effect when the access completes. Most I/O is
		// written 32-bit (high byte at a&3==3), but the game drives the 16-bit
		// interrupt registers with `sh` (I_MASK enable, I_STAT ack), which never
		// reaches a&3==3 — so also fire I_STAT/I_MASK on their halfword boundary.
		if a&3 == 3 || ((base == iStat || base == iMask) && a&3 == 1) {
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
	case 0x1F801814: // GPUSTAT
		return m.gpu.status()
	case 0x1F801810: // GPUREAD
		return m.gpu.read()
	case 0x1F801100, 0x1F801110, 0x1F801120: // timer current values
		m.timer += 0x100
		return m.timer & 0xFFFF
	case 0x1F8010F4: // DICR: control bits + per-channel flags + master IRQ flag
		v := m.io[base]&0x00FFFFFF | m.dmaFlags<<24
		if v&(1<<15) != 0 || (v&(1<<23) != 0 && m.dmaFlags&((v>>16)&0x7F) != 0) {
			v |= 1 << 31 // master IRQ flag
		}
		return v
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
	case base == 0x1F801810: // GP0 drawing command/data port
		m.gpu.gp0(word)
	case base == 0x1F801814: // GP1 display/control port
		m.gpu.gp1(word)
	case base == 0x1F8010F4: // DICR: bits 24-30 are write-1-to-clear flags
		m.dmaFlags &^= (word >> 24) & 0x7F
	case base >= 0x1F801080 && base < 0x1F801100:
		// DMA channel registers. A CHCR (offset +8) write with the start bit runs
		// the transfer; channel 3 (CDROM) actually moves the ready sector from the
		// CD data FIFO into RAM, the rest still auto-complete (real GPU/OTC DMA is
		// M12). MADR (+0) and BCR (+4) are the destination and block count.
		if base&0xF == 0x8 && word&0x01000000 != 0 {
			ch := (base - 0x1F801080) >> 4
			madr := m.io[base-8] & 0x1FFFFF
			bcr := m.io[base-4]
			switch ch {
			case 2: // GPU: feed a command list to GP0
				m.dmaGPU(madr, bcr, word)
			case 3: // CDROM: stream the ready sector into RAM
				m.cd.dmaTo(madr, bcr)
			}
			m.io[base] = word &^ 0x01000000 // clear the busy/start bit
			m.raiseIRQ(3)                    // DMA interrupt line
			m.dmaFlags |= 1 << ch            // DICR channel-done flag (write-1-to-clear)
		}
	}
}

// dmaGPU services a GPU DMA (channel 2): in linked-list sync mode (CHCR bits
// 9-10 == 2) it walks the ordering table, handing each node's words to the GPU
// (feedNode, which skips the game's disabled/zeroed primitive slots); in block
// mode it streams BCR words (a VRAM image upload).
func (m *Machine) dmaGPU(madr, bcr, chcr uint32) {
	if (chcr>>9)&3 == 2 { // linked list
		addr := madr
		nodeWords := make([]uint32, 0, 16)
		for i := 0; i < 0x40000; i++ { // guard against a broken/looping list
			addr &= 0x1FFFFF
			header := m.read32(addr)
			words := header >> 24
			nodeWords = nodeWords[:0]
			for j := uint32(0); j < words; j++ {
				addr = (addr + 4) & 0x1FFFFF
				nodeWords = append(nodeWords, m.read32(addr))
			}
			m.gpu.feedNode(nodeWords)
			next := header & 0xFFFFFF
			if next&0x800000 != 0 { // end-of-list marker
				break
			}
			addr = next
		}
		return
	}
	// Block / block-sync: BCR = blocksize | blockcount<<16, all words to GP0.
	n := bcr & 0xFFFF
	if bc := bcr >> 16; bc > 1 {
		n *= bc
	}
	for i := uint32(0); i < n; i++ {
		m.gpu.gp0(m.read32(madr))
		madr = (madr + 4) & 0x1FFFFF
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
