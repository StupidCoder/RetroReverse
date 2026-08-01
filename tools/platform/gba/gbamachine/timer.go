package gbamachine

// The four timers. Each counts up from its reload at the system clock over a
// prescaler (1, 64, 256, 1024), or — in cascade mode — once per overflow of the
// timer below it. Overflow raises the timer's interrupt if enabled. The
// scheduler ticks them once per scanline with the line's worth of cycles;
// that granularity is nominal-timing, consistent with the rest of the model
// (the sound FIFO timers will want finer treatment when sound exists).

type timer struct {
	counter uint16
	reload  uint16
	ctrl    uint16
	frac    int // cycles not yet converted into ticks
}

var timerPrescale = [4]int{1, 64, 256, 1024}

// timerRegWrite services a store to the 0x100-0x10E block.
func (m *Machine) timerRegWrite(reg uint32, v uint16) {
	n := int(reg-0x100) / 4
	t := &m.timers[n]
	if reg&2 == 0 { // TMxCNT_L: the RELOAD value, not the counter
		t.reload = v
		return
	}
	was := t.ctrl
	t.ctrl = v
	if v&(1<<7) != 0 && was&(1<<7) == 0 { // rising start edge: load the counter
		t.counter = t.reload
		t.frac = 0
	}
}

// tickTimers advances the chain by one scanline's cycles.
func (m *Machine) tickTimers(cycles int) {
	overflowBelow := false
	for n := range m.timers {
		t := &m.timers[n]
		if t.ctrl&(1<<7) == 0 {
			overflowBelow = false
			continue
		}
		ticks := 0
		if n > 0 && t.ctrl&(1<<2) != 0 { // cascade: count the timer below's overflows
			if overflowBelow {
				ticks = 1
			}
		} else {
			t.frac += cycles
			p := timerPrescale[t.ctrl&3]
			ticks = t.frac / p
			t.frac %= p
		}
		overflowBelow = false
		overflows := 0
		for ; ticks > 0; ticks-- {
			t.counter++
			if t.counter == 0 {
				t.counter = t.reload
				overflowBelow = true
				overflows++
				if t.ctrl&(1<<6) != 0 {
					m.raise(uint16(irqTimer0) << uint(n))
				}
			}
		}
		// Timers 0 and 1 are also the Direct Sound sample clock: every overflow
		// pops one sample out of whichever FIFO is bound to this timer.
		if overflows > 0 && n < 2 {
			m.fifoTimerOverflow(n, overflows)
		}
	}
}
