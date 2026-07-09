package n64

// si.go is the Serial Interface and the joybus behind it.
//
// The PIF holds a 64-byte block of RAM. To read the controllers or the save
// EEPROM, the CPU DMAs a command block into PIF RAM, sets the PIF's execute bit,
// and DMAs the results back — blocking, in libultra, on the SI interrupt that
// signals each transfer complete. osContInit and osEepromProbe both run this way
// during boot, so a machine that performs the DMA but never raises the interrupt
// stalls exactly as one that does nothing.
//
// The command block is a byte stream, one variable-length record per joybus
// channel:
//
//	0xFE        end of the command list
//	0xFF        channel skip / padding (also used as a NOP byte)
//	0x00        this channel is skipped
//	tx rx cmd…  a command: tx bytes out, rx bytes expected back
//
// The four controller channels come first, then channel 4 addresses the
// cartridge's save device. A channel with nothing attached must report so, or
// the game waits forever for a controller that will never answer.

const (
	siDramAddr  = 0x00
	siReadAddr  = 0x04 // write: PIF RAM -> RDRAM
	siWriteAddr = 0x10 // write: RDRAM -> PIF RAM
	siStatus    = 0x18
)

// SI_STATUS bits. A write to the status register acknowledges the interrupt.
const (
	siStatusDMABusy   = 1 << 0
	siStatusIOBusy    = 1 << 1
	siStatusInterrupt = 1 << 12
)

// pifRAMSize is the joybus mailbox at the top of the PIF's address space.
const pifRAMSize = 64

// Joybus commands.
const (
	jbInfo            = 0x00 // identify the device on this channel
	jbControllerState = 0x01
	jbReadAccessory   = 0x02
	jbWriteAccessory  = 0x03
	jbEepromRead      = 0x04
	jbEepromWrite     = 0x05
	jbReset           = 0xFF // same reply as Info
)

// Device identifiers returned by the Info command, high byte first.
const (
	devController uint16 = 0x0500 // a standard controller
)

// Controller buttons, in the order the joybus reports them (two bytes).
const (
	BtnA      = 1 << 15
	BtnB      = 1 << 14
	BtnZ      = 1 << 13
	BtnStart  = 1 << 12
	BtnDUp    = 1 << 11
	BtnDDown  = 1 << 10
	BtnDLeft  = 1 << 9
	BtnDRight = 1 << 8
	BtnL      = 1 << 5
	BtnR      = 1 << 4
	BtnCUp    = 1 << 3
	BtnCDown  = 1 << 2
	BtnCLeft  = 1 << 1
	BtnCRight = 1 << 0
)

// Fields are exported so encoding/gob carries them into a save-state.
type si struct {
	Regs     regFile
	DramAddr uint32
}

func (s *si) init() { s.Regs = regFile{} }

// Controller is the state of one of the four ports.
type Controller struct {
	Present bool
	Buttons uint16 // active-high, unlike the PSX pad
	StickX  int8
	StickY  int8
}

func (m *Machine) siRead(addr uint32) uint32 {
	switch addr & 0xFF {
	case siDramAddr:
		return m.si.DramAddr
	case siStatus:
		// The DMA has already completed by the time the write that started it
		// returns, so the busy bits are never observed set.
		var v uint32
		if m.mi.Intr&intrSI != 0 {
			v |= siStatusInterrupt
		}
		return v
	}
	return m.si.Regs[addr&0xFF]
}

func (m *Machine) siWrite(addr uint32, v uint32) {
	switch addr & 0xFF {
	case siDramAddr:
		m.si.DramAddr = v & 0x00FFFFFF
	case siWriteAddr: // RDRAM -> PIF RAM, then the PIF executes the block
		for i := 0; i < pifRAMSize; i++ {
			m.PIF[i] = m.RDRAM[(m.si.DramAddr+uint32(i))%uint32(len(m.RDRAM))]
		}
		m.joybus()
		m.raiseIRQ(intrSI)
	case siReadAddr: // PIF RAM -> RDRAM
		m.joybus()
		for i := 0; i < pifRAMSize; i++ {
			m.RDRAM[(m.si.DramAddr+uint32(i))%uint32(len(m.RDRAM))] = m.PIF[i]
		}
		m.raiseIRQ(intrSI)
	case siStatus:
		m.clearIRQ(intrSI) // any write acknowledges
	default:
		m.si.Regs[addr&0xFF] = v
	}
}

// joybus walks the command block in PIF RAM and answers each channel in place.
//
// The PIF only runs the block when the CPU sets bit 0 of the last byte; libultra
// writes that byte as part of the block it DMAs in. Answering unconditionally
// would be wrong for the "read back the results" transfer, which must not
// re-execute — so the flag is honoured and cleared, as the hardware does.
func (m *Machine) joybus() {
	if m.PIF[pifRAMSize-1]&1 == 0 {
		return
	}
	m.PIF[pifRAMSize-1] &^= 1

	ch := 0
	for i := 0; i < pifRAMSize-1; {
		tx := m.PIF[i]
		switch tx {
		case 0xFE: // end of the command list
			return
		case 0xFF, 0x00: // padding, or a channel to skip
			if tx == 0x00 {
				ch++
			}
			i++
			continue
		}
		if i+1 >= pifRAMSize-1 {
			return
		}
		rx := m.PIF[i+1] & 0x3F
		cmdAt := i + 2
		if cmdAt >= pifRAMSize-1 {
			return
		}
		resAt := cmdAt + int(tx&0x3F)
		if resAt+int(rx) > pifRAMSize-1 {
			return
		}
		// The receive-length byte is where the PIF reports a channel's error flags.
		m.joybusChannel(ch, m.PIF[cmdAt:resAt], m.PIF[resAt:resAt+int(rx)], i+1)
		i = resAt + int(rx)
		ch++
	}
}

// joybusChannel answers one command. res is the reply buffer; rxAt indexes the
// receive-length byte, whose high bits carry the error flags.
func (m *Machine) joybusChannel(ch int, cmd, res []byte, rxAt int) {
	if len(cmd) == 0 {
		return
	}
	// Channels 0..3 are the controller ports; channel 4 is the save device.
	if ch >= 4 {
		m.joybusEEPROM(cmd, res, rxAt)
		return
	}
	pad := m.Controllers[ch]

	switch cmd[0] {
	case jbInfo, jbReset:
		if !pad.Present {
			// Nothing attached: set the "no response" flag the PIF reports in the
			// top bits of the receive-length byte. Without it the game waits.
			m.PIF[rxAt] |= 0x80
			return
		}
		if len(res) >= 3 {
			res[0] = byte(devController >> 8)
			res[1] = byte(devController & 0xFF)
			res[2] = 0 // no accessory in the controller's expansion slot
		}
	case jbControllerState:
		if !pad.Present {
			m.PIF[rxAt] |= 0x80
			return
		}
		if len(res) >= 4 {
			res[0] = byte(pad.Buttons >> 8)
			res[1] = byte(pad.Buttons)
			res[2] = byte(pad.StickX)
			res[3] = byte(pad.StickY)
		}
	case jbReadAccessory, jbWriteAccessory:
		// No Controller Pak or Rumble Pak is fitted.
		m.PIF[rxAt] |= 0x80
	default:
		m.note("joybus: unmodelled command 0x%02X on controller channel %d", cmd[0], ch)
		m.PIF[rxAt] |= 0x80
	}
}
