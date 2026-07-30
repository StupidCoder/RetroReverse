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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	tcws := flag.Bool("tcws", false, "print the distinct texture control words the models reference")
	glbin := flag.String("glbin", "", "skip decoding: reopen this existing GLB and render -png from it")
	tex := flag.Bool("tex", false, "texture the export from the course's TEXDC file (course = the digit in -file)")
	texCourse := flag.Int("texcourse", -1, "override the texture course index (default: the digit in -file)")
	dumpTex := flag.String("dumptex", "", "decode every texture of the -file's course into this directory as PNGs")
	flag.Parse()

	if *glbin != "" {
		if *preview == "" {
			die("-glbin wants -png")
		}
		if err := renderGLB(*glbin, *preview); err != nil {
			die("preview: %v", err)
		}
		fmt.Printf("wrote %s (rendered from %s)\n", *preview, *glbin)
		return
	}
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
	nb, nv, maxAux, nTex, noAux := 0, 0, -1, 0, 0
	for _, m := range models {
		nb += len(m.Blocks)
		for _, b := range m.Blocks {
			if b.Textured() {
				nTex++
				if b.Aux == 0xFFFFFFFF {
					noAux++
				} else if int(b.Aux) > maxAux {
					maxAux = int(b.Aux)
				}
			}
			for _, s := range b.Strips {
				nv += len(s.Verts)
			}
		}
	}
	fmt.Printf("%s: %d models, %d blocks (%d textured, aux ids 0..%d, %d without), %d vertex records\n",
		*file, len(models), nb, nTex, maxAux, noAux, nv)
	if *tcws {
		type tex struct{ tcw, tsp uint32 }
		seen := map[tex]int{}
		for _, m := range models {
			for _, b := range m.Blocks {
				if b.Textured() {
					seen[tex{b.TCW, b.TSP}]++
				}
			}
		}
		keys := make([]tex, 0, len(seen))
		for k := range seen {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].tcw&0x1FFFFF < keys[j].tcw&0x1FFFFF })
		for _, k := range keys {
			u := 8 << (k.tsp >> 3 & 7)
			v := 8 << (k.tsp & 7)
			fmt.Printf("tcw %08X vram %06X  tsp %08X %4dx%-4d x%d\n", k.tcw, k.tcw&0x1FFFFF<<3, k.tsp, u, v, seen[k])
		}
		fmt.Printf("%d distinct textures\n", len(keys))
	}
	var td *assets.TexDir
	var texdc []byte
	if *tex || *dumpTex != "" {
		course := *texCourse
		if course < 0 {
			course = courseOf(*file)
		}
		if course < 0 {
			die("-tex: no course digit in %q (use -texcourse)", *file)
		}
		var err error
		if td, err = assets.OpenTexDir(readDisc(disc, "1ST_READ.BIN"), course); err != nil {
			die("%v", err)
		}
		texdc = readDisc(disc, fmt.Sprintf("TEXDC%d.BIN", course))
		fmt.Printf("texture directory: course %d, %d entries, %d payload bytes\n", course, len(td.Entries), len(texdc))
	}
	if *dumpTex != "" {
		if err := dumpTextures(td, texdc, *dumpTex); err != nil {
			die("%v", err)
		}
	}
	if *out == "" {
		return
	}
	if err := export(models, *out, td, texdc); err != nil {
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

// courseOf extracts the course digit from a POLDC<n>/TEXDC<n>/BINC<n> name.
func courseOf(name string) int {
	for i := 0; i < len(name); i++ {
		if name[i] >= '0' && name[i] <= '9' {
			return int(name[i] - '0')
		}
	}
	return -1
}

// dumpTextures decodes the whole directory into dir as NNN.png.
func dumpTextures(td *assets.TexDir, texdc []byte, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ok, fail := 0, 0
	for i := range td.Entries {
		img, err := td.Decode(i, texdc)
		if err != nil {
			fmt.Printf("  texture %d: %v\n", i, err)
			fail++
			continue
		}
		f, err := os.Create(fmt.Sprintf("%s/%03d.png", dir, i))
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		f.Close()
		ok++
	}
	fmt.Printf("decoded %d textures (%d failed) into %s\n", ok, fail, dir)
	return nil
}

// export writes all models into one GLB: shared position/uv/normal arrays.
// With a texture directory, blocks group by their aux texture id into one
// textured primitive per texture; blocks without one (aux 0xFFFFFFFF, or no
// directory) fall back to one colour group per TA list type.
func export(models []*assets.Model, path string, td *assets.TexDir, texdc []byte) error {
	var pos [][3]float32
	var uvs [][2]float32
	var normals [][3]float32
	groups := map[int]*glb.TriGroup{}
	texGroups := map[uint32]*glb.TexturedGroup{}
	images := map[uint32]image.Image{}
	for _, m := range models {
		for _, b := range m.Blocks {
			var tg *glb.TexturedGroup
			if td != nil && b.Textured() && b.Aux != 0xFFFFFFFF && int(b.Aux) < len(td.Entries) {
				tg = texGroups[b.Aux]
				if tg == nil {
					img, ok := images[b.Aux]
					if !ok {
						var err error
						if img, err = td.Decode(int(b.Aux), texdc); err != nil {
							fmt.Printf("  texture %d: %v\n", b.Aux, err)
							img = nil
						}
						images[b.Aux] = img
					}
					if img != nil {
						tg = &glb.TexturedGroup{Image: img, WrapS: 10497, WrapT: 10497, Blend: b.ListType() == 2}
						texGroups[b.Aux] = tg
					}
				}
			}
			var g *glb.TriGroup
			if tg == nil {
				g = groups[b.ListType()]
				if g == nil {
					g = &glb.TriGroup{Color: listColor(b.ListType())}
					if b.ListType() == 2 {
						g.Alpha = 0.6
					}
					groups[b.ListType()] = g
				}
			}
			for _, s := range b.Strips {
				base := len(pos)
				for _, v := range s.Verts {
					pos = append(pos, v.Pos)
					uvs = append(uvs, [2]float32{v.U, v.V})
					normals = append(normals, v.Normal)
				}
				for _, t := range s.Tris() {
					tri := [3]uint32{uint32(base + t[0]), uint32(base + t[1]), uint32(base + t[2])}
					if tg != nil {
						tg.Tris = append(tg.Tris, tri)
					} else {
						g.Tris = append(g.Tris, tri)
					}
				}
			}
		}
	}
	var texList []glb.TexturedGroup
	var auxIDs []int
	for aux := range texGroups {
		auxIDs = append(auxIDs, int(aux))
	}
	sort.Ints(auxIDs)
	for _, aux := range auxIDs {
		if g := texGroups[uint32(aux)]; len(g.Tris) > 0 {
			texList = append(texList, *g)
		}
	}
	var colorGroups []glb.TriGroup
	for _, k := range []int{0, 4, 2, 1, 3} {
		if g := groups[k]; g != nil && len(g.Tris) > 0 {
			colorGroups = append(colorGroups, *g)
		}
	}
	return glb.WriteTexturedLit(path, pos, uvs, normals, texList, colorGroups)
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
			Material   *int           `json:"material"`
		} `json:"primitives"`
	} `json:"meshes"`
	Materials []struct {
		PBR struct {
			BaseColorTexture *struct {
				Index int `json:"index"`
			} `json:"baseColorTexture"`
		} `json:"pbrMetallicRoughness"`
	} `json:"materials"`
	Textures []struct {
		Source int `json:"source"`
	} `json:"textures"`
	Images []struct {
		BufferView int `json:"bufferView"`
	} `json:"images"`
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
	// decode the embedded texture images once, material → image
	matImg := map[int]*image.RGBA{}
	for mi, m := range doc.Materials {
		if m.PBR.BaseColorTexture == nil {
			continue
		}
		src := doc.Textures[m.PBR.BaseColorTexture.Index].Source
		img, err := png.Decode(bytes.NewReader(acc2view(doc, bin, doc.Images[src].BufferView)))
		if err != nil {
			return fmt.Errorf("embedded image %d: %w", src, err)
		}
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
		matImg[mi] = rgba
	}
	type tri struct {
		p     [3][3]float32
		uv    [3][2]float32
		img   *image.RGBA
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
			var uvs [][2]float32
			var pimg *image.RGBA
			if prim.Material != nil {
				pimg = matImg[*prim.Material]
			}
			if ua, ok := prim.Attributes["TEXCOORD_0"]; ok && pimg != nil {
				ub := acc(ua)
				uvs = make([][2]float32, doc.Accessors[ua].Count)
				for i := range uvs {
					uvs[i][0] = math.Float32frombits(binary.LittleEndian.Uint32(ub[i*8:]))
					uvs[i][1] = math.Float32frombits(binary.LittleEndian.Uint32(ub[i*8+4:]))
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
				if uvs != nil {
					t.uv = [3][2]float32{uvs[idx[i]], uvs[idx[i+1]], uvs[idx[i+2]]}
					t.img = pimg
				}
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
		shade := 0.55 + d*0.45
		if t.img != nil {
			fillTriTex(img, q, t.uv, t.img, shade)
		} else {
			g := uint8(70 + d*170)
			fillTri(img, q, color.RGBA{g, g, uint8(min(float32(g)+20, 255)), 255})
		}
	}
	f, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// acc2view returns a raw bufferView's bytes (images use views directly,
// without an accessor).
func acc2view(doc glbDoc, bin []byte, vi int) []byte {
	v := doc.BufferViews[vi]
	return bin[v.ByteOffset : v.ByteOffset+v.ByteLength]
}

// fillTriTex rasterises one textured triangle: barycentric UVs, REPEAT wrap,
// nearest texel, scaled by the light shade.
func fillTriTex(img *image.RGBA, q [3][2]int, uv [3][2]float32, tex *image.RGBA, shade float32) {
	minX := min(q[0][0], min(q[1][0], q[2][0]))
	maxX := max(q[0][0], max(q[1][0], q[2][0]))
	minY := min(q[0][1], min(q[1][1], q[2][1]))
	maxY := max(q[0][1], max(q[1][1], q[2][1]))
	b := img.Bounds()
	minX, minY = max(minX, b.Min.X), max(minY, b.Min.Y)
	maxX, maxY = min(maxX, b.Max.X-1), min(maxY, b.Max.Y-1)
	area := (q[1][0]-q[0][0])*(q[2][1]-q[0][1]) - (q[1][1]-q[0][1])*(q[2][0]-q[0][0])
	if area == 0 {
		return
	}
	tw, th := tex.Bounds().Dx(), tex.Bounds().Dy()
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w0 := (q[1][0]-x)*(q[2][1]-y) - (q[1][1]-y)*(q[2][0]-x)
			w1 := (q[2][0]-x)*(q[0][1]-y) - (q[2][1]-y)*(q[0][0]-x)
			w2 := area - w0 - w1
			if !((w0 >= 0 && w1 >= 0 && w2 >= 0 && area > 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0 && area < 0)) {
				continue
			}
			fa := float32(area)
			u := (float32(w0)*uv[0][0] + float32(w1)*uv[1][0] + float32(w2)*uv[2][0]) / fa
			v := (float32(w0)*uv[0][1] + float32(w1)*uv[1][1] + float32(w2)*uv[2][1]) / fa
			tx := int(u*float32(tw)) % tw
			ty := int(v*float32(th)) % th
			if tx < 0 {
				tx += tw
			}
			if ty < 0 {
				ty += th
			}
			c := tex.RGBAAt(tex.Bounds().Min.X+tx, tex.Bounds().Min.Y+ty)
			if c.A < 128 {
				continue
			}
			img.SetRGBA(x, y, color.RGBA{uint8(float32(c.R) * shade), uint8(float32(c.G) * shade), uint8(float32(c.B) * shade), 255})
		}
	}
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
