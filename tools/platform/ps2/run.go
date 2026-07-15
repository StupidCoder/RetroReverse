package ps2

// run.go drives the machine: it steps the CPU, paces the frame clock, and stops for
// the reasons a caller cares about.
//
// Four of those reasons are worth naming, because between them they turn "the run
// ended" into "the run ended *here*, and this is why":
//
//	the HLE wall     the PC has left mapped memory. Almost always a call into the
//	                 BIOS that the kernel HLE does not implement, or a jump through a
//	                 pointer that was never written.
//	a spin           the PC has been going round a handful of addresses for a long
//	                 time and nothing is changing. A boot that waits forever on a
//	                 peripheral that does not exist looks exactly like this, and
//	                 without the detector it looks like nothing at all.
//	a thread exit    the PC reached the sentinel return address a thread was given.
//	a breakpoint     the caller asked.

import (
	"fmt"

	"retroreverse.com/tools/cpu/r5900"
)

// Result is what a run reports.
type Result struct {
	Steps  uint64
	PC     uint32
	Reason string
}

func (r Result) String() string {
	return fmt.Sprintf("stopped after %d steps at 0x%08X: %s", r.Steps, r.PC, r.Reason)
}

// spinWindow is how many instructions the detector watches before deciding a run is
// going nowhere; spinDistinct is how few distinct PCs in that window count as stuck.
const (
	spinWindow   = 0x200000
	spinDistinct = 6
)

// trailLen is how many recent control-transfer edges the run keeps, to print as a trail
// when it stops somewhere it should not have — a wild jump, an unhandled exception, a PC
// off the map. Recording *edges* (the instruction that branched and the address it
// branched to) rather than every PC is what makes a wild jump legible: a jump into a
// nop-slide of zeroed heap can slide for thousands of instructions, and a ring of raw PCs
// fills with the slide; the edge that entered it is the one line that matters. It is the
// EE's version of the IOP's -ioptrap trail.
const trailLen = 24

// Run steps the machine until the budget is spent or something stops it.
func (m *Machine) Run(maxSteps uint64) Result {
	var (
		steps     uint64
		vblAcc    uint64
		iopAcc    uint64
		spinSeen  = map[uint32]int{}
		spinAcc   int
		trail     [trailLen][2]uint32 // {from, to} for each control transfer
		trailPos  int
		trailPrev uint32
	)

	for steps < maxSteps {
		if m.Halted || m.CPU.Halted {
			break
		}
		if m.StopRequested {
			m.StopRequested = false
			return m.result(steps, "stop requested")
		}

		// The reboot the EE asked for, once the second processor has had time to perform it.
		// It happens here rather than where it was requested because it must not finish before
		// the routine that asked for it has — see iopReset.
		if m.iopRebootImage != "" && m.steps >= m.iopRebootAt {
			image := m.iopRebootImage
			m.iopRebootImage = ""
			if err := m.RebootIOPFrom(image); err != nil {
				m.note("SIF: the IOP did not reboot: %v", err)
			}
		}

		// The second processor. It runs at about an eighth of the EE's clock, and it runs
		// whether or not the EE does — a machine whose every thread is blocked is very
		// often a machine waiting on the IOP, and an IOP that only ticked when the EE did
		// would never arrive.
		if m.IOP != nil {
			iopAcc++
			if iopAcc >= iopStepRatio {
				iopAcc = 0
				m.IOP.Step()
			}
		}

		// The frame clock. Everything time-based in this machine hangs off it.
		vblAcc++
		if vblAcc >= stepsPerVBlank {
			vblAcc = 0
			m.deliverVBlank()
			if m.Halted {
				break
			}
		}

		// Every thread is blocked. The CPU runs nothing at all — the clock above is the
		// only thing still moving, and an interrupt or a reply from the IOP is what will
		// make a thread ready again.
		//
		// If nothing *can* do that, the machine is genuinely deadlocked rather than
		// waiting, and saying so is far more use than letting the step budget run out.
		if m.idle {
			steps++
			m.steps++
			if !m.resume() {
				if m.blocked() {
					return m.result(steps, "deadlocked: every thread is blocked, and nothing left can wake one")
				}
				continue
			}
		}

		pc := uint32(m.CPU.PC)

		// The sentinels. A thread that runs off the end of its entry point, and a guest
		// routine called by the machine, each return to an address that is not memory —
		// which is how the loop learns they have finished.
		if pc == threadExitAddr {
			m.onThreadExit()
			continue
		}

		if m.breakpoints[pc] {
			return m.result(steps, "breakpoint at "+m.Sym(pc))
		}

		// An exception with no handler behind it. The CPU vectors correctly, but under
		// this model nothing has been installed at the vector yet, so the PC lands on
		// zeroed memory — which decodes as `nop`, and the machine slides quietly through
		// memory until it falls into something. That failure is silent and looks like a
		// hang somewhere else entirely, so it is caught here and named.
		if r, caught := m.unhandledException(pc); caught {
			m.note("jump trail before the exception:%s", m.formatTrail(trail[:], trailPos))
			return m.result(steps, r)
		}

		if !m.mapped(phys(pc)) {
			m.note("jump trail before the PC left memory:%s", m.formatTrail(trail[:], trailPos))
			return m.result(steps, fmt.Sprintf(
				"the PC left mapped memory (0x%08X) — an unimplemented kernel call, or a jump through an unwritten pointer",
				pc))
		}

		// Record a control-transfer edge: any step where the PC is not the previous one plus
		// four is a branch, jump or exception return, and its {from, to} is what the trail is.
		if trailPrev != 0 && pc != trailPrev+4 {
			trail[trailPos] = [2]uint32{trailPrev, pc}
			trailPos = (trailPos + 1) % trailLen
		}
		trailPrev = pc

		if m.OnStep != nil {
			m.OnStep(m, pc)
			if m.Halted {
				break
			}
		}

		// The spin detector. It counts distinct PCs over a long window; a window that
		// sees only a handful of them is a loop that is not going anywhere.
		spinSeen[pc]++
		spinAcc++
		if spinAcc >= spinWindow {
			if len(spinSeen) <= spinDistinct {
				return m.result(steps, fmt.Sprintf(
					"spinning on %d addresses around %s — waiting for something that never happens",
					len(spinSeen), m.Sym(pc)))
			}
			spinSeen = map[uint32]int{}
			spinAcc = 0
		}

		m.CPU.Step()
		steps++
		m.steps++
	}

	if m.CPU.Halted {
		return m.result(steps, "cpu: "+m.CPU.HaltReason)
	}
	if m.Halted {
		return m.result(steps, m.HaltReason)
	}
	return m.result(steps, "step budget exhausted")
}

func (m *Machine) result(steps uint64, reason string) Result {
	return Result{Steps: steps, PC: uint32(m.CPU.PC), Reason: reason}
}

// formatTrail renders the ring of recent control-transfer edges oldest-first, each as the
// instruction that branched and the address it reached. The last line is the transfer that
// ended the run — for a wild jump, the routine it left from and the garbage it landed in.
func (m *Machine) formatTrail(trail [][2]uint32, pos int) string {
	s := ""
	for i := 0; i < len(trail); i++ {
		e := trail[(pos+i)%len(trail)]
		if e[0] == 0 && e[1] == 0 {
			continue
		}
		s += fmt.Sprintf("\n  %-34s -> %s", m.Sym(e[0]), m.Sym(e[1]))
	}
	return s
}

// The three addresses the EE vectors an exception to, with Status BEV clear.
var exceptionVectors = map[uint32]string{
	0x80000000: "TLB refill",
	0x80000080: "counter",
	0x80000180: "general",
}

// unhandledException reports whether the PC has landed on an exception vector that has
// no handler behind it, and if so says which exception it was and where it came from.
//
// EPC and Cause are still intact at this point, so the report can name the faulting
// instruction rather than just the vector — which is the whole difference between "the
// machine wandered off" and "the store at InitAlarm+0x1C touched an unmapped address".
func (m *Machine) unhandledException(pc uint32) (string, bool) {
	name, isVector := exceptionVectors[pc]
	if !isVector {
		return "", false
	}
	if m.Fetch32(pc) != 0 {
		return "", false // a handler is installed; let it run
	}

	cause := m.CPU.COP0[r5900.Cop0Cause]
	code := (cause >> 2) & 0x1F
	epc := uint32(m.CPU.COP0[r5900.Cop0EPC])
	bad := uint32(m.CPU.COP0[r5900.Cop0BadVAddr])

	return fmt.Sprintf("unhandled %s exception (%s) at %s — faulting address 0x%08X, and nothing is installed at the vector",
		name, excName(uint32(code)), m.Sym(epc), bad), true
}

func excName(code uint32) string {
	switch code {
	case 0x00:
		return "interrupt"
	case 0x01:
		return "TLB modification"
	case 0x02:
		return "TLB miss on a load or fetch"
	case 0x03:
		return "TLB miss on a store"
	case 0x04:
		return "address error on a load or fetch"
	case 0x05:
		return "address error on a store"
	case 0x08:
		return "syscall"
	case 0x09:
		return "breakpoint"
	case 0x0A:
		return "reserved instruction"
	case 0x0B:
		return "coprocessor unusable"
	case 0x0C:
		return "arithmetic overflow"
	case 0x0D:
		return "trap"
	}
	return fmt.Sprintf("code %d", code)
}
