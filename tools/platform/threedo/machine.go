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

	"retroreverse.com/tools/cpu/arm60"
)

const (
	dramSize    = 2 * 1024 * 1024 // 2 MiB main DRAM at 0x00000000
	vramBase    = 0x00200000
	vramSize    = 1024 * 1024 // 1 MiB VRAM
	madamBase   = 0x03300000  // Madam (CEL/matrix/DMA) registers
	madamEnd    = 0x03400000
	clioBase    = 0x03400000 // Clio (video/audio/timers/IRQ) registers
	clioEnd     = 0x03500000
	kernelBase  = 0x00180000 // synthetic Portfolio kernel/folio base (r7/r9)
	hleBase     = 0x0FE00000 // reserved window: a PC here is an intercepted folio call
	hleSize     = 0x00010000
	bootTaskNum = 1 // item number of the initial (boot) task

	// The File folio is a second folio, called through its own negative-offset
	// vector table (LookupItem of the "File" folio item yields its base). Its
	// vectors point into a distinct slice of the HLE window (hleBase+hleFileTag+N)
	// so an intercepted call can be told apart from a kernel-folio call.
	fileFolioBase  = 0x0017F000 // base returned by LookupItem("File" folio)
	hleFileTag     = 0x8000     // File-folio calls trap at hleBase+hleFileTag+offset
	otherFolioBase = 0x0017E800 // base for any other (not-yet-implemented) folio
	hleOtherTag    = 0xA000     // other-folio calls trap at hleBase+hleOtherTag+offset
	gfxFolioBase   = 0x0017E000 // base returned by LookupItem("Graphics" folio)
	hleGfxTag      = 0xC000     // Graphics-folio calls trap at hleBase+hleGfxTag+offset

	// The folio vector tables sit just below the kernel base (0x17E000..0x180000);
	// the AllocMem pool is below them and the boot stack (SP near the top of DRAM)
	// grows down through the ~0.5 MiB above, so a deep boot's stack never reaches
	// the tables. (A higher kernelBase left too little stack headroom and a deep
	// boot's stack corrupted the vectors.)
	// The OS "hardware context" struct the game reaches through the kernel base:
	// `[[0x3E1F4]+0x98]` points here. The game reads the current VBlank/field count
	// from +0xA (AddTimer), a device-ish field at +0x18 (timer setup) and the OS
	// MemList at +0xA8 (AllocMem's list arg, which our HLE ignores). We plant a real
	// pointer to this struct and advance the VBlank count so the game's timing loops
	// (the VBlank service task) see time passing.
	// [kernelBase+0x98] is kb_CurrentTask: a pointer to the running task's Task
	// struct. The HLE keeps one synthetic Task struct that always describes the
	// current task (updated on every switch) — the fields programs read off it:
	osCtxBase   = 0x0017D000 // the Task struct kb_CurrentTask points at
	osCtxPri    = 0x0A       // +0x0A: ItemNode n_Priority (threads inherit it)
	osCtxItem   = 0x18       // +0x18: ItemNode n_Item — the current task's Item
	osCtxMemLst = 0xA8       // +0xA8: t_FreeMemoryLists (ignored by our AllocMem)

	// The game sets up TWO memory managers (InitMemMgr, id 0xC8): one at boot entry
	// and one from a later dispatch. Both share one memlist array (@game 0x5BA68)
	// AND one block-node pool ([0x42368]); each manager's setup calls InitMemList
	// (which *replaces* the array with its own region) and rebuilds the node pool
	// at its region base. So the second manager wipes the first's block tracking,
	// and a later free of a first-manager block (e.g. the default screen buffer
	// 0x827c0) misses findmemblock and fatally calls the game's exit handler.
	//
	// This is the boot's current hard blocker and is NOT yet solved. A previous
	// attempt shrank this pool so manager 1's DRAM alloc returned 0 (skipping the
	// wipe); that was WRONG — the game then builds its node pool at base 0, whose
	// 0x28-strided .next writes corrupt the low-memory code (exception vectors,
	// folio trampolines, the 0x4C4 divide routine), so large-number divides loop
	// forever (proven: divide_test passes in isolation; an in-game trace showed
	// 0x4D0 overwritten). Keep manager 1's region VALID (non-zero) to avoid that
	// corruption; the real fix is a faithful OS MemList so manager 1's allocation
	// resolves the way it does on hardware (a valid region that doesn't orphan the
	// first manager's normal-path frees) — see the game's InitMemMgr flow.
	dheapBase = 0x00080000 // DRAM AllocMem pool: above the image + BSS
	dheapTop  = 0x0017D000 // ends below the OS context struct + folio vector tables
	vheapBase   = vramBase
	vramReserve = 0x4000 // reserve a VRAM sliver (over-commit workaround, see git log)
	vheapTop    = vramBase + vramSize - vramReserve
	// osCtx (0x17D000) and the folio vector tables (0x17E000..0x180000) sit above
	// the DRAM pool; the boot stack grows down from near the top of DRAM, clear of
	// both. (A higher kernelBase left too little stack headroom and a deep boot's
	// stack corrupted the vectors.)
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
	// Two allocation pools, matching the 3DO's split memory: AllocMem's flags
	// select VRAM (MEMTYPE_VRAM 0x10000) or DRAM (MEMTYPE_DRAM 0x80000, or ANY).
	// The game keeps them separate — its startup allocates a big DRAM working set,
	// then binary-searches VRAM for framebuffers — so a single heap corrupts its
	// bookkeeping.
	dheap *heap // DRAM pool
	vheap *heap // VRAM pool
	CPU   *arm60.CPU

	vol     *Volume                // the mounted disc, so the I/O HLE can read files (io.go)
	streams map[uint32]*diskStream // open File-folio streams, keyed by handle (filefolio.go)
	dirs    map[uint32]*dirScan    // open File-folio directory scans (filefolio.go)
	nvram   map[string][]byte      // battery-backed store, empty like a fresh console (filefolio.go)

	// Graphics folio state (graphicsfolio.go).
	bitmaps    map[int32]gfxBitmap // Bitmap item -> pixel buffer geometry
	screenBM   map[int32]int32     // Screen item -> its (first) Bitmap item
	displayBuf uint32              // the front buffer DisplayScreen last showed

	// Event-broker state (io.go): ports that connected as listeners, and a
	// step-scheduled control-pad script the run loop feeds to SendPadEvent.
	ebListeners []int32
	PadScript   []PadStep

	// CelDebug, when set, records a one-line summary of every cel DrawCels
	// renders (graphicsfolio.go) into CelDebugLog — for diagnosing what lands
	// on screen and what silently drops.
	CelDebug    bool
	CelDebugLog []string // cels drawn since the last DisplayScreen
	CelFrameLog []string // the last fully-displayed frame's cels
	SportDebug  bool     // log the full IOInfo of every SPORT request
	PerspTint   bool     // paint perspective (corner-engine) cels solid magenta

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi uint32
	OnWrite          func(addr, val, pc uint32)
	OnStep           func(m *Machine, pc uint32)

	// Kernel item system (kernel.go).
	items      map[int32]*item
	itemByType map[uint32]*item
	nextItem   int32

	// Cooperative task scheduler (task.go).
	tasks        []*task
	cur          int  // index of the running task
	switches     int  // context-switch count
	needSchedule bool // a WaitSignal/exit asked to switch after this instruction

	KernelCalls []KernelCall
	SWICalls    []KernelCall // SWI kernel calls (Offset = SWI comment)
	// SpinBreak, when set, lets the run loop poke past flag spin-waits (an
	// exploration aid — it advances the PC but not real OS state, so downstream
	// code that needed the awaited result runs on uninitialised data). Off by
	// default, so a plain run stops honestly at the first async-wait frontier.
	SpinBreak  bool
	SpinBreaks int
	// StallTolerance scales the run loop's deadlock threshold (default 1). The
	// guard calls a run dead after N fruitless task switches; a program whose
	// main loop legitimately re-executes only already-seen code (a settled frame
	// loop) needs a higher tolerance to keep running.
	StallTolerance int
	// NoStreams makes .stream movie files fail to resolve, skipping FMV playback
	// through the game's own missing-file fallback (see loadDiscFile). Set it
	// until the audio folio + DataStreamer are modelled well enough to play them.
	NoStreams bool
	simTime    uint64 // virtual microsecond clock (folio SampleSystemTimeTT)
	vblank     uint32 // virtual VBlank/field counter (osCtx +0xA)
	// vblMirror, if non-zero, is a game global the VBL manager must keep in step
	// with the elapsed-field count. On real hardware the graphics folio's VBL
	// interrupt increments a monotonic field counter that the game's frame/timer
	// loops read; we have no interrupts, so the HLE's virtual VBL manager writes
	// the same count here each field. The address is game-specific (the OS folio
	// writes wherever the program registered its counter), so the program's oracle
	// configures it via SetVBLMirror; the mechanism (advance every field) is generic.
	vblMirror uint32
	tty        []byte
	Log        []string
	logSeen    map[string]bool

	Halted     bool
	HaltReason string
}

// NewMachine builds a reset 3DO machine with DRAM, VRAM and an ARM60 core.
func NewMachine() *Machine {
	m := &Machine{
		dram:       make([]byte, dramSize),
		vram:       make([]byte, vramSize),
		dheap:      newHeap(dheapBase, dheapTop-dheapBase),
		vheap:      newHeap(vheapBase, vheapTop-vheapBase),
		items:      map[int32]*item{},
		itemByType: map[uint32]*item{},
		nextItem:   0x1000,
		logSeen:    map[string]bool{},
		streams:    map[uint32]*diskStream{},
		dirs:       map[uint32]*dirScan{},
		nvram:      map[string][]byte{},
		bitmaps:    map[int32]gfxBitmap{},
		screenBM:   map[int32]int32{},
	}
	m.CPU = arm60.NewCPU(m)
	m.CPU.SWI = m.swi
	m.initTasks()
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
	// Plant the File folio's vector table the same way, into its own HLE slice,
	// and a generic table for any other folio the game looks up.
	for off := uint32(4); off <= 0x100; off += 4 {
		m.writeWord(fileFolioBase-off, hleBase+hleFileTag+off)
		m.writeWord(otherFolioBase-off, hleBase+hleOtherTag+off)
		m.writeWord(gfxFolioBase-off, hleBase+hleGfxTag+off)
	}

	// Plant the current-task struct and point [kernelBase+0x98] (kb_CurrentTask)
	// at it. Programs read their own item number (n_Item — e.g. to tell a spawned
	// thread whom to signal back), their priority (threads inherit it), and the
	// task MemList pointer (just has to be non-zero; our AllocMem ignores it).
	m.writeWord(kernelBase+0x98, osCtxBase)
	m.dram[osCtxBase+osCtxPri] = 100
	m.writeWord(osCtxBase+osCtxItem, uint32(m.curTask().num))
	m.writeWord(osCtxBase+osCtxMemLst, osCtxBase+0x400)

	m.CPU.SetReg(5, 0)          // r5: argc-like
	m.CPU.SetReg(6, 0)          // r6: argv-like
	m.CPU.SetReg(7, kernelBase) // r7: kernel/folio base (entry copies it to r9)
	m.CPU.SetReg(13, dramSize-0x1000)
	m.CPU.SetPC(a.ImageBase)
}

// SetVolume mounts a disc image so the I/O HLE can read files from it.
func (m *Machine) SetVolume(v *Volume) { m.vol = v }

// DisplayBuffer returns the frame buffer the program last put on screen via
// DisplayScreen (0 if no screen has been shown yet).
func (m *Machine) DisplayBuffer() uint32 { return m.displayBuf }

// SetVBLMirror registers a game global that the virtual VBL manager keeps equal
// to the elapsed-field count (see the vblMirror field). Programs read the folio's
// monotonic VBlank counter out of their own BSS; the OS folio fills it from the
// VBL interrupt, which the HLE stands in for by writing the count each field.
func (m *Machine) SetVBLMirror(addr uint32) { m.vblMirror = addr }

// advanceVBlank moves the virtual VBlank/field clock forward by n fields and
// publishes the elapsed time to the program's registered field counter
// (SetVBLMirror), standing in for the graphics folio's VBL interrupt. It is
// driven both by a steady background tick (run loop) and by field-wait timer IOs
// (io.go), so field waits actually advance time.
//
// The counter's units are the game's own contract, read off its VBL service
// loop: each field it adds a field's worth (0x64) to an accumulator and
// subtracts the counter value, dispatching its per-field callbacks while the
// accumulator stays non-negative — so the counter must read ~100 per field
// (centifields since the last dispatch), NOT a monotonic total. A growing
// total drives the accumulator to minus infinity and the callbacks (frame
// flip, the race's frame-completion flag) never run again.
func (m *Machine) advanceVBlank(n uint32) {
	if n == 0 {
		n = 1
	}
	m.vblank += n
	if m.vblMirror != 0 {
		m.writeWord(m.vblMirror, n*100)
	}
	// GrafBase's gf_VBLNumber (+0x74): the folio's monotonic field counter,
	// which programs read straight off the folio base for frame timing (the
	// race's wait-until-field loops poll it).
	m.writeWord(gfxFolioBase+0x74, m.vblank)
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
	case addr >= 0xFFFFF000:
		// A folio call through a *null* base — `LDR pc, [0, #-N]` — reads near the
		// top of the address space. Return hleBase+N so the load lands in the HLE
		// trap window with the folio's offset, making a call through an
		// uninitialised folio base resolve like any other. (Without this the game
		// reads 0, jumps to 0, and re-runs the AIF header — the boot retry loop.)
		wbase := addr &^ 3
		w := hleBase - wbase // uint32 wrap: hleBase + (2^32 - wbase)
		return byte(w >> uint((3-(addr&3))*8))
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

// breakSpin tries to unstick a flag spin-wait. Only a load whose value is then
// compared against zero within the loop is treated as a flag; its target byte is
// set to 1 (the minimal non-zero, to avoid corrupting a value used as anything
// but a boolean), so a `while ([base+imm] == 0)` busy-wait — one an interrupt
// would normally end — falls through. Returns the addresses it poked. This is the
// oracle standing in for the interrupt/task that would set the flag.
// isFlagSpin reports whether the recent loop body contains a compare-against-zero
// (TST/TEQ/CMP rX,#0) — the signature of a busy-wait on a flag, as opposed to a
// working loop (memset, search) that just happens to revisit addresses. Used to
// decide when to switch tasks or poke.
func (m *Machine) isFlagSpin(pcs []uint32) bool {
	for _, pc := range pcs {
		if pc+4 > dramSize {
			continue
		}
		w := be32(m.dram[pc:])
		op := (w >> 21) & 0xF
		if (w>>26)&3 == 0 && (w>>20)&1 == 1 && (op == 0x8 || op == 0x9 || op == 0xA) && (w>>25)&1 == 1 && w&0xFFF == 0 {
			return true
		}
	}
	return false
}

func (m *Machine) breakSpin(pcs []uint32) []uint32 {
	if !m.isFlagSpin(pcs) {
		return nil
	}

	var poked []uint32
	done := map[uint32]bool{}
	for _, pc := range pcs {
		if done[pc] || pc+4 > dramSize {
			continue
		}
		done[pc] = true
		w := be32(m.dram[pc:])
		if (w>>26)&3 != 1 || (w>>25)&1 != 0 || (w>>20)&1 != 1 { // LDR/LDRB rd,[rn,#imm]
			continue
		}
		base := m.CPU.Reg((w >> 16) & 0xF)
		imm := w & 0xFFF
		addr := base
		if (w>>24)&1 == 1 {
			if (w>>23)&1 == 1 {
				addr = base + imm
			} else {
				addr = base - imm
			}
		}
		switch {
		case addr < dramSize:
			m.dram[addr] = 1
			poked = append(poked, addr)
		case addr >= vramBase && addr < vramBase+vramSize:
			m.vram[addr-vramBase] = 1
			poked = append(poked, addr)
		}
	}
	return poked
}

// swi services the ARM SWI gate. The AIF exit SWI (#0x11) stops the machine;
// other SWIs are the Portfolio kernel calls, logged with their arguments.
func (m *Machine) swi(c *arm60.CPU, comment uint32) bool {
	if comment == 0x11 { // program exit
		m.Halted, m.HaltReason = true, "program exit (SWI #0x11)"
		return true
	}
	m.SWICalls = append(m.SWICalls, KernelCall{
		Offset: comment,
		From:   c.CurPC(),
		Args:   [4]uint32{c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3)},
	})
	if !m.kernelSWI(c, comment) && !m.fileFolioSWI(c, comment) && !m.mathFolioSWI(c, comment) {
		m.note(fmt.Sprintf("SWI #0x%X (stub)", comment))
	}
	return true // serviced: do not vector to 0x08
}
