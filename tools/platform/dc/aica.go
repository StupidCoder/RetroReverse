package dc

// aica.go is the sound side: the AICA's ARM7DI running the driver the game
// uploads into the 2 MB sound RAM, paced off the SH-4's instruction count.
// The register block (SH-4 window 00700000, ARM window 00800000 — the same
// registers) follows the machine's discipline: what a run has demanded is
// modeled, everything else raw-stores and lands in the census. Sound
// synthesis — the 64 channels — is a later milestone; what exists here is
// what the boot needs: the ARM runs, the two CPUs converse through shared
// RAM and the interrupt registers.
//
// The ARM7DI's instruction set is a subset of the ARMv5TE core in
// tools/cpu/arm (no Thumb, no enhanced DSP), so the shared core executes the
// driver as-is; anything genuinely v5-only appearing in a trace would be a
// fact about the dump, not the machine.
//
// Pacing: the AICA ARM clocks at the audio master rate, 44100·512 ≈ 22.58
// MHz — close to a ninth of the SH-4's 200 MHz — so the run loop steps it
// once per nine SH-4 instructions. The ratio is a modeling choice, not a
// measured fact; nothing in a handshake depends on its exact value.

const armDivisor = 9

// One audio sample at 44.1 kHz, in SH-4 instructions at the model's
// one-instruction-per-clock pacing: 200 MHz / 44100.
const sampleInstrs = 4535

// AICA interrupt-pending bits (SCIPD/MCIPD): the three timers.
const (
	aicaIntTimerA = 1 << 6
	aicaIntTimerB = 1 << 7
	aicaIntTimerC = 1 << 8
)

// aicaTimers is the three 8-bit up-counters, each incrementing every
// 2^prescale samples and flagging its interrupt bit on overflow. Exported
// fields: it savestates.
type aicaTimers struct {
	Frac  uint32 // instructions accumulated toward the next sample
	Sub   [3]uint32
	SCIPD uint32 // pending, ARM side
	MCIPD uint32 // pending, SH-4 side
}

// timer register offsets in the block: TIMA 2890, TIMB 2894, TIMC 2898.
func timerReg(i int) uint32 { return 0x2890 + uint32(i)*4 }

// tickAICA advances the sample clock by one SH-4 instruction and the timers
// with it.
func (m *Machine) tickAICA() {
	t := &m.Timers
	t.Frac++
	if t.Frac < sampleInstrs {
		return
	}
	t.Frac = 0
	for i := 0; i < 3; i++ {
		reg := m.AICARegs[timerReg(i)]
		pre := (reg >> 8) & 7
		t.Sub[i]++
		if t.Sub[i] < 1<<pre {
			continue
		}
		t.Sub[i] = 0
		count := reg&0xFF + 1
		if count > 0xFF {
			count = 0
			bit := uint32(aicaIntTimerA) << i
			t.SCIPD |= bit
			t.MCIPD |= bit
			m.aicaMainIRQ()
		}
		m.AICARegs[timerReg(i)] = reg&^0xFF | count
	}
}

// aicaMainIRQ raises the Holly external-interrupt bit for the AICA when the
// SH-4-side mask admits any pending source. ISTEXT bit 1 is the AICA's line;
// it is level-like — cleared when the pending set drains, not by an ISTEXT
// write.
func (m *Machine) aicaMainIRQ() {
	if m.Timers.MCIPD&m.AICARegs[0x28B4] != 0 {
		m.Holly.ISTEXT |= 1 << 1
	} else {
		m.Holly.ISTEXT &^= 1 << 1
	}
	m.updateIRL()
}

// armBus is the AICA ARM's view of the world: sound RAM low, the register
// block at 00800000.
type armBus struct{ m *Machine }

func (b armBus) region(addr uint32) (mem []byte, off uint32, reg bool) {
	a := addr & 0x00FFFFFF
	if a < 0x00800000 {
		return b.m.AICARAM, a & (AICARAMSize - 1), false
	}
	return nil, a - 0x00800000, true
}

func (b armBus) Read(addr uint32) byte {
	if mem, off, reg := b.region(addr); !reg {
		return mem[off]
	} else {
		return byte(b.m.aicaRead(off&^3, 4, false) >> (8 * (off & 3)))
	}
}

func (b armBus) Write(addr uint32, v byte) {
	if mem, off, reg := b.region(addr); !reg {
		mem[off] = v
	} else {
		sh := 8 * (off & 3)
		old := b.m.aicaRead(off&^3, 4, false)
		b.m.aicaWrite(off&^3, 4, old&^(0xFF<<sh)|uint32(v)<<sh, false)
	}
}

func (b armBus) Read16(addr uint32) uint16 {
	if mem, off, reg := b.region(addr); !reg {
		return uint16(mem[off]) | uint16(mem[off+1])<<8
	} else {
		return uint16(b.m.aicaRead(off&^3, 2, false) >> (8 * (off & 2)))
	}
}

func (b armBus) Read32(addr uint32) uint32 {
	if mem, off, reg := b.region(addr); !reg {
		return uint32(mem[off]) | uint32(mem[off+1])<<8 | uint32(mem[off+2])<<16 | uint32(mem[off+3])<<24
	} else {
		return b.m.aicaRead(off, 4, false)
	}
}

func (b armBus) Write16(addr uint32, v uint16) {
	if mem, off, reg := b.region(addr); !reg {
		mem[off], mem[off+1] = uint8(v), uint8(v>>8)
	} else {
		b.m.aicaWrite(off, 2, uint32(v), false)
	}
}

func (b armBus) Write32(addr uint32, v uint32) {
	if mem, off, reg := b.region(addr); !reg {
		mem[off], mem[off+1], mem[off+2], mem[off+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
	} else {
		b.m.aicaWrite(off, 4, v, false)
	}
}

// aicaRead serves the register block; off is the offset within it. sh4 says
// which side is asking — the census names the asker.
func (m *Machine) aicaRead(off uint32, size int, sh4 bool) uint32 {
	switch off {
	case 0x2C00: // ARMRST readback
		if m.ARMRunning {
			return 0
		}
		return 1
	case 0x28A0: // SCIPD: the ARM's pending interrupts
		return m.Timers.SCIPD
	case 0x28B8: // MCIPD: the SH-4's
		return m.Timers.MCIPD
	}
	if v, ok := m.AICARegs[off&^3]; ok {
		return v
	}
	m.logf("AICA read %04X by %s (never written)", off, side(sh4))
	return 0
}

func (m *Machine) aicaWrite(off uint32, size int, v uint32, sh4 bool) {
	switch off {
	case 0x2C00: // ARMRST: bit 0 holds the ARM in reset; releasing it starts
		// the driver at the sound RAM's reset vector.
		if !sh4 {
			m.logf("AICA ARM wrote its own reset register")
			return
		}
		if v&1 != 0 {
			m.ARMRunning = false
			return
		}
		if !m.ARMRunning {
			m.ARM.Reset()
			m.ARM.R[15] = 0
			m.ARMRunning = true
		}
		return
	case 0x28A4: // SCIRE: writing 1s retires pending ARM interrupts
		m.Timers.SCIPD &^= v
		return
	case 0x28BC: // MCIRE: the SH-4-side acknowledge
		m.Timers.MCIPD &^= v
		m.aicaMainIRQ()
		return
	case 0x28B4: // MCIEB: re-evaluate the line when the mask moves
		m.AICARegs[off] = v
		m.aicaMainIRQ()
		return
	}
	m.AICARegs[off&^3] = v
}

func side(sh4 bool) string {
	if sh4 {
		return "sh4"
	}
	return "arm"
}

// stepARM advances the sound CPU by one instruction, if it is out of reset.
// An ARM halt is a census entry and a parked core, not a dead machine: the
// SH-4 side keeps its own story running.
func (m *Machine) stepARM() {
	if !m.ARMRunning || m.ARM.Halted {
		return
	}
	m.ARM.Step()
	if m.ARM.Halted {
		m.logf("AICA ARM halted: %s", m.ARM.HaltReason)
	}
}

// RTC: the two 16-bit halves of seconds-since-1950 at 00710000/00710004.
// The value is a fixed date rather than the host clock, because two runs of
// the oracle must be byte-identical; 0x5EE40000 seconds lands mid-2000,
// contemporary with the disc.
const rtcSeconds = 0x5EE40000

func (m *Machine) rtcRead(addr uint32) uint32 {
	switch addr {
	case 0x00710000:
		return rtcSeconds >> 16
	case 0x00710004:
		return rtcSeconds & 0xFFFF
	}
	m.logf("RTC read %08X", addr)
	return 0
}
