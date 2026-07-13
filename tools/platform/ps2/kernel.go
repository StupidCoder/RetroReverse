package ps2

// kernel.go is the EE kernel, high-level emulated.
//
// On real hardware the kernel lives in a 4 MiB BIOS ROM and is reached through the
// `syscall` instruction, with the call number in $v1 and arguments in $a0..$a3. That
// ROM is not on the game disc, and taking it from anywhere else is what the
// clean-room rule forbids — so the calls are Go functions instead.
//
// The numbers in the table below were not looked up. They were read out of the game:
// the boot ELF statically links Sony's SDK and ships its symbol table, so each
// wrapper — `FlushCache`, `CreateThread`, `sceSifSetDma` — is a named function whose
// body loads its own syscall number into $v1 and traps. Disassembling all 153 of them
// yields the whole table, derived from the image like everything else here. The
// negative numbers are real: a call made from interrupt context negates its number,
// which is how the kernel knows not to reschedule.
//
// Handlers come in three tiers, the discipline the PSP kernel established:
//
//	modelled    the call does what the hardware's would do
//	stubbed     the call returns a plausible success without doing the work
//	unmodelled  the call is logged with its arguments, and returns zero
//
// The third tier is the instrument, not a failure. One run prints every syscall the
// game made and how often, so the work list for the next phase is measured rather
// than guessed. SyscallCensus is that census, and it marks the unmodelled ones.

import (
	"fmt"
	"sort"

	"retroreverse.com/tools/cpu/r5900"
)

// arg reads syscall argument i (0..3) from $a0..$a3. Argument 4 is passed in $t0,
// which SetupThread needs.
func (m *Machine) arg(i uint32) uint32 { return uint32(m.CPU.Reg(4 + i)) }
func (m *Machine) argT0() uint32       { return uint32(m.CPU.Reg(8)) }

// setRet writes the syscall's return value to $v0.
func (m *Machine) setRet(v uint32) { m.CPU.SetReg(2, uint64(int64(int32(v)))) }

// handleSyscall is installed as the CPU's syscall hook. Returning true means the call
// was handled and execution continues after it.
func (m *Machine) handleSyscall(c *r5900.CPU) bool {
	num := int32(uint32(c.Reg(3))) // $v1
	if num < 0 {
		// The interrupt-context variant of the same call. Nothing in this model runs
		// with interrupts disabled in a way that would make the two differ, so they
		// share a handler — but the census keeps them apart, because a game that only
		// ever calls the "i" form is telling you where it thinks it is.
		num = -num
	}

	e, known := syscalls[uint32(num)]
	name := e.name
	if !known {
		name = fmt.Sprintf("syscall_%d", num)
	}
	m.SyscallCalls[name]++

	// A syscall the game has claimed for itself. SetSyscall lets a program replace a
	// kernel call with its own routine, and GOAL's kernel uses it heavily — `Copy` is
	// not the BIOS's memcpy at all, it is the game's own `kCopy`, installed at boot.
	//
	// Dispatch is a *jump*, not a call: the handler runs with $ra still pointing at
	// whoever called the SDK wrapper, and its own `jr $ra` returns there directly,
	// past the wrapper. That is what the real kernel's dispatcher does, and a model
	// that called it instead would return one frame too shallow.
	if h := m.userSyscalls[uint32(num)]; h != 0 {
		c.SetPC(uint64(h))
		return true
	}

	if e.fn == nil {
		m.note("%s (%d) unmodelled: a0=0x%08X a1=0x%08X a2=0x%08X a3=0x%08X, called from %s",
			name, num, m.arg(0), m.arg(1), m.arg(2), m.arg(3), m.Sym(uint32(c.CurPC())))
		m.setRet(0)
		return true
	}
	e.fn(m)
	return true
}

type syscallEntry struct {
	name string
	fn   func(m *Machine)
}

// syscalls is the EE kernel's call table, read out of the boot ELF's own SDK
// wrappers. Every entry is named; the ones with a handler are modelled or stubbed,
// and the rest are the work list.
var syscalls = map[uint32]syscallEntry{
	0: {"RFU000_FullReset", nil},
	1: {"ResetEE", func(m *Machine) { m.setRet(1) }},
	2: {"SetGsCrt", func(m *Machine) {
		m.gsInterlace, m.gsVideoMode, m.gsFieldMode = m.arg(0), m.arg(1), m.arg(2)
		m.note("SetGsCrt: interlace=%d mode=%d field=%d", m.arg(0), m.arg(1), m.arg(2))
		m.setRet(0)
	}},
	4: {"Exit", func(m *Machine) { m.Halt("the game called Exit(%d)", m.arg(0)) }},
	6: {"LoadExecPS2", nil},
	7: {"ExecPS2", nil},

	10: {"AddSbusIntcHandler", nil},
	11: {"RemoveSbusIntcHandler", nil},
	12: {"Interrupt2Iop", func(m *Machine) { m.setRet(0) }},
	13: {"SetVTLBRefillHandler", func(m *Machine) { m.setRet(0) }},
	14: {"SetVCommonHandler", func(m *Machine) { m.setRet(0) }},
	15: {"SetVInterruptHandler", func(m *Machine) { m.setRet(0) }},
	16: {"AddIntcHandler", func(m *Machine) { m.addIntcHandler() }},
	17: {"RemoveIntcHandler", func(m *Machine) { m.setRet(0) }},
	18: {"AddDmacHandler", func(m *Machine) { m.addDmacHandler() }},
	19: {"RemoveDmacHandler", func(m *Machine) { m.setRet(0) }},
	20: {"EnableIntc", func(m *Machine) { m.intcMask |= 1 << (m.arg(0) & 31); m.setRet(1) }},
	21: {"DisableIntc", func(m *Machine) { m.intcMask &^= 1 << (m.arg(0) & 31); m.setRet(1) }},
	22: {"EnableDmac", func(m *Machine) { m.dmacMask |= 1 << (m.arg(0) & 31); m.setRet(1) }},
	23: {"DisableDmac", func(m *Machine) { m.dmacMask &^= 1 << (m.arg(0) & 31); m.setRet(1) }},
	24: {"SetAlarm", nil},
	25: {"ReleaseAlarm", nil},

	32: {"CreateThread", func(m *Machine) { m.createThread() }},
	33: {"DeleteThread", func(m *Machine) { m.setRet(0) }},
	34: {"StartThread", func(m *Machine) { m.startThread() }},
	35: {"ExitThread", func(m *Machine) { m.exitThread() }},
	36: {"ExitDeleteThread", func(m *Machine) { m.exitThread() }},
	37: {"TerminateThread", func(m *Machine) { m.setRet(0) }},
	39: {"DisableDispatchThread", func(m *Machine) { m.setRet(0) }},
	40: {"EnableDispatchThread", func(m *Machine) { m.setRet(0) }},
	41: {"ChangeThreadPriority", func(m *Machine) { m.setRet(0) }},
	43: {"RotateThreadReadyQueue", func(m *Machine) { m.setRet(0) }},
	45: {"ReleaseWaitThread", func(m *Machine) { m.wakeupThread(m.arg(0)) }},
	47: {"GetThreadId", func(m *Machine) { m.setRet(m.currentThread) }},
	48: {"ReferThreadStatus", func(m *Machine) { m.referThreadStatus() }},
	50: {"SleepThread", func(m *Machine) { m.sleepThread() }},
	51: {"WakeupThread", func(m *Machine) { m.wakeupThread(m.arg(0)) }},
	52: {"iWakeupThread", func(m *Machine) { m.wakeupThread(m.arg(0)) }},
	53: {"CancelWakeupThread", func(m *Machine) { m.setRet(0) }},
	55: {"SuspendThread", func(m *Machine) { m.setRet(0) }},
	57: {"ResumeThread", func(m *Machine) { m.wakeupThread(m.arg(0)) }},
	59: {"JoinThread", func(m *Machine) { m.setRet(0) }},

	// The three calls crt0 makes before anything else exists.
	60: {"SetupThread", func(m *Machine) { m.setupThread() }},
	61: {"SetupHeap", func(m *Machine) { m.setupHeap() }},
	62: {"EndOfHeap", func(m *Machine) { m.setRet(m.heapEnd) }},

	64: {"CreateSema", func(m *Machine) { m.createSema() }},
	65: {"DeleteSema", func(m *Machine) { m.setRet(0) }},
	66: {"SignalSema", func(m *Machine) { m.signalSema(m.arg(0)) }},
	67: {"iSignalSema", func(m *Machine) { m.signalSema(m.arg(0)) }},
	68: {"WaitSema", func(m *Machine) { m.waitSema(m.arg(0)) }},
	69: {"PollSema", func(m *Machine) { m.pollSema(m.arg(0)) }},
	70: {"iPollSema", func(m *Machine) { m.pollSema(m.arg(0)) }},
	71: {"ReferSemaStatus", nil},

	74: {"SetOsdConfigParam", nil},
	75: {"GetOsdConfigParam", func(m *Machine) { m.setRet(0) }},
	76: {"GetGsHParam", nil},
	77: {"GetGsVParam", nil},
	78: {"SetGsHParam", nil},
	79: {"SetGsVParam", nil},

	85: {"PutTLBEntry", func(m *Machine) { m.setRet(0) }},
	86: {"SetTLBEntry", func(m *Machine) { m.setRet(0) }},
	87: {"GetTLBEntry", func(m *Machine) { m.setRet(0) }},
	88: {"ProbeTLBEntry", func(m *Machine) { m.setRet(0) }},
	89: {"ExpandScratchPad", func(m *Machine) { m.setRet(0) }},
	90: {"Copy", nil},
	91: {"GetEntryAddress", func(m *Machine) { m.setRet(0) }},
	92: {"EnableIntcHandler", func(m *Machine) { m.intcMask |= 1 << (m.arg(0) & 31); m.setRet(1) }},
	93: {"DisableIntcHandler", func(m *Machine) { m.intcMask &^= 1 << (m.arg(0) & 31); m.setRet(1) }},
	94: {"EnableDmacHandler", func(m *Machine) { m.dmacMask |= 1 << (m.arg(0) & 31); m.setRet(1) }},
	95: {"DisableDmacHandler", func(m *Machine) { m.dmacMask &^= 1 << (m.arg(0) & 31); m.setRet(1) }},

	96: {"KSeg0", func(m *Machine) { m.setRet(0) }},
	97: {"EnableCache", func(m *Machine) { m.setRet(0) }},
	98: {"DisableCache", func(m *Machine) { m.setRet(0) }},
	99: {"GetCop0", func(m *Machine) { m.setRet(uint32(m.CPU.COP0[m.arg(0)&31])) }},

	// The caches are not modelled, so a flush has nothing to do — but the call must
	// exist, because compiled code brackets every DMA with one.
	100: {"FlushCache", func(m *Machine) { m.setRet(0) }},
	102: {"CpuConfig", func(m *Machine) { m.setRet(0) }},

	107: {"sceSifStopDma", func(m *Machine) { m.setRet(0) }},
	108: {"SetCPUTimerHandler", func(m *Machine) { m.setRet(0) }},
	109: {"SetCPUTimer", func(m *Machine) { m.setRet(0) }},
	110: {"SetOsdConfigParam2", nil},
	111: {"GetOsdConfigParam2", func(m *Machine) { m.setRet(0) }},
	112: {"GsGetIMR", func(m *Machine) { m.setRet(m.gsIMR) }},
	113: {"GsPutIMR", func(m *Machine) { m.gsIMR = m.arg(0); m.setRet(0) }},
	114: {"SetPgifHandler", func(m *Machine) { m.setRet(0) }},
	115: {"SetVSyncFlag", func(m *Machine) {
		m.vsyncFlagPtr, m.vsyncFlag2Ptr = m.arg(0), m.arg(1)
		m.note("SetVSyncFlag: flags at 0x%08X and 0x%08X", m.arg(0), m.arg(1))
		m.setRet(0)
	}},
	116: {"SetSyscall", func(m *Machine) { m.setSyscall() }},

	118: {"sceSifDmaStat", func(m *Machine) { m.setRet(0xFFFFFFFF) }}, // nothing in flight
	119: {"sceSifSetDma", func(m *Machine) { m.sifSetDma() }},
	120: {"sceSifSetDChain", func(m *Machine) { m.setRet(0) }},
	121: {"sceSifSetReg", func(m *Machine) { m.sifSetReg() }},
	122: {"sceSifGetReg", func(m *Machine) { m.sifGetReg() }},

	123: {"ExecOSD", nil},
	124: {"Deci2Call", func(m *Machine) { m.setRet(0) }},
	125: {"PSMode", func(m *Machine) { m.setRet(0) }},
	126: {"MachineType", func(m *Machine) { m.setRet(0) }},
	127: {"GetMemorySize", func(m *Machine) { m.setRet(ramSize) }},
}

// --- the three calls crt0 makes ---------------------------------------------

// setupThread is syscall 60, the first thing the executable does after clearing its
// BSS. It is handed the global pointer, a stack, an argument block and the root
// function, and it *returns the stack pointer* — which crt0 immediately moves into
// $sp. Getting this wrong is not subtle: the program runs on a garbage stack and the
// first store faults.
//
// A stack of -1 means "put it at the top of memory".
func (m *Machine) setupThread() {
	gp, stack, stackSize, args := m.arg(0), m.arg(1), m.arg(2), m.arg(3)
	root := m.argT0()

	top := stack + stackSize
	if stack == 0xFFFFFFFF {
		top = ramSize // the top of main memory
	}
	// The kernel reserves a block at the top of the stack for the argument list, and
	// hands back a pointer below it, quadword-aligned.
	sp := (top - argBlockSize) &^ 0xF

	t := m.threads[m.currentThread]
	if t != nil {
		t.gp = gp
		t.stack = top - stackSize
		t.stackSz = stackSize
		t.entry = root
	}
	m.note("SetupThread: gp=0x%08X stack=0x%08X+0x%X args=0x%08X root=%s -> sp=0x%08X",
		gp, top-stackSize, stackSize, args, m.Sym(root), sp)

	// The argument list the program will be handed. There is no command line here —
	// the disc is the only thing that was "run" — so it is empty, which is what a
	// retail boot from the disc's own SYSTEM.CNF gives it too.
	m.argc = 0
	m.argv = 0

	m.setRet(sp)
}

// argBlockSize is the space the kernel keeps at the top of the stack for the argument
// list it copies there.
const argBlockSize = 0x2A0

// setSyscall is syscall 116: it replaces a kernel call with a routine of the
// program's own. A handler of zero puts the kernel's back.
//
// GOAL's kernel leans on this. It installs its own `kCopy` as syscall 90 and reaches
// it through the SDK's `Copy` wrapper, so a call that looks like the BIOS's memcpy is
// really the game's — which is why the unmodelled census reported `Copy` before this
// existed, even though the game had supplied the code for it.
func (m *Machine) setSyscall() {
	num, handler := m.arg(0), m.arg(1)
	if handler == 0 {
		delete(m.userSyscalls, num)
		m.note("SetSyscall %d: back to the kernel's", num)
	} else {
		m.userSyscalls[num] = handler
		m.note("SetSyscall %d -> %s", num, m.Sym(handler))
	}
	m.setRet(0)
}

// setupHeap is syscall 61: it establishes the malloc arena between the end of the
// program and the bottom of the stack. A size of -1 means "everything up to the
// stack".
func (m *Machine) setupHeap() {
	start, size := m.arg(0), m.arg(1)
	m.heapPtr = (start + 0xF) &^ 0xF
	if size == 0xFFFFFFFF {
		t := m.threads[m.currentThread]
		if t != nil && t.stack != 0 {
			m.heapEnd = t.stack
		} else {
			m.heapEnd = ramSize - 0x4000
		}
	} else {
		m.heapEnd = m.heapPtr + size
	}
	m.note("SetupHeap: 0x%08X..0x%08X (%d KiB)", m.heapPtr, m.heapEnd, (m.heapEnd-m.heapPtr)/1024)
	m.setRet(m.heapEnd)
}

// --- the census ---------------------------------------------------------------

// SyscallCensus renders which kernel calls the run made and how often, hottest first,
// marking the ones with no handler. Those are the work list.
func (m *Machine) SyscallCensus() string {
	if len(m.SyscallCalls) == 0 {
		return "no kernel calls\n"
	}
	handled := map[string]bool{}
	for _, e := range syscalls {
		if e.fn != nil {
			handled[e.name] = true
		}
	}

	type kv struct {
		name string
		n    int
	}
	var all []kv
	for k, v := range m.SyscallCalls {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].n != all[j].n {
			return all[i].n > all[j].n
		}
		return all[i].name < all[j].name
	})

	s := "kernel calls:\n"
	unmodelled := 0
	for _, e := range all {
		mark := " "
		if !handled[e.name] {
			mark = "*"
			unmodelled++
		}
		s += fmt.Sprintf("  %s %-24s %d\n", mark, e.name, e.n)
	}
	if unmodelled > 0 {
		s += fmt.Sprintf("  (* = unmodelled: %d call%s to write)\n", unmodelled, plural(unmodelled))
	}
	return s
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
