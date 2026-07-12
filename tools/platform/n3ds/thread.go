package n3ds

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// thread.go is the cooperative thread scheduler. Horizon threads are kernel-
// scheduled — unlike the N64's user-level libultra threads that only need an
// interrupt heartbeat — so the HLE must actually run them. Each thread is a saved
// ARM11 register/VFP context plus its own TLS page; the scheduler runs the highest-
// priority ready thread for a quantum of instructions, then switches, exactly the
// round-robin-with-park/wake shape of the DS dual-core scheduler
// (tools/platform/nds/dsmachine/run.go).
//
// The live state of the *current* thread is in Machine.CPU; every other thread's
// state lives in its ctx. switchTo swaps them with a whole-struct value copy — every
// arm.CPU field is a value type except the bus/SWI/Coproc pointers, which point at
// this machine and so survive the copy unchanged.
//
// A blocked svc yields cleanly: the svc handler sets the thread's state and
// m.reschedule, and returns normally (the PC advances past the svc as usual). The
// scheduler saves that context with the PC already past the svc; when the awaited
// object is later signalled, the waker writes the ABI result *into the blocked
// thread's saved ctx* and marks it ready, so on its next switch-in it resumes at the
// instruction after the svc with the right registers — no PC rewinding.

type threadState int

const (
	ready threadState = iota
	running
	waiting
	sleeping
	dead
)

func (s threadState) String() string {
	return [...]string{"ready", "running", "waiting", "sleeping", "dead"}[s]
}

// thread is one ARM11 thread of the process.
type thread struct {
	id       uint32
	handle   uint32
	ctx      arm.CPU // saved register/VFP context (live in Machine.CPU when running)
	tlsBase  uint32  // this thread's TLS page
	tpidr    uint32  // per-thread TPIDRURW scratch (CP15)
	priority int32   // lower is higher priority (Horizon convention)
	state    threadState
	wakeTick uint64   // tick to wake at, for sleeping threads
	waitAll  bool     // WaitSyncN: wait for all vs any
	waitOn   []uint32 // object handles being waited on
	arbAddr  uint32   // address-arbiter park address (0 = none)
}

// quantum is how many instructions a thread runs before the scheduler reconsiders.
// Small enough that a higher-priority wake (a signalled GSP thread) preempts
// promptly; large enough that switching overhead stays negligible.
const quantum = 2000

// threadExitSentinel is the LR a new thread starts with; if a thread function
// returns to it (rather than calling svcExitThread itself, as libctru does), the
// scheduler treats it as an implicit ExitThread.
const threadExitSentinel = 0xFFFF0000

// allocTLS hands out a fresh per-thread TLS page and returns its base.
func (m *Machine) allocTLS() uint32 {
	base := m.nextTLS
	m.nextTLS += tlsSize
	m.mapRegion("tls-thread", base, make([]byte, tlsSize))
	return base
}

// cmdBuf is the current thread's IPC command buffer (TLS + 0x80).
func (m *Machine) cmdBuf() uint32 { return m.curThread.tlsBase + 0x80 }

// createThread registers a runnable thread with the given entry/arg/stack/priority
// and returns its handle. Its context inherits the current CPU (for the bus/hook
// pointers and Arch), then the register file is reset to the thread's start state.
func (m *Machine) createThread(priority int32, entry, arg, stacktop uint32) uint32 {
	t := &thread{
		id:       m.nextThread,
		priority: priority,
		state:    ready,
		tlsBase:  m.allocTLS(),
	}
	m.nextThread++
	t.ctx = *m.CPU // inherit bus/SWI/Coproc + Arch (and VFP, harmless)
	t.ctx.R = [16]uint32{}
	t.ctx.R[0] = arg
	t.ctx.R[13] = stacktop
	t.ctx.R[14] = threadExitSentinel
	t.ctx.R[15] = entry &^ 1
	t.ctx.Thumb = entry&1 == 1
	t.ctx.Mode = arm.ModeSYS
	t.ctx.IRQDisable = false
	t.ctx.Halted = false

	h := m.newHandle("thread", false)
	m.handles[h].thread = t
	t.handle = h
	m.threads = append(m.threads, t)
	if m.Verbose {
		fmt.Printf("  createThread id=%d entry=0x%08X arg=0x%08X sp=0x%08X prio=%d -> handle 0x%08X\n",
			t.id, entry, arg, stacktop, priority, h)
	}
	return h
}

// svcSleepThread parks the current thread. r0:r1 = s64 nanoseconds; 0 is a plain
// yield, a positive value sleeps until the tick reaches the wake point, and a
// negative value (infinite) parks until explicitly woken.
func (m *Machine) svcSleepThread(c *arm.CPU) {
	ns := int64(uint64(c.R[0]) | uint64(c.R[1])<<32)
	c.R[0] = resultSuccess
	switch {
	case ns == 0:
		m.curThread.state = ready // yield
	case ns < 0:
		m.curThread.state = waiting // infinite
	default:
		m.curThread.state = sleeping
		m.curThread.wakeTick = m.tick + nsToTick(ns)
	}
	m.reschedule = true
}

// svcExitThread ends the current thread and wakes anyone joined on it.
func (m *Machine) svcExitThread(c *arm.CPU) {
	m.curThread.state = dead
	m.signalThreadExit(m.curThread)
	m.reschedule = true
}

// pickRunnable returns the highest-priority ready thread, rotating among equals
// (round-robin) so equal-priority threads share time. nil when none are ready.
func (m *Machine) pickRunnable() *thread {
	var best *thread
	n := len(m.threads)
	for i := 0; i < n; i++ {
		t := m.threads[(m.rrCursor+i)%n]
		if t.state == ready && (best == nil || t.priority < best.priority) {
			best = t
		}
	}
	if best != nil {
		m.rrCursor = (m.rrCursor + 1) % n
	}
	return best
}

// switchTo makes t the running thread, loading its context into the CPU.
func (m *Machine) switchTo(t *thread) {
	if m.curThread == t {
		t.state = running
		return
	}
	*m.CPU = t.ctx
	m.CPU.ClearExclusive() // a switch invalidates any pending LDREX (real OSes CLREX)
	m.curThread = t
	t.state = running
}

// soonestSleeper reports the earliest sleeping thread's wake tick, if any —
// one of the timed events the run loop's idle selection compares (run.go).
func (m *Machine) soonestSleeper() (uint64, bool) {
	var soonest uint64
	found := false
	for _, t := range m.threads {
		if t.state == sleeping && (!found || t.wakeTick < soonest) {
			soonest, found = t.wakeTick, true
		}
	}
	return soonest, found
}

// wakeDueSleepers readies every sleeping thread whose wake tick has passed.
// Called each scheduling iteration (not only when idle): a sleep deadline can
// pass while other threads execute, and a machine with a recurring event (the
// DSP audio frame) may never be idle at all.
func (m *Machine) wakeDueSleepers() {
	for _, t := range m.threads {
		if t.state == sleeping && t.wakeTick <= m.tick {
			t.state = ready
		}
	}
}

// wake sets a blocked thread ready and returns whether a reschedule is warranted
// (the woken thread outranks the current one).
func (m *Machine) wake(t *thread) bool {
	t.state = ready
	t.waitOn = nil
	t.arbAddr = 0
	return m.curThread == nil || t.priority < m.curThread.priority
}

// setResult writes an svc return value into a (possibly not-current) thread's
// context — used to deliver a wait's result. For the current thread it must go to
// the live CPU registers.
func (m *Machine) setResult(t *thread, reg int, v uint32) {
	if t == m.curThread {
		m.CPU.R[reg] = v
	} else {
		t.ctx.R[reg] = v
	}
}

// DumpThreads prints every thread's scheduler state and what it waits on, plus
// the handle table — the standard "what is everyone blocked on" bring-up
// instrument (bootoracle -threads).
func (m *Machine) DumpThreads() {
	m.dumpThreads()
	fmt.Printf("dsp: componentLoaded=%v state=%d semEvent=0x%08X nextFrame=%d (instrs=%d)\n",
		m.dsp.ComponentLoaded, m.dsp.State, m.dsp.SemEvent, m.dsp.NextFrame, m.instrs)
	fmt.Printf("handles:\n")
	for h, o := range m.handles {
		extra := ""
		if o.signal {
			extra = " signalled"
		}
		if o.kind == "thread" && o.thread != nil {
			extra += fmt.Sprintf(" (thread %d)", o.thread.id)
		}
		fmt.Printf("  0x%08X %-24s%s\n", h, o.kind, extra)
	}
}

// dumpThreads prints each thread's state and what it is blocked on — the
// diagnostic for a deadlock (which sync object nothing is signalling).
func (m *Machine) dumpThreads() {
	fmt.Printf("thread states at deadlock (%d GX commands pending):\n", len(m.gxPending))
	for _, t := range m.threads {
		wo := ""
		for _, h := range t.waitOn {
			kind := "?"
			if o := m.handles[h]; o != nil {
				kind = o.kind
			}
			wo += fmt.Sprintf(" 0x%08X(%s)", h, kind)
		}
		if t.arbAddr != 0 {
			wo += fmt.Sprintf(" arb@0x%08X", t.arbAddr)
		}
		fmt.Printf("  thread %d prio %d state %-8s pc=0x%08X sp=0x%08X lr=0x%08X waitOn:%s\n",
			t.id, t.priority, t.state, t.ctx.R[15], t.ctx.R[13], t.ctx.R[14], wo)
	}
}

// aliveThreads counts threads not yet dead.
func (m *Machine) aliveThreads() int {
	n := 0
	for _, t := range m.threads {
		if t.state != dead {
			n++
		}
	}
	return n
}

// tickToNs / nsToTick convert between the nominal system tick and nanoseconds.
// The 3DS CPU tick runs at ~268 MHz; the oracle advances m.tick by 2 per
// instruction, so this is a nominal mapping good enough for sleep ordering and
// the tick-vs-GetSystemTick timing loops the game runs.
const sysclockHz = 268111856

func nsToTick(ns int64) uint64 {
	if ns <= 0 {
		return 0
	}
	return uint64(ns) * sysclockHz / 1_000_000_000
}
