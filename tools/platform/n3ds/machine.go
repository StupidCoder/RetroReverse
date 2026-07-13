package n3ds

// machine.go is the Nintendo 3DS oracle: an ARM11 core (tools/cpu/arm in its
// ARMv6K variant) wired to the userland memory map an application process sees,
// with the Horizon kernel's supervisor calls high-level-emulated in Go (svc.go)
// the way tools/platform/psx HLEs the PSX BIOS and tools/platform/dos HLEs DOS.
//
// What it models is the *process* view, not the console. The ExeFS/.code segments
// are loaded at the addresses the ExHeader gives; a stack, a growable heap, the
// kernel→user shared config page and a thread-local-storage page are mapped; and
// the ARM11 boots at the code entry in User mode. The svc instruction traps to the
// HLE kernel (svc.go), and SendSyncRequest into a partial Horizon service tree
// (ipc.go / ipc_services.go: srv:, APT, GSP, hid, cfg, fs, dsp). The PICA200 GPU
// is modelled (gpu*.go: LLE vertex shader, rasteriser, TEV) and so is the DSP's
// firmware protocol (dsp.go: pipes, shared-memory audio frames, the frame clock);
// the ARM9 is not. The discipline throughout: run the C runtime and the OS
// handshake, and stop *explicitly* at the first kernel facility or service
// command not yet implemented rather than diverging.
//
// Every unimplemented supervisor call or IPC command halts with its identity and
// PC, so a run's reach is always a concrete, honest statement of what works.

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// Userland virtual-address landmarks (ARM11, Horizon). These match the layout the
// kernel gives an application-type process; addresses a title's own code assumes.
const (
	codeBase = 0x00100000 // .text/.rodata/.data/.bss load here (ExHeader confirms)

	heapBase = 0x08000000 // the process heap ControlMemory(ALLOC) grows
	heapMax  = 0x0E000000

	linearBase = 0x14000000 // the LINEAR heap (physically-contiguous allocations)
	linearMax  = 0x1C000000

	stackTop = 0x10000000 // the main thread's stack grows down from here

	configPage = 0x1FF80000 // kernel→user shared "config" page (read-only)
	tlsBase    = 0x1FF82000 // the main thread's thread-local storage page
	tlsSize    = 0x1000

	pageSize = 0x1000
)

// memRegion is one mapped, byte-addressable span of the process address space.
// (The container-side "region" in ncch.go is a different, offset/size, type.)
type memRegion struct {
	name string
	base uint32
	data []byte
}

func (r *memRegion) contains(a uint32) bool {
	return a >= r.base && a-r.base < uint32(len(r.data))
}

// ReadBytes copies length bytes starting at addr out of the mapped regions (zero
// for any unmapped byte). A bring-up instrument for inspecting a structure found
// in a loaded snapshot.
func (m *Machine) ReadBytes(addr, length uint32) []byte {
	b := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b[i] = m.Read(addr + i)
	}
	return b
}

// FindBytes scans every mapped region for the byte pattern and returns the
// virtual addresses where it occurs (a bring-up instrument for locating a known
// string or structure in a loaded snapshot). Capped so a common pattern does not
// flood.
func (m *Machine) FindBytes(pat []byte) []uint32 {
	var hits []uint32
	if len(pat) == 0 {
		return hits
	}
	for _, r := range m.regions {
		d := r.data
		for i := 0; i+len(pat) <= len(d); i++ {
			if d[i] == pat[0] && string(d[i:i+len(pat)]) == string(pat) {
				hits = append(hits, r.base+uint32(i))
				if len(hits) >= 256 {
					return hits
				}
			}
		}
	}
	return hits
}

// Machine is the 3DS process-level oracle.
type Machine struct {
	CPU      *arm.CPU
	regions  []*memRegion
	codeReg  *memRegion
	stackReg *memRegion
	tlsReg   *memRegion

	entry uint32 // code entry point (PC at start)

	// Heap bump allocators for ControlMemory.
	heapPtr   uint32
	linearPtr uint32
	heapReg   *memRegion
	linearReg *memRegion

	// Kernel-object bookkeeping for the HLE (svc.go).
	handles    map[uint32]*kobject
	nextHandle uint32
	ports      map[uint32]string // connected port handles → service name
	services   map[uint32]string // srv:-acquired service handles → service name
	tick       uint64            // GetSystemTick counter (advanced per step; JUMPED by idle sleeps)
	instrs     uint64            // machine-monotonic retired instructions. CPU.Instrs is
	// per-THREAD (switchTo does *m.CPU = t.ctx, so it moves backward on a
	// context switch); anything pacing the machine — the VBlank heartbeat, GX
	// completion deadlines — must ride this instead. Unlike tick it is never
	// jumped by sleep fast-forward, only by the deliberate idle-frame skip.

	// Threads and the cooperative scheduler (thread.go). curThread's live state
	// is in CPU; every other thread's state lives in its saved ctx. The main
	// thread is created in NewMachine.
	threads    []*thread
	curThread  *thread
	nextThread uint32 // thread id counter
	nextTLS    uint32 // next per-thread TLS page base to hand out
	rrCursor   int    // round-robin cursor into threads
	reschedule bool   // an svc asks the scheduler to switch after this Step
	stopped    bool   // a breakpoint asked the run loop to stop

	// Filesystem (fs.go): the cartridge RomFS (parsed + the raw IVFC region a
	// game opens to parse itself), and the open IFile sessions.
	romfs          *RomFS
	romfsRaw       []byte
	fsFiles        map[uint32]*fsFile
	fsDirs         map[uint32]*fsDir
	fsArchives     map[uint32]uint32 // archive handle → archive ID (OpenArchive)
	saveFiles      map[string][]byte // the writable (save-data) archive's contents
	saveFormatted  bool              // FormatSaveData (0x084C) seen; gates GetFormatInfo
	saveFormatInfo [4]uint32         // recorded format: size blocks, dirs, files, dup-data

	// The PICA200 GPU (gpu.go). Accessor: GPU().
	gpu *GPU

	// The DSP audio coprocessor (dsp.go / dsp_voice.go): the pipe protocol, the
	// shared-memory audio-frame exchange, the frame clock the game's sound thread
	// blocks on, and the 24-voice mixer behind it.
	dsp dspHLE

	// Audio capture (instrumentation, not machine state): when AudioCapture is
	// set, every audio frame's final stereo mix is appended to AudioPCM, which
	// bootoracle -wav writes out. Deliberately outside the savestate — it is a
	// recording of a run, not a property of the machine.
	AudioCapture bool
	AudioPCM     []int16

	// DSPTrace logs every source configuration the DSP consumes and every status
	// it publishes (bootoracle -dsptrace) — the instrument for the app↔DSP voice
	// conversation, which is otherwise invisible: it happens entirely in shared
	// memory, with no IPC to log.
	DSPTrace bool

	// IPC / graphics bring-up.
	notifyWaiters    []uint32 // thread ids parked in srv: ReceiveNotification
	aptNotifyEv      uint32   // events APT Initialize returned; the deferred wake signals them
	aptResumeEv      uint32
	aptWakePending   bool           // NotifyToWait seen; signal the APT events at the next VBlank
	aptParams        []aptParam     // queued applet answers ReceiveParameter delivers in order
	gxPending        []gxPendingCmd // accepted GX commands awaiting their completion deadline
	ipcLog           []ipcCall
	gspShared        uint32            // the GSP shared-memory block handle, once registered
	gspSharedAddr    uint32            // where the game mapped the GSP shared memory
	gspEvent         uint32            // event the GSP signals on each interrupt (VBlank)
	hidShared        uint32            // the HID shared-memory block handle (pad/touch/accel state)
	hidSharedAddr    uint32            // where the game mapped the HID shared memory
	hidEvents        []uint32          // HID interrupt events (pad0..pad1, accel, gyro, debugpad, touch)
	hidReadHist      map[uint32]int    // instrumentation: HID-shared read histogram by offset
	hidReadPC        map[uint32]uint32 // instrumentation: last PC that read each HID offset
	HidTrace         bool              // when set, tally reads within the HID shared block
	hidButtons       uint32            // buttons to publish as held in the HID shared memory (SetKeys)
	hidPrevButtons   uint32            // last frame's mask, for press/release edge computation
	hidRingIdx       uint32            // current HID sample-ring index (0..7)
	HidPulse         int               // if >0, release the injected buttons briefly every N frames (fresh edges)
	nextFrameInstr   uint64            // instruction count at which to deliver the next VBlank
	vblankCount      uint64            // VBlanks delivered
	framesSubmitted  int               // GSP TriggerCmdReqQueue calls (GPU command lists)
	framesSwapped    int               // framebuffer-info entries consumed at VBlank (frames presented)
	displayTransfers int               // GX DisplayTransfers executed (frames made visible)
	lastXferTop      xferRecord
	lastXferBottom   xferRecord
	screenFB         [2]fbPresent // last framebuffer-info entry consumed per screen (top, bottom)

	// Debugger hooks (debug.go). All nil/zero unless a debugger installs them, and
	// all deliberately outside the savestate: they observe a run, they are not a
	// property of the machine.
	OnPICACmd     func(w PICAWrite)                // each command-list register write, before it executes
	OnPixel       func(x, y uint32, ev PixelEvent) // each fragment the rasteriser produced, kept or killed
	OnFrame       func(m *Machine)                 // each VBlank, after the buffer swap is consumed
	StopRequested bool                             // a hook asking Run to return at the next opportunity

	// Memory watch windows. The ARM bus is byte-granular (the core issues four
	// Read calls for one LDR), so these fire per byte, with the PC that issued the
	// access — which is the question a watch is asked to answer.
	WatchLo, WatchHi   uint32
	OnWrite            func(addr, val, pc uint32)
	RWatchLo, RWatchHi uint32
	OnRead             func(addr, val, pc uint32)

	// picaLimit > 0 stops the run once picaCount command-list writes have
	// executed — the command scrubber's halt (RunStopAfterPICACommand).
	picaLimit int
	picaCount int

	// Per-subsystem frame timing (profile.go). Off by default; a run that does not
	// ask for it pays one predictable branch per coarse boundary. Outside the
	// savestate: it measures a run, it is not a property of the machine.
	Profile bool
	prof    profState

	// Instrumentation.
	GXCapture  bool       // record GX commands + ProcessCommandList buffers (gx.go)
	gxLog      []GXRecord // the captured commands, in submission order
	Trace      bool
	traceN     int
	traceMax   int
	bps        map[uint32]bool
	logpcs     map[uint32]bool
	tracefroms map[uint32]bool
	watches    []watch
	svcLog     []svcEvent // every supervisor call, in order
	debugOut   []byte     // svcOutputDebugString text
	Verbose    bool

	// Metadata pinned into a savestate so it cannot resume into another title.
	programID uint64
}

// kobject is a high-level-emulated kernel waitable object. Its semantics are real
// (sync.go): a thread that waits on an unavailable object blocks and is woken when
// the object is signalled. The fields used depend on kind:
//
//	event     — signal (set/clear), manualReset (sticky vs auto-clear on wake)
//	timer     — signal (fired)
//	mutex     — mutexOwner (holder thread id, 0 = free) + mutexDepth (recursion)
//	semaphore — semCount (available permits)
//	thread    — signal set true when the thread exits (WaitSync on a thread handle)
//	memblock  — blockAddr/blockSize (the backing region, Phase 3)
//
// waiters holds the ids of threads blocked on this object.
type kobject struct {
	kind        string
	name        string
	signal      bool
	manualReset bool
	semCount    int32
	mutexOwner  uint32
	mutexDepth  int
	waiters     []uint32
	blockAddr   uint32
	blockSize   uint32
	thread      *thread // for kind=="thread": the thread this handle names
}

// aptParam is a parameter queued for the app to receive — a library applet's
// answer this HLE fabricates, since applets are not run. Fields are exported
// for the savestate gob encoder.
type aptParam struct {
	Sender  uint32 // applet id the answer claims to come from
	Command uint32 // parameter command the app's receive loops dispatch on
	Handle  uint32 // shared-memory block handle carried along (0 = none)
	Data    []byte // payload copied into the receiver's static buffer
}

type watch struct {
	addr uint32
	len  uint32
	last map[uint32]uint32 // last observed value per word
	seen map[uint32]bool
}

// svcEvent records one supervisor call. Fields are exported so a snapshot's gob
// encoder captures the log.
type svcEvent struct {
	PC   uint32
	Num  uint32
	Name string
	Args [4]uint32
}

// NewMachine builds a 3DS oracle from a decrypted CCI image: it parses the
// application NCCH, loads and (if flagged) decompresses .code into the code
// region, lays out the stack/heap/config/TLS, and points the ARM11 at the entry.
func NewMachine(img []byte) (*Machine, error) {
	ncsd, err := ParseNCSD(img)
	if err != nil {
		return nil, err
	}
	cxi, err := ncsd.Executable()
	if err != nil {
		return nil, err
	}
	if cxi.Encrypted() {
		return nil, fmt.Errorf("n3ds: partition 0 is encrypted (%s); supply a decrypted dump", cxi.CryptoMethod())
	}
	ex, err := cxi.ExHeader()
	if err != nil {
		return nil, err
	}
	efs, err := cxi.ExeFS()
	if err != nil {
		return nil, err
	}
	code, err := efs.Code(ex)
	if err != nil {
		return nil, err
	}

	m := &Machine{
		handles:    map[uint32]*kobject{},
		nextHandle: 0x00010000,
		ports:      map[uint32]string{},
		services:   map[uint32]string{},
		fsFiles:    map[uint32]*fsFile{},
		fsDirs:     map[uint32]*fsDir{},
		fsArchives: map[uint32]uint32{},
		saveFiles:  map[string][]byte{},
		bps:        map[uint32]bool{},
		programID:  cxi.ProgramID,
		entry:      ex.Text.Address,
	}
	m.gpu = newGPU(m)
	// The RomFS backs the fs service (fs.go). A title without one still boots;
	// its file opens simply miss. The raw IVFC region is also kept: a game opens
	// ARCHIVE_ROMFS as one big file and walks the filesystem itself.
	m.romfs, _ = cxi.RomFS()
	m.romfsRaw, _ = cxi.RomFSBytes()

	// Code region: text..bss as one contiguous span (the segments tile without
	// gaps; validate() in the ExHeader proved it). BSS is the zero tail past the
	// loaded bytes, so size the span to include it.
	total := ex.CodeSize() + ex.BSSSize
	codeMem := make([]byte, total)
	copy(codeMem, code)
	m.codeReg = m.mapRegion("code", ex.Text.Address, codeMem)

	// Stack (grows down from stackTop), rounded to a page.
	stackSize := (ex.StackSize + pageSize - 1) &^ (pageSize - 1)
	if stackSize == 0 {
		stackSize = 0x4000
	}
	m.stackReg = m.mapRegion("stack", stackTop-stackSize, make([]byte, stackSize))

	// Heap and linear-heap regions grow on demand via ControlMemory; start empty.
	m.heapPtr, m.linearPtr = heapBase, linearBase
	m.heapReg = m.mapRegion("heap", heapBase, nil)
	m.linearReg = m.mapRegion("linear", linearBase, nil)

	// Config (shared) page and the main thread's TLS.
	m.mapRegion("config", configPage, m.buildConfigPage())
	m.tlsReg = m.mapRegion("tls", tlsBase, make([]byte, tlsSize))

	// VRAM, at the fixed virtual window an application sees it through. The
	// game DMAs its vertex/texture data here and points its render targets at
	// it (as physical 0x18000000 addresses in the GPU registers; gsp_mem.go
	// translates).
	m.mapRegion("vram", vramVirtBase, make([]byte, vramSize))

	// DSP RAM, likewise at its fixed window: the second half holds the two
	// shared-memory regions the app and the DSP exchange audio frames through
	// (dsp.go); ConvertProcessAddressFromDspDram hands the game pointers here.
	m.mapRegion("dspram", dspRAMBase, make([]byte, dspRAMSize))
	m.dsp.IntEvents = map[uint32]uint32{}

	// Boot the core: ARMv6K, User mode, PC at the entry, SP at the stack top.
	cpu := arm.NewCPU(m)
	cpu.Arch = arm.V6K
	cpu.Reset()
	cpu.Mode = arm.ModeSYS // a flat, unbanked user-ish mode for the HLE
	cpu.R[13] = stackTop
	cpu.R[15] = m.entry
	cpu.SWI = m.handleSVC
	cpu.Coproc = m.handleCP15
	m.CPU = cpu

	// The main thread: its live state is the CPU we just set up, and it uses the
	// pre-mapped TLS page. Per-thread TLS for children is handed out above it.
	m.nextTLS = tlsBase + tlsSize
	main := &thread{
		id: 1, handle: m.newHandle("thread", false), tlsBase: tlsBase,
		priority: 0x30, state: ready,
	}
	main.ctx = *cpu
	m.nextThread = 2
	m.threads = []*thread{main}
	m.curThread = main
	m.handles[main.handle].thread = main

	return m, nil
}

// mapRegion adds a region and returns it. A nil data slice makes an initially
// empty region whose backing store the heap allocators grow.
func (m *Machine) mapRegion(name string, base uint32, data []byte) *memRegion {
	r := &memRegion{name: name, base: base, data: data}
	m.regions = append(m.regions, r)
	return r
}

// buildConfigPage fills the handful of kernel→user config fields a runtime reads
// early. Most of the page is zero; the fields set here are the ones observed to
// matter (kernel version and the application memory-type/limit), left minimal and
// documented rather than dumped from a console.
func (m *Machine) buildConfigPage() []byte {
	p := make([]byte, pageSize)
	// 0x0000: kernel version (major.minor packed); report a late system version.
	binary.LittleEndian.PutUint32(p[0x00:], (2<<24)|(50<<16))
	// 0x0004: update flag / 0x0008: ns tid — left zero.
	// 0x0040: APPMEMALLOC — total application memory. The runtime derives its heap
	// size as (APPMEMALLOC − the COMMIT resource limit); the two must differ by a
	// sensible, page-aligned amount, which svcGetResourceLimitValues arranges.
	// Traced from the heap setup at 0x00100750, which reads config+0x40.
	binary.LittleEndian.PutUint32(p[0x40:], appMemAlloc)
	return p
}

// appMemAlloc is the total memory the application is granted. The heap is what
// remains after the committed base (committedBytes), so appMemAlloc −
// committedBytes is the initial heap size the runtime allocates.
const appMemAlloc = 0x04000000 // 64 MiB

// --- arm.Bus ---------------------------------------------------------------

func (m *Machine) regionOf(a uint32) *memRegion {
	// Small region count; linear scan is fine and keeps the map order-independent.
	for _, r := range m.regions {
		if r.contains(a) {
			return r
		}
	}
	return nil
}

func (m *Machine) Read(a uint32) byte {
	if m.HidTrace && m.hidSharedAddr != 0 && a >= m.hidSharedAddr && a < m.hidSharedAddr+0x1000 {
		if m.hidReadHist == nil {
			m.hidReadHist = map[uint32]int{}
			m.hidReadPC = map[uint32]uint32{}
		}
		off := (a - m.hidSharedAddr) &^ 3
		m.hidReadHist[off]++
		m.hidReadPC[off] = m.CPU.PC()
	}
	if r := m.regionOf(a); r != nil {
		v := r.data[a-r.base]
		if m.OnRead != nil && a >= m.RWatchLo && a < m.RWatchHi {
			m.OnRead(a, uint32(v), m.CPU.PC())
		}
		return v
	}
	if m.Verbose {
		fmt.Printf("  [unmapped read  0x%08X pc=0x%08X]\n", a, m.CPU.PC())
	}
	return 0
}

func (m *Machine) Write(a uint32, v byte) {
	if r := m.regionOf(a); r != nil {
		r.data[a-r.base] = v
		if m.OnWrite != nil && a >= m.WatchLo && a < m.WatchHi {
			m.OnWrite(a, uint32(v), m.CPU.PC())
		}
		return
	}
	if m.Verbose {
		fmt.Printf("  [unmapped write 0x%08X=0x%02X pc=0x%08X]\n", a, v, m.CPU.PC())
	}
}

// ReadWord/WriteWord are little-endian helpers the HLE uses to touch argument
// structures without going byte by byte.
func (m *Machine) ReadWord(a uint32) uint32 {
	return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 | uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
}

func (m *Machine) WriteWord(a, v uint32) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
	m.Write(a+2, byte(v>>16))
	m.Write(a+3, byte(v>>24))
}
