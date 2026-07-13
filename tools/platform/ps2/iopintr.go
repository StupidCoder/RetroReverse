package ps2

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
	// The two that will matter most later are #28 and #30. THREADMAN hands each of them
	// one of its own functions, and the function it hands to #30 is four instructions
	// long: it compares two of THREADMAN's globals and returns whether they differ.
	// That is "does the scheduler want to switch threads?" — so #30 registers the
	// predicate the kernel consults on the way out of an interrupt, and #28 registers
	// the routine that acts on it. They are recorded rather than called, because nothing
	// delivers an interrupt to the IOP yet; when something does, this is the contract it
	// has to honour, and honouring it is what will make the real THREADMAN schedule.
	lib("intrman", map[uint16]iopFunc{
		4:  {"RegisterIntrHandler", (*IOP).intrRegister},
		5:  {"ReleaseIntrHandler", (*IOP).intrRelease},
		6:  {"EnableIntr", (*IOP).intrEnable},
		7:  {"DisableIntr", (*IOP).intrDisable},
		8:  unknown(),
		9:  unknown(),
		14: unknown(),
		17: {"CpuSuspendIntr", (*IOP).intrSuspend},
		18: {"CpuResumeIntr", (*IOP).intrResume},
		23: unknown(),
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

// iopIRQs is how many interrupt lines the IOP has. The modules on this disc register
// handlers on lines up to 0x10; the controller has 32 and there is no reason to model
// fewer.
const iopIRQs = 32

// iopHandler is one registered interrupt handler.
type iopHandler struct {
	fn      uint32 // the module's own routine
	arg     uint32
	enabled bool
}

func (p *IOP) intrRegister() {
	irq, _, fn, arg := p.arg(0), p.arg(1), p.arg(2), p.arg(3)
	if irq >= iopIRQs {
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
		p.handlers[irq].enabled = true
	}
	p.setRet(0)
}

func (p *IOP) intrDisable() {
	if irq := p.arg(0); irq < iopIRQs {
		p.handlers[irq].enabled = false
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

// intrSetSwitchHook and intrSetReschedHook record THREADMAN's two hooks into the
// interrupt-exit path. See the note at the top of the file: they are stored, not
// called, because nothing delivers an interrupt to this IOP yet. When something does,
// the contract is: run the handler, ask the predicate, and if it says yes, call the
// switcher.
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

// iopHeap is one heap: a slab taken from the allocator and handed out of.
type iopHeap struct {
	base, size, ptr uint32
}

// heapCreate is CreateHeap(size, flags). It returns the heap's address, which is also
// the handle every later call passes back.
func (p *IOP) heapCreate() {
	size := p.arg(0)
	base := p.alloc(size)
	if base == 0 {
		p.setRet(0)
		return
	}
	p.heaps[base] = &iopHeap{base: base, size: size, ptr: base}
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
		p.setRet(0) // the heap is full; the caller checks
		return
	}
	h.ptr = a + size
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
