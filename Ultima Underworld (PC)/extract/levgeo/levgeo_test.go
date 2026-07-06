package levgeo

import (
	"testing"

	"ultimaunderworld/extract/lev"
)

// A slope whose floor meets a flat neighbour flush must not get a wall on that
// edge — even when the slope's SW corner (the one the old single-corner code
// sampled) is raised. SlopeS raises its south corners, so its NORTH edge sits at
// the base height, flush with a flat tile to the north at the same height.
func TestBuildSlopeFlushNoWall(t *testing.T) {
	g := &lev.Grid{W: 3, H: 4, Tiles: make([]lev.Tile, 12)}
	for i := range g.Tiles {
		g.Tiles[i] = lev.Tile{Type: lev.TileSolid}
	}
	g.Tiles[1*3+1] = lev.Tile{Type: lev.TileSlopeS, Height: 4} // south corners raised to 5, north edge = 4
	g.Tiles[2*3+1] = lev.Tile{Type: lev.TileOpen, Height: 4}   // flat, north of the slope, flush at 4
	tm := &lev.TexMap{Wall: make([]uint16, 48), Floor: make([]uint16, 10)}

	m := Build(g, tm, false)
	// The shared edge is the plane y==2 (between tiles (1,1) and (1,2)); no wall
	// quad should lie entirely in it.
	for _, q := range m.Quads {
		if !q.Wall {
			continue
		}
		if q.P[0][1] == 2 && q.P[1][1] == 2 && q.P[2][1] == 2 && q.P[3][1] == 2 {
			t.Errorf("spurious wall on flush slope edge at y=2: %+v", q.P)
		}
	}
}

// Wall textures use a uniform texel scale: one copy per tile width horizontally
// (U in 0..1) and per WallTexUnitsPerCopy vertically, so a full-height wall tiles
// the texture instead of stretching one copy over floor-to-ceiling. V starts at
// 0 at the foot (upright, not upside-down).
func TestBuildWallUVUniform(t *testing.T) {
	g := &lev.Grid{W: 3, H: 3, Tiles: make([]lev.Tile, 9)}
	for i := range g.Tiles {
		g.Tiles[i] = lev.Tile{Type: lev.TileSolid}
	}
	g.Tiles[1*3+1] = lev.Tile{Type: lev.TileOpen, Height: 0} // full-height walls to the ceiling
	tm := &lev.TexMap{Wall: make([]uint16, 48), Floor: make([]uint16, 10)}

	m := Build(g, tm, false)
	wantTop := float32(ceilingZ / WallTexUnitsPerCopy)
	for _, q := range m.Quads {
		if !q.Wall {
			continue
		}
		var minV, maxV float32 = 1e9, -1e9
		for _, uv := range q.UV {
			if uv[0] < 0 || uv[0] > 1 {
				t.Errorf("wall U outside 0..1: %v", uv)
			}
			if uv[1] < minV {
				minV = uv[1]
			}
			if uv[1] > maxV {
				maxV = uv[1]
			}
		}
		if minV != 0 {
			t.Errorf("wall V foot = %v, want 0 (upright, non-stretched)", minV)
		}
		if maxV != wantTop {
			t.Errorf("wall V top = %v, want %v (uniform tiling to the ceiling)", maxV, wantTop)
		}
	}
}

// A 2x2 grid: one open tile ringed by solid should get a floor and four walls.
func TestBuildOpenTileWalls(t *testing.T) {
	g := &lev.Grid{W: 3, H: 3, Tiles: make([]lev.Tile, 9)}
	for i := range g.Tiles {
		g.Tiles[i] = lev.Tile{Type: lev.TileSolid}
	}
	g.Tiles[1*3+1] = lev.Tile{Type: lev.TileOpen, Height: 2} // centre open
	tm := &lev.TexMap{Wall: make([]uint16, 48), Floor: make([]uint16, 10)}

	m := Build(g, tm, false)
	floors, walls := 0, 0
	for _, q := range m.Quads {
		if q.Wall {
			walls++
		} else {
			floors++
		}
	}
	if floors != 1 {
		t.Errorf("floors = %d, want 1", floors)
	}
	if walls != 4 {
		t.Errorf("walls = %d, want 4 (open tile ringed by solid)", walls)
	}
}
