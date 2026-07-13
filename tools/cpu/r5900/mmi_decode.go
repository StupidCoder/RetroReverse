package r5900

// mmi_decode.go names the MMI instruction space for the disassembler and the
// tracer. The semantics are in mmi.go; the tables here only have to agree with it
// on which encodings exist and what shape their operands take.

import "fmt"

// The operand shapes an MMI instruction can take. Most are "rd, rs, rt"; the rest
// are enumerated because getting an operand order wrong in a listing is the kind
// of mistake that survives for months.
type mmiForm int

const (
	mDST  mmiForm = iota // rd, rs, rt
	mDT                  // rd, rt        (the unary permutes and absolutes)
	mDTS                 // rd, rt, rs    (the variable shifts: the amount is in rs)
	mDTSA                // rd, rt, sa    (the immediate shifts)
	mD                   // rd            (pmfhi/pmflo/mfhi1/mflo1)
	mS                   // rs            (pmthi/pmtlo/mthi1/mtlo1)
	mST                  // rs, rt        (the divides)
)

type mmiOp struct {
	name string
	form mmiForm
}

// The four sub-tables, indexed by the shamt field.
var mmi0Tab = map[uint32]mmiOp{
	0x00: {"paddw", mDST}, 0x01: {"psubw", mDST}, 0x02: {"pcgtw", mDST}, 0x03: {"pmaxw", mDST},
	0x04: {"paddh", mDST}, 0x05: {"psubh", mDST}, 0x06: {"pcgth", mDST}, 0x07: {"pmaxh", mDST},
	0x08: {"paddb", mDST}, 0x09: {"psubb", mDST}, 0x0A: {"pcgtb", mDST},
	0x10: {"paddsw", mDST}, 0x11: {"psubsw", mDST}, 0x12: {"pextlw", mDST}, 0x13: {"ppacw", mDST},
	0x14: {"paddsh", mDST}, 0x15: {"psubsh", mDST}, 0x16: {"pextlh", mDST}, 0x17: {"ppach", mDST},
	0x18: {"paddsb", mDST}, 0x19: {"psubsb", mDST}, 0x1A: {"pextlb", mDST}, 0x1B: {"ppacb", mDST},
	0x1E: {"pext5", mDT}, 0x1F: {"ppac5", mDT},
}

var mmi1Tab = map[uint32]mmiOp{
	0x01: {"pabsw", mDT}, 0x02: {"pceqw", mDST}, 0x03: {"pminw", mDST},
	0x04: {"padsbh", mDST}, 0x05: {"pabsh", mDT}, 0x06: {"pceqh", mDST}, 0x07: {"pminh", mDST},
	0x0A: {"pceqb", mDST},
	0x10: {"padduw", mDST}, 0x11: {"psubuw", mDST}, 0x12: {"pextuw", mDST},
	0x14: {"padduh", mDST}, 0x15: {"psubuh", mDST}, 0x16: {"pextuh", mDST},
	0x18: {"paddub", mDST}, 0x19: {"psubub", mDST}, 0x1A: {"pextub", mDST},
	0x1B: {"qfsrv", mDST},
}

var mmi2Tab = map[uint32]mmiOp{
	0x00: {"pmaddw", mDST}, 0x02: {"psllvw", mDTS}, 0x03: {"psrlvw", mDTS},
	0x04: {"pmsubw", mDST}, 0x08: {"pmfhi", mD}, 0x09: {"pmflo", mD}, 0x0A: {"pinth", mDST},
	0x0C: {"pmultw", mDST}, 0x0D: {"pdivw", mST}, 0x0E: {"pcpyld", mDST},
	0x10: {"pmaddh", mDST}, 0x11: {"phmadh", mDST}, 0x12: {"pand", mDST}, 0x13: {"pxor", mDST},
	0x14: {"pmsubh", mDST}, 0x15: {"phmsbh", mDST},
	0x1A: {"pexeh", mDT}, 0x1B: {"prevh", mDT},
	0x1C: {"pmulth", mDST}, 0x1D: {"pdivbw", mST}, 0x1E: {"pexew", mDT}, 0x1F: {"prot3w", mDT},
}

var mmi3Tab = map[uint32]mmiOp{
	0x00: {"pmadduw", mDST}, 0x03: {"psravw", mDTS},
	0x08: {"pmthi", mS}, 0x09: {"pmtlo", mS}, 0x0A: {"pinteh", mDST},
	0x0C: {"pmultuw", mDST}, 0x0D: {"pdivuw", mST}, 0x0E: {"pcpyud", mDST},
	0x12: {"por", mDST}, 0x13: {"pnor", mDST},
	0x1A: {"pexch", mDT}, 0x1B: {"pcpyh", mDT}, 0x1E: {"pexcw", mDT},
}

// pmfhlSub names the slice of the accumulator pmfhl reads.
var pmfhlSub = [8]string{"lw", "uw", "slw", "lh", "sh"}

// decodeMMI handles op == 0x1C.
func decodeMMI(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }

	switch w & 0x3F {
	case 0x00:
		return set("madd", fmt.Sprintf("madd %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x01:
		return set("maddu", fmt.Sprintf("maddu %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x04:
		return set("plzcw", fmt.Sprintf("plzcw %s, %s", reg(rd), reg(rs)))
	case 0x08:
		return mmiSub(in, w, mmi0Tab, rs, rt, rd, shamt)
	case 0x09:
		return mmiSub(in, w, mmi2Tab, rs, rt, rd, shamt)
	case 0x10:
		return set("mfhi1", fmt.Sprintf("mfhi1 %s", reg(rd)))
	case 0x11:
		return set("mthi1", fmt.Sprintf("mthi1 %s", reg(rs)))
	case 0x12:
		return set("mflo1", fmt.Sprintf("mflo1 %s", reg(rd)))
	case 0x13:
		return set("mtlo1", fmt.Sprintf("mtlo1 %s", reg(rs)))
	case 0x18:
		return set("mult1", fmt.Sprintf("mult1 %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x19:
		return set("multu1", fmt.Sprintf("multu1 %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x1A:
		return set("div1", fmt.Sprintf("div1 %s, %s", reg(rs), reg(rt)))
	case 0x1B:
		return set("divu1", fmt.Sprintf("divu1 %s, %s", reg(rs), reg(rt)))
	case 0x20:
		return set("madd1", fmt.Sprintf("madd1 %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x21:
		return set("maddu1", fmt.Sprintf("maddu1 %s, %s, %s", reg(rd), reg(rs), reg(rt)))
	case 0x28:
		return mmiSub(in, w, mmi1Tab, rs, rt, rd, shamt)
	case 0x29:
		return mmiSub(in, w, mmi3Tab, rs, rt, rd, shamt)
	case 0x30:
		if shamt < uint32(len(pmfhlSub)) && pmfhlSub[shamt] != "" {
			m := "pmfhl." + pmfhlSub[shamt]
			return set(m, fmt.Sprintf("%s %s", m, reg(rd)))
		}
		return word(in, w)
	case 0x31:
		if shamt == 0 {
			return set("pmthl.lw", fmt.Sprintf("pmthl.lw %s", reg(rs)))
		}
		return word(in, w)

	case 0x34, 0x36, 0x37, 0x3C, 0x3E, 0x3F:
		m := map[uint32]string{
			0x34: "psllh", 0x36: "psrlh", 0x37: "psrah",
			0x3C: "psllw", 0x3E: "psrlw", 0x3F: "psraw",
		}[w&0x3F]
		return set(m, fmt.Sprintf("%s %s, %s, %d", m, reg(rd), reg(rt), shamt))
	}
	return word(in, w)
}

// mmiSub formats one entry of a sub-table.
func mmiSub(in Inst, w uint32, tab map[uint32]mmiOp, rs, rt, rd, shamt uint32) Inst {
	o, ok := tab[shamt]
	if !ok {
		return word(in, w)
	}
	in.Mnem = o.name
	switch o.form {
	case mDST:
		in.Text = fmt.Sprintf("%s %s, %s, %s", o.name, reg(rd), reg(rs), reg(rt))
	case mDT:
		in.Text = fmt.Sprintf("%s %s, %s", o.name, reg(rd), reg(rt))
	case mDTS:
		in.Text = fmt.Sprintf("%s %s, %s, %s", o.name, reg(rd), reg(rt), reg(rs))
	case mDTSA:
		in.Text = fmt.Sprintf("%s %s, %s, %d", o.name, reg(rd), reg(rt), shamt)
	case mD:
		in.Text = fmt.Sprintf("%s %s", o.name, reg(rd))
	case mS:
		in.Text = fmt.Sprintf("%s %s", o.name, reg(rs))
	case mST:
		in.Text = fmt.Sprintf("%s %s, %s", o.name, reg(rs), reg(rt))
	}
	return in
}
