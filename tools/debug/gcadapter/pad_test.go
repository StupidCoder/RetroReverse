package gcadapter

// pad_test.go pins the keyboard-to-controller mapping, and mostly it pins the stick.
//
// The arrows used to be the d-pad. They are the main stick now, because that is what it takes
// to play a GameCube game — and the interesting half of that change is not "the arrows moved",
// it is that A KEYBOARD HAS NO GATE AND THE STICK'S SHELL DOES. Every claim below is about one
// of those two things.
//
// None of it needs the disc. Key() and stickFrom touch nothing but the adapter's own two pad
// fields, so these run wherever the code does — which matters, because the game image is not
// committed and every test that needs it skips.

import (
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/platform/gc"
)

func padAdapter() *Adapter { return &Adapter{held: map[string]bool{}} }

func mustButton(t *testing.T, name string) uint16 {
	t.Helper()
	b, ok := gc.PadButton(name)
	if !ok {
		t.Fatalf("no pad button %q", name)
	}
	return b
}

// press drives a key down (or up) and returns the pad state the queue now ends on — what the
// next field poll will latch.
func press(t *testing.T, a *Adapter, name string, down bool) padState {
	t.Helper()
	if err := a.Key(debug.Key{Name: name, Down: down}); err != nil {
		t.Fatal(err)
	}
	if len(a.padQueue) == 0 {
		t.Fatalf("%s down=%v queued nothing", name, down)
	}
	return a.padQueue[len(a.padQueue)-1]
}

// TestArrowsDriveTheStickNotTheDPad is the change itself, stated as a test. Both halves
// matter: an arrow must move the stick, AND it must leave the d-pad bits alone.
func TestArrowsDriveTheStickNotTheDPad(t *testing.T) {
	dpad := mustButton(t, "up") | mustButton(t, "down") | mustButton(t, "left") | mustButton(t, "right")

	for _, c := range []struct {
		key          string
		wantX, wantY uint8
	}{
		// The pad reports UP as INCREASING — the opposite of every screen coordinate here,
		// and the reason the first cut of this shipped the stick upside down: the direction
		// table was written in the pad's convention AND negated on the way out. There is no
		// inversion to do; these four lines are what say so.
		{"ArrowUp", gc.PadStickCentre, gc.PadStickCentre + gc.PadStickFull},
		{"ArrowDown", gc.PadStickCentre, gc.PadStickCentre - gc.PadStickFull},
		{"ArrowLeft", gc.PadStickCentre - gc.PadStickFull, gc.PadStickCentre},
		{"ArrowRight", gc.PadStickCentre + gc.PadStickFull, gc.PadStickCentre},
	} {
		a := padAdapter()
		st := press(t, a, c.key, true)
		if st.stickX != c.wantX || st.stickY != c.wantY {
			t.Errorf("%s: stick = (0x%02X,0x%02X), want (0x%02X,0x%02X)",
				c.key, st.stickX, st.stickY, c.wantX, c.wantY)
		}
		if st.buttons&dpad != 0 {
			t.Errorf("%s pressed a d-pad button (buttons 0x%04X); the arrows are the STICK now",
				c.key, st.buttons)
		}
	}
}

// TestStickDiagonalRespectsTheGate is the claim that is easy to get wrong and invisible when
// you do: the stick's shell is an octagon whose eight notches are all the same distance from
// centre, so a corner reads about 0.7 of full on EACH axis. Driving both axes to full would
// hand the game a diagonal 1.41x longer than any physical pad can produce — a position the
// game is entitled to be confused by, and one no amount of looking at the picture would
// explain.
func TestStickDiagonalRespectsTheGate(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowUp", true)
	st := press(t, a, "ArrowRight", true)

	dx := int(st.stickX) - gc.PadStickCentre
	dy := int(st.stickY) - gc.PadStickCentre
	if dx <= 0 || dy <= 0 {
		t.Fatalf("up+right should push the stick up and right, got (0x%02X,0x%02X)", st.stickX, st.stickY)
	}
	if dx >= gc.PadStickFull || dy >= gc.PadStickFull {
		t.Errorf("up+right = (%+d,%+d): a diagonal must not reach full deflection on either axis "+
			"— the gate stops a real stick before it gets there", dx, dy)
	}
	// A corner of the octagon is as far from centre as a notch, so the magnitude should match
	// a cardinal push. Squared, with room for the rounding in padStickDiag.
	mag2, full2 := dx*dx+dy*dy, gc.PadStickFull*gc.PadStickFull
	if mag2 < full2*9/10 || mag2 > full2*11/10 {
		t.Errorf("up+right magnitude^2 = %d, want ~%d (a corner is as far out as a notch)", mag2, full2)
	}
}

// TestOppositeStickKeysCancel: a stick cannot be in two places, and a real one held both ways
// sits in the middle.
func TestOppositeStickKeysCancel(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowLeft", true)
	st := press(t, a, "ArrowRight", true)
	if st.stickX != gc.PadStickCentre || st.stickY != gc.PadStickCentre {
		t.Errorf("left+right = (0x%02X,0x%02X), want centred", st.stickX, st.stickY)
	}
}

// TestStickReleasesToCentre: letting go must return the stick, or the game walks forever.
func TestStickReleasesToCentre(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowUp", true)
	st := press(t, a, "ArrowUp", false)
	if st.stickX != gc.PadStickCentre || st.stickY != gc.PadStickCentre {
		t.Errorf("after release the stick is (0x%02X,0x%02X), want centred", st.stickX, st.stickY)
	}
}

// TestNumpadStillDrivesTheDPad: the d-pad was MOVED, not dropped.
func TestNumpadStillDrivesTheDPad(t *testing.T) {
	for _, c := range []struct{ key, button string }{
		{"8", "up"}, {"2", "down"}, {"4", "left"}, {"6", "right"},
	} {
		a := padAdapter()
		st := press(t, a, c.key, true)
		want := mustButton(t, c.button)
		if st.buttons&want == 0 {
			t.Errorf("%q should press d-pad %s (0x%04X), got buttons 0x%04X",
				c.key, c.button, want, st.buttons)
		}
		if st.stickX != gc.PadStickCentre || st.stickY != gc.PadStickCentre {
			t.Errorf("%q moved the stick to (0x%02X,0x%02X); the d-pad is not the stick",
				c.key, st.stickX, st.stickY)
		}
	}
}

// TestStickAndButtonsQueueTogether: one auto-poll latches the buttons and both axes at once,
// so a queued state carrying a new stick position and last field's buttons would be a
// controller that never existed. The queue therefore carries a whole pad.
func TestStickAndButtonsQueueTogether(t *testing.T) {
	a := padAdapter()
	press(t, a, "a", true)
	st := press(t, a, "ArrowLeft", true)

	if st.buttons&mustButton(t, "a") == 0 {
		t.Error("pushing the stick dropped the held A: the state must carry both")
	}
	if st.stickX != gc.PadStickCentre-gc.PadStickFull {
		t.Errorf("stick = 0x%02X, want pushed left while A is held", st.stickX)
	}
}

// TestUnmappedKeyQueuesNoPadState: the browser forwards every keystroke, so the ones the pad
// has no use for must be dropped without costing a field.
func TestUnmappedKeyQueuesNoPadState(t *testing.T) {
	a := padAdapter()
	if err := a.Key(debug.Key{Name: "F7", Down: true}); err != nil {
		t.Fatal(err)
	}
	if len(a.padQueue) != 0 {
		t.Errorf("an unmapped key queued %d pad state(s); it should cost no field", len(a.padQueue))
	}
}

// TestRepeatCostsNoField: a held key's browser auto-repeat is the same level, and a queue that
// took a field for it would make the pad lag behind the player.
func TestRepeatCostsNoField(t *testing.T) {
	a := padAdapter()
	press(t, a, "ArrowUp", true)
	n := len(a.padQueue)
	if err := a.Key(debug.Key{Name: "ArrowUp", Down: true}); err != nil {
		t.Fatal(err)
	}
	if len(a.padQueue) != n {
		t.Errorf("a repeated key queued another state (%d -> %d)", n, len(a.padQueue))
	}
}
