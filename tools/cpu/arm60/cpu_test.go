package arm60

import "testing"

// testBus is a flat little byte array with big-endian word helpers for tests.
type testBus struct{ mem []byte }

func (b *testBus) Read(a uint32) byte {
	if int(a) < len(b.mem) {
		return b.mem[a]
	}
	return 0
}
func (b *testBus) Write(a uint32, v byte) {
	if int(a) < len(b.mem) {
		b.mem[a] = v
	}
}
func (b *testBus) put32(a uint32, v uint32) {
	b.mem[a] = byte(v >> 24)
	b.mem[a+1] = byte(v >> 16)
	b.mem[a+2] = byte(v >> 8)
	b.mem[a+3] = byte(v)
}

// runProg loads big-endian words at 0x1000, runs n steps in System mode.
func runProg(prog []uint32, n int) (*CPU, *testBus) {
	bus := &testBus{mem: make([]byte, 0x4000)}
	for i, w := range prog {
		bus.put32(0x1000+uint32(i)*4, w)
	}
	c := NewCPU(bus)
	c.Mode = ModeSYS
	c.SetPC(0x1000)
	c.R[13] = 0x3000 // sp
	for i := 0; i < n; i++ {
		c.Step()
	}
	return c, bus
}

func TestDataProcessing(t *testing.T) {
	// MOV r0,#5 ; ADD r1,r0,#3 ; SUB r2,r0,r1 ; MOV r3,r2,LSL #2 ; ORR r4,r0,#0x100
	c, _ := runProg([]uint32{
		0xE3A00005, // MOV r0,#5
		0xE2801003, // ADD r1,r0,#3
		0xE0402001, // SUB r2,r0,r1
		0xE1A03102, // MOV r3,r2,LSL #2
		0xE3804C01, // ORR r4,r0,#0x100
	}, 5)
	if c.Halted {
		t.Fatalf("halted: %s", c.HaltReason)
	}
	want := map[int]uint32{0: 5, 1: 8, 2: 0xFFFFFFFD, 3: 0xFFFFFFF4, 4: 0x105}
	for r, v := range want {
		if c.R[r] != v {
			t.Errorf("r%d = 0x%08X, want 0x%08X", r, c.R[r], v)
		}
	}
}

func TestFlagsAndConditional(t *testing.T) {
	// CMP r0(=0),#0 sets Z; MOVEQ r1,#1 executes; MOVNE r2,#1 does not.
	c, _ := runProg([]uint32{
		0xE3500000, // CMP r0,#0
		0x03A01001, // MOVEQ r1,#1
		0x13A02001, // MOVNE r2,#1
	}, 3)
	if !c.Z {
		t.Errorf("Z not set after CMP r0,#0")
	}
	if c.R[1] != 1 {
		t.Errorf("MOVEQ did not execute (r1=%d)", c.R[1])
	}
	if c.R[2] != 0 {
		t.Errorf("MOVNE wrongly executed (r2=%d)", c.R[2])
	}
}

func TestLoadStoreBigEndian(t *testing.T) {
	// MOV r0,#0x2000 ; MOV r1,#0xFF ; STR r1,[r0] ; LDR r2,[r0] ; LDRB r3,[r0]
	c, bus := runProg([]uint32{
		0xE3A00A02, // MOV r0,#0x2000
		0xE3A010FF, // MOV r1,#0xFF
		0xE5801000, // STR r1,[r0]
		0xE5902000, // LDR r2,[r0]
		0xE5D03000, // LDRB r3,[r0]
	}, 5)
	if c.R[2] != 0xFF {
		t.Errorf("LDR r2 = 0x%08X, want 0xFF", c.R[2])
	}
	// Big-endian: 0x000000FF stored MSB-first => byte at 0x2000 is 0x00, at 0x2003 is 0xFF.
	if bus.mem[0x2000] != 0x00 || bus.mem[0x2003] != 0xFF {
		t.Errorf("store not big-endian: [%02X..%02X]", bus.mem[0x2000], bus.mem[0x2003])
	}
	if c.R[3] != 0x00 { // LDRB of the MSB
		t.Errorf("LDRB r3 = 0x%02X, want 0x00 (big-endian MSB)", c.R[3])
	}
}

func TestBlockTransfer(t *testing.T) {
	// MOV r0,#0x2000; MOV r1,#0x11; MOV r2,#0x22; STMIA r0!,{r1,r2}; LDMDB r0!,{r3,r4}
	c, _ := runProg([]uint32{
		0xE3A00A02, // MOV r0,#0x2000
		0xE3A01011, // MOV r1,#0x11
		0xE3A02022, // MOV r2,#0x22
		0xE8A00006, // STMIA r0!,{r1,r2}
		0xE930000C, // LDMDB r0!,{r2,r3}
	}, 5)
	if c.R[2] != 0x11 || c.R[3] != 0x22 {
		t.Errorf("LDMDB got r2=0x%X r3=0x%X, want 0x11,0x22", c.R[2], c.R[3])
	}
	if c.R[0] != 0x2000 {
		t.Errorf("writeback r0 = 0x%X, want 0x2000", c.R[0])
	}
}

func TestBranchAndReturn(t *testing.T) {
	// BL to 0x1010 (a MOV pc,lr that returns), then MOV r5,#7 after the call.
	c, _ := runProg([]uint32{
		0xEB000001, // 0x1000 BL 0x100C
		0xE3A05007, // 0x1004 MOV r5,#7
		0xEAFFFFFE, // 0x1008 B . (spin)
		0xE1A0F00E, // 0x100C MOV pc,lr
	}, 4)
	if c.R[5] != 7 {
		t.Errorf("did not return from BL (r5=%d)", c.R[5])
	}
	if c.R[14] != 0x1004 {
		t.Errorf("BL lr = 0x%08X, want 0x1004", c.R[14])
	}
}

func TestSWIHook(t *testing.T) {
	called := uint32(0xFFFF)
	bus := &testBus{mem: make([]byte, 0x2000)}
	bus.put32(0x1000, 0xEF000042) // SWI #0x42
	c := NewCPU(bus)
	c.Mode = ModeSYS
	c.SetPC(0x1000)
	c.SWI = func(cpu *CPU, comment uint32) bool { called = comment; return true }
	c.Step()
	if called != 0x42 {
		t.Errorf("SWI hook comment = 0x%X, want 0x42", called)
	}
	if c.R[15] != 0x1004 {
		t.Errorf("after serviced SWI, pc = 0x%08X, want 0x1004", c.R[15])
	}
}

func TestModeBankingViaMSR(t *testing.T) {
	c := NewCPU(&testBus{mem: make([]byte, 0x100)})
	c.Mode = ModeSYS
	c.R[13] = 0xAAAA // SYS sp
	// Switch to IRQ mode and give it its own sp, then switch back.
	c.SetCPSR(c.CPSR()&^0x1F | ModeIRQ)
	if c.R[13] == 0xAAAA {
		t.Errorf("IRQ sp should be banked, not the SYS sp")
	}
	c.R[13] = 0xBBBB
	c.SetCPSR(c.CPSR()&^0x1F | ModeSYS)
	if c.R[13] != 0xAAAA {
		t.Errorf("SYS sp = 0x%X after round-trip, want 0xAAAA", c.R[13])
	}
	c.SetCPSR(c.CPSR()&^0x1F | ModeIRQ)
	if c.R[13] != 0xBBBB {
		t.Errorf("IRQ sp = 0x%X, want 0xBBBB", c.R[13])
	}
}

func TestExceptionAndReturn(t *testing.T) {
	c := NewCPU(&testBus{mem: make([]byte, 0x100)})
	c.Mode = ModeSYS
	c.IRQDisable = false
	c.R[15] = 0x1234
	if !c.IRQ() {
		t.Fatal("IRQ not taken with I bit clear")
	}
	if c.Mode != ModeIRQ {
		t.Errorf("mode = 0x%X after IRQ, want IRQ", c.Mode)
	}
	if c.R[15] != 0x18 {
		t.Errorf("IRQ vector pc = 0x%X, want 0x18", c.R[15])
	}
	if c.R[14] != 0x1234+4 {
		t.Errorf("IRQ lr = 0x%X, want 0x1238", c.R[14])
	}
	if !c.IRQDisable {
		t.Errorf("IRQ should be masked inside the handler")
	}
}
