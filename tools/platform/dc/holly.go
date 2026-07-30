package dc

// holly.go is the system block's interrupt fan-in: three status registers
// (normal, external, error) gated by three mask sets, one per SH-4 interrupt
// level. A bit pending under a level-6 mask asserts IRL level 13, level 4
// masks assert 11, level 2 masks assert 9 — the KOS-documented routing — and
// the INTEVT codes are the SH7750's independent-line encodings (0x3A0, 0x360,
// 0x320). The VBlank heartbeat is ISTNRM bit 3, raised by the run loop every
// field.

import "encoding/binary"

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
		// Ch2 (PVR) DMA start: the source address sits in the SH-4 DMAC's
		// SAR2, which the CPU raw-stores; the destination decides the path.
		if v&1 != 0 {
			// Only the polygon path (10xxxxxx) carries TA parameters; a
			// texture-path job (11xxxxxx) is pixels, and parsing pixels as
			// parameters manufactured list-completions the game never made.
			src := m.CPU.OnchipReg(0xFFA00020)
			if m.C2DStat&0x01000000 == 0 {
				m.taSubmitDMA(src, m.C2DLen)
			} else {
				// The texture path: the pixels land in VRAM for the
				// rasteriser to sample.
				dst := m.C2DStat & 0x00FFFFFF
				if m.OnC2DTexture != nil {
					m.OnC2DTexture(src, dst, m.C2DLen)
				}
				for i := uint32(0); i+4 <= m.C2DLen && dst+i+4 <= VRAMSize; i += 4 {
					v := m.ram32(src + i)
					m.VRAM[dst+i] = uint8(v)
					m.VRAM[dst+i+1] = uint8(v >> 8)
					m.VRAM[dst+i+2] = uint8(v >> 16)
					m.VRAM[dst+i+3] = uint8(v >> 24)
				}
			}
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

// taSubmitDMA streams a ch2 DMA job's polygon-path words into the TA: each
// 32-byte chunk is recorded for the rasteriser and fed to the list tracker.
func (m *Machine) taSubmitDMA(src, byteLen uint32) {
	var chunk [32]byte
	for off := uint32(0); off+32 <= byteLen; off += 32 {
		for i := uint32(0); i < 32; i += 4 {
			binary.LittleEndian.PutUint32(chunk[i:], m.ram32(src+off+i))
		}
		m.TAFrame = append(m.TAFrame, chunk[:]...)
		m.taFeed(chunk[:])
	}
}

// taFifoWrite is the other submission path the hardware has: the game's
// store queues (or plain stores) into the polygon-path FIFO window. Words
// accumulate until a 32-byte chunk is complete, then follow the same
// record-and-feed as a DMA'd chunk; an END_OF_LIST arriving this way earns
// its list-done interrupt through a countdown of its own, deferred for the
// same reason a DMA completion is.
func (m *Machine) taFifoWrite(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	m.TAFifo = append(m.TAFifo, b[:]...)
	if len(m.TAFifo) < 32 {
		return
	}
	m.TAFrame = append(m.TAFrame, m.TAFifo...)
	m.taFeed(m.TAFifo)
	m.TAFifo = m.TAFifo[:0]
	if m.C2DPendingBits != 0 && m.C2DCountdown == 0 && m.TAFifoCountdown == 0 {
		m.TAFifoCountdown = 2000
	}
}

// taFeed consumes one 32-byte TA parameter chunk (control word first),
// tracking the current list type from each global parameter and collecting
// the matching list-complete bit at each END_OF_LIST. The current list
// persists in Machine.TAList across jobs — a list continued in a later chunk
// must not have its completion attributed to opaque — and resets the way the
// hardware's does: on TA_LIST_INIT and on soft reset (machine.go). A 64-byte
// parameter owes a second chunk, counted in TANeed, consumed uninterpreted
// even across job boundaries.
func (m *Machine) taFeed(p []byte) {
	if m.TANeed > 0 {
		m.TANeed -= 32
		return
	}
	pcw := binary.LittleEndian.Uint32(p)
	switch pcw >> 29 {
	case 0: // END_OF_LIST closes the open list
		if m.TAList >= 0 && m.TAList < 5 {
			m.C2DPendingBits |= taListDoneBit[m.TAList]
		}
		m.TAList = -1
	case 4: // polygon header: names the list and fixes the vertex size
		m.taOpenList(int32(pcw >> 24 & 7))
		textured := pcw&8 != 0
		colType := pcw >> 4 & 3
		volume := pcw&0x40 != 0
		m.TAVtx = 32
		// 64-byte vertices: two-volume parameters, floating-point base
		// color with a texture, and the modifier-volume lists' triangle
		// records.
		if volume || (textured && colType == 1) || m.TAList == 1 || m.TAList == 3 {
			m.TAVtx = 64
		}
		if volume && textured { // the two-volume header carries a second TSP set
			m.TANeed = 32
		} else if colType == 2 && pcw&4 != 0 && textured {
			m.TANeed = 32 // intensity-with-offset: face colours fill a second chunk
		}
	case 5: // sprite header: sprite vertices are always 64 bytes
		m.taOpenList(int32(pcw >> 24 & 7))
		m.TAVtx = 64
	case 7: // vertex parameter: the header decided its size
		if m.TAVtx == 64 {
			m.TANeed = 32
		}
	}
}

// taOpenList makes list n the TA's current list. Lists are strictly ordered
// on the hardware: a header naming a new list while a different one is open
// closes the open list and fires its completion — Crazy Taxi's frame stream
// depends on it (LIST_INIT leaves opaque open, the opaque-modifier header
// implicitly closes it, and the game's interrupt bookkeeping counts four
// completions against three explicit END_OF_LISTs).
func (m *Machine) taOpenList(n int32) {
	if m.TAList >= 0 && m.TAList != n {
		if m.TAList < 5 {
			m.C2DPendingBits |= taListDoneBit[m.TAList]
		}
	}
	m.TAList = n
}

// tickCompletions retires the deferred render and DMA countdowns.
func (m *Machine) tickCompletions() {
	if m.RenderCountdown > 0 {
		if m.RenderCountdown--; m.RenderCountdown == 0 {
			m.renderFrame()
			m.raiseNRM(istRenderDone)
			if m.OnRender != nil {
				// The frame boundary for a capture: the picture is complete
				// and still in the write framebuffer — the guest has not yet
				// seen the render-done interrupt, so it cannot have flipped.
				m.OnRender()
			}
		}
	}
	if m.C2DCountdown > 0 {
		if m.C2DCountdown--; m.C2DCountdown == 0 {
			m.raiseNRM(istCh2DMA | m.C2DPendingBits)
			m.C2DPendingBits = 0
		}
	}
	if m.TAFifoCountdown > 0 {
		if m.TAFifoCountdown--; m.TAFifoCountdown == 0 {
			m.raiseNRM(m.C2DPendingBits)
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
