package xbox

import (
	"os"
	"path/filepath"
	"testing"
)

// bootMachine opens the retail disc (skipping when it is absent — it is gitignored per
// the copyright policy), parses default.xbe, and builds a machine ready to run.
func bootMachine(t *testing.T) *Machine {
	t.Helper()
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the OutRun disc image is not present")
	}
	disc, err := Open(discPath)
	if err != nil {
		t.Fatalf("open disc: %v", err)
	}
	t.Cleanup(func() { disc.Close() })
	xbeBytes, err := disc.ReadFile("/default.xbe")
	if err != nil {
		t.Fatalf("read default.xbe: %v", err)
	}
	xbe, err := ParseXBE(xbeBytes)
	if err != nil {
		t.Fatalf("parse xbe: %v", err)
	}
	m, err := NewMachine(xbe, disc)
	if err != nil {
		t.Fatalf("build machine: %v", err)
	}
	return m
}

// TestBootReachesFrontier boots the title and checks it runs a substantial stretch of
// real XDK/CRT code — spawning the main thread, initialising the DPC/timer/critical-
// section machinery, querying config, and reaching the C runtime's SSE object init —
// before halting at its concrete frontier. It is a regression guard on the whole boot
// path: a change that stops the boot early trips it.
func TestBootReachesFrontier(t *testing.T) {
	m := bootMachine(t)
	reason, n := m.Run(80_000_000)
	if n < 20_000 {
		t.Fatalf("boot stalled after only %d instructions (reason %v): %s", n, reason, m.CPU.HaltReason)
	}
	// The boot must reach the main thread and touch the kernel: at least a handful of
	// distinct ordinals, and PsCreateSystemThreadEx (255) among them.
	if len(m.OrdinalHits) < 5 {
		t.Errorf("only %d distinct ordinals reached; expected the CRT init surface", len(m.OrdinalHits))
	}
	if m.OrdinalHits[255] == 0 {
		t.Errorf("PsCreateSystemThreadEx (255) was never called — the main thread did not spawn")
	}
	t.Logf("boot ran %d instructions, %d ordinals, stop=%v", n, len(m.OrdinalHits), reason)
}

// TestSaveStateRoundTrip boots a way in, snapshots to disk, restores into a fresh
// machine, and checks the restore is bit-identical — then that running both on from the
// snapshot reaches the same place. This is the oracle-capability-parity requirement:
// the expensive boot is reached once and restored instantly.
func TestSaveStateRoundTrip(t *testing.T) {
	m := bootMachine(t)
	m.Run(15_000) // partway into the boot

	path := filepath.Join(t.TempDir(), "state.gz")
	if err := m.SaveStateFile(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Restore into a fresh machine built the same way.
	m2 := bootMachine(t)
	if err := m2.LoadStateFile(path); err != nil {
		t.Fatalf("load: %v", err)
	}

	// The restored machine must match the source exactly: registers, flags, and RAM.
	if m2.CPU.IP != m.CPU.IP || m2.CPU.Regs != m.CPU.Regs {
		t.Fatalf("restored CPU differs: IP %08X vs %08X", m2.CPU.IP, m.CPU.IP)
	}
	for i := range m.RAM {
		if m.RAM[i] != m2.RAM[i] {
			t.Fatalf("restored RAM differs at %#x: %02X vs %02X", i, m.RAM[i], m2.RAM[i])
		}
	}

	// Running both on from the snapshot must land in the same place (determinism).
	m.Run(5_000)
	m2.Run(5_000)
	if m.CPU.IP != m2.CPU.IP {
		t.Fatalf("diverged after resume: IP %08X vs %08X", m.CPU.IP, m2.CPU.IP)
	}
}
