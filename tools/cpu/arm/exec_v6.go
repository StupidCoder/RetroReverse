package arm

import "math/bits"

// execARMv6 executes the conditional ARMv6K additions. It returns true when it
// recognised and ran the instruction, and false to let the ARMv5TE core handle
// it. The dispatch mirrors decodeARMv6 exactly, including the two aliasing cases
// (LDREX/STREX over SWP, UMAAL over MUL) that must be intercepted before the
// base core sees them.
func (c *CPU) execARMv6(w uint32) bool {
	// VFP coprocessor (cp 10/11): the load/store group (bits 27:25 == 110) and
	// the data-processing / register-transfer group (bits 27:24 == 1110).
	if isVFP(w) && ((w>>25)&7 == 0b110 || (w>>24)&0xF == 0b1110) {
		return c.execVFP(w)
	}
	// Synchronization primitives (bit 23 of the SWP slot set).
	if (w>>24)&0xF == 0b0001 && (w>>4)&0xF == 0b1001 && (w>>23)&1 == 1 {
		c.execSync(w)
		return true
	}
	// UMAAL, sharing the multiply slot.
	if (w>>24)&0xF == 0b0000 && (w>>4)&0xF == 0b1001 && (w>>20)&0xF == 0b0100 {
		c.execUMAAL(w)
		return true
	}
	// Media space.
	if (w>>25)&7 == 0b011 && (w>>4)&1 == 1 {
		return c.execMedia(w)
	}
	return false
}

// execUncondARMv6 runs the unconditional-space additions (CPS, SETEND, CLREX).
func (c *CPU) execUncondARMv6(w uint32) bool {
	switch {
	case (w>>20)&0xFF == 0b00010000 && (w>>16)&1 == 0: // CPS
		if (w>>17)&1 == 1 { // mode change
			c.switchMode(w & 0x1F)
		}
		imod := (w >> 18) & 3
		if imod == 0b10 || imod == 0b11 { // IE / ID
			disable := imod == 0b11
			if (w>>7)&1 == 1 {
				c.IRQDisable = disable
			}
			if (w>>6)&1 == 1 {
				c.FIQDisable = disable
			}
		}
		c.R[15] = c.cur + 4
		return true
	case (w>>16)&0xFFFF == 0b1111000100000001 && (w>>4)&1 == 0: // SETEND
		c.BigEndian = (w>>9)&1 == 1
		c.R[15] = c.cur + 4
		return true
	case (w>>20)&0xFF == 0b01010111 && (w>>4)&0xF == 0b0001: // CLREX
		c.exclValid = false
		c.R[15] = c.cur + 4
		return true
	}
	return false
}

// --- synchronization -------------------------------------------------------

func (c *CPU) execSync(w uint32) {
	load := (w>>20)&1 == 1
	sz := (w >> 21) & 3 // 0 word, 1 doubleword, 2 byte, 3 halfword
	addr := c.reg((w >> 16) & 0xF)

	if load {
		rt := (w >> 12) & 0xF
		switch sz {
		case 0:
			c.setReg(rt, c.read32(addr))
		case 1:
			// LDREXD Rt, Rt2, [Rn]: the doubleword lands in the register pair
			// Rt (even) and Rt+1, low word first — Rt2 is implicit, not encoded.
			c.setReg(rt, c.read32(addr))
			c.setReg(rt+1, c.read32(addr+4))
		case 2:
			c.setReg(rt, uint32(c.read8(addr)))
		case 3:
			c.setReg(rt, c.read16(addr))
		}
		c.exclValid = true
		c.exclAddr = addr
		return
	}

	// STREX Rd, Rt, [Rn]: succeed only if this core still holds the tag for the
	// address. A single-core HLE needs nothing stronger than an address match.
	rd := (w >> 12) & 0xF
	rt := c.reg(w & 0xF)
	if !c.exclValid || c.exclAddr != addr {
		c.setReg(rd, 1) // failed
		return
	}
	switch sz {
	case 0:
		c.write32(addr, rt)
	case 1:
		// STREXD Rd, Rt, Rt2, [Rn]: the value is the pair Rt (encoded at 3:0,
		// even) and Rt+1, low word first.
		c.write32(addr, rt)
		c.write32(addr+4, c.reg((w&0xF)+1))
	case 2:
		c.write8(addr, byte(rt))
	case 3:
		c.write16(addr, rt)
	}
	c.setReg(rd, 0) // succeeded
	c.exclValid = false
}

func (c *CPU) execUMAAL(w uint32) {
	rdHi := (w >> 16) & 0xF
	rdLo := (w >> 12) & 0xF
	rm := c.reg((w >> 8) & 0xF)
	rn := c.reg(w & 0xF)
	// UMAAL: {RdHi:RdLo} = Rn*Rm + RdHi + RdLo, all unsigned, no carry lost.
	res := uint64(rn)*uint64(rm) + uint64(c.reg(rdHi)) + uint64(c.reg(rdLo))
	c.setReg(rdLo, uint32(res))
	c.setReg(rdHi, uint32(res>>32))
}

// --- media -----------------------------------------------------------------

func (c *CPU) execMedia(w uint32) bool {
	op1 := (w >> 20) & 0x1F
	switch {
	case op1>>3 == 0b00 && op1&7 != 0:
		return c.execParallel(w)
	case op1>>3 == 0b01:
		return c.execMediaPack(w)
	case op1>>3 == 0b10:
		return c.execMediaMul(w)
	case op1 == 0b11000: // USAD8 / USADA8
		c.execUSAD(w)
		return true
	}
	c.Halt("unimplemented media instruction 0x%08X at 0x%08X", w, c.cur)
	return true
}

// execParallel runs the ARMv6 byte/halfword parallel (SIMD) add and subtract
// group. Three things vary independently and the encoding keeps them apart:
//
//   - the lane width and pairing (bits 7:5): ADD16, ASX, SAX, SUB16, ADD8, SUB8.
//     The two "exchange" forms cross the halfwords of Rm — ASX subtracts the high
//     halfword from the low lane and adds the low one into the high lane; SAX is
//     the mirror.
//   - signedness, and
//   - what happens to a lane result that will not fit (bits 22:20, the class
//     prefix): the plain forms (S, U) truncate and publish the per-lane GE flags
//     that a following SEL consumes; the saturating forms (Q, UQ) clamp to the
//     lane and leave GE alone; the halving forms (SH, UH) shift the full-precision
//     result right by one, so no lane can overflow, and also leave GE alone.
func (c *CPU) execParallel(w uint32) bool {
	class := (w >> 20) & 7 // 1 S, 2 Q, 3 SH, 5 U, 6 UQ, 7 UH
	op2 := (w >> 5) & 7
	rn := c.reg((w >> 16) & 0xF)
	rm := c.reg(w & 0xF)
	rd := (w >> 12) & 0xF

	var signed bool
	var mode parMode
	switch class {
	case 1:
		signed, mode = true, parWrap
	case 2:
		signed, mode = true, parSat
	case 3:
		signed, mode = true, parHalve
	case 5:
		signed, mode = false, parWrap
	case 6:
		signed, mode = false, parSat
	case 7:
		signed, mode = false, parHalve
	default:
		c.Halt("unimplemented parallel arithmetic 0x%08X (class %d) at 0x%08X", w, class, c.cur)
		return true
	}

	// half picks halfword h (0 = low, 1 = high) out of a register.
	half := func(v uint32, h uint) uint32 { return (v >> (16 * h)) & 0xFFFF }

	var res uint32
	var ge uint32 // one bit per byte lane, as the GE field is laid out

	switch op2 {
	case 0b000, 0b011: // ADD16 / SUB16 — lane-aligned halfwords
		sub := op2 == 0b011
		for h := uint(0); h < 2; h++ {
			v, g := parLane(half(rn, h), half(rm, h), 16, signed, mode, sub)
			res |= v << (16 * h)
			if g {
				ge |= 0x3 << (2 * h) // a halfword lane owns two GE bits
			}
		}
	case 0b001: // ASX: low lane subtracts Rm's HIGH half; high lane adds Rm's low
		lo, glo := parLane(half(rn, 0), half(rm, 1), 16, signed, mode, true)
		hi, ghi := parLane(half(rn, 1), half(rm, 0), 16, signed, mode, false)
		res = lo | hi<<16
		if glo {
			ge |= 0x3
		}
		if ghi {
			ge |= 0xC
		}
	case 0b010: // SAX: low lane adds Rm's high half; high lane subtracts Rm's low
		lo, glo := parLane(half(rn, 0), half(rm, 1), 16, signed, mode, false)
		hi, ghi := parLane(half(rn, 1), half(rm, 0), 16, signed, mode, true)
		res = lo | hi<<16
		if glo {
			ge |= 0x3
		}
		if ghi {
			ge |= 0xC
		}
	case 0b100, 0b111: // ADD8 / SUB8
		sub := op2 == 0b111
		for b := uint(0); b < 4; b++ {
			av := (rn >> (8 * b)) & 0xFF
			bv := (rm >> (8 * b)) & 0xFF
			v, g := parLane(av, bv, 8, signed, mode, sub)
			res |= v << (8 * b)
			if g {
				ge |= 1 << b
			}
		}
	default:
		c.Halt("unimplemented parallel op2=%d at 0x%08X", op2, c.cur)
		return true
	}

	c.setReg(rd, res)
	if mode == parWrap {
		// Only the truncating forms publish GE; Q/UQ and SH/UH leave it as it was.
		c.GE = ge
	}
	return true
}

// parMode is what the class prefix does to a lane result that will not fit.
type parMode int

const (
	parWrap  parMode = iota // S / U   — truncate, and set the GE flags
	parSat                  // Q / UQ  — saturate to the lane, GE untouched
	parHalve                // SH / UH — halve the full result, GE untouched
)

// parLane computes one SIMD lane: av op bv at the given width, interpreted signed
// or unsigned, with the class modifier applied. The bool is that lane's GE bit,
// which is meaningful only for the truncating classes: for signed lanes GE means
// "the result is not negative"; for unsigned it means a carry out of an add, or
// the absence of a borrow from a subtract (i.e. the operation did not wrap).
func parLane(av, bv uint32, width uint, signed bool, mode parMode, sub bool) (uint32, bool) {
	mask := uint32(1)<<width - 1

	var val int32
	if signed {
		a, b := signExtendLane(av, width), signExtendLane(bv, width)
		if sub {
			val = a - b
		} else {
			val = a + b
		}
	} else if sub {
		val = int32(av) - int32(bv)
	} else {
		val = int32(av) + int32(bv)
	}

	var ge bool
	switch {
	case signed:
		ge = val >= 0
	case sub:
		ge = av >= bv // no borrow
	default:
		ge = val > int32(mask) // carry out
	}

	switch mode {
	case parSat:
		if signed {
			hi := int32(1)<<(width-1) - 1
			lo := -(int32(1) << (width - 1))
			if val > hi {
				val = hi
			} else if val < lo {
				val = lo
			}
		} else {
			if val < 0 {
				val = 0
			} else if val > int32(mask) {
				val = int32(mask)
			}
		}
	case parHalve:
		val >>= 1 // arithmetic for signed, and the unsigned value is non-negative
	}
	return uint32(val) & mask, ge
}

// int16v sign-extends the low halfword of v — the lane accessor the signed dual
// multiplies use.
func int16v(v uint32) int32 { return int32(int16(uint16(v))) }

// signExtendLane sign-extends a width-bit lane to int32.
func signExtendLane(v uint32, width uint) int32 {
	sh := uint(32) - width
	return int32(v<<sh) >> sh
}

func (c *CPU) execMediaPack(w uint32) bool {
	op1 := (w >> 20) & 0x1F
	op2 := (w >> 5) & 7
	rd := (w >> 12) & 0xF
	rnv := c.reg((w >> 16) & 0xF)
	rmv := c.reg(w & 0xF)

	switch {
	case op1 == 0b01000 && op2&1 == 0: // PKHBT / PKHTB
		imm := (w >> 7) & 0x1F
		var res uint32
		if (w>>6)&1 == 0 { // PKHBT: bottom from Rn, top from Rm<<imm
			res = (rnv & 0xFFFF) | ((rmv << imm) & 0xFFFF0000)
		} else { // PKHTB: top from Rn, bottom from Rm>>imm (ASR)
			if imm == 0 {
				imm = 32
			}
			res = (rnv & 0xFFFF0000) | (uint32(int32(rmv)>>imm) & 0xFFFF)
		}
		c.setReg(rd, res)
		return true

	case (op1 == 0b01010 || op1 == 0b01011) && op2&1 == 0: // SSAT
		c.setReg(rd, c.doSat(w, true))
		return true
	case (op1 == 0b01110 || op1 == 0b01111) && op2&1 == 0: // USAT
		c.setReg(rd, c.doSat(w, false))
		return true

	case op2 == 0b011: // extend family
		if c.execExtend(w) {
			return true
		}
	case op1 == 0b01011 && op2 == 0b001: // REV
		c.setReg(rd, bits.ReverseBytes32(rmv))
		return true
	case op1 == 0b01011 && op2 == 0b101: // REV16 — reverse bytes in each halfword
		c.setReg(rd, (rmv&0x00FF00FF)<<8|(rmv&0xFF00FF00)>>8)
		return true
	case op1 == 0b01111 && op2 == 0b101: // REVSH — reverse low halfword, sign-extend
		v := (rmv&0xFF)<<8 | (rmv&0xFF00)>>8
		c.setReg(rd, uint32(int32(int16(uint16(v)))))
		return true
	case op1 == 0b01000 && op2 == 0b101: // SEL — pick bytes by GE
		var res uint32
		for i := 0; i < 4; i++ {
			sh := uint(i * 8)
			if c.GE&(1<<uint(i)) != 0 {
				res |= rnv & (0xFF << sh)
			} else {
				res |= rmv & (0xFF << sh)
			}
		}
		c.setReg(rd, res)
		return true
	}
	c.Halt("unimplemented media pack 0x%08X at 0x%08X", w, c.cur)
	return true
}

// doSat implements SSAT/USAT: rotate/shift Rm, then saturate to a bit width.
func (c *CPU) doSat(w uint32, signed bool) uint32 {
	rmv := c.reg(w & 0xF)
	amt := (w >> 7) & 0x1F
	val := int32(rmv)
	if (w>>6)&1 == 1 { // ASR
		if amt == 0 {
			amt = 32
		}
		val = val >> amt
	} else if amt != 0 { // LSL
		val = int32(rmv << amt)
	}
	if signed {
		n := (w>>16)&0x1F + 1
		hi := int32(1)<<(n-1) - 1
		lo := -int32(1) << (n - 1)
		if val > hi {
			c.Q = true
			return uint32(hi)
		}
		if val < lo {
			c.Q = true
			return uint32(lo)
		}
		return uint32(val)
	}
	n := (w >> 16) & 0x1F
	hi := int32((uint32(1) << n) - 1)
	if val > hi {
		c.Q = true
		return uint32(hi)
	}
	if val < 0 {
		c.Q = true
		return 0
	}
	return uint32(val)
}

// execExtend runs the sign/zero extend family (SXTB/SXTH/UXTB/UXTH, the 16-bit
// pair forms, and their accumulate variants when Rn != 15).
func (c *CPU) execExtend(w uint32) bool {
	op1 := (w >> 20) & 0x1F
	rd := (w >> 12) & 0xF
	rn := (w >> 16) & 0xF
	rmv := c.reg(w & 0xF)
	rot := (w >> 10) & 3 * 8
	rotated := ror32(rmv, rot)
	acc := rn != 0xF
	rnv := c.reg(rn)

	var res uint32
	switch op1 {
	case 0b01010: // SXTB / SXTAB
		res = uint32(int32(int8(byte(rotated))))
		if acc {
			res += rnv
		}
	case 0b01011: // SXTH / SXTAH
		res = uint32(int32(int16(uint16(rotated))))
		if acc {
			res += rnv
		}
	case 0b01110: // UXTB / UXTAB
		res = rotated & 0xFF
		if acc {
			res += rnv
		}
	case 0b01111: // UXTH / UXTAH
		res = rotated & 0xFFFF
		if acc {
			res += rnv
		}
	case 0b01000: // SXTB16 / SXTAB16
		lo := uint32(int32(int8(byte(rotated)))) & 0xFFFF
		hi := uint32(int32(int8(byte(rotated>>16)))) & 0xFFFF
		if acc {
			lo = (lo + rnv) & 0xFFFF
			hi = (hi + (rnv >> 16)) & 0xFFFF
		}
		res = lo | hi<<16
	case 0b01100: // UXTB16 / UXTAB16
		lo := rotated & 0xFF
		hi := (rotated >> 16) & 0xFF
		if acc {
			lo = (lo + rnv) & 0xFFFF
			hi = (hi + (rnv >> 16)) & 0xFFFF
		}
		res = lo | hi<<16
	default:
		return false
	}
	c.setReg(rd, res)
	return true
}

// execMediaMul runs the signed dual multiplies and the most-significant-word
// multiplies (SMUAD/SMUSD/SMLAD/SMLSD, SMMUL/SMMLA/SMMLS).
func (c *CPU) execMediaMul(w uint32) bool {
	op1 := (w >> 20) & 0x1F
	op2 := (w >> 5) & 7
	rd := (w >> 16) & 0xF
	ra := (w >> 12) & 0xF
	rmv := c.reg((w >> 8) & 0xF)
	rnv := c.reg(w & 0xF)
	swap := (w>>5)&1 == 1 // the X variants swap Rm's halves

	m := rmv
	if swap {
		m = (rmv >> 16) | (rmv << 16)
	}

	switch op1 {
	case 0b10000: // SMLAD/SMUAD/SMLSD/SMUSD
		p0 := int16v(rnv) * int16v(m)
		p1 := int16v(rnv>>16) * int16v(m>>16)
		var dual int32
		if op2>>1 == 0b00 { // add pair
			dual = p0 + p1
		} else { // op2>>1 == 01: subtract pair
			dual = p0 - p1
		}
		if ra != 0xF {
			dual += int32(c.reg(ra))
		}
		c.setReg(rd, uint32(dual))
		return true
	case 0b10101: // SMMUL/SMMLA/SMMLS
		prod := int64(int32(rnv)) * int64(int32(m))
		round := op2&1 == 1
		if round {
			prod += 0x80000000
		}
		var res int32
		switch op2 >> 1 {
		case 0b00: // SMMUL / SMMLA
			res = int32(prod >> 32)
			if ra != 0xF {
				res += int32(c.reg(ra))
			}
		case 0b11: // SMMLS
			res = int32(c.reg(ra)) - int32(prod>>32)
		default:
			return false
		}
		c.setReg(rd, uint32(res))
		return true
	}
	c.Halt("unimplemented media multiply 0x%08X at 0x%08X", w, c.cur)
	return true
}

// execUSAD runs USAD8 / USADA8: the sum of absolute byte differences.
func (c *CPU) execUSAD(w uint32) {
	rd := (w >> 16) & 0xF
	ra := (w >> 12) & 0xF
	rnv := c.reg(w & 0xF)
	rmv := c.reg((w >> 8) & 0xF)
	var sum uint32
	for i := 0; i < 4; i++ {
		sh := uint(i * 8)
		d := int32((rnv>>sh)&0xFF) - int32((rmv>>sh)&0xFF)
		if d < 0 {
			d = -d
		}
		sum += uint32(d)
	}
	if ra != 0xF {
		sum += c.reg(ra)
	}
	c.setReg(rd, sum)
}
