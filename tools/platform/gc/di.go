package gc

import (
	"fmt"
	"os"
)

var diTrace = os.Getenv("RR_GC_DITRACE") != ""

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

	// The transfer clock. A real drive takes time, and the game's own code depends on it:
	// Luigi's Mansion mounts its master archive (game_usa.szp) on a loader thread whose
	// completion callback registers a singleton, and the per-scene heap teardown frees
	// groups 1..26 at every scene transition — the design only works because the 2.6 MB
	// read is still in flight when the first transition happens, so the mount allocates
	// into the NEXT scene's heap. An instant drive completes the mount a scene early and
	// the game's own teardown destroys its own singleton. So a read command holds the
	// drive busy for a realistic count of instructions and completes from tickDI.
	BusyInstr int64  // instructions until the in-flight command completes (0 = idle)
	PendOff   int64  // the in-flight read's disc byte offset
	PendLen   uint32 // its length
	PendMAR   uint32 // its DMA target
	LastEnd   int64  // one past the previous read's end, for the sequential-read discount
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
		if diTrace && (off&0xFF) == 0x10 {
			fmt.Fprintf(os.Stderr, "DI Cmd2(offset word)=0x%08X -> byte 0x%X (pc 0x%08X)\n%s",
				v, int64(v)<<2, m.CPU.PC, m.BacktraceString())
		}
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
		// The command block is three words: the opcode in CMDBUF0, the disc offset in CMDBUF1
		// (Cmd[1]), and the length in CMDBUF2 (Cmd[2]). The offset is in 4-byte units — the
		// drive addresses the disc in words, not bytes. Reading it from Cmd[2] instead — the
		// length register — is the bug that made every game-issued read fetch from
		// length<<2 and hand the decompressor pure garbage; the apploader never caught it
		// because its reads are serviced directly, not through this command block.
		discOff := int64(d.Cmd[1]) << 2
		length := d.Length
		if m.OnDVDRead != nil {
			m.OnDVDRead(discOff, length, d.MAR)
		}
		if d.BusyInstr != 0 {
			// The SDK waits for the transfer-complete interrupt before issuing the next
			// command, so this should be unreachable; if it fires, the model is wrong.
			m.logf("DI read issued while a transfer is in flight (offset 0x%X)", discOff)
		}
		d.PendOff, d.PendLen, d.PendMAR = discOff, length, d.MAR
		d.BusyInstr = diCmdInstr + int64(length)*diInstrPerByte
		if discOff != d.LastEnd {
			d.BusyInstr += diSeekInstr
		}
		if diInstant {
			d.BusyInstr = 1
		}
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

// The drive's pace, on the instruction clock (fieldInstructions per 1/60 s field, so one
// emulated second is 60*fieldInstructions instructions). The real drive is CAV — roughly
// 2.0 MB/s at the inner edge to 3.1 MB/s at the outer — modelled here as a flat 2.5 MB/s;
// each command costs about a millisecond of handling, and a non-sequential offset pays a
// seek. RR_GC_DIINSTANT restores the old zero-time drive — a debug fiction, not hardware
// (it is what let the archive mount finish a scene early; see the di struct comment).
const (
	diInstrPerSec  = 60 * fieldInstructions
	diInstrPerByte = diInstrPerSec / 2_500_000 // ~2.5 MB/s
	diCmdInstr     = diInstrPerSec / 1000      // ~1 ms per command
	diSeekInstr    = 30 * diInstrPerSec / 1000 // ~30 ms per seek
)

var diInstant = os.Getenv("RR_GC_DIINSTANT") != ""

// tickDI advances the in-flight transfer one instruction; at zero the data lands and the
// transfer-complete interrupt fires.
func (m *Machine) tickDI() {
	d := &m.di
	if d.BusyInstr == 0 {
		return
	}
	d.BusyInstr--
	if d.BusyInstr != 0 {
		return
	}
	d.LastEnd = d.PendOff + int64(d.PendLen)
	if m.disc != nil {
		data, err := m.disc.Read(d.PendOff, int(d.PendLen))
		if err != nil {
			m.logf("DI read past the disc: offset 0x%X length %d: %v", d.PendOff, d.PendLen, err)
			d.raiseError(m)
			return
		}
		m.dmaToRAM(d.PendMAR, data)
	}
	d.complete(m)
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
