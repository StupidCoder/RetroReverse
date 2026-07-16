package gc

import "testing"

// The compare-mode tests below pin the encoding this game's shadow stage relies on. The
// arrangement is taken from that stage's own registers, read out of the frame that draws the
// map: the colour combiner compares the fragment's light-space depth (arriving as CPREV)
// against the shadow map's (arriving as TEXC), and the ALPHA combiner — whose own two
// operands are both ZERO — emits KONST or nothing on the strength of that colour comparison.

// TestTevCompareSelectors walks the eight comparisons the selector names, at each width, with
// operands either side of the boundary.
func TestTevCompareSelectors(t *testing.T) {
	// little-endian bytes of one number: GR16 reads green:red, BGR24 blue:green:red.
	rgb := func(r, g, b float32) [3]float32 { return [3]float32{r, g, b} }

	cases := []struct {
		name   string
		sel    int
		a, b   [3]float32
		ch     int
		expect bool
	}{
		{"R8_GT true", 0, rgb(9, 0, 0), rgb(8, 0, 0), -1, true},
		{"R8_GT false when equal", 0, rgb(8, 0, 0), rgb(8, 0, 0), -1, false},
		{"R8_GT ignores green", 0, rgb(8, 99, 0), rgb(8, 1, 0), -1, false},
		{"R8_EQ true", 1, rgb(8, 99, 0), rgb(8, 1, 0), -1, true},

		// GR16: green is the HIGH byte, so a bigger green wins regardless of red.
		{"GR16_GT high byte decides", 2, rgb(0, 2, 0), rgb(255, 1, 0), -1, true},
		{"GR16_GT low byte breaks tie", 2, rgb(5, 1, 0), rgb(4, 1, 0), -1, true},
		{"GR16_GT false when equal", 2, rgb(4, 1, 0), rgb(4, 1, 0), -1, false},
		{"GR16_EQ true", 3, rgb(4, 1, 9), rgb(4, 1, 7), -1, true},

		{"BGR24_GT high byte decides", 4, rgb(0, 0, 2), rgb(255, 255, 1), -1, true},
		{"BGR24_EQ true", 5, rgb(1, 2, 3), rgb(1, 2, 3), -1, true},
		{"BGR24_EQ false", 5, rgb(1, 2, 3), rgb(1, 2, 4), -1, false},

		// RGB8 compares each channel against its own opposite number.
		{"RGB8_GT red channel", 6, rgb(9, 0, 0), rgb(8, 5, 5), 0, true},
		{"RGB8_GT green channel same operands", 6, rgb(9, 0, 0), rgb(8, 5, 5), 1, false},
		{"RGB8_EQ blue channel", 7, rgb(0, 0, 5), rgb(9, 9, 5), 2, true},
	}
	for _, c := range cases {
		if got := tevCompare(c.sel, c.a, c.b, 0, 0, c.ch); got != c.expect {
			t.Errorf("%s: sel %d got %v want %v", c.name, c.sel, got, c.expect)
		}
	}
}

// TestTevCompareA8UsesAlphaOperands is the other half of the selector's top width: where the
// colour pipe compares per channel, the alpha pipe compares its own two operands.
func TestTevCompareA8UsesAlphaOperands(t *testing.T) {
	colour := [3]float32{255, 255, 255} // would win every colour comparison
	if tevCompare(6, colour, [3]float32{}, 3, 9, -1) {
		t.Error("A8_GT should compare the alpha operands (3 > 9 is false), not the colour ones")
	}
	if !tevCompare(6, [3]float32{}, colour, 9, 3, -1) {
		t.Error("A8_GT should compare the alpha operands (9 > 3 is true), not the colour ones")
	}
}

// TestAlphaCompareReadsColourOperands is the behaviour the shadow test is built on, and the
// one a per-combiner reading of the registers would get wrong: a narrow selector in the alpha
// combiner compares the COLOUR pipe's operands. Both alpha operands here are ZERO, exactly as
// the game programs them, so a combiner comparing its own inputs would answer the same way
// whatever the depths were — the bug that leaves the shadow missing.
func TestAlphaCompareReadsColourOperands(t *testing.T) {
	var reg [4][4]float32
	tex := [4]float32{}
	ras := [4]float32{}

	// alpha combiner: a=ZERO(7) b=ZERO(7) c=KONST(6) d=ZERO(7), bias=3 (compare),
	// op=0 (GT), scale=1 (GR16) — the game's stage 1 encoding.
	const ac = 7<<13 | 7<<10 | 6<<7 | 7<<4 | tevBiasCompare<<16 | 0<<18 | 1<<20
	const konst = 200

	// lightDepth > shadowDepth: in shadow, so the stage emits KONST.
	lit := [3]float32{0, 5, 0}    // GR16 = 0x0500
	shadow := [3]float32{0, 4, 0} // GR16 = 0x0400
	if got := combineAlpha(ac, reg, tex, ras, konst, lit, shadow); got != konst {
		t.Errorf("lit(0x0500) > shadow(0x0400) should emit KONST %d, got %v", konst, got)
	}
	// The other way round: not in shadow, so nothing is added.
	if got := combineAlpha(ac, reg, tex, ras, konst, shadow, lit); got != 0 {
		t.Errorf("shadow(0x0400) > lit(0x0500) is false, should emit 0, got %v", got)
	}
}

// TestCompareModeIsNotArithmetic guards the specific silent failure this replaced: bias 3 used
// to fall through the arithmetic switch with no case for it, and scale 3 was read as a *0.5
// shift. A compare stage therefore computed a plausible number instead of a verdict.
func TestCompareModeIsNotArithmetic(t *testing.T) {
	var reg [4][4]float32
	tex, ras := [4]float32{}, [4]float32{}
	konst := [3]float32{0, 0, 0}

	// colour: a=C0(2) b=C1(4) c=ZERO(15) d=ZERO(15), bias=3, op=0 (GT), scale=1 (GR16).
	const cc = 2<<12 | 4<<8 | 15<<4 | 15 | tevBiasCompare<<16 | 1<<20
	reg[1] = [4]float32{0, 5, 0, 0} // C0: GR16 = 0x0500
	reg[2] = [4]float32{0, 4, 0, 0} // C1: GR16 = 0x0400

	out, ca, cb := combineColor(cc, reg, tex, ras, konst)
	// c is ZERO and d is ZERO, so the verdict adds nothing either way — the colour result of
	// the game's shadow stage is deliberately inert; the verdict is the alpha combiner's.
	if out != ([3]float32{0, 0, 0}) {
		t.Errorf("c=ZERO d=ZERO should leave the colour result zero, got %v", out)
	}
	// The operands must come back out for the alpha combiner to compare.
	if ca != ([3]float32{0, 5, 0}) || cb != ([3]float32{0, 4, 0}) {
		t.Errorf("colour operands not returned: a=%v b=%v", ca, cb)
	}
}

// TestSwapTableDecode pins the layout: table t is built from KSEL 0xF6+2t (red, then green in
// the low nibble) and 0xF6+2t+1 (blue, then alpha).
func TestSwapTableDecode(t *testing.T) {
	g := &gpu{}
	// The tables this game programs for its shadow draws, read out of the frame that draws the
	// map. Table 1 lifts alpha into green; table 2 lifts alpha into red.
	g.BP[0xF6], g.BP[0xF7] = 0<<0|1<<2, 2<<0|3<<2 // table0: identity
	g.BP[0xF8], g.BP[0xF9] = 0<<0|3<<2, 2<<0|3<<2 // table1: r<-r g<-a
	g.BP[0xFA], g.BP[0xFB] = 3<<0|1<<2, 2<<0|3<<2 // table2: r<-a g<-g

	for _, c := range []struct {
		tbl  int
		want [4]int
	}{
		{0, [4]int{0, 1, 2, 3}},
		{1, [4]int{0, 3, 2, 3}},
		{2, [4]int{3, 1, 2, 3}},
	} {
		if got := g.swapTable(c.tbl); got != c.want {
			t.Errorf("table %d: got %v, want %v", c.tbl, got, c.want)
		}
	}
}

// TestSwapTableAssemblesSixteenBits is the reason the tables exist here at all. An IA8 texture
// samples as (I,I,I,A), so a 16-bit r:g compare would read (I,I) — the same byte twice, and no
// 16-bit number. The swap is what makes r:g the two halves of one value.
func TestSwapTableAssemblesSixteenBits(t *testing.T) {
	g := &gpu{}
	g.BP[0xF8], g.BP[0xF9] = 0<<0|3<<2, 2<<0|3<<2 // table1: r<-r, g<-a

	const lo, hi = 0x34, 0x12
	ia8 := [4]float32{lo, lo, lo, hi} // as an IA8 texel arrives: (I,I,I,A)

	if raw := int(ia8[1])<<8 | int(ia8[0]); raw != 0x3434 {
		t.Fatalf("unswapped r:g is the intensity twice over (0x%04X) — the premise of this test", raw)
	}
	sw := swizzle(g.swapTable(1), ia8)
	if got := int(sw[1])<<8 | int(sw[0]); got != 0x1234 {
		t.Errorf("swapped r:g = 0x%04X, want the assembled 0x1234", got)
	}
}

// TestZeroSwapTableIsNotIdentity guards the trap this cost a test run to find: an all-zero
// table selects RED for every channel. Identity is 0,1,2,3 and has to be programmed — by GXInit
// on the hardware, and explicitly by any test building a gpu of its own.
func TestZeroSwapTableIsNotIdentity(t *testing.T) {
	g := &gpu{}
	if got := g.swapTable(0); got != ([4]int{0, 0, 0, 0}) {
		t.Fatalf("a zeroed KSEL should decode to all-red, got %v", got)
	}
	grey := swizzle(g.swapTable(0), [4]float32{10, 20, 30, 40})
	if grey != ([4]float32{10, 10, 10, 10}) {
		t.Errorf("all-red table should flatten the colour to its red, got %v", grey)
	}
}
