package x86

import "testing"

// ram is a flat 1 MiB real-mode memory for the CPU tests.
type ram struct{ b []byte }

func (r *ram) Read(a uint32) byte     { return r.b[a&0xFFFFF] }
func (r *ram) Write(a uint32, v byte) { r.b[a&0xFFFFF] = v }

// runProg loads code at CS:IP = 0x1000:0000 (linear 0x10000), sets a stack at
// SS:SP = 0x2000:FFFE and DS=ES=0x4000, and runs until HLT or the step cap.
func runProg(t *testing.T, code []byte) (*CPU, *ram) {
	t.Helper()
	m := &ram{b: make([]byte, 1<<20)}
	copy(m.b[0x10000:], code)
	c := NewCPU(m)
	c.Seg[CS] = 0x1000
	c.Seg[SS] = 0x2000
	c.Seg[DS] = 0x4000
	c.Seg[ES] = 0x4000
	c.IP = 0
	c.sw(SP, 0xFFFE)
	c.Run(10000)
	if !c.Halted {
		t.Fatalf("program did not halt within step cap (IP=%04X)", c.IP)
	}
	return c, m
}

func TestSumLoop(t *testing.T) {
	// MOV CX,10; XOR AX,AX; loop: INC BX; ADD AX,BX; LOOP loop; HLT
	c, _ := runProg(t, []byte{
		0xB9, 0x0A, 0x00, // MOV CX, 10
		0x31, 0xC0, // XOR AX, AX
		0x43,       // INC BX          (offset 5)
		0x01, 0xD8, // ADD AX, BX
		0xE2, 0xFB, // LOOP -5 -> offset 5
		0xF4, // HLT
	})
	if got := c.gw(AX); got != 55 {
		t.Errorf("AX = %d, want 55", got)
	}
	if got := c.gw(BX); got != 10 {
		t.Errorf("BX = %d, want 10", got)
	}
	if got := c.gw(CX); got != 0 {
		t.Errorf("CX = %d, want 0", got)
	}
}

func TestCallRet(t *testing.T) {
	// MOV AX,0; CALL sub; MOV BX,0x2222; HLT; sub: MOV AX,0x1111; RET
	c, _ := runProg(t, []byte{
		0xB8, 0x00, 0x00, // MOV AX, 0
		0xE8, 0x04, 0x00, // CALL +4 -> offset 0x0A
		0xBB, 0x22, 0x22, // MOV BX, 0x2222
		0xF4,             // HLT
		0xB8, 0x11, 0x11, // (0x0A) MOV AX, 0x1111
		0xC3, // RET
	})
	if got := c.gw(AX); got != 0x1111 {
		t.Errorf("AX = $%04X, want $1111", got)
	}
	if got := c.gw(BX); got != 0x2222 {
		t.Errorf("BX = $%04X, want $2222", got)
	}
	if got := c.gw(SP); got != 0xFFFE {
		t.Errorf("SP = $%04X, want $FFFE (balanced)", got)
	}
}

func TestCondJumpAndCmp(t *testing.T) {
	// MOV AX,1; CMP AX,1; JZ skip; MOV BX,0xFFFF; skip: HLT
	c, _ := runProg(t, []byte{
		0xB8, 0x01, 0x00, // MOV AX, 1
		0x3D, 0x01, 0x00, // CMP AX, 1
		0x74, 0x03, // JZ +3 -> HLT
		0xBB, 0xFF, 0xFF, // MOV BX, 0xFFFF (skipped)
		0xF4, // HLT
	})
	if !c.ZF {
		t.Error("ZF should be set after CMP AX,1 == 1")
	}
	if got := c.gw(BX); got != 0 {
		t.Errorf("BX = $%04X, want 0 (branch should skip the MOV)", got)
	}
}

func TestSubFlags(t *testing.T) {
	// MOV AX,3; SUB AX,5; HLT  -> borrow (CF=1), negative (SF=1), AX=0xFFFE
	c, _ := runProg(t, []byte{
		0xB8, 0x03, 0x00, // MOV AX, 3
		0x2D, 0x05, 0x00, // SUB AX, 5
		0xF4,
	})
	if got := c.gw(AX); got != 0xFFFE {
		t.Errorf("AX = $%04X, want $FFFE", got)
	}
	if !c.CF {
		t.Error("CF should be set (borrow)")
	}
	if !c.SF {
		t.Error("SF should be set (negative result)")
	}
	if c.ZF {
		t.Error("ZF should be clear")
	}
}

func TestMemoryRoundTrip(t *testing.T) {
	// MOV AX,0x1234; MOV [0x100],AX; MOV BX,[0x100]; HLT   (DS=0x4000)
	c, m := runProg(t, []byte{
		0xB8, 0x34, 0x12, // MOV AX, 0x1234
		0xA3, 0x00, 0x01, // MOV [0x0100], AX
		0x8B, 0x1E, 0x00, 0x01, // MOV BX, [0x0100]
		0xF4,
	})
	if got := c.gw(BX); got != 0x1234 {
		t.Errorf("BX = $%04X, want $1234", got)
	}
	if m.b[0x40100] != 0x34 || m.b[0x40101] != 0x12 {
		t.Errorf("memory = %02X %02X, want 34 12", m.b[0x40100], m.b[0x40101])
	}
}

func TestRepStosb(t *testing.T) {
	// MOV AX,0x3000; MOV ES,AX; MOV AL,0xAB; MOV CX,4; CLD; REP STOSB; HLT
	c, m := runProg(t, []byte{
		0xB8, 0x00, 0x30, // MOV AX, 0x3000
		0x8E, 0xC0, // MOV ES, AX
		0xB0, 0xAB, // MOV AL, 0xAB
		0xB9, 0x04, 0x00, // MOV CX, 4
		0xFC,       // CLD
		0xF3, 0xAA, // REP STOSB
		0xF4,
	})
	for i := 0; i < 4; i++ {
		if m.b[0x30000+i] != 0xAB {
			t.Errorf("ES:%d = $%02X, want $AB", i, m.b[0x30000+i])
		}
	}
	if m.b[0x30004] != 0 {
		t.Error("STOSB wrote one byte too many")
	}
	if c.gw(DI) != 4 {
		t.Errorf("DI = %d, want 4", c.gw(DI))
	}
	if c.gw(CX) != 0 {
		t.Errorf("CX = %d, want 0", c.gw(CX))
	}
}

func TestMul(t *testing.T) {
	// MOV AL,7; MOV BL,3; MUL BL; HLT  -> AX=21
	c, _ := runProg(t, []byte{
		0xB0, 0x07, // MOV AL, 7
		0xB3, 0x03, // MOV BL, 3
		0xF6, 0xE3, // MUL BL
		0xF4,
	})
	if got := c.gw(AX); got != 21 {
		t.Errorf("AX = %d, want 21", got)
	}
}

func TestPushaPopa(t *testing.T) {
	// Load AX,CX,DX,BX with markers, PUSHA, scribble over them, POPA, HLT.
	c, _ := runProg(t, []byte{
		0xB8, 0x11, 0x11, // MOV AX, 0x1111
		0xB9, 0x22, 0x22, // MOV CX, 0x2222
		0xBA, 0x33, 0x33, // MOV DX, 0x3333
		0xBB, 0x44, 0x44, // MOV BX, 0x4444
		0x60,             // PUSHA
		0xB8, 0xFF, 0xFF, // MOV AX, 0xFFFF (clobber)
		0xB9, 0xFF, 0xFF, // MOV CX, 0xFFFF
		0x61, // POPA (restores AX..BX)
		0xF4, // HLT
	})
	if c.gw(AX) != 0x1111 || c.gw(CX) != 0x2222 || c.gw(DX) != 0x3333 || c.gw(BX) != 0x4444 {
		t.Errorf("PUSHA/POPA didn't restore: AX=%04X CX=%04X DX=%04X BX=%04X",
			c.gw(AX), c.gw(CX), c.gw(DX), c.gw(BX))
	}
	if c.gw(SP) != 0xFFFE {
		t.Errorf("SP = %04X, want FFFE (balanced)", c.gw(SP))
	}
}

func TestIntHook(t *testing.T) {
	// MOV AH,0x30; INT 0x21; HLT — a hook services INT 21h/AH=30h (DOS version).
	m := &ram{b: make([]byte, 1<<20)}
	copy(m.b[0x10000:], []byte{0xB4, 0x30, 0xCD, 0x21, 0xF4})
	c := NewCPU(m)
	c.Seg[CS], c.Seg[SS], c.IP = 0x1000, 0x2000, 0
	c.sw(SP, 0xFFFE)
	c.IntHook = func(cpu *CPU, n byte) bool {
		if n == 0x21 && cpu.g8(4) == 0x30 { // AH=0x30
			cpu.s8(0, 6) // AL=major
			cpu.s8(4, 0) // AH=minor
			return true
		}
		return false
	}
	c.Run(100)
	if !c.Halted {
		t.Fatal("did not halt")
	}
	if got := c.g8(0); got != 6 {
		t.Errorf("AL (DOS major) = %d, want 6", got)
	}
}
