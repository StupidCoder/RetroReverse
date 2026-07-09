package r4300

import (
	"encoding/binary"
	"testing"
)

func decodeAt(w, addr uint32) Inst { return DecodeWord(w, addr) }

func TestDecodeText(t *testing.T) {
	cases := []struct {
		w    uint32
		addr uint32
		text string
	}{
		{0x00000000, 0x80000000, "nop"},
		{0x27BDFFE8, 0x80000000, "addiu $sp, $sp, -24"},
		{0x3C088000, 0x80000000, "lui $t0, 0x8000"},
		{0x8FBF0014, 0x80000000, "lw $ra, 20($sp)"},
		{0x03E00008, 0x80000000, "jr $ra"},
		{0x0320F809, 0x80000000, "jalr $t9"},
		// MIPS III additions.
		{0x0064182D, 0x80000000, "daddu $v1, $v1, $a0"},
		{0x00041438, 0x80000000, "dsll $v0, $a0, 16"},
		{0x0004143C, 0x80000000, "dsll32 $v0, $a0, 16"},
		{0xDC820000, 0x80000000, "ld $v0, 0($a0)"},
		{0xFC820008, 0x80000000, "sd $v0, 8($a0)"},
		{0x9C820000, 0x80000000, "lwu $v0, 0($a0)"},
		// Branch-likely.
		{0x50400003, 0x80000000, "beql $v0, $zero, $80000010"},
		{0x54400003, 0x80000000, "bnel $v0, $zero, $80000010"},
		// COP0 and the TLB.
		{0x40086000, 0x80000000, "mfc0 $t0, Status"},
		{0x40088000, 0x80000000, "mfc0 $t0, Config"},
		{0x42000018, 0x80000000, "eret"},
		{0x42000002, 0x80000000, "tlbwi"},
		// COP1.
		{0x44810000, 0x80000000, "mtc1 $at, $f0"},
		{0x46000000, 0x80000000, "add.s $f0, $f0, $f0"},
		{0x4600018D, 0x80000000, "trunc.w.s $f6, $f0"},
		{0x46201032, 0x80000000, "c.eq.d $f2, $f0"},
		{0x45010003, 0x80000000, "bc1t $80000010"},
		// cache must decode, not fall through to .word.
		{0xBC010000, 0x80000000, "cache 0x1, 0($zero)"},
	}
	for _, c := range cases {
		if got := decodeAt(c.w, c.addr).Text; got != c.text {
			t.Errorf("0x%08X: got %q want %q", c.w, got, c.text)
		}
	}
}

// The tracer classifies control flow from Flow and Annul alone, so both must be
// right for every transfer instruction.
func TestDecodeFlow(t *testing.T) {
	cases := []struct {
		w     uint32
		flow  Flow
		delay bool
		annul bool
	}{
		{0x00000000, FlowSeq, false, false},    // nop
		{0x08000000, FlowJump, true, false},    // j
		{0x0C000000, FlowCall, true, false},    // jal
		{0x03E00008, FlowReturn, true, false},  // jr $ra
		{0x01000008, FlowIndJump, true, false}, // jr $t0
		{0x0320F809, FlowIndCall, true, false}, // jalr $t9
		{0x10000001, FlowJump, true, false},    // beq $0,$0 == b
		{0x14400001, FlowBranch, true, false},  // bne
		{0x50400001, FlowBranch, true, true},   // beql
		{0x54400001, FlowBranch, true, true},   // bnel
		{0x04620001, FlowBranch, true, true},   // bltzl
		{0x04710001, FlowCall, true, false},    // bgezal:  links, so a call
		{0x04720001, FlowCall, true, true},     // bgezall: links and annuls
		{0x45030001, FlowBranch, true, true},   // bc1tl
		{0x0000000C, FlowStop, false, false},   // syscall
		{0x42000018, FlowStop, false, false},   // eret
		{0x4C000000, FlowStop, false, false},   // COP3: .word
	}
	for _, c := range cases {
		in := decodeAt(c.w, 0x80000000)
		if in.Flow != c.flow || in.HasDelay != c.delay || in.Annul != c.annul {
			t.Errorf("0x%08X (%s): flow=%v delay=%v annul=%v; want flow=%v delay=%v annul=%v",
				c.w, in.Mnem, in.Flow, in.HasDelay, in.Annul, c.flow, c.delay, c.annul)
		}
	}
}

func TestDecodeTargets(t *testing.T) {
	// A jump's target takes its high nibble from the delay slot's address.
	in := decodeAt(0x08000100, 0x80000000) // j 0x400
	if !in.HasTarget || in.Target != 0x80000400 {
		t.Errorf("j target = %08X want 0x80000400", in.Target)
	}
	// A branch is relative to the delay slot.
	in = decodeAt(0x1000FFFE, 0x80000010) // b -2 -> back to itself
	if in.Target != 0x8000000C {
		t.Errorf("branch target = %08X want 0x8000000C", in.Target)
	}
}

func TestDecodeIsBigEndian(t *testing.T) {
	// The instruction stream is big-endian: the same bytes read the other way
	// round would decode as something else entirely.
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, 0x27BDFFE8)
	if got := Decode(b, 0x80000000).Mnem; got != "addiu" {
		t.Errorf("big-endian decode: got %q want addiu", got)
	}
}

func TestDecodeShortSlice(t *testing.T) {
	in := Decode([]byte{0x27, 0xBD}, 0x80000000)
	if in.Len != 0 || in.Flow != FlowStop {
		t.Errorf("short slice: Len=%d Flow=%v want 0 / stop", in.Len, in.Flow)
	}
}
