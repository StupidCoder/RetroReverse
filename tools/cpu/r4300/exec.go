package r4300

// exec.go is the instruction interpreter. See cpu.go for the branch delay slot
// and cop0.go for address translation.
//
// Two rules govern every 32-bit operation on this 64-bit core, and both are easy
// to get silently wrong:
//
//   - A 32-bit result is written back *sign-extended* into the full register.
//     `addu $t0, $t1, $t2` on values that fit in 32 bits still sets the high half
//     of $t0 to the sign of the result. Only the `d`-prefixed ops are truly
//     64-bit.
//   - A 32-bit shift or arithmetic op reads only the low half of its operands.

import "math/bits"

// Step executes one instruction and returns an (approximate) cycle count. It
// fetches at PC, advances the delay-slot machinery, and executes.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	c.tickCount()
	if c.checkInterrupt() {
		return 1
	}

	c.curPC = c.PC
	c.delaySlot = c.pendingDelay
	c.pendingDelay = false

	paddr, ok := c.translateFetch(c.PC)
	if !ok {
		return 1 // the fault has already redirected the PC
	}
	w := c.fetch(paddr)

	c.PC = c.nextPC
	c.nextPC += 4

	c.execute(w)
	c.Steps++
	return 1
}

// translateFetch translates an instruction fetch, raising an address error on a
// misaligned PC.
func (c *CPU) translateFetch(vaddr uint64) (uint32, bool) {
	if vaddr&3 != 0 {
		c.addrError(excAdEL, vaddr)
		return 0, false
	}
	return c.Translate(vaddr, false)
}

// doBranch enters a branch delay slot; when taken it redirects nextPC. The delay
// slot executes either way because PC was already advanced past the branch.
func (c *CPU) doBranch(taken bool, target uint64) {
	c.pendingDelay = true
	c.branchAddr = c.curPC
	if taken {
		c.nextPC = target
	}
}

// doBranchLikely is the branch-likely form: when the branch is not taken the
// delay slot is *annulled* — skipped entirely — by stepping PC over it. When
// taken it behaves exactly like doBranch.
func (c *CPU) doBranchLikely(taken bool, target uint64) {
	if taken {
		c.doBranch(true, target)
		return
	}
	c.PC = c.nextPC // skip the delay slot
	c.nextPC += 4
}

func (c *CPU) execute(w uint32) {
	op := w >> 26
	rs := (w >> 21) & 31
	rt := (w >> 16) & 31
	imm := w & 0xFFFF
	simm := uint64(int64(int16(imm))) // sign-extended to 64 bits
	branchT := c.curPC + 4 + simm<<2
	jumpT := (c.curPC+4)&0xFFFFFFFFF0000000 | uint64(w&0x03FFFFFF)<<2

	switch op {
	case 0x00:
		c.special(w, rs, rt)
	case 0x01:
		c.regimm(w, rs, rt, branchT, simm)
	case 0x02: // j
		c.doBranch(true, jumpT)
	case 0x03: // jal
		c.set(31, c.curPC+8)
		c.doBranch(true, jumpT)
	case 0x04: // beq
		c.doBranch(c.R[rs] == c.R[rt], branchT)
	case 0x05: // bne
		c.doBranch(c.R[rs] != c.R[rt], branchT)
	case 0x06: // blez
		c.doBranch(int64(c.R[rs]) <= 0, branchT)
	case 0x07: // bgtz
		c.doBranch(int64(c.R[rs]) > 0, branchT)

	case 0x14: // beql
		c.doBranchLikely(c.R[rs] == c.R[rt], branchT)
	case 0x15: // bnel
		c.doBranchLikely(c.R[rs] != c.R[rt], branchT)
	case 0x16: // blezl
		c.doBranchLikely(int64(c.R[rs]) <= 0, branchT)
	case 0x17: // bgtzl
		c.doBranchLikely(int64(c.R[rs]) > 0, branchT)

	case 0x08: // addi (trapping)
		if r, ov := addOv32(uint32(c.R[rs]), uint32(simm)); ov {
			c.Exception(excOv)
		} else {
			c.set(rt, sext32(r))
		}
	case 0x09: // addiu
		c.set(rt, sext32(uint32(c.R[rs])+uint32(simm)))
	case 0x0A: // slti
		c.set(rt, b2u(int64(c.R[rs]) < int64(simm)))
	case 0x0B: // sltiu
		c.set(rt, b2u(c.R[rs] < simm))
	case 0x0C: // andi
		c.set(rt, c.R[rs]&uint64(imm))
	case 0x0D: // ori
		c.set(rt, c.R[rs]|uint64(imm))
	case 0x0E: // xori
		c.set(rt, c.R[rs]^uint64(imm))
	case 0x0F: // lui
		c.set(rt, sext32(imm<<16))

	case 0x18: // daddi (trapping)
		if r, ov := addOv64(c.R[rs], simm); ov {
			c.Exception(excOv)
		} else {
			c.set(rt, r)
		}
	case 0x19: // daddiu
		c.set(rt, c.R[rs]+simm)

	case 0x10:
		c.cop0(w, rs, rt, (w>>11)&31)
	case 0x11:
		c.cop1(w, rs, rt, (w>>11)&31, (w>>6)&31, branchT)

	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x1A, 0x1B, 0x30, 0x34, 0x37:
		c.loadOp(op, rs, rt, simm)
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x38, 0x3C, 0x3F:
		c.storeOp(op, rs, rt, simm)

	case 0x2F:
		// cache: the caches are not modelled, so invalidation is a no-op. The
		// opcode itself must exist — libultra brackets every DMA with
		// osInvalDCache/osWritebackDCache, so halting here would stop the boot.

	case 0x31, 0x35, 0x39, 0x3D: // lwc1 / ldc1 / swc1 / sdc1
		c.cop1Mem(op, rs, rt, simm)

	default:
		c.Halt("unimplemented opcode 0x%02X (word 0x%08X) at 0x%08X", op, w, uint32(c.curPC))
	}
}

func (c *CPU) special(w, rs, rt uint32) {
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	funct := w & 63

	switch funct {
	case 0x00: // sll (also nop)
		c.set(rd, sext32(uint32(c.R[rt])<<shamt))
	case 0x02: // srl
		c.set(rd, sext32(uint32(c.R[rt])>>shamt))
	case 0x03: // sra
		c.set(rd, sext32(uint32(int32(uint32(c.R[rt]))>>shamt)))
	case 0x04: // sllv
		c.set(rd, sext32(uint32(c.R[rt])<<(c.R[rs]&31)))
	case 0x06: // srlv
		c.set(rd, sext32(uint32(c.R[rt])>>(c.R[rs]&31)))
	case 0x07: // srav
		c.set(rd, sext32(uint32(int32(uint32(c.R[rt]))>>(c.R[rs]&31))))

	case 0x08: // jr
		c.doBranch(true, c.R[rs])
	case 0x09: // jalr
		c.set(rd, c.curPC+8)
		c.doBranch(true, c.R[rs])

	case 0x0C: // syscall
		c.Exception(excSys)
	case 0x0D: // break
		c.Exception(excBp)
	case 0x0F: // sync — the store buffer is not modelled

	case 0x10: // mfhi
		c.set(rd, c.HI)
	case 0x11: // mthi
		c.HI = c.R[rs]
	case 0x12: // mflo
		c.set(rd, c.LO)
	case 0x13: // mtlo
		c.LO = c.R[rs]

	case 0x14: // dsllv
		c.set(rd, c.R[rt]<<(c.R[rs]&63))
	case 0x16: // dsrlv
		c.set(rd, c.R[rt]>>(c.R[rs]&63))
	case 0x17: // dsrav
		c.set(rd, uint64(int64(c.R[rt])>>(c.R[rs]&63)))

	case 0x18: // mult — 32-bit operands, sign-extended 64-bit halves
		p := int64(int32(uint32(c.R[rs]))) * int64(int32(uint32(c.R[rt])))
		c.LO, c.HI = sext32(uint32(p)), sext32(uint32(uint64(p)>>32))
	case 0x19: // multu
		p := uint64(uint32(c.R[rs])) * uint64(uint32(c.R[rt]))
		c.LO, c.HI = sext32(uint32(p)), sext32(uint32(p>>32))
	case 0x1A: // div
		c.divSigned(int32(uint32(c.R[rs])), int32(uint32(c.R[rt])))
	case 0x1B: // divu
		c.divUnsigned(uint32(c.R[rs]), uint32(c.R[rt]))
	case 0x1C: // dmult
		hi, lo := mul64Signed(int64(c.R[rs]), int64(c.R[rt]))
		c.LO, c.HI = lo, hi
	case 0x1D: // dmultu
		hi, lo := bits.Mul64(c.R[rs], c.R[rt])
		c.LO, c.HI = lo, hi
	case 0x1E: // ddiv
		c.ddivSigned(int64(c.R[rs]), int64(c.R[rt]))
	case 0x1F: // ddivu
		c.ddivUnsigned(c.R[rs], c.R[rt])

	case 0x20: // add (trapping)
		if r, ov := addOv32(uint32(c.R[rs]), uint32(c.R[rt])); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, sext32(r))
		}
	case 0x21: // addu
		c.set(rd, sext32(uint32(c.R[rs])+uint32(c.R[rt])))
	case 0x22: // sub (trapping)
		if r, ov := subOv32(uint32(c.R[rs]), uint32(c.R[rt])); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, sext32(r))
		}
	case 0x23: // subu
		c.set(rd, sext32(uint32(c.R[rs])-uint32(c.R[rt])))
	case 0x24: // and
		c.set(rd, c.R[rs]&c.R[rt])
	case 0x25: // or
		c.set(rd, c.R[rs]|c.R[rt])
	case 0x26: // xor
		c.set(rd, c.R[rs]^c.R[rt])
	case 0x27: // nor
		c.set(rd, ^(c.R[rs] | c.R[rt]))
	case 0x2A: // slt
		c.set(rd, b2u(int64(c.R[rs]) < int64(c.R[rt])))
	case 0x2B: // sltu
		c.set(rd, b2u(c.R[rs] < c.R[rt]))

	case 0x2C: // dadd (trapping)
		if r, ov := addOv64(c.R[rs], c.R[rt]); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x2D: // daddu
		c.set(rd, c.R[rs]+c.R[rt])
	case 0x2E: // dsub (trapping)
		if r, ov := subOv64(c.R[rs], c.R[rt]); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x2F: // dsubu
		c.set(rd, c.R[rs]-c.R[rt])

	case 0x30: // tge
		c.trapIf(int64(c.R[rs]) >= int64(c.R[rt]))
	case 0x31: // tgeu
		c.trapIf(c.R[rs] >= c.R[rt])
	case 0x32: // tlt
		c.trapIf(int64(c.R[rs]) < int64(c.R[rt]))
	case 0x33: // tltu
		c.trapIf(c.R[rs] < c.R[rt])
	case 0x34: // teq
		c.trapIf(c.R[rs] == c.R[rt])
	case 0x36: // tne
		c.trapIf(c.R[rs] != c.R[rt])

	case 0x38: // dsll
		c.set(rd, c.R[rt]<<shamt)
	case 0x3A: // dsrl
		c.set(rd, c.R[rt]>>shamt)
	case 0x3B: // dsra
		c.set(rd, uint64(int64(c.R[rt])>>shamt))
	case 0x3C: // dsll32 — shift amounts 32..63
		c.set(rd, c.R[rt]<<(shamt+32))
	case 0x3E: // dsrl32
		c.set(rd, c.R[rt]>>(shamt+32))
	case 0x3F: // dsra32
		c.set(rd, uint64(int64(c.R[rt])>>(shamt+32)))

	default:
		c.Halt("unimplemented special funct 0x%02X (word 0x%08X) at 0x%08X", funct, w, uint32(c.curPC))
	}
}

func (c *CPU) regimm(w, rs, rt uint32, branchT, simm uint64) {
	s := int64(c.R[rs])
	switch rt {
	case 0x00: // bltz
		c.doBranch(s < 0, branchT)
	case 0x01: // bgez
		c.doBranch(s >= 0, branchT)
	case 0x02: // bltzl
		c.doBranchLikely(s < 0, branchT)
	case 0x03: // bgezl
		c.doBranchLikely(s >= 0, branchT)

	case 0x08: // tgei
		c.trapIf(s >= int64(simm))
	case 0x09: // tgeiu
		c.trapIf(c.R[rs] >= simm)
	case 0x0A: // tlti
		c.trapIf(s < int64(simm))
	case 0x0B: // tltiu
		c.trapIf(c.R[rs] < simm)
	case 0x0C: // teqi
		c.trapIf(c.R[rs] == simm)
	case 0x0E: // tnei
		c.trapIf(c.R[rs] != simm)

	// The linking forms write $ra unconditionally, before the branch decides.
	case 0x10: // bltzal
		c.set(31, c.curPC+8)
		c.doBranch(s < 0, branchT)
	case 0x11: // bgezal
		c.set(31, c.curPC+8)
		c.doBranch(s >= 0, branchT)
	case 0x12: // bltzall
		c.set(31, c.curPC+8)
		c.doBranchLikely(s < 0, branchT)
	case 0x13: // bgezall
		c.set(31, c.curPC+8)
		c.doBranchLikely(s >= 0, branchT)

	default:
		c.Halt("unimplemented regimm rt=0x%02X (word 0x%08X) at 0x%08X", rt, w, uint32(c.curPC))
	}
}

func (c *CPU) trapIf(cond bool) {
	if cond {
		c.Exception(excTrap)
	}
}

// --- multiply and divide ----------------------------------------------------

// divSigned/divUnsigned follow the hardware's results for the degenerate cases
// (divide by zero and the INT_MIN/-1 overflow), which real code relies on.
func (c *CPU) divSigned(a, b int32) {
	switch {
	case b == 0:
		c.HI = sext32(uint32(a))
		if a >= 0 {
			c.LO = sext32(0xFFFFFFFF)
		} else {
			c.LO = 1
		}
	case a == -0x80000000 && b == -1:
		c.LO, c.HI = sext32(0x80000000), 0
	default:
		c.LO, c.HI = sext32(uint32(a/b)), sext32(uint32(a%b))
	}
}

func (c *CPU) divUnsigned(a, b uint32) {
	if b == 0 {
		c.LO, c.HI = sext32(0xFFFFFFFF), sext32(a)
		return
	}
	c.LO, c.HI = sext32(a/b), sext32(a%b)
}

func (c *CPU) ddivSigned(a, b int64) {
	switch {
	case b == 0:
		c.HI = uint64(a)
		if a >= 0 {
			c.LO = ^uint64(0)
		} else {
			c.LO = 1
		}
	case a == -0x8000000000000000 && b == -1:
		c.LO, c.HI = uint64(a), 0
	default:
		c.LO, c.HI = uint64(a/b), uint64(a%b)
	}
}

func (c *CPU) ddivUnsigned(a, b uint64) {
	if b == 0 {
		c.LO, c.HI = ^uint64(0), a
		return
	}
	c.LO, c.HI = a/b, a%b
}

// mul64Signed is the signed 64x64 -> 128 product, as dmult computes it. Go's
// bits.Mul64 is unsigned, so the sign corrections are applied afterwards.
func mul64Signed(a, b int64) (hi, lo uint64) {
	hi, lo = bits.Mul64(uint64(a), uint64(b))
	if a < 0 {
		hi -= uint64(b)
	}
	if b < 0 {
		hi -= uint64(a)
	}
	return hi, lo
}

// --- loads and stores -------------------------------------------------------

// The unaligned load/store family (lwl/lwr/ldl/ldr and their store forms) reads
// or writes the part of an aligned word or doubleword that lies on the same side
// of the address as the byte selected. On a big-endian machine "left" is the
// high-order end.

func (c *CPU) loadOp(op, rs, rt uint32, simm uint64) {
	vaddr := c.R[rs] + simm

	// align checks the access width and raises an address error on a misaligned
	// effective address, returning the physical address on success.
	align := func(n uint64) (uint32, bool) {
		if vaddr&(n-1) != 0 {
			c.addrError(excAdEL, vaddr)
			return 0, false
		}
		return c.Translate(vaddr, false)
	}

	switch op {
	case 0x20: // lb
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		c.set(rt, uint64(int64(int8(byte(c.read8(p))))))
	case 0x24: // lbu
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		c.set(rt, uint64(c.read8(p)))
	case 0x21: // lh
		p, ok := align(2)
		if !ok {
			return
		}
		c.set(rt, uint64(int64(int16(uint16(c.read16(p))))))
	case 0x25: // lhu
		p, ok := align(2)
		if !ok {
			return
		}
		c.set(rt, uint64(c.read16(p)))
	case 0x23: // lw
		p, ok := align(4)
		if !ok {
			return
		}
		c.set(rt, sext32(c.read32(p)))
	case 0x27: // lwu — zero-extends, unlike lw
		p, ok := align(4)
		if !ok {
			return
		}
		c.set(rt, uint64(c.read32(p)))
	case 0x37: // ld
		p, ok := align(8)
		if !ok {
			return
		}
		c.set(rt, c.read64(p))

	case 0x30: // ll — a load that arms the store-conditional
		p, ok := align(4)
		if !ok {
			return
		}
		c.set(rt, sext32(c.read32(p)))
		c.COP0[cop0LLAddr] = uint64(p >> 4)
		c.LLBit = true
	case 0x34: // lld
		p, ok := align(8)
		if !ok {
			return
		}
		c.set(rt, c.read64(p))
		c.COP0[cop0LLAddr] = uint64(p >> 4)
		c.LLBit = true

	case 0x22: // lwl — the bytes from vaddr to the end of the word, left-aligned
		p, ok := c.Translate(vaddr&^3, false)
		if !ok {
			return
		}
		shift := (vaddr & 3) * 8
		word := c.read32(p)
		cur := uint32(c.R[rt])
		c.set(rt, sext32((cur&(uint32(1)<<shift-1))|(word<<shift)))
	case 0x26: // lwr — the bytes from the start of the word to vaddr, right-aligned
		p, ok := c.Translate(vaddr&^3, false)
		if !ok {
			return
		}
		shift := (3 - vaddr&3) * 8
		word := c.read32(p)
		cur := uint32(c.R[rt])
		c.set(rt, sext32((cur&^(uint32(0xFFFFFFFF)>>shift))|(word>>shift)))

	case 0x1A: // ldl
		p, ok := c.Translate(vaddr&^7, false)
		if !ok {
			return
		}
		shift := (vaddr & 7) * 8
		d := c.read64(p)
		c.set(rt, (c.R[rt]&(uint64(1)<<shift-1))|(d<<shift))
	case 0x1B: // ldr
		p, ok := c.Translate(vaddr&^7, false)
		if !ok {
			return
		}
		shift := (7 - vaddr&7) * 8
		d := c.read64(p)
		c.set(rt, (c.R[rt]&^(^uint64(0)>>shift))|(d>>shift))
	}
}

func (c *CPU) storeOp(op, rs, rt uint32, simm uint64) {
	vaddr := c.R[rs] + simm
	v := c.R[rt]

	align := func(n uint64) (uint32, bool) {
		if vaddr&(n-1) != 0 {
			c.addrError(excAdES, vaddr)
			return 0, false
		}
		return c.Translate(vaddr, true)
	}

	switch op {
	case 0x28: // sb
		p, ok := c.Translate(vaddr, true)
		if !ok {
			return
		}
		c.write8(p, uint32(v)&0xFF)
	case 0x29: // sh
		p, ok := align(2)
		if !ok {
			return
		}
		c.write16(p, uint32(v)&0xFFFF)
	case 0x2B: // sw
		p, ok := align(4)
		if !ok {
			return
		}
		c.write32(p, uint32(v))
	case 0x3F: // sd
		p, ok := align(8)
		if !ok {
			return
		}
		c.write64(p, v)

	case 0x38: // sc — stores only if the load-linked is still valid
		p, ok := align(4)
		if !ok {
			return
		}
		if c.LLBit {
			c.write32(p, uint32(v))
		}
		c.set(rt, b2u(c.LLBit))
		c.LLBit = false
	case 0x3C: // scd
		p, ok := align(8)
		if !ok {
			return
		}
		if c.LLBit {
			c.write64(p, v)
		}
		c.set(rt, b2u(c.LLBit))
		c.LLBit = false

	case 0x2A: // swl — rt's high bytes, into the word from vaddr onward
		p, ok := c.Translate(vaddr&^3, true)
		if !ok {
			return
		}
		shift := (vaddr & 3) * 8
		word := c.read32(p)
		c.write32(p, (word&^(uint32(0xFFFFFFFF)>>shift))|(uint32(v)>>shift))
	case 0x2E: // swr — rt's low bytes, into the word up to vaddr
		p, ok := c.Translate(vaddr&^3, true)
		if !ok {
			return
		}
		shift := (3 - vaddr&3) * 8
		word := c.read32(p)
		c.write32(p, (word&^(uint32(0xFFFFFFFF)<<shift))|(uint32(v)<<shift))

	case 0x2C: // sdl
		p, ok := c.Translate(vaddr&^7, true)
		if !ok {
			return
		}
		shift := (vaddr & 7) * 8
		c.write64(p, (c.read64(p)&^(^uint64(0)>>shift))|(v>>shift))
	case 0x2D: // sdr
		p, ok := c.Translate(vaddr&^7, true)
		if !ok {
			return
		}
		shift := (7 - vaddr&7) * 8
		c.write64(p, (c.read64(p)&^(^uint64(0)<<shift))|(v<<shift))
	}
}

// --- overflow helpers -------------------------------------------------------

func addOv32(a, b uint32) (uint32, bool) {
	r := a + b
	return r, (a^r)&(b^r)&0x80000000 != 0
}
func subOv32(a, b uint32) (uint32, bool) {
	r := a - b
	return r, (a^b)&(a^r)&0x80000000 != 0
}
func addOv64(a, b uint64) (uint64, bool) {
	r := a + b
	return r, (a^r)&(b^r)&0x8000000000000000 != 0
}
func subOv64(a, b uint64) (uint64, bool) {
	r := a - b
	return r, (a^b)&(a^r)&0x8000000000000000 != 0
}

func b2u(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}
