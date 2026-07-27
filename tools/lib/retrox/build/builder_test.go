package build

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"retroreverse.com/tools/lib/retrox/schema"
)

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewNRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
}

func TestBuilderWritesValidTree(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "demo")
	b := New(out, "demo")
	b.SetTitle("Demo")
	b.SetDisplay(schema.Display{Native: schema.Size{W: 160, H: 144}, TickHz: 60})

	// A tilemap level whose atlas the exporter writes as a payload.
	writePNG(t, filepath.Join(out, "levels", "lv1", "atlas.png"), 16, 8)
	b.AddLevel(schema.Asset{ID: "lv1", Name: "Act 1"}, &schema.Level{
		Type: schema.LevelTilemap,
		Tilemap: &schema.Tilemap{
			TileSize: 8, Width: 2, Height: 1,
			Atlas: schema.TileAtlas{File: "lv1/atlas.png", Cols: 2},
			Cells: []int{0, 1},
		},
	})

	if err := b.Write(); err != nil {
		t.Fatal(err)
	}

	// Manifest is valid and headered.
	data, err := os.ReadFile(filepath.Join(out, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var man schema.Manifest
	if err := json.Unmarshal(data, &man); err != nil {
		t.Fatal(err)
	}
	if man.Format != schema.FormatName || man.Version != schema.Version {
		t.Fatalf("manifest header %q/%d", man.Format, man.Version)
	}
	if len(man.Assets) != 1 || man.Assets[0].Category != "level" || man.Assets[0].File != "levels/lv1.json" {
		t.Fatalf("assets: %+v", man.Assets)
	}

	// index.json created next to the game.
	idxData, err := os.ReadFile(filepath.Join(root, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx schema.Index
	if err := json.Unmarshal(idxData, &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx.Games) != 1 || idx.Games[0] != "demo" {
		t.Fatalf("index games: %v", idx.Games)
	}

	// Re-running keeps the index deduplicated.
	b2 := New(out, "demo")
	b2.SetTitle("Demo")
	b2.SetDisplay(schema.Display{Native: schema.Size{W: 160, H: 144}, TickHz: 60})
	b2.AddLevel(schema.Asset{ID: "lv1", Name: "Act 1"}, &schema.Level{
		Type: schema.LevelTilemap,
		Tilemap: &schema.Tilemap{
			TileSize: 8, Width: 2, Height: 1,
			Atlas: schema.TileAtlas{File: "lv1/atlas.png", Cols: 2},
			Cells: []int{0, 1},
		},
	})
	if err := b2.Write(); err != nil {
		t.Fatal(err)
	}
	idxData, _ = os.ReadFile(filepath.Join(root, "index.json"))
	_ = json.Unmarshal(idxData, &idx)
	if len(idx.Games) != 1 {
		t.Fatalf("index grew on re-export: %v", idx.Games)
	}
}

func TestBuilderRejectsInvalidTree(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "demo")
	b := New(out, "demo")
	b.SetTitle("Demo")
	b.SetDisplay(schema.Display{Native: schema.Size{W: 160, H: 144}, TickHz: 60})
	// Level references an atlas nobody wrote → Write must fail validation.
	b.AddLevel(schema.Asset{ID: "lv1", Name: "Act 1"}, &schema.Level{
		Type: schema.LevelTilemap,
		Tilemap: &schema.Tilemap{
			TileSize: 8, Width: 2, Height: 1,
			Atlas: schema.TileAtlas{File: "lv1/atlas.png", Cols: 2},
			Cells: []int{0, 1},
		},
	})
	if err := b.Write(); err == nil {
		t.Fatal("invalid tree accepted")
	}
}

// glb writes a minimal GLB so model validation has something to chew on.
func TestBuilderModelObject(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "demo")
	b := New(out, "demo")
	b.SetTitle("Demo")
	b.SetDisplay(schema.Display{Native: schema.Size{W: 320, H: 240}, TickHz: 60})

	j := []byte(`{"animations":[{"name":"spin"}],"materials":[]}`)
	for len(j)%4 != 0 {
		j = append(j, ' ')
	}
	var glb bytes.Buffer
	w32 := func(v uint32) { _ = binary.Write(&glb, binary.LittleEndian, v) }
	w32(0x46546C67)
	w32(2)
	w32(uint32(12 + 8 + len(j)))
	w32(uint32(len(j)))
	w32(0x4E4F534A)
	glb.Write(j)
	p, err := b.Path("objects", "kart.glb")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, glb.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	b.AddObject(schema.Asset{ID: "kart", Name: "Kart"}, &schema.Object{
		Type: schema.ObjectModel3D, Name: "Kart", Model: "kart.glb",
		Animations: []schema.Animation{{ID: "spin", Clip: "spin", Loop: "loop", FPS: 30}},
	})
	if err := b.Write(); err != nil {
		t.Fatal(err)
	}
}
