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
// (ipc.go / ipc_services.go: srv:, APT, GSP, hid, cfg, fs). It deliberately does
// not model the PICA200 GPU, the DSP or the ARM9 — turning the game's GPU command
// lists into pixels is a large separate effort — so this is a bring-up machine:
// run the C runtime and the OS handshake, and stop *explicitly* at the first
// kernel facility or service command not yet implemented rather than diverging.
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
	tick       uint64            // GetSystemTick counter (advanced per step)
	tpidr      uint32            // TPIDRURW/PRW writable per-thread scratch (CP15)

	// IPC / graphics bring-up.
	ipcLog          []ipcCall
	gspShared       uint32 // the GSP shared-memory block handle, once registered
	framesSubmitted int    // GSP TriggerCmdReqQueue calls (GPU command lists)
	framesSwapped   int    // GSP SetBufferSwap calls (frames presented)

	// Instrumentation.
	Trace     bool
	traceN    int
	traceMax  int
	bps       map[uint32]bool
	watches   []watch
	svcLog    []svcEvent // every supervisor call, in order
	debugOut  []byte     // svcOutputDebugString text
	Verbose   bool

	// Metadata pinned into a savestate so it cannot resume into another title.
	programID uint64
}

// kobject is a high-level-emulated kernel object (thread, event, mutex, …). The
// HLE does not model their semantics; it hands out handles and, for the waitable
// ones, reports them signalled so a runtime's init does not deadlock. Which
// objects are stubbed this way is spelled out in svc.go.
type kobject struct {
	kind   string
	name   string
	signal bool
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
		bps:        map[uint32]bool{},
		programID:  cxi.ProgramID,
		entry:      ex.Text.Address,
	}

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
	if r := m.regionOf(a); r != nil {
		return r.data[a-r.base]
	}
	if m.Verbose {
		fmt.Printf("  [unmapped read  0x%08X pc=0x%08X]\n", a, m.CPU.PC())
	}
	return 0
}

func (m *Machine) Write(a uint32, v byte) {
	if r := m.regionOf(a); r != nil {
		r.data[a-r.base] = v
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
