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

import "retroreverse.com/tools/cpu/x86"

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
	m.logf("newObject: %s @%08X signaled=%v from %08X", kind, addr, signaled, m.retAddr())
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

// guestObjAt adopts a raw DISPATCHER_HEADER living in guest memory (a KEVENT a title
// embeds in its own struct and initialises by hand — the Ke APIs take object addresses,
// not handles) into the object table, so waits and signals on it use the same machinery.
// The header's type byte distinguishes notification (0, stays signalled) from
// synchronisation (1, auto-clears) events; both behave as "event" here since satisfyWait
// consumes the signal exactly as a synchronisation event would — the D3D swap event is
// re-zeroed by its own code before every wait, so the distinction is unobservable.
func (m *Machine) guestObjAt(addr uint32) *kobject {
	if addr == 0 {
		return nil
	}
	if o := m.objAt(addr); o != nil {
		return o
	}
	o := &kobject{kind: "event", addr: addr, signaled: m.read32(addr+dhSignalState) != 0}
	m.objects[addr] = o
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
			m.openFile(m.arg(0), m.arg(2), m.arg(3), dispOpen)
			return 6
		}
	case 190: // NtCreateFile(FileHandle*, DesiredAccess, ObjectAttributes, IoStatusBlock,
		// AllocationSize*, FileAttributes, ShareAccess, CreateDisposition, CreateOptions)
		// -> NTSTATUS. Verified from its call site (0x43D08): nine pushes ending in the
		// out-handle, an access mask OR'd with SYNCHRONIZE|0x80, the OBJECT_ATTRIBUTES and
		// IOSB locals, and the caller comparing the status against 0xC0000035
		// (STATUS_OBJECT_NAME_COLLISION) — XAPI's CreateFile wrapper; table-185 + the Nt
		// block's +5 drift = 190. Disc paths open read-only through the XISO; the HDD
		// title partitions (T:/U:/Z:) honour the CreateDisposition against the writable
		// in-memory store (kernel_file.go) — a real console's cache partition always
		// exists, and OutRun's menu loader can only leave the loading screen once it has
		// built and re-opened z:\MENU.PAK there.
		return func(m *Machine) int {
			m.openFile(m.arg(0), m.arg(2), m.arg(3), m.arg(7))
			return 9
		}

	case 219: // NtReadFile(FileHandle, Event, ApcRoutine, ApcContext, IoStatusBlock,
		// Buffer, Length, ByteOffset*) — verified from its call site (0x440C1): eight args
		// in the OVERLAPPED shape (the caller pre-stores STATUS_PENDING 0x103 into the
		// IOSB's Internal field and passes the same block as ApcContext and IoStatusBlock,
		// Win32 ReadFile-over-lapped style), buffer 0x6B4430, length 0x10000, an explicit
		// ByteOffset pointer; table-214 + the Nt block's +5 drift = 219. The old provisional
		// binding of this function at ordinal 203 (never called — confirmed by the ordinal
		// histogram) is removed; 203 stays a halting frontier until a live site names it.
		return func(m *Machine) int {
			m.readFile(m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4), m.arg(5), m.arg(6), m.arg(7))
			return 8
		}

	case 236: // NtWriteFile(FileHandle, Event, ApcRoutine, ApcContext, IoStatusBlock,
		// Buffer, Length, ByteOffset*) — the write mirror of NtReadFile's 8-arg
		// OVERLAPPED shape. Canonical 236 = table-231 + the Nt block's established +5
		// drift, the same derivation every verified Nt binding has landed on; the first
		// live call (the z:\MENU.PAK cache build) passes a cache-file handle and a
		// buffer/length pair consistent with this shape. Only the writable HDD store
		// accepts writes (kernel_file.go writeFile halts on a disc handle).
		return func(m *Machine) int {
			m.writeFile(m.arg(0), m.arg(1), m.arg(4), m.arg(5), m.arg(6), m.arg(7))
			return 8
		}

	case 246: // ObReferenceObjectByHandle(Handle, ObjectType, Object**) -> NTSTATUS.
		// Verified from its call site (0x45291): three args — a handle, the *value* of a
		// data-export IAT slot (an OBJECT_TYPE pointer, ordinal 71, opaque here), and an
		// out-pointer the caller uses as the object body; table-241 + the Ob block's +5
		// drift = 246. Our handles ARE the guest addresses of their object blocks, so the
		// referenced object is the handle itself; the HLE does not refcount.
		return func(m *Machine) int {
			h, out := m.arg(0), m.arg(2)
			if m.objects[h] == nil {
				m.setRet(0xC0000008) // STATUS_INVALID_HANDLE
				return 3
			}
			if out != 0 {
				m.write32(out, h)
			}
			m.setRet(0)
			return 3
		}

	case 126, 127: // KeQueryPerformanceCounter / KeQueryPerformanceFrequency() -> ULONGLONG.
		// Verified from twin call sites (0x214B80/0x214B94): no argument pushes, the
		// 64-bit result taken from EDX:EAX and stored to the caller's out pointer —
		// exactly QueryPerformanceCounter/Frequency's kernel side; table-121/122 + the
		// Ke block's +5 drift. On hardware both ride the CPU's 733 MHz TSC; here the
		// counter is the machine tick scaled by the same 367 the RDTSC model uses, and
		// the frequency is that counter's true rate (367 x 2000 instrs/ms x 1000 =
		// 734 MHz) so guest-computed delta/frequency seconds match the tick clock.
		ord := ord
		return func(m *Machine) int {
			v := m.tick * 367
			if ord == 127 {
				v = 734000000
			}
			m.CPU.Regs[x86.AX] = uint32(v)
			m.CPU.Regs[x86.DX] = uint32(v >> 32)
			return 0
		}

	case 128: // KeQuerySystemTime(PLARGE_INTEGER CurrentTime) -> void. Verified from its
		// call site (0x214C5A): one pointer argument, the callee fills an 8-byte buffer the
		// caller then serialises verbatim into a token (LEA [EBP-2C]; PUSH; CALL; append 8
		// bytes) — the KeQuerySystemTime shape, not the table's KeRemoveByKeyDeviceQueue.
		// Table-123 + the Ke block's +5 drift = 128. Write the same monotonic 100-ns clock
		// the KeSystemTime data export advances so the two agree.
		return func(m *Machine) int {
			if p := m.arg(0); p != 0 {
				t := m.systemTime100ns()
				m.write32(p, uint32(t))
				m.write32(p+4, uint32(t>>32))
			}
			m.setRet(0)
			return 1
		}

	case 143: // KeSetBasePriorityThread(Thread, Increment) -> LONG (old priority).
		// Verified from its call site (0x44F10): XAPI's SetThreadPriority wrapper —
		// ObReferenceObjectByHandle with the thread type export (slot 0x24836C), the
		// Win32 priority mapped 15->16 / -15->-16, two args, result unread; table-138 +
		// the Ke block's +5 drift = 143. The increment offsets the normal base (16).
		return func(m *Machine) int {
			kt, inc := m.arg(0), int32(m.arg(1))
			old := int32(16)
			if o := m.objects[kt]; o != nil && o.thread != nil {
				old = o.thread.priority
				p := 16 + inc
				if p < 0 {
					p = 0
				} else if p > 31 {
					p = 31
				}
				o.thread.priority = p
			}
			m.setRet(uint32(old))
			return 2
		}

	case 250: // ObfDereferenceObject(Object@ECX) — fastcall, no stack args. Verified from
		// its call site (0x45331: MOV ECX,[EBP+8] then the call, at the tail of the
		// cancel-io path whose head is ObReferenceObjectByHandle); table-245 + the Ob
		// block's +5 drift = 250. The HLE does not refcount: success no-op.
		return func(m *Machine) int { m.setRet(0); return 0 }

	case 226: // NtSetInformationFile(FileHandle, IoStatusBlock, FileInformation, Length,
		// FileInformationClass) -> NTSTATUS. Verified from its call site (0x44378): the
		// same five-arg shape as NtQueryInformationFile but writing — a file handle from
		// the CreateFile wrapper, two locals, length 8, class 0xE (FilePositionInformation,
		// an 8-byte LARGE_INTEGER) — the XAPI SetFilePointer path; table-221 + the Nt
		// block's +5 drift = 226. Only the position class is modelled; others halt.
		return func(m *Machine) int {
			h, iosb, buf, ln, class := m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4)
			fo := m.files[h]
			if fo == nil {
				m.finishOpen(iosb, h, 0, 0xC0000008) // STATUS_INVALID_HANDLE
				return 5
			}
			if class != 0xE || ln < 8 {
				m.CPU.Halt("NtSetInformationFile: unmodelled class %d (len %d) from %08X",
					class, ln, m.retAddr())
				return 5
			}
			fo.off = m.read32(buf) // low dword; disc files stay under 4 GB
			m.finishOpen(iosb, h, 0, 0)
			return 5
		}

	case 224, 231: // NtResumeThread / NtSuspendThread (ThreadHandle, PreviousSuspendCount*).
		// Verified as a pair from twin 2-arg wrappers at 0x44F30/0x44F56 — each passes a
		// handle plus an out-count and maps failure to -1; slots 0x248370/0x248374 hold
		// 231/224, landing exactly on table 226/219 (NtSuspendThread/NtReadFile... the Nt
		// block's +5 drift). NT semantics, faithfully: suspension is a COUNT on the
		// thread, orthogonal to its dispatcher-wait state. Suspend increments; resume
		// decrements; the thread runs only at count 0 AND not blocked. A resume NEVER
		// completes a wait — the earlier model set any tsWaiting thread ready here, which
		// "satisfied" the streaming pump's infinite message-queue wait with an empty
		// queue: it popped a NULL message and double-released a buffer slot (see
		// thread.suspendCount).
		ord := ord
		return func(m *Machine) int {
			h, prevOut := m.arg(0), m.arg(1)
			o := m.objects[h]
			if o == nil || o.thread == nil {
				m.setRet(0xC0000008) // STATUS_INVALID_HANDLE
				return 2
			}
			t := o.thread
			prev := t.suspendCount
			if ord == 224 { // resume
				if t.suspendCount > 0 {
					t.suspendCount--
				}
			} else { // suspend
				t.suspendCount++
				if t == m.current {
					// Park the running thread the moment its own suspend lands; it
					// resumes exactly here (after the trap return) when the count
					// drops to zero. Its state stays tsReady — the wait axis is
					// untouched — and pickRunnable skips it via runnable().
					m.reschedule = true
				}
			}
			if prevOut != 0 {
				m.write32(prevOut, uint32(prev))
			}
			m.setRet(0)
			return 2
		}

	case 211: // NtQueryInformationFile(FileHandle, IoStatusBlock, FileInformation, Length,
		// FileInformationClass) -> NTSTATUS. Verified from its call site (0x445F6): five
		// args — a handle, two stack locals, length 0x38 and class 0x22
		// (FileNetworkOpenInformation, whose fixed size is exactly 0x38) — right after the
		// XAPI CreateFile wrapper; table-206 + the Nt block's +5 drift = 211. Times are 0
		// (the XISO carries none we have decoded), sizes are the disc entry's; any other
		// class halts and names itself.
		return func(m *Machine) int {
			h, iosb, buf, ln, class := m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4)
			fo := m.files[h]
			if fo == nil {
				m.finishOpen(iosb, h, 0, 0xC0000008) // STATUS_INVALID_HANDLE
				return 5
			}
			if class == 14 && ln >= 8 {
				// FilePositionInformation: the current byte offset as a LARGE_INTEGER
				// (the XAPI SetFilePointer/GetFileSize path reads it back, site 0x44321).
				m.write32(buf+0, fo.off)
				m.write32(buf+4, 0)
				m.finishOpen(iosb, h, 8, 0)
				return 5
			}
			if class != 0x22 || ln < 0x38 {
				m.CPU.Halt("NtQueryInformationFile: unmodelled class %d (len %d) from %08X",
					class, ln, m.retAddr())
				return 5
			}
			for i := uint32(0); i < 0x38; i += 4 {
				m.write32(buf+i, 0)
			}
			m.write32(buf+0x20, fo.size()) // AllocationSize (low; high already 0)
			m.write32(buf+0x28, fo.size()) // EndOfFile
			attrs := uint32(0x01 | 0x80)   // READONLY|NORMAL (a DVD file)
			if fo.entry.IsDir {
				attrs = 0x11 // READONLY|DIRECTORY
			}
			m.write32(buf+0x30, attrs)
			m.finishOpen(iosb, h, 0x38, 0)
			return 5
		}

	// --- Memory (Mm) — verified: 165 is a 1-arg allocation (the Mm block drifts +5) --
	case 165: // MmAllocateContiguousMemory(NumberOfBytes) -> physical base
		return func(m *Machine) int { m.setRet(m.allocPool(m.arg(0))); return 1 }

	case 166: // MmAllocateContiguousMemoryEx(NumberOfBytes, LowestAcceptableAddress,
		// HighestAcceptableAddress, Alignment, ProtectionType) -> physical base.
		// Verified from four live call sites (a 4-arg and a 5-arg .text wrapper, plus the
		// DSOUND and XONLINE callers): 5 stdcall args in the shape
		// (size, 0, 0xFFFFFFFF, align, PAGE_READWRITE|PAGE_WRITECOMBINE). The leading
		// PUSH ESI at the wrappers is a register save (balanced by POP ESI after the
		// call), not a sixth argument. We back it from the same down-growing contiguous
		// pool as ordinal 165, honouring the requested alignment; the [Lowest,Highest]
		// physical bounds are advisory (our pool already lives in low RAM under 0xFFFFFFFF)
		// so we do not constrain the base to them. Returns 0 on exhaustion, which the
		// callers treat as an allocation failure.
		return func(m *Machine) int {
			if m.arg(0) >= 1<<20 {
				m.logf("MmAllocateContiguousMemoryEx: %X bytes (lo=%X hi=%X align=%X prot=%X) from %08X (caller %08X)",
					m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4), m.retAddr(), m.read32(m.CPU.Regs[x86.BP]+4))
			}
			m.setRet(m.allocPoolAligned(m.arg(0), m.arg(3)))
			return 5
		}

	case 168: // MmClaimGpuInstanceMemory(NumberOfBytes, PSIZE_T PaddingBytes) -> end address.
		// Verified from its live call at the D3D device-init site: 2 args (0x5000, &pad),
		// the caller reads back (result - 0x5000) as the base of the GPU instance-memory
		// block and derives an NV2A DMA-object register from *pad. The kernel retains a
		// contiguous block at the top of physical RAM for the GPU (RAMIN: hash table, DMA
		// contexts) and returns the address just past its end. We back it from the same
		// down-growing contiguous pool (which lives near the top of RAM) and return
		// base+size so (result - size) is the claimed base; *pad = 0 (no alignment
		// remainder). Phase C (the NV2A) may refine the placement/padding to match the
		// instance-memory contract the GPU reads.
		return func(m *Machine) int {
			size := m.arg(0)
			base := m.allocPool(size)
			if p := m.arg(1); p != 0 {
				m.write32(p, 0) // NumberOfPaddingBytes
			}
			if base == 0 {
				m.setRet(0)
			} else {
				m.setRet(base + align32(size, 16))
			}
			return 2
		}

	case 65: // IoCreateDevice(DriverObject, DeviceExtensionSize, DeviceName(ANSI_STRING*),
		// DeviceType, Exclusive, DeviceObject**) -> NTSTATUS. Verified from its call site
		// (0x23F705): six args (obj 0x23F458 is a DRIVER_OBJECT with dispatch pointers at
		// +0x10/+0x14, ext size 0x170, the "\Device\MU_n" name, type 0x3A, FALSE, &local);
		// the caller reads the new device's +0x18 as the DeviceExtension pointer and zeroes
		// exactly DeviceExtensionSize bytes there — table-63 with the Io block's +2 drift
		// (the same +2 as Hal). XAPI creates one device object per memory-unit port; the
		// objects are bookkeeping (no MU media ever mounts here). Layout: a 0x40-byte
		// header with Type/Size and the extension pointer, the extension right behind it.
		return func(m *Machine) int {
			extSize := m.arg(1)
			out := m.arg(5)
			const hdr = 0x40
			dev := m.allocPool(hdr + extSize)
			if dev == 0 {
				m.setRet(0xC000009A) // STATUS_INSUFFICIENT_RESOURCES
				return 6
			}
			for i := uint32(0); i < hdr+extSize; i += 4 {
				m.write32(dev+i, 0)
			}
			m.write16(dev+0, 3)          // Type = IO_TYPE_DEVICE
			m.write16(dev+2, hdr)        // Size
			m.write32(dev+0x18, dev+hdr) // DeviceExtension
			if out != 0 {
				m.write32(out, dev)
			}
			name := ""
			if ns := m.arg(2); ns != 0 {
				name = m.cstr(m.read32(ns + 4))
			}
			m.logf("IoCreateDevice: %q type %02X ext %d -> %08X", name, m.arg(3), extSize, dev)
			m.setRet(0)
			return 6
		}

	case 173: // MmGetPhysicalAddress(BaseAddress) -> physical address. Verified from its
		// live call site (0x1DE100): one argument — the pointer the preceding contiguous
		// allocation returned — with the result stored alongside that pointer
		// (MOV [EDI],va; MOV [EDI+4],result); table-168 + the Mm block's +5 drift = 173.
		// The DSOUND library programs the APU's DMA with physical addresses. Our address
		// space folds the RAM windows onto one backing, so the physical address is the
		// translate() fold; a non-RAM argument would be a caller bug and halts.
		return func(m *Machine) int {
			va := m.arg(0)
			phys, mmio, ok := m.translate(va)
			if !ok || mmio {
				m.CPU.Halt("MmGetPhysicalAddress of non-RAM address %08X (from %08X)", va, m.retAddr())
				return 1
			}
			m.setRet(phys)
			return 1
		}

	case 175: // MmLockUnlockBufferPages(BaseAddress, NumberOfBytes, UnlockPages) — verified
		// from its call site (0x2401F1): three args (a fresh page-rounded allocation, its
		// size, FALSE), immediately followed by MmGetPhysicalAddress on the same base with
		// the virt-phys delta stored to a global — the canonical lock-before-DMA sequence;
		// table-170 + the Mm block's +5 drift = 175. Our flat RAM has no paging to lock:
		// success no-op, result unread by the caller.
		return func(m *Machine) int { m.setRet(0); return 3 }

	case 180: // MmQueryAllocationSize(BaseAddress) -> SIZE_T. Verified from its live call
		// site (0x1D6ADF): the single argument is the pointer the immediately preceding
		// MmAllocateContiguousMemoryEx (slot 0x248334, ordinal 166) returned, null-checked,
		// and the result is accumulated into a global allocated-bytes counter
		// (ADD [[0x1E0E3C]], EAX) — the Mm twin of the ExAllocatePoolWithTag /
		// ExQueryPoolBlockSize pair; table-175 + the Mm block's +5 drift = 180. Return the
		// size recorded at allocation (allocPoolAligned), 0 for an untracked block.
		return func(m *Machine) int {
			m.setRet(m.poolSizes[m.arg(0)])
			return 1
		}

	case 181: // MmQueryStatistics(PMM_STATISTICS) -> NTSTATUS. Between the two verified Mm
		// neighbours 180 (MmQueryAllocationSize) and 182 (MmSetAddressProtect), so pinned by
		// position; the call site (0x214C92) is the token/nonce builder, which passes a
		// stack struct and then reads AvailablePages (+8) as entropy without pre-setting
		// Length. Fill the MM_STATISTICS fields from the machine's real 64 MB accounting and
		// return success — a faithful page census, not an invented number, and the value is
		// only ever mixed into a self-produced nonce here.
		return func(m *Machine) int {
			p := m.arg(0)
			if p != 0 {
				const pageShift = 12
				total := uint32(ramSize >> pageShift)
				free := uint32(0)
				if m.poolNext > m.heapNext {
					free = (m.poolNext - m.heapNext) >> pageShift
				}
				poolPages := (uint32(kernelBandBase) - m.poolNext) >> pageShift
				m.write32(p+0x00, 0x24) // Length = sizeof(MM_STATISTICS)
				m.write32(p+0x04, total)
				m.write32(p+0x08, free)       // AvailablePages (read at +8 by the caller)
				m.write32(p+0x0C, total-free) // VirtualMemoryBytesCommitted (pages)
				m.write32(p+0x10, 0)          // VirtualMemoryBytesReserved
				m.write32(p+0x14, 0)          // CachePagesCommitted
				m.write32(p+0x18, poolPages)  // PoolPagesCommitted
				m.write32(p+0x1C, 0)          // StackPagesCommitted
				m.write32(p+0x20, 0)          // ImagePagesCommitted
			}
			m.setRet(0) // STATUS_SUCCESS
			return 1
		}

	case 182: // MmSetAddressProtect(BaseAddress, NumberOfBytes, NewProtect) — verified from
		// its one live call site: a 3-arg tail-call wrapper (JMP [slot]) that guards on
		// NumberOfBytes != 0, invoked here as (0x0128D000, 0x00280000, 0x404) right after
		// a contiguous allocation. Our RAM is flat with no page-protection enforcement and
		// the write-combine cache attribute does not change functional behaviour, so this
		// is a success no-op; the void return leaves EAX unused by the caller.
		return func(m *Machine) int { m.setRet(0); return 3 }

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
				m.logf("NtAllocateVirtualMemory: %X bytes type=%X prot=%X -> %08X (from %08X)",
					size, m.arg(3), m.arg(4), addr, m.retAddr())
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

	case 189: // NtCreateEvent(EventHandle*, ObjectAttributes, EventType, InitialState)
		// Canonical xboxkrnl ordinal, between the verified 187 NtClose and 190
		// NtCreateFile. Verified from its live call site (0x44D25, the XAPI CreateEvent
		// wrapper): four pushes — an out-handle local, the optional name attributes, an
		// EventType computed by SETZ from the wrapper's bManualReset argument
		// (Notification=0 when manual-reset, Synchronization=1 otherwise — exactly the
		// NT inversion), and InitialState forwarded. Synchronization events auto-clear
		// on a satisfied wait ("event-auto"); notification events stay signalled.
		return func(m *Machine) int {
			handleOut := m.arg(0)
			kind, typ := "event", byte(0)
			if m.arg(2) == 1 {
				kind, typ = "event-auto", 1
			}
			h := m.newObject(kind, typ, m.arg(3) != 0, 0, 0)
			if handleOut != 0 {
				m.write32(handleOut, h)
			}
			m.setRet(0) // STATUS_SUCCESS
			return 4
		}

	case 193: // NtCreateSemaphore(Handle*, ObjectAttributes, InitialCount, MaximumCount)
		// Canonical xboxkrnl ordinal (the verified Nt anchors 184/187/199/202 all match the
		// real table, so 193 is not drifted). The call site pushes 4 args with arg0 an
		// out-handle and checks the status for STATUS_OBJECT_NAME_EXISTS — a named object.
		// Mint a semaphore object; the HLE's wait logic (satisfyWait) reads its count.
		return func(m *Machine) int {
			handleOut := m.arg(0)
			initial := int32(m.arg(2))
			limit := int32(m.arg(3))
			h := m.newObject("semaphore", 5, initial > 0, initial, limit)
			if handleOut != 0 {
				m.write32(handleOut, h)
			}
			m.setRet(0) // STATUS_SUCCESS
			return 4
		}

	case 222: // NtReleaseSemaphore(Handle, ReleaseCount, PreviousCount*) -> NTSTATUS
		// Identified from the call site: 3 args forwarded through a wrapper that returns a
		// bool, in the same sync library as NtCreateSemaphore (193) and the timed wait (234)
		// — not the 5-arg NtSetIoCompletion the reconstructed table names. Raise the
		// semaphore's count, wake any thread blocked on it, and report the previous count.
		return func(m *Machine) int {
			handle, release, prevOut := m.arg(0), m.arg(1), m.arg(2)
			m.logf("NtReleaseSemaphore: handle %08X +%d from %08X (tick %d)", handle, release, m.retAddr(), m.tick)
			if o := m.objAt(handle); o != nil {
				if prevOut != 0 {
					m.write32(prevOut, uint32(o.count))
				}
				o.count += int32(release)
				o.signaled = o.count > 0
				m.writeSignal(o.addr, o.signaled)
				m.wakeWaiters(handle)
			} else if prevOut != 0 {
				m.write32(prevOut, 0)
			}
			m.setRet(0) // STATUS_SUCCESS
			return 3
		}

	case 234: // NtWaitForSingleObjectEx(Handle, WaitMode, Alertable, Timeout) -> NTSTATUS
		// Identified from the call site (NOT the reconstructed table, which misnames it
		// ObCreateObject): 4 args as (Handle, 1, Alertable, &Timeout), where the timeout
		// argument is built by a helper that multiplies a millisecond count by 10000 and
		// negates it — a relative LARGE_INTEGER. The wait is real (doWaitTimed): signalled
		// objects satisfy immediately, otherwise the thread parks until a signal or the
		// timeout — the earlier always-satisfied answer was a fiction that spun the
		// title's worker loops hot and starved its own producers.
		return func(m *Machine) int {
			m.doWaitTimed(m.arg(0), m.arg(3))
			return 4
		}

	case 199: // NtFreeVirtualMemory(BaseAddress**, RegionSize*, FreeType) -> NTSTATUS
		// Verified from its call site: 3 stdcall args wrapped in a lock/unlock pair, with
		// FreeType = 0x4000 (MEM_DECOMMIT) — not a PAGE_* value, so this is free, not the
		// 4-arg NtProtectVirtualMemory the reconstructed table names here. The Nt block's
		// +5 drift (194 -> 199) agrees with the verified NtAllocateVirtualMemory at 184. Our
		// allocators bump and never reclaim, so a decommit/release is a no-op success; leave
		// the caller's base/size in place.
		return func(m *Machine) int {
			m.setRet(0) // STATUS_SUCCESS
			return 3
		}

	// --- Handles (Nt) — verified: 187 is a 1-arg NtClose ------------------
	case 187: // NtClose(Handle) -> NTSTATUS. Verified from five live call sites, each
		// passing a single handle immediately after an open/create that returned it
		// (e.g. 0x427BF closes the handle from ordinal 207; 0x42846 closes the thread
		// handle from PsCreateSystemThreadEx). One argument — the earlier reading of 187
		// as a 3-arg NtCreateMutant was wrong and its over-pop derailed the boot thread.
		// The Nt block drifts +5 (the reconstructed table's NtClose at 182 lands at 187),
		// matching NtOpenFile at 202. We drop any open file backing the handle and report
		// success; other handle kinds (objects, threads) are reference-released as a
		// no-op since the HLE does not refcount them.
		return func(m *Machine) int {
			h := m.arg(0)
			delete(m.files, h)
			m.setRet(0) // STATUS_SUCCESS
			return 1
		}

	// --- PCI config space (Hal) — verified -------------------------------
	case 46: // HalReadWritePCISpace(BusNumber, SlotNumber, RegisterNumber, Buffer, Length,
		// WritePCISpace). Verified from the D3D device-init read-modify-write: it reads 4
		// bytes of config register 0x4C into a local, ORs a bit, and writes them back
		// (arg5 selects read vs write). Six args. The reconstructed Hal block drifts +2
		// (HalRegisterShutdownNotification at table index 45 is verified ordinal 47), so
		// table index 44 (HalReadWritePCISpace) lands at ordinal 46. We back config space
		// in a byte map keyed by (bus<<24|slot<<16|reg) so the read-after-write is coherent.
		return func(m *Machine) int {
			bus, slot, reg := m.arg(0), m.arg(1), m.arg(2)
			buf, length, write := m.arg(3), m.arg(4), m.arg(5)
			base := bus<<24 | slot<<16
			for i := uint32(0); i < length; i++ {
				key := base | ((reg + i) & 0xFFFF)
				if write != 0 {
					m.pciSpace[key] = m.Read(buf + i)
				} else {
					m.Write(buf+i, m.pciSpace[key])
				}
			}
			m.setRet(0)
			return 6
		}

	// --- Interrupts (Hal/Ke) — verified ----------------------------------
	case 44: // HalGetInterruptVector(BusInterruptLevel, PKIRQL Irql) -> Vector. Verified
		// two ways: (1) the live call at the D3D device-init site feeds its return value
		// (Vector) and its out-param (*Irql) straight into a 7-arg KeInitializeInterrupt
		// (ordinal 109) and then a 1-arg KeConnectInterrupt (ordinal 98) — the canonical
		// interrupt-hookup sequence; (2) the reconstructed Hal block drifts a uniform +2
		// (table's HalRegisterShutdownNotification at index 45 is verified ordinal 47), so
		// table index 42 (HalGetInterruptVector) lands at ordinal 44. Three of its five
		// call sites push exactly 2 args; the semantic arg count is 2.
		//
		// We do not dispatch hardware interrupts in this synchronous boot model (the
		// KeInitializeInterrupt/KeConnectInterrupt that consume these values only record
		// them), so the Vector/Irql are inert tokens: return the level as the vector and
		// write the level as the IRQL.
		return func(m *Machine) int {
			level := m.arg(0)
			if p := m.arg(1); p != 0 {
				m.Write(p, byte(level)) // *Irql (a KIRQL is one byte)
			}
			m.setRet(level)
			return 2
		}

	case 109: // KeInitializeInterrupt(Interrupt, ServiceRoutine, ServiceContext, Vector,
		// Irql, InterruptMode, ShareVector) — verified: the 7-arg call that consumes
		// HalGetInterruptVector's (Vector, Irql) at the D3D device-init site, immediately
		// before KeConnectInterrupt (ordinal 98). Records the routine, context and vector
		// into the KINTERRUPT block (leading fields, canonical across NT/Xbox); the
		// delivery machinery (interrupt.go) reads them back at fire time.
		return func(m *Machine) int {
			ki := m.arg(0)
			if ki != 0 {
				m.write32(ki+0x00, m.arg(1)) // ServiceRoutine
				m.write32(ki+0x04, m.arg(2)) // ServiceContext
				m.write32(ki+0x08, m.arg(3)) // BusInterruptLevel / vector
			}
			m.setRet(0)
			return 7
		}
	case 98: // KeConnectInterrupt(Interrupt) -> BOOLEAN. Verified: the 1-arg call right
		// after KeInitializeInterrupt on the same KINTERRUPT, whose AL result the caller
		// tests to decide success. Registers the interrupt with the delivery machinery
		// (interrupt.go) keyed by its vector.
		return func(m *Machine) int {
			ki := m.arg(0)
			if ki != 0 {
				vec := m.read32(ki + 0x08)
				m.interrupts[vec] = ki
				m.logf("KeConnectInterrupt: vector %d -> KINTERRUPT %08X (routine %08X ctx %08X)",
					vec, ki, m.read32(ki), m.read32(ki+4))
			}
			m.setRet(1)
			return 1
		}

	case 99: // KeGetCurrentThread() -> PKTHREAD. Canonical neighbour of the verified
		// KeConnectInterrupt (98); no arguments, the running thread's KTHREAD (which
		// doubles as its handle here). First called by XAPI's wait path (0x44E65).
		return func(m *Machine) int { m.setRet(m.currentKThread()); return 0 }

	case 119: // KeInsertQueueDpc(Dpc, SystemArgument1, SystemArgument2) -> BOOLEAN.
		// Verified from the D3D VBlank ISR (0x1B2E53, the first code the delivered
		// interrupt runs): three args, the KDPC at device+0x1BB64C whose routine
		// KeInitializeDpc recorded. Canonical ordinal (KeInitializeDpc at the verified
		// 107 anchors the block). The DPC runs after the ISR frame returns
		// (interrupt.go isrReturn), the hardware's ISR-then-DPC order; queuing an
		// already-queued DPC reports FALSE as the kernel does.
		return func(m *Machine) int {
			dpc := m.arg(0)
			for _, d := range m.dpcQueue {
				if d.Dpc == dpc {
					m.setRet(0) // already queued
					return 3
				}
			}
			m.dpcQueue = append(m.dpcQueue, dpcEntry{Dpc: dpc, Arg1: m.arg(1), Arg2: m.arg(2)})
			m.setRet(1)
			return 3
		}

	case 145: // KeSetEvent(Event*, Increment, Wait) -> LONG previous state. Verified from
		// the D3D VBlank DPC (0x1B301E): three args, the first the raw KEVENT at
		// 0x1BB75C — the very object the swap path's KeWaitForSingleObject (159) blocks
		// on. Canonical ordinal (KeSetBasePriorityThread at the verified 143 anchors
		// it). Signal, wake waiters, report the previous state.
		return func(m *Machine) int {
			o := m.guestObjAt(m.arg(0))
			prev := int32(0)
			if o != nil {
				if o.signaled {
					prev = 1
				}
				o.signaled = true
				m.writeSignal(o.addr, true)
				m.wakeWaiters(o.addr)
			}
			m.setRet(uint32(prev))
			return 3
		}

	case 159: // KeWaitForSingleObject(Object*, WaitReason, WaitMode, Alertable, Timeout*)
		// -> NTSTATUS. Verified from the D3D swap path (0x1A9C9D): five args — a raw
		// dispatcher object embedded in the device (device+0x1DBC, whose SignalState the
		// caller zeroes immediately before), WaitReason 6, WaitMode 1, no alert, NULL
		// timeout. Canonical ordinal (KeStallExecutionProcessor at the verified 151 and
		// KeTickCount at 156 anchor the block). Unlike the Nt waits this takes the OBJECT
		// ADDRESS, not a handle: guestObjAt adopts the raw DISPATCHER_HEADER so the same
		// wait/signal machinery applies. This is the swap's wait for the display's
		// vertical blank — the VBlank interrupt's ISR is what signals it.
		return func(m *Machine) int {
			m.guestObjAt(m.arg(0))
			m.doWaitTimed(m.arg(0), m.arg(4))
			return 5
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
		// (KTIMER, negative relative due time, KDPC). Arms the timer (timer.go); the DPC
		// runs when it expires. Returns whether the timer was already in the queue.
		//
		// This used to record the arguments and fire nothing, on the honest grounds that
		// nothing waited on it. XAPI's hub driver is what came to wait: it debounces a
		// port connect by 100 ms (0x241C01 pushes DueTime = -1,000,000) and resets the
		// port from the DPC, so a stub here meant the pad was seen, believed, and then
		// left for ever. See timer.go.
		return func(m *Machine) int {
			was := m.armTimer(m.arg(0), m.arg(3), m.arg(1), m.arg(2), 0)
			m.setRet(b2u(was))
			return 4
		}

	case 150: // KeSetTimerEx(Timer, DueTime(2 dwords), Period, Dpc) — verified from the NVNET
		// link-poll site (0x20D187): (this+0x1A8 KTIMER, 0xFFFFFFFFFFE17B80 = -200 ms
		// relative, Period=200 ms, this+0x18C KDPC) — the KTIMER/KDPC pair this object's
		// constructor built with KeInitializeTimerEx/KeInitializeDpc. Canonical neighbour
		// of the verified KeSetTimer (149), and armed the same way, with its period.
		return func(m *Machine) int {
			was := m.armTimer(m.arg(0), m.arg(4), m.arg(1), m.arg(2), m.arg(3))
			m.setRet(b2u(was))
			return 5
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

	case 97: // KeCancelTimer(Timer) -> BOOLEAN. Verified from the object destructor at
		// 0x20D1A9: one pointer arg = the KTIMER the constructor set up at this+0x1A8 with
		// KeInitializeTimerEx (ord 113), cancelled here as the object tears down.
		//
		// It used to answer FALSE always — "the timer was not pending" — which was true
		// only because KeSetTimer never queued one. That is the shape of stub worth being
		// wary of: correct, honest, and quietly wrong the moment its neighbour grows up.
		// It now answers for the real queue (timer.go).
		return func(m *Machine) int {
			tm := m.arg(0)
			was := m.cancelTimer(tm)
			if tm != 0 {
				m.write32(tm+dhSignalState, 0)
			}
			m.setRet(b2u(was))
			return 1
		}

	case 137: // KeRemoveQueueDpc(Dpc) -> BOOLEAN. Verified from the same destructor (0x20D1B6):
		// one pointer arg = the KDPC at this+0x18C (KeInitializeDpc, ord 107), dequeued as
		// the object tears down. Its old FALSE was the KeCancelTimer story again — "no DPC
		// is ever queued", true of a machine with no timer queue. Now it dequeues.
		return func(m *Machine) int {
			dpc := m.arg(0)
			found := false
			for i, d := range m.dpcQueue {
				if d.Dpc == dpc {
					m.dpcQueue = append(m.dpcQueue[:i], m.dpcQueue[i+1:]...)
					found = true
					break
				}
			}
			m.setRet(b2u(found))
			return 1
		}

	case 17: // ExFreePool(P) -> void. Verified from the destructor at 0x20D1DC: one pointer
		// arg = a block held at this+0x50, freed then nulled — the ExFreePool shape. Our
		// pool is a down-bumping arena with no reclamation, so this is a success no-op.
		return func(m *Machine) int { m.setRet(0); return 1 }

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
	case 301: // RtlNtStatusToDosError(NTSTATUS) -> Win32 error. The mapping is load-
		// bearing: XAPI funnels every NTSTATUS through here into the per-thread
		// last-error, and callers branch on EXACT Win32 codes. The predecessor's
		// "rough map" (st & 0xFFFF) sent STATUS_PENDING (0x103) to 259
		// (ERROR_NO_MORE_ITEMS) instead of 997 (ERROR_IO_PENDING) — the XMV movie
		// loader's GetLastError()==ERROR_IO_PENDING check could never pass, so the
		// title movie was aborted on every boot. Canonical documented mappings for
		// the statuses this HLE produces; anything unmapped returns
		// ERROR_MR_MID_NOT_FOUND (317) like the real function, and logs once so a
		// new status names itself instead of hiding behind a plausible number.
		return func(m *Machine) int {
			st := m.arg(0)
			var w uint32
			switch st {
			case 0:
				w = 0
			case 0x00000102: // STATUS_TIMEOUT -> WAIT_TIMEOUT
				w = 258
			case 0x00000103: // STATUS_PENDING -> ERROR_IO_PENDING
				w = 997
			case 0xC0000008: // STATUS_INVALID_HANDLE
				w = 6
			case 0xC0000011: // STATUS_END_OF_FILE -> ERROR_HANDLE_EOF
				w = 38
			case 0xC0000034: // STATUS_OBJECT_NAME_NOT_FOUND -> ERROR_FILE_NOT_FOUND
				w = 2
			case 0xC0000035: // STATUS_OBJECT_NAME_COLLISION -> ERROR_ALREADY_EXISTS
				w = 183
			case 0xC000003A: // STATUS_OBJECT_PATH_NOT_FOUND -> ERROR_PATH_NOT_FOUND
				w = 3
			default:
				w = 317 // ERROR_MR_MID_NOT_FOUND: the real fallback for unmapped statuses
				if !m.dosErrWarned[st] {
					if m.dosErrWarned == nil {
						m.dosErrWarned = map[uint32]bool{}
					}
					m.dosErrWarned[st] = true
					m.logf("RtlNtStatusToDosError: unmapped NTSTATUS %08X -> 317 (from %08X)", st, m.retAddr())
				}
			}
			m.setRet(w)
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

// doWaitTimed is the honest timed single-object wait: satisfy it now (STATUS_WAIT_0),
// or park the thread until the object is signalled (wakeWaiters -> WAIT_0) or the
// relative timeout expires (wakeDueSleepers -> STATUS_TIMEOUT). The predecessor of this
// function reported every timed wait as already satisfied; that fiction let a consumer
// thread spin hot through its wait-then-check-queue loop while the producer starved —
// the boot made no progress past resource loading until the wait actually waited.
func (m *Machine) doWaitTimed(handle, timeoutPtr uint32) {
	o := m.objAt(handle)
	if m.satisfyWait(o) {
		m.setRet(0) // STATUS_WAIT_0
		return
	}
	if m.current == nil {
		m.setRet(0) // no scheduler context (boot thread pre-spawn): do not deadlock
		return
	}
	wake := uint64(0) // 0 = wait forever (NULL timeout)
	if timeoutPtr != 0 {
		v := int64(uint64(m.read32(timeoutPtr+4))<<32 | uint64(m.read32(timeoutPtr)))
		switch {
		case v < 0: // relative, 100 ns units (the ms*10000-negate helper's shape)
			wake = m.tick + uint64(-v)/10000*instrsPerMs + 1
		case v == 0: // poll: no signal available right now
			m.setRet(0x102) // STATUS_TIMEOUT
			return
		default: // absolute system time, against the same clock sched.go publishes
			now := m.tick / instrsPerMs * 10000
			if uint64(v) <= now {
				m.setRet(0x102)
				return
			}
			wake = m.tick + (uint64(v)-now)/10000*instrsPerMs + 1
		}
	}
	t := m.current
	t.waitObjs = []uint32{handle}
	t.wakeTick = wake
	kind := "guest"
	if o != nil {
		kind = o.kind
	}
	m.logf("wait: tid=%d parks on %08X (%s) wake=%d from %08X", t.id, handle, kind, wake, m.retAddr())
	m.setRet(0x102) // the timeout result; a signal wake overwrites saved EAX with WAIT_0
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
				t.wakeTick = 0
				// A signal wake returns STATUS_WAIT_0; the parked context carries the
				// timeout result (doWaitTimed), so overwrite its saved EAX.
				t.ctx.Regs[x86.AX] = 0
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
