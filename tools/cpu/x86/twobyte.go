package x86

import "fmt"

// twoByte decodes an opcode on the 0x0F page (286/386 system and bit
// instructions, the near Jcc/SETcc forms, MOVZX/MOVSX, shifts, etc.).
func (d *dec) twoByte(op byte) Inst {
	switch {
	case op >= 0x80 && op <= 0x8F: // Jcc rel16/32 (near)
		mnem := "J" + ccName[op&0x0F]
		t := d.jrel(imWidth(d.opsize))
		return Inst{Mnem: mnem, Text: fmt.Sprintf("%s $%08X", mnem, t), Flow: FlowBranch, Target: t, HasTarget: true}
	case op >= 0x90 && op <= 0x9F: // SETcc r/m8
		_, rm := d.modrm()
		mnem := "SET" + ccName[op&0x0F]
		return op1(mnem, rm.atHint(8))
	case op >= 0x40 && op <= 0x4F: // CMOVcc r, r/m (P6)
		reg, rm := d.modrm()
		mnem := "CMOV" + ccName[op&0x0F]
		return op2(mnem, gpr(reg, d.opsize), rm.at(d.opsize))
	case op >= 0xC8 && op <= 0xCF: // BSWAP r32
		return op1("BSWAP", reg32[op&7])
	}

	switch op {
	case 0x00:
		return d.grp6()
	case 0x01:
		return d.grp7()
	case 0x02:
		reg, rm := d.modrm()
		return op2("LAR", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0x03:
		reg, rm := d.modrm()
		return op2("LSL", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0x06:
		return mk("CLTS", "CLTS")
	case 0x08:
		return mk("INVD", "INVD")
	case 0x09:
		return mk("WBINVD", "WBINVD")
	case 0x0B:
		return Inst{Mnem: "UD2", Text: "UD2", Flow: FlowStop}

	case 0x20:
		reg, rm := d.modrm()
		return op2("MOV", rm.at(32), fmt.Sprintf("CR%d", reg))
	case 0x21:
		reg, rm := d.modrm()
		return op2("MOV", rm.at(32), fmt.Sprintf("DR%d", reg))
	case 0x22:
		reg, rm := d.modrm()
		return op2("MOV", fmt.Sprintf("CR%d", reg), rm.at(32))
	case 0x23:
		reg, rm := d.modrm()
		return op2("MOV", fmt.Sprintf("DR%d", reg), rm.at(32))
	case 0x31:
		return mk("RDTSC", "RDTSC")

	case 0xA0:
		return op1("PUSH", "FS")
	case 0xA1:
		return op1("POP", "FS")
	case 0xA2:
		return mk("CPUID", "CPUID")
	case 0xA3:
		reg, rm := d.modrm()
		return op2("BT", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0xA4:
		reg, rm := d.modrm()
		return mk("SHLD", fmt.Sprintf("SHLD %s, %s, %s", rm.at(d.opsize), gpr(reg, d.opsize), hexImm(uint32(d.imm8()), 1)))
	case 0xA5:
		reg, rm := d.modrm()
		return mk("SHLD", fmt.Sprintf("SHLD %s, %s, CL", rm.at(d.opsize), gpr(reg, d.opsize)))
	case 0xA8:
		return op1("PUSH", "GS")
	case 0xA9:
		return op1("POP", "GS")
	case 0xAB:
		reg, rm := d.modrm()
		return op2("BTS", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0xAC:
		reg, rm := d.modrm()
		return mk("SHRD", fmt.Sprintf("SHRD %s, %s, %s", rm.at(d.opsize), gpr(reg, d.opsize), hexImm(uint32(d.imm8()), 1)))
	case 0xAD:
		reg, rm := d.modrm()
		return mk("SHRD", fmt.Sprintf("SHRD %s, %s, CL", rm.at(d.opsize), gpr(reg, d.opsize)))
	case 0xAF:
		reg, rm := d.modrm()
		return op2("IMUL", gpr(reg, d.opsize), rm.at(d.opsize))

	case 0xB0:
		reg, rm := d.modrm()
		return op2("CMPXCHG", rm.at(8), reg8[reg])
	case 0xB1:
		reg, rm := d.modrm()
		return op2("CMPXCHG", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0xB2:
		reg, rm := d.modrm()
		return op2("LSS", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xB3:
		reg, rm := d.modrm()
		return op2("BTR", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0xB4:
		reg, rm := d.modrm()
		return op2("LFS", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xB5:
		reg, rm := d.modrm()
		return op2("LGS", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xB6:
		reg, rm := d.modrm()
		return op2("MOVZX", gpr(reg, d.opsize), rm.at(8))
	case 0xB7:
		reg, rm := d.modrm()
		return op2("MOVZX", gpr(reg, d.opsize), rm.at(16))
	case 0xBA:
		return d.grp8()
	case 0xBB:
		reg, rm := d.modrm()
		return op2("BTC", rm.at(d.opsize), gpr(reg, d.opsize))
	case 0xBC:
		reg, rm := d.modrm()
		return op2("BSF", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xBD:
		reg, rm := d.modrm()
		return op2("BSR", gpr(reg, d.opsize), rm.at(d.opsize))
	case 0xBE:
		reg, rm := d.modrm()
		return op2("MOVSX", gpr(reg, d.opsize), rm.at(8))
	case 0xBF:
		reg, rm := d.modrm()
		return op2("MOVSX", gpr(reg, d.opsize), rm.at(16))

	case 0xC0:
		reg, rm := d.modrm()
		return op2("XADD", rm.at(8), reg8[reg])
	case 0xC1:
		reg, rm := d.modrm()
		return op2("XADD", rm.at(d.opsize), gpr(reg, d.opsize))
	}

	// Unrecognised 0x0F opcode: emit a two-byte data stop so the caller stays
	// aligned rather than mis-decoding the following byte.
	return Inst{Mnem: ".byte", Text: fmt.Sprintf(".byte $0F,$%02X", op), Flow: FlowStop}
}

// grp6 decodes 0x0F 0x00: SLDT/STR/LLDT/LTR/VERR/VERW r/m16.
func (d *dec) grp6() Inst {
	reg, rm := d.modrm()
	names := [6]string{"SLDT", "STR", "LLDT", "LTR", "VERR", "VERW"}
	if reg < 6 {
		return op1(names[reg], rm.atHint(16))
	}
	d.bad = true
	return Inst{}
}

// grp7 decodes 0x0F 0x01: SGDT/SIDT/LGDT/LIDT/SMSW/LMSW/INVLPG.
func (d *dec) grp7() Inst {
	reg, rm := d.modrm()
	switch reg {
	case 0:
		return op1("SGDT", rm.at(d.opsize))
	case 1:
		return op1("SIDT", rm.at(d.opsize))
	case 2:
		return op1("LGDT", rm.at(d.opsize))
	case 3:
		return op1("LIDT", rm.at(d.opsize))
	case 4:
		return op1("SMSW", rm.atHint(16))
	case 6:
		return op1("LMSW", rm.atHint(16))
	case 7:
		return op1("INVLPG", rm.at(d.opsize))
	}
	d.bad = true
	return Inst{}
}

// grp8 decodes 0x0F 0xBA: BT/BTS/BTR/BTC r/m, imm8 (reg 4..7).
func (d *dec) grp8() Inst {
	reg, rm := d.modrm()
	names := map[byte]string{4: "BT", 5: "BTS", 6: "BTR", 7: "BTC"}
	if n, ok := names[reg]; ok {
		return op2(n, rm.atHint(d.opsize), hexImm(uint32(d.imm8()), 1))
	}
	d.bad = true
	return Inst{}
}
