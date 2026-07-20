package xboxadapter

// pad_test.go pins the keyboard-to-pad mapping, and mostly it pins the stick.
//
// The arrows used to be the d-pad. They are the left stick now, because that is the Xbox's
// primary directional control — and the interesting half of that change is not "the arrows
// moved", it is that the pad's stick counts UP as POSITIVE while every screen coordinate in
// this debugger counts up as negative. The GameCube's first cut of this exact change shipped
// its stick upside down (gcadapter/pad_test.go says so, and this file exists because that is
// a mistake worth only making once).
//
// None of it needs the disc. Key() touches nothing but the adapter's own two pad fields and
// the platform's vocabulary, so these run wherever the code does — which matters, because the
// game image is not committed and every test that needs it skips.

import (
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/xbox"
)

func padAdapter() *Adapter { return &Adapter{held: map[string]bool{}} }

// press drives a key down (or up) and returns the pad state the queue now ends on — what the
// next frame's poll will report.
func press(t *testing.T, a *Adapter, name string, down bool) xbox.PadState {
	t.Helper()
	if err := a.Key(debug.Key{Name: name, Down: down}); err != nil {
		t.Fatal(err)
	}
	if len(a.padQueue) == 0 {
		t.Fatalf("%s down=%v queued nothing", name, down)
	}
	return a.padQueue[len(a.padQueue)-1]
}

func mustControl(t *testing.T, name string) xbox.PadControl {
	t.Helper()
	c, ok := xbox.PadControlByName(name)
	if !ok {
		t.Fatalf("no pad control %q", name)
	}
	return c
}

// TestArrowsDriveTheStickNotTheDPad is the change itself, stated as a test. Both halves
// matter: an arrow must move the stick, AND it must leave the d-pad bits alone.
func TestArrowsDriveTheStickNotTheDPad(t *testing.T) {
	dpad := mustControl(t, "up").Bit | mustControl(t, "down").Bit |
		mustControl(t, "left").Bit | mustControl(t, "right").Bit

	for _, c := range []struct {
		key          string
		wantX, wantY int16
	}{
		// The DIRECTIONS are the title's own, and they are not a convention this file chose:
		// each was derived by driving one axis at the game's on-screen keyboard and watching
		// which way the cursor stepped, with the matching d-pad bit producing the identical
		// frame. Axis 1 POSITIVE stepped the cursor UP. There is no inversion to do, and
		// these four lines are what say so.
		{"ArrowUp", 0, +xbox.PadStickFull},
		{"ArrowDown", 0, -xbox.PadStickFull},
		{"ArrowLeft", -xbox.PadStickFull, 0},
		{"ArrowRight", +xbox.PadStickFull, 0},
	} {
		a := padAdapter()
		st := press(t, a, c.key, true)
		if st.Axes[0] != c.wantX || st.Axes[1] != c.wantY {
			t.Errorf("%s: stick = (%d,%d), want (%d,%d) — the pad counts up and right as "+
				"POSITIVE (derived at the title's on-screen keyboard)",
				c.key, st.Axes[0], st.Axes[1], c.wantX, c.wantY)
		}
		if st.Buttons&dpad != 0 {
			t.Errorf("%s: set d-pad bits %04X — an arrow is the STICK, and a menu that "+
				"reads only the d-pad must still be reachable from the keypad alone",
				c.key, st.Buttons&dpad)
		}
		if st.Analog != ([8]byte{}) {
			t.Errorf("%s: pressed an analog button (%v) — an arrow is a direction", c.key, st.Analog)
		}
	}
}

// TestKeypadIsTheDPad is the other half: the d-pad kept a home of its own, so a screen that
// reads only the d-pad and a game that reads only the stick are both reachable.
func TestKeypadIsTheDPad(t *testing.T) {
	for _, c := range []struct {
		key  string
		want string
	}{{"8", "up"}, {"2", "down"}, {"4", "left"}, {"6", "right"}} {
		a := padAdapter()
		st := press(t, a, c.key, true)
		want := mustControl(t, c.want).Bit
		if st.Buttons != want {
			t.Errorf("key %q: buttons = %04X, want %04X (%s)", c.key, st.Buttons, want, c.want)
		}
		if st.Axes != ([4]int16{}) {
			t.Errorf("key %q: moved the stick to %v — the keypad is the DIGITAL pad", c.key, st.Axes)
		}
	}
}

// TestAnalogButtonKeys covers the two buttons the title named — and the fact that a pressure
// byte is not a bit, which is the whole reason the vocabulary had to change.
func TestAnalogButtonKeys(t *testing.T) {
	for _, name := range []string{"a", "b"} {
		a := padAdapter()
		st := press(t, a, name, true)
		c := mustControl(t, name)
		if st.Analog[c.Index] != xbox.PadPressed {
			t.Errorf("%q: analog[%d] = %02X, want %02X — the title thresholds this byte at "+
				"0x1E (0x147E5) and XAPI zeroes it under 0x20 (0x24390A)",
				name, c.Index, st.Analog[c.Index], xbox.PadPressed)
		}
		if st.Buttons != 0 {
			t.Errorf("%q: set wButtons %04X — A and B are PRESSURE BYTES at gamepad+2 and "+
				"gamepad+3, not digital bits", name, st.Buttons)
		}
		// The other seven bytes must be at rest: a name that leaked into a neighbouring
		// offset would press a button nobody asked for, and six of those offsets have no
		// name yet precisely because nothing has proven what they are.
		for i, p := range st.Analog {
			if i != c.Index && p != 0 {
				t.Errorf("%q: also pressed analog[%d] = %02X", name, i, p)
			}
		}
	}
}

// TestTriggerKeys pins the R and L keyboard keys to the analog triggers — OutRun's accelerate
// and brake. The R trigger is pressure byte gamepad+9 (control "an7"), the L trigger gamepad+8
// ("an6"); the R key holding an7 is the input Part XVIII drove the race with.
func TestTriggerKeys(t *testing.T) {
	for _, tc := range []struct{ key, control string }{{"r", "an7"}, {"l", "an6"}} {
		a := padAdapter()
		st := press(t, a, tc.key, true)
		c := mustControl(t, tc.control)
		if st.Analog[c.Index] != xbox.PadPressed {
			t.Errorf("%q: analog[%d] = %02X, want %02X — the %s key should press the trigger byte",
				tc.key, c.Index, st.Analog[c.Index], xbox.PadPressed, tc.key)
		}
		for i, p := range st.Analog {
			if i != c.Index && p != 0 {
				t.Errorf("%q: also pressed analog[%d] = %02X", tc.key, i, p)
			}
		}
		if st.Buttons != 0 {
			t.Errorf("%q: set wButtons %04X — a trigger is a pressure byte, not a bit", tc.key, st.Buttons)
		}
	}
}

// TestOppositeArrowsCancel is what a physical stick does, and what a keyboard will be asked
// to do the moment a thumb rolls across two arrow keys. It also pins the accumulate-then-
// clamp shape: an implementation that ASSIGNED each direction would answer differently
// depending on Go's map iteration order — the same input giving a different stick each run,
// and only sometimes, which is the worst way for this to be wrong.
func TestOppositeArrowsCancel(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowLeft", true)
	st := press(t, a, "ArrowRight", true)
	if st.Axes != ([4]int16{}) {
		t.Errorf("left+right = %v, want centred: opposite directions cancel", st.Axes)
	}
	// And releasing one must leave the other still driving — the cancel is a level, not a
	// latch.
	st = press(t, a, "ArrowLeft", false)
	if st.Axes[0] != +xbox.PadStickFull {
		t.Errorf("after releasing left, stick X = %d, want %d (right is still held)",
			st.Axes[0], xbox.PadStickFull)
	}
}

// TestDiagonalIsASquareGate pins the declared model choice, and the derived fact behind it.
//
// The GameCube splits a diagonal at full/sqrt2 because its shell has an octagonal gate. That
// number does NOT transfer: nothing in this image describes the Xbox's shell, and the title's
// own threshold refutes the shape anyway — a fresh direction must clear 0x5FFF (0x1469D), and
// 0x7FFF/sqrt2 is 0x5A82, which is under it. Split that way, a diagonal registers NEITHER
// direction and does it silently. So each axis goes to full, and this test is what makes that
// a decision rather than an accident.
func TestDiagonalIsASquareGate(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowUp", true)
	st := press(t, a, "ArrowLeft", true)
	if st.Axes[0] != -xbox.PadStickFull || st.Axes[1] != +xbox.PadStickFull {
		t.Fatalf("up+left = (%d,%d), want (%d,%d)",
			st.Axes[0], st.Axes[1], -xbox.PadStickFull, xbox.PadStickFull)
	}
	// The point of the choice, stated as the thing it buys: both axes clear the title's own
	// fresh-trigger threshold, so a diagonal is two live directions rather than none.
	const freshTrigger = 0x5FFF // 0x1469D CMP ECX,$00005FFF / SETLE
	for i, v := range [2]int16{st.Axes[0], st.Axes[1]} {
		mag := v
		if mag < 0 {
			mag = -mag
		}
		if int(mag) <= freshTrigger {
			t.Errorf("diagonal axis %d = %d, which does not clear the title's fresh-direction "+
				"threshold %#x — the direction would be silently dead", i, v, freshTrigger)
		}
	}
}

// TestUnmappedKeysAreIgnored: the browser sends every key, and most are not pad controls. It
// also pins the deliberate absence — six pressure bytes and the second stick have no name
// yet, and a key that quietly did nothing would be worse than one the vocabulary refuses.
func TestUnmappedKeysAreIgnored(t *testing.T) {
	a := padAdapter()
	for _, k := range []string{"x", "y", "black", "white", "l", "r", "Shift", "F5", "q"} {
		if err := a.Key(debug.Key{Name: k, Down: true}); err != nil {
			t.Errorf("Key(%q): %v", k, err)
		}
		if len(a.padQueue) != 0 {
			t.Fatalf("key %q queued a pad state, but no evidence names it — see usb_xid.go's "+
				"list of the ten controls this pad cannot yet spell", k)
		}
	}
}

// TestKeyRepeatCostsNoFrame: a held key repeats in the browser, and each repeat is the same
// level. Queuing it would spend a frame per repeat, and the queue drains one state per frame.
func TestKeyRepeatCostsNoFrame(t *testing.T) {
	a := padAdapter()
	press(t, a, "a", true)
	n := len(a.padQueue)
	for i := 0; i < 5; i++ {
		if err := a.Key(debug.Key{Name: "a", Down: true}); err != nil {
			t.Fatal(err)
		}
	}
	if len(a.padQueue) != n {
		t.Errorf("5 key repeats queued %d extra states, want 0", len(a.padQueue)-n)
	}
}
