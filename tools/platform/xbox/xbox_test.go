package xbox

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// The retail disc, when present. The image is gitignored (copyright policy), so every
// test that needs it skips without it and the suite stays green on a clean checkout.
const discPath = "../../../games/outrun-2006-xbox/image/OutRun 2006 - Coast 2 Coast (EUR).iso"

// --- synthetic XBE: exercises the header parse + XOR de-obfuscation with no asset ---

func buildXBE(t *testing.T) []byte {
	t.Helper()
	const (
		base      = 0x00010000
		imageSize = 0x00002000
		realEntry = 0x00011000 // inside the image → the retail key must win
		thunkVA   = 0x00010300 // thunk table in the header region
		secHdrVA  = 0x00010180
		nameVA    = 0x00010400
	)
	b := make([]byte, 0x1000)
	copy(b[0:4], "XBEH")
	put32 := func(off, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
	put32(0x104, base)
	put32(0x10C, imageSize)
	put32(0x118, 0)                        // no certificate for this fixture
	put32(0x11C, 1)                        // one section
	put32(0x120, secHdrVA)                 // section headers in the header region
	put32(0x128, realEntry^entryKeyRetail) // obfuscated entry
	put32(0x158, thunkVA^thunkKeyRetail)   // obfuscated thunk pointer

	// One section header at file offset 0x180 (VA secHdrVA).
	sh := 0x180
	put32(uint32(sh+0x00), SecExecutable)
	put32(uint32(sh+0x04), 0x00011000) // vaddr
	put32(uint32(sh+0x08), 0x00001000) // vsize
	put32(uint32(sh+0x0C), 0x00001000) // rawaddr
	put32(uint32(sh+0x10), 0x00001000) // rawsize
	put32(uint32(sh+0x14), nameVA)     // name VA

	// Thunk table at file 0x300: ordinals 1, 5, 3 (out of order, to prove sorting),
	// then a NUL terminator.
	binary.LittleEndian.PutUint32(b[0x300:], 0x80000001)
	binary.LittleEndian.PutUint32(b[0x304:], 0x80000005)
	binary.LittleEndian.PutUint32(b[0x308:], 0x80000003)
	binary.LittleEndian.PutUint32(b[0x30C:], 0)

	copy(b[0x400:], ".text\x00")
	return b
}

func TestParseXBESynthetic(t *testing.T) {
	x, err := ParseXBE(buildXBE(t))
	if err != nil {
		t.Fatal(err)
	}
	if x.Base != 0x00010000 {
		t.Errorf("base = %#x, want 0x10000", x.Base)
	}
	if x.Entry != 0x00011000 {
		t.Errorf("entry = %#x, want 0x11000 (retail key should have won)", x.Entry)
	}
	if !x.Retail {
		t.Error("expected the retail key to be chosen")
	}
	if x.ThunkAddr != 0x00010300 {
		t.Errorf("thunk addr = %#x, want 0x10300", x.ThunkAddr)
	}
	if len(x.Sections) != 1 || x.Sections[0].Name != ".text" {
		t.Fatalf("sections = %+v", x.Sections)
	}
	want := []uint16{1, 3, 5}
	if len(x.Ordinals) != len(want) {
		t.Fatalf("ordinals = %v, want %v", x.Ordinals, want)
	}
	for i, o := range want {
		if x.Ordinals[i] != o {
			t.Errorf("ordinal %d = %d, want %d", i, x.Ordinals[i], o)
		}
	}
}

func TestParseXBERejectsGarbage(t *testing.T) {
	if _, err := ParseXBE([]byte("not an xbe at all, but long enough................................")); err == nil {
		t.Error("a non-XBEH buffer was accepted")
	}
}

// --- synthetic XISO: exercises magic-scan + binary-tree directory walk with no asset ---

// buildXISO writes a one-file XDVDFS image to a temp file and returns its path. The
// partition base is 0, so the volume descriptor sits at sector 32 (0x10000).
func buildXISO(t *testing.T) string {
	t.Helper()
	const (
		rootSector = 33
		fileSector = 34
	)
	fileData := []byte("XBEH-ish payload")
	img := make([]byte, (fileSector+1)*sectorSize+sectorSize)

	// Volume descriptor at sector 32: head magic, root sector/size, tail magic.
	vd := 32 * sectorSize
	copy(img[vd:], xdvdfsMagic)
	binary.LittleEndian.PutUint32(img[vd+0x14:], rootSector)
	binary.LittleEndian.PutUint32(img[vd+0x18:], sectorSize) // root dir is one sector
	copy(img[vd+sectorSize-len(xdvdfsMagic):], xdvdfsMagic)

	// Root directory extent at sector 33: a single tree node (the root) with no children.
	rd := rootSector * sectorSize
	name := "default.xbe"
	binary.LittleEndian.PutUint16(img[rd+0:], 0) // left: none
	binary.LittleEndian.PutUint16(img[rd+2:], 0) // right: none
	binary.LittleEndian.PutUint32(img[rd+4:], fileSector)
	binary.LittleEndian.PutUint32(img[rd+8:], uint32(len(fileData)))
	img[rd+12] = 0 // attr: a regular file
	img[rd+13] = byte(len(name))
	copy(img[rd+14:], name)

	copy(img[fileSector*sectorSize:], fileData)

	path := filepath.Join(t.TempDir(), "synthetic.iso")
	if err := os.WriteFile(path, img, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestXISOSynthetic(t *testing.T) {
	img, err := Open(buildXISO(t))
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	if img.Base != 0 {
		t.Errorf("base = %#x, want 0", img.Base)
	}
	entries, err := img.ReadDir("/")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "default.xbe" || entries[0].IsDir {
		t.Fatalf("root entries = %+v", entries)
	}
	// Resolve by path, case-insensitively, and read the payload back.
	data, err := img.ReadFile("/DEFAULT.XBE")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "XBEH-ish payload" {
		t.Errorf("payload = %q", data)
	}
}

// --- the real image, when present ---

func TestRealDisc(t *testing.T) {
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the OutRun disc image is not present")
	}
	img, err := Open(discPath)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	data, err := img.ReadFile("/default.xbe")
	if err != nil {
		t.Fatal(err)
	}
	if string(data[0:4]) != "XBEH" {
		t.Fatalf("extracted default.xbe does not start with XBEH: %x", data[0:4])
	}
	x, err := ParseXBE(data)
	if err != nil {
		t.Fatal(err)
	}
	if x.Base != 0x00010000 {
		t.Errorf("base = %#x, want the retail 0x10000", x.Base)
	}
	if !x.inImage(x.Entry) {
		t.Errorf("entry %#x is not inside the image", x.Entry)
	}
	if len(x.Ordinals) == 0 {
		t.Error("no xboxkrnl ordinals were recovered")
	}
	if x.TitleName == "" {
		t.Error("no title name was read from the certificate")
	}
}
