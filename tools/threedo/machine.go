package threedo

// machine.go is the 3DO boot oracle: an ARM60 (tools/arm60) wired to the 3DO
// memory map, with the Madam/Clio custom chips stubbed and the Portfolio OS
// high-level-emulated, so the game's boot executable can be run and traced. It
// mirrors the PSX oracle in tools/psx/machine.go (an HLE BIOS + stubbed GPU/DMA).
//
// The 3DO boots by having the kernel load an AIF program and enter it with a
// register (r7 here) pointing at the kernel/folio base; the program then calls
// OS services indirectly through negative offsets from that base
// (`LDR pc, [r9, #-0x78]`, seen in LaunchMe's entry). We do not have the OS, so
// we synthesise that base: a vector table is planted in RAM whose every slot
// jumps into a reserved "HLE" address window, and the run loop intercepts a PC
// landing there, logs which folio offset was called, stubs a result and returns.
// That turns an un-runnable app into a live trace of its Portfolio call sequence.

import (
	"fmt"

	"retroreverse.com/tools/arm60"
)

const (
	dramSize   = 2 * 1024 * 1024 // 2 MiB main DRAM at 0x00000000
	vramBase   = 0x00200000
	vramSize   = 1024 * 1024 // 1 MiB VRAM
	madamBase  = 0x03300000  // Madam (CEL/matrix/DMA) registers
	madamEnd   = 0x03400000
	clioBase   = 0x03400000 // Clio (video/audio/timers/IRQ) registers
	clioEnd    = 0x03500000
	kernelBase = 0x00180000 // synthetic Portfolio kernel/folio base (r7/r9)
	hleBase    = 0x0FE00000 // reserved window: a PC here is an intercepted folio call
	hleSize    = 0x00010000
)

// KernelCall records one intercepted Portfolio folio/kernel call.
type KernelCall struct {
	Offset uint32    // the negative offset from the kernel base (the "vector")
	From   uint32    // the call site (LR-8)
	Args   [4]uint32 // r0..r3 at the call
}

// Machine is the 3DO oracle. It satisfies arm60.Bus.
type Machine struct {
	dram []byte
	vram []byte
	CPU  *arm60.CPU

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi uint32
	OnWrite          func(addr, val, pc uint32)
	OnStep           func(m *Machine, pc uint32)

	KernelCalls []KernelCall
	tty         []byte
	Log         []string
	logSeen     map[string]bool

	Halted     bool
	HaltReason string
}

// NewMachine builds a reset 3DO machine with DRAM, VRAM and an ARM60 core.
func NewMachine() *Machine {
	m := &Machine{
		dram:    make([]byte, dramSize),
		vram:    make([]byte, vramSize),
		logSeen: map[string]bool{},
	}
	m.CPU = arm60.NewCPU(m)
	m.CPU.SWI = m.swi
	return m
}

// LoadAIF copies an AIF image into DRAM at its base, plants the synthetic kernel
// vector table, seeds the entry registers the OS would supply, and points the CPU
// at the header (an executable AIF runs its decompress/reloc/zero-init/entry
// branch sequence from word 0).
func (m *Machine) LoadAIF(a *AIF) {
	base := a.ImageBase & (dramSize - 1)
	copy(m.dram[base:], a.Image)

	// Plant the folio vector table: each word just below the kernel base jumps
	// into the HLE window, encoding its own offset so the run loop can identify
	// the call. Cover a generous span of negative offsets.
	for off := uint32(4); off <= 0x1000; off += 4 {
		m.writeWord(kernelBase-off, hleBase+off)
	}

	m.CPU.SetReg(5, 0)          // r5: argc-like
	m.CPU.SetReg(6, 0)          // r6: argv-like
	m.CPU.SetReg(7, kernelBase) // r7: kernel/folio base (entry copies it to r9)
	m.CPU.SetReg(13, dramSize-0x1000)
	m.CPU.SetPC(a.ImageBase)
}

// writeWord stores a big-endian word directly into DRAM (setup helper).
func (m *Machine) writeWord(a, v uint32) {
	if int(a)+4 <= len(m.dram) {
		m.dram[a] = byte(v >> 24)
		m.dram[a+1] = byte(v >> 16)
		m.dram[a+2] = byte(v >> 8)
		m.dram[a+3] = byte(v)
	}
}

func (m *Machine) note(s string) {
	if !m.logSeen[s] {
		m.logSeen[s] = true
		m.Log = append(m.Log, s)
	}
}

// TTY returns anything the game wrote through the (HLE) console.
func (m *Machine) TTY() string { return string(m.tty) }

// --- arm60.Bus -------------------------------------------------------------

func (m *Machine) Read(addr uint32) byte {
	switch {
	case addr < dramSize:
		return m.dram[addr]
	case addr >= vramBase && addr < vramBase+vramSize:
		return m.vram[addr-vramBase]
	case addr >= madamBase && addr < clioEnd:
		return 0 // Madam/Clio registers read as 0 under HLE
	default:
		m.note(fmt.Sprintf("read unmapped 0x%08X", addr))
		return 0
	}
}

func (m *Machine) Write(addr uint32, v byte) {
	if m.OnWrite != nil && addr >= m.WatchLo && addr < m.WatchHi {
		m.OnWrite(addr, uint32(v), m.CPU.CurPC())
	}
	switch {
	case addr < dramSize:
		m.dram[addr] = v
	case addr >= vramBase && addr < vramBase+vramSize:
		m.vram[addr-vramBase] = v
	case addr >= madamBase && addr < clioEnd:
		// Madam/Clio register writes are accepted and ignored under HLE.
	default:
		m.note(fmt.Sprintf("write unmapped 0x%08X", addr))
	}
}

// swi services the ARM SWI gate. The AIF exit SWI (#0x11) stops the machine;
// other SWIs are logged and returned from.
func (m *Machine) swi(c *arm60.CPU, comment uint32) bool {
	switch comment {
	case 0x11: // program exit
		m.Halted, m.HaltReason = true, "program exit (SWI #0x11)"
	default:
		m.note(fmt.Sprintf("SWI #0x%X", comment))
	}
	return true // serviced: do not vector to 0x08
}
