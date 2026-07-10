package main

import (
	"math"
	"testing"
)

// A game-space rotation about the up axis (game Z) must, in glTF space, be a
// rotation about glTF's up axis (Y) by the same angle — the basis change carries
// the up vector Z -> Y and must carry a rotation about it likewise. This pins the
// quaternion transform's handedness, the one bit that a wrong sign would flip.
func TestGameZRotationBecomesGLTFYRotation(t *testing.T) {
	// 90 degrees about game Z: (0, 0, sin45, cos45).
	s := float32(math.Sqrt2 / 2)
	q := toGLQuat(0, 0, s, s)

	// Expect a rotation about +Y: (0, ±sin45, 0, cos45). The axis (x,z) must be
	// zero and w must be cos(45).
	if math.Abs(float64(q[0])) > 1e-5 || math.Abs(float64(q[2])) > 1e-5 {
		t.Errorf("axis leaked off Y: q = %v", q)
	}
	if math.Abs(float64(q[1])-float64(s)) > 1e-5 && math.Abs(float64(q[1])+float64(s)) > 1e-5 {
		t.Errorf("Y component %v, want ±%v", q[1], s)
	}
	if math.Abs(float64(q[3])-float64(s)) > 1e-5 {
		t.Errorf("w %v, want %v", q[3], s)
	}
}

// The transform preserves unit norm (b and its conjugate are unit quaternions),
// so an exported keyframe is still a valid rotation.
func TestToGLQuatStaysUnit(t *testing.T) {
	for _, in := range [][4]float32{{0.1, -0.2, 0.3, 0.9273618}, {0, 0, 0, 1}, {-0.5, 0.5, -0.5, 0.5}} {
		q := toGLQuat(in[0], in[1], in[2], in[3])
		n := math.Sqrt(float64(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]))
		if math.Abs(n-1) > 1e-5 {
			t.Errorf("norm %v for input %v", n, in)
		}
	}
}
