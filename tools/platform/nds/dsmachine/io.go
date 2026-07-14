package dsmachine

// Memory-mapped I/O: the register file both cores reach at 0x04000000.
//
// The two CPUs do not see the same machine through it. Most of the block is
// duplicated per core (each has its own interrupt controller, its own four DMA
// channels, its own four timers, its own DISPSTAT), some of it belongs to exactly
// one of them (the graphics engines, the maths units and the VRAM controls are the
// ARM9's; the SPI bus, the sound chip and the extra buttons are the ARM7's), and a
// little of it — the IPC block — is the seam where they meet.
//
// A register this model does not implement is *logged*, once, rather than silently
// reading back the last value written. A DS boot polls status bits constantly, and
// an unmodelled bit that happens to read "ready" is indistinguishable from one that
// works, right up until the frame it isn't.

const (
	regDISPSTAT   = 0x04000004
	regVCOUNT     = 0x04000006
	regDMA0SAD    = 0x040000B0
	regDMAFILL    = 0x040000E0
	regTM0CNT     = 0x04000100
	regIPCSYNC    = 0x04000180
	regIPCFIFOCNT = 0x04000184
	regIPCFIFOSND = 0x04000188
	regIME        = 0x04000208
	regIE         = 0x04000210
	regIF         = 0x04000214
	regVRAMCNT    = 0x04000240 // A..G at 0x240..0x246, H/I at 0x248/0x249
	regWRAMCNT    = 0x04000247 // ARM9-only: the shared-WRAM split
	regPOSTFLG    = 0x04000300
	regPOWCNT     = 0x04000304
	regIPCFIFORCV = 0x04100000
	regRTC        = 0x04000138
)

func (c *core) other() *core {
	if c.arm9 {
		return c.m.ARM7
	}
	return c.m.ARM9
}

// ioRead services a 32-bit-aligned read of the I/O block.
func (c *core) ioRead(a uint32) uint32 {
	m := c.m
	v, ok := c.ioReadReg(a)
	if m.OnIO != nil {
		m.OnIO(c.arm9, false, a, v, c.cpu.R[15])
	}
	if !ok {
		m.note("%s: read unmodelled I/O 0x%08X", c.name, a)
	}
	return v
}

func (c *core) ioReadReg(a uint32) (uint32, bool) {
	m := c.m
	switch a {
	case regIME:
		if c.ime {
			return 1, true
		}
		return 0, true
	case regIE:
		return c.ie, true
	case regIF:
		return c.if_, true
	case regDISPSTAT:
		return c.dispstat() | uint32(m.vid.line&0xFF)<<16, true // VCOUNT shares the word
	case regVCOUNT:
		return uint32(m.vid.line), true
	case regKEYINPUT:
		return m.keyinput() | c.io[regKEYCNT]<<16, true
	case regRCNT:
		// EXTKEYIN is at 0x04000136 — the HIGH half of this word, not a word of its
		// own. Matching it at its own address is matching an address the bus never
		// presents, and the register then reads as zero: X and Y are active low, so
		// the game sees both held, and bit 7 reads the lid as CLOSED. A machine that
		// believes it has been folded shut does not render anything.
		if !c.arm9 {
			return c.io[regRCNT]&0xFFFF | m.extkeyin()<<16, true
		}
	case regIPCSYNC:
		var in, out uint32
		if c.arm9 {
			in, out = uint32(m.ipc.sync7), uint32(m.ipc.sync9)
		} else {
			in, out = uint32(m.ipc.sync9), uint32(m.ipc.sync7)
		}
		return in | out<<8 | c.io[regIPCSYNC]&0x6000, true
	case regIPCFIFOCNT:
		return uint32(c.fifoCnt()), true
	case regIPCFIFORCV:
		return c.fifoRecv(), true
	case regROMCTRL:
		if c == m.cardCore() {
			return m.cd.ctrl, true
		}
		return 0, true
	case regCARDDATA:
		if c == m.cardCore() {
			return c.cardReadData(), true
		}
		return 0xFFFFFFFF, true
	case regAUXSPICNT:
		return c.io[regAUXSPICNT], true
	case regSPICNT:
		if !c.arm9 {
			// The busy bit (7) always reads clear: our transfers complete instantly.
			// SPIDATA is the high half of the same word.
			return c.io[regSPICNT]&^0x0080 | uint32(m.spi.out)<<16, true
		}
	case regRTC:
		if !c.arm9 {
			return 0, true // not modelled; nothing in this boot depends on the date
		}
	case regDIVCNT:
		return m.div.cnt, true
	case regSQRTCNT:
		return m.sqrt.cnt, true
	case regSQRT_RESULT:
		return m.sqrt.result, true
	case regPOSTFLG:
		return 1, true // the boot has completed
	case regPOWCNT:
		return m.powcnt, true
	case regEXMEMCNT:
		return c.io[regEXMEMCNT], true
	case regKEYCNT:
		return c.io[regKEYCNT], true
	}

	switch {
	case a >= regDIV_NUMER && a < regSQRTCNT:
		return m.divRead(a), true
	case a >= regSQRT_PARAM && a < regSQRT_PARAM+8:
		return half64(m.sqrt.param, a-regSQRT_PARAM), true
	case a >= regDMA0SAD && a < regDMAFILL:
		return c.dmaRead(a), true
	case a >= regTM0CNT && a < regTM0CNT+0x10:
		n := int(a-regTM0CNT) / 4
		return uint32(c.timers[n].counter) | uint32(c.timers[n].ctrl)<<16, true
	case a >= 0x04000320 && a < 0x040006A4 && c.arm9:
		return m.gpu3d.readReg(a), true
	case a >= 0x04000000 && a < 0x04000070, a >= 0x04001000 && a < 0x04001070:
		return c.io[a], true // the 2D engines' registers read back as written
	case a >= 0x04000400 && a < 0x04000520 && !c.arm9:
		return c.io[a], true // sound
	}
	return c.io[a], false
}

// ioWrite services a byte write into the I/O block. Registers are assembled a byte
// at a time and their side effects fire on the byte that *completes* the register,
// so a byte-wise store from the CPU core does not push four partial FIFO words or
// start the same DMA transfer four times.
func (c *core) ioWrite(a uint32, v byte) {
	m := c.m
	if m.OnIO != nil {
		m.OnIO(c.arm9, true, a, uint32(v), c.cpu.R[15])
	}

	// The byte-addressed registers come first: VRAMCNT is nine independent bytes,
	// and WRAMCNT sits in the middle of them, at 0x247.
	if a >= regVRAMCNT && a <= 0x04000249 && c.arm9 {
		// Keep the register file in step as well as the mapping, so a diagnostic that
		// reads VRAMCNT back sees what the game wrote rather than a zero that looks
		// exactly like "the game never mapped any VRAM".
		sh := 8 * (a & 3)
		c.io[a&^3] = c.io[a&^3]&^(0xFF<<sh) | uint32(v)<<sh
		if a == regWRAMCNT {
			m.wramcnt = v
		} else {
			m.vram.setCNT(vramBankIndex(a), v)
		}
		return
	}

	base := a &^ 3
	shift := 8 * (a & 3)
	lane := a & 3
	c.io[base] = c.io[base]&^(0xFF<<shift) | uint32(v)<<shift
	w := c.io[base]

	handled := true
	switch base {
	case regIME:
		c.ime = w&1 != 0
	case regIE:
		c.ie = w
	case regIF:
		// Write-1-to-clear, and only over the bits this byte actually carried.
		c.if_ &^= uint32(v) << shift
		c.io[base] = c.if_
	case regDISPSTAT:
		// VCOUNT occupies the top half of this word and is read-only, so a 32-bit
		// write must not be allowed to "set" the current scanline.
		c.io[regDISPSTAT] = w & 0xFFFF
	case regIPCSYNC:
		if lane != 1 { // the outgoing nibble and the send-IRQ bit live in the high byte
			return
		}
		out := uint8(w>>8) & 0xF
		if c.arm9 {
			m.ipc.sync9 = out
		} else {
			m.ipc.sync7 = out
		}
		if m.SyncTrace != nil {
			m.SyncTrace(c.name, out, c.cpu.R[15])
		}
		if w&0x2000 != 0 { // send an IRQ to the other CPU, if it asked for them
			o := c.other()
			if o.io[regIPCSYNC]&0x4000 != 0 {
				o.raise(irqIPCSync)
			}
		}
	case regIPCFIFOSND:
		if lane == 3 {
			c.fifoSend(w)
		}
	case regIPCFIFOCNT:
		if lane == 1 {
			c.fifoCntWrite(uint16(w))
		}
	case regROMCTRL:
		if lane == 3 && c == m.cardCore() && w&romctrlBusy != 0 {
			m.cd.start(w)
			// A card read is normally drained by DMA rather than by the CPU: a mode-5
			// channel fires the moment the data is there.
			m.runDMA(dmaCard)
		}
	case regSPICNT:
		if lane == 2 && !c.arm9 { // SPIDATA is the high half of this word
			c.spiTransfer(v)
		}
	case regDIVCNT:
		m.div.cnt = w
		m.div.run()
	case regSQRTCNT:
		m.sqrt.cnt = w
		m.sqrt.run()
	case regPOWCNT:
		// Bit 15 (ARM9) swaps which engine drives which physical screen — the reason
		// a game's "top screen" can be engine B.
		m.powcnt = w
	default:
		handled = false
	}
	if handled {
		return
	}

	switch {
	case base >= regCARDCMD && base < regCARDCMD+8:
		m.cd.cmd[a-regCARDCMD] = v
	case base >= regDIV_NUMER && base < regSQRTCNT:
		m.divWrite(base, w)
	case base >= regSQRT_PARAM && base < regSQRT_PARAM+8:
		m.sqrt.param = setHalf64(m.sqrt.param, base-regSQRT_PARAM, w)
		m.sqrt.run()
	case base >= regDMA0SAD && base < regDMAFILL:
		c.dmaWrite(base, w, lane)
	case base >= regTM0CNT && base < regTM0CNT+0x10:
		n := int(base-regTM0CNT) / 4
		if lane <= 1 {
			c.timers[n].reload = uint16(w)
		}
		if lane == 3 {
			c.writeTimerCtrl(n, uint16(w>>16))
		}
	case base >= 0x04000320 && base < 0x040006A4 && c.arm9:
		if lane == 3 {
			m.gpu3d.writeReg(base, w)
		}
	case base >= 0x04000000 && base < 0x04000070, base >= 0x04001000 && base < 0x04001070:
		// The 2D engines: latched here, read at scanout (gpu2d.go).
	case base >= 0x04000400 && base < 0x04000520 && !c.arm9:
		// Sound: the register file is real, the mixer is a declared gap (sound.go).
		if lane == 3 && base < regSOUNDCNT && (base-regSOUNDxCNT)%16 == 0 && w&0x80000000 != 0 {
			m.soundKeyed() // a voice has been started; say so once
		}
	default:
		if !ioKnown[base] {
			m.note("%s: wrote unmodelled I/O 0x%08X = 0x%08X", c.name, base, w)
		}
	}
}

// vramBankIndex maps a VRAMCNT byte address to its bank (0..8). The nine control
// bytes are not contiguous: WRAMCNT sits at 0x247, between bank G and bank H.
func vramBankIndex(a uint32) int {
	if a <= 0x04000246 { // A..G at 0x240..0x246
		return int(a - regVRAMCNT)
	}
	return int(a-0x04000248) + 7 // H at 0x248, I at 0x249
}

// ioKnown lists registers that are deliberately latched-and-ignored, so the "not
// modelled" log stays a list of real gaps rather than noise.
var ioKnown = map[uint32]bool{
	regKEYINPUT: true, regAUXSPICNT: true, regEXMEMCNT: true,
	regPOSTFLG: true, regPOWCNT: true, regDISPSTAT: true, regRTC: true,
	regDMAFILL: true, regDMAFILL + 4: true, regDMAFILL + 8: true, regDMAFILL + 12: true,
}

// --- the 64-bit maths registers ---------------------------------------------

func half64(v uint64, off uint32) uint32 {
	if off == 0 {
		return uint32(v)
	}
	return uint32(v >> 32)
}

func setHalf64(v uint64, off uint32, w uint32) uint64 {
	if off == 0 {
		return v&^0xFFFFFFFF | uint64(w)
	}
	return v&0xFFFFFFFF | uint64(w)<<32
}

func (m *Machine) divRead(a uint32) uint32 {
	switch {
	case a < regDIV_DENOM:
		return half64(m.div.numer, a-regDIV_NUMER)
	case a < regDIV_RESULT:
		return half64(m.div.denom, a-regDIV_DENOM)
	case a < regDIVREM:
		return half64(m.div.result, a-regDIV_RESULT)
	default:
		return half64(m.div.rem, a-regDIVREM)
	}
}

func (m *Machine) divWrite(a, w uint32) {
	switch {
	case a < regDIV_DENOM:
		m.div.numer = setHalf64(m.div.numer, a-regDIV_NUMER, w)
	case a < regDIV_RESULT:
		m.div.denom = setHalf64(m.div.denom, a-regDIV_DENOM, w)
	default:
		return // the result and remainder are read-only
	}
	m.div.run()
}

// --- DMA registers ----------------------------------------------------------

func (c *core) dmaRead(a uint32) uint32 {
	n := int(a-regDMA0SAD) / 12
	if n > 3 {
		return 0
	}
	switch (a - regDMA0SAD) % 12 {
	case 0:
		return c.dma[n].src
	case 4:
		return c.dma[n].dst
	default:
		return c.dma[n].count | c.dma[n].ctrl<<16
	}
}

func (c *core) dmaWrite(a, w, lane uint32) {
	n := int(a-regDMA0SAD) / 12
	if n > 3 {
		return
	}
	switch (a - regDMA0SAD) % 12 {
	case 0:
		c.dma[n].src = w
	case 4:
		c.dma[n].dst = w
	default: // the count/control word
		c.dma[n].count = w & c.maxCount()
		c.dma[n].ctrl = w >> 16
		if lane == 3 { // the control half is complete
			c.writeDMACtrl(n)
		}
	}
}

// --- IPC FIFO ---------------------------------------------------------------

func (c *core) sendQ() *[]uint32 {
	if c.arm9 {
		return &c.m.ipc.to7
	}
	return &c.m.ipc.to9
}

func (c *core) recvQ() *[]uint32 {
	if c.arm9 {
		return &c.m.ipc.to9
	}
	return &c.m.ipc.to7
}

func (c *core) fifoSend(v uint32) {
	if c.io[regIPCFIFOCNT]&0x8000 == 0 { // FIFO disabled
		return
	}
	q := c.sendQ()
	if len(*q) >= 16 {
		return // full: the word is lost, and the error bit would be set
	}
	*q = append(*q, v)
	o := c.other()
	if o.io[regIPCFIFOCNT]&0x0400 != 0 { // the receiver's not-empty IRQ
		o.raise(irqIPCRecv)
	}
}

func (c *core) fifoRecv() uint32 {
	q := c.recvQ()
	if len(*q) == 0 {
		return c.lastRecv // an empty FIFO reads back the last word, with an error bit
	}
	v := (*q)[0]
	*q = (*q)[1:]
	c.lastRecv = v
	if len(*q) == 0 {
		o := c.other()
		if o.io[regIPCFIFOCNT]&0x0004 != 0 { // the sender's send-empty IRQ
			o.raise(irqIPCSend)
		}
	}
	return v
}

func (c *core) fifoCnt() uint16 {
	send, recv := *c.sendQ(), *c.recvQ()
	v := uint16(c.io[regIPCFIFOCNT]) & 0xBC04 // keep the control/enable bits
	if len(send) == 0 {
		v |= 0x0001
	}
	if len(send) >= 16 {
		v |= 0x0002
	}
	if len(recv) == 0 {
		v |= 0x0100
	}
	if len(recv) >= 16 {
		v |= 0x0200
	}
	return v
}

func (c *core) fifoCntWrite(v uint16) {
	if v&0x0008 != 0 { // send-FIFO clear
		*c.sendQ() = (*c.sendQ())[:0]
	}
	c.io[regIPCFIFOCNT] = uint32(v)
}
