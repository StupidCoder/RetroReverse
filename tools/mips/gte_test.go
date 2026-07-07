package mips

import "testing"

// TestGTE_RTPS_Identity runs RTPS with an identity rotation (0x1000 = 1.0 in the
// 1.12 matrix format), sf=1 (>>12) and zero translation, so the transformed
// MAC1..3 / IR1..3 equal the input vector exactly — a hand-checkable case.
func TestGTE_RTPS_Identity(t *testing.T) {
	g := NewGTE()
	// Rotation matrix = identity: RT11=RT22=RT33=0x1000.
	g.ctrl[0] = 0x1000 // RT11 (RT12 = 0)
	g.ctrl[2] = 0x1000 // RT22 (RT23 = 0)
	g.ctrl[4] = 0x1000 // RT33
	// Translation TR = 0 (ctrl[5..7]); H = 0 so the divide stays in range.
	// V0 = (100, 200, 300).
	g.data[0] = uint32(100) | uint32(200)<<16
	g.data[1] = 300

	g.Command((1 << 19) | 0x01) // RTPS, sf=1

	if g.data[25] != 100 || g.data[26] != 200 || g.data[27] != 300 {
		t.Errorf("MAC1..3 = %d,%d,%d, want 100,200,300", int32(g.data[25]), int32(g.data[26]), int32(g.data[27]))
	}
	if g.irVal(1) != 100 || g.irVal(2) != 200 || g.irVal(3) != 300 {
		t.Errorf("IR1..3 = %d,%d,%d, want 100,200,300", g.irVal(1), g.irVal(2), g.irVal(3))
	}
	if g.data[19]&0xFFFF != 300 {
		t.Errorf("SZ3 = %d, want 300", g.data[19]&0xFFFF)
	}
	if g.ctrl[31] != 0 {
		t.Errorf("FLAG = 0x%08X, want 0 (no saturation/overflow)", g.ctrl[31])
	}
}

// TestGTE_NCLIP checks the signed-area cross product.
func TestGTE_NCLIP(t *testing.T) {
	g := NewGTE()
	// SXY0=(0,0), SXY1=(10,0), SXY2=(0,10): area = 100.
	g.data[12] = 0
	g.data[13] = uint32(uint16(10))
	g.data[14] = uint32(uint16(10)) << 16
	g.Command(0x06) // NCLIP
	if int32(g.data[24]) != 100 {
		t.Errorf("MAC0 = %d, want 100", int32(g.data[24]))
	}

	// Reverse winding -> negative area.
	g.data[13], g.data[14] = g.data[14], g.data[13]
	g.Command(0x06)
	if int32(g.data[24]) != -100 {
		t.Errorf("MAC0 (reversed) = %d, want -100", int32(g.data[24]))
	}
}

// TestGTE_AVSZ3 checks the ordering-table Z averaging.
func TestGTE_AVSZ3(t *testing.T) {
	g := NewGTE()
	g.ctrl[29] = 0x0300 // ZSF3 = 768
	g.data[17] = 100    // SZ1
	g.data[18] = 200    // SZ2
	g.data[19] = 300    // SZ3
	g.Command(0x2D)     // AVSZ3
	// MAC0 = 768 * (100+200+300) = 460800; OTZ = 460800 >> 12 = 112.
	if int32(g.data[24]) != 460800 {
		t.Errorf("MAC0 = %d, want 460800", int32(g.data[24]))
	}
	if g.data[7] != 112 {
		t.Errorf("OTZ = %d, want 112", g.data[7])
	}
}

// TestGTE_Divide spot-checks the perspective divide against the closed-form
// value for a few in-range inputs (the UNR result must match (h*0x10000)/sz3
// rounded, clamped to 0x1FFFF).
func TestGTE_Divide(t *testing.T) {
	g := NewGTE()
	cases := []struct{ h, sz3 uint32 }{
		{200, 400}, {100, 300}, {0x3FF, 0x400}, {1234, 5678},
	}
	for _, c := range cases {
		got := g.divide(c.h, c.sz3)
		want := (uint64(c.h)*0x20000/uint64(c.sz3) + 1) / 2
		if want > 0x1FFFF {
			want = 0x1FFFF
		}
		// The hardware UNR is accurate to within 1 ULP of the exact quotient.
		diff := int64(got) - int64(want)
		if diff < -1 || diff > 1 {
			t.Errorf("divide(%d,%d) = %d, want ~%d", c.h, c.sz3, got, want)
		}
	}
}

// TestGTE_DivideOverflow checks the divide-overflow clamp and flag.
func TestGTE_DivideOverflow(t *testing.T) {
	g := NewGTE()
	if got := g.divide(0x1000, 0); got != 0x1FFFF {
		t.Errorf("divide by zero = %d, want 0x1FFFF", got)
	}
	if g.ctrl[31]&flagDivOvf == 0 {
		t.Error("divide overflow flag not set")
	}
}
