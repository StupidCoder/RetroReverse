// webexport builds Elite's Retro-X game tree from the extracted tape files
// (extract -o work/extracted Elite.tap):
//
//	objects/<ship>.json     every XX21 ship blueprint as a wireframe3d object —
//	                        vertices, edges with their two adjacent faces, face
//	                        normals and face centroids, so the viewer reproduces
//	                        Elite's own back-face-culled hidden-surface wireframe
//	                        (Elite.md Part IV §1)
//	pictures/loading.png    the multicolor-bitmap loading picture
//	music/docking.mp3       the docking theme (The Blue Danube), synthesised from
//	                        the game image by the $BDDC sequencer reimplementation
//
// Ship type -> name is the documented Commodore 64 Elite blueprint table (XX21);
// the names live only in the manual, never in the program (Elite.md Part V §1),
// anchored by our own write-up's Thargoid = type $1D = 29.
//
// Usage (from games/elite-c64/): go run ./extract/cmd/webexport -in work/extracted
package main

import (
	"fmt"
	"math"

	"retroreverse.com/games/elite-c64/extract/shipmodel"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

func main() {
	cli.Main("elite-c64", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in work/extracted [-o DIR] [-only objects,pictures,music]")
	}

	b := ctx.Builder
	b.SetTitle("Elite")
	b.SetPlatform("Commodore 64")
	b.SetYear(1985)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 200},
		TickHz: 50,
		Filter: "crt",
	})

	if ctx.Stage("objects") {
		if err := exportShips(ctx); err != nil {
			return err
		}
	}
	if ctx.Stage("pictures") {
		if err := exportLoadingScreen(ctx); err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportDockingMusic(ctx); err != nil {
			return err
		}
	}
	return nil
}

// shipNames maps each XX21 blueprint slot to the ship every Elite player knows.
// Source: the Commodore 64 Elite ship table. Slots whose blueprint does not decode
// (e.g. 8 Splinter) simply never appear in the output.
var shipNames = map[int]string{
	1:  "Missile",
	2:  "Coriolis space station",
	3:  "Escape pod",
	4:  "Alloy plate",
	5:  "Cargo canister",
	6:  "Boulder",
	7:  "Asteroid",
	8:  "Splinter",
	9:  "Shuttle",
	10: "Transporter",
	11: "Cobra Mk III",
	12: "Python",
	13: "Boa",
	14: "Anaconda",
	15: "Rock hermit",
	16: "Viper",
	17: "Sidewinder",
	18: "Mamba",
	19: "Krait",
	20: "Adder",
	21: "Gecko",
	22: "Cobra Mk I",
	23: "Worm",
	24: "Cobra Mk III (pirate)",
	25: "Asp Mk II",
	26: "Python (pirate)",
	27: "Fer-de-Lance",
	28: "Moray",
	29: "Thargoid",
	30: "Thargon",
	31: "Constrictor",
	32: "Cougar",
	33: "Dodo space station",
}

// faceCentroid returns a representative point on face f — the average of its
// candidate vertices (the union of the edge endpoints that border f and the
// vertices whose own face list names f; either source alone is incomplete). For
// the back-face test only the point's position relative to the face plane
// matters, and the centroid sits squarely on (planar) or within (non-planar) the
// face. Faces named by nothing (the all-$F alloy plate) return the origin; their
// edges are drawn unconditionally anyway.
func faceCentroid(s *shipmodel.Ship, f int) [3]float64 {
	seen := map[int]bool{}
	var sx, sy, sz, n int
	add := func(v int) {
		if !seen[v] {
			seen[v] = true
			sx += s.Vertices[v].X
			sy += s.Vertices[v].Y
			sz += s.Vertices[v].Z
			n++
		}
	}
	for _, e := range s.Edges {
		if e.FaceA == f || e.FaceB == f {
			add(e.V1)
			add(e.V2)
		}
	}
	for v, vert := range s.Vertices {
		for _, vf := range vert.Faces {
			if vf == f {
				add(v)
				break
			}
		}
	}
	if n == 0 {
		return [3]float64{0, 0, 0}
	}
	round := func(t int) float64 { return math.Round(float64(t) / float64(n)) }
	return [3]float64{round(sx), round(sy), round(sz)}
}

// buildWireframe decodes one Ship into the wireframe3d payload the viewer's
// hidden-surface line renderer consumes, plus its bounding radius.
func buildWireframe(s *shipmodel.Ship) (*schema.Wireframe, float64) {
	wf := &schema.Wireframe{}
	var r float64
	for _, v := range s.Vertices {
		wf.Positions = append(wf.Positions, float64(v.X), float64(v.Y), float64(v.Z))
		if d := math.Hypot(math.Hypot(float64(v.X), float64(v.Y)), float64(v.Z)); d > r {
			r = d
		}
	}
	face := func(f int) int {
		if f >= len(s.Faces) { // the blueprint's $F sentinel: no face on this side
			return -1
		}
		return f
	}
	for _, e := range s.Edges {
		wf.Edges = append(wf.Edges, []int{e.V1, e.V2, face(e.FaceA), face(e.FaceB)})
	}
	for fi, f := range s.Faces {
		wf.Faces = append(wf.Faces, []float64{float64(f.NX), float64(f.NY), float64(f.NZ)})
		c := faceCentroid(s, fi)
		wf.FaceCenters = append(wf.FaceCenters, []float64{c[0], c[1], c[2]})
	}
	return wf, r
}
