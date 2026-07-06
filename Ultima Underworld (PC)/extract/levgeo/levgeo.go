// Package levgeo turns a decoded Ultima Underworld tile map into 3D polygons —
// the static level geometry — following the tile-geometry rules reverse-
// engineered in Part V: solid tiles are rock (walls come from their open
// neighbours), open tiles get a flat floor, slopes tilt two corners by one
// height unit, and diagonals cut the tile with a wall. A wall quad is emitted on
// an edge wherever the neighbour is solid (full height to the ceiling) or its
// floor is higher (a step up). Floor/ceiling faces take the tile's floor texture
// and walls its wall texture, mapped to global W64.TR/F32.TR numbers through the
// per-level texture list.
//
// One tile is a 1x1 cell in XY; heights are in tile-height units (the map's own
// 0-15 field). Coordinates: +X east, +Y north, +Z up.
package levgeo

import "ultimaunderworld/extract/lev"

// Ceiling is the level-wide ceiling height (tile-height units) walls rise to.
const Ceiling = 16

// HeightScale converts a tile-height unit into world units where a tile is 1x1.
// UW's floor heights are much finer than a tile width, so heights are scaled
// down for correct proportions (a full-height wall reads as roughly one tile).
const HeightScale = 0.14

// ceilingZ is the ceiling in world units.
const ceilingZ = Ceiling * HeightScale

// Quad is one textured polygon (a triangle repeats its last vertex). P holds the
// four corners; UV the per-corner texture coordinates (0..1, tiled).
type Quad struct {
	P    [4][3]float32
	UV   [4][2]float32
	Tex  uint16 // global texture number
	Wall bool   // true = W64.TR wall texture, false = F32.TR floor/ceiling
}

// Mesh is the whole level's geometry.
type Mesh struct{ Quads []Quad }

// cornerHeight returns the floor height at tile corner (cx,cy in {0,1}) for a
// tile of the given type and base height. Slopes raise the two corners on their
// high side by one unit.
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
	case lev.TileSlopeE: // toward +X
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

// Build generates the mesh for a grid, using tm to resolve texture indices to
// global texture numbers. ceilings adds the (constant-height) ceiling faces —
// authentic to the enclosed dungeon, but usually omitted so the layout is
// visible from above.
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

			// Floor: a quad whose corners carry the (possibly sloped) heights.
			fx := float32(x)
			fy := float32(y)
			m.Quads = append(m.Quads, Quad{
				P: [4][3]float32{
					{fx, fy, cornerHeight(t.Type, t.Height, 0, 0)},
					{fx + 1, fy, cornerHeight(t.Type, t.Height, 1, 0)},
					{fx + 1, fy + 1, cornerHeight(t.Type, t.Height, 1, 1)},
					{fx, fy + 1, cornerHeight(t.Type, t.Height, 0, 1)},
				},
				UV:  [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
				Tex: ftex,
			})
			if ceilings {
				m.Quads = append(m.Quads, Quad{
					P: [4][3]float32{
						{fx, fy, ceilingZ},
						{fx, fy + 1, ceilingZ},
						{fx + 1, fy + 1, ceilingZ},
						{fx + 1, fy, ceilingZ},
					},
					UV:  [4][2]float32{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
					Tex: ftex,
				})
			}

			// Walls on the four edges. edge = the two tile corners of that side.
			addWall(m, g, x, y, t, wtex)
		}
	}
	return m
}

// addWall emits wall quads on each of the tile's four edges where they are
// exposed (neighbour solid or higher-floored).
func addWall(m *Mesh, g *lev.Grid, x, y int, t lev.Tile, wtex uint16) {
	// Each edge: neighbour offset, and the two shared corners (cx,cy).
	edges := []struct {
		nx, ny   int
		c0x, c0y int
		c1x, c1y int
	}{
		{x, y - 1, 0, 0, 1, 0}, // south edge (toward -Y)
		{x + 1, y, 1, 0, 1, 1}, // east
		{x, y + 1, 1, 1, 0, 1}, // north
		{x - 1, y, 0, 1, 0, 0}, // west
	}
	for _, e := range edges {
		var nfloor float32 = ceilingZ // out-of-bounds / solid -> full wall
		solid := true
		if e.nx >= 0 && e.nx < g.W && e.ny >= 0 && e.ny < g.H {
			nt := g.At(e.nx, e.ny)
			if nt.Type != lev.TileSolid {
				solid = false
				nfloor = cornerHeight(nt.Type, nt.Height, 0, 0)
			}
		}
		// The wall rises from this tile's floor at the shared corners up to the
		// ceiling (solid neighbour) or the neighbour's floor (a step up).
		z0 := cornerHeight(t.Type, t.Height, e.c0x, e.c0y)
		z1 := cornerHeight(t.Type, t.Height, e.c1x, e.c1y)
		top := nfloor
		if !solid && nfloor <= z0 && nfloor <= z1 {
			continue // neighbour is level or lower: no wall on this side
		}
		p0 := [3]float32{float32(x + e.c0x), float32(y + e.c0y), z0}
		p1 := [3]float32{float32(x + e.c1x), float32(y + e.c1y), z1}
		hgt := top - min32(z0, z1)
		m.Quads = append(m.Quads, Quad{
			P: [4][3]float32{
				{p0[0], p0[1], z0},
				{p1[0], p1[1], z1},
				{p1[0], p1[1], top},
				{p0[0], p0[1], top},
			},
			UV:   [4][2]float32{{0, hgt}, {1, hgt}, {1, 0}, {0, 0}},
			Tex:  wtex,
			Wall: true,
		})
	}
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
