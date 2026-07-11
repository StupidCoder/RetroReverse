package gprs

// The decoder is exercised against the real image's boot directory archive,
// data/first_us.arc. Ground truth is the running game's own decompressor: the
// decoded buffer was captured from the oracle's RAM (the GARC registered at
// 0x09D90350) and compared byte-exact against this decoder's output once; the
// test pins the same invariants (declared size, GARC magic and version, the
// FILE directory chunk) so it needs only the image.

import (
	"os"
	"testing"

	"retroreverse.com/tools/platform/psp"
)

func testImage(t *testing.T) *psp.Image {
	t.Helper()
	for _, p := range []string{os.Getenv("PSP_IMAGE"), "../../image/LocoRoco.cso"} {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			im, err := psp.OpenImage(p)
			if err != nil {
				t.Fatalf("open %s: %v", p, err)
			}
			return im
		}
	}
	t.Skip("no PSP image (set PSP_IMAGE)")
	return nil
}

func TestDecompressFirstUS(t *testing.T) {
	im := testImage(t)
	defer im.Close()
	raw, err := im.ReadFile("PSP_GAME/USRDIR/data/first_us.arc")
	if err != nil {
		t.Fatalf("read first_us.arc: %v", err)
	}
	want, err := Size(raw)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decompress(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != want {
		t.Fatalf("size %d, want %d", len(out), want)
	}
	if string(out[:4]) != "GARC" {
		t.Errorf("magic %q, want GARC", out[:4])
	}
	// version float 1.0 little-endian at +4
	if out[4] != 0 || out[5] != 0 || out[6] != 0x80 || out[7] != 0x3F {
		t.Errorf("version bytes % X, want 00 00 80 3F", out[4:8])
	}
	// first chunk offset at +0x14 names the FILE directory
	off := int(uint32(out[0x14]) | uint32(out[0x15])<<8 | uint32(out[0x16])<<16 | uint32(out[0x17])<<24)
	if off+4 > len(out) || string(out[off:off+4]) != "FILE" {
		t.Errorf("no FILE chunk at +0x%X", off)
	}
}
