package allegrex

// Hand-assembled loop tests for the Allegrex-specific surface (MIPS32R2 integer,
// the single-precision FPU, and the VFPU single load/store). The MIPS I integer
// core is inherited verbatim from tools/cpu/mips, whose SingleStepTests suite covers
// it; these tests exercise what allegrex adds.

import (
	"math"
	"testing"
)

// memBus is a flat little-endian byte array addressed by the low 20 bits.
type memBus [0x100000]byte

func (m *memBus) Read(a uint32) byte     { return m[a&0xFFFFF] }
func (m *memBus) Write(a uint32, v byte) { m[a&0xFFFFF] = v }

func (m *memBus) w32(a, v uint32) {
	m[a&0xFFFFF] = byte(v)
	m[(a+1)&0xFFFFF] = byte(v >> 8)
	m[(a+2)&0xFFFFF] = byte(v >> 16)
	m[(a+3)&0xFFFFF] = byte(v >> 24)
}
func (m *memBus) r32(a uint32) uint32 {
	return uint32(m[a&0xFFFFF]) | uint32(m[(a+1)&0xFFFFF])<<8 |
		uint32(m[(a+2)&0xFFFFF])<<16 | uint32(m[(a+3)&0xFFFFF])<<24
}

// encoding helpers
func rtype(op, rs, rt, rd, sa, funct uint32) uint32 {
	return op<<26 | rs<<21 | rt<<16 | rd<<11 | sa<<6 | funct
}
func itype(op, rs, rt, imm uint32) uint32 { return op<<26 | rs<<21 | rt<<16 | imm&0xFFFF }

// run loads prog at 0x1000, runs len(prog) steps, and returns the CPU.
func run(t *testing.T, setup func(m *memBus, c *CPU), prog ...uint32) (*memBus, *CPU) {
	t.Helper()
	m := &memBus{}
	for i, w := range prog {
		m.w32(0x1000+uint32(i)*4, w)
	}
	c := NewCPU(m)
	c.SetPC(0x1000)
	if setup != nil {
		setup(m, c)
	}
	// Two extra steps run trailing nops (zeroed memory) so a delayed load issued by
	// the final instruction commits before we read the registers.
	for i := 0; i < len(prog)+2; i++ {
		c.Step()
		if c.Halted {
			t.Fatalf("halted at step %d: %s", i, c.HaltReason)
		}
	}
	return m, c
}

func TestR2Integer(t *testing.T) {
	const t0, t1 = 8, 9
	_, c := run(t, func(m *memBus, c *CPU) {},
		itype(0x0F, 0, t0, 0x1234),         // lui  $t0, 0x1234
		itype(0x0D, t0, t0, 0x5678),        // ori  $t0, $t0, 0x5678  => 0x12345678
		rtype(0x1F, 0, t0, t1, 0x10, 0x20), // seb  $t1, $t0  => 0x00000078
	)
	if got := c.Reg(t0); got != 0x12345678 {
		t.Fatalf("t0 = 0x%08X", got)
	}
	if got := c.Reg(t1); got != 0x00000078 {
		t.Errorf("seb = 0x%08X, want 0x78", got)
	}
}

func TestR2Ops(t *testing.T) {
	const s = 8 // source reg loaded with 0x12345678
	load := func() []uint32 {
		return []uint32{itype(0x0F, 0, s, 0x1234), itype(0x0D, s, s, 0x5678)}
	}
	cases := []struct {
		name string
		inst uint32
		rd   uint32
		want uint32
	}{
		{"seh", rtype(0x1F, 0, s, 10, 0x18, 0x20), 10, 0x00005678},
		{"wsbh", rtype(0x1F, 0, s, 10, 0x02, 0x20), 10, 0x34127856},
		{"ext_4_8", rtype(0x1F, s, 10, 8-1, 4, 0x00), 10, 0x67}, // rd=size-1, sa=pos
		{"clz", rtype(0x1C, s, 10, 10, 0, 0x20), 10, 3},
		{"mul", rtype(0x1C, s, s, 10, 0, 0x02), 10, 0x1DF4D840}, // 0x12345678^2 low32
		{"rotr8", rtype(0x00, 1, s, 10, 8, 0x02), 10, 0x78123456},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog := append(load(), tc.inst)
			_, c := run(t, nil, prog...)
			if got := c.Reg(tc.rd); got != tc.want {
				t.Errorf("%s = 0x%08X, want 0x%08X", tc.name, got, tc.want)
			}
		})
	}
}

func TestMovnMovz(t *testing.T) {
	const src, cond, dz, dn = 8, 9, 10, 11
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(src, 0xABCD)
		c.SetReg(cond, 0) // zero
	},
		rtype(0x00, src, cond, dz, 0, 0x0A), // movz $dz, $src, $cond  (cond==0 => moves)
		rtype(0x00, src, cond, dn, 0, 0x0B), // movn $dn, $src, $cond  (cond==0 => no move)
	)
	if c.Reg(dz) != 0xABCD {
		t.Errorf("movz did not move: 0x%08X", c.Reg(dz))
	}
	if c.Reg(dn) != 0 {
		t.Errorf("movn moved on zero: 0x%08X", c.Reg(dn))
	}
}

func TestFPUArith(t *testing.T) {
	const a, b, r = 4, 5, 6 // gprs
	const f0, f1, f2 = 0, 1, 2
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(a, math.Float32bits(1.5))
		c.SetReg(b, math.Float32bits(2.0))
	},
		rtype(0x11, 0x04, a, f0, 0, 0),      // mtc1 $a, $f0
		rtype(0x11, 0x04, b, f1, 0, 0),      // mtc1 $b, $f1
		rtype(0x11, 0x10, f1, f0, f2, 0x00), // add.s $f2, $f0, $f1  (ft=f1,fs=f0,fd=f2)
		rtype(0x11, 0x00, r, f2, 0, 0),      // mfc1 $r, $f2
	)
	if got := math.Float32frombits(c.Reg(r)); got != 3.5 {
		t.Errorf("1.5 + 2.0 = %v, want 3.5", got)
	}
}

func TestFPUConvertAndCompare(t *testing.T) {
	const a, r = 4, 6
	const f0, f1 = 0, 1
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(a, 7) // integer 7 in an FPU reg via mtc1
	},
		rtype(0x11, 0x04, a, f0, 0, 0),     // mtc1 $a, $f0  (bits = int 7)
		rtype(0x11, 0x14, 0, f0, f1, 0x20), // cvt.s.w $f1, $f0  => 7.0
		rtype(0x11, 0x00, r, f1, 0, 0),     // mfc1 $r, $f1
	)
	if got := math.Float32frombits(c.Reg(r)); got != 7.0 {
		t.Errorf("cvt.s.w 7 = %v, want 7.0", got)
	}
}

func TestFPUCompareConditions(t *testing.T) {
	// The c.cond.s predicate: cond bit 2 = less, bit 1 = equal, bit 0 = unordered.
	nan := float32(math.NaN())
	cases := []struct {
		funct uint32
		a, b  float32
		want  bool
	}{
		{0x32, 1.0, 1.0, true},  // c.eq.s equal
		{0x32, 1.0, 2.0, false}, // c.eq.s not equal
		{0x32, nan, 1.0, false}, // c.eq.s unordered
		{0x3C, 1.0, 2.0, true},  // c.lt.s less
		{0x3C, 2.0, 1.0, false}, // c.lt.s greater
		{0x3C, 1.0, 1.0, false}, // c.lt.s equal
		{0x3E, 1.0, 1.0, true},  // c.le.s equal
		{0x3E, 1.0, 2.0, true},  // c.le.s less
		{0x3E, 2.0, 1.0, false}, // c.le.s greater
		{0x34, 1.0, 2.0, true},  // c.olt.s less
		{0x35, nan, 1.0, true},  // c.ult.s unordered
		{0x30, 1.0, 1.0, false}, // c.f.s always false
		{0x31, nan, 1.0, true},  // c.un.s unordered
		{0x31, 1.0, 1.0, false}, // c.un.s ordered
	}
	for _, tc := range cases {
		if got := fcompare(tc.funct, tc.a, tc.b); got != tc.want {
			t.Errorf("fcompare(0x%02X, %v, %v) = %v, want %v", tc.funct, tc.a, tc.b, got, tc.want)
		}
	}
}

func TestFPUCompareBranch(t *testing.T) {
	const a, b, r = 4, 5, 6
	const f0, f1 = 0, 1
	// c.eq.s on equal values sets FCC; bc1t takes the branch.
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(a, math.Float32bits(1.0))
		c.SetReg(b, math.Float32bits(1.0))
	},
		rtype(0x11, 0x04, a, f0, 0, 0),     // mtc1 $a, $f0
		rtype(0x11, 0x04, b, f1, 0, 0),     // mtc1 $b, $f1
		rtype(0x11, 0x10, f1, f0, 0, 0x32), // c.eq.s $f0, $f1
		itype(0x11, 0x08, 0x01, 2),         // bc1t +2
		itype(0x09, 0, r, 99),              // delay slot: addiu $r,$0,99
		itype(0x09, 0, r, 1),               // skipped when branch taken
	)
	if c.Reg(r) != 99 {
		t.Errorf("bc1t after c.eq.s(1,1): r=%d, want 99 (delay slot then skip)", c.Reg(r))
	}
}

func TestVFPUSingleLoadStore(t *testing.T) {
	const base = 8
	_, c := run(t, func(m *memBus, c *CPU) {
		c.SetReg(base, 0x2000)
		m.w32(0x2000, 0xDEADBEEF)
	},
		itype(0x32, base, 0, 0),  // lv.s  $v0, 0($t0)   (vt=0)
		itype(0x3A, base, 0, 16), // sv.s  $v0, 16($t0)
	)
	if got := c.V[0]; got != 0xDEADBEEF {
		t.Errorf("VFPU v0 = 0x%08X after lv.s", got)
	}
	m := (*memBus)(nil)
	_ = m
}

func TestBranchDelaySlot(t *testing.T) {
	const t0, t1, t2 = 8, 9, 10
	// beq taken executes its delay slot, then jumps past the following instruction.
	_, c := run(t, func(m *memBus, c *CPU) {},
		itype(0x09, 0, t0, 5),  // addiu $t0, $0, 5
		itype(0x04, t0, t0, 2), // beq $t0,$t0,+2  (taken; target = 0x100C+8)
		itype(0x09, 0, t1, 99), // delay slot: addiu $t1,$0,99  (executes)
		itype(0x09, 0, t2, 7),  // skipped (branch lands past it)
	)
	if c.Reg(t1) != 99 {
		t.Errorf("delay slot did not execute: t1=%d", c.Reg(t1))
	}
	if c.Reg(t2) != 0 {
		t.Errorf("branch did not skip target+1: t2=%d", c.Reg(t2))
	}
}

func TestBranchLikelyNullify(t *testing.T) {
	const t0, t1 = 8, 9
	// bnel not taken nullifies its delay slot.
	_, c := run(t, func(m *memBus, c *CPU) {},
		itype(0x09, 0, t0, 5),  // addiu $t0,$0,5
		itype(0x15, t0, t0, 2), // bnel $t0,$t0,+2 (t0==t0 => NOT taken => nullify next)
		itype(0x09, 0, t1, 42), // nullified delay slot (must NOT execute)
	)
	if c.Reg(t1) != 0 {
		t.Errorf("bnel not-taken did not nullify delay slot: t1=%d", c.Reg(t1))
	}
}
