package debug

import (
	"reflect"
	"testing"
)

// The writer census is called once per stored fragment — several hundred thousand times in
// a 3DS frame — so what it must get right is: the order is execution order, a command that
// wrote a million pixels appears once, and a command whose fragments were interleaved with
// another's still appears once.
func TestMarkWrite(t *testing.T) {
	var fc FrameCapture
	fc.CountWrites()

	if fc.Writers == nil {
		t.Fatal("a counting capture that drew nothing must say [] — not null, which means nobody counted")
	}

	// Command 0 draws: the first write must be recorded, which it would not be if the
	// no-op guard started life at index 0.
	fc.MarkWrite(0)
	for i := 0; i < 1000; i++ {
		fc.MarkWrite(0) // the same command, a thousand more fragments
	}
	fc.MarkWrite(7)
	fc.MarkWrite(7)
	fc.MarkWrite(3) // a platform that interleaves two commands' fragments...
	fc.MarkWrite(7) // ...and comes back to one already listed
	fc.MarkWrite(9)

	if want := []int{0, 7, 3, 9}; !reflect.DeepEqual(fc.Writers, want) {
		t.Errorf("Writers = %v, want %v", fc.Writers, want)
	}
}

// A capture that never counted says so with a nil list, and the page falls back to scrubbing
// the whole stream. "Nobody counted" and "nothing drew" must not look alike.
func TestUncountedCaptureHasNoWriters(t *testing.T) {
	var fc FrameCapture
	if fc.Writers != nil {
		t.Errorf("Writers = %v, want nil", fc.Writers)
	}
}
