package assets

// Synthetic texdir tests: a fake 1ST_READ image with descriptors planted at
// the real array offsets, and payloads whose decodes are hand-computable.
// The real arrays were verified against the running game (all 216 truth-table
// anchors on TEXDC0; all 336 TEXDC1 entries byte- or texel-identical to the
// loader's VRAM uploads from the drive2 state).

import (
	"encoding/binary"
	"testing"
)

// buildFirstRead plants descriptor entries at course c's array offset; the
// array's remaining slots (the parser reads the full count) get valid dummies.
func buildFirstRead(course int, entries []TexEntry) []byte {
	le := binary.LittleEndian
	out := make([]byte, 0x140000)
	for i := 0; i < texDirs[course].n; i++ {
		e := TexEntry{W: 8, H: 8, Kind: TexTwiddled}
		if i < len(entries) {
			e = entries[i]
		}
		o := texDirs[course].off + i*16
		le.PutUint16(out[o:], uint16(e.W))
		le.PutUint16(out[o+2:], uint16(e.H))
		out[o+4] = e.Fmt
		out[o+5] = e.Kind
		le.PutUint32(out[o+8:], 0x0CB40000+e.Off)
	}
	return out
}

func TestTexDirSizes(t *testing.T) {
	// The size rules the baked consecutive offsets proved (32-byte-aligned):
	cases := []struct {
		e    TexEntry
		want uint32
	}{
		{TexEntry{W: 64, H: 64, Kind: TexVQMip}, 0xD60},   // 2048 + (1+1+4+16+64+256) + 1024
		{TexEntry{W: 64, H: 64, Kind: TexVQ}, 0xC00},      // 2048 + 1024
		{TexEntry{W: 32, H: 32, Kind: TexMip}, 0xAC0},     // (341+1)*2 + 2048, aligned
		{TexEntry{W: 16, H: 16, Kind: TexMip}, 0x2C0},     // (85+1)*2 + 512, aligned
		{TexEntry{W: 8, H: 8, Kind: TexMip}, 0xC0},        // (21+1)*2 + 128, aligned
		{TexEntry{W: 32, H: 32, Kind: TexTwiddled}, 0x800},
		{TexEntry{W: 128, H: 32, Kind: TexRect}, 0x2000},
	}
	for _, c := range cases {
		if got := c.e.Size(); got != c.want {
			t.Errorf("Size(kind %#x %dx%d) = %#x, want %#x", c.e.Kind, c.e.W, c.e.H, got, c.want)
		}
	}
}

func TestVQIndexTop(t *testing.T) {
	// The per-level ladder the oracle's rasteriser climbs: 8x8 → 6, 16x16 → 0x16.
	if got := vqIndexTop(8); got != 6 {
		t.Errorf("vqIndexTop(8) = %d, want 6", got)
	}
	if got := vqIndexTop(16); got != 0x16 {
		t.Errorf("vqIndexTop(16) = %d, want 0x16", got)
	}
}

func TestTexDirDecode(t *testing.T) {
	// Course 2 (arbitrary): one twiddled, one mip, one VQ, one non-square entry.
	entries := []TexEntry{
		{W: 2, H: 2, Fmt: 1, Kind: TexTwiddled, Off: 0},
		{W: 2, H: 2, Fmt: 1, Kind: TexMip, Off: 0x20},
		{W: 2, H: 2, Fmt: 1, Kind: TexVQ, Off: 0x40},
		{W: 4, H: 2, Fmt: 1, Kind: TexRect, Off: 0x8C0},
	}
	first := buildFirstRead(2, entries)
	d, err := OpenTexDir(first, 2)
	if err != nil {
		t.Fatalf("OpenTexDir: %v", err)
	}
	d.Entries = d.Entries[:len(entries)] // synthetic image holds only these

	texdc := make([]byte, 0x1000)
	le := binary.LittleEndian
	// entry 0, twiddled 2x2: twiddle order is (0,0)(0,1)(1,0)(1,1)
	// RGB565 values chosen so each texel is a distinct pure channel.
	red, green, blue, white := uint16(0xF800), uint16(0x07E0), uint16(0x001F), uint16(0xFFFF)
	le.PutUint16(texdc[0:], red)   // (0,0)
	le.PutUint16(texdc[2:], green) // (0,1)
	le.PutUint16(texdc[4:], blue)  // (1,0)
	le.PutUint16(texdc[6:], white) // (1,1)
	// entry 1, mip 2x2: file top level at ((4-1)/3+1)*2 = 4 bytes past base
	le.PutUint16(texdc[0x20+4:], red)
	le.PutUint16(texdc[0x20+6:], green)
	le.PutUint16(texdc[0x20+8:], blue)
	le.PutUint16(texdc[0x20+10:], white)
	// entry 2, VQ 2x2: one index selecting codebook entry 7, whose 4 texels
	// are red/green/blue/white in the codebook's t0..t3 order
	le.PutUint16(texdc[0x40+7*8+0:], red)   // t0 = (0,0)
	le.PutUint16(texdc[0x40+7*8+2:], green) // t1 = (0,1)
	le.PutUint16(texdc[0x40+7*8+4:], blue)  // t2 = (1,0)
	le.PutUint16(texdc[0x40+7*8+6:], white) // t3 = (1,1)
	texdc[0x40+2048] = 7
	// entry 3, non-square 4x2: two 2x2 twiddled squares side by side.
	// Square 0 texels in twiddle order (0,0)(0,1)(1,0)(1,1), then square 1
	// covering x 2-3 in the same order.
	for i, v := range []uint16{red, green, blue, white, white, blue, green, red} {
		le.PutUint16(texdc[0x8C0+2*i:], v)
	}

	wantQuad := [2][2][4]uint8{ // [y][x]
		{{248, 0, 0, 255}, {0, 0, 248, 255}},
		{{0, 252, 0, 255}, {248, 252, 248, 255}},
	}
	for i := 0; i < 3; i++ {
		img, err := d.Decode(i, texdc)
		if err != nil {
			t.Fatalf("Decode(%d): %v", i, err)
		}
		for y := 0; y < 2; y++ {
			for x := 0; x < 2; x++ {
				got := img.RGBAAt(x, y)
				w := wantQuad[y][x]
				if got.R != w[0] || got.G != w[1] || got.B != w[2] {
					t.Errorf("entry %d texel (%d,%d) = %v, want %v", i, x, y, got, w)
				}
			}
		}
	}
	img, err := d.Decode(3, texdc)
	if err != nil {
		t.Fatalf("Decode(3): %v", err)
	}
	// Square 0 is the same quad as the plain twiddled entry; square 1 puts
	// its own twiddle order at x 2-3 — (2,0) white, (3,0) green, (3,1) red.
	if g := img.RGBAAt(0, 1); g.G != 252 || g.R != 0 {
		t.Errorf("rect texel (0,1) = %v, want green", g)
	}
	if g := img.RGBAAt(2, 0); g.R != 248 || g.G != 252 || g.B != 248 {
		t.Errorf("rect texel (2,0) = %v, want white", g)
	}
	if g := img.RGBAAt(3, 0); g.G != 252 || g.R != 0 {
		t.Errorf("rect texel (3,0) = %v, want green", g)
	}
	if g := img.RGBAAt(3, 1); g.R != 248 || g.G != 0 || g.B != 0 {
		t.Errorf("rect texel (3,1) = %v, want red", g)
	}
}

func TestTexDirRefusesLies(t *testing.T) {
	if _, err := OpenTexDir(make([]byte, 16), 0); err == nil {
		t.Error("a short 1ST_READ was accepted")
	}
	if _, err := OpenTexDir(nil, 4); err == nil {
		t.Error("course 4 was accepted")
	}
	entries := []TexEntry{{W: 64, H: 64, Fmt: 1, Kind: TexVQMip, Off: 0}}
	d, err := OpenTexDir(buildFirstRead(0, entries), 0)
	if err != nil {
		t.Fatalf("OpenTexDir: %v", err)
	}
	// A payload shorter than the entry's size must error, not read past.
	if _, err := d.Decode(0, make([]byte, 16)); err == nil {
		t.Error("an overrunning texture was decoded")
	}
}
