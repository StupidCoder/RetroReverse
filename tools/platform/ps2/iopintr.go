package ps2

import "fmt"

// iopintr.go is intrman, sysmem, heaplib and sysclib: the four libraries the IOP's
// modules cannot take a step without, and the four we have to write.
//
// Everything here was read out of the code that calls it. Sony's kernel modules are
// stripped and their functions are imported by number, so an identification is a small
// argument rather than a lookup, and each one is recorded where it was made. The
// arguments are short because the callers are honest: a routine that passes a pointer,
// a zero and its own .bss size and then relies on the memory being clear is memset, and
// it is not much else.
//
// What is *not* here matters as much. intrman #8, #9, #14 and #23, sysmem #5 to #8,
// heaplib #5, #7 and #8, and most of sysclib are still unmodelled — called, counted,
// and answered with zero. They are in the census, and the census is the work list.

func init() {
	// intrman. The interrupt controller.
	//
	// The two that matter most are the pair every module brackets its critical sections
	// with, and they are named by their own call convention: #17 is passed a *pointer*
	// to a stack slot and #18 is passed the value that was in it. That is
	// CpuSuspendIntr(int *old) / CpuResumeIntr(int old), and nothing else has that
	// shape.
	//
	// #23 is QueryIntrContext, and it is named by the only use its answer is ever put to.
	// LIBSD calls it at the top of its transfer routine and branches on the result to one
	// of *two versions of the same THREADMAN call* — thevent#8 on one side, thevent#9 on
	// the other, with identical arguments. A kernel exports a pair like that for exactly
	// one reason: one of them is safe to call from inside an interrupt and the other is
	// not, and the caller has to know which it is. So the question being asked is "am I in
	// an interrupt?", and the answer is how deep in a handler we are.
	//
	// #28 and #30 are THREADMAN's hooks into interrupt exit. It hands each of them one of
	// its own functions, and the one it hands to #30 is four instructions long: it compares
	// two of THREADMAN's globals and returns whether they differ. That is "does the
	// scheduler want to switch threads?" — so #30 registers the predicate the kernel
	// consults on the way out of an interrupt, and #28 the routine that acts on it. Both
	// are called now, on the path in intrDeliver.
	lib("intrman", map[uint16]iopFunc{
		4:  {"RegisterIntrHandler", (*IOP).intrRegister},
		5:  {"ReleaseIntrHandler", (*IOP).intrRelease},
		6:  {"EnableIntr", (*IOP).intrEnable},
		7:  {"DisableIntr", (*IOP).intrDisable},
		8:  {"CpuDisableIntr", (*IOP).intrCpuDisable},
		9:  {"CpuEnableIntr", (*IOP).intrCpuEnable},
		14: {"CpuInvokeInKmode", (*IOP).intrInvokeInKmode},
		17: {"CpuSuspendIntr", (*IOP).intrSuspend},
		18: {"CpuResumeIntr", (*IOP).intrResume},
		23: {"QueryIntrContext", (*IOP).intrQueryContext},
		28: {"<register scheduler hook>", (*IOP).intrSetSwitchHook},
		30: {"<register reschedule predicate>", (*IOP).intrSetReschedHook},
	})

	// sysmem. The allocator, and — because it is the lowest library in the machine and
	// the one every other module can already see — the kernel's printf.
	//
	// #14 takes a pointer into a module's .rodata and up to three more registers, and
	// THREADMAN calls it on a debug path with the current thread's id. It is a printf.
	// That there are two of them (this and stdio #4) is not a mistake: the IOP has a
	// kernel printf and a library printf, and modules import whichever they can reach.
	lib("sysmem", map[uint16]iopFunc{
		4:  {"AllocSysMemory", (*IOP).sysmemAlloc},
		5:  {"FreeSysMemory", (*IOP).sysmemFree},
		6:  unknown(),
		7:  unknown(),
		8:  unknown(),
		9:  {"QueryBlockTopAddress", (*IOP).sysmemBlockTop},
		10: {"QueryBlockSize", (*IOP).sysmemBlockSize},
		14: {"Kprintf", (*IOP).sysmemKprintf},
	})

	// heaplib. A heap on top of sysmem. THREADMAN creates one and then takes every
	// thread control block, semaphore and event flag out of it.
	//
	// #6 is named by what its caller does next: THREADMAN checks the returned pointer
	// against zero and then memsets that many bytes at it. An allocator, and one that
	// does not clear.
	lib("heaplib", map[uint16]iopFunc{
		4: {"CreateHeap", (*IOP).heapCreate},
		5: unknown(),
		6: {"AllocHeapMemory", (*IOP).heapAlloc},
		7: unknown(),
		8: unknown(),
	})

	// sysclib. The C library.
	//
	// #14 is memset: THREADMAN's first act is to call it with its own .bss, a zero, and
	// 1220 — which is exactly the size of its .bss — and then to rely on the result
	// being clear.
	lib("sysclib", map[uint16]iopFunc{
		8: unknown(), 11: unknown(), 12: unknown(), 13: unknown(),
		14: {"memset", (*IOP).clibMemset},
		17: unknown(), 19: unknown(), 20: unknown(), 22: unknown(), 23: unknown(),
		25: unknown(), 27: unknown(), 29: unknown(), 30: unknown(), 36: unknown(),
	})
}

// --- intrman ----------------------------------------------------------------------

// iopIRQs is how big intrman's vector table is, and it is bigger than the interrupt
// controller.
//
// The chip has 32 lines. intrman offers more, and the extra ones are the DMA channels:
// the controller raises a single line for "some channel finished" and the kernel demuxes
// it, so from a driver's point of view each channel has an interrupt of its own. The
// numbers were read off the registrations, and four modules agree:
//
//	36  LIBSD    channel 4    the sound chip's first core
//	40  LIBSD    channel 7    its second
//	42  SIFMAN   channel 9    SIF0
//	43  SIFCMD   channel 10   SIF1
//
// which is 32 + n for the first block of channels and 40 + (n-7) for the second — eight
// numbers to a block, with the second block's numbering starting at a round 40 rather than
// carrying on from the first. That is not a guess between two readings; it is the only
// arithmetic that fits all four, and it independently confirms the channel identities
// derived from DPCR (iopdma.go), which were established from an entirely different
// register by an entirely different argument.
//
// A table of 32 does not merely lose these. It loses them *silently*: RegisterIntrHandler
// is handed 36, finds it out of range, returns an error the caller does not check, and the
// sound chip's completion interrupt is never registered at all. The handler that would have
// run is the one that calls the callback that sets the flag OVERLORD is spinning on.
const iopIRQs = 64

// iopHandler is one registered interrupt handler: an entry in the vector table, and
// nothing more.
//
// Whether the line is *unmasked* is deliberately not in here, and the reason is worth
// recording, because putting it here is a bug that hides perfectly. On the hardware the
// vector table and the mask register are two different pieces of state, and a driver may
// write them in either order — LIBSD writes them in the order nobody expects, calling
// EnableIntr(9) and only then RegisterIntrHandler(9). With the two conflated in one
// struct, registering the handler stores a fresh record over the top of the enable that
// had already happened, and the sound chip's line is masked from that moment on. The
// symptom is not a masked interrupt. The symptom is OVERLORD waiting forever for a DMA
// that completed, on a machine where the DMA controller, the SPU and the interrupt path
// are all working.
type iopHandler struct {
	fn  uint32 // the module's own routine
	arg uint32
}

func (p *IOP) intrRegister() {
	irq, _, fn, arg := p.arg(0), p.arg(1), p.arg(2), p.arg(3)
	if irq >= iopIRQs {
		p.ps2.note("IOP: a handler was registered on interrupt %d, which is past the end of the table", irq)
		p.setRet(0xFFFFFF9A) // out of range; the caller checks
		return
	}
	p.handlers[irq] = iopHandler{fn: fn, arg: arg}
	p.ps2.note("IOP: interrupt %d handled by %s", irq, p.Sym(fn))
	p.setRet(0)
}

func (p *IOP) intrRelease() {
	if irq := p.arg(0); irq < iopIRQs {
		p.handlers[irq] = iopHandler{}
	}
	p.setRet(0)
}

func (p *IOP) intrEnable() {
	if irq := p.arg(0); irq < iopIRQs {
		p.imask |= 1 << uint64(irq)
	}
	p.setRet(0)
}

func (p *IOP) intrDisable() {
	if irq := p.arg(0); irq < iopIRQs {
		p.imask &^= 1 << uint64(irq)
	}
	p.setRet(0)
}

// intrSuspend is CpuSuspendIntr(int *old): it saves the current interrupt-enable state
// through the pointer it is given, disables interrupts, and returns success.
//
// The pointer is the whole of the identification. Every module brackets its critical
// sections with this and its partner, and the partner is passed exactly the word this
// one wrote.
func (p *IOP) intrSuspend() {
	if ptr := p.arg(0); ptr != 0 {
		p.Write32(ptr, b2u(p.intrEnabled))
	}
	p.intrEnabled = false
	p.setRet(0)
}

// intrResume is CpuResumeIntr(int old).
func (p *IOP) intrResume() {
	p.intrEnabled = p.arg(0) != 0
	p.setRet(0)
}

// intrCpuDisable and intrCpuEnable are CpuDisableIntr() and CpuEnableIntr(): the
// processor's master interrupt switch, as against the per-line mask that #6 and #7 drive.
//
// Neither takes an argument — both are called with a `nop` in the delay slot and no
// register set up before them — and they bracket a module's *initialisation*, which is
// what distinguishes them from the suspend/resume pair. 989SND calls #8 as the first
// thing it does; THREADMAN calls #9 as the last instruction before its entry point
// returns.
//
// THREADMAN is the argument. Its entry point opens by calling CpuSuspendIntr, and it
// never calls CpuResumeIntr — not once, anywhere. If #9 is not the thing that turns the
// processor's interrupts back on, then nothing is, and the IOP runs the rest of its life
// with them off: every module's Suspend faithfully saves "disabled", every matching Resume
// faithfully restores it, and the machine looks entirely healthy while the sound chip's
// completed transfer waits at a masked door forever. Which is exactly what it did.
func (p *IOP) intrCpuDisable() {
	p.intrEnabled = false
	p.setRet(0)
}

func (p *IOP) intrCpuEnable() {
	p.intrEnabled = true
	p.setRet(0)
}

// intrSetSwitchHook and intrSetReschedHook record THREADMAN's two hooks into the
// interrupt-exit path. The contract they describe is kept in intrDeliver: run the
// handler, ask the predicate, and if it says yes, call the switcher.
func (p *IOP) intrSetSwitchHook() {
	p.schedSwitch = p.arg(0)
	p.ps2.note("IOP: the scheduler's switch routine is %s", p.Sym(p.schedSwitch))
	p.setRet(0)
}

func (p *IOP) intrSetReschedHook() {
	p.schedResched = p.arg(0)
	p.ps2.note("IOP: the scheduler's reschedule predicate is %s", p.Sym(p.schedResched))
	p.setRet(0)
}

// intrQueryContext is QueryIntrContext(): non-zero while an interrupt handler is running.
func (p *IOP) intrQueryContext() { p.setRet(b2u(p.inIntr > 0)) }

// intrInvokeInKmode is #14: run this function, and give me back what it returns.
//
// The first argument is a function pointer and the rest are its arguments, and that is the
// whole shape of it — THREADMAN wraps it in a three-line exported routine that passes one
// of its own functions and the caller's value, and the function it passes goes straight on
// to call a timer routine through timrman. So the name is about *where* the call runs, and
// the answer here is that it does not matter: this machine has no privilege levels to cross
// and no kernel mode to enter, so an invocation in kernel mode is an invocation.
//
// What matters is that it happens. Answering zero and calling nothing — which is what an
// unmodelled function does, and what this one did — quietly deletes whatever the caller
// asked to have run. Here it deleted the routine that arms THREADMAN's timer, and the
// second processor lost its heartbeat: the scheduler ran out of ready threads, fell into
// its idle loop, and waited for a clock that nobody had started.
func (p *IOP) intrInvokeInKmode() {
	fn := p.arg(0)
	if fn == 0 {
		p.setRet(0)
		return
	}
	a1, a2, a3 := p.arg(1), p.arg(2), p.arg(3)

	p.CPU.SetReg(4, a1)
	p.CPU.SetReg(5, a2)
	p.CPU.SetReg(6, a3)
	p.CPU.SetReg(7, 0)

	// On the caller's own stack, below its frame: this is a call the caller made, and it
	// returns to the caller. It is not an interrupt and it does not get a stack of its own.
	res, err := p.callGuestOn(fn, (p.CPU.Reg(29)-64)&^7)
	if err != nil {
		p.halt("the routine %s, invoked through intrman #14, did not return: %v", p.Sym(fn), err)
		return
	}
	p.setRet(res)
}

// --- delivering an interrupt ---------------------------------------------------------

// iopIntrStack is where an interrupt handler runs.
//
// It cannot be the stack callGuest uses, and that is not a detail. An interrupt arrives
// in the middle of whatever was running, so a handler starting at the top of the shared
// call stack would lay its own frame straight over the frames of the routine it
// interrupted — and the routine it interrupts is, almost always, the one waiting for it.
const iopIntrStack = iopRAMSizeBytes - 0x8400

// raiseIRQ marks an interrupt line. It is the hardware's I_STAT, and it is a Go field
// rather than a register because intrman is ours: no module on this disc reads the
// interrupt controller's registers, because none of them has to.
func (p *IOP) raiseIRQ(irq uint32) {
	p.pending |= 1 << uint64(irq)
	p.raised[irq]++
}

// serviceIntr delivers one pending interrupt, if the processor will take it.
//
// A line whose handler is not enabled stays raised. That is what the hardware does, and
// it is also the honest thing: the interrupt has happened whether or not anyone is
// listening yet, and a module that enables its handler afterwards should see it.
func (p *IOP) serviceIntr() {
	if p.inIntr > 0 || !p.intrEnabled || p.pending == 0 {
		return
	}

	// Not in the shadow of a branch, and not with a load still in the air.
	//
	// The frame a context is saved into holds one program counter, and one program counter
	// cannot describe a processor that is about to execute a delay slot — the branch it
	// belongs to has already gone, and the address it is going to is nowhere in the frame.
	// The same is true of the R3000A's load delay: the value is in flight, not in the
	// register file, and the frame has no room for it either.
	//
	// Both resolve in a single instruction, so declining here costs the interrupt one step
	// of latency and buys the guarantee that every context this machine ever saves is one
	// it can restore exactly. It is the same guarantee the hardware gives itself with the
	// BD bit, arrived at from the other side.
	if st := p.CPU.SaveState(); st.PendingDelay || st.LdReg != 0 {
		return
	}

	for irq := uint32(0); irq < iopIRQs; irq++ {
		bit := uint64(1) << uint64(irq)
		if p.pending&bit == 0 || p.imask&bit == 0 || p.handlers[irq].fn == 0 {
			continue
		}
		p.pending &^= bit
		p.delivered[irq]++
		p.intrDeliver(irq)
		return
	}
}

// intrDeliver is the exception path: save the interrupted thread into a frame on its own
// stack, run the handler, ask THREADMAN whether the interrupt has made a different thread
// the one that ought to be running, and resume whichever thread it names.
//
// The frame is not a Go struct, and that is the whole point. It lives in IOP memory, in
// the layout THREADMAN builds for a thread it starts (iopframe.go), because THREADMAN is
// going to be handed a pointer to it and is going to file it in the outgoing thread's
// control block. A context saved anywhere else would be a context Sony's scheduler could
// not put away — which is why the hooks sat recorded and uncalled for so long.
func (p *IOP) intrDeliver(irq uint32) {
	h := p.handlers[irq]

	// The frame goes below the interrupted thread's stack pointer, which is where the
	// kernel's exception prologue puts it and where THREADMAN's stack-overflow check
	// expects to find it.
	frame := (p.CPU.Reg(29) - iopFrameSize) &^ 7
	p.saveFrame(frame)

	// Whether the interrupted context is a thread at all.
	//
	// A module's entry point is not one. loadcore calls it, and here "loadcore" is Go, so
	// the entry runs as a nested interpreter loop on a stack of the machine's own —
	// whereas on the board it would be running on whichever thread called LoadModule, and
	// THREADMAN would know all about it. That difference is invisible until the scheduler
	// tries to *put the context away*: the switch routine files the frame in the control
	// block of the thread it believes is current, and the thread it believes is current is
	// not the one that is actually running. The frame goes to the wrong owner, the entry
	// point is never resumed, and the machine wanders into the low 64 KiB and executes
	// zeroes.
	//
	// So while a module is starting, the interrupt is delivered — the handler runs, the
	// DMA completion is seen, the flag the entry is spinning on gets set — but the thread
	// switch is declined. Nothing is lost by declining it: the interrupt has already made
	// its thread ready, and THREADMAN will pick it up at the next scheduling point, which
	// is the moment the entry point returns and the machine goes back to stepping threads.
	//
	// The honest fix is to give the loader a thread of its own, so there is always a real
	// context to file. That is a change to how modules are started, not to how interrupts
	// are delivered, and it belongs with the work that turns the IOP on in the main boot.
	preemptible := p.callDepth == 0

	wasEnabled := p.intrEnabled
	p.inIntr++
	p.intrEnabled = false // a handler runs masked, as on the hardware

	p.CPU.SetReg(4, h.arg) // the handler's one argument, given at registration
	_, err := p.callGuestOn(h.fn, iopIntrStack)

	resume := frame
	if err == nil && preemptible {
		if next, ok := p.intrReschedule(frame); ok {
			resume = next
		}
	}

	p.inIntr--
	p.loadFrame(resume)

	// The interrupt returns to the state it found, and *this* is the line that restores
	// it — not the handler, and not the scheduler hooks. Both of those are kernel code
	// that brackets its own critical sections with CpuSuspendIntr and CpuResumeIntr, and
	// both are entitled to leave the processor masked when they hand back, because on the
	// board what re-enables interrupts is the exception return itself, restoring the
	// status register the exception saved. There is no status register here, so the
	// restore has to be written down.
	//
	// Getting this the wrong way round — restoring the enable and *then* running the
	// hooks — costs the machine every interrupt after the first one that reschedules. The
	// timer goes on raising its line thirteen hundred times and not one of them is
	// delivered, because the scheduler quietly left the door shut on its way out.
	p.intrEnabled = wasEnabled

	if err != nil {
		p.halt("the handler for interrupt %d (%s) did not return: %v", irq, p.Sym(h.fn), err)
	}
}

// intrReschedule asks THREADMAN whether to switch threads, and switches if it says so.
// It returns the frame the processor should resume from.
//
// The predicate and the switch routine are THREADMAN's own, registered through intrman
// #30 and #28. The kernel does not need to understand the scheduler — it only has to ask
// the question the scheduler asked to be asked, and to believe the answer.
func (p *IOP) intrReschedule(frame uint32) (uint32, bool) {
	if p.schedResched == 0 || p.schedSwitch == 0 {
		return 0, false
	}
	want, err := p.callGuestOn(p.schedResched, iopIntrStack)
	if err != nil || want == 0 {
		return 0, false
	}

	// Hand the switch routine the frame we saved. It files it in the outgoing thread's
	// control block, makes the incoming thread current, and returns *its* frame.
	p.CPU.SetReg(4, frame)
	next, err := p.callGuestOn(p.schedSwitch, iopIntrStack)
	if err != nil {
		p.halt("THREADMAN's thread switch did not return: %v", err)
		return 0, false
	}
	if next == 0 {
		return 0, false
	}
	p.switches++
	return next, true
}

// tick advances the second processor's own clocks by one instruction: it finishes any
// DMA whose latency has run out, and delivers any interrupt that has become deliverable.
// Every loop that steps the IOP goes through it.
func (p *IOP) tick() {
	p.ioTraceFlush()
	p.steps++

	// The dead zone. Below the first module is the kernel's own memory, and on this machine
	// it is empty — which is worse than it sounds, because a word of zero is a `nop` on
	// MIPS. A processor that jumps to null does not fault here; it nops its way up through
	// sixty-four kilobytes of nothing and falls into the first module's code sideways, and
	// the boot goes on being wrong for a very long time. Catching the jump is the only way
	// to hear about it, and the address it came *from* is the only thing worth knowing.
	if pc := iopPhys(p.CPU.PC); pc < iopModuleBase && pc != iopIdleLoop && pc != iopIdleLoop+4 {
		p.halt("jumped into the empty kernel area at 0x%08X, from %s", p.CPU.PC, p.Sym(p.lastPC))
		return
	}
	p.lastPC = p.CPU.PC
	if p.steps%iopProfileEvery == 0 {
		if p.prof == nil {
			p.prof = map[string]int{}
		}
		p.prof[p.symFunc(p.CPU.PC)]++
	}
	p.timerTick()
	if len(p.dmaPending) > 0 {
		p.dmaTick()
	}
	if p.pending != 0 {
		p.serviceIntr()
	}
}

func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// --- sysmem -----------------------------------------------------------------------

// iopBlock is one allocation, remembered so that the two functions that ask about a
// block — its base and its size — can answer.
type iopBlock struct {
	base, size uint32
}

// sysmemAlloc is AllocSysMemory(mode, size, addr).
//
// The mode chooses where in the free list the block comes from; this allocator has no
// free list, so it bumps. The one part of the mode that cannot be ignored is an
// explicit address, because a caller that asked for a *particular* address is a caller
// that will use it whatever we return, so if one is given it is honoured.
func (p *IOP) sysmemAlloc() {
	size, addr := p.arg(1), p.arg(2)

	base := addr
	if base == 0 {
		base = p.alloc(size)
	}
	if base == 0 {
		p.setRet(0)
		return
	}
	p.blocks = append(p.blocks, iopBlock{base: base, size: size})
	p.setRet(base)
}

// sysmemFree succeeds and does nothing. Nothing on this disc's boot path frees memory
// it will miss: the modules allocate their threads and buffers once and keep them, and
// a 2 MiB arena that is never reclaimed is one that never moves, which during bring-up
// is worth more than the memory.
func (p *IOP) sysmemFree() { p.setRet(0) }

// sysmemBlockTop and sysmemBlockSize answer for the block containing a pointer.
//
// THREADMAN calls them as a pair and stores the results as a thread's stack base and
// stack size — the same two slots it fills from AllocSysMemory's return and its size
// argument on the path where it allocates the stack itself. That is what says which of
// the two is which.
func (p *IOP) sysmemBlockTop() {
	if b, ok := p.blockOf(p.arg(0)); ok {
		p.setRet(b.base)
		return
	}
	p.setRet(0)
}

func (p *IOP) sysmemBlockSize() {
	if b, ok := p.blockOf(p.arg(0)); ok {
		p.setRet(b.size)
		return
	}
	p.setRet(0)
}

func (p *IOP) blockOf(ptr uint32) (iopBlock, bool) {
	for _, b := range p.blocks {
		if ptr >= b.base && ptr < b.base+b.size {
			return b, true
		}
	}
	return iopBlock{}, false
}

// sysmemKprintf is the kernel's printf, and the IOP's other voice.
func (p *IOP) sysmemKprintf() {
	s := p.formatArgs(p.CString(p.arg(0)), 1)
	p.ps2.iopPrint(s)
	p.setRet(uint32(len(s)))
}

// --- heaplib ----------------------------------------------------------------------

// iopHeap is one heap: a chunk taken from the allocator and handed out of, and grown
// with another chunk when it runs out.
//
// The growth is the whole point, and it was learned the hard way. THREADMAN creates its
// heap with a *chunk* size, not a total — 2 KiB — and then takes every thread control
// block, every semaphore and every event flag in the machine out of it. A heap that
// stops at its first chunk runs dry about the time OVERLORD starts up, and what you see
// then is not "out of memory": it is CreateSema quietly returning an error code, and
// OVERLORD, which checks, jumping to itself forever. Two instructions, no message, in a
// module whose sound initialisation is the last thing it printed.
type iopHeap struct {
	chunk uint32 // how much to take from sysmem each time it grows
	base  uint32 // the current chunk
	size  uint32
	ptr   uint32 // the bump pointer within it
	total uint32 // handed out, over the heap's life
}

// heapCreate is CreateHeap(size, flags). It returns the heap's address, which is also
// the handle every later call passes back.
func (p *IOP) heapCreate() {
	chunk := p.arg(0)
	base := p.alloc(chunk)
	if base == 0 {
		p.setRet(0)
		return
	}
	p.heaps[base] = &iopHeap{chunk: chunk, base: base, size: chunk, ptr: base}
	p.setRet(base)
}

// heapAlloc is AllocHeapMemory(heap, size). The caller clears what it gets, so this
// does not have to.
func (p *IOP) heapAlloc() {
	h, size := p.heaps[p.arg(0)], p.arg(1)
	if h == nil {
		// A heap we never made. THREADMAN takes every thread block out of one, so a null
		// here would be a thread that silently does not exist — say so instead.
		p.halt("AllocHeapMemory from heap 0x%08X, which was never created", p.arg(0))
		p.setRet(0)
		return
	}

	a := (h.ptr + 7) &^ 7
	if a+size > h.base+h.size {
		// Grow it. A fresh chunk, at least as big as the request.
		n := h.chunk
		if size > n {
			n = size
		}
		base := p.alloc(n)
		if base == 0 {
			p.setRet(0)
			return
		}
		h.base, h.size, h.ptr = base, n, base
		a = base
	}
	h.ptr = a + size
	h.total += size
	p.setRet(a)
}

// --- sysclib ----------------------------------------------------------------------

// clibMemset is memset(dst, c, n).
func (p *IOP) clibMemset() {
	dst, c, n := p.arg(0), p.arg(1), p.arg(2)
	for i := uint32(0); i < n; i++ {
		p.Write(dst+i, byte(c))
	}
	p.setRet(dst)
}

// --- the kernel's own syscall --------------------------------------------------------

// iopSyscallReschedule is the service number THREADMAN asks the kernel for, and the only
// one anything on this disc asks for at all.
//
// It is the missing half of the scheduler, and its absence was invisible for a long time
// because of an accident of the instruction set. THREADMAN has exactly one `syscall` in
// twenty-eight kilobytes of code:
//
//	THREADMAN+0x46C   addiu $v0, $zero, 32
//	THREADMAN+0x470   syscall
//	THREADMAN+0x474   jr $ra
//
// — a three-instruction exported function that takes no arguments, returns nothing, and
// asks the kernel for service 32. The interrupt-exit path already had the machinery this
// needs: a predicate that says whether a switch is wanted and a routine that performs one,
// both of them THREADMAN's own, registered through intrman (iopframe.go). What was missing
// was the *other* door into it — the one a thread uses when it gives up the processor of
// its own accord rather than being taken off it. A thread that blocks on a semaphore does
// not wait for a timer; it asks to be rescheduled now.
//
// With no exception handler at 0x80000080 — and there is none, because intrman is Go — the
// core took the exception into empty memory. And a word of zero is a `nop` on MIPS, so the
// processor did not fault. It nopped its way up through sixty-four kilobytes of nothing and
// fell sideways into the first module's code, and the boot went on, wrongly, for a very
// long time. It took putting an idle loop in that empty memory to hear about it at all.
const iopSyscallReschedule = 32

// kernelSyscall serves a `syscall` whose service number is in $v0.
func (p *IOP) kernelSyscall() bool {
	switch svc := p.CPU.Reg(2); svc {
	case iopSyscallReschedule:
		p.yield()

	default:
		// Counted, not guessed at, and above all not taken as an exception into memory that
		// does not fault. The census is the work list.
		name := fmt.Sprintf("kernel syscall %d", svc)
		p.unmodelledCalls[name]++
		if p.unmodelledCalls[name] == 1 {
			p.ps2.note("IOP: %s from %s — unmodelled", name, p.Sym(p.CPU.CurPC()))
		}
	}
	return true
}

// yield is a thread giving up the processor: the synchronous twin of intrDeliver's exit
// path, and the same steps without the handler in front of them.
//
// The program counter saved is the instruction *after* the syscall, which is where the core
// has already advanced to — so a thread resumed from this frame comes back to the `jr $ra`
// and returns to whoever asked to be rescheduled, with no idea it was ever away.
func (p *IOP) yield() {
	frame := (p.CPU.Reg(29) - iopFrameSize) &^ 7
	p.saveFrame(frame)

	if next, ok := p.intrReschedule(frame); ok {
		p.loadFrame(next)
	}
	// And if the scheduler does not want a switch, nothing at all happens: the processor is
	// already running the thread it ought to be, and the frame just written is so much dead
	// wood below the stack pointer.
}
