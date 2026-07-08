package rr

// idx.go decodes IDX.HED: the 32×32 course grid. Each grid slot holds the
// MAP.RRM section index occupying that cell, or 0xFFFF for empty ground. The
// renderer (grid walk at 0x80012478) converts a position to a cell with
// x>>11, z>>11 and looks a cell up as grid[z*32 + 30 - x].
//
// Two unit systems meet here: positions (the camera, the cars) are kept in
// quarter model units — a cell is 2048 position units — while the section
// records are in model units. The grid walk shifts the rotated cell
// translation left by 2 (0x80012568) before it becomes the GTE translation,
// so in the records' own units a cell is 8192 across and a section's world
// origin is (cellX*8192, 0, cellZ*8192).

import "fmt"

// CellSize is the edge length of one grid cell in position units (the
// x>>11 of the grid walk).
const CellSize = 2048

// CellModel is the edge length of one grid cell in model units — the units
// of the MAP.RRM/OBJ.RRO vertices (position units × 4, the walk's <<2).
const CellModel = CellSize * 4

// Grid is the course grid; Empty marks an unoccupied cell.
type Grid struct {
	cells [1024]uint16
}

// Empty is the grid value of a cell with no track section.
const Empty = 0xFFFF

// ParseIDX decodes the 2048-byte IDX.HED image.
func ParseIDX(d []byte) (*Grid, error) {
	if len(d) != 2048 {
		return nil, fmt.Errorf("idx: want 2048 bytes, have %d", len(d))
	}
	g := &Grid{}
	for i := range g.cells {
		g.cells[i] = u16(d, i*2)
	}
	return g, nil
}

// Section returns the MAP.RRM section index at cell (x, z), or Empty.
// Coordinates outside the game's addressable range return Empty.
func (g *Grid) Section(x, z int) uint16 {
	if x < 0 || z < 0 || x >= 32 || z >= 32 {
		return Empty
	}
	i := z*32 + 30 - x
	if i < 0 || i >= len(g.cells) {
		return Empty
	}
	return g.cells[i]
}
