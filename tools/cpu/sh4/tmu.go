package sh4

// tmu.go is the on-chip timer unit: three identical down-counters clocked
// from the peripheral clock through a per-channel prescaler, each able to
// raise an interrupt on underflow. Katana's timebase runs on TMU0, so this is
// the first on-chip device a game actually needs.
//
// Pacing is by instruction count, the way the other cores pace their timers:
// this model retires roughly one instruction per CPU clock, and on the
// Dreamcast the peripheral clock is a quarter of the 200 MHz core clock, so a
// channel ticks every 4·prescale instructions. The fraction lives in Frac so
// a savestate carries it.

// TCR bits.
const (
	tmuUNF  = 1 << 8 // underflow happened
	tmuUNIE = 1 << 5 // underflow interrupts enabled
)

// TMUChannel is one counter. Fields are exported so gob carries them.
type TMUChannel struct {
	TCOR uint32 // reload constant
	TCNT uint32 // the live counter
	TCR  uint16
	Frac uint32 // instructions accumulated toward the next tick
}

// TMUState is the whole unit.
type TMUState struct {
	TOCR uint8
	TSTR uint8 // bits 0-2 start channels 0-2
	Ch   [3]TMUChannel
}

// tmuPeriod is the instructions per TCNT decrement for a channel's prescaler.
// The reserved and external-clock selections have no meaningful pace here;
// they count at the slowest defined rate rather than silently not at all, and
// the census (via Gaps) is the honest record if a game ever selects them.
func tmuPeriod(tcr uint16) uint32 {
	switch tcr & 7 {
	case 0:
		return 4 * 4
	case 1:
		return 4 * 16
	case 2:
		return 4 * 64
	case 3:
		return 4 * 256
	default:
		return 4 * 1024
	}
}

// tickTMU advances every running channel by one instruction.
func (c *CPU) tickTMU() {
	ts := c.TMU.TSTR & 7
	if ts == 0 {
		return
	}
	for i := 0; i < 3; i++ {
		if ts&(1<<i) == 0 {
			continue
		}
		ch := &c.TMU.Ch[i]
		ch.Frac++
		if ch.Frac < tmuPeriod(ch.TCR) {
			continue
		}
		ch.Frac = 0
		if ch.TCNT == 0 {
			ch.TCNT = ch.TCOR
			ch.TCR |= tmuUNF
		} else {
			ch.TCNT--
		}
	}
}

// tmuPending reports the highest-priority channel with an underflow both
// flagged and enabled: its IPRA priority level and INTEVT code.
func (c *CPU) tmuPending() (level, code uint32, ok bool) {
	shift := [3]uint32{12, 8, 4} // IPRA nibbles: TMU0 high, then TMU1, TMU2
	for i := 0; i < 3; i++ {
		ch := &c.TMU.Ch[i]
		if ch.TCR&tmuUNF == 0 || ch.TCR&tmuUNIE == 0 {
			continue
		}
		if l := (c.IPRA >> shift[i]) & 0xF; l > level {
			level, code, ok = l, 0x400+0x20*uint32(i), true
		}
	}
	return
}

// tmuRead and tmuWrite serve the register block at FFD80000.
func (c *CPU) tmuRead(addr uint32) (uint32, bool) {
	switch addr {
	case 0xFFD80000:
		return uint32(c.TMU.TOCR), true
	case 0xFFD80004:
		return uint32(c.TMU.TSTR), true
	}
	for i := uint32(0); i < 3; i++ {
		base := 0xFFD80008 + i*12
		switch addr {
		case base:
			return c.TMU.Ch[i].TCOR, true
		case base + 4:
			return c.TMU.Ch[i].TCNT, true
		case base + 8:
			return uint32(c.TMU.Ch[i].TCR), true
		}
	}
	return 0, false
}

func (c *CPU) tmuWrite(addr, v uint32) bool {
	switch addr {
	case 0xFFD80000:
		c.TMU.TOCR = uint8(v)
		return true
	case 0xFFD80004:
		c.TMU.TSTR = uint8(v)
		return true
	}
	for i := uint32(0); i < 3; i++ {
		base := 0xFFD80008 + i*12
		switch addr {
		case base:
			c.TMU.Ch[i].TCOR = v
			return true
		case base + 4:
			c.TMU.Ch[i].TCNT = v
			return true
		case base + 8:
			// UNF is a flag: software clears it by writing 0, and writing 1
			// cannot set it.
			old := c.TMU.Ch[i].TCR
			c.TMU.Ch[i].TCR = uint16(v)&^tmuUNF | old&uint16(v)&tmuUNF
			return true
		}
	}
	return false
}
