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
		case 2:
			c.setReg(rt, uint32(c.read8(addr)))
		case 3:
			c.setReg(rt, c.read16(addr))
		default:
			c.Halt("unimplemented LDREXD at 0x%08X", c.cur)
			return
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
	case 2:
		c.write8(addr, byte(rt))
	case 3:
		c.write16(addr, rt)
	default:
		c.Halt("unimplemented STREXD at 0x%08X", c.cur)
		return
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

// execParallel runs the byte/halfword parallel add and subtract. The plain
// signed/unsigned forms (classes S and U) are modelled, including the GE flags
// they set for a following SEL; the saturating (Q/UQ) and halving (SH/UH) and
// the exchange (ASX/SAX) variants Halt, so a program that needs them fails
// explicitly rather than computing the wrong thing.
func (c *CPU) execParallel(w uint32) bool {
	class := (w >> 20) & 7 // 1 S, 2 Q, 3 SH, 5 U, 6 UQ, 7 UH
	op2 := (w >> 5) & 7
	rn := c.reg((w >> 16) & 0xF)
	rm := c.reg(w & 0xF)
	rd := (w >> 12) & 0xF

	if class != 1 && class != 5 { // only plain S and U are modelled
		c.Halt("unimplemented parallel arithmetic 0x%08X (class %d) at 0x%08X", w, class, c.cur)
		return true
	}
	signed := class == 1

	var res uint32
	var ge uint32
	switch op2 {
	case 0b000: // ADD16
		res, ge = par16(rn, rm, signed, addOp)
	case 0b011: // SUB16
		res, ge = par16(rn, rm, signed, subOp)
	case 0b100: // ADD8
		res, ge = par8(rn, rm, signed, addOp)
	case 0b111: // SUB8
		res, ge = par8(rn, rm, signed, subOp)
	default:
		c.Halt("unimplemented parallel op2=%d at 0x%08X", op2, c.cur)
		return true
	}
	c.setReg(rd, res)
	c.GE = ge
	return true
}

type binOp int

const (
	addOp binOp = iota
	subOp
)

// par16 computes a two-lane 16-bit parallel add or subtract and the 4-bit GE
// value (two bits repeated per lane). GE marks, per lane, "no borrow" for a
// subtract or "carry/overflow-free non-negative" for an add — the exact rule the
// architecture uses so SEL picks the right bytes.
func par16(a, b uint32, signed bool, op binOp) (uint32, uint32) {
	var res, ge uint32
	for lane := 0; lane < 2; lane++ {
		sh := uint(lane * 16)
		av := (a >> sh) & 0xFFFF
		bv := (b >> sh) & 0xFFFF
		var lres uint32
		var setGE bool
		if op == addOp {
			var sum int32
			if signed {
				sum = int16v(av) + int16v(bv)
			} else {
				sum = int32(av) + int32(bv)
			}
			lres = uint32(sum) & 0xFFFF
			setGE = geAdd(av, bv, signed)
		} else {
			var diff int32
			if signed {
				diff = int16v(av) - int16v(bv)
			} else {
				diff = int32(av) - int32(bv)
			}
			lres = uint32(diff) & 0xFFFF
			setGE = geSub(av, bv, signed)
		}
		res |= lres << sh
		if setGE {
			ge |= 0x3 << uint(lane*2)
		}
	}
	return res, ge
}

func par8(a, b uint32, signed bool, op binOp) (uint32, uint32) {
	var res, ge uint32
	for lane := 0; lane < 4; lane++ {
		sh := uint(lane * 8)
		av := (a >> sh) & 0xFF
		bv := (b >> sh) & 0xFF
		var lres uint32
		var setGE bool
		if op == addOp {
			var sum int32
			if signed {
				sum = int8v(av) + int8v(bv)
			} else {
				sum = int32(av) + int32(bv)
			}
			lres = uint32(sum) & 0xFF
			setGE = geAdd8(av, bv, signed)
		} else {
			var diff int32
			if signed {
				diff = int8v(av) - int8v(bv)
			} else {
				diff = int32(av) - int32(bv)
			}
			lres = uint32(diff) & 0xFF
			setGE = geSub8(av, bv, signed)
		}
		res |= lres << sh
		if setGE {
			ge |= 1 << uint(lane)
		}
	}
	return res, ge
}

func int8v(v uint32) int32  { return int32(int8(byte(v))) }
func int16v(v uint32) int32 { return int32(int16(uint16(v))) }

// GE rules per the ARM ARM: for unsigned add, GE set if the sum did not carry
// out of the lane width... actually "sum >= 2^n" sets GE (result is >= the
// modulus, i.e. an unsigned carry). For unsigned sub, GE set if a >= b (no
// borrow). For signed add/sub, GE set if the true result is >= 0.
func geAdd(a, b uint32, signed bool) bool {
	if signed {
		return int16v(a)+int16v(b) >= 0
	}
	return a+b >= 0x10000
}
func geSub(a, b uint32, signed bool) bool {
	if signed {
		return int16v(a)-int16v(b) >= 0
	}
	return a >= b
}
func geAdd8(a, b uint32, signed bool) bool {
	if signed {
		return int8v(a)+int8v(b) >= 0
	}
	return a+b >= 0x100
}
func geSub8(a, b uint32, signed bool) bool {
	if signed {
		return int8v(a)-int8v(b) >= 0
	}
	return a >= b
}

// execMediaPack runs PKH, (U)SAT, the extends, SEL and the byte-reversals.
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
			res = (rnv & 0xFFFF0000) | ((uint32(int32(rmv)>>imm) & 0xFFFF))
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
