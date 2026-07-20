package xbox

import "testing"

// TestFlipVSyncCadence guards the FlipVSync clock model (Machine.FlipVSync, on by default). With
// it on, each FLIP_STALL advances the guest clock by one vblank period, so OutRun's RDTSC
// fixed-timestep loop (0x20AFA) sees one field of elapsed time per present and steps its
// simulation every frame — 60 FPS, instead of the old under-paced clock's ~1-sim-step-per-6-
// presents (a 10 FPS sim on a 60 FPS engine, because an instruction-paced clock undercounts a
// rendered frame's real cycle cost). It also checks the run stays stable — no freeze, no halt —
// from the pre-race countdown onward.
//
// The frame gate already runs this path (it is the default); this test pins the MECHANISM — that
// each present costs ~1.0 vblank, not the old ~0.18 — so a regression in the clock crediting is
// caught as a cadence drift, not only as a moved frame hash. It sets the field explicitly so it
// still means something if the default is ever flipped for an A/B.
func TestFlipVSyncCadence(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the NV2A for tens of fields; -short skips it")
	}
	m, disc := atState(t, outrunImage, outrunXBE, "../../../games/outrun-2006-xbox/work/states/shadow-caster.state")
	defer disc.Close()
	m.FlipVSync = true

	last := m.tick
	n := 0
	var sum float64
	m.OnFlip = func(mm *Machine) {
		if n > 0 { // flip 1 absorbs the partial field left from the state's capture point
			sum += float64(mm.tick-last) / float64(vblankPeriod)
		}
		last = mm.tick
		n++
		if n >= 40 {
			mm.StopRequested = true
		}
	}
	m.Run(300_000_000)

	if m.CPU.Halted {
		t.Fatalf("halted after %d flips: %s", n, m.CPU.HaltReason)
	}
	if n < 40 {
		t.Fatalf("only %d flips before the budget ran out — the sequence stalled", n)
	}
	if mean := sum / float64(n-1); mean < 0.98 || mean > 1.02 {
		t.Errorf("mean vblanks/flip = %.3f, want ~1.0 (each present should cost exactly one field)", mean)
	}
}
