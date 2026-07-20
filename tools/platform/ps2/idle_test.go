package ps2

// idle_test.go is the real correctness claim for the idle fast-forward (idle.go): not
// "the same as a pinned hash from last week" but "the same machine, instruction for
// instruction, as if it had never skipped". A pinned frame (bench_test.go) guards against
// change; this guards the skip against being change.

import "testing"

// TestIdleSkipMatchesSerial runs the same fields twice — once fast-forwarding idle loops,
// once interpreting every instruction — and demands byte-identical memory AND CPU state.
// This is the claim the whole optimisation rests on; a pinned hash cannot make it.
func TestIdleSkipMatchesSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for several fields twice; -short skips it")
	}
	if raceBuild {
		t.Skip("-race changes the floating-point result (FMA contraction); run the ordinary build")
	}

	run := func(skip bool) (ram, vram, cpu string, hits uint64) {
		m := atState(t, jakImage, stateTitle)
		m.SetIdleSkip(skip)
		runFields(m, benchFields, benchBudget)
		r, v, c := gateHashes(m)
		sk, h := m.IdleStats()
		_ = sk
		return r, v, c, h
	}

	ramOff, vramOff, cpuOff, _ := run(false)
	ramOn, vramOn, cpuOn, hits := run(true)

	if hits == 0 {
		t.Fatal("idle skip fired zero times over the title field — the test proves nothing; is the detector wired in?")
	}
	for _, c := range []struct{ name, off, on string }{
		{"ram", ramOff, ramOn},
		{"vram", vramOff, vramOn},
		{"cpu", cpuOff, cpuOn},
	} {
		if c.off != c.on {
			t.Errorf("%s hash differs between serial and idle-skipped runs:\n  serial  %s\n  skipped %s", c.name, c.off, c.on)
		}
	}
	if !t.Failed() {
		t.Logf("idle skip fired %d times and left the machine byte-identical", hits)
	}
}

// TestIdleSkipActuallyUsed stops TestIdleSkipMatchesSerial being a tautology: if the skip
// never fired, "skip == serial" would hold trivially. It asserts the title field really is
// fast-forwarded, and reports how far.
func TestIdleSkipActuallyUsed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for a field; -short skips it")
	}
	m := atState(t, jakImage, stateTitle)
	m.SetIdleSkip(true) // opt in — the fast-forward is off by default (idle.go)
	before := m.steps
	runFields(m, benchFields, benchBudget)
	skipped, hits := m.IdleStats()
	if hits == 0 || skipped == 0 {
		t.Fatalf("idle skip did not fire over the title field (hits=%d skipped=%d)", hits, skipped)
	}
	ran := m.steps - before
	t.Logf("%d/%d instruction-slots fast-forwarded (%.0f%%) over %d idle loops",
		skipped, ran, 100*float64(skipped)/float64(ran), hits)
	if skipped*2 < ran {
		t.Logf("note: less than half the field was skipped — the VSync idle was ~72%% by eeprof")
	}
}
