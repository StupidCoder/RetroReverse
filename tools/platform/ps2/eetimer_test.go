package ps2

import "testing"

// TestEETimerFieldRatio pins the contract the GOAL engine's frame pacing rests on:
// Timer 1 on BUSCLK/256 accrues just UNDER *ticks-per-frame* (0x2625 = 9765) in one
// field, so display-frame-start's ratio = count/9765 + 1 is 1 at healthy 60 fps —
// and 2 when a frame slips to two fields. The unmodelled-io fiction (reads return
// the last write) had the boot believing whole minutes per frame.
func TestEETimerFieldRatio(t *testing.T) {
	m := NewMachine()
	const t1 = 0x10000800

	m.ioWrite(t1+0x10, 0x82) // MODE: CLKS=/256, CUE
	m.ioWrite(t1, 0)         // COUNT = 0

	m.steps += stepsPerVBlank // one field passes
	count := m.ioRead(t1)
	if count == 0 || count >= 9765 {
		t.Errorf("one field on BUSCLK/256: count %d, want 0 < count < 9765", count)
	}
	if ratio := count/9765 + 1; ratio != 1 {
		t.Errorf("one-field ratio: got %d, want 1", ratio)
	}

	m.steps += stepsPerVBlank // a lagged frame: two fields since the reset
	if ratio := m.ioRead(t1)/9765 + 1; ratio != 2 {
		t.Errorf("two-field ratio: got %d, want 2", ratio)
	}

	// A COUNT write resets the epoch, exactly like the engine's timer-reset.
	m.ioWrite(t1, 0)
	if c := m.ioRead(t1); c != 0 {
		t.Errorf("count after reset: got %d, want 0", c)
	}

	// CUE off freezes the counter.
	m.steps += stepsPerVBlank / 2
	frozen := m.ioRead(t1)
	m.ioWrite(t1+0x10, 0x02) // CUE clear, same clock
	m.steps += stepsPerVBlank
	if c := m.ioRead(t1); c != frozen {
		t.Errorf("frozen counter moved: %d -> %d", frozen, c)
	}
}
