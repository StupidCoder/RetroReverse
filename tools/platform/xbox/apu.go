package xbox

// apu.go models the MCPX southbridge's two audio MMIO apertures — devices entirely
// separate from the NV2A (0xFD000000):
//
//	0xFE800000  the APU (audio processing unit; its GP DSP block sits at +0x20000)
//	0xFEC00000  the AC'97 audio codec interface
//
// The title's DirectSound library brings both up early in its init: it programs the
// APU across the whole 512 KB window (DSP command/handshake registers around +0x20000,
// mixer banks at +0x2050..0x2074, per-block words at +0x3FFFC/+0x5FFFC), then the
// AC'97 codec (a busy-wait on the register-access semaphore [+0x134] bit 0 being
// clear, and read-modify-writes of the global control word [+0x12C]).
//
// We render frames, not audio: nothing downstream ever consumes what the sound
// library programs here, so both apertures are latches — writes are stored sparsely
// and read back coherently, unwritten registers read 0 (which is also the AC'97
// semaphore's "not busy" truth). The one register with behaviour is the APU counter
// the bring-up *polls*:
//
//	001DE530  MOV EAX, [$FE820010]        001DDA83  MOV EAX, [$FE820010]
//	001DE535  AND EAX, $FFFFFFFC          001DDA88  AND EAX, $FFFFFFFC
//	001DE538  CMP EAX, $00000004          001DDA8B  CMP EAX, $00000020
//	001DE53B  JB  $001DE530               001DDA8E  JB  $001DDA83
//
// First it waits for +0x20010 to reach 4, later 0x20 — a counter the DSP advances as
// it runs, not a ready flag (a constant 4 satisfied the first gate and spun the
// second forever; the fiction surfaced exactly as the log-once note below intends).
// It reads the machine clock, scaled: monotonic and savestate-stable (tick is saved).
// The *rate* is pacing only — nothing consumes audio positions from it yet; if the
// title ever computes buffer cursors from deltas of this register, the scale must be
// re-derived from that code.
//
// An invented value must stay visible, never become quiet evidence: every read of a
// never-written register is logged once (and RR_APU_TRACE=1 traces all traffic).

import (
	"fmt"
	"os"
)

const (
	apuBase = 0xFE800000 // the MCPX APU MMIO aperture
	apuSize = 0x80000    // 512 KB
	apuTop  = apuBase + apuSize

	ac97Base = 0xFEC00000 // the MCPX AC'97 codec-interface aperture
	ac97Size = 0x1000     // 4 KB
	ac97Top  = ac97Base + ac97Size

	usbBase = 0xFED00000 // the MCPX USB OHCI host controller (the game-pad ports)
	usbSize = 0x1000     // one OHCI register file
	usbTop  = usbBase + usbSize

	nicBase = 0xFEF00000 // the MCPX NVNET Ethernet controller (XNET's link probe)
	nicSize = 0x1000
	nicTop  = nicBase + nicSize

	// OHCI operational registers with answers beyond the latch (generic OHCI spec —
	// platform knowledge, like the AC'97 semaphore): HcRevision reads 1.0, and
	// HcCommandStatus reads 0 — every self-clearing command bit (HCR reset, list
	// fills, OCR) has already completed in a synchronous model. XAPI's controller
	// stack (call site 0x2403B8) checks HcControl's InterruptRouting bit and HCFS
	// state, both satisfied by the latch's zero default (USBReset state, no SMM
	// ownership). No devices ever appear on the root hub — no pads yet; input is a
	// later phase, exactly like the GameCube SI.
	usbHcRevision      = 0x00
	usbHcCommandStatus = 0x08

	// The progressing DSP counter the bring-up polls (see the file comment).
	apuHandshake    = 0x20010
	apuCounterShift = 10 // counter = tick >> 10

	// AC'97 bus-master global registers (generic AC'97/ICH spec — platform
	// knowledge, like the OHCI answers above).
	ac97GlobalControl = 0x12C // bit 1 = ACLink cold reset deasserted (active low reset)
	ac97GlobalStatus  = 0x130 // bit 8 = primary codec ready
)

var apuTrace = os.Getenv("RR_APU_TRACE") != ""

// mmioLatch is a sparse latch model of a write-then-read-back device aperture.
type mmioLatch struct {
	name     string
	reg      map[uint32]uint32 // dword index -> value (sparse)
	seenCold map[uint32]bool   // unwritten registers read at least once (log-once)
}

func newMMIOLatch(name string) mmioLatch {
	return mmioLatch{name: name, reg: map[uint32]uint32{}, seenCold: map[uint32]bool{}}
}

// latchRead answers a byte read from a latch aperture (dw is the current dword value,
// already special-cased by the caller if the register has behaviour).
func (m *Machine) latchRead(l *mmioLatch, off, dw uint32, written bool) byte {
	if !written && !l.seenCold[off>>2] && len(l.seenCold) < 64 {
		l.seenCold[off>>2] = true
		m.logf("%s: unwritten register %05X read -> 0 (PC %08X)", l.name, off&^3, m.CPU.LinearPC())
	}
	if apuTrace && off&3 == 0 {
		fmt.Printf("%srd %05X -> %08X  PC=%08X\n", l.name, off, dw, m.CPU.LinearPC())
	}
	return byte(dw >> (8 * (off & 3)))
}

// latchWrite stores a byte write into the sparse latch.
func (m *Machine) latchWrite(l *mmioLatch, off uint32, v byte) {
	idx := off >> 2
	shift := 8 * (off & 3)
	l.reg[idx] = (l.reg[idx] &^ (0xFF << shift)) | uint32(v)<<shift
	if apuTrace && off&3 == 3 {
		fmt.Printf("%swr %05X <- %08X  PC=%08X\n", l.name, off&^3, l.reg[idx], m.CPU.LinearPC())
	}
}

// apuRead answers a byte read from the APU aperture (offset within 0xFE800000).
func (m *Machine) apuRead(off uint32) byte {
	dw, written := m.apu.reg[off>>2]
	if off&^3 == apuHandshake {
		dw, written = uint32(m.tick>>apuCounterShift), true
	}
	return m.latchRead(&m.apu, off, dw, written)
}

// apuWrite latches a byte write to the APU aperture.
func (m *Machine) apuWrite(off uint32, v byte) { m.latchWrite(&m.apu, off, v) }

// ac97Read / ac97Write are the AC'97 codec interface: a latch plus the one register with
// behaviour. The bring-up's semaphore poll (TEST BYTE [+0x134],1 until clear) is satisfied
// by the 0 default. The bus-master Control byte at each box+0x0B (offsets ...0B/1B/.../11B/
// 17B) carries the RR "Reset Registers" bit (0x02), which hardware SELF-CLEARS once the
// per-box reset completes; a pure latch would echo the written 1 forever and spin the
// codec-reset handshake at 0x1DE9EA (write 0x02 to [box+0x0B]; loop while bit 1 stays set).
// Reading a CR byte therefore clears bit 0x02 — the reset is instantaneous in our model —
// while leaving RPBM (bit 0, run) and the interrupt-enable bits untouched.
// The DirectSound device init's driver probe (0x1DE6F6) deasserts ACLink cold
// reset in Global Control and then polls Global Status for the primary codec
// ready bit (1000 tries, 20 ms apart) — DSERR_NODRIVER if it never rises. The
// codec is soldered onto every Xbox: once the link is out of cold reset it
// reports ready (instantaneously here, like the RR reset below).
func (m *Machine) ac97Read(off uint32) byte {
	dw, written := m.ac97.reg[off>>2]
	if off&^3 == ac97GlobalStatus && m.ac97.reg[ac97GlobalControl>>2]&0x02 != 0 {
		dw, written = dw|0x100, true // primary codec ready
	}
	b := m.latchRead(&m.ac97, off, dw, written)
	if off&0x0F == 0x0B {
		b &^= 0x02 // RR (Reset Registers) has already completed
	}
	return b
}
func (m *Machine) ac97Write(off uint32, v byte) { m.latchWrite(&m.ac97, off, v) }

// nicRead / nicWrite are the MCPX NVNET Ethernet controller: a pure latch. The
// title's XNET stack initialises the NIC during network bring-up; with no cable
// and no modelled PHY every unwritten register reads 0 (link down, nothing
// pending), which is the honest no-network truth — the log-once guard surfaces
// any register the stack polls for a rising bit.
func (m *Machine) nicRead(off uint32) byte {
	dw, written := m.nic.reg[off>>2]
	return m.latchRead(&m.nic, off, dw, written)
}
func (m *Machine) nicWrite(off uint32, v byte) { m.latchWrite(&m.nic, off, v) }

// usbRead / usbWrite are the USB OHCI host controller: a latch with the two
// spec-mandated answers above.
func (m *Machine) usbRead(off uint32) byte {
	dw, written := m.usb.reg[off>>2]
	switch off &^ 3 {
	case usbHcRevision:
		dw, written = 0x10, true // OHCI 1.0
	case usbHcCommandStatus:
		dw, written = 0, true // all self-clearing command bits already done
	}
	return m.latchRead(&m.usb, off, dw, written)
}
func (m *Machine) usbWrite(off uint32, v byte) { m.latchWrite(&m.usb, off, v) }
