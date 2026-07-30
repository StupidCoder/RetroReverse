package assets

// Synthetic tests only: the disc's own files are exercised by ctarc runs
// (the boot art and the EAST SIDE POLICE billboard were judged by eye);
// these pin the parsing contracts so a refactor cannot silently bend them.

import (
	"encoding/binary"
	"testing"
)

func buildAFS(entries [][]byte) []byte {
	const dataBase = 0x1000
	out := make([]byte, dataBase)
	copy(out, "AFS\x00")
	binary.LittleEndian.PutUint32(out[4:], uint32(len(entries)))
	off := dataBase
	for i, e := range entries {
		binary.LittleEndian.PutUint32(out[8+8*i:], uint32(off))
		binary.LittleEndian.PutUint32(out[12+8*i:], uint32(len(e)))
		off += len(e)
	}
	for _, e := range entries {
		out = append(out, e...)
	}
	return out
}

func TestAFSRoundTrip(t *testing.T) {
	want := [][]byte{[]byte("first"), []byte("second entry"), {}}
	a, err := OpenAFS(buildAFS(want))
	if err != nil {
		t.Fatalf("OpenAFS: %v", err)
	}
	if len(a.Entries) != len(want) {
		t.Fatalf("%d entries, want %d", len(a.Entries), len(want))
	}
	for i, w := range want {
		got, err := a.Data(i)
		if err != nil {
			t.Fatalf("Data(%d): %v", i, err)
		}
		if string(got) != string(w) {
			t.Errorf("entry %d = %q, want %q", i, got, w)
		}
	}
	if _, err := a.Data(len(want)); err == nil {
		t.Error("an out-of-range entry was served")
	}
}

func TestAFSRefusesLies(t *testing.T) {
	if _, err := OpenAFS([]byte("NOTA")); err == nil {
		t.Error("no magic accepted")
	}
	// An entry that overruns the file must be an error, not a panic later.
	b := buildAFS([][]byte{[]byte("x")})
	binary.LittleEndian.PutUint32(b[12:], 1<<30)
	if _, err := OpenAFS(b); err == nil {
		t.Error("an overrunning entry was accepted")
	}
}

// TestPVRTwiddledDecode pins the twiddle order with two independently-known
// cases (the stored-registers lesson: pin a decode with cases you can hand-
// compute): texel (0,0) is stored first, and texel (1,0) — x interleaves
// into the ODD bits — is stored at index 2, after (0,1) at index 1.
func TestPVRTwiddledDecode(t *testing.T) {
	// An 8x8 ARGB4444 twiddled texture: texel value = its storage index.
	data := make([]byte, 16+8*8*2)
	copy(data, "PVRT")
	data[8] = 2 // 4444
	data[9] = 1 // square twiddled
	binary.LittleEndian.PutUint16(data[12:], 8)
	binary.LittleEndian.PutUint16(data[14:], 8)
	for i := 0; i < 64; i++ {
		binary.LittleEndian.PutUint16(data[16+i*2:], uint16(0xF000|i))
	}
	tex, err := OpenPVR(data)
	if err != nil {
		t.Fatalf("OpenPVR: %v", err)
	}
	img, err := tex.Decode()
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	// index 0 -> (0,0): 4444 low nibbles 0,0,0 -> black. index 1 -> (0,1):
	// value 0xF001, b nibble 1 -> 0x11. index 2 -> (1,0): b nibble 2 -> 0x22.
	if b := img.Pix[img.PixOffset(0, 0)+2]; b != 0x00 {
		t.Errorf("texel (0,0) blue = %02x, want 00", b)
	}
	if b := img.Pix[img.PixOffset(0, 1)+2]; b != 0x11 {
		t.Errorf("texel (0,1) blue = %02x, want 11 (y interleaves into the even bits)", b)
	}
	if b := img.Pix[img.PixOffset(1, 0)+2]; b != 0x22 {
		t.Errorf("texel (1,0) blue = %02x, want 22 (x interleaves into the odd bits)", b)
	}
}

func TestDecode16Bounds(t *testing.T) {
	if _, err := Decode16(make([]byte, 10), 1, 4, 4); err == nil {
		t.Error("a page shorter than w*h*2 was accepted")
	}
	img, err := Decode16([]byte{0xFF, 0xFF, 0x00, 0x00}, 1, 2, 1)
	if err != nil {
		t.Fatalf("Decode16: %v", err)
	}
	if img.Pix[0] != 0xF8 || img.Pix[4] != 0 {
		t.Errorf("565 decode wrong: % x", img.Pix[:8])
	}
}
