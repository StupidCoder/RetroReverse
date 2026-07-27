package schema

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/png"
	"strings"
	"testing"
	"testing/fstest"
)

// --- fixture helpers -------------------------------------------------------

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func glbBytes(t *testing.T, doc string) []byte {
	t.Helper()
	j := []byte(doc)
	for len(j)%4 != 0 {
		j = append(j, ' ')
	}
	var buf bytes.Buffer
	w32 := func(v uint32) { _ = binary.Write(&buf, binary.LittleEndian, v) }
	w32(0x46546C67) // glTF
	w32(2)
	w32(uint32(12 + 8 + len(j)))
	w32(uint32(len(j)))
	w32(0x4E4F534A) // JSON
	buf.Write(j)
	return buf.Bytes()
}

func jsonBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// validGame builds a small but feature-rich in-memory game: a tilemap level
// with placements/pool/onClick, a scene3d level with a script + camera track,
// a sprite object, a model object, and media.
func validGame(t *testing.T) fstest.MapFS {
	t.Helper()
	visible := true
	room := 5

	man := Manifest{
		Header: NewHeader(), ID: "demo", Title: "Demo",
		Display: Display{Native: Size{W: 160, H: 144}, TickHz: 60, Filter: "gg"},
		Assets: []Asset{
			{ID: "lv1", Category: "level", Name: "Act 1", File: "levels/lv1.json"},
			{ID: "lv2", Category: "level", Name: "Scene", File: "levels/lv2.json"},
			{ID: "crab", Category: "object", Name: "Crab", File: "objects/crab.json"},
			{ID: "kart", Category: "object", Name: "Kart", File: "objects/kart.json"},
			{ID: "song", Category: "music", Name: "Song", File: "music/song.mp3", Loop: true},
			{ID: "ring", Category: "sfx", Name: "Ring", File: "sfx/ring.mp3"},
			{ID: "title", Category: "picture", Name: "Title", File: "pictures/title.png", W: 32, H: 16},
		},
	}

	lv1 := Level{
		Header: NewHeader(), Type: LevelTilemap, Music: "song",
		Tilemap: &Tilemap{
			TileSize: 8, Width: 4, Height: 2,
			Atlas: TileAtlas{File: "lv1/atlas.png", Cols: 4},
			Cells: []int{0, 1, 2, 3, 3, 2, 1, 0},
			TileAnims: []TileAnim{{Tiles: []int{2, 3}, Frames: [][]int{{3, 2}, {2, 3}}, PeriodFrames: 10}},
		},
		Collision: &Collision{Kind: "grid", Sub: 1, Solid: []int{0, 1, 1, 0}},
		Placements: []Placement{
			{ID: 1, Object: "crab", Pos: []float64{10, 20}, Anim: "walk",
				OnClick: &OnClick{Action: "animate", Clip: "walk", SFX: []SFXCue{{ID: "ring"}}}},
			{ID: 2, Object: "crab", Pos: []float64{40, 20},
				OnClick: &OnClick{Action: "text", Title: "Sign", Body: "Hello"}},
		},
		Pools: []Pool{{ID: "p1", Count: 1, Object: "crab",
			Candidates: [][]float64{{1, 2}, {3, 4}}, Seedable: true}},
	}

	track := CameraTrack{Header: NewHeader(), Frames: 2, FPS: 30,
		Track: []CamSample{
			{Pos: []float64{0, 0, 10}, Target: []float64{0, 0, 0}},
			{Pos: []float64{0, 0, 9}, Target: []float64{0, 0, 0}},
		}}

	script := Script{Header: NewHeader(), Name: "Intro", FPS: 30, Shots: []Shot{{
		ID: "s1", Frames: 2, Layers: []string{"terrain"},
		Actors: []Actor{{Placement: 7, Clip: "spin"}},
		// intro.json lives in levels/lv2/, so refs resolve from there.
		Camera: ShotCamera{TrackFile: "cam.json"},
		Sounds: []SoundCue{{SFX: "ring", Start: 0}},
	}}}

	lv2 := Level{
		Header: NewHeader(), Type: LevelScene3D,
		Camera: &Camera{Mode: "fly", Pos: []float64{0, 5, 10}, Target: []float64{0, 0, 0},
			Fly: &Fly{Speed: 100}},
		Scene: &Scene{
			Layers: []Layer{
				{ID: "terrain", File: "lv2/geo/terrain.glb"},
				{ID: "sky", File: "lv2/geo/sky.glb", Mode: "toggle", Attach: "camera", RenderOrder: -100},
				{ID: "streamA", File: "lv2/geo/a.glb", Mode: "exclusive:world"},
				{ID: "streamB", File: "lv2/geo/b.glb", Mode: "exclusive:world", Visible: boolPtr(false)},
			},
			Rooms: &Rooms{Areas: []Area{{ID: "ground", Name: "Ground"}}, List: []Room{
				{ID: 5, File: "lv2/rooms/r05.glb", Area: "ground"}}},
		},
		Variants: []Variant{{ID: "day", Name: "Day", Default: true}, {ID: "night", Name: "Night"}},
		Routes:   []Route{{ID: "lap", Loop: true, Points: [][]float64{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}}}},
		Placements: []Placement{
			{ID: 7, Object: "kart", Pos: []float64{0, 0, 0}, Anim: "spin",
				Room: &room, Variants: []string{"day"},
				Route: &RouteRef{ID: "lap", Speed: 100}},
			{ID: 8, Object: "kart", Matrix: identity16()},
		},
		Scripts: []ScriptRef{{ID: "intro", Name: "Intro", File: "lv2/intro.json"}},
	}
	_ = visible

	crab := Object{Header: NewHeader(), Type: ObjectSprite2D, Name: "Crab",
		Atlas: &SpriteAtlas{File: "crab.png", CellW: 16, CellH: 16, Anchor: []int{8, 16}},
		Animations: []Animation{
			{ID: "walk", Loop: "loop", Row: 0, Frames: 2, Durations: []int{8, 8},
				Events: []AnimEvent{{Frame: 1, SFX: "ring"}}},
			{ID: "idle", Loop: "hold", Row: 1, Frames: 1},
		}}

	kart := Object{Header: NewHeader(), Type: ObjectModel3D, Name: "Kart",
		Model: "kart.glb", Instanced: true,
		Animations: []Animation{{ID: "spin", Clip: "spin", Loop: "loop", FPS: 30}},
		UVAnims:    []UVAnim{{Material: "water", Frames: 4, TransS: &Channel{Samples: []float64{0, 1}, Step: 2}}}}

	glb := glbBytes(t, `{"animations":[{"name":"spin"}],"materials":[{"name":"water"}]}`)

	return fstest.MapFS{
		"manifest.json":        {Data: jsonBytes(t, &man)},
		"levels/lv1.json":      {Data: jsonBytes(t, &lv1)},
		"levels/lv1/atlas.png": {Data: pngBytes(t, 32, 8)}, // 4 cols × 1 row of 8px tiles
		"levels/lv2.json":      {Data: jsonBytes(t, &lv2)},
		"levels/lv2/geo/terrain.glb": {Data: glb},
		"levels/lv2/geo/sky.glb":     {Data: glb},
		"levels/lv2/geo/a.glb":       {Data: glb},
		"levels/lv2/geo/b.glb":       {Data: glb},
		"levels/lv2/rooms/r05.glb":   {Data: glb},
		"levels/lv2/cam.json":        {Data: jsonBytes(t, &track)},
		"levels/lv2/intro.json":      {Data: jsonBytes(t, &script)},
		"objects/crab.json":          {Data: jsonBytes(t, &crab)},
		"objects/crab.png":           {Data: pngBytes(t, 32, 32)}, // 2×2 cells of 16px
		"objects/kart.json":          {Data: jsonBytes(t, &kart)},
		"objects/kart.glb":           {Data: glb},
		"music/song.mp3":             {Data: []byte("mp3")},
		"sfx/ring.mp3":               {Data: []byte("mp3")},
		"pictures/title.png":         {Data: pngBytes(t, 32, 16)},
	}
}

func boolPtr(b bool) *bool { return &b }

func identity16() []float64 {
	return []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func errorsOf(issues []Issue) []string {
	var out []string
	for _, i := range issues {
		if i.Level == "error" {
			out = append(out, i.String())
		}
	}
	return out
}

// --- tests -----------------------------------------------------------------

func TestValidGamePasses(t *testing.T) {
	fsys := validGame(t)
	if errs := errorsOf(ValidateGame(fsys, "demo")); len(errs) != 0 {
		t.Fatalf("valid game reported errors:\n%s", strings.Join(errs, "\n"))
	}
}

// mutate reserializes one document with a tweak and expects a specific error.
func expectError(t *testing.T, fsys fstest.MapFS, substr string) {
	t.Helper()
	errs := errorsOf(ValidateGame(fsys, "demo"))
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Fatalf("expected an error containing %q, got:\n%s", substr, strings.Join(errs, "\n"))
}

func mutateLevel(t *testing.T, fsys fstest.MapFS, path string, f func(*Level)) {
	t.Helper()
	var lv Level
	if err := json.Unmarshal(fsys[path].Data, &lv); err != nil {
		t.Fatal(err)
	}
	f(&lv)
	fsys[path] = &fstest.MapFile{Data: jsonBytes(t, &lv)}
}

func mutateObject(t *testing.T, fsys fstest.MapFS, path string, f func(*Object)) {
	t.Helper()
	var o Object
	if err := json.Unmarshal(fsys[path].Data, &o); err != nil {
		t.Fatal(err)
	}
	f(&o)
	fsys[path] = &fstest.MapFile{Data: jsonBytes(t, &o)}
}

func TestRejectsNewerVersion(t *testing.T) {
	fsys := validGame(t)
	var man Manifest
	_ = json.Unmarshal(fsys["manifest.json"].Data, &man)
	man.Version = Version + 1
	fsys["manifest.json"] = &fstest.MapFile{Data: jsonBytes(t, &man)}
	expectError(t, fsys, "unsupported")
}

func TestBadObjectRef(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv1.json", func(l *Level) {
		l.Placements[0].Object = "nosuch"
	})
	expectError(t, fsys, `object "nosuch" is not an object asset`)
}

func TestBadAnimRef(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv1.json", func(l *Level) {
		l.Placements[0].Anim = "fly"
	})
	expectError(t, fsys, `anim "fly" does not exist`)
}

func TestCellsLengthMismatch(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv1.json", func(l *Level) {
		l.Tilemap.Cells = l.Tilemap.Cells[:5]
	})
	expectError(t, fsys, "cells has 5 entries")
}

func TestCellIDOutOfAtlas(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv1.json", func(l *Level) {
		l.Tilemap.Cells[0] = 99
	})
	expectError(t, fsys, "cell id 99 out of range")
}

func TestDurationsMismatch(t *testing.T) {
	fsys := validGame(t)
	mutateObject(t, fsys, "objects/crab.json", func(o *Object) {
		o.Animations[0].Durations = []int{8}
	})
	expectError(t, fsys, "1 durations for 2 frames")
}

func TestDurationsAndStepsExclusive(t *testing.T) {
	fsys := validGame(t)
	mutateObject(t, fsys, "objects/crab.json", func(o *Object) {
		o.Animations[0].Steps = [][]int{{0, 5}}
	})
	expectError(t, fsys, "mutually exclusive")
}

func TestClipNotInGLB(t *testing.T) {
	fsys := validGame(t)
	mutateObject(t, fsys, "objects/kart.json", func(o *Object) {
		o.Animations[0].Clip = "dance"
	})
	expectError(t, fsys, `clip "dance" is not in the GLB`)
}

func TestUVAnimMaterialNotInGLB(t *testing.T) {
	fsys := validGame(t)
	mutateObject(t, fsys, "objects/kart.json", func(o *Object) {
		o.UVAnims[0].Material = "lava"
	})
	expectError(t, fsys, `uvAnim material "lava" is not in the GLB`)
}

func TestOnClickClipMissing(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv1.json", func(l *Level) {
		l.Placements[0].OnClick.Clip = "open"
	})
	expectError(t, fsys, `clip "open" does not exist`)
}

func TestScriptActorMissingPlacement(t *testing.T) {
	fsys := validGame(t)
	var s Script
	_ = json.Unmarshal(fsys["levels/lv2/intro.json"].Data, &s)
	s.Shots[0].Actors[0].Placement = 99
	fsys["levels/lv2/intro.json"] = &fstest.MapFile{Data: jsonBytes(t, &s)}
	expectError(t, fsys, "actor placement 99 does not exist")
}

func TestCameraTrackLengthMismatch(t *testing.T) {
	fsys := validGame(t)
	var tr CameraTrack
	_ = json.Unmarshal(fsys["levels/lv2/cam.json"].Data, &tr)
	tr.Frames = 3
	fsys["levels/lv2/cam.json"] = &fstest.MapFile{Data: jsonBytes(t, &tr)}
	expectError(t, fsys, "track has 2 samples, want frames = 3")
}

func TestExclusiveGroupNeedsOneVisible(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv2.json", func(l *Level) {
		vis := true
		l.Scene.Layers[3].Visible = &vis // both exclusive layers visible now
	})
	expectError(t, fsys, "exactly one layer of an exclusive group")
}

func TestVariantMembershipChecked(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv2.json", func(l *Level) {
		l.Placements[0].Variants = []string{"dusk"}
	})
	expectError(t, fsys, `variant "dusk" is not declared`)
}

func TestPictureDimensionsChecked(t *testing.T) {
	fsys := validGame(t)
	var man Manifest
	_ = json.Unmarshal(fsys["manifest.json"].Data, &man)
	for i := range man.Assets {
		if man.Assets[i].ID == "title" {
			man.Assets[i].W = 999
		}
	}
	fsys["manifest.json"] = &fstest.MapFile{Data: jsonBytes(t, &man)}
	expectError(t, fsys, "declared 999x16 but PNG is 32x16")
}

func TestEscapingFileRefRejected(t *testing.T) {
	fsys := validGame(t)
	var man Manifest
	_ = json.Unmarshal(fsys["manifest.json"].Data, &man)
	man.Assets[0].File = "../evil.json"
	fsys["manifest.json"] = &fstest.MapFile{Data: jsonBytes(t, &man)}
	expectError(t, fsys, "must not contain '..'")
}

func TestMatrixLength(t *testing.T) {
	fsys := validGame(t)
	mutateLevel(t, fsys, "levels/lv2.json", func(l *Level) {
		l.Placements[1].Matrix = l.Placements[1].Matrix[:15]
	})
	expectError(t, fsys, "matrix must have 16 numbers")
}

func TestScaleRoundTrip(t *testing.T) {
	var p Placement
	if err := json.Unmarshal([]byte(`{"id":1,"object":"o","scale":2.5}`), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Scale) != 1 || p.Scale[0] != 2.5 {
		t.Fatalf("scalar scale mis-parsed: %v", p.Scale)
	}
	b, _ := json.Marshal(p.Scale)
	if string(b) != "2.5" {
		t.Fatalf("scalar scale mis-marshalled: %s", b)
	}
	if err := json.Unmarshal([]byte(`{"id":1,"object":"o","scale":[1,2,3]}`), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Scale) != 3 {
		t.Fatalf("vector scale mis-parsed: %v", p.Scale)
	}
}

func TestValidateTree(t *testing.T) {
	game := validGame(t)
	tree := fstest.MapFS{
		"index.json": {Data: jsonBytes(t, &Index{Header: NewHeader(), Games: []string{"demo"}})},
	}
	for p, f := range game {
		tree["demo/"+p] = f
	}
	if errs := errorsOf(ValidateTree(tree)); len(errs) != 0 {
		t.Fatalf("valid tree reported errors:\n%s", strings.Join(errs, "\n"))
	}
	// A listed game without a manifest is an error.
	tree["index.json"] = &fstest.MapFile{Data: jsonBytes(t, &Index{Header: NewHeader(), Games: []string{"demo", "ghost"}})}
	errs := errorsOf(ValidateTree(tree))
	found := false
	for _, e := range errs {
		if strings.Contains(e, `"ghost" has no manifest.json`) {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing-game error not reported: %v", errs)
	}
}

func TestReadGLBInfo(t *testing.T) {
	info, err := ReadGLBInfo(bytes.NewReader(glbBytes(t,
		`{"animations":[{"name":"walk"},{"name":"run"}],"materials":[{"name":"m0"}]}`)))
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Animations) != 2 || info.Animations[0] != "walk" || info.Animations[1] != "run" {
		t.Fatalf("animations: %v", info.Animations)
	}
	if len(info.Materials) != 1 || info.Materials[0] != "m0" {
		t.Fatalf("materials: %v", info.Materials)
	}
	if _, err := ReadGLBInfo(bytes.NewReader([]byte("not a glb at all"))); err == nil {
		t.Fatal("garbage accepted as GLB")
	}
}
