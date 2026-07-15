package xbox

// machine.go is the original-Xbox Machine: 64 MB of unified RAM, the Pentium III
// (the shared tools/cpu/x86 core in flat 32-bit protected mode), and the NV2A MMIO
// window. It loads a title's XBE at its fixed base, patches the kernel-import thunks
// so each xboxkrnl call traps into a Go handler (kernel.go), and runs the title's
// XDK/CRT boot code until it reaches the first NV2A push-buffer kick or halts on a
// kernel facility not yet modelled.
//
// This is the Phase-B bring-up host. The CPU (protected mode + x87) was proven in
// Phase A on a DOS-extender game (Quake, tools/platform/dos); the XISO/XBE static
// tooling and the 151-ordinal import census were Phase 0 (xiso.go, xbe.go). Here the
// two meet: real title code runs against an HLE of the exports it imports.
//
// Memory model. The Xbox address space is flat 32-bit with a handful of fixed
// windows onto the same 64 MB of physical RAM (there is no per-title paging that a
// title relies on — contiguous allocations are physically contiguous by contract):
//
//	0x00000000..0x03FFFFFF  physical RAM, identity (the title at 0x10000, stacks, heap)
//	0x80000000..0x83FFFFFF  cached kernel window   -> phys = va - 0x80000000
//	0xB0000000..0xB3FFFFFF  uncached kernel window  -> phys = va - 0xB0000000
//	0xD0000000..0xD3FFFFFF  physical/write-back     -> phys = va & 0x03FFFFFF
//	0xF0000000..0xFFFFFFFF  uncached; the NV2A MMIO aperture lives here (0xFD000000)
//
// Every guest access goes through translate(); the four RAM windows fold onto the
// one backing slice, and the MMIO aperture is handled by nv2a.go. Kernel bookkeeping
// (the KPCR, thread/dispatcher objects, the pool) lives in a reserved band at the top
// of physical RAM so it never collides with what the title allocates from below.

import (
	"fmt"

	"retroreverse.com/tools/cpu/x86"
)

const (
	ramSize = 64 << 20 // 0x04000000: the Xbox's unified 64 MB of RAM

	// The four RAM windows onto physical memory. Each is 64 MB wide (one RAM's worth).
	cachedBase   = 0x80000000
	uncachedBase = 0xB0000000
	physBase     = 0xD0000000
	windowMask   = 0x03FFFFFF // low 26 bits select a physical byte within a window

	// The NV2A lives at the top of the address space; its register aperture is 16 MB
	// at 0xFD000000 (nv2a.go decodes it). The GPU itself is Phase C — this phase only
	// needs to see the first push-buffer kick land here.
	mmioBase = 0xFD000000
	mmioTop  = 0xFE000000

	// Reserved kernel-bookkeeping band at the top of physical RAM. The KPCR, the
	// thread and dispatcher objects the HLE hands out, and the trap thunks' scratch
	// live here; the title's own allocators (heap, contiguous framebuffers) bump up
	// from low RAM and down from kernelBandBase, so the two never meet.
	kernelBandBase = ramSize - (2 << 20) // top 2 MiB of RAM
	kpcrAddr       = kernelBandBase      // the KPCR sits at the base of the band

	// Kernel-import trap region. It is NOT backed by RAM: each imported ordinal is
	// assigned a unique sentinel address here, written into the title's IAT thunk in
	// place of the real function pointer. When a CALL through the thunk lands the PC
	// on a sentinel, onStep dispatches the ordinal to its Go handler and simulates the
	// (stdcall) return — the sentinel bytes are never fetched. Chosen high and aligned
	// so no title code or data occupies it.
	trapBase   = 0x8F000000
	trapStride = 16  // bytes per ordinal sentinel (ordinal N -> trapBase + N*16)
	trapCount  = 512 // ordinals 1..511 fit; xboxkrnl tops out at ~366
	trapTop    = trapBase + trapCount*trapStride

	// Default title stack. XapiInitProcess runs on a stack the launcher set up; we
	// give the entry thread a generous one high in physical RAM, below the kernel band.
	titleStackTop  = kernelBandBase - 0x1000
	titleStackSize = 512 << 10
)

// Machine is a loaded Xbox title ready to run.
type Machine struct {
	RAM []byte // 64 MB of physical RAM, the backing for every window
	CPU *x86.CPU

	XBE  *XBE   // the parsed executable
	Disc *Image // the mounted disc, for kernel file I/O (nil if none)

	// Allocators. Physical RAM below kernelBandBase is split between the title's heap
	// (grows up from the image end) and contiguous/pool allocations (grow down from
	// kernelBandBase). poolNext bumps down; heapNext bumps up.
	poolNext uint32 // next contiguous/pool allocation (grows DOWN from kernelBandBase)
	heapNext uint32 // next virtual-memory allocation (grows UP from the image end)
	heapTop  uint32 // ceiling for the up-growing heap (meets poolNext)

	// Kernel HLE state (kernel.go, thread.go). Handles are opaque nonzero pointers
	// into the reserved band; objects are looked up by handle.
	objects     map[uint32]*kobject
	files       map[uint32]*fileObject // open disc file handles (kernel_file.go)
	nextObjAddr uint32                 // bump pointer within the kernel band for new dispatcher objects
	kbandNext   uint32                 // general bump pointer within the kernel band (KPCR follows, then objects)

	// Threads and scheduling (thread.go / sched.go).
	threads     []*thread
	current     *thread
	nextTID     uint32
	rrCursor    int
	reschedule  bool
	quantumLeft int

	tick uint64 // synthetic system tick, advanced per instruction (for timers/clocks)

	tickCountAddr  uint32 // guest address of the live KeTickCount data export (0 if none)
	systemTimeAddr uint32 // guest address of the live KeSystemTime data export (0 if none)

	// NV2A MMIO (nv2a.go), the PFIFO DMA pusher (nv2a_pfifo.go), and the graphics
	// engine (nv2a_pgraph.go).
	nv     nv2a
	push   pusherState
	pgraph *pgraph

	pciAddr uint32 // last value written to the PCI config-address port 0xCF8 (ports.go)

	// PCI configuration space, byte-addressed by (bus<<24 | slot<<16 | register). The
	// D3D device init read-modify-writes a few NV2A config registers through
	// HalReadWritePCISpace; backing them here keeps a read-after-write coherent.
	pciSpace map[uint32]byte

	// Instrumentation.
	Log         []string
	OrdinalHits map[uint16]int  // xboxkrnl ordinal -> call count
	dataDeref   map[uint16]bool // data-export ordinals that have been dereferenced (log-once)
	Halted        bool
	HaltReason    string
	firstPush     bool // the first NV2A push-buffer kick has been observed
	pusherEnabled bool // Phase C: run the DMA pusher on each DMA_PUT write (nv2a_pfifo.go)

	// Memory watches (the debugger's Watcher), same shape as the DOS host.
	wWLo, wWHi uint32
	onW        func(addr, val, pc uint32)
	wRLo, wRHi uint32
	onR        func(addr, val, pc uint32)

	verbose   bool
	traceLeft int // remaining instructions to print a PC/disasm trail for (-trace)
}

// NewMachine builds a machine from a parsed XBE, loading its sections into RAM at the
// image base and patching its kernel-import thunks. disc may be nil (the title then
// cannot open files, which surfaces as a kernel halt if it tries).
func NewMachine(xbe *XBE, disc *Image) (*Machine, error) {
	m := &Machine{
		RAM:         make([]byte, ramSize),
		XBE:         xbe,
		Disc:        disc,
		objects:     map[uint32]*kobject{},
		files:       map[uint32]*fileObject{},
		OrdinalHits: map[uint16]int{},
		dataDeref:   map[uint16]bool{},
	}
	m.nv.reg = map[uint32]uint32{}
	m.pciSpace = map[uint32]byte{}
	m.pgraph = newPgraph(m)

	if err := m.loadImage(); err != nil {
		return nil, err
	}
	m.setupMemoryLayout()
	m.setupKPCR()
	m.patchThunks()

	c := x86.NewCPU(m)
	c.Mode = x86.ModeProt
	// Flat selectors: CS/DS/ES/SS all base 0, limit 4 GB. FS points at the KPCR (Xbox
	// convention — the CRT and kernel read the current thread and stack limits through
	// FS). We give FS a distinct base via SegBase and never let a selector load reset
	// it (SegResolve returns the FS base for the FS selector and 0 for the rest).
	for i := range c.SegBase {
		c.SegBase[i] = 0
	}
	c.Seg[x86.CS] = 0x08 // flat code
	c.Seg[x86.DS] = 0x10
	c.Seg[x86.ES] = 0x10
	c.Seg[x86.SS] = 0x10
	c.Seg[x86.GS] = 0x10
	c.Seg[x86.FS] = fsSelector
	c.SegBase[x86.FS] = kpcrAddr
	c.SegResolve = m.resolveSel
	c.IP = xbe.Entry
	// Seed the boot stack with the thread-exit sentinel as the entry's return address:
	// the XBE entry is a launcher that spawns the game's main thread (PsCreateSystemThreadEx),
	// closes its handle, and returns — on hardware into a kernel thread-terminate stub. With
	// the sentinel on top, that outermost RET lands on threadExitAddr and onStep retires
	// thread 0 cleanly (like createThread does for spawned threads), handing the machine to
	// the main thread. Without it the RET popped 0 and derailed the boot into low memory.
	c.Regs[x86.SP] = titleStackTop - 4
	m.write32(titleStackTop-4, threadExitAddr)
	c.IF = true
	c.OnStep = m.onStep
	c.PortIn = m.portIn
	c.PortOut = m.portOut
	m.CPU = c

	// The entry thread is thread 0 — its context is simply the live CPU. The scheduler
	// (thread.go) treats a nil m.current as "the boot/entry context", the same way the
	// PSP HLE does; the first CreateThread/PsCreateSystemThreadEx materialises threads.
	m.bootThread()
	return m, nil
}

// fsSelector is the selector value we leave in FS; its base is the KPCR (set directly
// in SegBase and preserved because resolveSel returns the KPCR base for it).
const fsSelector = 0x38

// resolveSel maps a protected-mode selector to its linear base. Every flat selector is
// base 0; the FS selector alone carries the KPCR base. Loading DS/ES/SS/CS therefore
// stays flat, while an explicit FS reload keeps pointing at the KPCR.
func (m *Machine) resolveSel(sel uint16) uint32 {
	if sel == fsSelector {
		return kpcrAddr
	}
	return 0
}

// loadImage copies each XBE section into RAM at its virtual address. Xbox VAs are
// physical low addresses (base 0x10000), so a section at VA v lands at RAM[v]. .bss /
// uninitialised tail is left zeroed (make() did that).
func (m *Machine) loadImage() error {
	x := m.XBE
	for _, s := range x.Sections {
		if s.RawSize == 0 {
			continue // .bss-like: virtual size only, already zero
		}
		off, ok := x.atVA(s.VAddr)
		if !ok {
			return fmt.Errorf("xbox: section %q VA %#x has no file mapping", s.Name, s.VAddr)
		}
		if int(s.VAddr)+int(s.RawSize) > len(m.RAM) {
			return fmt.Errorf("xbox: section %q at VA %#x..%#x exceeds 64 MB RAM",
				s.Name, s.VAddr, s.VAddr+s.RawSize)
		}
		if off+int(s.RawSize) > len(x.raw) {
			return fmt.Errorf("xbox: section %q raw range %#x..%#x overruns the %d-byte image",
				s.Name, off, off+int(s.RawSize), len(x.raw))
		}
		copy(m.RAM[s.VAddr:s.VAddr+s.RawSize], x.raw[off:off+int(s.RawSize)])
	}
	return nil
}

// setupMemoryLayout fixes the allocator boundaries. The title's image occupies
// [base, imageEnd); the up-growing virtual-memory heap starts page-aligned above it,
// and the down-growing contiguous/pool allocator starts just below the kernel band.
func (m *Machine) setupMemoryLayout() {
	end := m.XBE.Base + m.XBE.ImageSize
	m.heapNext = align32(end, 0x1000)
	m.poolNext = titleStackTop - titleStackSize // reserve the entry stack below the band
	m.heapTop = m.poolNext
	// The kernel band: KPCR first, then a bump arena for dispatcher/thread objects.
	m.kbandNext = kpcrAddr + kpcrSize
	m.nextObjAddr = m.kbandNext
}

// --- x86.Bus: address-window translation over the one 64 MB backing ---

// translate folds a guest virtual address onto a physical RAM index, or reports that
// it is MMIO (handled separately) or out of range. ok=false with mmio=false is a fault.
func (m *Machine) translate(a uint32) (phys uint32, mmio, ok bool) {
	switch {
	case a < ramSize:
		return a, false, true
	case a >= cachedBase && a < cachedBase+ramSize:
		return a - cachedBase, false, true
	case a >= uncachedBase && a < uncachedBase+ramSize:
		return a - uncachedBase, false, true
	case a >= physBase && a < physBase+ramSize:
		return a & windowMask, false, true
	case a >= mmioBase && a < mmioTop:
		return a - mmioBase, true, true
	default:
		return 0, false, false
	}
}

func (m *Machine) Read(a uint32) byte {
	// A read inside the kernel-import trap region is a *data export* being dereferenced
	// (the game did MOV reg,[IAT slot] then read through it). Functions are dispatched
	// on execution (onStep sees the PC land on the sentinel); only data exports are read
	// as bytes. Returning 0 is the safe default — a zeroed flag/handle/pointer — and
	// avoids the fault that first revealed these. An ordinal that genuinely needs a
	// non-zero value gets an explicit populated block via dataExportSize instead.
	if a >= trapBase && a < trapTop {
		ord := uint16((a - trapBase) / trapStride)
		if !m.dataDeref[ord] {
			m.dataDeref[ord] = true
			m.logf("kernel: data export ordinal %d (%s) dereferenced -> 0", ord, ordinalName(ord))
		}
		return 0
	}
	phys, mmio, ok := m.translate(a)
	if !ok {
		m.fault("read", a)
		return 0xFF
	}
	if mmio {
		return m.nvRead(phys)
	}
	v := m.RAM[phys]
	if m.onR != nil && a >= m.wRLo && a < m.wRHi {
		m.onR(a, uint32(v), m.watchPC())
	}
	return v
}

func (m *Machine) Write(a uint32, v byte) {
	if a >= trapBase && a < trapTop {
		return // a write through a data-export slot: discard (nothing reads it back)
	}
	phys, mmio, ok := m.translate(a)
	if !ok {
		m.fault("write", a)
		return
	}
	if mmio {
		m.nvWrite(phys, v)
		return
	}
	m.RAM[phys] = v
	if m.onW != nil && a >= m.wWLo && a < m.wWHi {
		m.onW(a, uint32(v), m.watchPC())
	}
}

// watchPC is the linear PC of the instruction making a watched access.
func (m *Machine) watchPC() uint32 {
	if m.CPU == nil {
		return 0
	}
	return m.CPU.LinearPC()
}

// SetWriteWatch / SetReadWatch install the debugger's memory watches ([lo,hi)).
func (m *Machine) SetWriteWatch(lo, hi uint32, cb func(addr, val, pc uint32)) {
	m.wWLo, m.wWHi, m.onW = lo, hi, cb
}
func (m *Machine) SetReadWatch(lo, hi uint32, cb func(addr, val, pc uint32)) {
	m.wRLo, m.wRHi, m.onR = lo, hi, cb
}

func (m *Machine) fault(kind string, a uint32) {
	if m.CPU != nil && !m.CPU.Halted {
		m.CPU.Halt("out-of-range %s at %08X (PC %08X)", kind, a, m.CPU.LinearPC())
	}
}

// --- little-endian guest memory helpers (kernel HLE convenience) ---

func (m *Machine) read32(a uint32) uint32 {
	return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 | uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
}
func (m *Machine) read16(a uint32) uint16 { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
func (m *Machine) write32(a, v uint32) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
	m.Write(a+2, byte(v>>16))
	m.Write(a+3, byte(v>>24))
}
func (m *Machine) write16(a uint32, v uint16) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
}

// cstr reads a NUL-terminated ASCII string from guest memory.
func (m *Machine) cstr(a uint32) string {
	var b []byte
	for i := uint32(0); i < 1024 && a != 0; i++ {
		ch := m.Read(a + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

func (m *Machine) logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	m.Log = append(m.Log, s)
	if m.verbose {
		fmt.Println(s)
	}
}

// SetVerbose toggles live logging of kernel calls and events.
func (m *Machine) SetVerbose(v bool) { m.verbose = v }

// EnableGPU turns on the NV2A DMA pusher (Phase C): from now on a DMA_PUT write runs the
// push-buffer parser and the graphics engine, rather than only flagging the first kick.
// Run() then no longer stops at firstPush — it runs the title on so the pusher sees the
// real rendering command stream.
func (m *Machine) EnableGPU() { m.pusherEnabled = true }

// PGraph exposes the NV2A graphics engine (counters, method survey, framebuffer export).
func (m *Machine) PGraph() *pgraph { return m.pgraph }

// SetTrace prints the next n executed instructions (PC + disassembly), for bring-up.
func (m *Machine) SetTrace(n int) { m.traceLeft = n }

// MemReadByte / MemRead32 read guest memory through the address-window translation —
// the debugger/CLI view into the running machine.
func (m *Machine) MemReadByte(a uint32) byte { return m.Read(a) }
func (m *Machine) MemRead32(a uint32) uint32 { return m.read32(a) }

// disasmAt decodes the instruction at guest linear address a into text (best-effort:
// reads up to 16 bytes through the window translation).
func (m *Machine) disasmAt(a uint32) string {
	var buf [16]byte
	for i := range buf {
		buf[i] = m.Read(a + uint32(i))
	}
	inst := x86.Decode32(buf[:], a)
	return inst.Text
}

// DisasmForward returns n disassembled instructions starting at guest address a — the
// bring-up view of a call site (run it from a bit before a halt's return address to see
// the argument pushes, which pin the ordinal's signature).
func (m *Machine) DisasmForward(a uint32, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var buf [16]byte
		for j := range buf {
			buf[j] = m.Read(a + uint32(j))
		}
		inst := x86.Decode32(buf[:], a)
		out = append(out, fmt.Sprintf("  %08X  %s", a, inst.Text))
		if inst.Len <= 0 {
			break
		}
		a += uint32(inst.Len)
	}
	return out
}

// CallerReturnAddr is the return address on top of the stack — at a kernel-trap halt,
// the instruction after the offending CALL. Disassembling backward from it shows the
// argument frame.
func (m *Machine) CallerReturnAddr() uint32 { return m.read32(m.CPU.Regs[x86.SP]) }

func align32(v, a uint32) uint32 { return (v + a - 1) &^ (a - 1) }
