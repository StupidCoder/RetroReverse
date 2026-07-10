package f3d

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"retroreverse.com/tools/lib/glb"
)

// WriteObject merges draw groups into one GLB. Vertices are model-space; a
// group whose matrix differs from the first group's is baked through the
// relative transform, so an articulated model (a gyrocopter's rotor over its
// body) reassembles the way the frame posed it.
func (w *Walker) WriteObject(dir, name string, gs []*Group) {
	os.MkdirAll(dir, 0o755)
	root := InvertAffine(gs[0].Mtx)

	var pos [][3]float32
	var uvs [][2]float32
	var cols [][4]uint8
	var texGroups []glb.TexturedGroup
	for _, g := range gs {
		rel := Mul(g.Mtx, root)
		img := w.DecodeTexture(g)
		st := uv(g)

		var tris [][3]uint32
		for _, f := range g.Faces {
			base := uint32(len(pos))
			for _, vi := range f {
				v := g.Verts[vi]
				x, y, z := float64(v.X), float64(v.Y), float64(v.Z)
				pos = append(pos, [3]float32{
					float32(rel[0][0]*x + rel[1][0]*y + rel[2][0]*z + rel[3][0]),
					float32(rel[0][1]*x + rel[1][1]*y + rel[2][1]*z + rel[3][1]),
					float32(rel[0][2]*x + rel[1][2]*y + rel[2][2]*z + rel[3][2]),
				})
				u, vv := st(v.S, v.T)
				uvs = append(uvs, [2]float32{u, vv})
				cols = append(cols, [4]uint8{v.R, v.G, v.B, v.A})
			}
			tris = append(tris, [3]uint32{base, base + 1, base + 2})
		}
		if img == nil {
			img = whitePixel()
		}
		texGroups = append(texGroups, glb.TexturedGroup{
			Tris: tris, Image: img,
			WrapS: wrapMode(g.Tile.CmS, g.Tile.MaskS),
			WrapT: wrapMode(g.Tile.CmT, g.Tile.MaskT),
			// FORCE_BL in the render mode is the game's translucency: the
			// rotor-blur discs (vertex alpha 38..153 carrying the blur
			// amount) and the surf decals render through the blender.
			Blend: g.OtherL&OMForceBlend != 0,
		})
	}

	path := fmt.Sprintf("%s/%s.glb", dir, name)
	if err := glb.WriteTexturedColored(path, pos, uvs, cols, texGroups, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return
	}
	fmt.Printf("wrote %s: %d groups, %d verts\n", path, len(gs), len(pos))
}

// WriteGLBs writes one textured GLB per draw group, in model space (the raw
// vertex coordinates the game keeps in RDRAM), with UVs derived from the
// vertex s/t, the G_TEXTURE scale, and the tile extent.
func (w *Walker) WriteGLBs(dir string) {
	os.MkdirAll(dir, 0o755)
	for _, g := range w.Ordered() {
		if len(g.Faces) == 0 {
			continue
		}
		img := w.DecodeTexture(g)
		st := uv(g)

		var pos [][3]float32
		var uvs [][2]float32
		var tris [][3]uint32
		for _, f := range g.Faces {
			base := uint32(len(pos))
			for _, vi := range f {
				v := g.Verts[vi]
				pos = append(pos, [3]float32{float32(v.X), float32(v.Y), float32(v.Z)})
				u, vv := st(v.S, v.T)
				uvs = append(uvs, [2]float32{u, vv})
			}
			tris = append(tris, [3]uint32{base, base + 1, base + 2})
		}

		path := fmt.Sprintf("%s/%s.glb", dir, g.Name)
		var err error
		if img != nil {
			err = glb.WriteTextured(path, pos, uvs, []glb.TexturedGroup{{Tris: tris, Image: img}}, nil)
		} else {
			err = glb.WriteTrianglesMat(path, pos, []glb.TriGroup{{Tris: tris}})
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		}
	}
}

// uv maps a vertex s/t (10.5 texel coordinates) to a [0,1] coordinate of the
// group's decoded tile window: the G_TEXTURE scale applies first, then the
// tile's coordinate shift (0-10 right, 11-15 left by 16-n), then the window
// origin SL/TL (10.2) is subtracted and the result normalised by the extent.
func uv(g *Group) func(s, t int16) (float32, float32) {
	tw := float64(g.Tile.SH>>2-g.Tile.SL>>2) + 1
	th := float64(g.Tile.TH>>2-g.Tile.TL>>2) + 1
	sScale := float64(g.Scale[0]) / 65536 * shiftScale(g.Tile.ShiftS)
	tScale := float64(g.Scale[1]) / 65536 * shiftScale(g.Tile.ShiftT)
	s0 := float64(g.Tile.SL) / 4
	t0 := float64(g.Tile.TL) / 4
	return func(s, t int16) (float32, float32) {
		return float32((float64(s)/32*sScale - s0) / tw),
			float32((float64(t)/32*tScale - t0) / th)
	}
}

func shiftScale(sh uint32) float64 {
	if sh == 0 {
		return 1
	}
	if sh <= 10 {
		return 1 / float64(uint32(1)<<sh)
	}
	return float64(uint32(1) << (16 - sh))
}

// wrapMode maps a tile's cm/mask to the glTF sampler wrap enum. The cm bits
// are libultra's G_TX constants: 1 mirrors, 2 clamps; no mask forces a clamp.
func wrapMode(cm, mask uint32) int {
	switch {
	case cm&2 != 0 || mask == 0:
		return 33071 // CLAMP_TO_EDGE
	case cm&1 != 0:
		return 33648 // MIRRORED_REPEAT
	default:
		return 10497 // REPEAT
	}
}

func whitePixel() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})
	return img
}
