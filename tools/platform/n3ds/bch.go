package n3ds

import (
	"encoding/binary"
	"fmt"
	"image"
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
}

// BCHMesh is one drawable: a triangle list over its own vertices, bound to one
// of the model's materials.
type BCHMesh struct {
	MaterialIndex int
	Verts         []BCHVertex
	Indices       []uint32

	HasNormal bool
	HasColor  bool
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
	bchMeshStride     = 0x38
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
	bchInPosition = 0
	bchInNormal   = 1
	bchInTangent  = 2
	bchInColor    = 3
	bchInTexCoord = 4 // ...through 6
)

// DecodeModel decodes the model a Models-group entry points at.
func (f *BCH) DecodeModel(e BCHEntry) (*BCHModel, error) {
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
		for u := 0; u < 3; u++ {
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
	if off != stride {
		return fmt.Errorf("attributes sum to %d bytes but the stride is %d", off, stride)
	}

	if len(sh.Verts) == 0 {
		sh.ArrayOffset, sh.VertexStride = arr-f.dataX, stride
	}
	firstVert := uint32(len(sh.Verts))
	for v := int64(0); v < nVert; v++ {
		vo := arr + v*stride
		var out BCHVertex
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
			case a.input >= bchInTexCoord && a.input <= bchInTexCoord+2:
				set := a.input - bchInTexCoord
				out.UV[set] = [2]float32{vals[0], vals[1]}
				if set+1 > sh.UVCount {
					sh.UVCount = set + 1
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
// **Slot 0 is the texture the material samples on unit 0** in every material
// seen, and is what an exporter wants. Slots 1 and 2 are not reliably textures:
// the stage terrain's shader puts its normal map in slot 1 and the material's
// own short name in slot 2, while the sky's puts a shader name in slot 1. Read
// them as names, not as texture units.
type BCHMaterial struct {
	Name  string
	Names [3]string

	// How the material treats alpha, read from the PICA state it programs.
	// Ignoring this is not cosmetic: the stage's albedo textures are ETC1A4 and
	// therefore *all* carry a 4-bit alpha plane, but most of the materials
	// sampling them neither blend nor alpha-test, so that plane is not coverage
	// and must not be used as any. All five of the stone albedos carry the
	// identical alpha histogram (85..221) — the tell that it is not a mask.
	Blends    bool  // the blend function is anything other than src×1 + dst×0
	AlphaTest bool  // the alpha test is enabled
	AlphaFunc uint8 // its comparison, in the PICA's numbering (gpu_tev.go)
	AlphaRef  uint8 // its reference value
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

// Texture returns the material's unit-0 texture name.
func (m *BCHMaterial) Texture() string { return m.Names[0] }

// The model's per-material binding table: one 0x2C-byte entry per material,
// holding a pointer to the material object and the names of the textures it
// binds to units 0, 1 and 2.
const (
	bchMatBindStride = 0x2C
	bchMatBindObject = 0x00
	bchMatBindTex0   = 0x20
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
	for _, w := range ws {
		if w.Reg < 0x300 {
			regs[w.Reg] = w.Value
		}
	}
	bf := regs[regBlendFunc]
	srcRGB, dstRGB := bf>>16&0xF, bf>>20&0xF
	srcA, dstA := bf>>24&0xF, bf>>28&0xF
	m.Blends = !(srcRGB == blendOne && dstRGB == blendZero && srcA == blendOne && dstA == blendZero)
	at := regs[regAlphaTest]
	m.AlphaTest = at&1 != 0
	m.AlphaFunc = uint8(at >> 4 & 7)
	m.AlphaRef = uint8(at >> 8)
	return nil
}
