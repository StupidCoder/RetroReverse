// bgtexport writes Burnout Legends' tracks out as GLB models with their
// textures embedded, decoded straight from the UMD by the bgt package.
//
// Each track becomes two GLBs, because the game keeps its world in two layers
// and swaps between them:
//
//	<TRACK>.glb      the streamed world — the city you actually drive through
//	<TRACK>_env.glb  static.dat — the environment it sits in: terrain, cliffs,
//	                 forest, the mountain backdrop
//
// They are not additive. static.dat carries a COARSE stand-in for the ground the
// streamed cells cover, and the engine draws a static group only where the
// corresponding cell is not resident. Which groups those are is a runtime
// decision we have not reversed (the cell data lives in Gamedata.bgd, which
// holds no geometry), so drawing both layers at once double-covers the ground
// near the road — the coarse terrain wins the depth test over the real one.
// Keeping them apart lets a viewer show the world, the environment, or both.
//
// Meshes are batched by material, which is how the engine draws them too, so a
// GLB carries one primitive per texture.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/burnout-legends-psp/extract/bgt"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/psp"
)

func main() {
	img := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	out := flag.String("out", "", "output directory for the GLBs")
	only := flag.String("only", "", "export just this track (e.g. US/C3_V1)")
	pngDir := flag.String("png", "", "also write each track's textures to this directory")
	flag.Parse()

	if *img == "" || *out == "" {
		die("need -image and -out")
	}
	im, err := psp.OpenImage(*img)
	if err != nil {
		die("open image: %v", err)
	}
	defer im.Close()

	var dirs []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, "/static.dat") {
			dirs = append(dirs, path.Dir(e.Path))
		}
		return nil
	})
	sort.Strings(dirs)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}

	written := 0
	for _, dir := range dirs {
		name := strings.ReplaceAll(strings.TrimPrefix(dir, "PSP_GAME/USRDIR/Tracks/"), "/", "_")
		if *only != "" && !strings.Contains(dir, *only) {
			continue
		}
		sd, err := im.ReadFile(path.Join(dir, "static.dat"))
		if err != nil {
			die("read %s: %v", dir, err)
		}
		t, err := bgt.Parse(sd)
		if err != nil {
			die("%s: %v", name, err)
		}
		var world []bgt.Mesh
		blocks := 0
		if raw, err := im.ReadFile(path.Join(dir, "streamed.dat")); err == nil {
			bs, err := t.ParseStreamed(raw)
			if err != nil {
				die("%s: %v", name, err)
			}
			blocks = len(bs)
			for _, b := range bs {
				world = append(world, b.Meshes...)
			}
		}
		// A crash junction streams nothing: its only geometry is the resident set.
		suffix := "_env"
		if len(world) == 0 {
			world, t.Meshes = t.Meshes, nil
			suffix = ""
		}
		wTris := emit(*out, name, name, t, world)
		eTris := 0
		if len(t.Meshes) > 0 && suffix != "" {
			eTris = emit(*out, name+suffix, name+suffix, t, t.Meshes)
		}
		fmt.Printf("%-10s world: %4d meshes (%2d blocks) %6d tris   environment: %3d meshes %6d tris\n",
			name, len(world), blocks, wTris, len(t.Meshes), eTris)
		written++

		if *pngDir != "" {
			d := filepath.Join(*pngDir, name)
			if err := os.MkdirAll(d, 0o755); err != nil {
				die("%v", err)
			}
			for i := range t.Textures {
				tx, err := t.Texture(i)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s texture %d: %v\n", name, i, err)
					continue
				}
				writePNG(filepath.Join(d, fmt.Sprintf("tex%03d.png", i)), tx)
			}
		}
	}
	fmt.Printf("\nwrote %d track GLBs to %s\n", written, *out)
}

// emit writes one layer of a track as a GLB and returns its triangle count.
func emit(dir, file, name string, t *bgt.Track, meshes []bgt.Mesh) int {
	prims, tris, err := build(t, meshes)
	if err != nil {
		die("%s: %v", name, err)
	}
	if len(prims) == 0 {
		return 0
	}
	sc := glb.NewScene()
	node := sc.AddNode(name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	if err := sc.AddMesh(node, name, prims); err != nil {
		die("%s: %v", name, err)
	}
	if err := sc.Write(filepath.Join(dir, file+".glb"), name); err != nil {
		die("%s: %v", name, err)
	}
	return tris
}

// build batches the meshes by material — one primitive per texture, which is how
// the engine draws them too — and converts to glTF's coordinate convention.
func build(t *bgt.Track, meshes []bgt.Mesh) ([]glb.Prim, int, error) {
	type batch struct {
		pos  [][3]float32
		uv   [][2]float32
		col  [][4]uint8
		tris [][3]uint32
	}
	batches := map[int]*batch{}
	var order []int
	tris := 0
	for i := range meshes {
		m := &meshes[i]
		tr := m.Tris()
		if len(tr) == 0 {
			continue
		}
		b := batches[m.Material]
		if b == nil {
			b = &batch{}
			batches[m.Material] = b
			order = append(order, m.Material)
		}
		base := uint32(len(b.pos))
		for _, v := range m.Verts {
			// The track's own axes are already Y-up (heights run positive), as
			// glTF's are. Flip Z to swap the handedness, exactly as the vehicles do.
			b.pos = append(b.pos, [3]float32{v.X, v.Y, -v.Z})
			b.uv = append(b.uv, [2]float32{v.U, v.V})
			b.col = append(b.col, [4]uint8{v.R, v.G, v.B, v.A})
		}
		for _, f := range tr {
			b.tris = append(b.tris, [3]uint32{base + f[0], base + f[1], base + f[2]})
		}
		tris += len(tr)
	}
	sort.Ints(order)

	// The world is drawn unlit: the lighting is baked into the vertex colours.
	cache := map[int]image.Image{}
	var prims []glb.Prim
	for _, mat := range order {
		b := batches[mat]
		p := glb.Prim{
			Positions: b.pos,
			UVs:       b.uv,
			Colors:    b.col,
			Tris:      b.tris,
			BaseColor: [4]float32{1, 1, 1, 1},
			Unlit:     true,
		}
		if mat >= 0 {
			if ti := t.Materials[mat].Texture; ti >= 0 {
				tx, ok := cache[ti]
				if !ok {
					im, err := t.Texture(ti)
					if err != nil {
						return nil, 0, err
					}
					tx = im
					cache[ti] = tx
				}
				p.Image = tx
			}
		}
		prims = append(prims, p)
	}
	return prims, tris, nil
}

func writePNG(path string, im image.Image) {
	f, err := os.Create(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, im); err != nil {
		die("%v", err)
	}
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bgtexport: "+format+"\n", a...)
	os.Exit(1)
}
