package levgeo

import (
	"testing"

	"ultimaunderworld/extract/lev"
)

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
