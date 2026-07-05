package x86

// This file holds the ModR/M "group" opcodes, the shift/rotate engine, the
// REP string operations, the multiply/divide unit, and the executed subset of
// the 0x0F two-byte page.

// grp1 executes the immediate ALU group (0x80–0x83).
func (c *CPU) grp1(op byte) {
	reg, o := c.modrmE()
	var w int
	var imm uint32
	switch op {
	case 0x80, 0x82:
		w = 1
		imm = c.fetch8()
	case 0x81:
		w = c.osz()
		imm = c.fetchImm()
	default: // 0x83: imm8 sign-extended to the operand size
		w = c.osz()
		imm = signExtByte(c.fetch8()) & widthMask(w)
	}
	res, wr := c.alu(int(reg), c.rEA(o, w), imm, w)
	if wr {
		c.wEA(o, w, res)
	}
}

// grp2 executes a rotate/shift (0xC0/0xC1/0xD0-0xD3); count() supplies the
// shift amount, evaluated after the ModR/M byte is consumed.
func (c *CPU) grp2(w int, count func() uint32) {
	reg, o := c.modrmE()
	cnt := count()
	c.wEA(o, w, c.shiftOp(int(reg), c.rEA(o, w), cnt, w))
}

// shiftOp performs shift/rotate idx (0..7 = ROL ROR RCL RCR SHL SHR SAL SAR)
// on val by cnt at width w, sets the affected flags, and returns the result.
func (c *CPU) shiftOp(idx int, val, cnt uint32, w int) uint32 {
	m, sb, bits := widthMask(w), signMask(w), uint32(8*w)
	cnt &= 0x1F
	val &= m
	if cnt == 0 {
		return val
	}
	switch idx {
	case 4, 6: // SHL / SAL
		res := (val << cnt) & m
		if cnt <= bits {
			c.CF = (val>>(bits-cnt))&1 != 0
		} else {
			c.CF = false
		}
		if cnt == 1 {
			c.OF = (res&sb != 0) != c.CF
		}
		c.setSZP(res, w)
		return res
	case 5: // SHR
		c.CF = (val>>(cnt-1))&1 != 0
		res := (val >> cnt) & m
		if cnt == 1 {
			c.OF = val&sb != 0
		}
		c.setSZP(res, w)
		return res
	case 7: // SAR
		sv := signExtToInt(val, w)
		c.CF = (uint32(sv)>>(cnt-1))&1 != 0
		res := uint32(sv>>cnt) & m
		if cnt == 1 {
			c.OF = false
		}
		c.setSZP(res, w)
		return res
	case 0: // ROL
		n := cnt % bits
		res := val
		if n != 0 {
			res = ((val << n) | (val >> (bits - n))) & m
		}
		c.CF = res&1 != 0
		if cnt == 1 {
			c.OF = (res&sb != 0) != c.CF
		}
		return res
	case 1: // ROR
		n := cnt % bits
		res := val
		if n != 0 {
			res = ((val >> n) | (val << (bits - n))) & m
		}
		c.CF = res&sb != 0
		if cnt == 1 {
			c.OF = (res&sb != 0) != (res&(sb>>1) != 0)
		}
		return res
	case 2: // RCL
		res, cf := val, b2u(c.CF)
		for i := uint32(0); i < cnt; i++ {
			nc := (res >> (bits - 1)) & 1
			res = ((res << 1) | cf) & m
			cf = nc
		}
		c.CF = cf != 0
		if cnt == 1 {
			c.OF = (res&sb != 0) != c.CF
		}
		return res
	default: // 3: RCR
		res, cf := val, b2u(c.CF)
		for i := uint32(0); i < cnt; i++ {
			nc := res & 1
			res = (res >> 1) | (cf << (bits - 1))
			cf = nc
		}
		res &= m
		c.CF = cf != 0
		if cnt == 1 {
			c.OF = (res&sb != 0) != (res&(sb>>1) != 0)
		}
		return res
	}
}

// grp3 executes 0xF6/0xF7: TEST/NOT/NEG/MUL/IMUL/DIV/IDIV.
func (c *CPU) grp3(w int) {
	reg, o := c.modrmE()
	switch reg {
	case 0, 1: // TEST r/m, imm
		var imm uint32
		if w == 1 {
			imm = c.fetch8()
		} else {
			imm = c.fetchImm()
		}
		c.flagsLogic(c.rEA(o, w)&imm, w)
	case 2: // NOT
		c.wEA(o, w, ^c.rEA(o, w)&widthMask(w))
	case 3: // NEG
		v := c.rEA(o, w)
		res := c.flagsSub(0, v, 0, w)
		c.CF = v&widthMask(w) != 0
		c.wEA(o, w, res)
	case 4:
		c.mulOp(o, w, false)
	case 5:
		c.mulOp(o, w, true)
	case 6:
		c.divOp(o, w, false)
	case 7:
		c.divOp(o, w, true)
	}
}

// grp4/grp5 (0xFE/0xFF) live in exec.go alongside the dispatch (grp4 inline).

// grp5 executes 0xFF: INC/DEC/CALL/CALLF/JMP/JMPF/PUSH on an r/m operand.
func (c *CPU) grp5() {
	reg, o := c.modrmE()
	w := c.osz()
	switch reg {
	case 0:
		c.wEA(o, w, c.incDec(c.rEA(o, w), w, true))
	case 1:
		c.wEA(o, w, c.incDec(c.rEA(o, w), w, false))
	case 2: // CALL near indirect
		t := c.rEA(o, w) & 0xFFFF
		c.push16(c.IP)
		c.IP = t
	case 3: // CALLF m16:16
		off := c.rEA(o, 2)
		seg := c.memRead(o.seg, o.off+2, 2)
		c.push16(uint32(c.Seg[CS]))
		c.push16(c.IP)
		c.Seg[CS] = uint16(seg)
		c.IP = off & 0xFFFF
	case 4: // JMP near indirect
		c.IP = c.rEA(o, w) & 0xFFFF
	case 5: // JMPF m16:16
		off := c.rEA(o, 2)
		seg := c.memRead(o.seg, o.off+2, 2)
		c.Seg[CS] = uint16(seg)
		c.IP = off & 0xFFFF
	case 6: // PUSH r/m
		c.push(w, c.rEA(o, w))
	default:
		c.Halt("grp5 /7 (invalid) at %04X:%04X", c.Seg[CS], c.IP)
	}
}

// mulOp executes MUL/IMUL (one-operand form) at width w.
func (c *CPU) mulOp(o ea, w int, signed bool) {
	src := c.rEA(o, w)
	switch w {
	case 1:
		var p uint32
		if signed {
			p = uint32(int32(int8(byte(c.g8(0)))) * int32(int8(byte(src))))
		} else {
			p = (c.g8(0) & 0xFF) * (src & 0xFF)
		}
		c.s16(AX, p&0xFFFF)
		if signed {
			c.CF = int32(int16(uint16(p))) != int32(int8(byte(p)))
		} else {
			c.CF = (p>>8)&0xFF != 0
		}
		c.OF = c.CF
	case 2:
		var p uint32
		if signed {
			p = uint32(int32(int16(uint16(c.gw(AX)))) * int32(int16(uint16(src))))
		} else {
			p = c.gw(AX) * (src & 0xFFFF)
		}
		c.s16(AX, p&0xFFFF)
		c.s16(DX, (p>>16)&0xFFFF)
		if signed {
			c.CF = int32(p) != int32(int16(uint16(p)))
		} else {
			c.CF = p>>16 != 0
		}
		c.OF = c.CF
	default: // 4
		var p uint64
		if signed {
			p = uint64(int64(int32(c.Regs[AX])) * int64(int32(src)))
		} else {
			p = uint64(c.Regs[AX]) * uint64(src)
		}
		c.Regs[AX] = uint32(p)
		c.Regs[DX] = uint32(p >> 32)
		if signed {
			c.CF = int64(p) != int64(int32(uint32(p)))
		} else {
			c.CF = p>>32 != 0
		}
		c.OF = c.CF
	}
}

// divOp executes DIV/IDIV (one-operand form) at width w. A zero divisor or an
// overflowing quotient raises the divide-error exception (#DE / INT 0) through
// IVT[0] — it does NOT stop the core. This matters: fixed-point renderers (UW's
// among them) install their own #DE handler and deliberately let IDIV overflow,
// having the handler saturate the result. divErr returns true when it raised the
// exception, so the caller skips writing a result.
func (c *CPU) divOp(o ea, w int, signed bool) {
	src := c.rEA(o, w)
	switch w {
	case 1:
		if src&0xFF == 0 {
			c.divErr()
			return
		}
		if signed {
			num := int16(uint16(c.gw(AX)))
			d := int16(int8(byte(src)))
			q, r := num/d, num%d
			if q < -128 || q > 127 {
				c.divErr()
				return
			}
			c.s8(0, uint32(byte(int8(q))))
			c.s8(4, uint32(byte(int8(r))))
		} else {
			num := c.gw(AX)
			d := src & 0xFF
			q, r := num/d, num%d
			if q > 0xFF {
				c.divErr()
				return
			}
			c.s8(0, q)
			c.s8(4, r)
		}
	case 2:
		if src&0xFFFF == 0 {
			c.divErr()
			return
		}
		num := c.gw(DX)<<16 | c.gw(AX)
		if signed {
			n := int32(num)
			d := int32(int16(uint16(src)))
			q, r := n/d, n%d
			if q < -32768 || q > 32767 {
				c.divErr()
				return
			}
			c.s16(AX, uint32(uint16(int16(q))))
			c.s16(DX, uint32(uint16(int16(r))))
		} else {
			d := src & 0xFFFF
			q, r := num/d, num%d
			if q > 0xFFFF {
				c.divErr()
				return
			}
			c.s16(AX, q)
			c.s16(DX, r)
		}
	default: // 4
		if src == 0 {
			c.divErr()
			return
		}
		num := uint64(c.Regs[DX])<<32 | uint64(c.Regs[AX])
		if signed {
			n := int64(num)
			d := int64(int32(src))
			c.Regs[AX] = uint32(n / d)
			c.Regs[DX] = uint32(n % d)
		} else {
			d := uint64(src)
			c.Regs[AX] = uint32(num / d)
			c.Regs[DX] = uint32(num % d)
		}
	}
}

// imulTrunc is the two-/three-operand IMUL result: a signed product truncated
// to width w, with CF=OF set when significant bits were lost.
func (c *CPU) imulTrunc(a, b uint32, w int) uint32 {
	var full int64
	if w == 2 {
		full = int64(int16(uint16(a))) * int64(int16(uint16(b)))
	} else {
		full = int64(int32(a)) * int64(int32(b))
	}
	res := uint32(full) & widthMask(w)
	c.CF = full != int64(signExtToInt(res, w))
	c.OF = c.CF
	return res
}

// stringOp executes one string primitive, honouring a REP/REPE/REPNE prefix.
func (c *CPU) stringOp(op, rep byte) {
	w := 1
	switch op {
	case 0xA5, 0xA7, 0xAB, 0xAD, 0xAF:
		w = c.osz()
	}
	step := func() {
		switch op {
		case 0xA4, 0xA5: // MOVS  ES:DI <- DS:SI
			c.memWrite(c.Seg[ES], c.gw(DI), w, c.memRead(c.segFor(DS), c.gw(SI), w))
			c.advSI(w)
			c.advDI(w)
		case 0xA6, 0xA7: // CMPS  DS:SI ? ES:DI
			c.flagsSub(c.memRead(c.segFor(DS), c.gw(SI), w), c.memRead(c.Seg[ES], c.gw(DI), w), 0, w)
			c.advSI(w)
			c.advDI(w)
		case 0xAA, 0xAB: // STOS  ES:DI <- eAX
			c.memWrite(c.Seg[ES], c.gw(DI), w, c.getReg(AX, w))
			c.advDI(w)
		case 0xAC, 0xAD: // LODS  eAX <- DS:SI
			c.setReg(AX, w, c.memRead(c.segFor(DS), c.gw(SI), w))
			c.advSI(w)
		case 0xAE, 0xAF: // SCAS  eAX ? ES:DI
			c.flagsSub(c.getReg(AX, w), c.memRead(c.Seg[ES], c.gw(DI), w), 0, w)
			c.advDI(w)
		}
	}
	if rep == 0 {
		step()
		return
	}
	isCmp := op == 0xA6 || op == 0xA7 || op == 0xAE || op == 0xAF
	for c.gw(CX) != 0 {
		step()
		c.sw(CX, c.gw(CX)-1)
		if isCmp {
			if rep == 0xF3 && !c.ZF { // REPE: stop when a mismatch appears
				break
			}
			if rep == 0xF2 && c.ZF { // REPNE: stop when a match appears
				break
			}
		}
	}
}

// stringPortOp executes one INS/OUTS primitive, honouring a REP prefix.
func (c *CPU) stringPortOp(op, rep byte) {
	w := 1
	if op == 0x6D || op == 0x6F {
		w = c.osz()
	}
	port := c.Reg16(DX)
	step := func() {
		switch op {
		case 0x6C, 0x6D: // INS: ES:DI <- port[DX]
			c.memWrite(c.Seg[ES], c.gw(DI), w, c.inPort(port, w))
			c.advDI(w)
		default: // 0x6E,0x6F OUTS: port[DX] <- DS:SI
			c.outPort(port, w, c.memRead(c.segFor(DS), c.gw(SI), w))
			c.advSI(w)
		}
	}
	if rep == 0 {
		step()
		return
	}
	for c.gw(CX) != 0 {
		step()
		c.sw(CX, c.gw(CX)-1)
	}
}

func (c *CPU) advSI(w int) {
	if c.DF {
		c.sw(SI, c.gw(SI)-uint32(w))
	} else {
		c.sw(SI, c.gw(SI)+uint32(w))
	}
}
func (c *CPU) advDI(w int) {
	if c.DF {
		c.sw(DI, c.gw(DI)-uint32(w))
	} else {
		c.sw(DI, c.gw(DI)+uint32(w))
	}
}

// daa/das (0x27/0x2F): decimal-adjust AL after add/subtract.
func (c *CPU) daa(sub bool) {
	al := c.g8(0)
	oldAL, oldCF := al, c.CF
	c.CF = false
	if al&0x0F > 9 || c.AF {
		if sub {
			al -= 6
		} else {
			al += 6
		}
		c.AF = true
	} else {
		c.AF = false
	}
	if oldAL > 0x99 || oldCF {
		if sub {
			al -= 0x60
		} else {
			al += 0x60
		}
		c.CF = true
	}
	al &= 0xFF
	c.s8(0, al)
	c.setSZP(al, 1)
}

// aaa/aas (0x37/0x3F): ASCII-adjust AL after add/subtract.
func (c *CPU) aaa(sub bool) {
	al, ah := c.g8(0), c.g8(4)
	if al&0x0F > 9 || c.AF {
		if sub {
			al -= 6
			ah--
		} else {
			al += 6
			ah++
		}
		c.AF, c.CF = true, true
	} else {
		c.AF, c.CF = false, false
	}
	c.s8(0, al&0x0F)
	c.s8(4, ah&0xFF)
}

// aam/aad (0xD4/0xD5): ASCII-adjust for multiply/divide (base usually 10).
func (c *CPU) aam(base byte) {
	if base == 0 {
		c.Halt("AAM by zero at %04X:%04X", c.Seg[CS], c.IP)
		return
	}
	al := c.g8(0)
	c.s8(4, al/uint32(base))
	c.s8(0, al%uint32(base))
	c.setSZP(c.g8(0), 1)
}
func (c *CPU) aad(base byte) {
	res := (c.g8(0) + c.g8(4)*uint32(base)) & 0xFF
	c.s8(0, res)
	c.s8(4, 0)
	c.setSZP(res, 1)
}

// exec0F executes the implemented subset of the 0x0F two-byte page.
func (c *CPU) exec0F(op byte) {
	switch {
	case op >= 0x80 && op <= 0x8F: // Jcc near
		rel := c.fetchImm()
		if c.dOpsize != 32 {
			rel = signExtWord(rel)
		}
		if c.cond(op & 0x0F) {
			c.IP = (c.IP + rel) & 0xFFFF
		}
		return
	case op >= 0x90 && op <= 0x9F: // SETcc r/m8
		_, o := c.modrmE()
		c.wEA(o, 1, b2u(c.cond(op&0x0F)))
		return
	}
	switch op {
	case 0xA0:
		c.push16(uint32(c.Seg[FS]))
	case 0xA1:
		c.Seg[FS] = uint16(c.pop16())
	case 0xA8:
		c.push16(uint32(c.Seg[GS]))
	case 0xA9:
		c.Seg[GS] = uint16(c.pop16())
	case 0xAF: // IMUL r, r/m
		reg, o := c.modrmE()
		w := c.osz()
		c.setReg(reg, w, c.imulTrunc(c.getReg(reg, w), c.rEA(o, w), w))
	case 0xB6: // MOVZX r, r/m8
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), c.rEA(o, 1)&0xFF)
	case 0xB7: // MOVZX r, r/m16
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), c.rEA(o, 2)&0xFFFF)
	case 0xBE: // MOVSX r, r/m8
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), signExtByte(c.rEA(o, 1)))
	case 0xBF: // MOVSX r, r/m16
		reg, o := c.modrmE()
		c.setReg(reg, c.osz(), signExtWord(c.rEA(o, 2)))
	default:
		c.Halt("unimplemented 0F opcode $%02X at %04X:%04X", op, c.Seg[CS], c.IP)
	}
}
