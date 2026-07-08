package arm60

import "testing"

func decodeWord(w uint32, addr uint32) Inst {
	code := []byte{byte(w >> 24), byte(w >> 16), byte(w >> 8), byte(w)} // big-endian
	return Decode(code, addr)
}

func TestDisasmText(t *testing.T) {
	cases := []struct {
		w    uint32
		addr uint32
		text string
		flow Flow
	}{
		{0xE1A00000, 0, "MOV r0, r0", FlowSeq},
		{0xE3A00005, 0, "MOV r0, #5", FlowSeq},
		{0xE2801003, 0, "ADD r1, r0, #3", FlowSeq},
		{0xE1A03102, 0, "MOV r3, r2, LSL #2", FlowSeq},
		{0xE3500000, 0, "CMP r0, #0", FlowSeq},
		{0xE5901000, 0, "LDR r1, [r0]", FlowSeq},
		{0xE92D4000, 0, "STMDB sp!, {lr}", FlowSeq},
		{0xEB000001, 0x1000, "BL 0x0000100C", FlowCall},
		{0xEA000001, 0x1000, "B 0x0000100C", FlowJump},
		{0x1A000001, 0x1000, "BNE 0x0000100C", FlowBranch},
		{0xE1A0F00E, 0, "MOV pc, lr", FlowReturn},           // canonical return
		{0xE8BD8000, 0, "LDMIA sp!, {pc}", FlowReturn},      // pop pc
		{0xE10F0000, 0, "MRS r0, CPSR", FlowSeq},            // ARMv3 PSR transfer
		{0xE121F000, 0, "MSR CPSR_c, r0", FlowSeq},
		{0xEF000011, 0, "SWI #0x11", FlowSeq},
	}
	for _, c := range cases {
		in := decodeWord(c.w, c.addr)
		if in.Text != c.text {
			t.Errorf("0x%08X text = %q, want %q", c.w, in.Text, c.text)
		}
		if in.Flow != c.flow {
			t.Errorf("0x%08X flow = %d, want %d", c.w, in.Flow, c.flow)
		}
	}
}

func TestNeverConditionNeutralized(t *testing.T) {
	// A "never" branch (cond 1111) must not be followed as control flow.
	in := decodeWord(0xFA000001, 0x1000) // BNV
	if in.HasTarget {
		t.Errorf("NV branch should not expose a trace target")
	}
	if in.Flow != FlowSeq {
		t.Errorf("NV branch flow = %d, want FlowSeq", in.Flow)
	}
}

func TestArmv3HasNoBX(t *testing.T) {
	// The ARMv4 BX encoding decodes as data on the ARM60, not as a branch.
	in := decodeWord(0xE12FFF1E, 0) // "BX lr" on ARMv4
	if in.Flow != FlowStop {
		t.Errorf("BX encoding flow = %d, want FlowStop (undefined on ARMv3)", in.Flow)
	}
}
