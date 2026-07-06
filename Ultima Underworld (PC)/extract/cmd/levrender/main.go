// Command levrender is a verification tool: it software-renders a level's mesh
// (from levgeo) to a PNG with a z-buffer and simple lighting, each face flat-
// shaded by its texture's average colour. It confirms the reverse-engineered
// geometry forms a coherent 3D dungeon without needing the browser viewer.
//
// Usage: levrender [-game ../game] [-level 0] -o out.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"ultimaunderworld/extract/lev"
	"ultimaunderworld/extract/levgeo"
	"ultimaunderworld/extract/tex"
)

const (
	W, H = 900, 700
)

func main() {
	game := flag.String("game", "../game", "path to game/ folder")
	level := flag.Int("level", 0, "level index")
	palN := flag.Int("pal", 0, "palette index")
	out := flag.String("o", "level.png", "output PNG")
	flag.Parse()
	rd := func(n ...string) []byte {
		b, err := os.ReadFile(filepath.Join(append([]string{*game, "DATA"}, n...)...))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return b
	}
	ark, _ := lev.ParseArk(rd("LEV.ARK"))
	block, _ := ark.Block(*level)
	grid, _ := lev.DecodeGrid(block)
	tm, _ := ark.TexMapForLevel(*level)
	pal, _ := tex.LoadPalette(rd("PALS.DAT"), *palN)
	wallTR, _ := tex.ParseTR(rd("W64.TR"))
	floorTR, _ := tex.ParseTR(rd("F32.TR"))
	mesh := levgeo.Build(grid, tm, false)

	avg := func(wall bool, num uint16) color.RGBA {
		var t *tex.TR
		if wall {
			t = wallTR
		} else {
			t = floorTR
		}
		im, err := t.Image(int(num)%t.Count(), pal)
		if err != nil {
			return color.RGBA{128, 128, 128, 255}
		}
		var r, g, b, n uint64
		for i := 0; i < len(im.Pix); i += 4 {
			r += uint64(im.Pix[i])
			g += uint64(im.Pix[i+1])
			b += uint64(im.Pix[i+2])
			n++
		}
		return color.RGBA{byte(r / n), byte(g / n), byte(b / n), 255}
	}

	// Camera: an angled bird's-eye over the level centre.
	cx, cy := float64(grid.W)/2, float64(grid.H)/2
	eye := vec3{cx - 18, cy - 34, 20}
	target := vec3{cx, cy, 0}
	view := lookAt(eye, target, vec3{0, 0, 1})
	proj := perspective(50*math.Pi/180, float64(W)/H, 0.1, 400)
	light := norm(vec3{0.4, 0.5, 0.8})

	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for i := range img.Pix {
		img.Pix[i] = []byte{10, 11, 16, 255}[i%4]
	}
	zbuf := make([]float64, W*H)
	for i := range zbuf {
		zbuf[i] = math.Inf(1)
	}

	tri := [][3]int{{0, 1, 2}, {0, 2, 3}}
	for _, q := range mesh.Quads {
		base := avg(q.Wall, q.Tex)
		// Face normal for Lambert shading.
		e1 := sub(v3(q.P[1]), v3(q.P[0]))
		e2 := sub(v3(q.P[2]), v3(q.P[0]))
		nrm := norm(cross(e1, e2))
		sh := math.Abs(dot(nrm, light))*0.75 + 0.35
		col := color.RGBA{clamp(float64(base.R) * sh), clamp(float64(base.G) * sh), clamp(float64(base.B) * sh), 255}
		for _, t := range tri {
			var sp [3]vec3
			ok := true
			for k, ci := range t {
				cl := mul(proj, mul(view, vec4{float64(q.P[ci][0]), float64(q.P[ci][1]), float64(q.P[ci][2]), 1}))
				if cl.w <= 0.01 {
					ok = false
					break
				}
				sp[k] = vec3{(cl.x/cl.w*0.5 + 0.5) * W, (1 - (cl.y/cl.w*0.5 + 0.5)) * H, cl.w}
			}
			if ok {
				fillTri(img, zbuf, sp, col)
			}
		}
	}
	f, _ := os.Create(*out)
	defer f.Close()
	png.Encode(f, img)
	fmt.Printf("wrote %s: level %d, %d quads\n", *out, *level, len(mesh.Quads))
}

func fillTri(img *image.RGBA, zb []float64, p [3]vec3, c color.RGBA) {
	minx := int(math.Max(0, math.Floor(math.Min(p[0].x, math.Min(p[1].x, p[2].x)))))
	maxx := int(math.Min(W-1, math.Ceil(math.Max(p[0].x, math.Max(p[1].x, p[2].x)))))
	miny := int(math.Max(0, math.Floor(math.Min(p[0].y, math.Min(p[1].y, p[2].y)))))
	maxy := int(math.Min(H-1, math.Ceil(math.Max(p[0].y, math.Max(p[1].y, p[2].y)))))
	area := edge(p[0], p[1], p[2])
	if math.Abs(area) < 1e-6 {
		return
	}
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			pt := vec3{float64(x) + 0.5, float64(y) + 0.5, 0}
			w0 := edge(p[1], p[2], pt)
			w1 := edge(p[2], p[0], pt)
			w2 := edge(p[0], p[1], pt)
			if (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0) {
				b0, b1, b2 := w0/area, w1/area, w2/area
				z := b0*p[0].z + b1*p[1].z + b2*p[2].z
				idx := y*W + x
				if z < zb[idx] {
					zb[idx] = z
					o := idx * 4
					img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = c.R, c.G, c.B, 255
				}
			}
		}
	}
}

func edge(a, b, c vec3) float64 { return (c.x-a.x)*(b.y-a.y) - (c.y-a.y)*(b.x-a.x) }
func clamp(v float64) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

// --- tiny linear algebra ---
type vec3 struct{ x, y, z float64 }
type vec4 struct{ x, y, z, w float64 }
type mat4 [16]float64

func v3(p [3]float32) vec3  { return vec3{float64(p[0]), float64(p[1]), float64(p[2])} }
func sub(a, b vec3) vec3    { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func dot(a, b vec3) float64 { return a.x*b.x + a.y*b.y + a.z*b.z }
func cross(a, b vec3) vec3  { return vec3{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x} }
func norm(a vec3) vec3 {
	l := math.Sqrt(dot(a, a))
	if l == 0 {
		return a
	}
	return vec3{a.x / l, a.y / l, a.z / l}
}
func mul(m mat4, v vec4) vec4 {
	return vec4{
		m[0]*v.x + m[4]*v.y + m[8]*v.z + m[12]*v.w,
		m[1]*v.x + m[5]*v.y + m[9]*v.z + m[13]*v.w,
		m[2]*v.x + m[6]*v.y + m[10]*v.z + m[14]*v.w,
		m[3]*v.x + m[7]*v.y + m[11]*v.z + m[15]*v.w,
	}
}
func mulm(a, b mat4) mat4 {
	var r mat4
	for c := 0; c < 4; c++ {
		for row := 0; row < 4; row++ {
			for k := 0; k < 4; k++ {
				r[c*4+row] += a[k*4+row] * b[c*4+k]
			}
		}
	}
	return r
}
func lookAt(eye, at, up vec3) mat4 {
	f := norm(sub(at, eye))
	s := norm(cross(f, up))
	u := cross(s, f)
	return mat4{
		s.x, u.x, -f.x, 0,
		s.y, u.y, -f.y, 0,
		s.z, u.z, -f.z, 0,
		-dot(s, eye), -dot(u, eye), dot(f, eye), 1,
	}
}
func perspective(fovy, aspect, near, far float64) mat4 {
	t := 1 / math.Tan(fovy/2)
	return mat4{
		t / aspect, 0, 0, 0,
		0, t, 0, 0,
		0, 0, (far + near) / (near - far), -1,
		0, 0, 2 * far * near / (near - far), 0,
	}
}

func init() { _ = mulm } // keep mulm available for tweaks
