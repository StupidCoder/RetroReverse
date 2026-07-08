// models.go is the models stage: it decodes every Elite ship blueprint and writes
// each ship BOTH as a GLB LINES wireframe (models/<ship>.glb, loadable in any glTF
// viewer) and as its HSR blueprint (models/<ship>.hsr.json, consumed by the site's
// bespoke back-face-culling line renderer). This is the reconstruction that used
// to write the single ships.json; the per-ship split lets the manifest point a
// standard model file and a sibling data file at each ship.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/elite-c64/extract/shipmodel"
	"retroreverse.com/tools/lib/glb"
)

// wireColor is the light base colour of the exported GLB wireframes (the site's
// live render recolours per its own material; this is just what a neutral glTF
// viewer shows).
var wireColor = [3]float32{0.82, 0.90, 1.0}

func exportModels(inDir, outDir string) []ModelIndex {
	mem, err := shipmodel.LoadEngine(inDir)
	chk(err)
	ships := shipmodel.ParseAll(mem)

	modelsDir := filepath.Join(outDir, "models")
	chk(os.MkdirAll(modelsDir, 0o755))

	var idx []ModelIndex
	for i, s := range ships {
		js := buildShip(s)
		stem := slug(js.Name)

		// GLB wireframe: positions = ship vertices, edges = the ship's edge list.
		positions := make([][3]float32, len(js.Verts))
		for vi, v := range js.Verts {
			positions[vi] = [3]float32{float32(v[0]), float32(v[1]), float32(v[2])}
		}
		edges := make([][2]uint32, len(js.Edges))
		for ei, e := range js.Edges {
			edges[ei] = [2]uint32{uint32(e[0]), uint32(e[1])}
		}
		glbFile := filepath.Join("models", stem+".glb")
		chk(glb.WriteLines(filepath.Join(outDir, glbFile), positions, edges, wireColor))

		// HSR blueprint: the exact per-ship object the viewer's ShipMesh consumes.
		hsrFile := filepath.Join("models", stem+".hsr.json")
		writeJSON(filepath.Join(outDir, hsrFile), js)

		idx = append(idx, ModelIndex{
			Name: js.Name, Section: "Ships", Kind: "elite-ship",
			File: filepath.ToSlash(glbFile), Data: filepath.ToSlash(hsrFile),
		})
		fmt.Fprintf(os.Stderr, "[models] %2d/%2d  %-24s %2d verts %2d edges %2d faces\n",
			i+1, len(ships), stem, len(js.Verts), len(js.Edges), len(js.Faces))
	}
	fmt.Fprintf(os.Stderr, "[models] done: %d ships\n", len(idx))
	return idx
}
