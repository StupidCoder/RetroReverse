package allegrex

// exec.go is the instruction interpreter. The Allegrex keeps the branch delay slot
// (via PC/nextPC) but, unlike the R3000, has interlocked loads: MIPS32R2 removed the
// load delay slot, so a load is visible to the very next instruction. The R/out
// two-register-file split is retained only for its clean within-step commit; loads
// write to out immediately (see load), so `lw $v0; jalr $v0` uses the loaded value.

// Step executes one instruction and returns an (approximate) cycle count. It
// fetches at PC, advances the delay-slot machinery, executes, then commits register
// writes for the next instruction to see.
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

	// A not-taken likely branch nullifies its delay-slot instruction: advance past
	// it without executing.
	if c.nullifyNext {
		c.nullifyNext = false
		c.R = c.out
		c.Steps++
		return 1
	}

	c.execute(w)

	c.R = c.out
	c.Steps++
	return 1
}

// reg reads a general register (the pre-load architectural value).
func (c *CPU) reg(i uint32) uint32 { return c.R[i] }

// load retires a load immediately (MIPS32R2 has no load delay slot): the value is
// written to the output file and is visible to the next instruction.
func (c *CPU) load(reg, val uint32) { c.set(reg, val) }

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
	simm := uint32(int32(int16(imm))) // sign-extended
	branchT := c.curPC + 4 + simm<<2  // PC-relative branch target
	jumpT := (c.curPC+4)&0xF0000000 | (w&0x03FFFFFF)<<2

	switch op {
	case 0x00:
		c.special(w, rs, rt, rd, shamt)
	case 0x01: // REGIMM: bltz/bgez (+al link, +l likely variants)
		switch rt {
		case 0x00, 0x01, 0x02, 0x03, 0x10, 0x11, 0x12, 0x13:
			s := int32(c.reg(rs))
			if rt&0x10 != 0 { // bltzal/bgezal/bltzall/bgezall link unconditionally
				c.set(31, c.curPC+8)
			}
			taken := (rt&1 == 0 && s < 0) || (rt&1 == 1 && s >= 0)
			if rt&0x02 != 0 && !taken { // likely: not-taken nullifies the delay slot
				c.nullifyNext = true
			} else {
				c.doBranch(taken, branchT)
			}
		default:
			c.Halt("unimplemented regimm rt=0x%02X (word 0x%08X) at 0x%08X", rt, w, c.curPC)
		}
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
	case 0x11: // COP1 (FPU)
		c.cop1(w, rs, rt, rd, shamt)
	case 0x12: // COP2 (VFPU register moves / command)
		c.cop2(w, rs, rt, rd)
	case 0x14, 0x15, 0x16, 0x17: // branch-likely
		c.branchLikely(op, rs, rt, branchT)
	case 0x1C: // SPECIAL2
		c.special2(w, rs, rt, rd)
	case 0x1F: // SPECIAL3
		c.special3(w, rs, rt, rd, shamt)
	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26:
		c.loadOp(op, rs, rt, simm)
	case 0x28, 0x29, 0x2A, 0x2B, 0x2E:
		c.storeOp(op, rs, rt, simm)
	case 0x30: // ll — modelled as a plain lw (no multiprocessor)
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.load(rt, c.read32(addr))
	case 0x38: // sc — store and report success (always succeeds here)
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write32(addr, c.reg(rt))
		c.set(rt, 1)
	case 0x31: // lwc1 — load to FPU register (delayed)
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.F[rt&31] = c.read32(addr)
	case 0x39: // swc1 — store FPU register
		addr := c.reg(rs) + simm
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write32(addr, c.F[rt&31])
	case 0x32, 0x36, 0x35, 0x3A, 0x3E, 0x3D, 0x34, 0x37, 0x3C, 0x18, 0x19, 0x1B:
		c.vfpuOp(w, op, rs, rt, simm)
	case 0x3F: // VFPU pipeline hints: vnop (0xFFFF0000), vsync, vflush — no-ops here
		if w>>16 != 0xFFFF {
			c.Halt("unimplemented opcode 0x3F (word 0x%08X) at 0x%08X", w, c.curPC)
		}
	default:
		c.Halt("unimplemented opcode 0x%02X (word 0x%08X) at 0x%08X", op, w, c.curPC)
	}
}

// branchLikely implements the MIPS "likely" branches: when not taken, the delay slot
// is nullified (skipped). Taken behaves like the ordinary branch.
func (c *CPU) branchLikely(op, rs, rt, target uint32) {
	var taken bool
	switch op {
	case 0x14: // beql
		taken = c.reg(rs) == c.reg(rt)
	case 0x15: // bnel
		taken = c.reg(rs) != c.reg(rt)
	case 0x16: // blezl
		taken = int32(c.reg(rs)) <= 0
	case 0x17: // bgtzl
		taken = int32(c.reg(rs)) > 0
	}
	if taken {
		c.doBranch(true, target)
	} else {
		c.nullifyNext = true // skip the delay-slot instruction
	}
}

// special2 executes the MIPS32R2 SPECIAL2 group (op 0x1C).
func (c *CPU) special2(w, rs, rt, rd uint32) {
	switch w & 0x3F {
	case 0x00: // madd
		p := int64(c.hilo()) + int64(int32(c.reg(rs)))*int64(int32(c.reg(rt)))
		c.setHilo(uint64(p))
	case 0x01: // maddu
		p := c.hilo() + uint64(c.reg(rs))*uint64(c.reg(rt))
		c.setHilo(p)
	case 0x02: // mul
		c.set(rd, uint32(int32(c.reg(rs))*int32(c.reg(rt))))
	case 0x04: // msub
		p := int64(c.hilo()) - int64(int32(c.reg(rs)))*int64(int32(c.reg(rt)))
		c.setHilo(uint64(p))
	case 0x05: // msubu
		p := c.hilo() - uint64(c.reg(rs))*uint64(c.reg(rt))
		c.setHilo(p)
	case 0x20: // clz
		c.set(rd, clz32(c.reg(rs)))
	case 0x21: // clo
		c.set(rd, clz32(^c.reg(rs)))
	default:
		c.Halt("unimplemented special2 funct 0x%02X at 0x%08X", w&0x3F, c.curPC)
	}
}

// special3 executes the MIPS32R2 SPECIAL3 group (op 0x1F): ext/ins and the BSHFL
// byte/half operations.
func (c *CPU) special3(w, rs, rt, rd, shamt uint32) {
	switch w & 0x3F {
	case 0x00: // ext rt, rs, pos=shamt, size=rd+1
		pos, size := shamt, rd+1
		mask := uint32((uint64(1) << size) - 1)
		c.set(rt, (c.reg(rs)>>pos)&mask)
	case 0x04: // ins rt, rs, pos=shamt, msb=rd -> size=rd-pos+1
		pos, msb := shamt, rd
		size := msb - pos + 1
		mask := uint32((uint64(1)<<size)-1) << pos
		c.set(rt, (c.reg(rt)&^mask)|((c.reg(rs)<<pos)&mask))
	case 0x20: // BSHFL
		switch shamt {
		case 0x02: // wsbh — swap bytes within each halfword
			v := c.reg(rt)
			c.set(rd, (v&0x00FF00FF)<<8|(v&0xFF00FF00)>>8)
		case 0x10: // seb
			c.set(rd, uint32(int32(int8(byte(c.reg(rt))))))
		case 0x18: // seh
			c.set(rd, uint32(int32(int16(uint16(c.reg(rt))))))
		case 0x14: // bitrev
			c.set(rd, bitrev32(c.reg(rt)))
		default:
			c.Halt("unimplemented bshfl shamt 0x%02X at 0x%08X", shamt, c.curPC)
		}
	default:
		c.Halt("unimplemented special3 funct 0x%02X at 0x%08X", w&0x3F, c.curPC)
	}
}

func (c *CPU) hilo() uint64     { return uint64(c.HI)<<32 | uint64(c.LO) }
func (c *CPU) setHilo(v uint64) { c.HI, c.LO = uint32(v>>32), uint32(v) }

func clz32(v uint32) uint32 {
	n := uint32(0)
	for i := 31; i >= 0; i-- {
		if v&(1<<uint(i)) != 0 {
			break
		}
		n++
	}
	return n
}

func bitrev32(v uint32) uint32 {
	var r uint32
	for i := 0; i < 32; i++ {
		r = (r << 1) | (v & 1)
		v >>= 1
	}
	return r
}

func (c *CPU) special(w, rs, rt, rd, shamt uint32) {
	funct := w & 63
	switch funct {
	case 0x00: // sll
		c.set(rd, c.reg(rt)<<shamt)
	case 0x02: // srl / rotr
		if rs&1 != 0 {
			c.set(rd, rotr32(c.reg(rt), shamt))
		} else {
			c.set(rd, c.reg(rt)>>shamt)
		}
	case 0x03: // sra
		c.set(rd, uint32(int32(c.reg(rt))>>shamt))
	case 0x04: // sllv
		c.set(rd, c.reg(rt)<<(c.reg(rs)&31))
	case 0x06: // srlv / rotrv
		if shamt&1 != 0 {
			c.set(rd, rotr32(c.reg(rt), c.reg(rs)&31))
		} else {
			c.set(rd, c.reg(rt)>>(c.reg(rs)&31))
		}
	case 0x07: // srav
		c.set(rd, uint32(int32(c.reg(rt))>>(c.reg(rs)&31)))
	case 0x08: // jr
		c.doBranch(true, c.reg(rs))
	case 0x09: // jalr
		c.set(rd, c.curPC+8)
		c.doBranch(true, c.reg(rs))
	case 0x0A: // movz
		if c.reg(rt) == 0 {
			c.set(rd, c.reg(rs))
		}
	case 0x0B: // movn
		if c.reg(rt) != 0 {
			c.set(rd, c.reg(rs))
		}
	case 0x0C: // syscall
		if c.Syscall != nil && c.Syscall(c, (w>>6)&0xFFFFF) {
			return
		}
		c.Exception(excSys)
	case 0x0F: // sync — no-op on a single-core model
		return
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
	case 0x16: // clz (Allegrex: SPECIAL funct, unlike MIPS32's SPECIAL2)
		c.set(rd, clz32(c.reg(rs)))
	case 0x17: // clo (Allegrex)
		c.set(rd, clz32(^c.reg(rs)))
	case 0x2C: // max (Allegrex)
		if int32(c.reg(rs)) > int32(c.reg(rt)) {
			c.set(rd, c.reg(rs))
		} else {
			c.set(rd, c.reg(rt))
		}
	case 0x2D: // min (Allegrex)
		if int32(c.reg(rs)) < int32(c.reg(rt)) {
			c.set(rd, c.reg(rs))
		} else {
			c.set(rd, c.reg(rt))
		}
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

func rotr32(v, n uint32) uint32 {
	n &= 31
	if n == 0 {
		return v
	}
	return v>>n | v<<(32-n)
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
