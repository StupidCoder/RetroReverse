package r5900

// cpu_test.go covers what diff_test.go structurally cannot: the instructions the
// R5900 has and the VR4300 does not. That is the MMI SIMD space, the 128-bit
// load/store, the three-operand multiply, the shift-amount register, the
// conditional moves, and a floating-point unit that is not IEEE 754.

import (
	"testing"
)

// testBus is a flat little-endian RAM with the program at 0. It is 2 MiB, which is
// enough to hold a physical address above 0x00100000 — the address the Jak boot ELF
// is linked at, and the one TestUnmappedKusegFaults needs to reach for real rather
// than have wrapped back onto the program.
const testRAMSize = 2 << 20
const testRAMMask = testRAMSize - 1

type testBus struct{ ram []byte }

func newTestBus() *testBus { return &testBus{ram: make([]byte, testRAMSize)} }

func (b *testBus) Read(addr uint32) byte     { return b.ram[addr&testRAMMask] }
func (b *testBus) Write(addr uint32, v byte) { b.ram[addr&testRAMMask] = v }

func (b *testBus) Read32(addr uint32) uint32 {
	a := addr & testRAMMask
	return uint32(b.ram[a]) | uint32(b.ram[a+1])<<8 | uint32(b.ram[a+2])<<16 | uint32(b.ram[a+3])<<24
}

func (b *testBus) Write32(addr uint32, v uint32) {
	a := addr & testRAMMask
	b.ram[a] = byte(v)
	b.ram[a+1] = byte(v >> 8)
	b.ram[a+2] = byte(v >> 16)
	b.ram[a+3] = byte(v >> 24)
}

// Encoders, so a test reads as the instruction it means.
func special(rs, rt, rd, sa, funct uint32) uint32 {
	return rs<<21 | rt<<16 | rd<<11 | sa<<6 | funct
}
func mmiW(rs, rt, rd, sub, funct uint32) uint32 {
	return 0x1C<<26 | rs<<21 | rt<<16 | rd<<11 | sub<<6 | funct
}
func itype(op, rs, rt, imm uint32) uint32 { return op<<26 | rs<<21 | rt<<16 | imm&0xFFFF }
func cop1(fmtf, ft, fs, fd, funct uint32) uint32 {
	return 0x11<<26 | fmtf<<21 | ft<<16 | fs<<11 | fd<<6 | funct
}

// --- the 128-bit register file ----------------------------------------------

func TestQuadPreservedByOrdinaryOps(t *testing.T) {
	// The defining property of this core: a 64-bit MIPS operation writes the low
	// half and must leave the high half — the MMI world — untouched.
	bus := newTestBus()
	bus.Write32(0, special(1, 2, 3, 0, 0x2D)) // daddu $v1, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 5, Hi: 0xAAAAAAAAAAAAAAAA}
	c.R[2] = Quad{Lo: 7, Hi: 0xBBBBBBBBBBBBBBBB}
	c.R[3] = Quad{Lo: 0, Hi: 0xCCCCCCCCCCCCCCCC}
	c.Step()

	if c.R[3].Lo != 12 {
		t.Errorf("daddu low half = %d, want 12", c.R[3].Lo)
	}
	if c.R[3].Hi != 0xCCCCCCCCCCCCCCCC {
		t.Errorf("daddu disturbed the upper 64 bits: got 0x%016X, want 0xCCCCCCCCCCCCCCCC", c.R[3].Hi)
	}
}

// kseg0Base is a data base register value that translates without a TLB entry.
// Data addresses in these tests go through it: the low addresses are in KUSEG,
// which is *mapped*, so a bare 0x100 would correctly take a TLB miss.
const kseg0Base = 0x80000000

func TestLoadStoreQuadword(t *testing.T) {
	bus := newTestBus()
	// sq $t0, 0x100($t2) ; lq $t1, 0x100($t2)   with $t2 = 0x80000000
	bus.Write32(0, itype(0x1F, 10, 8, 0x100))
	bus.Write32(4, itype(0x1E, 10, 9, 0x100))
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[10] = Quad{Lo: kseg0Base}
	want := Quad{Lo: 0x0123456789ABCDEF, Hi: 0xFEDCBA9876543210}
	c.R[8] = want
	c.Step()
	c.Step()

	if c.R[9] != want {
		t.Errorf("sq/lq round-trip: got %v, want %v", c.R[9], want)
	}
	// Little-endian: the low byte of the low doubleword is at the lowest address.
	if bus.ram[0x100] != 0xEF || bus.ram[0x10F] != 0xFE {
		t.Errorf("sq byte order wrong: [0x100]=0x%02X [0x10F]=0x%02X, want 0xEF and 0xFE",
			bus.ram[0x100], bus.ram[0x10F])
	}
}

func TestLoadQuadForcesAlignment(t *testing.T) {
	// lq and sq mask the address down to a quadword rather than faulting on a
	// misaligned one — unlike every other load on the machine.
	bus := newTestBus()
	bus.Write32(0, itype(0x1E, 10, 9, 0x107)) // lq $t1, 0x107($t2)
	bus.Write32(0x100, 0xDEADBEEF)
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[10] = Quad{Lo: kseg0Base}
	c.Step()

	if c.Halted {
		t.Fatalf("halted: %s", c.HaltReason)
	}
	if uint32(c.R[9].Lo) != 0xDEADBEEF {
		t.Errorf("lq from 0x107 read 0x%08X, want the quadword at 0x100 (0xDEADBEEF)", uint32(c.R[9].Lo))
	}
}

func TestUnmappedKusegFaults(t *testing.T) {
	// The counterpart of the two tests above: an address in KUSEG with nothing
	// mapping it takes a TLB miss rather than reading memory. This is not a detail —
	// the Jak boot ELF is linked at 0x00100000, inside KUSEG, so the machine model
	// must install the mapping the BIOS kernel would have before the game runs, and
	// this is the fault it would otherwise take on its very first instruction.
	//
	// 0x00100000 is used rather than a low address on purpose. An all-zero TLB entry
	// tags VPN2 0 with ASID 0, so an address near zero genuinely *matches* it and
	// takes the "entry present but invalid" vector at 0x180. Only an address outside
	// every entry's tag misses entirely and reaches the refill vector at 0x000.
	const jakLoadAddr = 0x00100000

	bus := newTestBus()
	bus.Write32(0, itype(0x23, 10, 9, 0)) // lw $t1, 0($t2)
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = 0
	c.R[10] = Quad{Lo: jakLoadAddr}
	c.Step()

	if c.PC != 0xFFFFFFFF80000000 {
		t.Errorf("an unmapped KUSEG load did not vector to the TLB refill handler: PC = 0x%016X", c.PC)
	}
	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excTLBL {
		t.Errorf("exception code = %d, want excTLBL (%d)", got, excTLBL)
	}
	if c.COP0[cop0BadVAddr] != jakLoadAddr {
		t.Errorf("BadVAddr = 0x%X, want 0x%X", c.COP0[cop0BadVAddr], jakLoadAddr)
	}

	// And with the mapping installed, the same load succeeds — which is exactly what
	// the machine model will do at boot. The mapping is identity, so the virtual
	// address 0x00100000 reads physical 0x00100000.
	bus2 := newTestBus()
	bus2.Write32(0, itype(0x23, 10, 9, 0))
	bus2.Write32(jakLoadAddr, 0xCAFEBABE)
	d := NewCPU(bus2)
	d.SetPC(0x80000000)
	d.COP0[cop0Status] = 0
	d.R[10] = Quad{Lo: jakLoadAddr}
	// One 16 MiB page pair, identity-mapped, covering 0x00000000..0x01FFFFFF — all of
	// main memory, which is what the kernel lays down for a game.
	d.SetTLB(0, TLBEntry{
		PageMask: 0x01FFE000, // 16 MiB pages
		EntryHi:  0,
		EntryLo0: (0 << 6) | entryLoV | entryLoD | entryLoG,      // PFN 0        -> 0x00000000
		EntryLo1: (0x1000 << 6) | entryLoV | entryLoD | entryLoG, // PFN 0x1000   -> 0x01000000
	})
	d.Step()

	if d.Halted {
		t.Fatalf("halted: %s", d.HaltReason)
	}
	if uint32(d.R[9].Lo) != 0xCAFEBABE {
		t.Errorf("with the mapping installed, lw from 0x%08X read 0x%08X, want 0xCAFEBABE",
			jakLoadAddr, uint32(d.R[9].Lo))
	}
}

// --- the R5900's own integer instructions -----------------------------------

func TestMultWritesDestination(t *testing.T) {
	// "mult rd, rs, rt" writes LO into rd as well as the accumulator. The VR4300's
	// mult has no rd, which is why diff_test.go forces it to zero.
	bus := newTestBus()
	bus.Write32(0, special(1, 2, 3, 0, 0x18)) // mult $v1, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 7}
	c.R[2] = Quad{Lo: 6}
	c.Step()

	if c.LO != 42 {
		t.Errorf("LO = %d, want 42", c.LO)
	}
	if c.R[3].Lo != 42 {
		t.Errorf("mult did not write rd: $v1 = %d, want 42", c.R[3].Lo)
	}
}

func TestSecondAccumulator(t *testing.T) {
	// mult1 targets HI1/LO1, leaving HI/LO alone — the point of having two.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0, 0x18)) // mult1 $v1, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 100}
	c.R[2] = Quad{Lo: 3}
	c.HI, c.LO = 0xDEAD, 0xBEEF
	c.Step()

	if c.LO1 != 300 {
		t.Errorf("LO1 = %d, want 300", c.LO1)
	}
	if c.LO != 0xBEEF || c.HI != 0xDEAD {
		t.Errorf("mult1 clobbered the first accumulator: HI=0x%X LO=0x%X", c.HI, c.LO)
	}
}

func TestConditionalMoves(t *testing.T) {
	// movz $v1, $at, $v0 — move when rt is zero.
	bus := newTestBus()
	bus.Write32(0, special(1, 2, 3, 0, 0x0A))
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 99}
	c.R[2] = Quad{Lo: 0}
	c.R[3] = Quad{Lo: 1}
	c.Step()
	if c.R[3].Lo != 99 {
		t.Errorf("movz with rt==0 did not move: got %d, want 99", c.R[3].Lo)
	}

	bus = newTestBus()
	bus.Write32(0, special(1, 2, 3, 0, 0x0A))
	c = NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 99}
	c.R[2] = Quad{Lo: 5} // non-zero: must not move
	c.R[3] = Quad{Lo: 1}
	c.Step()
	if c.R[3].Lo != 1 {
		t.Errorf("movz with rt!=0 moved anyway: got %d, want 1", c.R[3].Lo)
	}
}

func TestShiftAmountRegisterAndFunnelShift(t *testing.T) {
	// mtsa sets SA; qfsrv shifts the 256-bit pair (rs:rt) right by SA bits and keeps
	// the low 128. With SA = 64, the result is rt's high half under rs's low half.
	bus := newTestBus()
	bus.Write32(0, special(1, 0, 0, 0, 0x29)) // mtsa $at
	bus.Write32(4, mmiW(2, 3, 4, 0x1B, 0x28)) // qfsrv $a0, $v0, $v1
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = Quad{Lo: 64}                                         // SA = 64 bits
	c.R[2] = Quad{Lo: 0x1111111111111111, Hi: 0x2222222222222222} // rs — the high source
	c.R[3] = Quad{Lo: 0x3333333333333333, Hi: 0x4444444444444444} // rt — the low source
	c.Step()
	c.Step()

	want := Quad{Lo: 0x4444444444444444, Hi: 0x1111111111111111}
	if c.R[4] != want {
		t.Errorf("qfsrv SA=64: got %v, want %v", c.R[4], want)
	}
}

// --- MMI ---------------------------------------------------------------------

func TestPackedAddHalfword(t *testing.T) {
	// paddh adds eight 16-bit lanes independently — the lanes must not carry into
	// each other, which is the whole point of a packed add.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0x04, 0x08)) // paddh $v1, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromHalves([8]uint16{0xFFFF, 1, 2, 3, 4, 5, 6, 7})
	c.R[2] = fromHalves([8]uint16{1, 1, 1, 1, 1, 1, 1, 1})
	c.Step()

	got := halves(c.R[3])
	want := [8]uint16{0, 2, 3, 4, 5, 6, 7, 8} // lane 0 wraps to 0; no carry into lane 1
	if got != want {
		t.Errorf("paddh = %v, want %v", got, want)
	}
}

func TestPackedSaturatingAdd(t *testing.T) {
	// paddsh saturates each signed 16-bit lane instead of wrapping.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0x14, 0x08)) // paddsh
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromHalves([8]uint16{0x7FFF, 0x8000, 100, 0, 0, 0, 0, 0})
	c.R[2] = fromHalves([8]uint16{1, 0xFFFF, 100, 0, 0, 0, 0, 0})
	c.Step()

	got := halves(c.R[3])
	if got[0] != 0x7FFF {
		t.Errorf("paddsh positive overflow = 0x%04X, want 0x7FFF", got[0])
	}
	if got[1] != 0x8000 {
		t.Errorf("paddsh negative overflow = 0x%04X, want 0x8000", got[1])
	}
	if got[2] != 200 {
		t.Errorf("paddsh in-range lane = %d, want 200", got[2])
	}
}

func TestPackedCompareProducesMask(t *testing.T) {
	// pcgtw sets a lane to all ones when rs > rt, all zeros otherwise — a mask, not
	// a boolean, so it can be used to select.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0x02, 0x08)) // pcgtw
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromWords([4]uint32{5, 0xFFFFFFFF, 0, 100}) // 5, -1, 0, 100
	c.R[2] = fromWords([4]uint32{3, 1, 0, 200})          // 3,  1, 0, 200
	c.Step()

	got := words(c.R[3])
	want := [4]uint32{0xFFFFFFFF, 0, 0, 0} // 5>3 yes; -1>1 no; 0>0 no; 100>200 no
	if got != want {
		t.Errorf("pcgtw = %08X, want %08X", got, want)
	}
}

func TestInterleaveAndCopy(t *testing.T) {
	// pextlw interleaves the low words of rt and rs; pcpyud takes the high
	// doublewords. These two carry most of the register shuffling in vector code.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0x12, 0x08)) // pextlw $v1, $at, $v0
	bus.Write32(4, mmiW(1, 2, 4, 0x0E, 0x29)) // pcpyud $a0, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromWords([4]uint32{0xA0, 0xA1, 0xA2, 0xA3}) // rs
	c.R[2] = fromWords([4]uint32{0xB0, 0xB1, 0xB2, 0xB3}) // rt
	c.Step()
	c.Step()

	if got, want := words(c.R[3]), [4]uint32{0xB0, 0xA0, 0xB1, 0xA1}; got != want {
		t.Errorf("pextlw = %X, want %X", got, want)
	}
	// pcpyud: rd.Lo = rs.Hi, rd.Hi = rt.Hi.
	if got, want := words(c.R[4]), [4]uint32{0xA2, 0xA3, 0xB2, 0xB3}; got != want {
		t.Errorf("pcpyud = %X, want %X", got, want)
	}
}

func TestPackedMultiplyHalfwordFillsAccumulator(t *testing.T) {
	// pmulth computes eight 16x16 products into the 128-bit HI/LO pair and returns
	// four of them in rd.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 2, 3, 0x1C, 0x09)) // pmulth $v1, $at, $v0
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromHalves([8]uint16{1, 2, 3, 4, 5, 6, 7, 8})
	c.R[2] = fromHalves([8]uint16{10, 10, 10, 10, 10, 10, 10, 10})
	c.Step()

	// Products: 10 20 30 40 50 60 70 80, laid out LO, HI, LO1, HI1.
	if c.LO != 10|20<<32 {
		t.Errorf("LO = 0x%016X, want products 0 and 1", c.LO)
	}
	if c.HI != 30|40<<32 {
		t.Errorf("HI = 0x%016X, want products 2 and 3", c.HI)
	}
	if c.LO1 != 50|60<<32 {
		t.Errorf("LO1 = 0x%016X, want products 4 and 5", c.LO1)
	}
	// rd gets the LO halves: products 0, 1, 4, 5.
	if got, want := words(c.R[3]), [4]uint32{10, 20, 50, 60}; got != want {
		t.Errorf("pmulth rd = %v, want %v", got, want)
	}
}

func TestPlzcw(t *testing.T) {
	// plzcw counts the leading bits that repeat the sign, less one — the shift a
	// normaliser applies. It runs on the two low words independently.
	bus := newTestBus()
	bus.Write32(0, mmiW(1, 0, 3, 0, 0x04)) // plzcw $v1, $at
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.R[1] = fromWords([4]uint32{0x00000001, 0xFFFFFFFF, 0, 0})
	c.Step()

	got := words(c.R[3])
	if got[0] != 30 { // 0x00000001: 31 leading zeros, less one
		t.Errorf("plzcw(1) = %d, want 30", got[0])
	}
	if got[1] != 31 { // all ones: every bit repeats the sign
		t.Errorf("plzcw(-1) = %d, want 31", got[1])
	}
}

// --- COP1: the ways it is not IEEE 754 ---------------------------------------

func TestFPUHasNoInfinity(t *testing.T) {
	// Multiplying two large values overflows. IEEE gives +Inf; this unit saturates.
	bus := newTestBus()
	bus.Write32(0, cop1(0x10, 2, 1, 3, 0x02)) // mul.s $f3, $f1, $f2
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = statusCU1
	c.FPR[1] = 0x7F000000 // ~1.7e38
	c.FPR[2] = 0x7F000000
	c.Step()

	if c.FPR[3] != fMax {
		t.Errorf("mul.s overflow = 0x%08X, want the saturated maximum 0x%08X (not an infinity)",
			c.FPR[3], uint32(fMax))
	}
}

func TestFPUDivideByZeroSaturates(t *testing.T) {
	bus := newTestBus()
	bus.Write32(0, cop1(0x10, 2, 1, 3, 0x03)) // div.s $f3, $f1, $f2
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = statusCU1
	c.FPR[1] = 0x3F800000 // 1.0
	c.FPR[2] = 0          // 0.0
	c.Step()

	if c.FPR[3] != fMax {
		t.Errorf("1.0/0.0 = 0x%08X, want the saturated maximum 0x%08X", c.FPR[3], uint32(fMax))
	}
}

func TestFPUDenormalsReadAsZero(t *testing.T) {
	// A denormal source operand is zero to this unit. 1.0 + smallest-denormal is
	// exactly 1.0, which is also true on IEEE — so the test that bites is a denormal
	// compared against zero.
	bus := newTestBus()
	bus.Write32(0, cop1(0x10, 2, 1, 0, 0x32)) // c.eq.s $f1, $f2
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = statusCU1
	c.FPR[1] = 0x00000001 // the smallest IEEE denormal
	c.FPR[2] = 0x00000000 // zero
	c.Step()

	if c.FCR31&fcr31C == 0 {
		t.Error("c.eq.s: a denormal did not compare equal to zero, but this unit has no denormals")
	}
}

func TestFPURoundsTowardZero(t *testing.T) {
	// 1.0 + 2^-24 is exactly halfway between 1.0 and the next representable single.
	// IEEE round-to-nearest-even gives 1.0; chopping gives 1.0 too. The case that
	// separates them is a value strictly between, biased up: 1.0 + 1.5 * 2^-24
	// rounds *up* under IEEE and *down* under truncation.
	bus := newTestBus()
	bus.Write32(0, cop1(0x10, 2, 1, 3, 0x00)) // add.s $f3, $f1, $f2
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = statusCU1
	c.FPR[1] = 0x3F800000 // 1.0
	c.FPR[2] = 0x33C00000 // 1.5 * 2^-24
	c.Step()

	if c.FPR[3] != 0x3F800000 {
		t.Errorf("add.s = 0x%08X, want 0x3F800000: the unit truncates, it does not round to nearest",
			c.FPR[3])
	}
}

func TestFPUAccumulator(t *testing.T) {
	// madd.s reads the accumulator and writes a register, leaving ACC alone.
	bus := newTestBus()
	bus.Write32(0, cop1(0x10, 2, 1, 0, 0x1A)) // mula.s $f1, $f2   -> ACC = f1*f2
	bus.Write32(4, cop1(0x10, 2, 1, 3, 0x1C)) // madd.s $f3, $f1, $f2 -> f3 = ACC + f1*f2
	c := NewCPU(bus)
	c.SetPC(0x80000000)
	c.COP0[cop0Status] = statusCU1
	c.FPR[1] = 0x40400000 // 3.0
	c.FPR[2] = 0x40800000 // 4.0
	c.Step()
	c.Step()

	if got := fdec(c.ACC); got != 12 {
		t.Errorf("mula.s: ACC = %v, want 12", got)
	}
	if got := fdec(c.FPR[3]); got != 24 { // 12 + 3*4
		t.Errorf("madd.s: $f3 = %v, want 24", got)
	}
	if got := fdec(c.ACC); got != 12 {
		t.Errorf("madd.s must not write the accumulator: ACC = %v, want 12", got)
	}
}

// --- savestates ---------------------------------------------------------------

func TestSnapshotRoundTripsTheUpperHalves(t *testing.T) {
	// A savestate that dropped the 128-bit halves, the second accumulator or SA
	// would restore a core that diverges the moment it reaches MMI code.
	bus := newTestBus()
	c := NewCPU(bus)
	c.R[5] = Quad{Lo: 0x1122334455667788, Hi: 0x99AABBCCDDEEFF00}
	c.HI1, c.LO1 = 0xAAAA, 0xBBBB
	c.SA = 40
	c.ACC = 0x40490FDB

	s := c.Snapshot()
	d := NewCPU(bus)
	d.Restore(s)

	if d.R[5] != c.R[5] {
		t.Errorf("R5 did not round-trip: got %v, want %v", d.R[5], c.R[5])
	}
	if d.HI1 != c.HI1 || d.LO1 != c.LO1 {
		t.Errorf("the second accumulator did not round-trip")
	}
	if d.SA != c.SA {
		t.Errorf("SA did not round-trip: got %d, want %d", d.SA, c.SA)
	}
	if d.ACC != c.ACC {
		t.Errorf("ACC did not round-trip")
	}
}

// --- the decoder --------------------------------------------------------------

func TestDecodeRendersMMI(t *testing.T) {
	cases := []struct {
		word uint32
		want string
	}{
		{mmiW(1, 2, 3, 0x04, 0x08), "paddh $v1, $at, $v0"},
		{mmiW(1, 2, 3, 0x12, 0x08), "pextlw $v1, $at, $v0"},
		{mmiW(1, 2, 3, 0x1B, 0x28), "qfsrv $v1, $at, $v0"},
		{mmiW(1, 2, 3, 0x1C, 0x09), "pmulth $v1, $at, $v0"},
		{mmiW(1, 2, 3, 0, 0x18), "mult1 $v1, $at, $v0"},
		{mmiW(0, 0, 3, 0, 0x30), "pmfhl.lw $v1"},
		{mmiW(0, 2, 3, 7, 0x34), "psllh $v1, $v0, 7"},
		{special(1, 2, 3, 0, 0x18), "mult $v1, $at, $v0"},
		{special(0, 0, 3, 0, 0x28), "mfsa $v1"},
		{itype(0x1E, 4, 5, 0x20), "lq $a1, 32($a0)"},
	}
	for _, c := range cases {
		if got := DecodeWord(c.word, 0).Text; got != c.want {
			t.Errorf("DecodeWord(0x%08X) = %q, want %q", c.word, got, c.want)
		}
	}
}
