package n64

import "testing"

func newBareMachine() *Machine {
	m := &Machine{
		RDRAM:   make([]byte, rdramSize),
		DMEM:    make([]byte, spMemSize),
		IMEM:    make([]byte, spMemSize),
		PIF:     make([]byte, pifRAMSize),
		ROM:     make([]byte, 1024),
		sp:      regFile{},
		dp:      regFile{},
		logSeen: map[string]bool{},
	}
	m.CPU = newBareCPU(m)
	return m
}

// A test ROM prints its results here rather than onto the screen, which is what
// makes a conformance suite usable from a program.
func TestISViewerPrintsWhatWasBuffered(t *testing.T) {
	m := newBareMachine()
	var got []string
	m.OnPrint = func(_ *Machine, line string) { got = append(got, line) }

	msg := "Done! Tests: 262. Failed: 0\n"
	for i, c := range []byte(msg) {
		m.Write(isvBuffer+uint32(i), c) // the ROM stores a byte at a time
	}
	m.Write32(isvLen, uint32(len(msg)))

	if len(got) != 1 || got[0] != "Done! Tests: 262. Failed: 0" {
		t.Fatalf("printed %q, want the message without its newline", got)
	}
	if lines := m.ISViewerLines(); len(lines) != 1 {
		t.Errorf("ISViewerLines returned %d lines, want 1", len(lines))
	}
}

// The IS-Viewer window lies inside the cartridge's address range and must be
// decoded ahead of it, or its writes are swallowed as writes to a read-only ROM.
func TestISViewerIsDecodedBeforeTheCartridge(t *testing.T) {
	m := newBareMachine()
	m.Write(isvBuffer, 'X')
	if m.isv.Buf[0] != 'X' {
		t.Error("a byte written to the IS-Viewer buffer went to the cartridge instead")
	}
	for _, l := range m.Log {
		t.Errorf("unexpected machine note: %s", l)
	}
}

// RDRAM does not mirror. Past the populated chips a read returns zero and a
// write vanishes — which is how n64-systemtest's boot code clears IMEM, by
// DMAing from an address above 8 MiB.
func TestRDRAMDoesNotMirror(t *testing.T) {
	m := newBareMachine()
	m.RDRAM[0x10] = 0xAB

	const beyond = rdramSize + 0x10
	if got := m.rdramRead(beyond); got != 0 {
		t.Errorf("read past RDRAM returned %02X, want 0 (it mirrored the low half)", got)
	}
	m.rdramWrite(beyond, 0xCD)
	if m.RDRAM[0x10] != 0xAB {
		t.Error("a write past RDRAM landed in the low half: RDRAM is mirroring")
	}
}

func TestSPDMAFromUnpopulatedRDRAMClearsIMEM(t *testing.T) {
	// The exact sequence n64-systemtest's bootcode relies on.
	m := newBareMachine()
	for i := range m.IMEM {
		m.IMEM[i] = 0xFF
	}
	m.sp[spMemAddr] = spMemSize // select IMEM
	m.sp[spDramAddr] = rdramSize + 0x1000
	m.spDMA(0x7FF, false) // RDRAM -> SP memory, 2 KiB

	for i := 0; i < 0x800; i++ {
		if m.IMEM[i] != 0 {
			t.Fatalf("IMEM[%d] = %02X: the DMA read mirrored RDRAM instead of zeroes", i, m.IMEM[i])
		}
	}
}

func TestSPDMALongerThanIMEMWrapsInsideIt(t *testing.T) {
	// A transfer longer than 4 KiB wraps within the SP memory it started in; it
	// does not run on into the neighbouring one.
	m := newBareMachine()
	for i := range m.IMEM {
		m.IMEM[i] = byte(i)
	}
	for i := range m.DMEM {
		m.DMEM[i] = 0xEE
	}
	m.sp[spMemAddr] = spMemSize // IMEM
	m.sp[spDramAddr] = 0
	m.spDMA(0xFFF, true) // SP memory -> RDRAM, the full 4 KiB

	// The DMEM must be untouched by an IMEM transfer.
	for i := range m.DMEM {
		if m.DMEM[i] != 0xEE {
			t.Fatalf("DMEM[%d] was disturbed by an IMEM transfer", i)
		}
	}
	if m.RDRAM[0] != 0 || m.RDRAM[0xFFF] != 0xFF {
		t.Errorf("IMEM was not copied verbatim: RDRAM[0]=%02X RDRAM[0xFFF]=%02X",
			m.RDRAM[0], m.RDRAM[0xFFF])
	}
}
