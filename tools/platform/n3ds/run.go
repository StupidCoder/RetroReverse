package n3ds

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// Run executes up to budget instructions across all threads and returns the
// number actually run. It is the cooperative scheduler: each iteration picks the
// highest-priority ready thread (thread.go), runs it a quantum, then reconsiders
// — a blocked svc breaks the quantum early via m.reschedule. When every thread is
// blocked it advances idle time (waking sleepers) or, if truly deadlocked, halts.
// It stops early when the core halts. Pacing is by total instruction count, which
// keeps a resumed savestate deterministic (the N64/DOS discipline).
//
// maxIdleFrames bounds how many successive timed wakeups that make no thread
// runnable (VBlanks, and DSP audio frames — which recur ~3.4× per VBlank
// whether or not anything listens) the run tolerates before calling it a
// genuine deadlock rather than an idle wait between frames. 40 mixed wakeups
// ≈ the 8 video frames the pre-DSP bound allowed.
const maxIdleFrames = 40

func (m *Machine) Run(budget int) int {
	n := 0
	idleFrames := 0
	for n < budget {
		if m.CPU.Halted {
			break
		}
		if m.vblankDue() {
			m.deliverVBlank()
		}
		if m.dspDue() {
			m.dspTick() // the audio-frame clock (dsp.go), independent of the VBlank
		}
		m.processGXQueue() // drain any GPU commands the game posted; completions are deadline-paced
		t := m.pickRunnable()
		if t == nil {
			// The nearest machine event due before the next VBlank — a pending GX
			// completion or the DSP's audio frame — is the nearest wake source:
			// jump the clock to it and deliver. A GX completion is one-shot (real
			// progress, resets the idle bound); the DSP frame recurs whether or
			// not anyone listens, so it counts toward the idle bound like a
			// VBlank does — else a sound-less title would livelock here.
			gxDl, gxOK := m.gxDeadline()
			dl, ok, isGX := gxDl, gxOK, gxOK
			if d, o := m.dspDeadline(); o && (!ok || d < dl) {
				dl, ok, isGX = d, true, false
			}
			bound := m.nextFrameInstr
			if m.gspEvent == 0 {
				bound = ^uint64(0) // no graphics heartbeat yet: the DSP paces itself
			}
			if ok && dl < bound && idleFrames < maxIdleFrames {
				if m.instrs < dl {
					m.instrs = dl
				}
				m.pumpGX()
				if m.dspDue() {
					m.dspTick()
				}
				if isGX {
					idleFrames = 0
				} else {
					idleFrames++
				}
				continue
			}
			if m.advanceIdle() {
				continue
			}
			// Every thread is blocked with no timed wake pending. If the graphics
			// heartbeat is live, the game is waiting for the next VBlank — jump to
			// the frame boundary and deliver it rather than declaring a deadlock.
			// Bound it: if several successive VBlanks wake nothing, the game is
			// genuinely stuck (not merely idling between frames) — report it.
			if m.gspEvent != 0 && idleFrames < maxIdleFrames {
				idleFrames++
				m.instrs = m.nextFrameInstr
				m.deliverVBlank()
				continue
			}
			m.dumpThreads()
			m.CPU.Halt("all threads blocked (deadlock): %d live, none runnable, after %d instructions",
				m.aliveThreads(), m.CPU.Instrs)
			break
		}
		idleFrames = 0 // a thread is runnable — real progress, not an idle frame
		m.switchTo(t)

		for q := 0; q < quantum && n < budget; q++ {
			if m.CPU.Halted {
				break
			}
			pc := m.CPU.R[15]
			if pc == threadExitSentinel { // a thread function returned to LR
				m.svcExitThread(m.CPU)
				break
			}
			if m.bps[pc] {
				fmt.Printf("breakpoint [t%d] at 0x%08X r0=%08X r1=%08X r4=%08X r5=%08X lr=%08X after %d\n",
					m.curThread.id, pc, m.CPU.R[0], m.CPU.R[1], m.CPU.R[4], m.CPU.R[5], m.CPU.R[14], n)
				m.stopped = true
				break
			}
			if m.tracefroms[pc] {
				m.Trace = true
				m.traceN = 0
			}
			if m.logpcs[pc] {
				sp := m.CPU.R[13]
				fmt.Printf("logpc [t%d] 0x%08X r0=%08X r1=%08X r2=%08X r3=%08X r4=%08X lr=%08X sp=[%08X %08X %08X] instr=%d\n",
					m.curThread.id, pc, m.CPU.R[0], m.CPU.R[1], m.CPU.R[2], m.CPU.R[3], m.CPU.R[4], m.CPU.R[14],
					m.ReadWord(sp), m.ReadWord(sp+4), m.ReadWord(sp+8), m.CPU.Instrs)
			}
			if m.Trace && m.traceN < m.traceMax {
				m.traceOne(pc)
				m.traceN++
			}
			m.checkWatches(pc)
			m.instrs++  // machine-monotonic (CPU.Instrs is per-thread; see machine.go)
			m.tick += 2 // GetSystemTick advances; nominal, like the PSX timer
			m.reschedule = false
			m.CPU.Step()
			n++
			if m.reschedule || t.state != running {
				break
			}
		}
		if m.stopped {
			break
		}
		// Save the thread's context back; if it is still running, it merely used
		// up its quantum — return it to the ready pool.
		t.ctx = *m.CPU
		if t.state == running {
			t.state = ready
		}
	}
	return n
}

// SetTrace turns on instruction tracing to stdout, up to max instructions.
func (m *Machine) SetTrace(on bool, max int) {
	m.Trace = on
	m.traceMax = max
}

// AddBreakpoint registers a PC breakpoint.
func (m *Machine) AddBreakpoint(addr uint32) { m.bps[addr] = true }

// AddTraceFrom registers a PC that, when first reached, switches on instruction
// tracing for the next tracen instructions — a way to trace a specific routine
// deep in a long boot without drowning in the millions of instructions before it.
func (m *Machine) AddTraceFrom(addr uint32) {
	if m.tracefroms == nil {
		m.tracefroms = map[uint32]bool{}
	}
	m.tracefroms[addr] = true
}

// AddLogPC registers a PC that logs register context each time it executes and
// continues — the non-halting counterpart to a breakpoint, for watching how often
// and with what arguments a routine runs during a long boot.
func (m *Machine) AddLogPC(addr uint32) {
	if m.logpcs == nil {
		m.logpcs = map[uint32]bool{}
	}
	m.logpcs[addr] = true
}

// AddWatch registers a memory watch over [addr, addr+length).
func (m *Machine) AddWatch(addr, length uint32) {
	if length == 0 {
		length = 1
	}
	m.watches = append(m.watches, watch{addr: addr, len: length})
}

func (m *Machine) traceOne(pc uint32) {
	var buf [4]byte
	for i := uint32(0); i < 4; i++ {
		buf[i] = m.Read(pc + i)
	}
	in := arm.DecodeVariant(buf[:], pc, m.CPU.Thumb, arm.V6K)
	fmt.Printf("[t%d] %08X: %-22s  r0=%08X r1=%08X r2=%08X r3=%08X sp=%08X lr=%08X\n",
		m.curThread.id, pc, in.Text, m.CPU.R[0], m.CPU.R[1], m.CPU.R[2], m.CPU.R[3], m.CPU.R[13], m.CPU.R[14])
}

// checkWatches reports each watched word whose value changed since the previous
// step, tagged with the PC that was about to execute — the standard bring-up
// instrument for "when does this location get written, and by what."
func (m *Machine) checkWatches(pc uint32) {
	for i := range m.watches {
		w := &m.watches[i]
		for off := uint32(0); off < w.len; off += 4 {
			a := w.addr + off
			v := m.ReadWord(a)
			if w.last == nil {
				w.last = map[uint32]uint32{}
				w.seen = map[uint32]bool{}
			}
			if !w.seen[a] {
				w.seen[a] = true
				w.last[a] = v
				continue
			}
			if v != w.last[a] {
				fmt.Printf("watch 0x%08X: 0x%08X -> 0x%08X at pc=0x%08X\n", a, w.last[a], v, pc)
				w.last[a] = v
			}
		}
	}
}

// HaltReason reports why the machine stopped (empty if it is still runnable).
func (m *Machine) HaltReason() string { return m.CPU.HaltReason }

// DebugString returns the text accumulated from svcOutputDebugString calls.
func (m *Machine) DebugString() string { return string(m.debugOut) }

// Entry is the code entry point the machine boots at.
func (m *Machine) Entry() uint32 { return m.entry }

// Instrs is the total instructions the core has retired.
func (m *Machine) Instrs() uint64 { return m.CPU.Instrs }
