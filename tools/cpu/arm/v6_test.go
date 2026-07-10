package arm

import "testing"

// The encodings below are hand-assembled from the ARM Architecture Reference
// Manual (ARMv6K), so the decoder and the execution core are checked against a
// source independent of themselves.

func TestV6Decode(t *testing.T) {
	cases := []struct {
		w    uint32
		mnem string
		text string
	}{
		{0xE6BF0F31, "REV", "REV r0, r1"},
		{0xE6BF0FB1, "REV16", "REV16 r0, r1"},
		{0xE6FF0FB1, "REVSH", "REVSH r0, r1"},
		{0xE6EF0071, "UXTB", "UXTB r0, r1"},
		{0xE6BF0071, "SXTH", "SXTH r0, r1"},
		{0xE1910F9F, "LDREX", "LDREX r0, [r1]"},
		{0xE1810F92, "STREX", "STREX r0, r2, [r1]"},
		{0xE0410293, "UMAAL", "UMAAL r0, r1, r3, r2"},
		{0xE6810FB2, "SEL", "SEL r0, r1, r2"},
		{0xE6510F92, "UADD8", "UADD8 r0, r1, r2"},
		{0xE750F211, "SMMUL", "SMMUL r0, r1, r2"},
		{0xE700F211, "SMUAD", "SMUAD r0, r1, r2"},
		{0xE6AF0011, "SSAT", "SSAT r0, #16, r1"},
		{0xE6E70011, "USAT", "USAT r0, #7, r1"},
		{0xE6810012, "PKHBT", "PKHBT r0, r1, r2, LSL #0"},
	}
	for _, tc := range cases {
		code := []byte{byte(tc.w), byte(tc.w >> 8), byte(tc.w >> 16), byte(tc.w >> 24)}
		in := DecodeVariant(code, 0x1000, false, V6K)
		if in.Mnem != tc.mnem {
			t.Errorf("0x%08X: mnem = %q, want %q", tc.w, in.Mnem, tc.mnem)
		}
		if in.Text != tc.text {
			t.Errorf("0x%08X: text = %q, want %q", tc.w, in.Text, tc.text)
		}
	}
}

// A V6K instruction fed to the V5TE decoder must NOT silently decode as its
// aliased base instruction. LDREX shares SWP's slot, so the V5TE decoder reads
// it as SWP — the very confusion the variant split exists to prevent.
func TestV6AliasingUnderV5(t *testing.T) {
	ldrex := uint32(0xE1910F9F)
	code := []byte{byte(ldrex), byte(ldrex >> 8), byte(ldrex >> 16), byte(ldrex >> 24)}
	v5 := DecodeVariant(code, 0x1000, false, V5TE)
	v6 := DecodeVariant(code, 0x1000, false, V6K)
	if v5.Mnem == v6.Mnem {
		t.Fatalf("V5TE and V6K decode the LDREX encoding identically (%q); the alias is not being intercepted", v5.Mnem)
	}
	if v6.Mnem != "LDREX" {
		t.Errorf("V6K LDREX decoded as %q", v6.Mnem)
	}
}

func v6cpu() (*CPU, *flatBus) {
	b := newBus(0x10000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS
	return c, b
}

// stepOne executes exactly the instruction at 0x1000.
func stepOne(c *CPU, b *flatBus, w uint32) {
	b.loadARM(0x1000, w)
	c.R[15] = 0x1000
	c.Step()
}

func TestV6Exec(t *testing.T) {
	tests := []struct {
		name string
		w    uint32
		in   map[int]uint32 // register presets
		reg  int            // register to check
		want uint32
	}{
		{"REV", 0xE6BF0F31, map[int]uint32{1: 0x11223344}, 0, 0x44332211},
		{"REV16", 0xE6BF0FB1, map[int]uint32{1: 0x11223344}, 0, 0x22114433},
		{"REVSH", 0xE6FF0FB1, map[int]uint32{1: 0x000080FF}, 0, 0xFFFFFF80},
		{"UXTB", 0xE6EF0071, map[int]uint32{1: 0x123456FE}, 0, 0x000000FE},
		{"SXTB", 0xE6AF0071, map[int]uint32{1: 0x123456FE}, 0, 0xFFFFFFFE},
		{"SXTH", 0xE6BF0071, map[int]uint32{1: 0x1234F001}, 0, 0xFFFFF001},
		{"UXTH", 0xE6FF0071, map[int]uint32{1: 0x1234F001}, 0, 0x0000F001},
		// UMAAL RdLo=r0, RdHi=r1, Rm=r2, Rn=r3: {r1:r0} = r3*r2 + r1 + r0.
		// 0x10000*0x10000 = 0x1_00000000, so RdHi(r1) receives the carry-out 1.
		{"UMAAL-hi", 0xE0410293, map[int]uint32{0: 0, 1: 0, 2: 0x00010000, 3: 0x00010000}, 1, 0x00000001},
		{"UMAAL-lo", 0xE0410293, map[int]uint32{0: 0x00000010, 1: 0x00000020, 2: 0x00010000, 3: 0x00000004}, 0, 0x00040030},
		{"SSAT-hi", 0xE6AF0011, map[int]uint32{1: 0x00010000}, 0, 0x00007FFF},
		{"SSAT-ok", 0xE6AF0011, map[int]uint32{1: 0x00000005}, 0, 0x00000005},
		{"USAT-hi", 0xE6E70011, map[int]uint32{1: 0x00000200}, 0, 0x0000007F},
		{"PKHBT", 0xE6810012, map[int]uint32{1: 0x0000AAAA, 2: 0xBBBB0000}, 0, 0xBBBBAAAA},
		{"UADD8", 0xE6510F92, map[int]uint32{1: 0x01020304, 2: 0x10203040}, 0, 0x11223344},
		{"SMMUL", 0xE750F211, map[int]uint32{1: 0x40000000, 2: 0x40000000}, 0, 0x10000000},
		{"SMUAD", 0xE700F211, map[int]uint32{1: 0x00020003, 2: 0x00040005}, 0, 0x00000017}, // 2*4 + 3*5 = 23
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, b := v6cpu()
			for r, v := range tt.in {
				c.R[r] = v
			}
			stepOne(c, b, tt.w)
			if c.Halted {
				t.Fatalf("core halted: %s", c.HaltReason)
			}
			if got := c.R[tt.reg]; got != tt.want {
				t.Fatalf("r%d = 0x%08X, want 0x%08X", tt.reg, got, tt.want)
			}
		})
	}
}

// LDREX tags an address and a matching STREX to it succeeds (writes 0 to Rd and
// stores the value); a STREX to a different address, or after CLREX, fails.
func TestV6Exclusive(t *testing.T) {
	c, b := v6cpu()
	c.R[1] = 0x2000 // address
	c.R[2] = 0xCAFEF00D

	stepOne(c, b, 0xE1910F9F) // LDREX r0, [r1]
	stepOne(c, b, 0xE1810F92) // STREX r0, r2, [r1]
	if c.R[0] != 0 {
		t.Fatalf("STREX to the tagged address failed (Rd=%d)", c.R[0])
	}
	if b.word(0x2000) != 0xCAFEF00D {
		t.Fatalf("STREX did not store: mem=0x%08X", b.word(0x2000))
	}

	// A second STREX without a fresh LDREX must fail (monitor cleared on success).
	c.R[2] = 0x11111111
	stepOne(c, b, 0xE1810F92)
	if c.R[0] != 1 {
		t.Fatal("STREX succeeded without holding the exclusive monitor")
	}
	if b.word(0x2000) != 0xCAFEF00D {
		t.Fatal("failed STREX wrote memory anyway")
	}
}

// SEL picks bytes from Rn where the GE flag is set and Rm elsewhere; UADD8 sets
// GE per lane on an unsigned carry, so the pair implements a saturating-style
// select in the way compilers emit.
func TestV6SelAfterParallel(t *testing.T) {
	c, b := v6cpu()
	// UADD8 with carries only in some lanes: 0x80+0x80 carries, 0x01+0x01 does not.
	c.R[1] = 0x80018001
	c.R[2] = 0x80018001
	stepOne(c, b, 0xE6510F92) // UADD8 r0, r1, r2
	// lanes 0 and 2 (0x01+0x01) no carry -> GE 0; lanes 1,3 (0x80+0x80) carry -> GE 1.
	if c.GE != 0b1010 {
		t.Fatalf("GE = %04b, want 1010", c.GE)
	}
	c.R[3] = 0xAAAAAAAA
	c.R[4] = 0xBBBBBBBB
	stepOne(c, b, 0xE6830FB4) // SEL r0, r3, r4
	// bytes: lane GE=1 -> from r3(AA), GE=0 -> from r4(BB): BB AA BB AA
	if c.R[0] != 0xAABBAABB {
		t.Fatalf("SEL = 0x%08X, want 0xAABBAABB", c.R[0])
	}
}
