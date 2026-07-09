package main

// specials.go exports the course's dynamic objects — everything placed by
// dedicated code rather than the placement tables: the number girl, the
// helicopter and airplane, the start balloons and lap banner, the big screens,
// the start banner, the swinging tunnel sign, the rotating sign, the beacon,
// the start-gate crowd and semaphore, the camera crane and the night hot-air
// balloons. The catalog comes from rr.Dynamics() (traced code blocks + EXE
// position vectors; see rr/dynamics.go). Each entry becomes one GLB under the
// Studio's "Objects" section — the base object plus, for the helicopter, its
// rotor — with the drawer's upright pitch baked into the geometry. Entries
// with a fixed world position are also appended to course.objects.json so the
// course view shows them in place; the airplane has no fixed position (its
// path is integrated at runtime) and is exported as a model only.

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

// specialPlacements is filled by exportSpecials and merged into
// course.objects.json by exportObjects (which runs in the levels stage).
func dynamicPlacements(a *assets, cps []rr.Checkpoint) []ObjectPlacement {
	var out []ObjectPlacement
	for _, d := range rr.Dynamics(a.exe) {
		if d.X == 0 && d.Z == 0 {
			continue // no fixed world position (the airplane)
		}
		flags := "dynamic"
		if d.Night {
			flags = "dynamic, night"
		}
		out = append(out, ObjectPlacement{
			Model: fmt.Sprintf("models/special-%03d.glb", d.Objs[0]),
			Pos:   [3]float64{float64(d.X) / 1024, -float64(d.Y) / 1024, -float64(d.Z) / 1024},
			Rot:   [3]float64{0, float64(d.Yaw) / 4096 * 2 * 3.141592653589793, 0},
			ID:    d.Objs[0],
			Addr:  fmt.Sprintf("0x%08X", d.Addr),
			Flags: flags,
			Yaw:   int(d.Yaw),
		})
	}
	return out
}

func exportSpecials(a *assets, out string) ([]ModelIndex, error) {
	cps, err := rr.Checkpoints(a.exe)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(out, "models")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var models []ModelIndex
	seen := map[int]bool{}
	for _, d := range rr.Dynamics(a.exe) {
		if seen[d.Objs[0]] {
			continue // both big screens share one model
		}
		seen[d.Objs[0]] = true
		// Texture state: the scenery set at the entry's position.
		set := 0
		if d.X != 0 || d.Z != 0 {
			seg := rr.NearestSegment(cps, d.X, d.Z)
			set = rr.SetForProgress(int32(seg) * 256)
		}
		R := rr.YawMatrix(0)
		if d.Pitch != 0 {
			R = rr.PitchMatrix(d.Pitch)
		}
		// Multi-part entries assemble every listed object (joints at zero);
		// everywhere else the extra ids are animation frames or state
		// variants, so only the base is meshed.
		ids := d.Objs[:1]
		if d.Name == "Helicopter" || d.Name == "Camera crane" {
			ids = d.Objs
		}
		b := newMeshBuilder(a.vrams[set])
		for _, id := range ids {
			if id < 0 || id >= len(a.objs) {
				continue
			}
			addObjectXform(b, &a.objs[id], R, [3]int32{0, 0, 0}, set)
		}
		file := fmt.Sprintf("models/special-%03d.glb", d.Objs[0])
		if err := b.Write(filepath.Join(out, file)); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[objects] %s (obj %v, set %d, %d verts)\n",
			d.Name, d.Objs, set, len(b.verts))
		models = append(models, ModelIndex{
			Name:    fmt.Sprintf("%s (obj %d)", d.Name, d.Objs[0]),
			File:    file,
			Kind:    "mesh3d",
			Section: "Objects",
		})
	}
	return models, nil
}
