package dc

// run.go is the run loop: instruction-budgeted, with breakpoints, the VBlank
// heartbeat, and tight-spin detection. The budget counts instructions
// *retired* so a later fast-forward optimisation cannot silently change what
// -steps N means.

import "fmt"

// Result says why a run stopped.
type Result struct {
	Steps  uint64
	PC     uint32
	Reason string
}

func (r Result) String() string {
	return fmt.Sprintf("stopped after %d steps at PC %08X: %s", r.Steps, r.PC, r.Reason)
}

// spinWindow is how long the PC set may stay tiny before the run is declared
// a spin: two video fields, so an idle loop that wakes every VBlank is never
// misdiagnosed.
const spinWindow = 2 * fieldInstructions

// RunConfig tunes a run; the zero value is a plain bounded run.
type RunConfig struct {
	Breakpoints map[uint32]bool
	NoSpin      bool // disable spin detection
}

// Run executes up to maxSteps instructions.
func (m *Machine) Run(maxSteps uint64, cfg RunConfig) Result {
	var spinPCs [4]uint32
	var spinCount int
	var spinSteps uint64

	for start := m.Instrs; m.Instrs-start < maxSteps; {
		pc := m.CPU.PC
		if cfg.Breakpoints[pc] {
			return Result{m.Instrs, pc, "breakpoint"}
		}
		if m.OnStep != nil {
			m.OnStep(pc)
		}
		if m.trapPC(pc) {
			continue // an HLE call serviced; PC has moved on
		}

		m.CPU.Step()
		m.Instrs++
		if m.CPU.Halted {
			return Result{m.Instrs, m.CPU.CurPC(), "halt: " + m.CPU.HaltReason}
		}

		if m.armAcc++; m.armAcc >= armDivisor {
			m.armAcc = 0
			m.stepARM()
		}
		m.tickAICA()
		m.tickCompletions()

		m.tickField()

		if !cfg.NoSpin {
			known := false
			for i := 0; i < spinCount; i++ {
				if spinPCs[i] == pc {
					known = true
					break
				}
			}
			if !known {
				if spinCount < len(spinPCs) {
					spinPCs[spinCount] = pc
					spinCount++
				} else {
					spinCount, spinSteps = 0, 0 // a fifth PC: not a spin, restart the window
					spinPCs[0], spinCount = pc, 1
				}
			}
			spinSteps++
			if spinSteps >= spinWindow && spinCount <= len(spinPCs) {
				return Result{m.Instrs, pc, fmt.Sprintf("spin: %d distinct PCs over %d instructions", spinCount, spinSteps)}
			}
		}
	}
	return Result{m.Instrs, m.CPU.PC, "steps"}
}

// RunFields runs until n more VBlank fields have been delivered, under a
// step budget. A field, not a flip, is the unit, because a boot has no flips.
func (m *Machine) RunFields(n uint64, budget uint64) Result {
	target := m.Fields + n
	for m.Fields < target {
		r := m.Run(minU64(budget, fieldInstructions+1), RunConfig{})
		budget -= minU64(budget, fieldInstructions+1)
		if r.Reason != "steps" {
			return r
		}
		if budget == 0 {
			return Result{m.Instrs, m.CPU.PC, "field budget exhausted"}
		}
	}
	return Result{m.Instrs, m.CPU.PC, "fields"}
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// tickField advances the scan position. The field is a real line sweep, not
// a bare heartbeat: the counter walks 0..total-1 at fieldInstructions per
// field, VBlank-in and -out fire at the lines the game programmed into
// SPG_VBLANK_INT, and SPG_STATUS reads the live line — a game that verifies
// "am I truly in blanking" before flipping must be able to see yes.
func (m *Machine) tickField() {
	m.instrInField++
	total := m.spgTotalLines()
	if m.instrInField < uint64(fieldInstructions/total) {
		return
	}
	m.instrInField = 0
	m.CurLine++
	if m.CurLine >= total {
		m.CurLine = 0
		m.FieldNum ^= 1
	}
	vbl := m.PVRRegs[0xCC/4]
	switch m.CurLine {
	case vbl & 0x3FF: // vblank-in: the field boundary
		m.Fields++
		m.raiseNRM(istVBlankIn)
		if m.OnDisplay != nil {
			m.OnDisplay(m.Fields)
		}
	case vbl >> 16 & 0x3FF:
		m.raiseNRM(istVBlankOut)
	}
}

// spgTotalLines is the field's line count from SPG_LOAD, with the NTSC total
// standing in until the game programs one.
func (m *Machine) spgTotalLines() uint32 {
	if t := m.PVRRegs[0xF8/4] >> 16 & 0x3FF; t != 0 {
		return t + 1
	}
	return 525
}

