// Package carmodel reimplements the game's in-race car object renderer — the
// 3-D vector car that represents the computer player. The car is not a stored
// vertex model: $599E2 builds it procedurally, in SCREEN space, from the four
// projected wheel points of the physics/AI rig:
//
//   - $5A186/$5A40C place the four wheel contact points on the decoded track
//     surface (front axle = the across interpolation of the rail samples,
//     rear axle = front displaced by 1.5x the axle vector rotated 90 deg,
//     $5A468); the drawn car is 81 plan units wide and 122 long on the
//     standard 384-unit road
//   - the height planes: body plane $1BD66 = surface + $68 (+ the bounce
//     rand), wheel-top plane $1BF64 = body + $50, deck plane = min(that,
//     $1BD86+$50) = surface + $50 ($599E2 head, $641B6, $63A58)
//   - $5A0CE/$59B7A derive halves/quarters/eighths of the projected axle
//     vector and its perpendicular; $59C7C/$59DF6/$59F78/$59F98/$59B1A
//     assemble the front/rear face rectangles, the four connector edges, the
//     four tyre parallelograms and the deck quad from those fractions
//   - $67AA6 fills the faces with hard-coded palette colours: tyres 0, deck
//     5, sides $C, front rectangle $A, top $F, bottom 9
//
// Build is a verbatim port of the screen-space construction (verified
// edge-exact against the engine by cmd/caroracle). Because the construction is
// linear in the projected wheel points, running it through an orthographic
// "projection" yields the exact 3-D shape the engine's math encodes; RestPose
// does that from a front and a side view to assemble the rest-pose model the
// webexport ships as a GLB.
package carmodel

// Pt is one projected (screen-space) point in the engine's convention:
// X right, Y down.
type Pt struct{ X, Y int }

// Edge is one display-list edge: two screen points captured at emit time,
// tagged with its display-list slot (offset/4 relative to the car's $5E0
// region base).
type Edge struct {
	Slot int
	A, B Pt
}

// BuildIn is the screen-space input of one $599E2 run: the four wheel points
// projected twice — Hi at the wheel-top plane ($1BF64 heights, slots
// $F4/$F6/$F8/$FA) and Lo at the deck plane (min-clamped, slots $118..$11E).
// Order: [0]=$F4, [1]=$F6, [2]=$F8, [3]=$FA (front-left, rear-left,
// front-right, rear-right in axle terms: $F4/$F8 form one axle, $F6/$FA the
// other).
type BuildIn struct {
	Hi [4]Pt
	Lo [4]Pt
}

// state mirrors the $1BC64..$1BC8A register family.
type state struct {
	bc64, bc66, bc68, bc6A int
	bc6C, bc6E, bc70, bc72 int
	bc74, bc76, bc78, bc7A int
	bc7C, bc7E, bc80, bc82 int
	bc84, bc86             int
	bc88, bc8A             int
}

func clamp255(v int) int {
	if v > 255 {
		return 255
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// axes reproduces $5A0CE + $59B7A: derive the axle-vector fraction family from
// the two projected wheel points a and b (the axle pair), and the axle screen
// midpoint.
func (s *state) axes(a, b Pt) {
	d4 := (b.X - a.X) >> 1
	d5 := (b.Y - a.Y) >> 1
	s.bc6A = clamp255(abs(d4))
	s.bc68 = clamp255(abs(d5))
	s.bc64 = s.bc6A
	if d4 >= 0 {
		s.bc6A = -s.bc6A
	} else {
		s.bc64 = -s.bc64
	}
	s.bc66 = s.bc68
	if d5 < 0 {
		s.bc66 = -s.bc66
		s.bc68 = -s.bc68
	}
	s.bc66 >>= 1
	s.bc6A >>= 1
	s.bc84 = a.X + d4
	s.bc86 = a.Y + d5

	// $59B7A: signed halves/quarters/eighths of each component
	frac := func(v int) (h, q, e int) {
		a := abs(v)
		h, q, e = a>>1, a>>2, a>>3
		if v < 0 {
			h, q, e = -h, -q, -e
		}
		return
	}
	s.bc6C, s.bc74, s.bc7C = frac(s.bc64)
	s.bc6E, s.bc76, s.bc7E = frac(s.bc66)
	s.bc70, s.bc78, s.bc80 = frac(s.bc68)
	s.bc72, s.bc7A, s.bc82 = frac(s.bc6A)
}

// wheelQuad reproduces the shared tail of $59F78/$59F98: build one tyre
// parallelogram from the current axle fractions and emit its four edges.
func (s *state) wheelQuad(first bool, slot int, out *[]Edge) {
	if first { // $59F78
		s.bc88 = -s.bc64
		s.bc8A = -s.bc68
	} else { // $59F98
		s.bc88 = s.bc64 - s.bc6C
		s.bc8A = s.bc68 - s.bc70
	}
	var v100, v102, v104, v106 Pt
	v102 = Pt{s.bc88, s.bc8A}
	v104 = Pt{s.bc88 + s.bc6C, s.bc8A + s.bc70}
	v100 = Pt{s.bc88 - s.bc66, s.bc8A - s.bc6A}
	v106 = Pt{s.bc88 - s.bc66 + s.bc6C, s.bc8A - s.bc6A + s.bc70}
	for _, p := range []*Pt{&v100, &v102, &v104, &v106} {
		p.X += s.bc84
		p.Y += s.bc86
	}
	v104.X++
	v106.X++
	v100.Y++
	v106.Y++
	*out = append(*out,
		Edge{slot + 0, v100, v102},
		Edge{slot + 1, v102, v104},
		Edge{slot + 2, v104, v106},
		Edge{slot + 3, v106, v100},
	)
}

// Build is the verbatim port of $599E2's construction over already-projected
// wheel points. It returns every display-list edge with its slot (the face
// fills of $67AA6 reference these slots; see Faces).
func Build(in BuildIn) []Edge {
	var out []Edge
	var s state

	v118 := in.Lo[0]
	v11A := in.Lo[1]
	v11C := in.Lo[2]
	v11E := in.Lo[3]

	// front axle: $5A0CE(d1=$F4) uses slots $F4/$F8
	s.axes(in.Hi[0], in.Hi[2])

	// $59C7C: two front tyre quads at slots $10/$14 (list offsets $40/$50)
	s.wheelQuad(true, 0x10, &out)
	s.wheelQuad(false, 0x14, &out)

	// front face rectangle at slots $4..$7 (offsets $10-$1C)
	var v108, v10A, v10C, v10E Pt
	v10E = Pt{s.bc6C - s.bc6E, s.bc70 - s.bc72}
	v108 = Pt{-s.bc6C - s.bc6E, -s.bc70 - s.bc72}
	v10C = Pt{s.bc66 + s.bc74 + s.bc7C, s.bc6A + s.bc78 + s.bc80}
	v10A = Pt{s.bc66 - s.bc74 - s.bc7C, s.bc6A - s.bc78 - s.bc80}
	for _, p := range []*Pt{&v108, &v10A, &v10C, &v10E} {
		p.X += s.bc84
		p.Y += s.bc86
	}
	v10C.X++
	v10E.X++
	v108.Y++
	v10E.Y++
	out = append(out,
		Edge{0x4, v108, v10A},
		Edge{0x5, v10A, v10C},
		Edge{0x6, v10C, v10E},
		Edge{0x7, v10E, v108},
	)

	// $59AB8: shift the front deck corners by the front-axle fractions
	v118.X -= s.bc66
	v11C.X -= s.bc66
	v118.Y -= s.bc6A - 1
	v11C.Y -= s.bc6A - 1

	// rear axle: $5A0CE(d1=$F6) uses slots $F6/$FA
	s.axes(in.Hi[1], in.Hi[3])

	// $59DF6: rear face rectangle at slots $8..$B (offsets $20-$2C)
	var v110, v112, v114, v116 Pt
	v114 = Pt{s.bc6C, s.bc70}
	v116 = Pt{s.bc6C - s.bc6E, s.bc70 - s.bc72}
	v112 = Pt{-s.bc6C, -s.bc70}
	v110 = Pt{-s.bc6C - s.bc6E, -s.bc70 - s.bc72}
	for _, p := range []*Pt{&v110, &v112, &v114, &v116} {
		p.X += s.bc84
		p.Y += s.bc86
	}
	v114.X++
	v116.X++
	v110.Y++
	v116.Y++
	out = append(out,
		Edge{0x8, v110, v112},
		Edge{0x9, v112, v114},
		Edge{0xA, v114, v116},
		Edge{0xB, v116, v110},
	)
	// connectors at slots $C..$F (offsets $30-$3C)
	out = append(out,
		Edge{0xC, v108, v110},
		Edge{0xD, v10E, v116},
		Edge{0xE, v10A, v112},
		Edge{0xF, v10C, v114},
	)
	// rear tyre quads at slots $18/$1C (offsets $60/$70)
	s.wheelQuad(true, 0x18, &out)
	s.wheelQuad(false, 0x1C, &out)

	// $59AEE: shift the rear deck corners by the rear-axle fractions (no
	// SUBQ here — only the front shift uses bc6A-1)
	v11A.X -= s.bc66
	v11E.X -= s.bc66
	v11A.Y -= s.bc6A
	v11E.Y -= s.bc6A

	// $59B1A: deck quad at slots $20..$23 (offsets $80-$8C)
	out = append(out,
		Edge{0x20, v118, v11C},
		Edge{0x21, v118, v11A},
		Edge{0x22, v11A, v11E},
		Edge{0x23, v11C, v11E},
	)
	return out
}

// Face is one filled face of the car: the display-list slots whose edges bound
// it (in $67AA6's read order) and the hard-coded palette colour.
type Face struct {
	Slots [4]int
	Pal   byte
}

// Faces is $67AA6's unrolled fill list (in draw order). Slot values match the
// Edge.Slot tags Build emits: offsets/4 relative to the car's list region.
var Faces = []Face{
	{[4]int{0x20, 0x21, 0x22, 0x23}, 5}, // deck plane (pitch-gated in-game)
	{[4]int{0x18, 0x19, 0x1A, 0x1B}, 0}, // rear tyre A
	{[4]int{0x1C, 0x1D, 0x1E, 0x1F}, 0}, // rear tyre B
	{[4]int{0x8, 0xE, 0x4, 0xC}, 0xC},   // left flank
	{[4]int{0x6, 0xF, 0xA, 0xD}, 0xC},   // right flank
	{[4]int{0x4, 0x5, 0x6, 0x7}, 0xA},   // front face rectangle (red)
	{[4]int{0xE, 0x9, 0xF, 0x5}, 0xF},   // top (white)
	{[4]int{0xC, 0x7, 0xD, 0xB}, 9},     // bottom (dark red)
	{[4]int{0x10, 0x11, 0x12, 0x13}, 0}, // front tyre A
	{[4]int{0x14, 0x15, 0x16, 0x17}, 0}, // front tyre B
}

// --- rest pose ---

// The car's world dimensions (track plan units; heights in the full track
// height domain where 4 height units = 1 plan unit):
const (
	Width     = 81  // axle track: road width 384 x the $1BBBE across offsets
	Length    = 122 // wheelbase: 1.5 x the axle vector ($5A468)
	HDeck     = 80  // deck plane above the surface ($1BD86 + $50)
	HTop      = 184 // wheel-top/body plane ($1BD66 + $50, bounce 0)
	HeightDiv = 4   // full height units per plan unit
)

// Vert3 is one rest-pose vertex in car-local plan units: X across (right
// positive), Y up, Z along (front positive).
type Vert3 struct{ X, Y, Z float64 }

// RestModel is the 3-D rest pose: deduplicated vertices plus the faces (quads
// of vertex indexes, engine fill order) and their palette colours.
type RestModel struct {
	Verts []Vert3
	Quads []struct {
		V   [4]int
		Pal byte
	}
}

// ortho runs Build over an orthographic projection of the rest-pose wheels:
// axis selects the horizontal coordinate (0 = across/X, 1 = along/Z); the
// vertical is -height/HeightDiv (screen Y grows down).
func ortho(axis int) []Edge {
	// wheel order [ $F4, $F6, $F8, $FA ] = front-left, rear-left, front-right,
	// rear-right: $F4/$F8 = the front axle, $F6/$FA = the rear axle
	wx := [4]float64{-Width / 2.0, -Width / 2.0, Width / 2.0, Width / 2.0}
	wz := [4]float64{Length / 2.0, -Length / 2.0, Length / 2.0, -Length / 2.0}
	var in BuildIn
	for i := 0; i < 4; i++ {
		h := wx[i]
		if axis == 1 {
			h = wz[i]
		}
		// scale x2 keeps sub-unit precision through the >>3 fractions while
		// staying inside the $FF axle clamp ($5A0CE) that is part of the
		// algorithm
		in.Hi[i] = Pt{int(h * 2), -HTop * 2 / HeightDiv}
		in.Lo[i] = Pt{int(h * 2), -HDeck * 2 / HeightDiv}
	}
	return Build(in)
}

// Rest assembles the 3-D rest pose by combining the across (front) and along
// (side) orthographic runs of the verbatim construction: each display-list
// slot's endpoints give X from the front view, Z from the side view and Y from
// both (they agree by construction for a level pose).
func Rest() RestModel {
	front := ortho(0)
	side := ortho(1)
	bySlotF := map[int]Edge{}
	bySlotS := map[int]Edge{}
	for _, e := range front {
		bySlotF[e.Slot] = e
	}
	for _, e := range side {
		bySlotS[e.Slot] = e
	}

	var m RestModel
	index := map[[3]int]int{}
	vert := func(f, s Pt) int {
		key := [3]int{f.X, s.X, f.Y}
		if i, ok := index[key]; ok {
			return i
		}
		i := len(m.Verts)
		index[key] = i
		m.Verts = append(m.Verts, Vert3{
			X: float64(f.X) / 2,
			Z: float64(s.X) / 2,
			Y: -float64(f.Y) / 2, // screen Y down -> world up (plan units)
		})
		return i
	}
	type corner struct{ f, s Pt }
	key := func(c corner) [4]int { return [4]int{c.f.X, c.f.Y, c.s.X, c.s.Y} }
	for _, face := range Faces {
		// chain the four boundary edges into a ring: corners correspond
		// across the two views by construction (same code path, same roles)
		type ce struct{ a, b corner }
		var es []ce
		for _, slot := range face.Slots {
			ef, e2 := bySlotF[slot], bySlotS[slot]
			es = append(es, ce{corner{ef.A, e2.A}, corner{ef.B, e2.B}})
		}
		ring := []corner{es[0].a, es[0].b}
		used := map[int]bool{0: true}
		for len(ring) < 4 {
			cur := ring[len(ring)-1]
			found := false
			for i, e := range es {
				if used[i] {
					continue
				}
				if key(e.a) == key(cur) {
					ring = append(ring, e.b)
					used[i] = true
					found = true
					break
				}
				if key(e.b) == key(cur) {
					ring = append(ring, e.a)
					used[i] = true
					found = true
					break
				}
			}
			if !found {
				// degenerate (coincident corners): fall back to A endpoints
				ring = []corner{es[0].a, es[1].a, es[2].a, es[3].a}
				break
			}
		}
		var q [4]int
		for i, c := range ring[:4] {
			q[i] = vert(c.f, c.s)
		}
		m.Quads = append(m.Quads, struct {
			V   [4]int
			Pal byte
		}{q, face.Pal})
	}
	return m
}
