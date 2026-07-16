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

// TestOffsetNegativeWrap pins the double-buffer registers' width: BASE, OFFSET and
// TOPS are 10 bits, and the title's merc chains lean on it — they program
// OFFSET = -BASE (0xFE46 for base 442) so the second buffer sits at
// (442 + (-442 mod 1024)) mod 1024 = 0. Kept at 16 bits, TOPS became 65536, every
// FLG unpack landed out of range and was dropped, and the skinning streamer ran
// away over another chain's leavings — the title's black screen.
func TestOffsetNegativeWrap(t *testing.T) {
	m := NewMachine()
	v := m.ensureVIF(1)

	code := func(cmd, num, imm uint32) []byte {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], imm&0xFFFF|num<<16|cmd<<24)
		return b[:]
	}

	var stream []byte
	stream = append(stream, code(0x03, 0, 442)...)    // BASE 442
	stream = append(stream, code(0x02, 0, 0xFE46)...) // OFFSET -442 (10 bits: 582)
	stream = append(stream, code(0x14, 0, 0)...)      // MSCAL 0: TOP=442, TOPS flips to (442+582)&0x3FF = 0
	// V4-32, 1 vector, at 140 with the FLG bit (1<<15): lands at TOPS+140 = 140.
	stream = append(stream, code(0x6C, 1, 140|1<<15)...)
	stream = append(stream, []byte{0xEF, 0xBE, 0xAD, 0xDE, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	v.feed(stream)

	if v.tops != 0 {
		t.Errorf("TOPS after the flip: got %d, want 0", v.tops)
	}
	if got := binary.LittleEndian.Uint32(v.data[140*16:]); got != 0xDEADBEEF {
		t.Errorf("FLG unpack at TOPS+140: got %08X at qw 140, want DEADBEEF", got)
	}
	if v.vu.Top != 442 {
		t.Errorf("TOP handed to the first MSCAL: got %d, want 442", v.vu.Top)
	}
}
