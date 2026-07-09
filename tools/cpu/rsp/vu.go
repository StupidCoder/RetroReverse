package rsp

// vu.go is COP2, the RSP's vector unit — where the geometry actually happens.
//
// A vector register is 128 bits seen as eight signed 16-bit lanes, lane 0 being
// the most significant. Behind them sits an accumulator of eight 48-bit lanes,
// which the multiply instructions write and the multiply-accumulate ones add
// into. Nothing returns a 48-bit value: each instruction extracts a 16-bit slice
// and clamps it, and *which* slice, and *how* it clamps, is the only difference
// between most of the multiply opcodes.
//
// Every ALU instruction carries a four-bit element specifier that rewrites which
// lane of the second operand each lane reads. It is how a microcode multiplies a
// vector by a scalar without moving anything: element 8..15 broadcast a single
// lane across all eight, and 2..7 select within pairs or quads. See element().
//
// The reciprocal instructions index a table burnt into the RSP's silicon, which
// is not on the cartridge and could not be derived from it — see rcprom.go.

// accMask is the accumulator's 48-bit width.
const accMask = 0xFFFFFFFFFFFF

// element rewrites which lane of the second operand lane i reads.
//
//	0,1     identity: lane i reads lane i
//	2,3     broadcast within each pair
//	4..7    broadcast within each quad
//	8..15   broadcast one lane across all eight
func element(e, i uint32) uint32 {
	switch {
	case e < 2:
		return i
	case e < 4:
		return i&^1 | e&1
	case e < 8:
		return i&^3 | e&3
	default:
		return e & 7
	}
}

// vte reads lane i of a source-register snapshot under the element specifier.
func vte(v [8]uint16, e, i uint32) uint16 { return v[element(e, i)] }

// acc reads a lane of the accumulator, sign-extended from 48 bits.
func (c *CPU) acc(i uint32) int64 { return int64(c.Acc[i]<<16) >> 16 }

func (c *CPU) setAcc(i uint32, v int64) { c.Acc[i] = uint64(v) & accMask }

// setAccLo writes only the accumulator's low 16 bits, leaving the middle and
// high slices as they were.
//
// This is the difference between the multiply instructions and everything else.
// A multiply produces a 48-bit product and owns the whole accumulator lane. An
// add, a compare, a logical operation or a lane move produces sixteen bits, and
// writes only those — the slices above survive. Microcode relies on it: it reads
// them back with VSAR long after the operation that filled them, and an
// implementation that zeroes them on every VADD quietly destroys geometry.
func (c *CPU) setAccLo(i uint32, v uint16) {
	c.Acc[i] = c.Acc[i]&^0xFFFF | uint64(v)
}

// clampS is the signed saturation the "high slice" instructions apply.
func clampS(v int64) uint16 {
	if v > 32767 {
		return 0x7FFF
	}
	if v < -32768 {
		return 0x8000
	}
	return uint16(int16(v))
}

// clampU is VMULU's and VMACU's saturation. It is not a clamp to the unsigned
// 16-bit range: the comparison happens at the *signed* boundary, so an
// accumulator slice of 0x8000 — representable unsigned — still saturates to
// 0xFFFF. Negative accumulators clamp to zero. n64-systemtest pins 0x8000×0x8000
// to 0xFFFF, which a true unsigned clamp would return as 0x8000.
func clampU(v int64) uint16 {
	if v < 0 {
		return 0
	}
	if v > 32767 {
		return 0xFFFF
	}
	return uint16(v)
}

// clampLow extracts the accumulator's low slice. It returns the raw low 16 bits
// unless the value above them has overflowed a signed 16-bit range, in which
// case it saturates — this is what separates VMUDN/VMADN from a plain truncation.
func clampLow(acc int64) uint16 {
	hi := acc >> 16
	if hi < -32768 {
		return 0x0000
	}
	if hi > 32767 {
		return 0xFFFF
	}
	return uint16(acc)
}

func s16(v uint16) int64 { return int64(int16(v)) }
func u16(v uint16) int64 { return int64(v) }

// bit reads flag lane i out of a 16-bit flag register.
func bit(f uint16, i uint32) bool { return f&(1<<i) != 0 }

func setBit(f *uint16, i uint32, on bool) {
	if on {
		*f |= 1 << i
	} else {
		*f &^= 1 << i
	}
}

// cop2 dispatches a COP2 instruction: a register move, or the vector ALU.
func (c *CPU) cop2(w uint32) {
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31

	if w&(1<<25) == 0 { // register moves
		e := (w >> 7) & 15
		switch rs {
		case 0x00: // mfc2: a 16-bit lane, read at a byte offset, sign-extended
			b := e & 15
			v := uint32(c.vecByte(rd, b))<<8 | uint32(c.vecByte(rd, (b+1)&15))
			c.set(rt, uint32(int32(int16(uint16(v)))))
		case 0x04: // mtc2
			b := e & 15
			c.setVecByte(rd, b, byte(c.R[rt]>>8))
			if b+1 < 16 {
				c.setVecByte(rd, b+1, byte(c.R[rt]))
			}
		case 0x02: // cfc2
			c.set(rt, uint32(int32(int16(c.ctrl(rd&3)))))
		case 0x06: // ctc2
			c.setCtrl(rd&3, uint16(c.R[rt]))
		default:
			c.unimpl(w)
		}
		return
	}
	c.vectorALU(w, w&0x3F, rs&15, rt, rd, shamt)
}

// ctrl and setCtrl map the control-register index to a flag register. There are
// only three, but the two-bit index has four values: the hardware maps both 2
// and 3 to VCE, and CFC2/CTC2 with indices above 3 reduce to their low two bits.
func (c *CPU) ctrl(i uint32) uint16 {
	switch i & 3 {
	case 0:
		return c.VCO
	case 1:
		return c.VCC
	default:
		return uint16(c.VCE)
	}
}

func (c *CPU) setCtrl(i uint32, v uint16) {
	switch i & 3 {
	case 0:
		c.VCO = v
	case 1:
		c.VCC = v
	default:
		c.VCE = uint8(v)
	}
}

// vecByte and setVecByte address a vector register as sixteen bytes, which is
// how the load/store family and the scalar moves see it.
func (c *CPU) vecByte(r, b uint32) byte {
	lane := (b >> 1) & 7
	if b&1 == 0 {
		return byte(c.V[r][lane] >> 8)
	}
	return byte(c.V[r][lane])
}

func (c *CPU) setVecByte(r, b uint32, v byte) {
	lane := (b >> 1) & 7
	if b&1 == 0 {
		c.V[r][lane] = c.V[r][lane]&0x00FF | uint16(v)<<8
	} else {
		c.V[r][lane] = c.V[r][lane]&0xFF00 | uint16(v)
	}
}

// vectorALU runs one vector operation. vd is the destination, vs the first
// source, vt the second (through the element specifier).
//
// Both sources are copied before anything is written. The hardware reads its
// operands in full before writeback, so an instruction whose destination is also
// a source still computes from the old values — visible whenever the element
// specifier makes lane i read some other lane j that was already overwritten.
// n64-systemtest exercises exactly this with vd == vt and vd == vs.
func (c *CPU) vectorALU(w, funct, e, vt, vs, vd uint32) {
	vsv, vtv := c.V[vs], c.V[vt]
	switch funct {
	// --- multiply: write the accumulator outright ---------------------------
	case 0x00: // vmulf: signed fractional multiply, rounded
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, p<<1+0x8000)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x01: // vmulu: as vmulf, saturating unsigned
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, p<<1+0x8000)
			c.V[vd][i] = clampU(c.acc(i) >> 16)
		}
	case 0x04: // vmudl: unsigned * unsigned, low slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, u16(vsv[i])*u16(vte(vtv, e, i))>>16)
			c.V[vd][i] = uint16(c.acc(i))
		}
	case 0x05: // vmudm: signed * unsigned, high slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, s16(vsv[i])*u16(vte(vtv, e, i)))
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x06: // vmudn: unsigned * signed, low slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, u16(vsv[i])*s16(vte(vtv, e, i)))
			c.V[vd][i] = uint16(c.acc(i))
		}
	case 0x07: // vmudh: signed * signed, shifted into the high slice
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, p<<16)
			c.V[vd][i] = clampS(p)
		}
	case 0x03: // vmulq: the MPEG quantiser multiply
		// The product sits in the accumulator's upper 32 bits; a negative one is
		// rounded up by 31 in its integer part before the result is taken from
		// one bit below the mid slice, with the low nibble stripped.
		for i := uint32(0); i < 8; i++ {
			a := s16(vsv[i]) * s16(vte(vtv, e, i)) << 16
			if a < 0 {
				a += 0x1F0000
			}
			c.setAcc(i, a)
			c.V[vd][i] = clampS(a>>17) & 0xFFF0
		}
	case 0x02, 0x0A: // vrndp, vrndn: the MPEG rounding adds
		// The operand is vt through the element specifier; the vs *field* is not
		// a register here but a flag — odd values shift the product into the
		// accumulator's upper half. VRNDP adds only while the accumulator is
		// non-negative, VRNDN only while negative, which is what makes them
		// rounds toward the respective direction.
		for i := uint32(0); i < 8; i++ {
			p := s16(vte(vtv, e, i))
			if vs&1 != 0 {
				p <<= 16
			}
			a := c.acc(i)
			if (funct == 0x02) == (a >= 0) {
				a += p
			}
			c.setAcc(i, a)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x0B: // vmacq: the matching accumulator rounder; reads no operands
		for i := uint32(0); i < 8; i++ {
			a := c.acc(i)
			if a&0x200000 == 0 {
				if a>>22 < 0 {
					a += 0x200000
				} else if a>>22 > 0 {
					a -= 0x200000
				}
			}
			c.setAcc(i, a)
			r := uint16(a >> 17)
			if a < 0 && (^a)>>32 != 0 {
				r = 0x8000
			} else if a >= 0 && a>>32 != 0 {
				r = 0x7FFF
			}
			c.V[vd][i] = r & 0xFFF0
		}

	// --- multiply-accumulate: add into the accumulator -----------------------
	case 0x08: // vmacf
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, c.acc(i)+p<<1)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x09: // vmacu
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, c.acc(i)+p<<1)
			c.V[vd][i] = clampU(c.acc(i) >> 16)
		}
	case 0x0C: // vmadl
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+u16(vsv[i])*u16(vte(vtv, e, i))>>16)
			c.V[vd][i] = clampLow(c.acc(i))
		}
	case 0x0D: // vmadm
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+s16(vsv[i])*u16(vte(vtv, e, i)))
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x0E: // vmadn
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+u16(vsv[i])*s16(vte(vtv, e, i)))
			c.V[vd][i] = clampLow(c.acc(i))
		}
	case 0x0F: // vmadh
		for i := uint32(0); i < 8; i++ {
			p := s16(vsv[i]) * s16(vte(vtv, e, i))
			c.setAcc(i, c.acc(i)+p<<16)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}

	// --- add and subtract, through the carry flags ---------------------------
	case 0x10: // vadd: adds the carry VADDC left behind
		for i := uint32(0); i < 8; i++ {
			sum := s16(vsv[i]) + s16(vte(vtv, e, i)) + b2i(bit(c.VCO, i))
			c.setAccLo(i, uint16(sum))
			c.V[vd][i] = clampS(sum)
		}
		c.VCO = 0
	case 0x11: // vsub
		for i := uint32(0); i < 8; i++ {
			d := s16(vsv[i]) - s16(vte(vtv, e, i)) - b2i(bit(c.VCO, i))
			c.setAccLo(i, uint16(d))
			c.V[vd][i] = clampS(d)
		}
		c.VCO = 0
	case 0x13: // vabs: the sign of vs applied to vt
		for i := uint32(0); i < 8; i++ {
			s := s16(vsv[i])
			t := vte(vtv, e, i)
			var r, a uint16
			switch {
			case s < 0:
				// Negating the most negative value saturates in the register,
				// but the accumulator keeps the raw two's complement — the
				// saturation happens on the way out, not inside the adder.
				a = uint16(-int16(t))
				r = a
				if t == 0x8000 {
					r = 0x7FFF
				}
			case s > 0:
				r, a = t, t
			}
			c.setAccLo(i, a)
			c.V[vd][i] = r
		}
	case 0x14: // vaddc: unsigned add, leaving the carry in VCO
		c.VCO = 0
		for i := uint32(0); i < 8; i++ {
			sum := u16(vsv[i]) + u16(vte(vtv, e, i))
			c.setAccLo(i, uint16(sum))
			c.V[vd][i] = uint16(sum)
			setBit(&c.VCO, i, sum > 0xFFFF)
		}
	case 0x15: // vsubc: unsigned subtract, leaving borrow and not-equal in VCO
		c.VCO = 0
		for i := uint32(0); i < 8; i++ {
			d := u16(vsv[i]) - u16(vte(vtv, e, i))
			c.setAccLo(i, uint16(d))
			c.V[vd][i] = uint16(d)
			setBit(&c.VCO, i, d < 0)
			setBit(&c.VCO, i+8, d != 0)
		}

	case 0x1D: // vsar: read a slice of the accumulator into vd
		for i := uint32(0); i < 8; i++ {
			switch e {
			case 8:
				c.V[vd][i] = uint16(c.Acc[i] >> 32)
			case 9:
				c.V[vd][i] = uint16(c.Acc[i] >> 16)
			case 10:
				c.V[vd][i] = uint16(c.Acc[i])
			default:
				c.V[vd][i] = 0
			}
		}

	// --- compares: write VCC, and select into vd -----------------------------
	case 0x20, 0x21, 0x22, 0x23:
		c.compare(funct, e, vsv, vtv, vd)
	case 0x24:
		c.vcl(e, vsv, vtv, vd)
	case 0x25:
		c.vch(e, vsv, vtv, vd)
	case 0x26:
		c.vcr(e, vsv, vtv, vd)
	case 0x27: // vmrg: select by VCC, clearing the carries but not VCC itself
		for i := uint32(0); i < 8; i++ {
			r := vte(vtv, e, i)
			if bit(c.VCC, i) {
				r = vsv[i]
			}
			c.setAccLo(i, r)
			c.V[vd][i] = r
		}
		c.VCO = 0

	// --- bitwise -------------------------------------------------------------
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D:
		for i := uint32(0); i < 8; i++ {
			a, b := vsv[i], vte(vtv, e, i)
			var r uint16
			switch funct {
			case 0x28:
				r = a & b
			case 0x29:
				r = ^(a & b)
			case 0x2A:
				r = a | b
			case 0x2B:
				r = ^(a | b)
			case 0x2C:
				r = a ^ b
			case 0x2D:
				r = ^(a ^ b)
			}
			c.setAccLo(i, r)
			c.V[vd][i] = r
		}

	case 0x33: // vmov: copy one lane
		// The single-lane operations name their destination lane in the low bits
		// of the vs field. VMOV's source goes through the ordinary element
		// broadcast, indexed at the destination lane; the divide family reads the
		// raw lane e&7 instead. n64-systemtest distinguishes the two.
		de := vs & 7
		c.V[vd][de] = vte(vtv, e, de)
		for i := uint32(0); i < 8; i++ {
			c.setAccLo(i, vte(vtv, e, i))
		}
	case 0x37: // vnop

	// The holes in the opcode map are not no-ops. The hardware's adder still
	// runs: the accumulator's low slice receives vs + vt (wrapping), and the
	// destination register is zeroed. n64-systemtest probes every one of these
	// encodings against hardware; an emulator that treats them as no-ops or
	// faults is detectable.
	case 0x12, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1E, 0x1F,
		0x2E, 0x2F, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x3E:
		for i := uint32(0); i < 8; i++ {
			c.setAccLo(i, vsv[i]+vte(vtv, e, i))
			c.V[vd][i] = 0
		}
	case 0x3F: // vnull: does nothing at all

	case 0x30: // vrcp: the reciprocal of a 16-bit input
		c.divIn, c.divInLoaded = 0, false
		c.divide(e, vtv, vs, vd, c.reciprocal(int32(int16(vtv[e&7]))))
	case 0x31: // vrcpl: the low half of a 32-bit reciprocal
		c.divide(e, vtv, vs, vd, c.reciprocal(c.divInput(vtv[e&7])))
	case 0x34: // vrsq: the reciprocal square root of a 16-bit input
		c.divIn, c.divInLoaded = 0, false
		c.divide(e, vtv, vs, vd, c.rsqrt(int32(int16(vtv[e&7]))))
	case 0x35: // vrsql: the low half of a 32-bit reciprocal square root
		c.divide(e, vtv, vs, vd, c.rsqrt(c.divInput(vtv[e&7])))
	case 0x32, 0x36: // vrcph, vrsqh: exchange the high halves
		// No arithmetic happens here. The previous result's high half is read
		// out, and this operand's becomes the high half of the next divide.
		c.V[vd][vs&7] = c.divOut
		c.divIn, c.divInLoaded = vtv[e&7], true
		for i := uint32(0); i < 8; i++ {
			c.setAccLo(i, vte(vtv, e, i))
		}

	default:
		c.unimpl(w)
	}
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// compare implements VLT, VEQ, VNE and VGE, which differ only in the predicate
// and in how they consult the carry and not-equal flags VSUBC left behind.
func (c *CPU) compare(funct, e uint32, vsv, vtv [8]uint16, vd uint32) {
	c.VCC = 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(vsv[i]), s16(vte(vtv, e, i))
		carry, ne := bit(c.VCO, i), bit(c.VCO, i+8)

		var cond bool
		switch funct {
		case 0x20: // vlt
			cond = a < b || (a == b && carry && ne)
		case 0x21: // veq
			cond = a == b && !ne
		case 0x22: // vne
			cond = a != b || ne
		case 0x23: // vge
			cond = a > b || (a == b && !(carry && ne))
		}
		setBit(&c.VCC, i, cond)

		r := uint16(b)
		if cond {
			r = uint16(a)
		}
		c.setAccLo(i, r)
		c.V[vd][i] = r
	}
	c.VCO = 0
}

// vch is the clip test used when a triangle straddles a frustum plane: it
// compares vs against ±vt, choosing the sign from whether they point the same
// way, and records both the result and enough state for VCL to finish the job.
//
// Every lane writes all five flag bits, so the flags can be cleared up front and
// set where true. The selection bits VCC carries out are what the microcode's
// clipper steers by; n64-systemtest checks each against hardware, per lane, for
// every prior flag state.
func (c *CPU) vch(e uint32, vsv, vtv [8]uint16, vd uint32) {
	c.VCO, c.VCC, c.VCE = 0, 0, 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(vsv[i]), s16(vte(vtv, e, i))

		var r int64
		if (a ^ b) < 0 { // opposite signs: compare against -vt, the lower plane
			sum := a + b
			le := sum <= 0
			setBit(&c.VCC, i, le)
			setBit(&c.VCC, i+8, b < 0)
			if le {
				r = -b
			} else {
				r = a
			}
			setBit(&c.VCO, i, true)
			setBit(&c.VCO, i+8, sum != 0 && uint16(a) != ^uint16(b))
			if sum == -1 {
				c.VCE |= 1 << i
			}
		} else { // same signs: compare against +vt, the upper plane
			ge := a-b >= 0
			setBit(&c.VCC, i, b < 0)
			setBit(&c.VCC, i+8, ge)
			if ge {
				r = b
			} else {
				r = a
			}
			setBit(&c.VCO, i+8, a-b != 0)
		}
		c.setAccLo(i, uint16(r))
		c.V[vd][i] = uint16(r)
	}
}

// vcl completes the clip test VCH began, using the flags it left behind. The
// VCC bit for a lane is rewritten only when VCH found the operands equal enough
// (its not-equal bit clear); otherwise the earlier verdict stands and only the
// selection happens.
func (c *CPU) vcl(e uint32, vsv, vtv [8]uint16, vd uint32) {
	for i := uint32(0); i < 8; i++ {
		a, b := vsv[i], vte(vtv, e, i)
		carry, ne := bit(c.VCO, i), bit(c.VCO, i+8)
		vce := c.VCE&(1<<i) != 0

		var r uint16
		if carry { // opposite signs: test against the lower plane
			if !ne {
				sum := uint32(a) + uint32(b)
				exact := sum&0xFFFF == 0    // -vs == vt precisely
				noCarry := sum&0x10000 == 0 // the 16-bit add did not wrap
				le := exact && noCarry
				if vce {
					le = exact || noCarry
				}
				setBit(&c.VCC, i, le)
			}
			if bit(c.VCC, i) {
				r = uint16(-int16(b))
			} else {
				r = a
			}
		} else { // same signs: test against the upper plane
			if !ne {
				setBit(&c.VCC, i+8, a >= b)
			}
			if bit(c.VCC, i+8) {
				r = b
			} else {
				r = a
			}
		}
		c.setAccLo(i, r)
		c.V[vd][i] = r
	}
	c.VCO, c.VCE = 0, 0
}

// vcr is the clip test for single-precision inputs: like VCH but with a
// one's-complement select on the lower plane and no carry bookkeeping for a
// following VCL.
func (c *CPU) vcr(e uint32, vsv, vtv [8]uint16, vd uint32) {
	c.VCO, c.VCC, c.VCE = 0, 0, 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(vsv[i]), s16(vte(vtv, e, i))

		var r int64
		if (a ^ b) < 0 { // opposite signs
			le := a+b < 0
			setBit(&c.VCC, i, le)
			setBit(&c.VCC, i+8, b < 0)
			if le {
				r = ^b
			} else {
				r = a
			}
		} else { // same signs
			ge := a-b >= 0
			setBit(&c.VCC, i, b < 0)
			setBit(&c.VCC, i+8, ge)
			if ge {
				r = b
			} else {
				r = a
			}
		}
		c.setAccLo(i, uint16(r))
		c.V[vd][i] = uint16(r)
	}
}

// --- the element-addressed load/store family --------------------------------

// vecLoad implements LWC2. The offset is a 7-bit signed value scaled by the
// access width, and the element specifier selects where in the register the
// bytes land.
func (c *CPU) vecLoad(w, rs, vt uint32) {
	op := (w >> 11) & 31
	e := (w >> 7) & 15
	off := int32(w&0x7F) << 25 >> 25
	addr := c.R[rs] + uint32(off<<memShift[op])

	switch op {
	case 0x00, 0x01, 0x02, 0x03: // lbv, lsv, llv, ldv: 1, 2, 4 or 8 bytes
		n := uint32(1) << op
		for i := uint32(0); i < n && e+i < 16; i++ {
			c.setVecByte(vt, e+i, c.DMEM[(addr+i)&0xFFF])
		}
	case 0x04: // lqv: up to the next 16-byte boundary
		end := (addr &^ 15) + 16
		for b := e; b < 16 && addr < end; b, addr = b+1, addr+1 {
			c.setVecByte(vt, b, c.DMEM[addr&0xFFF])
		}
	case 0x05: // lrv: the bytes before the address, right-aligned in the register
		// LQV and LRV are the LWL and LWR of a 128-bit load. LQV takes the bytes
		// from the address up to the next 16-byte boundary; LRV takes the bytes
		// from the previous boundary up to, but excluding, the address — and
		// lands them at the register's tail.
		//
		// The element is where the pair's first byte goes, so a non-zero element
		// shortens the *load*: LRV fills from 16-n+e to 15, reading from the
		// boundary. Fewer bytes are read, not different ones.
		n := addr & 15
		base := addr &^ 15
		for i := uint32(0); 16-n+e+i < 16; i++ {
			c.setVecByte(vt, 16-n+e+i, c.DMEM[(base+i)&0xFFF])
		}
	case 0x06: // lpv: eight bytes, one per lane, as signed values
		c.packedLoad(vt, e, addr, 8, 1)
	case 0x07: // luv: eight bytes, one per lane, as unsigned values
		c.packedLoad(vt, e, addr, 7, 1)
	case 0x08: // lhv: every second byte, one per lane, unsigned
		c.packedLoad(vt, e, addr, 7, 2)
	case 0x09: // lfv: every fourth byte, into one half of the register
		c.lfv(vt, e, addr)
	case 0x0A: // lwv: does not exist in the hardware; the encoding does nothing
	case 0x0B: // ltv: one lane into each of eight registers — a transposed load
		c.ltv(vt, e, addr)
	default:
		c.unimpl(w)
	}
}

// packedLoad implements LPV, LUV and LHV: eight bytes unpacked one per lane,
// every stride-th byte from a 16-byte window at the address's 8-byte boundary.
//
// The element does not choose destination lanes — all eight are always written.
// It rotates the *source*: lane i reads the byte at (16 - e + i*stride + mis)
// mod 16 within the window, mis being the address's misalignment. n64-systemtest
// walks every element against every misalignment, including the DMEM wrap.
//
// The shift decides how the byte sits in its 16-bit lane: LPV puts it in bits
// 15..8, which makes it a signed value; LUV and LHV in bits 14..7, unsigned.
func (c *CPU) packedLoad(vt, e, addr, shift, stride uint32) {
	base, mis := addr&^7, addr&7
	for i := uint32(0); i < 8; i++ {
		b := c.DMEM[(base+(16-e+i*stride+mis)&0xF)&0xFFF]
		c.V[vt][i] = uint16(b) << shift
	}
}

// lfv loads every fourth byte through a fixed source pattern into one half of
// the register — bytes e onward, at most eight. The pattern is the hardware's
// own, transcribed from n64-systemtest's LFV expectation, quirks included: lane
// 0 offsets by +e where every other lane offsets by -e, and lanes 4..7 repeat
// sources rather than continue them.
func (c *CPU) lfv(vt, e, addr uint32) {
	base, mis := addr&^7, addr&7
	off := [8]uint32{mis + e, mis + 4 - e, mis + 8 - e, mis + 12 - e,
		mis + 8 - e, mis + 12 - e, mis - e, mis + 4 - e}
	var tmp [16]byte
	for i := uint32(0); i < 8; i++ {
		v := uint16(c.DMEM[(base+off[i]&0xF)&0xFFF]) << 7
		tmp[i*2] = byte(v >> 8)
		tmp[i*2+1] = byte(v)
	}
	for b := e; b < 16 && b < e+8; b++ {
		c.setVecByte(vt, b, tmp[b])
	}
}

// ltv is the transposed load: byte pair i lands in register group-base+i, each
// pair one lane further along, so eight registers each receive one lane of a
// diagonal. The group is the register's aligned block of eight; the element's
// top three bits pick which diagonal, and bit 3 of the address rotates the
// 16-byte source window by half.
func (c *CPU) ltv(vt, e, addr uint32) {
	base := addr &^ 7
	rot := base & 8
	for i := uint32(0); i < 8; i++ {
		reg := vt&^7 + (e/2+i)&7
		for h := uint32(0); h < 2; h++ {
			c.setVecByte(reg, i*2+h, c.DMEM[(base+(rot+e+i*2+h)&0xF)&0xFFF])
		}
	}
}

// vecStore implements SWC2.
func (c *CPU) vecStore(w, rs, vt uint32) {
	op := (w >> 11) & 31
	e := (w >> 7) & 15
	off := int32(w&0x7F) << 25 >> 25
	addr := c.R[rs] + uint32(off<<memShift[op])

	switch op {
	case 0x00, 0x01, 0x02, 0x03: // sbv, ssv, slv, sdv
		n := uint32(1) << op
		for i := uint32(0); i < n; i++ {
			c.DMEM[(addr+i)&0xFFF] = c.vecByte(vt, (e+i)&15)
		}
	case 0x04: // sqv
		end := (addr &^ 15) + 16
		for b := e; addr < end; b, addr = b+1, addr+1 {
			c.DMEM[addr&0xFFF] = c.vecByte(vt, b&15)
		}
	case 0x05: // srv
		// The stores differ from the loads in how the element is used: they
		// always write full width, fetching from the register with a wrap, so a
		// non-zero element rotates the data rather than shortening it.
		n := addr & 15
		base := addr &^ 15
		for i := uint32(0); i < n; i++ {
			c.DMEM[(base+i)&0xFFF] = c.vecByte(vt, (e+16-n+i)&15)
		}
	case 0x06: // spv: eight lanes packed to bytes, signed scale
		c.packedStore(vt, e, addr, 8, 7)
	case 0x07: // suv: as spv, unsigned scale
		c.packedStore(vt, e, addr, 7, 8)
	case 0x08: // shv: every second byte, reading lanes at byte granularity
		c.shv(vt, e, addr)
	case 0x09: // sfv: every fourth byte, through a fixed source-lane table
		c.sfv(vt, e, addr)
	case 0x0A: // swv: the whole register, rotated through a 16-byte window
		base, mis := addr&^7, addr&7
		for i := uint32(0); i < 16; i++ {
			c.DMEM[(base+(mis+i)&0xF)&0xFFF] = c.vecByte(vt, (e+i)&15)
		}
	case 0x0B: // stv: the transposed store, one lane from each of eight registers
		c.stv(vt, e, addr)
	default:
		c.unimpl(w)
	}
}

// packedStore implements SPV and SUV: eight lanes packed to eight contiguous
// bytes. The two differ only in scale — SPV keeps bits 15..8 (signed bytes),
// SUV bits 14..7 (unsigned) — and walking the element past 7 switches a lane to
// the *other* instruction's scale. The swap is decided per element index as the
// walk crosses 8, not wholesale by the starting element.
func (c *CPU) packedStore(vt, e, addr, shift, altShift uint32) {
	for i := uint32(0); i < 8; i++ {
		ei := e + i
		s := shift
		if ei&8 != 0 {
			s = altShift
		}
		c.DMEM[(addr+i)&0xFFF] = byte(c.V[vt][ei&7] >> s)
	}
}

// shv packs eight 16-bit values read at *byte* offsets — element e gives the
// first byte, so an odd element reads values straddling two lanes — into every
// second byte of a 16-byte window at the address's 8-byte boundary.
func (c *CPU) shv(vt, e, addr uint32) {
	base, mis := addr&^7, addr&7
	for i := uint32(0); i < 8; i++ {
		ei := e + i*2
		v := uint16(c.vecByte(vt, ei&15))<<8 | uint16(c.vecByte(vt, (ei+1)&15))
		c.DMEM[(base+(mis+i*2)&0xF)&0xFFF] = byte(v >> 7)
	}
}

// sfvStart gives SFV's first source lane for the elements the hardware defines;
// the three following lanes stay within the same register half. Undefined
// elements store zeroes. Transcribed from n64-systemtest's SFV expectation.
var sfvStart = map[uint32]uint32{0: 0, 1: 6, 4: 1, 5: 7, 8: 4, 11: 3, 12: 5, 15: 0}

// sfv packs four lanes into every fourth byte of a 16-byte window.
func (c *CPU) sfv(vt, e, addr uint32) {
	base, mis := addr&^7, addr&7
	start, ok := sfvStart[e]
	for i := uint32(0); i < 4; i++ {
		var b byte
		if ok {
			lane := start&4 | (start+i)&3
			b = byte(c.V[vt][lane] >> 7)
		}
		c.DMEM[(base+(mis+i*4)&0xF)&0xFFF] = b
	}
}

// stv is the transposed store, LTV's counterpart: byte pair i comes from
// register group-base+i's lane i rotated by the element, writing a full
// 16-byte window. Bit 3 of the address rotates which register the diagonal
// starts in, mirroring LTV's source rotation.
func (c *CPU) stv(vt, e, addr uint32) {
	base := addr &^ 7
	rot := (base & 8) / 2
	for i := uint32(0); i < 16; i++ {
		reg := vt&^7 + (i/2-rot+e/2)&7
		c.DMEM[(base+(addr+i)&0xF)&0xFFF] = c.vecByte(reg, (i+base)&15)
	}
}

// --- the reciprocal unit ----------------------------------------------------
//
// VRCP and VRSQ turn a fixed-point value into its reciprocal (or reciprocal
// square root) by normalising it, looking the mantissa up in a ROM (rcprom.go),
// and shifting the 17-bit result — whose top bit is an implicit 1 — back into
// place. The result is 32 bits wide: the low half lands in the destination lane,
// the high half in an internal register that the next VRCPH/VRSQH reads out.
//
// A 32-bit divide is therefore three instructions. VRCPH latches the high half
// of the input, VRCPL supplies the low half and does the work, and a second
// VRCPH collects the high half of the answer. That is the sequence at the top of
// this game's perspective-divide overlay.

// divide computes a reciprocal, writes its low half to the destination lane, and
// leaves the high half in divOut for the following VRCPH to collect.
func (c *CPU) divide(e uint32, vtv [8]uint16, vs, vd uint32, result uint32) {
	c.V[vd][vs&7] = uint16(result)
	for i := uint32(0); i < 8; i++ {
		c.setAccLo(i, vte(vtv, e, i))
	}
}

// divInput assembles the input for the L-form divides: the operand supplies the
// low half, and a preceding VRCPH/VRSQH may have latched the high half. Without
// one, the input is the operand sign-extended. The latch is consumed either way.
func (c *CPU) divInput(lo uint16) int32 {
	in := int32(int16(lo))
	if c.divInLoaded {
		in = int32(uint32(c.divIn)<<16 | uint32(lo))
	}
	c.divIn, c.divInLoaded = 0, false
	return in
}

// reciprocal normalises the input, looks its top nine mantissa bits up in the
// ROM, prepends the implicit leading 1, and shifts the resulting 17-bit value
// down by the input's exponent. A negative input inverts the result rather than
// negating it, so the sign falls out of the same complement that produced the
// magnitude.
//
// The scaling is the hardware's own and looks arbitrary out of context: the
// result approximates 2^31/x rather than 1/x, and microcode carries the exponent
// itself. What matters is that it is consistent.
//
// Two details are worth stating because getting either wrong yields geometry
// that is almost right. The first is the shift. Published pseudocode for this
// unit writes the output scale as 32 - scale_in, which places the 17-bit value
// two bits too high; the true fixed point puts its top bit at bit 30, so the
// shift is 14 - scale_in. n64-systemtest pins this from both ends: its RCP table
// test asserts reciprocal(0x1000) == 0x0007FFFC, and its exhaustive 16-bit test
// asserts the low half of reciprocal(1) is 0xC000. Only 14 satisfies both.
//
// The second is that the input's magnitude is taken as ~in, not -in, once it
// reaches -32768 or below. Above that the two agree; at exactly -32768 the
// hardware short-circuits, because negating it overflows.
func (c *CPU) reciprocal(in int32) uint32 {
	// The sign, smeared across the word: 0 for a positive input, ~0 for a
	// negative one. It both un-negates the input and re-negates the result.
	mask := in >> 31
	x := uint32(in ^ mask)
	if in > -32768 {
		x = uint32((in ^ mask) - mask)
	}

	switch {
	case x == 0:
		// There is no divide-by-zero exception on the RSP; the result saturates.
		c.divOut = 0x7FFF
		return 0x7FFFFFFF
	case in == -32768:
		c.divOut = 0xFFFF
		return 0xFFFF0000
	}

	// scaleIn is the position of the highest set bit: the input's exponent.
	scaleIn := uint32(31)
	for x&(1<<scaleIn) == 0 {
		scaleIn--
	}

	// The nine mantissa bits immediately below the leading 1. Bit positions
	// below zero read as zero, which the shift direction handles.
	var idx uint32
	if scaleIn >= 9 {
		idx = x >> (scaleIn - 9) & 0x1FF
	} else {
		idx = x << (9 - scaleIn) & 0x1FF
	}

	// Place the 17-bit value (1 || rom) with its top bit at bit 30 - scaleIn.
	v := uint64(0x10000) | uint64(rcpROM[idx])
	var result uint32
	if scaleIn <= 14 {
		result = uint32(v << (14 - scaleIn))
	} else {
		result = uint32(v >> (scaleIn - 14))
	}
	result ^= uint32(mask)

	c.divOut = uint16(result >> 16)
	return result
}

// rsqrt is the reciprocal square root, sharing reciprocal's shape with two
// differences that are the square root's own. The table index carries eight
// mantissa bits plus the exponent's parity — sqrt(2·x) and sqrt(x) need
// different mantissa curves, so odd exponents select the table's upper half.
// And the output shift halves the exponent, since sqrt halves magnitudes'
// scales. The special cases match reciprocal's, and n64-systemtest's RSQ table
// test pins every entry against hardware.
func (c *CPU) rsqrt(in int32) uint32 {
	mask := in >> 31
	x := uint32(in ^ mask)
	if in > -32768 {
		x = uint32((in ^ mask) - mask)
	}

	switch {
	case x == 0:
		c.divOut = 0x7FFF
		return 0x7FFFFFFF
	case in == -32768:
		c.divOut = 0xFFFF
		return 0xFFFF0000
	}

	scaleIn := uint32(31)
	for x&(1<<scaleIn) == 0 {
		scaleIn--
	}

	// Eight mantissa bits below the leading 1, with the exponent's parity above
	// them selecting the table half.
	var idx uint32
	if scaleIn >= 8 {
		idx = x >> (scaleIn - 8) & 0xFF
	} else {
		idx = x << (8 - scaleIn) & 0xFF
	}
	idx |= (scaleIn & 1) << 8

	// Place the 17-bit value with its top bit at bit 30 - scaleIn/2.
	v := uint64(0x10000) | uint64(rsqROM[idx])
	result := uint32(v << 14 >> (scaleIn >> 1))
	result ^= uint32(mask)

	c.divOut = uint16(result >> 16)
	return result
}
