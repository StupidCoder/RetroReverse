package glb

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
)

// scene.go is the scene-level GLB writer: named nodes in a hierarchy with TRS
// transforms, meshes of textured/vertex-coloured primitives attached to nodes,
// and keyframed node-translation animation. The flat writers in glb.go emit one
// mesh at the origin, which suits extracted level geometry; this API is for
// articulated scenes — the 3DS banner's bone-per-part logo is the first user.
//
// Animation is emitted as glTF CUBICSPLINE translation samplers, which are
// exactly cubic hermite: the source formats' (frame, value, slope) keys map
// loss-lessly onto (value, inTangent, outTangent) triplets, so no resampling
// or approximation happens on the way through.

// Scene accumulates nodes, meshes and animation channels, then writes one GLB.
type Scene struct {
	b         builder
	nodes     []map[string]any
	meshes    []map[string]any
	materials []map[string]any
	images    []map[string]any
	textures  []map[string]any
	samplers  []map[string]any
	channels  []map[string]any
	animSamp  []map[string]any
	skins     []map[string]any
	clips     []map[string]any
	imageIdx  map[image.Image]int
	roots     []int
}

func NewScene() *Scene {
	return &Scene{imageIdx: map[image.Image]int{}}
}

// AddNode adds a node with a TRS transform under parent (-1 for a scene root)
// and returns its index. rotation is a unit quaternion (x, y, z, w).
func (s *Scene) AddNode(name string, parent int, translation [3]float32, rotation [4]float32, scale [3]float32) int {
	n := map[string]any{"name": name}
	if translation != ([3]float32{}) {
		n["translation"] = []float32{translation[0], translation[1], translation[2]}
	}
	if rotation != ([4]float32{0, 0, 0, 1}) && rotation != ([4]float32{}) {
		n["rotation"] = []float32{rotation[0], rotation[1], rotation[2], rotation[3]}
	}
	if scale != ([3]float32{1, 1, 1}) && scale != ([3]float32{}) {
		n["scale"] = []float32{scale[0], scale[1], scale[2]}
	}
	idx := len(s.nodes)
	s.nodes = append(s.nodes, n)
	if parent >= 0 {
		p := s.nodes[parent]
		kids, _ := p["children"].([]int)
		p["children"] = append(kids, idx)
	} else {
		s.roots = append(s.roots, idx)
	}
	return idx
}

// Prim is one primitive of a mesh: indexed triangles over its own vertex
// arrays. Normals, UVs and Colors may be nil. A nil Image gives an untextured
// material tinted BaseColor; with an Image, BaseColor multiplies the texture
// (leave it {1,1,1,1}).
type Prim struct {
	Positions [][3]float32
	Normals   [][3]float32
	UVs       [][2]float32
	Colors    [][4]uint8
	Tris      [][3]uint32

	Image       image.Image
	Joints      []uint8      // per-vertex joint index (single influence) for skinning
	JointsW     [][4]uint8   // per-vertex joint indices (multi-influence); pairs with Weights
	Weights     [][4]float32 // per-vertex joint weights for JointsW
	Layer       int          // draw-order layer (source submission order); >0 emits extras.layer
	BaseColor   [4]float32
	Unlit       bool
	DoubleSided bool
	Blend       bool // alphaMode BLEND instead of MASK when the texture has alpha
	// Opaque forces alphaMode OPAQUE: the texture's alpha channel is ignored.
	// A guest format can oblige every texture to carry an alpha plane whether
	// or not the material samples it as coverage (the 3DS's ETC1A4 does), and
	// masking such a texture at any cutoff punches holes the game never has.
	Opaque      bool
	AlphaCutoff float32 // custom MASK cutoff (0 = the default 0.5); the guests' alpha tests often use 1/255
	Additive    bool    // additive blending: BLEND + Retro-X extras {blend: "additive"}
	WrapS       int     // glTF sampler wrap enums; 0 = REPEAT
	WrapT       int
}

// AddMesh attaches a mesh built from prims to a node.
func (s *Scene) AddMesh(node int, name string, prims []Prim) error {
	var out []map[string]any
	for pi := range prims {
		p := &prims[pi]
		if len(p.Positions) == 0 || len(p.Tris) == 0 {
			continue
		}
		attrs := map[string]int{"POSITION": s.b.addPositions(p.Positions)}
		if p.Normals != nil {
			attrs["NORMAL"] = s.addNormals(p.Normals)
		}
		if p.UVs != nil {
			attrs["TEXCOORD_0"] = s.b.addUVs(p.UVs)
		}
		if p.Colors != nil {
			attrs["COLOR_0"] = s.b.addColors(p.Colors)
		}
		if p.Joints != nil {
			attrs["JOINTS_0"] = s.addJoints(p.Joints)
			attrs["WEIGHTS_0"] = s.addUnitWeights(len(p.Joints))
		} else if p.JointsW != nil {
			attrs["JOINTS_0"] = s.addJoints4(p.JointsW)
			attrs["WEIGHTS_0"] = s.addWeights(p.Weights)
		}
		idx := make([]uint32, 0, len(p.Tris)*3)
		for _, t := range p.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		prim := map[string]any{
			"attributes": attrs,
			"indices":    s.b.addIndices(idx),
			"mode":       4,
			"material":   s.addMaterial(p),
		}
		out = append(out, prim)
	}
	if len(out) == 0 {
		return fmt.Errorf("glb: mesh %q has no non-empty primitives", name)
	}
	mi := len(s.meshes)
	s.meshes = append(s.meshes, map[string]any{"name": name, "primitives": out})
	s.nodes[node]["mesh"] = mi
	return nil
}

func (s *Scene) addMaterial(p *Prim) int {
	base := p.BaseColor
	if base == ([4]float32{}) {
		base = [4]float32{1, 1, 1, 1}
	}
	pbr := map[string]any{
		"baseColorFactor": []float32{base[0], base[1], base[2], base[3]},
		"metallicFactor":  0,
		"roughnessFactor": 1,
	}
	mat := map[string]any{
		"pbrMetallicRoughness": pbr,
		"doubleSided":          p.DoubleSided,
	}
	extras := map[string]any{}
	if p.Layer > 0 {
		extras["layer"] = p.Layer
	}
	if p.Additive {
		extras["blend"] = "additive"
	}
	// Blending is a property of the material, not of having a texture. A prim
	// that asks to blend and carries its coverage in COLOR_0 — a soft-edged
	// decal, a blob shadow — got no alphaMode at all while this lived inside
	// the texture branch below, and glTF's default is OPAQUE: the alpha was
	// written into the file and ignored by every renderer reading it.
	if p.Blend || p.Additive {
		mat["alphaMode"] = "BLEND"
	}
	if len(extras) > 0 {
		mat["extras"] = extras
	}
	if p.Unlit {
		mat["extensions"] = map[string]any{"KHR_materials_unlit": struct{}{}}
	}
	if p.Image != nil {
		wrapS, wrapT := p.WrapS, p.WrapT
		if wrapS == 0 {
			wrapS = 10497 // REPEAT
		}
		if wrapT == 0 {
			wrapT = 10497
		}
		s.samplers = append(s.samplers, map[string]any{
			"magFilter": 9729, "minFilter": 9729, "wrapS": wrapS, "wrapT": wrapT,
		})
		img, ok := s.imageIdx[p.Image]
		if !ok {
			var png bytes.Buffer
			if err := encodePNG(&png, p.Image); err == nil {
				vi := s.b.addView(png.Bytes())
				img = len(s.images)
				s.imageIdx[p.Image] = img
				s.images = append(s.images, map[string]any{"bufferView": vi, "mimeType": "image/png"})
			}
		}
		s.textures = append(s.textures, map[string]any{"sampler": len(s.samplers) - 1, "source": img})
		pbr["baseColorTexture"] = map[string]int{"index": len(s.textures) - 1}
		switch {
		case p.Blend || p.Additive:
			// already BLEND
		case p.Opaque:
			mat["alphaMode"] = "OPAQUE"
		default:
			mat["alphaMode"] = "MASK"
			if p.AlphaCutoff > 0 {
				mat["alphaCutoff"] = p.AlphaCutoff
			} else {
				mat["alphaCutoff"] = 0.5
			}
		}
	}
	s.materials = append(s.materials, mat)
	return len(s.materials) - 1
}

func (s *Scene) addNormals(n [][3]float32) int {
	buf := make([]byte, 12*len(n))
	for i, v := range n {
		for c := 0; c < 3; c++ {
			putF32(buf[i*12+c*4:], v[c])
		}
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(n), Type: "VEC3",
	})
	return len(s.b.accessors) - 1
}

// AddTranslationTrack adds a CUBICSPLINE translation channel on a node. times
// is in seconds; values/inTan/outTan are parallel (tangents in units/second).
func (s *Scene) AddTranslationTrack(node int, times []float32, values, inTan, outTan [][3]float32) {
	tbuf := make([]byte, 4*len(times))
	mn, mx := float32(math.Inf(1)), float32(math.Inf(-1))
	for i, t := range times {
		putF32(tbuf[i*4:], t)
		if t < mn {
			mn = t
		}
		if t > mx {
			mx = t
		}
	}
	ti := s.b.addView(tbuf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: ti, ComponentType: 5126, Count: len(times), Type: "SCALAR",
		Min: []float64{float64(mn)}, Max: []float64{float64(mx)},
	})
	tAcc := len(s.b.accessors) - 1

	// CUBICSPLINE output: (inTangent, value, outTangent) per key.
	vbuf := make([]byte, 36*len(values))
	for i := range values {
		for c := 0; c < 3; c++ {
			putF32(vbuf[i*36+c*4:], inTan[i][c])
			putF32(vbuf[i*36+12+c*4:], values[i][c])
			putF32(vbuf[i*36+24+c*4:], outTan[i][c])
		}
	}
	vi := s.b.addView(vbuf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(values) * 3, Type: "VEC3",
	})
	vAcc := len(s.b.accessors) - 1

	s.animSamp = append(s.animSamp, map[string]any{
		"input": tAcc, "output": vAcc, "interpolation": "CUBICSPLINE",
	})
	s.channels = append(s.channels, map[string]any{
		"sampler": len(s.animSamp) - 1,
		"target":  map[string]any{"node": node, "path": "translation"},
	})
}

// Write assembles and writes the GLB.
func (s *Scene) Write(path, sceneName string) error {
	doc := map[string]any{
		"asset":       map[string]string{"version": "2.0", "generator": "retroreverse"},
		"scene":       0,
		"scenes":      []map[string]any{{"name": sceneName, "nodes": s.roots}},
		"nodes":       s.nodes,
		"meshes":      s.meshes,
		"materials":   s.materials,
		"bufferViews": s.b.views,
		"accessors":   s.b.accessors,
		// The BIN chunk still has to be declared: a loader resolves a bufferView
		// through buffers[0], and without this every GLB this writer produced was
		// unloadable ("cannot read properties of undefined"). Our own PNG-render
		// checks never caught it because they read the structs, not the file.
		"buffers": []map[string]int{{"byteLength": s.b.bin.Len()}},
	}
	if len(s.images) > 0 {
		doc["images"] = s.images
		doc["textures"] = s.textures
		doc["samplers"] = s.samplers
	}
	unlit := false
	for _, m := range s.materials {
		if _, ok := m["extensions"]; ok {
			unlit = true
		}
	}
	if unlit {
		doc["extensionsUsed"] = []string{"KHR_materials_unlit"}
	}
	anims := s.clips
	if len(s.channels) > 0 {
		anims = append([]map[string]any{{
			"name": "banner", "channels": s.channels, "samplers": s.animSamp,
		}}, anims...)
	}
	if len(anims) > 0 {
		doc["animations"] = anims
	}
	if len(s.skins) > 0 {
		doc["skins"] = s.skins
	}
	data, err := pack(doc, s.b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// addJoints writes a VEC4 ubyte JOINTS_0 accessor (single influence).
func (s *Scene) addJoints(j []uint8) int {
	buf := make([]byte, 4*len(j))
	for i, v := range j {
		buf[i*4] = v
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5121, Count: len(j), Type: "VEC4",
	})
	return len(s.b.accessors) - 1
}

// addJoints4 writes a VEC4 ubyte JOINTS_0 accessor (multi-influence).
func (s *Scene) addJoints4(j [][4]uint8) int {
	buf := make([]byte, 4*len(j))
	for i, v := range j {
		buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3] = v[0], v[1], v[2], v[3]
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5121, Count: len(j), Type: "VEC4",
	})
	return len(s.b.accessors) - 1
}

// addWeights writes a VEC4 float WEIGHTS_0 accessor.
func (s *Scene) addWeights(w [][4]float32) int {
	buf := make([]byte, 16*len(w))
	for i, v := range w {
		for k := 0; k < 4; k++ {
			putF32(buf[i*16+k*4:], v[k])
		}
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(w), Type: "VEC4",
	})
	return len(s.b.accessors) - 1
}

// addUnitWeights writes a VEC4 float WEIGHTS_0 accessor of {1,0,0,0}.
func (s *Scene) addUnitWeights(n int) int {
	buf := make([]byte, 16*n)
	for i := 0; i < n; i++ {
		putF32(buf[i*16:], 1)
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: n, Type: "VEC4",
	})
	return len(s.b.accessors) - 1
}

func putF32(b []byte, v float32) {
	u := math.Float32bits(v)
	b[0], b[1], b[2], b[3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
}
