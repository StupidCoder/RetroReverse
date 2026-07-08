// Package glb is a minimal, dependency-free glTF 2.0 binary (.glb) writer: it uses
// only encoding/binary and encoding/json to emit a self-contained single-buffer
// GLB (one BIN chunk, no external .bin) that loads in any standard glTF viewer
// (three.js, Blender, …). It is deliberately tiny — just the primitive kinds the
// RetroReverse tools need (LINES for wireframes, TRIANGLES for solids) with an
// unlit base-colour material — and carries no platform-specific decoding, unlike
// tools/platform/nds/nitro's texture-and-skin GLB exporter.
//
// A .glb is a 12-byte header followed by length-prefixed chunks; this writer
// emits exactly two: a JSON chunk (the glTF document) and a BIN chunk (the packed
// vertex/index data referenced by bufferViews with buffer 0). Everything is
// little-endian and 4-byte aligned, per the spec.
package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
)

// accessor, bufferView and the doc types below mirror the glTF 2.0 JSON schema
// (only the fields this writer sets). They are marshalled straight to the JSON
// chunk.
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

// builder accumulates the BIN buffer plus its bufferViews and accessors, keeping
// the running byte offset 4-aligned as required by the glTF spec.
type builder struct {
	bin       bytes.Buffer
	views     []bufferView
	accessors []accessor
}

func (b *builder) addView(data []byte) int {
	for b.bin.Len()%4 != 0 {
		b.bin.WriteByte(0)
	}
	b.views = append(b.views, bufferView{Buffer: 0, ByteOffset: b.bin.Len(), ByteLength: len(data)})
	b.bin.Write(data)
	return len(b.views) - 1
}

// addPositions writes a tightly packed VEC3 float32 POSITION accessor (with the
// min/max bounds glTF requires on POSITION) and returns its index.
func (b *builder) addPositions(pos [][3]float32) int {
	buf := make([]byte, 12*len(pos))
	mn := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	mx := [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for i, p := range pos {
		for c := 0; c < 3; c++ {
			binary.LittleEndian.PutUint32(buf[i*12+c*4:], math.Float32bits(p[c]))
			v := float64(p[c])
			mn[c] = math.Min(mn[c], v)
			mx[c] = math.Max(mx[c], v)
		}
	}
	if len(pos) == 0 {
		mn, mx = [3]float64{0, 0, 0}, [3]float64{0, 0, 0}
	}
	vi := b.addView(buf)
	b.accessors = append(b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(pos), Type: "VEC3",
		Min: mn[:], Max: mx[:],
	})
	return len(b.accessors) - 1
}

// addIndices writes a SCALAR unsigned-int (5125) index accessor and returns it.
func (b *builder) addIndices(idx []uint32) int {
	buf := make([]byte, 4*len(idx))
	for i, v := range idx {
		binary.LittleEndian.PutUint32(buf[i*4:], v)
	}
	vi := b.addView(buf)
	b.accessors = append(b.accessors, accessor{
		BufferView: vi, ComponentType: 5125, Count: len(idx), Type: "SCALAR",
	})
	return len(b.accessors) - 1
}

// TriGroup is one triangle sub-mesh of a WriteTrianglesMat call: its own triangle
// list (each tri a triple of indices into the shared positions array), an unlit
// base colour, and a face-culling choice. SingleSided marks the emitted material
// doubleSided:false (the glTF default), which three.js honours by back-face
// culling — used for one-way geometry such as a ceiling seen only from below. The
// zero value is double-sided, matching WriteLines/WriteTriangles.
type TriGroup struct {
	Tris        [][3]uint32
	Color       [3]float32
	SingleSided bool
}

// unlitMaterial returns a KHR_materials_unlit material with the given RGB base
// colour (alpha 1). doubleSided is emitted explicitly (true keeps the legacy
// two-sided default; false lets three.js back-face cull, e.g. for ceilings).
func unlitMaterial(color [3]float32, doubleSided bool) map[string]any {
	return map[string]any{
		"name": "wire",
		"pbrMetallicRoughness": map[string]any{
			"baseColorFactor": []float64{float64(color[0]), float64(color[1]), float64(color[2]), 1},
			"metallicFactor":  0,
			"roughnessFactor": 1,
		},
		"extensions":  map[string]any{"KHR_materials_unlit": struct{}{}},
		"doubleSided": doubleSided,
	}
}

// pack finishes a document into a .glb byte stream: it marshals doc to the JSON
// chunk, appends the BIN chunk, and frames both under the 12-byte header.
func pack(doc map[string]any, bin []byte) ([]byte, error) {
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ') // JSON chunk pads with spaces
	}
	binBytes := append([]byte(nil), bin...)
	for len(binBytes)%4 != 0 {
		binBytes = append(binBytes, 0) // BIN chunk pads with zeros
	}

	var out bytes.Buffer
	total := 12 + 8 + len(jsonBytes) + 8 + len(binBytes)
	hdr := make([]byte, 12)
	copy(hdr, "glTF")
	binary.LittleEndian.PutUint32(hdr[4:], 2) // version
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

// assemble wraps a set of primitives and materials into the shared top-level
// glTF document (one mesh in one node in one scene). Each primitive's "material"
// indexes materials; a LINES/TRIANGLES writer passes one of each, WriteTrianglesMat
// passes one per group.
func assemble(name string, b *builder, prims, materials []map[string]any) map[string]any {
	return map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse tools/lib/glb"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         []map[string]any{{"nodes": []int{0}, "name": name}},
		"nodes":          []map[string]any{{"mesh": 0, "name": name}},
		"meshes":         []map[string]any{{"primitives": prims, "name": name}},
		"materials":      materials,
		"accessors":      b.accessors,
		"bufferViews":    b.views,
		"buffers":        []map[string]int{{"byteLength": b.bin.Len()}},
	}
}

// primitive builds one glTF primitive over the shared POSITION accessor with the
// given optional index accessor (idxAcc < 0 = none), draw mode, and material index.
func primitive(posAcc, idxAcc, mode, material int) map[string]any {
	prim := map[string]any{
		"attributes": map[string]int{"POSITION": posAcc},
		"material":   material,
		"mode":       mode,
	}
	if idxAcc >= 0 {
		prim["indices"] = idxAcc
	}
	return prim
}

// document assembles a single-mesh, single-primitive, single-material (double-sided)
// glTF document — the shared body of WriteLines and WriteTriangles.
func document(name string, b *builder, posAcc, idxAcc, mode int, color [3]float32) map[string]any {
	prim := primitive(posAcc, idxAcc, mode, 0)
	return assemble(name, b, []map[string]any{prim}, []map[string]any{unlitMaterial(color, true)})
}

// WriteLines writes an indexed LINES (mode 1) GLB: one mesh, one primitive, a
// POSITION accessor over positions and a line-index accessor over the flattened
// edge endpoints, with an unlit material of the given base colour. Each edge is a
// pair of indices into positions. The mesh is named after the file's base name.
func WriteLines(path string, positions [][3]float32, edges [][2]uint32, color [3]float32) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	idx := make([]uint32, 0, len(edges)*2)
	for _, e := range edges {
		idx = append(idx, e[0], e[1])
	}
	idxAcc := b.addIndices(idx)
	doc := document(baseName(path), b, posAcc, idxAcc, 1, color) // mode 1 = LINES
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteTriangles writes an indexed TRIANGLES (mode 4) GLB: one mesh, one
// primitive, a POSITION accessor and a triangle-index accessor (each tri is a
// triple of indices into positions), with an unlit material of the given colour.
func WriteTriangles(path string, positions [][3]float32, tris [][3]uint32, color [3]float32) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	idx := make([]uint32, 0, len(tris)*3)
	for _, t := range tris {
		idx = append(idx, t[0], t[1], t[2])
	}
	idxAcc := b.addIndices(idx)
	doc := document(baseName(path), b, posAcc, idxAcc, 4, color) // mode 4 = TRIANGLES
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteTrianglesMat writes a multi-material TRIANGLES (mode 4) GLB: one mesh whose
// primitives all share a single POSITION accessor over positions, one primitive
// (and one unlit material) per group. A group may be marked SingleSided so its
// material serialises doubleSided:false and three.js back-face culls it (e.g.
// ceilings); the default is double-sided, matching WriteTriangles. Groups with no
// triangles are skipped. With a single default-sided group this is equivalent to
// WriteTriangles.
func WriteTrianglesMat(path string, positions [][3]float32, groups []TriGroup) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	var prims, materials []map[string]any
	for _, g := range groups {
		if len(g.Tris) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		idxAcc := b.addIndices(idx)
		mat := len(materials)
		prims = append(prims, primitive(posAcc, idxAcc, 4, mat))
		materials = append(materials, unlitMaterial(g.Color, !g.SingleSided))
	}
	doc := assemble(baseName(path), b, prims, materials)
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LineGroup is one LINES sub-mesh of a WriteMixed call: its own edge list (each
// edge a pair of indices into the shared positions array) and an unlit base
// colour. Line materials are always double-sided (culling has no meaning for
// lines).
type LineGroup struct {
	Lines [][2]uint32
	Color [3]float32
}

// WriteMixed writes a multi-material GLB mixing filled and line geometry in one
// mesh: every primitive shares a single POSITION accessor over positions, with
// one TRIANGLES (mode 4) primitive + unlit material per TriGroup followed by one
// LINES (mode 1) primitive + unlit material per LineGroup, in caller order.
// Groups with no geometry are skipped. With only TriGroups this is equivalent to
// WriteTrianglesMat; with only one LineGroup, to WriteLines.
func WriteMixed(path string, positions [][3]float32, tris []TriGroup, lines []LineGroup) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	var prims, materials []map[string]any
	for _, g := range tris {
		if len(g.Tris) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		idxAcc := b.addIndices(idx)
		mat := len(materials)
		prims = append(prims, primitive(posAcc, idxAcc, 4, mat))
		materials = append(materials, unlitMaterial(g.Color, !g.SingleSided))
	}
	for _, g := range lines {
		if len(g.Lines) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Lines)*2)
		for _, e := range g.Lines {
			idx = append(idx, e[0], e[1])
		}
		idxAcc := b.addIndices(idx)
		mat := len(materials)
		prims = append(prims, primitive(posAcc, idxAcc, 1, mat))
		materials = append(materials, unlitMaterial(g.Color, true))
	}
	doc := assemble(baseName(path), b, prims, materials)
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// MorphAnim is one morph target with an animated weight track for
// WriteMixedMorph: per-vertex POSITION displacements (same length and order as
// the positions array), a LINEAR-interpolated weight animation (Times seconds,
// Weights 0..1), and the mesh's resting weight for viewers that do not play
// the animation.
type MorphAnim struct {
	Name    string
	Deltas  [][3]float32
	Times   []float32
	Weights []float32
	Default float32
}

// WriteMixedMorph writes a WriteMixed-style GLB whose mesh carries one morph
// target: every primitive gets the shared POSITION-displacement target, the
// mesh weight defaults to m.Default, and one animation drives the weight along
// m.Times/m.Weights with LINEAR interpolation (make the last keyframe equal
// the first for a seamless loop).
func WriteMixedMorph(path string, positions [][3]float32, tris []TriGroup, lines []LineGroup, m *MorphAnim) error {
	b := &builder{}
	posAcc := b.addPositions(positions)
	deltaAcc := b.addPositions(m.Deltas)
	var prims, materials []map[string]any
	addPrim := func(idx []uint32, mode int, mat map[string]any) {
		idxAcc := b.addIndices(idx)
		prim := primitive(posAcc, idxAcc, mode, len(materials))
		prim["targets"] = []map[string]int{{"POSITION": deltaAcc}}
		prims = append(prims, prim)
		materials = append(materials, mat)
	}
	for _, g := range tris {
		if len(g.Tris) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Tris)*3)
		for _, t := range g.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
		addPrim(idx, 4, unlitMaterial(g.Color, !g.SingleSided))
	}
	for _, g := range lines {
		if len(g.Lines) == 0 {
			continue
		}
		idx := make([]uint32, 0, len(g.Lines)*2)
		for _, e := range g.Lines {
			idx = append(idx, e[0], e[1])
		}
		addPrim(idx, 1, unlitMaterial(g.Color, true))
	}

	tAcc := b.addScalars(m.Times)
	wAcc := b.addScalars(m.Weights)
	doc := assemble(baseName(path), b, prims, materials)
	doc["meshes"].([]map[string]any)[0]["weights"] = []float64{float64(m.Default)}
	doc["animations"] = []map[string]any{{
		"name": m.Name,
		"samplers": []map[string]any{{
			"input": tAcc, "output": wAcc, "interpolation": "LINEAR",
		}},
		"channels": []map[string]any{{
			"sampler": 0,
			"target":  map[string]any{"node": 0, "path": "weights"},
		}},
	}}
	data, err := pack(doc, b.bin.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// addScalars writes a tightly packed SCALAR float32 accessor with min/max (the
// glTF spec requires bounds on animation-sampler inputs) and returns its index.
func (b *builder) addScalars(vals []float32) int {
	buf := make([]byte, 4*len(vals))
	mn, mx := math.Inf(1), math.Inf(-1)
	for i, v := range vals {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
		mn = math.Min(mn, float64(v))
		mx = math.Max(mx, float64(v))
	}
	if len(vals) == 0 {
		mn, mx = 0, 0
	}
	vi := b.addView(buf)
	b.accessors = append(b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(vals), Type: "SCALAR",
		Min: []float64{mn}, Max: []float64{mx},
	})
	return len(b.accessors) - 1
}

// baseName returns the file name without its directory or extension, for use as
// the mesh/scene name. Kept tiny to avoid a path/filepath import cost for callers.
func baseName(path string) string {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' && path[i] != '\\' {
		i--
	}
	name := path[i+1:]
	if dot := lastDot(name); dot > 0 {
		name = name[:dot]
	}
	if name == "" {
		return "mesh"
	}
	return name
}

func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
