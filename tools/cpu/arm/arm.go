// Package arm is a disassembler, recursive-descent code-tracer front-end and
// execution core for the two ARM CPUs in the Nintendo DS, mirroring the mos6502,
// z80, sm83 and m68k packages in this repository. It operates on raw ARM machine
// code and makes no assumptions about where that code came from (the DS memory
// map — main RAM, the two TCMs, the shared WRAM, cartridge, BIOS — is the caller's
// concern, supplied through the Bus in cpu.go).
//
// The DS has two ARM cores:
//
//   - ARM9 — an ARM946E-S, ARMv5TE. The main CPU (67 MHz). Adds over ARMv4T the
//     BLX interworking branches, CLZ, the DSP saturating (QADD/QSUB…) and signed
//     multiply (SMLAxy…) instructions, and BKPT. Has caches, tightly-coupled
//     memories (ITCM/DTCM) and an MPU (CP15), not a full MMU.
//   - ARM7 — an ARM7TDMI, ARMv4T. The secondary CPU (33 MHz), driving sound, wifi,
//     the touchscreen and other I/O.
//
// Both implement two instruction sets that this package decodes and executes:
//
//   - ARM — fixed 32-bit instructions, every one carrying a 4-bit condition field
//     so almost any instruction can be made conditional; operand 2 of the
//     data-processing group runs through a barrel shifter.
//   - Thumb — 16-bit instructions (thumb.go), a compressed re-encoding of a subset
//     of ARM, used heavily to save cartridge and RAM space.
//
// The processor switches between the two at run time (the CPSR T bit), reached via
// the BX/BLX interworking branches — bit 0 of the target address selects the state.
// Because of that a caller (and the codetracearm worklist) must track which state
// each address is decoded in; Decode takes that state as a parameter and each Inst
// records it (Inst.Thumb) and the state its branch target lands in (Inst.TargetThumb).
package arm

import "fmt"

// Flow classifies how an instruction affects control flow — the information a
// recursive-descent disassembler needs to follow every reachable path. The first
// seven categories mirror the mos6502/m68k/z80/sm83 packages; FlowIndCall is an
// ARM-specific eighth: unlike those CPUs, ARM has a common *indirect call* form
// (BLX Rm, and calls through a function pointer) whose target is not statically
// known yet which returns and so must NOT end the trace of the fall-through path.
type Flow int

const (
	FlowSeq     Flow = iota // continues to the next instruction
	FlowBranch              // conditional branch (B<cc>): continues AND may go to Target
	FlowJump                // unconditional branch (B / B AL): goes to Target, no fall-through
	FlowCall                // BL / BLX <imm>: goes to Target, normally returns after it
	FlowReturn              // return (BX lr, MOV pc,lr, POP/LDM {…,pc}): path ends
	FlowIndJump             // computed jump (BX Rm, MOV pc,Rm, LDR pc,…): target not statically known
	FlowStop                // undefined/unmodelled opcode or a truncated read: treat as data/stop
	FlowIndCall             // indirect call (BLX Rm, call through a pointer): target unknown, but returns
)

// Inst is one decoded instruction (one ARM word, or one Thumb halfword, or the
// two-halfword Thumb BL/BLX pair).
type Inst struct {
	Addr        uint32
	Len         int    // 4 for ARM, 2 for Thumb (4 for the Thumb BL/BLX pair)
	Mnem        string // the bare mnemonic including any condition/size suffix ("ADDNE", "LDRB", …)
	Text        string // formatted "MNEM operands"
	Flow        Flow
	Target      uint32 // branch/call destination (when HasTarget)
	HasTarget   bool
	Thumb       bool // decoded in Thumb state
	TargetThumb bool // the ISA state (Thumb vs ARM) execution lands in at Target — for interworking
	Cond        int  // the 4-bit condition field (condAL == unconditional); informational
}

// Condition codes (bits 31:28 of every ARM instruction).
const (
	condEQ = iota
	condNE
	condCS // == HS
	condCC // == LO
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
	condAL // always (no suffix)
	condNV // never / ARMv5 unconditional-extension space (bits 31:28 == 1111)
)

// condName is the mnemonic suffix for each condition (condAL is empty).
var condName = [16]string{
	"EQ", "NE", "CS", "CC", "MI", "PL", "VS", "VC",
	"HI", "LS", "GE", "LT", "GT", "LE", "", "NV",
}

// regName maps register numbers to their assembler names (SP/LR/PC aliases).
var regName = [16]string{
	"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7",
	"r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc",
}

// dpOps are the 16 data-processing operations, indexed by bits 24:21.
var dpOps = [16]string{
	"AND", "EOR", "SUB", "RSB", "ADD", "ADC", "SBC", "RSC",
	"TST", "TEQ", "CMP", "CMN", "ORR", "MOV", "BIC", "MVN",
}

var shiftName = [4]string{"LSL", "LSR", "ASR", "ROR"}

// signExtend sign-extends the low n bits of v to 32 bits.
func signExtend(v uint32, n uint) uint32 {
	m := uint32(1) << (n - 1)
	return (v ^ m) - m
}

// imm formats an immediate operand: small values decimal, larger ones hex, for
// readable listings.
func imm(v uint32) string {
	if v < 10 {
		return fmt.Sprintf("#%d", v)
	}
	return fmt.Sprintf("#0x%X", v)
}

// Decode decodes one instruction at the start of code (code[0] is the first byte),
// which is loaded at address addr, interpreting it in Thumb state when thumb is
// true and ARM state otherwise. A truncated read or an unmodelled encoding decodes
// as data with FlowStop.
func Decode(code []byte, addr uint32, thumb bool) Inst {
	if thumb {
		return DecodeThumb(code, addr)
	}
	return DecodeARM(code, addr)
}

// word reads a little-endian 32-bit word from code, reporting whether all four
// bytes were present.
func word(code []byte) (uint32, bool) {
	if len(code) < 4 {
		return 0, false
	}
	return uint32(code[0]) | uint32(code[1])<<8 | uint32(code[2])<<16 | uint32(code[3])<<24, true
}

// DecodeARM decodes one 32-bit ARM instruction at addr.
func DecodeARM(code []byte, addr uint32) Inst {
	w, ok := word(code)
	if !ok {
		return Inst{Addr: addr, Len: len(code), Mnem: ".word", Text: ".word ; truncated", Flow: FlowStop, Cond: condAL}
	}
	cond := int(w >> 28)
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq, Cond: cond}

	if cond == condNV {
		return decodeUncond(w, addr, in)
	}

	switch (w >> 25) & 7 { // bits 27:25
	case 0b000:
		return decodeDataMisc(w, addr, in, false)
	case 0b001:
		return decodeDataMisc(w, addr, in, true)
	case 0b010: // LDR/STR, immediate offset
		return decodeSingle(w, in, false)
	case 0b011: // LDR/STR, register offset (bit4==0); bit4==1 is the media space (undefined on ARMv5)
		if (w>>4)&1 == 1 {
			return undef(w, in)
		}
		return decodeSingle(w, in, true)
	case 0b100: // LDM/STM
		return decodeBlock(w, in)
	case 0b101: // B / BL
		return decodeBranch(w, addr, in)
	case 0b110: // LDC/STC coprocessor transfer
		return decodeCoproLS(w, in)
	default: // 0b111: CDP / MCR / MRC, or SWI
		if (w>>24)&1 == 1 {
			in.Mnem = "SWI" + cn(cond)
			in.Text = fmt.Sprintf("%s #0x%X", in.Mnem, w&0xFFFFFF)
			// A software interrupt traps to the BIOS and returns; keep tracing after it.
			return in
		}
		return decodeCopro(w, in)
	}
}

// cn returns the condition suffix for a condition code (empty for AL).
func cn(cond int) string { return condName[cond] }

func undef(w uint32, in Inst) Inst {
	in.Mnem, in.Flow = ".word", FlowStop
	in.Text = fmt.Sprintf(".word 0x%08X", w)
	return in
}

// decodeDataMisc handles bits 27:25 == 000 (register/shift forms and the
// multiply/swap/halfword extension space) and 001 (immediate forms). The
// TST/TEQ/CMP/CMN opcodes with S==0 do not exist as data-processing (they would
// compute nothing) so that slot is reused for the miscellaneous instructions
// (MRS/MSR/BX/BLX/CLZ/BKPT and the DSP ops).
func decodeDataMisc(w, addr uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF
	s := (w >> 20) & 1

	if !immForm && (w>>4)&1 == 1 && (w>>7)&1 == 1 {
		// bits 7 and 4 both set: multiply / swap / long-multiply / halfword transfer.
		return decodeExtension(w, in)
	}
	if s == 0 && op >= 0b1000 && op <= 0b1011 {
		// TST/TEQ/CMP/CMN with S clear → miscellaneous / PSR-transfer / DSP.
		return decodeMisc(w, in, immForm)
	}
	return decodeDataProc(w, addr, in, immForm)
}

// decodeDataProc formats a data-processing instruction (AND…MVN).
func decodeDataProc(w, addr uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF
	s := (w >> 20) & 1
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF

	var op2 string
	if immForm {
		v := ror32(w&0xFF, ((w>>8)&0xF)*2)
		op2 = imm(v)
	} else {
		op2 = shiftOperand(w)
	}

	sfx := ""
	if s == 1 {
		sfx = "S"
	}
	name := dpOps[op]
	in.Mnem = name + cn(in.Cond) + sfx

	switch op {
	case 0b1000, 0b1001, 0b1010, 0b1011: // TST/TEQ/CMP/CMN — no destination, always set flags
		in.Mnem = dpOps[op] + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[rn], op2)
	case 0b1101, 0b1111: // MOV/MVN — no first operand
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[rd], op2)
	default:
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rn], op2)
	}

	// A data-processing op that writes the PC is a computed control transfer.
	if rd == 15 && op != 0b1000 && op != 0b1001 && op != 0b1010 && op != 0b1011 {
		classifyPCWrite(&in, w, op)
	}
	return in
}

// classifyPCWrite marks a data-processing instruction that targets r15 as a jump
// or a return. "MOV pc, lr" (and "MOVS pc, lr") is the canonical function return.
func classifyPCWrite(in *Inst, w, op uint32) {
	rm := w & 0xF
	isReg := (w>>25)&1 == 0 && (w>>4)&0xFF == 0 // plain "MOV pc, Rm"
	if op == 0b1101 && isReg && rm == 14 {
		in.Flow = FlowReturn
		return
	}
	in.Flow = FlowIndJump
}

// shiftOperand formats operand 2 in its register/barrel-shifter form.
func shiftOperand(w uint32) string {
	rm := w & 0xF
	styp := (w >> 5) & 3
	if (w>>4)&1 == 1 { // register-specified shift amount
		rs := (w >> 8) & 0xF
		return fmt.Sprintf("%s, %s %s", regName[rm], shiftName[styp], regName[rs])
	}
	amt := (w >> 7) & 0x1F
	switch styp {
	case 0: // LSL
		if amt == 0 {
			return regName[rm] // "Rm"
		}
		return fmt.Sprintf("%s, LSL #%d", regName[rm], amt)
	case 1: // LSR (a #0 encodes #32)
		if amt == 0 {
			amt = 32
		}
		return fmt.Sprintf("%s, LSR #%d", regName[rm], amt)
	case 2: // ASR (a #0 encodes #32)
		if amt == 0 {
			amt = 32
		}
		return fmt.Sprintf("%s, ASR #%d", regName[rm], amt)
	default: // ROR (a #0 encodes RRX, a rotate-right-through-carry by one)
		if amt == 0 {
			return fmt.Sprintf("%s, RRX", regName[rm])
		}
		return fmt.Sprintf("%s, ROR #%d", regName[rm], amt)
	}
}

// ror32 rotates v right by n (mod 32).
func ror32(v, n uint32) uint32 {
	n &= 31
	return v>>n | v<<(32-n)
}

// decodeMisc handles the miscellaneous space carved out of the TST/TEQ/CMP/CMN
// (S==0) data-processing slot: MRS/MSR, BX/BLX/BXJ, CLZ, BKPT, the saturating
// arithmetic (QADD…) and the signed multiplies (SMLAxy…).
func decodeMisc(w uint32, in Inst, immForm bool) Inst {
	op := (w >> 21) & 0xF

	if immForm {
		// MSR immediate (into CPSR/SPSR flags).
		return decodeMSRimm(w, in)
	}
	if (w>>7)&1 == 1 {
		// bit7 set, bit4 clear here → DSP signed multiplies (SMLAxy family).
		return decodeSignedMul(w, in)
	}

	op2 := (w >> 4) & 0xF
	switch op2 {
	case 0b0000: // MRS / MSR (register)
		if op == 0b1000 || op == 0b1010 { // TST/CMP slot → MRS
			psr := "CPSR"
			if (w>>22)&1 == 1 {
				psr = "SPSR"
			}
			in.Mnem = "MRS" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[(w>>12)&0xF], psr)
			return in
		}
		return decodeMSRreg(w, in)
	case 0b0001:
		switch op {
		case 0b1001: // BX
			rm := w & 0xF
			in.Mnem = "BX" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s", in.Mnem, regName[rm])
			if rm == 14 {
				in.Flow = FlowReturn
			} else {
				in.Flow = FlowIndJump
			}
			return in
		case 0b1011: // CLZ
			in.Mnem = "CLZ" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, regName[(w>>12)&0xF], regName[w&0xF])
			return in
		}
	case 0b0011: // BLX (register) — indirect call
		if op == 0b1001 {
			in.Mnem = "BLX" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s", in.Mnem, regName[w&0xF])
			in.Flow = FlowIndCall
			return in
		}
	case 0b0101: // saturating arithmetic
		sat := [4]string{"QADD", "QSUB", "QDADD", "QDSUB"}[(w>>21)&3]
		in.Mnem = sat + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[(w>>12)&0xF], regName[w&0xF], regName[(w>>16)&0xF])
		return in
	case 0b0111: // BKPT
		if op == 0b1010 {
			in.Mnem = "BKPT"
			imm16 := (w>>4)&0xFFF0 | (w & 0xF)
			in.Text = fmt.Sprintf("BKPT #0x%X", imm16)
			in.Flow = FlowStop
			return in
		}
	}
	return undef(w, in)
}

// decodeMSRreg formats "MSR CPSR_fields, Rm".
func decodeMSRreg(w uint32, in Inst) Inst {
	psr := "CPSR"
	if (w>>22)&1 == 1 {
		psr = "SPSR"
	}
	in.Mnem = "MSR" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s %s%s, %s", in.Mnem, psr, msrFields(w), regName[w&0xF])
	return in
}

// decodeMSRimm formats "MSR CPSR_fields, #imm".
func decodeMSRimm(w uint32, in Inst) Inst {
	psr := "CPSR"
	if (w>>22)&1 == 1 {
		psr = "SPSR"
	}
	v := ror32(w&0xFF, ((w>>8)&0xF)*2)
	in.Mnem = "MSR" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s %s%s, %s", in.Mnem, psr, msrFields(w), imm(v))
	return in
}

// msrFields renders the "_cxsf" field mask suffix of an MSR.
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

// decodeSignedMul formats the ARMv5TE signed multiplies (SMLAxy/SMULxy/SMLAWy/
// SMULWy/SMLALxy).
func decodeSignedMul(w uint32, in Inst) Inst {
	op := (w >> 21) & 0xF
	x := (w >> 5) & 1
	y := (w >> 6) & 1
	bt := func(b uint32) byte {
		if b == 0 {
			return 'B'
		}
		return 'T'
	}
	rd := (w >> 16) & 0xF
	rn := (w >> 12) & 0xF
	rs := (w >> 8) & 0xF
	rm := w & 0xF
	switch op {
	case 0b1000: // SMLA<x><y>
		in.Mnem = fmt.Sprintf("SMLA%c%c%s", bt(x), bt(y), cn(in.Cond))
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs], regName[rn])
	case 0b1001: // SMLAW<y> / SMULW<y>
		if x == 0 {
			in.Mnem = fmt.Sprintf("SMLAW%c%s", bt(y), cn(in.Cond))
			in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs], regName[rn])
		} else {
			in.Mnem = fmt.Sprintf("SMULW%c%s", bt(y), cn(in.Cond))
			in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs])
		}
	case 0b1010: // SMLAL<x><y>
		in.Mnem = fmt.Sprintf("SMLAL%c%c%s", bt(x), bt(y), cn(in.Cond))
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[rn], regName[rd], regName[rm], regName[rs])
	default: // 0b1011: SMUL<x><y>
		in.Mnem = fmt.Sprintf("SMUL%c%c%s", bt(x), bt(y), cn(in.Cond))
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs])
	}
	return in
}

// decodeExtension handles the bit7==1 && bit4==1 space in the 000 group: the
// multiply group (MUL/MLA, the long multiplies, SWP) and the halfword/signed
// byte load-store group.
func decodeExtension(w uint32, in Inst) Inst {
	sh := (w >> 5) & 3
	if sh == 0 { // multiply / swap
		switch {
		case (w>>24)&1 == 1: // SWP/SWPB
			b := ""
			if (w>>22)&1 == 1 {
				b = "B"
			}
			in.Mnem = "SWP" + cn(in.Cond) + b
			in.Text = fmt.Sprintf("%s %s, %s, [%s]", in.Mnem, regName[(w>>12)&0xF], regName[w&0xF], regName[(w>>16)&0xF])
			return in
		case (w>>23)&1 == 1: // long multiply
			return decodeMulLong(w, in)
		default: // MUL / MLA
			return decodeMul(w, in)
		}
	}
	// Halfword / signed load-store.
	return decodeHalf(w, in, sh)
}

// decodeMul formats MUL/MLA.
func decodeMul(w uint32, in Inst) Inst {
	s := ""
	if (w>>20)&1 == 1 {
		s = "S"
	}
	rd := (w >> 16) & 0xF
	rn := (w >> 12) & 0xF
	rs := (w >> 8) & 0xF
	rm := w & 0xF
	if (w>>21)&1 == 1 { // MLA
		in.Mnem = "MLA" + cn(in.Cond) + s
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs], regName[rn])
	} else { // MUL
		in.Mnem = "MUL" + cn(in.Cond) + s
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, regName[rd], regName[rm], regName[rs])
	}
	return in
}

// decodeMulLong formats UMULL/UMLAL/SMULL/SMLAL.
func decodeMulLong(w uint32, in Inst) Inst {
	names := [4]string{"UMULL", "UMLAL", "SMULL", "SMLAL"} // by bits 22:21 (U, A)
	s := ""
	if (w>>20)&1 == 1 {
		s = "S"
	}
	in.Mnem = names[(w>>21)&3] + cn(in.Cond) + s
	in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem,
		regName[(w>>12)&0xF], regName[(w>>16)&0xF], regName[w&0xF], regName[(w>>8)&0xF])
	return in
}

// decodeHalf formats the halfword / signed-byte load-store group (LDRH/STRH/
// LDRSB/LDRSH), with its immediate or register offset and pre/post-index modes.
func decodeHalf(w uint32, in Inst, sh uint32) Inst {
	l := (w >> 20) & 1
	name := ""
	switch {
	case l == 0 && sh == 1:
		name = "STRH"
	case l == 1 && sh == 1:
		name = "LDRH"
	case l == 1 && sh == 2:
		name = "LDRSB"
	case l == 1 && sh == 3:
		name = "LDRSH"
	default: // STRD/LDRD (ARMv5TE) at sh==2/3, L==0 — model as their names
		if sh == 2 {
			name = "LDRD"
		} else {
			name = "STRD"
		}
	}
	in.Mnem = name + cn(in.Cond)
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	sign := "+"
	if u == 0 {
		sign = "-"
	}
	var off string
	if (w>>22)&1 == 1 { // immediate offset
		v := (w>>4)&0xF0 | (w & 0xF)
		if v == 0 {
			off = ""
		} else {
			off = fmt.Sprintf(", #%s0x%X", plusMinus(sign), v)
		}
	} else { // register offset
		off = fmt.Sprintf(", %s%s", plusMinus(sign), regName[w&0xF])
	}
	in.Text = addrForm(in.Mnem, rd, rn, off, p, (w>>21)&1)
	return in
}

// plusMinus renders "-" as "-" and "+" as "" (positive offsets are implicit).
func plusMinus(s string) string {
	if s == "-" {
		return "-"
	}
	return ""
}

// decodeSingle formats the single data-transfer group (LDR/STR/LDRB/STRB).
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
		v := w & 0xFFF
		if v == 0 {
			off = ""
		} else {
			off = fmt.Sprintf(", #%s0x%X", sign, v)
		}
	} else {
		off = ", " + sign + shiftOperand(w)
	}
	in.Text = addrForm(in.Mnem, rd, rn, off, p, (w>>21)&1)

	// LDR into the PC is a computed jump / return (e.g. "LDR pc, [pc, #x]" dispatch,
	// or "LDR pc, [sp], #4" style returns). Target isn't statically known.
	if l == 1 && rd == 15 {
		in.Flow = FlowIndJump
	}
	return in
}

// addrForm renders the "Rd, [Rn ...]" addressing operand for load/store, honouring
// pre-index (p==1) with optional write-back (wb), and post-index (p==0).
func addrForm(mnem string, rd, rn uint32, off string, p, wb uint32) string {
	if p == 1 { // pre-indexed: [Rn, off]{!}
		bang := ""
		if wb == 1 {
			bang = "!"
		}
		return fmt.Sprintf("%s %s, [%s%s]%s", mnem, regName[rd], regName[rn], off, bang)
	}
	// post-indexed: [Rn], off
	return fmt.Sprintf("%s %s, [%s]%s", mnem, regName[rd], regName[rn], off)
}

// decodeBlock formats the block data-transfer group (LDM/STM, i.e. push/pop and
// context save/restore). Loading the PC (r15 in the register list) ends the path
// as a return — the standard function epilogue "LDMFD sp!, {…, pc}".
func decodeBlock(w uint32, in Inst) Inst {
	l := (w >> 20) & 1
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	base := "STM"
	if l == 1 {
		base = "LDM"
	}
	// Addressing-mode suffix: IA/IB/DA/DB (u,p).
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

	if l == 1 && w&(1<<15) != 0 { // r15 loaded → return
		in.Flow = FlowReturn
	}
	return in
}

// regList renders a 16-bit register bitmap as "r0-r3, r5, lr" style ranges.
func regList(mask uint32) string {
	var parts []string
	i := 0
	for i < 16 {
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
	return joinComma(parts)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

// decodeBranch formats B / BL. The target is PC-relative, with the ARM pipeline's
// +8 read of the PC folded in.
func decodeBranch(w, addr uint32, in Inst) Inst {
	off := signExtend(w&0xFFFFFF, 24) << 2
	target := addr + 8 + off
	in.Target, in.HasTarget, in.TargetThumb = target, true, false
	if (w>>24)&1 == 1 { // BL — a call (returns), whether or not it is conditional
		in.Mnem = "BL" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s 0x%08X", in.Mnem, target)
		in.Flow = FlowCall
		return in
	}
	in.Mnem = "B" + cn(in.Cond)
	in.Text = fmt.Sprintf("%s 0x%08X", in.Mnem, target)
	if in.Cond == condAL {
		in.Flow = FlowJump
	} else {
		in.Flow = FlowBranch
	}
	return in
}

// decodeUncond handles the cond==1111 space: on ARMv5 this holds BLX <imm> (an
// unconditional interworking call that switches to Thumb) and hint/cache ops such
// as PLD. Anything else here is treated as data.
func decodeUncond(w, addr uint32, in Inst) Inst {
	switch {
	case (w>>25)&7 == 0b101: // BLX <imm>
		h := (w >> 24) & 1
		off := signExtend(w&0xFFFFFF, 24)<<2 | h<<1
		target := addr + 8 + off
		in.Mnem = "BLX"
		in.Text = fmt.Sprintf("BLX 0x%08X", target)
		in.Flow, in.Target, in.HasTarget, in.TargetThumb = FlowCall, target, true, true // switches to Thumb
		return in
	case (w>>26)&3 == 0b01 && (w>>20)&0xF7 == 0xF5: // PLD [Rn, …] — a cache hint, no effect on flow
		in.Mnem = "PLD"
		in.Text = fmt.Sprintf("PLD [%s]", regName[(w>>16)&0xF])
		return in
	}
	return undef(w, in)
}

// decodeCoproLS formats coprocessor load/store (LDC/STC). Rare in DS game code
// (mostly floating-point emulation packages); modelled for completeness of the
// listing, with no control-flow effect.
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

// decodeCopro formats the coprocessor data operations and register transfers
// (CDP, and MCR/MRC — the CP15 system-control accesses the ARM9 uses to drive
// caches, the TCMs and the MPU).
func decodeCopro(w uint32, in Inst) Inst {
	cp := (w >> 8) & 0xF
	if (w>>4)&1 == 0 { // CDP
		in.Mnem = "CDP" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s p%d, #%d, c%d, c%d, c%d", in.Mnem, cp, (w>>20)&0xF, (w>>12)&0xF, (w>>16)&0xF, w&0xF)
		return in
	}
	// MCR / MRC.
	name := "MCR"
	if (w>>20)&1 == 1 {
		name = "MRC"
	}
	in.Mnem = name + cn(in.Cond)
	in.Text = fmt.Sprintf("%s p%d, #%d, %s, c%d, c%d, #%d", in.Mnem, cp, (w>>21)&7, regName[(w>>12)&0xF], (w>>16)&0xF, w&0xF, (w>>5)&7)
	return in
}
