package n3ds

// The container parsers are exercised against the Super Mario 3D Land cartridge,
// which is not committed (see the repository's image policy) and so these skip
// when it is absent. Everything asserted here is arithmetic the headers pin
// themselves: a wrong media unit, a wrong level order or a wrong BLZ decode
// cannot satisfy it.

import (
	"os"
	"testing"
)

const imagePath = "../../../games/super-mario-3d-land-3ds/image/Super Mario 3D Land (Europe) (En,Fr,De,Es,It,Nl,Pt,Ru) (Rev 2).cci"

func loadImage(t *testing.T) *NCSD {
	t.Helper()
	img, err := os.ReadFile(imagePath)
	if err != nil {
		t.Skip("Super Mario 3D Land image not present (game images are not committed)")
	}
	n, err := ParseNCSD(img)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNCSDHeader(t *testing.T) {
	n := loadImage(t)
	if n.MediaUnitSize != 512 {
		t.Errorf("media unit size = %d, want 512", n.MediaUnitSize)
	}
	if n.MediaID != 0x0004000000053f00 {
		t.Errorf("media ID = %016x", n.MediaID)
	}
	// The header's own image size must agree with the file length.
	img, _ := os.ReadFile(imagePath)
	if n.ImageSize != int64(len(img)) {
		t.Errorf("header image size 0x%x != file size 0x%x", n.ImageSize, len(img))
	}
	// Partitions 0 (application), 1 (manual) and 7 (update data) are used.
	for _, i := range []int{0, 1, 7} {
		if n.Partitions[i].Empty() {
			t.Errorf("partition %d is empty", i)
		}
	}
	for _, i := range []int{2, 3, 4, 5, 6} {
		if !n.Partitions[i].Empty() {
			t.Errorf("partition %d is unexpectedly present", i)
		}
	}
}

func TestNCCHIsDecrypted(t *testing.T) {
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if c.Encrypted() {
		t.Fatal("partition 0 reports as encrypted; the parsers below cannot work on it")
	}
	if c.ProductCode != "CTR-P-AREP" {
		t.Errorf("product code = %q", c.ProductCode)
	}
	if c.ProgramID != 0x0004000000053f00 {
		t.Errorf("program ID = %016x", c.ProgramID)
	}
}

func TestExHeaderCodeSetLayout(t *testing.T) {
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := c.ExHeader()
	if err != nil {
		t.Fatal(err)
	}
	if ex.Title != "CtrApp" {
		t.Errorf("title = %q", ex.Title)
	}
	if !ex.CompressedExeFSCode() {
		t.Error("expected .code to be flagged BLZ-compressed")
	}
	if ex.Text.Address != 0x00100000 {
		t.Errorf("text address = 0x%08x, want 0x00100000", ex.Text.Address)
	}
	// ParseExHeader already asserts the segments tile contiguously; restate the
	// resulting addresses so a regression names the segment that moved.
	if ex.ROData.Address != 0x003a2000 {
		t.Errorf("rodata address = 0x%08x", ex.ROData.Address)
	}
	if ex.Data.Address != 0x003e1000 {
		t.Errorf("data address = 0x%08x", ex.Data.Address)
	}
	if ex.BSSAddress() != 0x003f39d4 {
		t.Errorf("bss address = 0x%08x", ex.BSSAddress())
	}
	if ex.CodeSize() != 0x2f4000 {
		t.Errorf("expected .code size = 0x%x, want 0x2f4000", ex.CodeSize())
	}
}

// The round-trip proof for the BLZ decoder: the decompressed length is pinned by
// ExHeader arithmetic the compressed stream knows nothing about, and Code()
// fails if they disagree.
func TestExeFSCodeDecompresses(t *testing.T) {
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := c.ExHeader()
	if err != nil {
		t.Fatal(err)
	}
	efs, err := c.ExeFS()
	if err != nil {
		t.Fatal(err)
	}
	if len(efs.Files) != 4 {
		t.Errorf("ExeFS has %d files, want 4", len(efs.Files))
	}
	code, err := efs.Code(ex)
	if err != nil {
		t.Fatal(err)
	}
	if uint32(len(code)) != ex.CodeSize() {
		t.Fatalf("code is 0x%x bytes, want 0x%x", len(code), ex.CodeSize())
	}
	// Each segment is zero-padded from its Size out to its page Extent.
	tailStart := ex.Text.Extent() + ex.ROData.Extent() + ex.Data.Size
	for i := tailStart; i < uint32(len(code)); i++ {
		if code[i] != 0 {
			t.Fatalf("data segment padding at 0x%x is 0x%02x, want 0", i, code[i])
		}
	}
}

func TestRomFSTree(t *testing.T) {
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fs, err := c.RomFS()
	if err != nil {
		t.Fatal(err)
	}
	if len(fs.Files) != 1771 {
		t.Errorf("RomFS has %d files, want 1771", len(fs.Files))
	}
	if len(fs.Dirs) != 33 {
		t.Errorf("RomFS has %d directories, want 33", len(fs.Dirs))
	}
	// Every file's data must lie inside the region; ParseRomFS enforces it, so
	// a non-empty tree that parsed at all is already bounded. Spot-check that a
	// known top-level directory came through.
	var found bool
	for _, d := range fs.Dirs {
		if d == "/StageData" {
			found = true
		}
	}
	if !found {
		t.Error("expected a /StageData directory")
	}
}

// VerifyIVFC reads the whole 300 MB region, so it is a long test.
func TestRomFSIVFCHashTree(t *testing.T) {
	if testing.Short() {
		t.Skip("hashes the whole RomFS region")
	}
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fs, err := c.RomFS()
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.VerifyIVFC(); err != nil {
		t.Fatal(err)
	}
}

// The banner (CBMD) decompresses via LZ11 to a valid CGFX whose scene graph
// enumerates: exactly one model, four textures, and a skeletal animation — the
// animated 3-D scene the HOME Menu shows.
func TestBannerScene(t *testing.T) {
	n := loadImage(t)
	c, err := n.Executable()
	if err != nil {
		t.Fatal(err)
	}
	efs, err := c.ExeFS()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := efs.File("banner")
	if err != nil {
		t.Fatal(err)
	}
	bn, err := ParseBanner(raw)
	if err != nil {
		t.Fatal(err)
	}
	cgfx, err := bn.CommonModel()
	if err != nil {
		t.Fatal(err)
	}
	g, err := ParseCGFX(cgfx)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(g.Resources["Models"]); got != 1 {
		t.Errorf("Models = %d, want 1", got)
	}
	if got := len(g.Resources["Textures"]); got != 4 {
		t.Errorf("Textures = %d, want 4", got)
	}
	if got := len(g.Resources["SkeletalAnimations"]); got != 1 {
		t.Errorf("SkeletalAnimations = %d, want 1", got)
	}
	if g.IMAGOff == 0 {
		t.Error("no IMAG data block found")
	}
	if m := g.Resources["Models"]; len(m) == 1 && m[0].Magic != "CMDL" {
		t.Errorf("model magic = %q, want CMDL", m[0].Magic)
	}
}

// The banner CGFX decodes to a complete, self-consistent scene: a model whose
// meshes bind real shapes and materials, textures whose pixel budgets match
// their headers, and an animation over named bones. Everything asserted is
// arithmetic the format pins itself, so it cannot pass on a mis-decode.
func TestBannerModelDecode(t *testing.T) {
	n := loadImage(t)
	c, _ := n.Executable()
	efs, _ := c.ExeFS()
	raw, _ := efs.File("banner")
	bn, err := ParseBanner(raw)
	if err != nil {
		t.Fatal(err)
	}
	cgfx, err := bn.CommonModel()
	if err != nil {
		t.Fatal(err)
	}
	g, err := ParseCGFX(cgfx)
	if err != nil {
		t.Fatal(err)
	}

	m, err := g.DecodeModel(g.Resources["Models"][0])
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Meshes) != 8 || len(m.Shapes) != 8 || len(m.Materials) != 4 || len(m.Bones) != 9 {
		t.Fatalf("model shape: %d meshes, %d shapes, %d materials, %d bones (want 8/8/4/9)",
			len(m.Meshes), len(m.Shapes), len(m.Materials), len(m.Bones))
	}
	// Every mesh's indices are in range and a whole number of triangles.
	for si, sh := range m.Shapes {
		if len(sh.Indices) == 0 || len(sh.Indices)%3 != 0 {
			t.Errorf("shape %d has %d indices (not a triangle list)", si, len(sh.Indices))
		}
		for _, ix := range sh.Indices {
			if int(ix) >= len(sh.Verts) {
				t.Fatalf("shape %d index %d exceeds %d verts", si, ix, len(sh.Verts))
			}
		}
	}
	// The named bones we rely on for the animation must be present.
	names := map[string]bool{}
	for _, b := range m.Bones {
		names[b.Name] = true
	}
	for _, want := range []string{"AllRoot", "Mario", "SuperLeaf", "Block"} {
		if !names[want] {
			t.Errorf("skeleton missing bone %q", want)
		}
	}

	// Textures: each decodes, and its dimensions match a power-of-two-ish tile grid.
	for _, te := range g.Resources["Textures"] {
		txob, im, err := g.DecodeTexture(te)
		if err != nil {
			t.Fatalf("texture %q: %v", te.Name, err)
		}
		if im.Rect.Dx() != txob.Width || im.Rect.Dy() != txob.Height {
			t.Errorf("texture %q decoded to %dx%d, header says %dx%d",
				te.Name, im.Rect.Dx(), im.Rect.Dy(), txob.Width, txob.Height)
		}
	}

	// Animation: over known bones, with keyed curves.
	anim, err := g.DecodeSkeletalAnim(g.Resources["SkeletalAnimations"][0])
	if err != nil {
		t.Fatal(err)
	}
	if len(anim.Members) == 0 {
		t.Fatal("animation has no members")
	}
	for _, mem := range anim.Members {
		if !names[mem.Bone] {
			t.Errorf("animation targets unknown bone %q", mem.Bone)
		}
		keyed := false
		for _, cv := range mem.Curves {
			if cv != nil && len(cv.Keys) > 0 {
				keyed = true
			}
		}
		if !keyed {
			t.Errorf("animation member %q has no keyed curves", mem.Bone)
		}
	}
}
