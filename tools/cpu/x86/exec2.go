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
		t := c.rEA(o, w)
		c.push(c.osz(), c.IP)
		c.IP = c.ipMask(t)
	case 3: // CALLF m16:16 / m16:32
		off := c.rEA(o, w)
		seg := c.memRead(o.base, o.off+uint32(w), 2)
		c.push(c.osz(), uint32(c.Seg[CS]))
		c.push(c.osz(), c.IP)
		c.loadSeg(CS, uint16(seg))
		c.IP = c.ipMask(off)
	case 4: // JMP near indirect
		c.IP = c.ipMask(c.rEA(o, w))
	case 5: // JMPF m16:16 / m16:32
		off := c.rEA(o, w)
		seg := c.memRead(o.base, o.off+uint32(w), 2)
		c.loadSeg(CS, uint16(seg))
		c.IP = c.ipMask(off)
	case 6: // PUSH r/m
		c.push(w, c.rEA(o, w))
	default:
		c.Halt("grp5 /7 (invalid) at %s", c.at())
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
// The operand width w (1/2/4) is the data size; the address width aw (2/4) is
// the size of the SI/DI pointers and the CX counter — so REP STOSD (w=4, aw=4
// in 32-bit code) walks ESI/EDI/ECX, the very first thing a go32 crt0 does.
func (c *CPU) stringOp(op, rep byte) {
	w := 1
	switch op {
	case 0xA5, 0xA7, 0xAB, 0xAD, 0xAF:
		w = c.osz()
	}
	aw := c.asz()
	dsBase := c.segBaseFor(DS)
	esBase := c.segBase(ES)
	step := func() {
		switch op {
		case 0xA4, 0xA5: // MOVS  ES:DI <- DS:SI
			c.memWrite(esBase, c.getReg(DI, aw), w, c.memRead(dsBase, c.getReg(SI, aw), w))
			c.advSI(w, aw)
			c.advDI(w, aw)
		case 0xA6, 0xA7: // CMPS  DS:SI ? ES:DI
			c.flagsSub(c.memRead(dsBase, c.getReg(SI, aw), w), c.memRead(esBase, c.getReg(DI, aw), w), 0, w)
			c.advSI(w, aw)
			c.advDI(w, aw)
		case 0xAA, 0xAB: // STOS  ES:DI <- eAX
			c.memWrite(esBase, c.getReg(DI, aw), w, c.getReg(AX, w))
			c.advDI(w, aw)
		case 0xAC, 0xAD: // LODS  eAX <- DS:SI
			c.setReg(AX, w, c.memRead(dsBase, c.getReg(SI, aw), w))
			c.advSI(w, aw)
		case 0xAE, 0xAF: // SCAS  eAX ? ES:DI
			c.flagsSub(c.getReg(AX, w), c.memRead(esBase, c.getReg(DI, aw), w), 0, w)
			c.advDI(w, aw)
		}
	}
	if rep == 0 {
		step()
		return
	}
	isCmp := op == 0xA6 || op == 0xA7 || op == 0xAE || op == 0xAF
	for c.getReg(CX, aw) != 0 {
		step()
		c.setReg(CX, aw, c.getReg(CX, aw)-1)
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
	aw := c.asz()
	dsBase := c.segBaseFor(DS)
	esBase := c.segBase(ES)
	port := c.Reg16(DX)
	step := func() {
		switch op {
		case 0x6C, 0x6D: // INS: ES:DI <- port[DX]
			c.memWrite(esBase, c.getReg(DI, aw), w, c.inPort(port, w))
			c.advDI(w, aw)
		default: // 0x6E,0x6F OUTS: port[DX] <- DS:SI
			c.outPort(port, w, c.memRead(dsBase, c.getReg(SI, aw), w))
			c.advSI(w, aw)
		}
	}
	if rep == 0 {
		step()
		return
	}
	for c.getReg(CX, aw) != 0 {
		step()
		c.setReg(CX, aw, c.getReg(CX, aw)-1)
	}
}

// advSI/advDI step the string index register by the data width w (up or down
// per DF), at address width aw (SI vs ESI).
func (c *CPU) advSI(w, aw int) {
	if c.DF {
		c.setReg(SI, aw, c.getReg(SI, aw)-uint32(w))
	} else {
		c.setReg(SI, aw, c.getReg(SI, aw)+uint32(w))
	}
}
func (c *CPU) advDI(w, aw int) {
	if c.DF {
		c.setReg(DI, aw, c.getReg(DI, aw)-uint32(w))
	} else {
		c.setReg(DI, aw, c.getReg(DI, aw)+uint32(w))
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
		c.Halt("AAM by zero at %s", c.at())
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

// exec0F executes the implemented subset of the 0x0F two-byte page. rep carries any
// 0xF2/0xF3 prefix (a REP byte in the string ops, a mandatory prefix in SSE); a 0x66
// prefix is visible as dOpsize==16.
func (c *CPU) exec0F(op, rep byte) {
	switch {
	case op >= 0x80 && op <= 0x8F: // Jcc near
		rel := c.fetchImm()
		if c.dOpsize != 32 {
			rel = signExtWord(rel)
		}
		if c.cond(op & 0x0F) {
			c.IP = c.ipMask(c.IP + rel)
		}
		return
	case op >= 0x90 && op <= 0x9F: // SETcc r/m8
		_, o := c.modrmE()
		c.wEA(o, 1, b2u(c.cond(op&0x0F)))
		return
	}
	switch op {
	case 0x08, 0x09: // INVD / WBINVD — cache maintenance; no caches modelled, so no-ops.
		// The XDK graphics init flushes the CPU cache (WBINVD) before it kicks the NV2A so
		// the pusher sees coherent push-buffer bytes; with one flat RAM backing the flush
		// is already a no-op for us.
	case 0x31: // RDTSC — EDX:EAX = the time-stamp counter. Modelled as the retired-
		// instruction count scaled by TSCMul, so guest-visible time advances at a rate
		// the host machine chooses consistently with its other clocks (the Xbox sets
		// 367: a 733 MHz TSC against its 2000-instructions-per-millisecond tick).
		mul := c.TSCMul
		if mul == 0 {
			mul = 1
		}
		tsc := c.Steps * mul
		c.Regs[AX] = uint32(tsc)
		c.Regs[DX] = uint32(tsc >> 32)
	case 0xB0, 0xB1: // CMPXCHG r/m, r — compare the accumulator with r/m: equal (ZF=1)
		// stores r into r/m, else (ZF=0) loads r/m into the accumulator. Flags as CMP.
		// The single-core model needs no LOCK semantics. OutRun's own interlocked paths
		// are the first user.
		w := 1
		if op == 0xB1 {
			w = c.osz()
		}
		reg, o := c.modrmE()
		dst := c.rEA(o, w)
		c.flagsSub(c.getReg(0, w), dst, 0, w)
		if c.ZF {
			c.wEA(o, w, c.getReg(reg, w))
		} else {
			c.setReg(0, w, dst)
		}
	case 0xC0, 0xC1: // XADD r/m, r — exchange and add: r/m gets the sum, r the old r/m.
		w := 1
		if op == 0xC1 {
			w = c.osz()
		}
		reg, o := c.modrmE()
		dst := c.rEA(o, w)
		sum := c.flagsAdd(dst, c.getReg(reg, w), 0, w)
		c.setReg(reg, w, dst)
		c.wEA(o, w, sum)
	case 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F:
		// PREFETCH* (0F 18) and the reserved multi-byte NOP group (0F 19-1F, incl. the
		// canonical long-NOP 0F 1F). All carry a ModR/M operand and have no architectural
		// effect but consuming it; decode and discard. The XDK inserts prefetch hints
		// throughout its vertex/matrix code.
		c.modrmE()
	case 0xAE: // group 15: fences / LDMXCSR / STMXCSR / CLFLUSH / FXSAVE / FXRSTOR
		reg, o := c.modrmE()
		switch {
		case o.isReg:
			// mod=11: reg 5 LFENCE, 6 MFENCE, 7 SFENCE — memory ordering barriers, and no-ops
			// on an in-order single-core model. The XDK kickoff SFENCEs before writing PUT.
		case reg == 2: // LDMXCSR m32
			c.MXCSR = c.rEA(o, 4)
		case reg == 3: // STMXCSR m32
			c.wEA(o, 4, c.MXCSR)
		case reg == 7: // CLFLUSH m8 — flush a cache line; no-op with flat RAM
		default:
			c.Halt("unimplemented 0F AE /%d (FXSAVE/FXRSTOR) at %s", reg, c.at())
		}
	case 0xA0:
		c.push(c.osz(), uint32(c.Seg[FS]))
	case 0xA1:
		c.loadSeg(FS, uint16(c.pop(c.osz())))
	case 0xA8:
		c.push(c.osz(), uint32(c.Seg[GS]))
	case 0xA9:
		c.loadSeg(GS, uint16(c.pop(c.osz())))
	case 0xA4: // SHLD r/m, r, imm8
		reg, o := c.modrmE()
		c.doubleShift(true, o, reg, c.fetch8(), c.osz())
	case 0xA5: // SHLD r/m, r, CL
		reg, o := c.modrmE()
		c.doubleShift(true, o, reg, c.g8(1), c.osz())
	case 0xAC: // SHRD r/m, r, imm8
		reg, o := c.modrmE()
		c.doubleShift(false, o, reg, c.fetch8(), c.osz())
	case 0xAD: // SHRD r/m, r, CL
		reg, o := c.modrmE()
		c.doubleShift(false, o, reg, c.g8(1), c.osz())
	case 0xA3: // BT  r/m, r
		reg, o := c.modrmE()
		c.bitOp(o, c.getReg(reg, c.osz()), c.osz(), 0, true)
	case 0xAB: // BTS r/m, r
		reg, o := c.modrmE()
		c.bitOp(o, c.getReg(reg, c.osz()), c.osz(), 1, true)
	case 0xB3: // BTR r/m, r
		reg, o := c.modrmE()
		c.bitOp(o, c.getReg(reg, c.osz()), c.osz(), 2, true)
	case 0xBB: // BTC r/m, r
		reg, o := c.modrmE()
		c.bitOp(o, c.getReg(reg, c.osz()), c.osz(), 3, true)
	case 0xBA: // grp8: BT/BTS/BTR/BTC r/m, imm8 (reg field 4/5/6/7)
		sub, o := c.modrmE()
		imm := c.fetch8()
		c.bitOp(o, imm, c.osz(), int(sub&3), false)
	case 0xBC: // BSF r, r/m
		reg, o := c.modrmE()
		c.bitScan(reg, c.rEA(o, c.osz()), c.osz(), false)
	case 0xBD: // BSR r, r/m
		reg, o := c.modrmE()
		c.bitScan(reg, c.rEA(o, c.osz()), c.osz(), true)
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
		if c.execSSE(op, rep) {
			return
		}
		c.Halt("unimplemented 0F opcode $%02X at %s", op, c.at())
	}
}

// bitOp implements BT/BTS/BTR/BTC (mode 0/1/2/3). CF receives the selected bit;
// BTS/BTR/BTC then set/clear/complement it. With a register destination the bit
// index is taken modulo the operand width. With a memory destination and a
// register-sourced index (memBitString), the index is a *signed* bit-string
// offset that selects a byte away from the operand address — the addressing an
// assembler-level `bt [mem], eax` uses to walk a bitmap. The imm8 form
// (memBitString false) instead operates on the whole operand at its own address.
func (c *CPU) bitOp(o ea, idx uint32, w, mode int, memBitString bool) {
	if o.isReg {
		b := idx % uint32(w*8)
		v := c.getReg(o.reg, w)
		c.CF = (v>>b)&1 != 0
		if mode != 0 {
			c.setReg(o.reg, w, applyBit(v, b, mode))
		}
		return
	}
	base, off, b := o.base, o.off, idx
	bw := w
	if memBitString {
		off += uint32(int32(idx) >> 3) // arithmetic shift: negative offsets walk back
		b = idx & 7
		bw = 1 // a memory bit string is addressed a byte at a time
	} else {
		b = idx % uint32(w*8)
	}
	v := c.memRead(base, off, bw)
	c.CF = (v>>b)&1 != 0
	if mode != 0 {
		c.memWrite(base, off, bw, applyBit(v, b, mode))
	}
}

func applyBit(v, b uint32, mode int) uint32 {
	switch mode {
	case 1: // BTS
		return v | (1 << b)
	case 2: // BTR
		return v &^ (1 << b)
	case 3: // BTC
		return v ^ (1 << b)
	}
	return v
}

// bitScan implements BSF (forward) and BSR (reverse): find the index of the
// lowest / highest set bit of src. ZF is set when src is zero (and the
// destination is left unchanged, matching the architectural "undefined but in
// practice preserved" behaviour); otherwise ZF is cleared and the index is stored.
func (c *CPU) bitScan(reg byte, src uint32, w int, reverse bool) {
	src &= widthMask(w)
	if src == 0 {
		c.ZF = true
		return
	}
	c.ZF = false
	var idx uint32
	if reverse {
		idx = uint32(w*8) - 1
		for src&(1<<idx) == 0 {
			idx--
		}
	} else {
		for src&(1<<idx) == 0 {
			idx++
		}
	}
	c.setReg(reg, w, idx)
}

// doubleShift implements SHLD (left) and SHRD (!left): a double-precision shift of
// the destination r/m by count bit positions, feeding in bits from the source
// register at the vacated end. The count is masked to 5 bits and a zero count is a
// no-op that leaves the flags untouched (as on real hardware). CF takes the last
// bit shifted out of the destination; SF/ZF/PF follow the result; OF is defined
// only for a count of 1 (a sign change), AF is left undefined.
func (c *CPU) doubleShift(left bool, o ea, reg byte, count uint32, w int) {
	count &= 0x1F
	dst := c.rEA(o, w)
	if count == 0 {
		return
	}
	bits := uint(w * 8)
	if count > uint32(bits) {
		count = uint32(bits) // a 16-bit count > 16 is architecturally undefined; clamp
	}
	n := uint(count)
	src := c.getReg(reg, w)
	var res uint32
	if left {
		res = dst << n
		if n < bits {
			res |= src >> (bits - n)
		}
		c.CF = (dst>>(bits-n))&1 != 0
	} else {
		res = dst >> n
		if n < bits {
			res |= src << (bits - n)
		}
		c.CF = (dst>>(n-1))&1 != 0
	}
	res &= widthMask(w)
	if count == 1 {
		c.OF = (res^dst)&signMask(w) != 0
	}
	c.setSZP(res, w)
	c.wEA(o, w, res)
}
