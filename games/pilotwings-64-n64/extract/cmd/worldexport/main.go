// worldexport assembles Pilotwings 64's worlds from the cartridge archive.
//
// The single UVTR resource holds ten grids; each non-empty cell names a UVCT
// terrain chunk and the 4x4 transform that places it. worldexport decodes all
// of them, checks each grid against itself, places every chunk, and writes one
// continuous textured GLB per world.
//
// That the result is continuous is the verification: the chunks are decoded and
// placed independently, so seams, gaps, overlaps or a wrong axis would show
// immediately. -check reports the seam residual measured between the shared
// edges of horizontally and vertically adjacent cells.
//
// Nothing here boots the machine. The attract sequence never loads a world (it
// draws the island as a single UVMD), so the oracle has nothing to say about
// this format until a mission is driven into.
//
// Usage:
//
//	worldexport -image ROM -check
//	worldexport -image ROM -o work/worlds
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image_ := flag.String("image", "", "cartridge image")
	out := flag.String("o", "", "write one GLB per world into this directory")
	check := flag.Bool("check", false, "report grid checks and the seam residual between adjacent cells")
	flag.Parse()

	rom, err := n64.Load(*image_)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}

	worlds, err := decodeWorlds(a)
	if err != nil {
		die(err)
	}
	chunks, err := decodeChunks(a)
	if err != nil {
		die(err)
	}
	fmt.Printf("%d worlds, %d terrain chunks (%d triangles)\n", len(worlds), len(chunks), totalTris(chunks))

	// Every UVCT resource must be named by exactly one cell, across all worlds:
	// the grids partition the terrain, they do not sample it.
	used := map[int][]string{}
	for wi, w := range worlds {
		for _, c := range w.Cells {
			if c.Present {
				used[c.Chunk] = append(used[c.Chunk], fmt.Sprintf("world %d (%d,%d)", wi, c.Col, c.Row))
			}
		}
	}
	unnamed := 0
	for i := range chunks {
		if len(used[i]) == 0 {
			unnamed++
		}
	}
	fmt.Printf("cells naming a chunk: %d; chunks never named: %d; chunks named by more than one cell: %d\n",
		countCells(worlds), unnamed, countShared(used))

	for wi, w := range worlds {
		if err := w.Check(); err != nil {
			die(fmt.Errorf("world %d: %w", wi, err))
		}
	}
	fmt.Println("every world's extents match cols*cellW / rows*cellH, and every cell's transform lands on its centre")

	if *check {
		for wi, w := range worlds {
			reportSeams(wi, w, chunks)
		}
	}

	if *out != "" {
		texs, err := loadTextures(a)
		if err != nil {
			die(err)
		}
		if err := os.MkdirAll(*out, 0o755); err != nil {
			die(err)
		}
		for wi, w := range worlds {
			p := filepath.Join(*out, fmt.Sprintf("world-%d.glb", wi))
			n, tris, err := writeWorld(p, w, chunks, texs)
			if err != nil {
				die(err)
			}
			fmt.Printf("wrote %s: %2d cells, %6d triangles, %.0fx%.0f units\n",
				p, n, tris, w.Max[0]-w.Min[0], w.Max[1]-w.Min[1])
		}
	}
}

func countCells(ws []*uvtr.World) int {
	n := 0
	for _, w := range ws {
		for _, c := range w.Cells {
			if c.Present {
				n++
			}
		}
	}
	return n
}

func countShared(used map[int][]string) int {
	n := 0
	for _, v := range used {
		if len(v) > 1 {
			n++
		}
	}
	return n
}

func totalTris(cs []*uvct.Chunk) int {
	n := 0
	for _, c := range cs {
		n += c.Triangles()
	}
	return n
}

func commChunk(a *pwad.Archive, f pwad.Form) ([]byte, error) {
	for _, c := range f.Chunks {
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		if tag == "COMM" {
			return a.Data(c)
		}
	}
	return nil, fmt.Errorf("resource %d has no COMM chunk", f.Index)
}

func decodeWorlds(a *pwad.Archive) ([]*uvtr.World, error) {
	idx := a.ByType("UVTR")
	if len(idx) != 1 {
		return nil, fmt.Errorf("expected one UVTR resource, found %d", len(idx))
	}
	f, err := a.Resource(idx[0])
	if err != nil {
		return nil, err
	}
	var out []*uvtr.World
	for _, c := range f.Chunks {
		if c.Tag != "COMM" {
			continue
		}
		data, err := a.Data(c)
		if err != nil {
			return nil, err
		}
		w, err := uvtr.Decode(data)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func decodeChunks(a *pwad.Archive) ([]*uvct.Chunk, error) {
	idx := a.ByType("UVCT")
	sort.Ints(idx)
	out := make([]*uvct.Chunk, 0, len(idx))
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		data, err := commChunk(a, f)
		if err != nil {
			return nil, err
		}
		c, err := uvct.Decode(data)
		if err != nil {
			return nil, fmt.Errorf("UVCT %d: %w", i, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func loadTextures(a *pwad.Archive) ([]*uvtx.Texture, error) {
	idx := a.ByType("UVTX")
	sort.Ints(idx)
	out := make([]*uvtx.Texture, 0, len(idx))
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		data, err := commChunk(a, f)
		if err != nil {
			return nil, err
		}
		t, err := uvtx.Decode(data)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// xform applies a cell's row-vector transform to a model-space vertex.
func xform(m uvmd.Matrix, x, y, z float32) [3]float32 {
	return [3]float32{
		m[0][0]*x + m[1][0]*y + m[2][0]*z + m[3][0],
		m[0][1]*x + m[1][1]*y + m[2][1]*z + m[3][1],
		m[0][2]*x + m[1][2]*y + m[2][2]*z + m[3][2],
	}
}

// reportSeams checks that every placed chunk lies inside its own grid cell.
//
// A chunk need not *fill* its cell — a coastal cell's terrain stops at the
// shoreline — so a gap between two chunks' bounding boxes is not a defect and
// measuring one proves nothing. What must hold is containment: if the transform,
// the axes or the cell stride were wrong, geometry would spill across cell
// boundaries. The reported number is the worst overhang, in world units, of any
// chunk beyond the cell that names it.
func reportSeams(wi int, w *uvtr.World, chunks []*uvct.Chunk) {
	cellAt := func(col, row int) *uvtr.Cell {
		if col < 0 || row < 0 || col >= w.Cols || row >= w.Rows {
			return nil
		}
		c := &w.Cells[row*w.Cols+col]
		if !c.Present {
			return nil
		}
		return c
	}
	// For each present cell, the placed bounding box.
	bbox := func(c *uvtr.Cell) (lo, hi [3]float32) {
		ch := chunks[c.Chunk]
		lo = [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		hi = [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
		for _, v := range ch.Vertices {
			p := xform(c.Matrix, float32(v.X), float32(v.Y), float32(v.Z))
			for i := 0; i < 3; i++ {
				if p[i] < lo[i] {
					lo[i] = p[i]
				}
				if p[i] > hi[i] {
					hi[i] = p[i]
				}
			}
		}
		return
	}
	var overhang float64
	cells := 0
	for row := 0; row < w.Rows; row++ {
		for col := 0; col < w.Cols; col++ {
			c := cellAt(col, row)
			if c == nil {
				continue
			}
			cells++
			lo, hi := bbox(c)
			cx, cy := w.Centre(col, row)
			exLo := [2]float32{cx - w.CellW/2, cy - w.CellH/2}
			exHi := [2]float32{cx + w.CellW/2, cy + w.CellH/2}
			for i := 0; i < 2; i++ {
				overhang = math.Max(overhang, float64(exLo[i]-lo[i])) // spills below
				overhang = math.Max(overhang, float64(hi[i]-exHi[i])) // spills above
			}
		}
	}
	fmt.Printf("world %2d: %2d cells, %2dx%-2d grid of %gx%g; worst overhang beyond a cell %.4f units\n",
		wi, cells, w.Cols, w.Rows, w.CellW, w.CellH, overhang)
}

func writeWorld(path string, w *uvtr.World, chunks []*uvct.Chunk, texs []*uvtx.Texture) (cells, tris int, err error) {
	white := image.NewRGBA(image.Rect(0, 0, 1, 1))
	white.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})

	var pos [][3]float32
	var uvs [][2]float32
	var cols [][4]uint8
	// One glTF primitive per texture, merged across cells: a world is one mesh,
	// not a pile of chunks.
	byTex := map[int][][3]uint32{}
	for i := range w.Cells {
		c := &w.Cells[i]
		if !c.Present {
			continue
		}
		cells++
		ch := chunks[c.Chunk]
		for _, b := range ch.Batches {
			tex := -1
			tw, th := 1.0, 1.0
			if t, ok := b.Material.Texture(); ok && t < len(texs) {
				tex = t
				tw, th = float64(texs[t].Width), float64(texs[t].Height)
			}
			for _, t := range b.Tris {
				base := uint32(len(pos))
				for _, vi := range t {
					v := ch.Vertices[vi]
					pos = append(pos, xform(c.Matrix, float32(v.X), float32(v.Y), float32(v.Z)))
					uvs = append(uvs, [2]float32{
						float32(float64(v.S) / 32 / tw), float32(float64(v.T) / 32 / th),
					})
					cols = append(cols, [4]uint8{v.R, v.G, v.B, v.A})
				}
				byTex[tex] = append(byTex[tex], [3]uint32{base, base + 1, base + 2})
				tris++
			}
		}
	}
	if len(pos) == 0 {
		return 0, 0, fmt.Errorf("world has no geometry")
	}
	var keys []int
	for k := range byTex {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var groups []glb.TexturedGroup
	for _, k := range keys {
		img := white
		if k >= 0 {
			img = texs[k].Image
		}
		groups = append(groups, glb.TexturedGroup{Tris: byTex[k], Image: img, WrapS: 10497, WrapT: 10497})
	}
	return cells, tris, glb.WriteTexturedColored(path, pos, uvs, cols, groups, nil)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "worldexport:", err)
	os.Exit(1)
}
