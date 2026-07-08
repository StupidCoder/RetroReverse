package mips

// exec.go is the instruction interpreter. See cpu.go for the two pipeline
// hazards (branch delay via PC/nextPC, load delay via the R/out register files).

// Step executes one instruction and returns an (approximate) cycle count. It
// fetches at PC, advances the delay-slot machinery, commits any pending load,
// executes, then commits register writes.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	c.curPC = c.PC
	c.delaySlot = c.pendingDelay
	c.pendingDelay = false

	w := c.read32(c.PC)
	c.PC = c.nextPC
	c.nextPC += 4

	// Commit the load issued by the previous instruction into the output file
	// before this instruction runs; reads below still see the old value via c.R.
	if c.ld.reg != 0 {
		c.out[c.ld.reg] = c.ld.val
	}
	c.ld = loadSlot{}

	c.execute(w)

	c.R = c.out
	c.Steps++
	return 1
}

// reg reads a general register (the pre-load architectural value).
func (c *CPU) reg(i uint32) uint32 { return c.R[i] }

// load issues a delayed load: it becomes visible one instruction later.
func (c *CPU) load(reg, val uint32) { c.ld = loadSlot{reg, val} }

// doBranch enters a branch delay slot; when taken it redirects nextPC. The delay
// slot executes either way because PC was already advanced past the branch.
func (c *CPU) doBranch(taken bool, target uint32) {
	c.pendingDelay = true
	c.branchAddr = c.curPC
	if taken {
		c.nextPC = target
	}
}

func (c *CPU) execute(w uint32) {
	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	imm := w & 0xFFFF
	simm := uint32(int32(int16(imm)))       // sign-extended
	branchT := c.curPC + 4 + simm<<2         // PC-relative branch target
	jumpT := (c.curPC+4)&0xF0000000 | (w&0x03FFFFFF)<<2

	switch op {
	case 0x00:
		c.special(w, rs, rt, rd, shamt)
	case 0x01: // REGIMM: bltz/bgez/bltzal/bgezal
		s := int32(c.reg(rs))
		link := rt&0x1E == 0x10 // bltzal(0x10)/bgezal(0x11) link unconditionally
		if link {
			c.set(31, c.curPC+8)
		}
		taken := (rt&1 == 0 && s < 0) || (rt&1 == 1 && s >= 0)
		c.doBranch(taken, branchT)
	case 0x02: // j
		c.doBranch(true, jumpT)
	case 0x03: // jal
		c.set(31, c.curPC+8)
		c.doBranch(true, jumpT)
	case 0x04: // beq
		c.doBranch(c.reg(rs) == c.reg(rt), branchT)
	case 0x05: // bne
		c.doBranch(c.reg(rs) != c.reg(rt), branchT)
	case 0x06: // blez
		c.doBranch(int32(c.reg(rs)) <= 0, branchT)
	case 0x07: // bgtz
		c.doBranch(int32(c.reg(rs)) > 0, branchT)
	case 0x08: // addi (trapping)
		if r, ov := addOv(c.reg(rs), simm); ov {
			c.Exception(excOv)
		} else {
			c.set(rt, r)
		}
	case 0x09: // addiu
		c.set(rt, c.reg(rs)+simm)
	case 0x0A: // slti
		c.set(rt, b2u(int32(c.reg(rs)) < int32(simm)))
	case 0x0B: // sltiu
		c.set(rt, b2u(c.reg(rs) < simm))
	case 0x0C: // andi
		c.set(rt, c.reg(rs)&imm)
	case 0x0D: // ori
		c.set(rt, c.reg(rs)|imm)
	case 0x0E: // xori
		c.set(rt, c.reg(rs)^imm)
	case 0x0F: // lui
		c.set(rt, imm<<16)
	case 0x10: // COP0
		c.cop0(w, rs, rt, rd)
	case 0x12: // COP2 / GTE
		c.cop2(w, rs, rt, rd)
	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26:
		c.loadOp(op, rs, rt, simm)
	case 0x28, 0x29, 0x2A, 0x2B, 0x2E:
		c.storeOp(op, rs, rt, simm)
	case 0x32: // lwc2
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		v := c.read32(addr)
		if c.GTE != nil {
			c.GTE.Write(rt, v)
		}
	case 0x3A: // swc2
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		var v uint32
		if c.GTE != nil {
			v = c.GTE.Read(rt)
		}
		c.write32(addr, v)
	default:
		c.Halt("unimplemented opcode 0x%02X (word 0x%08X) at 0x%08X", op, w, c.curPC)
	}
}

func (c *CPU) special(w, rs, rt, rd, shamt uint32) {
	funct := w & 63
	switch funct {
	case 0x00: // sll
		c.set(rd, c.reg(rt)<<shamt)
	case 0x02: // srl
		c.set(rd, c.reg(rt)>>shamt)
	case 0x03: // sra
		c.set(rd, uint32(int32(c.reg(rt))>>shamt))
	case 0x04: // sllv
		c.set(rd, c.reg(rt)<<(c.reg(rs)&31))
	case 0x06: // srlv
		c.set(rd, c.reg(rt)>>(c.reg(rs)&31))
	case 0x07: // srav
		c.set(rd, uint32(int32(c.reg(rt))>>(c.reg(rs)&31)))
	case 0x08: // jr
		c.doBranch(true, c.reg(rs))
	case 0x09: // jalr
		c.set(rd, c.curPC+8)
		c.doBranch(true, c.reg(rs))
	case 0x0C: // syscall
		if c.Syscall != nil && c.Syscall(c) {
			return
		}
		c.Exception(excSys)
	case 0x0D: // break
		c.Exception(excBp)
	case 0x10: // mfhi
		c.set(rd, c.HI)
	case 0x11: // mthi
		c.HI = c.reg(rs)
	case 0x12: // mflo
		c.set(rd, c.LO)
	case 0x13: // mtlo
		c.LO = c.reg(rs)
	case 0x18: // mult
		p := int64(int32(c.reg(rs))) * int64(int32(c.reg(rt)))
		c.LO, c.HI = uint32(p), uint32(uint64(p)>>32)
	case 0x19: // multu
		p := uint64(c.reg(rs)) * uint64(c.reg(rt))
		c.LO, c.HI = uint32(p), uint32(p>>32)
	case 0x1A: // div
		c.divSigned(int32(c.reg(rs)), int32(c.reg(rt)))
	case 0x1B: // divu
		c.divUnsigned(c.reg(rs), c.reg(rt))
	case 0x20: // add (trapping)
		if r, ov := addOv(c.reg(rs), c.reg(rt)); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x21: // addu
		c.set(rd, c.reg(rs)+c.reg(rt))
	case 0x22: // sub (trapping)
		if r, ov := subOv(c.reg(rs), c.reg(rt)); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x23: // subu
		c.set(rd, c.reg(rs)-c.reg(rt))
	case 0x24: // and
		c.set(rd, c.reg(rs)&c.reg(rt))
	case 0x25: // or
		c.set(rd, c.reg(rs)|c.reg(rt))
	case 0x26: // xor
		c.set(rd, c.reg(rs)^c.reg(rt))
	case 0x27: // nor
		c.set(rd, ^(c.reg(rs) | c.reg(rt)))
	case 0x2A: // slt
		c.set(rd, b2u(int32(c.reg(rs)) < int32(c.reg(rt))))
	case 0x2B: // sltu
		c.set(rd, b2u(c.reg(rs) < c.reg(rt)))
	default:
		c.Halt("unimplemented special funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, c.curPC)
	}
}

// divSigned/divUnsigned follow the R3000 results for the degenerate cases
// (divide by zero and the INT_MIN/-1 overflow), which real code relies on.
func (c *CPU) divSigned(a, b int32) {
	switch {
	case b == 0:
		c.HI = uint32(a)
		if a >= 0 {
			c.LO = 0xFFFFFFFF
		} else {
			c.LO = 1
		}
	case a == -0x80000000 && b == -1:
		c.LO, c.HI = 0x80000000, 0
	default:
		c.LO, c.HI = uint32(a/b), uint32(a%b)
	}
}

func (c *CPU) divUnsigned(a, b uint32) {
	if b == 0 {
		c.LO, c.HI = 0xFFFFFFFF, a
		return
	}
	c.LO, c.HI = a/b, a%b
}

func (c *CPU) loadOp(op, rs, rt, simm uint32) {
	addr := c.reg(rs) + simm
	switch op {
	case 0x20: // lb
		c.load(rt, uint32(int32(int8(byte(c.read8(addr))))))
	case 0x24: // lbu
		c.load(rt, c.read8(addr))
	case 0x21: // lh
		if addr&1 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.load(rt, uint32(int32(int16(uint16(c.read16(addr))))))
	case 0x25: // lhu
		if addr&1 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.load(rt, c.read16(addr))
	case 0x23: // lw
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.load(rt, c.read32(addr))
	case 0x22: // lwl
		word := c.read32(addr &^ 3)
		shift := (addr & 3) * 8
		cur := c.out[rt] // forward the just-committed load for the lwl;lwr idiom
		c.load(rt, (cur&(0x00FFFFFF>>shift))|(word<<(24-shift)))
	case 0x26: // lwr
		word := c.read32(addr &^ 3)
		shift := (addr & 3) * 8
		cur := c.out[rt]
		c.load(rt, (cur&(0xFFFFFF00<<(24-shift)))|(word>>shift))
	}
}

func (c *CPU) storeOp(op, rs, rt, simm uint32) {
	addr := c.reg(rs) + simm
	v := c.reg(rt)
	switch op {
	case 0x28: // sb
		c.write8(addr, v&0xFF)
	case 0x29: // sh
		if addr&1 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write16(addr, v&0xFFFF)
	case 0x2B: // sw
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write32(addr, v)
	case 0x2A: // swl
		word := c.read32(addr &^ 3)
		shift := (addr & 3) * 8
		c.write32(addr&^3, (word&(0xFFFFFF00<<shift))|(v>>(24-shift)))
	case 0x2E: // swr
		word := c.read32(addr &^ 3)
		shift := (addr & 3) * 8
		c.write32(addr&^3, (word&(0x00FFFFFF>>(24-shift)))|(v<<shift))
	}
}

func (c *CPU) cop0(w, rs, rt, rd uint32) {
	if w&(1<<25) != 0 { // COP0 command
		if w&0x3F == 0x10 { // rfe
			c.rfe()
			return
		}
		c.Halt("unimplemented cop0 command 0x%08X at 0x%08X", w, c.curPC)
		return
	}
	switch rs {
	case 0x00: // mfc0 (delayed like a load)
		c.load(rt, c.COP0[rd])
	case 0x04: // mtc0
		c.COP0[rd] = c.reg(rt)
	default:
		c.Halt("unimplemented cop0 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, c.curPC)
	}
}

func (c *CPU) cop2(w, rs, rt, rd uint32) {
	if w&(1<<25) != 0 { // GTE command
		if c.GTE != nil {
			c.GTE.Command(w & 0x01FFFFFF)
		}
		return
	}
	switch rs {
	case 0x00: // mfc2 (delayed)
		var v uint32
		if c.GTE != nil {
			v = c.GTE.Read(rd)
		}
		c.load(rt, v)
	case 0x02: // cfc2 (delayed)
		var v uint32
		if c.GTE != nil {
			v = c.GTE.ReadCtrl(rd)
		}
		c.load(rt, v)
	case 0x04: // mtc2
		if c.GTE != nil {
			c.GTE.Write(rd, c.reg(rt))
		}
	case 0x06: // ctc2
		if c.GTE != nil {
			c.GTE.WriteCtrl(rd, c.reg(rt))
		}
	default:
		c.Halt("unimplemented cop2 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, c.curPC)
	}
}

// addOv/subOv compute signed add/sub and report two's-complement overflow.
func addOv(a, b uint32) (uint32, bool) {
	r := a + b
	return r, (a^r)&(b^r)&0x80000000 != 0
}
func subOv(a, b uint32) (uint32, bool) {
	r := a - b
	return r, (a^b)&(a^r)&0x80000000 != 0
}

func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
