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
