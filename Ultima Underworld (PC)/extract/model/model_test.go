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

// The doorframe (model 1) and pillar (model 10) adapt to the room: their upper
// vertices add an environment pool slot (v256) the emitter pre-loads with the
// tile's floor-to-ceiling vector; the portcullis (12/13) uses v128. Their
// height extent is the 1024 "variable" sentinel.
func TestCeilingAdaptiveModels(t *testing.T) {
	exe, err := os.ReadFile(filepath.Join("..", "..", "game", "UW.EXE"))
	if err != nil {
		t.Skip("game data not present")
	}
	models, _ := Decode(exe)
	has := func(m *Model, slot uint16) bool {
		for _, s := range m.EnvSlots {
			if s == slot {
				return true
			}
		}
		return false
	}
	if !has(models[1], 256) || models[1].Extents[1] != 1024 {
		t.Errorf("doorframe: env=%v extents=%v", models[1].EnvSlots, models[1].Extents)
	}
	if !has(models[10], 256) || models[10].Extents[1] != 1024 {
		t.Errorf("pillar: env=%v extents=%v", models[10].EnvSlots, models[10].Extents)
	}
	if !has(models[12], 128) {
		t.Errorf("portcullis: env=%v", models[12].EnvSlots)
	}
	if len(models[14].EnvSlots) != 0 {
		t.Errorf("door leaf should not be ceiling-adaptive: %v", models[14].EnvSlots)
	}
}
