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

const istVBlankIn = 1 << 3

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
	default:
		m.logf("SB write %08X = %08X (PC %08X)", addr, v, m.CPU.CurPC())
		return
	}
	m.updateIRL()
}

// raiseNRM sets ISTNRM bits and re-evaluates the interrupt line.
func (m *Machine) raiseNRM(bits uint32) {
	m.Holly.ISTNRM |= bits
	m.updateIRL()
}

// updateIRL asserts the highest Holly level with a pending, unmasked bit.
func (m *Machine) updateIRL() {
	h := &m.Holly
	switch {
	case h.ISTNRM&h.IML6NRM != 0 || h.ISTEXT&h.IML6EXT != 0 || h.ISTERR&h.IML6ERR != 0:
		m.CPU.SetIRL(13, 0x3A0)
	case h.ISTNRM&h.IML4NRM != 0 || h.ISTEXT&h.IML4EXT != 0 || h.ISTERR&h.IML4ERR != 0:
		m.CPU.SetIRL(11, 0x360)
	case h.ISTNRM&h.IML2NRM != 0 || h.ISTEXT&h.IML2EXT != 0 || h.ISTERR&h.IML2ERR != 0:
		m.CPU.SetIRL(9, 0x320)
	default:
		m.CPU.SetIRL(0, 0)
	}
}

// spgStatus derives the scanline counter and vsync flag from the field
// phase: 525 lines sweep per field-pair on the NTSC timing the games
// configure, folded here to a 262-line field advancing with the instruction
// count. Precise enough for polling loops that wait for a line range or the
// vsync edge.
func (m *Machine) spgStatus() uint32 {
	line := uint32(m.instrInField * 262 / fieldInstructions)
	var vsync uint32
	if line >= 256 { // the blanking tail of the field
		vsync = 1 << 13
	}
	return vsync | line&0x3FF
}
