package gcdsp

import "testing"

// nullBus is a hardware bus that is never expected to be touched in these tests; if the program
// reaches a hardware address the test fails loudly rather than reading a silent zero.
type nullBus struct{ t *testing.T }

func (b nullBus) HWRead(a uint16) uint16  { b.t.Fatalf("unexpected HW read @0x%04X", a); return 0 }
func (b nullBus) HWWrite(a uint16, v uint16) { b.t.Fatalf("unexpected HW write @0x%04X = 0x%04X", a, v) }

// run loads a program into IRAM and steps until PC reaches stopAt, the core halts, or the step
// budget is spent. It returns the CPU for inspection.
func run(t *testing.T, prog []uint16, stopAt uint16) *CPU {
	c := New(nullBus{t})
	copy(c.IRAM[:], prog)
	for i := 0; i < 100000; i++ {
		if c.PC == stopAt {
			return c
		}
		if !c.Step() {
			break
		}
	}
	if c.Halted {
		t.Fatalf("core halted: %s (PC=0x%04X)", c.Reason, c.PC)
	}
	if c.PC != stopAt {
		t.Fatalf("did not reach 0x%04X (stopped at 0x%04X)", stopAt, c.PC)
	}
	return c
}

// TestInitIdiom runs the register-init idiom the real ucode opens with: load 0xFFFF into an
// accumulator middle word and copy it into the wrapping registers. It checks the value moves
// and that writing acX.m sign-extends the high word.
func TestInitIdiom(t *testing.T) {
	c := run(t, []uint16{
		0x009E, 0xFFFF, // 0000: lri ac0.m, #0xFFFF
		0x1D1E,         // 0002: mrr wr0, ac0.m
		0x1D3E,         // 0003: mrr wr1, ac0.m
		0x0021,         // 0004: halt (sentinel; run stops before executing it)
	}, 0x0004)
	if c.Reg[regWR0] != 0xFFFF || c.Reg[regWR0+1] != 0xFFFF {
		t.Errorf("wr0/wr1 = 0x%04X/0x%04X, want 0xFFFF", c.Reg[regWR0], c.Reg[regWR0+1])
	}
	if c.Reg[regAC0H] != 0xFFFF {
		t.Errorf("ac0.h = 0x%04X, want 0xFFFF (sign-extended from ac0.m)", c.Reg[regAC0H])
	}
}

// TestBlockLoop runs a block loop that stores a counter to consecutive DRAM words, stepping the
// address register and incrementing the accumulator each pass — the shape of the ucode's DRAM
// clear and table fills.
func TestBlockLoop(t *testing.T) {
	c := run(t, []uint16{
		0x0088, 0xFFFF, // 0000: lri wr0, #0xFFFF (no-wrap, as the ucode sets up)
		0x0080, 0x0100, // 0002: lri ar0, #0x0100
		0x009E, 0x0000, // 0004: lri ac0.m, #0x0000
		0x1104, 0x000A, // 0006: bloopi #4, 0x000A  (body 0x0008..0x000A)
		0x1B1E,         // 0008: srr @ar0, ac0.m
		0x0008,         // 0009: iar ar0
		0x0200, 0x0001, // 000A: addi ac0, #0x0001  (loop end)
		0x0021,         // 000C: halt (sentinel)
	}, 0x000C)
	for i := uint16(0); i < 4; i++ {
		if got := c.DRAM[0x100+i]; got != i {
			t.Errorf("DRAM[0x%03X] = 0x%04X, want 0x%04X", 0x100+i, got, i)
		}
	}
	if c.Reg[regAC0M] != 4 {
		t.Errorf("ac0.m = 0x%04X, want 4", c.Reg[regAC0M])
	}
}

// TestCompareBranch checks that cmpi sets the zero flag on equality and jmpz takes the branch,
// skipping the halt that would otherwise stop the run.
func TestCompareBranch(t *testing.T) {
	c := run(t, []uint16{
		0x009E, 0x0004, // 0000: lri ac0.m, #0x0004
		0x0280, 0x0004, // 0002: cmpi ac0, #0x0004  (ac0 == 4<<16, equal)
		0x0295, 0x0008, // 0004: jmpz 0x0008
		0x0021,         // 0006: halt (must be skipped)
		0x0021,         // 0007: pad
		0x0000,         // 0008: nop (sentinel)
	}, 0x0008)
	if c.Reg[regSR]&srZero == 0 {
		t.Error("zero flag not set after equal compare")
	}
}

// TestAndfAndcfFlags pins the sibling logic-test ops that the mailbox handshake turns on.
// andf sets the logic-zero flag when NONE of the immediate's bits are present in acD.m; andcf
// sets it when ALL of them are. Reading andcf as an OR (the old bug) never sets the flag for a
// present bit, so the microcode's "wait until the mailbox present bit is set" loop never exits.
func TestAndfAndcfFlags(t *testing.T) {
	cases := []struct {
		name    string
		op      uint16 // 0x02A0 andf, 0x02C0 andcf
		acm     uint16
		imm     uint16
		wantLZ  bool
	}{
		{"andf/bit-clear-sets-LZ", 0x02A0, 0x0000, 0x8000, true},
		{"andf/bit-set-clears-LZ", 0x02A0, 0x8000, 0x8000, false},
		{"andcf/bit-set-sets-LZ", 0x02C0, 0x8000, 0x8000, true},
		{"andcf/bit-clear-clears-LZ", 0x02C0, 0x0000, 0x8000, false},
		{"andcf/partial-clears-LZ", 0x02C0, 0x8000, 0xC000, false},
	}
	for _, tc := range cases {
		c := run(t, []uint16{
			0x009E, tc.acm, // 0000: lri ac0.m, #acm
			tc.op, tc.imm, //  0002: andf/andcf ac0.m, #imm
			0x0000, //         0004: nop (sentinel)
		}, 0x0004)
		if gotLZ := c.Reg[regSR]&srLogicZero != 0; gotLZ != tc.wantLZ {
			t.Errorf("%s: logic-zero=%v, want %v", tc.name, gotLZ, tc.wantLZ)
		}
	}
}

// TestMailboxWaitIdiom runs the exact wait-for-CPU-mail loop the AX microcode boots into: read
// the from-CPU mailbox high half, andcf its present bit, and loop while it is clear. With the
// mailbox reporting a present bit the loop must exit and consume the low half; the old orf
// reading of andcf spun here forever, which stalled the whole audio bring-up.
func TestMailboxWaitIdiom(t *testing.T) {
	bus := &mailboxBus{cmbh: 0x8000} // present bit set
	c := New(bus)
	copy(c.IRAM[:], []uint16{
		0x8100,         // 0000: clr ac0
		0x26FE,         // 0001: lrs ac0.m, @0xFFFE (read CMBH)
		0x02C0, 0x8000, // 0002: andcf ac0.m, #0x8000
		0x029C, 0x0000, // 0004: jmplnz 0x0000 (loop while present bit clear)
		0x27FF,         // 0006: lrs ac1.m, @0xFFFF (consume CMBL)
		0x0021,         // 0007: halt (sentinel)
	})
	for i := 0; i < 100000 && c.PC != 0x0007; i++ {
		if !c.Step() {
			break
		}
	}
	if c.Halted {
		t.Fatalf("core halted: %s (PC=0x%04X)", c.Reason, c.PC)
	}
	if c.PC != 0x0007 {
		t.Fatalf("mailbox wait did not exit (PC=0x%04X); present bit was not detected", c.PC)
	}
	if !bus.consumed {
		t.Error("loop exited but never read the mailbox low half")
	}
}

// mailboxBus answers the CPU-mailbox reads the wait idiom makes: the high half carries the
// present bit, and reading the low half consumes the mail.
type mailboxBus struct {
	cmbh     uint16
	consumed bool
}

func (b *mailboxBus) HWRead(a uint16) uint16 {
	switch a {
	case 0xFFFE: // CMBH
		return b.cmbh
	case 0xFFFF: // CMBL — consume
		b.consumed = true
		b.cmbh &^= 0x8000
		return 0
	}
	return 0
}
func (b *mailboxBus) HWWrite(a uint16, v uint16) {}

// TestCallRet checks the call stack: a call runs a subroutine that loads a register, and the
// ret returns to the instruction after the call.
func TestCallRet(t *testing.T) {
	c := run(t, []uint16{
		0x02BF, 0x0006, // 0000: call 0x0006
		0x009F, 0xBEEF, // 0002: lri ac1.m, #0xBEEF (runs after ret)
		0x0021,         // 0004: halt (sentinel)
		0x0000,         // 0005: pad
		0x009D, 0xCAFE, // 0006: lri ac1.l, #0xCAFE
		0x02DF,         // 0008: ret
	}, 0x0004)
	if c.Reg[regAC1L] != 0xCAFE {
		t.Errorf("ac1.l = 0x%04X, want 0xCAFE (subroutine ran)", c.Reg[regAC1L])
	}
	if c.Reg[regAC1M] != 0xBEEF {
		t.Errorf("ac1.m = 0x%04X, want 0xBEEF (returned to caller)", c.Reg[regAC1M])
	}
}
