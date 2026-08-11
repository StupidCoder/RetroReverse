package n3ds

import (
	"encoding/binary"
	"fmt"
	"math"
)

// cgfx_model.go decodes a CGFX model (CMDL) into plain geometry: meshes, shapes
// with de-interleaved vertices and triangle indices, material→texture bindings,
// and the skeleton. Every offset below was derived by dumping Super Mario 3D
// Land's banner CGFX and cross-checking the arithmetic (stride × vertex count =
// buffer size, attribute offsets summing to the stride, table pointers landing
// on magics); the writeup's Part III records the layout.
//
// All pointers are self-relative. "M" in the offset comments is the object's
// magic word; dictionary data pointers address the typeId word 4 bytes before it
// for models/textures (but not animations — see cgfx_anim.go).

// Vertex is one de-interleaved vertex. Fields a shape lacks stay zero; the
// shape's Has* flags say which are meaningful.
type Vertex struct {
	Pos    [3]float32
	Normal [3]float32
	Color  [4]uint8
	UV0    [2]float32
	UV1    [2]float32
}

// Shape is one SOBJ shape: a triangle list over its own vertex array, rigidly
// bound to one bone.
type Shape struct {
	BoneIndex int // index into Model.Bones (from the primitive set's bone table)
	OBBCenter [3]float32

	Verts     []Vertex
	Indices   []uint32
	HasNormal bool
	HasColor  bool
	UVCount   int
}

// Mesh binds a shape to a material.
type Mesh struct {
	ShapeIndex    int
	MaterialIndex int
}

// Material is the slice of an MTOB this exporter needs: its name and its
// texture mappers, in mapper order.
type Material struct {
	Name    string
	Mappers []TexMapper
}

// PICAWrap is a sampler's wrap mode, as the PICA200 encodes it in the texture
// unit's parameter register (0x083 / 0x093 / 0x09B, bits 12-14 for S and 8-10
// for T).
type PICAWrap uint8

// The four wrap modes. gpu_texture.go's wrapCoord implements them.
const (
	WrapClampToEdge    PICAWrap = 0
	WrapClampToBorder  PICAWrap = 1
	WrapRepeat         PICAWrap = 2
	WrapMirroredRepeat PICAWrap = 3
)

// TexMapper is one of a material's texture mappers: the texture it samples,
// which of the shape's UV sets feeds it, the transform applied to that UV, and
// the sampler's wrap modes.
//
// None of this is decoration. The banner's atlases are 512x256 while their
// vertex UVs address them as if they were square, and the coordinator's
// scaleT = 2 is what stretches the V range back over the whole image; ignoring
// it samples half the atlas and cuts the logo in two. Mario's UVs run outside
// [0,1] and his sampler is MIRRORED_REPEAT, not the REPEAT that a glTF viewer
// assumes by default.
type TexMapper struct {
	Texture  string
	SourceUV int // the coordinator's sourceCoordinate: which vertex UV set feeds this mapper

	// The coordinator's authored transform. Matrix is what the runtime uploads
	// to the shader and what Apply uses; the scalars are kept because they name
	// what the matrix means.
	ScaleS, ScaleT float32
	Rotate         float32
	TransS, TransT float32

	// Matrix is the two rows of the coordinator's baked 3x4 texture matrix that
	// produce (s, t) from (u, v, 1). The third row is the pass-through
	// [0 0 1 0] for the unused third coordinate.
	Matrix [2][3]float32

	WrapS, WrapT PICAWrap
}

// Apply transforms one UV pair through the mapper's texture matrix. It works in
// PICA texture space, where t = 0 is the *bottom* row — flip V for glTF after
// this, not before.
func (m *TexMapper) Apply(uv [2]float32) [2]float32 {
	return [2]float32{
		m.Matrix[0][0]*uv[0] + m.Matrix[0][1]*uv[1] + m.Matrix[0][2],
		m.Matrix[1][0]*uv[0] + m.Matrix[1][1]*uv[1] + m.Matrix[1][2],
	}
}

// Bone is one skeleton joint. Rotation is the CGFX XYZ Euler triple.
type Bone struct {
	Name   string
	Index  int
	Parent int // -1 for the root
	Scale  [3]float32
	Rot    [3]float32
	Trans  [3]float32
}

// Model is a fully decoded CMDL.
type Model struct {
	Name      string
	Meshes    []Mesh
	Shapes    []Shape
	Materials []Material
	Bones     []Bone
}

// GL type enums as used by the attribute and index-stream descriptors.
const (
	glByte   = 0x1400
	glUByte  = 0x1401
	glShort  = 0x1402
	glUShort = 0x1403
	glFloat  = 0x1406
)

func (c *CGFX) u32(off int64) uint32 { return binary.LittleEndian.Uint32(c.raw[off:]) }
func (c *CGFX) f32(off int64) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(c.raw[off:]))
}
func (c *CGFX) rel(off int64) int64 { return selfRel32(c.raw, off) }

func (c *CGFX) check(off int64, n int64, what string) error {
	if off <= 0 || off+n > int64(len(c.raw)) {
		return fmt.Errorf("cgfx: %s at 0x%x (+%d) runs outside the blob", what, off, n)
	}
	return nil
}

// DecodeModel decodes the CMDL a Models-dictionary entry points at.
func (c *CGFX) DecodeModel(e CGFXEntry) (*Model, error) {
	M := e.Offset + 4 // magic word
	if err := c.check(M, 0xE0, "CMDL header"); err != nil {
		return nil, err
	}
	if string(c.raw[M:M+4]) != "CMDL" {
		return nil, fmt.Errorf("cgfx: model entry %q is not a CMDL (magic %q)", e.Name, c.raw[M:M+4])
	}
	m := &Model{Name: e.Name}

	// Counts and tables. Offsets from the magic word M:
	//   meshes   count M+0xB0, table M+0xB4
	//   materials count M+0xB8, DICT  M+0xBC
	//   shapes   count M+0xC0, table M+0xC4
	//   skeleton ptr   M+0xDC
	meshN, meshT := int(c.u32(M+0xB0)), c.rel(M+0xB4)
	matN, matD := int(c.u32(M+0xB8)), c.rel(M+0xBC)
	shpN, shpT := int(c.u32(M+0xC0)), c.rel(M+0xC4)
	skel := c.rel(M + 0xDC)

	for i := 0; i < shpN; i++ {
		s, err := c.decodeShape(c.rel(shpT + int64(i)*4))
		if err != nil {
			return nil, fmt.Errorf("shape %d: %w", i, err)
		}
		m.Shapes = append(m.Shapes, *s)
	}
	for i := 0; i < meshN; i++ {
		X := c.rel(meshT+int64(i)*4) + 4 // magic
		mesh := Mesh{
			ShapeIndex:    int(c.u32(X + 0x14)),
			MaterialIndex: int(c.u32(X + 0x18)),
		}
		if mesh.ShapeIndex >= len(m.Shapes) || mesh.MaterialIndex >= matN {
			return nil, fmt.Errorf("mesh %d: shape %d / material %d out of range", i, mesh.ShapeIndex, mesh.MaterialIndex)
		}
		m.Meshes = append(m.Meshes, mesh)
	}

	if matD != 0 {
		entries, err := c.parseDict(matD)
		if err != nil {
			return nil, fmt.Errorf("materials dict: %w", err)
		}
		for _, me := range entries {
			mat, err := c.decodeMaterial(me)
			if err != nil {
				return nil, fmt.Errorf("material %q: %w", me.Name, err)
			}
			m.Materials = append(m.Materials, *mat)
		}
	}

	if skel != 0 {
		bones, err := c.decodeSkeleton(skel)
		if err != nil {
			return nil, fmt.Errorf("skeleton: %w", err)
		}
		m.Bones = bones
	}
	return m, nil
}

// decodeShape reads one SOBJ shape: its primitive sets (bone binding + index
// streams) and its interleaved vertex buffer.
func (c *CGFX) decodeShape(off int64) (*Shape, error) {
	S := off + 4 // magic
	if err := c.check(S, 0x40, "SOBJ header"); err != nil {
		return nil, err
	}
	if string(c.raw[S:S+4]) != "SOBJ" {
		return nil, fmt.Errorf("not a SOBJ (magic %q)", c.raw[S:S+4])
	}
	sh := &Shape{BoneIndex: -1}

	if obb := c.rel(S + 0x18); obb != 0 {
		for i := 0; i < 3; i++ {
			sh.OBBCenter[i] = c.f32(obb + 4 + int64(i)*4)
		}
	}

	// Primitive sets: bone table + primitives + index streams.
	psN, psT := int(c.u32(S+0x28)), c.rel(S+0x2C)
	for pi := 0; pi < psN; pi++ {
		ps := c.rel(psT + int64(pi)*4)
		boneN, boneT := int(c.u32(ps)), c.rel(ps+4)
		if boneN > 0 && boneT != 0 {
			sh.BoneIndex = int(c.u32(boneT)) // rigid: single-bone binding
			if boneN > 1 {
				return nil, fmt.Errorf("primitive set %d has %d bones; only rigid single-bone shapes are supported", pi, boneN)
			}
		}
		prN, prT := int(c.u32(ps+0xC)), c.rel(ps+0x10)
		for ri := 0; ri < prN; ri++ {
			pr := c.rel(prT + int64(ri)*4)
			isN, isT := int(c.u32(pr)), c.rel(pr+4)
			for ii := 0; ii < isN; ii++ {
				st := c.rel(isT + int64(ii)*4)
				if err := c.readIndexStream(st, sh); err != nil {
					return nil, err
				}
			}
		}
	}

	// Vertex buffers: the interleaved one carries the attributes; a Fixed
	// buffer (type 0x80000000) is a per-shape constant and is skipped.
	vbN, vbT := int(c.u32(S+0x34)), c.rel(S+0x38)
	for vi := 0; vi < vbN; vi++ {
		vb := c.rel(vbT + int64(vi)*4)
		switch ty := c.u32(vb); ty {
		case 0x40000002: // interleaved
			if err := c.readInterleaved(vb, sh); err != nil {
				return nil, err
			}
		case 0x80000000: // fixed (constant attribute) — no per-vertex data
		default:
			return nil, fmt.Errorf("vertex buffer %d has unsupported type 0x%08x", vi, ty)
		}
	}
	if len(sh.Verts) == 0 {
		return nil, fmt.Errorf("shape has no interleaved vertex buffer")
	}
	for _, ix := range sh.Indices {
		if int(ix) >= len(sh.Verts) {
			return nil, fmt.Errorf("index %d out of range (%d vertices)", ix, len(sh.Verts))
		}
	}
	return sh, nil
}

// readIndexStream appends one index stream's indices.
// Layout: glType @+0, size(bytes) @+8, data ptr @+0xC.
func (c *CGFX) readIndexStream(st int64, sh *Shape) error {
	ty := c.u32(st)
	size := int64(c.u32(st + 8))
	data := c.rel(st + 0xC)
	if err := c.check(data, size, "index stream"); err != nil {
		return err
	}
	switch ty {
	case glUShort:
		for p := data; p < data+size; p += 2 {
			sh.Indices = append(sh.Indices, uint32(binary.LittleEndian.Uint16(c.raw[p:])))
		}
	case glUByte:
		for p := data; p < data+size; p++ {
			sh.Indices = append(sh.Indices, uint32(c.raw[p]))
		}
	default:
		return fmt.Errorf("index stream has unsupported GL type 0x%04x", ty)
	}
	return nil
}

// readInterleaved decodes the interleaved vertex buffer.
// Layout: size @+0x14, data @+0x18, stride @+0x24, attrCount @+0x28,
// attrTable @+0x2C; each attribute: name @+4, glType @+0x24, components @+0x28,
// scale @+0x2C, byte offset within the stride @+0x30.
func (c *CGFX) readInterleaved(vb int64, sh *Shape) error {
	size := int64(c.u32(vb + 0x14))
	data := c.rel(vb + 0x18)
	stride := int64(c.u32(vb + 0x24))
	attrN, attrT := int(c.u32(vb+0x28)), c.rel(vb+0x2C)
	if stride == 0 || size%stride != 0 {
		return fmt.Errorf("vertex buffer size 0x%x is not a multiple of stride %d", size, stride)
	}
	if err := c.check(data, size, "vertex data"); err != nil {
		return err
	}
	n := int(size / stride)
	sh.Verts = make([]Vertex, n)

	for ai := 0; ai < attrN; ai++ {
		a := c.rel(attrT + int64(ai)*4)
		name := c.u32(a + 4)
		ty := c.u32(a + 0x24)
		comps := int(c.u32(a + 0x28))
		scale := c.f32(a + 0x2C)
		off := int64(c.u32(a + 0x30))

		for v := 0; v < n; v++ {
			base := data + int64(v)*stride + off
			var vals [4]float32
			for k := 0; k < comps; k++ {
				switch ty {
				case glFloat:
					vals[k] = c.f32(base + int64(k)*4)
				case glUByte:
					vals[k] = float32(c.raw[base+int64(k)]) * scale
				case glByte:
					vals[k] = float32(int8(c.raw[base+int64(k)])) * scale
				case glShort:
					vals[k] = float32(int16(binary.LittleEndian.Uint16(c.raw[base+int64(k)*2:]))) * scale
				case glUShort:
					vals[k] = float32(binary.LittleEndian.Uint16(c.raw[base+int64(k)*2:])) * scale
				default:
					return fmt.Errorf("attribute %d has unsupported GL type 0x%04x", ai, ty)
				}
			}
			switch name {
			case 0: // position
				copy(sh.Verts[v].Pos[:], vals[:3])
			case 1: // normal
				copy(sh.Verts[v].Normal[:], vals[:3])
			case 3: // color — scaled to [0,1]; store as RGBA8
				for k := 0; k < 4; k++ {
					f := vals[k]
					if f < 0 {
						f = 0
					}
					if f > 1 {
						f = 1
					}
					sh.Verts[v].Color[k] = uint8(f*255 + 0.5)
				}
			case 4:
				copy(sh.Verts[v].UV0[:], vals[:2])
			case 5:
				copy(sh.Verts[v].UV1[:], vals[:2])
			}
		}
		switch name {
		case 1:
			sh.HasNormal = true
		case 3:
			sh.HasColor = true
		case 4:
			if sh.UVCount < 1 {
				sh.UVCount = 1
			}
		case 5:
			sh.UVCount = 2
		}
	}
	return nil
}

// MTOB offsets, all from the object's typeId word and identical across the
// banner's four materials:
//
//	+0x168  u32  live texture-coordinator count (1 or 2 here)
//	+0x16C  the coordinator array, stride 0x58
//	+0x2F4  the mapper's sampler command entry, stride 0x8C
//	+0x33C  self-relative pointer to the mapper's texture name, stride 0x8C
//
// The count is the authority on how many mappers a material has: an MTOB always
// carries three name slots, and MarioMat's unused two point at *empty strings*
// rather than being null, so counting non-null pointers over-reports.
const (
	mtobCoordCount  = 0x168
	mtobCoordArray  = 0x16C
	mtobCoordStride = 0x58
	mtobSamplerCmd  = 0x2F4
	mtobTexName     = 0x33C
	mtobMapperStrid = 0x8C
)

// Texture-coordinator field offsets within one 0x58-byte entry.
const (
	coordSourceUV   = 0x00 // u32: which vertex UV set feeds this mapper
	coordMethod     = 0x04 // u32: mapping method (UV / cube / sphere env)
	coordScaleS     = 0x10
	coordScaleT     = 0x14
	coordRotate     = 0x18
	coordTransS     = 0x1C
	coordTransT     = 0x20
	coordMatrix     = 0x28 // f32[3][4], row-major
	coordMethodUV   = 0    // the only method this decoder handles
	coordMatrixRow  = 4    // floats per matrix row
	coordMatrixRows = 3
)

// picaBorderReg is the first register of each texture unit's config block; the
// unit's wrap/filter PARAM register is two further on. The three blocks are not
// evenly spaced in the register file (0x081, 0x091, 0x099), so they are named
// rather than computed.
var picaBorderReg = [3]uint16{0x081, 0x091, 0x099}

// decodeMaterial reads an MTOB into a Material: its name, and one TexMapper per
// live texture coordinator carrying the texture name, the source UV set, the
// coordinate transform and the sampler's wrap modes.
func (c *CGFX) decodeMaterial(e CGFXEntry) (*Material, error) {
	base := e.Offset // typeId word (magic at +4)
	if err := c.check(base, mtobTexName+3*mtobMapperStrid, "MTOB"); err != nil {
		return nil, err
	}
	if string(c.raw[base+4:base+8]) != "MTOB" {
		return nil, fmt.Errorf("not an MTOB (magic %q)", c.raw[base+4:base+8])
	}
	mat := &Material{Name: e.Name}

	n := int(c.u32(base + mtobCoordCount))
	if n > 3 {
		return nil, fmt.Errorf("material has %d texture coordinators; the PICA has 3 units", n)
	}
	for i := 0; i < n; i++ {
		m, err := c.decodeMapper(base, i)
		if err != nil {
			return nil, fmt.Errorf("mapper %d: %w", i, err)
		}
		mat.Mappers = append(mat.Mappers, *m)
	}
	return mat, nil
}

// decodeMapper reads one texture coordinator plus the texture name and sampler
// state that belong to the same mapper slot.
func (c *CGFX) decodeMapper(base int64, i int) (*TexMapper, error) {
	nameOff := c.rel(base + mtobTexName + int64(i)*mtobMapperStrid)
	if nameOff <= 0 || nameOff >= int64(len(c.raw)) {
		return nil, fmt.Errorf("texture name pointer is out of range")
	}
	name := readCStr(c.raw, nameOff)
	if name == "" {
		return nil, fmt.Errorf("live coordinator names no texture")
	}

	C := base + mtobCoordArray + int64(i)*mtobCoordStride
	if method := c.u32(C + coordMethod); method != coordMethodUV {
		return nil, fmt.Errorf("texture %q uses mapping method %d; only plain UV mapping is implemented", name, method)
	}
	m := &TexMapper{
		Texture:  name,
		SourceUV: int(c.u32(C + coordSourceUV)),
		ScaleS:   c.f32(C + coordScaleS),
		ScaleT:   c.f32(C + coordScaleT),
		Rotate:   c.f32(C + coordRotate),
		TransS:   c.f32(C + coordTransS),
		TransT:   c.f32(C + coordTransT),
	}
	// The matrix's third row must be the pass-through the 2-D reading assumes.
	row := func(r int) [4]float32 {
		var out [4]float32
		for k := 0; k < 4; k++ {
			out[k] = c.f32(C + coordMatrix + int64(r*coordMatrixRow+k)*4)
		}
		return out
	}
	if r2 := row(2); r2 != [4]float32{0, 0, 1, 0} {
		return nil, fmt.Errorf("texture %q has a non-planar texture matrix (third row %v)", name, r2)
	}
	for r := 0; r < 2; r++ {
		v := row(r)
		if v[2] != 0 {
			return nil, fmt.Errorf("texture %q's texture matrix uses the third coordinate (row %d = %v)", name, r, v)
		}
		m.Matrix[r] = [3]float32{v[0], v[1], v[3]}
	}

	ws, wt, err := c.samplerWrap(base, i)
	if err != nil {
		return nil, fmt.Errorf("texture %q: %w", name, err)
	}
	m.WrapS, m.WrapT = ws, wt
	return m, nil
}

// samplerWrap reads a mapper's wrap modes out of the PICA command entry the
// MTOB carries for that texture unit. The entry is a real command-buffer pair —
// parameter word, then a header writing the unit's border colour with nine
// consecutive extra parameters — so the unit's PARAM register (border + 2) is
// extra parameter 1. Reading the game's own register write beats inventing a
// struct field for it, and the header shape is checked so a layout change is a
// loud error rather than a plausible wrap mode.
func (c *CGFX) samplerWrap(base int64, i int) (PICAWrap, PICAWrap, error) {
	hdrOff := base + mtobSamplerCmd + int64(i)*mtobMapperStrid
	if err := c.check(hdrOff, 4+9*4, "sampler command entry"); err != nil {
		return 0, 0, err
	}
	hdr := c.u32(hdrOff)
	reg, mask, extras, consec := uint16(hdr&0xFFFF), hdr>>16&0xF, hdr>>20&0xFF, hdr>>31
	if reg != picaBorderReg[i] || mask != 0xF || extras != 9 || consec != 1 {
		return 0, 0, fmt.Errorf("sampler command header 0x%08x is not the expected write of 0x%03X+9", hdr, picaBorderReg[i])
	}
	param := c.u32(hdrOff + 4 + 1*4) // extra 1 writes register border+2 = PARAM
	return PICAWrap(param >> 12 & 7), PICAWrap(param >> 8 & 7), nil
}

// decodeSkeleton reads the bones dictionary.
// Skeleton: bone count @+0x18, bones DICT ptr @+0x1C (offsets from the typeId).
// Bone: name @+0, index @+8, parent index @+0xC, scale @+0x20, rotation @+0x2C,
// translation @+0x38.
func (c *CGFX) decodeSkeleton(off int64) ([]Bone, error) {
	if string(c.raw[off+4:off+8]) != "SOBJ" {
		return nil, fmt.Errorf("skeleton is not a SOBJ (magic %q)", c.raw[off+4:off+8])
	}
	dict := c.rel(off + 0x1C)
	entries, err := c.parseDict(dict)
	if err != nil {
		return nil, err
	}
	bones := make([]Bone, len(entries))
	for _, e := range entries {
		B := e.Offset
		idx := int(c.u32(B + 8))
		if idx < 0 || idx >= len(bones) {
			return nil, fmt.Errorf("bone %q has index %d out of %d", e.Name, idx, len(bones))
		}
		b := Bone{Name: e.Name, Index: idx, Parent: int(int32(c.u32(B + 0xC)))}
		for i := 0; i < 3; i++ {
			b.Scale[i] = c.f32(B + 0x20 + int64(i)*4)
			b.Rot[i] = c.f32(B + 0x2C + int64(i)*4)
			b.Trans[i] = c.f32(B + 0x38 + int64(i)*4)
		}
		bones[idx] = b
	}
	return bones, nil
}
