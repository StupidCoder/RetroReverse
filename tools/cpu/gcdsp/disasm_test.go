package gcdsp

import "testing"

// reader turns a word slice into the read closure Disasm wants.
func reader(words []uint16) func(uint16) uint16 {
	return func(a uint16) uint16 {
		if int(a) < len(words) {
			return words[a]
		}
		return 0
	}
}

// TestDecode pins the encodings confirmed against the real Luigi's Mansion AX microcode: the
// reset/interrupt vector forms, the mode-setup ops, the register-init idioms, and the two-word
// loads and branches. Each entry is the word (or word pair) and the text and length expected.
func TestDecode(t *testing.T) {
	cases := []struct {
		words []uint16
		want  string
		span  uint16
	}{
		{[]uint16{0x0000}, "nop", 1},
		{[]uint16{0x0021}, "halt", 1},
		{[]uint16{0x029F, 0x0010}, "jmp    0x0010", 2},
		{[]uint16{0x02FF}, "rti", 1},
		{[]uint16{0x02DF}, "ret", 1},
		{[]uint16{0x02BF, 0x005A}, "call   0x005A", 2},
		{[]uint16{0x029C, 0x002C}, "jmplnz 0x002C", 2},
		{[]uint16{0x1302}, "sbset  #2", 1},
		{[]uint16{0x1204}, "sbclr  #4", 1},
		{[]uint16{0x8E00}, "set16", 1},
		{[]uint16{0x8C00}, "clr15", 1},
		{[]uint16{0x8B00}, "m0", 1},
		{[]uint16{0x8100}, "clr    ac0", 1},
		{[]uint16{0x8900}, "clr    ac1", 1},
		{[]uint16{0x009E, 0xFFFF}, "lri    ac0.m, #0xFFFF", 2},
		{[]uint16{0x1D1E}, "mrr    wr0, ac0.m", 1},
		{[]uint16{0x00DC, 0x0343}, "lr     ac0.l, @0x0343", 2},
		{[]uint16{0x00FE, 0x0343}, "sr     @0x0343, ac0.m", 2},
		{[]uint16{0x16FC, 0x8888}, "si     @0xFC, #0x8888", 2},
		{[]uint16{0x005F}, "loop   ac1.m", 1},
		// The register-indirect load/store family: bit 9 = store, bits 8..7 = post-modify.
		{[]uint16{0x1B1E}, "srri   @ar0, ac0.m", 1}, // store, post-increment (the DRAM-clear idiom)
		{[]uint16{0x1A5E}, "srr    @ar2, ac0.m", 1}, // store, no modify
		{[]uint16{0x1ABC}, "srrd   @ar1, ac0.l", 1}, // store, post-decrement
		{[]uint16{0x193E}, "lrri   ac0.m, @ar1", 1}, // load, post-increment
		{[]uint16{0x199E}, "lrrn   ac0.m, @ar0", 1}, // load, post-add index
		{[]uint16{0x181E}, "lrr    ac0.m, @ar0", 1}, // load, no modify
		{[]uint16{0x26FF}, "lrs    ac0.m, @0xFFFF", 1},
		{[]uint16{0x2800}, "srs    @0xFF00, ax0.l", 1},
		{[]uint16{0x0200, 0x0062}, "addi   ac0, #0x0062", 2},
		{[]uint16{0x0240, 0x007E}, "andi   ac0.m, #0x007E", 2},
		{[]uint16{0x170F}, "jmpr   ar0", 1},
		// The register-indirect control family: bits 7..5 register, bit 4 = call flag, cond 0xF.
		{[]uint16{0x173F}, "callr  ar1", 1}, // the callr at ucode 0x04E0
		{[]uint16{0x175F}, "callr  ar2", 1}, // the jump-table dispatch at ucode 0x0276
		{[]uint16{0x174F}, "jmpr   ar2", 1},
		// The shift group: bit 8 = accumulator, bit 7 = arithmetic, bits 6..0 signed amount.
		{[]uint16{0x1401}, "lsl    ac0, #1", 1},
		{[]uint16{0x1479}, "lsr    ac0, #7", 1}, // -7 in the 7-bit field
		{[]uint16{0x147F}, "lsr    ac0, #1", 1},
		{[]uint16{0x1488}, "asl    ac0, #8", 1},
		{[]uint16{0x14FB}, "asr    ac0, #5", 1}, // arithmetic, -5
		{[]uint16{0x1501}, "lsl    ac1, #1", 1},
		{[]uint16{0x1570}, "lsr    ac1, #16", 1}, // -16
		{[]uint16{0x1585}, "asl    ac1, #5", 1},
	}
	for _, c := range cases {
		got, span := Disasm(reader(c.words), 0)
		if got != c.want || span != c.span {
			t.Errorf("word %#04x: got %q (%d words), want %q (%d words)", c.words[0], got, span, c.want, c.span)
		}
	}
}

// TestAlignment builds a small program shaped like a real ucode — a reset jump over an
// interrupt-vector table, a body with a two-word load and a block loop — and checks the
// validator finds every branch target on an instruction boundary and the stream tiling the
// whole image. This is the property that proved the decoder's lengths correct against the real
// 3056-word microcode.
func TestAlignment(t *testing.T) {
	prog := []uint16{
		0x029F, 0x0008, // 0000: jmp 0x0008 (reset, over the vectors)
		0x02FF, 0x0000, // 0002: rti / nop
		0x02FF, 0x0000, // 0004: rti / nop
		0x02FF, 0x0000, // 0006: rti / nop
		0x8E00,         // 0008: set16
		0x009E, 0xFFFF, // 0009: lri ac0.m,#0xFFFF (two words)
		0x1100, 0x0010, // 000B: bloopi #0, 0x0010
		0x1B1E,         // 000D: srr @ar0, ac0.m
		0x02BF, 0x0008, // 000E: call 0x0008
		0x02DF, // 0010: ret
	}
	mis, boundaries, overrun := DisasmValidate(reader(prog), uint16(len(prog)))
	if len(mis) != 0 {
		t.Errorf("misaligned branch targets at %v", mis)
	}
	if overrun {
		t.Error("decode overran the image")
	}
	if boundaries == 0 {
		t.Error("no instructions decoded")
	}
}
