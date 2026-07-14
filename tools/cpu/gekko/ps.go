package gekko

// ps.go is the paired-single unit and the quantised load/store — the two pieces of
// hardware Nintendo added to the PowerPC 750, and the reason a GameCube can do the vector
// arithmetic a 3-D game needs on a processor that would otherwise be a desktop G3.
//
// # Paired singles
//
// Each floating-point register holds two singles, PS0 and PS1, and one ps_ instruction
// operates on both. That much is ordinary SIMD. What is not ordinary is the cross-slot
// wiring: several instructions deliberately read the *wrong* slot, because that is what
// makes a dot product cheap.
//
//	ps_muls0   D = A × C[ps0]     — multiply both slots by C's ps0, broadcast
//	ps_muls1   D = A × C[ps1]     — ...or by its ps1
//	ps_madds0  D = A × C[ps0] + B
//	ps_sum0    D[ps0] = A[ps0] + B[ps1];  D[ps1] = C[ps1]   — a horizontal add
//	ps_sum1    D[ps0] = C[ps0];           D[ps1] = A[ps0] + B[ps1]
//	ps_mergeXY D[ps0] = A[psX];   D[ps1] = B[psY]           — a shuffle
//
// ps_sum0 and ps_sum1 are the horizontal add — the operation SIMD units usually make you
// beg for — and the merges are a full 2-way shuffle. Together with the broadcast multiplies
// they are exactly the instruction set for a 3×3 matrix multiply, which is what the machine
// spends its time doing. Getting a slot wrong in this table produces a core that runs and
// draws geometry that is subtly, unfixably wrong, so the table above is the specification
// and the code below is a transcription of it.
//
// # The quantiser
//
// psq_l reads one or two values from memory in one of five formats, multiplies by a power
// of two, and lands two floats in a register. psq_st reverses it, saturating on the way
// out. The format and the scale are in neither the instruction nor the operands: they come
// from one of eight graphics quantisation registers, chosen by a three-bit field. So a
// GameCube stores its vertex data as bytes and halfwords and reads it as floats, for free,
// in the load — which is why its models are small.
//
// The scale is a *signed 6-bit* exponent, and the load and the store use it in opposite
// directions: a load multiplies by 2^-scale, a store by 2^+scale. Reversing that is the
// single easiest way to produce a core that works on every test and puts the geometry in
// the wrong place, so the vector suite walks all 64 scales × 5 formats × both directions
// exhaustively rather than sampling them.

import "math"

// The quantisation formats, as the GQR's type field names them.
const (
	qFloat = 0
	qU8    = 4
	qU16   = 5
	qS8    = 6
	qS16   = 7
)

// gqrLoad reads the load half of a graphics quantisation register: type in bits 16-18,
// scale in bits 24-29 (a signed six-bit value).
func gqrLoad(g uint32) (typ uint32, scale int32) {
	typ = (g >> 16) & 7
	scale = int32(g<<2) >> 26 // sign-extend the six-bit field at bits 24-29
	return
}

// gqrStore reads the store half: type in bits 0-2, scale in bits 8-13.
func gqrStore(g uint32) (typ uint32, scale int32) {
	typ = g & 7
	scale = int32(g<<18) >> 26
	return
}

// dequantize turns one stored value into a float. The scale divides: a value stored with
// scale s was multiplied by 2^s on the way in, so it is multiplied by 2^-s on the way out.
func (c *CPU) dequantize(raw uint32, typ uint32, scale int32) float64 {
	var v float64
	switch typ {
	case qFloat:
		return float64(math.Float32frombits(raw))
	case qU8:
		v = float64(uint8(raw))
	case qU16:
		v = float64(uint16(raw))
	case qS8:
		v = float64(int8(raw))
	case qS16:
		v = float64(int16(raw))
	default:
		// Types 1-3 are not defined. The hardware does something; we do not know what,
		// and inventing a behaviour would be inventing a fact.
		c.Halt("gekko: a quantised load used the undefined GQR type %d at 0x%08X", typ, c.PC-4)
		return 0
	}
	return v * math.Ldexp(1, int(-scale))
}

// quantize turns a float into a stored value, saturating at the format's limits. The
// saturation is not optional and not a rounding detail: a game that scales its vertices to
// fill a signed byte relies on the clamp, and a core that wrapped instead would fold the
// far side of a model through the near side.
func (c *CPU) quantize(v float64, typ uint32, scale int32) uint32 {
	if typ == qFloat {
		return math.Float32bits(float32(v))
	}
	v *= math.Ldexp(1, int(scale))

	// A NaN quantises to zero rather than to whatever the conversion happens to produce.
	if math.IsNaN(v) {
		v = 0
	}
	switch typ {
	case qU8:
		return uint32(uint8(clamp(v, 0, 255)))
	case qU16:
		return uint32(uint16(clamp(v, 0, 65535)))
	case qS8:
		return uint32(uint8(int8(clamp(v, -128, 127))))
	case qS16:
		return uint32(uint16(int16(clamp(v, -32768, 32767))))
	}
	c.Halt("gekko: a quantised store used the undefined GQR type %d at 0x%08X", typ, c.PC-4)
	return 0
}

// clamp saturates and rounds. The order matters: the value is rounded to an integer first
// and then clamped, so that a value just past the limit lands exactly on it.
func clamp(v, lo, hi float64) float64 {
	v = math.Trunc(v)
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// qsize is how many bytes one value of a format occupies.
func qsize(typ uint32) uint32 {
	switch typ {
	case qFloat:
		return 4
	case qU16, qS16:
		return 2
	case qU8, qS8:
		return 1
	}
	return 4
}

// readQ reads one stored value of a format from memory.
func (c *CPU) readQ(ea uint32, typ uint32) uint32 {
	switch qsize(typ) {
	case 1:
		return uint32(c.read8(ea))
	case 2:
		return uint32(c.read16(ea))
	}
	return c.read32(ea)
}

func (c *CPU) writeQ(ea uint32, typ uint32, v uint32) {
	switch qsize(typ) {
	case 1:
		c.write8(ea, uint8(v))
	case 2:
		c.write16(ea, uint16(v))
	default:
		c.write32(ea, v)
	}
}

// psqLoad performs a quantised load. wBit selects one value rather than a pair — and when
// only one is loaded, PS1 is set to 1.0, not left alone and not zeroed. That is the
// hardware's choice and it is a load-bearing one: it makes a 3-vector loaded into a pair
// of registers come out as a homogeneous coordinate with w = 1.
func (c *CPU) psqLoad(d, ea, gqr, wBit uint32) {
	typ, scale := gqrLoad(c.GQR[gqr&7])
	if wBit != 0 {
		c.FPR[d].PS0 = c.dequantize(c.readQ(ea, typ), typ, scale)
		c.FPR[d].PS1 = 1.0
		return
	}
	sz := qsize(typ)
	c.FPR[d].PS0 = c.dequantize(c.readQ(ea, typ), typ, scale)
	c.FPR[d].PS1 = c.dequantize(c.readQ(ea+sz, typ), typ, scale)
}

func (c *CPU) psqStore(s, ea, gqr, wBit uint32) {
	typ, scale := gqrStore(c.GQR[gqr&7])
	c.writeQ(ea, typ, c.quantize(c.FPR[s].PS0, typ, scale))
	if wBit != 0 {
		return // one value only; PS1 is not written
	}
	c.writeQ(ea+qsize(typ), typ, c.quantize(c.FPR[s].PS1, typ, scale))
}

// execPSQ runs the non-indexed quantised loads and stores: opcodes 56, 57, 60, 61.
func (c *CPU) execPSQ(w uint32) {
	if !c.psEnabled() {
		c.Halt("gekko: a quantised load/store ran with the paired-single unit off (HID2 = 0x%08X) at 0x%08X", c.HID2, c.PC-4)
		return
	}
	d := rs(w)
	gqr := (w >> 12) & 7
	wBit := (w >> 15) & 1
	disp := uint32(psqDisp(w))

	switch opcd(w) {
	case 56: // psq_l
		c.psqLoad(d, c.raOrZero(w)+disp, gqr, wBit)
	case 57: // psq_lu
		ea := c.GPR[ra(w)] + disp
		c.psqLoad(d, ea, gqr, wBit)
		c.GPR[ra(w)] = ea
	case 60: // psq_st
		c.psqStore(d, c.raOrZero(w)+disp, gqr, wBit)
	case 61: // psq_stu
		ea := c.GPR[ra(w)] + disp
		c.psqStore(d, ea, gqr, wBit)
		c.GPR[ra(w)] = ea
	}
}

// execPS runs primary opcode 4: the paired-single arithmetic, the indexed quantised
// loads, and dcbz_l.
func (c *CPU) execPS(w, pc uint32) {
	// The indexed quantised forms and dcbz_l come first, because dcbz_l is not a
	// floating-point instruction and must work with the paired-single unit off.
	switch xo6(w) {
	case 6, 7, 38, 39:
		if !c.psEnabled() {
			c.Halt("gekko: an indexed quantised load/store ran with the paired-single unit off at 0x%08X", pc)
			return
		}
		d := rs(w)
		gqr := (w >> 12) & 7
		wBit := (w >> 15) & 1
		switch xo6(w) {
		case 6: // psq_lx
			c.psqLoad(d, c.eax(w), gqr, wBit)
		case 7: // psq_stx
			c.psqStore(d, c.eax(w), gqr, wBit)
		case 38: // psq_lux
			ea := c.eaxU(w)
			c.psqLoad(d, ea, gqr, wBit)
			c.GPR[ra(w)] = ea
		case 39: // psq_stux
			ea := c.eaxU(w)
			c.psqStore(d, ea, gqr, wBit)
			c.GPR[ra(w)] = ea
		}
		return
	}
	if xo10(w) == 1014 { // dcbz_l
		c.dcbzL(c.eax(w))
		return
	}

	if !c.psEnabled() {
		c.Halt("gekko: a paired-single instruction ran with the unit off (HID2 = 0x%08X) at 0x%08X", c.HID2, pc)
		return
	}

	d, a, b, cc := rs(w), ra(w), rb(w), rc(w)
	A, B, C := c.FPR[a], c.FPR[b], c.FPR[cc]

	// The A-forms. Note which slot of C each broadcast multiply reads — that is the whole
	// of the cross-slot wiring, and it is the part that must not be guessed.
	switch xo5(w) {
	case 21: // ps_add
		c.psResult(w, d, c.fadd(A.PS0, B.PS0), c.fadd(A.PS1, B.PS1))
		return
	case 20: // ps_sub
		c.psResult(w, d, c.fsub(A.PS0, B.PS0), c.fsub(A.PS1, B.PS1))
		return
	case 25: // ps_mul
		c.psResult(w, d, c.fmul(A.PS0, C.PS0), c.fmul(A.PS1, C.PS1))
		return
	case 18: // ps_div
		c.psResult(w, d, c.fdiv(A.PS0, B.PS0), c.fdiv(A.PS1, B.PS1))
		return
	case 12: // ps_muls0 — both slots multiplied by C's ps0
		c.psResult(w, d, c.fmul(A.PS0, C.PS0), c.fmul(A.PS1, C.PS0))
		return
	case 13: // ps_muls1 — both slots multiplied by C's ps1
		c.psResult(w, d, c.fmul(A.PS0, C.PS1), c.fmul(A.PS1, C.PS1))
		return
	case 29: // ps_madd
		c.psResult(w, d, c.fmaddRaw(A.PS0, C.PS0, B.PS0), c.fmaddRaw(A.PS1, C.PS1, B.PS1))
		return
	case 28: // ps_msub
		c.psResult(w, d, c.fmaddRaw(A.PS0, C.PS0, -B.PS0), c.fmaddRaw(A.PS1, C.PS1, -B.PS1))
		return
	case 31: // ps_nmadd
		c.psResult(w, d, -c.fmaddRaw(A.PS0, C.PS0, B.PS0), -c.fmaddRaw(A.PS1, C.PS1, B.PS1))
		return
	case 30: // ps_nmsub
		c.psResult(w, d, -c.fmaddRaw(A.PS0, C.PS0, -B.PS0), -c.fmaddRaw(A.PS1, C.PS1, -B.PS1))
		return
	case 14: // ps_madds0 — the broadcast multiply-accumulate
		c.psResult(w, d, c.fmaddRaw(A.PS0, C.PS0, B.PS0), c.fmaddRaw(A.PS1, C.PS0, B.PS1))
		return
	case 15: // ps_madds1
		c.psResult(w, d, c.fmaddRaw(A.PS0, C.PS1, B.PS0), c.fmaddRaw(A.PS1, C.PS1, B.PS1))
		return
	case 10: // ps_sum0 — the horizontal add: ps0 gets A.ps0 + B.ps1, ps1 passes C through
		c.psResult(w, d, c.fadd(A.PS0, B.PS1), C.PS1)
		return
	case 11: // ps_sum1 — the other way round
		c.psResult(w, d, C.PS0, c.fadd(A.PS0, B.PS1))
		return
	case 23: // ps_sel — a branchless select, per slot
		p0, p1 := B.PS0, B.PS1
		if A.PS0 >= 0 && !math.IsNaN(A.PS0) {
			p0 = C.PS0
		}
		if A.PS1 >= 0 && !math.IsNaN(A.PS1) {
			p1 = C.PS1
		}
		c.FPR[d].PS0, c.FPR[d].PS1 = p0, p1
		c.rcF(w)
		return
	case 24: // ps_res
		c.psResult(w, d, c.fres(B.PS0), c.fres(B.PS1))
		return
	case 26: // ps_rsqrte
		c.psResult(w, d, c.frsqrte(B.PS0), c.frsqrte(B.PS1))
		return
	}

	// The ten-bit forms: the shuffles, the sign flips, and the compares.
	switch xo10(w) {
	case 528: // ps_merge00 — D = (A.ps0, B.ps0)
		c.FPR[d] = FPR{PS0: A.PS0, PS1: B.PS0}
		c.rcF(w)
	case 560: // ps_merge01 — D = (A.ps0, B.ps1)
		c.FPR[d] = FPR{PS0: A.PS0, PS1: B.PS1}
		c.rcF(w)
	case 592: // ps_merge10 — D = (A.ps1, B.ps0)
		c.FPR[d] = FPR{PS0: A.PS1, PS1: B.PS0}
		c.rcF(w)
	case 624: // ps_merge11 — D = (A.ps1, B.ps1)
		c.FPR[d] = FPR{PS0: A.PS1, PS1: B.PS1}
		c.rcF(w)
	case 40: // ps_neg
		c.FPR[d] = FPR{PS0: negf(B.PS0), PS1: negf(B.PS1)}
		c.rcF(w)
	case 72: // ps_mr
		c.FPR[d] = B
		c.rcF(w)
	case 136: // ps_nabs
		c.FPR[d] = FPR{PS0: -math.Abs(B.PS0), PS1: -math.Abs(B.PS1)}
		c.rcF(w)
	case 264: // ps_abs
		c.FPR[d] = FPR{PS0: math.Abs(B.PS0), PS1: math.Abs(B.PS1)}
		c.rcF(w)
	case 0: // ps_cmpu0 — compare the ps0 slots
		c.fcmp(crfD(w), A.PS0, B.PS0, false)
	case 32: // ps_cmpo0
		c.fcmp(crfD(w), A.PS0, B.PS0, true)
	case 64: // ps_cmpu1 — compare the ps1 slots
		c.fcmp(crfD(w), A.PS1, B.PS1, false)
	case 96: // ps_cmpo1
		c.fcmp(crfD(w), A.PS1, B.PS1, true)
	default:
		c.Halt("gekko: unimplemented opcode 4 extended %d (word 0x%08X) at 0x%08X", xo10(w), w, pc)
	}
}

// negf flips a sign bit without arithmetic, so that negating a NaN keeps its payload.
func negf(v float64) float64 {
	return math.Float64frombits(math.Float64bits(v) ^ (1 << 63))
}

// psResult stores both halves of a paired-single result. Each slot is rounded to single —
// they are singles, whatever the internal width — and FPRF reports the ps0 slot, which is
// what the architecture says and what a following ps_cmp reads.
func (c *CPU) psResult(w, d uint32, p0, p1 float64) {
	c.FPR[d].PS0 = f32(p0)
	c.FPR[d].PS1 = f32(p1)
	c.setFPRF(c.FPR[d].PS0)
	c.rcF(w)
}
