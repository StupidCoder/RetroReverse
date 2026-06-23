package sm83

import "testing"

// flatBus is a 64 KB RAM for unit tests.
type flatBus [0x10000]byte

func (b *flatBus) Read(a uint16) byte     { return b[a] }
func (b *flatBus) Write(a uint16, v byte) { b[a] = v }

// newCPU returns a core at PC=$0000 with flags cleared (not the boot state) so
// individual opcodes are easy to set up.
func newCPU(prog ...byte) (*CPU, *flatBus) {
	bus := &flatBus{}
	copy(bus[:], prog)
	c := NewCPU(bus)
	c.PC, c.SP, c.F = 0, 0xFFFE, 0
	return c, bus
}

func TestALUFlags(t *testing.T) {
	// ADD A,B with half-carry and carry.
	c, _ := newCPU(0x80) // ADD A,B
	c.A, c.B = 0xF8, 0x08
	c.Step()
	if c.A != 0x00 || c.F != flagZ|flagH|flagC {
		t.Errorf("ADD: A=%02X F=%02X, want A=00 F=%02X", c.A, c.F, flagZ|flagH|flagC)
	}
	// SUB producing a borrow.
	c, _ = newCPU(0x90) // SUB B
	c.A, c.B = 0x10, 0x20
	c.Step()
	if c.A != 0xF0 || c.F != flagN|flagC {
		t.Errorf("SUB: A=%02X F=%02X, want A=F0 F=%02X", c.A, c.F, flagN|flagC)
	}
	// AND sets H, clears C.
	c, _ = newCPU(0xA0) // AND B
	c.A, c.B, c.F = 0x0F, 0xF0, flagC
	c.Step()
	if c.A != 0x00 || c.F != flagZ|flagH {
		t.Errorf("AND: A=%02X F=%02X, want A=00 F=%02X", c.A, c.F, flagZ|flagH)
	}
	// INC preserves carry, sets half-carry.
	c, _ = newCPU(0x04) // INC B
	c.B, c.F = 0x0F, flagC
	c.Step()
	if c.B != 0x10 || c.F != flagH|flagC {
		t.Errorf("INC: B=%02X F=%02X, want B=10 F=%02X", c.B, c.F, flagH|flagC)
	}
}

func TestADDHLAndSP(t *testing.T) {
	// ADD HL,DE: Z preserved, H from bit 11, C from bit 15.
	c, _ := newCPU(0x19) // ADD HL,DE
	c.setHL(0x0FFF)
	c.setDE(0x0001)
	c.F = flagZ // must be preserved
	c.Step()
	if c.hl() != 0x1000 || c.F&flagZ == 0 || c.F&flagH == 0 || c.F&flagC != 0 || c.F&flagN != 0 {
		t.Errorf("ADD HL,DE: HL=%04X F=%02X", c.hl(), c.F)
	}
	// ADD SP,e (signed) with low-byte half/carry.
	c, _ = newCPU(0xE8, 0x01) // ADD SP,+1
	c.SP = 0x000F
	c.Step()
	if c.SP != 0x0010 || c.F != flagH {
		t.Errorf("ADD SP,1: SP=%04X F=%02X, want SP=0010 F=%02X", c.SP, c.F, flagH)
	}
	// LD HL,SP-1 shares the flag behaviour and sign-extends e.
	c, _ = newCPU(0xF8, 0xFF) // LD HL,SP-1
	c.SP = 0x0100
	c.Step()
	if c.hl() != 0x00FF {
		t.Errorf("LD HL,SP-1: HL=%04X, want 00FF", c.hl())
	}
}

func TestGBSpecificLoads(t *testing.T) {
	// LD (HL+),A then LD A,(HL-).
	c, bus := newCPU(0x22, 0x2A) // LD (HL+),A ; LD A,(HL+)  -> use 22 then 2A
	c.A, c.H, c.L = 0xAB, 0xC0, 0x00
	c.Step() // (C000)=AB, HL=C001
	if bus[0xC000] != 0xAB || c.hl() != 0xC001 {
		t.Errorf("LD (HL+),A: (C000)=%02X HL=%04X", bus[0xC000], c.hl())
	}
	c.A = 0
	c.setHL(0xC000)
	c.Step() // A=(C000)=AB, HL=C001
	if c.A != 0xAB || c.hl() != 0xC001 {
		t.Errorf("LD A,(HL+): A=%02X HL=%04X", c.A, c.hl())
	}
	// LDH (a8),A writes $FF00+a8.
	c, bus = newCPU(0xE0, 0x80) // LDH ($FF80),A
	c.A = 0x5A
	c.Step()
	if bus[0xFF80] != 0x5A {
		t.Errorf("LDH ($FF80),A: (FF80)=%02X, want 5A", bus[0xFF80])
	}
	// LD ($nnnn),SP.
	c, bus = newCPU(0x08, 0x00, 0xC0) // LD ($C000),SP
	c.SP = 0x1234
	c.Step()
	if bus[0xC000] != 0x34 || bus[0xC001] != 0x12 {
		t.Errorf("LD ($C000),SP: %02X %02X, want 34 12", bus[0xC000], bus[0xC001])
	}
}

func TestCBOps(t *testing.T) {
	// SWAP A.
	c, _ := newCPU(0xCB, 0x37) // SWAP A
	c.A = 0xAB
	c.Step()
	if c.A != 0xBA || c.F != 0 {
		t.Errorf("SWAP A: A=%02X F=%02X, want BA 00", c.A, c.F)
	}
	// BIT 7,A with bit set -> Z clear, H set.
	c, _ = newCPU(0xCB, 0x7F) // BIT 7,A
	c.A, c.F = 0x80, flagC
	c.Step()
	if c.F != flagC|flagH { // Z clear (bit set), H set, C preserved
		t.Errorf("BIT 7,A: F=%02X, want %02X", c.F, flagC|flagH)
	}
	// SET 3,B then RES 3,B.
	c, _ = newCPU(0xCB, 0xD8, 0xCB, 0x98) // SET 3,B ; RES 3,B
	c.B = 0
	c.Step()
	if c.B != 0x08 {
		t.Errorf("SET 3,B: B=%02X, want 08", c.B)
	}
	c.Step()
	if c.B != 0x00 {
		t.Errorf("RES 3,B: B=%02X, want 00", c.B)
	}
}

func TestDAA(t *testing.T) {
	// 0x09 + 0x08 = 0x11 in BCD: ADD then DAA.
	c, _ := newCPU(0x80, 0x27) // ADD A,B ; DAA
	c.A, c.B = 0x09, 0x08
	c.Step() // A=0x11 binary, H set
	c.Step() // DAA -> 0x17 (BCD of 9+8=17)
	if c.A != 0x17 {
		t.Errorf("DAA after 09+08: A=%02X, want 17", c.A)
	}
}

func TestCallRet(t *testing.T) {
	// CALL $0010 ; (at 0010) RET. Check SP and PC and cycle counts.
	c, _ := newCPU(0xCD, 0x10, 0x00)
	c.bus.Write(0x0010, 0xC9) // RET
	cyc := c.Step()           // CALL
	if c.PC != 0x0010 || c.SP != 0xFFFC || cyc != 24 {
		t.Errorf("CALL: PC=%04X SP=%04X cyc=%d, want 0010 FFFC 24", c.PC, c.SP, cyc)
	}
	cyc = c.Step() // RET
	if c.PC != 0x0003 || c.SP != 0xFFFE || cyc != 16 {
		t.Errorf("RET: PC=%04X SP=%04X cyc=%d, want 0003 FFFE 16", c.PC, c.SP, cyc)
	}
}

func TestConditionalCycles(t *testing.T) {
	// JR NZ taken vs not taken: 12 vs 8.
	c, _ := newCPU(0x20, 0x05) // JR NZ,+5
	c.F = 0                    // Z clear -> taken
	if cyc := c.Step(); cyc != 12 {
		t.Errorf("JR NZ taken cyc=%d, want 12", cyc)
	}
	c, _ = newCPU(0x20, 0x05)
	c.F = flagZ // Z set -> not taken
	if cyc := c.Step(); cyc != 8 {
		t.Errorf("JR NZ not taken cyc=%d, want 8", cyc)
	}
}

func TestInterruptDispatch(t *testing.T) {
	c, bus := newCPU(0x00) // NOP at 0
	c.IME = true
	c.PC = 0x0200
	bus[regIE] = 0x01 // VBlank enabled
	bus[regIF] = 0x01 // VBlank requested
	cyc := c.Step()
	if c.PC != 0x0040 || c.IME || bus[regIF]&0x01 != 0 || c.SP != 0xFFFC || cyc != 20 {
		t.Errorf("IRQ: PC=%04X IME=%v IF=%02X SP=%04X cyc=%d", c.PC, c.IME, bus[regIF], c.SP, cyc)
	}
	if got := c.read16(c.SP); got != 0x0200 {
		t.Errorf("IRQ: pushed return=%04X, want 0200", got)
	}
}

func TestEIDelay(t *testing.T) {
	// EI ; NOP ; (IME becomes true after the NOP, not during it).
	c, _ := newCPU(0xFB, 0x00, 0x00) // EI ; NOP ; NOP
	c.IME = false
	c.Step() // EI
	if c.IME {
		t.Errorf("IME should still be false right after EI")
	}
	c.Step() // NOP: IME turns on at the end
	if !c.IME {
		t.Errorf("IME should be true after the instruction following EI")
	}
}

// TestSumProgram runs a small routine: sum of 4 bytes via a loop, to exercise the
// core end to end (loads, DEC, conditional jump, ADD).
func TestSumProgram(t *testing.T) {
	// HL -> data; B = count; A = 0; loop: ADD A,(HL); INC HL; DEC B; JR NZ,loop; HALT
	prog := []byte{
		0x21, 0x00, 0xC0, // LD HL,$C000
		0x06, 0x04, // LD B,4
		0xAF,       // XOR A
		0x86,       // (loop) ADD A,(HL)
		0x23,       // INC HL
		0x05,       // DEC B
		0x20, 0xFB, // JR NZ,loop  (-5, back to the ADD)
		0x76, // HALT
	}
	c, bus := newCPU(prog...)
	copy(bus[0xC000:], []byte{10, 20, 30, 40})
	for i := 0; i < 100 && !c.halt && !c.Halted; i++ {
		c.Step()
	}
	if c.A != 100 {
		t.Errorf("sum program: A=%d, want 100", c.A)
	}
}
