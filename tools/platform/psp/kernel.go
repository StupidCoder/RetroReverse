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
	"strings"

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
	entry uint32 // thread entry, for threads; callback function, for callbacks
	addr  uint32 // block address, for memory blocks and pools
	size  uint32 // pool size, for vpl/fpl
	used  uint32 // pool bump pointer, for vpl/fpl

	// Event-flag fields.
	bits uint32 // current bit pattern

	// Semaphore fields.
	count int32

	// Thread fields (see sched.go).
	priority uint32
	stackTop uint32
	tstate   threadState
	ctx      allegrex.CPUState // saved register context while not running

	// While a thread blocks in sceKernelWaitEventFlag or sceKernelWaitSema: the
	// object handle and the wait condition, re-evaluated on every set/signal.
	waitEv     uint32 // event flag handle (0 = not waiting on a flag)
	waitBits   uint32
	waitMode   uint32
	waitOutPtr uint32
	waitSema   uint32 // semaphore handle (0 = not waiting on a semaphore)
	waitNeed   int32
	wakeVblank uint32 // wake when the VBlank counter reaches this (timed waits)
}

// evMatch checks (and on success consumes) an event-flag wait condition: mode
// bit 0 = OR (any requested bit) vs AND (all), 0x20 clears the requested bits,
// 0x40 clears the whole flag; the pre-clear pattern is written to outPtr.
func (m *Machine) evMatch(o *kobject, bits, mode, outPtr uint32) bool {
	ok := o.bits&bits == bits
	if mode&1 != 0 {
		ok = o.bits&bits != 0
	}
	if !ok {
		return false
	}
	if outPtr != 0 {
		m.write32(outPtr, o.bits)
	}
	if mode&0x20 != 0 {
		o.bits &^= bits
	}
	if mode&0x40 != 0 {
		o.bits = 0
	}
	return true
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
	"sceKernelDeleteCallback",
	// sceAudio
	"sceAudioChReserve", "sceAudioChRelease", "sceAudioOutputBlocking", "sceAudioOutputPannedBlocking",
	"sceAudioSetChannelDataLen", "sceAudioGetChannelRestLen", "sceAudioChangeChannelVolume",
	"sceAudioOutput2Reserve", "sceAudioOutput2OutputBlocking", "sceAudioOutput2Release",
	"sceAudioSRCChReserve", "sceAudioSRCOutputBlocking", "sceAudioSRCChRelease",
	"sceKernelCreateEventFlag", "sceKernelWaitEventFlag", "sceKernelSetEventFlag", "sceKernelClearEventFlag",
	"sceKernelCreateMutex", "sceKernelLockMutex", "sceKernelUnlockMutex",
	"sceKernelCreateMsgPipe", "sceKernelGetSystemTimeWide", "sceKernelGetSystemTimeLow",
	"sceKernelGetThreadCurrentPriority", "sceKernelChangeThreadPriority",
	"sceKernelCreateVpl", "sceKernelDeleteVpl", "sceKernelAllocateVpl", "sceKernelAllocateVplCB", "sceKernelTryAllocateVpl",
	"sceKernelFreeVpl", "sceKernelReferVplStatus", "sceKernelCreateFpl", "sceKernelDeleteFpl",
	"sceKernelAllocateFpl", "sceKernelTryAllocateFpl", "sceKernelFreeFpl",
	"sceKernelChangeCurrentThreadAttr", "sceKernelCreateMbx", "sceKernelCreateVTimer",
	"sceKernelPollSema", "sceKernelSignalSema", "sceKernelDeleteSema",
	"sceKernelWaitEventFlagCB", "sceKernelPollEventFlag", "sceKernelDeleteEventFlag",
	"sceKernelCheckCallback", "sceKernelNotifyCallback", "sceKernelRotateThreadReadyQueue",
	"sceKernelWaitThreadEnd", "sceKernelWaitThreadEndCB", "sceKernelSuspendThread",
	"sceKernelResumeThread", "sceKernelGetSystemTime", "sceKernelUSec2SysClock",
	"sceKernelSysClock2USec", "sceKernelDelaySysClockThread", "sceKernelSetGPO",
	// Kernel_Library (interrupt control, used for critical sections)
	"sceKernelCpuSuspendIntr", "sceKernelCpuResumeIntr", "sceKernelCpuResumeIntrWithSync",
	// InterruptManager
	"sceKernelRegisterSubIntrHandler", "sceKernelReleaseSubIntrHandler",
	"sceKernelEnableSubIntr", "sceKernelDisableSubIntr",
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
	"sceDisplayWaitVblankStartCB", "sceDisplayWaitVblank", "sceDisplayWaitVblankCB",
	"sceDisplayGetVcount", "sceDisplayIsVblank", "sceDisplayGetCurrentHcount",
	"sceDisplayGetAccumulatedHcount", "sceDisplayGetFramePerSec",
	// sceGe_user
	"sceGeEdramGetAddr", "sceGeEdramGetSize", "sceGeListEnQueue", "sceGeListEnQueueHead",
	"sceGeListSync", "sceGeDrawSync", "sceGeSetCallback", "sceGeUnsetCallback",
	"sceGeListUpdateStallAddr", "sceGeBreak", "sceGeContinue", "sceGeGetCmd", "sceGeGetMtx",
	// sceCtrl
	"sceCtrlSetSamplingCycle", "sceCtrlSetSamplingMode", "sceCtrlGetSamplingCycle",
	"sceCtrlGetSamplingMode", "sceCtrlReadBufferPositive", "sceCtrlPeekBufferPositive",
	"sceCtrlReadLatch", "sceCtrlPeekLatch", "sceCtrlSetIdleCancelThreshold",
	// StdioForUser
	"sceKernelStdout", "sceKernelStdin", "sceKernelStderr",
	// IoFileMgrForUser
	"sceIoOpen", "sceIoClose", "sceIoRead", "sceIoWrite", "sceIoLseek", "sceIoLseek32",
	"sceIoOpenAsync", "sceIoCloseAsync", "sceIoReadAsync", "sceIoLseekAsync", "sceIoLseek32Async",
	"sceIoWaitAsync", "sceIoWaitAsyncCB", "sceIoPollAsync", "sceIoGetAsyncStat",
	"sceIoChangeAsyncPriority", "sceIoSetAsyncCallback",
	"sceIoGetstat", "sceIoChdir", "sceIoDevctl", "sceIoIoctl", "sceIoIoctlAsync",
	"sceIoDopen", "sceIoDread", "sceIoDclose", "sceIoSync", "sceIoRemove", "sceIoMkdir",
	// sceUmdUser
	"sceUmdActivate", "sceUmdDeactivate", "sceUmdCheckMedium", "sceUmdGetDriveStat",
	"sceUmdWaitDriveStat", "sceUmdWaitDriveStatCB", "sceUmdWaitDriveStatWithTimer",
	"sceUmdRegisterUMDCallBack", "sceUmdUnRegisterUMDCallBack", "sceUmdGetErrorStat",
	// sceRtc / scePower
	"sceRtcGetCurrentTick", "scePowerGetCpuClockFrequency", "scePowerRegisterCallback",
	// sceUtility
	"sceUtilityGetSystemParamInt", "sceUtilityGetSystemParamString",
	"sceUtilityMsgDialogInitStart", "sceUtilityMsgDialogGetStatus", "sceUtilityMsgDialogUpdate",
	"sceUtilityMsgDialogShutdownStart",
	"sceUtilitySavedataInitStart", "sceUtilitySavedataGetStatus", "sceUtilitySavedataUpdate",
	"sceUtilitySavedataShutdownStart",
	// sceSasCore
	"__sceSasInit", "__sceSasCore", "__sceSasCoreWithMix", "__sceSasGetEndFlag",
	"__sceSasGetAllEnvelopeHeights", "__sceSasSetVoice", "__sceSasSetVoicePCM", "__sceSasSetPitch",
	"__sceSasSetVolume", "__sceSasSetSL", "__sceSasSetADSR", "__sceSasSetADSRmode",
	"__sceSasSetSimpleADSR", "__sceSasSetKeyOn", "__sceSasSetKeyOff", "__sceSasSetPause",
	"__sceSasGetPauseFlag", "__sceSasSetNoise", "__sceSasSetGrain", "__sceSasGetGrain",
	"__sceSasSetOutputmode", "__sceSasGetOutputmode", "__sceSasRevType", "__sceSasRevParam",
	"__sceSasRevEVOL", "__sceSasRevVON",
	// sceMpeg (PSMF movie player)
	"sceMpegInit", "sceMpegFinish", "sceMpegQueryMemSize", "sceMpegCreate", "sceMpegDelete",
	"sceMpegRegistStream", "sceMpegUnRegistStream", "sceMpegMallocAvcEsBuf", "sceMpegFreeAvcEsBuf",
	"sceMpegInitAu", "sceMpegGetAvcAu", "sceMpegGetAtracAu", "sceMpegGetPcmAu",
	"sceMpegQueryStreamOffset", "sceMpegQueryStreamSize", "sceMpegQueryAtracEsSize",
	"sceMpegAvcDecode", "sceMpegAvcDecodeMode", "sceMpegAvcDecodeStop", "sceMpegAvcDecodeYCbCr",
	"sceMpegAvcDecodeStopYCbCr", "sceMpegAvcQueryYCbCrSize", "sceMpegAvcInitYCbCr",
	"sceMpegAvcCsc", "sceMpegAtracDecode", "sceMpegChangeGetAvcAuMode", "sceMpegChangeGetAuMode",
	"sceMpegRingbufferQueryMemSize", "sceMpegRingbufferConstruct", "sceMpegRingbufferDestruct",
	"sceMpegRingbufferPut", "sceMpegRingbufferAvailableSize",
	"sceMpegFlushStream", "sceMpegFlushAllStream",
	// sceAtrac3plus
	"sceAtracGetAtracID", "sceAtracSetDataAndGetID", "sceAtracSetHalfwayBufferAndGetID",
	"sceAtracSetData", "sceAtracSetHalfwayBuffer", "sceAtracDecodeData", "sceAtracGetRemainFrame",
	"sceAtracGetStreamDataInfo", "sceAtracAddStreamData", "sceAtracGetSecondBufferInfo",
	"sceAtracSetSecondBuffer", "sceAtracGetNextDecodePosition", "sceAtracGetSoundSample",
	"sceAtracGetChannel", "sceAtracGetMaxSample", "sceAtracGetNextSample", "sceAtracGetBitrate",
	"sceAtracGetLoopStatus", "sceAtracSetLoopNum", "sceAtracResetPlayPosition",
	"sceAtracGetInternalErrorInfo", "sceAtracReleaseAtracID",
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
func (m *Machine) setRet64(v uint64) { // $v0 = low, $v1 = high
	m.CPU.SetReg(2, uint32(v))
	m.CPU.SetReg(3, uint32(v>>32))
}

// clock is the system time in microseconds, derived from the instruction count so
// it advances consistently with the synthetic VBlank: stepsPerVBlank steps = one
// 60 Hz frame = 16667 us.
func (m *Machine) clock() uint64 {
	return m.CPU.Steps * 16667 / stepsPerVBlank
}
func (m *Machine) newHandle(o *kobject) uint32 {
	h := m.nextHandle
	m.nextHandle++
	m.handles[h] = o
	return h
}

// formatPrintf renders a guest printf format with integer/string arguments taken
// from $a1.. and the stack, enough for kernel debug output (%s %d %u %x %c %f and
// width/flag prefixes are consumed; %f prints the raw word).
func (m *Machine) formatPrintf(format string, firstArg uint32) string {
	argn := firstArg
	nextArg := func() uint32 {
		var v uint32
		if argn < 4 {
			v = m.CPU.Reg(4 + argn)
		} else {
			v = m.read32(m.CPU.Reg(29) + argn*4)
		}
		argn++
		return v
	}
	var out []byte
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch != '%' {
			out = append(out, ch)
			continue
		}
		j := i + 1
		for j < len(format) && (format[j] == '-' || format[j] == '0' || format[j] == '+' ||
			format[j] == ' ' || format[j] == '.' || (format[j] >= '0' && format[j] <= '9')) {
			j++
		}
		if j >= len(format) {
			break
		}
		switch format[j] {
		case '%':
			out = append(out, '%')
		case 's':
			out = append(out, m.cstr(nextArg())...)
		case 'd', 'i':
			out = append(out, fmt.Sprintf("%d", int32(nextArg()))...)
		case 'u':
			out = append(out, fmt.Sprintf("%d", nextArg())...)
		case 'x', 'X', 'p':
			out = append(out, fmt.Sprintf("%x", nextArg())...)
		case 'c':
			out = append(out, byte(nextArg()))
		case 'f':
			out = append(out, fmt.Sprintf("<%08X>", nextArg())...)
		default:
			out = append(out, '%', format[j])
		}
		i = j
	}
	return string(out)
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
	case "sceKernelMaxFreeMemSize", "sceKernelTotalFreeMemSize":
		// Games size their heap zones from this; it must reflect the bump heap.
		return func(m *Machine) { m.setRet(m.heapEnd - m.heapPtr) }
	case "sceKernelCreateMutex",
		"sceKernelCreateMsgPipe", "sceKernelCreateMbx", "sceKernelCreateVTimer":
		// Kernel objects must be real positive uids: a returned 0 reads as failure
		// and cascades (heap zones guard themselves with semaphores).
		return func(m *Machine) {
			m.setRet(m.newHandle(&kobject{kind: "sync", name: m.cstr(m.arg(0))}))
		}
	case "sceKernelCreateSema":
		return func(m *Machine) {
			// (name, attr, initCount, maxCount, opt)
			m.setRet(m.newHandle(&kobject{
				kind: "sema", name: m.cstr(m.arg(0)), count: int32(m.arg(2)),
			}))
		}
	case "sceKernelWaitSema", "sceKernelWaitSemaCB":
		return func(m *Machine) {
			// (semaid, need, timeout*) — take the count or block until signalled.
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "sema" {
				m.setRet(0x800201C3) // ERROR_NOT_FOUND_SEMAPHORE
				return
			}
			need := int32(m.arg(1))
			if o.count >= need {
				o.count -= need
				m.setRet(0)
				return
			}
			if m.current == nil {
				m.note("WaitSema would block the module-start context")
				m.setRet(0x800201B8)
				return
			}
			m.setRet(0)
			m.current.waitSema = m.arg(0)
			m.current.waitNeed = need
			m.yieldCurrent(thWaiting)
		}
	case "sceKernelPollSema":
		return func(m *Machine) {
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "sema" {
				m.setRet(0x800201C3)
				return
			}
			need := int32(m.arg(1))
			if o.count >= need {
				o.count -= need
				m.setRet(0)
			} else {
				m.setRet(0x800201AD) // ERROR_SEMA_ZERO
			}
		}
	case "sceKernelSignalSema":
		return func(m *Machine) {
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "sema" {
				m.setRet(0x800201C3)
				return
			}
			o.count += int32(m.arg(1))
			for _, th := range m.handles {
				if th.kind == "thread" && th.tstate == thWaiting && th.waitSema == m.arg(0) &&
					o.count >= th.waitNeed {
					o.count -= th.waitNeed
					th.waitSema = 0
					th.tstate = thReady
				}
			}
			m.setRet(0)
		}
	case "sceKernelDeleteSema":
		return func(m *Machine) {
			delete(m.handles, m.arg(0))
			m.setRet(0)
		}
	case "sceKernelCreateEventFlag":
		return func(m *Machine) {
			// (name, attr, initPattern, opt)
			m.setRet(m.newHandle(&kobject{kind: "evflag", name: m.cstr(m.arg(0)), bits: m.arg(2)}))
		}
	case "sceKernelSetEventFlag":
		return func(m *Machine) {
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "evflag" {
				m.setRet(0x800201C9) // ERROR_NOT_FOUND_EVENT_FLAG
				return
			}
			o.bits |= m.arg(1)
			for _, th := range m.handles {
				if th.kind == "thread" && th.tstate == thWaiting && th.waitEv == m.arg(0) {
					if m.evMatch(o, th.waitBits, th.waitMode, th.waitOutPtr) {
						th.waitEv = 0
						th.tstate = thReady
					}
				}
			}
			m.setRet(0)
		}
	case "sceKernelClearEventFlag":
		return func(m *Machine) {
			if o := m.handles[m.arg(0)]; o != nil && o.kind == "evflag" {
				o.bits &= m.arg(1)
			}
			m.setRet(0)
		}
	case "sceKernelWaitEventFlag", "sceKernelWaitEventFlagCB":
		return func(m *Machine) {
			// (evid, bits, mode, outBits*, timeout*)
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "evflag" {
				m.setRet(0x800201C9)
				return
			}
			if m.evMatch(o, m.arg(1), m.arg(2), m.arg(3)) {
				m.setRet(0)
				return
			}
			if m.current == nil {
				m.note("WaitEventFlag would block the module-start context")
				m.setRet(0x800201B8) // ERROR_WAIT_TIMEOUT
				return
			}
			// Block: $v0 = 0 is committed into the saved context now; the pattern
			// is written to outBits when a SetEventFlag satisfies the wait.
			m.setRet(0)
			m.current.waitEv = m.arg(0)
			m.current.waitBits = m.arg(1)
			m.current.waitMode = m.arg(2)
			m.current.waitOutPtr = m.arg(3)
			m.yieldCurrent(thWaiting)
		}
	case "sceKernelPollEventFlag":
		return func(m *Machine) {
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "evflag" {
				m.setRet(0x800201C9)
				return
			}
			if m.evMatch(o, m.arg(1), m.arg(2), m.arg(3)) {
				m.setRet(0)
			} else {
				m.setRet(0x800201C1) // ERROR_EVENT_FLAG_POLL_FAILED
			}
		}
	case "sceKernelDeleteEventFlag":
		return func(m *Machine) {
			delete(m.handles, m.arg(0))
			m.setRet(0)
		}
	case "sceAudioChReserve":
		return func(m *Machine) {
			ch := m.arg(0)
			if int32(ch) < 0 {
				ch = m.audioCh
			}
			m.audioCh = (m.audioCh + 1) & 7
			m.setRet(ch)
		}
	case "sceAudioOutputBlocking", "sceAudioOutputPannedBlocking",
		"sceAudioSRCOutputBlocking", "sceAudioOutput2OutputBlocking":
		// The hardware would drain the sample buffer (~a frame's worth): a timed
		// wait, or the mixer threads starve everyone below their priority.
		return func(m *Machine) {
			m.setRet(0)
			if m.current != nil {
				m.current.wakeVblank = m.vblanks + 1
				m.yieldCurrent(thWaiting)
			}
		}
	case "sceKernelCreateCallback":
		return func(m *Machine) {
			// (name, function, arg)
			m.setRet(m.newHandle(&kobject{
				kind: "callback", name: m.cstr(m.arg(0)), entry: m.arg(1), addr: m.arg(2),
			}))
		}
	case "sceKernelCreateVpl", "sceKernelCreateFpl":
		return func(m *Machine) {
			// (name, partition, attr, size, opt) — carve the pool out of the bump
			// heap now; Allocate bump-allocates within it.
			size := m.arg(3)
			addr := m.heapPtr
			m.heapPtr = (m.heapPtr + size + 0xFF) &^ 0xFF
			m.setRet(m.newHandle(&kobject{
				kind: "vpl", name: m.cstr(m.arg(0)), addr: addr, size: size,
			}))
		}
	case "sceKernelAllocateVpl", "sceKernelAllocateVplCB", "sceKernelTryAllocateVpl":
		return func(m *Machine) {
			// (uid, size, void **data, timeout*)
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "vpl" {
				m.setRet(0x800201C7) // ERROR_NOT_FOUND_VPOOL
				return
			}
			size := (m.arg(1) + 7) &^ 7
			if o.used+size > o.size {
				m.setRet(0x800201D9) // ERROR_NO_MEMORY
				return
			}
			m.write32(m.arg(2), o.addr+o.used)
			o.used += size
			m.setRet(0)
		}
	case "sceKernelFreeVpl":
		return func(m *Machine) { m.setRet(0) } // pool frees are not reclaimed
	case "sceKernelReferVplStatus":
		return func(m *Machine) {
			// (uid, SceKernelVplInfo *info{size,name[32],attr,poolSize,freeSize,waiters})
			o := m.handles[m.arg(0)]
			if o == nil || o.kind != "vpl" {
				m.setRet(0x800201C7)
				return
			}
			p := m.arg(1)
			m.write32(p, 52)
			for i := uint32(0); i < 32; i++ {
				var b byte
				if int(i) < len(o.name) {
					b = o.name[i]
				}
				m.Write(p+4+i, b)
			}
			m.write32(p+36, 0)             // attr
			m.write32(p+40, o.size)        // poolSize
			m.write32(p+44, o.size-o.used) // freeSize
			m.write32(p+48, 0)             // numWaitThreads
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
				kind: "thread", name: m.cstr(m.arg(0)), entry: m.arg(1), addr: base,
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
			m.startThread(m.arg(0), o, m.arg(1), m.arg(2))
			m.setRet(0)
		}
	case "sceKernelExitGame", "sceKernelSelfStopUnloadModule":
		return func(m *Machine) {
			m.Halted = true
			m.HaltReason = "game requested exit (sceKernelExitGame)"
		}
	case "sceKernelExitThread", "sceKernelExitDeleteThread", "sceKernelTerminateDeleteThread":
		return func(m *Machine) { m.onThreadExit() }
	case "sceKernelRegisterSubIntrHandler":
		return func(m *Machine) {
			m.registerSubIntr(m.arg(0), m.arg(1), m.arg(2), m.arg(3))
			m.setRet(0)
		}
	case "sceKernelReleaseSubIntrHandler":
		return func(m *Machine) {
			delete(m.subIntrs, m.arg(0)<<16|m.arg(1))
			m.setRet(0)
		}
	case "sceKernelEnableSubIntr":
		return func(m *Machine) {
			m.setSubIntrEnabled(m.arg(0), m.arg(1), true)
			m.setRet(0)
		}
	case "sceKernelDisableSubIntr":
		return func(m *Machine) {
			m.setSubIntrEnabled(m.arg(0), m.arg(1), false)
			m.setRet(0)
		}
	case "sceDisplayGetVcount":
		return func(m *Machine) { m.setRet(m.vblanks) }
	case "sceKernelGetSystemTimeWide":
		return func(m *Machine) { m.setRet64(m.clock()) }
	case "sceKernelGetSystemTimeLow":
		return func(m *Machine) { m.setRet(uint32(m.clock())) }
	case "sceKernelGetSystemTime", "sceRtcGetCurrentTick":
		return func(m *Machine) { // write a u64 clock to *a0
			t := m.clock()
			m.write32(m.arg(0), uint32(t))
			m.write32(m.arg(0)+4, uint32(t>>32))
			m.setRet(0)
		}
	case "sceCtrlReadBufferPositive", "sceCtrlPeekBufferPositive":
		return func(m *Machine) { // fill one SceCtrlData with a neutral pad
			p := m.arg(0)
			m.write32(p+0, uint32(m.clock())) // timestamp
			m.write32(p+4, m.pad)             // buttons
			m.Write(p+8, 0x80)                // analog X center
			m.Write(p+9, 0x80)                // analog Y center
			for i := uint32(10); i < 16; i++ {
				m.Write(p+i, 0)
			}
			m.setRet(1) // one buffer read
		}
	case "sceCtrlReadLatch", "sceCtrlPeekLatch":
		return func(m *Machine) { // fill a SceCtrlLatch (make/break/press/release)
			p := m.arg(0)
			m.write32(p+0, m.pad&^m.padPrev) // uiMake: newly pressed
			m.write32(p+4, m.padPrev&^m.pad) // uiBreak: newly released
			m.write32(p+8, m.pad)            // uiPress: currently down
			m.write32(p+12, ^m.pad)          // uiRelease: currently up
			if strings.HasSuffix(name, "ReadLatch") {
				m.padPrev = m.pad // Read consumes the edge; Peek does not
			}
			m.setRet(1) // one latch sample
		}
	case "sceKernelSleepThread", "sceKernelSleepThreadCB":
		return func(m *Machine) { m.yieldCurrent(thWaiting) }
	case "sceKernelWakeupThread":
		return func(m *Machine) {
			if o := m.handles[m.arg(0)]; o != nil && o.kind == "thread" && o.tstate == thWaiting {
				o.tstate = thReady
			}
			m.setRet(0)
		}
	case "sceKernelDelayThread", "sceKernelDelayThreadCB":
		// A real timed wait — lower-priority threads must get the CPU, or a
		// polling loop starves every priority below it.
		return func(m *Machine) {
			usec := m.arg(0)
			m.setRet(0)
			if m.current == nil {
				return
			}
			frames := usec / 16667
			if frames == 0 {
				frames = 1
			}
			m.current.wakeVblank = m.vblanks + frames
			m.yieldCurrent(thWaiting)
		}
	case "sceDisplayWaitVblank", "sceDisplayWaitVblankCB",
		"sceDisplayWaitVblankStart", "sceDisplayWaitVblankStartCB":
		return func(m *Machine) {
			m.setRet(0)
			if m.current == nil {
				return
			}
			m.current.wakeVblank = m.vblanks + 1
			m.yieldCurrent(thWaiting)
		}
	case "sceDisplaySetFrameBuf":
		return func(m *Machine) {
			// (topaddr, bufferwidth, pixelformat, sync)
			m.fbAddr = m.arg(0)
			m.fbWidth = m.arg(1)
			m.fbFormat = m.arg(2)
			m.setRet(0)
		}
	case "sceDisplayGetFrameBuf":
		return func(m *Machine) {
			// (topaddr*, bufferwidth*, pixelformat*, sync) — games verify their
			// display setup against this before advancing past init.
			m.write32(m.arg(0), m.fbAddr)
			m.write32(m.arg(1), m.fbWidth)
			m.write32(m.arg(2), m.fbFormat)
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
	case "sceKernelGetThreadId":
		// The current thread's real handle: libc keys its per-thread reentrancy
		// structures on this, so it must be honest per thread.
		return func(m *Machine) { m.setRet(m.currentThreadID()) }
	case "sceKernelGetModuleIdByAddress", "sceKernelGetModuleId":
		return func(m *Machine) { m.setRet(1) }
	case "sceKernelGetGPI":
		// The devkit DIP switches; bit 0 enables the game's own debug logging,
		// which lands in the TTY — the game diagnoses itself.
		return func(m *Machine) { m.setRet(1) }
	case "sceKernelPrintf":
		return func(m *Machine) {
			m.tty = append(m.tty, m.formatPrintf(m.cstr(m.arg(0)), 1)...)
			m.setRet(0)
		}
	case "sceKernelStdin":
		return func(m *Machine) { m.setRet(0) }
	case "sceKernelStdout":
		return func(m *Machine) { m.setRet(fdStdout) }
	case "sceKernelStderr":
		return func(m *Machine) { m.setRet(fdStderr) }
	case "sceIoOpen":
		return func(m *Machine) { m.setRet(m.ioOpen(m.cstr(m.arg(0)))) }
	case "sceIoClose":
		return func(m *Machine) { m.setRet(m.ioClose(m.arg(0))) }
	case "sceIoRead":
		return func(m *Machine) { m.setRet(m.ioRead(m.arg(0), m.arg(1), m.arg(2))) }
	case "sceIoWrite":
		return func(m *Machine) { m.setRet(m.ioWrite(m.arg(0), m.arg(1), m.arg(2))) }
	case "sceIoLseek":
		return func(m *Machine) {
			// (fd, SceOff offset, whence): the 64-bit offset rides the aligned
			// pair $a2/$a3; whence lands in $t0 (the PSP passes 8 args in regs).
			off := int64(uint64(m.arg(3))<<32 | uint64(m.arg(2)))
			m.setRet64(uint64(m.ioLseek(m.arg(0), off, m.arg(4))))
		}
	case "sceIoGetstat":
		return func(m *Machine) { m.setRet(m.ioGetstat(m.cstr(m.arg(0)), m.arg(1))) }
	case "sceIoDopen":
		return func(m *Machine) { m.setRet(m.ioDopen(m.cstr(m.arg(0)))) }
	case "sceIoDread":
		return func(m *Machine) { m.setRet(m.ioDread(m.arg(0), m.arg(1))) }
	case "sceIoDclose":
		return func(m *Machine) { m.setRet(m.ioDclose(m.arg(0))) }
	case "sceIoOpenAsync":
		return func(m *Machine) { m.setRet(m.ioOpenAsync(m.cstr(m.arg(0)))) }
	case "sceIoReadAsync":
		return func(m *Machine) { m.setRet(m.ioReadAsync(m.arg(0), m.arg(1), m.arg(2))) }
	case "sceIoCloseAsync":
		return func(m *Machine) { m.setRet(m.ioCloseAsync(m.arg(0))) }
	case "sceIoLseekAsync":
		return func(m *Machine) {
			off := int64(uint64(m.arg(3))<<32 | uint64(m.arg(2)))
			m.setRet(m.ioLseekAsync(m.arg(0), off, m.arg(4)))
		}
	case "sceIoLseek32Async":
		return func(m *Machine) {
			m.setRet(m.ioLseekAsync(m.arg(0), int64(int32(m.arg(1))), m.arg(2)))
		}
	case "sceIoWaitAsync", "sceIoWaitAsyncCB", "sceIoGetAsyncStat":
		return func(m *Machine) { m.setRet(m.ioWaitAsync(m.arg(0), m.arg(1))) }
	case "sceIoPollAsync":
		return func(m *Machine) { m.setRet(m.ioPollAsync(m.arg(0), m.arg(1))) }
	case "sceIoChangeAsyncPriority", "sceIoSetAsyncCallback":
		return func(m *Machine) { m.setRet(0) }
	case "sceIoLseek32":
		return func(m *Machine) {
			m.setRet(uint32(m.ioLseek(m.arg(0), int64(int32(m.arg(1))), m.arg(2))))
		}
	case "sceUmdActivate", "sceUmdDeactivate", "sceUmdWaitDriveStat",
		"sceUmdWaitDriveStatCB", "sceUmdWaitDriveStatWithTimer":
		return func(m *Machine) { m.setRet(0) }
	case "sceMpegQueryMemSize":
		return func(m *Machine) { m.setRet(0x10000) }
	case "sceMpegRingbufferQueryMemSize":
		// packets * (one 2048-byte packet + the 104-byte packet header)
		return func(m *Machine) { m.setRet(m.arg(0) * (mpegPacketSize + 104)) }
	case "sceMpegCreate":
		// (SceMpeg*, data, size, SceMpegRingbuffer*, frameWidth, mode, ddrtop):
		// the handle points at the caller's workspace.
		return func(m *Machine) {
			m.write32(m.arg(0), m.arg(1))
			m.mpeg.Handle = m.arg(1)
			if rb := m.arg(3); rb != 0 {
				m.write32(rb+40, m.arg(1)) // ringbuffer's mpeg backlink
			}
			m.note("sceMpegCreate: handle 0x%08X, ringbuffer 0x%08X", m.arg(1), m.arg(3))
			m.setRet(0)
		}
	case "sceMpegDelete", "sceMpegFinish", "sceMpegRingbufferDestruct":
		return func(m *Machine) {
			m.mpeg = mpegState{}
			m.setRet(0)
		}
	case "sceMpegRingbufferConstruct":
		return func(m *Machine) {
			m.setRet(m.mpegRingbufferConstruct(m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4), m.arg(5)))
		}
	case "sceMpegRingbufferAvailableSize":
		return func(m *Machine) { m.setRet(m.mpeg.Packets - m.mpeg.In) }
	case "sceMpegRingbufferPut":
		return func(m *Machine) { m.setRet(m.mpegRingbufferPut(m.arg(0), m.arg(1), m.arg(2))) }
	case "sceMpegQueryStreamOffset":
		// (mpeg, buffer, offset*): the buffer holds the PSMF container header
		// (big-endian): magic "PSMF", version +4, stream offset +8, size +12.
		return func(m *Machine) {
			buf, out := m.arg(1), m.arg(2)
			if m.Read(buf) != 'P' || m.Read(buf+1) != 'S' || m.Read(buf+2) != 'M' || m.Read(buf+3) != 'F' {
				m.note("sceMpegQueryStreamOffset: no PSMF magic at 0x%08X", buf)
				m.write32(out, 0)
				m.setRet(errMpegInvalid)
				return
			}
			m.write32(out, m.beGuest32(buf+8))
			m.setRet(0)
		}
	case "sceMpegQueryStreamSize":
		return func(m *Machine) {
			m.write32(m.arg(1), m.beGuest32(m.arg(0)+12))
			m.setRet(0)
		}
	case "sceMpegAvcQueryYCbCrSize":
		// (mpeg, mode, width, height, size*): a 4:2:0 YCbCr frame buffer
		return func(m *Machine) {
			w, h := m.arg(2), m.arg(3)
			m.write32(m.arg(4), w*h*3/2)
			m.setRet(0)
		}
	case "sceMpegRegistStream":
		// (mpeg, streamType, streamNum) -> an opaque nonzero stream handle
		return func(m *Machine) { m.setRet(0x2000 + m.arg(1)) }
	case "sceMpegMallocAvcEsBuf":
		return func(m *Machine) { m.setRet(1) }
	case "sceMpegInitAu":
		// (mpeg, esBuffer, SceMpegAu*): timestamps start out invalid (-1)
		return func(m *Machine) {
			au := m.arg(2)
			for i := uint32(0); i < 16; i += 4 {
				m.write32(au+i, 0xFFFFFFFF)
			}
			m.write32(au+16, m.arg(1))
			m.write32(au+20, 0)
			m.setRet(0)
		}
	case "sceMpegGetAvcAu", "sceMpegGetAtracAu", "sceMpegGetPcmAu":
		// (mpeg, streamHandle, SceMpegAu*, attr*)
		return func(m *Machine) {
			r := m.mpegGetAu(m.arg(2))
			if r == 0 && m.arg(3) != 0 {
				m.write32(m.arg(3), 0)
			}
			m.setRet(r)
		}
	case "sceMpegAvcDecodeYCbCr", "sceMpegAtracDecode":
		// (mpeg, au, buffer, init*): report a frame produced; no decoder here —
		// the pixels are not modelled.
		return func(m *Machine) {
			if p := m.arg(3); p != 0 {
				m.write32(p, 1)
			}
			m.setRet(0)
		}
	case "sceMpegAvcDecode":
		// (mpeg, au, frameWidth, buffer, init*)
		return func(m *Machine) {
			if p := m.arg(4); p != 0 {
				m.write32(p, 1)
			}
			m.setRet(0)
		}
	case "sceMpegFlushStream", "sceMpegFlushAllStream":
		return func(m *Machine) {
			m.mpeg.In = 0
			if rb := m.mpeg.Ringbuf; rb != 0 {
				m.write32(rb+12, m.mpeg.Packets)
			}
			m.setRet(0)
		}
	case "sceAtracSetDataAndGetID", "sceAtracSetHalfwayBufferAndGetID":
		// (buffer, bufferSize[, readSize]) -> atrac id
		return func(m *Machine) {
			id := m.nextAtrac
			m.nextAtrac++
			a := &atracState{Buf: m.arg(0), Size: m.arg(1)}
			m.atracParseRiff(a)
			m.atrac[id] = a
			m.note("sceAtracSetDataAndGetID(0x%08X, %d) -> id %d (%d frames, %d ch)",
				a.Buf, a.Size, id, a.Frames, a.Channels)
			m.setRet(id)
		}
	case "sceAtracDecodeData":
		// (id, outPcm, samples*, end*, remainFrames*)
		return func(m *Machine) {
			m.setRet(m.atracDecode(m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.arg(4)))
		}
	case "sceAtracGetRemainFrame":
		return func(m *Machine) {
			if a := m.atrac[m.arg(0)]; a != nil {
				m.write32(m.arg(1), a.Frames-a.Pos)
				m.setRet(0)
			} else {
				m.setRet(errAtracBadID)
			}
		}
	case "sceAtracGetStreamDataInfo":
		// (id, writePtr*, writableBytes*, readOffset*): the whole file is in the
		// buffer already — nothing more to stream in.
		return func(m *Machine) {
			a := m.atrac[m.arg(0)]
			if a == nil {
				m.setRet(errAtracBadID)
				return
			}
			if p := m.arg(1); p != 0 {
				m.write32(p, a.Buf)
			}
			if p := m.arg(2); p != 0 {
				m.write32(p, 0)
			}
			if p := m.arg(3); p != 0 {
				m.write32(p, a.Size)
			}
			m.setRet(0)
		}
	case "sceAtracGetNextSample", "sceAtracGetMaxSample":
		return func(m *Machine) {
			if a := m.atrac[m.arg(0)]; a != nil {
				n := uint32(atracMaxSamples)
				if a.Pos >= a.Frames {
					n = 0
				}
				m.write32(m.arg(1), n)
				m.setRet(0)
			} else {
				m.setRet(errAtracBadID)
			}
		}
	case "sceAtracGetChannel":
		return func(m *Machine) {
			if a := m.atrac[m.arg(0)]; a != nil {
				m.write32(m.arg(1), a.Channels)
				m.setRet(0)
			} else {
				m.setRet(errAtracBadID)
			}
		}
	case "sceAtracGetNextDecodePosition":
		return func(m *Machine) {
			a := m.atrac[m.arg(0)]
			if a == nil {
				m.setRet(errAtracBadID)
				return
			}
			if a.Pos >= a.Frames {
				m.setRet(errAtracAllDecoded)
				return
			}
			m.write32(m.arg(1), a.Pos*atracMaxSamples)
			m.setRet(0)
		}
	case "sceAtracReleaseAtracID":
		return func(m *Machine) {
			delete(m.atrac, m.arg(0))
			m.setRet(0)
		}
	case "sceUmdCheckMedium":
		return func(m *Machine) { m.setRet(1) } // medium present
	case "sceUtilitySavedataInitStart":
		return func(m *Machine) {
			// (SceUtilitySavedataParam*): the dialog-common header is 48 bytes
			// (size, language, buttonSwap, four thread priorities, result at
			// +28, reserved); the operation mode follows at +48. There is no
			// memory stick with save data, so the load modes complete with
			// SCE_UTILITY_SAVEDATA_ERROR_LOAD_NO_DATA in the result field and
			// the game proceeds as a fresh start.
			p := m.arg(0)
			mode := m.read32(p + 48)
			var result uint32
			switch mode {
			case 0, 2, 4: // AUTOLOAD / LOAD / LISTLOAD
				result = 0x80110307 // no save data
			}
			m.write32(p+28, result)
			m.savedataStatus = 1 // INIT
			m.setRet(0)
		}
	case "sceUtilitySavedataGetStatus":
		return func(m *Machine) {
			// The dialog runs without interaction: INIT is reported once, then
			// RUNNING, then FINISHED until the game shuts the utility down.
			st := m.savedataStatus
			m.setRet(st)
			switch st {
			case 1, 2:
				m.savedataStatus = st + 1
			case 4:
				m.savedataStatus = 0
			}
		}
	case "sceUtilitySavedataUpdate":
		return func(m *Machine) { m.setRet(0) }
	case "sceUtilitySavedataShutdownStart":
		return func(m *Machine) {
			m.savedataStatus = 4
			m.setRet(0)
		}
	case "sceUtilityGetSystemParamInt":
		return func(m *Machine) {
			// (id, *value): 8 = language (1 = English), 9 = button assign
			// (1 = cross confirms). Others read as 0.
			var v uint32
			switch m.arg(0) {
			case 8, 9:
				v = 1
			}
			m.write32(m.arg(1), v)
			m.setRet(0)
		}
	case "sceUmdGetDriveStat":
		return func(m *Machine) { m.setRet(0x32) } // present | ready | readable
	case "sceKernelDcacheWritebackAll", "sceKernelDcacheWritebackRange",
		"sceKernelDcacheWritebackInvalidateAll", "sceKernelIcacheInvalidateAll",
		"sceDisplaySetMode":
		return func(m *Machine) { m.setRet(0) }
	}
	return nil // stub: report success (0)
}
