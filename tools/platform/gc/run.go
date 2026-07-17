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

// RunStopAfterGXCommand runs until the FIFO interpreter has executed n graphics commands,
// or until the machine stops for one of the usual reasons. It is what the command scrubber
// is built on: restore a frame's start snapshot, run with n = k+1, and the embedded
// framebuffer holds exactly what that frame's first k+1 commands drew.
//
// Stopping mid-list needs no sentinel panic here, unlike the N64's RDP and the PSX's GP0.
// Those consume a DMA'd command list from inside a single CPU instruction, so a plain
// return cannot get out of one and the interpreter has to be unwound through. The
// GameCube's FIFO is drained by an ordinary Go loop that the write-gather pipe feeds, so
// the interpreter simply declines the next command and leaves it in the queue.
//
// It leaves the machine mid-frame with a half-drained pipe and a display list abandoned
// part-way through — a state the machine cannot honestly run on from. That is deliberate
// and it is why this is a scratch-machine call: replay into a copy, read its picture,
// throw it away.
func (m *Machine) RunStopAfterGXCommand(n int, maxSteps uint64) Result {
	m.gxCmdCount, m.gxStopAfter, m.gxStopped = 0, n, false
	res := m.Run(maxSteps)
	m.gxStopAfter, m.gxStopped = 0, false
	return res
}

// GXCommandCount is how many FIFO commands the interpreter has run since the last
// RunStopAfterGXCommand reset it.
func (m *Machine) GXCommandCount() int { return m.gxCmdCount }

// RunFields runs until the video clock has completed n fields, the instruction budget is
// spent, or the machine stops for one of the usual reasons.
//
// A FIELD, NOT A FLIP, IS THE UNIT, and that is the whole point of this call. The pixel
// engine's copy out to the external framebuffer is the honest frame boundary — it is what
// OnFlip fires on and what the profiler closes a frame on — but A BOOT HAS NO FLIPS. The
// machine spends its first emulated seconds running the apploader and reading the disc with
// the graphics pipe idle, so a run bounded by flips would sit there until the budget ran out
// and report nothing. The video clock is instruction-paced (fieldInstructions) and always
// advances, so it bounds a loading stretch and a drawing one alike — which is what a profile
// of "where does the time go before anything is drawn" needs.
//
// The stop lands one instruction after the field boundary: OnDisplay fires from inside
// tickVI, and the run loop consults StopRequested at the top of the next iteration. That is
// deterministic, which is the only property that matters here.
func (m *Machine) RunFields(n int, budget uint64) Result {
	if n <= 0 {
		return Result{PC: m.CPU.PC, Reason: "no fields requested"}
	}
	target := m.vi.Field + uint64(n)

	prev := m.OnDisplay
	m.OnDisplay = func(mm *Machine) {
		if prev != nil {
			prev(mm)
		}
		if mm.vi.Field >= target {
			mm.StopRequested = true
		}
	}
	defer func() { m.OnDisplay = prev }()

	res := m.Run(budget)
	if res.Reason == "stop requested" && m.vi.Field >= target {
		res.Reason = fmt.Sprintf("%d fields", n)
	}
	return res
}

// Run executes up to maxSteps instructions and returns why it stopped. The breakpoint at
// the current PC is not retaken, so a run that stopped at one resumes cleanly.
func (m *Machine) Run(maxSteps uint64) Result {
	m.profRunEnter()
	defer m.profRunExit()

	// THE BUDGET IS EMULATED INSTRUCTIONS, NOT INTERPRETED ONES, and since the idle
	// fast-forward (idle.go) those are no longer the same number. A budget counted in loop
	// iterations would mean "-steps 500000000" quietly ran five times further into the game
	// than it used to, because four fifths of this machine's instructions are an idle loop it
	// no longer executes. The flag names a point to run to, so it has to keep naming the same
	// point whether or not the skip is on.
	start := m.Instrs
	steps := func() uint64 { return m.Instrs - start }
	first := true

	// Tight-spin detection: if only a couple of PCs recur across a whole window, the
	// program is stuck. A real loop — even a wait loop that polls a device — touches more
	// addresses than that, because the poll itself is several instructions.
	//
	// The window is counted with a four-entry array rather than the set this obviously
	// wants, because the set was a map insert PER GEKKO INSTRUCTION — two million of them a
	// field, and ~10% of a boot stretch measured by A/B-ing -nospin, which is the whole of
	// what that flag turns off.
	//
	// The array is not an approximation of the set; it decides the same thing. The
	// predicate is "fewer than four distinct PCs in the window", so ONCE THE FOURTH
	// DISTINCT PC HAS BEEN SEEN THE ANSWER IS ALREADY NO, and nothing later in the window
	// can change it back. So spinN saturates at four and the scan stops: past that point
	// the cost is an increment and a compare, and a run of real code is past that point
	// within a handful of instructions. Below four, a linear scan of at most three entries
	// is cheaper than hashing.
	//
	// spinN < 4 is exactly len(spin) < 4, so the run stops at the same instruction with
	// the same reason. This is not bit-exact by argument, it is the same function.
	// THE WINDOW IS COUNTED IN EMULATED INSTRUCTIONS, not in trips round this loop, and since
	// the idle fast-forward those differ. Counting trips would quietly disable the heuristic
	// exactly where it is most wanted: a machine wedged in `b .` is precisely what the skip
	// fast-forwards, so the window would take a thousand times longer to close and a genuine
	// hang would be reported as an exhausted budget instead of the spin it is.
	var spinPCs [4]uint32
	var spinN int
	const spinWindow = 0x400000
	spinStart := m.Instrs

	for steps() < maxSteps {
		pc := m.CPU.PC

		if m.StopRequested {
			m.StopRequested = false
			return Result{steps(), pc, "stop requested"}
		}
		if m.run.breakpoints[pc] && !first {
			return Result{steps(), pc, fmt.Sprintf("breakpoint at 0x%08X", pc)}
		}
		first = false
		if m.OnStep != nil {
			m.OnStep(m, pc)
		}

		// The call-stack sampler. It reads the stack the CPU is standing on, so it samples
		// before the ticks: an interrupt delivered below would replace that stack with the
		// handler's and book the sample to the wrong work.
		if m.stack.every != 0 && m.Instrs >= m.stack.next {
			m.sampleStack(pc)
		}

		// The idle fast-forward. The detector runs before the ticks so that a loop proved
		// idle skips to the edge of the next event and the ordinary tick below delivers it.
		if !m.noIdle && m.idleStep(pc) {
			if n := m.idleDeadline(); n > 0 {
				// Never skip past the budget. A skip jumps the instruction clock in one go,
				// so without this a run asked for exactly N instructions would sail up to a
				// whole field beyond it — and the caller asking for N is usually asking to
				// stop at a particular place.
				if rem := maxSteps - steps(); n > rem {
					n = rem
				}
				m.idleSkip(n)
			}
		}

		m.tickVI()
		m.tickDSP()
		m.tickAID()
		m.tickDI()

		if !m.noSpin {
			if spinN < 4 {
				seen := false
				for i := 0; i < spinN; i++ {
					if spinPCs[i] == pc {
						seen = true
						break
					}
				}
				if !seen {
					spinPCs[spinN] = pc
					spinN++
				}
			}
			if m.Instrs-spinStart >= spinWindow {
				if spinN < 4 {
					return Result{steps(), pc, "spin (tight loop)"}
				}
				spinN = 0
				spinStart = m.Instrs
			}
		}

		m.CPU.Step()
		m.Instrs++

		if m.CPU.Halted {
			return Result{steps(), m.CPU.PC, m.CPU.HaltReason}
		}
	}
	return Result{steps(), m.CPU.PC, "step budget exhausted"}
}
