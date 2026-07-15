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
// stack; every other register reads plainly.
func (c *CPU) getReg(r uint16) uint16 {
	if r >= regST0 && r <= regST3 {
		return c.pop(r)
	}
	return c.Reg[r]
}

// setReg writes a register-file entry. Writing an ST register pushes its stack; writing an
// accumulator middle word sign-extends the accumulator's high byte to match, which is the
// hardware's behaviour and is what keeps a value loaded into acX.m readable as a signed 40-bit
// accumulator.
func (c *CPU) setReg(r, v uint16) {
	switch {
	case r >= regST0 && r <= regST3:
		c.push(r, v)
	case r == regAC0M || r == regAC1M:
		n := int(r - regAC0M)
		c.Reg[regAC0M+n] = v
		if v&0x8000 != 0 {
			c.Reg[regAC0H+n] = 0xFFFF
		} else {
			c.Reg[regAC0H+n] = 0x0000
		}
	default:
		c.Reg[r] = v
	}
}

// --- address registers -------------------------------------------------------------------

// arStep advances address register n by a signed delta. The wrapping register wr bounds the
// step modulo a circular buffer; the common configuration this ucode sets is wr = 0xFFFF, no
// wrap, a plain add. A non-trivial wrap is not yet modelled and halts, naming the register,
// rather than stepping wrong.
func (c *CPU) arStep(n int, delta int) {
	wr := c.Reg[regWR0+n]
	if wr == 0xFFFF {
		c.Reg[regAR0+n] = uint16(int32(c.Reg[regAR0+n]) + int32(delta))
		return
	}
	c.Halt("DSP address-register wrap (ar%d, wr=0x%04X) not yet modelled at 0x%04X", n, wr, c.PC)
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

// --- status flags ------------------------------------------------------------------------

// setArithFlags sets the zero and sign flags from a full 40-bit accumulator result. The finer
// carry/overflow bits are added as the ops that produce them are implemented.
func (c *CPU) setArithFlags(n int) {
	v := c.ac(n)
	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
}

// aluAddSub adds (or, when sub, subtracts) a 40-bit operand into accumulator d, stores the
// result, and sets the zero, sign, carry and overflow flags the branch conditions read. The
// carry means "no borrow" for a subtract; overflow is the ordinary two's-complement signed
// overflow, taken over the 40-bit accumulator width, and is also latched into the sticky
// overflow bit. b is the operand already sign-extended into an int64.
func (c *CPU) aluAddSub(d int, b int64, sub bool) {
	a := c.ac(d)
	var res int64
	if sub {
		res = a - b
	} else {
		res = a + b
	}
	c.setAc(d, res)
	v := c.ac(d) // the stored 40-bit signed result

	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)

	ua := uint64(a) & 0xFFFFFFFFFF
	ub := uint64(b) & 0xFFFFFFFFFF
	if sub {
		c.setFlag(srCarry, ua >= ub) // no borrow out of bit 39
		// overflow: operands differ in sign and the result took the subtrahend's sign
		c.setFlag(srOverflow, (a^b)&(a^v)&(1<<39) != 0)
	} else {
		c.setFlag(srCarry, ua+ub > 0xFFFFFFFFFF) // carry out of bit 39
		// overflow: operands share a sign and the result's sign flipped
		c.setFlag(srOverflow, ^(a^b)&(a^v)&(1<<39) != 0)
	}
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

// setLogicFlags sets the zero and sign flags from an accumulator after a logic op. The DSP
// judges these on the whole accumulator, so read it back the same way.
func (c *CPU) setLogicFlags(n int) {
	v := c.ac(n)
	c.setFlag(srZero, v == 0)
	c.setFlag(srSign, v < 0)
	c.setFlag(srLogicZero, uint16(v>>16) == 0)
}

// subFlags sets the flags a compare produces: the zero, sign, carry and overflow of a-b, taken
// over the 40-bit accumulator width. It does not store the difference.
func (c *CPU) subFlags(a, b int64) {
	d := a - b
	c.setFlag(srZero, (d&0xFFFFFFFFFF) == 0)
	c.setFlag(srSign, d < 0)
	c.setFlag(srCarry, uint64(a&0xFFFFFFFFFF) >= uint64(b&0xFFFFFFFFFF))
	// Signed overflow: operands differ in sign and the result takes the subtrahend's sign.
	c.setFlag(srOverflow, ((a^b)&(a^d))&(1<<39) != 0)
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
