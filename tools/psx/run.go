package psx

// run.go drives the oracle: it steps the CPU, intercepting the BIOS-HLE call
// vectors and the exception vector, and stops on a budget, an unimplemented
// opcode, a program exit, or a tight spin.

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

// Run executes up to maxSteps instructions and returns why it stopped. BIOS
// calls and exceptions are handled between instructions and do not count as
// steps.
func (m *Machine) Run(maxSteps uint64) Result {
	var steps uint64
	spin := map[uint32]bool{}
	const spinWindow = 0x40000
	var sinceReset uint64

	for steps < maxSteps {
		// Apply any scripted controller input whose step has been reached.
		for m.padCursor < len(m.PadScript) && m.CPU.Steps >= m.PadScript[m.padCursor].AtStep {
			m.PadButtons = m.PadScript[m.padCursor].Buttons
			m.padCursor++
		}
		// Synthetic vertical-blank: raise I_STAT bit 0 on a fixed cadence. The game
		// polls I_STAT for timing (and, in-game, unmasks it for a vectored VBlank).
		if m.vblankAcc++; m.vblankAcc >= stepsPerVBlank {
			m.vblankAcc = 0
			m.raiseIRQ(0)
			m.writePad() // refresh the controller buffer, as the BIOS would
		}
		// Advance the CD-ROM controller's queued interrupt responses.
		m.cd.tick()
		// Deliver a pending, enabled, unmasked interrupt: this vectors the CPU to
		// 0x80000080, caught below as PC 0x80. Harmless while the game keeps
		// interrupts masked (Ridge Racer's boot polls instead).
		// Don't deliver a new interrupt while a handler is still running: the ISR
		// trampoline keeps a single saved context and one interrupt stack, so a
		// nested dispatch would corrupt both (the game runs its handler to
		// completion anyway — the BIOS enters with interrupts disabled).
		m.CPU.Interrupt((m.irqStat&m.irqMask) != 0 && !m.isr.active)

		pc := phys(m.CPU.PC)
		switch pc {
		case 0xA0:
			m.biosCall('A')
			continue
		case 0xB0:
			m.biosCall('B')
			continue
		case 0xC0:
			m.biosCall('C')
			continue
		case 0x80:
			m.handleException()
			continue
		case isrReturn:
			m.returnFromISR()
			continue
		}
		if m.CPU.PC == 0 { // returned to the null $ra: program exit
			return Result{steps, m.CPU.PC, "program returned to $ra=0 (exit)"}
		}

		if m.OnStep != nil {
			m.OnStep(m, m.CPU.PC)
		}

		// Tight-spin detection: if only a couple of PCs recur across a whole
		// window, the program is stuck (e.g. `b .`); a real loop touches more.
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
			return Result{steps, m.CPU.PC, "halt: " + m.CPU.HaltReason}
		}
	}
	return Result{steps, m.CPU.PC, "budget reached"}
}
