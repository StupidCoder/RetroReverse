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
//	  +0x0C/+0x10 texture-matrix scale (fx12), +0x20 GX texture addressing
//	  (repeat bits 16/17, flip bits 18/19 — ORed onto the texture's param at
//	  runtime), +0x24 GX polygon attribute.
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
	Skel     []SkelJoint // bind-pose skeleton (for skinned export)
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
		name        string
		dataOff, sz int
		param       uint32
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
		// GX polygon attribute at +$24: bits 16-20 are the polygon alpha the
		// engine sends with the material (31 = opaque; 1-30 = hardware-blended
		// translucency — the water materials use 15/16, star_base 20; the stage
		// loader at $0202B5EC gives such materials their own polygon ID).
		m.Alpha = int(b.u32(r+0x24)>>16) & 0x1F
		// Texture-matrix scale at +$0C/+$10 (fx12) — applied by the hardware
		// when the addressing word's texgen mode (bits 30-31) is 1. The coin
		// (arc0[5]) maps its 16x64 texture with scaleS=2 onto mirrored-S
		// addressing; without it the face is stretched 2x horizontally.
		if ss, st := int32(b.u32(r+0xC)), int32(b.u32(r+0x10)); ss != 0 && st != 0 {
			m.ScaleS, m.ScaleT = float64(ss)/4096, float64(st)/4096
		}
		ti, pi := int(int32(b.u32(r+4))), int(int32(b.u32(r+8)))
		if name, w, h, err := decodeTex(ti, pi); err == nil && name != "" {
			m.Texture, m.Width, m.Height = name, w, h
			// The GX TEXIMAGE_PARAM the engine sends is the material's addressing
			// word at +$20 (repeat bits 16/17, flip bits 18/19) ORed onto the
			// texture record's format/size param — traced in the render-object
			// init at $02046374, which ORs [material+$20] with [texture+$10].
			// The castle hall's sun emblem (mat_sun, flip S+T) mirrors its four
			// quarters this way; clamp materials (mat_kuppa) keep all bits clear.
			m.TexParam = texRecs[ti].param | b.u32(r+0x20)
		}
		mats[i] = m
	}

	// Bones position and draw the display lists. Walk each bone's render list,
	// decoding the referenced display list into triangles under its material.
	//
	// The 32 addressable matrix slots must hold each bone's world transform —
	// INCLUDING the model-wide 2^shift — because that is what the engine keeps
	// there: its bone compose ($0204488C) MTX_STOREs the composed matrix, whose
	// command stream ends with MTX_SCALE(1 << (header.shift + 12)), into the
	// bone's slot. A display list that begins with MTX_RESTORE (most do) starts
	// from that scaled matrix; slots initialised to identity silently drop the
	// 2^shift bake for every such model (a shift-1 stage exported at half size,
	// the shift-4 trees at 1/16).
	worlds := boneWorlds(b, oBone, numBone, scale)
	// Artifact guard bound: vertices are fx4.12 (±8) baked x2^shift, but bone
	// translations legitimately push geometry further, so keep the historical
	// ±10000 floor and only widen it for high-shift models (the shift-13
	// skyboxes reach ±65536).
	bound := math.Max(1e4, 16*scale)
	// The matrix-slot map: a display list's MTX_RESTORE indices are SLOTS, and
	// slot k belongs to bone u16map[byteList[k]] — the byte list lives in the
	// sub-header words we previously ignored ({u32 count @+0, u32 listOff @+4}),
	// the u16 map at header +$2C (the engine reads both in its multi-matrix
	// path, $02044534/$020445E8). The goomba's list is [1,0,2]: slot 1 is the
	// body, slots 0/2 the feet — identity seeding put the body on a leg matrix.
	mapOff := int(b.u32(0x2C))
	boneOfSlot := func(sub int, k int) int {
		cnt, listOff := int(b.u32(sub)), int(b.u32(sub+4))
		if k >= cnt || listOff <= 0 || listOff+cnt > len(data) {
			return -1
		}
		mi := int(data[listOff+k])
		o := mapOff + mi*2
		if mapOff <= 0 || o+2 > len(data) {
			return -1
		}
		bone := int(le.Uint16(data[o:]))
		if bone >= numBone {
			return -1
		}
		return bone
	}
	skel := make([]SkelJoint, numBone)
	for i := 0; i < numBone; i++ {
		r := oBone + i*0x40
		fxv := func(o int) float64 { return float64(int32(b.u32(o))) / 4096 }
		angv := func(o int) float64 {
			return float64(uint16(le.Uint16(b.d[o:]))<<4) / 65536 * 2 * math.Pi
		}
		par := i + int(int16(le.Uint16(b.d[r+8:])))
		if par == i || par < 0 || par >= i {
			par = -1
		}
		skel[i] = SkelJoint{
			Name:   b.name(int(b.u32(r + 4))),
			Parent: par,
			// +$3C bit 0 marks a camera-facing bone: set on the bob-omb's
			// "body_bill" and the tree quads, clear on roots and feet — the
			// flag the engine's bone compose treats as the billboard bit.
			Billboard: b.u32(r+0x3C)&1 != 0,
			S:         [3]float64{fxv(r + 0x10), fxv(r + 0x14), fxv(r + 0x18)},
			R:         [3]float64{angv(r + 0x1C), angv(r + 0x1E), angv(r + 0x20)},
			T:         [3]float64{fxv(r + 0x24), fxv(r + 0x28), fxv(r + 0x2C)},
		}
	}
	stack := make([]nitro.Mat43, 32)
	for i := range stack {
		if i < numBone {
			stack[i] = worlds[i]
		} else if numBone > 0 {
			stack[i] = worlds[0]
		}
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
		// Seed the matrix slots from the first sub-header's slot->bone list and
		// remember it to remap vertex joints below. (The record's chunks share
		// one command stream and, in the shipped models, one matrix list.)
		slotBone := map[int]int{}
		if subCount > 0 && subPtr > 0 && subPtr+0x10 <= len(data) {
			for k := 0; k < 32; k++ {
				if bone := boneOfSlot(subPtr, k); bone >= 0 {
					stack[k] = worlds[bone]
					slotBone[k] = bone
				}
			}
		}
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
		tris := nitro.DecodeDL(gx, stack, world, 0)
		// vertex joints: slot -> bone (skinned export skins by bone index)
		for i := range tris {
			for vi := range tris[i].V {
				if bone, ok := slotBone[tris[i].V[vi].J]; ok {
					tris[i].V[vi].J = bone
				}
			}
		}
		return tris
	}
	for bi := 0; bi < numBone; bi++ {
		bone := oBone + bi*0x40
		world := worlds[bi]
		cnt := int(b.u32(bone + 0x30))
		moff, doff := int(b.u32(bone+0x34)), int(b.u32(bone+0x38))
		for i := 0; i < cnt; i++ {
			if moff+i >= len(data) || doff+i >= len(data) {
				break
			}
			matIdx := int(data[moff+i])
			dlIdx := int(data[doff+i])
			for _, t := range decodeDL(dlIdx, world) {
				if !finiteTri(t, bound) {
					continue // guard: drop any triangle a bad matrix flung out of range
				}
				byMat[matIdx] = append(byMat[matIdx], t)
			}
		}
	}
	if len(byMat) == 0 {
		return nil, fmt.Errorf("sm64ds: %s has no geometry", name)
	}
	return &Model{Name: name, ByMat: byMat, Mats: mats, Texs: texs, NumBones: numBone, Skel: skel}, nil
}

// boneWorlds builds every bone's world transform from the bind pose. A bone
// record carries a full local TRS — scale at +$10 (fx12), rotation at
// +$1C/+$1E/+$20 (u16, <<4 = DS angle units; traced through the runtime copy at
// $020462D0), translation at +$24 (fx12) — and RELATIVE links at +$08/+$0A/+$0C
// (parent/sibling/child, in records; the walker $02045074 recurses over them).
// The engine's rotation builder $02045178 composes R = Rx·Ry·Rz (row-vector
// order: a vertex sees X first, then Y, then Z — its rows read
// (cy·cz, cy·sz, -sy) / (sx·sy·cz - cx·sz, ...) exactly). A vertex is
// transformed by the model-wide 2^shift first (the engine's trailing
// MTX_SCALE), then S, Rx, Ry, Rz, T, then the parent chain.
func boneWorlds(b bmd, oBone, n int, scale float64) []nitro.Mat43 {
	rot := func(rx, ry, rz float64) nitro.Mat43 {
		cx, sx := math.Cos(rx), math.Sin(rx)
		cy, sy := math.Cos(ry), math.Sin(ry)
		cz, sz := math.Cos(rz), math.Sin(rz)
		return nitro.Mat43{
			cy * cz, cy * sz, -sy,
			sx*sy*cz - cx*sz, cx*cz + sx*sy*sz, sx * cy,
			cx*sy*cz + sx*sz, cx*sy*sz - sx*cz, cx * cy,
			0, 0, 0,
		}
	}
	ang := func(o int) float64 {
		raw := uint16(le.Uint16(b.d[o:])) << 4 // runtime stores u16<<4 (STRH truncates)
		return float64(raw) / 65536 * 2 * math.Pi
	}
	fx := func(o int) float64 { return float64(int32(b.u32(o))) / 4096 }
	ws := make([]nitro.Mat43, n)
	for i := 0; i < n; i++ {
		r := oBone + i*0x40
		local := rot(ang(r+0x1C), ang(r+0x1E), ang(r+0x20))
		sx, sy, sz := fx(r+0x10), fx(r+0x14), fx(r+0x18)
		if sx != 1 || sy != 1 || sz != 1 {
			local = local.Mul(nitro.Mat43{sx, 0, 0, 0, sy, 0, 0, 0, sz, 0, 0, 0})
		}
		local[9], local[10], local[11] = fx(r+0x24), fx(r+0x28), fx(r+0x2C)
		par := i + int(int16(le.Uint16(b.d[r+8:])))
		if par != i && par >= 0 && par < i {
			ws[i] = ws[par].Mul(local)
		} else {
			ws[i] = local
		}
	}
	sc := nitro.Mat43{scale, 0, 0, 0, scale, 0, 0, 0, scale, 0, 0, 0}
	out := make([]nitro.Mat43, n)
	for i := range ws {
		out[i] = ws[i].Mul(sc)
	}
	return out
}

// GLB exports the model to binary glTF via the shared nitro exporter.
func (m *Model) GLB() ([]byte, error) {
	return nitro.ExportTrisGLB(m.Name, m.ByMat, m.Mats, m.Texs)
}

// finiteTri reports whether every vertex is finite and within the model's own
// legitimate range: DS vertices are fx4.12 (±8) and the exporter bakes 2^shift
// onto them, so anything past ~16 x 2^shift is a decode artifact (a bad matrix
// flinging a vertex out). The old fixed 1e4 bound silently discarded the
// skyboxes, which are authored at shift 13 (vertices to ±8 x 8192 = ±65536).
func finiteTri(t nitro.Tri, bound float64) bool {
	for _, v := range t.V {
		for _, c := range []float64{v.X, v.Y, v.Z} {
			if math.IsNaN(c) || math.IsInf(c, 0) || c > bound || c < -bound {
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
