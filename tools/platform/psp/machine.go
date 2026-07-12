package psp

// machine.go is the PSP oracle: an Allegrex core (tools/cpu/allegrex) wired to the
// PSP memory map and a high-level-emulated kernel (kernel.go). It boots a decrypted
// module and runs it, exposing the tracing/watch instrumentation the other machine
// models in this repo provide so the game can be watched producing its data.
//
// The kernel is HLE: instead of loading the firmware, each module import stub's
// `syscall` is patched to a synthetic code that the CPU's Syscall hook dispatches to
// a Go handler, the way the PSX oracle intercepts the BIOS call vectors.

import (
	"fmt"

	"retroreverse.com/tools/cpu/allegrex"
)

const (
	ramBase     = 0x08000000
	ramSize     = 32 * 1024 * 1024 // 32 MiB main RAM
	vramBase    = 0x04000000
	vramSize    = 2 * 1024 * 1024 // 2 MiB
	scratchBase = 0x00010000
	scratchSize = 16 * 1024 // 16 KiB

	// userBase is where the main module is loaded and relocated: the user memory
	// partition (0x08800000) past its reserved head.
	userBase = 0x08804000
	// stackTop is the initial $sp: high in user RAM, clear of the module image.
	stackTop = 0x09FF0000

	// stepsPerVBlank paces the synthetic display interrupt (see run.go).
	stepsPerVBlank = 1000000
)

// phys folds the MIPS kseg mirrors (kseg0/kseg1 at 0x80000000/0xA0000000) down onto
// the physical address; PSP RAM/VRAM/scratchpad already sit in the low range.
func phys(addr uint32) uint32 { return addr & 0x1FFFFFFF }

// Machine is the PSP oracle.
type Machine struct {
	ram     []byte
	vram    []byte
	scratch []byte
	CPU     *allegrex.CPU

	io map[uint32]uint32 // last-written value of otherwise-unmodelled I/O registers

	// Kernel-HLE state (kernel.go).
	syscalls       map[uint32]*syscall // synthetic code -> handler
	nextSyscall    uint32
	handles        map[uint32]*kobject
	nextHandle     uint32
	heapPtr        uint32
	heapEnd        uint32
	threadEntry    uint32   // entry of the "main thread" created by sceKernelCreateThread
	current        *kobject // the running thread (nil = the initial module-start context)
	doneReason     string   // set when scheduling finds nothing runnable
	subIntrs       map[uint32]*subIntr
	vblanks        uint32 // display VBlank counter (sceDisplayGetVcount)
	pad            uint32 // current pad button bits (sceCtrl reads)
	padPrev        uint32 // pad at the previous latch read (edge detection)
	padScript      []PadEvent
	savedataStatus uint32    // savedata-utility dialog status (0 none, 1 init, 2 running, 3 finished, 4 shutdown)
	volatileLocked bool      // the 4 MiB volatile block (sceKernelVolatileMemLock) is held
	mpeg           mpegState // the active sceMpeg session (movie playback)
	atrac          map[uint32]*atracState
	nextAtrac      uint32
	vol            *Volume
	files          map[uint32]*ioFile
	nextFd         uint32
	audioCh        uint32
	SyscallCalls   map[string]int
	tty            []byte
	fbAddr         uint32 // framebuffer set via sceDisplaySetFrameBuf
	fbWidth        uint32 // framebuffer line stride (pixels)
	fbFormat       uint32 // pixel format from sceDisplaySetFrameBuf

	// GE (GPU) display lists submitted via sceGeListEnQueue.
	GeLists  []GeList
	OnGeList func(GeList)
	geSt     *geState // persistent GE register state (survives across lists)

	imageHash string // pinned into savestates

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi   uint32
	OnWrite            func(addr, val, pc uint32)
	RWatchLo, RWatchHi uint32
	OnRead             func(addr, val, pc uint32)
	OnStep             func(m *Machine, pc uint32)

	// Diagnostics.
	Log     []string
	logSeen map[string]bool

	Halted     bool
	HaltReason string
}

// NewMachine builds a reset machine with RAM, VRAM, scratchpad, the CPU and kernel.
func NewMachine() *Machine {
	m := &Machine{
		ram:          make([]byte, ramSize),
		vram:         make([]byte, vramSize),
		scratch:      make([]byte, scratchSize),
		io:           map[uint32]uint32{},
		syscalls:     map[uint32]*syscall{},
		nextSyscall:  0x1000,
		handles:      map[uint32]*kobject{},
		nextHandle:   1,
		subIntrs:     map[uint32]*subIntr{},
		files:        map[uint32]*ioFile{},
		atrac:        map[uint32]*atracState{},
		nextFd:       fdFirstFile,
		SyscallCalls: map[string]int{},
		logSeen:      map[string]bool{},
	}
	m.CPU = allegrex.NewCPU(m)
	m.CPU.Syscall = m.handleSyscall
	return m
}

// PadEvent sets the pad button state from a given VBlank on (the pad holds the
// buttons until the next event). Events are applied in order of AtVblank.
type PadEvent struct {
	AtVblank uint32
	Buttons  uint32
}

// SetPadScript installs a scheduled pad input script for this run.
func (m *Machine) SetPadScript(evs []PadEvent) { m.padScript = evs }

// Vblanks reports the display VBlank counter.
func (m *Machine) Vblanks() uint32 { return m.vblanks }

// note records a one-time diagnostic message.
func (m *Machine) note(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	if m.logSeen[s] {
		return
	}
	m.logSeen[s] = true
	m.Log = append(m.Log, s)
}

// --- allegrex.Bus ----------------------------------------------------------

// Read returns the byte at addr, decoding the PSP memory map.
func (m *Machine) Read(addr uint32) byte {
	p := phys(addr)
	var v byte
	switch {
	case p >= ramBase && p < ramBase+ramSize:
		v = m.ram[p-ramBase]
	case p >= vramBase && p < vramBase+vramSize:
		v = m.vram[p-vramBase]
	case p >= scratchBase && p < scratchBase+scratchSize:
		v = m.scratch[p-scratchBase]
	default:
		m.note("read from unmapped 0x%08X at pc 0x%08X", addr, m.CPU.CurPC())
		v = 0
	}
	if m.OnRead != nil && addr >= m.RWatchLo && addr < m.RWatchHi {
		m.OnRead(addr, uint32(v), m.CPU.CurPC())
	}
	return v
}

// Write stores v at addr, decoding the PSP memory map.
func (m *Machine) Write(addr uint32, v byte) {
	p := phys(addr)
	switch {
	case p >= ramBase && p < ramBase+ramSize:
		m.ram[p-ramBase] = v
	case p >= vramBase && p < vramBase+vramSize:
		m.vram[p-vramBase] = v
	case p >= scratchBase && p < scratchBase+scratchSize:
		m.scratch[p-scratchBase] = v
	default:
		m.io[p] = uint32(v)
		m.note("write 0x%02X to unmapped 0x%08X at pc 0x%08X", v, addr, m.CPU.CurPC())
	}
	if m.OnWrite != nil && addr >= m.WatchLo && addr < m.WatchHi {
		m.OnWrite(addr, uint32(v), m.CPU.CurPC())
	}
}

// --- reads/writes for loaders and instrumentation --------------------------

func (m *Machine) read32(a uint32) uint32 {
	return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 | uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
}
func (m *Machine) write32(a, v uint32) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
	m.Write(a+2, byte(v>>16))
	m.Write(a+3, byte(v>>24))
}

// DisasmAt returns the disassembly of the instruction at addr (for -trace).
func (m *Machine) DisasmAt(addr uint32) string {
	var b [4]byte
	for i := range b {
		b[i] = m.Read(addr + uint32(i))
	}
	return allegrex.Decode(b[:], addr).Text
}

// TTY returns the accumulated kernel stdout/Kprintf output.
func (m *Machine) TTY() string { return string(m.tty) }

// CurrentThread names the running thread ("" for the anonymous module-start
// context), so a trace can be filtered to one thread.
func (m *Machine) CurrentThread() string {
	if m.current == nil {
		return ""
	}
	return m.current.name
}

// KObjects describes every non-thread kernel object (semaphores, event flags,
// pools) with its state, plus which object each waiting thread blocks on.
func (m *Machine) KObjects() []string {
	var out []string
	for h, o := range m.handles {
		switch o.kind {
		case "sema":
			out = append(out, fmt.Sprintf("sema %d %q count %d", h, o.name, o.count))
		case "evflag":
			out = append(out, fmt.Sprintf("evflag %d %q bits 0x%X", h, o.name, o.bits))
		case "vpl":
			out = append(out, fmt.Sprintf("vpl %d %q at 0x%08X used %d/%d", h, o.name, o.addr, o.used, o.size))
		case "thread":
			if o.tstate == thWaiting {
				switch {
				case o.waitEv != 0:
					out = append(out, fmt.Sprintf("thread %q waits evflag %d (bits 0x%X mode 0x%X)",
						o.name, o.waitEv, o.waitBits, o.waitMode))
				case o.waitSema != 0:
					out = append(out, fmt.Sprintf("thread %q waits sema %d (need %d)",
						o.name, o.waitSema, o.waitNeed))
				case o.wakeVblank != 0:
					out = append(out, fmt.Sprintf("thread %q sleeps until vblank %d (now %d)",
						o.name, o.wakeVblank, m.vblanks))
				default:
					out = append(out, fmt.Sprintf("thread %q sleeps (WakeupThread)", o.name))
				}
			}
		}
	}
	return out
}

// Threads describes every thread kobject (name, entry, priority, state, and the
// PC it is parked at), for the oracle's diagnostics.
func (m *Machine) Threads() []string {
	states := [...]string{"dormant", "ready", "running", "waiting"}
	var out []string
	for h, o := range m.handles {
		if o.kind != "thread" {
			continue
		}
		pc := o.ctx.PC
		if o == m.current {
			pc = m.CPU.PC
		}
		out = append(out, fmt.Sprintf("thread %d %q entry 0x%08X prio %d %s pc 0x%08X",
			h, o.name, o.entry, o.priority, states[o.tstate], pc))
	}
	return out
}

// --- module loading --------------------------------------------------------

// LoadModule relocates a module to the user partition, copies its segments into RAM,
// seeds the entry/gp/sp, and installs the kernel-HLE syscall stubs from its imports.
func (m *Machine) LoadModule(mod *Module) error {
	if mod.Type == etPRX {
		mod.Relocate(userBase)
	}
	for _, s := range mod.Segments {
		for i, b := range s.Data {
			m.writeRAM(s.VAddr+uint32(i), b)
		}
		// The rest up to MemSize is bss (already zero in a fresh RAM).
	}
	k0 := m.threadK0(0x1000, stackTop) // the module-start context area ($k0)
	m.CPU.SetPC(mod.EntryPC)
	m.CPU.SetReg(26, k0)             // $k0
	m.CPU.SetReg(28, mod.GP)         // $gp
	m.CPU.SetReg(29, k0)             // $sp
	m.CPU.SetReg(30, k0)             // $fp
	m.CPU.SetReg(31, threadExitAddr) // $ra: module_start "returns" to the scheduler
	m.CPU.SetReg(4, 0)               // $a0 = argc
	m.CPU.SetReg(5, 0)               // $a1 = argv

	// A bump heap above the module image for sceKernelAllocPartitionMemory.
	var top uint32
	for _, s := range mod.Segments {
		if end := s.VAddr + s.MemSize; end > top {
			top = end
		}
	}
	m.heapPtr = (top + 0xFFF) &^ 0xFFF
	m.heapEnd = stackTop - 0x100000

	m.installStubs(mod)
	return nil
}

// writeRAM writes a byte directly to a mapped region without firing watch hooks
// (used by the loader).
func (m *Machine) writeRAM(addr uint32, v byte) {
	p := phys(addr)
	switch {
	case p >= ramBase && p < ramBase+ramSize:
		m.ram[p-ramBase] = v
	case p >= vramBase && p < vramBase+vramSize:
		m.vram[p-vramBase] = v
	case p >= scratchBase && p < scratchBase+scratchSize:
		m.scratch[p-scratchBase] = v
	}
}

// SetImageHash pins the source image's hash for savestate validation.
func (m *Machine) SetImageHash(h string) { m.imageHash = h }
