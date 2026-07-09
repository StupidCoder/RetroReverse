package n64

// Conformance against lemmy-64/n64-systemtest, a self-checking homebrew ROM that
// exercises the VR4300, the RCP and the RSP against real-hardware behaviour and
// reports each test by name.
//
// There is no published per-instruction suite for this CPU the way there is for
// the R3000A (SingleStepTests/psx), so this stands in its place. It is a better
// instrument than a rendered frame: a failure names one instruction or one
// register, where a wrong image names nothing.
//
// The ROM prints to the IS-Viewer as well as to the screen, which is what makes
// it readable from a program (see isviewer.go).
//
// Point N64_SYSTEMTEST at the ROM; the test skips when it is unset, so `go test`
// runs without it. To build the ROM:
//
//	rustup toolchain install nightly && cargo install nust64
//	git clone https://github.com/lemmy-64/n64-systemtest && cd n64-systemtest
//	cargo run --release
//	# target/mips-nintendo64-none/release/n64-systemtest.z64
//
// Not every test is expected to pass, and chasing all of them is not the point.
// The ROM deliberately executes illegal encodings and probes cartridge-bus
// timing quirks that no game depends on. What this asserts is that the machine
// boots the ROM, that the ROM's own startup check of the boot state passes, and
// that a substantial number of tests run. The named failures are logged, and
// they are the work list.

import (
	"os"
	"strings"
	"testing"
)

func TestN64SystemTest(t *testing.T) {
	path := os.Getenv("N64_SYSTEMTEST")
	if path == "" {
		t.Skip("set N64_SYSTEMTEST to a built n64-systemtest.z64 (see the file comment)")
	}
	rom, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", rom)

	m := NewMachine(rom)
	if err := m.Boot(rom, DefaultBoot()); err != nil {
		t.Fatal(err)
	}
	// The ROM parks in loops on purpose; the oracle's tight-loop heuristic is a
	// convenience for debugging games, not a hardware behaviour.
	m.SetSpinDetect(false)

	var out strings.Builder
	m.OnPrint = func(_ *Machine, line string) { out.WriteString(line) }

	res := m.Run(3_000_000_000)
	text := out.String()
	if text == "" {
		t.Fatalf("the ROM printed nothing: %s", res)
	}

	// The ROM prints in fragments, so a test's name and its verdict arrive
	// separately. Splitting on the "Test '" that precedes every failure recovers
	// the name.
	var failed []string
	for _, seg := range strings.Split(text, "Test '")[1:] {
		if !strings.Contains(seg, "failed") {
			continue
		}
		name := seg
		if i := strings.Index(seg, "'"); i > 0 {
			name = seg[:i]
		}
		failed = append(failed, strings.TrimSpace(name))
	}
	ran := strings.Count(text, "Running ")

	// StartupTest checks the state the boot handoff left behind — the Status
	// register above all. If that is wrong, everything after it is measured
	// against the wrong machine.
	for _, f := range failed {
		if f == "StartupTest" {
			t.Error("StartupTest failed: the PIF boot handoff leaves the machine in the wrong state")
		}
	}
	if ran < 50 {
		t.Errorf("only %d tests ran before %s: the ROM did not get far", ran, res)
	}

	t.Logf("%s", res)
	t.Logf("%d tests started, %d reported a failure", ran, len(failed))
	for _, f := range failed {
		t.Logf("  fail: %s", f)
	}
	if m.RSP != nil && len(m.RSP.Unimplemented) > 0 {
		t.Logf("RSP encodings met but not modelled: %d distinct", len(m.RSP.Unimplemented))
		for w, n := range m.RSP.Unimplemented {
			t.Logf("  0x%08X  x%d", w, n)
		}
	}
}
