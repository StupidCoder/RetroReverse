package arm

import "testing"

// LDREXD/STREXD are the 64-bit members of the load/store-exclusive family: the
// value is a register *pair*, Rt and Rt+1, low word first, and Rt2 is implicit —
// it is not encoded anywhere in the instruction. The exclusive monitor tags the
// base address exactly as the word-sized forms do.
//
// Encodings (ARMv6K, always-condition), hand-assembled from bits 27:24 = 0001,
// 7:4 = 1001, bit 23 = 1 (exclusive, not SWP), bits 22:21 = 01 (doubleword):
//
//	LDREXD r0, r1, [r2] = 0xE1B20F9F  (L=1, Rn=2, Rt=0)
//	STREXD r3, r4, r5, [r2] = 0xE1A23F94  (L=0, Rn=2, Rd=3, Rt=4)
func TestLDREXDSTREXD(t *testing.T) {
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS

	c.R[2] = 0x2000
	b.loadARM(0x2000, 0x11223344, 0x55667788)

	b.loadARM(0x1000, 0xE1B20F9F) // LDREXD r0, r1, [r2]
	c.R[15] = 0x1000
	c.Step()
	if c.R[0] != 0x11223344 || c.R[1] != 0x55667788 {
		t.Fatalf("LDREXD gave r0=0x%08X r1=0x%08X, want 0x11223344 0x55667788", c.R[0], c.R[1])
	}
	if !c.exclValid || c.exclAddr != 0x2000 {
		t.Fatalf("LDREXD did not tag the address: valid=%v addr=0x%08X", c.exclValid, c.exclAddr)
	}

	// The tagged STREXD stores BOTH words and reports success (0).
	c.R[4] = 0xDEADBEEF
	c.R[5] = 0xFEEDFACE
	b.loadARM(0x1000, 0xE1A23F94) // STREXD r3, r4, r5, [r2]
	c.R[15] = 0x1000
	c.Step()
	if c.R[3] != 0 {
		t.Fatalf("tagged STREXD reported failure (r3=%d), want success (0)", c.R[3])
	}
	if got := b.word(0x2000); got != 0xDEADBEEF {
		t.Fatalf("STREXD low word = 0x%08X, want 0xDEADBEEF", got)
	}
	if got := b.word(0x2004); got != 0xFEEDFACE {
		t.Fatalf("STREXD high word = 0x%08X, want 0xFEEDFACE", got)
	}
}

// Without a matching LDREXD the store must fail, report 1, and leave memory
// untouched — the property the lock code's retry loop depends on.
func TestSTREXDFailsUntagged(t *testing.T) {
	b := newBus(0x4000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS

	c.R[2] = 0x2000
	b.loadARM(0x2000, 0xAAAAAAAA, 0xBBBBBBBB)
	c.R[4] = 0xDEADBEEF
	c.R[5] = 0xFEEDFACE

	c.ClearExclusive()
	b.loadARM(0x1000, 0xE1A23F94) // STREXD r3, r4, r5, [r2]
	c.R[15] = 0x1000
	c.Step()
	if c.R[3] != 1 {
		t.Fatalf("untagged STREXD reported success (r3=%d), want failure (1)", c.R[3])
	}
	if b.word(0x2000) != 0xAAAAAAAA || b.word(0x2004) != 0xBBBBBBBB {
		t.Fatalf("failed STREXD wrote memory: 0x%08X 0x%08X", b.word(0x2000), b.word(0x2004))
	}
}
