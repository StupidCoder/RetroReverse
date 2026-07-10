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

// Material is the slice of an MTOB this exporter needs: its name and the names
// of the textures its texture mappers reference, in mapper order.
type Material struct {
	Name     string
	Textures []string
}

// Bone is one skeleton joint. Rotation is the CGFX XYZ Euler triple.
type Bone struct {
	Name    string
	Index   int
	Parent  int // -1 for the root
	Scale   [3]float32
	Rot     [3]float32
	Trans   [3]float32
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

// decodeMaterial reads an MTOB's name and texture-mapper references. The three
// mapper slots' texture-name pointers sit at a fixed offset from the MTOB's
// typeId word (+0x33C, stride 0x8C) — verified identical across all four banner
// materials; a zero slot means no mapper.
func (c *CGFX) decodeMaterial(e CGFXEntry) (*Material, error) {
	base := e.Offset // typeId word (magic at +4)
	if string(c.raw[base+4:base+8]) != "MTOB" {
		return nil, fmt.Errorf("not an MTOB (magic %q)", c.raw[base+4:base+8])
	}
	mat := &Material{Name: e.Name}
	for i := 0; i < 3; i++ {
		slot := base + 0x33C + int64(i)*0x8C
		if slot+4 > int64(len(c.raw)) {
			break
		}
		nameOff := c.rel(slot)
		if nameOff == 0 || nameOff >= int64(len(c.raw)) {
			continue
		}
		name := readCStr(c.raw, nameOff)
		if name != "" {
			mat.Textures = append(mat.Textures, name)
		}
	}
	return mat, nil
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
