// exportkcl converts .kcl collision meshes to viewer GLBs (Part VI §7).
//
// The triangles are reconstructed from the prism records exactly as the
// game's walker defines them — each prism is the region of its face plane cut
// by the three edge half-planes, so the corners are the pairwise plane
// intersections (sm64ds/kcl.go Corners). Every level's collision map is
// exported, plus every object collider the actor-binding oracle saw an actor
// load (actorbind.json "kcl" entries — platform/mechanism actors registering
// dBgW_Kc colliders in their init).
//
// Output space matches the stage GLBs: world-fx / 4096 / 1000 (a placement
// short is world/1000 in stage-GLB units, Part V §5), so the meshes overlay
// the render geometry directly and object colliders sit at the placement
// with no extra scale. Triangles carry a normal-shaded red as baked vertex
// colors (unlit, like all static DS geometry).
//
//	exportkcl [-rom img] [-extracted dir] [-bind actorbind.json] [-o dir]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/nds"
	"retroreverse.com/tools/nds/nitro"
	"supermario64ds/extract/sm64ds"
)

const toGLB = 1.0 / 4096 / 1000 // fx20.12 world -> stage-GLB units

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	bindPath := flag.String("bind", "../extracted/actorbind.json", "oracle binding table")
	outDir := flag.String("o", "../extracted/collision", "output dir")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		sm64ds.Die(err)
	}

	// stem -> internal path for every .kcl in the file table
	kclPath := map[string]string{}
	for i := 0; i < 2058; i++ {
		if n := ls.InternalName(i); strings.HasSuffix(n, ".kcl") {
			stem := strings.TrimSuffix(filepath.Base(n), ".kcl")
			if _, dup := kclPath[stem]; !dup {
				kclPath[stem] = n
			}
		}
	}

	export := func(stem string) error {
		p, ok := kclPath[stem]
		if !ok {
			return fmt.Errorf("no .kcl named %s in the file table", stem)
		}
		data, err := os.ReadFile(filepath.Join(*ext, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
		if err != nil {
			return err
		}
		if len(data) > 4 && string(data[:4]) == "LZ77" {
			data = nds.Decompress(data[4:])
		}
		k, err := sm64ds.ParseKCL(data)
		if err != nil {
			return err
		}
		if bad := checkCorners(k); bad > 0 {
			fmt.Fprintf(os.Stderr, "  %s: %d prisms whose centroid FAILS the walker's own tests\n", stem, bad)
		}
		tris, skipped := trisOf(k)
		if len(tris) == 0 {
			return fmt.Errorf("%s: no triangles", stem)
		}
		glb, err := nitro.ExportTrisGLB(stem+"_col", map[int][]nitro.Tri{0: tris},
			[]nitro.Material{{Name: "collision", Alpha: 31}}, nil)
		if err != nil {
			return err
		}
		if skipped > 0 {
			fmt.Printf("  %s: %d prisms, %d degenerate skipped\n", stem, k.NumPrisms()-1, skipped)
		}
		return os.WriteFile(filepath.Join(*outDir, stem+".glb"), glb, 0o644)
	}

	// every level's collision map
	done := map[string]bool{}
	levels := 0
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl")
		if stem == "" || done[stem] {
			continue
		}
		done[stem] = true
		if err := export(stem); err != nil {
			fmt.Fprintf(os.Stderr, "  level %d %s: %v\n", i, stem, err)
			continue
		}
		levels++
	}

	// every object collider an actor's own code loaded (oracle bindings)
	var bindings map[int][]struct {
		KCL []string `json:"kcl"`
	}
	objects := 0
	if buf, err := os.ReadFile(*bindPath); err == nil {
		json.Unmarshal(buf, &bindings)
		for _, bs := range bindings {
			for _, b := range bs {
				for _, stem := range b.KCL {
					if done[stem] {
						continue
					}
					done[stem] = true
					if err := export(stem); err != nil {
						fmt.Fprintf(os.Stderr, "  object %s: %v\n", stem, err)
						continue
					}
					objects++
				}
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "no binding table (%v): level maps only\n", err)
	}
	fmt.Printf("exported %d level + %d object collision meshes to %s\n", levels, objects, *outDir)
}

// checkCorners validates the reconstruction against the walker's own accept
// region: every triangle centroid must lie in the face plane, inside edges A
// and B (dot ≤ 0) and between 0 and length along edge C — the exact
// inequalities of $01FFD3F8, evaluated at the reconstruction's units.
func checkCorners(k *sm64ds.KCL) (bad int) {
	for i := 1; i < k.NumPrisms(); i++ {
		c, ok := k.Corners(i)
		if !ok {
			continue
		}
		p := k.PrismAt(i)
		v := k.VertexAt(p.Vertex)
		var d [3]float64 // centroid - anchor, back in world>>6 units
		for j := 0; j < 3; j++ {
			d[j] = ((c[0][j]+c[1][j]+c[2][j])/3 - float64(v[j])) / 64
		}
		dot := func(ni int) float64 {
			n := k.NormalAt(ni)
			return (d[0]*float64(n[0]) + d[1]*float64(n[1]) + d[2]*float64(n[2])) / 4
			// NormalAt is <<2; /4 restores the raw fx.10 the walker multiplies with
		}
		const tol = 0x20000 // the walker's edge tolerance
		if math.Abs(dot(p.FaceNormal)) > tol ||
			dot(p.EdgeA) > tol || dot(p.EdgeB) > tol ||
			dot(p.EdgeC) < -tol || dot(p.EdgeC) > float64(p.Length)+tol {
			bad++
		}
	}
	return bad
}

// trisOf reconstructs all prisms as viewer triangles: stage-GLB units, a
// normal-shaded red baked as vertex color (top-lit so slopes read).
func trisOf(k *sm64ds.KCL) (tris []nitro.Tri, skipped int) {
	const lx, ly, lz = 0.30, 0.90, 0.30 // light direction (normalized-ish)
	for i := 1; i < k.NumPrisms(); i++ {
		c, ok := k.Corners(i)
		if !ok {
			skipped++
			continue
		}
		n := k.NormalAt(k.PrismAt(i).FaceNormal)
		nl := math.Sqrt(float64(n[0])*float64(n[0]) + float64(n[1])*float64(n[1]) + float64(n[2])*float64(n[2]))
		shade := 0.55
		if nl > 0 {
			d := (float64(n[0])*lx + float64(n[1])*ly + float64(n[2])*lz) / nl
			if d > 0 {
				shade += 0.45 * d
			}
		}
		col := color.NRGBA{R: uint8(210 * shade), G: uint8(38 * shade), B: uint8(38 * shade), A: 255}
		var t nitro.Tri
		bad := false
		for j := 0; j < 3; j++ {
			x, y, z := c[j][0]*toGLB, c[j][1]*toGLB, c[j][2]*toGLB
			if math.Abs(x) > 1e4 || math.Abs(y) > 1e4 || math.Abs(z) > 1e4 {
				bad = true // near-degenerate prism blew a corner out of the area
				break
			}
			t.V[j] = nitro.Vertex{X: x, Y: y, Z: z, C: col}
		}
		if bad {
			skipped++
			continue
		}
		tris = append(tris, t)
	}
	return tris, skipped
}
