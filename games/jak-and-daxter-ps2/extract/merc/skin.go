package merc

// skin.go: the skeleton side of the export. GOAL matrices are row-vector
// convention (v' = v * M, translation in row 3); a joint's Bind field is the
// INVERSE bind pose (world -> bone), which is exactly glTF's
// inverseBindMatrices after transposition. The merc vertex's matrix byte
// (lump q0.x) selects the bone palette slot: addr = byte & 0x7F resolves
// through the running MatXfer dest map to a palette index, and palette
// index - 1 is the joint number; 0xFF means "the previous vertex's matrix"
// (the microcode just keeps the loaded rows).

import "math"

// Mat4 is row-major, GOAL row-vector convention.
type Mat4 [16]float32

func (m Mat4) Mul(n Mat4) (r Mat4) { // r = m then n (v*m*n)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += m[i*4+k] * n[k*4+j]
			}
			r[i*4+j] = s
		}
	}
	return
}

// Inverse inverts a rigid-with-scale transform (R upper 3x3, T row 3).
func (m Mat4) Inverse() (r Mat4) {
	// invert 3x3 by adjugate
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[4], m[5], m[6]
	g, h, i := m[8], m[9], m[10]
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	if det == 0 {
		det = 1
	}
	id := 1 / det
	r[0], r[1], r[2] = (e*i-f*h)*id, (c*h-b*i)*id, (b*f-c*e)*id
	r[4], r[5], r[6] = (f*g-d*i)*id, (a*i-c*g)*id, (c*d-a*f)*id
	r[8], r[9], r[10] = (d*h-e*g)*id, (b*g-a*h)*id, (a*e-b*d)*id
	// row-vector: T' = -T * Rinv
	tx, ty, tz := m[12], m[13], m[14]
	r[12] = -(tx*r[0] + ty*r[4] + tz*r[8])
	r[13] = -(tx*r[1] + ty*r[5] + tz*r[9])
	r[14] = -(tx*r[2] + ty*r[6] + tz*r[10])
	r[15] = 1
	return
}

// TRS decomposes (uniform-ish scale assumed).
func (m Mat4) TRS() (t [3]float32, q [4]float32, s [3]float32) {
	t = [3]float32{m[12], m[13], m[14]}
	sx := float32(math.Sqrt(float64(m[0]*m[0] + m[1]*m[1] + m[2]*m[2])))
	sy := float32(math.Sqrt(float64(m[4]*m[4] + m[5]*m[5] + m[6]*m[6])))
	sz := float32(math.Sqrt(float64(m[8]*m[8] + m[9]*m[9] + m[10]*m[10])))
	s = [3]float32{sx, sy, sz}
	if sx == 0 || sy == 0 || sz == 0 {
		q = [4]float32{0, 0, 0, 1}
		return
	}
	// normalized rotation rows (row-vector convention)
	r00, r01, r02 := float64(m[0]/sx), float64(m[1]/sx), float64(m[2]/sx)
	r10, r11, r12 := float64(m[4]/sy), float64(m[5]/sy), float64(m[6]/sy)
	r20, r21, r22 := float64(m[8]/sz), float64(m[9]/sz), float64(m[10]/sz)
	// quaternion of the COLUMN-vector matrix = transpose, i.e. swap the
	// off-diagonal differences' signs relative to the row form.
	tr := r00 + r11 + r22
	var x, y, z, w float64
	switch {
	case tr > 0:
		s4 := math.Sqrt(tr+1) * 2
		w = s4 / 4
		x = (r12 - r21) / s4
		y = (r20 - r02) / s4
		z = (r01 - r10) / s4
	case r00 > r11 && r00 > r22:
		s4 := math.Sqrt(1+r00-r11-r22) * 2
		x = s4 / 4
		w = (r12 - r21) / s4
		y = (r10 + r01) / s4
		z = (r20 + r02) / s4
	case r11 > r22:
		s4 := math.Sqrt(1+r11-r00-r22) * 2
		y = s4 / 4
		w = (r20 - r02) / s4
		x = (r10 + r01) / s4
		z = (r21 + r12) / s4
	default:
		s4 := math.Sqrt(1+r22-r00-r11) * 2
		z = s4 / 4
		w = (r01 - r10) / s4
		x = (r20 + r02) / s4
		y = (r21 + r12) / s4
	}
	q = [4]float32{float32(x), float32(y), float32(z), float32(w)}
	return
}

// QuatMat builds the row-vector rotation matrix of a quaternion (the
// engine's matrix-from-quat convention, transposed vs column form).
func QuatMat(q [4]float32) (m Mat4) {
	x, y, z, w := q[0], q[1], q[2], q[3]
	m[0] = 1 - 2*(y*y+z*z)
	m[1] = 2 * (x*y + z*w)
	m[2] = 2 * (x*z - y*w)
	m[4] = 2 * (x*y - z*w)
	m[5] = 1 - 2*(x*x+z*z)
	m[6] = 2 * (y*z + x*w)
	m[8] = 2 * (x*z + y*w)
	m[9] = 2 * (y*z - x*w)
	m[10] = 1 - 2*(x*x+y*y)
	m[15] = 1
	return
}

// PoseLocal builds a joint pose's local matrix (row convention):
// scale, rotate, translate.
func PoseLocal(p *JointPose) Mat4 {
	if p.Matrix != nil {
		return Mat4(*p.Matrix)
	}
	m := QuatMat(p.Quat)
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			m[r*4+c] *= p.Scale[r]
		}
	}
	m[12], m[13], m[14] = p.Trans[0], p.Trans[1], p.Trans[2]
	return m
}

// GeoJoints reads an art-joint-geo's joint list: count at basic+8, joint
// basic pointers from basic+28.
func GeoJoints(obj []byte, geoBasic uint32, jointType uint32) []Joint {
	n := int(u32(obj, geoBasic+8))
	all := Joints(obj, jointType)
	byOff := map[uint32]int{}
	// rebuild offsets the same way Joints found them
	idx := 0
	for o := uint32(0); int(o)+84 <= len(obj) && idx < len(all); o += 4 {
		if u32(obj, o) == jointType {
			byOff[o+4] = idx
			idx++
			o += 76
		}
	}
	out := make([]Joint, 0, n)
	remap := map[int]int{}
	for k := 0; k < n; k++ {
		p := u32(obj, geoBasic+28+uint32(k)*4)
		if gi, ok := byOff[p]; ok {
			remap[gi] = len(out)
			out = append(out, all[gi])
		}
	}
	// reindex parents into the geo-local list
	for i := range out {
		if out[i].Parent >= 0 {
			if l, ok := remap[out[i].Parent]; ok {
				out[i].Parent = l
			} else {
				out[i].Parent = -1
			}
		}
	}
	return out
}

// FindGeos scans for art-joint-geo basics.
func FindGeos(obj []byte, geoType uint32) []uint32 {
	var out []uint32
	for o := uint32(0); int(o)+32 <= len(obj); o += 4 {
		if u32(obj, o) == geoType {
			out = append(out, o+4)
		}
	}
	return out
}
