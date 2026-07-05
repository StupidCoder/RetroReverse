package x86

import "fmt"

// fpu decodes an x87 escape opcode (0xD8–0xDF). The memory forms take a ModR/M
// memory operand whose reg field selects the operation; the register forms
// (mod==3) are a per-opcode table over ST(i). Either way the ModR/M byte (and
// any displacement) is fully consumed, so Inst.Len is always correct even for
// the rarer encodings.
func (d *dec) fpu(op byte) Inst {
	b := d.next()
	mod, reg, rm := b>>6, (b>>3)&7, b&7
	if mod != 3 {
		var memtext string
		if d.addrsize == 32 {
			memtext = d.mem32(mod, rm)
		} else {
			memtext = d.mem16(mod, rm)
		}
		return d.fpuMem(op, reg, memtext)
	}
	return d.fpuReg(op, b, rm)
}

// fpuMem names a memory-operand x87 instruction.
func (d *dec) fpuMem(op, reg byte, mem string) Inst {
	var name, kw string
	switch op {
	case 0xD8:
		name, kw = [8]string{"FADD", "FMUL", "FCOM", "FCOMP", "FSUB", "FSUBR", "FDIV", "FDIVR"}[reg], "DWORD "
	case 0xD9:
		name = [8]string{"FLD", "", "FST", "FSTP", "FLDENV", "FLDCW", "FNSTENV", "FNSTCW"}[reg]
		if reg == 0 || reg == 2 || reg == 3 {
			kw = "DWORD "
		} else if reg == 5 || reg == 7 {
			kw = "WORD "
		}
	case 0xDA:
		name, kw = [8]string{"FIADD", "FIMUL", "FICOM", "FICOMP", "FISUB", "FISUBR", "FIDIV", "FIDIVR"}[reg], "DWORD "
	case 0xDB:
		name = [8]string{"FILD", "FISTTP", "FIST", "FISTP", "", "FLD", "", "FSTP"}[reg]
		if reg == 5 || reg == 7 {
			kw = "TBYTE "
		} else {
			kw = "DWORD "
		}
	case 0xDC:
		name, kw = [8]string{"FADD", "FMUL", "FCOM", "FCOMP", "FSUB", "FSUBR", "FDIV", "FDIVR"}[reg], "QWORD "
	case 0xDD:
		name = [8]string{"FLD", "FISTTP", "FST", "FSTP", "FRSTOR", "", "FNSAVE", "FNSTSW"}[reg]
		if reg == 0 || reg == 2 || reg == 3 {
			kw = "QWORD "
		} else if reg == 7 {
			kw = "WORD "
		}
	case 0xDE:
		name, kw = [8]string{"FIADD", "FIMUL", "FICOM", "FICOMP", "FISUB", "FISUBR", "FIDIV", "FIDIVR"}[reg], "WORD "
	default: // 0xDF
		name = [8]string{"FILD", "FISTTP", "FIST", "FISTP", "FBLD", "FILD", "FBSTP", "FISTP"}[reg]
		switch reg {
		case 4, 6:
			kw = "TBYTE "
		case 5, 7:
			kw = "QWORD "
		default:
			kw = "WORD "
		}
	}
	if name == "" {
		return Inst{Mnem: ".byte", Text: fmt.Sprintf(".byte $%02X,$%02X", op, d.mem[d.p-1]), Flow: FlowStop}
	}
	return op1(name, kw+mem)
}

// fpuReg names a register-stack (mod==3) x87 instruction. b is the full ModR/M
// byte; sti is its low three bits (the stack index).
func (d *dec) fpuReg(op, b, sti byte) Inst {
	st := fmt.Sprintf("ST(%d)", sti)
	sub := (b >> 3) & 7 // the reg field, selecting the arithmetic sub-op
	switch op {
	case 0xD8:
		return op2([8]string{"FADD", "FMUL", "FCOM", "FCOMP", "FSUB", "FSUBR", "FDIV", "FDIVR"}[sub], "ST", st)
	case 0xD9:
		switch {
		case b >= 0xC0 && b <= 0xC7:
			return op2("FLD", "ST", st)
		case b >= 0xC8 && b <= 0xCF:
			return op2("FXCH", "ST", st)
		}
		if n, ok := d9tab[b]; ok {
			return mk(n, n)
		}
	case 0xDA:
		if b == 0xE9 {
			return mk("FUCOMPP", "FUCOMPP")
		}
		if n, ok := map[byte]string{0xC0: "FCMOVB", 0xC8: "FCMOVE", 0xD0: "FCMOVBE", 0xD8: "FCMOVU"}[b&0xF8]; ok {
			return op2(n, "ST", st)
		}
	case 0xDB:
		switch b {
		case 0xE2:
			return mk("FNCLEX", "FNCLEX")
		case 0xE3:
			return mk("FNINIT", "FNINIT")
		}
		if n, ok := map[byte]string{0xC0: "FCMOVNB", 0xC8: "FCMOVNE", 0xD0: "FCMOVNBE", 0xD8: "FCMOVNU", 0xE8: "FUCOMI", 0xF0: "FCOMI"}[b&0xF8]; ok {
			return op2(n, "ST", st)
		}
	case 0xDC:
		return op2([8]string{"FADD", "FMUL", "FCOM", "FCOMP", "FSUBR", "FSUB", "FDIVR", "FDIV"}[sub], st, "ST")
	case 0xDD:
		if n, ok := map[byte]string{0xC0: "FFREE", 0xD0: "FST", 0xD8: "FSTP", 0xE0: "FUCOM", 0xE8: "FUCOMP"}[b&0xF8]; ok {
			return op1(n, st)
		}
	case 0xDE:
		if b == 0xD9 {
			return mk("FCOMPP", "FCOMPP")
		}
		return op2([8]string{"FADDP", "FMULP", "FCOMP", "FCOMPP", "FSUBRP", "FSUBP", "FDIVRP", "FDIVP"}[sub], st, "ST")
	}
	// DF special forms + generic fallback.
	if op == 0xDF {
		switch b {
		case 0xE0:
			return op1("FNSTSW", "AX")
		}
		if n, ok := map[byte]string{0xC0: "FFREEP", 0xE8: "FUCOMIP", 0xF0: "FCOMIP"}[b&0xF8]; ok {
			return op2(n, "ST", st)
		}
	}
	// Anything not individually named still has the right length (2 bytes);
	// render it generically so a trace stays readable and aligned.
	return mk("ESC", fmt.Sprintf("ESC $%02X,$%02X", op, b))
}

// d9tab holds the D9 no-operand register-form instructions (constants and
// transcendental/setup ops).
var d9tab = map[byte]string{
	0xD0: "FNOP", 0xE0: "FCHS", 0xE1: "FABS", 0xE4: "FTST", 0xE5: "FXAM",
	0xE8: "FLD1", 0xE9: "FLDL2T", 0xEA: "FLDL2E", 0xEB: "FLDPI",
	0xEC: "FLDLG2", 0xED: "FLDLN2", 0xEE: "FLDZ",
	0xF0: "F2XM1", 0xF1: "FYL2X", 0xF2: "FPTAN", 0xF3: "FPATAN",
	0xF4: "FXTRACT", 0xF5: "FPREM1", 0xF6: "FDECSTP", 0xF7: "FINCSTP",
	0xF8: "FPREM", 0xF9: "FYL2XP1", 0xFA: "FSQRT", 0xFB: "FSINCOS",
	0xFC: "FRNDINT", 0xFD: "FSCALE", 0xFE: "FSIN", 0xFF: "FCOS",
}
