package ps2

import "testing"

// TestSwizzlesAreInjective proves each format's address function maps a buffer's texels
// onto distinct storage — the property that makes an upload readable back. A collision
// or an out-of-page escape would silently corrupt a neighbour.
func TestSwizzlesAreInjective(t *testing.T) {
	// PSMT8: two pages side by side (256x64, bw = 256/64 = 4).
	seen := map[uint32]bool{}
	for y := uint32(0); y < 64; y++ {
		for x := uint32(0); x < 256; x++ {
			a := addrPSMT8(0, 4, x, y)
			if seen[a] {
				t.Fatalf("PSMT8 (%d,%d) collides at 0x%X", x, y, a)
			}
			seen[a] = true
			if a >= 2*8192 {
				t.Fatalf("PSMT8 (%d,%d) escapes its two pages: 0x%X", x, y, a)
			}
		}
	}

	// PSMT4: one page, 128x128, bw = 2.
	seenNib := map[uint32]bool{}
	for y := uint32(0); y < 128; y++ {
		for x := uint32(0); x < 128; x++ {
			a, nib := addrPSMT4(0, 2, x, y)
			key := a*2 + nib
			if seenNib[key] {
				t.Fatalf("PSMT4 (%d,%d) collides at 0x%X.%d", x, y, a, nib)
			}
			seenNib[key] = true
			if a >= 8192 {
				t.Fatalf("PSMT4 (%d,%d) escapes its page: 0x%X", x, y, a)
			}
		}
	}

	// PSMCT16 and CT16S: one page, 64x64, bw = 1.
	for _, s := range []bool{false, true} {
		seen16 := map[uint32]bool{}
		for y := uint32(0); y < 64; y++ {
			for x := uint32(0); x < 64; x++ {
				a := addrPSMCT16(0, 1, x, y, s)
				if seen16[a] {
					t.Fatalf("PSMCT16(s=%v) (%d,%d) collides at 0x%X", s, x, y, a)
				}
				seen16[a] = true
				if a&1 != 0 || a >= 8192 {
					t.Fatalf("PSMCT16(s=%v) (%d,%d) misaligned or escaped: 0x%X", s, x, y, a)
				}
			}
		}
	}
}

// TestTexturedSpriteThroughCLUT walks the whole road a 2D screen uses: upload a CLUT
// (PSMCT32 image transfer), upload 8-bit indices (PSMT8 transfer), point TEX0 at both
// with CLD=1, and draw a textured sprite. The pixels must come back as the CLUT colours
// the indices name, through the upload swizzle and the sampler's.
func TestTexturedSpriteThroughCLUT(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16) // FBP=0, FBW=1 (64px), PSM=CT32
	gs.write(gsSCISSOR1, 63<<16|uint64(63)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	// The CLUT: 256 CT32 entries at block 0x100; entry i = solid (i, 255-i, i^0x55, 0x80).
	gs.write(gsBITBLTBUF, uint64(0x100)<<32|uint64(1)<<48|uint64(psmCT32)<<56)
	gs.write(gsTRXPOS, 0)
	gs.write(gsTRXREG, 16|uint64(16)<<32)
	gs.write(gsTRXDIR, 0)
	var clutData []byte
	for i := 0; i < 256; i++ {
		// CSM1 swizzles entries into 8x2 half-blocks; feeding raster order means the
		// pixel at (x,y) must be the entry whose CSM1 slot is (x,y).
		x, y := uint32(i%16), uint32(i/16)
		// Invert the CSM1 mapping: which entry lands at (x, y)?
		e := x&7 | (y&1)<<3 | (x&8)<<1 | (y&0xE)<<4
		clutData = append(clutData, byte(e), byte(255-e), byte(e^0x55), 0x80)
	}
	for i := 0; i < len(clutData); i += 8 {
		gs.write(gsHWREG, le64(clutData[i:]))
	}

	// The texture: 64x64 PSMT8 at block 0x140, texel (x,y) = index (x+y)&0xFF.
	gs.write(gsBITBLTBUF, uint64(0x140)<<32|uint64(1)<<48|uint64(psmT8)<<56)
	gs.write(gsTRXPOS, 0)
	gs.write(gsTRXREG, 64|uint64(64)<<32)
	gs.write(gsTRXDIR, 0)
	var texData []byte
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			texData = append(texData, byte(x+y))
		}
	}
	for i := 0; i < len(texData); i += 8 {
		gs.write(gsHWREG, le64(texData[i:]))
	}

	// TEX0: base 0x140, width 1, PSMT8, 64x64, TCC=1, DECAL, CLUT at 0x100 CT32 CSM1 CLD=1.
	tex0 := uint64(0x140) | 1<<14 | uint64(psmT8)<<20 | 6<<26 | 6<<30 | 1<<34 | 1<<35 |
		uint64(0x100)<<37 | uint64(psmCT32)<<51 | 1<<61
	gs.write(gsTEX0_1, tex0)
	gs.write(gsCLAMP1, 0) // repeat

	// PRIM: sprite, textured, FST=1 (UV). Draw (0,0)..(32,32) mapping texels
	// (0.5,0.5)..(32.5,32.5) — the half-texel every real sprite authors (sprites
	// interpolate at the integer pixel coordinate and the sample point sits half a
	// texel below, so 0.5-authored UVs read texel x exactly; 0.0-authored ones read
	// texel x-1 on silicon too).
	gs.write(gsPRIM, 6|1<<4|1<<8)
	gs.write(gsRGBAQ, 0x80808080)
	gs.write(gsUV, 8|8<<16)
	gs.write(gsXYZ3, 0)
	gs.write(gsUV, (32*16+8)|uint64(32*16+8)<<16)
	gs.write(gsXYZ2, 32*16|32*16<<16)

	if gs.prims != 1 {
		t.Fatalf("expected 1 primitive, got %d", gs.prims)
	}
	for _, at := range [][2]int32{{0, 0}, {5, 9}, {31, 31}, {17, 3}} {
		x, y := at[0], at[1]
		idx := byte(x + y)
		want := uint32(idx) | uint32(255-idx)<<8 | uint32(idx^0x55)<<16 | 0x80<<24
		got := le32gs(gs.vram[addrPSMCT32(0, 1, uint32(x), uint32(y)):])
		if got != want {
			t.Errorf("pixel (%d,%d) = %08X, want %08X (index %d)", x, y, got, want, idx)
		}
	}
}

// TestExpand16AppliesTEXA covers the three CT16 alpha cases: bit 15 set takes TA1,
// clear takes TA0, and AEM makes an all-zero pixel transparent.
func TestExpand16AppliesTEXA(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()
	gs.write(gsTEXA, uint64(0x40)|1<<15|uint64(0xC0)<<32) // TA0=0x40, AEM, TA1=0xC0

	if got := gs.expand16(0x8000); got>>24 != 0xC0 {
		t.Errorf("bit15 pixel alpha = %02X, want C0 (TA1)", got>>24)
	}
	if got := gs.expand16(0x001F); got>>24 != 0x40 || got&0xFF != 0xF8 {
		t.Errorf("red pixel = %08X, want alpha 40 (TA0), red F8", got)
	}
	if got := gs.expand16(0x0000); got != 0 {
		t.Errorf("zero pixel with AEM = %08X, want fully transparent", got)
	}
}

// TestBlendEquation checks the standard source-over blend and the FIX constant path.
func TestBlendEquation(t *testing.T) {
	// (Cs - Cd) * As >> 7 + Cd. Source is blue (B=0xFF) with As = 0x80 (1.0), dest is
	// red (R=0xFF): full alpha replaces the dest.
	alpha := uint64(0<<0 | 1<<2 | 0<<4 | 1<<6)
	if got := blendPixel(alpha, 0x80FF0000, 0x000000FF, 0x80, true); got&0xFFFFFF != 0xFF0000 {
		t.Errorf("full-alpha source-over = %08X, want blue only", got)
	}
	// As = 0x40 (0.5): both channels land halfway.
	half := blendPixel(alpha, 0x40FF0000, 0x000000FF, 0x80, true)
	if r, b := half&0xFF, half>>16&0xFF; r < 0x7E || r > 0x81 || b < 0x7E || b > 0x81 {
		t.Errorf("half blend = %08X, want red ~0x80 and blue ~0x7F", half)
	}
	// C = FIX (2), FIX = 0x80, A=src B=0 D=0 -> source through.
	fixed := uint64(0<<0|2<<2|2<<4|2<<6) | uint64(0x80)<<32
	if got := blendPixel(fixed, 0x00123456, 0xFFFFFFFF, 0x80, true); got&0xFFFFFF != 0x123456 {
		t.Errorf("FIX passthrough = %08X, want 123456", got)
	}
}

// TestAlphaTestCutsOut proves a sprite drawn with ATST=NOTEQUAL 0 drops its transparent
// pixels instead of painting them.
func TestAlphaTestCutsOut(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()
	gs.write(gsFRAME1, 1<<16)
	gs.write(gsSCISSOR1, 63<<16|uint64(63)<<48)
	gs.write(gsPRMODECONT, 1)
	gs.write(gsTEST1, 1|7<<1|0<<4) // ATE, NOTEQUAL, AREF=0

	// An untextured sprite with alpha 0: every pixel fails and nothing is written.
	gs.write(gsPRIM, 6)
	gs.write(gsRGBAQ, 0x00FF00FF)
	gs.write(gsXYZ3, 0)
	gs.write(gsXYZ2, 8*16|8*16<<16)
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 4, 4):]); got != 0 {
		t.Errorf("alpha-0 pixel written: %08X", got)
	}

	// The same sprite with alpha 0x80 passes.
	gs.write(gsRGBAQ, 0x80FF00FF)
	gs.write(gsXYZ3, 0)
	gs.write(gsXYZ2, 8*16|8*16<<16)
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 4, 4):]); got != 0x80FF00FF {
		t.Errorf("alpha-0x80 pixel = %08X, want 80FF00FF", got)
	}
}

// TestDepthTestOrdersSprites draws a near sprite then a far one with ZTST=GREATER: the
// far one must lose, and the Z buffer must hold the near depth.
func TestDepthTestOrdersSprites(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()
	gs.write(gsFRAME1, 1<<16) // FBP=0, FBW=1
	gs.write(gsSCISSOR1, 63<<16|uint64(31)<<48)
	gs.write(gsPRMODECONT, 1)
	gs.write(gsZBUF1, 8|0<<24)     // Z buffer at page 8, Z32
	gs.write(gsTEST1, 1<<16|3<<17) // ZTE, GREATER
	gs.write(gsPRIM, 6)            // flat sprite

	draw := func(z uint64, rgba uint64) {
		gs.write(gsRGBAQ, rgba)
		gs.write(gsXYZ3, z<<32)
		gs.write(gsXYZ2, 8*16|8*16<<16|z<<32)
	}
	draw(1000, 0x800000FF) // near, red
	draw(500, 0x8000FF00)  // far, green — must lose

	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 4, 4):]); got != 0x800000FF {
		t.Errorf("far sprite won the depth test: %08X, want 800000FF", got)
	}
	if got := le32gs(gs.vram[addrPSMZ32(8*32, 1, 4, 4):]); got != 1000 {
		t.Errorf("Z buffer holds %d, want 1000", got)
	}

	draw(2000, 0x80FF0000) // nearer, blue — must win
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 4, 4):]); got != 0x80FF0000 {
		t.Errorf("nearer sprite lost: %08X, want 80FF0000", got)
	}
}

// TestNearestSampleSitsHalfBelow pins the sampler's coordinate arithmetic: the
// GS's sample point sits half a texel below the interpolated coordinate, so a
// 1:1 sprite whose UVs are authored at +0.5 (u = x+1.0 at pixel centres) reads
// texel x, not x+1. RRV's FM-card grade pass is built on this: its CT32-as-T8
// gathers author +0.5 and its per-page composites +1.0, and under floor(u) the
// gathers land one T8 row low — on the wrong byte lane of the pun, which is
// what shredded the flyover into period-4 channel noise.
func TestNearestSampleSitsHalfBelow(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16) // FBP=0, FBW=1, PSM=CT32
	gs.write(gsSCISSOR1, 63<<16|uint64(63)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	// A 4x1 CT32 texture at block 0x140: red, green, blue, white.
	gs.write(gsBITBLTBUF, uint64(0x140)<<32|uint64(1)<<48|uint64(psmCT32)<<56)
	gs.write(gsTRXPOS, 0)
	gs.write(gsTRXREG, 4|uint64(1)<<32)
	gs.write(gsTRXDIR, 0)
	gs.write(gsHWREG, uint64(0x800000FF)|uint64(0x8000FF00)<<32)
	gs.write(gsHWREG, uint64(0x80FF0000)|uint64(0x80FFFFFF)<<32)

	// TEX0: 4x1 CT32, TCC=1, DECAL. TEX1=0: nearest.
	gs.write(gsTEX0_1, uint64(0x140)|1<<14|uint64(psmCT32)<<20|2<<26|0<<30|1<<34|1<<35)
	gs.write(gsTEX1_1, 0)
	gs.write(gsCLAMP1, 0)

	// Sprite (0,0)-(2,1), UV (0.5,0.5)-(2.5,1.5): pixel centres sample u = 1.0, 2.0.
	gs.write(gsPRIM, 6|1<<4|1<<8)
	gs.write(gsRGBAQ, 0x80808080)
	gs.write(gsUV, 8|8<<16)
	gs.write(gsXYZ3, 0)
	gs.write(gsUV, (2*16+8)|(1*16+8)<<16)
	gs.write(gsXYZ2, (2*16)|(1*16)<<16)

	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 0, 0):]); got != 0x800000FF {
		t.Errorf("pixel (0,0) = %08X, want 800000FF (texel 0 — sampling must sit half a texel below u=1.0)", got)
	}
	if got := le32gs(gs.vram[addrPSMCT32(0, 1, 1, 0):]); got != 0x8000FF00 {
		t.Errorf("pixel (1,0) = %08X, want 8000FF00 (texel 1)", got)
	}
}

// TestBilinearBlendsWhenTEX1Asks pins the MMAG=LINEAR path: a sample point
// exactly between a black and a white texel comes back mid-grey, not one of
// the two texels.
func TestBilinearBlendsWhenTEX1Asks(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16)
	gs.write(gsSCISSOR1, 63<<16|uint64(63)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	// A 2x1 CT32 texture at block 0x140: black then white.
	gs.write(gsBITBLTBUF, uint64(0x140)<<32|uint64(1)<<48|uint64(psmCT32)<<56)
	gs.write(gsTRXPOS, 0)
	gs.write(gsTRXREG, 2|uint64(1)<<32)
	gs.write(gsTRXDIR, 0)
	gs.write(gsHWREG, uint64(0x80000000)|uint64(0x80FFFFFF)<<32)

	gs.write(gsTEX0_1, uint64(0x140)|1<<14|uint64(psmCT32)<<20|1<<26|0<<30|1<<34|1<<35)
	gs.write(gsTEX1_1, 1<<5) // MMAG=LINEAR
	gs.write(gsCLAMP1, uint64(1)|1<<2) // clamp, so the blend partners are the two texels

	// A 1x1 sprite whose pixel (0,0) — sprites interpolate at the integer pixel
	// coordinate — carries u = 1.0: the sample point u-0.5 = 0.5 sits exactly
	// between the two texels, so LINEAR must return the 50/50 blend.
	gs.write(gsPRIM, 6|1<<4|1<<8)
	gs.write(gsRGBAQ, 0x80808080)
	gs.write(gsUV, 16|16<<16)
	gs.write(gsXYZ3, 0)
	gs.write(gsUV, (2*16)|(2*16)<<16)
	gs.write(gsXYZ2, (1*16)|(1*16)<<16)

	got := le32gs(gs.vram[addrPSMCT32(0, 1, 0, 0):])
	r := got & 0xFF
	if r < 0x7E || r > 0x81 {
		t.Errorf("pixel (0,0) = %08X, want mid-grey (~7F): TEX1's MMAG bit must blend, not pick", got)
	}
}

// TestSpriteSamplesAtIntegerPixel pins WHERE a sprite interpolates its texture
// coordinate: at the integer pixel coordinate, not the pixel centre. The
// discriminating draw is RRV's pyramid gather — a 1:3-stride sprite (6 dest rows
// spanning 18 texel rows, v authored 0.5..18.5). At dest row y the coordinate is
// v = 0.5 + 3y, and the sample half below reads texel 3y exactly; centre
// interpolation would land on 3y+2 and read texel 3y+1, and on the game's real
// packets (v 48.5..66.5 + a REGION_REPEAT byte-lane fold that cannot wrap v>=64)
// it overshoots the 64-row page and reads 32 pixel rows away. 1:1-stride sprites
// read identically under either rule, which is why nothing else ever noticed.
func TestSpriteSamplesAtIntegerPixel(t *testing.T) {
	m := NewMachine()
	gs := m.ensureGS()

	gs.write(gsFRAME1, 1<<16)
	gs.write(gsSCISSOR1, 63<<16|uint64(63)<<48)
	gs.write(gsXYOFFSET1, 0)
	gs.write(gsPRMODECONT, 1)

	// A 2x32 CT32 texture at block 0x140 whose row r is red=r.
	gs.write(gsBITBLTBUF, uint64(0x140)<<32|uint64(1)<<48|uint64(psmCT32)<<56)
	gs.write(gsTRXPOS, 0)
	gs.write(gsTRXREG, 2|uint64(32)<<32)
	gs.write(gsTRXDIR, 0)
	for r := 0; r < 32; r++ {
		px := uint64(0x80)<<24 | uint64(r)
		gs.write(gsHWREG, px|px<<32)
	}

	gs.write(gsTEX0_1, uint64(0x140)|1<<14|uint64(psmCT32)<<20|1<<26|5<<30|1<<34|1<<35)
	gs.write(gsTEX1_1, 0)      // NEAREST
	gs.write(gsCLAMP1, 1|1<<2) // clamp

	// Sprite: dest (0,0)..(1,6), uv (0.5,0.5)..(1.5,18.5) — 1:3 vertical.
	gs.write(gsPRIM, 6|1<<4|1<<8)
	gs.write(gsRGBAQ, 0x80808080)
	gs.write(gsUV, 8|8<<16)
	gs.write(gsXYZ3, 0)
	gs.write(gsUV, (1*16+8)|uint64(18*16+8)<<16)
	gs.write(gsXYZ2, (1*16)|uint64(6*16)<<16)

	for _, y := range []uint32{0, 2, 5} {
		got := le32gs(gs.vram[addrPSMCT32(0, 1, 0, y):]) & 0xFF
		if got != 3*y {
			t.Errorf("dest row %d sampled texel row %d, want %d (integer-pixel interpolation)", y, got, 3*y)
		}
	}
}
