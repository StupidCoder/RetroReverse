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
		c.arDec(int(op & 3))
		return 1
	case op&0xFFFC == 0x0008: // iar
		c.arInc(int(op & 3))
		return 1
	case op&0xFFFC == 0x000C: // subarn
		n := int(op & 3)
		c.arSub(n, int16(c.Reg[regIX0+n]))
		return 1
	case op&0xFFF0 == 0x0010: // addarn ar[d] += ix[s]
		d := int(op & 3)
		s := int((op >> 2) & 3)
		c.arAdd(d, int16(c.Reg[regIX0+s]))
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
		c.setRegExtend(op&0x1F, c.imem(pc+1))
		return 2
	case op&0xFFE0 == 0x00C0: // lr reg, @addr
		c.setRegExtend(op&0x1F, c.dataRead(c.imem(pc+1)))
		return 2
	case op&0xFFE0 == 0x00E0: // sr @addr, reg
		c.dataWrite(c.imem(pc+1), c.getReg(op&0x1F))
		return 2
	case op&0xFF00 == 0x1600: // si @M, #imm  — store immediate to 0xFF00|M
		c.dataWrite(0xFF00|(op&0xFF), c.imem(pc+1))
		return 2
	case op&0xF800 == 0x2000: // lrs reg(0x18+r), @0xFF00|M
		c.setRegExtend(0x18+((op>>8)&7), c.dataRead(0xFF00|(op&0xFF)))
		return 1
	case op&0xF800 == 0x2800: // srs @0xFF00|M, reg(0x18+r)
		c.dataWrite(0xFF00|(op&0xFF), c.getReg(0x18+((op>>8)&7)))
		return 1

	// --- register-indirect loads and stores (via an address register) --------------------
	case op&0xFC00 == 0x1800:
		// The register-indirect memory family: load or store a register through an address
		// register, with an optional post-modification of that address register. Encoding
		// 0001 10L MN ssd dddd — bit 9 (L) stores rather than loads, bits 8..7 (MN) choose the
		// post-modify (00 none, 01 decrement, 10 increment, 11 add the index register), bits
		// 6..5 select the address register, bits 4..0 the data register. Getting the increment
		// right is what makes the DRAM-clear loop, the two-word address fetches, and the block
		// copies to the hardware registers walk their buffers instead of hammering one cell.
		s := int((op >> 5) & 3)
		reg := op & 0x1F
		addr := c.Reg[regAR0+s]
		if op&0x0200 != 0 { // store
			c.dataWrite(addr, c.getReg(reg))
		} else { // load
			c.setRegExtend(reg, c.dataRead(addr))
		}
		switch op & 0x0180 {
		case 0x0080: // post-decrement
			c.arDec(s)
		case 0x0100: // post-increment
			c.arInc(s)
		case 0x0180: // post-add the index register
			c.arAdd(s, int16(c.Reg[regIX0+s]))
		}
		return 1

	// --- register to register ------------------------------------------------------------
	case op&0xFC00 == 0x1C00: // mrr d, s
		c.setRegExtend((op>>5)&0x1F, c.getReg(op&0x1F))
		return 1

	// --- accumulator shift by immediate --------------------------------------------------
	case op&0xFE00 == 0x1400: // shifti: LSL/LSR/ASL/ASR acR, #shift
		// Encoding 0001 010r aiii iiii: bit 8 selects the accumulator, bit 7 the arithmetic
		// (sign-preserving) vs logical form, and bits 6..0 are a SIGNED 7-bit shift amount —
		// positive shifts the 40-bit accumulator left, negative shifts it right by the
		// magnitude. So 0x1479 (a=0, imm=0x79=-7) is a logical right shift by 7, and 0x1501
		// (r=1, imm=+1) a logical left shift by 1. The command dispatch at ucode 0x0040 leans
		// on exactly this: it normalises the mailbox command word with a shift before masking
		// out the jump-table index, so a wrong amount or direction lands on the wrong entry.
		r := int((op >> 8) & 1)
		arith := op&0x0080 != 0
		amt := int(op & 0x7F)
		if amt&0x40 != 0 { // sign-extend the 7-bit field
			amt -= 0x80
		}
		c.shiftAcc(r, arith, amt)
		return 1

	// --- load short immediate ------------------------------------------------------------
	case op&0xF800 == 0x0800: // lris reg(0x18+d), #imm8 (sign-extended)
		// Encoding 0000 1ddd iiii iiii: an 8-bit signed immediate into one of the eight
		// short registers 0x18..0x1F (the ax and accumulator low/mid words). The DMA-setup
		// helpers use it to drop a small direction constant into ac1.m between programming the
		// DSP-DMA address and control registers.
		c.setRegExtend(0x18+((op>>8)&7), uint16(int16(int8(op&0xFF))))
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

	// --- single-word short-immediate arithmetic (an 8-bit immediate into the middle word) ---
	case op&0xFE00 == 0x0400: // addis acD, #imm8 — add a signed 8-bit immediate, shifted into the
		// accumulator middle word. The mixer builds a second buffer address 0x50 words along by
		// addis-ing 0x50 to a pointer held in an accumulator middle (ucode 0x0464).
		d := int((op >> 8) & 1)
		c.aluAddSub(d, int64(int8(op))<<16, false)
		return 1
	case op&0xFE00 == 0x0600: // cmpis acD, #imm8 — compare the accumulator with imm<<16 (flags only)
		d := int((op >> 8) & 1)
		c.subFlags(c.ac(d), int64(int8(op))<<16)
		return 1

	// --- branches, calls, returns --------------------------------------------------------
	case op&0xFFF0 == 0x0270: // if cc — execute the next instruction only when the condition
		// holds; a false condition skips the WHOLE next instruction (one or two words). The
		// ucode's idioms are conditional single-word steps (ifg incm / ifl decm at 0x0616, the
		// conditional stores at 0x02A5, the abs-value ifg neg at 0x0BDB).
		if !c.cond(op & 0xF) {
			_, span := Disasm(func(a uint16) uint16 { return c.imem(a) }, pc+1)
			return 1 + span
		}
		return 1

	// --- shift the accumulator by a register amount (the 7-bit signed convention: zero low
	// six bits mean no shift regardless of the sign bit; POSITIVE means RIGHT here) ----------
	case op == 0x02CA: // lsrn — ac0 logically shifted by ac1.m
		c.shiftAcc(0, false, -shiftAmount7(c.Reg[regAC1M]))
		return 1
	case op == 0x02CB: // asrn — ac0 arithmetically shifted by ac1.m
		c.shiftAcc(0, true, -shiftAmount7(c.Reg[regAC1M]))
		return 1
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
	case op&0xFF00 == 0x1700: // jmpr/callr $R, cc — the register-indirect control transfer.
		// bits 7..5 select the address register, bit 4 is the call flag (a call stacks the
		// return address), bits 3..0 are the condition. The register holds a code address the
		// ucode computed — the jump-table dispatch at ucode 0x0276 builds it with mrr then callr.
		if c.cond(op & 0xF) {
			if op&0x10 != 0 { // callr: stack the return (this op is one word)
				c.push(regST0, pc+1)
			}
			c.PC = c.Reg[(op>>5)&7]
			c.Branched = true
		}
		return 1
	}

	// --- arithmetic/logic/multiply/move ops (0x3000 and up), operation in the high bits, a
	// parallel extension in the low byte. The extension runs alongside the main op. ---------
	if op >= 0x3000 {
		return c.execArith(pc, op)
	}

	c.Halt("unmodelled DSP instruction 0x%04X at 0x%04X (%s)", op, pc, mustText(c.imem, pc))
	return 1
}

// execArith handles the 0x3000+ family: the logic ops, the accumulator arithmetic, the multiply/
// accumulate ops, and the moves — each of which may carry a parallel extension in its low byte.
// The extension's read phase runs before the main op and its write phase after (see exec_ext.go),
// so a caller runs `p := c.extBegin(op&0xFF)`, the main op, then `p.commit(c)`. Opcodes and their
// operand bit-fields are the documented set (gamecube-tools). Ops not yet reached halt loudly.
func (c *CPU) execArith(pc, op uint16) uint16 {
	// The mode and standalone ops that carry no accumulator write are matched first, before the
	// wider masks below could swallow them. They take no parallel extension.
	switch {
	case op&0xFF00 == 0x8A00: // m2 — products doubled (clears the modify bit; doubling is default)
		c.setFlag(srMulNoDouble, false)
		return 1
	case op&0xFF00 == 0x8B00: // m0 — products as-is
		c.setFlag(srMulNoDouble, true)
		return 1
	case op&0xFF00 == 0x8C00: // clr15 — MULX low halves signed
		c.setFlag(srMulUnsigned, false)
		return 1
	case op&0xFF00 == 0x8D00: // set15 — MULX low halves unsigned
		c.setFlag(srMulUnsigned, true)
		return 1
	case op&0xFF00 == 0x8E00: // set16 — plain mid loads, no read saturation. NOTE: 0x8E is SET16
		// and 0x8F is SET40 (hardware-verified) — the manual's "SET40/16 1000 111x" line reads
		// the other way around and had these swapped here for a while.
		c.setFlag(srMode40, false)
		return 1
	case op&0xFF00 == 0x8F00: // set40 — mid loads extend + clear low, mid reads saturate
		c.setFlag(srMode40, true)
		return 1
	}

	// Everything below may carry a parallel extension: read phase, main op, write phase.
	// The 0x3xxx logic row carries only a SEVEN-bit extension — bit 7 there is opcode, doubling
	// the row (NOT lives at 0x3280 above XORR at 0x3000, and several rows above ANDR/ORR are
	// ops the documentation predates). The proof is in the table masks (every documented 0x3xxx
	// op masks 0x..80) and in the ucode: read with an 8-bit field, `not ac0.m` at 0x075A would
	// carry a phantom `ls` (a store through AR3 that nothing set up), and the interpolator's
	// 0x38C3 would dual-load over the constant its own preamble just placed in ax0.h.
	extBits := op & 0xFF
	if op&0xF000 == 0x3000 {
		extBits = op & 0x7F
	}
	p := c.extBegin(extBits)
	switch {
	case op&0xFF00 == 0x8000: // nx — no arithmetic, the extension is the whole instruction

	case op&0xFF00 == 0x8400: // clrp — clear the product register. The hardware's "zero" is a
		// biased set of the four pieces (manual p.66: l=0x0000, m1=0xfff0, h=0x00ff, m2=0x0010),
		// which sums to zero through prod()'s 40-bit read. The FIR prologue is `clrp : ld`.
		c.Reg[regPRODL] = 0x0000
		c.Reg[regPRODM1] = 0xFFF0
		c.Reg[regPRODH] = 0x00FF
		c.Reg[regPRODM2] = 0x0010

	// --- logic ops on an accumulator middle word (bit 8 selects the accumulator) ---------
	case op&0xFE80 == 0x3280: // not acD.m
		d := int((op >> 8) & 1)
		c.setReg(uint16(regAC0M+d), ^c.Reg[regAC0M+d])
		c.setLogicFlags(d)
	case op&0xFC80 == 0x3000: // xorr acD.m, axS.h
		d := int((op >> 8) & 1)
		c.setReg(uint16(regAC0M+d), c.Reg[regAC0M+d]^c.Reg[regAX0H+((op>>9)&1)])
		c.setLogicFlags(d)
	case op&0xFC80 == 0x3400: // andr acD.m, axS.h
		d := int((op >> 8) & 1)
		c.setReg(uint16(regAC0M+d), c.Reg[regAC0M+d]&c.Reg[regAX0H+((op>>9)&1)])
		c.setLogicFlags(d)
	case op&0xFC80 == 0x3800: // orr acD.m, axS.h
		d := int((op >> 8) & 1)
		c.setReg(uint16(regAC0M+d), c.Reg[regAC0M+d]|c.Reg[regAX0H+((op>>9)&1)])
		c.setLogicFlags(d)
	case op&0xFE80 == 0x3080: // xorc acD.m — XOR with the OTHER accumulator's middle
		d := int((op >> 8) & 1)
		c.setReg(uint16(regAC0M+d), c.Reg[regAC0M+d]^c.Reg[regAC0M+(1-d)])
		c.setLogicFlags(d)

	// --- shift the accumulator by a register amount (bit7=1 rows of the logic space; the
	// 7-bit signed amount convention, POSITIVE meaning LEFT for these — the opposite of
	// lsrn/asrn). The interpolator at ucode 0x013B rides asrnrx with its constant 4 in ax0.h.
	case op&0xFC80 == 0x3480: // lsrnrx acD, axS.h (s=bit9) — logical shift by the ax high
		c.shiftAcc(int((op>>8)&1), false, shiftAmount7(c.Reg[regAX0H+((op>>9)&1)]))
	case op&0xFC80 == 0x3880: // asrnrx acD, axS.h — arithmetic shift by the ax high
		c.shiftAcc(int((op>>8)&1), true, shiftAmount7(c.Reg[regAX0H+((op>>9)&1)]))
	case op&0xFE80 == 0x3C80: // lsrnr acD — logical shift by the OTHER accumulator's middle
		d := int((op >> 8) & 1)
		c.shiftAcc(d, false, shiftAmount7(c.Reg[regAC0M+(1-d)]))
	case op&0xFE80 == 0x3E80: // asrnr acD — arithmetic shift by the other accumulator's middle
		d := int((op >> 8) & 1)
		c.shiftAcc(d, true, shiftAmount7(c.Reg[regAC0M+(1-d)]))

	// --- clr / add / sub of an accumulator or extended register --------------------------
	case op&0xF700 == 0x8100: // clr acR (bit 11)
		r := int((op >> 11) & 1)
		c.setAc(r, 0)
		c.setArithFlags(r)
	case op&0xFE00 == 0x8600: // tstaxh axR.h (bit 8) — flags from the 16-bit HIGH HALF of an ax
		// register, not from an accumulator: the accumulator test (tst acR) is 0xB100, with its
		// selector up in bit 11. Both appear in this ucode. (An earlier build ran 0x8600 as
		// "tst acR" — self-consistent with its own test, silently wrong against the ISA.)
		v := int16(c.Reg[regAX0H+((op>>8)&1)])
		c.setFlag(srZero, v == 0)
		c.setFlag(srSign, v < 0)
		c.setFlag(srCarry, false)
		c.setFlag(srOverflow, false)
	case op&0xF700 == 0xB100: // tst acR (bit 11) — flags from the accumulator vs zero
		c.setTestFlags(int((op >> 11) & 1))
	case op&0xE700 == 0xC100: // cmpaxh acS, axR.h (s=bit11, r=bit12 — NOT the devkitPro table's
		// assignment, which has these two swapped) — compare the accumulator with an ax high
		// half taken into the middle word, the same alignment addr/cmpis use
		c.subFlags(c.ac(int((op>>11)&1)), int64(int16(c.Reg[regAX0H+((op>>12)&1)]))<<16)
	case op&0xF800 == 0x4000: // addr acD, reg — add a register (bits 10..9 -> 0x18+field) into the middle
		d := int((op >> 8) & 1)
		reg := 0x18 + ((op >> 9) & 3)
		c.aluAddSub(d, int64(int16(c.getReg(reg)))<<16, false)
	case op&0xFC00 == 0x4800: // addax acD, axS (s=bit9, d=bit8) — add the full 32-bit ax
		c.aluAddSub(int((op>>8)&1), c.ax(int((op>>9)&1)), false)
	case op&0xFE00 == 0x4C00: // add acD — add the other accumulator
		d := int((op >> 8) & 1)
		c.aluAddSub(d, c.ac(1-d), false)
	case op&0xF800 == 0x5000: // subr acD, reg — subtract a register (into the middle)
		d := int((op >> 8) & 1)
		reg := 0x18 + ((op >> 9) & 3)
		c.aluAddSub(d, int64(int16(c.getReg(reg)))<<16, true)
	case op&0xFC00 == 0x5800: // subax acD, axS — subtract the full 32-bit ax
		c.aluAddSub(int((op>>8)&1), c.ax(int((op>>9)&1)), true)
	case op&0xFE00 == 0x5C00: // sub acD — subtract the other accumulator
		d := int((op >> 8) & 1)
		c.aluAddSub(d, c.ac(1-d), true)
	case op&0xFC00 == 0x7000: // addaxl acD, axS.l — add the low 16 bits of an ax register
		d := int((op >> 8) & 1)
		s := int((op >> 9) & 1)
		c.aluAddSub(d, int64(c.Reg[regAX0L+s]), false)

	// --- single-accumulator steps (0111 0/1 rows) -----------------------------------------
	case op&0xFE00 == 0x7400: // incm acD — +1 in the middle word (a pointer step)
		c.aluAddSub(int((op>>8)&1), 1<<16, false)
	case op&0xFE00 == 0x7600: // inc acD
		c.aluAddSub(int((op>>8)&1), 1, false)
	case op&0xFE00 == 0x7800: // decm acD — −1 in the middle word
		c.aluAddSub(int((op>>8)&1), 1<<16, true)
	case op&0xFE00 == 0x7A00: // dec acD
		c.aluAddSub(int((op>>8)&1), 1, true)
	case op&0xFE00 == 0x7C00: // neg acD
		d := int((op >> 8) & 1)
		c.setAc(d, -c.ac(d))
		c.setArithFlags(d)

	// --- accumulator shifts by a fixed 16 --------------------------------------------------
	case op&0xFE00 == 0xF000: // lsl16 acR
		c.shiftAcc(int((op>>8)&1), false, 16)
	case op&0xFE00 == 0xF400: // lsr16 acR
		c.shiftAcc(int((op>>8)&1), false, -16)
	case op&0xF700 == 0x9100: // asr16 acR (bit 11)
		c.shiftAcc(int((op>>11)&1), true, -16)

	// --- moves into the accumulator ------------------------------------------------------
	case op&0xF800 == 0x6000: // movr acD, reg (reg = bits 10..9 -> 0x18+field) — the value lands
		// in the middle with the low cleared and the sign extended, regardless of mode
		d := int((op >> 8) & 1)
		reg := 0x18 + ((op >> 9) & 3)
		c.setAc(d, int64(int16(c.Reg[reg]))<<16)
		c.setArithFlags(d)
	case op&0xFC00 == 0x6800: // movax acD, axS — the full 32-bit ax, sign-extended
		d := int((op >> 8) & 1)
		c.setAc(d, c.ax(int((op>>9)&1)))
		c.setArithFlags(d)
	case op&0xFE00 == 0x6C00: // mov acD, ac(1-d) — copy the other 40-bit accumulator
		d := int((op >> 8) & 1)
		c.setAc(d, c.ac(1-d))
		c.setArithFlags(d)

	// --- the multiply / multiply-accumulate family ---------------------------------------
	case op&0xF700 == 0x9000: // mul axS.l, axS.h (s=bit11) — prod = product
		s := int((op >> 11) & 1)
		c.setProd(c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))
	case op&0xF600 == 0x9400: // mulac acD: acD += prod; prod = new product
		s := int((op >> 11) & 1)
		d := int((op >> 8) & 1)
		c.setAc(d, c.ac(d)+c.prod())
		c.setArithFlags(d)
		c.setProd(c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))
	case op&0xF600 == 0x9600: // mulmv acD: acD = prod; prod = new product
		s := int((op >> 11) & 1)
		d := int((op >> 8) & 1)
		c.setAc(d, c.prod())
		c.setArithFlags(d)
		c.setProd(c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))
	case op&0xF600 == 0x9200: // mulmvz acD: acD = prod with the low word cleared; prod = product
		s := int((op >> 11) & 1)
		d := int((op >> 8) & 1)
		c.setAc(d, c.prod()&^0xFFFF)
		c.setArithFlags(d)
		c.setProd(c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))
	case op&0xE700 == 0xC000: // mulc acS1.m, axS2.h (s1=bit12, s2=bit11) — prod = product
		s1 := int((op >> 12) & 1)
		s2 := int((op >> 11) & 1)
		c.setProd(c.mul16(c.Reg[regAC0M+s1], c.Reg[regAX0H+s2]))
	case op&0xE600 == 0xC200: // mulcmvz acR: acR = prod with the low word cleared; prod = new product
		s1 := int((op >> 12) & 1)
		s2 := int((op >> 11) & 1)
		r := int((op >> 8) & 1)
		c.setAc(r, c.prod()&^0xFFFF)
		c.setArithFlags(r)
		c.setProd(c.mul16(c.Reg[regAC0M+s1], c.Reg[regAX0H+s2]))
	case op&0xE600 == 0xC400: // mulcac acR: acR += prod; prod = acS.m * axT.h — the volume loop
		// at ucode 0x011E rides this with the LD2 dual load refreshing ax1 each pass.
		s1 := int((op >> 12) & 1)
		s2 := int((op >> 11) & 1)
		r := int((op >> 8) & 1)
		c.setAc(r, c.ac(r)+c.prod())
		c.setArithFlags(r)
		c.setProd(c.mul16(c.Reg[regAC0M+s1], c.Reg[regAX0H+s2]))
	case op&0xE600 == 0xC600: // mulcmv acR: acR = prod; prod = new product
		s1 := int((op >> 12) & 1)
		s2 := int((op >> 11) & 1)
		r := int((op >> 8) & 1)
		c.setAc(r, c.prod())
		c.setArithFlags(r)
		c.setProd(c.mul16(c.Reg[regAC0M+s1], c.Reg[regAX0H+s2]))

	// --- multiply-accumulate into prod without touching an accumulator --------------------
	case op&0xFC00 == 0xE000: // maddx — prod += (ax0 half S) * (ax1 half T), s=bit9, t=bit8
		c.setProd(c.prod() + c.mul16(c.Reg[regAX0L+((op>>9)&1)*2], c.Reg[regAX1L+((op>>8)&1)*2]))
	case op&0xFC00 == 0xE400: // msubx
		c.setProd(c.prod() - c.mul16(c.Reg[regAX0L+((op>>9)&1)*2], c.Reg[regAX1L+((op>>8)&1)*2]))
	case op&0xFC00 == 0xE800: // maddc — prod += acS.m * axT.h, s=bit9, t=bit8
		c.setProd(c.prod() + c.mul16(c.Reg[regAC0M+((op>>9)&1)], c.Reg[regAX0H+((op>>8)&1)]))
	case op&0xFC00 == 0xEC00: // msubc
		c.setProd(c.prod() - c.mul16(c.Reg[regAC0M+((op>>9)&1)], c.Reg[regAX0H+((op>>8)&1)]))

	// --- the cross-multiply family: a half of ax0 times a half of ax1. The two selector bits
	// pick low vs high as a register offset (0 -> ax_.l at 0x18/0x19, 2 -> ax_.h at 0x1A/0x1B);
	// which halves are used also decides the set15 unsigned treatment (see mulx16).
	case op&0xE700 == 0xA000: // mulx ax0.[lh], ax1.[lh] — prod = product
		c.setProd(c.mulxProd(op))
	case op&0xE600 == 0xA400: // mulxac acD: acD += prod; prod = new product
		d := int((op >> 8) & 1)
		p := c.mulxProd(op)
		c.setAc(d, c.ac(d)+c.prod())
		c.setArithFlags(d)
		c.setProd(p)
	case op&0xE600 == 0xA600: // mulxmv acD: acD = prod; prod = new product
		d := int((op >> 8) & 1)
		p := c.mulxProd(op)
		c.setAc(d, c.prod())
		c.setArithFlags(d)
		c.setProd(p)
	case op&0xE600 == 0xA200: // mulxmvz acD: acD = prod with the low word cleared; prod = product
		d := int((op >> 8) & 1)
		p := c.mulxProd(op)
		c.setAc(d, c.prod()&^0xFFFF)
		c.setArithFlags(d)
		c.setProd(p)
	case op&0xFE00 == 0xF200: // madd axS.l, axS.h (s=bit8) — prod += product
		s := int((op >> 8) & 1)
		c.setProd(c.prod() + c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))
	case op&0xFE00 == 0xF600: // msub axS.l, axS.h — prod -= product
		s := int((op >> 8) & 1)
		c.setProd(c.prod() - c.mul16(c.Reg[regAX0L+s], c.Reg[regAX0H+s]))

	// --- move / add the product into the accumulator -------------------------------------
	case op&0xFE00 == 0x4E00: // addp acD — acD += prod
		d := int((op >> 8) & 1)
		c.setAc(d, c.ac(d)+c.prod())
		c.setArithFlags(d)
	case op&0xFE00 == 0x6E00: // movp acD — the product into the accumulator
		d := int((op >> 8) & 1)
		c.setAc(d, c.prod())
		c.setArithFlags(d)
	case op&0xFE00 == 0x7E00: // movnp acD — the negated product
		d := int((op >> 8) & 1)
		c.setAc(d, -c.prod())
		c.setArithFlags(d)
	case op&0xFE00 == 0xFE00: // movpz acD — the product ROUNDED to its middle word (round half
		// to even at bit 16, not a truncation): the FIR epilogue (clrp; madd×8; movpz; store
		// acD.m) reads its Q15 result out this way.
		d := int((op >> 8) & 1)
		c.setAc(d, c.prodRounded())
		c.setArithFlags(d)
	case op&0xFC00 == 0xF800: // addpaxz acD, axS (d=bit8, s=bit9) — acD = rounded prod + the ax
		// high half <<16, the low word landing zero; carry compares the unrounded product
		// against the result
		d := int((op >> 8) & 1)
		s := int((op >> 9) & 1)
		oldProd := c.prod()
		v := c.prodRounded() + (c.ax(s) &^ 0xFFFF)
		c.setAc(d, v)
		res := c.ac(d)
		c.setFlag(srZero, res == 0)
		c.setFlag(srSign, res < 0)
		c.setFlag(srCarry, uint64(oldProd) > uint64(res))
		c.setFlag(srOverflow, false)

	default:
		c.Halt("unmodelled DSP arithmetic op 0x%04X at 0x%04X (%s)", op, pc, mustText(c.imem, pc))
		return 1
	}
	p.commit(c)
	return 1
}

// mustText disassembles for a halt message; it never fails.
func mustText(read func(uint16) uint16, pc uint16) string {
	t, _ := Disasm(read, pc)
	return t
}
