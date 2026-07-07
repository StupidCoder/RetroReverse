// Package arm60 is a disassembler, code-tracer front-end and execution core for
// the ARM60 — the CPU of the 3DO Interactive Multiplayer. It mirrors the sibling
// tools/arm (the DS's ARMv5TE/ARMv4T cores) but targets the older, simpler ARM60
// and, critically, runs BIG-endian: the 3DO wires the ARM in big-endian mode, so
// both instructions and data are most-significant-byte first.
//
// The ARM60 is an ARMv3 core. Versus the ARMv4T/ARMv5TE that tools/arm models it
// lacks: Thumb, the BX/BLX interworking branches, the halfword/signed loads
// (LDRH/LDRSB/LDRSH), CLZ, the DSP saturating and signed-multiply instructions,
// and BKPT. The cond==1111 encoding is the old "never" condition (NV), not the
// ARMv5 unconditional space. What it has is the classic set: the conditional
// data-processing group with the barrel shifter, MRS/MSR, MUL/MLA (and the
// ARMv3M long multiplies), SWP, LDR/STR, LDM/STM, B/BL, SWI, and coprocessor
// transfers. This package models the 32-bit programmer's mode (address mode 32,
// which the 3DO uses — confirmed from the game's AIF header); the ARMv3 26-bit
// modes are not modelled.
package arm60

import "fmt"

// Flow classifies how an instruction affects control flow, for a recursive-descent
// tracer. Mirrors the categories in tools/arm (minus the Thumb interworking cases).
type Flow int

const (
	FlowSeq     Flow = iota // continues to the next instruction
	FlowBranch              // conditional branch: continues AND may go to Target
	FlowJump                // unconditional branch: goes to Target, no fall-through
	FlowCall                // BL: goes to Target, returns after
	FlowReturn              // return (MOV pc,lr; LDM {…,pc}): path ends
	FlowIndJump             // computed jump (MOV pc,Rm; LDR pc,…): target unknown
	FlowStop                // undefined/unmodelled: treat as data/stop
)

// Inst is one decoded 32-bit ARM instruction.
type Inst struct {
	Addr      uint32
	Len       int
	Mnem      string
	Text      string
	Flow      Flow
	Target    uint32
	HasTarget bool
	Cond      int
}

// Condition codes (bits 31:28).
const (
	condEQ = iota
	condNE
	condCS
	condCC
	condMI
	condPL
	condVS
	condVC
	condHI
	condLS
	condGE
	condLT
	condGT
	condLE
	condAL // always
	condNV // never (ARMv3): the instruction is not executed
)

var condName = [16]string{
	"EQ", "NE", "CS", "CC", "MI", "PL", "VS", "VC",
	"HI", "LS", "GE", "LT", "GT", "LE", "", "NV",
}

var regName = [16]string{
	"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7",
	"r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc",
}

var dpOps = [16]string{
	"AND", "EOR", "SUB", "RSB", "ADD", "ADC", "SBC", "RSC",
	"TST", "TEQ", "CMP", "CMN", "ORR", "MOV", "BIC", "MVN",
}

var shiftName = [4]string{"LSL", "LSR", "ASR", "ROR"}

func signExtend(v uint32, n uint) uint32 {
	m := uint32(1) << (n - 1)
	return (v ^ m) - m
}

func ror32(v, n uint32) uint32 {
	n &= 31
	return v>>n | v<<(32-n)
}

func cn(cond int) string { return condName[cond] }

// word reads a big-endian 32-bit word from code.
func word(code []byte) (uint32, bool) {
	if len(code) < 4 {
		return 0, false
	}
	return uint32(code[0])<<24 | uint32(code[1])<<16 | uint32(code[2])<<8 | uint32(code[3]), true
}

// Decode decodes one ARM instruction at addr (code[0] is the first byte).
func Decode(code []byte, addr uint32) Inst {
	w, ok := word(code)
	if !ok {
		return Inst{Addr: addr, Len: len(code), Mnem: ".word", Text: ".word ; truncated", Flow: FlowStop, Cond: condAL}
	}
	cond := int(w >> 28)
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq, Cond: cond}

	switch (w >> 25) & 7 { // bits 27:25
	case 0b000:
		in = decodeDataMisc(w, in, false)
	case 0b001:
		in = decodeDataMisc(w, in, true)
	case 0b010:
		in = decodeSingle(w, in, false)
	case 0b011:
		if (w>>4)&1 == 1 { // register-shift LDR/STR would set bit 4 — undefined on ARMv3
			in = undef(w, in)
		} else {
			in = decodeSingle(w, in, true)
		}
	case 0b100:
		in = decodeBlock(w, in)
	case 0b101:
		in = decodeBranch(w, addr, in)
	case 0b110:
		in = decodeCoproLS(w, in)
	default: // 0b111
		if (w>>24)&1 == 1 {
			in.Mnem = "SWI" + cn(cond)
			in.Text = fmt.Sprintf("%s #0x%X", in.Mnem, w&0xFFFFFF)
		} else {
			in = decodeCopro(w, in)
		}
	}

	// A "never" instruction is architecturally a no-op: keep the disassembly text
	// but neutralize its control flow so the tracer does not follow a dead branch.
	if cond == condNV && in.Flow != FlowStop {
		in.Flow, in.HasTarget = FlowSeq, false
	}
	return in
}

func undef(w uint32, in Inst) Inst {
	in.Mnem, in.Flow = ".word", FlowStop
	in.Text = fmt.Sprintf(".word 0x%08X", w)
	return in
}

// decodeDataMisc handles the 000/001 groups: data processing, the PSR-transfer
// space (MRS/MSR, carved from the S==0 TST/TEQ/CMP/CMN slot), and the multiply/
// swap extension space (bits 7 and 4 both set).
func decodeDataMisc(w uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF
	s := (w >> 20) & 1
	if !immForm && (w>>4)&1 == 1 && (w>>7)&1 == 1 {
		return decodeExtension(w, in)
	}
	if s == 0 && op >= 0b1000 && op <= 0b1011 {
		return decodeMisc(w, in, immForm)
	}
	return decodeDataProc(w, in, immForm)
}

func decodeDataProc(w uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF
	s := (w >> 20) & 1
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF

	var op2 string
	if immForm {
		op2 = imm(ror32(w&0xFF, ((w>>8)&0xF)*2))
	} else {
		op2 = shiftOperand(w)
	}
	sfx := ""
	if s == 1 {
		sfx = "S"
	}
	in.Mnem = dpOps[op] + cn(in.Cond) + sfx

	switch op {
	case 0b1000, 0b1001, 0b1010, 0b1011: // TST/TEQ/CMP/CMN
		in.Mnem = dpOps[op] + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[rn], op2)
	case 0b1101, 0b1111: // MOV/MVN
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[rd], op2)
	default:
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rn], op2)
	}

	if rd == 15 && !(op >= 0b1000 && op <= 0b1011) {
		rm := w & 0xF
		isReg := (w>>25)&1 == 0 && (w>>4)&0xFF == 0
		if op == 0b1101 && isReg && rm == 14 { // MOV pc, lr
			in.Flow = FlowReturn
		} else {
			in.Flow = FlowIndJump
		}
	}
	return in
}

func imm(v uint32) string {
	if v < 10 {
		return fmt.Sprintf("#%d", v)
	}
	return fmt.Sprintf("#0x%X", v)
}

func shiftOperand(w uint32) string {
	rm := w & 0xF
	styp := (w >> 5) & 3
	if (w>>4)&1 == 1 {
		rs := (w >> 8) & 0xF
		return fmt.Sprintf("%s, %s %s", regName[rm], shiftName[styp], regName[rs])
	}
	amt := (w >> 7) & 0x1F
	switch styp {
	case 0:
		if amt == 0 {
			return regName[rm]
		}
		return fmt.Sprintf("%s, LSL #%d", regName[rm], amt)
	case 1:
		if amt == 0 {
			amt = 32
		}
		return fmt.Sprintf("%s, LSR #%d", regName[rm], amt)
	case 2:
		if amt == 0 {
			amt = 32
		}
		return fmt.Sprintf("%s, ASR #%d", regName[rm], amt)
	default:
		if amt == 0 {
			return fmt.Sprintf("%s, RRX", regName[rm])
		}
		return fmt.Sprintf("%s, ROR #%d", regName[rm], amt)
	}
}

// decodeMisc handles the MRS/MSR PSR-transfer space. On ARMv3 the BX/CLZ/BKPT/
// saturating encodings that ARMv4+ place here do not exist and decode as data.
func decodeMisc(w uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF
	if immForm {
		return decodeMSRimm(w, in)
	}
	if (w>>4)&0xF == 0 { // MRS / MSR (register)
		if op == 0b1000 || op == 0b1010 { // MRS
			psr := "CPSR"
			if (w>>22)&1 == 1 {
				psr = "SPSR"
			}
			in.Mnem = "MRS" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[(w>>12)&0xF], psr)
			return in
		}
		return decodeMSRreg(w, in)
	}
	return undef(w, in)
}

func decodeMSRreg(w uint32, in Inst) Inst {
	psr := "CPSR"
	if (w>>22)&1 == 1 {
		psr = "SPSR"
	}
	in.Mnem = "MSR" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s %s%s, %s", in.Mnem, psr, msrFields(w), regName[w&0xF])
	return in
}

func decodeMSRimm(w uint32, in Inst) Inst {
	psr := "CPSR"
	if (w>>22)&1 == 1 {
		psr = "SPSR"
	}
	in.Mnem = "MSR" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s %s%s, %s", in.Mnem, psr, msrFields(w), imm(ror32(w&0xFF, ((w>>8)&0xF)*2)))
	return in
}

func msrFields(w uint32) string {
	m := (w >> 16) & 0xF
	if m == 0 {
		return ""
	}
	s := "_"
	for i, f := range []string{"c", "x", "s", "f"} {
		if m&(1<<uint(i)) != 0 {
			s += f
		}
	}
	return s
}

// decodeExtension handles the bits-7-and-4 space: MUL/MLA, the long multiplies,
// and SWP. Halfword/signed loads (sh != 0) do not exist on ARMv3.
func decodeExtension(w uint32, in Inst) Inst {
	if (w>>5)&3 != 0 { // halfword/signed load-store space — undefined on ARMv3
		return undef(w, in)
	}
	switch {
	case (w>>24)&1 == 1: // SWP/SWPB
		b := ""
		if (w>>22)&1 == 1 {
			b = "B"
		}
		in.Mnem = "SWP" + cn(in.Cond) + b
		in.Text = fmt.Sprintf("%s %s, %s, [%s]", in.Mnem, regName[(w>>12)&0xF], regName[w&0xF], regName[(w>>16)&0xF])
		return in
	case (w>>23)&1 == 1: // long multiply (ARMv3M)
		names := [4]string{"UMULL", "UMLAL", "SMULL", "SMLAL"}
		s := ""
		if (w>>20)&1 == 1 {
			s = "S"
		}
		in.Mnem = names[(w>>21)&3] + cn(in.Cond) + s
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem,
			regName[(w>>12)&0xF], regName[(w>>16)&0xF], regName[w&0xF], regName[(w>>8)&0xF])
		return in
	default: // MUL / MLA
		s := ""
		if (w>>20)&1 == 1 {
			s = "S"
		}
		rd := (w >> 16) & 0xF
		rn := (w >> 12) & 0xF
		rs := (w >> 8) & 0xF
		rm := w & 0xF
		if (w>>21)&1 == 1 {
			in.Mnem = "MLA" + cn(in.Cond) + s
			in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs], regName[rn])
		} else {
			in.Mnem = "MUL" + cn(in.Cond) + s
			in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs])
		}
		return in
	}
}

func decodeSingle(w uint32, in Inst, regOff bool) Inst {
	l := (w >> 20) & 1
	b := (w >> 22) & 1
	name := "STR"
	if l == 1 {
		name = "LDR"
	}
	if b == 1 {
		name += "B"
	}
	in.Mnem = name + cn(in.Cond)

	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	sign := ""
	if u == 0 {
		sign = "-"
	}
	var off string
	if !regOff {
		if v := w & 0xFFF; v != 0 {
			off = fmt.Sprintf(", #%s0x%X", sign, v)
		}
	} else {
		off = ", " + sign + shiftOperand(w)
	}
	in.Text = addrForm(in.Mnem, rd, rn, off, p, (w>>21)&1)
	if l == 1 && rd == 15 {
		in.Flow = FlowIndJump
	}
	return in
}

func addrForm(mnem string, rd, rn uint32, off string, p, wb uint32) string {
	if p == 1 {
		bang := ""
		if wb == 1 {
			bang = "!"
		}
		return fmt.Sprintf("%s %s, [%s%s]%s", mnem, regName[rd], regName[rn], off, bang)
	}
	return fmt.Sprintf("%s %s, [%s]%s", mnem, regName[rd], regName[rn], off)
}

func decodeBlock(w uint32, in Inst) Inst {
	l := (w >> 20) & 1
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	base := "STM"
	if l == 1 {
		base = "LDM"
	}
	mode := [4]string{"DA", "DB", "IA", "IB"}[u<<1|p]
	in.Mnem = base + cn(in.Cond) + mode

	rn := (w >> 16) & 0xF
	bang := ""
	if (w>>21)&1 == 1 {
		bang = "!"
	}
	usr := ""
	if (w>>22)&1 == 1 {
		usr = "^"
	}
	in.Text = fmt.Sprintf("%s %s%s, {%s}%s", in.Mnem, regName[rn], bang, regList(w&0xFFFF), usr)
	if l == 1 && w&(1<<15) != 0 {
		in.Flow = FlowReturn
	}
	return in
}

func regList(mask uint32) string {
	var parts []string
	for i := 0; i < 16; {
		if mask&(1<<uint(i)) == 0 {
			i++
			continue
		}
		j := i
		for j+1 < 16 && mask&(1<<uint(j+1)) != 0 {
			j++
		}
		switch {
		case j == i:
			parts = append(parts, regName[i])
		case j == i+1:
			parts = append(parts, regName[i], regName[j])
		default:
			parts = append(parts, regName[i]+"-"+regName[j])
		}
		i = j + 1
	}
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

func decodeBranch(w, addr uint32, in Inst) Inst {
	off := signExtend(w&0xFFFFFF, 24) << 2
	target := addr + 8 + off
	in.Target, in.HasTarget = target, true
	if (w>>24)&1 == 1 { // BL
		in.Mnem = "BL" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s 0x%08X", in.Mnem, target)
		in.Flow = FlowCall
		return in
	}
	in.Mnem = "B" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s 0x%08X", in.Mnem, target)
	if in.Cond == condAL || in.Cond == condNV {
		in.Flow = FlowJump
	} else {
		in.Flow = FlowBranch
	}
	return in
}

func decodeCoproLS(w uint32, in Inst) Inst {
	l := (w >> 20) & 1
	name := "STC"
	if l == 1 {
		name = "LDC"
	}
	in.Mnem = name + cn(in.Cond)
	in.Text = fmt.Sprintf("%s p%d, c%d, [%s]", in.Mnem, (w>>8)&0xF, (w>>12)&0xF, regName[(w>>16)&0xF])
	return in
}

func decodeCopro(w uint32, in Inst) Inst {
	cp := (w >> 8) & 0xF
	if (w>>4)&1 == 0 { // CDP
		in.Mnem = "CDP" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s p%d, #%d, c%d, c%d, c%d", in.Mnem, cp, (w>>20)&0xF, (w>>12)&0xF, (w>>16)&0xF, w&0xF)
		return in
	}
	name := "MCR"
	if (w>>20)&1 == 1 {
		name = "MRC"
	}
	in.Mnem = name + cn(in.Cond)
	in.Text = fmt.Sprintf("%s p%d, #%d, %s, c%d, c%d, #%d", in.Mnem, cp, (w>>21)&7, regName[(w>>12)&0xF], (w>>16)&0xF, w&0xF, (w>>5)&7)
	return in
}

// Disassemble decodes a run of ARM code starting at base into assembler lines.
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for off := 0; off+4 <= len(code); off += 4 {
		in := Decode(code[off:], base+uint32(off))
		out = append(out, fmt.Sprintf("%08X  %s", in.Addr, in.Text))
	}
	return out
}
