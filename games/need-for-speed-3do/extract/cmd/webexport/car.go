package main

// car.go exports one car .wrapFam as a textured GLB: the highest-detail
// (model, shape) LOD, each quad face textured by its SPoT via the "!ori"
// face map. ORI3 vertices are 16.16 world units, so the same 1 unit = 1 m
// scale as the course applies.
//
// Two departures from a naive quad dump, both forced by the target being a
// depth-tested triangle rasterizer instead of the 3DO's order-painted cel
// engine:
//
//   - The cel engine's corner matrix maps a texture onto a quad BILINEARLY
//     (X(c,r) = XPos + r·VDX + c·(HDX + r·HDDX) is linear in c and r
//     together, never affine per half). Splitting a non-parallelogram quad
//     into two affinely-textured triangles bends every straight texture line
//     at the diagonal — the Rodeo's window pillars kink where the cut runs.
//     Faces are therefore subdivided into an N×N bilinear grid until the
//     worst surface deviation is under ~2 model units, which converges on
//     the cel engine's mapping.
//
//   - The game draws a car's faces in model order with no depth buffer, and
//     the faces flagged 0x04 in the material word (wheels, and the dark
//     axle strips behind them) come last, painting over the body. Their
//     quads sit INBOARD of the body panels (the Diablo's rims at |x|=127/112
//     against sills reaching |x|=129), so a depth test hands the win to the
//     body's black arch backdrop instead. The outermost side-facing overlay
//     faces are pushed just outside whatever body surface covers their
//     footprint, which restores the painted-over look from every angle.

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/threedo"
)

// fleet is every car .wrapFam on the disc that decodes (28 of 29 — only the
// dev leftover CopMust.WrapFam.old has a type byte even the retail engine
// would misread). Names come from the filenames and the ORI3 model names;
// sections group the Studio browse list: the 8 player cars, the traffic
// fleet the player weaves through (plus the police Mustang that chases), and
// the cars nothing in the game spawns.
var fleet = []struct{ file, name, section string }{
	{"LDiablo.WrapFam", "Diablo", "Player cars"},
	{"F512TR.WrapFam", "512 TR", "Player cars"},
	{"P911.WrapFam", "911", "Player cars"},
	{"CZR1.WrapFam", "ZR-1", "Player cars"},
	{"DVIPER.WrapFam", "Viper", "Player cars"},
	{"ANSX.WrapFam", "NSX", "Player cars"},
	{"MRX7.WrapFam", "RX-7", "Player cars"},
	{"TSupra.WrapFam", "Supra", "Player cars"},
	{"CopMust.WrapFam", "Police Mustang", "Traffic vehicles"},
	{"BMW.WrapFam", "BMW", "Traffic vehicles"},
	{"CRX.WrapFam", "CRX", "Traffic vehicles"},
	{"GMCTRUCK.WrapFam", "GMC Truck", "Traffic vehicles"},
	{"Jeep.WrapFam", "Jeep", "Traffic vehicles"},
	{"Jetta.WrapFam", "Jetta", "Traffic vehicles"},
	{"Lemans.WrapFam", "Lemans", "Traffic vehicles"},
	{"Pickup.WrapFam", "Pickup", "Traffic vehicles"},
	{"PRELUDE.WrapFam", "Prelude", "Traffic vehicles"},
	{"PROBE.WrapFam", "Probe", "Traffic vehicles"},
	{"RODEO.WrapFam", "Rodeo", "Traffic vehicles"},
	{"SunBird.WrapFam", "Sunbird", "Traffic vehicles"},
	{"Vandura.WrapFam", "Vandura", "Traffic vehicles"},
	{"Wagon.WrapFam", "Wagon", "Traffic vehicles"},
	{"axxess.WrapFam", "Axxess", "Traffic vehicles"},
	{"Scooter.WrapFam", "Scooter", "Unused vehicles"},
	{"Porsche.WrapFam", "911 (classic body)", "Unused vehicles"},
	{"CopMust.WrapFam.new", "Police Mustang (dev)", "Unused vehicles"},
	{"Probe94.WrapFam", "Probe94 (SASCO model)", "Unused vehicles"},
	{"SASCO.WrapFam", "SASCO", "Unused vehicles"},
}

// exportCars writes every fleet car as one textured GLB and returns their
// manifest entries. Cars whose shapes carry the shared plt recolour chunks
// get one glTF scene per colour scheme in the same GLB — the Retro-X
// model-variant contract — named by each scheme's dominant body colour.
func exportCars(vol *threedo.Volume, out string) ([]ModelIndex, error) {
	var models []ModelIndex
	for _, c := range fleet {
		fam, err := vol.ReadFile("DriveData/CarData/" + c.file)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", c.file, err)
		}
		// "CopMust.WrapFam.new" -> "copmust-new"
		slug := strings.ToLower(strings.ReplaceAll(strings.Replace(c.file, ".WrapFam", "", 1), ".", "-"))

		lods, err := nfs.ParseCarFam(fam)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", c.file, err)
		}
		// highest detail = most faces; the same LOD index holds for every
		// recolour scheme (schemes swap palettes, not geometry)
		bi := 0
		for i, l := range lods {
			if len(l.Model.Faces) > len(lods[bi].Model.Faces) {
				bi = i
			}
		}
		schemes := lods[bi].Palettes
		if schemes < 1 {
			schemes = 1
		}
		schemeLODs := []*nfs.CarLOD{lods[bi]}
		for plt := 1; plt < schemes; plt++ {
			pl, err := nfs.ParseCarFamPalette(fam, plt)
			if err != nil {
				return nil, fmt.Errorf("%s plt%d: %v", c.file, plt, err)
			}
			schemeLODs = append(schemeLODs, pl[bi])
		}

		file := fmt.Sprintf("models/car-%s.glb", slug)
		path := filepath.Join(out, file)
		mi := ModelIndex{Name: c.name, File: file, Kind: "mesh3d", Section: c.section}
		if schemes == 1 {
			if err := writeORI3(schemeLODs[0], 1.0/128, path); err != nil {
				return nil, fmt.Errorf("%s: %v", c.file, err)
			}
		} else {
			var vars []glb.ModelVariant
			for plt, lod := range schemeLODs {
				v := buildORI3(lod, 1.0/128)
				v.Name = fmt.Sprintf("plt%d", plt)
				vars = append(vars, v)
			}
			if err := glb.WriteVariantScenes(path, vars); err != nil {
				return nil, fmt.Errorf("%s: %v", c.file, err)
			}
			for plt, name := range schemeNames(schemeLODs) {
				mi.Variants = append(mi.Variants, schema.ModelVariant{
					ID:          fmt.Sprintf("plt%d", plt),
					Name:        name,
					Scene:       fmt.Sprintf("plt%d", plt),
					Description: fmt.Sprintf("Shared-palette recolour scheme %d of %d (the game retargets each cel's plt0 PLUT pointer per spawned instance)", plt+1, schemes),
				})
			}
		}
		models = append(models, mi)
	}
	return models, nil
}

// writeORI3 writes a single-scene textured GLB from an ORI3 model + its
// face-texture map — the car path with one palette, and the track packets'
// scenery objects.
func writeORI3(lod *nfs.CarLOD, scale float64, path string) error {
	v := buildORI3(lod, scale)
	return glb.WriteTextured(path, v.Positions, v.UVs, v.TexGroups, nil)
}

// buildORI3 turns an ORI3 model + face-texture map into GLB-ready arrays.
// scale converts model units to world metres: ORI3 vertices are 1/128 m
// (the render path shifts them <<9 into 16.16 world space; the Diablo's
// ±292-unit length = 4.56 m checks out against the real car).
func buildORI3(lod *nfs.CarLOD, scale float64) glb.ModelVariant {
	m := lod.Model
	push := overlayPush(lod)

	var positions [][3]float32
	var uvs [][2]float32
	byTex := map[int][][3]uint32{}
	var order []int

	for fi, f := range m.Faces {
		ti := 0
		if fi < len(lod.FaceTex) {
			ti = lod.FaceTex[fi]
		}
		// quad corners in model units (v0=UV 0,0 v1=1,0 v2=1,1 v3=0,1),
		// overlay faces shifted out past the body they paint over
		var p [4][3]float64
		for k := 0; k < 4; k++ {
			v := m.Verts[f.V[k]]
			p[k] = [3]float64{float64(v.X) + push[fi], float64(v.Y), float64(v.Z)}
		}
		n := subdivN(p)
		// N×N bilinear grid: interior points sample the cel engine's own
		// (c,r) mapping, so texture lines stay straight across the quad
		base := uint32(len(positions))
		for j := 0; j <= n; j++ {
			for i := 0; i <= n; i++ {
				u, w := float64(i)/float64(n), float64(j)/float64(n)
				q := bilerp(p, u, w)
				positions = append(positions, [3]float32{
					float32(q[0] * scale),
					float32(q[1] * scale),
					-float32(q[2] * scale),
				})
				uvs = append(uvs, [2]float32{float32(u), float32(w)})
			}
		}
		if _, ok := byTex[ti]; !ok {
			order = append(order, ti)
		}
		stride := uint32(n + 1)
		for j := uint32(0); j < uint32(n); j++ {
			for i := uint32(0); i < uint32(n); i++ {
				a := base + j*stride + i
				b := a + 1
				c := a + stride + 1
				d := a + stride
				byTex[ti] = append(byTex[ti], [3]uint32{a, b, c}, [3]uint32{a, c, d})
			}
		}
	}
	v := glb.ModelVariant{Positions: positions, UVs: uvs}
	for _, ti := range order {
		v.TexGroups = append(v.TexGroups, glb.TexturedGroup{Tris: byTex[ti], Image: lod.Textures[ti]})
	}
	return v
}

// bilerp evaluates the quad's bilinear surface at (u,w): u runs along the
// v0→v1 edge, w along v0→v3 — matching the full-texture UV corners.
func bilerp(p [4][3]float64, u, w float64) [3]float64 {
	var q [3]float64
	for a := 0; a < 3; a++ {
		q[a] = (1-u)*(1-w)*p[0][a] + u*(1-w)*p[1][a] + u*w*p[2][a] + (1-u)*w*p[3][a]
	}
	return q
}

// subdivN picks the grid resolution for one quad. The two-triangle cut of a
// non-parallelogram quad deviates from the bilinear surface by up to |e|/4,
// e = v0−v1+v2−v3 (zero exactly when the quad is a parallelogram and the
// affine and bilinear mappings coincide). N×N subdivision scales the residual
// by 1/N²; N is chosen to keep it under tol model units (~1.6 cm).
func subdivN(p [4][3]float64) int {
	var e float64
	for a := 0; a < 3; a++ {
		d := p[0][a] - p[1][a] + p[2][a] - p[3][a]
		e += d * d
	}
	e = math.Sqrt(e)
	const tol = 2.0
	if e/4 <= tol {
		return 1
	}
	n := int(math.Ceil(math.Sqrt(e / 4 / tol)))
	if n > 8 {
		n = 8
	}
	return n
}

// overlayPush returns a signed X displacement (model units) per face index
// for the material-0x04 overlay faces that must clear the body: the game
// draws these last with no depth buffer, so their quads may sit inboard of
// the panels they cover. A face qualifies when it faces sideways (X-dominant
// normal, all corners one side of the centreline), its texture is bright
// enough to matter (the silver rims — a near-black backdrop hidden behind a
// near-black arch loses nothing, and the Viper flags ONLY backdrop discs),
// and no other overlay face lies outside it over the same footprint. The
// push lands it just outside the outermost body surface sampled inside its
// (y,z) footprint.
func overlayPush(lod *nfs.CarLOD) map[int]float64 {
	m := lod.Model
	bright := func(ti int) bool {
		if ti < 0 || ti >= len(lod.Textures) {
			return false
		}
		img := lod.Textures[ti]
		b := img.Bounds()
		var lum, n float64
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := img.RGBAAt(x, y)
				if c.A == 0 {
					continue
				}
				lum += (float64(c.R) + float64(c.G) + float64(c.B)) / 3 / 255
				n++
			}
		}
		return n > 0 && lum/n > 0.15
	}
	corners := func(f threedo.Face) (p [4][3]float64) {
		for k := 0; k < 4; k++ {
			v := m.Verts[f.V[k]]
			p[k] = [3]float64{float64(v.X), float64(v.Y), float64(v.Z)}
		}
		return
	}
	type cand struct {
		face   int
		sign   float64
		maxAbs float64
		bb     [2][2]float64 // (y,z) min/max
	}
	var cands []cand
	for fi, f := range m.Faces {
		if f.Material>>24&4 == 0 {
			continue
		}
		if fi >= len(lod.FaceTex) || !bright(lod.FaceTex[fi]) {
			continue
		}
		p := corners(f)
		// X-dominant normal (diagonal cross product), corners on one side
		n := [3]float64{}
		d1 := sub(p[2], p[0])
		d2 := sub(p[3], p[1])
		n[0] = d1[1]*d2[2] - d1[2]*d2[1]
		n[1] = d1[2]*d2[0] - d1[0]*d2[2]
		n[2] = d1[0]*d2[1] - d1[1]*d2[0]
		if math.Abs(n[0]) < math.Abs(n[1]) || math.Abs(n[0]) < math.Abs(n[2]) {
			continue
		}
		sign, maxAbs := 0.0, 0.0
		bb := [2][2]float64{{math.Inf(1), math.Inf(-1)}, {math.Inf(1), math.Inf(-1)}}
		ok := true
		for k := 0; k < 4; k++ {
			s := math.Copysign(1, p[k][0])
			if p[k][0] == 0 || (sign != 0 && s != sign) {
				ok = false
				break
			}
			sign = s
			maxAbs = math.Max(maxAbs, math.Abs(p[k][0]))
			bb[0][0] = math.Min(bb[0][0], p[k][1])
			bb[0][1] = math.Max(bb[0][1], p[k][1])
			bb[1][0] = math.Min(bb[1][0], p[k][2])
			bb[1][1] = math.Max(bb[1][1], p[k][2])
		}
		if ok {
			cands = append(cands, cand{fi, sign, maxAbs, bb})
		}
	}
	overlap := func(a, b [2][2]float64) bool {
		return a[0][0] <= b[0][1] && b[0][0] <= a[0][1] && a[1][0] <= b[1][1] && b[1][0] <= a[1][1]
	}
	push := map[int]float64{}
	for _, c := range cands {
		outermost := true
		for _, o := range cands {
			if o.face != c.face && o.sign == c.sign && o.maxAbs > c.maxAbs+0.5 && overlap(o.bb, c.bb) {
				outermost = false
				break
			}
		}
		if !outermost {
			continue
		}
		// outermost body surface inside the footprint, same side
		bodyMax := 0.0
		for fi, f := range m.Faces {
			if f.Material>>24&4 != 0 || fi == c.face {
				continue
			}
			p := corners(f)
			const grid = 8
			for j := 0; j <= grid; j++ {
				for i := 0; i <= grid; i++ {
					q := bilerp(p, float64(i)/grid, float64(j)/grid)
					if math.Copysign(1, q[0]) != c.sign {
						continue
					}
					if q[1] < c.bb[0][0] || q[1] > c.bb[0][1] || q[2] < c.bb[1][0] || q[2] > c.bb[1][1] {
						continue
					}
					bodyMax = math.Max(bodyMax, math.Abs(q[0]))
				}
			}
		}
		if bodyMax+0.5 >= c.maxAbs {
			push[c.face] = c.sign * (bodyMax + 2 - c.maxAbs)
		}
	}
	return push
}

func sub(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

// schemeNames labels each recolour scheme by the dominant colour of the
// pixels the scheme actually changes — the body ramp (palette entries 14+),
// since entries 0-13 (chrome, glass, tyres) are shared across schemes.
// Scheme k's changed pixels are found by diffing its decoded textures
// against scheme 0's (and scheme 0's against scheme 1's).
func schemeNames(lods []*nfs.CarLOD) []string {
	mean := func(a, b *nfs.CarLOD) (r, g, bl float64, n int) {
		for ti := range a.Textures {
			ia, ib := a.Textures[ti], b.Textures[ti]
			bounds := ia.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					ca, cb := ia.RGBAAt(x, y), ib.RGBAAt(x, y)
					if ca == cb || ca.A == 0 {
						continue
					}
					r += float64(ca.R) / 255
					g += float64(ca.G) / 255
					bl += float64(ca.B) / 255
					n++
				}
			}
		}
		return
	}
	names := make([]string, len(lods))
	used := map[string]int{}
	for k := range lods {
		ref := lods[0]
		if k == 0 {
			ref = lods[1]
		}
		r, g, b, n := mean(lods[k], ref)
		name := fmt.Sprintf("Scheme %d", k+1)
		if n > 0 {
			name = colourLabel(r/float64(n), g/float64(n), b/float64(n))
		}
		used[name]++
		if c := used[name]; c > 1 {
			name = fmt.Sprintf("%s %d", name, c)
		}
		names[k] = name
	}
	return names
}

// colourLabel names an sRGB colour (components 0..1) with a coarse
// human-readable bucket — enough to tell the Studio's variant entries apart.
func colourLabel(r, g, b float64) string {
	mx := math.Max(r, math.Max(g, b))
	mn := math.Min(r, math.Min(g, b))
	l := (mx + mn) / 2
	var s float64
	if mx > mn {
		if l > 0.5 {
			s = (mx - mn) / (2 - mx - mn)
		} else {
			s = (mx - mn) / (mx + mn)
		}
	}
	if s < 0.15 {
		switch {
		case l < 0.13:
			return "Black"
		case l < 0.3:
			return "Dark grey"
		case l < 0.55:
			return "Grey"
		case l < 0.8:
			return "Silver"
		default:
			return "White"
		}
	}
	var h float64 // hue in degrees
	d := mx - mn
	switch mx {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	name := ""
	switch {
	case h < 15 || h >= 345:
		name = "red"
	case h < 42:
		name = "orange"
	case h < 70:
		name = "yellow"
	case h < 158:
		name = "green"
	case h < 200:
		name = "teal"
	case h < 258:
		name = "blue"
	case h < 290:
		name = "purple"
	default:
		name = "pink"
	}
	if l < 0.22 {
		name = "dark " + name
	} else if l > 0.75 {
		name = "light " + name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
