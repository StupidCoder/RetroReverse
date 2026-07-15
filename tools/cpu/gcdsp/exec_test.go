package gcdsp

import "testing"

// nullBus is a hardware bus that is never expected to be touched in these tests; if the program
// reaches a hardware address the test fails loudly rather than reading a silent zero.
type nullBus struct{ t *testing.T }

func (b nullBus) HWRead(a uint16) uint16 { b.t.Fatalf("unexpected HW read @0x%04X", a); return 0 }
func (b nullBus) HWWrite(a uint16, v uint16) {
	b.t.Fatalf("unexpected HW write @0x%04X = 0x%04X", a, v)
}

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
// accumulator middle word and copy it into the wrapping registers. The mid-load extension is
// MODE-DEPENDENT: at power-on (set16) the rest of the accumulator is untouched; under set40 a
// mid load sign-extends the high byte and clears the low word.
func TestInitIdiom(t *testing.T) {
	c := run(t, []uint16{
		0x009E, 0xFFFF, // 0000: lri ac0.m, #0xFFFF
		0x1D1E, // 0002: mrr wr0, ac0.m
		0x1D3E, // 0003: mrr wr1, ac0.m
		0x0021, // 0004: halt (sentinel; run stops before executing it)
	}, 0x0004)
	if c.Reg[regWR0] != 0xFFFF || c.Reg[regWR0+1] != 0xFFFF {
		t.Errorf("wr0/wr1 = 0x%04X/0x%04X, want 0xFFFF", c.Reg[regWR0], c.Reg[regWR0+1])
	}
	if c.Reg[regAC0H] != 0x0000 {
		t.Errorf("ac0.h = 0x%04X, want untouched 0x0000 (set16 is the power-on mode)", c.Reg[regAC0H])
	}

	// The same load under set40: the high byte extends and the low word clears.
	c2 := run(t, []uint16{
		0x8F00,         // 0000: set40
		0x009C, 0x1234, // 0001: lri ac0.l, #0x1234 (must be cleared by the mid load)
		0x009E, 0xFFFF, // 0003: lri ac0.m, #0xFFFF
		0x0021, // 0005: halt (sentinel)
	}, 0x0005)
	if c2.Reg[regAC0H] != 0xFFFF || c2.Reg[regAC0L] != 0x0000 {
		t.Errorf("set40 mid load: ac0.h/l = 0x%04X/0x%04X, want 0xFFFF/0x0000",
			c2.Reg[regAC0H], c2.Reg[regAC0L])
	}
}

// TestBlockLoop runs a block loop that stores a counter to consecutive DRAM words and increments
// the accumulator each pass — the shape of the ucode's DRAM clear and table fills. The store
// (srri) post-increments the address register itself, walking the buffer with no separate step;
// getting that auto-increment right is what lets the real DRAM-clear loop cover all of memory.
func TestBlockLoop(t *testing.T) {
	c := run(t, []uint16{
		0x0088, 0xFFFF, // 0000: lri wr0, #0xFFFF (no-wrap, as the ucode sets up)
		0x0080, 0x0100, // 0002: lri ar0, #0x0100
		0x009E, 0x0000, // 0004: lri ac0.m, #0x0000
		0x1104, 0x0009, // 0006: bloopi #4, 0x0009  (body 0x0008..0x0009)
		0x1B1E,         // 0008: srri @ar0, ac0.m  (store and post-increment ar0)
		0x0200, 0x0001, // 0009: addi ac0, #0x0001  (loop end)
		0x0021, // 000B: halt (sentinel)
	}, 0x000B)
	for i := uint16(0); i < 4; i++ {
		if got := c.DRAM[0x100+i]; got != i {
			t.Errorf("DRAM[0x%03X] = 0x%04X, want 0x%04X", 0x100+i, got, i)
		}
	}
	if c.Reg[regAR0] != 0x0104 {
		t.Errorf("ar0 = 0x%04X, want 0x0104 (post-incremented four times)", c.Reg[regAR0])
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
		0x0021, // 0006: halt (must be skipped)
		0x0021, // 0007: pad
		0x0000, // 0008: nop (sentinel)
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
		name   string
		op     uint16 // 0x02A0 andf, 0x02C0 andcf
		acm    uint16
		imm    uint16
		wantLZ bool
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
		0x27FF, // 0006: lrs ac1.m, @0xFFFF (consume CMBL)
		0x0021, // 0007: halt (sentinel)
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

// TestShiftAccumulator pins the shift group's semantics on the full 40-bit accumulator: a
// positive amount shifts left, a negative one shifts right; the logical form zero-fills a right
// shift while the arithmetic form carries the sign; and bit 8 selects the accumulator. Each case
// sets the accumulator directly, steps the single shift word, and reads the 40-bit result back.
func TestShiftAccumulator(t *testing.T) {
	cases := []struct {
		name string
		op   uint16
		in   int64
		want int64
	}{
		{"lsl-1", 0x1401, 0x0000001234, 0x0000002468},
		{"lsr-4", 0x147C, 0x0000012340, 0x0000001234},
		// The command-dispatch shift: ac0.m=0x8000 sign-extends to -0x80000000; LSR by 7
		// zero-fills to 0x01FF000000, whose middle word (0xFF00) the ucode then masks away.
		{"lsr-7-of-neg", 0x1479, -0x80000000, 0x01FF000000},
		{"asr-4-neg", 0x14FC, -0x10000000, -0x1000000},
		{"lsr-4-of-neg-zerofills", 0x147C, -0x10000000, 0x0FFF000000},
		{"lsl-1-ac1", 0x1501, 0x0000001234, 0x0000002468},
	}
	for _, tc := range cases {
		c := New(nullBus{t})
		r := int((tc.op >> 8) & 1)
		c.setAc(r, tc.in)
		c.IRAM[0] = tc.op
		c.IRAM[1] = 0x0000 // nop sentinel
		if !c.Step() {
			t.Fatalf("%s: core halted: %s", tc.name, c.Reason)
		}
		if got := c.ac(r); got != tc.want {
			t.Errorf("%s: ac%d = 0x%010X, want 0x%010X", tc.name, r, uint64(got)&0xFFFFFFFFFF, uint64(tc.want)&0xFFFFFFFFFF)
		}
	}
}

// TestCommandDispatchShift runs the exact command-dispatch idiom at ucode 0x0040: the mailbox
// command word in ac0.m is shifted, masked to an even jump-table index, and offset by the table
// base. For command 0 (mailbox high = 0x8000, present bit only) the computed target is the table
// base 0x0062 itself — the first entry. A wrong shift lands on a different entry.
func TestCommandDispatchShift(t *testing.T) {
	c := run(t, []uint16{
		0x009E, 0x8000, // 0000: lri ac0.m, #0x8000 (the mailbox-high command word)
		0x1479,         // 0002: lsr ac0, #7
		0x0240, 0x007E, // 0003: andi ac0.m, #0x007E
		0x0200, 0x0062, // 0005: addi ac0, #0x0062
		0x0000, // 0007: nop sentinel
	}, 0x0007)
	if c.Reg[regAC0M] != 0x0062 {
		t.Errorf("dispatch target ac0.m = 0x%04X, want 0x0062 (jump-table entry 0 for command 0)", c.Reg[regAC0M])
	}
}

// TestAddSubAx pins the accumulator add/subtract of an extended register and the flags the
// branch conditions read. addax/subax take the 32-bit ax.h:ax.l, sign-extended into the 40-bit
// accumulator; the ax-select bit picks ax0 or ax1. Each case sets up the operands, runs the one
// op, and checks the accumulator and the sign/zero flags.
func TestAddSubAx(t *testing.T) {
	cases := []struct {
		name     string
		op       uint16
		ac       int64
		axh, axl uint16
		axsel    int // which ax register to load
		want     int64
		wantZero bool
		wantSign bool
	}{
		// subax ac0, ax0 (0x5800): 0x00010000 - 0x00008000 = 0x00008000 (the full 32-bit ax)
		{"subax-ac0-ax0", 0x5800, 0x00010000, 0x0000, 0x8000, 0, 0x00008000, false, false},
		// subax ac0, ax0 to zero
		{"subax-to-zero", 0x5800, 0x00008000, 0x0000, 0x8000, 0, 0x0000000000, true, false},
		// addax ac0, ax1 (0x4A00: d=0,s=1): negative result
		{"addax-ac0-ax1-neg", 0x4A00, -0x00010000, 0xFFFF, 0x0000, 1, -0x00020000, false, true},
		// subax ac1, ax1 (0x5B00: d=1,s=1)
		{"subax-ac1-ax1", 0x5B00, 0x00040000, 0x0001, 0x0000, 1, 0x00030000, false, false},
		// addr ac0, ax0.h (0x4400: reg=ax0.h): adds only the high 16 bits, shifted into the
		// middle — distinct from addax, which would also add ax0.l. Here ax0.l is set but ignored.
		{"addr-ac0-ax0h", 0x4400, 0x00000000, 0x0002, 0x8000, 0, 0x00020000, false, false},
		// subr ac0, ax0.h (0x5400): subtract only the high 16 bits
		{"subr-ac0-ax0h", 0x5400, 0x00030000, 0x0001, 0xFFFF, 0, 0x00020000, false, false},
	}
	for _, tc := range cases {
		c := New(nullBus{t})
		d := int((tc.op >> 8) & 1)
		c.setAc(d, tc.ac)
		c.Reg[regAX0H+tc.axsel] = tc.axh
		c.Reg[regAX0L+tc.axsel] = tc.axl
		c.IRAM[0] = tc.op
		c.IRAM[1] = 0x0000
		if !c.Step() {
			t.Fatalf("%s: core halted: %s", tc.name, c.Reason)
		}
		if got := c.ac(d); got != tc.want {
			t.Errorf("%s: ac%d = 0x%010X, want 0x%010X", tc.name, d, uint64(got)&0xFFFFFFFFFF, uint64(tc.want)&0xFFFFFFFFFF)
		}
		if z := c.Reg[regSR]&srZero != 0; z != tc.wantZero {
			t.Errorf("%s: zero=%v, want %v", tc.name, z, tc.wantZero)
		}
		if s := c.Reg[regSR]&srSign != 0; s != tc.wantSign {
			t.Errorf("%s: sign=%v, want %v", tc.name, s, tc.wantSign)
		}
	}
}

// TestMixerLogicAndMoves pins the logic ops and the moves the mixer's format-conversion routines
// use (ucode 0x03B3 and the copy loops). The opcodes and their operand bits are the documented set.
func TestMixerLogicAndMoves(t *testing.T) {
	// not ac0.m (0x3280): complement the middle word, sign-extending the high byte.
	c := New(nullBus{t})
	c.setReg(regAC0M, 0x1234)
	c.IRAM[0] = 0x3280
	c.Step()
	if got := c.Reg[regAC0M]; got != 0xEDCB {
		t.Errorf("not ac0.m = 0x%04X, want 0xEDCB", got)
	}

	// xorr ac1.m, ax0.h (0x3100: d=bit8=1 ac1, s=bit9=0 ax0): ac1.m ^= ax0.h.
	c = New(nullBus{t})
	c.setReg(regAC1M, 0xFF00)
	c.Reg[regAX0H] = 0x0F0F
	c.IRAM[0] = 0x3100
	c.Step()
	if got := c.Reg[regAC1M]; got != 0xF00F {
		t.Errorf("xorr ac1.m = 0x%04X, want 0xF00F", got)
	}

	// movr ac0, ax1.h (0x6000 | reg field 3 in bits 10..9 = 0x0600): middle = ax1.h, low = 0, sign.
	c = New(nullBus{t})
	c.Reg[regAX1H] = 0x8001
	c.setAc(0, 0x00FFFFFFFF)
	c.IRAM[0] = 0x6000 | 0x0600
	c.Step()
	if got := c.ac(0); got != -0x7FFF0000 {
		t.Errorf("movr ac0 = 0x%010X, want the sign-extended ax1.h in the middle", uint64(got)&0xFFFFFFFFFF)
	}

	// movax ac1, ax0 (0x6800 | d=1 in bit8 = 0x0100): full 32-bit ax into the 40-bit accumulator.
	c = New(nullBus{t})
	c.Reg[regAX0H] = 0xFFFF
	c.Reg[regAX0L] = 0x0001
	c.IRAM[0] = 0x6800 | 0x0100
	c.Step()
	if got := c.ac(1); got != -0xFFFF {
		t.Errorf("movax ac1 = 0x%010X, want -0xFFFF (sign-extended ax0)", uint64(got)&0xFFFFFFFFFF)
	}
}

// TestMul pins the multiplier: prod = a signed 16x16 product, doubled in m2 mode. The mixer's
// resampling loops depend on both the product value and the m0/m2 shift.
func TestMul(t *testing.T) {
	// mul ax0 (0x9000) in m0 mode: prod = ax0.l * ax0.h, no doubling (M0 SETS the modify bit).
	c := New(nullBus{t})
	c.setFlag(srMulNoDouble, true)
	c.Reg[regAX0L] = 0x0004
	c.Reg[regAX0H] = 0x0100
	c.IRAM[0] = 0x9000
	c.Step()
	if got := c.prod(); got != 0x400 {
		t.Errorf("mul m0 prod = 0x%X, want 0x400", got)
	}

	// In m2 mode (the modify bit clear — the POWER-ON DEFAULT) the product is doubled.
	c = New(nullBus{t})
	c.Reg[regAX0L] = 0x0004
	c.Reg[regAX0H] = 0x0100
	c.IRAM[0] = 0x9000
	c.Step()
	if got := c.prod(); got != 0x800 {
		t.Errorf("mul m2/default prod = 0x%X, want 0x800 (doubling is the default)", got)
	}

	// A signed product under m0: -2 * 3 = -6.
	c = New(nullBus{t})
	c.setFlag(srMulNoDouble, true)
	c.Reg[regAX0L] = 0xFFFE // -2
	c.Reg[regAX0H] = 0x0003
	c.IRAM[0] = 0x9000
	c.Step()
	if got := c.prod(); got != -6 {
		t.Errorf("mul signed prod = %d, want -6", got)
	}

	// mulx ax0.h, ax1.h (0xB800): the cross product uses a half of ax0 and a half of ax1.
	c = New(nullBus{t})
	c.setFlag(srMulNoDouble, true)
	c.Reg[regAX0H] = 0x0002
	c.Reg[regAX1H] = 0x0003
	c.IRAM[0] = 0xB800 // bit12,11 set -> ax0.h, ax1.h
	c.Step()
	if got := c.prod(); got != 6 {
		t.Errorf("mulx prod = %d, want 6", got)
	}

	// set15 makes the MULX low halves unsigned — and ONLY the MULX family; a plain mul stays
	// signed even in set15.
	c = New(nullBus{t})
	c.setFlag(srMulNoDouble, true)
	c.setFlag(srMulUnsigned, true)
	c.Reg[regAX0L] = 0xFFFE // 65534 unsigned
	c.Reg[regAX1L] = 0x0002
	c.IRAM[0] = 0xA000 // mulx ax0.l, ax1.l
	c.Step()
	if got := c.prod(); got != 65534*2 {
		t.Errorf("set15 mulx .l*.l prod = %d, want %d (fully unsigned)", got, 65534*2)
	}
	c = New(nullBus{t})
	c.setFlag(srMulNoDouble, true)
	c.setFlag(srMulUnsigned, true)
	c.Reg[regAX0L] = 0xFFFE // -2 — plain mul is signed regardless of set15
	c.Reg[regAX0H] = 0x0003
	c.IRAM[0] = 0x9000
	c.Step()
	if got := c.prod(); got != -6 {
		t.Errorf("set15 plain mul prod = %d, want -6 (set15 is MULX-only)", got)
	}
}

// TestExtParallelStore pins the hazard-free timing of a parallel store: it must capture the
// accumulator's value from BEFORE the main op changes it. Two instructions pin it: the ucode's
// own `not ac0.m : s @ar2, ac0.m` (0x32B2, at 0x03B7 — a 7-bit extension; the stored word must
// be the original), and an ls pair on a row with the full 8-bit field (0x64B1: movr overwrites
// ac0 while the ls stores its pre-op middle and loads ax0.h).
func TestExtParallelStore(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.setReg(regAC0M, 0x1234)
	c.Reg[regAR0+2] = 0x0200 // store destination (ar2)
	c.IRAM[0] = 0x32B2       // not ac0.m : s @ar2, ac0.m
	c.Step()
	if got := c.DRAM[0x0200]; got != 0x1234 {
		t.Errorf("parallel store wrote 0x%04X, want 0x1234 (the pre-op accumulator middle)", got)
	}
	if got := c.Reg[regAC0M]; got != 0xEDCB {
		t.Errorf("not left ac0.m = 0x%04X, want 0xEDCB", got)
	}
	if c.Reg[regAR0+2] != 0x0201 {
		t.Errorf("parallel store did not post-increment ar2: 0x%04X", c.Reg[regAR0+2])
	}

	// The full-8-bit row: movr ac0, ax1.l : ls ax0.h, ac0.m (0x62B1 with the ls in the low
	// byte) — the ls stores the PRE-movr ac0.m through AR3 and loads ax0.h through AR0.
	c2 := New(nullBus{t})
	noWrap(c2)
	c2.setReg(regAC0M, 0x4321)
	c2.Reg[regAX1L] = 0x0055
	c2.Reg[regAR0] = 0x0100   // load source
	c2.Reg[regAR0+3] = 0x0200 // store destination
	c2.DRAM[0x0100] = 0xBEEF
	c2.IRAM[0] = 0x62A0 // movr ac0, ax1.l : ls ax0.h, ac0.m
	c2.Step()
	if got := c2.DRAM[0x0200]; got != 0x4321 {
		t.Errorf("ls stored 0x%04X, want 0x4321 (the pre-movr accumulator middle)", got)
	}
	if got := c2.Reg[regAC0M]; got != 0x0055 {
		t.Errorf("movr left ac0.m = 0x%04X, want 0x0055", got)
	}
	if got := c2.Reg[regAX0H]; got != 0xBEEF {
		t.Errorf("ls loaded ax0.h = 0x%04X, want 0xBEEF", got)
	}
	if c2.Reg[regAR0] != 0x0101 || c2.Reg[regAR0+3] != 0x0201 {
		t.Errorf("ls did not post-increment both pointers: ar0=0x%04X ar3=0x%04X",
			c2.Reg[regAR0], c2.Reg[regAR0+3])
	}
}

// TestAddisCmpis pins the single-word short-immediate ops: addis adds an 8-bit immediate into the
// accumulator middle (how the mixer builds a second buffer pointer 0x50 words along), cmpis
// compares against imm<<16.
func TestAddisCmpis(t *testing.T) {
	c := New(nullBus{t})
	c.setAc(1, 0x00_0010_0000)
	c.IRAM[0] = 0x0500 | 0x0050 // addis ac1, #0x50
	c.Step()
	if got := c.ac(1); got != 0x00_0060_0000 {
		t.Errorf("addis ac1 = 0x%010X, want 0x0000600000", uint64(got)&0xFFFFFFFFFF)
	}

	// cmpis ac1, #0x60 against a middle of 0x60 -> zero flag set, no store.
	c = New(nullBus{t})
	c.setAc(1, 0x00_0060_0000)
	c.IRAM[0] = 0x0700 | 0x0060 // cmpis ac1, #0x60
	c.Step()
	if c.Reg[regSR]&srZero == 0 {
		t.Error("cmpis did not set the zero flag on an equal compare")
	}
	if got := c.ac(1); got != 0x00_0060_0000 {
		t.Errorf("cmpis modified the accumulator (0x%010X)", uint64(got)&0xFFFFFFFFFF)
	}
}

// TestTstBranch runs the mixer's "is this sample negative" idiom: tst the accumulator then take
// a GE branch, which with overflow cleared reduces to the sign. A negative accumulator must fall
// through (not GE); a non-negative one must branch.
// TestTstBranch pins the two distinct test ops: 0xB100 is tst acR (flags from the 40-bit
// accumulator, selector in bit 11), 0x8600 is tstaxh axR.h (flags from the 16-bit HIGH half of
// an ax register, selector in bit 8). An earlier build ran 0x8600 as the accumulator test —
// self-consistent with its own version of this test, silently wrong against the ISA — so the
// two cases here deliberately disagree: the accumulator is set opposite in sign to ax0.h and
// each op must follow its own operand.
func TestTstBranch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		op     uint16
		ac     int64
		axh    uint16
		wantGE bool
	}{
		{"tst-negative", 0xB100, -0x0010000, 0x0001, false}, // ac0 < 0 while ax0.h > 0
		{"tst-positive", 0xB100, 0x0010000, 0x8000, true},   // ac0 > 0 while ax0.h < 0
		{"tst-zero", 0xB100, 0, 0x8000, true},
		{"tstaxh-negative", 0x8600, 0x0010000, 0x8000, false}, // ax0.h < 0 while ac0 > 0
		{"tstaxh-positive", 0x8600, -0x0010000, 0x0001, true}, // ax0.h > 0 while ac0 < 0
		{"tstaxh-zero", 0x8600, -0x0010000, 0x0000, true},
	} {
		c := New(nullBus{t})
		c.setAc(0, tc.ac)
		c.Reg[regAX0H] = tc.axh
		copy(c.IRAM[:], []uint16{
			tc.op,          // 0000: tst ac0 / tstaxh ax0.h
			0x0290, 0x0005, // 0001: jmpge 0x0005
			0x0021, // 0003: halt (taken when not GE)
			0x0000, // 0004: pad
			0x0000, // 0005: nop (GE lands here)
		})
		for i := 0; i < 100 && c.PC != 0x0005 && !c.Halted; i++ {
			c.Step()
		}
		gotGE := c.PC == 0x0005 && !c.Halted
		if gotGE != tc.wantGE {
			t.Errorf("%s: GE-branch taken=%v, want %v (PC=0x%04X)", tc.name, gotGE, tc.wantGE, c.PC)
		}
	}
}

// TestCallRet checks the call stack: a call runs a subroutine that loads a register, and the
// ret returns to the instruction after the call.
func TestCallRet(t *testing.T) {
	c := run(t, []uint16{
		0x02BF, 0x0006, // 0000: call 0x0006
		0x009F, 0xBEEF, // 0002: lri ac1.m, #0xBEEF (runs after ret)
		0x0021,         // 0004: halt (sentinel)
		0x0000,         // 0005: pad
		0x009D, 0xCAFE, // 0006: lri ac1.l, #0xCAFE
		0x02DF, // 0008: ret
	}, 0x0004)
	if c.Reg[regAC1L] != 0xCAFE {
		t.Errorf("ac1.l = 0x%04X, want 0xCAFE (subroutine ran)", c.Reg[regAC1L])
	}
	if c.Reg[regAC1M] != 0xBEEF {
		t.Errorf("ac1.m = 0x%04X, want 0xBEEF (returned to caller)", c.Reg[regAC1M])
	}
}

// TestCallrJmpr covers the register-indirect control transfer the mixer's command decode uses.
// callr $ar1 (0x173F) stacks the return and jumps to the address the register holds; the handler
// rets back to the instruction after the callr — the idiom at ucode 0x04E0 and the 0x0276 jump
// table. jmpr $ar2 (0x174F) transfers with nothing stacked.
func TestCallrJmpr(t *testing.T) {
	c := run(t, []uint16{
		0x0081, 0x0006, // 0000: lri ar1, #0x0006
		0x173F,         // 0002: callr ar1  -> 0x0006, returns to 0x0003
		0x009F, 0xBEEF, // 0003: lri ac1.m, #0xBEEF (runs after the handler returns)
		0x0021,         // 0005: halt sentinel
		0x009D, 0xCAFE, // 0006: lri ac1.l, #0xCAFE (the handler)
		0x02DF, // 0008: ret
	}, 0x0005)
	if c.Reg[regAC1L] != 0xCAFE {
		t.Errorf("ac1.l = 0x%04X, want 0xCAFE (callr handler ran)", c.Reg[regAC1L])
	}
	if c.Reg[regAC1M] != 0xBEEF {
		t.Errorf("ac1.m = 0x%04X, want 0xBEEF (returned to the caller)", c.Reg[regAC1M])
	}

	c2 := run(t, []uint16{
		0x0082, 0x0004, // 0000: lri ar2, #0x0004
		0x174F,         // 0002: jmpr ar2 -> 0x0004 (no return stacked)
		0x0021,         // 0003: halt (skipped over)
		0x009F, 0xF00D, // 0004: lri ac1.m, #0xF00D
		0x0021, // 0006: halt sentinel
	}, 0x0006)
	if c2.Reg[regAC1M] != 0xF00D {
		t.Errorf("ac1.m = 0x%04X, want 0xF00D (jmpr landed)", c2.Reg[regAC1M])
	}
	if len(c2.Stacks[regST0-regST0]) != 0 {
		t.Errorf("jmpr pushed a return address (stack depth %d), want none", len(c2.Stacks[0]))
	}
}

// runOn steps a pre-configured CPU (registers and DRAM already set) until PC reaches stopAt.
func runOn(t *testing.T, c *CPU, stopAt uint16) {
	t.Helper()
	for i := 0; i < 200000; i++ {
		if c.PC == stopAt {
			return
		}
		if !c.Step() {
			break
		}
	}
	if c.Halted {
		t.Fatalf("core halted: %s (PC=0x%04X)", c.Reason, c.PC)
	}
	t.Fatalf("did not reach 0x%04X (stopped at 0x%04X)", stopAt, c.PC)
}

// noWrap puts every address register in the no-wrap configuration the ucode's init sets.
func noWrap(c *CPU) {
	for i := 0; i < 4; i++ {
		c.Reg[regWR0+i] = 0xFFFF
	}
}

// TestExtDualLoad pins the plain LD extension (11mn barr): one half of AX0 (bit 5) loads
// through arS, one half of AX1 (bit 4) through AR3, both post-modified — and, like every
// parallel load, the writes land only after the main op has read the old register values.
func TestExtDualLoad(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAR0+1] = 0x140
	c.Reg[regAR0+3] = 0x180
	c.Reg[regAX0L] = 0x5678 // the value the main op must still see
	c.DRAM[0x140] = 0x1111
	c.DRAM[0x180] = 0x2222
	copy(c.IRAM[:], []uint16{
		0x60C1, // movr ac0, ax0.l : ld ax0.l, ax1.l, @ar1
		0x0021,
	})
	runOn(t, c, 0x0001)
	if c.Reg[regAC0M] != 0x5678 {
		t.Errorf("main op saw ax0.l = 0x%04X, want the pre-load 0x5678", c.Reg[regAC0M])
	}
	if c.Reg[regAX0L] != 0x1111 || c.Reg[regAX1L] != 0x2222 {
		t.Errorf("ld loaded ax0.l/ax1.l = 0x%04X/0x%04X, want 0x1111/0x2222", c.Reg[regAX0L], c.Reg[regAX1L])
	}
	if c.Reg[regAR0+1] != 0x141 || c.Reg[regAR0+3] != 0x181 {
		t.Errorf("ld stepped ar1/ar3 to 0x%04X/0x%04X, want 0x0141/0x0181", c.Reg[regAR0+1], c.Reg[regAR0+3])
	}

	// The M variant the resampler uses (ext 0xE8 = ldm ax0.h, ax1.l, @ar0): AR3 steps by IX3.
	c2 := New(nullBus{t})
	noWrap(c2)
	c2.Reg[regAR0] = 0x100
	c2.Reg[regAR0+3] = 0x180
	c2.Reg[regIX0+3] = 0x0005
	c2.DRAM[0x100] = 0xAAAA
	c2.DRAM[0x180] = 0xBBBB
	copy(c2.IRAM[:], []uint16{
		0x80E8, // nx : ldm ax0.h, ax1.l, @ar0
		0x0021,
	})
	runOn(t, c2, 0x0001)
	if c2.Reg[regAX0H] != 0xAAAA || c2.Reg[regAX1L] != 0xBBBB {
		t.Errorf("ldm loaded ax0.h/ax1.l = 0x%04X/0x%04X, want 0xAAAA/0xBBBB", c2.Reg[regAX0H], c2.Reg[regAX1L])
	}
	if c2.Reg[regAR0] != 0x101 || c2.Reg[regAR0+3] != 0x185 {
		t.Errorf("ldm stepped ar0/ar3 to 0x%04X/0x%04X, want 0x0101/0x0185 (+1, +ix3)", c2.Reg[regAR0], c2.Reg[regAR0+3])
	}
}

// TestExtDualLoadBothHalves pins the LD2 form (low bits 11): BOTH halves of one ax register
// load at once — the high half through arS (bit 5), the low through AR3 — with bit 4 choosing
// the ax. This is the FIR's per-tap fetch (coefficient into .h via the circular AR0, sample
// into .l via AR3) and the volume loop's stream refresh.
func TestExtDualLoadBothHalves(t *testing.T) {
	for _, tc := range []struct {
		ext   uint16
		wantH uint16 // register receiving [arS]
		wantL uint16 // register receiving [AR3]
		arS   int
	}{
		{0xC3, regAX0H, regAX0L, 0}, // ld2 ax0, @ar0 — the FIR's
		{0xD3, regAX1H, regAX1L, 0}, // ld2 ax1, @ar0 — the volume loop's
		{0xE3, regAX0H, regAX0L, 1}, // ld2 ax0, @ar1
	} {
		c := New(nullBus{t})
		noWrap(c)
		c.Reg[regAR0+tc.arS] = 0x100
		c.Reg[regAR0+3] = 0x180
		c.DRAM[0x100] = 0xC0EF // the arS-side value (a coefficient in the FIR)
		c.DRAM[0x180] = 0x5A5A // the AR3-side value (a sample)
		copy(c.IRAM[:], []uint16{
			0x8000 | tc.ext, // nx : ld2
			0x0021,
		})
		runOn(t, c, 0x0001)
		if c.Reg[tc.wantH] != 0xC0EF {
			t.Errorf("ext 0x%02X: high-half reg = 0x%04X, want 0xC0EF from [ar%d]", tc.ext, c.Reg[tc.wantH], tc.arS)
		}
		if c.Reg[tc.wantL] != 0x5A5A {
			t.Errorf("ext 0x%02X: low-half reg = 0x%04X, want 0x5A5A from [ar3]", tc.ext, c.Reg[tc.wantL])
		}
		if c.Reg[regAR0+tc.arS] != 0x101 || c.Reg[regAR0+3] != 0x181 {
			t.Errorf("ext 0x%02X: ar%d/ar3 = 0x%04X/0x%04X, want both post-incremented",
				tc.ext, tc.arS, c.Reg[regAR0+tc.arS], c.Reg[regAR0+3])
		}
	}
}

// TestClrpMovpz pins the product-register idioms around the FIR: clrp sets the documented
// biased pieces that read back as zero, movpz moves prod into an accumulator with the low
// word cleared, and movp moves it exactly.
func TestClrpMovpz(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.setAc(0, 0x123456789A) // must be overwritten
	copy(c.IRAM[:], []uint16{
		0x8400, // clrp
		0xFE00, // movpz ac0
		0x0021,
	})
	runOn(t, c, 0x0002)
	if c.Reg[regPRODL] != 0x0000 || c.Reg[regPRODM1] != 0xFFF0 || c.Reg[regPRODH] != 0x00FF || c.Reg[regPRODM2] != 0x0010 {
		t.Errorf("clrp pieces = %04X/%04X/%04X/%04X, want 0000/FFF0/00FF/0010",
			c.Reg[regPRODL], c.Reg[regPRODM1], c.Reg[regPRODH], c.Reg[regPRODM2])
	}
	if got := c.ac(0); got != 0 {
		t.Errorf("movpz after clrp: ac0 = 0x%X, want 0 (the biased pieces must sum to zero)", got)
	}

	// m0; clrp; madd; movp — the product lands exactly (signed operands, no doubling).
	c2 := New(nullBus{t})
	noWrap(c2)
	c2.Reg[regAX0L] = 0x4000 // +16384
	c2.Reg[regAX0H] = 0xE000 // -8192
	copy(c2.IRAM[:], []uint16{
		0x8B00, // m0 (doubling is otherwise the power-on default)
		0x8400, // clrp
		0xF200, // madd ax0.l, ax0.h
		0x6E00, // movp ac0
		0x0021,
	})
	runOn(t, c2, 0x0004)
	if want := int64(16384) * -8192; c2.ac(0) != want {
		t.Errorf("clrp; madd; movp: ac0 = %d, want %d", c2.ac(0), want)
	}
}

// TestMulcac pins MULCAC (110s t10r): the old product is added into acR while a new product
// acS.m * axT.h starts — the volume loop's pipeline.
func TestMulcac(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAC1M] = 0x4000 // the volume, in ac1.m
	c.Reg[regAX1H] = 0x0100 // first sample
	copy(c.IRAM[:], []uint16{
		0x8B00,         // m0
		0xD800,         // mulc ac1.m, ax1.h — prod = vol * s1
		0x009B, 0xFF00, // lri ax1.h, #0xFF00 — second sample (-256)
		0xDC00, // mulcac ac1.m, ax1.h, ac0 — ac0 += old prod; prod = vol * s2
		0x0021,
	})
	runOn(t, c, 0x0005)
	if want := int64(0x4000) * 0x0100; c.ac(0) != want {
		t.Errorf("mulcac accumulated ac0 = %d, want the FIRST product %d", c.ac(0), want)
	}
	if want := int64(0x4000) * -256; c.prod() != want {
		t.Errorf("mulcac left prod = %d, want the SECOND product %d", c.prod(), want)
	}
}

// TestMaddcMaddx pins the two-operand multiply-accumulates: maddc (acS.m * axT.h) and maddx
// (a half of ax0 times a half of ax1), both adding into prod.
func TestMaddcMaddx(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAC1M] = 0x0123
	c.Reg[regAX0H] = 0xFFFE // -2
	c.Reg[regAX0L] = 0x0040
	c.Reg[regAX1H] = 0x0010
	copy(c.IRAM[:], []uint16{
		0x8B00, // m0
		0x8400, // clrp
		0xEB00, // maddc ac1.m, ax1.h — prod += 0x0123 * 0x0010
		0xE100, // maddx ax0.l, ax1.h — prod += 0x0040 * 0x0010
		0xE600, // msubx ax0.h, ax1.l — prod -= (-2) * ax1.l(0)
		0x6E00, // movp ac0
		0x0021,
	})
	runOn(t, c, 0x0006)
	want := int64(0x0123)*0x0010 + int64(0x0040)*0x0010
	if c.ac(0) != want {
		t.Errorf("maddc+maddx: ac0 = %d, want %d", c.ac(0), want)
	}
}

// TestFIR runs the game's own 8-tap FIR routine (ucode 0x01C2..0x01E0, lifted verbatim with
// only the bloopi target rebased): copy 8 coefficients into the circular table at 0x03E8,
// then filter 0x50 samples in place through `clrp:ld2; (madd:ld2)x7; madd; movpz; srri`,
// with AR0 circling the coefficients under wr0=7 and AR3 walking the sample window. Verified
// against a direct Go convolution in the DSP's fixed-point convention (signed 16x16, m2
// product doubling, the result word taken from prod bits 16..31).
func TestFIR(t *testing.T) {
	coefs := []uint16{0x0100, 0xFF00, 0x2000, 0x8000, 0x0001, 0x7FFF, 0xF000, 0x0800}
	const nOut = 0x50
	samples := make([]uint16, nOut+8)
	for k := range samples {
		samples[k] = uint16(0x1234*k + 0x89)
	}

	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAR0] = 0x100   // the caller's coefficient list
	c.Reg[regAR0+1] = 0x200 // the sample buffer, filtered in place
	copy(c.DRAM[0x100:], coefs)
	copy(c.DRAM[0x200:], samples)
	copy(c.IRAM[:], []uint16{
		0x8A00,         // 00: m2
		0x0083, 0x03E8, // 01: lri ar3, #0x03E8
		0x191E,         // 03: lrri ac0.m, @ar0
		0x191A,         // 04: lrri ax0.h, @ar0
		0x1006,         // 05: loopi #6
		0x64A0,         // 06: movr ac0, ax0.h : ls ax0.h, ac0.m
		0x1B7E,         // 07: srri @ar3, ac0.m
		0x1B7A,         // 08: srri @ar3, ax0.h
		0x0080, 0x03E8, // 09: lri ar0, #0x03E8
		0x0088, 0x0007, // 0B: lri wr0, #0x0007
		0x1150, 0x001A, // 0D: bloopi #0x50, 0x001A
		0x1C61,                                         // 0F: mrr ar3, ar1
		0x84C3,                                         // 10: clrp : ld2 ax0, @ar0
		0xF2C3,                                         // 11: madd ax0 : ld2 ax0, @ar0
		0xF2C3, 0xF2C3, 0xF2C3, 0xF2C3, 0xF2C3, 0xF2C3, // 12..17
		0xF200,         // 18: madd ax0
		0xFE00,         // 19: movpz ac0
		0x1B3E,         // 1A: srri @ar1, ac0.m (loop end)
		0x0088, 0xFFFF, // 1B: lri wr0, #0xFFFF
		0x8B00, // 1D: m0
		0x0021, // 1E: halt sentinel
	})
	runOn(t, c, 0x001E)

	// The coefficient table must have been copied in caller order.
	for i, want := range coefs {
		if got := c.DRAM[0x3E8+i]; got != want {
			t.Fatalf("coefficient table [%d] = 0x%04X, want 0x%04X", i, got, want)
		}
	}
	// Every output: the Q15 MAC over the original window (outputs land at index n, windows
	// read n..n+7, so the in-place write never aliases a later read), read out through
	// movpz's ROUNDING of the product to its middle word.
	for n := 0; n < nOut; n++ {
		var acc int64
		for i := 0; i < 8; i++ {
			acc += 2 * int64(int16(coefs[i])) * int64(int16(samples[n+i]))
		}
		if acc&0x10000 != 0 {
			acc += 0x8000
		} else {
			acc += 0x7FFF
		}
		if got, want := c.DRAM[0x200+n], uint16(acc>>16); got != want {
			t.Fatalf("output %d = 0x%04X, want 0x%04X", n, got, want)
		}
	}
	if c.Reg[regAR0] != 0x3E8 {
		t.Errorf("ar0 = 0x%04X, want 0x03E8 (80 full circles of the 8-word table)", c.Reg[regAR0])
	}
	if c.Reg[regWR0] != 0xFFFF {
		t.Errorf("wr0 = 0x%04X, want restored 0xFFFF", c.Reg[regWR0])
	}
}

// TestIfCC runs the ucode's sign-tracked step idiom (0x0616: ifg incm / ifl decm): after a
// tst, exactly one of the two guarded single-word steps must execute.
func TestIfCC(t *testing.T) {
	for _, tc := range []struct {
		name string
		ac1  int64
		want uint16 // ac0.m after the pair
	}{
		{"positive-increments", 0x10000, 0x0011},
		{"negative-decrements", -0x10000, 0x000F},
		{"zero-does-neither", 0, 0x0010},
	} {
		c := New(nullBus{t})
		noWrap(c)
		c.setAc(1, tc.ac1)
		c.Reg[regAC0M] = 0x0010
		copy(c.IRAM[:], []uint16{
			0xB900, // tst ac1
			0x0272, // ifg
			0x7400, // incm ac0.m
			0x0271, // ifl
			0x7800, // decm ac0.m
			0x0021,
		})
		runOn(t, c, 0x0005)
		if c.Reg[regAC0M] != tc.want {
			t.Errorf("%s: ac0.m = 0x%04X, want 0x%04X", tc.name, c.Reg[regAC0M], tc.want)
		}
	}
}

// TestArWrap pins the carry-detect wrap on the two rings the ucode actually configures: the
// FIR's 8-word coefficient table (wr=7, 0x03E8..0x03EF) stepped +1, and the pitch resampler's
// 160-word sample ring (wr=0x9F, 0x0B60..0x0BFF) stepped by an arbitrary positive ix — the
// wrap fires when the step flips an address bit above the ring's span, which is why the ucode
// parks both rings' ends just below aligned boundaries.
func TestArWrap(t *testing.T) {
	for _, tc := range []struct {
		name  string
		wr    uint16
		ar    uint16
		delta int
		want  uint16
	}{
		{"fir-mid", 7, 0x03EB, 1, 0x03EC},
		{"fir-end-wraps", 7, 0x03EF, 1, 0x03E8},
		{"ring-mid", 0x9F, 0x0B9F, 1, 0x0BA0}, // low byte passes 0x9F mid-ring: NOT a wrap point
		{"ring-end-wraps", 0x9F, 0x0BFF, 1, 0x0B60},
		{"ring-step-2-over-end", 0x9F, 0x0BFE, 3, 0x0B61},
		{"ring-step-mid", 0x9F, 0x0BE0, 0x30, 0x0B70},
		{"ring-step-zero", 0x9F, 0x0BFF, 0, 0x0BFF},
	} {
		c := New(nullBus{t})
		c.Reg[regWR0+3] = tc.wr
		c.Reg[regAR0+3] = tc.ar
		if tc.delta == 1 {
			c.arInc(3) // the dedicated +1 form srri/lrri use
		} else {
			c.arAdd(3, int16(tc.delta)) // the index form addarn uses
		}
		if c.Halted {
			t.Fatalf("%s: halted: %s", tc.name, c.Reason)
		}
		if got := c.Reg[regAR0+3]; got != tc.want {
			t.Errorf("%s: ar3 = 0x%04X, want 0x%04X", tc.name, got, tc.want)
		}
	}
}

// TestSevenBitExtRow pins the 0x3xxx row's SEVEN-bit extension field: `not ac0.m` at 0x3280
// has bit 7 as opcode, not an `ls` extension — it must execute as a pure register op (no
// store through AR3, no load into ax0.l, no pointer steps). An 8-bit misread here performs a
// phantom memory store the ucode never asked for. A real 7-bit extension (0x3644: andr with
// a parallel load) must still run.
func TestSevenBitExtRow(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAR0] = 0x100
	c.Reg[regAR0+3] = 0x180
	c.Reg[regAX0L] = 0x5A5A
	c.Reg[regAC0M] = 0x00FF
	c.DRAM[0x100] = 0x1111
	c.DRAM[0x180] = 0x2222
	copy(c.IRAM[:], []uint16{
		0x3280, // not ac0.m — ext field is 0x00, NOT "ls"
		0x0021,
	})
	runOn(t, c, 0x0001)
	if c.Reg[regAC0M] != 0xFF00 {
		t.Errorf("not ac0.m = 0x%04X, want 0xFF00", c.Reg[regAC0M])
	}
	if c.Reg[regAX0L] != 0x5A5A {
		t.Errorf("ax0.l = 0x%04X, want untouched 0x5A5A (phantom ls load)", c.Reg[regAX0L])
	}
	if c.DRAM[0x180] != 0x2222 {
		t.Errorf("[ar3] = 0x%04X, want untouched 0x2222 (phantom ls store)", c.DRAM[0x180])
	}
	if c.Reg[regAR0] != 0x100 || c.Reg[regAR0+3] != 0x180 {
		t.Errorf("ar0/ar3 = 0x%04X/0x%04X, want unstepped 0x100/0x180", c.Reg[regAR0], c.Reg[regAR0+3])
	}

	// A genuine 7-bit extension on the same row: andr with a parallel 'ln (ucode 0x3644 —
	// ext 0x44 loads ax0.l through AR0 and steps AR0 by its index register).
	c2 := New(nullBus{t})
	noWrap(c2)
	c2.Reg[regAR0] = 0x100
	c2.Reg[regIX0] = 0x0002
	c2.Reg[regAC0M] = 0x0F0F
	c2.Reg[regAX1H] = 0x00FF
	c2.DRAM[0x100] = 0x7777
	copy(c2.IRAM[:], []uint16{
		0x3644, // andr ac0.m, ax1.h : ln ax0.l, @ar0
		0x0021,
	})
	runOn(t, c2, 0x0001)
	if c2.Reg[regAC0M] != 0x000F {
		t.Errorf("andr ac0.m = 0x%04X, want 0x000F", c2.Reg[regAC0M])
	}
	if c2.Reg[regAX0L] != 0x7777 || c2.Reg[regAR0] != 0x102 {
		t.Errorf("parallel ln: ax0.l = 0x%04X ar0 = 0x%04X, want 0x7777/0x0102 (stepped by ix0)", c2.Reg[regAX0L], c2.Reg[regAR0])
	}
}

// TestShiftByRegister pins the register-amount shift family found by hardware testing (absent
// from both paper sources): the 7-bit signed amount convention (low six bits of zero mean NO
// shift even with the sign bit set), and the direction split — lsrn/asrn (0x02CA/CB, amount in
// ac1.m, acting on ac0) shift RIGHT for positive amounts, while lsrnrx/asrnrx (amount in an ax
// high) and lsrnr/asrnr (amount in the other accumulator's middle) shift LEFT.
func TestShiftByRegister(t *testing.T) {
	for _, tc := range []struct {
		name   string
		op     uint16
		amtReg uint16 // where the amount goes
		amt    uint16
		in     int64 // preset into the target accumulator
		want   int64
	}{
		{"lsrn-right-4", 0x02CA, regAC1M, 0x0004, 0x0000012340, 0x0000001234},
		{"lsrn-left-neg4", 0x02CA, regAC1M, 0x7C, 0x0000001234, 0x0000012340}, // -4 in 7 bits
		{"lsrn-zero-low-bits", 0x02CA, regAC1M, 0x0040, 0x0000001234, 0x0000001234},
		{"lsrn-logical-zero-fill", 0x02CA, regAC1M, 0x0004, -0x1000000000, 0x0F00000000},
		{"asrn-arith-sign-fill", 0x02CB, regAC1M, 0x0004, -0x1000000000, -0x0100000000},
		{"asrnrx-ax0h-left-4", 0x3880, regAX0H, 0x0004, 0x0000001234, 0x0000012340},
		{"asrnrx-ax0h-right-neg8", 0x3880, regAX0H, 0x78, -0x1000000000, -0x0010000000},
		{"asrnrx-ax1h-ac1", 0x3B80, regAX1H, 0x0004, 0x0000001234, 0x0000012340},
		{"lsrnrx-ax0h-right-neg8-zerofill", 0x3480, regAX0H, 0x78, -0x1000000000, 0x00F0000000},
		{"lsrnr-by-other-mid", 0x3C80, regAC1M, 0x0004, 0x0000001234, 0x0000012340},
		{"asrnr-ac1-by-ac0m", 0x3E80 | 0x0100, regAC0M, 0x0004, 0x0000001234, 0x0000012340},
	} {
		c := New(nullBus{t})
		noWrap(c)
		d := 0
		if tc.op&0x0100 != 0 && tc.op != 0x02CA && tc.op != 0x02CB {
			d = 1
		}
		c.setAc(d, tc.in)
		c.Reg[tc.amtReg] = tc.amt
		copy(c.IRAM[:], []uint16{tc.op, 0x0021})
		runOn(t, c, 0x0001)
		if got := c.ac(d); got != tc.want {
			t.Errorf("%s: ac%d = 0x%010X, want 0x%010X", tc.name, d, uint64(got)&0xFFFFFFFFFF, uint64(tc.want)&0xFFFFFFFFFF)
		}
	}
}

// TestXorc pins xorc acD.m ^= ac(1-D).m — the bit7=1 sibling of the xorr row.
func TestXorc(t *testing.T) {
	c := New(nullBus{t})
	noWrap(c)
	c.Reg[regAC0M] = 0xFF00
	c.Reg[regAC1M] = 0x0F0F
	copy(c.IRAM[:], []uint16{0x3080, 0x0021})
	runOn(t, c, 0x0001)
	if got := c.Reg[regAC0M]; got != 0xF00F {
		t.Errorf("xorc ac0.m = 0x%04X, want 0xF00F", got)
	}
}

// TestMidReadSaturation pins the mode-dependent accumulator-middle READ: in set40 a middle
// whose accumulator does not fit in 32 bits reads clamped to 0x7FFF/0x8000 (the mixer's MAC
// results clip through exactly this on their way to memory); in set16 it reads plainly.
func TestMidReadSaturation(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode uint16 // 0x8F00 set40 / 0x8E00 set16
		in   int64
		want uint16
	}{
		{"set40-positive-clamps", 0x8F00, 0x0123456789, 0x7FFF},
		{"set40-negative-clamps", 0x8F00, -0x0123456789, 0x8000},
		{"set40-fitting-plain", 0x8F00, 0x00003456 << 16, 0x3456},
		{"set16-overflow-plain", 0x8E00, 0x0123456789, 0x2345},
	} {
		c := New(nullBus{t})
		noWrap(c)
		c.setAc(0, tc.in)
		c.Reg[regAR0+2] = 0x0100
		copy(c.IRAM[:], []uint16{
			tc.mode,
			0x1B5E, // srri @ar2, ac0.m — a store reads the middle through the register file
			0x0021,
		})
		runOn(t, c, 0x0002)
		if got := c.DRAM[0x100]; got != tc.want {
			t.Errorf("%s: stored 0x%04X, want 0x%04X", tc.name, got, tc.want)
		}
	}
}

// TestAcHighWrite pins that a write to ac?.h keeps only its signed low byte — only 8 bits of
// the high word are real.
func TestAcHighWrite(t *testing.T) {
	c := run(t, []uint16{
		0x0090, 0x1280, // lri ac0.h, #0x1280 — only the 0x80 byte is real, sign-extended
		0x0021,
	}, 0x0002)
	if got := c.Reg[regAC0H]; got != 0xFF80 {
		t.Errorf("ac0.h = 0x%04X, want 0xFF80 (sign-extended from the low byte)", got)
	}
}
