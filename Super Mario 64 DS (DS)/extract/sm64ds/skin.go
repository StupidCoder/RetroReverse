// Skinned export: the .bmd skeleton (bind pose) plus .bca clips as a
// nitro.Skin. Conventions traced from the engine: a bone's local transform is
// v -> S -> Rx -> Ry -> Rz -> T in row-vector terms ($02045178/$02045074),
// which in glTF's column T·R·S composition maps to translation = T, rotation =
// the transposed row matrix as a quaternion, scale = S. The inverse-bind
// matrices invert the hierarchy world WITHOUT the model-wide 2^shift — that
// scale is baked into the exported vertices and must stay.
package sm64ds

import (
	"math"

	"retroreverse.com/tools/nds/nitro"
)

// SkelJoint is one .bmd bone's bind-pose local transform.
type SkelJoint struct {
	Name    string
	Parent  int // absolute index, -1 for roots
	S, T    [3]float64
	R       [3]float64 // radians, engine Rx·Ry·Rz order
}

// NamedBCA pairs a decoded animation with its clip name.
type NamedBCA struct {
	Name string
	Anim *BCA
}

// rowRot builds the engine's rotation matrix (row-vector convention).
func rowRot(rx, ry, rz float64) [9]float64 {
	cx, sx := math.Cos(rx), math.Sin(rx)
	cy, sy := math.Cos(ry), math.Sin(ry)
	cz, sz := math.Cos(rz), math.Sin(rz)
	return [9]float64{
		cy * cz, cy * sz, -sy,
		sx*sy*cz - cx*sz, cx*cz + sx*sy*sz, sx * cy,
		cx*sy*cz + sx*sz, cx*sy*sz - sx*cz, cx * cy,
	}
}

// quatFromEuler converts engine angles to a glTF quaternion (xyzw): the
// column-convention rotation matrix is the transpose of the row one.
func quatFromEuler(rx, ry, rz float64) [4]float64 {
	r := rowRot(rx, ry, rz)
	// column matrix c[i][j] = r[j*3+i] (transpose)
	m00, m01, m02 := r[0], r[3], r[6]
	m10, m11, m12 := r[1], r[4], r[7]
	m20, m21, m22 := r[2], r[5], r[8]
	tr := m00 + m11 + m22
	var x, y, z, w float64
	switch {
	case tr > 0:
		s := math.Sqrt(tr+1) * 2
		w = s / 4
		x = (m21 - m12) / s
		y = (m02 - m20) / s
		z = (m10 - m01) / s
	case m00 > m11 && m00 > m22:
		s := math.Sqrt(1+m00-m11-m22) * 2
		x = s / 4
		w = (m21 - m12) / s
		y = (m01 + m10) / s
		z = (m02 + m20) / s
	case m11 > m22:
		s := math.Sqrt(1+m11-m00-m22) * 2
		y = s / 4
		w = (m02 - m20) / s
		x = (m01 + m10) / s
		z = (m12 + m21) / s
	default:
		s := math.Sqrt(1+m22-m00-m11) * 2
		z = s / 4
		w = (m10 - m01) / s
		x = (m02 + m20) / s
		y = (m12 + m21) / s
	}
	return [4]float64{x, y, z, w}
}

// localMat43 composes a joint's local transform in nitro.Mat43 row convention.
func localMat43(s, r, t [3]float64) nitro.Mat43 {
	rm := rowRot(r[0], r[1], r[2])
	m := nitro.Mat43{
		rm[0], rm[1], rm[2],
		rm[3], rm[4], rm[5],
		rm[6], rm[7], rm[8],
		t[0], t[1], t[2],
	}
	if s != [3]float64{1, 1, 1} {
		m = m.Mul(nitro.Mat43{s[0], 0, 0, 0, s[1], 0, 0, 0, s[2], 0, 0, 0})
	}
	return m
}

// invertAffine inverts a row-convention 4x3 (general 3x3 + translation).
func invertAffine(m nitro.Mat43) nitro.Mat43 {
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[3], m[4], m[5]
	g, h, i := m[6], m[7], m[8]
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	if det == 0 {
		return nitro.Identity43()
	}
	id := 1 / det
	var r nitro.Mat43
	r[0], r[1], r[2] = (e*i-f*h)*id, (c*h-b*i)*id, (b*f-c*e)*id
	r[3], r[4], r[5] = (f*g-d*i)*id, (a*i-c*g)*id, (c*d-a*f)*id
	r[6], r[7], r[8] = (d*h-e*g)*id, (b*g-a*h)*id, (a*e-b*d)*id
	// translation: -T·Rinv (row convention: v' = (v - T)·Rinv)
	tx, ty, tz := m[9], m[10], m[11]
	r[9] = -(tx*r[0] + ty*r[3] + tz*r[6])
	r[10] = -(tx*r[1] + ty*r[4] + tz*r[7])
	r[11] = -(tx*r[2] + ty*r[5] + tz*r[8])
	return r
}

// BuildSkin assembles the glTF skin from the model's skeleton and clips.
func (m *Model) BuildSkin(anims []NamedBCA) *nitro.Skin {
	if len(m.Skel) == 0 {
		return nil
	}
	sk := &nitro.Skin{}
	worlds := make([]nitro.Mat43, len(m.Skel))
	for i, j := range m.Skel {
		local := localMat43(j.S, j.R, j.T)
		if j.Parent >= 0 {
			worlds[i] = worlds[j.Parent].Mul(local)
		} else {
			worlds[i] = local
		}
		inv := invertAffine(worlds[i])
		var ibm [16]float64 // column-major = transpose of the row matrix
		ibm[0], ibm[4], ibm[8], ibm[12] = inv[0], inv[3], inv[6], inv[9]
		ibm[1], ibm[5], ibm[9], ibm[13] = inv[1], inv[4], inv[7], inv[10]
		ibm[2], ibm[6], ibm[10], ibm[14] = inv[2], inv[5], inv[8], inv[11]
		ibm[15] = 1
		sk.Joints = append(sk.Joints, nitro.Joint{
			Name:   j.Name,
			Parent: j.Parent,
			T:      j.T,
			R:      quatFromEuler(j.R[0], j.R[1], j.R[2]),
			S:      j.S,
			IBM:    ibm,
		})
	}
	for _, na := range anims {
		if na.Anim.NumBones != len(m.Skel) {
			continue
		}
		a := nitro.SkinAnim{Name: na.Name, FPS: 30}
		nf := na.Anim.NumFrames
		for ji := range m.Skel {
			var ts [][3]float64
			var rs [][4]float64
			var ss [][3]float64
			for f := 0; f < nf; f++ {
				v := na.Anim.BoneTRS(ji, f)
				ss = append(ss, [3]float64{v[0], v[1], v[2]})
				rs = append(rs, quatFromEuler(v[3], v[4], v[5]))
				ts = append(ts, [3]float64{v[6], v[7], v[8]})
			}
			a.T = append(a.T, ts)
			a.R = append(a.R, rs)
			a.S = append(a.S, ss)
		}
		sk.Anims = append(sk.Anims, a)
	}
	return sk
}

// SkinnedGLB exports the model with its skeleton and the given clips.
func (m *Model) SkinnedGLB(anims []NamedBCA) ([]byte, error) {
	return nitro.ExportSkinnedGLB(m.Name, m.ByMat, m.Mats, m.Texs, m.BuildSkin(anims))
}
