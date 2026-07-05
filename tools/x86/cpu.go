// This file adds an instruction-level 16-bit real-mode x86 execution core to
// the x86 package — the executable counterpart to the disassembler, mirroring
// the shape of the mos6502/m68k CPU cores. Memory is reached through the Bus
// interface as a flat little-endian byte space; the CPU folds seg:off to a
// 20-bit linear address (8086 wrap) itself, so a machine model just supplies
// RAM and can service software interrupts through an IntHook (the DOS layer for
// UW.EXE does exactly that).
//
// Scope: a "minimal but real" core — MOV/MOVZX/MOVSX, LEA, the PUSH/POP family
// (incl. PUSHA/POPA/PUSHF/POPF), the eight ALU ops in every addressing form
// plus INC/DEC/NEG/NOT/TEST, MUL/IMUL/DIV/IDIV, the shift/rotate family, the
// REP string ops (MOVS/STOS/LODS/SCAS/CMPS), the full near/short/far and
// indirect control transfers (JMP/CALL/RET/RETF/Jcc/LOOP/JCXZ), the flag ops,
// XCHG, CBW/CWD, XLAT, and INT/IRET — all with correct real-mode flags. Anything
// outside that set (notably the x87 FPU) halts the CPU with the offending
// opcode, so gaps are obvious and easy to fill on demand.
package x86

import "fmt"

// Bus is the memory interface seen by the CPU: a flat, byte-addressed, little-
// endian 1 MiB real-mode space. The CPU composes words/dwords from it.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
}

// register indices (match the x86 encoding of the reg/rm fields).
const (
	AX = iota
	CX
	DX
	BX
	SP
	BP
	SI
	DI
)

// 8-bit register indices (low then high bytes), for Reg8/SetReg8.
const (
	AL = iota
	CL
	DL
	BL
	AH
	CH
	DH
	BH
)

// segment-register indices (match the encoding used by MOV sreg and overrides).
const (
	ES = iota
	CS
	SS
	DS
	FS
	GS
)

// CPU is a real-mode x86 execution core. Registers are exported so a machine
// model and tests can seed and inspect them.
type CPU struct {
	Regs [8]uint32 // EAX ECX EDX EBX ESP EBP ESI EDI (16-bit access = low word)
	Seg  [8]uint16 // ES CS SS DS FS GS (indices 6,7 unused, padded so a bad sreg field can't panic)
	IP   uint32    // instruction pointer (16-bit in real mode)

	// arithmetic + control flags
	CF, PF, AF, ZF, SF, TF, IF, DF, OF bool

	Halted     bool
	HaltReason string
	Steps      uint64
	Ext386     uint64 // count of 0x0F two-byte / 0x66 / 0x67 (386) instructions executed

	// IntHook services a software INT n. Returning true means the interrupt was
	// fully emulated (registers/flags set, IP already past the INT); returning
	// false falls back to the real-mode IVT dispatch. Nil = always IVT.
	IntHook func(c *CPU, n byte) bool

	// OnStep, if set, is called before each instruction is fetched — a place for
	// a machine model to log the PC, watch memory, or detect spins (it may call
	// Halt to stop the run).
	OnStep func(c *CPU)

	// PortIn/PortOut model I/O-port access (IN/OUT). Unset means no device is
	// present: IN reads all-ones, OUT is discarded. size is in bytes (1/2).
	PortIn  func(port uint16, size int) uint32
	PortOut func(port uint16, size int, v uint32)

	bus Bus

	// ssShadow inhibits maskable interrupts for the one instruction following a
	// load of SS (MOV SS,x / POP SS), matching the hardware so a machine model's
	// injected IRQ can't land between a MOV SS / MOV SP stack switch.
	ssShadow bool

	// transient per-instruction decode state (segment override / sizes)
	dSeg      int // segment-override index, or -1
	dOpsize   int // 16 or 32
	dAddrsize int // 16 or 32
}

// NewCPU returns a CPU bound to bus.
func NewCPU(bus Bus) *CPU { return &CPU{bus: bus} }

// Halt stops execution and records why.
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

// --- linear addressing (20-bit real-mode wrap) ---

func (c *CPU) lin(seg uint16, off uint32) uint32 {
	return ((uint32(seg) << 4) + (off & 0xFFFF)) & 0xFFFFF
}

func (c *CPU) rd8(a uint32) uint32 { return uint32(c.bus.Read(a & 0xFFFFF)) }
func (c *CPU) rd16(a uint32) uint32 {
	return c.rd8(a) | c.rd8(a+1)<<8
}
func (c *CPU) rd32(a uint32) uint32 {
	return c.rd16(a) | c.rd16(a+2)<<16
}
func (c *CPU) wr8(a, v uint32) { c.bus.Write(a&0xFFFFF, byte(v)) }
func (c *CPU) wr16(a, v uint32) {
	c.wr8(a, v)
	c.wr8(a+1, v>>8)
}
func (c *CPU) wr32(a, v uint32) {
	c.wr16(a, v)
	c.wr16(a+2, v>>16)
}

// memRead/memWrite access a seg:off operand at a byte width (1/2/4). Each byte's
// address is computed from the wrapped 16-bit offset (real-mode segment wrap), so
// a word/dword straddling offset 0xFFFF reads/writes its high byte(s) from the
// start of the same segment — not the next linear paragraph (8086 semantics,
// verified against the SingleStepTests suite).
func (c *CPU) memRead(seg uint16, off uint32, b int) uint32 {
	v := c.rd8(c.lin(seg, off))
	if b >= 2 {
		v |= c.rd8(c.lin(seg, off+1)) << 8
	}
	if b == 4 {
		v |= c.rd8(c.lin(seg, off+2))<<16 | c.rd8(c.lin(seg, off+3))<<24
	}
	return v
}
func (c *CPU) memWrite(seg uint16, off uint32, b int, v uint32) {
	c.wr8(c.lin(seg, off), v)
	if b >= 2 {
		c.wr8(c.lin(seg, off+1), v>>8)
	}
	if b == 4 {
		c.wr8(c.lin(seg, off+2), v>>16)
		c.wr8(c.lin(seg, off+3), v>>24)
	}
}

// --- instruction fetch (CS:IP) ---

func (c *CPU) fetch8() uint32 {
	v := c.rd8(c.lin(c.Seg[CS], c.IP))
	c.IP = (c.IP + 1) & 0xFFFF
	return v
}
func (c *CPU) fetch16() uint32 {
	lo := c.fetch8()
	hi := c.fetch8()
	return lo | hi<<8
}
func (c *CPU) fetch32() uint32 {
	lo := c.fetch16()
	hi := c.fetch16()
	return lo | hi<<16
}

// fetchImm reads an operand-size immediate (2 or 4 bytes).
func (c *CPU) fetchImm() uint32 {
	if c.dOpsize == 32 {
		return c.fetch32()
	}
	return c.fetch16()
}

// --- register access ---

func (c *CPU) g8(i byte) uint32 {
	if i < 4 {
		return c.Regs[i] & 0xFF
	}
	return (c.Regs[i-4] >> 8) & 0xFF
}
func (c *CPU) s8(i byte, v uint32) {
	if i < 4 {
		c.Regs[i] = (c.Regs[i] &^ 0xFF) | (v & 0xFF)
	} else {
		c.Regs[i-4] = (c.Regs[i-4] &^ 0xFF00) | ((v & 0xFF) << 8)
	}
}
func (c *CPU) g16(i byte) uint32    { return c.Regs[i] & 0xFFFF }
func (c *CPU) s16(i byte, v uint32) { c.Regs[i] = (c.Regs[i] &^ 0xFFFF) | (v & 0xFFFF) }

// getReg/setReg access general register i at byte width b (1/2/4).
func (c *CPU) getReg(i byte, b int) uint32 {
	switch b {
	case 1:
		return c.g8(i)
	case 2:
		return c.g16(i)
	default:
		return c.Regs[i]
	}
}
func (c *CPU) setReg(i byte, b int, v uint32) {
	switch b {
	case 1:
		c.s8(i, v)
	case 2:
		c.s16(i, v)
	default:
		c.Regs[i] = v
	}
}

// convenience 16-bit accessors used pervasively
func (c *CPU) gw(i int) uint32    { return c.Regs[i] & 0xFFFF }
func (c *CPU) sw(i int, v uint32) { c.Regs[i] = (c.Regs[i] &^ 0xFFFF) | (v & 0xFFFF) }

// Exported register views for a machine model (e.g. a DOS INT-21h layer).

// Reg8 reads an 8-bit register by index (AL,CL,DL,BL,AH,CH,DH,BH).
func (c *CPU) Reg8(i int) byte { return byte(c.g8(byte(i))) }

// SetReg8 writes an 8-bit register by index.
func (c *CPU) SetReg8(i int, v byte) { c.s8(byte(i), uint32(v)) }

// Reg16 reads a 16-bit register by index (AX,CX,DX,BX,SP,BP,SI,DI).
func (c *CPU) Reg16(i int) uint16 { return uint16(c.gw(i)) }

// SetReg16 writes a 16-bit register by index.
func (c *CPU) SetReg16(i int, v uint16) { c.sw(i, uint32(v)) }

// --- stack ---

func (c *CPU) push16(v uint32) {
	c.sw(SP, c.gw(SP)-2)
	c.memWrite(c.Seg[SS], c.gw(SP), 2, v)
}
func (c *CPU) pop16() uint32 {
	v := c.memRead(c.Seg[SS], c.gw(SP), 2)
	c.sw(SP, c.gw(SP)+2)
	return v
}
func (c *CPU) push(b int, v uint32) {
	if b == 4 {
		c.sw(SP, c.gw(SP)-4)
		c.memWrite(c.Seg[SS], c.gw(SP), 4, v)
		return
	}
	c.push16(v)
}
func (c *CPU) pop(b int) uint32 {
	if b == 4 {
		v := c.memRead(c.Seg[SS], c.gw(SP), 4)
		c.sw(SP, c.gw(SP)+4)
		return v
	}
	return c.pop16()
}

// --- flags ---

// EFlags assembles the flag bits (the low 16 real-mode bits).
func (c *CPU) EFlags() uint16 {
	f := uint16(0x0002) // bit 1 reads as 1
	set := func(cond bool, bit uint) {
		if cond {
			f |= 1 << bit
		}
	}
	set(c.CF, 0)
	set(c.PF, 2)
	set(c.AF, 4)
	set(c.ZF, 6)
	set(c.SF, 7)
	set(c.TF, 8)
	set(c.IF, 9)
	set(c.DF, 10)
	set(c.OF, 11)
	return f
}

// SetEFlags loads the flag bits.
func (c *CPU) SetEFlags(f uint16) {
	c.CF = f&(1<<0) != 0
	c.PF = f&(1<<2) != 0
	c.AF = f&(1<<4) != 0
	c.ZF = f&(1<<6) != 0
	c.SF = f&(1<<7) != 0
	c.TF = f&(1<<8) != 0
	c.IF = f&(1<<9) != 0
	c.DF = f&(1<<10) != 0
	c.OF = f&(1<<11) != 0
}

// --- width helpers ---

func widthMask(b int) uint32 {
	switch b {
	case 1:
		return 0xFF
	case 2:
		return 0xFFFF
	default:
		return 0xFFFFFFFF
	}
}
func signMask(b int) uint32 {
	switch b {
	case 1:
		return 0x80
	case 2:
		return 0x8000
	default:
		return 0x80000000
	}
}
func signExtByte(v uint32) uint32 { return uint32(int32(int8(byte(v)))) }
func signExtWord(v uint32) uint32 { return uint32(int32(int16(uint16(v)))) }

// setSZP sets SF, ZF and PF from a result at width b.
func (c *CPU) setSZP(res uint32, b int) {
	m := widthMask(b)
	c.ZF = res&m == 0
	c.SF = res&signMask(b) != 0
	c.PF = parity(byte(res))
}

func parity(v byte) bool {
	v ^= v >> 4
	v ^= v >> 2
	v ^= v >> 1
	return v&1 == 0
}

// flagsAdd sets flags for a+b+cin at width w and returns the truncated result.
func (c *CPU) flagsAdd(a, b uint32, cin uint32, w int) uint32 {
	m := widthMask(w)
	full := uint64(a&m) + uint64(b&m) + uint64(cin)
	res := uint32(full) & m
	c.CF = full&(uint64(m)+1) != 0
	c.AF = (a^b^res)&0x10 != 0
	c.OF = (^(a^b)&(a^res))&signMask(w) != 0
	c.setSZP(res, w)
	return res
}

// flagsSub sets flags for a-b-cin at width w and returns the truncated result.
func (c *CPU) flagsSub(a, b uint32, cin uint32, w int) uint32 {
	m := widthMask(w)
	res := (a - b - cin) & m
	c.CF = uint64(b&m)+uint64(cin) > uint64(a&m)
	c.AF = (a^b^res)&0x10 != 0
	c.OF = ((a^b)&(a^res))&signMask(w) != 0
	c.setSZP(res, w)
	return res
}

// flagsLogic sets flags for a bitwise result: CF=OF=0, AF cleared, SZP from res.
func (c *CPU) flagsLogic(res uint32, w int) uint32 {
	res &= widthMask(w)
	c.CF, c.OF, c.AF = false, false, false
	c.setSZP(res, w)
	return res
}

// cond evaluates a condition code (0..15: O,NO,B,NB,Z,NZ,BE,A,S,NS,P,NP,L,GE,LE,G).
func (c *CPU) cond(cc byte) bool {
	switch cc {
	case 0x0:
		return c.OF
	case 0x1:
		return !c.OF
	case 0x2:
		return c.CF
	case 0x3:
		return !c.CF
	case 0x4:
		return c.ZF
	case 0x5:
		return !c.ZF
	case 0x6:
		return c.CF || c.ZF
	case 0x7:
		return !c.CF && !c.ZF
	case 0x8:
		return c.SF
	case 0x9:
		return !c.SF
	case 0xA:
		return c.PF
	case 0xB:
		return !c.PF
	case 0xC:
		return c.SF != c.OF
	case 0xD:
		return c.SF == c.OF
	case 0xE:
		return c.ZF || (c.SF != c.OF)
	default: // 0xF
		return !c.ZF && (c.SF == c.OF)
	}
}
