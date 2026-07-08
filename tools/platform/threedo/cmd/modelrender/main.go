// modelrender decodes 3DO ORI3 car models (from Need for Speed's .WrapFam / .3OR
// files) and renders orthographic wireframe views to a PNG — side, top and front.
//
//	modelrender -image disc.bin -path DriveData/DriveArt/CarArt/VIPER.WrapFam -o viper.png
//	modelrender -f VIPER.WrapFam -o viper.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	imagePath := flag.String("image", "", "3DO disc image")
	path := flag.String("path", "", "model file path within the disc")
	file := flag.String("f", "", "standalone .WrapFam/.3OR file")
	out := flag.String("o", "model.png", "output PNG")
	flag.Parse()

	var data []byte
	var err error
	switch {
	case *file != "":
		data, err = os.ReadFile(*file)
	case *imagePath != "" && *path != "":
		var img []byte
		if img, err = os.ReadFile(*imagePath); err == nil {
			var vol *threedo.Volume
			if vol, err = threedo.Open(img); err == nil {
				data, err = vol.ReadFile(*path)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: modelrender -image DISC -path P -o out.png | -f FILE -o out.png")
		os.Exit(2)
	}
	if err != nil {
		die(err)
	}

	models, err := threedo.ParseModels(data)
	if err != nil {
		die(err)
	}
	for _, m := range models {
		mn, mx := m.Bounds()
		fmt.Printf("model %-10q  %d verts  %d quads   bounds X[%d,%d] Y[%d,%d] Z[%d,%d]\n",
			m.Name, len(m.Verts), len(m.Faces), mn.X, mx.X, mn.Y, mx.Y, mn.Z, mx.Z)
	}
	// Render the highest-detail model (a WrapFam may hold several LODs).
	best := models[0]
	for _, m := range models[1:] {
		if len(m.Faces) > len(best.Faces) {
			best = m
		}
	}
	img := renderViews(best)
	f, err := os.Create(*out)
	if err != nil {
		die(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		die(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
}

// renderViews draws side / top / front wireframes of a model into one image.
func renderViews(m *threedo.Model) *image.RGBA {
	const W, H = 560, 420
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for i := range img.Pix { // opaque black background
		if i%4 == 3 {
			img.Pix[i] = 0xFF
		}
	}
	mn, mx := m.Bounds()
	span := func(a, b int32) float64 {
		if b-a == 0 {
			return 1
		}
		return float64(b - a)
	}
	// Scale so the longest axis (car length, Z) fits ~40% of the width.
	scale := 0.40 * float64(W) / span(mn.Z, mx.Z)

	line := func(x0, y0, x1, y1 int, c color.RGBA) {
		dx := abs(x1 - x0)
		dy := -abs(y1 - y0)
		sx, sy := 1, 1
		if x0 >= x1 {
			sx = -1
		}
		if y0 >= y1 {
			sy = -1
		}
		err := dx + dy
		for {
			if x0 >= 0 && x0 < W && y0 >= 0 && y0 < H {
				img.SetRGBA(x0, y0, c)
			}
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * err
			if e2 >= dy {
				err += dy
				x0 += sx
			}
			if e2 <= dx {
				err += dx
				y0 += sy
			}
		}
	}
	// view: pick two axes to project (0=side ZY, 1=top ZX, 2=front XY).
	draw := func(view int, ox, oy float64, c color.RGBA) {
		proj := func(v threedo.Vec3) (float64, float64) {
			switch view {
			case 0:
				return float64(v.Z), -float64(v.Y) // side
			case 1:
				return float64(v.Z), -float64(v.X) // top
			default:
				return float64(v.X), -float64(v.Y) // front
			}
		}
		for _, f := range m.Faces {
			for k := 0; k < 4; k++ {
				a, b := m.Verts[f.V[k]], m.Verts[f.V[(k+1)%4]]
				ax, ay := proj(a)
				bx, by := proj(b)
				line(int(ox+ax*scale), int(oy+ay*scale), int(ox+bx*scale), int(oy+by*scale), c)
			}
		}
	}
	blue := color.RGBA{0x78, 0xC8, 0xFF, 0xFF}
	orange := color.RGBA{0xFF, 0xC8, 0x78, 0xFF}
	green := color.RGBA{0x9E, 0xE6, 0x9E, 0xFF}
	draw(0, float64(W)*0.28, float64(H)*0.28, blue)   // side, upper-left
	draw(1, float64(W)*0.28, float64(H)*0.72, orange) // top, lower-left
	draw(2, float64(W)*0.80, float64(H)*0.28, green)  // front, upper-right
	return img
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "modelrender:", err)
	os.Exit(1)
}
