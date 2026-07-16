package xbox

// run.go is the boot driver: it steps the CPU under an instruction budget and reports
// how far the title's XDK/CRT boot code reached. A run ends for one of three reasons —
// the instruction budget was spent, the CPU halted (an unmodelled ordinal, a fault, a
// deadlock), or the first NV2A push-buffer kick landed, which is the Phase-B goal.

import (
	"fmt"

	"retroreverse.com/tools/cpu/x86"
)

// StopReason explains why Run returned.
type StopReason int

const (
	StopBudget    StopReason = iota // the instruction budget was exhausted
	StopHalt                        // the CPU halted (see HaltReason)
	StopFirstPush                   // the first NV2A push-buffer kick was reached (goal)
	StopRequest                     // a hook set StopRequested (a flip, a breaking watch)
	StopBreak                       // an execution breakpoint was reached
)

func (r StopReason) String() string {
	switch r {
	case StopBudget:
		return "budget exhausted"
	case StopHalt:
		return "halted"
	case StopFirstPush:
		return "first NV2A push reached"
	case StopRequest:
		return "stop requested"
	case StopBreak:
		return "breakpoint"
	default:
		return "?"
	}
}

// Run steps up to maxSteps instructions, stopping early on a halt, the first NV2A push,
// a breakpoint, or a hook's stop request. It returns the reason and the number of
// instructions executed.
func (m *Machine) Run(maxSteps uint64) (StopReason, uint64) {
	m.StopRequested = false
	var n uint64
	for n < maxSteps {
		if m.CPU.Halted {
			return StopHalt, n
		}
		if m.firstPush && !m.pusherEnabled {
			return StopFirstPush, n
		}
		// A breakpoint stops BEFORE its instruction executes, so the machine is parked
		// exactly at the PC the user asked about and stepping on from here runs it.
		//
		// The address tested is the NEXT instruction's (SegBase[CS]+IP, the same thing
		// onStep dispatches on), not CPU.LinearPC() — that reports instrIP, the
		// instruction currently executing, which between steps is the one just RETIRED.
		// Testing it would fire every breakpoint exactly one instruction late.
		//
		// Checked only after at least one step, so a Continue from a breakpoint's own PC
		// makes progress instead of stopping on the spot forever.
		if n > 0 && len(m.bps) > 0 && m.bps[m.PC()] {
			return StopBreak, n
		}
		m.CPU.Step()
		n++
		if m.StopRequested {
			m.StopRequested = false
			return StopRequest, n
		}
	}
	return StopBudget, n
}

// RunStopAfterNVMethod runs until the pusher has dispatched k more methods, then stops.
// It is the command scrubber's engine: replay a frame from its start snapshot and stop
// after command k, leaving the render target holding the frame as it stood right then.
//
// Unlike the GameCube's FIFO — drained by a loop the caller owns — the NV2A's pusher runs
// inside the guest's own store to DMA_PUT, so this cannot simply decline the next command:
// it arms a countdown that the pusher's method dispatch trips, which stops the pusher
// mid-buffer and the CPU with it. The machine is left mid-frame with GET short of PUT,
// which is exactly why this belongs on a scratch machine and never on the live one.
func (m *Machine) RunStopAfterNVMethod(k int, maxSteps uint64) (StopReason, uint64) {
	m.stopAfterMethod, m.stopAfterArmed = k, true
	defer func() { m.stopAfterMethod, m.stopAfterArmed = 0, false }()
	return m.Run(maxSteps)
}

// PC is the linear address of the instruction the machine will execute NEXT — where a
// parked machine sits. It is deliberately not CPU.LinearPC(), which reports the
// instruction being executed (and so, between steps, the one just retired): a debugger
// asks "where am I stopped?", and this answers it.
func (m *Machine) PC() uint32 { return m.CPU.SegBase[x86.CS] + m.CPU.IP }

// SetBreakpoint / ClearBreakpoint / ClearBreakpoints / Breakpoints hold execution
// breakpoints, checked against the linear PC between instructions.
func (m *Machine) SetBreakpoint(pc uint32) {
	if m.bps == nil {
		m.bps = map[uint32]bool{}
	}
	m.bps[pc] = true
}

func (m *Machine) ClearBreakpoint(pc uint32) { delete(m.bps, pc) }

func (m *Machine) ClearBreakpoints() { m.bps = nil }

func (m *Machine) Breakpoints() []uint32 {
	out := make([]uint32, 0, len(m.bps))
	for pc := range m.bps {
		out = append(out, pc)
	}
	sortU32(out)
	return out
}

// ClearHalt clears a halted machine so a run can resume. An unimplemented-ordinal
// halt stops with EIP still at the trap sentinel and nothing mutated (dispatchKernel
// halts instead of dispatching), so after the ordinal gains a handler, clearing the
// halt retries the very call that stopped the previous run — the fix-and-resume
// workflow for frontier savestates.
func (m *Machine) ClearHalt() {
	m.CPU.Halted, m.CPU.HaltReason = false, ""
	m.Halted, m.HaltReason = false, ""
}

// Report renders a one-screen summary of a run's end state — the standard oracle
// frontier statement.
func (m *Machine) Report() string {
	c := m.CPU
	s := fmt.Sprintf("xbox: %q (title id %08X)\n", m.XBE.TitleName, m.XBE.TitleID)
	s += fmt.Sprintf("  entry %08X  base %08X  imageSize %08X\n", m.XBE.Entry, m.XBE.Base, m.XBE.ImageSize)
	s += fmt.Sprintf("  steps=%d  PC=%08X  EAX=%08X EBX=%08X ECX=%08X EDX=%08X ESP=%08X\n",
		c.Steps, c.LinearPC(), c.Regs[0], c.Regs[3], c.Regs[1], c.Regs[2], c.Regs[4])
	if c.Halted {
		s += "  HALT: " + c.HaltReason + "\n"
	}
	if m.firstPush {
		s += "  ★ first NV2A push-buffer kick reached\n"
	}
	s += fmt.Sprintf("  distinct ordinals called: %d\n", len(m.OrdinalHits))
	return s
}

// OrdinalHistogram returns the called ordinals with their names and counts, in
// ordinal order — the boot's kernel-surface reach.
func (m *Machine) OrdinalHistogram() []string {
	// Collect and sort the keys.
	keys := make([]int, 0, len(m.OrdinalHits))
	for o := range m.OrdinalHits {
		keys = append(keys, int(o))
	}
	// insertion sort — the set is small
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	out := make([]string, 0, len(keys))
	for _, o := range keys {
		out = append(out, fmt.Sprintf("  ord %3d %-34s x%d", o, ordinalName(uint16(o)), m.OrdinalHits[uint16(o)]))
	}
	return out
}
