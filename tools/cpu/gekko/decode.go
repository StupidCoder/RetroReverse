package gekko

// decode.go turns a 32-bit word into an Inst. It is pure: no CPU, no memory, no state.
//
// PowerPC's encoding is regular in a way MIPS's is not. Every instruction is one word;
// the primary opcode is always the top six bits; and where a primary opcode is a family
// rather than an instruction, the member is named by an "extended opcode" in the low
// bits — ten bits at 21-30 for most, five at 26-30 for the four-operand floating-point
// forms, six at 25-30 for the indexed quantised loads. Three of the families (4, 19,
// 31, 63) therefore need two lookups, and two of them (4 and 63) mix extended-opcode
// widths in one primary opcode, which is the only genuinely awkward corner of the map.
//
// The widths do not actually collide, and that is a property worth stating rather than
// hoping for: the five-bit A-form values in opcode 63 are {18,20,21,22,23,25,26,28-31},
// and every ten-bit form in opcode 63 reduces, modulo 32, to one of {0,6,7,8,12,14,15} —
// disjoint. The same holds in opcode 4. So the decoder may test the narrow field first
// and fall through to the wide one, and TestExtendedOpcodesDoNotCollide in decode_test.go
// asserts it exhaustively rather than trusting this paragraph.
//
// Extended mnemonics are produced where PowerPC assembly has them, because a disassembly
// of two megabytes of compiler output is unreadable without them: `bne cr0,+8` rather
// than `bc 4,2,+8`, `li r3,0` rather than `addi r3,r0,0`, `mr r4,r5` rather than
// `or r4,r5,r5`, `nop` rather than `ori r0,r0,0`. The bare form is always still correct;
// this is presentation, and it changes no Flow and no Target.

import (
	"encoding/binary"
	"fmt"
)

// Decode reads one instruction from the front of code, which is big-endian.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 4 {
		return Inst{Addr: addr, Len: 0, Mnem: "?", Text: "(truncated)", Flow: FlowStop}
	}
	return DecodeWord(binary.BigEndian.Uint32(code), addr)
}

// DecodeWord decodes one already-loaded instruction word.
func DecodeWord(w, addr uint32) Inst {
	in := Inst{Addr: addr, Word: w, Len: 4, Flow: FlowSeq}
	decode(&in, w, addr)
	if in.Mnem == "" {
		in.Mnem = ".word"
		in.Text = fmt.Sprintf(".word 0x%08X", w)
		in.Flow = FlowStop // an unknown encoding ends a traced path rather than inventing one
	}
	return in
}

// set fills in the mnemonic and the rendered text together, so the two cannot drift.
func (in *Inst) set(mnem, format string, args ...interface{}) {
	in.Mnem = mnem
	if format == "" {
		in.Text = mnem
		return
	}
	in.Text = mnem + " " + fmt.Sprintf(format, args...)
}

func r(n uint32) string  { return fmt.Sprintf("r%d", n) }
func fr(n uint32) string { return fmt.Sprintf("f%d", n) }

// dot appends the record-bit suffix. `add` and `add.` differ only in whether they set
// CR0, and the dot is how PowerPC says so.
func dot(mnem string, w uint32) string {
	if rcbit(w) {
		return mnem + "."
	}
	return mnem
}

// oeDot appends the overflow-enable and record suffixes: add, add., addo, addo.
func oeDot(mnem string, w uint32) string {
	if oe(w) {
		mnem += "o"
	}
	return dot(mnem, w)
}

func decode(in *Inst, w, addr uint32) {
	switch opcd(w) {
	case 3:
		in.set("twi", "%d,%s,%d", rs(w), r(ra(w)), simm(w))
		in.Flow = FlowStop // a trap that fires does not return; treat it as a path end
	case 4:
		decodePS(in, w) // the paired-single space — see decode_ps.go
	case 7:
		in.set("mulli", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
	case 8:
		in.set("subfic", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
	case 10:
		in.set("cmplwi", "cr%d,%s,%d", crfD(w), r(ra(w)), uimm(w))
	case 11:
		in.set("cmpwi", "cr%d,%s,%d", crfD(w), r(ra(w)), simm(w))
	case 12:
		in.set("addic", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
	case 13:
		in.set("addic.", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
	case 14:
		// addi with rA = 0 reads zero rather than r0, which makes it "load immediate".
		if ra(w) == 0 {
			in.set("li", "%s,%d", r(rs(w)), simm(w))
		} else {
			in.set("addi", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
		}
	case 15:
		if ra(w) == 0 {
			in.set("lis", "%s,0x%X", r(rs(w)), uimm(w))
		} else {
			in.set("addis", "%s,%s,%d", r(rs(w)), r(ra(w)), simm(w))
		}
	case 16:
		decodeBC(in, w, addr)
	case 17:
		in.set("sc", "")
		in.Flow = FlowStop
	case 18:
		decodeB(in, w, addr)
	case 19:
		decode19(in, w)
	case 20:
		in.set(dot("rlwimi", w), "%s,%s,%d,%d,%d", r(ra(w)), r(rs(w)), shOf(w), mbOf(w), meOf(w))
	case 21:
		decodeRlwinm(in, w)
	case 23:
		in.set(dot("rlwnm", w), "%s,%s,%s,%d,%d", r(ra(w)), r(rs(w)), r(rb(w)), mbOf(w), meOf(w))
	case 24:
		// ori r0,r0,0 is the canonical no-op, and compilers emit a lot of them.
		if rs(w) == 0 && ra(w) == 0 && uimm(w) == 0 {
			in.set("nop", "")
		} else {
			in.set("ori", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
		}
	case 25:
		in.set("oris", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
	case 26:
		in.set("xori", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
	case 27:
		in.set("xoris", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
	case 28:
		in.set("andi.", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
	case 29:
		in.set("andis.", "%s,%s,0x%X", r(ra(w)), r(rs(w)), uimm(w))
	case 31:
		decode31(in, w)

	// The load/store space. rA = 0 means "no base register, use zero" for the
	// non-update forms, which is how a program addresses absolute low memory.
	case 32:
		mem(in, "lwz", r(rs(w)), w)
	case 33:
		mem(in, "lwzu", r(rs(w)), w)
	case 34:
		mem(in, "lbz", r(rs(w)), w)
	case 35:
		mem(in, "lbzu", r(rs(w)), w)
	case 36:
		mem(in, "stw", r(rs(w)), w)
	case 37:
		mem(in, "stwu", r(rs(w)), w)
	case 38:
		mem(in, "stb", r(rs(w)), w)
	case 39:
		mem(in, "stbu", r(rs(w)), w)
	case 40:
		mem(in, "lhz", r(rs(w)), w)
	case 41:
		mem(in, "lhzu", r(rs(w)), w)
	case 42:
		mem(in, "lha", r(rs(w)), w)
	case 43:
		mem(in, "lhau", r(rs(w)), w)
	case 44:
		mem(in, "sth", r(rs(w)), w)
	case 45:
		mem(in, "sthu", r(rs(w)), w)
	case 46:
		mem(in, "lmw", r(rs(w)), w)
	case 47:
		mem(in, "stmw", r(rs(w)), w)
	case 48:
		mem(in, "lfs", fr(rs(w)), w)
	case 49:
		mem(in, "lfsu", fr(rs(w)), w)
	case 50:
		mem(in, "lfd", fr(rs(w)), w)
	case 51:
		mem(in, "lfdu", fr(rs(w)), w)
	case 52:
		mem(in, "stfs", fr(rs(w)), w)
	case 53:
		mem(in, "stfsu", fr(rs(w)), w)
	case 54:
		mem(in, "stfd", fr(rs(w)), w)
	case 55:
		mem(in, "stfdu", fr(rs(w)), w)

	case 56, 57, 60, 61:
		decodePSQ(in, w) // psq_l / psq_lu / psq_st / psq_stu — see decode_ps.go

	case 59:
		decode59(in, w)
	case 63:
		decode63(in, w)
	}
}

// mem renders the d(rA) addressing mode shared by every non-indexed load and store.
func mem(in *Inst, mnem, reg string, w uint32) {
	d := simm(w)
	if ra(w) == 0 {
		// Not an update form's illegal rA=0; for the plain forms this genuinely means
		// "address zero plus the displacement", and printing r0 would be a lie.
		in.set(mnem, "%s,%d(0)", reg, d)
		return
	}
	in.set(mnem, "%s,%d(%s)", reg, d, r(ra(w)))
}

// memx renders the indexed forms: rA + rB, with rA = 0 again meaning zero.
func memx(in *Inst, mnem, reg string, w uint32) {
	base := r(ra(w))
	if ra(w) == 0 {
		base = "0"
	}
	in.set(mnem, "%s,%s,%s", reg, base, r(rb(w)))
}

// decodeRlwinm names the three extended mnemonics that cover almost every use of it.
// The compiler emits rlwinm constantly and almost never means "rotate": it means
// "shift", or "extract a bit-field", and reading it as a rotate-and-mask every time is
// how a disassembly becomes unreadable.
func decodeRlwinm(in *Inst, w uint32) {
	sh, mb, me := shOf(w), mbOf(w), meOf(w)
	name := dot("rlwinm", w)
	switch {
	case mb == 0 && me == 31-sh:
		// rotate left by sh, keep the top (32-sh) bits: a left shift.
		in.set(dot("slwi", w), "%s,%s,%d", r(ra(w)), r(rs(w)), sh)
	case me == 31 && mb == 32-sh && sh != 0:
		// rotate left by sh == rotate right by 32-sh, keep the bottom: a right shift.
		in.set(dot("srwi", w), "%s,%s,%d", r(ra(w)), r(rs(w)), 32-sh)
	case sh == 0 && mb == 0:
		in.set(dot("clrrwi", w), "%s,%s,%d", r(ra(w)), r(rs(w)), 31-me)
	case sh == 0 && me == 31:
		in.set(dot("clrlwi", w), "%s,%s,%d", r(ra(w)), r(rs(w)), mb)
	default:
		in.set(name, "%s,%s,%d,%d,%d", r(ra(w)), r(rs(w)), sh, mb, me)
	}
}

// --- Branches ------------------------------------------------------------------

// The BO field, bit by bit (PowerPC numbers them from the top, so BO0 is worth 16).
const (
	boNoCond  = 16 // BO0: do not test the condition register
	boCondSet = 8  // BO1: the CR bit value to branch on
	boNoDec   = 4  // BO2: do not decrement CTR
	boCTRZero = 2  // BO3: the CTR test — branch when CTR == 0
)

// boAlways reports whether a BO field means "branch, unconditionally": it neither tests
// the condition register nor the count register.
func boAlways(bo uint32) bool { return bo&boNoCond != 0 && bo&boNoDec != 0 }

// crBit names one bit of one condition-register field, as a branch's BI selects it.
var crBitName = [4]string{"lt", "gt", "eq", "so"}

// The extended branch mnemonics: branch-if-true and branch-if-false, per CR bit.
var brTrue = [4]string{"blt", "bgt", "beq", "bso"}
var brFalse = [4]string{"bge", "ble", "bne", "bns"}

// branchMnem builds the readable name for a conditional branch, and reports whether it
// managed to. `suffix` is "" for a relative branch, "lr" for bclr, "ctr" for bcctr.
func branchMnem(bo, bi uint32, suffix string, link bool) (string, bool) {
	l := ""
	if link {
		l = "l"
	}
	dec := bo&boNoDec == 0 // this branch decrements CTR
	cond := bo&boNoCond == 0

	switch {
	case dec && !cond:
		// A pure loop branch: bdnz / bdz.
		name := "bdnz"
		if bo&boCTRZero != 0 {
			name = "bdz"
		}
		return name + suffix + l, true
	case dec && cond:
		// Both tests: bdnzt/bdnzf/bdzt/bdzf, which take the CR bit as an operand.
		name := "bdnz"
		if bo&boCTRZero != 0 {
			name = "bdz"
		}
		if bo&boCondSet != 0 {
			name += "t"
		} else {
			name += "f"
		}
		return name + suffix + l, false // false: the CR bit still has to be printed
	case !dec && cond:
		tbl := brFalse
		if bo&boCondSet != 0 {
			tbl = brTrue
		}
		return tbl[bi&3] + suffix + l, true
	}
	// Neither test: an unconditional branch through this encoding.
	return "b" + suffix + l, true
}

func decodeBC(in *Inst, w, addr uint32) {
	bo, bi := rs(w), ra(w)
	// The displacement is a signed 14-bit word offset, held in bits 16-29.
	d := int32(int16(w & 0xFFFC)) // the low two bits are AA and LK, and are not part of it
	target := uint32(int32(addr) + d)
	if aa(w) {
		target = uint32(d)
	}
	in.Target, in.HasTarget = target, true

	name, complete := branchMnem(bo, bi, "", lk(w))
	crf := bi >> 2
	switch {
	case complete && boAlways(bo):
		in.set(name, "0x%08X", target)
	case complete:
		if crf == 0 {
			in.set(name, "0x%08X", target)
		} else {
			in.set(name, "cr%d,0x%08X", crf, target)
		}
	default:
		in.set(name, "cr%d[%s],0x%08X", crf, crBitName[bi&3], target)
	}

	switch {
	case lk(w):
		in.Flow = FlowCall // a conditional call still returns; the tracer walks both ways
	case boAlways(bo):
		in.Flow = FlowJump
	default:
		in.Flow = FlowBranch
	}
}

func decodeB(in *Inst, w, addr uint32) {
	// A signed 26-bit word offset in bits 6-29.
	d := int32(w&0x03FFFFFC) << 6 >> 6
	target := uint32(int32(addr) + d)
	if aa(w) {
		target = uint32(d)
	}
	in.Target, in.HasTarget = target, true

	name := "b"
	if aa(w) {
		name += "a"
	}
	if lk(w) {
		name += "l"
	}
	in.set(name, "0x%08X", target)

	switch {
	case !lk(w):
		in.Flow = FlowJump
	case target == addr+4:
		// `bl .+4` is not a call. It is how position-independent code reads its own PC:
		// the link register is the point, and control simply carries on. Classifying it
		// as a call would invent a function at every one of them and fill the tracer's
		// caller graph with noise.
		in.Flow = FlowSeq
	default:
		in.Flow = FlowCall
	}
}

// decode19 is the branch-to-register and condition-register-logic family.
func decode19(in *Inst, w uint32) {
	bo, bi := rs(w), ra(w)
	switch xo10(w) {
	case 0:
		in.set("mcrf", "cr%d,cr%d", crfD(w), crfS(w))
	case 16, 528:
		toCTR := xo10(w) == 528
		suffix := "lr"
		if toCTR {
			suffix = "ctr"
		}
		name, complete := branchMnem(bo, bi, suffix, lk(w))
		crf := bi >> 2
		switch {
		case complete && boAlways(bo):
			in.set(name, "")
		case complete:
			in.set(name, "cr%d", crf)
		default:
			in.set(name, "cr%d[%s]", crf, crBitName[bi&3])
		}

		// This is the classification PowerPC needs and MIPS does not: the branch is
		// indirect AND it may be conditional, so "returns" and "falls through" are not
		// mutually exclusive. A conditional blr continues down the fall-through path.
		switch {
		case lk(w):
			in.Flow = FlowIndCall
		case !boAlways(bo):
			in.Flow = FlowBranch // conditional: it may return, but it may also carry on
		case toCTR:
			in.Flow = FlowIndJump
		default:
			in.Flow = FlowReturn
		}
	case 50:
		in.set("rfi", "")
		in.Flow = FlowStop
	case 150:
		in.set("isync", "")
	case 33, 129, 193, 225, 257, 289, 417, 449:
		crLogic(in, w)
	}
}

var crLogicName = map[uint32]string{
	33: "crnor", 129: "crandc", 193: "crxor", 225: "crnand",
	257: "crand", 289: "creqv", 417: "crorc", 449: "cror",
}

func crLogic(in *Inst, w uint32) {
	name := crLogicName[xo10(w)]
	d, a, b := rs(w), ra(w), rb(w)
	// crxor of a bit with itself clears it, and creqv of a bit with itself sets it;
	// both idioms are common enough in compiler output to be worth naming.
	switch {
	case name == "crxor" && d == a && a == b:
		in.set("crclr", "cr%d[%s]", d>>2, crBitName[d&3])
	case name == "creqv" && d == a && a == b:
		in.set("crset", "cr%d[%s]", d>>2, crBitName[d&3])
	case name == "cror" && a == b:
		in.set("crmove", "cr%d[%s],cr%d[%s]", d>>2, crBitName[d&3], a>>2, crBitName[a&3])
	default:
		in.set(name, "cr%d[%s],cr%d[%s],cr%d[%s]",
			d>>2, crBitName[d&3], a>>2, crBitName[a&3], b>>2, crBitName[b&3])
	}
}

// --- Opcode 31: the integer, load/store-indexed and system family -----------------

// oeForm lists the opcode-31 instructions that carry an overflow-enable bit. It exists
// because of an encoding trap: OE is bit 21, which is also the *top bit of the ten-bit
// extended-opcode field*. So `addo` does not encode extended opcode 266 with a flag set
// somewhere else — it encodes 266+512 = 778, and a decoder that switches on the ten-bit
// field alone silently fails to recognise every arithmetic instruction that can overflow.
// Stripping the bit unconditionally would be just as wrong, since the X-form instructions
// really do use all ten bits. So it is stripped for exactly these, and nothing else.
var oeForm = map[uint32]bool{
	8: true, 10: true, 40: true, 104: true, 136: true, 138: true,
	200: true, 202: true, 232: true, 234: true, 235: true, 266: true,
	459: true, 491: true,
}

func decode31(in *Inst, w uint32) {
	d, a, b := rs(w), ra(w), rb(w)

	x := xo10(w)
	if x >= 512 && oeForm[x-512] {
		x -= 512
	}

	switch x {
	// Compares. L selects 32- or 64-bit; the Gekko is 32-bit, so the mnemonic is fixed.
	case 0:
		in.set("cmpw", "cr%d,%s,%s", crfD(w), r(a), r(b))
	case 32:
		in.set("cmplw", "cr%d,%s,%s", crfD(w), r(a), r(b))
	case 4:
		in.set("tw", "%d,%s,%s", d, r(a), r(b))
		in.Flow = FlowStop

	// Arithmetic. The three-operand forms all share the OE/Rc suffix machinery.
	case 8:
		in.set(oeDot("subfc", w), "%s,%s,%s", r(d), r(a), r(b))
	case 10:
		in.set(oeDot("addc", w), "%s,%s,%s", r(d), r(a), r(b))
	case 11:
		in.set(dot("mulhwu", w), "%s,%s,%s", r(d), r(a), r(b))
	case 40:
		// subf rD,rA,rB is rB - rA; `sub rD,rB,rA` is the readable spelling.
		in.set(oeDot("subf", w), "%s,%s,%s", r(d), r(a), r(b))
	case 75:
		in.set(dot("mulhw", w), "%s,%s,%s", r(d), r(a), r(b))
	case 104:
		in.set(oeDot("neg", w), "%s,%s", r(d), r(a))
	case 136:
		in.set(oeDot("subfe", w), "%s,%s,%s", r(d), r(a), r(b))
	case 138:
		in.set(oeDot("adde", w), "%s,%s,%s", r(d), r(a), r(b))
	case 200:
		in.set(oeDot("subfze", w), "%s,%s", r(d), r(a))
	case 202:
		in.set(oeDot("addze", w), "%s,%s", r(d), r(a))
	case 232:
		in.set(oeDot("subfme", w), "%s,%s", r(d), r(a))
	case 234:
		in.set(oeDot("addme", w), "%s,%s", r(d), r(a))
	case 235:
		in.set(oeDot("mullw", w), "%s,%s,%s", r(d), r(a), r(b))
	case 266:
		in.set(oeDot("add", w), "%s,%s,%s", r(d), r(a), r(b))
	case 459:
		in.set(oeDot("divwu", w), "%s,%s,%s", r(d), r(a), r(b))
	case 491:
		in.set(oeDot("divw", w), "%s,%s,%s", r(d), r(a), r(b))

	// Logical and shift. `or rA,rS,rS` is a register move, and it is everywhere.
	case 28:
		in.set(dot("and", w), "%s,%s,%s", r(a), r(d), r(b))
	case 60:
		in.set(dot("andc", w), "%s,%s,%s", r(a), r(d), r(b))
	case 124:
		if d == b {
			in.set(dot("not", w), "%s,%s", r(a), r(d))
		} else {
			in.set(dot("nor", w), "%s,%s,%s", r(a), r(d), r(b))
		}
	case 284:
		in.set(dot("eqv", w), "%s,%s,%s", r(a), r(d), r(b))
	case 316:
		in.set(dot("xor", w), "%s,%s,%s", r(a), r(d), r(b))
	case 412:
		in.set(dot("orc", w), "%s,%s,%s", r(a), r(d), r(b))
	case 444:
		if d == b {
			in.set(dot("mr", w), "%s,%s", r(a), r(d))
		} else {
			in.set(dot("or", w), "%s,%s,%s", r(a), r(d), r(b))
		}
	case 476:
		in.set(dot("nand", w), "%s,%s,%s", r(a), r(d), r(b))
	case 24:
		in.set(dot("slw", w), "%s,%s,%s", r(a), r(d), r(b))
	case 536:
		in.set(dot("srw", w), "%s,%s,%s", r(a), r(d), r(b))
	case 792:
		in.set(dot("sraw", w), "%s,%s,%s", r(a), r(d), r(b))
	case 824:
		in.set(dot("srawi", w), "%s,%s,%d", r(a), r(d), shOf(w))
	case 26:
		in.set(dot("cntlzw", w), "%s,%s", r(a), r(d))
	case 922:
		in.set(dot("extsh", w), "%s,%s", r(a), r(d))
	case 954:
		in.set(dot("extsb", w), "%s,%s", r(a), r(d))

	// Moves to and from the special registers.
	case 19:
		in.set("mfcr", "%s", r(d))
	case 144:
		// mtcrf with a full mask is mtcr: the compiler's usual spelling.
		mask := (w >> 12) & 0xFF
		if mask == 0xFF {
			in.set("mtcr", "%s", r(d))
		} else {
			in.set("mtcrf", "0x%02X,%s", mask, r(d))
		}
	case 339:
		in.set("mfspr", "%s,%s", r(d), sprStr(sprOf(w)))
	case 467:
		in.set("mtspr", "%s,%s", sprStr(sprOf(w)), r(d))
	case 371:
		in.set("mftb", "%s,%s", r(d), sprStr(sprOf(w)))
	case 83:
		in.set("mfmsr", "%s", r(d))
	case 146:
		in.set("mtmsr", "%s", r(d))
	case 210:
		in.set("mtsr", "%d,%s", a&15, r(d))
	case 242:
		in.set("mtsrin", "%s,%s", r(d), r(b))
	case 595:
		in.set("mfsr", "%s,%d", r(d), a&15)
	case 659:
		in.set("mfsrin", "%s,%s", r(d), r(b))
	case 512:
		in.set("mcrxr", "cr%d", crfD(w))

	// Indexed loads and stores.
	case 23:
		memx(in, "lwzx", r(d), w)
	case 55:
		memx(in, "lwzux", r(d), w)
	case 87:
		memx(in, "lbzx", r(d), w)
	case 119:
		memx(in, "lbzux", r(d), w)
	case 279:
		memx(in, "lhzx", r(d), w)
	case 311:
		memx(in, "lhzux", r(d), w)
	case 343:
		memx(in, "lhax", r(d), w)
	case 375:
		memx(in, "lhaux", r(d), w)
	case 151:
		memx(in, "stwx", r(d), w)
	case 183:
		memx(in, "stwux", r(d), w)
	case 215:
		memx(in, "stbx", r(d), w)
	case 247:
		memx(in, "stbux", r(d), w)
	case 407:
		memx(in, "sthx", r(d), w)
	case 439:
		memx(in, "sthux", r(d), w)

	// The byte-reversed forms. On a big-endian machine these are the little-endian
	// accessors, and they are exactly the ones a careless transcription gets backwards.
	case 534:
		memx(in, "lwbrx", r(d), w)
	case 790:
		memx(in, "lhbrx", r(d), w)
	case 662:
		memx(in, "stwbrx", r(d), w)
	case 918:
		memx(in, "sthbrx", r(d), w)

	// Load/store string and multiple.
	case 533:
		memx(in, "lswx", r(d), w)
	case 597:
		in.set("lswi", "%s,%s,%d", r(d), r(a), rb(w))
	case 661:
		memx(in, "stswx", r(d), w)
	case 725:
		in.set("stswi", "%s,%s,%d", r(d), r(a), rb(w))

	// The atomic pair. lwarx reserves; stwcx. stores only if the reservation held, and
	// sets CR0[eq] to say whether it did.
	case 20:
		memx(in, "lwarx", r(d), w)
	case 150:
		memx(in, "stwcx.", r(d), w)

	// Indexed floating-point loads and stores.
	case 535:
		memx(in, "lfsx", fr(d), w)
	case 567:
		memx(in, "lfsux", fr(d), w)
	case 599:
		memx(in, "lfdx", fr(d), w)
	case 631:
		memx(in, "lfdux", fr(d), w)
	case 663:
		memx(in, "stfsx", fr(d), w)
	case 695:
		memx(in, "stfsux", fr(d), w)
	case 727:
		memx(in, "stfdx", fr(d), w)
	case 759:
		memx(in, "stfdux", fr(d), w)
	case 983:
		memx(in, "stfiwx", fr(d), w)

	// Cache and synchronisation. These are not hints on this machine: dcbz writes
	// zeroes, and the game uses it to clear memory.
	case 54:
		in.set("dcbst", "%s,%s", r(a), r(b))
	case 86:
		in.set("dcbf", "%s,%s", r(a), r(b))
	case 246:
		in.set("dcbtst", "%s,%s", r(a), r(b))
	case 278:
		in.set("dcbt", "%s,%s", r(a), r(b))
	case 470:
		in.set("dcbi", "%s,%s", r(a), r(b))
	case 1014:
		in.set("dcbz", "%s,%s", r(a), r(b))
	case 982:
		in.set("icbi", "%s,%s", r(a), r(b))
	case 598:
		in.set("sync", "")
	case 854:
		in.set("eieio", "")

	// The MMU maintenance and external-control instructions. A GameCube runs on the
	// block-translation registers and never installs a page table, so if one of these
	// ever executes it is news — but decoding it is still the decoder's job.
	case 306:
		in.set("tlbie", "%s", r(b))
	case 370:
		in.set("tlbia", "")
	case 566:
		in.set("tlbsync", "")
	case 310:
		memx(in, "eciwx", r(d), w)
	case 438:
		memx(in, "ecowx", r(d), w)
	}
}

// --- The floating-point families ---------------------------------------------------

// decode59 is single-precision floating point. Every one of these rounds its double
// result to single precision, and the "rounds twice" that implies is a real numerical
// effect, not a formality — see fpu.go.
func decode59(in *Inst, w uint32) {
	d, a, b, c := rs(w), ra(w), rb(w), rc(w)
	switch xo5(w) {
	case 18:
		in.set(dot("fdivs", w), "%s,%s,%s", fr(d), fr(a), fr(b))
	case 20:
		in.set(dot("fsubs", w), "%s,%s,%s", fr(d), fr(a), fr(b))
	case 21:
		in.set(dot("fadds", w), "%s,%s,%s", fr(d), fr(a), fr(b))
	case 22:
		in.set(dot("fsqrts", w), "%s,%s", fr(d), fr(b))
	case 24:
		in.set(dot("fres", w), "%s,%s", fr(d), fr(b))
	case 25:
		in.set(dot("fmuls", w), "%s,%s,%s", fr(d), fr(a), fr(c))
	case 28:
		in.set(dot("fmsubs", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
	case 29:
		in.set(dot("fmadds", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
	case 30:
		in.set(dot("fnmsubs", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
	case 31:
		in.set(dot("fnmadds", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
	}
}

// decode63 is double-precision floating point, plus the FPSCR and register moves. The
// five-bit and ten-bit extended opcodes share this primary opcode; see the file comment
// for why testing the narrow one first is safe.
func decode63(in *Inst, w uint32) {
	d, a, b, c := rs(w), ra(w), rb(w), rc(w)

	switch xo5(w) {
	case 18:
		in.set(dot("fdiv", w), "%s,%s,%s", fr(d), fr(a), fr(b))
		return
	case 20:
		in.set(dot("fsub", w), "%s,%s,%s", fr(d), fr(a), fr(b))
		return
	case 21:
		in.set(dot("fadd", w), "%s,%s,%s", fr(d), fr(a), fr(b))
		return
	case 22:
		in.set(dot("fsqrt", w), "%s,%s", fr(d), fr(b))
		return
	case 23:
		in.set(dot("fsel", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		return
	case 25:
		in.set(dot("fmul", w), "%s,%s,%s", fr(d), fr(a), fr(c))
		return
	case 26:
		in.set(dot("frsqrte", w), "%s,%s", fr(d), fr(b))
		return
	case 28:
		in.set(dot("fmsub", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		return
	case 29:
		in.set(dot("fmadd", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		return
	case 30:
		in.set(dot("fnmsub", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		return
	case 31:
		in.set(dot("fnmadd", w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		return
	}

	switch xo10(w) {
	case 0:
		in.set("fcmpu", "cr%d,%s,%s", crfD(w), fr(a), fr(b))
	case 32:
		in.set("fcmpo", "cr%d,%s,%s", crfD(w), fr(a), fr(b))
	case 12:
		in.set(dot("frsp", w), "%s,%s", fr(d), fr(b))
	case 14:
		in.set(dot("fctiw", w), "%s,%s", fr(d), fr(b))
	case 15:
		in.set(dot("fctiwz", w), "%s,%s", fr(d), fr(b))
	case 40:
		in.set(dot("fneg", w), "%s,%s", fr(d), fr(b))
	case 72:
		in.set(dot("fmr", w), "%s,%s", fr(d), fr(b))
	case 136:
		in.set(dot("fnabs", w), "%s,%s", fr(d), fr(b))
	case 264:
		in.set(dot("fabs", w), "%s,%s", fr(d), fr(b))
	case 64:
		in.set("mcrfs", "cr%d,cr%d", crfD(w), crfS(w))
	case 38:
		in.set(dot("mtfsb1", w), "%d", d)
	case 70:
		in.set(dot("mtfsb0", w), "%d", d)
	case 134:
		in.set(dot("mtfsfi", w), "cr%d,%d", crfD(w), (w>>12)&15)
	case 583:
		in.set(dot("mffs", w), "%s", fr(d))
	case 711:
		in.set(dot("mtfsf", w), "0x%02X,%s", (w>>17)&0xFF, fr(b))
	}
}
