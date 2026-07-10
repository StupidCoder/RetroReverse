package threedo

import (
	"encoding/binary"
	"image/color"
	"testing"
)

// bitWriter emits big-endian (MSB-first) bit fields — the inverse of bitReader.
type bitWriter struct {
	data []byte
	nbit int
}

func (bw *bitWriter) write(v uint32, n int) {
	for i := n - 1; i >= 0; i-- {
		if bw.nbit%8 == 0 {
			bw.data = append(bw.data, 0)
		}
		if (v>>uint(i))&1 == 1 {
			bw.data[len(bw.data)-1] |= 1 << uint(7-bw.nbit%8)
		}
		bw.nbit++
	}
}
func (bw *bitWriter) alignByte() {
	for bw.nbit%8 != 0 {
		bw.nbit++
	}
	for len(bw.data)*8 < bw.nbit {
		bw.data = append(bw.data, 0)
	}
}

// a packet in an encoded line: op is one of packLiteral/packTransparent/packRepeat.
type pkt struct {
	op   int
	vals []uint32 // literal pixels, or a single repeat pixel
	n    int      // transparent/repeat run length
}

// encodeLine emits one packed scanline (offset byte + packets + EOL), padded to
// a whole number of 32-bit words, with the leading offset byte backfilled.
func encodeLine(bpp int, pkts []pkt) []byte {
	bw := &bitWriter{}
	bw.write(0, 8) // offset-byte placeholder (bpp<=8 => 1 byte)
	for _, p := range pkts {
		switch p.op {
		case packLiteral:
			bw.write(uint32(packLiteral), 2)
			bw.write(uint32(len(p.vals)-1), 6)
			for _, v := range p.vals {
				bw.write(v, bpp)
			}
		case packTransparent:
			bw.write(uint32(packTransparent), 2)
			bw.write(uint32(p.n-1), 6)
		case packRepeat:
			bw.write(uint32(packRepeat), 2)
			bw.write(uint32(p.n-1), 6)
			bw.write(p.vals[0], bpp)
		}
	}
	bw.write(packEOL, 2)
	bw.alignByte()
	for len(bw.data)%4 != 0 { // pad to word boundary
		bw.data = append(bw.data, 0)
	}
	bw.data[0] = byte(len(bw.data)/4 - 2) // offset = words - 2
	return bw.data
}

// buildCel assembles a 4bpp coded, packed .cel from a PLUT and encoded lines.
func buildCel(w, h int, plut []uint16, pdat []byte) []byte {
	be := binary.BigEndian
	chunk := func(tag string, body []byte) []byte {
		out := make([]byte, 8+len(body))
		copy(out, tag)
		be.PutUint32(out[4:], uint32(8+len(body)))
		copy(out[8:], body)
		return out
	}
	ccb := make([]byte, 0x48)
	// CCB_CCBPRE: this synthetic cel carries its preamble in the CCB struct
	// (a cel without the flag leads its source data with the preamble words,
	// and ParseCel consumes them from there).
	be.PutUint32(ccb[0x04:], ccbPacked|ccbCCBPre)
	be.PutUint32(ccb[0x38:], 3|uint32(h-1)<<6) // PRE0: bpp code 3 (4bpp), VCNT
	be.PutUint32(ccb[0x3C:], uint32(w-1))      // PRE1: HPCNT
	be.PutUint32(ccb[0x40:], uint32(w))
	be.PutUint32(ccb[0x44:], uint32(h))

	pl := make([]byte, 4+len(plut)*2)
	be.PutUint32(pl, uint32(len(plut)))
	for i, c := range plut {
		be.PutUint16(pl[4+i*2:], c)
	}

	out := chunk("CCB ", ccb)
	out = append(out, chunk("PLUT", pl)...)
	out = append(out, chunk("PDAT", pdat)...)
	return out
}

func TestCelPackedRoundTrip(t *testing.T) {
	const w, h, bpp = 6, 2, 4
	plut := []uint16{0x0000, 0x7C00, 0x03E0, 0x7FFF} // black, red, green, white

	// Row 0: a literal run then a repeat run. Row 1: transparent-bordered literal.
	pdat := encodeLine(bpp, []pkt{
		{op: packLiteral, vals: []uint32{1, 2, 3}},
		{op: packRepeat, vals: []uint32{2}, n: 3},
	})
	pdat = append(pdat, encodeLine(bpp, []pkt{
		{op: packTransparent, n: 2},
		{op: packLiteral, vals: []uint32{3, 3}},
		{op: packTransparent, n: 2},
	})...)

	cel, err := ParseCel(buildCel(w, h, plut, pdat))
	if err != nil {
		t.Fatalf("ParseCel: %v", err)
	}
	if cel.Width != w || cel.Height != h || cel.BPP != bpp || !cel.Packed || !cel.Coded {
		t.Fatalf("header: %dx%d bpp=%d packed=%v coded=%v", cel.Width, cel.Height, cel.BPP, cel.Packed, cel.Coded)
	}
	img, err := cel.Image()
	if err != nil {
		t.Fatalf("Image: %v", err)
	}

	want := [][]color.RGBA{
		{rgb555(0x7C00), rgb555(0x03E0), rgb555(0x7FFF), rgb555(0x03E0), rgb555(0x03E0), rgb555(0x03E0)},
		{{}, {}, rgb555(0x7FFF), rgb555(0x7FFF), {}, {}}, // {} = transparent (alpha 0)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if got := img.RGBAAt(x, y); got != want[y][x] {
				t.Errorf("pixel (%d,%d) = %v, want %v", x, y, got, want[y][x])
			}
		}
	}
}

func TestCelRejectsNonCel(t *testing.T) {
	if _, err := ParseCel([]byte("not a cel at all")); err == nil {
		t.Errorf("ParseCel(junk) succeeded, want error")
	}
}
