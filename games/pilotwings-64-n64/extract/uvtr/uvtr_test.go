package uvtr

import (
	"os"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func archive(t *testing.T) *pwad.Archive {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func worlds(t *testing.T, a *pwad.Archive) []*World {
	t.Helper()
	idx := a.ByType("UVTR")
	if len(idx) != 1 {
		t.Fatalf("%d UVTR resources, want 1", len(idx))
	}
	f, err := a.Resource(idx[0])
	if err != nil {
		t.Fatal(err)
	}
	var out []*World
	for _, c := range f.Chunks {
		if c.Tag != "COMM" {
			continue
		}
		data, err := a.Data(c)
		if err != nil {
			t.Fatal(err)
		}
		w, err := Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, w)
	}
	return out
}

// The one UVTR resource holds ten worlds, and each is self-consistent: its
// extents are exactly cols*cellW by rows*cellH, and every present cell's
// transform translates to that cell's centre. The second condition is the one
// that cannot be satisfied by a misread — a wrong field order or stride would
// scatter the translations.
func TestTenWorldsCheckOut(t *testing.T) {
	a := archive(t)
	ws := worlds(t, a)
	if got, want := len(ws), 10; got != want {
		t.Fatalf("%d worlds, want %d", got, want)
	}
	for i, w := range ws {
		if err := w.Check(); err != nil {
			t.Errorf("world %d: %v", i, err)
		}
		if w.Padding >= 8 {
			t.Errorf("world %d: %d bytes unconsumed", i, w.Padding)
		}
	}
}

// The grids partition the terrain: every one of the 101 UVCT resources is named
// by at least one cell, and no cell names a resource that does not exist.
func TestEveryTerrainChunkIsPlaced(t *testing.T) {
	a := archive(t)
	ws := worlds(t, a)
	n := len(a.ByType("UVCT"))
	if n != 101 {
		t.Fatalf("%d UVCT resources, want 101", n)
	}
	named := map[int]int{}
	cells := 0
	for wi, w := range ws {
		for _, c := range w.Cells {
			if !c.Present {
				continue
			}
			cells++
			if c.Chunk < 0 || c.Chunk >= n {
				t.Fatalf("world %d cell (%d,%d) names chunk %d of %d", wi, c.Col, c.Row, c.Chunk, n)
			}
			named[c.Chunk]++
		}
	}
	if len(named) != n {
		t.Errorf("%d of %d chunks are named by a cell", len(named), n)
	}
	if cells != 120 {
		t.Errorf("%d cells name a chunk, want 120", cells)
	}
}

// Every terrain chunk decodes, and its own internal references hold: each
// batch's face slice lies inside the face array, each face's vertex indices
// inside the vertex pool.
func TestEveryTerrainChunkDecodes(t *testing.T) {
	a := archive(t)
	idx := a.ByType("UVCT")
	tris := 0
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			if tag != "COMM" {
				continue
			}
			data, err := a.Data(c)
			if err != nil {
				t.Fatal(err)
			}
			ch, err := uvct.Decode(data)
			if err != nil {
				t.Fatalf("UVCT %d: %v", i, err)
			}
			if ch.Padding >= 8 {
				t.Errorf("UVCT %d: %d bytes unconsumed", i, ch.Padding)
			}
			tris += ch.Triangles()
		}
	}
	if want := 20846; tris != want {
		t.Errorf("terrain has %d triangles, want %d", tris, want)
	}
}

// No placed chunk may spill outside the cell that names it. A chunk need not
// fill its cell — a coastal cell stops at the shoreline — but geometry crossing
// a cell boundary would mean the transform, the axes or the cell stride is wrong.
func TestChunksStayInsideTheirCells(t *testing.T) {
	a := archive(t)
	ws := worlds(t, a)
	chunks := map[int]*uvct.Chunk{}
	for _, i := range a.ByType("UVCT") {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			if tag != "COMM" {
				continue
			}
			data, err := a.Data(c)
			if err != nil {
				t.Fatal(err)
			}
			ch, err := uvct.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			chunks[i-a.ByType("UVCT")[0]] = ch
		}
	}
	const eps = 0.01
	for wi, w := range ws {
		for _, c := range w.Cells {
			if !c.Present {
				continue
			}
			ch := chunks[c.Chunk]
			cx, cy := w.Centre(c.Col, c.Row)
			loX, hiX := cx-w.CellW/2-eps, cx+w.CellW/2+eps
			loY, hiY := cy-w.CellH/2-eps, cy+w.CellH/2+eps
			for _, v := range ch.Vertices {
				x := c.Matrix[0][0]*float32(v.X) + c.Matrix[1][0]*float32(v.Y) + c.Matrix[2][0]*float32(v.Z) + c.Matrix[3][0]
				y := c.Matrix[0][1]*float32(v.X) + c.Matrix[1][1]*float32(v.Y) + c.Matrix[2][1]*float32(v.Z) + c.Matrix[3][1]
				if x < loX || x > hiX || y < loY || y > hiY {
					t.Fatalf("world %d cell (%d,%d) chunk %d: vertex at (%g,%g) escapes the cell x[%g,%g] y[%g,%g]",
						wi, c.Col, c.Row, c.Chunk, x, y, loX, hiX, loY, hiY)
				}
			}
		}
	}
}
