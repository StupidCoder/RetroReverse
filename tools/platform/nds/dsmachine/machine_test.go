package dsmachine

import "testing"

// testMachine builds the shared state and two bare cores without a ROM, to
// exercise the game-neutral IPC cross-wiring in isolation.
func testMachine() *Machine {
	m := &Machine{ram: make([]byte, mainSize), swram: make([]byte, swramSize), logSeen: map[string]bool{}}
	m.ARM9 = &core{m: m, name: "ARM9", arm9: true, io: map[uint32]uint32{}}
	m.ARM7 = &core{m: m, name: "ARM7", arm9: false, io: map[uint32]uint32{}}
	return m
}

func writeHalf(c *core, addr uint32, v uint16) {
	c.ioWrite(addr, byte(v))
	c.ioWrite(addr+1, byte(v>>8))
}

// TestIPCSyncCrossWire verifies that each core's outgoing IPCSYNC nibble (bits
// 8-11) appears as the other core's incoming nibble (bits 0-3) — the mechanism
// the ARM9↔ARM7 rendezvous handshake polls.
func TestIPCSyncCrossWire(t *testing.T) {
	m := testMachine()
	writeHalf(m.ARM9, regIPCSYNC, 0x0500) // ARM9 posts nibble 5
	if got := m.ARM7.ioRead(regIPCSYNC) & 0xF; got != 5 {
		t.Fatalf("ARM7 incoming = %d, want 5", got)
	}
	writeHalf(m.ARM7, regIPCSYNC, 0x0A00) // ARM7 posts nibble 0xA
	if got := m.ARM9.ioRead(regIPCSYNC) & 0xF; got != 0xA {
		t.Fatalf("ARM9 incoming = %d, want 0xA", got)
	}
	// a core reads its own posted value back in bits 8-11
	if got := (m.ARM9.ioRead(regIPCSYNC) >> 8) & 0xF; got != 5 {
		t.Fatalf("ARM9 outgoing = %d, want 5", got)
	}
}

// TestIPCFifoCrossWire verifies the two directional FIFOs: a word sent by one
// core is received by the other (and only the other), FIFO-empty status tracks
// the queue, and a recv IRQ is posted to the receiver when it enabled it.
func TestIPCFifoCrossWire(t *testing.T) {
	m := testMachine()
	// enable both FIFOs; ARM7 enables its recv IRQ (IPCFIFOCNT bit 10)
	m.ARM9.io[regIPCFIFOCNT] = 0x8000
	m.ARM7.io[regIPCFIFOCNT] = 0x8000 | 0x0400
	m.ARM7.ie, m.ARM7.ime = irqIPCRecv, true

	if m.ARM9.fifoCnt()&0x0001 == 0 {
		t.Fatal("ARM9 send FIFO should start empty")
	}
	m.ARM9.ioWrite(regIPCFIFOSND, 0xEF) // low byte; assemble a full word
	m.ARM9.ioWrite(regIPCFIFOSND+1, 0xBE)
	m.ARM9.ioWrite(regIPCFIFOSND+2, 0xAD)
	m.ARM9.ioWrite(regIPCFIFOSND+3, 0xDE) // 0xDEADBEEF

	if to7, _ := m.FifoLens(); to7 != 1 {
		t.Fatalf("ARM9→ARM7 FIFO depth = %d, want 1", to7)
	}
	if m.ARM7.if_&irqIPCRecv == 0 {
		t.Fatal("ARM7 should have a pending recv-FIFO interrupt")
	}
	if m.ARM9.ioRead(regIPCFIFORCV) != 0 { // ARM9's own recv FIFO is empty
		t.Fatal("ARM9 recv FIFO should be empty (its send is the ARM7's recv)")
	}
	if got := m.ARM7.ioRead(regIPCFIFORCV); got != 0xDEADBEEF {
		t.Fatalf("ARM7 received 0x%08X, want 0xDEADBEEF", got)
	}
	if to7, _ := m.FifoLens(); to7 != 0 {
		t.Fatalf("FIFO should be drained, depth = %d", to7)
	}
}
