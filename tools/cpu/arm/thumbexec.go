package arm

import "math/bits"

// stepThumb fetches and executes one 16-bit Thumb instruction. The BL/BLX long
// branch is executed as its two natural halfwords (the prefix loads LR, the suffix
// branches) rather than as a fused pair, matching the hardware and keeping each
// Step one halfword.
func (c *CPU) stepThumb() int {
	h := c.read16(c.cur)
	c.execThumb(h)
	if !c.branched {
		c.R[15] = c.cur + 2
	}
	return 1
}

func (c *CPU) execThumb(h uint32) {
	switch {
	case h>>13 == 0b000 && (h>>11)&3 != 0b11: // 1: move shifted register
		typ := (h >> 11) & 3
		res, carry := c.shift(typ, (h>>6)&0x1F, c.R[(h>>3)&7], false, c.boolToU(c.C))
		c.R[h&7] = res
		c.setNZ(res)
		c.C = carry == 1
	case h>>11 == 0b00011: // 2: add/subtract
		c.thumbAddSub(h)
	case h>>13 == 0b001: // 3: mov/cmp/add/sub immediate
		c.thumbImm(h)
	case h>>10 == 0b010000: // 4: ALU operations
		c.thumbALUExec(h)
	case h>>10 == 0b010001: // 5: hi-register operations / BX
		c.thumbHiRegExec(h)
	case h>>11 == 0b01001: // 6: PC-relative load
		addr := ((c.cur + 4) &^ 3) + (h&0xFF)*4
		c.R[(h>>8)&7] = c.read32(addr)
	case h>>12 == 0b0101 && (h>>9)&1 == 0: // 7: load/store with register offset
		c.thumbLoadStoreReg(h)
	case h>>12 == 0b0101: // 8: load/store sign-extended
		c.thumbLoadStoreSExt(h)
	case h>>13 == 0b011: // 9: load/store with immediate offset
		c.thumbLoadStoreImm(h)
	case h>>12 == 0b1000: // 10: load/store halfword
		addr := c.R[(h>>3)&7] + ((h>>6)&0x1F)*2
		if (h>>11)&1 == 1 {
			c.R[h&7] = c.read16(addr)
		} else {
			c.write16(addr, c.R[h&7])
		}
	case h>>12 == 0b1001: // 11: SP-relative load/store
		addr := c.R[13] + (h&0xFF)*4
		if (h>>11)&1 == 1 {
			c.R[(h>>8)&7] = c.read32(addr)
		} else {
			c.write32(addr, c.R[(h>>8)&7])
		}
	case h>>12 == 0b1010: // 12: load address
		off := (h & 0xFF) * 4
		if (h>>11)&1 == 1 {
			c.R[(h>>8)&7] = c.R[13] + off
		} else {
			c.R[(h>>8)&7] = ((c.cur + 4) &^ 3) + off
		}
	case h>>8 == 0b10110000: // 13: add offset to SP
		off := (h & 0x7F) * 4
		if (h>>7)&1 == 1 {
			c.R[13] -= off
		} else {
			c.R[13] += off
		}
	case h>>12 == 0b1011 && (h>>9)&3 == 0b10: // 14: push/pop
		c.thumbPushPop(h)
	case h>>12 == 0b1100: // 15: multiple load/store
		c.thumbBlock(h)
	case h>>12 == 0b1101: // 16: conditional branch / SWI
		c.thumbCondBranch(h)
	case h>>11 == 0b11100: // 18: unconditional branch
		c.branchTo(c.cur + 4 + signExtend(h&0x7FF, 11)<<1)
	case h>>11 == 0b11110: // 19: BL/BLX prefix — load the high part of LR
		c.R[14] = c.cur + 4 + signExtend(h&0x7FF, 11)<<12
	case h>>11 == 0b11111: // 19: BL suffix — stay in Thumb
		target := c.R[14] + (h&0x7FF)<<1
		c.R[14] = (c.cur + 2) | 1
		c.branchTo(target)
	case h>>11 == 0b11101: // 19: BLX suffix — switch to ARM
		target := (c.R[14] + (h&0x7FF)<<1) &^ 3
		c.R[14] = (c.cur + 2) | 1
		c.Thumb = false
		c.R[15] = target
		c.branched = true
	default:
		c.Halt("undefined Thumb 0x%04X at 0x%08X", h, c.cur)
	}
}

func (c *CPU) thumbAddSub(h uint32) {
	rd, rs := h&7, (h>>3)&7
	var operand uint32
	if (h>>10)&1 == 1 { // immediate
		operand = (h >> 6) & 7
	} else {
		operand = c.R[(h>>6)&7]
	}
	if (h>>9)&1 == 1 { // SUB
		c.R[rd] = c.sub(c.R[rs], operand, 1)
	} else { // ADD
		c.R[rd] = c.add(c.R[rs], operand, 0)
	}
}

func (c *CPU) thumbImm(h uint32) {
	rd := (h >> 8) & 7
	imm := h & 0xFF
	switch (h >> 11) & 3 {
	case 0b00: // MOV
		c.R[rd] = imm
		c.setNZ(imm)
	case 0b01: // CMP
		c.sub(c.R[rd], imm, 1)
	case 0b10: // ADD
		c.R[rd] = c.add(c.R[rd], imm, 0)
	default: // SUB
		c.R[rd] = c.sub(c.R[rd], imm, 1)
	}
}

func (c *CPU) thumbALUExec(h uint32) {
	rd, rs := h&7, (h>>3)&7
	a, b := c.R[rd], c.R[rs]
	switch (h >> 6) & 0xF {
	case 0x0: // AND
		c.R[rd] = a & b
		c.setNZ(c.R[rd])
	case 0x1: // EOR
		c.R[rd] = a ^ b
		c.setNZ(c.R[rd])
	case 0x2: // LSL
		res, carry := c.shift(0, b, a, true, c.boolToU(c.C))
		c.R[rd] = res
		c.setNZ(res)
		c.C = carry == 1
	case 0x3: // LSR
		res, carry := c.shift(1, b, a, true, c.boolToU(c.C))
		c.R[rd] = res
		c.setNZ(res)
		c.C = carry == 1
	case 0x4: // ASR
		res, carry := c.shift(2, b, a, true, c.boolToU(c.C))
		c.R[rd] = res
		c.setNZ(res)
		c.C = carry == 1
	case 0x5: // ADC
		c.R[rd] = c.add(a, b, c.boolToU(c.C))
	case 0x6: // SBC
		c.R[rd] = c.sub(a, b, c.boolToU(c.C))
	case 0x7: // ROR
		res, carry := c.shift(3, b, a, true, c.boolToU(c.C))
		c.R[rd] = res
		c.setNZ(res)
		c.C = carry == 1
	case 0x8: // TST
		c.setNZ(a & b)
	case 0x9: // NEG
		c.R[rd] = c.sub(0, b, 1)
	case 0xA: // CMP
		c.sub(a, b, 1)
	case 0xB: // CMN
		c.add(a, b, 0)
	case 0xC: // ORR
		c.R[rd] = a | b
		c.setNZ(c.R[rd])
	case 0xD: // MUL
		c.R[rd] = a * b
		c.setNZ(c.R[rd])
	case 0xE: // BIC
		c.R[rd] = a &^ b
		c.setNZ(c.R[rd])
	default: // 0xF: MVN
		c.R[rd] = ^b
		c.setNZ(c.R[rd])
	}
}

func (c *CPU) thumbHiRegExec(h uint32) {
	rd := (h & 7) | (h>>4)&8
	rm := (h >> 3) & 0xF
	switch (h >> 8) & 3 {
	case 0b00: // ADD (no flags)
		v := c.reg(rd) + c.reg(rm)
		if rd == 15 {
			c.branchTo(v)
		} else {
			c.R[rd] = v
		}
	case 0b01: // CMP
		c.sub(c.reg(rd), c.reg(rm), 1)
	case 0b10: // MOV (no flags)
		v := c.reg(rm)
		if rd == 15 {
			c.branchTo(v)
		} else {
			c.R[rd] = v
		}
	default: // 0b11: BX / BLX
		if (h>>7)&1 == 1 { // BLX
			c.R[14] = (c.cur + 2) | 1
		}
		c.bxTo(c.reg(rm))
	}
}

func (c *CPU) thumbLoadStoreReg(h uint32) {
	addr := c.R[(h>>3)&7] + c.R[(h>>6)&7]
	rd := h & 7
	l, b := (h>>11)&1, (h>>10)&1
	switch {
	case l == 1 && b == 1:
		c.R[rd] = uint32(c.read8(addr))
	case l == 1:
		c.R[rd] = c.read32(addr)
	case b == 1:
		c.write8(addr, byte(c.R[rd]))
	default:
		c.write32(addr, c.R[rd])
	}
}

func (c *CPU) thumbLoadStoreSExt(h uint32) {
	addr := c.R[(h>>3)&7] + c.R[(h>>6)&7]
	rd := h & 7
	switch (h >> 10) & 3 {
	case 0b00: // STRH
		c.write16(addr, c.R[rd])
	case 0b01: // LDSB
		c.R[rd] = uint32(int32(int8(c.read8(addr))))
	case 0b10: // LDRH
		c.R[rd] = c.read16(addr)
	default: // LDSH
		c.R[rd] = uint32(int32(int16(c.read16(addr))))
	}
}

func (c *CPU) thumbLoadStoreImm(h uint32) {
	rd, rb := h&7, (h>>3)&7
	l, b := (h>>11)&1, (h>>12)&1
	off := (h >> 6) & 0x1F
	if b == 0 {
		off *= 4 // word offsets scale by 4
	}
	addr := c.R[rb] + off
	switch {
	case l == 1 && b == 1:
		c.R[rd] = uint32(c.read8(addr))
	case l == 1:
		c.R[rd] = c.read32(addr)
	case b == 1:
		c.write8(addr, byte(c.R[rd]))
	default:
		c.write32(addr, c.R[rd])
	}
}

func (c *CPU) thumbPushPop(h uint32) {
	mask := h & 0xFF
	if (h>>11)&1 == 1 { // POP
		if (h>>8)&1 == 1 {
			mask |= 1 << 15 // include PC
		}
		addr := c.R[13]
		for i := uint32(0); i < 16; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			v := c.read32aligned(addr)
			if i == 15 {
				c.bxTo(v) // POP {…, pc} interworks on ARMv5
			} else {
				c.R[i] = v
			}
			addr += 4
		}
		c.R[13] = addr
	} else { // PUSH
		if (h>>8)&1 == 1 {
			mask |= 1 << 14 // include LR
		}
		addr := c.R[13] - 4*uint32(bits.OnesCount32(mask))
		c.R[13] = addr
		for i := uint32(0); i < 16; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			c.write32(addr, c.R[i])
			addr += 4
		}
	}
}

func (c *CPU) thumbBlock(h uint32) {
	rb := (h >> 8) & 7
	mask := h & 0xFF
	addr := c.R[rb]
	load := (h>>11)&1 == 1
	for i := uint32(0); i < 8; i++ {
		if mask&(1<<i) == 0 {
			continue
		}
		if load {
			c.R[i] = c.read32aligned(addr)
		} else {
			c.write32(addr, c.R[i])
		}
		addr += 4
	}
	// Write-back, unless a load reloaded the base register.
	if !(load && mask&(1<<rb) != 0) {
		c.R[rb] = addr
	}
}

func (c *CPU) thumbCondBranch(h uint32) {
	cond := int((h >> 8) & 0xF)
	if cond == 0b1111 { // SWI
		if c.SWI != nil && c.SWI(c, h&0xFF) {
			return
		}
		c.exception(ModeSVC, 0x08, c.cur+2)
		return
	}
	if cond == 0b1110 {
		c.Halt("undefined Thumb 0x%04X at 0x%08X", h, c.cur)
		return
	}
	if c.cond(cond) {
		c.branchTo(c.cur + 4 + signExtend(h&0xFF, 8)<<1)
	}
}
