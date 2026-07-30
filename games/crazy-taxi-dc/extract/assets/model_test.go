package assets

// Synthetic model tests pinning the stream grammar read off the game's
// renderer (the walker at 0C080E20/0C080E44): 24-byte model header, 80-byte
// block headers with a strip-stream byte size, [flags][count] strips whose
// vertices are either 32-byte inline records (LSB of x set) or 8-byte
// back-references, and the [0][total-vertex-count] trailer.

import (
	"encoding/binary"
	"math"
	"testing"
)

func fbits(f float32) uint32 { return math.Float32bits(f) }

func buildModel(t *testing.T) []byte {
	var b []byte
	w := func(v uint32) {
		var q [4]byte
		binary.LittleEndian.PutUint32(q[:], v)
		b = append(b, q[:]...)
	}
	// model header: kind 1, sub 3, sphere
	w(1)
	w(3)
	w(fbits(0))
	w(fbits(1))
	w(fbits(0))
	w(fbits(5))
	// block header (80 bytes): PCW translucent+textured, then a strip
	// stream of one 3-vertex strip and one 1-triangle list
	w(0x828C00AC)
	w(0x83000000)
	w(0x9408245B)
	w(0xC80754A4)
	for _, f := range []float32{0, 1, 0, 5} { // block sphere
		w(fbits(f))
	}
	w(0xAC)                  // aux
	w(0xFFFFFFFF)            // mode -1
	w(fbits(0.75))           // intensity
	for i := 0; i < 8; i++ { // base+offset colours
		w(fbits(1))
	}
	inline := func(x, y, z float32) {
		w(fbits(x) | 1) // the inline flag lives in x's LSB
		w(fbits(y))
		w(fbits(z))
		w(fbits(0))
		w(fbits(1))
		w(fbits(0))
		w(fbits(0.5))
		w(fbits(0.5))
	}
	// strip stream size: strip hdr 8 + 2 inline + 1 indexed (8) +
	// tri-list hdr 8 + 3 inline
	w(8 + 2*32 + 8 + 8 + 3*32)
	w(0x52) // strip flags (bit3 clear = tristrip)
	w(3)    // 3 vertices
	inline(1, 0, 0)
	inline(2, 0, 0)
	w(0x100)                    // indexed: LSB clear
	w(0xFFFFFFFF & ^uint32(71)) // -72: back to the first inline vertex
	w(0x5A)                     // flags bit3 = triangle list
	w(1)                        // one triangle
	inline(0, 1, 0)
	inline(0, 2, 0)
	inline(0, 3, 0)
	// terminator + vertex-record count trailer
	w(0)
	w(6)
	return b
}

func TestModelRoundTrip(t *testing.T) {
	data := buildModel(t)
	m, err := OpenModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if m.Kind != 1 || m.Sub != 3 {
		t.Fatalf("kind/sub = %d/%d", m.Kind, m.Sub)
	}
	if !m.Counted {
		t.Fatal("trailer count did not match")
	}
	if m.Size != len(data) {
		t.Fatalf("size %d, want %d", m.Size, len(data))
	}
	if len(m.Blocks) != 1 {
		t.Fatalf("%d blocks", len(m.Blocks))
	}
	b := m.Blocks[0]
	if b.ListType() != 2 || !b.Textured() {
		t.Fatalf("list %d textured %v", b.ListType(), b.Textured())
	}
	if len(b.Strips) != 2 {
		t.Fatalf("%d strips", len(b.Strips))
	}
	st, tl := b.Strips[0], b.Strips[1]
	if len(st.Verts) != 3 || len(tl.Verts) != 3 {
		t.Fatalf("verts %d/%d", len(st.Verts), len(tl.Verts))
	}
	if !st.Verts[2].Indexed {
		t.Fatal("third strip vertex should be a back-reference")
	}
	// the back-reference points at the first inline vertex
	if got := st.Verts[2].Pos; math.Abs(float64(got[0]-1)) > 1e-5 {
		t.Fatalf("indexed vertex pos %v", got)
	}
	if n := st.Verts[0].Normal; n != [3]float32{0, 1, 0} {
		t.Fatalf("normal %v", n)
	}
	if got, want := len(st.Tris()), 1; got != want {
		t.Fatalf("strip tris %d, want %d", got, want)
	}
	if got, want := len(tl.Tris()), 1; got != want {
		t.Fatalf("list tris %d, want %d", got, want)
	}
}

func TestModelRejectsBadStream(t *testing.T) {
	data := buildModel(t)
	// corrupt the block's size so the strip walk misses its end
	binary.LittleEndian.PutUint32(data[100:], 12345)
	if _, err := OpenModel(data); err == nil {
		t.Fatal("bad block size not rejected")
	}
}
