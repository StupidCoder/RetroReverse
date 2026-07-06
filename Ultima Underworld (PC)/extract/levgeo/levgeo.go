// Package levgeo turns a decoded Ultima Underworld tile map into 3D polygons —
// the static level geometry — following the tile-geometry rules reverse-
// engineered in Part V:
//
//   - solid tiles are rock (walls come from their open neighbours);
//   - open tiles get a flat floor quad;
//   - slopes (types 6-9) tilt two corners up by one height unit;
//   - diagonals (types 2-5) are half-open: the solid corner (NW/NE/SW/SE for
//     2/3/4/5, derived from neighbour solidity in the real levels) is cut off by
//     a diagonal wall, leaving a triangular floor.
//
// A wall quad is emitted on an edge wherever the neighbour is solid (up to the
// ceiling) or its floor is higher (a step up). Floors take the tile's floor
// texture, walls its wall texture, mapped to global W64.TR/F32.TR numbers
// through the per-level texture list.
//
// One tile is a 1x1 cell in XY; heights are scaled by HeightScale because the
// game's vertex transform (07F7:50BE) halves Z relative to X/Y (SAR CX,1 after a
// shared doubling), so a height unit is half a tile width. Coordinates: +X east,
// +Y north, +Z up.
package levgeo

import "ultimaunderworld/extract/lev"

// Ceiling is the level-wide ceiling height (tile-height units) walls rise to.
const Ceiling = 16

// HeightScale converts a tile-height unit into world units where a tile is 1x1.
// The game's vertex transform scales Z to half of X/Y, so one height unit is
// half a tile.
const HeightScale = 0.5

const ceilingZ = Ceiling * HeightScale

// Quad is one textured polygon. For a triangle, Tri is set and P[3]/UV[3] are
// unused. P holds the corners (CCW seen from the front); UV the texture coords.
type Quad struct {
	P    [4][3]float32
	UV   [4][2]float32
	Tex  uint16 // global texture number
	Wall bool   // W64.TR wall texture (vs F32.TR floor)
	Tri  bool   // triangle (3 verts) rather than quad (4)
}

// Mesh is the whole level's geometry.
type Mesh struct{ Quads []Quad }

// corner (cx,cy in {0,1}) world position and (possibly sloped) height.
func cornerHeight(t, h uint8, cx, cy int) float32 {
	z := float32(h)
	switch t {
	case lev.TileSlopeN: // rises toward north (+Y)
		if cy == 1 {
			z++
		}
	case lev.TileSlopeS:
		if cy == 0 {
			z++
		}
	case lev.TileSlopeE:
		if cx == 1 {
			z++
		}
	case lev.TileSlopeW:
		if cx == 0 {
			z++
		}
	}
	return z * HeightScale
}

// The four tile edges, each with its index (0=S,1=E,2=N,3=W), neighbour offset
// and the two shared corners.
type edge struct {
	idx      int
	nx, ny   int
	c0x, c0y int
	c1x, c1y int
}

func tileEdges(x, y int) [4]edge {
	return [4]edge{
		{0, x, y - 1, 0, 0, 1, 0}, // south (-Y)
		{1, x + 1, y, 1, 0, 1, 1}, // east
		{2, x, y + 1, 1, 1, 0, 1}, // north
		{3, x - 1, y, 0, 1, 0, 0}, // west
	}
}

// diagonal descriptors, indexed by type-2 (so type 2 -> [0]). Each names the
// solid corner (cut off), the two floor-triangle corners on the hypotenuse, and
// which two edges stay open (get normal walls).
var diagonals = [4]struct {
	solidX, solidY int
	hyp            [2][2]int // the two hypotenuse corners (cx,cy)
	openEdges      [2]int    // edge indices that remain (0=S,1=E,2=N,3=W)
}{
	{0, 1, [2][2]int{{1, 1}, {0, 0}}, [2]int{0, 1}}, // 2 NW solid: hyp NE-SW, open S,E
	{1, 1, [2][2]int{{0, 1}, {1, 0}}, [2]int{0, 3}}, // 3 NE solid: hyp NW-SE, open S,W
	{0, 0, [2][2]int{{0, 1}, {1, 0}}, [2]int{1, 2}}, // 4 SW solid: hyp NW-SE, open E,N
	{1, 0, [2][2]int{{1, 1}, {0, 0}}, [2]int{2, 3}}, // 5 SE solid: hyp NE-SW, open N,W
}

// Build generates the mesh for a grid. ceilings adds the constant-height ceiling
// faces (authentic enclosed dungeon; usually omitted so the layout shows).
func Build(g *lev.Grid, tm *lev.TexMap, ceilings bool) *Mesh {
	m := &Mesh{}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			t := g.At(x, y)
			if t.Type == lev.TileSolid {
				continue
			}
			ftex := tm.FloorTexture(t.FloorTex)
			wtex := tm.WallTexture(t.WallTex)
			fx, fy := float32(x), float32(y)
			z := func(cx, cy int) float32 { return cornerHeight(t.Type, t.Height, cx, cy) }
			corner := func(cx, cy int) [3]float32 { return [3]float32{fx + float32(cx), fy + float32(cy), z(cx, cy)} }

			if t.Type >= lev.TileDiagSE && t.Type <= lev.TileDiagNW {
				buildDiagonal(m, g, x, y, t, ftex, wtex, ceilings)
				continue
			}

			// Floor quad (CCW from above): SW, SE, NE, NW.
			m.Quads = append(m.Quads, Quad{
				P:   [4][3]float32{corner(0, 0), corner(1, 0), corner(1, 1), corner(0, 1)},
				UV:  [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
				Tex: ftex,
			})
			if ceilings {
				m.Quads = append(m.Quads, Quad{
					P:   [4][3]float32{{fx, fy, ceilingZ}, {fx, fy + 1, ceilingZ}, {fx + 1, fy + 1, ceilingZ}, {fx + 1, fy, ceilingZ}},
					UV:  [4][2]float32{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
					Tex: ftex,
				})
			}
			for _, e := range tileEdges(x, y) {
				addWall(m, g, x, y, t, e, wtex)
			}
		}
	}
	return m
}

// buildDiagonal emits a diagonal tile: a triangular floor (and ceiling), the
// diagonal wall across the hypotenuse, and walls on the two open edges.
func buildDiagonal(m *Mesh, g *lev.Grid, x, y int, t lev.Tile, ftex, wtex uint16, ceilings bool) {
	d := diagonals[t.Type-lev.TileDiagSE]
	fx, fy := float32(x), float32(y)
	z := func(cx, cy int) float32 { return cornerHeight(t.Type, t.Height, cx, cy) }
	corner := func(cx, cy int) [3]float32 { return [3]float32{fx + float32(cx), fy + float32(cy), z(cx, cy)} }

	// Floor triangle = the quad's four corners minus the solid one, in CCW order.
	ring := [4][2]int{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	var tp [3][3]float32
	var tuv [3][2]float32
	uvRing := [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	k := 0
	for i, c := range ring {
		if c[0] == d.solidX && c[1] == d.solidY {
			continue
		}
		tp[k] = corner(c[0], c[1])
		tuv[k] = uvRing[i]
		k++
	}
	m.Quads = append(m.Quads, Quad{
		P:   [4][3]float32{tp[0], tp[1], tp[2], tp[2]},
		UV:  [4][2]float32{tuv[0], tuv[1], tuv[2], tuv[2]},
		Tex: ftex, Tri: true,
	})
	if ceilings {
		m.Quads = append(m.Quads, Quad{
			P:   [4][3]float32{{tp[0][0], tp[0][1], ceilingZ}, {tp[2][0], tp[2][1], ceilingZ}, {tp[1][0], tp[1][1], ceilingZ}, {tp[1][0], tp[1][1], ceilingZ}},
			UV:  [4][2]float32{tuv[0], tuv[2], tuv[1], tuv[1]},
			Tex: ftex, Tri: true,
		})
	}

	// Diagonal wall along the hypotenuse, floor up to the ceiling (one texture).
	a, b := d.hyp[0], d.hyp[1]
	za, zb := z(a[0], a[1]), z(b[0], b[1])
	pa := [3]float32{fx + float32(a[0]), fy + float32(a[1]), za}
	pb := [3]float32{fx + float32(b[0]), fy + float32(b[1]), zb}
	m.Quads = append(m.Quads, Quad{
		P: [4][3]float32{
			{pa[0], pa[1], za}, {pb[0], pb[1], zb}, {pb[0], pb[1], ceilingZ}, {pa[0], pa[1], ceilingZ},
		},
		UV:   [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		Tex:  wtex,
		Wall: true,
	})

	// Walls on the two open edges.
	all := tileEdges(x, y)
	for _, ei := range d.openEdges {
		addWall(m, g, x, y, t, all[ei], wtex)
	}
}

// diagSolidEdge reports whether edge e (0=S,1=E,2=N,3=W) of a diagonal tile of
// the given type is a closed (solid-rock) edge — the two edges bordering its
// solid corner. The other two are open.
func diagSolidEdge(dt uint8, e int) bool {
	d := diagonals[dt-lev.TileDiagSE]
	for _, oe := range d.openEdges {
		if oe == e {
			return false
		}
	}
	return true
}

// addWall emits a wall quad on one edge wherever the neighbour rises above this
// tile's floor — up to the neighbour's own floor (a step) or the ceiling (solid
// rock). The neighbour's height is sampled at BOTH shared corners so a ramp
// meets a flush neighbour with no wall and produces a triangular side wall
// instead of a spurious full-width vertical segment. One texture spans the wall
// face (UV 0..1), so a door doesn't tile into repeated copies.
func addWall(m *Mesh, g *lev.Grid, x, y int, t lev.Tile, e edge, wtex uint16) {
	z0 := cornerHeight(t.Type, t.Height, e.c0x, e.c0y)
	z1 := cornerHeight(t.Type, t.Height, e.c1x, e.c1y)

	top0, top1 := float32(ceilingZ), float32(ceilingZ) // out-of-bounds / solid -> full wall
	if e.nx >= 0 && e.nx < g.W && e.ny >= 0 && e.ny < g.H {
		nt := g.At(e.nx, e.ny)
		// A diagonal neighbour is solid rock along the two edges by its solid
		// corner: the edge facing us (the opposite of ours, idx^2) may be closed
		// even though the tile isn't fully solid — then it still needs a wall.
		diagClosed := nt.Type >= lev.TileDiagSE && nt.Type <= lev.TileDiagNW &&
			diagSolidEdge(nt.Type, e.idx^2)
		if nt.Type != lev.TileSolid && !diagClosed {
			// neighbour floor sampled at the SAME two shared world corners
			top0 = cornerHeight(nt.Type, nt.Height, x+e.c0x-e.nx, y+e.c0y-e.ny)
			top1 = cornerHeight(nt.Type, nt.Height, x+e.c1x-e.nx, y+e.c1y-e.ny)
		}
	}
	if top0 <= z0 && top1 <= z1 {
		return // neighbour level or lower along this edge: it owns any wall
	}
	if top0 < z0 { // clamp so the wall never inverts where the neighbour dips below us
		top0 = z0
	}
	if top1 < z1 {
		top1 = z1
	}
	fx0, fy0 := float32(x+e.c0x), float32(y+e.c0y)
	fx1, fy1 := float32(x+e.c1x), float32(y+e.c1y)
	m.Quads = append(m.Quads, Quad{
		P: [4][3]float32{
			{fx0, fy0, z0}, {fx1, fy1, z1}, {fx1, fy1, top1}, {fx0, fy0, top0},
		},
		UV:   [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		Tex:  wtex,
		Wall: true,
	})
}
