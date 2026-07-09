package threedo

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm60"
)

// vblankPeriod is how many instructions map to one virtual VBlank/field. The
// real ratio depends on emulated CPU speed; this is tuned so timing loops see a
// steady stream of fields without spinning the whole step budget on one frame.
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
		// Advance the virtual VBlank counter the game reads out of the OS context
		// struct (osCtx +0xA), so timer/VBlank waits see fields elapse. ~60 Hz is
		// modelled as one field per vblankPeriod instructions.
		if steps%vblankPeriod == 0 {
			// A steady background tick, so timing loops that never issue a field-wait
			// IO still see the clock move. Field-wait IOs advance it too (io.go).
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
				return Result{steps, pc, fmt.Sprintf("all tasks finished (%d switches)", m.switches)}
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
					m.curTask().state = stRunning
				}
			}
			continue
		}
		if m.Halted {
			return Result{steps, pc, m.HaltReason}
		}
		if m.OnStep != nil {
			m.OnStep(m, pc)
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
					// This task has made no progress for a while. In cooperative
					// multitasking it should yield: switch to another runnable task
					// (which may set the flag / build the list it's waiting on).
					m.curTask().state = stReady
					if m.switchTask() {
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
				m.curTask().state = stRunning // nothing else runnable: proceed optimistically
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
