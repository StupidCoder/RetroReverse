package main

// car.go exports one car .wrapFam as a textured GLB: the highest-detail
// (model, shape) LOD, each quad face textured by its SPoT via the "!ori"
// face map. ORI3 vertices are 16.16 world units, so the same 1 unit = 1 m
// scale as the course applies.

import (
	"fmt"
	"path/filepath"
	"strings"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/lib/glb"
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

// exportCars writes the highest-detail LOD of every fleet car as a textured
// GLB and returns their manifest entries. Cars whose shapes carry the shared
// plt recolour chunks get one GLB per colour scheme.
func exportCars(vol *threedo.Volume, out string) ([]ModelIndex, error) {
	var models []ModelIndex
	for _, c := range fleet {
		fam, err := vol.ReadFile("DriveData/CarData/" + c.file)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", c.file, err)
		}
		// "CopMust.WrapFam.new" -> "copmust-new"
		slug := strings.ToLower(strings.ReplaceAll(strings.Replace(c.file, ".WrapFam", "", 1), ".", "-"))
		schemes := 1
		for plt := 0; plt < schemes; plt++ {
			lods, err := nfs.ParseCarFamPalette(fam, plt)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", c.file, err)
			}
			// highest detail = most faces
			best := lods[0]
			for _, l := range lods[1:] {
				if len(l.Model.Faces) > len(best.Model.Faces) {
					best = l
				}
			}
			if plt == 0 && best.Palettes > 1 {
				schemes = best.Palettes
			}
			file := fmt.Sprintf("models/car-%s.glb", slug)
			name := c.name
			if plt > 0 {
				file = fmt.Sprintf("models/car-%s-plt%d.glb", slug, plt)
				name = fmt.Sprintf("%s (scheme %d)", c.name, plt+1)
			}
			if err := writeORI3(best, 1.0/128, filepath.Join(out, file)); err != nil {
				return nil, fmt.Errorf("%s: %v", c.file, err)
			}
			models = append(models, ModelIndex{
				Name: name, File: file, Kind: "mesh3d", Section: c.section,
			})
		}
	}
	return models, nil
}

// writeORI3 writes a textured GLB from an ORI3 model + its face-texture map.
// scale converts model units to world metres: ORI3 vertices are 1/128 m
// (the render path shifts them <<9 into 16.16 world space; the Diablo's
// ±292-unit length = 4.56 m checks out against the real car).
func writeORI3(lod *nfs.CarLOD, scale float64, path string) error {
	type vk struct {
		v    int
		u, w float32
	}
	index := map[vk]uint32{}
	var positions [][3]float32
	var uvs [][2]float32
	vertex := func(v int, u, w float32) uint32 {
		k := vk{v, u, w}
		if i, ok := index[k]; ok {
			return i
		}
		mv := lod.Model.Verts[v]
		i := uint32(len(positions))
		positions = append(positions, [3]float32{
			float32(float64(mv.X) * scale),
			float32(float64(mv.Y) * scale),
			-float32(float64(mv.Z) * scale),
		})
		uvs = append(uvs, [2]float32{u, w})
		index[k] = i
		return i
	}

	uvc := [4][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	byTex := map[int][][3]uint32{}
	var order []int
	for fi, f := range lod.Model.Faces {
		ti := 0
		if fi < len(lod.FaceTex) {
			ti = lod.FaceTex[fi]
		}
		var idx [4]uint32
		for k := 0; k < 4; k++ {
			idx[k] = vertex(f.V[k], uvc[k][0], uvc[k][1])
		}
		if _, ok := byTex[ti]; !ok {
			order = append(order, ti)
		}
		byTex[ti] = append(byTex[ti],
			[3]uint32{idx[0], idx[1], idx[2]}, [3]uint32{idx[0], idx[2], idx[3]})
	}
	var groups []glb.TexturedGroup
	for _, ti := range order {
		groups = append(groups, glb.TexturedGroup{Tris: byTex[ti], Image: lod.Textures[ti]})
	}
	return glb.WriteTextured(path, positions, uvs, groups, nil)
}
