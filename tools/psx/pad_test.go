package psx

import "testing"

// TestPadPacketFormat checks writePad lays out the digital-controller packet the
// InitPad-based reader expects: present flag, pad id, then the active-low buttons.
func TestPadPacketFormat(t *testing.T) {
	m := NewMachine()
	m.padBuf = 0x80001000
	m.padActive = true
	m.PadButtons = PadReleased &^ PadStart // START held
	m.writePad()
	got := [4]byte{m.Read(m.padBuf), m.Read(m.padBuf + 1), m.Read(m.padBuf + 2), m.Read(m.padBuf + 3)}
	want := [4]byte{0x00, 0x41, byte((PadReleased &^ PadStart) & 0xFF), 0xFF}
	if got != want {
		t.Fatalf("pad packet = % X, want % X", got, want)
	}
	// A disarmed pad must not touch the buffer.
	m2 := NewMachine()
	m2.padBuf = 0x80001000
	m2.writePad() // padActive false
	if m2.Read(m2.padBuf) != 0 {
		t.Fatalf("disarmed pad wrote to buffer")
	}
}

// TestPadStartAdvancesTitle drives the real game: it boots to the attract title
// screen, injects a START press via the scripted-input path, and checks the
// framebuffer changes — the game left the title for the GAME START menu. Skips
// when the disc image is absent (like the other disc-backed tests).
func TestPadStartAdvancesTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("boots the full game")
	}
	v := loadDisc(t)
	data, err := v.ReadFile("SCUS-943.00;1")
	if err != nil {
		t.Fatal(err)
	}
	e, err := ParseEXE(data)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine()
	m.SetDisc(v)
	m.ISRHandler = 0x8004DF48
	m.LoadEXE(e)
	m.PadScript = []PadEvent{
		{AtStep: 380_000_000, Buttons: PadReleased &^ PadStart}, // press START
		{AtStep: 383_000_000, Buttons: PadReleased},             // release
	}
	m.Run(380_000_000)
	title := frameHash(m)
	m.Run(10_000_000) // let START register and the menu draw
	if menu := frameHash(m); menu == title {
		t.Fatalf("framebuffer unchanged after START: still on the title screen")
	}
	if !m.padActive || m.padBuf == 0 {
		t.Fatalf("game never armed the pad (InitPad/StartPad): active=%v buf=%08X", m.padActive, m.padBuf)
	}
}

func frameHash(m *Machine) uint64 {
	var h uint64 = 1469598103934665603
	for y := 0; y < m.gpu.dispH; y++ {
		for x := 0; x < m.gpu.dispW; x++ {
			px := m.gpu.vram[((m.gpu.dispY+y)&(vramH-1))*vramW+((m.gpu.dispX+x)&(vramW-1))]
			h = (h ^ uint64(px)) * 1099511628211
		}
	}
	return h
}
