package gcdsp

// exec_support.go holds the machinery the interpreter leans on: the three memory spaces, the
// register file's read/write side effects, the four hardware stacks, the block-loop service,
// the status-flag computations, and the branch-condition test. Kept apart from exec.go so that
// file reads as the instruction set and this one as the plumbing beneath it.

// --- memory ------------------------------------------------------------------------------

// imem fetches an instruction word. Instructions live in IRAM at 0x0000 and in the absent boot
// IROM at 0x8000; a fetch from anywhere else, or from the IROM this core does not carry, halts.
func (c *CPU) imem(a uint16) uint16 {
	switch {
	case a < 0x1000:
		return c.IRAM[a]
	case a >= 0x8000:
		if c.IROM != nil && int(a-0x8000) < len(c.IROM) {
			return c.IROM[a-0x8000]
		}
		c.Halt("instruction fetch from boot IROM @0x%04X — IROM not present", a)
	default:
		c.Halt("instruction fetch @0x%04X — unmapped", a)
	}
	return 0
}

// dataRead reads a data word. DRAM is at 0x0000, the coefficient DROM at 0x1000, and the
// hardware registers at 0xFF00 and up; a read of the absent DROM or of unmapped space halts.
func (c *CPU) dataRead(a uint16) uint16 {
	switch {
	case a < 0x1000:
		return c.DRAM[a]
	case a < 0x2000:
		if c.DROM != nil && int(a-0x1000) < len(c.DROM) {
			return c.DROM[a-0x1000]
		}
		c.Halt("DSP read of coefficient ROM @0x%04X — DROM not present (a resampling table the ucode wants)", a)
	case a >= 0xFF00:
		if c.bus != nil {
			return c.bus.HWRead(a)
		}
		c.Halt("DSP hardware read @0x%04X — no bus attached", a)
	default:
		c.Halt("DSP data read @0x%04X — unmapped", a)
	}
	return 0
}

// dataWrite writes a data word to DRAM or, at the top of the space, to a hardware register.
func (c *CPU) dataWrite(a, v uint16) {
	switch {
	case a < 0x1000:
		c.DRAM[a] = v
	case a >= 0xFF00:
		if c.bus != nil {
			c.bus.HWWrite(a, v)
			return
		}
		c.Halt("DSP hardware write @0x%04X — no bus attached", a)
	default:
		c.Halt("DSP data write @0x%04X = 0x%04X — unmapped", a, v)
	}
}

// --- register file with side effects -----------------------------------------------------

// getReg reads a register-file entry. Reading one of the four ST registers pops its hardware
// stack, and reading an accumulator MIDDLE word in 40-bit mode saturates: a value that does not
// fit in 32 bits reads as 0x7FFF/0x8000 — the clamp the mixer's stores of MAC results ride on.
// In 16-bit mode the middle reads plainly.
func (c *CPU) getReg(r uint16) uint16 {
	switch {
	case r >= regST0 && r <= regST3:
		return c.pop(r)
	case (r == regAC0M || r == regAC1M) && c.sr()&srMode40 != 0:
		n := int(r - regAC0M)
		v := c.ac(n)
		if v != int64(int32(v)) {
			if v > 0 {
				return 0x7FFF
			}
			return 0x8000
		}
		return c.Reg[r]
	default:
		return c.Reg[r]
	}
}

// setReg writes a register-file entry, with the hardware's canonicalisations: an ST write pushes
// its stack, an accumulator HIGH write sign-extends from its low byte (only 8 bits are real),
// the product high keeps only its low byte, the config register its low byte, and an SR write
// drops the always-zero bit 8. An accumulator MIDDLE write is plain here — the mode-dependent
// whole-accumulator extension belongs to the plain load instructions (setRegExtend), NOT to the
// parallel-extension loads, which land as bare register writes.
func (c *CPU) setReg(r, v uint16) {
	switch {
	case r >= regST0 && r <= regST3:
		c.push(r, v)
	case r == regAC0H || r == regAC1H:
		c.Reg[r] = uint16(int16(int8(v)))
	case r == regPRODH:
		c.Reg[r] = v & 0x00FF
	case r == regCONFIG:
		c.Reg[r] = v & 0x00FF
	case r == regSR:
		c.Reg[r] = v &^ 0x0100
	default:
		c.Reg[r] = v
	}
}

// setRegExtend is setReg plus the load-instruction rule for accumulator middles: in 40-bit mode
// (set40) a mid load sign-extends the high byte AND CLEARS THE LOW WORD, making the accumulator
// the loaded value <<16; in 16-bit mode (set16, the power-on state) the mid loads plainly and
// the rest of the accumulator keeps its old bits. lri/lris/lr/lrs/lrr/mrr use this; the parallel
// extensions do not.
func (c *CPU) setRegExtend(r, v uint16) {
	c.setReg(r, v)
	if (r == regAC0M || r == regAC1M) && c.sr()&srMode40 != 0 {
		n := int(r - regAC0M)
		if v&0x8000 != 0 {
			c.Reg[regAC0H+n] = 0xFFFF
		} else {
			c.Reg[regAC0H+n] = 0x0000
		}
		c.Reg[regAC0L+n] = 0
	}
}

// --- address registers -------------------------------------------------------------------

// The address registers step under the wrapping registers, which bound each step to a circular
// buffer of wr+1 words (wr = 0xFFFF, the ucode's usual setting, is a plain 16-bit add). The
// wrap is a carry detector, not a true modulo: a step wraps by the buffer length when it flips
// an address bit above the buffer's span, which is only equivalent to a circular buffer when
// the ring sits against a suitably aligned boundary — and this ucode's rings do exactly that
// (the FIR's 8-word table at 0x03E8 ends at the 16-boundary 0x3F0; the pitch resampler's
// 160-word ring, wr3=0x9F, ends at the 1024-boundary 0x0C00). The manual predates the wrapping
// registers entirely (it leaves $8..$11 unnamed); the four formulas below are the
// hardware-verified ones, adopted from Dolphin's DSP-LLE interpreter with the user's approval
// (see doc.go). The single-step forms and the add-index forms detect the carry differently, so
// they are kept as four distinct operations, matching which instruction uses which.

// arInc post-increments address register n (the +1 of lrri/srri and the plain extension steps).
func (c *CPU) arInc(n int) {
	ar, wr := uint32(c.Reg[regAR0+n]), uint32(c.Reg[regWR0+n])
	nar := ar + 1
	if (nar ^ ar) > (wr|1)<<1 {
		nar -= wr + 1
	}
	c.Reg[regAR0+n] = uint16(nar)
}

// arDec post-decrements address register n (dar and the decrementing loads/stores).
func (c *CPU) arDec(n int) {
	ar, wr := uint32(c.Reg[regAR0+n]), uint32(c.Reg[regWR0+n])
	nar := ar + wr
	if (nar^ar)&((wr|1)<<1) > wr {
		nar -= wr + 1
	}
	c.Reg[regAR0+n] = uint16(nar)
}

// arAdd steps address register n by the signed index amount ix (addarn and the N-variant
// extension steps).
func (c *CPU) arAdd(n int, ix int16) {
	ar, wr := uint32(c.Reg[regAR0+n]), uint32(c.Reg[regWR0+n])
	mx := (wr | 1) << 1
	nar := ar + uint32(int32(ix))
	dar := (nar ^ ar ^ uint32(int32(ix))) & mx
	if ix >= 0 {
		if dar > wr { // carried past the ring
			nar -= wr + 1
		}
	} else {
		if ((nar+wr+1)^nar)&dar <= wr { // borrowed past it
			nar += wr + 1
		}
	}
	c.Reg[regAR0+n] = uint16(nar)
}

// arSub steps address register n by minus the signed index amount (subarn).
func (c *CPU) arSub(n int, ix int16) {
	ar, wr := uint32(c.Reg[regAR0+n]), uint32(c.Reg[regWR0+n])
	mx := (wr | 1) << 1
	nar := ar - uint32(int32(ix))
	dar := (nar ^ ar ^ ^uint32(int32(ix))) & mx
	if ix < 0 && ix != -0x8000 {
		if dar > wr {
			nar -= wr + 1
		}
	} else {
		if ((nar+wr+1)^nar)&dar <= wr {
			nar += wr + 1
		}
	}
	c.Reg[regAR0+n] = uint16(nar)
}

// --- hardware stacks ---------------------------------------------------------------------

func (c *CPU) push(reg, v uint16) {
	i := reg - regST0
	c.Stacks[i] = append(c.Stacks[i], v)
}

func (c *CPU) pop(reg uint16) uint16 {
	i := reg - regST0
	if len(c.Stacks[i]) == 0 {
		c.Halt("DSP stack ST%d underflow at 0x%04X", i, c.PC)
		return 0
	}
	v := c.Stacks[i][len(c.Stacks[i])-1]
	c.Stacks[i] = c.Stacks[i][:len(c.Stacks[i])-1]
	return v
}

// --- hardware loops ----------------------------------------------------------------------

// startLoop begins a repeat of the instructions from start to end (inclusive), count times.
func (c *CPU) startLoop(start, end, count uint16) {
	if count == 0 {
		c.Halt("DSP loop with count 0 at 0x%04X — not yet modelled", c.PC)
		return
	}
	c.Loops = append(c.Loops, LoopFrame{Start: start, End: end, Count: count})
}

// serviceLoops runs after each instruction. If the instruction just executed sits at the end
// of the innermost active loop and did not itself branch away, the loop either jumps back to
// its start (another iteration remains) or finishes and is popped.
func (c *CPU) serviceLoops(execAddr uint16, branched bool) {
	if len(c.Loops) == 0 || branched {
		return
	}
	top := &c.Loops[len(c.Loops)-1]
	if execAddr != top.End {
		return
	}
	top.Count--
	if top.Count > 0 {
		c.PC = top.Start
	} else {
		c.Loops = c.Loops[:len(c.Loops)-1]
	}
}

// --- accumulator shift -------------------------------------------------------------------

// shiftAcc shifts the 40-bit accumulator acR by a signed amount: positive shifts left,
// negative shifts right by the magnitude. The logical form treats the accumulator as an
// unsigned 40-bit quantity (zero-fill on the right shift); the arithmetic form preserves the
// sign (bit 39 replicated on the right shift). A left shift is the same bit operation either
// way. The result is written back sign-extended and the zero/sign flags are set from it.
func (c *CPU) shiftAcc(r int, arith bool, amt int) {
	switch {
	case amt >= 0:
		v := (uint64(c.ac(r)) << uint(amt)) & 0xFFFFFFFFFF
		c.setAc(r, int64(v))
	case arith:
		// Arithmetic right: c.ac already returns the value sign-extended into an int64, so a
		// Go arithmetic shift carries the sign; setAc masks it back to 40 bits.
		c.setAc(r, c.ac(r)>>uint(-amt))
	default:
		// Logical right: shift the unsigned 40-bit value, zero-filling from the top.
		v := (uint64(c.ac(r)) & 0xFFFFFFFFFF) >> uint(-amt)
		c.setAc(r, int64(v))
	}
	c.setArithFlags(r)
}

// shiftAmount7 decodes the register-shift ops' 7-bit signed amount: low six bits of zero mean
// no shift regardless of the sign bit; otherwise bit 6 makes the low six bits negative.
func shiftAmount7(v uint16) int {
	switch {
	case v&0x3F == 0:
		return 0
	case v&0x40 != 0:
		return -0x40 + int(v&0x3F)
	default:
		return int(v & 0x3F)
	}
}

// --- status flags ------------------------------------------------------------------------

// setArithFlags sets the compare flags from a full 40-bit accumulator result the way the moves,
// shifts and multiply-accumulates do: zero and sign from the value, carry and overflow CLEARED
// (every flag-writing op rewrites the whole compare group; only the adds and subtracts put
// something in the carry/overflow bits).
func (c *CPU) setArithFlags(n int) {
	v := c.ac(n)
	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
	c.setFlag(srCarry, false)
	c.setFlag(srOverflow, false)
}

// aluAddSub adds (or, when sub, subtracts) a 40-bit operand into accumulator d, stores the
// result, and sets the zero, sign, carry and overflow flags the branch conditions read. The
// carry compares the first operand against the sign-extended result (add carries when the
// result dropped below the operand, a subtract's carry means "no borrow"); overflow is the
// two's-complement sign rule over the sign-extended values, latched into the sticky bit. b is
// the operand already sign-extended into an int64.
func (c *CPU) aluAddSub(d int, b int64, sub bool) {
	a := c.ac(d)
	if sub {
		b = -b
	}
	c.setAc(d, a+b)
	v := c.ac(d) // the stored 40-bit sign-extended result

	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
	if sub {
		c.setFlag(srCarry, uint64(a) >= uint64(v))
	} else {
		c.setFlag(srCarry, uint64(a) > uint64(v))
	}
	c.setFlag(srOverflow, (a^v)&(b^v) < 0)
	if c.Reg[regSR]&srOverflow != 0 {
		c.Reg[regSR] |= srOverSticky
	}
}

// setTestFlags sets the flags a tst produces: it compares the accumulator with zero, so the
// zero and sign flags come straight from the value and the carry and overflow are cleared. With
// overflow clear, the GE/LE branch conditions reduce to the accumulator's sign, which is what
// the mixer's "is this sample negative" tests rely on.
func (c *CPU) setTestFlags(n int) {
	v := c.ac(n)
	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
	c.setFlag(srCarry, false)
	c.setFlag(srOverflow, false)
}

// setLogicFlags sets the compare flags from an accumulator MIDDLE word after a logic op: zero
// and sign are judged on the 16-bit result alone, carry and overflow are cleared, and the
// LOGIC-ZERO flag is NOT touched — only andf/andcf write that one (the mailbox-wait idiom
// depends on it surviving the surrounding logic ops).
func (c *CPU) setLogicFlags(n int) {
	v := int16(c.Reg[regAC0M+n])
	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
	c.setFlag(srCarry, false)
	c.setFlag(srOverflow, false)
}

// subFlags sets the flags a compare produces: the zero, sign, carry and overflow of a-b over
// the 40-bit accumulator width, without storing the difference. The carry compares the first
// operand against the sign-extended result ("no borrow"); overflow is the two's-complement
// sign rule.
func (c *CPU) subFlags(a, b int64) {
	res := ((a - b) << 24) >> 24 // sign-extend the 40-bit difference
	c.setFlag(srZero, res == 0)
	c.setFlag(srSign, res < 0)
	c.setFlag(srCarry, uint64(a) >= uint64(res))
	c.setFlag(srOverflow, (a^res)&(-b^res) < 0)
	if c.Reg[regSR]&srOverflow != 0 {
		c.Reg[regSR] |= srOverSticky
	}
}

// --- branch conditions -------------------------------------------------------------------

// cond evaluates a 4-bit branch condition against the status flags. The unconditional code and
// the arithmetic/logic conditions the microcode uses are implemented; an unrecognised code
// halts rather than guess a direction and corrupt control flow.
func (c *CPU) cond(cc uint16) bool {
	sr := c.Reg[regSR]
	z := sr&srZero != 0
	s := sr&srSign != 0
	o := sr&srOverflow != 0
	cf := sr&srCarry != 0
	lz := sr&srLogicZero != 0
	switch cc {
	case 0x0: // GE: sign == overflow
		return s == o
	case 0x1: // L: sign != overflow
		return s != o
	case 0x2: // G: greater and not zero
		return s == o && !z
	case 0x3: // LE: less or zero
		return s != o || z
	case 0x4: // NZ
		return !z
	case 0x5: // Z
		return z
	case 0x6: // NC: carry clear
		return !cf
	case 0x7: // C: carry set
		return cf
	case 0xC: // LNZ: logic result not zero
		return !lz
	case 0xD: // LZ: logic result zero
		return lz
	case 0xF: // always
		return true
	}
	c.Halt("unmodelled branch condition 0x%X at 0x%04X", cc, c.PC)
	return false
}
