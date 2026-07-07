package mips

import (
	"encoding/binary"
	"fmt"
)

// regName is the MIPS o32 register mnemonic for each of the 32 GPRs.
var regName = [32]string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

func reg(i uint32) string { return "$" + regName[i&31] }

// gteCmd maps a COP2 command's function field (low 6 bits) to a GTE mnemonic,
// for readable disassembly of the geometry engine's command stream.
var gteCmd = map[uint32]string{
	0x01: "rtps", 0x06: "nclip", 0x0C: "op", 0x10: "dpcs",
	0x11: "intpl", 0x12: "mvmva", 0x13: "ncds", 0x14: "cdp",
	0x16: "ncdt", 0x1B: "nccs", 0x1C: "cc", 0x1E: "ncs",
	0x20: "nct", 0x28: "sqr", 0x29: "dcpl", 0x2A: "dpct",
	0x2D: "avsz3", 0x2E: "avsz4", 0x30: "rtpt", 0x3D: "gpf",
	0x3E: "gpl", 0x3F: "ncct",
}

// Decode decodes the MIPS instruction at code[0:], whose CPU address is addr
// (used to compute branch and jump targets). A short slice yields Len 0.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 4 {
		return Inst{Addr: addr, Len: 0, Mnem: ".word", Text: ".word <truncated>", Flow: FlowStop}
	}
	w := binary.LittleEndian.Uint32(code)
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq}

	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	funct := w & 63
	imm := w & 0xFFFF
	simm := int32(int16(imm))          // sign-extended immediate
	branchT := addr + 4 + uint32(simm)*4 // PC-relative branch target
	jumpT := (addr+4)&0xF0000000 | (w&0x03FFFFFF)<<2

	set := func(mnem, text string) { in.Mnem, in.Text = mnem, text }

	switch op {
	case 0x00: // SPECIAL — dispatch on funct
		return decodeSpecial(in, w, rs, rt, rd, shamt, funct)

	case 0x01: // REGIMM — dispatch on rt
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		switch rt {
		case 0x00:
			set("bltz", fmt.Sprintf("bltz %s, $%08X", reg(rs), branchT))
		case 0x01:
			set("bgez", fmt.Sprintf("bgez %s, $%08X", reg(rs), branchT))
		case 0x10:
			set("bltzal", fmt.Sprintf("bltzal %s, $%08X", reg(rs), branchT))
		case 0x11:
			set("bgezal", fmt.Sprintf("bgezal %s, $%08X", reg(rs), branchT))
		default:
			return word(in, w)
		}

	case 0x02: // j
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowJump, jumpT, true, true
		set("j", fmt.Sprintf("j $%08X", jumpT))
	case 0x03: // jal
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowCall, jumpT, true, true
		set("jal", fmt.Sprintf("jal $%08X", jumpT))

	case 0x04: // beq
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		if rs == 0 && rt == 0 {
			set("b", fmt.Sprintf("b $%08X", branchT)) // beq $0,$0 == unconditional b
			in.Flow = FlowJump
		} else {
			set("beq", fmt.Sprintf("beq %s, %s, $%08X", reg(rs), reg(rt), branchT))
		}
	case 0x05: // bne
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		set("bne", fmt.Sprintf("bne %s, %s, $%08X", reg(rs), reg(rt), branchT))
	case 0x06: // blez
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		set("blez", fmt.Sprintf("blez %s, $%08X", reg(rs), branchT))
	case 0x07: // bgtz
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		set("bgtz", fmt.Sprintf("bgtz %s, $%08X", reg(rs), branchT))

	case 0x08: // addi
		set("addi", fmt.Sprintf("addi %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x09: // addiu
		set("addiu", fmt.Sprintf("addiu %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0A: // slti
		set("slti", fmt.Sprintf("slti %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0B: // sltiu
		set("sltiu", fmt.Sprintf("sltiu %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0C: // andi
		set("andi", fmt.Sprintf("andi %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0D: // ori
		set("ori", fmt.Sprintf("ori %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0E: // xori
		set("xori", fmt.Sprintf("xori %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0F: // lui
		set("lui", fmt.Sprintf("lui %s, 0x%X", reg(rt), imm))

	case 0x10: // COP0
		return decodeCop0(in, w, rs, rt, rd)
	case 0x12: // COP2 (GTE)
		return decodeCop2(in, w, rs, rt, rd)

	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26: // loads
		m := [...]string{0x20: "lb", 0x21: "lh", 0x22: "lwl", 0x23: "lw", 0x24: "lbu", 0x25: "lhu", 0x26: "lwr"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))
	case 0x28, 0x29, 0x2A, 0x2B, 0x2E: // stores
		m := map[uint32]string{0x28: "sb", 0x29: "sh", 0x2A: "swl", 0x2B: "sw", 0x2E: "swr"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))
	case 0x32: // lwc2 — load word to GTE data register
		set("lwc2", fmt.Sprintf("lwc2 $%d, %d(%s)", rt, simm, reg(rs)))
	case 0x3A: // swc2 — store GTE data register
		set("swc2", fmt.Sprintf("swc2 $%d, %d(%s)", rt, simm, reg(rs)))

	default:
		return word(in, w)
	}
	return in
}

// decodeSpecial handles op == 0 (register-format arithmetic, shifts, jr/jalr,
// mult/div, syscall/break, HI/LO moves).
func decodeSpecial(in Inst, w, rs, rt, rd, shamt, funct uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	switch funct {
	case 0x00:
		if w == 0 {
			return set("nop", "nop")
		}
		return set("sll", fmt.Sprintf("sll %s, %s, %d", reg(rd), reg(rt), shamt))
	case 0x02:
		return set("srl", fmt.Sprintf("srl %s, %s, %d", reg(rd), reg(rt), shamt))
	case 0x03:
		return set("sra", fmt.Sprintf("sra %s, %s, %d", reg(rd), reg(rt), shamt))
	case 0x04:
		return set("sllv", fmt.Sprintf("sllv %s, %s, %s", reg(rd), reg(rt), reg(rs)))
	case 0x06:
		return set("srlv", fmt.Sprintf("srlv %s, %s, %s", reg(rd), reg(rt), reg(rs)))
	case 0x07:
		return set("srav", fmt.Sprintf("srav %s, %s, %s", reg(rd), reg(rt), reg(rs)))
	case 0x08: // jr
		in.HasDelay = true
		if rs == 31 {
			in.Flow = FlowReturn
		} else {
			in.Flow = FlowIndJump
		}
		return set("jr", fmt.Sprintf("jr %s", reg(rs)))
	case 0x09: // jalr
		in.Flow, in.HasDelay = FlowIndCall, true
		if rd == 31 {
			return set("jalr", fmt.Sprintf("jalr %s", reg(rs)))
		}
		return set("jalr", fmt.Sprintf("jalr %s, %s", reg(rd), reg(rs)))
	case 0x0C:
		in.Flow = FlowStop
		return set("syscall", "syscall")
	case 0x0D:
		in.Flow = FlowStop
		return set("break", "break")
	case 0x10:
		return set("mfhi", fmt.Sprintf("mfhi %s", reg(rd)))
	case 0x11:
		return set("mthi", fmt.Sprintf("mthi %s", reg(rs)))
	case 0x12:
		return set("mflo", fmt.Sprintf("mflo %s", reg(rd)))
	case 0x13:
		return set("mtlo", fmt.Sprintf("mtlo %s", reg(rs)))
	case 0x18:
		return set("mult", fmt.Sprintf("mult %s, %s", reg(rs), reg(rt)))
	case 0x19:
		return set("multu", fmt.Sprintf("multu %s, %s", reg(rs), reg(rt)))
	case 0x1A:
		return set("div", fmt.Sprintf("div %s, %s", reg(rs), reg(rt)))
	case 0x1B:
		return set("divu", fmt.Sprintf("divu %s, %s", reg(rs), reg(rt)))
	case 0x20:
		return set("add", fmt.Sprintf("add %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x21:
		return set("addu", fmt.Sprintf("addu %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x22:
		return set("sub", fmt.Sprintf("sub %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x23:
		return set("subu", fmt.Sprintf("subu %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x24:
		return set("and", fmt.Sprintf("and %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x25:
		if rt == 0 {
			return set("move", fmt.Sprintf("move %s, %s", reg(rd), reg(rs))) // or $rd,$rs,$zero
		}
		return set("or", fmt.Sprintf("or %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x26:
		return set("xor", fmt.Sprintf("xor %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x27:
		return set("nor", fmt.Sprintf("nor %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x2A:
		return set("slt", fmt.Sprintf("slt %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x2B:
		return set("sltu", fmt.Sprintf("sltu %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	}
	return word(in, w)
}

// decodeCop0 handles mfc0/mtc0/cfc0/ctc0 and rfe.
func decodeCop0(in Inst, w, rs, rt, rd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	if w&(1<<25) != 0 { // COP0 command: only rfe (funct 0x10) is used
		if w&0x3F == 0x10 {
			return set("rfe", "rfe")
		}
		return word(in, w)
	}
	switch rs {
	case 0x00:
		return set("mfc0", fmt.Sprintf("mfc0 %s, $%d", reg(rt), rd))
	case 0x04:
		return set("mtc0", fmt.Sprintf("mtc0 %s, $%d", reg(rt), rd))
	}
	return word(in, w)
}

// decodeCop2 handles the GTE register moves and command stream.
func decodeCop2(in Inst, w, rs, rt, rd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	if w&(1<<25) != 0 { // GTE command
		cmd := w & 0x01FFFFFF
		name := gteCmd[w&0x3F]
		if name == "" {
			return set("cop2", fmt.Sprintf("cop2 0x%07X", cmd))
		}
		return set(name, fmt.Sprintf("%s\t; cop2 0x%07X", name, cmd))
	}
	switch rs {
	case 0x00:
		return set("mfc2", fmt.Sprintf("mfc2 %s, $%d", reg(rt), rd))
	case 0x02:
		return set("cfc2", fmt.Sprintf("cfc2 %s, $%d", reg(rt), rd))
	case 0x04:
		return set("mtc2", fmt.Sprintf("mtc2 %s, $%d", reg(rt), rd))
	case 0x06:
		return set("ctc2", fmt.Sprintf("ctc2 %s, $%d", reg(rt), rd))
	}
	return word(in, w)
}

// word marks the instruction as an unmodelled data word that stops a trace.
func word(in Inst, w uint32) Inst {
	in.Mnem, in.Text, in.Flow = ".word", fmt.Sprintf(".word 0x%08X", w), FlowStop
	in.HasTarget, in.HasDelay = false, false
	return in
}
