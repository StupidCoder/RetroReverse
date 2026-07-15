package gc

import (
	"encoding/binary"
	"os"
	"testing"
)

// The retail disc, when it is present. Everything that needs it skips without it, so
// the suite runs on a clean checkout (the image is not committed — see the repository
// copyright policy).
const discPath = "../../../games/luigis-mansion-gc/image/Luigi's Mansion (USA).iso"

func openDisc(t *testing.T) *Disc {
	t.Helper()
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the disc image is not present")
	}
	d, err := Open(discPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestDiscHeader(t *testing.T) {
	d := openDisc(t)
	if d.Header.GameID != "GLME01" {
		t.Errorf("game id = %q, want GLME01", d.Header.GameID)
	}
	if d.Header.Title != "LUIGI'S MANSION" {
		t.Errorf("title = %q", d.Header.Title)
	}
	if d.Apploader.Entry != 0x81200194 {
		t.Errorf("apploader entry = %#08X, want 0x81200194", d.Apploader.Entry)
	}
	if d.Header.DOLOffset != 0x015600 || d.Header.FSTOffset != 0x3B5100 {
		t.Errorf("dol at %#X, fst at %#X", d.Header.DOLOffset, d.Header.FSTOffset)
	}
}

// The apploader is PowerPC, and its first instructions are the standard prologue. If
// this passes, the loader really is code and really is big-endian — which is the whole
// premise of running it rather than emulating what it would have done.
func TestApploaderIsPowerPC(t *testing.T) {
	d := openDisc(t)
	code, err := d.ApploaderCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != d.Apploader.Body() {
		t.Fatalf("read %d bytes of loader, want %d", len(code), d.Apploader.Body())
	}
	want := []uint32{
		0x7C0802A6, // mflr r0
		0x90010004, // stw  r0,4(r1)
		0x9421FFE0, // stwu r1,-32(r1)
	}
	for i, w := range want {
		if got := binary.BigEndian.Uint32(code[i*4:]); got != w {
			t.Errorf("loader word %d = %#08X, want %#08X", i, got, w)
		}
	}
}

func TestDOL(t *testing.T) {
	d := openDisc(t)
	dol, err := d.DOL()
	if err != nil {
		t.Fatal(err)
	}
	if dol.Entry != 0x80003100 {
		t.Errorf("entry = %#08X, want 0x80003100", dol.Entry)
	}
	// The executable's segments must not overlap each other, and the entry must land
	// in one of the code segments — otherwise the game would begin by executing data.
	if !dol.Text(dol.Entry) {
		t.Errorf("the entry point %#08X is not in a text segment", dol.Entry)
	}
	for i, a := range dol.Segments {
		for _, b := range dol.Segments[i+1:] {
			if a.Addr < b.Addr+b.Size && b.Addr < a.Addr+a.Size {
				t.Errorf("%s (%#08X+%#X) overlaps %s (%#08X+%#X)", a.Name(), a.Addr, a.Size, b.Name(), b.Addr, b.Size)
			}
		}
	}
	// Flat must reproduce each segment where the segment says it lives.
	base, mem := dol.Flat()
	for _, s := range dol.Segments {
		got := mem[s.Addr-base : s.Addr-base+s.Size]
		if string(got) != string(s.Data) {
			t.Errorf("%s is not where Flat put it", s.Name())
		}
	}
}

// The extent check is the proof that the filesystem was read correctly: a misparse
// yields wild offsets, and wild offsets leave the disc or collide with a neighbour.
func TestFSTValidates(t *testing.T) {
	d := openDisc(t)
	if err := d.FST.Validate(d.Size); err != nil {
		t.Fatal(err)
	}
	if n := len(d.FST.Files()); n != 847 {
		t.Errorf("%d files, want 847", n)
	}
	// A file the disc is known to carry, found by path, and then found again by the
	// offset it lives at — which is the round trip the oracle's disc log depends on.
	f, ok := d.FST.ByPath("/Game/game.szp")
	if !ok {
		t.Fatal("/Game/game.szp is not in the filesystem")
	}
	got, within, ok := d.FST.ByOffset(f.Offset + 16)
	if !ok || got.Path != f.Path || within != 16 {
		t.Errorf("ByOffset(%#X+16) = %q+%d, %v", f.Offset, got.Path, within, ok)
	}
	// The byte before a file belongs to no file, or to the one before it — never to
	// this one.
	if got, _, ok := d.FST.ByOffset(f.Offset - 1); ok && got.Path == f.Path {
		t.Errorf("the byte before %s is claimed by it", f.Path)
	}
}

// fstEntry builds one raw 12-byte entry.
func fstEntry(dir bool, nameOff, a, c uint32) []byte {
	b := make([]byte, fstEntrySize)
	if dir {
		b[0] = 1
	}
	b[1], b[2], b[3] = byte(nameOff>>16), byte(nameOff>>8), byte(nameOff)
	binary.BigEndian.PutUint32(b[4:], a)
	binary.BigEndian.PutUint32(b[8:], c)
	return b
}

// A hand-built filesystem, so the flattening is tested independently of the disc:
//
//	/            entries 0..3
//	  dir/       entry 1, holding entry 2
//	    a.bin    entry 2
//	  b.bin      entry 3
func TestParseFSTHierarchy(t *testing.T) {
	names := []byte("dir\x00a.bin\x00b.bin\x00")
	var b []byte
	b = append(b, fstEntry(true, 0, 0, 4)...) // root: 4 entries
	b = append(b, fstEntry(true, 0, 0, 3)...) // dir/, ends before 3
	b = append(b, fstEntry(false, 4, 0x1000, 0x10)...)
	b = append(b, fstEntry(false, 10, 0x2000, 0x20)...)
	b = append(b, names...)

	f, err := ParseFST(b)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/", "/dir", "/dir/a.bin", "/b.bin"}
	for i, w := range want {
		if f.Entries[i].Path != w {
			t.Errorf("entry %d = %q, want %q", i, f.Entries[i].Path, w)
		}
	}
	if _, _, ok := f.ByOffset(0x1008); !ok {
		t.Error("0x1008 should be inside /dir/a.bin")
	}
	if _, _, ok := f.ByOffset(0x1800); ok {
		t.Error("0x1800 is between the two files and should be in neither")
	}
}

// A filesystem whose entry count runs past the buffer must be rejected, not read into
// whatever memory happens to follow.
func TestParseFSTRejectsGarbage(t *testing.T) {
	b := fstEntry(true, 0, 0, 9999)
	if _, err := ParseFST(b); err == nil {
		t.Error("a 1-entry buffer claiming 9999 entries was accepted")
	}
}
