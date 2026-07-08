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

	"retroreverse.com/games/ultima-underworld-pc/extract/lev"
	"retroreverse.com/games/ultima-underworld-pc/extract/levgeo"
	"retroreverse.com/games/ultima-underworld-pc/extract/tex"
)

const (
	W, H = 900, 700
)

func main() {
	game := flag.String("game", "../game", "path to game/ folder")
	level := flag.Int("level", 0, "level index")
	palN := flag.Int("pal", 0, "palette index")
	out := flag.String("o", "level.png", "output PNG")
	top := flag.Bool("top", false, "top-down camera (verify walls/layout)")
	zoom := flag.Float64("zoom", 0, "half-extent to frame in top view (0=whole level)")
	cxf := flag.Float64("cx", -1, "top-view centre X (-1=level centre)")
	cyf := flag.Float64("cy", -1, "top-view centre Y (-1=level centre)")
	ceil := flag.Bool("ceilings", false, "include ceiling faces")
	pitch := flag.Float64("pitch", 2.0, "top-view camera height multiplier (lower=more side-on)")
	textured := flag.Bool("textured", false, "sample real textures (verify UVs) instead of flat shading")
	uvt := flag.Bool("uvtest", false, "colour walls by UV (R=U, G=V) to read texture orientation")
	fp := flag.Bool("fp", false, "first-person camera at -cx,-cy,-fpz looking along -yaw")
	fpz := flag.Float64("fpz", 3.0, "first-person eye height")
	yaw := flag.Float64("yaw", 0, "first-person look yaw in degrees (0=+X east)")
	ceiltex := flag.Int("ceiltex", -1, "override the ceiling texture (global F32 index; -1 = default)")
	flag.Parse()
	uvTest = *uvt
	if *ceiltex >= 0 {
		levgeo.CeilingTex = uint16(*ceiltex)
	}
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
	mesh := levgeo.Build(grid, tm, *ceil)

	imgCache := map[[2]int]*image.RGBA{}
	texImg := func(wall bool, num uint16) *image.RGBA {
		key := [2]int{b2i(wall), int(num)}
		if im, ok := imgCache[key]; ok {
			return im
		}
		t := floorTR
		if wall {
			t = wallTR
		}
		im, err := t.Image(int(num)%t.Count(), pal)
		if err != nil {
			im = image.NewRGBA(image.Rect(0, 0, 1, 1))
			im.Pix = []byte{128, 128, 128, 255}
		}
		imgCache[key] = im
		return im
	}
	avg := func(wall bool, num uint16) color.RGBA {
		im := texImg(wall, num)
		var r, g, b, n uint64
		for i := 0; i < len(im.Pix); i += 4 {
			r += uint64(im.Pix[i])
			g += uint64(im.Pix[i+1])
			b += uint64(im.Pix[i+2])
			n++
		}
		return color.RGBA{byte(r / n), byte(g / n), byte(b / n), 255}
	}

	// Camera: an angled bird's-eye over the level centre, or straight-down.
	cx, cy := float64(grid.W)/2, float64(grid.H)/2
	if *cxf >= 0 {
		cx = *cxf
	}
	if *cyf >= 0 {
		cy = *cyf
	}
	eye := vec3{cx - 18, cy - 34, 20}
	target := vec3{cx, cy, 0}
	up := vec3{0, 0, 1}
	if *top {
		ext := float64(grid.W) / 2
		if *zoom > 0 {
			ext = *zoom
		}
		// Oblique from the south so wall faces stay visible; -pitch lowers it.
		eye = vec3{cx, cy - ext*1.3, ext * (*pitch)}
		target = vec3{cx, cy, 2}
		up = vec3{0, 0, 1}
	}
	if *fp {
		a := *yaw * math.Pi / 180
		eye = vec3{cx, cy, *fpz}
		target = vec3{cx + math.Cos(a), cy + math.Sin(a), *fpz}
		up = vec3{0, 0, 1}
	}
	view := lookAt(eye, target, up)
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
		var texi *image.RGBA
		if *textured || *uvt {
			texi = texImg(q.Wall, q.Tex)
		}
		for _, t := range tri {
			var sp [3]vert
			ok := true
			for k, ci := range t {
				cl := mul(proj, mul(view, vec4{float64(q.P[ci][0]), float64(q.P[ci][1]), float64(q.P[ci][2]), 1}))
				if cl.w <= 0.01 {
					ok = false
					break
				}
				inv := 1 / cl.w
				sp[k] = vert{
					vec3{(cl.x*inv*0.5 + 0.5) * W, (1 - (cl.y*inv*0.5 + 0.5)) * H, cl.w},
					float64(q.UV[ci][0]) * inv, float64(q.UV[ci][1]) * inv, inv,
				}
			}
			if ok {
				fillTri(img, zbuf, sp, col, texi, sh)
			}
		}
	}
	f, _ := os.Create(*out)
	defer f.Close()
	png.Encode(f, img)
	fmt.Printf("wrote %s: level %d, %d quads\n", *out, *level, len(mesh.Quads))
}

// vert is a rasteriser vertex: screen pos (z=clip w for the z-buffer) plus
// perspective-correct u/w, v/w and 1/w.
type vert struct {
	p          vec3
	uw, vw, iw float64
}

func fillTri(img *image.RGBA, zb []float64, p [3]vert, c color.RGBA, texi *image.RGBA, sh float64) {
	minx := int(math.Max(0, math.Floor(math.Min(p[0].p.x, math.Min(p[1].p.x, p[2].p.x)))))
	maxx := int(math.Min(W-1, math.Ceil(math.Max(p[0].p.x, math.Max(p[1].p.x, p[2].p.x)))))
	miny := int(math.Max(0, math.Floor(math.Min(p[0].p.y, math.Min(p[1].p.y, p[2].p.y)))))
	maxy := int(math.Min(H-1, math.Ceil(math.Max(p[0].p.y, math.Max(p[1].p.y, p[2].p.y)))))
	area := edge(p[0].p, p[1].p, p[2].p)
	if math.Abs(area) < 1e-6 {
		return
	}
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			pt := vec3{float64(x) + 0.5, float64(y) + 0.5, 0}
			w0 := edge(p[1].p, p[2].p, pt)
			w1 := edge(p[2].p, p[0].p, pt)
			w2 := edge(p[0].p, p[1].p, pt)
			if (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0) {
				b0, b1, b2 := w0/area, w1/area, w2/area
				z := b0*p[0].p.z + b1*p[1].p.z + b2*p[2].p.z
				idx := y*W + x
				if z >= zb[idx] {
					continue
				}
				zb[idx] = z
				col := c
				if texi != nil {
					iw := b0*p[0].iw + b1*p[1].iw + b2*p[2].iw
					u := (b0*p[0].uw + b1*p[1].uw + b2*p[2].uw) / iw
					v := (b0*p[0].vw + b1*p[1].vw + b2*p[2].vw) / iw
					if uvTest {
						// R=U, G=V within one tile (fractional). Reads: red at U=1,
						// green at V=1 (top). Lets us see mirrors/upside-down directly.
						col = color.RGBA{clamp((u - math.Floor(u)) * 255), clamp((v - math.Floor(v)) * 255), 40, 255}
					} else {
						col = sampleTex(texi, u, v, sh)
					}
				}
				o := idx * 4
				img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
			}
		}
	}
}

// sampleTex nearest-samples a texture at (u,v) with V flipped to match three.js
// (flipY): v=0 is the bottom row. Tiles via fractional wrap.
func sampleTex(im *image.RGBA, u, v, sh float64) color.RGBA {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	fu := u - math.Floor(u)
	fv := v - math.Floor(v)
	tx := int(fu * float64(w))
	ty := int((1 - fv) * float64(h)) // flipY: v=0 -> bottom row
	if tx < 0 {
		tx = 0
	} else if tx >= w {
		tx = w - 1
	}
	if ty < 0 {
		ty = 0
	} else if ty >= h {
		ty = h - 1
	}
	o := (ty*w + tx) * 4
	return color.RGBA{clamp(float64(im.Pix[o]) * sh), clamp(float64(im.Pix[o+1]) * sh), clamp(float64(im.Pix[o+2]) * sh), 255}
}

var uvTest bool

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
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
