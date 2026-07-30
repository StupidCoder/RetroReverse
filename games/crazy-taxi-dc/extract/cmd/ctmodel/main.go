// ctmodel decodes the Crazy Taxi model family (assets/model.go — the
// format read off the game's own renderer) and exports GLBs.
//
//	ctmodel -census                            parse every model on the disc, report coverage
//	ctmodel -file POLDC1.BIN -o out.glb        export a whole POLDC file (all its models)
//	ctmodel -file BINC1.AFS -entry 3 -o o.glb  export one streamed object
//	ctmodel ... -png preview.png               reopen the WRITTEN GLB and render it
//
// The -png preview deliberately reads the exported file back (positions,
// normals, indices out of the GLB's own accessors) rather than re-rendering
// the decoded structs — the shipped artefact is what gets verified.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/crazy-taxi-dc/extract/assets"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/dc"
)

func die(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "ctmodel: "+f+"\n", a...)
	os.Exit(1)
}

func main() {
	image_ := flag.String("image", "image/Crazy Taxi (US).cue", "disc image (.cue)")
	file := flag.String("file", "", "disc file (POLDC*.BIN, or *.AFS with -entry)")
	entry := flag.Int("entry", -1, "AFS entry index")
	out := flag.String("o", "", "write a GLB")
	preview := flag.String("png", "", "reopen the written GLB and render this preview PNG")
	census := flag.Bool("census", false, "parse every model container on the disc and report")
	flag.Parse()

	disc, err := dc.OpenDisc(*image_)
	if err != nil {
		die("%v", err)
	}

	if *census {
		runCensus(disc)
		return
	}
	if *file == "" {
		die("need -file (or -census)")
	}
	data := readDisc(disc, *file)
	var models []*assets.Model
	if strings.HasSuffix(strings.ToUpper(*file), ".AFS") {
		a, err := assets.OpenAFS(data)
		if err != nil {
			die("%v", err)
		}
		if *entry >= 0 {
			if data, err = a.Data(*entry); err != nil {
				die("%v", err)
			}
			if models, err = assets.OpenModels(data); err != nil {
				die("parse: %v", err)
			}
		} else {
			// whole container: the BINC entries are world-placed chunks,
			// so concatenating them assembles the course
			for i := range a.Entries {
				e, err := a.Data(i)
				if err != nil || len(e) == 0 {
					continue
				}
				ms, err := assets.OpenModels(e)
				if err != nil {
					die("entry %d: %v", i, err)
				}
				models = append(models, ms...)
			}
		}
	} else {
		var err error
		if models, err = assets.OpenModels(data); err != nil {
			die("parse: %v (after %d models)", err, len(models))
		}
	}
	nb, nv := 0, 0
	for _, m := range models {
		nb += len(m.Blocks)
		for _, b := range m.Blocks {
			for _, s := range b.Strips {
				nv += len(s.Verts)
			}
		}
	}
	fmt.Printf("%s: %d models, %d blocks, %d vertex records\n", *file, len(models), nb, nv)
	if *out == "" {
		return
	}
	if err := export(models, *out); err != nil {
		die("export: %v", err)
	}
	fmt.Printf("wrote %s\n", *out)
	if *preview != "" {
		if err := renderGLB(*out, *preview); err != nil {
			die("preview: %v", err)
		}
		fmt.Printf("wrote %s (rendered from the GLB file)\n", *preview)
	}
}

func readDisc(disc *dc.Disc, name string) []byte {
	for _, cand := range []string{name, name + ";1"} {
		if data, err := disc.Vol.ReadFile(cand); err == nil {
			return data
		}
	}
	die("no file %q on the disc", name)
	return nil
}

func runCensus(disc *dc.Disc) {
	names := []string{"POLDC0.BIN", "POLDC1.BIN", "POLDC2.BIN", "POLDC3.BIN"}
	for _, n := range names {
		models, err := assets.OpenModels(readDisc(disc, n))
		nb, nv, counted := 0, 0, 0
		for _, m := range models {
			if m.Counted {
				counted++
			}
			nb += len(m.Blocks)
			for _, b := range m.Blocks {
				for _, s := range b.Strips {
					nv += len(s.Verts)
				}
			}
		}
		fmt.Printf("%-11s %5d models (%d counted), %6d blocks, %7d verts, err=%v\n", n, len(models), counted, nb, nv, err)
	}
	for _, n := range []string{"BINC1.AFS", "BINC2.AFS", "BINC3.AFS"} {
		a, err := assets.OpenAFS(readDisc(disc, n))
		if err != nil {
			fmt.Printf("%s: %v\n", n, err)
			continue
		}
		ok, fail, nb, nv := 0, 0, 0, 0
		for i := range a.Entries {
			e, err := a.Data(i)
			if err != nil || len(e) == 0 {
				continue
			}
			ms, err := assets.OpenModels(e)
			if err != nil {
				fail++
				fmt.Printf("  entry %d: %v\n", i, err)
				continue
			}
			ok++
			for _, m := range ms {
				nb += len(m.Blocks)
				for _, b := range m.Blocks {
					for _, s := range b.Strips {
						nv += len(s.Verts)
					}
				}
			}
		}
		fmt.Printf("%-11s %5d entries ok, %d fail, %6d blocks, %7d verts\n", n, ok, fail, nb, nv)
	}
}

// export writes all models into one GLB: shared position/uv/normal arrays,
// one colour group per TA list type (opaque / punch-through / translucent) —
// texture mapping is a later pass, the geometry is what this proves.
func export(models []*assets.Model, path string) error {
	var pos [][3]float32
	var uvs [][2]float32
	var normals [][3]float32
	groups := map[int]*glb.TriGroup{}
	for _, m := range models {
		for _, b := range m.Blocks {
			g := groups[b.ListType()]
			if g == nil {
				g = &glb.TriGroup{Color: listColor(b.ListType())}
				if b.ListType() == 2 {
					g.Alpha = 0.6
				}
				groups[b.ListType()] = g
			}
			for _, s := range b.Strips {
				base := len(pos)
				for _, v := range s.Verts {
					pos = append(pos, v.Pos)
					uvs = append(uvs, [2]float32{v.U, v.V})
					normals = append(normals, v.Normal)
				}
				for _, t := range s.Tris() {
					g.Tris = append(g.Tris, [3]uint32{uint32(base + t[0]), uint32(base + t[1]), uint32(base + t[2])})
				}
			}
		}
	}
	var colorGroups []glb.TriGroup
	for _, k := range []int{0, 4, 2, 1, 3} {
		if g := groups[k]; g != nil && len(g.Tris) > 0 {
			colorGroups = append(colorGroups, *g)
		}
	}
	_ = uvs
	return glb.WriteTexturedLit(path, pos, uvs, normals, nil, colorGroups)
}

func listColor(list int) [3]float32 {
	switch list {
	case 0:
		return [3]float32{0.75, 0.75, 0.78} // opaque
	case 2:
		return [3]float32{0.4, 0.6, 0.9} // translucent
	case 4:
		return [3]float32{0.8, 0.7, 0.4} // punch-through
	}
	return [3]float32{0.9, 0.3, 0.9}
}

// ---- GLB reopen + preview render (the shipped file, not the structs) ----

type glbDoc struct {
	Accessors []struct {
		BufferView    int    `json:"bufferView"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
	} `json:"accessors"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
	} `json:"bufferViews"`
	Meshes []struct {
		Primitives []struct {
			Attributes map[string]int `json:"attributes"`
			Indices    *int           `json:"indices"`
		} `json:"primitives"`
	} `json:"meshes"`
}

func renderGLB(glbPath, pngPath string) error {
	raw, err := os.ReadFile(glbPath)
	if err != nil {
		return err
	}
	if len(raw) < 20 || string(raw[:4]) != "glTF" {
		return fmt.Errorf("%s: not a GLB", glbPath)
	}
	jlen := int(binary.LittleEndian.Uint32(raw[12:]))
	var doc glbDoc
	if err := json.Unmarshal(raw[20:20+jlen], &doc); err != nil {
		return fmt.Errorf("GLB json: %w", err)
	}
	binOff := 20 + jlen + 8
	if binOff > len(raw) {
		return fmt.Errorf("GLB has no BIN chunk")
	}
	bin := raw[binOff:]
	acc := func(i int) []byte {
		v := doc.BufferViews[doc.Accessors[i].BufferView]
		return bin[v.ByteOffset : v.ByteOffset+v.ByteLength]
	}
	type tri struct {
		p     [3][3]float32
		n     [3]float32
		depth float32
	}
	var tris []tri
	for _, mesh := range doc.Meshes {
		for _, prim := range mesh.Primitives {
			pa, ok := prim.Attributes["POSITION"]
			if !ok || prim.Indices == nil {
				continue
			}
			pb := acc(pa)
			np := doc.Accessors[pa].Count
			pos := make([][3]float32, np)
			for i := range pos {
				for c := 0; c < 3; c++ {
					pos[i][c] = math.Float32frombits(binary.LittleEndian.Uint32(pb[i*12+c*4:]))
				}
			}
			ib := acc(*prim.Indices)
			ct := doc.Accessors[*prim.Indices].ComponentType
			n := doc.Accessors[*prim.Indices].Count
			idx := make([]int, n)
			for i := range idx {
				switch ct {
				case 5125:
					idx[i] = int(binary.LittleEndian.Uint32(ib[i*4:]))
				case 5123:
					idx[i] = int(binary.LittleEndian.Uint16(ib[i*2:]))
				default:
					return fmt.Errorf("index componentType %d", ct)
				}
			}
			for i := 0; i+2 < n; i += 3 {
				t := tri{p: [3][3]float32{pos[idx[i]], pos[idx[i+1]], pos[idx[i+2]]}}
				tris = append(tris, t)
			}
		}
	}
	if len(tris) == 0 {
		return fmt.Errorf("GLB contains no triangles")
	}
	// dimetric projection: x' = x - z*0.5, y' = -y + (x+z)*0.25
	lo := [2]float32{math.MaxFloat32, math.MaxFloat32}
	hi := [2]float32{-math.MaxFloat32, -math.MaxFloat32}
	project := func(p [3]float32) [2]float32 {
		return [2]float32{p[0] - p[2]*0.5, -p[1] + (p[0]+p[2])*0.25}
	}
	for i := range tris {
		for _, p := range tris[i].p {
			q := project(p)
			for c := 0; c < 2; c++ {
				lo[c] = min(lo[c], q[c])
				hi[c] = max(hi[c], q[c])
			}
			tris[i].depth += (p[0] + p[2]) / 3
		}
		a, b, c := tris[i].p[0], tris[i].p[1], tris[i].p[2]
		u := [3]float32{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
		v := [3]float32{c[0] - a[0], c[1] - a[1], c[2] - a[2]}
		nx := u[1]*v[2] - u[2]*v[1]
		ny := u[2]*v[0] - u[0]*v[2]
		nz := u[0]*v[1] - u[1]*v[0]
		l := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if l > 0 {
			tris[i].n = [3]float32{nx / l, ny / l, nz / l}
		}
	}
	sort.Slice(tris, func(i, j int) bool { return tris[i].depth < tris[j].depth })
	const W, H = 1024, 768
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for i := range img.Pix {
		img.Pix[i] = 24
		if i%4 == 3 {
			img.Pix[i] = 255
		}
	}
	sx := float32(W-40) / (hi[0] - lo[0])
	sy := float32(H-40) / (hi[1] - lo[1])
	s := min(sx, sy)
	light := [3]float32{0.44, 0.77, 0.44}
	for _, t := range tris {
		var q [3][2]int
		for i, p := range t.p {
			pr := project(p)
			q[i] = [2]int{int((pr[0]-lo[0])*s) + 20, int((pr[1]-lo[1])*s) + 20}
		}
		d := t.n[0]*light[0] + t.n[1]*light[1] + t.n[2]*light[2]
		if d < 0 {
			d = -d
		}
		g := uint8(70 + d*170)
		fillTri(img, q, color.RGBA{g, g, uint8(min(float32(g)+20, 255)), 255})
	}
	f, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fillTri(img *image.RGBA, q [3][2]int, c color.RGBA) {
	minX := min(q[0][0], min(q[1][0], q[2][0]))
	maxX := max(q[0][0], max(q[1][0], q[2][0]))
	minY := min(q[0][1], min(q[1][1], q[2][1]))
	maxY := max(q[0][1], max(q[1][1], q[2][1]))
	b := img.Bounds()
	minX, minY = max(minX, b.Min.X), max(minY, b.Min.Y)
	maxX, maxY = min(maxX, b.Max.X-1), min(maxY, b.Max.Y-1)
	edge := func(ax, ay, bx, by, px, py int) int { return (bx-ax)*(py-ay) - (by-ay)*(px-ax) }
	a0 := edge(q[0][0], q[0][1], q[1][0], q[1][1], q[2][0], q[2][1])
	if a0 == 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w0 := edge(q[0][0], q[0][1], q[1][0], q[1][1], x, y)
			w1 := edge(q[1][0], q[1][1], q[2][0], q[2][1], x, y)
			w2 := edge(q[2][0], q[2][1], q[0][0], q[0][1], x, y)
			if (w0 >= 0 && w1 >= 0 && w2 >= 0 && a0 > 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0 && a0 < 0) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}
