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

// unlitMaterial returns a doubled-sided KHR_materials_unlit material with the
// given RGB base colour (alpha 1).
func unlitMaterial(color [3]float32) map[string]any {
	return map[string]any{
		"name": "wire",
		"pbrMetallicRoughness": map[string]any{
			"baseColorFactor": []float64{float64(color[0]), float64(color[1]), float64(color[2]), 1},
			"metallicFactor":  0,
			"roughnessFactor": 1,
		},
		"extensions":  map[string]any{"KHR_materials_unlit": struct{}{}},
		"doubleSided": true,
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

// document assembles the shared top-level glTF document for a single mesh with
// one primitive (the caller supplies POSITION, the optional index accessor, and
// the primitive mode).
func document(name string, b *builder, posAcc, idxAcc, mode int, color [3]float32) map[string]any {
	prim := map[string]any{
		"attributes": map[string]int{"POSITION": posAcc},
		"material":   0,
		"mode":       mode,
	}
	if idxAcc >= 0 {
		prim["indices"] = idxAcc
	}
	return map[string]any{
		"asset":          map[string]string{"version": "2.0", "generator": "retroreverse tools/lib/glb"},
		"extensionsUsed": []string{"KHR_materials_unlit"},
		"scene":          0,
		"scenes":         []map[string]any{{"nodes": []int{0}, "name": name}},
		"nodes":          []map[string]any{{"mesh": 0, "name": name}},
		"meshes":         []map[string]any{{"primitives": []map[string]any{prim}, "name": name}},
		"materials":      []map[string]any{unlitMaterial(color)},
		"accessors":      b.accessors,
		"bufferViews":    b.views,
		"buffers":        []map[string]int{{"byteLength": b.bin.Len()}},
	}
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
