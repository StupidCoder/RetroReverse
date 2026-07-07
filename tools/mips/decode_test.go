package mips

import (
	"encoding/binary"
	"testing"
)

// enc builds the little-endian bytes of a 32-bit instruction word.
func enc(w uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, w)
	return b
}

func TestDecodeGolden(t *testing.T) {
	cases := []struct {
		word   uint32
		addr   uint32
		text   string
		flow   Flow
		target uint32 // checked when flow has a static target
	}{
		// R-type
		{0x00851020, 0, "add $v0, $a0, $a1", FlowSeq, 0},   // add $2,$4,$5
		{0x00851021, 0, "addu $v0, $a0, $a1", FlowSeq, 0},  // addu
		{0x00854022, 0, "sub $t0, $a0, $a1", FlowSeq, 0},   // sub $8,$4,$5
		{0x00a41825, 0, "or $v1, $a1, $a0", FlowSeq, 0},    // or $3,$5,$4
		{0x00041080, 0, "sll $v0, $a0, 2", FlowSeq, 0},     // sll $2,$4,2
		{0x00000000, 0, "nop", FlowSeq, 0},                 // sll $0,$0,0
		{0x0064182a, 0, "slt $v1, $v1, $a0", FlowSeq, 0},   // slt $3,$3,$4
		// I-type
		{0x3c1e8002, 0, "lui $fp, 0x8002", FlowSeq, 0},        // lui $30,0x8002
		{0x24840010, 0, "addiu $a0, $a0, 16", FlowSeq, 0},     // addiu $4,$4,16
		{0x2484fff0, 0, "addiu $a0, $a0, -16", FlowSeq, 0},    // negative imm
		{0x34420001, 0, "ori $v0, $v0, 0x1", FlowSeq, 0},      // ori $2,$2,1
		{0x8c820004, 0, "lw $v0, 4($a0)", FlowSeq, 0},         // lw $2,4($4)
		{0xac820004, 0, "sw $v0, 4($a0)", FlowSeq, 0},         // sw $2,4($4)
		{0xa082fff8, 0, "sb $v0, -8($a0)", FlowSeq, 0},        // sb, negative off
		// branches (delay slot, PC-relative). addr=0x1000.
		{0x10850002, 0x1000, "beq $a0, $a1, $0000100C", FlowBranch, 0x100C},
		{0x14850002, 0x1000, "bne $a0, $a1, $0000100C", FlowBranch, 0x100C},
		{0x1000ffff, 0x1000, "b $00001000", FlowJump, 0x1000}, // beq $0,$0,-1 => b
		{0x1880ffff, 0x1000, "blez $a0, $00001000", FlowBranch, 0x1000},
		{0x0480ffff, 0x1000, "bltz $a0, $00001000", FlowBranch, 0x1000},
		// jumps. addr=0x80010000.
		{0x08004000, 0x80010000, "j $80010000", FlowJump, 0x80010000},
		{0x0c010000, 0x80010000, "jal $80040000", FlowCall, 0x80040000},
		{0x03e00008, 0, "jr $ra", FlowReturn, 0},         // jr $31
		{0x00800008, 0, "jr $a0", FlowIndJump, 0},        // jr $4
		{0x0080f809, 0, "jalr $a0", FlowIndCall, 0},      // jalr $31,$4
		{0x0000000c, 0, "syscall", FlowStop, 0},
		// coprocessors
		{0x40026000, 0, "mfc0 $v0, $12", FlowSeq, 0},  // mfc0 $2, SR (rs=0)
		{0x40886000, 0, "mtc0 $t0, $12", FlowSeq, 0},  // mtc0 $8, SR (rs=4)
		{0x4a000012, 0, "mvmva\t; cop2 0x0000012", FlowSeq, 0}, // GTE command
	}
	for _, c := range cases {
		in := Decode(enc(c.word), c.addr)
		if in.Text != c.text {
			t.Errorf("word 0x%08X @ 0x%X: text = %q, want %q", c.word, c.addr, in.Text, c.text)
		}
		if in.Flow != c.flow {
			t.Errorf("word 0x%08X: flow = %v, want %v", c.word, in.Flow, c.flow)
		}
		if in.HasTarget && in.Target != c.target {
			t.Errorf("word 0x%08X: target = 0x%08X, want 0x%08X", c.word, in.Target, c.target)
		}
		if in.Len != 4 {
			t.Errorf("word 0x%08X: len = %d, want 4", c.word, in.Len)
		}
	}
}

func TestDecodeTruncated(t *testing.T) {
	in := Decode([]byte{0x00, 0x01}, 0)
	if in.Len != 0 || in.Flow != FlowStop {
		t.Errorf("truncated decode = %+v, want Len 0 / FlowStop", in)
	}
}
