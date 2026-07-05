package x86

import "testing"

func TestDecode(t *testing.T) {
	cases := []struct {
		name   string
		code   []byte
		addr   uint32
		len    int
		mnem   string
		text   string
		flow   Flow
		hasTgt bool
		target uint32
	}{
		{"nop", []byte{0x90}, 0x100, 1, "NOP", "NOP", FlowSeq, false, 0},
		{"mov ax,imm16", []byte{0xB8, 0x34, 0x12}, 0x100, 3, "MOV", "MOV AX, $1234", FlowSeq, false, 0},
		{"mov eax,imm32 (66)", []byte{0x66, 0xB8, 0x78, 0x56, 0x34, 0x12}, 0x100, 6, "MOV", "MOV EAX, $12345678", FlowSeq, false, 0},
		{"mov ax,[bp-2]", []byte{0x8B, 0x46, 0xFE}, 0x100, 3, "MOV", "MOV AX, [BP-$2]", FlowSeq, false, 0},
		{"mov [bx],ax", []byte{0x89, 0x07}, 0x100, 2, "MOV", "MOV [BX], AX", FlowSeq, false, 0},
		{"mov ax,[es:bx]", []byte{0x26, 0x8B, 0x07}, 0x100, 3, "MOV", "MOV AX, [ES:BX]", FlowSeq, false, 0},
		{"add ax,imm16", []byte{0x05, 0x34, 0x12}, 0x100, 3, "ADD", "ADD AX, $1234", FlowSeq, false, 0},
		{"add bx,imm16 (grp1)", []byte{0x81, 0xC3, 0x34, 0x12}, 0x100, 4, "ADD", "ADD BX, $1234", FlowSeq, false, 0},
		{"add bx,imm8sx (grp1/83)", []byte{0x83, 0xC3, 0x05}, 0x100, 3, "ADD", "ADD BX, $0005", FlowSeq, false, 0},
		{"mov byte [imm16],imm8", []byte{0xC6, 0x06, 0x00, 0x20, 0x41}, 0x100, 5, "MOV", "MOV BYTE [$2000], $41", FlowSeq, false, 0},
		{"neg ax (grp3)", []byte{0xF7, 0xD8}, 0x100, 2, "NEG", "NEG AX", FlowSeq, false, 0},
		{"shl ax,1 (grp2)", []byte{0xD1, 0xE0}, 0x100, 2, "SHL", "SHL AX, 1", FlowSeq, false, 0},
		{"movzx ax,bl", []byte{0x0F, 0xB6, 0xC3}, 0x100, 3, "MOVZX", "MOVZX AX, BL", FlowSeq, false, 0},
		{"stosb", []byte{0xAA}, 0x100, 1, "STOSB", "STOSB", FlowSeq, false, 0},
		{"rep stosb", []byte{0xF3, 0xAA}, 0x100, 2, "STOSB", "REP STOSB", FlowSeq, false, 0},
		{"sib mov ax,[ebx+ecx*4] (67)", []byte{0x67, 0x8B, 0x04, 0x8B}, 0x100, 4, "MOV", "MOV AX, [EBX+ECX*4]", FlowSeq, false, 0},

		// control flow
		{"jmp short -2", []byte{0xEB, 0xFE}, 0x100, 2, "JMP", "JMP $00000100", FlowJump, true, 0x100},
		{"jmp rel16 -3", []byte{0xE9, 0xFD, 0xFF}, 0x100, 3, "JMP", "JMP $00000100", FlowJump, true, 0x100},
		{"call rel16 +0", []byte{0xE8, 0x00, 0x00}, 0x100, 3, "CALL", "CALL $00000103", FlowCall, true, 0x103},
		{"jz rel8 +16", []byte{0x74, 0x10}, 0x100, 2, "JZ", "JZ $00000112", FlowBranch, true, 0x112},
		{"jz near rel16 +16", []byte{0x0F, 0x84, 0x10, 0x00}, 0x100, 4, "JZ", "JZ $00000114", FlowBranch, true, 0x114},
		{"loop -2", []byte{0xE2, 0xFE}, 0x100, 2, "LOOP", "LOOP $00000100", FlowBranch, true, 0x100},
		{"ret", []byte{0xC3}, 0x100, 1, "RET", "RET", FlowReturn, false, 0},
		{"ret imm16", []byte{0xC2, 0x04, 0x00}, 0x100, 3, "RET", "RET $0004", FlowReturn, false, 0},
		{"retf", []byte{0xCB}, 0x100, 1, "RETF", "RETF", FlowReturn, false, 0},
		{"iret", []byte{0xCF}, 0x100, 1, "IRET", "IRET", FlowReturn, false, 0},
		{"int 21h", []byte{0xCD, 0x21}, 0x100, 2, "INT", "INT $21", FlowSeq, false, 0},
		{"jmp far ptr", []byte{0xEA, 0x00, 0x10, 0x00, 0x20}, 0x100, 5, "JMPF", "JMPF $2000:$1000", FlowJump, true, 0x21000},
		{"call far ptr", []byte{0x9A, 0x00, 0x10, 0x00, 0x20}, 0x100, 5, "CALLF", "CALLF $2000:$1000", FlowCall, true, 0x21000},
		{"jmp near [imm16] (grp5/4)", []byte{0xFF, 0x26, 0x34, 0x12}, 0x100, 4, "JMP", "JMP [$1234]", FlowIndJump, false, 0},
		{"call near [imm16] (grp5/2)", []byte{0xFF, 0x16, 0x34, 0x12}, 0x100, 4, "CALL", "CALL [$1234]", FlowCall, false, 0},
		{"hlt", []byte{0xF4}, 0x100, 1, "HLT", "HLT", FlowStop, false, 0},

		// x87 length correctness
		{"fld dword [bp-4]", []byte{0xD9, 0x46, 0xFC}, 0x100, 3, "FLD", "FLD DWORD [BP-$4]", FlowSeq, false, 0},
		{"faddp st(1),st", []byte{0xDE, 0xC1}, 0x100, 2, "FADDP", "FADDP ST(1), ST", FlowSeq, false, 0},

		// unknown 0F stays 2 bytes so a trace stays aligned
		{"unknown 0F", []byte{0x0F, 0xFF}, 0x100, 2, ".byte", ".byte $0F,$FF", FlowStop, false, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := Decode(c.code, c.addr)
			if in.Len != c.len {
				t.Errorf("Len = %d, want %d", in.Len, c.len)
			}
			if in.Mnem != c.mnem {
				t.Errorf("Mnem = %q, want %q", in.Mnem, c.mnem)
			}
			if in.Text != c.text {
				t.Errorf("Text = %q, want %q", in.Text, c.text)
			}
			if in.Flow != c.flow {
				t.Errorf("Flow = %d, want %d", in.Flow, c.flow)
			}
			if in.HasTarget != c.hasTgt {
				t.Errorf("HasTarget = %v, want %v", in.HasTarget, c.hasTgt)
			}
			if c.hasTgt && in.Target != c.target {
				t.Errorf("Target = $%X, want $%X", in.Target, c.target)
			}
		})
	}
}

// TestDecodeAlignment feeds a hand-assembled straight-line routine and checks
// that decoding stays byte-aligned across every instruction to the final RET.
func TestDecodeAlignment(t *testing.T) {
	// push bp; mov bp,sp; sub sp,4; mov ax,[bp+4]; add ax,1; mov [bp-2],ax;
	// mov sp,bp; pop bp; ret
	code := []byte{
		0x55,       // PUSH BP
		0x8B, 0xEC, // MOV BP, SP
		0x83, 0xEC, 0x04, // SUB SP, $0004
		0x8B, 0x46, 0x04, // MOV AX, [BP+$4]
		0x05, 0x01, 0x00, // ADD AX, $0001
		0x89, 0x46, 0xFE, // MOV [BP-$2], AX
		0x8B, 0xE5, // MOV SP, BP
		0x5D, // POP BP
		0xC3, // RET
	}
	want := []string{
		"PUSH BP", "MOV BP, SP", "SUB SP, $0004", "MOV AX, [BP+$4]",
		"ADD AX, $0001", "MOV [BP-$2], AX", "MOV SP, BP", "POP BP", "RET",
	}
	off, i := 0, 0
	for off < len(code) {
		in := Decode(code[off:], uint32(off))
		if i >= len(want) {
			t.Fatalf("decoded more instructions than expected at off %d", off)
		}
		if in.Text != want[i] {
			t.Errorf("instr %d: Text = %q, want %q", i, in.Text, want[i])
		}
		if in.Flow == FlowStop && in.Mnem == ".byte" {
			t.Fatalf("desync at off %d: %q", off, in.Text)
		}
		off += in.Len
		i++
	}
	if i != len(want) {
		t.Errorf("decoded %d instructions, want %d", i, len(want))
	}
}
