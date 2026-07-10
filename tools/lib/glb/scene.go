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
	BaseColor   [4]float32
	Unlit       bool
	DoubleSided bool
	Blend       bool // alphaMode BLEND instead of MASK when the texture has alpha
	WrapS       int  // glTF sampler wrap enums; 0 = REPEAT
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
		if p.Blend {
			mat["alphaMode"] = "BLEND"
		} else {
			mat["alphaMode"] = "MASK"
			mat["alphaCutoff"] = 0.5
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
	if len(s.channels) > 0 {
		doc["animations"] = []map[string]any{{
			"name": "banner", "channels": s.channels, "samplers": s.animSamp,
		}}
	}
	data, err := pack(doc, s.b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func putF32(b []byte, v float32) {
	u := math.Float32bits(v)
	b[0], b[1], b[2], b[3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
}
