package threedo

import "retroreverse.com/tools/cpu/arm60"

// operamath.go reimplements the Operamath (math) folio's fixed-point vector and
// matrix SWIs — the 3D math the game's renderer runs every frame. These were
// stubbed, so the world->screen transform of the road vertices did nothing and
// the projected road collapsed to a vertical line at screen centre (all four
// corners of every road strip landed on the same X). Reimplemented from
// operamath.h (MATHFOLIO = 5, MATHSWI = 5<<16); values are frac16 (16.16 fixed
// point), so each product of two frac16 is shifted right by 16.
//
//	+0x00 MulVec3Mat33_F16(dest, vec, mat)
//	+0x01 MulMat33Mat33_F16(dest, src1, src2)
//	+0x02 MulManyVec3Mat33_F16(dest, src, mat, count)
//	+0x0C Dot3_F16(v1, v2) -> frac16
//	+0x0E Cross3_F16(dest, v1, v2)
const (
	swiMulVec3Mat33     = 0x50000
	swiMulMat33Mat33    = 0x50001
	swiMulManyVec3Mat33 = 0x50002
	swiDot3             = 0x5000C
	swiCross3           = 0x5000E
)

// mathFolioSWI services a math-folio SWI, returning false for numbers it does
// not implement so the caller logs them as stubs.
func (m *Machine) mathFolioSWI(c *arm60.CPU, swi uint32) bool {
	switch swi {
	case swiMulVec3Mat33:
		m.mulVec3Mat33(c.Reg(0), c.Reg(1), c.Reg(2))
	case swiMulManyVec3Mat33:
		dest, src, mat, count := c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3)
		for i := uint32(0); i < count; i++ {
			m.mulVec3Mat33(dest+i*12, src+i*12, mat)
		}
	case swiMulMat33Mat33:
		m.mulMat33Mat33(c.Reg(0), c.Reg(1), c.Reg(2))
	case swiDot3:
		v1, v2 := c.Reg(0), c.Reg(1)
		var acc int64
		for i := uint32(0); i < 3; i++ {
			acc += int64(int32(m.read32(v1+i*4))) * int64(int32(m.read32(v2+i*4)))
		}
		c.SetReg(0, uint32(acc>>16))
	case swiCross3:
		m.cross3(c.Reg(0), c.Reg(1), c.Reg(2))
	default:
		return false
	}
	return true
}

// fmul16 multiplies two frac16 (16.16) values, keeping the result in 16.16.
func fmul16(a, b int32) int32 { return int32((int64(a) * int64(b)) >> 16) }

// readVec3 loads a vec3f16 from memory.
func (m *Machine) readVec3(p uint32) [3]int32 {
	return [3]int32{int32(m.read32(p)), int32(m.read32(p + 4)), int32(m.read32(p + 8))}
}

// mulVec3Mat33 computes dest = vec * mat, a row-vector times a 3x3 matrix:
// dest[j] = sum_i vec[i] * mat[i][j] (each product a frac16 multiply). This is
// the Operamath convention the game's matrices are built for.
func (m *Machine) mulVec3Mat33(dest, vec, mat uint32) {
	v := m.readVec3(vec)
	var out [3]int64
	for i := uint32(0); i < 3; i++ {
		for j := uint32(0); j < 3; j++ {
			out[j] += int64(v[i]) * int64(int32(m.read32(mat+(i*3+j)*4)))
		}
	}
	for j := uint32(0); j < 3; j++ {
		m.write32(dest+j*4, uint32(out[j]>>16))
	}
}

// mulMat33Mat33 computes dest = src1 * src2 (3x3 times 3x3), the transform
// composition the game uses to build its camera/object matrices.
func (m *Machine) mulMat33Mat33(dest, src1, src2 uint32) {
	var a, b [3][3]int32
	for i := uint32(0); i < 3; i++ {
		for j := uint32(0); j < 3; j++ {
			a[i][j] = int32(m.read32(src1 + (i*3+j)*4))
			b[i][j] = int32(m.read32(src2 + (i*3+j)*4))
		}
	}
	// dest and a source may alias, so compute into a temporary first.
	var out [3][3]int32
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			var acc int64
			for k := 0; k < 3; k++ {
				acc += int64(a[i][k]) * int64(b[k][j])
			}
			out[i][j] = int32(acc >> 16)
		}
	}
	for i := uint32(0); i < 3; i++ {
		for j := uint32(0); j < 3; j++ {
			m.write32(dest+(i*3+j)*4, uint32(out[i][j]))
		}
	}
}

// cross3 computes dest = v1 x v2 (vector cross product).
func (m *Machine) cross3(dest, v1p, v2p uint32) {
	a, b := m.readVec3(v1p), m.readVec3(v2p)
	out := [3]int32{
		fmul16(a[1], b[2]) - fmul16(a[2], b[1]),
		fmul16(a[2], b[0]) - fmul16(a[0], b[2]),
		fmul16(a[0], b[1]) - fmul16(a[1], b[0]),
	}
	for j := uint32(0); j < 3; j++ {
		m.write32(dest+j*4, uint32(out[j]))
	}
}
