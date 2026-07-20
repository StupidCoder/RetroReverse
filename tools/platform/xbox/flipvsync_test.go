package xbox

import "testing"

// TestFlipVSyncCadence guards the FlipVSync fidelity experiment (Machine.FlipVSync). With it on,
// each FLIP_STALL advances the guest clock by one vblank period, so OutRun's RDTSC fixed-timestep
// loop (0x20AFA) sees one field of elapsed time per present and steps its simulation every frame
// — 60 FPS, instead of the default path's ~1-sim-step-per-6-presents (a 10 FPS sim on a 60 FPS
// engine, because the instruction-paced clock undercounts a rendered frame's real cycle cost).
// It also checks the run stays stable — no freeze, no halt — from the pre-race countdown onward.
//
// This is the A/B that proved the fix: the default path leaves the tick advancing only by the
// render code retired (~0.18 of a field per present); this asserts the experiment brings it to
// ~1.0. The experiment is off by default (it re-times every trajectory), so the frame gate stays
// byte-identical and this is the only thing that exercises the on path.
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
