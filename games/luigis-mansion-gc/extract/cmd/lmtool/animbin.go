package main

// animbin.go exports a .bin furniture model with its .anm clips as an
// animated GLB: one glTF node per scene-graph node carrying its own mesh
// (vertices are node-local), and one glTF animation per clip driving the
// nodes' TRS — the .anm channels replace the node TRS outright, sampled
// here at the format's own 30 fps.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"sort"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

type anmClip struct {
	Name string
	Anm  *lm.Anm
}

// eulerQuat converts degrees to a quaternion via the game's Rz·Ry·Rx.
func eulerQuat(deg [3]float32) [4]float32 {
	half := func(d float32) (float32, float32) {
		a := float64(d) * math.Pi / 180 / 2
		return float32(math.Sin(a)), float32(math.Cos(a))
	}
	sx, cx := half(deg[0])
	sy, cy := half(deg[1])
	sz, cz := half(deg[2])
	return [4]float32{
		cz*cy*sx - sz*sy*cx,
		cz*sy*cx + sz*cy*sx,
		sz*cy*cx - cz*sy*sx,
		cz*cy*cx + sz*sy*sx,
	}
}

// binGLBAnimated writes the model as a node hierarchy with animations.
// With no clips it still writes the hierarchy (equivalent to the baked
// export, but instanceable per node).
func binGLBAnimated(m *lm.Bin, clips []anmClip, path, name string) error {
	w := &binWriter{}

	type prim struct {
		pos [][3]float32
		nrm [][3]float32
		uv  [][2]float32
		idx []uint32
	}

	var materials, images, textures, samplers []map[string]any
	imgIdx := map[int]int{}
	matIdx := map[int]int{}
	var meshes []map[string]any

	material := func(mi int) (int, error) {
		if gi, ok := matIdx[mi]; ok {
			return gi, nil
		}
		mat := map[string]any{
			"name":        fmt.Sprintf("mat%02d", mi),
			"doubleSided": true,
			"pbrMetallicRoughness": map[string]any{
				"baseColorFactor": []float32{1, 1, 1, 1},
				"metallicFactor":  0, "roughnessFactor": 1,
			},
			"alphaMode": "MASK", "alphaCutoff": 0.5,
		}
		bm := &m.Materials[mi]
		textured := false
		if bm.Sampler >= 0 && bm.Sampler < len(m.Samplers) {
			s := m.Samplers[bm.Sampler]
			if s.Texture >= 0 && s.Texture < len(m.Textures) {
				textured = true
				mat["extensions"] = map[string]any{"KHR_materials_unlit": struct{}{}}
				ii, ok := imgIdx[s.Texture]
				if !ok {
					t := m.Textures[s.Texture]
					img := image.NewRGBA(image.Rect(0, 0, t.Width, t.Height))
					copy(img.Pix, t.Pixels)
					var pngBuf bytes.Buffer
					if err := png.Encode(&pngBuf, img); err != nil {
						return 0, err
					}
					ii = len(images)
					imgIdx[s.Texture] = ii
					images = append(images, map[string]any{"bufferView": w.addView(pngBuf.Bytes()), "mimeType": "image/png"})
				}
				samplers = append(samplers, map[string]any{
					"magFilter": 9729, "minFilter": 9729,
					"wrapS": gxWrap(uint8(s.WrapS)), "wrapT": gxWrap(uint8(s.WrapT)),
				})
				textures = append(textures, map[string]any{"sampler": len(samplers) - 1, "source": ii})
				mat["pbrMetallicRoughness"].(map[string]any)["baseColorTexture"] = map[string]int{"index": len(textures) - 1}
			}
		}
		if !textured {
			// Untextured materials stay lit and tinted, like the baked export.
			mat["pbrMetallicRoughness"].(map[string]any)["baseColorFactor"] = []float32{
				float32(bm.Tint[0]) / 255, float32(bm.Tint[1]) / 255,
				float32(bm.Tint[2]) / 255, float32(bm.Tint[3]) / 255,
			}
		}
		materials = append(materials, mat)
		matIdx[mi] = len(materials) - 1
		return len(materials) - 1, nil
	}

	// One glTF node per graph node, meshes in node-local space.
	var nodes []map[string]any
	for ni := range m.Nodes {
		bn := &m.Nodes[ni]
		prims := map[int]*prim{}
		var order []int
		for _, pair := range bn.Pairs {
			if pair.Mesh >= len(m.Meshes) || pair.Material >= len(m.Materials) {
				continue
			}
			pr, ok := prims[pair.Material]
			if !ok {
				pr = &prim{}
				prims[pair.Material] = pr
				order = append(order, pair.Material)
			}
			dl, err := m.ParseBinDL(&m.Meshes[pair.Mesh])
			if err != nil {
				return fmt.Errorf("%s mesh %d: %w", name, pair.Mesh, err)
			}
			for _, p := range dl {
				base := len(pr.pos)
				for _, v := range p.Verts {
					pr.pos = append(pr.pos, m.Positions[v.Pos])
					if v.HasNrm && int(v.Nrm) < len(m.Normals) {
						pr.nrm = append(pr.nrm, m.Normals[v.Nrm])
					}
					if v.HasTex && int(v.Tex) < len(m.Texcoords) {
						pr.uv = append(pr.uv, m.Texcoords[v.Tex])
					}
				}
				for _, t := range p.Triangulate() {
					pr.idx = append(pr.idx, uint32(base+t[0]), uint32(base+t[1]), uint32(base+t[2]))
				}
			}
		}
		var meshPrims []map[string]any
		for _, mi := range order {
			pr := prims[mi]
			if len(pr.idx) == 0 {
				continue
			}
			attrs := map[string]any{"POSITION": w.addVec3(pr.pos)}
			if len(pr.nrm) == len(pr.pos) {
				attrs["NORMAL"] = w.addVec3(pr.nrm)
			}
			if len(pr.uv) == len(pr.pos) {
				attrs["TEXCOORD_0"] = w.addVec2(pr.uv)
			}
			gi, err := material(mi)
			if err != nil {
				return err
			}
			meshPrims = append(meshPrims, map[string]any{
				"attributes": attrs,
				"indices":    w.addIndices(pr.idx),
				"mode":       4,
				"material":   gi,
			})
		}
		q := eulerQuat(bn.Rot)
		n := map[string]any{
			"name":        fmt.Sprintf("%s_n%02d", name, ni),
			"translation": bn.Trans[:],
			"rotation":    q[:],
			"scale":       bn.Scale[:],
		}
		if len(meshPrims) > 0 {
			meshes = append(meshes, map[string]any{"name": fmt.Sprintf("%s_m%02d", name, ni), "primitives": meshPrims})
			n["mesh"] = len(meshes) - 1
		}
		var kids []int
		for c := bn.Child; c >= 0 && c < len(m.Nodes); c = m.Nodes[c].Next {
			kids = append(kids, c)
		}
		if len(kids) > 0 {
			n["children"] = kids
		}
		nodes = append(nodes, n)
	}
	var roots []int
	for ni := range m.Nodes {
		if m.Nodes[ni].Parent < 0 {
			roots = append(roots, ni)
		}
	}
	if len(roots) == 0 && len(nodes) > 0 {
		roots = []int{0}
	}

	// Animations: sample each clip's channels per frame at 30 fps.
	var anims []map[string]any
	for _, clip := range clips {
		a := clip.Anm
		frames := a.Frames
		if frames < 2 {
			frames = 2
		}
		times := make([]float32, frames)
		for f := range times {
			times[f] = float32(f) / 30
		}
		tAcc := w.addScalars(times)
		var chans, samps []map[string]any
		for ni := range m.Nodes {
			if ni >= len(a.Nodes) {
				break
			}
			ch := &a.Nodes[ni]
			trans := make([][3]float32, frames)
			rots := make([][4]float32, frames)
			scales := make([][3]float32, frames)
			for f := 0; f < frames; f++ {
				t := float32(f)
				var deg [3]float32
				for c := 0; c < 3; c++ {
					scales[f][c] = ch[c].Eval(t)
					deg[c] = ch[3+c].Eval(t)
					trans[f][c] = ch[6+c].Eval(t)
				}
				rots[f] = eulerQuat(deg)
			}
			samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec3(trans), "interpolation": "LINEAR"})
			chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": ni, "path": "translation"}})
			samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec4(rots), "interpolation": "LINEAR"})
			chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": ni, "path": "rotation"}})
			samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec3(scales), "interpolation": "LINEAR"})
			chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": ni, "path": "scale"}})
		}
		anims = append(anims, map[string]any{"name": clip.Name, "channels": chans, "samplers": samps})
	}

	doc := map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse lmtool"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         []map[string]any{{"name": name, "nodes": roots}},
		"nodes":          nodes,
		"meshes":         meshes,
		"materials":      materials,
		"accessors":      w.accessors,
		"bufferViews":    w.views,
	}
	if len(anims) > 0 {
		doc["animations"] = anims
	}
	if len(images) > 0 {
		doc["images"] = images
		doc["textures"] = textures
		doc["samplers"] = samplers
	}

	w.align(4)
	doc["buffers"] = []map[string]int{{"byteLength": w.buf.Len()}}
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}
	var out bytes.Buffer
	total := 12 + 8 + len(jsonBytes) + 8 + w.buf.Len()
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
	binary.LittleEndian.PutUint32(chunk, uint32(w.buf.Len()))
	copy(chunk[4:], "BIN\x00")
	out.Write(chunk)
	out.Write(w.buf.Bytes())
	return os.WriteFile(path, out.Bytes(), 0o644)
}

// sortClips orders clips by name so chest_0, chest_1 … come out stable.
func sortClips(clips []anmClip) {
	sort.Slice(clips, func(i, j int) bool { return clips[i].Name < clips[j].Name })
}
