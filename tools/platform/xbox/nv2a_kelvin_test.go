package xbox

// nv2a_kelvin_test.go exercises the Kelvin pipeline without the disc: the vertex-
// program interpreter against hand-encoded instructions, the DXT1 block decoder
// against a hand-computed block, and an end-to-end synthetic draw (program upload,
// attribute formats, BEGIN/INLINE_ARRAY/END) whose pixels land in RAM.

import (
	"math"
	"testing"
)

// gpuMachine builds a disc-less machine from a loadable synthetic XBE (one section
// whose raw bytes exist in the fixture) for pipeline tests.
func gpuMachine(t *testing.T) *Machine {
	t.Helper()
	b := buildXBE(t)
	// buildXBE's section claims raw 0x1000..0x2000, beyond the 4 KB fixture — the
	// parse tests never load it. Point the section at an in-fixture range instead.
	put32 := func(off, v uint32) {
		b[off] = byte(v)
		b[off+1] = byte(v >> 8)
		b[off+2] = byte(v >> 16)
		b[off+3] = byte(v >> 24)
	}
	put32(0x180+0x08, 0x800) // vsize
	put32(0x180+0x0C, 0x800) // rawaddr
	put32(0x180+0x10, 0x800) // rawsize
	xbe, err := ParseXBE(b)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMachine(xbe, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.EnableGPU()
	return m
}

// vshEncode builds one 4-dword instruction from the field map (the inverse of
// vshDecode; swizzles are identity, no negates).
type vshEnc struct {
	mac, ilu           int
	constIdx, inputReg int
	aMux, aReg         int
	bMux, bReg         int
	cMux, cReg         int
	macMask, macDst    uint32
	iluMask            uint32
	outMask            uint32
	outAddr            int
	outFromILU         bool
	final              bool
}

func (e vshEnc) words() [4]uint32 {
	const identSwz = 0x1B // x,y,z,w in 2-bit fields
	var w [4]uint32
	w[1] = uint32(e.ilu)<<25 | uint32(e.mac)<<21 | uint32(e.constIdx)<<13 |
		uint32(e.inputReg)<<9 | identSwz
	w[2] = uint32(e.aReg)<<28 | uint32(e.aMux)<<26 |
		uint32(identSwz)<<17 | uint32(e.bReg)<<13 | uint32(e.bMux)<<11 |
		uint32(identSwz)<<2 | uint32(e.cReg>>2)
	w[3] = uint32(e.cReg&3)<<30 | uint32(e.cMux)<<28 |
		e.macMask<<24 | e.macDst<<20 | e.iluMask<<16 | e.outMask<<12 |
		1<<11 | uint32(e.outAddr)<<3
	if e.outFromILU {
		w[3] |= 1 << 2
	}
	if e.final {
		w[3] |= 1
	}
	return w
}

func fbits(f float32) uint32 { return math.Float32bits(f) }

// TestVSHInterpreter runs a hand-encoded program: R1 = v0 * c0; oPos = R1 + c1;
// oD0 = v3; dual-issue RCP into R1.w — checking arithmetic, temp writes, output
// mapping, and the FINAL stop.
func TestVSHInterpreter(t *testing.T) {
	m := gpuMachine(t)
	g := m.pgraph

	prog := [][4]uint32{
		// MUL R1.xyzw, v0, c0
		{0, 0, 0, 0}, // placeholder, filled below
		// ADD oPos.xyzw, R1, c1  (ADD reads A and C)
		{0, 0, 0, 0},
		// MOV oD0, v3 ; RCP R1.w <- c2.x(w swizzled? identity: c2.x used)  [FINAL]
		{0, 0, 0, 0},
	}
	prog[0] = vshEnc{mac: macMUL, inputReg: 0, constIdx: 0,
		aMux: 2, bMux: 3, cMux: 1, macMask: 0xF, macDst: 1, outAddr: 0xFF}.words()
	prog[1] = vshEnc{mac: macADD, constIdx: 1,
		aMux: 1, aReg: 1, bMux: 1, cMux: 3, outMask: 0xF, outAddr: 0}.words()
	prog[2] = vshEnc{mac: macMOV, ilu: iluRCP, inputReg: 3, constIdx: 2,
		aMux: 2, bMux: 2, cMux: 3, outMask: 0xF, outAddr: 3,
		iluMask: 0x1, final: true}.words()

	for i, p := range prog {
		g.Prog[i] = p
	}
	g.Const[0] = [4]uint32{fbits(2), fbits(2), fbits(2), fbits(2)}
	g.Const[1] = [4]uint32{fbits(10), fbits(20), fbits(30), fbits(40)}
	g.Const[2] = [4]uint32{fbits(4), fbits(4), fbits(4), fbits(4)}

	var in [16][4]float32
	in[0] = [4]float32{1, 2, 3, 1}
	in[3] = [4]float32{0.5, 0.25, 0.125, 1}
	var out [13][4]float32
	if !g.vshRun(&in, &out) {
		t.Fatalf("vshRun halted: %s", m.CPU.HaltReason)
	}
	wantPos := [4]float32{12, 24, 36, 42}
	if out[0] != wantPos {
		t.Errorf("oPos = %v, want %v", out[0], wantPos)
	}
	if out[3] != in[3] {
		t.Errorf("oD0 = %v, want %v", out[3], in[3])
	}
}

// TestVSHDisasmAgainstLiveProgram pins the field map against four instruction words
// captured from OutRun's own uploaded program (the survey's 0x0B00 firstArgs): the
// decode must read MOV R1,v0 / dual MOV-oD0+RCP / MOV oFog / dual MUL+MOV-oD1.
func TestVSHDisasmAgainstLiveProgram(t *testing.T) {
	m := gpuMachine(t)
	g := m.pgraph
	g.Prog[0] = [4]uint32{0, 0x0020001B, 0x0836106C, 0x2F100FF8}
	g.Prog[1] = [4]uint32{0, 0x0420061B, 0x083613FC, 0x5011F818}
	g.Prog[2] = [4]uint32{0, 0x002008FF, 0x0836106C, 0x2070F828}
	g.Prog[3] = [4]uint32{0, 0x0240081B, 0x1436186C, 0x2F20F824}
	want := []string{
		"MOV R1, v0",
		"MOV oD0, v3 ; RCP R1.w, R1.wwww",
		"MOV oFog, v4.wwww",
		"MUL R2, R1, c0 ; MOV oD1, v4",
	}
	for i, w := range want {
		if got := g.vshDisasm(i); got != w {
			t.Errorf("inst %d: %q, want %q", i, got, w)
		}
	}
}

// TestDXT1Decode checks one hand-built block: c0=white > c1=black, selectors walking
// 0,1,2,3 give white, black, 2/3 white, 1/3 white.
func TestDXT1Decode(t *testing.T) {
	img := &texImage{w: 4, h: 4, pix: make([]byte, 4*4*4)}
	ram := make([]byte, 16)
	// c0 = 0xFFFF (white), c1 = 0x0000 (black), selectors: texel i uses code i&3.
	ram[0], ram[1] = 0xFF, 0xFF
	ram[2], ram[3] = 0x00, 0x00
	ram[4], ram[5], ram[6], ram[7] = 0xE4, 0xE4, 0xE4, 0xE4 // 11100100 = 3,2,1,0
	decodeDXT(img, ram, 0, 1)
	// Texel 0: code 0 = white; texel 1: code 1 = black; texel 2: code 2 = 2/3 white.
	if img.pix[0] != 0xFF || img.pix[1] != 0xFF || img.pix[2] != 0xFF {
		t.Errorf("texel 0 = %v, want white", img.pix[0:4])
	}
	if img.pix[4] != 0 || img.pix[5] != 0 || img.pix[6] != 0 {
		t.Errorf("texel 1 = %v, want black", img.pix[4:8])
	}
	if img.pix[8] != 0xA9 && img.pix[8] != 0xAA {
		t.Errorf("texel 2 r = %d, want ~2/3 of 255", img.pix[8])
	}
}

// TestKelvinInlineDraw drives the whole pipeline as the pusher would: program +
// constants + surface state + formats through kelvinMethod, then BEGIN, 21 inline
// words (3 vertices of float4 pos + D3DCOLOR + float2 uv), END — and checks the
// triangle rasterised into RAM with the diffuse color (no texture stage enabled).
func TestKelvinInlineDraw(t *testing.T) {
	m := gpuMachine(t)
	g := m.pgraph
	kv := func(method, arg uint32) { g.kelvinMethod(method, arg) }

	// Passthrough program: MOV oPos, v0 / MOV oD0, v3 [FINAL] via the upload FIFO.
	kv(kelvinProgLoad, 0)
	inst1 := vshEnc{mac: macMOV, inputReg: 0,
		aMux: 2, bMux: 2, cMux: 2, outMask: 0xF, outAddr: 0}.words()
	inst2 := vshEnc{mac: macMOV, inputReg: 3,
		aMux: 2, bMux: 2, cMux: 2, outMask: 0xF, outAddr: 3, final: true}.words()
	for _, w := range append(inst1[:], inst2[:]...) {
		kv(kelvinProgData, w)
	}
	kv(kelvinProgStart, 0)
	kv(kelvinTransformExecMode, 6)

	// Surface: 64x64 A8R8G8B8 + Z24S8, pitch 256, color at 1MB, no AA.
	const colorBase = 0x100000
	kv(kelvinSurfaceFormat, 0x0128)
	kv(kelvinSurfacePitch, 0x0100_0100)
	kv(kelvinSurfaceColorOffset, colorBase)
	kv(kelvinSurfaceZetaOffset, 0x180000)
	kv(kelvinSurfaceClipH, 64<<16)
	kv(kelvinSurfaceClipV, 64<<16)
	kv(kelvinColorMask, 0x01010101)
	kv(kelvinCombinerControl, 1)
	// Stage 0: spare0 = COL0 (A=col0, B=1-zero); alpha likewise; final D = spare0.
	kv(kelvinCombinerColorICW, 0x04_20_00_00) // A=col0(4,map0) B=zero-invert(0x20)=1 C=0 D=0
	kv(kelvinCombinerAlphaICW, 0x14_30_00_00) // A=col0.a B=1 (alpha side)
	kv(kelvinCombinerColorOCW, 0xC00)         // sum -> spare0
	kv(kelvinCombinerAlphaOCW, 0xC00)
	kv(kelvinSpecFogCW0, 0x0000000C) // D = spare0
	kv(kelvinSpecFogCW1, 0x00001C00) // G = spare0 alpha

	// Attribute formats: v0 = float x4, v3 = D3DCOLOR, v9 = float x2; rest disabled.
	for i := uint32(0); i < 16; i++ {
		kv(kelvinVtxArrayFormat+4*i, 0x02)
	}
	kv(kelvinVtxArrayFormat+0, 0x42)
	kv(kelvinVtxArrayFormat+4*3, 0x40)
	kv(kelvinVtxArrayFormat+4*9, 0x22)

	// One right triangle covering the surface's upper-left, diffuse = pure green.
	kv(kelvinBeginEnd, primTriangles)
	tri := [][7]uint32{
		{fbits(0), fbits(0), 0, fbits(1), 0xFF00FF00, fbits(0), fbits(0)},
		{fbits(60), fbits(0), 0, fbits(1), 0xFF00FF00, fbits(1), fbits(0)},
		{fbits(0), fbits(60), 0, fbits(1), 0xFF00FF00, fbits(0), fbits(1)},
	}
	for _, v := range tri {
		for _, w := range v {
			kv(kelvinInlineArray, w)
		}
	}
	kv(kelvinBeginEnd, 0)

	if m.CPU.Halted {
		t.Fatalf("draw halted: %s", m.CPU.HaltReason)
	}
	if g.Draws != 1 {
		t.Fatalf("Draws = %d, want 1", g.Draws)
	}
	// A pixel inside the triangle: (10, 10) -> B,G,R,A = 0,255,0,255.
	o := uint32(colorBase) + 10*256 + 10*4
	if m.RAM[o] != 0 || m.RAM[o+1] != 255 || m.RAM[o+2] != 0 {
		t.Errorf("pixel (10,10) = B%d G%d R%d, want green", m.RAM[o], m.RAM[o+1], m.RAM[o+2])
	}
	// A pixel outside (55, 55) stays zero.
	o = uint32(colorBase) + 55*256 + 55*4
	if m.RAM[o+1] != 0 {
		t.Errorf("pixel (55,55) G = %d, want untouched 0", m.RAM[o+1])
	}
}

// TestKelvinStateRoundTrip covers the new pipeline fields in the savestate: program,
// constants, staging, and an open BEGIN batch survive a save/load.
func TestKelvinStateRoundTrip(t *testing.T) {
	m := gpuMachine(t)
	g := m.pgraph
	g.kelvinMethod(kelvinProgLoad, 7)
	g.kelvinMethod(kelvinProgData, 0x11111111) // partial staging: 1 of 4 dwords
	g.kelvinMethod(kelvinConstLoad, 9)
	for _, w := range []uint32{fbits(1), fbits(2), fbits(3), fbits(4)} {
		g.kelvinMethod(kelvinConstData, w)
	}
	g.kelvinMethod(kelvinBeginEnd, primQuadStrip)
	g.kelvinMethod(kelvinInlineArray, 0xAABBCCDD)
	g.kelvinMethod(kelvinVertexData4C+4*5, 0x80402010)

	st := m.SaveState()
	m2 := gpuMachine(t)
	if err := m2.LoadState(st); err != nil {
		t.Fatal(err)
	}
	g2 := m2.pgraph
	if g2.ProgLoad != 7 || g2.progBufN != 1 || g2.progBuf[0] != 0x11111111 {
		t.Errorf("program staging: load=%d n=%d buf0=%08X", g2.ProgLoad, g2.progBufN, g2.progBuf[0])
	}
	if g2.ConstLoad != 10 || g2.Const[9] != [4]uint32{fbits(1), fbits(2), fbits(3), fbits(4)} {
		t.Errorf("constants: load=%d c9=%v", g2.ConstLoad, g2.Const[9])
	}
	if g2.prim != primQuadStrip || len(g2.inline) != 1 || g2.inline[0] != 0xAABBCCDD {
		t.Errorf("open batch: prim=%d inline=%v", g2.prim, g2.inline)
	}
	if g2.vtxAttr[5] != g.vtxAttr[5] {
		t.Errorf("vtxAttr[5] = %v, want %v", g2.vtxAttr[5], g.vtxAttr[5])
	}
}
