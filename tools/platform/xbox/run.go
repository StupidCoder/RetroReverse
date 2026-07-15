package xbox

// run.go is the boot driver: it steps the CPU under an instruction budget and reports
// how far the title's XDK/CRT boot code reached. A run ends for one of three reasons —
// the instruction budget was spent, the CPU halted (an unmodelled ordinal, a fault, a
// deadlock), or the first NV2A push-buffer kick landed, which is the Phase-B goal.

import "fmt"

// StopReason explains why Run returned.
type StopReason int

const (
	StopBudget    StopReason = iota // the instruction budget was exhausted
	StopHalt                        // the CPU halted (see HaltReason)
	StopFirstPush                   // the first NV2A push-buffer kick was reached (goal)
)

func (r StopReason) String() string {
	switch r {
	case StopBudget:
		return "budget exhausted"
	case StopHalt:
		return "halted"
	case StopFirstPush:
		return "first NV2A push reached"
	default:
		return "?"
	}
}

// Run steps up to maxSteps instructions, stopping early on a halt or the first NV2A
// push. It returns the reason and the number of instructions executed.
func (m *Machine) Run(maxSteps uint64) (StopReason, uint64) {
	var n uint64
	for n < maxSteps {
		if m.CPU.Halted {
			return StopHalt, n
		}
		if m.firstPush && !m.pusherEnabled {
			return StopFirstPush, n
		}
		m.CPU.Step()
		n++
	}
	return StopBudget, n
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
