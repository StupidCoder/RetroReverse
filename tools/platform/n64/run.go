package n64

// run.go drives the oracle: it steps the CPU, delivers RCP interrupts, and stops
// on a budget, a breakpoint, an unimplemented opcode, or a tight spin.

import "fmt"

// Result reports why a run stopped and where.
type Result struct {
	Steps  uint64 // instructions executed
	PC     uint32 // program counter at stop
	Reason string
}

func (r Result) String() string {
	return fmt.Sprintf("stopped at 0x%08X after %d steps: %s", r.PC, r.Steps, r.Reason)
}

// Breakpoints and watch-run control.
type runState struct {
	breakpoints map[uint32]bool
}

// NoSpinDetect turns off the tight-loop heuristic.
//
// Spin detection is an oracle convenience, not a hardware behaviour: a game
// stuck in `b .` has usually hit a gap in the model, and saying so beats running
// out of budget. But a test ROM deliberately spins — waiting on a timer, or
// parked in a handler it means to reach — so the heuristic has to be defeated
// when the thing being run knows what it is doing.
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

// Run executes up to maxSteps instructions and returns why it stopped.
//
// The breakpoint at the current PC is not retaken, so a run that stopped at one
// can simply be resumed.
func (m *Machine) Run(maxSteps uint64) Result {
	var steps uint64
	first := true

	// Tight-spin detection: if only a couple of PCs recur across a whole window,
	// the program is stuck (IPL3 dead-loops this way on a bad CIC seed); a real
	// loop touches more.
	spin := map[uint32]bool{}
	const spinWindow = 0x100000
	var sinceReset uint64

	for steps < maxSteps {
		pc := uint32(m.CPU.PC)

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

		// Deliver a pending, unmasked RCP interrupt. The game installs its own
		// handler at the exception vector: nothing is emulated on its behalf.
		m.CPU.Interrupt(m.irqPending())

		if !m.noSpin {
			spin[pc] = true
			sinceReset++
			if sinceReset >= spinWindow {
				if len(spin) < 3 {
					return Result{steps, pc, "spin (tight loop)"}
				}
				spin = map[uint32]bool{}
				sinceReset = 0
			}
		}

		m.CPU.Step()
		steps++

		if m.CPU.Halted {
			return Result{steps, uint32(m.CPU.CurPC()), m.CPU.HaltReason}
		}
	}
	return Result{steps, uint32(m.CPU.PC), "step budget exhausted"}
}
