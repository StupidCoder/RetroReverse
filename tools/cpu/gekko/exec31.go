package gekko

// exec31.go executes primary opcode 31 — the integer arithmetic, the indexed loads and
// stores, the cache instructions and the moves to and from the special registers. It is
// the largest family, and it is where the rules that a manual states in one line and an
// implementation gets wrong in one character live:
//
//   - XER[CA] on a subtract is the carry out of a + ¬b + 1, so "carry set" means "no
//     borrow" — the opposite of the flag most architectures set.
//   - srawi and sraw set CA if and only if the operand is negative AND a one-bit was
//     shifted out. A negative operand alone is not enough; a shift of zero never sets it.
//   - divw by zero, and 0x80000000 / -1, leave the destination register *architecturally
//     undefined* — but they still set OV and SO. This core writes zero, and says so; the
//     vector suite marks the register don't-care and asserts the flags, which is the only
//     honest way to test an instruction whose result the architecture declines to define.
//   - The overflow-enable bit is the top bit of the extended-opcode field, so it is
//     stripped in the decoder for exactly the instructions that have it (see oeForm).

import "math/bits"

func (c *CPU) exec31(w, pc uint32) {
	d, a, b := rs(w), ra(w), rb(w)

	x := xo10(w)
	if x >= 512 && oeForm[x-512] {
		x -= 512
	}

	switch x {
	// --- Compares and traps ---
	case 0: // cmp
		c.compareArith(crfD(w), c.GPR[a], c.GPR[b])
	case 32: // cmpl
		c.compareLogical(crfD(w), c.GPR[a], c.GPR[b])
	case 4: // tw
		if c.trapCond(d, c.GPR[a], c.GPR[b]) {
			c.programException(SRR1Trap)
		}

	// --- Addition and subtraction ---
	case 266: // add
		va, vb := c.GPR[a], c.GPR[b]
		r := va + vb
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(va, vb, r))
	case 10: // addc — add and record the carry out
		va, vb := c.GPR[a], c.GPR[b]
		r := va + vb
		c.setCA(carryAdd(va, vb, 0))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(va, vb, r))
	case 138: // adde — add with the carry in
		va, vb, ci := c.GPR[a], c.GPR[b], c.ca()
		r := va + vb + ci
		c.setCA(carryAdd(va, vb, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(va, vb, r))
	case 234: // addme — add -1 and the carry in
		va, ci := c.GPR[a], c.ca()
		r := va + 0xFFFFFFFF + ci
		c.setCA(carryAdd(va, 0xFFFFFFFF, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(va, 0xFFFFFFFF, r))
	case 202: // addze — add the carry in
		va, ci := c.GPR[a], c.ca()
		r := va + ci
		c.setCA(carryAdd(va, 0, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(va, 0, r))
	case 40: // subf — rB minus rA, which is why the operand order reads backwards
		va, vb := c.GPR[a], c.GPR[b]
		r := vb - va
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(^va, vb, r))
	case 8: // subfc
		va, vb := c.GPR[a], c.GPR[b]
		r := vb - va
		c.setCA(carrySub(vb, va))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(^va, vb, r))
	case 136: // subfe — rB + ¬rA + CA
		va, vb, ci := c.GPR[a], c.GPR[b], c.ca()
		r := vb + ^va + ci
		c.setCA(carryAdd(vb, ^va, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(^va, vb, r))
	case 232: // subfme — ¬rA + CA - 1
		va, ci := c.GPR[a], c.ca()
		r := ^va + 0xFFFFFFFF + ci
		c.setCA(carryAdd(^va, 0xFFFFFFFF, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(^va, 0xFFFFFFFF, r))
	case 200: // subfze — ¬rA + CA
		va, ci := c.GPR[a], c.ca()
		r := ^va + ci
		c.setCA(carryAdd(^va, 0, ci))
		c.GPR[d] = r
		c.oeRc(w, r, overflowAdd(^va, 0, r))
	case 104: // neg — the one overflow case is negating the most negative integer
		va := c.GPR[a]
		r := -va
		c.GPR[d] = r
		c.oeRc(w, r, va == 0x80000000)

	// --- Multiplication and division ---
	case 235: // mullw
		va, vb := int32(c.GPR[a]), int32(c.GPR[b])
		full := int64(va) * int64(vb)
		r := uint32(full)
		c.GPR[d] = r
		// Overflow is "the 64-bit product did not fit in 32 bits", not the sign trick.
		c.oeRc(w, r, full != int64(int32(full)))
	case 75: // mulhw
		full := int64(int32(c.GPR[a])) * int64(int32(c.GPR[b]))
		r := uint32(full >> 32)
		c.GPR[d] = r
		c.rc(w, r)
	case 11: // mulhwu
		full := uint64(c.GPR[a]) * uint64(c.GPR[b])
		r := uint32(full >> 32)
		c.GPR[d] = r
		c.rc(w, r)
	case 491: // divw
		va, vb := int32(c.GPR[a]), int32(c.GPR[b])
		// Two cases the architecture declines to define a result for. It DOES define the
		// overflow flags, so those are set, and the register gets zero — recorded as a
		// choice, not asserted as a fact.
		if vb == 0 || (va == -0x80000000 && vb == -1) {
			c.GPR[d] = 0
			c.oeRc(w, 0, true)
			return
		}
		r := uint32(va / vb)
		c.GPR[d] = r
		c.oeRc(w, r, false)
	case 459: // divwu
		va, vb := c.GPR[a], c.GPR[b]
		if vb == 0 {
			c.GPR[d] = 0
			c.oeRc(w, 0, true)
			return
		}
		r := va / vb
		c.GPR[d] = r
		c.oeRc(w, r, false)

	// --- Logical ---
	case 28: // and
		r := c.GPR[d] & c.GPR[b]
		c.GPR[a] = r
		c.rc(w, r)
	case 60: // andc
		r := c.GPR[d] &^ c.GPR[b]
		c.GPR[a] = r
		c.rc(w, r)
	case 444: // or (and mr, when rS == rB)
		r := c.GPR[d] | c.GPR[b]
		c.GPR[a] = r
		c.rc(w, r)
	case 412: // orc
		r := c.GPR[d] | ^c.GPR[b]
		c.GPR[a] = r
		c.rc(w, r)
	case 316: // xor
		r := c.GPR[d] ^ c.GPR[b]
		c.GPR[a] = r
		c.rc(w, r)
	case 124: // nor (and not, when rS == rB)
		r := ^(c.GPR[d] | c.GPR[b])
		c.GPR[a] = r
		c.rc(w, r)
	case 476: // nand
		r := ^(c.GPR[d] & c.GPR[b])
		c.GPR[a] = r
		c.rc(w, r)
	case 284: // eqv
		r := ^(c.GPR[d] ^ c.GPR[b])
		c.GPR[a] = r
		c.rc(w, r)
	case 26: // cntlzw
		r := uint32(bits.LeadingZeros32(c.GPR[d]))
		c.GPR[a] = r
		c.rc(w, r)
	case 922: // extsh
		r := uint32(int32(int16(c.GPR[d])))
		c.GPR[a] = r
		c.rc(w, r)
	case 954: // extsb
		r := uint32(int32(int8(c.GPR[d])))
		c.GPR[a] = r
		c.rc(w, r)

	// --- Shifts. The shift amount is SIX bits, not five: a shift of 32 or more produces
	// zero (or, for the arithmetic shift, the sign), rather than wrapping round to a
	// shift of zero, which is what a naive & 31 would do.
	case 24: // slw
		sh := c.GPR[b] & 63
		r := uint32(0)
		if sh < 32 {
			r = c.GPR[d] << sh
		}
		c.GPR[a] = r
		c.rc(w, r)
	case 536: // srw
		sh := c.GPR[b] & 63
		r := uint32(0)
		if sh < 32 {
			r = c.GPR[d] >> sh
		}
		c.GPR[a] = r
		c.rc(w, r)
	case 792: // sraw
		c.srawi(w, c.GPR[b]&63)
	case 824: // srawi
		c.srawi(w, shOf(w))

	// --- The condition register and the special registers ---
	case 19: // mfcr
		c.GPR[d] = c.CR
	case 144: // mtcrf — one bit of the mask per CR field
		mask := (w >> 12) & 0xFF
		m := uint32(0)
		for i := 0; i < 8; i++ {
			if mask&(0x80>>i) != 0 {
				m |= 0xF << (28 - 4*i)
			}
		}
		c.CR = (c.CR &^ m) | (c.GPR[d] & m)
	case 512: // mcrxr — move XER's top nibble to a CR field, AND CLEAR IT. This is the
		// only instruction that clears the sticky summary-overflow bit.
		c.SetCRField(crfD(w), c.XER>>28)
		c.XER &^= XERSO | XEROV | XERCA
	case 339: // mfspr
		c.GPR[d] = c.readSPR(sprOf(w), pc)
	case 467: // mtspr
		c.writeSPR(sprOf(w), c.GPR[d], pc)
	case 371: // mftb — the time base, read through its OWN SPR numbers (268/269), which are
		// not the ones mtspr writes it through (284/285). Conflating the two is an easy
		// mistake, and it halts a game the moment it reads the clock.
		switch sprOf(w) {
		case 268: // TBL
			c.GPR[d] = uint32(c.TB)
		case 269: // TBU
			c.GPR[d] = uint32(c.TB >> 32)
		default:
			c.Halt("gekko: mftb of SPR %d at 0x%08X", sprOf(w), pc)
		}
	case 83: // mfmsr
		c.GPR[d] = c.MSR
	case 146: // mtmsr
		c.MSR = c.GPR[d]
	case 210: // mtsr
		c.SR[a&15] = c.GPR[d]
	case 242: // mtsrin
		c.SR[(c.GPR[b]>>28)&15] = c.GPR[d]
	case 595: // mfsr
		c.GPR[d] = c.SR[a&15]
	case 659: // mfsrin
		c.GPR[d] = c.SR[(c.GPR[b]>>28)&15]

	// --- Indexed loads and stores ---
	case 23: // lwzx
		c.GPR[d] = c.read32(c.eax(w))
	case 55: // lwzux
		ea := c.eaxU(w)
		c.GPR[d] = c.read32(ea)
		c.GPR[a] = ea
	case 87: // lbzx
		c.GPR[d] = uint32(c.read8(c.eax(w)))
	case 119: // lbzux
		ea := c.eaxU(w)
		c.GPR[d] = uint32(c.read8(ea))
		c.GPR[a] = ea
	case 279: // lhzx
		c.GPR[d] = uint32(c.read16(c.eax(w)))
	case 311: // lhzux
		ea := c.eaxU(w)
		c.GPR[d] = uint32(c.read16(ea))
		c.GPR[a] = ea
	case 343: // lhax
		c.GPR[d] = uint32(int32(int16(c.read16(c.eax(w)))))
	case 375: // lhaux
		ea := c.eaxU(w)
		c.GPR[d] = uint32(int32(int16(c.read16(ea))))
		c.GPR[a] = ea
	case 151: // stwx
		c.write32(c.eax(w), c.GPR[d])
	case 183: // stwux
		ea := c.eaxU(w)
		c.write32(ea, c.GPR[d])
		c.GPR[a] = ea
	case 215: // stbx
		c.write8(c.eax(w), uint8(c.GPR[d]))
	case 247: // stbux
		ea := c.eaxU(w)
		c.write8(ea, uint8(c.GPR[d]))
		c.GPR[a] = ea
	case 407: // sthx
		c.write16(c.eax(w), uint16(c.GPR[d]))
	case 439: // sthux
		ea := c.eaxU(w)
		c.write16(ea, uint16(c.GPR[d]))
		c.GPR[a] = ea

	// --- The byte-reversed forms. This is a big-endian machine, so these are its
	// little-endian accessors — and they are exactly the instructions a careless
	// transcription implements backwards, because "reversed" is relative to something.
	case 534: // lwbrx
		v := c.read32(c.eax(w))
		c.GPR[d] = bits.ReverseBytes32(v)
	case 790: // lhbrx
		v := c.read16(c.eax(w))
		c.GPR[d] = uint32(bits.ReverseBytes16(v))
	case 662: // stwbrx
		c.write32(c.eax(w), bits.ReverseBytes32(c.GPR[d]))
	case 918: // sthbrx
		c.write16(c.eax(w), bits.ReverseBytes16(uint16(c.GPR[d])))

	// --- Load and store multiple/string ---
	case 597: // lswi
		n := rb(w)
		if n == 0 {
			n = 32
		}
		c.loadString(d, c.raOrZero(w), n)
	case 725: // stswi
		n := rb(w)
		if n == 0 {
			n = 32
		}
		c.storeString(d, c.raOrZero(w), n)
	case 533: // lswx
		c.loadString(d, c.eax(w), c.XER&0x7F)
	case 661: // stswx
		c.storeString(d, c.eax(w), c.XER&0x7F)

	// --- The atomic pair. lwarx takes a reservation on the line; stwcx. stores only if
	// the reservation is still held, and reports in CR0[eq] whether it was. Any write to
	// the line, from anywhere, clears it.
	case 20: // lwarx
		ea := c.eax(w)
		c.GPR[d] = c.read32(ea)
		c.Reserved = true
		c.ReserveAddr = ea
	case 150: // stwcx.
		ea := c.eax(w)
		f := uint32(0)
		if c.XER&XERSO != 0 {
			f |= crSO
		}
		if c.Reserved && c.ReserveAddr&^31 == ea&^31 {
			c.write32(ea, c.GPR[d])
			f |= crEQ // the store happened
		}
		c.Reserved = false
		c.SetCRField(0, f)

	// --- Indexed floating-point ---
	case 535: // lfsx
		c.loadFS(d, c.eax(w))
	case 567: // lfsux
		ea := c.eaxU(w)
		c.loadFS(d, ea)
		c.GPR[a] = ea
	case 599: // lfdx
		c.FPR[d].PS0 = f64from(c.read64(c.eax(w)))
	case 631: // lfdux
		ea := c.eaxU(w)
		c.FPR[d].PS0 = f64from(c.read64(ea))
		c.GPR[a] = ea
	case 663: // stfsx
		c.write32(c.eax(w), float32bitsOf(c.FPR[d].PS0))
	case 695: // stfsux
		ea := c.eaxU(w)
		c.write32(ea, float32bitsOf(c.FPR[d].PS0))
		c.GPR[a] = ea
	case 727: // stfdx
		c.write64(c.eax(w), bits64(c.FPR[d].PS0))
	case 759: // stfdux
		ea := c.eaxU(w)
		c.write64(ea, bits64(c.FPR[d].PS0))
		c.GPR[a] = ea
	case 983: // stfiwx — store the low word of the register's bit pattern, unconverted
		c.write32(c.eax(w), uint32(bits64(c.FPR[d].PS0)))

	// --- Cache and synchronisation ---
	case 1014: // dcbz — writes 32 bytes of zeroes. Not a hint.
		c.dcbz(c.eax(w))
	case 54, 86, 246, 278, 470, 982: // dcbst, dcbf, dcbtst, dcbt, dcbi, icbi
		// This core has no write-back cache to flush and no instruction cache to
		// invalidate, so these have nothing to do. That is a property of the model, not
		// of the hardware, and it is safe here only because nothing in the machine reads
		// memory that the CPU has written and not yet flushed — the graphics FIFO is fed
		// through the uncached alias, deliberately, by the program itself.
	case 598, 854, 566: // sync, eieio, tlbsync
	case 306, 370: // tlbie, tlbia — no page table to invalidate an entry of
	case 310, 438: // eciwx, ecowx — the external control facility; unused on a GameCube
		c.Halt("gekko: external control (eciwx/ecowx) at 0x%08X; this machine has no such device", pc)

	default:
		c.Halt("gekko: unimplemented opcode 31 extended %d (word 0x%08X) at 0x%08X", xo10(w), w, pc)
	}
}

// srawi is the arithmetic right shift, and its carry rule is the one every PowerPC
// implementation gets wrong once: CA is set if and only if the value is negative AND at
// least one one-bit was shifted out of it. A negative value shifted by zero does not set
// it; a negative value whose shifted-out bits are all zero does not set it.
func (c *CPU) srawi(w, sh uint32) {
	v := int32(c.GPR[rs(w)])
	var r int32
	if sh >= 32 {
		// The whole value is shifted out; the result is all sign bits, and the carry is
		// set exactly when there was a sign bit to shift out.
		r = v >> 31
		c.setCA(v < 0)
	} else {
		r = v >> sh
		// The bits that fell off the bottom.
		lost := uint32(v) & ((1 << sh) - 1)
		c.setCA(v < 0 && lost != 0)
	}
	c.GPR[ra(w)] = uint32(r)
	c.rc(w, uint32(r))
}

// oeRc applies the overflow-enable and record bits together, in that order — OV must be
// set before CR0 is computed, because CR0's summary bit is copied out of XER.
func (c *CPU) oeRc(w, result uint32, overflow bool) {
	if oe(w) {
		c.setOV(overflow)
	}
	c.rc(w, result)
}

// loadString and storeString move n bytes into or out of consecutive registers, wrapping
// from r31 to r0, packing four bytes to a register, big end first.
func (c *CPU) loadString(d, ea, n uint32) {
	reg := d
	c.GPR[reg] = 0
	for i := uint32(0); i < n; i++ {
		shift := 24 - 8*(i%4)
		c.GPR[reg] |= uint32(c.read8(ea+i)) << shift
		if i%4 == 3 {
			reg = (reg + 1) & 31
			c.GPR[reg] = 0
		}
	}
}

func (c *CPU) storeString(d, ea, n uint32) {
	reg := d
	for i := uint32(0); i < n; i++ {
		shift := 24 - 8*(i%4)
		c.write8(ea+i, uint8(c.GPR[reg]>>shift))
		if i%4 == 3 {
			reg = (reg + 1) & 31
		}
	}
}
