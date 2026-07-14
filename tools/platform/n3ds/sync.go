package n3ds

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

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
// r0 = handle, **r2:r3** = s64 timeout ns — the 64-bit value is register-pair
// aligned, so it skips r1 (every one of the game's 33 call sites sets r2/r3 and
// leaves r1 holding whatever was there before). Reading it from r1:r2 mostly
// "worked", because an infinite wait (-1) still came out non-zero — but a
// *try-wait* (timeout 0, which the game writes as MOV r2,#0; MOV r3,r2) then
// read garbage from r1, looked like a long timeout, and blocked forever instead
// of returning TIMEOUT at once.
func (m *Machine) svcWaitSync1(c *arm.CPU) {
	handle := c.R[0]
	timeoutNs := int64(uint64(c.R[2]) | uint64(c.R[3])<<32)
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
	m.armWaitDeadline(timeoutNs)
	obj.waiters = append(obj.waiters, m.curThread.id)
	m.curThread.state = waiting
	m.reschedule = true
}

// armWaitDeadline bounds the running thread's wait when the caller asked for a
// finite timeout. A negative timeout is Horizon's "wait for ever" (the callers
// write -1); only a positive one is a deadline. thread.go expires it.
func (m *Machine) armWaitDeadline(ns int64) {
	if ns > 0 {
		m.curThread.waitDeadline = m.tick + nsToTick(ns)
		return
	}
	m.curThread.waitDeadline = 0
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
	m.armWaitDeadline(timeoutNs)
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
// libctru uses (LightLock contention, cond_wait). r1 = addr, r2 = type, r3 = value.
//
// The type numbering, and how it was pinned. The obvious reading — that the types come
// in two pairs, wait and decrement-and-wait, each with a timeout variant "next to" it —
// is wrong, and it was briefly coded here before the machine was asked. The order is:
//
//	0 SIGNAL                                  wake up to `value` parked threads (<0 = all)
//	1 WAIT_IF_LESS_THAN                       park iff *addr < value
//	2 DECREMENT_AND_WAIT_IF_LESS_THAN         the same, decrementing the word first
//	3 WAIT_IF_LESS_THAN_TIMEOUT               1, bounded by an s64 ns timeout
//	4 DECREMENT_AND_WAIT_IF_LESS_THAN_TIMEOUT 2, bounded by an s64 ns timeout
//
// The timeout is what the HIGH numbers add; the decrement is what the EVEN ones add.
// The game settles it: at a type-2 call the register file holds no timeout at all (r4
// is a copy of the address, r5 is 1 — leftovers, not an s64 in r4:r5, which is where a
// wrapper that had one would have put it). A caller that meant "wait with a timeout"
// would have loaded one. So 2 is the untimed decrement-and-wait, and this is what the
// code always did.
func (m *Machine) svcArbitrateAddress(c *arm.CPU) {
	addr := c.R[1]
	typ := c.R[2]
	value := int32(c.R[3])
	switch typ {
	case 0: // SIGNAL
		m.arbSignal(addr, value)
		c.R[0] = resultSuccess
	case 1, 3: // WAIT_IF_LESS_THAN (3: with timeout)
		if int32(m.ReadWord(addr)) < value {
			m.arbPark(addr, typ == 3)
		} else {
			c.R[0] = resultSuccess
		}
	case 2, 4: // DECREMENT_AND_WAIT_IF_LESS_THAN (4: with timeout)
		cur := int32(m.ReadWord(addr))
		if cur <= value {
			m.WriteWord(addr, uint32(cur-1))
			m.arbPark(addr, typ == 4)
		} else {
			c.R[0] = resultSuccess
		}
	default:
		c.R[0] = resultSuccess
	}
}

// arbPark blocks the running thread on an arbiter address until someone signals it.
//
// The timed variants (3, 4) are not modelled: no title has been seen to use them, and a
// deadline written against no observation is a guess. But an unmodelled timeout degrades
// into a thread that waits for ever — which is how this platform's worst bugs present —
// so it says so out loud rather than hanging quietly. If this ever prints, read the
// timeout out of the caller's registers and implement it.
func (m *Machine) arbPark(addr uint32, timed bool) {
	if timed && !m.arbTimedWarned[addr] {
		if m.arbTimedWarned == nil {
			m.arbTimedWarned = map[uint32]bool{}
		}
		m.arbTimedWarned[addr] = true
		fmt.Printf("n3ds: thread %d parked on a TIMED arbiter wait (arb@0x%08X) — the timeout is not modelled, "+
			"so this thread will only wake if something signals the address\n", m.curThread.id, addr)
	}
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
