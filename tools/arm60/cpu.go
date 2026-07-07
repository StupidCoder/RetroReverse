package arm60

import "fmt"

// Bus is the flat 32-bit address space the CPU drives (byte-addressed). The 3DO
// machine model (tools/threedo) decodes the memory map and the Madam/Clio I/O.
// Words are composed big-endian by the core.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
}

// Processor modes (CPSR bits 4:0), 32-bit-mode encodings.
const (
	ModeUSR = 0x10
	ModeFIQ = 0x11
	ModeIRQ = 0x12
	ModeSVC = 0x13
	ModeABT = 0x17
	ModeUND = 0x1B
	ModeSYS = 0x1F
)

// CPU is the ARM60 programmer's-model core.
type CPU struct {
	R          [16]uint32 // R[13]=SP, R[14]=LR, R[15]=PC
	N, Z, C, V bool       // CPSR condition flags
	IRQDisable bool       // CPSR I bit
	FIQDisable bool       // CPSR F bit
	Mode       uint32     // CPSR mode field

	bankR13  [6]uint32
	bankR14  [6]uint32
	bankSPSR [6]uint32
	fiqR8_12 [5]uint32
	usrR8_12 [5]uint32

	// SWI handles a software interrupt (return true if serviced, so the core does
	// not vector to 0x08). Used to high-level-emulate the Portfolio OS SWI gate.
	SWI func(c *CPU, comment uint32) bool

	bus        Bus
	Halted     bool
	HaltReason string
	Instrs     uint64

	cur      uint32 // address of the instruction currently executing
	branched bool   // an instruction wrote R[15]
}

// NewCPU makes a core over bus in a reset state.
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset puts the core at the reset vector in Supervisor mode with interrupts off.
func (c *CPU) Reset() {
	c.R = [16]uint32{}
	c.N, c.Z, c.C, c.V = false, false, false, false
	c.IRQDisable, c.FIQDisable = true, true
	c.Mode = ModeSVC
	c.Halted, c.HaltReason = false, ""
}

// Halt stops the core, recording why.
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// CurPC returns the address of the instruction currently executing.
func (c *CPU) CurPC() uint32 { return c.cur }

// SetPC / SetReg seed state before entering a program.
func (c *CPU) SetPC(v uint32)      { c.R[15] = v }
func (c *CPU) SetReg(i, v uint32)  { c.R[i&0xF] = v }
func (c *CPU) Reg(i uint32) uint32 { return c.R[i&0xF] }

// --- CPSR packing ----------------------------------------------------------

func (c *CPU) CPSR() uint32 {
	var v uint32
	if c.N {
		v |= 1 << 31
	}
	if c.Z {
		v |= 1 << 30
	}
	if c.C {
		v |= 1 << 29
	}
	if c.V {
		v |= 1 << 28
	}
	if c.IRQDisable {
		v |= 1 << 7
	}
	if c.FIQDisable {
		v |= 1 << 6
	}
	return v | (c.Mode & 0x1F)
}

func (c *CPU) SetCPSR(v uint32) {
	c.N = v&(1<<31) != 0
	c.Z = v&(1<<30) != 0
	c.C = v&(1<<29) != 0
	c.V = v&(1<<28) != 0
	c.IRQDisable = v&(1<<7) != 0
	c.FIQDisable = v&(1<<6) != 0
	c.switchMode(v & 0x1F)
}

func modeIndex(mode uint32) int {
	switch mode {
	case ModeFIQ:
		return 1
	case ModeIRQ:
		return 2
	case ModeSVC:
		return 3
	case ModeABT:
		return 4
	case ModeUND:
		return 5
	default: // USR / SYS
		return 0
	}
}

func (c *CPU) switchMode(mode uint32) {
	mode &= 0x1F
	if mode == c.Mode {
		return
	}
	from, to := modeIndex(c.Mode), modeIndex(mode)
	c.bankR13[from] = c.R[13]
	c.bankR14[from] = c.R[14]
	if c.Mode == ModeFIQ {
		copy(c.fiqR8_12[:], c.R[8:13])
	} else {
		copy(c.usrR8_12[:], c.R[8:13])
	}
	c.R[13] = c.bankR13[to]
	c.R[14] = c.bankR14[to]
	if mode == ModeFIQ {
		copy(c.R[8:13], c.fiqR8_12[:])
	} else {
		copy(c.R[8:13], c.usrR8_12[:])
	}
	c.Mode = mode
}

func (c *CPU) SPSR() uint32     { return c.bankSPSR[modeIndex(c.Mode)] }
func (c *CPU) SetSPSR(v uint32) { c.bankSPSR[modeIndex(c.Mode)] = v }

// --- memory helpers (big-endian) -------------------------------------------

func (c *CPU) read8(a uint32) byte     { return c.bus.Read(a) }
func (c *CPU) write8(a uint32, v byte) { c.bus.Write(a, v) }

func (c *CPU) read32aligned(a uint32) uint32 {
	a &^= 3
	return uint32(c.bus.Read(a))<<24 | uint32(c.bus.Read(a+1))<<16 | uint32(c.bus.Read(a+2))<<8 | uint32(c.bus.Read(a+3))
}

// read32 honours the ARM unaligned-word rotate (games are usually aligned, but an
// oracle should match hardware).
func (c *CPU) read32(a uint32) uint32 {
	v := c.read32aligned(a)
	if r := (a & 3) * 8; r != 0 {
		v = ror32(v, r)
	}
	return v
}

func (c *CPU) write32(a, v uint32) {
	a &^= 3
	c.bus.Write(a, byte(v>>24))
	c.bus.Write(a+1, byte(v>>16))
	c.bus.Write(a+2, byte(v>>8))
	c.bus.Write(a+3, byte(v))
}

// --- register access with R15 pipeline behaviour ---------------------------

// reg reads register i; reading R15 yields the instruction address + 8 (the ARM
// pipeline value).
func (c *CPU) reg(i uint32) uint32 {
	if i == 15 {
		return c.cur + 8
	}
	return c.R[i]
}

func (c *CPU) setReg(i, v uint32) {
	c.R[i] = v
	if i == 15 {
		c.branched = true
	}
}

// --- flags -----------------------------------------------------------------

func (c *CPU) setNZ(v uint32) {
	c.Z = v == 0
	c.N = v&(1<<31) != 0
}

func (c *CPU) add(a, b, cin uint32) uint32 {
	r := uint64(a) + uint64(b) + uint64(cin)
	res := uint32(r)
	c.setNZ(res)
	c.C = r > 0xFFFFFFFF
	c.V = (a^res)&(b^res)&0x80000000 != 0
	return res
}

func (c *CPU) sub(a, b, cin uint32) uint32 { return c.add(a, ^b, cin) }

func (c *CPU) boolToU(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func (c *CPU) cond(cc int) bool {
	switch cc {
	case condEQ:
		return c.Z
	case condNE:
		return !c.Z
	case condCS:
		return c.C
	case condCC:
		return !c.C
	case condMI:
		return c.N
	case condPL:
		return !c.N
	case condVS:
		return c.V
	case condVC:
		return !c.V
	case condHI:
		return c.C && !c.Z
	case condLS:
		return !c.C || c.Z
	case condGE:
		return c.N == c.V
	case condLT:
		return c.N != c.V
	case condGT:
		return !c.Z && c.N == c.V
	case condLE:
		return c.Z || c.N != c.V
	case condNV:
		return false // never
	default: // condAL
		return true
	}
}

// --- barrel shifter --------------------------------------------------------

func (c *CPU) shift(typ, amt, val uint32, regForm bool, cin uint32) (uint32, uint32) {
	if regForm {
		amt &= 0xFF
		if amt == 0 {
			return val, cin
		}
	}
	switch typ {
	case 0: // LSL
		switch {
		case amt == 0:
			return val, cin
		case amt < 32:
			return val << amt, (val >> (32 - amt)) & 1
		case amt == 32:
			return 0, val & 1
		default:
			return 0, 0
		}
	case 1: // LSR
		if amt == 0 && !regForm {
			amt = 32
		}
		switch {
		case amt == 0:
			return val, cin
		case amt < 32:
			return val >> amt, (val >> (amt - 1)) & 1
		case amt == 32:
			return 0, (val >> 31) & 1
		default:
			return 0, 0
		}
	case 2: // ASR
		if amt == 0 && !regForm {
			amt = 32
		}
		sv := int32(val)
		switch {
		case amt == 0:
			return val, cin
		case amt < 32:
			return uint32(sv >> amt), uint32(val>>(amt-1)) & 1
		default:
			if val&(1<<31) != 0 {
				return 0xFFFFFFFF, 1
			}
			return 0, 0
		}
	default: // ROR
		if amt == 0 && !regForm { // RRX
			return cin<<31 | val>>1, val & 1
		}
		if amt == 0 {
			return val, cin
		}
		amt &= 31
		if amt == 0 {
			return val, (val >> 31) & 1
		}
		return ror32(val, amt), (val >> (amt - 1)) & 1
	}
}
