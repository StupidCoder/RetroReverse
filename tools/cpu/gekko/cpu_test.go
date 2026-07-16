package gekko

import (
	"math"
	"testing"
)

// testRAM is a flat 16 MiB big-endian memory implementing Bus. The locked cache is not
// part of it — the CPU answers those addresses itself — so a test can tell a scratchpad
// write from a memory write by looking at which one changed.
type testRAM struct {
	b [16 << 20]byte
}

func (m *testRAM) at(a uint32) uint32 { return a & 0xFFFFFF }

func (m *testRAM) Read8(a uint32) uint8 { return m.b[m.at(a)] }
func (m *testRAM) Read16(a uint32) uint16 {
	o := m.at(a)
	return uint16(m.b[o])<<8 | uint16(m.b[o+1])
}
func (m *testRAM) Read32(a uint32) uint32 {
	o := m.at(a)
	return uint32(m.b[o])<<24 | uint32(m.b[o+1])<<16 | uint32(m.b[o+2])<<8 | uint32(m.b[o+3])
}
func (m *testRAM) Write8(a uint32, v uint8) { m.b[m.at(a)] = v }
func (m *testRAM) Write16(a uint32, v uint16) {
	o := m.at(a)
	m.b[o], m.b[o+1] = uint8(v>>8), uint8(v)
}
func (m *testRAM) Write32(a uint32, v uint32) {
	o := m.at(a)
	m.b[o], m.b[o+1] = uint8(v>>24), uint8(v>>16)
	m.b[o+2], m.b[o+3] = uint8(v>>8), uint8(v)
}

// newTest assembles a program at 0x1000 and returns a CPU sitting at its first
// instruction, with translation off (so effective addresses are physical).
func newTest(words ...uint32) (*CPU, *testRAM) {
	m := &testRAM{}
	c := NewCPU(m)
	// No translation, no interrupts, vectors low — but the FPU on. MSR[FP] is not a
	// detail of the harness: with it clear every floating-point instruction traps to the
	// FP-unavailable vector instead of executing, so a test of floating-point behaviour
	// has to own the FPU the way the thread it stands for would.
	c.MSR = MSRFP
	c.PC = 0x1000
	for i, w := range words {
		m.Write32(0x1000+uint32(i*4), w)
	}
	return c, m
}

// run steps n instructions, failing if the core halts.
func run(t *testing.T, c *CPU, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		c.Step()
		if c.Halted {
			t.Fatalf("halted after %d steps: %s", i+1, c.HaltReason)
		}
	}
}

// --- Instruction encoders, so a test reads as assembly rather than as hex ---

func xform(op, d, a, b, xo, rc uint32) uint32 {
	return op<<26 | d<<21 | a<<16 | b<<11 | xo<<1 | rc
}
func dform(op, d, a, imm uint32) uint32 { return op<<26 | d<<21 | a<<16 | imm&0xFFFF }

// --- The tests ---

// srawi's carry rule: set only when the value is negative AND a one-bit fell off the
// bottom. Three of these four cases are the ones an implementation gets wrong.
func TestSrawiCarry(t *testing.T) {
	cases := []struct {
		name string
		v    uint32
		sh   uint32
		want uint32 // the expected result
		ca   bool
	}{
		{"negative, bits lost", 0xFFFFFFFF, 4, 0xFFFFFFFF, true},     // -1 >> 4 = -1, and ones fell off
		{"negative, no bits lost", 0xFFFFFFF0, 4, 0xFFFFFFFF, false}, // the low nibble was zero
		{"positive, bits lost", 0x0000000F, 4, 0, false},             // positive never sets carry
		{"shift of zero", 0xFFFFFFFF, 0, 0xFFFFFFFF, false},          // nothing was shifted out
	}
	for _, tc := range cases {
		c, _ := newTest(xform(31, 3, 4, tc.sh, 824, 0)) // srawi r4,r3,sh
		c.GPR[3] = tc.v
		run(t, c, 1)
		if c.GPR[4] != tc.want {
			t.Errorf("%s: srawi gave 0x%08X, want 0x%08X", tc.name, c.GPR[4], tc.want)
		}
		if got := c.XER&XERCA != 0; got != tc.ca {
			t.Errorf("%s: carry = %v, want %v", tc.name, got, tc.ca)
		}
	}
}

// The subtract carry is the carry out of a + ¬b + 1, so "set" means "no borrow".
func TestSubtractCarryIsNotBorrow(t *testing.T) {
	// subfc r5,r3,r4  =>  r5 = r4 - r3
	c, _ := newTest(xform(31, 5, 3, 4, 8, 0))
	c.GPR[3], c.GPR[4] = 1, 5 // 5 - 1: no borrow, so carry SET
	run(t, c, 1)
	if c.GPR[5] != 4 {
		t.Errorf("5-1 = %d", c.GPR[5])
	}
	if c.XER&XERCA == 0 {
		t.Error("5-1 did not borrow, so the carry should be set")
	}

	c, _ = newTest(xform(31, 5, 3, 4, 8, 0))
	c.GPR[3], c.GPR[4] = 5, 1 // 1 - 5: a borrow, so carry CLEAR
	run(t, c, 1)
	if c.XER&XERCA != 0 {
		t.Error("1-5 borrowed, so the carry should be clear")
	}
}

// divw by zero and the most-negative-over-minus-one both leave the result undefined but
// DO set overflow. The flags are the contract; the register is not.
func TestDivideOverflowSetsFlagsNotResult(t *testing.T) {
	// divwo. r5,r3,r4
	c, _ := newTest(xform(31, 5, 3, 4, 491|512, 1))
	c.GPR[3], c.GPR[4] = 100, 0
	run(t, c, 1)
	if c.XER&XEROV == 0 || c.XER&XERSO == 0 {
		t.Errorf("divide by zero: XER = 0x%08X, want OV and SO set", c.XER)
	}

	c, _ = newTest(xform(31, 5, 3, 4, 491|512, 0))
	c.GPR[3], c.GPR[4] = 0x80000000, 0xFFFFFFFF
	run(t, c, 1)
	if c.XER&XEROV == 0 {
		t.Error("0x80000000 / -1 should overflow")
	}
}

// The summary-overflow bit is sticky: it survives an operation that does not overflow,
// and only mcrxr (or a write to XER) clears it.
func TestSummaryOverflowIsSticky(t *testing.T) {
	c, _ := newTest(
		xform(31, 5, 3, 4, 266|512, 0), // addo r5,r3,r4 — overflows
		xform(31, 6, 3, 3, 266|512, 0), // addo r6,r3,r3 — does not
		xform(31, 0, 0, 0, 512, 0),     // mcrxr cr0 — the only thing that clears SO
	)
	c.GPR[3], c.GPR[4] = 0x7FFFFFFF, 0x7FFFFFFF
	run(t, c, 1)
	if c.XER&XERSO == 0 {
		t.Fatal("the overflow did not set SO")
	}
	c.GPR[3] = 1
	run(t, c, 1)
	if c.XER&XEROV != 0 {
		t.Error("OV should have been cleared by the non-overflowing add")
	}
	if c.XER&XERSO == 0 {
		t.Error("SO is sticky and should have survived")
	}
	run(t, c, 1)
	if c.XER&XERSO != 0 {
		t.Error("mcrxr should have cleared SO")
	}
}

// The rotate-and-mask, including the wrapped mask (mb > me) that the compiler uses for
// field extraction and that a transcription gets wrong.
func TestRotateMask(t *testing.T) {
	if got := mask32(0, 31); got != 0xFFFFFFFF {
		t.Errorf("mask(0,31) = 0x%08X", got)
	}
	if got := mask32(28, 31); got != 0x0000000F {
		t.Errorf("mask(28,31) = 0x%08X, want the low nibble", got)
	}
	if got := mask32(0, 3); got != 0xF0000000 {
		t.Errorf("mask(0,3) = 0x%08X, want the high nibble", got)
	}
	// The wrapped case: ones at both ends, a hole in the middle.
	if got := mask32(28, 3); got != 0xF000000F {
		t.Errorf("mask(28,3) = 0x%08X, want 0xF000000F — the mask must wrap", got)
	}
}

// A shift of 32 or more produces zero, rather than wrapping to a shift of zero, which is
// what a naive mask of five bits would do.
func TestShiftBeyondWidth(t *testing.T) {
	c, _ := newTest(xform(31, 3, 4, 5, 24, 0)) // slw r4,r3,r5
	c.GPR[3], c.GPR[5] = 0xFFFFFFFF, 32
	run(t, c, 1)
	if c.GPR[4] != 0 {
		t.Errorf("a shift by 32 gave 0x%08X, want 0 — the shift amount is six bits", c.GPR[4])
	}
}

// lwarx takes a reservation; stwcx. stores only while it holds, and says so in CR0[eq].
func TestReservation(t *testing.T) {
	c, m := newTest(
		xform(31, 3, 0, 4, 20, 0),  // lwarx r3,0,r4
		xform(31, 5, 0, 4, 150, 1), // stwcx. r5,0,r4
	)
	c.GPR[4] = 0x2000
	c.GPR[5] = 0xCAFE
	m.Write32(0x2000, 0xBEEF)
	run(t, c, 2)
	if m.Read32(0x2000) != 0xCAFE {
		t.Error("the store-conditional should have succeeded")
	}
	if c.CRField(0)&crEQ == 0 {
		t.Error("a successful stwcx. must set CR0[eq]")
	}

	// Now break the reservation with an intervening write, and the store must fail.
	c, m = newTest(
		xform(31, 3, 0, 4, 20, 0),  // lwarx
		dform(36, 6, 4, 0),         // stw r6,0(r4) — clears the reservation
		xform(31, 5, 0, 4, 150, 1), // stwcx.
	)
	c.GPR[4], c.GPR[5], c.GPR[6] = 0x2000, 0xCAFE, 0x1234
	run(t, c, 3)
	if m.Read32(0x2000) != 0x1234 {
		t.Errorf("the stwcx. wrote 0x%08X over the intervening store", m.Read32(0x2000))
	}
	if c.CRField(0)&crEQ != 0 {
		t.Error("a failed stwcx. must leave CR0[eq] clear")
	}
}

// A conditional branch to the link register may return AND fall through — the PowerPC
// shape that has no MIPS equivalent.
func TestConditionalBranchToLinkRegister(t *testing.T) {
	// bnelr, then a marker instruction. With CR0[eq] set, it does NOT return.
	c, _ := newTest(
		0x4C820020,         // bnelr — BO=4 (branch if false), BI=2 (eq)
		dform(14, 3, 0, 7), // li r3,7 — the fall-through
	)
	c.LR = 0x9000
	c.SetCRField(0, crEQ) // equal, so "not equal" is false, so it does not branch
	run(t, c, 2)
	if c.GPR[3] != 7 {
		t.Error("the conditional blr should have fallen through")
	}

	c, _ = newTest(0x4C820020, dform(14, 3, 0, 7))
	c.LR = 0x9000
	c.SetCRField(0, crGT) // not equal, so it returns
	run(t, c, 1)
	if c.PC != 0x9000 {
		t.Errorf("the conditional blr should have returned to 0x9000, went to 0x%08X", c.PC)
	}
}

// dcbz writes 32 zero bytes. It is not a hint, and a core that treated it as one would
// leave stale data where the program expects a cleared line.
func TestDCBZWritesZeroes(t *testing.T) {
	c, m := newTest(xform(31, 0, 0, 4, 1014, 0)) // dcbz 0,r4
	c.GPR[4] = 0x2010                            // deliberately not line-aligned
	for i := uint32(0); i < 64; i++ {
		m.Write8(0x2000+i, 0xFF)
	}
	run(t, c, 1)
	// The line containing 0x2010 is 0x2000..0x2020, and only that line.
	for i := uint32(0x2000); i < 0x2020; i++ {
		if m.Read8(i) != 0 {
			t.Fatalf("0x%X was not zeroed", i)
		}
	}
	if m.Read8(0x2020) != 0xFF {
		t.Error("dcbz zeroed past the end of its cache line")
	}
}

// The locked cache is memory that is not memory: writes to it must not reach the bus, and
// they must survive.
func TestLockedCache(t *testing.T) {
	c, m := newTest(
		xform(4, 0, 0, 4, 1014, 0), // dcbz_l 0,r4 — allocate a line in the scratchpad
		dform(36, 5, 4, 0),         // stw r5,0(r4)
		dform(32, 6, 4, 0),         // lwz r6,0(r4)
	)
	c.setHID2(HID2LCE | HID2PSE)
	c.GPR[4] = LockedCacheBase
	c.GPR[5] = 0xABCD1234
	run(t, c, 3)

	if c.GPR[6] != 0xABCD1234 {
		t.Errorf("the locked cache read back 0x%08X, want 0xABCD1234", c.GPR[6])
	}
	// The write must NOT have reached main memory. testRAM masks to 24 bits, so the
	// scratchpad address 0xE0000000 would alias to 0 — and finding it there is exactly
	// the bug this checks for.
	if m.Read32(0) != 0 {
		t.Errorf("a locked-cache write reached the bus at 0x%08X", m.Read32(0))
	}
}

// The decrementer counts down and raises one exception when it passes below zero —
// one, not one per instruction spent negative.
func TestDecrementerFiresOnce(t *testing.T) {
	// A loop of nops at the vector, so the handler is somewhere to land.
	c, m := newTest()
	for i := 0; i < 4000; i++ {
		m.Write32(0x1000+uint32(i*4), 0x60000000) // nop
	}
	m.Write32(VecDecrementer, 0x60000000) // a nop at the handler
	c.MSR = MSREE                         // interrupts on, vectors low
	c.setDEC(1)

	fired := 0
	for i := 0; i < 100; i++ {
		before := c.PC
		c.Step()
		if c.Halted {
			t.Fatal(c.HaltReason)
		}
		if c.PC == VecDecrementer && before != VecDecrementer {
			fired++
			// The handler returns to where we were.
			c.MSR |= MSREE
			c.PC = c.SRR0
		}
	}
	if fired != 1 {
		t.Errorf("the decrementer fired %d times, want exactly 1", fired)
	}
}

// An exception saves the resume address and the machine state, and rfi puts them back.
func TestExceptionAndRFI(t *testing.T) {
	c, m := newTest()
	m.Write32(0x1000, 0x44000002)          // sc
	m.Write32(VecSyscall, 0x4C000064)      // rfi at the handler
	m.Write32(0x1004, dform(14, 3, 0, 42)) // li r3,42 — where we should come back to
	c.MSR = MSREE | MSRDR | MSRIR

	// Translation must be OFF inside the handler: that is what makes a handler able to
	// run when translation is what failed.
	c.IBAT[0] = [2]uint32{0x00001FFF | 2, 0x00000000 | 2} // identity-map the low 256 MiB
	c.DBAT[0] = [2]uint32{0x00001FFF | 2, 0x00000000 | 2}

	c.Step() // sc
	if c.PC != VecSyscall {
		t.Fatalf("sc went to 0x%08X, want the syscall vector", c.PC)
	}
	if c.SRR0 != 0x1004 {
		t.Errorf("SRR0 = 0x%08X, want 0x1004 (the instruction after sc)", c.SRR0)
	}
	if c.MSR&MSREE != 0 {
		t.Error("an exception must clear MSR[EE], or the handler could be preempted")
	}
	if c.MSR&(MSRIR|MSRDR) != 0 {
		t.Error("an exception must clear translation, or a handler could not run when translation failed")
	}

	c.Step() // rfi
	if c.PC != 0x1004 {
		t.Errorf("rfi resumed at 0x%08X, want 0x1004", c.PC)
	}
	if c.MSR&MSREE == 0 {
		t.Error("rfi must restore MSR[EE] from SRR1")
	}
}

// The block-address translation is how a GameCube reaches memory at all: 0x80000000 is
// physical 0, and so is 0xC0000000.
func TestBATTranslation(t *testing.T) {
	c, m := newTest()
	// A 256 MiB block at 0x80000000 -> physical 0, valid in supervisor state.
	c.DBAT[0] = [2]uint32{0x80000000 | (0x7FF << 2) | 2, 0x00000000 | 2}
	c.MSR = MSRDR

	m.Write32(0x1234, 0xFEEDFACE)
	pa, ok := c.Translate(0x80001234, false, false)
	if !ok || pa != 0x1234 {
		t.Errorf("0x80001234 translated to 0x%08X (ok=%v), want 0x1234", pa, ok)
	}
	if got := c.read32(0x80001234); got != 0xFEEDFACE {
		t.Errorf("read through the BAT gave 0x%08X", got)
	}
}

// An address no BAT maps must halt loudly rather than silently reading zero — because on
// this machine there is no page table to fall back to, and a zero would be a lie.
func TestUnmappedAddressHalts(t *testing.T) {
	c, _ := newTest()
	c.MSR = MSRDR // translation on, but no BAT set
	c.read32(0x80001234)
	if !c.Halted {
		t.Fatal("an unmapped address must halt, not read zero")
	}
}

// lfs broadcasts: a single-precision load lands the value in BOTH halves of the register.
func TestLoadSingleBroadcasts(t *testing.T) {
	c, m := newTest(dform(48, 1, 4, 0)) // lfs f1,0(r4)
	c.GPR[4] = 0x2000
	m.Write32(0x2000, math.Float32bits(1.5))
	run(t, c, 1)
	if c.FPR[1].PS0 != 1.5 || c.FPR[1].PS1 != 1.5 {
		t.Errorf("lfs gave (%v, %v), want both halves 1.5", c.FPR[1].PS0, c.FPR[1].PS1)
	}
}

// A scalar double operation must leave PS1 alone. Getting this wrong yields a core that
// passes every scalar test and then draws wrong geometry.
func TestScalarOpPreservesPS1(t *testing.T) {
	c, _ := newTest(xform(63, 3, 1, 2, 21, 0) | 0) // fadd f3,f1,f2
	c.FPR[1] = FPR{PS0: 1, PS1: 111}
	c.FPR[2] = FPR{PS0: 2, PS1: 222}
	c.FPR[3] = FPR{PS0: 0, PS1: 999}
	run(t, c, 1)
	if c.FPR[3].PS0 != 3 {
		t.Errorf("fadd gave %v, want 3", c.FPR[3].PS0)
	}
	if c.FPR[3].PS1 != 999 {
		t.Errorf("fadd clobbered PS1 (%v); a scalar double op must not touch it", c.FPR[3].PS1)
	}
}

// The paired-single cross-slot wiring — the part that must not be guessed.
func TestPairedSingleSlots(t *testing.T) {
	// ps_sum0 f3,f1,f2,f4 : ps0 = A.ps0 + B.ps1, ps1 = C.ps1
	c, _ := newTest(4<<26 | 3<<21 | 1<<16 | 2<<11 | 4<<6 | 10<<1)
	c.setHID2(HID2PSE)
	c.FPR[1] = FPR{PS0: 1, PS1: 10}
	c.FPR[2] = FPR{PS0: 2, PS1: 20}
	c.FPR[4] = FPR{PS0: 3, PS1: 30}
	run(t, c, 1)
	if c.FPR[3].PS0 != 21 { // A.ps0(1) + B.ps1(20)
		t.Errorf("ps_sum0 ps0 = %v, want 21 (A.ps0 + B.ps1)", c.FPR[3].PS0)
	}
	if c.FPR[3].PS1 != 30 { // C.ps1
		t.Errorf("ps_sum0 ps1 = %v, want 30 (C.ps1)", c.FPR[3].PS1)
	}

	// ps_muls0 f3,f1,f4 : both slots multiplied by C's ps0 — the broadcast.
	c, _ = newTest(4<<26 | 3<<21 | 1<<16 | 0<<11 | 4<<6 | 12<<1)
	c.setHID2(HID2PSE)
	c.FPR[1] = FPR{PS0: 2, PS1: 3}
	c.FPR[4] = FPR{PS0: 5, PS1: 100}
	run(t, c, 1)
	if c.FPR[3].PS0 != 10 || c.FPR[3].PS1 != 15 {
		t.Errorf("ps_muls0 gave (%v,%v), want (10,15) — both slots times C.ps0",
			c.FPR[3].PS0, c.FPR[3].PS1)
	}

	// ps_merge10 f3,f1,f2 : (A.ps1, B.ps0)
	c, _ = newTest(4<<26 | 3<<21 | 1<<16 | 2<<11 | 592<<1)
	c.setHID2(HID2PSE)
	c.FPR[1] = FPR{PS0: 1, PS1: 10}
	c.FPR[2] = FPR{PS0: 2, PS1: 20}
	run(t, c, 1)
	if c.FPR[3].PS0 != 10 || c.FPR[3].PS1 != 2 {
		t.Errorf("ps_merge10 gave (%v,%v), want (10,2)", c.FPR[3].PS0, c.FPR[3].PS1)
	}
}

// A paired-single instruction with the unit switched off must halt, not quietly execute.
func TestPairedSingleRequiresHID2(t *testing.T) {
	c, _ := newTest(4<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1) // ps_add, unit off
	c.Step()
	if !c.Halted {
		t.Error("a ps_ instruction with HID2[PSE] clear must halt")
	}
}

// The quantiser, in both directions and at both ends of the scale. The load divides by
// 2^scale and the store multiplies by it — reversing that is the single easiest way to
// build a core that works and puts the geometry in the wrong place.
func TestQuantisedLoadStore(t *testing.T) {
	// GQR0: store type s8, load type s8, both scaled by 2^2 = 4.
	//   store: v * 4 -> byte      load: byte / 4 -> v
	c, m := newTest(
		0xF0040000|1<<15, // psq_st f0,0(r4),1,gqr0  (one value)
		0xE0240000|1<<15, // psq_l  f1,0(r4),1,gqr0
	)
	c.setHID2(HID2PSE)
	c.GQR[0] = (2 << 8) | qS8 | (2 << 24) | (qS8 << 16)
	c.GPR[4] = 0x2000
	c.FPR[0] = FPR{PS0: 3.0, PS1: 0}
	run(t, c, 2)

	// 3.0 * 4 = 12, stored as a signed byte.
	if got := m.Read8(0x2000); got != 12 {
		t.Errorf("psq_st stored %d, want 12 (3.0 scaled by 2^2)", got)
	}
	// And back: 12 / 4 = 3.0.
	if c.FPR[1].PS0 != 3.0 {
		t.Errorf("psq_l read back %v, want 3.0", c.FPR[1].PS0)
	}
	// A single-value load sets PS1 to 1.0 — not zero, and not left alone. It is what
	// makes a 3-vector load come out as a homogeneous coordinate.
	if c.FPR[1].PS1 != 1.0 {
		t.Errorf("a one-value psq_l set PS1 to %v, want 1.0", c.FPR[1].PS1)
	}
}

// The store saturates rather than wrapping. A game that scales its vertices to fill a
// signed byte relies on the clamp; a core that wrapped would fold the far side of a model
// through the near side.
func TestQuantisedStoreSaturates(t *testing.T) {
	c, m := newTest(0xF0040000 | 1<<15) // psq_st f0,0(r4),1,gqr0
	c.setHID2(HID2PSE)
	c.GQR[0] = qS8 // signed byte, scale 0
	c.GPR[4] = 0x2000
	c.FPR[0] = FPR{PS0: 1000.0} // far past a signed byte
	run(t, c, 1)
	if got := int8(m.Read8(0x2000)); got != 127 {
		t.Errorf("psq_st of 1000 into a signed byte gave %d, want 127 (saturated)", got)
	}

	c, m = newTest(0xF0040000 | 1<<15)
	c.setHID2(HID2PSE)
	c.GQR[0] = qS8
	c.GPR[4] = 0x2000
	c.FPR[0] = FPR{PS0: -1000.0}
	run(t, c, 1)
	if got := int8(m.Read8(0x2000)); got != -128 {
		t.Errorf("psq_st of -1000 gave %d, want -128 (saturated)", got)
	}
}

// fmadd rounds ONCE, on the exact product-plus-addend. A core that computed a*c + b would
// round twice and differ in the last bit — on an operation a 3-D game runs millions of
// times a frame.
func TestFusedMultiplyAddRoundsOnce(t *testing.T) {
	// Operands chosen so the product needs more than 53 bits and the fused and unfused
	// results genuinely differ.
	a, cc, b := 1.0000000000000002, 1.0000000000000002, -1.0
	c, _ := newTest(63<<26 | 3<<21 | 1<<16 | 2<<11 | 4<<6 | 29<<1) // fmadd f3,f1,f4,f2
	c.FPR[1] = FPR{PS0: a}
	c.FPR[4] = FPR{PS0: cc}
	c.FPR[2] = FPR{PS0: b}
	run(t, c, 1)

	want := math.FMA(a, cc, b)
	unfused := a*cc + b
	if want == unfused {
		t.Skip("these operands do not distinguish the two; the test needs better ones")
	}
	if c.FPR[3].PS0 != want {
		t.Errorf("fmadd gave %v, want the fused %v (the unfused answer is %v)",
			c.FPR[3].PS0, want, unfused)
	}
}

// A single-precision op computes in double and rounds once to single.
func TestSinglePrecisionRounds(t *testing.T) {
	c, _ := newTest(59<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1) // fadds f3,f1,f2
	c.FPR[1] = FPR{PS0: 1.0}
	c.FPR[2] = FPR{PS0: 1e-20} // far below single precision: the sum rounds back to 1.0
	run(t, c, 1)
	if c.FPR[3].PS0 != 1.0 {
		t.Errorf("fadds gave %v, want exactly 1.0 after rounding to single", c.FPR[3].PS0)
	}
	// And it broadcasts, like every single-precision result.
	if c.FPR[3].PS1 != 1.0 {
		t.Errorf("fadds left PS1 = %v, want the result broadcast", c.FPR[3].PS1)
	}
}

// lmw and stmw move a run of registers up to r31.
func TestLoadStoreMultiple(t *testing.T) {
	c, m := newTest(
		dform(47, 29, 4, 0), // stmw r29,0(r4) — r29, r30, r31
		dform(46, 26, 4, 0), // lmw  r26,0(r4) — r26..r31, reading what we wrote plus more
	)
	c.GPR[4] = 0x2000
	c.GPR[29], c.GPR[30], c.GPR[31] = 0xAAAA, 0xBBBB, 0xCCCC
	run(t, c, 1)
	if m.Read32(0x2000) != 0xAAAA || m.Read32(0x2008) != 0xCCCC {
		t.Error("stmw did not lay the registers down in order")
	}
	run(t, c, 1)
	if c.GPR[26] != 0xAAAA || c.GPR[28] != 0xCCCC {
		t.Errorf("lmw read back r26=0x%X r28=0x%X", c.GPR[26], c.GPR[28])
	}
}

// The byte-reversed loads and stores — the little-endian accessors on a big-endian
// machine, and exactly the ones a careless transcription implements backwards.
func TestByteReversed(t *testing.T) {
	c, m := newTest(xform(31, 3, 0, 4, 534, 0)) // lwbrx r3,0,r4
	c.GPR[4] = 0x2000
	m.Write32(0x2000, 0x11223344)
	run(t, c, 1)
	if c.GPR[3] != 0x44332211 {
		t.Errorf("lwbrx gave 0x%08X, want 0x44332211", c.GPR[3])
	}
}

// A snapshot must restore to a machine that runs identically — including the locked
// cache, which holds data that exists nowhere else.
func TestSnapshotRoundTrip(t *testing.T) {
	c, _ := newTest(dform(14, 3, 3, 1)) // addi r3,r3,1
	c.setHID2(HID2LCE)
	c.LC.Write32(LockedCacheBase, 0xDEADBEEF)
	c.GPR[3] = 41
	c.XER = XERSO
	c.FPR[7] = FPR{PS0: 1.5, PS1: 2.5}

	s := c.Snapshot()
	run(t, c, 1)
	if c.GPR[3] != 42 {
		t.Fatal("the program did not run")
	}

	c.Restore(s)
	if c.GPR[3] != 41 || c.PC != 0x1000 {
		t.Error("Restore did not put the registers back")
	}
	if c.LC.Read32(LockedCacheBase) != 0xDEADBEEF {
		t.Error("the locked cache was not restored — its contents exist nowhere else")
	}
	if c.FPR[7].PS1 != 2.5 {
		t.Error("the second half of a paired register was not restored")
	}
	run(t, c, 1)
	if c.GPR[3] != 42 {
		t.Error("the restored machine did not run identically")
	}
}

// An unknown SPR halts rather than reading zero. A zero is an answer, and a wrong one.
func TestUnknownSPRHalts(t *testing.T) {
	c, _ := newTest(xform(31, 3, 0, 0, 339, 0) | (999&0x1F)<<16 | (999>>5)<<11)
	c.Step()
	if !c.Halted {
		t.Error("mfspr of an unknown register must halt, not read zero")
	}
}
