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

// TestLineRasterises drives a LINE primitive — the shape RRV's intro grid and its HUD
// dials are built from, previously counted "not rasterised" and dropped. A horizontal
// line is axis-aligned so the DDA is exact: every pixel between the endpoints is painted,
// none above or below. Mutation test: make gs.line a no-op and the on-line assertion fails.
func TestLineRasterises(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16) // FBP=0, FBW=1, PSM=CT32
	gs.write(gsSCISSOR1, 63<<16|uint64(31)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	var packet []byte
	tag := make([]byte, 16)
	putLE64(tag[0:], uint64(1)|1<<15|uint64(gifPacked)<<58|uint64(4)<<60)
	putLE64(tag[8:], 0x0|0x1<<4|0x5<<8|0x5<<12) // PRIM, RGBAQ, XYZ2, XYZ2
	packet = append(packet, tag...)
	packet = append(packet, packedPrim(1)...) // line, flat
	packet = append(packet, packedRGBA(0xFF, 0xFF, 0xFF)...)
	packet = append(packet, packedXYZ2(2, 5, false)...)
	packet = append(packet, packedXYZ2(12, 5, false)...)
	m.gifPacket(packet)

	if gs.prims != 1 {
		t.Fatalf("expected 1 line primitive, got %d", gs.prims)
	}
	// A pixel on the line is white...
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 7, 5):]); got != 0x80FFFFFF {
		t.Errorf("on-line pixel (7,5) = %08X, want 80FFFFFF — the line was not rasterised", got)
	}
	// ...and the rows above and below it are untouched.
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 7, 4):]); got != 0 {
		t.Errorf("pixel (7,4) above the line = %08X, want 0", got)
	}
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 7, 6):]); got != 0 {
		t.Errorf("pixel (7,6) below the line = %08X, want 0", got)
	}
}

// TestFrameZAsColourWritesThroughZSwizzle drives the "clear/render Z as colour" idiom:
// a game points FRAME at its Z buffer with FRAME.PSM set to a Z format so the ordinary
// colour pipeline fills the depth memory (RRV's intro clears its 0x46000 Z buffer this
// way). PSMZ24 shares PSMCT32's pixel packing but a DIFFERENT block swizzle (^0x18), so
// the colour must land at the Z-swizzled address — writing it through the CT32 swizzle
// would "clear" the wrong words and leave the Z the geometry reads stale. Mutation test:
// route the store through addrPSMCT32 instead of addrPSMZ32 and both assertions flip.
func TestFrameZAsColourWritesThroughZSwizzle(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	// A 64-pixel-wide frame at base 0 whose PSM is PSMZ24 — the colour pipeline writes
	// into Z memory. Z-write masked and Z-test off so the depth stage stays out of the way.
	gs.write(gsFRAME1, 1<<16|uint64(psmZ24)<<24)
	gs.write(gsSCISSOR1, 63<<16|uint64(31)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsZBUF1, 1<<32) // ZMSK=1: no depth write to interfere
	gs.write(gsTEST1, 0)     // alpha test off, ZTE off (always pass)
	gs.write(gsPRMODECONT, 1)

	// One sprite covering (4,4)..(11,9) in a distinctive colour.
	var packet []byte
	tag := make([]byte, 16)
	putLE64(tag[0:], uint64(1)|1<<15|uint64(gifPacked)<<58|uint64(4)<<60)
	putLE64(tag[8:], 0x0|0x1<<4|0x5<<8|0x5<<12)
	packet = append(packet, tag...)
	packet = append(packet, packedPrim(6)...) // sprite
	packet = append(packet, packedRGBA(0x12, 0x34, 0x56)...)
	packet = append(packet, packedXYZ2(4, 4, false)...)
	packet = append(packet, packedXYZ2(12, 10, false)...)
	m.gifPacket(packet)

	// The colour landed at the Z-swizzled word for (5,5).
	const want = 0x80563412 // a=0x80 b=0x56 g=0x34 r=0x12
	if got := le32gs(gs.vram[addrPSMZ32(0, 1, 5, 5):]); got != want {
		t.Errorf("Z-swizzled (5,5) = %08X, want %08X — colour did not reach the Z buffer", got, want)
	}
	// ...and NOT at the CT32-swizzled word, which the depth reader never touches.
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 5, 5):]); got == want {
		t.Errorf("colour reached the CT32-swizzled word for (5,5) — the Z clear would miss the geometry's depth")
	}
}

// TestTriStripADCBreaksWithoutBridging drives a triangle strip that packs two disjoint
// triangles into one primitive and breaks between them with the ADC bit — the exact idiom
// Ridge Racer V's flyover uses. A left triangle (red) and a right triangle (green) sit far
// apart; the two triangles that would bridge them are the strip's ADC=1 vertices, marked
// "register but do not draw". The GS still slides the strip's window through those vertices,
// so the right triangle draws in its own place; it only withholds the bridging kicks.
//
// The bug this guards: treating a non-kicking vertex as a plain append froze the window on
// the first three vertices. The dropped vertices never reached the queue, so the next
// drawing kick stitched a triangle across the gap — a huge stray polygon jutting between the
// two objects, and the far triangle (its vertices dropped) never appeared at all.
//
// Mutation test: make pushVertex append-only for non-kicking writes (the old code) and the
// bridge pixel lights up while the right triangle vanishes — both remaining assertions flip.
func TestTriStripADCBreaksWithoutBridging(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16) // FBP=0, FBW=1, PSMCT32
	gs.write(gsSCISSOR1, 63<<16|uint64(31)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	// tag 1: PRIM(tristrip), RGBAQ(red), and the three vertices of the left triangle.
	var packet []byte
	tag := make([]byte, 16)
	putLE64(tag[0:], uint64(1)|1<<15|uint64(gifPacked)<<58|uint64(5)<<60)
	putLE64(tag[8:], 0x0|0x1<<4|0x5<<8|0x5<<12|0x5<<16) // PRIM,RGBAQ,XYZ2,XYZ2,XYZ2
	packet = append(packet, tag...)
	packet = append(packet, packedPrim(4)...)          // tristrip, flat
	packet = append(packet, packedRGBA(0xFF, 0, 0)...) // red
	packet = append(packet, packedXYZ2(4, 4, false)...)
	packet = append(packet, packedXYZ2(16, 4, false)...)
	packet = append(packet, packedXYZ2(4, 16, false)...)

	// tag 2: recolour green, then the right triangle — its first two vertices carry ADC=1,
	// so the two triangles that would bridge from the left object are suppressed. The window
	// still advances through them, so the third (ADC=0) vertex draws the right triangle.
	tag2 := make([]byte, 16)
	putLE64(tag2[0:], uint64(1)|1<<15|uint64(gifPacked)<<58|uint64(4)<<60)
	putLE64(tag2[8:], 0x1|0x5<<4|0x5<<8|0x5<<12) // RGBAQ,XYZ2,XYZ2,XYZ2
	packet = append(packet, tag2...)
	packet = append(packet, packedRGBA(0, 0xFF, 0)...) // green
	packet = append(packet, packedXYZ2(44, 4, true)...)
	packet = append(packet, packedXYZ2(56, 4, true)...)
	packet = append(packet, packedXYZ2(44, 16, false)...)
	m.gifPacket(packet)

	// Exactly two triangles are drawn — the two objects, not the bridges.
	if gs.prims != 2 {
		t.Fatalf("expected 2 drawn triangles, got %d", gs.prims)
	}
	// The left triangle is red where it covers.
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 7, 7):]); got != 0x800000FF {
		t.Errorf("left-triangle pixel (7,7) = %08X, want 800000FF (red)", got)
	}
	// The right triangle drew in its own place — proof the window slid through the ADC
	// vertices instead of dropping them.
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 47, 7):]); got != 0x8000FF00 {
		t.Errorf("right-triangle pixel (47,7) = %08X, want 8000FF00 (green) — the ADC vertices were dropped, not slid", got)
	}
	// The gap between them, which only a suppressed bridging triangle would cover, is clean.
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 24, 6):]); got != 0 {
		t.Errorf("bridge pixel (24,6) = %08X, want 0 — a suppressed triangle jutted across the gap", got)
	}
}
