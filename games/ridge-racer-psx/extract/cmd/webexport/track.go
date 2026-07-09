package main

// track.go exports the whole course as one textured GLB. IDX.HED's 32×32 grid
// places each MAP.RRM section at its cell origin (cellX*8192, 0, cellZ*8192)
// in the records' model units; a section's three record classes are all
// 40-byte textured quads in cell space, so the world mesh is a straight
// composition — the same placement the renderer's grid walk performs each
// frame.

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

func exportTrack(a *assets, out string) (ModelIndex, error) {
	dir := filepath.Join(out, "models")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ModelIndex{}, err
	}
	cps, err := rr.Checkpoints(a.exe)
	if err != nil {
		return ModelIndex{}, err
	}
	b := newMeshBuilder(a.vrams[0], a.vrams[1], a.vrams[2])
	cells := 0
	for z := 0; z < 32; z++ {
		for x := 0; x < 32; x++ {
			sec := a.grid.Section(x, z)
			if sec == rr.Empty || int(sec) >= len(a.track.Sections) {
				continue
			}
			cells++
			off := [3]int32{int32(x) * rr.CellModel, 0, int32(z) * rr.CellModel}
			// The scenery set this section renders under: the set active when
			// the car passes its cell (course.go).
			seg := rr.NearestSegment(cps, off[0]+rr.CellModel/2, off[2]+rr.CellModel/2)
			set := rr.SetForProgress(int32(seg) * 256)
			s := &a.track.Sections[sec]
			for _, class := range [][]rr.TrackQuad{s.A, s.B, s.C} {
				for _, q := range class {
					b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off, set, rr.TexWindow{})
				}
			}
		}
	}

	file := "models/track.glb"
	if err := b.Write(filepath.Join(out, file)); err != nil {
		return ModelIndex{}, err
	}
	fmt.Fprintf(os.Stderr, "[levels] track.glb (%d cells, %d verts, %d tiles)\n",
		cells, len(b.verts), len(b.tiles))

	// The roadside objects ship as separate GLBs placed by course.objects.json,
	// so the viewer can toggle and inspect them.
	objectsFile, err := exportObjects(a, out, cps)
	if err != nil {
		return ModelIndex{}, err
	}
	return ModelIndex{
		Name: "Ridge Racer Course", File: file, Kind: "rr-course",
		ObjectsFile: objectsFile, Fly: true,
	}, nil
}

// rotv applies a 4096-scaled rotation matrix to an int16 vertex, returning an
// int16 vertex (object-local magnitudes stay in range).
func rotv(R [3][3]int32, v [3]int16) [3]int16 {
	var o [3]int16
	for i := 0; i < 3; i++ {
		s := int32(v[0])*R[i][0] + int32(v[1])*R[i][1] + int32(v[2])*R[i][2]
		o[i] = int16(s / 4096)
	}
	return o
}

// addObjectXform adds one OBJ.RRO object rotated by R and translated by off,
// sampling scenery set `set` for its quadrant-page textures.
func addObjectXform(b *meshBuilder, o *rr.Object, R [3][3]int32, off [3]int32, set int) {
	rot := func(v [4][3]int16) [4][3]int16 {
		var r [4][3]int16
		for i := range v {
			r[i] = rotv(R, v[i])
		}
		return r
	}
	for _, q := range o.FT {
		b.AddTextured(rot(q.V), q.UV, q.TPage, q.CLUT, off, set, rr.TexWindow{})
	}
	for _, q := range o.FT8 {
		b.AddTextured(rot(q.V), q.UV, q.TPage, q.CLUT, off, set, q.Window())
	}
	for _, q := range o.F {
		b.AddFlat(rot(q.V), q.RGB, off)
	}
	for _, q := range o.GT {
		b.AddTextured(rot(q.V), q.UV, q.TPage, q.CLUT, off, set, rr.TexWindow{})
	}
	for _, q := range o.GT8 {
		b.AddTextured(rot(q.V), q.UV, q.TPage, q.CLUT, off, set, rr.TexWindow{})
	}
	for _, q := range o.G {
		b.AddFlat(rot(q.V), q.RGB, off)
	}
}
