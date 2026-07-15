package xbox

// sched.go is the preemptive scheduler. It is deliberately minimal at Phase B's start
// — the boot runs on a single thread — but the machinery is here so that the first
// PsCreateSystemThreadEx and the first blocking wait have a scheduler to plug into,
// following the shape proven in tools/platform/n3ds (pickRunnable / switchTo / a
// per-instruction quantum) and tools/platform/psp (yieldCurrent + wake-on-signal).
//
// The quantum is charged in onStep (machine.go wires schedTick there). When a thread
// blocks (a wait that cannot be satisfied) it yields; the scheduler saves its context
// with the PC already past the trap and switches to the next ready thread. A later
// signal writes the wait result into the blocked thread's saved context and marks it
// ready, so it resumes correctly on its next switch-in — no PC rewinding, exactly the
// n3ds/psp discipline.

const schedQuantum = 4000 // instructions a thread runs before the scheduler reconsiders
const instrsPerMs = 2000  // nominal instruction-to-millisecond scale for the live counters

// schedTick charges the running thread's quantum and reschedules when it expires or a
// wake is pending. With one thread this is almost free; it becomes load-bearing once
// the title spawns worker threads.
func (m *Machine) schedTick() {
	// Keep the live kernel counters advancing. KeTickCount is a millisecond counter;
	// the tick advances once per instruction, so scale it to a plausible rate. Updating
	// on a coarse boundary keeps this cheap.
	if m.tick&0x3FF == 0 {
		if m.tickCountAddr != 0 {
			m.write32(m.tickCountAddr, uint32(m.tick/instrsPerMs))
		}
		if m.systemTimeAddr != 0 {
			t := m.tick / instrsPerMs * 10000 // 100 ns units
			m.write32(m.systemTimeAddr, uint32(t))
			m.write32(m.systemTimeAddr+4, uint32(t>>32))
		}
	}
	m.wakeDueSleepers()
	if m.reschedule {
		m.reschedule = false
		m.dispatch()
		return
	}
	m.quantumLeft--
	if m.quantumLeft <= 0 {
		m.quantumLeft = schedQuantum
		if len(m.threads) > 1 {
			m.dispatch()
		}
	}
}

// yieldCurrent parks the running thread in the given state and forces a reschedule at
// the next tick. The trap dispatcher has already advanced the PC past the kernel call,
// so the saved context resumes cleanly.
func (m *Machine) yieldCurrent(state threadState) {
	if m.current != nil {
		m.current.state = state
	}
	m.reschedule = true
}

// dispatch picks the highest-priority ready thread and switches to it. If none is
// ready and the current thread is not runnable, the machine is deadlocked as far as
// the ready set goes — reported by the boot driver, not silently spun.
func (m *Machine) dispatch() {
	next := m.pickRunnable()
	if next == nil {
		if m.current != nil && m.current.state == tsRunning {
			return // the current thread is still runnable; keep going
		}
		m.CPU.Halt("scheduler: no runnable thread (deadlock); %d threads", len(m.threads))
		m.Halted, m.HaltReason = true, m.CPU.HaltReason
		return
	}
	m.switchTo(next)
}

// pickRunnable returns the highest-priority ready thread, rotating among equals.
func (m *Machine) pickRunnable() *thread {
	var best *thread
	n := len(m.threads)
	if n == 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		t := m.threads[(m.rrCursor+i)%n]
		if t.state == tsReady && (best == nil || t.priority > best.priority) {
			best = t
		}
	}
	if best != nil {
		m.rrCursor = (m.rrCursor + 1) % n
	}
	return best
}

// switchTo makes t the running thread, saving the outgoing context and loading t's.
func (m *Machine) switchTo(t *thread) {
	if m.current == t {
		t.state = tsRunning
		return
	}
	if m.current != nil {
		m.current.ctx = *m.CPU // save outgoing (bus/hook pointers are copied verbatim)
		if m.current.state == tsRunning {
			m.current.state = tsReady
		}
	}
	*m.CPU = t.ctx
	m.current = t
	t.state = tsRunning
	m.quantumLeft = schedQuantum
	m.write32(kpcrAddr+kpcrPrcbData+prcbCurrentThread, t.kthread)
}

// wakeDueSleepers readies every sleeping thread whose wake tick has passed, and expires
// timed waits (doWaitTimed): a waiting thread with a nonzero wakeTick whose deadline has
// passed resumes with the STATUS_TIMEOUT its parked context already carries. Suspended
// threads (create-suspended, NtSuspendThread) are tsWaiting with wakeTick 0 — untouched.
func (m *Machine) wakeDueSleepers() {
	for _, t := range m.threads {
		switch {
		case t.state == tsSleeping && t.wakeTick <= m.tick:
			t.state = tsReady
		case t.state == tsWaiting && t.wakeTick != 0 && t.wakeTick <= m.tick:
			t.state = tsReady
			t.waitObjs = nil
			t.wakeTick = 0
		}
	}
}

// aliveThreads counts threads not yet dead.
func (m *Machine) aliveThreads() int {
	n := 0
	for _, t := range m.threads {
		if t.state != tsDead {
			n++
		}
	}
	return n
}
