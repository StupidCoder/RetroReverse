package rsp

// Assembled-loop tests. There is no published per-instruction conformance suite
// for the RSP, so these concentrate on what makes it different from an ordinary
// MIPS core — the wrapping program counter, the wrapping data memory, and above
// all the vector unit's accumulator, whose 48-bit width and per-instruction
// choice of output slice is where a wrong answer hides most easily.

import "testing"

// testRegs is a COP0 window that just remembers what was written.
type testRegs struct{ r [16]uint32 }

func (t *testRegs) ReadCop0(i uint32) uint32     { return t.r[i&15] }
func (t *testRegs) WriteCop0(i uint32, v uint32) { t.r[i&15] = v }

func newTest(words ...uint32) (*CPU, *testRegs) {
	regs := &testRegs{}
	dmem := make([]byte, DMEMSize)
	imem := make([]byte, IMEMSize)
	for i, w := range words {
		imem[i*4+0] = byte(w >> 24)
		imem[i*4+1] = byte(w >> 16)
		imem[i*4+2] = byte(w >> 8)
		imem[i*4+3] = byte(w)
	}
	c := NewCPU(dmem, imem, regs)
	c.Start(0)
	return c, regs
}

func rt(op, rs, rtr, rd, sh, funct uint32) uint32 {
	return op<<26 | rs<<21 | rtr<<16 | rd<<11 | sh<<6 | funct
}
func it(op, rs, rtr uint32, imm uint16) uint32 { return op<<26 | rs<<21 | rtr<<16 | uint32(imm) }

// vec assembles a COP2 vector ALU instruction.
func vec(e, vt, vs, vd, funct uint32) uint32 {
	return 0x12<<26 | 1<<25 | e<<21 | vt<<16 | vs<<11 | vd<<6 | funct
}

const brk = 0x0000000D

func run(t *testing.T, c *CPU, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		c.Step()
		if c.Halted && !c.Broke {
			t.Fatalf("halted after %d steps: %s", i+1, c.HaltReason)
		}
	}
}

func TestScalarArithmeticDoesNotTrap(t *testing.T) {
	// The RSP has no arithmetic exceptions, so `add` on an overflowing pair is
	// just `addu`. A core that traps here stops a microcode dead.
	c, _ := newTest(rt(0, 1, 2, 3, 0, 0x20)) // add $3, $1, $2
	c.R[1] = 0x7FFFFFFF
	c.R[2] = 1
	run(t, c, 1)
	if c.R[3] != 0x80000000 {
		t.Errorf("add: got %08X want 80000000", c.R[3])
	}
}

func TestBreakStopsTheCore(t *testing.T) {
	c, _ := newTest(brk, it(0x09, 0, 1, 0xBAD)) // break; addiu $1,$0,0xBAD
	c.Step()
	if !c.Broke || !c.Halted {
		t.Fatalf("break did not stop the core: broke=%v halted=%v", c.Broke, c.Halted)
	}
	c.Step() // a halted core does nothing
	if c.R[1] != 0 {
		t.Errorf("the core executed past a break: $1 = %08X", c.R[1])
	}
}

func TestProgramCounterWrapsInIMEM(t *testing.T) {
	// A jump past the end of instruction memory lands back at the start: the PC
	// is 12 bits and there is nowhere else to go.
	c, _ := newTest()
	c.SetPC(0xFFC)
	// At 0xFFC: j 0x800 -> the delay slot is at 0x000, which wraps.
	w := uint32(0x02)<<26 | (0x800 >> 2)
	c.IMEM[0xFFC] = byte(w >> 24)
	c.IMEM[0xFFD] = byte(w >> 16)
	c.IMEM[0xFFE] = byte(w >> 8)
	c.IMEM[0xFFF] = byte(w)
	c.Step() // the jump
	if c.PC != 0x000 {
		t.Errorf("delay slot after a jump at 0xFFC: PC = %03X want 000", c.PC)
	}
	c.Step() // the delay slot
	if c.PC != 0x800 {
		t.Errorf("jump target: PC = %03X want 800", c.PC)
	}
}

func TestDataMemoryWraps(t *testing.T) {
	// DMEM is 4 KiB and every access wraps inside it; there is no fault.
	c, _ := newTest(it(0x2B, 1, 2, 0), it(0x23, 0, 3, 0)) // sw $2,0($1); lw $3,0($0)
	c.R[1] = 0x1000                                       // wraps to 0
	c.R[2] = 0xDEADBEEF
	run(t, c, 2)
	if c.R[3] != 0xDEADBEEF {
		t.Errorf("wrapped store/load: got %08X want DEADBEEF", c.R[3])
	}
}

func TestCop0ReachesTheRegisterWindow(t *testing.T) {
	// The RSP's COP0 is not a system-control unit: it is the SP and DP registers.
	c, regs := newTest(rt(0x10, 0x04, 2, 4, 0, 0), rt(0x10, 0x00, 3, 4, 0, 0))
	c.R[2] = 0x1234
	run(t, c, 2)
	if regs.r[4] != 0x1234 {
		t.Errorf("mtc0 to SP_STATUS: got %08X want 1234", regs.r[4])
	}
	if c.R[3] != 0x1234 {
		t.Errorf("mfc0 from SP_STATUS: got %08X want 1234", c.R[3])
	}
}

// --- the vector unit --------------------------------------------------------

func TestElementBroadcast(t *testing.T) {
	// Elements 8..15 broadcast one lane across all eight; 4..7 select within a
	// quad; 2..3 within a pair. This is how microcode scales a vector by a scalar.
	for _, tc := range []struct{ e, i, want uint32 }{
		{0, 5, 5}, {1, 5, 5},
		{2, 5, 4}, {3, 5, 5},
		{4, 5, 4}, {7, 5, 7},
		{8, 5, 0}, {13, 5, 5}, {15, 5, 7},
	} {
		if got := element(tc.e, tc.i); got != tc.want {
			t.Errorf("element(e=%d, lane=%d) = %d want %d", tc.e, tc.i, got, tc.want)
		}
	}
}

func TestVMUDNAccumulatesIntoVMADN(t *testing.T) {
	// The multiply pair that microcode uses for a 32-bit product: vmudn writes
	// the accumulator, vmadn adds into it, and both return the low slice.
	c, _ := newTest(
		vec(0, 2, 1, 3, 0x06), // vmudn $v3, $v1, $v2
		vec(0, 2, 1, 4, 0x0E), // vmadn $v4, $v1, $v2
	)
	c.V[1][0] = 0x0002
	c.V[2][0] = 0x0003
	run(t, c, 2)

	if c.V[3][0] != 6 {
		t.Errorf("vmudn lane 0: got %04X want 0006", c.V[3][0])
	}
	// The accumulator now holds 6 + 6 = 12.
	if c.V[4][0] != 12 {
		t.Errorf("vmadn lane 0: got %04X want 000C", c.V[4][0])
	}
	if c.Acc[0] != 12 {
		t.Errorf("accumulator lane 0: got %012X want 00000000000C", c.Acc[0])
	}
}

func TestVMUDHWritesTheHighSlice(t *testing.T) {
	// vmudh puts the product in the accumulator's high half, so the result is
	// the product itself, saturated.
	c, _ := newTest(vec(0, 2, 1, 3, 0x07))
	c.V[1][0] = 0x0100 // 256
	c.V[2][0] = 0x0100
	c.V[1][1] = 0x7FFF // a product that overflows a signed 16-bit result
	c.V[2][1] = 0x7FFF
	run(t, c, 1)

	// 256*256 is 0x10000, one past a signed 16-bit result, so it saturates —
	// while the accumulator keeps the exact product in its high half.
	if c.V[3][0] != 0x7FFF {
		t.Errorf("vmudh 256*256 must saturate: got %04X want 7FFF", c.V[3][0])
	}
	if c.Acc[0]>>16 != 0x10000 {
		t.Errorf("vmudh accumulator lane 0: got %012X want the exact product 0x10000", c.Acc[0]>>16)
	}
	if c.V[3][1] != 0x7FFF {
		t.Errorf("vmudh saturation: lane 1 got %04X want 7FFF", c.V[3][1])
	}
}

func TestVMULFRoundsAndSaturates(t *testing.T) {
	// vmulf is the signed fractional multiply: (a*b*2 + 0x8000) >> 16.
	c, _ := newTest(vec(0, 2, 1, 3, 0x00))
	c.V[1][0] = 0x4000 // 0.5
	c.V[2][0] = 0x4000 // 0.5
	c.V[1][1] = 0x8000 // -1.0, whose square saturates
	c.V[2][1] = 0x8000
	run(t, c, 1)

	if c.V[3][0] != 0x2000 { // 0.25
		t.Errorf("vmulf 0.5*0.5: got %04X want 2000", c.V[3][0])
	}
	if c.V[3][1] != 0x7FFF {
		t.Errorf("vmulf -1*-1 must saturate: got %04X want 7FFF", c.V[3][1])
	}
}

func TestVADDCarriesThroughVCO(t *testing.T) {
	// vaddc leaves a carry in VCO that the following vadd consumes; that pair is
	// how microcode adds 32-bit quantities held as two vectors.
	c, _ := newTest(
		vec(0, 2, 1, 3, 0x14), // vaddc $v3, $v1, $v2  (low halves)
		vec(0, 4, 5, 6, 0x10), // vadd  $v6, $v5, $v4  (high halves + carry)
	)
	c.V[1][0] = 0xFFFF
	c.V[2][0] = 0x0001 // 0xFFFF + 1 carries
	c.V[5][0] = 0x0000
	c.V[4][0] = 0x0000
	run(t, c, 2)

	if c.V[3][0] != 0x0000 {
		t.Errorf("vaddc low: got %04X want 0000", c.V[3][0])
	}
	if c.V[6][0] != 0x0001 {
		t.Errorf("vadd did not consume the carry: got %04X want 0001", c.V[6][0])
	}
	if c.VCO != 0 {
		t.Errorf("vadd must clear VCO: got %04X", c.VCO)
	}
}

func TestVSARReadsAccumulatorSlices(t *testing.T) {
	c, _ := newTest(
		vec(8, 0, 0, 1, 0x1D),  // vsar $v1, $v0, e=8  (high)
		vec(9, 0, 0, 2, 0x1D),  // vsar $v2, $v0, e=9  (mid)
		vec(10, 0, 0, 3, 0x1D), // vsar $v3, $v0, e=10 (low)
	)
	c.Acc[0] = 0xAAAABBBBCCCC
	run(t, c, 3)
	if c.V[1][0] != 0xAAAA || c.V[2][0] != 0xBBBB || c.V[3][0] != 0xCCCC {
		t.Errorf("vsar slices: %04X %04X %04X want AAAA BBBB CCCC", c.V[1][0], c.V[2][0], c.V[3][0])
	}
}

func TestVectorLoadStoreQuadword(t *testing.T) {
	// lqv loads up to the next 16-byte boundary, which is what makes it safe to
	// use on a pointer that is not itself aligned.
	c, _ := newTest(
		0x32<<26|1<<21|1<<16|4<<11|0, // lqv $v1[0], 0($1)
		0x3A<<26|2<<21|1<<16|4<<11|0, // sqv $v1[0], 0($2)
	)
	for i := 0; i < 16; i++ {
		c.DMEM[i] = byte(0x10 + i)
	}
	c.R[1] = 0
	c.R[2] = 0x100
	run(t, c, 2)

	if c.V[1][0] != 0x1011 || c.V[1][7] != 0x1E1F {
		t.Errorf("lqv: lane0=%04X lane7=%04X want 1011 1E1F", c.V[1][0], c.V[1][7])
	}
	for i := 0; i < 16; i++ {
		if c.DMEM[0x100+i] != byte(0x10+i) {
			t.Fatalf("sqv byte %d: got %02X want %02X", i, c.DMEM[0x100+i], 0x10+i)
		}
	}
}

func TestReciprocalScalesConsistently(t *testing.T) {
	// The hardware's reciprocal is scaled: it approximates 2^33/x rather than
	// 1/x, and microcode carries the exponent itself. What must hold is that
	// doubling the input exactly halves the result, which is what makes the
	// scaling usable at all.
	c, _ := newTest()
	one := c.reciprocal(0x00010000)  // 1.0 in 16.16
	two := c.reciprocal(0x00020000)  // 2.0
	four := c.reciprocal(0x00040000) // 4.0

	if two != one>>1 || four != one>>2 {
		t.Errorf("reciprocal is not scale-consistent: rcp(1)=%08X rcp(2)=%08X rcp(4)=%08X",
			one, two, four)
	}
	// The leading 1 and the first ROM entry: 1 || 0xFFFF.
	if one != 0x0001FFFF {
		t.Errorf("rcp(1.0) = %08X want 0001FFFF", one)
	}
}

func TestReciprocalOfZeroAndNegatives(t *testing.T) {
	c, _ := newTest()
	if got := c.reciprocal(0); got != 0xFFFF || c.divOut != 0xFFFF {
		t.Errorf("rcp(0) = %04X/%04X want FFFF/FFFF (it saturates; there is no exception)",
			c.divOut, got)
	}
	// A negative input inverts the result rather than negating it.
	pos := c.reciprocal(0x00010000)
	neg := c.reciprocal(-0x00010000)
	if neg != ^pos {
		t.Errorf("rcp(-x) = %08X want ^rcp(x) = %08X", neg, ^pos)
	}
}

func TestReciprocalROMResistsAClosedForm(t *testing.T) {
	// The obvious closed form, floor(65536*(512-i)/(512+i)), reproduces 510 of
	// the ROM's 512 entries. It is wrong at exactly two — indices 241 and 273,
	// each low by one.
	//
	// That is the whole argument for transcribing the table rather than
	// computing it. A reciprocal that is correct 99.6% of the time yields
	// geometry that is almost right, everywhere, occasionally: the hardest
	// possible thing to notice and to chase.
	var wrong []int
	for i := 0; i < 512; i++ {
		v := 65536 * (512 - i) / (512 + i)
		if v > 0xFFFF {
			v = 0xFFFF
		}
		if uint16(v) != rcpROM[i] {
			wrong = append(wrong, i)
		}
	}
	if len(wrong) != 2 || wrong[0] != 241 || wrong[1] != 273 {
		t.Errorf("the closed form diverges at %v; expected exactly [241 273]", wrong)
	}
	if rcpROM[241] != 23586 || rcpROM[273] != 19953 {
		t.Errorf("ROM entries 241/273 = %d/%d, want 23586/19953", rcpROM[241], rcpROM[273])
	}
}

func TestLongReciprocalSequence(t *testing.T) {
	// A 32-bit divide takes three instructions, and this is the sequence at the
	// top of the game's perspective-divide overlay: VRCPH latches the input's
	// high half, VRCPL supplies the low half and does the work, and a second
	// VRCPH reads the answer's high half back out.
	//
	// vrcph $v2[0], $v1[0]   -- latch the high half of the input
	// vrcpl $v3[0], $v1[1]   -- supply the low half; $v3 gets the low result
	// vrcph $v4[0], $v0[0]   -- $v4 gets the high result
	c, _ := newTest(
		vec(0, 1, 0, 2, 0x32),
		vec(1, 1, 0, 3, 0x31),
		vec(0, 0, 0, 4, 0x32),
	)
	// The input 0x00010000 spread across two lanes: high 0x0001, low 0x0000.
	c.V[1][0] = 0x0001
	c.V[1][1] = 0x0000

	run(t, c, 1)
	if !c.divInLoaded {
		t.Fatal("vrcph did not latch the input's high half")
	}
	run(t, c, 1)
	if c.divInLoaded {
		t.Error("vrcpl did not consume the latched high half")
	}
	run(t, c, 1) // the second vrcph collects the result's high half

	want := c.reciprocal(0x00010000)
	got := uint32(c.V[4][0])<<16 | uint32(c.V[3][0])
	if got != want {
		t.Errorf("32-bit reciprocal: got %08X want %08X", got, want)
	}
}

func TestUnimplementedEncodingsAreRecordedNotFatal(t *testing.T) {
	// The RSP has no exception mechanism: no interrupts, no faults, nowhere to
	// vector to. An encoding it does not implement does nothing at all, so a core
	// that halted would stop microcode real hardware runs. The gap is recorded
	// instead.
	//
	// VRSQ is one: its published algorithm is self-inconsistent, so rather than
	// invent a plausible answer the core leaves the destination alone and says so.
	w := vec(0, 1, 0, 2, 0x34) // vrsq
	c, _ := newTest(w)
	c.V[2][0] = 0xBEEF
	c.Step()

	if c.Halted {
		t.Fatalf("vrsq halted the core: %s", c.HaltReason)
	}
	if c.V[2][0] != 0xBEEF {
		t.Error("vrsq wrote a destination lane it does not model")
	}
	if c.Unimplemented[w] != 1 {
		t.Errorf("vrsq was not recorded as unimplemented: %v", c.Unimplemented)
	}
}

func TestOnlyMultipliesOwnTheWholeAccumulator(t *testing.T) {
	// A multiply produces a 48-bit product and owns its accumulator lane. An add,
	// a compare, a logical or a lane move produces sixteen bits and writes only
	// those — the slices above survive. Microcode reads them back with VSAR long
	// after the operation that filled them, so an implementation that zeroes them
	// on every VADD quietly destroys geometry.
	c, _ := newTest(
		vec(0, 2, 1, 3, 0x07), // vmudh $v3, $v1, $v2  -- fills the whole lane
		vec(0, 5, 4, 6, 0x10), // vadd  $v6, $v4, $v5  -- must touch only the low 16
	)
	c.V[1][0] = 0x0100
	c.V[2][0] = 0x0100 // product 0x10000, so acc lane 0 = 0x0001_0000_0000
	run(t, c, 1)
	before := c.Acc[0]
	if before>>16 == 0 {
		t.Fatalf("vmudh did not fill the accumulator's upper slices: %012X", before)
	}

	c.V[4][0] = 3
	c.V[5][0] = 4
	run(t, c, 1)

	if got := c.Acc[0] & 0xFFFF; got != 7 {
		t.Errorf("vadd wrote acc low = %04X, want 7", got)
	}
	if got, want := c.Acc[0]>>16, before>>16; got != want {
		t.Errorf("vadd disturbed the accumulator above bit 16: %08X want %08X", got, want)
	}
}

func TestLRVLoadsTheBytesBeforeTheAddress(t *testing.T) {
	// LQV and LRV are the LWL and LWR of a 128-bit load. LQV takes the bytes from
	// the address up to the next 16-byte boundary; LRV takes those from the
	// previous boundary up to, but excluding, the address, and lands them at the
	// register's tail.
	c, _ := newTest(
		0x32<<26|1<<21|1<<16|4<<11|0, // lqv $v1[0], 0($1)
		0x32<<26|2<<21|1<<16|5<<11|0, // lrv $v1[0], 0($2)
	)
	for i := 0; i < 32; i++ {
		c.DMEM[i] = byte(0x10 + i)
	}
	c.R[1] = 8  // eight bytes, up to the boundary at 16
	c.R[2] = 24 // the eight bytes from 16 up to 24
	run(t, c, 2)

	// Bytes 8..15 of DMEM land in the register's first half, 16..23 in its second.
	if got := c.vecByte(1, 0); got != 0x18 {
		t.Errorf("lqv put %02X in byte 0, want 18", got)
	}
	if got := c.vecByte(1, 8); got != 0x20 {
		t.Errorf("lrv put %02X in byte 8, want 20", got)
	}
	if got := c.vecByte(1, 15); got != 0x27 {
		t.Errorf("lrv put %02X in byte 15, want 27", got)
	}
}
