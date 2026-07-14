package gekko

// exec.go executes one instruction.
//
// The shape is the one every interpreter in this repository uses: fetch, decode the
// fields inline (not through the disassembler's Decode, which builds strings), dispatch,
// and advance. What is worth knowing before reading it:
//
//   - There is no delay slot. A taken branch sets PC and nothing else runs.
//   - The record bit ("add." rather than "add") means "and set CR0 from the result".
//   - The overflow-enable bit ("addo") means "and set XER[OV], and set XER[SO] for good".
//     SO is sticky: once set, only writing XER or executing mcrxr clears it.
//   - Every unimplemented encoding halts, naming the word and the address. A gap in this
//     core is a fact about this core, and it should be loud rather than a wrong answer.

import "math"

// Step executes one instruction and returns the cycles it took. The count is nominal —
// this is an interpreter, not a pipeline model — but it paces the timers, and the timers
// are what the game schedules against.
func (c *CPU) Step() int {
	if c.Halted {
		return 0
	}
	if c.checkInterrupt() {
		c.Steps++
		c.tick(1)
		return 1
	}

	w := c.fetch(c.PC)
	if c.Halted {
		return 0 // the fetch itself failed translation
	}
	pc := c.PC
	c.PC = pc + 4 // the default; a branch overwrites it

	c.execute(w, pc)

	c.Steps++
	c.tick(1)
	return 1
}

func (c *CPU) execute(w, pc uint32) {
	switch opcd(w) {
	case 3: // twi
		if c.trapCond(rs(w), c.GPR[ra(w)], uint32(simm(w))) {
			c.programException(SRR1Trap)
		}
	case 4:
		c.execPS(w, pc)
	case 7: // mulli
		c.GPR[rs(w)] = uint32(int32(c.GPR[ra(w)]) * simm(w))
	case 8: // subfic
		a := c.GPR[ra(w)]
		imm := uint32(simm(w))
		r := imm - a
		c.setCA(carrySub(imm, a))
		c.GPR[rs(w)] = r
	case 10: // cmpli
		c.compareLogical(crfD(w), c.GPR[ra(w)], uimm(w))
	case 11: // cmpi
		c.compareArith(crfD(w), c.GPR[ra(w)], uint32(simm(w)))
	case 12: // addic
		a := c.GPR[ra(w)]
		imm := uint32(simm(w))
		r := a + imm
		c.setCA(carryAdd(a, imm, 0))
		c.GPR[rs(w)] = r
	case 13: // addic.
		a := c.GPR[ra(w)]
		imm := uint32(simm(w))
		r := a + imm
		c.setCA(carryAdd(a, imm, 0))
		c.GPR[rs(w)] = r
		c.setCR0(r)
	case 14: // addi (and li, when rA is 0 — which reads zero, not r0)
		c.GPR[rs(w)] = c.raOrZero(w) + uint32(simm(w))
	case 15: // addis
		c.GPR[rs(w)] = c.raOrZero(w) + uimm(w)<<16
	case 16:
		c.execBC(w, pc)
	case 17: // sc
		// A GameCube has no operating system to call, so nothing normally handles this;
		// the machine may install a hook to trap the apploader's callbacks, and if it
		// does not, the architectural path is taken and the game's own handler runs.
		if c.SC != nil && c.SC(c) {
			return
		}
		c.Exception(VecSyscall, c.PC, 0)
	case 18:
		c.execB(w, pc)
	case 19:
		c.exec19(w, pc)
	case 20: // rlwimi
		n := shOf(w)
		m := mask32(mbOf(w), meOf(w))
		r := (rotl32(c.GPR[rs(w)], n) & m) | (c.GPR[ra(w)] &^ m)
		c.GPR[ra(w)] = r
		c.rc(w, r)
	case 21: // rlwinm
		r := rotl32(c.GPR[rs(w)], shOf(w)) & mask32(mbOf(w), meOf(w))
		c.GPR[ra(w)] = r
		c.rc(w, r)
	case 23: // rlwnm
		r := rotl32(c.GPR[rs(w)], c.GPR[rb(w)]&31) & mask32(mbOf(w), meOf(w))
		c.GPR[ra(w)] = r
		c.rc(w, r)
	case 24: // ori
		c.GPR[ra(w)] = c.GPR[rs(w)] | uimm(w)
	case 25: // oris
		c.GPR[ra(w)] = c.GPR[rs(w)] | uimm(w)<<16
	case 26: // xori
		c.GPR[ra(w)] = c.GPR[rs(w)] ^ uimm(w)
	case 27: // xoris
		c.GPR[ra(w)] = c.GPR[rs(w)] ^ uimm(w)<<16
	case 28: // andi.
		r := c.GPR[rs(w)] & uimm(w)
		c.GPR[ra(w)] = r
		c.setCR0(r)
	case 29: // andis.
		r := c.GPR[rs(w)] & (uimm(w) << 16)
		c.GPR[ra(w)] = r
		c.setCR0(r)
	case 31:
		c.exec31(w, pc)

	// Loads and stores, non-indexed. The update forms write the computed address back
	// into rA, which is how a compiler walks an array.
	case 32: // lwz
		c.GPR[rs(w)] = c.read32(c.ea(w))
	case 33: // lwzu
		ea := c.eaU(w)
		c.GPR[rs(w)] = c.read32(ea)
		c.GPR[ra(w)] = ea
	case 34: // lbz
		c.GPR[rs(w)] = uint32(c.read8(c.ea(w)))
	case 35: // lbzu
		ea := c.eaU(w)
		c.GPR[rs(w)] = uint32(c.read8(ea))
		c.GPR[ra(w)] = ea
	case 36: // stw
		c.write32(c.ea(w), c.GPR[rs(w)])
	case 37: // stwu
		ea := c.eaU(w)
		c.write32(ea, c.GPR[rs(w)])
		c.GPR[ra(w)] = ea
	case 38: // stb
		c.write8(c.ea(w), uint8(c.GPR[rs(w)]))
	case 39: // stbu
		ea := c.eaU(w)
		c.write8(ea, uint8(c.GPR[rs(w)]))
		c.GPR[ra(w)] = ea
	case 40: // lhz
		c.GPR[rs(w)] = uint32(c.read16(c.ea(w)))
	case 41: // lhzu
		ea := c.eaU(w)
		c.GPR[rs(w)] = uint32(c.read16(ea))
		c.GPR[ra(w)] = ea
	case 42: // lha — sign-extending
		c.GPR[rs(w)] = uint32(int32(int16(c.read16(c.ea(w)))))
	case 43: // lhau
		ea := c.eaU(w)
		c.GPR[rs(w)] = uint32(int32(int16(c.read16(ea))))
		c.GPR[ra(w)] = ea
	case 44: // sth
		c.write16(c.ea(w), uint16(c.GPR[rs(w)]))
	case 45: // sthu
		ea := c.eaU(w)
		c.write16(ea, uint16(c.GPR[rs(w)]))
		c.GPR[ra(w)] = ea
	case 46: // lmw — load every register from rS up to r31
		ea := c.ea(w)
		for i := rs(w); i < 32; i++ {
			c.GPR[i] = c.read32(ea)
			ea += 4
		}
	case 47: // stmw
		ea := c.ea(w)
		for i := rs(w); i < 32; i++ {
			c.write32(ea, c.GPR[i])
			ea += 4
		}

	// Floating-point loads and stores. lfs loads a single and lands it in BOTH halves of
	// the register, which is the paired-single unit showing through: a scalar load is a
	// broadcast.
	case 48: // lfs
		c.loadFS(rs(w), c.ea(w))
	case 49: // lfsu
		ea := c.eaU(w)
		c.loadFS(rs(w), ea)
		c.GPR[ra(w)] = ea
	case 50: // lfd
		c.FPR[rs(w)].PS0 = f64from(c.read64(c.ea(w)))
	case 51: // lfdu
		ea := c.eaU(w)
		c.FPR[rs(w)].PS0 = f64from(c.read64(ea))
		c.GPR[ra(w)] = ea
	case 52: // stfs
		c.write32(c.ea(w), math.Float32bits(float32(c.FPR[rs(w)].PS0)))
	case 53: // stfsu
		ea := c.eaU(w)
		c.write32(ea, math.Float32bits(float32(c.FPR[rs(w)].PS0)))
		c.GPR[ra(w)] = ea
	case 54: // stfd
		c.write64(c.ea(w), f64bits(c.FPR[rs(w)].PS0))
	case 55: // stfdu
		ea := c.eaU(w)
		c.write64(ea, f64bits(c.FPR[rs(w)].PS0))
		c.GPR[ra(w)] = ea

	case 56, 57, 60, 61:
		c.execPSQ(w)

	case 59:
		c.exec59(w)
	case 63:
		c.exec63(w)

	default:
		c.Halt("gekko: unimplemented primary opcode %d (word 0x%08X) at 0x%08X", opcd(w), w, pc)
	}
}

// --- Operand helpers ----------------------------------------------------------------

// raOrZero reads rA, except that register 0 means the literal zero in the addressing and
// add-immediate forms. This is not r0 being zero — r0 is an ordinary register — it is the
// *encoding* of rA=0 meaning "no base".
func (c *CPU) raOrZero(w uint32) uint32 {
	if ra(w) == 0 {
		return 0
	}
	return c.GPR[ra(w)]
}

// ea is the effective address of a non-indexed load or store.
func (c *CPU) ea(w uint32) uint32 { return c.raOrZero(w) + uint32(simm(w)) }

// eaU is the effective address of an *update* form, where rA=0 is not "no base" but an
// invalid encoding — there is nowhere to write the update back to.
func (c *CPU) eaU(w uint32) uint32 { return c.GPR[ra(w)] + uint32(simm(w)) }

// eax and eaxU are the indexed equivalents.
func (c *CPU) eax(w uint32) uint32  { return c.raOrZero(w) + c.GPR[rb(w)] }
func (c *CPU) eaxU(w uint32) uint32 { return c.GPR[ra(w)] + c.GPR[rb(w)] }

// rc applies the record bit.
func (c *CPU) rc(w, result uint32) {
	if rcbit(w) {
		c.setCR0(result)
	}
}

// loadFS loads a single-precision value into both halves of a register. The hardware does
// this because in paired-single mode a scalar operand has to appear in both slots; the
// effect is that lfs is a broadcast, and code that then runs a ps_ instruction on the
// register gets the value twice, deliberately.
func (c *CPU) loadFS(d, ea uint32) {
	v := float64(math.Float32frombits(c.read32(ea)))
	c.FPR[d].PS0 = v
	c.FPR[d].PS1 = v
}

func rotl32(v, n uint32) uint32 { return v<<n | v>>(32-n)&^(^uint32(0)<<n) }

// mask32 builds the PowerPC MASK(mb, me): the bits from mb to me inclusive, counting from
// the most significant end — and *wrapping* when mb > me, which is the case that a
// transcription gets wrong and which the compiler uses constantly for field extraction.
func mask32(mb, me uint32) uint32 {
	if mb <= me {
		// A run of ones from bit mb to bit me, MSB-numbered.
		return (^uint32(0) >> mb) & (^uint32(0) << (31 - me))
	}
	// Wrapped: ones at the top and the bottom, zeroes in the middle.
	return (^uint32(0) >> mb) | (^uint32(0) << (31 - me))
}

// carryAdd reports whether a + b + carryIn overflows 32 bits — the unsigned carry out,
// which is what XER[CA] holds. It is computed in 64 bits rather than by the usual
// wraparound trick, so that it says what it means.
func carryAdd(a, b, carryIn uint32) bool {
	return uint64(a)+uint64(b)+uint64(carryIn) > 0xFFFFFFFF
}

// carrySub reports the carry out of a - b, which PowerPC defines as the carry out of
// a + ^b + 1. So "carry set" means "no borrow", which is the opposite of the intuition
// most architectures give you.
func carrySub(a, b uint32) bool {
	return carryAdd(a, ^b, 1)
}

// overflowAdd reports a signed overflow: the operands agreed in sign and the result did
// not.
func overflowAdd(a, b, r uint32) bool {
	return (a^r)&(b^r)&0x80000000 != 0
}

// --- Compares -----------------------------------------------------------------------

func (c *CPU) compareArith(crf, a, b uint32) {
	f := uint32(0)
	switch {
	case int32(a) < int32(b):
		f = crLT
	case int32(a) > int32(b):
		f = crGT
	default:
		f = crEQ
	}
	if c.XER&XERSO != 0 {
		f |= crSO
	}
	c.SetCRField(crf, f)
}

func (c *CPU) compareLogical(crf, a, b uint32) {
	f := uint32(0)
	switch {
	case a < b:
		f = crLT
	case a > b:
		f = crGT
	default:
		f = crEQ
	}
	if c.XER&XERSO != 0 {
		f |= crSO
	}
	c.SetCRField(crf, f)
}

// trapCond evaluates a trap instruction's five-bit condition against two operands.
func (c *CPU) trapCond(to, a, b uint32) bool {
	sa, sb := int32(a), int32(b)
	switch {
	case to&0x10 != 0 && sa < sb:
		return true
	case to&0x08 != 0 && sa > sb:
		return true
	case to&0x04 != 0 && a == b:
		return true
	case to&0x02 != 0 && a < b:
		return true
	case to&0x01 != 0 && a > b:
		return true
	}
	return false
}

// --- Branches -------------------------------------------------------------------------

// branchTaken evaluates a conditional branch's BO/BI, decrementing CTR if the encoding
// says to. It is called exactly once per branch, because decrementing twice would be a
// bug that only shows up in a loop.
func (c *CPU) branchTaken(bo, bi uint32) bool {
	ctrOK := true
	if bo&boNoDec == 0 {
		c.CTR--
		if bo&boCTRZero != 0 {
			ctrOK = c.CTR == 0
		} else {
			ctrOK = c.CTR != 0
		}
	}
	condOK := true
	if bo&boNoCond == 0 {
		bit := (c.CR >> (31 - bi)) & 1
		want := uint32(0)
		if bo&boCondSet != 0 {
			want = 1
		}
		condOK = bit == want
	}
	return ctrOK && condOK
}

func (c *CPU) execBC(w, pc uint32) {
	if lk(w) {
		c.LR = pc + 4
	}
	if !c.branchTaken(rs(w), ra(w)) {
		return
	}
	d := int32(int16(w & 0xFFFC))
	if aa(w) {
		c.PC = uint32(d)
	} else {
		c.PC = uint32(int32(pc) + d)
	}
}

func (c *CPU) execB(w, pc uint32) {
	d := int32(w&0x03FFFFFC) << 6 >> 6
	if lk(w) {
		c.LR = pc + 4
	}
	if aa(w) {
		c.PC = uint32(d)
	} else {
		c.PC = uint32(int32(pc) + d)
	}
}

func (c *CPU) exec19(w, pc uint32) {
	switch xo10(w) {
	case 0: // mcrf
		c.SetCRField(crfD(w), c.CRField(crfS(w)))
	case 16: // bclr
		// The target is read BEFORE the link register is written, which matters for
		// blrl: it branches to the old LR and leaves the new one behind.
		target := c.LR &^ 3
		taken := c.branchTaken(rs(w), ra(w))
		if lk(w) {
			c.LR = pc + 4
		}
		if taken {
			c.PC = target
		}
	case 528: // bcctr — CTR is never decremented by this form, whatever BO says
		target := c.CTR &^ 3
		bo := rs(w)
		condOK := true
		if bo&boNoCond == 0 {
			bit := (c.CR >> (31 - ra(w))) & 1
			want := uint32(0)
			if bo&boCondSet != 0 {
				want = 1
			}
			condOK = bit == want
		}
		if lk(w) {
			c.LR = pc + 4
		}
		if condOK {
			c.PC = target
		}
	case 50: // rfi
		// Restore the machine state the exception saved, and resume where it said.
		c.MSR = (c.MSR &^ 0x87C0FF73) | (c.SRR1 & 0x87C0FF73)
		c.MSR &^= 1 << 18 // POW is never restored
		c.PC = c.SRR0 &^ 3
		c.LC.Enabled_ = c.HID2&HID2LCE != 0
	case 150: // isync — nothing to do in an interpreter with no pipeline
	case 33, 129, 193, 225, 257, 289, 417, 449:
		c.execCRLogic(w)
	default:
		c.Halt("gekko: unimplemented opcode 19 extended %d (word 0x%08X) at 0x%08X", xo10(w), w, pc)
	}
}

func (c *CPU) execCRLogic(w uint32) {
	d, a, b := rs(w), ra(w), rb(w)
	ba := (c.CR >> (31 - a)) & 1
	bb := (c.CR >> (31 - b)) & 1
	var v uint32
	switch xo10(w) {
	case 33: // crnor
		v = ^(ba | bb) & 1
	case 129: // crandc
		v = ba &^ bb & 1
	case 193: // crxor
		v = ba ^ bb
	case 225: // crnand
		v = ^(ba & bb) & 1
	case 257: // crand
		v = ba & bb
	case 289: // creqv
		v = ^(ba ^ bb) & 1
	case 417: // crorc
		v = (ba | (^bb & 1)) & 1
	case 449: // cror
		v = ba | bb
	}
	sh := 31 - d
	c.CR = (c.CR &^ (1 << sh)) | (v << sh)
}
