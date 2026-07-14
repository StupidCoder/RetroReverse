package ps2

// machine.go is the PlayStation 2 machine model: 32 MiB of main memory, the
// scratchpad, the memory-mapped registers of the peripherals, and the Emotion
// Engine wired to all of it.
//
// The Machine *is* the CPU's Bus — Read/Write/Read32/Write32 decode the physical
// address map — and it drives the core itself in Run (run.go). That is the shape
// every platform in this repository uses.
//
// Two things about this machine are not like the others.
//
// The BIOS. A PS2 boots from a 4 MiB ROM holding the EE kernel, and that ROM is not
// on the game disc; deriving it from the disc is impossible, and taking it from
// elsewhere is what the clean-room rule forbids. So the kernel is high-level
// emulated: its syscalls are Go functions (kernel.go), and the machine also has to
// stand in for what the kernel *did* before it handed over — chiefly installing the
// TLB entries that map main memory. See mapMemory.
//
// The IOP. The second processor is faked at the SIF boundary for now (sif.go), not
// executed. It is a MIPS R3000A, which tools/cpu/mips already models, so running the
// real IOP modules is a later phase rather than a rewrite.

import (
	"bytes"
	"fmt"
	"sort"

	"retroreverse.com/tools/cpu/r5900"
	"retroreverse.com/tools/lib/iso9660"
)

// The physical address map.
const (
	ramBase = 0x00000000
	ramSize = 32 << 20 // 32 MiB of main memory

	// The scratchpad answers at its virtual address, 0x70000000, with no TLB entry
	// behind it — the CPU's Translate returns it unchanged, so the bus sees it here.
	spramBase = 0x70000000
	spramSize = 0x4000 // 16 KiB

	// The peripheral registers: timers, the DMA controller, the two vector-unit
	// interfaces, the GIF, the image processor and the SIF, all inside one page.
	ioBase = 0x10000000
	ioEnd  = 0x10010000

	// The vector units' instruction and data memories.
	vuBase = 0x11000000
	vuEnd  = 0x11010000

	// The GS's privileged registers — the ones the EE writes directly rather than
	// through the GIF.
	gsRegBase = 0x12000000
	gsRegEnd  = 0x12002000

	// The IOP's 2 MiB of RAM, which the EE can see directly.
	iopRAMBase = 0x1C000000
	iopRAMSize = 2 << 20
)

// stepsPerVBlank paces the synthetic vertical blank. The EE runs at ~294 MHz and a
// frame is 1/60 s, so the true figure is nearer 4.9 million; this model runs the
// interpreter far slower than real time, and a smaller quantum keeps a boot that
// waits on VBlank from spending its whole step budget inside one frame.
const stepsPerVBlank = 1_000_000

// Machine is a PlayStation 2.
type Machine struct {
	ram    []byte
	spram  []byte
	iopRAM []byte

	CPU *r5900.CPU

	// io holds the last value written to a peripheral register that is not otherwise
	// modelled, and unmodelled records which ones were touched — the census that says
	// what the game actually asks the hardware for, and therefore what to model next.
	io         map[uint32]uint32
	unmodelled map[uint32]int

	// The loaded executable, kept for its symbol table: it is what lets the oracle
	// name an address instead of printing a number.
	exe *Executable

	// The mounted disc.
	vol *iso9660.Volume

	// Kernel HLE state (kernel.go).
	SyscallCalls map[string]int
	tty          []byte
	heapPtr      uint32
	heapEnd      uint32

	// Threads (sched.go).
	threads       map[uint32]*thread
	nextThreadID  uint32
	currentThread uint32

	// Interrupts (intr.go).
	intcHandlers []handler
	dmacHandlers []handler
	intcMask     uint32
	intcStat     uint32
	dmacMask     uint32
	dmacStat     uint32

	// The vertical-sync flags the kernel writes each frame on the game's behalf
	// (SetVSyncFlag). A game's idle loop spins on one of these.
	vsyncFlagPtr  uint32
	vsyncFlag2Ptr uint32

	// The GS display mode, recorded by SetGsCrt until the GS itself exists.
	gsInterlace, gsVideoMode, gsFieldMode uint32
	gsIMR                                 uint32

	// Semaphores (sema.go).
	semas      map[uint32]*sema
	nextSemaID uint32

	// Syscalls the game has replaced with routines of its own (SetSyscall).
	userSyscalls map[uint32]uint32

	// The argument list SetupThread establishes for the program.
	argc, argv uint32

	// DECI2 sockets handed out to the game's debug-output layer (kernel.go), and the
	// descriptor each one was opened with.
	deci2Sockets uint32
	deci2Desc    map[uint32]uint32

	// The SIF (sif.go): the handshake registers, the EE's command buffer and the
	// handler it registered to consume packets, the queue of replies the fake IOP owes
	// it, and the census of commands nobody answered.
	sifRegs       [32]uint32
	sifDmaID      uint32
	sifCmdBuf     uint32
	sifCmdHandler uint32
	sifPending    []sifPacket
	sifUnmodelled map[uint32]int

	// The second processor (iop.go). It is nil until the game reboots the IOP, which
	// is when the modules it will run are chosen.
	IOP *IOP

	// OnIOPStart, if set, is handed the second processor the moment it is built — which
	// is the only chance to attach an instrument to it, because the modules start driving
	// hardware inside the very entry points RebootIOP calls.
	OnIOPStart func(*IOP)

	// The six registers both processors can see (sifbus.go).
	sbus [sbusRegs]uint32

	// The buffer of arguments the EE DMA'd across just before its last RPC call.
	rpcSendBuf, rpcSendSize uint32

	// The RPC servers the EE has bound to, and the calls it has made to them.
	rpcServers  map[uint32]uint32 // server id -> synthetic handle
	rpcServerOf map[uint32]uint32 // handle -> server id
	rpcCalls    map[sifRPCKey]int

	// idle is set when no thread can run. The CPU stops; only the clock advances, until
	// an interrupt or a reply from the IOP makes something ready again.
	idle bool

	// steps counts every instruction the machine has run, across the whole session. The
	// SIF's reply latency is measured in it.
	steps uint64

	vblanks uint32

	imageHash string // pinned into savestates

	// Instrumentation (opt-in; checked in Read/Write and the run loop).
	WatchLo, WatchHi   uint32
	OnWrite            func(addr, val, pc uint32)
	RWatchLo, RWatchHi uint32
	OnRead             func(addr, val, pc uint32)
	OnStep             func(m *Machine, pc uint32)

	// hookMuted suppresses the read hooks during the machine's own reads (a
	// disassembly, a memory dump), which are not the guest reading anything.
	hookMuted bool

	breakpoints map[uint32]bool

	// StopRequested ends the run at the next instruction boundary.
	StopRequested bool

	// The IOP's stdio output, buffered until it has a whole line to log.
	iopTTYLine []byte

	Log     []string
	logSeen map[string]bool

	Halted     bool
	HaltReason string
}

// NewMachine makes a PS2 with memory and a CPU, and nothing running on it.
func NewMachine() *Machine {
	m := &Machine{
		ram:           make([]byte, ramSize),
		spram:         make([]byte, spramSize),
		iopRAM:        make([]byte, iopRAMSize),
		io:            map[uint32]uint32{},
		unmodelled:    map[uint32]int{},
		SyscallCalls:  map[string]int{},
		breakpoints:   map[uint32]bool{},
		logSeen:       map[string]bool{},
		threads:       map[uint32]*thread{},
		semas:         map[uint32]*sema{},
		nextSemaID:    1,
		userSyscalls:  map[uint32]uint32{},
		sifUnmodelled: map[uint32]int{},
		rpcServers:    map[uint32]uint32{},
		rpcServerOf:   map[uint32]uint32{},
		rpcCalls:      map[sifRPCKey]int{},
		deci2Desc:     map[uint32]uint32{},
	}
	m.CPU = r5900.NewCPU(m)
	m.CPU.Syscall = m.handleSyscall
	return m
}

// sprintf is fmt.Sprintf under a shorter name, used by the table renderers.
func sprintf(format string, args ...interface{}) string { return fmt.Sprintf(format, args...) }

// StartIOP builds the second processor over the memory the EE already shares with it.
// It runs nothing until a module is loaded onto it (iopload.go).
func (m *Machine) StartIOP() *IOP {
	m.IOP = newIOP(m, m.iopRAM)
	if m.OnIOPStart != nil {
		m.OnIOPStart(m.IOP)
	}
	return m.IOP
}

// iopPrint receives what the IOP's modules print through stdio. It is the IOP's half
// of the narration the EE's kernel already gives us over DECI2, and during bring-up it
// is very often the only thing that will say what went wrong.
func (m *Machine) iopPrint(s string) {
	m.iopTTYLine = append(m.iopTTYLine, s...)
	m.IOP.tty = append(m.IOP.tty, s...)
	for {
		i := bytes.IndexByte(m.iopTTYLine, '\n')
		if i < 0 {
			break
		}
		m.note("iop: %s", string(m.iopTTYLine[:i]))
		m.iopTTYLine = m.iopTTYLine[i+1:]
	}
}

// SetImageHash pins the disc's MD5 so a savestate cannot be loaded against a
// different image.
func (m *Machine) SetImageHash(md5 string) { m.imageHash = md5 }

// SetVolume mounts the disc the game will read from.
func (m *Machine) SetVolume(v *iso9660.Volume) { m.vol = v }

// Volume returns the mounted disc.
func (m *Machine) Volume() *iso9660.Volume { return m.vol }

// Exe returns the loaded executable, for its symbol table.
func (m *Machine) Exe() *Executable { return m.exe }

// --- loading ----------------------------------------------------------------

// LoadExecutable writes an ELF into memory, installs the mapping the BIOS kernel
// would have created, and sets the CPU up to enter at the executable's entry point.
func (m *Machine) LoadExecutable(e *Executable) {
	m.exe = e
	for _, seg := range e.Segments {
		copy(m.ram[seg.VAddr:], seg.Data)
		// The rest of the segment is BSS. It is already zero in a fresh machine, but
		// clearing it explicitly means a reload over a dirty machine behaves the same.
		for i := uint32(len(seg.Data)); i < seg.MemSz; i++ {
			m.ram[seg.VAddr+i] = 0
		}
	}

	m.mapMemory()

	// The state the kernel leaves the core in before it jumps to the entry point:
	// out of error level, using the RAM exception vectors, with the coprocessors
	// enabled and interrupts armed.
	m.CPU.COP0[r5900.Cop0Status] = r5900.StatusCU0 | r5900.StatusCU1 | r5900.StatusCU2 |
		r5900.StatusEIE | 0x0000FF00 // every interrupt line unmasked
	m.CPU.SetPC(uint64(e.Entry))

	// A stack at the top of main memory, and the global pointer the compiler expects.
	// The boot ELF sets both itself in _start, but a caller that jumps straight to an
	// inner routine (an oracle probe) needs them to already be sane.
	m.CPU.SetReg(29, uint64(ramSize-0x4000)) // $sp
	m.CPU.SetReg(28, 0)                      // $gp

	m.heapPtr = 0x01C00000 // a bump heap above the game, below the stack
	m.heapEnd = 0x01F00000

	// The context the entry point runs in is a thread like any other — the kernel's
	// GetThreadId has to answer something, and a later CreateThread has to be able to
	// out-rank it. It is not created by CreateThread; it simply exists.
	m.threads = map[uint32]*thread{
		1: {id: 1, entry: e.Entry, priority: 64, state: thRunning},
	}
	m.currentThread = 1
	m.nextThreadID = 2
}

// mapMemory installs the TLB entries the EE kernel creates before it hands control
// to a game.
//
// This is not an optimisation or a shortcut — it is a necessary part of standing in
// for the BIOS. Two regions need it, and a program dies immediately without either:
//
//   - Main memory. A PS2 executable is linked at 0x00100000, which is inside KUSEG,
//     and KUSEG is a *mapped* segment: with an empty TLB the very first instruction
//     fetch takes a miss.
//   - The hardware registers. Compiled PS2 code reaches the DMA controller, the
//     timers, the GIF and the rest at their bare physical addresses — 0x1000E000 and
//     the like — which are in KUSEG too. The kernel maps them, so a peripheral can be
//     driven with a plain store. Jak's very first library call reads a timer at
//     0x10001810, and without this mapping it faults on that one instruction.
//
// Both are mapped identity, with 16 MiB pages. A TLB entry covers a *pair* of pages,
// so one entry spans 32 MiB.
func (m *Machine) mapMemory() {
	const (
		pageMask16M = 0x01FFE000                // selects a 16 MiB page, so a 32 MiB pair
		page16M     = 0x1000                    // 16 MiB, counted in the 4 KiB frames a PFN counts
		flags       = 0x01 | 0x02 | 0x04 | 0x18 // global | valid | dirty | cached
	)
	// entry maps 32 MiB of virtual space at vaddr onto 32 MiB of physical space at
	// paddr. The two differ for the aliases below.
	entry := func(i int, vaddr, paddr uint32) {
		pfn := uint64(paddr >> 12)
		m.CPU.SetTLB(i, r5900.TLBEntry{
			PageMask: pageMask16M,
			EntryHi:  uint64(vaddr),
			EntryLo0: (pfn << 6) | flags,
			EntryLo1: ((pfn + page16M) << 6) | flags,
		})
	}

	// Main memory: 0x00000000..0x01FFFFFF, cached.
	entry(0, 0x00000000, 0x00000000)

	// The peripherals and everything above them: 0x10000000..0x1FFFFFFF, in eight
	// 32 MiB entries. Covering the range in one sweep — the EE's registers, the vector
	// units' memories, the GS's privileged registers, the IOP's RAM — is cheaper than
	// enumerating them and cannot leave a hole for a later peripheral to fall into.
	for i := 0; i < 8; i++ {
		entry(1+i, 0x10000000+uint32(i)*0x02000000, 0x10000000+uint32(i)*0x02000000)
	}

	// Main memory again, twice more.
	//
	// The kernel maps RAM at three virtual addresses, not one: cached at 0x00000000,
	// *uncached* at 0x20000000, and uncached-accelerated at 0x30000000. They are the
	// same 32 MiB seen through different cache behaviour, and PS2 code switches between
	// them constantly — anything a DMA engine will read is written through an uncached
	// alias, because a cached write might still be sitting in the cache when the DMA
	// starts.
	//
	// This model has no caches, so the three are literally the same memory. But the
	// *addresses* must translate, and a machine that maps only the cached one dies the
	// first time the game touches a DMA structure. It is not a corner case: the EE's own
	// SIF handler reads its command buffer through 0x2013B9C0.
	entry(9, 0x20000000, 0x00000000)
	entry(10, 0x30000000, 0x00000000)
}

// --- the bus ----------------------------------------------------------------

// phys folds the KSEG0/KSEG1 mirrors onto the physical address. The scratchpad is
// left alone: it lives at 0x70000000 and masking it would land it in the middle of
// nothing.
func phys(addr uint32) uint32 {
	if addr >= spramBase && addr < spramBase+spramSize {
		return addr
	}
	return addr & 0x1FFFFFFF
}

// mapped reports whether an address is backed by memory the CPU can execute from.
func (m *Machine) mapped(p uint32) bool {
	return p < ramSize || (p >= spramBase && p < spramBase+spramSize)
}

// slice returns the backing store for a physical address, and the offset into it.
func (m *Machine) slice(p uint32) ([]byte, uint32, bool) {
	switch {
	case p < ramSize:
		return m.ram, p, true
	case p >= spramBase && p < spramBase+spramSize:
		return m.spram, p - spramBase, true
	case p >= iopRAMBase && p < iopRAMBase+iopRAMSize:
		return m.iopRAM, p - iopRAMBase, true
	}
	return nil, 0, false
}

func (m *Machine) Read(addr uint32) byte {
	p := phys(addr)
	if buf, off, ok := m.slice(p); ok {
		v := buf[off]
		m.noteRead(p, uint32(v))
		return v
	}
	v := m.ioRead(p)
	return byte(v >> (8 * (p & 3)))
}

func (m *Machine) Write(addr uint32, v byte) {
	p := phys(addr)
	if buf, off, ok := m.slice(p); ok {
		buf[off] = v
		m.noteWrite(p, uint32(v))
		return
	}
	// A byte write to a register: merge it into the word.
	sh := 8 * (p & 3)
	w := m.io[p&^3]&^(0xFF<<sh) | uint32(v)<<sh
	m.ioWrite(p&^3, w)
}

func (m *Machine) Read32(addr uint32) uint32 {
	p := phys(addr)
	if buf, off, ok := m.slice(p); ok && off+4 <= uint32(len(buf)) {
		v := uint32(buf[off]) | uint32(buf[off+1])<<8 | uint32(buf[off+2])<<16 | uint32(buf[off+3])<<24
		m.noteRead(p, v)
		return v
	}
	return m.ioRead(p)
}

func (m *Machine) Write32(addr uint32, v uint32) {
	p := phys(addr)
	if buf, off, ok := m.slice(p); ok && off+4 <= uint32(len(buf)) {
		buf[off] = byte(v)
		buf[off+1] = byte(v >> 8)
		buf[off+2] = byte(v >> 16)
		buf[off+3] = byte(v >> 24)
		m.noteWrite(p, v)
		return
	}
	m.ioWrite(p, v)
}

// Fetch32 serves an instruction fetch. It bypasses the read hooks: a "who reads this
// address" watch is about data, and would otherwise be drowned by the fetch of every
// instruction that runs inside the window.
func (m *Machine) Fetch32(addr uint32) uint32 {
	p := phys(addr)
	if buf, off, ok := m.slice(p); ok && off+4 <= uint32(len(buf)) {
		return uint32(buf[off]) | uint32(buf[off+1])<<8 | uint32(buf[off+2])<<16 | uint32(buf[off+3])<<24
	}
	return 0
}

func (m *Machine) noteRead(p, v uint32) {
	if m.OnRead != nil && !m.hookMuted && p >= m.RWatchLo && p < m.RWatchHi {
		m.OnRead(p, v, uint32(m.CPU.CurPC()))
	}
}

func (m *Machine) noteWrite(p, v uint32) {
	if m.OnWrite != nil && p >= m.WatchLo && p < m.WatchHi {
		m.OnWrite(p, v, uint32(m.CPU.CurPC()))
	}
}

// --- unmodelled peripherals --------------------------------------------------

// ioRead and ioWrite stand in for every peripheral the machine does not yet have.
// A read returns whatever was last written (so a register the game polls for its own
// value behaves), and both tally the address.
//
// The tally is the point. One run of the boot prints exactly which registers the game
// touches and how often, which is the work list for the phases that follow — the same
// "unmodelled tier" the PSP kernel uses to enumerate its syscall surface, applied to
// hardware.
func (m *Machine) ioRead(p uint32) uint32 {
	if p >= sbusEEBase && p < sbusEEBase+sbusSpan {
		return m.sbusRead(p - sbusEEBase)
	}
	m.unmodelled[p]++
	return m.io[p]
}

func (m *Machine) ioWrite(p, v uint32) {
	if p >= sbusEEBase && p < sbusEEBase+sbusSpan {
		m.sbusWrite(p-sbusEEBase, v)
		return
	}
	m.unmodelled[p]++
	m.io[p] = v
}

// Unmodelled reports the peripheral registers the run touched, with hit counts.
func (m *Machine) Unmodelled() map[uint32]int { return m.unmodelled }

// HardwareCensus renders which peripheral registers the run touched, grouped by the
// unit they belong to. It is the work list for the phases that model them: a boot
// that hammers the DMAC and the GIF and never looks at the IPU says exactly which of
// those to build next.
func (m *Machine) HardwareCensus() string {
	if len(m.unmodelled) == 0 {
		return "no unmodelled hardware touched\n"
	}
	type unit struct {
		hits int
		regs map[uint32]int
	}
	units := map[string]*unit{}
	for addr, n := range m.unmodelled {
		name := RegionName(addr)
		u := units[name]
		if u == nil {
			u = &unit{regs: map[uint32]int{}}
			units[name] = u
		}
		u.hits += n
		u.regs[addr] += n
	}

	type kv struct {
		name string
		u    *unit
	}
	var all []kv
	for k, v := range units {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].u.hits > all[j].u.hits })

	s := "unmodelled hardware touched (the work list):\n"
	for _, e := range all {
		s += fmt.Sprintf("  %-8s %8d accesses over %d registers\n", e.name, e.u.hits, len(e.u.regs))
		// The hottest few registers of each unit, which is what says *what* the game
		// wants of it rather than merely that it wants something.
		var regs []uint32
		for r := range e.u.regs {
			regs = append(regs, r)
		}
		sort.Slice(regs, func(i, j int) bool { return e.u.regs[regs[i]] > e.u.regs[regs[j]] })
		for i, r := range regs {
			if i == 4 {
				s += fmt.Sprintf("      ... and %d more\n", len(regs)-4)
				break
			}
			s += fmt.Sprintf("      0x%08X  %d\n", r, e.u.regs[r])
		}
	}
	return s
}

// RegionName labels a peripheral address range, so the census reads as hardware
// rather than as numbers.
func RegionName(p uint32) string {
	switch {
	case p >= 0x10000000 && p < 0x10002000:
		return "TIMER"
	case p >= 0x10002000 && p < 0x10003000:
		return "IPU"
	case p >= 0x10003000 && p < 0x10003800:
		return "GIF"
	case p >= 0x10003800 && p < 0x10004000:
		return "VIF"
	case p >= 0x10004000 && p < 0x10008000:
		return "VIF-FIFO"
	case p >= 0x10008000 && p < 0x1000F000:
		return "DMAC"
	case p >= 0x1000F000 && p < 0x1000F200:
		return "INTC"
	case p >= 0x1000F200 && p < 0x1000F600:
		return "SIF"
	case p >= 0x1000F400 && p < 0x10010000:
		return "MCH"
	case p >= vuBase && p < vuEnd:
		return "VU"
	case p >= gsRegBase && p < gsRegEnd:
		return "GS"
	case p >= iopRAMBase && p < iopRAMBase+iopRAMSize:
		return "IOP-RAM"
	case p >= 0x1FC00000:
		return "BIOS"
	}
	return "?"
}

// --- diagnostics --------------------------------------------------------------

func (m *Machine) note(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if m.logSeen[s] {
		return
	}
	m.logSeen[s] = true
	m.Log = append(m.Log, s)
}

// Halt stops the machine.
func (m *Machine) Halt(format string, args ...interface{}) {
	m.Halted = true
	m.HaltReason = fmt.Sprintf(format, args...)
}

// SetBreakpoint halts the run when the CPU reaches vaddr, before the instruction.
func (m *Machine) SetBreakpoint(vaddr uint32) { m.breakpoints[vaddr] = true }

// ClearBreakpoints removes every breakpoint.
func (m *Machine) ClearBreakpoints() { m.breakpoints = map[uint32]bool{} }

// TTY returns whatever the game has printed through the kernel.
func (m *Machine) TTY() string { return string(m.tty) }

// VBlanks reports how many synthetic vertical blanks have elapsed.
func (m *Machine) VBlanks() uint32 { return m.vblanks }

// ReadMem copies n bytes of guest memory, without disturbing the read hooks.
func (m *Machine) ReadMem(addr uint32, n int) []byte {
	m.hookMuted = true
	defer func() { m.hookMuted = false }()
	out := make([]byte, n)
	for i := range out {
		out[i] = m.Read(addr + uint32(i))
	}
	return out
}

// CString reads a NUL-terminated string out of guest memory.
func (m *Machine) CString(addr uint32) string {
	m.hookMuted = true
	defer func() { m.hookMuted = false }()
	var b []byte
	for i := 0; i < 1024; i++ {
		c := m.Read(addr + uint32(i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

// DisasmAt renders the instruction at a virtual address.
func (m *Machine) DisasmAt(vaddr uint32) string {
	m.hookMuted = true
	defer func() { m.hookMuted = false }()
	return r5900.DecodeWord(m.Fetch32(vaddr), vaddr).Text
}

// Sym names the function containing a virtual address, as "name+0x1c", or the bare
// address when nothing covers it. Every instrument that prints a PC goes through
// this — which on a target whose executable ships symbols is the difference between
// a trace you can read and one you have to cross-reference.
func (m *Machine) Sym(addr uint32) string {
	if m.exe != nil {
		if name, off, ok := m.exe.Lookup(addr); ok {
			if off == 0 {
				return name
			}
			return fmt.Sprintf("%s+0x%X", name, off)
		}
	}
	return fmt.Sprintf("0x%08X", addr)
}

// Registers renders the CPU's general registers, for a breakpoint dump.
func (m *Machine) Registers() string {
	names := [32]string{
		"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
		"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
		"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
		"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
	}
	s := fmt.Sprintf("pc=%s\n", m.Sym(uint32(m.CPU.PC)))
	for i := 0; i < 32; i += 4 {
		for j := 0; j < 4; j++ {
			s += fmt.Sprintf("%4s=%016X  ", names[i+j], m.CPU.Reg(uint32(i+j)))
		}
		s += "\n"
	}
	return s
}
