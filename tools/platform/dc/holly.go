package dc

// holly.go is the system block's interrupt fan-in: three status registers
// (normal, external, error) gated by three mask sets, one per SH-4 interrupt
// level. A bit pending under a level-6 mask asserts IRL level 13, level 4
// masks assert 11, level 2 masks assert 9 — the KOS-documented routing — and
// the INTEVT codes are the SH7750's independent-line encodings (0x3A0, 0x360,
// 0x320). The VBlank heartbeat is ISTNRM bit 3, raised by the run loop every
// field.

// Holly is the interrupt state of the system block.
type Holly struct {
	ISTNRM, ISTEXT, ISTERR    uint32
	IML2NRM, IML2EXT, IML2ERR uint32
	IML4NRM, IML4EXT, IML4ERR uint32
	IML6NRM, IML6EXT, IML6ERR uint32
}

const (
	istRenderDone = 7 << 0 // video, ISP, TSP — raised together
	istVBlankIn   = 1 << 3
	istVBlankOut  = 1 << 4
	istCh2DMA     = 1 << 19
)

// taListDoneBit maps a TA list type (a global parameter's bits 24-26) to its
// list-complete interrupt: opaque, opaque-modifier, translucent,
// translucent-modifier, punch-through.
var taListDoneBit = [5]uint32{1 << 7, 1 << 8, 1 << 9, 1 << 10, 1 << 21}

// hollyRead serves the system block.
func (m *Machine) hollyRead(addr uint32) uint32 {
	h := &m.Holly
	switch addr {
	case 0x005F6900:
		// The status read folds the other two registers' any-pending into the
		// top bits, so a handler can dispatch off one read.
		v := h.ISTNRM
		if h.ISTEXT != 0 {
			v |= 1 << 30
		}
		if h.ISTERR != 0 {
			v |= 1 << 31
		}
		return v
	case 0x005F6904:
		return h.ISTEXT
	case 0x005F6908:
		return h.ISTERR
	case 0x005F6910:
		return h.IML2NRM
	case 0x005F6914:
		return h.IML2EXT
	case 0x005F6918:
		return h.IML2ERR
	case 0x005F6920:
		return h.IML4NRM
	case 0x005F6924:
		return h.IML4EXT
	case 0x005F6928:
		return h.IML4ERR
	case 0x005F6930:
		return h.IML6NRM
	case 0x005F6934:
		return h.IML6EXT
	case 0x005F6938:
		return h.IML6ERR
	case 0x005F688C: // SB_FFST: every write FIFO empty
		return 0
	case 0x005F689C: // SB_SBREV
		return 0x0B
	case 0x005F6800: // SB_C2DSTAT
		return m.C2DStat
	case 0x005F6804: // SB_C2DLEN
		return m.C2DLen
	case 0x005F6808: // SB_C2DST: the completed-instantly model reads idle
		return 0
	}
	m.logf("SB read %08X (PC %08X)", addr, m.CPU.CurPC())
	return 0
}

func (m *Machine) hollyWrite(addr, v uint32) {
	h := &m.Holly
	switch addr {
	case 0x005F6900: // write-1-to-clear
		h.ISTNRM &^= v
	case 0x005F6904:
		// External status clears at the source, not here.
	case 0x005F6908:
		h.ISTERR &^= v
	case 0x005F6910:
		h.IML2NRM = v
	case 0x005F6914:
		h.IML2EXT = v
	case 0x005F6918:
		h.IML2ERR = v
	case 0x005F6920:
		h.IML4NRM = v
	case 0x005F6924:
		h.IML4EXT = v
	case 0x005F6928:
		h.IML4ERR = v
	case 0x005F6930:
		h.IML6NRM = v
	case 0x005F6934:
		h.IML6EXT = v
	case 0x005F6938:
		h.IML6ERR = v
	case 0x005F6800:
		m.C2DStat = v
	case 0x005F6804:
		m.C2DLen = v
	case 0x005F6808:
		// Ch2 (PVR) DMA start. The TA is a later milestone, so nothing is
		// rendered — but the parameter stream is real and its END_OF_LIST
		// markers are scanned (the source address sits in the SH-4 DMAC's
		// SAR2, which the CPU raw-stores), so the list-complete interrupts a
		// frame loop waits on fire for exactly the lists the game closed.
		if v&1 != 0 {
			m.scanTAStream(m.CPU.OnchipReg(0xFFA00020), m.C2DLen)
			m.TAWrites += uint64(m.C2DLen / 4)
			// Completion (the DMA-end interrupt and any END_OF_LIST bits the
			// scan collected) is deferred: a transfer that finishes inside
			// its own start-write finishes before the guest can wait for it.
			m.C2DCountdown = 2000 + m.C2DLen/16
			m.C2DLen = 0
		}
	default:
		m.logf("SB write %08X = %08X (PC %08X)", addr, v, m.CPU.CurPC())
		return
	}
	m.updateIRL()
}

// scanTAStream walks a submitted TA parameter stream (32-byte parameters,
// control word first) tracking the current list type from each global
// parameter and raising the matching list-complete bit at each END_OF_LIST.
func (m *Machine) scanTAStream(src, byteLen uint32) {
	list := uint32(0)
	for off := uint32(0); off+32 <= byteLen; off += 32 {
		pcw := m.ram32(src + off)
		switch pcw >> 29 {
		case 0: // END_OF_LIST
			if list < 5 {
				m.C2DPendingBits |= taListDoneBit[list]
			}
		case 4, 5, 6: // polygon / sprite headers name their list
			list = pcw >> 24 & 7
		}
	}
}

// tickCompletions retires the deferred render and DMA countdowns.
func (m *Machine) tickCompletions() {
	if m.RenderCountdown > 0 {
		if m.RenderCountdown--; m.RenderCountdown == 0 {
			m.raiseNRM(istRenderDone)
		}
	}
	if m.C2DCountdown > 0 {
		if m.C2DCountdown--; m.C2DCountdown == 0 {
			m.raiseNRM(istCh2DMA | m.C2DPendingBits)
			m.C2DPendingBits = 0
		}
	}
}

// raiseNRM sets ISTNRM bits and re-evaluates the interrupt line.
func (m *Machine) raiseNRM(bits uint32) {
	m.Holly.ISTNRM |= bits
	m.updateIRL()
}

// updateIRL asserts the highest Holly level with a pending, unmasked bit.
//
// The IML register names carry the SH-4 *level*: IML6 masks raise level 6,
// whose IRL pin value is 9 and whose INTEVT is therefore 0x320
// (0x200+0x20·(15−L)) — the code a game's dispatcher tests for its VBlank.
// Getting this upside-down (level 13 for IML6) made Crazy Taxi's handler
// dispatch an INTEVT it does not own, return without acknowledging, and
// re-enter forever.
func (m *Machine) updateIRL() {
	h := &m.Holly
	switch {
	case h.ISTNRM&h.IML6NRM != 0 || h.ISTEXT&h.IML6EXT != 0 || h.ISTERR&h.IML6ERR != 0:
		m.CPU.SetIRL(6, 0x320)
	case h.ISTNRM&h.IML4NRM != 0 || h.ISTEXT&h.IML4EXT != 0 || h.ISTERR&h.IML4ERR != 0:
		m.CPU.SetIRL(4, 0x360)
	case h.ISTNRM&h.IML2NRM != 0 || h.ISTEXT&h.IML2EXT != 0 || h.ISTERR&h.IML2ERR != 0:
		m.CPU.SetIRL(2, 0x3A0)
	default:
		m.CPU.SetIRL(0, 0)
	}
}

// spgStatus is the live scan state: the line counter run.go sweeps, the
// field number, and vsync asserted inside the blanking interval the game's
// own SPG_VBLANK_INT delimits.
func (m *Machine) spgStatus() uint32 {
	vbl := m.PVRRegs[0xCC/4]
	vbIn, vbOut := vbl&0x3FF, vbl>>16&0x3FF
	line := m.CurLine
	var v uint32
	inBlank := false
	if vbIn > vbOut { // the usual shape: blanking wraps the field boundary
		inBlank = line >= vbIn || line < vbOut
	} else if vbIn != vbOut {
		inBlank = line >= vbIn && line < vbOut
	}
	if inBlank {
		v |= 1 << 13
	}
	v |= m.FieldNum << 10
	return v | line&0x3FF
}
