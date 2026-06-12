package m68k

import "testing"

type mem struct{ b []byte }

func (m *mem) Read(a uint32) byte     { return m.b[a&0xFFFF] }
func (m *mem) Write(a uint32, v byte) { m.b[a&0xFFFF] = v }

func newCPU(words ...uint16) (*CPU, *mem) {
	m := &mem{b: make([]byte, 0x10000)}
	for i, w := range words {
		m.b[i*2] = byte(w >> 8)
		m.b[i*2+1] = byte(w)
	}
	return NewCPU(m), m
}

// run single-steps until PC reaches stop, the CPU halts, or the step budget runs out.
func run(t *testing.T, c *CPU, stop uint32) {
	t.Helper()
	for i := 0; i < 1000; i++ {
		if c.PC == stop {
			return
		}
		if c.Halted {
			t.Fatalf("halted: %s", c.HaltReason)
		}
		c.Step()
	}
	t.Fatalf("did not reach $%X (pc=$%X)", stop, c.PC)
}

func TestFlagMath(t *testing.T) {
	c, _ := newCPU()
	if r := c.add(0x7F, 0x01, 0); r != 0x80 || !c.V || c.C || !c.N {
		t.Errorf("add 7F+01.b = %X V=%v C=%v N=%v, want 80 V C̄ N", r, c.V, c.C, c.N)
	}
	if r := c.add(0xFF, 0x01, 0); r != 0x00 || !c.C || !c.Z || !c.X {
		t.Errorf("add FF+01.b = %X C=%v Z=%v X=%v, want 00 C Z X", r, c.C, c.Z, c.X)
	}
	if r := c.sub(0x00, 0x01, 1); r != 0xFFFF || !c.C || !c.N || c.Z {
		t.Errorf("sub 0-1.w = %X C=%v N=%v, want FFFF C N", r, c.C, c.N)
	}
	c.X = false
	c.cmp(0x10, 0x10, 1) // equal
	if !c.Z || c.C || c.X {
		t.Errorf("cmp equal: Z=%v C=%v X=%v, want Z C̄ X̄", c.Z, c.C, c.X)
	}
}

// TestSumLoop runs a real loop: D0 = 10+9+...+0 = 55 via ADD/DBRA.
func TestSumLoop(t *testing.T) {
	c, _ := newCPU(
		0x7000,         // MOVEQ #0,d0
		0x720A,         // MOVEQ #10,d1
		0xD081,         // loop: ADD.l d1,d0
		0x51C9, 0xFFFC, // DBRA d1,loop
	)
	run(t, c, 10)
	if c.D[0] != 55 {
		t.Errorf("sum = %d, want 55", c.D[0])
	}
	if int16(c.D[1]) != -1 {
		t.Errorf("d1 = %d, want -1 after DBRA exit", int16(c.D[1]))
	}
}

// TestCallReturn exercises BSR/RTS and the stack pointer.
func TestCallReturn(t *testing.T) {
	c, _ := newCPU(
		0x6104, // BSR.S +4 -> $6
		0x4E71, // NOP (return lands here, $2)
		0x4E71, // NOP ($4)
		0x7007, // sub ($6): MOVEQ #7,d0
		0x4E75, // RTS
	)
	c.A[7] = 0x1000
	run(t, c, 4)
	if c.D[0] != 7 {
		t.Errorf("d0 = %d, want 7 (sub ran)", c.D[0])
	}
	if c.A[7] != 0x1000 {
		t.Errorf("sp = $%X, want $1000 (restored)", c.A[7])
	}
}

func TestShifts(t *testing.T) {
	c, _ := newCPU()
	c.D[0] = 0x80000000
	if r := c.doShift(c.D[0], 1, 2, 1, true); r != 0 || !c.C || !c.X || !c.Z { // LSL.l #1
		t.Errorf("LSL 80000000 = %X C=%v X=%v Z=%v", r, c.C, c.X, c.Z)
	}
	c.D[0] = 1
	if r := c.doShift(c.D[0], 1, 2, 1, false); r != 0 || !c.C || !c.Z { // LSR.l #1
		t.Errorf("LSR 1 = %X C=%v Z=%v", r, c.C, c.Z)
	}
	c.D[0] = 0xFFFFFFFE
	if r := c.doShift(c.D[0], 1, 2, 0, false); r != 0xFFFFFFFF || c.C || !c.N { // ASR.l #1
		t.Errorf("ASR -2 = %X C=%v N=%v", r, c.C, c.N)
	}
}

// TestMovem pushes two registers and pops them into two others.
func TestMovem(t *testing.T) {
	c, _ := newCPU(
		0x48E7, 0xC000, // MOVEM.L d0/d1,-(a7)
		0x4CDF, 0x000C, // MOVEM.L (a7)+,d2/d3
	)
	c.A[7] = 0x1000
	c.D[0] = 0x11111111
	c.D[1] = 0x22222222
	run(t, c, 8)
	if c.D[2] != 0x11111111 || c.D[3] != 0x22222222 {
		t.Errorf("movem round-trip: d2=%X d3=%X, want 11111111 22222222", c.D[2], c.D[3])
	}
	if c.A[7] != 0x1000 {
		t.Errorf("sp = $%X, want $1000 (balanced)", c.A[7])
	}
}

func TestMoveSetsFlags(t *testing.T) {
	c, _ := newCPU(0x7005, 0x2200) // MOVEQ #5,d0 ; MOVE.L d0,d1
	run(t, c, 4)
	if c.D[1] != 5 || c.Z || c.N {
		t.Errorf("move: d1=%X Z=%v N=%v, want 5 Z̄ N̄", c.D[1], c.Z, c.N)
	}
}
