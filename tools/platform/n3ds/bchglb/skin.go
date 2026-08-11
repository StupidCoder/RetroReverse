// Package bchglb turns a decoded BCH model into glTF scene nodes: the skeleton
// as a node hierarchy, and each mesh as a skinned primitive bound to it. It is
// shared by the development instrument (tools/cmd/bchglb) and the game's web
// export, so both build a character the same way.
//
// # The two skinning spaces
//
// A BCH mesh does not have one skinning rule; it has one of two, and the face
// header's skinning mode says which:
//
//	mode 1 (smooth)  vertices are in the MODEL's space, the matrix is world(bone) x invBind(bone)
//	mode 2 (rigid)   vertices are in the BONE's own space, the matrix is world(bone)
//
// That is not inferred from the geometry: it was read out of the running game.
// The PICA has no matrix-palette registers, so a skinned draw's matrices arrive
// as ordinary vertex-shader float uniforms, three rows each from c25 in the
// palette's order — and the oracle prints them (`bootoracle -gputrace
// -gpuuniforms`). For Captain Toad's opening stage, with Toad placed at
// (-400, -800, 550):
//
//   - his rucksack (mode 1, bones Hip/Spine2/ShoulderL/ShoulderR/Ruck1/Ruck2)
//     is handed six matrices that are all within three units of the placement
//     itself — which is what world x invBind is at a pose near the bind pose,
//     and is nothing like six bones' world matrices, which would be a hundred
//     units apart in Y.
//   - his lamp strap (mode 2, one bone, HeadBelt) is handed a matrix
//     translating to (-414, -677, 541): 123 units above the placement, and
//     HeadBelt stands 124 units up the bind pose. His head mesh's bone stands
//     at 72 and is handed -728; his headlamp's at 146 and is handed -653.
//     Those are the bones' world matrices, with no inverse bind anywhere.
//
// Both rules are real, and applying either one to a whole character breaks the
// half of it that wanted the other — which is exactly what "the body hangs
// below the floor" and "closer to standing, still a jumble" were.
//
// glTF has only the one rule (world x inverseBindMatrix), so a rigid mesh is
// expressed by giving its joints an inverse bind matrix of *identity*. Since
// the same bone therefore needs different inverse-bind matrices in different
// meshes, each mesh gets its own skin — which glTF allows, a skin being
// referenced by a node and a node carrying one mesh.
package bchglb

import (
	"math"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
)

// Rig is a model's skeleton after it has been written into a scene: the glTF
// node index of each bone, in the model's own bone order.
type Rig struct {
	Nodes []int
	model *n3ds.BCHModel
}

// AddSkeleton writes a model's bones into the scene as a node hierarchy under
// parent (-1 for a scene root), naming each node after its bone.
//
// The local transform is the bone table's own TRS, with the Euler triple
// composed Rz·Ry·Rx — the order the bind pose closes in.
func AddSkeleton(s *glb.Scene, m *n3ds.BCHModel, parent int, prefix string) *Rig {
	r := &Rig{Nodes: make([]int, len(m.Bones)), model: m}
	for i, b := range m.Bones {
		p := parent
		if b.Parent >= 0 && b.Parent < i {
			p = r.Nodes[b.Parent]
		}
		r.Nodes[i] = s.AddNode(prefix+b.Name, p, b.Trans, quatZYX(b.Rotate), b.Scale)
	}
	return r
}

// Skin registers the skin one mesh needs and returns its index: the palette's
// bones as joints, with inverse-bind matrices chosen by the mesh's skinning
// mode. A mesh the model does not skin returns -1.
func (r *Rig) Skin(s *glb.Scene, sh *n3ds.BCHMesh) int {
	if len(sh.Palette) == 0 || len(r.Nodes) == 0 {
		return -1
	}
	joints := make([]int, 0, len(sh.Palette))
	ibm := make([][16]float32, 0, len(sh.Palette))
	for _, b := range sh.Palette {
		if b < 0 || b >= len(r.Nodes) {
			return -1
		}
		joints = append(joints, r.Nodes[b])
		switch sh.SkinMode {
		case n3ds.BCHSkinRigid:
			// The vertices are already in this bone's space, so the only
			// transform they want is the bone's world matrix. glTF always
			// applies world x inverseBind, so the inverse bind is the identity.
			ibm = append(ibm, [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1})
		default:
			ibm = append(ibm, colMajor(r.model.Bones[b].InvBind4()))
		}
	}
	return s.AddSkin(joints, ibm)
}

// BindJoints fills a primitive's joint and weight attributes from a mesh. The
// joints are palette slots, which is what the skin's joint list is indexed by,
// so no remapping happens here or anywhere.
func BindJoints(p *glb.Prim, sh *n3ds.BCHMesh) {
	if len(sh.Palette) == 0 {
		return
	}
	p.JointsW = make([][4]uint8, len(sh.Verts))
	p.Weights = make([][4]float32, len(sh.Verts))
	for i, v := range sh.Verts {
		p.JointsW[i] = v.Joints
		p.Weights[i] = v.Weights
	}
}

// colMajor transposes a row-major 4x4 into the column-major order glTF stores
// matrices in.
func colMajor(m [16]float64) [16]float32 {
	var o [16]float32
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			o[c*4+r] = float32(m[r*4+c])
		}
	}
	return o
}

// quatZYX converts a bone's XYZ Euler triple (radians) to the quaternion of
// Rz·Ry·Rx, the order the skeleton's bind pose closes in.
func quatZYX(r [3]float32) [4]float32 {
	cx, sx := math.Cos(float64(r[0])/2), math.Sin(float64(r[0])/2)
	cy, sy := math.Cos(float64(r[1])/2), math.Sin(float64(r[1])/2)
	cz, sz := math.Cos(float64(r[2])/2), math.Sin(float64(r[2])/2)
	return [4]float32{
		float32(sx*cy*cz - cx*sy*sz),
		float32(cx*sy*cz + sx*cy*sz),
		float32(cx*cy*sz - sx*sy*cz),
		float32(cx*cy*cz + sx*sy*sz),
	}
}
