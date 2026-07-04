package nds

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestParseNARC builds a minimal NARC with two files and checks they round-trip.
func TestParseNARC(t *testing.T) {
	le := binary.LittleEndian
	img := []byte("HELLOworld!") // file0="HELLO" (0..5), file1="world!" (5..11)
	btaf := make([]byte, 0x0C+2*8)
	copy(btaf, "BTAF")
	le.PutUint32(btaf[4:], uint32(len(btaf)))
	le.PutUint16(btaf[8:], 2) // 2 files
	le.PutUint32(btaf[0x0C:], 0)
	le.PutUint32(btaf[0x10:], 5)
	le.PutUint32(btaf[0x14:], 5)
	le.PutUint32(btaf[0x18:], 11)
	btnf := make([]byte, 8)
	copy(btnf, "BTNF")
	le.PutUint32(btnf[4:], 8)
	gmif := make([]byte, 8+len(img))
	copy(gmif, "GMIF")
	le.PutUint32(gmif[4:], uint32(len(gmif)))
	copy(gmif[8:], img)

	var buf bytes.Buffer
	hdr := make([]byte, 0x10)
	copy(hdr, "NARC")
	le.PutUint16(hdr[4:], 0xFFFE)
	le.PutUint16(hdr[6:], 0x0100)
	le.PutUint16(hdr[0x0C:], 0x10) // header size → first block
	le.PutUint16(hdr[0x0E:], 3)    // 3 blocks
	buf.Write(hdr)
	buf.Write(btaf)
	buf.Write(btnf)
	buf.Write(gmif)
	le.PutUint32(buf.Bytes()[8:], uint32(buf.Len())) // file size

	files, err := ParseNARC(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || string(files[0]) != "HELLO" || string(files[1]) != "world!" {
		t.Fatalf("NARC files = %q", files)
	}
}
