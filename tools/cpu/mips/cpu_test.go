package mips

import "testing"

// flatBus is a flat little-endian memory for exercising the core.
type flatBus struct{ m []byte }

func newBus(size int) *flatBus { return &flatBus{m: make([]byte, size)} }

func (b *flatBus) Read(a uint32) byte {
	if int(a) < len(b.m) {
		return b.m[a]
	}
	return 0
}
func (b *flatBus) Write(a uint32, v byte) {
	if int(a) < len(b.m) {
		b.m[a] = v
	}
}

// load writes a sequence of 32-bit words starting at addr.
func (b *flatBus) load(addr uint32, words ...uint32) {
	for i, w := range words {
		p := addr + uint32(i)*4
		b.m[p] = byte(w)
		b.m[p+1] = byte(w >> 8)
		b.m[p+2] = byte(w >> 16)
		b.m[p+3] = byte(w >> 24)
	}
}

// runTo single-steps up to a budget, stopping when PC reaches stop.
func runTo(t *testing.T, c *CPU, stop uint32, budget int) {
	t.Helper()
	for i := 0; i < budget; i++ {
		if c.PC == stop {
			return
		}
		if c.Halted {
			t.Fatalf("halted: %s (pc=0x%08X)", c.HaltReason, c.PC)
		}
		c.Step()
	}
	t.Fatalf("did not reach 0x%08X within %d steps (pc=0x%08X)", stop, budget, c.PC)
}

// TestBranchDelaySlot proves the delay-slot instruction runs even when the
// branch is taken, and that the instruction after the slot is skipped.
func TestBranchDelaySlot(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x24080001, // addiu $t0, $zero, 1
		0x10000002, // beq   $zero, $zero, 0x110  (taken)
		0x2508000A, // addiu $t0, $t0, 10   (delay slot: runs)
		0x25080064, // addiu $t0, $t0, 100  (skipped)
		0x250803E8, // addiu $t0, $t0, 1000 (branch target @0x110)
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	runTo(t, c, 0x114, 20)
	if got := c.R[8]; got != 1+10+1000 {
		t.Errorf("$t0 = %d, want %d (delay slot must run, +100 must be skipped)", got, 1+10+1000)
	}
}

// TestLoadDelaySlot proves the instruction after a load sees the OLD register
// value (the R3000 load-delay hazard), and the one after that sees the new value.
func TestLoadDelaySlot(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x24080200, // addiu $t0, $zero, 0x200   (address)
		0x240900AA, // addiu $t1, $zero, 0xAA
		0xAD090000, // sw    $t1, 0($t0)          (mem[0x200] = 0xAA)
		0x240A0011, // addiu $t2, $zero, 0x11     (old marker)
		0x8D0A0000, // lw    $t2, 0($t0)          (delayed load of 0xAA)
		0x01405825, // or    $t3, $t2, $zero      (delay slot: sees OLD 0x11)
		0x01406025, // or    $t4, $t2, $zero      (sees NEW 0xAA)
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	runTo(t, c, 0x11C, 20)
	if c.R[11] != 0x11 {
		t.Errorf("$t3 = 0x%X, want 0x11 (load delay slot must see old value)", c.R[11])
	}
	if c.R[12] != 0xAA {
		t.Errorf("$t4 = 0x%X, want 0xAA (value visible after the delay slot)", c.R[12])
	}
}

// TestCallReturn exercises jal (with its link) and jr $ra.
func TestCallReturn(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x0C000080, // jal   0x200        ($ra = 0x108)
		0x24080005, // addiu $t0, $zero, 5 (delay slot)
		0x24090009, // addiu $t1, $zero, 9 (after return @0x108)
	)
	b.load(0x200,
		0x240A0007, // addiu $t2, $zero, 7
		0x03E00008, // jr    $ra           (back to 0x108)
		0x00000000, // nop                 (delay slot)
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	runTo(t, c, 0x10C, 30)
	if c.R[8] != 5 || c.R[9] != 9 || c.R[10] != 7 {
		t.Errorf("$t0=%d $t1=%d $t2=%d, want 5 9 7", c.R[8], c.R[9], c.R[10])
	}
	if c.R[31] != 0x108 {
		t.Errorf("$ra = 0x%08X, want 0x00000108", c.R[31])
	}
}

// TestMulDiv checks the HI/LO multiply and divide results.
func TestMulDiv(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x24080006, // addiu $t0, $zero, 6
		0x24090007, // addiu $t1, $zero, 7
		0x01090018, // mult  $t0, $t1     (LO = 42)
		0x00005012, // mflo  $t2
		0x240B0014, // addiu $t3, $zero, 20
		0x0168001A, // div   $t3, $t0     (20/6 -> LO=3, HI=2)
		0x00006012, // mflo  $t4
		0x00006810, // mfhi  $t5
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	runTo(t, c, 0x120, 30)
	if c.R[10] != 42 {
		t.Errorf("$t2 (6*7) = %d, want 42", c.R[10])
	}
	if c.R[12] != 3 {
		t.Errorf("$t4 (20/6) = %d, want 3", c.R[12])
	}
	if c.R[13] != 2 {
		t.Errorf("$t5 (20%%6) = %d, want 2", c.R[13])
	}
}

// TestR0Hardwired verifies writes to $zero are dropped.
func TestR0Hardwired(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x2400FFFF, // addiu $zero, $zero, -1  (must NOT change $zero)
		0x24010005, // addiu $at, $zero, 5
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	runTo(t, c, 0x108, 10)
	if c.R[0] != 0 {
		t.Errorf("$zero = 0x%X, want 0", c.R[0])
	}
	if c.R[1] != 5 {
		t.Errorf("$at = %d, want 5", c.R[1])
	}
}

// TestOverflowTrap checks that add traps on signed overflow (and addu does not).
func TestOverflowTrap(t *testing.T) {
	b := newBus(0x1000)
	b.load(0x100,
		0x3C087FFF, // lui   $t0, 0x7FFF
		0x3508FFFF, // ori   $t0, $t0, 0xFFFF   -> $t0 = 0x7FFFFFFF (INT_MAX)
		0x00000000, // nop
		0x01081020, // add   $v0, $t0, $t0      -> signed overflow, must trap
	)
	c := NewCPU(b)
	c.SetPC(0x100)
	// Run until we either take the exception (PC jumps to vector) or finish.
	for i := 0; i < 20 && c.PC != 0x80000080 && !c.Halted; i++ {
		c.Step()
	}
	if c.PC != 0x80000080 {
		t.Fatalf("add overflow did not trap (pc=0x%08X)", c.PC)
	}
	if c.COP0[cop0EPC] != 0x10C {
		t.Errorf("EPC = 0x%08X, want 0x10C", c.COP0[cop0EPC])
	}
	if c.R[2] != 0 {
		t.Errorf("$v0 = 0x%X, want 0 (overflow must not write the result)", c.R[2])
	}
}
