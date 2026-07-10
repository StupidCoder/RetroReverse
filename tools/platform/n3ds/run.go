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
func (m *Machine) Run(budget int) int {
	n := 0
	for n < budget {
		if m.CPU.Halted {
			break
		}
		t := m.pickRunnable()
		if t == nil {
			if !m.advanceIdle() {
				m.dumpThreads()
				m.CPU.Halt("all threads blocked (deadlock): %d live, none runnable, after %d instructions",
					m.aliveThreads(), m.CPU.Instrs)
				break
			}
			continue
		}
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
				fmt.Printf("breakpoint at 0x%08X after %d instructions\n", pc, n)
				m.stopped = true
				break
			}
			if m.Trace && m.traceN < m.traceMax {
				m.traceOne(pc)
				m.traceN++
			}
			m.checkWatches(pc)
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
	fmt.Printf("%08X: %-24s  r0=%08X r1=%08X r2=%08X r3=%08X sp=%08X lr=%08X\n",
		pc, in.Text, m.CPU.R[0], m.CPU.R[1], m.CPU.R[2], m.CPU.R[3], m.CPU.R[13], m.CPU.R[14])
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
