package allegrex

// vfpu.go is the Allegrex COP2 vector unit (the VFPU): 128 single-precision
// registers organized as 8 4x4 matrices, addressed by 7-bit register numbers that
// select single/pair/triple/quad vectors (columns, or rows via a transpose bit) and
// 2x2/3x3/4x4 matrices, plus 16 control registers (the vpfxs/vpfxt/vpfxd operand
// prefixes, the vcmp condition codes, and the vrnd state).
//
// The VFPU occupies the COP2 opcode (op 0x12: register moves mfv/mtv/mfvc/mtvc and
// the bvf/bvt branches) plus a family of primary opcodes: loads/stores (0x32/0x36/
// 0x35 lv, 0x3A/0x3E/0x3D sv), the arithmetic groups VFPU0/1/3 (0x18/0x19/0x1B),
// the single-operand group and conversions (0x34), the prefix/immediate group
// (0x37) and the matrix group (0x3C).
//
// The instruction encodings, register addressing, prefix semantics and operation
// behaviour follow the PSP platform specification as documented by the PPSSPP
// project's interpreter (Core/MIPS/MIPSIntVFPU.cpp, MIPSVFPUUtils.cpp) — used as a
// hardware reference the way the KIRK and GE constants are; the implementation here
// is written from that behaviour, not translated code. Two documented
// simplifications, both invisible to code that leaves the prefixes at their
// defaults (which is what compiled SDK code does):
//
//   - the operand prefixes are applied with their standard semantics to every
//     element of vector operations; the hardware's per-op quirks (prefixes acting
//     only on the last lane of vdiv/vrcp-family ops, forced-constant rewriting
//     inside vhdp/vcrs/vscl) are folded into the ops' direct formulas instead;
//   - matrix operations (vmmul/vtfm/vmscl/vmmov and the matrix set group) ignore
//     a pending prefix rather than modelling its last-row-only application.
//
// Unimplemented tail ops Halt with their word and PC (or report to OnVFPU), so
// gaps are explicit and grow lazily from real stops.

import (
	"fmt"
	"math"
)

// VFPU control registers (mtvc/mfvc register space, VfpuCtrl indices).
const (
	vfpuCtlSPfx = 0 // vpfxs latch
	vfpuCtlTPfx = 1 // vpfxt latch
	vfpuCtlDPfx = 2 // vpfxd latch
	vfpuCtlCC   = 3 // vcmp condition codes (6 bits)

	// pfxIdentity is the reset/consumed value of the S and T prefixes: the
	// swizzle 0b11100100 = (x,y,z,w) with no abs/constant/negate bits.
	pfxIdentity = 0xE4
)

// vfpuCtrlMask is the writable-bit mask of each control register (writes through
// mtvc/vmtvc are masked; REV and the reserved registers are read-only).
var vfpuCtrlMask = [16]uint32{
	0x000FFFFF, 0x000FFFFF, 0x00000FFF, 0x0000003F, // SPFX TPFX DPFX CC
	0xFFFFFFFF, 0, 0, 0, // INF4 RSV5 RSV6 REV
	0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF, // RCX0..3
	0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF, // RCX4..7
}

// --- register addressing -----------------------------------------------------
//
// The flat register file V[128] is column-major within each matrix:
// V[mtx*16 + col*4 + row]. A 7-bit vector register number decomposes into
// mtx = bits 2..4, col = bits 0..1, and a start row + transpose whose encoding
// depends on the operand size; the transpose bit switches a column vector to a
// row vector (and M to E for matrices).

// vfpuSingle returns the flat V index of a 7-bit single register number.
func vfpuSingle(reg uint32) uint32 {
	m := (reg >> 2) & 7
	c := reg & 3
	r := (reg >> 5) & 3
	return m*16 + c*4 + r
}

// vecIdx returns the flat V indices of an n-element vector register.
func vecIdx(reg, n uint32) (idx [4]uint32) {
	if n == 1 {
		idx[0] = vfpuSingle(reg)
		return
	}
	var row uint32
	switch n {
	case 2, 4:
		row = (reg >> 5) & 2
	case 3:
		row = (reg >> 6) & 1
	}
	mtx := (reg >> 2) & 7
	col := reg & 3
	if reg&0x20 != 0 { // transposed: a row vector, stride 4
		base := mtx*16 + col
		for i := uint32(0); i < n; i++ {
			idx[i] = base + ((row+i)&3)*4
		}
	} else { // a column vector, stride 1
		base := mtx*16 + col*4
		for i := uint32(0); i < n; i++ {
			idx[i] = base + ((row + i) & 3)
		}
	}
	return
}

// matIdx returns the flat V index of element (a,b) of a side×side matrix register:
// a indexes the columns of the register matrix (rows when transposed), b the rows.
func matIdx(reg, side, a, b uint32) uint32 {
	var row uint32
	switch side {
	case 2, 4:
		row = (reg >> 5) & 2
	case 3:
		row = (reg >> 6) & 1
	}
	mtx := (reg >> 2) & 7
	col := reg & 3
	if reg&0x20 != 0 {
		return mtx*16 + ((row+b)&3)*4 + ((col + a) & 3)
	}
	return mtx*16 + ((col+a)&3)*4 + ((row + b) & 3)
}

func (c *CPU) vread(reg, n uint32) (v [4]float32) {
	idx := vecIdx(reg, n)
	for i := uint32(0); i < n; i++ {
		v[i] = math.Float32frombits(c.V[idx[i]&127])
	}
	return
}

func (c *CPU) vreadBits(reg, n uint32) (v [4]uint32) {
	idx := vecIdx(reg, n)
	for i := uint32(0); i < n; i++ {
		v[i] = c.V[idx[i]&127]
	}
	return
}

// vwrite writes an n-element vector, honoring the vpfxd write mask (a set mask
// bit keeps the destination element unchanged).
func (c *CPU) vwrite(v [4]float32, reg, n uint32) {
	var bits [4]uint32
	for i := uint32(0); i < n; i++ {
		bits[i] = math.Float32bits(v[i])
	}
	c.vwriteBits(bits, reg, n)
}

func (c *CPU) vwriteBits(v [4]uint32, reg, n uint32) {
	mask := c.VfpuCtrl[vfpuCtlDPfx] >> 8
	idx := vecIdx(reg, n)
	for i := uint32(0); i < n; i++ {
		if mask&(1<<i) == 0 {
			c.V[idx[i]&127] = v[i]
		}
	}
}

// mread reads a side×side matrix: m[a*4+b] = element (col a, row b) of the
// register matrix (transposed when the register's E bit is set), so m[a*4..a*4+3]
// is one column as the arithmetic below consumes it. Unfilled elements stay zero.
func (c *CPU) mread(reg, side uint32) (m [16]float32) {
	for a := uint32(0); a < side; a++ {
		for b := uint32(0); b < side; b++ {
			m[a*4+b] = math.Float32frombits(c.V[matIdx(reg, side, a, b)&127])
		}
	}
	return
}

// mwrite writes a side×side matrix (the vpfxd mask does not apply to matrix ops).
func (c *CPU) mwrite(m [16]float32, reg, side uint32) {
	for a := uint32(0); a < side; a++ {
		for b := uint32(0); b < side; b++ {
			c.V[matIdx(reg, side, a, b)&127] = math.Float32bits(m[a*4+b])
		}
	}
}

// --- operand prefixes ----------------------------------------------------------
//
// vpfxs/vpfxt latch a 20-bit source prefix that transforms the next op's operand:
// per element, bits 0..7 pick a swizzle source (2 bits each), bits 8..11 force
// absolute value, bits 12..15 substitute a constant (the swizzle+abs bits then
// select which), bits 16..19 negate. vpfxd latches a 12-bit destination prefix:
// bits 0..7 saturate per element (1 = clamp [0,1], 3 = clamp [-1,1]), bits 8..11
// mask the write. A prefix applies to one op and resets.

// pfxConstants are the constants a source prefix can substitute, indexed by
// swizzle | abs<<2.
var pfxConstants = [8]float32{0, 1, 2, 0.5, 3, 1.0 / 3.0, 0.25, 1.0 / 6.0}

// applyPfx transforms v[0:n] per a source-prefix word. abs and negate operate on
// the sign bit so NaN payloads pass through unchanged.
func applyPfx(v *[4]float32, data, n uint32) {
	if data == pfxIdentity {
		return
	}
	var orig [4]float32
	copy(orig[:], v[:n])
	for i := uint32(0); i < n; i++ {
		swz := (data >> (i * 2)) & 3
		abs := (data >> (8 + i)) & 1
		neg := (data >> (16 + i)) & 1
		cst := (data >> (12 + i)) & 1
		var r float32
		if cst != 0 {
			r = pfxConstants[swz|abs<<2]
		} else {
			r = orig[swz&3]
			if abs != 0 {
				r = math.Float32frombits(math.Float32bits(r) &^ 0x80000000)
			}
		}
		if neg != 0 {
			r = math.Float32frombits(math.Float32bits(r) ^ 0x80000000)
		}
		v[i] = r
	}
}

func (c *CPU) applyPfxS(v *[4]float32, n uint32) { applyPfx(v, c.VfpuCtrl[vfpuCtlSPfx], n) }
func (c *CPU) applyPfxT(v *[4]float32, n uint32) { applyPfx(v, c.VfpuCtrl[vfpuCtlTPfx], n) }

// applyPfxD applies the destination prefix's saturation (the write mask is
// honored by vwrite).
func (c *CPU) applyPfxD(v *[4]float32, n uint32) {
	data := c.VfpuCtrl[vfpuCtlDPfx]
	if data == 0 {
		return
	}
	for i := uint32(0); i < n; i++ {
		switch (data >> (i * 2)) & 3 {
		case 1:
			v[i] = vfpuClamp(v[i], 0, 1)
		case 3:
			v[i] = vfpuClamp(v[i], -1, 1)
		}
	}
}

// vfpuClamp saturates preserving NaN (comparisons with NaN are false).
func vfpuClamp(v, lo, hi float32) float32 {
	if v >= hi {
		return hi
	}
	if v <= lo {
		return lo
	}
	return v
}

// eatPfx consumes the prefixes after an op.
func (c *CPU) eatPfx() {
	c.VfpuCtrl[vfpuCtlSPfx] = pfxIdentity
	c.VfpuCtrl[vfpuCtlTPfx] = pfxIdentity
	c.VfpuCtrl[vfpuCtlDPfx] = 0
}

// --- helpers ---------------------------------------------------------------

// vfpuVecN decodes an op's vector size (1..4 = .s/.p/.t/.q) from bits 7 and 15.
func vfpuVecN(w uint32) uint32 { return (w>>7)&1 + (w>>14)&2 + 1 }

var vfpuSuffix = [5]string{"", "s", "p", "t", "q"}

// vfpuSin/vfpuCos: the VFPU measures angles in quarter turns (x=1 is 90°), which
// makes quadrant values exact; integral inputs are special-cased so rotation
// matrices built from vrot/vsin/vcos are exact, as on hardware.
func vfpuSin(x float32) float32 {
	k := math.Mod(float64(x), 4)
	if k < 0 {
		k += 4
	}
	switch k {
	case 0, 2:
		return 0
	case 1:
		return 1
	case 3:
		return -1
	}
	return float32(math.Sin(k * math.Pi / 2))
}

func vfpuCos(x float32) float32 {
	k := math.Mod(float64(x), 4)
	if k < 0 {
		k += 4
	}
	switch k {
	case 0:
		return 1
	case 2:
		return -1
	case 1, 3:
		return 0
	}
	return float32(math.Cos(k * math.Pi / 2))
}

// float16to32 expands the VFPU's half-float immediates (vfim.s).
func float16to32(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h>>10) & 0x1F
	frac := uint32(h) & 0x3FF
	switch exp {
	case 0: // zero / denormal
		if frac == 0 {
			return math.Float32frombits(sign)
		}
		return math.Float32frombits(sign) + float32(frac)*float32(math.Pow(2, -24))*sign2(sign)
	case 0x1F: // inf / NaN
		return math.Float32frombits(sign | 0x7F800000 | frac<<13)
	}
	return math.Float32frombits(sign | (exp+112)<<23 | frac<<13)
}

func sign2(signBit uint32) float32 {
	if signBit != 0 {
		return -1
	}
	return 1
}

// vcstConstants is the vcst table (indices beyond 19 read as 0 on hardware).
var vcstConstants = [32]float32{
	0,
	math.MaxFloat32,
	math.Sqrt2,
	float32(math.Sqrt(0.5)),
	float32(2 / math.Sqrt(math.Pi)),
	float32(2 / math.Pi),
	float32(1 / math.Pi),
	math.Pi / 4,
	math.Pi / 2,
	math.Pi,
	math.E,
	math.Log2E,
	math.Log10E,
	math.Ln2,
	float32(math.Log(10)),
	2 * math.Pi,
	math.Pi / 6,
	float32(math.Log10(2)),
	float32(math.Log(10) / math.Log(2)),
	float32(math.Sqrt(3) / 2),
}

// --- disassembly -------------------------------------------------------------

// vnot names an n-element vector register (S=single, C=column, R=row).
func vnot(reg, n uint32) string {
	mtx := (reg >> 2) & 7
	col := reg & 3
	var row uint32
	ch := byte('C')
	switch n {
	case 1:
		return fmt.Sprintf("S%d%d%d", mtx, col, (reg>>5)&3)
	case 2, 4:
		row = (reg >> 5) & 2
	case 3:
		row = (reg >> 6) & 1
	}
	if reg&0x20 != 0 {
		ch = 'R'
		return fmt.Sprintf("%c%d%d%d", ch, mtx, row, col)
	}
	return fmt.Sprintf("%c%d%d%d", ch, mtx, col, row)
}

// mnot names a matrix register (M, or E when transposed).
func mnot(reg, side uint32) string {
	mtx := (reg >> 2) & 7
	col := reg & 3
	var row uint32
	switch side {
	case 2, 4:
		row = (reg >> 5) & 2
	case 3:
		row = (reg >> 6) & 1
	}
	if reg&0x20 != 0 {
		return fmt.Sprintf("E%d%d%d", mtx, row, col)
	}
	return fmt.Sprintf("M%d%d%d", mtx, col, row)
}

// decodeCop2 disassembles a COP2 (VFPU) register move or branch.
func decodeCop2(in Inst, w, rs, rt, rd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	imm := w & 0xFF
	switch rs {
	case 0x03: // mfv / mfvc
		if imm >= 128 {
			return set("mfvc", fmt.Sprintf("mfvc %s, $%d", reg(rt), imm-128))
		}
		return set("mfv", fmt.Sprintf("mfv %s, %s", reg(rt), vnot(imm, 1)))
	case 0x07: // mtv / mtvc
		if imm >= 128 {
			return set("mtvc", fmt.Sprintf("mtvc %s, $%d", reg(rt), imm-128))
		}
		return set("mtv", fmt.Sprintf("mtv %s, %s", reg(rt), vnot(imm, 1)))
	case 0x08: // bvf/bvt/bvfl/bvtl on CC bit (w>>18)&7
		in.Flow, in.HasTarget, in.HasDelay = FlowBranch, true, true
		in.Target = in.Addr + 4 + uint32(int32(int16(w)))*4
		m := [...]string{"bvf", "bvt", "bvfl", "bvtl"}[(w>>16)&3]
		return set(m, fmt.Sprintf("%s %d, $%08X", m, (w>>18)&7, in.Target))
	}
	return set("cop2", fmt.Sprintf("cop2 0x%08X", w))
}

// decodeVFPU disassembles the VFPU load/store and compute primary opcodes; ops
// beyond the executed set keep a generic-but-named rendering so a trace stays
// legible without stopping on .word.
func decodeVFPU(in Inst, w, op, rs, rt, simm uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	n := vfpuVecN(w)
	sfx := vfpuSuffix[n]
	vd := w & 0x7F
	vs := (w >> 8) & 0x7F
	vt := (w >> 16) & 0x7F
	off := int32(simm) &^ 3

	vv := func(name string) Inst {
		return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s", name, sfx, vnot(vd, n), vnot(vs, n)))
	}
	vvv := func(name string) Inst {
		return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s, %s", name, sfx, vnot(vd, n), vnot(vs, n), vnot(vt, n)))
	}
	v1 := func(name string) Inst {
		return set(name+"."+sfx, fmt.Sprintf("%s.%s %s", name, sfx, vnot(vd, n)))
	}

	switch op {
	case 0x32:
		vt := rt | (w&3)<<5
		return set("lv.s", fmt.Sprintf("lv.s %s, %d(%s)", vnot(vt, 1), off, reg(rs)))
	case 0x3A:
		vt := rt | (w&3)<<5
		return set("sv.s", fmt.Sprintf("sv.s %s, %d(%s)", vnot(vt, 1), off, reg(rs)))
	case 0x36:
		vt := rt | (w&1)<<5
		return set("lv.q", fmt.Sprintf("lv.q %s, %d(%s)", vnot(vt, 4), off, reg(rs)))
	case 0x3E:
		vt := rt | (w&1)<<5
		return set("sv.q", fmt.Sprintf("sv.q %s, %d(%s)", vnot(vt, 4), off, reg(rs)))
	case 0x35:
		vt := rt | (w&1)<<5
		return set("lv"+lr(w)+".q", fmt.Sprintf("lv%s.q %s, %d(%s)", lr(w), vnot(vt, 4), off, reg(rs)))
	case 0x3D:
		vt := rt | (w&1)<<5
		return set("sv"+lr(w)+".q", fmt.Sprintf("sv%s.q %s, %d(%s)", lr(w), vnot(vt, 4), off, reg(rs)))

	case 0x18: // VFPU0
		switch (w >> 23) & 7 {
		case 0:
			return vvv("vadd")
		case 1:
			return vvv("vsub")
		case 2:
			return vvv("vsbn")
		case 7:
			return vvv("vdiv")
		}
	case 0x19: // VFPU1
		switch (w >> 23) & 7 {
		case 0:
			return vvv("vmul")
		case 1:
			return set("vdot."+sfx, fmt.Sprintf("vdot.%s %s, %s, %s", sfx, vnot(vd, 1), vnot(vs, n), vnot(vt, n)))
		case 2:
			return set("vscl."+sfx, fmt.Sprintf("vscl.%s %s, %s, %s", sfx, vnot(vd, n), vnot(vs, n), vnot(vt, 1)))
		case 4:
			return set("vhdp."+sfx, fmt.Sprintf("vhdp.%s %s, %s, %s", sfx, vnot(vd, 1), vnot(vs, n), vnot(vt, n)))
		case 5:
			return vvv("vcrs")
		case 6:
			return set("vdet."+sfx, fmt.Sprintf("vdet.%s %s, %s, %s", sfx, vnot(vd, 1), vnot(vs, n), vnot(vt, n)))
		}
	case 0x1B: // VFPU3
		switch (w >> 23) & 7 {
		case 0:
			return set("vcmp."+sfx, fmt.Sprintf("vcmp.%s %d, %s, %s", sfx, w&0xF, vnot(vs, n), vnot(vt, n)))
		case 2:
			return vvv("vmin")
		case 3:
			return vvv("vmax")
		case 5:
			return vvv("vscmp")
		case 6:
			return vvv("vsge")
		case 7:
			return vvv("vslt")
		}
	case 0x34: // VFPU4 groups
		switch rs {
		case 0x00:
			names := map[uint32]string{
				0: "vmov", 1: "vabs", 2: "vneg", 3: "vidt", 4: "vsat0", 5: "vsat1",
				6: "vzero", 7: "vone", 16: "vrcp", 17: "vrsq", 18: "vsin", 19: "vcos",
				20: "vexp2", 21: "vlog2", 22: "vsqrt", 23: "vasin", 24: "vnrcp",
				26: "vnsin", 28: "vrexp2",
			}
			if name, ok := names[rt]; ok {
				switch rt {
				case 3, 6, 7:
					return v1(name)
				}
				return vv(name)
			}
		case 0x01:
			names := map[uint32]string{
				0: "vrnds", 1: "vrndi", 2: "vrndf1", 3: "vrndf2", 18: "vf2h", 19: "vh2f",
				22: "vsbz", 23: "vlgb", 24: "vuc2ifs", 25: "vc2i", 26: "vus2i", 27: "vs2i",
				28: "vi2uc", 29: "vi2c", 30: "vi2us", 31: "vi2s",
			}
			if name, ok := names[rt]; ok {
				return vv(name)
			}
		case 0x02:
			names := map[uint32]string{
				0: "vsrt1", 1: "vsrt2", 2: "vbfy1", 3: "vbfy2", 4: "vocp", 5: "vsocp",
				6: "vfad", 7: "vavg", 8: "vsrt3", 9: "vsrt4", 10: "vsgn",
				16: "vmfvc", 17: "vmtvc", 25: "vt4444", 26: "vt5551", 27: "vt5650",
			}
			if name, ok := names[rt]; ok {
				return vv(name)
			}
		case 0x03:
			return set("vcst."+sfx, fmt.Sprintf("vcst.%s %s, %d", sfx, vnot(vd, n), rt))
		case 0x10, 0x11, 0x12, 0x13:
			name := [...]string{"vf2in", "vf2iz", "vf2iu", "vf2id"}[rs-0x10]
			return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s, %d", name, sfx, vnot(vd, n), vnot(vs, n), rt))
		case 0x14:
			return set("vi2f."+sfx, fmt.Sprintf("vi2f.%s %s, %s, %d", sfx, vnot(vd, n), vnot(vs, n), rt))
		case 0x15:
			name := "vcmovt"
			if (w>>19)&1 != 0 {
				name = "vcmovf"
			}
			return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s, %d", name, sfx, vnot(vd, n), vnot(vs, n), (w>>16)&7))
		}
		if rs >= 0x18 {
			return vv("vwbn")
		}
	case 0x37: // VFPU5: prefixes and immediates
		switch (w >> 24) & 3 {
		case 0:
			return set("vpfxs", fmt.Sprintf("vpfxs 0x%05X", w&0xFFFFF))
		case 1:
			return set("vpfxt", fmt.Sprintf("vpfxt 0x%05X", w&0xFFFFF))
		case 2:
			return set("vpfxd", fmt.Sprintf("vpfxd 0x%03X", w&0xFFF))
		case 3:
			vt7 := (w >> 16) & 0x7F
			if w&(1<<23) != 0 {
				return set("vfim.s", fmt.Sprintf("vfim.s %s, 0x%04X", vnot(vt7, 1), w&0xFFFF))
			}
			return set("viim.s", fmt.Sprintf("viim.s %s, %d", vnot(vt7, 1), int32(int16(w))))
		}
	case 0x3C: // VFPU6: matrix group
		switch (w >> 21) & 31 {
		case 0, 1, 2, 3:
			return set("vmmul."+sfx, fmt.Sprintf("vmmul.%s %s, %s, %s", sfx, mnot(vd, n), mnot(vs, n), mnot(vt, n)))
		case 16, 17, 18, 19:
			return set("vmscl."+sfx, fmt.Sprintf("vmscl.%s %s, %s, %s", sfx, mnot(vd, n), mnot(vs, n), vnot(vt, 1)))
		case 20, 21, 22, 23:
			name := "vcrsp"
			if n == 4 {
				name = "vqmul"
			}
			return vvv(name)
		case 28:
			switch rt & 0xF {
			case 0:
				return set("vmmov."+sfx, fmt.Sprintf("vmmov.%s %s, %s", sfx, mnot(vd, n), mnot(vs, n)))
			case 3:
				return set("vmidt."+sfx, fmt.Sprintf("vmidt.%s %s", sfx, mnot(vd, n)))
			case 6:
				return set("vmzero."+sfx, fmt.Sprintf("vmzero.%s %s", sfx, mnot(vd, n)))
			case 7:
				return set("vmone."+sfx, fmt.Sprintf("vmone.%s %s", sfx, mnot(vd, n)))
			}
		case 29:
			return set("vrot."+sfx, fmt.Sprintf("vrot.%s %s, %s, %d", sfx, vnot(vd, n), vnot(vs, 1), (w>>16)&0x1F))
		default:
			ins := (w >> 23) & 3
			if ins >= 1 {
				name := fmt.Sprintf("vtfm%d", ins+1)
				if n != ins+1 {
					name = fmt.Sprintf("vhtfm%d", ins+1)
				}
				return set(name, fmt.Sprintf("%s.%s %s, %s, %s", name, sfx, vnot(vd, ins+1), mnot(vs, ins+1), vnot(vt, n)))
			}
		}
	}
	return set("vfpu", fmt.Sprintf("vfpu.%02X 0x%08X", op, w))
}

func lr(w uint32) string {
	if w&2 != 0 {
		return "r"
	}
	return "l"
}

// --- execution ---------------------------------------------------------------

// cop2 executes a COP2 (VFPU) register move or branch.
func (c *CPU) cop2(w, rs, rt, rd uint32) {
	imm := w & 0xFF
	switch rs {
	case 0x03: // mfv / mfvc
		if imm >= 128 {
			var v uint32
			if imm < 128+16 {
				v = c.VfpuCtrl[imm-128]
			}
			c.load(rt, v)
			return
		}
		c.load(rt, c.V[vfpuSingle(imm)&127])
	case 0x07: // mtv / mtvc
		if imm >= 128 {
			if imm < 128+16 {
				i := imm - 128
				if vfpuCtrlMask[i] != 0 {
					c.VfpuCtrl[i] = c.reg(rt) & vfpuCtrlMask[i]
				}
			}
			return
		}
		c.V[vfpuSingle(imm)&127] = c.reg(rt)
	case 0x08: // bvf / bvt / bvfl / bvtl
		target := c.curPC + 4 + uint32(int32(int16(w)))<<2
		val := (c.VfpuCtrl[vfpuCtlCC]>>((w>>18)&7))&1 != 0
		switch (w >> 16) & 3 {
		case 0:
			c.doBranch(!val, target)
		case 1:
			c.doBranch(val, target)
		case 2:
			if !val {
				c.doBranch(true, target)
			} else {
				c.nullifyNext = true
			}
		case 3:
			if val {
				c.doBranch(true, target)
			} else {
				c.nullifyNext = true
			}
		}
	default:
		c.Halt("unimplemented cop2 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, c.curPC)
	}
}

// vfpuOp executes the VFPU load/store and compute primary opcodes.
func (c *CPU) vfpuOp(w, op, rs, rt, simm uint32) {
	switch op {
	case 0x32: // lv.s
		vt := rt | (w&3)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.V[vfpuSingle(vt)&127] = c.read32(addr)
	case 0x3A: // sv.s
		vt := rt | (w&3)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write32(addr, c.V[vfpuSingle(vt)&127])
	case 0x36: // lv.q
		vt := rt | (w&1)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&15 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		var v [4]uint32
		for i := uint32(0); i < 4; i++ {
			v[i] = c.read32(addr + i*4)
		}
		idx := vecIdx(vt, 4)
		for i := uint32(0); i < 4; i++ {
			c.V[idx[i]&127] = v[i]
		}
	case 0x3E: // sv.q
		vt := rt | (w&1)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&15 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		idx := vecIdx(vt, 4)
		for i := uint32(0); i < 4; i++ {
			c.write32(addr+i*4, c.V[idx[i]&127])
		}
	case 0x35: // lvl.q / lvr.q: fill the quad from an unaligned word address
		vt := rt | (w&1)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		idx := vecIdx(vt, 4)
		offset := (addr >> 2) & 3
		if w&2 == 0 { // lvl.q
			for i := uint32(0); i <= offset; i++ {
				c.V[idx[3-i]&127] = c.read32(addr - 4*i)
			}
		} else { // lvr.q
			for i := uint32(0); i <= 3-offset; i++ {
				c.V[idx[i]&127] = c.read32(addr + 4*i)
			}
		}
	case 0x3D: // svl.q / svr.q
		vt := rt | (w&1)<<5
		addr := c.reg(rs) + (simm &^ 3)
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		idx := vecIdx(vt, 4)
		offset := (addr >> 2) & 3
		if w&2 == 0 { // svl.q
			for i := uint32(0); i <= offset; i++ {
				c.write32(addr-4*i, c.V[idx[3-i]&127])
			}
		} else { // svr.q
			for i := uint32(0); i <= 3-offset; i++ {
				c.write32(addr+4*i, c.V[idx[i]&127])
			}
		}

	case 0x18, 0x19, 0x1B:
		c.vfpuALU(w, op)
	case 0x34:
		c.vfpu4(w, rs, rt)
	case 0x37:
		c.vfpu5(w)
	case 0x3C:
		c.vfpuMatrix(w, rt)
	default:
		c.vfpuUnimpl(w, op)
	}
}

// vfpuUnimpl reports an unimplemented VFPU op to OnVFPU or halts.
func (c *CPU) vfpuUnimpl(w, op uint32) {
	if c.OnVFPU != nil {
		c.OnVFPU(w, op)
		return
	}
	c.Halt("unimplemented VFPU op 0x%02X %q (word 0x%08X) at 0x%08X",
		op, decodeVFPU(Inst{}, w, op, (w>>21)&31, (w>>16)&31, uint32(int32(int16(w)))).Mnem, w, c.curPC)
}

// vfpuALU executes the element-wise and reduction groups VFPU0/1/3.
func (c *CPU) vfpuALU(w, op uint32) {
	n := vfpuVecN(w)
	vd := w & 0x7F
	vs := (w >> 8) & 0x7F
	vt := (w >> 16) & 0x7F
	sub := (w >> 23) & 7
	var d [4]float32

	switch op {
	case 0x18: // VFPU0: vadd/vsub/vdiv
		s := c.vread(vs, n)
		t := c.vread(vt, n)
		c.applyPfxS(&s, n)
		c.applyPfxT(&t, n)
		for i := uint32(0); i < n; i++ {
			switch sub {
			case 0:
				d[i] = s[i] + t[i]
			case 1:
				d[i] = s[i] - t[i]
			case 7:
				d[i] = s[i] / t[i]
			default:
				c.vfpuUnimpl(w, op)
				return
			}
		}
	case 0x19: // VFPU1
		switch sub {
		case 0: // vmul
			s := c.vread(vs, n)
			t := c.vread(vt, n)
			c.applyPfxS(&s, n)
			c.applyPfxT(&t, n)
			for i := uint32(0); i < n; i++ {
				d[i] = s[i] * t[i]
			}
		case 1: // vdot -> single
			s := c.vread(vs, n)
			t := c.vread(vt, n)
			c.applyPfxS(&s, n)
			c.applyPfxT(&t, n)
			var sum float32
			for i := uint32(0); i < n; i++ {
				sum += s[i] * t[i]
			}
			d[0] = sum
			c.applyPfxD(&d, 1)
			c.vwrite(d, vd, 1)
			c.eatPfx()
			return
		case 2: // vscl: scale by the single T
			s := c.vread(vs, n)
			t := c.vread(vt, 1)
			c.applyPfxS(&s, n)
			for i := uint32(0); i < n; i++ {
				d[i] = s[i] * t[0]
			}
		case 4: // vhdp: homogeneous dot (S's last lane wired to 1) -> single
			s := c.vread(vs, n)
			t := c.vread(vt, n)
			c.applyPfxT(&t, n)
			var sum float32
			for i := uint32(0); i < n-1; i++ {
				sum += s[i] * t[i]
			}
			sum += t[n-1]
			d[0] = sum
			c.applyPfxD(&d, 1)
			c.vwrite(d, vd, 1)
			c.eatPfx()
			return
		case 5: // vcrs: half cross product d = s.yzx * t.zxy
			s := c.vread(vs, n)
			t := c.vread(vt, n)
			d[0] = s[1] * t[2]
			d[1] = s[2] * t[0]
			d[2] = s[0] * t[1]
		case 6: // vdet: 2x2 determinant -> single
			s := c.vread(vs, n)
			t := c.vread(vt, n)
			d[0] = s[0]*t[1] - s[1]*t[0]
			c.applyPfxD(&d, 1)
			c.vwrite(d, vd, 1)
			c.eatPfx()
			return
		default:
			c.vfpuUnimpl(w, op)
			return
		}
	case 0x1B: // VFPU3
		s := c.vread(vs, n)
		t := c.vread(vt, n)
		c.applyPfxS(&s, n)
		c.applyPfxT(&t, n)
		switch sub {
		case 0: // vcmp: set CC bits + any/all
			cond := w & 0xF
			var cc, orv uint32
			andv := uint32(1)
			affected := uint32(1<<4 | 1<<5)
			for i := uint32(0); i < n; i++ {
				var ci bool
				switch cond {
				case 0x0: // FL
					ci = false
				case 0x1: // EQ
					ci = s[i] == t[i]
				case 0x2: // LT
					ci = s[i] < t[i]
				case 0x3: // LE
					ci = s[i] <= t[i]
				case 0x4: // TR
					ci = true
				case 0x5: // NE
					ci = s[i] != t[i]
				case 0x6: // GE
					ci = s[i] >= t[i]
				case 0x7: // GT
					ci = s[i] > t[i]
				case 0x8: // EZ
					ci = s[i] == 0
				case 0x9: // EN (NaN)
					ci = s[i] != s[i]
				case 0xA: // EI (infinite)
					ci = isInf32(s[i])
				case 0xB: // ES (NaN or inf)
					ci = s[i] != s[i] || isInf32(s[i])
				case 0xC: // NZ
					ci = s[i] != 0
				case 0xD: // NN
					ci = s[i] == s[i]
				case 0xE: // NI
					ci = !isInf32(s[i])
				case 0xF: // NS
					ci = !(s[i] != s[i] || isInf32(s[i]))
				}
				var cb uint32
				if ci {
					cb = 1
				}
				cc |= cb << i
				orv |= cb
				andv &= cb
				affected |= 1 << i
			}
			c.VfpuCtrl[vfpuCtlCC] = (c.VfpuCtrl[vfpuCtlCC] &^ affected) |
				((cc | orv<<4 | andv<<5) & affected)
			c.eatPfx()
			return
		case 2, 3: // vmin / vmax (NaN/inf ordered by magnitude, as hardware)
			for i := uint32(0); i < n; i++ {
				d[i] = vfpuMinMax(s[i], t[i], sub == 3)
			}
		case 6: // vsge
			for i := uint32(0); i < n; i++ {
				if s[i] != s[i] || t[i] != t[i] {
					d[i] = 0
				} else if s[i] >= t[i] {
					d[i] = 1
				} else {
					d[i] = 0
				}
			}
			c.vwrite(d, vd, n) // sat cannot matter for 0/1
			c.eatPfx()
			return
		case 7: // vslt
			for i := uint32(0); i < n; i++ {
				if s[i] != s[i] || t[i] != t[i] {
					d[i] = 0
				} else if s[i] < t[i] {
					d[i] = 1
				} else {
					d[i] = 0
				}
			}
			c.vwrite(d, vd, n)
			c.eatPfx()
			return
		case 5: // vscmp: sign of s-t
			for i := uint32(0); i < n; i++ {
				switch {
				case s[i] < t[i]:
					d[i] = -1
				case s[i] > t[i]:
					d[i] = 1
				default:
					d[i] = 0
				}
			}
		default:
			c.vfpuUnimpl(w, op)
			return
		}
	}
	c.applyPfxD(&d, n)
	c.vwrite(d, vd, n)
	c.eatPfx()
}

func isInf32(f float32) bool { return math.Float32bits(f)&0x7FFFFFFF == 0x7F800000 }

// vfpuMinMax orders NaN/inf by magnitude bits like the hardware: among negatives
// the integer comparison flips.
func vfpuMinMax(s, t float32, wantMax bool) float32 {
	if s != s || t != t || isInf32(s) || isInf32(t) {
		si, ti := int32(math.Float32bits(s)), int32(math.Float32bits(t))
		flip := si < 0 && ti < 0
		if (si < ti) != (wantMax != flip) {
			return s
		}
		return t
	}
	if (s < t) != wantMax {
		return s
	}
	return t
}

// vfpu4 executes the 0x34 group: single-operand ops, conversions, vcst, vcmov.
func (c *CPU) vfpu4(w, rs, rt uint32) {
	n := vfpuVecN(w)
	vd := w & 0x7F
	vs := (w >> 8) & 0x7F
	var d [4]float32

	switch rs {
	case 0x00: // VFPU4: vmov family
		s := c.vread(vs, n)
		c.applyPfxS(&s, n)
		for i := uint32(0); i < n; i++ {
			x := s[i]
			switch rt {
			case 0: // vmov
				d[i] = x
			case 1: // vabs (hardware folds this into the source prefix)
				d[i] = math.Float32frombits(math.Float32bits(x) &^ 0x80000000)
			case 2: // vneg
				d[i] = math.Float32frombits(math.Float32bits(x) ^ 0x80000000)
			case 4: // vsat0: clamp [0,1], -0 becomes +0, NaN passes
				if x <= 0 {
					d[i] = 0
				} else if x > 1 {
					d[i] = 1
				} else {
					d[i] = x
				}
			case 5: // vsat1: clamp [-1,1]
				if x < -1 {
					d[i] = -1
				} else if x > 1 {
					d[i] = 1
				} else {
					d[i] = x
				}
			case 16: // vrcp
				d[i] = 1 / x
			case 17: // vrsq
				d[i] = float32(1 / math.Sqrt(float64(x)))
			case 18: // vsin
				d[i] = vfpuSin(x)
			case 19: // vcos
				d[i] = vfpuCos(x)
			case 20: // vexp2
				d[i] = float32(math.Exp2(float64(x)))
			case 21: // vlog2
				d[i] = float32(math.Log2(float64(x)))
			case 22: // vsqrt
				d[i] = float32(math.Abs(math.Sqrt(float64(x))))
			case 23: // vasin (result in quarter turns)
				d[i] = float32(math.Asin(float64(x)) / (math.Pi / 2))
			case 24: // vnrcp
				d[i] = -(1 / x)
			case 26: // vnsin
				d[i] = -vfpuSin(x)
			case 28: // vrexp2
				d[i] = float32(1 / math.Exp2(float64(x)))
			default:
				switch rt {
				case 3: // vidt: identity vector (1 at the register's lane)
					off := vd & 3
					if n < 3 {
						off = vd & 1
					}
					for j := uint32(0); j < n; j++ {
						if j == off {
							d[j] = 1
						} else {
							d[j] = 0
						}
					}
				case 6: // vzero
					for j := uint32(0); j < n; j++ {
						d[j] = 0
					}
				case 7: // vone
					for j := uint32(0); j < n; j++ {
						d[j] = 1
					}
				default:
					c.vfpuUnimpl(w, 0x34)
					return
				}
				c.applyPfxD(&d, n)
				c.vwrite(d, vd, n)
				c.eatPfx()
				return
			}
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	case 0x01: // int <-> packed-integer conversions (bit data; prefixes don't apply)
		switch rt {
		case 27: // vs2i: unpack each s16 lane to the high half of an int
			s := c.vreadBits(vs, n)
			var bits [4]uint32
			for i := uint32(0); i < n; i++ {
				bits[2*i] = s[i] << 16
				bits[2*i+1] = s[i] & 0xFFFF0000
			}
			c.vwriteBits(bits, vd, 2*n)
			c.eatPfx()
		case 26: // vus2i: unpack each u16 lane, expanded to the positive int range
			s := c.vreadBits(vs, n)
			var bits [4]uint32
			for i := uint32(0); i < n; i++ {
				bits[2*i] = (s[i] & 0xFFFF) << 15
				bits[2*i+1] = (s[i] >> 16) << 15
			}
			c.vwriteBits(bits, vd, 2*n)
			c.eatPfx()
		case 30, 31: // vi2us / vi2s: pack int pairs to 16-bit halves
			s := c.vreadBits(vs, n)
			var bits [4]uint32
			for i := uint32(0); i < n; i += 2 {
				var lo, hi uint32
				if rt == 31 { // vi2s: arithmetic >>16
					lo = s[i] >> 16
					hi = s[i+1] >> 16
				} else { // vi2us: clamp negatives to 0, then >>15
					if int32(s[i]) > 0 {
						lo = s[i] >> 15
					}
					if int32(s[i+1]) > 0 {
						hi = s[i+1] >> 15
					}
				}
				bits[i/2] = lo&0xFFFF | hi<<16
			}
			c.vwriteBits(bits, vd, n/2)
			c.eatPfx()
		case 28, 29: // vi2uc / vi2c: pack four ints to bytes
			s := c.vreadBits(vs, n)
			var bits [4]uint32
			var out uint32
			for i := uint32(0); i < n; i++ {
				var b uint32
				if rt == 29 { // vi2c: top byte
					b = s[i] >> 24
				} else if int32(s[i]) > 0 { // vi2uc: clamp negatives, saturate via >>23
					b = s[i] >> 23
					if b > 0xFF {
						b = 0xFF
					}
				}
				out |= b << (8 * i)
			}
			bits[0] = out
			c.vwriteBits(bits, vd, 1)
			c.eatPfx()
		default:
			c.vfpuUnimpl(w, 0x34)
		}

	case 0x02: // VFPU9 subset
		switch rt {
		case 2: // vbfy1
			s := c.vread(vs, n)
			c.applyPfxS(&s, n)
			d[0], d[1] = s[0]+s[1], s[0]-s[1]
			if n == 4 {
				d[2], d[3] = s[2]+s[3], s[2]-s[3]
			}
		case 3: // vbfy2 (quad only)
			s := c.vread(vs, n)
			c.applyPfxS(&s, n)
			d[0], d[1] = s[0]+s[2], s[1]+s[3]
			d[2], d[3] = s[0]-s[2], s[1]-s[3]
		case 4: // vocp: one's complement 1-x
			s := c.vread(vs, n)
			c.applyPfxS(&s, n)
			for i := uint32(0); i < n; i++ {
				d[i] = 1 - s[i]
			}
		case 6, 7: // vfad (sum) / vavg (mean) -> single
			s := c.vread(vs, n)
			c.applyPfxS(&s, n)
			var sum float32
			for i := uint32(0); i < n; i++ {
				sum += s[i]
			}
			if rt == 7 {
				sum /= float32(n)
			}
			d[0] = sum
			c.applyPfxD(&d, 1)
			c.vwrite(d, vd, 1)
			c.eatPfx()
			return
		case 10: // vsgn
			s := c.vread(vs, n)
			c.applyPfxS(&s, n)
			for i := uint32(0); i < n; i++ {
				switch {
				case s[i] > 0:
					d[i] = 1
				case s[i] < 0:
					d[i] = -1
				default:
					d[i] = 0 // 0 and NaN
				}
			}
		case 16: // vmfvc: vd = ctrl[imm7 from vs field]
			imm := (w >> 8) & 0x7F
			var v uint32
			if imm < 16 {
				v = c.VfpuCtrl[imm]
			}
			c.V[vfpuSingle(vd)&127] = v
			return
		case 17: // vmtvc: ctrl[imm7 from vd field] = vs
			imm := w & 0x7F
			if imm < 16 && vfpuCtrlMask[imm] != 0 {
				c.VfpuCtrl[imm] = c.V[vfpuSingle(vs)&127] & vfpuCtrlMask[imm]
			}
			return
		default:
			c.vfpuUnimpl(w, 0x34)
			return
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	case 0x03: // vcst
		cst := vcstConstants[rt&31]
		for i := uint32(0); i < n; i++ {
			d[i] = cst
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	case 0x10, 0x11, 0x12, 0x13: // vf2in/vf2iz/vf2iu/vf2id: float -> int, scale 2^imm
		s := c.vread(vs, n)
		c.applyPfxS(&s, n)
		mult := float64(uint64(1) << (rt & 31))
		var bits [4]uint32
		for i := uint32(0); i < n; i++ {
			if s[i] != s[i] {
				bits[i] = 0x7FFFFFFF
				continue
			}
			sv := float64(s[i]) * mult
			switch {
			case sv > float64(math.MaxInt32):
				bits[i] = 0x7FFFFFFF
			case sv <= float64(math.MinInt32):
				bits[i] = 0x80000000
			default:
				var r float64
				switch rs {
				case 0x10:
					r = math.RoundToEven(sv)
				case 0x11:
					r = math.Trunc(sv)
				case 0x12:
					r = math.Ceil(sv)
				case 0x13:
					r = math.Floor(sv)
				}
				bits[i] = uint32(int32(r))
			}
		}
		c.vwriteBits(bits, vd, n) // mask applies; sat does not
		c.eatPfx()

	case 0x14: // vi2f: int -> float, scale 2^-imm
		s := c.vreadBits(vs, n)
		mult := float32(1) / float32(uint64(1)<<(rt&31))
		for i := uint32(0); i < n; i++ {
			d[i] = float32(int32(s[i])) * mult
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	case 0x15: // vcmov: conditional move on the vcmp condition codes
		tf := (w >> 19) & 1
		imm3 := (w >> 16) & 7
		s := c.vread(vs, n)
		c.applyPfxS(&s, n)
		d = c.vread(vd, n)
		c.applyPfxT(&d, n) // the T prefix applies to the destination-as-source
		cc := c.VfpuCtrl[vfpuCtlCC]
		if imm3 < 6 {
			if (cc>>imm3)&1 == 1-tf {
				for i := uint32(0); i < n; i++ {
					d[i] = s[i]
				}
			}
		} else if imm3 == 6 { // per-lane bits
			for i := uint32(0); i < n; i++ {
				if (cc>>i)&1 == 1-tf {
					d[i] = s[i]
				}
			}
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	default:
		c.vfpuUnimpl(w, 0x34)
	}
}

// vfpu5 executes the prefix latches and the immediate loads.
func (c *CPU) vfpu5(w uint32) {
	switch (w >> 24) & 3 {
	case 0:
		c.VfpuCtrl[vfpuCtlSPfx] = w & 0xFFFFF
	case 1:
		c.VfpuCtrl[vfpuCtlTPfx] = w & 0xFFFFF
	case 2:
		c.VfpuCtrl[vfpuCtlDPfx] = w & 0xFFF
	case 3: // viim.s / vfim.s
		vt := (w >> 16) & 0x7F
		var d [4]float32
		if w&(1<<23) != 0 {
			d[0] = float16to32(uint16(w))
		} else {
			d[0] = float32(int32(int16(w)))
		}
		c.applyPfxD(&d, 1)
		c.vwrite(d, vt, 1)
		c.eatPfx()
	}
}

// vfpuMatrix executes the 0x3C group: vmmul, vtfm/vhtfm, vmscl, vqmul/vcrsp,
// vrot and the matrix set/move ops. Matrix ops ignore pending prefixes (see the
// package comment); vector-output ops honor the destination prefix.
func (c *CPU) vfpuMatrix(w, rt uint32) {
	n := vfpuVecN(w)
	vd := w & 0x7F
	vs := (w >> 8) & 0x7F
	vt := (w >> 16) & 0x7F

	switch (w >> 21) & 31 {
	case 0, 1, 2, 3: // vmmul: D = S^T x T over the read layout (m[col*4+row])
		s := c.mread(vs, n)
		t := c.mread(vt, n)
		var d [16]float32
		for a := uint32(0); a < n; a++ {
			for b := uint32(0); b < n; b++ {
				var sum float32
				for k := uint32(0); k < n; k++ {
					sum += s[b*4+k] * t[a*4+k]
				}
				d[a*4+b] = sum
			}
		}
		c.mwrite(d, vd, n)
		c.eatPfx()

	case 16, 17, 18, 19: // vmscl: scale a matrix by the single T
		s := c.mread(vs, n)
		t := c.vread(vt, 1)
		var d [16]float32
		for a := uint32(0); a < n; a++ {
			for b := uint32(0); b < n; b++ {
				d[a*4+b] = s[a*4+b] * t[0]
			}
		}
		c.mwrite(d, vd, n)
		c.eatPfx()

	case 20, 21, 22, 23: // vcrsp.t (cross product) / vqmul.q (quaternion product)
		s := c.vread(vs, n)
		t := c.vread(vt, n)
		var d [4]float32
		switch n {
		case 3:
			d[0] = s[1]*t[2] - s[2]*t[1]
			d[1] = s[2]*t[0] - s[0]*t[2]
			d[2] = s[0]*t[1] - s[1]*t[0]
		case 4:
			d[0] = s[0]*t[3] + s[1]*t[2] - s[2]*t[1] + s[3]*t[0]
			d[1] = -s[0]*t[2] + s[1]*t[3] + s[2]*t[0] + s[3]*t[1]
			d[2] = s[0]*t[1] - s[1]*t[0] + s[2]*t[3] + s[3]*t[2]
			d[3] = -s[0]*t[0] - s[1]*t[1] - s[2]*t[2] + s[3]*t[3]
		default:
			c.vfpuUnimpl(w, 0x3C)
			return
		}
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	case 28: // matrix set/move
		switch rt & 0xF {
		case 0: // vmmov
			c.mwrite(c.mread(vs, n), vd, n)
		case 3: // vmidt
			var d [16]float32
			for i := uint32(0); i < n; i++ {
				d[i*4+i] = 1
			}
			c.mwrite(d, vd, n)
		case 6: // vmzero
			c.mwrite([16]float32{}, vd, n)
		case 7: // vmone
			var d [16]float32
			for a := uint32(0); a < n; a++ {
				for b := uint32(0); b < n; b++ {
					d[a*4+b] = 1
				}
			}
			c.mwrite(d, vd, n)
		default:
			c.vfpuUnimpl(w, 0x3C)
			return
		}
		c.eatPfx()

	case 29: // vrot: one row of a rotation matrix from the single S (quarter turns)
		imm := (w >> 16) & 0x1F
		x := c.vread(vs, 1)[0]
		sin, cos := vfpuSin(x), vfpuCos(x)
		if imm&0x10 != 0 {
			sin = -sin
		}
		sinLane := (imm >> 2) & 3
		cosLane := imm & 3
		var d [4]float32
		if sinLane == cosLane {
			for i := uint32(0); i < n; i++ {
				d[i] = sin
			}
		} else {
			d[sinLane] = sin
		}
		d[cosLane] = cos
		c.applyPfxD(&d, n)
		c.vwrite(d, vd, n)
		c.eatPfx()

	default: // vtfm2/3/4, vhtfm2/3/4: matrix x vector transform
		ins := (w >> 23) & 3 // matrix side - 1
		if ins == 0 {
			c.vfpuUnimpl(w, 0x3C)
			return
		}
		side := ins + 1
		s := c.mread(vs, side)
		t := c.vread(vt, n)
		hom := ins >= n // vhtfm: input one element short, w wired to 1
		tn := n
		if side < tn {
			tn = side
		}
		var d [4]float32
		for i := uint32(0); i <= ins; i++ {
			var sum float32
			for k := uint32(0); k < tn; k++ {
				sum += s[i*4+k] * t[k]
			}
			if hom {
				sum += s[i*4+ins]
			}
			d[i] = sum
		}
		c.applyPfxD(&d, side)
		c.vwrite(d, vd, side)
		c.eatPfx()
	}
}
