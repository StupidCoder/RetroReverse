package n3ds

// The game-asset formats — SARC archives, BCH models, BYML configuration — are
// exercised against Captain Toad's opening stage, the scene the oracle boots
// into. The cartridge is not committed, so these skip when it is absent.

import (
	"os"
	"strings"
	"testing"
)

const (
	toadStage   = "Season1OpeningStage"    // the stage whose map places the terrain
	toadTerrain = "Season1OpeningStepA"    // the object archive holding that terrain
	toadTexArc  = "Season1OpeningTextures" // the archive its InitModel.byml names
)

func toadRomFS(t *testing.T) *RomFS {
	t.Helper()
	img, err := os.ReadFile(toadImagePath)
	if err != nil {
		t.Skip("Captain Toad image not present (game images are not committed)")
	}
	n, err := ParseNCSD(img)
	if err != nil {
		t.Fatal(err)
	}
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fs, err := c.RomFS()
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func openSZS(t *testing.T, fs *RomFS, path string) *SARC {
	t.Helper()
	raw, err := fs.File(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	a, err := OpenSZS(raw)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return a
}

// TestSARCNameHashes leans on the archive's own redundancy: every entry stores
// both its name and a hash of that name, and ParseSARC refuses any archive
// where the two disagree. Opening a handful proves the node stride, the string
// table offsets and the hash key at once — a misread layout cannot produce
// names that hash correctly.
func TestSARCNameHashes(t *testing.T) {
	fs := toadRomFS(t)
	for _, p := range []string{
		"/StageData/" + toadStage + "Map1.szs",
		"/ObjectData/" + toadTerrain + ".szs",
		"/ObjectData/" + toadTexArc + ".szs",
	} {
		a := openSZS(t, fs, p)
		if len(a.Files) == 0 {
			t.Errorf("%s: no files", p)
		}
		for _, f := range a.Files {
			if f.Name == "" || len(f.Data) == 0 {
				t.Errorf("%s: entry %q is empty", p, f.Name)
			}
		}
	}
}

// TestBCHVertexArraysTile is the decode's own proof. A model's meshes take their
// vertices from one shared array in the extended data section, each naming its
// own byte offset into it, and the arrays sit end to end: mesh k's offset plus
// its vertex count times its stride is exactly where some other mesh begins.
//
// The vertex count is not stored anywhere — it comes from the largest index in
// the mesh's index array, which is only right if the index *width* is right,
// which comes from the relocation type of the pointer that patches it. So a
// wrong stride, a wrong array offset, a wrong index width or a wrong index
// count each break the chain. All 53 arrays tiling with no gap and no overlap
// is a coincidence nothing but a correct decode produces.
func TestBCHVertexArraysTile(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/ObjectData/"+toadTerrain+".szs")
	blob, ok := a.File(toadTerrain + ".bch")
	if !ok {
		t.Fatal("the object archive has no .bch")
	}
	f, err := ParseBCH(blob)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(f.Groups[BCHModels]); n != 2 {
		t.Fatalf("%d models, want 2 (the object and its depth-shadow proxy)", n)
	}
	m, err := f.DecodeModel(f.Groups[BCHModels][0])
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Meshes) != 53 || len(m.Materials) != 53 {
		t.Fatalf("%d meshes / %d materials, want 53/53", len(m.Meshes), len(m.Materials))
	}

	starts := map[int64]bool{}
	for _, sh := range m.Meshes {
		starts[sh.ArrayOffset] = true
	}
	for i, sh := range m.Meshes {
		if sh.VertexStride == 0 {
			t.Fatalf("mesh %d has no stride", i)
		}
		end := sh.ArrayOffset + int64(len(sh.Verts))*sh.VertexStride
		// The last array in storage order has nothing after it; every other
		// must land exactly on some mesh's start.
		last := true
		for s := range starts {
			if s >= end {
				last = false
				break
			}
		}
		if !last && !starts[end] {
			t.Errorf("mesh %d: %d vertices at stride %d end at 0x%x, which starts no other mesh",
				i, len(sh.Verts), sh.VertexStride, end)
		}
		for _, ix := range sh.Indices {
			if int(ix) >= len(sh.Verts) {
				t.Fatalf("mesh %d: index %d exceeds %d vertices", i, ix, len(sh.Verts))
			}
		}
		if len(sh.Indices)%3 != 0 {
			t.Errorf("mesh %d has %d indices, not a triangle list", i, len(sh.Indices))
		}
	}
}

// TestBCHMaterialTextures checks the model's per-material binding table against
// the texture archive its own InitModel.byml names — the two files are only
// consistent if both the binding table's layout and the archive's texture group
// are read correctly.
func TestBCHMaterialTextures(t *testing.T) {
	fs := toadRomFS(t)
	obj := openSZS(t, fs, "/ObjectData/"+toadTerrain+".szs")

	// The object names its texture archive; nothing is inferred from filenames.
	init, ok := obj.File("InitModel.byml")
	if !ok {
		t.Fatal("the object archive has no InitModel.byml")
	}
	doc, err := ParseBYML(init)
	if err != nil {
		t.Fatal(err)
	}
	arcName, _ := doc.(BYMLDict)["TextureArc"].(string)
	if arcName != toadTexArc {
		t.Fatalf("TextureArc = %q, want %s", arcName, toadTexArc)
	}

	texArc := openSZS(t, fs, "/ObjectData/"+arcName+".szs")
	texBlob, _ := texArc.File(arcName + ".bch")
	tf, err := ParseBCH(texBlob)
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, e := range tf.Groups[BCHTextures] {
		tex, err := tf.DecodeTexture(e)
		if err != nil {
			t.Fatalf("texture %q: %v", e.Name, err)
		}
		if tex.Format != 0xC && tex.Format != 0xD {
			t.Errorf("texture %q has format 0x%X; this stage's are all ETC1 or ETC1A4", tex.Name, tex.Format)
		}
		if tex.Image.Rect.Dx() != tex.Width || tex.Image.Rect.Dy() != tex.Height {
			t.Errorf("texture %q decoded to %dx%d, the registers say %dx%d",
				tex.Name, tex.Image.Rect.Dx(), tex.Image.Rect.Dy(), tex.Width, tex.Height)
		}
		have[tex.Name] = true
	}
	if len(have) != 30 {
		t.Errorf("%d textures in %s, want 30", len(have), arcName)
	}

	blob, _ := obj.File(toadTerrain + ".bch")
	f, _ := ParseBCH(blob)
	m, err := f.DecodeModel(f.Groups[BCHModels][0])
	if err != nil {
		t.Fatal(err)
	}
	// Every material's unit-0 texture must exist in the archive the object
	// named. A binding table read at the wrong offset yields names that are not
	// textures at all.
	textured := 0
	for _, mat := range m.Materials {
		tex := mat.Texture()
		if tex == "" {
			continue // a vertex-coloured material samples nothing
		}
		textured++
		if !have[tex] {
			t.Errorf("material %q binds texture %q, which %s does not hold", mat.Name, tex, arcName)
		}
	}
	if textured != 51 {
		t.Errorf("%d materials bind a texture, want 51 (two of the 53 are vertex-coloured grass)", textured)
	}
	// And the pairing is meaningful, not just present.
	for _, want := range []struct{ mat, tex string }{
		{"ivybillboard", "ivybillboard_alb"},
		{"moss", "moss_alb"},
		{"Clovers", "clowers_alb"},
	} {
		found := false
		for _, mat := range m.Materials {
			if mat.Name == want.mat && mat.Texture() == want.tex {
				found = true
			}
		}
		if !found {
			t.Errorf("no material %q binding %q", want.mat, want.tex)
		}
	}
}

// TestBYMLStageMap decodes the opening stage's placement map. The values are
// the ones the level export depends on, and a BYML read with the wrong node
// stride or the wrong key table produces neither the key names nor the numbers.
func TestBYMLStageMap(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/StageData/"+toadStage+"Map1.szs")
	blob, ok := a.File(toadStage + "Map.byml")
	if !ok {
		t.Fatal("the map archive has no map BYML")
	}
	doc, err := ParseBYML(blob)
	if err != nil {
		t.Fatal(err)
	}
	root, ok := doc.(BYMLDict)
	if !ok {
		t.Fatalf("root is %T, want a dictionary", doc)
	}
	objs, _ := root["ObjectList"].([]any)
	if len(objs) == 0 {
		t.Fatal("the map places no objects")
	}
	var found BYMLDict
	for _, o := range objs {
		d, _ := o.(BYMLDict)
		if n, _ := d["UnitConfigName"].(string); n == toadTerrain {
			found = d
		}
	}
	if found == nil {
		t.Fatalf("the map does not place %s", toadTerrain)
	}
	for _, want := range []struct {
		axis string
		v    float32
	}{{"X", 200}, {"Y", -750}, {"Z", 50}} {
		got := found.Get("Translate", want.axis)
		if f, ok := got.(float32); !ok || f != want.v {
			t.Errorf("StepA Translate.%s = %v, want %v", want.axis, got, want.v)
		}
	}
	if s, _ := root["FilePath"].(string); !strings.HasSuffix(s, ".muunt") {
		t.Errorf("FilePath = %q, want the map's authoring path", s)
	}
}

// TestBCHMaterialAlpha pins how the stage's materials treat alpha, which is not
// something the textures can be asked.
//
// Every one of this stage's albedo textures is ETC1A4 and therefore carries a
// 4-bit alpha plane — but only ten of the fifty-three materials sampling them
// enable the alpha test, and none blends at all. For the other forty-three the
// plane is not coverage, and honouring it masks holes through the stone and
// deletes the soil layers under the diorama. The tell is in the textures too:
// all five stone albedos carry the *identical* alpha histogram, which is not
// what per-texture coverage looks like.
func TestBCHMaterialAlpha(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/ObjectData/"+toadTerrain+".szs")
	blob, _ := a.File(toadTerrain + ".bch")
	f, err := ParseBCH(blob)
	if err != nil {
		t.Fatal(err)
	}
	m, err := f.DecodeModel(f.Groups[BCHModels][0])
	if err != nil {
		t.Fatal(err)
	}

	blend, test, opaque := 0, 0, 0
	byName := map[string]*BCHMaterial{}
	for i := range m.Materials {
		mat := &m.Materials[i]
		byName[mat.Name] = mat
		switch {
		case mat.Blends:
			blend++
		case mat.AlphaTest:
			test++
			c, ok := mat.AlphaCutoff()
			if !ok {
				t.Errorf("material %q alpha-tests with comparison %d, which is not a cutoff", mat.Name, mat.AlphaFunc)
			}
			if c <= 0 || c > 1 {
				t.Errorf("material %q cutoff %v is outside [0,1]", mat.Name, c)
			}
		default:
			opaque++
		}
	}
	if blend != 0 || test != 10 || opaque != 43 {
		t.Errorf("%d blend / %d alpha-test / %d opaque, want 0/10/43", blend, test, opaque)
	}

	// The split is meaningful, not just numerically right: the cut-out decals
	// test, and the solid stone does not.
	for _, name := range []string{"Clovers", "ivybillboard", "moss", "OpStone03_AI_Tile"} {
		mat := byName[name]
		if mat == nil {
			t.Errorf("no material %q", name)
			continue
		}
		if !mat.AlphaTest {
			t.Errorf("material %q is a cut-out decal but does not alpha-test", name)
		}
	}
	for _, name := range []string{"OpStone00_AI", "OpStone03_AI", "Lawn01_S", "DanmenSoilLayer00_S"} {
		mat := byName[name]
		if mat == nil {
			t.Errorf("no material %q", name)
			continue
		}
		if mat.AlphaTest || mat.Blends {
			t.Errorf("material %q is solid but reports alphaTest=%v blends=%v", name, mat.AlphaTest, mat.Blends)
		}
	}
}
