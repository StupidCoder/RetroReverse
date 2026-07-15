package gc

import (
	"fmt"
	"os"
)

// siTrace, set by RR_GC_SITRACE, dumps every SI register access — the tool that pinned the
// register map and that verifies the transfer/poll protocol as it runs.
var siTrace = os.Getenv("RR_GC_SITRACE") != ""

func siLog(format string, a ...any) { fmt.Fprintf(os.Stderr, format, a...) }

// si.go is the Serial Interface: the four controller ports, the transfer engine that
// carries a command out to a controller and its answer back, and the once-a-field auto-poll
// that keeps each pad's state fresh. It is what turns a plugged-in standard controller into
// bytes the game can read, and it is the last device the boot needs before it stops waiting.
//
// The boot's wait, and why this device is what ends it, is worth stating because it looked
// for a long time like a hang: Luigi's Mansion boots cleanly all the way to its first
// interactive screen and then correctly parks, waiting for the player to press Start or A.
// The game's scene director advances only when the pad object's per-frame event field carries
// button bits 0x1100 (START|A). Those never arrive because PADRead skips every port: PADReset
// marks each channel "reset pending" (its bit set in a mask that read back 0xF0000000) and a
// port clears only when its reset/probe transfer *completes and raises the SI interrupt* so
// PADReset's callback can record the controller type. A serial interface that completes no
// transfer and raises no interrupt leaves every port pending forever. So this device has to
// actually run the protocol.
//
// The register map, pinned empirically from the boot's own accesses (RR_GC_SITRACE), base
// 0xCC006400:
//
//	+0x00/+0x04/+0x08   channel 0 OUTBUF / INBUFH / INBUFL   (each channel is 12 bytes)
//	+0x0C.. +0x24..     channels 1, 2, 3
//	+0x30               SIPOLL    — the auto-poll configuration (which channels, how often)
//	+0x34               SICOMCSR  — the transfer command/status (TSTART, channel, lengths)
//	+0x38               SISR      — per-channel status (one byte each: RDST + error nibble)
//	+0x3C               SIEXILK
//	+0x80.. (128 bytes) the transfer I/O buffer: the command goes out of it, the reply into it
//
// SICOMCSR is at +0x34, not +0x30 — the poll register is what sits at +0x30. (This corrects
// an earlier note; the boot's TSTART command 0xC0010301 lands at +0x34 and the poll config
// 0x00F70200 at +0x30, which the trace makes unambiguous.) The command word packs, MSB-first:
// bit0 an always-set enable, bit1 TCINTMSK (set when a completion callback is registered),
// bits9-15 the output length, bits17-23 the input length (either length 0 meaning 128),
// bits29-30 the channel, and bit31 TSTART. On completion TSTART clears, bit0 (TCINT) latches,
// and the SI interrupt is raised if TCINTMSK is set — the game's SIInterruptHandler then reads
// the reply out of the I/O buffer, acknowledges TCINT by writing bit0 back, and invokes the
// per-channel callback.

// SICOMCSR bits.
const (
	siTStart   = 0x00000001 // bit31: start the transfer (auto-clears when done)
	siTCInt    = 0x80000000 // bit0:  transfer-complete status (write-one-to-clear)
	siTCIntMsk = 0x40000000 // bit1:  raise the SI interrupt on completion
)

// SISR per-channel status byte. Each channel occupies one byte of the word (channel 0 the
// most significant), and PADRead reads a channel's byte to decide whether the last poll or
// transfer produced usable data: RDST (0x20) set with the error nibble (0x0F) clear means a
// fresh, valid response; a set error bit means the port did not answer.
const (
	siRDST       = 0x20 // read status: a response is present in this channel's registers
	siNoResponse = 0x08 // the low-nibble "no controller answered" error
)

type si struct {
	// The channel registers: [channel][0]=OUTBUF, [1]=INBUFH, [2]=INBUFL. The auto-poll
	// writes INBUFH/INBUFL; SIGetResponse reads them.
	Chan   [4][3]uint32
	Poll   uint32    // +0x30 SIPOLL
	ComCSR uint32    // +0x34 SICOMCSR
	Status uint32    // +0x38 SISR
	ExiLk  uint32    // +0x3C
	IOBuf  [128]byte // +0x80 the transfer I/O buffer

	// The controllers plugged into the four ports.
	Pad [4]padPort
}

// padPort is one standard controller. Buttons is the digital button halfword the pad reports
// (START=0x1000, A=0x0100, B=0x0200, X=0x0400, Y=0x0800, and the rest); the analog fields are
// absolute 8-bit samples, the sticks centred at 0x80 and the triggers at rest at 0.
type padPort struct {
	Connected                                      bool
	Buttons                                        uint16
	StickX, StickY, SubX, SubY, TriggerL, TriggerR uint8
}

// connectPad plugs a standard controller into a port with its sticks physically centred, the
// resting state a real pad reports when nothing is touched.
func (d *si) connectPad(port int) {
	d.Pad[port] = padPort{Connected: true, StickX: 0x80, StickY: 0x80, SubX: 0x80, SubY: 0x80}
}

func (d *si) read(m *Machine, off uint32, size int) uint32 {
	r := off & 0xFF
	v := d.readReg(r, size)
	if siTrace {
		siLog("SI rd 0x%02X -> 0x%08X (pc 0x%08X)\n", r, v, m.CPU.PC)
	}
	return v
}

func (d *si) readReg(r uint32, size int) uint32 {
	switch {
	case r < 0x30:
		c, w := r/12, (r%12)/4
		return d.Chan[c][w]
	case r == 0x30:
		return d.Poll
	case r == 0x34:
		return d.ComCSR
	case r == 0x38:
		return d.Status
	case r == 0x3C:
		return d.ExiLk
	case r >= 0x80 && r < 0x100:
		return d.iobufRead(r-0x80, size)
	}
	return 0
}

func (d *si) write(m *Machine, off uint32, v uint32, size int) {
	r := off & 0xFF
	if siTrace {
		siLog("SI wr 0x%02X = 0x%08X (pc 0x%08X)\n", r, v, m.CPU.PC)
	}
	switch {
	case r < 0x30:
		c, w := r/12, (r%12)/4
		d.Chan[c][w] = v
	case r == 0x30:
		d.Poll = v
	case r == 0x34:
		d.writeComCSR(m, v)
	case r == 0x38:
		// SITransfer masks SISR to one channel's bits before a transfer; the game owns the
		// register, so honour the write. The transfer engine sets the completion bits after.
		d.Status = v
	case r == 0x3C:
		d.ExiLk = v
	case r >= 0x80 && r < 0x100:
		d.iobufWrite(r-0x80, v, size)
	default:
		m.logf("SI write unmodelled 0x%02X = 0x%08X", r, v)
	}
}

// writeComCSR handles a write to the command/status register: it acknowledges a completion
// interrupt (write-one-to-clear on TCINT), records the command bits, and, if the write starts
// a transfer, runs it.
func (d *si) writeComCSR(m *Machine, v uint32) {
	// A write with TCINT set acknowledges the completed transfer's interrupt. Resolve that
	// first, then take the remaining control bits from the write.
	tcint := d.ComCSR & siTCInt
	if v&siTCInt != 0 {
		tcint = 0
	}
	d.ComCSR = tcint | (v &^ siTCInt)
	if v&siTStart != 0 {
		d.startTransfer(m)
		return
	}
	m.siRefreshIRQ()
}

// startTransfer runs one immediate transfer: it sends the channel's command out of the I/O
// buffer and writes the controller's reply back into it, sets the channel's SISR status, and
// completes with the transfer-complete interrupt.
func (d *si) startTransfer(m *Machine) {
	c := (d.ComCSR >> 1) & 3
	inlen := (d.ComCSR >> 8) & 0x7F
	if inlen == 0 {
		inlen = 128
	}
	cmd := d.IOBuf[0]
	shift := (3 - c) * 8

	d.Status &^= 0xFF << shift // clear this channel's status byte
	if d.Pad[c].Connected {
		resp := d.Pad[c].respond(cmd, int(inlen))
		copy(d.IOBuf[:], resp)
		d.Status |= siRDST << shift // a valid reply is present, error nibble clear
	} else {
		// Nothing on this port: the interface times out and reports no response, leaving the
		// read buffer as it was.
		d.Status |= siNoResponse << shift
	}

	// The transfer is done: TSTART clears, TCINT latches, and the SI line is asserted if the
	// completion interrupt is enabled.
	d.ComCSR = (d.ComCSR &^ siTStart) | siTCInt
	m.siRefreshIRQ()
}

// respond builds a standard controller's reply to one SI command. The command byte is the
// first byte of the output the game placed in the I/O buffer; the reply length is what the
// command register asked for. These are the standard-controller wire formats — a device ID
// for the reset/probe, the neutral origin for the calibration read, and the button+analog
// status for a poll — so the game's own SIGetType/PADReset/PADRead parse them unchanged.
func (p *padPort) respond(cmd byte, inlen int) []byte {
	buf := make([]byte, inlen)
	switch cmd {
	case 0x00, 0xFF: // reset / get-ID: the three-byte SI device identifier
		if inlen >= 3 {
			// The GameCube standard controller's ID.
			buf[0], buf[1], buf[2] = 0x09, 0x00, 0x00
		}
	default: // 0x40 read status, 0x41 read origin, 0x42 recalibrate: the button+analog status
		hi, lo := p.status()
		putWord(buf, 0, hi)
		putWord(buf, 4, lo)
	}
	return buf
}

// status packs the pad's live state into the two response words a poll returns: the button
// halfword and the two stick samples in the first, the C-stick and both triggers in the
// second. It is exactly what lands in INBUFH/INBUFL for SIGetResponse to read.
func (p *padPort) status() (hi, lo uint32) {
	hi = uint32(p.Buttons)<<16 | uint32(p.StickX)<<8 | uint32(p.StickY)
	lo = uint32(p.SubX)<<24 | uint32(p.SubY)<<16 | uint32(p.TriggerL)<<8 | uint32(p.TriggerR)
	return
}

// tickSI is the auto-poll: once per video field, while polling is enabled, it latches each
// connected pad's status into that channel's input registers and marks the channel's SISR
// status "response ready". PADRead rides this the way the rest of the game rides the retrace.
func (m *Machine) tickSI() {
	if m.si.Poll == 0 { // polling disabled: SIPOLL is zero out of reset and during PADReset
		return
	}
	for c := 0; c < 4; c++ {
		p := &m.si.Pad[c]
		if !p.Connected {
			continue
		}
		hi, lo := p.status()
		m.si.Chan[c][1] = hi
		m.si.Chan[c][2] = lo
		shift := uint(3-c) * 8
		m.si.Status = (m.si.Status &^ (0xFF << shift)) | (siRDST << shift)
	}
}

// siRefreshIRQ follows the SI interrupt line from the command register: the line is asserted
// while a transfer's completion is latched (TCINT) and its interrupt is enabled (TCINTMSK),
// and it drops when the handler acknowledges TCINT. Level-driven, like every other cause.
func (m *Machine) siRefreshIRQ() {
	if m.si.ComCSR&siTCInt != 0 && m.si.ComCSR&siTCIntMsk != 0 {
		m.raiseInt(IntSI)
	} else {
		m.clearInt(IntSI)
	}
}

// --- controller API (used by the oracle's -keys injection) --------------------------------

// SetController plugs a standard controller into a port or unplugs it.
func (m *Machine) SetController(port int, connected bool) {
	if port < 0 || port > 3 {
		return
	}
	if connected {
		m.si.connectPad(port)
	} else {
		m.si.Pad[port] = padPort{}
	}
}

// SetPadButtons sets a connected controller's digital buttons (START=0x1000, A=0x0100, …).
// The next auto-poll latches them where PADRead will find them.
func (m *Machine) SetPadButtons(port int, buttons uint16) {
	if port < 0 || port > 3 {
		return
	}
	m.si.Pad[port].Buttons = buttons
}

// --- I/O buffer byte access ---------------------------------------------------------------

func (d *si) iobufRead(o uint32, size int) uint32 {
	switch size {
	case 1:
		return uint32(d.IOBuf[o])
	case 2:
		return uint32(d.IOBuf[o])<<8 | uint32(d.IOBuf[o+1])
	default:
		return uint32(d.IOBuf[o])<<24 | uint32(d.IOBuf[o+1])<<16 | uint32(d.IOBuf[o+2])<<8 | uint32(d.IOBuf[o+3])
	}
}

func (d *si) iobufWrite(o uint32, v uint32, size int) {
	switch size {
	case 1:
		d.IOBuf[o] = byte(v)
	case 2:
		d.IOBuf[o], d.IOBuf[o+1] = byte(v>>8), byte(v)
	default:
		d.IOBuf[o], d.IOBuf[o+1], d.IOBuf[o+2], d.IOBuf[o+3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
	}
}

func putWord(b []byte, o int, v uint32) {
	if o+3 < len(b) {
		b[o], b[o+1], b[o+2], b[o+3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
	}
}
