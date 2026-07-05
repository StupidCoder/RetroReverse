package dos

// Minimal PC I/O-port models, enough to get UW past its hardware initialisation
// (it takes over the keyboard, VGA and timer directly). None of this produces
// real output — the point is to answer the status polls the game blocks on so
// its logic keeps running under the oracle:
//
//   - 8042 keyboard controller (ports 0x60/0x64): report "ready to accept", and
//     acknowledge keyboard/controller commands with the expected response byte,
//     so the command handshakes complete instead of spinning.
//   - VGA input status (0x3DA): toggle the vertical-retrace/display-enable bits
//     each read, so retrace-wait loops progress.
//   - PIT (0x40-0x43) and PIC (0x20/0xA0): accept writes, return advancing/zero
//     reads.

import "retroreverse.com/tools/x86"

type ioState struct {
	kbdOut     byte // 8042 output buffer (byte the game will read from 0x60)
	kbdOutFull bool
	expectData bool   // a 0x64 command expecting a following 0x60 data byte
	retrace    bool   // 0x3DA toggle
	pit        uint16 // rotating PIT counter read value
	oplReg     byte   // last OPL (FM) register selected
	oplTimer   bool   // OPL timer-1 running (drives the detection status bits)

	tick uint32 // instructions since the last injected timer tick

	// Sound Blaster Pro (base 0x220): mixer (read-back) + DSP reset/command.
	mixReg   byte
	mixRegs  [256]byte
	dspQueue []byte // bytes the DSP will return via 0x22A
	dspReset byte   // last value written to the reset port 0x226

	// VGA DAC (ports 3C8/3C9): the 256-colour palette the game programs.
	dacIndex int // current register × 3 + component cursor
	Pal      [768]byte

	seen map[uint16]bool
}

func (m *Machine) portIn(port uint16, size int) uint32 {
	io := m.io
	switch port {
	case 0x60: // keyboard data
		io.kbdOutFull = false
		return uint32(io.kbdOut)
	case 0x64: // keyboard controller status: bit0 = output full, bit1 = input full (0 = ready)
		s := uint32(0)
		if io.kbdOutFull {
			s |= 0x01
		}
		return s
	case 0x3DA, 0x3BA: // VGA/MDA input status #1: bit0 = display disabled, bit3 = vertical retrace
		io.retrace = !io.retrace
		if io.retrace {
			return 0x09
		}
		return 0x00
	case 0x40, 0x41, 0x42: // PIT counter latch — return an advancing value
		io.pit -= 0x137
		return uint32(io.pit>>8) & 0xFF
	case 0x228, 0x388: // OPL2/OPL3 FM status: bits 7,6 report timer expiry
		if io.oplTimer {
			return 0xC0 // both timers expired — the "OPL present" signature
		}
		return 0x00
	case 0x201: // joystick/game port: buttons up, no joystick
		return 0xF0
	case 0x225: // SB Pro mixer data: read back what was written (detection)
		return uint32(io.mixRegs[io.mixReg])
	case 0x22A: // SB DSP read-data port
		if len(io.dspQueue) > 0 {
			v := io.dspQueue[0]
			io.dspQueue = io.dspQueue[1:]
			return uint32(v)
		}
		return 0
	case 0x22C: // SB DSP write-buffer status: bit7 = 0 means ready to write
		return 0x00
	case 0x22E: // SB DSP read-buffer status: bit7 = data available
		if len(io.dspQueue) > 0 {
			return 0x80
		}
		return 0x00
	case 0x20, 0x21, 0xA0, 0xA1: // PIC
		return 0
	default:
		if io.seen == nil {
			io.seen = map[uint16]bool{}
		}
		if !io.seen[port] {
			io.seen[port] = true
			m.logf("IN port $%03X (unmodeled) at %04X:%04X", port, m.CPU.Seg[1], m.CPU.IP)
		}
		return widthMask8(size)
	}
}

func (m *Machine) portOut(port uint16, size int, v uint32) {
	// A word OUT to a VGA index port is the classic index+data pair
	// (e.g. `MOV AX,$0C0C; OUT DX,AX` programs CRTC register AL with AH).
	if size == 2 && (port == 0x3C4 || port == 0x3CE || port == 0x3D4) {
		m.portOut(port, 1, v&0xFF)
		m.portOut(port+1, 1, (v>>8)&0xFF)
		return
	}
	io := m.io
	b := byte(v)
	switch port {
	case 0x3C4, 0x3C5, 0x3CE, 0x3CF, 0x3D4, 0x3D5: // VGA sequencer/GC/CRTC
		m.vgaRegOut(port, b)
	case 0x60: // data to keyboard (or a controller command's data byte)
		if io.expectData {
			io.expectData = false // consumed by the previous controller command
			return
		}
		io.kbdOut, io.kbdOutFull = 0xFA, true // keyboard ACK
	case 0x64: // command to the 8042 controller
		switch b {
		case 0xAA: // controller self-test
			io.kbdOut, io.kbdOutFull = 0x55, true
		case 0xAB: // interface test
			io.kbdOut, io.kbdOutFull = 0x00, true
		case 0xEE: // echo
			io.kbdOut, io.kbdOutFull = 0xEE, true
		case 0x60, 0xD1, 0xD2, 0xD3, 0xD4: // commands that take a following data byte
			io.expectData = true
		}
	case 0x228, 0x388: // OPL FM register-select port
		io.oplReg = b
	case 0x229, 0x389: // OPL FM data port
		if io.oplReg == 0x04 { // timer-control register
			if b&0x80 != 0 {
				io.oplTimer = false // reset / IRQ reset
			} else if b&0x01 != 0 {
				io.oplTimer = true // start timer 1 (detection expects it to "expire")
			}
		}
	case 0x3C8: // VGA DAC write index
		io.dacIndex = int(b) * 3
	case 0x3C9: // VGA DAC data: R,G,B 6-bit components, auto-advancing
		io.Pal[io.dacIndex%768] = b
		io.dacIndex++
	case 0x224: // SB Pro mixer register-select
		io.mixReg = b
	case 0x225: // SB Pro mixer data (stored for read-back)
		io.mixRegs[io.mixReg] = b
	case 0x226: // SB DSP reset: a 1 then 0 pulse makes the DSP report 0xAA ready
		if io.dspReset == 1 && b == 0 {
			io.dspQueue = append(io.dspQueue, 0xAA)
		}
		io.dspReset = b
	case 0x22C: // SB DSP command port
		switch b {
		case 0xE1: // get DSP version -> 3.01 (SB Pro)
			io.dspQueue = append(io.dspQueue, 0x03, 0x01)
		}
	}
	// All other ports (PIT, PIC, VGA, DMA, DSP) accept and discard.
}

func widthMask8(size int) uint32 {
	if size == 2 {
		return 0xFFFF
	}
	return 0xFF
}

// onStep is the CPU per-instruction hook: it injects a periodic timer interrupt
// (IRQ0 / INT 8) and advances the BIOS tick counter at 0040:006C, so the game's
// timer-driven initialisation and main loop make progress. The tick period is
// in instructions, not real time — the oracle just needs ticks to *happen*.
const ticksEveryInstrs = 800

func (m *Machine) onStep(c *x86.CPU) {
	if !m.EnableIRQ {
		return
	}
	io := m.io
	io.tick++
	// Keyboard injection rides a HALF-tick phase offset from the timer. Delivering
	// it here — not at the tick wrap below — matters: the timer's Interrupt(8)
	// dispatch clears IF for the duration of its ISR, so a keyboard IRQ raised in
	// the same breath would always find interrupts masked and never land. At the
	// half-tick the game is in its normal flow (IF set ~half the time), exactly
	// when a real IRQ1 would be accepted.
	if io.tick == ticksEveryInstrs/2 {
		m.pumpKeys()
		return
	}
	if io.tick < ticksEveryInstrs {
		return
	}
	io.tick = 0
	// Advance the BIOS timer-tick dword at 0040:006C.
	t := uint32(m.r16(0x46C)) | uint32(m.r16(0x46E))<<16
	t++
	m.w16(0x46C, uint16(t))
	m.w16(0x46E, uint16(t>>16))
	// Deliver IRQ0 only if the program installed a handler (IVT[8] segment set).
	if m.r16(8*4+2) != 0 {
		c.Interrupt(8)
	}
}
