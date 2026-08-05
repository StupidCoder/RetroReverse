package glb

import "math"

// orient.go — deriving a winding for geometry whose source never had one.
//
// glTF says a triangle's front face is the one wound counter-clockwise, and
// consumers that care — a backface cull, a walkable-surface index, a normal
// generator — read that and nothing else. Several of the machines we export
// from never had the concept. The 3DO cel engine, the PSX GPU as Ridge Racer
// drives it and Crazy Taxi's renderer all paint quads back to front with no
// culling whatsoever, so the corner order in their face lists was free to be
// anything, and measurably is: 48% of the near-horizontal triangles in a Need
// for Speed course come out facing down, 83% in Crazy Taxi, 94% in Ridge Racer.
//
// That is not a decode we got wrong. The disc does not contain the answer. But
// glTF requires SOME answer, and "whatever the face list happened to say" is
// the one answer that is guaranteed to be meaningless — so the winding these
// exporters emit is DERIVED, and this is where it is derived.
//
// The rule is deliberately narrow: a triangle that isnear horizontal has an
// obvious right answer (ground faces up), and a triangle that is near vertical
// does not (a wall faces whichever way the room is, which the geometry alone
// cannot say). So OrientUp fixes the first kind and leaves the second alone
// rather than inventing a convention it cannot defend.
//
// This is safe to apply to double-sided material groups, which is what all
// three of those games emit — the render cannot change, because neither face
// was ever culled.

// OrientUp reverses every triangle whose geometric normal points downward,
// among those steeper than cosMin from horizontal. Triangles nearer to vertical
// than that are left exactly as they were. Returns how many were reversed.
//
// cosMin is the cosine of the maximum slope still treated as ground: 0.707 is
// 45 degrees, which matches what a walkable-surface index asks for.
func OrientUp(positions [][3]float32, tris [][3]uint32, cosMin float32) int {
	flipped := 0
	for i, t := range tris {
		a, b, c := positions[t[0]], positions[t[1]], positions[t[2]]
		ux, uy, uz := b[0]-a[0], b[1]-a[1], b[2]-a[2]
		vx, vy, vz := c[0]-a[0], c[1]-a[1], c[2]-a[2]
		nx := uy*vz - uz*vy
		ny := uz*vx - ux*vz
		nz := ux*vy - uy*vx
		l := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if l == 0 {
			continue // degenerate: no normal, nothing to orient
		}
		y := ny / l
		if y >= -cosMin {
			// Facing up, or too close to vertical to have an opinion. The
			// boundary itself counts as "no opinion" — see FacingCounts.
			continue
		}
		tris[i] = [3]uint32{t[0], t[2], t[1]}
		flipped++
	}
	return flipped
}

// FacingCounts reports how many triangles face up, down, or neither, at the
// same threshold. Exporters print it so a re-export SAYS what it changed
// instead of leaving the reader to open the file.
func FacingCounts(positions [][3]float32, tris [][3]uint32, cosMin float32) (up, down, side int) {
	for _, t := range tris {
		a, b, c := positions[t[0]], positions[t[1]], positions[t[2]]
		ux, uy, uz := b[0]-a[0], b[1]-a[1], b[2]-a[2]
		vx, vy, vz := c[0]-a[0], c[1]-a[1], c[2]-a[2]
		nx := uy*vz - uz*vy
		ny := uz*vx - ux*vz
		nz := ux*vy - uy*vx
		l := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if l == 0 {
			continue
		}
		// The same predicate OrientUp uses, so the two cannot disagree: a
		// triangle sitting EXACTLY on the threshold was both "not steep enough
		// to flip" and "down-facing", and the report said so while the file
		// said otherwise.
		switch y := ny / l; {
		case y > cosMin:
			up++
		case y < -cosMin:
			down++
		default:
			side++
		}
	}
	return
}
