package psp

// sched.go is a cooperative thread scheduler for the kernel HLE. The PSP is
// preemptive, but games are structured so that a thread runs until it blocks
// (sleep/wait/delay) or exits; modelling that cooperatively is enough to carry the
// boot through its thread hand-offs without a timer-driven scheduler.
//
// Threads are kobjects with a saved register context. sceKernelStartThread makes a
// thread runnable and lets the caller continue; when the caller sleeps or exits, the
// scheduler saves its context and switches to the highest-priority ready thread. The
// initial module-start runs as an anonymous context (current == nil) until it yields.

import "retroreverse.com/tools/cpu/allegrex"

// threadExitAddr is the sentinel return address a thread's $ra is seeded with; when
// the run loop sees the PC reach it, the current thread has returned and the
// scheduler switches. It is an unmapped address the run loop traps explicitly.
const threadExitAddr = 0x0F000000

type threadState int

const (
	thDormant threadState = iota // created, not started
	thReady                      // runnable
	thRunning                    // currently executing
	thWaiting                    // blocked (sleep/wait)
)

// startThread builds a thread's initial register context and marks it ready. The gp
// is inherited from the running context (all threads share the module gp).
//
// The 256 bytes at the top of the stack are the thread's context area, pointed to
// by $k0: the kernel stores the thread uid at +0xC0 and the stack base at +0xC8
// (and the uid again at the stack base). Sony's libc walks $k0 to find its
// per-thread reentrancy data, so this is load-bearing.
func (m *Machine) startThread(uid uint32, o *kobject, argLen, argPtr uint32) {
	k0 := m.threadK0(uid, o.stackTop)
	ctx := allegrex.CPUState{}
	ctx.VfpuCtrl[0], ctx.VfpuCtrl[1] = 0xE4, 0xE4 // vpfxs/vpfxt at their identity
	ctx.R[26] = k0                                // $k0
	ctx.R[28] = m.CPU.Reg(28)                     // $gp
	ctx.R[29] = k0                                // $sp: below the context area
	ctx.R[30] = k0                                // $fp
	ctx.R[31] = threadExitAddr
	ctx.R[4] = argLen // $a0
	ctx.R[5] = argPtr // $a1
	ctx.Out = ctx.R
	ctx.PC = o.entry
	ctx.NextPC = o.entry + 4
	ctx.Steps = m.CPU.Steps
	o.ctx = ctx
	o.tstate = thReady
}

// threadK0 lays out a thread's 256-byte $k0 context area at the top of its stack
// and returns its address.
func (m *Machine) threadK0(uid, stackTop uint32) uint32 {
	k0 := (stackTop - 0x100) &^ 0xF
	for i := uint32(0); i < 0x100; i++ {
		m.Write(k0+i, 0)
	}
	m.write32(k0+0xC0, uid)
	stackBase := k0 // without the block's true base, point at the area itself
	if o := m.handles[uid]; o != nil && o.kind == "thread" {
		stackBase = o.addr
	}
	m.write32(k0+0xC8, stackBase)
	m.write32(k0+0xF8, 0xFFFFFFFF)
	m.write32(k0+0xFC, 0xFFFFFFFF)
	m.write32(stackBase, uid)
	return k0
}

// schedule saves the running context and switches to the highest-priority ready
// thread (lower priority number = higher priority on the PSP). With nothing ready
// but timed waits pending, it idles to the next VBlank (waking timed waiters and
// running the VBlank handlers, which may ready more threads). It returns false
// and sets doneReason only when nothing can ever run again.
func (m *Machine) schedule(currentBecomes threadState) bool {
	if m.current != nil {
		m.current.ctx = m.CPU.SaveState()
		if m.current.tstate == thRunning {
			m.current.tstate = currentBecomes
		}
	}
	for tries := 0; tries < 600; tries++ {
		var best *kobject
		for _, o := range m.handles {
			if o.kind == "thread" && o.tstate == thReady {
				if best == nil || o.priority < best.priority {
					best = o
				}
			}
		}
		if best != nil {
			m.current = best
			best.tstate = thRunning
			m.CPU.LoadState(best.ctx)
			return true
		}
		timed := false
		for _, o := range m.handles {
			if o.kind == "thread" && o.tstate == thWaiting && o.wakeVblank != 0 {
				timed = true
				break
			}
		}
		if !timed {
			break
		}
		m.deliverVBlank()
	}
	if m.doneReason == "" {
		m.doneReason = "no runnable threads"
	}
	return false
}

// yieldCurrent is called when the running context blocks or exits: the current
// context takes newState and the scheduler picks the next thread. If none is ready,
// the machine is marked done.
func (m *Machine) yieldCurrent(newState threadState) {
	if !m.schedule(newState) {
		m.Halted = true
		m.HaltReason = m.doneReason
	}
}

// currentThreadID is the running thread's handle; the anonymous module-start
// context gets a fixed pseudo-handle clear of real ones.
func (m *Machine) currentThreadID() uint32 {
	if m.current == nil {
		return 0x1000
	}
	for h, o := range m.handles {
		if o == m.current {
			return h
		}
	}
	return 0x1000
}

// onThreadExit handles a thread returning to threadExitAddr.
func (m *Machine) onThreadExit() {
	if m.current != nil {
		m.current.tstate = thDormant
	}
	// The exiting context does not need its registers saved.
	saved := m.current
	m.current = nil
	if !m.schedule(thDormant) {
		m.Halted = true
		if m.doneReason == "" {
			m.doneReason = "all threads exited"
		}
		m.HaltReason = m.doneReason
	}
	_ = saved
}
