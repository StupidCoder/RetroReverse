package r5900

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

func vfreg(i uint32) string { return fmt.Sprintf("$vf%d", i&31) }

// cop0Name labels the system-control registers the kernel touches.
var cop0Name = [32]string{
	0: "Index", 1: "Random", 2: "EntryLo0", 3: "EntryLo1", 4: "Context",
	5: "PageMask", 6: "Wired", 8: "BadVAddr", 9: "Count", 10: "EntryHi",
	11: "Compare", 12: "Status", 13: "Cause", 14: "EPC", 15: "PRId",
	16: "Config", 23: "BadPAddr", 24: "Debug", 25: "Perf", 30: "ErrorEPC",
}

func cop0Reg(i uint32) string {
	if n := cop0Name[i&31]; n != "" {
		return n
	}
	return fmt.Sprintf("$%d", i&31)
}

// Decode decodes the instruction at code[0:], whose CPU address is addr (used to
// compute branch and jump targets). A short slice yields Len 0.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 4 {
		return Inst{Addr: addr, Len: 0, Mnem: ".word", Text: ".word <truncated>", Flow: FlowStop}
	}
	return DecodeWord(binary.LittleEndian.Uint32(code), addr)
}

// DecodeWord decodes an already-assembled instruction word. The machine model uses
// it to disassemble memory it has fetched itself.
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
	branch := func(mnem, text string, likely bool) {
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		in.Annul = likely
		set(mnem, text)
	}
	mem := func(m string, r string) {
		set(m, fmt.Sprintf("%s %s, %d(%s)", m, r, simm, reg(rs)))
	}

	switch op {
	case 0x00: // SPECIAL — dispatch on funct
		return decodeSpecial(in, w, rs, rt, rd, shamt)

	case 0x01: // REGIMM — dispatch on rt
		return decodeRegimm(in, w, rs, rt, branchT, simm)

	case 0x1C: // MMI — the SIMD space, dispatch on funct
		return decodeMMI(in, w, rs, rt, rd, shamt)

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
	case 0x11: // COP1 (the FPU)
		return decodeCop1(in, w, rs, rt, rd, shamt, branchT)
	case 0x12: // COP2 (VU0 macro mode)
		return decodeCop2(in, w, rs, rt, rd, branchT)

	case 0x1A: // ldl
		mem("ldl", reg(rt))
	case 0x1B: // ldr
		mem("ldr", reg(rt))

	case 0x1E: // lq — the 128-bit load. The address is quadword-aligned by masking.
		mem("lq", reg(rt))
	case 0x1F: // sq
		mem("sq", reg(rt))

	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27: // loads
		m := [...]string{0x20: "lb", 0x21: "lh", 0x22: "lwl", 0x23: "lw",
			0x24: "lbu", 0x25: "lhu", 0x26: "lwr", 0x27: "lwu"}[op]
		mem(m, reg(rt))

	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E: // stores
		m := [...]string{0x28: "sb", 0x29: "sh", 0x2A: "swl", 0x2B: "sw",
			0x2C: "sdl", 0x2D: "sdr", 0x2E: "swr"}[op]
		mem(m, reg(rt))

	case 0x2F: // cache — a no-op here; see exec.go
		set("cache", fmt.Sprintf("cache 0x%X, %d(%s)", rt, simm, reg(rs)))
	case 0x33: // pref — a hint, and a no-op
		set("pref", fmt.Sprintf("pref 0x%X, %d(%s)", rt, simm, reg(rs)))

	case 0x30: // ll
		mem("ll", reg(rt))
	case 0x34: // lld
		mem("lld", reg(rt))
	case 0x37: // ld
		mem("ld", reg(rt))
	case 0x38: // sc
		mem("sc", reg(rt))
	case 0x3C: // scd
		mem("scd", reg(rt))
	case 0x3F: // sd
		mem("sd", reg(rt))

	case 0x31: // lwc1
		mem("lwc1", freg(rt))
	case 0x39: // swc1
		mem("swc1", freg(rt))

	case 0x36: // lqc2 — a whole vector register from memory
		mem("lqc2", vfreg(rt))
	case 0x3E: // sqc2
		mem("sqc2", vfreg(rt))

	default:
		return word(in, w)
	}
	return in
}

// decodeSpecial handles op == 0: register-format arithmetic, shifts, jr/jalr,
// mult/div, syscall/break/sync, HI/LO and SA moves, traps, and the 64-bit forms.
func decodeSpecial(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	funct := w & 63
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
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

	// The MIPS IV conditional moves, which the EE has and the VR4300 does not.
	case 0x0A:
		return r3("movz")
	case 0x0B:
		return r3("movn")

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

	// The EE's mult and multu take a destination register as well as writing
	// HI/LO — "mult rd, rs, rt" — which is why they are formatted as three-operand
	// instructions here and two-operand on the VR4300.
	case 0x18:
		return set("mult", fmt.Sprintf("mult %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x19:
		return set("multu", fmt.Sprintf("multu %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x1A:
		return r2("div")
	case 0x1B:
		return r2("divu")

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

	// The shift-amount register, which qfsrv reads to extract an unaligned quadword.
	case 0x28:
		return set("mfsa", fmt.Sprintf("mfsa %s", reg(rd)))
	case 0x29:
		return set("mtsa", fmt.Sprintf("mtsa %s", reg(rs)))

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
// likely and linking forms), the immediate traps, and the two SA loads.
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

	// The linking forms write $ra, so a tracer treats them as calls.
	case 0x10:
		return call(br("bltzal", false))
	case 0x11:
		return call(br("bgezal", false))
	case 0x12:
		return call(br("bltzall", true))
	case 0x13:
		return call(br("bgezall", true))

	// mtsab/mtsah set the shift-amount register from a register plus an immediate,
	// in bytes and halfwords respectively.
	case 0x18:
		return set("mtsab", fmt.Sprintf("mtsab %s, 0x%X", reg(rs), uint16(simm)))
	case 0x19:
		return set("mtsah", fmt.Sprintf("mtsah %s, 0x%X", reg(rs), uint16(simm)))
	}
	return word(in, w)
}

// decodeCop0 handles the system-control moves and the TLB / ERET / EI / DI
// commands.
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
		case 0x38:
			return set("ei", "ei")
		case 0x39:
			return set("di", "di")
		}
		return word(in, w)
	}
	switch rs {
	case 0x00:
		return set("mfc0", fmt.Sprintf("mfc0 %s, %s", reg(rt), cop0Reg(rd)))
	case 0x04:
		return set("mtc0", fmt.Sprintf("mtc0 %s, %s", reg(rt), cop0Reg(rd)))
	}
	return word(in, w)
}

// decodeCop1 handles the FPU. The EE's COP1 is single-precision only, so there is
// no format field to dispatch on in the VR4300's sense: rs == 0x10 is "single" and
// rs == 0x14 is "word", and that is the whole of it.
func decodeCop1(in Inst, w, rs, rt, rd, shamt, branchT uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }

	switch rs {
	case 0x00:
		return set("mfc1", fmt.Sprintf("mfc1 %s, %s", reg(rt), freg(rd)))
	case 0x02:
		return set("cfc1", fmt.Sprintf("cfc1 %s, $%d", reg(rt), rd))
	case 0x04:
		return set("mtc1", fmt.Sprintf("mtc1 %s, %s", reg(rt), freg(rd)))
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

	case 0x10: // single-precision arithmetic
		ft, fs, fd := rt, rd, shamt
		funct := w & 0x3F
		// The comparisons write the condition flag rather than a register.
		switch funct {
		case 0x30:
			return set("c.f.s", fmt.Sprintf("c.f.s %s, %s", freg(fs), freg(ft)))
		case 0x32:
			return set("c.eq.s", fmt.Sprintf("c.eq.s %s, %s", freg(fs), freg(ft)))
		case 0x34:
			return set("c.lt.s", fmt.Sprintf("c.lt.s %s, %s", freg(fs), freg(ft)))
		case 0x36:
			return set("c.le.s", fmt.Sprintf("c.le.s %s, %s", freg(fs), freg(ft)))
		}
		// Three-operand: fd = fs op ft.
		if n, ok := map[uint32]string{
			0x00: "add.s", 0x01: "sub.s", 0x02: "mul.s", 0x03: "div.s",
			0x28: "max.s", 0x29: "min.s",
		}[funct]; ok {
			return set(n, fmt.Sprintf("%s %s, %s, %s", n, freg(fd), freg(fs), freg(ft)))
		}
		// Two-operand: fd = op fs.
		if n, ok := map[uint32]string{
			0x04: "sqrt.s", 0x05: "abs.s", 0x06: "mov.s", 0x07: "neg.s",
			0x16: "rsqrt.s", 0x24: "cvt.w.s",
		}[funct]; ok {
			return set(n, fmt.Sprintf("%s %s, %s", n, freg(fd), freg(fs)))
		}
		// The accumulator family. "adda/suba/mula" set the accumulator; "madd/msub"
		// use it and write fd; "madda/msuba" use it and write it back.
		if n, ok := map[uint32]string{
			0x18: "adda.s", 0x19: "suba.s", 0x1A: "mula.s",
			0x1E: "madda.s", 0x1F: "msuba.s",
		}[funct]; ok {
			return set(n, fmt.Sprintf("%s %s, %s", n, freg(fs), freg(ft)))
		}
		if n, ok := map[uint32]string{0x1C: "madd.s", 0x1D: "msub.s"}[funct]; ok {
			return set(n, fmt.Sprintf("%s %s, %s, %s", n, freg(fd), freg(fs), freg(ft)))
		}
		return word(in, w)

	case 0x14: // word (integer) source
		if w&0x3F == 0x20 {
			return set("cvt.s.w", fmt.Sprintf("cvt.s.w %s, %s", freg(shamt), freg(rd)))
		}
		return word(in, w)
	}
	return word(in, w)
}

// decodeCop2 handles VU0 macro mode: the moves between the EE and the vector unit,
// the branches on the VU's condition flags, and the vector instructions themselves.
//
// The vector instruction space is large and is disassembled by tools/cpu/vu, which
// owns the semantics. Here it is enough to name the moves and mark everything else
// as a VU operation, so a trace stays readable and the flow analysis is right.
func decodeCop2(in Inst, w, rs, rt, rd, branchT uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }

	if w&(1<<25) != 0 { // a vector operation, not a move
		return set("vu", fmt.Sprintf("vu 0x%08X", w))
	}
	switch rs {
	case 0x01:
		return set("qmfc2", fmt.Sprintf("qmfc2 %s, %s", reg(rt), vfreg(rd)))
	case 0x02:
		return set("cfc2", fmt.Sprintf("cfc2 %s, $vi%d", reg(rt), rd))
	case 0x05:
		return set("qmtc2", fmt.Sprintf("qmtc2 %s, %s", reg(rt), vfreg(rd)))
	case 0x06:
		return set("ctc2", fmt.Sprintf("ctc2 %s, $vi%d", reg(rt), rd))
	case 0x08: // BC2: branch on the VU0 condition flag
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, branchT, true, true
		var m string
		switch rt & 3 {
		case 0:
			m = "bc2f"
		case 1:
			m = "bc2t"
		case 2:
			m, in.Annul = "bc2fl", true
		case 3:
			m, in.Annul = "bc2tl", true
		}
		return set(m, fmt.Sprintf("%s $%08X", m, branchT))
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
