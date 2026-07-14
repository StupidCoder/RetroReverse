package dsmachine

import (
	"encoding/binary"
	"unicode/utf16"

	"retroreverse.com/tools/platform/nds"
)

// The ARM7's SPI bus, and the three devices on it: the power-management chip, the
// 256 KiB firmware flash, and the touchscreen controller. One chip-select field in
// SPICNT says which is listening; bytes go out and come back through SPIDATA, one
// at a time, with the transaction held open across bytes by the "chip-select hold"
// bit.
//
// The firmware is the interesting one. It is not just a boot ROM: it carries the
// *user settings* — the owner's name, birthday, favourite colour, language, and
// the touchscreen calibration — and a DS game reads them at startup. NitroSDK does
// not read them trustingly: it takes the two copies of the settings block, checks
// each one's CRC, and picks the newer valid one. Hand it a firmware of zeroes and
// both copies fail their checksum, and the boot does not proceed with defaults —
// it proceeds with nothing.
//
// So we synthesise a firmware that is *correct*, not merely present. It is
// generated rather than dumped: none of it is Nintendo's, all of it is what the
// documented format says those bytes mean.

const (
	regSPICNT  = 0x040001C0
	regSPIDATA = 0x040001C2

	firmwareSize  = 0x40000 // 256 KiB
	userSettings1 = 0x3FE00
	userSettings2 = 0x3FF00
)

// SPI device select (SPICNT bits 8..9).
const (
	spiPower = 0
	spiFirm  = 1
	spiTouch = 2
)

// Touchscreen calibration, as written into the synthesised firmware. The touch
// controller must invert exactly this mapping (see touchSample), or every stylus
// point the game computes lands somewhere other than where we put it.
const (
	calADCX1, calADCY1 = 0x02DF, 0x032C
	calScrX1, calScrY1 = 32, 32
	calADCX2, calADCY2 = 0x0D3B, 0x0CE7
	calScrX2, calScrY2 = 224, 158
)

// spibus holds the transaction in flight. Every device on this bus is a state
// machine over a *byte sequence* rather than a set of addressable registers, so
// what matters is which byte of the current transaction we are on: `phase` counts
// bytes since chip-select went active, and `cmd` remembers the one that opened it.
type spibus struct {
	firmware []byte

	dev   int
	phase int
	cmd   byte
	addr  uint32 // the flash address a firmware read is walking
	out   byte   // the byte the last exchange produced; SPIDATA reads it back

	chanSel   int // the touchscreen channel the last control byte selected
	resultIdx int // which result byte of the current conversion is next

	touchX, touchY int // where the stylus is held (screen pixels)
	touchDown      bool
}

func newSPI() *spibus {
	s := &spibus{firmware: make([]byte, firmwareSize)}
	s.buildFirmware()
	return s
}

// buildFirmware synthesises a firmware flash whose header and user-settings blocks
// are internally consistent — above all, whose CRCs check out.
func (s *spibus) buildFirmware() {
	f := s.firmware
	le := binary.LittleEndian

	// Header. The field a game genuinely needs is at 0x20: the offset of the
	// user-settings area, stored divided by eight.
	le.PutUint16(f[0x20:], uint16(userSettings1/8))
	copy(f[0x08:], []byte("MACP")) // firmware identifier
	f[0x1D] = 0xFF                 // console type: an original DS

	var us [0x100]byte
	us[0x00] = 5 // settings version
	us[0x02] = 4 // favourite colour
	us[0x03] = 7 // birthday month
	us[0x04] = 4 // birthday day

	name := utf16.Encode([]rune("RETRO"))
	for i, r := range name {
		le.PutUint16(us[0x06+i*2:], r)
	}
	le.PutUint16(us[0x1A:], uint16(len(name)))
	le.PutUint16(us[0x50:], 0) // empty personal message

	// Touchscreen calibration: two reference points, each a raw ADC reading paired
	// with the screen pixel it corresponds to.
	le.PutUint16(us[0x58:], calADCX1)
	le.PutUint16(us[0x5A:], calADCY1)
	us[0x5C] = calScrX1
	us[0x5D] = calScrY1
	le.PutUint16(us[0x5E:], calADCX2)
	le.PutUint16(us[0x60:], calADCY2)
	us[0x62] = calScrX2
	us[0x63] = calScrY2

	// Language and flags: bits 0..2 select the language (1 = English).
	le.PutUint16(us[0x64:], 0x8001|(1<<3)|(1<<6))

	// The two copies are distinguished by an update counter — the higher one is the
	// live settings — and each is checksummed over its first 0x70 bytes.
	for i, off := range []int{userSettings1, userSettings2} {
		blk := us
		le.PutUint16(blk[0x70:], uint16(i)) // counter 0, then 1: the second one wins
		le.PutUint16(blk[0x72:], nds.CRC16(blk[0x00:0x70]))
		copy(f[off:], blk[:])
	}
}

// SetTouch positions the stylus (screen pixel coordinates), or lifts it.
func (m *Machine) SetTouch(x, y int, down bool) {
	m.spi.touchX, m.spi.touchY, m.spi.touchDown = x, y, down
}

// spiTransfer exchanges one byte with the selected device and latches the reply.
func (c *core) spiTransfer(v byte) {
	s := c.m.spi
	cnt := c.io[regSPICNT]
	if cnt&0x8000 == 0 {
		return // the bus is disabled
	}
	dev := int(cnt>>8) & 3
	if dev != s.dev {
		s.dev, s.phase = dev, 0
	}
	if s.phase == 0 {
		s.cmd = v
	}

	switch dev {
	case spiFirm:
		s.out = s.firmwareByte(v)
	case spiTouch:
		s.out = s.touchByte(v)
	default: // power management: writes take effect, reads read back zero
		s.out = 0
	}
	s.phase++

	// The chip-select hold bit (11) keeps the transaction open for the next byte.
	// When it is clear, the device is deselected as soon as this byte completes and
	// its state machine rewinds — that is what *ends* a firmware read.
	if cnt&0x0800 == 0 {
		s.phase = 0
		s.dev = -1
	}
	if cnt&0x4000 != 0 { // transfer-complete interrupt
		c.raise(irqSPI)
	}
}

// firmwareByte is the flash's command state machine. Only the commands a boot
// issues are answered.
func (s *spibus) firmwareByte(v byte) byte {
	switch s.cmd {
	case 0x03: // READ: a 24-bit big-endian address, then a stream of data bytes
		switch {
		case s.phase == 0:
			s.addr = 0
			return 0
		case s.phase <= 3:
			s.addr = s.addr<<8 | uint32(v)
			return 0
		default:
			b := byte(0)
			if int(s.addr) < len(s.firmware) {
				b = s.firmware[s.addr]
			}
			s.addr++
			return b
		}
	case 0x05: // READ STATUS: zero means "ready, no write in progress"
		return 0x00
	case 0x9F: // JEDEC id
		ids := []byte{0x20, 0x40, 0x12}
		if s.phase >= 1 && s.phase <= 3 {
			return ids[s.phase-1]
		}
		return 0
	}
	return 0 // write-enable/disable and anything else: no reply
}

// touchByte is the touchscreen controller. Software writes a control byte (bit 7
// starts a conversion, bits 4..6 name the channel) and then clocks two more bytes out
// to receive the 12-bit result: the FIRST carries bits 11..5, the second the low five
// bits left-aligned. Big-endian, in other words, like everything else on this bus.
//
// The order matters and cannot be inferred from the byte count: swap the two and every
// coordinate the game computes is a plausible number in the right range and the wrong
// place, so the stylus works, reports a position, and never lands where you put it.
//
// The result index is counted from the last control byte rather than from the start of
// the transaction, because the driver holds chip-select across several conversions and
// clocks them back to back — control, high, low, control, high, low — so parity of the
// byte position within the transaction is not the same thing at all.
func (s *spibus) touchByte(v byte) byte {
	if v&0x80 != 0 { // a control byte: latch the channel, begin a conversion
		s.chanSel = int(v>>4) & 7
		s.resultIdx = 0
		return 0
	}
	val := s.touchSample(s.chanSel)
	s.resultIdx++
	if s.resultIdx == 1 {
		return byte(val >> 5) // bits 11..5
	}
	return byte(val<<3) & 0xF8 // bits 4..0, left-aligned in the byte
}

// touchSample converts the held stylus position back into the ADC reading the
// controller would have produced for it — the inverse of the calibration the game
// is about to apply. Reporting pixels directly would be a lie the game then
// "calibrates" into the wrong place.
func (s *spibus) touchSample(ch int) uint16 {
	if !s.touchDown {
		return 0 // an untouched panel: no pressure, floating position
	}
	switch ch {
	case 1: // Y
		return uint16(calADCY1 + (s.touchY-calScrY1)*(calADCY2-calADCY1)/(calScrY2-calScrY1))
	case 5: // X
		return uint16(calADCX1 + (s.touchX-calScrX1)*(calADCX2-calADCX1)/(calScrX2-calScrX1))
	case 3, 4: // Z1/Z2 — the pressure channels; any sane pair reads as "touched"
		return 0x0200
	}
	return 0
}
