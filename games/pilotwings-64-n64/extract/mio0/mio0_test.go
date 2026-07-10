package mio0

import (
	"bytes"
	"testing"
)

// TestDecompress hand-assembles a minimal stream: three literals then a
// six-byte back-reference at distance 3, expanding "abc" to "abcabcabc".
// The distance field stores distance-1 and the length field length-3.
func TestDecompress(t *testing.T) {
	src := []byte{
		'M', 'I', 'O', '0',
		0, 0, 0, 9, // decompressed length
		0, 0, 0, 20, // back-reference stream offset
		0, 0, 0, 22, // literal stream offset
		0xE0, 0x00, 0x00, 0x00, // flags: literal, literal, literal, back-ref
		0x30, 0x02, // len 6, dist 3
		'a', 'b', 'c',
	}
	out, err := Decompress(src)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, []byte("abcabcabc")) {
		t.Fatalf("got %q", out)
	}
}

func TestDecompressRejectsBadMagic(t *testing.T) {
	if _, err := Decompress([]byte("MIO1aaaaaaaaaaaaaaaa")); err == nil {
		t.Fatal("bad magic accepted")
	}
}
