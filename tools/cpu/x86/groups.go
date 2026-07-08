package x86

// grp1 decodes the immediate ALU group (opcodes 0x80–0x83): the reg field of
// the ModR/M byte selects ADD/OR/ADC/SBB/AND/SUB/XOR/CMP and the opcode selects
// the operand width and immediate form.
//
//	0x80  r/m8,  imm8
//	0x81  r/m16, imm16/imm32 (operand-size)
//	0x82  r/m8,  imm8   (an alias of 0x80)
//	0x83  r/m16, imm8   (sign-extended to the operand size)
func (d *dec) grp1(op byte) Inst {
	reg, rm := d.modrm()
	mnem := alu[reg]
	switch op {
	case 0x80, 0x82:
		return op2(mnem, rm.atHint(8), hexImm(uint32(d.imm8()), 1))
	case 0x81:
		v, w := d.immOsz()
		return op2(mnem, rm.atHint(d.opsize), hexImm(v, w))
	default: // 0x83: imm8 sign-extended
		imm := uint32(int32(int8(d.next())))
		return op2(mnem, rm.atHint(d.opsize), hexImm(imm, imWidth(d.opsize)))
	}
}

// grp2 decodes the rotate/shift group (0xC0/0xC1 imm8, 0xD0/0xD1 by 1,
// 0xD2/0xD3 by CL). count is "imm", "1" or "CL".
func (d *dec) grp2(size int, count string) Inst {
	reg, rm := d.modrm()
	mnem := shf[reg]
	amt := count
	if count == "imm" {
		amt = hexImm(uint32(d.imm8()), 1)
	}
	return op2(mnem, rm.atHint(size), amt)
}

// grp3 decodes 0xF6 (byte) / 0xF7 (operand-size): TEST/NOT/NEG/MUL/IMUL/DIV/IDIV.
func (d *dec) grp3(size int) Inst {
	reg, rm := d.modrm()
	switch reg {
	case 0, 1: // TEST r/m, imm
		if size == 8 {
			return op2("TEST", rm.atHint(8), hexImm(uint32(d.imm8()), 1))
		}
		v, w := d.immOsz()
		return op2("TEST", rm.atHint(size), hexImm(v, w))
	case 2:
		return op1("NOT", rm.atHint(size))
	case 3:
		return op1("NEG", rm.atHint(size))
	case 4:
		return op1("MUL", rm.atHint(size))
	case 5:
		return op1("IMUL", rm.atHint(size))
	case 6:
		return op1("DIV", rm.atHint(size))
	default:
		return op1("IDIV", rm.atHint(size))
	}
}

// grp4 decodes 0xFE: INC/DEC r/m8 (other reg values are undefined).
func (d *dec) grp4() Inst {
	reg, rm := d.modrm()
	switch reg {
	case 0:
		return op1("INC", rm.atHint(8))
	case 1:
		return op1("DEC", rm.atHint(8))
	}
	d.bad = true
	return Inst{}
}

// grp5 decodes 0xFF: INC/DEC/CALL/CALLF/JMP/JMPF/PUSH on an operand-size r/m.
// The indirect CALL/JMP forms have no statically-known target: an indirect JMP
// ends the traced path (FlowIndJump), while an indirect CALL keeps the
// fall-through (FlowCall with HasTarget=false).
func (d *dec) grp5() Inst {
	reg, rm := d.modrm()
	switch reg {
	case 0:
		return op1("INC", rm.atHint(d.opsize))
	case 1:
		return op1("DEC", rm.atHint(d.opsize))
	case 2:
		return Inst{Mnem: "CALL", Text: "CALL " + rm.at(d.opsize), Flow: FlowCall}
	case 3:
		return Inst{Mnem: "CALLF", Text: "CALLF " + rm.atHint(d.opsize), Flow: FlowCall}
	case 4:
		return Inst{Mnem: "JMP", Text: "JMP " + rm.at(d.opsize), Flow: FlowIndJump}
	case 5:
		return Inst{Mnem: "JMPF", Text: "JMPF " + rm.atHint(d.opsize), Flow: FlowIndJump}
	case 6:
		return op1("PUSH", rm.atHint(d.opsize))
	}
	d.bad = true
	return Inst{}
}
