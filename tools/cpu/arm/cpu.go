package arm

// An ARMv5TE / ARMv4T integer execution core — the executable counterpart of the
// decoder, mirroring mos6502.CPU / m68k.CPU / z80.CPU / sm83.CPU. It implements the
// two instruction sets of the DS's ARM9 (ARM946E-S) and ARM7 (ARM7TDMI): the 32-bit
// ARM set (exec_arm.go) and the 16-bit Thumb set (exec_thumb.go), switching between
// them through the CPSR T bit and the BX/BLX interworking branches.
//
// Memory goes through the Bus, so the caller supplies the DS machine model (main
// RAM, the ITCM/DTCM, shared WRAM, the memory-mapped I/O, the cartridge). This is a
// programmer's-model interpreter — condition codes, the banked register file and
// mode switching are modelled; caches, the MPU, cycle-accurate timing and the 2D/3D
// video hardware are not (that is the "full machine" the caller layers on top, and
// is the hard part the DS makes genuinely challenging).
//
// As in the sibling packages, instructions are implemented as needed and anything
// unmodelled calls Halt with the offending encoding and PC, so gaps are explicit
// rather than silently wrong. Coprocessor accesses (CP15 on the ARM9, and the BIOS
// SWIs) are routed to optional caller hooks.

import "fmt"

// Bus is the flat 32-bit address space the CPU drives (byte-addressed, little-endian
// words composed by the core). The machine model decodes the DS memory map and I/O.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
}

// Processor modes (CPSR bits 4:0).
const (
	ModeUSR = 0x10
	ModeFIQ = 0x11
	ModeIRQ = 0x12
	ModeSVC = 0x13
	ModeABT = 0x17
	ModeUND = 0x1B
	ModeSYS = 0x1F
)

// CPU is the ARM programmer's-model core. The active register file is R; the banked
// copies for the other modes live in the bank* fields and are swapped in on a mode
// change.
type CPU struct {
	R          [16]uint32 // active registers; R[13]=SP, R[14]=LR, R[15]=PC
	N, Z, C, V bool       // CPSR condition flags
	Q          bool       // CPSR sticky saturation flag (set by QADD/SMLAxy overflow)
	Thumb      bool       // CPSR T bit — Thumb state when set
	IRQDisable bool       // CPSR I bit
	FIQDisable bool       // CPSR F bit
	Mode       uint32     // CPSR mode field (ModeUSR…ModeSYS)

	// Banked registers, indexed by modeIndex(). USR/SYS share bank 0.
	bankR13  [6]uint32
	bankR14  [6]uint32
	bankSPSR [6]uint32
	fiqR8_12 [5]uint32 // FIQ has its own r8-r12
	usrR8_12 [5]uint32 // the USR/SYS r8-r12 saved while in FIQ

	// Optional caller hooks. SWI handles a software interrupt (return true if
	// serviced, so the core does not vector to 0x08). Coproc handles MCR/MRC (e.g.
	// CP15 on the ARM9); when nil these are ignored.
	SWI    func(c *CPU, comment uint32) bool
	Coproc func(c *CPU, load bool, cp, op1, crn, crm, op2 uint32, rd *uint32)

	bus        Bus
	Halted     bool
	HaltReason string
	Instrs     uint64

	cur      uint32 // address of the instruction currently executing
	branched bool   // an instruction wrote R[15] (skip the normal PC advance)
}

// NewCPU makes a core over bus in a reset ARM/System-mode state.
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset puts the core in the ARM state at the reset vector in Supervisor mode with
// interrupts disabled, as an ARM core comes out of reset.
func (c *CPU) Reset() {
	c.R = [16]uint32{}
	c.N, c.Z, c.C, c.V, c.Q = false, false, false, false, false
	c.Thumb = false
	c.IRQDisable, c.FIQDisable = true, true
	c.Mode = ModeSVC
	c.Halted, c.HaltReason = false, ""
}

// Halt stops the core (used for an unimplemented instruction or a fatal fault),
// recording why.
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// --- CPSR packing ----------------------------------------------------------

// CPSR assembles the condition/control bits into the 32-bit program status word.
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
	if c.Q {
		v |= 1 << 27
	}
	if c.IRQDisable {
		v |= 1 << 7
	}
	if c.FIQDisable {
		v |= 1 << 6
	}
	if c.Thumb {
		v |= 1 << 5
	}
	return v | (c.Mode & 0x1F)
}

// SetCPSR unpacks a program status word, switching mode/bank if the mode changed.
func (c *CPU) SetCPSR(v uint32) {
	c.N = v&(1<<31) != 0
	c.Z = v&(1<<30) != 0
	c.C = v&(1<<29) != 0
	c.V = v&(1<<28) != 0
	c.Q = v&(1<<27) != 0
	c.IRQDisable = v&(1<<7) != 0
	c.FIQDisable = v&(1<<6) != 0
	c.Thumb = v&(1<<5) != 0
	c.switchMode(v & 0x1F)
}

// modeIndex maps a mode to its bank slot (USR and SYS share slot 0).
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

// switchMode saves the current mode's banked registers and loads the target mode's.
func (c *CPU) switchMode(mode uint32) {
	mode &= 0x1F
	if mode == c.Mode {
		return
	}
	from, to := modeIndex(c.Mode), modeIndex(mode)
	// Save r13/r14 of the current mode.
	c.bankR13[from] = c.R[13]
	c.bankR14[from] = c.R[14]
	// FIQ swaps r8-r12 as well.
	if c.Mode == ModeFIQ {
		copy(c.fiqR8_12[:], c.R[8:13])
	} else {
		copy(c.usrR8_12[:], c.R[8:13])
	}
	// Load target mode.
	c.R[13] = c.bankR13[to]
	c.R[14] = c.bankR14[to]
	if mode == ModeFIQ {
		copy(c.R[8:13], c.fiqR8_12[:])
	} else {
		copy(c.R[8:13], c.usrR8_12[:])
	}
	c.Mode = mode
}

// SPSR returns the saved program status of the current exception mode.
func (c *CPU) SPSR() uint32 { return c.bankSPSR[modeIndex(c.Mode)] }

// SetSPSR sets the saved program status of the current exception mode.
func (c *CPU) SetSPSR(v uint32) { c.bankSPSR[modeIndex(c.Mode)] = v }

// --- memory helpers (little-endian) ----------------------------------------

func (c *CPU) read8(a uint32) byte     { return c.bus.Read(a) }
func (c *CPU) write8(a uint32, v byte) { c.bus.Write(a, v) }

func (c *CPU) read16(a uint32) uint32 {
	return uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8
}
func (c *CPU) write16(a, v uint32) {
	c.bus.Write(a, byte(v))
	c.bus.Write(a+1, byte(v>>8))
}
func (c *CPU) read32(a uint32) uint32 {
	// Unaligned word reads on ARM rotate; game code is almost always aligned, but
	// honour the rotate so an oracle matches hardware.
	v := uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8 | uint32(c.bus.Read(a+2))<<16 | uint32(c.bus.Read(a+3))<<24
	if r := (a & 3) * 8; r != 0 {
		v = ror32(v, r)
	}
	return v
}
func (c *CPU) read32aligned(a uint32) uint32 {
	a &^= 3
	return uint32(c.bus.Read(a)) | uint32(c.bus.Read(a+1))<<8 | uint32(c.bus.Read(a+2))<<16 | uint32(c.bus.Read(a+3))<<24
}
func (c *CPU) write32(a, v uint32) {
	a &^= 3
	c.bus.Write(a, byte(v))
	c.bus.Write(a+1, byte(v>>8))
	c.bus.Write(a+2, byte(v>>16))
	c.bus.Write(a+3, byte(v>>24))
}

// --- register access with R15 pipeline behaviour ---------------------------

// reg reads register i. Reading R15 yields the instruction address + 8 (ARM) or + 4
// (Thumb), the value the pipeline exposes.
func (c *CPU) reg(i uint32) uint32 {
	if i == 15 {
		if c.Thumb {
			return c.cur + 4
		}
		return c.cur + 8
	}
	return c.R[i]
}

// setReg writes register i, flagging a branch when it targets R15.
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

// add computes a+b+cin, sets NZCV, and returns the result.
func (c *CPU) add(a, b, cin uint32) uint32 {
	r := uint64(a) + uint64(b) + uint64(cin)
	res := uint32(r)
	c.setNZ(res)
	c.C = r > 0xFFFFFFFF
	c.V = (a^res)&(b^res)&0x80000000 != 0
	return res
}

// sub computes a-b (with borrow via cin=1 for plain subtract), sets NZCV.
func (c *CPU) sub(a, b, cin uint32) uint32 { return c.add(a, ^b, cin) }

// cond evaluates the 4-bit condition field against the flags.
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
	default: // condAL, condNV
		return true
	}
}

// --- barrel shifter --------------------------------------------------------

// shift applies barrel-shifter type typ by amount amt to val, returning the shifted
// value and the carry-out. regForm is true for a register-specified shift amount
// (which has no special #0 encodings). cin is the current carry (used by LSL #0 and
// RRX).
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
			amt = 32 // immediate LSR #0 means #32
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
			amt = 32 // immediate ASR #0 means #32
		}
		sv := int32(val)
		switch {
		case amt == 0:
			return val, cin
		case amt < 32:
			return uint32(sv >> amt), uint32(val>>(amt-1)) & 1
		default: // >= 32: fill with the sign bit
			if val&(1<<31) != 0 {
				return 0xFFFFFFFF, 1
			}
			return 0, 0
		}
	default: // ROR
		if amt == 0 && !regForm { // RRX: rotate right through carry by one
			return cin<<31 | val>>1, val & 1
		}
		if amt == 0 {
			return val, cin
		}
		amt &= 31
		if amt == 0 { // multiple of 32
			return val, (val >> 31) & 1
		}
		return ror32(val, amt), (val >> (amt - 1)) & 1
	}
}
