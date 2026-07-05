package nitro

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"sort"
)

// ExportGLB serialises a decoded model as a self-contained binary glTF 2.0 (.glb):
// non-indexed triangle primitives grouped per material, with positions, normalised
// UVs and vertex colours, and the material-bound NITRO textures embedded as PNGs
// (sampler wrap modes mapped from the GX repeat/flip bits). The output loads in any
// standard glTF viewer (three.js, Blender, …).
func ExportGLB(m Model, texs map[string]Texture) ([]byte, error) {
	// Gather triangles grouped by material.
	byMat := map[int][]Tri{}
	for _, d := range RunSBC(m) {
		if d.Shape >= len(m.Shapes) {
			continue
		}
		for _, t := range DecodeDL(m.Shapes[d.Shape].DL, d.Stack, d.M, d.Mat) {
			byMat[t.Mat] = append(byMat[t.Mat], t)
		}
	}
	return ExportTrisGLB(m.Name, byMat, m.Materials, texs)
}

// ExportTrisGLB serialises pre-decoded triangles (grouped by material index) as a
// binary glTF, for model formats that carry their own scene structure instead of
// NITRO's SBC bytecode — e.g. Super Mario 64 DS's BMD, whose bones and display
// lists are decoded elsewhere into this same (byMat, materials, textures) shape.
// Joint is one skeleton node for skinned export: a local bind TRS (glTF
// conventions: quaternion xyzw, column composition T·R·S) plus the
// inverse-bind matrix (column-major 4x4) of the joint's bind world transform.
type Joint struct {
	Name      string
	Parent    int // -1 for the root
	T         [3]float64
	R         [4]float64 // quaternion xyzw
	S         [3]float64
	IBM       [16]float64
	Billboard bool // camera-facing bone (exported as node extras)
}

// SkinAnim is one clip: per-joint, per-frame local TRS, sampled at FPS.
type SkinAnim struct {
	Name string
	FPS  float64
	T    [][][3]float64 // [joint][frame]
	R    [][][4]float64 // [joint][frame] quaternion xyzw
	S    [][][3]float64
}

// Skin carries a skeleton and its clips; vertices reference joints via
// Vertex.J (the DS matrix-stack slot they were emitted under).
type Skin struct {
	Joints []Joint
	Anims  []SkinAnim
}

// ExportTrisGLB writes a static GLB from pre-transformed triangles.
func ExportTrisGLB(name string, byMat map[int][]Tri, mats []Material, texs map[string]Texture) ([]byte, error) {
	return exportGLB(name, byMat, mats, texs, nil)
}

// ExportSkinnedGLB writes a GLB whose vertices are skinned to skin.Joints and
// which carries skin.Anims as glTF animations.
func ExportSkinnedGLB(name string, byMat map[int][]Tri, mats []Material, texs map[string]Texture, skin *Skin) ([]byte, error) {
	return exportGLB(name, byMat, mats, texs, skin)
}

func toInts(v interface{}) []int {
	if v == nil {
		return nil
	}
	return v.([]int)
}

func exportGLB(name string, byMat map[int][]Tri, mats []Material, texs map[string]Texture, skin *Skin) ([]byte, error) {
	if len(byMat) == 0 {
		return nil, fmt.Errorf("nitro: model %q has no geometry", name)
	}
	m := Model{Name: name, Materials: mats}

	var bin bytes.Buffer
	type accessor struct {
		BufferView    int       `json:"bufferView"`
		ComponentType int       `json:"componentType"`
		Count         int       `json:"count"`
		Type          string    `json:"type"`
		Min           []float64 `json:"min,omitempty"`
		Max           []float64 `json:"max,omitempty"`
	}
	type bufferView struct {
		Buffer     int `json:"buffer"`
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
	}
	var views []bufferView
	var accessors []accessor

	addView := func(data []byte) int {
		for bin.Len()%4 != 0 {
			bin.WriteByte(0)
		}
		views = append(views, bufferView{Buffer: 0, ByteOffset: bin.Len(), ByteLength: len(data)})
		bin.Write(data)
		return len(views) - 1
	}
	addFloats := func(vals []float32, typ string, withMinMax bool, comps int) int {
		buf := make([]byte, 4*len(vals))
		for i, v := range vals {
			binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
		}
		vi := addView(buf)
		acc := accessor{BufferView: vi, ComponentType: 5126, Count: len(vals) / comps, Type: typ}
		if withMinMax {
			mn := make([]float64, comps)
			mx := make([]float64, comps)
			for c := 0; c < comps; c++ {
				mn[c], mx[c] = math.Inf(1), math.Inf(-1)
			}
			for i := 0; i < len(vals); i += comps {
				for c := 0; c < comps; c++ {
					v := float64(vals[i+c])
					mn[c] = math.Min(mn[c], v)
					mx[c] = math.Max(mx[c], v)
				}
			}
			acc.Min, acc.Max = mn, mx
		}
		accessors = append(accessors, acc)
		return len(accessors) - 1
	}

	// Materials and textures.
	type texRef struct {
		Index int `json:"index"`
	}
	type pbr struct {
		BaseColorTexture *texRef   `json:"baseColorTexture,omitempty"`
		BaseColorFactor  []float64 `json:"baseColorFactor,omitempty"`
		MetallicFactor   float64   `json:"metallicFactor"`
		RoughnessFactor  float64   `json:"roughnessFactor"`
	}
	type gMaterial struct {
		Name        string  `json:"name"`
		PBR         pbr     `json:"pbrMetallicRoughness"`
		AlphaMode   string  `json:"alphaMode,omitempty"`
		AlphaCutoff float64 `json:"alphaCutoff,omitempty"`
		DoubleSided bool    `json:"doubleSided"`
	}
	type gSampler struct {
		WrapS int `json:"wrapS"`
		WrapT int `json:"wrapT"`
	}
	type gImage struct {
		BufferView int    `json:"bufferView"`
		MimeType   string `json:"mimeType"`
		Name       string `json:"name,omitempty"`
	}
	type gTexture struct {
		Sampler int `json:"sampler"`
		Source  int `json:"source"`
	}
	var gmats []gMaterial
	var samplers []gSampler
	var images []gImage
	var gtexs []gTexture
	matIndex := map[int]int{} // model material index -> glTF material index
	texDims := map[int][2]int{}

	wrapMode := func(rep, flip bool) int {
		switch {
		case flip:
			return 33648 // MIRRORED_REPEAT
		case rep:
			return 10497 // REPEAT
		default:
			return 33071 // CLAMP_TO_EDGE
		}
	}

	// Stable material order: iterate the material keys sorted, so the emitted
	// materials, images, buffer views and primitives are deterministic — the same
	// model re-exports byte-for-byte (no map-iteration churn in the committed GLBs).
	var order []int
	for mi := range byMat {
		order = append(order, mi)
	}
	sort.Ints(order)
	for _, mi := range order {
		gm := gMaterial{PBR: pbr{MetallicFactor: 0, RoughnessFactor: 1}, DoubleSided: true}
		gm.Name = fmt.Sprintf("mat%d", mi)
		if mi < len(m.Materials) {
			mat := m.Materials[mi]
			gm.Name = mat.Name
			if tex, ok := texs[mat.Texture]; ok && mat.Texture != "" {
				var pngBuf bytes.Buffer
				if err := png.Encode(&pngBuf, tex.Img); err != nil {
					return nil, err
				}
				vi := addView(pngBuf.Bytes())
				images = append(images, gImage{BufferView: vi, MimeType: "image/png", Name: mat.Texture})
				samplers = append(samplers, gSampler{
					WrapS: wrapMode(mat.TexParam&(1<<16) != 0, mat.TexParam&(1<<18) != 0),
					WrapT: wrapMode(mat.TexParam&(1<<17) != 0, mat.TexParam&(1<<19) != 0),
				})
				gtexs = append(gtexs, gTexture{Sampler: len(samplers) - 1, Source: len(images) - 1})
				gm.PBR.BaseColorTexture = &texRef{Index: len(gtexs) - 1}
				// Alpha-test only textures that actually have transparent texels — a
				// blanket MASK makes opaque surfaces do a needless cut-off and can nibble
				// edges. Opaque textures stay OPAQUE.
				if hasTransparency(tex.Img) {
					gm.AlphaMode = "MASK"
					gm.AlphaCutoff = 0.5
				}
				texDims[mi] = [2]int{tex.Width, tex.Height}
			}
		}
		matIndex[mi] = len(gmats)
		gmats = append(gmats, gm)
	}

	// Primitives (one per material), non-indexed.
	type primitive struct {
		Attributes map[string]int `json:"attributes"`
		Material   int            `json:"material"`
		Mode       int            `json:"mode"`
	}
	var prims []primitive
	for _, mi := range order { // same sorted order as the materials above
		tris := byMat[mi]
		var pos, uv, col, weights []float32
		var joints []byte
		dims, hasTex := texDims[mi]
		us, vs := 1.0, 1.0
		if mi < len(m.Materials) {
			us, vs = m.Materials[mi].UVScale()
		}
		for _, t := range tris {
			for _, v := range t.V {
				pos = append(pos, float32(v.X), float32(v.Y), float32(v.Z))
				if hasTex {
					uv = append(uv, float32(v.U*us/float64(dims[0])), float32(v.V*vs/float64(dims[1])))
				}
				col = append(col, float32(v.C.R)/255, float32(v.C.G)/255, float32(v.C.B)/255)
				if skin != nil {
					j := v.J
					if j >= len(skin.Joints) {
						j = 0
					}
					joints = append(joints, byte(j), 0, 0, 0)
					weights = append(weights, 1, 0, 0, 0)
				}
			}
		}
		attrs := map[string]int{"POSITION": addFloats(pos, "VEC3", true, 3)}
		if hasTex {
			attrs["TEXCOORD_0"] = addFloats(uv, "VEC2", false, 2)
		}
		attrs["COLOR_0"] = addFloats(col, "VEC3", false, 3)
		if skin != nil {
			vi := addView(joints)
			accessors = append(accessors, accessor{BufferView: vi, ComponentType: 5121, Count: len(joints) / 4, Type: "VEC4"})
			attrs["JOINTS_0"] = len(accessors) - 1
			attrs["WEIGHTS_0"] = addFloats(weights, "VEC4", false, 4)
		}
		prims = append(prims, primitive{Attributes: attrs, Material: matIndex[mi], Mode: 4})
	}

	meshNode := map[string]interface{}{"mesh": 0, "name": m.Name}
	nodes := []map[string]interface{}{meshNode}
	sceneRoots := []int{0}
	doc := map[string]interface{}{
		"asset":       map[string]string{"version": "2.0", "generator": "retroreverse tools/nds/nitro"},
		"scene":       0,
		"scenes":      []map[string]interface{}{{"nodes": sceneRoots, "name": m.Name}},
		"nodes":       nodes,
		"meshes":      []map[string]interface{}{{"primitives": prims, "name": m.Name}},
		"materials":   gmats,
		"accessors":   accessors,
		"bufferViews": views,
		"buffers":     []map[string]int{{"byteLength": bin.Len()}},
	}
	if len(images) > 0 {
		doc["images"] = images
		doc["samplers"] = samplers
		doc["textures"] = gtexs
	}
	if skin != nil {
		base := len(nodes) // joint node indices start here
		for ji, j := range skin.Joints {
			n := map[string]interface{}{
				"name":        j.Name,
				"translation": j.T[:],
				"rotation":    j.R[:],
				"scale":       j.S[:],
			}
			if j.Billboard {
				n["extras"] = map[string]bool{"billboard": true}
			}
			nodes = append(nodes, n)
			if j.Parent >= 0 {
				pn := nodes[base+j.Parent]
				pn["children"] = append(toInts(pn["children"]), base+ji)
			} else {
				sceneRoots = append(sceneRoots, base+ji)
			}
		}
		var ibm []float32
		for _, j := range skin.Joints {
			for _, v := range j.IBM {
				ibm = append(ibm, float32(v))
			}
		}
		ibmAcc := addFloats(ibm, "MAT4", false, 16)
		jointIdx := make([]int, len(skin.Joints))
		for i := range jointIdx {
			jointIdx[i] = base + i
		}
		meshNode["skin"] = 0
		doc["skins"] = []map[string]interface{}{{"joints": jointIdx, "inverseBindMatrices": ibmAcc, "skeleton": base}}
		var anims []map[string]interface{}
		for _, a := range skin.Anims {
			frames := 0
			if len(a.T) > 0 {
				frames = len(a.T[0])
			}
			times := make([]float32, frames)
			for f := range times {
				times[f] = float32(float64(f) / a.FPS)
			}
			input := addFloats(times, "SCALAR", true, 1)
			var samplers []map[string]interface{}
			var channels []map[string]interface{}
			addChan := func(joint int, path string, out int) {
				samplers = append(samplers, map[string]interface{}{"input": input, "output": out, "interpolation": "LINEAR"})
				channels = append(channels, map[string]interface{}{
					"sampler": len(samplers) - 1,
					"target":  map[string]interface{}{"node": base + joint, "path": path},
				})
			}
			for ji := range skin.Joints {
				var tv, rv, sv []float32
				for f := 0; f < frames; f++ {
					tv = append(tv, float32(a.T[ji][f][0]), float32(a.T[ji][f][1]), float32(a.T[ji][f][2]))
					rv = append(rv, float32(a.R[ji][f][0]), float32(a.R[ji][f][1]), float32(a.R[ji][f][2]), float32(a.R[ji][f][3]))
					sv = append(sv, float32(a.S[ji][f][0]), float32(a.S[ji][f][1]), float32(a.S[ji][f][2]))
				}
				addChan(ji, "translation", addFloats(tv, "VEC3", false, 3))
				addChan(ji, "rotation", addFloats(rv, "VEC4", false, 4))
				addChan(ji, "scale", addFloats(sv, "VEC3", false, 3))
			}
			anims = append(anims, map[string]interface{}{"name": a.Name, "samplers": samplers, "channels": channels})
		}
		if len(anims) > 0 {
			doc["animations"] = anims
		}
		doc["nodes"] = nodes
		doc["scenes"] = []map[string]interface{}{{"nodes": sceneRoots, "name": m.Name}}
		doc["accessors"] = accessors
		doc["bufferViews"] = views
		doc["buffers"] = []map[string]int{{"byteLength": bin.Len()}}
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}
	binBytes := bin.Bytes()
	for len(binBytes)%4 != 0 {
		binBytes = append(binBytes, 0)
	}

	var out bytes.Buffer
	total := 12 + 8 + len(jsonBytes) + 8 + len(binBytes)
	hdr := make([]byte, 12)
	copy(hdr, "glTF")
	binary.LittleEndian.PutUint32(hdr[4:], 2)
	binary.LittleEndian.PutUint32(hdr[8:], uint32(total))
	out.Write(hdr)
	chunk := make([]byte, 8)
	binary.LittleEndian.PutUint32(chunk, uint32(len(jsonBytes)))
	copy(chunk[4:], "JSON")
	out.Write(chunk)
	out.Write(jsonBytes)
	binary.LittleEndian.PutUint32(chunk, uint32(len(binBytes)))
	copy(chunk[4:], "BIN\x00")
	out.Write(chunk)
	out.Write(binBytes)
	return out.Bytes(), nil
}

// hasTransparency reports whether an image has any fully-transparent texel — the
// signal that its material needs alpha testing rather than opaque rendering.
func hasTransparency(img *image.NRGBA) bool {
	if img == nil {
		return false
	}
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] == 0 {
			return true
		}
	}
	return false
}
