package n3ds

import "retroreverse.com/tools/cpu/arm"

// sync.go implements the real Horizon synchronization primitives over kobjects:
// a thread that waits on an unavailable object blocks (thread.go), and a signaller
// wakes waiters and hands each its result. This is what lets a game's threads
// hand off work and its init not spin — the previous always-signalled stubs let
// single-threaded init limp along but cannot support real threads.
//
// Result codes: success is 0; a zero-timeout poll that finds the object
// unavailable returns the Horizon "timeout" error so the caller's non-blocking
// fast paths behave.
const resultTimeout uint32 = 0x09401BFE

// objAvailable reports whether a wait on obj would succeed right now, without
// consuming anything.
func (m *Machine) objAvailable(obj *kobject) bool {
	switch obj.kind {
	case "mutex":
		return obj.mutexOwner == 0 || obj.mutexOwner == m.curThread.id
	case "semaphore":
		return obj.semCount > 0
	default: // event, timer, thread-exit, notification-semaphore, ports treated as ready
		return obj.signal
	}
}

// consume applies the side effect of a successful wait (auto-reset event clears,
// semaphore decrements, mutex takes ownership) for the given thread.
func (m *Machine) consume(obj *kobject, t *thread) {
	switch obj.kind {
	case "mutex":
		obj.mutexOwner = t.id
		obj.mutexDepth++
	case "semaphore":
		obj.semCount--
	case "event":
		if !obj.manualReset {
			obj.signal = false // auto-reset clears on the wake
		}
	}
}

// svcWaitSync1 blocks the current thread on one object until it is signalled.
// r0 = handle, r1:r2 = s64 timeout ns.
func (m *Machine) svcWaitSync1(c *arm.CPU) {
	handle := c.R[0]
	timeoutNs := int64(uint64(c.R[1]) | uint64(c.R[2])<<32)
	obj := m.handles[handle]
	if obj == nil {
		c.R[0] = resultTimeout // unknown handle: fail the wait rather than block forever
		return
	}
	if m.objAvailable(obj) {
		m.consume(obj, m.curThread)
		c.R[0] = resultSuccess
		return
	}
	if timeoutNs == 0 {
		c.R[0] = resultTimeout
		return
	}
	// Block: park the thread and register it as a waiter.
	m.curThread.waitOn = []uint32{handle}
	m.curThread.waitAll = false
	obj.waiters = append(obj.waiters, m.curThread.id)
	m.curThread.state = waiting
	m.reschedule = true
}

// svcWaitSyncN blocks on a set of objects. r0/r1 = handles ptr / result-index out
// slot (ABI: handles in r1, count in r2, waitAll in r3, timeout in r0:r4 — the
// libctru wrapper passes handlesPtr in r1). We read: r1=handlesPtr, r2=count,
// r3=waitAll, r0:r4 = timeout. On any-wait success r1 receives the index.
func (m *Machine) svcWaitSyncN(c *arm.CPU) {
	handlesPtr := c.R[1]
	count := int(c.R[2])
	waitAll := c.R[3] != 0
	timeoutNs := int64(uint64(c.R[0]) | uint64(c.R[4])<<32)

	handles := make([]uint32, count)
	for i := 0; i < count; i++ {
		handles[i] = m.ReadWord(handlesPtr + uint32(i)*4)
	}

	if waitAll {
		all := true
		for _, h := range handles {
			if obj := m.handles[h]; obj == nil || !m.objAvailable(obj) {
				all = false
				break
			}
		}
		if all {
			for _, h := range handles {
				m.consume(m.handles[h], m.curThread)
			}
			c.R[0] = resultSuccess
			return
		}
	} else {
		for i, h := range handles {
			if obj := m.handles[h]; obj != nil && m.objAvailable(obj) {
				m.consume(obj, m.curThread)
				c.R[0], c.R[1] = resultSuccess, uint32(i)
				return
			}
		}
	}
	if timeoutNs == 0 {
		c.R[0] = resultTimeout
		return
	}
	m.curThread.waitOn = handles
	m.curThread.waitAll = waitAll
	for _, h := range handles {
		if obj := m.handles[h]; obj != nil {
			obj.waiters = append(obj.waiters, m.curThread.id)
		}
	}
	m.curThread.state = waiting
	m.reschedule = true
}

// signalObject marks obj signalled and wakes waiters. For an auto-reset event
// exactly one waiter is released (and it re-clears); a manual-reset event wakes
// all and stays set; a semaphore releases up to n; a mutex release wakes one new
// owner. Returns whether a reschedule is warranted.
func (m *Machine) signalObject(obj *kobject) bool {
	resched := false
	// Try to satisfy waiters whose whole wait-condition is now met.
	i := 0
	for i < len(obj.waiters) {
		tid := obj.waiters[i]
		t := m.threadByID(tid)
		if t == nil || t.state != waiting {
			obj.waiters = append(obj.waiters[:i], obj.waiters[i+1:]...)
			continue
		}
		if m.tryComplete(t) {
			obj.waiters = append(obj.waiters[:i], obj.waiters[i+1:]...)
			if m.wake(t) {
				resched = true
			}
			// auto-reset / single-permit: stop after releasing one where the
			// object can no longer satisfy another waiter.
			if !m.objAvailable(obj) {
				break
			}
			continue
		}
		i++
	}
	return resched
}

// tryComplete checks whether all of thread t's wait conditions are now satisfied
// and, if so, consumes them and writes the result into t's context. Returns
// whether t was completed (and should be removed from waiter lists).
func (m *Machine) tryComplete(t *thread) bool {
	if t.waitAll {
		for _, h := range t.waitOn {
			if obj := m.handles[h]; obj == nil || !m.availableFor(obj, t) {
				return false
			}
		}
		for _, h := range t.waitOn {
			m.consumeFor(m.handles[h], t)
		}
		m.setResult(t, 0, resultSuccess)
		return true
	}
	for i, h := range t.waitOn {
		if obj := m.handles[h]; obj != nil && m.availableFor(obj, t) {
			m.consumeFor(obj, t)
			m.setResult(t, 0, resultSuccess)
			m.setResult(t, 1, uint32(i))
			return true
		}
	}
	return false
}

// availableFor / consumeFor are objAvailable / consume for a specific (possibly
// non-current) thread — mutex ownership is checked against t, not curThread.
func (m *Machine) availableFor(obj *kobject, t *thread) bool {
	switch obj.kind {
	case "mutex":
		return obj.mutexOwner == 0 || obj.mutexOwner == t.id
	case "semaphore":
		return obj.semCount > 0
	default:
		return obj.signal
	}
}

func (m *Machine) consumeFor(obj *kobject, t *thread) {
	switch obj.kind {
	case "mutex":
		obj.mutexOwner = t.id
		obj.mutexDepth++
	case "semaphore":
		obj.semCount--
	case "event":
		if !obj.manualReset {
			obj.signal = false
		}
	}
}

func (m *Machine) threadByID(id uint32) *thread {
	for _, t := range m.threads {
		if t.id == id {
			return t
		}
	}
	return nil
}

// --- signallers (svc entry points) ---

func (m *Machine) svcSignalEvent(c *arm.CPU) {
	obj := m.handles[c.R[0]]
	if obj != nil {
		obj.signal = true
		if m.signalObject(obj) {
			m.reschedule = true
		}
	}
	c.R[0] = resultSuccess
}

func (m *Machine) svcClearEvent(c *arm.CPU) {
	if obj := m.handles[c.R[0]]; obj != nil {
		obj.signal = false
	}
	c.R[0] = resultSuccess
}

func (m *Machine) svcReleaseMutex(c *arm.CPU) {
	obj := m.handles[c.R[0]]
	if obj != nil && obj.kind == "mutex" {
		if obj.mutexDepth > 0 {
			obj.mutexDepth--
		}
		if obj.mutexDepth == 0 {
			obj.mutexOwner = 0
			if m.signalObject(obj) {
				m.reschedule = true
			}
		}
	}
	c.R[0] = resultSuccess
}

func (m *Machine) svcReleaseSemaphore(c *arm.CPU) {
	// r0 = count-out slot, r1 = handle, r2 = release count.
	obj := m.handles[c.R[1]]
	if obj != nil && obj.kind == "semaphore" {
		prev := obj.semCount
		obj.semCount += int32(c.R[2])
		if m.signalObject(obj) {
			m.reschedule = true
		}
		c.R[0], c.R[1] = resultSuccess, uint32(prev)
		return
	}
	c.R[0] = resultSuccess
}

// signalThreadExit marks a thread's handle-object signalled and wakes anyone in
// WaitSync on it (join). Called from svcExitThread.
func (m *Machine) signalThreadExit(t *thread) {
	for _, o := range m.handles {
		if o.kind == "thread" && o.thread == t {
			o.signal = true
			if m.signalObject(o) {
				m.reschedule = true
			}
		}
	}
}

// svcArbitrateAddress is the address arbiter — the futex/condition-variable path
// libctru uses (LightLock contention, cond_wait). r1 = addr, r2 = type,
// r3 = value. Types: 0 SIGNAL (wake up to `value` parked threads, <0 = all),
// 1/3 WAIT_IF_LESS_THAN, 2/4 DECREMENT_AND_WAIT_IF_LESS_THAN.
func (m *Machine) svcArbitrateAddress(c *arm.CPU) {
	addr := c.R[1]
	typ := c.R[2]
	value := int32(c.R[3])
	switch typ {
	case 0: // SIGNAL
		m.arbSignal(addr, value)
		c.R[0] = resultSuccess
	case 1, 3: // WAIT_IF_LESS_THAN
		if int32(m.ReadWord(addr)) < value {
			m.arbPark(addr)
		} else {
			c.R[0] = resultSuccess
		}
	case 2, 4: // DECREMENT_AND_WAIT_IF_LESS_THAN
		cur := int32(m.ReadWord(addr))
		if cur < value {
			m.WriteWord(addr, uint32(cur-1))
			m.arbPark(addr)
		} else {
			c.R[0] = resultSuccess
		}
	default:
		c.R[0] = resultSuccess
	}
}

func (m *Machine) arbPark(addr uint32) {
	m.curThread.arbAddr = addr
	m.curThread.state = waiting
	m.reschedule = true // result (success) delivered on wake
}

func (m *Machine) arbSignal(addr uint32, count int32) {
	var n int32
	for _, t := range m.threads {
		if t.state == waiting && t.arbAddr == addr {
			if count >= 0 && n >= count {
				break
			}
			m.setResult(t, 0, resultSuccess)
			if m.wake(t) {
				m.reschedule = true
			}
			n++
		}
	}
}
