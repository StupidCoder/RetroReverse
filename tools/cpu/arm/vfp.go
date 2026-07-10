package arm

import (
	"fmt"
	"math"
)

// vfp.go implements VFPv2 — the ARM11's hardware floating-point unit, and so the
// 3DS's. The DS ARM cores have no FPU, which is why this lives behind the V6K
// variant gate: VFP instructions are coprocessor-10 (single precision) and
// coprocessor-11 (double precision) accesses, and on a V5TE core those encodings
// are undefined.
//
// The register file is the VFPv2 shape: 32 single-precision registers S0-S31,
// which pair up as 16 double-precision registers D0-D15 (Dn occupies S[2n] as its
// low word and S[2n+1] as its high). FPSCR carries the N/Z/C/V comparison flags
// (VCMP writes them; VMRS APSR_nzcv copies them to the CPSR flags) and the
// rounding-mode/exception bits, which this core keeps but does not act on beyond
// the default round-to-nearest that Go's float arithmetic already provides.
//
// The four encoding groups, all with coprocessor number 10 or 11:
//
//	load/store (bits 27:25 = 110)  VLDR VSTR VLDM VSTM VPUSH VPOP
//	data-proc  (bits 27:24 = 1110, bit4 = 0)  VADD VSUB VMUL VDIV VMLA … VCMP VCVT
//	1-reg move (bits 27:24 = 1110, bit4 = 1)  VMOV(core↔s) VMSR VMRS
//	2-reg move (bits 27:21 = 1100010)         VMOV(2×core↔d)
//
// Instructions outside the implemented subset halt with their encoding, so a gap
// is explicit rather than a silently wrong float.

// vfpState is the coprocessor register file, embedded in CPU.
type vfpState struct {
	S     [32]uint32 // S0-S31; Dn = {S[2n] (low), S[2n+1] (high)}
	FPSCR uint32
	FPEXC uint32 // only the EN bit is modelled (VFP enabled)
}

// --- register access -------------------------------------------------------

func (c *CPU) sGet(n uint32) float32 { return math.Float32frombits(c.VFP.S[n]) }
func (c *CPU) sSet(n uint32, v float32) { c.VFP.S[n] = math.Float32bits(v) }
func (c *CPU) sBits(n uint32) uint32    { return c.VFP.S[n] }
func (c *CPU) sSetBits(n, v uint32)     { c.VFP.S[n] = v }

func (c *CPU) dGet(n uint32) float64 {
	return math.Float64frombits(uint64(c.VFP.S[2*n]) | uint64(c.VFP.S[2*n+1])<<32)
}
func (c *CPU) dSet(n uint32, v float64) {
	b := math.Float64bits(v)
	c.VFP.S[2*n] = uint32(b)
	c.VFP.S[2*n+1] = uint32(b >> 32)
}

// Register-number assembly differs between precisions: a single-precision number
// puts the extra bit low (Sn = Vn<<1 | bit), a double-precision number puts it
// high (Dn = bit<<4 | Vn).
func sReg(v, bit uint32) uint32 { return v<<1 | bit }
func dReg(v, bit uint32) uint32 { return bit<<4 | v }

// isVFP reports whether an ARM word is a VFP coprocessor access (cp 10 or 11).
func isVFP(w uint32) bool {
	cp := (w >> 8) & 0xF
	return cp == 10 || cp == 11
}

// --- disassembly -----------------------------------------------------------

// decodeVFP renders a VFP instruction for a listing. It names the common
// load/store, arithmetic, compare, convert and transfer forms; anything it does
// not individually name falls back to a "VFP" placeholder carrying the word, so a
// listing never silently mis-prints a float instruction as something else.
func decodeVFP(w uint32, in Inst) Inst {
	single := (w>>8)&0xF == 10
	prec := "F64"
	if single {
		prec = "F32"
	}
	sfx := cn(in.Cond)

	switch {
	case (w>>25)&7 == 0b110: // load/store or 64-bit move
		if (w>>21)&0x7F == 0b1100010 || ((w>>23)&0x1F == 0b11000 && (w>>21)&3 == 2) {
			in.Mnem = "VMOV" + sfx
			in.Text = fmt.Sprintf("%s (64-bit move) ; 0x%08X", in.Mnem, w)
			return in
		}
		p, wb, l := (w>>24)&1, (w>>21)&1, (w>>20)&1
		rn := regName[(w>>16)&0xF]
		vd := vfpRegName(single, (w>>12)&0xF, (w>>22)&1)
		if p == 1 && wb == 0 { // VLDR / VSTR
			name := "VSTR"
			if l == 1 {
				name = "VLDR"
			}
			off := (w & 0xFF) << 2
			sign := "+"
			if (w>>23)&1 == 0 {
				sign = "-"
			}
			in.Mnem = name + sfx
			in.Text = fmt.Sprintf("%s %s, [%s, #%s0x%X]", in.Mnem, vd, rn, sign, off)
			return in
		}
		name := "VSTM"
		if l == 1 {
			name = "VLDM"
		}
		in.Mnem = name + sfx
		in.Text = fmt.Sprintf("%s %s%s, %s{%d regs}", in.Mnem, rn, wbMark(wb), "", w&0xFF)
		return in

	case (w>>24)&0xF == 0b1110:
		if (w>>4)&1 == 1 { // core↔VFP transfer / VMRS / VMSR
			return decodeVFPMove(w, in)
		}
		return decodeVFPData(w, in, single, prec, sfx)
	}
	in.Mnem = "VFP?"
	in.Text = fmt.Sprintf("VFP? 0x%08X", w)
	return in
}

func decodeVFPMove(w uint32, in Inst) Inst {
	l := (w >> 20) & 1
	rt := regName[(w>>12)&0xF]
	if (w>>21)&7 == 0b111 { // VMRS / VMSR
		if l == 1 {
			in.Mnem = "VMRS" + cn(in.Cond)
			dst := rt
			if (w>>12)&0xF == 15 {
				dst = "APSR_nzcv"
			}
			in.Text = fmt.Sprintf("%s %s, FPSCR", in.Mnem, dst)
		} else {
			in.Mnem = "VMSR" + cn(in.Cond)
			in.Text = fmt.Sprintf("%s FPSCR, %s", in.Mnem, rt)
		}
		return in
	}
	sn := vfpRegName(true, (w>>16)&0xF, (w>>7)&1)
	in.Mnem = "VMOV" + cn(in.Cond)
	if l == 1 {
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, rt, sn)
	} else {
		in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, sn, rt)
	}
	return in
}

func decodeVFPData(w uint32, in Inst, single bool, prec, sfx string) Inst {
	pqr := (w>>23)&1<<2 | (w>>21)&1<<1 | (w>>20)&1
	op := (w >> 6) & 1
	vd := vfpRegName(single, (w>>12)&0xF, (w>>22)&1)
	vn := vfpRegName(single, (w>>16)&0xF, (w>>7)&1)
	vm := vfpRegName(single, w&0xF, (w>>5)&1)

	name := ""
	switch pqr {
	case 0b011:
		if op == 0 {
			name = "VADD"
		} else {
			name = "VSUB"
		}
	case 0b010:
		if op == 0 {
			name = "VMUL"
		} else {
			name = "VNMUL"
		}
	case 0b100:
		name = "VDIV"
	case 0b000:
		if op == 0 {
			name = "VMLA"
		} else {
			name = "VMLS"
		}
	case 0b001:
		if op == 0 {
			name = "VNMLS"
		} else {
			name = "VNMLA"
		}
	case 0b111:
		return decodeVFPExt(w, in, single, prec, sfx, vd, vm)
	}
	if name == "" {
		in.Mnem, in.Text = "VFP?", fmt.Sprintf("VFP? 0x%08X", w)
		return in
	}
	in.Mnem = name + sfx + "." + prec
	in.Text = fmt.Sprintf("%s %s, %s, %s", in.Mnem, vd, vn, vm)
	return in
}

func decodeVFPExt(w uint32, in Inst, single bool, prec, sfx, vd, vm string) Inst {
	opc2 := (w >> 16) & 0xF
	opc3 := (w >> 6) & 3
	if opc3&1 == 0 { // VMOV immediate
		in.Mnem = "VMOV" + sfx + "." + prec
		in.Text = fmt.Sprintf("%s %s, #imm", in.Mnem, vd)
		return in
	}
	name := ""
	switch opc2 {
	case 0b0000:
		if opc3 == 0b01 {
			name = "VMOV"
		} else {
			name = "VABS"
		}
	case 0b0001:
		if opc3 == 0b01 {
			name = "VNEG"
		} else {
			name = "VSQRT"
		}
	case 0b0100, 0b0101:
		name = "VCMP"
	case 0b0111:
		name = "VCVT"
	case 0b1000, 0b1100, 0b1101:
		name = "VCVT"
	}
	if name == "" {
		in.Mnem, in.Text = "VFP?", fmt.Sprintf("VFP? 0x%08X", w)
		return in
	}
	in.Mnem = name + sfx + "." + prec
	in.Text = fmt.Sprintf("%s %s, %s", in.Mnem, vd, vm)
	return in
}

func vfpRegName(single bool, v, bit uint32) string {
	if single {
		return fmt.Sprintf("s%d", sReg(v, bit))
	}
	return fmt.Sprintf("d%d", dReg(v, bit))
}

func wbMark(wb uint32) string {
	if wb == 1 {
		return "!"
	}
	return ""
}

// --- execution -------------------------------------------------------------

// execVFP runs a VFP instruction, returning false if the encoding is not VFP (so
// the caller falls through). Only reached on the V6K variant.
func (c *CPU) execVFP(w uint32) bool {
	if !isVFP(w) {
		return false
	}
	switch {
	case (w>>25)&7 == 0b110: // load/store, or 64-bit move
		if (w>>21)&0x7F == 0b1100010 || (w>>23)&0x1F == 0b11000 && (w>>21)&3 == 2 {
			return c.execVFPMove64(w)
		}
		c.execVFPLoadStore(w)
		return true
	case (w>>24)&0xF == 0b1110:
		if (w>>4)&1 == 1 {
			return c.execVFPMove(w) // core↔VFP single-register transfer / VMRS / VMSR
		}
		return c.execVFPData(w)
	}
	return false
}

// execVFPLoadStore handles VLDR/VSTR (single register, immediate offset) and
// VLDM/VSTM/VPUSH/VPOP (register lists).
func (c *CPU) execVFPLoadStore(w uint32) {
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	d := (w >> 22) & 1
	wb := (w >> 21) & 1
	l := (w >> 20) & 1
	rn := (w >> 16) & 0xF
	vd := (w >> 12) & 0xF
	single := (w>>8)&0xF == 10
	imm8 := w & 0xFF
	base := c.reg(rn)

	if p == 1 && wb == 0 { // VLDR / VSTR
		offset := imm8 << 2
		addr := base + offset
		if u == 0 {
			addr = base - offset
		}
		if single {
			sd := sReg(vd, d)
			if l == 1 {
				c.sSetBits(sd, c.read32(addr))
			} else {
				c.write32(addr, c.sBits(sd))
			}
		} else {
			dd := dReg(vd, d)
			if l == 1 {
				c.VFP.S[2*dd] = c.read32(addr)
				c.VFP.S[2*dd+1] = c.read32(addr + 4)
			} else {
				c.write32(addr, c.VFP.S[2*dd])
				c.write32(addr+4, c.VFP.S[2*dd+1])
			}
		}
		return
	}

	// VLDM / VSTM / VPUSH / VPOP. imm8 counts single registers (double regs use
	// imm8/2). U selects increment-after (1) vs decrement-before (0).
	count := imm8
	regs := count
	if !single {
		regs = count / 2
	}
	addr := base
	if u == 0 { // decrement-before: start below the base
		addr = base - count*4
	}
	start := addr
	for i := uint32(0); i < regs; i++ {
		if single {
			sd := sReg(vd, d) + i
			if l == 1 {
				c.sSetBits(sd, c.read32(addr))
			} else {
				c.write32(addr, c.sBits(sd))
			}
			addr += 4
		} else {
			dd := dReg(vd, d) + i
			if l == 1 {
				c.VFP.S[2*dd] = c.read32(addr)
				c.VFP.S[2*dd+1] = c.read32(addr + 4)
			} else {
				c.write32(addr, c.VFP.S[2*dd])
				c.write32(addr+4, c.VFP.S[2*dd+1])
			}
			addr += 8
		}
	}
	if wb == 1 {
		if u == 1 {
			c.setReg(rn, base+count*4)
		} else {
			c.setReg(rn, start) // decrement-before leaves Rn at the lowered base
		}
	}
}

// execVFPData handles the three-register arithmetic and the extension group
// (VABS/VNEG/VSQRT/VMOV/VCMP/VCVT).
//
// The opcode is {bit23, bit21, bit20} — NOT bits[23:20]. Bit 22 is the D bit (the
// destination register's high bit), so folding it into the opcode would make an
// instruction with a high destination register fail to decode. Bit 6 is the
// operation's sub-select (add vs subtract, and so on). The extension group is
// {bit23,bit21,bit20} == 111.
func (c *CPU) execVFPData(w uint32) bool {
	pqr := (w>>23)&1<<2 | (w>>21)&1<<1 | (w>>20)&1 // {bit23,bit21,bit20}
	single := (w>>8)&0xF == 10
	op := (w >> 6) & 1

	dn := (w >> 16) & 0xF
	dd := (w >> 12) & 0xF
	dm := w & 0xF
	nBit := (w >> 7) & 1
	mBit := (w >> 5) & 1
	dBit := (w >> 22) & 1

	// Register numbers depend on precision.
	var vd, vn, vm uint32
	if single {
		vd, vn, vm = sReg(dd, dBit), sReg(dn, nBit), sReg(dm, mBit)
	} else {
		vd, vn, vm = dReg(dd, dBit), dReg(dn, nBit), dReg(dm, mBit)
	}

	// binop applies a float operation at the active precision.
	binop := func(f32 func(a, b float32) float32, f64 func(a, b float64) float64) {
		if single {
			c.sSet(vd, f32(c.sGet(vn), c.sGet(vm)))
		} else {
			c.dSet(vd, f64(c.dGet(vn), c.dGet(vm)))
		}
	}

	switch pqr {
	case 0b011: // VADD / VSUB
		if op == 0 {
			binop(func(a, b float32) float32 { return a + b }, func(a, b float64) float64 { return a + b })
		} else {
			binop(func(a, b float32) float32 { return a - b }, func(a, b float64) float64 { return a - b })
		}
		return true
	case 0b010: // VMUL / VNMUL
		binop(func(a, b float32) float32 { return a * b }, func(a, b float64) float64 { return a * b })
		if op == 1 { // VNMUL negates
			c.vfpNegate(vd, single)
		}
		return true
	case 0b100: // VDIV (op bit is 0 for VDIV)
		binop(func(a, b float32) float32 { return a / b }, func(a, b float64) float64 { return a / b })
		return true
	case 0b000: // VMLA / VMLS: Vd = Vd ± (Vn*Vm)
		c.vfpMulAcc(vd, vn, vm, single, op == 1, false)
		return true
	case 0b001: // VNMLA / VNMLS
		c.vfpMulAcc(vd, vn, vm, single, op == 0, true)
		return true
	case 0b111: // extension group (VABS/VNEG/VSQRT/VMOV/VCMP/VCVT)
		return c.execVFPExt(w, single, vd, vm)
	}
	c.Halt("unimplemented VFP data op %03b (0x%08X) at 0x%08X", pqr, w, c.cur)
	return true
}

func (c *CPU) vfpNegate(vd uint32, single bool) {
	if single {
		c.sSet(vd, -c.sGet(vd))
	} else {
		c.dSet(vd, -c.dGet(vd))
	}
}

// vfpMulAcc implements the multiply-accumulate family. sub negates the product;
// negAcc negates the accumulator first (the VNMLA/VNMLS forms).
func (c *CPU) vfpMulAcc(vd, vn, vm uint32, single, sub, negAcc bool) {
	if single {
		acc := c.sGet(vd)
		if negAcc {
			acc = -acc
		}
		p := c.sGet(vn) * c.sGet(vm)
		if sub {
			p = -p
		}
		c.sSet(vd, acc+p)
	} else {
		acc := c.dGet(vd)
		if negAcc {
			acc = -acc
		}
		p := c.dGet(vn) * c.dGet(vm)
		if sub {
			p = -p
		}
		c.dSet(vd, acc+p)
	}
}

// execVFPExt handles the opc1==1011 extension group: VMOV(imm/reg), VABS, VNEG,
// VSQRT, VCMP/VCMPE and VCVT.
func (c *CPU) execVFPExt(w uint32, single bool, vd, vm uint32) bool {
	opc2 := (w >> 16) & 0xF
	opc3 := (w >> 6) & 3 // {bit7, bit6}

	// VMOV immediate: opc3 low bit (bit6) == 0.
	if opc3&1 == 0 {
		imm := vfpExpandImm(w, single)
		if single {
			c.sSetBits(vd, uint32(imm))
		} else {
			c.VFP.S[2*vd] = uint32(imm)
			c.VFP.S[2*vd+1] = uint32(imm >> 32)
		}
		return true
	}

	switch opc2 {
	case 0b0000: // VMOV reg / VABS
		if opc3 == 0b01 { // VMOV (register copy)
			c.vfpCopy(vd, vm, single)
		} else { // VABS
			c.vfpUnary(vd, vm, single, math.Abs)
		}
		return true
	case 0b0001: // VNEG / VSQRT
		if opc3 == 0b01 { // VNEG
			c.vfpUnary(vd, vm, single, func(f float64) float64 { return -f })
		} else { // VSQRT
			c.vfpUnary(vd, vm, single, math.Sqrt)
		}
		return true
	case 0b0100, 0b0101: // VCMP / VCMPE (0b0101 compares against zero)
		c.execVCMP(w, single, vd, vm, opc2 == 0b0101)
		return true
	case 0b0111: // VCVT between single and double
		c.execVCVTPrec(w, single)
		return true
	case 0b1000: // VCVT integer -> float
		c.execVCVTFromInt(w, single)
		return true
	case 0b1100, 0b1101: // VCVT float -> integer
		c.execVCVTToInt(w, single)
		return true
	}
	c.Halt("unimplemented VFP ext op2=0x%X (0x%08X) at 0x%08X", opc2, w, c.cur)
	return true
}

func (c *CPU) vfpCopy(vd, vm uint32, single bool) {
	if single {
		c.sSetBits(vd, c.sBits(vm))
	} else {
		c.VFP.S[2*vd] = c.VFP.S[2*vm]
		c.VFP.S[2*vd+1] = c.VFP.S[2*vm+1]
	}
}

func (c *CPU) vfpUnary(vd, vm uint32, single bool, f func(float64) float64) {
	if single {
		c.sSet(vd, float32(f(float64(c.sGet(vm)))))
	} else {
		c.dSet(vd, f(c.dGet(vm)))
	}
}

// execVCMP compares two floats (or one against zero) and writes the VFP N/Z/C/V
// flags into FPSCR, following the IEEE ordered/unordered rules. A later VMRS
// APSR_nzcv copies these to the CPSR so a conditional branch can use them.
func (c *CPU) execVCMP(w uint32, single bool, vd, vm uint32, withZero bool) {
	var a, b float64
	if single {
		a = float64(c.sGet(vd))
		if !withZero {
			b = float64(c.sGet(vm))
		}
	} else {
		a = c.dGet(vd)
		if !withZero {
			b = c.dGet(vm)
		}
	}
	var n, z, cc, v bool
	switch {
	case math.IsNaN(a) || math.IsNaN(b): // unordered
		cc, v = true, true
	case a == b:
		z, cc = true, true
	case a < b:
		n = true
	default: // a > b
		cc = true
	}
	f := c.VFP.FPSCR &^ (0xF << 28)
	if n {
		f |= 1 << 31
	}
	if z {
		f |= 1 << 30
	}
	if cc {
		f |= 1 << 29
	}
	if v {
		f |= 1 << 28
	}
	c.VFP.FPSCR = f
}

// The three VCVT forms below each recompute their register numbers from the raw
// encoding, because the source and destination can be *different precisions*
// (and the integer side is always a single-precision slot regardless of the
// instruction's sz bit): reusing the caller's uniformly-typed vd/vm would number
// one operand wrong.
func vdOf(w uint32, single bool) uint32 {
	if single {
		return sReg((w>>12)&0xF, (w>>22)&1)
	}
	return dReg((w>>12)&0xF, (w>>22)&1)
}
func vmOf(w uint32, single bool) uint32 {
	if single {
		return sReg(w&0xF, (w>>5)&1)
	}
	return dReg(w&0xF, (w>>5)&1)
}

// execVCVTPrec converts between single and double precision (VCVT.F64.F32 and
// VCVT.F32.F64). The sz bit gives the source precision; the destination is the
// other, so the two registers are numbered at opposite precisions.
func (c *CPU) execVCVTPrec(w uint32, srcSingle bool) {
	if srcSingle { // F32 -> F64
		c.dSet(vdOf(w, false), float64(c.sGet(vmOf(w, true))))
	} else { // F64 -> F32
		c.sSet(vdOf(w, true), float32(c.dGet(vmOf(w, false))))
	}
}

// execVCVTFromInt converts a signed/unsigned integer (always in a single-
// precision source slot) to a float at the instruction's precision.
func (c *CPU) execVCVTFromInt(w uint32, single bool) {
	signed := (w>>7)&1 == 1
	bits := c.sBits(vmOf(w, true)) // integer source is single-precision numbered
	var f float64
	if signed {
		f = float64(int32(bits))
	} else {
		f = float64(bits)
	}
	if single {
		c.sSet(vdOf(w, true), float32(f))
	} else {
		c.dSet(vdOf(w, false), f)
	}
}

// execVCVTToInt converts a float at the instruction's precision to a signed or
// unsigned integer in a single-precision destination slot, truncating toward
// zero (the form C casts compile to).
func (c *CPU) execVCVTToInt(w uint32, single bool) {
	signed := (w>>16)&1 == 1
	var f float64
	if single {
		f = float64(c.sGet(vmOf(w, true)))
	} else {
		f = c.dGet(vmOf(w, false))
	}
	var out uint32
	if signed {
		out = uint32(int32(f)) // truncates toward zero
	} else {
		if f < 0 {
			f = 0
		}
		out = uint32(f)
	}
	c.sSetBits(vdOf(w, true), out) // integer destination is single-precision numbered
}

// execVFPMove handles the single-register core↔VFP transfers and VMRS/VMSR.
func (c *CPU) execVFPMove(w uint32) bool {
	l := (w >> 20) & 1
	rt := (w >> 12) & 0xF

	// VMSR/VMRS: coprocessor 10, register bits 19:16 select the system register.
	if (w>>21)&0x7 == 0b111 { // MRC/MCR with opc1==7 → system-register form
		sysreg := (w >> 16) & 0xF
		if l == 1 { // VMRS
			if sysreg == 0x1 { // FPSCR
				if rt == 15 { // VMRS APSR_nzcv, FPSCR → copy flags to CPSR
					c.N = c.VFP.FPSCR&(1<<31) != 0
					c.Z = c.VFP.FPSCR&(1<<30) != 0
					c.C = c.VFP.FPSCR&(1<<29) != 0
					c.V = c.VFP.FPSCR&(1<<28) != 0
				} else {
					c.setReg(rt, c.VFP.FPSCR)
				}
			} else if sysreg == 0x0 { // FPSID
				c.setReg(rt, 0x410120B4) // a plausible VFPv2 FPSID
			} else {
				c.setReg(rt, 0)
			}
		} else { // VMSR
			if sysreg == 0x1 {
				c.VFP.FPSCR = c.reg(rt)
			}
		}
		return true
	}

	// VMOV core register ↔ single-precision register.
	vn := sReg((w>>16)&0xF, (w>>7)&1)
	if l == 1 { // VMOV Rt, Sn
		c.setReg(rt, c.sBits(vn))
	} else { // VMOV Sn, Rt
		c.sSetBits(vn, c.reg(rt))
	}
	return true
}

// execVFPMove64 handles the two-core-register ↔ doubleword transfers (VMOV Dm,
// Rt, Rt2 and the reverse), used to load a double or a register pair at once.
func (c *CPU) execVFPMove64(w uint32) bool {
	l := (w >> 20) & 1
	rt := (w >> 12) & 0xF
	rt2 := (w >> 16) & 0xF
	single := (w>>8)&0xF == 10
	mBit := (w >> 5) & 1
	vm := w & 0xF

	if single {
		// VMOV Sm, Sm1, Rt, Rt2 — two consecutive single registers.
		sm := sReg(vm, mBit)
		if l == 1 {
			c.setReg(rt, c.sBits(sm))
			c.setReg(rt2, c.sBits(sm+1))
		} else {
			c.sSetBits(sm, c.reg(rt))
			c.sSetBits(sm+1, c.reg(rt2))
		}
	} else {
		dm := dReg(vm, mBit)
		if l == 1 { // to core registers
			c.setReg(rt, c.VFP.S[2*dm])
			c.setReg(rt2, c.VFP.S[2*dm+1])
		} else { // to the double register
			c.VFP.S[2*dm] = c.reg(rt)
			c.VFP.S[2*dm+1] = c.reg(rt2)
		}
	}
	return true
}

// vfpExpandImm reconstructs a VFP immediate constant. The 8-bit immediate is
// split across bits 19:16 (high nibble) and 3:0 (low nibble), then expanded to a
// float per the ARM ARM's VFPExpandImm: for single precision
//
//	imm32 = a : NOT(b) : b×5 : c : d : (fraction efgh) : 19 zeros
//
// where the imm8 bits are a b c d e f g h (a = bit 7). Double precision is the
// same shape with an 11-bit exponent (b×8) and a 52-bit fraction. Returns the raw
// bit pattern (in the low 32 bits for single, all 64 for double).
func vfpExpandImm(w uint32, single bool) uint64 {
	imm8 := (((w >> 16) & 0xF) << 4) | (w & 0xF)
	sign := (imm8 >> 7) & 1
	b := (imm8 >> 6) & 1
	if single {
		exp := ((^b & 1) << 7) | (repeat(b, 5) << 2) | ((imm8 >> 4) & 3)
		frac := imm8 & 0xF
		return uint64(sign<<31 | exp<<23 | frac<<19)
	}
	exp := ((uint64(^b&1) << 10) | (repeat64(uint64(b), 8) << 2) | uint64((imm8>>4)&3))
	frac := uint64(imm8 & 0xF)
	return uint64(sign)<<63 | exp<<52 | frac<<48
}

// repeat returns n copies of bit's low bit as a small bit field.
func repeat(bit, n uint32) uint32 {
	var v uint32
	for i := uint32(0); i < n; i++ {
		v = v<<1 | (bit & 1)
	}
	return v
}
func repeat64(bit uint64, n uint32) uint64 {
	var v uint64
	for i := uint32(0); i < n; i++ {
		v = v<<1 | (bit & 1)
	}
	return v
}
