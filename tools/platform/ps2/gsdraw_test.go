package ps2

import "testing"

// TestSpriteAndTriangleRasterise drives the GS the way a game does — a PACKED GIF
// packet: PRIM, colour, two XYZ2 kicks for a sprite, then three for a flat triangle —
// and reads the pixels back through the same swizzle the display readout uses.
func TestSpriteAndTriangleRasterise(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	// A 64-pixel-wide PSMCT32 frame at base 0, XYOFFSET 0, scissor the full page.
	gs.write(gsFRAME1, 1<<16)                   // FBP=0, FBW=1, PSM=CT32
	gs.write(gsSCISSOR1, 63<<16|uint64(31)<<48) // x 0..63, y 0..31
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	packet := buildPackedDraw()
	m.gifPacket(packet)

	if gs.prims != 2 {
		t.Fatalf("expected 2 primitives (one sprite, one triangle), got %d", gs.prims)
	}

	// The sprite filled (4,4)..(11,9) with solid red (the second vertex's colour).
	red := le32gs(gs.vram[addrPSMCT32(0, 1, 5, 5):])
	if red != 0x800000FF {
		t.Errorf("sprite pixel (5,5) = %08X, want 800000FF (red)", red)
	}
	if out := le32gs(gs.vram[addrPSMCT32(0, 1, 12, 5):]); out == 0x000000FF {
		t.Errorf("pixel (12,5) right of the sprite was painted")
	}

	// The triangle (20,4) (30,4) (20,14) covers its own corner but not the far one.
	green := le32gs(gs.vram[addrPSMCT32(0, 1, 21, 5):])
	if green != 0x8000FF00 {
		t.Errorf("triangle pixel (21,5) = %08X, want 8000FF00 (green)", green)
	}
	if out := le32gs(gs.vram[addrPSMCT32(0, 1, 29, 13):]); out == 0x0000FF00 {
		t.Errorf("pixel (29,13) outside the triangle's hypotenuse was painted")
	}
}

// buildPackedDraw is a PACKED packet with PRIM/RGBAQ/XYZ2 descriptors, not A+D — the
// layout a real display list uses.
func buildPackedDraw() []byte {
	var out []byte
	tag := make([]byte, 16)
	// NLOOP=8, EOP, PACKED, NREG=1... simpler: one descriptor per loop won't sequence
	// registers; use NREG with the descriptor list PRIM,RGBAQ,XYZ2 x2 then RGBAQ,XYZ2 x3.
	// Two tags keep it readable: sprite then triangle.
	// Sprite tag: NREG=4, regs = PRIM(0), RGBAQ(1), XYZ2(5), XYZ2(5); NLOOP=1.
	lo := uint64(1) | 1<<15 | uint64(gifPacked)<<58 | uint64(4)<<60
	putLE64(tag[0:], lo)
	putLE64(tag[8:], 0x0|0x1<<4|0x5<<8|0x5<<12)
	out = append(out, tag...)
	out = append(out, packedPrim(6)...)          // sprite
	out = append(out, packedRGBA(0xFF, 0, 0)...) // red
	out = append(out, packedXYZ2(4, 4, false)...)
	out = append(out, packedXYZ2(12, 10, false)...)

	tag2 := make([]byte, 16)
	lo2 := uint64(1) | 1<<15 | uint64(gifPacked)<<58 | uint64(5)<<60
	putLE64(tag2[0:], lo2)
	putLE64(tag2[8:], 0x0|0x1<<4|0x5<<8|0x5<<12|0x5<<16)
	out = append(out, tag2...)
	out = append(out, packedPrim(3)...)          // triangle, flat
	out = append(out, packedRGBA(0, 0xFF, 0)...) // green
	out = append(out, packedXYZ2(20, 4, false)...)
	out = append(out, packedXYZ2(30, 4, false)...)
	out = append(out, packedXYZ2(20, 14, false)...)
	return out
}

func packedPrim(typ uint64) []byte {
	q := make([]byte, 16)
	putLE64(q[0:], typ)
	return q
}

func packedRGBA(r, g, b uint64) []byte {
	q := make([]byte, 16)
	putLE64(q[0:], r|g<<32)
	putLE64(q[8:], b|0x80<<32)
	return q
}

func packedXYZ2(x, y int, adc bool) []byte {
	q := make([]byte, 16)
	putLE64(q[0:], uint64(x<<4)|uint64(y<<4)<<32) // 12.4 fixed
	var hi uint64
	if adc {
		hi |= 1 << 47
	}
	putLE64(q[8:], hi)
	return q
}
