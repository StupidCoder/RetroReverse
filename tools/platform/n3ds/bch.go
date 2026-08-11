package n3ds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"math"
)

// bch.go reads BCH ("Binary CTR H3D"), the model container this platform's
// *game* assets use — as opposed to the CGFX (cgfx.go) that its HOME Menu
// banner uses. Everything under /ObjectData and /StageData that draws anything
// is a `.bch` inside a SARC inside a Yaz0 stream.
//
// The file is six sections, and the header gives each an offset and a length
// that tile the file exactly — main header, string table, GPU command words,
// data, extended data, relocation table:
//
//	0x00 "BCH\0"  u8 backwardCompat  u8 forwardCompat  u16 version
//	0x08 u32 mainHeaderOffset   0x20 u32 mainHeaderLength
//	0x0C u32 stringTableOffset  0x24 u32 stringTableLength
//	0x10 u32 gpuCommandsOffset  0x28 u32 gpuCommandsLength
//	0x14 u32 dataOffset         0x2C u32 dataLength
//	0x18 u32 dataExtOffset      0x30 u32 dataExtLength
//	0x1C u32 relocationOffset   0x34 u32 relocationLength
//	0x40 u16 flags              0x42 u16 addressCount
//
// The main header is fifteen (pointerTable, count, dictionary) triples, one per
// content type, in a fixed order; a dictionary is a patricia tree whose nodes a
// linear walk enumerates, exactly like CGFX's.
//
// **A mesh is not a struct describing geometry — it is a GPU command list.**
// The vertex format, the buffer offsets, the index array and the draw call are
// all stored as PICA register writes, in the same encoding the game submits at
// run time, and this decoder reads them with the same DecodePICA the emulated
// GPU uses. That is why so little of this file is layout guessing: the parts
// that matter are the hardware's own registers, and gpu_raster.go already says
// what they mean.
const bchMagic = "BCH\x00"

// BCH content groups, in main-header order.
const (
	BCHModels = iota
	BCHMaterials
	BCHShaders
	BCHTextures
	BCHMaterialLUTs
	BCHLights
	BCHCameras
	BCHFogs
	BCHSkeletalAnims
	BCHMaterialAnims
	BCHVisibilityAnims
	BCHLightAnims
	BCHCameraAnims
	BCHFogAnims
	BCHScenes
	bchGroupCount
)

// BCHEntry is one named object of a content group.
type BCHEntry struct {
	Name   string
	Offset int64 // absolute file offset of the object
}

// BCH is a parsed container.
type BCH struct {
	raw    []byte
	main   int64
	str    int64
	gpu    int64
	data   int64
	dataX  int64
	reloc  int64
	relocN int64

	Groups [bchGroupCount][]BCHEntry
}

func (f *BCH) u32(o int64) uint32  { return binary.LittleEndian.Uint32(f.raw[o:]) }
func (f *BCH) u16(o int64) uint16  { return binary.LittleEndian.Uint16(f.raw[o:]) }
func (f *BCH) f32(o int64) float32 { return math.Float32frombits(f.u32(o)) }
func (f *BCH) inRange(o, n int64) bool {
	return o >= 0 && n >= 0 && o+n <= int64(len(f.raw))
}

// ParseBCH parses a BCH container's structure.
func ParseBCH(b []byte) (*BCH, error) {
	if len(b) < 0x44 || string(b[:4]) != bchMagic {
		return nil, fmt.Errorf("bch: bad magic")
	}
	f := &BCH{raw: b}
	f.main = int64(binary.LittleEndian.Uint32(b[0x08:]))
	f.str = int64(binary.LittleEndian.Uint32(b[0x0C:]))
	f.gpu = int64(binary.LittleEndian.Uint32(b[0x10:]))
	f.data = int64(binary.LittleEndian.Uint32(b[0x14:]))
	f.dataX = int64(binary.LittleEndian.Uint32(b[0x18:]))
	f.reloc = int64(binary.LittleEndian.Uint32(b[0x1C:]))
	f.relocN = int64(binary.LittleEndian.Uint32(b[0x34:])) / 4

	// The sections tile the file: the last one ends exactly at the file's end.
	// That is the one arithmetic check the header cannot fake, and it is what
	// pins every offset above.
	if end := f.reloc + f.relocN*4; end != int64(len(b)) {
		return nil, fmt.Errorf("bch: relocation table ends at 0x%x, file is 0x%x", end, len(b))
	}
	if !f.inRange(f.main, int64(binary.LittleEndian.Uint32(b[0x20:]))) {
		return nil, fmt.Errorf("bch: main header runs outside the file")
	}

	for g := 0; g < bchGroupCount; g++ {
		p := f.main + int64(g)*12
		tbl, n, dict := int64(f.u32(p)), int64(f.u32(p+4)), int64(f.u32(p+8))
		if n == 0 {
			continue
		}
		if !f.inRange(f.main+tbl, n*4) {
			return nil, fmt.Errorf("bch: group %d pointer table runs outside the file", g)
		}
		es := make([]BCHEntry, n)
		for i := int64(0); i < n; i++ {
			// A dictionary is a root node then one 12-byte node per entry:
			// {u32 referenceBit, u16 left, u16 right, u32 nameOffset}.
			node := f.main + dict + 12 + i*12
			if !f.inRange(node, 12) {
				return nil, fmt.Errorf("bch: group %d dictionary node %d runs outside the file", g, i)
			}
			es[i] = BCHEntry{
				Name:   readCStr(b, f.str+int64(f.u32(node+8))),
				Offset: f.main + int64(f.u32(f.main+tbl+i*4)),
			}
		}
		f.Groups[g] = es
	}
	return f, nil
}

// relocType returns the relocation type of the entry that patches a given word
// index, or -1 if the word is not relocated. The type is not decoration: for
// the index-array pointer it is the *only* place the index width is recorded.
func (f *BCH) relocType(word uint32) int {
	for i := int64(0); i < f.relocN; i++ {
		v := f.u32(f.reloc + i*4)
		if v&0xFFFFFF == word {
			return int(v >> 24)
		}
	}
	return -1
}

// Relocation types this decoder acts on. Each face's index-array pointer is
// relocated with one of two types, and *that* is what says how wide its indices
// are — PICA's own format bit (register 0x227 bit 31) is zero in the file,
// because the address it sits beside has not been relocated yet.
const (
	bchRelocIndex16 = 0x58
	bchRelocIndex8  = 0x5A
)

// BCHVertex is one de-interleaved vertex. Which fields are meaningful is given
// by the shape's Has* flags.
type BCHVertex struct {
	Pos    [3]float32
	Normal [3]float32
	Color  [4]uint8
	UV     [3][2]float32

	// Joints index the mesh's Palette; Weights sum to 1 over the influences the
	// vertex actually has.
	Joints  [4]uint8
	Weights [4]float32
}

// BCHBone is one joint of a model's skeleton. The rotation is an XYZ Euler
// triple in radians; InvBind is the 3x4 matrix taking a vertex from the model's
// space into the bone's.
type BCHBone struct {
	Name    string
	Parent  int // -1 for the root
	Scale   [3]float32
	Rotate  [3]float32
	Trans   [3]float32
	InvBind [3][4]float32
}

// BCHMesh is one drawable: a triangle list over its own vertices, bound to one
// of the model's materials.
type BCHMesh struct {
	MaterialIndex int
	Verts         []BCHVertex
	Indices       []uint32

	// Palette maps this mesh's own bone indices onto the model's skeleton. A
	// skinned vertex does not name a bone directly: it names a slot in the
	// small per-draw matrix palette the hardware holds, and the palette names
	// the bone. Empty when the mesh is not skinned.
	Palette []int

	// SkinMode is the face header's own statement of how the mesh is bound:
	// BCHSkinSmooth (per-vertex indices and weights), BCHSkinRigid (an index
	// per vertex, full weight) or BCHSkinNone.
	SkinMode int

	HasNormal bool
	HasColor  bool
	HasSkin   bool
	UVCount   int

	// ArrayOffset is where this mesh's vertices start in the extended data
	// section, and VertexStride how far apart they sit. They are kept because
	// they are checkable: a model's meshes tile that section end to end, so
	// ArrayOffset + len(Verts)*VertexStride reaches the next mesh's offset. A
	// wrong stride, a wrong index width or a wrong array offset all break that
	// chain, which makes it the decode's own proof.
	ArrayOffset  int64
	VertexStride int64
}

// BCHModel is a decoded model object.
type BCHModel struct {
	Name      string
	Transform [3][4]float32 // the model's world matrix, row-major
	Meshes    []BCHMesh
	Materials []BCHMaterial // in the model's own material order
	Bones     []BCHBone     // the skeleton, empty for static geometry
}

// Model-object field offsets, all from the object's start. The pointers are
// relative to the main header, like every pointer in that section.
const (
	bchModelTransform = 0x04 // f32[3][4]
	bchModelMatTable  = 0x34
	bchModelMatCount  = 0x38
	bchModelMatDict   = 0x3C
	bchModelMeshTable = 0x40
	bchModelMeshCount = 0x44
	bchModelBoneTable = 0x70
	bchModelBoneCount = 0x74
	bchMeshStride     = 0x38

	bchBoneStride  = 0x64
	bchBoneFlags   = 0x00
	bchBoneParent  = 0x04 // int16
	bchBoneScale   = 0x08
	bchBoneRotate  = 0x14 // XYZ Euler, radians
	bchBoneTrans   = 0x20
	bchBoneInvBind = 0x2C // f32[3][4]
	bchBoneName    = 0x5C
)

// A face's matrix palette sits at the head of the face structure: a skinning
// mode, a bone count, and then that many bone indices, all u16. The reserved
// space is twenty u16 — what the hardware's palette holds — and the entries
// past the count are its unused tail.
//
// The count is not decoration, and reading the whole reserved block as bones
// gets every vertex's joint wrong by two slots. The game says so itself: for
// each of Captain Toad's skinned draws the vertex shader is handed exactly
// *count* three-row matrices, starting at uniform c25, in this table's order —
// Toad's rucksack (count 6) uploads c25-c42, his body (count 4) reuses the
// first four of them, and his scarf (count 3, bones 8/14/18) is handed the
// rucksack's slots 1-3 verbatim.
const (
	bchFaceSkinMode  = 0x00 // u16: 1 = smooth (per-vertex weights), 2 = rigid
	bchFaceBoneCount = 0x02 // u16
	bchFacePalette   = 0x04
	bchPaletteMax    = 18 // the reserved block is 20 u16, two of which are the above
)

// Skinning modes, as the face header states them.
const (
	BCHSkinNone   = 0 // not skinned: the mesh's vertices are the model's own
	BCHSkinSmooth = 1 // per-vertex bone indices and weights
	BCHSkinRigid  = 2 // per-vertex bone index, full weight
)

// Mesh field offsets.
const (
	bchMeshMaterial = 0x00 // u16
	bchMeshVertCmds = 0x08 // (offset, word count) into the GPU command section
	bchMeshVertLen  = 0x0C
	bchMeshFaces    = 0x10 // (offset into the main header, count)
	bchMeshFaceN    = 0x14
	bchFaceStride   = 0xA0
	bchFaceCmds     = 0x2C // (offset, word count) into the GPU command section
	bchFaceCmdLen   = 0x30
)

// The two vertex-loader registers gpu.go does not already name. Everything else
// this decoder reads (regAttrBase, regAttrFmtLow/High, regIndexConfig,
// regNumVertices, regVshAttrPermL/H) is declared there, and means the same
// thing here as it does when the game submits it at run time.
const (
	regLoaderOff  = 0x203 // loader 0: a *byte* offset from the attribute base
	regLoaderComp = 0x204 // its component list: 4 bits per slot, naming an attribute
	regLoaderCfg  = 0x205 // component list high half, stride in 16-23, slot count in 28-31
)

// Shader input registers, whose index *is* the attribute's meaning: the H3D
// vertex shader takes position in v0, normal in v1, tangent in v2, colour in v3
// and the texture coordinate sets in v4-v6. The attribute→input permutation
// (0x2BB/0x2BC) is therefore the file's own statement of what each attribute is,
// and no attribute order has to be assumed.
const (
	bchInPosition   = 0
	bchInNormal     = 1
	bchInTangent    = 2
	bchInColor      = 3
	bchInTexCoord   = 4 // ...through 6
	bchInBoneIndex  = 7
	bchInBoneWeight = 8
)

// attrScale undoes attrValue's normalisation for attributes that are counts
// rather than fractions: a bone *index* wants the raw byte back.
func attrScale(typ int) float32 {
	if typ == 1 {
		return 255
	}
	return 1
}

// weightScale turns a stored bone weight into a fraction. They are unsigned
// bytes, and they are **percentages** — a two-influence vertex stores 70 and 30,
// not 179 and 76. attrValue has already divided by 255, so the remaining factor
// puts them over 100 instead. The rule is checkable and checked: every skinned
// vertex's weights sum to one.
func weightScale(typ int) float32 {
	if typ == 1 {
		return 2.55 // 255/100
	}
	return 1
}

// DecodeModel decodes the model a Models-group entry points at.
func (f *BCH) DecodeModel(e2 BCHEntry) (*BCHModel, error) {
	e := e2
	M := e.Offset
	if !f.inRange(M, 0x48) {
		return nil, fmt.Errorf("bch: model %q header runs outside the file", e.Name)
	}
	m := &BCHModel{Name: e.Name}
	for r := 0; r < 3; r++ {
		for c := 0; c < 4; c++ {
			m.Transform[r][c] = f.f32(M + bchModelTransform + int64(r*4+c)*4)
		}
	}

	matN := int64(f.u32(M + bchModelMatCount))
	matDict := int64(f.u32(M + bchModelMatDict))
	bind := f.main + int64(f.u32(M+bchModelMatTable))
	for i := int64(0); i < matN; i++ {
		node := f.main + matDict + 12 + i*12
		be := bind + i*bchMatBindStride
		if !f.inRange(node, 12) || !f.inRange(be, bchMatBindStride) {
			return nil, fmt.Errorf("bch: model %q material %d runs outside the file", e.Name, i)
		}
		mat := BCHMaterial{Name: readCStr(f.raw, f.str+int64(f.u32(node+8)))}
		obj := f.main + int64(f.u32(be+bchMatBindObject))
		if !f.inRange(obj, bchMatCmdLen+4) {
			return nil, fmt.Errorf("bch: model %q material %d object runs outside the file", e.Name, i)
		}
		if err := f.decodeMaterialState(obj, &mat); err != nil {
			return nil, fmt.Errorf("bch: model %q: %w", e.Name, err)
		}
		for u := 0; u < 4; u++ {
			// An unset slot is a zero word, and zero is a valid string-table
			// offset — it lands on the table's first entry. Resolving it would
			// give every textureless material the same plausible-looking name,
			// so an unset slot stays empty.
			if w := f.u32(be + bchMatBindTex0 + int64(u)*4); w != 0 {
				mat.Names[u] = readCStr(f.raw, f.str+int64(w))
			}
		}
		m.Materials = append(m.Materials, mat)
	}

	boneT := f.main + int64(f.u32(M+bchModelBoneTable))
	boneN := int64(f.u32(M + bchModelBoneCount))
	for i := int64(0); i < boneN; i++ {
		e := boneT + i*bchBoneStride
		if !f.inRange(e, bchBoneStride) {
			return nil, fmt.Errorf("bch: model %q bone %d runs outside the file", e2.Name, i)
		}
		b := BCHBone{
			Name:   readCStr(f.raw, f.str+int64(f.u32(e+bchBoneName))),
			Parent: int(int16(f.u16(e + bchBoneParent))),
		}
		for k := 0; k < 3; k++ {
			b.Scale[k] = f.f32(e + bchBoneScale + int64(k)*4)
			b.Rotate[k] = f.f32(e + bchBoneRotate + int64(k)*4)
			b.Trans[k] = f.f32(e + bchBoneTrans + int64(k)*4)
		}
		for r := 0; r < 3; r++ {
			for c := 0; c < 4; c++ {
				b.InvBind[r][c] = f.f32(e + bchBoneInvBind + int64(r*4+c)*4)
			}
		}
		if b.Parent >= int(boneN) {
			return nil, fmt.Errorf("bch: model %q bone %q has parent %d of %d", e2.Name, b.Name, b.Parent, boneN)
		}
		m.Bones = append(m.Bones, b)
	}

	meshT := f.main + int64(f.u32(M+bchModelMeshTable))
	meshN := int64(f.u32(M + bchModelMeshCount))
	for i := int64(0); i < meshN; i++ {
		sh, err := f.decodeMesh(meshT + i*bchMeshStride)
		if err != nil {
			return nil, fmt.Errorf("bch: model %q mesh %d: %w", e.Name, i, err)
		}
		m.Meshes = append(m.Meshes, *sh)
	}
	return m, nil
}

// decodeMesh runs one mesh's command lists through a bare register file and
// reads the geometry the registers describe.
func (f *BCH) decodeMesh(e int64) (*BCHMesh, error) {
	if !f.inRange(e, bchMeshStride) {
		return nil, fmt.Errorf("mesh header runs outside the file")
	}
	sh := &BCHMesh{MaterialIndex: int(f.u16(e + bchMeshMaterial))}

	var regs [0x300]uint32
	run := func(off, words int64) error {
		if words == 0 {
			return nil
		}
		if !f.inRange(f.gpu+off, words*4) {
			return fmt.Errorf("command list at gpu+0x%x (%d words) runs outside the file", off, words)
		}
		ws, err := DecodePICA(f.raw[f.gpu+off : f.gpu+off+words*4])
		if err != nil {
			return err
		}
		for _, w := range ws {
			if w.Reg < 0x300 {
				regs[w.Reg] = w.Value
			}
		}
		return nil
	}
	if err := run(int64(f.u32(e+bchMeshVertCmds)), int64(f.u32(e+bchMeshVertLen))); err != nil {
		return nil, err
	}

	faces := f.main + int64(f.u32(e+bchMeshFaces))
	faceN := int64(f.u32(e + bchMeshFaceN))
	for i := int64(0); i < faceN; i++ {
		fe := faces + i*bchFaceStride
		if !f.inRange(fe, bchFaceStride) {
			return nil, fmt.Errorf("face %d runs outside the file", i)
		}
		// The face's matrix palette: the header says how long it is, so the
		// zeros past the end stay out of it rather than becoming bone 0.
		if i == 0 {
			sh.SkinMode = int(f.u16(fe + bchFaceSkinMode))
			n := int64(f.u16(fe + bchFaceBoneCount))
			if n > bchPaletteMax {
				return nil, fmt.Errorf("face %d names %d palette bones, more than the %d the block holds",
					i, n, bchPaletteMax)
			}
			switch sh.SkinMode {
			case BCHSkinNone, BCHSkinSmooth, BCHSkinRigid:
			default:
				return nil, fmt.Errorf("face %d has skinning mode %d, which is not one this decoder models", i, sh.SkinMode)
			}
			for p := int64(0); p < n; p++ {
				sh.Palette = append(sh.Palette, int(f.u16(fe+bchFacePalette+p*2)))
			}
		}
		cmds, words := int64(f.u32(fe+bchFaceCmds)), int64(f.u32(fe+bchFaceCmdLen))
		if err := run(cmds, words); err != nil {
			return nil, err
		}
		// The index-array pointer is the third entry of the face's list; its
		// relocation type carries the index width.
		idxWord := uint32((cmds + 2*8) / 4)
		wide := false
		switch t := f.relocType(idxWord); t {
		case bchRelocIndex16:
			wide = true
		case bchRelocIndex8:
		default:
			return nil, fmt.Errorf("face %d index pointer has relocation type %d, expected 0x%02X or 0x%02X",
				i, t, bchRelocIndex16, bchRelocIndex8)
		}
		if err := f.appendFace(sh, &regs, wide); err != nil {
			return nil, fmt.Errorf("face %d: %w", i, err)
		}
	}
	return sh, nil
}

// appendFace reads one draw's vertices and indices out of the register state,
// appending them to the mesh with the indices rebased onto what is already
// there (a mesh's faces are separate draws over separate slices of the array).
func (f *BCH) appendFace(sh *BCHMesh, regs *[0x300]uint32, wide bool) error {
	base := f.dataX + int64(regs[regAttrBase])<<3
	arr := base + int64(regs[regLoaderOff])
	stride := int64(regs[regLoaderCfg] >> 16 & 0xFF)
	if stride == 0 {
		return fmt.Errorf("vertex stride is zero")
	}

	// Indices first: they say how much of the array the draw touches.
	count := int64(regs[regNumVertices])
	idxAddr := base + int64(regs[regIndexConfig]&0x7FFFFFFF)
	width := int64(1)
	if wide {
		width = 2
	}
	if !f.inRange(idxAddr, count*width) {
		return fmt.Errorf("index array at 0x%x (%d x %d) runs outside the file", idxAddr, count, width)
	}
	idx := make([]uint32, count)
	maxI := uint32(0)
	for i := range idx {
		if wide {
			idx[i] = uint32(f.u16(idxAddr + int64(i)*2))
		} else {
			idx[i] = uint32(f.raw[idxAddr+int64(i)])
		}
		if idx[i] > maxI {
			maxI = idx[i]
		}
	}
	nVert := int64(maxI) + 1
	if !f.inRange(arr, nVert*stride) {
		return fmt.Errorf("vertex array at 0x%x (%d x %d) runs outside the file", arr, nVert, stride)
	}

	// The loader's component list names, in storage order, which attribute each
	// slot supplies; the attribute's own format nibble (0x201/0x202) gives its
	// component count and type, and the shader-input permutation (0x2BB/0x2BC)
	// gives its meaning. Reading the list rather than assuming slot i holds
	// attribute i costs nothing and is what the hardware does.
	type attr struct {
		input, size, typ int
		off              int64
	}
	slots := int(regs[regLoaderCfg] >> 28 & 0xF)
	compList := uint64(regs[regLoaderComp]) | uint64(regs[regLoaderCfg]&0xFFFF)<<32
	fmtWord := uint64(regs[regAttrFmtLow]) | uint64(regs[regAttrFmtHigh]&0xFFFF)<<32
	inPerm := uint64(regs[regVshAttrPermL]) | uint64(regs[regVshAttrPermH])<<32
	var attrs []attr
	var off int64
	for s := 0; s < slots; s++ {
		a := int(compList >> (4 * uint(s)) & 0xF)
		if a >= 12 { // 12-15 are padding slots, sized 1-4 bytes
			off += int64(a - 11)
			continue
		}
		nib := uint32(fmtWord >> (4 * uint(a)) & 0xF)
		size, typ := int(nib>>2)+1, int(nib&3)
		attrs = append(attrs, attr{
			input: int(inPerm >> (4 * uint(a)) & 0xF),
			size:  size, typ: typ, off: off,
		})
		off += int64(size) * bchAttrTypeSize[typ]
	}
	// The attributes may fall short of the stride by up to three bytes: a
	// vertex is padded to a multiple of four. Anything else means the layout
	// was misread, so the tolerance is exactly the alignment and no more.
	if (off+3)/4*4 != stride {
		return fmt.Errorf("attributes sum to %d bytes, which does not pad to the stride of %d", off, stride)
	}

	if len(sh.Verts) == 0 {
		sh.ArrayOffset, sh.VertexStride = arr-f.dataX, stride
	}
	// Skinning shape: a mesh may carry indices and weights, indices alone
	// (rigid, one bone per vertex), or neither (rigid, the whole draw on the
	// palette's first bone). All three occur in one character.
	hasIndex, hasWeight := false, false
	for _, a := range attrs {
		switch a.input {
		case bchInBoneIndex:
			hasIndex = true
		case bchInBoneWeight:
			hasWeight = true
		}
	}
	if len(sh.Palette) > 0 {
		sh.HasSkin = true
	}

	firstVert := uint32(len(sh.Verts))
	for v := int64(0); v < nVert; v++ {
		vo := arr + v*stride
		// White, not black, when the mesh has no colour attribute. The shader
		// input is simply never written for such a mesh, and a material whose
		// combiner multiplies by the vertex colour has to come out unchanged —
		// Captain Toad's hands carry position, normal and a bone index and
		// nothing else, and their chain ends `x vtxcol`, so a zero here paints
		// them black. HasColor stays false: this is the identity for a mesh
		// that has no opinion, not a colour anyone stored.
		out := BCHVertex{Color: [4]uint8{255, 255, 255, 255}}
		for _, a := range attrs {
			var vals [4]float32
			for k := 0; k < a.size; k++ {
				vals[k] = f.attrValue(vo+a.off, a.typ, k)
			}
			switch {
			case a.input == bchInPosition:
				copy(out.Pos[:], vals[:3])
			case a.input == bchInNormal:
				copy(out.Normal[:], vals[:3])
				sh.HasNormal = true
			case a.input == bchInColor:
				for k := 0; k < 4; k++ {
					out.Color[k] = uint8(clampf01(vals[k])*255 + 0.5)
				}
				sh.HasColor = true
			case a.input == bchInBoneIndex:
				for k := 0; k < a.size && k < 4; k++ {
					out.Joints[k] = uint8(vals[k]*attrScale(a.typ) + 0.5)
				}
			case a.input == bchInBoneWeight:
				for k := 0; k < a.size && k < 4; k++ {
					out.Weights[k] = vals[k] * weightScale(a.typ)
				}
			case a.input >= bchInTexCoord && a.input <= bchInTexCoord+2:
				set := a.input - bchInTexCoord
				out.UV[set] = [2]float32{vals[0], vals[1]}
				if set+1 > sh.UVCount {
					sh.UVCount = set + 1
				}
			}
		}
		if !hasWeight {
			// Rigid: one bone, full weight. With no index attribute either, the
			// draw rides the palette's first slot.
			out.Weights[0] = 1
			if !hasIndex {
				out.Joints[0] = 0
			}
		} else {
			var sum float32
			for _, w := range out.Weights {
				sum += w
			}
			if sum < 0.99 || sum > 1.01 {
				return fmt.Errorf("vertex %d weights %v sum to %g, not 1", v, out.Weights, sum)
			}
		}
		// A joint names a palette slot, so it has to be one the palette has.
		// This is the check that fails loudly if the palette's length were ever
		// misread — which it was, when the header's mode and count words were
		// taken for bones and every joint landed two slots off.
		if len(sh.Palette) > 0 {
			for k, w := range out.Weights {
				if w != 0 && int(out.Joints[k]) >= len(sh.Palette) {
					return fmt.Errorf("vertex %d influence %d names palette slot %d of %d",
						v, k, out.Joints[k], len(sh.Palette))
				}
			}
		}
		sh.Verts = append(sh.Verts, out)
	}
	for _, i := range idx {
		sh.Indices = append(sh.Indices, firstVert+i)
	}
	return nil
}

// BindPose composes each bone's world matrix down the parent chain, in the pose
// the bone table stores. It is the matrix a skinned vertex is posed *out of*:
// multiplying it by the bone's own inverse-bind matrix gives the identity, which
// is the invariant the skeleton decode is proved by (TestBCHSkeletonBindPose).
//
// Row-major 4x4s, in float64 because composing twenty-odd of them in float32
// loses more than the residual that check allows.
func (m *BCHModel) BindPose() [][16]float64 {
	world := make([][16]float64, len(m.Bones))
	for i, b := range m.Bones {
		l := mul4(mul4(trans4(b.Trans), eulerZYX(b.Rotate)), scale4(b.Scale))
		if b.Parent < 0 || b.Parent >= i {
			world[i] = l
			continue
		}
		world[i] = mul4(world[b.Parent], l)
	}
	return world
}

func mul4(a, b [16]float64) [16]float64 {
	var o [16]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			s := 0.0
			for k := 0; k < 4; k++ {
				s += a[r*4+k] * b[k*4+c]
			}
			o[r*4+c] = s
		}
	}
	return o
}

func trans4(t [3]float32) [16]float64 {
	return [16]float64{1, 0, 0, float64(t[0]), 0, 1, 0, float64(t[1]), 0, 0, 1, float64(t[2]), 0, 0, 0, 1}
}

func scale4(s [3]float32) [16]float64 {
	return [16]float64{float64(s[0]), 0, 0, 0, 0, float64(s[1]), 0, 0, 0, 0, float64(s[2]), 0, 0, 0, 0, 1}
}

// eulerZYX builds Rz·Ry·Rx from an XYZ triple in radians — the order the bind
// pose closes in, and the one a bone's animated rotation composes in too.
func eulerZYX(r [3]float32) [16]float64 {
	rot := func(ax int, a float64) [16]float64 {
		c, s := math.Cos(a), math.Sin(a)
		switch ax {
		case 0:
			return [16]float64{1, 0, 0, 0, 0, c, -s, 0, 0, s, c, 0, 0, 0, 0, 1}
		case 1:
			return [16]float64{c, 0, s, 0, 0, 1, 0, 0, -s, 0, c, 0, 0, 0, 0, 1}
		}
		return [16]float64{c, -s, 0, 0, s, c, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	}
	return mul4(mul4(rot(2, float64(r[2])), rot(1, float64(r[1]))), rot(0, float64(r[0])))
}

// InvBind4 is a bone's stored inverse-bind matrix as a full row-major 4x4.
func (b *BCHBone) InvBind4() [16]float64 {
	var o [16]float64
	for r := 0; r < 3; r++ {
		for c := 0; c < 4; c++ {
			o[r*4+c] = float64(b.InvBind[r][c])
		}
	}
	o[15] = 1
	return o
}

// bchAttrTypeSize is the byte width of each PICA attribute component type.
var bchAttrTypeSize = [4]int64{1, 1, 2, 4} // byte, ubyte, short, float

// attrValue reads component k of an attribute. Integer types are the raw value:
// BCH's colours are unsigned bytes meaning 0-255 and its positions are floats,
// so no attribute scale is involved (unlike CGFX, which carries one).
func (f *BCH) attrValue(o int64, typ, k int) float32 {
	switch typ {
	case 0:
		return float32(int8(f.raw[o+int64(k)]))
	case 1:
		return float32(f.raw[o+int64(k)]) / 255
	case 2:
		return float32(int16(f.u16(o + int64(k)*2)))
	default:
		return f.f32(o + int64(k)*4)
	}
}

func clampf01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// A texture object is a triple of (command-list offset, word count), one per
// texture unit the texture can be bound to; the unit-0 list is the one that
// carries its size, format and pixel address. Like a mesh, a texture describes
// itself in PICA registers rather than in a struct of its own.
const (
	bchTexCmds   = 0x00
	bchTexCmdLen = 0x04
)

// Texture-unit registers, as the unit-0 block writes them.
const (
	regTex0Dim    = 0x082 // width in bits 16-26, height in bits 0-10
	regTex0Addr   = 0x085 // a byte offset into the data section, not yet relocated
	regTex0Format = 0x08E
)

// BCHTexture is one decoded texture.
type BCHTexture struct {
	Name          string
	Width, Height int
	Format        uint32
	Image         *image.NRGBA
}

// DecodeTexture decodes the texture a Textures-group entry points at. Only the
// top mip level is decoded — the smaller levels follow it in the data section
// and an exporter has no use for them.
func (f *BCH) DecodeTexture(e BCHEntry) (*BCHTexture, error) {
	if !f.inRange(e.Offset, 8) {
		return nil, fmt.Errorf("bch: texture %q header runs outside the file", e.Name)
	}
	off, words := int64(f.u32(e.Offset+bchTexCmds)), int64(f.u32(e.Offset+bchTexCmdLen))
	if words == 0 || !f.inRange(f.gpu+off, words*4) {
		return nil, fmt.Errorf("bch: texture %q has no unit-0 command list", e.Name)
	}
	ws, err := DecodePICA(f.raw[f.gpu+off : f.gpu+off+words*4])
	if err != nil {
		return nil, fmt.Errorf("bch: texture %q: %w", e.Name, err)
	}
	var regs [0x300]uint32
	for _, w := range ws {
		if w.Reg < 0x300 {
			regs[w.Reg] = w.Value
		}
	}
	t := &BCHTexture{
		Name:   e.Name,
		Width:  int(regs[regTex0Dim] >> 16 & 0x7FF),
		Height: int(regs[regTex0Dim] & 0x7FF),
		Format: regs[regTex0Format] & 0xF,
	}
	addr := f.data + int64(regs[regTex0Addr])
	need := int64(t.Width) * int64(t.Height) * int64(picaTexBits[t.Format]) / 8
	if !f.inRange(addr, need) {
		return nil, fmt.Errorf("bch: texture %q pixels at 0x%x (+%d) run outside the file", e.Name, addr, need)
	}
	img, err := DecodePICATexture(f.raw[addr:addr+need], t.Format, t.Width, t.Height)
	if err != nil {
		return nil, fmt.Errorf("bch: texture %q: %w", e.Name, err)
	}
	t.Image = img
	return t, nil
}

// BCHMaterial is a model's binding of one material: its name and the three name
// references the binding carries. The names are resolved against whatever
// container holds the textures — a stage's models keep their textures in a
// separate `.bch` beside them, so the binding is by name, not by pointer.
//
// The four slots are the material's three **texture units** followed by its own
// short name. This stage's terrain shader binds the depth-shadow map to unit 0,
// the albedo to unit 1 and the normal map to unit 2 — and its TEV stage 0 is
// `Modulate(PrimaryFragmentColor, Texture1)`, the lighting unit's output times
// that albedo, which is where the game's whole look comes from. Slot 0 is an
// unset name on the sky's materials, so read them as names, not as guarantees.
type BCHMaterial struct {
	Name  string
	Names [4]string

	// How the material treats alpha, read from the PICA state it programs.
	// Ignoring this is not cosmetic: the stage's albedo textures are ETC1A4 and
	// therefore *all* carry a 4-bit alpha plane, but most of the materials
	// sampling them neither blend nor alpha-test, so that plane is not coverage
	// and must not be used as any. All five of the stone albedos carry the
	// identical alpha histogram (85..221) — the tell that it is not a mask.
	// BlendSrc/BlendDst are the colour blend factors the material programs, in
	// the PICA's numbering. They are kept rather than reduced to a flag because
	// "it blends" is not one behaviour: a glow quad that adds its colour to the
	// frame is a different thing from a leaf that covers what is behind it, and
	// only the factors tell them apart.
	BlendSrc, BlendDst uint8

	Blends    bool  // the blend function is anything other than src×1 + dst×0
	AlphaTest bool  // the alpha test is enabled
	AlphaFunc uint8 // its comparison, in the PICA's numbering (gpu_tev.go)
	AlphaRef  uint8 // its reference value

	// WrapS/WrapT are the texture-unit wrap modes the material programs, one
	// pair per unit. Hard-coding REPEAT tiles the goal star's single eye
	// texture into a row of slots across its face.
	WrapS, WrapT [3]PICAWrap
	wrapSet      [3]bool

	// texMatrix is the affine transform each unit applies to the vertex UV
	// before sampling: u' = m[0]u + m[1]v + m[2], v' = m[3]u + m[4]v + m[5].
	texMatrix [3][6]float32
	texMatSet [3]bool

	// The material's own combiner program, and which of the six stages its
	// command list actually writes. The distinction matters: PICA registers
	// latch, so the register file during a draw also holds whatever the
	// *previous* material left in the stages this one does not touch. Only the
	// file says which stages are this material's.
	tev    tevState
	stages int
}

// PICA alpha-test comparisons, as register 0x104 numbers them.
const (
	AlphaNever = iota
	AlphaAlways
	AlphaEqual
	AlphaNotEqual
	AlphaLess
	AlphaLEqual
	AlphaGreater
	AlphaGEqual
)

// TexMatrix returns the affine transform a texture unit applies to the vertex
// UV, and whether the material states one. Apply it as
//
//	u' = m[0]*u + m[1]*v + m[2]
//	v' = m[3]*u + m[4]*v + m[5]
func (m *BCHMaterial) TexMatrix(unit int) ([6]float32, bool) {
	return m.texMatrix[unit], m.texMatSet[unit]
}

// WrapModes returns the wrap modes the material programs for a texture unit,
// and whether it programs them at all. A material that never writes the unit's
// parameter register leaves whatever the previous draw set — the register file
// is not the material.
func (m *BCHMaterial) WrapModes(unit int) (s, t PICAWrap, ok bool) {
	return m.WrapS[unit], m.WrapT[unit], m.wrapSet[unit]
}

// PICA blend factors this decoder names.
const (
	BlendZero        = 0
	BlendOne         = 1
	BlendSrcAlpha    = 6
	BlendOneMinusSrc = 7
)

// Additive reports whether the material adds its colour to what is already
// there rather than covering it — `src + dst`, however the source is weighted.
// A glow or a sparkle wants to be exported as additive; a cut-out leaf does not.
func (m *BCHMaterial) Additive() bool {
	return m.Blends && m.BlendDst == BlendOne &&
		(m.BlendSrc == BlendOne || m.BlendSrc == BlendSrcAlpha)
}

// AlphaCutoff turns an enabled alpha test into the equivalent "keep alpha >= c"
// threshold in [0,1], and reports whether the test is expressible that way. A
// GREATER test keeps strictly above its reference, so its threshold is one
// texel value higher than a GEQUAL one with the same reference.
func (m *BCHMaterial) AlphaCutoff() (float32, bool) {
	if !m.AlphaTest {
		return 0, false
	}
	switch m.AlphaFunc {
	case AlphaGreater:
		return float32(int(m.AlphaRef)+1) / 255, true
	case AlphaGEqual:
		return float32(m.AlphaRef) / 255, true
	}
	return 0, false
}

// AlbedoUnit is the texture unit carrying the image this material's exported
// texture is built from: the lowest-numbered unit whose named texture the
// archive actually holds.
//
// It is not always unit 1, though the terrain and the characters make it look
// that way — their unit 0 is `$shadowmap`, a run-time target no archive holds,
// so the first real texture is on unit 1. The goal star's sparkles and Toad's
// headlamp put theirs on unit 0, and assuming unit 1 for them reads a texture
// matrix that is all zeros, collapsing every UV onto one texel and painting the
// quad a flat colour.
func (m *BCHMaterial) AlbedoUnit(textures map[string]*image.NRGBA) int {
	for u := 0; u < 3; u++ {
		if textures[m.Names[u]] != nil {
			return u
		}
	}
	return bchUnitAlbedo
}

// Texture returns the name of the albedo the material samples — the texture on
// unit 1, which is the one its combiner multiplies the lighting by.
func (m *BCHMaterial) Texture() string { return m.Names[bchUnitAlbedo] }

// NormalMap returns the material's unit-2 normal map, if it binds one.
func (m *BCHMaterial) NormalMap() string { return m.Names[bchUnitNormal] }

// Texture-unit assignments this stage's shader uses.
const (
	bchUnitShadow = 0
	bchUnitAlbedo = 1
	bchUnitNormal = 2
)

// The model's per-material binding table: one 0x2C-byte entry per material,
// holding a pointer to the material object and the names of the textures it
// binds to units 0, 1 and 2.
const (
	bchMatBindStride = 0x2C
	bchMatBindObject = 0x00
	bchMatBindTex0   = 0x1C // the unit-0 name; units 1 and 2 follow, then the short name
)

// The material object's own PICA command list: (offset, word count) into the
// GPU command section, holding the material's whole fragment state — the TEV
// stages, the texture-unit setup, the blend function and the alpha test.
const (
	bchMatCmds   = 0xC8
	bchMatCmdLen = 0xCC
)

// The output-merger registers the material programs.
const (
	regBlendFunc = 0x101 // src/dst factors in bits 16-31
	regAlphaTest = 0x104 // bit0 enable, bits 4-6 function, bits 8-15 reference
)

// PICA blend factors, of which only these two are needed to recognise a
// material that does not blend at all.
const (
	blendZero = 0
	blendOne  = 1
)

// A material's texture MATRICES are not in its structure at all: it uploads
// them as vertex-shader float uniforms, three rows each, at c11-c13 for unit 0,
// c14-c16 for unit 1 and c17-c19 for unit 2. The shader multiplies the vertex's
// UV by them before sampling.
//
// They are load-bearing and they are not identities. Captain Toad's cap runs
// its UV through `u' = 2u - 2, v' = 2v - 3` — a scale and a whole-texture
// offset — so ignoring the matrix samples a different part of the mask
// entirely, and his spots come out in the wrong places or not at all. His face
// is `2u, 2v - 1`; the goal star's eye is `8u`, which with the mirrored wrap on
// that unit is how ONE half-eye texture becomes a pair of eyes.
//
// This is the same lesson the banner's CGFX taught (uv-set-is-not-a-texture-
// coord): a vertex UV set is not a texture coordinate until the material's
// mapper has had its say.
const (
	bchMatTexMatrix = 11 // uniform index of unit 0's first row
	bchMatTexRows   = 3  // uniform rows per unit
)

// A material's texture mappers: one 0x10-byte block per texture unit, holding
// the sampler state the engine programs into that unit's parameter register at
// bind time.
//
//	+0x00 ?          +0x01 wrap S      +0x02 wrap T      +0x03 ?
//	+0x04 ?          +0x08 f32 ?       +0x0C ? ... +0x0F 0xFF
//
// ⚠ **The wrap is not in any command list.** Not the material's, not the
// texture object's (which carries three lists of its own, one per unit, and
// writes only the dimensions, the address and the format), not the mesh's.
// Reading it out of the register file gives CLAMP_TO_EDGE for every material in
// the game, which is what an unwritten register reads and not what anything
// said.
//
// These two bytes are, and it is checked rather than assumed: against the
// values the running game programs for Captain Toad's own draws — sixteen
// measurements over three units, read from `bootoracle -gputrace` — every one
// matches. They are also load-bearing. Toad's cap is MIRRORED_REPEAT and his
// face is CLAMP_TO_EDGE while his rucksack is REPEAT, so a single guessed mode
// is wrong for two thirds of him: the spots run off the edge of the cap, and
// the face tiles.
const (
	bchMatMappers   = 0x110
	bchMatMapStride = 0x10
	bchMapWrapS     = 0x01
	bchMapWrapT     = 0x02
)

// decodeMaterialState runs a material object's command list and reads the
// output-merger state out of the registers it sets.
func (f *BCH) decodeMaterialState(obj int64, m *BCHMaterial) error {
	off, words := int64(f.u32(obj+bchMatCmds)), int64(f.u32(obj+bchMatCmdLen))
	if words == 0 {
		return nil
	}
	if !f.inRange(f.gpu+off, words*4) {
		return fmt.Errorf("material %q: command list at gpu+0x%x (%d words) runs outside the file", m.Name, off, words)
	}
	ws, err := DecodePICA(f.raw[f.gpu+off : f.gpu+off+words*4])
	if err != nil {
		return fmt.Errorf("material %q: %w", m.Name, err)
	}
	var regs [0x300]uint32
	var written [0x300]bool
	uniforms := map[int][4]float32{}
	var uIdx int
	var uF32 bool
	var uBuf []uint32
	for _, w := range ws {
		if w.Reg < 0x300 {
			regs[w.Reg] = w.Value
			written[w.Reg] = true
		}
		// The float-uniform FIFO, decoded exactly as the GPU decodes it: a
		// config word names the destination and the mode, then the data words
		// arrive w first (gpu.go's floatUniformWord).
		switch {
		case w.Reg == regVshFloatCfg:
			uIdx, uF32, uBuf = int(w.Value&0xFF), w.Value>>31 != 0, uBuf[:0]
		case w.Reg >= regVshFloatData && w.Reg < regVshFloatData+8:
			uBuf = append(uBuf, w.Value)
			n := 3
			if uF32 {
				n = 4
			}
			if len(uBuf) == n {
				if uF32 {
					uniforms[uIdx] = [4]float32{toF24(f32bits(uBuf[3])), toF24(f32bits(uBuf[2])),
						toF24(f32bits(uBuf[1])), toF24(f32bits(uBuf[0]))}
				} else {
					uniforms[uIdx] = unpackF24x4(uBuf[0], uBuf[1], uBuf[2])
				}
				uIdx++
				uBuf = uBuf[:0]
			}
		}
	}
	// Which stages the list programs. A stage is this material's if its source
	// register was written; the ones past that are the register file's history.
	m.stages = 0
	for i, base := range tevStageBase {
		if written[base] {
			m.stages = i + 1
		}
	}
	m.tev = tevStateFromRegs(&regs)
	_ = written

	// The texture matrices, from the uniforms the command list uploads.
	for u := 0; u < 3; u++ {
		r0, ok0 := uniforms[bchMatTexMatrix+u*bchMatTexRows]
		r1, ok1 := uniforms[bchMatTexMatrix+u*bchMatTexRows+1]
		if !ok0 || !ok1 {
			continue
		}
		m.texMatrix[u] = [6]float32{r0[0], r0[1], r0[3], r1[0], r1[1], r1[3]}
		m.texMatSet[u] = true
	}

	for u := int64(0); u < 3; u++ {
		b := obj + bchMatMappers + u*bchMatMapStride
		if !f.inRange(b, bchMatMapStride) {
			return fmt.Errorf("material %q: texture mapper %d runs outside the file", m.Name, u)
		}
		ws, wt := PICAWrap(f.raw[b+bchMapWrapS]), PICAWrap(f.raw[b+bchMapWrapT])
		if ws > WrapMirroredRepeat || wt > WrapMirroredRepeat {
			return fmt.Errorf("material %q: texture mapper %d has wrap modes %d/%d, which are not modes",
				m.Name, u, ws, wt)
		}
		m.WrapS[u], m.WrapT[u], m.wrapSet[u] = ws, wt, true
	}

	bf := regs[regBlendFunc]
	srcRGB, dstRGB := bf>>16&0xF, bf>>20&0xF
	srcA, dstA := bf>>24&0xF, bf>>28&0xF
	m.Blends = !(srcRGB == blendOne && dstRGB == blendZero && srcA == blendOne && dstA == blendZero)
	m.BlendSrc, m.BlendDst = uint8(srcRGB), uint8(dstRGB)
	at := regs[regAlphaTest]
	m.AlphaTest = at&1 != 0
	m.AlphaFunc = uint8(at >> 4 & 7)
	m.AlphaRef = uint8(at >> 8)
	return nil
}

// Stages is how many combiner stages the material's own command list programs.
// The register file always holds six; only this many are the material's.
func (m *BCHMaterial) Stages() int { return m.stages }

// Describe writes the material's combiner program out one stage at a time, in
// the same terms the running GPU's per-draw dump uses. It is the instrument for
// answering "where does this surface's colour come from" without guessing.
func (m *BCHMaterial) Describe() []string {
	out := make([]string, 0, m.stages)
	for i := 0; i < m.stages; i++ {
		s := &m.tev.stages[i]
		out = append(out, fmt.Sprintf(
			"rgb %s(%s) %s(%s) %s(%s) %s x%d | a %s(%s) %s(%s) %s(%s) %s x%d | konst (%d,%d,%d,%d) buf=%v/%v",
			tevSrcName(uint32(s.colr[0].src)), tevCOpName(uint32(s.colr[0].op)),
			tevSrcName(uint32(s.colr[1].src)), tevCOpName(uint32(s.colr[1].op)),
			tevSrcName(uint32(s.colr[2].src)), tevCOpName(uint32(s.colr[2].op)),
			tevOpName(uint32(s.combC)), 1<<s.scaleC,
			tevSrcName(uint32(s.alph[0].src)), tevAOpName(uint32(s.alph[0].op)),
			tevSrcName(uint32(s.alph[1].src)), tevAOpName(uint32(s.alph[1].op)),
			tevSrcName(uint32(s.alph[2].src)), tevAOpName(uint32(s.alph[2].op)),
			tevOpName(uint32(s.combA)), 1<<s.scaleA,
			s.konst.r, s.konst.g, s.konst.b, s.konst.a, s.updC, s.updA))
	}
	return out
}

// VertexColorScalar reports whether the material reads the vertex colour as a
// single number rather than as a colour.
//
// It matters because a mesh can carry a vertex colour whose channels are not a
// colour at all. Captain Toad's own meshes are the case: the red channel runs
// the whole 0-255 range while green and blue sit at 255, and every colour stage
// that reads the vertex colour reads it through the *source-red* operand — the
// artists' occlusion, stored once and replicated. Multiplying a baked light by
// all three channels then drags green and blue down independently of red and
// paints the character teal. The stage terrain, by contrast, reads `vtxcol(rgb)`
// and means it.
//
// The material is the one that knows, so it is the one asked.
func (m *BCHMaterial) VertexColorScalar() bool {
	reads := 0
	for i := 0; i < m.stages; i++ {
		for _, o := range m.tev.stages[i].colr {
			if o.src != tevSrcVertexColor {
				continue
			}
			reads++
			if o.op == tevOpRGB || o.op == tevOpOneMinusRGB {
				return false
			}
		}
	}
	return reads > 0
}

// The TEV source and colour-operand codes this file names.
const (
	tevSrcVertexColor = 0
	tevOpRGB          = 0
	tevOpOneMinusRGB  = 1
)

// Shade evaluates how much the lighting and the vertex colour darken this
// material, as the factor to multiply its baked albedo by.
//
// It is a ratio of the material's own program to itself: the chain run with
// this vertex's lighting, over the chain run with the lighting neutral — the
// second being exactly what BakeAlbedo wrote into the texture. Anything the
// chain does to the texture cancels, and what is left is the part that varies
// per vertex.
//
// A ratio rather than a second evaluation, because the albedo is not a factor
// of these chains. Captain Toad's cap is built out of its texture rather than
// multiplied by it — `(1-mask) x white + mask x red` — so running the chain
// with a white texture does not yield "the shading with no albedo", it yields
// the colour of his spots.
//
// The vertex colour is left to the chain too, and that matters: his cap ends
// with `(vtxcol.r + 0.46) x everything`, an ADD before the multiply, so an
// occlusion above about half does not darken him at all. Multiplying by the
// vertex colour by hand makes him a third too dark and, because only the red
// channel carries the occlusion, green.
//
// texel is the albedo texel to take the ratio at; the caller should pick a
// bright one, and BakeCheck says whether the choice matters.
func (m *BCHMaterial) Shade(vertexColor [4]uint8, light [3]float64, texel [4]uint8) [3]float64 {
	num, den, ok := m.shadePair(vertexColor, light, texel)
	if !ok {
		return [3]float64{1, 1, 1}
	}
	var out [3]float64
	for i := 0; i < 3; i++ {
		if den[i] <= 0 {
			out[i] = 0
			continue
		}
		out[i] = float64(num[i]) / float64(den[i])
	}
	return out
}

func (m *BCHMaterial) shadePair(vertexColor [4]uint8, light [3]float64, texel [4]uint8) (num, den [3]int32, ok bool) {
	white := rgba{255, 255, 255, 255}
	tex := [3]rgba{white,
		{int32(texel[0]), int32(texel[1]), int32(texel[2]), int32(texel[3])},
		{128, 128, 255, 255}}
	v := rgba{int32(vertexColor[0]), int32(vertexColor[1]), int32(vertexColor[2]), int32(vertexColor[3])}
	prim := rgba{clamp255(float32(light[0] * 255)), clamp255(float32(light[1] * 255)),
		clamp255(float32(light[2] * 255)), 255}
	lit, ok1 := m.tev.run(v, prim, rgba{0, 0, 0, 0}, tex)
	flat, ok2 := m.tev.run(white, white, rgba{0, 0, 0, 0}, tex)
	if !ok1 || !ok2 {
		return num, den, false
	}
	return [3]int32{lit.r, lit.g, lit.b}, [3]int32{flat.r, flat.g, flat.b}, true
}

// BakeCheck reports how much the shading factor depends on which albedo texel
// it is measured at. The split into "a baked albedo texture" and "a per-vertex
// factor" is only meaningful if it does not: zero means the material multiplies
// its albedo by a shading term, and a large value means the two are entangled
// and must not be separated.
func (m *BCHMaterial) BakeCheck() float64 {
	worst := 0.0
	for _, l := range [][3]float64{{0.2, 0.2, 0.2}, {0.6, 0.6, 0.6}, {1, 0.8, 0.5}} {
		for _, vc := range [][4]uint8{{0, 255, 255, 255}, {128, 255, 255, 255}, {255, 255, 255, 255}} {
			var ref [3]float64
			first := true
			for _, t := range [][4]uint8{{255, 255, 255, 255}, {160, 160, 160, 255}, {96, 96, 96, 255}} {
				num, den, ok := m.shadePair(vc, l, t)
				if !ok {
					return 1
				}
				var f [3]float64
				dark := false
				for i := 0; i < 3; i++ {
					if den[i] < 16 { // too dark to measure a ratio against
						dark = true
						break
					}
					f[i] = float64(num[i]) / float64(den[i])
				}
				if dark {
					continue
				}
				if first {
					ref, first = f, false
					continue
				}
				for i := 0; i < 3; i++ {
					if d := math.Abs(f[i] - ref[i]); d > worst {
						worst = d
					}
				}
			}
		}
	}
	return worst
}

// BakeAlbedo evaluates the material's own combiner over its unit-1 texture and
// returns the surface colour, unlit.
//
// A character's colour is not in its texture. Captain Toad's cap samples a
// 32x32 black-and-white *mask* and builds the surface from it with two stage
// constants — `(1-mask) x (242,246,255)` for the white cap, `+ mask x
// (196,30,25)` for the spots — so binding that texture as an albedo, which is
// what a decoder does when it stops at the texture name, paints him a black cap
// with white spots. The colours are in the material, and the material is a
// program.
//
// Running that program with the lighting inputs *neutral* — diffuse white,
// specular black, vertex colour white — is what "unlit albedo" means for a
// combiner, and on these materials it reduces exactly: the stages that consume
// the lighting become identities (`fragpri x prev` with fragpri white is prev),
// and what survives is the texture-and-constants expression. The lighting is
// then applied the way the terrain's is, baked per vertex in gamma space.
//
// tex2 is sampled at the same coordinate when the chain reads it; a material
// whose second texture is on a different UV set would need more than that, and
// none of the characters' do (checked: their colour stages read tex1 only).
func (m *BCHMaterial) BakeAlbedo(textures map[string]*image.NRGBA) (*image.NRGBA, error) {
	// Every unit the material binds, sampled from the texture it names.
	//
	// ⚠ Unit 0 is NOT always the shadow map. It is on the terrain and on the
	// characters, whose materials name `$shadowmap` there — a target the engine
	// renders at run time, which no archive holds. The goal star binds real
	// textures to it and BUILDS ITS ALPHA OUT OF THEM: its glow quad's chain is
	// `alpha = tex0.r + tex1.r + tex2.r`. Substituting an opaque white for unit
	// 0 because the terrain's is a shadow map turned that glow into a solid
	// yellow card with the star behind it.
	//
	// So the rule is what the archive can answer, not what the name looks like:
	// a bound name the archive holds is sampled, and a name it does not hold is
	// something the engine makes at run time and gets the neutral value that
	// means "not there" — unoccluded for the shadow unit, a flat normal for the
	// bump unit.
	albedo := m.AlbedoUnit(textures)
	var units [3]*image.NRGBA
	w, h := 0, 0
	for u := 0; u < 3; u++ {
		img := textures[m.Names[u]]
		units[u] = img
		if img == nil {
			continue
		}
		if u == albedo {
			w, h = img.Bounds().Dx(), img.Bounds().Dy()
		}
	}
	if w == 0 || h == 0 {
		// The chain is constants alone, so one texel says everything about it.
		w, h = 1, 1
	}

	// The baked image lives in the ALBEDO unit's texture space, because that is
	// the space the exported UVs will be in once the albedo unit's own matrix
	// has been applied to them. The other units read the same vertex UV through
	// a matrix of their OWN, so their sample position is this texel put back
	// through the albedo matrix's inverse and forward through theirs.
	mapToUnit := func(u int, s, t float32) (float32, float32) {
		a, oka := m.TexMatrix(albedo)
		b, okb := m.TexMatrix(u)
		if !oka || !okb || u == albedo {
			return s, t
		}
		iu, iv, ok := invAffine2(a, s, t)
		if !ok {
			return s, t
		}
		return b[0]*iu + b[1]*iv + b[2], b[3]*iu + b[4]*iv + b[5]
	}

	neutral := [3]rgba{{255, 255, 255, 255}, {255, 255, 255, 255}, {128, 128, 255, 255}}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	white := rgba{255, 255, 255, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			tex := neutral
			s, t := (float32(x)+0.5)/float32(w), (float32(y)+0.5)/float32(h)
			for u := 0; u < 3; u++ {
				img := units[u]
				if img == nil {
					continue
				}
				su, sv := mapToUnit(u, s, t)
				c := sampleWrapped(img, su, sv, m.WrapS[u], m.WrapT[u])
				tex[u] = rgba{int32(c.R), int32(c.G), int32(c.B), int32(c.A)}
			}
			r, ok := m.tev.run(white, white, rgba{0, 0, 0, 0}, tex)
			if !ok {
				return nil, fmt.Errorf("material %q: its combiner uses a source or op this model does not implement", m.Name)
			}
			out.SetNRGBA(x, y, color.NRGBA{uint8(r.r), uint8(r.g), uint8(r.b), uint8(r.a)})
		}
	}
	return out, nil
}

// invAffine2 inverts the 2x3 affine a texture matrix applies to a UV.
func invAffine2(m [6]float32, x, y float32) (u, v float32, ok bool) {
	det := m[0]*m[4] - m[1]*m[3]
	if det == 0 {
		return 0, 0, false
	}
	x, y = x-m[2], y-m[5]
	return (m[4]*x - m[1]*y) / det, (m[0]*y - m[3]*x) / det, true
}

// sampleWrapped reads a texel with the unit's own wrap modes, nearest-neighbour.
func sampleWrapped(img *image.NRGBA, u, v float32, ws, wt PICAWrap) color.NRGBA {
	b := img.Bounds()
	x, okx := wrapTexel(int(math.Floor(float64(u)*float64(b.Dx()))), b.Dx(), ws)
	y, oky := wrapTexel(int(math.Floor(float64(v)*float64(b.Dy()))), b.Dy(), wt)
	if !okx || !oky {
		return color.NRGBA{} // clamp-to-border: outside is the border, and the
		// engine's border here is transparent black
	}
	return img.NRGBAAt(b.Min.X+x, b.Min.Y+y)
}

func wrapTexel(v, n int, mode PICAWrap) (int, bool) {
	switch mode {
	case WrapClampToBorder:
		if v < 0 || v >= n {
			return 0, false
		}
	case WrapRepeat:
		v %= n
		if v < 0 {
			v += n
		}
	case WrapMirroredRepeat:
		p := 2 * n
		v %= p
		if v < 0 {
			v += p
		}
		if v >= n {
			v = p - 1 - v
		}
	default: // clamp to edge
		if v < 0 {
			v = 0
		}
		if v >= n {
			v = n - 1
		}
	}
	return v, true
}

// A scene's lights live in their own `.bch` beside the stage's models — the
// stage `<Base>Stage` is lit by the scene `<Base>Scene` — and the stage's
// Design archive names, per light area, which of them apply.
//
// Light object fields, from the object's start:
//
//	+0x00 u32 name offset       +0x28 u32 flags
//	+0x30 u32 ambient  RGBA     +0x34 u32 diffuse RGBA
//	+0x38 u32 specular0         +0x3C u32 specular1
//	+0x40 f32[3] direction      (only when the flags say the light has one)
//
// The flags' bit 3 says whether the light carries a direction, and the objects
// are sized accordingly (0x34 bytes without one). That reading is checked
// against the data rather than assumed: across every scene in the cartridge,
// each light the bit marks carries a vector of **exactly** unit length, which a
// misread offset does not produce.
const (
	bchLightFlags     = 0x28
	bchLightAmbient   = 0x30
	bchLightDiffuse   = 0x34
	bchLightSpecular0 = 0x38
	bchLightSpecular1 = 0x3C
	bchLightDirection = 0x40

	bchLightHasVector = 1 << 3 // the flags bit that says +0x40 is meaningful
)

// BCHLight is one of a scene's lights.
type BCHLight struct {
	Name    string
	Flags   uint32
	Ambient [3]uint8 // 255 is full scale, as the hardware's 10-bit fields read it

	// Directional lights only.
	Directional bool
	Diffuse     [3]uint8
	Specular0   [3]uint8
	Specular1   [3]uint8

	// Direction is the direction the light *travels*, in the model's world
	// space. The vector towards the light — the L of an N·L — is its negation.
	Direction [3]float32
}

// picaColor reads one of the file's RGBA colour words: four bytes, red first.
func (f *BCH) picaColor(o int64) [3]uint8 {
	v := f.u32(o)
	return [3]uint8{uint8(v), uint8(v >> 8), uint8(v >> 16)}
}

// DecodeLight decodes the light a Lights-group entry points at.
func (f *BCH) DecodeLight(e BCHEntry) (*BCHLight, error) {
	if !f.inRange(e.Offset, bchLightDirection) {
		return nil, fmt.Errorf("bch: light %q header runs outside the file", e.Name)
	}
	l := &BCHLight{
		Name:    e.Name,
		Flags:   f.u32(e.Offset + bchLightFlags),
		Ambient: f.picaColor(e.Offset + bchLightAmbient),
	}
	if l.Flags&bchLightHasVector == 0 {
		return l, nil // an ambient-only light: the object ends before the rest
	}
	if !f.inRange(e.Offset, bchLightDirection+12) {
		return nil, fmt.Errorf("bch: light %q direction runs outside the file", e.Name)
	}
	l.Directional = true
	l.Diffuse = f.picaColor(e.Offset + bchLightDiffuse)
	l.Specular0 = f.picaColor(e.Offset + bchLightSpecular0)
	l.Specular1 = f.picaColor(e.Offset + bchLightSpecular1)
	var sum float64
	for k := 0; k < 3; k++ {
		l.Direction[k] = f.f32(e.Offset + bchLightDirection + int64(k)*4)
		sum += float64(l.Direction[k]) * float64(l.Direction[k])
	}
	if d := math.Abs(math.Sqrt(sum) - 1); d > 1e-3 {
		return nil, fmt.Errorf("bch: light %q direction %v is not a unit vector (off by %g)", e.Name, l.Direction, d)
	}
	return l, nil
}

// Lights decodes every light in the container, by name.
func (f *BCH) Lights() (map[string]*BCHLight, error) {
	out := make(map[string]*BCHLight, len(f.Groups[BCHLights]))
	for _, e := range f.Groups[BCHLights] {
		l, err := f.DecodeLight(e)
		if err != nil {
			return nil, err
		}
		out[l.Name] = l
	}
	return out, nil
}
