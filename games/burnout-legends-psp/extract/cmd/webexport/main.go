// webexport builds the Studio's asset tree for Burnout Legends: every vehicle
// and every track, decoded from the UMD and written as GLBs with a manifest.
//
// Vehicles come from the bgv package (89 cars, one GLB each, assembled from the
// body meshes plus the wheel instanced at its four placements). Tracks come from
// the bgt package, and each ships as TWO GLBs — the streamed world you drive
// through, and the static environment it sits in. The game swaps between those
// layers rather than adding them, so the viewer keeps them apart too.
//
// Tracks are named by their identity on the disc (region and track directory,
// e.g. "US C3 V1"). The game holds no readable track names — tlist.bin is
// binary and the front end's strings live in its XUI resources — so inventing
// prettier ones would mean going outside the image.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/burnout-legends-psp/extract/bgt"
	"retroreverse.com/games/burnout-legends-psp/extract/bgv"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/psp"
)

type model struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	Kind    string `json:"kind"`
	File    string `json:"file"`
	Env     string `json:"env,omitempty"`   // bl-track: the environment layer
	Spawn   *spawn `json:"spawn,omitempty"` // bl-track: where to open the camera
}

// spawn opens the viewer on the road rather than out at the bounding sphere of a
// long thin circuit, which leaves it a speck on the screen.
//
// This is a presentation choice, not a fact about the format: nothing in the
// files says "the camera goes here". What the files do give is the road itself,
// and roadFrom() picks it out of a cell. Cell anchors and cell bounding boxes
// both looked promising and both put the camera in a hedge — the anchors are the
// corners of a cell's footprint, and a cell holds so much hillside that the
// middle of its box is off in the scenery.
type spawn struct {
	Pos    [3]float32 `json:"pos"`
	Target [3]float32 `json:"target"`
}

// roadFrom returns a point on the road in a cell, in glTF coordinates (Z
// flipped, as the meshes are).
//
// The road is the one thing in a cell shaped like a road: WIDE and FLAT. Score
// each mesh by its footprint over its thickness and the tarmac wins outright —
// buildings are tall, props are small, the hillside is neither wide nor flat. The
// winning mesh's centre sits mid-slab, so lifting the eye by a few times the
// slab's own half-thickness puts it just above the surface, at a scale that
// travels from track to track.
func roadFrom(b bgt.Block) ([3]float32, bool) {
	best, bestScore := -1, float32(0)
	for i := range b.Meshes {
		e := b.Meshes[i].Extent
		thick := max(e[1], 1)
		if s := e[0] * e[2] / thick; s > bestScore {
			best, bestScore = i, s
		}
	}
	if best < 0 {
		return [3]float32{}, false
	}
	m := &b.Meshes[best]
	return [3]float32{m.Centre[0], m.Centre[1] + 3*max(m.Extent[1], 1), -m.Centre[2]}, true
}

type manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Models   []model        `json:"models"`
}

func main() {
	in := flag.String("in", "", "PSP UMD image (.cso or .iso)")
	out := flag.String("o", "../../site/public/burnout-legends-psp", "output root")
	only := flag.String("only", "all", "models|all")
	flag.Parse()

	if *in == "" {
		die("need -in")
	}
	im, err := psp.OpenImage(*in)
	if err != nil {
		die("open image: %v", err)
	}
	defer im.Close()

	models := filepath.Join(*out, "models")
	if err := os.MkdirAll(models, 0o755); err != nil {
		die("%v", err)
	}
	want := *only == "all" || strings.Contains(*only, "models")
	if !want {
		die("nothing to do for -only %q", *only)
	}

	m := manifest{
		Format: 2, Game: "burnout-legends-psp", Platform: "Sony PSP",
		Native: map[string]int{"w": 480, "h": 272}, TickHz: 60,
	}
	m.Models = append(m.Models, exportTracks(im, models)...)
	m.Models = append(m.Models, exportCars(im, models)...)

	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		die("%v", err)
	}
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(buf, '\n'), 0o644); err != nil {
		die("%v", err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d models -> %s\n", len(m.Models), filepath.Join(*out, "manifest.json"))
}

// exportTracks writes each track's two layers and returns their manifest entries.
func exportTracks(im *psp.Image, dir string) []model {
	var dirs []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, "/static.dat") {
			dirs = append(dirs, path.Dir(e.Path))
		}
		return nil
	})
	sort.Strings(dirs)

	var out []model
	for i, d := range dirs {
		id := strings.TrimPrefix(d, "PSP_GAME/USRDIR/Tracks/") // e.g. US/C3_V1
		slug := strings.ReplaceAll(id, "/", "_")
		region, track, _ := strings.Cut(id, "/")

		sd, err := im.ReadFile(path.Join(d, "static.dat"))
		if err != nil {
			die("read %s: %v", d, err)
		}
		t, err := bgt.Parse(sd)
		if err != nil {
			die("%s: %v", id, err)
		}
		var world []bgt.Mesh
		var anchors []bgt.Block
		if raw, err := im.ReadFile(path.Join(d, "streamed.dat")); err == nil {
			bs, err := t.ParseStreamed(raw)
			if err != nil {
				die("%s: %v", id, err)
			}
			for _, b := range bs {
				if len(b.Meshes) == 0 {
					continue // the other block of each cell carries no geometry
				}
				world = append(world, b.Meshes...)
				anchors = append(anchors, b)
			}
		}
		env := t.Meshes
		if len(world) == 0 { // a crash junction streams nothing
			world, env = env, nil
		}

		e := model{
			Name:    region + " " + strings.ReplaceAll(track, "_", " "),
			Section: "Tracks",
			Kind:    "bl-track",
			File:    "models/track-" + slug + ".glb",
		}
		if len(anchors) >= 2 {
			pos, ok1 := roadFrom(anchors[0])
			tgt, ok2 := roadFrom(anchors[1])
			if ok1 && ok2 {
				tgt[1] = pos[1] // look along the road, not up or down the next cell
				e.Spawn = &spawn{Pos: pos, Target: tgt}
			}
		}
		writeGLB(filepath.Join(dir, "track-"+slug+".glb"), slug, t, world)
		if len(env) > 0 {
			e.Env = "models/track-" + slug + "_env.glb"
			writeGLB(filepath.Join(dir, "track-"+slug+"_env.glb"), slug+"_env", t, env)
		}
		out = append(out, e)
		fmt.Fprintf(os.Stderr, "[tracks] %2d/%d  %s (%d world meshes, %d environment)\n",
			i+1, len(dirs), id, len(world), len(env))
	}
	return out
}

// exportCars writes one GLB per vehicle: its highest detail level, assembled.
func exportCars(im *psp.Image, dir string) []model {
	var paths []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, ".bgv") {
			paths = append(paths, e.Path)
		}
		return nil
	})
	sort.Strings(paths)

	var out []model
	for i, p := range paths {
		// Vehicles are grouped by class — the classes the SELECT CAR menu offers —
		// and the numbering restarts in each, so the class belongs in the name.
		class := filepath.Base(filepath.Dir(p))
		car := strings.TrimSuffix(filepath.Base(p), ".bgv")
		slug := class + "_" + car

		data, err := im.ReadFile(p)
		if err != nil {
			die("read %s: %v", p, err)
		}
		mdl, err := bgv.Parse(data)
		if err != nil {
			die("%s: %v", slug, err)
		}
		atlas, err := bgv.Texture(data)
		if err != nil {
			die("%s: %v", slug, err)
		}
		wheels, err := bgv.Wheels(data)
		if err != nil {
			die("%s: %v", slug, err)
		}

		// The three mesh tables are detail levels, least detailed first; ship the
		// one the game shows you in the car park.
		meshes := mdl.Levels[len(mdl.Levels)-1]

		var prims []glb.Prim
		for mi := range meshes {
			me := &meshes[mi]
			tris := me.Tris()
			if len(tris) == 0 {
				continue
			}
			// Body parts carry their position in their vertices; the wheel is
			// modelled about the origin and instanced at the four placements.
			places := []bgv.Mat4{identity}
			if me.AtOrigin() {
				places = wheels
			}
			for _, w := range places {
				pos := make([][3]float32, len(me.Verts))
				nrm := make([][3]float32, len(me.Verts))
				uv := make([][2]float32, len(me.Verts))
				for j, v := range me.Verts {
					x, y, z := w.Apply(v.X, v.Y, v.Z)
					nx, ny, nz := w.ApplyDir(v.NX, v.NY, v.NZ)
					pos[j] = [3]float32{x, y, -z} // the models are already Y-up; flip Z
					nrm[j] = [3]float32{nx, ny, -nz}
					uv[j] = [2]float32{v.U, v.V}
				}
				// Unlit, like everything else the Studio ships: the stage carries
				// no lights, so a PBR material renders black. The car's paint,
				// decals and lettering are all in its atlas anyway.
				prims = append(prims, glb.Prim{
					Positions: pos, Normals: nrm, UVs: uv, Tris: tris,
					Image: atlas, BaseColor: [4]float32{1, 1, 1, 1}, Unlit: true,
				})
			}
		}
		if len(prims) == 0 {
			continue
		}
		sc := glb.NewScene()
		node := sc.AddNode(slug, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		if err := sc.AddMesh(node, slug, prims); err != nil {
			die("%s: %v", slug, err)
		}
		file := "models/car-" + slug + ".glb"
		if err := sc.Write(filepath.Join(dir, "car-"+slug+".glb"), slug); err != nil {
			die("%s: %v", slug, err)
		}
		out = append(out, model{
			Name:    class + " " + car,
			Section: "Cars",
			Kind:    "mesh3d",
			File:    file,
		})
		if (i+1)%10 == 0 || i+1 == len(paths) {
			fmt.Fprintf(os.Stderr, "[cars] %2d/%d  %s\n", i+1, len(paths), slug)
		}
	}
	return out
}

// writeGLB emits one track layer, batched by material as the engine draws it.
func writeGLB(file, name string, t *bgt.Track, meshes []bgt.Mesh) {
	type batch struct {
		pos  [][3]float32
		uv   [][2]float32
		col  [][4]uint8
		tris [][3]uint32
	}
	batches := map[int]*batch{}
	var order []int
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
			// The track's axes are already Y-up, as glTF's are; flip Z for handedness.
			b.pos = append(b.pos, [3]float32{v.X, v.Y, -v.Z})
			b.uv = append(b.uv, [2]float32{v.U, v.V})
			b.col = append(b.col, [4]uint8{v.R, v.G, v.B, v.A})
		}
		for _, f := range tr {
			b.tris = append(b.tris, [3]uint32{base + f[0], base + f[1], base + f[2]})
		}
	}
	sort.Ints(order)

	// The world is drawn unlit — its lighting is baked into the vertex colours.
	cache := map[int]image.Image{}
	var prims []glb.Prim
	for _, mat := range order {
		b := batches[mat]
		p := glb.Prim{
			Positions: b.pos, UVs: b.uv, Colors: b.col, Tris: b.tris,
			BaseColor: [4]float32{1, 1, 1, 1}, Unlit: true,
		}
		if mat >= 0 {
			if ti := t.Materials[mat].Texture; ti >= 0 {
				tx, ok := cache[ti]
				if !ok {
					img, err := t.Texture(ti)
					if err != nil {
						die("%s: texture %d: %v", name, ti, err)
					}
					// Run the engine's own alpha test, so glTF's 0.5 mask cutoff
					// keeps the solid world and cuts only the real cutouts.
					tx = bgt.AlphaTest(img)
					cache[ti] = tx
				}
				p.Image = tx
			}
		}
		prims = append(prims, p)
	}
	if len(prims) == 0 {
		return
	}
	sc := glb.NewScene()
	node := sc.AddNode(name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	if err := sc.AddMesh(node, name, prims); err != nil {
		die("%s: %v", name, err)
	}
	if err := sc.Write(file, name); err != nil {
		die("%s: %v", name, err)
	}
}

// identity leaves a mesh where its vertices already put it.
var identity = bgv.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(1)
}
