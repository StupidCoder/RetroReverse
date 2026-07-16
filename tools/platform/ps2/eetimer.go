package ps2

// eetimer.go is the EE's four timers, the machine's only readable clock besides the
// vblank. Until they existed, a timer read fell into the unmodelled-io tally and
// returned the last write — and that fiction became the game's day: display-frame-start
// computes its frame ratio as count/ticks-per-frame+1, the boot's stale reads pumped
// the GOAL clock ~60x fast, and by the title the engine believed a quarter of an hour
// had passed — the time-of-day system put the sun at noon and painted the night title
// in daylight.
//
// Each timer is four registers, 0x800 apart per timer from 0x10000000:
//
//	+0x00 COUNT  16 bits, writing sets it
//	+0x10 MODE   bits 0-1 CLKS (0 BUSCLK, 1 /16, 2 /256, 3 HBLANK), bit 6 ZRET
//	             (clear on reaching COMP), bit 7 CUE (count enable), bits 8/9
//	             compare/overflow interrupt enables, bits 10/11 the flags (W1C)
//	+0x20 COMP   the compare value
//	+0x30 HOLD   T0/T1 only: COUNT latched on SBUS interrupt (nothing raises it here)
//
// The counter is not ticked; it is computed. The machine's clock is its step counter
// (a field is stepsPerVBlank steps by construction), so COUNT is base + elapsed steps
// scaled by the clock select — evaluated on read, exact at every moment, and free.
// BUSCLK is 147.456 MHz and a NTSC field 1/59.94 s, so one field is 2,460,180 BUSCLK
// ticks; the /256 rate makes that 9,610 per field, which is why the engine's
// *ticks-per-frame* constant is 0x2625 = 9765: one healthy field's count divides to
// zero and the ratio is 1. Timer interrupts are not delivered (nothing here asks for
// them yet); a MODE write that enables them is noted so the day they matter is loud.

const (
	eeTimerBase = 0x10000000
	eeTimerEnd  = 0x10001830
)

// busclkPerField is one NTSC field of BUSCLK ticks: 147,456,000 / 59.94.
const busclkPerField = 2460180

// eeTimer is one timer's state. count is evaluated lazily against m.steps.
type eeTimer struct {
	base      uint32 // COUNT at baseSteps (16 bits used)
	baseSteps uint64 // when base was set
	mode      uint32
	comp      uint32
	hold      uint32
}

// rate returns the timer's ticks-per-field numerator for its clock select; the
// denominator is stepsPerVBlank.
func (t *eeTimer) rate() uint64 {
	switch t.mode & 3 {
	case 0:
		return busclkPerField
	case 1:
		return busclkPerField / 16
	case 2:
		return busclkPerField / 256
	default:
		return 263 // HBLANK: 262.5 lines a field, and a counter has no halves
	}
}

// countAt evaluates COUNT at the given step count.
func (t *eeTimer) countAt(steps uint64) uint32 {
	if t.mode&0x80 == 0 { // CUE clear: the counter holds
		return t.base & 0xFFFF
	}
	elapsed := (steps - t.baseSteps) * t.rate() / stepsPerVBlank
	c := uint64(t.base) + elapsed
	if t.mode&0x40 != 0 && t.comp > 0 { // ZRET: clear on reaching COMP
		return uint32(c % uint64(t.comp))
	}
	return uint32(c & 0xFFFF)
}

// rebase files the current count as the new base, so a MODE change (clock select,
// CUE) keeps continuity.
func (t *eeTimer) rebase(steps uint64) {
	t.base = t.countAt(steps)
	t.baseSteps = steps
}

// eeTimerAt maps an address to its timer, or -1.
func eeTimerAt(p uint32) (n int, reg uint32) {
	if p < eeTimerBase || p >= eeTimerEnd {
		return -1, 0
	}
	off := p - eeTimerBase
	if off&0x7FF >= 0x40 || off&0xF != 0 {
		return -1, 0
	}
	return int(off >> 11), off & 0x3F
}

func (m *Machine) eeTimerRead(p uint32) (uint32, bool) {
	n, reg := eeTimerAt(p)
	if n < 0 {
		return 0, false
	}
	t := &m.eeTimers[n]
	switch reg {
	case 0x00:
		return t.countAt(m.steps), true
	case 0x10:
		return t.mode, true
	case 0x20:
		return t.comp, true
	case 0x30:
		return t.hold, true
	}
	return 0, false
}

func (m *Machine) eeTimerWrite(p, v uint32) bool {
	n, reg := eeTimerAt(p)
	if n < 0 {
		return false
	}
	t := &m.eeTimers[n]
	switch reg {
	case 0x00:
		t.base = v & 0xFFFF
		t.baseSteps = m.steps
	case 0x10:
		t.rebase(m.steps)
		// Bits 10/11 are the compare/overflow flags, write-one-to-clear; they are
		// never set here (no interrupt delivery), so clearing is dropping the bits.
		t.mode = v &^ 0xC00
		if v&0x300 != 0 {
			m.note("EE timer %d MODE 0x%X enables interrupts — not delivered", n, v)
		}
	case 0x20:
		t.comp = v & 0xFFFF
	case 0x30:
		t.hold = v & 0xFFFF
	}
	return true
}
