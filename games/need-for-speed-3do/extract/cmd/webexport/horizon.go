package main

// horizon.go exports a course's horizon backdrop as a camera-centred sky
// dome. The packet's root children 3 and 4 are two 6-cel panorama rings,
// stored rotated 90° (cel columns are screen rows — the engine draws them
// with a swapped corner matrix): child 4 is the opaque FAR ring (sky
// gradient + distant terrain), child 3 the NEAR ring (treeline/skyline with
// a PDEC-transparent sky) drawn over it. Child 3 carries a seventh cel — an
// unrotated 128×64 pre-composed preview of the stacked layers (different CCB
// flags, 0x47EE4420), which is what pinned the layer order. Coastal courses
// use 128-wide near cels and collapse the far ring to a single 4×4 solid.
//
// The dome is a pair of concentric cylinders closed by cone caps in the
// rings' edge colours, with the wrap count and band placement traced from
// the game's own horizon draw (see skyWraps and the band comment below).
// The viewer pins the dome to the camera every frame (see
// course-renderer.js), the same camera-centred contract as the DS viewers'
// skyboxes.

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/threedo"
)

// rotateCCW turns a stored horizon cel upright (columns → rows): the result
// pixel (x, y) is the source pixel (w-1-y, x).
func rotateCCW(src *image.RGBA) *image.RGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < w; y++ {
		for x := 0; x < h; x++ {
			dst.SetRGBA(x, y, src.RGBAAt(w-1-y, x))
		}
	}
	return dst
}

// stitchRing decodes a ring node's six cels, uprights them and joins them
// into one panorama strip.
func stitchRing(a *assets, node *threedo.WrapNode) (*image.RGBA, error) {
	var cels []*image.RGBA
	w, h := 0, 0
	for k := 0; k < 6; k++ {
		img, err := a.tex.celImage(child(node, k))
		if err != nil {
			return nil, fmt.Errorf("ring cel %d: %v", k, err)
		}
		r := rotateCCW(img)
		cels = append(cels, r)
		w += r.Bounds().Dx()
		if r.Bounds().Dy() > h {
			h = r.Bounds().Dy()
		}
	}
	pano := image.NewRGBA(image.Rect(0, 0, w, h))
	x := 0
	for _, c := range cels {
		cw, ch := c.Bounds().Dx(), c.Bounds().Dy()
		for py := 0; py < ch; py++ {
			for px := 0; px < cw; px++ {
				pano.SetRGBA(x+px, py, c.RGBAAt(px, py))
			}
		}
		x += cw
	}
	return pano, nil
}

// edgeColor averages a panorama row's opaque pixels into a 1×1 fill texture.
func edgeColor(pano *image.RGBA, row int) *image.RGBA {
	var r, g, b, n uint32
	for x := 0; x < pano.Bounds().Dx(); x++ {
		c := pano.RGBAAt(x, row)
		if c.A == 0 {
			continue
		}
		r += uint32(c.R)
		g += uint32(c.G)
		b += uint32(c.B)
		n++
	}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if n > 0 {
		img.Pix[0], img.Pix[1], img.Pix[2], img.Pix[3] =
			uint8(r/n), uint8(g/n), uint8(b/n), 0xFF
	}
	return img
}

// skyBuilder accumulates the dome's cylinders and caps.
type skyBuilder struct {
	positions [][3]float32
	uvs       [][2]float32
	groups    []glb.TexturedGroup
}

// skyWraps is how many times a ring's panorama repeats per revolution. The
// game draws each horizon cel over exactly 15° of bearing (the CCB corner
// table the horizon builder feeds MapCel puts the cel boundaries at
// tan(k·15°)×223 px: -61, 31, 100, 160, 219, 288 in the traced frame — a
// 71° hFOV), so 24 cel slots cycle the 6-cel ring exactly 4× per circle.
const skyWraps = 4

// cylinder adds an open cylinder of radius r spanning y0..y1, the panorama
// repeated skyWraps times around it (u grows with bearing). v0/v1 are the
// texture rows at the top and bottom edge; a v beyond [0,1] samples the
// clamped edge row (WrapT CLAMP), which extends the band's own edge colours
// per column instead of a flat fill.
func (b *skyBuilder) cylinder(img image.Image, r, y0, y1, v0, v1 float32) {
	const n = 96
	base := uint32(len(b.positions))
	for s := 0; s <= n; s++ {
		th := 2 * math.Pi * float64(s) / n
		x, z := float32(math.Sin(th)), float32(math.Cos(th))
		u := skyWraps * float32(s) / n
		b.positions = append(b.positions,
			[3]float32{r * x, y1, -r * z}, [3]float32{r * x, y0, -r * z})
		b.uvs = append(b.uvs, [2]float32{u, v0}, [2]float32{u, v1})
	}
	var tris [][3]uint32
	for s := uint32(0); s < n; s++ {
		i := base + 2*s
		tris = append(tris, [3]uint32{i, i + 2, i + 3}, [3]uint32{i, i + 3, i + 1})
	}
	b.groups = append(b.groups, glb.TexturedGroup{Tris: tris, Image: img, WrapS: 10497})
}

// cap adds a cone fan from the circle (r, y) to the apex (0, apexY, 0).
func (b *skyBuilder) cap(img image.Image, r, y, apexY float32) {
	const n = 48
	base := uint32(len(b.positions))
	b.positions = append(b.positions, [3]float32{0, apexY, 0})
	b.uvs = append(b.uvs, [2]float32{0.5, 0.5})
	for s := 0; s <= n; s++ {
		th := 2 * math.Pi * float64(s) / n
		b.positions = append(b.positions,
			[3]float32{r * float32(math.Sin(th)), y, -r * float32(math.Cos(th))})
		b.uvs = append(b.uvs, [2]float32{0.5, 0.5})
	}
	var tris [][3]uint32
	for s := uint32(0); s < n; s++ {
		tris = append(tris, [3]uint32{base, base + 1 + s, base + 2 + s})
	}
	b.groups = append(b.groups, glb.TexturedGroup{Tris: tris, Image: img})
}

// exportSky writes models/sky-<course>.glb: far ring, sky cap, ground cap,
// then the near ring last so it paints over the far one (the viewer draws
// the whole dome first with depth writes off, in primitive order).
func exportSky(a *assets, out string) (string, error) {
	root := a.root
	if len(root.Children) < 5 {
		return "", fmt.Errorf("packet has no horizon children")
	}
	near, err := stitchRing(a, root.Children[3])
	if err != nil {
		return "", fmt.Errorf("near ring: %v", err)
	}
	far, err := stitchRing(a, root.Children[4])
	if err != nil {
		return "", fmt.Errorf("far ring: %v", err)
	}
	if far.Bounds().Dy() <= 8 {
		// Coastal collapses the far ring to a 4×4 destination-shading cel
		// (PIXC 0x9F00 — it dims the framebuffer, not a picture), which
		// bakes to translucent black. Continue the sky from the near
		// ring's own top row instead.
		far = edgeColor(near, 0)
	}

	// Band placement is the traced frame's, in tan units (f = 223 px). The
	// horizon CCBs' HDY is NEGATIVE — source columns walk UP the screen —
	// so each band RISES from its YPos line: the near ring from y 85 up to
	// 35, the far ring from ~68 up to ~0 (start-line frame; same anchors
	// mid-drive). Eye level is y ≈ 85: the projection converges there (the
	// distant roadside-object mips cluster at y 79-84 dead ahead), so the
	// game anchors the treeline base exactly ON the horizon line, with the
	// far ring's hills from +0.08 up — showing over the treeline, only the
	// sky gradient higher, the stacking the pre-composed preview cel
	// depicts. Taller near rings (Coastal's 128-row ocean) keep the same
	// base and grow upward.
	nearRows := near.Bounds().Dy()
	nearY0 := float32(0.0)
	nearTop := nearY0 + 0.22*float32(nearRows)/64
	farY0, farTop := float32(0.08), float32(0.39)
	// The near cylinder runs a little past its traced base with v clamped
	// past 1: each column continues its own bottom-row colour (treeline
	// base, lake shore, ocean) below the horizon line, where the game's
	// terrain otherwise covers from the low driving camera.
	const ext = -0.08
	vExt := 1 + (nearY0-ext)/(nearTop-nearY0)
	b := &skyBuilder{}
	b.cylinder(far, 1.0, farY0, farTop, 0, 1)
	b.cap(edgeColor(far, 0), 1.0, farTop, 1.2)                   // sky above
	b.cap(edgeColor(near, near.Bounds().Dy()-1), 1.0, ext, -0.5) // ground below
	b.cylinder(near, 0.98, ext, nearTop, 0, vExt)

	file := fmt.Sprintf("models/sky-%s.glb", a.course)
	if err := glb.WriteTextured(filepath.Join(out, file), b.positions, b.uvs, b.groups, nil); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[sky] near %dx%d far %dx%d -> %s\n",
		near.Bounds().Dx(), near.Bounds().Dy(), far.Bounds().Dx(), far.Bounds().Dy(), file)
	return file, nil
}
