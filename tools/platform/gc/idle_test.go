package gc

// idle_test.go carries the correctness claim for the idle fast-forward.
//
// The pinned hashes in bench_test.go do gate it — the skip is on by default and they did not
// move — but they cannot make the claim on their own, for the same reason they cannot make it
// for the parallel rasteriser: they say "the same as last week", and the question here is
// "the same as this build with the skip turned off". That is what these tests ask, on both
// states, on every hash the gate knows plus the clocks the game can read.
//
// The first version of the skip PASSED the picture and failed this, by two bytes: it jumped to
// the raw deadline, so the retrace interrupt always landed on the PC the detector happened to
// snapshot rather than wherever the loop had actually got to. The game's own frame-time
// counter noticed. That is why idleSkip rounds to a whole number of loop periods, and why this
// file compares RAM and not merely the frame.

import (
	"testing"

	"retroreverse.com/tools/cpu/gekko"
)

// TestIdleSkipChangesNothing is the claim: skip on and skip off produce the same machine.
func TestIdleSkipChangesNothing(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the machine for several fields twice; -short skips it")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			off := atState(t, p.state)
			off.SetIdleSkip(false)
			defer off.Close()
			off.RunFields(benchFields, benchBudget)

			on := atState(t, p.state)
			defer on.Close()
			on.RunFields(benchFields, benchBudget)

			ramA, xfbA, aramA, cpuA := gateHashes(off)
			ramB, xfbB, aramB, cpuB := gateHashes(on)
			for _, c := range []struct{ name, a, b string }{
				{"ram", ramA, ramB},
				{"xfb", xfbA, xfbB},
				{"aram", aramA, aramB},
				{"cpu", cpuA, cpuB},
			} {
				if c.a != c.b {
					t.Errorf("%s differs with the idle skip on:\n off %s\n on  %s", c.name, c.a, c.b)
				}
			}

			// The clocks the GAME can read must be identical, not merely close. TB is what
			// its frame timer samples; Instrs is what every device here is paced by. Only
			// the emulator's own count of interpreted instructions is allowed to differ,
			// and that is the entire point of the exercise.
			if off.CPU.TB != on.CPU.TB {
				t.Errorf("time base differs: off %d, on %d", off.CPU.TB, on.CPU.TB)
			}
			if off.CPU.DEC != on.CPU.DEC {
				t.Errorf("decrementer differs: off 0x%08X, on 0x%08X", off.CPU.DEC, on.CPU.DEC)
			}
			if off.Instrs != on.Instrs {
				t.Errorf("instruction clock differs: off %d, on %d", off.Instrs, on.Instrs)
			}
			if off.vi.Field != on.vi.Field {
				t.Errorf("video field differs: off %d, on %d", off.vi.Field, on.vi.Field)
			}
		})
	}
}

// TestIdleSkipIsActuallyUsed stops the test above from being a tautology: a detector that
// never fired would agree with itself perfectly and the machine would simply be slow.
func TestIdleSkipIsActuallyUsed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the machine for several fields; -short skips it")
	}
	m := atState(t, stateCutscene)
	defer m.Close()
	res := m.RunFields(benchFields, benchBudget)
	skipped, hits := m.IdleStats()

	if hits == 0 {
		t.Fatal("the idle detector never fired; every other claim in this file is vacuous")
	}
	// res.Steps is EMULATED instructions and so already counts the skipped ones — the whole
	// point of the budget change in run.go. The interpreter only ran the difference.
	share := float64(skipped) / float64(res.Steps) * 100
	t.Logf("%d emulated instructions, of which %d were SKIPPED in %d hits — the interpreter ran %d (%.1f%% avoided)",
		res.Steps, skipped, hits, res.Steps-skipped, share)

	// The histogram that motivated this put the idle loop at 79.7% of the cutscene's
	// instructions. Anything far below that means the detector has stopped finding it — a
	// silent regression that nothing else here would catch, because the machine would still
	// be perfectly correct.
	if share < 50 {
		t.Errorf("only %.1f%% of the field was skipped; the idle loop is 79.7%% of it, so the "+
			"detector has stopped recognising it", share)
	}
}

// TestIdleSkipOffIsSilent: the switch must turn the thing off completely, since it is what a
// bisect uses to rule the skip out.
func TestIdleSkipOffIsSilent(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the machine for several fields; -short skips it")
	}
	m := atState(t, stateCutscene)
	m.SetIdleSkip(false)
	defer m.Close()
	m.RunFields(benchFields, benchBudget)
	if skipped, hits := m.IdleStats(); skipped != 0 || hits != 0 {
		t.Errorf("with the skip off, %d instructions were skipped in %d hits", skipped, hits)
	}
}

// TestIdleSkipRoundsToWholeLoops pins the property the exactness rests on, directly rather
// than through a whole field: a skip must never leave the processor at a PC the unskipped run
// would not have been at, so it must advance a multiple of the proved loop period.
func TestIdleSkipRoundsToWholeLoops(t *testing.T) {
	m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
	m.CPU = gekko.NewCPU(m)

	for _, period := range []int{1, 3, 5, 7, 64} {
		for _, want := range []uint64{0, 1, 2, 1000, 999_999} {
			m.idle = idleState{period: period}
			m.vi.Counter = 0
			m.idleSkip(want)
			got := m.idle.Skipped
			if got%uint64(period) != 0 {
				t.Errorf("period %d, asked for %d: skipped %d, which is not a whole number of loops",
					period, want, got)
			}
			if got > want {
				t.Errorf("period %d, asked for %d: skipped %d, which is further than allowed",
					period, want, got)
			}
			if want >= uint64(period) && got == 0 {
				t.Errorf("period %d, asked for %d: skipped nothing", period, want)
			}
			// The clock must have moved by exactly what was skipped.
			if uint32(got) != m.vi.Counter {
				t.Errorf("period %d: skipped %d but the video clock moved %d", period, got, m.vi.Counter)
			}
		}
	}
}
