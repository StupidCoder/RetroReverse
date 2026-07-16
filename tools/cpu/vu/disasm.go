package vu

// disasm.go renders VU instructions as text. A VU instruction is one 64-bit word holding
// two operations that issue together: the UPPER (bits 32..63) runs on the floating-point
// FMAC pipes, the LOWER (bits 0..31) on the integer/load-store/branch pipe. The upper
// word's top five bits are flags, not opcode: I says the lower word is a 32-bit float
// immediate (the I register's new value), E says the program ends after the next pair,
// M/D/T are the VU0-macro and debug bits.
//
// Field layout, shared by most operations:
//
//	dest  bits 21..24   the write mask, printed .xyzw
//	ft    bits 16..20   a vector register (or it, an integer register)
//	fs    bits 11..15   a vector register (or is)
//	fd    bits  6..10   a vector register (or id / the second opcode field)
//	op    bits  0..5    the opcode; 0x3C..0x3F select the wide group, where the
//	                    operation is bits 6..10 and 0..1 combined
//
// The broadcast forms (ADDbc and family) take their scalar from field bc = bits 0..1.

import "fmt"

// Inst is one decoded pair.
type Inst struct {
	Upper, Lower string
	I, E         bool // the flag bits that change control flow / decoding
	Raw          uint64
}

// Decode renders the 64-bit instruction pair at byte address addr (used only for branch
// targets in the text).
func Decode(raw uint64, addr uint32) Inst {
	up := uint32(raw >> 32)
	lo := uint32(raw)
	inst := Inst{
		Raw: raw,
		I:   up&(1<<31) != 0,
		E:   up&(1<<30) != 0,
	}
	inst.Upper = decodeUpper(up)
	if inst.I {
		inst.Lower = fmt.Sprintf("loi %g", float32frombits(lo))
	} else {
		inst.Lower = decodeLower(lo, addr)
	}
	return inst
}

// DisasmMacro renders one macro-mode COP2 instruction word, dispatching exactly the
// way Macro executes it, so the text names what the executor would run. Macro-mode
// mnemonics carry the conventional "v" prefix that tells an EE listing's vadd from
// the FPU's add.s.
func DisasmMacro(w uint32) string {
	op := w & 0x3F
	switch {
	case op < 0x30:
		return "v" + decodeUpper(w)
	case op == 0x38:
		return fmt.Sprintf("vcallms 0x%X", w>>6&0x7FFF*8)
	case op == 0x39:
		return "vcallmsr"
	case op < 0x3C:
		// The integer ops (viadd and family) share the lower-special encodings.
		return "v" + decodeLowerSpecial(w)
	default:
		if op2 := w>>4&0x7C | w&3; op2 <= 0x2F {
			return "v" + decodeUpper(w) // the wide group's FMAC half (adda, mula, ftoi...)
		}
		return "v" + decodeLowerSpecial(w) // div, move, mtir, lqi/sqi, the EFU...
	}
}

var destNames = [16]string{
	"", "w", "z", "zw", "y", "yw", "yz", "yzw",
	"x", "xw", "xz", "xzw", "xy", "xyw", "xyz", "xyzw",
}

var bcNames = [4]string{"x", "y", "z", "w"}

func dest(i uint32) string { return destNames[i>>21&0xF] }
func ft(i uint32) uint32   { return i >> 16 & 0x1F }
func fs(i uint32) uint32   { return i >> 11 & 0x1F }
func fd(i uint32) uint32   { return i >> 6 & 0x1F }

// decodeUpper renders the FMAC half.
func decodeUpper(i uint32) string {
	op := i & 0x3F
	d, t, s, f := dest(i), ft(i), fs(i), fd(i)
	bc := bcNames[i&3]

	three := func(name string) string {
		return fmt.Sprintf("%s.%s vf%02d, vf%02d, vf%02d", name, d, f, s, t)
	}
	threeBC := func(name string) string {
		return fmt.Sprintf("%s%s.%s vf%02d, vf%02d, vf%02d%s", name, bc, d, f, s, t, bc)
	}
	threeQ := func(name string) string {
		return fmt.Sprintf("%sq.%s vf%02d, vf%02d, q", name, d, f, s)
	}
	threeI := func(name string) string {
		return fmt.Sprintf("%si.%s vf%02d, vf%02d, i", name, d, f, s)
	}

	switch {
	case op < 0x04:
		return threeBC("add")
	case op < 0x08:
		return threeBC("sub")
	case op < 0x0C:
		return threeBC("madd")
	case op < 0x10:
		return threeBC("msub")
	case op < 0x14:
		return threeBC("max")
	case op < 0x18:
		return threeBC("mini")
	case op < 0x1C:
		return threeBC("mul")
	case op == 0x1C:
		return threeQ("mul")
	case op == 0x1D:
		return threeI("max")
	case op == 0x1E:
		return threeI("mul")
	case op == 0x1F:
		return threeI("mini")
	case op == 0x20:
		return threeQ("add")
	case op == 0x21:
		return threeQ("madd")
	case op == 0x22:
		return threeI("add")
	case op == 0x23:
		return threeI("madd")
	case op == 0x24:
		return threeQ("sub")
	case op == 0x25:
		return threeQ("msub")
	case op == 0x26:
		return threeI("sub")
	case op == 0x27:
		return threeI("msub")
	case op == 0x28:
		return three("add")
	case op == 0x29:
		return three("madd")
	case op == 0x2A:
		return three("mul")
	case op == 0x2B:
		return three("max")
	case op == 0x2C:
		return three("sub")
	case op == 0x2D:
		return three("msub")
	case op == 0x2E:
		return three("opmsub")
	case op == 0x2F:
		return three("mini")
	case op >= 0x3C:
		return decodeUpperWide(i)
	}
	return fmt.Sprintf("upper? 0x%08X", i)
}

// decodeUpperWide is the 0x3C..0x3F group: the operation is bits 6..10 and 0..1.
func decodeUpperWide(i uint32) string {
	op2 := i>>4&0x7C | i&3
	d, t, s := dest(i), ft(i), fs(i)
	bc := bcNames[i&3]

	acc := func(name string) string {
		return fmt.Sprintf("%s.%s acc, vf%02d, vf%02d", name, d, s, t)
	}
	accBC := func(name string) string {
		return fmt.Sprintf("%s%s.%s acc, vf%02d, vf%02d%s", name, bc, d, s, t, bc)
	}
	accQ := func(name string) string {
		return fmt.Sprintf("%sq.%s acc, vf%02d, q", name, d, s)
	}
	accI := func(name string) string {
		return fmt.Sprintf("%si.%s acc, vf%02d, i", name, d, s)
	}
	conv := func(name string) string {
		return fmt.Sprintf("%s.%s vf%02d, vf%02d", name, d, t, s)
	}

	switch {
	case op2 < 0x04:
		return accBC("adda")
	case op2 < 0x08:
		return accBC("suba")
	case op2 < 0x0C:
		return accBC("madda")
	case op2 < 0x10:
		return accBC("msuba")
	case op2 == 0x10:
		return conv("itof0")
	case op2 == 0x11:
		return conv("itof4")
	case op2 == 0x12:
		return conv("itof12")
	case op2 == 0x13:
		return conv("itof15")
	case op2 == 0x14:
		return conv("ftoi0")
	case op2 == 0x15:
		return conv("ftoi4")
	case op2 == 0x16:
		return conv("ftoi12")
	case op2 == 0x17:
		return conv("ftoi15")
	case op2 < 0x1C:
		return accBC("mula")
	case op2 == 0x1C:
		return accQ("mula")
	case op2 == 0x1D:
		return conv("abs")
	case op2 == 0x1E:
		return accI("mula")
	case op2 == 0x1F:
		return fmt.Sprintf("clipw.xyz vf%02d, vf%02dw", fs(i), ft(i))
	case op2 == 0x20:
		return accQ("adda")
	case op2 == 0x21:
		return accQ("madda")
	case op2 == 0x22:
		return accI("adda")
	case op2 == 0x23:
		return accI("madda")
	case op2 == 0x24:
		return accQ("suba")
	case op2 == 0x25:
		return accQ("msuba")
	case op2 == 0x26:
		return accI("suba")
	case op2 == 0x27:
		return accI("msuba")
	case op2 == 0x28:
		return acc("adda")
	case op2 == 0x29:
		return acc("madda")
	case op2 == 0x2A:
		return acc("mula")
	case op2 == 0x2C:
		return acc("suba")
	case op2 == 0x2D:
		return acc("msuba")
	case op2 == 0x2E:
		return acc("opmula")
	case op2 == 0x2F:
		return "nop"
	}
	return fmt.Sprintf("upper? 0x%08X (wide 0x%02X)", i, op2)
}

// decodeLower renders the integer/branch/load-store half.
func decodeLower(i uint32, addr uint32) string {
	op7 := i >> 25 & 0x7F
	d, t, s := dest(i), ft(i), fs(i)
	imm11 := int32(i<<21) >> 21 // sign-extended 11-bit
	imm15 := i>>10&0x7800 | i&0x7FF
	branch := func(name string, args string) string {
		target := addr + 8 + uint32(imm11)*8
		if args != "" {
			args += ", "
		}
		return fmt.Sprintf("%s %s0x%X", name, args, target)
	}

	switch op7 {
	case 0x00:
		return fmt.Sprintf("lq.%s vf%02d, %d(vi%02d)", d, t, imm11, s)
	case 0x01:
		return fmt.Sprintf("sq.%s vf%02d, %d(vi%02d)", d, s, imm11, t)
	case 0x04:
		return fmt.Sprintf("ilw.%s vi%02d, %d(vi%02d)", d, t, imm11, s)
	case 0x05:
		return fmt.Sprintf("isw.%s vi%02d, %d(vi%02d)", d, t, imm11, s)
	case 0x08:
		return fmt.Sprintf("iaddiu vi%02d, vi%02d, %d", t, s, imm15)
	case 0x09:
		return fmt.Sprintf("isubiu vi%02d, vi%02d, %d", t, s, imm15)
	case 0x10:
		return fmt.Sprintf("fceq 0x%06X", i&0xFFFFFF)
	case 0x11:
		return fmt.Sprintf("fcset 0x%06X", i&0xFFFFFF)
	case 0x12:
		return fmt.Sprintf("fcand 0x%06X", i&0xFFFFFF)
	case 0x13:
		return fmt.Sprintf("fcor 0x%06X", i&0xFFFFFF)
	case 0x14:
		return fmt.Sprintf("fseq vi%02d, 0x%03X", t, i>>10&0x800|i&0x7FF)
	case 0x15:
		return fmt.Sprintf("fsset 0x%03X", i>>10&0x800|i&0x7FF)
	case 0x16:
		return fmt.Sprintf("fsand vi%02d, 0x%03X", t, i>>10&0x800|i&0x7FF)
	case 0x17:
		return fmt.Sprintf("fsor vi%02d, 0x%03X", t, i>>10&0x800|i&0x7FF)
	case 0x18:
		return fmt.Sprintf("fmeq vi%02d, vi%02d", t, s)
	case 0x1A:
		return fmt.Sprintf("fmand vi%02d, vi%02d", t, s)
	case 0x1B:
		return fmt.Sprintf("fmor vi%02d, vi%02d", t, s)
	case 0x1C:
		return fmt.Sprintf("fcget vi%02d", t)
	case 0x20:
		return branch("b", "")
	case 0x21:
		return branch("bal", fmt.Sprintf("vi%02d", t))
	case 0x24:
		return fmt.Sprintf("jr vi%02d", s)
	case 0x25:
		return fmt.Sprintf("jalr vi%02d, vi%02d", t, s)
	case 0x28:
		return branch("ibeq", fmt.Sprintf("vi%02d, vi%02d", t, s))
	case 0x29:
		return branch("ibne", fmt.Sprintf("vi%02d, vi%02d", t, s))
	case 0x2C:
		return branch("ibltz", fmt.Sprintf("vi%02d", s))
	case 0x2D:
		return branch("ibgtz", fmt.Sprintf("vi%02d", s))
	case 0x2E:
		return branch("iblez", fmt.Sprintf("vi%02d", s))
	case 0x2F:
		return branch("ibgez", fmt.Sprintf("vi%02d", s))
	case 0x40:
		return decodeLowerSpecial(i)
	}
	if i == 0 {
		return "nop"
	}
	return fmt.Sprintf("lower? 0x%08X", i)
}

// decodeLowerSpecial is the op7 == 0x40 group.
func decodeLowerSpecial(i uint32) string {
	d, t, s, f := dest(i), ft(i), fs(i), fd(i)
	op := i & 0x3F
	switch op {
	case 0x30:
		return fmt.Sprintf("iadd vi%02d, vi%02d, vi%02d", f, s, t)
	case 0x31:
		return fmt.Sprintf("isub vi%02d, vi%02d, vi%02d", f, s, t)
	case 0x32:
		return fmt.Sprintf("iaddi vi%02d, vi%02d, %d", t, s, int32(i<<21)>>27)
	case 0x34:
		return fmt.Sprintf("iand vi%02d, vi%02d, vi%02d", f, s, t)
	case 0x35:
		return fmt.Sprintf("ior vi%02d, vi%02d, vi%02d", f, s, t)
	}
	if op < 0x3C {
		if i == 0 {
			return "nop"
		}
		return fmt.Sprintf("lower? 0x%08X", i)
	}

	op2 := i>>4&0x7C | i&3
	fsf := bcNames[i>>21&3]
	ftf := bcNames[i>>23&3]
	switch op2 {
	case 0x30:
		return fmt.Sprintf("move.%s vf%02d, vf%02d", d, t, s)
	case 0x31:
		return fmt.Sprintf("mr32.%s vf%02d, vf%02d", d, t, s)
	case 0x34:
		return fmt.Sprintf("lqi.%s vf%02d, (vi%02d++)", d, t, s)
	case 0x35:
		return fmt.Sprintf("sqi.%s vf%02d, (vi%02d++)", d, s, t)
	case 0x36:
		return fmt.Sprintf("lqd.%s vf%02d, (--vi%02d)", d, t, s)
	case 0x37:
		return fmt.Sprintf("sqd.%s vf%02d, (--vi%02d)", d, s, t)
	case 0x38:
		return fmt.Sprintf("div q, vf%02d%s, vf%02d%s", s, fsf, t, ftf)
	case 0x39:
		return fmt.Sprintf("sqrt q, vf%02d%s", t, ftf)
	case 0x3A:
		return fmt.Sprintf("rsqrt q, vf%02d%s, vf%02d%s", s, fsf, t, ftf)
	case 0x3B:
		return "waitq"
	case 0x3C:
		return fmt.Sprintf("mtir vi%02d, vf%02d%s", t, s, fsf)
	case 0x3D:
		return fmt.Sprintf("mfir.%s vf%02d, vi%02d", d, t, s)
	case 0x3E:
		return fmt.Sprintf("ilwr.%s vi%02d, (vi%02d)", d, t, s)
	case 0x3F:
		return fmt.Sprintf("iswr.%s vi%02d, (vi%02d)", d, t, s)
	case 0x40:
		return "rnext"
	case 0x41:
		return fmt.Sprintf("rget.%s vf%02d, r", d, t)
	case 0x42:
		return fmt.Sprintf("rinit r, vf%02d%s", s, fsf)
	case 0x43:
		return fmt.Sprintf("rxor r, vf%02d%s", s, fsf)
	case 0x64:
		return fmt.Sprintf("mfp.%s vf%02d, p", d, t)
	case 0x68:
		return fmt.Sprintf("xtop vi%02d", t)
	case 0x69:
		return fmt.Sprintf("xitop vi%02d", t)
	case 0x6C:
		return fmt.Sprintf("xgkick vi%02d", s)
	case 0x70:
		return fmt.Sprintf("esadd p, vf%02d", s)
	case 0x71:
		return fmt.Sprintf("ersadd p, vf%02d", s)
	case 0x72:
		return fmt.Sprintf("eleng p, vf%02d", s)
	case 0x73:
		return fmt.Sprintf("erleng p, vf%02d", s)
	case 0x74:
		return fmt.Sprintf("eatanxy p, vf%02d", s)
	case 0x75:
		return fmt.Sprintf("eatanxz p, vf%02d", s)
	case 0x76:
		return fmt.Sprintf("esum p, vf%02d", s)
	case 0x78:
		return fmt.Sprintf("esqrt p, vf%02d%s", s, fsf)
	case 0x79:
		return fmt.Sprintf("ersqrt p, vf%02d%s", s, fsf)
	case 0x7A:
		return fmt.Sprintf("ercpr p, vf%02d%s", s, fsf)
	case 0x7B:
		return "waitp"
	case 0x7C:
		return fmt.Sprintf("esin p, vf%02d%s", s, fsf)
	case 0x7D:
		return fmt.Sprintf("eatan p, vf%02d%s", s, fsf)
	case 0x7E:
		return fmt.Sprintf("eexp p, vf%02d%s", s, fsf)
	}
	return fmt.Sprintf("lower? 0x%08X (special2 0x%02X)", i, op2)
}
