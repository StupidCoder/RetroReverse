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
	wcBase       = 0xF0000000 // write-combined RAM alias (D3D texture/surface pointers)
	windowMask   = 0x03FFFFFF // low 26 bits select a physical byte within a window

	// The NV2A lives at the top of the address space; its register aperture is 16 MB
	// at 0xFD000000 (nv2a.go decodes it). The GPU itself is Phase C — this phase only
	// needs to see the first push-buffer kick land here.
	mmioBase = 0xFD000000
	mmioTop  = 0xFE000000

	// Reserved kernel-bookkeeping band at the top of physical RAM. The KPCR and the
	// thread and dispatcher objects the HLE hands out live here; the title's own
	// allocators (heap, contiguous framebuffers) bump up from low RAM and down from
	// kernelBandBase, so the two never meet. The band must stay SMALL: OutRun's own
	// memory plan (a fixed 0x2AE147A contiguous arena + a 0xB4CCCD committed heap
	// arena + the 0x652AC0 image) totals 63.1 MB of the 64, so the real kernel plus
	// every runtime pool allocation provably fits in the remaining ~900 KB — a 2 MB
	// band pushed that title's own last allocations (the DirectSound voice buffers,
	// worker-thread stacks) into honest-looking E_OUTOFMEMORY failures the real
	// console never sees.
	kernelBandBase = ramSize - (256 << 10) // top 256 KiB of RAM
	kpcrAddr       = kernelBandBase        // the KPCR sits at the base of the band

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

	// Default title stack. XapiInitProcess runs on the launcher's stack only briefly
	// (it spawns the game's main thread, which gets its own PsCreateSystemThreadEx
	// stack, and returns); 64 KiB matches the modest footprint the title's memory
	// plan leaves for it.
	titleStackTop  = kernelBandBase - 0x1000
	titleStackSize = 64 << 10
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
	files       map[uint32]*fileObject // open disc/HDD file handles (kernel_file.go)
	cacheFS     map[string]*cacheFile  // the writable HDD title partitions (T:/U:/Z:)
	fileBasic   map[string][]byte      // per-key FileBasicInformation blobs the title set (NtSetInformationFile class 4)
	pendingIO   []pendingIO            // paced async read completions (kernel_file.go)
	poolSizes   map[uint32]uint32      // ExAllocatePoolWithTag block -> byte size (ExQueryPoolBlockSize)
	nextObjAddr uint32                 // bump pointer within the kernel band for new dispatcher objects
	kbandNext   uint32                 // general bump pointer within the kernel band (KPCR follows, then objects)

	// Threads and scheduling (thread.go / sched.go).
	threads     []*thread
	current     *thread
	nextTID     uint32
	rrCursor    int
	reschedule  bool
	quantumLeft int

	// Hardware interrupt delivery (interrupt.go): connected KINTERRUPTs by vector,
	// the VBlank pacing, and the saved context while an ISR frame runs.
	interrupts map[uint32]uint32 // vector -> KINTERRUPT guest address (KeConnectInterrupt)
	nextVBlank uint64            // tick of the next vertical blank
	isrActive  bool              // an ISR/DPC frame is running on the current thread
	isrSaved   x86.CPU           // the interrupted context, restored at isrExitAddr
	dpcQueue   []dpcEntry        // DPCs queued by the running ISR (KeInsertQueueDpc)

	tick uint64 // synthetic system tick, advanced per instruction (for timers/clocks)

	// The guest-clock epoch: systemTime100ns/guestTSC are affine in (tick - clockBaseTick)
	// so a savestate taken under an older instrsPerMs declaration resumes with continuous
	// time and TSC (see sched.go systemTime100ns and state.go's legacy migration).
	clockBaseTick  uint64
	clockBase100ns uint64
	tscBase        uint64

	tickCountAddr  uint32 // guest address of the live KeTickCount data export (0 if none)
	systemTimeAddr uint32 // guest address of the live KeSystemTime data export (0 if none)

	// XcSHA* streaming contexts (kernel_crypto.go), keyed by the guest context address the
	// title passes to XcSHAInit/Update/Final. The value is a crypto/sha1 digest marshalled
	// via BinaryMarshaler, so it gob-serialises directly into a savestate.
	shaCtx map[uint32][]byte

	// XcRC4* key schedules (kernel_crypto.go), keyed by the guest key-table address. The
	// value is the 258-byte RC4 state (S[256] then i, j) so it gob-serialises directly. The
	// state is kept host-side rather than in the (opaque, unknown-layout) guest buffer.
	rc4Ctx map[uint32][]byte

	// NV2A MMIO (nv2a.go), the PFIFO DMA pusher (nv2a_pfifo.go), and the graphics
	// engine (nv2a_pgraph.go).
	nv     nv2a
	push   pusherState
	pgraph *pgraph

	// MCPX device MMIO: the APU's, AC'97 codec's and NIC's latch apertures (apu.go),
	// and the USB OHCI host controller (usb.go), whose register file is still backed by
	// a latch's sparse map — that map is the savestate's USBReg, so the controller the
	// title programmed survives a restore — with its semantics layered over the top.
	apu  mmioLatch
	ac97 mmioLatch
	usb  mmioLatch
	nic  mmioLatch

	// usbFrameServed is the last USB frame number usbTick ran (usb.go). The frame
	// number itself is derived from tick and stored nowhere; this is only the
	// catch-up cursor, and it rides in the savestate so a restore does not replay.
	usbFrameServed uint64

	// usbWrDword reassembles the dword a guest store is writing to the OHCI, byte by
	// byte, so the trace can report the command the driver issued rather than what the
	// register was left holding. Instrumentation, not state: it cannot outlive the
	// single instruction that fills it.
	usbWrDword uint32

	// Armed KTIMERs (timer.go). The queue is small — a handful of drivers' heartbeats
	// and XAPI's port debounce — so it is a slice scanned per coarse tick rather than a
	// heap: the cost is in the scan, and the clarity is worth more than the ordering.
	timers []ktimer

	// The devices on the root hub's four ports (usb_xid.go), and the transfer engine's
	// working state (usb_ohci.go): the TDs retired this frame and awaiting writeback,
	// and the control pipe's in-flight IN data with its cursor.
	usbDev      [usbPorts]usbDevice
	usbDone     []uint32
	usbCtrlData []byte
	usbCtrlOff  int

	pciAddr uint32 // last value written to the PCI config-address port 0xCF8 (ports.go)

	// PCI configuration space, byte-addressed by (bus<<24 | slot<<16 | register). The
	// D3D device init read-modify-writes a few NV2A config registers through
	// HalReadWritePCISpace; backing them here keeps a read-after-write coherent.
	pciSpace map[uint32]byte

	// Instrumentation.
	Log           []string
	OrdinalHits   map[uint16]int  // xboxkrnl ordinal -> call count
	dataDeref     map[uint16]bool // data-export ordinals that have been dereferenced (log-once)
	dosErrWarned  map[uint32]bool // NTSTATUS values RtlNtStatusToDosError could not map (log-once)
	Halted        bool
	HaltReason    string
	firstPush     bool // the first NV2A push-buffer kick has been observed
	pusherEnabled bool // Phase C: run the DMA pusher on each DMA_PUT write (nv2a_pfifo.go)

	// Memory watches (the debugger's Watcher), same shape as the DOS host.
	wWLo, wWHi uint32
	onW        func(addr, val, pc uint32)
	wRLo, wRHi uint32
	onR        func(addr, val, pc uint32)

	// OnPixel is the debugger's per-fragment provenance hook (the n3ds/gc shape):
	// every fragment the rasteriser produces — drawn, depth-killed, or alpha-killed —
	// reports here when installed. Nil by default and outside the savestate.
	OnPixel func(x, y uint32, ev PixelEvent)

	// OnNVMethod is the debugger's command hook: every (subchannel, method, argument)
	// the PFIFO pusher decodes reports here BEFORE the engine acts on it, so a hook can
	// number the command that is about to draw. Nil by default, outside the savestate.
	OnNVMethod func(m *Machine, subchan, method, arg uint32)

	// OnFlip fires when the title presents a frame — when it registers a new scanout
	// through AvSetDisplayMode, which on this console is the swap (see kernel.go
	// ordinal 3). Nil by default, outside the savestate.
	OnFlip func(*Machine)

	// FlipVSync models FLIP_STALL's real time cost: a Present blocks until the vblank the
	// CRTC owes. The guest TSC is instruction-paced, so without this a rendered frame
	// advances the clock only by the instructions its render code retired (~0.18 of a
	// field), and OutRun's RDTSC fixed-timestep loop (0x20AFA) then steps its simulation
	// only every ~6th present — a 10 FPS sim on a 60 FPS engine. When set, each FLIP_STALL
	// advances the tick to the next vblank so one present costs one field and the loop
	// steps every present. A FIDELITY change: it re-times every trajectory (savestates
	// move), so it is off by default and gated, not the standard behaviour. See sched.go's
	// note on why the guest clock undercounts a rendered frame's real cycle cost.
	FlipVSync bool

	// StopRequested asks the run loops to stop at the next safe boundary: the CPU
	// between instructions, the pusher between commands. A hook sets it to end a run
	// (the frame hook at a flip, a breaking watch); Run clears it when it returns.
	//
	// It is a flag rather than a method because the machine is single-threaded by
	// contract — only a hook running INSIDE the run may set it, never another
	// goroutine, which is what keeps the whole model race-free.
	StopRequested bool

	// stopAfterMethod counts down the pusher methods left to dispatch before the run
	// stops, for the debugger's command scrubber. Zero means no limit.
	stopAfterMethod int
	stopAfterArmed  bool

	bps map[uint32]bool // execution breakpoints, by linear PC

	verbose   bool
	traceLeft int // remaining instructions to print a PC/disasm trail for (-trace)

	// hotpc is the sampling PC profiler (RR_HOTPC=1): every 256th tick records the
	// current PC. Diagnostic only — outside the savestate, nil unless enabled.
	hotpc map[uint32]uint64

	// Profile gates the per-subsystem frame profiler (profile.go). Off by default and
	// outside the savestate; the debugger and the oracle's -profile flag turn it on.
	Profile bool
	prof    profState
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
		cacheFS:     map[string]*cacheFile{},
		fileBasic:   map[string][]byte{},
		poolSizes:   map[uint32]uint32{},
		OrdinalHits: map[uint16]int{},
		dataDeref:   map[uint16]bool{},
		shaCtx:      map[uint32][]byte{},
		rc4Ctx:      map[uint32][]byte{},
		interrupts:  map[uint32]uint32{},
	}
	m.FlipVSync = flipVSyncDefault // RR_FLIP_VSYNC (nv2a_kelvin.go) — a fidelity experiment, off by default
	m.nv.reg = map[uint32]uint32{}
	m.apu = newMMIOLatch("APU")
	m.ac97 = newMMIOLatch("AC97")
	m.usb = newMMIOLatch("USB")
	m.nic = newMMIOLatch("NIC")
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
	// RDTSC comes from the machine's single timebase (sched.go guestTSC): one count
	// per instruction-clock of guest time, continuous across idle-advance jumps —
	// which retire no instructions but do pass time — and across savestates from the
	// old 2000-instrs-per-ms declaration (state.go's clock-epoch migration).
	c.TSCFunc = m.guestTSC
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
	case a >= wcBase && a < wcBase+ramSize:
		// The write-combined RAM alias: Xbox D3D hands out texture/surface
		// pointers as 0xF0000000 | physical so CPU writes bypass the cache. The
		// XMV movie player's YUV->RGB blit writes its locked texture through this
		// window — the first code to reach it.
		return a - wcBase, false, true
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
	if a >= apuBase && a < apuTop {
		return m.apuRead(a - apuBase)
	}
	if a >= ac97Base && a < ac97Top {
		return m.ac97Read(a - ac97Base)
	}
	if a >= usbBase && a < usbTop {
		return m.usbRead(a - usbBase)
	}
	if a >= nicBase && a < nicTop {
		return m.nicRead(a - nicBase)
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
	if a >= apuBase && a < apuTop {
		m.apuWrite(a-apuBase, v)
		return
	}
	if a >= ac97Base && a < ac97Top {
		m.ac97Write(a-ac97Base, v)
		return
	}
	if a >= usbBase && a < usbTop {
		m.usbWrite(a-usbBase, v)
		return
	}
	if a >= nicBase && a < nicTop {
		m.nicWrite(a-nicBase, v)
		return
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

// Poke writes a dword to guest memory the way the guest itself would — through the full
// address-window translation, MMIO side effects and all.
//
// That is the opposite choice from ReadRAM below, and deliberately so. ReadRAM refuses
// the aperture because a debugger's memory pane polls continuously at addresses the user
// merely happened to scroll past, so a side effect there is an accident. A poke is the
// other kind of act: it is one write, at one address, that someone typed on purpose. If
// they aimed it at a register, servicing the register is the answer they asked for.
//
// It exists for the oracle's -poke — a probe, never a model. A poke that makes the title
// advance has proved that the address is read; it has not earned a place in the machine.
func (m *Machine) Poke(a, v uint32) { m.write32(a, v) }

// ReadRAM fills buf from guest address a, reading RAM only: anything outside RAM — the
// NV2A aperture above all — reads as zero rather than being fetched.
//
// That restriction is the point, and it is why this exists beside Read. A debugger's
// memory pane and its disassembler poll continuously and at addresses the user picked,
// and reading an MMIO register HAS SIDE EFFECTS on this machine (the PCRTC interrupt
// status is write-1-to-clear, the FIFO status is what a spin loop is waiting on). A pane
// that quietly serviced the title's own interrupt ack would be a debugger that changes
// the bug it is being used to find.
func (m *Machine) ReadRAM(a uint32, buf []byte) {
	for i := range buf {
		buf[i] = 0
		if phys, mmio, ok := m.translate(a + uint32(i)); ok && !mmio && int(phys) < len(m.RAM) {
			buf[i] = m.RAM[phys]
		}
	}
}

// ReadCode fills buf with the bytes at a for disassembly. It is ReadRAM: code lives in
// RAM, and a disassembler walking off into the register aperture must not touch it.
func (m *Machine) ReadCode(a uint32, buf []byte) { m.ReadRAM(a, buf) }

// NVSubchannelClass is the object class currently bound to a push-buffer subchannel —
// what a command's method number has to be read against, since the method space is the
// bound object's, not the machine's.
func (m *Machine) NVSubchannelClass(subchan uint32) uint32 { return m.pgraph.subClass[subchan&7] }

// AllocStats reports the two bump arenas' current edges (heap grows up toward
// heapTop, the contiguous/pool arena grows down toward the heap) — the
// debugger/CLI view of guest memory pressure.
func (m *Machine) AllocStats() (heapNext, heapTop, poolNext uint32) {
	return m.heapNext, m.heapTop, m.poolNext
}

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
