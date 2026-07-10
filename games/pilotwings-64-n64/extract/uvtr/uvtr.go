// Package uvtr decodes Pilotwings 64's world grids.
//
// The archive holds exactly one UVTR resource, and it carries ten uncompressed
// COMM chunks — ten worlds. Each is a rectangular grid of cells, and a cell that
// is not empty names one UVCT terrain chunk and the transform that places it.
//
// The reader is at 0x802270BC:
//
//	f32 minX, minY, minZ        the world's lower bound
//	f32 maxX, maxY, maxZ        its upper bound
//	u8  cols, u8 rows
//	f32 cellW, f32 cellH
//	f32 radius                  half-diagonal of the world, unused here
//	Cell[cols*rows]             row-major
//
// and a cell is:
//
//	u8 present                  0 -> the loader zeroes a 72-byte slot and moves on
//	if present:
//	    f32[16] matrix          a 4x4 row-vector transform
//	    u8  flag
//	    u16 chunk               the UVCT resource's ordinal, 0..100
//
// The terrain's ground plane is X/Y with Z up: a cell's matrix translation is
// its grid cell's **centre**, `(minX + (col+0.5)*cellW, minY + (row+0.5)*cellH)`,
// which Check verifies for every cell of every world. Together with the extents
// agreeing exactly with `cols*cellW` and `rows*cellH`, and with the ten worlds'
// cells naming each of the 101 UVCT resources, that is what pins the grid.
package uvtr

import (
	"encoding/binary"
	"fmt"
	"math"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
)

// headerSize is the fixed header ahead of the cells.
const headerSize = 0x26

// Cell is one grid cell. An empty cell has no terrain.
type Cell struct {
	Present bool
	Matrix  uvmd.Matrix
	Flag    uint8
	Chunk   int // UVCT ordinal, 0..100
	Col     int
	Row     int
}

// World is one of the ten grids.
type World struct {
	Min, Max     [3]float32
	Cols, Rows   int
	CellW, CellH float32
	Radius       float32
	Cells        []Cell
	Padding      int
}

// Centre is the world-space centre of cell (col,row).
func (w *World) Centre(col, row int) (x, y float32) {
	return w.Min[0] + (float32(col)+0.5)*w.CellW, w.Min[1] + (float32(row)+0.5)*w.CellH
}

func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func f32(b []byte, o int) float32 { return math.Float32frombits(be32(b, o)) }
func be16(b []byte, o int) uint16 { return binary.BigEndian.Uint16(b[o:]) }

// Decode parses one of the UVTR resource's COMM chunks.
func Decode(data []byte) (*World, error) {
	if len(data) < headerSize {
		return nil, fmt.Errorf("uvtr: chunk shorter than its header")
	}
	w := &World{
		Min:    [3]float32{f32(data, 0), f32(data, 4), f32(data, 8)},
		Max:    [3]float32{f32(data, 0xC), f32(data, 0x10), f32(data, 0x14)},
		Cols:   int(data[0x18]),
		Rows:   int(data[0x19]),
		CellW:  f32(data, 0x1A),
		CellH:  f32(data, 0x1E),
		Radius: f32(data, 0x22),
	}
	if w.Cols == 0 || w.Rows == 0 {
		return nil, fmt.Errorf("uvtr: %dx%d grid", w.Cols, w.Rows)
	}

	p := headerSize
	w.Cells = make([]Cell, w.Cols*w.Rows)
	for i := range w.Cells {
		if p >= len(data) {
			return nil, fmt.Errorf("uvtr: cell %d runs past the chunk", i)
		}
		c := &w.Cells[i]
		c.Col, c.Row = i%w.Cols, i/w.Cols
		present := data[p]
		p++
		if present == 0 {
			continue
		}
		if present != 1 {
			return nil, fmt.Errorf("uvtr: cell %d has present=%d", i, present)
		}
		if p+67 > len(data) {
			return nil, fmt.Errorf("uvtr: cell %d body runs past the chunk", i)
		}
		c.Present = true
		for r := 0; r < 4; r++ {
			for col := 0; col < 4; col++ {
				c.Matrix[r][col] = f32(data, p+(r*4+col)*4)
			}
		}
		p += 64
		c.Flag = data[p]
		p++
		c.Chunk = int(be16(data, p))
		p += 2
	}
	w.Padding = len(data) - p
	if w.Padding < 0 || w.Padding >= 8 {
		return nil, fmt.Errorf("uvtr: %d bytes left after the parse", w.Padding)
	}
	return w, nil
}

// Check verifies the grid against itself: the extents must equal cols*cellW and
// rows*cellH exactly, and every present cell's transform must translate to that
// cell's centre. A misread of the cell record — the wrong field order, a wrong
// stride — cannot satisfy the second condition by accident.
func (w *World) Check() error {
	if got, want := w.Max[0]-w.Min[0], float32(w.Cols)*w.CellW; !near(got, want, 0.01) {
		return fmt.Errorf("uvtr: x extent %g but %d cells of %g = %g", got, w.Cols, w.CellW, want)
	}
	if got, want := w.Max[1]-w.Min[1], float32(w.Rows)*w.CellH; !near(got, want, 0.01) {
		return fmt.Errorf("uvtr: y extent %g but %d cells of %g = %g", got, w.Rows, w.CellH, want)
	}
	for _, c := range w.Cells {
		if !c.Present {
			continue
		}
		cx, cy := w.Centre(c.Col, c.Row)
		if !near(c.Matrix[3][0], cx, 0.01) || !near(c.Matrix[3][1], cy, 0.01) {
			return fmt.Errorf("uvtr: cell (%d,%d) translates to (%g,%g), its centre is (%g,%g)",
				c.Col, c.Row, c.Matrix[3][0], c.Matrix[3][1], cx, cy)
		}
	}
	return nil
}

func near(a, b, eps float32) bool { return a-b < eps && b-a < eps }
