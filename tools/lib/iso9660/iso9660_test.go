package iso9660

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildCooked assembles a minimal but valid ISO 9660 volume: a PVD at LBA 16, a
// terminator at 17, a root directory at 18 holding one subdirectory and one file,
// the subdirectory at 19 holding one file, and the two file extents at 20 and 21.
func buildCooked() []byte {
	const (
		rootLBA = 18
		subLBA  = 19
		fileLBA = 20
		deepLBA = 21
		total   = 22
	)
	helloBody := []byte("BOOT2=cdrom0:\\SCUS_971.24;1\r\n")
	deepBody := bytes.Repeat([]byte("x"), BlockSize+7) // spans two blocks

	img := make([]byte, total*BlockSize+2*BlockSize)

	// --- Primary Volume Descriptor at LBA 16 ---
	pvd := img[16*BlockSize:]
	pvd[0] = 1
	copy(pvd[1:6], "CD001")
	pvd[6] = 1
	copy(pvd[8:40], []byte("PLAYSTATION                     "))
	copy(pvd[40:72], []byte("JAKTEST                         "))
	putBoth32(pvd[80:], total+2) // volume space size, in logical blocks
	putBoth16(pvd[128:], BlockSize)
	// Root directory record, 34 bytes at offset 156.
	writeDirRec(pvd[156:], "\x00", rootLBA, 1*BlockSize, true)

	// --- Volume descriptor set terminator at LBA 17 ---
	term := img[17*BlockSize:]
	term[0] = 0xFF
	copy(term[1:6], "CD001")
	term[6] = 1

	// --- Root directory at LBA 18 ---
	root := img[rootLBA*BlockSize:]
	p := 0
	p += writeDirRec(root[p:], "\x00", rootLBA, BlockSize, true)  // .
	p += writeDirRec(root[p:], "\x01", rootLBA, BlockSize, true)  // ..
	p += writeDirRec(root[p:], "SYSTEM.CNF;1", fileLBA, len(helloBody), false)
	p += writeDirRec(root[p:], "DGO", subLBA, BlockSize, true)

	// --- Subdirectory at LBA 19 ---
	sub := img[subLBA*BlockSize:]
	p = 0
	p += writeDirRec(sub[p:], "\x00", subLBA, BlockSize, true)
	p += writeDirRec(sub[p:], "\x01", rootLBA, BlockSize, true)
	p += writeDirRec(sub[p:], "VI1.DGO;1", deepLBA, len(deepBody), false)

	copy(img[fileLBA*BlockSize:], helloBody)
	copy(img[deepLBA*BlockSize:], deepBody)
	return img
}

func putBoth32(b []byte, v int) {
	binary.LittleEndian.PutUint32(b[0:], uint32(v))
	binary.BigEndian.PutUint32(b[4:], uint32(v))
}

func putBoth16(b []byte, v int) {
	binary.LittleEndian.PutUint16(b[0:], uint16(v))
	binary.BigEndian.PutUint16(b[2:], uint16(v))
}

// writeDirRec lays down one directory record and returns its length.
func writeDirRec(b []byte, name string, lba, size int, isDir bool) int {
	n := 33 + len(name)
	if n%2 == 1 {
		n++ // records are padded to an even length
	}
	b[0] = byte(n)
	putBoth32(b[2:], lba)
	putBoth32(b[10:], size)
	if isDir {
		b[25] = 0x02
	}
	b[32] = byte(len(name))
	copy(b[33:], name)
	return n
}

// rewrap re-lays a cooked image under g: each 2048-byte block is placed at
// SectorSize*n + DataOffset, with the surrounding sector bytes left as filler.
func rewrap(cooked []byte, g Geometry) []byte {
	blocks := len(cooked) / BlockSize
	out := make([]byte, blocks*g.SectorSize)
	for i := 0; i < blocks; i++ {
		copy(out[i*g.SectorSize+g.DataOffset:], cooked[i*BlockSize:(i+1)*BlockSize])
	}
	return out
}

func TestDetectAndRead(t *testing.T) {
	cooked := buildCooked()

	for _, g := range geometries {
		img := rewrap(cooked, g)

		got, err := Detect(newByteReaderAt(img), int64(len(img)))
		if err != nil {
			t.Errorf("%v: Detect: %v", g, err)
			continue
		}
		if got != g {
			// Several layouts can place a plausible-looking PVD; the volume-size
			// check is what disambiguates. If this fires, that check is too weak.
			t.Errorf("%v: Detect returned %v", g, got)
			continue
		}

		v, err := OpenBytes(img)
		if err != nil {
			t.Errorf("%v: OpenBytes: %v", g, err)
			continue
		}
		if v.System != "PLAYSTATION" {
			t.Errorf("%v: system id = %q, want PLAYSTATION", g, v.System)
		}

		body, err := v.ReadFile("SYSTEM.CNF")
		if err != nil {
			t.Errorf("%v: ReadFile: %v", g, err)
			continue
		}
		if want := "BOOT2=cdrom0:\\SCUS_971.24;1\r\n"; string(body) != want {
			t.Errorf("%v: SYSTEM.CNF = %q, want %q", g, body, want)
		}
	}
}

func TestWalkAndMultiBlockFile(t *testing.T) {
	v, err := OpenBytes(buildCooked())
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	if err := v.Walk(func(e Entry) error {
		paths = append(paths, e.Path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	want := []string{"SYSTEM.CNF;1", "DGO", "DGO/VI1.DGO;1"}
	if len(paths) != len(want) {
		t.Fatalf("walk found %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("walk[%d] = %q, want %q", i, paths[i], want[i])
		}
	}

	// A file that spans two logical blocks must be reassembled exactly, and its
	// length must come from the directory record, not the block count.
	deep, err := v.ReadFile("/DGO/VI1.DGO")
	if err != nil {
		t.Fatal(err)
	}
	if len(deep) != BlockSize+7 {
		t.Fatalf("VI1.DGO is %d bytes, want %d", len(deep), BlockSize+7)
	}
	if deep[BlockSize+6] != 'x' || bytes.Contains(deep, []byte{0}) {
		t.Errorf("VI1.DGO body is not the expected filler")
	}
}

func TestResolveAcceptsBootPaths(t *testing.T) {
	v, err := OpenBytes(buildCooked())
	if err != nil {
		t.Fatal(err)
	}
	// The path SYSTEM.CNF names the boot ELF with is a device path with backslashes.
	for _, p := range []string{"cdrom0:\\SYSTEM.CNF;1", "\\SYSTEM.CNF", "/SYSTEM.CNF;1", "system.cnf"} {
		if _, err := v.Resolve(p); err != nil {
			t.Errorf("Resolve(%q): %v", p, err)
		}
	}
}

func TestFileAt(t *testing.T) {
	v, err := OpenBytes(buildCooked())
	if err != nil {
		t.Fatal(err)
	}
	// LBA 21 is the first block of the two-block DGO; 22 is its second.
	for _, lba := range []int{21, 22} {
		e, ok := v.FileAt(lba)
		if !ok || e.Path != "DGO/VI1.DGO;1" {
			t.Errorf("FileAt(%d) = %q, %v; want DGO/VI1.DGO;1", lba, e.Path, ok)
		}
	}
	if _, ok := v.FileAt(17); ok {
		t.Errorf("FileAt(17) resolved to a file, but LBA 17 is the descriptor terminator")
	}
}

func TestDetectRejectsGarbage(t *testing.T) {
	if _, err := Detect(newByteReaderAt(bytes.Repeat([]byte{0x25}, 4<<20)), 4<<20); err == nil {
		t.Error("Detect accepted an image with no volume descriptor")
	}
}
