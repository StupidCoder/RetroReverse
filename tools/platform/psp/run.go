package psp

// run.go drives the oracle: it steps the CPU and stops on a budget, a kernel exit,
// a return to the null $ra, an unimplemented instruction, or a tight spin.

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

// Run executes up to maxSteps instructions and returns why it stopped.
func (m *Machine) Run(maxSteps uint64) Result {
	var steps uint64
	spin := map[uint32]bool{}
	const spinWindow = 0x40000
	var sinceReset uint64

	var untilVBlank uint64 = stepsPerVBlank
	for steps < maxSteps {
		// The synthetic display VBlank: deliver the sub-interrupt callbacks games
		// pace their frame loop on.
		if untilVBlank--; untilVBlank == 0 {
			untilVBlank = stepsPerVBlank
			m.deliverVBlank()
			if m.CPU.Halted {
				return Result{steps, m.CPU.PC, "cpu halt in interrupt handler: " + m.CPU.HaltReason}
			}
		}
		if m.CPU.PC == threadExitAddr { // a thread returned: hand to the scheduler
			m.onThreadExit()
			if m.Halted {
				return Result{steps, m.CPU.PC, m.HaltReason}
			}
			continue
		}
		if m.CPU.PC == 0 { // returned to a null $ra
			return Result{steps, 0, "returned to $ra=0 (exit)"}
		}
		if !m.mapped(phys(m.CPU.PC)) {
			// Execution left mapped memory — the HLE wall: a stubbed syscall returned
			// 0 where a real pointer/handle was needed and the game jumped through it.
			return Result{steps, m.CPU.PC, "PC left mapped memory (HLE wall)"}
		}
		if m.OnStep != nil {
			m.OnStep(m, m.CPU.PC)
		}

		// Tight-spin detection: if only a couple of PCs recur across a whole window,
		// the program is stuck (e.g. an idle wait we do not model).
		spin[m.CPU.PC] = true
		sinceReset++
		if sinceReset >= spinWindow {
			if len(spin) < 3 {
				return Result{steps, m.CPU.PC, "spin (tight loop)"}
			}
			spin = map[uint32]bool{}
			sinceReset = 0
		}

		m.CPU.Step()
		steps++
		if m.CPU.Halted {
			return Result{steps, m.CPU.PC, "cpu halt: " + m.CPU.HaltReason}
		}
		if m.Halted {
			return Result{steps, m.CPU.PC, m.HaltReason}
		}
	}
	return Result{steps, m.CPU.PC, "budget reached"}
}

// mapped reports whether a physical address falls in RAM, VRAM or the scratchpad.
func (m *Machine) mapped(p uint32) bool {
	switch {
	case p >= ramBase && p < ramBase+ramSize:
		return true
	case p >= vramBase && p < vramBase+vramSize:
		return true
	case p >= scratchBase && p < scratchBase+scratchSize:
		return true
	}
	return false
}
