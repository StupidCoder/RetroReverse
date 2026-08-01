package arm

import (
	"encoding/binary"
	"testing"
)

// The ARMv5 encodings that do not exist on the ARM7TDMI, with the mnemonic the
// V5TE decoder produces for each.
var v5Only = []struct {
	w    uint32
	name string
}{
	{0xE12FFF30, "BLX"},  // BLX r0
	{0xFA000000, "BLX"},  // BLX <imm> (unconditional space)
	{0xE16F0F10, "CLZ"},  // CLZ r0, r0
	{0xE1000050, "QADD"}, // QADD r0, r0, r0
	{0xE1000080, "SMLA"}, // SMLABB r0, r0, r0, r0
	{0xE1C000D0, "LDRD"}, // LDRD r0, [r0]
	{0xE1C000F0, "STRD"}, // STRD r0, [r0]
	{0xE1200070, "BKPT"}, // BKPT #0
	{0xF5D0F000, "PLD"},  // PLD [r0]
}

// TestV4TDecodeRejectsV5 checks the DECODER: on V4T each v5-only word must come
// back as an explicit undefined word, not as the instruction a v5 chip runs.
func TestV4TDecodeRejectsV5(t *testing.T) {
	for _, c := range v5Only {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], c.w)

		// The control: a V5TE decode must RECOGNISE it. Without this the test
		// would pass just as happily against a decoder that understood nothing.
		if in := DecodeARMVariant(b[:], 0x08000000, V5TE); in.Mnem == ".word" {
			t.Fatalf("control failed: V5TE does not decode 0x%08X (%s) either", c.w, c.name)
		}
		if in := DecodeARMVariant(b[:], 0x08000000, V4T); in.Mnem != ".word" {
			t.Errorf("V4T decoded 0x%08X as %q; want an undefined word", c.w, in.Text)
		}
	}
}

// stubBus is flat RAM for the execution tests.
type stubBus struct{ mem [256]byte }

func (b *stubBus) Read(a uint32) byte     { return b.mem[a&0xFF] }
func (b *stubBus) Write(a uint32, v byte) { b.mem[a&0xFF] = v }

// TestV4TExecRejectsV5 checks the CORE: executing a v5-only encoding on a V4T
// core must halt with a reason, not quietly do the v5 thing.
func TestV4TExecRejectsV5(t *testing.T) {
	// Each run gets a FRESH bus. Sharing one is not a tidiness point: STRD's
	// control run stores through r0 == 0 and overwrites the very word under
	// test, so the V4T core then decodes a zeroed word, does not halt, and the
	// test reports a missing guard that is actually present.
	run := func(v Variant, w uint32) *CPU {
		bus := &stubBus{}
		binary.LittleEndian.PutUint32(bus.mem[0:], w)
		cpu := NewCPU(bus)
		cpu.Arch = v
		cpu.Mode, cpu.R[15] = ModeSYS, 0
		cpu.Step()
		return cpu
	}
	for _, c := range v5Only {
		// The control: V5TE must EXECUTE it (BKPT legitimately halts there too).
		if cpu := run(V5TE, c.w); cpu.Halted && c.name != "BKPT" {
			t.Fatalf("control failed: V5TE halted on 0x%08X (%s): %s", c.w, c.name, cpu.HaltReason)
		}
		if cpu := run(V4T, c.w); !cpu.Halted {
			t.Errorf("V4T executed 0x%08X (%s) instead of halting", c.w, c.name)
		}
	}
}

// TestV4TThumbBLX covers the two Thumb BLX forms, which are ARMv5 additions to
// an otherwise identical 16-bit set.
func TestV4TThumbBLX(t *testing.T) {
	cases := []struct {
		halfwords []uint16
		name      string
	}{
		{[]uint16{0x4780}, "BLX r0 (register)"},
		{[]uint16{0xF000, 0xE800}, "BLX <imm> (long branch pair)"},
	}
	for _, c := range cases {
		bus := &stubBus{}
		for i, h := range c.halfwords {
			binary.LittleEndian.PutUint16(bus.mem[i*2:], h)
		}
		cpu := NewCPU(bus)
		cpu.Arch = V4T
		cpu.Mode, cpu.Thumb, cpu.R[15] = ModeSYS, true, 0
		for range c.halfwords {
			cpu.Step()
		}
		if !cpu.Halted {
			t.Errorf("V4T executed Thumb %s instead of halting", c.name)
		}
	}
}

// TestV4TKeepsV4Instructions is the other half of the claim: the subset must
// still execute everything the ARM7TDMI DOES have. A variant that rejected too
// much would pass every test above.
func TestV4TKeepsV4Instructions(t *testing.T) {
	bus := &stubBus{}
	// MOV r0, #0x12 / ADD r0, r0, #1 / BX lr — plain ARMv4 code.
	for i, w := range []uint32{0xE3A00012, 0xE2800001, 0xE12FFF1E} {
		binary.LittleEndian.PutUint32(bus.mem[i*4:], w)
	}
	cpu := NewCPU(bus)
	cpu.Arch = V4T
	cpu.Mode, cpu.R[15], cpu.R[14] = ModeSYS, 0, 0x40
	for i := 0; i < 3; i++ {
		cpu.Step()
		if cpu.Halted {
			t.Fatalf("V4T halted on plain ARMv4 code at step %d: %s", i, cpu.HaltReason)
		}
	}
	if cpu.R[0] != 0x13 {
		t.Errorf("r0 = %#x, want 0x13", cpu.R[0])
	}
	if cpu.R[15] != 0x40 {
		t.Errorf("BX lr went to %#x, want 0x40", cpu.R[15])
	}
}

// TestVariantZeroValueIsV5TE pins the thing that would break the DS silently:
// a CPU built without setting Arch must still be the ARMv5TE core.
func TestVariantZeroValueIsV5TE(t *testing.T) {
	if (Variant(0)) != V5TE {
		t.Fatal("the zero Variant is no longer V5TE — every existing machine model just changed CPU")
	}
	cpu := NewCPU(&stubBus{})
	if cpu.Arch != V5TE || !cpu.Arch.v5OrLater() || cpu.Arch.isV6() {
		t.Errorf("default core is %v (v5=%v, v6=%v)", cpu.Arch, cpu.Arch.v5OrLater(), cpu.Arch.isV6())
	}
}
