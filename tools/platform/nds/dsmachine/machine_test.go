package dsmachine

import (
	"testing"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/nds"
)

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

// TestSaveStateRoundTrip checks that a snapshot restores a machine that is
// bit-identical where it matters — including the ARM's BANKED registers, which are
// the part a savestate is most likely to drop and least likely to be caught dropping.
func TestSaveStateRoundTrip(t *testing.T) {
	rom := &nds.ROM{Data: make([]byte, 0x8000)}
	copy(rom.Data[0x0C:], "TEST")
	rom.Header.ARM9RAMAddr, rom.Header.ARM9Entry = 0x02000000, 0x02000000
	rom.Header.ARM7RAMAddr, rom.Header.ARM7Entry = 0x02380000, 0x02380000

	m := New(rom, 0x023C0000)
	// Dirty some state that a naive snapshot would miss.
	m.ARM9.cpu.R[5] = 0xCAFEF00D
	// Switch modes through the CPSR, not by poking Mode: only SetCPSR banks the
	// registers, and a test that assigns Mode directly is testing nothing.
	setMode(m.ARM9.cpu, arm.ModeIRQ)
	m.ARM9.cpu.R[13] = 0x0BADF00D // the IRQ-mode stack pointer
	setMode(m.ARM9.cpu, arm.ModeSVC)
	m.ARM9.cpu.R[13] = 0x02345678 // ...and the supervisor one
	m.ARM7.ie, m.ARM7.if_, m.ARM7.ime = 0x1234, 0x0002, true
	m.ARM9.timers[2] = timer{counter: 0x4321, reload: 0x1111, ctrl: 0x00C1}
	m.vram.setCNT(0, 0x83) // bank A -> texture memory
	m.vram.bank[0][0x1234] = 0xAB
	m.pal[0x400] = 0x99
	m.Poke(true, 0x02001000, []byte{1, 2, 3, 4})
	m.SetTouch(77, 88, true)
	m.spi.resultIdx = 1

	dir := t.TempDir() + "/s.state"
	if err := m.SaveState(dir); err != nil {
		t.Fatal(err)
	}

	// A fresh machine, then restore into it.
	n := New(rom, 0x023C0000)
	if err := n.LoadState(dir); err != nil {
		t.Fatal(err)
	}

	if got := n.ARM9.cpu.R[5]; got != 0xCAFEF00D {
		t.Errorf("ARM9 r5 = %08X, want CAFEF00D", got)
	}
	if got := n.ARM9.cpu.R[13]; got != 0x02345678 {
		t.Errorf("ARM9 svc sp = %08X, want 02345678", got)
	}
	// The IRQ bank is the one that matters: switch to it and see if it survived.
	setMode(n.ARM9.cpu, arm.ModeIRQ)
	if got := n.ARM9.cpu.R[13]; got != 0x0BADF00D {
		t.Errorf("ARM9 IRQ-mode sp = %08X, want 0BADF00D — the banked registers were not restored", got)
	}
	if n.ARM7.ie != 0x1234 || n.ARM7.if_ != 0x0002 || !n.ARM7.ime {
		t.Errorf("ARM7 interrupt controller not restored: IE=%X IF=%X IME=%v", n.ARM7.ie, n.ARM7.if_, n.ARM7.ime)
	}
	if n.ARM9.timers[2].counter != 0x4321 || n.ARM9.timers[2].ctrl != 0x00C1 {
		t.Errorf("ARM9 timer 2 not restored: %+v", n.ARM9.timers[2])
	}
	if n.vram.bank[0][0x1234] != 0xAB {
		t.Error("VRAM bank A not restored")
	}
	// The page tables are derived, not stored — the restore must rebuild them.
	if n.vram.cnt[0] != 0x83 || n.vram.read8(spTex, 0x1234) != 0xAB {
		t.Error("VRAM mapping not rebuilt from VRAMCNT after restore")
	}
	if n.pal[0x400] != 0x99 {
		t.Error("palette RAM not restored")
	}
	if got := n.Snapshot(true, 0x02001000, 4); got[0] != 1 || got[3] != 4 {
		t.Errorf("main RAM not restored: %v", got)
	}
	if !n.spi.touchDown || n.spi.touchX != 77 || n.spi.resultIdx != 1 {
		t.Error("the SPI bus / stylus state was not restored")
	}
}

// TestSaveStateRejectsWrongGame guards the one mistake that produces a machine that
// runs and is nonsense: restoring one game's state into another's memory map.
func TestSaveStateRejectsWrongGame(t *testing.T) {
	a := &nds.ROM{Data: make([]byte, 0x8000)}
	copy(a.Data[0x0C:], "AAAA")
	b := &nds.ROM{Data: make([]byte, 0x8000)}
	copy(b.Data[0x0C:], "BBBB")

	ma := New(a, 0)
	p := t.TempDir() + "/a.state"
	if err := ma.SaveState(p); err != nil {
		t.Fatal(err)
	}
	if err := New(b, 0).LoadState(p); err == nil {
		t.Fatal("loading game A's savestate into game B was allowed")
	}
}

// setMode switches a core's processor mode the way the hardware does — through the
// CPSR, so the banked registers are swapped.
func setMode(c *arm.CPU, mode uint32) { c.SetCPSR(c.CPSR()&^0x1F | mode) }
