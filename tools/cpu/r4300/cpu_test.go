package r4300

// Assembled-loop tests: each case builds a short instruction sequence, runs it,
// and checks the architectural state. There is no published per-instruction
// conformance suite for the VR4300 (the SingleStepTests sets cover the R3000A
// but not this core), so the cases here concentrate on what MIPS III adds over
// the R3000A already validated in tools/cpu/mips, and on the places where a
// 64-bit core silently diverges from a 32-bit one:
//
//   - 32-bit results sign-extend into the full register
//   - the branch-likely family annuls its delay slot
//   - the unaligned load/store family, whose "left" is the high-order end
//   - 64-bit multiply/divide, including the degenerate cases
//
// The oracle boots lemmy-64/n64-systemtest for broad conformance; see
// tools/platform/n64.

import (
	"encoding/binary"
	"testing"
)

// testRAM is a flat big-endian memory implementing Bus, addressed physically.
type testRAM struct{ b []byte }

func newRAM() *testRAM { return &testRAM{b: make([]byte, 0x10000)} }

func (r *testRAM) Read(a uint32) byte         { return r.b[a&0xFFFF] }
func (r *testRAM) Write(a uint32, v byte)     { r.b[a&0xFFFF] = v }
func (r *testRAM) Read32(a uint32) uint32     { return binary.BigEndian.Uint32(r.b[a&0xFFFF:]) }
func (r *testRAM) Write32(a uint32, v uint32) { binary.BigEndian.PutUint32(r.b[a&0xFFFF:], v) }

// base is where test programs are assembled: KSEG0, which maps to physical 0.
const testBase = 0x80000000

// newTest builds a CPU running a program at testBase, with the FPU enabled and
// the 32-register FP file selected (FR), as libultra leaves it.
func newTest(words ...uint32) (*CPU, *testRAM) {
	ram := newRAM()
	for i, w := range words {
		ram.Write32(uint32(i*4), w)
	}
	c := NewCPU(ram)
	c.COP0[cop0Status] = statusCU1 | statusFR // clears ERL/BEV: KSEG0 is live
	c.SetPC(testBase)
	return c, ram
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

// --- instruction encoders ---------------------------------------------------

func rt(op, rs, rtr, rd, sh, funct uint32) uint32 {
	return op<<26 | rs<<21 | rtr<<16 | rd<<11 | sh<<6 | funct
}
func it(op, rs, rtr uint32, imm uint16) uint32 {
	return op<<26 | rs<<21 | rtr<<16 | uint32(imm)
}

const nop = 0

func TestSignExtension32BitOps(t *testing.T) {
	// addu of two positives whose sum sets bit 31 must sign-extend into bits
	// 32..63; daddu of the same operands must not. This is the single easiest
	// way to get a 64-bit MIPS core subtly wrong.
	c, _ := newTest(
		rt(0, 1, 2, 3, 0, 0x21), // addu  $3, $1, $2
		rt(0, 1, 2, 4, 0, 0x2D), // daddu $4, $1, $2
	)
	c.SetReg(1, 0x7FFFFFFF)
	c.SetReg(2, 1)
	run(t, c, 2)

	if got, want := c.R[3], uint64(0xFFFFFFFF80000000); got != want {
		t.Errorf("addu: got %016X want %016X", got, want)
	}
	if got, want := c.R[4], uint64(0x0000000080000000); got != want {
		t.Errorf("daddu: got %016X want %016X", got, want)
	}
}

func TestDoubleShifts(t *testing.T) {
	// The "32" forms add 32 to the shift amount, reaching 32..63.
	c, _ := newTest(
		rt(0, 0, 1, 2, 0, 0x3C), // dsll32 $2, $1, 0   -> $1 << 32
		rt(0, 0, 2, 3, 4, 0x3E), // dsrl32 $3, $2, 4   -> $2 >> 36
		rt(0, 0, 1, 4, 8, 0x38), // dsll   $4, $1, 8
	)
	c.SetReg(1, 1)
	run(t, c, 3)

	if got, want := c.R[2], uint64(1)<<32; got != want {
		t.Errorf("dsll32: got %016X want %016X", got, want)
	}
	if got, want := c.R[3], uint64(0); got != want {
		t.Errorf("dsrl32: got %016X want %016X", got, want)
	}
	if got, want := c.R[4], uint64(1)<<8; got != want {
		t.Errorf("dsll: got %016X want %016X", got, want)
	}
}

func TestArithmeticShiftIgnoresTheHighWord(t *testing.T) {
	// sra and srav are 32-BIT operations on a 64-bit register. MIPS defines the
	// result as rt[31:0] shifted right with bit 31 replicated into the vacated bits,
	// then sign-extended to 64: the high word is never read, and the sign comes from
	// bit 31, not bit 63.
	//
	// This test previously asserted the opposite — that the shift reads all 64 bits —
	// and both this core and the R5900 implemented that. Nothing external ever
	// checked it: the R5900's diff test compares the two cores against each other,
	// and they were wrong together. The evidence that settles it:
	//
	//   - The MIPS III/64 pseudocode: temp <- (GPR[rt]_31)^s || GPR[rt]_31..s.
	//   - tools/cpu/mips, the 32-bit R3000A core checked against an external vector
	//     set, shifts int32(rt).
	//   - Shipped game code depends on it. Jak and Daxter's draw-string does
	//     `qmfc2 $t5,$vf1` / `bltz` / `sra $t5,$t5,31` / `bltz` to test a vector's
	//     y then x lane: sra 31 exists purely to bring bit 31 into the sign for the
	//     second bltz. Shifting 64 bits drags the y lane down into the sign and
	//     answers about the wrong lane — every glyph clipped, no text ever drawn.
	//
	// The manual calls rt UNPREDICTABLE when it is not a sign-extended word, which
	// is why an implementer can talk themselves into either rule. Silicon ignores
	// the high half, and the code above only runs because it does.
	c, _ := newTest(
		rt(0, 0, 1, 2, 4, 0x03), // sra  $2, $1, 4
		rt(0, 3, 1, 4, 0, 0x07), // srav $4, $1, $3
	)
	c.SetReg(1, 0x00000000FFFFFFFF) // low half all ones, high half zero
	c.SetReg(3, 4)
	run(t, c, 2)

	// rt[31:0] is 0xFFFFFFFF = -1; -1 >> 4 is -1, sign-extended to 64. The high
	// word's zeros must not turn it into 0x0FFFFFFF.
	if got, want := c.R[2], uint64(0xFFFFFFFFFFFFFFFF); got != want {
		t.Errorf("sra: got %016X want %016X", got, want)
	}
	if got, want := c.R[4], uint64(0xFFFFFFFFFFFFFFFF); got != want {
		t.Errorf("srav: got %016X want %016X", got, want)
	}

	// The mirror case: a high word of all ones must not make a positive low word
	// negative.
	c2, _ := newTest(rt(0, 0, 1, 2, 4, 0x03))
	c2.SetReg(1, 0xFFFFFFFF7FFFFFF0)
	run(t, c2, 1)
	if got, want := c2.R[2], uint64(0x07FFFFFF); got != want {
		t.Errorf("sra of a positive low word under a negative high word: got %016X want %016X", got, want)
	}

	// And an ordinary sign-extended negative word still sign-extends.
	c3, _ := newTest(rt(0, 0, 1, 2, 4, 0x03))
	c3.SetReg(1, 0xFFFFFFFFFFFFFF00)
	run(t, c3, 1)
	if got, want := c3.R[2], uint64(0xFFFFFFFFFFFFFFF0); got != want {
		t.Errorf("sra of a negative value: got %016X want %016X", got, want)
	}
}

func TestBranchLikelyAnnulsWhenNotTaken(t *testing.T) {
	// beql $1,$0 with $1 != 0: not taken, so the delay slot is annulled and the
	// instruction after it runs.
	c, _ := newTest(
		it(0x14, 1, 0, 2),      // beql  $1, $0, +2
		it(0x09, 0, 2, 0x0BAD), // addiu $2, $0, 0xBAD   (delay slot: annulled)
		it(0x09, 0, 3, 7),      // addiu $3, $0, 7
	)
	c.SetReg(1, 1)
	run(t, c, 2)

	if c.R[2] != 0 {
		t.Errorf("annulled delay slot executed: $2 = %016X", c.R[2])
	}
	if c.R[3] != 7 {
		t.Errorf("instruction after the annulled slot did not run: $3 = %016X", c.R[3])
	}
}

func TestBranchLikelyRunsSlotWhenTaken(t *testing.T) {
	// beql $1,$0 with $1 == 0: taken, so the delay slot runs and control lands
	// on the target, skipping the instruction between.
	c, _ := newTest(
		it(0x14, 1, 0, 2), // beql  $1, $0, +2  -> target is word[3]
		it(0x09, 0, 2, 5), // addiu $2, $0, 5   (delay slot: runs)
		it(0x09, 0, 3, 7), // addiu $3, $0, 7   (skipped)
		it(0x09, 0, 4, 9), // addiu $4, $0, 9
	)
	c.SetReg(1, 0)
	run(t, c, 3)

	if c.R[2] != 5 {
		t.Errorf("taken delay slot did not run: $2 = %016X", c.R[2])
	}
	if c.R[3] != 0 {
		t.Errorf("branch did not skip to its target: $3 = %016X", c.R[3])
	}
	if c.R[4] != 9 {
		t.Errorf("target did not run: $4 = %016X", c.R[4])
	}
}

func TestUnalignedLoadMerge(t *testing.T) {
	// The lwl/lwr idiom assembles an unaligned word. On a big-endian machine
	// "left" is the high-order end, so lwl takes the bytes from the effective
	// address to the end of its aligned word.
	//
	// Memory: 0x100: 11 22 33 44   0x104: 55 66 77 88
	// The unaligned word at 0x102 is 33 44 55 66.
	c, ram := newTest(
		it(0x22, 5, 1, 2), // lwl $1, 2($5)
		it(0x26, 5, 1, 5), // lwr $1, 5($5)
	)
	ram.Write32(0x100, 0x11223344)
	ram.Write32(0x104, 0x55667788)
	c.SetReg(5, 0x80000100)
	c.SetReg(1, 0xFFFFFFFFFFFFFFFF)
	run(t, c, 2)

	if got, want := c.R[1], uint64(0x33445566); got != want {
		t.Errorf("lwl/lwr: got %016X want %016X", got, want)
	}
}

func TestUnalignedStoreMerge(t *testing.T) {
	// swl/swr write an unaligned word without disturbing its neighbours.
	c, ram := newTest(
		it(0x2A, 5, 1, 2), // swl $1, 2($5)
		it(0x2E, 5, 1, 5), // swr $1, 5($5)
	)
	ram.Write32(0x100, 0x11223344)
	ram.Write32(0x104, 0x55667788)
	c.SetReg(5, 0x80000100)
	c.SetReg(1, 0xAABBCCDD)
	run(t, c, 2)

	if got, want := ram.Read32(0x100), uint32(0x1122AABB); got != want {
		t.Errorf("swl: [0x100] got %08X want %08X", got, want)
	}
	if got, want := ram.Read32(0x104), uint32(0xCCDD7788); got != want {
		t.Errorf("swr: [0x104] got %08X want %08X", got, want)
	}
}

func TestLoadLinkedStoreConditional(t *testing.T) {
	c, ram := newTest(
		it(0x30, 5, 1, 0), // ll  $1, 0($5)
		it(0x38, 5, 2, 0), // sc  $2, 0($5)
		it(0x38, 5, 3, 0), // sc  $3, 0($5)   -- LLBit consumed: must fail
	)
	ram.Write32(0x200, 0xCAFEF00D)
	c.SetReg(5, 0x80000200)
	c.SetReg(2, 0x12345678)
	c.SetReg(3, 0xDEADBEEF)
	run(t, c, 3)

	if got, want := c.R[1], sext32(0xCAFEF00D); got != want {
		t.Errorf("ll: got %016X want %016X", got, want)
	}
	if c.R[2] != 1 {
		t.Errorf("sc should have succeeded: $2 = %016X", c.R[2])
	}
	if got, want := ram.Read32(0x200), uint32(0x12345678); got != want {
		t.Errorf("sc did not store: got %08X want %08X", got, want)
	}
	if c.R[3] != 0 {
		t.Errorf("second sc should have failed: $3 = %016X", c.R[3])
	}
	if got, want := ram.Read32(0x200), uint32(0x12345678); got != want {
		t.Errorf("failed sc must not store: got %08X want %08X", got, want)
	}
}

func TestDoubleMultiplyDivide(t *testing.T) {
	c, _ := newTest(
		rt(0, 1, 2, 0, 0, 0x1C), // dmult $1, $2   -2 * 3 = -6
		rt(0, 3, 4, 0, 0, 0x1E), // ddiv  $3, $4   7 / -2 = -3 rem 1
	)
	c.SetReg(1, ^uint64(1)) // -2
	c.SetReg(2, 3)
	run(t, c, 1)
	if got, want := c.LO, ^uint64(5); got != want { // -6
		t.Errorf("dmult LO: got %016X want %016X", got, want)
	}
	if got, want := c.HI, ^uint64(0); got != want { // -1
		t.Errorf("dmult HI: got %016X want %016X", got, want)
	}

	c.SetReg(3, 7)
	c.SetReg(4, ^uint64(1)) // -2
	run(t, c, 1)
	if got, want := c.LO, ^uint64(2); got != want { // -3
		t.Errorf("ddiv LO: got %016X want %016X", got, want)
	}
	if got, want := c.HI, uint64(1); got != want {
		t.Errorf("ddiv HI: got %016X want %016X", got, want)
	}
}

func TestDivideByZeroIsDefined(t *testing.T) {
	// Real code relies on the hardware's degenerate results rather than trapping.
	c, _ := newTest(
		rt(0, 1, 2, 0, 0, 0x1E), // ddiv $1, $2   with $2 == 0
	)
	c.SetReg(1, 5)
	c.SetReg(2, 0)
	run(t, c, 1)
	if got, want := c.HI, uint64(5); got != want {
		t.Errorf("ddiv by zero HI: got %016X want %016X", got, want)
	}
	if got, want := c.LO, ^uint64(0); got != want {
		t.Errorf("ddiv by zero LO: got %016X want %016X", got, want)
	}
}

func TestMult32SignExtendsResult(t *testing.T) {
	c, _ := newTest(rt(0, 1, 2, 0, 0, 0x18)) // mult $1, $2
	c.SetReg(1, 0x00010000)
	c.SetReg(2, 0x00010000) // product 0x1_0000_0000: LO = 0, HI = 1
	run(t, c, 1)
	if c.LO != 0 || c.HI != 1 {
		t.Errorf("mult: LO=%016X HI=%016X want 0 / 1", c.LO, c.HI)
	}

	c2, _ := newTest(rt(0, 1, 2, 0, 0, 0x18))
	c2.SetReg(1, 0xFFFFFFFF) // -1 as a 32-bit operand
	c2.SetReg(2, 2)
	run(t, c2, 1)
	if got, want := c2.LO, sext32(0xFFFFFFFE); got != want {
		t.Errorf("mult LO sign-extension: got %016X want %016X", got, want)
	}
}

func TestCacheIsANoOp(t *testing.T) {
	// libultra brackets every DMA with osInvalDCache / osWritebackDCache, so a
	// core that halts on `cache` never finishes booting.
	c, _ := newTest(it(0x2F, 0, 0, 0)) // cache 0, 0($0)
	c.Step()
	if c.Halted {
		t.Fatalf("cache halted the core: %s", c.HaltReason)
	}
	if c.Steps != 1 {
		t.Errorf("cache did not retire: Steps = %d", c.Steps)
	}
}

func TestCountCompareRaisesIP7(t *testing.T) {
	// Count advances once per two instructions; reaching Compare sets the timer
	// interrupt. Interrupts stay disabled here so the bit can be observed
	// without taking the exception.
	words := make([]uint32, 64)
	c, _ := newTest(words...) // all nop
	c.COP0[cop0Compare] = 10

	for i := 0; i < 19; i++ {
		c.Step()
	}
	if c.COP0[cop0Cause]&causeIP7 != 0 {
		t.Fatalf("IP7 raised early: Count = %d", uint32(c.COP0[cop0Count]))
	}
	c.Step()
	if uint32(c.COP0[cop0Count]) != 10 {
		t.Fatalf("Count = %d, want 10", uint32(c.COP0[cop0Count]))
	}
	if c.COP0[cop0Cause]&causeIP7 == 0 {
		t.Error("Count reached Compare but IP7 was not raised")
	}

	// Writing Compare acknowledges the interrupt.
	c.writeCop0(cop0Compare, 500)
	if c.COP0[cop0Cause]&causeIP7 != 0 {
		t.Error("writing Compare did not clear IP7")
	}
}

func TestExceptionVectorsFollowBEV(t *testing.T) {
	// A syscall vectors to the general handler; the base moves with BEV, and the
	// EPC records the faulting instruction.
	for _, tc := range []struct {
		bev  bool
		want uint64
	}{
		{false, 0xFFFFFFFF80000180},
		{true, 0xFFFFFFFFBFC00380},
	} {
		c, _ := newTest(rt(0, 0, 0, 0, 0, 0x0C)) // syscall
		if tc.bev {
			c.COP0[cop0Status] |= statusBEV
		}
		c.Step()
		if c.PC != tc.want {
			t.Errorf("BEV=%v: vectored to %016X want %016X", tc.bev, c.PC, tc.want)
		}
		if c.COP0[cop0EPC] != testBase {
			t.Errorf("BEV=%v: EPC = %016X want %016X", tc.bev, c.COP0[cop0EPC], uint64(testBase))
		}
		if c.COP0[cop0Status]&statusEXL == 0 {
			t.Errorf("BEV=%v: EXL not set on exception entry", tc.bev)
		}
	}
}

func TestExceptionInDelaySlotSetsBD(t *testing.T) {
	// EPC must point at the branch, not at the faulting delay-slot instruction,
	// so that eret re-executes the branch.
	c, _ := newTest(
		it(0x04, 0, 0, 4),       // beq $0,$0,+4   (always taken)
		rt(0, 0, 0, 0, 0, 0x0C), // syscall        (delay slot)
	)
	run(t, c, 1)
	c.Step() // the delay slot faults

	if c.COP0[cop0Cause]&causeBD == 0 {
		t.Error("BD not set for a fault in a delay slot")
	}
	if got, want := c.COP0[cop0EPC], uint64(testBase); got != want {
		t.Errorf("EPC = %016X want the branch at %016X", got, want)
	}
}

func TestEretReturnsToEPC(t *testing.T) {
	c, _ := newTest(rt(0x10, 0x10, 0, 0, 0, 0x18)) // eret
	c.COP0[cop0Status] |= statusEXL
	c.COP0[cop0EPC] = 0x80001234
	c.LLBit = true
	c.Step()

	if c.PC != 0x80001234 {
		t.Errorf("eret went to %016X want 0x80001234", c.PC)
	}
	if c.COP0[cop0Status]&statusEXL != 0 {
		t.Error("eret did not clear EXL")
	}
	if c.LLBit {
		t.Error("eret did not break the load-linked")
	}
}

func TestTLBMapsAndFaults(t *testing.T) {
	// A mapped read through KUSEG: fill one 4 KiB entry, then load through it.
	c, ram := newTest(it(0x23, 0, 1, 0)) // lw $1, 0($0)  -- vaddr 0 is mapped
	ram.Write32(0x3000, 0x1234ABCD)

	// EntryLo PFN is the physical page number; bits: V | G, and D for writable.
	pfn := uint64(0x3000 >> 12)
	c.COP0[cop0PageMask] = 0
	c.COP0[cop0EntryHi] = 0 // VPN2 = 0 covers vaddr 0x0000..0x1FFF
	c.COP0[cop0EntryLo0] = pfn<<6 | entryLoV | entryLoG | entryLoD
	c.COP0[cop0EntryLo1] = entryLoG // odd page invalid
	c.COP0[cop0Index] = 0
	c.tlbw(0)

	run(t, c, 1)
	if got, want := c.R[1], sext32(0x1234ABCD); got != want {
		t.Errorf("mapped load: got %016X want %016X", got, want)
	}

	// The odd page of the pair is invalid: a load there must fault, and since an
	// entry does match, it is not a refill — it takes the general vector.
	c2, _ := newTest(it(0x23, 5, 1, 0)) // lw $1, 0($5)
	c2.COP0[cop0EntryLo0] = pfn<<6 | entryLoV | entryLoG | entryLoD
	c2.COP0[cop0EntryLo1] = entryLoG
	c2.tlbw(0)
	c2.SetReg(5, 0x1000) // the odd page
	c2.Step()

	if c2.PC != 0xFFFFFFFF80000180 {
		t.Errorf("invalid-page fault vectored to %016X want the general handler", c2.PC)
	}
	if got := (c2.COP0[cop0Cause] >> 2) & 0x1F; got != excTLBL {
		t.Errorf("exception code %d want TLBL (%d)", got, excTLBL)
	}
	if c2.COP0[cop0BadVAddr] != 0x1000 {
		t.Errorf("BadVAddr = %016X want 0x1000", c2.COP0[cop0BadVAddr])
	}
}

func TestTLBRefillUsesItsOwnVector(t *testing.T) {
	// No entry matches at all: the refill vector at offset 0x000, so the handler
	// can be a fast path.
	c, _ := newTest(it(0x23, 5, 1, 0)) // lw $1, 0($5)
	c.SetReg(5, 0x40000000)
	c.Step()

	if c.PC != 0xFFFFFFFF80000000 {
		t.Errorf("refill vectored to %016X want 0x...80000000", c.PC)
	}
}

func TestFPUArithmeticAndCompare(t *testing.T) {
	c, _ := newTest(
		rt(0x11, 0x04, 1, 0, 0, 0),    // mtc1  $1, $f0
		rt(0x11, 0x04, 2, 1, 0, 0),    // mtc1  $2, $f1
		rt(0x11, 0x10, 1, 0, 2, 0x00), // add.s $f2, $f0, $f1
		rt(0x11, 0x10, 0, 2, 3, 0x24), // cvt.w.s $f3, $f2
		rt(0x11, 0x10, 1, 0, 0, 0x3C), // c.lt.s $f0, $f1
	)
	c.SetReg(1, 0x3FC00000) // 1.5f
	c.SetReg(2, 0x40200000) // 2.5f
	run(t, c, 5)

	if got := c.fs(2); got != 4.0 {
		t.Errorf("add.s: got %v want 4", got)
	}
	if got := c.readFGR32(3); got != 4 {
		t.Errorf("cvt.w.s: got %d want 4", got)
	}
	if c.FCR31&fcr31Cond == 0 {
		t.Error("c.lt.s 1.5 < 2.5 did not set the condition bit")
	}

	// And the false direction clears it again.
	c2, _ := newTest(
		rt(0x11, 0x04, 1, 0, 0, 0),
		rt(0x11, 0x04, 2, 1, 0, 0),
		rt(0x11, 0x10, 0, 1, 0, 0x3C), // c.lt.s $f1, $f0   (2.5 < 1.5 is false)
	)
	c2.FCR31 = fcr31Cond
	c2.SetReg(1, 0x3FC00000)
	c2.SetReg(2, 0x40200000)
	run(t, c2, 3)
	if c2.FCR31&fcr31Cond != 0 {
		t.Error("c.lt.s 2.5 < 1.5 left the condition bit set")
	}
}

func TestFPURegisterFileShape(t *testing.T) {
	// With FR clear a double occupies an even/odd pair of 32-bit registers: the
	// even register is the low half, the odd one the high half, and the double
	// is addressed by the even register of the pair.
	c, _ := newTest()
	c.COP0[cop0Status] &^= statusFR

	c.writeFGR32(0, 0x11111111) // low half of the $f0 pair
	c.writeFGR32(1, 0x22222222) // high half
	if got, want := c.readFGR64(0), uint64(0x2222222211111111); got != want {
		t.Errorf("paired double: got %016X want %016X", got, want)
	}
	// Naming the odd register of the pair reaches the same double.
	if got, want := c.readFGR64(1), uint64(0x2222222211111111); got != want {
		t.Errorf("odd-register double: got %016X want %016X", got, want)
	}

	// With FR set the 32 registers are independent: a 64-bit write to $f0 leaves
	// $f1 alone, where with FR clear it would have overwritten $f1's half.
	c2, _ := newTest()
	c2.COP0[cop0Status] |= statusFR
	c2.writeFGR64(0, 0xAAAAAAAABBBBBBBB)
	c2.writeFGR32(1, 0xCCCCCCCC)
	if got, want := c2.readFGR64(0), uint64(0xAAAAAAAABBBBBBBB); got != want {
		t.Errorf("FR=1: $f0 disturbed by a write to $f1: got %016X want %016X", got, want)
	}
	if got, want := c2.readFGR32(1), uint32(0xCCCCCCCC); got != want {
		t.Errorf("FR=1: $f1 got %08X want %08X", got, want)
	}
}

func TestUndefinedOpcodeRaisesReservedInstruction(t *testing.T) {
	// Every encoding the VR4300 does not define faults rather than halting. A
	// program may execute one deliberately to see what happens — n64-systemtest
	// does — and real hardware obliges it with an exception.
	for _, tc := range []struct {
		name string
		w    uint32
	}{
		{"COP3 (there is no third coprocessor)", 0x4C000000},
		{"SPECIAL3", 0x7C03E83B},
		{"an undefined SPECIAL function", 0x00000015},
		{"an undefined REGIMM function", 0x04170001},
	} {
		c, _ := newTest(tc.w)
		c.Step()
		if c.Halted {
			t.Errorf("%s halted the core instead of faulting: %s", tc.name, c.HaltReason)
			continue
		}
		if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excRI {
			t.Errorf("%s: exception code %d, want RI (%d)", tc.name, got, excRI)
		}
		if c.PC != 0xFFFFFFFF80000180 {
			t.Errorf("%s: vectored to %016X, want the general handler", tc.name, c.PC)
		}
	}
}

func TestCoprocessorUnusable(t *testing.T) {
	// Touching a coprocessor that Status has not enabled faults, and the Cause
	// register records which one — a handler cannot act without knowing.
	c, _ := newTest(rt(0x11, 0x04, 1, 0, 0, 0)) // mtc1, with CU1 cleared below
	c.COP0[cop0Status] &^= statusCU1
	c.Step()
	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excCpU {
		t.Errorf("COP1 while disabled: exception code %d, want CpU (%d)", got, excCpU)
	}
	if got := (c.COP0[cop0Cause] >> 28) & 3; got != 1 {
		t.Errorf("Cause CE = %d, want 1 (the coprocessor that faulted)", got)
	}
}

func TestCOP2IsALatch(t *testing.T) {
	// The VR4300's second coprocessor has no function unit — the N64's vector
	// unit lives in the RSP — but its moves work, on one shared latch.
	c, _ := newTest(
		rt(0x12, 0x05, 1, 0, 0, 0), // dmtc2 $1, $0
		rt(0x12, 0x01, 2, 0, 0, 0), // dmfc2 $2, $0
	)
	c.COP0[cop0Status] |= statusCU2
	c.SetReg(1, 0x0123456789ABCDEF)
	run(t, c, 2)
	if c.R[2] != 0x0123456789ABCDEF {
		t.Errorf("the COP2 latch did not round-trip: got %016X", c.R[2])
	}
}

func TestUnimplementedFPOperationFaults(t *testing.T) {
	// cvt.s.s names a conversion from a format to itself. The FPU has no unit for
	// it and traps unconditionally, recording the reason in FCR31.
	c, _ := newTest(0x46000120) // cvt.s.s $f4, $f0
	c.Step()
	if c.Halted {
		t.Fatalf("cvt.s.s halted the core: %s", c.HaltReason)
	}
	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excFPE {
		t.Errorf("exception code %d, want FPE (%d)", got, excFPE)
	}
	if c.FCR31&fcr31CauseUnimpl == 0 {
		t.Error("FCR31 does not record an unimplemented operation")
	}
}

func TestFPUCauseFieldIsClearedEachOperation(t *testing.T) {
	// The causes describe the operation that just ran. Letting them accumulate
	// reports an FPU permanently in every error state at once, which is what a
	// program polling FCR31 would then believe.
	c, _ := newTest(
		rt(0x11, 0x10, 1, 0, 2, 0x00), // add.s $f2, $f0, $f1  (inexact)
		rt(0x11, 0x10, 1, 0, 3, 0x00), // add.s $f3, $f0, $f1  (exact)
	)
	c.setS(0, 1.0)
	c.setS(1, 1e-30) // 1 + 1e-30 rounds back to 1: inexact
	run(t, c, 1)
	if c.FCR31&(fpInexact<<fcr31CauseShift) == 0 {
		t.Fatal("an inexact add did not raise the inexact cause")
	}

	c.setS(1, 1.0) // 1 + 1 is exact
	run(t, c, 1)
	if c.FCR31&fcr31CauseMask != 0 {
		t.Errorf("an exact add left causes set: FCR31 = %08X", c.FCR31)
	}
	// The sticky flag from the first operation survives, though.
	if c.FCR31&(fpInexact<<fcr31FlagShift) == 0 {
		t.Error("the sticky inexact flag was cleared: flags are not sticky")
	}
}

func TestFPUDivideByZeroAndInvalid(t *testing.T) {
	c, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x03)) // div.s $f2, $f0, $f1
	c.setS(0, 1.0)
	c.setS(1, 0.0)
	run(t, c, 1)
	if c.FCR31&(fpDivByZero<<fcr31CauseShift) == 0 {
		t.Error("1/0 did not raise divide-by-zero")
	}

	c2, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x03))
	c2.setS(0, 0.0)
	c2.setS(1, 0.0)
	run(t, c2, 1)
	if c2.FCR31&(fpInvalid<<fcr31CauseShift) == 0 {
		t.Error("0/0 did not raise invalid")
	}
}

func TestFPUEnabledCauseTrapsAndLeavesFlagsAlone(t *testing.T) {
	// When an enabled cause traps, the sticky flags are not updated: the handler
	// is meant to see the causes untouched.
	c, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x00)) // add.s
	c.FCR31 = fpInexact << fcr31EnableShift
	c.setS(0, 1.0)
	c.setS(1, 1e-30)
	c.Step()

	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excFPE {
		t.Fatalf("an enabled inexact cause did not trap: exception code %d", got)
	}
	if c.FCR31&(fpInexact<<fcr31FlagShift) != 0 {
		t.Error("a trapping operation updated the sticky flags")
	}
}

func TestAddressErrorUpdatesContext(t *testing.T) {
	// Context and XContext are page-table pointers: a base the OS wrote, with the
	// faulting page number spliced in. Every fault with an address updates them,
	// not just a TLB miss — and their low four bits are hardwired to zero.
	c, _ := newTest(it(0x23, 5, 1, 0)) // lw $1, 0($5), misaligned
	c.COP0[cop0Context] = 0xFFFFFFFFFF000000
	c.SetReg(5, 0x80002002)
	c.Step()

	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excAdEL {
		t.Fatalf("exception code %d, want AdEL", got)
	}
	if c.COP0[cop0BadVAddr] != 0x80002002 {
		t.Errorf("BadVAddr = %016X", c.COP0[cop0BadVAddr])
	}
	// BadVPN2 is the address shifted right 13, placed at bit 4.
	wantVPN := (uint64(0x80002002) >> 13) & 0x7FFFF << 4
	if got := c.COP0[cop0Context] & 0x7FFFFF; got != wantVPN {
		t.Errorf("Context page number = %06X want %06X", got, wantVPN)
	}
	if c.COP0[cop0Context]&0xF != 0 {
		t.Error("Context's low four bits are not zero")
	}
	if c.COP0[cop0Context]&0xFFFFFFFFFF000000 != 0xFFFFFFFFFF000000 {
		t.Error("Context's base was not preserved")
	}
}

func TestFPUHasNoDenormals(t *testing.T) {
	// A denormal operand is something this FPU cannot compute with at all: it
	// raises unimplemented-operation and leaves the arithmetic to software.
	c, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x00)) // add.s
	c.writeFGR32(0, 0x00000001)                    // the smallest denormal
	c.setS(1, 1.0)
	c.Step()
	if c.FCR31&fcr31CauseUnimpl == 0 {
		t.Error("a denormal operand did not raise unimplemented-operation")
	}
	if got := (c.COP0[cop0Cause] >> 2) & 0x1F; got != excFPE {
		t.Errorf("exception code %d, want FPE", got)
	}
}

func TestFPUDenormalResultFollowsTheFlushBit(t *testing.T) {
	// A result too small to represent flushes to a signed zero when FS is set,
	// and raises unimplemented-operation when it is not.
	sum := func(flush bool) *CPU {
		c, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x00)) // add.s $f2, $f0, $f1
		if flush {
			c.FCR31 |= fcr31FS
		}
		c.setS(0, 1.5285104e-37)
		c.setS(1, -1.5391543e-37) // the exact sum is denormal
		c.Step()
		return c
	}

	c := sum(true)
	if c.FCR31&fcr31CauseUnimpl != 0 {
		t.Error("with FS set, a denormal result should flush rather than trap")
	}
	if got := c.readFGR32(2); got != 0x80000000 {
		t.Errorf("flushed result = %08X, want a negative zero", got)
	}
	if c.FCR31&((fpUnderflow|fpInexact)<<fcr31CauseShift) != (fpUnderflow|fpInexact)<<fcr31CauseShift {
		t.Errorf("flushing did not report underflow and inexact: FCR31 = %08X", c.FCR31)
	}

	c = sum(false)
	if c.FCR31&fcr31CauseUnimpl == 0 {
		t.Error("with FS clear, a denormal result should raise unimplemented-operation")
	}
}

func TestFPURoundingModes(t *testing.T) {
	// 1/3 is not representable, so each mode lands on a different neighbour.
	div := func(rm uint32) float32 {
		c, _ := newTest(rt(0x11, 0x10, 1, 0, 2, 0x03)) // div.s $f2, $f0, $f1
		c.FCR31 = rm
		c.setS(0, 1)
		c.setS(1, 3)
		run(t, c, 1)
		return c.fs(2)
	}
	// The exact value lies just below the nearest float, so nearest rounds up.
	nearest := div(roundNearest)
	toZero := div(roundToZero)
	ceil := div(roundCeil)
	floor := div(roundFloor)

	if !(toZero < nearest) {
		t.Errorf("toward zero (%v) should step below nearest (%v)", toZero, nearest)
	}
	if ceil != nearest {
		t.Errorf("toward +inf (%v) should agree with nearest (%v) when nearest rounded up", ceil, nearest)
	}
	if floor != toZero {
		t.Errorf("toward -inf (%v) should agree with toward zero (%v) for a positive value", floor, toZero)
	}
}

func TestFPUProducesTheHardwareQuietNaN(t *testing.T) {
	// The VR4300's quiet NaN is not the one IEEE suggests, and neg does not flip
	// a NaN's sign.
	c, _ := newTest(
		rt(0x11, 0x10, 1, 0, 2, 0x03), // div.s $f2, $f0, $f1  -> 0/0 = NaN
		rt(0x11, 0x10, 0, 2, 3, 0x07), // neg.s $f3, $f2
	)
	c.setS(0, 0)
	c.setS(1, 0)
	run(t, c, 2)
	if got := c.readFGR32(2); got != qNaN32 {
		t.Errorf("0/0 gave %08X, want the hardware quiet NaN %08X", got, uint32(qNaN32))
	}
	if got := c.readFGR32(3); got != qNaN32 {
		t.Errorf("neg of a NaN gave %08X, want %08X unchanged", got, uint32(qNaN32))
	}
}

func TestSignallingNaNAndSourceFormats(t *testing.T) {
	// The VR4300 predates IEEE-754-2008 and uses the opposite NaN convention: a
	// mantissa whose top bit is set is *signalling*. So the NaN every modern
	// language calls quiet raises invalid-operation here.
	c, _ := newTest(rt(0x11, 0x10, 0, 0, 1, 0x05)) // abs.s $f1, $f0
	c.writeFGR32(0, 0x7FC00000)                    // Go's "quiet" NaN
	c.Step()
	if c.FCR31&(fpInvalid<<fcr31CauseShift) == 0 {
		t.Error("abs.s of a signalling NaN did not raise invalid")
	}
	if got := c.readFGR32(1); got != qNaN32 {
		t.Errorf("result %08X, want the hardware quiet NaN %08X", got, uint32(qNaN32))
	}

	// And a genuinely quiet NaN propagates in silence.
	c2, _ := newTest(rt(0x11, 0x10, 0, 0, 1, 0x05))
	c2.writeFGR32(0, qNaN32)
	c2.Step()
	if c2.FCR31&(fpInvalid<<fcr31CauseShift) != 0 {
		t.Error("abs.s of a quiet NaN raised invalid")
	}

	// cvt.s.d reads a *double*. Testing its source register as a single would
	// misread the low half of an ordinary value as a NaN pattern — a bug that
	// silently replaced most of a game's arithmetic.
	c3, _ := newTest(rt(0x11, 0x11, 0, 0, 2, 0x20)) // cvt.s.d $f2, $f0
	c3.setD(0, 1.5)                                 // low half is 0x00000000
	run(t, c3, 1)
	if got := c3.fs(2); got != 1.5 {
		t.Errorf("cvt.s.d gave %v, want 1.5", got)
	}
	if c3.FCR31&(fpInvalid<<fcr31CauseShift) != 0 {
		t.Error("cvt.s.d of an ordinary double raised invalid: it read its source as a single")
	}
}
