package arm

import "fmt"

// DecodeThumb decodes one Thumb instruction at addr. Thumb instructions are 16-bit,
// except the BL/BLX long branch which is a pair of halfwords decoded here as a
// single 4-byte Inst. A truncated read decodes as data with FlowStop.
//
// The Thumb encoding is the classic 19-"format" layout (ARM7TDMI Data Sheet), and
// this decoder is organised the same way, dispatching on the top bits of the
// halfword.
func DecodeThumb(code []byte, addr uint32) Inst {
	if len(code) < 2 {
		return Inst{Addr: addr, Len: len(code), Mnem: ".hword", Text: ".hword ; truncated", Flow: FlowStop, Thumb: true, Cond: condAL}
	}
	h := uint32(code[0]) | uint32(code[1])<<8
	in := Inst{Addr: addr, Len: 2, Flow: FlowSeq, Thumb: true, TargetThumb: true, Cond: condAL}

	switch {
	case h>>13 == 0b000 && (h>>11)&3 != 0b11: // 1: move shifted register
		op := shiftName[(h>>11)&3]
		in.Mnem = op
		in.Text = fmt.Sprintf("%s %s, %s, #%d", op, lo(h), lo(h>>3), (h>>6)&0x1F)
	case h>>11 == 0b00011: // 2: add/subtract
		return thumbAddSub(h, in)
	case h>>13 == 0b001: // 3: move/compare/add/subtract immediate
		op := [4]string{"MOV", "CMP", "ADD", "SUB"}[(h>>11)&3]
		in.Mnem = op
		in.Text = fmt.Sprintf("%s %s, #0x%X", op, lo(h>>8), h&0xFF)
	case h>>10 == 0b010000: // 4: ALU operations
		return thumbALU(h, in)
	case h>>10 == 0b010001: // 5: hi-register operations / branch-exchange
		return thumbHiReg(h, in)
	case h>>11 == 0b01001: // 6: PC-relative load
		in.Mnem = "LDR"
		in.Text = fmt.Sprintf("LDR %s, [pc, #0x%X]", lo(h>>8), (h&0xFF)*4)
	case h>>12 == 0b0101 && (h>>9)&1 == 0: // 7: load/store with register offset
		return thumbLoadStoreReg(h, in)
	case h>>12 == 0b0101: // 8: load/store sign-extended byte/halfword
		return thumbLoadStoreSExt(h, in)
	case h>>13 == 0b011: // 9: load/store with immediate offset
		return thumbLoadStoreImm(h, in)
	case h>>12 == 0b1000: // 10: load/store halfword
		l := (h >> 11) & 1
		name := "STRH"
		if l == 1 {
			name = "LDRH"
		}
		in.Mnem = name
		in.Text = fmt.Sprintf("%s %s, [%s, #0x%X]", name, lo(h), lo(h>>3), ((h>>6)&0x1F)*2)
	case h>>12 == 0b1001: // 11: SP-relative load/store
		l := (h >> 11) & 1
		name := "STR"
		if l == 1 {
			name = "LDR"
		}
		in.Mnem = name
		in.Text = fmt.Sprintf("%s %s, [sp, #0x%X]", name, lo(h>>8), (h&0xFF)*4)
	case h>>12 == 0b1010: // 12: load address (ADD Rd, PC/SP, #imm)
		base := "pc"
		if (h>>11)&1 == 1 {
			base = "sp"
		}
		in.Mnem = "ADD"
		in.Text = fmt.Sprintf("ADD %s, %s, #0x%X", lo(h>>8), base, (h&0xFF)*4)
	case h>>8 == 0b10110000: // 13: add offset to stack pointer
		sign := ""
		if (h>>7)&1 == 1 {
			sign = "-"
		}
		in.Mnem = "ADD"
		in.Text = fmt.Sprintf("ADD sp, #%s0x%X", sign, (h&0x7F)*4)
	case h>>12 == 0b1011 && (h>>9)&3 == 0b10: // 14: push/pop registers
		return thumbPushPop(h, in)
	case h>>12 == 0b1100: // 15: multiple load/store
		l := (h >> 11) & 1
		name := "STMIA"
		if l == 1 {
			name = "LDMIA"
		}
		in.Mnem = name
		in.Text = fmt.Sprintf("%s %s!, {%s}", name, lo(h>>8), regList(h&0xFF))
	case h>>12 == 0b1101: // 16: conditional branch (and SWI at cond==1111)
		return thumbCondBranch(h, addr, in)
	case h>>11 == 0b11100: // 18: unconditional branch
		off := signExtend(h&0x7FF, 11) << 1
		target := addr + 4 + off
		in.Mnem = "B"
		in.Text = fmt.Sprintf("B 0x%08X", target)
		in.Flow, in.Target, in.HasTarget = FlowJump, target, true
	case h>>11 == 0b11110 || h>>11 == 0b11111 || h>>11 == 0b11101: // 19: long branch with link
		return thumbLongBranch(code, h, addr, in)
	default:
		in.Mnem, in.Flow = ".hword", FlowStop
		in.Text = fmt.Sprintf(".hword 0x%04X", h)
	}
	return in
}

// lo renders a low register (r0-r7) from the bottom 3 bits of v.
func lo(v uint32) string { return regName[v&7] }

func thumbAddSub(h uint32, in Inst) Inst {
	i := (h >> 10) & 1 // immediate operand
	op := (h >> 9) & 1 // 0 ADD, 1 SUB
	rnOff := (h >> 6) & 7
	name := "ADD"
	if op == 1 {
		name = "SUB"
	}
	in.Mnem = name
	if i == 1 {
		in.Text = fmt.Sprintf("%s %s, %s, #%d", name, lo(h), lo(h>>3), rnOff)
	} else {
		in.Text = fmt.Sprintf("%s %s, %s, %s", name, lo(h), lo(h>>3), regName[rnOff])
	}
	return in
}

var thumbALUOps = [16]string{
	"AND", "EOR", "LSL", "LSR", "ASR", "ADC", "SBC", "ROR",
	"TST", "NEG", "CMP", "CMN", "ORR", "MUL", "BIC", "MVN",
}

func thumbALU(h uint32, in Inst) Inst {
	op := thumbALUOps[(h>>6)&0xF]
	in.Mnem = op
	in.Text = fmt.Sprintf("%s %s, %s", op, lo(h), lo(h>>3))
	return in
}

// thumbHiReg handles the hi-register ops (ADD/CMP/MOV that can reach r8-r15) and
// BX/BLX. Writing the PC, or branching through a register, is the control transfer
// the tracer needs.
func thumbHiReg(h uint32, in Inst) Inst {
	op := (h >> 8) & 3
	rd := (h & 7) | (h>>4)&8 // H1:Rd
	rm := (h >> 3) & 0xF     // H2:Rm
	switch op {
	case 0b00: // ADD
		in.Mnem = "ADD"
		in.Text = fmt.Sprintf("ADD %s, %s", regName[rd], regName[rm])
		if rd == 15 {
			in.Flow = FlowIndJump
		}
	case 0b01: // CMP
		in.Mnem = "CMP"
		in.Text = fmt.Sprintf("CMP %s, %s", regName[rd], regName[rm])
	case 0b10: // MOV
		in.Mnem = "MOV"
		in.Text = fmt.Sprintf("MOV %s, %s", regName[rd], regName[rm])
		if rd == 15 {
			if rm == 14 {
				in.Flow = FlowReturn // MOV pc, lr
			} else {
				in.Flow = FlowIndJump
			}
		}
	default: // 0b11: BX / BLX
		if (h>>7)&1 == 0 { // BX
			in.Mnem = "BX"
			in.Text = fmt.Sprintf("BX %s", regName[rm])
			if rm == 14 {
				in.Flow = FlowReturn
			} else {
				in.Flow = FlowIndJump
			}
		} else { // BLX (register) — indirect call
			in.Mnem = "BLX"
			in.Text = fmt.Sprintf("BLX %s", regName[rm])
			in.Flow = FlowIndCall
		}
	}
	return in
}

func thumbLoadStoreReg(h uint32, in Inst) Inst {
	l := (h >> 11) & 1
	b := (h >> 10) & 1
	name := "STR"
	if l == 1 {
		name = "LDR"
	}
	if b == 1 {
		name += "B"
	}
	in.Mnem = name
	in.Text = fmt.Sprintf("%s %s, [%s, %s]", name, lo(h), lo(h>>3), lo(h>>6))
	return in
}

func thumbLoadStoreSExt(h uint32, in Inst) Inst {
	name := [4]string{"STRH", "LDSB", "LDRH", "LDSH"}[(h>>10)&3]
	in.Mnem = name
	in.Text = fmt.Sprintf("%s %s, [%s, %s]", name, lo(h), lo(h>>3), lo(h>>6))
	return in
}

func thumbLoadStoreImm(h uint32, in Inst) Inst {
	b := (h >> 12) & 1
	l := (h >> 11) & 1
	name := "STR"
	if l == 1 {
		name = "LDR"
	}
	off := (h >> 6) & 0x1F
	if b == 1 {
		name += "B" // byte offsets are unscaled
	} else {
		off *= 4 // word offsets are scaled by 4
	}
	in.Mnem = name
	in.Text = fmt.Sprintf("%s %s, [%s, #0x%X]", name, lo(h), lo(h>>3), off)
	return in
}

func thumbPushPop(h uint32, in Inst) Inst {
	l := (h >> 11) & 1
	r := (h >> 8) & 1 // include LR (push) / PC (pop)
	mask := h & 0xFF
	if l == 1 { // POP
		extra := uint32(0)
		if r == 1 {
			extra = 1 << 15 // pc
		}
		in.Mnem = "POP"
		in.Text = fmt.Sprintf("POP {%s}", regList(mask|extra))
		if r == 1 { // POP {…, pc} → return
			in.Flow = FlowReturn
		}
	} else { // PUSH
		extra := uint32(0)
		if r == 1 {
			extra = 1 << 14 // lr
		}
		in.Mnem = "PUSH"
		in.Text = fmt.Sprintf("PUSH {%s}", regList(mask|extra))
	}
	return in
}

func thumbCondBranch(h, addr uint32, in Inst) Inst {
	cond := int((h >> 8) & 0xF)
	if cond == 0b1111 { // SWI
		in.Mnem = "SWI"
		in.Text = fmt.Sprintf("SWI #0x%X", h&0xFF)
		in.Cond = condAL
		return in // returns from the BIOS; keep tracing
	}
	if cond == 0b1110 { // undefined
		in.Mnem, in.Flow = ".hword", FlowStop
		in.Text = fmt.Sprintf(".hword 0x%04X", h)
		return in
	}
	off := signExtend(h&0xFF, 8) << 1
	target := addr + 4 + off
	in.Cond = cond
	in.Mnem = "B" + condName[cond]
	in.Text = fmt.Sprintf("%s 0x%08X", in.Mnem, target)
	in.Flow, in.Target, in.HasTarget = FlowBranch, target, true
	return in
}

// thumbLongBranch decodes the BL/BLX long branch. It is a pair of halfwords: a
// high-offset prefix (11110) then a BL (11111) or BLX (11101) suffix. When the
// following halfword is not a valid suffix the prefix is decoded alone (it merely
// loads LR) so the trace stays halfword-aligned.
func thumbLongBranch(code []byte, h, addr uint32, in Inst) Inst {
	top := h >> 11
	if top != 0b11110 { // a stray suffix with no preceding prefix — treat as data-ish
		in.Mnem = "BL(suffix)"
		in.Text = fmt.Sprintf("; BL/BLX suffix (no prefix) 0x%04X", h)
		return in
	}
	if len(code) < 4 {
		in.Mnem = "BL(hi)"
		in.Text = fmt.Sprintf("BL (hi) #0x%X", h&0x7FF)
		return in // just sets LR; keep tracing
	}
	h2 := uint32(code[2]) | uint32(code[3])<<8
	suf := h2 >> 11
	if suf != 0b11111 && suf != 0b11101 {
		in.Mnem = "BL(hi)"
		in.Text = fmt.Sprintf("BL (hi) #0x%X", h&0x7FF)
		return in
	}
	in.Len = 4
	hiOff := signExtend(h&0x7FF, 11) << 12
	loOff := (h2 & 0x7FF) << 1
	target := addr + 4 + hiOff + loOff
	if suf == 0b11101 { // BLX — switch to ARM, word-align
		target &^= 3
		in.Mnem = "BLX"
		in.Text = fmt.Sprintf("BLX 0x%08X", target)
		in.Flow, in.Target, in.HasTarget, in.TargetThumb = FlowCall, target, true, false
	} else { // BL — stay in Thumb
		in.Mnem = "BL"
		in.Text = fmt.Sprintf("BL 0x%08X", target)
		in.Flow, in.Target, in.HasTarget, in.TargetThumb = FlowCall, target, true, true
	}
	return in
}
