package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteLinesCube writes the 12 edges of a unit cube as a LINES GLB, then
// re-reads the raw file to assert the .glb chunk framing (magic, version, two
// length-prefixed chunks) and the decoded JSON structure (one LINES primitive
// with a POSITION accessor, a line-index accessor, and an unlit material).
func TestWriteLinesCube(t *testing.T) {
	pos := [][3]float32{
		{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0},
		{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1},
	}
	edges := [][2]uint32{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, // bottom
		{4, 5}, {5, 6}, {6, 7}, {7, 4}, // top
		{0, 4}, {1, 5}, {2, 6}, {3, 7}, // verticals
	}
	path := filepath.Join(t.TempDir(), "cube.glb")
	if err := WriteLines(path, pos, edges, [3]float32{0.8, 0.9, 1}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 20 {
		t.Fatalf("file too short: %d bytes", len(data))
	}
	if string(data[0:4]) != "glTF" {
		t.Fatalf("bad magic %q", data[0:4])
	}
	if v := binary.LittleEndian.Uint32(data[4:8]); v != 2 {
		t.Fatalf("bad version %d", v)
	}
	if total := binary.LittleEndian.Uint32(data[8:12]); int(total) != len(data) {
		t.Fatalf("header total %d != file len %d", total, len(data))
	}

	// Chunk 0: JSON.
	jLen := binary.LittleEndian.Uint32(data[12:16])
	if string(data[16:20]) != "JSON" {
		t.Fatalf("chunk 0 not JSON: %q", data[16:20])
	}
	jStart := 20
	jEnd := jStart + int(jLen)
	if jEnd > len(data) {
		t.Fatalf("JSON chunk overruns file")
	}
	// Chunk 1: BIN.
	if jEnd+8 > len(data) {
		t.Fatalf("missing BIN chunk")
	}
	bLen := binary.LittleEndian.Uint32(data[jEnd : jEnd+4])
	if string(data[jEnd+4:jEnd+8]) != "BIN\x00" {
		t.Fatalf("chunk 1 not BIN: %q", data[jEnd+4:jEnd+8])
	}
	if jEnd+8+int(bLen) != len(data) {
		t.Fatalf("BIN chunk length mismatch: %d+8+%d != %d", jEnd, bLen, len(data))
	}

	var doc struct {
		Asset  struct{ Version string } `json:"asset"`
		Meshes []struct {
			Primitives []struct {
				Attributes map[string]int `json:"attributes"`
				Indices    *int           `json:"indices"`
				Material   int            `json:"material"`
				Mode       int            `json:"mode"`
			} `json:"primitives"`
		} `json:"meshes"`
		Accessors []struct {
			ComponentType int    `json:"componentType"`
			Count         int    `json:"count"`
			Type          string `json:"type"`
		} `json:"accessors"`
		Materials []struct {
			Extensions map[string]json.RawMessage `json:"extensions"`
		} `json:"materials"`
		Buffers []struct {
			ByteLength int `json:"byteLength"`
		} `json:"buffers"`
	}
	if err := json.Unmarshal(data[jStart:jEnd], &doc); err != nil {
		t.Fatalf("JSON chunk not valid: %v", err)
	}
	if doc.Asset.Version != "2.0" {
		t.Fatalf("asset.version = %q", doc.Asset.Version)
	}
	if len(doc.Meshes) != 1 || len(doc.Meshes[0].Primitives) != 1 {
		t.Fatalf("want one mesh with one primitive, got %+v", doc.Meshes)
	}
	prim := doc.Meshes[0].Primitives[0]
	if prim.Mode != 1 {
		t.Fatalf("primitive mode = %d, want 1 (LINES)", prim.Mode)
	}
	if _, ok := prim.Attributes["POSITION"]; !ok {
		t.Fatalf("primitive missing POSITION attribute")
	}
	if prim.Indices == nil {
		t.Fatalf("LINES primitive missing indices accessor")
	}
	// POSITION accessor: 8 verts, VEC3 float32.
	pa := doc.Accessors[prim.Attributes["POSITION"]]
	if pa.Count != len(pos) || pa.Type != "VEC3" || pa.ComponentType != 5126 {
		t.Fatalf("POSITION accessor = %+v", pa)
	}
	// Index accessor: 24 indices (12 edges * 2), SCALAR uint.
	ia := doc.Accessors[*prim.Indices]
	if ia.Count != len(edges)*2 || ia.Type != "SCALAR" || ia.ComponentType != 5125 {
		t.Fatalf("index accessor = %+v", ia)
	}
	if len(doc.Materials) != 1 {
		t.Fatalf("want one material, got %d", len(doc.Materials))
	}
	if _, ok := doc.Materials[0].Extensions["KHR_materials_unlit"]; !ok {
		t.Fatalf("material is not unlit: %+v", doc.Materials[0].Extensions)
	}
	if len(doc.Buffers) != 1 || doc.Buffers[0].ByteLength != int(bLen)-pad(bLen) {
		// byteLength is the logical (pre-pad) buffer size; allow the <=3 byte pad.
		if doc.Buffers[0].ByteLength > int(bLen) || doc.Buffers[0].ByteLength < int(bLen)-3 {
			t.Fatalf("buffer byteLength %d vs BIN chunk %d", doc.Buffers[0].ByteLength, bLen)
		}
	}
}

func pad(n uint32) int {
	if n%4 == 0 {
		return 0
	}
	return int(4 - n%4)
}

// TestWriteTriangles is a smoke test that the TRIANGLES path emits mode 4 and a
// valid GLB header.
func TestWriteTriangles(t *testing.T) {
	pos := [][3]float32{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}
	tris := [][3]uint32{{0, 1, 2}}
	path := filepath.Join(t.TempDir(), "tri.glb")
	if err := WriteTriangles(path, pos, tris, [3]float32{1, 0, 0}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data[0:4]) != "glTF" {
		t.Fatalf("bad magic")
	}
}

// TestWriteMixed writes one quad (two triangles) plus two decal line groups and
// asserts the mesh carries the primitives in order (TRIANGLES then LINES), one
// material per group with its own colour, all sharing the single POSITION
// accessor — and that the output is byte-stable across two writes.
func TestWriteMixed(t *testing.T) {
	pos := [][3]float32{
		{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}, // quad
		{0, 0.01, 0}, {1, 0.01, 0}, // decal A
		{0, 0.01, 1}, {1, 0.01, 1}, // decal B
	}
	tris := []TriGroup{
		{Tris: [][3]uint32{{0, 1, 2}, {0, 2, 3}}, Color: [3]float32{0.5, 0.5, 0.5}},
		{}, // empty group: skipped
	}
	lines := []LineGroup{
		{Lines: [][2]uint32{{4, 5}}, Color: [3]float32{1, 1, 0}},
		{Lines: [][2]uint32{{6, 7}}, Color: [3]float32{0, 0, 0}},
	}
	path := filepath.Join(t.TempDir(), "mixed.glb")
	if err := WriteMixed(path, pos, tris, lines); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data[0:4]) != "glTF" {
		t.Fatalf("bad magic")
	}
	jLen := binary.LittleEndian.Uint32(data[12:16])
	var doc struct {
		Meshes []struct {
			Primitives []struct {
				Attributes map[string]int `json:"attributes"`
				Material   int            `json:"material"`
				Mode       int            `json:"mode"`
			} `json:"primitives"`
		} `json:"meshes"`
		Materials []struct {
			PBR struct {
				BaseColorFactor []float64 `json:"baseColorFactor"`
			} `json:"pbrMetallicRoughness"`
			Extensions map[string]json.RawMessage `json:"extensions"`
		} `json:"materials"`
	}
	if err := json.Unmarshal(data[20:20+int(jLen)], &doc); err != nil {
		t.Fatalf("JSON chunk not valid: %v", err)
	}
	if len(doc.Meshes) != 1 || len(doc.Meshes[0].Primitives) != 3 {
		t.Fatalf("want one mesh with three primitives (empty group skipped), got %+v", doc.Meshes)
	}
	prims := doc.Meshes[0].Primitives
	wantModes := []int{4, 1, 1}
	for i, p := range prims {
		if p.Mode != wantModes[i] {
			t.Fatalf("primitive %d mode = %d, want %d", i, p.Mode, wantModes[i])
		}
		if p.Material != i {
			t.Fatalf("primitive %d material = %d, want %d", i, p.Material, i)
		}
		if p.Attributes["POSITION"] != prims[0].Attributes["POSITION"] {
			t.Fatalf("primitive %d does not share the POSITION accessor", i)
		}
	}
	if len(doc.Materials) != 3 {
		t.Fatalf("want three materials, got %d", len(doc.Materials))
	}
	wantColor := [][3]float64{{0.5, 0.5, 0.5}, {1, 1, 0}, {0, 0, 0}}
	for i, m := range doc.Materials {
		c := m.PBR.BaseColorFactor
		if len(c) != 4 || c[0] != wantColor[i][0] || c[1] != wantColor[i][1] || c[2] != wantColor[i][2] {
			t.Fatalf("material %d baseColorFactor = %v, want %v", i, c, wantColor[i])
		}
		if _, ok := m.Extensions["KHR_materials_unlit"]; !ok {
			t.Fatalf("material %d is not unlit", i)
		}
	}
	// Byte-stable: a second write of the same input is identical.
	path2 := filepath.Join(t.TempDir(), "mixed.glb")
	if err := WriteMixed(path2, pos, tris, lines); err != nil {
		t.Fatal(err)
	}
	data2, err := os.ReadFile(path2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, data2) {
		t.Fatalf("WriteMixed output not byte-stable")
	}
}

// TestWriteMixedMorph writes a quad with one morph target and a weight
// animation, and asserts every primitive carries the shared POSITION target,
// the mesh default weight, and a LINEAR weights animation with bounded input.
func TestWriteMixedMorph(t *testing.T) {
	pos := [][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}}
	deltas := [][3]float32{{0, 1, 0}, {0, 1, 0}, {0, 0, 0}, {0, 0, 0}}
	tris := []TriGroup{{Tris: [][3]uint32{{0, 1, 2}, {0, 2, 3}}, Color: [3]float32{1, 0, 0}}}
	lines := []LineGroup{{Lines: [][2]uint32{{0, 1}}, Color: [3]float32{0, 0, 0}}}
	m := &MorphAnim{
		Name: "flap", Deltas: deltas,
		Times: []float32{0, 1, 2}, Weights: []float32{1, 0, 1}, Default: 0.5,
	}
	path := filepath.Join(t.TempDir(), "morph.glb")
	if err := WriteMixedMorph(path, pos, tris, lines, m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	jLen := binary.LittleEndian.Uint32(data[12:16])
	var doc struct {
		Meshes []struct {
			Primitives []struct {
				Targets []map[string]int `json:"targets"`
			} `json:"primitives"`
			Weights []float64 `json:"weights"`
		} `json:"meshes"`
		Animations []struct {
			Samplers []struct {
				Input         int    `json:"input"`
				Output        int    `json:"output"`
				Interpolation string `json:"interpolation"`
			} `json:"samplers"`
			Channels []struct {
				Target struct {
					Node int    `json:"node"`
					Path string `json:"path"`
				} `json:"target"`
			} `json:"channels"`
		} `json:"animations"`
		Accessors []struct {
			Count int       `json:"count"`
			Min   []float64 `json:"min"`
			Max   []float64 `json:"max"`
		} `json:"accessors"`
	}
	if err := json.Unmarshal(data[20:20+int(jLen)], &doc); err != nil {
		t.Fatalf("JSON chunk not valid: %v", err)
	}
	mesh := doc.Meshes[0]
	if len(mesh.Primitives) != 2 {
		t.Fatalf("want 2 primitives, got %d", len(mesh.Primitives))
	}
	for i, p := range mesh.Primitives {
		if len(p.Targets) != 1 {
			t.Fatalf("primitive %d: want 1 morph target, got %d", i, len(p.Targets))
		}
		if _, ok := p.Targets[0]["POSITION"]; !ok {
			t.Fatalf("primitive %d target missing POSITION", i)
		}
	}
	if len(mesh.Weights) != 1 || mesh.Weights[0] != 0.5 {
		t.Fatalf("mesh weights = %v, want [0.5]", mesh.Weights)
	}
	if len(doc.Animations) != 1 {
		t.Fatalf("want one animation")
	}
	a := doc.Animations[0]
	if a.Samplers[0].Interpolation != "LINEAR" || a.Channels[0].Target.Path != "weights" {
		t.Fatalf("animation wiring wrong: %+v", a)
	}
	ta := doc.Accessors[a.Samplers[0].Input]
	if ta.Count != 3 || ta.Min[0] != 0 || ta.Max[0] != 2 {
		t.Fatalf("time accessor = %+v", ta)
	}
}

// TestWriteTrianglesMat writes two triangle groups — one default (double-sided)
// and one SingleSided — and asserts the GLB carries one primitive + one material
// per group, that both primitives share the single POSITION accessor, and that
// the single-sided group serialises doubleSided:false while the default group
// serialises doubleSided:true.
func TestWriteTrianglesMat(t *testing.T) {
	pos := [][3]float32{
		{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, // group 0
		{0, 0, 1}, {1, 0, 1}, {0, 1, 1}, // group 1
	}
	groups := []TriGroup{
		{Tris: [][3]uint32{{0, 1, 2}}, Color: [3]float32{1, 0, 0}},                    // double-sided (default)
		{Tris: [][3]uint32{{3, 4, 5}}, Color: [3]float32{0, 1, 0}, SingleSided: true}, // ceiling-style
	}
	path := filepath.Join(t.TempDir(), "mat.glb")
	if err := WriteTrianglesMat(path, pos, groups); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data[0:4]) != "glTF" {
		t.Fatalf("bad magic")
	}
	jLen := binary.LittleEndian.Uint32(data[12:16])
	var doc struct {
		Meshes []struct {
			Primitives []struct {
				Attributes map[string]int `json:"attributes"`
				Material   int            `json:"material"`
				Mode       int            `json:"mode"`
			} `json:"primitives"`
		} `json:"meshes"`
		Materials []struct {
			DoubleSided bool                       `json:"doubleSided"`
			Extensions  map[string]json.RawMessage `json:"extensions"`
		} `json:"materials"`
	}
	if err := json.Unmarshal(data[20:20+int(jLen)], &doc); err != nil {
		t.Fatalf("JSON chunk not valid: %v", err)
	}
	if len(doc.Meshes) != 1 || len(doc.Meshes[0].Primitives) != 2 {
		t.Fatalf("want one mesh with two primitives, got %+v", doc.Meshes)
	}
	if len(doc.Materials) != 2 {
		t.Fatalf("want two materials, got %d", len(doc.Materials))
	}
	// Both primitives share the one POSITION accessor and index distinct materials.
	p0, p1 := doc.Meshes[0].Primitives[0], doc.Meshes[0].Primitives[1]
	if p0.Attributes["POSITION"] != p1.Attributes["POSITION"] {
		t.Fatalf("groups should share one POSITION accessor: %d vs %d",
			p0.Attributes["POSITION"], p1.Attributes["POSITION"])
	}
	if p0.Material != 0 || p1.Material != 1 || p0.Mode != 4 || p1.Mode != 4 {
		t.Fatalf("primitive material/mode wiring wrong: %+v %+v", p0, p1)
	}
	if !doc.Materials[0].DoubleSided {
		t.Fatalf("group 0 should be double-sided (doubleSided:true)")
	}
	if doc.Materials[1].DoubleSided {
		t.Fatalf("group 1 (SingleSided) should serialise doubleSided:false")
	}
	for i, m := range doc.Materials {
		if _, ok := m.Extensions["KHR_materials_unlit"]; !ok {
			t.Fatalf("material %d is not unlit: %+v", i, m.Extensions)
		}
	}
}
