package main

// objects.go exports the roadside scenery as separate GLB models plus a
// placement manifest, Ridge Racer style: each RoadObjects def referenced by a
// placement becomes one models/obj-NN.glb (billboards as textured quads, 3D
// defs as their ORI3 mesh), and <course>.objects.json lists every placement
// as {model, pos, rot} with the source record's diagnostics for the info
// card. Repeated props ship once and the viewer instances them.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/threedo"
)

// ObjectPlacement is one entry of <course>.objects.json.
type ObjectPlacement struct {
	Model string     `json:"model"`
	Pos   [3]float64 `json:"pos"`
	Rot   [3]float64 `json:"rot"` // radians; only Y is nonzero
	// Diagnostics surfaced in the object info card.
	Def     int    `json:"def"`
	Type    int    `json:"type"`
	Segment int    `json:"segment"`
	Index   int    `json:"index"` // placement record index
	RawYaw  int    `json:"rawYaw"`
	Flags   string `json:"flags"`
}

type objectsDoc struct {
	Objects []ObjectPlacement `json:"objects"`
}

func exportObjects(a *assets, out string) (string, error) {
	// which defs are actually placed?
	used := map[int]bool{}
	for _, p := range a.track.Objects.Placements {
		used[int(p.Def)] = true
	}
	var ids []int
	for id := range used {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	written := map[int]string{}
	for _, id := range ids {
		def := &a.track.Objects.Defs[id]
		file := fmt.Sprintf("models/obj-%02d.glb", id)
		var err error
		switch def.Type {
		case 1:
			err = writeModelObject(a, def, filepath.Join(out, file))
		default: // 4 = upright billboard, 6 = two-anchor cel (approximated flat)
			err = writeBillboard(a, def, filepath.Join(out, file))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "[objects] obj-%02d (type %d): %v — skipped\n", id, def.Type, err)
			continue
		}
		written[id] = file
	}

	var doc objectsDoc
	for i, p := range a.track.Objects.Placements {
		file, ok := written[int(p.Def)]
		if !ok || int(p.Segment) >= len(a.track.Segments) {
			continue
		}
		seg := &a.track.Segments[p.Segment]
		w := p.World(seg)
		pos := gl(w)
		def := &a.track.Objects.Defs[p.Def]
		doc.Objects = append(doc.Objects, ObjectPlacement{
			Model: file,
			Pos:   [3]float64{float64(pos[0]), float64(pos[1]), float64(pos[2])},
			Rot:   [3]float64{0, objectYaw(p, seg), 0},
			Def:   int(p.Def), Type: int(def.Type), Segment: int(p.Segment), Index: i,
			RawYaw: int(p.Yaw), Flags: fmt.Sprintf("% x", p.Raw[:]),
		})
	}
	rel := a.course + ".objects.json"
	j, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(filepath.Join(out, rel), append(j, '\n'), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[objects] %d models, %d placements -> %s\n", len(written), len(doc.Objects), rel)
	return rel, nil
}

// objectYaw converts a placement's yaw to GLB radians. The billboard corner
// code (0x161E4) rotates by a = -(yaw<<16) - (heading<<10), full circle
// 0x1000000, and spans the quad along (cos a, sin a) in game X/Z. Mapping to
// GLB (x, y, -z) and three.js's +Y rotation of the +X-aligned quad gives
// theta = a — the same negative angle. (The sign error this replaces was
// invisible at the start grid, where the heading is 0, and mirrored every
// sign and lamp post once the track curved.)
func objectYaw(p nfs.Placement, seg *nfs.Segment) float64 {
	return -float64(int32(p.Yaw)<<16+int32(seg.Heading)<<10) / float64(1<<24) * 2 * math.Pi
}

// writeBillboard emits one textured quad: width W1 centred on the origin,
// height H upward (or depth H flat on the ground for Flat defs).
func writeBillboard(a *assets, def *nfs.ObjectDef, path string) error {
	img, err := a.tex.objectTexture(int(def.Tex))
	if err != nil {
		return err
	}
	w := float32(nfs.Float(def.W1))
	h := float32(nfs.Float(def.H))
	var positions [][3]float32
	if def.Flat {
		positions = [][3]float32{
			{-w / 2, 0.01, -h / 2}, {w / 2, 0.01, -h / 2}, {w / 2, 0.01, h / 2}, {-w / 2, 0.01, h / 2},
		}
	} else {
		positions = [][3]float32{
			{-w / 2, h, 0}, {w / 2, h, 0}, {w / 2, 0, 0}, {-w / 2, 0, 0},
		}
	}
	uvs := [][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	groups := []glb.TexturedGroup{{
		Tris:  [][3]uint32{{0, 1, 2}, {0, 2, 3}},
		Image: img,
	}}
	return glb.WriteTextured(path, positions, uvs, groups, nil)
}

// writeModelObject emits a type-1 def's ORI3 mesh. The packet's model group
// (root child 2) holds (model, shape) pairs exactly like the car families;
// the def's texture byte picks the pair.
func writeModelObject(a *assets, def *nfs.ObjectDef, path string) error {
	models := a.root.Children[2]
	if models == nil || len(models.Children) == 0 || models.Children[0] == nil {
		return fmt.Errorf("packet has no model group")
	}
	list := models.Children[0]
	idx := int(def.Tex)
	if idx >= len(list.Children) {
		return fmt.Errorf("model index %d out of range (%d)", idx, len(list.Children))
	}
	pair := list.Children[idx]
	if pair == nil || len(pair.Children) < 2 {
		return fmt.Errorf("model pair %d missing", idx)
	}
	m, s := pair.Children[0], pair.Children[1]
	parsed, err := threedo.ParseModels(a.pkt[m.Offset:])
	if err != nil || len(parsed) == 0 {
		return fmt.Errorf("bad ORI3: %v", err)
	}
	lod := &nfs.CarLOD{Model: parsed[0]}
	if err := nfs.ParseShapeInto(a.pkt[s.Offset:], lod); err != nil {
		return err
	}
	return writeORI3(lod, 1.0/128, path)
}
