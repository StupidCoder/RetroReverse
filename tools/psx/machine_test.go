package psx

import (
	"strings"
	"testing"
)

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
