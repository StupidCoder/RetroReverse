package threedo

import (
	"fmt"

	"retroreverse.com/tools/arm60"
)

// Result summarises a run.
type Result struct {
	Steps  uint64
	PC     uint32
	Reason string
}

// Run steps the ARM60 up to maxSteps, intercepting synthetic Portfolio folio
// calls (a PC in the HLE window), stubbing each and returning to the caller. It
// stops on a program exit, an unmodelled instruction (CPU halt), a tight self-
// branch spin, or the step budget.
func (m *Machine) Run(maxSteps uint64) Result {
	var steps uint64
	var lastPC uint32
	var spin int
	for steps < maxSteps {
		pc := m.CPU.Reg(15)

		// A PC in the HLE window is an intercepted folio/kernel call.
		if pc >= hleBase && pc < hleBase+hleSize {
			m.serviceKernelCall(pc)
			steps++
			continue
		}
		if m.Halted {
			return Result{steps, pc, m.HaltReason}
		}
		if m.OnStep != nil {
			m.OnStep(m, pc)
		}

		// Tight-spin detection: a branch-to-self (the AIF's post-exit "B .").
		if pc == lastPC {
			spin++
			if spin > 64 {
				return Result{steps, pc, fmt.Sprintf("spin at 0x%08X", pc)}
			}
		} else {
			spin = 0
		}
		lastPC = pc

		m.CPU.Step()
		steps++
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
	m.SetResultAndReturn(0)
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
