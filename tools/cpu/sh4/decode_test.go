package sh4

import "testing"

// TestDecodeGolden pins the rendering of at least one member of every group,
// every Flow classification, both literal-pool forms, and the .hword
// fallback. The first twelve entries are the opening of Crazy Taxi's
// 1ST_READ.BIN — the bootstrap copy loop the writeup's Part II walks through —
// so the table starts from halfwords known to be real compiler output rather
// than hand-picked encodings.
func TestDecodeGolden(t *testing.T) {
	cases := []struct {
		h    uint16
		addr uint32
		text string
		flow Flow
		// optional expectations; target checked when hasTarget, lit when litSize != 0
		target    uint32
		hasTarget bool
		delay     bool
		litAddr   uint32
		litSize   int
	}{
		// Crazy Taxi 1ST_READ.BIN, first twelve halfwords at 8C010000.
		{h: 0xD006, addr: 0x8C010000, text: "mov.l @(0x18,pc), r0", flow: FlowSeq, litAddr: 0x8C01001C, litSize: 4},
		{h: 0xD107, addr: 0x8C010002, text: "mov.l @(0x1C,pc), r1", flow: FlowSeq, litAddr: 0x8C010020, litSize: 4},
		{h: 0xD207, addr: 0x8C010004, text: "mov.l @(0x1C,pc), r2", flow: FlowSeq, litAddr: 0x8C010024, litSize: 4},
		{h: 0x6302, addr: 0x8C010006, text: "mov.l @r0, r3", flow: FlowSeq},
		{h: 0x2232, addr: 0x8C010008, text: "mov.l r3, @r2", flow: FlowSeq},
		{h: 0x7004, addr: 0x8C01000A, text: "add #4, r0", flow: FlowSeq},
		{h: 0x7204, addr: 0x8C01000C, text: "add #4, r2", flow: FlowSeq},
		{h: 0x3100, addr: 0x8C01000E, text: "cmp/eq r0, r1", flow: FlowSeq},
		{h: 0x8BF9, addr: 0x8C010010, text: "bf 0x8C010006", flow: FlowBranch, target: 0x8C010006, hasTarget: true},
		{h: 0xD004, addr: 0x8C010012, text: "mov.l @(0x10,pc), r0", flow: FlowSeq, litAddr: 0x8C010024, litSize: 4},
		{h: 0x402B, addr: 0x8C010014, text: "jmp @r0", flow: FlowIndJump, delay: true},
		{h: 0x0009, addr: 0x8C010016, text: "nop", flow: FlowSeq},

		// Group 0: system singles, sts/stc, @(R0,...) moves, control transfers.
		{h: 0x0008, addr: 0, text: "clrt", flow: FlowSeq},
		{h: 0x0018, addr: 0, text: "sett", flow: FlowSeq},
		{h: 0x0028, addr: 0, text: "clrmac", flow: FlowSeq},
		{h: 0x0038, addr: 0, text: "ldtlb", flow: FlowSeq},
		{h: 0x0019, addr: 0, text: "div0u", flow: FlowSeq},
		{h: 0x0329, addr: 0, text: "movt r3", flow: FlowSeq},
		{h: 0x000B, addr: 0, text: "rts", flow: FlowReturn, delay: true},
		{h: 0x002B, addr: 0, text: "rte", flow: FlowStop, delay: true},
		{h: 0x001B, addr: 0, text: "sleep", flow: FlowStop},
		{h: 0x0502, addr: 0, text: "stc sr, r5", flow: FlowSeq},
		{h: 0x0522, addr: 0, text: "stc vbr, r5", flow: FlowSeq},
		{h: 0x0532, addr: 0, text: "stc ssr, r5", flow: FlowSeq},
		{h: 0x05A2, addr: 0, text: "stc r2_bank, r5", flow: FlowSeq},
		{h: 0x050A, addr: 0, text: "sts mach, r5", flow: FlowSeq},
		{h: 0x052A, addr: 0, text: "sts pr, r5", flow: FlowSeq},
		{h: 0x053A, addr: 0, text: "stc sgr, r5", flow: FlowSeq},
		{h: 0x05FA, addr: 0, text: "stc dbr, r5", flow: FlowSeq},
		{h: 0x0083, addr: 0, text: "pref @r0", flow: FlowSeq},
		{h: 0x00C3, addr: 0, text: "movca.l r0, @r0", flow: FlowSeq},
		{h: 0x0403, addr: 0, text: "bsrf r4", flow: FlowIndCall, delay: true},
		{h: 0x0423, addr: 0, text: "braf r4", flow: FlowIndJump, delay: true},
		{h: 0x0154, addr: 0, text: "mov.b r5, @(r0,r1)", flow: FlowSeq},
		{h: 0x015E, addr: 0, text: "mov.l @(r0,r5), r1", flow: FlowSeq},
		{h: 0x0157, addr: 0, text: "mul.l r5, r1", flow: FlowSeq},
		{h: 0x015F, addr: 0, text: "mac.l @r5+, @r1+", flow: FlowSeq},

		// Groups 1/5: displacement longword moves.
		{h: 0x1F23, addr: 0, text: "mov.l r2, @(0xC,r15)", flow: FlowSeq},
		{h: 0x5F23, addr: 0, text: "mov.l @(0xC,r2), r15", flow: FlowSeq},

		// Group 2: stores and two-register logic.
		{h: 0x2270, addr: 0, text: "mov.b r7, @r2", flow: FlowSeq},
		{h: 0x2276, addr: 0, text: "mov.l r7, @-r2", flow: FlowSeq},
		{h: 0x2277, addr: 0, text: "div0s r7, r2", flow: FlowSeq},
		{h: 0x2278, addr: 0, text: "tst r7, r2", flow: FlowSeq},
		{h: 0x227C, addr: 0, text: "cmp/str r7, r2", flow: FlowSeq},
		{h: 0x227E, addr: 0, text: "mulu.w r7, r2", flow: FlowSeq},

		// Group 3: compares and arithmetic.
		{h: 0x3123, addr: 0, text: "cmp/ge r2, r1", flow: FlowSeq},
		{h: 0x3126, addr: 0, text: "cmp/hi r2, r1", flow: FlowSeq},
		{h: 0x3124, addr: 0, text: "div1 r2, r1", flow: FlowSeq},
		{h: 0x3128, addr: 0, text: "sub r2, r1", flow: FlowSeq},
		{h: 0x312A, addr: 0, text: "subc r2, r1", flow: FlowSeq},
		{h: 0x312C, addr: 0, text: "add r2, r1", flow: FlowSeq},
		{h: 0x312D, addr: 0, text: "dmuls.l r2, r1", flow: FlowSeq},
		{h: 0x312E, addr: 0, text: "addc r2, r1", flow: FlowSeq},

		// Group 4: shifts, system loads/stores, banked forms, jsr/jmp.
		{h: 0x4400, addr: 0, text: "shll r4", flow: FlowSeq},
		{h: 0x4401, addr: 0, text: "shlr r4", flow: FlowSeq},
		{h: 0x4408, addr: 0, text: "shll2 r4", flow: FlowSeq},
		{h: 0x4428, addr: 0, text: "shll16 r4", flow: FlowSeq},
		{h: 0x4410, addr: 0, text: "dt r4", flow: FlowSeq},
		{h: 0x4411, addr: 0, text: "cmp/pz r4", flow: FlowSeq},
		{h: 0x4415, addr: 0, text: "cmp/pl r4", flow: FlowSeq},
		{h: 0x441B, addr: 0, text: "tas.b @r4", flow: FlowSeq},
		{h: 0x4424, addr: 0, text: "rotcl r4", flow: FlowSeq},
		{h: 0x440B, addr: 0, text: "jsr @r4", flow: FlowIndCall, delay: true},
		{h: 0x442B, addr: 0, text: "jmp @r4", flow: FlowIndJump, delay: true},
		{h: 0x440E, addr: 0, text: "ldc r4, sr", flow: FlowSeq},
		{h: 0x443E, addr: 0, text: "ldc r4, ssr", flow: FlowSeq},
		{h: 0x444E, addr: 0, text: "ldc r4, spc", flow: FlowSeq},
		{h: 0x4F22, addr: 0, text: "sts.l pr, @-r15", flow: FlowSeq},
		{h: 0x4F26, addr: 0, text: "lds.l @r15+, pr", flow: FlowSeq},
		{h: 0x4F52, addr: 0, text: "sts.l fpul, @-r15", flow: FlowSeq},
		{h: 0x4F66, addr: 0, text: "lds.l @r15+, fpscr", flow: FlowSeq},
		{h: 0x456A, addr: 0, text: "lds r5, fpscr", flow: FlowSeq},
		{h: 0x4FB3, addr: 0, text: "stc.l r3_bank, @-r15", flow: FlowSeq},
		{h: 0x45A7, addr: 0, text: "ldc.l @r5+, r2_bank", flow: FlowSeq},
		{h: 0x45AE, addr: 0, text: "ldc r5, r2_bank", flow: FlowSeq},
		{h: 0x412C, addr: 0, text: "shad r2, r1", flow: FlowSeq},
		{h: 0x412D, addr: 0, text: "shld r2, r1", flow: FlowSeq},
		{h: 0x412F, addr: 0, text: "mac.w @r2+, @r1+", flow: FlowSeq},

		// Group 6: loads and transforms.
		{h: 0x6274, addr: 0, text: "mov.b @r7+, r2", flow: FlowSeq},
		{h: 0x6273, addr: 0, text: "mov r7, r2", flow: FlowSeq},
		{h: 0x6277, addr: 0, text: "not r7, r2", flow: FlowSeq},
		{h: 0x6278, addr: 0, text: "swap.b r7, r2", flow: FlowSeq},
		{h: 0x627B, addr: 0, text: "neg r7, r2", flow: FlowSeq},
		{h: 0x627C, addr: 0, text: "extu.b r7, r2", flow: FlowSeq},
		{h: 0x627F, addr: 0, text: "exts.w r7, r2", flow: FlowSeq},

		// Group 8: short displacement moves and all four conditional branches.
		{h: 0x8153, addr: 0, text: "mov.w r0, @(0x6,r5)", flow: FlowSeq},
		{h: 0x8453, addr: 0, text: "mov.b @(0x3,r5), r0", flow: FlowSeq},
		{h: 0x88FE, addr: 0, text: "cmp/eq #-2, r0", flow: FlowSeq},
		{h: 0x8905, addr: 0x100, text: "bt 0x0000010E", flow: FlowBranch, target: 0x10E, hasTarget: true},
		{h: 0x8B05, addr: 0x100, text: "bf 0x0000010E", flow: FlowBranch, target: 0x10E, hasTarget: true},
		{h: 0x8D05, addr: 0x100, text: "bt/s 0x0000010E", flow: FlowBranch, target: 0x10E, hasTarget: true, delay: true},
		{h: 0x8FFC, addr: 0x100, text: "bf/s 0x000000FC", flow: FlowBranch, target: 0xFC, hasTarget: true, delay: true},

		// Group 9: the halfword literal pool.
		{h: 0x9105, addr: 0x100, text: "mov.w @(0xA,pc), r1", flow: FlowSeq, litAddr: 0x10E, litSize: 2},

		// Groups A/B: the 12-bit displacement transfers.
		{h: 0xA003, addr: 0x2000, text: "bra 0x0000200A", flow: FlowJump, target: 0x200A, hasTarget: true, delay: true},
		{h: 0xAFFE, addr: 0x2000, text: "bra 0x00002000", flow: FlowJump, target: 0x2000, hasTarget: true, delay: true},
		{h: 0xB005, addr: 0x2000, text: "bsr 0x0000200E", flow: FlowCall, target: 0x200E, hasTarget: true, delay: true},

		// Group C: GBR moves, trapa, mova, immediate logic.
		{h: 0xC007, addr: 0, text: "mov.b r0, @(0x7,gbr)", flow: FlowSeq},
		{h: 0xC207, addr: 0, text: "mov.l r0, @(0x1C,gbr)", flow: FlowSeq},
		{h: 0xC607, addr: 0, text: "mov.l @(0x1C,gbr), r0", flow: FlowSeq},
		{h: 0xC320, addr: 0, text: "trapa #32", flow: FlowStop},
		{h: 0xC706, addr: 0x102, text: "mova @(0x18,pc), r0", flow: FlowSeq, litAddr: 0x11C, litSize: 4},
		{h: 0xC880, addr: 0, text: "tst #128, r0", flow: FlowSeq},
		{h: 0xC901, addr: 0, text: "and #1, r0", flow: FlowSeq},
		{h: 0xCB80, addr: 0, text: "or #128, r0", flow: FlowSeq},
		{h: 0xCD01, addr: 0, text: "and.b #1, @(r0,gbr)", flow: FlowSeq},

		// Group E: immediate move.
		{h: 0xE2FF, addr: 0, text: "mov #-1, r2", flow: FlowSeq},

		// Group F: the FPU space, single-precision reading.
		{h: 0xF210, addr: 0, text: "fadd fr1, fr2", flow: FlowSeq},
		{h: 0xF211, addr: 0, text: "fsub fr1, fr2", flow: FlowSeq},
		{h: 0xF212, addr: 0, text: "fmul fr1, fr2", flow: FlowSeq},
		{h: 0xF213, addr: 0, text: "fdiv fr1, fr2", flow: FlowSeq},
		{h: 0xF214, addr: 0, text: "fcmp/eq fr1, fr2", flow: FlowSeq},
		{h: 0xF215, addr: 0, text: "fcmp/gt fr1, fr2", flow: FlowSeq},
		{h: 0xF216, addr: 0, text: "fmov.s @(r0,r1), fr2", flow: FlowSeq},
		{h: 0xF217, addr: 0, text: "fmov.s fr1, @(r0,r2)", flow: FlowSeq},
		{h: 0xF218, addr: 0, text: "fmov.s @r1, fr2", flow: FlowSeq},
		{h: 0xF219, addr: 0, text: "fmov.s @r1+, fr2", flow: FlowSeq},
		{h: 0xF21A, addr: 0, text: "fmov.s fr1, @r2", flow: FlowSeq},
		{h: 0xF21B, addr: 0, text: "fmov.s fr1, @-r2", flow: FlowSeq},
		{h: 0xF21C, addr: 0, text: "fmov fr1, fr2", flow: FlowSeq},
		{h: 0xF21E, addr: 0, text: "fmac fr0, fr1, fr2", flow: FlowSeq},
		{h: 0xF20D, addr: 0, text: "fsts fpul, fr2", flow: FlowSeq},
		{h: 0xF21D, addr: 0, text: "flds fr2, fpul", flow: FlowSeq},
		{h: 0xF22D, addr: 0, text: "float fpul, fr2", flow: FlowSeq},
		{h: 0xF23D, addr: 0, text: "ftrc fr2, fpul", flow: FlowSeq},
		{h: 0xF24D, addr: 0, text: "fneg fr2", flow: FlowSeq},
		{h: 0xF25D, addr: 0, text: "fabs fr2", flow: FlowSeq},
		{h: 0xF26D, addr: 0, text: "fsqrt fr2", flow: FlowSeq},
		{h: 0xF27D, addr: 0, text: "fsrra fr2", flow: FlowSeq},
		{h: 0xF28D, addr: 0, text: "fldi0 fr2", flow: FlowSeq},
		{h: 0xF29D, addr: 0, text: "fldi1 fr2", flow: FlowSeq},
		{h: 0xF2AD, addr: 0, text: "fcnvsd fpul, dr2", flow: FlowSeq},
		{h: 0xF2BD, addr: 0, text: "fcnvds dr2, fpul", flow: FlowSeq},
		{h: 0xF6ED, addr: 0, text: "fipr fv8, fv4", flow: FlowSeq},
		{h: 0xF0FD, addr: 0, text: "fsca fpul, dr0", flow: FlowSeq},
		{h: 0xF6FD, addr: 0, text: "fsca fpul, dr6", flow: FlowSeq},
		{h: 0xF5FD, addr: 0, text: "ftrv xmtrx, fv4", flow: FlowSeq},
		{h: 0xF9FD, addr: 0, text: "ftrv xmtrx, fv8", flow: FlowSeq},
		{h: 0xFBFD, addr: 0, text: "frchg", flow: FlowSeq},
		{h: 0xF3FD, addr: 0, text: "fschg", flow: FlowSeq},

		// Holes decode as data and stop a traced path.
		{h: 0x0000, addr: 0, text: ".hword 0x0000", flow: FlowStop},
		{h: 0x2233, addr: 0, text: ".hword 0x2233", flow: FlowStop},
		{h: 0x3121, addr: 0, text: ".hword 0x3121", flow: FlowStop},
		{h: 0x4414, addr: 0, text: ".hword 0x4414", flow: FlowStop},
		{h: 0xF7FD, addr: 0, text: ".hword 0xF7FD", flow: FlowStop},
		{h: 0xFFFF, addr: 0, text: ".hword 0xFFFF", flow: FlowStop},
	}
	for _, c := range cases {
		in := DecodeHalfword(c.h, c.addr)
		if in.Text != c.text {
			t.Errorf("%04X @%08X: text %q, want %q", c.h, c.addr, in.Text, c.text)
		}
		if in.Flow != c.flow {
			t.Errorf("%04X @%08X: flow %v, want %v", c.h, c.addr, in.Flow, c.flow)
		}
		if in.HasTarget != c.hasTarget || (c.hasTarget && in.Target != c.target) {
			t.Errorf("%04X @%08X: target %v/%08X, want %v/%08X", c.h, c.addr, in.HasTarget, in.Target, c.hasTarget, c.target)
		}
		if in.HasDelay != c.delay {
			t.Errorf("%04X @%08X: delay %v, want %v", c.h, c.addr, in.HasDelay, c.delay)
		}
		if in.LitSize != c.litSize || (c.litSize != 0 && in.LitAddr != c.litAddr) {
			t.Errorf("%04X @%08X: lit %d/%08X, want %d/%08X", c.h, c.addr, in.LitSize, in.LitAddr, c.litSize, c.litAddr)
		}
	}
}

// TestDecodeTruncated pins the short-slice path.
func TestDecodeTruncated(t *testing.T) {
	in := Decode([]byte{0x09}, 0x100)
	if in.Len != 0 || in.Flow != FlowStop {
		t.Errorf("truncated decode: Len %d Flow %v, want 0/stop", in.Len, in.Flow)
	}
}

// TestDecodeLittleEndian pins the byte order once: 0x402B stored 2B 40 is
// jmp @r0.
func TestDecodeLittleEndian(t *testing.T) {
	in := Decode([]byte{0x2B, 0x40}, 0)
	if in.Text != "jmp @r0" {
		t.Errorf("LE decode: %q, want %q", in.Text, "jmp @r0")
	}
}
