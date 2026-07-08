// trackmap draws a course's full track layout: the course model's textured
// geometry seen top-down, overlaid with the NKM course-map data — the checkpoint
// gates (key checkpoints and the lap line emphasised), the CPU racers' drive line
// (EPOI), the item-probe line (IPOI), Lakitu respawn points, object placements and
// the grid start. The overlay coordinates come straight from the NKM; that they
// land on the road is the verification that both decoders (model and map) agree.
//
//	trackmap [-size N] [-o DIR] course.carc
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

	"retroreverse.com/games/mario-kart-ds/extract/mkds"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
)

func main() {
	size := flag.Int("size", 1024, "output image size (px)")
	outDir := flag.String("o", "../rendered/tracks", "output directory")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: trackmap [-size N] [-o DIR] course.carc")
		os.Exit(2)
	}
	path := flag.Arg(0)

	models, err := mkds.LoadModels(path)
	if err != nil {
		die(err)
	}
	texs := mkds.LoadTextures(path)

	// The main course model is the one named like the archive (largest, in practice
	// the first "<x>_course"-named model; fall back to the biggest).
	var course *nitro.Model
	for i := range models {
		if strings.HasSuffix(models[i].Name, "_course") || strings.HasSuffix(models[i].Name, "_stage") {
			course = &models[i]
			break
		}
	}
	if course == nil && len(models) > 0 {
		course = &models[0]
	}
	if course == nil {
		die(fmt.Errorf("no course model in %s", path))
	}

	// First NKMD in the archive = the single-player course map.
	raw, err := os.ReadFile(path)
	if err != nil {
		die(err)
	}
	files, err := nds.ParseNARC(raw)
	if err != nil {
		die(err)
	}
	var nkm *mkds.NKM
	for _, f := range files {
		if len(f) >= 4 && string(f[:4]) == "NKMD" {
			if nkm, err = mkds.ParseNKM(f); err != nil {
				die(err)
			}
			break
		}
	}
	if nkm == nil {
		die(fmt.Errorf("no NKMD in %s", path))
	}

	img := render(*course, texs, nkm, *size)
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	out := filepath.Join(*outDir, stem+"-layout.png")
	if err := writePNG(out, img); err != nil {
		die(err)
	}
	fmt.Printf("  %-24s %3d checkpoints, %2d enemy-line sections (%d pts), %2d objects → %s\n",
		course.Name, len(nkm.Checkpoints), len(nkm.EnemyPaths), len(nkm.EnemyPoints), len(nkm.Objects), filepath.Base(out))
}

func render(m nitro.Model, texs map[string]nitro.Texture, nkm *mkds.NKM, size int) *image.NRGBA {
	var tris []nitro.Tri
	for _, d := range nitro.RunSBC(m) {
		if d.Shape < len(m.Shapes) {
			tris = append(tris, nitro.DecodeDL(m.Shapes[d.Shape].DL, d.Stack, d.M, d.Mat)...)
		}
	}
	// The render model lives at 1/16 of world scale: NKM (and collision) coordinates
	// are kart-world units, and the engine scales course geometry up by 16 when
	// drawing. Scale the triangles so the two sources share one frame.
	for i := range tris {
		for j := range tris[i].V {
			tris[i].V[j].X *= 16
			tris[i].V[j].Y *= 16
			tris[i].V[j].Z *= 16
		}
	}
	// Bounds over geometry AND the map data, top-down (x → screen x, z → screen y).
	minX, minZ := math.Inf(1), math.Inf(1)
	maxX, maxZ := math.Inf(-1), math.Inf(-1)
	ext := func(x, z float64) {
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minZ, maxZ = math.Min(minZ, z), math.Max(maxZ, z)
	}
	for _, t := range tris {
		for _, v := range t.V {
			ext(v.X, v.Z)
		}
	}
	for _, c := range nkm.Checkpoints {
		ext(c.X1, c.Z1)
		ext(c.X2, c.Z2)
	}
	span := math.Max(maxX-minX, maxZ-minZ) * 1.06
	scale := float64(size) / span
	ox := (float64(size) - (maxX-minX)*scale) / 2
	oz := (float64(size) - (maxZ-minZ)*scale) / 2
	px := func(x, z float64) (float64, float64) {
		return ox + (x-minX)*scale, oz + (z-minZ)*scale
	}

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	zbuf := make([]float64, size*size)
	for i := range zbuf {
		zbuf[i] = math.Inf(-1)
	}
	// Geometry pass: top-down, textured, dimmed so the overlay reads.
	for _, t := range tris {
		var xs, ys, hs [3]float64
		for i, v := range t.V {
			xs[i], ys[i] = px(v.X, v.Z)
			hs[i] = v.Y
		}
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
		raster(img, zbuf, size, xs, ys, hs, t, tex, repS, repT, flipS, flipT, us, vsc)
	}

	// Overlay: item line (yellow), enemy line (cyan), checkpoints (red; key gold),
	// respawns (magenta), objects (white), start (green).
	drawSections(img, nkm.ItemPaths, func(i int) (float64, float64) {
		p := nkm.ItemPoints[i]
		return px(p.Pos.X, p.Pos.Z)
	}, len(nkm.ItemPoints), color.NRGBA{240, 200, 40, 255}, 1)
	drawSections(img, nkm.EnemyPaths, func(i int) (float64, float64) {
		p := nkm.EnemyPoints[i]
		return px(p.Pos.X, p.Pos.Z)
	}, len(nkm.EnemyPoints), color.NRGBA{40, 220, 240, 255}, 2)
	for _, c := range nkm.Checkpoints {
		x1, y1 := px(c.X1, c.Z1)
		x2, y2 := px(c.X2, c.Z2)
		col := color.NRGBA{240, 60, 60, 255}
		w := 1
		if c.KeyID == 0 {
			col = color.NRGBA{255, 255, 255, 255} // the lap line
			w = 3
		} else if c.KeyID > 0 {
			col = color.NRGBA{255, 180, 40, 255} // key checkpoint
			w = 2
		}
		line(img, x1, y1, x2, y2, col, w)
	}
	for _, r := range nkm.Respawns {
		x, y := px(r.Pos.X, r.Pos.Z)
		disc(img, x, y, 4, color.NRGBA{230, 80, 230, 255})
	}
	for _, o := range nkm.Objects {
		x, y := px(o.Pos.X, o.Pos.Z)
		disc(img, x, y, 2.5, color.NRGBA{255, 255, 255, 220})
	}
	for _, s := range nkm.Starts {
		x, y := px(s.Pos.X, s.Pos.Z)
		disc(img, x, y, 6, color.NRGBA{60, 230, 90, 255})
	}
	return img
}

// drawSections draws poly-lines section by section, following Next links so
// branches (alternate routes) are all drawn.
func drawSections(img *image.NRGBA, paths []mkds.SectPath, at func(i int) (float64, float64), npts int, col color.NRGBA, w int) {
	for _, sp := range paths {
		for i := sp.Start; i < sp.Start+sp.Len-1 && i+1 < npts; i++ {
			x1, y1 := at(i)
			x2, y2 := at(i + 1)
			line(img, x1, y1, x2, y2, col, w)
		}
		// connect the section's last point to each successor section's first
		if sp.Len > 0 {
			lx, ly := at(sp.Start + sp.Len - 1)
			for _, nx := range sp.Next {
				if nx >= 0 && nx < len(paths) && paths[nx].Len > 0 && paths[nx].Start < npts {
					tx, ty := at(paths[nx].Start)
					line(img, lx, ly, tx, ty, col, w)
				}
			}
		}
	}
}

func raster(img *image.NRGBA, zbuf []float64, size int, xs, ys, hs [3]float64, t nitro.Tri, tex *nitro.Texture, repS, repT, flipS, flipT bool, us, vsc float64) {
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
			pxf, pyf := float64(x)+0.5, float64(y)+0.5
			w0 := ((ys[1]-ys[2])*(pxf-xs[2]) + (xs[2]-xs[1])*(pyf-ys[2])) / d
			w1 := ((ys[2]-ys[0])*(pxf-xs[2]) + (xs[0]-xs[2])*(pyf-ys[2])) / d
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			h := w0*hs[0] + w1*hs[1] + w2*hs[2] // height as depth: highest wins
			if h <= zbuf[y*size+x] {
				continue
			}
			c := color.NRGBA{
				R: t.V[0].C.R, G: t.V[0].C.G, B: t.V[0].C.B, A: 0xFF,
			}
			if tex != nil {
				u := (w0*t.V[0].U + w1*t.V[1].U + w2*t.V[2].U) * us
				v := (w0*t.V[0].V + w1*t.V[1].V + w2*t.V[2].V) * vsc
				tc := sample(tex, u, v, repS, repT, flipS, flipT)
				if tc.A == 0 {
					continue
				}
				c = tc
			}
			// dim so the overlay stands out
			c.R = uint8(float64(c.R) * 0.55)
			c.G = uint8(float64(c.G) * 0.55)
			c.B = uint8(float64(c.B) * 0.55)
			img.SetNRGBA(x, y, c)
			zbuf[y*size+x] = h
		}
	}
}

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
	default:
		if c < 0 {
			return 0
		}
		if c >= n {
			return n - 1
		}
		return c
	}
}

func line(img *image.NRGBA, x1, y1, x2, y2 float64, col color.NRGBA, w int) {
	steps := int(math.Hypot(x2-x1, y2-y1)) + 1
	for s := 0; s <= steps; s++ {
		t := float64(s) / float64(steps)
		x, y := x1+(x2-x1)*t, y1+(y2-y1)*t
		for dy := -w / 2; dy <= w/2; dy++ {
			for dx := -w / 2; dx <= w/2; dx++ {
				setPx(img, int(x)+dx, int(y)+dy, col)
			}
		}
	}
}

func disc(img *image.NRGBA, x, y, r float64, col color.NRGBA) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				setPx(img, int(x+dx), int(y+dy), col)
			}
		}
	}
}

func setPx(img *image.NRGBA, x, y int, col color.NRGBA) {
	if x >= 0 && y >= 0 && x < img.Bounds().Dx() && y < img.Bounds().Dy() {
		img.SetNRGBA(x, y, col)
	}
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
	fmt.Fprintln(os.Stderr, "trackmap:", err)
	os.Exit(1)
}
