package n3ds

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// A stored (uncompressed) BLZ file has a zero originalBottom and decodes to
// itself.
func TestDecompressBLZStored(t *testing.T) {
	in := []byte("hello world, not compressed\x00\x00\x00\x00")
	binary.LittleEndian.PutUint32(in[len(in)-4:], 0) // originalBottom = 0

	out, err := DecompressBLZ(in)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, in) {
		t.Fatalf("stored file changed: %q -> %q", in, out)
	}
}

// A hand-assembled stream, so the decoder is checked against bytes derived from
// the format description rather than against an encoder that could share its
// bugs.
//
// The intended output is a 4-byte verbatim head "0123" followed by "abcdefgh"
// repeated four times (36 bytes total). The encoder's view, reading the tokens
// in the order the decoder consumes them (from the top of the stream down):
//
//	flag 0x00  — eight literals: 'h','g','f','e','d','c','b','a'  -> out[35..28]
//	flag 0xC0  — two matches:
//	             0xF005: length 18, distance 8                    -> out[27..10]
//	             0x3005: length  6, distance 8                    -> out[ 9.. 4]
//
// Laid out in ascending memory (the decoder walks it backwards), a match token's
// high byte sits above its low byte, and each flag byte sits above the eight
// tokens it governs.
func TestDecompressBLZHandBuilt(t *testing.T) {
	const (
		headSize   = 4
		footerSize = 8
		rawLen     = 36
	)

	stream := []byte{
		0x05, 0x30, // match B: tok 0x3005 (len 6, dist 8)
		0x05, 0xF0, // match A: tok 0xF005 (len 18, dist 8)
		0xC0,                                   // flag: both tokens above are matches
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', // literals, lowest = last consumed
		0x00, // flag: all eight tokens below are literals
	}

	pak := make([]byte, 0, headSize+len(stream)+footerSize)
	pak = append(pak, '0', '1', '2', '3') // the verbatim head
	pak = append(pak, stream...)

	compressedSize := len(stream) + footerSize
	originalBottom := rawLen - (headSize + len(stream) + footerSize)

	var footer [8]byte
	binary.LittleEndian.PutUint32(footer[0:], uint32(compressedSize)|uint32(footerSize)<<24)
	binary.LittleEndian.PutUint32(footer[4:], uint32(originalBottom))
	pak = append(pak, footer[:]...)

	out, err := DecompressBLZ(pak)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("0123" + "abcdefgh" + "abcdefgh" + "abcdefgh" + "abcdefgh")
	if !bytes.Equal(out, want) {
		t.Fatalf("decompressed:\n got %q\nwant %q", out, want)
	}
}

func TestDecompressBLZRejectsBadFooter(t *testing.T) {
	for _, tc := range []struct {
		name              string
		topBottom, bottom uint32
	}{
		{"footer size too small", 0x00000010, 4},
		{"footer size too large", 0x20000010, 4},
		{"compressed size past the buffer", 0x08FFFFFF, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := make([]byte, 32)
			binary.LittleEndian.PutUint32(in[len(in)-8:], tc.topBottom)
			binary.LittleEndian.PutUint32(in[len(in)-4:], tc.bottom)
			if _, err := DecompressBLZ(in); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

func TestDecompressBLZRejectsShortInput(t *testing.T) {
	if _, err := DecompressBLZ([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected an error for a 3-byte input")
	}
}
