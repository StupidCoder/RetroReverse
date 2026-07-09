package n64

// The byte-order conversions are exercised on synthetic images so they run
// everywhere; the header/CIC checks need the cartridge, which is not committed
// (see the repository's image policy) and so skip when it is absent.

import (
	"encoding/binary"
	"os"
	"testing"
)

// romPath is the Pilotwings 64 image, relative to this package.
const romPath = "../../../games/pilotwings-64-n64/image/Pilotwings 64 (USA).z64"

// synthImage builds a minimal big-endian image: the magic word, an entry point,
// and IPL3-sized filler.
func synthImage() []byte {
	d := make([]byte, IPL3End)
	binary.BigEndian.PutUint32(d[0x00:], magicZ64)
	binary.BigEndian.PutUint32(d[0x08:], 0x80200050)
	for i := IPL3Start; i < IPL3End; i++ {
		d[i] = byte(i)
	}
	return d
}

func TestRomByteOrder(t *testing.T) {
	native := synthImage()

	v64 := append([]byte(nil), native...)
	swapPairs(v64)
	if got := binary.BigEndian.Uint32(v64); got != magicV64 {
		t.Fatalf("v64 magic: got 0x%08X want 0x%08X", got, magicV64)
	}

	n64 := append([]byte(nil), native...)
	reverseWords(n64)
	if got := binary.BigEndian.Uint32(n64); got != magicN64 {
		t.Fatalf("n64 magic: got 0x%08X want 0x%08X", got, magicN64)
	}

	// Each order must normalise back to the identical native image.
	for _, tc := range []struct {
		name string
		data []byte
		want ByteOrder
	}{
		{"z64", append([]byte(nil), native...), OrderZ64},
		{"v64", v64, OrderV64},
		{"n64", n64, OrderN64},
	} {
		r, err := Decode(tc.data)
		// A synthetic image has no real IPL3, so Decode reports an unknown CIC.
		// Byte order and the header are resolved before that, and are what we check.
		if err == nil {
			t.Fatalf("%s: expected an unrecognised-IPL3 error for a synthetic image", tc.name)
		}
		if r != nil {
			t.Fatalf("%s: expected a nil ROM alongside the error", tc.name)
		}
		if got := binary.BigEndian.Uint32(tc.data); got != magicZ64 {
			t.Errorf("%s: after Decode the image is not big-endian: first word 0x%08X", tc.name, got)
		}
		for i := range native {
			if tc.data[i] != native[i] {
				t.Fatalf("%s: byte %d differs after normalising: got 0x%02X want 0x%02X",
					tc.name, i, tc.data[i], native[i])
			}
		}
	}
}

func TestRomRejectsUnknownMagic(t *testing.T) {
	d := make([]byte, IPL3End)
	binary.BigEndian.PutUint32(d, 0xDEADBEEF)
	if _, err := Decode(d); err == nil {
		t.Fatal("expected an error for an image with no recognisable magic")
	}
}

func TestRomShortImage(t *testing.T) {
	if _, err := Decode(make([]byte, 16)); err == nil {
		t.Fatal("expected an error for an image shorter than header + IPL3")
	}
}

// TestRomPilotwings pins what the cartridge actually says. Every value here was
// read out of the image, not assumed: the file is stored byteswapped despite its
// .z64 extension, and the title carries a space.
func TestRomPilotwings(t *testing.T) {
	if _, err := os.Stat(romPath); os.IsNotExist(err) {
		t.Skip("Pilotwings 64 image not present (game images are not committed)")
	}
	r, err := Load(romPath)
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		what string
		got  any
		want any
	}{
		{"byte order", r.Order, OrderV64},
		{"md5", r.MD5, "c5569227242e04138aac8457b7f83e6c"},
		{"size", len(r.Data), 8 * 1024 * 1024},
		{"name", r.Header.Name, "Pilot Wings64"},
		{"entry", r.Header.Entry, uint32(0x80200050)},
		{"crc1", r.Header.CRC1, uint32(0xC851961C)},
		{"crc2", r.Header.CRC2, uint32(0x78FCAAFA)},
		{"cart type", r.Header.CartType, byte('N')},
		{"cart id", r.Header.CartID, "PW"},
		{"region", r.Header.Region(), "USA"},
		{"ipl3 crc32", r.IPL3, uint32(0x90BB6CB5)},
		{"cic", r.CIC.Name, "CIC-NUS-6102"},
		{"cic seed", r.CIC.Seed, uint32(0x3F)},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v want %v", c.what, c.got, c.want)
		}
	}

	// Normalisation must leave the cartridge bus's own view of the first word.
	if got := binary.BigEndian.Uint32(r.Data); got != magicZ64 {
		t.Errorf("normalised first word: got 0x%08X want 0x%08X", got, magicZ64)
	}
	t.Log(r)
}
