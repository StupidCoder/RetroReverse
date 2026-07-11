// clvdump decodes a stage file (st_*.clv) and reports its layout, cells,
// batches and strips.
//
//	clvdump -in st_flower01.clv [-cells] [-verify PRIMLOG -base HEX]
//
// -verify cross-checks the decoded strips against a PSP_GE_DEBUG PRIM log from
// the running game: every logged PRIM whose vertex address falls inside the
// loaded stage image must correspond byte-exactly to a decoded strip (same
// file offset, same vertex count).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"regexp"
	"sort"
	"strconv"

	"retroreverse.com/games/loco-roco-psp/extract/clv"
)

func main() {
	in := flag.String("in", "", "stage file (.clv)")
	cells := flag.Bool("cells", false, "list every cell's batches")
	texDir := flag.String("textures", "", "decode every batch material's texture to PNGs in this directory")
	preview := flag.String("preview", "", "render an orthographic flat-colour preview of the stage to this PNG")
	objects := flag.Bool("objects", false, "list the stage's prop placements")
	verify := flag.String("verify", "", "PSP_GE_DEBUG PRIM log to cross-check strips against")
	baseS := flag.String("base", "094400C0", "with -verify, RAM base the stage image was loaded at (hex)")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "need -in FILE.clv")
		os.Exit(1)
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		die(err)
	}
	c, err := clv.Parse(raw)
	if err != nil {
		die(err)
	}
	L := &c.Layout
	fmt.Printf("image %d bytes, scene @%#x, reloc @%#x (%d pointer slots, all valid)\n",
		len(c.Data), c.SceneOff, c.RelocOff, len(c.Reloc))
	fmt.Printf("layout: origin (%.0f,%.0f) size %.0fx%.0f z %.0f..%.0f, %dx%d cells of %.0f\n",
		L.X, L.Y, L.W, L.H, L.Z0, L.Z1, L.Cols, L.Rows, L.CellSize)
	nb, ns, nv := 0, 0, 0
	mats := map[string]int{}
	for i := range L.Cells {
		for _, b := range L.Cells[i].Batches {
			nb++
			mats[b.MaterialName]++
			for _, s := range b.Strips {
				ns++
				nv += len(s.Verts)
			}
		}
	}
	fmt.Printf("%d non-empty cells, %d batches, %d strips, %d vertices\n",
		countNonEmpty(L.Cells), nb, ns, nv)
	col := &c.Collision
	if len(col.Points) > 0 {
		fmt.Printf("collision: %d points, %d layers (", len(col.Points), len(col.Layers))
		for i, cl := range col.Layers {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d edges p=%g", len(cl.Edges), cl.Param)
		}
		fmt.Printf("), %d entities\n", col.EntityCount)
	}
	fmt.Printf("materials used by batches:\n")
	for m, n := range mats {
		fmt.Printf("  %4d  %s\n", n, m)
	}
	if *cells {
		for i := range L.Cells {
			if len(L.Cells[i].Batches) == 0 {
				continue
			}
			fmt.Printf("cell %d (row %d col %d):\n", i, i/int(L.Cols), i%int(L.Cols))
			for _, b := range L.Cells[i].Batches {
				fmt.Printf("  %-24s color %08X  %d strips\n", b.MaterialName, b.Color, len(b.Strips))
			}
		}
	}
	if *texDir != "" {
		if err := os.MkdirAll(*texDir, 0755); err != nil {
			die(err)
		}
		done := map[string]bool{}
		for i := range L.Cells {
			for _, b := range L.Cells[i].Batches {
				m, err := c.Material(b)
				if err != nil {
					fmt.Printf("material %q: %v\n", b.MaterialName, err)
					continue
				}
				if done[m.TexName] {
					continue
				}
				done[m.TexName] = true
				img, err := c.DecodeTexture(m)
				if err != nil {
					fmt.Printf("texture %q: %v\n", m.TexName, err)
					continue
				}
				p := *texDir + "/" + m.TexName + ".png"
				f, err := os.Create(p)
				if err != nil {
					die(err)
				}
				if err := png.Encode(f, img); err != nil {
					die(err)
				}
				f.Close()
				fmt.Printf("wrote %s (%dx%d, %d mips, uv x%.1f,%.1f)\n", p, m.TexW, m.TexH, m.Mips, m.UScale, m.VScale)
			}
		}
	}
	if *objects {
		pls := c.Placements()
		fmt.Printf("%d prop placements:\n", len(pls))
		for _, p := range pls {
			fmt.Printf("  %-28s pos (%8.1f,%8.1f,%6.1f) rotZ %7.2f scale (%.2f,%.2f,%.2f)\n",
				p.Name, p.Pos[0], p.Pos[1], p.Pos[2], p.RotZ, p.Scale[0], p.Scale[1], p.Scale[2])
		}
	}
	if *preview != "" {
		if err := renderPreview(c, *preview); err != nil {
			die(err)
		}
		fmt.Printf("wrote %s\n", *preview)
	}
	if *verify != "" {
		base, err := strconv.ParseUint(*baseS, 16, 32)
		if err != nil {
			die(fmt.Errorf("bad -base"))
		}
		verifyLog(c, uint32(base), *verify)
	}
}

// renderPreview draws every strip as flat-colour triangles in an orthographic
// top view (world x right, world y up), batches in cell order — a quick visual
// check that the decoded geometry is the stage.
func renderPreview(c *clv.Clv, path string) error {
	const W = 1040
	L := &c.Layout
	scale := float32(W) / L.W
	H := int(L.H*scale) + 1
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	// batches sorted back-to-front by their first vertex's z, as the painter
	// order approximates the game's layering
	type job struct {
		z   float32
		col [4]uint8
		b   clv.Batch
	}
	var jobs []job
	for i := range L.Cells {
		for _, b := range L.Cells[i].Batches {
			if len(b.Strips) == 0 || len(b.Strips[0].Verts) == 0 {
				continue
			}
			m, _ := c.Material(b)
			col := [4]uint8{byte(m.Color), byte(m.Color >> 8), byte(m.Color >> 16), 255}
			if m.TexName != "" && m.Color == 0xFFFFFFFF {
				// untinted textured batch: use a mid-grey so shape is visible
				col = [4]uint8{160, 160, 160, 255}
			}
			jobs = append(jobs, job{b.Strips[0].Verts[0].Z, col, b})
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].z < jobs[j].z })
	px := func(x, y float32) (int, int) {
		return int((x - L.X) * scale), H - 1 - int((y-L.Y)*scale)
	}
	for _, j := range jobs {
		for _, s := range j.b.Strips {
			for i := 0; i+2 < len(s.Verts); i++ {
				x0, y0 := px(s.Verts[i].X, s.Verts[i].Y)
				x1, y1 := px(s.Verts[i+1].X, s.Verts[i+1].Y)
				x2, y2 := px(s.Verts[i+2].X, s.Verts[i+2].Y)
				fillTri(img, x0, y0, x1, y1, x2, y2, j.col)
			}
		}
	}
	// collision edges over the geometry, one colour per layer
	layerCols := [][4]uint8{{255, 0, 0, 255}, {0, 160, 0, 255}, {0, 0, 255, 255}, {255, 0, 255, 255}}
	for li, cl := range c.Collision.Layers {
		col := layerCols[li%len(layerCols)]
		for _, e := range cl.Edges {
			a, b := c.Collision.Points[e[0]], c.Collision.Points[e[1]]
			x0, y0 := px(a[0], a[1])
			x1, y1 := px(b[0], b[1])
			drawLine(img, x0, y0, x1, y1, col)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fillTri(img *image.RGBA, x0, y0, x1, y1, x2, y2 int, col [4]uint8) {
	minX, maxX := min3(x0, x1, x2), max3(x0, x1, x2)
	minY, maxY := min3(y0, y1, y2), max3(y0, y1, y2)
	b := img.Bounds()
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= b.Dx() {
		maxX = b.Dx() - 1
	}
	if maxY >= b.Dy() {
		maxY = b.Dy() - 1
	}
	e := func(ax, ay, bx, by, px, py int) int { return (px-ax)*(by-ay) - (py-ay)*(bx-ax) }
	area := e(x0, y0, x1, y1, x2, y2)
	if area == 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w0, w1, w2 := e(x0, y0, x1, y1, x, y), e(x1, y1, x2, y2, x, y), e(x2, y2, x0, y0, x, y)
			if (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0) {
				i := img.PixOffset(x, y)
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = col[0], col[1], col[2], col[3]
			}
		}
	}
}

// drawLine draws a 1-px Bresenham line, clipped to the image.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col [4]uint8) {
	b := img.Bounds()
	dx, dy := x1-x0, y1-y0
	steps := max3(abs(dx), abs(dy), 1)
	for i := 0; i <= steps; i++ {
		x := x0 + dx*i/steps
		y := y0 + dy*i/steps
		if x < 0 || y < 0 || x >= b.Dx() || y >= b.Dy() {
			continue
		}
		o := img.PixOffset(x, y)
		img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col[0], col[1], col[2], col[3]
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
func max3(a, b, c int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func countNonEmpty(cells []clv.Cell) int {
	n := 0
	for i := range cells {
		if len(cells[i].Batches) > 0 {
			n++
		}
	}
	return n
}

// verifyLog checks every in-image PRIM of a PSP_GE_DEBUG log against the
// decoded strip set.
func verifyLog(c *clv.Clv, base uint32, path string) {
	strips := map[uint32]int{} // vertex-data offset -> vertex count
	for i := range c.Layout.Cells {
		for _, b := range c.Layout.Cells[i].Batches {
			for _, s := range b.Strips {
				strips[s.Off] = len(s.Verts)
			}
		}
	}
	f, err := os.Open(path)
	if err != nil {
		die(err)
	}
	defer f.Close()
	re := regexp.MustCompile(`PRIM#\d+ t(\d) n(\d+) vt=([0-9A-F]+) va=([0-9A-F]+)`)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	total, matched, mismatched := 0, 0, 0
	seen := map[uint32]bool{}
	for sc.Scan() {
		m := re.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		va, _ := strconv.ParseUint(m[4], 16, 32)
		off := uint32(va) - base
		if uint32(va) < base || off >= uint32(len(c.Data)) {
			continue // not a stage-image draw (HUD, characters, ...)
		}
		n, _ := strconv.Atoi(m[2])
		total++
		if seen[off] {
			continue
		}
		seen[off] = true
		if want, ok := strips[off]; ok && want == n {
			matched++
		} else if ok {
			mismatched++
			fmt.Printf("MISMATCH: PRIM va offset %#x has %d verts, decoder says %d\n", off, n, want)
		} else {
			mismatched++
			fmt.Printf("MISS: PRIM va offset %#x (%d verts, t%s vt=%s) not among decoded strips\n", off, n, m[1], m[3])
		}
	}
	fmt.Printf("verify: %d in-image PRIMs, %d distinct strips matched, %d problems (decoder knows %d strips)\n",
		total, matched, mismatched, len(strips))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
