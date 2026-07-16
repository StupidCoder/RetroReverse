package ps2

import (
	"encoding/binary"
	"testing"
)

// TestUnpackSkipMode pins STCYCL's skip-write addressing: with CL > WL the unit
// writes WL vectors then skips ahead to the next CL boundary, which is how the
// game interleaves separate position/texcoord/colour streams into one vertex
// array (the title streams three cl3/wl1 unpacks per batch). A contiguous write
// scrambles every record — the bug that drew the title's text mesh as garbage.
func TestUnpackSkipMode(t *testing.T) {
	m := NewMachine()
	v := m.ensureVIF(1)

	code := func(cmd, num, imm uint32) []byte {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], imm&0xFFFF|num<<16|cmd<<24)
		return b[:]
	}
	qw := func(a, b, c, d uint32) []byte {
		var buf [16]byte
		binary.LittleEndian.PutUint32(buf[0:], a)
		binary.LittleEndian.PutUint32(buf[4:], b)
		binary.LittleEndian.PutUint32(buf[8:], c)
		binary.LittleEndian.PutUint32(buf[12:], d)
		return buf[:]
	}

	// STCYCL CL=3 WL=1, then three V4-32 unpacks of 2 vectors each at VU
	// addresses 0, 1, 2 — three streams interleaving into two 3-qw records.
	var stream []byte
	stream = append(stream, code(0x01, 0, 0x0103)...) // STCYCL cl=3 wl=1
	stream = append(stream, code(0x6C, 2, 0)...)      // V4-32, 2 vectors, at 0
	stream = append(stream, qw(0xA0, 0xA1, 0xA2, 0xA3)...)
	stream = append(stream, qw(0xB0, 0xB1, 0xB2, 0xB3)...)
	stream = append(stream, code(0x6C, 2, 1)...) // V4-32, 2 vectors, at 1
	stream = append(stream, qw(0xC0, 0xC1, 0xC2, 0xC3)...)
	stream = append(stream, qw(0xD0, 0xD1, 0xD2, 0xD3)...)
	stream = append(stream, code(0x6C, 2, 2)...) // V4-32, 2 vectors, at 2
	stream = append(stream, qw(0xE0, 0xE1, 0xE2, 0xE3)...)
	stream = append(stream, qw(0xF0, 0xF1, 0xF2, 0xF3)...)
	v.feed(stream)

	// The records the microprogram walks: {A,C,E} then {B,D,F}, 3 qw apart.
	want := []uint32{0xA0, 0xC0, 0xE0, 0xB0, 0xD0, 0xF0}
	for i, w := range want {
		got := binary.LittleEndian.Uint32(v.data[i*16:])
		if got != w {
			t.Errorf("record qw %d: got %08X, want %08X", i, got, w)
		}
	}
}
