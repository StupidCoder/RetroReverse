package gc

// run.go drives the machine: it steps the Gekko, ticks the video clock that is the game's
// heartbeat, delivers interrupts, and stops on a budget, a breakpoint, a halt, or a tight
// spin. It is the N64 oracle's run loop with the GameCube's devices — the same shape,
// because the two machines are the same shape: hardware to model, and no operating system
// to service.

import "fmt"

// Result reports why a run stopped and where.
type Result struct {
	Steps  uint64
	PC     uint32
	Reason string
}

func (r Result) String() string {
	return fmt.Sprintf("stopped at 0x%08X after %d steps: %s", r.PC, r.Steps, r.Reason)
}

type runState struct {
	breakpoints map[uint32]bool
}

// SetSpinDetect turns the tight-loop heuristic on or off. A game stuck in `b .` has usually
// hit a gap in the model, and saying so beats running out of budget — but a boot that means
// to park in a wait loop needs the heuristic defeated, so it is a switch.
func (m *Machine) SetSpinDetect(on bool) { m.noSpin = !on }

// SetBreakpoint stops the run when the CPU is about to execute vaddr.
func (m *Machine) SetBreakpoint(vaddr uint32) {
	if m.run.breakpoints == nil {
		m.run.breakpoints = map[uint32]bool{}
	}
	m.run.breakpoints[vaddr] = true
}

// ClearBreakpoints removes every breakpoint.
func (m *Machine) ClearBreakpoints() { m.run.breakpoints = nil }

// Run executes up to maxSteps instructions and returns why it stopped. The breakpoint at
// the current PC is not retaken, so a run that stopped at one resumes cleanly.
func (m *Machine) Run(maxSteps uint64) Result {
	var steps uint64
	first := true

	// Tight-spin detection: if only a couple of PCs recur across a whole window, the
	// program is stuck. A real loop — even a wait loop that polls a device — touches more
	// addresses than that, because the poll itself is several instructions.
	spin := map[uint32]bool{}
	const spinWindow = 0x400000
	var sinceReset uint64

	for steps < maxSteps {
		pc := m.CPU.PC

		if m.StopRequested {
			m.StopRequested = false
			return Result{steps, pc, "stop requested"}
		}
		if m.run.breakpoints[pc] && !first {
			return Result{steps, pc, fmt.Sprintf("breakpoint at 0x%08X", pc)}
		}
		first = false
		if m.OnStep != nil {
			m.OnStep(m, pc)
		}

		m.tickVI()
		m.tickDSP()

		if !m.noSpin {
			spin[pc] = true
			sinceReset++
			if sinceReset >= spinWindow {
				if len(spin) < 4 {
					return Result{steps, pc, "spin (tight loop)"}
				}
				spin = map[uint32]bool{}
				sinceReset = 0
			}
		}

		m.CPU.Step()
		steps++

		if m.CPU.Halted {
			return Result{steps, m.CPU.PC, m.CPU.HaltReason}
		}
	}
	return Result{steps, m.CPU.PC, "step budget exhausted"}
}
