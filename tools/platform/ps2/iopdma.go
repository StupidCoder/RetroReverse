package ps2

// iopdma.go is the IOP's DMA controller.
//
// It is thirteen channels in two blocks, and it is the piece of silicon that stands
// between the second processor's modules and everything they own: the disc, the sound
// chip, and the other processor. Until it existed, OVERLORD got as far as handing the
// SPU a block of sound data and then waited for a transfer that was never going to
// happen, in a loop seven instructions long.
//
// Which channel is which was not looked up. It was read off the boot, because a module
// enables its own channel before it uses it and no other module enables that one:
//
//	2   SIF2    SIFMAN sets nibble 2 of DPCR
//	3   CDVD    CDVDMAN sets nibble 3
//	4   SPU     LIBSD sets nibble 4, then drives MADR/BCR/CHCR at 0x1F8010C0
//	7   SPU2    LIBSD sets nibble 0 of DPCR2 — the second sound core
//	9   SIF0    SIFMAN sets nibble 2 of DPCR2, and writes CHCR at 0x1F801528
//	10  SIF1    SIFMAN sets nibble 3 of DPCR2
//
// That is six of the thirteen, and it is exactly the six this disc uses. The rest are
// counted and left alone rather than invented; a channel nothing on the disc drives is
// a channel we have no evidence about.
//
// The block count is derived the same way. LIBSD programs BCR = 0x00010010 for a
// transfer OVERLORD had just rounded up to a multiple of 64 bytes — 0x10 words in a
// block, one block, 64 bytes — so BCR's low half is a block size in words and its high
// half is a count of blocks. The two modules agree, which is the only kind of
// confirmation available for a register nobody documents.

import "fmt"

// The DMA controller's registers.
//
// Two blocks, because the PlayStation's controller had seven channels and the PS2's
// IOP needed thirteen. The second block is the first one's layout again at a different
// address, which is why one piece of code serves both.
const (
	iopDMA1Base = 0x1F801080 // channels 0..6
	iopDMA1End  = 0x1F8010F0
	iopDPCR     = 0x1F8010F0 // which channels are enabled, and at what priority
	iopDICR     = 0x1F8010F4 // which channels may interrupt, and which have

	iopDMA2Base  = 0x1F801500 // channels 7..12
	iopDMA2End   = 0x1F801570
	iopDPCR2     = 0x1F801570
	iopDICR2     = 0x1F801574
	iopDMACEN    = 0x1F801578
	iopDMACINTEN = 0x1F80157C

	iopDMAChannels = 13
)

// The channels this disc drives, and the module that proved each one.
const (
	iopDMAChSIF2 = 2  // SIFMAN
	iopDMAChCDVD = 3  // CDVDMAN
	iopDMAChSPU0 = 4  // LIBSD, sound core 0
	iopDMAChSPU1 = 7  // LIBSD, sound core 1
	iopDMAChSIF0 = 9  // SIFMAN, IOP -> EE
	iopDMAChSIF1 = 10 // SIFMAN, EE -> IOP
)

// A channel's four registers, at base + 0x10*channel.
const (
	iopDMAMadr = 0x0 // where in IOP memory the data is
	iopDMABcr  = 0x4 // block size in words, and a count of blocks
	iopDMAChcr = 0x8 // direction, sync mode, and the start bit
	iopDMATadr = 0xC // the chain's next link, for the modes that use one
)

// CHCR's bits.
//
// Only three of them are load-bearing here, and the start bit is the one that matters:
// it is in the *top* byte, and the R3000A's bus writes a word low byte first, so by the
// time the controller sees the start bit the address and the block count beneath it have
// already arrived. A controller that triggers on it therefore never acts on half a
// command. See IOP.Write.
const (
	iopChcrFromRAM  = 1 << 0  // 1: IOP memory -> device. 0: device -> IOP memory.
	iopChcrStart    = 1 << 24 // set to begin; the controller clears it when done
	iopChcrTrigger  = 1 << 28 // the burst modes' "go" bit
	iopChcrSyncMask = 3 << 9
)

// iopDMAChan is one channel's register file.
type iopDMAChan struct {
	madr, bcr, chcr, tadr uint32
}

// iopDMADone is a transfer that has moved its bytes and is waiting out its latency
// before it says so.
type iopDMADone struct {
	at uint64 // the IOP step count at which the interrupt should arrive
	ch int
}

// iopDMALatency is how long a transfer takes, in IOP instructions.
//
// It is deliberately not zero. LIBSD writes the start bit and then goes on doing its own
// bookkeeping before it returns to its caller; a completion interrupt delivered on the
// very next instruction would run the callback — and wake the thread waiting on it —
// before the code that started the transfer had finished recording that it had. On the
// board a 64-byte transfer takes microseconds and the interrupt cannot possibly beat the
// store that follows the start bit. A few hundred instructions reproduces that ordering
// without pretending to model the bus.
const iopDMALatency = 400

// dmaReg decodes a DMA register address into a channel and a register, or reports that
// it is one of the controller-wide ones.
func iopDMAReg(a uint32) (ch int, reg uint32, ok bool) {
	switch {
	case a >= iopDMA1Base && a < iopDMA1End:
		return int(a-iopDMA1Base) / 0x10, a & 0xC, true
	case a >= iopDMA2Base && a < iopDMA2End:
		return 7 + int(a-iopDMA2Base)/0x10, a & 0xC, true
	}
	return 0, 0, false
}

// dmaRead serves a read of the controller.
func (p *IOP) dmaRead(a uint32) (uint32, bool) {
	if ch, reg, ok := iopDMAReg(a); ok {
		c := &p.dma[ch]
		switch reg {
		case iopDMAMadr:
			return c.madr, true
		case iopDMABcr:
			return c.bcr, true
		case iopDMAChcr:
			// The busy bit is what a module reads this for: LIBSD checks that the channel
			// is idle before it programs one. A transfer that has already moved its bytes
			// but not yet reported is still busy, and says so.
			return c.chcr, true
		case iopDMATadr:
			return c.tadr, true
		}
	}
	switch a {
	case iopDPCR:
		return p.dpcr, true
	case iopDPCR2:
		return p.dpcr2, true
	case iopDICR:
		return p.dicr, true
	case iopDICR2:
		return p.dicr2, true
	}
	return 0, false
}

// dmaWrite serves a write, and starts a transfer when one is asked for.
func (p *IOP) dmaWrite(a, v uint32) bool {
	if ch, reg, ok := iopDMAReg(a); ok {
		c := &p.dma[ch]
		switch reg {
		case iopDMAMadr:
			c.madr = v & 0x00FFFFFF // the IOP has 2 MiB; the address is 24 bits of it
		case iopDMABcr:
			c.bcr = v
		case iopDMAChcr:
			c.chcr = v
			if v&iopChcrStart != 0 {
				p.dmaStart(ch)
			}
		case iopDMATadr:
			c.tadr = v
		}
		return true
	}
	switch a {
	case iopDPCR:
		p.dpcr = v
		return true
	case iopDPCR2:
		p.dpcr2 = v
		return true
	case iopDICR:
		// The channel-interrupt flags in the top byte are write-1-to-clear, which is what
		// makes the register's read half meaningful. Nothing on this disc enables a DMA
		// interrupt — the SPU signals its own completion on its own line — so this path is
		// correct and unexercised, and it is written out rather than stubbed so that the
		// first module that does enable one is not met with a lie.
		p.dicr = (p.dicr &^ (v & 0x7F000000)) | (v &^ 0x7F000000)
		return true
	case iopDICR2:
		p.dicr2 = (p.dicr2 &^ (v & 0x3F000000)) | (v &^ 0x3F000000)
		return true
	}
	return false
}

// dmaEnabled reports whether DPCR/DPCR2 has this channel switched on. A transfer
// started on a channel nobody enabled is a bug in the module or in our decode of the
// register, and either way it is worth hearing about rather than silently performing.
func (p *IOP) dmaEnabled(ch int) bool {
	if ch < 7 {
		return p.dpcr>>(uint(ch)*4)&0x8 != 0
	}
	return p.dpcr2>>(uint(ch-7)*4)&0x8 != 0
}

// dmaStart performs a transfer.
//
// The bytes move now and the completion is reported later (dmaTick). That split is the
// whole of the timing model, and it is enough: the data is there the instant the device
// could possibly want it, and the *interrupt* — which is the only part any module
// observes the timing of — arrives after the code that started the transfer has finished.
func (p *IOP) dmaStart(ch int) {
	c := &p.dma[ch]

	if !p.dmaEnabled(ch) {
		p.ps2.note("IOP DMA: channel %d started but DPCR never enabled it — check the decode", ch)
	}

	// BCR: a block size in words, and a count of blocks. In the burst modes the count is
	// unused and a size of zero means 0x10000 words, which is the one place this register
	// is not simply a product.
	words := c.bcr & 0xFFFF
	blocks := c.bcr >> 16
	if (c.chcr&iopChcrSyncMask)>>9 == 0 { // burst
		if words == 0 {
			words = 0x10000
		}
		blocks = 1
	}
	if blocks == 0 {
		blocks = 1
	}
	n := words * blocks * 4 // bytes
	toRAM := c.chcr&iopChcrFromRAM == 0

	switch ch {
	case iopDMAChSPU0:
		p.spu.dma(0, p, c.madr, n, toRAM)
	case iopDMAChSPU1:
		p.spu.dma(1, p, c.madr, n, toRAM)

	case iopDMAChSIF0:
		// The second processor is sending the first one something, and it is the only channel on
		// the IOP whose other end is a CPU rather than a device.
		//
		// It is also the only one that is a *chain*: SIFMAN writes this channel's TADR and never
		// its MADR, so neither the source nor the destination is in the registers — both are in
		// the tag the channel points at, and sifFromIOP reads them from there. Which is why n,
		// computed above from BCR, is not passed on: BCR describes the chain buffer, not the
		// transfer, and 128 bytes from address zero is what believing otherwise looks like.
		if !toRAM {
			p.ps2.sifFromIOP()
		}

	case iopDMAChSIF1:
		// The receiving half of EE -> IOP, and the one channel here that does not complete when
		// it is started.
		//
		// Starting it is not a transfer. SIFMAN arms it — a block size, a start bit, and no MADR
		// at all — and what that means is "I am ready for the next packet"; the transfer happens
		// later, when the *other* processor sends one. So there is no completion to schedule and
		// no interrupt to raise, and raising one anyway is not a harmless approximation. It is a
		// doorbell rung by nobody: SIFCMD's handler runs, finds no packet, arms the channel
		// again, and is rung again — which it did, three hundred and eighty-seven times, and
		// which looks from the outside like a machine hard at work.
		//
		// The bytes and the interrupt both arrive in sifPump, which is the sender's side — and
		// arming this channel is precisely the event it waits for, so it is asked now whether
		// the EE has anything for a receiver that is finally listening.
		p.ps2.sifPump()
		return

	default:
		// A channel we have no evidence about. It completes — refusing to would hang the
		// module rather than teach us anything — but it is counted and named, and the
		// count is the work list, exactly as an unmodelled kernel call is.
		p.unmodelledCalls[fmt.Sprintf("DMA channel %d", ch)]++
		if p.unmodelledCalls[fmt.Sprintf("DMA channel %d", ch)] == 1 {
			p.ps2.note("IOP DMA: channel %d moved %d bytes at 0x%08X, and nothing models that channel",
				ch, n, c.madr)
		}
	}

	p.dmaPending = append(p.dmaPending, iopDMADone{at: p.steps + iopDMALatency, ch: ch})
}

// dmaTick finishes any transfer whose latency has run out: it clears the busy bit and
// lets the device raise its interrupt.
func (p *IOP) dmaTick() {
	for len(p.dmaPending) > 0 && p.steps >= p.dmaPending[0].at {
		d := p.dmaPending[0]
		p.dmaPending = p.dmaPending[1:]

		p.dma[d.ch].chcr &^= iopChcrStart | iopChcrTrigger

		switch d.ch {
		case iopDMAChSPU0:
			p.spu.complete(0, p)
		case iopDMAChSPU1:
			p.spu.complete(1, p)
		}

		// And the channel's interrupt. Each channel has one of its own — see iopIRQs, where
		// the numbering is established — because intrman demuxes the controller's single
		// line into one interrupt per channel before any driver sees it. So a driver
		// registers a handler against a *channel*, and this is where that handler is
		// summoned.
		//
		// DICR is not consulted, and that is not an omission. DICR is how the *hardware*
		// controller is told which channels may interrupt, and the thing that programs it is
		// intrman — which is us. What a driver actually does to arm a channel's interrupt is
		// call EnableIntr on the channel's number, and that is exactly what the mask in
		// serviceIntr answers. Not one module on this disc ever writes DICR, and now it is
		// clear why: none of them has any business doing so.
		p.raiseIRQ(iopDMAIRQ(d.ch))
	}
}

// iopDMAIRQ is the interrupt number a channel's completion raises. See iopIRQs for how
// the numbering was established, and from what.
func iopDMAIRQ(ch int) uint32 {
	if ch < 7 {
		return uint32(32 + ch)
	}
	return uint32(40 + ch - 7)
}
