package main

// course.go bakes the course surface: for every streamed segment, the slice
// cross-section row (11 world-space points) is stitched to the next row by
// the static face list of the segment's full-detail slice type, each quad
// textured by the cel its material byte resolves to. This is exactly the
// geometry the game's own scan (0x1A7DC) draws for near segments — the
// road surface, curbs, walls and trackside terrain are all slice faces.

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/lib/glb"
)

// vkey dedupes vertices on position + UV corner.
type vkey struct {
	x, y, z nfs.Fx16
	u, v    float32
}

type courseBuilder struct {
	verts []vkey
	index map[vkey]uint32
	byTex map[*image.RGBA][][3]uint32
	order []*image.RGBA
}

func (b *courseBuilder) vertex(p nfs.Vec3, u, v float32) uint32 {
	k := vkey{p.X, p.Y, p.Z, u, v}
	if i, ok := b.index[k]; ok {
		return i
	}
	i := uint32(len(b.verts))
	b.verts = append(b.verts, k)
	b.index[k] = i
	return i
}

// quad adds one textured quad (corner order v0..v3 around the perimeter,
// matching the cel engine's MapCel corner convention).
func (b *courseBuilder) quad(img *image.RGBA, p [4]nfs.Vec3) {
	uv := [4][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	var idx [4]uint32
	for k := 0; k < 4; k++ {
		idx[k] = b.vertex(p[k], uv[k][0], uv[k][1])
	}
	if _, ok := b.byTex[img]; !ok {
		b.order = append(b.order, img)
	}
	b.byTex[img] = append(b.byTex[img],
		[3]uint32{idx[0], idx[1], idx[2]}, [3]uint32{idx[0], idx[2], idx[3]})
}

func exportCourse(a *assets, out string) (string, error) {
	b := &courseBuilder{index: map[vkey]uint32{}, byTex: map[*image.RGBA][][3]uint32{}}

	// Face vertex indices address the group's compacted 4-row vertex batch:
	// the concatenation of each row's slice-type vertex subset (sparse rows
	// carry fewer points — walls repeat every second segment), then the NEXT
	// group's first row from the type's vertex total onward. This is exactly
	// the batch the game's world-vertex builder (0x15F70) accumulates and the
	// face scan indexes.
	used := a.track.UsedSegments()
	missing := map[string]int{}
	for g := 0; g+1 < len(a.track.Groups); g++ {
		grp := &a.track.Groups[g]
		typ, err := a.slices.TypeFor(a.track.Segments[4*g].SliceFamily)
		if err != nil {
			return "", fmt.Errorf("group %d: %v", g, err)
		}
		var batch []nfs.Vec3
		for r := 0; r < 4; r++ {
			for _, pi := range typ.RowVerts[r] {
				batch = append(batch, grp.Rows[r][pi])
			}
		}
		typN, err := a.slices.TypeFor(a.track.Segments[4*(g+1)].SliceFamily)
		if err != nil {
			return "", fmt.Errorf("group %d: %v", g+1, err)
		}
		nextRow := &a.track.Groups[g+1].Rows[0]
		var next []nfs.Vec3
		for _, pi := range typN.RowVerts[0] {
			next = append(next, nextRow[pi])
		}

		for _, f := range a.slices.Faces(typ) {
			var p [4]nfs.Vec3
			ok := true
			for k := 0; k < 4; k++ {
				vi := f.V[k]
				switch {
				case vi < len(batch):
					p[k] = batch[vi]
				case vi-typ.VertTotal >= 0 && vi-typ.VertTotal < len(next):
					p[k] = next[vi-typ.VertTotal]
				default:
					ok = false
				}
			}
			if !ok {
				missing["face vertex out of range"]++
				continue
			}
			img, err := a.tex.sliceTexture(int(grp.Materials[f.Material]), int(f.TexSub))
			if err != nil {
				missing[err.Error()]++
				continue
			}
			b.quad(img, p)
		}
	}
	for e, n := range missing {
		fmt.Fprintf(os.Stderr, "[course] skipped %d faces: %s\n", n, e)
	}

	positions := make([][3]float32, len(b.verts))
	uvs := make([][2]float32, len(b.verts))
	for i, k := range b.verts {
		positions[i] = gl(nfs.Vec3{X: k.x, Y: k.y, Z: k.z})
		uvs[i] = [2]float32{k.u, k.v}
	}
	var groups []glb.TexturedGroup
	for _, img := range b.order {
		groups = append(groups, glb.TexturedGroup{Tris: b.byTex[img], Image: img})
	}
	file := fmt.Sprintf("models/course-%s.glb", a.course)
	if err := glb.WriteTextured(filepath.Join(out, file), positions, uvs, groups, nil); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[course] %d segments, %d verts, %d textures -> %s\n",
		used, len(positions), len(groups), file)
	return file, nil
}
