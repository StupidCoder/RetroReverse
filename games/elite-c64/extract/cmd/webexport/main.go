// webexport reconstructs Elite's web assets from the extracted C64 game image and
// writes the common format-2 asset tree (see FORMAT2.md / STANDARDS.md §4)
// under the output root:
//
//	manifest.json               game index: native res, tick rate, model/bitmap/music list
//	models/<ship>.glb           per ship: the wireframe as GLB LINES (tools/lib/glb)
//	models/<ship>.hsr.json      per ship: the hidden-surface-removal blueprint the three.js
//	                            viewer's ShipMesh consumes (verts, edges+adjacent faces,
//	                            face normals, face centroids) — Elite.md Part IV §1
//	bitmaps/loading.png         the multicolor-bitmap loading picture
//	music/docking.mp3           the docking theme, synthesised from the game image
//
// Every ship is emitted TWICE: as a standard GLB wireframe (loads in any glTF
// viewer) and as the raw HSR blob (drives the site's authentic back-face-culled,
// flickering line render). The three producers (models, bitmaps, music) are folded
// into this one command; -only gates which stages run.
//
// Ship type -> name is the documented Commodore 64 Elite blueprint table (XX21);
// the names live only in the manual, never in the program (Elite.md Part V §1),
// anchored by our own write-up's Thargoid = type $1D = 29.
//
// Usage: webexport [-in <extracted dir>] [-o <outdir>] [-only models,bitmaps,music,all]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/elite-c64/extract/shipmodel"
)

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

// jsonShip is one ship's HSR blueprint — the exact per-ship object the three.js
// viewer's ShipMesh consumes. It carries everything the viewer needs to reproduce
// Elite's own wireframe hidden-surface removal: each face's outward normal and a
// point on it (faceC), and each edge with the two faces beside it. The viewer
// back-face tests every face by the sign of dot(normal, eye - faceCenter) — the
// face is visible when it points toward the eye from where the face actually sits
// — and draws an edge only when one of its faces is visible (Elite.md Part IV §1).
// Face nibble $F (15) on an edge means "no face here"; such an edge is always drawn.
//
// This is written per ship as models/<ship>.hsr.json (the field names/shape are
// unchanged from the old single ships.json, so the render math is identical).
type jsonShip struct {
	Type   int      `json:"type"`
	Name   string   `json:"name"`
	Radius float64  `json:"radius"`
	Verts  [][3]int `json:"verts"`
	Edges  [][4]int `json:"edges"` // v1, v2, faceA, faceB
	Faces  [][3]int `json:"faces"` // outward normal nx, ny, nz
	FaceC  [][3]int `json:"faceC"` // a point on each face (its vertex centroid)
}

// Manifest is the format-2 game index (STANDARDS.md §4.2). Sections gated out by
// -only stay empty and omitempty drops them, so a partial run is self-consistent.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Models   []ModelIndex   `json:"models,omitempty"`
	Bitmaps  []AssetIndex   `json:"bitmaps,omitempty"`
	Music    []AssetIndex   `json:"music,omitempty"`
}

// ModelIndex is one manifest models[] entry: the GLB wireframe plus the sibling
// HSR blob the bespoke elite-ship viewer reads.
type ModelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"` // models/<ship>.glb
	Kind    string `json:"kind"` // "elite-ship"
	Data    string `json:"data"` // models/<ship>.hsr.json
}

// AssetIndex is one bitmaps[]/music[] entry.
type AssetIndex struct {
	Name string `json:"name"`
	File string `json:"file"`
}

// faceCentroid returns a representative point on face f — the average of its
// candidate vertices (the union of the edge endpoints that border f and the
// vertices whose own face list names f; either source alone is incomplete). For
// the back-face test only the point's position relative to the face plane
// matters, and the centroid sits squarely on (planar) or within (non-planar) the
// face. Faces named by nothing (the all-$F alloy plate) return the origin; their
// edges are drawn unconditionally anyway.
func faceCentroid(s *shipmodel.Ship, f int) [3]int {
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
		return [3]int{0, 0, 0}
	}
	round := func(t int) int { return int(math.Round(float64(t) / float64(n))) }
	return [3]int{round(sx), round(sy), round(sz)}
}

// buildShip decodes one Ship into the jsonShip HSR blueprint (and computes the
// bounding radius the viewer frames the camera by).
func buildShip(s *shipmodel.Ship) jsonShip {
	js := jsonShip{Type: s.Type, Name: shipNames[s.Type]}
	if js.Name == "" {
		js.Name = fmt.Sprintf("Ship %d", s.Type)
	}
	var r float64
	for _, v := range s.Vertices {
		js.Verts = append(js.Verts, [3]int{v.X, v.Y, v.Z})
		if d := math.Hypot(math.Hypot(float64(v.X), float64(v.Y)), float64(v.Z)); d > r {
			r = d
		}
	}
	js.Radius = r
	for _, e := range s.Edges {
		js.Edges = append(js.Edges, [4]int{e.V1, e.V2, e.FaceA, e.FaceB})
	}
	for fi, f := range s.Faces {
		js.Faces = append(js.Faces, [3]int{f.NX, f.NY, f.NZ})
		js.FaceC = append(js.FaceC, faceCentroid(s, fi))
	}
	return js
}

// slug turns a ship name into a stable kebab-case file stem ("Cobra Mk III
// (pirate)" -> "cobra-mk-iii-pirate"). Names are unique across the XX21 table, so
// the slug is a stable per-ship identity.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects
// every stage; otherwise only the named stages (models,bitmaps,music) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == "all" {
			out["models"], out["bitmaps"], out["music"] = true, true, true
			continue
		}
		if s != "models" && s != "bitmaps" && s != "music" {
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want models,bitmaps,music,all)\n", s)
			os.Exit(2)
		}
		out[s] = true
	}
	return out
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func writeJSON(path string, v any) {
	b, err := json.Marshal(v)
	chk(err)
	chk(os.WriteFile(path, b, 0o644))
}

func main() {
	in := flag.String("in", "../extracted", "input directory of extracted C64 files")
	outdir := flag.String("o", "../../site/public/elite-c64", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages: models,bitmaps,music,all")
	flag.Parse()
	sel := parseOnly(*only)
	chk(os.MkdirAll(*outdir, 0o755))

	man := Manifest{
		Format: 2, Game: "elite-c64", Platform: "Commodore 64",
		Native: map[string]int{"w": 320, "h": 200}, TickHz: 50,
	}
	if sel["models"] {
		man.Models = exportModels(*in, *outdir)
	}
	if sel["bitmaps"] {
		man.Bitmaps = exportBitmaps(*in, *outdir)
	}
	if sel["music"] {
		man.Music = exportMusic(*in, *outdir)
	}
	writeJSON(filepath.Join(*outdir, "manifest.json"), man)
	fmt.Fprintf(os.Stderr, "[manifest] %d models, %d bitmaps, %d music -> %s\n",
		len(man.Models), len(man.Bitmaps), len(man.Music), *outdir)
}
