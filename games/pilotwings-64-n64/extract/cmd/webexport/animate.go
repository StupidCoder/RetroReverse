package main

// Exporting the game's articulated models as animated glTF.
//
// A UVMD's parts are rigid pieces, each with a rest pose that is pure
// translation (identity 3x3 — checked across the animated models): part 0 at the
// origin, the others at their pivots. A UVAN animates such a model by, per part,
// replacing that part matrix's rotation with a keyframed quaternion while keeping
// the translation. There is no skeleton hierarchy and no skinning: the apply loop
// (0x80200638) rotates each part about its own pivot and stores the matrix, with
// no parent concatenation. So each part becomes one glTF node — translation at the
// pivot, an animated rotation — and the animation is a rotation channel per part.
//
// Two frames of reference meet here. The game is Z-up and glTF is Y-up, joined by
// B: (x, y, z) -> (x, z, -y), a -90 degrees rotation about X. A vertex the game
// rotates by quaternion q about a pivot must, in glTF space, rotate by the same
// rotation expressed in the new basis: q' = b * q * b^-1, with b the quaternion of
// B. The node translation is B applied to the rest pivot. Worked through, this
// makes the glTF node reproduce the game's transform exactly (the quaternion
// convention already matches: the routine at 0x80229D78 is the standard formula,
// and row-vector times matrix equals glTF's matrix times column-vector for the
// same q).

import (
	"fmt"
	"path/filepath"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvan"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/lib/glb"
)

// animFPS turns a UVAN frame index into a playback time. The game's exact rate is
// not pinned (the header's rate word is 1 for all but six resources), so this is
// a presentation choice, not a decoded value: it sets how fast the clips play.
const animFPS = 12.0

// bX is the quaternion of the Z-up -> Y-up basis change B: a -90 degrees rotation
// about X, so B(x,y,z) = (x, z, -y). Its conjugate is bXconj.
var bX = quat{-0.70710678, 0, 0, 0.70710678}
var bXconj = quat{0.70710678, 0, 0, 0.70710678}

type quat struct{ X, Y, Z, W float32 }

func (a quat) mul(b quat) quat {
	return quat{
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y,
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X,
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W,
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z,
	}
}

// toGLQuat expresses a game-space rotation quaternion in glTF's Y-up basis.
func toGLQuat(x, y, z, w float32) [4]float32 {
	q := bX.mul(quat{x, y, z, w}).mul(bXconj)
	return [4]float32{q.X, q.Y, q.Z, q.W}
}

// gatherAnimations groups every UVAN by the UVMD ordinal it targets.
func gatherAnimations(a *pwad.Archive) map[int][]*uvan.Animation {
	out := map[int][]*uvan.Animation{}
	for _, r := range a.ByType("UVAN") {
		f, err := a.Resource(r)
		if err != nil {
			die(err)
		}
		var comm []byte
		var parts [][]byte
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			data, err := a.Data(c)
			if err != nil {
				die(err)
			}
			switch tag {
			case "COMM":
				comm = data
			case "PART":
				parts = append(parts, data)
			}
		}
		anim, err := uvan.Decode(comm, parts)
		if err != nil {
			die(fmt.Errorf("UVAN %d: %w", r, err))
		}
		out[anim.Model] = append(out[anim.Model], anim)
	}
	return out
}

// writeAnimatedModel writes one model as a rigged, animated GLB and returns the
// names of the clips it carries.
func writeAnimatedModel(path string, m *uvmd.Model, anims []*uvan.Animation, texs []*uvtx.Texture) []string {
	lod := m.LODs[0]

	// Shared vertex arrays, in part-local space (toGL'd but NOT placed by the rest
	// pose — the node translation does that). Triangles index into them, grouped
	// per part and per texture.
	var pos [][3]float32
	var uvs [][2]float32
	var cols [][4]uint8

	parts := make([]glb.RigPart, len(lod.Parts))
	for pi, part := range lod.Parts {
		byTex := map[int][][3]uint32{}
		for _, bt := range part.Batches {
			tex := -1
			tw, th := 1.0, 1.0
			if t, ok := bt.Material.Texture(); ok && t < len(texs) {
				tex = t
				tw, th = float64(texs[t].Width), float64(texs[t].Height)
			}
			for _, tri := range bt.Tris {
				base := uint32(len(pos))
				for _, vi := range tri {
					v := m.Vertices[vi]
					// Keep the vertex part-local (no rest transform); the node
					// translation places the part. Only the axis flip is baked.
					pos = append(pos, toGL(float32(v.X), float32(v.Y), float32(v.Z)))
					uvs = append(uvs, [2]float32{float32(float64(v.S) / 32 / tw), float32(float64(v.T) / 32 / th)})
					cols = append(cols, [4]uint8{v.R, v.G, v.B, v.A})
				}
				byTex[tex] = append(byTex[tex], [3]uint32{base, base + 1, base + 2})
			}
		}

		rt := m.Matrices[pi]
		parts[pi] = glb.RigPart{
			Name:        fmt.Sprintf("part-%02d", pi),
			Translation: toGL(rt[3][0], rt[3][1], rt[3][2]),
		}
		for tex, tris := range byTex {
			g := glb.TexturedGroup{Tris: tris, WrapS: 10497, WrapT: 10497}
			if tex >= 0 {
				g.Image, g.Blend = exportImage(texs[tex])
			} else {
				g.Image = white()
			}
			parts[pi].Tex = append(parts[pi].Tex, g)
		}
	}

	// One clip per UVAN, one rotation track per animated part.
	var clips []glb.RigClip
	var names []string
	for ci, an := range anims {
		clip := glb.RigClip{Name: fmt.Sprintf("clip-%02d", ci)}
		for _, tr := range an.Tracks {
			if tr.Part >= len(parts) || len(tr.Keys) == 0 {
				continue
			}
			rc := glb.RigTrack{Part: tr.Part}
			for _, k := range tr.Keys {
				rc.Times = append(rc.Times, float32(k.Frame)/animFPS)
				rc.Rots = append(rc.Rots, toGLQuat(k.Rot.X, k.Rot.Y, k.Rot.Z, k.Rot.W))
			}
			clip.Tracks = append(clip.Tracks, rc)
		}
		if len(clip.Tracks) > 0 {
			clips = append(clips, clip)
			names = append(names, clip.Name)
		}
	}

	if err := glb.WriteRiggedAnimated(path, pos, uvs, cols, parts, clips); err != nil {
		die(fmt.Errorf("%s: %w", filepath.Base(path), err))
	}
	return names
}
