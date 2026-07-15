package gc

// vi.go is the Video Interface: it reads a framebuffer out of memory and scans it to the
// screen, and once per field it raises the vertical-retrace interrupt that is the whole
// machine's clock.
//
// The retrace interrupt is the heartbeat. A GameCube game does almost nothing in a busy
// loop; it arms the retrace, sleeps, and is woken sixty times a second to advance one
// frame. So a machine whose VI never fires is a machine whose game runs its initialisation
// and then waits forever — which is the most common way a first boot appears to hang, and
// why this device is one of the first that has to work.
//
// The framebuffer VI reads is the XFB — the "external" framebuffer, in main memory, in
// YUV 4:2:2. It is not what the graphics pipe draws into (that is the EFB, on the
// graphics chip); it is where the pipe's finished frame is copied to be displayed. So the
// address in TFBL is the ground truth for "what is on screen", and -shot reads it.

type vi struct {
	DI      [4]uint32 // the display-interrupt registers: scanline, enable (bit 28), pending (bit 31)
	TFBL    uint32    // top field base: the XFB address, plus the addressing-mode bits
	BFBL    uint32    // bottom field base
	DCR     uint32    // display configuration: enable, interlace, format
	Field   uint64    // fields elapsed since reset — the frame counter
	Line    uint32    // the current scanline, for DPV reads
	Counter uint32    // instructions into the current field, for the heartbeat
}

func (v *vi) init() {
	// Nothing to pre-load: the game programs every VI register itself. The zero value is
	// "display off", which is correct out of reset.
}

func (v *vi) read(m *Machine, off uint32, size int) uint32 {
	r := off & 0xFFF
	switch {
	case r == 0x02:
		return v.DCR
	case r == 0x1C || r == 0x1E:
		return halfword(v.TFBL, r, size)
	case r == 0x24 || r == 0x26:
		return halfword(v.BFBL, r, size)
	case r == 0x2C:
		// Display position, vertical: which scanline is being scanned out. A game polls
		// this to sync finely; returning a moving value keeps such a poll from spinning.
		return v.Line
	case r >= 0x30 && r < 0x40:
		// The display-interrupt registers are 32 bits but the game accesses them a halfword
		// at a time: the high halfword (offset 0/4/8/C) holds the enable, the status bit and
		// the vertical position; the low halfword holds the horizontal position.
		full := v.DI[(r-0x30)/4]
		if size == 2 {
			if r&2 == 0 {
				return full >> 16
			}
			return full & 0xFFFF
		}
		return full
	}
	// The many timing registers a game writes and never reads back: report them once, but
	// do not pretend the read means something.
	m.logf("VI read unmodelled 0x%03X", r)
	return 0
}

func (v *vi) write(m *Machine, off uint32, val uint32, size int) {
	r := off & 0xFFF
	switch {
	case r == 0x02:
		v.DCR = val
	case r == 0x1C || r == 0x1E:
		// The framebuffer address is a 32-bit register the game programs a halfword at a time:
		// offset 0x1C is its high half, 0x1E its low half. Composing the whole register from
		// both — rather than taking a halfword store as the entire value — is what keeps the
		// low half of the address, without which the video interface scans out of the wrong
		// place. This is the same halfword hazard the display-interrupt registers carry.
		v.TFBL = composeHalfword(v.TFBL, val, r, size)
	case r == 0x24 || r == 0x26:
		v.BFBL = composeHalfword(v.BFBL, val, r, size)
	case r >= 0x30 && r < 0x40:
		i := (r - 0x30) / 4
		// The display-interrupt registers are 32 bits, but the game programs them a halfword
		// at a time. The high halfword (offset 0/4/8/C into the block) carries the enable
		// (bit 28), the status/pending bit (31) and the vertical position; the low halfword
		// carries the horizontal position. Composing the full register from its halves is
		// what makes the enable bit visible — a whole-word view of a high-halfword write puts
		// the enable at bit 12, where nothing looks for it, and the retrace never arms.
		switch {
		case size == 2 && r&2 == 0:
			v.DI[i] = (v.DI[i] & 0x0000FFFF) | (val&0xFFFF)<<16
		case size == 2:
			v.DI[i] = (v.DI[i] & 0xFFFF0000) | (val & 0xFFFF)
		default:
			v.DI[i] = val
		}
		if v.DI[i]&(1<<31) == 0 {
			// The status bit is clear — either never set, or just acknowledged by the
			// handler. If no display interrupt is still pending, lower the shared VI line.
			m.viRefreshIRQ()
		}
	default:
		// The timing registers, the horizontal scan configuration, the filter
		// coefficients: all written at init and not modelled here.
	}
}

// composeHalfword updates one half of a 32-bit VI register from a halfword store, or the
// whole register from a word store. The even offset within the register's pair (2-aligned) is
// the high half, the odd is the low half, on this big-endian machine.
func composeHalfword(reg, val, r uint32, size int) uint32 {
	if size != 2 {
		return val
	}
	if r&2 == 0 {
		return (reg & 0x0000FFFF) | (val&0xFFFF)<<16
	}
	return (reg & 0xFFFF0000) | (val & 0xFFFF)
}

// halfword reads one half of a 32-bit VI register, or the whole register for a word read.
func halfword(reg, r uint32, size int) uint32 {
	if size != 2 {
		return reg
	}
	if r&2 == 0 {
		return reg >> 16
	}
	return reg & 0xFFFF
}

// XFBAddr resolves the framebuffer address VI is scanning out. The addressing has a
// history: bit 28 of TFBL selects whether the stored value is a byte address or a value
// to be multiplied by 32, a hangover from a time when memory was smaller.
func (v *vi) XFBAddr() uint32 {
	a := v.TFBL & 0x00FFFFFF
	if v.TFBL&0x10000000 != 0 {
		a <<= 5
	}
	return a
}

// tickVI advances the video clock. Called from the run loop, it counts instructions into
// fields, and once per field raises the retrace interrupt the game is waiting on.
//
// The field length in instructions is a modelling choice, not a hardware constant: the
// hardware measures a field in scanline time, and this interpreter measures time in
// instructions. What matters is that the game gets a steady heartbeat it can schedule
// against, and that a savestate resumes to the same phase — which it does, because the
// counter is instruction-paced and part of the state.
func (m *Machine) tickVI() {
	m.vi.Counter++
	if m.vi.Counter < fieldInstructions {
		return
	}
	m.vi.Counter = 0
	m.vi.Field++
	m.vi.Line = 0

	// A field ended. Raise the retrace on every enabled display interrupt, and let the
	// game's handler run. The frame it has just finished is on screen now, so this is the
	// moment to capture it.
	fired := false
	for i := range m.vi.DI {
		if m.vi.DI[i]&(1<<28) != 0 { // enabled
			m.vi.DI[i] |= 1 << 31 // pending
			fired = true
		}
	}
	if fired {
		m.raiseInt(IntVI)
	}
	// The serial interface auto-polls the controllers on the field boundary, latching each
	// pad's state where PADRead will read it — the same VBlank cadence real hardware uses.
	m.tickSI()
	if m.OnDisplay != nil {
		m.OnDisplay(m)
	}
}

// viRefreshIRQ lowers the VI line when no display interrupt is still pending.
func (m *Machine) viRefreshIRQ() {
	for _, d := range m.vi.DI {
		if d&(1<<28) != 0 && d&(1<<31) != 0 {
			return // one is still pending
		}
	}
	m.clearInt(IntVI)
}

// fieldInstructions is how many instructions make one video field. A GameCube runs at
// ~486 MHz and displays ~60 fields a second, so a field is roughly eight million
// instructions — but the interpreter does not retire one per cycle, so this is tuned for
// a heartbeat that arrives often enough to make progress without drowning the run in
// handler entries. It is instruction-paced, so it is deterministic.
const fieldInstructions = 2_000_000
