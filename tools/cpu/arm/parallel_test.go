package arm

import "testing"

// The ARMv6 parallel (SIMD) add/subtract group varies three things independently:
// lane width/pairing, signedness, and what happens to a lane that overflows —
// truncate (S/U, which publish GE), saturate (Q/UQ), or halve (SH/UH). These pin
// the modifiers, since a lane that merely truncates where it should saturate
// produces plausible-looking but wrong pixels/audio rather than an obvious crash.
//
// Encoding: cond=E, bits 27:25=011, bits 22:20 = class, bits 7:5 = op2, bit 4 = 1.
//
//	0xE6_C_D_F_O_M  where C = 2x|class, D = Rd, Rn at 19:16, Rm at 3:0.
func parWord(class, op2, rd, rn, rm uint32) uint32 {
	return 0xE6000000 | class<<20 | rn<<16 | rd<<12 | 0xF<<8 | op2<<5 | 1<<4 | rm
}

func runPar(t *testing.T, class, op2, rnv, rmv uint32) (uint32, uint32) {
	t.Helper()
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS
	c.R[2] = rnv
	c.R[3] = rmv
	b.loadARM(0x1000, parWord(class, op2, 1, 2, 3))
	c.R[15] = 0x1000
	c.Step()
	if c.Halted {
		t.Fatalf("halted: %s", c.HaltReason)
	}
	return c.R[1], c.GE
}

const (
	clsS  = 1
	clsQ  = 2
	clsSH = 3
	clsU  = 5
	clsUQ = 6
	clsUH = 7

	opADD16 = 0b000
	opASX   = 0b001
	opSAX   = 0b010
	opSUB16 = 0b011
	opADD8  = 0b100
	opSUB8  = 0b111
)

// UQSUB8 is the instruction Captain Toad's boot halted on: each unsigned byte
// lane clamps at 0 instead of wrapping to 0xFF.
func TestUQSUB8Saturates(t *testing.T) {
	// lanes: 0x10-0x20 -> 0 (clamped), 0x80-0x01 = 0x7F, 0x00-0xFF -> 0, 0xFF-0x0F = 0xF0
	got, _ := runPar(t, clsUQ, opSUB8, 0xFF008010, 0x0FFF0120)
	want := uint32(0xF0007F00)
	if got != want {
		t.Fatalf("UQSUB8 = 0x%08X, want 0x%08X", got, want)
	}
	// The plain (wrapping) U form must NOT clamp — this is what makes the
	// saturating variant worth distinguishing.
	got, _ = runPar(t, clsU, opSUB8, 0xFF008010, 0x0FFF0120)
	if got == want {
		t.Fatalf("USUB8 saturated like UQSUB8 (0x%08X); it must wrap", got)
	}
}

func TestUQADD8Saturates(t *testing.T) {
	// 0xF0+0x20 -> 0xFF (clamped), 0x01+0x02 = 0x03
	got, _ := runPar(t, clsUQ, opADD8, 0x000000F0|0x01<<8, 0x00000020|0x02<<8)
	if lane0 := got & 0xFF; lane0 != 0xFF {
		t.Fatalf("UQADD8 lane0 = 0x%02X, want 0xFF (saturated)", lane0)
	}
	if lane1 := (got >> 8) & 0xFF; lane1 != 0x03 {
		t.Fatalf("UQADD8 lane1 = 0x%02X, want 0x03", lane1)
	}
}

func TestQSUB16SignedSaturates(t *testing.T) {
	// signed halfwords: 0x8000 (-32768) - 0x0001 -> clamps to 0x8000 (-32768)
	got, _ := runPar(t, clsQ, opSUB16, 0x00008000, 0x00000001)
	if lane0 := got & 0xFFFF; lane0 != 0x8000 {
		t.Fatalf("QSUB16 lane0 = 0x%04X, want 0x8000 (negative saturation)", lane0)
	}
}

func TestHalvingForms(t *testing.T) {
	// UHADD8: (0xFF + 0xFF) >> 1 = 0xFF — the halving forms cannot overflow.
	got, _ := runPar(t, clsUH, opADD8, 0x000000FF, 0x000000FF)
	if lane0 := got & 0xFF; lane0 != 0xFF {
		t.Fatalf("UHADD8 lane0 = 0x%02X, want 0xFF", lane0)
	}
	// SHADD16: signed (0x7FFF + 0x7FFF) >> 1 = 0x7FFF
	got, _ = runPar(t, clsSH, opADD16, 0x00007FFF, 0x00007FFF)
	if lane0 := got & 0xFFFF; lane0 != 0x7FFF {
		t.Fatalf("SHADD16 lane0 = 0x%04X, want 0x7FFF", lane0)
	}
}

// The exchange forms cross Rm's halfwords: ASX subtracts Rm's HIGH half in the
// low lane and adds Rm's low half into the high lane; SAX is the mirror.
func TestExchangeForms(t *testing.T) {
	// UASX: lo = Rn.lo - Rm.hi = 0x0030 - 0x0010 = 0x0020
	//       hi = Rn.hi + Rm.lo = 0x0005 + 0x0002 = 0x0007
	got, _ := runPar(t, clsU, opASX, 0x00050030, 0x00100002)
	if got != 0x00070020 {
		t.Fatalf("UASX = 0x%08X, want 0x00070020", got)
	}
	// USAX: lo = Rn.lo + Rm.hi = 0x0030 + 0x0010 = 0x0040
	//       hi = Rn.hi - Rm.lo = 0x0005 - 0x0002 = 0x0003
	got, _ = runPar(t, clsU, opSAX, 0x00050030, 0x00100002)
	if got != 0x00030040 {
		t.Fatalf("USAX = 0x%08X, want 0x00030040", got)
	}
}

// Only the truncating classes publish GE (a following SEL consumes it); the
// saturating and halving classes must leave it untouched.
func TestGEOnlyFromTruncatingClasses(t *testing.T) {
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS
	c.GE = 0xA // a value that must survive a UQ op
	c.R[2] = 0x00000010
	c.R[3] = 0x00000020
	b.loadARM(0x1000, parWord(clsUQ, opSUB8, 1, 2, 3))
	c.R[15] = 0x1000
	c.Step()
	if c.GE != 0xA {
		t.Fatalf("UQSUB8 clobbered GE (0x%X); saturating forms must not write it", c.GE)
	}

	// USUB8 (truncating) does set GE: lane0 0x10-0x20 borrows, so GE bit 0 clears.
	c.GE = 0xF
	b.loadARM(0x1000, parWord(clsU, opSUB8, 1, 2, 3))
	c.R[15] = 0x1000
	c.Step()
	if c.GE&1 != 0 {
		t.Fatalf("USUB8 GE bit0 = 1, want 0 (0x10 - 0x20 borrows)")
	}
}
