package main

// track.go exports the whole course as one textured GLB. IDX.HED's 32×32 grid
// places each MAP.RRM section at its cell origin (cellX*2048, 0, cellZ*2048);
// a section's three record classes are all 40-byte textured quads in cell
// space, so the world mesh is a straight composition — the same placement the
// renderer's grid walk performs each frame.

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
	b := newMeshBuilder(a.vram)
	cells := 0
	for z := 0; z < 32; z++ {
		for x := 0; x < 32; x++ {
			sec := a.grid.Section(x, z)
			if sec == rr.Empty || int(sec) >= len(a.track.Sections) {
				continue
			}
			cells++
			off := [3]int32{int32(x) * rr.CellSize, 0, int32(z) * rr.CellSize}
			s := &a.track.Sections[sec]
			for _, class := range [][]rr.TrackQuad{s.A, s.B, s.C} {
				for _, q := range class {
					b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off)
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
	return ModelIndex{Name: "Ridge Racer Course", File: file, Kind: "mesh3d"}, nil
}
