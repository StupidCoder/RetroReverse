package crit

import "testing"

func TestUnpack5MSBFirst(t *testing.T) {
	// bytes 0x87,0x65 = 10000111 01100101 -> 5-bit codes 16, 29, 18.
	got := unpack5([]byte{0x87, 0x65}, 3)
	want := []byte{16, 29, 18}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("code %d = %d, want %d (%v)", i, got[i], want[i], got)
		}
	}
}

func TestCodeBits(t *testing.T) {
	for f, want := range map[byte]int{6: 5, 8: 4, 4: 8} {
		if got := codeBits(f); got != want {
			t.Errorf("codeBits(%d) = %d, want %d", f, got, want)
		}
	}
}
