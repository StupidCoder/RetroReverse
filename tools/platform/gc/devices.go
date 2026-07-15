package gc

import (
	"fmt"
	"os"
)

var siTrace = os.Getenv("RR_GC_SITRACE") != ""

// devices.go holds the register blocks that are, for now, thin: they keep their registers
// so a read returns what was written, they satisfy the accesses the boot path makes, and
// they log once anything the model does not yet answer for. Each will grow its own file
// when a phase needs it — the pixel engine and the command processor when the graphics
// pipe is built, the audio interface when sound is wanted. Until then, a thin device that
// does not stall the boot is exactly right, and a stub that lied would not be.

// --- MI: the Memory Interface -------------------------------------------------------
//
// It configures the memory protection regions and reports memory errors. The boot path
// programs it and moves on; nothing here needs to do more than remember what it was told.

type mi struct {
	Reg [0x40]uint16
}

func (d *mi) read(m *Machine, off uint32, size int) uint32 {
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		return uint32(d.Reg[i])
	}
	m.logf("MI read unmodelled 0x%03X", off&0xFFF)
	return 0
}

func (d *mi) write(m *Machine, off uint32, v uint32, size int) {
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		d.Reg[i] = uint16(v)
		return
	}
	m.logf("MI write unmodelled 0x%03X = 0x%08X", off&0xFFF, v)
}

// --- SI: the Serial Interface -------------------------------------------------------
//
// It polls the four controller ports. No controller is attached yet, so a poll reads back
// "nothing there". For the boot up to the title that was harmless — but it is exactly what
// now gates further progress, and the mechanism is worth recording because it looked for a
// long time like a hang.
//
// Luigi's Mansion boots cleanly all the way through: the apploader runs, the DOL loads, the
// four boot loader-tasks stream every asset off the disc (game_usa.szp + all the audio
// banks — 486 reads), and the game's scene director reaches its first *interactive* screen
// and **waits for the player to press Start or A**. That wait is not a bug and not an I/O
// race (an earlier theory blamed instant disc completion; pacing the drive changes nothing —
// all four tasks complete and are consumed, the last by the director's own PHASE-1 poll).
// It is the game correctly parked on a press-Start screen. Proof: forcing the director's
// PHASE-2 advance makes it immediately load /Kawano/ENGLISH/res_slct.szp — the select screen.
//
// The director advances when the pad object's per-frame event field carries button bits
// 0x1100 (START|A). It never does because PADRead skips every port: PADReset leaves each
// channel's bit set in its "reset pending" mask (0xF0000000) and only the SI reset/probe
// *completing* — the transfer raising its interrupt so PADReset's callback clears the bit
// and records a standard-controller type — takes a port out of that mask. Our SI completes
// no transfers and raises no interrupt, so the ports stay perpetually pending.
//
// So the real next step is a faithful SI: model the transfer engine (SICOMCSR at +0x30 with
// its TSTART/channel/length fields, SISR at +0x38 for SIGetStatus's per-channel OK bit) and
// the VBLANK auto-poll (SIPOLL at +0x34) so that a connected standard controller answers
// each poll into the channel INBUF registers (+0x04/+0x08, read by SIGetResponse at
// 0x801E0010), raises the SI interrupt, and lets the game's own PADReset/PADRead deliver a
// scripted Start press. RR_GC_SITRACE dumps every SI register access for that work.

type si struct {
	Reg [0x40]uint32
}

func (d *si) read(m *Machine, off uint32, size int) uint32 {
	r := off & 0xFF
	if siTrace {
		v := uint32(0)
		if r < 0x30 {
			v = 0
		} else if int(r/4) < len(d.Reg) {
			v = d.Reg[r/4]
		}
		fmt.Fprintf(os.Stderr, "SI rd 0x%02X -> 0x%08X (pc 0x%08X)\n", r, v, m.CPU.PC)
	}
	// The four channel input registers report the pad state. Zero — no buttons, sticks
	// centred at zero — is a plausible idle pad; a game reads it without stalling.
	if r < 0x30 {
		return 0
	}
	switch r {
	case 0x30:
		// The poll/status register. Report "no response error", so a game does not treat
		// the empty ports as a fault.
		return 0
	}
	i := r / 4
	if int(i) < len(d.Reg) {
		return d.Reg[i]
	}
	m.logf("SI read unmodelled 0x%02X", r)
	return 0
}

func (d *si) write(m *Machine, off uint32, v uint32, size int) {
	if siTrace {
		fmt.Fprintf(os.Stderr, "SI wr 0x%02X = 0x%08X (pc 0x%08X)\n", off&0xFF, v, m.CPU.PC)
	}
	i := (off & 0xFF) / 4
	if int(i) < len(d.Reg) {
		d.Reg[i] = v
		return
	}
	m.logf("SI write unmodelled 0x%02X = 0x%08X", off&0xFF, v)
}

// --- AI: the Audio Interface --------------------------------------------------------
//
// The streaming DAC's sample clock and volume. Sound is a later phase; for now the
// registers are kept so a program that configures the DAC and reads back its settings sees
// what it wrote, and the sample-count register advances so a poll on it makes progress.

type ai struct {
	Control uint32
	Volume  uint32
	SCnt    uint32 // the running sample counter
	ITCnt   uint32 // the interrupt-trigger count
}

func (d *ai) read(m *Machine, off uint32, size int) uint32 {
	switch off & 0xFF {
	case 0x00:
		return d.Control
	case 0x04:
		return d.Volume
	case 0x08:
		d.SCnt++ // a poll on the sample counter must see it move
		return d.SCnt
	case 0x0C:
		return d.ITCnt
	}
	m.logf("AI read unmodelled 0x%02X", off&0xFF)
	return 0
}

func (d *ai) write(m *Machine, off uint32, v uint32, size int) {
	switch off & 0xFF {
	case 0x00:
		d.Control = v
	case 0x04:
		d.Volume = v
	case 0x08:
		d.SCnt = v
	case 0x0C:
		d.ITCnt = v
	default:
		m.logf("AI write unmodelled 0x%02X = 0x%08X", off&0xFF, v)
	}
}

// --- CP: the Command Processor ------------------------------------------------------
//
// The graphics FIFO's control: where it reads from, how full it is, and whether it is
// running. The graphics pipe is Phase 3; here the CP keeps its registers and reports the
// FIFO as empty and idle, so a program that waits for the pipe to drain is not told it is
// perpetually busy.

type cp struct {
	Status  uint16
	Control uint16
	Clear   uint16
	Reg     [0x40]uint16
}

func (d *cp) read(m *Machine, off uint32, size int) uint32 {
	switch off & 0xFFF {
	case 0x00:
		// The status register. Report the FIFO empty and the pipe idle: bit for
		// "read-idle" and "command-idle" set, "overflow"/"underflow" clear.
		return 0x0006
	case 0x02:
		return uint32(d.Control)
	}
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		return uint32(d.Reg[i])
	}
	m.logf("CP read unmodelled 0x%03X", off&0xFFF)
	return 0
}

func (d *cp) write(m *Machine, off uint32, v uint32, size int) {
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		d.Reg[i] = uint16(v)
		return
	}
	m.logf("CP write unmodelled 0x%03X = 0x%08X", off&0xFFF, v)
}

// --- PE: the Pixel Engine -----------------------------------------------------------
//
// The end of the graphics pipe, and the source of the two interrupts a frame-timed game
// waits on: the token (a marker the game plants in the command stream) and the finish
// (the pipe has drained). The command processor (gpu.go) reaches those markers as it walks
// the FIFO and calls setToken/setFinish here; this is where the interrupt is actually
// raised, gated by the enables the game programmed, and later acknowledged.
//
// The control register at 0x0A is the whole of that protocol. Its low two bits enable the
// token and finish interrupts; its next two are write-one-to-clear acknowledgements that a
// handler uses to dismiss the interrupt it is servicing. The interrupt line itself is the
// shared PE cause in the processor interface, so raising and clearing here is level-driven
// through pi.go exactly as every other device is.

type pe struct {
	Reg          [0x20]uint16
	Token        uint16
	TokenEnable  bool
	FinishEnable bool
}

func (d *pe) read(m *Machine, off uint32, size int) uint32 {
	switch off & 0xFFF {
	case 0x0A:
		// The control register reads back the enables it was given and the pending status
		// of each interrupt, which is what a handler inspects to tell token from finish.
		var v uint16
		if d.TokenEnable {
			v |= 1 << 0
		}
		if d.FinishEnable {
			v |= 1 << 1
		}
		if m.pi.Cause&(1<<IntPEToken) != 0 {
			v |= 1 << 2
		}
		if m.pi.Cause&(1<<IntPEFinish) != 0 {
			v |= 1 << 3
		}
		return uint32(v)
	case 0x0E:
		return uint32(d.Token) // the last token the pipe passed
	}
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		return uint32(d.Reg[i])
	}
	m.logf("PE read unmodelled 0x%03X", off&0xFFF)
	return 0
}

func (d *pe) write(m *Machine, off uint32, v uint32, size int) {
	switch off & 0xFFF {
	case 0x0A:
		d.TokenEnable = v&(1<<0) != 0
		d.FinishEnable = v&(1<<1) != 0
		if v&(1<<2) != 0 { // acknowledge the token interrupt
			m.clearInt(IntPEToken)
		}
		if v&(1<<3) != 0 { // acknowledge the finish interrupt
			m.clearInt(IntPEFinish)
		}
		return
	}
	i := (off & 0xFFF) / 2
	if int(i) < len(d.Reg) {
		d.Reg[i] = uint16(v)
		return
	}
	m.logf("PE write unmodelled 0x%03X = 0x%08X", off&0xFFF, v)
}

// setFinish is called by the command processor when it reaches the draw-done marker in the
// FIFO: the pipe has drained to that point. It raises the finish interrupt if the game
// enabled it — a game that polls instead simply leaves it disabled and reads the status.
func (d *pe) setFinish(m *Machine) {
	if d.FinishEnable {
		m.raiseInt(IntPEFinish)
	}
}

// setToken records a token the pipe passed (readable at 0x0E) and, for the interrupting
// form, raises the token interrupt if the game enabled it.
func (d *pe) setToken(m *Machine, tok uint16, raise bool) {
	d.Token = tok
	if raise && d.TokenEnable {
		m.raiseInt(IntPEToken)
	}
}

// --- The write-gather pipe ----------------------------------------------------------
//
// The Gekko has a special store path: writes to one physical address (0x0C008000) are not
// stored, they are gathered into 32-byte bursts and pushed into the graphics FIFO. It is
// how a GameCube program feeds the graphics pipe at speed, without the CPU addressing the
// FIFO word by word.
//
// In Phase 2 the pipe counts its bytes and hands them to OnFIFO if anyone is listening;
// Phase 3's command processor is what consumes them. Counting them is enough to prove the
// path works and to see the game beginning to draw.

type wgPipe struct {
	Bytes uint64 // total bytes pushed, across the run
	Buf   []byte // the current burst, handed to OnFIFO in 32-byte lines
}

func (w *wgPipe) push(m *Machine, b []byte) {
	w.Bytes += uint64(len(b))
	// The command processor consumes the stream: this is what feeds the graphics pipe.
	m.gpu.feed(m, b)
	// The capture tool, if one is listening, gets the same bytes in 32-byte lines.
	if m.OnFIFO == nil {
		return
	}
	w.Buf = append(w.Buf, b...)
	for len(w.Buf) >= 32 {
		m.OnFIFO(w.Buf[:32])
		w.Buf = w.Buf[32:]
	}
}

func (w *wgPipe) write8(m *Machine, v uint8)   { w.push(m, []byte{v}) }
func (w *wgPipe) write16(m *Machine, v uint16) { w.push(m, []byte{uint8(v >> 8), uint8(v)}) }
func (w *wgPipe) write32(m *Machine, v uint32) {
	w.push(m, []byte{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)})
}
