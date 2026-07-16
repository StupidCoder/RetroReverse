package xbox

// thread.go defines the kernel objects the HLE hands out and the thread model. Xbox
// threads are kernel-scheduled (like the 3DS's Horizon threads, tools/platform/n3ds/
// thread.go, not the N64's user-level ones): the HLE must actually run them. Each
// thread is a saved x86.CPU register context plus a KTHREAD address in guest memory;
// the scheduler (sched.go) runs the highest-priority ready thread for a quantum of
// instructions, then switches.
//
// Phase B begins single-threaded: the XBE entry runs on the boot thread (thread 0,
// whose live state is Machine.CPU) until the title's own PsCreateSystemThreadEx calls
// materialise more. The kobject/handle machinery is here from the start so the
// dispatcher calls (events, semaphores, waits) have something to model the moment the
// boot path reaches them.

import (
	"fmt"

	"retroreverse.com/tools/cpu/x86"
)

// threadState is a thread's scheduler state.
type threadState int

const (
	tsReady threadState = iota
	tsRunning
	tsWaiting  // blocked on a dispatcher object, wakeable by a signal
	tsSleeping // KeDelayExecutionThread — wakes at a tick
	tsDead
)

func (s threadState) String() string {
	return [...]string{"ready", "running", "waiting", "sleeping", "dead"}[s]
}

// thread is one Xbox thread.
type thread struct {
	id       uint32
	kthread  uint32  // guest address of its KTHREAD (also its object handle)
	ctx      x86.CPU // saved register context (live in Machine.CPU while running)
	priority int32   // Xbox: higher value = higher priority (0..31)
	state    threadState
	wakeTick uint64 // tick to wake at, for sleeping threads

	// While waiting on dispatcher objects.
	waitAll  bool
	waitObjs []uint32 // KWAIT: object handles being waited on
	waitReg  int      // register to receive the wait result (EAX)
}

// kobject is a dispatcher object handed out by the HLE (event, semaphore, mutant,
// timer, thread). Its guest address is its handle; the header fields the title reads
// (signal state, count) are also written into guest memory at that address so code
// that inspects the object directly sees a coherent DISPATCHER_HEADER.
type kobject struct {
	kind     string
	addr     uint32 // guest address of the object (== handle)
	signaled bool
	count    int32   // semaphore count / mutant recursion
	limit    int32   // semaphore limit
	thread   *thread // for thread objects
}

// DebugThreads returns a one-line-per-thread summary of the scheduler state (id, state,
// priority, saved PC, wait targets) — a diagnostic accessor for the boot driver, like
// OrdinalHistogram. The current thread's live PC comes from the CPU, not its stale ctx.
func (m *Machine) DebugThreads() []string {
	out := make([]string, 0, len(m.threads))
	for _, t := range m.threads {
		pc := t.ctx.SegBase[x86.CS] + t.ctx.IP
		mark := " "
		if t == m.current {
			pc = m.CPU.SegBase[x86.CS] + m.CPU.IP
			mark = "*"
		}
		out = append(out, fmt.Sprintf("%s tid=%d %-8s prio=%d PC=%08X wakeTick=%d waitObjs=%v",
			mark, t.id, t.state, t.priority, pc, t.wakeTick, t.waitObjs))
	}
	return out
}

// bootThread creates thread 0 — the entry context. Its ctx is not used while it runs
// (the live CPU is authoritative); it exists so KeGetCurrentThread has a KTHREAD to
// return and so the scheduler has a home for the boot context when it first blocks.
func (m *Machine) bootThread() {
	kt := m.allocKObject(0x100) // a KTHREAD-sized block in the kernel band
	t := &thread{
		id:       0,
		kthread:  kt,
		priority: 16, // normal
		state:    tsRunning,
	}
	m.nextTID = 1
	m.threads = append(m.threads, t)
	m.current = t
	m.objects[kt] = &kobject{kind: "thread", addr: kt, thread: t}
	// KPCR.Prcb.CurrentThread points at the running thread's KTHREAD.
	m.write32(kpcrAddr+kpcrPrcbData+prcbCurrentThread, kt)
}

// currentKThread is the KTHREAD address of the running thread (KeGetCurrentThread).
func (m *Machine) currentKThread() uint32 {
	if m.current == nil {
		return 0
	}
	return m.current.kthread
}

// threadExitAddr is the return address a new thread starts with. When a start routine
// returns to it (rather than calling PsTerminateSystemThread itself), onStep treats it
// as an implicit thread exit. It sits just below the kernel-trap region, backed by
// nothing — it is a sentinel PC, never fetched.
const threadExitAddr = trapBase - 0x100

// createThread materialises a ready thread whose start routine runs with two context
// arguments (the Xbox KSTART_ROUTINE convention: StartRoutine(ctx1, ctx2)). It carves
// a stack from the pool arena, seeds a fresh register context inheriting the live CPU's
// bus/hook pointers, and returns the thread. The thread does not run until the
// scheduler picks it.
func (m *Machine) createThread(entry, ctx1, ctx2 uint32, stackSize uint32, priority int32, suspended bool) *thread {
	// Honour the caller's KernelStackSize (PsCreateSystemThreadEx arg 2), page-rounded;
	// 0 means the kernel default (16 KiB). Inflating small requests wastes the pool the
	// title budgets to the last few hundred KiB (see kernelBandBase).
	if stackSize == 0 {
		stackSize = 0x4000
	}
	stackSize = align32(stackSize, 0x1000)
	stackBase := m.allocPool(stackSize)
	if stackBase == 0 {
		// A silent 0 here becomes a thread running on low RAM — a fiction that
		// corrupts page zero and surfaces as unrelated crashes much later.
		m.CPU.Halt("createThread: pool exhausted allocating a %X-byte stack", stackSize)
	}
	sp := stackBase + stackSize - 16
	kt := m.allocKObject(0x100)

	t := &thread{
		id:       m.nextTID,
		kthread:  kt,
		priority: priority,
		state:    tsReady,
	}
	if suspended {
		t.state = tsWaiting // resumed by KeResumeThread / NtResumeThread
	}
	m.nextTID++

	// Seed the context from the live CPU so the bus and the OnStep/SegResolve/port
	// hooks (which close over this machine) are carried over, then reset the register
	// file to the thread's start state.
	t.ctx = *m.CPU
	t.ctx.Regs = [8]uint32{}
	t.ctx.IP = entry
	// KSTART_ROUTINE(ctx1, ctx2): push the two args and the exit-sentinel return addr.
	t.ctx.Regs[x86.SP] = sp
	m.pushCtx(&t.ctx, ctx2)
	m.pushCtx(&t.ctx, ctx1)
	m.pushCtx(&t.ctx, threadExitAddr)
	t.ctx.Halted = false
	t.ctx.HaltReason = ""

	m.threads = append(m.threads, t)
	m.objects[kt] = &kobject{kind: "thread", addr: kt, thread: t}
	m.logf("createThread id=%d entry=%08X ctx1=%08X ctx2=%08X sp=%08X prio=%d susp=%v -> KTHREAD %08X",
		t.id, entry, ctx1, ctx2, sp, priority, suspended, kt)
	return t
}

// pushCtx pushes a dword onto a (possibly non-running) thread's stack context.
func (m *Machine) pushCtx(c *x86.CPU, v uint32) {
	c.Regs[x86.SP] -= 4
	m.write32(c.Regs[x86.SP], v)
}

// exitCurrentThread ends the running thread and reschedules.
func (m *Machine) exitCurrentThread() {
	m.logf("thread %d exited (returned to sentinel) at tick %d", m.threadID(), m.tick)
	if m.current != nil {
		m.current.state = tsDead
		delete(m.objects, m.current.kthread)
	}
	m.reschedule = true
	m.dispatch()
}

func (m *Machine) threadID() uint32 {
	if m.current == nil {
		return 0
	}
	return m.current.id
}
