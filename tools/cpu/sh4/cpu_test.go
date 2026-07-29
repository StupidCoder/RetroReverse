package sh4

import (
	"math"
	"testing"
)

// testRAM is a flat 64 KiB bus; every address folds into it, so code at the
// usual 8C010000 and an interrupt handler at VBR+0x600 coexist.
type testRAM struct{ b [65536]byte }

func (r *testRAM) at(a uint32) uint32 { return a & 0xFFFF }
func (r *testRAM) Read8(a uint32) uint8 {
	return r.b[r.at(a)]
}
func (r *testRAM) Read16(a uint32) uint16 {
	i := r.at(a)
	return uint16(r.b[i]) | uint16(r.b[i+1])<<8
}
func (r *testRAM) Read32(a uint32) uint32 {
	i := r.at(a)
	return uint32(r.b[i]) | uint32(r.b[i+1])<<8 | uint32(r.b[i+2])<<16 | uint32(r.b[i+3])<<24
}
func (r *testRAM) Write8(a uint32, v uint8) { r.b[r.at(a)] = v }
func (r *testRAM) Write16(a uint32, v uint16) {
	i := r.at(a)
	r.b[i], r.b[i+1] = uint8(v), uint8(v>>8)
}
func (r *testRAM) Write32(a uint32, v uint32) {
	i := r.at(a)
	r.b[i], r.b[i+1], r.b[i+2], r.b[i+3] = uint8(v), uint8(v>>8), uint8(v>>16), uint8(v>>24)
}

const testBase = 0x8C010000

// newTest lays halfwords at testBase and points a fresh CPU there, with
// interrupts unmasked and unblocked so tests control acceptance through what
// they assert, not through reset defaults.
func newTest(hs ...uint16) (*CPU, *testRAM) {
	ram := &testRAM{}
	c := NewCPU(ram)
	for i, h := range hs {
		ram.Write16(testBase+uint32(i)*2, h)
	}
	c.SetSR(SRMD | SRRB) // privileged, IMASK 0, BL clear
	c.SetPC(testBase)
	return c, ram
}

// run steps n instructions, failing the test on a halt: an unimplemented
// instruction in a test program is a test bug worth naming.
func run(t *testing.T, c *CPU, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		c.Step()
		if c.Halted {
			t.Fatalf("halted after %d steps: %s", i+1, c.HaltReason)
		}
	}
}

// --- encoding helpers, one per format --------------------------------------

func movImm(imm int8, n uint32) uint16  { return 0xE000 | uint16(n)<<8 | uint16(uint8(imm)) }
func addImm(imm int8, n uint32) uint16  { return 0x7000 | uint16(n)<<8 | uint16(uint8(imm)) }
func nm(op uint16, n, m uint32) uint16  { return op | uint16(n)<<8 | uint16(m)<<4 }
func one(op uint16, n uint32) uint16    { return op | uint16(n)<<8 }
func d8(op uint16, disp int8) uint16    { return op | uint16(uint8(disp)) }
func d12(op uint16, disp int16) uint16  { return op | uint16(disp)&0xFFF }
func movlPC(disp uint8, n uint32) uint16 { return 0xD000 | uint16(n)<<8 | uint16(disp) }

const (
	opNop   = 0x0009
	opRts   = 0x000B
	opClrT  = 0x0008
	opSetT  = 0x0018
	opDiv0U = 0x0019
)

func TestMovAddChain(t *testing.T) {
	c, _ := newTest(
		movImm(5, 0),   // mov #5, r0
		addImm(-2, 0),  // add #-2, r0
		nm(0x6003, 1, 0), // mov r0, r1
		addImm(10, 1),  // add #10, r1
	)
	run(t, c, 4)
	if c.R[0] != 3 || c.R[1] != 13 {
		t.Fatalf("r0=%d r1=%d, want 3 13", c.R[0], c.R[1])
	}
}

// TestDelaySlotExecutes pins the ordering: the slot after bra runs, the
// instruction after the slot does not, and control lands on the target.
func TestDelaySlotExecutes(t *testing.T) {
	c, _ := newTest(
		movImm(0, 0),   // 8C010000  mov #0, r0
		d12(0xA000, 3), // 8C010002  bra 8C01000C
		addImm(1, 0),   // 8C010004  slot: add #1, r0   (runs)
		addImm(16, 0),  // 8C010006  skipped
		opNop,          // 8C010008
		opNop,          // 8C01000A
		addImm(2, 0),   // 8C01000C  target: add #2, r0
	)
	run(t, c, 4) // mov, bra, slot, target
	if c.R[0] != 3 {
		t.Fatalf("r0=%d, want 3 (slot ran once, skipped instruction never)", c.R[0])
	}
	if c.PC != testBase+0xE {
		t.Fatalf("PC=%08X, want %08X", c.PC, uint32(testBase+0xE))
	}
}

// TestJmpTargetLatchedBeforeSlot: the slot may clobber the register the jump
// went through; the transfer uses the value at branch time.
func TestJmpTargetLatchedBeforeSlot(t *testing.T) {
	c, ram := newTest(
		movlPC(1, 0),     // 8C010000  mov.l @(4,pc), r0 -> pool at base+8
		nm(0x402B, 0, 0), // 8C010002  jmp @r0
		movImm(0, 0),     // 8C010004  slot clobbers r0
		opNop,            // 8C010006
		0, 0,             // 8C010008  pool (patched below)
		addImm(7, 1),     // 8C01000C  target
	)
	ram.Write32(testBase+8, testBase+0xC)
	run(t, c, 4) // mov.l, jmp, slot, target
	if c.R[1] != 7 {
		t.Fatalf("r1=%d, want 7 (jump target latched before slot clobbered r0)", c.R[1])
	}
	if c.R[0] != 0 {
		t.Fatalf("r0=%08X, want 0 (slot ran)", c.R[0])
	}
}

// TestBtsSlotRunsUntaken: SH-4's delayed conditionals are not branch-likely —
// the slot runs on both paths.
func TestBtsSlotRunsUntaken(t *testing.T) {
	c, _ := newTest(
		opClrT,        // T=0
		d8(0x8D00, 2), // bt/s +: not taken
		addImm(1, 0),  // slot: runs anyway
		addImm(2, 0),  // fall-through continues here
	)
	run(t, c, 4)
	if c.R[0] != 3 {
		t.Fatalf("r0=%d, want 3 (slot and fall-through both ran)", c.R[0])
	}
}

func TestAddcSubcCarryChain(t *testing.T) {
	c, _ := newTest(
		opClrT,
		nm(0x300E, 0, 1), // addc r1, r0
		nm(0x300E, 2, 3), // addc r3, r2 — consumes the carry
		nm(0x300A, 4, 5), // subc r5, r4
		nm(0x300A, 6, 7), // subc r7, r6 — consumes the borrow
	)
	c.R[0], c.R[1] = 0xFFFFFFFF, 1 // overflow: r0=0, T=1
	c.R[2], c.R[3] = 5, 0          // 5+0+T = 6
	c.R[4], c.R[5] = 0, 1          // 0-1 borrows: r4=0xFFFFFFFF, T=1
	c.R[6], c.R[7] = 5, 0          // 5-0-T = 4
	run(t, c, 5)
	if c.R[0] != 0 || c.R[2] != 6 {
		t.Fatalf("addc chain: r0=%08X r2=%d, want 0 6", c.R[0], c.R[2])
	}
	if c.R[4] != 0xFFFFFFFF || c.R[6] != 4 {
		t.Fatalf("subc chain: r4=%08X r6=%d, want FFFFFFFF 4", c.R[4], c.R[6])
	}
}

func TestDTLoop(t *testing.T) {
	c, _ := newTest(
		movImm(3, 0), // counter
		movImm(0, 1),
		one(0x4010, 0), // loop: dt r0
		addImm(1, 1),   //   body count
		d8(0x8B00, -4), //   bf loop (back to dt)
	)
	run(t, c, 2+3*3)
	if c.R[0] != 0 || c.R[1] != 3 {
		t.Fatalf("r0=%d r1=%d, want 0 3", c.R[0], c.R[1])
	}
}

// TestDiv1Unsigned runs the manual's unsigned 32÷32 sequence — div0u then 32
// rotcl/div1 pairs and a final rotcl — on values with a known quotient.
func TestDiv1Unsigned(t *testing.T) {
	prog := []uint16{opDiv0U}
	for i := 0; i < 32; i++ {
		prog = append(prog, one(0x4024, 4))  // rotcl r4 (dividend/quotient)
		prog = append(prog, nm(0x3004, 0, 5)) // div1 r5, r0 (r0 = remainder acc)
	}
	prog = append(prog, one(0x4024, 4)) // final rotcl completes the quotient
	c, _ := newTest(prog...)
	c.R[0], c.R[4], c.R[5] = 0, 100, 7
	run(t, c, len(prog))
	if c.R[4] != 14 {
		t.Fatalf("100/7: quotient r4=%d, want 14", c.R[4])
	}
}

func TestShadShld(t *testing.T) {
	c, _ := newTest(
		nm(0x400C, 0, 1), // shad r1, r0
		nm(0x400C, 2, 3), // shad r3, r2 (negative: arithmetic right)
		nm(0x400D, 4, 5), // shld r5, r4 (negative: logical right)
	)
	c.R[0], c.R[1] = 1, 4                   // 1<<4 = 16
	c.R[2], c.R[3] = 0x80000000, ^uint32(3) // -4: >> 4 arithmetic
	c.R[4], c.R[5] = 0x80000000, ^uint32(3) // -4: >> 4 logical
	run(t, c, 3)
	if c.R[0] != 16 {
		t.Fatalf("shad left: %d, want 16", c.R[0])
	}
	if c.R[2] != 0xF8000000 {
		t.Fatalf("shad right: %08X, want F8000000", c.R[2])
	}
	if c.R[4] != 0x08000000 {
		t.Fatalf("shld right: %08X, want 08000000", c.R[4])
	}
}

func TestBankSwitchOnSRWrite(t *testing.T) {
	// newTest leaves RB set (bank 1 live); the ldc clears it, swapping bank 0 in.
	c, _ := newTest(
		nm(0x400E, 1, 0), // ldc r1, sr — clears RB: banks swap
		nm(0x6003, 2, 0), // mov r0, r2 — reads the newly live bank's r0
		nm(0x0002, 3, 8), // stc r0_bank, r3 — the now-dormant bank
	)
	c.R[0] = 111     // live (bank 1) r0
	c.R[1] = SRMD    // target SR: MD set, RB clear
	c.Rbank[0] = 222 // dormant (bank 0) r0
	run(t, c, 3)
	if c.R[2] != 222 {
		t.Fatalf("after RB cleared, r0 reads %d, want 222 (bank 0)", c.R[2])
	}
	if c.R[3] != 111 {
		t.Fatalf("stc r0_bank reads %d, want 111 (the now-dormant bank 1)", c.R[3])
	}
}

func TestPCRelativeLiterals(t *testing.T) {
	c, ram := newTest(
		movlPC(1, 0), // 8C010000: mov.l @(4,pc), r0 -> pool at base+8
		0x9003,       // 8C010002: mov.w @(6,pc), r0 -> litW(base+2, 3) = base+0xC
		opNop,        // 8C010004
		opNop,        // 8C010006
		0, 0,         // 8C010008: .word pool
		0,            // 8C01000C: .hword pool
	)
	ram.Write32(testBase+8, 0x11223344)
	ram.Write16(testBase+0xC, 0x8000) // sign-extends
	run(t, c, 1)
	if c.R[0] != 0x11223344 {
		t.Fatalf("mov.l pc-literal: %08X, want 11223344", c.R[0])
	}
	run(t, c, 1)
	if c.R[0] != 0xFFFF8000 {
		t.Fatalf("mov.w pc-literal: %08X, want FFFF8000 (sign-extended)", c.R[0])
	}
}

// TestMovaAlignment pins the &^3: from an unaligned (addr%4==2) instruction,
// mova's base drops back to the aligned longword.
func TestMovaAlignment(t *testing.T) {
	c, _ := newTest(
		opNop,         // 8C010000
		d8(0xC700, 1), // 8C010002: mova @(4,pc), r0 — base (8C010006)&^3 = 8C010004, +4 = 8C010008
	)
	run(t, c, 2)
	if c.R[0] != testBase+8 {
		t.Fatalf("mova: r0=%08X, want %08X", c.R[0], uint32(testBase+8))
	}
}

func TestSQFlush(t *testing.T) {
	c, ram := newTest(
		nm(0x2002, 1, 2), // mov.l r2, @r1 — into SQ0 word 0
		addImm(4, 1),
		nm(0x2002, 1, 3), // SQ0 word 1
		one(0x0083, 4),   // pref @r4 — flush SQ0
	)
	c.R[1] = 0xE0000000
	c.R[2], c.R[3] = 0xAABBCCDD, 0x11223344
	c.R[4] = 0xE0000000
	c.QACR0 = 0x0C // external bits 28-26 = 011 -> 0x0C000000
	run(t, c, 4)
	if got := ram.Read32(0x0C000000); got != 0xAABBCCDD {
		t.Fatalf("SQ flush word 0: %08X, want AABBCCDD", got)
	}
	if got := ram.Read32(0x0C000004); got != 0x11223344 {
		t.Fatalf("SQ flush word 1: %08X, want 11223344", got)
	}
}

// TestTMUInterrupt: a running channel underflows, and the interrupt is taken
// only once IMASK admits it. Entry must switch to the other register bank and
// record INTEVT.
func TestTMUInterrupt(t *testing.T) {
	prog := make([]uint16, 0, 64)
	for i := 0; i < 40; i++ {
		prog = append(prog, opNop)
	}
	c, ram := newTest(prog...)
	ram.Write16(0x0600, opNop) // the handler: VBR=0, vector +0x600
	c.VBR = 0
	c.IPRA = 0xF000 // TMU0 at priority 15
	c.TMU.TSTR = 1
	c.TMU.Ch[0].TCOR = 2
	c.TMU.Ch[0].TCNT = 0 // underflows on the first tick, at instruction 16
	c.TMU.Ch[0].TCR = tmuUNIE // fastest prescaler, interrupt enabled

	// With IMASK=15 the underflow must NOT be accepted.
	c.SetSR(SRMD | srIMASK)
	run(t, c, 30)
	if c.PC < testBase {
		t.Fatalf("interrupt accepted through IMASK=15: PC=%08X", c.PC)
	}
	if c.TMU.Ch[0].TCR&tmuUNF == 0 {
		t.Fatalf("TMU never underflowed in 30 steps")
	}

	// Unmask: the next step must take it.
	c.SetSR(SRMD)
	c.Step()
	if c.PC != 0x0600 {
		t.Fatalf("PC=%08X, want the 0x600 vector", c.PC)
	}
	if c.INTEVT != 0x400 {
		t.Fatalf("INTEVT=%03X, want 400", c.INTEVT)
	}
	if c.SR&SRBL == 0 || c.SR&SRRB == 0 {
		t.Fatalf("SR=%08X: BL and RB must be set on entry", c.SR)
	}
}

// TestNoInterruptInDelaySlot: an interrupt arriving while a delayed transfer
// is in flight waits for the slot; SPC then points at the branch target,
// never into the slot.
func TestNoInterruptInDelaySlot(t *testing.T) {
	c, ram := newTest(
		d12(0xA000, 4), // 8C010000: bra 8C01000C
		addImm(1, 0),   // 8C010002: slot
		opNop,          // 8C010004
		opNop, opNop, opNop,
		addImm(2, 0), // 8C01000C: target
		opNop,
	)
	ram.Write16(0x0600, opNop)
	c.VBR = 0
	c.Step() // bra: pendingDelay set
	c.SetIRL(14, 0x2C0)
	c.Step() // must be the slot, not the interrupt
	if c.R[0] != 1 {
		t.Fatalf("slot did not run under a pending interrupt: r0=%d", c.R[0])
	}
	c.Step() // now the interrupt
	if c.SPC != testBase+0xC {
		t.Fatalf("SPC=%08X, want the branch target %08X", c.SPC, uint32(testBase+0xC))
	}
	if c.INTEVT != 0x2C0 {
		t.Fatalf("INTEVT=%03X, want 2C0", c.INTEVT)
	}
}

func TestTrapa(t *testing.T) {
	c, ram := newTest(
		d8(0xC300, 4), // trapa #4
	)
	ram.Write16(0x0100, opNop)
	c.VBR = 0
	c.Step()
	if c.EXPEVT != 0x160 || c.TRA != 16 {
		t.Fatalf("EXPEVT=%03X TRA=%d, want 160 16", c.EXPEVT, c.TRA)
	}
	if c.SPC != testBase+2 {
		t.Fatalf("SPC=%08X, want the next instruction %08X", c.SPC, uint32(testBase+2))
	}
	if c.PC != 0x0100 {
		t.Fatalf("PC=%08X, want the 0x100 vector", c.PC)
	}
}

// TestRte returns through SSR/SPC with the bank swapping back.
func TestRte(t *testing.T) {
	c, _ := newTest(
		0x002B, // rte
		opNop,  // slot
		opNop,
	)
	c.SSR = SRMD // RB clear: bank swaps back on restore
	c.SPC = testBase + 0x10
	c.SetSR(SRMD | SRRB)
	c.Rbank[0] = 77
	c.Step() // rte
	c.Step() // slot
	if c.PC != testBase+0x10 {
		t.Fatalf("PC=%08X, want %08X", c.PC, uint32(testBase+0x10))
	}
	if c.R[0] != 77 {
		t.Fatalf("bank did not swap back on rte: r0=%d, want 77", c.R[0])
	}
}

func TestFPUBasics(t *testing.T) {
	c, _ := newTest(
		one(0xF09D, 0),   // fldi1 fr0
		one(0xF09D, 1),   // fldi1 fr1
		nm(0xF000, 0, 1), // fadd fr1, fr0 -> 2
		one(0xF03D, 0),   // ftrc fr0, fpul -> 2
	)
	run(t, c, 4)
	if got := math.Float32frombits(c.fpr[0][0]); got != 2 {
		t.Fatalf("fadd: fr0=%v, want 2", got)
	}
	if c.FPUL != 2 {
		t.Fatalf("ftrc: FPUL=%d, want 2", c.FPUL)
	}
}

func TestFPUFloatAndSca(t *testing.T) {
	c, _ := newTest(
		one(0xF02D, 2), // float fpul, fr2
		one(0xF0FD, 4), // fsca fpul, dr4
	)
	c.FPUL = 3
	run(t, c, 1)
	if got := math.Float32frombits(c.fpr[0][2]); got != 3 {
		t.Fatalf("float: fr2=%v, want 3", got)
	}
	c.FPUL = 0 // angle 0: sin 0, cos 1
	run(t, c, 1)
	if s := math.Float32frombits(c.fpr[0][4]); s != 0 {
		t.Fatalf("fsca sin(0)=%v, want 0", s)
	}
	if co := math.Float32frombits(c.fpr[0][5]); co != 1 {
		t.Fatalf("fsca cos(0)=%v, want 1", co)
	}
}

func TestFrchgBanks(t *testing.T) {
	c, _ := newTest(
		one(0xF09D, 0), // fldi1 fr0 (bank 0)
		0xFBFD,         // frchg
		one(0xF08D, 0), // fldi0 fr0 (bank 1)
	)
	run(t, c, 3)
	if math.Float32frombits(c.fpr[0][0]) != 1 || c.fpr[1][0] != 0 {
		t.Fatalf("banks: fpr0[0]=%08X fpr1[0]=%08X, want 1.0 and 0", c.fpr[0][0], c.fpr[1][0])
	}
	if c.FPSCR&FPSCRFR == 0 {
		t.Fatalf("FPSCR.FR not flipped")
	}
}

// TestFschgFmov64 moves a DR pair through memory with SZ set: high word at
// the lower address.
func TestFschgFmov64(t *testing.T) {
	c, ram := newTest(
		0xF3FD,           // fschg: SZ=1
		nm(0xF008, 0, 1), // fmov @r1, dr0
		nm(0xF00A, 2, 0), // fmov dr0, @r2
	)
	c.R[1] = 0x8C012000
	c.R[2] = 0x8C013000
	ram.Write32(0x8C012000, 0x40080000) // 3.0's high word
	ram.Write32(0x8C012004, 0x00000000)
	run(t, c, 3)
	if got := getDR(&c.fpr[0], 0); got != 3 {
		t.Fatalf("fmov 64 load: dr0=%v, want 3", got)
	}
	if hi, lo := ram.Read32(0x8C013000), ram.Read32(0x8C013004); hi != 0x40080000 || lo != 0 {
		t.Fatalf("fmov 64 store: %08X %08X, want 40080000 0", hi, lo)
	}
}

func TestFPUDoublePR(t *testing.T) {
	c, _ := newTest(
		nm(0xF000, 0, 2), // fadd dr2, dr0 (PR mode)
	)
	c.SetFPSCR(FPSCRPR)
	setDR(&c.fpr[0], 0, 1.5)
	setDR(&c.fpr[0], 2, 2.25)
	run(t, c, 1)
	if got := getDR(&c.fpr[0], 0); got != 3.75 {
		t.Fatalf("PR fadd: dr0=%v, want 3.75", got)
	}
}

func TestMacWSaturation(t *testing.T) {
	c, ram := newTest(
		nm(0x400F, 1, 0), // mac.w @r0+, @r1+
	)
	c.SetSR(c.SR | SRS)
	c.R[0], c.R[1] = 0x8C012000, 0x8C012010
	ram.Write16(0x8C012000, 0x7FFF)
	ram.Write16(0x8C012010, 0x7FFF)
	c.MACL = 0x7FFFFFFF
	run(t, c, 1)
	if c.MACL != 0x7FFFFFFF || c.MACH&1 == 0 {
		t.Fatalf("mac.w saturation: MACL=%08X MACH=%08X, want 7FFFFFFF and MACH bit0", c.MACL, c.MACH)
	}
}

func TestSnapshotRoundTripMidDelaySlot(t *testing.T) {
	c, _ := newTest(
		d12(0xA000, 4), // bra
		addImm(1, 0),   // slot
		opNop, opNop, opNop, opNop,
		addImm(2, 0), // target
	)
	c.Step() // bra executed: pendingDelay latched
	s := c.Snapshot()

	c2 := NewCPU(&testRAM{})
	// Same program in the second RAM.
	ram2 := c2.bus.(*testRAM)
	for i, h := range []uint16{d12(0xA000, 4), addImm(1, 0), opNop, opNop, opNop, opNop, addImm(2, 0)} {
		ram2.Write16(testBase+uint32(i)*2, h)
	}
	c2.Restore(s)
	c2.Step() // slot
	c2.Step() // target
	if c2.R[0] != 3 {
		t.Fatalf("restored mid-delay-slot: r0=%d, want 3", c2.R[0])
	}
}

// TestSnapshotMapsIndependent: the on-chip maps must be deep-copied both
// directions.
func TestSnapshotMapsIndependent(t *testing.T) {
	c, _ := newTest(opNop)
	c.onchipRaw[0xFF800000] = 1
	s := c.Snapshot()
	c.onchipRaw[0xFF800000] = 2
	if s.OnchipRaw[0xFF800000] != 1 {
		t.Fatalf("snapshot map aliases the live map")
	}
	c2 := NewCPU(&testRAM{})
	c2.Restore(s)
	c2.onchipRaw[0xFF800000] = 3
	if s.OnchipRaw[0xFF800000] != 1 {
		t.Fatalf("restore aliased the state map")
	}
}
