// stageverify renders an exported stage GLB through the GAME'S OWN transform
// — the composite MVP rows (c160-163) and viewport constants (c58/c59)
// captured off a live draw with bootoracle -vshdump — and writes the result
// next to the game's captured frame for comparison. This is the course-side
// version of the standard the cars met: the exported file, opened as a
// viewer would open it, judged against the game's own render of the scene.
//
// Usage:
//
//	stageverify -glb stage-beac.glb -o out.png \
//	  -mvp "r0x r0y r0z r0w r1x ... r3w" -vp "sx sy ox oy"
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type gltf struct {
	Nodes []struct {
		Mesh *int   `json:"mesh"`
		Name string `json:"name"`
	} `json:"nodes"`
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
		AlphaMode string         `json:"alphaMode"`
		Extras    map[string]any `json:"extras"`
	} `json:"materials"`
	Textures []struct {
		Source int `json:"source"`
	} `json:"textures"`
	Images []struct {
		BufferView int `json:"bufferView"`
	} `json:"images"`
	Accessors []struct {
		BufferView    *int   `json:"bufferView"`
		ByteOffset    int    `json:"byteOffset"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
		Normalized    bool   `json:"normalized"`
	} `json:"accessors"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
		ByteStride int `json:"byteStride"`
	} `json:"bufferViews"`
}

var (
	g   gltf
	bin []byte
)

func accF(ai, n int) [][]float64 {
	a := g.Accessors[ai]
	bv := g.BufferViews[*a.BufferView]
	cs := 4
	if a.ComponentType == 5121 {
		cs = 1
	}
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
			if a.ComponentType == 5126 {
				v[k] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bin[o+k*4:])))
			} else {
				x := float64(bin[o+k])
				if a.Normalized {
					x /= 255
				}
				v[k] = x
			}
		}
		out[i] = v
	}
	return out
}

func accI(ai int) []int {
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

type vtx struct {
	sx, sy float64 // screen
	iw     float64 // 1/clip.w
	z      float64 // clip.z/clip.w
	u, v   float64 // uv * iw (perspective-corrected numerators)
	r, g, b, a float64 // color * iw
}

type tri struct {
	v        [3]vtx
	tex      *image.NRGBA
	blend    bool
	additive bool
	mask     bool
	zc       float64
}

func main() {
	glbPath := flag.String("glb", "", "exported GLB to open")
	out := flag.String("o", "verify.png", "output PNG")
	mvpS := flag.String("mvp", "", "16 floats: c160..c163 rows (clip = row·pos)")
	vpS := flag.String("vp", "320 -240 320.53125 240.53125", "viewport sx sy ox oy (c58.xy c59.xy)")
	w := flag.Int("w", 640, "width")
	h := flag.Int("h", 480, "height")
	flag.Parse()

	var mvp [16]float64
	for i, tok := range strings.Fields(*mvpS) {
		f, err := strconv.ParseFloat(tok, 64)
		if err != nil || i >= 16 {
			fmt.Fprintln(os.Stderr, "stageverify: bad -mvp")
			os.Exit(2)
		}
		mvp[i] = f
	}
	var vp [4]float64
	for i, tok := range strings.Fields(*vpS) {
		f, _ := strconv.ParseFloat(tok, 64)
		if i < 4 {
			vp[i] = f
		}
	}

	raw, err := os.ReadFile(*glbPath)
	if err != nil {
		panic(err)
	}
	jlen := binary.LittleEndian.Uint32(raw[12:])
	if err := json.Unmarshal(raw[20:20+jlen], &g); err != nil {
		panic(err)
	}
	blen := binary.LittleEndian.Uint32(raw[20+jlen:])
	bin = raw[20+jlen+8 : 20+jlen+8+blen]

	texs := map[int]*image.NRGBA{}
	getTex := func(mi *int) *image.NRGBA {
		if mi == nil || g.Materials[*mi].PBR.BaseColorTexture == nil {
			return nil
		}
		ti := g.Materials[*mi].PBR.BaseColorTexture.Index
		if im, ok := texs[ti]; ok {
			return im
		}
		bv := g.BufferViews[g.Images[g.Textures[ti].Source].BufferView]
		img, err := png.Decode(bytes.NewReader(bin[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]))
		if err != nil {
			panic(err)
		}
		r := image.NewNRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				r.Set(x, y, img.At(x, y))
			}
		}
		texs[ti] = r
		return r
	}

	var opaque, blended []tri
	for _, n := range g.Nodes {
		if n.Mesh == nil {
			continue
		}
		for _, pr := range g.Meshes[*n.Mesh].Primitives {
			pos := accF(pr.Attributes["POSITION"], 3)
			var uvs, cols [][]float64
			if ai, ok := pr.Attributes["TEXCOORD_0"]; ok {
				uvs = accF(ai, 2)
			}
			if ai, ok := pr.Attributes["COLOR_0"]; ok {
				cols = accF(ai, 4)
			}
			tex := getTex(pr.Material)
			blend, additive, mask := false, false, false
			if pr.Material != nil {
				m := g.Materials[*pr.Material]
				blend = m.AlphaMode == "BLEND"
				mask = m.AlphaMode == "MASK"
				if b, ok := m.Extras["blend"].(string); ok && b == "additive" {
					additive = true
				}
			}
			idx := accI(*pr.Indices)
			// pre-transform vertices of this primitive
			vt := make([]vtx, len(pos))
			okv := make([]bool, len(pos))
			for i, p := range pos {
				cx := mvp[0]*p[0] + mvp[1]*p[1] + mvp[2]*p[2] + mvp[3]
				cy := mvp[4]*p[0] + mvp[5]*p[1] + mvp[6]*p[2] + mvp[7]
				cz := mvp[8]*p[0] + mvp[9]*p[1] + mvp[10]*p[2] + mvp[11]
				cw := mvp[12]*p[0] + mvp[13]*p[1] + mvp[14]*p[2] + mvp[15]
				if cw < 0.01 {
					continue
				}
				iw := 1 / cw
				v := vtx{
					sx: cx*iw*vp[0] + vp[2],
					sy: cy*iw*vp[1] + vp[3],
					iw: iw,
					z:  cz * iw,
				}
				if uvs != nil {
					v.u, v.v = uvs[i][0]*iw, uvs[i][1]*iw
				}
				if cols != nil {
					v.r, v.g, v.b, v.a = cols[i][0]*iw, cols[i][1]*iw, cols[i][2]*iw, cols[i][3]*iw
				} else {
					v.r, v.g, v.b, v.a = iw, iw, iw, iw
				}
				vt[i] = v
				okv[i] = true
			}
			for i := 0; i+2 < len(idx); i += 3 {
				a, b, c := idx[i], idx[i+1], idx[i+2]
				if !okv[a] || !okv[b] || !okv[c] {
					continue
				}
				t := tri{v: [3]vtx{vt[a], vt[b], vt[c]}, tex: tex, blend: blend, additive: additive, mask: mask}
				t.zc = (vt[a].z + vt[b].z + vt[c].z) / 3
				if blend || additive {
					blended = append(blended, t)
				} else {
					opaque = append(opaque, t)
				}
			}
		}
	}

	W, H := *w, *h
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for i := range img.Pix {
		img.Pix[i] = 0
	}
	zbuf := make([]float64, W*H)
	for i := range zbuf {
		zbuf[i] = math.Inf(1)
	}
	raster := func(t *tri, zwrite bool) {
		xs := [3]float64{t.v[0].sx, t.v[1].sx, t.v[2].sx}
		ys := [3]float64{t.v[0].sy, t.v[1].sy, t.v[2].sy}
		d := (ys[1]-ys[2])*(xs[0]-xs[2]) + (xs[2]-xs[1])*(ys[0]-ys[2])
		if math.Abs(d) < 1e-12 {
			return
		}
		minx := int(math.Max(0, math.Floor(math.Min(xs[0], math.Min(xs[1], xs[2])))))
		maxx := int(math.Min(float64(W-1), math.Ceil(math.Max(xs[0], math.Max(xs[1], xs[2])))))
		miny := int(math.Max(0, math.Floor(math.Min(ys[0], math.Min(ys[1], ys[2])))))
		maxy := int(math.Min(float64(H-1), math.Ceil(math.Max(ys[0], math.Max(ys[1], ys[2])))))
		for y := miny; y <= maxy; y++ {
			for x := minx; x <= maxx; x++ {
				px, py := float64(x)+0.5, float64(y)+0.5
				w0 := ((ys[1]-ys[2])*(px-xs[2]) + (xs[2]-xs[1])*(py-ys[2])) / d
				w1 := ((ys[2]-ys[0])*(px-xs[2]) + (xs[0]-xs[2])*(py-ys[2])) / d
				w2 := 1 - w0 - w1
				if w0 < 0 || w1 < 0 || w2 < 0 {
					continue
				}
				iw := w0*t.v[0].iw + w1*t.v[1].iw + w2*t.v[2].iw
				if iw <= 0 {
					continue
				}
				z := (w0*t.v[0].z + w1*t.v[1].z + w2*t.v[2].z)
				pi := y*W + x
				if z >= zbuf[pi] {
					continue
				}
				cr, cg, cb, ca := 1.0, 1.0, 1.0, 1.0
				if t.tex != nil {
					u := (w0*t.v[0].u + w1*t.v[1].u + w2*t.v[2].u) / iw
					v := (w0*t.v[0].v + w1*t.v[1].v + w2*t.v[2].v) / iw
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
					c := t.tex.NRGBAAt(b.Min.X+tx, b.Min.Y+ty)
					cr, cg, cb, ca = float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, float64(c.A)/255
				}
				vr := (w0*t.v[0].r + w1*t.v[1].r + w2*t.v[2].r) / iw
				vg := (w0*t.v[0].g + w1*t.v[1].g + w2*t.v[2].g) / iw
				vb := (w0*t.v[0].b + w1*t.v[1].b + w2*t.v[2].b) / iw
				va := (w0*t.v[0].a + w1*t.v[1].a + w2*t.v[2].a) / iw
				cr, cg, cb = cr*vr, cg*vg, cb*vb
				ca *= va
				if t.mask && ca < 0.5 {
					continue
				}
				o := pi * 4
				if t.additive {
					img.Pix[o] = add255(img.Pix[o], cr*ca)
					img.Pix[o+1] = add255(img.Pix[o+1], cg*ca)
					img.Pix[o+2] = add255(img.Pix[o+2], cb*ca)
				} else if t.blend {
					img.Pix[o] = mix255(img.Pix[o], cr, ca)
					img.Pix[o+1] = mix255(img.Pix[o+1], cg, ca)
					img.Pix[o+2] = mix255(img.Pix[o+2], cb, ca)
				} else {
					img.Pix[o] = clamp255(cr)
					img.Pix[o+1] = clamp255(cg)
					img.Pix[o+2] = clamp255(cb)
				}
				img.Pix[o+3] = 255
				if zwrite {
					zbuf[pi] = z
				}
			}
		}
	}
	for i := range opaque {
		raster(&opaque[i], true)
	}
	sort.Slice(blended, func(i, j int) bool { return blended[i].zc > blended[j].zc })
	for i := range blended {
		raster(&blended[i], false)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	f.Close()
	fmt.Printf("stageverify: %d opaque + %d blended tris -> %s\n", len(opaque), len(blended), *out)
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
func add255(d uint8, v float64) uint8 {
	x := int(float64(d) + v*255)
	if x > 255 {
		x = 255
	}
	return uint8(x)
}
func mix255(d uint8, v, a float64) uint8 {
	x := int(v*a*255 + float64(d)*(1-a))
	if x > 255 {
		x = 255
	}
	return uint8(x)
}
