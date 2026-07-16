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

	// The MMX / SSE-integer group (mmxint.go executes it): named so a census of a
	// codec's inner loops reads as instructions, not `.byte` noise. Register names
	// stay generic ("mmN") — the 0x66 xmm forms share the opcodes.
	if mnem, ok := mmxIntName[op]; ok {
		switch op {
		case 0x70, 0xC4, 0xC5, 0xC6: // shuffle/insert/extract carry an imm8
			reg, rm := d.modrm()
			return Inst{Mnem: mnem, Text: fmt.Sprintf("%s mm%d, %s, $%02X", mnem, reg, rm.atHint(32), d.imm8())}
		case 0x71, 0x72, 0x73: // shift-by-imm groups: reg field selects the op
			reg, rm := d.modrm()
			grp := map[byte]string{2: "PSRL", 3: "PSRLDQ", 4: "PSRA", 6: "PSLL", 7: "PSLLDQ"}
			sz := map[byte]string{0x71: "W", 0x72: "D", 0x73: "Q"}
			if g, ok := grp[reg]; ok {
				n := g + sz[op]
				if reg == 3 || reg == 7 {
					n = g
				}
				return Inst{Mnem: n, Text: fmt.Sprintf("%s %s, $%02X", n, rm.atHint(32), d.imm8())}
			}
		default:
			reg, rm := d.modrm()
			return Inst{Mnem: mnem, Text: fmt.Sprintf("%s mm%d, %s", mnem, reg, rm.atHint(32))}
		}
	}

	// Unrecognised 0x0F opcode: emit a two-byte data stop so the caller stays
	// aligned rather than mis-decoding the following byte.
	return Inst{Mnem: ".byte", Text: fmt.Sprintf(".byte $0F,$%02X", op), Flow: FlowStop}
}

// mmxIntName names the packed-integer opcodes (the no-prefix mm forms; the 0x66 xmm
// variants share the bytes and read fine under the same names).
var mmxIntName = map[byte]string{
	0x2A: "CVTPI2PS", 0x2C: "CVTTPS2PI", 0x2D: "CVTPS2PI",
	0x6E: "MOVD", 0x7E: "MOVD", 0x6F: "MOVQ", 0x7F: "MOVQ",
	// The float-SSE page (executed in sse.go), named generically by its
	// no-prefix form — enough for censuses and call-site reading to stay in sync.
	0x10: "MOVUPS", 0x11: "MOVUPS", 0x12: "MOVLPS", 0x13: "MOVLPS",
	0x14: "UNPCKLPS", 0x15: "UNPCKHPS", 0x16: "MOVHPS", 0x17: "MOVHPS",
	0x28: "MOVAPS", 0x29: "MOVAPS", 0x2B: "MOVNTPS", 0x2E: "UCOMISS", 0x2F: "COMISS",
	0x51: "SQRTPS", 0x52: "RSQRTPS", 0x53: "RCPPS",
	0x54: "ANDPS", 0x55: "ANDNPS", 0x56: "ORPS", 0x57: "XORPS",
	0x58: "ADDPS", 0x59: "MULPS", 0x5A: "CVTPS2PD", 0x5B: "CVTDQ2PS",
	0x5C: "SUBPS", 0x5D: "MINPS", 0x5E: "DIVPS", 0x5F: "MAXPS",
	0xC6: "SHUFPS", 0xD6: "MOVQ",
	0x60: "PUNPCKLBW", 0x61: "PUNPCKLWD", 0x62: "PUNPCKLDQ", 0x63: "PACKSSWB",
	0x64: "PCMPGTB", 0x65: "PCMPGTW", 0x66: "PCMPGTD", 0x67: "PACKUSWB",
	0x68: "PUNPCKHBW", 0x69: "PUNPCKHWD", 0x6A: "PUNPCKHDQ", 0x6B: "PACKSSDW",
	0x6C: "PUNPCKLQDQ", 0x6D: "PUNPCKHQDQ",
	0x70: "PSHUFW", 0x71: "PSHIFTW", 0x72: "PSHIFTD", 0x73: "PSHIFTQ",
	0x74: "PCMPEQB", 0x75: "PCMPEQW", 0x76: "PCMPEQD",
	0xC4: "PINSRW", 0xC5: "PEXTRW",
	0xD1: "PSRLW", 0xD2: "PSRLD", 0xD3: "PSRLQ", 0xD4: "PADDQ", 0xD5: "PMULLW",
	0xD7: "PMOVMSKB", 0xD8: "PSUBUSB", 0xD9: "PSUBUSW", 0xDA: "PMINUB", 0xDB: "PAND",
	0xDC: "PADDUSB", 0xDD: "PADDUSW", 0xDE: "PMAXUB", 0xDF: "PANDN",
	0xE0: "PAVGB", 0xE1: "PSRAW", 0xE2: "PSRAD", 0xE3: "PAVGW", 0xE4: "PMULHUW",
	0xE5: "PMULHW", 0xE7: "MOVNTQ", 0xE8: "PSUBSB", 0xE9: "PSUBSW", 0xEA: "PMINSW",
	0xEB: "POR", 0xEC: "PADDSB", 0xED: "PADDSW", 0xEE: "PMAXSW", 0xEF: "PXOR",
	0xF1: "PSLLW", 0xF2: "PSLLD", 0xF3: "PSLLQ", 0xF4: "PMULUDQ", 0xF5: "PMADDWD",
	0xF6: "PSADBW", 0xF8: "PSUBB", 0xF9: "PSUBW", 0xFA: "PSUBD", 0xFB: "PSUBQ",
	0xFC: "PADDB", 0xFD: "PADDW", 0xFE: "PADDD",
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
