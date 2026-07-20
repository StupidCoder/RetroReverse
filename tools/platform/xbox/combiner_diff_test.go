package xbox

// combiner_diff_test.go guards the register-combiner optimisation the way the repo guards every
// reimplementation: against an independent reference. refCombine below is a FROZEN copy of the
// combiner as it stood before the array-register-file rewrite — its own types, its own switch-
// based register file, byte-for-byte the pre-optimisation arithmetic. TestCombinerMatchesRef runs
// thousands of random combiner configurations and fragment inputs through both the live combine()
// and refCombine() and asserts the [4]float32 outputs are bit-identical (float bits, so a NaN that
// compares unequal still fails). This is stronger than the an3-drive frame gate, which exercises
// only OutRun's single MODULATE stage: a config the game never programs but the hardware allows
// would slip past the gate and be caught here ([[reference-match-hides-bugs]]).

import (
	"math"
	"testing"
)

// --- the frozen reference: the combiner before the rewrite, self-contained ---

type refRegs struct {
	c0, c1     [4]float32
	fog        [4]float32
	col0, col1 [4]float32
	tex        [4][4]float32
	spare      [2][4]float32
}

func (r *refRegs) read(reg uint32) [4]float32 {
	switch reg {
	case 1:
		return r.c0
	case 2:
		return r.c1
	case 3:
		return r.fog
	case 4:
		return r.col0
	case 5:
		return r.col1
	case 8, 9, 10, 11:
		return r.tex[reg-8]
	case 12:
		return r.spare[0]
	case 13:
		return r.spare[1]
	}
	return [4]float32{}
}

func (r *refRegs) write(reg uint32, v [4]float32, alpha bool) {
	var dst *[4]float32
	switch reg {
	case 4:
		dst = &r.col0
	case 5:
		dst = &r.col1
	case 8, 9, 10, 11:
		dst = &r.tex[reg-8]
	case 12:
		dst = &r.spare[0]
	case 13:
		dst = &r.spare[1]
	default:
		return
	}
	if alpha {
		dst[3] = v[3]
	} else {
		dst[0], dst[1], dst[2] = v[0], v[1], v[2]
	}
}

func refMap(v [4]float32, mapping uint32) [4]float32 {
	var out [4]float32
	switch mapping {
	case 0:
		for i, x := range v {
			out[i] = clamp01(x)
		}
	case 1:
		for i, x := range v {
			out[i] = 1 - clamp01(x)
		}
	case 2:
		for i, x := range v {
			out[i] = 2*clamp01(x) - 1
		}
	case 3:
		for i, x := range v {
			out[i] = -(2*clamp01(x) - 1)
		}
	case 4:
		for i, x := range v {
			out[i] = clamp01(x) - 0.5
		}
	case 5:
		for i, x := range v {
			out[i] = -(clamp01(x) - 0.5)
		}
	case 6:
		out = v
	case 7:
		for i, x := range v {
			out[i] = -x
		}
	}
	return out
}

func refFetch(r *refRegs, b uint32, alphaSide bool) [4]float32 {
	v := r.read(b & 0xF)
	if b>>4&1 == 1 {
		v = [4]float32{v[3], v[3], v[3], v[3]}
	} else if alphaSide {
		v = [4]float32{v[2], v[2], v[2], v[2]}
	}
	return refMap(v, b>>5&7)
}

func refOp(v [4]float32, op uint32) [4]float32 {
	switch op {
	case 1:
		for i, x := range v {
			v[i] = clampSigned(x - 0.5)
		}
	case 2:
		for i, x := range v {
			v[i] = clampSigned(x * 2)
		}
	case 3:
		for i, x := range v {
			v[i] = clampSigned((x - 0.5) * 2)
		}
	case 4:
		for i, x := range v {
			v[i] = clampSigned(x * 4)
		}
	case 6:
		for i, x := range v {
			v[i] = clampSigned(x * 0.5)
		}
	default:
		for i, x := range v {
			v[i] = clampSigned(x)
		}
	}
	return v
}

func refFinalFetch(r *refRegs, b uint32, ef [4]float32) [4]float32 {
	var v [4]float32
	switch b & 0xF {
	case 14:
		for i := range v {
			v[i] = clamp01(r.spare[0][i] + r.col1[i])
		}
	case 15:
		v = ef
	default:
		v = r.read(b & 0xF)
	}
	if b>>4&1 == 1 {
		v = [4]float32{v[3], v[3], v[3], v[3]}
	}
	if b>>5&7 == 1 {
		return [4]float32{1 - clamp01(v[0]), 1 - clamp01(v[1]), 1 - clamp01(v[2]), 1 - clamp01(v[3])}
	}
	return [4]float32{clamp01(v[0]), clamp01(v[1]), clamp01(v[2]), clamp01(v[3])}
}

func refCombine(cs *combState, in *combInput) [4]float32 {
	r := refRegs{fog: cs.fogColor, col0: in.col0, col1: in.col1, tex: in.tex}
	for si := range cs.stages {
		st := &cs.stages[si]
		r.c0 = st.factor0
		r.c1 = st.factor1
		a := refFetch(&r, st.colorICW>>24&0xFF, false)
		b := refFetch(&r, st.colorICW>>16&0xFF, false)
		c := refFetch(&r, st.colorICW>>8&0xFF, false)
		d := refFetch(&r, st.colorICW&0xFF, false)
		ab := mul4(a, b)
		cd := mul4(c, d)
		if st.colorOCW>>13&1 == 1 {
			dp := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
			ab = [4]float32{dp, dp, dp, dp}
		}
		if st.colorOCW>>12&1 == 1 {
			dp := c[0]*d[0] + c[1]*d[1] + c[2]*d[2]
			cd = [4]float32{dp, dp, dp, dp}
		}
		var sum [4]float32
		if st.colorOCW>>14&1 == 1 {
			if r.spare[0][3] < 0.5 {
				sum = ab
			} else {
				sum = cd
			}
		} else {
			sum = add4(ab, cd)
		}
		op := st.colorOCW >> 15 & 7
		ab, cd, sum = refOp(ab, op), refOp(cd, op), refOp(sum, op)
		r.write(st.colorOCW>>4&0xF, ab, false)
		r.write(st.colorOCW&0xF, cd, false)
		r.write(st.colorOCW>>8&0xF, sum, false)

		aa := refFetch(&r, st.alphaICW>>24&0xFF, true)
		ba := refFetch(&r, st.alphaICW>>16&0xFF, true)
		ca := refFetch(&r, st.alphaICW>>8&0xFF, true)
		da := refFetch(&r, st.alphaICW&0xFF, true)
		abA := [4]float32{0, 0, 0, aa[3] * ba[3]}
		cdA := [4]float32{0, 0, 0, ca[3] * da[3]}
		var sumA [4]float32
		if st.alphaOCW>>14&1 == 1 {
			if r.spare[0][3] < 0.5 {
				sumA = abA
			} else {
				sumA = cdA
			}
		} else {
			sumA = [4]float32{0, 0, 0, abA[3] + cdA[3]}
		}
		opA := st.alphaOCW >> 15 & 7
		abA, cdA, sumA = refOp(abA, opA), refOp(cdA, opA), refOp(sumA, opA)
		r.write(st.alphaOCW>>4&0xF, abA, true)
		r.write(st.alphaOCW&0xF, cdA, true)
		r.write(st.alphaOCW>>8&0xF, sumA, true)
	}
	e := refFinalFetch(&r, cs.cw1>>24&0xFF, [4]float32{})
	f := refFinalFetch(&r, cs.cw1>>16&0xFF, [4]float32{})
	ef := mul4(e, f)
	a := refFinalFetch(&r, cs.cw0>>24&0xFF, ef)
	b := refFinalFetch(&r, cs.cw0>>16&0xFF, ef)
	c := refFinalFetch(&r, cs.cw0>>8&0xFF, ef)
	d := refFinalFetch(&r, cs.cw0&0xFF, ef)
	gv := refFinalFetch(&r, cs.cw1>>8&0xFF, ef)
	var out [4]float32
	for i := 0; i < 3; i++ {
		out[i] = clamp01(a[i]*b[i] + (1-a[i])*c[i] + d[i])
	}
	out[3] = clamp01(gv[3])
	return out
}

// --- the differential test ---

// lcg is a tiny deterministic PRNG (Math.random() is unavailable and we want reproducibility).
type lcg uint64

func (s *lcg) next() uint32 {
	*s = *s*6364136223846793005 + 1442695040888963407
	return uint32(*s >> 32)
}
func (s *lcg) f32() float32 {
	// values in [-2,2], so mappings/ops that expand out of [0,1] are exercised
	return (float32(s.next()>>8)/float32(1<<24))*4 - 2
}

func TestCombinerMatchesRef(t *testing.T) {
	g := &pgraph{}
	rng := lcg(0x1234567)
	for iter := 0; iter < 20000; iter++ {
		// A random config: 1-4 stages, random ICW/OCW/final control words and factors.
		var cs combState
		n := int(rng.next()%4) + 1
		cs.stages = cs.buf[:n]
		for i := 0; i < n; i++ {
			cs.stages[i] = combStage{
				colorICW: rng.next(),
				alphaICW: rng.next(),
				colorOCW: rng.next(),
				alphaOCW: rng.next(),
				factor0:  [4]float32{rng.f32(), rng.f32(), rng.f32(), rng.f32()},
				factor1:  [4]float32{rng.f32(), rng.f32(), rng.f32(), rng.f32()},
			}
		}
		cs.cw0 = rng.next()
		cs.cw1 = rng.next()
		cs.fogColor = [4]float32{rng.f32(), rng.f32(), rng.f32(), 1}

		var in combInput
		in.col0 = [4]float32{rng.f32(), rng.f32(), rng.f32(), rng.f32()}
		in.col1 = [4]float32{rng.f32(), rng.f32(), rng.f32(), rng.f32()}
		for u := 0; u < 4; u++ {
			in.tex[u] = [4]float32{rng.f32(), rng.f32(), rng.f32(), rng.f32()}
		}

		got := g.combine(&cs, &in)
		want := refCombine(&cs, &in)
		for c := 0; c < 4; c++ {
			if math.Float32bits(got[c]) != math.Float32bits(want[c]) {
				t.Fatalf("iter %d comp %d: combine=%v (%08x) ref=%v (%08x)\n  cs=%+v in=%+v",
					iter, c, got[c], math.Float32bits(got[c]), want[c], math.Float32bits(want[c]), cs, in)
			}
		}
	}
}
