package gcdsp

// exec_ext.go is the parallel-extension engine. Many of the DSP's arithmetic, logic, multiply and
// move instructions carry a second operation in their low byte — a load, a store, a register move,
// or an address-register step — that runs "alongside" the main op. On the hardware the two are
// simultaneous; an interpreter has to pick an order that reproduces the hazard-free result.
//
// The model here is the standard one, and the microcode's own idioms confirm it: the copy loop at
// ucode 0x03AC preloads an accumulator before the loop, which is only meaningful if a parallel
// store writes the accumulator's value from BEFORE the main op changes it. So:
//
//   1. read phase  — capture the values the extension stores/moves FROM (the pre-op register
//                    state), perform the memory stores, and read the memory the extension loads;
//   2. the main op runs (updating accumulators / prod / flags);
//   3. write phase — write the loaded values into their destination registers and apply the
//                    address-register post-modifications.
//
// A caller runs `p := c.extBegin(op & 0xFF)`, then the main op, then `p.commit(c)`.
//
// The extension encodings are the documented set (gamecube-tools' opcodes_ext table). The address
// registers a two-address form uses are fixed by the hardware: the load side is AR0, the store
// side is AR3. The suffix bits choose the post-modification: N (bit 2) steps the load address by
// its index register IX0 instead of +1; M (bit 3) steps the store address by IX3 instead of +1.

// extPending holds a parallel load's register writes, deferred until after the main op so the main
// op sees the register's old value (the hazard-free ordering).
type extPending struct {
	n   int
	reg [2]uint16
	val [2]uint16
}

func (p *extPending) add(reg, val uint16) {
	if p.n < len(p.reg) {
		p.reg[p.n] = reg
		p.val[p.n] = val
		p.n++
	}
}

func (p extPending) commit(c *CPU) {
	for i := 0; i < p.n; i++ {
		c.setReg(p.reg[i], p.val[i])
	}
}

// extStep steps an address register: by its index register when byIndex, otherwise by +1. It is
// the post-modification the parallel two-address moves apply to AR0 (load) and AR3 (store).
func (c *CPU) extStep(n int, byIndex bool) {
	if byIndex {
		c.arStep(n, int(int16(c.Reg[regIX0+n])))
	} else {
		c.arStep(n, +1)
	}
}

// extBegin performs a parallel extension's read phase (stores with pre-op values, memory loads
// captured for deferred write-back) and returns the deferred register writes. ext == 0 is the
// no-op case (the instruction carries no extension). An unmodelled extension halts, so a silent
// wrong move never slips through.
func (c *CPU) extBegin(ext uint16) extPending {
	var p extPending
	switch {
	case ext == 0x00:
		// no extension

	case ext&0xC0 == 0x80:
		// LS / SL and their N/M variants: a simultaneous load through AR0 and store through AR3.
		// bit 1 selects the assembly form (LS vs SL) but the effect is the same pair of accesses;
		// bits 2 (N) and 3 (M) choose the two address steps. reg = bits 5..4 (0x18+field), the
		// stored accumulator middle = bit 0.
		reg := 0x18 + ((ext >> 4) & 3)
		accm := regAC0M + (ext & 1)
		c.dataWrite(c.Reg[regAR0+3], c.getReg(accm)) // store: pre-op accumulator middle -> [AR3]
		p.add(reg, c.dataRead(c.Reg[regAR0+0]))      // load:  [AR0] -> reg (deferred)
		c.extStep(0, ext&0x04 != 0)                  // AR0 step (N -> IX0)
		c.extStep(3, ext&0x08 != 0)                  // AR3 step (M -> IX3)

	case ext&0xC0 == 0x40:
		// L / LN: load a register through the address register PRG (bits 1..0), no store. reg =
		// bits 5..3. N (bit 2) steps by the index register.
		reg := 0x18 + ((ext >> 3) & 7)
		prg := int(ext & 3)
		p.add(reg, c.dataRead(c.Reg[regAR0+prg]))
		c.extStep(prg, ext&0x04 != 0)

	case ext&0xE0 == 0x20:
		// S / SN: store an accumulator low/middle word through address register PRG (bits 1..0).
		// The stored register is bits 4..3 (0x1C+field: ac0.l, ac1.l, ac0.m, ac1.m).
		prg := int(ext & 3)
		src := 0x1C + ((ext >> 3) & 3)
		c.dataWrite(c.Reg[regAR0+prg], c.getReg(src))
		c.extStep(prg, ext&0x04 != 0)

	case ext&0xF0 == 0x10:
		// MV: move an accumulator low/middle word (bits 1..0 -> 0x1C+field) into a register
		// (bits 3..2 -> 0x18+field). A register-to-register move, no memory.
		dst := 0x18 + ((ext >> 2) & 3)
		src := 0x1C + (ext & 3)
		p.add(dst, c.getReg(src))

	case ext&0xC0 == 0xC0:
		// The dual-load family: two simultaneous reads, one through an address register arS and
		// one through AR3, landing in the ax half-registers (deferred past the main op like every
		// parallel load — the FIR's `clrp : ld` primes the first operand pair on the clrp itself).
		// Bit 2 (N) steps arS by its index register instead of +1, bit 3 (M) does the same for AR3
		// with IX3. Two forms share the space (the manual's bit-decoding summary §7: LD = 11mn barr,
		// LD2 = 11rm ba11; gamecube-tools carries only the first):
		//   - LD (low bits != 3): one half of AX0 (bit 5: 0=l, 1=h) from [arS], one half of AX1
		//     (bit 4) from [AR3]; arS is bits 1..0 — the value 3 is the LD2 marker, not AR3.
		//   - LD2 (low bits == 3): BOTH halves of one ax register — high from [arS], low from
		//     [AR3]; bit 4 picks the ax register, bit 5 the arS (AR0 or AR1).
		// LD2's operand assignment is pinned by the ucode two independent ways: the 8-tap FIR at
		// 0x01D2 (`madd ax0` = ax0.l*ax0.h needs both halves fresh each tap; its circular ar0
		// coefficient buffer is only read if arS feeds a half) and the volume loop at 0x011E,
		// whose preamble primes ax1.h from [ar0] twice before `mulcac : ld2 0xD3` continues the
		// same stream — so bit 4 = the ax (ax1 there), high half = the arS side.
		if ext&0x03 == 0x03 { // LD2 axR, @arS: axR.h <- [arS], axR.l <- [AR3]
			s := int((ext >> 5) & 1)
			r := (ext >> 4) & 1
			p.add(regAX0H+r, c.dataRead(c.Reg[regAR0+s]))
			p.add(regAX0L+r, c.dataRead(c.Reg[regAR0+3]))
			c.extStep(s, ext&0x04 != 0)
			c.extStep(3, ext&0x08 != 0)
		} else { // LD: ax0 half (bit 5) <- [arS], ax1 half (bit 4) <- [AR3]
			s := int(ext & 3)
			p.add(regAX0L+((ext>>5)&1)*2, c.dataRead(c.Reg[regAR0+s]))
			p.add(regAX1L+((ext>>4)&1)*2, c.dataRead(c.Reg[regAR0+3]))
			c.extStep(s, ext&0x04 != 0)
			c.extStep(3, ext&0x08 != 0)
		}

	case ext&0xFC == 0x04: // DR: decrement address register (bits 1..0)
		c.arStep(int(ext&3), -1)
	case ext&0xFC == 0x08: // IR: increment address register
		c.arStep(int(ext&3), +1)
	case ext&0xFC == 0x0C: // NR: step address register by its index register
		n := int(ext & 3)
		c.arStep(n, int(int16(c.Reg[regIX0+n])))

	default:
		c.Halt("unmodelled parallel extension 0x%02X at 0x%04X", ext, c.PC)
	}
	return p
}

// --- the multiplier ----------------------------------------------------------------------
//
// The DSP multiplies two 16-bit operands into the 40-bit PROD register. clr15/set15 choose signed
// or unsigned operands (the mixing ucode multiplies signed samples), and m0/m2 choose whether the
// product is used as-is or doubled (the fractional-format left shift). These are documented mode
// bits; the exact fixed-point conventions are the part most worth checking against a reference DSP.

// mul16 forms the 16x16 product under the current sign (set15) and shift (m2) modes.
func (c *CPU) mul16(a, b uint16) int64 {
	var p int64
	if c.sr()&srMulSigned != 0 { // set15: unsigned operands
		p = int64(uint32(a)) * int64(uint32(b))
	} else { // clr15 (default): signed operands
		p = int64(int16(a)) * int64(int16(b))
	}
	if c.sr()&srMulShift != 0 { // m2: product doubled
		p <<= 1
	}
	return p
}

// setProd writes a 40-bit value into the product register's four pieces, clearing the second
// middle word so a later prod() read returns exactly this value.
func (c *CPU) setProd(v int64) {
	c.Reg[regPRODL] = uint16(v)
	c.Reg[regPRODM1] = uint16(v >> 16)
	c.Reg[regPRODH] = uint16(v >> 32)
	c.Reg[regPRODM2] = 0
}
