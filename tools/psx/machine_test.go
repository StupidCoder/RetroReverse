package psx

import (
	"strings"
	"testing"
)

// textFromWords packs MIPS words little-endian into a text image.
func textFromWords(words []uint32) []byte {
	text := make([]byte, len(words)*4)
	for i, w := range words {
		text[i*4] = byte(w)
		text[i*4+1] = byte(w >> 8)
		text[i*4+2] = byte(w >> 16)
		text[i*4+3] = byte(w >> 24)
	}
	return text
}

// TestInterruptDispatch proves the M9 interrupt path end to end: with interrupts
// enabled and a handler registered (as HookEntryInt would), a raised IRQ vectors
// through the BIOS-HLE exception handler into the game handler, which runs, acks
// the IRQ and returns — and the interrupted context is restored transparently.
func TestInterruptDispatch(t *testing.T) {
	const base = 0x80010000
	// Main spin loop: wait for the sentinel at base+0x1010, then return to $ra=0.
	//   0x00 lui  $t0, 0x8001            ; $t0 = base
	//   0x04 lw   $t2, 0x1010($t0)       ; loop:
	//   0x08 nop                         ; load delay
	//   0x0C beq  $t2, $zero, loop
	//   0x10 nop
	//   0x14 jr   $ra
	//   0x18 nop
	main := []uint32{
		0x3C088001, // lui   $t0, 0x8001
		0x8D0A1010, // lw    $t2, 0x1010($t0)
		0x00000000, // nop
		0x1140FFFD, // beq   $t2, $zero, -3 (loop)
		0x00000000, // nop
		0x03E00008, // jr    $ra
		0x00000000, // nop
	}
	// Handler at base+0x100: write 0xDEAD sentinel, ack I_STAT, clobber $t0 (to
	// prove context restore), return.
	//   lui $t0,0x8001 ; ori $t1,$zero,0xDEAD ; sw $t1,0x1010($t0)
	//   lui $t3,0x1F80 ; sw $zero,0x1070($t3)  ; lui $t0,0xDEAD ; jr $ra ; nop
	handler := []uint32{
		0x3C088001, // lui   $t0, 0x8001
		0x3409DEAD, // ori   $t1, $zero, 0xDEAD
		0xAD091010, // sw    $t1, 0x1010($t0)
		0x3C0B1F80, // lui   $t3, 0x1F80
		0xAD601070, // sw    $zero, 0x1070($t3)   ; ack I_STAT (0x1F801070)
		0x3C08DEAD, // lui   $t0, 0xDEAD          ; clobber (must be restored)
		0x03E00008, // jr    $ra
		0x00000000, // nop
	}

	text := make([]byte, 0x200)
	copy(text, textFromWords(main))
	copy(text[0x100:], textFromWords(handler))
	e := &EXE{PC0: base, TAddr: base, TSize: uint32(len(text)), Text: text}

	m := NewMachine()
	m.LoadEXE(e)
	m.CPU.SetReg(31, 0) // $ra = 0 so the main loop's jr $ra exits the run

	// Register the handler chain {next, handler} at base+0x1000, as HookEntryInt
	// would, and point the machine at it.
	chain := uint32(base + 0x1000)
	m.write32(chain+0, 0)
	m.write32(chain+4, base+0x100)
	m.isrChain = chain

	// Enable interrupts (SR: IEc | IM2) and raise a masked CD IRQ (bit 2).
	m.CPU.COP0[12] = 0x0401
	m.irqMask = 1 << 2
	m.raiseIRQ(2)

	res := m.Run(100000)
	if got := m.read32(base + 0x1010); got != 0xDEAD {
		t.Fatalf("handler did not run: sentinel = 0x%X, want 0xDEAD (%s)", got, res)
	}
	if res.PC != 0 {
		t.Fatalf("run did not exit cleanly (context restore likely failed): %s", res)
	}
	if m.isr.active {
		t.Fatalf("ISR state left active after return")
	}
}

// TestBootReachesGameInit loads the real Ridge Racer EXE and runs it. Success is
// that boot proceeds deep into the game's initialization without the CPU halting
// on an unimplemented opcode — the milestone for the oracle.
func TestBootReachesGameInit(t *testing.T) {
	v := loadDisc(t) // from cd_test.go; skips when the image is absent
	data, err := v.ReadFile("SCUS-943.00;1")
	if err != nil {
		t.Fatalf("read boot exe: %v", err)
	}
	e, err := ParseEXE(data)
	if err != nil {
		t.Fatalf("parse exe: %v", err)
	}

	m := NewMachine()
	m.LoadEXE(e)
	res := m.Run(50_000_000)
	t.Logf("%s", res)
	t.Logf("BIOS calls: %v", m.BiosCalls())
	if tty := m.TTY(); tty != "" {
		t.Logf("TTY: %q", tty)
	}
	if len(m.Log) > 0 {
		t.Logf("notes: %s", strings.Join(m.Log, " | "))
	}

	if strings.HasPrefix(res.Reason, "halt: unimplemented") {
		t.Fatalf("boot halted on an unimplemented instruction: %s", res)
	}
	if m.CPU.Steps < 1000 {
		t.Fatalf("boot barely executed (%d steps): %s", m.CPU.Steps, res)
	}
}

// TestTinyProgram exercises LoadEXE + Run on a hand-built program that writes a
// sentinel to RAM and returns, independent of the disc image.
func TestTinyProgram(t *testing.T) {
	// lui $t0,0x8001 ; ori $t0,$t0,0x0000 ; addiu $t1,$zero,0x1234 ;
	// sw $t1,0($t0) ; jr $ra ; nop
	prog := []uint32{
		0x3C088001, // lui   $t0, 0x8001
		0x35080000, // ori   $t0, $t0, 0
		0x24091234, // addiu $t1, $zero, 0x1234
		0xAD090000, // sw    $t1, 0($t0)
		0x03E00008, // jr    $ra
		0x00000000, // nop
	}
	text := make([]byte, len(prog)*4)
	for i, w := range prog {
		text[i*4] = byte(w)
		text[i*4+1] = byte(w >> 8)
		text[i*4+2] = byte(w >> 16)
		text[i*4+3] = byte(w >> 24)
	}
	e := &EXE{PC0: 0x80010000, TAddr: 0x80010000, TSize: uint32(len(text)), Text: text}

	m := NewMachine()
	m.LoadEXE(e)
	m.CPU.SetReg(31, 0) // $ra = 0 so jr $ra exits the run
	res := m.Run(100)
	if got := uint32(m.ram[0x10000]) | uint32(m.ram[0x10001])<<8; got != 0x1234 {
		t.Errorf("mem[0x80010000] = 0x%X, want 0x1234 (%s)", got, res)
	}
}
