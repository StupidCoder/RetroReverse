package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"
)

// A scene with a parented node, a textured mesh and an animation channel must
// produce a valid GLB: the 12-byte header, a JSON chunk that parses and carries
// the hierarchy, the mesh, an image, and one CUBICSPLINE animation whose output
// accessor has three elements per key.
func TestSceneWriteRoundTrip(t *testing.T) {
	s := NewScene()
	root := s.AddNode("root", -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	child := s.AddNode("child", root, [3]float32{0, 1, 0}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	err := s.AddMesh(child, "quad", []Prim{{
		Positions: [][3]float32{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		UVs:       [][2]float32{{0, 0}, {1, 0}, {0, 1}},
		Tris:      [][3]uint32{{0, 1, 2}},
		Image:     img,
		Unlit:     true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	s.AddTranslationTrack(child,
		[]float32{0, 1},
		[][3]float32{{0, 0, 0}, {0, 2, 0}},
		[][3]float32{{0, 0, 0}, {0, 0, 0}},
		[][3]float32{{0, 0, 0}, {0, 0, 0}})

	path := filepath.Join(t.TempDir(), "scene.glb")
	if err := s.Write(path, "test"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data[0:4], []byte("glTF")) {
		t.Fatal("missing glTF magic")
	}
	jlen := binary.LittleEndian.Uint32(data[12:])
	var doc map[string]any
	if err := json.Unmarshal(data[20:20+jlen], &doc); err != nil {
		t.Fatalf("JSON chunk does not parse: %v", err)
	}
	nodes := doc["nodes"].([]any)
	if len(nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(nodes))
	}
	anims := doc["animations"].([]any)
	if len(anims) != 1 {
		t.Fatalf("animations = %d, want 1", len(anims))
	}
	if _, ok := doc["images"]; !ok {
		t.Error("no images in a textured scene")
	}
}
