package dsmachine

// Memory-mapped I/O. The two cores share the IPC block (IPCSYNC + the FIFOs) and
// each has its own interrupt controller (IME/IE/IF). Everything else is stubbed so
// status polls read "ready/idle" and the boot keeps moving; each distinct stub is
// logged once.

const (
	regDISPSTAT   = 0x04000004
	regVCOUNT     = 0x04000006
	regIPCSYNC    = 0x04000180
	regIPCFIFOCNT = 0x04000184
	regIPCFIFOSND = 0x04000188
	regIME        = 0x04000208
	regIE         = 0x04000210
	regIF         = 0x04000214
	regIPCFIFORCV = 0x04100000
	regSPICNT     = 0x040001C0
	regAUXSPICNT  = 0x04000040 // touch/backup SPI on some maps; harmless stub
)

// interrupt-source bits (subset the boot uses)
const (
	irqVBlank  = 1 << 0
	irqIPCSync = 1 << 16
	irqIPCSend = 1 << 17
	irqIPCRecv = 1 << 18
)

func (c *core) other() *core {
	if c.arm9 {
		return c.m.ARM7
	}
	return c.m.ARM9
}

func (c *core) ioRead(a uint32) uint32 {
	m := c.m
	switch a {
	case regIME:
		if c.ime {
			return 1
		}
		return 0
	case regIE:
		return c.ie
	case regIF:
		return c.if_
	case regDISPSTAT:
		// bit0 VBlank flag: report set so a naive poll clears; VCOUNT advances too.
		return 0x1
	case regVCOUNT:
		c.m.vcount = (c.m.vcount + 1) & 0x1FF
		return uint32(c.m.vcount)
	case regIPCSYNC:
		var in, out uint32
		if c.arm9 {
			in, out = uint32(m.ipc.sync7), uint32(m.ipc.sync9)
		} else {
			in, out = uint32(m.ipc.sync9), uint32(m.ipc.sync7)
		}
		ctrl := uint32(c.io[regIPCSYNC]) & 0x6000 // keep the IRQ-enable/-send bits
		return in | out<<8 | ctrl
	case regIPCFIFOCNT:
		return uint32(c.fifoCnt())
	case regIPCFIFORCV:
		return c.fifoRecv()
	case regSPICNT, regAUXSPICNT:
		// clear the "busy" bit (7) so a transfer-complete poll returns immediately
		return c.io[a] &^ 0x80
	}
	if a>>24 == 0x04 {
		m.note("%s: read stub I/O 0x%08X → last/0", c.name, a)
	}
	return c.io[a]
}

func (c *core) ioWrite(a uint32, v byte) {
	base := a &^ 3
	if _, seen := c.io[base]; !seen {
		c.ioseq = append(c.ioseq, base)
	}
	shift := 8 * (a & 3)
	c.io[base] = c.io[base]&^(0xFF<<shift) | uint32(v)<<shift
	w := c.io[base]

	// Side effects fire once, on the byte that completes the register (a 16-bit reg
	// completes on its high byte, a 32-bit reg on its top byte), so a byte-at-a-time
	// STR/STRH from the CPU core doesn't push four partial FIFO words.
	lane := a & 3
	switch base {
	case regIME:
		c.ime = w&1 != 0
	case regIE:
		c.ie = w
	case regIF:
		c.if_ &^= w // write-1-to-clear (per-byte is fine)
	case regIPCSYNC:
		if lane != 1 { // out nibble + send-IRQ live in the high byte
			return
		}
		out := uint8(w>>8) & 0xF
		if c.arm9 {
			c.m.ipc.sync9 = out
		} else {
			c.m.ipc.sync7 = out
		}
		if c.m.SyncTrace != nil {
			c.m.SyncTrace(c.name, out, c.cpu.R[15])
		}
		if w&0x2000 != 0 { // bit13 = send IRQ to the other CPU (if it enabled bit14)
			o := c.other()
			if uint32(o.io[regIPCSYNC])&0x4000 != 0 {
				o.if_ |= irqIPCSync
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
	}
}

// --- IPC FIFO --------------------------------------------------------------

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
	if uint32(c.io[regIPCFIFOCNT])&0x8000 == 0 { // FIFO disabled: ignore
		return
	}
	q := c.sendQ()
	if len(*q) >= 16 {
		return // full
	}
	*q = append(*q, v)
	// receiver's recv-not-empty IRQ (its IPCFIFOCNT bit10)
	o := c.other()
	if uint32(o.io[regIPCFIFOCNT])&0x0400 != 0 {
		o.if_ |= irqIPCRecv
	}
}

func (c *core) fifoRecv() uint32 {
	q := c.recvQ()
	if len(*q) == 0 {
		return c.lastRecv // reading an empty FIFO returns the last word (+ error bit)
	}
	v := (*q)[0]
	*q = (*q)[1:]
	c.lastRecv = v
	// sender's send-empty IRQ when its FIFO drains (its IPCFIFOCNT bit2)
	if len(*q) == 0 {
		o := c.other()
		if uint32(o.io[regIPCFIFOCNT])&0x0004 != 0 {
			o.if_ |= irqIPCSend
		}
	}
	return v
}

// fifoCnt assembles this core's IPCFIFOCNT status from the live queue depths.
func (c *core) fifoCnt() uint16 {
	send := *c.sendQ()
	recv := *c.recvQ()
	v := uint16(c.io[regIPCFIFOCNT]) & 0xBC04 // keep the control/enable bits
	if len(send) == 0 {
		v |= 0x0001 // send empty
	}
	if len(send) >= 16 {
		v |= 0x0002 // send full
	}
	if len(recv) == 0 {
		v |= 0x0100 // recv empty
	}
	if len(recv) >= 16 {
		v |= 0x0200 // recv full
	}
	return v
}

func (c *core) fifoCntWrite(v uint16) {
	if v&0x0008 != 0 { // send-FIFO clear
		*c.sendQ() = (*c.sendQ())[:0]
	}
	c.io[regIPCFIFOCNT] = uint32(v)
}
