package xbox

// kernel.go is the xboxkrnl.exe HLE. A title imports kernel exports by *ordinal*: its
// IAT (the thunk table at XBE.ThunkAddr, in .rdata) is an array of DWORDs, each with
// the high bit set and the ordinal in its low 16 bits. The loader would overwrite each
// slot with the real function address; we overwrite it with a unique sentinel in the
// trap region instead (patchThunks). When title code does `CALL DWORD PTR [slot]` the
// PC lands on the sentinel; onStep recognises the trap range, decodes the ordinal, and
// dispatches to a Go handler — the x86 analogue of the PSP's `jr $ra; syscall` stub
// patch (tools/platform/psp/kernel.go).
//
// Calling convention. xboxkrnl exports are __stdcall (the callee pops its arguments),
// except the few varargs ones (DbgPrint) which are cdecl. A handler reads its
// arguments from the stack (arg(0) = [ESP+4], the word above the return address),
// writes its result to EAX, and returns the number of DWORD arguments to pop; the
// dispatcher then simulates the RET: pop the return address, add argWords*4 to ESP for
// stdcall, and resume at the caller. The sentinel bytes are never fetched.
//
// Three tiers, after the PSP/3DS HLEs: modelled (memory, critical sections, threads,
// the file and object calls the boot path reads the results of), stubbed (report
// success so single-threaded init proceeds), and halt-on-unknown — an ordinal with no
// handler stops the machine, naming itself, so each run states exactly how far the
// boot reached. The ordinal->name map is general Xbox-platform ABI (the sanctioned
// clean-room exception, like the PSP KIRK keys); every game fact stays image-derived.

import (
	"fmt"

	"retroreverse.com/tools/cpu/x86"
)

// KPCR layout. FS points here on Xbox; the CRT and kernel read the exception-handler
// list, the stack bounds, and the current thread through it.
const (
	kpcrSize          = 0x2C0
	kpcrExceptionList = 0x00 // NT_TIB.ExceptionList (0xFFFFFFFF = end of chain)
	kpcrStackBase     = 0x04
	kpcrStackLimit    = 0x08
	kpcrSelfPcr       = 0x1C // KPCR.SelfPcr -> the KPCR's own address
	kpcrPrcb          = 0x20 // KPCR.Prcb -> the (inline) KPRCB
	kpcrIrql          = 0x24 // current IRQL (0 = PASSIVE_LEVEL)
	kpcrPrcbData      = 0x28 // the inline KPRCB begins here
	prcbCurrentThread = 0x04 // KPRCB.CurrentThread
)

// setupKPCR builds the processor control region FS points at.
func (m *Machine) setupKPCR() {
	m.write32(kpcrAddr+kpcrExceptionList, 0xFFFFFFFF)
	m.write32(kpcrAddr+kpcrStackBase, titleStackTop)
	m.write32(kpcrAddr+kpcrStackLimit, titleStackTop-titleStackSize)
	m.write32(kpcrAddr+kpcrSelfPcr, kpcrAddr)
	m.write32(kpcrAddr+kpcrPrcb, kpcrAddr+kpcrPrcbData)
	m.write32(kpcrAddr+kpcrIrql, 0)
	// CurrentThread is filled in once the boot thread's KTHREAD exists (bootThread).
}

// patchThunks walks the IAT and rewrites each ordinal import. A function export points
// at its code trap; a data export points at a populated block in the kernel band, so a
// dereference of the slot reads coherent data rather than faulting in the trap region.
func (m *Machine) patchThunks() {
	code, data := 0, 0
	for p := m.XBE.ThunkAddr; ; p += 4 {
		v := m.read32(p)
		if v == 0 {
			break // NUL terminator
		}
		if v&0x80000000 == 0 {
			continue // a by-name import (does not occur for xboxkrnl); leave it
		}
		ord := uint16(v & 0xFFFF)
		if size, isData := dataExportSize(ord); isData {
			addr := m.allocKObject(uint32(size))
			m.initDataExport(ord, addr)
			m.write32(p, addr)
			data++
		} else {
			m.write32(p, trapBase+uint32(ord)*trapStride)
			code++
		}
	}
	m.logf("kernel: patched %d function + %d data import thunks", code, data)
}

// onStep is the CPU's per-instruction hook. Its first duty is to catch a call that has
// landed on a kernel trap sentinel and service the ordinal; otherwise it advances the
// synthetic clock and lets the instruction run. The scheduler's quantum accounting
// lives here too (sched.go).
func (m *Machine) onStep(c *x86.CPU) {
	pc := c.SegBase[x86.CS] + c.IP
	if pc >= trapBase && pc < trapTop {
		ord := uint16((pc - trapBase) / trapStride)
		m.dispatchKernel(ord)
		return
	}
	if pc == threadExitAddr {
		m.exitCurrentThread()
		return
	}
	if m.traceLeft > 0 {
		m.traceLeft--
		fmt.Printf("%08X  %s\n", pc, m.disasmAt(pc))
	}
	m.tick++
	m.schedTick()
}

// dispatchKernel services one kernel import: run its handler, then simulate the return.
func (m *Machine) dispatchKernel(ord uint16) {
	m.OrdinalHits[ord]++
	name := ordinalName(ord)
	h := kernelHandler(ord)
	if h == nil {
		m.CPU.Halt("unimplemented xboxkrnl ordinal %d (%s), called from %08X",
			ord, name, m.retAddr())
		m.Halted, m.HaltReason = true, m.CPU.HaltReason
		return
	}
	argWords := h(m)
	m.kret(argWords)
	// A handler that blocked (a wait that could not be satisfied) set reschedule; the
	// return PC is already saved into this thread's context by kret, so switching now
	// parks it cleanly and resumes it after the call once woken.
	if m.reschedule {
		m.reschedule = false
		m.dispatch()
	}
}

// retAddr is the return address on the stack at a kernel-trap entry ([ESP]).
func (m *Machine) retAddr() uint32 { return m.read32(m.CPU.Regs[x86.SP]) }

// arg reads stdcall argument i (i=0 is the first, at [ESP+4]).
func (m *Machine) arg(i int) uint32 { return m.read32(m.CPU.Regs[x86.SP] + 4 + uint32(i)*4) }

// setRet writes the EAX return value.
func (m *Machine) setRet(v uint32) { m.CPU.Regs[x86.AX] = v }

// kret simulates the stdcall return: pop the return address into EIP and drop the
// callee's argWords arguments from the stack. A cdecl handler returns 0 (the caller
// cleans up).
func (m *Machine) kret(argWords int) {
	sp := m.CPU.Regs[x86.SP]
	ret := m.read32(sp)
	m.CPU.Regs[x86.SP] = sp + 4 + uint32(argWords)*4
	m.CPU.IP = ret
}

// --- helpers -------------------------------------------------------------

// allocPool bump-allocates size bytes of physical RAM from the down-growing
// contiguous/pool arena, 16-byte aligned. Returns 0 on exhaustion.
func (m *Machine) allocPool(size uint32) uint32 { return m.allocPoolAligned(size, 16) }

// allocPoolAligned is allocPool with an explicit result alignment — the Alignment
// argument MmAllocateContiguousMemoryEx carries. The arena grows DOWN, so the base is
// (poolNext-size) rounded down to the alignment. An alignment of 0 means the default 16.
// Returns 0 on exhaustion.
func (m *Machine) allocPoolAligned(size, align uint32) uint32 {
	if align < 16 {
		align = 16
	}
	size = align32(size, 16)
	if size == 0 || size > m.poolNext {
		return 0
	}
	base := (m.poolNext - size) &^ (align - 1)
	if base < m.heapNext {
		return 0
	}
	m.poolNext = base
	return base
}

// allocVirtual bump-allocates from the up-growing heap arena, page aligned.
func (m *Machine) allocVirtual(size uint32) uint32 {
	size = align32(size, 0x1000)
	if size == 0 || m.heapNext+size > m.heapTop {
		return 0
	}
	a := m.heapNext
	m.heapNext += size
	return a
}

// allocKObject carves a dispatcher/kernel object out of the reserved kernel band and
// returns its guest address (which also serves as its handle).
func (m *Machine) allocKObject(size uint32) uint32 {
	size = align32(size, 16)
	a := m.nextObjAddr
	m.nextObjAddr += size
	return a
}

// --- the modelled handler set --------------------------------------------
//
// A handler reads its arguments, does its work, sets EAX, and returns the number of
// DWORD arguments to pop (stdcall). Bind an ordinal here only when its number->function
// identity and its argument count are certain; everything else halts and names itself,
// which is the intended concrete frontier.

// kernelHandler returns the handler for an ordinal, or nil to halt-on-unknown. The
// dispatcher-object / memory / sync surface lives in kernel_objects.go; the core set
// below covers the remaining boot-touched exports.
func kernelHandler(ord uint16) func(*Machine) int {
	if h := kernelObjectHandler(ord); h != nil {
		return h
	}
	switch ord {
	case 8: // DbgPrint(format, ...) — cdecl/varargs. Stack-safe even if the low-block
		// numbering is off, because a cdecl callee pops nothing (the caller cleans up).
		return func(m *Machine) int {
			m.logf("DbgPrint: %s", m.cstr(m.arg(0)))
			m.setRet(0)
			return 0
		}
	case 24: // ExQueryNonVolatileSetting(Index, Type*, Value*, Length, ResultLength*)
		// Verified: a 5-arg tail-call from the CRT reading system config (the index
		// 0x11 setting). Return STATUS_SUCCESS with a zeroed value — no configured
		// EEPROM here, so every setting reads as its default.
		return func(m *Machine) int {
			if p := m.arg(1); p != 0 {
				m.write32(p, 4) // Type = REG_DWORD-ish
			}
			if p := m.arg(2); p != 0 {
				m.write32(p, 0) // the setting value (default 0)
			}
			if p := m.arg(4); p != 0 {
				m.write32(p, 4) // ResultLength
			}
			m.setRet(0) // STATUS_SUCCESS
			return 5
		}
	case 47: // HalRegisterShutdownNotification(&HAL_SHUTDOWN_REGISTRATION, Register)
		// Verified from its call site: 2 stdcall args, arg1 = TRUE, returns void. We do
		// not run shutdown, so recording the registration is unnecessary — accept it.
		return func(m *Machine) int { m.setRet(0); return 2 }

	case 255: // PsCreateSystemThreadEx — verified from the CRT's 10-arg call pattern
		// (ThreadHandle, ThreadExtSize, KernelStackSize, TlsDataSize, ThreadId,
		//  StartContext1, StartContext2, CreateSuspended, DebuggerThread, StartRoutine)
		return func(m *Machine) int {
			handleOut := m.arg(0)
			stackSize := m.arg(2)
			threadIDOut := m.arg(4)
			ctx1 := m.arg(5)
			ctx2 := m.arg(6)
			suspended := m.arg(7) != 0
			startRoutine := m.arg(9)
			t := m.createThread(startRoutine, ctx1, ctx2, stackSize, 16, suspended)
			if handleOut != 0 {
				m.write32(handleOut, t.kthread) // the KTHREAD address doubles as the handle
			}
			if threadIDOut != 0 {
				m.write32(threadIDOut, t.id)
			}
			m.setRet(0) // STATUS_SUCCESS
			return 10
		}
	}
	// Everything else halts and names itself: the concrete boot frontier. Ordinal
	// semantics are added here one at a time, each verified against the live call
	// (its stack arguments and how the caller consumes the result) rather than
	// pre-guessed — the honest bring-up loop.
	return nil
}
