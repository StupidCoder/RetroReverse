package tex

import "testing"

// The format-8 decoder was reverse-engineered from the game's own routine
// (07F7:018D) and verified pixel-exact against a live-decoded OBJECTS.GR frame
// (256/256 silhouette match on frame 128). These unit tests pin the core RLE
// paths with hand-built frames whose output can be checked by hand.
//
// A frame is [08][W][H][aux][u16 nibbleCount][packed nibbles]; the nibbles are
// a stream of RUN-then-DUMP operations.

func eq(t *testing.T, got, want []byte) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pixel %d = %d, want %d (%v)", i, got[i], want[i], got)
		}
	}
}

func TestFormat8RunAndDump(t *testing.T) {
	// nibbles: R=4,px=3, D=2,lit9,lit8  => 3 3 3 3 9 8, then padded to 16.
	// packed: 43 29 80  (nibbles 4,3,2,9,8; last byte low-padded)
	frame := []byte{0x08, 4, 4, 0, 0x05, 0x00, 0x43, 0x29, 0x80}
	eq(t, decodeFormat8(frame, 4, 4),
		[]byte{3, 3, 3, 3, 9, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func TestFormat8ExtendedRun(t *testing.T) {
	// R=0 (extended run), E=1, next=0 => count = (1<<4)|0 = 16, px=3 => all 3s.
	// nibbles: 0 1 0 3 -> packed 01 03
	frame := []byte{0x08, 4, 4, 0, 0x04, 0x00, 0x01, 0x03}
	want := make([]byte, 16)
	for i := range want {
		want[i] = 3
	}
	eq(t, decodeFormat8(frame, 4, 4), want)
}

func TestFormat8ExhaustionPads(t *testing.T) {
	// A short frame must still return exactly W*H pixels, zero-padded.
	frame := []byte{0x08, 2, 2, 0, 0x02, 0x00, 0x53} // R=5,px=3 -> 4 threes fill 2x2
	eq(t, decodeFormat8(frame, 2, 2), []byte{3, 3, 3, 3})
}
