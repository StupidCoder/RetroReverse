package arm

import "testing"

// LDRD/STRD move a register *pair* (Rd, Rd+1) to/from two consecutive words.
// They live in the "extra load/store" space where the L bit does not separate
// load from store — LDRD is L=0,sh=2 (a load) and STRD is L=0,sh=3 — so a decoder
// keyed on L alone runs them as a 16-bit STRH and corrupts memory. These check the
// real doubleword behaviour.
//
// Encodings (immediate, offset 0, P=1,U=1,W=0), hand-assembled:
//   LDRD r0, [r2]  = 0xE1C200D0  (bits27:20=0x1C, bits7:4=1101 → sh=2)
//   STRD r4, [r2]  = 0xE1C240F0  (bits7:4=1111 → sh=3)
func TestLDRDSTRD(t *testing.T) {
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Mode = ModeSYS

	// Data lives at 0x2000, clear of the instruction word at 0x1000.
	c.R[2] = 0x2000
	b.loadARM(0x2000, 0x11223344, 0x55667788)

	b.loadARM(0x1000, 0xE1C200D0) // LDRD r0, [r2]
	c.R[15] = 0x1000
	c.Step()
	if c.R[0] != 0x11223344 || c.R[1] != 0x55667788 {
		t.Fatalf("LDRD gave r0=0x%08X r1=0x%08X, want 0x11223344 0x55667788", c.R[0], c.R[1])
	}

	// STRD r4,r5 to 0x2008 and read the pair back.
	c.R[4] = 0xDEADBEEF
	c.R[5] = 0xFEEDFACE
	c.R[2] = 0x2008
	b.loadARM(0x1000, 0xE1C240F0) // STRD r4, [r2]
	c.R[15] = 0x1000
	c.Step()
	if got := b.word(0x2008); got != 0xDEADBEEF {
		t.Fatalf("STRD low word = 0x%08X, want 0xDEADBEEF", got)
	}
	if got := b.word(0x200C); got != 0xFEEDFACE {
		t.Fatalf("STRD high word = 0x%08X, want 0xFEEDFACE", got)
	}
}

// A regression pin for the specific corruption bug: STRD must NOT behave like a
// 16-bit STRH (which would write only the low halfword of Rd to [addr] and leave
// the rest untouched).
func TestSTRDIsNotSTRH(t *testing.T) {
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[2] = 0x2000
	c.R[0] = 0xAABBCCDD
	c.R[1] = 0x01020304
	b.loadARM(0x2000, 0xFFFFFFFF, 0xFFFFFFFF)
	b.loadARM(0x1000, 0xE1C200F0) // STRD r0, [r2]
	c.R[15] = 0x1000
	c.Step()
	if b.word(0x2000) != 0xAABBCCDD || b.word(0x2004) != 0x01020304 {
		t.Fatalf("STRD wrote 0x%08X 0x%08X (STRH bug leaves high word as 0xFFFFFFFF)", b.word(0x2000), b.word(0x2004))
	}
}
