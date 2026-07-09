package n64

import "testing"

// newRDPTest builds a machine with a 320x240 16-bit colour image and a full-screen
// scissor, which is what every command below draws into.
func newRDPTest(t *testing.T, cycle uint32) *Machine {
	t.Helper()
	m := &Machine{
		RDRAM:   make([]byte, rdramSize),
		DMEM:    make([]byte, spMemSize),
		IMEM:    make([]byte, spMemSize),
		PIF:     make([]byte, pifRAMSize),
		dp:      regFile{},
		logSeen: map[string]bool{},
	}
	m.CPU = newBareCPU(m)
	m.rdp.Color = image{Format: fmtRGBA, Size: size16, Width: 320, Addr: 0x100000}
	m.rdp.Scissor.XL, m.rdp.Scissor.YL = 320<<2, 240<<2
	m.rdp.OtherModes = uint64(cycle) << 52
	return m
}

func pixel(m *Machine, x, y uint32) uint16 {
	a := m.rdp.pixelAddr(x, y)
	return uint16(m.RDRAM[a])<<8 | uint16(m.RDRAM[a+1])
}

func TestFillRectangleWritesTwoPixelsPerWord(t *testing.T) {
	// In FILL mode the fill colour is written 32 bits at a time, so for a 16-bit
	// framebuffer it carries two distinct pixels: the high half lands on even
	// columns and the low half on odd ones. A model that writes one of them
	// everywhere produces a screen that is subtly the wrong colour.
	m := newRDPTest(t, cycleFill)
	m.rdp.FillColor = 0xAAAA5555

	w := uint64(cmdFillRect)<<56 | uint64(9<<2)<<44 | uint64(4<<2)<<32 | uint64(2<<2)<<12 | uint64(1<<2)
	m.fillRect(w)
	if m.CPU.Halted {
		t.Fatalf("halted: %s", m.CPU.HaltReason)
	}

	if got := pixel(m, 2, 1); got != 0xAAAA {
		t.Errorf("even column: got %04X want AAAA", got)
	}
	if got := pixel(m, 3, 1); got != 0x5555 {
		t.Errorf("odd column: got %04X want 5555", got)
	}
	// The encoded corner is inclusive, so column 9 and row 4 are drawn...
	if got := pixel(m, 9, 4); got != 0x5555 {
		t.Errorf("inclusive corner (9,4): got %04X want 5555", got)
	}
	// ...and nothing outside it is.
	if got := pixel(m, 10, 4); got != 0 {
		t.Errorf("pixel past the right edge was drawn: %04X", got)
	}
	if got := pixel(m, 2, 0); got != 0 {
		t.Errorf("pixel above the top edge was drawn: %04X", got)
	}
}

func TestFillRectangleClipsToScissor(t *testing.T) {
	m := newRDPTest(t, cycleFill)
	m.rdp.FillColor = 0xFFFFFFFF
	m.rdp.Scissor.XH, m.rdp.Scissor.YH = 4<<2, 4<<2
	m.rdp.Scissor.XL, m.rdp.Scissor.YL = 8<<2, 8<<2

	// A rectangle covering the whole screen.
	w := uint64(cmdFillRect)<<56 | uint64(319<<2)<<44 | uint64(239<<2)<<32
	m.fillRect(w)

	if got := pixel(m, 3, 5); got != 0 {
		t.Errorf("pixel left of the scissor was drawn: %04X", got)
	}
	if got := pixel(m, 5, 5); got == 0 {
		t.Error("pixel inside the scissor was not drawn")
	}
	if got := pixel(m, 8, 5); got != 0 {
		t.Errorf("pixel right of the scissor was drawn: %04X", got)
	}
}

func TestFillRectangleRefusesAnUnmodelledCycle(t *testing.T) {
	// Drawing something plausible in a mode we have not built is worse than
	// stopping: the image would look almost right.
	m := newRDPTest(t, cycle1)
	m.fillRect(uint64(cmdFillRect) << 56)
	if !m.CPU.Halted {
		t.Fatal("Fill_Rectangle in 1-cycle mode did not halt")
	}
}

func TestCommandLengths(t *testing.T) {
	// The triangle family grows with what it interpolates: four words of edge
	// coefficients, plus eight for texture, eight for shade, two for depth. A
	// wrong length desynchronises the whole command stream.
	for _, tc := range []struct {
		op   uint32
		want int
	}{
		{cmdTriFill, 4}, {cmdTriFillZ, 6},
		{cmdTriTex, 12}, {cmdTriTexZ, 14},
		{cmdTriShade, 12}, {cmdTriShadeZ, 14},
		{cmdTriShadeTex, 20}, {cmdTriShadeTexZ, 22},
		{cmdTexRect, 2}, {cmdTexRectFlip, 2},
		{cmdSetColorImage, 1}, {cmdSyncFull, 1}, {cmdFillRect, 1},
	} {
		if got := cmdLen(tc.op); got != tc.want {
			t.Errorf("%s: length %d want %d", cmdNameOf(tc.op), got, tc.want)
		}
	}
}

// The queue resumes at DPC_CURRENT, not DPC_START. Microcode appends commands
// and walks DPC_END forward a word at a time; restarting at START would
// re-execute the whole frame on every append.
func TestQueueResumesAtCurrentNotStart(t *testing.T) {
	m := newRDPTest(t, cycleFill)
	const base = 0x2000

	// Two Set_Fill_Color commands, appended one at a time.
	put := func(off uint32, w uint64) {
		for i := uint32(0); i < 8; i++ {
			m.RDRAM[base+off+i] = byte(w >> (56 - 8*i))
		}
	}
	put(0, uint64(cmdSetFillColor)<<56|0x1111)
	put(8, uint64(cmdSetFillColor)<<56|0x2222)

	seen := 0
	m.OnRDPCmd = func(*Machine, uint32, []uint64) { seen++ }

	m.dpWrite(dpRegBase+dpStart, base)
	m.dpWrite(dpRegBase+dpEnd, base+8) // the first command only
	if seen != 1 || m.rdp.FillColor != 0x1111 {
		t.Fatalf("after the first append: %d commands, fill colour %04X", seen, m.rdp.FillColor)
	}
	m.dpWrite(dpRegBase+dpEnd, base+16) // append the second
	if seen != 2 {
		t.Errorf("after the second append: %d commands ran, want 2 — the queue re-ran from START", seen)
	}
	if m.rdp.FillColor != 0x2222 {
		t.Errorf("fill colour %04X want 2222", m.rdp.FillColor)
	}
}

// A command whose last word has not been appended yet must not run, and must be
// picked up once DPC_END moves past it.
func TestPartiallyAppendedCommandWaits(t *testing.T) {
	m := newRDPTest(t, cycleFill)
	const base = 0x3000
	// A 2-word Texture_Rectangle, appended one word at a time.
	for i := uint32(0); i < 8; i++ {
		m.RDRAM[base+i] = byte(uint64(cmdTexRect) << 56 >> (56 - 8*i))
	}
	seen := 0
	m.OnRDPCmd = func(*Machine, uint32, []uint64) { seen++ }

	m.dpWrite(dpRegBase+dpStart, base)
	m.dpWrite(dpRegBase+dpEnd, base+8) // only the first of its two words
	if seen != 0 {
		t.Fatal("a half-appended command was executed")
	}
	if m.dp[dpCurrent] != base {
		t.Errorf("DPC_CURRENT = %08X, want it left before the incomplete command (%08X)",
			m.dp[dpCurrent], base)
	}
}

func TestSyncFullRaisesTheDPInterrupt(t *testing.T) {
	m := newRDPTest(t, cycleFill)
	m.execRDP(cmdSyncFull, []uint64{uint64(cmdSyncFull) << 56})
	if m.mi.Intr&intrDP == 0 {
		t.Error("Sync_Full did not raise the DP interrupt: the game never learns the frame is done")
	}
}

func TestUnmodelledCommandHalts(t *testing.T) {
	m := newRDPTest(t, cycleFill)
	m.execRDP(0x1F, []uint64{uint64(0x1F) << 56})
	if !m.CPU.Halted {
		t.Fatal("an unmodelled RDP command did not halt the machine")
	}
}
