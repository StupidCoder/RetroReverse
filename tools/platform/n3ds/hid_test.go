package n3ds

import (
	"os"
	"testing"
)

// hidTestMachine is a machine with nothing but the HID shared page mapped — enough to ask
// what updateHIDShared publishes, and nothing else.
func hidTestMachine(t *testing.T) *Machine {
	t.Helper()
	m := &Machine{}
	const base = 0x1FF00000
	m.mapRegion("hid-shared", base, make([]byte, hidSharedSize))
	m.hidSharedAddr = base
	return m
}

// TestHIDSectionSpans is the reason we can write a touch sample without a second trace.
//
// The section BASES were read off the game's own polling loop. The ring shapes then have
// to fit exactly between them — and they do, which is what pins TouchData's shorter header
// and its 8-byte entry. Reusing PadData's 0x10 stride would run the eighth touch sample
// past the end of the section and into the accelerometer's ring.
func TestHIDSectionSpans(t *testing.T) {
	if got := hidEntries + hidRingLen*hidEntSz; got != hidTouchBase-hidPadBase {
		t.Errorf("PadData spans 0x%X, but TouchData starts 0x%X after it", got, hidTouchBase-hidPadBase)
	}
	if got := hidTouchEntries + hidRingLen*hidTouchEntSz; got != hidAccelBase-hidTouchBase {
		t.Errorf("TouchData spans 0x%X, but Accelerometer starts 0x%X after it", got, hidAccelBase-hidTouchBase)
	}
}

// TestHIDTouchSample checks what the game would read: the pen where it was put, and the
// valid flag telling it apart from a screen nobody is touching.
func TestHIDTouchSample(t *testing.T) {
	m := hidTestMachine(t)

	entry := func(idx uint32) (x, y uint16, valid uint32) {
		a := m.hidSharedAddr + hidTouchBase + hidTouchEntries + idx*hidTouchEntSz
		pos := m.ReadWord(a)
		return uint16(pos), uint16(pos >> 16), m.ReadWord(a + 4)
	}

	// Nothing touched: a published sample that says so.
	m.updateHIDShared()
	if x, y, v := entry(0); v != 0 || x != 0 || y != 0 {
		t.Errorf("idle sample = (%d,%d) valid=%d, want (0,0) valid=0", x, y, v)
	}

	m.SetTouch(160, 120, true)
	m.updateHIDShared()
	if x, y, v := entry(1); x != 160 || y != 120 || v != 1 {
		t.Errorf("pen down = (%d,%d) valid=%d, want (160,120) valid=1", x, y, v)
	}
	// The header must name the entry just written, or the game reads a stale slot.
	if got := m.ReadWord(m.hidSharedAddr + hidTouchBase + hidHdrIndex); got != 1 {
		t.Errorf("TouchData latest-entry index = %d, want 1", got)
	}

	// Lifting reports the origin, not the last place the pen was: a game that trusts the
	// position without testing the flag must not keep being touched where it was touched.
	m.SetTouch(160, 120, false)
	m.updateHIDShared()
	if x, y, v := entry(2); v != 0 || x != 0 || y != 0 {
		t.Errorf("pen up = (%d,%d) valid=%d, want (0,0) valid=0", x, y, v)
	}
}

// TestHIDTouchClamp keeps a stray coordinate on the screen. The debugger clamps a drag to
// the panel, but a caller need not, and a y of 240 would be published as a touch a real
// digitiser cannot report.
func TestHIDTouchClamp(t *testing.T) {
	m := hidTestMachine(t)
	m.SetTouch(-5, 999, true)
	if m.hidTouchX != 0 || m.hidTouchY != 239 {
		t.Errorf("clamped to (%d,%d), want (0,239)", m.hidTouchX, m.hidTouchY)
	}
}

// TestHIDTouchRideSnapshot: the frame debugger replays a captured frame on a scratch
// machine restored from that frame's snapshot. If the pen does not ride the snapshot, the
// replay is a frame nobody was touching — a different frame from the one captured.
func TestHIDTouchRideSnapshot(t *testing.T) {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		t.Skip("Super Mario 3D Land image not present (game images are not committed)")
	}
	m, err := NewMachine(img)
	if err != nil {
		t.Fatal(err)
	}
	m.SetTouch(64, 96, true)
	if err := m.SetKeys("a,start"); err != nil {
		t.Fatal(err)
	}

	other, err := NewMachine(img)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.RestoreState(m.SnapshotState()); err != nil {
		t.Fatal(err)
	}
	if other.hidTouchX != 64 || other.hidTouchY != 96 || !other.hidTouchDown {
		t.Errorf("restored stylus = (%d,%d) down=%v, want (64,96) down=true",
			other.hidTouchX, other.hidTouchY, other.hidTouchDown)
	}
	if other.hidButtons != m.hidButtons {
		t.Errorf("restored buttons = %#x, want %#x", other.hidButtons, m.hidButtons)
	}
}
