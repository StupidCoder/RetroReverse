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

// TestGTE_NCS_PreservesCode checks the normal-colour lighting pipeline: with a
// unit +Z normal and identity light/colour matrices it produces a non-zero lit
// colour on the RGB FIFO, and — crucially for building lit primitives — carries
// the GPU command byte from RGBC (bits 24-31) through to the pushed colour word.
func TestGTE_NCS_PreservesCode(t *testing.T) {
	g := NewGTE()
	// Light matrix LLM (ctrl 8..12) = identity (0x1000 diagonal).
	g.ctrl[8], g.ctrl[10], g.ctrl[12] = 0x1000, 0x1000, 0x1000
	// Light-colour matrix LCM (ctrl 16..20) = identity.
	g.ctrl[16], g.ctrl[18], g.ctrl[20] = 0x1000, 0x1000, 0x1000
	// RGBC: colour grey, command byte 0x2C (a flat textured quad).
	g.data[6] = 0x2C808080
	// Normal V0 = (0,0,0x1000) — points along +Z.
	g.data[0], g.data[1] = 0, 0x1000

	g.Command((1 << 19) | (1 << 10) | 0x1E) // NCS, sf=12, lm=1

	rgb2 := g.Read(22)
	if code := rgb2 >> 24; code != 0x2C {
		t.Errorf("pushed colour code byte = 0x%02X, want 0x2C (RGBC command byte)", code)
	}
	if rgb2&0xFFFFFF == 0 {
		t.Errorf("lit colour is zero (0x%06X); a lit surface should be non-black", rgb2&0xFFFFFF)
	}
}

// TestGTE_INTPL_Fog checks the fog interpolation: with IR0=0 (near) the source
// colour passes through, and the RGBC command byte is preserved on the FIFO.
func TestGTE_INTPL_Fog(t *testing.T) {
	g := NewGTE()
	g.data[6] = 0x24000000       // command byte 0x24, colour ignored by INTPL
	g.data[9], g.data[10], g.data[11] = 0x400, 0x400, 0x400 // IR1..3 base colour
	g.data[8] = 0                // IR0 = 0 -> near, no fog: keep the source colour
	// Far colour FC (ctrl 21..23) arbitrary; with IR0=0 it must not matter.
	g.ctrl[21], g.ctrl[22], g.ctrl[23] = 0x1000, 0, 0

	g.Command((1 << 19) | 0x11) // INTPL, sf=12

	rgb2 := g.Read(22)
	if code := rgb2 >> 24; code != 0x24 {
		t.Errorf("INTPL code byte = 0x%02X, want 0x24", code)
	}
	// IR base 0x400 -> MAC 0x400<<12 -> >>12 -> 0x400 -> colour /16 = 0x40 each lane.
	if r := rgb2 & 0xFF; r != 0x40 {
		t.Errorf("INTPL near colour R = 0x%02X, want 0x40 (source colour passthrough)", r)
	}
}
