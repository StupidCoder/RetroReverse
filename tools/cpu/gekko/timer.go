package gekko

// timer.go is the time base and the decrementer — the processor's two clocks, and the
// only things in this core that advance without an instruction asking them to.
//
// Both are driven by the bus clock divided by four, which on a GameCube is 162 MHz / 4 =
// 40.5 MHz. They are paced here by the instruction count rather than by wall time, for
// the reason every core in this repository paces its timers that way: a run must be
// reproducible. A savestate taken at instruction N and resumed must reach instruction
// N+M in exactly the state an uninterrupted run would have, and a clock that consulted
// the host's time of day would make that false.
//
// The time base is a 64-bit up-counter that software reads with mftb and uses as a
// monotonic clock. The decrementer is a 32-bit down-counter that raises an exception when
// it passes below zero, and it is how a GameCube program schedules anything — the
// alarm queue, the thread quantum, every timeout in the game.

// The Gekko runs at 486 MHz against a 162 MHz bus, and the timers tick at a quarter of
// the bus clock. So one timer tick is twelve core clocks — and, since this interpreter
// retires roughly one instruction per core clock, twelve instructions.
const (
	CoreClock     = 486_000_000
	BusClock      = 162_000_000
	TimerClock    = BusClock / 4 // 40.5 MHz
	ClocksPerTick = CoreClock / TimerClock
)

// tick advances the clocks by one instruction's worth of time.
//
// The fractional accumulator is what keeps the ratio honest: twelve instructions per
// timer tick is exact here, but stating it as an accumulator rather than a counter means
// a different clock ratio does not silently round to zero.
func (c *CPU) tick(cycles int) {
	c.clockFrac += uint32(cycles)
	for c.clockFrac >= ClocksPerTick {
		c.clockFrac -= ClocksPerTick
		c.TB++
		// The decrementer counts down, and the exception fires on the transition into
		// negative — not on every instruction while it is negative. decArmed is that
		// edge: software rearms the timer by writing DEC, and until it does, one
		// underflow raises one exception.
		prev := c.DEC
		c.DEC--
		if prev&0x80000000 == 0 && c.DEC&0x80000000 != 0 {
			c.decArmed = true
		}
	}
}

// setDEC writes the decrementer, which rearms it.
func (c *CPU) setDEC(v uint32) {
	c.DEC = v
	c.decArmed = false
}

// InstrsToDecUnderflow is how many instructions may pass before the decrementer would cross
// below zero — the moment tick above latches decArmed and the machine owes an exception.
//
// It exists for the idle skip (see the platform's idle.go), which needs to know every way the
// machine can change while the program is going round a loop that changes nothing. The
// decrementer is one of them, and the least obvious: it is the only clock here that raises an
// exception on its own, with no device involved.
//
// A decrementer that has ALREADY underflowed and not been rearmed owes nothing further — the
// latch is set, the exception is the CPU's business, and counting on down cannot latch it
// twice. That returns "never", because from the skip's point of view nothing more will happen.
func (c *CPU) InstrsToDecUnderflow() uint64 {
	if c.decArmed || c.DEC&0x80000000 != 0 {
		return ^uint64(0)
	}
	// DEC decrements once every ClocksPerTick instructions, and the underflow is the step
	// from 0 to 0xFFFFFFFF — so it takes DEC+1 more decrements to get there.
	ticks := uint64(c.DEC) + 1
	return ticks*ClocksPerTick - uint64(c.clockFrac)
}

// SkipInstructions advances the clocks as though n instructions had been retired, without
// retiring them.
//
// The caller must have established that nothing in those n instructions could change the
// machine — that is the idle skip's whole proof obligation, not this function's — and must
// not skip past a decrementer underflow (InstrsToDecUnderflow says where that is), because
// this deliberately does not latch decArmed: a skip that could hide an exception would be
// changing the trajectory rather than fast-forwarding it.
func (c *CPU) SkipInstructions(n uint64) {
	c.Steps += n
	total := uint64(c.clockFrac) + n
	ticks := total / ClocksPerTick
	c.clockFrac = uint32(total % ClocksPerTick)
	c.TB += ticks
	c.DEC -= uint32(ticks)
}
