package glb

import "testing"

// A unit quad at height y, wound so its normal comes out +Y, as two triangles.
func upQuad(y float32) ([][3]float32, [][3]uint32) {
	p := [][3]float32{{0, y, 0}, {0, y, 1}, {1, y, 1}, {1, y, 0}}
	return p, [][3]uint32{{0, 1, 2}, {0, 2, 3}}
}

func TestOrientUpFlipsOnlyWhatItShould(t *testing.T) {
	p, tris := upQuad(0)
	if n := OrientUp(p, tris, 0.707); n != 0 {
		t.Fatalf("already up-facing, flipped %d", n)
	}
	up, down, side := FacingCounts(p, tris, 0.707)
	if up != 2 || down != 0 || side != 0 {
		t.Fatalf("up=%d down=%d side=%d", up, down, side)
	}

	// Reverse them and it must put them back — bit-for-bit the original order,
	// since a reversal is its own inverse for the middle-and-last swap.
	tris[0] = [3]uint32{0, 2, 1}
	tris[1] = [3]uint32{0, 3, 2}
	if n := OrientUp(p, tris, 0.707); n != 2 {
		t.Fatalf("flipped %d, want 2", n)
	}
	if tris[0] != ([3]uint32{0, 1, 2}) || tris[1] != ([3]uint32{0, 2, 3}) {
		t.Fatalf("not restored: %v", tris)
	}
}

func TestOrientUpLeavesWallsAlone(t *testing.T) {
	// A vertical quad: no defensible answer, so it must come out untouched
	// whichever way it was wound.
	p := [][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}
	for _, want := range [][][3]uint32{
		{{0, 1, 2}, {0, 2, 3}},
		{{0, 2, 1}, {0, 3, 2}},
	} {
		tris := append([][3]uint32(nil), want...)
		if n := OrientUp(p, tris, 0.707); n != 0 {
			t.Fatalf("flipped %d vertical triangles", n)
		}
		for i := range tris {
			if tris[i] != want[i] {
				t.Fatalf("vertical triangle rewritten: %v -> %v", want, tris)
			}
		}
		if up, down, side := FacingCounts(p, tris, 0.707); up != 0 || down != 0 || side != 2 {
			t.Fatalf("up=%d down=%d side=%d, want all side", up, down, side)
		}
	}
}

func TestOrientUpKeepsTheSurfaceItself(t *testing.T) {
	// The whole safety argument: reversing a triangle changes which way it
	// FACES and nothing else. Same three corners, same positions, same area.
	p, tris := upQuad(3)
	tris[0] = [3]uint32{0, 2, 1}
	before := map[uint32]bool{tris[0][0]: true, tris[0][1]: true, tris[0][2]: true}
	OrientUp(p, tris, 0.707)
	after := map[uint32]bool{tris[0][0]: true, tris[0][1]: true, tris[0][2]: true}
	if len(before) != len(after) {
		t.Fatal("corner count changed")
	}
	for k := range before {
		if !after[k] {
			t.Fatalf("corner %d lost; %v", k, tris[0])
		}
	}
}

func TestOrientUpSurvivesDegenerates(t *testing.T) {
	// Zero-area triangles have no normal. They must not be reversed, and must
	// not divide by zero either.
	p := [][3]float32{{0, 0, 0}, {1, 0, 0}, {2, 0, 0}}
	tris := [][3]uint32{{0, 1, 2}}
	if n := OrientUp(p, tris, 0.707); n != 0 {
		t.Fatalf("flipped a degenerate")
	}
	if up, down, side := FacingCounts(p, tris, 0.707); up+down+side != 0 {
		t.Fatalf("degenerate counted: %d %d %d", up, down, side)
	}
}

func TestOrientUpThresholdIsTheSlope(t *testing.T) {
	// A ramp at 60 degrees is "ground" at a 70-degree threshold and "wall" at
	// a 45-degree one, and the flip must follow that and not the other way.
	// A ramp whose normal is 60 degrees off vertical, wound so it faces DOWN
	// (normal.y = -0.5) — otherwise there is nothing for either threshold to do.
	p := [][3]float32{{0, 0, 0}, {1, 0, 0}, {0.5, 1.732, 1}}
	mk := func() [][3]uint32 { return [][3]uint32{{0, 1, 2}} }

	tris := mk()
	if n := OrientUp(p, tris, 0.9397); n != 0 { // 20 deg max slope: this is a wall
		t.Fatalf("steep threshold flipped %d", n)
	}
	tris = mk()
	if n := OrientUp(p, tris, 0.342); n != 1 { // 70 deg max slope: this is ground
		t.Fatalf("shallow threshold flipped %d, want 1", n)
	}
}
