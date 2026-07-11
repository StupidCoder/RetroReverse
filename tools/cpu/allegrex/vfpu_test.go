package allegrex

// Hand-assembled loop tests for the VFPU: register addressing (columns, rows,
// matrices, transposition), the quad loads/stores, the operand prefixes, and each
// executed compute op. Expected values are computed independently in the test
// bodies; register numbers and flat V indices are written out per the layout
// V[mtx*16 + col*4 + row].

import (
	"math"
	"testing"
)

// --- encoders ----------------------------------------------------------------

// szBits encodes an op's vector/matrix size (1..4 = .s/.p/.t/.q) into bits 7/15.
func szBits(n uint32) uint32 { return (n-1)&1<<7 | ((n-1)>>1&1)<<15 }

// vop3 encodes a VFPU0/1/3 three-operand op.
func vop3(op, sub, vd, vs, vt, n uint32) uint32 {
	return op<<26 | sub<<23 | vt<<16 | vs<<8 | vd | szBits(n)
}

// vop4 encodes a 0x34-group op (rs5 selects the group, rt5 the op/immediate).
func vop4(rs5, rt5, vd, vs, n uint32) uint32 {
	return 0x34<<26 | rs5<<21 | rt5<<16 | vs<<8 | vd | szBits(n)
}

// vmat encodes a 0x3C-group op.
func vmat(sub, vd, vs, vt, n uint32) uint32 {
	return 0x3C<<26 | sub<<23 | vt<<16 | vs<<8 | vd | szBits(n)
}

func vmset(fn, vd, vs, n uint32) uint32 {
	return 0x3C<<26 | 28<<21 | fn<<16 | vs<<8 | vd | szBits(n)
}

func vpfx(which, data uint32) uint32 { return 0x37<<26 | which<<24 | data }

func lvq(rs, vt, off uint32) uint32 {
	return 0x36<<26 | rs<<21 | (vt&31)<<16 | off&0xFFFC | vt>>5&1
}
func svq(rs, vt, off uint32) uint32 {
	return 0x3E<<26 | rs<<21 | (vt&31)<<16 | off&0xFFFC | vt>>5&1
}

func mtv(rt, vreg uint32) uint32 { return 0x12<<26 | 7<<21 | rt<<16 | vreg }
func mfv(rt, vreg uint32) uint32 { return 0x12<<26 | 3<<21 | rt<<16 | vreg }

func f2b(f float32) uint32 { return math.Float32bits(f) }

// setV seeds flat V indices with float values.
func setV(c *CPU, base uint32, vals ...float32) {
	for i, v := range vals {
		c.V[base+uint32(i)] = math.Float32bits(v)
	}
}

// getV reads a flat V index as a float.
func getV(c *CPU, i uint32) float32 { return math.Float32frombits(c.V[i]) }

func checkV(t *testing.T, c *CPU, base uint32, want ...float32) {
	t.Helper()
	for i, w := range want {
		if got := getV(c, base+uint32(i)); got != w {
			t.Errorf("V[%d] = %v, want %v", base+uint32(i), got, w)
		}
	}
}

// --- tests ---------------------------------------------------------------------

func TestVFPUQuadLoadStore(t *testing.T) {
	const base = 8
	m, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(base, 0x2000)
		for i := uint32(0); i < 4; i++ {
			m.w32(0x2000+i*4, f2b(float32(i+1)))
		}
	},
		lvq(base, 0, 0),     // lv.q C000, 0($t0)
		svq(base, 0, 0x20),  // sv.q C000, 32($t0)
		lvq(base, 0x21, 0),  // lv.q R010, 0($t0)  (transposed: row 0 of col.. row regs)
		svq(base, 0x21, 64), // sv.q R010, 64($t0)
	)
	// C000 (flats 0..3) round-trips through the sv.q copy at 0x2020. (The register
	// itself is checked via memory: the later row load overlaps flat 1.)
	for i := uint32(0); i < 4; i++ {
		if got := m.r32(0x2020 + i*4); got != f2b(float32(i+1)) {
			t.Errorf("sv.q word %d = 0x%08X", i, got)
		}
	}
	// R010 (reg 0x21: transpose, col 1) strides 4: flats 1, 5, 9, 13.
	for i, flat := range []uint32{1, 5, 9, 13} {
		if got := getV(c, flat); got != float32(i+1) {
			t.Errorf("row reg element %d (flat %d) = %v", i, flat, got)
		}
	}
	for i := uint32(0); i < 4; i++ {
		if got := m.r32(0x2040 + i*4); got != f2b(float32(i+1)) {
			t.Errorf("sv.q row word %d = 0x%08X", i, got)
		}
	}
}

func TestVFPUArith(t *testing.T) {
	// C100 = (1,2,3,4), C110 = (10,20,30,40); results into C000.
	seed := func(m *memBus, c *CPU) {
		setV(c, 16, 1, 2, 3, 4)
		setV(c, 20, 10, 20, 30, 40)
	}
	const vd, vs, vt = 0, 4, 5 // C000, C100, C110
	cases := []struct {
		name string
		inst uint32
		want [4]float32
	}{
		{"vadd.q", vop3(0x18, 0, vd, vs, vt, 4), [4]float32{11, 22, 33, 44}},
		{"vsub.q", vop3(0x18, 1, vd, vs, vt, 4), [4]float32{-9, -18, -27, -36}},
		{"vdiv.q", vop3(0x18, 7, vd, vt, vs, 4), [4]float32{10, 10, 10, 10}},
		{"vmul.q", vop3(0x19, 0, vd, vs, vt, 4), [4]float32{10, 40, 90, 160}},
		{"vmin.q", vop3(0x1B, 2, vd, vs, vt, 4), [4]float32{1, 2, 3, 4}},
		{"vmax.q", vop3(0x1B, 3, vd, vs, vt, 4), [4]float32{10, 20, 30, 40}},
		{"vslt.q", vop3(0x1B, 7, vd, vs, vt, 4), [4]float32{1, 1, 1, 1}},
		{"vsge.q", vop3(0x1B, 6, vd, vs, vt, 4), [4]float32{0, 0, 0, 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, c := run(t, seed, tc.inst)
			checkV(t, c, 0, tc.want[0], tc.want[1], tc.want[2], tc.want[3])
		})
	}
}

func TestVFPUDotHdpScl(t *testing.T) {
	seed := func(m *memBus, c *CPU) {
		setV(c, 16, 1, 2, 3, 4)     // C100
		setV(c, 20, 10, 20, 30, 40) // C110
	}
	_, c := run(t, seed, vop3(0x19, 1, 0, 4, 5, 4)) // vdot.q S000, C100, C110
	if got := getV(c, 0); got != 1*10+2*20+3*30+4*40 {
		t.Errorf("vdot = %v", got)
	}
	// vhdp.q: S's last lane wired to 1 => 1*10+2*20+3*30+40.
	_, c = run(t, seed, vop3(0x19, 4, 0, 4, 5, 4))
	if got := getV(c, 0); got != 1*10+2*20+3*30+40 {
		t.Errorf("vhdp = %v", got)
	}
	// vscl.t C000, C100, S110 (scale by 10).
	_, c = run(t, seed, vop3(0x19, 2, 0, 4, 5, 3))
	checkV(t, c, 0, 10, 20, 30)
}

func TestVFPUPrefixes(t *testing.T) {
	seed := func(m *memBus, c *CPU) {
		setV(c, 16, 1, 2, 3, 4)         // C100
		setV(c, 20, 100, 100, 100, 100) // C110
		setV(c, 0, -1, -1, -1, -1)      // C000 (to observe the write mask)
	}
	// vpfxs broadcast x: swizzle bits all select lane 0.
	_, c := run(t, seed, vpfx(0, 0x00), vop3(0x18, 0, 0, 4, 5, 4))
	checkV(t, c, 0, 101, 101, 101, 101)

	// vpfxs negate all lanes (identity swizzle + negate bits 16..19).
	_, c = run(t, seed, vpfx(0, 0xF00E4), vop3(0x18, 0, 0, 4, 5, 4))
	checkV(t, c, 0, 99, 98, 97, 96)

	// vpfxt constants: every lane the constant 1 (constant bits + swizzle=1).
	_, c = run(t, seed, vpfx(1, 0xF055), vop3(0x18, 0, 0, 4, 5, 4))
	checkV(t, c, 0, 2, 3, 4, 5)

	// vpfxd write mask on lanes 1,2: vmov.q writes only lanes 0,3.
	_, c = run(t, seed, vpfx(2, 0x600), vop4(0, 0, 0, 4, 4))
	checkV(t, c, 0, 1, -1, -1, 4)

	// vpfxd saturate [0:1] on all lanes.
	seedSat := func(m *memBus, c *CPU) { setV(c, 16, -2, 0.5, 2, 1) }
	_, c = run(t, seedSat, vpfx(2, 0x55), vop4(0, 0, 0, 4, 4))
	checkV(t, c, 0, 0, 0.5, 1, 1)

	// The prefix is consumed: a second vadd runs unprefixed.
	_, c = run(t, seed, vpfx(0, 0x00), vop3(0x18, 0, 0, 4, 5, 4), vop3(0x18, 0, 0, 4, 5, 4))
	checkV(t, c, 0, 101, 102, 103, 104)
}

func TestVFPUSingleOps(t *testing.T) {
	cases := []struct {
		name string
		rt   uint32
		in   float32
		want float32
	}{
		{"vmov", 0, 42, 42},
		{"vabs", 1, -3, 3},
		{"vneg", 2, 3, -3},
		{"vsat0", 4, 2, 1},
		{"vsat1", 5, -2, -1},
		{"vrcp", 16, 4, 0.25},
		{"vrsq", 17, 4, 0.5},
		{"vsin", 18, 1, 1},  // quarter turns: sin(90°) = 1 exactly
		{"vcos", 19, 2, -1}, // cos(180°) = -1 exactly
		{"vsqrt", 22, 16, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, c := run(t, func(m *memBus, c *CPU) { setV(c, 16, tc.in) },
				vop4(0, tc.rt, 0, 4, 1)) // op.s S000, S100
			if got := getV(c, 0); got != tc.want {
				t.Errorf("%s(%v) = %v, want %v", tc.name, tc.in, got, tc.want)
			}
		})
	}
	// vocp (VFPU9): 1-x. vfad: sum. vavg: mean. vsgn.
	_, c := run(t, func(m *memBus, c *CPU) { setV(c, 16, 0.25) }, vop4(2, 4, 0, 4, 1))
	if got := getV(c, 0); got != 0.75 {
		t.Errorf("vocp(0.25) = %v", got)
	}
	seedQ := func(m *memBus, c *CPU) { setV(c, 16, 1, 2, 3, 4) }
	_, c = run(t, seedQ, vop4(2, 6, 0, 4, 4)) // vfad.q
	if got := getV(c, 0); got != 10 {
		t.Errorf("vfad = %v", got)
	}
	_, c = run(t, seedQ, vop4(2, 7, 0, 4, 4)) // vavg.q
	if got := getV(c, 0); got != 2.5 {
		t.Errorf("vavg = %v", got)
	}
	_, c = run(t, func(m *memBus, c *CPU) { setV(c, 16, -5, 0, 7, -0.5) }, vop4(2, 10, 0, 4, 4))
	checkV(t, c, 0, -1, 0, 1, -1)
}

func TestVFPUVectorInit(t *testing.T) {
	seed := func(m *memBus, c *CPU) { setV(c, 0, 9, 9, 9, 9) }
	_, c := run(t, seed, vop4(0, 6, 0, 0, 4)) // vzero.q C000
	checkV(t, c, 0, 0, 0, 0, 0)
	_, c = run(t, seed, vop4(0, 7, 0, 0, 4)) // vone.q C000
	checkV(t, c, 0, 1, 1, 1, 1)
	_, c = run(t, seed, vop4(0, 3, 1, 0, 4)) // vidt.q C001 => (0,1,0,0)
	checkV(t, c, 4, 0, 1, 0, 0)
}

func TestVFPUConvert(t *testing.T) {
	// vi2f.s with scale 4: 32 / 2^4 = 2.0.
	_, c := run(t, func(m *memBus, c *CPU) { c.V[16] = 32 }, vop4(0x14, 4, 0, 4, 1))
	if got := getV(c, 0); got != 2.0 {
		t.Errorf("vi2f(32, 4) = %v", got)
	}
	// vf2iz.s scale 0 truncates toward zero.
	_, c = run(t, func(m *memBus, c *CPU) { setV(c, 16, -2.7) }, vop4(0x11, 0, 0, 4, 1))
	if got := int32(c.V[0]); got != -2 {
		t.Errorf("vf2iz(-2.7) = %d", got)
	}
	// vf2in.s rounds to nearest even.
	_, c = run(t, func(m *memBus, c *CPU) { setV(c, 16, 2.5) }, vop4(0x10, 0, 0, 4, 1))
	if got := int32(c.V[0]); got != 2 {
		t.Errorf("vf2in(2.5) = %d", got)
	}
	// vcst.s: constant 9 is pi.
	_, c = run(t, nil, vop4(3, 9, 0, 0, 1))
	if got := getV(c, 0); got != float32(math.Pi) {
		t.Errorf("vcst(pi) = %v", got)
	}
	// viim.s / vfim.s.
	_, c = run(t, nil, 0x37<<26|3<<24|0<<23|0<<16|100) // viim.s S000, 100
	if got := getV(c, 0); got != 100 {
		t.Errorf("viim 100 = %v", got)
	}
	_, c = run(t, nil, 0x37<<26|3<<24|1<<23|0<<16|0x3C00) // vfim.s S000, 1.0h
	if got := getV(c, 0); got != 1 {
		t.Errorf("vfim 0x3C00 = %v", got)
	}
}

func TestVFPUMatrixSetAndMove(t *testing.T) {
	_, c := run(t, nil, vmset(3, 0, 0, 4)) // vmidt.q M000
	for a := uint32(0); a < 4; a++ {
		for b := uint32(0); b < 4; b++ {
			want := float32(0)
			if a == b {
				want = 1
			}
			if got := getV(c, a*4+b); got != want {
				t.Errorf("vmidt [%d][%d] = %v", a, b, got)
			}
		}
	}
	_, c = run(t, func(m *memBus, c *CPU) {
		setV(c, 16, 1, 2, 3, 4)
		setV(c, 20, 5, 6, 7, 8)
		setV(c, 24, 9, 10, 11, 12)
		setV(c, 28, 13, 14, 15, 16)
	}, vmset(0, 0, 4, 4)) // vmmov.q M000, M100
	for i := uint32(0); i < 16; i++ {
		if got := getV(c, i); got != float32(i+1) {
			t.Errorf("vmmov flat %d = %v", i, got)
		}
	}
	_, c = run(t, func(m *memBus, c *CPU) {
		setV(c, 0, 9, 9, 9, 9)
		setV(c, 4, 9, 9, 9, 9)
	}, vmset(7, 0, 0, 2)) // vmone.p M000
	checkV(t, c, 0, 1, 1, 9, 9)
	checkV(t, c, 4, 1, 1, 9, 9)
}

func TestVFPUMmul(t *testing.T) {
	// 2x2: with standard matrices S=[[1,2],[3,4]], T=[[5,6],[7,8]] stored
	// column-major (V[mtx*16+col*4+row]), vmmul computes D = S^T x T:
	// D = [[26,30],[38,44]].
	_, c := run(t, func(m *memBus, c *CPU) {
		c.V[16], c.V[17] = f2b(1), f2b(3) // S col 0
		c.V[20], c.V[21] = f2b(2), f2b(4) // S col 1
		c.V[32], c.V[33] = f2b(5), f2b(7) // T col 0
		c.V[36], c.V[37] = f2b(6), f2b(8) // T col 1
	}, vmat(0, 0, 4, 8, 2)) // vmmul.p M000, M100, M200
	if getV(c, 0) != 26 || getV(c, 1) != 38 || getV(c, 4) != 30 || getV(c, 5) != 44 {
		t.Errorf("vmmul.p = [[%v,%v],[%v,%v]], want [[26,30],[38,44]]",
			getV(c, 0), getV(c, 4), getV(c, 1), getV(c, 5))
	}
	// With the S operand transposed (E100), D = S x T = [[19,22],[43,50]].
	_, c = run(t, func(m *memBus, c *CPU) {
		c.V[16], c.V[17] = f2b(1), f2b(3)
		c.V[20], c.V[21] = f2b(2), f2b(4)
		c.V[32], c.V[33] = f2b(5), f2b(7)
		c.V[36], c.V[37] = f2b(6), f2b(8)
	}, vmat(0, 0, 4|0x20, 8, 2)) // vmmul.p M000, E100, M200
	if getV(c, 0) != 19 || getV(c, 1) != 43 || getV(c, 4) != 22 || getV(c, 5) != 50 {
		t.Errorf("vmmul.p (E) = [[%v,%v],[%v,%v]], want [[19,22],[43,50]]",
			getV(c, 0), getV(c, 4), getV(c, 1), getV(c, 5))
	}
	// 4x4 against identity: D = I^T x T = T.
	_, c = run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 4; i++ {
			c.V[16+i*4+i] = f2b(1) // M100 = I
		}
		for i := uint32(0); i < 16; i++ {
			c.V[32+i] = f2b(float32(i + 1)) // M200
		}
	}, vmat(0, 0, 4, 8, 4)) // vmmul.q M000, M100, M200
	for i := uint32(0); i < 16; i++ {
		if got := getV(c, i); got != float32(i+1) {
			t.Errorf("vmmul.q I flat %d = %v", i, got)
		}
	}
}

func TestVFPUTfm(t *testing.T) {
	// vtfm4.q C000, M100, C200: d[i] = dot(register column i of M100, t).
	// M100 is flats 16..31; t = e0 picks out each column's first element.
	_, c := run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 16; i++ {
			c.V[16+i] = f2b(float32(i + 1))
		}
		setV(c, 32, 1, 0, 0, 0)
	}, vmat(3, 0, 4, 8, 4)) // vtfm4.q C000, M100, C200
	// d[i] = V[16+i*4+0]*1 = 1, 5, 9, 13.
	checkV(t, c, 0, 1, 5, 9, 13)

	// General t.
	_, c = run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 16; i++ {
			c.V[16+i] = f2b(float32(i + 1))
		}
		setV(c, 32, 1, 2, 3, 4)
	}, vmat(3, 0, 4, 8, 4))
	var want [4]float32
	for i := uint32(0); i < 4; i++ {
		var sum float32
		for k := uint32(0); k < 4; k++ {
			sum += float32(16+i*4+k-15) * float32(k+1)
		}
		want[i] = sum
	}
	checkV(t, c, 0, want[0], want[1], want[2], want[3])

	// vhtfm4 (encoded .t): homogeneous point transform, w wired to 1.
	_, c = run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 16; i++ {
			c.V[16+i] = f2b(float32(i + 1))
		}
		setV(c, 32, 1, 2, 3)
	}, vmat(3, 0, 4, 8, 3)) // vhtfm4.t C000, M100, C200
	for i := uint32(0); i < 4; i++ {
		var sum float32
		for k := uint32(0); k < 3; k++ {
			sum += float32(16+i*4+k-15) * float32(k+1)
		}
		sum += float32(16 + i*4 + 3 - 15) // translation column, t[3]=1
		if got := getV(c, i); got != sum {
			t.Errorf("vhtfm4 d[%d] = %v, want %v", i, got, sum)
		}
	}

	// vtfm3.t on the identity: passthrough.
	_, c = run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 3; i++ {
			c.V[16+i*4+i] = f2b(1)
		}
		setV(c, 32, 7, 8, 9)
	}, vmat(2, 0, 4, 8, 3)) // vtfm3.t C000, M100, C200
	checkV(t, c, 0, 7, 8, 9)
}

func TestVFPUMscl(t *testing.T) {
	_, c := run(t, func(m *memBus, c *CPU) {
		for i := uint32(0); i < 16; i++ {
			c.V[16+i] = f2b(float32(i + 1))
		}
		setV(c, 32, 2) // S200 = 2
	}, vmat(4, 0, 4, 8, 4)) // vmscl.q M000, M100, S200
	for i := uint32(0); i < 16; i++ {
		if got := getV(c, i); got != float32(i+1)*2 {
			t.Errorf("vmscl flat %d = %v", i, got)
		}
	}
}

func TestVFPUCrossQuat(t *testing.T) {
	// vcrsp.t: cross product e0 x e1 = e2.
	_, c := run(t, func(m *memBus, c *CPU) {
		setV(c, 16, 1, 0, 0)
		setV(c, 32, 0, 1, 0)
	}, vmat(5, 0, 4, 8, 3)) // vcrsp.t C000, C100, C200
	checkV(t, c, 0, 0, 0, 1)
	// vqmul.q: i * j = k  (quaternions x,y,z,w).
	_, c = run(t, func(m *memBus, c *CPU) {
		setV(c, 16, 1, 0, 0, 0)
		setV(c, 32, 0, 1, 0, 0)
	}, vmat(5, 0, 4, 8, 4))
	checkV(t, c, 0, 0, 0, 1, 0)
}

func TestVFPURot(t *testing.T) {
	// vrot.p C000, S100, [imm 0: cos in lane 0, sin in lane 1? imm decodes
	// sinLane=(imm>>2)&3, cosLane=imm&3] — imm=4: sin lane 1, cos lane 0.
	_, c := run(t, func(m *memBus, c *CPU) { setV(c, 16, 1) }, // 90°
		0x3C<<26|29<<21|4<<16|4<<8|0|szBits(2))
	checkV(t, c, 0, 0, 1) // cos 90° = 0, sin 90° = 1, exactly
}

func TestVFPUCmpCmovBranch(t *testing.T) {
	const t0, t1, t2 = 8, 9, 10
	// vcmp.s LT S100,S110 sets CC0; bvt takes; delay slot runs; next is skipped.
	_, c := run(t, func(m *memBus, c *CPU) {
		setV(c, 16, 1) // S100
		setV(c, 20, 2) // S110
	},
		vop3(0x1B, 0, 2, 4, 5, 1),    // vcmp.s LT (cond 2)
		0x12<<26|8<<21|0<<18|1<<16|2, // bvt 0, +2
		itype(0x09, 0, t0, 1),        // delay slot
		itype(0x09, 0, t1, 99),       // skipped
		itype(0x09, 0, t2, 7),        // branch target
	)
	if c.Reg(t0) != 1 || c.Reg(t1) != 0 || c.Reg(t2) != 7 {
		t.Errorf("bvt path: t0=%d t1=%d t2=%d", c.Reg(t0), c.Reg(t1), c.Reg(t2))
	}
	if cc := c.VfpuCtrl[vfpuCtlCC]; cc&1 == 0 {
		t.Errorf("vcmp LT did not set CC0: 0x%X", cc)
	}

	// vcmovt.q copies where the per-lane CC bits are set (imm3=6).
	_, c = run(t, func(m *memBus, c *CPU) {
		setV(c, 16, 1, 5, 3, 8) // C100
		setV(c, 20, 2, 2, 9, 9) // C110
		setV(c, 0, -1, -1, -1, -1)
	},
		vop3(0x1B, 0, 2, 4, 5, 4), // vcmp.q LT => lanes 0,2,3? (1<2, 5>2, 3<9, 8<9)
		vop4(0x15, 6, 0, 4, 4),    // vcmovt.q C000, C100, 6
	)
	checkV(t, c, 0, 1, -1, 3, 8)
}

func TestVFPUMtvMfv(t *testing.T) {
	const a, r = 4, 6
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(a, f2b(2.5))
	},
		mtv(a, 0x11), // mtv $a0, S011 (mtx0 col1 row... reg 0x11 = mtx 4? no: 0x11 = 0b0010001 => mtx 4)
		mfv(r, 0x11), // mfv $a2, same
	)
	if got := math.Float32frombits(c.Reg(r)); got != 2.5 {
		t.Errorf("mtv/mfv roundtrip = %v", got)
	}
	if got := getV(c, vfpuSingle(0x11)); got != 2.5 {
		t.Errorf("mtv register content = %v", got)
	}
	// mtvc/mfvc: control register 3 is the CC register (mask 0x3F).
	_, c = run(t, func(m *memBus, c *CPU) { c.SetReg(a, 0xFFFFFFFF) },
		mtv(a, 128+3), mfv(r, 128+3))
	if got := c.Reg(r); got != 0x3F {
		t.Errorf("mtvc/mfvc CC = 0x%X, want 0x3F", got)
	}
}

func TestVFPULvlLvr(t *testing.T) {
	const base = 8
	// lvl.q at address 0x2008 (offset word 2) fills the top of the quad from
	// descending addresses; lvr.q at 0x2008 fills the bottom from ascending.
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(base, 0x2008)
		for i := uint32(0); i < 4; i++ {
			m.w32(0x2000+i*4, f2b(float32(i+1)))
		}
		setV(c, 0, -1, -1, -1, -1)
		setV(c, 4, -1, -1, -1, -1)
	},
		0x35<<26|base<<21|0<<16|0, // lvl.q C000, 0($t0)
		0x35<<26|base<<21|1<<16|2, // lvr.q C010, 0($t0)
	)
	// lvl at word offset 2: d[3]=mem[2]=3, d[2]=mem[1]=2, d[1]=mem[0]=1; d[0] kept.
	checkV(t, c, 0, -1, 1, 2, 3)
	// lvr at word offset 2: d[0]=mem[2]=3, d[1]=mem[3]=4; d[2],d[3] kept.
	checkV(t, c, 4, 3, 4, -1, -1)
}
