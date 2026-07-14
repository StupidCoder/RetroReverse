package ps2

// ioptimer.go is the IOP's clock.
//
// Without it the second processor has no sense of time passing, and the consequence is
// not that things happen at the wrong moment — it is that a whole class of them never
// happen at all. THREADMAN's DelayThread computes a wake-up time from a counter that
// never advances; the thread is put to sleep against a deadline that will never arrive;
// the scheduler finds nothing to run and goes round its idle loop; and the boot, from
// outside, looks like a machine that is working hard and getting nowhere. The profile
// said so first: a quarter of the IOP's time in a 64-bit division routine inside
// THREADMAN, which is what converting microseconds to timer ticks looks like when it is
// being done over and over because the answer never comes true.
//
// Six counters in two banks, and the disc drives two of them:
//
//	4   0x1F801490   mode 0x58 — reset on target, interrupt on target, repeat.
//	                 A periodic tick. TIMEMANI takes interrupt 15 for it.
//	5   0x1F8014A0   mode 0x70 — interrupt on target, no reset. An alarm, re-armed by
//	                 hand each time. THREADMAN takes interrupt 16 for it.
//
// The line each timer raises is derived from that pairing, not assumed: THREADMAN
// allocates a timer through timrman, TIMEMANI programs counter 5's target on its behalf,
// and THREADMAN registers its handler on interrupt 16 — so counter 5 raises 16. TIMEMANI
// keeps counter 4 for itself and registers on 15. Both fit "interrupt = 11 + counter" for
// the upper bank, and the lower bank's 4, 5, 6 follow the same shape one bank down. The
// upper bank is what this disc uses and the upper bank is what is confirmed.
//
// The counter advances once per instruction. That is not a fudge: the IOP is an R3000A at
// about 36.9 MHz retiring roughly one instruction a cycle, and both counters here are
// programmed with clock source 0, which is that same system clock. So one tick per
// instruction is the board's own ratio, near enough — and it is the ratio the modules were
// written against, which is what makes their delays come out at plausible lengths rather
// than merely monotonic ones.

// The two banks of counters.
const (
	iopTimerLoBase = 0x1F801100 // counters 0..2, sixteen bits each
	iopTimerLoEnd  = 0x1F801130
	iopTimerHiBase = 0x1F801480 // counters 3..5, thirty-two bits
	iopTimerHiEnd  = 0x1F8014B0

	iopTimers = 6
)

// A counter's three registers, at base + 0x10*n.
const (
	iopTimerCount  = 0x0
	iopTimerMode   = 0x4
	iopTimerTarget = 0x8
)

// The mode bits this model acts on. The rest are stored and read back.
const (
	iopTimerResetOnTarget = 1 << 3  // go back to zero on reaching the target
	iopTimerIRQOnTarget   = 1 << 4  // raise the line on reaching it
	iopTimerIRQOnOverflow = 1 << 5  // and on wrapping
	iopTimerRepeat        = 1 << 6  // keep doing it, rather than once
	iopTimerHitTarget     = 1 << 11 // set by the hardware, cleared by reading the mode
	iopTimerHitOverflow   = 1 << 12
)

// iopTimer is one counter.
type iopTimer struct {
	count, mode, target uint32
	fired               bool // armed-and-spent: cleared whenever the guest reprograms it
}

// iopTimerIRQ is the interrupt line a counter raises. See the note above for how the
// upper bank's two were established; the lower bank's three follow the same shape and
// nothing on this disc exercises them.
func iopTimerIRQ(n int) uint32 {
	if n < 3 {
		return uint32(4 + n)
	}
	return uint32(11 + n)
}

// iopTimerMax is how far a counter counts before it wraps. The lower bank is sixteen bits
// and the upper bank thirty-two — which is not a detail: THREADMAN programs counter 5
// with a target of 0x12000, and a sixteen-bit counter could not reach it.
func iopTimerMax(n int) uint32 {
	if n < 3 {
		return 0xFFFF
	}
	return 0xFFFFFFFF
}

// iopTimerReg decodes a timer register address.
func iopTimerReg(a uint32) (n int, reg uint32, ok bool) {
	switch {
	case a >= iopTimerLoBase && a < iopTimerLoEnd:
		return int(a-iopTimerLoBase) / 0x10, a & 0xC, true
	case a >= iopTimerHiBase && a < iopTimerHiEnd:
		return 3 + int(a-iopTimerHiBase)/0x10, a & 0xC, true
	}
	return 0, 0, false
}

// timerRead serves a read by the guest.
func (p *IOP) timerRead(a uint32) (uint32, bool) {
	n, reg, ok := iopTimerReg(a)
	if !ok {
		return 0, false
	}
	v, _ := p.timerPeek(a)
	if reg == iopTimerMode {
		// Reading the mode clears the two "it happened" bits. That is how the hardware works
		// and it is how the driver expects to be told: TIMEMANI reads this register to find
		// out *why* it was interrupted, and a bit that did not clear would have it answer the
		// same question the same way forever.
		//
		// But THE CLEAR CANNOT HAPPEN NOW, and this is the whole reason the arming is
		// deferred rather than done here.
		//
		// The R3000A's bus is byte-wide. TIMEMANI reads this register with an `lw`, and an
		// `lw` arrives as four separate reads of the same word — and "reading the mode clears
		// the flags" applied to each of them means the *first* byte's read clears them and the
		// other three see a register in which nothing ever happened. The flags live in bits 11
		// and 12, which is byte 1: the byte that carries the answer is read second, and by
		// then the answer is gone. The mode comes back 0x0070 instead of 0x0870 on the one
		// read that mattered — THREADMAN's alarm handler, which dispatches on exactly that bit
		// — so the handler concluded that the alarm it had just been interrupted for had not
		// gone off, and every thread that ever called DelayThread slept for ever.
		//
		// It is invisible from every other angle. The interrupt is raised, and delivered, and
		// the handler runs, and the counter really did reach its target; the machine looks
		// like one whose threads have simply nothing to do.
		//
		// So the clear settles at the end of the instruction, exactly as the register trace
		// does (see ioTrace): every byte of one load sees the same register, and the flags go
		// out once the load is over.
		p.timerAck |= 1 << uint(n)
	}
	return v, true
}

// timerAckFlush clears the "it happened" bits of any timer whose mode was read during the
// instruction that has just finished. See timerRead.
func (p *IOP) timerAckFlush() {
	for n := 0; p.timerAck != 0 && n < iopTimers; n++ {
		if p.timerAck&(1<<uint(n)) == 0 {
			continue
		}
		p.timers[n].mode &^= iopTimerHitTarget | iopTimerHitOverflow
	}
	p.timerAck = 0
}

// timerPeek reads a timer register without disturbing it.
func (p *IOP) timerPeek(a uint32) (uint32, bool) {
	n, reg, ok := iopTimerReg(a)
	if !ok {
		return 0, false
	}
	t := &p.timers[n]
	switch reg {
	case iopTimerCount:
		return t.count, true
	case iopTimerMode:
		return t.mode, true
	case iopTimerTarget:
		return t.target, true
	}
	return 0, false
}

// timerWrite serves a write.
func (p *IOP) timerWrite(a, v uint32) bool {
	n, reg, ok := iopTimerReg(a)
	if !ok {
		return false
	}
	t := &p.timers[n]
	switch reg {
	case iopTimerCount:
		t.count = v
	case iopTimerMode:
		// Writing the mode restarts the counter. The hardware does this, and the driver
		// relies on it: TIMEMANI sets a target and then writes the mode, and expects the
		// interval to be measured from that moment.
		t.mode = v
		t.count = 0
	case iopTimerTarget:
		t.target = v
	}
	// Any of the three is a re-arming. A counter that has already fired against an old
	// target must be allowed to fire again against the new one — this is what lets
	// THREADMAN's alarm, which has no auto-reset and is re-armed by hand, go off more than
	// once.
	t.fired = false
	return true
}

// timerTick advances every counter by one, and raises the line of any that has arrived.
func (p *IOP) timerTick() {
	for n := range p.timers {
		t := &p.timers[n]
		if t.mode == 0 {
			continue // never programmed: not running
		}
		max := iopTimerMax(n)

		if t.count == max {
			t.count = 0
			t.mode |= iopTimerHitOverflow
			if t.mode&iopTimerIRQOnOverflow != 0 {
				p.raiseIRQ(iopTimerIRQ(n))
			}
			continue
		}
		t.count++

		if t.target == 0 || t.count < t.target {
			continue
		}
		// It has reached the target. Whether that is worth an interrupt, and whether the
		// counter starts again, are two separate bits and the two counters on this disc use
		// them differently: counter 4 resets and re-fires forever, counter 5 does neither
		// and is re-armed by the driver.
		if t.mode&iopTimerResetOnTarget != 0 {
			t.count = 0
			t.fired = false
		}
		if t.fired {
			continue
		}
		t.mode |= iopTimerHitTarget
		if t.mode&iopTimerIRQOnTarget != 0 {
			p.raiseIRQ(iopTimerIRQ(n))
		}
		if t.mode&iopTimerRepeat == 0 || t.mode&iopTimerResetOnTarget == 0 {
			t.fired = true
		}
	}
}
