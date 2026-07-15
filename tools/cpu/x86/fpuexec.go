package x86

// This file executes the x87 FPU escape opcodes (0xD8–0xDF); fpu.go only decodes
// them for the disassembler. The model is an eight-deep register stack of
// extended-precision values, a top-of-stack pointer, and the control, status and
// tag words — the real 80387 stack machine, with one deliberate simplification:
// the registers hold Go float64 rather than true 80-bit extended values. That is
// the same choice most emulators make; the only place 53- vs 64-bit significand
// precision shows is the last handful of bits of a long transcendental chain, far
// below what a game's geometry or an asset pipeline depends on. FLD/FSTP m80 still
// round-trip through a real 80-bit encoding so a stored long double keeps its
// exponent range. If a future computation ever needs the extra mantissa bits, the
// stack element type is the single thing to widen.

import "math"

// FPUState is the x87 register file and status. St holds the eight PHYSICAL
// registers R0..R7; the programmer-visible ST(i) is R[(Top+i)&7] (see st/setst).
// Tag is the 2-bit tag per physical register (0=valid, 1=zero, 2=special, 3=empty).
// Stat carries the condition-code and exception bits but NOT the TOP field, which
// lives in Top and is folded in only when the status word is read (statusWord).
type FPUState struct {
	St   [8]float64
	Tag  [8]uint8
	Top  int
	Ctrl uint16
	Stat uint16
}

// tag values.
const (
	tagValid = iota
	tagZero
	tagSpecial
	tagEmpty
)

// finit resets the FPU to its power-on / FNINIT state: control word 0x037F (all
// exceptions masked, round-to-nearest, 64-bit precision), cleared status, an
// empty stack with TOP=0.
func (f *FPUState) finit() {
	f.Ctrl = 0x037F
	f.Stat = 0
	f.Top = 0
	for i := range f.Tag {
		f.Tag[i] = tagEmpty
		f.St[i] = 0
	}
}

func classify(v float64) uint8 {
	switch {
	case v == 0:
		return tagZero
	case math.IsNaN(v) || math.IsInf(v, 0):
		return tagSpecial
	default:
		return tagValid
	}
}

// physical register index of ST(i).
func (f *FPUState) phys(i int) int { return (f.Top + i) & 7 }

// st reads ST(i); setst writes it (updating the tag).
func (f *FPUState) st(i int) float64 { return f.St[f.phys(i)] }
func (f *FPUState) setst(i int, v float64) {
	p := f.phys(i)
	f.St[p] = v
	f.Tag[p] = classify(v)
}

// push decrements TOP and stores v as the new ST(0); pop marks ST(0) empty and
// increments TOP. Neither raises a stack fault — games do not depend on the x87
// stack-overflow/underflow trap, and every masked exception here would just be
// absorbed by the default handler anyway.
func (f *FPUState) push(v float64) {
	f.Top = (f.Top - 1) & 7
	f.St[f.Top] = v
	f.Tag[f.Top] = classify(v)
}
func (f *FPUState) pop() {
	f.Tag[f.Top] = tagEmpty
	f.Top = (f.Top + 1) & 7
}

// statusWord assembles the 16-bit status word with the live TOP folded into bits
// 11–13 (FNSTSW/FNSTENV report it there; internally we keep TOP in Top).
func (f *FPUState) statusWord() uint16 {
	return (f.Stat &^ (7 << 11)) | uint16(f.Top&7)<<11
}

// setCC sets the three primary condition-code bits C3,C2,C0 (C1 is handled per-op).
func (f *FPUState) setCC(c3, c2, c0 bool) {
	f.Stat &^= (1 << 8) | (1 << 10) | (1 << 14) // C0, C2, C3
	if c0 {
		f.Stat |= 1 << 8
	}
	if c2 {
		f.Stat |= 1 << 10
	}
	if c3 {
		f.Stat |= 1 << 14
	}
}

// round applies the control word's rounding mode (RC, bits 10–11) to v, giving
// the integer used by FRNDINT and the FIST family.
func (f *FPUState) round(v float64) float64 {
	switch (f.Ctrl >> 10) & 3 {
	case 0:
		return math.RoundToEven(v)
	case 1:
		return math.Floor(v)
	case 2:
		return math.Ceil(v)
	default:
		return math.Trunc(v)
	}
}

// --- memory operand read/write (mode-aware via memRead/memWrite) ---

func (c *CPU) fLoad32(o ea) float64 { return float64(math.Float32frombits(c.memRead(o.base, o.off, 4))) }
func (c *CPU) fStore32(o ea, v float64) {
	c.memWrite(o.base, o.off, 4, math.Float32bits(float32(v)))
}
func (c *CPU) fLoad64(o ea) float64 {
	lo := uint64(c.memRead(o.base, o.off, 4))
	hi := uint64(c.memRead(o.base, o.off+4, 4))
	return math.Float64frombits(lo | hi<<32)
}
func (c *CPU) fStore64(o ea, v float64) {
	b := math.Float64bits(v)
	c.memWrite(o.base, o.off, 4, uint32(b))
	c.memWrite(o.base, o.off+4, 4, uint32(b>>32))
}
func (c *CPU) fLoad80(o ea) float64 {
	mant := uint64(c.memRead(o.base, o.off, 4)) | uint64(c.memRead(o.base, o.off+4, 4))<<32
	se := uint16(c.memRead(o.base, o.off+8, 2))
	return f80ToF64(mant, se)
}
func (c *CPU) fStore80(o ea, v float64) {
	mant, se := f64ToF80(v)
	c.memWrite(o.base, o.off, 4, uint32(mant))
	c.memWrite(o.base, o.off+4, 4, uint32(mant>>32))
	c.memWrite(o.base, o.off+8, 2, uint32(se))
}

// f80ToF64 decodes an 80-bit extended value (explicit-integer-bit mantissa +
// 15-bit biased exponent + sign) to float64. f64ToF80 is the inverse.
func f80ToF64(mant uint64, se uint16) float64 {
	sign := se >> 15
	exp := int(se & 0x7FFF)
	neg := func(x float64) float64 {
		if sign == 1 {
			return -x
		}
		return x
	}
	switch {
	case exp == 0 && mant == 0:
		return neg(0)
	case exp == 0x7FFF:
		if mant<<1 == 0 { // integer bit only -> infinity
			return neg(math.Inf(1))
		}
		return math.NaN()
	}
	// value = mant * 2^(exp - 16383 - 63)
	return neg(math.Ldexp(float64(mant), exp-16446))
}

func f64ToF80(v float64) (uint64, uint16) {
	bits := math.Float64bits(v)
	sign := uint16(bits>>63) & 1
	switch {
	case v == 0:
		return 0, sign << 15
	case math.IsInf(v, 0):
		return 0x8000000000000000, sign<<15 | 0x7FFF
	case math.IsNaN(v):
		return 0xC000000000000000, sign<<15 | 0x7FFF
	}
	// Frexp normalises subnormals too: v = m * 2^e, m in [0.5,1).
	m, e := math.Frexp(math.Abs(v))
	mant := uint64(m * math.Exp2(64)) // top bit set (m>=0.5); value = mant*2^(e-64)
	e80 := e + 16382                  // exp-16446 == e-64  ->  exp = e+16382
	return mant, sign<<15 | uint16(e80)
}

// --- dispatch ---

// fpuExec executes one x87 escape opcode (op is 0xD8..0xDF). It re-decodes the
// ModR/M itself (rather than going through modrmE) because the register forms
// need the whole byte to pick the operation, and the memory forms use the reg
// field as the sub-opcode.
func (c *CPU) fpuExec(op byte) {
	mb := byte(c.fetch8())
	mod, reg, rm := mb>>6, (mb>>3)&7, mb&7
	if mod != 3 {
		var o ea
		if c.dAddrsize == 32 {
			o = c.ea32(mod, rm)
		} else {
			o = c.ea16(mod, rm)
		}
		c.fpuMemExec(op, reg, o)
		return
	}
	c.fpuRegExec(op, mb, reg, rm)
}

// arithST computes ST(0) <op> operand for the D8/DC/DA/DE grid where ST(0) is the
// destination (sub is the reg field 0..7). Returns (result, isCompare).
func arithST(sub int, a, b float64) (float64, bool) {
	switch sub {
	case 0:
		return a + b, false // FADD
	case 1:
		return a * b, false // FMUL
	case 2, 3:
		return 0, true // FCOM / FCOMP
	case 4:
		return a - b, false // FSUB
	case 5:
		return b - a, false // FSUBR
	case 6:
		return a / b, false // FDIV
	default:
		return b / a, false // FDIVR
	}
}

// fcom sets C3,C2,C0 from comparing a to b (C2=1 means unordered).
func (f *FPUState) fcom(a, b float64) {
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		f.setCC(true, true, true)
	case a > b:
		f.setCC(false, false, false)
	case a < b:
		f.setCC(false, false, true)
	default:
		f.setCC(true, false, false)
	}
}

// fcomi sets the integer EFLAGS (ZF,PF,CF) from comparing a to b — the FCOMI /
// FUCOMI form that feeds ordinary conditional branches. OF,SF,AF are cleared.
func (c *CPU) fcomi(a, b float64) {
	c.OF, c.SF, c.AF = false, false, false
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		c.ZF, c.PF, c.CF = true, true, true
	case a > b:
		c.ZF, c.PF, c.CF = false, false, false
	case a < b:
		c.ZF, c.PF, c.CF = false, false, true
	default:
		c.ZF, c.PF, c.CF = true, false, false
	}
}

// fpuMemExec handles the memory-operand forms; reg is the sub-opcode.
func (c *CPU) fpuMemExec(op, reg byte, o ea) {
	f := &c.FPU
	switch op {
	case 0xD8: // arith ST(0), m32real
		v := c.fLoad32(o)
		if res, cmp := arithST(int(reg), f.st(0), v); cmp {
			f.fcom(f.st(0), v)
			if reg == 3 {
				f.pop()
			}
		} else {
			f.setst(0, res)
		}
	case 0xDC: // arith ST(0), m64real
		v := c.fLoad64(o)
		if res, cmp := arithST(int(reg), f.st(0), v); cmp {
			f.fcom(f.st(0), v)
			if reg == 3 {
				f.pop()
			}
		} else {
			f.setst(0, res)
		}
	case 0xDA: // arith ST(0), m32int
		v := float64(int32(c.memRead(o.base, o.off, 4)))
		if res, cmp := arithST(int(reg), f.st(0), v); cmp {
			f.fcom(f.st(0), v)
			if reg == 3 {
				f.pop()
			}
		} else {
			f.setst(0, res)
		}
	case 0xDE: // arith ST(0), m16int
		v := float64(int16(uint16(c.memRead(o.base, o.off, 2))))
		if res, cmp := arithST(int(reg), f.st(0), v); cmp {
			f.fcom(f.st(0), v)
			if reg == 3 {
				f.pop()
			}
		} else {
			f.setst(0, res)
		}

	case 0xD9:
		switch reg {
		case 0: // FLD m32real
			f.push(c.fLoad32(o))
		case 2: // FST m32real
			c.fStore32(o, f.st(0))
		case 3: // FSTP m32real
			c.fStore32(o, f.st(0))
			f.pop()
		case 4: // FLDENV m14/28
			c.fldenv(o)
		case 5: // FLDCW m16
			f.Ctrl = uint16(c.memRead(o.base, o.off, 2))
		case 6: // FNSTENV m14/28
			c.fnstenv(o)
		case 7: // FNSTCW m16
			c.memWrite(o.base, o.off, 2, uint32(f.Ctrl))
		}

	case 0xDB:
		switch reg {
		case 0: // FILD m32int
			f.push(float64(int32(c.memRead(o.base, o.off, 4))))
		case 1: // FISTTP m32int (truncate)
			c.fistStore(o, 4, math.Trunc(f.st(0)))
			f.pop()
		case 2: // FIST m32int
			c.fistStore(o, 4, f.round(f.st(0)))
		case 3: // FISTP m32int
			c.fistStore(o, 4, f.round(f.st(0)))
			f.pop()
		case 5: // FLD m80real
			f.push(c.fLoad80(o))
		case 7: // FSTP m80real
			c.fStore80(o, f.st(0))
			f.pop()
		}

	case 0xDD:
		switch reg {
		case 0: // FLD m64real
			f.push(c.fLoad64(o))
		case 1: // FISTTP m64int
			c.fistStore(o, 8, math.Trunc(f.st(0)))
			f.pop()
		case 2: // FST m64real
			c.fStore64(o, f.st(0))
		case 3: // FSTP m64real
			c.fStore64(o, f.st(0))
			f.pop()
		case 4: // FRSTOR
			c.frstor(o)
		case 6: // FNSAVE
			c.fnsave(o)
		case 7: // FNSTSW m16
			c.memWrite(o.base, o.off, 2, uint32(f.statusWord()))
		}

	case 0xDF:
		switch reg {
		case 0: // FILD m16int
			f.push(float64(int16(uint16(c.memRead(o.base, o.off, 2)))))
		case 1: // FISTTP m16int
			c.fistStore(o, 2, math.Trunc(f.st(0)))
			f.pop()
		case 2: // FIST m16int
			c.fistStore(o, 2, f.round(f.st(0)))
		case 3: // FISTP m16int
			c.fistStore(o, 2, f.round(f.st(0)))
			f.pop()
		case 4: // FBLD m80bcd — packed-BCD load, rare; treat as unsupported no-op push
			f.push(0)
		case 5: // FILD m64int
			lo := uint64(c.memRead(o.base, o.off, 4)) | uint64(c.memRead(o.base, o.off+4, 4))<<32
			f.push(float64(int64(lo)))
		case 6: // FBSTP m80bcd — rare; store zero, pop
			for i := uint32(0); i < 10; i++ {
				c.wr8(c.linear(o.base, o.off+i), 0)
			}
			f.pop()
		case 7: // FISTP m64int
			c.fistStore(o, 8, f.round(f.st(0)))
			f.pop()
		}
	}
}

// fistStore writes rounded integer iv into an n-byte (2/4/8) memory slot.
func (c *CPU) fistStore(o ea, n int, fv float64) {
	iv := int64(fv)
	// out-of-range -> integer indefinite (the most-negative value of the width).
	switch n {
	case 2:
		if fv < -32768 || fv > 32767 {
			iv = -32768
		}
		c.memWrite(o.base, o.off, 2, uint32(uint16(int16(iv))))
	case 4:
		if fv < -2147483648 || fv > 2147483647 {
			iv = -2147483648
		}
		c.memWrite(o.base, o.off, 4, uint32(int32(iv)))
	default: // 8
		c.memWrite(o.base, o.off, 4, uint32(iv))
		c.memWrite(o.base, o.off+4, 4, uint32(uint64(iv)>>32))
	}
}

// fpuRegExec handles the mod==3 register/no-operand forms. mb is the full ModR/M
// byte, reg its middle field, rm the ST index.
func (c *CPU) fpuRegExec(op, mb, reg, rm byte) {
	f := &c.FPU
	i := int(rm)
	switch op {
	case 0xD8: // arith ST(0), ST(i)
		v := f.st(i)
		if res, cmp := arithST(int(reg), f.st(0), v); cmp {
			f.fcom(f.st(0), v)
			if reg == 3 {
				f.pop()
			}
		} else {
			f.setst(0, res)
		}
		return
	case 0xDC: // arith ST(i), ST(0) — ST(i) is destination
		if reg == 2 || reg == 3 { // FCOM/FCOMP encodings alias here on some assemblers
			f.fcom(f.st(0), f.st(i))
			if reg == 3 {
				f.pop()
			}
			return
		}
		f.setst(i, dstArith(int(reg), f.st(i), f.st(0)))
		return
	case 0xDE: // arith-and-pop ST(i), ST(0)
		if mb == 0xD9 { // FCOMPP
			f.fcom(f.st(0), f.st(1))
			f.pop()
			f.pop()
			return
		}
		if reg == 2 || reg == 3 {
			f.fcom(f.st(0), f.st(i))
			f.pop()
			if reg == 3 {
				f.pop()
			}
			return
		}
		f.setst(i, dstArith(int(reg), f.st(i), f.st(0)))
		f.pop()
		return
	}

	// The remaining pages are keyed on the whole byte.
	switch op {
	case 0xD9:
		c.fpuD9(mb, i)
	case 0xDA:
		c.fpuDA(mb, i)
	case 0xDB:
		c.fpuDB(mb, i)
	case 0xDD:
		c.fpuDD(mb, i)
	case 0xDF:
		c.fpuDF(mb, i)
	}
}

// dstArith computes dst <op> src for the DC/DE grid (dst = ST(i), src = ST(0)),
// honouring the reversed FSUBR/FSUB and FDIVR/FDIV encoding order.
func dstArith(sub int, dst, src float64) float64 {
	switch sub {
	case 0:
		return dst + src // FADD(P)
	case 1:
		return dst * src // FMUL(P)
	case 4:
		return src - dst // FSUBR(P):  ST(i) <- ST0 - ST(i)
	case 5:
		return dst - src // FSUB(P):   ST(i) <- ST(i) - ST0
	case 6:
		return src / dst // FDIVR(P):  ST(i) <- ST0 / ST(i)
	default:
		return dst / src // FDIV(P):   ST(i) <- ST(i) / ST0
	}
}

// fpuD9 executes the D9 register page: FLD/FXCH ST(i), the constant loads, and
// the unary/transcendental no-operand ops.
func (c *CPU) fpuD9(mb byte, i int) {
	f := &c.FPU
	switch {
	case mb >= 0xC0 && mb <= 0xC7: // FLD ST(i)
		f.push(f.st(i))
		return
	case mb >= 0xC8 && mb <= 0xCF: // FXCH ST(i)
		p0, pi := f.phys(0), f.phys(i)
		f.St[p0], f.St[pi] = f.St[pi], f.St[p0]
		f.Tag[p0], f.Tag[pi] = f.Tag[pi], f.Tag[p0]
		return
	}
	switch mb {
	case 0xD0: // FNOP
	case 0xE0: // FCHS
		f.setst(0, -f.st(0))
	case 0xE1: // FABS
		f.setst(0, math.Abs(f.st(0)))
	case 0xE4: // FTST
		f.fcom(f.st(0), 0)
	case 0xE5: // FXAM
		c.fxam()
	case 0xE8: // FLD1
		f.push(1)
	case 0xE9: // FLDL2T
		f.push(math.Log2(10))
	case 0xEA: // FLDL2E
		f.push(math.Log2(math.E))
	case 0xEB: // FLDPI
		f.push(math.Pi)
	case 0xEC: // FLDLG2
		f.push(math.Log10(2))
	case 0xED: // FLDLN2
		f.push(math.Ln2)
	case 0xEE: // FLDZ
		f.push(0)
	case 0xF0: // F2XM1
		f.setst(0, math.Exp2(f.st(0))-1)
	case 0xF1: // FYL2X: ST(1) <- ST(1)*log2(ST(0)); pop
		f.setst(1, f.st(1)*math.Log2(f.st(0)))
		f.pop()
	case 0xF2: // FPTAN: ST(0) <- tan(ST(0)); push 1
		f.setst(0, math.Tan(f.st(0)))
		f.push(1)
		f.Stat &^= 1 << 10 // C2=0: reduction complete
	case 0xF3: // FPATAN: ST(1) <- atan2(ST(1),ST(0)); pop
		f.setst(1, math.Atan2(f.st(1), f.st(0)))
		f.pop()
	case 0xF4: // FXTRACT: ST(0) <- exponent; push significand
		c.fxtract()
	case 0xF5: // FPREM1 (IEEE remainder)
		c.fprem(true)
	case 0xF6: // FDECSTP
		f.Top = (f.Top - 1) & 7
	case 0xF7: // FINCSTP
		f.Top = (f.Top + 1) & 7
	case 0xF8: // FPREM
		c.fprem(false)
	case 0xF9: // FYL2XP1: ST(1) <- ST(1)*log2(ST(0)+1); pop
		f.setst(1, f.st(1)*math.Log2(f.st(0)+1))
		f.pop()
	case 0xFA: // FSQRT
		f.setst(0, math.Sqrt(f.st(0)))
	case 0xFB: // FSINCOS: ST(0) <- sin; push cos
		x := f.st(0)
		f.setst(0, math.Sin(x))
		f.push(math.Cos(x))
		f.Stat &^= 1 << 10
	case 0xFC: // FRNDINT
		f.setst(0, f.round(f.st(0)))
	case 0xFD: // FSCALE: ST(0) <- ST(0)*2^trunc(ST(1))
		f.setst(0, math.Ldexp(f.st(0), int(math.Trunc(f.st(1)))))
	case 0xFE: // FSIN
		f.setst(0, math.Sin(f.st(0)))
		f.Stat &^= 1 << 10
	case 0xFF: // FCOS
		f.setst(0, math.Cos(f.st(0)))
		f.Stat &^= 1 << 10
	}
}

// fxam classifies ST(0) into C3,C2,C0 and puts the sign in C1.
func (c *CPU) fxam() {
	f := &c.FPU
	v := f.st(0)
	sign := math.Signbit(v)
	var c3, c2, c0 bool
	switch {
	case f.Tag[f.phys(0)] == tagEmpty:
		c3, c2, c0 = true, false, true // empty
	case math.IsNaN(v):
		c3, c2, c0 = false, false, true // NaN
	case math.IsInf(v, 0):
		c3, c2, c0 = false, true, true // infinity
	case v == 0:
		c3, c2, c0 = true, false, false // zero
	default:
		c3, c2, c0 = false, true, false // normal finite
	}
	f.setCC(c3, c2, c0)
	f.Stat &^= 1 << 9
	if sign {
		f.Stat |= 1 << 9 // C1 = sign
	}
}

// fxtract splits ST(0) into a floating exponent and significand.
func (c *CPU) fxtract() {
	f := &c.FPU
	x := f.st(0)
	if x == 0 || math.IsInf(x, 0) || math.IsNaN(x) {
		f.push(x) // degenerate: leave value, push a copy
		return
	}
	exp := math.Floor(math.Log2(math.Abs(x)))
	sig := x / math.Exp2(exp)
	f.setst(0, exp)
	f.push(sig)
}

// fprem computes the partial remainder ST(0) mod ST(1). ieee selects FPREM1
// (round-to-nearest quotient) vs FPREM (truncated quotient). We reduce fully in
// one step, so C2 (incomplete) is cleared and the low quotient bits go to C0/C3/C1.
func (c *CPU) fprem(ieee bool) {
	f := &c.FPU
	a, b := f.st(0), f.st(1)
	var r, q float64
	if ieee {
		r = math.Remainder(a, b)
		q = math.RoundToEven(a / b)
	} else {
		r = math.Mod(a, b)
		q = math.Trunc(a / b)
	}
	f.setst(0, r)
	qi := int64(math.Abs(q))
	f.Stat &^= (1 << 8) | (1 << 9) | (1 << 10) | (1 << 14) // clear C0,C1,C2,C3
	if qi&1 != 0 {
		f.Stat |= 1 << 9 // C1 = q0
	}
	if qi&2 != 0 {
		f.Stat |= 1 << 14 // C3 = q1
	}
	if qi&4 != 0 {
		f.Stat |= 1 << 8 // C0 = q2
	}
}

// fpuDA: FCMOVcc ST(0),ST(i) on carry/zero, and FUCOMPP.
func (c *CPU) fpuDA(mb byte, i int) {
	f := &c.FPU
	if mb == 0xE9 { // FUCOMPP
		f.fcom(f.st(0), f.st(1))
		f.pop()
		f.pop()
		return
	}
	var move bool
	switch mb & 0xF8 {
	case 0xC0: // FCMOVB   (CF=1)
		move = c.CF
	case 0xC8: // FCMOVE   (ZF=1)
		move = c.ZF
	case 0xD0: // FCMOVBE  (CF|ZF)
		move = c.CF || c.ZF
	case 0xD8: // FCMOVU   (PF=1)
		move = c.PF
	}
	if move {
		f.setst(0, f.st(i))
	}
}

// fpuDB: FCMOVNcc, FNCLEX, FNINIT, and FUCOMI/FCOMI ST(0),ST(i).
func (c *CPU) fpuDB(mb byte, i int) {
	f := &c.FPU
	switch mb {
	case 0xE2: // FNCLEX — clear exception + busy bits
		f.Stat &^= 0x80FF
		return
	case 0xE3: // FNINIT
		f.finit()
		return
	}
	switch mb & 0xF8 {
	case 0xC0, 0xC8, 0xD0, 0xD8: // FCMOVN{B,E,BE,U}
		var move bool
		switch mb & 0xF8 {
		case 0xC0:
			move = !c.CF
		case 0xC8:
			move = !c.ZF
		case 0xD0:
			move = !(c.CF || c.ZF)
		case 0xD8:
			move = !c.PF
		}
		if move {
			f.setst(0, f.st(i))
		}
	case 0xE8: // FUCOMI ST(0),ST(i)
		c.fcomi(f.st(0), f.st(i))
	case 0xF0: // FCOMI ST(0),ST(i)
		c.fcomi(f.st(0), f.st(i))
	}
}

// fpuDD: FFREE / FST / FSTP / FUCOM / FUCOMP on ST(i).
func (c *CPU) fpuDD(mb byte, i int) {
	f := &c.FPU
	switch mb & 0xF8 {
	case 0xC0: // FFREE ST(i)
		f.Tag[f.phys(i)] = tagEmpty
	case 0xD0: // FST ST(i)
		f.setst(i, f.st(0))
	case 0xD8: // FSTP ST(i)
		f.setst(i, f.st(0))
		f.pop()
	case 0xE0: // FUCOM ST(i)
		f.fcom(f.st(0), f.st(i))
	case 0xE8: // FUCOMP ST(i)
		f.fcom(f.st(0), f.st(i))
		f.pop()
	}
}

// fpuDF: FNSTSW AX, FFREEP, and FUCOMIP/FCOMIP ST(0),ST(i).
func (c *CPU) fpuDF(mb byte, i int) {
	f := &c.FPU
	if mb == 0xE0 { // FNSTSW AX
		c.s16(AX, uint32(f.statusWord()))
		return
	}
	switch mb & 0xF8 {
	case 0xC0: // FFREEP ST(i) — free then pop
		f.Tag[f.phys(i)] = tagEmpty
		f.pop()
	case 0xE8: // FUCOMIP
		c.fcomi(f.st(0), f.st(i))
		f.pop()
	case 0xF0: // FCOMIP
		c.fcomi(f.st(0), f.st(i))
		f.pop()
	}
}

// --- environment / full-state save & restore ---
//
// FLDENV/FNSTENV move the 14- or 28-byte environment (control, status, tag words
// and the instruction/operand pointers); FNSAVE/FRSTOR add the eight 80-bit
// register slots. We honour the operand size for the 16- vs 32-bit protected-mode
// layout and preserve control/status/tag; the instruction/data pointers, which no
// game reads back, are written as zero.

func (c *CPU) envIs32() bool { return c.dOpsize == 32 }

func (c *CPU) fnstenv(o ea) {
	f := &c.FPU
	c.envWriteHeader(o, f)
}
func (c *CPU) fldenv(o ea) {
	f := &c.FPU
	c.envReadHeader(o, f)
}
func (c *CPU) fnsave(o ea) {
	f := &c.FPU
	c.envWriteHeader(o, f)
	base := uint32(14)
	if c.envIs32() {
		base = 28
	}
	for i := 0; i < 8; i++ { // stored in physical order R0..R7
		mant, se := f64ToF80(f.St[i])
		off := o.off + base + uint32(i)*10
		c.memWrite(o.base, off, 4, uint32(mant))
		c.memWrite(o.base, off+4, 4, uint32(mant>>32))
		c.memWrite(o.base, off+8, 2, uint32(se))
	}
	f.finit() // FNSAVE reinitialises the FPU
}
func (c *CPU) frstor(o ea) {
	f := &c.FPU
	c.envReadHeader(o, f)
	base := uint32(14)
	if c.envIs32() {
		base = 28
	}
	for i := 0; i < 8; i++ {
		off := o.off + base + uint32(i)*10
		mant := uint64(c.memRead(o.base, off, 4)) | uint64(c.memRead(o.base, off+4, 4))<<32
		se := uint16(c.memRead(o.base, off+8, 2))
		f.St[i] = f80ToF64(mant, se)
	}
}

func (c *CPU) envWriteHeader(o ea, f *FPUState) {
	if c.envIs32() {
		c.memWrite(o.base, o.off, 4, uint32(f.Ctrl))
		c.memWrite(o.base, o.off+4, 4, uint32(f.statusWord()))
		c.memWrite(o.base, o.off+8, 4, uint32(f.tagWord()))
		for i := uint32(12); i < 28; i += 4 { // instruction/operand pointers -> 0
			c.memWrite(o.base, o.off+i, 4, 0)
		}
	} else {
		c.memWrite(o.base, o.off, 2, uint32(f.Ctrl))
		c.memWrite(o.base, o.off+2, 2, uint32(f.statusWord()))
		c.memWrite(o.base, o.off+4, 2, uint32(f.tagWord()))
		for i := uint32(6); i < 14; i += 2 {
			c.memWrite(o.base, o.off+i, 2, 0)
		}
	}
}

func (c *CPU) envReadHeader(o ea, f *FPUState) {
	if c.envIs32() {
		f.Ctrl = uint16(c.memRead(o.base, o.off, 4))
		f.Stat = uint16(c.memRead(o.base, o.off+4, 4))
		f.setTagWord(uint16(c.memRead(o.base, o.off+8, 4)))
	} else {
		f.Ctrl = uint16(c.memRead(o.base, o.off, 2))
		f.Stat = uint16(c.memRead(o.base, o.off+2, 2))
		f.setTagWord(uint16(c.memRead(o.base, o.off+4, 2)))
	}
	f.Top = int(f.Stat>>11) & 7
	f.Stat &^= 7 << 11
}

// tagWord packs the eight 2-bit physical tags; setTagWord unpacks them.
func (f *FPUState) tagWord() uint16 {
	var w uint16
	for i := 0; i < 8; i++ {
		w |= uint16(f.Tag[i]&3) << (uint(i) * 2)
	}
	return w
}
func (f *FPUState) setTagWord(w uint16) {
	for i := 0; i < 8; i++ {
		f.Tag[i] = uint8(w>>(uint(i)*2)) & 3
	}
}
