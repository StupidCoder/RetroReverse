package sh4

// exec.go is the instruction interpreter for everything outside group 1111 —
// the FPU space executes in exec_fpu.go, mirroring the decoder split.
//
// The delay-slot machinery is the r4300's: PC and nextPC advance in lockstep,
// a delayed transfer sets pendingDelay and redirects nextPC, and the slot at
// PC executes before the redirect lands. The SH-4-specific rules sit on top:
// an interrupt is never accepted between a delayed transfer and its slot
// (checkInterrupt tests pendingDelay), and a transfer decoded inside a delay
// slot is the architectural "slot illegal" — this core halts on it rather
// than modeling the exception, because compiled code never does it and a halt
// names the corruption that made it happen.

// Step executes one instruction and returns an (approximate) cycle count. It
// fetches at PC, advances the delay-slot machinery, and executes.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	c.tickTMU()
	if c.checkInterrupt() {
		return 1
	}

	c.curPC = c.PC
	c.delaySlot = c.pendingDelay
	c.pendingDelay = false

	if c.PC&1 != 0 {
		c.Halt("misaligned PC %08X", c.PC)
		return 1
	}
	h := c.fetchInstr(c.PC)

	c.PC = c.nextPC
	c.nextPC += 2

	c.execute(h)
	c.Steps++
	return 1
}

// checkInterrupt accepts the highest-priority pending interrupt, if any
// outranks SR.IMASK. The candidates are the external IRL request (level and
// INTEVT code supplied by the machine) and the three TMU channels (priority
// from IPRA). Acceptance is blocked by SR.BL and by a delayed transfer whose
// slot has not executed yet.
func (c *CPU) checkInterrupt() bool {
	if c.pendingDelay || c.SR&SRBL != 0 {
		return false
	}
	level, code := c.irlLevel, c.irlCode
	if l, cd, ok := c.tmuPending(); ok && l > level {
		level, code = l, cd
	}
	if level == 0 || level <= (c.SR>>4)&0xF {
		return false
	}
	c.INTEVT = code
	c.SSR = c.SR
	c.SPC = c.PC
	c.SGR = c.R[15]
	c.SetSR(c.SR | SRMD | SRRB | SRBL) // the handler wakes on the other bank
	c.PC = c.VBR + 0x600
	c.nextPC = c.PC + 2
	return true
}

// exception enters a general exception: trapa and its relatives. SPC is the
// address execution resumes at.
func (c *CPU) exception(expevt, spc uint32) {
	c.EXPEVT = expevt
	c.SSR = c.SR
	c.SPC = spc
	c.SGR = c.R[15]
	c.SetSR(c.SR | SRMD | SRRB | SRBL)
	c.PC = c.VBR + 0x100
	c.nextPC = c.PC + 2
}

// doBranch enters a delay slot; when taken it redirects nextPC. The slot
// executes either way because PC was already advanced past the transfer —
// SH-4's delayed conditionals (bt/s, bf/s) run their slot on both paths,
// unlike MIPS's branch-likely family.
func (c *CPU) doBranch(taken bool, target uint32) {
	if c.delaySlot {
		c.Halt("branch in delay slot at %08X", c.curPC)
		return
	}
	c.pendingDelay = true
	if taken {
		c.nextPC = target
	}
}

// doJumpNow is the undelayed conditional branch (bt, bf): no slot, PC simply
// redirects.
func (c *CPU) doJumpNow(taken bool, target uint32) {
	if c.delaySlot {
		c.Halt("branch in delay slot at %08X", c.curPC)
		return
	}
	if taken {
		c.PC = target
		c.nextPC = target + 2
	}
}

func (c *CPU) execute(h uint16) {
	n, m := rn(h), rm(h)
	switch h >> 12 {
	case 0x0:
		c.exec0(h, n, m)
	case 0x1: // mov.l Rm,@(disp,Rn)
		c.write32(c.R[n]+uint32(h&0xF)*4, c.R[m])
	case 0x2:
		c.exec2(h, n, m)
	case 0x3:
		c.exec3(h, n, m)
	case 0x4:
		c.exec4(h, n, m)
	case 0x5: // mov.l @(disp,Rm),Rn
		c.R[n] = c.read32(c.R[m] + uint32(h&0xF)*4)
	case 0x6:
		c.exec6(h, n, m)
	case 0x7: // add #imm,Rn
		c.R[n] += uint32(s8(h))
	case 0x8:
		c.exec8(h)
	case 0x9: // mov.w @(disp,PC),Rn
		c.R[n] = uint32(int32(int16(c.read16(litW(c.curPC, uint32(h&0xFF))))))
	case 0xA: // bra
		c.doBranch(true, uint32(int32(c.curPC+4)+s12(h)*2))
	case 0xB: // bsr
		c.PR = c.curPC + 4
		c.doBranch(true, uint32(int32(c.curPC+4)+s12(h)*2))
	case 0xC:
		c.execC(h)
	case 0xD: // mov.l @(disp,PC),Rn
		c.R[n] = c.read32(litL(c.curPC, uint32(h&0xFF)))
	case 0xE: // mov #imm,Rn
		c.R[n] = uint32(s8(h))
	case 0xF:
		c.execFPU(h, n, m)
	}
}

func (c *CPU) exec0(h uint16, n, m uint32) {
	switch h & 0xF {
	case 0x2: // stc <ctrl>,Rn
		switch {
		case m == 0:
			c.R[n] = c.SR
		case m == 1:
			c.R[n] = c.GBR
		case m == 2:
			c.R[n] = c.VBR
		case m == 3:
			c.R[n] = c.SSR
		case m == 4:
			c.R[n] = c.SPC
		case m >= 8:
			c.R[n] = c.bankReg(m)
		default:
			c.unknown(h)
		}
	case 0x3:
		switch m {
		case 0x0: // bsrf
			c.PR = c.curPC + 4
			c.doBranch(true, c.curPC+4+c.R[n])
		case 0x2: // braf
			c.doBranch(true, c.curPC+4+c.R[n])
		case 0x8: // pref: a hint everywhere except the store-queue window
			if a := c.R[n]; a >= 0xE0000000 && a < 0xE4000000 {
				c.sqFlush(a)
			}
		case 0x9, 0xA, 0xB: // ocbi/ocbp/ocbwb: cache ops, and there is no cache
		case 0xC: // movca.l: a plain store without the cache allocation
			c.write32(c.R[n], c.R[0])
		default:
			c.unknown(h)
		}
	case 0x4:
		c.write8(c.R[0]+c.R[n], uint8(c.R[m]))
	case 0x5:
		c.write16(c.R[0]+c.R[n], uint16(c.R[m]))
	case 0x6:
		c.write32(c.R[0]+c.R[n], c.R[m])
	case 0x7:
		c.MACL = c.R[n] * c.R[m]
	case 0x8:
		if n != 0 {
			c.unknown(h)
			return
		}
		switch m {
		case 0x0:
			c.setT(false)
		case 0x1:
			c.setT(true)
		case 0x2:
			c.MACH, c.MACL = 0, 0
		case 0x3:
			c.Halt("ldtlb with the MMU unmodelled (PC %08X)", c.curPC)
		case 0x4:
			c.SR &^= SRS
		case 0x5:
			c.SR |= SRS
		default:
			c.unknown(h)
		}
	case 0x9:
		switch {
		case h == 0x0009: // nop
		case h == 0x0019: // div0u
			c.SR &^= SRQ | SRM | SRT
		case m == 2: // movt
			c.R[n] = c.T()
		default:
			c.unknown(h)
		}
	case 0xA:
		switch m {
		case 0x0:
			c.R[n] = c.MACH
		case 0x1:
			c.R[n] = c.MACL
		case 0x2:
			c.R[n] = c.PR
		case 0x3:
			c.R[n] = c.SGR
		case 0x5:
			c.R[n] = c.FPUL
		case 0x6:
			c.R[n] = c.FPSCR
		case 0xF:
			c.R[n] = c.DBR
		default:
			c.unknown(h)
		}
	case 0xB:
		switch h {
		case 0x000B: // rts
			c.doBranch(true, c.PR)
		case 0x001B: // sleep: wait for an interrupt by re-executing until one lands
			c.PC = c.curPC
			c.nextPC = c.curPC + 2
		case 0x002B: // rte: SR restores now, so the slot runs with the restored SR
			target := c.SPC
			c.SetSR(c.SSR)
			c.doBranch(true, target)
		default:
			c.unknown(h)
		}
	case 0xC:
		c.R[n] = uint32(int32(int8(c.read8(c.R[0] + c.R[m]))))
	case 0xD:
		c.R[n] = uint32(int32(int16(c.read16(c.R[0] + c.R[m]))))
	case 0xE:
		c.R[n] = c.read32(c.R[0] + c.R[m])
	case 0xF:
		c.macL(n, m)
	default:
		c.unknown(h)
	}
}

func (c *CPU) exec2(h uint16, n, m uint32) {
	switch h & 0xF {
	case 0x0:
		c.write8(c.R[n], uint8(c.R[m]))
	case 0x1:
		c.write16(c.R[n], uint16(c.R[m]))
	case 0x2:
		c.write32(c.R[n], c.R[m])
	case 0x4:
		c.R[n] -= 1
		c.write8(c.R[n], uint8(c.R[m]))
	case 0x5:
		c.R[n] -= 2
		c.write16(c.R[n], uint16(c.R[m]))
	case 0x6:
		c.R[n] -= 4
		c.write32(c.R[n], c.R[m])
	case 0x7: // div0s
		c.SR &^= SRQ | SRM
		if c.R[n]&0x80000000 != 0 {
			c.SR |= SRQ
		}
		if c.R[m]&0x80000000 != 0 {
			c.SR |= SRM
		}
		c.setT((c.SR&SRQ != 0) != (c.SR&SRM != 0))
	case 0x8:
		c.setT(c.R[n]&c.R[m] == 0)
	case 0x9:
		c.R[n] &= c.R[m]
	case 0xA:
		c.R[n] ^= c.R[m]
	case 0xB:
		c.R[n] |= c.R[m]
	case 0xC: // cmp/str: T if any byte equal
		d := c.R[n] ^ c.R[m]
		c.setT(d&0xFF000000 == 0 || d&0x00FF0000 == 0 || d&0x0000FF00 == 0 || d&0x000000FF == 0)
	case 0xD: // xtrct
		c.R[n] = c.R[m]<<16 | c.R[n]>>16
	case 0xE:
		c.MACL = uint32(c.R[n]&0xFFFF) * uint32(c.R[m]&0xFFFF)
	case 0xF:
		c.MACL = uint32(int32(int16(c.R[n])) * int32(int16(c.R[m])))
	default:
		c.unknown(h)
	}
}

func (c *CPU) exec3(h uint16, n, m uint32) {
	switch h & 0xF {
	case 0x0:
		c.setT(c.R[n] == c.R[m])
	case 0x2:
		c.setT(c.R[n] >= c.R[m])
	case 0x3:
		c.setT(int32(c.R[n]) >= int32(c.R[m]))
	case 0x4:
		c.div1(n, m)
	case 0x5:
		p := uint64(c.R[n]) * uint64(c.R[m])
		c.MACH, c.MACL = uint32(p>>32), uint32(p)
	case 0x6:
		c.setT(c.R[n] > c.R[m])
	case 0x7:
		c.setT(int32(c.R[n]) > int32(c.R[m]))
	case 0x8:
		c.R[n] -= c.R[m]
	case 0xA: // subc
		tmp := c.R[n] - c.R[m]
		res := tmp - c.T()
		c.setT(c.R[n] < tmp || tmp < res)
		c.R[n] = res
	case 0xB: // subv
		res := c.R[n] - c.R[m]
		c.setT(int32((c.R[n]^c.R[m])&(c.R[n]^res)) < 0)
		c.R[n] = res
	case 0xC:
		c.R[n] += c.R[m]
	case 0xD:
		p := uint64(int64(int32(c.R[n])) * int64(int32(c.R[m])))
		c.MACH, c.MACL = uint32(p>>32), uint32(p)
	case 0xE: // addc
		tmp := c.R[n] + c.R[m]
		res := tmp + c.T()
		c.setT(tmp < c.R[n] || res < tmp)
		c.R[n] = res
	case 0xF: // addv
		res := c.R[n] + c.R[m]
		c.setT(int32(^(c.R[n]^c.R[m])&(c.R[n]^res)) < 0)
		c.R[n] = res
	default:
		c.unknown(h)
	}
}

func (c *CPU) exec4(h uint16, n, m uint32) {
	switch h & 0xF {
	case 0xC: // shad: left on positive, arithmetic right on negative
		if sh := c.R[m]; int32(sh) >= 0 {
			c.R[n] <<= sh & 31
		} else if sh&31 == 0 {
			c.R[n] = uint32(int32(c.R[n]) >> 31)
		} else {
			c.R[n] = uint32(int32(c.R[n]) >> (32 - sh&31))
		}
		return
	case 0xD: // shld: as shad, logically
		if sh := c.R[m]; int32(sh) >= 0 {
			c.R[n] <<= sh & 31
		} else if sh&31 == 0 {
			c.R[n] = 0
		} else {
			c.R[n] >>= 32 - sh&31
		}
		return
	case 0xF:
		c.macW(n, m)
		return
	case 0x3:
		if h&0x80 != 0 { // stc.l Rm_bank,@-Rn
			c.R[n] -= 4
			c.write32(c.R[n], c.bankReg(m))
			return
		}
	case 0x7:
		if h&0x80 != 0 { // ldc.l @Rm+,Rn_bank
			c.setBankReg(m, c.read32(c.R[n]))
			c.R[n] += 4
			return
		}
	case 0xE:
		if h&0x80 != 0 { // ldc Rm,Rn_bank
			c.setBankReg(m, c.R[n])
			return
		}
	}
	switch h & 0xFF {
	case 0x00: // shll
		c.setT(c.R[n]&0x80000000 != 0)
		c.R[n] <<= 1
	case 0x01: // shlr
		c.setT(c.R[n]&1 != 0)
		c.R[n] >>= 1
	case 0x02:
		c.R[n] -= 4
		c.write32(c.R[n], c.MACH)
	case 0x03:
		c.R[n] -= 4
		c.write32(c.R[n], c.SR)
	case 0x04: // rotl
		c.setT(c.R[n]&0x80000000 != 0)
		c.R[n] = c.R[n]<<1 | c.R[n]>>31
	case 0x05: // rotr
		c.setT(c.R[n]&1 != 0)
		c.R[n] = c.R[n]>>1 | c.R[n]<<31
	case 0x06:
		c.MACH = c.read32(c.R[n])
		c.R[n] += 4
	case 0x07:
		c.SetSR(c.read32(c.R[n]))
		c.R[n] += 4
	case 0x08:
		c.R[n] <<= 2
	case 0x09:
		c.R[n] >>= 2
	case 0x0A:
		c.MACH = c.R[n]
	case 0x0B: // jsr
		c.PR = c.curPC + 4
		c.doBranch(true, c.R[n])
	case 0x0E:
		c.SetSR(c.R[n])
	case 0x10: // dt
		c.R[n]--
		c.setT(c.R[n] == 0)
	case 0x11:
		c.setT(int32(c.R[n]) >= 0)
	case 0x12:
		c.R[n] -= 4
		c.write32(c.R[n], c.MACL)
	case 0x13:
		c.R[n] -= 4
		c.write32(c.R[n], c.GBR)
	case 0x15:
		c.setT(int32(c.R[n]) > 0)
	case 0x16:
		c.MACL = c.read32(c.R[n])
		c.R[n] += 4
	case 0x17:
		c.GBR = c.read32(c.R[n])
		c.R[n] += 4
	case 0x18:
		c.R[n] <<= 8
	case 0x19:
		c.R[n] >>= 8
	case 0x1A:
		c.MACL = c.R[n]
	case 0x1B: // tas.b
		b := c.read8(c.R[n])
		c.setT(b == 0)
		c.write8(c.R[n], b|0x80)
	case 0x1E:
		c.GBR = c.R[n]
	case 0x20: // shal
		c.setT(c.R[n]&0x80000000 != 0)
		c.R[n] <<= 1
	case 0x21: // shar
		c.setT(c.R[n]&1 != 0)
		c.R[n] = uint32(int32(c.R[n]) >> 1)
	case 0x22:
		c.R[n] -= 4
		c.write32(c.R[n], c.PR)
	case 0x23:
		c.R[n] -= 4
		c.write32(c.R[n], c.VBR)
	case 0x24: // rotcl
		t := c.R[n] >> 31
		c.R[n] = c.R[n]<<1 | c.T()
		c.setT(t != 0)
	case 0x25: // rotcr
		t := c.R[n] & 1
		c.R[n] = c.R[n]>>1 | c.T()<<31
		c.setT(t != 0)
	case 0x26:
		c.PR = c.read32(c.R[n])
		c.R[n] += 4
	case 0x27:
		c.VBR = c.read32(c.R[n])
		c.R[n] += 4
	case 0x28:
		c.R[n] <<= 16
	case 0x29:
		c.R[n] >>= 16
	case 0x2A:
		c.PR = c.R[n]
	case 0x2B: // jmp
		c.doBranch(true, c.R[n])
	case 0x2E:
		c.VBR = c.R[n]
	case 0x32:
		c.R[n] -= 4
		c.write32(c.R[n], c.SGR)
	case 0x33:
		c.R[n] -= 4
		c.write32(c.R[n], c.SSR)
	case 0x37:
		c.SSR = c.read32(c.R[n])
		c.R[n] += 4
	case 0x3E:
		c.SSR = c.R[n]
	case 0x43:
		c.R[n] -= 4
		c.write32(c.R[n], c.SPC)
	case 0x47:
		c.SPC = c.read32(c.R[n])
		c.R[n] += 4
	case 0x4E:
		c.SPC = c.R[n]
	case 0x52:
		c.R[n] -= 4
		c.write32(c.R[n], c.FPUL)
	case 0x56:
		c.FPUL = c.read32(c.R[n])
		c.R[n] += 4
	case 0x5A:
		c.FPUL = c.R[n]
	case 0x62:
		c.R[n] -= 4
		c.write32(c.R[n], c.FPSCR)
	case 0x66:
		c.SetFPSCR(c.read32(c.R[n]))
		c.R[n] += 4
	case 0x6A:
		c.SetFPSCR(c.R[n])
	case 0xF2:
		c.R[n] -= 4
		c.write32(c.R[n], c.DBR)
	case 0xF6:
		c.DBR = c.read32(c.R[n])
		c.R[n] += 4
	case 0xFA:
		c.DBR = c.R[n]
	default:
		c.unknown(h)
	}
}

func (c *CPU) exec6(h uint16, n, m uint32) {
	switch h & 0xF {
	case 0x0:
		c.R[n] = uint32(int32(int8(c.read8(c.R[m]))))
	case 0x1:
		c.R[n] = uint32(int32(int16(c.read16(c.R[m]))))
	case 0x2:
		c.R[n] = c.read32(c.R[m])
	case 0x3:
		c.R[n] = c.R[m]
	case 0x4:
		c.R[n] = uint32(int32(int8(c.read8(c.R[m]))))
		if n != m {
			c.R[m]++
		}
	case 0x5:
		c.R[n] = uint32(int32(int16(c.read16(c.R[m]))))
		if n != m {
			c.R[m] += 2
		}
	case 0x6:
		c.R[n] = c.read32(c.R[m])
		if n != m {
			c.R[m] += 4
		}
	case 0x7:
		c.R[n] = ^c.R[m]
	case 0x8: // swap.b: the low two bytes swap, the high half rides along
		v := c.R[m]
		c.R[n] = v&0xFFFF0000 | v<<8&0xFF00 | v>>8&0xFF
	case 0x9: // swap.w
		c.R[n] = c.R[m]<<16 | c.R[m]>>16
	case 0xA: // negc
		res := 0 - c.R[m] - c.T()
		c.setT(c.R[m] != 0 || c.T() != 0)
		c.R[n] = res
	case 0xB:
		c.R[n] = -c.R[m]
	case 0xC:
		c.R[n] = c.R[m] & 0xFF
	case 0xD:
		c.R[n] = c.R[m] & 0xFFFF
	case 0xE:
		c.R[n] = uint32(int32(int8(c.R[m])))
	case 0xF:
		c.R[n] = uint32(int32(int16(c.R[m])))
	}
}

func (c *CPU) exec8(h uint16) {
	m := rm(h)
	switch (h >> 8) & 0xF {
	case 0x0:
		c.write8(c.R[m]+uint32(h&0xF), uint8(c.R[0]))
	case 0x1:
		c.write16(c.R[m]+uint32(h&0xF)*2, uint16(c.R[0]))
	case 0x4:
		c.R[0] = uint32(int32(int8(c.read8(c.R[m] + uint32(h&0xF)))))
	case 0x5:
		c.R[0] = uint32(int32(int16(c.read16(c.R[m] + uint32(h&0xF)*2))))
	case 0x8:
		c.setT(int32(c.R[0]) == s8(h))
	case 0x9:
		c.doJumpNow(c.T() != 0, uint32(int32(c.curPC+4)+s8(h)*2))
	case 0xB:
		c.doJumpNow(c.T() == 0, uint32(int32(c.curPC+4)+s8(h)*2))
	case 0xD:
		c.doBranch(c.T() != 0, uint32(int32(c.curPC+4)+s8(h)*2))
	case 0xF:
		c.doBranch(c.T() == 0, uint32(int32(c.curPC+4)+s8(h)*2))
	default:
		c.unknown(h)
	}
}

func (c *CPU) execC(h uint16) {
	d := uint32(h & 0xFF)
	switch (h >> 8) & 0xF {
	case 0x0:
		c.write8(c.GBR+d, uint8(c.R[0]))
	case 0x1:
		c.write16(c.GBR+d*2, uint16(c.R[0]))
	case 0x2:
		c.write32(c.GBR+d*4, c.R[0])
	case 0x3: // trapa
		c.TRA = d << 2
		c.exception(0x160, c.PC)
	case 0x4:
		c.R[0] = uint32(int32(int8(c.read8(c.GBR + d))))
	case 0x5:
		c.R[0] = uint32(int32(int16(c.read16(c.GBR + d*2))))
	case 0x6:
		c.R[0] = c.read32(c.GBR + d*4)
	case 0x7: // mova
		c.R[0] = litL(c.curPC, d)
	case 0x8:
		c.setT(c.R[0]&d == 0)
	case 0x9:
		c.R[0] &= d
	case 0xA:
		c.R[0] ^= d
	case 0xB:
		c.R[0] |= d
	case 0xC:
		c.setT(uint32(c.read8(c.GBR+c.R[0]))&d == 0)
	case 0xD:
		a := c.GBR + c.R[0]
		c.write8(a, c.read8(a)&uint8(d))
	case 0xE:
		a := c.GBR + c.R[0]
		c.write8(a, c.read8(a)^uint8(d))
	case 0xF:
		a := c.GBR + c.R[0]
		c.write8(a, c.read8(a)|uint8(d))
	}
}

// bankReg reads the other bank's R0-R7 — the one SR.RB has not made live.
func (c *CPU) bankReg(m uint32) uint32 { return c.Rbank[m&7] }
func (c *CPU) setBankReg(m, v uint32)  { c.Rbank[m&7] = v }

// div1 is one non-restoring division step, transcribed from the manual's
// algorithm: div0s/div0u seed Q, M and T, and 32 div1 steps produce a 32-bit
// quotient bit-by-bit through T.
func (c *CPU) div1(n, m uint32) {
	oldQ := c.SR&SRQ != 0
	q := c.R[n]&0x80000000 != 0
	mf := c.SR&SRM != 0
	tmp2 := c.R[m]
	c.R[n] = c.R[n]<<1 | c.T()
	tmp0 := c.R[n]
	var tmp1 bool
	switch {
	case !oldQ && !mf:
		c.R[n] -= tmp2
		tmp1 = c.R[n] > tmp0
		q = q != tmp1 // Q==0: Q=tmp1; Q==1: Q=!tmp1
	case !oldQ && mf:
		c.R[n] += tmp2
		tmp1 = c.R[n] < tmp0
		q = q == tmp1 // Q==0: Q=!tmp1; Q==1: Q=tmp1
	case oldQ && !mf:
		c.R[n] += tmp2
		tmp1 = c.R[n] < tmp0
		q = q != tmp1
	default: // oldQ && mf
		c.R[n] -= tmp2
		tmp1 = c.R[n] > tmp0
		q = q == tmp1
	}
	if q {
		c.SR |= SRQ
	} else {
		c.SR &^= SRQ
	}
	c.setT(q == mf)
}

// macL is mac.l: a signed 64-bit multiply-accumulate through memory, with
// 48-bit saturation when SR.S is set.
func (c *CPU) macL(n, m uint32) {
	a := int64(int32(c.read32(c.R[n])))
	c.R[n] += 4
	b := int64(int32(c.read32(c.R[m])))
	c.R[m] += 4
	mac := int64(uint64(c.MACH)<<32|uint64(c.MACL)) + a*b
	if c.SR&SRS != 0 {
		const hi, lo = int64(0x00007FFFFFFFFFFF), int64(-0x0000800000000000)
		if mac > hi {
			mac = hi
		} else if mac < lo {
			mac = lo
		}
	}
	c.MACH, c.MACL = uint32(uint64(mac)>>32), uint32(uint64(mac))
}

// macW is mac.w: a signed 16-bit multiply-accumulate through memory. With
// SR.S set the accumulation saturates to 32 bits in MACL and flags the
// overflow in MACH's low bit, per the manual.
func (c *CPU) macW(n, m uint32) {
	a := int64(int16(c.read16(c.R[n])))
	c.R[n] += 2
	b := int64(int16(c.read16(c.R[m])))
	c.R[m] += 2
	if c.SR&SRS != 0 {
		sum := int64(int32(c.MACL)) + a*b
		if sum > 0x7FFFFFFF {
			sum = 0x7FFFFFFF
			c.MACH |= 1
		} else if sum < -0x80000000 {
			sum = -0x80000000
			c.MACH |= 1
		}
		c.MACL = uint32(sum)
		return
	}
	mac := int64(uint64(c.MACH)<<32|uint64(c.MACL)) + a*b
	c.MACH, c.MACL = uint32(uint64(mac)>>32), uint32(uint64(mac))
}

func (c *CPU) unknown(h uint16) {
	c.Halt("unimplemented instruction %04X (%s) at %08X", h, DecodeHalfword(h, c.curPC).Text, c.curPC)
}
