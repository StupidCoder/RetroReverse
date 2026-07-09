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
// The reciprocal instructions (VRCP, VRSQ and their long forms) index a table
// burnt into the RSP's silicon. That table is not on the cartridge and has no
// published closed form, so they halt rather than return a plausible-looking
// wrong answer — a perspective divide that is quietly off by a bit would show up
// as geometry that almost works, which is the worst kind of bug to chase.

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
		src := element(e, vd&7)
		c.V[vd][vd&7] = c.V[vt][src]
		for i := uint32(0); i < 8; i++ {
			c.setAcc(i, s16(c.vt(vt, e, i)))
		}
	case 0x37: // vnop

	case 0x30, 0x31, 0x32, 0x34, 0x35, 0x36:
		c.Halt("unmodelled reciprocal instruction %q at $%03X: the RSP's reciprocal ROM is "+
			"burnt into silicon, is not on the cartridge, and has no published closed form",
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
	default:
		c.Halt("unmodelled vector load %q at $%03X", lwc2Op[op], c.curPC)
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
	default:
		c.Halt("unmodelled vector store %q at $%03X", swc2Op[op], c.curPC)
	}
}
