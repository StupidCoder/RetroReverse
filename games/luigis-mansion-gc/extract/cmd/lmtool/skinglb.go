package main

// skinglb.go writes a skinned, animated GLB of an .mdl + .key pair (the
// cutscene Luigi). The mdl's node tree becomes the glTF joint hierarchy, its
// per-node matrices become the skin's inverse bind matrices, and vertices are
// emitted in bind (model) space: envelope-skinned vertices already are, and a
// rigid vertex is joint-local, so it goes through the joint's bind matrix (the
// inverse of the file's inverse-bind). The .key animation is sampled at its
// native 30 fps into LINEAR translation/rotation channels per joint.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

// invert34 inverts an affine 3x4 matrix.
func invert34(m lm.Mtx34) lm.Mtx34 {
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[4], m[5], m[6]
	g, h, i := m[8], m[9], m[10]
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	id := 1 / det
	var r lm.Mtx34
	r[0] = (e*i - f*h) * id
	r[1] = (c*h - b*i) * id
	r[2] = (b*f - c*e) * id
	r[4] = (f*g - d*i) * id
	r[5] = (a*i - c*g) * id
	r[6] = (c*d - a*f) * id
	r[8] = (d*h - e*g) * id
	r[9] = (b*g - a*h) * id
	r[10] = (a*e - b*d) * id
	tx, ty, tz := m[3], m[7], m[11]
	r[3] = -(r[0]*tx + r[1]*ty + r[2]*tz)
	r[7] = -(r[4]*tx + r[5]*ty + r[6]*tz)
	r[11] = -(r[8]*tx + r[9]*ty + r[10]*tz)
	return r
}

type binWriter struct {
	buf       bytes.Buffer
	views     []map[string]any
	accessors []map[string]any
}

func (w *binWriter) align(n int) {
	for w.buf.Len()%n != 0 {
		w.buf.WriteByte(0)
	}
}

func (w *binWriter) addView(data []byte) int {
	w.align(4)
	v := map[string]any{"buffer": 0, "byteOffset": w.buf.Len(), "byteLength": len(data)}
	w.buf.Write(data)
	w.views = append(w.views, v)
	return len(w.views) - 1
}

func (w *binWriter) addAccessor(view, compType int, typ string, count int, extra map[string]any) int {
	a := map[string]any{"bufferView": view, "componentType": compType, "type": typ, "count": count}
	for k, v := range extra {
		a[k] = v
	}
	w.accessors = append(w.accessors, a)
	return len(w.accessors) - 1
}

func f32bytes(vals []float32) []byte {
	b := make([]byte, 4*len(vals))
	for i, v := range vals {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return b
}

func (w *binWriter) addVec3(v [][3]float32) int {
	flat := make([]float32, 0, len(v)*3)
	mn := [3]float32{float32(math.Inf(1)), float32(math.Inf(1)), float32(math.Inf(1))}
	mx := [3]float32{float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1))}
	for _, p := range v {
		for c := 0; c < 3; c++ {
			flat = append(flat, p[c])
			if p[c] < mn[c] {
				mn[c] = p[c]
			}
			if p[c] > mx[c] {
				mx[c] = p[c]
			}
		}
	}
	vi := w.addView(f32bytes(flat))
	return w.addAccessor(vi, 5126, "VEC3", len(v), map[string]any{
		"min": []float32{mn[0], mn[1], mn[2]}, "max": []float32{mx[0], mx[1], mx[2]},
	})
}

func (w *binWriter) addVec2(v [][2]float32) int {
	flat := make([]float32, 0, len(v)*2)
	for _, p := range v {
		flat = append(flat, p[0], p[1])
	}
	return w.addAccessor(w.addView(f32bytes(flat)), 5126, "VEC2", len(v), nil)
}

func (w *binWriter) addVec4(v [][4]float32) int {
	flat := make([]float32, 0, len(v)*4)
	for _, p := range v {
		flat = append(flat, p[0], p[1], p[2], p[3])
	}
	return w.addAccessor(w.addView(f32bytes(flat)), 5126, "VEC4", len(v), nil)
}

func (w *binWriter) addScalars(v []float32) int {
	mn, mx := float32(math.Inf(1)), float32(math.Inf(-1))
	for _, t := range v {
		if t < mn {
			mn = t
		}
		if t > mx {
			mx = t
		}
	}
	return w.addAccessor(w.addView(f32bytes(v)), 5126, "SCALAR", len(v), map[string]any{
		"min": []float32{mn}, "max": []float32{mx},
	})
}

func (w *binWriter) addJoints(j [][4]uint16) int {
	b := make([]byte, 8*len(j))
	for i, v := range j {
		for c := 0; c < 4; c++ {
			binary.LittleEndian.PutUint16(b[i*8+c*2:], v[c])
		}
	}
	return w.addAccessor(w.addView(b), 5123, "VEC4", len(j), nil)
}

func (w *binWriter) addIndices(idx []uint32) int {
	b := make([]byte, 4*len(idx))
	for i, v := range idx {
		binary.LittleEndian.PutUint32(b[i*4:], v)
	}
	vi := w.addView(b)
	return w.addAccessor(vi, 5125, "SCALAR", len(idx), nil)
}

// skinnedGLB writes the model with its animation clip.
func skinnedGLB(m *lm.MDL, key *lm.Key, path, name, clipName string) error {
	if len(key.Tracks) != len(m.Nodes) {
		return fmt.Errorf("key has %d tracks for %d nodes", len(key.Tracks), len(m.Nodes))
	}
	w := &binWriter{}

	// Bind matrices: file matrices are world->joint (inverse bind).
	bind := make([]lm.Mtx34, len(m.Nodes))
	for i := range m.Nodes {
		bind[i] = invert34(m.Nodes[i].Mtx)
	}

	// Per-material primitive accumulation.
	type prim struct {
		pos     [][3]float32
		nrm     [][3]float32
		uv      [][2]float32
		joints  [][4]uint16
		weights [][4]float32
		idx     []uint32
	}
	prims := map[int]*prim{}
	var order []int

	// GX PN-matrix slots persist across packets: a packet's -1 entries keep
	// whatever an earlier packet loaded. Walk nodes in the game's draw order
	// (the tree) carrying the slot state.
	slots := make([]int, 10)
	for i := range slots {
		slots[i] = -1
	}
	var drawOrder []int
	var visit func(idx int)
	visit = func(idx int) {
		for idx >= 0 && idx < len(m.Nodes) {
			drawOrder = append(drawOrder, idx)
			if c := m.Nodes[idx].Child; c >= 0 {
				visit(c)
			}
			idx = m.Nodes[idx].Sibling
		}
	}
	visit(0)

	for _, ni := range drawOrder {
		for _, pair := range m.Nodes[ni].Pairs {
			if pair.Shape >= len(m.Shapes) || pair.Material >= len(m.Materials) {
				continue
			}
			shape := &m.Shapes[pair.Shape]
			nbt := shape.Flags&0x02000000 != 0
			pr, ok := prims[pair.Material]
			if !ok {
				pr = &prim{}
				prims[pair.Material] = pr
				order = append(order, pair.Material)
			}
			for pi := range shape.Packets {
				pk := &shape.Packets[pi]
				for s, mi := range pk.MtxIdx {
					if mi >= 0 && s < len(slots) {
						slots[s] = int(mi)
					}
				}
				dl, err := m.ParseDL(pk.DL, nbt)
				if err != nil {
					return fmt.Errorf("node %d: %w", ni, err)
				}
				for _, p := range dl {
					base := len(pr.pos)
					for _, v := range p.Verts {
						mi := 0
						if s := int(v.MtxSlot) / 3; s < len(slots) && slots[s] >= 0 {
							mi = slots[s]
						}
						pos := m.Positions[v.Pos]
						var nrm [3]float32
						if v.HasNrm {
							nrm = m.Normals[v.Nrm]
						}
						var js [4]uint16
						var ws [4]float32
						if mi < len(m.Nodes) {
							// Rigid: joint-local vertex -> bind space.
							pos = bind[mi].Apply(pos)
							nrm = bind[mi].ApplyVec(nrm)
							js[0], ws[0] = uint16(mi), 1
						} else {
							e := &m.Envelopes[mi-len(m.Nodes)]
							// glTF holds four influences; keep the heaviest.
							idx := make([]int, len(e.Joints))
							for c := range idx {
								idx[c] = c
							}
							for a := 0; a < len(idx); a++ {
								for b := a + 1; b < len(idx); b++ {
									if e.Weights[idx[b]] > e.Weights[idx[a]] {
										idx[a], idx[b] = idx[b], idx[a]
									}
								}
							}
							n := len(idx)
							if n > 4 {
								n = 4
							}
							var sum float32
							for c := 0; c < n; c++ {
								js[c] = e.Joints[idx[c]]
								ws[c] = e.Weights[idx[c]]
								sum += ws[c]
							}
							if sum > 0 {
								for c := 0; c < n; c++ {
									ws[c] /= sum
								}
							}
						}
						pr.pos = append(pr.pos, pos)
						pr.nrm = append(pr.nrm, norm3(nrm))
						if v.HasTex {
							pr.uv = append(pr.uv, m.Texcoords[v.Tex])
						} else {
							pr.uv = append(pr.uv, [2]float32{})
						}
						pr.joints = append(pr.joints, js)
						pr.weights = append(pr.weights, ws)
					}
					for _, t := range p.Triangulate() {
						pr.idx = append(pr.idx, uint32(base+t[0]), uint32(base+t[1]), uint32(base+t[2]))
					}
				}
			}
		}
	}

	// Joint nodes: local rest pose from the animation's first frame.
	hasScale := key.Flags&0x2 != 0
	var nodes []map[string]any
	for i := range m.Nodes {
		p := key.Eval(i, 0)
		q := p.Quat()
		n := map[string]any{
			"name":        fmt.Sprintf("joint%03d", i),
			"translation": p.Translate[:],
			"rotation":    q[:],
		}
		if hasScale {
			n["scale"] = p.Scale[:]
		}
		var kids []int
		if m.Nodes[i].Child >= 0 {
			for c := m.Nodes[i].Child; c >= 0 && c < len(m.Nodes); c = m.Nodes[c].Sibling {
				kids = append(kids, c)
			}
		}
		if len(kids) > 0 {
			n["children"] = kids
		}
		nodes = append(nodes, n)
	}

	// Inverse bind matrices, glTF column-major 4x4.
	ibm := make([]byte, 64*len(m.Nodes))
	for i, n := range m.Nodes {
		mtx := n.Mtx
		cols := [16]float32{
			mtx[0], mtx[4], mtx[8], 0,
			mtx[1], mtx[5], mtx[9], 0,
			mtx[2], mtx[6], mtx[10], 0,
			mtx[3], mtx[7], mtx[11], 1,
		}
		for j, v := range cols {
			binary.LittleEndian.PutUint32(ibm[i*64+j*4:], math.Float32bits(v))
		}
	}
	ibmAcc := w.addAccessor(w.addView(ibm), 5126, "MAT4", len(m.Nodes), nil)

	// Materials and textures.
	var materials, images, textures, samplers []map[string]any
	imgIdx := map[int]int{}
	var meshPrims []map[string]any
	for _, mi := range order {
		pr := prims[mi]
		if len(pr.idx) == 0 {
			continue
		}
		attrs := map[string]any{
			"POSITION":   w.addVec3(pr.pos),
			"NORMAL":     w.addVec3(pr.nrm),
			"TEXCOORD_0": w.addVec2(pr.uv),
			"JOINTS_0":   w.addJoints(pr.joints),
			"WEIGHTS_0":  w.addVec4(pr.weights),
		}
		mat := map[string]any{
			"name":        fmt.Sprintf("mat%02d", mi),
			"doubleSided": true,
			"extensions":  map[string]any{"KHR_materials_unlit": struct{}{}},
			"pbrMetallicRoughness": map[string]any{
				"baseColorFactor": []float32{1, 1, 1, 1},
				"metallicFactor":  0, "roughnessFactor": 1,
			},
			"alphaMode": "MASK", "alphaCutoff": 0.5,
		}
		if s := m.Materials[mi].Samplers; len(s) > 0 {
			smp := m.Samplers[s[0]]
			if smp.Texture < len(m.Textures) {
				ii, ok := imgIdx[smp.Texture]
				if !ok {
					t := m.Textures[smp.Texture]
					img := image.NewRGBA(image.Rect(0, 0, t.Width, t.Height))
					copy(img.Pix, t.Pixels)
					var pngBuf bytes.Buffer
					if err := png.Encode(&pngBuf, img); err != nil {
						return err
					}
					ii = len(images)
					imgIdx[smp.Texture] = ii
					images = append(images, map[string]any{"bufferView": w.addView(pngBuf.Bytes()), "mimeType": "image/png"})
				}
				samplers = append(samplers, map[string]any{
					"magFilter": 9729, "minFilter": 9729,
					"wrapS": gxWrap(smp.WrapS), "wrapT": gxWrap(smp.WrapT),
				})
				textures = append(textures, map[string]any{"sampler": len(samplers) - 1, "source": ii})
				mat["pbrMetallicRoughness"].(map[string]any)["baseColorTexture"] = map[string]int{"index": len(textures) - 1}
			}
		}
		materials = append(materials, mat)
		meshPrims = append(meshPrims, map[string]any{
			"attributes": attrs,
			"indices":    w.addIndices(pr.idx),
			"mode":       4,
			"material":   len(materials) - 1,
		})
	}

	// Skin + mesh node.
	joints := make([]int, len(m.Nodes))
	for i := range joints {
		joints[i] = i
	}
	meshNode := len(nodes)
	nodes = append(nodes, map[string]any{"name": name, "mesh": 0, "skin": 0})
	rootIdx := len(nodes)
	// The model's own space is y-down (the demo places it through an actor
	// matrix that flips it); a 180-degree X rotation on the root stands the
	// standalone export upright.
	nodes = append(nodes, map[string]any{
		"name": name + "_root", "children": []int{0, meshNode},
		"rotation": []float32{1, 0, 0, 0},
	})

	// Animation: sample every frame at 30 fps.
	frames := int(key.Duration()) + 1
	times := make([]float32, frames)
	for f := 0; f < frames; f++ {
		times[f] = float32(f) / 30
	}
	tAcc := w.addScalars(times)
	var chans, samps []map[string]any
	for i := range m.Nodes {
		trans := make([][3]float32, frames)
		rots := make([][4]float32, frames)
		scales := make([][3]float32, frames)
		for f := 0; f < frames; f++ {
			p := key.Eval(i, float32(f))
			trans[f] = p.Translate
			rots[f] = p.Quat()
			scales[f] = p.Scale
		}
		samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec3(trans), "interpolation": "LINEAR"})
		chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": i, "path": "translation"}})
		samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec4(rots), "interpolation": "LINEAR"})
		chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": i, "path": "rotation"}})
		if hasScale {
			samps = append(samps, map[string]any{"input": tAcc, "output": w.addVec3(scales), "interpolation": "LINEAR"})
			chans = append(chans, map[string]any{"sampler": len(samps) - 1, "target": map[string]any{"node": i, "path": "scale"}})
		}
	}

	doc := map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse lmtool"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         []map[string]any{{"name": name, "nodes": []int{rootIdx}}},
		"nodes":          nodes,
		"meshes":         []map[string]any{{"name": name, "primitives": meshPrims}},
		"skins": []map[string]any{{
			"joints": joints, "inverseBindMatrices": ibmAcc, "skeleton": 0,
		}},
		"materials":   materials,
		"accessors":   w.accessors,
		"bufferViews": w.views,
		"animations": []map[string]any{{
			"name": clipName, "channels": chans, "samplers": samps,
		}},
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

func norm3(v [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	if l == 0 {
		return [3]float32{0, 1, 0}
	}
	return [3]float32{v[0] / l, v[1] / l, v[2] / l}
}
