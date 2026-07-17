package threedo

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm60"
)

// vblankPeriod is how many instructions map to one virtual field when field
// pacing is OFF: a steady background tick advances the clock so timing loops that
// never issue a field-wait still see it move. With pacing on the clock advances
// per idle field instead (fieldTick), so this is unused on that path.
const vblankPeriod = 20000

// Result summarises a run.
type Result struct {
	Steps  uint64
	PC     uint32
	Reason string
}

// PadStep schedules a control-pad state change: at instruction AtStep the
// buttons become Buttons (0 = all released) until the next entry. The oracle
// fills Machine.PadScript with these to drive menus without a real pad.
type PadStep struct {
	AtStep  uint64
	Buttons uint32
}

// Run steps the ARM60 up to maxSteps, intercepting synthetic Portfolio folio
// calls (a PC in the HLE window), stubbing each and returning to the caller. It
// stops on a program exit, an unmodelled instruction (CPU halt), a tight self-
// branch spin, or the step budget.
func (m *Machine) Run(maxSteps uint64) Result {
	// Time only while the machine is actually running: a debugger paused at a frame
	// boundary for ten minutes must not report ten minutes of ARM60.
	m.profRunEnter()
	defer m.profRunExit()

	var steps uint64
	// Forward-progress tracking: a task that executes no never-before-seen
	// instruction for a while is busy-waiting; if it's a flag spin we switch to
	// another runnable task (the cooperative scheduler), else eventually give up.
	seen := map[uint32]bool{}
	var sinceNew uint64
	const (
		noProgress = 1_000_000
		switchAt   = 20_000 // check for a flag-spin / try a task switch this often
	)
	var ring [64]uint32
	var ri int
	var unstickTries int
	var stallSwitches int // context switches since the last real progress
	for steps < maxSteps {
		// A hook asked the run to end — DisplayScreen closing a frame, or the cel
		// scrubber reaching its limit. Both fire from inside a folio call, which
		// returns to the top of this loop rather than to the interpreter, so this is
		// where the request is honoured: at a clean instruction boundary, with nothing
		// half-done.
		if m.StopRequested {
			m.StopRequested = false
			return Result{steps, m.CPU.Reg(15), "stop requested"}
		}

		// With field pacing OFF the clock advances on a fixed instruction cadence, so
		// timing loops that never issue a field-wait still see it move. With pacing ON
		// the clock instead ticks one field whenever the machine would otherwise idle
		// (fieldTick, in the scheduler paths below), so it tracks displayed frames — a
		// WaitVBL blocks its caller until the field arrives rather than completing
		// instantly, which stops in-game timers racing far ahead of the frames.
		if !m.PaceFields && steps%vblankPeriod == 0 {
			m.advanceVBlank(1)
		}
		// Feed the scheduled control-pad script: deliver each state change once
		// its step arrives (in order; entries already delivered are dropped).
		for len(m.PadScript) > 0 && steps >= m.PadScript[0].AtStep {
			m.SendPadEvent(m.PadScript[0].Buttons)
			m.PadScript = m.PadScript[1:]
		}

		pc := m.CPU.Reg(15)

		// A spawned task returned to its exit trampoline.
		if pc == taskExitTramp {
			m.curTask().state = stDone
			if !m.switchTask() {
				// No other task is ready — but the rest may just be parked on WaitVBL
				// (a transient boot task finishing while the frame loop sleeps between
				// fields). Idle to the next field to wake them before calling it done.
				if !m.wakeByFieldTick() {
					return Result{steps, pc, fmt.Sprintf("all tasks finished (%d switches)", m.switches)}
				}
			}
			continue
		}
		// A PC in the HLE window is an intercepted folio call.
		if pc >= hleBase && pc < hleBase+hleSize {
			m.serviceKernelCall(pc)
			steps++
			if m.needSchedule { // a blocking folio call (SleepUntilTime) yielded
				m.needSchedule = false
				if !m.switchTask() {
					if m.wakeByFieldTick() {
						sinceNew, stallSwitches = 0, 0 // a field completed: progress
					} else {
						m.curTask().state = stRunning
					}
				}
			}
			continue
		}
		if m.Halted {
			return Result{steps, pc, m.HaltReason}
		}
		if m.OnStep != nil {
			m.OnStep(m, pc)
			// A breakpoint stops AT its instruction, not after it.
			if m.StopRequested {
				m.StopRequested = false
				return Result{steps, pc, "stop requested"}
			}
		}

		ring[ri&63] = pc
		ri++
		if pc < dramSize {
			if !seen[pc] {
				seen[pc] = true
				sinceNew = 0
				stallSwitches = 0 // real forward progress
			} else {
				sinceNew++
				if sinceNew%switchAt == 0 {
					// This task has made no new-code progress for a while. Idle the
					// machine one field forward — the console waiting for the next
					// vertical blank — which wakes any WaitVBL waiter and lets a
					// field-counter busy-wait see the new count; then yield to another
					// runnable task (which may set the flag / build the list it waits
					// on). A completed field-wait is real progress: a frame loop that
					// re-runs the same code every field is alive, not deadlocked, so it
					// resets the stall count. Only a machine where the field wakes
					// nothing and no other task can run is the true deadlock.
					woke := false
					if m.PaceFields {
						woke = m.fieldTick()
					}
					m.curTask().state = stReady
					switched := m.switchTask()
					if woke {
						stallSwitches = 0
						sinceNew = 0
						continue
					}
					if switched {
						stallSwitches++
						if stallSwitches > (32*len(m.tasks)+8)*max(1, m.StallTolerance) {
							return Result{steps, pc, fmt.Sprintf("deadlock: all %d tasks stalled near 0x%08X", len(m.tasks), pc)}
						}
						sinceNew = 0
						continue
					}
					m.curTask().state = stRunning
					if m.SpinBreak && m.isFlagSpin(ring[:]) {
						if poked := m.breakSpin(ring[:]); len(poked) > 0 && unstickTries < 2000 {
							unstickTries++
							m.SpinBreaks++
							sinceNew = 0
						}
					}
				}
				if sinceNew > noProgress {
					return Result{steps, pc, fmt.Sprintf("no forward progress (0x%08X, %d switches, %d tasks)", pc, m.switches, len(m.tasks))}
				}
			}
		}

		m.CPU.Step()
		steps++
		if m.needSchedule { // a WaitSignal blocked the current task
			m.needSchedule = false
			if !m.switchTask() {
				if m.wakeByFieldTick() {
					sinceNew, stallSwitches = 0, 0 // a field completed: progress
				} else {
					m.curTask().state = stRunning // nothing else runnable: proceed optimistically
				}
			}
		}
		if m.CPU.Halted {
			return Result{steps, m.CPU.CurPC(), "cpu: " + m.CPU.HaltReason}
		}
		if m.Halted {
			return Result{steps, m.CPU.CurPC(), m.HaltReason}
		}
	}
	return Result{steps, m.CPU.Reg(15), "step budget reached"}
}

// serviceKernelCall handles a PC that landed in the HLE window: it records which
// folio offset was invoked and its arguments, stubs a zero result and returns to
// the caller (LR).
func (m *Machine) serviceKernelCall(pc uint32) {
	// Time the folio HLE, less any cel work done inside it. DrawCels draws a whole
	// chain without returning, so a folio bucket that did not subtract the cel engine
	// would count every cel twice and drive the derived remainder negative.
	m.celCnt.folios++
	tf, gfxBefore, gen := m.profStart(), m.profGfxNs(), m.prof.gen
	defer m.profEndFolio(tf, gfxBefore, gen) // a method value, not a closure: no allocation

	off := pc - hleBase
	m.KernelCalls = append(m.KernelCalls, KernelCall{
		Offset: off,
		From:   m.CPU.Reg(14) - 8,
		Args:   [4]uint32{m.CPU.Reg(0), m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3)},
	})
	// A call in a folio slice of the window is that folio's vector call. Slices are
	// ordered by tag; test the highest tag first.
	if off >= hleAudioTag {
		m.serviceAudioFolio(off - hleAudioTag)
		return
	}
	if off >= hleGfxTag {
		m.serviceGraphicsFolio(off - hleGfxTag)
		return
	}
	if off >= hleOtherTag {
		m.serviceOtherFolio(off - hleOtherTag)
		return
	}
	if off >= hleFileTag {
		m.serviceFileFolio(off - hleFileTag)
		return
	}
	if off >= hleMathTag {
		m.serviceMathFolio(off - hleMathTag)
		return
	}
	// Reimplemented folios handle themselves (including the return); anything not
	// yet reimplemented falls back to a zero-result stub.
	if !m.serviceFolio(off) {
		m.SetResultAndReturn(0)
	}
}

// SetResultAndReturn stubs r0 and returns to LR — the default for an unmodelled
// folio call, and a hook point once individual calls are reimplemented.
func (m *Machine) SetResultAndReturn(result uint32) {
	m.CPU.SetReg(0, result)
	m.CPU.SetPC(m.CPU.Reg(14))
}

// DisasmAt disassembles one instruction at addr from DRAM (for trace output).
func (m *Machine) DisasmAt(addr uint32) string {
	if int(addr)+4 > len(m.dram) {
		return fmt.Sprintf("%08X  ????", addr)
	}
	in := arm60.Decode(m.dram[addr:], addr)
	return fmt.Sprintf("%08X  %s", addr, in.Text)
}
