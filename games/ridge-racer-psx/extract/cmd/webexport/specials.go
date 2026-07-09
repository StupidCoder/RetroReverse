package main

// specials.go exports the course's bespoke / animated objects as standalone,
// inspectable GLBs under the Studio's "Objects" section. These are drawn by
// dedicated code (a helicopter that flies a fixed-altitude path, a 2-D
// billboard sprite, trackside screens and banners) rather than the static
// placement tables (objects.go), so they carry no course.objects.json entry and
// no fixed world transform. Each is exported at its own origin — the viewer
// shows the raw model. The recovered world position is used only to pick the
// scenery set its quadrant-page textures sample; a zero position falls back to
// set 0.
//
// The ids and positions were found by capturing every object the running game
// draws over an attract lap that is not in the placement tables, then anchoring
// each to a known-placement object drawn in the same frame (cam-relative world
// position via Rᵀ·TR). 175 flies at a constant altitude (helicopter); 250 has
// no stable world position (a screen-space 2-D billboard — the number girl);
// 182/185 are the finish banner and a trackside screen.

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

// specialObj is one bespoke object: its OBJ.RRO id, a display name, and the
// world position (model units) recovered for it — used only to select the
// scenery set (0,0 = unknown → set 0).
type specialObj struct {
	ID   int
	Name string
	X, Z int32
}

// These are real OBJ.RRO objects the game draws in the race that are not in the
// static placement tables (objects.go). Only 182 is identified — the finish
// banner. The rest are miscellaneous course objects, kept under neutral names;
// they are NOT the animated helicopter / number girl / airplane / big screen
// (those are drawn by dedicated code from other geometry/sprite sources and are
// not yet located).
var specials = []specialObj{
	{182, "Finish banner", 158750, 150511},
	{175, "Object 175", 177237, 140359},
	{185, "Object 185", 98133, 68082},
	{250, "Object 250", 134180, 158126},
	{257, "Object 257", 0, 0},
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
	identity := [3][3]int32{{4096, 0, 0}, {0, 4096, 0}, {0, 0, 4096}}
	var models []ModelIndex
	for _, s := range specials {
		if s.ID < 0 || s.ID >= len(a.objs) {
			fmt.Fprintf(os.Stderr, "[objects] skip special %d (out of range)\n", s.ID)
			continue
		}
		set := 0
		if s.X != 0 || s.Z != 0 {
			seg := rr.NearestSegment(cps, s.X, s.Z)
			set = rr.SetForProgress(int32(seg) * 256)
		}
		b := newMeshBuilder(a.vrams[set])
		addObjectXform(b, &a.objs[s.ID], identity, [3]int32{0, 0, 0}, set)
		file := fmt.Sprintf("models/special-%03d.glb", s.ID)
		if err := b.Write(filepath.Join(out, file)); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[objects] special %s (obj %d, set %d, %d verts)\n",
			s.Name, s.ID, set, len(b.verts))
		name := s.Name
		if s.Name == fmt.Sprintf("Object %d", s.ID) {
			name = s.Name // already carries the id
		} else {
			name = fmt.Sprintf("%s (obj %d)", s.Name, s.ID)
		}
		models = append(models, ModelIndex{
			Name: name, File: file, Kind: "mesh3d", Section: "Objects",
		})
	}
	return models, nil
}
