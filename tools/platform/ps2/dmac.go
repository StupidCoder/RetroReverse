package ps2

// dmac.go is the Emotion Engine's DMA controller — the piece of silicon that moves
// bulk data between main memory and the peripherals without the CPU copying it word
// by word. On the EE it is what every render frame is made of: the game builds a list
// of GS commands in memory and hands it to a DMA channel, and the channel walks it
// into the GIF while the CPU goes on to build the next one.
//
// The controller has ten channels, each owning one peripheral:
//
//	0  VIF0     the first vector unit's data path
//	1  VIF1     the second vector unit's data path
//	2  GIF      the Graphics Synthesizer — the render channel
//	3  IPU_FROM the image processor, its output
//	4  IPU_TO   the image processor, its input
//	5  SIF0     from the IOP  (the second processor's replies land here)
//	6  SIF1     to the IOP    (the EE's requests go out here)
//	7  SIF2     the third SIF channel
//	8  SPR_FROM the scratchpad, reading out
//	9  SPR_TO   the scratchpad, writing in
//
// This is the mirror of the IOP's controller (iopdma.go): a channel is armed by
// writing its registers and started by setting CHCR's STR bit, and the controller
// clears STR when the transfer is done. The game watches STR to know a transfer has
// finished — sceGsExecLoadImage writes STR and then polls it — so the whole contract
// the boot depends on is: when STR is written, move the data and clear STR.
//
// Which of the ten channels the game actually drives is left to the game to say. The
// GIF channel is the one the render path uses, and it is the one modelled through to
// its device (the GIF, gif.go); the SIF channels are handled by the SIF's own HLE
// (sif.go), which owns the data movement, so here they only complete; the rest
// complete and are counted, exactly as the IOP's unmodelled channels are.

// The channel register block and the controller-wide registers.
const (
	dmacBase = 0x10008000
	dmacEnd  = 0x1000F000

	dmacChannels = 10

	// The controller-wide registers, above the channels.
	dCTRL    = 0x1000E000 // enable, and the stall/transfer-tag controls
	dSTAT    = 0x1000E010 // channel interrupt status (low 16) and its mask (high 16)
	dPCR     = 0x1000E020 // per-channel priority / enable
	dSQWC    = 0x1000E030 // the quadword count that triggers a stall
	dRBSR    = 0x1000E040 // the ring-buffer size, for the SIF's chain
	dRBOR    = 0x1000E050 // the ring-buffer origin
	dSTADR   = 0x1000E060 // the stall address
	dENABLER = 0x1000F520
	dENABLEW = 0x1000F590
)

// A channel's registers, at dmacBase + 0x1000*channel.
const (
	dChcr = 0x00 // direction, mode, and the start bit
	dMadr = 0x10 // the address in memory the data is at
	dQwc  = 0x20 // how many quadwords to move
	dTadr = 0x30 // the next chain tag, for the chain modes
	dAsr0 = 0x40 // a saved tag address, for call/ret chain tags
	dAsr1 = 0x50
	dSadr = 0x80 // the scratchpad address, for the SPR channels
)

// CHCR's bits. STR is the one that matters: the game sets it to start a transfer and
// polls it to learn the transfer is done.
const (
	dChcrDir   = 1 << 0 // 1: memory -> device. 0: device -> memory.
	dChcrModeM = 3 << 2 // 00 normal, 01 chain, 10 interleave
	dChcrTTE   = 1 << 6 // transfer the tag itself, in chain mode
	dChcrTIE   = 1 << 7 // interrupt when a tag with the IRQ bit completes
	dChcrStart = 1 << 8 // set to begin; the controller clears it when done
)

// The channels this game drives, named where the render path uses them.
const (
	dmacChVIF0    = 0
	dmacChVIF1    = 1
	dmacChGIF     = 2
	dmacChSIF0    = 5
	dmacChSIF1    = 6
	dmacChSPRfrom = 8
	dmacChSPRto   = 9
)

// dmacChan is one channel's register file.
type dmacChan struct {
	chcr, madr, qwc, tadr, asr0, asr1, sadr uint32
}

// dmacChanReg decodes an address into a channel and a register, or reports that it is
// one of the controller-wide registers above the channel block.
func dmacChanReg(a uint32) (ch int, reg uint32, ok bool) {
	if a < dmacBase || a >= dCTRL {
		return 0, 0, false
	}
	ch = int(a-dmacBase) / 0x1000
	if ch >= dmacChannels {
		return 0, 0, false
	}
	return ch, a & 0xFF, true
}

// dmacRead serves a read of the controller.
func (m *Machine) dmacRead(a uint32) (uint32, bool) {
	if ch, reg, ok := dmacChanReg(a); ok {
		c := &m.dmac[ch]
		switch reg {
		case dChcr:
			return c.chcr, true
		case dMadr:
			return c.madr, true
		case dQwc:
			return c.qwc, true
		case dTadr:
			return c.tadr, true
		case dAsr0:
			return c.asr0, true
		case dAsr1:
			return c.asr1, true
		case dSadr:
			return c.sadr, true
		}
		return 0, true
	}
	switch a {
	case dCTRL:
		return m.dCtrl, true
	case dSTAT:
		return m.dmacStat | m.dmacMask<<16, true
	case dPCR:
		return m.dPcr, true
	case dSQWC:
		return m.dSqwc, true
	case dRBSR:
		return m.dRbsr, true
	case dRBOR:
		return m.dRbor, true
	}
	return 0, false
}

// dmacWrite serves a write, and starts a transfer when STR is set.
func (m *Machine) dmacWrite(a, v uint32) bool {
	if ch, reg, ok := dmacChanReg(a); ok {
		c := &m.dmac[ch]
		switch reg {
		case dChcr:
			c.chcr = v
			if v&dChcrStart != 0 {
				m.dmacStart(ch)
			}
		case dMadr:
			c.madr = v
		case dQwc:
			c.qwc = v & 0xFFFF
		case dTadr:
			c.tadr = v
		case dAsr0:
			c.asr0 = v
		case dAsr1:
			c.asr1 = v
		case dSadr:
			c.sadr = v
		}
		return true
	}
	switch a {
	case dCTRL:
		m.dCtrl = v
	case dSTAT:
		// D_STAT has two halves and they are written differently, which is what makes the
		// register usable at all. The low 16 bits (the channel interrupt status) are
		// write-1-to-clear — an acknowledgement. The high 16 bits (the mask) are
		// write-1-to-*toggle*, which is the EE's oddity: the kernel flips a mask bit rather
		// than setting it. Storing the word plainly would make an acknowledgement look like
		// a fresh interrupt and a mask update wipe every other channel's.
		m.dmacStat &^= v & 0xFFFF
		m.dmacMask ^= (v >> 16) & 0xFFFF
	case dPCR:
		m.dPcr = v
	case dSQWC:
		m.dSqwc = v
	case dRBSR:
		m.dRbsr = v
	case dRBOR:
		m.dRbor = v
	default:
		return false
	}
	return true
}

// dmacStart performs a transfer.
//
// The bytes move now and STR clears now — the EE's DMA is fast and the code that
// started it goes straight on to poll STR, so unlike the IOP's sound channel there is
// no ordering hazard that a latency would fix. What the channel does with the bytes is
// the channel's business: the GIF channel feeds them to the Graphics Synthesizer, the
// SIF channels are already served by the SIF's HLE, and the rest are counted.
func (m *Machine) dmacStart(ch int) {
	c := &m.dmac[ch]

	switch ch {
	case dmacChGIF:
		m.gifStart(c)

	case dmacChSIF0, dmacChSIF1:
		// The SIF is moved by its own high-level path (sif.go): the reply from the IOP is
		// carried by sifFromIOP, the request to it by the sceSifSetDma syscall. So starting
		// the channel here is not a transfer — the data has already crossed, or is about to,
		// by another road. It only has to complete, so the kernel's "is the channel idle?"
		// check reads idle.

	default:
		// A channel we have no evidence this game drives. It completes — refusing would hang
		// the caller rather than teach anything — but it is named once, the same way the IOP's
		// unmodelled channels are, so the first game that drives one is a line in the log and
		// not a silent nothing.
		m.note("EE DMA: channel %d started (mode 0x%X, %d qwords at 0x%08X), and nothing models it",
			ch, (c.chcr&dChcrModeM)>>2, c.qwc, c.madr)
	}

	m.dmacComplete(ch)
}

// dmacComplete clears the channel's busy bit and records its interrupt in D_STAT.
//
// It records the interrupt; it does not deliver it. This machine's kernel is high-level
// emulated, so interrupt handlers are run directly (deliverVBlank) rather than by
// vectoring the CPU through a table nothing has filled in. A boot that uses the GIF
// channel polls STR and never registers a handler for it, so for the render path this is
// only the STR clear and the status bit a later reader of D_STAT will see. Vectoring the
// CPU here instead would land it on an empty exception vector — an interrupt the HLE was
// never structured to take.
func (m *Machine) dmacComplete(ch int) {
	m.dmac[ch].chcr &^= dChcrStart
	m.dmacStat |= 1 << uint(ch)
}

// dmaBytes reads a channel's source data out of memory: n quadwords from MADR, honouring
// the scratchpad-source flag the EE folds into the top bit of MADR.
//
// sceGsExecLoadImage is the authority on the encoding — it masks the source to 28 bits
// and, when the source is the scratchpad (its address has 0x70000000 set), sets bit 31 of
// MADR. So bit 31 means "read from the scratchpad", and the low bits are the offset.
func (m *Machine) dmaBytes(madr, qwc uint32) []byte {
	n := qwc * 16
	out := make([]byte, n)
	spr := madr&0x80000000 != 0
	addr := madr & 0x0FFFFFFF
	m.hookMuted = true
	defer func() { m.hookMuted = false }()
	for i := uint32(0); i < n; i++ {
		if spr {
			out[i] = m.spram[(addr+i)&(spramSize-1)]
		} else {
			out[i] = m.Read(addr + i)
		}
	}
	return out
}
