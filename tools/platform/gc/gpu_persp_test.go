package gc

import (
	"math"
	"testing"
)

// TestPerspUVOrthoReducesToAffine pins the property that keeps every 2D screen in the game
// exactly as it was: an orthographic draw has w = 1 at every vertex, so perspective correction
// must be an identity on it. The title screen is rendered entirely this way, and it stayed
// bit-identical across this change because of this.
func TestPerspUVOrthoReducesToAffine(t *testing.T) {
	v0 := sv(0, 0, 1)
	v1 := sv(1, 0, 1)
	v2 := sv(0, 1, 1)

	for _, b := range [][3]float32{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1},
		{0.5, 0.5, 0}, {1.0 / 3, 1.0 / 3, 1.0 / 3}, {0.25, 0.25, 0.5},
	} {
		u, v := puv(b[0], b[1], b[2], v0, v1, v2)
		wantU := b[0]*v0.tc[0].s + b[1]*v1.tc[0].s + b[2]*v2.tc[0].s
		wantV := b[0]*v0.tc[0].t + b[1]*v1.tc[0].t + b[2]*v2.tc[0].t
		if u != wantU || v != wantV {
			t.Errorf("weights %v: got (%v,%v), want the affine (%v,%v)", b, u, v, wantU, wantV)
		}
	}
}

// TestPerspUVAtVertices checks the boundary: at a vertex the interpolation must return that
// vertex's own coordinate exactly, whatever the w disparity.
func TestPerspUVAtVertices(t *testing.T) {
	v0 := sv(0.1, 0.2, 1)
	v1 := sv(0.7, 0.9, 1.0 / 50)
	v2 := sv(0.4, 0.3, 1.0 / 1000)

	for i, c := range []struct {
		b  [3]float32
		vx screenVertex
	}{
		{[3]float32{1, 0, 0}, v0},
		{[3]float32{0, 1, 0}, v1},
		{[3]float32{0, 0, 1}, v2},
	} {
		u, v := puv(c.b[0], c.b[1], c.b[2], v0, v1, v2)
		if math.Abs(float64(u-c.vx.tc[0].s)) > 1e-6 || math.Abs(float64(v-c.vx.tc[0].t)) > 1e-6 {
			t.Errorf("vertex %d: got (%v,%v), want (%v,%v)", i, u, v, c.vx.tc[0].s, c.vx.tc[0].t)
		}
	}
}

// TestPerspUVRecedingEdge is the case affine interpolation gets wrong, with the arithmetic
// worked out by hand so the test states the right answer rather than echoing the code.
//
// Two vertices of a receding surface: one at w = 1 with u = 0, one at w = 3 with u = 1. At the
// halfway point across the SCREEN, the affine answer is u = 0.5. The correct answer weights by
// 1/w: interpolating u/w gives 0.5*0 + 0.5*(1/3) = 1/6, and interpolating 1/w gives
// 0.5*1 + 0.5*(1/3) = 2/3, so u = (1/6)/(2/3) = 0.25. The texture is a quarter of the way
// along, not half — the far half of the surface is compressed, which is what perspective does.
func TestPerspUVRecedingEdge(t *testing.T) {
	near := sv(0, 0, 1.0 / 1)
	far := sv(1, 0, 1.0 / 3)
	unused := sv(0, 0, 1)

	u, _ := puv(0.5, 0.5, 0, near, far, unused)
	if math.Abs(float64(u-0.25)) > 1e-6 {
		t.Errorf("screen midpoint of a receding edge: u = %v, want 0.25", u)
	}
	// And the affine answer — the bug — must not be what comes out.
	if math.Abs(float64(u-0.5)) < 1e-6 {
		t.Error("u = 0.5: the interpolation is still affine")
	}
}

// TestPerspUVMatchesGroundTruthAlongAnEdge sweeps a receding edge and checks the interpolation
// against the independent closed form: at screen parameter s between two vertices, the true
// surface parameter is t = s/w1 / ((1-s)/w0 + s/w1), and u is the linear interpolation of the
// endpoints at t. Deriving the expected value a different way than the implementation does is
// what makes this a check rather than a restatement.
func TestPerspUVMatchesGroundTruthAlongAnEdge(t *testing.T) {
	const w0, w1 float32 = 1, 100
	v0 := sv(0, 3, 1 / w0)
	v1 := sv(1, 7, 1 / w1)
	v2 := sv(9, 9, 1 / w0)

	for i := 0; i <= 10; i++ {
		s := float32(i) / 10
		gotU, gotV := puv(1-s, s, 0, v0, v1, v2)

		num := s / w1
		den := (1-s)/w0 + s/w1
		tt := num / den
		wantU := v0.tc[0].s + (v1.tc[0].s-v0.tc[0].s)*tt
		wantV := v0.tc[0].t + (v1.tc[0].t-v0.tc[0].t)*tt

		if math.Abs(float64(gotU-wantU)) > 1e-5 || math.Abs(float64(gotV-wantV)) > 1e-5 {
			t.Errorf("s=%.1f: got (%v,%v), want (%v,%v)", s, gotU, gotV, wantU, wantV)
		}
	}
}

// sv builds a one-coordinate screen vertex for the interpolation tests: the coordinate is
// non-projective, so its divisor is the 1 the transform unit would have written (see
// gpu_texgen.go), and the perspective divide by it is an identity.
func sv(u, v, invW float32) screenVertex {
	s := screenVertex{ntc: 1, invW: invW}
	s.tc[0] = texCoord{s: u, t: v, q: 1}
	return s
}

// puv is the two-coordinate answer the old single-coordinate interpolator returned, so these
// tests keep asserting what they always asserted.
func puv(b0, b1, b2 float32, v0, v1, v2 screenVertex) (u, v float32) {
	var out [maxTexCoord]texCoord
	perspTexCoords(b0, b1, b2, v0, v1, v2, &out)
	return out[0].s, out[0].t
}
