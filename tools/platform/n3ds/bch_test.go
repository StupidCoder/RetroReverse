package n3ds

// The game-asset formats — SARC archives, BCH models, BYML configuration — are
// exercised against Captain Toad's opening stage, the scene the oracle boots
// into. The cartridge is not committed, so these skip when it is absent.

import (
	"math"
	"os"
	"strings"
	"testing"
)

const (
	toadStage   = "Season1OpeningStage"    // the stage whose map places the terrain
	toadTerrain = "Season1OpeningStepA"    // the object archive holding that terrain
	toadTexArc  = "Season1OpeningTextures" // the archive its InitModel.byml names
	toadScene   = "Season1OpeningScene"    // the scene archive holding its lights
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

// TestBCHSceneLights decodes the opening stage's light rig and checks it
// against the registers the machine actually programs when it draws that stage.
//
// The values here were read back from a running frame (`bootoracle -gputrace`,
// which dumps the lighting unit's state per draw): the terrain's draws report
// two lights, a global ambient of raw 0x06C1D07F and a key diffuse of raw
// 0x0FA2F04C. Those are 10-bit fields holding 8-bit colours, so they decode to
// (108,116,127) and (254,188,76) — the file's (109,117,128) and (255,189,77) to
// within one unit of quantisation. That agreement is the point: it says the
// scene file's lights are the ones the hardware is handed.
func TestBCHSceneLights(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/ObjectData/"+toadScene+".szs")
	blob, ok := a.File(toadScene + ".bch")
	if !ok {
		t.Fatal("the scene archive has no .bch")
	}
	f, err := ParseBCH(blob)
	if err != nil {
		t.Fatal(err)
	}
	lights, err := f.Lights()
	if err != nil {
		t.Fatal(err)
	}
	if len(lights) != 4 {
		t.Fatalf("%d lights, want 4", len(lights))
	}

	key := lights["mainLight"]
	if key == nil {
		t.Fatal("no mainLight")
	}
	if !key.Directional {
		t.Error("mainLight is not directional")
	}
	if key.Diffuse != [3]uint8{255, 189, 77} {
		t.Errorf("mainLight diffuse = %v, want the warm key {255 189 77}", key.Diffuse)
	}
	// The stored vector is the direction the light travels: it points *down*,
	// and the sun is above. Reading it as the direction towards the light would
	// light the underside of everything.
	if key.Direction[1] >= 0 {
		t.Errorf("mainLight direction %v does not travel downwards", key.Direction)
	}

	amb := lights["AmbientLight"]
	if amb == nil {
		t.Fatal("no AmbientLight")
	}
	if amb.Directional {
		t.Error("AmbientLight should carry no direction")
	}
	if amb.Ambient != [3]uint8{109, 117, 128} {
		t.Errorf("AmbientLight = %v, want the cool sky {109 117 128}", amb.Ambient)
	}

	// Both directional lights are unit vectors — DecodeLight enforces it, and
	// it is what makes the +0x40 offset trustworthy.
	for name, l := range lights {
		if !l.Directional {
			continue
		}
		var sum float64
		for _, c := range l.Direction {
			sum += float64(c) * float64(c)
		}
		if d := sum - 1; d > 1e-5 || d < -1e-5 {
			t.Errorf("light %q direction is not unit length (|d|^2 = %v)", name, sum)
		}
	}
}

// TestBCHStageLightAreas checks the other half of the chain: the stage's own
// Design archive says which of the scene's lights light it.
func TestBCHStageLightAreas(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/StageData/"+toadStage+"Design1.szs")
	blob, ok := a.File("LightAreas.byml")
	if !ok {
		t.Fatal("the design archive has no LightAreas.byml")
	}
	doc, err := ParseBYML(blob)
	if err != nil {
		t.Fatal(err)
	}
	areas, _ := doc.(BYMLDict)["LightAreas"].([]any)
	if len(areas) != 2 {
		t.Fatalf("%d light areas, want 2", len(areas))
	}
	// Both areas light the terrain with a key plus the ambient; they differ
	// only in whether that key casts a depth shadow, which is why either can be
	// read for the rig's colours.
	want := [][]string{{"mainLight", "AmbientLight"}, {"mainLight_noshadow", "AmbientLight"}}
	for i, area := range areas {
		d, _ := area.(BYMLDict)
		names, _ := d["Lights"].([]any)
		var got []string
		for _, n := range names {
			nd, _ := n.(BYMLDict)
			s, _ := nd["lightName"].(string)
			got = append(got, s)
		}
		if len(got) != len(want[i]) {
			t.Errorf("area %d has %v, want %v", i, got, want[i])
			continue
		}
		for j := range got {
			if got[j] != want[i][j] {
				t.Errorf("area %d light %d = %q, want %q", i, j, got[j], want[i][j])
			}
		}
	}
}

// TestBYMLDepthShadow reads the stage's depth-shadow settings, which its Design
// archive ships beside the light areas.
//
// BiasFactor is the number that matters and it is not a slope-acne nudge: a
// fraction of the near-far range, 0.0185 of 4,300 units is about eighty world
// units. That is the *stand-off of the caster*. The depth-shadow proxy is a
// coarse shell stretched over the light-facing side of the object — 128 of its
// 136 triangles face the light, and it is not a closed solid — standing well
// clear of the real surface, so every recess it bridges over needs that much
// bias to stay lit. Comparing against it with a small bias shadows geometry the
// game does not, and moirés every face lying near the shell.
func TestBYMLDepthShadow(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/StageData/"+toadStage+"Design1.szs")
	blob, ok := a.File("DepthShadow.byml")
	if !ok {
		t.Fatal("the design archive has no DepthShadow.byml")
	}
	doc, err := ParseBYML(blob)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := doc.(BYMLDict)
	p, _ := d.Get("DepthShadowParam", "DefaultDepthShadowParam").(BYMLDict)
	if p == nil {
		t.Fatal("no DefaultDepthShadowParam")
	}
	if on, _ := p["IsDepthShadowEnable"].(bool); !on {
		t.Error("the stage's default depth shadow is disabled")
	}
	for _, want := range []struct {
		key string
		v   float32
	}{{"BiasFactor", 0.0185}, {"Near", 500}, {"Far", 4800}, {"ColorA", 150}} {
		got, ok := p[want.key].(float32)
		if !ok {
			if i, isInt := p[want.key].(int32); isInt {
				got, ok = float32(i), true
			}
		}
		if !ok || got != want.v {
			t.Errorf("%s = %v, want %v", want.key, p[want.key], want.v)
		}
	}
	// The bias in world units is what says it is a stand-off, not an epsilon.
	near, _ := p["Near"].(float32)
	far, _ := p["Far"].(float32)
	bias, _ := p["BiasFactor"].(float32)
	if u := bias * (far - near); u < 50 || u > 120 {
		t.Errorf("bias is %.1f world units; the caster stand-off should put it near eighty", u)
	}
	g, _ := d["GlobalParam"].(BYMLDict)
	if w, _ := g["ShadowMapWidth"].(int32); w != 512 {
		t.Errorf("ShadowMapWidth = %v, want 512", g["ShadowMapWidth"])
	}
}

// TestBCHSkeletonBindPose proves the skeleton decode with the invariant a
// skeleton carries about itself: composing a bone's world matrix down the
// parent chain and multiplying by the inverse bind matrix stored beside it must
// give the identity, because that is what "inverse bind" means.
//
// It settles four things at once — the bone stride, the parent links, the
// order the Euler triple composes in, and the inverse-bind layout — and it is
// sharp. The rotation is Rz·Ry·Rx, the X term applied first: over Toadette's
// skeleton, whose bones turn about all three axes, that order lands within 0.01
// while the next best is out by 30 and the one after by 67. A skeleton with no
// three-axis bone cannot tell those apart, which is why the check runs over the
// one that has them.
func TestBCHSkeletonBindPose(t *testing.T) {
	fs := toadRomFS(t)
	for _, obj := range []string{"Kinopio", "KinopicoNpc"} {
		a := openSZS(t, fs, "/ObjectData/"+obj+".szs")
		blob, ok := a.File(obj + ".bch")
		if !ok {
			t.Fatalf("%s: no .bch", obj)
		}
		f, err := ParseBCH(blob)
		if err != nil {
			t.Fatal(err)
		}
		m, err := f.DecodeModel(f.Groups[BCHModels][0])
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Bones) == 0 {
			t.Fatalf("%s has no skeleton", obj)
		}
		world := make([][16]float64, len(m.Bones))
		threeAxis := 0
		for i, b := range m.Bones {
			if b.Rotate[0] != 0 && b.Rotate[1] != 0 && b.Rotate[2] != 0 {
				threeAxis++
			}
			l := mul4(mul4(trans4(b.Trans), eulerZYX(b.Rotate)), scale4(b.Scale))
			if b.Parent < 0 {
				world[i] = l
			} else {
				if b.Parent >= i {
					t.Fatalf("%s bone %q has parent %d, which is not before it", obj, b.Name, b.Parent)
				}
				world[i] = mul4(world[b.Parent], l)
			}
		}
		worst := 0.0
		for i, b := range m.Bones {
			var ib [16]float64
			for r := 0; r < 3; r++ {
				for c := 0; c < 4; c++ {
					ib[r*4+c] = float64(b.InvBind[r][c])
				}
			}
			ib[15] = 1
			p := mul4(world[i], ib)
			for r := 0; r < 4; r++ {
				for c := 0; c < 4; c++ {
					want := 0.0
					if r == c {
						want = 1
					}
					if d := math.Abs(p[r*4+c] - want); d > worst {
						worst = d
					}
				}
			}
		}
		if worst > 0.02 {
			t.Errorf("%s: worst |world x invBind - I| = %g over %d bones; the bind pose does not close",
				obj, worst, len(m.Bones))
		}
		t.Logf("%s: %d bones (%d turning about all three axes), worst bind-pose residual %.4f",
			obj, len(m.Bones), threeAxis, worst)
	}
}

// TestBCHSkinWeights checks the other half of the skin: the stored weights are
// percentages, and every skinned vertex's influences sum to one. DecodeModel
// enforces it, so this pins that the enforcement is exercised and that the
// palette a vertex indexes is in range.
func TestBCHSkinWeights(t *testing.T) {
	fs := toadRomFS(t)
	a := openSZS(t, fs, "/ObjectData/Kinopio.szs")
	blob, _ := a.File("Kinopio.bch")
	f, err := ParseBCH(blob)
	if err != nil {
		t.Fatal(err)
	}
	m, err := f.DecodeModel(f.Groups[BCHModels][0])
	if err != nil {
		t.Fatal(err)
	}
	skinned, influences := 0, 0
	for i, sh := range m.Meshes {
		if !sh.HasSkin {
			continue
		}
		skinned++
		for _, v := range sh.Verts {
			var sum float32
			for k, w := range v.Weights {
				if w == 0 {
					continue
				}
				influences++
				sum += w
				if int(v.Joints[k]) >= len(sh.Palette) {
					t.Fatalf("mesh %d: joint %d indexes a palette of %d", i, v.Joints[k], len(sh.Palette))
				}
				if b := sh.Palette[v.Joints[k]]; b >= len(m.Bones) {
					t.Fatalf("mesh %d: palette slot %d names bone %d of %d", i, v.Joints[k], b, len(m.Bones))
				}
			}
			if sum < 0.99 || sum > 1.01 {
				t.Fatalf("mesh %d: weights sum to %g", i, sum)
			}
		}
	}
	if skinned == 0 || influences == 0 {
		t.Fatalf("%d skinned meshes, %d influences — the skin decode is not being exercised", skinned, influences)
	}
	t.Logf("%d skinned meshes, %d vertex influences, all weights sum to one", skinned, influences)
}

func mul4(a, b [16]float64) [16]float64 {
	var o [16]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			s := 0.0
			for k := 0; k < 4; k++ {
				s += a[r*4+k] * b[k*4+c]
			}
			o[r*4+c] = s
		}
	}
	return o
}
func trans4(t [3]float32) [16]float64 {
	return [16]float64{1, 0, 0, float64(t[0]), 0, 1, 0, float64(t[1]), 0, 0, 1, float64(t[2]), 0, 0, 0, 1}
}
func scale4(s [3]float32) [16]float64 {
	return [16]float64{float64(s[0]), 0, 0, 0, 0, float64(s[1]), 0, 0, 0, 0, float64(s[2]), 0, 0, 0, 0, 1}
}

// eulerZYX builds Rz·Ry·Rx from an XYZ triple in radians.
func eulerZYX(r [3]float32) [16]float64 {
	rot := func(ax int, a float64) [16]float64 {
		c, s := math.Cos(a), math.Sin(a)
		switch ax {
		case 0:
			return [16]float64{1, 0, 0, 0, 0, c, -s, 0, 0, s, c, 0, 0, 0, 0, 1}
		case 1:
			return [16]float64{c, 0, s, 0, 0, 1, 0, 0, -s, 0, c, 0, 0, 0, 0, 1}
		}
		return [16]float64{c, -s, 0, 0, s, c, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	}
	return mul4(mul4(rot(2, float64(r[2])), rot(1, float64(r[1]))), rot(0, float64(r[0])))
}
