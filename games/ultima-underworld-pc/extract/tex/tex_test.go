package tex

import (
	"encoding/binary"
	"testing"
)

func TestParseTRAndImage(t *testing.T) {
	// A tiny 2x2, one-texture .TR: pixels {0,1,2,3}.
	var d []byte
	d = append(d, 2, 2)                                  // version, dim=2
	d = binary.LittleEndian.AppendUint16(d, 1)           // count
	d = binary.LittleEndian.AppendUint32(d, uint32(4+4)) // offset to pixels
	d = append(d, 0, 1, 2, 3)                            // 2x2 palette indices
	tr, err := ParseTR(d)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Dim != 2 || tr.Count() != 1 {
		t.Fatalf("dim=%d count=%d, want 2/1", tr.Dim, tr.Count())
	}
	// palette: index i -> grey (i*10)
	pals := make([]byte, 768)
	for i := 0; i < 256; i++ {
		pals[i*3] = byte(i)
	}
	pal, _ := LoadPalette(pals, 0)
	img, err := tr.Image(0, pal)
	if err != nil {
		t.Fatal(err)
	}
	if img.Pix[0] != 0 || img.Pix[4] != 1<<2 { // index 1 -> R=1<<2
		t.Errorf("pixel decode wrong: %d %d", img.Pix[0], img.Pix[4])
	}
}
