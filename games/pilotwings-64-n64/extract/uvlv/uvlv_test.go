package uvlv

import (
	"os"
	"sort"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func archive(t *testing.T) *pwad.Archive {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// commChunks decodes every COMM chunk of the archive's single resource of a type.
func commChunks(t *testing.T, a *pwad.Archive, fourCC string) [][]byte {
	t.Helper()
	idx := a.ByType(fourCC)
	if len(idx) != 1 {
		t.Fatalf("%d %s resources, want 1", len(idx), fourCC)
	}
	f, err := a.Resource(idx[0])
	if err != nil {
		t.Fatal(err)
	}
	var out [][]byte
	for _, c := range f.Chunks {
		if c.Tag != "COMM" && !(c.Compressed() && c.InnerTag == "COMM") {
			continue
		}
		data, err := a.Data(c)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, data)
	}
	return out
}

func scenes(t *testing.T, a *pwad.Archive) []*Scene {
	t.Helper()
	var out []*Scene
	for i, data := range commChunks(t, a, "UVLV") {
		s, err := Decode(data)
		if err != nil {
			t.Fatalf("scene %d: %v", i, err)
		}
		out = append(out, s)
	}
	if len(out) != 136 {
		t.Fatalf("%d scenes, want 136", len(out))
	}
	return out
}

// Every scene parses to exactly ten arrays and lands within six bytes of the
// chunk's end, on zeroes. A slot boundary in the wrong place would run an array
// off the end or leave a non-zero remainder. Nine slots also parse cleanly —
// which is why the count came from the parser's ten unrolled reads, not from
// this test.
func TestEverySceneParsesToTenArrays(t *testing.T) {
	a := archive(t)
	used := [Slots]int{}
	for i, s := range scenes(t, a) {
		if s.Padding > 6 {
			t.Errorf("scene %d: %d bytes of padding, want at most 6", i, s.Padding)
		}
		if len(s.Slot[1]) != 0 {
			t.Errorf("scene %d: slot 1 is not empty (%d entries)", i, len(s.Slot[1]))
		}
		for k := range s.Slot {
			if len(s.Slot[k]) > 0 {
				used[k]++
			}
		}
	}
	t.Logf("scenes using each slot: %v", used)
}

// Slot 8 is the nine fonts, resident in every scene. It is what makes the slots
// legible as ordinals-within-a-type in the first place: nine values, 0..8, and
// the archive holds exactly nine UVFT resources.
func TestFontsAreResidentInEveryScene(t *testing.T) {
	a := archive(t)
	if n := len(a.ByType("UVFT")); n != 9 {
		t.Fatalf("%d UVFT resources, want 9", n)
	}
	for i, s := range scenes(t, a) {
		f := s.Fonts()
		if len(f) != 9 {
			t.Fatalf("scene %d: %d fonts, want 9", i, len(f))
		}
		for j, v := range f {
			if int(v) != j {
				t.Fatalf("scene %d: font[%d] = %d, want %d", i, j, v, j)
			}
		}
	}
}

// The load-bearing check on slots 0 and 4: a scene names worlds, and the terrain
// chunks it loads are exactly the ones those worlds' grid cells name. UVTR and
// UVLV were decoded independently and neither mentions the other; that they
// agree, thirteen times, is what identifies both slots.
func TestTerrainListsAgreeWithTheWorldGrids(t *testing.T) {
	a := archive(t)
	var worlds []*uvtr.World
	for i, data := range commChunks(t, a, "UVTR") {
		w, err := uvtr.Decode(data)
		if err != nil {
			t.Fatalf("world %d: %v", i, err)
		}
		worlds = append(worlds, w)
	}
	if len(worlds) != 10 {
		t.Fatalf("%d worlds, want 10", len(worlds))
	}

	withTerrain := 0
	for i, s := range scenes(t, a) {
		if len(s.Terrain()) == 0 {
			if len(s.Worlds()) != 0 {
				t.Errorf("scene %d names worlds but no terrain", i)
			}
			continue
		}
		withTerrain++
		want := map[int]bool{}
		for _, w := range s.Worlds() {
			if int(w) >= len(worlds) {
				t.Fatalf("scene %d: world %d out of range", i, w)
			}
			for _, c := range worlds[w].Cells {
				if c.Present {
					want[c.Chunk] = true
				}
			}
		}
		got := map[int]bool{}
		for _, c := range s.Terrain() {
			got[int(c)] = true
		}
		if len(want) != len(got) {
			t.Errorf("scene %d: worlds %v name %d chunks, slot 4 lists %d", i, s.Worlds(), len(want), len(got))
			continue
		}
		for c := range want {
			if !got[c] {
				t.Errorf("scene %d: worlds %v name UVCT %d, absent from slot 4", i, s.Worlds(), c)
			}
		}
	}
	if withTerrain != 13 {
		t.Errorf("%d scenes name terrain, want 13", withTerrain)
	}
}

// An object's `type` is a UVMD ordinal — read out of the engine at 0x8022C59C,
// where it indexes a table of loaded models raw. If that is right, then every
// object type in a scene's terrain must appear in the scene's own model list,
// because a type whose model the scene never loaded would index a zero and the
// engine would fault instead of drawing. Thirteen scenes, 183 distinct types,
// checked against the ROM with no machine running.
func TestEveryObjectTypeIsAModelTheSceneLoads(t *testing.T) {
	a := archive(t)
	uvctIdx := a.ByType("UVCT")
	chunks := make([]*uvct.Chunk, len(uvctIdx))
	for i, idx := range uvctIdx {
		f, err := a.Resource(idx)
		if err != nil {
			t.Fatal(err)
		}
		var data []byte
		for _, c := range f.Chunks {
			// A UVCT's COMM is compressed, so it arrives wrapped in a GZIP chunk.
			if c.Tag == "COMM" || (c.Compressed() && c.InnerTag == "COMM") {
				if data, err = a.Data(c); err != nil {
					t.Fatal(err)
				}
			}
		}
		if chunks[i], err = uvct.Decode(data); err != nil {
			t.Fatalf("UVCT %d: %v", i, err)
		}
	}
	nUVMD := len(a.ByType("UVMD"))

	allTypes := map[uint16]bool{}
	checked := 0
	for i, s := range scenes(t, a) {
		if len(s.Terrain()) == 0 {
			continue
		}
		checked++
		models := map[uint16]bool{}
		for _, m := range s.Models() {
			if int(m) >= nUVMD {
				t.Fatalf("scene %d: UVMD ordinal %d out of range (%d resources)", i, m, nUVMD)
			}
			models[m] = true
		}
		for _, c := range s.Terrain() {
			if int(c) >= len(chunks) {
				t.Fatalf("scene %d: UVCT ordinal %d out of range", i, c)
			}
			for _, o := range chunks[c].Objects {
				allTypes[o.Type] = true
				if !models[o.Type] {
					t.Errorf("scene %d: UVCT %d places type %d, not in the scene's %d models",
						i, c, o.Type, len(models))
				}
			}
		}
	}
	if checked != 13 {
		t.Errorf("checked %d scenes, want 13", checked)
	}
	if len(allTypes) != 183 {
		t.Errorf("%d distinct object types, want 183", len(allTypes))
	}
	types := make([]int, 0, len(allTypes))
	for tp := range allTypes {
		types = append(types, int(tp))
	}
	sort.Ints(types)
	t.Logf("%d scenes, %d distinct object types (%d..%d), every one a model its scene loads",
		checked, len(types), types[0], types[len(types)-1])
}
