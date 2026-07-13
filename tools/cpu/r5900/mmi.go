package r5900

// mmi.go is the multimedia instruction set: the Emotion Engine's packed-SIMD
// integer unit, reached through primary opcode 0x1C.
//
// It is a whole second instruction space. The primary opcode selects a funct,
// which either names an instruction directly (madd, mult1, the packed shifts) or
// opens one of four sub-tables — MMI0, MMI1, MMI2, MMI3 — indexed by the shamt
// field. Those sub-tables hold the packed arithmetic: the same add, subtract,
// compare, min/max and interleave, spelled once for each of the three element
// widths a 128-bit register can be cut into (four 32-bit words, eight 16-bit
// halfwords, sixteen bytes).
//
// Two pieces of architectural state exist only for this world:
//
//   - A second multiply accumulator. The "1" instructions (mult1, div1, madd1,
//     mfhi1 …) target HI1/LO1 instead of HI/LO, so two multiply chains can be in
//     flight at once. The packed multiplies go further and treat the pair as one
//     128-bit accumulator: LO holds words 0 and 1, LO1 words 2 and 3.
//   - The shift-amount register SA, set by mtsa/mtsab/mtsah and read only by
//     qfsrv, the funnel shift that extracts a quadword straddling two registers.
//
// The decoder and the interpreter live in the same file because the four
// sub-tables are the hard part to get right, and splitting them across files puts
// the encoding and the meaning of an opcode two screens apart.

import "math/bits"

// --- element views ----------------------------------------------------------
//
// A Quad is addressed as four words, eight halfwords or sixteen bytes, element 0
// always being the least significant. Everything below is expressed through these
// four helpers, so an off-by-one in the packing is a single place to fix.

func words(q Quad) [4]uint32 {
	return [4]uint32{uint32(q.Lo), uint32(q.Lo >> 32), uint32(q.Hi), uint32(q.Hi >> 32)}
}

func fromWords(w [4]uint32) Quad {
	return Quad{
		Lo: uint64(w[0]) | uint64(w[1])<<32,
		Hi: uint64(w[2]) | uint64(w[3])<<32,
	}
}

func halves(q Quad) [8]uint16 {
	var h [8]uint16
	for i := 0; i < 4; i++ {
		h[i] = uint16(q.Lo >> (16 * uint(i)))
		h[i+4] = uint16(q.Hi >> (16 * uint(i)))
	}
	return h
}

func fromHalves(h [8]uint16) Quad {
	var q Quad
	for i := 0; i < 4; i++ {
		q.Lo |= uint64(h[i]) << (16 * uint(i))
		q.Hi |= uint64(h[i+4]) << (16 * uint(i))
	}
	return q
}

func octets(q Quad) [16]uint8 {
	var b [16]uint8
	for i := 0; i < 8; i++ {
		b[i] = uint8(q.Lo >> (8 * uint(i)))
		b[i+8] = uint8(q.Hi >> (8 * uint(i)))
	}
	return b
}

func fromOctets(b [16]uint8) Quad {
	var q Quad
	for i := 0; i < 8; i++ {
		q.Lo |= uint64(b[i]) << (8 * uint(i))
		q.Hi |= uint64(b[i+8]) << (8 * uint(i))
	}
	return q
}

// --- saturation -------------------------------------------------------------

func satS32(v int64) uint32 {
	if v > 0x7FFFFFFF {
		return 0x7FFFFFFF
	}
	if v < -0x80000000 {
		return 0x80000000
	}
	return uint32(int32(v))
}

func satS16(v int32) uint16 {
	if v > 0x7FFF {
		return 0x7FFF
	}
	if v < -0x8000 {
		return 0x8000
	}
	return uint16(int16(v))
}

func satS8(v int32) uint8 {
	if v > 0x7F {
		return 0x7F
	}
	if v < -0x80 {
		return 0x80
	}
	return uint8(int8(v))
}

func satU32(v int64) uint32 {
	if v > 0xFFFFFFFF {
		return 0xFFFFFFFF
	}
	if v < 0 {
		return 0
	}
	return uint32(v)
}

func satU16(v int32) uint16 {
	if v > 0xFFFF {
		return 0xFFFF
	}
	if v < 0 {
		return 0
	}
	return uint16(v)
}

func satU8(v int32) uint8 {
	if v > 0xFF {
		return 0xFF
	}
	if v < 0 {
		return 0
	}
	return uint8(v)
}

// --- the 128-bit accumulator view -------------------------------------------
//
// The packed multiplies see HI/HI1 and LO/LO1 as two 128-bit registers, word i of
// which lives in the low half for i < 2 and the "1" half for i >= 2.

func (c *CPU) loQ() Quad { return Quad{Lo: c.LO, Hi: c.LO1} }
func (c *CPU) hiQ() Quad { return Quad{Lo: c.HI, Hi: c.HI1} }

func (c *CPU) setLoQ(q Quad) { c.LO, c.LO1 = q.Lo, q.Hi }
func (c *CPU) setHiQ(q Quad) { c.HI, c.HI1 = q.Lo, q.Hi }

// --- MMI dispatch -----------------------------------------------------------

func (c *CPU) mmi(w, rs, rt, rd, shamt uint32) {
	switch w & 0x3F {
	case 0x00: // madd — accumulate into HI/LO, and write the low half to rd
		c.maddAcc(rd, int64(int32(uint32(c.R[rs].Lo)))*int64(int32(uint32(c.R[rt].Lo))), false)
	case 0x01: // maddu
		c.maddAcc(rd, int64(uint64(uint32(c.R[rs].Lo))*uint64(uint32(c.R[rt].Lo))), false)
	case 0x04: // plzcw — leading redundant sign bits of each of the two low words
		s := words(c.R[rs])
		c.set(rd, uint64(lzcw(s[0]))|uint64(lzcw(s[1]))<<32)
	case 0x08:
		c.mmi0(rs, rt, rd, shamt)
	case 0x09:
		c.mmi2(rs, rt, rd, shamt)
	case 0x10: // mfhi1
		c.set(rd, c.HI1)
	case 0x11: // mthi1
		c.HI1 = c.R[rs].Lo
	case 0x12: // mflo1
		c.set(rd, c.LO1)
	case 0x13: // mtlo1
		c.LO1 = c.R[rs].Lo
	case 0x18: // mult1
		p := int64(int32(uint32(c.R[rs].Lo))) * int64(int32(uint32(c.R[rt].Lo)))
		c.LO1, c.HI1 = sext32(uint32(p)), sext32(uint32(uint64(p)>>32))
		c.set(rd, c.LO1)
	case 0x19: // multu1
		p := uint64(uint32(c.R[rs].Lo)) * uint64(uint32(c.R[rt].Lo))
		c.LO1, c.HI1 = sext32(uint32(p)), sext32(uint32(p>>32))
		c.set(rd, c.LO1)
	case 0x1A: // div1
		lo, hi := divSigned32(int32(uint32(c.R[rs].Lo)), int32(uint32(c.R[rt].Lo)))
		c.LO1, c.HI1 = lo, hi
	case 0x1B: // divu1
		lo, hi := divUnsigned32(uint32(c.R[rs].Lo), uint32(c.R[rt].Lo))
		c.LO1, c.HI1 = lo, hi
	case 0x20: // madd1
		c.maddAcc(rd, int64(int32(uint32(c.R[rs].Lo)))*int64(int32(uint32(c.R[rt].Lo))), true)
	case 0x21: // maddu1
		c.maddAcc(rd, int64(uint64(uint32(c.R[rs].Lo))*uint64(uint32(c.R[rt].Lo))), true)
	case 0x28:
		c.mmi1(rs, rt, rd, shamt)
	case 0x29:
		c.mmi3(rs, rt, rd, shamt)
	case 0x30:
		c.pmfhl(rd, shamt)
	case 0x31:
		c.pmthl(rs, shamt)

	// The packed shifts. The halfword forms take a 4-bit amount, the word forms 5.
	// The word forms sign-extend each 32-bit lane result, as every 32-bit operation
	// on this core does.
	case 0x34: // psllh
		h := halves(c.R[rt])
		for i := range h {
			h[i] <<= shamt & 15
		}
		c.putH(rd, h)
	case 0x36: // psrlh
		h := halves(c.R[rt])
		for i := range h {
			h[i] >>= shamt & 15
		}
		c.putH(rd, h)
	case 0x37: // psrah
		h := halves(c.R[rt])
		for i := range h {
			h[i] = uint16(int16(h[i]) >> (shamt & 15))
		}
		c.putH(rd, h)
	case 0x3C: // psllw
		v := words(c.R[rt])
		for i := range v {
			v[i] <<= shamt & 31
		}
		c.putW(rd, v)
	case 0x3E: // psrlw
		v := words(c.R[rt])
		for i := range v {
			v[i] >>= shamt & 31
		}
		c.putW(rd, v)
	case 0x3F: // psraw
		v := words(c.R[rt])
		for i := range v {
			v[i] = uint32(int32(v[i]) >> (shamt & 31))
		}
		c.putW(rd, v)

	default:
		c.Exception(excRI)
	}
}

// maddAcc accumulates a 64-bit product into HI/LO (or HI1/LO1 for the "1" form)
// and writes the resulting low half to rd. The accumulator halves are each kept
// sign-extended from 32 bits, as the hardware does.
func (c *CPU) maddAcc(rd uint32, p int64, second bool) {
	hi, lo := &c.HI, &c.LO
	if second {
		hi, lo = &c.HI1, &c.LO1
	}
	acc := int64(uint64(uint32(*lo)) | uint64(uint32(*hi))<<32)
	sum := uint64(acc + p)
	*lo, *hi = sext32(uint32(sum)), sext32(uint32(sum>>32))
	c.set(rd, *lo)
}

// lzcw counts the leading bits that repeat the sign bit, less one — the shift a
// normaliser would apply. A value of all zeros or all ones gives 31.
func lzcw(v uint32) uint32 {
	if v&0x80000000 != 0 {
		v = ^v
	}
	return uint32(bits.LeadingZeros32(v)) - 1
}

func divSigned32(a, b int32) (lo, hi uint64) {
	switch {
	case b == 0:
		hi = sext32(uint32(a))
		if a >= 0 {
			lo = sext32(0xFFFFFFFF)
		} else {
			lo = 1
		}
	case a == -0x80000000 && b == -1:
		lo, hi = sext32(0x80000000), 0
	default:
		lo, hi = sext32(uint32(a/b)), sext32(uint32(a%b))
	}
	return
}

func divUnsigned32(a, b uint32) (lo, hi uint64) {
	if b == 0 {
		return sext32(0xFFFFFFFF), sext32(a)
	}
	return sext32(a / b), sext32(a % b)
}

// --- MMI0: packed add/subtract/compare/max, the signed saturating forms, the
// low interleaves and the packs ----------------------------------------------

func (c *CPU) mmi0(rs, rt, rd, sub uint32) {
	s, t := c.R[rs], c.R[rt]

	switch sub {
	case 0x00: // paddw
		a, b := words(s), words(t)
		for i := range a {
			a[i] += b[i]
		}
		c.putW(rd, a)
	case 0x01: // psubw
		a, b := words(s), words(t)
		for i := range a {
			a[i] -= b[i]
		}
		c.putW(rd, a)
	case 0x02: // pcgtw — a lane-wise "set to all ones if greater"
		a, b := words(s), words(t)
		for i := range a {
			a[i] = maskIf(int32(a[i]) > int32(b[i]))
		}
		c.putW(rd, a)
	case 0x03: // pmaxw
		a, b := words(s), words(t)
		for i := range a {
			if int32(b[i]) > int32(a[i]) {
				a[i] = b[i]
			}
		}
		c.putW(rd, a)

	case 0x04: // paddh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] += b[i]
		}
		c.putH(rd, a)
	case 0x05: // psubh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] -= b[i]
		}
		c.putH(rd, a)
	case 0x06: // pcgth
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = uint16(maskIf(int16(a[i]) > int16(b[i])))
		}
		c.putH(rd, a)
	case 0x07: // pmaxh
		a, b := halves(s), halves(t)
		for i := range a {
			if int16(b[i]) > int16(a[i]) {
				a[i] = b[i]
			}
		}
		c.putH(rd, a)

	case 0x08: // paddb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] += b[i]
		}
		c.putB(rd, a)
	case 0x09: // psubb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] -= b[i]
		}
		c.putB(rd, a)
	case 0x0A: // pcgtb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = uint8(maskIf(int8(a[i]) > int8(b[i])))
		}
		c.putB(rd, a)

	case 0x10: // paddsw — saturating signed
		a, b := words(s), words(t)
		for i := range a {
			a[i] = satS32(int64(int32(a[i])) + int64(int32(b[i])))
		}
		c.putW(rd, a)
	case 0x11: // psubsw
		a, b := words(s), words(t)
		for i := range a {
			a[i] = satS32(int64(int32(a[i])) - int64(int32(b[i])))
		}
		c.putW(rd, a)
	case 0x12: // pextlw — interleave the low words of rt and rs
		a, b := words(s), words(t)
		c.putW(rd, [4]uint32{b[0], a[0], b[1], a[1]})
	case 0x13: // ppacw — pack the even words of rt then rs
		a, b := words(s), words(t)
		c.putW(rd, [4]uint32{b[0], b[2], a[0], a[2]})

	case 0x14: // paddsh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = satS16(int32(int16(a[i])) + int32(int16(b[i])))
		}
		c.putH(rd, a)
	case 0x15: // psubsh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = satS16(int32(int16(a[i])) - int32(int16(b[i])))
		}
		c.putH(rd, a)
	case 0x16: // pextlh
		a, b := halves(s), halves(t)
		c.putH(rd, [8]uint16{b[0], a[0], b[1], a[1], b[2], a[2], b[3], a[3]})
	case 0x17: // ppach — pack the even halfwords of rt then rs
		a, b := halves(s), halves(t)
		c.putH(rd, [8]uint16{b[0], b[2], b[4], b[6], a[0], a[2], a[4], a[6]})

	case 0x18: // paddsb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = satS8(int32(int8(a[i])) + int32(int8(b[i])))
		}
		c.putB(rd, a)
	case 0x19: // psubsb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = satS8(int32(int8(a[i])) - int32(int8(b[i])))
		}
		c.putB(rd, a)
	case 0x1A: // pextlb
		a, b := octets(s), octets(t)
		var o [16]uint8
		for i := 0; i < 8; i++ {
			o[2*i], o[2*i+1] = b[i], a[i]
		}
		c.putB(rd, o)
	case 0x1B: // ppacb — pack the even bytes of rt then rs
		a, b := octets(s), octets(t)
		var o [16]uint8
		for i := 0; i < 8; i++ {
			o[i], o[i+8] = b[2*i], a[2*i]
		}
		c.putB(rd, o)

	case 0x1E: // pext5 — expand four 16-bit 1:5:5:5 pixels into 32-bit words
		b := words(t)
		var o [4]uint32
		for i := range b {
			v := b[i] & 0xFFFF
			o[i] = (v&0x1F)<<3 | ((v>>5)&0x1F)<<11 | ((v>>10)&0x1F)<<19 | ((v>>15)&1)<<31
		}
		c.putW(rd, o)
	case 0x1F: // ppac5 — pack four 32-bit words back into 1:5:5:5
		b := words(t)
		var o [4]uint32
		for i := range b {
			v := b[i]
			o[i] = (v>>3)&0x1F | ((v>>11)&0x1F)<<5 | ((v>>19)&0x1F)<<10 | ((v>>31)&1)<<15
		}
		c.putW(rd, o)

	default:
		c.Exception(excRI)
	}
}

// --- MMI1: the absolute values, the equality compares, min, the unsigned
// saturating forms, the high interleaves, and the funnel shift ----------------

func (c *CPU) mmi1(rs, rt, rd, sub uint32) {
	s, t := c.R[rs], c.R[rt]

	switch sub {
	case 0x01: // pabsw
		b := words(t)
		for i := range b {
			b[i] = absW(b[i])
		}
		c.putW(rd, b)
	case 0x02: // pceqw
		a, b := words(s), words(t)
		for i := range a {
			a[i] = maskIf(a[i] == b[i])
		}
		c.putW(rd, a)
	case 0x03: // pminw
		a, b := words(s), words(t)
		for i := range a {
			if int32(b[i]) < int32(a[i]) {
				a[i] = b[i]
			}
		}
		c.putW(rd, a)

	case 0x04: // padsbh — the low half subtracts, the high half adds
		a, b := halves(s), halves(t)
		for i := 0; i < 4; i++ {
			a[i] -= b[i]
		}
		for i := 4; i < 8; i++ {
			a[i] += b[i]
		}
		c.putH(rd, a)
	case 0x05: // pabsh
		b := halves(t)
		for i := range b {
			b[i] = absH(b[i])
		}
		c.putH(rd, b)
	case 0x06: // pceqh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = uint16(maskIf(a[i] == b[i]))
		}
		c.putH(rd, a)
	case 0x07: // pminh
		a, b := halves(s), halves(t)
		for i := range a {
			if int16(b[i]) < int16(a[i]) {
				a[i] = b[i]
			}
		}
		c.putH(rd, a)

	case 0x0A: // pceqb
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = uint8(maskIf(a[i] == b[i]))
		}
		c.putB(rd, a)

	case 0x10: // padduw — saturating unsigned
		a, b := words(s), words(t)
		for i := range a {
			a[i] = satU32(int64(a[i]) + int64(b[i]))
		}
		c.putW(rd, a)
	case 0x11: // psubuw
		a, b := words(s), words(t)
		for i := range a {
			a[i] = satU32(int64(a[i]) - int64(b[i]))
		}
		c.putW(rd, a)
	case 0x12: // pextuw — interleave the high words
		a, b := words(s), words(t)
		c.putW(rd, [4]uint32{b[2], a[2], b[3], a[3]})

	case 0x14: // padduh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = satU16(int32(a[i]) + int32(b[i]))
		}
		c.putH(rd, a)
	case 0x15: // psubuh
		a, b := halves(s), halves(t)
		for i := range a {
			a[i] = satU16(int32(a[i]) - int32(b[i]))
		}
		c.putH(rd, a)
	case 0x16: // pextuh
		a, b := halves(s), halves(t)
		c.putH(rd, [8]uint16{b[4], a[4], b[5], a[5], b[6], a[6], b[7], a[7]})

	case 0x18: // paddub
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = satU8(int32(a[i]) + int32(b[i]))
		}
		c.putB(rd, a)
	case 0x19: // psubub
		a, b := octets(s), octets(t)
		for i := range a {
			a[i] = satU8(int32(a[i]) - int32(b[i]))
		}
		c.putB(rd, a)
	case 0x1A: // pextub
		a, b := octets(s), octets(t)
		var o [16]uint8
		for i := 0; i < 8; i++ {
			o[2*i], o[2*i+1] = b[i+8], a[i+8]
		}
		c.putB(rd, o)

	case 0x1B: // qfsrv — the funnel shift: SA bytes out of the pair (rs:rt)
		// SA is a *bit* count here (mtsab/mtsah scale their operand into it), and
		// the shift runs right across the 256-bit concatenation rs:rt.
		n := c.SA & 0x7F
		c.putQ(rd, funnelRight(s, t, n))

	default:
		c.Exception(excRI)
	}
}

// --- MMI2: the word multiplies and divides, the variable shifts, the halfword
// multiply-accumulates, and the permutes ------------------------------------

func (c *CPU) mmi2(rs, rt, rd, sub uint32) {
	s, t := c.R[rs], c.R[rt]

	switch sub {
	case 0x00: // pmaddw
		c.pmulW(rd, s, t, true, false, true)
	case 0x02: // psllvw — a per-lane variable shift, on words 0 and 2 only
		a, b := words(s), words(t)
		c.putQ(rd, Quad{
			Lo: sext32(b[0] << (a[0] & 31)),
			Hi: sext32(b[2] << (a[2] & 31)),
		})
	case 0x03: // psrlvw
		a, b := words(s), words(t)
		c.putQ(rd, Quad{
			Lo: sext32(b[0] >> (a[0] & 31)),
			Hi: sext32(b[2] >> (a[2] & 31)),
		})
	case 0x04: // pmsubw
		c.pmulW(rd, s, t, true, true, true)
	case 0x08: // pmfhi
		c.putQ(rd, c.hiQ())
	case 0x09: // pmflo
		c.putQ(rd, c.loQ())
	case 0x0A: // pinth — rt's low halves interleaved with rs's high halves
		a, b := halves(s), halves(t)
		c.putH(rd, [8]uint16{b[0], a[4], b[1], a[5], b[2], a[6], b[3], a[7]})
	case 0x0C: // pmultw
		c.pmulW(rd, s, t, false, false, true)
	case 0x0D: // pdivw
		a, b := words(s), words(t)
		lo0, hi0 := divSigned32(int32(a[0]), int32(b[0]))
		lo1, hi1 := divSigned32(int32(a[2]), int32(b[2]))
		c.LO, c.HI, c.LO1, c.HI1 = lo0, hi0, lo1, hi1
	case 0x0E: // pcpyld — rt's low doubleword under rs's low doubleword
		c.putQ(rd, Quad{Lo: t.Lo, Hi: s.Lo})
	case 0x10: // pmaddh
		c.pmulH(rd, s, t, true, false)
	case 0x11: // phmadh
		c.phmulH(rd, s, t, false)
	case 0x12: // pand
		c.putQ(rd, Quad{Lo: s.Lo & t.Lo, Hi: s.Hi & t.Hi})
	case 0x13: // pxor
		c.putQ(rd, Quad{Lo: s.Lo ^ t.Lo, Hi: s.Hi ^ t.Hi})
	case 0x14: // pmsubh
		c.pmulH(rd, s, t, true, true)
	case 0x15: // phmsbh
		c.phmulH(rd, s, t, true)
	case 0x1A: // pexeh — exchange the even halfwords of each doubleword
		h := halves(t)
		c.putH(rd, [8]uint16{h[2], h[1], h[0], h[3], h[6], h[5], h[4], h[7]})
	case 0x1B: // prevh — reverse the halfwords of each doubleword
		h := halves(t)
		c.putH(rd, [8]uint16{h[3], h[2], h[1], h[0], h[7], h[6], h[5], h[4]})
	case 0x1C: // pmulth
		c.pmulH(rd, s, t, false, false)
	case 0x1D: // pdivbw — divide each word of rs by rt's low halfword
		a, b := words(s), words(t)
		d := int32(int16(uint16(b[0])))
		var lo, hi [4]uint32
		for i := range a {
			q, r := divSigned32(int32(a[i]), d)
			lo[i], hi[i] = uint32(q), uint32(r)
		}
		c.setLoQ(fromWords(lo))
		c.setHiQ(fromWords(hi))
	case 0x1E: // pexew — exchange the even words of each doubleword pair
		v := words(t)
		c.putW(rd, [4]uint32{v[2], v[1], v[0], v[3]})
	case 0x1F: // prot3w — rotate the low three words
		v := words(t)
		c.putW(rd, [4]uint32{v[1], v[2], v[0], v[3]})

	default:
		c.Exception(excRI)
	}
}

// --- MMI3: the unsigned word multiplies, the arithmetic variable shift, the
// accumulator moves and the remaining permutes -------------------------------

func (c *CPU) mmi3(rs, rt, rd, sub uint32) {
	s, t := c.R[rs], c.R[rt]

	switch sub {
	case 0x00: // pmadduw
		c.pmulW(rd, s, t, true, false, false)
	case 0x03: // psravw
		a, b := words(s), words(t)
		c.putQ(rd, Quad{
			Lo: sext32(uint32(int32(b[0]) >> (a[0] & 31))),
			Hi: sext32(uint32(int32(b[2]) >> (a[2] & 31))),
		})
	case 0x08: // pmthi
		c.HI, c.HI1 = s.Lo, s.Hi
	case 0x09: // pmtlo
		c.LO, c.LO1 = s.Lo, s.Hi
	case 0x0A: // pinteh — rt's even halves interleaved with rs's even halves
		a, b := halves(s), halves(t)
		c.putH(rd, [8]uint16{b[0], a[0], b[2], a[2], b[4], a[4], b[6], a[6]})
	case 0x0C: // pmultuw
		c.pmulW(rd, s, t, false, false, false)
	case 0x0D: // pdivuw
		a, b := words(s), words(t)
		lo0, hi0 := divUnsigned32(a[0], b[0])
		lo1, hi1 := divUnsigned32(a[2], b[2])
		c.LO, c.HI, c.LO1, c.HI1 = lo0, hi0, lo1, hi1
	case 0x0E: // pcpyud — rs's high doubleword under rt's high doubleword
		c.putQ(rd, Quad{Lo: s.Hi, Hi: t.Hi})
	case 0x12: // por
		c.putQ(rd, Quad{Lo: s.Lo | t.Lo, Hi: s.Hi | t.Hi})
	case 0x13: // pnor
		c.putQ(rd, Quad{Lo: ^(s.Lo | t.Lo), Hi: ^(s.Hi | t.Hi)})
	case 0x1A: // pexch — exchange the centre halfwords of each doubleword
		h := halves(t)
		c.putH(rd, [8]uint16{h[0], h[2], h[1], h[3], h[4], h[6], h[5], h[7]})
	case 0x1B: // pcpyh — broadcast the low halfword of each doubleword
		h := halves(t)
		c.putH(rd, [8]uint16{h[0], h[0], h[0], h[0], h[4], h[4], h[4], h[4]})
	case 0x1E: // pexcw — exchange the centre words
		v := words(t)
		c.putW(rd, [4]uint32{v[0], v[2], v[1], v[3]})

	default:
		c.Exception(excRI)
	}
}

// --- the packed multiplies --------------------------------------------------

// pmulW is the word-multiply family: pmultw/pmaddw/pmsubw and their unsigned
// forms. Words 0 and 2 of each operand are multiplied to 64 bits; the results land
// in the 128-bit accumulator (word pair 0/1 in HI/LO, pair 2/3 in HI1/LO1), and
// the same two 64-bit values are written to rd.
func (c *CPU) pmulW(rd uint32, s, t Quad, accumulate, subtract, signed bool) {
	a, b := words(s), words(t)

	prod := func(i int) uint64 {
		if signed {
			return uint64(int64(int32(a[i])) * int64(int32(b[i])))
		}
		return uint64(a[i]) * uint64(b[i])
	}

	p0, p1 := prod(0), prod(2)

	if accumulate {
		acc0 := uint64(uint32(c.LO)) | uint64(uint32(c.HI))<<32
		acc1 := uint64(uint32(c.LO1)) | uint64(uint32(c.HI1))<<32
		if subtract {
			p0, p1 = acc0-p0, acc1-p1
		} else {
			p0, p1 = acc0+p0, acc1+p1
		}
	}

	c.LO, c.HI = sext32(uint32(p0)), sext32(uint32(p0>>32))
	c.LO1, c.HI1 = sext32(uint32(p1)), sext32(uint32(p1>>32))
	c.putQ(rd, Quad{Lo: p0, Hi: p1})
}

// pmulH is the halfword-multiply family: pmulth/pmaddh/pmsubh. Eight 16x16
// products become eight 32-bit values, laid across the 128-bit HI/LO pair in the
// order LO.w0, LO.w1, HI.w0, HI.w1, LO1.w0, LO1.w1, HI1.w0, HI1.w1. rd receives
// products 0, 1, 4 and 5 — the LO halves — which is what makes a chain of these
// compute a dot product with the accumulator left holding the rest.
func (c *CPU) pmulH(rd uint32, s, t Quad, accumulate, subtract bool) {
	a, b := halves(s), halves(t)
	var p [8]uint32
	for i := 0; i < 8; i++ {
		p[i] = uint32(int32(int16(a[i])) * int32(int16(b[i])))
	}
	if accumulate {
		acc := c.accWords()
		for i := 0; i < 8; i++ {
			if subtract {
				p[i] = acc[i] - p[i]
			} else {
				p[i] = acc[i] + p[i]
			}
		}
	}
	c.setAccWords(p)
	c.putW(rd, [4]uint32{p[0], p[1], p[4], p[5]})
}

// phmulH is the horizontal form: phmadh/phmsbh multiply adjacent halfword pairs
// and sum each pair into one 32-bit result, giving four results rather than eight.
// The accumulator lanes that the vertical form would have filled with the second
// product of each pair are written with the same total, which is what lets pmfhl
// read the result back either way.
func (c *CPU) phmulH(rd uint32, s, t Quad, subtract bool) {
	a, b := halves(s), halves(t)
	var r [4]uint32
	for i := 0; i < 4; i++ {
		lo := int32(int16(a[2*i])) * int32(int16(b[2*i]))
		hi := int32(int16(a[2*i+1])) * int32(int16(b[2*i+1]))
		if subtract {
			r[i] = uint32(hi - lo)
		} else {
			r[i] = uint32(hi + lo)
		}
	}
	// Lanes 0,1 -> LO; 2,3 -> HI; 4,5 -> LO1; 6,7 -> HI1.
	c.setAccWords([8]uint32{r[0], r[0], r[1], r[1], r[2], r[2], r[3], r[3]})
	c.putW(rd, [4]uint32{r[0], r[1], r[2], r[3]})
}

// accWords views the 128-bit HI/LO accumulator pair as eight 32-bit lanes, in the
// order the packed multiplies write them.
func (c *CPU) accWords() [8]uint32 {
	return [8]uint32{
		uint32(c.LO), uint32(c.LO >> 32),
		uint32(c.HI), uint32(c.HI >> 32),
		uint32(c.LO1), uint32(c.LO1 >> 32),
		uint32(c.HI1), uint32(c.HI1 >> 32),
	}
}

func (c *CPU) setAccWords(p [8]uint32) {
	c.LO = uint64(p[0]) | uint64(p[1])<<32
	c.HI = uint64(p[2]) | uint64(p[3])<<32
	c.LO1 = uint64(p[4]) | uint64(p[5])<<32
	c.HI1 = uint64(p[6]) | uint64(p[7])<<32
}

// --- pmfhl / pmthl ----------------------------------------------------------

// pmfhl moves the accumulator pair into a general register. The sub-function
// selects which slice: the low or high words of each lane, a sign-extended
// concatenation, or a saturated halfword pack.
func (c *CPU) pmfhl(rd, sub uint32) {
	acc := c.accWords()
	switch sub {
	case 0x00: // pmfhl.lw — the low word of each of the four 64-bit lanes
		c.putW(rd, [4]uint32{acc[0], acc[2], acc[4], acc[6]})
	case 0x01: // pmfhl.uw — the high word of each
		c.putW(rd, [4]uint32{acc[1], acc[3], acc[5], acc[7]})
	case 0x02: // pmfhl.slw — the two 64-bit values, saturated to 64 bits
		c.putQ(rd, Quad{
			Lo: satS64(int64(int32(acc[2]))<<32 | int64(uint32(acc[0]))),
			Hi: satS64(int64(int32(acc[6]))<<32 | int64(uint32(acc[4]))),
		})
	case 0x03: // pmfhl.lh — the low halfword of each of the eight 32-bit lanes
		var h [8]uint16
		for i := 0; i < 8; i++ {
			h[i] = uint16(acc[i])
		}
		c.putH(rd, h)
	case 0x04: // pmfhl.sh — each lane saturated to a halfword
		var h [8]uint16
		for i := 0; i < 8; i++ {
			h[i] = satS16(int32(acc[i]))
		}
		c.putH(rd, h)
	default:
		c.Exception(excRI)
	}
}

// pmthl writes a general register back into the accumulator. Only the "lw" form
// exists: it replaces the low word of each 64-bit lane, leaving the high words.
func (c *CPU) pmthl(rs, sub uint32) {
	if sub != 0 {
		c.Exception(excRI)
		return
	}
	v := words(c.R[rs])
	acc := c.accWords()
	acc[0], acc[2], acc[4], acc[6] = v[0], v[1], v[2], v[3]
	c.setAccWords(acc)
}

// satS64 is the identity here — the value already is 64 bits — but naming it keeps
// pmfhl.slw's intent legible next to its saturating siblings.
func satS64(v int64) uint64 { return uint64(v) }

// --- small helpers ----------------------------------------------------------

// maskIf turns a lane-wise comparison into the all-ones / all-zeros mask the
// packed compares produce.
func maskIf(b bool) uint32 {
	if b {
		return 0xFFFFFFFF
	}
	return 0
}

func absW(v uint32) uint32 {
	if v == 0x80000000 {
		return 0x7FFFFFFF // the negation would overflow, so it saturates
	}
	if int32(v) < 0 {
		return uint32(-int32(v))
	}
	return v
}

func absH(v uint16) uint16 {
	if v == 0x8000 {
		return 0x7FFF
	}
	if int16(v) < 0 {
		return uint16(-int16(v))
	}
	return v
}

// funnelRight shifts the 256-bit concatenation (hi:lo) right by n bits and returns
// the low 128. qfsrv uses it to lift a quadword out of an unaligned pair.
func funnelRight(hi, lo Quad, n uint32) Quad {
	if n == 0 {
		return lo
	}
	if n >= 128 {
		return shiftRight128(hi, n-128)
	}
	// The result is lo >> n, with the bits shifted out of hi's bottom shifted in.
	r := shiftRight128(lo, n)
	l := shiftLeft128(hi, 128-n)
	return Quad{Lo: r.Lo | l.Lo, Hi: r.Hi | l.Hi}
}

func shiftRight128(q Quad, n uint32) Quad {
	switch {
	case n == 0:
		return q
	case n >= 128:
		return Quad{}
	case n >= 64:
		return Quad{Lo: q.Hi >> (n - 64)}
	default:
		return Quad{Lo: q.Lo>>n | q.Hi<<(64-n), Hi: q.Hi >> n}
	}
}

func shiftLeft128(q Quad, n uint32) Quad {
	switch {
	case n == 0:
		return q
	case n >= 128:
		return Quad{}
	case n >= 64:
		return Quad{Hi: q.Lo << (n - 64)}
	default:
		return Quad{Lo: q.Lo << n, Hi: q.Hi<<n | q.Lo>>(64-n)}
	}
}

// putW / putH / putB / putQ write a whole 128-bit register from an element view.
func (c *CPU) putW(rd uint32, v [4]uint32) { c.putQ(rd, fromWords(v)) }
func (c *CPU) putH(rd uint32, v [8]uint16) { c.putQ(rd, fromHalves(v)) }
func (c *CPU) putB(rd uint32, v [16]uint8) { c.putQ(rd, fromOctets(v)) }

func (c *CPU) putQ(rd uint32, q Quad) { c.setQ(rd, q.Lo, q.Hi) }
