package nds

import (
	"bytes"
	"testing"
)

func TestLZ77(t *testing.T) {
	// literal-only: header (type 0x10, size 4) + one flag byte (all-literal) + "AAAA"
	lit := []byte{0x10, 0x04, 0x00, 0x00, 0x00, 'A', 'A', 'A', 'A'}
	if got, err := DecompressLZ77(lit); err != nil || !bytes.Equal(got, []byte("AAAA")) {
		t.Fatalf("literal decode = %q, %v", got, err)
	}
	// back-reference: 'A' literal, then copy 3 bytes at displacement 1 → "AAAA"
	ref := []byte{0x10, 0x04, 0x00, 0x00, 0x40, 'A', 0x00, 0x00}
	if got, err := DecompressLZ77(ref); err != nil || !bytes.Equal(got, []byte("AAAA")) {
		t.Fatalf("back-ref decode = %q, %v", got, err)
	}
	if !IsLZ77(lit) || IsLZ77([]byte{0x00, 1, 2, 3}) {
		t.Fatalf("IsLZ77 classification wrong")
	}
}
