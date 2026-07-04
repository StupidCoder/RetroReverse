package arm

import "math/bits"

// Step executes one instruction (servicing the current CPU state) and returns an
// approximate cycle count. Timing is only nominal — enough for an oracle to budget
// a run by "cycles"; it is not cycle-accurate (that needs the memory-system wait
// states this core deliberately omits). Inspect Instrs for an exact instruction
// count. A Halted core returns 0.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	c.cur = c.R[15]
	c.branched = false
	c.Instrs++
	if c.Thumb {
		return c.stepThumb()
	}
	return c.stepARM()
}

func (c *CPU) boolToU(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// stepARM fetches, condition-checks and executes one ARM instruction.
func (c *CPU) stepARM() int {
	w := c.read32aligned(c.cur)
	cond := int(w >> 28)
	if cond == condNV {
		return c.execUncondARM(w)
	}
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

func (c *CPU) execUncondARM(w uint32) int {
	if (w>>25)&7 == 0b101 { // BLX <imm> → switch to Thumb
		h := (w >> 24) & 1
		off := signExtend(w&0xFFFFFF, 24)<<2 | h<<1
		c.R[14] = c.cur + 4
		c.Thumb = true
		c.R[15] = (c.cur + 8 + off) &^ 1
		c.branched = true
		return 3
	}
	// PLD and other hints: no architectural effect on this core.
	c.R[15] = c.cur + 4
	return 1
}

// execARM dispatches a condition-passed ARM instruction.
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
			c.Halt("undefined media instruction 0x%08X at 0x%08X", w, c.cur)
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

// dpOperand computes operand 2 of a data-processing instruction and its shifter
// carry-out.
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
	if (w>>4)&1 == 1 { // register-specified shift amount
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

	sn, sz, sc, sv := c.N, c.Z, c.C, c.V // snapshot to restore when flags aren't set
	var res uint32
	logical := true // whether flags come from the shifter (vs the adder)
	switch op {
	case 0b0000: // AND
		res = a & op2
	case 0b0001: // EOR
		res = a ^ op2
	case 0b0010: // SUB
		res, logical = c.sub(a, op2, 1), false
	case 0b0011: // RSB
		res, logical = c.sub(op2, a, 1), false
	case 0b0100: // ADD
		res, logical = c.add(a, op2, 0), false
	case 0b0101: // ADC
		res, logical = c.add(a, op2, c.boolToU(sc)), false
	case 0b0110: // SBC
		res, logical = c.sub(a, op2, c.boolToU(sc)), false
	case 0b0111: // RSC
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
	case 0b1100: // ORR
		res = a | op2
	case 0b1101: // MOV
		res = op2
	case 0b1110: // BIC
		res = a &^ op2
	default: // 0b1111: MVN
		res = ^op2
	}

	isTest := op >= 0b1000 && op <= 0b1011
	if isTest {
		// Test ops always update flags. Logical tests take C from the shifter and
		// leave V unchanged; the arithmetic tests (CMP/CMN) already set NZCV.
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
		// arithmetic ops already set NZCV via add/sub
	} else {
		c.N, c.Z, c.C, c.V = sn, sz, sc, sv // no flag change
	}

	if rd == 15 {
		if setFlags { // exception return: restore CPSR from SPSR
			c.SetCPSR(c.SPSR())
		}
		c.branchTo(res)
		return
	}
	c.setReg(rd, res)
}

// branchTo writes the PC, word/halfword-aligning for the current state.
func (c *CPU) branchTo(v uint32) {
	if c.Thumb {
		c.R[15] = v &^ 1
	} else {
		c.R[15] = v &^ 3
	}
	c.branched = true
}

// bxTo writes the PC and selects ARM/Thumb state from bit 0 of the target (the
// BX/BLX and ARMv5 LDR/LDM-to-PC interworking rule).
func (c *CPU) bxTo(v uint32) {
	c.Thumb = v&1 == 1
	c.branchTo(v)
}

func (c *CPU) execBranch(w uint32) {
	off := signExtend(w&0xFFFFFF, 24) << 2
	target := c.cur + 8 + off
	if (w>>24)&1 == 1 { // BL
		c.R[14] = c.cur + 4
	}
	c.branchTo(target)
}

// execExtension runs the multiply / swap / halfword-transfer space.
func (c *CPU) execExtension(w uint32) {
	sh := (w >> 5) & 3
	if sh == 0 {
		switch {
		case (w>>24)&1 == 1:
			c.execSwap(w)
		case (w>>23)&1 == 1:
			c.execMulLong(w)
		default:
			c.execMul(w)
		}
		return
	}
	c.execHalf(w, sh)
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

func (c *CPU) execHalf(w uint32, sh uint32) {
	l := (w >> 20) & 1
	p := (w >> 24) & 1
	u := (w >> 23) & 1
	wbit := (w >> 21) & 1
	rn := (w >> 16) & 0xF
	rd := (w >> 12) & 0xF
	base := c.reg(rn)

	var off uint32
	if (w>>22)&1 == 1 { // immediate
		off = (w>>4)&0xF0 | (w & 0xF)
	} else {
		off = c.reg(w & 0xF)
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
		switch sh {
		case 1: // LDRH
			v = c.read16(addr)
		case 2: // LDRSB
			v = uint32(int32(int8(c.read8(addr))))
		default: // LDRSH
			v = uint32(int32(int16(c.read16(addr))))
		}
		writeback()
		c.setReg(rd, v)
	} else { // STRH
		writeback()
		c.write16(addr, c.reg(rd))
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
			c.bxTo(v) // LDR into PC interworks on ARMv5
		} else {
			c.setReg(rd, v)
		}
	} else {
		v := c.reg(rd)
		if rd == 15 {
			v = c.cur + 12 // storing PC stores the current instruction + 12
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

	// The lowest-numbered register always uses the lowest address; compute that
	// start address and the final write-back value for the four modes.
	var start, final uint32
	if u == 1 { // increment
		final = base + n*4
		if p == 1 { // IB
			start = base + 4
		} else { // IA
			start = base
		}
	} else { // decrement
		final = base - n*4
		if p == 1 { // DB
			start = base - n*4
		} else { // DA
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
				if s == 1 { // restore CPSR from SPSR (exception return form)
					c.SetCPSR(c.SPSR())
				}
				c.bxTo(v)
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
	if wbit == 1 && !(l == 1 && loadedRn) { // a loaded base isn't overwritten by write-back
		c.setReg(rn, final)
	}
}
