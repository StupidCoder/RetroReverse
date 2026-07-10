package glb

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
)

// RigPart is one rigid piece of an animated model: a node placed at a fixed
// Translation (its pivot) with an animatable Rotation, carrying the part's
// textured and untextured triangle groups over the shared vertex arrays. The
// vertices a part references are in the part's local space (the pivot at the
// origin), so the node's translation/rotation places and turns them — which is
// exactly how a rigid-part model with no skinning animates: rotate the part
// about its pivot, then set the pivot down at its rest position.
type RigPart struct {
	Name        string
	Translation [3]float32
	Rotation    [4]float32 // bind-pose quaternion (x, y, z, w); zero value means identity
	Tex         []TexturedGroup
	Color       []TriGroup
}

// RigTrack animates one part's rotation over a clip: a quaternion at each time.
type RigTrack struct {
	Part  int          // index into the parts slice
	Times []float32    // seconds, ascending
	Rots  [][4]float32 // one quaternion (x, y, z, w) per time
}

// RigClip is a named animation over a model's parts.
type RigClip struct {
	Name   string
	Tracks []RigTrack
}

// addRotations writes a tightly packed VEC4 float32 accessor (quaternion keys).
func (b *builder) addRotations(q [][4]float32) int {
	buf := make([]byte, 16*len(q))
	for i, r := range q {
		for c := 0; c < 4; c++ {
			binary.LittleEndian.PutUint32(buf[i*16+c*4:], math.Float32bits(r[c]))
		}
	}
	vi := b.addView(buf)
	b.accessors = append(b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(q), Type: "VEC4",
	})
	return len(b.accessors) - 1
}

// WriteRiggedAnimated writes a GLB whose parts are separate nodes under one root,
// each a rigid piece placed by its own translation/rotation, plus glTF animations
// that drive the parts' rotations. positions/uvs/colors are the shared,
// part-local vertex arrays every part indexes into (a part's group triangles are
// indices into them). Textures are shared across parts by image pointer.
func WriteRiggedAnimated(path string, positions [][3]float32, uvs [][2]float32,
	colors [][4]uint8, parts []RigPart, clips []RigClip) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	uvAcc := -1
	if len(uvs) > 0 {
		uvAcc = b.addUVs(uvs)
	}
	colAcc := -1
	if len(colors) > 0 {
		colAcc = b.addColors(colors)
	}

	var materials, images, textures, samplers []map[string]any
	samplerIndex := map[[2]int]int{}
	imageIndex := map[interface{}]int{}

	texturePrim := func(g TexturedGroup) (map[string]any, bool) {
		if len(g.Tris) == 0 {
			return nil, false
		}
		wrapS, wrapT := g.WrapS, g.WrapT
		if wrapS == 0 {
			wrapS = 33071
		}
		if wrapT == 0 {
			wrapT = 33071
		}
		smp, ok := samplerIndex[[2]int{wrapS, wrapT}]
		if !ok {
			smp = len(samplers)
			samplerIndex[[2]int{wrapS, wrapT}] = smp
			samplers = append(samplers, map[string]any{
				"magFilter": 9728, "minFilter": 9728, "wrapS": wrapS, "wrapT": wrapT,
			})
		}
		img, ok := imageIndex[g.Image]
		if !ok {
			var png bytes.Buffer
			if err := encodePNG(&png, g.Image); err != nil {
				panic(err) // encodePNG only fails on a broken image.Image
			}
			vi := b.addView(png.Bytes())
			img = len(images)
			imageIndex[g.Image] = img
			images = append(images, map[string]any{"bufferView": vi, "mimeType": "image/png"})
		}
		textures = append(textures, map[string]any{"sampler": smp, "source": img})
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		prim := primitive(posAcc, b.addIndices(idx), 4, len(materials))
		attrs := map[string]int{"POSITION": posAcc}
		if uvAcc >= 0 {
			attrs["TEXCOORD_0"] = uvAcc
		}
		if colAcc >= 0 {
			attrs["COLOR_0"] = colAcc
		}
		prim["attributes"] = attrs
		mat := map[string]any{
			"name": "tex",
			"pbrMetallicRoughness": map[string]any{
				"baseColorTexture": map[string]int{"index": len(textures) - 1},
				"baseColorFactor":  []float64{1, 1, 1, 1},
				"metallicFactor":   0, "roughnessFactor": 1,
			},
			"extensions":  map[string]any{"KHR_materials_unlit": struct{}{}},
			"alphaMode":   "MASK",
			"alphaCutoff": 0.5,
			"doubleSided": !g.SingleSided,
		}
		if g.Blend {
			mat["alphaMode"] = "BLEND"
			delete(mat, "alphaCutoff")
		}
		materials = append(materials, mat)
		return prim, true
	}

	// One mesh and one node per part; the root node parents them all.
	var meshes []map[string]any
	nodes := []map[string]any{nil} // slot 0 reserved for the root
	var childIdx []int
	for pi, part := range parts {
		var prims []map[string]any
		for _, g := range part.Tex {
			if prim, ok := texturePrim(g); ok {
				prims = append(prims, prim)
			}
		}
		for _, g := range part.Color {
			if len(g.Tris) == 0 {
				continue
			}
			idx := make([]uint32, 0, len(g.Tris)*3)
			for _, t := range g.Tris {
				idx = append(idx, t[0], t[1], t[2])
			}
			prims = append(prims, primitive(posAcc, b.addIndices(idx), 4, len(materials)))
			materials = append(materials, unlitMaterial(g.Color, !g.SingleSided))
		}

		name := part.Name
		if name == "" {
			name = "part"
		}
		node := map[string]any{
			"name":        name,
			"translation": []float64{float64(part.Translation[0]), float64(part.Translation[1]), float64(part.Translation[2])},
		}
		rot := part.Rotation
		if rot == ([4]float32{}) {
			rot = [4]float32{0, 0, 0, 1}
		}
		node["rotation"] = []float64{float64(rot[0]), float64(rot[1]), float64(rot[2]), float64(rot[3])}
		if len(prims) > 0 {
			node["mesh"] = len(meshes)
			meshes = append(meshes, map[string]any{"primitives": prims, "name": name})
		}
		nodes = append(nodes, node)
		childIdx = append(childIdx, pi+1)
		_ = pi
	}
	nodes[0] = map[string]any{"name": baseName(path), "children": childIdx}

	// Animations: a sampler (time -> quaternion) and a channel per animated part.
	var anims []map[string]any
	for _, clip := range clips {
		var samps, chans []map[string]any
		for _, tr := range clip.Tracks {
			if len(tr.Times) == 0 {
				continue
			}
			in := b.addScalars(tr.Times)
			out := b.addRotations(tr.Rots)
			samps = append(samps, map[string]any{"input": in, "output": out, "interpolation": "LINEAR"})
			chans = append(chans, map[string]any{
				"sampler": len(samps) - 1,
				"target":  map[string]any{"node": tr.Part + 1, "path": "rotation"},
			})
		}
		if len(samps) == 0 {
			continue
		}
		anims = append(anims, map[string]any{"name": clip.Name, "samplers": samps, "channels": chans})
	}

	name := baseName(path)
	doc := map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse tools/lib/glb"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         []map[string]any{{"nodes": []int{0}, "name": name}},
		"nodes":          nodes,
		"meshes":         meshes,
		"materials":      materials,
		"accessors":      b.accessors,
		"bufferViews":    b.views,
		"buffers":        []map[string]int{{"byteLength": b.bin.Len()}},
	}
	if len(images) > 0 {
		doc["images"] = images
		doc["textures"] = textures
		doc["samplers"] = samplers
	}
	if len(anims) > 0 {
		doc["animations"] = anims
	}
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
