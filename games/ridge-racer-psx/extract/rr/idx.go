package rr

// idx.go decodes IDX.HED: the 32×32 course grid. The world is a square of
// 32×32 cells of 2048 units; each grid slot holds the MAP.RRM section index
// occupying that cell, or 0xFFFF for empty ground. The renderer (grid walk at
// 0x80012478) converts a world position to a cell with x>>11, z>>11 and looks
// a cell up as grid[z*32 + 30 - x]; a section's world origin is
// (cellX*2048, 0, cellZ*2048).

import "fmt"

// CellSize is the world-unit edge length of one grid cell.
const CellSize = 2048

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
