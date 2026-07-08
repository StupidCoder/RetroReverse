package arm

import "testing"

// leWord packs a 32-bit ARM instruction into little-endian bytes for a test case.
func leWord(w uint32) []byte {
	return []byte{byte(w), byte(w >> 8), byte(w >> 16), byte(w >> 24)}
}

func TestDecodeARM(t *testing.T) {
	cases := []struct {
		w    uint32
		addr uint32
		text string
		flow Flow
		tgt  uint32
	}{
		// data processing
		{0xE1A00000, 0, "MOV r0, r0", FlowSeq, 0}, // the canonical ARM NOP
		{0xE3A00001, 0, "MOV r0, #1", FlowSeq, 0},
		{0xE2811004, 0, "ADD r1, r1, #4", FlowSeq, 0},
		{0xE0802001, 0, "ADD r2, r0, r1", FlowSeq, 0},
		{0xE0812102, 0, "ADD r2, r1, r2, LSL #2", FlowSeq, 0},
		{0xE3500000, 0, "CMP r0, #0", FlowSeq, 0},
		{0xE1500001, 0, "CMP r0, r1", FlowSeq, 0},
		{0xE2400001, 0, "SUB r0, r0, #1", FlowSeq, 0},
		{0xE1B01002, 0, "MOVS r1, r2", FlowSeq, 0},
		// multiply
		{0xE0010293, 0, "MUL r1, r3, r2", FlowSeq, 0},
		{0xE0214392, 0, "MLA r1, r2, r3, r4", FlowSeq, 0},
		// load / store
		{0xE5901000, 0, "LDR r1, [r0]", FlowSeq, 0},
		{0xE5801004, 0, "STR r1, [r0, #0x4]", FlowSeq, 0},
		{0xE7901002, 0, "LDR r1, [r0, r2]", FlowSeq, 0},
		{0xE1D010B4, 0, "LDRH r1, [r0, #0x4]", FlowSeq, 0},
		{0xE51FF004, 0x100, "LDR pc, [pc, #-0x4]", FlowIndJump, 0}, // computed jump
		// block transfer
		{0xE92D4010, 0, "STMDB sp!, {r4, lr}", FlowSeq, 0},    // PUSH {r4, lr}
		{0xE8BD8010, 0, "LDMIA sp!, {r4, pc}", FlowReturn, 0}, // POP {r4, pc}
		// branches (addr chosen so targets are tidy)
		{0xEAFFFFFE, 0x100, "B 0x00000100", FlowJump, 0x100}, // B to self
		{0xEBFFFFFE, 0x100, "BL 0x00000100", FlowCall, 0x100},
		{0x1AFFFFFD, 0x100, "BNE 0x000000FC", FlowBranch, 0xFC},
		{0xEA000000, 0x100, "B 0x00000108", FlowJump, 0x108},
		// interworking
		{0xE12FFF1E, 0, "BX lr", FlowReturn, 0},
		{0xE12FFF11, 0, "BX r1", FlowIndJump, 0},
		{0xE12FFF31, 0, "BLX r1", FlowIndCall, 0},
		{0xFAFFFFFE, 0x100, "BLX 0x00000100", FlowCall, 0x100}, // BLX <imm> → Thumb
		// system / misc
		{0xEF000000, 0, "SWI #0x0", FlowSeq, 0},
		{0xE10F0000, 0, "MRS r0, CPSR", FlowSeq, 0},
		{0xE16F0F11, 0, "CLZ r0, r1", FlowSeq, 0},
	}
	for _, c := range cases {
		got := DecodeARM(leWord(c.w), c.addr)
		if got.Text != c.text || got.Flow != c.flow {
			t.Errorf("DecodeARM(%08X)@%X = {%q flow=%d}, want {%q flow=%d}",
				c.w, c.addr, got.Text, got.Flow, c.text, c.flow)
		}
		if got.HasTarget && got.Target != c.tgt {
			t.Errorf("DecodeARM(%08X) target = %08X, want %08X", c.w, got.Target, c.tgt)
		}
	}
}

func TestDecodeThumb(t *testing.T) {
	cases := []struct {
		bytes []byte
		addr  uint32
		text  string
		len   int
		flow  Flow
		tgt   uint32
	}{
		{[]byte{0x00, 0x20}, 0, "MOV r0, #0x0", 2, FlowSeq, 0},
		{[]byte{0x01, 0x30}, 0, "ADD r0, #0x1", 2, FlowSeq, 0},
		{[]byte{0x40, 0x18}, 0, "ADD r0, r0, r1", 2, FlowSeq, 0},
		{[]byte{0x08, 0x1C}, 0, "ADD r0, r1, #0", 2, FlowSeq, 0}, // ADD Rd,Rs,#0 (MOV alias)
		{[]byte{0x88, 0x42}, 0, "CMP r0, r1", 2, FlowSeq, 0},
		{[]byte{0x01, 0x48}, 0, "LDR r0, [pc, #0x4]", 2, FlowSeq, 0},
		{[]byte{0x00, 0xB5}, 0, "PUSH {lr}", 2, FlowSeq, 0},
		{[]byte{0x10, 0xB5}, 0, "PUSH {r4, lr}", 2, FlowSeq, 0},
		{[]byte{0x00, 0xBD}, 0, "POP {pc}", 2, FlowReturn, 0},
		{[]byte{0x70, 0x47}, 0, "BX lr", 2, FlowReturn, 0},
		{[]byte{0x88, 0x47}, 0, "BLX r1", 2, FlowIndCall, 0},
		{[]byte{0x00, 0xD0}, 0, "BEQ 0x00000004", 2, FlowBranch, 4},
		{[]byte{0xFE, 0xE7}, 0, "B 0x00000000", 2, FlowJump, 0},
		{[]byte{0x00, 0xDF}, 0, "SWI #0x0", 2, FlowSeq, 0},
		// BL pair (4 bytes)
		{[]byte{0x00, 0xF0, 0x01, 0xF8}, 0, "BL 0x00000006", 4, FlowCall, 6},
		// BLX pair (4 bytes, switches to ARM)
		{[]byte{0x00, 0xF0, 0x02, 0xE8}, 0, "BLX 0x00000008", 4, FlowCall, 8},
	}
	for _, c := range cases {
		got := DecodeThumb(c.bytes, c.addr)
		if got.Text != c.text || got.Len != c.len || got.Flow != c.flow {
			t.Errorf("DecodeThumb(%X)@%X = {%q len=%d flow=%d}, want {%q len=%d flow=%d}",
				c.bytes, c.addr, got.Text, got.Len, got.Flow, c.text, c.len, c.flow)
		}
		if got.HasTarget && got.Target != c.tgt {
			t.Errorf("DecodeThumb(%X) target = %08X, want %08X", c.bytes, got.Target, c.tgt)
		}
	}
}
