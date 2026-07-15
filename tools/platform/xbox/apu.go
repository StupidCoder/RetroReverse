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

	// The progressing DSP counter the bring-up polls (see the file comment).
	apuHandshake    = 0x20010
	apuCounterShift = 10 // counter = tick >> 10
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

// ac97Read / ac97Write are the AC'97 codec interface: a pure latch. The bring-up's
// semaphore poll (TEST BYTE [+0x134],1 until clear) is satisfied by the 0 default.
func (m *Machine) ac97Read(off uint32) byte {
	dw, written := m.ac97.reg[off>>2]
	return m.latchRead(&m.ac97, off, dw, written)
}
func (m *Machine) ac97Write(off uint32, v byte) { m.latchWrite(&m.ac97, off, v) }
