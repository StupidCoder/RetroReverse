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
