package gc

// di.go is the Disc Interface: the drive, as the game's own code drives it.
//
// A GameCube reads its disc by writing a command into three registers, a memory address
// and a length into two more, and a "go" bit into the control register; the drive DMAs
// the bytes into memory and raises an interrupt. There is exactly one command the boot
// path uses — read (0xA8) — and its parameters are a disc offset (in 4-byte units, a
// hangover from the drive's sector addressing) and a length.
//
// This is where the filesystem parser earns its place. Every read arrives here as a raw
// offset and a length and nothing else, and OnDVDRead resolves the offset back to the file
// it lands in — so the oracle can report not "the game read 0x4B970B08" but "the game read
// /Ajioka/ADemo/codemo01.szp". That is the single most useful thing the machine can say
// while it runs, because it is the game telling us, in its own words, what it is loading.

type di struct {
	SR     uint32    // status: the interrupt flags and their masks
	Cover  uint32    // the cover/lid state
	Cmd    [3]uint32 // the command buffer: opcode and two parameters
	MAR    uint32    // the memory address to DMA into
	Length uint32    // how many bytes
	CR     uint32    // control: the "go" bit and the DMA direction
	ImmBuf uint32    // the immediate result of a non-DMA command
	Cfg    uint32
}

// The DI status-register bits.
const (
	diBreakInt  = 1 << 6 // a break completed
	diBreakMask = 1 << 5
	diTCInt     = 1 << 4 // transfer complete
	diTCMask    = 1 << 3
	diErrInt    = 1 << 2 // a device error
	diErrMask   = 1 << 1
	diBreak     = 1 << 0
)

func (d *di) init() {
	// The lid is closed and a disc is present. Bit 0 of the cover register is the lid
	// state; a program checks it before trusting the drive, and an open lid takes a
	// "please insert a disc" path that never reaches the game.
	d.Cover = 0 // 0 = closed, disc present
	d.SR = diTCMask | diErrMask | diBreakMask
}

func (d *di) read(m *Machine, off uint32, size int) uint32 {
	switch off & 0xFF {
	case 0x00:
		return d.SR
	case 0x04:
		return d.Cover
	case 0x08, 0x0C, 0x10:
		return d.Cmd[(off&0xFF-0x08)/4]
	case 0x14:
		return d.MAR
	case 0x18:
		return d.Length
	case 0x1C:
		return d.CR
	case 0x20:
		return d.ImmBuf
	case 0x24:
		return d.Cfg
	}
	m.logf("DI read unmodelled 0x%02X", off&0xFF)
	return 0
}

func (d *di) write(m *Machine, off uint32, v uint32, size int) {
	switch off & 0xFF {
	case 0x00:
		// The status register: the mask bits are written directly, and writing a 1 to an
		// interrupt-flag bit acknowledges it.
		d.SR &^= v & (diBreakInt | diTCInt | diErrInt)
		d.SR = (d.SR &^ (diBreakMask | diTCMask | diErrMask)) | (v & (diBreakMask | diTCMask | diErrMask))
		if v&diBreakInt != 0 || v&diTCInt != 0 || v&diErrInt != 0 {
			m.diRefreshIRQ()
		}
	case 0x04:
		d.Cover &^= v & 0x4 // the cover-change interrupt flag
	case 0x08, 0x0C, 0x10:
		d.Cmd[(off&0xFF-0x08)/4] = v
	case 0x14:
		d.MAR = v & 0x03FFFFE0
	case 0x18:
		d.Length = v & 0x03FFFFE0
	case 0x1C:
		d.CR = v
		if v&1 != 0 { // TSTART: begin the command
			d.exec(m)
		}
	case 0x24:
		d.Cfg = v
	default:
		m.logf("DI write unmodelled 0x%02X = 0x%08X", off&0xFF, v)
	}
}

// exec runs a disc command. Only the read is implemented; anything else is logged so it
// becomes a work item rather than a silent success.
func (d *di) exec(m *Machine) {
	opcode := d.Cmd[0] >> 24
	switch opcode {
	case 0xA8: // read
		// The offset is in 4-byte units: the drive addresses the disc in words, not bytes.
		discOff := int64(d.Cmd[2]) << 2
		length := d.Length
		if m.OnDVDRead != nil {
			m.OnDVDRead(discOff, length, d.MAR)
		}
		if m.disc != nil {
			data, err := m.disc.Read(discOff, int(length))
			if err != nil {
				m.logf("DI read past the disc: offset 0x%X length %d: %v", discOff, length, err)
				d.raiseError(m)
				return
			}
			m.dmaToRAM(d.MAR, data)
		}
		d.complete(m)
	case 0xE0: // request error / get status — answered from the immediate buffer
		d.ImmBuf = 0
		d.complete(m)
	case 0x12: // inquiry: the drive reports its model and firmware
		// A small, plausible identity so a program that inquires does not stall. The
		// bytes are the drive's own; a game reads them only to confirm a drive is there.
		m.dmaToRAM(d.MAR, []byte{0x00, 0x00, 0x00, 0x00, 0x20, 0x01, 0x06, 0x03,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		d.complete(m)
	default:
		m.logf("DI unimplemented command 0x%02X (0x%08X 0x%08X 0x%08X)", opcode, d.Cmd[0], d.Cmd[1], d.Cmd[2])
		d.complete(m) // complete it anyway, so the game does not wait forever on a stub
	}
}

// complete finishes a command: clear the go bit, set transfer-complete, and interrupt if
// the game armed it.
func (d *di) complete(m *Machine) {
	d.CR &^= 1
	d.Length = 0
	d.SR |= diTCInt
	m.diRefreshIRQ()
}

func (d *di) raiseError(m *Machine) {
	d.CR &^= 1
	d.SR |= diErrInt
	m.diRefreshIRQ()
}

// diRefreshIRQ raises or lowers the disc's line from its flags and masks.
func (m *Machine) diRefreshIRQ() {
	s := m.di.SR
	pending := (s&diTCInt != 0 && s&diTCMask != 0) ||
		(s&diErrInt != 0 && s&diErrMask != 0) ||
		(s&diBreakInt != 0 && s&diBreakMask != 0)
	if pending {
		m.raiseInt(IntDI)
	} else {
		m.clearInt(IntDI)
	}
}
