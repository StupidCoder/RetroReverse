package x86

import "fmt"

// oneByte decodes a primary (non-0x0F) opcode.
func (d *dec) oneByte(op byte) Inst {
	// The 0x00–0x3F block is the eight basic ALU ops laid out in a regular
	// grid; the six operand columns (op&7 < 6) are the r/m and accumulator
	// forms, and columns 6/7 are the segment push/pop and BCD adjusts.
	if op < 0x40 && op&7 < 6 {
		return d.aluForm(alu[op>>3], op&7)
	}

	switch op {
	// segment PUSH/POP and BCD adjusts interleaved in the ALU block
	case 0x06:
		return op1("PUSH", "ES")
	case 0x07:
		return op1("POP", "ES")
	case 0x0E:
		return op1("PUSH", "CS")
	case 0x16:
		return op1("PUSH", "SS")
	case 0x17:
		return op1("POP", "SS")
	case 0x1E:
		return op1("PUSH", "DS")
	case 0x1F:
		return op1("POP", "DS")
	case 0x27:
		return mk("DAA", "DAA")
	case 0x2F:
		return mk("DAS", "DAS")
	case 0x37:
		return mk("AAA", "AAA")
	case 0x3F:
		return mk("AAS", "AAS")

	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47:
		return op1("INC", gpr(op&7, d.opsize))
	case 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F:
		return op1("DEC", gpr(op&7, d.opsize))
	case 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57:
		return op1("PUSH", gpr(op&7, d.opsize))
	case 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F:
		return op1("POP", gpr(op&7, d.opsize))

	case 0x60:
		return mk("PUSHA", pick(d.opsize, "PUSHA", "PUSHAD"))
	case 0x61:
		return mk("POPA", pick(d.opsize, "POPA", "POPAD"))
	case 0x62:
		reg, rm := d.modrm()
		return op2("BOUND", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0x63:
		reg, rm := d.modrm()
		return op2("ARPL", rm.at(16), gpr(reg, 16))
	case 0x68:
		v, w := d.immOsz()
		return op1("PUSH", hexImm(v, w))
	case 0x69:
		reg, rm := d.modrm()
		v, w := d.immOsz()
		return mk("IMUL", fmt.Sprintf("IMUL %s, %s, %s", gpr(reg, d.opsize), rm.at(d.opsize), hexImm(v, w)))
	case 0x6A:
		return op1("PUSH", hexImm(uint32(int32(int8(d.next()))), imWidth(d.opsize)))
	case 0x6B:
		reg, rm := d.modrm()
		imm := uint32(int32(int8(d.next())))
		return mk("IMUL", fmt.Sprintf("IMUL %s, %s, %s", gpr(reg, d.opsize), rm.at(d.opsize), hexImm(imm, imWidth(d.opsize))))
	case 0x6C:
		return mk("INSB", d.rep+"INSB")
	case 0x6D:
		return mk("INS", d.rep+pick(d.opsize, "INSW", "INSD"))
	case 0x6E:
		return mk("OUTSB", d.rep+"OUTSB")
	case 0x6F:
		return mk("OUTS", d.rep+pick(d.opsize, "OUTSW", "OUTSD"))

	case 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77,
		0x78, 0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x7F:
		mnem := "J" + ccName[op&0x0F]
		t := d.jrel(1)
		return Inst{Mnem: mnem, Text: fmt.Sprintf("%s $%08X", mnem, t), Flow: FlowBranch, Target: t, HasTarget: true}

	case 0x80, 0x81, 0x82, 0x83:
		return d.grp1(op)

	case 0x84:
		reg, rm := d.modrm()
		return op2("TEST", rm.at(8), reg8[reg])
	case 0x85:
		reg, rm := d.modrm()
		return op2("TEST", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0x86:
		reg, rm := d.modrm()
		return op2("XCHG", rm.at(8), reg8[reg])
	case 0x87:
		reg, rm := d.modrm()
		return op2("XCHG", rm.at(d.opsize), gpr(reg, d.opsize))

	case 0x88:
		reg, rm := d.modrm()
		return op2("MOV", rm.at(8), reg8[reg])
	case 0x89:
		reg, rm := d.modrm()
		return op2("MOV", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0x8A:
		reg, rm := d.modrm()
		return op2("MOV", reg8[reg], rm.at(8))
	case 0x8B:
		reg, rm := d.modrm()
		return op2("MOV", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0x8C:
		reg, rm := d.modrm()
		return op2("MOV", rm.at(16), sreg[reg])
	case 0x8D:
		reg, rm := d.modrm()
		return op2("LEA", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0x8E:
		reg, rm := d.modrm()
		return op2("MOV", sreg[reg], rm.at(16))
	case 0x8F:
		_, rm := d.modrm()
		return op1("POP", rm.atHint(d.opsize))

	case 0x90:
		if d.rep == "REP " {
			return mk("PAUSE", "PAUSE")
		}
		return mk("NOP", "NOP")
	case 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97:
		return op2("XCHG", gpr(0, d.opsize), gpr(op&7, d.opsize))

	case 0x98:
		return mk("CBW", pick(d.opsize, "CBW", "CWDE"))
	case 0x99:
		return mk("CWD", pick(d.opsize, "CWD", "CDQ"))
	case 0x9A:
		return d.farPtr("CALLF", FlowCall)
	case 0x9B:
		return mk("FWAIT", "FWAIT")
	case 0x9C:
		return mk("PUSHF", pick(d.opsize, "PUSHF", "PUSHFD"))
	case 0x9D:
		return mk("POPF", pick(d.opsize, "POPF", "POPFD"))
	case 0x9E:
		return mk("SAHF", "SAHF")
	case 0x9F:
		return mk("LAHF", "LAHF")

	case 0xA0:
		return op2("MOV", "AL", d.moffs())
	case 0xA1:
		return op2("MOV", gpr(0, d.opsize), d.moffs())
	case 0xA2:
		return op2("MOV", d.moffs(), "AL")
	case 0xA3:
		return op2("MOV", d.moffs(), gpr(0, d.opsize))
	case 0xA4:
		return mk("MOVSB", d.rep+"MOVSB")
	case 0xA5:
		return mk("MOVS", d.rep+pick(d.opsize, "MOVSW", "MOVSD"))
	case 0xA6:
		return mk("CMPSB", d.rep+"CMPSB")
	case 0xA7:
		return mk("CMPS", d.rep+pick(d.opsize, "CMPSW", "CMPSD"))
	case 0xA8:
		return op2("TEST", "AL", hexImm(uint32(d.imm8()), 1))
	case 0xA9:
		v, w := d.immOsz()
		return op2("TEST", gpr(0, d.opsize), hexImm(v, w))
	case 0xAA:
		return mk("STOSB", d.rep+"STOSB")
	case 0xAB:
		return mk("STOS", d.rep+pick(d.opsize, "STOSW", "STOSD"))
	case 0xAC:
		return mk("LODSB", d.rep+"LODSB")
	case 0xAD:
		return mk("LODS", d.rep+pick(d.opsize, "LODSW", "LODSD"))
	case 0xAE:
		return mk("SCASB", d.rep+"SCASB")
	case 0xAF:
		return mk("SCAS", d.rep+pick(d.opsize, "SCASW", "SCASD"))

	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7:
		return op2("MOV", reg8[op&7], hexImm(uint32(d.imm8()), 1))
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
		v, w := d.immOsz()
		return op2("MOV", gpr(op&7, d.opsize), hexImm(v, w))

	case 0xC0:
		return d.grp2(8, "imm")
	case 0xC1:
		return d.grp2(d.opsize, "imm")
	case 0xC2:
		return Inst{Mnem: "RET", Text: "RET " + hexImm(uint32(d.imm16()), 2), Flow: FlowReturn}
	case 0xC3:
		return Inst{Mnem: "RET", Text: "RET", Flow: FlowReturn}
	case 0xC4:
		reg, rm := d.modrm()
		return op2("LES", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xC5:
		reg, rm := d.modrm()
		return op2("LDS", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xC6:
		_, rm := d.modrm()
		return op2("MOV", rm.atHint(8), hexImm(uint32(d.imm8()), 1))
	case 0xC7:
		_, rm := d.modrm()
		v, w := d.immOsz()
		return op2("MOV", rm.atHint(d.opsize), hexImm(v, w))
	case 0xC8:
		sz := d.imm16()
		lvl := d.imm8()
		return op2("ENTER", hexImm(uint32(sz), 2), hexImm(uint32(lvl), 1))
	case 0xC9:
		return mk("LEAVE", "LEAVE")
	case 0xCA:
		return Inst{Mnem: "RETF", Text: "RETF " + hexImm(uint32(d.imm16()), 2), Flow: FlowReturn}
	case 0xCB:
		return Inst{Mnem: "RETF", Text: "RETF", Flow: FlowReturn}
	case 0xCC:
		return mk("INT3", "INT3")
	case 0xCD:
		return op1("INT", hexImm(uint32(d.imm8()), 1))
	case 0xCE:
		return mk("INTO", "INTO")
	case 0xCF:
		return Inst{Mnem: "IRET", Text: pick(d.opsize, "IRET", "IRETD"), Flow: FlowReturn}

	case 0xD0:
		return d.grp2(8, "1")
	case 0xD1:
		return d.grp2(d.opsize, "1")
	case 0xD2:
		return d.grp2(8, "CL")
	case 0xD3:
		return d.grp2(d.opsize, "CL")
	case 0xD4:
		if b := d.imm8(); b != 0x0A {
			return op1("AAM", hexImm(uint32(b), 1))
		}
		return mk("AAM", "AAM")
	case 0xD5:
		if b := d.imm8(); b != 0x0A {
			return op1("AAD", hexImm(uint32(b), 1))
		}
		return mk("AAD", "AAD")
	case 0xD6:
		return mk("SALC", "SALC")
	case 0xD7:
		return mk("XLAT", "XLAT")
	case 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF:
		return d.fpu(op)

	case 0xE0:
		return d.loop("LOOPNE")
	case 0xE1:
		return d.loop("LOOPE")
	case 0xE2:
		return d.loop("LOOP")
	case 0xE3:
		return d.loop(pick(d.addrsize, "JCXZ", "JECXZ"))
	case 0xE4:
		return op2("IN", "AL", hexImm(uint32(d.imm8()), 1))
	case 0xE5:
		return op2("IN", gpr(0, d.opsize), hexImm(uint32(d.imm8()), 1))
	case 0xE6:
		return op2("OUT", hexImm(uint32(d.imm8()), 1), "AL")
	case 0xE7:
		return op2("OUT", hexImm(uint32(d.imm8()), 1), gpr(0, d.opsize))
	case 0xE8:
		t := d.jrel(imWidth(d.opsize))
		return Inst{Mnem: "CALL", Text: fmt.Sprintf("CALL $%08X", t), Flow: FlowCall, Target: t, HasTarget: true}
	case 0xE9:
		t := d.jrel(imWidth(d.opsize))
		return Inst{Mnem: "JMP", Text: fmt.Sprintf("JMP $%08X", t), Flow: FlowJump, Target: t, HasTarget: true}
	case 0xEA:
		return d.farPtr("JMPF", FlowJump)
	case 0xEB:
		t := d.jrel(1)
		return Inst{Mnem: "JMP", Text: fmt.Sprintf("JMP $%08X", t), Flow: FlowJump, Target: t, HasTarget: true}
	case 0xEC:
		return op2("IN", "AL", "DX")
	case 0xED:
		return op2("IN", gpr(0, d.opsize), "DX")
	case 0xEE:
		return op2("OUT", "DX", "AL")
	case 0xEF:
		return op2("OUT", "DX", gpr(0, d.opsize))

	case 0xF1:
		return mk("INT1", "INT1")
	case 0xF4:
		return Inst{Mnem: "HLT", Text: "HLT", Flow: FlowStop}
	case 0xF5:
		return mk("CMC", "CMC")
	case 0xF6:
		return d.grp3(8)
	case 0xF7:
		return d.grp3(d.opsize)
	case 0xF8:
		return mk("CLC", "CLC")
	case 0xF9:
		return mk("STC", "STC")
	case 0xFA:
		return mk("CLI", "CLI")
	case 0xFB:
		return mk("STI", "STI")
	case 0xFC:
		return mk("CLD", "CLD")
	case 0xFD:
		return mk("STD", "STD")
	case 0xFE:
		return d.grp4()
	case 0xFF:
		return d.grp5()
	}

	d.bad = true
	return Inst{}
}

// aluForm decodes one operand column (z = 0..5) of a basic ALU op.
func (d *dec) aluForm(mnem string, z byte) Inst {
	switch z {
	case 0:
		reg, rm := d.modrm()
		return op2(mnem, rm.at(8), reg8[reg])
	case 1:
		reg, rm := d.modrm()
		return op2(mnem, rm.at(d.opsize), gpr(reg, d.opsize))
	case 2:
		reg, rm := d.modrm()
		return op2(mnem, reg8[reg], rm.at(8))
	case 3:
		reg, rm := d.modrm()
		return op2(mnem, gpr(reg, d.opsize), rm.at(d.opsize))
	case 4:
		return op2(mnem, "AL", hexImm(uint32(d.imm8()), 1))
	default: // 5
		v, w := d.immOsz()
		return op2(mnem, gpr(0, d.opsize), hexImm(v, w))
	}
}

// loop decodes a LOOP/LOOPcc/JCXZ rel8 branch.
func (d *dec) loop(mnem string) Inst {
	t := d.jrel(1)
	return Inst{Mnem: mnem, Text: fmt.Sprintf("%s $%08X", mnem, t), Flow: FlowBranch, Target: t, HasTarget: true}
}

// farPtr decodes a direct far pointer operand (ptr16:16 / ptr16:32) for the
// CALLF/JMPF forms and folds seg:off to a linear Target.
func (d *dec) farPtr(mnem string, flow Flow) Inst {
	off, w := d.immOsz()
	seg := d.imm16()
	t := uint32(seg)<<4 + off
	return Inst{Mnem: mnem, Text: fmt.Sprintf("%s $%04X:%s", mnem, seg, hexImm(off, w)),
		Flow: flow, Target: t, HasTarget: true}
}

// moffs formats a MOV moffs direct-memory operand (address-size-wide offset).
func (d *dec) moffs() string {
	if d.addrsize == 32 {
		return fmt.Sprintf("[%s%s]", d.seg, hexImm(d.imm32(), 4))
	}
	return fmt.Sprintf("[%s%s]", d.seg, hexImm(uint32(d.imm16()), 2))
}

// pick returns a16 for 16-bit operand/address size, else a32.
func pick(size int, a16, a32 string) string {
	if size == 32 {
		return a32
	}
	return a16
}

// imWidth is the display byte-width for an operand-size immediate (2 or 4).
func imWidth(size int) int {
	if size == 32 {
		return 4
	}
	return 2
}
