package nds

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestDecompressBLZ decompresses a real (tiny) BLZ vector: Mario Kart DS's fourth
// ARM9 overlay is a 20-byte BSS-init stub that decompresses to 32 zero bytes. This
// pins the backward-LZSS decode against known-good input/output.
func TestDecompressBLZ(t *testing.T) {
	compressed := []byte{
		0x05, 0x50, 0x09, 0x90, 0x03, 0x30, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x1E, 0x14, 0x00, 0x00, 0x08, 0x0C, 0x00, 0x00, 0x00,
	}
	if !IsBLZ(compressed) {
		t.Fatal("IsBLZ = false, want true for a valid BLZ footer")
	}
	got := DecompressBLZ(compressed)
	want := make([]byte, 32) // all zero
	if !bytes.Equal(got, want) {
		t.Fatalf("DecompressBLZ = %d bytes %x, want 32 zero bytes", len(got), got)
	}
}

// TestBLZUncompressed checks the inc_len==0 passthrough (data that is not compressed).
func TestBLZUncompressed(t *testing.T) {
	data := make([]byte, 16)
	copy(data, []byte("hello"))
	binary.LittleEndian.PutUint32(data[12:], 0) // inc_len == 0 → not compressed
	if IsBLZ(data) {
		t.Fatal("IsBLZ = true for inc_len==0 data")
	}
	got := DecompressBLZ(data)
	if !bytes.Equal(got, data) {
		t.Fatalf("DecompressBLZ of uncompressed data changed it")
	}
}
