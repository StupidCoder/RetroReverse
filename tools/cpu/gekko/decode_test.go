package gekko

import (
	"testing"
)

func TestDecodeText(t *testing.T) {
	cases := []struct {
		w    uint32
		addr uint32
		want string
	}{
		// The apploader's own opening words, off the disc. If these three do not decode,
		// nothing downstream can be trusted.
		{0x7C0802A6, 0x81200194, "mfspr r0,LR"},
		{0x90010004, 0x81200198, "stw r0,4(r1)"},
		{0x9421FFE0, 0x8120019C, "stwu r1,-32(r1)"},

		// Integer arithmetic and the suffix machinery.
		{0x7C221A14, 0, "add r1,r2,r3"},
		{0x7C221A15, 0, "add. r1,r2,r3"},
		{0x7C221E14, 0, "addo r1,r2,r3"},
		{0x7C221E15, 0, "addo. r1,r2,r3"},
		{0x38600000, 0, "li r3,0"},
		{0x3C60800A, 0, "lis r3,0x800A"},
		{0x38630010, 0, "addi r3,r3,16"},
		{0x60000000, 0, "nop"},
		{0x7C641B78, 0, "mr r4,r3"},
		{0x7C0903A6, 0, "mtspr CTR,r0"},

		// srawi, whose carry rule is the classic PowerPC trap — see the vector suite.
		{0x7C640E70, 0, "srawi r4,r3,1"},

		// rlwinm, and the extended mnemonics that make it readable.
		{0x5463103A, 0, "slwi r3,r3,2"},
		{0x5463E8FE, 0, "srwi r3,r3,3"},

		// Branches: relative, absolute, linked, and the CR-field conditionals.
		{0x48000010, 0x80003100, "b 0x80003110"},
		{0x48000011, 0x80003100, "bl 0x80003110"},
		{0x4E800020, 0, "blr"},
		{0x4E800420, 0, "bctr"},
		{0x4E800421, 0, "bctrl"},
		{0x41820010, 0x80003100, "beq 0x80003110"},
		{0x40820010, 0x80003100, "bne 0x80003110"},
		{0x4181000C, 0x80003100, "bgt 0x8000310C"},
		{0x41980010, 0x80003100, "blt cr6,0x80003110"}, // BI = 6*4+0: field 6, bit lt
		{0x42000010, 0x80003100, "bdnz 0x80003110"},

		// Compares.
		{0x2C030000, 0, "cmpwi cr0,r3,0"},
		{0x7C032000, 0, "cmpw cr0,r3,r4"},
		{0x7C032040, 0, "cmplw cr0,r3,r4"},

		// Floating point, single and double, and the four-operand accumulate.
		{0xFC22182A, 0, "fadd f1,f2,f3"},
		{0xEC22182A, 0, "fadds f1,f2,f3"},
		{0xFC2200F2, 0, "fmul f1,f2,f3"},
		{0xFC2320FA, 0, "fmadd f1,f3,f3,f4"},
		{0xFC000090, 0, "fmr f0,f0"},

		// The Gekko's own: paired singles, and the quantised load/store.
		{0x1022182A, 0, ""}, // placeholder, replaced below — see TestDecodePairedSingle
	}
	for _, c := range cases {
		if c.want == "" {
			continue
		}
		got := DecodeWord(c.w, c.addr)
		if got.Text != c.want {
			t.Errorf("DecodeWord(%08X) = %q, want %q", c.w, got.Text, c.want)
		}
	}
}

func TestDecodePairedSingle(t *testing.T) {
	cases := []struct {
		w    uint32
		want string
	}{
		{0x1022182A, "ps_add f1,f2,f3"}, // opcode 4, XO5 = 21
		{0x10221828, "ps_sub f1,f2,f3"}, // XO5 = 20
		{0x102200F2, "ps_mul f1,f2,f3"}, // XO5 = 25
		{0x1022193A, "ps_madd f1,f2,f4,f3"},
		{0x10220420, "ps_merge00 f1,f2,f0"},
		{0x10000090, "ps_mr f0,f0"},  // XO10 = 72
		{0x10000210, "ps_abs f0,f0"}, // XO10 = 264
		// The quantised loads: the GQR is named, the format is not — it is not here.
		{0xE0030008, "psq_l f0,8(r3),0,gqr0"},
		{0xF0030008, "psq_st f0,8(r3),0,gqr0"},
		{0xE0039008, "psq_l f0,8(r3),1,gqr1"},
	}
	for _, c := range cases {
		if got := DecodeWord(c.w, 0); got.Text != c.want {
			t.Errorf("DecodeWord(%08X) = %q, want %q", c.w, got.Text, c.want)
		}
	}
}

// The decoder tries the narrow extended-opcode field before the wide one in opcodes 4
// and 63, which is only sound if their value sets are disjoint modulo 32. That is
// asserted here rather than trusted: if a future addition to either table breaks it,
// this fails loudly instead of silently mis-decoding one instruction as another.
func TestExtendedOpcodesDoNotCollide(t *testing.T) {
	// Opcode 4: the A-forms are keyed by XO5, everything else by XO10 or XO6.
	for xo := range psX {
		if name, clash := psAForm[xo&0x1F]; clash {
			t.Errorf("opcode 4: the ten-bit form %d collides with the A-form %s (both are %d mod 32)",
				xo, name, xo&0x1F)
		}
	}
	for _, xo := range []uint32{6, 7, 38, 39} { // the psq_*x forms, by XO6
		if name, clash := psAForm[xo&0x1F]; clash {
			t.Errorf("opcode 4: the six-bit form %d collides with the A-form %s", xo, name)
		}
	}

	// Opcode 63: the A-forms are 18, 20-23, 25, 26, 28-31; the ten-bit forms must all
	// miss that set.
	aForm63 := map[uint32]bool{18: true, 20: true, 21: true, 22: true, 23: true,
		25: true, 26: true, 28: true, 29: true, 30: true, 31: true}
	tenBit63 := []uint32{0, 12, 14, 15, 32, 38, 40, 64, 70, 72, 134, 136, 264, 583, 711}
	for _, xo := range tenBit63 {
		if aForm63[xo&0x1F] {
			t.Errorf("opcode 63: the ten-bit form %d collides with an A-form (%d mod 32)", xo, xo&0x1F)
		}
	}
}

func TestDecodeFlow(t *testing.T) {
	cases := []struct {
		w    uint32
		addr uint32
		flow Flow
		tgt  uint32
	}{
		{0x48000010, 0x80003100, FlowJump, 0x80003110},   // b
		{0x48000011, 0x80003100, FlowCall, 0x80003110},   // bl
		{0x41820010, 0x80003100, FlowBranch, 0x80003110}, // beq: continues too
		{0x4E800020, 0, FlowReturn, 0},                   // blr
		{0x4E800420, 0, FlowIndJump, 0},                  // bctr
		{0x4E800421, 0, FlowIndCall, 0},                  // bctrl
		{0x4C820020, 0, FlowBranch, 0},                   // a CONDITIONAL blr (bnelr): it may also fall through
		{0x44000002, 0, FlowStop, 0},                     // sc
		{0x4C000064, 0, FlowStop, 0},                     // rfi
		{0x7C221A14, 0, FlowSeq, 0},                      // add
		{0x00000000, 0, FlowStop, 0},                     // an illegal word stops a traced path
	}
	for _, c := range cases {
		got := DecodeWord(c.w, c.addr)
		if got.Flow != c.flow {
			t.Errorf("DecodeWord(%08X).Flow = %v, want %v (%s)", c.w, got.Flow, c.flow, got.Text)
		}
		if c.tgt != 0 && got.Target != c.tgt {
			t.Errorf("DecodeWord(%08X).Target = %08X, want %08X", c.w, got.Target, c.tgt)
		}
	}
}

// `bl .+4` is the position-independent-code idiom for reading the program counter, not
// a call. Classifying it as one would invent a function at every occurrence.
func TestBranchAndLinkToNextIsNotACall(t *testing.T) {
	in := DecodeWord(0x48000005, 0x80003100) // bl 0x80003104
	if in.Flow != FlowSeq {
		t.Errorf("bl .+4 classified as %v, want seq", in.Flow)
	}
}

// PowerPC is big-endian, and the decoder must read it that way. The same four bytes read
// the other way round are a different instruction entirely.
func TestDecodeIsBigEndian(t *testing.T) {
	in := Decode([]byte{0x7C, 0x08, 0x02, 0xA6}, 0)
	if in.Mnem != "mfspr" {
		t.Errorf("decoded %q, want mfspr — the bytes were read little-endian", in.Text)
	}
}

func TestDecodeShortSlice(t *testing.T) {
	in := Decode([]byte{0x7C, 0x08}, 0)
	if in.Len != 0 || in.Flow != FlowStop {
		t.Errorf("a two-byte slice decoded to len %d flow %v, want 0/stop", in.Len, in.Flow)
	}
}
