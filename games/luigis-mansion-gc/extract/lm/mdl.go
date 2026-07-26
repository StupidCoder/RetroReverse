package lm

// mdl.go reads the .mdl model format the opening-demo files carry (opwf_bg.mdl,
// opwf_luigi.mdl, ...), as reverse-engineered from the game's own renderer at
// 0x80058A00..0x8005A100 (found by read-watching the model in RAM while the
// cutscene drew it).
//
// The file is a header of counts and section pointers, followed by raw GX
// display lists, texture images, and the arrays the display lists index. The
// game patches the section offsets into pointers in place at load; on disc they
// are plain file offsets:
//
//	+0x04 u16 faceCount
//	+0x08 u16 nodeCount    +0x0c u16 weightedCount (blended matrices)
//	+0x10 u16 posCount     +0x12 u16 nrmCount
//	+0x14 u16 clrCount     +0x16 u16 texCount
//	+0x20 u16 texHdrCount  +0x24 u16 samplerCount  +0x28 u16 materialCount
//	+0x2a u16 shapeCount (= draw-pair count in both opening models)
//	+0x30 nodes    16 B: {s16 id, s16 sibling, s16 child, u16 mode,
//	                      u16 pairCount, u16 firstPair}  (sibling/child relative)
//	+0x34 packets  32 B: {u32 dlOffset, u32 dlSize, u16 -, u16 mtxCount,
//	                      s16 mtxIdx[10]}  (mtxIdx i loads GX PN-matrix slot 3i)
//	+0x38 matrices 48 B: 3x4 float32 row-major per node
//	+0x3c weight values   f32     (envelope streams, indexed by running count)
//	+0x40 weight joints   u16
//	+0x44 weight counts   u8 per weighted matrix
//	+0x48 positions f32[3]  +0x4c normals f32[3]  +0x50 colours rgba8
//	+0x54 texcoords f32[2]
//	+0x60 texture-header pointer table (u32 each)
//	       texture header: {u8 fmtIdx, u8 -, u16 w, u16 h; image at +0x20};
//	       fmtIdx maps through the game's table at 0x80338C30 to a GX format
//	+0x68 materials 0x120 B: {u8[4] -, u8 tint rgba? @4..7, u8 flags@8, u8 stageCount@7,
//	       32-B stage entries from +0x20 with s16 samplerIdx at entry+2}
//	+0x6c samplers 8 B: {u16 texHdrIdx, u16 -, u8 wrapS, u8 wrapT}
//	+0x70 shapes   8 B: {u32 flags (bit 0x02000000 NBT, bit 0x01000000 lit),
//	                     u16 packetCount, u16 firstPacket}
//	+0x74 draw pairs {u16 materialIdx, u16 shapeIdx}
//
// A display list is a GX command stream: opcode byte (primitive | vat), u16
// vertex count, then per vertex a u8 PN-matrix slot plus one u16 index per
// enabled attribute in attribute order — position, normal (3 entries for NBT),
// colour, texcoord — an attribute is enabled when its array count is nonzero.

import (
	"encoding/binary"
	"fmt"
)

// Mtx34 is a GC 3x4 row-major matrix (rows of x-basis/y-basis/z-basis with the
// translation in the fourth column).
type Mtx34 [12]float32

// Identity34 is the identity Mtx34.
var Identity34 = Mtx34{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0}

// Mul multiplies two 3x4 matrices (treating both as 4x4 with a 0,0,0,1 row).
func (a Mtx34) Mul(b Mtx34) Mtx34 {
	var r Mtx34
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			s := a[i*4]*b[j] + a[i*4+1]*b[4+j] + a[i*4+2]*b[8+j]
			if j == 3 {
				s += a[i*4+3]
			}
			r[i*4+j] = s
		}
	}
	return r
}

// Apply transforms a point.
func (a Mtx34) Apply(v [3]float32) [3]float32 {
	return [3]float32{
		a[0]*v[0] + a[1]*v[1] + a[2]*v[2] + a[3],
		a[4]*v[0] + a[5]*v[1] + a[6]*v[2] + a[7],
		a[8]*v[0] + a[9]*v[1] + a[10]*v[2] + a[11],
	}
}

// ApplyVec transforms a direction (no translation).
func (a Mtx34) ApplyVec(v [3]float32) [3]float32 {
	return [3]float32{
		a[0]*v[0] + a[1]*v[1] + a[2]*v[2],
		a[4]*v[0] + a[5]*v[1] + a[6]*v[2],
		a[8]*v[0] + a[9]*v[1] + a[10]*v[2],
	}
}

// MDLNode is one node of the model's tree; Sibling and Child are indices into
// Nodes (-1 for none).
type MDLNode struct {
	ID      int
	Sibling int
	Child   int
	Mode    uint16
	Pairs   []MDLPair // the (material, shape) pairs this node draws
	Mtx     Mtx34     // the node's matrix from the file
}

// MDLPair draws one shape with one material.
type MDLPair struct{ Material, Shape int }

// MDLPacket is a slice of a shape: a display list plus the matrix table it
// expects loaded. MtxIdx[i] is the node (or nodeCount+envelope) matrix loaded
// into GX PN slot 3*i; -1 keeps the previous load.
type MDLPacket struct {
	DL     []byte
	MtxIdx []int16
}

// MDLShape is a run of packets sharing one vertex descriptor.
type MDLShape struct {
	Flags   uint32
	Packets []MDLPacket
}

// MDLSampler binds a texture with wrap modes (GX enum: 0 clamp, 1 repeat, 2 mirror).
type MDLSampler struct {
	Texture      int
	WrapS, WrapT uint8
}

// MDLTexture is a decoded texture image.
type MDLTexture struct {
	Format        uint8 // the file's format index (0..10)
	Width, Height int
	Pixels        []byte // RGBA, Width*Height*4
}

// MDLMaterial keeps what the exporter needs of the 0x120-byte material record.
type MDLMaterial struct {
	Tint     [4]uint8
	Flags    uint8
	Samplers []int // sampler index per texture stage
}

// MDLEnvelope is one blended matrix: vertices bound to it blend the listed
// joints with the listed weights.
type MDLEnvelope struct {
	Joints  []uint16
	Weights []float32
}

// MDL is a parsed .mdl model.
type MDL struct {
	FaceCount int
	Nodes     []MDLNode
	Shapes    []MDLShape
	Envelopes []MDLEnvelope
	Positions [][3]float32
	Normals   [][3]float32
	Colors    [][4]uint8
	Texcoords [][2]float32
	Textures  []MDLTexture
	Samplers  []MDLSampler
	Materials []MDLMaterial
}

// ParseMDL reads a .mdl file.
func ParseMDL(b []byte) (*MDL, error) {
	if len(b) < 0x80 {
		return nil, fmt.Errorf("lm: mdl too short")
	}
	u16 := func(off uint32) int { return int(binary.BigEndian.Uint16(b[off:])) }
	u32 := func(off uint32) uint32 { return binary.BigEndian.Uint32(b[off:]) }
	f32 := func(off uint32) float32 {
		return float32frombits(binary.BigEndian.Uint32(b[off:]))
	}

	m := &MDL{FaceCount: u16(0x04)}
	nodeCount := u16(0x08)
	weightedCount := u16(0x0c)
	posCount, nrmCount := u16(0x10), u16(0x12)
	clrCount, texCount := u16(0x14), u16(0x16)
	texHdrCount, samplerCount := u16(0x20), u16(0x24)
	materialCount := u16(0x28)
	pairCount := u16(0x2a)

	nodesOff := u32(0x30)
	packetsOff := u32(0x34)
	mtxOff := u32(0x38)
	weightValOff := u32(0x3c)
	weightJointOff := u32(0x40)
	weightCntOff := u32(0x44)
	posOff := u32(0x48)
	nrmOff := u32(0x4c)
	clrOff := u32(0x50)
	texOff := u32(0x54)
	texTabOff := u32(0x60)
	matOff := u32(0x68)
	samplerOff := u32(0x6c)
	shapeOff := u32(0x70)
	pairOff := u32(0x74)

	// Vertex arrays.
	m.Positions = make([][3]float32, posCount)
	for i := range m.Positions {
		o := posOff + uint32(i)*12
		m.Positions[i] = [3]float32{f32(o), f32(o + 4), f32(o + 8)}
	}
	m.Normals = make([][3]float32, nrmCount)
	for i := range m.Normals {
		o := nrmOff + uint32(i)*12
		m.Normals[i] = [3]float32{f32(o), f32(o + 4), f32(o + 8)}
	}
	m.Colors = make([][4]uint8, clrCount)
	for i := range m.Colors {
		o := clrOff + uint32(i)*4
		m.Colors[i] = [4]uint8{b[o], b[o+1], b[o+2], b[o+3]}
	}
	m.Texcoords = make([][2]float32, texCount)
	for i := range m.Texcoords {
		o := texOff + uint32(i)*8
		m.Texcoords[i] = [2]float32{f32(o), f32(o + 4)}
	}

	// Nodes and their draw pairs.
	pairs := make([]MDLPair, pairCount)
	for i := range pairs {
		o := pairOff + uint32(i)*4
		pairs[i] = MDLPair{Material: u16(o), Shape: u16(o + 2)}
	}
	m.Nodes = make([]MDLNode, nodeCount)
	for i := range m.Nodes {
		o := nodesOff + uint32(i)*16
		n := &m.Nodes[i]
		n.ID = int(int16(u16(o)))
		// Verified against the game's runtime world-matrix array: +2 is the
		// child offset and +4 the sibling offset, both relative node counts.
		if c := int(int16(u16(o + 2))); c != 0 {
			n.Child = i + c
		} else {
			n.Child = -1
		}
		if s := int(int16(u16(o + 4))); s != 0 {
			n.Sibling = i + s
		} else {
			n.Sibling = -1
		}
		n.Mode = uint16(u16(o + 6))
		cnt, first := u16(o+8), u16(o+10)
		if first+cnt <= len(pairs) {
			n.Pairs = pairs[first : first+cnt]
		}
		mo := mtxOff + uint32(i)*48
		for j := 0; j < 12; j++ {
			n.Mtx[j] = f32(mo + uint32(j)*4)
		}
	}

	// Envelopes (weighted matrices).
	m.Envelopes = make([]MDLEnvelope, weightedCount)
	run := uint32(0)
	for i := range m.Envelopes {
		cnt := uint32(b[weightCntOff+uint32(i)])
		e := &m.Envelopes[i]
		for j := uint32(0); j < cnt; j++ {
			e.Joints = append(e.Joints, binary.BigEndian.Uint16(b[weightJointOff+(run+j)*2:]))
			e.Weights = append(e.Weights, f32(weightValOff+(run+j)*4))
		}
		run += cnt
	}

	// Shapes and packets. The header's +0x2a is the pair count; the shape
	// array's own length is bounded by the highest shape the pairs reference.
	m.Shapes = make([]MDLShape, shapeSectionCount(pairs))
	for i := range m.Shapes {
		o := shapeOff + uint32(i)*8
		s := &m.Shapes[i]
		s.Flags = u32(o)
		pktCnt, first := u16(o+4), u16(o+6)
		for p := 0; p < pktCnt; p++ {
			po := packetsOff + uint32(first+p)*32
			dlOff, dlSize := u32(po), u32(po+4)
			if dlOff+dlSize > uint32(len(b)) {
				return nil, fmt.Errorf("lm: packet %d display list out of range", first+p)
			}
			pk := MDLPacket{DL: b[dlOff : dlOff+dlSize]}
			mtxCount := u16(po + 10)
			for k := 0; k < mtxCount; k++ {
				pk.MtxIdx = append(pk.MtxIdx, int16(u16(po+12+uint32(k)*2)))
			}
			m.Shapes[i].Packets = append(s.Packets, pk)
		}
	}

	// Textures, samplers, materials.
	for i := 0; i < texHdrCount; i++ {
		p := u32(texTabOff + uint32(i)*4)
		if p+0x20 > uint32(len(b)) {
			return nil, fmt.Errorf("lm: texture header %d out of range", i)
		}
		fmtIdx := b[p]
		w, h := u16(p+2), u16(p+4)
		img, err := decodeGXTexture(fmtIdx, w, h, b[p+0x20:])
		if err != nil {
			return nil, fmt.Errorf("lm: texture %d: %w", i, err)
		}
		m.Textures = append(m.Textures, MDLTexture{Format: fmtIdx, Width: w, Height: h, Pixels: img})
	}
	for i := 0; i < samplerCount; i++ {
		o := samplerOff + uint32(i)*8
		m.Samplers = append(m.Samplers, MDLSampler{Texture: u16(o), WrapS: b[o+4], WrapT: b[o+5]})
	}
	for i := 0; i < materialCount; i++ {
		o := matOff + uint32(i)*0x120
		mat := MDLMaterial{
			Tint:  [4]uint8{b[o+4], b[o+5], b[o+6], b[o+7]},
			Flags: b[o+8],
		}
		stageCount := int(b[o+7])
		if stageCount > 8 {
			stageCount = 8
		}
		for s := 0; s < stageCount; s++ {
			so := o + 0x20 + uint32(s)*0x20
			idx := int(int16(u16(so + 2)))
			if idx >= 0 && idx < samplerCount {
				mat.Samplers = append(mat.Samplers, idx)
			}
		}
		m.Materials = append(m.Materials, mat)
	}
	return m, nil
}

// shapeSectionCount bounds the shape array by the highest shape index the draw
// pairs actually reference.
func shapeSectionCount(pairs []MDLPair) int {
	max := 0
	for _, p := range pairs {
		if p.Shape+1 > max {
			max = p.Shape + 1
		}
	}
	return max
}

// DLVertex is one vertex fetched from a display list: indices into the model's
// arrays, and the PN-matrix slot it uses (slot/3 indexes the packet's MtxIdx).
type DLVertex struct {
	MtxSlot uint8
	Pos     uint16
	Nrm     uint16
	Clr     uint16
	Tex     uint16
	HasNrm  bool
	HasClr  bool
	HasTex  bool
}

// DLPrimitive is one GX primitive from a display list.
type DLPrimitive struct {
	Kind  uint8 // GX opcode top bits: 0x80 quads, 0x90 tris, 0x98 strip, 0xA0 fan
	Verts []DLVertex
}

// ParseDL walks a packet's display list given the model's attribute presence.
// nbt selects three consecutive normal indices per vertex (the shape flag).
func (m *MDL) ParseDL(dl []byte, nbt bool) ([]DLPrimitive, error) {
	hasNrm := len(m.Normals) > 0
	hasClr := len(m.Colors) > 0
	hasTex := len(m.Texcoords) > 0
	var prims []DLPrimitive
	i := 0
	for i < len(dl) {
		op := dl[i]
		if op == 0 { // NOP padding at the tail
			i++
			continue
		}
		kind := op & 0xF8
		if kind < 0x80 || kind > 0xB8 {
			return nil, fmt.Errorf("lm: display-list opcode %#x at %d", op, i)
		}
		if i+3 > len(dl) {
			return nil, fmt.Errorf("lm: display list truncated at %d", i)
		}
		count := int(binary.BigEndian.Uint16(dl[i+1:]))
		i += 3
		p := DLPrimitive{Kind: kind}
		for v := 0; v < count; v++ {
			var vx DLVertex
			// Three direct matrix bytes: the PN-matrix slot and two texture-
			// matrix slots (GX attrs 7/8 on static models, 1/2 on skinned ones —
			// the renderer enables one pair or the other, but both are one byte).
			vx.MtxSlot = dl[i]
			i += 3
			vx.Pos = binary.BigEndian.Uint16(dl[i:])
			i += 2
			if hasNrm {
				vx.HasNrm = true
				vx.Nrm = binary.BigEndian.Uint16(dl[i:])
				i += 2
				if nbt {
					i += 4 // tangent and binormal indices follow
				}
			}
			if hasClr {
				vx.HasClr = true
				vx.Clr = binary.BigEndian.Uint16(dl[i:])
				i += 2
			}
			if hasTex {
				vx.HasTex = true
				vx.Tex = binary.BigEndian.Uint16(dl[i:])
				i += 2
			}
			p.Verts = append(p.Verts, vx)
		}
		prims = append(prims, p)
	}
	return prims, nil
}

// Triangulate converts a primitive to triangle index triples (indices into
// p.Verts), honouring GX winding.
func (p *DLPrimitive) Triangulate() [][3]int {
	var tris [][3]int
	n := len(p.Verts)
	switch p.Kind {
	case 0x80: // quads
		for i := 0; i+3 < n; i += 4 {
			tris = append(tris, [3]int{i, i + 1, i + 2}, [3]int{i, i + 2, i + 3})
		}
	case 0x90: // triangles
		for i := 0; i+2 < n; i += 3 {
			tris = append(tris, [3]int{i, i + 1, i + 2})
		}
	case 0x98: // strip
		for i := 0; i+2 < n; i++ {
			if i%2 == 0 {
				tris = append(tris, [3]int{i, i + 1, i + 2})
			} else {
				tris = append(tris, [3]int{i, i + 2, i + 1})
			}
		}
	case 0xA0: // fan
		for i := 1; i+1 < n; i++ {
			tris = append(tris, [3]int{0, i, i + 1})
		}
	}
	return tris
}
