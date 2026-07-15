package gc

// exi.go is the External Interface: a small serial bus with three channels, off which
// hang the memory cards, the modem, and — the reason this device matters at boot — the
// real-time clock and the console's SRAM.
//
// The SRAM is a landmine, named in the package doc as one of the three console-resident
// substitutions. It holds the console's settings: the video mode, the language, the sound
// mode, and a checksum over them. OSInit and VIInit read it early, and if it is absent or
// its checksum does not verify, the machine does not fail loudly — it takes a "the
// settings are corrupt, ask the user" path that a booting game never returns from. So the
// SRAM is synthesized here, with a valid checksum, before the first instruction runs.
//
// The real-time clock is read the same way and matters almost as much: OSInit reads it to
// bias the time base, and a program that computed a delay from it against a clock that
// never advanced would wait forever.
//
// EXI is a shift register: software selects a channel and a device, writes a command into
// the data register, sets the transfer-control register's "go" bit, and reads the reply
// out of the same data register. So a transaction is a tiny state machine per channel, and
// most of this file is that machine for the two devices boot depends on.

import (
	"fmt"
	"os"
)

var exiTrace = os.Getenv("RR_GC_EXITRACE") != ""

type exiChannel struct {
	CSR     uint32 // channel status/control: the selected device and the interrupts
	Data    uint32 // the shift register
	CR      uint32 // transfer control: length, direction, and the go bit
	DMAAddr uint32 // main-memory address for a DMA transfer
	DMALen  uint32 // DMA length in bytes
	Dev     int    // the selected device
	Phase   int    // words into a multi-word transaction
	Command uint32
}

// The channel status register's bit layout: interrupt masks and their write-one-to-clear
// status bits, the clock divisor, the one-hot chip select, and the read-only device-present
// bit. The three status bits and the present bit must survive a control write — the SDK
// acknowledges an interrupt by writing its bit back, not by rewriting the whole register.
const (
	exiCSRTCIntMask = 1 << 2  // transfer-complete interrupt enable
	exiCSRTCInt     = 1 << 3  // transfer-complete status (write one to clear)
	exiCSRStatus    = 1<<1 | 1<<3 | 1<<11 // EXIINT, TCINT, EXTINT — the w1c bits
)

type exi struct {
	Ch   [3]exiChannel
	SRAM [64]byte // the console settings block
	RTC  uint32   // seconds since the GameCube epoch (2000-01-01)
}

func (e *exi) init() {
	// Synthesize a valid SRAM. The layout: a 16-bit checksum, its inverse, a 32-bit RTC
	// bias, then the settings bytes. The checksum is a sum of the settings words; getting
	// it right is what keeps OSInit on the happy path.
	//
	// The settings themselves are the defaults a boxed console ships with: NTSC video,
	// English, stereo sound, no progressive scan.
	s := e.SRAM[:]
	// Offsets 6.. are the flags block. 0x08 language (0 = English), 0x0A flags.
	s[0x0C] = 0x00              // sound mode / flags
	writeBE16(s[0x08:], 0x0000) // language English + the mode flags in the low bits

	// The checksum is over the 8 words that follow the two checksum halves.
	var sum uint16
	for i := 4; i < 20; i += 2 {
		sum += beU16(s[i:])
	}
	writeBE16(s[0x00:], sum)
	writeBE16(s[0x02:], ^sum)

	e.RTC = 0x30000000 // an arbitrary but non-zero clock, well past the epoch
}

// The device numbers on channel 0.
const (
	exiDevMemCard = 0 // channel 0, device 0
	exiDevRTC     = 1 // channel 0, device 1: the RTC and the SRAM sit behind one device
	exiDevNone    = -1
)

func (e *exi) read(m *Machine, off uint32, size int) uint32 {
	ch := (off & 0xFF) / 0x14
	if ch >= 3 {
		m.logf("EXI read unmodelled 0x%02X", off&0xFF)
		return 0
	}
	c := &e.Ch[ch]
	switch (off & 0xFF) % 0x14 {
	case 0x00:
		return c.CSR
	case 0x04:
		return c.DMAAddr
	case 0x08:
		return c.DMALen
	case 0x0C:
		return c.CR
	case 0x10:
		return c.Data
	}
	m.logf("EXI read unmodelled channel %d 0x%02X", ch, (off&0xFF)%0x14)
	return 0
}

func (e *exi) write(m *Machine, off uint32, v uint32, size int) {
	ch := (off & 0xFF) / 0x14
	if ch >= 3 {
		m.logf("EXI write unmodelled 0x%02X = 0x%08X", off&0xFF, v)
		return
	}
	if exiTrace {
		fmt.Fprintf(os.Stderr, "  EXI wr ch%d 0x%02X = 0x%08X (pc 0x%08X)\n", ch, (off&0xFF)%0x14, v, m.CPU.PC)
	}
	c := &e.Ch[ch]
	switch (off & 0xFF) % 0x14 {
	case 0x00:
		// The channel status/control register. Bit 7 downward selects the chip; writing
		// zero to the select deasserts it, which ends a transaction. The three interrupt
		// status bits acknowledge (clear) when written with a one and otherwise persist.
		ack := v & exiCSRStatus
		c.CSR = (v &^ exiCSRStatus) | (c.CSR & exiCSRStatus &^ ack)
		if v&(0x380) == 0 { // no device selected
			c.Dev = exiDevNone
			c.Phase = 0
		} else {
			// Bits 7-9 are a one-hot device select on this channel.
			switch (v >> 7) & 7 {
			case 1:
				c.Dev = exiDevMemCard
			case 2:
				c.Dev = exiDevRTC
			default:
				c.Dev = exiDevNone
			}
			c.Phase = 0
		}
		e.refreshIRQ(m)
	case 0x04:
		c.DMAAddr = v & 0x03FFFFE0
	case 0x08:
		c.DMALen = v & 0x03FFFFE0
	case 0x0C:
		c.CR = v
		if v&1 != 0 { // the transfer go bit
			if v&2 != 0 { // DMA mode: bits 2-3 are the direction, 0 read / 1 write
				e.dmaTransfer(m, ch, (v>>2)&3 == 1)
			} else {
				e.transfer(m, ch)
			}
			c.CR &^= 1
			// Completing a transfer raises the transfer-complete status; whether it also
			// interrupts the CPU is the mask's decision. The SDK's asynchronous SRAM sync
			// and the memory-card driver both ride this interrupt — a transfer that
			// completes without it leaves their callbacks waiting forever.
			c.CSR |= exiCSRTCInt
			e.refreshIRQ(m)
		}
	case 0x10:
		c.Data = v
	default:
		m.logf("EXI write unmodelled channel %d 0x%02X = 0x%08X", ch, (off&0xFF)%0x14, v)
	}
}

// refreshIRQ recomputes the EXI line to the CPU: any channel with an interrupt status bit
// whose mask (the bit below it) is set holds the line up.
func (e *exi) refreshIRQ(m *Machine) {
	pending := false
	for i := range e.Ch {
		csr := e.Ch[i].CSR
		// Each status bit (1, 3, 11) is enabled by the mask bit directly below it (0, 2, 10);
		// shifting the register left one aligns each mask under its status.
		if csr&(csr<<1)&exiCSRStatus != 0 {
			pending = true
		}
	}
	if pending {
		m.raiseInt(IntEXI)
	} else {
		m.clearInt(IntEXI)
	}
}

// transfer runs one EXI word exchange for the selected device.
func (e *exi) transfer(m *Machine, ch uint32) {
	c := &e.Ch[ch]
	switch c.Dev {
	case exiDevRTC:
		e.rtcTransfer(m, c)
	case exiDevMemCard:
		// No card is fitted. A memory-card probe reads back an all-ones ID, which a
		// program reads as "no device", and it moves on rather than waiting.
		c.Data = 0xFFFFFFFF
	default:
		c.Data = 0
	}
}

// dmaTransfer runs a whole-block exchange between main memory and the selected device — the
// form the SDK uses for the 64-byte SRAM block (an immediate command word, then one DMA).
// asWrite is the direction from the CPU's side: true pushes main memory into the device.
func (e *exi) dmaTransfer(m *Machine, ch uint32, asWrite bool) {
	c := &e.Ch[ch]
	addr, n := c.DMAAddr, c.DMALen
	if int(addr)+int(n) > len(m.RAM) {
		m.logf("EXI DMA out of range 0x%08X+0x%X", addr, n)
		return
	}
	switch c.Dev {
	case exiDevRTC:
		dev := rtcDevAddr(c.Command)
		if dev < 4 {
			m.logf("EXI DMA to RTC device address %d (command 0x%08X) unmodelled", dev, c.Command)
			return
		}
		off := dev - 4 // SRAM starts at device address 4
		for i := uint32(0); i < n && off+i < uint32(len(e.SRAM)); i++ {
			if asWrite {
				e.SRAM[off+i] = m.RAM[addr+i]
			} else {
				m.RAM[addr+i] = e.SRAM[off+i]
			}
		}
	case exiDevMemCard:
		// No card is fitted: a read shifts in all-ones, and a write falls off the end of
		// the bus. The probe reads an ID of ones and reports the slot empty.
		if !asWrite {
			for i := uint32(0); i < n; i++ {
				m.RAM[addr+i] = 0xFF
			}
		}
	default:
		m.logf("EXI DMA with no device selected (channel %d)", ch)
	}
}

// rtcDevAddr extracts the device address from a clock-chip command word: bits 6 and up
// (the 0x20000000 command prefix masked away). Address 0 is the RTC counter; the SRAM
// block begins at 4 — the observed SRAM-block command is 0x20000100, device address 4.
func rtcDevAddr(cmd uint32) uint32 { return (cmd >> 6) & 0x7FFF }

// rtcTransfer is the clock-and-SRAM device's protocol, one immediate word at a time: the
// first word is a command — bit 31 selects write, bits 6.. the device address — and the
// words after it carry the data, the device address advancing a word per exchange.
func (e *exi) rtcTransfer(m *Machine, c *exiChannel) {
	if c.Phase == 0 {
		c.Command = c.Data
		c.Phase = 1
		// The reply to the command word is not meaningful; the data comes next.
		return
	}
	write := c.Command&0x80000000 != 0
	dev := rtcDevAddr(c.Command)
	word := uint32(c.Phase-1) * 4
	c.Phase++
	switch {
	case dev == 0: // the RTC counter
		if !write {
			c.Data = e.RTC
		}
	case dev >= 4 && dev-4+word+3 < uint32(len(e.SRAM)): // the SRAM, a word per exchange
		idx := dev - 4 + word
		if write {
			writeBE32(e.SRAM[idx:], c.Data)
		} else {
			c.Data = beU32(e.SRAM[idx:])
		}
	default:
		// The flash, the UART, the diagnostics: read back zero, and note it once so an
		// unexpected access is visible rather than silently satisfied.
		m.logf("EXI RTC-device access at device address %d (command 0x%08X) unmodelled", dev, c.Command)
		if !write {
			c.Data = 0
		}
	}
}

func beU16(b []byte) uint16 { return uint16(b[0])<<8 | uint16(b[1]) }
func beU32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
func writeBE16(b []byte, v uint16) {
	b[0], b[1] = uint8(v>>8), uint8(v)
}
func writeBE32(b []byte, v uint32) {
	b[0], b[1], b[2], b[3] = uint8(v>>24), uint8(v>>16), uint8(v>>8), uint8(v)
}
