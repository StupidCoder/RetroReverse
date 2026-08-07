package main

// rendersheet: parse a GLB (its own buffers) and software-render a sheet of
// views: rows = {all nodes, node0, node1, ...}, cols = yaw angles.
// Textured: samples baseColorTexture (nearest, REPEAT) x COLOR_0 x shade.
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
)

type gltf struct {
	Nodes []struct {
		Mesh        *int       `json:"mesh"`
		Skin        *int       `json:"skin"`
		Name        string     `json:"name"`
		Children    []int      `json:"children"`
		Translation []float64  `json:"translation"`
		Rotation    []float64  `json:"rotation"`
		Scale       []float64  `json:"scale"`
	} `json:"nodes"`
	Skins []struct {
		Joints              []int `json:"joints"`
		InverseBindMatrices int   `json:"inverseBindMatrices"`
	} `json:"skins"`
	Animations []struct {
		Name     string `json:"name"`
		Channels []struct {
			Sampler int `json:"sampler"`
			Target  struct {
				Node int    `json:"node"`
				Path string `json:"path"`
			} `json:"target"`
		} `json:"channels"`
		Samplers []struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		} `json:"samplers"`
	} `json:"animations"`
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
			BaseColorFactor []float64 `json:"baseColorFactor"`
		} `json:"pbrMetallicRoughness"`
		Extras map[string]any `json:"extras"`
	} `json:"materials"`
	Textures []struct {
		Source int `json:"source"`
	} `json:"textures"`
	Images []struct {
		BufferView int `json:"bufferView"`
	} `json:"images"`
	Accessors []struct {
		BufferView    *int   `json:"bufferView"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
		ByteOffset    int    `json:"byteOffset"`
		Normalized    bool   `json:"normalized"`
	} `json:"accessors"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
		ByteStride int `json:"byteStride"`
	} `json:"bufferViews"`
}

type tri struct {
	p        [3][3]float64
	uv       [3][2]float64
	col      [3][4]float64
	tex      *image.RGBA
	node     int
	additive bool
}

var g gltf
var bin []byte

func compSize(t int) int {
	switch t {
	case 5120, 5121:
		return 1
	case 5122, 5123:
		return 2
	default:
		return 4
	}
}

func accVec(ai, n int) [][]float64 {
	a := g.Accessors[ai]
	if os.Getenv("DBG") != "" {
		fmt.Printf("accVec ai=%d type=%s comp=%d count=%d bv=%v\n", ai, a.Type, a.ComponentType, a.Count, *a.BufferView)
	}
	bv := g.BufferViews[*a.BufferView]
	cs := compSize(a.ComponentType)
	stride := bv.ByteStride
	if stride == 0 {
		stride = cs * n
	}
	base := bv.ByteOffset + a.ByteOffset
	out := make([][]float64, a.Count)
	for i := 0; i < a.Count; i++ {
		o := base + i*stride
		v := make([]float64, n)
		for k := 0; k < n; k++ {
			switch a.ComponentType {
			case 5126:
				v[k] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bin[o+k*4:])))
			case 5121:
				x := float64(bin[o+k])
				if a.Normalized {
					x /= 255
				}
				v[k] = x
			case 5123:
				x := float64(binary.LittleEndian.Uint16(bin[o+k*2:]))
				if a.Normalized {
					x /= 65535
				}
				v[k] = x
			}
		}
		out[i] = v
	}
	return out
}

func accIdx(ai int) []int {
	a := g.Accessors[ai]
	bv := g.BufferViews[*a.BufferView]
	base := bv.ByteOffset + a.ByteOffset
	out := make([]int, a.Count)
	for i := 0; i < a.Count; i++ {
		if a.ComponentType == 5123 {
			out[i] = int(binary.LittleEndian.Uint16(bin[base+i*2:]))
		} else {
			out[i] = int(binary.LittleEndian.Uint32(bin[base+i*4:]))
		}
	}
	return out
}

func main() {
	raw, err := os.ReadFile(os.Args[1])
	check(err)
	out := os.Args[2]
	jlen := binary.LittleEndian.Uint32(raw[12:])
	check(json.Unmarshal(raw[20:20+jlen], &g))
	blen := binary.LittleEndian.Uint32(raw[20+jlen:])
	bin = raw[20+jlen+8 : 20+jlen+8+blen]

	texs := map[int]*image.RGBA{}
	getTex := func(mi *int) *image.RGBA {
		if mi == nil {
			return nil
		}
		m := g.Materials[*mi]
		if m.PBR.BaseColorTexture == nil {
			return nil
		}
		ti := m.PBR.BaseColorTexture.Index
		if im, ok := texs[ti]; ok {
			return im
		}
		bv := g.BufferViews[g.Images[g.Textures[ti].Source].BufferView]
		img, err := png.Decode(bytes.NewReader(bin[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]))
		check(err)
		r := image.NewRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				r.Set(x, y, img.At(x, y))
			}
		}
		texs[ti] = r
		return r
	}

	only := -1
	if v := os.Getenv("PRIM"); v != "" {
		fmt.Sscanf(v, "%d", &only)
	}
	flatMode = os.Getenv("FLAT") != ""
	// Node world transforms are always composed and applied to non-skinned
	// meshes (the character exports animate plain node TRS, no skin);
	// POSE=clipname:frame overrides the locals from a clip first. Skinned
	// meshes ignore their node transform per glTF and pose via jointMats.
	jointMats := map[int][]mat4{} // skin index -> per-joint vertex matrices
	var clip string
	frame := 0
	if pv := os.Getenv("POSE"); pv != "" {
		fmt.Sscanf(pv, "%s", &clip)
		if i := len(clip) - 1; i > 0 {
			for k := range clip {
				if clip[k] == ':' {
					fmt.Sscanf(clip[k+1:], "%d", &frame)
					clip = clip[:k]
					break
				}
			}
		}
	}
	// node local TRS (possibly overridden by the clip at the frame)
	type trs struct {
		t [3]float64
		q [4]float64
		s [3]float64
	}
	locals := make([]trs, len(g.Nodes))
	for ni, n := range g.Nodes {
		l := trs{q: [4]float64{0, 0, 0, 1}, s: [3]float64{1, 1, 1}}
		if len(n.Translation) == 3 {
			copy(l.t[:], n.Translation)
		}
		if len(n.Rotation) == 4 {
			copy(l.q[:], n.Rotation)
		}
		if len(n.Scale) == 3 {
			copy(l.s[:], n.Scale)
		}
		locals[ni] = l
	}
	if clip != "" {
		for _, an := range g.Animations {
			if clip == "BIND" || an.Name != clip {
				continue
			}
			for _, ch := range an.Channels {
				out := an.Samplers[ch.Sampler].Output
				f := frame
				if f >= g.Accessors[out].Count {
					f = g.Accessors[out].Count - 1
				}
				switch ch.Target.Path {
				case "rotation":
					v := accVec(out, 4)[f]
					copy(locals[ch.Target.Node].q[:], v)
				case "translation":
					v := accVec(out, 3)[f]
					copy(locals[ch.Target.Node].t[:], v)
				case "scale":
					v := accVec(out, 3)[f]
					copy(locals[ch.Target.Node].s[:], v)
				}
			}
		}
	}
	// compose world matrices (column-major, parent * local)
	nodeWorld := make([]mat4, len(g.Nodes))
	haveW := make([]bool, len(g.Nodes))
	parent := make([]int, len(g.Nodes))
	for i := range parent {
		parent[i] = -1
	}
	for ni, n := range g.Nodes {
		for _, c := range n.Children {
			parent[c] = ni
		}
	}
	var compose func(ni int) mat4
	compose = func(ni int) mat4 {
		if haveW[ni] {
			return nodeWorld[ni]
		}
		m := trsMat(locals[ni].t, locals[ni].q, locals[ni].s)
		if parent[ni] >= 0 {
			m = matMul(compose(parent[ni]), m)
		}
		nodeWorld[ni] = m
		haveW[ni] = true
		return m
	}
	for si, sk := range g.Skins {
		ib := accVec(sk.InverseBindMatrices, 16)
		jm := make([]mat4, len(sk.Joints))
		for k, jn := range sk.Joints {
			var ibm mat4
			copy(ibm[:], ib[k])
			jm[k] = matMul(compose(jn), ibm)
			if os.Getenv("DBGJM") != "" && (k == 3 || k == 10 || k == 30) {
				fmt.Printf("jm[%d] = %.3f\n", k, jm[k])
			}
		}
		jointMats[si] = jm
	}

	var tris []tri
	var nodeIDs []int
	for ni, n := range g.Nodes {
		if n.Mesh == nil {
			continue
		}
		nodeIDs = append(nodeIDs, ni)
		for pi, pr := range g.Meshes[*n.Mesh].Primitives {
			if only >= 0 && pi != only {
				continue
			}
			pos := accVec(pr.Attributes["POSITION"], 3)
			if n.Skin == nil {
				w := compose(ni)
				for vi := range pos {
					pos[vi] = matApply(w, pos[vi])
				}
			}
			if n.Skin != nil {
				if jm, ok := jointMats[*n.Skin]; ok {
					if ji, ok2 := pr.Attributes["JOINTS_0"]; ok2 {
						js := accVec(ji, 4)
						var ws [][]float64
						if wi, ok3 := pr.Attributes["WEIGHTS_0"]; ok3 {
							ws = accVec(wi, 4)
						}
						for vi := range pos {
							if ws == nil {
								// single influence, weight 1
								if j := int(js[vi][0]); j < len(jm) {
									pos[vi] = matApply(jm[j], pos[vi])
								}
								continue
							}
							// linear blend skinning over up to 4 joints
							var out [3]float64
							for k := 0; k < 4; k++ {
								w := ws[vi][k]
								if w == 0 {
									continue
								}
								j := int(js[vi][k])
								if j >= len(jm) {
									continue
								}
								p := matApply(jm[j], pos[vi])
								for c := 0; c < 3; c++ {
									out[c] += w * p[c]
								}
							}
							pos[vi] = out[:]
						}
					}
				}
			}
			if os.Getenv("DBGBOX") != "" {
				mn := []float64{1e30, 1e30, 1e30}
				mx := []float64{-1e30, -1e30, -1e30}
				for _, p := range pos {
					for c := 0; c < 3; c++ {
						if p[c] < mn[c] {
							mn[c] = p[c]
						}
						if p[c] > mx[c] {
							mx[c] = p[c]
						}
					}
				}
				fmt.Printf("dbgbox node %d %q prim %d skin=%v bbox (%.2f %.2f %.2f)..(%.2f %.2f %.2f)\n",
					ni, n.Name, pi, n.Skin != nil, mn[0], mn[1], mn[2], mx[0], mx[1], mx[2])
			}
			var uvs, cols [][]float64
			if ai, ok := pr.Attributes["TEXCOORD_0"]; ok {
				uvs = accVec(ai, 2)
			}
			if ai, ok := pr.Attributes["COLOR_0"]; ok {
				cols = accVec(ai, 4)
			}
			tex := getTex(pr.Material)
			additive := false
			if pr.Material != nil {
				if b, ok := g.Materials[*pr.Material].Extras["blend"].(string); ok && b == "additive" {
					additive = true
				}
			}
			idx := accIdx(*pr.Indices)
			for i := 0; i+2 < len(idx); i += 3 {
				var t tri
				t.tex = tex
				t.node = ni
				t.additive = additive
				for k := 0; k < 3; k++ {
					v := idx[i+k]
					copy(t.p[k][:], pos[v])
					if uvs != nil {
						copy(t.uv[k][:], uvs[v])
					}
					if cols != nil {
						copy(t.col[k][:], cols[v])
					} else {
						t.col[k] = [4]float64{1, 1, 1, 1}
					}
				}
				tris = append(tris, t)
			}
		}
	}
	fmt.Printf("%d tris, %d nodes, %d textures\n", len(tris), len(nodeIDs), len(g.Textures))

	angles := []float64{0, 40, 90}
	rows := 1 + len(nodeIDs)
	const W, H = 460, 320
	sheet := image.NewRGBA(image.Rect(0, 0, W*len(angles), H*rows))
	for row := 0; row < rows; row++ {
		var sel []tri
		if row == 0 {
			sel = tris
		} else {
			for _, t := range tris {
				if t.node == nodeIDs[row-1] {
					sel = append(sel, t)
				}
			}
		}
		for col, ang := range angles {
			renderPanel(sheet, col*W, row*H, W, H, sel, ang*math.Pi/180)
		}
	}
	f, err := os.Create(out)
	check(err)
	check(png.Encode(f, sheet))
	check(f.Close())
	fmt.Println("wrote", out)
}

var flatMode bool

type mat4 [16]float64 // column-major, glTF order

func matMul(a, b mat4) (r mat4) {
	for c := 0; c < 4; c++ {
		for w := 0; w < 4; w++ {
			var s float64
			for k := 0; k < 4; k++ {
				s += a[k*4+w] * b[c*4+k]
			}
			r[c*4+w] = s
		}
	}
	return
}

func matApply(m mat4, v []float64) []float64 {
	x, y, z := v[0], v[1], v[2]
	return []float64{
		m[0]*x + m[4]*y + m[8]*z + m[12],
		m[1]*x + m[5]*y + m[9]*z + m[13],
		m[2]*x + m[6]*y + m[10]*z + m[14],
	}
}

func trsMat(t [3]float64, q [4]float64, s [3]float64) mat4 {
	x, y, z, w := q[0], q[1], q[2], q[3]
	var m mat4
	m[0] = (1 - 2*(y*y+z*z)) * s[0]
	m[1] = (2 * (x*y + z*w)) * s[0]
	m[2] = (2 * (x*z - y*w)) * s[0]
	m[4] = (2 * (x*y - z*w)) * s[1]
	m[5] = (1 - 2*(x*x+z*z)) * s[1]
	m[6] = (2 * (y*z + x*w)) * s[1]
	m[8] = (2 * (x*z + y*w)) * s[2]
	m[9] = (2 * (y*z - x*w)) * s[2]
	m[10] = (1 - 2*(x*x+y*y)) * s[2]
	m[12], m[13], m[14] = t[0], t[1], t[2]
	m[15] = 1
	return m
}

type ptri struct {
	x, y  [3]float64
	z     float64
	shade float64
	t     *tri
}

func renderPanel(img *image.RGBA, ox, oy, w, h int, tris []tri, yaw float64) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(ox+x, oy+y, color.RGBA{22, 22, 30, 255})
		}
	}
	if len(tris) == 0 {
		return
	}
	sy, cy := math.Sin(yaw), math.Cos(yaw)
	rot := func(p [3]float64) [3]float64 {
		return [3]float64{p[0]*cy + p[2]*sy, p[1], -p[0]*sy + p[2]*cy}
	}
	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)
	for _, t := range tris {
		for _, p := range t.p {
			r := rot(p)
			minX, maxX = math.Min(minX, r[0]), math.Max(maxX, r[0])
			minY, maxY = math.Min(minY, r[1]), math.Max(maxY, r[1])
		}
	}
	scale := math.Min(float64(w-30)/(maxX-minX+1e-6), float64(h-30)/(maxY-minY+1e-6))
	cx, cyy := (minX+maxX)/2, (minY+maxY)/2
	light := [3]float64{0.35, 0.55, 0.76}
	var pts []ptri
	for i := range tris {
		t := &tris[i]
		var pt ptri
		pt.t = t
		var r [3][3]float64
		for k := 0; k < 3; k++ {
			r[k] = rot(t.p[k])
			pt.x[k] = float64(ox) + float64(w)/2 + (r[k][0]-cx)*scale
			pt.y[k] = float64(oy) + float64(h)/2 - (r[k][1]-cyy)*scale
			pt.z += r[k][2] / 3
		}
		ax, ay, az := r[1][0]-r[0][0], r[1][1]-r[0][1], r[1][2]-r[0][2]
		bx, by, bz := r[2][0]-r[0][0], r[2][1]-r[0][1], r[2][2]-r[0][2]
		nx, ny, nz := ay*bz-az*by, az*bx-ax*bz, ax*by-ay*bx
		nl := math.Sqrt(nx*nx + ny*ny + nz*nz)
		if nl < 1e-9 {
			continue
		}
		d := (nx*light[0] + ny*light[1] + nz*light[2]) / nl
		pt.shade = 0.42 + 0.58*math.Abs(d)
		if flatMode {
			pt.shade = 1
			pt.t.tex = nil
		}
		pts = append(pts, pt)
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].z < pts[j].z })
	for _, pt := range pts {
		fillTri(img, &pt)
	}
}

func fillTri(img *image.RGBA, pt *ptri) {
	xs, ys := pt.x, pt.y
	minx := int(math.Floor(math.Min(xs[0], math.Min(xs[1], xs[2]))))
	maxx := int(math.Ceil(math.Max(xs[0], math.Max(xs[1], xs[2]))))
	miny := int(math.Floor(math.Min(ys[0], math.Min(ys[1], ys[2]))))
	maxy := int(math.Ceil(math.Max(ys[0], math.Max(ys[1], ys[2]))))
	d := (ys[1]-ys[2])*(xs[0]-xs[2]) + (xs[2]-xs[1])*(ys[0]-ys[2])
	if math.Abs(d) < 1e-9 {
		return
	}
	t := pt.t
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			w0 := ((ys[1]-ys[2])*(px-xs[2]) + (xs[2]-xs[1])*(py-ys[2])) / d
			w1 := ((ys[2]-ys[0])*(px-xs[2]) + (xs[0]-xs[2])*(py-ys[2])) / d
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			cr, cg, cb := 0.83, 0.68, 0.28
			if flatMode {
				cr, cg, cb = 1, 1, 1
			}
			ta := 1.0
			if t.tex != nil {
				u := w0*t.uv[0][0] + w1*t.uv[1][0] + w2*t.uv[2][0]
				v := w0*t.uv[0][1] + w1*t.uv[1][1] + w2*t.uv[2][1]
				b := t.tex.Bounds()
				tw, th := b.Dx(), b.Dy()
				tx := int(math.Floor(u*float64(tw))) % tw
				ty := int(math.Floor(v*float64(th))) % th
				if tx < 0 {
					tx += tw
				}
				if ty < 0 {
					ty += th
				}
				c := t.tex.RGBAAt(b.Min.X+tx, b.Min.Y+ty)
				cr, cg, cb = float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
				ta = float64(c.A) / 255
			}
			vr := w0*t.col[0][0] + w1*t.col[1][0] + w2*t.col[2][0]
			vg := w0*t.col[0][1] + w1*t.col[1][1] + w2*t.col[2][1]
			vb := w0*t.col[0][2] + w1*t.col[1][2] + w2*t.col[2][2]
			va := w0*t.col[0][3] + w1*t.col[1][3] + w2*t.col[2][3]
			a := ta * va
			if a < 0.02 {
				continue
			}
			s := pt.shade
			d := img.RGBAAt(x, y)
			var nr, ng, nb float64
			if t.additive {
				nr = float64(d.R)/255 + cr*vr*a
				ng = float64(d.G)/255 + cg*vg*a
				nb = float64(d.B)/255 + cb*vb*a
			} else {
				nr = cr*vr*s*a + float64(d.R)/255*(1-a)
				ng = cg*vg*s*a + float64(d.G)/255*(1-a)
				nb = cb*vb*s*a + float64(d.B)/255*(1-a)
			}
			img.Set(x, y, color.RGBA{clamp255(nr), clamp255(ng), clamp255(nb), 255})
		}
	}
}

func clamp255(v float64) uint8 {
	x := int(v * 255)
	if x > 255 {
		x = 255
	}
	if x < 0 {
		x = 0
	}
	return uint8(x)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
