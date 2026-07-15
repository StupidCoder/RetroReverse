package r5900

// exec.go is the instruction interpreter. See cpu.go for the branch delay slot,
// cop0.go for address translation, mmi.go for the SIMD unit and fpu.go for COP1.
//
// Three rules govern every operation on this core, and all three are easy to get
// silently wrong:
//
//   - A 32-bit result is written back *sign-extended* into the low 64 bits of the
//     register. `addu $t0, $t1, $t2` on values that fit in 32 bits still sets bits
//     63..32 of $t0 to the sign of the result. Only the `d`-prefixed ops are truly
//     64-bit.
//   - A 64-bit operation leaves the register's *upper* 64 bits alone. Those belong
//     to the 128-bit world of lq/sq and the MMI instructions, and a daddu must not
//     disturb them.
//   - A 32-bit shift or arithmetic op reads only the low half of its operands.

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

// doBranchLikely is the branch-likely form: when the branch is not taken the delay
// slot is *annulled* — skipped entirely — by stepping PC over it. When taken it
// behaves exactly like doBranch.
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
	case 0x1C:
		c.mmi(w, rs, rt, (w>>11)&31, (w>>6)&31)

	case 0x02: // j
		c.doBranch(true, jumpT)
	case 0x03: // jal
		c.set(31, c.curPC+8)
		c.doBranch(true, jumpT)
	case 0x04: // beq
		c.doBranch(c.R[rs].Lo == c.R[rt].Lo, branchT)
	case 0x05: // bne
		c.doBranch(c.R[rs].Lo != c.R[rt].Lo, branchT)
	case 0x06: // blez
		c.doBranch(int64(c.R[rs].Lo) <= 0, branchT)
	case 0x07: // bgtz
		c.doBranch(int64(c.R[rs].Lo) > 0, branchT)

	case 0x14: // beql
		c.doBranchLikely(c.R[rs].Lo == c.R[rt].Lo, branchT)
	case 0x15: // bnel
		c.doBranchLikely(c.R[rs].Lo != c.R[rt].Lo, branchT)
	case 0x16: // blezl
		c.doBranchLikely(int64(c.R[rs].Lo) <= 0, branchT)
	case 0x17: // bgtzl
		c.doBranchLikely(int64(c.R[rs].Lo) > 0, branchT)

	case 0x08: // addi (trapping)
		if r, ov := addOv32(uint32(c.R[rs].Lo), uint32(simm)); ov {
			c.Exception(excOv)
		} else {
			c.set(rt, sext32(r))
		}
	case 0x09: // addiu
		c.set(rt, sext32(uint32(c.R[rs].Lo)+uint32(simm)))
	case 0x0A: // slti
		c.set(rt, b2u(int64(c.R[rs].Lo) < int64(simm)))
	case 0x0B: // sltiu
		c.set(rt, b2u(c.R[rs].Lo < simm))
	case 0x0C: // andi
		c.set(rt, c.R[rs].Lo&uint64(imm))
	case 0x0D: // ori
		c.set(rt, c.R[rs].Lo|uint64(imm))
	case 0x0E: // xori
		c.set(rt, c.R[rs].Lo^uint64(imm))
	case 0x0F: // lui
		c.set(rt, sext32(imm<<16))

	case 0x18: // daddi (trapping)
		if r, ov := addOv64(c.R[rs].Lo, simm); ov {
			c.Exception(excOv)
		} else {
			c.set(rt, r)
		}
	case 0x19: // daddiu
		c.set(rt, c.R[rs].Lo+simm)

	case 0x10:
		c.cop0(w, rs, rt, (w>>11)&31)
	case 0x11:
		if c.COP0[cop0Status]&statusCU1 == 0 {
			c.coprocessorUnusable(1)
			return
		}
		c.cop1(w, rs, rt, (w>>11)&31, (w>>6)&31, branchT)
	case 0x12:
		c.cop2(w, rs, rt, (w>>11)&31, branchT)

	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x1A, 0x1B, 0x30, 0x34, 0x37, 0x1E:
		c.loadOp(op, rs, rt, simm)
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x38, 0x3C, 0x3F, 0x1F:
		c.storeOp(op, rs, rt, simm)

	case 0x2F, 0x33:
		// cache and pref: the caches are not modelled, so invalidation and
		// prefetch are no-ops. The opcodes must exist — compiled code brackets
		// every DMA with a cache writeback, and halting here would stop the boot.

	case 0x31, 0x39: // lwc1 / swc1
		if c.COP0[cop0Status]&statusCU1 == 0 {
			c.coprocessorUnusable(1)
			return
		}
		c.cop1Mem(op, rs, rt, simm)

	case 0x36, 0x3E: // lqc2 / sqc2 — a whole vector register to or from memory
		c.cop2Mem(op, rs, rt, simm)

	default:
		// Every encoding the R5900 does not define raises a reserved-instruction
		// fault rather than halting the core. A program may execute one on purpose
		// to probe the machine, and halting would stop code that real hardware runs.
		c.Exception(excRI)
	}
}

// cop2 is VU0 in macro mode: the moves between the EE's registers and the vector
// unit's, the branches on its condition flags, and the vector operations, which are
// handed to tools/cpu/vu.
func (c *CPU) cop2(w, rs, rt, rd uint32, branchT uint64) {
	if c.COP0[cop0Status]&statusCU2 == 0 {
		c.coprocessorUnusable(2)
		return
	}
	if c.COP2 == nil {
		// No vector unit attached. Writes and operations are dropped, but the READS must
		// still write their destination: cfc2 and qmfc2 are register writes whatever the
		// coprocessor holds, and skipping them leaves the destination with whatever it held
		// before — a stale value that then gets believed. Jak and Daxter's sceGsSyncPath
		// polls `cfc2 $a2, $vi29` (VPU_STAT) for VU1's busy bit, and with the write dropped
		// it read its own leftover $a2, whose bit 8 happened to be set, and concluded "VU1
		// does not terminate" every frame — a vector unit that did not exist reported busy
		// forever. Zero is the honest answer this machine can give: no VU is attached, so
		// no VU is running.
		switch rs {
		case 0x01: // qmfc2
			c.setQ(rt, 0, 0)
		case 0x02: // cfc2
			c.set(rt, 0)
		}
		return
	}

	if w&(1<<25) != 0 { // a vector operation
		c.COP2.Macro(w)
		return
	}

	switch rs {
	case 0x01: // qmfc2 — a whole 128-bit vector register into a GPR
		c.setQ(rt,
			uint64(c.COP2.ReadVF(rd, 0))|uint64(c.COP2.ReadVF(rd, 1))<<32,
			uint64(c.COP2.ReadVF(rd, 2))|uint64(c.COP2.ReadVF(rd, 3))<<32)
	case 0x02: // cfc2
		c.set(rt, sext32(c.COP2.ReadCtrl(rd)))
	case 0x05: // qmtc2
		q := c.R[rt]
		c.COP2.WriteVF(rd, 0, uint32(q.Lo))
		c.COP2.WriteVF(rd, 1, uint32(q.Lo>>32))
		c.COP2.WriteVF(rd, 2, uint32(q.Hi))
		c.COP2.WriteVF(rd, 3, uint32(q.Hi>>32))
	case 0x06: // ctc2
		c.COP2.WriteCtrl(rd, uint32(c.R[rt].Lo))
	case 0x08: // bc2f / bc2t / bc2fl / bc2tl — on VU0's condition flag
		cond := c.COP2.ReadCtrl(vuCtrlCMSAR) != 0
		switch rt & 3 {
		case 0:
			c.doBranch(!cond, branchT)
		case 1:
			c.doBranch(cond, branchT)
		case 2:
			c.doBranchLikely(!cond, branchT)
		case 3:
			c.doBranchLikely(cond, branchT)
		}
	default:
		c.Exception(excRI)
	}
}

// vuCtrlCMSAR is the VU control register the COP2 branches test. It is named here
// rather than in tools/cpu/vu because the EE's branch instructions are the only
// thing on this side of the interface that needs it.
const vuCtrlCMSAR = 18

// cop2Mem is lqc2 / sqc2: a whole vector register loaded from or stored to memory.
func (c *CPU) cop2Mem(op, rs, rt uint32, simm uint64) {
	if c.COP2 == nil {
		return
	}
	vaddr := c.R[rs].Lo + simm
	if vaddr&15 != 0 {
		code := uint32(excAdEL)
		if op == 0x3E {
			code = excAdES
		}
		c.addrError(code, vaddr)
		return
	}
	switch op {
	case 0x36: // lqc2
		p, ok := c.Translate(vaddr, false)
		if !ok {
			return
		}
		for i := uint32(0); i < 4; i++ {
			c.COP2.WriteVF(rt, i, c.read32(p+i*4))
		}
	case 0x3E: // sqc2
		p, ok := c.Translate(vaddr, true)
		if !ok {
			return
		}
		for i := uint32(0); i < 4; i++ {
			c.write32(p+i*4, c.COP2.ReadVF(rt, i))
		}
	}
}

func (c *CPU) special(w, rs, rt uint32) {
	rd := (w >> 11) & 31
	shamt := (w >> 6) & 31
	funct := w & 63

	switch funct {
	case 0x00: // sll (also nop)
		c.set(rd, sext32(uint32(c.R[rt].Lo)<<shamt))
	case 0x02: // srl
		c.set(rd, sext32(uint32(c.R[rt].Lo)>>shamt))
	case 0x03: // sra
		// The arithmetic right shifts read the *whole* 64-bit half, not its low 32
		// bits, and only then sign-extend the 32-bit result. A core that shifts the
		// low word alone gives a different answer whenever the high word is not
		// already the sign extension of the low one — which is exactly what happens
		// after a 64-bit operation.
		c.set(rd, sext32(uint32(int64(c.R[rt].Lo)>>shamt)))
	case 0x04: // sllv
		c.set(rd, sext32(uint32(c.R[rt].Lo)<<(c.R[rs].Lo&31)))
	case 0x06: // srlv
		c.set(rd, sext32(uint32(c.R[rt].Lo)>>(c.R[rs].Lo&31)))
	case 0x07: // srav
		c.set(rd, sext32(uint32(int64(c.R[rt].Lo)>>(c.R[rs].Lo&31))))

	case 0x08: // jr
		c.doBranch(true, c.R[rs].Lo)
	case 0x09: // jalr
		c.set(rd, c.curPC+8)
		c.doBranch(true, c.R[rs].Lo)

	// The MIPS IV conditional moves. Note the test is on the whole 64-bit half.
	case 0x0A: // movz
		if c.R[rt].Lo == 0 {
			c.set(rd, c.R[rs].Lo)
		}
	case 0x0B: // movn
		if c.R[rt].Lo != 0 {
			c.set(rd, c.R[rs].Lo)
		}

	case 0x0C: // syscall
		// The EE kernel lives in a BIOS ROM this model does not have, so the machine
		// may HLE it. If it declines, the architectural exception is taken.
		if c.Syscall != nil && c.Syscall(c) {
			return
		}
		c.Exception(excSys)
	case 0x0D: // break
		c.Exception(excBp)
	case 0x0F: // sync — the store buffer is not modelled

	case 0x10: // mfhi
		c.set(rd, c.HI)
	case 0x11: // mthi
		c.HI = c.R[rs].Lo
	case 0x12: // mflo
		c.set(rd, c.LO)
	case 0x13: // mtlo
		c.LO = c.R[rs].Lo

	case 0x14: // dsllv
		c.set(rd, c.R[rt].Lo<<(c.R[rs].Lo&63))
	case 0x16: // dsrlv
		c.set(rd, c.R[rt].Lo>>(c.R[rs].Lo&63))
	case 0x17: // dsrav
		c.set(rd, uint64(int64(c.R[rt].Lo)>>(c.R[rs].Lo&63)))

	// mult and multu write a destination register as well as HI/LO on this core —
	// "mult rd, rs, rt" — which the VR4300 does not do.
	case 0x18: // mult
		p := int64(int32(uint32(c.R[rs].Lo))) * int64(int32(uint32(c.R[rt].Lo)))
		c.LO, c.HI = sext32(uint32(p)), sext32(uint32(uint64(p)>>32))
		c.set(rd, c.LO)
	case 0x19: // multu
		p := uint64(uint32(c.R[rs].Lo)) * uint64(uint32(c.R[rt].Lo))
		c.LO, c.HI = sext32(uint32(p)), sext32(uint32(p>>32))
		c.set(rd, c.LO)
	case 0x1A: // div
		c.LO, c.HI = divSigned32(int32(uint32(c.R[rs].Lo)), int32(uint32(c.R[rt].Lo)))
	case 0x1B: // divu
		c.LO, c.HI = divUnsigned32(uint32(c.R[rs].Lo), uint32(c.R[rt].Lo))

	case 0x20: // add (trapping)
		if r, ov := addOv32(uint32(c.R[rs].Lo), uint32(c.R[rt].Lo)); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, sext32(r))
		}
	case 0x21: // addu
		c.set(rd, sext32(uint32(c.R[rs].Lo)+uint32(c.R[rt].Lo)))
	case 0x22: // sub (trapping)
		if r, ov := subOv32(uint32(c.R[rs].Lo), uint32(c.R[rt].Lo)); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, sext32(r))
		}
	case 0x23: // subu
		c.set(rd, sext32(uint32(c.R[rs].Lo)-uint32(c.R[rt].Lo)))
	case 0x24: // and
		c.set(rd, c.R[rs].Lo&c.R[rt].Lo)
	case 0x25: // or
		c.set(rd, c.R[rs].Lo|c.R[rt].Lo)
	case 0x26: // xor
		c.set(rd, c.R[rs].Lo^c.R[rt].Lo)
	case 0x27: // nor
		c.set(rd, ^(c.R[rs].Lo | c.R[rt].Lo))

	case 0x28: // mfsa
		c.set(rd, uint64(c.SA))
	case 0x29: // mtsa
		c.SA = uint32(c.R[rs].Lo) & 0x7F

	case 0x2A: // slt
		c.set(rd, b2u(int64(c.R[rs].Lo) < int64(c.R[rt].Lo)))
	case 0x2B: // sltu
		c.set(rd, b2u(c.R[rs].Lo < c.R[rt].Lo))

	case 0x2C: // dadd (trapping)
		if r, ov := addOv64(c.R[rs].Lo, c.R[rt].Lo); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x2D: // daddu
		c.set(rd, c.R[rs].Lo+c.R[rt].Lo)
	case 0x2E: // dsub (trapping)
		if r, ov := subOv64(c.R[rs].Lo, c.R[rt].Lo); ov {
			c.Exception(excOv)
		} else {
			c.set(rd, r)
		}
	case 0x2F: // dsubu
		c.set(rd, c.R[rs].Lo-c.R[rt].Lo)

	case 0x30: // tge
		c.trapIf(int64(c.R[rs].Lo) >= int64(c.R[rt].Lo))
	case 0x31: // tgeu
		c.trapIf(c.R[rs].Lo >= c.R[rt].Lo)
	case 0x32: // tlt
		c.trapIf(int64(c.R[rs].Lo) < int64(c.R[rt].Lo))
	case 0x33: // tltu
		c.trapIf(c.R[rs].Lo < c.R[rt].Lo)
	case 0x34: // teq
		c.trapIf(c.R[rs].Lo == c.R[rt].Lo)
	case 0x36: // tne
		c.trapIf(c.R[rs].Lo != c.R[rt].Lo)

	case 0x38: // dsll
		c.set(rd, c.R[rt].Lo<<shamt)
	case 0x3A: // dsrl
		c.set(rd, c.R[rt].Lo>>shamt)
	case 0x3B: // dsra
		c.set(rd, uint64(int64(c.R[rt].Lo)>>shamt))
	case 0x3C: // dsll32 — shift amounts 32..63
		c.set(rd, c.R[rt].Lo<<(shamt+32))
	case 0x3E: // dsrl32
		c.set(rd, c.R[rt].Lo>>(shamt+32))
	case 0x3F: // dsra32
		c.set(rd, uint64(int64(c.R[rt].Lo)>>(shamt+32)))

	default:
		c.Exception(excRI) // an undefined SPECIAL function
	}
}

func (c *CPU) regimm(w, rs, rt uint32, branchT, simm uint64) {
	s := int64(c.R[rs].Lo)
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
		c.trapIf(c.R[rs].Lo >= simm)
	case 0x0A: // tlti
		c.trapIf(s < int64(simm))
	case 0x0B: // tltiu
		c.trapIf(c.R[rs].Lo < simm)
	case 0x0C: // teqi
		c.trapIf(c.R[rs].Lo == simm)
	case 0x0E: // tnei
		c.trapIf(c.R[rs].Lo != simm)

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

	// mtsab/mtsah set the shift-amount register, which only qfsrv reads. The byte
	// form counts bytes, the halfword form halfwords, and both land in SA as a *bit*
	// count — so a qfsrv never has to know which one set it.
	case 0x18: // mtsab
		c.SA = ((uint32(c.R[rs].Lo) & 0xF) ^ (uint32(simm) & 0xF)) * 8
	case 0x19: // mtsah
		c.SA = ((uint32(c.R[rs].Lo) & 0x7) ^ (uint32(simm) & 0x7)) * 16

	default:
		c.Exception(excRI) // an undefined REGIMM function
	}
}

func (c *CPU) trapIf(cond bool) {
	if cond {
		c.Exception(excTrap)
	}
}

// --- loads and stores -------------------------------------------------------
//
// The unaligned load/store family (lwl/lwr/ldl/ldr and their store forms) reads or
// writes the part of an aligned word or doubleword that lies on the same side of
// the address as the byte selected. On a little-endian machine "left" is the
// high-order end, which is the *far* end from the addressed byte — the mirror of
// the big-endian case in tools/cpu/r4300.

func (c *CPU) loadOp(op, rs, rt uint32, simm uint64) {
	vaddr := c.R[rs].Lo + simm

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

	case 0x1E: // lq — 128 bits. The address is *forced* aligned rather than faulting.
		p, ok := c.Translate(vaddr&^15, false)
		if !ok {
			return
		}
		q := c.read128(p)
		c.setQ(rt, q.Lo, q.Hi)

	case 0x30: // ll — a load that arms the store-conditional
		p, ok := align(4)
		if !ok {
			return
		}
		c.set(rt, sext32(c.read32(p)))
		c.LLBit = true
	case 0x34: // lld
		p, ok := align(8)
		if !ok {
			return
		}
		c.set(rt, c.read64(p))
		c.LLBit = true

	case 0x22: // lwl — the bytes from vaddr up to the top of the word
		p, ok := c.Translate(vaddr&^3, false)
		if !ok {
			return
		}
		shift := (3 - vaddr&3) * 8
		word := c.read32(p)
		cur := uint32(c.R[rt].Lo)
		c.set(rt, sext32((cur&(uint32(1)<<shift-1))|(word<<shift)))
	case 0x26: // lwr — the bytes from the bottom of the word up to vaddr
		p, ok := c.Translate(vaddr&^3, false)
		if !ok {
			return
		}
		shift := (vaddr & 3) * 8
		word := c.read32(p)
		cur := uint32(c.R[rt].Lo)
		c.set(rt, sext32((cur&^(uint32(0xFFFFFFFF)>>shift))|(word>>shift)))

	case 0x1A: // ldl
		p, ok := c.Translate(vaddr&^7, false)
		if !ok {
			return
		}
		shift := (7 - vaddr&7) * 8
		d := c.read64(p)
		c.set(rt, (c.R[rt].Lo&(uint64(1)<<shift-1))|(d<<shift))
	case 0x1B: // ldr
		p, ok := c.Translate(vaddr&^7, false)
		if !ok {
			return
		}
		shift := (vaddr & 7) * 8
		d := c.read64(p)
		c.set(rt, (c.R[rt].Lo&^(^uint64(0)>>shift))|(d>>shift))
	}
}

func (c *CPU) storeOp(op, rs, rt uint32, simm uint64) {
	vaddr := c.R[rs].Lo + simm
	v := c.R[rt].Lo

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

	case 0x1F: // sq — 128 bits, with the address forced aligned
		p, ok := c.Translate(vaddr&^15, true)
		if !ok {
			return
		}
		c.write128(p, c.R[rt])

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

	case 0x2A: // swl — rt's low bytes, into the word from vaddr up to its top
		p, ok := c.Translate(vaddr&^3, true)
		if !ok {
			return
		}
		shift := (3 - vaddr&3) * 8
		word := c.read32(p)
		c.write32(p, (word&^(uint32(0xFFFFFFFF)>>shift))|(uint32(v)>>shift))
	case 0x2E: // swr — rt's low bytes, into the word from its bottom up to vaddr
		p, ok := c.Translate(vaddr&^3, true)
		if !ok {
			return
		}
		shift := (vaddr & 3) * 8
		word := c.read32(p)
		c.write32(p, (word&^(uint32(0xFFFFFFFF)<<shift))|(uint32(v)<<shift))

	case 0x2C: // sdl
		p, ok := c.Translate(vaddr&^7, true)
		if !ok {
			return
		}
		shift := (7 - vaddr&7) * 8
		c.write64(p, (c.read64(p)&^(^uint64(0)>>shift))|(v>>shift))
	case 0x2D: // sdr
		p, ok := c.Translate(vaddr&^7, true)
		if !ok {
			return
		}
		shift := (vaddr & 7) * 8
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
