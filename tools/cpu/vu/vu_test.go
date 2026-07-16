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

// TestDisasmMacro pins the macro-mode dispatcher's decode against words lifted from
// Jak and Daxter's generic renderer (hand-decoded once, then trusted): a broadcast
// multiply-accumulate chain, a plain FMAC op, an integer op, and a lower-special.
func TestDisasmMacro(t *testing.T) {
	cases := []struct{ w uint32; want string }{
		{0x4BE149BE, "vmulaz.xyzw acc, vf09, vf01z"},  // ACC = vf9 * vf1.z
		{0x4BE150BC, "vmaddax.xyzw acc, vf10, vf01x"}, // ACC += vf10 * vf1.x
		{0x4BE15849, "vmaddy.xyzw vf01, vf11, vf01y"}, // vf1 = ACC + vf11 * vf1.y
		{0x4A0DA030, "viadd vi00, vi20, vi13"},
		{0x4A005BBC, "vdiv q, vf11x, vf00x"},
	}
	for _, c := range cases {
		if got := DisasmMacro(c.w); got != c.want {
			t.Errorf("DisasmMacro(%08X) = %q, want %q", c.w, got, c.want)
		}
	}
}

// TestMacroIntegerOps: the macro face routes the integer group (0x30..0x37) to the
// same executor micro mode uses — a dropped VIADD is silent and poisons every
// address computed from it.
func TestMacroIntegerOps(t *testing.T) {
	v := New(make([]byte, 16<<10), make([]byte, 16<<10))
	v.VI[1] = 7
	v.VI[2] = 5
	// viadd vi3, vi1, vi2 : COP2 macro form, op 0x30, fd=3, fs=1, ft=2.
	v.Macro(1<<25 | 2<<16 | 1<<11 | 3<<6 | 0x30)
	if v.VI[3] != 12 {
		t.Errorf("macro viadd: vi3 = %d, want 12", v.VI[3])
	}
	// visub vi4, vi1, vi2
	v.Macro(1<<25 | 2<<16 | 1<<11 | 4<<6 | 0x31)
	if v.VI[4] != 2 {
		t.Errorf("macro visub: vi4 = %d, want 2", v.VI[4])
	}
}
