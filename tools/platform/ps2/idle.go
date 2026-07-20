package ps2

// idle.go fast-forwards the clock through a loop the EE cannot break from where it is.
//
// SEVENTY-TWO PER CENT OF A DRAWING FIELD IS ONE FUNCTION. A PC histogram over the Jak
// title field (bootoracle -eeprof) puts 72% of every EE instruction retired inside the
// engine's `VSync`, which is eight instructions:
//
//	1237F0  lui   $v0, 0x1001
//	1237F4  lw    $v0, -4096($v0)   ; read I_STAT
//	1237F8  andi  $v0, $v0, 0x4     ; the VBlank-start bit
//	1237FC  nop
//	123800  nop
//	123804  nop
//	123808  beq   $v0, $zero, 1237F0 ; spin until the blank arrives
//	12380C  nop
//
// It is not a gap in the model — it is the game waiting, correctly, for the vertical
// blank. stepsPerVBlank (machine.go) gives the interpreter a million instructions a
// field where real hardware retires nearer five, so the game finishes its frame's work
// early and spends the rest of every field here. The emulator was spending most of a
// field proving that a bit it cannot set from the CPU is still clear.
//
// WHAT MAKES THE SKIP EXACT — the same argument as the GameCube's (gc/idle.go), with a
// PlayStation 2's own list of what can change behind the processor's back:
//
// If the EE returns to a state it has already been in — the same PC, the same registers,
// the same everything the interpreter reads — having stored nothing, then it is a
// function with no inputs that has repeated itself. It will repeat forever until
// SOMETHING OUTSIDE THE PROCESSOR changes. So the skip does not approximate the loop; it
// computes its limit. The whole risk sits in "what can change while the EE is not
// looking", and on this machine that list is:
//
//	the vertical blank            deliverVBlank sets I_STAT's VBlank-start bit; the
//	                              trailing edge sets VBlank-end half a field later. Both
//	                              are instruction-paced off stepsPerVBlank.
//	the Count/Compare timer       an interrupt with no device — the EE's own clock,
//	                              ticked every second instruction, raising IP7 at Compare.
//	                              The least obvious, and only a deadline when it can fire.
//	the second processor          the IOP writes EE memory and wakes EE threads across the
//	                              SIF. Unlike the GameCube's DSP it is NOT a veto (it never
//	                              sleeps), so it is stepped THROUGH the skip at its real
//	                              1:8 cadence, and the skip STOPS the instant it reaches
//	                              across (eeDisturbGen) — the woken thread has to run at the
//	                              instruction it was woken on, not at the end of a jump.
//
// The vblank and the timer are deadlines: go to the nearest and stop one loop-period
// short, so the run loop runs the last few iterations and delivers the event itself. The
// IOP is interleaved, not skipped, which is why this saves the EE's idle instructions but
// not the IOP's — the IOP does exactly the work it always would.
//
// SKIP A WHOLE NUMBER OF LOOP PERIODS, not to the raw deadline. The state repeat gives the
// period; advancing a multiple of it leaves the EE at the very instruction an un-skipped
// run would, so the vblank interrupts the same PC and the machine is not merely
// equivalent but identical — which for an oracle is the only test that counts. The
// GameCube learned this cost two bytes of RAM the other way (its frame-time counter
// noticed); here TestIdleSkipMatchesSerial is the standing proof, and the frame-hash gate
// the backstop.
//
// Count is carried by SkipInstructions, not snapshotted: it ticks every second
// instruction regardless of the loop, so a snapshot that included it would never repeat
// and the skip would silently never fire — the R5900's version of the NaN that made the
// GameCube hash floats by bits.

import "retroreverse.com/tools/cpu/r5900"

// idleSnap is the architectural state a loop must return to, unchanged, to prove it is
// idle. It is everything the interpreter reads to decide what to do next EXCEPT the
// free-running Count timer and the COP0 file behind it (see the file comment), and the
// TLB (a poll loop does not remap memory). It is a plain comparable value, so a repeat is
// one `==`.
//
// Floating point is carried by BITS (FPR is already the raw uint32 the core stores), so a
// register file holding a NaN — which compares unequal to itself — does not make every
// snapshot look different and defeat the skip.
type idleSnap struct {
	PC, NextPC, BranchAddr uint64
	HI, LO, HI1, LO1       uint64
	SA, ACC, FCR31         uint32
	DelaySlot, PendingDelay, LLBit bool
	R                      [32]r5900.Quad
	FPR                    [32]uint32
}

// idleState is the search for a state repeat, plus what it has bought.
type idleState struct {
	armed  bool
	snap   idleSnap
	stores uint64 // m.stores when the snapshot was taken
	insns  int    // instructions watched since arming
	period int    // the proved loop length, in instructions
	next   uint64 // the step count at which to arm again

	// phase[i] is the full core state i instructions into the loop from its boundary.
	// A skip freezes the EE at the boundary and only ever pays for the loop it elides IF
	// nothing looks at where in the loop the EE actually is — but the IOP can reach across
	// and park or wake the idle thread mid-period, and then its exact phase matters. This
	// is captured once per skip so that phase can be restored in O(1) at the rare instant
	// it is observed, without re-running the loop against memory the disturbance changed.
	phase []r5900.State

	Skipped uint64 // instruction-slots fast-forwarded — the diagnostic to look at
	Hits    uint64 // how many times a loop was proved idle and skipped
}

// How the detector paces itself. Arming costs a snapshot and watching costs a compare per
// instruction, so neither runs all the time: arm once every idleArmEvery instructions,
// give up after idleWatch. VSync's loop is eight instructions and repeats well inside the
// window; a loop that does not repeat within sixty-four instructions is doing enough work
// that a skip would not have paid.
const (
	idleArmEvery = 4096
	idleWatch    = 64
)

// snapshotEE takes the comparable core state. It reads the exported register file
// directly and the delay-slot machinery through the core's getters; nothing here is a
// clock.
func (m *Machine) snapshotEE() idleSnap {
	c := m.CPU
	return idleSnap{
		PC: c.PC, NextPC: c.NextPC(), BranchAddr: c.BranchAddr(),
		HI: c.HI, LO: c.LO, HI1: c.HI1, LO1: c.LO1,
		SA: c.SA, ACC: c.ACC, FCR31: c.FCR31,
		DelaySlot: c.InDelaySlot(), PendingDelay: c.PendingDelay(), LLBit: c.LLBit,
		R: c.R, FPR: c.FPR,
	}
}

// idleStep is the detector, on the run loop's hot path: one compare per instruction when
// it is not armed, the register-file compare on top when it is. It reports whether the EE
// has just been proved idle at the state it is about to step from.
func (m *Machine) idleStep() bool {
	d := &m.idleDet
	if !d.armed {
		if m.steps < d.next {
			return false
		}
		d.armed = true
		d.insns = 0
		d.stores = m.stores
		d.snap = m.snapshotEE()
		return false
	}

	d.insns++
	if d.insns > idleWatch {
		d.armed = false
		d.next = m.steps + idleArmEvery
		return false
	}
	// The cheap tests first — the same PC, and nothing stored on the way round — then the
	// whole register file only if those pass.
	if m.CPU.PC != d.snap.PC || m.stores != d.stores {
		return false
	}
	if m.snapshotEE() != d.snap {
		return false
	}
	d.period = d.insns
	d.armed = false
	d.next = m.steps + idleArmEvery
	return true
}

// idleDeadline is how many instruction-slots may pass before something outside the EE
// would change state the loop could see, measured from the current vblank phase. It is
// the distance to the NEAREST such event; the caller stops a period short of it.
func (m *Machine) idleDeadline(vblAcc uint64) uint64 {
	// The next vertical blank's leading edge.
	D := stepsPerVBlank - vblAcc
	// Its trailing edge, half a field in, if that is still ahead.
	if half := uint64(stepsPerVBlank / 2); vblAcc < half {
		if d := half - vblAcc; d < D {
			D = d
		}
	}
	// The Count/Compare timer, but only when it could actually be taken — otherwise
	// reaching Compare just sets IP7, which SkipInstructions reproduces without help.
	if m.CPU.TimerIRQDeliverable() {
		ticks := uint64(m.CPU.Compare() - m.CPU.Count()) // forward distance, mod 2^32
		if ticks == 0 {
			ticks = 1 << 32
		}
		// Count rises every second instruction; an odd fraction carry means the very next
		// instruction increments it, so the deadline is one instruction nearer.
		steps := ticks * 2
		if m.CPU.CountFrac()&1 != 0 {
			steps--
		}
		if steps < D {
			D = steps
		}
	}
	// A pending IOP reboot is a hard barrier: it must happen on schedule (run.go), so a
	// skip may not step across m.iopRebootAt.
	if m.iopRebootImage != "" {
		if m.steps >= m.iopRebootAt {
			return 0
		}
		if d := m.iopRebootAt - m.steps; d < D {
			D = d
		}
	}
	return D
}

// idleSkip fast-forwards the machine through a proven-idle loop. vblAcc and iopAcc are the
// run loop's own frame- and IOP-clock accumulators, advanced here in lockstep; budget is
// the instructions left in the step budget. It returns the number of instruction-slots
// skipped, which the run loop folds into its own counters, or 0 if it did not skip.
//
// It advances the frame clock, the Count timer and the IOP exactly as the run loop would,
// but does not run the EE — sound only for a whole number of the loop's periods, so the EE
// is left on the very instruction an un-skipped run would reach.
func (m *Machine) idleSkip(vblAcc, iopAcc *uint64, budget uint64) uint64 {
	P := uint64(m.idleDet.period)
	if P == 0 {
		return 0
	}
	// Never step over an interrupt the run loop is about to take.
	if m.CPU.InterruptDeliverable() {
		return 0
	}
	// The IOP veto. If the second processor is streaming data in — a disc read, a SIF
	// transfer — do not skip: the fast-forward would carry the EE past the very instruction
	// the arriving bytes interrupt, and stepping the IOP across the skip cannot reproduce
	// the interleave exactly. Two forms: something is in flight RIGHT NOW (BusyToEE), or the
	// IOP reached into EE memory so recently that a skip started now would likely run into
	// the next one before its deadline. A title screen the IOP touches once in many fields
	// still gets fast-forwarded between; a menu it answers several times a field does not.
	if m.IOP != nil && m.IOP.BusyToEE() {
		return 0
	}

	D := m.idleDeadline(*vblAcc)
	if D <= P {
		return 0 // no room to skip a whole period and still stop short
	}

	// Whole periods only, stopping at least one period short of the deadline.
	N := ((D - 1) / P) * P
	if N > budget {
		N = (budget / P) * P
	}
	if N == 0 {
		return 0
	}

	// Capture the loop's P phase states before freezing the EE — one pass round the loop,
	// then rewound. It is cheap (P is a handful of instructions) and, unlike re-running the
	// loop later, it sees the memory the loop reads BEFORE any disturbance edits it.
	m.captureIdlePhases(P)

	done := uint64(0)
	broke := false
	for done < N {
		// Advance to the next IOP step, no further, so the IOP fires at exactly its
		// 1:iopStepRatio cadence.
		chunk := N - done
		if m.IOP != nil {
			if toIOP := uint64(iopStepRatio) - *iopAcc; toIOP < chunk {
				chunk = toIOP
			}
		}
		stepsIOP := m.IOP != nil && *iopAcc+chunk >= uint64(iopStepRatio)

		if !stepsIOP {
			// No IOP step in this span — pure clock, elide it whole.
			*vblAcc += chunk
			*iopAcc += chunk
			m.CPU.SkipInstructions(chunk)
			m.steps += chunk
			done += chunk
			continue
		}

		// The IOP tick fires at the TOP of a slot, BEFORE that slot's own instruction
		// retires — so the run loop's step clock is one lower than after the slot. Advance
		// every slot up to (not through) the IOP-step slot, so the IOP sees exactly the
		// step count, Count timer and loop phase a serial run would. Getting this wrong by
		// one is invisible for a loop the IOP never touches (VSync), and diverges the moment
		// it does: sifFromIOP and the SIF handler read the step clock and the EE timers.
		*vblAcc += chunk
		*iopAcc = 0 // the tick that reaches iopStepRatio resets it
		m.CPU.SkipInstructions(chunk - 1)
		m.steps += chunk - 1
		done += chunk

		// Place the idle thread where the un-skipped run has it at this instant — `done-1`
		// instructions round the loop — so a SIF handler reads the right context and a
		// preempt parks the right registers.
		m.CPU.SetThreadRegs(m.idleDet.phase[(done-1)%P])
		gen := m.eeDisturbGen
		m.IOP.Step()

		if m.eeDisturbGen != gen {
			// The IOP reached across into the EE. Retire the IOP-step slot's own
			// instruction for real — it belongs to whatever thread is now current (the idle
			// thread still, or the one just switched in) — leaving the machine on a clean
			// slot boundary for the run loop to take over.
			m.CPU.Step()
			m.steps++
			broke = true
			break
		}
		// Undisturbed: that slot's instruction is the idle loop's own; elide it too.
		m.CPU.SkipInstructions(1)
		m.steps++
	}

	// A clean skip elided a whole number of periods; leave the idle thread where a run that
	// interpreted them would be — which for `done` a multiple of P is the loop boundary it
	// started from. (After a disturbance the thread already ran its slot for real above.)
	if !broke {
		m.CPU.SetThreadRegs(m.idleDet.phase[done%P])
	}

	m.idleDet.Skipped += done
	m.idleDet.Hits++
	return done
}

// captureIdlePhases records the core state at each of the P instructions of the idle loop
// and rewinds to where it started, so a later disturbance can be handed the exact phase
// the idle thread should be at without re-interpreting the loop.
func (m *Machine) captureIdlePhases(P uint64) {
	if uint64(cap(m.idleDet.phase)) < P {
		m.idleDet.phase = make([]r5900.State, P)
	}
	m.idleDet.phase = m.idleDet.phase[:P]
	base := m.CPU.Snapshot()
	m.idleDet.phase[0] = base
	for i := uint64(1); i < P; i++ {
		m.CPU.Step()
		m.idleDet.phase[i] = m.CPU.Snapshot()
	}
	m.CPU.Restore(base) // rewind: the clock those steps ticked is undone with the rest
}

// SetIdleSkip turns the idle fast-forward on or off. It is OFF by default (NewMachine sets
// noIdleSkip), and it is opt-in for a reason worth stating plainly.
//
// The skip is byte-identical to a serial run on a field that stands still — a title screen, a
// paused menu, the frame the gate pins ([[ps2-platform]]): TestIdleSkipMatchesSerial and the
// RRV title/intro tests prove it, and it is ~1.3x there. It is NOT byte-identical across disc
// streaming. This machine's IOP is the GameCube's DSP — a second processor that reaches into
// EE memory — but unlike the DSP it never truly sleeps, and its SIF receive channel sits
// armed the whole time a game runs, so there is no cheap "is it quiet" flag to veto on: the
// BusyToEE veto catches a disc read, but not an RPC reply the IOP arms and fires inside one
// step. Left on through a boot, the skip eventually fast-forwards the EE past one of those and
// the whole-boot render pin moves. So a run that must be bit-exact from cold boot — the asset
// pipeline, the regression — leaves it off, and a caller re-rendering a still field for speed
// (a frame debugger, a repeated capture) turns it on knowing the field it is on.
func (m *Machine) SetIdleSkip(on bool) { m.noIdleSkip = !on }

// IdleSkip reports whether the fast-forward is currently on, so a UI that toggles it (the
// frame debugger's checkbox) can draw the box in the right state.
func (m *Machine) IdleSkip() bool { return !m.noIdleSkip }

// IdleStats reports how much of a run was fast-forwarded rather than interpreted: the
// instruction-slots skipped and how many times a loop was proved idle.
func (m *Machine) IdleStats() (skipped, hits uint64) {
	return m.idleDet.Skipped, m.idleDet.Hits
}