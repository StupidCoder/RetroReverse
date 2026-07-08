package main

// objects.go exports the roadside scenery as separate GLB models plus a
// placement manifest. Each distinct OBJ.RRO object referenced by the placement
// tables (rr.Placements) becomes one models/obj-NN.glb — its local geometry,
// textured with the scenery set its placements fall in (no object id spans two
// sets). levels/course.objects.json then lists every placement as a
// {model, pos, rot} the viewer's object layer loads, carrying the source
// record's address, flag and yaw for the per-object info card.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

// ObjectPlacement is one entry of course.objects.json.
type ObjectPlacement struct {
	Model string     `json:"model"`
	Pos   [3]float64 `json:"pos"`
	Rot   [3]float64 `json:"rot"` // radians; only Y is nonzero
	// Diagnostics surfaced in the object info card.
	ID    int    `json:"id"`
	Addr  string `json:"addr"`
	Flags string `json:"flags"`
	Yaw   int    `json:"yaw"` // raw units, 4096 = 360°
}

type objectsDoc struct {
	Objects []ObjectPlacement `json:"objects"`
}

// exportObjects writes one GLB per referenced object and course.objects.json,
// returning the manifest-relative path of the placement file.
func exportObjects(a *assets, out string, cps []rr.Checkpoint) (string, error) {
	placements := rr.Placements(a.exe)

	// The scenery set each object id renders under (its placements never span
	// two sets, verified). Build one GLB per used id.
	setOf := map[int]int{}
	for _, p := range placements {
		if p.Obj < 0 || p.Obj >= len(a.objs) {
			continue
		}
		seg := rr.NearestSegment(cps, p.X, p.Z)
		setOf[p.Obj] = rr.SetForProgress(int32(seg) * 256)
	}
	ids := make([]int, 0, len(setOf))
	for id := range setOf {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	if err := os.MkdirAll(filepath.Join(out, "models"), 0o755); err != nil {
		return "", err
	}
	identity := [3][3]int32{{4096, 0, 0}, {0, 4096, 0}, {0, 0, 4096}}
	for _, id := range ids {
		b := newMeshBuilder(a.vrams[setOf[id]])
		addObjectXform(b, &a.objs[id], identity, [3]int32{0, 0, 0}, setOf[id])
		if err := b.Write(filepath.Join(out, fmt.Sprintf("models/obj-%02d.glb", id))); err != nil {
			return "", err
		}
	}

	var doc objectsDoc
	for _, p := range placements {
		if p.Obj < 0 || p.Obj >= len(a.objs) {
			continue
		}
		doc.Objects = append(doc.Objects, ObjectPlacement{
			Model: fmt.Sprintf("models/obj-%02d.glb", p.Obj),
			Pos:   [3]float64{float64(p.X) / 1024, -float64(p.Y) / 1024, -float64(p.Z) / 1024},
			Rot:   [3]float64{0, float64(p.Yaw) / 4096 * 2 * math.Pi, 0},
			ID:    p.Obj,
			Addr:  fmt.Sprintf("0x%08X", p.Addr),
			Flags: fmt.Sprintf("0x%08X", p.Flag),
			Yaw:   int(p.Yaw),
		})
	}
	if err := os.MkdirAll(filepath.Join(out, "levels"), 0o755); err != nil {
		return "", err
	}
	rel := "levels/course.objects.json"
	j, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(filepath.Join(out, rel), append(j, '\n'), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[levels] %d object models, %d placements\n", len(ids), len(doc.Objects))
	return rel, nil
}
