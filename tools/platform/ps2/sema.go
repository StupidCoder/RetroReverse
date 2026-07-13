package ps2

// sema.go is the kernel's counting semaphores.
//
// They are how the game's threads wait on each other and on the hardware: a thread
// that has asked the IOP for a file blocks on a semaphore, and the interrupt that
// says the transfer finished signals it. That makes them the place a boot most easily
// deadlocks — a semaphore nobody signals is a thread that never runs again — so the
// count is modelled honestly rather than made permissive.

// The struct CreateSema is handed a pointer to:
//
//	0x00 count        the *live* count — output, not input
//	0x04 maxCount
//	0x08 initCount    what the semaphore starts at
//	0x0C waitThreads
//	0x10 attr
//	0x14 option
//
// The initial count comes from initCount, not from count. Jak creates two semaphores
// at boot: one with initCount 0 (a signal, nobody may pass until someone posts) and one
// with initCount 1 (a mutex, free). Read the wrong field and the mutex is born locked,
// so the first thread to take it blocks forever — and the boot deadlocks with two
// threads waiting on semaphores and no clue why.
type sema struct {
	id       uint32
	count    int32
	maxCount int32
	waiting  []uint32 // thread ids, in the order they blocked
}

func (m *Machine) createSema() {
	p := m.arg(0)
	s := &sema{
		id:       m.nextSemaID,
		count:    int32(m.Read32(p + 0x08)), // initCount
		maxCount: int32(m.Read32(p + 0x04)),
	}
	m.nextSemaID++
	m.semas[s.id] = s
	m.setRet(s.id)
}

// waitSema takes a count, blocking the calling thread when there is none to take.
func (m *Machine) waitSema(id uint32) {
	s := m.semas[id]
	if s == nil {
		m.setRet(0xFFFFFFFF)
		return
	}
	m.setRet(id)
	if s.count > 0 {
		s.count--
		return
	}
	// No count: the caller blocks. It goes on the semaphore's queue and off the
	// scheduler's, and only a SignalSema will put it back.
	s.waiting = append(s.waiting, m.currentThread)
	m.switchAway(thWaitSema)
}

// signalSema returns a count, releasing the longest-waiting thread if there is one.
func (m *Machine) signalSema(id uint32) {
	s := m.semas[id]
	if s == nil {
		m.setRet(0xFFFFFFFF)
		return
	}
	m.setRet(id)
	if len(s.waiting) > 0 {
		// A waiter takes the count directly; it is never added to the counter and then
		// removed, because a second thread could take it in between.
		tid := s.waiting[0]
		s.waiting = s.waiting[1:]
		if t := m.threads[tid]; t != nil && t.state == thWaitSema {
			t.state = thReady
		}
		return
	}
	if s.count < s.maxCount || s.maxCount == 0 {
		s.count++
	}
}

// pollSema takes a count if one is free, and reports failure rather than blocking.
func (m *Machine) pollSema(id uint32) {
	s := m.semas[id]
	if s == nil || s.count <= 0 {
		m.setRet(0xFFFFFFFF)
		return
	}
	s.count--
	m.setRet(id)
}
