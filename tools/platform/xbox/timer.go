package xbox

import "retroreverse.com/tools/cpu/x86"

// timer.go is the kernel's timer queue: KeSetTimer's armed KTIMERs, the deadline that
// expires them, and the DPC each one runs when it does.
//
// It exists because of a comment that was true when it was written. KeSetTimer used to
// record its arguments and fire nothing, and said so —
//
//	// Record the association; the DPC is not fired here (nothing yet waits on it).
//
// — which was an honest description of a machine whose only timer callers were an NVNET
// link poll and a sound-library heartbeat that the title also drove from its own frame
// loop. Nothing waited on it, so nothing missed it.
//
// The pad is what came to wait on it. XAPI's hub driver does not reset a port the moment
// it sees a device: a physical connector bounces, so it DEBOUNCES, and the whole of that
// debounce is a timer —
//
//	00241C01  MOV BYTE [ESI], $FD          tag the deferred work item
//	00241C04  PUSH $0057CE94               a KDPC
//	00241C0E  OR ECX, $FFFFFFFF            \  DueTime = 0xFFFFFFFF_FFF0BDC0
//	00241C12  MOV EAX, $FFF0BDC0            > = -1,000,000 x 100ns
//	00241C17  PUSH EAX                     /  = RELATIVE 100 ms
//	00241C18  MOV BYTE [$0057CE60], $01    "delayed work pending"
//	00241C35  PUSH $0057CEB0               a KTIMER
//	00241C40  CALL [$002483B4]             -> ordinal 149, KeSetTimer
//
// — so with a record-only stub the driver saw our pad, believed it, filed the reset for
// 100 ms later, and waited for ever. The port was never reset, the device never
// enumerated, and the machine sat in a state indistinguishable from having no pad at all.
// A differential run is what pinned it: without a pad KeSetTimer is called once, with a
// pad twice, and the second call is the debounce.
//
// The other branch of that same function (0x241BCA, taken when [0x57CE60] is set) is the
// same 100 ms by hand — KeQuerySystemTime plus 0xF4240, onto a deadline-sorted list at
// 0x57CEDC. It is the fallback when a timer is already in flight, and nothing read that
// list either. Both routes were dead; this file is why the first one is not.
//
// THE STUBS AROUND IT WERE ONLY CORRECT BECAUSE THIS DID NOT EXIST. KeCancelTimer
// answered "not pending" and KeRemoveQueueDpc answered "not queued", and both were true
// of a machine with no timer queue — which is exactly the kind of correctness that stops
// being correct without anyone editing it. They are wired to the real queue below.

// b2u renders a Go bool as the BOOLEAN a kernel export returns.
func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// ktimer is one armed KTIMER: the guest object, the DPC to run when it expires, its
// deadline on the same clock KeQuerySystemTime reports, and its period if it repeats.
//
// Due is an ABSOLUTE deadline in 100-ns units, computed once at arm time, rather than a
// countdown decremented per tick. That is the apuHandshake discipline again: a counter
// that advances per call drifts the moment a savestate skips a call, while a deadline
// against a clock derived from the machine tick means a restored state resumes the same
// timer with the same expiry.
type ktimer struct {
	Timer  uint32 // KTIMER guest address — the identity a re-arm or a cancel names
	Dpc    uint32 // KDPC to queue on expiry; 0 for a timer that only signals
	Due    uint64 // deadline, systemTime100ns units
	Period uint32 // milliseconds; 0 = one-shot
}

// armTimer arms (or re-arms) a KTIMER, reporting whether it was already in the queue —
// which is what KeSetTimer returns.
//
// The due time is the kernel's own convention and not ours to choose: NEGATIVE is an
// interval from now, positive is an absolute time on the system clock. XAPI's debounce
// passes -1,000,000, and reading that as an absolute deadline would arm a timer for a
// moment 3,000 years in the past — which would fire instantly and skip the very
// debounce the driver asked for. It would even look like it worked.
func (m *Machine) armTimer(tm, dpc uint32, dueLo, dueHi uint32, period uint32) bool {
	if tm == 0 {
		return false
	}
	due := uint64(dueHi)<<32 | uint64(dueLo)
	var deadline uint64
	if int64(due) < 0 {
		deadline = m.systemTime100ns() + uint64(-int64(due))
	} else {
		deadline = due
	}
	was := m.cancelTimer(tm) // re-arming replaces, it does not stack
	m.write32(tm+dhSignalState, 0)
	m.timers = append(m.timers, ktimer{Timer: tm, Dpc: dpc, Due: deadline, Period: period})
	return was
}

// cancelTimer removes a KTIMER from the queue, reporting whether it was there.
func (m *Machine) cancelTimer(tm uint32) bool {
	for i := range m.timers {
		if m.timers[i].Timer == tm {
			m.timers = append(m.timers[:i], m.timers[i+1:]...)
			return true
		}
	}
	return false
}

// queueDPC appends a DPC unless it is already queued, the way KeInsertQueueDpc does.
func (m *Machine) queueDPC(dpc, arg1, arg2 uint32) bool {
	if dpc == 0 {
		return false
	}
	for _, d := range m.dpcQueue {
		if d.Dpc == dpc {
			return false
		}
	}
	m.dpcQueue = append(m.dpcQueue, dpcEntry{Dpc: dpc, Arg1: arg1, Arg2: arg2})
	return true
}

// timerTick expires every timer whose deadline has passed: it signals the KTIMER, wakes
// anything waiting on it, and queues its DPC. Called from schedTick's coarse block.
//
// It expires by DEADLINE rather than by "one per tick", so a coarse block that runs every
// 1024 instructions against a 100 ms timer cannot miss it, and a savestate restored with
// the clock well past a deadline expires it immediately rather than waiting all over
// again.
func (m *Machine) timerTick() {
	if len(m.timers) == 0 {
		return
	}
	now := m.systemTime100ns()
	for i := 0; i < len(m.timers); {
		t := m.timers[i]
		if now < t.Due {
			i++
			continue
		}
		// Signalling and waking use the same path KeSetEvent does, because a KTIMER is a
		// dispatcher object like any other and a thread may be blocked on it.
		if o := m.guestObjAt(t.Timer); o != nil {
			o.signaled = true
			m.writeSignal(o.addr, true)
			o.noteSignal("timer", 0, m.tick)
			m.wakeWaiters(o.addr)
		}
		m.queueDPC(t.Dpc, 0, 0)
		if t.Period != 0 {
			// A periodic timer re-arms from NOW, not from its old deadline: catching up
			// on missed periods would queue the same DPC repeatedly (queueDPC would drop
			// them anyway) and is not what a driver polling at an interval wants.
			m.timers[i].Due = now + uint64(t.Period)*10000
			i++
		} else {
			m.timers = append(m.timers[:i], m.timers[i+1:]...)
		}
	}
	m.deliverDPC()
}

// deliverDPC runs a queued DPC when no interrupt frame is already running.
//
// It exists because until now every DPC arrived behind an ISR — KeInsertQueueDpc was only
// ever called FROM one, and isrReturn drains the queue as that frame unwinds. A timer's
// DPC has no ISR in front of it, so without an entry point of its own it would sit in the
// queue until the next vblank happened to deliver one and drain it behind itself. That
// would mostly work, which is the problem: it would make every timer's latency depend on
// an unrelated device's interrupt, and a machine with the display asleep would never run
// a timer at all.
//
// The gates are deliverPending's, for the same reasons, and the frame is built the way
// isrReturn builds a DPC frame — so isrReturn drains any remaining DPCs and restores the
// interrupted context with no further help.
func (m *Machine) deliverDPC() {
	if m.isrActive || m.CPU == nil || m.CPU.Halted || !m.CPU.IF {
		return
	}
	if m.Read(kpcrAddr+kpcrIrql) != 0 {
		return
	}
	if len(m.dpcQueue) == 0 {
		return
	}
	d := m.dpcQueue[0]
	routine := m.read32(d.Dpc + 0x0C)
	if routine == 0 {
		m.dpcQueue = m.dpcQueue[1:] // a DPC with no routine is not a frame; drop it
		return
	}
	m.dpcQueue = m.dpcQueue[1:]
	m.isrSaved = *m.CPU
	m.isrActive = true
	c := m.CPU
	push := func(v uint32) {
		c.Regs[x86.SP] -= 4
		m.write32(c.Regs[x86.SP], v)
	}
	// DeferredRoutine(Dpc, DeferredContext, SystemArgument1, SystemArgument2)
	push(d.Arg2)
	push(d.Arg1)
	push(m.read32(d.Dpc + 0x10)) // DeferredContext
	push(d.Dpc)
	push(isrExitAddr)
	c.IP = routine
}
