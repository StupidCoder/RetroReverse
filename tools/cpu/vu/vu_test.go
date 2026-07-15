package vu

import (
	"math"
	"testing"
)

// The tests assemble tiny programs by hand. The encoders mirror the field layout
// disasm.go documents.

const upNOP = 0x2FF // the wide-group NOP

func pair(upper, lower uint32) []byte {
	return []byte{
		byte(lower), byte(lower >> 8), byte(lower >> 16), byte(lower >> 24),
		byte(upper), byte(upper >> 8), byte(upper >> 16), byte(upper >> 24),
	}
}

func enIADDIU(t, s, imm uint32) uint32 {
	return 0x08<<25 | imm&0x7800<<10 | t<<16 | s<<11 | imm&0x7FF
}
func enSQ(dest, fs, it uint32, imm int32) uint32 {
	return 0x01<<25 | dest<<21 | it<<16 | fs<<11 | uint32(imm)&0x7FF
}
func enB(imm int32) uint32 { return 0x20<<25 | uint32(imm)&0x7FF }
func enUpperI(op, dest, ft, fs, fd uint32) uint32 {
	return dest<<21 | ft<<16 | fs<<11 | fd<<6 | op
}
func enDIV(fsf, ftf, fs, ft uint32) uint32 {
	return 0x40<<25 | ftf<<23 | fsf<<21 | ft<<16 | fs<<11 | 0x38>>2<<6 | 0x3C
}
func enXGKICK(is uint32) uint32 {
	return 0x40<<25 | is<<11 | 0x6C>>2<<6 | 0x3C
}

func newTestVU(prog []byte) *VU {
	micro := make([]byte, 16<<10)
	copy(micro, prog)
	return New(micro, make([]byte, 16<<10))
}

// TestProgramRunsToEBit: load I, multiply VF00 by it, store the result, end.
func TestProgramRunsToEBit(t *testing.T) {
	var prog []byte
	prog = append(prog, pair(upNOP|1<<31, math.Float32bits(2.0))...)  // loi 2.0
	prog = append(prog, pair(upNOP, enIADDIU(1, 0, 3))...)            // vi01 = 3
	prog = append(prog, pair(enUpperI(0x1E, 0xF, 0, 0, 1), 0)...)     // muli.xyzw vf01, vf00, i
	prog = append(prog, pair(upNOP|1<<30, enSQ(0xF, 1, 1, 0))...)     // [E] sq.xyzw vf01, 0(vi01)
	prog = append(prog, pair(upNOP, 0)...)                            // the E delay slot

	v := newTestVU(prog)
	steps, ended := v.Run(0, 1000)
	if !ended {
		t.Fatalf("program did not end by its E bit (%d steps)", steps)
	}
	// VF00 is (0,0,0,1); times 2 -> (0,0,0,2), stored at quadword 3.
	got := math.Float32frombits(le64mLo(v.Data[3*16+12:]))
	if got != 2.0 {
		t.Errorf("stored w = %g, want 2.0", got)
	}
}

// TestBranchDelaySlot: the instruction after a taken branch executes; the one at the
// fall-through target beyond it does not.
func TestBranchDelaySlot(t *testing.T) {
	var prog []byte
	prog = append(prog, pair(upNOP, enIADDIU(1, 0, 1))...) // 0000
	prog = append(prog, pair(upNOP, enB(2))...)            // 0008: b 0x20
	prog = append(prog, pair(upNOP, enIADDIU(2, 0, 5))...) // 0010: delay slot — runs
	prog = append(prog, pair(upNOP, enIADDIU(3, 0, 7))...) // 0018: skipped
	prog = append(prog, pair(upNOP|1<<30, 0)...)           // 0020: [E]
	prog = append(prog, pair(upNOP, 0)...)                 // 0028

	v := newTestVU(prog)
	if _, ended := v.Run(0, 1000); !ended {
		t.Fatal("program did not end")
	}
	if v.VI[2] != 5 {
		t.Errorf("the delay slot did not execute: vi02 = %d, want 5", v.VI[2])
	}
	if v.VI[3] != 0 {
		t.Errorf("the skipped instruction executed: vi03 = %d, want 0", v.VI[3])
	}
}

// TestPairReadsPrePairState: when the upper half writes a register the lower half
// stores, the store sees the value from before the pair.
func TestPairReadsPrePairState(t *testing.T) {
	var prog []byte
	prog = append(prog, pair(upNOP|1<<31, math.Float32bits(3.0))...) // loi 3.0
	// addi.xyzw vf01, vf00, i  |  sq.xyzw vf01, 0(vi00)
	prog = append(prog, pair(enUpperI(0x22, 0xF, 0, 0, 1), enSQ(0xF, 1, 0, 0))...)
	prog = append(prog, pair(upNOP|1<<30, enSQ(0xF, 1, 0, 1))...) // [E] sq the new value
	prog = append(prog, pair(upNOP, 0)...)

	v := newTestVU(prog)
	if _, ended := v.Run(0, 1000); !ended {
		t.Fatal("program did not end")
	}
	if got := math.Float32frombits(le64mLo(v.Data[12:])); got != 0 {
		t.Errorf("same-pair store saw the upper result: w = %g, want 0 (the old vf01)", got)
	}
	if got := math.Float32frombits(le64mLo(v.Data[16+12:])); got != 4.0 {
		t.Errorf("next-pair store w = %g, want 4.0 (vf00.w + 3)", got)
	}
}

// TestDivByZeroSaturates: the DIV contract — a zero divisor gives the largest value of
// the quotient's sign, not an infinity.
func TestDivByZeroSaturates(t *testing.T) {
	var prog []byte
	prog = append(prog, pair(upNOP, enDIV(3, 0, 0, 0))...)        // div q, vf00w (1), vf00x (0)
	prog = append(prog, pair(enUpperI(0x1C, 0xF, 0, 0, 1), 0)...) // mulq.xyzw vf01, vf00, q
	prog = append(prog, pair(upNOP|1<<30, enSQ(0xF, 1, 0, 0))...)
	prog = append(prog, pair(upNOP, 0)...)

	v := newTestVU(prog)
	if _, ended := v.Run(0, 1000); !ended {
		t.Fatal("program did not end")
	}
	got := math.Float32frombits(le64mLo(v.Data[12:]))
	if got != math.MaxFloat32 {
		t.Errorf("1/0 = %g, want MaxFloat32", got)
	}
}

// TestXGKickHandsOverTheAddress: the callback receives the quadword address the
// program names.
func TestXGKickHandsOverTheAddress(t *testing.T) {
	var prog []byte
	prog = append(prog, pair(upNOP, enIADDIU(1, 0, 7))...)
	prog = append(prog, pair(upNOP, enXGKICK(1))...)
	prog = append(prog, pair(upNOP|1<<30, 0)...)
	prog = append(prog, pair(upNOP, 0)...)

	v := newTestVU(prog)
	var kicked []uint32
	v.XGKick = func(qw uint32) { kicked = append(kicked, qw) }
	if _, ended := v.Run(0, 1000); !ended {
		t.Fatal("program did not end")
	}
	if len(kicked) != 1 || kicked[0] != 7 {
		t.Errorf("XGKICK = %v, want [7]", kicked)
	}
}

// le64mLo reads the low 32 bits at b, for pulling one lane out of data memory.
func le64mLo(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
