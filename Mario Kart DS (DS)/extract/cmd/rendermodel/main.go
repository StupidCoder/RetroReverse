// rendermodel software-renders an NSBMD model to PNG: parse the model, run its SBC
// to place the shapes, decode the display lists to triangles, and rasterise with a
// z-buffer — textured from the companion NSBTX via the material bindings (the
// model's own authoritative texture↔palette pairing), modulated by vertex colours
// and a simple face-normal light so the shape reads.
//
//	rendermodel [-size N] [-yaw D] [-pitch D] [-tex FILE] [-o DIR] model.nsbmd
//
// The texture set defaults to the .nsbtx next to the model (same stem, then any
// sibling). Output is <stem>.png, orthographic, auto-fitted.
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
	"strings"

	"mariokartds/extract/mkds"

	"retroreverse.com/tools/nds"
	"retroreverse.com/tools/nds/nitro"
)

func main() {
	size := flag.Int("size", 512, "output image size (px)")
	yaw := flag.Float64("yaw", 40, "camera yaw (degrees)")
	pitch := flag.Float64("pitch", 25, "camera pitch (degrees)")
	texFile := flag.String("tex", "", "NSBTX texture file (default: sibling of the model)")
	outDir := flag.String("o", "../rendered/models", "output directory")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: rendermodel [-size N] [-yaw D] [-pitch D] [-tex F] [-o DIR] model.nsbmd")
		os.Exit(2)
	}
	path := flag.Arg(0)
	models, err := mkds.LoadModels(path)
	if err != nil {
		die(err)
	}

	texs := mkds.LoadTextures(path)
	if *texFile != "" {
		if data, err := os.ReadFile(*texFile); err == nil {
			if ts, err := nitro.DecodeNSBTX(nds.Decompress(data)); err == nil {
				for _, t := range ts {
					texs[t.Name] = t
				}
			}
		}
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	for _, m := range models {
		img, ntris := render(m, texs, *size, *yaw, *pitch)
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		name := stem + ".png"
		if len(models) > 1 {
			name = stem + "-" + m.Name + ".png"
		}
		if err := writePNG(filepath.Join(*outDir, name), img); err != nil {
			die(err)
		}
		fmt.Printf("  %-24s %4d tris  → %s\n", m.Name, ntris, name)
	}
}

func render(m nitro.Model, texs map[string]nitro.Texture, size int, yawDeg, pitchDeg float64) (*image.NRGBA, int) {
	// Gather all triangles.
	var tris []nitro.Tri
	for _, d := range nitro.RunSBC(m) {
		if d.Shape < len(m.Shapes) {
			tris = append(tris, nitro.DecodeDL(m.Shapes[d.Shape].DL, d.Stack, d.M, d.Mat)...)
		}
	}
	// View rotation.
	ya, pa := yawDeg*math.Pi/180, pitchDeg*math.Pi/180
	sy, cy := math.Sin(ya), math.Cos(ya)
	sp, cp := math.Sin(pa), math.Cos(pa)
	rot := func(v nitro.Vertex) (float64, float64, float64) {
		x, z := v.X*cy+v.Z*sy, -v.X*sy+v.Z*cy
		y, z2 := v.Y*cp-z*sp, v.Y*sp+z*cp
		return x, y, z2
	}
	// Bounds → fit.
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, t := range tris {
		for _, v := range t.V {
			x, y, _ := rot(v)
			minX, maxX = math.Min(minX, x), math.Max(maxX, x)
			minY, maxY = math.Min(minY, y), math.Max(maxY, y)
		}
	}
	if len(tris) == 0 || maxX <= minX {
		return image.NewNRGBA(image.Rect(0, 0, size, size)), 0
	}
	span := math.Max(maxX-minX, maxY-minY) * 1.1
	scale := float64(size) / span
	ox := (float64(size) - (maxX-minX)*scale) / 2
	oy := (float64(size) - (maxY-minY)*scale) / 2

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	zbuf := make([]float64, size*size)
	for i := range zbuf {
		zbuf[i] = math.Inf(-1)
	}

	for _, t := range tris {
		var sxs, sys, szs [3]float64
		for i, v := range t.V {
			x, y, z := rot(v)
			sxs[i] = ox + (x-minX)*scale
			sys[i] = float64(size) - (oy + (y-minY)*scale) // +y up
			szs[i] = z
		}
		// Face normal light (screen-space z of the normal).
		ax, ay := sxs[1]-sxs[0], sys[1]-sys[0]
		bx, by := sxs[2]-sxs[0], sys[2]-sys[0]
		area := ax*by - ay*bx
		if area == 0 {
			continue
		}
		shade := 0.72 + 0.28*math.Abs(area)/(math.Hypot(ax, ay)*math.Hypot(bx, by)+1e-9)

		var tex *nitro.Texture
		repS, repT, flipS, flipT := false, false, false, false
		us, vsc := 1.0, 1.0
		if t.Mat < len(m.Materials) {
			mat := m.Materials[t.Mat]
			if tx, ok := texs[mat.Texture]; ok && mat.Texture != "" {
				tex = &tx
			}
			repS = mat.TexParam&(1<<16) != 0
			repT = mat.TexParam&(1<<17) != 0
			flipS = mat.TexParam&(1<<18) != 0
			flipT = mat.TexParam&(1<<19) != 0
			us, vsc = mat.UVScale()
		}
		rasterise(img, zbuf, size, sxs, sys, szs, t, tex, shade, repS, repT, flipS, flipT, us, vsc)
	}
	return img, len(tris)
}

func rasterise(img *image.NRGBA, zbuf []float64, size int, xs, ys, zs [3]float64, t nitro.Tri, tex *nitro.Texture, shade float64, repS, repT, flipS, flipT bool, us, vsc float64) {
	minx := int(math.Max(0, math.Floor(min3(xs))))
	maxx := int(math.Min(float64(size-1), math.Ceil(max3(xs))))
	miny := int(math.Max(0, math.Floor(min3(ys))))
	maxy := int(math.Min(float64(size-1), math.Ceil(max3(ys))))
	d := (ys[1]-ys[2])*(xs[0]-xs[2]) + (xs[2]-xs[1])*(ys[0]-ys[2])
	if d == 0 {
		return
	}
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			w0 := ((ys[1]-ys[2])*(px-xs[2]) + (xs[2]-xs[1])*(py-ys[2])) / d
			w1 := ((ys[2]-ys[0])*(px-xs[2]) + (xs[0]-xs[2])*(py-ys[2])) / d
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			z := w0*zs[0] + w1*zs[1] + w2*zs[2]
			if z <= zbuf[y*size+x] {
				continue
			}
			c := color.NRGBA{
				R: lerp3(t.V[0].C.R, t.V[1].C.R, t.V[2].C.R, w0, w1, w2),
				G: lerp3(t.V[0].C.G, t.V[1].C.G, t.V[2].C.G, w0, w1, w2),
				B: lerp3(t.V[0].C.B, t.V[1].C.B, t.V[2].C.B, w0, w1, w2),
				A: 0xFF,
			}
			if tex != nil {
				u := (w0*t.V[0].U + w1*t.V[1].U + w2*t.V[2].U) * us
				v := (w0*t.V[0].V + w1*t.V[1].V + w2*t.V[2].V) * vsc
				tc := sample(tex, u, v, repS, repT, flipS, flipT)
				if tc.A == 0 {
					continue // transparent texel
				}
				c.R = uint8(int(c.R) * int(tc.R) / 255)
				c.G = uint8(int(c.G) * int(tc.G) / 255)
				c.B = uint8(int(c.B) * int(tc.B) / 255)
			}
			c.R = uint8(math.Min(255, float64(c.R)*shade))
			c.G = uint8(math.Min(255, float64(c.G)*shade))
			c.B = uint8(math.Min(255, float64(c.B)*shade))
			img.SetNRGBA(x, y, c)
			zbuf[y*size+x] = z
		}
	}
}

// sample fetches a texel honouring the GX repeat/flip addressing.
func sample(t *nitro.Texture, u, v float64, repS, repT, flipS, flipT bool) color.NRGBA {
	x := wrap(int(math.Floor(u)), t.Width, repS, flipS)
	y := wrap(int(math.Floor(v)), t.Height, repT, flipT)
	return t.Img.NRGBAAt(x, y)
}

func wrap(c, n int, rep, flip bool) int {
	if n <= 0 {
		return 0
	}
	switch {
	case flip:
		m := c & (2*n - 1)
		if m < 0 {
			m += 2 * n
		}
		if m >= n {
			m = 2*n - 1 - m
		}
		return m
	case rep:
		m := c % n
		if m < 0 {
			m += n
		}
		return m
	default: // clamp
		if c < 0 {
			return 0
		}
		if c >= n {
			return n - 1
		}
		return c
	}
}

func lerp3(a, b, c uint8, w0, w1, w2 float64) uint8 {
	return uint8(math.Min(255, w0*float64(a)+w1*float64(b)+w2*float64(c)))
}

func min3(v [3]float64) float64 { return math.Min(v[0], math.Min(v[1], v[2])) }
func max3(v [3]float64) float64 { return math.Max(v[0], math.Max(v[1], v[2])) }

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "rendermodel:", err)
	os.Exit(1)
}
