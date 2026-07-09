package n64

import "testing"

// The combiner's multiplexer selects. Note that the colour *mul* input has no
// constant one: select 6 there is KEYSCALE, not 1. Only subA and add offer a
// one. A combiner that wants to pass a value through untouched therefore adds
// it rather than multiplying by it.
const (
	ccCombined = 0
	ccTexel0   = 1
	ccPrim     = 3
	ccShade    = 4
	ccOne      = 6  // subA and add only
	ccZeroC    = 8  // any colour select 8..15 is zero
	ccZeroMul  = 16 // any colour mul select 16..31 is zero
	ccZeroA    = 7  // alpha select 7 is zero
)

// combineWord packs one cycle's eight multiplexers, leaving the second cycle
// selecting zero.
func combineWord(subAC, subBC, mulC, addC, subAA, subBA, mulA, addA uint64) uint64 {
	return subAC<<52 | mulC<<47 | subAA<<44 | mulA<<41 |
		subBC<<28 | addC<<15 | subBA<<12 | addA<<9
}

func TestCombinerModulatesTextureByShade(t *testing.T) {
	// (texel - 0) * shade + 0, the single most common combiner a game sets.
	r := &rdp{OtherModes: uint64(cycle1) << 52}
	r.Combine = combineWord(ccTexel0, ccZeroC, ccShade, ccZeroC, ccTexel0, ccZeroA, ccShade, ccZeroA)

	in := combineInputs{
		Texel0: rgba{200, 100, 50, 255},
		Shade:  rgba{255, 255, 255, 255},
	}
	// A full-brightness shade is 255, not 256, so modulating by it costs the
	// texel a fraction: 200*255 >> 8 is 199. The hardware is off by that same
	// one, which is why an N64 texture drawn "unmodified" through the combiner
	// is imperceptibly darker than its texels.
	got := r.combine(&in)
	for i, pair := range [][2]uint32{{got.R, 200}, {got.G, 100}, {got.B, 50}} {
		if pair[0] > pair[1] || pair[1]-pair[0] > 1 {
			t.Errorf("channel %d: texel * white shade gave %d, want %d or one less", i, pair[0], pair[1])
		}
	}

	in = combineInputs{Texel0: rgba{200, 100, 50, 255}, Shade: rgba{128, 128, 128, 255}}
	got = r.combine(&in)
	if got.R != 100 || got.G != 50 || got.B != 25 {
		t.Errorf("texel * half shade: got %+v want roughly half of each channel", got)
	}
}

func TestCombinerClampsRatherThanWrapping(t *testing.T) {
	// (1 - 0) * shade + 1 overflows when shade is white; the hardware clamps.
	r := &rdp{OtherModes: uint64(cycle1) << 52}
	r.Combine = combineWord(ccOne, ccZeroC, ccShade, ccOne, ccOne, ccZeroA, ccShade, ccOne)
	in := combineInputs{Shade: rgba{255, 255, 255, 255}}
	if got := r.combine(&in); got.R != 255 {
		t.Errorf("an overflowing combine gave %d, want 255", got.R)
	}
}

func TestSecondCycleSeesTheFirstThroughCOMBINED(t *testing.T) {
	// This is what makes multi-texturing possible: cycle 1 reads cycle 0's output.
	r := &rdp{OtherModes: uint64(cycle2) << 52}
	// Cycle 0: (0 - 0) * 0 + texel  ->  the texel, passed through by the adder.
	w := combineWord(ccZeroC, ccZeroC, ccZeroMul, ccTexel0, ccZeroA, ccZeroA, ccZeroA, 1)
	// Cycle 1: (0 - combined) * shade + prim  ->  prim minus the texel.
	w |= uint64(ccZeroC)<<37 | uint64(ccShade)<<32 | uint64(ccCombined)<<24 | uint64(ccPrim)<<6
	w |= uint64(ccZeroA)<<21 | uint64(ccZeroA)<<18 | uint64(ccZeroA)<<3 | uint64(3)
	r.Combine = w

	in := combineInputs{
		Texel0: rgba{60, 30, 10, 128},
		Shade:  rgba{255, 255, 255, 255},
		Prim:   rgba{200, 200, 200, 255},
	}
	got := r.combine(&in)
	// If the second cycle could not see the first, COMBINED would be zero and
	// the result would simply be prim.
	if got.R > 145 || got.R < 138 {
		t.Errorf("the second cycle did not see the first: got R=%d, want prim - texel (about 140)", got.R)
	}
	if got.R == 200 {
		t.Error("COMBINED read as zero: the cycles are not chained")
	}
}

func TestBlenderInterpolatesByAlpha(t *testing.T) {
	// The classic source-over: p = the pixel just shaded, m = the framebuffer,
	// a = its alpha, b = 1 - a. The division by (a + b) is what normalises it.
	r := &rdp{}
	r.OtherModes = uint64(cycle1)<<52 |
		0<<30 | 0<<26 | 1<<22 | 0<<18 // p=combined a=combined_alpha m=memory b=1-a

	src := rgba{200, 0, 0, 128}
	dst := rgba{0, 200, 0, 255}
	got := r.blend(src, dst, 255)
	// 128/255 of the source, the rest of the destination.
	if got.R < 95 || got.R > 105 || got.G < 95 || got.G > 105 {
		t.Errorf("half-alpha source over destination: got %+v want roughly (100,100,0)", got)
	}
}

func TestBlenderKnowsWhenItReadsMemory(t *testing.T) {
	// An opaque draw must not pay for a framebuffer read.
	r := &rdp{OtherModes: uint64(cycle1)<<52 | 0<<30 | 0<<26 | 0<<22 | 3<<18}
	if r.blenderReadsMemory() {
		t.Error("a blend selecting neither memory source claims to read memory")
	}
	r.OtherModes |= 1 << 22 // m = memory
	if !r.blenderReadsMemory() {
		t.Error("a blend whose M is the framebuffer does not report reading memory")
	}
}

func TestTexCoordWrapsAndMirrors(t *testing.T) {
	// A three-bit mask wraps every eight texels; mirroring reflects every other
	// repeat, which is how a game tiles a texture without a seam. The cm bits
	// are libultra's G_TX constants: 1 mirrors, 2 clamps — Pilotwings' sky is
	// the witness for which is which (see texCoord).
	for _, tc := range []struct {
		v        int32
		mask, cm uint32
		want     uint32
	}{
		{3, 3, 0, 3}, {8, 3, 0, 0}, {11, 3, 0, 3}, // wrap
		{3, 3, 1, 3}, {8, 3, 1, 7}, {11, 3, 1, 4}, // mirror
		{3, 3, 2, 3}, {8, 3, 2, 7}, {11, 3, 2, 7}, // clamp to the extent, 0..7
		{-2, 3, 2, 0}, {9, 3, 3, 7},               // clamp below; clamp wins before mirror
	} {
		if got := texCoord(tc.v, tc.mask, tc.cm, 0, 7); got != tc.want {
			t.Errorf("texCoord(v=%d mask=%d cm=%d) = %d want %d", tc.v, tc.mask, tc.cm, got, tc.want)
		}
	}
}
