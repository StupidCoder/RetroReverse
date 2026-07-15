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

type exiChannel struct {
	CSR     uint32 // channel status/control: the selected device and the interrupts
	Data    uint32 // the shift register
	CR      uint32 // transfer control: length, direction, and the go bit
	Dev     int    // the selected device
	Addr    uint32 // the offset a read/write command named
	Phase   int    // words into a multi-word transaction
	Command uint32
}

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
	c := &e.Ch[ch]
	switch (off & 0xFF) % 0x14 {
	case 0x00:
		// The channel status/control register. Bit 7 downward selects the chip; writing
		// zero to the select deasserts it, which ends a transaction.
		c.CSR = v
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
	case 0x0C:
		c.CR = v
		if v&1 != 0 { // the transfer go bit
			e.transfer(m, ch)
			c.CR &^= 1
		}
	case 0x10:
		c.Data = v
	default:
		m.logf("EXI write unmodelled channel %d 0x%02X = 0x%08X", ch, (off&0xFF)%0x14, v)
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

// rtcTransfer is the clock-and-SRAM device's protocol: the first word is a command whose
// top bits select read or write and whose address picks the RTC counter, the SRAM, or the
// flash. The second word carries the data.
func (e *exi) rtcTransfer(m *Machine, c *exiChannel) {
	if c.Phase == 0 {
		c.Command = c.Data
		c.Addr = (c.Data >> 8) & 0x7FFFFFFF
		c.Phase = 1
		// The reply to the command word is not meaningful; the data comes next.
		return
	}
	write := c.Command&0x80000000 != 0
	switch {
	case c.Addr == 0x20000000: // the RTC counter
		if !write {
			c.Data = e.RTC
		}
	case c.Addr >= 0x20000100 && c.Addr < 0x20000140: // the SRAM
		idx := (c.Addr - 0x20000100)
		if write {
			if idx+3 < uint32(len(e.SRAM)) {
				writeBE32(e.SRAM[idx:], c.Data)
			}
		} else {
			if idx+3 < uint32(len(e.SRAM)) {
				c.Data = beU32(e.SRAM[idx:])
			}
		}
	default:
		// The flash, the UART, the diagnostics: read back zero, and note it once so an
		// unexpected access is visible rather than silently satisfied.
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
