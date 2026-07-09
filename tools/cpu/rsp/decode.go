package rsp

import (
	"encoding/binary"
	"fmt"
)

var regName = [32]string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

func reg(i uint32) string  { return "$" + regName[i&31] }
func vreg(i uint32) string { return fmt.Sprintf("$v%d", i&31) }

// cop0Name labels the RSP's COP0 registers, which are the SP and DP register
// blocks rather than a system-control unit.
var cop0Name = [16]string{
	"SP_MEM_ADDR", "SP_DRAM_ADDR", "SP_RD_LEN", "SP_WR_LEN",
	"SP_STATUS", "SP_DMA_FULL", "SP_DMA_BUSY", "SP_SEMAPHORE",
	"DPC_START", "DPC_END", "DPC_CURRENT", "DPC_STATUS",
	"DPC_CLOCK", "DPC_BUFBUSY", "DPC_PIPEBUSY", "DPC_TMEM",
}

// vuOp names the COP2 vector operations by their function field.
var vuOp = map[uint32]string{
	0x00: "vmulf", 0x01: "vmulu", 0x02: "vrndp", 0x03: "vmulq",
	0x04: "vmudl", 0x05: "vmudm", 0x06: "vmudn", 0x07: "vmudh",
	0x08: "vmacf", 0x09: "vmacu", 0x0A: "vrndn", 0x0B: "vmacq",
	0x0C: "vmadl", 0x0D: "vmadm", 0x0E: "vmadn", 0x0F: "vmadh",
	0x10: "vadd", 0x11: "vsub", 0x13: "vabs", 0x14: "vaddc", 0x15: "vsubc",
	0x1D: "vsar",
	0x20: "vlt", 0x21: "veq", 0x22: "vne", 0x23: "vge",
	0x24: "vcl", 0x25: "vch", 0x26: "vcr", 0x27: "vmrg",
	0x28: "vand", 0x29: "vnand", 0x2A: "vor", 0x2B: "vnor",
	0x2C: "vxor", 0x2D: "vnxor",
	0x30: "vrcp", 0x31: "vrcpl", 0x32: "vrcph", 0x33: "vmov",
	0x34: "vrsq", 0x35: "vrsql", 0x36: "vrsqh", 0x37: "vnop",
}

// lwc2Op and swc2Op name the element-addressed vector load/store family. The
// opcode occupies the rd field, not the funct field.
var lwc2Op = [32]string{
	0x00: "lbv", 0x01: "lsv", 0x02: "llv", 0x03: "ldv", 0x04: "lqv", 0x05: "lrv",
	0x06: "lpv", 0x07: "luv", 0x08: "lhv", 0x09: "lfv", 0x0A: "lwv", 0x0B: "ltv",
}

var swc2Op = [32]string{
	0x00: "sbv", 0x01: "ssv", 0x02: "slv", 0x03: "sdv", 0x04: "sqv", 0x05: "srv",
	0x06: "spv", 0x07: "suv", 0x08: "shv", 0x09: "sfv", 0x0A: "swv", 0x0B: "stv",
}

// memShift is how far the 7-bit signed offset is scaled for each load/store,
// which is the access's element width in bytes.
var memShift = [32]uint32{
	0x00: 0, 0x01: 1, 0x02: 2, 0x03: 3, 0x04: 4, 0x05: 4,
	0x06: 3, 0x07: 3, 0x08: 4, 0x09: 4, 0x0A: 4, 0x0B: 4,
}

// Decode decodes the instruction at code[0:], whose IMEM address is addr.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 4 {
		return Inst{Addr: addr, Len: 0, Mnem: ".word", Text: ".word <truncated>", Flow: FlowStop}
	}
	return DecodeWord(binary.BigEndian.Uint32(code), addr)
}

// DecodeWord decodes an already-assembled instruction word.
func DecodeWord(w, addr uint32) Inst {
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq}

	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	imm := w & 0xFFFF
	simm := int32(int16(imm))

	// Branch and jump targets wrap inside the 4 KiB of instruction memory.
	branchT := (addr + 4 + uint32(simm)*4) & 0xFFC
	jumpT := (w & 0x03FFFFFF << 2) & 0xFFC

	set := func(mnem, text string) { in.Mnem, in.Text = mnem, text }
	branch := func(m, text string) {
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		set(m, text)
	}

	switch op {
	case 0x00:
		return decodeSpecial(in, w, rs, rt, rd, shamt)
	case 0x01: // REGIMM
		switch rt {
		case 0x00:
			branch("bltz", fmt.Sprintf("bltz %s, $%03X", reg(rs), branchT))
		case 0x01:
			branch("bgez", fmt.Sprintf("bgez %s, $%03X", reg(rs), branchT))
		case 0x10:
			branch("bltzal", fmt.Sprintf("bltzal %s, $%03X", reg(rs), branchT))
			in.Flow = FlowCall
		case 0x11:
			branch("bgezal", fmt.Sprintf("bgezal %s, $%03X", reg(rs), branchT))
			in.Flow = FlowCall
		default:
			return word(in, w)
		}
	case 0x02:
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowJump, jumpT, true, true
		set("j", fmt.Sprintf("j $%03X", jumpT))
	case 0x03:
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowCall, jumpT, true, true
		set("jal", fmt.Sprintf("jal $%03X", jumpT))
	case 0x04:
		if rs == 0 && rt == 0 {
			in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowJump, branchT, true, true
			set("b", fmt.Sprintf("b $%03X", branchT))
		} else {
			branch("beq", fmt.Sprintf("beq %s, %s, $%03X", reg(rs), reg(rt), branchT))
		}
	case 0x05:
		branch("bne", fmt.Sprintf("bne %s, %s, $%03X", reg(rs), reg(rt), branchT))
	case 0x06:
		branch("blez", fmt.Sprintf("blez %s, $%03X", reg(rs), branchT))
	case 0x07:
		branch("bgtz", fmt.Sprintf("bgtz %s, $%03X", reg(rs), branchT))

	// The RSP has no arithmetic exceptions, so addi and addiu are the same
	// instruction; the encodings are kept apart only for readability.
	case 0x08:
		set("addi", fmt.Sprintf("addi %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x09:
		set("addiu", fmt.Sprintf("addiu %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0A:
		set("slti", fmt.Sprintf("slti %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0B:
		set("sltiu", fmt.Sprintf("sltiu %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x0C:
		set("andi", fmt.Sprintf("andi %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0D:
		set("ori", fmt.Sprintf("ori %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0E:
		set("xori", fmt.Sprintf("xori %s, %s, 0x%X", reg(rt), reg(rs), imm))
	case 0x0F:
		set("lui", fmt.Sprintf("lui %s, 0x%X", reg(rt), imm))

	case 0x10: // COP0: the SP and DP registers
		switch rs {
		case 0x00:
			set("mfc0", fmt.Sprintf("mfc0 %s, %s", reg(rt), cop0Reg(rd)))
		case 0x04:
			set("mtc0", fmt.Sprintf("mtc0 %s, %s", reg(rt), cop0Reg(rd)))
		default:
			return word(in, w)
		}
	case 0x12: // COP2: the vector unit
		return decodeCop2(in, w, rs, rt, rd, shamt)

	case 0x20, 0x21, 0x23, 0x24, 0x25: // scalar loads
		m := [...]string{0x20: "lb", 0x21: "lh", 0x23: "lw", 0x24: "lbu", 0x25: "lhu"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))
	case 0x28, 0x29, 0x2B: // scalar stores
		m := [...]string{0x28: "sb", 0x29: "sh", 0x2B: "sw"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))

	case 0x32: // LWC2
		return decodeVecMem(in, w, rs, rt, rd, lwc2Op[:])
	case 0x3A: // SWC2
		return decodeVecMem(in, w, rs, rt, rd, swc2Op[:])

	default:
		return word(in, w)
	}
	return in
}

func cop0Reg(i uint32) string {
	if i < 16 {
		return cop0Name[i]
	}
	return fmt.Sprintf("$%d", i)
}

func decodeSpecial(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	funct := w & 63
	set := func(m, text string) Inst { in.Mnem, in.Text = m, text; return in }
	r3 := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %s", m, reg(rd), reg(rs), reg(rt))) }
	sh := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %d", m, reg(rd), reg(rt), shamt)) }
	shv := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %s", m, reg(rd), reg(rt), reg(rs))) }

	switch funct {
	case 0x00:
		if w == 0 {
			return set("nop", "nop")
		}
		return sh("sll")
	case 0x02:
		return sh("srl")
	case 0x03:
		return sh("sra")
	case 0x04:
		return shv("sllv")
	case 0x06:
		return shv("srlv")
	case 0x07:
		return shv("srav")
	case 0x08:
		in.HasDelay = true
		if rs == 31 {
			in.Flow = FlowReturn
		} else {
			in.Flow = FlowIndJump
		}
		return set("jr", fmt.Sprintf("jr %s", reg(rs)))
	case 0x09:
		in.Flow, in.HasDelay = FlowIndCall, true
		return set("jalr", fmt.Sprintf("jalr %s, %s", reg(rd), reg(rs)))
	case 0x0D:
		in.Flow = FlowStop
		return set("break", "break")
	case 0x20:
		return r3("add")
	case 0x21:
		return r3("addu")
	case 0x22:
		return r3("sub")
	case 0x23:
		return r3("subu")
	case 0x24:
		return r3("and")
	case 0x25:
		if rt == 0 {
			return set("move", fmt.Sprintf("move %s, %s", reg(rd), reg(rs)))
		}
		return r3("or")
	case 0x26:
		return r3("xor")
	case 0x27:
		return r3("nor")
	case 0x2A:
		return r3("slt")
	case 0x2B:
		return r3("sltu")
	}
	return word(in, w)
}

// decodeCop2 handles the vector-unit register moves and the vector ALU.
func decodeCop2(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	set := func(m, text string) Inst { in.Mnem, in.Text = m, text; return in }

	if w&(1<<25) == 0 { // register moves; rd names the vector register
		e := (w >> 7) & 15
		switch rs {
		case 0x00:
			return set("mfc2", fmt.Sprintf("mfc2 %s, %s[%d]", reg(rt), vreg(rd), e))
		case 0x02:
			return set("cfc2", fmt.Sprintf("cfc2 %s, $%d", reg(rt), rd&3))
		case 0x04:
			return set("mtc2", fmt.Sprintf("mtc2 %s, %s[%d]", reg(rt), vreg(rd), e))
		case 0x06:
			return set("ctc2", fmt.Sprintf("ctc2 %s, $%d", reg(rt), rd&3))
		}
		return word(in, w)
	}

	// Vector ALU: vd = shamt, vs = rd, vt = rt, with an element specifier in rs.
	funct := w & 0x3F
	name, ok := vuOp[funct]
	if !ok {
		return word(in, w)
	}
	e := rs & 15
	switch name {
	case "vnop":
		return set(name, name)
	case "vsar":
		return set(name, fmt.Sprintf("%s %s, %s[%d]", name, vreg(shamt), vreg(rd), e))
	case "vmov", "vrcp", "vrcpl", "vrcph", "vrsq", "vrsql", "vrsqh":
		// The single-lane ops name a source lane in the element field and a
		// destination lane in the shift field of vd.
		return set(name, fmt.Sprintf("%s %s[%d], %s[%d]", name, vreg(shamt), rd&7, vreg(rt), e))
	}
	return set(name, fmt.Sprintf("%s %s, %s, %s[%d]", name, vreg(shamt), vreg(rd), vreg(rt), e))
}

// decodeVecMem handles LWC2/SWC2: the opcode is in the rd field, the element in
// bits 10..7, and the offset is a 7-bit signed value scaled by the access width.
func decodeVecMem(in Inst, w, rs, rt, rd uint32, names []string) Inst {
	name := names[rd&31]
	if name == "" {
		return word(in, w)
	}
	e := (w >> 7) & 15
	off := int32(w&0x7F) << 25 >> 25 // sign-extend 7 bits
	in.Mnem = name
	in.Text = fmt.Sprintf("%s %s[%d], %d(%s)", name, vreg(rt), e, off<<memShift[rd&31], reg(rs))
	return in
}

func word(in Inst, w uint32) Inst {
	in.Mnem, in.Text, in.Flow = ".word", fmt.Sprintf(".word 0x%08X", w), FlowStop
	in.HasTarget, in.HasDelay = false, false
	return in
}

// Disassemble renders code as a linear listing starting at IMEM address base.
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for i := 0; i+4 <= len(code); i += 4 {
		addr := base + uint32(i)
		in := Decode(code[i:], addr)
		out = append(out, fmt.Sprintf("%03X  %02X %02X %02X %02X  %s",
			addr, code[i], code[i+1], code[i+2], code[i+3], in.Text))
	}
	return out
}
