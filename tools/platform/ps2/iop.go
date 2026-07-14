package ps2

// iop.go is the PlayStation 2's second processor.
//
// The IOP is a MIPS R3000A with 2 MiB of its own memory — a PlayStation, near
// enough, kept on board to own the disc drive, the sound chip, the controllers and
// the memory cards. The EE talks to it across the SIF and asks it for things. Until
// now this machine answered those requests in Go (sif.go), because the IOP's base
// kernel lives in a BIOS ROM we do not have and will not take.
//
// The disc supplies the rest. IOPRP221.IMG carries the kernel modules a game reboots
// the IOP onto, and /DRIVERS/ carries the game's own — OVERLORD.IRX above all, which
// is where Naughty Dog put the disc streaming, the DGO loader and the sound. So the
// IOP here is real: a core executing the modules off the disc, with only the ROM's
// own base libraries standing in as Go (iopkernel.go).
//
// The division is worth stating plainly, because it is the whole design:
//
//	run for real   anything that is logic. SIFCMD, SIFMAN, THREADMAN, IOMAN,
//	               MODLOAD, LOADFILE, FILEIO, CDVDFSV, MCMAN, MCSERV, PADMAN,
//	               989SND, and OVERLORD. The game's protocol is then not something
//	               we have inferred; it is something it is doing.
//
//	stand in       anything that touches silicon we do not model: the interrupt
//	               controller, the allocator, the C library, the CD hardware, the
//	               SPU. All of it Sony's, none of it the game's.
//
// The 2 MiB of IOP memory is the same []byte the EE sees at 0x1C000000 — one buffer,
// two processors, which is what it is on the board too, and what makes a DMA between
// them a copy rather than a fiction.

import (
	"fmt"
	"sort"

	"retroreverse.com/tools/cpu/mips"
)

// The IOP's address map.
const (
	iopRAMSizeBytes = 2 << 20 // 2 MiB

	// The scratchpad: 1 KiB of fast memory in the D-cache's place, exactly as on the
	// PlayStation.
	iopSPRAMBase = 0x1F800000
	iopSPRAMSize = 0x400

	// The hardware registers: the interrupt controller, the DMA controller, the timers,
	// the SPU2, the SIO2, and the SIF's own doorbells.
	iopIOBase = 0x1F801000
	iopIOEnd  = 0x1FA00000
)

// iopModuleBase is where the first module is placed.
//
// The low 64 KiB is left alone: it is where the exception vectors live and where a
// real IOP keeps its kernel, and a module loaded over the top of them would work
// right up until the first interrupt.
const iopModuleBase = 0x00010000

// IOP is the second processor.
type IOP struct {
	ram []byte // the same slice the EE sees at 0x1C000000
	spr []byte

	CPU *mips.CPU
	ps2 *Machine // the machine it belongs to, for the SIF and the log

	// The modules loaded into it, and the libraries they provide (iopload.go).
	modules []*IOPModule
	exports map[string]*IRXExport // library name -> the export table serving it
	owner   map[string]string     // library name -> the module that exports it

	// The bump allocator that places modules and anything the kernel hands out.
	// The IOP has no MMU: an address is an address, and once given it is given.
	allocPtr uint32

	// The syscall table. A stub for a library we model in Go is patched to
	// `jr $ra; syscall n`, and n indexes this (iopkernel.go). bound remembers which
	// code was handed to which (library, function), so every stub for the same
	// function shares one.
	calls []*iopCall
	bound map[iopBinding]uint32

	// stubName names every import stub the linker patched, by the library and function
	// number it was patched to serve. It is what turns `jal 0x00088670` — an address in
	// the middle of a module's .text, which no symbol covers because a stub is not a
	// function anybody named — into `libsd#7`. Both kinds of stub are in here: the ones
	// bound into Go and the ones linked to a real module on the disc. The second kind is
	// the one that matters, because those are the calls whose meaning has to be earned.
	stubName map[uint32]string

	// Whatever the IOP's modules have printed through stdio.
	tty []byte

	// The kernel HLE's state (iopintr.go): the interrupt controller, the allocator's
	// blocks, and the heaps.
	handlers    [iopIRQs]iopHandler
	imask       uint64 // which lines are unmasked: the interrupt controller's I_MASK
	intrEnabled bool
	blocks      []iopBlock
	heaps       map[uint32]*iopHeap

	// THREADMAN's two hooks into the interrupt-exit path: the predicate that says
	// whether a thread switch is wanted, and the routine that performs one. Registered
	// through intrman, and the contract the interrupt path has to keep.
	schedSwitch  uint32
	schedResched uint32

	// The interrupt controller (iopintr.go). pending is the set of lines raised and not
	// yet delivered — the hardware's I_STAT, except that intrman is ours, so no module
	// ever looks at it and it does not need to live at an address. inIntr counts how deep
	// in a handler we are: it is what QueryIntrContext answers, and what stops an
	// interrupt from being delivered inside one.
	pending uint64
	inIntr  int

	// The DMA controller (iopdma.go) and the sound chip (iopspu.go).
	dma        [iopDMAChannels]iopDMAChan
	dpcr       uint32
	dpcr2      uint32
	dicr       uint32
	dicr2      uint32
	dmaPending []iopDMADone
	spu        *spu2

	// The six counters (ioptimer.go). They are the second processor's only sense of time
	// passing, and THREADMAN's scheduler runs on them.
	timers [iopTimers]iopTimer

	// steps counts the instructions this processor has retired. A DMA's latency is
	// measured in it.
	steps uint64

	// prof is a sampling profile: every so many instructions, the routine the IOP is in.
	//
	// It answers the one question a bring-up asks over and over — "it is running, but what
	// is it running?" — which the spin detector cannot, because a machine going round a
	// large enough loop is not spinning by any definition it can apply. It is the only
	// instrument that tells the difference between a scheduler idling and a scheduler
	// working.
	prof map[string]int

	// The census: every call to a function nothing models, and every peripheral
	// register nobody has claimed. This is the work list, and it is the only honest
	// account of how much of the IOP is still missing.
	unmodelledCalls map[string]int
	io              map[uint32]uint32
	unmodelledIO    map[uint32]int

	// OnIO, if set, receives every access the IOP's own code makes to a peripheral
	// register: the address, the value, whether it was a write, and the PC that did it.
	//
	// This is the instrument the DMA controller and the SPU were built from. Sony's
	// modules are stripped, so the only account of how a transfer is programmed is the
	// order in which the registers are written — and that is a thing you watch, not a
	// thing you look up.
	OnIO func(addr, val uint32, write bool, pc uint32)

	// OnWrite, if set, reports every store the IOP makes into the window [WatchLo, WatchHi).
	// It is the same instrument the EE has, and it answers the same question: who wrote
	// this, and from where.
	WatchLo, WatchHi uint32
	OnWrite          func(addr, val uint32, pc uint32)

	// ioPend holds the register access the current instruction made, until the
	// instruction is over and the register has settled. See ioTrace.
	ioPend []ioTouch

	// lastPC is where the processor was an instruction ago, kept for one purpose: to say
	// where a jump into nothing came from.
	lastPC uint32

	// running is false until the first module is loaded. The machine does not step a
	// processor that has nothing on it.
	running bool

	// What the interrupt controller actually did: how often each line was raised, how
	// often it was delivered, and how many times THREADMAN chose to switch threads on the
	// way out. The gap between raised and delivered is the instrument that matters — a line
	// raised thirteen hundred times and delivered none is a masked interrupt, and it looks
	// from every other angle like a machine that is simply busy.
	raised    [iopIRQs]int
	delivered [iopIRQs]int
	switches  int

	// callDepth counts how deep we are inside a guest routine the *machine* called — a
	// module entry point, an interrupt handler, a scheduler hook. At depth zero the IOP is
	// running its own threads under THREADMAN and may be preempted freely; above zero it
	// is running something Go asked for, and the context is not one THREADMAN can file.
	// See intrDeliver.
	callDepth int

	Halted     bool
	HaltReason string
}

// IOPInterrupts reports what the second processor's interrupt controller did: every line
// that was raised, how many of those were delivered, and whether the scheduler switched
// threads as a result.
//
// The raised-versus-delivered column is the whole point of it. An interrupt that is raised
// and never delivered is not a quiet machine; it is a masked one, and from every other
// vantage point — the profile, the trace, the census — it looks exactly like a processor
// that is merely busy.
func (p *IOP) IOPInterrupts() string {
	s := fmt.Sprintf("the IOP's interrupts (%d thread switches by THREADMAN on interrupt exit):\n", p.switches)
	for i := 0; i < iopIRQs; i++ {
		if p.raised[i] == 0 && p.delivered[i] == 0 {
			continue
		}
		masked := ""
		if p.imask>>uint64(i)&1 == 0 {
			masked = "  [masked]"
		}
		s += fmt.Sprintf("      %2d  raised %7d  delivered %7d   %s%s\n",
			i, p.raised[i], p.delivered[i], p.Sym(p.handlers[i].fn), masked)
	}
	return s
}

// newIOP makes the second processor over the memory the EE already shares with it.
func newIOP(m *Machine, ram []byte) *IOP {
	p := &IOP{
		ram:             ram,
		spr:             make([]byte, iopSPRAMSize),
		ps2:             m,
		exports:         map[string]*IRXExport{},
		owner:           map[string]string{},
		allocPtr:        iopModuleBase,
		unmodelledCalls: map[string]int{},
		io:              map[uint32]uint32{},
		unmodelledIO:    map[uint32]int{},
		heaps:           map[uint32]*iopHeap{},
		stubName:        map[uint32]string{},
		spu:             newSPU2(),

		// Interrupts are on before the first module runs, because on the board the kernel
		// that loads a module has already armed them — the module is being called, not
		// booted. Starting this at false is a trap that closes silently: every module
		// brackets its critical sections with CpuSuspendIntr/CpuResumeIntr, and Resume is
		// passed *the value Suspend saved*. Suspend saves false, Resume restores false, the
		// round trip is faithful, and the processor never takes an interrupt again as long
		// as it lives. What that looks like from outside is not "interrupts are disabled";
		// it is OVERLORD spinning on a sound-transfer flag that the transfer really did set.
		intrEnabled: true,
	}
	p.CPU = mips.NewCPU(p)
	p.CPU.Syscall = p.handleSyscall
	p.registerLibraries()
	p.installIdle()
	return p
}

// --- the bus ------------------------------------------------------------------

// iopPhys folds the KSEG0 and KSEG1 mirrors onto the physical address. The IOP runs
// its kernel through 0x80000000 and reaches hardware through 0xA0000000, and both are
// the same 2 MiB and the same registers.
func iopPhys(addr uint32) uint32 { return addr & 0x1FFFFFFF }

func (p *IOP) Read(addr uint32) byte {
	a := iopPhys(addr)
	switch {
	case a < iopRAMSizeBytes:
		return p.ram[a]
	case a >= iopSPRAMBase && a < iopSPRAMBase+iopSPRAMSize:
		return p.spr[a-iopSPRAMBase]
	}
	v := p.ioRead(a &^ 3)
	p.ioTrace(a&^3, false)
	return byte(v >> (8 * (a & 3)))
}

func (p *IOP) Write(addr uint32, v byte) {
	a := iopPhys(addr)
	if p.OnWrite != nil && a >= p.WatchLo && a < p.WatchHi {
		p.OnWrite(a, uint32(v), p.CPU.CurPC())
	}
	switch {
	case a < iopRAMSizeBytes:
		p.ram[a] = v
		return
	case a >= iopSPRAMBase && a < iopSPRAMBase+iopSPRAMSize:
		p.spr[a-iopSPRAMBase] = v
		return
	}
	// A byte write to a register: merge it into the word. The word it merges into has to
	// be the register's *current* value — which for a register the other processor also
	// writes is not the same as the last value this one wrote. The R3000A's bus is
	// byte-wide, so every `sw` to a SIF doorbell arrives here four times, and a merge
	// against a stale word would drop half of it.
	sh := 8 * (a & 3)
	w := p.ioPeek(a&^3)&^(0xFF<<sh) | uint32(v)<<sh
	p.ioWrite(a&^3, w)
	p.ioTrace(a&^3, true)
}

// ioPeek reads a register without counting the access. It is for the read half of a
// read-modify-write, which is not the guest asking for anything.
func (p *IOP) ioPeek(a uint32) uint32 {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		return p.ps2.sbusRead(a - sbusIOPBase)
	}
	// Every modelled register has to answer here too, and not just in ioRead. This is the
	// read half of a byte-store's read-modify-write, and a word merged against the wrong
	// old value is a word half of which is nonsense — which for a DMA channel's CHCR
	// means a transfer with the right start bit and the wrong direction.
	if v, ok := p.dmaRead(a); ok {
		return v
	}
	if v, ok := p.timerPeek(a); ok {
		return v
	}
	if a >= iopSPU2Base && a < iopSPU2End {
		return p.spu.read(a - iopSPU2Base)
	}
	return p.io[a]
}

// Read32 and Write32 are for the machine's own use — the loader, the kernel HLE and
// the instruments. The core composes its own words a byte at a time.
//
// A register write must go through as *one word*, not as four bytes. The byte path
// merges its byte into the last value written to that address, and a shared register is
// one the other processor may have changed since — so four byte-writes to the SIF's
// doorbell would compose the new word out of a stale one and drop half of it. It is the
// kind of mistake that leaves a handshake that works locally and means nothing.
func (p *IOP) Read32(addr uint32) uint32 {
	a := iopPhys(addr)
	switch {
	case a+4 <= iopRAMSizeBytes:
		return uint32(p.ram[a]) | uint32(p.ram[a+1])<<8 | uint32(p.ram[a+2])<<16 | uint32(p.ram[a+3])<<24
	case a >= iopSPRAMBase && a+4 <= iopSPRAMBase+iopSPRAMSize:
		o := a - iopSPRAMBase
		return uint32(p.spr[o]) | uint32(p.spr[o+1])<<8 | uint32(p.spr[o+2])<<16 | uint32(p.spr[o+3])<<24
	}
	return p.ioRead(a &^ 3)
}

func (p *IOP) Write32(addr, v uint32) {
	a := iopPhys(addr)
	switch {
	case a+4 <= iopRAMSizeBytes:
		p.ram[a] = byte(v)
		p.ram[a+1] = byte(v >> 8)
		p.ram[a+2] = byte(v >> 16)
		p.ram[a+3] = byte(v >> 24)
	case a >= iopSPRAMBase && a+4 <= iopSPRAMBase+iopSPRAMSize:
		o := a - iopSPRAMBase
		p.spr[o] = byte(v)
		p.spr[o+1] = byte(v >> 8)
		p.spr[o+2] = byte(v >> 16)
		p.spr[o+3] = byte(v >> 24)
	default:
		p.ioWrite(a&^3, v)
	}
}

// CString reads a NUL-terminated string out of IOP memory.
func (p *IOP) CString(addr uint32) string {
	var b []byte
	for i := uint32(0); i < 1024; i++ {
		c := p.Read(addr + i)
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

// ioRead and ioWrite stand in for every IOP peripheral, and tally what was touched.
// The tally is the same instrument the EE's has: it says which silicon the modules
// actually drive, and therefore what to model next.
func (p *IOP) ioRead(a uint32) uint32 {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		return p.ps2.sbusRead(a - sbusIOPBase)
	}
	if v, ok := p.dmaRead(a); ok {
		return v
	}
	if v, ok := p.timerRead(a); ok {
		return v
	}
	if a >= iopSPU2Base && a < iopSPU2End {
		return p.spu.read(a - iopSPU2Base)
	}
	p.unmodelledIO[a]++
	return p.io[a]
}

func (p *IOP) ioWrite(a, v uint32) {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		p.ps2.sbusWrite(a-sbusIOPBase, v)
		return
	}
	if p.dmaWrite(a, v) {
		return
	}
	if p.timerWrite(a, v) {
		return
	}
	if a >= iopSPU2Base && a < iopSPU2End {
		p.spu.write(a-iopSPU2Base, v)
		return
	}
	p.unmodelledIO[a]++
	p.io[a] = v
}

// --- memory ---------------------------------------------------------------------

// alloc takes n bytes of IOP memory, 64-byte aligned.
//
// It is a bump allocator and it never frees. That is not a placeholder: an IOP boot
// loads its modules once and keeps them, and a kernel that cannot free is a kernel
// whose addresses stay put — which is worth a great deal when the thing being
// debugged is a machine you are still writing.
func (p *IOP) alloc(n uint32) uint32 {
	a := (p.allocPtr + 63) &^ 63
	p.allocPtr = a + n
	if p.allocPtr > iopStackArea {
		p.halt("out of IOP memory: %d bytes wanted, and the allocator has reached the stacks at 0x%08X",
			n, uint32(iopStackArea))
		return 0
	}
	return a
}

// iopStackArea is where the allocator has to stop.
//
// The top 64 KiB of IOP memory belongs to the machine rather than to the guest: the stack
// a routine called from Go runs on, the separate one an interrupt handler runs on, and the
// one the idle loop sits on. They are not allocations and nothing on the IOP knows they are
// there, so the allocator has to be told — a heap that grew into them would corrupt the
// stack of the very code that asked it to grow.
const (
	iopStackArea = iopRAMSizeBytes - 0x10000
	iopIdleStack = iopRAMSizeBytes - 0x4400
)

// --- the idle loop ------------------------------------------------------------------

// iopIdleLoop is where the second processor goes when it has nothing else to do.
//
// It needs one, and the need is not obvious until the IOP is asked to run on its own for
// the first time. Every module's entry point is called from Go and returns to Go, and when
// the last one has returned the processor's program counter is back at its reset vector —
// which on this machine is a BIOS we do not have. Step it from there and it reads zeroes,
// executes them, and walks up through memory four bytes at a time, and the profile of a
// perfectly healthy boot is a list of addresses beginning at zero.
//
// A real IOP does not have this problem, because the kernel that loaded the modules is
// itself a thread, and when it has finished loading them it goes round an idle loop
// forever. So this is that loop: two instructions, in the low 64 KiB where a real IOP keeps
// its kernel, and the machine parks the processor there whenever it is not inside a call.
//
// What happens next is the entire scheduler, and none of it is ours. The loop spins with
// interrupts on. The timer fires. THREADMAN's handler runs and makes some thread ready. On
// the way out, the predicate says the running thread is no longer the one that ought to be
// running — and the thread it names is a real one, and the frame the machine is holding
// (the idle loop's) is filed away as the outgoing thread's, exactly as it should be,
// because idling *is* what that thread was doing.
const (
	iopIdleLoop = 0x00000200

	// `beq $zero, $zero, -1` — a branch to itself, with a nop in its delay slot.
	iopIdleInsn = 0x1000FFFF
)

// installIdle writes the idle loop into memory and parks the processor on it.
func (p *IOP) installIdle() {
	p.Write32(iopIdleLoop, iopIdleInsn)
	p.Write32(iopIdleLoop+4, insnNop())
	p.CPU.SetPC(iopIdleLoop)
	p.CPU.SetReg(29, iopIdleStack)
}

// --- running ---------------------------------------------------------------------

// iopStepRatio is how many EE instructions pass for each IOP instruction.
//
// The EE runs at about 294 MHz and the IOP at about 36 MHz, so the true ratio is
// near eight. It matters more than it looks: the two processors hand work to each
// other and then wait, and an IOP running at the wrong speed turns a handshake into
// either a spin or a race — neither of which is a bug in the code being debugged.
const iopStepRatio = 8

// Step runs the IOP for one instruction, if it has anything to run.
func (p *IOP) Step() {
	if !p.running || p.Halted || p.CPU.Halted {
		return
	}
	p.tick()
	p.CPU.Step()
}

func (p *IOP) halt(format string, args ...interface{}) {
	p.Halted = true
	p.HaltReason = fmt.Sprintf(format, args...)
	p.ps2.note("IOP halted: %s", p.HaltReason)
}

// TTY returns whatever the IOP's modules have printed.
func (p *IOP) TTY() string { return string(p.tty) }

// DisasmAt renders IOP memory as instructions, as loaded — relocated, linked, and with
// every import stub already patched to whatever it was bound to.
//
// That last part is why this exists rather than a static disassembler over the file.
// Sony's kernel modules are stripped, so the only way to learn what `intrman #4` is is
// to read the code that calls it — and the code that calls it is only legible once you
// can see, at the call site, that the thing being called is a jump into THREADMAN's own
// interrupt handler rather than a number.
func (p *IOP) DisasmAt(addr uint32) string {
	code := []byte{p.Read(addr), p.Read(addr + 1), p.Read(addr + 2), p.Read(addr + 3)}
	in := mips.Decode(code, addr)

	// A call into the kernel HLE is a jump to a stub whose second word is a syscall.
	// Name it: `jal 0x00013abc` says nothing, `jal intrman.CpuSuspendIntr` says
	// everything.
	if in.HasTarget {
		if name := p.callName(in.Target); name != "" {
			return fmt.Sprintf("%-28s ; %s", in.Text, name)
		}
		// A stub linked to a real module on the disc. The linker rewrote it to jump
		// straight into the callee, so following the target lands inside the *other*
		// module and tells you nothing about what was asked for; the library and the
		// function number are what was asked for.
		if name := p.stubName[in.Target]; name != "" {
			return fmt.Sprintf("%-28s ; %s -> %s", in.Text, name, p.Sym(p.stubTarget(in.Target)))
		}
		if s := p.Sym(in.Target); s != "" {
			return fmt.Sprintf("%-28s ; %s", in.Text, s)
		}
	}
	return in.Text
}

// stubTarget follows a direct-linked stub to the code it jumps to, so the disassembly
// can show both what was imported and where it went.
func (p *IOP) stubTarget(addr uint32) uint32 {
	w := p.Read32(addr)
	if w>>26 != 0x02 { // j
		return addr
	}
	return (addr &^ 0x0FFFFFFF) | (w&0x03FFFFFF)<<2
}

// callName reports the kernel function a stub address stands for, if it is one.
func (p *IOP) callName(addr uint32) string {
	if p.Read32(addr) != insnJR(regRA()) {
		return ""
	}
	w := p.Read32(addr + 4)
	if w&0x3F != 0x0C {
		return ""
	}
	code := (w >> 6) & 0xFFFFF
	if code == 0 || int(code) >= len(p.calls) {
		return ""
	}
	return p.calls[code].name
}

// Modules returns the modules resident in IOP memory.
func (p *IOP) Modules() []*IOPModule { return p.modules }

// Sym names an address in IOP memory, using the symbol tables of the modules loaded
// there. Sony's kernel modules are stripped, but the game's are not: an address
// inside OVERLORD comes back as `ISOThread+0x1c`, which is the difference between
// reading a trace and decoding one.
func (p *IOP) Sym(addr uint32) string {
	for _, m := range p.modules {
		if addr < m.Base || addr >= m.Base+m.Size {
			continue
		}
		off := addr - m.Base
		var best *Symbol
		for i := range m.IRX.Symbols {
			s := &m.IRX.Symbols[i]
			if s.Func && s.Addr <= off && (best == nil || s.Addr > best.Addr) {
				best = s
			}
		}
		if best != nil {
			if d := off - best.Addr; d != 0 {
				return fmt.Sprintf("%s+0x%X", best.Name, d)
			}
			return best.Name
		}
		return fmt.Sprintf("%s+0x%X", m.Name, off)
	}
	return fmt.Sprintf("0x%08X", addr)
}

// SymAddr resolves a symbol name to its address in IOP memory, as loaded.
//
// It is the inverse of Sym, and it exists so that an instrument can be pointed at
// `DMA_SendToSPUAndSync` rather than at a number that changes the moment a module ahead
// of it in the load order grows by a byte. Only the game's modules carry symbols;
// Sony's are stripped, and for those the address is still the only handle there is.
func (p *IOP) SymAddr(name string) (uint32, bool) {
	for _, m := range p.modules {
		for i := range m.IRX.Symbols {
			if s := &m.IRX.Symbols[i]; s.Name == name {
				return m.Base + s.Addr, true
			}
		}
	}
	return 0, false
}

// --- the census -------------------------------------------------------------------

// iopProfileEvery is how often the profiler takes a sample, in instructions. It is a
// power of two so the test is a mask, and coarse enough that the sampling costs nothing
// next to the work being sampled.
const iopProfileEvery = 4096

// IOPProfile reports where the second processor spent its time, hottest first.
func (p *IOP) IOPProfile() string {
	if len(p.prof) == 0 {
		return ""
	}
	type kv struct {
		name string
		n    int
	}
	var all []kv
	total := 0
	for k, n := range p.prof {
		all = append(all, kv{k, n})
		total += n
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].n != all[j].n {
			return all[i].n > all[j].n
		}
		return all[i].name < all[j].name
	})
	s := fmt.Sprintf("where the IOP spent its time (%d samples, one per %d instructions):\n",
		total, iopProfileEvery)
	for i, e := range all {
		if i == 12 {
			s += fmt.Sprintf("      ... and %d more routines\n", len(all)-12)
			break
		}
		s += fmt.Sprintf("      %5.1f%%  %s\n", 100*float64(e.n)/float64(total), e.name)
	}
	return s
}

// IOPCensus reports what the IOP asked for that nothing answered: the kernel
// functions still unmodelled, and the hardware registers still unclaimed.
//
// It is the same instrument as the EE's syscall census and the hardware census, and
// it exists for the same reason. Modules resolve their imports by *index* — there
// are no names on the wire — so the only way to learn what the IOP kernel owes the
// disc is to let the disc ask, and write down what it asked for.
func (p *IOP) IOPCensus() string {
	if len(p.unmodelledCalls) == 0 && len(p.unmodelledIO) == 0 {
		return ""
	}
	s := "the IOP's unanswered requests (the work list):\n"

	if len(p.unmodelledCalls) > 0 {
		type kv struct {
			name string
			n    int
		}
		var all []kv
		for k, n := range p.unmodelledCalls {
			all = append(all, kv{k, n})
		}
		sort.Slice(all, func(i, j int) bool {
			if all[i].n != all[j].n {
				return all[i].n > all[j].n
			}
			return all[i].name < all[j].name
		})
		s += "  kernel functions nothing models:\n"
		for _, e := range all {
			s += fmt.Sprintf("      %-20s %d call%s\n", e.name, e.n, plural(e.n))
		}
	}

	if len(p.unmodelledIO) > 0 {
		var regs []uint32
		for a := range p.unmodelledIO {
			regs = append(regs, a)
		}
		sort.Slice(regs, func(i, j int) bool { return p.unmodelledIO[regs[i]] > p.unmodelledIO[regs[j]] })
		s += fmt.Sprintf("  IOP hardware touched: %d registers\n", len(regs))
		for i, a := range regs {
			if i == 8 {
				s += fmt.Sprintf("      ... and %d more\n", len(regs)-8)
				break
			}
			s += fmt.Sprintf("      0x%08X  %s  %d\n", a, IOPRegionName(a), p.unmodelledIO[a])
		}
	}
	return s
}

// IOPRegionName labels an IOP peripheral, so the census reads as hardware.
func IOPRegionName(a uint32) string {
	switch {
	case a >= 0x1F801040 && a < 0x1F801060:
		return "SIO"
	case a >= 0x1F801070 && a < 0x1F801080:
		return "INTC"
	case a >= 0x1F801080 && a < 0x1F801100:
		return "DMA"
	case a >= 0x1F801100 && a < 0x1F801140:
		return "TIMER"
	case a >= 0x1F801450 && a < 0x1F801460:
		return "?"
	case a >= 0x1F801500 && a < 0x1F801580:
		return "DMA2"
	case a >= 0x1F802070 && a < 0x1F802080:
		return "POST"
	case a >= 0x1F808200 && a < 0x1F808300:
		return "SIO2"
	case a >= 0x1F900000 && a < 0x1FA00000:
		return "SPU2"
	case a >= 0x1D000000 && a < 0x1D000100:
		return "SIF"
	}
	return "?"
}

// symFunc names the routine containing an address, without the offset. It is Sym for a
// profile, where "ISOThread+0x1c" and "ISOThread+0x20" are the same answer.
func (p *IOP) symFunc(addr uint32) string {
	s := p.Sym(addr)
	if i := len(s) - 1; i > 0 {
		if j := indexByte(s, '+'); j > 0 {
			return s[:j]
		}
	}
	return s
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// --- the register trace ------------------------------------------------------------
//
// The R3000A's bus is byte-wide, so one `sw` to a register arrives at the bus four
// times and one `sh` twice, and neither announces how wide it was. A trace that fired on
// every byte would print four lines for one store, each carrying a partly-merged word; a
// trace that fired only on the byte that completes a *word* would print nothing at all
// for the halfword stores that drive the timers' mode registers — which is exactly the
// register whose absence from an earlier trace sent this bring-up looking in the wrong
// place.
//
// So the trace is deferred instead. An access records which register was touched, and the
// line is emitted at the start of the next instruction, by which time every byte of the
// access has landed and the register holds the value the guest meant to leave in it. One
// instruction that touches a register produces exactly one line, whatever its width.

// ioTrace records that the current instruction touched a register.
func (p *IOP) ioTrace(a uint32, write bool) {
	if p.OnIO == nil {
		return
	}
	p.ioPend = append(p.ioPend[:0], ioTouch{addr: a, write: write, pc: p.CPU.CurPC()})
}

// ioTouch is one register access, waiting for the instruction to finish.
type ioTouch struct {
	addr  uint32
	pc    uint32
	write bool
}

// ioTraceFlush emits the previous instruction's register access, if it made one.
func (p *IOP) ioTraceFlush() {
	if p.OnIO == nil || len(p.ioPend) == 0 {
		return
	}
	t := p.ioPend[0]
	p.ioPend = p.ioPend[:0]
	p.OnIO(t.addr, p.ioPeek(t.addr), t.write, t.pc)
}

// Run steps the IOP on its own, for up to n instructions.
//
// This is the second processor running as a processor rather than as a called routine:
// its threads are scheduled by THREADMAN, its interrupts arrive from its own timers and
// devices, and nothing outside it is driving. It is what the IOP does for the whole of a
// game — the EE asks it for things across the SIF, and the rest of the time it is simply
// alive.
//
// It stops early if the machine halts. It does not stop when there is nothing to do:
// an IOP with every thread blocked is an IOP waiting for an interrupt, and the interrupt
// is on its way.
func (p *IOP) Run(n uint64) {
	for i := uint64(0); i < n; i++ {
		if !p.running || p.Halted || p.CPU.Halted {
			return
		}
		p.tick()
		p.CPU.Step()
	}
}

// Steps reports how many instructions the second processor has retired.
func (p *IOP) Steps() uint64 { return p.steps }
