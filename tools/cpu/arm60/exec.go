package arm60

import "math/bits"

// Step executes one instruction and returns a nominal cycle count (1). Timing is
// not modelled; inspect Instrs for an exact instruction count. A Halted core
// returns 0.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	c.cur = c.R[15]
	c.branched = false
	c.Instrs++

	w := c.read32aligned(c.cur)
	cond := int(w >> 28)
	if !c.cond(cond) {
		c.R[15] = c.cur + 4
		return 1
	}
	c.execARM(w)
	if !c.branched {
		c.R[15] = c.cur + 4
	}
	return 1
}

func (c *CPU) execARM(w uint32) {
	switch (w >> 25) & 7 {
	case 0b000:
		c.execDataMisc(w, false)
	case 0b001:
		c.execDataMisc(w, true)
	case 0b010:
		c.execSingle(w, false)
	case 0b011:
		if (w>>4)&1 == 1 {
			c.Halt("undefined instruction 0x%08X at 0x%08X", w, c.cur)
			return
		}
		c.execSingle(w, true)
	case 0b100:
		c.execBlock(w)
	case 0b101:
		c.execBranch(w)
	case 0b110:
		c.Halt("unimplemented coprocessor transfer 0x%08X at 0x%08X", w, c.cur)
	default: // 0b111
		if (w>>24)&1 == 1 {
			c.execSWI(w)
		} else {
			c.execCopro(w)
		}
	}
}

func (c *CPU) execDataMisc(w uint32, immForm bool) {
	op := (w >> 21) & 0xF
	s := (w >> 20) & 1
	if !immForm && (w>>4)&1 == 1 && (w>>7)&1 == 1 {
		c.execExtension(w)
		return
	}
	if s == 0 && op >= 0b1000 && op <= 0b1011 {
		c.execMisc(w, immForm)
		return
	}
	c.execDataProc(w, immForm)
}

func (c *CPU) dpOperand(w uint32, immForm bool) (uint32, uint32) {
	if immForm {
		rot := ((w >> 8) & 0xF) * 2
		v := ror32(w&0xFF, rot)
		if rot == 0 {
			return v, c.boolToU(c.C)
		}
		return v, (v >> 31) & 1
	}
	rm := c.reg(w & 0xF)
	typ := (w >> 5) & 3
	if (w>>4)&1 == 1 {
		return c.shift(typ, c.reg((w>>8)&0xF), rm, true, c.boolToU(c.C))
	}
	return c.shift(typ, (w>>7)&0x1F, rm, false, c.boolToU(c.C))
}

func (c *CPU) execDataProc(w uint32, immForm bool) {
	op := (w >> 21) & 0xF
	setFlags := (w>>20)&1 == 1
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF
	op2, shc := c.dpOperand(w, immForm)
	a := c.reg(rn)

	sn, sz, sc, sv := c.N, c.Z, c.C, c.V
	var res uint32
	logical := true
	switch op {
	case 0b0000:
		res = a & op2
	case 0b0001:
		res = a ^ op2
	case 0b0010:
		res, logical = c.sub(a, op2, 1), false
	case 0b0011:
		res, logical = c.sub(op2, a, 1), false
	case 0b0100:
		res, logical = c.add(a, op2, 0), false
	case 0b0101:
		res, logical = c.add(a, op2, c.boolToU(sc)), false
	case 0b0110:
		res, logical = c.sub(a, op2, c.boolToU(sc)), false
	case 0b0111:
		res, logical = c.sub(op2, a, c.boolToU(sc)), false
	case 0b1000: // TST
		res = a & op2
	case 0b1001: // TEQ
		res = a ^ op2
	case 0b1010: // CMP
		c.sub(a, op2, 1)
		logical = false
	case 0b1011: // CMN
		c.add(a, op2, 0)
		logical = false
	case 0b1100:
		res = a | op2
	case 0b1101:
		res = op2
	case 0b1110:
		res = a &^ op2
	default: // MVN
		res = ^op2
	}

	isTest := op >= 0b1000 && op <= 0b1011
	if isTest {
		if logical {
			c.setNZ(res)
			c.C, c.V = shc == 1, sv
		}
		return
	}
	if setFlags {
		if logical {
			c.setNZ(res)
			c.C, c.V = shc == 1, sv
		}
	} else {
		c.N, c.Z, c.C, c.V = sn, sz, sc, sv
	}

	if rd == 15 {
		if setFlags { // exception return: restore CPSR from SPSR
			c.SetCPSR(c.SPSR())
		}
		c.R[15] = res &^ 3
		c.branched = true
		return
	}
	c.setReg(rd, res)
}

func (c *CPU) execBranch(w uint32) {
	off := signExtend(w&0xFFFFFF, 24) << 2
	target := c.cur + 8 + off
	if (w>>24)&1 == 1 { // BL
		c.R[14] = c.cur + 4
	}
	c.R[15] = target &^ 3
	c.branched = true
}

func (c *CPU) execExtension(w uint32) {
	if (w>>5)&3 != 0 {
		c.Halt("undefined halfword/signed transfer 0x%08X at 0x%08X", w, c.cur)
		return
	}
	switch {
	case (w>>24)&1 == 1:
		c.execSwap(w)
	case (w>>23)&1 == 1:
		c.execMulLong(w)
	default:
		c.execMul(w)
	}
}

func (c *CPU) execMul(w uint32) {
	rd := (w >> 16) & 0xF
	rn := (w >> 12) & 0xF
	rs := (w >> 8) & 0xF
	rm := w & 0xF
	res := c.reg(rm) * c.reg(rs)
	if (w>>21)&1 == 1 { // MLA
		res += c.reg(rn)
	}
	if (w>>20)&1 == 1 {
		c.setNZ(res)
	}
	c.setReg(rd, res)
}

func (c *CPU) execMulLong(w uint32) {
	rdHi := (w >> 16) & 0xF
	rdLo := (w >> 12) & 0xF
	rs := c.reg((w >> 8) & 0xF)
	rm := c.reg(w & 0xF)
	signed := (w>>22)&1 == 1
	accum := (w>>21)&1 == 1

	var result uint64
	if signed {
		result = uint64(int64(int32(rm)) * int64(int32(rs)))
	} else {
		result = uint64(rm) * uint64(rs)
	}
	if accum {
		result += uint64(c.reg(rdHi))<<32 | uint64(c.reg(rdLo))
	}
	if (w>>20)&1 == 1 {
		c.Z = result == 0
		c.N = result&(1<<63) != 0
	}
	c.setReg(rdLo, uint32(result))
	c.setReg(rdHi, uint32(result>>32))
}

func (c *CPU) execSwap(w uint32) {
	rn := c.reg((w >> 16) & 0xF)
	rd := (w >> 12) & 0xF
	rm := w & 0xF
	if (w>>22)&1 == 1 { // SWPB
		old := uint32(c.read8(rn))
		c.write8(rn, byte(c.reg(rm)))
		c.setReg(rd, old)
	} else {
		old := c.read32(rn)
		c.write32(rn, c.reg(rm))
		c.setReg(rd, old)
	}
}

func (c *CPU) execSingle(w uint32, regOff bool) {
	l := (w >> 20) & 1
	b := (w >> 22) & 1
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	wbit := (w >> 21) & 1
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF
	base := c.reg(rn)

	var off uint32
	if !regOff {
		off = w & 0xFFF
	} else {
		off, _ = c.shift((w>>5)&3, (w>>7)&0x1F, c.reg(w&0xF), false, c.boolToU(c.C))
	}
	addr := base
	if p == 1 {
		if u == 1 {
			addr += off
		} else {
			addr -= off
		}
	}
	writeback := func() {
		if p == 0 {
			if u == 1 {
				base += off
			} else {
				base -= off
			}
			c.setReg(rn, base)
		} else if wbit == 1 {
			c.setReg(rn, addr)
		}
	}
	if l == 1 {
		var v uint32
		if b == 1 {
			v = uint32(c.read8(addr))
		} else {
			v = c.read32(addr)
		}
		writeback()
		if rd == 15 {
			c.R[15] = v &^ 3
			c.branched = true
		} else {
			c.setReg(rd, v)
		}
	} else {
		v := c.reg(rd)
		if rd == 15 {
			v = c.cur + 12
		}
		if b == 1 {
			c.write8(addr, byte(v))
		} else {
			c.write32(addr, v)
		}
		writeback()
	}
}

func (c *CPU) execBlock(w uint32) {
	l := (w >> 20) & 1
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	wbit := (w >> 21) & 1
	s := (w >> 22) & 1
	rn := (w >> 16) & 0xF
	mask := w & 0xFFFF
	base := c.R[rn]
	n := uint32(bits.OnesCount32(mask))

	var start, final uint32
	if u == 1 {
		final = base + n*4
		if p == 1 {
			start = base + 4
		} else {
			start = base
		}
	} else {
		final = base - n*4
		if p == 1 {
			start = base - n*4
		} else {
			start = base - n*4 + 4
		}
	}

	addr := start
	loadedRn := false
	for i := uint32(0); i < 16; i++ {
		if mask&(1<<i) == 0 {
			continue
		}
		if l == 1 {
			v := c.read32aligned(addr)
			if i == 15 {
				if s == 1 { // exception-return form: restore CPSR from SPSR
					c.SetCPSR(c.SPSR())
				}
				c.R[15] = v &^ 3
				c.branched = true
			} else {
				c.R[i] = v
				if i == rn {
					loadedRn = true
				}
			}
		} else {
			v := c.R[i]
			if i == 15 {
				v = c.cur + 12
			}
			c.write32(addr, v)
		}
		addr += 4
	}
	if wbit == 1 && !(l == 1 && loadedRn) {
		c.setReg(rn, final)
	}
}

func (c *CPU) execMisc(w uint32, immForm bool) {
	op := (w >> 21) & 0xF
	if immForm { // MSR immediate
		c.execMSR(w, true)
		return
	}
	if (w>>4)&0xF == 0 {
		if op == 0b1000 || op == 0b1010 { // MRS
			v := c.CPSR()
			if (w>>22)&1 == 1 {
				v = c.SPSR()
			}
			c.setReg((w>>12)&0xF, v)
		} else { // MSR register
			c.execMSR(w, false)
		}
		return
	}
	c.Halt("undefined instruction 0x%08X at 0x%08X", w, c.cur)
}

func (c *CPU) execMSR(w uint32, immForm bool) {
	var val uint32
	if immForm {
		val = ror32(w&0xFF, ((w>>8)&0xF)*2)
	} else {
		val = c.reg(w & 0xF)
	}
	fields := (w >> 16) & 0xF
	var mask uint32
	if fields&1 != 0 {
		mask |= 0x000000FF
	}
	if fields&2 != 0 {
		mask |= 0x0000FF00
	}
	if fields&4 != 0 {
		mask |= 0x00FF0000
	}
	if fields&8 != 0 {
		mask |= 0xFF000000
	}
	if (w>>22)&1 == 1 { // SPSR
		c.SetSPSR(c.SPSR()&^mask | val&mask)
		return
	}
	if c.Mode == ModeUSR { // user mode: only the flag byte is writable
		mask &= 0xFF000000
	}
	c.SetCPSR(c.CPSR()&^mask | val&mask)
}

func (c *CPU) execSWI(w uint32) {
	if c.SWI != nil && c.SWI(c, w&0xFFFFFF) {
		c.R[15] = c.cur + 4
		c.branched = true
		return
	}
	c.exception(ModeSVC, 0x08, c.cur+4)
}

// exception enters an exception mode: bank registers, save the return address and
// CPSR, disable IRQs, and jump to the vector.
func (c *CPU) exception(mode, vector, returnAddr uint32) {
	saved := c.CPSR()
	c.switchMode(mode)
	c.SetSPSR(saved)
	c.R[14] = returnAddr
	c.IRQDisable = true
	c.R[15] = vector
	c.branched = true
}

// Exception enters an exception at an arbitrary vector — exported so the machine
// model can dispatch a Portfolio-style interrupt to a handler of its choosing.
func (c *CPU) Exception(mode, vector, returnAddr uint32) {
	c.exception(mode, vector, returnAddr)
}

// IRQ raises a normal interrupt if enabled (I bit clear); returns whether taken.
func (c *CPU) IRQ() bool {
	if c.IRQDisable {
		return false
	}
	c.exception(ModeIRQ, 0x18, c.R[15]+4)
	return true
}

// execCopro handles CDP/MCR/MRC. The 3DO's ARM60 has no coprocessors in normal
// use, so with no hook these are ignored (MRC reads zero, MCR is dropped) to keep
// a run from stalling.
func (c *CPU) execCopro(w uint32) {
	if (w>>4)&1 == 0 { // CDP
		return
	}
	load := (w>>20)&1 == 1
	rd := (w >> 12) & 0xF
	if load && rd != 15 {
		c.setReg(rd, 0)
	}
}
