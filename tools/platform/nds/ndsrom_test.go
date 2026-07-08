package nds

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildImage hand-assembles a minimal but valid cartridge image: a header, a FAT with
// two files, and an FNT with the root holding file "a.bin" and a subdir "sub" holding
// "b.bin". It exercises the directory-tree walk, sequential file-ID assignment and the
// header CRC without needing a copyrighted ROM.
func buildImage() []byte {
	le := binary.LittleEndian
	img := make([]byte, 0x400)

	// Header identity + capacity.
	copy(img[0x00:], "TESTROM")
	copy(img[0x0C:], "TEST")
	copy(img[0x10:], "01")
	img[0x14] = 0x02 // 128KB << 2 = 512KB nominal chip

	const fntOff, fatOff = 0x200, 0x2C0
	le.PutUint32(img[0x40:], fntOff) // FNT offset
	le.PutUint32(img[0x44:], 0x30)   // FNT size
	le.PutUint32(img[0x48:], fatOff) // FAT offset
	le.PutUint32(img[0x4C:], 0x10)   // FAT size → 2 files

	// FAT: file 0 = "ABCD" at 0x300, file 1 = "wxyz" at 0x304.
	le.PutUint32(img[fatOff+0:], 0x300)
	le.PutUint32(img[fatOff+4:], 0x304)
	le.PutUint32(img[fatOff+8:], 0x304)
	le.PutUint32(img[fatOff+12:], 0x308)
	copy(img[0x300:], "ABCDwxyz")

	// FNT main table: root record then the subdir record.
	le.PutUint32(img[fntOff+0:], 0x10) // root sub-table at FNT+0x10
	le.PutUint16(img[fntOff+4:], 0)    // root first file id
	le.PutUint16(img[fntOff+6:], 2)    // total directory count
	le.PutUint32(img[fntOff+8:], 0x1E) // subdir sub-table at FNT+0x1E
	le.PutUint16(img[fntOff+12:], 1)   // subdir first file id
	le.PutUint16(img[fntOff+14:], 0xF000)

	// Root sub-table (FNT+0x10): file "a.bin", subdir "sub" (child 0xF001), end.
	p := fntOff + 0x10
	img[p] = 0x05
	copy(img[p+1:], "a.bin")
	p += 6
	img[p] = 0x83
	copy(img[p+1:], "sub")
	le.PutUint16(img[p+4:], 0xF001)
	p += 6
	img[p] = 0x00

	// Subdir sub-table (FNT+0x1E): file "b.bin", end.
	p = fntOff + 0x1E
	img[p] = 0x05
	copy(img[p+1:], "b.bin")
	img[p+6] = 0x00

	// Header checksum over 0x00..0x15D.
	le.PutUint16(img[0x15E:], CRC16(img[0x00:0x15E]))
	return img
}

func TestOpen(t *testing.T) {
	rom, err := Open(buildImage())
	if err != nil {
		t.Fatal(err)
	}
	if rom.Header.Title != "TESTROM" || rom.Header.GameCode != "TEST" {
		t.Fatalf("header identity = %q/%q", rom.Header.Title, rom.Header.GameCode)
	}
	if rom.Header.ChipBytes() != 512*1024 {
		t.Fatalf("chip size = %d, want 512KB", rom.Header.ChipBytes())
	}
	if _, ok := rom.VerifyHeaderCRC(); !ok {
		t.Fatalf("header CRC failed to verify")
	}
	if len(rom.Files) != 2 {
		t.Fatalf("named files = %d, want 2", len(rom.Files))
	}
	if got := rom.FileByPath("a.bin"); !bytes.Equal(got, []byte("ABCD")) {
		t.Fatalf("a.bin = %q, want ABCD", got)
	}
	if got := rom.FileByPath("sub/b.bin"); !bytes.Equal(got, []byte("wxyz")) {
		t.Fatalf("sub/b.bin = %q, want wxyz", got)
	}
	if rom.File(0) == nil || string(rom.File(1)) != "wxyz" {
		t.Fatalf("File() by id mismatch")
	}
}

// TestCRC16 checks the CRC-16/MODBUS against a known vector ("123456789" → 0x4B37).
func TestCRC16(t *testing.T) {
	if got := CRC16([]byte("123456789")); got != 0x4B37 {
		t.Fatalf("CRC16 = 0x%04X, want 0x4B37", got)
	}
}
