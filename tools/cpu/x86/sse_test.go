package x86

import (
	"math"
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
