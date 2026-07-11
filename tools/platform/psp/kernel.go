package psp

// kernel.go is the PSP kernel HLE. A game reaches the kernel by calling an imported
// sceXxx function whose call stub ends in a `syscall`; at load time each stub is
// patched to `jr $ra; syscall <synthetic code>` and the code is mapped to a Go
// handler. The CPU's Syscall hook dispatches here.
//
// Functions are named by hashing: a PSP import is identified by a 32-bit NID that is
// the first four bytes (little-endian) of SHA-1(function name). Hashing a curated
// list of names gives the NID->name map used to label the trace and to bind the
// handful of functions modelled for real.
//
// Three tiers, after tools/platform/n3ds's svc.go:
//   - modelled: the memory / thread-entry / display / time calls the C runtime and
//     module start touch, to the extent their results are read;
//   - stubbed: kernel objects (sema/event/mutex) hand out handles and report success
//     so single-threaded init proceeds without a scheduler;
//   - everything else logs its (library,NID) and returns 0, so one run enumerates the
//     whole syscall surface the boot path reaches — the work list.

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/allegrex"
)

// syscall binds a synthetic code to a named handler.
type syscall struct {
	name    string
	handler func(m *Machine)
}

// kobject is a kernel object handed out by the HLE (thread, memory block, sema, …).
type kobject struct {
	kind  string
	name  string
	entry uint32 // thread entry, for threads
	addr  uint32 // block address, for memory blocks

	// Thread fields (see sched.go).
	priority uint32
	stackTop uint32
	tstate   threadState
	ctx      allegrex.CPUState // saved register context while not running
}

// nidOf computes a function's NID: SHA-1(name)[0:4] little-endian.
func nidOf(name string) uint32 {
	h := sha1.Sum([]byte(name))
	return binary.LittleEndian.Uint32(h[:4])
}

// knownNIDs are the function names hashed to build the NID->name label map. Adding a
// name here only makes a trace legible; a name in modelled/stubbed also binds behaviour.
var knownNIDs = []string{
	// SysMemUserForUser
	"sceKernelAllocPartitionMemory", "sceKernelFreePartitionMemory", "sceKernelGetBlockHeadAddr",
	"sceKernelTotalFreeMemSize", "sceKernelMaxFreeMemSize", "sceKernelPrintf", "sceKernelDevkitVersion",
	// ThreadManForUser
	"sceKernelCreateThread", "sceKernelStartThread", "sceKernelExitThread", "sceKernelExitDeleteThread",
	"sceKernelDeleteThread", "sceKernelGetThreadId", "sceKernelDelayThread", "sceKernelDelayThreadCB",
	"sceKernelSleepThread", "sceKernelSleepThreadCB", "sceKernelWakeupThread", "sceKernelTerminateDeleteThread",
	"sceKernelCreateCallback", "sceKernelReferThreadStatus", "sceKernelGetThreadExitStatus",
	"sceKernelGetGPI", "sceKernelGetGPO",
	"sceKernelCreateSema", "sceKernelDeleteSema", "sceKernelWaitSema", "sceKernelWaitSemaCB", "sceKernelSignalSema",
	"sceKernelCreateEventFlag", "sceKernelWaitEventFlag", "sceKernelSetEventFlag", "sceKernelClearEventFlag",
	"sceKernelCreateMutex", "sceKernelLockMutex", "sceKernelUnlockMutex",
	"sceKernelCreateMsgPipe", "sceKernelGetSystemTimeWide", "sceKernelGetSystemTimeLow",
	"sceKernelGetThreadCurrentPriority", "sceKernelChangeThreadPriority",
	// Kernel_Library (interrupt control, used for critical sections)
	"sceKernelCpuSuspendIntr", "sceKernelCpuResumeIntr", "sceKernelCpuResumeIntrWithSync",
	// LoadExecForUser / ModuleMgrForUser
	"sceKernelExitGame", "sceKernelRegisterExitCallback", "sceKernelLoadModule", "sceKernelStartModule",
	"sceKernelSelfStopUnloadModule", "sceKernelGetModuleIdByAddress", "sceKernelGetModuleId",
	"sceKernelSetCompiledSdkVersion", "sceKernelSetCompilerVersion",
	// UtilsForUser
	"sceKernelDcacheWritebackAll", "sceKernelDcacheWritebackRange", "sceKernelDcacheWritebackInvalidateAll",
	"sceKernelIcacheInvalidateAll", "sceKernelLibcTime", "sceKernelLibcClock", "sceKernelLibcGettimeofday",
	"sceKernelUtilsMt19937Init", "sceKernelUtilsMt19937UInt",
	// sceDisplay
	"sceDisplaySetMode", "sceDisplaySetFrameBuf", "sceDisplayGetFrameBuf", "sceDisplayWaitVblankStart",
	"sceDisplayWaitVblankStartCB", "sceDisplayGetVcount", "sceDisplayIsVblank",
	// sceGe_user
	"sceGeEdramGetAddr", "sceGeEdramGetSize", "sceGeListEnQueue", "sceGeListSync", "sceGeDrawSync",
	"sceGeSetCallback",
	// sceCtrl
	"sceCtrlSetSamplingCycle", "sceCtrlSetSamplingMode", "sceCtrlReadBufferPositive", "sceCtrlPeekBufferPositive",
	// StdioForUser
	"sceKernelStdout", "sceKernelStdin", "sceKernelStderr",
	// sceRtc / scePower
	"sceRtcGetCurrentTick", "scePowerGetCpuClockFrequency", "scePowerRegisterCallback",
}

// nidName maps a (library, NID) to a legible name, or "library:0xNID" if unknown.
func nidName(lib string, nid uint32) string {
	if name, ok := nidLookup[nid]; ok {
		return name
	}
	return fmt.Sprintf("%s:0x%08X", lib, nid)
}

var nidLookup = func() map[uint32]string {
	m := make(map[uint32]string, len(knownNIDs))
	for _, n := range knownNIDs {
		m[nidOf(n)] = n
	}
	return m
}()

// installStubs patches every import stub to `jr $ra; syscall <code>` and binds the
// code to the handler for that function.
func (m *Machine) installStubs(mod *Module) {
	for _, imp := range mod.Imports {
		for i, nid := range imp.NIDs {
			stub := userBase + imp.StubAddr + uint32(i)*8
			name := nidName(imp.Library, nid)
			code := m.nextSyscall
			m.nextSyscall++
			m.syscalls[code] = &syscall{name: name, handler: handlerFor(name)}
			m.write32(stub+0, 0x03E00008)           // jr $ra
			m.write32(stub+4, (code<<6)|0x0000000C) // syscall <code>
		}
	}
}

// handleSyscall is the CPU Syscall hook: dispatch the synthetic code to its handler.
func (m *Machine) handleSyscall(c *allegrex.CPU, code uint32) bool {
	sc := m.syscalls[code]
	if sc == nil {
		m.note("unknown syscall code 0x%X at pc 0x%08X", code, m.CPU.CurPC())
		m.setRet(0)
		return true
	}
	m.SyscallCalls[sc.name]++
	if sc.handler != nil {
		sc.handler(m)
	} else {
		m.setRet(0) // stubbed / unmodelled: report success
	}
	return true
}

// --- handler helpers -------------------------------------------------------

func (m *Machine) arg(i uint32) uint32 { return m.CPU.Reg(4 + i) } // $a0..$a3
func (m *Machine) setRet(v uint32)     { m.CPU.SetReg(2, v) }      // $v0
func (m *Machine) newHandle(o *kobject) uint32 {
	h := m.nextHandle
	m.nextHandle++
	m.handles[h] = o
	return h
}

// cstr reads a NUL-terminated string from guest memory.
func (m *Machine) cstr(addr uint32) string {
	var b []byte
	for i := uint32(0); i < 256; i++ {
		ch := m.Read(addr + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

// handlerFor returns the modelled handler for a function name, or nil (stub) for the
// rest. Kept in one place so the modelled surface is auditable.
func handlerFor(name string) func(m *Machine) {
	switch name {
	case "sceKernelAllocPartitionMemory":
		return func(m *Machine) {
			// (partition, name, type, size, addr) — bump-allocate size bytes.
			size := m.arg(3)
			addr := m.heapPtr
			m.heapPtr = (m.heapPtr + size + 0xFF) &^ 0xFF
			h := m.newHandle(&kobject{kind: "block", name: m.cstr(m.arg(1)), addr: addr})
			m.setRet(h)
		}
	case "sceKernelGetBlockHeadAddr":
		return func(m *Machine) {
			if o := m.handles[m.arg(0)]; o != nil {
				m.setRet(o.addr)
				return
			}
			m.setRet(0)
		}
	case "sceKernelCreateThread":
		return func(m *Machine) {
			// (name, entry, priority, stacksize, attr, option)
			stackSize := m.arg(3)
			if stackSize < 0x1000 {
				stackSize = 0x4000
			}
			base := m.heapPtr
			m.heapPtr = (base + stackSize + 0xFF) &^ 0xFF
			o := &kobject{
				kind: "thread", name: m.cstr(m.arg(0)), entry: m.arg(1),
				priority: m.arg(2), stackTop: m.heapPtr - 16, tstate: thDormant,
			}
			m.threadEntry = o.entry
			m.setRet(m.newHandle(o))
		}
	case "sceKernelStartThread":
		return func(m *Machine) {
			// (thid, arglen, argp): make the thread runnable; the caller continues.
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "thread" {
				m.setRet(0x80020198) // ERROR_NOT_FOUND_THREAD
				return
			}
			m.startThread(o, m.arg(1), m.arg(2))
			m.setRet(0)
		}
	case "sceKernelExitGame", "sceKernelSelfStopUnloadModule":
		return func(m *Machine) {
			m.Halted = true
			m.HaltReason = "game requested exit (sceKernelExitGame)"
		}
	case "sceKernelExitThread", "sceKernelExitDeleteThread", "sceKernelTerminateDeleteThread":
		return func(m *Machine) { m.onThreadExit() }
	case "sceKernelSleepThread", "sceKernelSleepThreadCB":
		return func(m *Machine) { m.yieldCurrent(thWaiting) }
	case "sceKernelDelayThread", "sceKernelDelayThreadCB":
		return func(m *Machine) { m.setRet(0) } // treated as a no-op yield
	case "sceDisplaySetFrameBuf":
		return func(m *Machine) {
			// (topaddr, bufferwidth, pixelformat, sync)
			m.fbAddr = m.arg(0)
			m.fbWidth = m.arg(1)
			m.fbFormat = m.arg(2)
			m.setRet(0)
		}
	case "sceGeListEnQueue", "sceGeListEnQueueHead":
		return func(m *Machine) {
			// (list, stall, cbid, arg): capture and execute the display list.
			list := m.captureList(m.arg(0))
			m.GeLists = append(m.GeLists, list)
			if m.OnGeList != nil {
				m.OnGeList(list)
			}
			m.execGeList(list)
			m.setRet(uint32(len(m.GeLists))) // a nonzero queue id
		}
	case "sceGeEdramGetAddr":
		return func(m *Machine) { m.setRet(vramBase) }
	case "sceGeEdramGetSize":
		return func(m *Machine) { m.setRet(vramSize) }
	case "sceKernelGetThreadId", "sceKernelGetModuleIdByAddress", "sceKernelGetModuleId":
		return func(m *Machine) { m.setRet(1) }
	case "sceKernelDcacheWritebackAll", "sceKernelDcacheWritebackRange",
		"sceKernelDcacheWritebackInvalidateAll", "sceKernelIcacheInvalidateAll",
		"sceDisplayWaitVblankStart", "sceDisplaySetMode":
		return func(m *Machine) { m.setRet(0) }
	}
	return nil // stub: report success (0)
}
