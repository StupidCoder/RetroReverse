package sh4

// onchip.go is the P4 register area at 0xFC000000-0xFFFFFFFF: the SH7091's
// own control registers, which never reach the external bus. The modeled set
// is what a Dreamcast boot actually exercises — CCN (with the MMU refusal),
// the interrupt controller's priority registers, the TMU (tmu.go), and the
// store-queue address registers.
//
// Everything else follows the raw-store-plus-gap-log discipline: a write
// lands in onchipRaw so a later read of the same register sees it (the bus
// controller's configuration dance is write-then-verify), and a read of a
// register nothing wrote is recorded in onchipGaps and returns zero — the gap
// list, surfaced through Gaps(), is the worklist of registers worth modeling,
// and a zero that mattered will have its address on it.

import (
	"fmt"
	"sort"
)

// SCIF: the on-chip serial port games print debug text through. The
// transmitter is modeled as always drained — true of a UART nobody is
// listening to — and every byte written to the FIFO lands in SerialTX, so
// the guest's own log is readable instead of lost. The receive side stays
// honestly empty.
const (
	scifSCFTDR = 0xFFE8000C // transmit FIFO
	scifSCFSR  = 0xFFE80010 // status: TEND|TDFE always, nothing received
)

func (c *CPU) onchipRead(addr uint32, size int) uint32 {
	if v, ok := c.tmuRead(addr); ok {
		return v
	}
	switch addr {
	case scifSCFSR:
		return 0x0060 // TEND | TDFE
	case 0xFF000000:
		return c.PTEH
	case 0xFF000004:
		return c.PTEL
	case 0xFF000008:
		return c.TTB
	case 0xFF00000C:
		return c.TEA
	case 0xFF000010:
		return c.MMUCR
	case 0xFF00001C:
		return c.CCR
	case 0xFF000020:
		return c.TRA
	case 0xFF000024:
		return c.EXPEVT
	case 0xFF000028:
		return c.INTEVT
	case 0xFF000030: // PVR: the processor names itself; boot code checks
		return 0x040205C1
	case 0xFF000038:
		return c.QACR0
	case 0xFF00003C:
		return c.QACR1
	case 0xFFD00000:
		return c.ICR
	case 0xFFD00004:
		return c.IPRA
	case 0xFFD00008:
		return c.IPRB
	case 0xFFD0000C:
		return c.IPRC
	}
	if v, ok := c.onchipRaw[addr]; ok {
		return v
	}
	c.onchipGaps[addr]++
	return 0
}

func (c *CPU) onchipWrite(addr uint32, size int, v uint32) {
	if c.tmuWrite(addr, v) {
		return
	}
	switch addr {
	case scifSCFTDR:
		c.SerialTX = append(c.SerialTX, uint8(v))
		return
	case scifSCFSR:
		return // flag-clearing writes; our flags are derived, not stored
	case 0xFF000000:
		c.PTEH = v
	case 0xFF000004:
		c.PTEL = v
	case 0xFF000008:
		c.TTB = v
	case 0xFF00000C:
		c.TEA = v
	case 0xFF000010:
		c.MMUCR = v
		if v&1 != 0 {
			// Retail software runs untranslated; a set AT bit means either a
			// path this model has not met or a corrupted boot, and both are
			// worth stopping for.
			c.Halt("MMUCR.AT enabled at PC %08X — the MMU is unmodelled", c.curPC)
		}
	case 0xFF00001C:
		c.CCR = v &^ 0x0808 // the flush bits read back as 0, self-clearing
	case 0xFF000020:
		c.TRA = v
	case 0xFF000024:
		c.EXPEVT = v
	case 0xFF000028:
		c.INTEVT = v
	case 0xFF000038:
		c.QACR0 = v
	case 0xFF00003C:
		c.QACR1 = v
	case 0xFFD00000:
		c.ICR = v
	case 0xFFD00004:
		c.IPRA = v
	case 0xFFD00008:
		c.IPRB = v
	case 0xFFD0000C:
		c.IPRC = v
	default:
		c.onchipRaw[addr] = v
	}
}

// Gaps lists the on-chip registers read without ever having been modeled or
// written, most-read first: the worklist for onchip.go.
func (c *CPU) Gaps() []string {
	type gap struct {
		addr  uint32
		reads int
	}
	var gs []gap
	for a, n := range c.onchipGaps {
		gs = append(gs, gap{a, n})
	}
	sort.Slice(gs, func(i, j int) bool { return gs[i].reads > gs[j].reads })
	var out []string
	for _, g := range gs {
		out = append(out, fmt.Sprintf("onchip read %08X x%d (never written)", g.addr, g.reads))
	}
	return out
}
