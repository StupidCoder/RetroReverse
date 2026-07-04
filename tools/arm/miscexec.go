package arm

import "math/bits"

// execMisc runs the miscellaneous space (MRS/MSR, BX/BLX, CLZ, the saturating
// arithmetic and the signed multiplies).
func (c *CPU) execMisc(w uint32, immForm bool) {
	op := (w >> 21) & 0xF
	if immForm { // MSR immediate
		c.execMSR(w, true)
		return
	}
	if (w>>7)&1 == 1 { // signed multiplies (SMLAxy family)
		c.execSignedMul(w)
		return
	}
	switch (w >> 4) & 0xF {
	case 0b0000:
		if op == 0b1000 || op == 0b1010 { // MRS
			v := c.CPSR()
			if (w>>22)&1 == 1 {
				v = c.SPSR()
			}
			c.setReg((w>>12)&0xF, v)
		} else { // MSR register
			c.execMSR(w, false)
		}
	case 0b0001:
		switch op {
		case 0b1001: // BX
			c.bxTo(c.reg(w & 0xF))
		case 0b1011: // CLZ
			c.setReg((w>>12)&0xF, uint32(bits.LeadingZeros32(c.reg(w&0xF))))
		default:
			c.Halt("unimplemented misc 0x%08X at 0x%08X", w, c.cur)
		}
	case 0b0011: // BLX (register)
		if op == 0b1001 {
			c.R[14] = c.cur + 4
			c.bxTo(c.reg(w & 0xF))
		} else {
			c.Halt("unimplemented misc 0x%08X at 0x%08X", w, c.cur)
		}
	case 0b0101: // QADD / QSUB / QDADD / QDSUB
		c.execSaturating(w)
	case 0b0111: // BKPT
		c.Halt("BKPT #0x%X at 0x%08X", (w>>4)&0xFFF0|(w&0xF), c.cur)
	default:
		c.Halt("unimplemented misc 0x%08X at 0x%08X", w, c.cur)
	}
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
	// In User mode only the flag byte is writable; other modes may write the whole
	// masked word.
	if c.Mode == ModeUSR {
		mask &= 0xFF000000
	}
	c.SetCPSR(c.CPSR()&^mask | val&mask)
}

// qsat saturates a 64-bit signed value to the signed 32-bit range, setting Q on
// overflow.
func (c *CPU) qsat(v int64) uint32 {
	if v > 0x7FFFFFFF {
		c.Q = true
		return 0x7FFFFFFF
	}
	if v < -0x80000000 {
		c.Q = true
		return 0x80000000 // int32 minimum, as its uint32 bit pattern
	}
	return uint32(int32(v))
}

func (c *CPU) execSaturating(w uint32) {
	rn := int64(int32(c.reg((w >> 16) & 0xF)))
	rm := int64(int32(c.reg(w & 0xF)))
	rd := (w >> 12) & 0xF
	switch (w >> 21) & 3 {
	case 0b00: // QADD
		c.setReg(rd, c.qsat(rm+rn))
	case 0b01: // QSUB
		c.setReg(rd, c.qsat(rm-rn))
	case 0b10: // QDADD: Rm + saturate(2*Rn)
		dbl := c.qsat(2 * rn)
		c.setReg(rd, c.qsat(rm+int64(int32(dbl))))
	default: // QDSUB: Rm - saturate(2*Rn)
		dbl := c.qsat(2 * rn)
		c.setReg(rd, c.qsat(rm-int64(int32(dbl))))
	}
}

// half16 selects the bottom (t==0) or top (t==1) signed 16-bit half of v.
func half16(v uint32, t uint32) int64 {
	if t == 1 {
		return int64(int16(v >> 16))
	}
	return int64(int16(v))
}

func (c *CPU) execSignedMul(w uint32) {
	op := (w >> 21) & 0xF
	x := (w >> 5) & 1
	y := (w >> 6) & 1
	rd := (w >> 16) & 0xF
	rn := (w >> 12) & 0xF
	rs := (w >> 8) & 0xF
	rm := w & 0xF
	switch op {
	case 0b1000: // SMLA<x><y>: (Rm.x * Rs.y) + Rn, 32-bit with Q on overflow
		prod := half16(c.reg(rm), x) * half16(c.reg(rs), y)
		acc := prod + int64(int32(c.reg(rn)))
		c.setReg(rd, uint32(acc))
		if acc != int64(int32(acc)) {
			c.Q = true
		}
	case 0b1001: // SMLAW<y> / SMULW<y>
		wide := int64(int32(c.reg(rm))) * half16(c.reg(rs), y)
		res := int32(wide >> 16)
		if x == 0 { // SMLAW
			acc := int64(res) + int64(int32(c.reg(rn)))
			c.setReg(rd, uint32(acc))
			if acc != int64(int32(acc)) {
				c.Q = true
			}
		} else { // SMULW
			c.setReg(rd, uint32(res))
		}
	case 0b1010: // SMLAL<x><y>: 64-bit accumulate into RdHi:RdLo (rn:rd here are lo:hi)
		prod := half16(c.reg(rm), x) * half16(c.reg(rs), y)
		acc := int64(uint64(c.reg(rd))<<32|uint64(c.reg(rn))) + prod
		c.setReg(rn, uint32(acc))             // RdLo
		c.setReg(rd, uint32(uint64(acc)>>32)) // RdHi
	default: // 0b1011: SMUL<x><y>
		c.setReg(rd, uint32(half16(c.reg(rm), x)*half16(c.reg(rs), y)))
	}
}

func (c *CPU) execSWI(w uint32) {
	if c.SWI != nil && c.SWI(c, w&0xFFFFFF) {
		c.R[15] = c.cur + 4
		c.branched = true
		return
	}
	c.exception(ModeSVC, 0x08, c.cur+4)
}

// exception enters an exception mode: bank in the mode's registers, save the return
// address and old CPSR, disable IRQs, switch to ARM state, and jump to the vector.
func (c *CPU) exception(mode, vector, returnAddr uint32) {
	saved := c.CPSR()
	c.switchMode(mode)
	c.SetSPSR(saved)
	c.R[14] = returnAddr
	c.IRQDisable = true
	c.Thumb = false
	c.R[15] = vector
	c.branched = true
}

// IRQ raises a normal interrupt, if enabled (I bit clear). Returns whether it was
// taken. The machine model calls this when its interrupt controller has a pending,
// unmasked request.
func (c *CPU) IRQ() bool {
	if c.IRQDisable {
		return false
	}
	ret := c.R[15]
	if c.Thumb {
		ret += 2 // Thumb: LR = PC of next instruction; IRQ handler adjusts by 4 as in ARM
	}
	// ARM IRQ return address is (address of next instruction) + 4.
	c.exception(ModeIRQ, 0x18, ret+4-c.boolToU(c.Thumb)*2)
	return true
}

// execCopro handles CDP/MCR/MRC. MCR/MRC are routed to the Coproc hook (the ARM9's
// CP15 lives here); with no hook they are ignored so a trace/run does not stall on
// cache/MPU setup.
func (c *CPU) execCopro(w uint32) {
	if (w>>4)&1 == 0 { // CDP — no general-purpose effect modelled
		return
	}
	load := (w>>20)&1 == 1
	cp := (w >> 8) & 0xF
	op1 := (w >> 21) & 7
	crn := (w >> 16) & 0xF
	crm := w & 0xF
	op2 := (w >> 5) & 7
	rd := (w >> 12) & 0xF
	if c.Coproc != nil {
		v := c.reg(rd)
		c.Coproc(c, load, cp, op1, crn, crm, op2, &v)
		if load && rd != 15 {
			c.setReg(rd, v)
		}
		return
	}
	// No coprocessor hook: MRC reads as zero, MCR is dropped.
	if load && rd != 15 {
		c.setReg(rd, 0)
	}
}
