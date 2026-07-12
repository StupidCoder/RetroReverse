package n3ds

import (
	"testing"

	"retroreverse.com/tools/cpu/arm"
)

// newSyncMachine builds a bare Machine (no cartridge) with a flat memory region
// and one ready main thread — enough to exercise the scheduler and the
// synchronization primitives without booting an image.
func newSyncMachine() *Machine {
	m := &Machine{
		handles:  map[uint32]*kobject{},
		services: map[uint32]string{},
		ports:    map[uint32]string{},
	}
	m.nextHandle = 0x1000
	m.mapRegion("mem", 0, make([]byte, 0x20000)) // flat, for handle arrays + TLS
	m.CPU = arm.NewCPU(m)
	m.CPU.Arch = arm.V6K
	m.nextTLS = tlsBase + tlsSize
	main := &thread{id: 1, priority: 0x30, state: ready, tlsBase: 0x1000}
	main.ctx = *m.CPU
	m.nextThread = 2
	m.threads = []*thread{main}
	m.curThread = main
	return m
}

// addThread registers an extra ready thread at a priority for scheduling tests.
func (m *Machine) addThread(prio int32) *thread {
	t := &thread{id: m.nextThread, priority: prio, state: ready, tlsBase: 0x2000 + m.nextThread*0x1000}
	m.nextThread++
	m.threads = append(m.threads, t)
	return t
}

func TestPickRunnablePriorityAndRoundRobin(t *testing.T) {
	m := newSyncMachine()
	m.threads[0].priority = 20
	lo1 := m.addThread(30)
	lo2 := m.addThread(30)

	// Highest priority (lowest number) wins.
	if got := m.pickRunnable(); got != m.threads[0] {
		t.Fatalf("expected the priority-20 thread, got id %d", got.id)
	}
	// With the high-priority thread blocked, the two priority-30 threads share
	// time in round-robin order.
	m.threads[0].state = waiting
	first := m.pickRunnable()
	second := m.pickRunnable()
	if first == second {
		t.Fatalf("round-robin returned the same thread twice (id %d)", first.id)
	}
	if (first != lo1 && first != lo2) || (second != lo1 && second != lo2) {
		t.Fatalf("round-robin picked outside the priority-30 group")
	}
}

func TestEventWaitSignal(t *testing.T) {
	m := newSyncMachine()
	h := m.newHandle("event", false) // auto-reset, not signalled

	// Main waits on the unsignalled event → it blocks.
	m.CPU.R[0], m.CPU.R[2], m.CPU.R[3] = h, 1, 0 // handle, timeout ns = 1 in r2:r3 (the real ABI)
	m.svcWaitSync1(m.CPU)
	if m.curThread.state != waiting || !m.reschedule {
		t.Fatalf("wait on an unsignalled event did not block (state %v)", m.curThread.state)
	}

	// Signalling wakes it exactly once and delivers success; the auto event clears.
	m.svcSignalEvent2(h)
	if m.threads[0].state != ready {
		t.Fatalf("signal did not wake the waiter (state %v)", m.threads[0].state)
	}
	if m.threads[0].ctx.R[0] != resultSuccess {
		t.Fatalf("woken thread result = 0x%08X, want success", m.threads[0].ctx.R[0])
	}
	if m.handles[h].signal {
		t.Fatal("auto-reset event stayed signalled after waking a waiter")
	}
}

// svcSignalEvent2 is svcSignalEvent driven by a handle argument, for tests.
func (m *Machine) svcSignalEvent2(h uint32) {
	m.CPU.R[0] = h
	m.svcSignalEvent(m.CPU)
}

func TestMutexMutualExclusion(t *testing.T) {
	m := newSyncMachine()
	other := m.addThread(0x30)
	h := m.newHandle("mutex", false)

	// Main acquires the mutex (WaitSync1 on a free mutex).
	m.CPU.R[0], m.CPU.R[2], m.CPU.R[3] = h, 0, 0 // timeout 0: a try-wait must not block
	m.svcWaitSync1(m.CPU)
	if m.handles[h].mutexOwner != m.curThread.id {
		t.Fatalf("mutex not owned after acquire")
	}

	// `other` tries to acquire → it blocks.
	m.curThread = other
	m.CPU.R[0], m.CPU.R[2], m.CPU.R[3] = h, 1, 0
	m.svcWaitSync1(m.CPU)
	if other.state != waiting {
		t.Fatalf("second acquirer did not block on a held mutex")
	}

	// Main releases → `other` becomes the owner.
	m.curThread = m.threads[0]
	m.CPU.R[0] = h
	m.svcReleaseMutex(m.CPU)
	if other.state != ready {
		t.Fatalf("release did not wake the waiter")
	}
	if m.handles[h].mutexOwner != other.id {
		t.Fatalf("mutex owner = %d after handoff, want %d", m.handles[h].mutexOwner, other.id)
	}
}

func TestSemaphoreFanOut(t *testing.T) {
	m := newSyncMachine()
	a := m.addThread(0x30)
	b := m.addThread(0x30)
	h := m.newHandle("semaphore", false)
	m.handles[h].semCount = 0

	// Two threads wait on the empty semaphore.
	for _, tr := range []*thread{a, b} {
		m.curThread = tr
		m.CPU.R[0], m.CPU.R[2], m.CPU.R[3] = h, 1, 0
		m.svcWaitSync1(m.CPU)
		if tr.state != waiting {
			t.Fatalf("thread %d did not block on the empty semaphore", tr.id)
		}
	}
	// Release 2 permits → both wake.
	m.curThread = m.threads[0]
	m.CPU.R[1], m.CPU.R[2] = h, 2 // releaseSemaphore: r1=handle, r2=count
	m.svcReleaseSemaphore(m.CPU)
	if a.state != ready || b.state != ready {
		t.Fatalf("semaphore release did not wake both waiters (a=%v b=%v)", a.state, b.state)
	}
	if m.handles[h].semCount != 0 {
		t.Fatalf("semaphore count = %d after two wakes, want 0", m.handles[h].semCount)
	}
}

func TestArbitrateAddressWaitSignal(t *testing.T) {
	m := newSyncMachine()
	a := m.addThread(0x30)
	b := m.addThread(0x30)
	const addr = 0x1000
	m.WriteWord(addr, 0) // value 0

	// Two threads WAIT_IF_LESS_THAN(addr, 1): 0 < 1 → both park.
	for _, tr := range []*thread{a, b} {
		m.curThread = tr
		m.CPU.R[1], m.CPU.R[2], m.CPU.R[3] = addr, 1, 1 // addr, type WAIT_IF_LESS_THAN, value
		m.svcArbitrateAddress(m.CPU)
		if tr.state != waiting || tr.arbAddr != addr {
			t.Fatalf("thread %d did not park on the arbiter", tr.id)
		}
	}
	// SIGNAL one.
	m.curThread = m.threads[0]
	m.CPU.R[1], m.CPU.R[2], m.CPU.R[3] = addr, 0, 1 // addr, type SIGNAL, count 1
	m.svcArbitrateAddress(m.CPU)
	woke := 0
	if a.state == ready {
		woke++
	}
	if b.state == ready {
		woke++
	}
	if woke != 1 {
		t.Fatalf("SIGNAL count 1 woke %d threads, want 1", woke)
	}
}

func TestWaitSyncNAnyIndex(t *testing.T) {
	m := newSyncMachine()
	h0 := m.newHandle("event", false)
	h1 := m.newHandle("event", false)
	m.handles[h1].signal = true // only the second is signalled

	const arr = 0x1000
	m.WriteWord(arr, h0)
	m.WriteWord(arr+4, h1)
	// WaitSyncN(handles=arr, count=2, waitAll=false): returns index 1.
	m.CPU.R[0], m.CPU.R[1], m.CPU.R[2], m.CPU.R[3], m.CPU.R[4] = 0, arr, 2, 0, 0
	m.svcWaitSyncN(m.CPU)
	if m.CPU.R[0] != resultSuccess || m.CPU.R[1] != 1 {
		t.Fatalf("WaitSyncN any = (0x%08X, idx %d), want (success, 1)", m.CPU.R[0], m.CPU.R[1])
	}
}
