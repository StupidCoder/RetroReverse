package x86

// This file is the execution half of the CPU core: Step() fetches prefixes and
// one opcode from CS:IP and dispatches it, and the helpers below implement the
// instruction subset (see the package/cpu.go scope note). Effective addresses
// go through a small ea value so register and memory operands share one path.

// ea is a decoded ModR/M operand: a register index, or a base:off memory ref
// (base is the resolved segment linear base; see segBase).
type ea struct {
	isReg bool
	reg   byte
	base  uint32
	off   uint32
}

func (c *CPU) rEA(o ea, b int) uint32 {
	if o.isReg {
		return c.getReg(o.reg, b)
	}
	return c.memRead(o.base, o.off, b)
}
func (c *CPU) wEA(o ea, b int, v uint32) {
	if o.isReg {
		c.setReg(o.reg, b, v)
		return
	}
	c.memWrite(o.base, o.off, b, v)
}

// osz is the current operand size in bytes (2 or 4).
func (c *CPU) osz() int {
	if c.dOpsize == 32 {
		return 4
	}
	return 2
}

// asz is the current address size in bytes (2 or 4) — the width of the string
// index registers (SI/DI), REP/LOOP counters (CX), and XLAT's base register.
func (c *CPU) asz() int {
	if c.dAddrsize == 32 {
		return 4
	}
	return 2
}

// segBaseFor returns the linear base of segment def, or of the override segment
// when a segment-prefix is active.
func (c *CPU) segBaseFor(def int) uint32 {
	if c.dSeg >= 0 {
		return c.segBase(c.dSeg)
	}
	return c.segBase(def)
}

func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// inPort/outPort route IN/OUT through the machine's port hooks (an absent
// device reads as all-ones). size is in bytes (1/2).
func (c *CPU) inPort(port uint16, size int) uint32 {
	if c.PortIn != nil {
		return c.PortIn(port, size) & widthMask(size)
	}
	return widthMask(size)
}
func (c *CPU) outPort(port uint16, size int, v uint32) {
	if c.PortOut != nil {
		c.PortOut(port, size, v)
	}
}

func signExtToInt(v uint32, w int) int32 {
	switch w {
	case 1:
		return int32(int8(byte(v)))
	case 2:
		return int32(int16(uint16(v)))
	default:
		return int32(v)
	}
}

// modrmE decodes a ModR/M byte for execution (fetching SIB/displacement).
func (c *CPU) modrmE() (byte, ea) {
	mb := c.fetch8()
	mod, reg, rm := byte(mb>>6), byte((mb>>3)&7), byte(mb&7)
	if mod == 3 {
		return reg, ea{isReg: true, reg: rm}
	}
	if c.dAddrsize == 32 {
		return reg, c.ea32(mod, rm)
	}
	return reg, c.ea16(mod, rm)
}

func (c *CPU) ea16(mod, rm byte) ea {
	var off uint32
	defSeg := DS
	switch rm {
	case 0:
		off = c.gw(BX) + c.gw(SI)
	case 1:
		off = c.gw(BX) + c.gw(DI)
	case 2:
		off = c.gw(BP) + c.gw(SI)
		defSeg = SS
	case 3:
		off = c.gw(BP) + c.gw(DI)
		defSeg = SS
	case 4:
		off = c.gw(SI)
	case 5:
		off = c.gw(DI)
	case 6:
		if mod == 0 {
			off = c.fetch16()
		} else {
			off = c.gw(BP)
			defSeg = SS
		}
	case 7:
		off = c.gw(BX)
	}
	switch mod {
	case 1:
		off += signExtByte(c.fetch8())
	case 2:
		off += c.fetch16()
	}
	seg := defSeg
	if c.dSeg >= 0 {
		seg = c.dSeg
	}
	return ea{base: c.segBase(seg), off: off & 0xFFFF}
}

func (c *CPU) ea32(mod, rm byte) ea {
	var off uint32
	defSeg := DS
	switch {
	case rm == 4: // SIB
		sib := c.fetch8()
		scale, idx, base := byte(sib>>6), byte((sib>>3)&7), byte(sib&7)
		if idx != 4 {
			off += c.Regs[idx] << scale
		}
		if base == 5 && mod == 0 {
			off += c.fetch32()
		} else {
			off += c.Regs[base]
			if base == SP || base == BP {
				defSeg = SS
			}
		}
	case rm == 5 && mod == 0:
		off = c.fetch32()
	default:
		off = c.Regs[rm]
		if rm == SP || rm == BP {
			defSeg = SS
		}
	}
	switch mod {
	case 1:
		off += signExtByte(c.fetch8())
	case 2:
		off += c.fetch32()
	}
	seg := defSeg
	if c.dSeg >= 0 {
		seg = c.dSeg
	}
	return ea{base: c.segBase(seg), off: off}
}

// Step executes one instruction (prefixes + opcode).
func (c *CPU) Step() {
	if c.Halted {
		return
	}
	if c.OnStep != nil {
		c.OnStep(c) // may inject an IRQ (Interrupt() honours ssShadow)
		if c.Halted {
			return
		}
	}
	c.ssShadow = false // the shadow blocks only the single boundary above
	c.Steps++
	c.instrIP = c.IP // start of this instruction, for a fault's pushed CS:(E)IP
	// Default operand/address size follows the mode (the CS descriptor's D bit):
	// 16-bit in real mode, 32-bit in flat protected mode. A 0x66/0x67 prefix
	// selects the opposite size.
	c.dSeg = -1
	alt := 32
	if c.Mode == ModeProt {
		c.dOpsize, c.dAddrsize = 32, 32
		alt = 16
	} else {
		c.dOpsize, c.dAddrsize = 16, 16
	}
	var rep byte
	var op byte
	for {
		op = byte(c.fetch8())
		prefix := true
		switch op {
		case 0x26:
			c.dSeg = ES
		case 0x2E:
			c.dSeg = CS
		case 0x36:
			c.dSeg = SS
		case 0x3E:
			c.dSeg = DS
		case 0x64:
			c.dSeg = FS
		case 0x65:
			c.dSeg = GS
		case 0x66:
			c.dOpsize = alt
			c.Ext386++
		case 0x67:
			c.dAddrsize = alt
			c.Ext386++
		case 0xF0: // LOCK — no effect on a single-core model
		case 0xF2:
			rep = 0xF2
		case 0xF3:
			rep = 0xF3
		default:
			prefix = false
		}
		if !prefix {
			break
		}
	}
	c.exec(op, rep)
}

// Run steps until the CPU halts or maxSteps instructions have executed. It
// returns the number of steps taken.
func (c *CPU) Run(maxSteps uint64) uint64 {
	var n uint64
	for !c.Halted && n < maxSteps {
		c.Step()
		n++
	}
	return n
}

// alu applies ALU op idx (0..7 = ADD OR ADC SBB AND SUB XOR CMP) to a,b at
// width w, sets flags, and returns (result, writeBack).
func (c *CPU) alu(idx int, a, b uint32, w int) (uint32, bool) {
	switch idx {
	case 0:
		return c.flagsAdd(a, b, 0, w), true
	case 1:
		return c.flagsLogic(a|b, w), true
	case 2:
		return c.flagsAdd(a, b, b2u(c.CF), w), true
	case 3:
		return c.flagsSub(a, b, b2u(c.CF), w), true
	case 4:
		return c.flagsLogic(a&b, w), true
	case 5:
		return c.flagsSub(a, b, 0, w), true
	case 6:
		return c.flagsLogic(a^b, w), true
	default: // CMP
		return c.flagsSub(a, b, 0, w), false
	}
}

// exec dispatches one opcode. rep is 0/0xF2/0xF3 (a string-op prefix).
func (c *CPU) exec(op, rep byte) {
	// The 0x00–0x3F ALU grid (columns op&7 < 6).
	if op < 0x40 && op&7 < 6 {
		c.aluGrid(int(op>>3), op&7)
		return
	}

	switch {
	case op >= 0x40 && op <= 0x47: // INC r16
		c.setReg(op&7, c.osz(), c.incDec(c.getReg(op&7, c.osz()), c.osz(), true))
		return
	case op >= 0x48 && op <= 0x4F: // DEC r16
		c.setReg(op&7, c.osz(), c.incDec(c.getReg(op&7, c.osz()), c.osz(), false))
		return
	case op >= 0x50 && op <= 0x57: // PUSH r16
		c.push(c.osz(), c.getReg(op&7, c.osz()))
		return
	case op >= 0x58 && op <= 0x5F: // POP r16
		c.setReg(op&7, c.osz(), c.pop(c.osz()))
		return
	case op >= 0x70 && op <= 0x7F: // Jcc rel8
		rel := signExtByte(c.fetch8())
		if c.cond(op & 0x0F) {
			c.IP = c.ipMask(c.IP + rel)
		}
		return
	case op >= 0x91 && op <= 0x97: // XCHG AX, r16
		i := op & 7
		w := c.osz()
		t := c.getReg(0, w)
		c.setReg(0, w, c.getReg(i, w))
		c.setReg(i, w, t)
		return
	case op >= 0xB0 && op <= 0xB7: // MOV r8, imm8
		c.s8(op&7, c.fetch8())
		return
	case op >= 0xB8 && op <= 0xBF: // MOV r16, imm
		c.setReg(op&7, c.osz(), c.fetchImm())
		return
	}

	switch op {
	case 0x06:
		c.push(c.osz(), uint32(c.Seg[ES]))
	case 0x07:
		c.loadSeg(ES, uint16(c.pop(c.osz())))
	case 0x0E:
		c.push(c.osz(), uint32(c.Seg[CS]))
	case 0x16:
		c.push(c.osz(), uint32(c.Seg[SS]))
	case 0x17:
		c.loadSeg(SS, uint16(c.pop(c.osz())))
		c.ssShadow = true
	case 0x1E:
		c.push(c.osz(), uint32(c.Seg[DS]))
	case 0x1F:
		c.loadSeg(DS, uint16(c.pop(c.osz())))

	case 0x27:
		c.daa(false)
	case 0x2F:
		c.daa(true)
	case 0x37:
		c.aaa(false)
	case 0x3F:
		c.aaa(true)
	case 0xD4:
		c.aam(byte(c.fetch8()))
	case 0xD5:
		c.aad(byte(c.fetch8()))

	case 0x0F:
		c.Ext386++
		c.exec0F(byte(c.fetch8()))
	case 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF: // x87 escape
		c.fpuExec(op)

	case 0x60: // PUSHA / PUSHAD
		w := c.osz()
		sp := c.getReg(SP, w)
		for _, r := range []byte{AX, CX, DX, BX} {
			c.push(w, c.getReg(r, w))
		}
		c.push(w, sp) // the SP value before this instruction
		for _, r := range []byte{BP, SI, DI} {
			c.push(w, c.getReg(r, w))
		}
	case 0x61: // POPA / POPAD
		w := c.osz()
		for _, r := range []byte{DI, SI, BP} {
			c.setReg(r, w, c.pop(w))
		}
		c.pop(w) // the saved SP slot is discarded
		for _, r := range []byte{BX, DX, CX, AX} {
			c.setReg(r, w, c.pop(w))
		}
	case 0x62: // BOUND r, m — array bounds check; modelled as a no-op
		c.modrmE()
	case 0x63: // ARPL r/m16, r16 — protected-mode only; no-op in real mode
		c.modrmE()
	case 0x6C, 0x6D, 0x6E, 0x6F: // INS / OUTS (port string ops)
		c.stringPortOp(op, rep)
	case 0x9B: // FWAIT / WAIT — no FPU exceptions modelled
	case 0xF1: // INT1 / ICEBP
		c.doInt(1)

	case 0x68: // PUSH imm
		c.push(c.osz(), c.fetchImm())
	case 0x69: // IMUL r, r/m, imm
		reg, o := c.modrmE()
		w := c.osz()
		src := c.rEA(o, w)
		imm := c.fetchImm()
		c.setReg(reg, w, c.imulTrunc(src, imm, w))
	case 0x6A: // PUSH imm8 (sign-extended)
		c.push(c.osz(), signExtByte(c.fetch8()))
	case 0x6B: // IMUL r, r/m, imm8
		reg, o := c.modrmE()
		w := c.osz()
		src := c.rEA(o, w)
		imm := signExtByte(c.fetch8())
		c.setReg(reg, w, c.imulTrunc(src, imm, w))

	case 0x80, 0x81, 0x82, 0x83:
		c.grp1(op)

	case 0x84: // TEST r/m8, r8
		reg, o := c.modrmE()
		c.flagsLogic(c.rEA(o, 1)&c.g8(reg), 1)
	case 0x85: // TEST r/m, r
		reg, o := c.modrmE()
		w := c.osz()
		c.flagsLogic(c.rEA(o, w)&c.getReg(reg, w), w)
	case 0x86: // XCHG r/m8, r8
		reg, o := c.modrmE()
		t := c.rEA(o, 1)
		c.wEA(o, 1, c.g8(reg))
		c.s8(reg, t)
	case 0x87: // XCHG r/m, r
		reg, o := c.modrmE()
		w := c.osz()
		t := c.rEA(o, w)
		c.wEA(o, w, c.getReg(reg, w))
		c.setReg(reg, w, t)

	case 0x88: // MOV r/m8, r8
		reg, o := c.modrmE()
		c.wEA(o, 1, c.g8(reg))
	case 0x89: // MOV r/m, r
		reg, o := c.modrmE()
		w := c.osz()
		c.wEA(o, w, c.getReg(reg, w))
	case 0x8A: // MOV r8, r/m8
		reg, o := c.modrmE()
		c.s8(reg, c.rEA(o, 1))
	case 0x8B: // MOV r, r/m
		reg, o := c.modrmE()
		w := c.osz()
		c.setReg(reg, w, c.rEA(o, w))
	case 0x8C: // MOV r/m16, sreg
		reg, o := c.modrmE()
		c.wEA(o, 2, uint32(c.Seg[reg&7]))
	case 0x8D: // LEA r, m
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), o.off)
	case 0x8E: // MOV sreg, r/m16
		reg, o := c.modrmE()
		c.loadSeg(int(reg&7), uint16(c.rEA(o, 2)))
		if reg&7 == SS {
			c.ssShadow = true
		}
	case 0x8F: // POP r/m
		_, o := c.modrmE()
		c.wEA(o, c.osz(), c.pop(c.osz()))

	case 0x90: // NOP (XCHG AX,AX)
	case 0x98: // CBW / CWDE
		if c.dOpsize == 32 {
			c.Regs[AX] = uint32(int32(int16(uint16(c.gw(AX)))))
		} else {
			c.s16(AX, signExtByte(c.g8(0)))
		}
	case 0x99: // CWD / CDQ
		if c.dOpsize == 32 {
			if c.Regs[AX]&0x80000000 != 0 {
				c.Regs[DX] = 0xFFFFFFFF
			} else {
				c.Regs[DX] = 0
			}
		} else if c.gw(AX)&0x8000 != 0 {
			c.s16(DX, 0xFFFF)
		} else {
			c.s16(DX, 0)
		}
	case 0x9A: // CALL far ptr16:16 / ptr16:32
		off := c.fetchImm()
		seg := c.fetch16()
		c.push(c.osz(), uint32(c.Seg[CS]))
		c.push(c.osz(), c.IP)
		c.loadSeg(CS, uint16(seg))
		c.IP = c.ipMask(off)
	case 0x9C: // PUSHF / PUSHFD
		c.push(c.osz(), uint32(c.EFlags()))
	case 0x9D: // POPF / POPFD
		c.SetEFlags(uint16(c.pop(c.osz())))
	case 0x9E: // SAHF
		c.SetEFlags((c.EFlags() &^ 0xFF) | uint16(c.g8(4))) // AH into low byte
	case 0x9F: // LAHF
		c.s8(4, uint32(byte(c.EFlags())))

	case 0xA0: // MOV AL, moffs8
		c.s8(0, c.memRead(c.segBaseFor(DS), c.moffs(), 1))
	case 0xA1: // MOV eAX, moffs
		w := c.osz()
		c.setReg(AX, w, c.memRead(c.segBaseFor(DS), c.moffs(), w))
	case 0xA2: // MOV moffs8, AL
		c.memWrite(c.segBaseFor(DS), c.moffs(), 1, c.g8(0))
	case 0xA3: // MOV moffs, eAX
		w := c.osz()
		c.memWrite(c.segBaseFor(DS), c.moffs(), w, c.getReg(AX, w))
	case 0xA4, 0xA5, 0xA6, 0xA7, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF:
		c.stringOp(op, rep)
	case 0xA8: // TEST AL, imm8
		c.flagsLogic(c.g8(0)&c.fetch8(), 1)
	case 0xA9: // TEST eAX, imm
		w := c.osz()
		c.flagsLogic(c.getReg(AX, w)&c.fetchImm(), w)

	case 0xC0: // grp2 r/m8, imm8
		c.grp2(1, func() uint32 { return c.fetch8() })
	case 0xC1: // grp2 r/m, imm8
		c.grp2(c.osz(), func() uint32 { return c.fetch8() })
	case 0xC2: // RET imm16
		n := c.fetch16()
		c.IP = c.ipMask(c.pop(c.osz()))
		c.spAdd(n)
	case 0xC3: // RET
		c.IP = c.ipMask(c.pop(c.osz()))
	case 0xC4: // LES r, m
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), c.rEA(o, 2))
		c.loadSeg(ES, uint16(c.memRead(o.base, o.off+2, 2)))
	case 0xC5: // LDS r, m
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), c.rEA(o, 2))
		c.loadSeg(DS, uint16(c.memRead(o.base, o.off+2, 2)))
	case 0xC6: // MOV r/m8, imm8
		_, o := c.modrmE()
		c.wEA(o, 1, c.fetch8())
	case 0xC7: // MOV r/m, imm
		_, o := c.modrmE()
		c.wEA(o, c.osz(), c.fetchImm())
	case 0xC8: // ENTER imm16, imm8
		sz := c.fetch16()
		lvl := uint32(c.fetch8() & 0x1F)
		w := c.osz()
		c.push(w, c.getReg(BP, w))
		frame := c.getReg(SP, w)
		for i := uint32(1); i < lvl; i++ {
			c.setReg(BP, w, c.getReg(BP, w)-uint32(w))
			c.push(w, c.getReg(BP, w))
		}
		if lvl > 0 {
			c.push(w, frame)
		}
		c.setReg(BP, w, frame)
		c.setReg(SP, w, c.getReg(SP, w)-sz)
	case 0xC9: // LEAVE
		w := c.osz()
		c.setReg(SP, w, c.getReg(BP, w))
		c.setReg(BP, w, c.pop(w))
	case 0xCA: // RETF imm16
		n := c.fetch16()
		c.IP = c.ipMask(c.pop(c.osz()))
		c.loadSeg(CS, uint16(c.pop(c.osz())))
		c.spAdd(n)
	case 0xCB: // RETF
		c.IP = c.ipMask(c.pop(c.osz()))
		c.loadSeg(CS, uint16(c.pop(c.osz())))
	case 0xCC: // INT3
		c.doInt(3)
	case 0xCD: // INT imm8
		c.doInt(byte(c.fetch8()))
	case 0xCE: // INTO
		if c.OF {
			c.doInt(4)
		}
	case 0xCF: // IRET / IRETD
		c.IP = c.ipMask(c.pop(c.osz()))
		c.loadSeg(CS, uint16(c.pop(c.osz())))
		c.SetEFlags(uint16(c.pop(c.osz())))

	case 0xD0: // grp2 r/m8, 1
		c.grp2(1, func() uint32 { return 1 })
	case 0xD1: // grp2 r/m, 1
		c.grp2(c.osz(), func() uint32 { return 1 })
	case 0xD2: // grp2 r/m8, CL
		c.grp2(1, func() uint32 { return c.g8(1) })
	case 0xD3: // grp2 r/m, CL
		c.grp2(c.osz(), func() uint32 { return c.g8(1) })
	case 0xD7: // XLAT
		aw := c.asz()
		c.s8(0, c.memRead(c.segBaseFor(DS), c.getReg(BX, aw)+c.g8(0), 1))

	case 0xE0, 0xE1, 0xE2: // LOOPNE / LOOPE / LOOP
		rel := signExtByte(c.fetch8())
		aw := c.asz()
		c.setReg(CX, aw, c.getReg(CX, aw)-1)
		take := c.getReg(CX, aw) != 0
		if op == 0xE1 {
			take = take && c.ZF
		} else if op == 0xE0 {
			take = take && !c.ZF
		}
		if take {
			c.IP = c.ipMask(c.IP + rel)
		}
	case 0xE3: // JCXZ / JECXZ
		rel := signExtByte(c.fetch8())
		if c.getReg(CX, c.asz()) == 0 {
			c.IP = c.ipMask(c.IP + rel)
		}
	case 0xE8: // CALL rel
		rel := c.fetchImm()
		if c.dOpsize != 32 {
			rel = signExtWord(rel)
		}
		c.push(c.osz(), c.IP)
		c.IP = c.ipMask(c.IP + rel)
	case 0xE9: // JMP rel
		rel := c.fetchImm()
		if c.dOpsize != 32 {
			rel = signExtWord(rel)
		}
		c.IP = c.ipMask(c.IP + rel)
	case 0xEA: // JMP far ptr16:16 / ptr16:32
		off := c.fetchImm()
		seg := c.fetch16()
		c.loadSeg(CS, uint16(seg))
		c.IP = c.ipMask(off)
	case 0xEB: // JMP rel8
		rel := signExtByte(c.fetch8())
		c.IP = c.ipMask(c.IP + rel)

	case 0xE4: // IN AL, imm8
		c.s8(0, c.inPort(uint16(c.fetch8()), 1))
	case 0xE5: // IN eAX, imm8
		c.setReg(AX, c.osz(), c.inPort(uint16(c.fetch8()), c.osz()))
	case 0xE6: // OUT imm8, AL
		c.outPort(uint16(c.fetch8()), 1, c.g8(0))
	case 0xE7: // OUT imm8, eAX
		c.outPort(uint16(c.fetch8()), c.osz(), c.getReg(AX, c.osz()))
	case 0xEC: // IN AL, DX
		c.s8(0, c.inPort(uint16(c.gw(DX)), 1))
	case 0xED: // IN eAX, DX
		c.setReg(AX, c.osz(), c.inPort(uint16(c.gw(DX)), c.osz()))
	case 0xEE: // OUT DX, AL
		c.outPort(uint16(c.gw(DX)), 1, c.g8(0))
	case 0xEF: // OUT DX, eAX
		c.outPort(uint16(c.gw(DX)), c.osz(), c.getReg(AX, c.osz()))

	case 0xF4: // HLT
		c.Halt("HLT at %s", c.at())
	case 0xF5: // CMC
		c.CF = !c.CF
	case 0xF6: // grp3 r/m8
		c.grp3(1)
	case 0xF7: // grp3 r/m
		c.grp3(c.osz())
	case 0xF8:
		c.CF = false
	case 0xF9:
		c.CF = true
	case 0xFA:
		c.IF = false
	case 0xFB:
		c.IF = true
	case 0xFC:
		c.DF = false
	case 0xFD:
		c.DF = true
	case 0xFE: // grp4 (INC/DEC r/m8)
		reg, o := c.modrmE()
		v := c.rEA(o, 1)
		c.wEA(o, 1, c.incDec(v, 1, reg == 0))
	case 0xFF: // grp5
		c.grp5()

	default:
		c.Halt("unimplemented opcode $%02X at %s", op, c.at())
	}
}

// moffs fetches a MOV moffs direct offset (address-size wide).
func (c *CPU) moffs() uint32 {
	if c.dAddrsize == 32 {
		return c.fetch32()
	}
	return c.fetch16()
}

// aluGrid executes one column (z) of the primary ALU grid for op idx.
func (c *CPU) aluGrid(idx int, z byte) {
	switch z {
	case 0:
		reg, o := c.modrmE()
		res, wr := c.alu(idx, c.rEA(o, 1), c.g8(reg), 1)
		if wr {
			c.wEA(o, 1, res)
		}
	case 1:
		reg, o := c.modrmE()
		w := c.osz()
		res, wr := c.alu(idx, c.rEA(o, w), c.getReg(reg, w), w)
		if wr {
			c.wEA(o, w, res)
		}
	case 2:
		reg, o := c.modrmE()
		res, wr := c.alu(idx, c.g8(reg), c.rEA(o, 1), 1)
		if wr {
			c.s8(reg, res)
		}
	case 3:
		reg, o := c.modrmE()
		w := c.osz()
		res, wr := c.alu(idx, c.getReg(reg, w), c.rEA(o, w), w)
		if wr {
			c.setReg(reg, w, res)
		}
	case 4:
		res, wr := c.alu(idx, c.g8(0), c.fetch8(), 1)
		if wr {
			c.s8(0, res)
		}
	default: // 5
		w := c.osz()
		res, wr := c.alu(idx, c.getReg(AX, w), c.fetchImm(), w)
		if wr {
			c.setReg(AX, w, res)
		}
	}
}

// incDec adds or subtracts 1 preserving CF (the defining trait of INC/DEC).
func (c *CPU) incDec(v uint32, w int, inc bool) uint32 {
	saved := c.CF
	var res uint32
	if inc {
		res = c.flagsAdd(v, 1, 0, w)
	} else {
		res = c.flagsSub(v, 1, 0, w)
	}
	c.CF = saved
	return res
}

// doInt services a software interrupt, preferring the IntHook. In protected
// mode there is no real-mode IVT to fall back to (a flat go32/Xbox image routes
// every INT through the host's IntHook), so an unhandled INT halts with a clear
// reason rather than vectoring through low-memory garbage.
func (c *CPU) doInt(n byte) {
	if c.IntHook != nil && c.IntHook(c, n) {
		return
	}
	if c.Mode == ModeProt {
		c.Halt("unhandled INT $%02X in protected mode at %08X", n, c.instrIP)
		return
	}
	c.dispatchIVT(n)
}

// divErr raises the divide-error exception (#DE) by vectoring through IVT[0],
// as a real CPU does on a zero divisor or overflowing quotient. It pushes the
// address of the FAULTING instruction (the 286-and-later behaviour), not the
// following one (the 8086/8088 quirk) — this core models a 386-class machine.
// The distinction is load-bearing: a #DE handler that restarts or adjusts the
// faulting instruction (UW's fixed-point renderer skips the 2-byte IDIV by
// adding 2 to the pushed IP, so it lands on the instruction's own result-check)
// only lands correctly when the pushed IP points at the IDIV itself.
func (c *CPU) divErr() {
	c.IP = c.instrIP
	c.dispatchIVT(0)
}

// dispatchIVT performs the real-mode interrupt sequence: push FLAGS/CS/IP,
// clear IF/TF, and vector through IVT[n].
func (c *CPU) dispatchIVT(n byte) {
	c.push16(uint32(c.EFlags()))
	c.IF, c.TF = false, false
	c.push16(uint32(c.Seg[CS]))
	c.push16(c.IP)
	c.IP = c.rd16(uint32(n) * 4)
	c.Seg[CS] = uint16(c.rd16(uint32(n)*4 + 2))
}

// Interrupt delivers a maskable hardware interrupt n through the IVT, honoring
// the interrupt-enable flag. It returns true if the interrupt was delivered
// (IF was set); the caller is responsible for only invoking it at an
// instruction boundary. Used by a machine model to inject the timer IRQ, etc.
func (c *CPU) Interrupt(n byte) bool {
	if !c.IF || c.Halted || c.ssShadow {
		return false
	}
	c.dispatchIVT(n)
	return true
}

// InterruptPM delivers a maskable hardware interrupt in protected mode: it pushes
// the 32-bit interrupt frame (EFLAGS, CS, EIP) and vectors to the handler at
// cs:eip, clearing IF/TF — the counterpart to Interrupt for a flat 32-bit client
// whose handler ends in IRETD (which pops that same frame, see 0xCF). The vector
// is not in any IVT here: a DPMI host records its clients' PM handlers (INT 31h
// 0205h), so the caller supplies the recorded cs:eip. Masking matches Interrupt.
func (c *CPU) InterruptPM(cs uint16, eip uint32) bool {
	if !c.IF || c.Halted || c.ssShadow {
		return false
	}
	c.push(4, uint32(c.EFlags()))
	c.IF, c.TF = false, false
	c.push(4, uint32(c.Seg[CS]))
	c.push(4, c.IP)
	c.loadSeg(CS, cs)
	c.IP = eip
	return true
}
