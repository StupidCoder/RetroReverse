package ps2

// rrv_idle_test.go checks the idle fast-forward (idle.go) on the repo's OTHER PlayStation 2
// game, which paces itself differently from Jak — Ridge Racer V polls I_STAT in a userspace
// loop woken by a priority-0 callback, and streams its menus in over the SIF. It is the case
// the IOP veto (idle.go / IOP.BusyToEE) exists for: skipping the EE while the second
// processor reaches across the SIF cannot reproduce the interleave, so the skip refuses to
// start while the IOP is streaming.
//
// WHERE IT STANDS, honestly: on a title screen and on the intro the machine is byte-identical
// with the skip on or off. On the loaded main menu — where the IOP answers RPCs the veto has
// no lead on (a reply armed and fired inside a single IOP step) — the RENDER is byte-identical
// (the GS VRAM matches exactly), but main memory and the EE's instruction clock drift: the
// skip re-phases the every-thread-blocked span, and a value the game reads back from a timer
// lands a few ticks off. It is a real gap in the disturbance handling, asserted below as the
// render being exact rather than the whole machine, so the boundary is visible rather than
// hidden. The Jak whole-boot pin ([[ps2-platform]]) and TestFrameHashes both hold with the
// skip on, because the veto keeps it off through the disc loading that a boot is.

import (
	"crypto/md5"
	"fmt"
	"os"
	"testing"

	"retroreverse.com/tools/lib/iso9660"
)

const (
	rrvImage    = "../../../games/ridge-racer-v-ps2/image/Ridge Racer V (USA).bin"
	rrvBIOS     = "../../../games/ridge-racer-v-ps2/image/scph10000.bin"
	rrvStateDir = "../../../games/ridge-racer-v-ps2/work/states/"
)

func rrvAt(tb testing.TB, state string) *Machine {
	tb.Helper()
	raw, err := os.ReadFile(rrvImage)
	if err != nil {
		tb.Skipf("RRV image not present: %v", err)
	}
	bios, err := os.ReadFile(rrvBIOS)
	if err != nil {
		tb.Skipf("RRV BIOS not present: %v", err)
	}
	vol, err := iso9660.OpenBytes(raw)
	if err != nil {
		tb.Fatal(err)
	}
	exePath, err := bootExe(vol)
	if err != nil {
		tb.Fatal(err)
	}
	elfRaw, err := vol.ReadFile(exePath)
	if err != nil {
		tb.Fatal(err)
	}
	exe, err := LoadELF(elfRaw)
	if err != nil {
		tb.Fatal(err)
	}
	m := NewMachine()
	m.SetImageHash(fmt.Sprintf("%x", md5.Sum(raw)))
	m.SetVolume(vol)
	m.SetBIOS(bios)
	m.LoadExecutable(exe)
	if err := m.LoadStateFile(rrvStateDir + state); err != nil {
		tb.Skipf("no RRV state %s (work/ is not committed): %v", state, err)
	}
	return m
}

func rrvHashes(t *testing.T, state string, skip bool) (ram, vram, cpu string, hits uint64) {
	m := rrvAt(t, state)
	m.SetIdleSkip(skip)
	runFields(m, benchFields, benchBudget)
	r, v, c := gateHashes(m)
	_, h := m.IdleStats()
	return r, v, c, h
}

// TestRRVIdleSkipMatchesSerial: on the states with no IOP streaming mid-skip, every hash is
// identical serial vs skipped. It is the byte-for-byte claim, the same one the Jak test makes.
func TestRRVIdleSkipMatchesSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for several fields twice; -short skips it")
	}
	if raceBuild {
		t.Skip("-race changes the floating-point result (FMA contraction)")
	}
	for _, st := range []string{"title.state", "intro.state"} {
		st := st
		t.Run(st, func(t *testing.T) {
			rOff, vOff, cOff, _ := rrvHashes(t, st, false)
			rOn, vOn, cOn, hits := rrvHashes(t, st, true)
			for _, c := range []struct{ name, off, on string }{
				{"ram", rOff, rOn}, {"vram", vOff, vOn}, {"cpu", cOff, cOn},
			} {
				if c.off != c.on {
					t.Errorf("%s differs between serial and idle-skipped runs", c.name)
				}
			}
			if !t.Failed() {
				t.Logf("byte-identical; idle skip fired %d times", hits)
			}
		})
	}
}

// TestRRVIdleSkipRenderExactOnMenu pins the KNOWN boundary: on the loaded main menu the skip
// keeps the render exact (GS VRAM identical) but not main memory or the instruction clock. It
// asserts the render — the oracle's product — and reports the RAM/CPU drift rather than hiding
// it, so a future fix to the disturbance handling turns this into a full match and folds back
// into the test above.
func TestRRVIdleSkipRenderExactOnMenu(t *testing.T) {
	if testing.Short() || raceBuild {
		t.Skip("full-GS byte comparison; skipped under -short/-race")
	}
	rOff, vOff, cOff, _ := rrvHashes(t, "main menu.state", false)
	rOn, vOn, cOn, hits := rrvHashes(t, "main menu.state", true)
	if vOn != vOff {
		t.Errorf("VRAM (the render) differs — that is a real break, not the known clock drift")
	}
	if rOn == rOff && cOn == cOff {
		t.Logf("byte-identical (the disturbance gap appears closed — fold into TestRRVIdleSkipMatchesSerial)")
	} else {
		t.Logf("KNOWN LIMITATION: main memory / EE clock drift on the streaming menu (ram match=%v cpu match=%v); the render is exact. See file comment. Skip fired %d times.",
			rOn == rOff, cOn == cOff, hits)
	}
}
