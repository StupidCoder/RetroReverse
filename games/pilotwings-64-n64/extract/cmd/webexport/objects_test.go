package main

import (
	"math"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
)

// applyGL transforms a glTF-space point by a node matrix in the layout the
// exporter writes: the row-vector matrix listed row by row, which is how glTF
// stores a column-vector matrix column-major.
func applyGL(m []float32, p [3]float32) [3]float32 {
	var out [3]float32
	for j := 0; j < 3; j++ {
		out[j] = m[0*4+j]*p[0] + m[1*4+j]*p[1] + m[2*4+j]*p[2] + m[3*4+j]
	}
	return out
}

func closeTo(a, b [3]float32, eps float32) bool {
	for i := range a {
		if float32(math.Abs(float64(a[i]-b[i]))) > eps {
			return false
		}
	}
	return true
}

// The node matrix must transform the *already-rotated* model. Take a point in
// game space, push it through the game's own transforms and rotate the result
// to Y-up; then rotate the point first and push it through the node matrix. The
// two must agree — which is only true if the matrix is R^-1 * M * R, and not M.
func TestWorldMatrixTransformsTheRotatedModel(t *testing.T) {
	// A pose with a scale, a yaw and a translation, and a cell that offsets it:
	// all three of the things a naive composition gets wrong in a different way.
	s, c := float32(0.1)*0.6, float32(0.1)*0.8 // scale 0.1, yaw with sin/cos 0.6/0.8
	pose := uvmd.Matrix{
		{c, s, 0, 0},
		{-s, c, 0, 0},
		{0, 0, 0.1, 0},
		{-154, -176, 5, 1},
	}
	cell := uvmd.Matrix{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{-300, -300, 0, 1},
	}
	got := worldMatrix(pose, cell)

	for _, p := range [][3]float32{{0, 0, 0}, {100, 0, 0}, {0, 100, 0}, {0, 0, 100}, {143, -20, 7}} {
		// The game's own answer, then rotated to glTF space.
		m := mul(pose, cell)
		var w [3]float32
		for j := 0; j < 3; j++ {
			w[j] = m[0][j]*p[0] + m[1][j]*p[1] + m[2][j]*p[2] + m[3][j]
		}
		want := toGL(w[0], w[1], w[2])

		// The viewer's answer: rotate the model's vertex, then apply the node matrix.
		gl := toGL(p[0], p[1], p[2])
		have := applyGL(got, gl)

		if !closeTo(have, want, 1e-3) {
			t.Errorf("point %v: node matrix gives %v, the game gives %v", p, have, want)
		}
	}
}

// The translation is the object's position, rotated — nothing more. This is the
// invariant the ROM-side containment test leans on, so it is worth stating
// separately from the general case above.
func TestWorldMatrixTranslationIsTheRotatedPosition(t *testing.T) {
	pose := uvmd.Matrix{{0.1, 0, 0, 0}, {0, 0.1, 0, 0}, {0, 0, 0.1, 0}, {-126, -206, 5, 1}}
	cell := uvmd.Matrix{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {-300, 300, 0, 1}}
	m := worldMatrix(pose, cell)
	got := [3]float32{m[12], m[13], m[14]}
	want := toGL(-426, 94, 5)
	if !closeTo(got, want, 1e-3) {
		t.Errorf("translation %v, want %v", got, want)
	}
}
