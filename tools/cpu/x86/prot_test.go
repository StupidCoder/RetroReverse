package x86

import "testing"

// flatRAM is a flat 16 MiB protected-mode memory: no real-mode 20-bit wrap, so
// a test can prove the core reaches linear addresses above 1 MiB.
type flatRAM struct{ b []byte }

func (r *flatRAM) Read(a uint32) byte     { return r.b[a] }
func (r *flatRAM) Write(a uint32, v byte) { r.b[a] = v }

// runProt loads code at flat linear `entry`, puts a 32-bit stack at ESP=stack
// (flat SS base 0), sets flat DS/ES/CS/SS (all base 0), enters ModeProt, and
// runs until HLT or the step cap.
func runProt(t *testing.T, entry, stack uint32, code []byte) (*CPU, *flatRAM) {
	t.Helper()
	m := &flatRAM{b: make([]byte, 16<<20)}
	copy(m.b[entry:], code)
	c := NewCPU(m)
	c.Mode = ModeProt
	// Flat selectors: bases all 0. (Selector values are cosmetic in flat PM.)
	c.Seg[CS], c.Seg[DS], c.Seg[ES], c.Seg[SS] = 0x08, 0x10, 0x10, 0x10
	c.IP = entry
	c.Regs[SP] = stack
	c.Run(200000)
	if !c.Halted {
		t.Fatalf("program did not halt within step cap (EIP=%08X)", c.IP)
	}
	return c, m
}

// TestProtFlat32 exercises a 32-bit MOV to a linear address above 1 MiB (proving
// the real-mode 0xFFFFF mask is gone), a 32-bit CALL/RET on an ESP stack, and a
// 32-bit register round-trip.
func TestProtFlat32(t *testing.T) {
	// entry 0x100000 (1 MiB):
	//   MOV EAX, 0xDEADBEEF
	//   MOV [0x00200000], EAX     ; write at 2 MiB — impossible under 20-bit wrap
	//   CALL sub                  ; 32-bit near call, ESP stack
	//   MOV EBX, [0x00200000]
	//   HLT
	// sub:
	//   ADD EAX, 1
	//   RET
	code := []byte{
		0xB8, 0xEF, 0xBE, 0xAD, 0xDE, // MOV EAX, 0xDEADBEEF
		0xA3, 0x00, 0x00, 0x20, 0x00, // MOV [0x00200000], EAX
		0xE8, 0x07, 0x00, 0x00, 0x00, // CALL +7 -> sub
		0x8B, 0x1D, 0x00, 0x00, 0x20, 0x00, // MOV EBX, [0x00200000]
		0xF4,             // HLT
		0x83, 0xC0, 0x01, // (sub) ADD EAX, 1
		0xC3, // RET
	}
	c, m := runProt(t, 0x100000, 0x00300000, code)
	if got := c.Regs[AX]; got != 0xDEADBEF0 {
		t.Errorf("EAX = %08X, want DEADBEF0 (0xDEADBEEF + 1 from sub)", got)
	}
	if got := c.Regs[BX]; got != 0xDEADBEEF {
		t.Errorf("EBX = %08X, want DEADBEEF (read back from 2 MiB)", got)
	}
	// The store really landed at linear 2 MiB.
	if lo := m.b[0x200000]; lo != 0xEF {
		t.Errorf("mem[0x200000] = %02X, want EF", lo)
	}
	// ESP returned to its start after the balanced CALL/RET.
	if c.Regs[SP] != 0x00300000 {
		t.Errorf("ESP = %08X, want 00300000 (balanced call/ret)", c.Regs[SP])
	}
}

// TestProtRepStosd reproduces the go32 crt0's very first act: REP STOSD to clear
// a BSS region above 1 MiB with ECX/EDI (32-bit address size).
func TestProtRepStosd(t *testing.T) {
	// Pre-fill the target so we can see it get cleared.
	// entry:
	//   MOV EDI, 0x00201000
	//   MOV ECX, 0x00000010   ; 16 dwords = 64 bytes
	//   XOR EAX, EAX
	//   CLD
	//   REP STOSD
	//   HLT
	code := []byte{
		0xBF, 0x00, 0x10, 0x20, 0x00, // MOV EDI, 0x00201000
		0xB9, 0x10, 0x00, 0x00, 0x00, // MOV ECX, 16
		0x31, 0xC0, // XOR EAX, EAX
		0xFC,       // CLD
		0xF3, 0xAB, // REP STOSD
		0xF4, // HLT
	}
	m := &flatRAM{b: make([]byte, 16<<20)}
	for i := 0; i < 64; i++ {
		m.b[0x201000+i] = 0xAA
	}
	copy(m.b[0x100000:], code)
	c := NewCPU(m)
	c.Mode = ModeProt
	c.IP = 0x100000
	c.Regs[SP] = 0x00300000
	c.Run(200000)
	if !c.Halted {
		t.Fatalf("did not halt (EIP=%08X)", c.IP)
	}
	for i := 0; i < 64; i++ {
		if m.b[0x201000+i] != 0 {
			t.Fatalf("byte %d not cleared: %02X", i, m.b[0x201000+i])
		}
	}
	if c.Regs[CX] != 0 {
		t.Errorf("ECX = %08X, want 0 after REP", c.Regs[CX])
	}
	if c.Regs[DI] != 0x00201000+64 {
		t.Errorf("EDI = %08X, want %08X", c.Regs[DI], 0x00201000+64)
	}
}

// TestProt66Prefix checks that in protected mode a 0x66 prefix selects 16-bit
// operands (the inverse of real mode), and a bare opcode uses 32-bit.
func TestProt66Prefix(t *testing.T) {
	// entry:
	//   MOV EAX, 0x11112222      ; 32-bit default (B8 imm32)
	//   66 B8 3333               ; MOV AX, 0x3333  (16-bit via 0x66) -> EAX=0x11113333
	//   HLT
	code := []byte{
		0xB8, 0x22, 0x22, 0x11, 0x11, // MOV EAX, 0x11112222
		0x66, 0xB8, 0x33, 0x33, // MOV AX, 0x3333
		0xF4, // HLT
	}
	c, _ := runProt(t, 0x100000, 0x00300000, code)
	if got := c.Regs[AX]; got != 0x11113333 {
		t.Errorf("EAX = %08X, want 11113333 (0x66 selected 16-bit MOV)", got)
	}
}
