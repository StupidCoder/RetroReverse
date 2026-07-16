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

// systemTime100ns is the current synthetic system time in 100-ns units — the same clock
// KeSystemTime advances and KeQuerySystemTime reports, so the data export and the call
// agree. It is a relative uptime, not a wall-clock date (no console RTC is modelled);
// what the title needs it for is monotonic nonce/timestamp material, which this supplies.
func (m *Machine) systemTime100ns() uint64 {
	return m.tick / instrsPerMs * 10000
}

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
			t := m.systemTime100ns()
			m.write32(m.systemTimeAddr, uint32(t))
			m.write32(m.systemTimeAddr+4, uint32(t>>32))
		}
		m.apuTick()    // the GP DSP's command-mailbox poll (apu.go)
		m.ioTick()     // due asynchronous read completions (kernel_file.go)
		m.vblankTick() // the 60 Hz PCRTC vertical blank (interrupt.go)
	}
	if m.isrActive {
		return // an ISR frame runs to completion: no wakes, no thread switches
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

// dispatch picks the highest-priority ready thread and switches to it. With nothing
// runnable it idles the machine forward to the next future wake source (a sleeper's
// deadline, a timed wait, a pending I/O completion); only when no such source exists
// is the machine genuinely deadlocked — reported by the boot driver, not silently
// spun.
func (m *Machine) dispatch() {
	next := m.pickRunnable()
	if next == nil && m.current != nil && m.current.state == tsRunning && m.current.suspendCount == 0 {
		return // the current thread is still runnable; keep going
	}
	for next == nil {
		if !m.idleAdvance() {
			m.CPU.Halt("scheduler: no runnable thread (deadlock); %d threads", len(m.threads))
			m.Halted, m.HaltReason = true, m.CPU.HaltReason
			return
		}
		next = m.pickRunnable()
	}
	m.switchTo(next)
}

// idleAdvance jumps the clock to the earliest future wake source and delivers it:
// a sleeping thread's deadline, a timed wait's expiry, or a pending asynchronous
// I/O completion (whose IOSB write can satisfy a poller once it runs again, and
// whose event signal can complete a wait). Returns false when no source exists.
func (m *Machine) idleAdvance() bool {
	var min uint64
	for _, t := range m.threads {
		if (t.state == tsSleeping || (t.state == tsWaiting && t.wakeTick != 0)) &&
			(min == 0 || t.wakeTick < min) {
			min = t.wakeTick
		}
	}
	for _, p := range m.pendingIO {
		if min == 0 || p.Due < min {
			min = p.Due
		}
	}
	if min == 0 {
		return false
	}
	if min > m.tick {
		m.tick = min
	}
	m.wakeDueSleepers()
	m.ioTick()
	return true
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
		if t.runnable() && (best == nil || t.priority > best.priority) {
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
// passed resumes with the STATUS_TIMEOUT its parked context already carries. Suspension
// (thread.suspendCount) is orthogonal: a woken-but-suspended thread becomes tsReady here
// and simply stays unpicked until its count drops to zero.
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
