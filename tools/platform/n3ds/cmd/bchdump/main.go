// bchdump prints what a game asset archive holds: the objects of every BCH
// content group, and, for a model, its skeleton and the shape of each mesh.
//
//	bchdump [-model NAME] [-bones] [-meshes] file.szs|file.bch [...]
//
// It reads either a `.bch` directly or any `.szs` (Yaz0 → SARC), dumping each
// BCH inside. Where bchglb converts a model to geometry, this one reports the
// numbers a decode is checked against — vertex counts, strides, matrix
// palettes, bind poses — so a mesh in a file can be matched against a draw the
// running game makes.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	model := flag.String("model", "", "only dump models whose name contains this")
	bones := flag.Bool("bones", false, "print every bone of each model's skeleton")
	texdump := flag.String("texdump", "", "write every decoded texture into this directory as PNG")
	listAllF := flag.Bool("list", false, "list every entry of every group, however many there are")
	mats := flag.Bool("mats", false, "with -meshes, print each material's combiner program")
	meshes := flag.Bool("meshes", false, "print every mesh: vertex count, stride, skin shape, matrix palette")
	flag.Parse()
	listAll = *listAllF
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: bchdump [-model NAME] [-bones] [-meshes] file.szs|file.bch […]")
		os.Exit(2)
	}
	for _, p := range flag.Args() {
		if err := dumpFile(p, *model, *bones, *meshes, *texdump, *mats); err != nil {
			fmt.Fprintln(os.Stderr, "bchdump:", err)
			os.Exit(1)
		}
	}
}

func dumpFile(path, want string, bones, meshes bool, texdump string, mats bool) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.EqualFold(filepath.Ext(path), ".bch") {
		return dumpBCH(path, raw, want, bones, meshes, texdump, mats)
	}
	arc, err := n3ds.OpenSZS(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	fmt.Printf("%s: %d files\n", path, len(arc.Files))
	for _, f := range arc.Files {
		fmt.Printf("  %-48s %8d bytes\n", f.Name, len(f.Data))
	}
	for _, f := range arc.Files {
		if !strings.HasSuffix(f.Name, ".bch") {
			continue
		}
		if err := dumpBCH(path+"/"+f.Name, f.Data, want, bones, meshes, texdump, mats); err != nil {
			return err
		}
	}
	return nil
}

var listAll bool

var groupNames = map[int]string{
	n3ds.BCHModels: "models", n3ds.BCHMaterials: "materials", n3ds.BCHShaders: "shaders",
	n3ds.BCHTextures: "textures", n3ds.BCHMaterialLUTs: "material LUTs", n3ds.BCHLights: "lights",
	n3ds.BCHCameras: "cameras", n3ds.BCHFogs: "fogs", n3ds.BCHSkeletalAnims: "skeletal anims",
	n3ds.BCHMaterialAnims: "material anims", n3ds.BCHVisibilityAnims: "visibility anims",
	n3ds.BCHLightAnims: "light anims", n3ds.BCHCameraAnims: "camera anims",
	n3ds.BCHFogAnims: "fog anims", n3ds.BCHScenes: "scenes",
}

func dumpBCH(name string, raw []byte, want string, bones, meshes bool, texdump string, mats bool) error {
	f, err := n3ds.ParseBCH(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	fmt.Printf("\n== %s ==\n", name)
	for g, es := range f.Groups {
		if len(es) == 0 {
			continue
		}
		fmt.Printf("  %-18s %d\n", groupNames[g], len(es))
		if g != n3ds.BCHModels && (len(es) <= 24 || listAll) {
			for _, e := range es {
				fmt.Printf("      %s\n", e.Name)
			}
		}
	}

	if texdump != "" {
		if err := os.MkdirAll(texdump, 0o755); err != nil {
			return err
		}
		for _, e := range f.Groups[n3ds.BCHTextures] {
			t, err := f.DecodeTexture(e)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			out, err := os.Create(filepath.Join(texdump, t.Name+".png"))
			if err != nil {
				return err
			}
			if err := png.Encode(out, t.Image); err != nil {
				out.Close()
				return err
			}
			out.Close()
			fmt.Printf("  texture %-32s %3dx%-3d format 0x%X\n", t.Name, t.Width, t.Height, t.Format)
		}
	}

	for _, e := range f.Groups[n3ds.BCHModels] {
		if want != "" && !strings.Contains(e.Name, want) {
			continue
		}
		m, err := f.DecodeModel(e)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		tris := 0
		for _, sh := range m.Meshes {
			tris += len(sh.Indices) / 3
		}
		fmt.Printf("\n  model %s: %d meshes, %d materials, %d bones, %d tris\n",
			m.Name, len(m.Meshes), len(m.Materials), len(m.Bones), tris)

		// The two extents side by side. A skinned model's vertices are stored in
		// the model's own space, so at the bind pose they must envelop the
		// skeleton posed in that same space: if the two boxes do not overlap,
		// one of the decodes is in the wrong space and no amount of skinning
		// arithmetic will bring them together.
		lo, hi := extent(func(add func([3]float64)) {
			for _, sh := range m.Meshes {
				for _, v := range sh.Verts {
					add([3]float64{float64(v.Pos[0]), float64(v.Pos[1]), float64(v.Pos[2])})
				}
			}
		})
		fmt.Printf("    vertices  x %8.1f..%-8.1f y %8.1f..%-8.1f z %8.1f..%-8.1f\n",
			lo[0], hi[0], lo[1], hi[1], lo[2], hi[2])
		world := m.BindPose()
		if len(world) > 0 {
			blo, bhi := extent(func(add func([3]float64)) {
				for _, w := range world {
					add([3]float64{w[3], w[7], w[11]})
				}
			})
			fmt.Printf("    bones     x %8.1f..%-8.1f y %8.1f..%-8.1f z %8.1f..%-8.1f\n",
				blo[0], bhi[0], blo[1], bhi[1], blo[2], bhi[2])
		}

		if bones {
			for i, b := range m.Bones {
				w := world[i]
				fmt.Printf("    bone %2d %-20s parent %3d  world (%8.2f %8.2f %8.2f)  t=%v r=%v\n",
					i, b.Name, b.Parent, w[3], w[7], w[11], b.Trans, b.Rotate)
			}
		}
		if meshes {
			for i, sh := range m.Meshes {
				mat, tex := "", ""
				if sh.MaterialIndex < len(m.Materials) {
					mat = m.Materials[sh.MaterialIndex].Name
					tex = fmt.Sprintf("%q", m.Materials[sh.MaterialIndex].Names)
				}
				mlo, mhi := extent(func(add func([3]float64)) {
					for _, v := range sh.Verts {
						add([3]float64{float64(v.Pos[0]), float64(v.Pos[1]), float64(v.Pos[2])})
					}
				})
				fmt.Printf("    mesh %2d %-24s verts %5d idx %5d stride %3d uv %d skin=%d palette %v\n",
					i, mat, len(sh.Verts), len(sh.Indices), sh.VertexStride, sh.UVCount,
					sh.SkinMode, sh.Palette)
				fmt.Printf("             extent  (%7.1f %7.1f %7.1f)..(%7.1f %7.1f %7.1f)  textures %s\n",
					mlo[0], mlo[1], mlo[2], mhi[0], mhi[1], mhi[2], tex)
				if mats && sh.MaterialIndex < len(m.Materials) {
					mm := &m.Materials[sh.MaterialIndex]
					fmt.Printf("             blend %v (src %d dst %d, additive %v) alphaTest %v func %d ref %d\n",
						mm.Blends, mm.BlendSrc, mm.BlendDst, mm.Additive(), mm.AlphaTest, mm.AlphaFunc, mm.AlphaRef)
					for si, line := range m.Materials[sh.MaterialIndex].Describe() {
						fmt.Printf("             tev%d %s\n", si, line)
					}
				}
				// Beside it, where the palette's bones stand in the bind pose. If
				// the mesh's vertices are in the model's space the two agree; if
				// they are in a bone's, the mesh sits at the origin instead.
				for si, b := range sh.Palette {
					if b < len(world) {
						fmt.Printf("             slot %d = bone %2d %-16s at (%7.1f %7.1f %7.1f)\n",
							si, b, m.Bones[b].Name, world[b][3], world[b][7], world[b][11])
					}
				}
			}
		}
	}
	return nil
}

// extent runs a collector over a point set and returns its bounding box.
func extent(each func(add func([3]float64))) (lo, hi [3]float64) {
	first := true
	each(func(p [3]float64) {
		if first {
			lo, hi, first = p, p, false
			return
		}
		for i := 0; i < 3; i++ {
			if p[i] < lo[i] {
				lo[i] = p[i]
			}
			if p[i] > hi[i] {
				hi[i] = p[i]
			}
		}
	})
	return lo, hi
}
