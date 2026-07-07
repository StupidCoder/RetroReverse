package threedo

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildImage assembles a minimal cooked (2048-byte-per-block) Opera image with a
// volume label at block 0, a single-block root directory at block 2 holding one
// file entry, and that file's data at block 3. It exercises every field the
// reader depends on without needing the real (git-ignored) disc.
func buildImage(t *testing.T, fileName string, fileData []byte) []byte {
	t.Helper()
	const nblocks = 4
	img := make([]byte, nblocks*userSize)
	be := binary.BigEndian

	// --- block 0: volume label ---
	lbl := img[0:userSize]
	lbl[0] = 0x01
	copy(lbl[1:6], []byte{0x5A, 0x5A, 0x5A, 0x5A, 0x5A})
	lbl[6] = 0x01
	copy(lbl[40:72], []byte("TEST"))
	be.PutUint32(lbl[72:], 0xCAFEBABE) // id
	be.PutUint32(lbl[76:], userSize)   // block size
	be.PutUint32(lbl[80:], nblocks)    // block count
	be.PutUint32(lbl[84:], 0x1234)     // root dir id
	be.PutUint32(lbl[88:], 1)          // root dir blocks
	be.PutUint32(lbl[92:], userSize)   // root dir block size
	be.PutUint32(lbl[96:], 0)          // last root copy index (=> 1 copy)
	be.PutUint32(lbl[100:], 2)         // root copies[0] -> block 2

	// --- block 2: root directory (one file entry) ---
	dir := img[2*userSize : 3*userSize]
	be.PutUint32(dir[0:], 0xFFFFFFFF) // next = -1
	be.PutUint32(dir[4:], 0xFFFFFFFF) // prev = -1
	be.PutUint32(dir[8:], 0)          // flags
	entryLen := 0x44 + 4              // fixed header + one avatar
	be.PutUint32(dir[12:], uint32(20+entryLen)) // endOffset (first free byte)
	be.PutUint32(dir[16:], 20)                  // first entry offset

	e := dir[20:]
	be.PutUint32(e[0:], 0x02)              // flags: file
	be.PutUint32(e[4:], 0x9999)            // id
	copy(e[8:12], []byte("Txt"))           // type
	be.PutUint32(e[12:], userSize)         // block size
	be.PutUint32(e[16:], uint32(len(fileData)))
	be.PutUint32(e[20:], 1)                // block count
	copy(e[0x20:0x40], []byte(fileName))   // name
	be.PutUint32(e[0x40:], 0)              // last avatar index (=> 1 copy)
	be.PutUint32(e[0x44:], 3)              // avatar[0] -> block 3

	// --- block 3: file data ---
	copy(img[3*userSize:], fileData)
	return img
}

func TestOperaFS(t *testing.T) {
	data := []byte("hello, stygian abyss? no — the open road.")
	img := buildImage(t, "README", data)

	v, err := Open(img)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if v.Label != "TEST" {
		t.Errorf("label = %q, want TEST", v.Label)
	}
	if v.ID != 0xCAFEBABE {
		t.Errorf("id = 0x%08X, want 0xCAFEBABE", v.ID)
	}

	var names []string
	if err := v.Walk(func(e Entry) error {
		names = append(names, e.Path)
		if e.Path == "README" {
			if e.Size != len(data) {
				t.Errorf("README size = %d, want %d", e.Size, len(data))
			}
			if e.Type != "Txt" {
				t.Errorf("README type = %q, want Txt", e.Type)
			}
			if e.Block != 3 {
				t.Errorf("README block = %d, want 3", e.Block)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(names) != 1 || names[0] != "README" {
		t.Fatalf("Walk entries = %v, want [README]", names)
	}

	got, err := v.ReadFile("README")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadFile = %q, want %q", got, data)
	}

	// A path miss should be a clean error, not a panic.
	if _, err := v.ReadFile("nope"); err == nil {
		t.Errorf("ReadFile(nope) succeeded, want error")
	}
}

func TestOpenRejectsNonOpera(t *testing.T) {
	junk := make([]byte, 4*userSize) // valid geometry, no label
	if _, err := Open(junk); err == nil {
		t.Errorf("Open(junk) succeeded, want error")
	}
}
