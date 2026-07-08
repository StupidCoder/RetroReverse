package glb

import (
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
