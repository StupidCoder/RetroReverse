package model

import (
	"os"
	"path/filepath"
	"testing"
)

// The wooden door (model 14) is the anchor: a 128x208x8 leaf whose program
// shifts the pivot to the hinge, rotates by the angle variable [292A], builds
// the 8 box corners, draws 4 flat edge strips and textured big faces.
func TestDecodeDoor(t *testing.T) {
	exe, err := os.ReadFile(filepath.Join("..", "..", "game", "UW.EXE"))
	if err != nil {
		t.Skip("game data not present")
	}
	models, err := Decode(exe)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 64 {
		t.Fatalf("models = %d, want 64", len(models))
	}
	m := models[14]
	if m.Extents != [3]int16{192, 208, 8} {
		t.Errorf("door extents = %v", m.Extents)
	}
	if len(m.Pool) != 8 {
		t.Errorf("door verts = %d, want 8 (box corners)", len(m.Pool))
	}
	// the leaf outline: (0,0,0) (0,208,0) (128,208,0) (128,0,0) + Z+8 copies
	if v := m.Pool[1]; v != (Vertex{0, 208, 0}) {
		t.Errorf("v1 = %v", v)
	}
	if v := m.Pool[6]; v != (Vertex{128, 208, 8}) {
		t.Errorf("v6 = %v", v)
	}
	// it must draw flat edge strips and at least one textured face
	if len(m.Polys) < 4 {
		t.Errorf("door polys = %d, want >= 4 (edges + textured faces)", len(m.Polys))
	}
	// no undecoded opcodes across ALL models
	for _, mm := range models {
		for _, w := range mm.Warnings {
			if w != "empty/stub" {
				t.Errorf("model %d: %s", mm.Index, w)
			}
		}
	}
}
