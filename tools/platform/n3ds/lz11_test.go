package n3ds

import (
	"bytes"
	"testing"
)

// A hand-assembled LZ11 stream, built from the format description so the decoder
// is checked against an independent source. It encodes "ABABABAB": two literals
// then a 6-byte back-reference at distance 2 (an overlapping copy that expands
// run-length).
//
//	0x11              type
//	0x08 0x00 0x00    decompressed size = 8
//	0x20              flag byte: tokens are literal, literal, back-ref, …
//	'A' 'B'           the two literals
//	0x50 0x01         back-ref: indicator 5 → count 6; disp field 0 → distance 2
func TestDecompressLZ11HandBuilt(t *testing.T) {
	stream := []byte{0x11, 0x08, 0x00, 0x00, 0x20, 'A', 'B', 0x50, 0x01}
	out, err := DecompressLZ11(stream)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte("ABABABAB"); !bytes.Equal(out, want) {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestDecompressLZ11RejectsBadType(t *testing.T) {
	if _, err := DecompressLZ11([]byte{0x10, 0x04, 0x00, 0x00}); err == nil {
		t.Fatal("expected an error for a non-0x11 type byte")
	}
}

// The all-literal path: a flag byte of 0x00 means eight literals in a row.
func TestDecompressLZ11AllLiterals(t *testing.T) {
	stream := []byte{0x11, 0x04, 0x00, 0x00, 0x00, 'W', 'X', 'Y', 'Z'}
	out, err := DecompressLZ11(stream)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte("WXYZ"); !bytes.Equal(out, want) {
		t.Fatalf("got %q, want %q", out, want)
	}
}
