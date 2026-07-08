package arm60

import "testing"

// TestUnsignedDivideRoutine runs the game's own unrolled restoring-division
// routine (LaunchMe 0x4C4: RSBS/SBCCS/ADC per bit, returning quotient in r0) for
// a dividend with bit 31 set — the case the printf/itoa hit when formatting VDL
// words. It exercises RSBS-then-conditional-SBC carry chaining across all 32
// shift amounts, which a small dividend never reaches.
func TestUnsignedDivideRoutine(t *testing.T) {
	prog := []uint32{
		0xE0702FA1, 0x20C11F80, 0xE0A33003, 0xE0702F21, 0x20C11F00, 0xE0A33003,
		0xE0702EA1, 0x20C11E80, 0xE0A33003, 0xE0702E21, 0x20C11E00, 0xE0A33003,
		0xE0702DA1, 0x20C11D80, 0xE0A33003, 0xE0702D21, 0x20C11D00, 0xE0A33003,
		0xE0702CA1, 0x20C11C80, 0xE0A33003, 0xE0702C21, 0x20C11C00, 0xE0A33003,
		0xE0702BA1, 0x20C11B80, 0xE0A33003, 0xE0702B21, 0x20C11B00, 0xE0A33003,
		0xE0702AA1, 0x20C11A80, 0xE0A33003, 0xE0702A21, 0x20C11A00, 0xE0A33003,
		0xE07029A1, 0x20C11980, 0xE0A33003, 0xE0702921, 0x20C11900, 0xE0A33003,
		0xE07028A1, 0x20C11880, 0xE0A33003, 0xE0702821, 0x20C11800, 0xE0A33003,
		0xE07027A1, 0x20C11780, 0xE0A33003, 0xE0702721, 0x20C11700, 0xE0A33003,
		0xE07026A1, 0x20C11680, 0xE0A33003, 0xE0702621, 0x20C11600, 0xE0A33003,
		0xE07025A1, 0x20C11580, 0xE0A33003, 0xE0702521, 0x20C11500, 0xE0A33003,
		0xE07024A1, 0x20C11480, 0xE0A33003, 0xE0702421, 0x20C11400, 0xE0A33003,
		0xE07023A1, 0x20C11380, 0xE0A33003, 0xE0702321, 0x20C11300, 0xE0A33003,
		0xE07022A1, 0x20C11280, 0xE0A33003, 0xE0702221, 0x20C11200, 0xE0A33003,
		0xE07021A1, 0x20C11180, 0xE0A33003, 0xE0702121, 0x20C11100, 0xE0A33003,
		0xE07020A1, 0x20C11080, 0xE0A33003, 0xE0702001, 0x20C11000, 0xE0A30003,
		// final MOV pc,lr omitted; we run exactly the 96 arithmetic steps.
	}
	cases := []struct{ dividend, divisor, want uint32 }{
		{0xE1FFFFFF, 16, 0xE1FFFFFF / 16},
		{0xE1FFFFFF, 10, 0xE1FFFFFF / 10},
		{0x80000000, 10, 0x80000000 / 10},
		{0xFFFFFFFF, 16, 0x0FFFFFFF},
		{0x00001234, 10, 0x1234 / 10},
	}
	for _, tc := range cases {
		bus := &testBus{mem: make([]byte, 0x4000)}
		for i, w := range prog {
			bus.put32(0x1000+uint32(i)*4, w)
		}
		c := NewCPU(bus)
		c.Mode = ModeSYS
		c.SetPC(0x1000)
		c.R[0] = tc.divisor
		c.R[1] = tc.dividend
		c.R[3] = 0xF0FFFFFF // the game enters with r3 uninitialised; 32 doublings flush it
		for i := 0; i < len(prog); i++ {
			c.Step()
		}
		if c.Halted {
			t.Fatalf("halted: %s", c.HaltReason)
		}
		if c.R[0] != tc.want {
			t.Errorf("%#08x / %d = %#08x (r1 rem %#08x), want %#08x",
				tc.dividend, tc.divisor, c.R[0], c.R[1], tc.want)
		}
	}
}
