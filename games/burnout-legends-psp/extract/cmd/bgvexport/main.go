// bgvexport writes Burnout Legends' vehicles out as GLB models with their
// texture atlas embedded, decoded straight from the UMD by the bgv package.
//
// Each car becomes one GLB: the highest detail level's meshes, textured with
// the model's own 256x256 atlas. -png also writes the atlas out, which is the
// quickest way to see that the texture decode is right.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/burnout-legends-psp/extract/bgt"
	"retroreverse.com/games/burnout-legends-psp/extract/bgv"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/psp"
)

func main() {
	image := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	out := flag.String("out", "", "output directory for the GLBs")
	only := flag.String("only", "", "export just this car (e.g. Car1)")
	level := flag.Int("level", -1, "detail level to export (default: the highest)")
	pngDir := flag.String("png", "", "also write each car's texture atlas to this directory")
	flag.Parse()

	if *image == "" || *out == "" {
		die("need -image and -out")
	}
	im, err := psp.OpenImage(*image)
	if err != nil {
		die("open image: %v", err)
	}
	defer im.Close()

	var paths []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, ".bgv") {
			if *only == "" || strings.Contains(e.Path, *only) {
				paths = append(paths, e.Path)
			}
		}
		return nil
	})
	sort.Strings(paths)
	if len(paths) == 0 {
		die("no vehicle models matched")
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}
	if *pngDir != "" {
		if err := os.MkdirAll(*pngDir, 0o755); err != nil {
			die("%v", err)
		}
	}

	written := 0
	for _, p := range paths {
		// Vehicles are grouped by class — COMP, CUPE, HEVY, MSCL, SPRT, SUPR,
		// the same classes the SELECT CAR menu offers — and the numbering
		// restarts in each, so the class belongs in the name.
		name := strings.TrimSuffix(filepath.Base(p), ".bgv")
		if dir := filepath.Base(filepath.Dir(p)); dir != "" {
			name = dir + "_" + name
		}
		data, err := im.ReadFile(p)
		if err != nil {
			die("read %s: %v", p, err)
		}
		m, err := bgv.Parse(data)
		if err != nil {
			die("parse %s: %v", name, err)
		}
		atlas, err := bgv.Texture(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: no texture (%v) — exporting untextured\n", name, err)
		}
		// Run the engine's alpha test before the atlas reaches a glTF material:
		// a car's atlas is 15% middling alpha, and glTF's 0.5 mask cutoff punches
		// every one of those texels out, returning the body as a shell full of
		// holes. (The -png dump above stays the raw decode.)
		masked := atlas
		if atlas != nil {
			masked = bgt.AlphaTest(atlas)
		}
		if *pngDir != "" && atlas != nil {
			f, err := os.Create(filepath.Join(*pngDir, name+".png"))
			if err != nil {
				die("%v", err)
			}
			if err := png.Encode(f, atlas); err != nil {
				die("%v", err)
			}
			f.Close()
		}

		lvl := *level
		if lvl < 0 || lvl >= len(m.Levels) {
			lvl = len(m.Levels) - 1 // the last table is the most detailed
		}
		meshes := m.Levels[lvl]

		places, err := bgv.Placements(data, meshes)
		if err != nil {
			die("%s: %v", name, err)
		}

		sc := glb.NewScene()
		node := sc.AddNode(name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		var prims []glb.Prim
		tris := 0
		for i := range meshes {
			me := &meshes[i]
			t := me.Tris()
			if len(t) == 0 {
				continue
			}
			// The body where its vertices already put it, each detachable panel
			// at its own placement, the wheel at all four corners.
			for _, m := range places[i] {
				tris += len(t)
				pos := make([][3]float32, len(me.Verts))
				nrm := make([][3]float32, len(me.Verts))
				uv := make([][2]float32, len(me.Verts))
				for j, v := range me.Verts {
					x, y, z := m.Apply(v.X, v.Y, v.Z)
					nx, ny, nz := m.ApplyDir(v.NX, v.NY, v.NZ)
					// The model's own axes are already Y-up (the body's vertices
					// run from the floor up to +1.08); glTF is Y-up too. Flip Z
					// instead, to swap the handedness.
					pos[j] = [3]float32{x, y, z}
					nrm[j] = [3]float32{nx, ny, nz}
					uv[j] = [2]float32{v.U, v.V}
				}
				// The models are right-handed and Y-up already — glTF's convention
				// — so no axis flip. Only a mirroring placement (the wheels on one
				// side, instanced through a negated basis) reverses winding.
				prims = append(prims, glb.Prim{
					Positions: pos,
					Normals:   nrm,
					UVs:       uv,
					Tris:      wound(t, m.Mirrors()),
					Image:     masked,
					BaseColor: [4]float32{1, 1, 1, 1},
					Unlit:     true,
				})
			}
		}
		if len(prims) == 0 {
			fmt.Fprintf(os.Stderr, "%s: nothing to export\n", name)
			continue
		}
		if err := sc.AddMesh(node, name, prims); err != nil {
			die("%s: %v", name, err)
		}
		dst := filepath.Join(*out, name+".glb")
		if err := sc.Write(dst, name); err != nil {
			die("%s: %v", name, err)
		}
		st, _ := os.Stat(dst)
		fmt.Printf("%-8s level %d  %2d meshes  %5d tris  -> %s (%d KB)\n",
			name, lvl, len(prims), tris, filepath.Base(dst), st.Size()/1024)
		written++
	}
	fmt.Printf("\nwrote %d vehicle GLBs to %s\n", written, *out)
}

// wound reverses a triangle list's winding when a transform has flipped
// handedness, so glTF's counter-clockwise front face still faces out.
func wound(t [][3]uint32, reverse bool) [][3]uint32 {
	if !reverse {
		return t
	}
	out := make([][3]uint32, len(t))
	for i, f := range t {
		out[i] = [3]uint32{f[0], f[2], f[1]}
	}
	return out
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bgvexport: "+format+"\n", a...)
	os.Exit(1)
}
