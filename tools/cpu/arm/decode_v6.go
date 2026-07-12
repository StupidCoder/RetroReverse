package arm

import "fmt"

// decodeARMv6 handles the ARM instructions ARMv6K adds over ARMv5TE. It returns
// handled==false for anything that is not a v6 addition, so the caller falls
// through to the ARMv5TE decoder. The two aliasing cases (LDREX/STREX over SWP,
// UMAAL over MUL) are caught here first; the rest occupy space ARMv5TE leaves
// undefined.
func decodeARMv6(w, addr uint32, in Inst) (Inst, bool) {
	cond := int(w >> 28)

	// VFP coprocessor (cp 10/11): load/store and data-processing/transfer groups.
	// Handled for every condition (VFP instructions are conditional).
	if isVFP(w) && ((w>>25)&7 == 0b110 || (w>>24)&0xF == 0b1110) {
		return decodeVFP(w, in), true
	}

	// Unconditional space (cond == 1111): CPS, SETEND, CLREX. Anything else
	// there (BLX <imm>, PLD) belongs to the v5 decoder, so report not-handled.
	if cond == condNV {
		if out, ok := decodeV6Uncond(w, in); ok {
			return out, true
		}
		return in, false
	}

	// Synchronization primitives share the SWP slot: bits 27:24 == 0001,
	// bits 7:4 == 1001. Bit 23 selects LDREX/STREX (1) over SWP (0, left to v5).
	if (w>>24)&0xF == 0b0001 && (w>>4)&0xF == 0b1001 && (w>>23)&1 == 1 {
		return decodeSync(w, in), true
	}

	// UMAAL shares the multiply slot: bits 27:24 == 0000, bits 7:4 == 1001,
	// bits 23:20 == 0100. ARMv5TE would read this as MUL.
	if (w>>24)&0xF == 0b0000 && (w>>4)&0xF == 0b1001 && (w>>20)&0xF == 0b0100 {
		in.Mnem = "UMAAL" + cn(cond)
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem,
			regName[(w>>12)&0xF], regName[(w>>16)&0xF], regName[w&0xF], regName[(w>>8)&0xF])
		return in, true
	}

	// The media space: bits 27:25 == 011, bit 4 == 1.
	if (w>>25)&7 == 0b011 && (w>>4)&1 == 1 {
		return decodeMedia(w, in), true
	}
	return in, false
}

// decodeSync formats the load/store-exclusive family (word, doubleword, byte,
// halfword), selected by bits 22:21, with bit 20 the load flag.
func decodeSync(w uint32, in Inst) Inst {
	load := (w>>20)&1 == 1
	sz := (w >> 21) & 3 // 0 word, 1 doubleword, 2 byte, 3 halfword
	suffix := [4]string{"", "D", "B", "H"}[sz]
	rn := regName[(w>>16)&0xF]
	if load {
		in.Mnem = "LDREX" + suffix + cn(in.Cond)
		rt := (w >> 12) & 0xF
		if sz == 1 { // LDREXD Rt, Rt2, [Rn] — Rt2 = Rt+1, implicit in the encoding
			in.Text = fmt.Sprintf("%s %s, %s, [%s]", in.Mnem, regName[rt], regName[(rt+1)&0xF], rn)
		} else {
			in.Text = fmt.Sprintf("%s %s, [%s]", in.Mnem, regName[rt], rn)
		}
	} else {
		in.Mnem = "STREX" + suffix + cn(in.Cond)
		// STREX Rd, Rt, [Rn]: Rd (result) at 15:12, Rt (value) at 3:0.
		rt := w & 0xF
		if sz == 1 { // STREXD Rd, Rt, Rt2, [Rn]
			in.Text = fmt.Sprintf("%s %s, %s, %s, [%s]", in.Mnem,
				regName[(w>>12)&0xF], regName[rt], regName[(rt+1)&0xF], rn)
		} else {
			in.Text = fmt.Sprintf("%s %s, %s, [%s]", in.Mnem, regName[(w>>12)&0xF], regName[rt], rn)
		}
	}
	return in
}

// decodeMedia formats the ARMv6 media space (bits 27:25 == 011, bit 4 == 1):
// parallel arithmetic, packing/saturation/extension/reversal, and the signed
// dual multiplies.
func decodeMedia(w uint32, in Inst) Inst {
	op1 := (w >> 20) & 0x1F
	op2 := (w >> 5) & 7
	rd := regName[(w>>12)&0xF]
	rn := regName[(w>>16)&0xF]
	rm := regName[w&0xF]

	switch {
	case op1>>3 == 0b00 && op1&7 != 0: // parallel add/sub (op1 = 0b00xxx, xxx != 0)
		if name := parallelName(op1&7, op2); name != "" {
			in.Mnem = name + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rd, rn, rm)
			return in
		}
	case op1>>3 == 0b01: // packing, saturation, extension, reversal, select
		return decodeMediaPack(w, in)
	case op1>>3 == 0b10: // signed multiplies (SMLAD, SMUAD, SMLSD, SMMLA, …)
		return decodeMediaMul(w, in)
	case op1 == 0b11000: // USAD8 / USADA8
		ra := (w >> 12) & 0xF
		if ra == 15 {
			in.Mnem = "USAD8" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rn, rm, regName[(w>>8)&0xF])
		} else {
			in.Mnem = "USADA8" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, rn, rm, regName[(w>>8)&0xF], regName[ra])
		}
		return in
	}
	return undef(w, in)
}

// parallelName builds a parallel-arithmetic mnemonic from its signedness/mode
// class (op1&7) and operation (op2). Empty for a reserved combination.
func parallelName(class, op2 uint32) string {
	prefix := map[uint32]string{1: "S", 2: "Q", 3: "SH", 5: "U", 6: "UQ", 7: "UH"}[class]
	op := map[uint32]string{0: "ADD16", 1: "ASX", 2: "SAX", 3: "SUB16", 4: "ADD8", 7: "SUB8"}[op2]
	if prefix == "" || op == "" {
		return ""
	}
	return prefix + op
}

// decodeMediaPack handles op1 == 0b01xxx: PKH, (U)SAT, (U)SAT16, the extend
// family (SXT/UXT and their accumulate forms), SEL, and REV/REV16/REVSH.
func decodeMediaPack(w uint32, in Inst) Inst {
	op1 := (w >> 20) & 0x1F
	op2 := (w >> 5) & 7
	rd := regName[(w>>12)&0xF]
	rn := regName[(w>>16)&0xF]
	rm := regName[w&0xF]

	switch {
	case op1 == 0b01000 && op2&1 == 0: // PKHBT / PKHTB
		imm := (w >> 7) & 0x1F
		if (w>>6)&1 == 0 {
			in.Mnem = "PKHBT" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s, LSL #%d", in.Mnem, rd, rn, rm, imm)
		} else {
			in.Mnem = "PKHTB" + cn(in.Cond)
			if imm == 0 {
				imm = 32
			}
			in.Text = fmt.Sprintf("%s %s, %s, %s, ASR #%d", in.Mnem, rd, rn, rm, imm)
		}
		return in

	case (op1 == 0b01010 || op1 == 0b01011) && op2&1 == 0: // SSAT
		sat := (w>>16)&0x1F + 1
		return satText(w, in, "SSAT", sat)
	case (op1 == 0b01110 || op1 == 0b01111) && op2&1 == 0: // USAT
		sat := (w >> 16) & 0x1F
		return satText(w, in, "USAT", sat)
	case op1 == 0b01010 && op2 == 0b001: // SSAT16
		in.Mnem = "SSAT16" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, #%d, %s", in.Mnem, rd, (w>>16)&0xF+1, rm)
		return in
	case op1 == 0b01110 && op2 == 0b001: // USAT16
		in.Mnem = "USAT16" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, #%d, %s", in.Mnem, rd, (w>>16)&0xF, rm)
		return in

	case op2 == 0b011: // extend family (SXTB16/SXTB/SXTH/UXT…, with/without Rn)
		if name := extendName(op1); name != "" {
			rot := (w >> 10) & 3 * 8
			acc := (w>>16)&0xF != 0xF
			in.Mnem = name + cn(in.Cond)
			if acc {
				in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rd, rn, rm)
			} else {
				in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, rd, rm)
			}
			if rot != 0 {
				in.Text += fmt.Sprintf(", ROR #%d", rot)
			}
			return in
		}
	case op1 == 0b01011 && op2 == 0b001: // REV
		in.Mnem = "REV" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, rd, rm)
		return in
	case op1 == 0b01011 && op2 == 0b101: // REV16
		in.Mnem = "REV16" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, rd, rm)
		return in
	case op1 == 0b01111 && op2 == 0b101: // REVSH
		in.Mnem = "REVSH" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, rd, rm)
		return in
	case op1 == 0b01000 && op2 == 0b101: // SEL
		in.Mnem = "SEL" + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rd, rn, rm)
		return in
	}
	return undef(w, in)
}

func satText(w uint32, in Inst, name string, sat uint32) Inst {
	rd := regName[(w>>12)&0xF]
	rm := regName[w&0xF]
	amt := (w >> 7) & 0x1F
	sh := "LSL"
	if (w>>6)&1 == 1 {
		sh = "ASR"
		if amt == 0 {
			amt = 32
		}
	}
	in.Mnem = name + cn(in.Cond)
	if amt == 0 {
		in.Text = fmt.Sprintf("%s %s, #%d, %s", in.Mnem, rd, sat, rm)
	} else {
		in.Text = fmt.Sprintf("%s %s, #%d, %s, %s #%d", in.Mnem, rd, sat, rm, sh, amt)
	}
	return in
}

// extendName maps the op1 of the extend group to its mnemonic. The accumulate
// (…AB…/…AH…) forms differ only by Rn != 15, resolved at format time.
func extendName(op1 uint32) string {
	switch op1 {
	case 0b01000:
		return "SXTB16" // or SXTAB16 with Rn
	case 0b01010:
		return "SXTB"
	case 0b01011:
		return "SXTH"
	case 0b01100:
		return "UXTB16"
	case 0b01110:
		return "UXTB"
	case 0b01111:
		return "UXTH"
	}
	return ""
}

// decodeMediaMul handles op1 == 0b10xxx: the signed dual multiplies and the
// most-significant-word multiplies.
func decodeMediaMul(w uint32, in Inst) Inst {
	op1 := (w >> 20) & 0x1F
	op2 := (w >> 5) & 7
	rd := regName[(w>>16)&0xF] // note: media multiplies put Rd at 19:16
	ra := (w >> 12) & 0xF
	rm := regName[(w>>8)&0xF]
	rn := regName[w&0xF]
	x := "X"
	if (w>>5)&1 == 0 {
		x = ""
	}

	switch op1 {
	case 0b10000: // SMLAD / SMUAD / SMLSD / SMUSD (op2 selects add vs subtract)
		var base string
		switch op2 >> 1 {
		case 0b00:
			base = "SMLAD"
		case 0b01:
			base = "SMLSD"
		default:
			return undef(w, in)
		}
		if ra == 15 { // no accumulate → SMUAD / SMUSD
			base = map[string]string{"SMLAD": "SMUAD", "SMLSD": "SMUSD"}[base]
			in.Mnem = base + x + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rd, rn, rm)
		} else {
			in.Mnem = base + x + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, rd, rn, rm, regName[ra])
		}
		return in
	case 0b10100: // SMLALD / SMLSLD
		base := "SMLALD"
		if op2>>1 == 1 {
			base = "SMLSLD"
		}
		in.Mnem = base + x + cn(in.Cond)
		in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, regName[ra], rd, rn, rm)
		return in
	case 0b10101: // SMMLA / SMMUL / SMMLS
		r := ""
		if (w>>5)&1 == 1 {
			r = "R" // rounding variant
		}
		switch op2 >> 1 {
		case 0b00:
			if ra == 15 {
				in.Mnem = "SMMUL" + r + cn(in.Cond)
				in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, rd, rn, rm)
			} else {
				in.Mnem = "SMMLA" + r + cn(in.Cond)
				in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, rd, rn, rm, regName[ra])
			}
			return in
		case 0b11:
			in.Mnem = "SMMLS" + r + cn(in.Cond)
			in.Text = fmt.Sprintf("%s %s, %s, %s, %s", in.Mnem, rd, rn, rm, regName[ra])
			return in
		}
	}
	return undef(w, in)
}

// decodeV6Uncond handles the unconditional-space additions: CPS, SETEND and
// CLREX. It reports handled==false for the rest of the unconditional space so
// the caller can defer to the ARMv5TE decoder (BLX <imm>, PLD).
func decodeV6Uncond(w uint32, in Inst) (Inst, bool) {
	switch {
	case (w>>20)&0xFF == 0b00010000 && (w>>16)&1 == 0: // CPS
		imod := (w >> 18) & 3
		in.Mnem = "CPS"
		switch imod {
		case 0b10:
			in.Mnem = "CPSIE"
		case 0b11:
			in.Mnem = "CPSID"
		}
		flags := ""
		for _, ch := range []struct {
			bit  uint32
			name string
		}{{8, "a"}, {7, "i"}, {6, "f"}} {
			if (w>>ch.bit)&1 == 1 {
				flags += ch.name
			}
		}
		if (w>>17)&1 == 1 { // mode change present
			in.Text = fmt.Sprintf("%s %s, #0x%X", in.Mnem, flags, w&0x1F)
		} else {
			in.Text = fmt.Sprintf("%s %s", in.Mnem, flags)
		}
		return in, true
	case (w>>16)&0xFFFF == 0b1111000100000001 && (w>>4)&1 == 0: // SETEND
		e := "LE"
		if (w>>9)&1 == 1 {
			e = "BE"
		}
		in.Mnem = "SETEND"
		in.Text = "SETEND " + e
		return in, true
	case (w>>20)&0xFF == 0b01010111 && (w>>4)&0xF == 0b0001: // CLREX
		in.Mnem = "CLREX"
		in.Text = "CLREX"
		return in, true
	}
	// Not a v6 unconditional addition; defer to the v5 decoder (BLX <imm>, PLD).
	return in, false
}
