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
		info := make([]uint32, 28)
		info[0] = 0x01000000 // FT: CONTROLLER (big-endian function mask, the bus convention)
		info[1] = 0x000F06FE // FD: the standard pad's capability word
		name := "RetroReverse Dreamcast Pad    "
		for i := 0; i < 28 && i < len(name); i++ {
			info[3+uint32(i)/4] |= uint32(name[i]) << (8 * (uint32(i) % 4))
		}
		m.Write32(recv, 5|src<<8|host<<16|uint32(len(info))<<24)
		for i, w := range info {
			m.Write32(recv+4+uint32(i)*4, w)
		}
	case 4: // GETCOND -> 8: data transfer
		m.Write32(recv, 8|src<<8|host<<16|3<<24)
		m.Write32(recv+4, 0x01000000) // the function replying
		m.Write32(recv+8, uint32(m.Pad.Buttons)|uint32(m.Pad.RT)<<16|uint32(m.Pad.LT)<<24)
		m.Write32(recv+12, uint32(m.Pad.JoyX)|uint32(m.Pad.JoyY)<<8|0x80<<16|0x80<<24)
	default:
		// The controller exists but does not speak this function: the bus's
		// "function code unsupported" reply (0xFE), never the empty-socket
		// word — a device must not vanish between commands.
		m.logf("maple command %d answered as unsupported", cmd)
		m.Write32(recv, 0xFE|src<<8|host<<16)
	}
}
