package n3ds

// Unit tests for the PICA200 shader VM: tiny hand-assembled programs, the same
// style as the CPU cores' tests. The encodings mirror gpu_shader.go's decode;
// the game-shader disassembly (picadump -shader) established the field layout
// against real code.

import (
	"math"
	"testing"

	"retroreverse.com/tools/cpu/arm"
)

// swizIdentity is the identity operand descriptor: full write mask, xyzw
// selectors on all three operands, no negation.
const swizIdentity = 0xF | 0x1B<<5 | 0x1B<<14 | 0x1B<<23

// fmt1 assembles a format-1 arithmetic instruction.
func fmt1(op, dst, idx, s1, s2, desc uint32) uint32 {
	return op<<26 | dst<<21 | idx<<19 | s1<<12 | s2<<7 | desc
}

func testGPU() *GPU {
	m := &Machine{}
	m.CPU = arm.NewCPU(m)
	return newGPU(m)
}

func TestShaderMovAddMul(t *testing.T) {
	g := testGPU()
	g.Opdesc[0] = swizIdentity
	// r0 = v0; r1 = r0 + v1; o0 = r1 * v1; end
	g.Code[0] = fmt1(0x13, 0x10, 0, 0x00, 0, 0) // mov r0, v0
	g.Code[1] = fmt1(0x00, 0x11, 0, 0x10, 1, 0) // add r1, r0, v1
	g.Code[2] = fmt1(0x08, 0x00, 0, 0x11, 1, 0) // mul o0, r1, v1
	g.Code[3] = 0x22 << 26                      // end

	var v [16][4]float32
	v[0] = [4]float32{1, 2, 3, 4}
	v[1] = [4]float32{10, 20, 30, 40}
	o, ok := runShader(g, &v)
	if !ok {
		t.Fatal("shader halted")
	}
	want := [4]float32{110, 440, 990, 1760} // (v0+v1)*v1
	if o[0] != want {
		t.Errorf("o0 = %v, want %v", o[0], want)
	}
}

func TestShaderDP4MatrixRow(t *testing.T) {
	g := testGPU()
	g.Opdesc[0] = swizIdentity
	// The idiom every captured shader uses: dp4 o0.x, c0.wzyx, v0 — a matrix
	// row stored w-first (upload order) read back through the reversing
	// swizzle.
	wzyx := uint32(0xF | 0xE4<<5 | 0x1B<<14 | 0x1B<<23) // src1 selectors w,z,y,x
	g.Opdesc[1] = wzyx
	g.Code[0] = fmt1(0x02, 0x00, 0, 0x20, 0, 1) // dp4 o0, c0.wzyx, v0
	g.Code[1] = 0x22 << 26

	// Simulate the upload order the game uses: w,z,y,x words for the row
	// (2, 0, 0, 5) meaning x-coefficient 2, w-coefficient 5.
	g.fltF32 = true
	g.fltIdx = 0
	for _, w := range []float32{5, 0, 0, 2} { // w z y x
		g.floatUniformWord(math.Float32bits(w))
	}

	var v [16][4]float32
	v[0] = [4]float32{3, 7, 11, 1}
	o, ok := runShader(g, &v)
	if !ok {
		t.Fatal("shader halted")
	}
	// c0 stored as {x:2, y:0, z:0, w:5}; read back .wzyx = (5,0,0,2);
	// dp4 with v0 = 5*3 + 0 + 0 + 2*1 = 17.
	if o[0][0] != 17 {
		t.Errorf("dp4 = %v, want 17", o[0][0])
	}
}

func TestShaderMAD(t *testing.T) {
	g := testGPU()
	g.Opdesc[0] = 0xF | 0x1B<<5 | 0x1B<<14 | 0x1B<<23&0x1F // MAD descs are 5 bits in the instr; table entry still full
	g.Opdesc[0] = swizIdentity
	// mad r2, v0, v1, v2  (wide src2 form, op 0x38)
	g.Code[0] = 0x38<<26 | 0x12<<24 | 0<<22 | 0x00<<17 | 0x01<<10 | 0x02<<5 | 0
	// mov o0, r2
	g.Code[1] = fmt1(0x13, 0x00, 0, 0x12, 0, 0)
	g.Code[2] = 0x22 << 26

	var v [16][4]float32
	v[0] = [4]float32{2, 3, 4, 5}
	v[1] = [4]float32{10, 10, 10, 10}
	v[2] = [4]float32{1, 1, 1, 1}
	o, ok := runShader(g, &v)
	if !ok {
		t.Fatal("shader halted")
	}
	want := [4]float32{21, 31, 41, 51}
	if o[0] != want {
		t.Errorf("mad = %v, want %v", o[0], want)
	}
}

func TestShaderCMPAndIFC(t *testing.T) {
	g := testGPU()
	g.Opdesc[0] = swizIdentity
	// cmp v0 eq|eq v1 ; ifc(x) → o0 = v0 else o0 = v1
	g.Code[0] = 0x2E<<26 | 0<<24 | 0<<21 | 0x00<<12 | 0x01<<7 | 0 // cmp eq,eq
	g.Code[1] = 0x28<<26 | 1<<25 | 2<<22 | 3<<10 | 1              // ifc x==true: else at 3, num 1
	g.Code[2] = fmt1(0x13, 0x00, 0, 0x00, 0, 0)                   // then: mov o0, v0
	g.Code[3] = fmt1(0x13, 0x00, 0, 0x01, 0, 0)                   // else: mov o0, v1
	g.Code[4] = 0x22 << 26

	var v [16][4]float32
	v[0] = [4]float32{5, 5, 5, 5}
	v[1] = [4]float32{5, 9, 9, 9}
	o, ok := runShader(g, &v)
	if !ok {
		t.Fatal("shader halted")
	}
	if o[0] != v[0] {
		t.Errorf("ifc took the wrong branch: o0 = %v", o[0])
	}
}

func TestF24(t *testing.T) {
	cases := []struct {
		bits uint32
		want float32
	}{
		{0x000000, 0},
		{0x3F0000, 1},   // exp 63, mantissa 0
		{0xBF0000, -1},  // sign + exp 63
		{0x400000, 2},   // exp 64
		{0x3E0000, 0.5}, // exp 62
	}
	for _, c := range cases {
		if got := f24bits(c.bits); got != c.want {
			t.Errorf("f24(0x%06X) = %v, want %v", c.bits, got, c.want)
		}
	}
}

func TestUnpackF24x4(t *testing.T) {
	// Pack (x=1, y=-1, z=2, w=0.5) by hand: w in the top of word0, x in the
	// low 24 bits of word2.
	x, y, z, w := uint32(0x3F0000), uint32(0xBF0000), uint32(0x400000), uint32(0x3E0000)
	w0 := w<<8 | z>>16
	w1 := z<<16 | y>>8
	w2 := y<<24 | x
	got := unpackF24x4(w0, w1, w2)
	want := [4]float32{1, -1, 2, 0.5}
	if got != want {
		t.Errorf("unpackF24x4 = %v, want %v", got, want)
	}
}

// TestAttrInputPermutation pins the direction of the vertex-attribute → shader
// input-register map (registers 0x2BB/0x2BC). Captain Toad's opening stage is
// the case that settles it: a draw fetching attributes 0-2 (position, UV,
// colour) with permutation 0x340. Delivered attribute-first, every fetched
// attribute lands in a distinct register. Register-first — the inverse — v1 and
// v2 would take attributes 4 and 3, which this draw never fetches, and the UV
// and colour would reach no register at all. That inversion is what turned the
// stage into spiky radiating triangles.
func TestAttrInputPermutation(t *testing.T) {
	var attrs [16][4]float32
	pos := [4]float32{-601.65, 448.437, 22.0549, 1}
	uv := [4]float32{0.0054, 0.7544, 0, 1}
	col := [4]float32{198, 198, 198, 255}
	attrs[0], attrs[1], attrs[2] = pos, uv, col

	var v [16][4]float32
	mapAttrsToInputs(&v, &attrs, 0x340, 3)
	if v[0] != pos {
		t.Errorf("v0 = %v, want the position %v", v[0], pos)
	}
	if v[4] != uv {
		t.Errorf("v4 = %v, want the UV %v", v[4], uv)
	}
	if v[3] != col {
		t.Errorf("v3 = %v, want the colour %v", v[3], col)
	}
	// Registers no attribute names keep the default, and in particular the
	// position must not leak into them: reading the permutation backwards fed
	// v5-v15 the position, which is exactly how the UV came out as coordinates.
	for _, j := range []int{1, 2, 5, 15} {
		if v[j] != ([4]float32{0, 0, 0, 1}) {
			t.Errorf("v%d = %v, want the unwritten default", j, v[j])
		}
	}
}

// runShader is the tests' hand on shaderRun, which the draw path calls with a
// caller-owned output file and a hoisted entry point (both per draw, not per
// vertex). Here each case wants a fresh output and the entry from the registers.
func runShader(g *GPU, v *[16][4]float32) ([16][4]float32, bool) {
	var o [16][4]float32
	ok := g.shaderRun(v, &o, int(g.Regs[regVshEntry]&0xFFF))
	return o, ok
}
