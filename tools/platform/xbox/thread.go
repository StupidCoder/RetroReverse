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

	// suspendCount is NT's suspension axis, ORTHOGONAL to the wait state: a
	// suspended thread does not run, but a suspend never cancels a dispatcher
	// wait and — critically — a RESUME never satisfies one. Conflating the two
	// (an early model set any tsWaiting thread ready on NtResumeThread) let a
	// producer's ResumeThread "complete" the streaming pump's infinite
	// message-queue wait: the pump popped a NULL message, took the zero-length
	// close path, and double-released a streaming buffer slot into a negative
	// refcount — the wedge that kept OutRun's font/sprite loads (and with them
	// the whole attract sequence) from ever finishing.
	suspendCount int32

	// While waiting on dispatcher objects.
	waitAll  bool
	waitObjs []uint32 // KWAIT: object handles being waited on
	waitReg  int      // register to receive the wait result (EAX)

	// stackTop / stackLimit are what KPCR.NtTib.StackBase / StackLimit hold while
	// this thread runs — the kernel swaps them at every context switch. The TLS
	// area (PsCreateSystemThreadEx's TlsDataSize) is carved at the top of the stack
	// and reached RELATIVE to StackBase: the XAPI TLS "index" global is a NEGATIVE
	// dword offset (-TlsDataSize/4), so [FS:[4] + idx*4] resolves to the TLS area's
	// self-pointer, and through it every per-thread datum — GetLastError first among
	// them. A KPCR whose StackBase never changes hands makes every thread read one
	// shared slot in the middle of the boot thread's LIVE STACK — GetLastError then
	// returns stack noise, and the XMV movie loader, which distinguishes "read
	// pending" (ERROR_IO_PENDING) from failure by it, aborts the title movie.
	stackTop   uint32
	stackLimit uint32
}

// runnable reports whether the scheduler may pick this thread: ready and not
// suspended.
func (t *thread) runnable() bool { return t.state == tsReady && t.suspendCount == 0 }

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
		out = append(out, fmt.Sprintf("%s tid=%d %-8s prio=%d susp=%d PC=%08X wakeTick=%d waitObjs=%v",
			mark, t.id, t.state, t.priority, t.suspendCount, pc, t.wakeTick, t.waitObjs))
	}
	return out
}

// DebugCurrentTid reports the id of the thread the CPU is executing — the companion
// accessor to DebugThreads for single-step diagnostics that follow one thread.
func (m *Machine) DebugCurrentTid() int {
	if m.current == nil {
		return -1
	}
	return int(m.current.id)
}

// bootThread creates thread 0 — the entry context. Its ctx is not used while it runs
// (the live CPU is authoritative); it exists so KeGetCurrentThread has a KTHREAD to
// return and so the scheduler has a home for the boot context when it first blocks.
func (m *Machine) bootThread() {
	kt := m.allocKObject(0x100) // a KTHREAD-sized block in the kernel band
	t := &thread{
		id:         0,
		kthread:    kt,
		priority:   16, // normal
		state:      tsRunning,
		stackTop:   titleStackTop, // what setupKPCR published for the boot thread
		stackLimit: titleStackTop - titleStackSize,
	}
	m.nextTID = 1
	m.threads = append(m.threads, t)
	m.current = t
	m.objects[kt] = &kobject{kind: "thread", addr: kt, thread: t}
	// The boot thread's TLS area sits in the gap above its stack top (titleStackTop
	// up to the kernel band is unused); the XBE entry path may run the same XAPI
	// thread bootstrap as any other thread, and a zero here would send its template
	// copy to page zero.
	m.write32(kt+kthreadTlsData, titleStackTop)
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

// kthreadTlsData is the KTHREAD field holding the thread's TLS area pointer, read
// directly by title code (XAPI's start thunk: MOV EAX,FS:[0x28]; MOV EDX,[EAX+0x28]).
const kthreadTlsData = 0x28

// createThread materialises a ready thread whose start routine runs with two context
// arguments (the Xbox KSTART_ROUTINE convention: StartRoutine(ctx1, ctx2)). It carves
// a stack from the pool arena, seeds a fresh register context inheriting the live CPU's
// bus/hook pointers, and returns the thread. The thread does not run until the
// scheduler picks it.
func (m *Machine) createThread(entry, ctx1, ctx2 uint32, stackSize, tlsSize uint32, priority int32, suspended bool) *thread {
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
	// Carve the TLS area (TlsDataSize, dword-rounded) at the top of the stack and
	// initialise it from the XBE's TLS template — the real kernel's layout, which
	// KPCR.NtTib.StackBase publishes while the thread runs (switchTo).
	tlsData := stackBase + stackSize
	if tlsSize != 0 {
		tlsSize = align32(tlsSize, 4)
		tlsData -= tlsSize
		m.initTLSArea(tlsData, tlsSize)
	}
	sp := tlsData - 16
	kt := m.allocKObject(0x100)
	// KTHREAD.TlsData (+0x28): the title reads this field directly — XAPI's thread
	// start thunk takes its TLS area from here (KeGetCurrentThread via FS:[0x28],
	// then +0x28), so it must be coherent in guest memory, not just scheduler state.
	m.write32(kt+kthreadTlsData, tlsData)

	t := &thread{
		id:         m.nextTID,
		kthread:    kt,
		priority:   priority,
		state:      tsReady,
		stackTop:   stackBase + stackSize,
		stackLimit: stackBase,
	}
	if suspended {
		t.suspendCount = 1 // create-suspended: NtResumeThread drops the count to 0
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

// initTLSArea zeroes a freshly carved per-thread TLS area. The kernel only carves;
// the TEMPLATE fill is the title's job — XAPI's own thread-start thunk (0x45069 in
// OutRun) reads TlsData from KTHREAD+0x28, writes the [TlsData] = TlsData+4
// self-pointer, and REP MOVSDs the XBE TLS template in itself. The kernel's whole
// contract is: carve TlsDataSize at the stack top, point KTHREAD+0x28 and (while the
// thread runs) KPCR.NtTib.StackBase at it.
func (m *Machine) initTLSArea(addr, size uint32) {
	for i := uint32(0); i < size; i++ {
		m.Write(addr+i, 0)
	}
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
