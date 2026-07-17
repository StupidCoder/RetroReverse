package x86

import (
	"math"
	"strings"
	"testing"
)

func putF32(b []byte, v float32) {
	u := math.Float32bits(v)
	b[0], b[1], b[2], b[3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
}
func getF32(b []byte) float32 {
	u := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return math.Float32frombits(u)
}

// TestSSEScalar exercises the scalar-single SSE path the Xbox C runtime leans on:
// MOVSS load/store, ADDSS, MULSS, XORPS (register zero), and the integer<->float
// conversions CVTSI2SS / CVTTSS2SI. These opcodes are on the two-byte page behind the
// mandatory 0xF3 prefix and never appear in real mode, so they are guarded here rather
// than by the SingleStepTests real-mode harness.
func TestSSEScalar(t *testing.T) {
	m := &flatRAM{b: make([]byte, 16<<20)}
	putF32(m.b[0x2000:], 2.0)
	putF32(m.b[0x2004:], 3.0)

	code := []byte{
		0xF3, 0x0F, 0x10, 0x05, 0x00, 0x20, 0x00, 0x00, // MOVSS xmm0, [0x2000]
		0xF3, 0x0F, 0x58, 0x05, 0x04, 0x20, 0x00, 0x00, // ADDSS xmm0, [0x2004] -> 5.0
		0xF3, 0x0F, 0x11, 0x05, 0x00, 0x30, 0x00, 0x00, // MOVSS [0x3000], xmm0
		0x0F, 0x57, 0xC9, // XORPS xmm1, xmm1 -> 0
		0xF3, 0x0F, 0x11, 0x0D, 0x04, 0x30, 0x00, 0x00, // MOVSS [0x3004], xmm1
		0xB8, 0x07, 0x00, 0x00, 0x00, // MOV EAX, 7
		0xF3, 0x0F, 0x2A, 0xD0, // CVTSI2SS xmm2, EAX -> 7.0
		0xF3, 0x0F, 0x59, 0x15, 0x00, 0x20, 0x00, 0x00, // MULSS xmm2, [0x2000] -> 14.0
		0xF3, 0x0F, 0x11, 0x15, 0x08, 0x30, 0x00, 0x00, // MOVSS [0x3008], xmm2
		0xF3, 0x0F, 0x2C, 0xC2, // CVTTSS2SI EAX, xmm2 -> 14
		0xF4, // HLT
	}
	copy(m.b[0x1000:], code)

	c := NewCPU(m)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = 0x1000
	c.Regs[SP] = 0x8000
	c.Run(100000)
	if !c.Halted {
		t.Fatalf("SSE program did not halt (EIP=%08X, reason=%q)", c.IP, c.HaltReason)
	}
	if got := getF32(m.b[0x3000:]); got != 5.0 {
		t.Errorf("ADDSS: [0x3000] = %v, want 5.0", got)
	}
	if got := getF32(m.b[0x3004:]); got != 0.0 {
		t.Errorf("XORPS zero: [0x3004] = %v, want 0.0", got)
	}
	if got := getF32(m.b[0x3008:]); got != 14.0 {
		t.Errorf("CVTSI2SS+MULSS: [0x3008] = %v, want 14.0", got)
	}
	if got := c.Regs[AX]; got != 14 {
		t.Errorf("CVTTSS2SI: EAX = %d, want 14", got)
	}
}

// TestSSEPacked exercises the packed-single path: MOVUPS a 4-lane vector, ADDPS it to a
// second vector, and read the lanes back.
func TestSSEPacked(t *testing.T) {
	m := &flatRAM{b: make([]byte, 16<<20)}
	for i, v := range []float32{1, 2, 3, 4} {
		putF32(m.b[0x2000+i*4:], v)
	}
	for i, v := range []float32{10, 20, 30, 40} {
		putF32(m.b[0x2010+i*4:], v)
	}
	code := []byte{
		0x0F, 0x10, 0x05, 0x00, 0x20, 0x00, 0x00, // MOVUPS xmm0, [0x2000]
		0x0F, 0x58, 0x05, 0x10, 0x20, 0x00, 0x00, // ADDPS xmm0, [0x2010]
		0x0F, 0x11, 0x05, 0x00, 0x30, 0x00, 0x00, // MOVUPS [0x3000], xmm0
		0xF4, // HLT
	}
	copy(m.b[0x1000:], code)
	c := NewCPU(m)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = 0x1000
	c.Regs[SP] = 0x8000
	c.Run(100000)
	if !c.Halted {
		t.Fatalf("packed SSE program did not halt (reason=%q)", c.HaltReason)
	}
	want := []float32{11, 22, 33, 44}
	for i, w := range want {
		if got := getF32(m.b[0x3000+i*4:]); got != w {
			t.Errorf("ADDPS lane %d = %v, want %v", i, got, w)
		}
	}
}

// TestSSEShuffleMoveMMX covers the ops the Xbox D3D vertex/matrix code and the XDK fast
// memcpy add on top of the scalar/packed core: SHUFPS (lane select), MOVDQU (128-bit
// integer move), and the MMX MOVQ/EMMS 64-bit copy. None appear in real mode.
func TestSSEShuffleMoveMMX(t *testing.T) {
	m := &flatRAM{b: make([]byte, 16<<20)}
	for i, v := range []float32{1, 2, 3, 4} { // xmm0 source
		putF32(m.b[0x2000+i*4:], v)
	}
	for i, v := range []float32{5, 6, 7, 8} { // xmm1 source
		putF32(m.b[0x2010+i*4:], v)
	}
	for i := 0; i < 8; i++ { // 8 bytes for the MMX copy
		m.b[0x2020+i] = byte(0xA0 + i)
	}
	code := []byte{
		0x0F, 0x10, 0x05, 0x00, 0x20, 0x00, 0x00, // MOVUPS xmm0, [0x2000] = {1,2,3,4}
		0x0F, 0x10, 0x0D, 0x10, 0x20, 0x00, 0x00, // MOVUPS xmm1, [0x2010] = {5,6,7,8}
		0x0F, 0xC6, 0xC1, 0x4E, //                   SHUFPS xmm0, xmm1, 0x4E -> {3,4,5,6}
		0x0F, 0x11, 0x05, 0x00, 0x30, 0x00, 0x00, // MOVUPS [0x3000], xmm0
		0xF3, 0x0F, 0x6F, 0x15, 0x00, 0x20, 0x00, 0x00, // MOVDQU xmm2, [0x2000]
		0xF3, 0x0F, 0x7F, 0x15, 0x10, 0x30, 0x00, 0x00, // MOVDQU [0x3010], xmm2 = {1,2,3,4}
		0x0F, 0x6F, 0x05, 0x20, 0x20, 0x00, 0x00, // MOVQ mm0, [0x2020]
		0x0F, 0x7F, 0x05, 0x20, 0x30, 0x00, 0x00, // MOVQ [0x3020], mm0
		0x0F, 0x77, // EMMS
		0xF4, // HLT
	}
	copy(m.b[0x1000:], code)
	c := NewCPU(m)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = 0x1000
	c.Regs[SP] = 0x8000
	c.Run(100000)
	if !c.Halted {
		t.Fatalf("shuffle/move program did not halt (EIP=%08X reason=%q)", c.IP, c.HaltReason)
	}
	for i, w := range []float32{3, 4, 5, 6} {
		if got := getF32(m.b[0x3000+i*4:]); got != w {
			t.Errorf("SHUFPS lane %d = %v, want %v", i, got, w)
		}
	}
	for i, w := range []float32{1, 2, 3, 4} {
		if got := getF32(m.b[0x3010+i*4:]); got != w {
			t.Errorf("MOVDQU lane %d = %v, want %v", i, got, w)
		}
	}
	for i := 0; i < 8; i++ {
		if got := m.b[0x3020+i]; got != byte(0xA0+i) {
			t.Errorf("MMX MOVQ byte %d = %02X, want %02X", i, got, 0xA0+i)
		}
	}
}

// TestMMXIntOps drives the packed-integer group (mmxint.go) through a small MMX
// program — the op mix OutRun's XMV movie decoder leans on — and checks the lane
// arithmetic: wrap and saturating adds, multiplies, pack/unpack, shifts, compares,
// average, and PMOVMSKB/PSADBW.
func TestMMXIntOps(t *testing.T) {
	m := &flatRAM{b: make([]byte, 16<<20)}
	a := []uint16{0x0001, 0x7FFF, 0x8000, 0x00FF} // mm0 words
	b := []uint16{0x0002, 0x0001, 0xFFFF, 0xFF00} // mm1 words
	for i := range a {
		putw16(m.b[0x2000+i*2:], a[i])
		putw16(m.b[0x2008+i*2:], b[i])
	}
	code := []byte{
		0x0F, 0x6F, 0x05, 0x00, 0x20, 0x00, 0x00, // MOVQ mm0, [0x2000]
		0x0F, 0x6F, 0x0D, 0x08, 0x20, 0x00, 0x00, // MOVQ mm1, [0x2008]
		0x0F, 0x6F, 0xD0, //                         MOVQ mm2, mm0
		0x0F, 0xFD, 0xD1, //                         PADDW mm2, mm1 (wrap)
		0x0F, 0x7F, 0x15, 0x00, 0x30, 0x00, 0x00, // MOVQ [0x3000], mm2
		0x0F, 0x6F, 0xD8, //                         MOVQ mm3, mm0
		0x0F, 0xED, 0xD9, //                         PADDSW mm3, mm1 (signed saturate)
		0x0F, 0x7F, 0x1D, 0x08, 0x30, 0x00, 0x00, // MOVQ [0x3008], mm3
		0x0F, 0x6F, 0xE0, //                         MOVQ mm4, mm0
		0x0F, 0xD5, 0xE1, //                         PMULLW mm4, mm1
		0x0F, 0x7F, 0x25, 0x10, 0x30, 0x00, 0x00, // MOVQ [0x3010], mm4
		0x0F, 0x6F, 0xE8, //                         MOVQ mm5, mm0
		0x0F, 0x71, 0xD5, 0x04, //                   PSRLW mm5, 4
		0x0F, 0x7F, 0x2D, 0x18, 0x30, 0x00, 0x00, // MOVQ [0x3018], mm5
		0x0F, 0x6F, 0xF0, //                         MOVQ mm6, mm0
		0x0F, 0x67, 0xF1, //                         PACKUSWB mm6, mm1
		0x0F, 0x7F, 0x35, 0x20, 0x30, 0x00, 0x00, // MOVQ [0x3020], mm6
		0x0F, 0x6F, 0xF8, //                         MOVQ mm7, mm0
		0x0F, 0x75, 0xF9, //                         PCMPEQW mm7, mm1
		0x0F, 0x7F, 0x3D, 0x28, 0x30, 0x00, 0x00, // MOVQ [0x3028], mm7
		0x0F, 0xD7, 0xC8, //                         PMOVMSKB ecx, mm0
		0x0F, 0xEF, 0xC0, //                         PXOR mm0, mm0
		0x0F, 0x7F, 0x05, 0x30, 0x30, 0x00, 0x00, // MOVQ [0x3030], mm0
		0x0F, 0x77, // EMMS
		0xF4, // HLT
	}
	copy(m.b[0x1000:], code)
	c := NewCPU(m)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = 0x1000
	c.Regs[SP] = 0x8000
	c.Run(100000)
	if !c.Halted || !strings.HasPrefix(c.HaltReason, "HLT") {
		t.Fatalf("MMX program did not halt cleanly (EIP=%08X reason=%q)", c.IP, c.HaltReason)
	}
	check := func(base int, want []uint16, what string) {
		t.Helper()
		for i, w := range want {
			if got := w16(m.b[base+i*2:]); got != w {
				t.Errorf("%s lane %d = %04X, want %04X", what, i, got, w)
			}
		}
	}
	check(0x3000, []uint16{0x0003, 0x8000, 0x7FFF, 0xFFFF}, "PADDW")
	check(0x3008, []uint16{0x0003, 0x7FFF, 0x8000, 0xFFFF}, "PADDSW")
	check(0x3010, []uint16{0x0002, 0x7FFF, 0x8000, 0x0100}, "PMULLW")
	check(0x3018, []uint16{0x0000, 0x07FF, 0x0800, 0x000F}, "PSRLW")
	// PACKUSWB: mm0 words {1,7FFF,8000,FF} -> bytes {01,FF,00,FF}; mm1 {2,1,FFFF,FF00} -> {02,01,00,00}
	if got := le32b(m.b[0x3020:]); got != 0xFF00FF01 {
		t.Errorf("PACKUSWB low = %08X, want FF00FF01", got)
	}
	if got := le32b(m.b[0x3024:]); got != 0x00000102 {
		t.Errorf("PACKUSWB high = %08X, want 00000102", got)
	}
	check(0x3028, []uint16{0x0000, 0x0000, 0x0000, 0x0000}, "PCMPEQW")
	check(0x3030, []uint16{0, 0, 0, 0}, "PXOR")
	// PMOVMSKB of {01 00 FF 7F 00 80 FF 00}: bytes with MSB set -> mask
	if c.Regs[CX] != 0b01100100 { // bytes 01 00 FF 7F 00 80 FF 00: MSBs at 2, 5, 6
		t.Errorf("PMOVMSKB = %08b, want 01100100", c.Regs[CX])
	}
}

// TestSSEUnpck pins UNPCKLPS/UNPCKHPS (0F 14 / 0F 15) and their PD forms against the
// lane order the manual specifies, rather than only against "the WMA decoder stopped
// halting". The single-precision forms interleave two dwords from each operand; the
// high form takes lanes 2/3 where the low form takes 0/1. Getting the halves the wrong
// way round still decodes, still runs, and silently scrambles every sample.
func TestSSEUnpck(t *testing.T) {
	m := &flatRAM{b: make([]byte, 16<<20)}
	// xmm0 = [1,2,3,4], the memory operand = [5,6,7,8].
	for i, v := range []float32{1, 2, 3, 4} {
		putF32(m.b[0x2000+i*4:], v)
	}
	for i, v := range []float32{5, 6, 7, 8} {
		putF32(m.b[0x2010+i*4:], v)
	}

	code := []byte{
		0x0F, 0x10, 0x05, 0x00, 0x20, 0x00, 0x00, // MOVUPS xmm0, [0x2000]
		0x0F, 0x28, 0xC8, // MOVAPS xmm1, xmm0
		0x0F, 0x14, 0x0D, 0x10, 0x20, 0x00, 0x00, // UNPCKLPS xmm1, [0x2010] -> 1,5,2,6
		0x0F, 0x11, 0x0D, 0x00, 0x30, 0x00, 0x00, // MOVUPS [0x3000], xmm1
		0x0F, 0x28, 0xD0, // MOVAPS xmm2, xmm0
		0x0F, 0x15, 0x15, 0x10, 0x20, 0x00, 0x00, // UNPCKHPS xmm2, [0x2010] -> 3,7,4,8
		0x0F, 0x11, 0x15, 0x10, 0x30, 0x00, 0x00, // MOVUPS [0x3010], xmm2
		0x66, 0x0F, 0x28, 0xD8, // MOVAPD xmm3, xmm0
		0x66, 0x0F, 0x15, 0x1D, 0x10, 0x20, 0x00, 0x00, // UNPCKHPD xmm3, [0x2010] -> 3,4,7,8
		0x0F, 0x11, 0x1D, 0x20, 0x30, 0x00, 0x00, // MOVUPS [0x3020], xmm3
		0xF4, // HLT
	}
	copy(m.b[0x1000:], code)

	c := NewCPU(m)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = 0x1000
	c.Regs[SP] = 0x8000
	c.Run(100000)
	if !c.Halted || strings.Contains(c.HaltReason, "unimplemented") {
		t.Fatalf("UNPCK program did not run (EIP=%08X, reason=%q)", c.IP, c.HaltReason)
	}
	for _, tc := range []struct {
		name string
		addr int
		want []float32
	}{
		{"UNPCKLPS", 0x3000, []float32{1, 5, 2, 6}},
		{"UNPCKHPS", 0x3010, []float32{3, 7, 4, 8}},
		{"UNPCKHPD", 0x3020, []float32{3, 4, 7, 8}},
	} {
		for i, want := range tc.want {
			if got := getF32(m.b[tc.addr+i*4:]); got != want {
				t.Errorf("%s lane %d = %v, want %v", tc.name, i, got, want)
			}
		}
	}
}
