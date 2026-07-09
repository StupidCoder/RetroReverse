package n64

// eeprom.go is the cartridge's save device, reached over the joybus on channel 4
// rather than through the cartridge bus.
//
// Pilotwings 64 carries a 4 kbit EEPROM: 512 bytes in 64 blocks of 8. libultra's
// osEepromProbe runs an Info command on that channel during boot and waits on the
// SI interrupt for the reply; a channel that never answers leaves the probe
// spinning.
//
// The contents are the player's saved data, so they belong to whoever runs the
// oracle rather than to the repository. Nothing here loads or stores a file: the
// device starts erased, and a caller that wants persistence can read and write
// Machine.EEPROM directly.

const (
	eepromBlockSize = 8
	eeprom4KBlocks  = 64  // 4 kbit
	eeprom16KBlocks = 256 // 16 kbit
)

// EEPROM device identifiers, as the Info command reports them.
const (
	devEEPROM4K  uint16 = 0x0080
	devEEPROM16K uint16 = 0x00C0
)

func (m *Machine) joybusEEPROM(cmd, res []byte, rxAt int) {
	if len(m.EEPROM) == 0 {
		m.PIF[rxAt] |= 0x80 // no save device fitted
		return
	}
	switch cmd[0] {
	case jbInfo, jbReset:
		if len(res) >= 3 {
			id := devEEPROM4K
			if len(m.EEPROM) > eeprom4KBlocks*eepromBlockSize {
				id = devEEPROM16K
			}
			res[0] = byte(id >> 8)
			res[1] = byte(id)
			res[2] = 0 // not write-busy
		}
	case jbEepromRead:
		if len(cmd) < 2 || len(res) < eepromBlockSize {
			m.PIF[rxAt] |= 0x40
			return
		}
		off := int(cmd[1]) * eepromBlockSize
		if off+eepromBlockSize > len(m.EEPROM) {
			m.PIF[rxAt] |= 0x40
			return
		}
		copy(res, m.EEPROM[off:off+eepromBlockSize])
	case jbEepromWrite:
		if len(cmd) < 2+eepromBlockSize {
			m.PIF[rxAt] |= 0x40
			return
		}
		off := int(cmd[1]) * eepromBlockSize
		if off+eepromBlockSize > len(m.EEPROM) {
			m.PIF[rxAt] |= 0x40
			return
		}
		copy(m.EEPROM[off:off+eepromBlockSize], cmd[2:2+eepromBlockSize])
		if len(res) >= 1 {
			res[0] = 0 // write complete
		}
	default:
		m.note("joybus: unmodelled EEPROM command 0x%02X", cmd[0])
		m.PIF[rxAt] |= 0x80
	}
}
