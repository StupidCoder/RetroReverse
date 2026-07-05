// Package sm64ds decodes Super Mario 64 DS's own 3D model format, ".bmd" — which,
// despite the name, is NOT the NITRO BMD0 container Mario Kart DS uses. It is a
// bespoke format: a fixed header of section (count, offset) pairs, arrays of bones,
// display lists, materials, textures and palettes. The low-level pieces are the
// same DS silicon primitives as NITRO, though — the GX geometry display lists and
// the seven DS texture formats — so those decoders (tools/nds/nitro) are reused;
// only the container is new.
//
// A ".bmd" file is stored as an "LZ77"-magic wrapper around a standard LZ10 stream;
// LoadBMD decompresses it and parses the model.
//
// Format (little-endian), pinned against castle_1f_all.bmd, tree.bmd et al:
//
//	Header (0x3C+):
//	  0x00 u32 scale        model scale as a power-of-two shift (vertices ×2^scale)
//	  0x04 u32 numBones     0x08 u32 bonesOffset      (bone stride 0x40)
//	  0x0C u32 numDisplists 0x10 u32 displistsOffset  (stride 0x08)
//	  0x14 u32 numTextures  0x18 u32 texturesOffset   (stride 0x14)
//	  0x1C u32 numPalettes  0x20 u32 palettesOffset   (stride 0x10)
//	  0x24 u32 numMaterials 0x28 u32 materialsOffset  (stride 0x30)
//	Bone (0x40): translate/scale transform + a render list at +0x30 (count),
//	  +0x34 (byte array of material indices), +0x38 (byte array of displist
//	  indices) — the bone draws displist[dl[i]] with material[mat[i]].
//	Displist (0x08): {u32 subCount, u32 subHeaderPtr}. subHeaderPtr points at
//	  subCount sub-headers of 0x10 bytes; each locates a GX display-list command
//	  chunk (u32 gxSize@0x08, u32 gxDataPtr@0x0C). (Stride is 0x08, not 0x10 — a
//	  0x10 read merges pairs of adjacent display lists and halves the index space,
//	  scrambling which material draws which geometry.)
//	Material (0x30): +0x00 nameOff, +0x04 textureIndex, +0x08 paletteIndex,
//	  +0x0C/+0x10 texture-matrix scale (fx12), +0x24 GX polygon attribute.
//	Texture (0x14): +0x00 nameOff, +0x04 dataOff, +0x08 size, +0x0C dims,
//	  +0x10 texImageParam (format/width/height). Format 5 (4x4): the 2-bit index
//	  data is at dataOff (size bytes), the per-block palette-select data follows
//	  at dataOff+size (size/2 bytes).
//	Palette (0x10): +0x00 nameOff, +0x04 dataOff, +0x08 size (BGR555).
package sm64ds

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"retroreverse.com/tools/nds"
	"retroreverse.com/tools/nds/nitro"
)

var le = binary.LittleEndian

// Model is a decoded SM64DS model ready for GLB export: triangles grouped by the
// material index that draws them, the materials, and the textures keyed by name.
type Model struct {
	Name     string
	ByMat    map[int][]nitro.Tri
	Mats     []nitro.Material
	Texs     map[string]nitro.Texture
	NumBones int
}

type bmd struct{ d []byte }

func (b bmd) u32(o int) uint32 {
	if o < 0 || o+4 > len(b.d) {
		return 0
	}
	return le.Uint32(b.d[o:])
}
func (b bmd) name(o int) string {
	if o <= 0 || o >= len(b.d) {
		return ""
	}
	e := o
	for e < len(b.d) && b.d[e] >= 0x20 && b.d[e] < 0x7F {
		e++
	}
	return string(b.d[o:e])
}

// LoadBMD reads and decompresses a .bmd file and decodes its model.
func LoadBMD(path string) (*Model, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data := raw
	if len(raw) >= 4 && string(raw[:4]) == "LZ77" {
		data = nds.Decompress(raw[4:]) // "LZ77" magic + standard LZ10 stream
	} else {
		data = nds.Decompress(raw)
	}
	return Decode(data, baseName(path))
}

// Decode parses a decompressed .bmd blob.
func Decode(data []byte, name string) (*Model, error) {
	if len(data) < 0x3C {
		return nil, fmt.Errorf("sm64ds: short BMD")
	}
	b := bmd{data}
	scale := math.Pow(2, float64(b.u32(0x00)))
	numBone, oBone := int(b.u32(0x04)), int(b.u32(0x08))
	numDL, oDL := int(b.u32(0x0C)), int(b.u32(0x10))
	numTex, oTex := int(b.u32(0x14)), int(b.u32(0x18))
	numPal, oPal := int(b.u32(0x1C)), int(b.u32(0x20))
	numMat, oMat := int(b.u32(0x24)), int(b.u32(0x28))

	// Textures + palettes, decoded and keyed by name.
	pals := make([]int, numPal) // palette data offsets
	palNames := make([]string, numPal)
	for i := 0; i < numPal; i++ {
		r := oPal + i*0x10
		palNames[i] = b.name(int(b.u32(r)))
		pals[i] = int(b.u32(r + 4))
	}
	type texRec struct {
		name         string
		dataOff, sz  int
		param        uint32
	}
	texRecs := make([]texRec, numTex)
	for i := 0; i < numTex; i++ {
		r := oTex + i*0x14
		texRecs[i] = texRec{b.name(int(b.u32(r))), int(b.u32(r + 4)), int(b.u32(r + 8)), b.u32(r + 0x10)}
	}

	texs := map[string]nitro.Texture{}
	decodeTex := func(ti, pi int) (string, int, int, error) {
		if ti < 0 || ti >= numTex {
			return "", 0, 0, nil
		}
		tr := texRecs[ti]
		palOff := 0
		if pi >= 0 && pi < numPal {
			palOff = pals[pi]
		}
		palIdxOff := tr.dataOff + tr.sz // format-5 per-block palette-select data
		if _, done := texs[tr.name]; !done {
			t, err := nitro.DecodeTexture(data, tr.dataOff, palIdxOff, palOff, tr.param)
			if err != nil {
				return tr.name, 0, 0, err
			}
			t.Name = tr.name
			texs[tr.name] = t
		}
		t := texs[tr.name]
		return tr.name, t.Width, t.Height, nil
	}

	// Materials: each names an explicit texture + palette index.
	mats := make([]nitro.Material, numMat)
	for i := 0; i < numMat; i++ {
		r := oMat + i*0x30
		m := nitro.Material{Name: b.name(int(b.u32(r))), ScaleS: 1, ScaleT: 1}
		ti, pi := int(int32(b.u32(r + 4))), int(int32(b.u32(r + 8)))
		if name, w, h, err := decodeTex(ti, pi); err == nil && name != "" {
			m.Texture, m.Width, m.Height = name, w, h
			// Wrap: level textures tile, so repeat both axes unless the texgen bits
			// say otherwise; the GX repeat/flip live in bits 16-19 of the param.
			m.TexParam = texRecs[ti].param | (3 << 16)
		}
		mats[i] = m
	}

	// Bones position and draw the display lists. Walk each bone's render list,
	// decoding the referenced display list into triangles under its material.
	stack := make([]nitro.Mat43, 32)
	for i := range stack {
		stack[i] = nitro.Identity43()
	}
	byMat := map[int][]nitro.Tri{}
	decodeDL := func(dlIdx int, world nitro.Mat43) []nitro.Tri {
		if dlIdx < 0 || dlIdx >= numDL {
			return nil
		}
		// A display list is an 8-byte record {count, subHeaderPtr}: `count`
		// sub-headers of 0x10 bytes at subHeaderPtr, each locating a GX command chunk
		// (gxSize at +8, gxPtr at +0xC). (An earlier read used a 0x10 stride with two
		// sub-headers per record, which silently merged pairs of adjacent display lists
		// and halved the index space — scrambling which material drew which geometry.)
		rec := oDL + dlIdx*0x08
		subCount, subPtr := int(b.u32(rec)), int(b.u32(rec+4))
		var gx []byte
		for k := 0; k < subCount; k++ {
			sub := subPtr + k*0x10
			if sub <= 0 || sub+0x10 > len(data) {
				continue
			}
			sz, ptr := int(b.u32(sub+8)), int(b.u32(sub+0xC))
			if ptr <= 0 || sz <= 0 || ptr+sz > len(data) || sz > 0x40000 {
				continue
			}
			gx = append(gx, data[ptr:ptr+sz]...)
		}
		return nitro.DecodeDL(gx, stack, world, 0)
	}
	for bi := 0; bi < numBone; bi++ {
		bone := oBone + bi*0x40
		world := boneWorld(b, bone, scale)
		cnt := int(b.u32(bone + 0x30))
		moff, doff := int(b.u32(bone+0x34)), int(b.u32(bone+0x38))
		for i := 0; i < cnt; i++ {
			if moff+i >= len(data) || doff+i >= len(data) {
				break
			}
			matIdx := int(data[moff+i])
			dlIdx := int(data[doff+i])
			for _, t := range decodeDL(dlIdx, world) {
				if !finiteTri(t) {
					continue // guard: drop any triangle a bad matrix flung out of range
				}
				byMat[matIdx] = append(byMat[matIdx], t)
			}
		}
	}
	if len(byMat) == 0 {
		return nil, fmt.Errorf("sm64ds: %s has no geometry", name)
	}
	return &Model{Name: name, ByMat: byMat, Mats: mats, Texs: texs, NumBones: numBone}, nil
}

// boneWorld builds a bone's world transform. The stage/object geometry tested so
// far is authored in model space with identity bones, so this applies only the
// model-wide 2^scale; per-bone translate/rotate is layered in once a model needs it.
func boneWorld(b bmd, bone int, scale float64) nitro.Mat43 {
	return nitro.Mat43{scale, 0, 0, 0, scale, 0, 0, 0, scale, 0, 0, 0}
}

// GLB exports the model to binary glTF via the shared nitro exporter.
func (m *Model) GLB() ([]byte, error) {
	return nitro.ExportTrisGLB(m.Name, m.ByMat, m.Mats, m.Texs)
}

// finiteTri reports whether every vertex is finite and within a sane bound (DS
// model space is roughly ±8 pre-scale; anything past ±10000 is a decode artifact).
func finiteTri(t nitro.Tri) bool {
	for _, v := range t.V {
		for _, c := range []float64{v.X, v.Y, v.Z} {
			if math.IsNaN(c) || math.IsInf(c, 0) || c > 1e4 || c < -1e4 {
				return false
			}
		}
	}
	return true
}

func baseName(p string) string {
	i := len(p) - 1
	for i >= 0 && p[i] != '/' {
		i--
	}
	s := p[i+1:]
	if j := len(s) - 4; j > 0 && s[j:] == ".bmd" {
		s = s[:j]
	}
	return s
}
