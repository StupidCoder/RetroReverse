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

// vt reads the second operand's lane i under the element specifier.
func (c *CPU) vt(r, e, i uint32) uint16 { return c.V[r][element(e, i)] }

// acc reads a lane of the accumulator, sign-extended from 48 bits.
func (c *CPU) acc(i uint32) int64 { return int64(c.Acc[i]<<16) >> 16 }

func (c *CPU) setAcc(i uint32, v int64) { c.Acc[i] = uint64(v) & accMask }

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

// clampU saturates to an unsigned 16-bit range, which VMULU and VMACU use.
func clampU(v int64) uint16 {
	if v < 0 {
		return 0
	}
	if v > 65535 {
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
			c.Halt("unimplemented cop2 rs=0x%02X (word 0x%08X) at $%03X", rs, w, c.curPC)
		}
		return
	}
	c.vectorALU(w&0x3F, rs&15, rt, rd, shamt)
}

func (c *CPU) ctrl(i uint32) uint16 {
	switch i {
	case 0:
		return c.VCO
	case 1:
		return c.VCC
	case 2:
		return uint16(c.VCE)
	}
	return 0
}

func (c *CPU) setCtrl(i uint32, v uint16) {
	switch i {
	case 0:
		c.VCO = v
	case 1:
		c.VCC = v
	case 2:
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
func (c *CPU) vectorALU(funct, e, vt, vs, vd uint32) {
	switch funct {
	// --- multiply: write the accumulator outright ---------------------------
	case 0x00: // vmulf: signed fractional multiply, rounded
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, p<<1+0x8000)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x01: // vmulu: as vmulf, saturating unsigned
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, p<<1+0x8000)
			c.V[vd][i] = clampU(c.acc(i) >> 16)
		}
	case 0x04: // vmudl: unsigned * unsigned, low slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, u16(c.V[vs][i])*u16(c.vt(vt, e, i))>>16)
			c.V[vd][i] = uint16(c.acc(i))
		}
	case 0x05: // vmudm: signed * unsigned, high slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, s16(c.V[vs][i])*u16(c.vt(vt, e, i)))
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x06: // vmudn: unsigned * signed, low slice
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, u16(c.V[vs][i])*s16(c.vt(vt, e, i)))
			c.V[vd][i] = uint16(c.acc(i))
		}
	case 0x07: // vmudh: signed * signed, shifted into the high slice
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, p<<16)
			c.V[vd][i] = clampS(p)
		}

	// --- multiply-accumulate: add into the accumulator -----------------------
	case 0x08: // vmacf
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, c.acc(i)+p<<1)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x09: // vmacu
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, c.acc(i)+p<<1)
			c.V[vd][i] = clampU(c.acc(i) >> 16)
		}
	case 0x0C: // vmadl
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+u16(c.V[vs][i])*u16(c.vt(vt, e, i))>>16)
			c.V[vd][i] = clampLow(c.acc(i))
		}
	case 0x0D: // vmadm
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+s16(c.V[vs][i])*u16(c.vt(vt, e, i)))
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}
	case 0x0E: // vmadn
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, c.acc(i)+u16(c.V[vs][i])*s16(c.vt(vt, e, i)))
			c.V[vd][i] = clampLow(c.acc(i))
		}
	case 0x0F: // vmadh
		for i := uint32(0); i < 8; i++ {
			p := s16(c.V[vs][i]) * s16(c.vt(vt, e, i))
			c.setAcc(i, c.acc(i)+p<<16)
			c.V[vd][i] = clampS(c.acc(i) >> 16)
		}

	// --- add and subtract, through the carry flags ---------------------------
	case 0x10: // vadd: adds the carry VADDC left behind
		for i := uint32(0); i < 8; i++ {
			sum := s16(c.V[vs][i]) + s16(c.vt(vt, e, i)) + b2i(bit(c.VCO, i))
			c.setAcc(i, sum)
			c.V[vd][i] = clampS(sum)
		}
		c.VCO = 0
	case 0x11: // vsub
		for i := uint32(0); i < 8; i++ {
			d := s16(c.V[vs][i]) - s16(c.vt(vt, e, i)) - b2i(bit(c.VCO, i))
			c.setAcc(i, d)
			c.V[vd][i] = clampS(d)
		}
		c.VCO = 0
	case 0x13: // vabs: the sign of vs applied to vt
		for i := uint32(0); i < 8; i++ {
			s := s16(c.V[vs][i])
			t := c.vt(vt, e, i)
			var r uint16
			switch {
			case s < 0:
				if t == 0x8000 {
					r = 0x7FFF // negating the most negative value saturates
				} else {
					r = uint16(-int16(t))
				}
			case s > 0:
				r = t
			}
			c.setAcc(i, s16(r))
			c.V[vd][i] = r
		}
	case 0x14: // vaddc: unsigned add, leaving the carry in VCO
		c.VCO = 0
		for i := uint32(0); i < 8; i++ {
			sum := u16(c.V[vs][i]) + u16(c.vt(vt, e, i))
			c.setAcc(i, sum)
			c.V[vd][i] = uint16(sum)
			setBit(&c.VCO, i, sum > 0xFFFF)
		}
	case 0x15: // vsubc: unsigned subtract, leaving borrow and not-equal in VCO
		c.VCO = 0
		for i := uint32(0); i < 8; i++ {
			d := u16(c.V[vs][i]) - u16(c.vt(vt, e, i))
			c.setAcc(i, d)
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
		c.compare(funct, e, vt, vs, vd)
	case 0x24:
		c.vcl(e, vt, vs, vd)
	case 0x25:
		c.vch(e, vt, vs, vd)
	case 0x26:
		c.vcr(e, vt, vs, vd)
	case 0x27: // vmrg: select by VCC, without touching the flags
		for i := uint32(0); i < 8; i++ {
			r := c.vt(vt, e, i)
			if bit(c.VCC, i) {
				r = c.V[vs][i]
			}
			c.setAcc(i, s16(r))
			c.V[vd][i] = r
		}
		c.VCO = 0

	// --- bitwise -------------------------------------------------------------
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D:
		for i := uint32(0); i < 8; i++ {
			a, b := c.V[vs][i], c.vt(vt, e, i)
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
			c.setAcc(i, s16(r))
			c.V[vd][i] = r
		}

	case 0x33: // vmov: copy one lane
		// The single-lane operations name their destination lane in the low bits
		// of the vs field and their source lane in the element field, not by the
		// broadcast rule the ALU ops use.
		de, se := vs&7, e&7
		c.V[vd][de] = c.V[vt][se]
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, s16(c.vt(vt, e, i)))
		}
	case 0x37: // vnop

	case 0x30: // vrcp: the reciprocal of a 16-bit input
		c.divide(e, vt, vs, vd, int32(int16(c.V[vt][e&7])))
	case 0x31: // vrcpl: the low half of a 32-bit reciprocal
		lo := c.V[vt][e&7]
		in := int32(int16(lo)) // with no high half latched, the input sign-extends
		if c.divInLoaded {
			in = int32(uint32(c.divIn)<<16 | uint32(lo))
		}
		c.divide(e, vt, vs, vd, in)
		c.divIn, c.divInLoaded = 0, false
	case 0x32, 0x36: // vrcph, vrsqh: exchange the high halves
		// No arithmetic happens here. The previous result's high half is read
		// out, and this operand's becomes the high half of the next divide.
		c.V[vd][vs&7] = c.divOut
		c.divIn, c.divInLoaded = c.V[vt][e&7], true
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, s16(c.vt(vt, e, i)))
		}
	case 0x34, 0x35: // vrsq, vrsql
		// No microcode seen so far executes these, and the published pseudocode
		// for them is self-inconsistent: it derives scale_out from the input's
		// scale before halving that scale, and its final line invokes the
		// reciprocal rather than the reciprocal square root. Implementing it from
		// a description that cannot be right would produce a wrong answer that
		// looks plausible, so it waits for a microcode that exercises it.
		c.Halt("unmodelled %q at $%03X: the published algorithm for the reciprocal "+
			"square root is self-inconsistent and no microcode has exercised it",
			vuOp[funct], c.curPC)

	default:
		c.Halt("unimplemented vector op funct 0x%02X at $%03X", funct, c.curPC)
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
func (c *CPU) compare(funct, e, vt, vs, vd uint32) {
	c.VCC = 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(c.V[vs][i]), s16(c.vt(vt, e, i))
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
		c.setAcc(i, s16(r))
		c.V[vd][i] = r
	}
	c.VCO = 0
}

// vch is the clip test used when a triangle straddles a frustum plane: it
// compares vs against ±vt, choosing the sign from whether they point the same
// way, and records both the result and enough state for VCL to finish the job.
func (c *CPU) vch(e, vt, vs, vd uint32) {
	c.VCO, c.VCC, c.VCE = 0, 0, 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(c.V[vs][i]), s16(c.vt(vt, e, i))
		sign := (a ^ b) < 0 // the operands have opposite signs

		var r int64
		if sign {
			// Compare against -vt: the lower clip plane.
			ge := b >= 0
			le := a+b <= 0
			setBit(&c.VCC, i, le)
			setBit(&c.VCC, i+8, ge)
			if le {
				r = -b
			} else {
				r = a
			}
			setBit(&c.VCO, i, true)
			setBit(&c.VCO, i+8, a+b != 0 && uint16(a) != ^uint16(b))
			if a+b == -1 {
				c.VCE |= 1 << i
			}
		} else {
			ge := a-b >= 0
			le := b < 0
			setBit(&c.VCC, i, le)
			setBit(&c.VCC, i+8, ge)
			if ge {
				r = b
			} else {
				r = a
			}
			setBit(&c.VCO, i+8, a-b != 0)
		}
		c.setAcc(i, r)
		c.V[vd][i] = uint16(r)
	}
}

// vcl completes the clip test VCH began, using the flags it left behind.
func (c *CPU) vcl(e, vt, vs, vd uint32) {
	for i := uint32(0); i < 8; i++ {
		a, b := c.V[vs][i], c.vt(vt, e, i)
		carry, ne := bit(c.VCO, i), bit(c.VCO, i+8)
		vce := c.VCE&(1<<i) != 0

		var r uint16
		if carry { // opposite signs: test against the lower plane
			if !ne {
				sum := uint32(a) + uint32(b)
				le := sum&0x10000 == 0 || sum&0xFFFF == 0
				if vce {
					le = sum&0xFFFF <= 0x10000
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
				setBit(&c.VCC, i+8, uint32(a)-uint32(b) >= 0 && a >= b)
			}
			if bit(c.VCC, i+8) {
				r = b
			} else {
				r = a
			}
		}
		c.setAcc(i, s16(r))
		c.V[vd][i] = r
	}
	c.VCO, c.VCE = 0, 0
}

// vcr is the clip test without the carry bookkeeping, used for the simpler
// single-sided planes.
func (c *CPU) vcr(e, vt, vs, vd uint32) {
	c.VCO, c.VCC, c.VCE = 0, 0, 0
	for i := uint32(0); i < 8; i++ {
		a, b := s16(c.V[vs][i]), s16(c.vt(vt, e, i))
		sign := (a ^ b) < 0

		var r int64
		if sign {
			le := a+b < 0
			setBit(&c.VCC, i, le)
			if le {
				r = ^b
			} else {
				r = a
			}
		} else {
			ge := a-b >= 0
			setBit(&c.VCC, i+8, ge)
			if ge {
				r = b
			} else {
				r = a
			}
		}
		c.setAcc(i, r)
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
	case 0x05: // lrv: the bytes before the boundary, right-aligned
		n := addr & 15
		addr &^= 15
		for b := 16 - n; b < 16; b, addr = b+1, addr+1 {
			c.setVecByte(vt, b, c.DMEM[addr&0xFFF])
		}
	case 0x06: // lpv: eight bytes, one per lane, as signed values
		c.packedLoad(vt, e, addr, 8)
	case 0x07: // luv: eight bytes, one per lane, as unsigned values
		c.packedLoad(vt, e, addr, 7)
	default:
		c.Halt("unmodelled vector load %q at $%03X", lwc2Op[op], c.curPC)
	}
}

// packedLoad implements LPV and LUV: eight bytes, one into each lane.
//
// Two independent wraps make this instruction what it is. The bytes are read
// starting at the effective address and wrap at its 8-byte boundary; the lanes
// they land in start at the element and wrap at eight. So an unaligned address
// and a non-zero element rotate the source and the destination separately.
//
// The shift decides how the byte sits in its 16-bit lane: LPV puts it in bits
// 15..8, which makes it a signed value, and LUV in bits 14..7, which makes it
// unsigned. For loads an element above 7 simply wraps, its top bit ignored.
func (c *CPU) packedLoad(vt, e, addr, shift uint32) {
	base, start := addr&^7, addr&7
	for i := uint32(0); i < 8; i++ {
		b := c.DMEM[(base+(start+i)&7)&0xFFF]
		c.V[vt][(e+i)&7] = uint16(b) << shift
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
		n := addr & 15
		addr &^= 15
		for b := 16 - n; b < 16; b, addr = b+1, addr+1 {
			c.DMEM[addr&0xFFF] = c.vecByte(vt, (e+b)&15)
		}
	case 0x06, 0x07: // spv, suv
		// Unlike the loads, an element above 7 does not merely wrap here: it
		// swaps the two instructions' meaning, so SPV with element 8..15 stores
		// as SUV does, and the other way round.
		shift := uint32(8)
		if op == 0x07 {
			shift = 7
		}
		if e >= 8 {
			shift = 15 - shift
		}
		c.packedStore(vt, e, addr, shift)
	default:
		c.Halt("unmodelled vector store %q at $%03X", swc2Op[op], c.curPC)
	}
}

// packedStore implements SPV and SUV, the counterpart of packedLoad.
func (c *CPU) packedStore(vt, e, addr, shift uint32) {
	base, start := addr&^7, addr&7
	for i := uint32(0); i < 8; i++ {
		v := c.V[vt][(e+i)&7] >> shift
		c.DMEM[(base+(start+i)&7)&0xFFF] = byte(v)
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
func (c *CPU) divide(e, vt, vs, vd uint32, in int32) {
	c.V[vd][vs&7] = uint16(c.reciprocal(in))
	for i := uint32(0); i < 8; i++ {
		c.setAcc(i, s16(c.vt(vt, e, i)))
	}
}

// reciprocal follows the hardware's algorithm exactly: normalise the input, look
// its top nine mantissa bits up in the ROM, prepend the implicit leading 1, and
// shift the resulting 17-bit value so that its top bit lands at the output's
// scale. A negative input inverts the result rather than negating it.
//
// The scaling is the hardware's own and looks arbitrary out of context: the
// result approximates 2^33/x rather than 1/x, and microcode carries the exponent
// itself. What matters is that it is consistent, which the shift below keeps.
func (c *CPU) reciprocal(in int32) uint32 {
	if in == 0 {
		// There is no divide-by-zero exception on the RSP; the result saturates.
		c.divOut = 0xFFFF
		return 0xFFFF
	}
	neg := in < 0
	x := uint32(in)
	if neg {
		x = uint32(-in)
	}

	// scaleIn is the position of the highest set bit: the input's exponent.
	scaleIn := uint32(31)
	for x&(1<<scaleIn) == 0 {
		scaleIn--
	}
	scaleOut := 32 - scaleIn

	// The nine mantissa bits immediately below the leading 1. Bit positions
	// below zero read as zero, which the shift direction handles.
	var idx uint32
	if scaleIn >= 9 {
		idx = x >> (scaleIn - 9) & 0x1FF
	} else {
		idx = x << (9 - scaleIn) & 0x1FF
	}

	// Place the 17-bit value (1 || rom) with its top bit at scaleOut.
	v := uint64(0x10000) | uint64(rcpROM[idx])
	var result uint32
	if scaleOut >= 16 {
		result = uint32(v << (scaleOut - 16))
	} else {
		result = uint32(v >> (16 - scaleOut))
	}
	if neg {
		result = ^result
	}
	c.divOut = uint16(result >> 16)
	return result
}
