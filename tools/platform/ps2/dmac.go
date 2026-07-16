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

// A channel's registers, at its base (dmacChanBase) + the register offset.
const (
	dChcr = 0x00 // direction, mode, and the start bit
	dMadr = 0x10 // the address in memory the data is at
	dQwc  = 0x20 // how many quadwords to move
	dTadr = 0x30 // the next chain tag, for the chain modes
	dAsr0 = 0x40 // a saved tag address, for call/ret chain tags
	dAsr1 = 0x50
	dSadr = 0x80 // the scratchpad address, for the SPR channels
)

// The channel register blocks are NOT uniformly spaced. Channels 0-2 sit 0x1000
// apart, but from IPU onward the EE packs them 0x400 apart in three groups. Decoding
// by (addr-base)/0x1000 — as if the spacing were uniform — folds 0x1000D000/D400 (the
// scratchpad channels) onto channel 5, so a scratchpad transfer arrives looking like a
// SIF start and moves nothing. This is the map the silicon actually uses, read off the
// addresses the game's own `ultimate-memcpy` writes (0x1000D400 SPR_TO, 0x1000D000
// SPR_FROM): the bases are a table, not an arithmetic progression.
var dmacChanBase = [dmacChannels]uint32{
	0x10008000, // 0  VIF0
	0x10009000, // 1  VIF1
	0x1000A000, // 2  GIF
	0x1000B000, // 3  IPU_FROM
	0x1000B400, // 4  IPU_TO
	0x1000C000, // 5  SIF0
	0x1000C400, // 6  SIF1
	0x1000C800, // 7  SIF2
	0x1000D000, // 8  SPR_FROM (scratchpad -> memory)
	0x1000D400, // 9  SPR_TO   (memory -> scratchpad)
}

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
// one of the controller-wide registers above the channel block. A channel's register
// window is the 0x100 bytes at its base (registers run 0x00..0x80); the bases are the
// non-uniform table above, so the decode is a lookup, not a division.
func dmacChanReg(a uint32) (ch int, reg uint32, ok bool) {
	if a < dmacBase || a >= dCTRL {
		return 0, 0, false
	}
	for i, base := range dmacChanBase {
		if a >= base && a < base+0x100 {
			return i, a - base, true
		}
	}
	return 0, 0, false
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

	case dmacChVIF0:
		m.vifStart(0, c)

	case dmacChVIF1:
		m.vifStart(1, c)

	case dmacChSIF0, dmacChSIF1:
		// The SIF is moved by its own high-level path (sif.go): the reply from the IOP is
		// carried by sifFromIOP, the request to it by the sceSifSetDma syscall. So starting
		// the channel here is not a transfer — the data has already crossed, or is about to,
		// by another road. It only has to complete, so the kernel's "is the channel idle?"
		// check reads idle.

	case dmacChSPRto:
		// Memory -> scratchpad. The GOAL runtime's `ultimate-memcpy` bounces every large,
		// aligned block through the 16 KiB scratchpad: this channel reads main memory at
		// MADR into the scratchpad at SADR, and channel 8 reads it back out to the
		// destination. It is a real mover of bytes, and the kernel links against it — once
		// the GOAL `ultimate-memcpy` is defined it becomes the copy for every object main
		// segment 4 KiB and up, which is most of them. Left unmodelled, the segment arrives
		// as zeros and the first method that runs from it nop-slides into a break.
		//
		// The channel also takes a source chain, and the merc renderer drives it that way:
		// generic-merc-execute-asm kicks CHCR=0x144 (chain + TTE) with QWC=0 and TADR at the
		// chain generic-merc-add-to-cue built — a CNT header link followed by REF links that
		// scatter-gather the model's geometry out of the merc heap into the SPR's double
		// buffer. Treating that start as a normal transfer moves QWC=0 quadwords — nothing —
		// and mercneric-convert then converts whatever the scratchpad last held.
		if c.chcr&dChcrModeM == 1<<2 {
			m.dmacSPRChainTo(c)
		} else {
			m.dmacSPR(c, false)
		}

	case dmacChSPRfrom:
		// Scratchpad -> memory, the second half of the bounce: SADR in the scratchpad out
		// to MADR in main memory.
		m.dmacSPR(c, true)

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

// dmacSPR moves a scratchpad channel's quadwords between main memory (MADR) and the
// scratchpad (SADR). fromSPR picks the direction: false is channel 9 (SPR_TO, memory ->
// scratchpad), true is channel 8 (SPR_FROM, scratchpad -> memory).
//
// The scratchpad address is its own register here, not the top-bit-of-MADR encoding
// sceGsExecLoadImage uses for the GIF channel — the SPR channels have a real SADR (reg
// 0x80). MADR is a plain main-memory address. The transfer is n = QWC*16 bytes, byte at
// a time so it goes through the same memory path (and the same write hook) as any other
// store; these copies happen at link time, not per frame, so the cost does not matter.
func (m *Machine) dmacSPR(c *dmacChan, fromSPR bool) {
	n := c.qwc * 16
	madr := c.madr & 0x0FFFFFFF
	for i := uint32(0); i < n; i++ {
		s := (c.sadr + i) & (spramSize - 1)
		if fromSPR {
			// Scratchpad -> memory: the destination store is a real write, so leave the
			// write hook live for a watch to see the copy land.
			m.Write(madr+i, m.spram[s])
		} else {
			// Memory -> scratchpad: mute the read hook so a data watch is not drowned by
			// the source read; the scratchpad store fires no hook of its own.
			m.hookMuted = true
			m.spram[s] = m.Read(madr + i)
			m.hookMuted = false
		}
	}
}

// dmacSPRChainTo walks a source chain into the scratchpad: each link's quadwords land at
// SADR and SADR advances, so a chain of REF links is a gather — which is what the merc
// renderer uses it for, pulling a model's scattered fragments out of the merc heap into
// the SPR's double buffer in one kick. With TTE set the whole 16-byte tag lands too (see
// dmacSourceChain on why the scratchpad takes all 16), which is how the buffer's header
// — its quadword count and the next chain segment's address — arrives: the game rides it
// in the tag's device half.
func (m *Machine) dmacSPRChainTo(c *dmacChan) {
	sadr := c.sadr & (spramSize - 1)
	m.dmacSourceChain(dmacChSPRto, c, func(b []byte) {
		for _, x := range b {
			m.spram[sadr] = x
			sadr = (sadr + 1) & (spramSize - 1)
		}
	}, true)
	c.sadr = sadr // left where the transfer ended, like MADR; the game rewrites it each kick
}

// The source-chain DMAtag IDs. A chain is a linked list the game builds in memory: each
// tag says where this link's quadwords are and where the next tag is, and the DMA
// controller walks it so the CPU can hand over a whole frame's worth of scattered
// buffers in one start. This is how the render channels are driven — the GOAL engine
// never uses the flat mode for drawing.
const (
	dtagREFE = 0 // data at ADDR, then end
	dtagCNT  = 1 // data follows this tag; the next tag follows the data
	dtagNEXT = 2 // data follows this tag; the next tag is at ADDR
	dtagREF  = 3 // data at ADDR; the next tag follows this one
	dtagREFS = 4 // REF, with the stall control watching it
	dtagCALL = 5 // data follows; push the after-data address, continue at ADDR
	dtagRET  = 6 // data follows; continue at the pushed address
	dtagEND  = 7 // data follows, then end
)

// dmacSourceChain walks a source chain from the channel's TADR, handing each link's data
// to feed. It is the shape both render channels share; what the bytes mean is the
// device's business (the GIF parses GIFtags, the VIF parses VIFcodes).
//
// The tag is a quadword. Its low 64 bits belong to the DMA controller — QWC in the low
// halfword, the ID in bits 28..30, the interrupt request in bit 31, the address in bits
// 32..62 and the scratchpad flag in bit 63. The high 64 bits belong to the DEVICE: with
// CHCR's TTE bit set the controller forwards them, which is how a VIF chain rides two
// VIFcodes along on every tag without spending a link on them.
//
// What TTE forwards depends on what is listening. The VIF and the GIF are parsers and
// take the tag's upper 64 bits; the scratchpad is memory and takes the whole quadword
// (wholeTag) — the merc chain proves it from its own reads: the buffer's consumer reads
// its quadword count at +8 and the next segment's address at +12, which are exactly the
// tag's two upper words, sitting where only a 16-byte tag landing at SADR would put them.
//
// The scratchpad flag is folded into bit 31 of the address, which is the encoding
// dmaBytes already understands (it is the one sceGsExecLoadImage uses for MADR).
func (m *Machine) dmacSourceChain(ch int, c *dmacChan, feed func([]byte), wholeTag bool) {
	tte := c.chcr&dChcrTTE != 0

	// The walk is bounded, and the bound is reported when hit rather than silently
	// truncating a frame: a chain that long is a corrupt chain, and the address where it
	// happened is the diagnosis.
	const maxLinks = 1 << 16
	for i := 0; i < maxLinks; i++ {
		tag := m.dmaBytes(c.tadr, 1)
		lo := le64(tag)

		qwc := uint32(lo & 0xFFFF)
		id := uint32(lo>>28) & 7
		irq := lo&(1<<31) != 0
		addr := uint32(lo>>32) &^ 0xF & 0x7FFFFFFF
		if lo&(1<<63) != 0 {
			addr |= 0x80000000 // the scratchpad, in dmaBytes' encoding
		}

		if tte {
			if wholeTag {
				feed(tag[0:16])
			} else {
				feed(tag[8:16])
			}
		}

		data := c.tadr&0x80000000 | (c.tadr&0x7FFFFFFF + 16) // "follows the tag", staying in the tag's memory
		end := false
		switch id {
		case dtagREFE:
			data = addr
			end = true
		case dtagCNT:
			c.tadr = data + qwc*16
		case dtagNEXT:
			c.tadr = addr
		case dtagREF, dtagREFS:
			data = addr
			c.tadr += 16
		case dtagCALL:
			c.asr1 = c.asr0
			c.asr0 = data&0x80000000 | (data&0x7FFFFFFF + qwc*16)
			c.tadr = addr
		case dtagRET:
			if c.asr0 == 0 && c.asr1 == 0 {
				end = true
				break
			}
			c.tadr = c.asr0
			c.asr0, c.asr1 = c.asr1, 0
		case dtagEND:
			end = true
		}

		if qwc > 0 {
			feed(m.dmaBytes(data, qwc))
		}
		if end {
			return
		}
		if irq && c.chcr&dChcrTIE != 0 {
			return // the tag asked for an interrupt and the channel honours it: stop here
		}
		if i == maxLinks-1 {
			m.note("EE DMA: channel %d's chain ran %d links without ending — truncated at TADR 0x%08X",
				ch, maxLinks, c.tadr)
		}
	}
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
