package gc

// run_test.go pins the run loop's stopping conditions — and specifically the tight-spin
// heuristic, which had no test at all until the spin set became a spin array.
//
// That gap was worth closing on its own account. Every other test in this package calls
// SetSpinDetect(false) before it runs, and the frame gate never spins, so the heuristic
// could have been inverted, or wired to fire always, or wired to fire never, and the whole
// suite would have stayed green. The claim that the array decides the same thing as the map
// is only worth what tests it, so: the boundary is tested from both sides.

import (
	"testing"

	"retroreverse.com/tools/cpu/gekko"
)

// PowerPC encodings, by hand, because a spin test wants the smallest possible program.
const (
	ppcNop = 0x60000000 // ori r0,r0,0
)

// ppcBranch encodes an unconditional relative branch of disp bytes. The displacement lives
// in bits 6..29 and is a byte offset with the low two bits implied zero, so it is masked
// rather than shifted.
func ppcBranch(disp int32) uint32 {
	return 0x48000000 | uint32(disp)&0x03FFFFFC
}

// spinMachine lays a loop of n instructions at 0x1000 — (n-1) nops and a branch back — and
// leaves the machine ready to run it. The loop therefore touches exactly n distinct PCs,
// which is the only variable the heuristic reads.
//
// Address translation is off (MSR is zero at reset), so the effective address is the
// physical one and the program is simply where it was written. MSR[EE] is zero too, which
// matters more than it looks: the video clock raises a retrace interrupt every two million
// instructions, and a taken exception would jump the PC out of the loop and add distinct
// addresses the test never asked for.
func spinMachine(t *testing.T, n int) *Machine {
	t.Helper()
	if n < 1 {
		t.Fatalf("a loop needs at least one instruction, got %d", n)
	}
	m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
	m.CPU = gekko.NewCPU(m)

	const base = 0x1000
	for i := 0; i < n-1; i++ {
		m.setRAM32(base+uint32(i*4), ppcNop)
	}
	m.setRAM32(base+uint32((n-1)*4), ppcBranch(int32(-4*(n-1))))
	m.CPU.PC = base
	return m
}

// TestSpinDetectFiresOnATightLoop: the case the heuristic exists for. `b .` is one PC.
func TestSpinDetectFiresOnATightLoop(t *testing.T) {
	m := spinMachine(t, 1)
	res := m.Run(8 * 1024 * 1024)
	if res.Reason != "spin (tight loop)" {
		t.Fatalf("a one-instruction loop should be a spin, got %q at 0x%08X after %d steps",
			res.Reason, res.PC, res.Steps)
	}
}

// TestSpinDetectBoundary is the whole equivalence argument, tested from both sides.
//
// The predicate is "fewer than four distinct PCs in the window". Three must fire and four
// must not — and four must not fire NO MATTER HOW LONG IT RUNS, which is the property that
// lets the counter saturate and stop looking. A rewrite that counted occurrences instead of
// distinct addresses, or that saturated at the wrong number, fails exactly here.
func TestSpinDetectBoundary(t *testing.T) {
	for _, c := range []struct {
		distinctPCs int
		wantSpin    bool
	}{
		{1, true},
		{2, true},
		{3, true},  // three distinct: still a spin
		{4, false}, // four distinct: not a spin, and never will be
		{5, false},
		{8, false},
	} {
		m := spinMachine(t, c.distinctPCs)
		res := m.Run(8 * 1024 * 1024)
		gotSpin := res.Reason == "spin (tight loop)"
		if gotSpin != c.wantSpin {
			t.Errorf("a %d-PC loop: spin = %v, want %v (reason %q after %d steps)",
				c.distinctPCs, gotSpin, c.wantSpin, res.Reason, res.Steps)
		}
		if !c.wantSpin && res.Reason != "step budget exhausted" {
			t.Errorf("a %d-PC loop should have run out its budget, got %q",
				c.distinctPCs, res.Reason)
		}
	}
}

// TestSpinDetectOffNeverFires: -nospin is what a boot that means to park in a wait loop
// needs, and it is also the flag the spin cost was measured by A/B-ing. It must turn the
// heuristic off completely, not merely make it rarer.
func TestSpinDetectOffNeverFires(t *testing.T) {
	m := spinMachine(t, 1)
	m.SetSpinDetect(false)
	res := m.Run(8 * 1024 * 1024)
	if res.Reason != "step budget exhausted" {
		t.Fatalf("with spin detection off, a tight loop should run its budget out, got %q", res.Reason)
	}
}

// TestSpinWindowResets checks the part a saturating counter is most likely to get wrong: the
// window is a WINDOW, so the count must start again at each boundary. A counter that
// saturated at four and never reset would never convict anything again, and the heuristic
// would be silently dead — which is the failure this whole file exists to make loud.
//
// So the program starts as a four-PC loop (not a spin) and BECOMES a one-PC spin. The first
// window sees four distinct PCs and must acquit. A later window sees one and must convict.
// If the counter does not reset, the second window still believes it has seen four, the run
// reaches its budget, and this fails.
//
// THE INNOCENT LOOP MUST DO SOMETHING, and the first version of this test forgot: it was three
// nops and a branch, which changes no register and stores nothing — an idle loop by the exact
// definition in idle.go, so the fast-forward skipped straight over it and the patch below,
// which was counting trips round the loop, never landed. The loop counts a register up
// instead, which is what an honest not-idle loop looks like.
func TestSpinWindowResets(t *testing.T) {
	m := spinMachine(t, 4)
	// addi r3,r3,1 in place of the first nop: the loop now makes progress, so it is not idle
	// and must be interpreted rather than skipped.
	m.setRAM32(0x1000, 0x38630001)

	// Well inside the first window: the loop is still innocent when the patch lands. Counted
	// in EMULATED instructions, because that is the clock the spin window is on and the one
	// the fast-forward advances.
	const patchAt = 0x100000
	m.OnStep = func(mm *Machine, pc uint32) {
		if mm.Instrs == patchAt {
			// The loop's branch becomes a branch to itself. From here the program is stuck
			// on one address — a real spin, arriving after an innocent stretch.
			mm.setRAM32(0x100C, ppcBranch(0))
		}
	}

	// Budget enough to cross two window boundaries: the first acquits (four distinct PCs
	// were seen before the patch), the second must convict.
	res := m.Run(12 * 1024 * 1024)
	if res.Reason != "spin (tight loop)" {
		t.Fatalf("a loop that turned into a spin must be convicted by a later window; "+
			"got %q at 0x%08X after %d steps (a counter that never resets looks exactly like this)",
			res.Reason, res.PC, res.Steps)
	}
	if res.PC != 0x100C {
		t.Errorf("convicted at 0x%08X, want the spinning branch at 0x0000100C", res.PC)
	}
	if m.CPU.Halted {
		t.Fatalf("the machine halted rather than spinning: %s", m.CPU.HaltReason)
	}
}
