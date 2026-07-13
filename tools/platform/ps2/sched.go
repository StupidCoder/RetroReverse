package ps2

// sched.go is the EE kernel's thread scheduler, high-level emulated.
//
// The model is cooperative, as the PSP's is: a thread is a saved CPU state plus a
// priority, and a switch happens only where the kernel would have made one — a sleep,
// a wakeup, a thread exiting, or a vertical blank finding nothing runnable. Nothing
// here preempts, because nothing in a boot depends on being preempted.
//
// The thread the executable starts in is not created by CreateThread; it is simply
// the context the entry point runs in. It is thread 1, and it exists from the moment
// the machine loads an executable.

import "retroreverse.com/tools/cpu/r5900"

type threadState int

const (
	thRunning threadState = iota
	thReady
	thSleeping // waiting for WakeupThread
	thWaitSema // blocked on a semaphore
	thDormant  // created but not started
	thDead
)

func (s threadState) String() string {
	switch s {
	case thRunning:
		return "running"
	case thReady:
		return "ready"
	case thSleeping:
		return "sleeping"
	case thWaitSema:
		return "wait-sema"
	case thDormant:
		return "dormant"
	}
	return "dead"
}

type thread struct {
	id       uint32
	entry    uint32
	stack    uint32
	stackSz  uint32
	gp       uint32
	priority uint32
	state    threadState

	// wakeupCount is the kernel's counter, not a boolean: a WakeupThread that arrives
	// before the SleepThread is remembered, and the sleep returns immediately. A model
	// that used a flag loses the race and hangs.
	wakeupCount int32

	ctx r5900.State // saved when the thread is not running
}

// createThread is syscall 32. The struct it is handed a pointer to opens with the
// thread's *status*, not its entry point:
//
//	0x00 status
//	0x04 func            entry point
//	0x08 stack           stack base
//	0x0C stackSize
//	0x10 gp              global pointer
//	0x14 initialPriority
//	0x18 currentPriority
//	0x1C attr
//	0x20 option
//
// Reading it as though `func` came first shifts every field by one and yields a thread
// with a null entry point and a nonsense priority — which starts, jumps to zero, and
// faults, a long way from the mistake.
func (m *Machine) createThread() {
	p := m.arg(0)
	t := &thread{
		id:       m.nextThreadID,
		entry:    m.Read32(p + 0x04),
		stack:    m.Read32(p + 0x08),
		stackSz:  m.Read32(p + 0x0C),
		gp:       m.Read32(p + 0x10),
		priority: m.Read32(p + 0x14),
		state:    thDormant,
	}
	m.nextThreadID++
	m.threads[t.id] = t
	m.note("CreateThread %d: entry=%s stack=0x%08X+0x%X prio=%d",
		t.id, m.Sym(t.entry), t.stack, t.stackSz, t.priority)
	m.setRet(t.id)
}

// startThread makes a dormant thread runnable. The EE kernel does not switch to it
// immediately unless it outranks the caller — and a lower number is a *higher*
// priority — so the caller usually keeps running until it sleeps.
func (m *Machine) startThread() {
	id, arg := m.arg(0), m.arg(1)
	t := m.threads[id]
	if t == nil {
		m.note("StartThread %d: no such thread", id)
		m.setRet(0xFFFFFFFF)
		return
	}

	// The thread's initial context is taken from the *running* core and then
	// overwritten, rather than built from zero.
	//
	// That is not laziness. A CPU state carries more than registers — it carries the
	// TLB and the COP0 file, and those belong to the machine, not to the thread. A
	// context built from a zero value restores an empty TLB the moment the thread is
	// switched to, and every address it touches then faults, including addresses that
	// are plainly mapped. Snapshotting and overriding cannot forget a field; enumerating
	// the fields to copy can, and did.
	t.ctx = m.CPU.Snapshot()
	t.ctx.R = [32]r5900.Quad{}
	t.ctx.HI, t.ctx.LO, t.ctx.HI1, t.ctx.LO1, t.ctx.SA = 0, 0, 0, 0, 0
	t.ctx.DelaySlot, t.ctx.PendingDelay = false, false
	t.ctx.PC = uint64(t.entry)
	t.ctx.NextPC = uint64(t.entry) + 4
	t.ctx.R[4] = r5900.Quad{Lo: uint64(arg)}                         // $a0
	t.ctx.R[28] = r5900.Quad{Lo: uint64(t.gp)}                       // $gp
	t.ctx.R[29] = r5900.Quad{Lo: uint64(t.stack + t.stackSz - 0x10)} // $sp: stacks grow down
	t.ctx.R[31] = r5900.Quad{Lo: threadExitAddr}                     // $ra
	t.state = thReady

	m.setRet(id)

	cur := m.threads[m.currentThread]
	if cur != nil && t.priority < cur.priority {
		m.switchTo(t, thReady) // it outranks us: it runs now
	}
}

// threadExitAddr is the sentinel $ra a started thread returns to. The run loop traps
// a PC landing on it and retires the thread.
const threadExitAddr = 0x0F100000

func (m *Machine) sleepThread() {
	t := m.threads[m.currentThread]
	if t == nil {
		m.setRet(0)
		return
	}
	// A wakeup that arrived first is consumed here, and the sleep does not happen.
	if t.wakeupCount > 0 {
		t.wakeupCount--
		m.setRet(0)
		return
	}
	m.setRet(0)
	m.switchAway(thSleeping)
}

func (m *Machine) wakeupThread(id uint32) {
	t := m.threads[id]
	if t == nil {
		m.setRet(0xFFFFFFFF)
		return
	}
	if t.state == thSleeping {
		t.state = thReady
	} else {
		t.wakeupCount++ // remembered, so a later SleepThread returns at once
	}
	m.setRet(id)
}

// switchAway blocks the running thread in the given state and picks another. When
// nothing else can run, the machine goes *idle*: the CPU stops entirely, and only the
// clock advances until an interrupt or a reply from the IOP makes something ready.
//
// Idling is not an optimisation. A model that marks a thread blocked and then keeps
// executing it has not blocked it at all: SleepThread returns immediately, and every
// piece of code shaped like "ask the IOP, sleep, and read the answer when you wake" —
// which is all of them — reads the answer before it has arrived. The symptom is a
// blocking call that behaves like a polling one, and it is invisible until you notice
// the game re-asking a question it should already have the answer to.
func (m *Machine) switchAway(newState threadState) {
	if cur := m.threads[m.currentThread]; cur != nil {
		cur.ctx = m.CPU.Snapshot()
		cur.state = newState
	}
	next := m.pickReady()
	if next == nil {
		m.idle = true // nothing to run; the run loop will wait for something to happen
		return
	}
	next.state = thRunning
	m.currentThread = next.id
	m.CPU.Restore(next.ctx)
}

// resume looks for a thread to run after the machine has been idle. It reports whether
// one was found.
func (m *Machine) resume() bool {
	next := m.pickReady()
	if next == nil {
		return false
	}
	m.idle = false
	next.state = thRunning
	m.currentThread = next.id
	m.CPU.Restore(next.ctx)
	return true
}

// blocked reports whether anything could still wake the machine up. Nothing pending
// from the IOP, no interrupt handler to run, and no ready thread is a genuine
// deadlock rather than a wait.
func (m *Machine) blocked() bool {
	if len(m.sifPending) > 0 || len(m.intcHandlers) > 0 {
		return false
	}
	return m.pickReady() == nil
}

// switchTo saves the current thread and resumes another.
func (m *Machine) switchTo(next *thread, curState threadState) {
	if cur := m.threads[m.currentThread]; cur != nil {
		cur.ctx = m.CPU.Snapshot()
		if cur.state == thRunning {
			cur.state = curState
		}
	}
	next.state = thRunning
	m.currentThread = next.id
	m.idle = false
	m.CPU.Restore(next.ctx)
}

// pickReady returns the highest-priority ready thread — the lowest priority number.
func (m *Machine) pickReady() *thread {
	var best *thread
	for _, t := range m.threads {
		if t.state != thReady {
			continue
		}
		if best == nil || t.priority < best.priority ||
			(t.priority == best.priority && t.id < best.id) {
			best = t
		}
	}
	return best
}

// onThreadExit retires the thread that ran off its sentinel return address.
func (m *Machine) onThreadExit() {
	if t := m.threads[m.currentThread]; t != nil {
		t.state = thDead
		m.note("thread %d exited", t.id)
	}
	if !m.resume() {
		m.idle = true // maybe an interrupt or the IOP will make something runnable
	}
}

// Threads renders the thread table, the first thing to look at when a boot stops
// making progress.
func (m *Machine) Threads() string {
	if len(m.threads) == 0 {
		return "no threads\n"
	}
	s := "threads:\n"
	for id := uint32(1); id < m.nextThreadID; id++ {
		t := m.threads[id]
		if t == nil {
			continue
		}
		mark := " "
		if id == m.currentThread {
			mark = ">"
		}
		pc := uint32(t.ctx.PC)
		if id == m.currentThread {
			pc = uint32(m.CPU.PC)
		}
		s += sprintf("%s %2d  %-9s prio=%-3d pc=%-32s entry=%s\n",
			mark, t.id, t.state, t.priority, m.Sym(pc), m.Sym(t.entry))
	}
	return s
}

// exitThread retires the running thread from inside a syscall, rather than by running
// off its sentinel return address.
func (m *Machine) exitThread() {
	m.onThreadExit()
}

// referThreadStatus fills the caller's struct with the thread's state. The game polls
// it to find out whether a worker has finished.
func (m *Machine) referThreadStatus() {
	id, p := m.arg(0), m.arg(1)
	if id == 0 {
		id = m.currentThread
	}
	t := m.threads[id]
	if t == nil {
		m.setRet(0xFFFFFFFF)
		return
	}
	if p != 0 {
		m.Write32(p+0x00, uint32(eeThreadStatus(t.state)))
		m.Write32(p+0x04, t.entry)
		m.Write32(p+0x08, t.stack)
		m.Write32(p+0x0C, t.gp)
		m.Write32(p+0x10, t.priority)
		m.Write32(p+0x14, t.priority)
	}
	m.setRet(uint32(eeThreadStatus(t.state)))
}

// eeThreadStatus maps this model's thread states onto the kernel's own numbering,
// which is what the game compares against.
func eeThreadStatus(s threadState) int {
	switch s {
	case thRunning:
		return 0x01 // THS_RUN
	case thReady:
		return 0x02 // THS_READY
	case thWaitSema, thSleeping:
		return 0x04 // THS_WAIT
	case thDormant:
		return 0x10 // THS_DORMANT
	}
	return 0x00
}
