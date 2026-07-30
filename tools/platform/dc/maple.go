package dc

// maple.go is the controller bus, driven the way the hardware is: the game
// builds a command table in RAM, points SB_MDSTAR at it, and starts the DMA;
// the responses land in the receive buffers each descriptor names and the
// Maple-DMA-end interrupt fires. One standard controller answers on port A;
// every other address gets the no-connection word, which is exactly what an
// empty port says.
//
// The pad state is injectable (Machine.Pad), which is what a -keys flag
// needs: buttons are active-low in the condition response, the hardware's
// own convention.

import "encoding/binary"

// Maple SB registers.
const (
	sbMDSTAR = 0x005F6C04
	sbMDTSEL = 0x005F6C10
	sbMDEN   = 0x005F6C14
	sbMDST   = 0x005F6C18
)

const istMapleDMA = 1 << 12

// PadState is the port-A controller's condition, in hardware conventions:
// Buttons is active-low (a pressed button clears its bit).
type PadState struct {
	Buttons uint16 // active low; 0xFFFF = nothing pressed
	LT, RT  uint8  // analog triggers, 0 released
	JoyX    uint8  // 0x80 centered
	JoyY    uint8
}

// Button bits in PadState.Buttons (clear = pressed).
const (
	PadC     = 1 << 0
	PadB     = 1 << 1
	PadA     = 1 << 2
	PadStart = 1 << 3
	PadUp    = 1 << 4
	PadDown  = 1 << 5
	PadLeft  = 1 << 6
	PadRight = 1 << 7
	PadZ     = 1 << 8
	PadY     = 1 << 9
	PadX     = 1 << 10
	PadD     = 1 << 11
)

// bswap flips a word into the bus's big-endian byte order.
func bswap(v uint32) uint32 {
	return v<<24 | v>>24 | v<<8&0xFF0000 | v>>8&0xFF00
}

// mapleState is the register file; it savestates.
type mapleState struct {
	MDSTAR, MDTSEL, MDEN uint32
}

// mapleRead serves the 005F6C00 block.
func (m *Machine) mapleRead(addr uint32) uint32 {
	switch addr {
	case sbMDSTAR:
		return m.Maple.MDSTAR
	case sbMDTSEL:
		return m.Maple.MDTSEL
	case sbMDEN:
		return m.Maple.MDEN
	case sbMDST:
		return 0 // DMA completes within the write that starts it
	}
	m.logf("maple read %08X (PC %08X)", addr, m.CPU.CurPC())
	return 0
}

func (m *Machine) mapleWrite(addr, v uint32) {
	switch addr {
	case sbMDSTAR:
		m.Maple.MDSTAR = v & 0x1FFFFFE0
	case sbMDTSEL:
		m.Maple.MDTSEL = v
	case sbMDEN:
		m.Maple.MDEN = v & 1
	case sbMDST:
		if v&1 != 0 && m.Maple.MDEN != 0 {
			m.mapleDMA()
		}
	default:
		m.logf("maple write %08X = %08X (PC %08X)", addr, v, m.CPU.CurPC())
	}
}

// mapleDMA walks the command table and answers every frame.
func (m *Machine) mapleDMA() {
	addr := m.Maple.MDSTAR
	for i := 0; i < 64; i++ { // a sane descriptor bound; real tables are short
		ctrl := m.ram32(addr)
		recv := m.ram32(addr+4) & 0x1FFFFFFF
		frame := m.ram32(addr + 8)
		words := frame >> 24 & 0xFF
		payload := addr + 12
		addr = payload + 4*words

		cmd := frame & 0xFF
		dst := frame >> 8 & 0xFF
		m.mapleRespond(cmd, dst, payload, recv)
		if cmd != 1 && cmd != 4 {
			m.logf("maple frame cmd %d dst %02X fn %08X arg %08X recv %08X", cmd, dst, m.ram32(payload), m.ram32(payload+4), recv)
		}

		if ctrl&(1<<31) != 0 {
			break
		}
	}
	m.raiseNRM(istMapleDMA)
}

// mapleRespond writes one frame's response. The only device is port A's main
// peripheral (AP 0x20); everyone else is an empty socket.
func (m *Machine) mapleRespond(cmd, dst, payload, recv uint32) {
	if dst != 0x20 {
		m.Write32(recv, 0xFFFFFFFF) // no connection
		return
	}
	src := uint32(0x20)
	host := uint32(0x00)
	switch cmd {
	case 1: // DEVINFO request -> 5: device status
		// The device-status block, in the hardware's own layout: FT, three FD
		// words, destination area code, connector direction, 30 bytes of
		// product name, 60 of license, standby and maximum current. Crazy
		// Taxi's driver byte-reverses the FT word it receives and tests the
		// CONTROLLER bit in the low byte, then classifies the device by
		// product-name bytes ('R' and 'F' prefixes select the wheel and rod
		// paths) — so the block carries the retail pad's own constants, and
		// the FT goes out raw (wire bytes 00 00 00 01), not pre-swapped.
		var blk [112]byte
		binary.LittleEndian.PutUint32(blk[0:], 0x01000000)        // FT: CONTROLLER
		binary.LittleEndian.PutUint32(blk[4:], bswap(0x000F06FE)) // FD1: the standard pad's capability word
		blk[16] = 0xFF                                            // destination area: all regions
		blk[17] = 0                                               // connector direction
		copy(blk[18:], "Dreamcast Controller          ")
		copy(blk[48:], "Produced By or Under License From SEGA ENTERPRISES,LTD.    ")
		binary.LittleEndian.PutUint16(blk[108:], 0x01AE) // standby current, 0.1mA units
		binary.LittleEndian.PutUint16(blk[110:], 0x01F4) // maximum current
		m.Write32(recv, 5|host<<8|src<<16|uint32(len(blk)/4)<<24) // responses answer TO the host FROM the device
		for i := uint32(0); i < uint32(len(blk)); i += 4 {
			m.Write32(recv+4+i, binary.LittleEndian.Uint32(blk[i:]))
		}
	case 9: // GET CONDITION -> 8: data transfer (command 4 is device shutdown,
		// not GETCOND — the game polls 9 every field, and answering it
		// "unsupported" was a controller that never felt a button)
		m.Write32(recv, 8|host<<8|src<<16|3<<24)
		m.Write32(recv+4, bswap(0x01000000)) // the function replying, bus byte order
		m.Write32(recv+8, uint32(m.Pad.Buttons)|uint32(m.Pad.RT)<<16|uint32(m.Pad.LT)<<24)
		m.Write32(recv+12, uint32(m.Pad.JoyX)|uint32(m.Pad.JoyY)<<8|0x80<<16|0x80<<24)
	default:
		// The controller exists but does not speak this function: the bus's
		// "function code unsupported" reply (0xFE), never the empty-socket
		// word — a device must not vanish between commands.
		m.logf("maple command %d answered as unsupported", cmd)
		m.Write32(recv, 0xFE|host<<8|src<<16)
	}
}
