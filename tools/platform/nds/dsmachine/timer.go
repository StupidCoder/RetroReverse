package dsmachine

// The four timers each core owns. A timer counts up at a prescaled fraction of the
// system clock; when it wraps past 0xFFFF it reloads from its latched reload value
// and, if asked, raises an interrupt — and if the *next* timer up is in cascade
// mode, it ticks that one instead of the clock.
//
// Timers are driven here from the scanline scheduler rather than from an
// instruction count, because that is what they measure on hardware: a DS timer is
// a clock, and code that uses one to pace itself (an audio mixer's frame, a
// tick-based scheduler) is measuring the *video clock*, not how fast we interpret.

// timerPrescale is the number of system cycles per tick, by the two-bit
// prescaler field.
var timerPrescale = [4]int{1, 64, 256, 1024}

type timer struct {
	counter uint16 // the live count
	reload  uint16 // reloaded on overflow
	ctrl    uint16 // TMxCNT_H
	frac    int    // system cycles carried over between updates
}

func (t *timer) enabled() bool { return t.ctrl&0x0080 != 0 }
func (t *timer) cascade() bool { return t.ctrl&0x0004 != 0 }
func (t *timer) irqOn() bool   { return t.ctrl&0x0040 != 0 }

// tickTimers advances a core's four timers by the system cycles one scanline takes.
// It returns nothing: overflow effects (interrupt, cascade) are applied in place.
func (c *core) tickTimers(cycles int) {
	var carry int // overflows produced by timer n, for timer n+1 to consume in cascade
	for n := range c.timers {
		t := &c.timers[n]
		if !t.enabled() {
			carry = 0
			continue
		}
		var ticks int
		if n > 0 && t.cascade() {
			// A cascading timer ignores the clock entirely and counts the overflows of
			// the timer below it.
			ticks = carry
		} else {
			p := timerPrescale[t.ctrl&3]
			t.frac += cycles
			ticks = t.frac / p
			t.frac -= ticks * p
		}
		carry = 0
		for i := 0; i < ticks; i++ {
			if t.counter == 0xFFFF {
				t.counter = t.reload
				carry++
				if t.irqOn() {
					c.raise(irqTimer0 << uint(n))
				}
			} else {
				t.counter++
			}
		}
	}
}

// writeTimerCtrl handles a write to TMxCNT_H. The reload value is latched into the
// counter on the enable *edge* — a running timer whose control is rewritten keeps
// counting from where it is, which is how software adjusts a timer's interrupt
// without losing its phase.
func (c *core) writeTimerCtrl(n int, v uint16) {
	t := &c.timers[n]
	wasOn := t.enabled()
	t.ctrl = v
	if !wasOn && t.enabled() {
		t.counter = t.reload
		t.frac = 0
	}
}
