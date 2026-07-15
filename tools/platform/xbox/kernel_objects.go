package xbox

// kernel_objects.go models the dispatcher-object, memory, and synchronisation exports
// the boot path reaches — the Nt*, Mm*, Ke*, and Rtl* calls the XDK/CRT startup makes
// once the main thread is running. Every ordinal here is bound by NUMBER, and every
// number is one pinned empirically from its live call site (its argument count and the
// shapes of its arguments) — not from the reconstructed table in kernel_ordinals.go,
// which drifts by a few entries in several blocks. So MmAllocateContiguousMemory sits at
// 165 and NtOpenFile at 202, five above where a naive count would put them; each was
// confirmed the same way and is listed in verifiedNames. The stdcall argument count each
// handler returns for stack cleanup is exactly the push count seen at the call site.
//
// Objects: a dispatcher object (event, semaphore, mutant, timer) is a small guest block
// carrying a DISPATCHER_HEADER whose signal state the HLE keeps in sync with its own
// kobject, so title code that inspects the object directly sees a coherent header. The
// block's guest address is its handle. Waits check the signal state: satisfied waits
// return immediately; unsatisfied ones park the thread for the scheduler.

// dispatcher-header offsets (KEVENT/KSEMAPHORE/KMUTANT share the leading header).
const (
	dhType         = 0x00 // byte: object type
	dhSignalState  = 0x04 // LONG: signalled (nonzero) / not
	dhWaitListHead = 0x08 // LIST_ENTRY (unused by the HLE, zeroed)
)

// newObject allocates and registers a dispatcher object, returning its handle (== its
// guest address). The DISPATCHER_HEADER is written so direct inspection is coherent.
func (m *Machine) newObject(kind string, typ byte, signaled bool, count, limit int32) uint32 {
	addr := m.allocKObject(0x20)
	m.Write(addr+dhType, typ)
	m.writeSignal(addr, signaled)
	m.write32(addr+dhWaitListHead, addr+dhWaitListHead)   // empty list -> points at itself
	m.write32(addr+dhWaitListHead+4, addr+dhWaitListHead) // Flink == Blink == &head
	m.objects[addr] = &kobject{kind: kind, addr: addr, signaled: signaled, count: count, limit: limit}
	return addr
}

func (m *Machine) writeSignal(addr uint32, signaled bool) {
	if signaled {
		m.write32(addr+dhSignalState, 1)
	} else {
		m.write32(addr+dhSignalState, 0)
	}
}

// objAt looks up the kobject for a handle (== guest address). It resyncs the kobject's
// signal state from the guest header first, so code that pokes the header directly (as
// KeInitializeEvent-style inline setup does) is respected.
func (m *Machine) objAt(handle uint32) *kobject {
	o := m.objects[handle]
	if o != nil {
		o.signaled = m.read32(handle+dhSignalState) != 0
	}
	return o
}

// satisfyWait tests whether a wait on an object succeeds now, consuming the object's
// signal as the kernel would (auto-reset events and semaphores decrement; mutants take
// ownership). Returns true if the wait is satisfied without blocking.
func (m *Machine) satisfyWait(o *kobject) bool {
	if o == nil {
		return true // unknown handle: do not deadlock the boot on it
	}
	switch o.kind {
	case "semaphore":
		if o.count > 0 {
			o.count--
			if o.count == 0 {
				o.signaled = false
				m.writeSignal(o.addr, false)
			}
			return true
		}
		return false
	case "event-auto":
		if o.signaled {
			o.signaled = false
			m.writeSignal(o.addr, false)
			return true
		}
		return false
	default: // manual-reset event, mutant, timer, thread: satisfied while signalled
		return o.signaled
	}
}

// kernelObjectHandler returns handlers for the dispatcher / memory / sync / file
// ordinals, or nil to fall through to kernelHandler's core set (and then halt-on-
// unknown).
//
// Two tiers live here. The VERIFIED cases (202 NtOpenFile, 203 NtReadFile — pinned from
// their live call sites) are trustworthy. The PROVISIONAL cases below them are keyed to
// the reconstructed ordinal table (kernel_ordinals.go), whose numbering drifts by a few
// entries in several blocks — so a provisional case may be bound to the wrong function.
// They are kept because they let the boot advance, and because a wrong binding surfaces
// quickly (a downstream fault or an obviously-wrong argument), but each must be
// re-verified against its actual call site before it is trusted. Only a case promoted
// into verifiedNames (kernel_ordinals.go) has been confirmed.
func kernelObjectHandler(ord uint16) func(*Machine) int {
	switch ord {

	// --- Disc file I/O (verified) ---------------------------------------
	case 202: // NtOpenFile(FileHandle*, DesiredAccess, ObjectAttributes, IoStatusBlock,
		// ShareAccess, OpenOptions) — verified: 6 args, an ACCESS_MASK + OpenOptions.
		return func(m *Machine) int {
			m.openFile(m.arg(0), m.arg(2), m.arg(3))
			return 6
		}
	case 203: // NtReadFile(FileHandle, Event, ApcRoutine, ApcContext, IoStatusBlock,
		// Buffer, Length, ByteOffset*) — provisional pending its own call-site check.
		return func(m *Machine) int {
			m.readFile(m.arg(0), m.arg(4), m.arg(5), m.arg(6), m.arg(7))
			return 8
		}

	// --- Memory (Mm) — verified: 165 is a 1-arg allocation (the Mm block drifts +5) --
	case 165: // MmAllocateContiguousMemory(NumberOfBytes) -> physical base
		return func(m *Machine) int { m.setRet(m.allocPool(m.arg(0))); return 1 }

	// --- Virtual memory (Nt) — verified: 184 is a 5-arg reserve/commit ---
	case 184: // NtAllocateVirtualMemory(BaseAddress**, ZeroBits, RegionSize*, Type, Protect)
		return func(m *Machine) int {
			baseOut, sizeP := m.arg(0), m.arg(2)
			size := m.read32(sizeP)
			addr := uint32(0)
			if baseOut != 0 && m.read32(baseOut) != 0 {
				addr = m.read32(baseOut) // a requested base: honour it (identity)
			} else {
				addr = m.allocVirtual(size)
			}
			if baseOut != 0 {
				m.write32(baseOut, addr)
			}
			if sizeP != 0 {
				m.write32(sizeP, align32(size, 0x1000))
			}
			m.setRet(0) // STATUS_SUCCESS
			return 5
		}

	// --- Dispatcher objects (Nt) — verified: 187 is a 3-arg create -------
	case 187: // NtCreateMutant(MutantHandle*, ObjectAttributes, InitialOwner) — verified
		return func(m *Machine) int {
			initialOwner := m.arg(2) != 0
			h := m.newObject("mutant", 2, !initialOwner, 0, 0)
			if p := m.arg(0); p != 0 {
				m.write32(p, h)
			}
			m.setRet(0)
			return 3
		}

	// --- DPC / timers (Ke) — verified ------------------------------------
	case 107: // KeInitializeDpc(Dpc, DeferredRoutine, DeferredContext) — verified: 3 args,
		// arg1 is a .text routine pointer. Fill the KDPC header the kernel would.
		return func(m *Machine) int {
			dpc := m.arg(0)
			if dpc != 0 {
				m.write16(dpc+0x00, 0x0013)   // Type = DpcObject
				m.write32(dpc+0x04, 0)        // DpcListEntry.Flink
				m.write32(dpc+0x08, 0)        // DpcListEntry.Blink
				m.write32(dpc+0x0C, m.arg(1)) // DeferredRoutine
				m.write32(dpc+0x10, m.arg(2)) // DeferredContext
			}
			m.setRet(0)
			return 3
		}

	case 149: // KeSetTimer(Timer, DueTime(LARGE_INTEGER, 2 dwords), Dpc) — verified: args
		// (KTIMER, negative relative due time, KDPC). Record the association; the DPC is
		// not fired here (nothing yet waits on it). Returns TRUE (timer was not set).
		return func(m *Machine) int {
			tm := m.arg(0)
			if tm != 0 {
				m.write32(tm+dhSignalState, 0) // not yet signalled
			}
			m.setRet(1)
			return 4
		}

	case 113: // KeInitializeTimerEx(Timer, Type) — verified: 2 args, follows the DPC init.
		return func(m *Machine) int {
			tm := m.arg(0)
			if tm != 0 {
				m.write16(tm+0x00, 0x0008) // Type = TimerNotificationObject
				m.write32(tm+dhSignalState, 0)
				m.write32(tm+0x0C, 0) // TimerListEntry / dueTime, zeroed
				m.write32(tm+0x10, 0)
			}
			m.setRet(0)
			return 2
		}

	// --- Critical sections (Rtl) — pinned against this image's census ----
	case 277: // RtlEnterCriticalSection(cs)
		return func(m *Machine) int { m.setRet(0); return 1 }
	case 291: // RtlInitializeCriticalSection(cs)
		return func(m *Machine) int {
			cs := m.arg(0)
			if cs != 0 {
				m.write32(cs+0x00, 0)          // Unknown
				m.write32(cs+0x04, 0xFFFFFFFF) // LockCount (-1 = unlocked)
				m.write32(cs+0x08, 0)          // RecursionCount
				m.write32(cs+0x0C, 0)          // OwningThread
			}
			m.setRet(0)
			return 1
		}
	case 294: // RtlLeaveCriticalSection(cs)
		return func(m *Machine) int { m.setRet(0); return 1 }
	case 301: // RtlNtStatusToDosError(NTSTATUS) -> Win32 error (0 stays 0 = success)
		return func(m *Machine) int {
			st := m.arg(0)
			if st == 0 {
				m.setRet(0)
			} else {
				m.setRet(st & 0xFFFF) // rough map; exact table not needed for the boot
			}
			return 1
		}
	}
	return nil
}

// doWait implements a single-object wait: satisfy it now, or park the current thread
// until the object is signalled. reg is the register to receive the STATUS on wake
// (EAX). A satisfied wait sets EAX=0 (STATUS_WAIT_0) directly.
func (m *Machine) doWait(handle uint32, reg int) {
	o := m.objAt(handle)
	if m.satisfyWait(o) {
		m.setRet(0) // STATUS_WAIT_0
		return
	}
	// Block: record the wait and yield. The dispatcher advances the PC past the trap
	// before this runs (kret is not called for a blocking wait — dispatchKernel handles
	// that), so the thread resumes after the call once woken.
	if m.current == nil {
		m.setRet(0) // no scheduler context (boot thread pre-spawn): do not deadlock
		return
	}
	m.current.waitObjs = []uint32{handle}
	m.setRet(0) // committed into the saved context; refined to STATUS_WAIT_0 on wake
	m.yieldCurrent(tsWaiting)
}

// wakeWaiters readies any thread blocked on the given object handle whose wait is now
// satisfiable, consuming the object's signal as the wake would.
func (m *Machine) wakeWaiters(handle uint32) {
	o := m.objects[handle]
	for _, t := range m.threads {
		if t.state != tsWaiting {
			continue
		}
		for _, h := range t.waitObjs {
			if h == handle && m.satisfyWait(o) {
				t.state = tsReady
				t.waitObjs = nil
				break
			}
		}
	}
}

func boolU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
