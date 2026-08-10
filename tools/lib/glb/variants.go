package glb

import (
	"bytes"
	"fmt"
	"image"
	"os"
)

// variants.go is the multi-model GLB writer: one glTF *scene* per variant of a
// model, in a single .glb. glTF 2.0 allows any number of scenes with a default
// `scene` index, so a file can carry independent alternates of one subject —
// LOD levels, a shadow-caster proxy, livery recolours — while sharing a single
// BIN chunk and one deduplicated set of embedded textures. A viewer that knows
// nothing of variants (Blender, three.js's default `gltf.scene`) shows only
// the first scene; a Retro-X viewer offers the rest via the object document's
// `variants` list, which names scenes by their glTF scene name.

// ModelVariant is one variant: an independent model with its own vertex
// arrays and groups (the same shapes WriteTexturedLit takes). Images that
// compare equal by pointer across variants are embedded once.
//
// A variant is either one anonymous mesh (the flat fields) or a list of named
// Nodes — one glTF node per logical part, so a viewer can pick and isolate
// them. The two forms are mutually exclusive.
type ModelVariant struct {
	Name        string
	Positions   [][3]float32
	Normals     [][3]float32 // optional (nil = no NORMAL accessor)
	UVs         [][2]float32
	UV2         [][2]float32 // optional second UV set (nil = no TEXCOORD_1)
	Colors      [][4]uint8   // optional per-vertex COLOR_0 (nil = none)
	TexGroups   []TexturedGroup
	ColorGroups []TriGroup

	Nodes []VariantNode
}

// VariantNode is one named part of a variant.
//
// Matrix, when non-nil, becomes the glTF node's local transform, in glTF's
// column-major order — which for a row-vector engine (vertex·M, translation in
// the fourth row) is the engine matrix's own 16 floats in storage order.
// MeshKey, when non-nil, lets nodes share one mesh: the first node with a
// given key contributes the geometry, later nodes with an equal key become
// pure instances of it (their own arrays are ignored). Extras is attached to
// the node verbatim.
type VariantNode struct {
	Name        string
	Positions   [][3]float32
	Normals     [][3]float32
	UVs         [][2]float32
	UV2         [][2]float32
	Colors      [][4]uint8
	TexGroups   []TexturedGroup
	ColorGroups []TriGroup

	Matrix  *[16]float32
	MeshKey any
	Extras  map[string]any
}

// sharedTex tracks image/sampler dedupe across every primitive of a document.
type sharedTex struct {
	images       []map[string]any
	textures     []map[string]any
	samplers     []map[string]any
	samplerIndex map[[2]int]int
	imageIndex   map[image.Image]int
}

// appendTextured turns one variant's arrays into glTF primitives + materials,
// appending accessors to b and textures to st. It is the shared body of
// writeTextured (single-scene) and WriteVariantScenes.
func appendTextured(b *builder, st *sharedTex, matBase int,
	positions [][3]float32, uvs, uv2 [][2]float32, normals [][3]float32, colors [][4]uint8,
	texGroups []TexturedGroup, colorGroups []TriGroup) (prims, materials []map[string]any, err error) {

	posAcc := b.addPositions(positions)
	uvAcc := b.addUVs(uvs)
	uv2Acc := -1
	if len(uv2) > 0 {
		uv2Acc = b.addUVs(uv2)
	}
	nrmAcc := -1
	if len(normals) > 0 {
		nrmAcc = b.addVec3(normals)
	}
	colAcc := -1
	if len(colors) > 0 {
		colAcc = b.addColors(colors)
	}

	for _, g := range texGroups {
		if len(g.Tris) == 0 {
			continue
		}
		wrapS, wrapT := g.WrapS, g.WrapT
		if wrapS == 0 {
			wrapS = 33071 // CLAMP_TO_EDGE
		}
		if wrapT == 0 {
			wrapT = 33071
		}
		smp, ok := st.samplerIndex[[2]int{wrapS, wrapT}]
		if !ok {
			smp = len(st.samplers)
			st.samplerIndex[[2]int{wrapS, wrapT}] = smp
			st.samplers = append(st.samplers, map[string]any{
				"magFilter": 9728, "minFilter": 9728, "wrapS": wrapS, "wrapT": wrapT,
			})
		}
		img, ok := st.imageIndex[g.Image]
		if !ok {
			var png bytes.Buffer
			if err := encodePNG(&png, g.Image); err != nil {
				return nil, nil, err
			}
			vi := b.addView(png.Bytes())
			img = len(st.images)
			st.imageIndex[g.Image] = img
			st.images = append(st.images, map[string]any{"bufferView": vi, "mimeType": "image/png"})
		}
		st.textures = append(st.textures, map[string]any{"sampler": smp, "source": img})
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		idxAcc := b.addIndices(idx)
		prim := primitive(posAcc, idxAcc, 4, matBase+len(materials))
		attrs := map[string]int{"POSITION": posAcc, "TEXCOORD_0": uvAcc}
		if uv2Acc >= 0 {
			attrs["TEXCOORD_1"] = uv2Acc
		}
		if nrmAcc >= 0 {
			attrs["NORMAL"] = nrmAcc
		}
		if colAcc >= 0 {
			attrs["COLOR_0"] = colAcc
		}
		prim["attributes"] = attrs
		prims = append(prims, prim)
		mat := map[string]any{
			"name": "tex",
			"pbrMetallicRoughness": map[string]any{
				"baseColorTexture": map[string]int{"index": len(st.textures) - 1},
				"baseColorFactor":  []float64{1, 1, 1, 1},
				"metallicFactor":   0,
				"roughnessFactor":  1,
			},
			"extensions":  map[string]any{"KHR_materials_unlit": struct{}{}},
			"alphaMode":   "MASK",
			"alphaCutoff": 0.5,
			"doubleSided": !g.SingleSided,
		}
		if g.AlphaCutoff > 0 {
			mat["alphaCutoff"] = g.AlphaCutoff
		}
		// Opaque first: a group that also asks for BLEND is translucent, and
		// translucency is not a claim about texels that "no holes" can overrule.
		if g.Opaque {
			mat["alphaMode"] = "OPAQUE"
			delete(mat, "alphaCutoff")
		}
		if g.Blend || g.Additive || g.Decal || g.Alpha > 0 {
			mat["alphaMode"] = "BLEND"
			delete(mat, "alphaCutoff")
		}
		if g.Alpha > 0 && g.Alpha < 1 {
			mat["pbrMetallicRoughness"].(map[string]any)["baseColorFactor"] = []float64{1, 1, 1, g.Alpha}
		}
		extras := groupExtras(g.Additive, g.Sheen, g.Matcap)
		if g.Decal {
			if extras == nil {
				extras = map[string]any{}
			}
			extras["decal"] = true
		}
		if extras != nil {
			mat["extras"] = extras
		}
		materials = append(materials, mat)
	}
	for _, g := range colorGroups {
		if len(g.Tris) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		idxAcc := b.addIndices(idx)
		prim := primitive(posAcc, idxAcc, 4, matBase+len(materials))
		if nrmAcc >= 0 {
			prim["attributes"].(map[string]int)["NORMAL"] = nrmAcc
		}
		if colAcc >= 0 {
			prim["attributes"].(map[string]int)["COLOR_0"] = colAcc
		}
		prims = append(prims, prim)
		mat := unlitMaterial(g.Color, g.alphaOr1(), !g.SingleSided)
		if g.Additive || g.Blend {
			mat["alphaMode"] = "BLEND"
		}
		if extras := groupExtras(g.Additive, g.Sheen, false); extras != nil {
			mat["extras"] = extras
		}
		materials = append(materials, mat)
	}
	return prims, materials, nil
}

// groupExtras builds the Retro-X material extras for additive/sheen flags.
func groupExtras(additive, sheen, matcap bool) map[string]any {
	if !additive && !sheen && !matcap {
		return nil
	}
	extras := map[string]any{}
	if additive {
		extras["blend"] = "additive"
	}
	if sheen {
		extras["sheen"] = true
	}
	if matcap {
		extras["matcap"] = true
	}
	return extras
}

// WriteVariantScenes writes a multi-scene GLB: variant k becomes glTF scene k
// (named variant.Name) holding one node + mesh; scene 0 is the document
// default, so plain viewers show the first variant only. All variants share
// the BIN buffer and embedded images.
func WriteVariantScenes(path string, variants []ModelVariant) error {
	if len(variants) == 0 {
		return fmt.Errorf("glb: no variants")
	}
	b := &builder{}
	st := &sharedTex{samplerIndex: map[[2]int]int{}, imageIndex: map[image.Image]int{}}
	var meshes, nodes, scenes []map[string]any
	var materials []map[string]any
	meshByKey := map[any]int{}
	addNode := func(n VariantNode) (int, error) {
		mi := -1
		if n.MeshKey != nil {
			if have, ok := meshByKey[n.MeshKey]; ok {
				mi = have
			}
		}
		if mi < 0 {
			prims, mats, err := appendTextured(b, st, len(materials), n.Positions, n.UVs, n.UV2, n.Normals, n.Colors, n.TexGroups, n.ColorGroups)
			if err != nil {
				return -1, err
			}
			materials = append(materials, mats...)
			if len(prims) == 0 {
				return -1, nil
			}
			mi = len(meshes)
			meshes = append(meshes, map[string]any{"primitives": prims, "name": n.Name})
			if n.MeshKey != nil {
				meshByKey[n.MeshKey] = mi
			}
		}
		ni := len(nodes)
		node := map[string]any{"mesh": mi, "name": n.Name}
		if n.Matrix != nil {
			node["matrix"] = n.Matrix[:]
		}
		if len(n.Extras) > 0 {
			node["extras"] = n.Extras
		}
		nodes = append(nodes, node)
		return ni, nil
	}
	for _, v := range variants {
		var sceneNodes []int
		vnodes := v.Nodes
		if len(vnodes) == 0 {
			vnodes = []VariantNode{{Name: v.Name, Positions: v.Positions, Normals: v.Normals,
				UVs: v.UVs, UV2: v.UV2, Colors: v.Colors, TexGroups: v.TexGroups, ColorGroups: v.ColorGroups}}
		}
		for _, n := range vnodes {
			ni, err := addNode(n)
			if err != nil {
				return err
			}
			if ni >= 0 {
				sceneNodes = append(sceneNodes, ni)
			}
		}
		if len(sceneNodes) == 0 {
			return fmt.Errorf("glb: variant %q is empty", v.Name)
		}
		scenes = append(scenes, map[string]any{"nodes": sceneNodes, "name": v.Name})
	}
	doc := map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse tools/lib/glb"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         scenes,
		"nodes":          nodes,
		"meshes":         meshes,
		"materials":      materials,
		"accessors":      b.accessors,
		"bufferViews":    b.views,
		"buffers":        []map[string]int{{"byteLength": b.bin.Len()}},
	}
	if len(st.images) > 0 {
		doc["images"] = st.images
		doc["textures"] = st.textures
		doc["samplers"] = st.samplers
	}
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
