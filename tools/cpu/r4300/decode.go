package r4300

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

func freg(i uint32) string { return fmt.Sprintf("$f%d", i&31) }

// cop0Name labels the system-control registers the boot code and libultra touch.
var cop0Name = [32]string{
	0: "Index", 1: "Random", 2: "EntryLo0", 3: "EntryLo1", 4: "Context",
	5: "PageMask", 6: "Wired", 8: "BadVAddr", 9: "Count", 10: "EntryHi",
	11: "Compare", 12: "Status", 13: "Cause", 14: "EPC", 15: "PRId",
	16: "Config", 17: "LLAddr", 18: "WatchLo", 19: "WatchHi", 20: "XContext",
	26: "ParityError", 28: "TagLo", 29: "TagHi", 30: "ErrorEPC",
}

func cop0Reg(i uint32) string {
	if n := cop0Name[i&31]; n != "" {
		return n
	}
	return fmt.Sprintf("$%d", i&31)
}

// fmtName is the COP1 format field: the operand type an FPU op works on.
var fmtName = map[uint32]string{0x10: "s", 0x11: "d", 0x14: "w", 0x15: "l"}

// fpuCond is the 16-way condition encoded in the low 4 bits of a c.cond.fmt.
var fpuCond = [16]string{
	"f", "un", "eq", "ueq", "olt", "ult", "ole", "ule",
	"sf", "ngle", "seq", "ngl", "lt", "nge", "le", "ngt",
}

// Decode decodes the VR4300 instruction at code[0:], whose CPU address is addr
// (used to compute branch and jump targets). A short slice yields Len 0.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 4 {
		return Inst{Addr: addr, Len: 0, Mnem: ".word", Text: ".word <truncated>", Flow: FlowStop}
	}
	return DecodeWord(binary.BigEndian.Uint32(code), addr)
}

// DecodeWord decodes an already-assembled instruction word. The machine model
// uses it to disassemble memory it has fetched itself.
func DecodeWord(w, addr uint32) Inst {
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq}

	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	imm := w & 0xFFFF
	simm := int32(int16(imm))            // sign-extended immediate
	branchT := addr + 4 + uint32(simm)*4 // PC-relative branch target
	jumpT := (addr+4)&0xF0000000 | (w&0x03FFFFFF)<<2

	set := func(mnem, text string) { in.Mnem, in.Text = mnem, text }
	// branch marks a conditional PC-relative branch; likely ones annul their slot.
	branch := func(mnem, text string, likely bool) {
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		in.Annul = likely
		set(mnem, text)
	}

	switch op {
	case 0x00: // SPECIAL — dispatch on funct
		return decodeSpecial(in, w, rs, rt, rd, shamt)

	case 0x01: // REGIMM — dispatch on rt
		return decodeRegimm(in, w, rs, rt, branchT, simm)

	case 0x02: // j
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowJump, jumpT, true, true
		set("j", fmt.Sprintf("j $%08X", jumpT))
	case 0x03: // jal
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowCall, jumpT, true, true
		set("jal", fmt.Sprintf("jal $%08X", jumpT))

	case 0x04: // beq
		if rs == 0 && rt == 0 { // beq $0,$0 == unconditional b
			in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowJump, branchT, true, true
			set("b", fmt.Sprintf("b $%08X", branchT))
		} else {
			branch("beq", fmt.Sprintf("beq %s, %s, $%08X", reg(rs), reg(rt), branchT), false)
		}
	case 0x05: // bne
		branch("bne", fmt.Sprintf("bne %s, %s, $%08X", reg(rs), reg(rt), branchT), false)
	case 0x06: // blez
		branch("blez", fmt.Sprintf("blez %s, $%08X", reg(rs), branchT), false)
	case 0x07: // bgtz
		branch("bgtz", fmt.Sprintf("bgtz %s, $%08X", reg(rs), branchT), false)

	// The branch-likely family: the delay slot is annulled when not taken.
	case 0x14: // beql
		branch("beql", fmt.Sprintf("beql %s, %s, $%08X", reg(rs), reg(rt), branchT), true)
	case 0x15: // bnel
		branch("bnel", fmt.Sprintf("bnel %s, %s, $%08X", reg(rs), reg(rt), branchT), true)
	case 0x16: // blezl
		branch("blezl", fmt.Sprintf("blezl %s, $%08X", reg(rs), branchT), true)
	case 0x17: // bgtzl
		branch("bgtzl", fmt.Sprintf("bgtzl %s, $%08X", reg(rs), branchT), true)

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

	case 0x18: // daddi
		set("daddi", fmt.Sprintf("daddi %s, %s, %d", reg(rt), reg(rs), simm))
	case 0x19: // daddiu
		set("daddiu", fmt.Sprintf("daddiu %s, %s, %d", reg(rt), reg(rs), simm))

	case 0x10: // COP0
		return decodeCop0(in, w, rs, rt, rd)
	case 0x11: // COP1 (FPU)
		return decodeCop1(in, w, rs, rt, rd, shamt, branchT)

	case 0x1A: // ldl
		set("ldl", fmt.Sprintf("ldl %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x1B: // ldr
		set("ldr", fmt.Sprintf("ldr %s, %d(%s)", reg(rt), simm, reg(rs)))

	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27: // loads
		m := [...]string{0x20: "lb", 0x21: "lh", 0x22: "lwl", 0x23: "lw",
			0x24: "lbu", 0x25: "lhu", 0x26: "lwr", 0x27: "lwu"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))

	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E: // stores
		m := [...]string{0x28: "sb", 0x29: "sh", 0x2A: "swl", 0x2B: "sw",
			0x2C: "sdl", 0x2D: "sdr", 0x2E: "swr"}[op]
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, reg(rt), simm, reg(rs)))

	case 0x2F: // cache — a no-op here; see exec.go
		set("cache", fmt.Sprintf("cache 0x%X, %d(%s)", rt, simm, reg(rs)))

	case 0x30: // ll
		set("ll", fmt.Sprintf("ll %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x34: // lld
		set("lld", fmt.Sprintf("lld %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x37: // ld
		set("ld", fmt.Sprintf("ld %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x38: // sc
		set("sc", fmt.Sprintf("sc %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x3C: // scd
		set("scd", fmt.Sprintf("scd %s, %d(%s)", reg(rt), simm, reg(rs)))
	case 0x3F: // sd
		set("sd", fmt.Sprintf("sd %s, %d(%s)", reg(rt), simm, reg(rs)))

	case 0x31: // lwc1
		set("lwc1", fmt.Sprintf("lwc1 %s, %d(%s)", freg(rt), simm, reg(rs)))
	case 0x35: // ldc1
		set("ldc1", fmt.Sprintf("ldc1 %s, %d(%s)", freg(rt), simm, reg(rs)))
	case 0x39: // swc1
		set("swc1", fmt.Sprintf("swc1 %s, %d(%s)", freg(rt), simm, reg(rs)))
	case 0x3D: // sdc1
		set("sdc1", fmt.Sprintf("sdc1 %s, %d(%s)", freg(rt), simm, reg(rs)))

	default:
		return word(in, w)
	}
	return in
}

// decodeSpecial handles op == 0: register-format arithmetic, shifts, jr/jalr,
// mult/div, syscall/break/sync, HI/LO moves, traps, and the 64-bit forms.
func decodeSpecial(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	funct := w & 63
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	// r3 formats a three-register op, r2 a two-register one, sh a shift.
	r3 := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %s", m, reg(rd), reg(rs), reg(rt))) }
	sh := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %d", m, reg(rd), reg(rt), shamt)) }
	shv := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s, %s", m, reg(rd), reg(rt), reg(rs))) }
	r2 := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %s", m, reg(rs), reg(rt))) }

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
	case 0x0F:
		return set("sync", "sync")

	case 0x10:
		return set("mfhi", fmt.Sprintf("mfhi %s", reg(rd)))
	case 0x11:
		return set("mthi", fmt.Sprintf("mthi %s", reg(rs)))
	case 0x12:
		return set("mflo", fmt.Sprintf("mflo %s", reg(rd)))
	case 0x13:
		return set("mtlo", fmt.Sprintf("mtlo %s", reg(rs)))
	case 0x14:
		return shv("dsllv")
	case 0x16:
		return shv("dsrlv")
	case 0x17:
		return shv("dsrav")

	case 0x18:
		return r2("mult")
	case 0x19:
		return r2("multu")
	case 0x1A:
		return r2("div")
	case 0x1B:
		return r2("divu")
	case 0x1C:
		return r2("dmult")
	case 0x1D:
		return r2("dmultu")
	case 0x1E:
		return r2("ddiv")
	case 0x1F:
		return r2("ddivu")

	case 0x20:
		return r3("add")
	case 0x21:
		if rt == 0 {
			return set("move", fmt.Sprintf("move %s, %s", reg(rd), reg(rs)))
		}
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
	case 0x2C:
		return r3("dadd")
	case 0x2D:
		return r3("daddu")
	case 0x2E:
		return r3("dsub")
	case 0x2F:
		return r3("dsubu")

	case 0x30:
		return r2("tge")
	case 0x31:
		return r2("tgeu")
	case 0x32:
		return r2("tlt")
	case 0x33:
		return r2("tltu")
	case 0x34:
		return r2("teq")
	case 0x36:
		return r2("tne")

	case 0x38:
		return sh("dsll")
	case 0x3A:
		return sh("dsrl")
	case 0x3B:
		return sh("dsra")
	// The "32" forms add 32 to the shift amount, reaching 32..63.
	case 0x3C:
		return set("dsll32", fmt.Sprintf("dsll32 %s, %s, %d", reg(rd), reg(rt), shamt))
	case 0x3E:
		return set("dsrl32", fmt.Sprintf("dsrl32 %s, %s, %d", reg(rd), reg(rt), shamt))
	case 0x3F:
		return set("dsra32", fmt.Sprintf("dsra32 %s, %s, %d", reg(rd), reg(rt), shamt))
	}
	return word(in, w)
}

// decodeRegimm handles op == 1: the register-immediate branches (including the
// likely and linking forms) and the immediate traps, all selected by rt.
func decodeRegimm(in Inst, w, rs, rt, branchT uint32, simm int32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	br := func(m string, likely bool) Inst {
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		in.Annul = likely
		return set(m, fmt.Sprintf("%s %s, $%08X", m, reg(rs), branchT))
	}
	trap := func(m string) Inst { return set(m, fmt.Sprintf("%s %s, %d", m, reg(rs), simm)) }

	switch rt {
	case 0x00:
		return br("bltz", false)
	case 0x01:
		return br("bgez", false)
	case 0x02:
		return br("bltzl", true)
	case 0x03:
		return br("bgezl", true)
	case 0x08:
		return trap("tgei")
	case 0x09:
		return trap("tgeiu")
	case 0x0A:
		return trap("tlti")
	case 0x0B:
		return trap("tltiu")
	case 0x0C:
		return trap("teqi")
	case 0x0E:
		return trap("tnei")
	// The linking forms write $ra, so a tracer treats them as calls. br sets
	// FlowBranch, so the reclassification has to follow it.
	case 0x10:
		return call(br("bltzal", false))
	case 0x11:
		return call(br("bgezal", false))
	case 0x12:
		return call(br("bltzall", true))
	case 0x13:
		return call(br("bgezall", true))
	}
	return word(in, w)
}

// decodeCop0 handles the system-control moves and the TLB / ERET commands.
func decodeCop0(in Inst, w, rs, rt, rd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	if w&(1<<25) != 0 { // COP0 command
		switch w & 0x3F {
		case 0x01:
			return set("tlbr", "tlbr")
		case 0x02:
			return set("tlbwi", "tlbwi")
		case 0x06:
			return set("tlbwr", "tlbwr")
		case 0x08:
			return set("tlbp", "tlbp")
		case 0x18:
			in.Flow = FlowStop // returns to EPC: the static path ends here
			return set("eret", "eret")
		}
		return word(in, w)
	}
	switch rs {
	case 0x00:
		return set("mfc0", fmt.Sprintf("mfc0 %s, %s", reg(rt), cop0Reg(rd)))
	case 0x01:
		return set("dmfc0", fmt.Sprintf("dmfc0 %s, %s", reg(rt), cop0Reg(rd)))
	case 0x04:
		return set("mtc0", fmt.Sprintf("mtc0 %s, %s", reg(rt), cop0Reg(rd)))
	case 0x05:
		return set("dmtc0", fmt.Sprintf("dmtc0 %s, %s", reg(rt), cop0Reg(rd)))
	}
	return word(in, w)
}

// decodeCop1 handles the FPU: register moves, the FP conditional branches, and
// the format-dispatched arithmetic.
func decodeCop1(in Inst, w, rs, rt, rd, shamt, branchT uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }

	switch rs {
	case 0x00:
		return set("mfc1", fmt.Sprintf("mfc1 %s, %s", reg(rt), freg(rd)))
	case 0x01:
		return set("dmfc1", fmt.Sprintf("dmfc1 %s, %s", reg(rt), freg(rd)))
	case 0x02:
		return set("cfc1", fmt.Sprintf("cfc1 %s, $%d", reg(rt), rd))
	case 0x04:
		return set("mtc1", fmt.Sprintf("mtc1 %s, %s", reg(rt), freg(rd)))
	case 0x05:
		return set("dmtc1", fmt.Sprintf("dmtc1 %s, %s", reg(rt), freg(rd)))
	case 0x06:
		return set("ctc1", fmt.Sprintf("ctc1 %s, $%d", reg(rt), rd))

	case 0x08: // BC1: branch on the FPU condition flag
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		var m string
		switch rt & 3 {
		case 0:
			m = "bc1f"
		case 1:
			m = "bc1t"
		case 2:
			m, in.Annul = "bc1fl", true
		case 3:
			m, in.Annul = "bc1tl", true
		}
		return set(m, fmt.Sprintf("%s $%08X", m, branchT))
	}

	// Format-dispatched arithmetic: rs is the format (s/d/w/l), funct the op.
	f, ok := fmtName[rs]
	if !ok {
		return word(in, w)
	}
	funct := w & 0x3F
	ft, fs, fd := rt, rd, shamt // in FP arithmetic the fields name FP registers

	// c.cond.fmt occupies funct 0x30..0x3F.
	if funct&0x30 == 0x30 {
		m := "c." + fpuCond[funct&0xF] + "." + f
		return set(m, fmt.Sprintf("%s %s, %s", m, freg(fs), freg(ft)))
	}

	// Three-operand arithmetic.
	if n, ok := map[uint32]string{0x00: "add", 0x01: "sub", 0x02: "mul", 0x03: "div"}[funct]; ok {
		m := n + "." + f
		return set(m, fmt.Sprintf("%s %s, %s, %s", m, freg(fd), freg(fs), freg(ft)))
	}
	// Two-operand: unary arithmetic, the rounding conversions, and cvt.
	unary := map[uint32]string{
		0x04: "sqrt", 0x05: "abs", 0x06: "mov", 0x07: "neg",
		0x08: "round.l", 0x09: "trunc.l", 0x0A: "ceil.l", 0x0B: "floor.l",
		0x0C: "round.w", 0x0D: "trunc.w", 0x0E: "ceil.w", 0x0F: "floor.w",
		0x20: "cvt.s", 0x21: "cvt.d", 0x24: "cvt.w", 0x25: "cvt.l",
	}
	if n, ok := unary[funct]; ok {
		m := n + "." + f
		return set(m, fmt.Sprintf("%s %s, %s", m, freg(fd), freg(fs)))
	}
	return word(in, w)
}

// call reclassifies a decoded branch as a call, for the linking forms.
func call(in Inst) Inst { in.Flow = FlowCall; return in }

// word marks the instruction as an unmodelled data word that stops a trace.
func word(in Inst, w uint32) Inst {
	in.Mnem, in.Text, in.Flow = ".word", fmt.Sprintf(".word 0x%08X", w), FlowStop
	in.HasTarget, in.HasDelay, in.Annul = false, false, false
	return in
}
