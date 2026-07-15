package gcdsp

// exec.go steps the DSP: fetch the instruction at PC, run it, advance. It implements the
// control flow, the loads and stores, the address-register arithmetic, the status/mode ops and
// the immediate ALU ops — everything the microcode's setup and command-dispatch need — and
// halts loudly, naming the opcode, on anything not yet modelled, so the remaining ops are added
// as real execution reaches them rather than guessed at in advance.
//
// Memory splits three ways. Instructions come from IRAM (0x0000) or the absent IROM (0x8000).
// Data reads and writes hit DRAM (0x0000), the absent coefficient DROM (0x1000), or — at the
// top of the space, 0xFF00 and up — the hardware registers the host models behind the Bus:
// the mailboxes, the DMA engine, the sample accelerator.

// Step runs one instruction. It returns false once the core has halted.
func (c *CPU) Step() bool {
	if c.Halted {
		return false
	}
	pc := c.PC
	op := c.imem(pc)
	span := c.execute(pc, op)
	if c.Halted {
		return false
	}
	// A branch or return sets branched and leaves PC where it wants; otherwise advance past
	// this instruction (one or two words).
	branched := c.Branched
	c.Branched = false
	if !branched {
		c.PC = pc + span
	}
	// A hardware loop whose last instruction just ran jumps back or falls through — unless that
	// instruction itself branched away.
	c.serviceLoops(pc, branched)
	return !c.Halted
}

// execute runs the instruction word op fetched at pc and returns its length in words. It sets
// c.Branched when it has already placed PC at the destination.
func (c *CPU) execute(pc, op uint16) (span uint16) {
	switch {
	case op == 0x0000: // nop
		return 1
	case op == 0x0021: // halt — the ucode stopping itself
		c.Halt("ucode executed HALT at 0x%04X", pc)
		return 1

	// --- address-register arithmetic -----------------------------------------------------
	case op&0xFFFC == 0x0004: // dar
		n := int(op & 3)
		c.arStep(n, -1)
		return 1
	case op&0xFFFC == 0x0008: // iar
		n := int(op & 3)
		c.arStep(n, +1)
		return 1
	case op&0xFFFC == 0x000C: // subarn
		n := int(op & 3)
		c.arStep(n, -int(int16(c.Reg[regIX0+n])))
		return 1
	case op&0xFFF0 == 0x0010: // addarn ar[d] += ix[s]
		d := int(op & 3)
		s := int((op >> 2) & 3)
		c.arStep(d, int(int16(c.Reg[regIX0+s])))
		return 1

	// --- hardware loops ------------------------------------------------------------------
	case op&0xFFE0 == 0x0040: // loop $R
		c.startLoop(pc+1, pc+1, c.Reg[op&0x1F])
		return 1
	case op&0xFFE0 == 0x0060: // bloop $R, end
		end := c.imem(pc + 1)
		c.startLoop(pc+2, end, c.Reg[op&0x1F])
		return 2
	case op&0xFF00 == 0x1000: // loopi #I
		c.startLoop(pc+1, pc+1, op&0xFF)
		return 1
	case op&0xFF00 == 0x1100: // bloopi #I, end
		end := c.imem(pc + 1)
		c.startLoop(pc+2, end, op&0xFF)
		return 2

	// --- status-bit set/clear ------------------------------------------------------------
	case op&0xFF00 == 0x1200: // sbclr #b  — clears SR bit (6+b), the mode/config bits
		c.Reg[regSR] &^= 1 << (6 + (op & 0xFF))
		return 1
	case op&0xFF00 == 0x1300: // sbset #b
		c.Reg[regSR] |= 1 << (6 + (op & 0xFF))
		return 1

	// --- immediate / direct loads and stores ---------------------------------------------
	case op&0xFFE0 == 0x0080: // lri reg, #imm
		c.setReg(op&0x1F, c.imem(pc+1))
		return 2
	case op&0xFFE0 == 0x00C0: // lr reg, @addr
		c.setReg(op&0x1F, c.dataRead(c.imem(pc+1)))
		return 2
	case op&0xFFE0 == 0x00E0: // sr @addr, reg
		c.dataWrite(c.imem(pc+1), c.getReg(op&0x1F))
		return 2
	case op&0xFF00 == 0x1600: // si @M, #imm  — store immediate to 0xFF00|M
		c.dataWrite(0xFF00|(op&0xFF), c.imem(pc+1))
		return 2
	case op&0xF800 == 0x2000: // lrs reg(0x18+r), @0xFF00|M
		c.setReg(0x18+((op>>8)&7), c.dataRead(0xFF00|(op&0xFF)))
		return 1
	case op&0xF800 == 0x2800: // srs @0xFF00|M, reg(0x18+r)
		c.dataWrite(0xFF00|(op&0xFF), c.getReg(0x18+((op>>8)&7)))
		return 1

	// --- register-indirect loads and stores (via an address register) --------------------
	case op&0xFF00 == 0x1900: // lrr reg, @arS
		s := int((op >> 5) & 3)
		c.setReg(op&0x1F, c.dataRead(c.Reg[regAR0+s]))
		return 1
	case op&0xFF00 == 0x1A00: // lrri/lrrd reg, @arS (post-increment/decrement)
		s := int((op >> 5) & 3)
		c.setReg(op&0x1F, c.dataRead(c.Reg[regAR0+s]))
		if op&0x0080 != 0 { // bit 7 chooses decrement vs increment in this row
			c.arStep(s, -1)
		} else {
			c.arStep(s, +1)
		}
		return 1
	case op&0xFF00 == 0x1B00: // srr @arS, reg
		s := int((op >> 5) & 3)
		c.dataWrite(c.Reg[regAR0+s], c.getReg(op&0x1F))
		return 1

	// --- register to register ------------------------------------------------------------
	case op&0xFC00 == 0x1C00: // mrr d, s
		c.setReg((op>>5)&0x1F, c.getReg(op&0x1F))
		return 1

	// --- immediate ALU to accumulator middle ---------------------------------------------
	case op&0xFEFF == 0x0200: // addi acD, #imm  (sign-extended into the 40-bit accumulator)
		d := int((op >> 8) & 1)
		c.setAc(d, c.ac(d)+int64(int16(c.imem(pc+1)))<<16)
		c.setArithFlags(d)
		return 2
	case op&0xFEFF == 0x0220: // xori acD.m
		d := int((op >> 8) & 1)
		c.Reg[regAC0M+d] ^= c.imem(pc + 1)
		c.setLogicFlags(d)
		return 2
	case op&0xFEFF == 0x0240: // andi acD.m
		d := int((op >> 8) & 1)
		c.Reg[regAC0M+d] &= c.imem(pc + 1)
		c.setLogicFlags(d)
		return 2
	case op&0xFEFF == 0x0260: // ori acD.m
		d := int((op >> 8) & 1)
		c.Reg[regAC0M+d] |= c.imem(pc + 1)
		c.setLogicFlags(d)
		return 2
	case op&0xFEFF == 0x0280: // cmpi acD, #imm — compare, set flags only
		d := int((op >> 8) & 1)
		c.subFlags(c.ac(d), int64(int16(c.imem(pc+1)))<<16)
		return 2
	case op&0xFEFF == 0x02A0: // andf acD.m — test AND, set logic-zero flag
		d := int((op >> 8) & 1)
		c.setFlag(srLogicZero, c.Reg[regAC0M+d]&c.imem(pc+1) == 0)
		return 2
	case op&0xFEFF == 0x02C0: // andcf acD.m, #I — test AND, set the logic-zero flag when ALL the
		// immediate's bits are present in acD.m, i.e. (acD.m & I) == I. It is the mirror of andf
		// (0x02A0), which sets the flag when NONE are (the AND is zero); together they are the
		// DSP's "spin until these bits are set / clear" idiom. The microcode's wait for the CPU
		// mailbox present bit (andcf ac0.m, #0x8000; jmplnz) is exactly this test — read as an OR
		// it never sees the bit and the mailbox handshake spins forever.
		d := int((op >> 8) & 1)
		imm := c.imem(pc + 1)
		c.setFlag(srLogicZero, c.Reg[regAC0M+d]&imm == imm)
		return 2

	// --- branches, calls, returns --------------------------------------------------------
	case op&0xFFF0 == 0x0290: // jmp cc, addr
		dst := c.imem(pc + 1)
		if c.cond(op & 0xF) {
			c.PC = dst
			c.Branched = true
		}
		return 2
	case op&0xFFF0 == 0x02B0: // call cc, addr
		dst := c.imem(pc + 1)
		if c.cond(op & 0xF) {
			c.push(regST0, pc+2)
			c.PC = dst
			c.Branched = true
		}
		return 2
	case op&0xFFF0 == 0x02D0: // ret cc
		if c.cond(op & 0xF) {
			c.PC = c.pop(regST0)
			c.Branched = true
		}
		return 1
	case op&0xFFF0 == 0x02F0: // rti cc — return from interrupt
		if c.cond(op & 0xF) {
			c.PC = c.pop(regST0)
			c.Branched = true
			c.InInterrupt = false
		}
		return 1
	case op&0xFFF0 == 0x1700: // jmpr $R
		if c.cond(op & 0xF) {
			c.PC = c.Reg[(op>>5)&7]
			c.Branched = true
		}
		return 1
	case op&0xFFF0 == 0x1710: // callr $R
		if c.cond(op & 0xF) {
			c.push(regST0, pc+1)
			c.PC = c.Reg[(op>>5)&7]
			c.Branched = true
		}
		return 1
	}

	// --- arithmetic ops (0x4000 and up), operation in the high bits, a parallel extension in
	// the low byte. The extension runs alongside the main op. -----------------------------
	if op >= 0x4000 {
		return c.execArith(pc, op)
	}

	c.Halt("unmodelled DSP instruction 0x%04X at 0x%04X (%s)", op, pc, mustText(c.imem, pc))
	return 1
}

// execArith handles the 0x4000+ arithmetic family. Only the ops the microcode has been seen to
// run are implemented; the rest halt loudly with their disassembly so they can be added when a
// run reaches them.
func (c *CPU) execArith(pc, op uint16) uint16 {
	switch {
	case op&0xFF00 == 0x8000: // nx — no arithmetic, extension only
		c.runExt(op & 0xFF)
		return 1
	case op&0xFF00 == 0x8A00: // m2 — multiply results shift left one (x2 mode)
		c.setFlag(srMulShift, true)
		return 1
	case op&0xFF00 == 0x8B00: // m0 — multiply results not shifted
		c.setFlag(srMulShift, false)
		return 1
	case op&0xFF00 == 0x8C00: // clr15 — operands signed
		c.setFlag(srMulSigned, false)
		return 1
	case op&0xFF00 == 0x8D00: // set15 — operands unsigned
		c.setFlag(srMulSigned, true)
		return 1
	case op&0xFF00 == 0x8E00: // set40 — 40-bit accumulator mode
		c.setFlag(srMode40, true)
		return 1
	case op&0xFF00 == 0x8F00: // set16 — 16-bit accumulator saturation mode
		c.setFlag(srMode40, false)
		return 1
	case op&0xF700 == 0x8100: // clr acR
		r := int((op >> 11) & 1)
		c.setAc(r, 0)
		c.setArithFlags(r)
		c.runExt(op & 0xFF)
		return 1
	}
	c.Halt("unmodelled DSP arithmetic op 0x%04X at 0x%04X (%s)", op, pc, mustText(c.imem, pc))
	return 1
}

// runExt executes a parallel extension carried in an arithmetic op's low byte. Only the empty
// extension is a no-op; any real extension halts until modelled, so a silent wrong move never
// slips through.
func (c *CPU) runExt(ext uint16) {
	if ext == 0 {
		return
	}
	c.Halt("unmodelled parallel extension 0x%02X at 0x%04X", ext, c.PC)
}

// mustText disassembles for a halt message; it never fails.
func mustText(read func(uint16) uint16, pc uint16) string {
	t, _ := Disasm(read, pc)
	return t
}
