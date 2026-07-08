package sm83

// A practically complete Sharp LR35902 (Game Boy CPU) execution core — the
// executable counterpart of the decoder in decode.go, mirroring mos6502.CPU /
// m68k.CPU / z80.CPU. Memory goes through the Bus, so the caller supplies the
// machine model (cartridge mapper, VRAM, WRAM, OAM, I/O registers). It implements
// the documented instruction set with the LR35902's own four-flag (Z/N/H/C)
// behaviour and Game-Boy-specific opcodes; Step returns the instruction's T-cycle
// count so the machine can drive timing (the timer, the LCD scanline counter).
//
// This is NOT a Z80: there are no IX/IY, no alternate register set, no DD/FD/ED
// prefixes (only CB), and no IN/OUT ports.

import "fmt"

// Bus is the flat 16-bit memory the CPU drives. The Game Boy has no separate I/O
// space — registers live in the memory map (the machine model decodes them).
type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
}

// Flag bits in F. The low nibble of F is always zero on the LR35902.
const (
	flagC = 1 << 4 // carry
	flagH = 1 << 5 // half carry
	flagN = 1 << 6 // add/subtract
	flagZ = 1 << 7 // zero
)

// Interrupt registers in the memory map (the CPU reads/acks them through the Bus).
const (
	regIF = 0xFF0F // interrupt flags (requests)
	regIE = 0xFFFF // interrupt enable
)

type CPU struct {
	A, F, B, C, D, E, H, L byte
	SP, PC                 uint16

	IME       bool // interrupt master enable
	imeEnable bool // EI scheduled IME for after the next instruction

	Halted     bool   // fatal stop (unimplemented opcode); inspect HaltReason
	HaltReason string
	halt       bool // executed HALT, idling until an interrupt is pending
	Stopped    bool // executed STOP
	Instrs     uint64

	bus        Bus
	takenExtra int // extra T-cycles when a conditional branch/call/ret is taken
}

// NewCPU makes a core over bus, in the DMG post-boot-ROM register state (as if the
// internal boot ROM has just handed control to the cartridge at $0100).
func NewCPU(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset restores the DMG post-boot-ROM register state.
func (c *CPU) Reset() {
	c.A, c.F = 0x01, 0xB0
	c.B, c.C = 0x00, 0x13
	c.D, c.E = 0x00, 0xD8
	c.H, c.L = 0x01, 0x4D
	c.SP, c.PC = 0xFFFE, 0x0100
	c.IME, c.imeEnable = false, false
	c.halt, c.Stopped, c.Halted = false, false, false
}

// Halt stops the core (used for an unimplemented opcode), recording why.
func (c *CPU) Halt(format string, args ...interface{}) {
	c.Halted = true
	c.HaltReason = fmt.Sprintf(format, args...)
}

func (c *CPU) read(a uint16) byte     { return c.bus.Read(a) }
func (c *CPU) write(a uint16, v byte) { c.bus.Write(a, v) }
func (c *CPU) read16(a uint16) uint16 { return uint16(c.read(a)) | uint16(c.read(a+1))<<8 }
func (c *CPU) write16(a, v uint16)    { c.write(a, byte(v)); c.write(a+1, byte(v>>8)) }
func (c *CPU) fetch() byte            { v := c.read(c.PC); c.PC++; return v }
func (c *CPU) fetch16() uint16        { v := c.read16(c.PC); c.PC += 2; return v }
func (c *CPU) push16(v uint16)        { c.SP -= 2; c.write16(c.SP, v) }
func (c *CPU) pop16() uint16          { v := c.read16(c.SP); c.SP += 2; return v }

// --- register pairs --------------------------------------------------------

func (c *CPU) bc() uint16     { return uint16(c.B)<<8 | uint16(c.C) }
func (c *CPU) de() uint16     { return uint16(c.D)<<8 | uint16(c.E) }
func (c *CPU) hl() uint16     { return uint16(c.H)<<8 | uint16(c.L) }
func (c *CPU) af() uint16     { return uint16(c.A)<<8 | uint16(c.F) }
func (c *CPU) setBC(v uint16) { c.B, c.C = byte(v>>8), byte(v) }
func (c *CPU) setDE(v uint16) { c.D, c.E = byte(v>>8), byte(v) }
func (c *CPU) setHL(v uint16) { c.H, c.L = byte(v>>8), byte(v) }
func (c *CPU) setAF(v uint16) { c.A, c.F = byte(v>>8), byte(v)&0xF0 }

// getRP/setRP: p=0 BC,1 DE,2 HL,3 SP. rp2 uses AF for p=3.
func (c *CPU) getRP(p int, af bool) uint16 {
	switch p {
	case 0:
		return c.bc()
	case 1:
		return c.de()
	case 2:
		return c.hl()
	default:
		if af {
			return c.af()
		}
		return c.SP
	}
}
func (c *CPU) setRP(p int, v uint16, af bool) {
	switch p {
	case 0:
		c.setBC(v)
	case 1:
		c.setDE(v)
	case 2:
		c.setHL(v)
	default:
		if af {
			c.setAF(v)
		} else {
			c.SP = v
		}
	}
}

// getR/setR: 8-bit register by index 0..7 (B,C,D,E,H,L,(HL),A).
func (c *CPU) getR(i int) byte {
	switch i {
	case 0:
		return c.B
	case 1:
		return c.C
	case 2:
		return c.D
	case 3:
		return c.E
	case 4:
		return c.H
	case 5:
		return c.L
	case 6:
		return c.read(c.hl())
	default:
		return c.A
	}
}
func (c *CPU) setR(i int, v byte) {
	switch i {
	case 0:
		c.B = v
	case 1:
		c.C = v
	case 2:
		c.D = v
	case 3:
		c.E = v
	case 4:
		c.H = v
	case 5:
		c.L = v
	case 6:
		c.write(c.hl(), v)
	default:
		c.A = v
	}
}

// --- flag helpers ----------------------------------------------------------

func (c *CPU) getf(b byte) bool { return c.F&b != 0 }
func (c *CPU) setFlags(z, n, h, cy bool) {
	c.F = 0
	if z {
		c.F |= flagZ
	}
	if n {
		c.F |= flagN
	}
	if h {
		c.F |= flagH
	}
	if cy {
		c.F |= flagC
	}
}

// --- 8-bit ALU (Z/N/H/C only) ----------------------------------------------

func (c *CPU) add8(n byte, carry bool) {
	cin := byte(0)
	if carry && c.getf(flagC) {
		cin = 1
	}
	r := uint16(c.A) + uint16(n) + uint16(cin)
	res := byte(r)
	c.setFlags(res == 0, false, (c.A&0xF)+(n&0xF)+cin > 0xF, r > 0xFF)
	c.A = res
}

// sub8 computes A-n(-carry); stored is false for CP (flags only).
func (c *CPU) sub8(n byte, carry, store bool) {
	cin := byte(0)
	if carry && c.getf(flagC) {
		cin = 1
	}
	r := int(c.A) - int(n) - int(cin)
	res := byte(r)
	c.setFlags(res == 0, true, int(c.A&0xF)-int(n&0xF)-int(cin) < 0, r < 0)
	if store {
		c.A = res
	}
}

func (c *CPU) and8(n byte) { c.A &= n; c.setFlags(c.A == 0, false, true, false) }
func (c *CPU) or8(n byte)  { c.A |= n; c.setFlags(c.A == 0, false, false, false) }
func (c *CPU) xor8(n byte) { c.A ^= n; c.setFlags(c.A == 0, false, false, false) }

func (c *CPU) alu(y int, n byte) {
	switch y {
	case 0:
		c.add8(n, false)
	case 1:
		c.add8(n, true)
	case 2:
		c.sub8(n, false, true)
	case 3:
		c.sub8(n, true, true)
	case 4:
		c.and8(n)
	case 5:
		c.xor8(n)
	case 6:
		c.or8(n)
	default:
		c.sub8(n, false, false) // CP
	}
}

func (c *CPU) inc8(v byte) byte {
	r := v + 1
	c.setFlags(r == 0, false, v&0xF == 0xF, c.getf(flagC)) // C preserved
	return r
}
func (c *CPU) dec8(v byte) byte {
	r := v - 1
	c.setFlags(r == 0, true, v&0xF == 0, c.getf(flagC)) // C preserved
	return r
}

// add16 is ADD HL,rp: Z preserved, N=0, H from bit 11, C from bit 15.
func (c *CPU) add16(a, b uint16) uint16 {
	r := uint32(a) + uint32(b)
	z := c.getf(flagZ)
	c.setFlags(z, false, (a&0xFFF)+(b&0xFFF) > 0xFFF, r > 0xFFFF)
	return uint16(r)
}

// addSP computes SP+e (signed); flags Z=0,N=0 with H/C from the low-byte add.
func (c *CPU) addSP(e byte) uint16 {
	res := c.SP + uint16(int8(e))
	c.setFlags(false, false, (c.SP&0xF)+(uint16(e)&0xF) > 0xF, (c.SP&0xFF)+(uint16(e)&0xFF) > 0xFF)
	return res
}

// rot runs a CB rotate/shift op y (0..7): RLC RRC RL RR SLA SRA SWAP SRL.
func (c *CPU) rot(y int, v byte) byte {
	var r byte
	cf := c.getf(flagC)
	var newC bool
	switch y {
	case 0: // RLC
		newC = v&0x80 != 0
		r = v<<1 | v>>7
	case 1: // RRC
		newC = v&1 != 0
		r = v>>1 | v<<7
	case 2: // RL
		newC = v&0x80 != 0
		r = v << 1
		if cf {
			r |= 1
		}
	case 3: // RR
		newC = v&1 != 0
		r = v >> 1
		if cf {
			r |= 0x80
		}
	case 4: // SLA
		newC = v&0x80 != 0
		r = v << 1
	case 5: // SRA
		newC = v&1 != 0
		r = v>>1 | v&0x80
	case 6: // SWAP
		r = v<<4 | v>>4
		newC = false
	default: // SRL
		newC = v&1 != 0
		r = v >> 1
	}
	c.setFlags(r == 0, false, false, newC)
	return r
}

// --- the instruction step --------------------------------------------------

// Step services any pending interrupt, then executes one instruction, and returns
// the T-cycles consumed (4 per machine cycle). A HALTed/STOPped core idles (4).
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	// Interrupt dispatch: an enabled+requested interrupt wakes HALT and, if IME is
	// set, is serviced (push PC, jump to its vector, clear its request bit).
	iff := c.read(regIF)
	pend := c.read(regIE) & iff & 0x1F
	if pend != 0 {
		c.halt = false
	}
	if c.IME && pend != 0 {
		c.IME, c.imeEnable = false, false
		bit := 0
		for ; bit < 5; bit++ {
			if pend&(1<<bit) != 0 {
				break
			}
		}
		c.write(regIF, iff&^(1<<uint(bit)))
		c.push16(c.PC)
		c.PC = uint16(0x40 + bit*8)
		c.Instrs++
		return 20
	}
	if c.halt || c.Stopped {
		c.Instrs++
		return 4
	}

	pendingEI := c.imeEnable
	c.takenExtra = 0
	c.Instrs++
	op := c.fetch()
	cyc := 0
	if op == 0xCB {
		cb := c.fetch()
		c.execCB(cb)
		cyc = cbCycles(cb)
	} else {
		c.execMain(op)
		cyc = int(mainCycles[op]) + c.takenExtra
	}
	// A pending EI (set by an EI executed in the PREVIOUS step) takes effect now,
	// unless the just-executed instruction cancelled it (DI clears imeEnable).
	if pendingEI && c.imeEnable {
		c.IME, c.imeEnable = true, false
	}
	return cyc
}

func (c *CPU) execMain(op byte) {
	x, y, z := op>>6, int(op>>3)&7, int(op)&7
	p, q := y>>1, y&1
	switch x {
	case 0:
		c.execX0(y, z, p, q)
	case 1:
		if z == 6 && y == 6 {
			c.halt = true // HALT
			return
		}
		c.setR(y, c.getR(z)) // LD r,r
	case 2:
		c.alu(y, c.getR(z))
	default:
		c.execX3(y, z, p, q)
	}
}

func (c *CPU) execX0(y, z, p, q int) {
	switch z {
	case 0:
		switch y {
		case 0: // NOP
		case 1: // LD (a16),SP
			c.write16(c.fetch16(), c.SP)
		case 2: // STOP
			c.fetch() // ignored padding byte
			c.Stopped = true
		case 3: // JR e
			off := int8(c.fetch())
			c.PC = uint16(int(c.PC) + int(off))
		default: // JR cc,e
			off := int8(c.fetch())
			if c.cond(y - 4) {
				c.PC = uint16(int(c.PC) + int(off))
				c.takenExtra = 4
			}
		}
	case 1:
		if q == 0 { // LD rp,nn
			c.setRP(p, c.fetch16(), false)
		} else { // ADD HL,rp
			c.setHL(c.add16(c.hl(), c.getRP(p, false)))
		}
	case 2:
		switch {
		case q == 0 && p == 0:
			c.write(c.bc(), c.A) // LD (BC),A
		case q == 0 && p == 1:
			c.write(c.de(), c.A) // LD (DE),A
		case q == 0 && p == 2: // LD (HL+),A
			c.write(c.hl(), c.A)
			c.setHL(c.hl() + 1)
		case q == 0 && p == 3: // LD (HL-),A
			c.write(c.hl(), c.A)
			c.setHL(c.hl() - 1)
		case q == 1 && p == 0:
			c.A = c.read(c.bc()) // LD A,(BC)
		case q == 1 && p == 1:
			c.A = c.read(c.de()) // LD A,(DE)
		case q == 1 && p == 2: // LD A,(HL+)
			c.A = c.read(c.hl())
			c.setHL(c.hl() + 1)
		default: // LD A,(HL-)
			c.A = c.read(c.hl())
			c.setHL(c.hl() - 1)
		}
	case 3:
		if q == 0 {
			c.setRP(p, c.getRP(p, false)+1, false) // INC rp
		} else {
			c.setRP(p, c.getRP(p, false)-1, false) // DEC rp
		}
	case 4: // INC r
		c.setR(y, c.inc8(c.getR(y)))
	case 5: // DEC r
		c.setR(y, c.dec8(c.getR(y)))
	case 6: // LD r,n
		c.setR(y, c.fetch())
	default: // z==7: accumulator/flag ops
		c.accOp(y)
	}
}

func (c *CPU) accOp(y int) {
	switch y {
	case 0: // RLCA
		cy := c.A&0x80 != 0
		c.A = c.A<<1 | c.A>>7
		c.setFlags(false, false, false, cy)
	case 1: // RRCA
		cy := c.A&1 != 0
		c.A = c.A>>1 | c.A<<7
		c.setFlags(false, false, false, cy)
	case 2: // RLA
		old := c.getf(flagC)
		cy := c.A&0x80 != 0
		c.A <<= 1
		if old {
			c.A |= 1
		}
		c.setFlags(false, false, false, cy)
	case 3: // RRA
		old := c.getf(flagC)
		cy := c.A&1 != 0
		c.A >>= 1
		if old {
			c.A |= 0x80
		}
		c.setFlags(false, false, false, cy)
	case 4: // DAA
		c.daa()
	case 5: // CPL
		c.A = ^c.A
		c.F |= flagN | flagH
	case 6: // SCF
		c.F = c.F&flagZ | flagC
	default: // CCF
		c.F = c.F&flagZ | (c.F^flagC)&flagC
	}
}

func (c *CPU) daa() {
	var corr byte
	carry := c.getf(flagC)
	if c.getf(flagH) || (!c.getf(flagN) && c.A&0x0F > 9) {
		corr |= 0x06
	}
	if c.getf(flagC) || (!c.getf(flagN) && c.A > 0x99) {
		corr |= 0x60
		carry = true
	}
	if c.getf(flagN) {
		c.A -= corr
	} else {
		c.A += corr
	}
	c.F &^= flagZ | flagH | flagC
	if c.A == 0 {
		c.F |= flagZ
	}
	if carry {
		c.F |= flagC
	}
}

func (c *CPU) execX3(y, z, p, q int) {
	switch z {
	case 0:
		switch y {
		case 0, 1, 2, 3: // RET cc
			if c.cond(y) {
				c.PC = c.pop16()
				c.takenExtra = 12
			}
		case 4: // LDH (a8),A
			c.write(0xFF00|uint16(c.fetch()), c.A)
		case 5: // ADD SP,e
			c.SP = c.addSP(c.fetch())
		case 6: // LDH A,(a8)
			c.A = c.read(0xFF00 | uint16(c.fetch()))
		default: // LD HL,SP+e
			c.setHL(c.addSP(c.fetch()))
		}
	case 1:
		if q == 0 {
			c.setRP(p, c.pop16(), true) // POP rp2
		} else {
			switch p {
			case 0: // RET
				c.PC = c.pop16()
			case 1: // RETI
				c.PC = c.pop16()
				c.IME, c.imeEnable = true, false
			case 2: // JP (HL)
				c.PC = c.hl()
			default: // LD SP,HL
				c.SP = c.hl()
			}
		}
	case 2:
		switch y {
		case 0, 1, 2, 3: // JP cc,nn
			a := c.fetch16()
			if c.cond(y) {
				c.PC = a
				c.takenExtra = 4
			}
		case 4: // LDH (C),A
			c.write(0xFF00|uint16(c.C), c.A)
		case 5: // LD (a16),A
			c.write(c.fetch16(), c.A)
		case 6: // LDH A,(C)
			c.A = c.read(0xFF00 | uint16(c.C))
		default: // LD A,(a16)
			c.A = c.read(c.fetch16())
		}
	case 3:
		switch y {
		case 0: // JP nn
			c.PC = c.fetch16()
		case 6: // DI
			c.IME, c.imeEnable = false, false
		case 7: // EI
			c.imeEnable = true
		default: // illegal ($D3/$DB/$E3/$EB)
			c.Halt("illegal opcode at $%04X", c.PC-1)
		}
	case 4:
		if y <= 3 { // CALL cc,nn
			a := c.fetch16()
			if c.cond(y) {
				c.push16(c.PC)
				c.PC = a
				c.takenExtra = 12
			}
		} else {
			c.Halt("illegal opcode at $%04X", c.PC-1)
		}
	case 5:
		if q == 0 {
			c.push16(c.getRP(p, true)) // PUSH rp2
		} else if p == 0 { // CALL nn
			a := c.fetch16()
			c.push16(c.PC)
			c.PC = a
		} else {
			c.Halt("illegal opcode at $%04X", c.PC-1)
		}
	case 6: // alu n
		c.alu(y, c.fetch())
	default: // z==7: RST
		c.push16(c.PC)
		c.PC = uint16(y) * 8
	}
}

// cond evaluates condition cc[y]: NZ Z NC C.
func (c *CPU) cond(y int) bool {
	switch y {
	case 0:
		return !c.getf(flagZ)
	case 1:
		return c.getf(flagZ)
	case 2:
		return !c.getf(flagC)
	default:
		return c.getf(flagC)
	}
}

func (c *CPU) execCB(op byte) {
	x, y, z := op>>6, int(op>>3)&7, int(op)&7
	switch x {
	case 0:
		c.setR(z, c.rot(y, c.getR(z)))
	case 1: // BIT y,r (Z=bit==0, N=0, H=1, C preserved)
		v := c.getR(z)
		c.F = c.F&flagC | flagH
		if v&(1<<uint(y)) == 0 {
			c.F |= flagZ
		}
	case 2: // RES
		c.setR(z, c.getR(z)&^(1<<uint(y)))
	default: // SET
		c.setR(z, c.getR(z)|1<<uint(y))
	}
}

// cbCycles returns the T-cycle count of a CB-prefixed op: 8 for register operands,
// 12 for BIT b,(HL) (read only), 16 for the read-modify-write (HL) forms.
func cbCycles(op byte) int {
	if op&7 != 6 { // register operand
		return 8
	}
	if op>>6 == 1 { // BIT b,(HL)
		return 12
	}
	return 16 // rot/RES/SET (HL)
}

// mainCycles is the base T-cycle count of each unprefixed opcode (the untaken count
// for conditional branch/call/ret, whose taken case adds takenExtra). Illegal
// opcodes are 0.
var mainCycles = [256]uint8{
	4, 12, 8, 8, 4, 4, 8, 4, 20, 8, 8, 8, 4, 4, 8, 4, // 0x
	4, 12, 8, 8, 4, 4, 8, 4, 12, 8, 8, 8, 4, 4, 8, 4, // 1x
	8, 12, 8, 8, 4, 4, 8, 4, 8, 8, 8, 8, 4, 4, 8, 4, // 2x
	8, 12, 8, 8, 12, 12, 12, 4, 8, 8, 8, 8, 4, 4, 8, 4, // 3x
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 4x
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 5x
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 6x
	8, 8, 8, 8, 8, 8, 4, 8, 4, 4, 4, 4, 4, 4, 8, 4, // 7x (0x76 HALT=4)
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 8x
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 9x
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // Ax
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // Bx
	8, 12, 12, 16, 12, 16, 8, 16, 8, 16, 12, 4, 12, 24, 8, 16, // Cx
	8, 12, 12, 0, 12, 16, 8, 16, 8, 16, 12, 0, 12, 0, 8, 16, // Dx
	12, 12, 8, 0, 0, 16, 8, 16, 16, 4, 16, 0, 0, 0, 8, 16, // Ex
	12, 12, 8, 4, 0, 16, 8, 16, 12, 8, 16, 4, 0, 0, 8, 16, // Fx
}
