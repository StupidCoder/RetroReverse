package gc

// idle.go fast-forwards the clock through a loop that cannot do anything.
//
// FOUR FIFTHS OF THIS MACHINE'S INSTRUCTIONS ARE THREE ADDRESSES. A PC histogram over the
// intro cutscene puts 79.7% of every instruction retired at 0x801E0864..0x801E086C, and the
// shadow scene agrees at 76.9%:
//
//	801E0864  lwz    r0,5784(r13)   ; load a flag out of the small-data area
//	801E0868  cmplwi cr0,r0,0       ; is it still zero?
//	801E086C  beq    0x801E0864     ; then go round again
//
// That is the operating system's idle loop, and it is not a gap in the model — it is the game
// waiting, correctly, for an interrupt handler to set a flag. A GameCube retires ~8.1M
// instructions a field and this interpreter gives the game 2M (vi.go's fieldInstructions is a
// modelling choice, not a hardware constant), so the game finishes its work early and spends
// the rest of every field here. The emulator was spending four fifths of its time proving that
// zero is still zero.
//
// WHAT MAKES THE SKIP EXACT, and it is worth being precise because "idle loop detection" is
// usually a heuristic and this one is not:
//
// If the machine returns to a state it has already been in — the same PC, the same registers,
// the same everything — having performed no stores, then it is a function with no inputs that
// has repeated itself. It will go on repeating, identically, forever, UNTIL SOMETHING OUTSIDE
// THE PROCESSOR CHANGES. Running that loop a million more times leaves exactly the state that
// running it once more leaves. So the skip does not approximate the loop; it computes its
// limit.
//
// The whole risk therefore sits in one question: what can change while the CPU is not looking?
// On this machine the complete list is the ticks, and every one of them is instruction-paced,
// which is what makes the deadline computable rather than guessable:
//
//	the video field boundary      raises the retrace interrupt
//	the audio DMA's next block    raises the AID interrupt
//	an in-flight disc transfer    WRITES MAIN MEMORY and raises an interrupt
//	the decrementer underflowing  raises an exception with no device involved at all
//	the DSP, if it is running     writes memory and mailboxes behind the CPU's back
//
// The first four are deadlines: the skip goes to the nearest and stops. The fifth is not a
// deadline but a veto — a running DSP is stepped 64 instructions at a time by tickDSP, and
// skipping the Gekko would starve it — so a draw with the audio core awake does not skip at
// all. It is asleep 99.8% of the time (3,644 batches against two million ticks a field).
//
// So the machine after a skip is the machine an uninterrupted run would have reached, with one
// deliberate exception: it retired fewer instructions to get there. Instrs and CPU.Steps are
// the emulator's bookkeeping, not the console's condition; TB and DEC are the console's, and
// they are advanced exactly as if the instructions had run, because the game reads them.

import "math"

// idleSnap is the processor state a loop must return to, unchanged, to prove it is idle.
//
// Floating point is carried by BITS rather than by value, and that is not pedantry: a NaN
// compares unequal to itself, so a register file holding one — which is ordinary — would make
// every snapshot look different from every other and the skip would silently never fire.
type idleSnap struct {
	PC, LR, CTR, CR, XER, MSR uint32
	GPR                       [32]uint32
	FPR                       [32][2]uint64
}

// idleState is the search for a state repeat.
type idleState struct {
	armed  bool
	snap   idleSnap
	stores uint64 // the store count when the snapshot was taken
	insns  int    // instructions watched since arming
	period int    // the proved loop length, in instructions
	next   uint64 // the instruction count at which to arm again

	Skipped uint64 // instructions fast-forwarded — diagnostic, and the thing to look at
	Hits    uint64 // how many times a loop was proved idle
}

// How the detector paces itself. Arming costs a snapshot, and watching costs a compare of one
// word per instruction, so neither is done all the time: the machine arms once every
// idleArmEvery instructions and gives up after idleWatch. A loop of three instructions repeats
// well inside that; a loop that does not repeat within sixty-four instructions is doing enough
// work that a skip would not have paid anyway.
const (
	idleArmEvery = 4096
	idleWatch    = 64
)

// snapshotCPU takes the processor's whole architectural state.
func (m *Machine) snapshotCPU() idleSnap {
	c := m.CPU
	s := idleSnap{
		PC: c.PC, LR: c.LR, CTR: c.CTR, CR: c.CR, XER: c.XER, MSR: c.MSR,
		GPR: c.GPR,
	}
	for i := range c.FPR {
		s.FPR[i][0] = math.Float64bits(c.FPR[i].PS0)
		s.FPR[i][1] = math.Float64bits(c.FPR[i].PS1)
	}
	return s
}

// idleStep is the detector, and it runs on the hot path — one compare per instruction when it
// is not armed, one more when it is.
//
// It reports whether the machine has just been proved idle at pc.
func (m *Machine) idleStep(pc uint32) bool {
	if !m.idle.armed {
		if m.Instrs < m.idle.next {
			return false
		}
		m.idle.armed = true
		m.idle.insns = 0
		m.idle.stores = m.stores
		m.idle.snap = m.snapshotCPU()
		return false
	}

	m.idle.insns++
	if m.idle.insns > idleWatch {
		m.idle.armed = false
		m.idle.next = m.Instrs + idleArmEvery
		return false
	}
	// The cheap tests first: the same address, and nothing stored on the way round. Only then
	// is the whole register file worth comparing.
	if pc != m.idle.snap.PC || m.stores != m.idle.stores {
		return false
	}
	if m.snapshotCPU() != m.idle.snap {
		return false
	}
	// The instructions it took to come back around ARE the loop's period, and that number is
	// what makes the skip exact rather than merely equivalent. See idleSkip.
	m.idle.period = m.idle.insns
	m.idle.armed = false
	m.idle.next = m.Instrs + idleArmEvery
	return true
}

// idleDeadline is how many instructions may be skipped before something outside the processor
// would do something. Zero means "do not skip".
func (m *Machine) idleDeadline() uint64 {
	// A running audio core is a veto, not a deadline: tickDSP steps it 64 instructions at a
	// time off the Gekko's clock, so skipping the Gekko would stop the DSP dead.
	d := &m.dsp
	if d.Core != nil && !d.CoreHalt && !d.CoreBlocked && !d.Core.Halted {
		return 0
	}

	n := uint64(fieldInstructions - m.vi.Counter) // the next video field

	if d.AIDControl&0x8000 != 0 && d.AIDControl&0x7FFF != 0 {
		if left := uint64(aidInstrPerBlock) - d.AIDAccum; left < n {
			n = left
		}
	}
	if m.di.BusyInstr > 0 && uint64(m.di.BusyInstr) < n {
		n = uint64(m.di.BusyInstr)
	}
	if dec := m.CPU.InstrsToDecUnderflow(); dec < n {
		n = dec
	}

	// Stop one instruction short, so the tick that fires the event is executed by the ordinary
	// run loop rather than reimplemented here. The skip's job is to reach the edge of the next
	// event, not to deliver it.
	if n == 0 {
		return 0
	}
	return n - 1
}

// idleSkip fast-forwards n instructions' worth of clock without retiring them.
//
// IT SKIPS A WHOLE NUMBER OF LOOP ITERATIONS, and that is the difference between a skip that is
// equivalent and a skip that is exact. The loop returns to the same PC every period
// instructions, so advancing a multiple of the period leaves the processor exactly where the
// unskipped run would have left it — same PC, same registers, same instruction count. The
// remainder, always fewer instructions than the period, is left for the run loop to interpret
// normally.
//
// Skipping the raw deadline instead is what the first version of this did, and it cost two
// bytes of RAM: the interrupt then always landed on the PC the detector happened to snapshot,
// rather than wherever the loop had got to, so SRR0 differed, the loop took a couple of
// instructions longer to leave, and THE GAME'S OWN FRAME-TIME COUNTER — it reads the time base
// at 0x801A14E0 and rings the delta into a buffer — came out one tick different. The picture
// was identical and the machine was arguably just as correct; it simply was not the same
// machine, which for an oracle is the only question.
func (m *Machine) idleSkip(n uint64) {
	if m.idle.period <= 0 {
		return
	}
	n -= n % uint64(m.idle.period)
	if n == 0 {
		return
	}

	m.vi.Counter += uint32(n)
	if d := &m.dsp; d.AIDControl&0x8000 != 0 && d.AIDControl&0x7FFF != 0 {
		d.AIDAccum += n
	}
	if m.di.BusyInstr > 0 {
		m.di.BusyInstr -= int64(n)
	}
	m.CPU.SkipInstructions(n)
	m.Instrs += n
	m.idle.Skipped += n
	m.idle.Hits++
}

// IdleStats reports how much of a run was fast-forwarded rather than interpreted.
func (m *Machine) IdleStats() (skipped, hits uint64) { return m.idle.Skipped, m.idle.Hits }

// SetIdleSkip turns the fast-forward on or off. It is on by default; turning it off is how a
// run proves the skip changed nothing, and how a bisect rules it out.
func (m *Machine) SetIdleSkip(on bool) { m.noIdle = !on }
