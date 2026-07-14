package n3ds

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// svc.go high-level-emulates the Horizon supervisor calls. An svc #n on the
// ARM11 traps here through the CPU's SWI hook. The ABI: arguments in r0-r3 (a
// few calls take a fifth in r4), the result code in r0, and any output values in
// r1+ — exactly what the userland svc wrappers expect on return.
//
// The surface is deliberately partial and honest. Three tiers:
//
//   - Memory and information calls (ControlMemory, QueryMemory, GetSystemTick,
//     GetSystemInfo, GetThreadPriority, …) are modelled for real: a runtime's
//     early setup depends on their actual results.
//   - Kernel-object calls (CreateThread/Event/Mutex/Semaphore, WaitSynchronization,
//     CloseHandle, ConnectToPort) are *stubbed*: they hand out handles and report
//     waitable objects as already signalled, so init proceeds without a scheduler.
//     This is a scaffold, not a faithful kernel — a program whose correctness
//     depends on real thread/sync semantics will behave differently, and that is
//     called out here rather than hidden.
//   - Everything else halts with its number and PC. Reaching an unimplemented svc
//     is the machine reporting exactly how far the bring-up currently goes.
//
// The IPC path (SendSyncRequest) is where the real work of talking to the srv:,
// GSP and APT services would begin; it currently logs the command header and
// returns an error, which is the next milestone's starting point.
const (
	svcControlMemory      = 0x01
	svcQueryMemory        = 0x02
	svcExitProcess        = 0x03
	svcGetProcessAffinity = 0x04
	svcCreateThread       = 0x08
	svcExitThread         = 0x09
	svcSleepThread        = 0x0A
	svcGetThreadPriority  = 0x0B
	svcSetThreadPriority  = 0x0C
	svcCreateMutex        = 0x13
	svcReleaseMutex       = 0x14
	svcCreateSemaphore    = 0x15
	svcReleaseSemaphore   = 0x16
	svcCreateEvent        = 0x17
	svcSignalEvent        = 0x18
	svcClearEvent         = 0x19
	svcCreateTimer        = 0x1A
	svcCreateMemoryBlock  = 0x1E
	svcMapMemoryBlock     = 0x1F
	svcUnmapMemoryBlock   = 0x20
	svcCreateAddressArb   = 0x21
	svcArbitrateAddress   = 0x22
	svcCloseHandle        = 0x23
	svcWaitSync1          = 0x24
	svcWaitSyncN          = 0x25
	svcDuplicateHandle    = 0x27
	svcGetSystemTick      = 0x28
	svcGetHandleInfo      = 0x29
	svcGetSystemInfo      = 0x2A
	svcGetProcessInfo     = 0x2B
	svcGetThreadInfo      = 0x2C
	svcConnectToPort      = 0x2D
	svcSendSyncRequest    = 0x32
	svcGetProcessId       = 0x35
	svcGetThreadId        = 0x37
	svcGetResourceLimit   = 0x38
	svcGetResLimitCurrent = 0x39
	svcGetResLimitLimit   = 0x3A
	svcBreak              = 0x3C
	svcOutputDebugString  = 0x3D
)

// Horizon result codes used by the HLE.
const (
	resultSuccess uint32 = 0x00000000
)

// handleSVC is the CPU's SWI hook. Returning true tells the core the call was
// serviced (so it does not vector to the ARM exception base).
func (m *Machine) handleSVC(c *arm.CPU, comment uint32) bool {
	defer m.profEnd(bucketSVC, m.profStart()) // profile.go
	num := comment & 0xFF
	ev := svcEvent{PC: c.PC(), Num: num, Name: svcName(num)}
	ev.Args = [4]uint32{c.R[0], c.R[1], c.R[2], c.R[3]}
	m.svcLog = append(m.svcLog, ev)
	if m.Verbose {
		fmt.Printf("[t%d] svc 0x%02X %-20s r0=%08X r1=%08X r2=%08X r3=%08X  pc=%08X\n",
			m.curThread.id, num, ev.Name, c.R[0], c.R[1], c.R[2], c.R[3], c.PC())
	}

	switch num {
	case svcControlMemory:
		m.svcControlMemory(c)
	case svcQueryMemory:
		m.svcQueryMemory(c)
	case svcExitProcess:
		c.Halt("svcExitProcess after %d instructions", c.Instrs)
	case svcGetSystemTick:
		c.R[0] = uint32(m.tick)
		c.R[1] = uint32(m.tick >> 32)
	case svcGetSystemInfo:
		c.R[0], c.R[1], c.R[2] = resultSuccess, 0, 0
	case svcGetProcessInfo, svcGetThreadInfo, svcGetHandleInfo:
		c.R[0], c.R[1], c.R[2] = resultSuccess, 0, 0
	case svcGetThreadPriority:
		c.R[0], c.R[1] = resultSuccess, uint32(m.curThread.priority)
	case svcSetThreadPriority:
		m.curThread.priority = int32(c.R[1])
		c.R[0] = resultSuccess
	case svcSleepThread:
		m.svcSleepThread(c)
	case svcExitThread:
		m.svcExitThread(c)
	case svcSignalEvent:
		m.svcSignalEvent(c)
	case svcClearEvent:
		m.svcClearEvent(c)
	case svcReleaseMutex:
		m.svcReleaseMutex(c)
	case svcReleaseSemaphore:
		m.svcReleaseSemaphore(c)
	case svcArbitrateAddress:
		m.svcArbitrateAddress(c)
	case svcGetProcessId:
		c.R[0], c.R[1] = resultSuccess, 1
	case svcGetThreadId:
		c.R[0], c.R[1] = resultSuccess, m.curThread.id
	case svcCreateThread:
		// r0=priority, r1=entry, r2=arg, r3=stacktop.
		h := m.createThread(int32(c.R[0]), c.R[1], c.R[2], c.R[3])
		c.R[0], c.R[1] = resultSuccess, h
	case svcCreateEvent:
		h := m.newHandle("event", false)
		m.handles[h].manualReset = c.R[1] != 0 // r1 = reset type (1 = sticky)
		c.R[0], c.R[1] = resultSuccess, h
	case svcCreateMutex:
		h := m.newHandle("mutex", false)
		if c.R[1] != 0 { // r1 = initially locked
			m.handles[h].mutexOwner, m.handles[h].mutexDepth = m.curThread.id, 1
		}
		c.R[0], c.R[1] = resultSuccess, h
	case svcCreateSemaphore:
		h := m.newHandle("semaphore", false)
		m.handles[h].semCount = int32(c.R[1]) // r1 = initial count
		c.R[0], c.R[1] = resultSuccess, h
	case svcCreateTimer:
		m.svcCreateHandle(c, "timer", false, 1)
	case svcCreateAddressArb:
		m.svcCreateHandle(c, "arbiter", false, 1)
	case svcCreateMemoryBlock:
		m.svcCreateHandle(c, "memblock", false, 0) // handle out in r0
	case svcMapMemoryBlock:
		m.svcMapMemoryBlock(c)
	case svcUnmapMemoryBlock:
		m.svcUnmapMemoryBlock(c)
	case svcDuplicateHandle:
		m.svcCreateHandle(c, "dup", false, 1)
	case svcConnectToPort:
		m.svcConnectToPort(c)
	case svcWaitSync1:
		m.svcWaitSync1(c)
	case svcWaitSyncN:
		m.svcWaitSyncN(c)
	case svcCloseHandle:
		delete(m.handles, c.R[0])
		delete(m.ports, c.R[0])
		c.R[0] = resultSuccess
	case svcSendSyncRequest:
		m.svcSendSyncRequest(c)
	case svcGetResourceLimit:
		m.svcCreateHandle(c, "resourcelimit", false, 1)
	case svcGetResLimitCurrent:
		m.svcGetResourceLimitValues(c, true)
	case svcGetResLimitLimit:
		m.svcGetResourceLimitValues(c, false)
	case svcOutputDebugString:
		m.svcOutputDebugString(c)
	case svcBreak:
		c.Halt("svcBreak (reason %d) at 0x%08X after %d instructions", c.R[0], c.PC(), c.Instrs)
	default:
		c.Halt("unimplemented svc 0x%02X (%s) at 0x%08X after %d instructions",
			num, svcName(num), c.PC(), c.Instrs)
	}
	return true
}

// svcControlMemory services heap and linear-heap allocation. Real behaviour: the
// runtime's allocator and the LINEAR heap depend on getting back a valid, mapped
// address, so this grows the appropriate region and returns it. Only the ALLOC
// operation is modelled; FREE and MAP shrink/alias, which the bring-up does not
// need yet and which halt.
//
// The kernel-entry ABI (confirmed by disassembling the wrapper at 0x00293054:
// PUSH {r0,r4}; LDR r0,[sp,#8]; LDR r4,[sp,#0xC]; svc #1) is r0=operation,
// r1=addr0, r2=addr1, r3=size, r4=perm — NOT the C-prototype order. Result in r0,
// the mapped address in r1.
func (m *Machine) svcControlMemory(c *arm.CPU) {
	memop := c.R[0]
	size := (c.R[3] + pageSize - 1) &^ (pageSize - 1)
	switch memop & 0xFF {
	case 3: // MEMOP_ALLOC — handled below
	case 1, 4, 6: // MEMOP_FREE / MAP / PROTECT
		// The HLE's heaps are bump allocators that never reclaim, so a free is a
		// no-op leak and a map/protect is accepted unchanged. r1 (the mapped
		// address the caller passed in addr0) is left untouched for a free-check.
		c.R[0] = resultSuccess
		return
	default:
		c.Halt("svcControlMemory op 0x%X unimplemented at 0x%08X", memop&0xFF, c.PC())
		return
	}
	// Guard against an implausible request (a mis-derived size the runtime would
	// itself reject on hardware) turning into a multi-gigabyte host allocation.
	if size > heapMax-heapBase {
		c.Halt("svcControlMemory: implausible size 0x%X at 0x%08X (heap-size derivation likely wrong)", c.R[3], c.PC())
		return
	}
	// LINEAR memory carries the MEMOP_LINEAR flag (0x10000); everything else is
	// ordinary process heap.
	linear := memop&0x10000 != 0
	var addr uint32
	if linear {
		addr = m.linearPtr
		m.linearPtr += size
		m.linearReg.data = append(m.linearReg.data, make([]byte, size)...)
		m.indexRegion(m.linearReg) // the region grew: the new pages are now backed
		if m.linearPtr > linearMax {
			c.Halt("linear heap exhausted at 0x%08X", c.PC())
			return
		}
	} else {
		addr = m.heapPtr
		m.heapPtr += size
		m.heapReg.data = append(m.heapReg.data, make([]byte, size)...)
		m.indexRegion(m.heapReg) // the region grew: the new pages are now backed
		if m.heapPtr > heapMax {
			c.Halt("process heap exhausted at 0x%08X", c.PC())
			return
		}
	}
	if m.MemTrace {
		fmt.Printf("[mem] ControlMemory op=%06X addr0=%08X addr1=%08X size=%08X %s -> %08X (linear now %08X)\n",
			memop, c.R[1], c.R[2], size, map[bool]string{true: "LINEAR", false: "heap"}[linear], addr, m.linearPtr)
	}
	c.R[0] = resultSuccess
	c.R[1] = addr
}

// svcQueryMemory returns a minimal MemoryInfo/PageInfo for the queried address:
// its containing region's base and size, a "private, read/write" state, and zero
// page flags. Enough for an allocator that sanity-checks its own mappings.
func (m *Machine) svcQueryMemory(c *arm.CPU) {
	addr := c.R[2]
	r := m.regionOf(addr)
	if r == nil {
		// FREE region: base 0, size to the next mapping is unknown; report a
		// large free span from the address.
		c.R[0], c.R[1], c.R[2], c.R[3], c.R[4], c.R[5] = resultSuccess, addr, 0x1000, 0, 0, 0
		return
	}
	if m.MemTrace {
		fmt.Printf("[mem] QueryMemory(%08X) -> region %q base=%08X size=%08X\n", addr, r.name, r.base, len(r.data))
	}
	c.R[0] = resultSuccess
	c.R[1] = r.base              // base address
	c.R[2] = uint32(len(r.data)) // size
	c.R[3] = 3                   // MEMSTATE_PRIVATE-ish
	c.R[4] = 3                   // permission RW
	c.R[5] = 0                   // page flags
}

// svcCreateHandle allocates a stub kernel object and returns its handle. handleReg
// selects which register receives the handle (0 for the calls that return it in
// r0, 1 for those that return it in r1); signalled marks a waitable object as
// permanently ready so a later WaitSynchronization succeeds immediately.
func (m *Machine) svcCreateHandle(c *arm.CPU, kind string, signalled bool, handleReg int) {
	h := m.nextHandle
	m.nextHandle++
	m.handles[h] = &kobject{kind: kind, signal: signalled}
	c.R[0] = resultSuccess
	c.R[uint32(handleReg)] = h // r0 for the calls that return the handle there
}

// svcConnectToPort records the named port and returns a handle. The name is a
// pointer in r1 to a NUL-terminated string ("srv:", "APT:U", …). Capturing it is
// itself useful: the sequence of ports a title opens is a map of the services it
// needs, which is the work list for the IPC milestone.
func (m *Machine) svcConnectToPort(c *arm.CPU) {
	namePtr := c.R[1]
	name := m.readCString(namePtr, 12)
	h := m.nextHandle
	m.nextHandle++
	m.handles[h] = &kobject{kind: "port", name: name}
	m.ports[h] = name
	c.R[0] = resultSuccess
	c.R[1] = h
	if m.Verbose {
		fmt.Printf("  ConnectToPort %q -> handle 0x%08X\n", name, h)
	}
}

// svcSendSyncRequest is the IPC entry point: the command header and parameters
// sit in the calling thread's TLS command buffer at TLS+0x80, addressed by the
// session handle in r0. It dispatches to the HLE service layer (ipc.go), which
// writes the reply back into the same buffer. The svc itself succeeds; a service
// this layer does not implement halts there with the service and command ID.
func (m *Machine) svcSendSyncRequest(c *arm.CPU) {
	m.handleIPC(c.R[0])
	c.R[0] = resultSuccess
}

// svcGetResourceLimitValues services GetResourceLimitCurrentValues (current==true)
// and GetResourceLimitLimitValues (current==false): for each requested
// ResourceLimitType name it writes an s64 value into the output array. The runtime
// sizes its heap from these — `limit(COMMIT) − current(COMMIT)` is the free memory
// it then allocates — and it checks that the difference is page-aligned before
// committing, branching to an error path otherwise. So the two must differ by a
// page-aligned, positive amount:
//
//   - COMMIT limit   = the application memory budget (APPMEMALLOC, 64 MiB).
//   - COMMIT current = the amount already committed (code + stack, rounded up).
//
// Their difference is a valid, page-aligned heap size. The other resource kinds
// (thread/event/mutex counts, priority, CPU time) are reported generously; these
// are HLE stubs, not console-measured figures.
//
// r0 = s64* values, r1 = resourceLimit handle, r2 = ResourceLimitType* names,
// r3 = name count.
func (m *Machine) svcGetResourceLimitValues(c *arm.CPU, current bool) {
	values := c.R[0]
	names := c.R[2]
	count := int32(c.R[3])
	for i := int32(0); i < count && i < 32; i++ {
		name := m.ReadWord(names + uint32(i)*4)
		var v int64
		switch name {
		case 1: // COMMIT — bytes of committed memory
			// The runtime computes its heap as APPMEMALLOC (config+0x40) minus the
			// COMMIT limit, so the COMMIT figure is the committed base (code +
			// stack), not the whole budget — otherwise the heap would be zero or
			// negative. Both current and limit report that base here.
			v = int64(committedBytes)
		case 0: // PRIORITY (lowest numerical priority allowed)
			v = 0x18
		case 9: // CPUTIME
			v = 0
		default: // THREAD/EVENT/MUTEX/SEMAPHORE/TIMER/SHAREDMEMORY/ARBITER counts
			if current {
				v = 0
			} else {
				v = 0x100
			}
		}
		off := values + uint32(i)*8
		m.WriteWord(off, uint32(v))
		m.WriteWord(off+4, uint32(v>>32))
	}
	c.R[0] = resultSuccess
}

// committedBytes is the memory reported as already committed at startup: enough
// to cover the loaded code image and the main stack, page-aligned, so that the
// heap size the runtime derives (limit − current) is a sane, aligned figure.
const committedBytes = 0x00400000 // 4 MiB

// svcOutputDebugString appends the runtime's debug text (r0=ptr, r1=len) to the
// captured buffer — a direct window into how far init got and any error it
// printed.
func (m *Machine) svcOutputDebugString(c *arm.CPU) {
	ptr, n := c.R[0], c.R[1]
	for i := uint32(0); i < n && i < 0x1000; i++ {
		m.debugOut = append(m.debugOut, m.Read(ptr+i))
	}
	m.debugOut = append(m.debugOut, '\n')
	c.R[0] = resultSuccess
}

// readCString reads a NUL-terminated string of at most max bytes.
func (m *Machine) readCString(addr, max uint32) string {
	var b []byte
	for i := uint32(0); i < max; i++ {
		ch := m.Read(addr + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

func svcName(n uint32) string {
	names := map[uint32]string{
		svcControlMemory: "ControlMemory", svcQueryMemory: "QueryMemory",
		svcExitProcess: "ExitProcess", svcCreateThread: "CreateThread",
		svcExitThread: "ExitThread", svcSleepThread: "SleepThread",
		svcGetThreadPriority: "GetThreadPriority", svcSetThreadPriority: "SetThreadPriority",
		svcCreateMutex: "CreateMutex", svcReleaseMutex: "ReleaseMutex",
		svcCreateSemaphore: "CreateSemaphore", svcReleaseSemaphore: "ReleaseSemaphore",
		svcCreateEvent: "CreateEvent", svcSignalEvent: "SignalEvent", svcClearEvent: "ClearEvent",
		svcCreateTimer: "CreateTimer", svcCreateMemoryBlock: "CreateMemoryBlock",
		svcMapMemoryBlock: "MapMemoryBlock", svcCreateAddressArb: "CreateAddressArbiter",
		svcArbitrateAddress: "ArbitrateAddress", svcCloseHandle: "CloseHandle",
		svcWaitSync1: "WaitSynchronization1", svcWaitSyncN: "WaitSynchronizationN",
		svcDuplicateHandle: "DuplicateHandle", svcGetSystemTick: "GetSystemTick",
		svcGetSystemInfo: "GetSystemInfo", svcGetProcessInfo: "GetProcessInfo",
		svcGetThreadInfo: "GetThreadInfo", svcConnectToPort: "ConnectToPort",
		svcSendSyncRequest: "SendSyncRequest", svcGetProcessId: "GetProcessId",
		svcGetThreadId: "GetThreadId", svcGetResourceLimit: "GetResourceLimit",
		svcGetResLimitCurrent: "GetResourceLimitCurrentValues",
		svcGetResLimitLimit:   "GetResourceLimitLimitValues",
		svcBreak:              "Break", svcOutputDebugString: "OutputDebugString",
	}
	if s, ok := names[n]; ok {
		return s
	}
	return fmt.Sprintf("svc_0x%02X", n)
}

// SVCLog returns the ordered list of supervisor calls made so far.
func (m *Machine) SVCLog() []svcEvent { return m.svcLog }

// Ports returns the service ports connected so far (handle → name).
func (m *Machine) Ports() map[uint32]string { return m.ports }
