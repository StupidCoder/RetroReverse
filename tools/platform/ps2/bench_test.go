package ps2

// bench_test.go is the gate the coming rasteriser parallelisation reports against — the thing
// that stops a "3x faster" claim from quietly being a "3x faster and subtly wrong" one. The
// PS2 oracle is ground truth for the exported assets, so a change that moves one pixel has not
// sped the machine up, it has broken it. Every claim is made twice: once in milliseconds
// (BenchmarkField), once in hashes (TestFrameHashes).
//
// WHAT IS HASHED, AND WHY. Unlike the GameCube, the PS2 does not erase what it drew as it
// delivers the frame — the title composites straight into the buffer DISPFB scans out — so
// the picture is durable and the choice is simpler:
//
//	ram    all 32 MiB of main memory. The GS's local memory is separate (below), but RAM
//	       carries every CPU-side and DMA-side trajectory, so it is the pin that catches a
//	       change nothing draws.
//	vram   the GS's 4 MiB of local memory: the frame buffers, the Z buffer, the textures and
//	       CLUTs. THIS is where a rasteriser change shows up first — a mis-parallelised fill
//	       lands here and nowhere else.
//	cpu    the EE register file and its clocks: the trajectory no picture can see. Both halves
//	       of every 128-bit GPR, the two MAC units, COP0, the COP1 float registers.
//
// The pins are a gate against CHANGE, not a proof of correctness — the frame they defend was
// judged by eye first (the Jak title logo, Daxter on it), and a mismatch writes the frame out
// so the next person judges it again rather than trusting the hash ([[pinned-hash-guards-change]]).
// They are machine-specific: Go permits FMA contraction, so a float result can differ between
// hosts. These were pinned on darwin/arm64; a mismatch elsewhere must be explained, not re-pinned.

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"retroreverse.com/tools/lib/iso9660"
)

// The recipe. Neither the disc image (copyright) nor the savestates (work/ is ignored) are
// committed, so everything here skips when they are absent.
const (
	jakImage = "../../../games/jak-and-daxter-ps2/image/Jak and Daxter - The Precursor Legacy.iso"

	// The title-logo field: the Jak and Daxter logo with Daxter on it. It is the field the
	// adapter's own capture counts — ~5,430 GS primitives — so the numbers this harness
	// reports are comparable to the ones the frame debugger shows.
	stateTitle = "../../../games/jak-and-daxter-ps2/work/states/title-logo.state"

	// benchFields is what the pins below were taken over. It is part of the recipe: change it
	// and every hash changes with it.
	benchFields = 2

	// benchBudget caps a field that never ends, so a hang is a failed gate rather than a
	// wedged test run. A field is ~1,000,000 EE instructions (stepsPerVBlank); this is generous.
	benchBudget = 60_000_000
)

// gatePin is one state's pinned output after benchFields fields.
type gatePin struct {
	name           string
	state          string
	ram, vram, cpu string
}

// The pins, taken on darwin/arm64 from the frame this recipe reproduces — the Jak title logo,
// which was LOOKED AT before these were written down. If one moves, open the PNG the failure
// prints and JUDGE THE PICTURE AGAIN before touching this block.
var gatePins = []gatePin{
	{
		name:  "title-logo",
		state: stateTitle,
		ram:   "613fbfc1882ec42dc2216193d3904275fd75030c90bb1ce164d1ad650e5382f9",
		vram:  "d9971679de816c66b00752f1d7a7e8e81fc4e4152e48f4db42573dfe07afd965",
		cpu:   "687a1ccb65c391d75967f14f499bbfe28e806f8e9a4b5cfc236eb18696ddd2a6",
	},
}

// atState boots the disc the way the adapter does and restores one of the recipe's savestates.
func atState(tb testing.TB, image, state string) *Machine {
	tb.Helper()
	raw, err := os.ReadFile(image)
	if err != nil {
		tb.Skipf("disc image not present (game images are not committed): %v", err)
	}
	vol, err := iso9660.OpenBytes(raw)
	if err != nil {
		tb.Fatalf("opening the disc: %v", err)
	}
	exePath, err := bootExe(vol)
	if err != nil {
		tb.Fatal(err)
	}
	elfRaw, err := vol.ReadFile(exePath)
	if err != nil {
		tb.Fatalf("reading boot executable %s: %v", exePath, err)
	}
	exe, err := LoadELF(elfRaw)
	if err != nil {
		tb.Fatal(err)
	}
	m := NewMachine()
	m.SetImageHash(fmt.Sprintf("%x", md5.Sum(raw)))
	m.SetVolume(vol)
	m.LoadExecutable(exe)
	if err := m.RebootIOP(); err != nil {
		tb.Fatalf("bringing the IOP up: %v", err)
	}
	if err := m.LoadStateFile(state); err != nil {
		tb.Skipf("no savestate at %s (work/ is not committed): %v", filepath.Base(state), err)
	}
	return m
}

// bootExe reads SYSTEM.CNF's BOOT2 line — the disc's own name for the executable to run.
func bootExe(vol *iso9660.Volume) (string, error) {
	cnf, err := vol.ReadFile("SYSTEM.CNF")
	if err != nil {
		return "", fmt.Errorf("reading SYSTEM.CNF: %w", err)
	}
	for _, line := range strings.Split(string(cnf), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "BOOT2"); ok {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "=")), nil
		}
	}
	return "", fmt.Errorf("SYSTEM.CNF names no BOOT2 executable")
}

// runFields advances the machine exactly n vertical blanks — the field boundary the profiler
// and the debugger both stop on. It borrows OnVBlank the way the adapter's StepFast does.
func runFields(m *Machine, n int, budget uint64) {
	got := 0
	m.OnVBlank = func(mm *Machine) {
		got++
		if got >= n {
			mm.StopRequested = true
		}
	}
	m.Run(budget)
	m.OnVBlank = nil
}

// gateHashes fingerprints the machine's memory and its EE trajectory.
func gateHashes(m *Machine) (ram, vram, cpu string) {
	sum := func(f func(h *hasher)) string {
		h := &hasher{Hash: sha256.New()}
		f(h)
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	ram = sum(func(h *hasher) { h.Write(m.ram) })
	vram = sum(func(h *hasher) {
		if m.gs != nil {
			h.Write(m.gs.vram)
		} else {
			h.Write([]byte("no gs"))
		}
	})
	cpu = sum(func(h *hasher) {
		c := m.CPU
		for _, r := range c.R {
			h.u64(r.Lo)
			h.u64(r.Hi)
		}
		for _, v := range []uint64{c.HI, c.LO, c.HI1, c.LO1, c.PC} {
			h.u64(v)
		}
		h.u64(uint64(c.SA))
		for _, v := range c.COP0 {
			h.u64(v)
		}
		// The COP1 float registers as bit patterns: two NaNs that compare unequal must still
		// fingerprint the same, and a geometry-affecting float lives here.
		for _, v := range c.FPR {
			h.u64(uint64(v))
		}
		h.u64(uint64(c.ACC))
		h.u64(uint64(c.FCR31))
		h.u64(c.Steps)
		// The machine's own frame clock: a run that drifted by a field would show nowhere else.
		h.u64(m.steps)
		h.u64(uint64(m.vblanks))
	})
	return ram, vram, cpu
}

type hasher struct{ hash.Hash }

func (h *hasher) u64(v uint64) {
	var w [8]byte
	binary.LittleEndian.PutUint64(w[:], v)
	h.Write(w[:])
}

// gateCounters snapshots the GS's per-run tallies, so a benchmark can report the work done
// alongside the time taken. FRAGMENTS PER MILLISECOND is the number that says whether the
// rasteriser got faster; wall time alone cannot tell that from a field that drew less.
type gateCounters struct {
	prims, frags, zRej int
}

func snapGate(m *Machine) gateCounters {
	if m.gs == nil {
		return gateCounters{}
	}
	return gateCounters{prims: m.gs.prims, frags: int(m.gs.plotted), zRej: int(m.gs.rejZ)}
}

func (c gateCounters) sub(o gateCounters) gateCounters {
	return gateCounters{c.prims - o.prims, c.frags - o.frags, c.zRej - o.zRej}
}

// TestFrameHashes is the gate. It runs the recipe and compares the machine's memory against
// the pins. On a mismatch it writes the frame it actually produced to a PNG and names the
// path, because a hash says only "different".
func TestFrameHashes(t *testing.T) {
	if testing.Short() {
		t.Skip("the frame gate runs the GS for several fields; -short skips it")
	}
	if raceBuild {
		t.Skip("-race changes the floating-point result (FMA contraction); the pins are for the ordinary build")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			m := atState(t, jakImage, p.state)

			before := snapGate(m)
			start := time.Now()
			runFields(m, benchFields, benchBudget)
			elapsed := time.Since(start)
			did := snapGate(m).sub(before)

			t.Logf("%d fields in %s (%s/field) — %d gs primitives, %d fragments drawn, %d depth-rejected",
				benchFields, elapsed.Round(time.Millisecond),
				(elapsed / benchFields).Round(time.Millisecond), did.prims, did.frags, did.zRej)
			if ms := elapsed.Seconds() * 1000; ms > 0 {
				t.Logf("  %.0f fragments/ms", float64(did.frags)/ms)
			}

			ram, vram, cpu := gateHashes(m)
			for _, c := range []struct{ name, got, want string }{
				{"ram", ram, p.ram},
				{"vram", vram, p.vram},
				{"cpu", cpu, p.cpu},
			} {
				if c.want == "PIN_ME" {
					t.Errorf("%s hash is unpinned; it is %s", c.name, c.got)
					continue
				}
				if c.got != c.want {
					t.Errorf("%s hash = %s, want %s", c.name, c.got, c.want)
				}
			}
			if t.Failed() {
				writeGateFrame(t, m, filepath.Join(os.TempDir(), "ps2-gate-fail-"+p.name+".png"))
			}
		})
	}
}

// writeGateFrame puts the picture the machine actually produced on disk and says so loudly.
func writeGateFrame(t *testing.T, m *Machine, path string) {
	t.Helper()
	pix, w, h := m.GSFrame()
	if pix == nil || w == 0 || h == 0 {
		t.Logf("(no displayable frame to write out)")
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(path)
	if err != nil {
		t.Logf("(could not write the frame out: %v)", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Logf("(could not encode the frame out: %v)", err)
		return
	}
	t.Logf("LOOK AT THE FRAME before re-pinning: %s", path)
}

// TestFrameDeterminism checks the property the gate rests on: the same state, run the same
// way, twice, produces the same machine. If this fails, no pinned hash means anything.
func TestFrameDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for several fields twice; -short skips it")
	}
	var first [3]string
	for run := 0; run < 2; run++ {
		m := atState(t, jakImage, stateTitle)
		runFields(m, benchFields, benchBudget)
		ram, vram, cpu := gateHashes(m)
		got := [3]string{ram, vram, cpu}
		if run == 0 {
			first = got
			continue
		}
		for i, name := range []string{"ram", "vram", "cpu"} {
			if got[i] != first[i] {
				t.Errorf("%s differs between two identical runs: %s vs %s", name, got[i], first[i])
			}
		}
	}
}

// BenchmarkField reports ns/field and — because a field is only comparable to another field
// that drew the same thing — the work it did alongside. It re-restores the savestate for each
// measured batch so an animating title does not drift into a different scene.
func BenchmarkField(b *testing.B) {
	m := atState(b, jakImage, stateTitle)
	snap := m.SaveState()

	var did gateCounters
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := m.LoadState(snap); err != nil {
			b.Fatal(err)
		}
		before := snapGate(m)
		b.StartTimer()

		runFields(m, 1, benchBudget)

		b.StopTimer()
		did = snapGate(m).sub(before)
		b.StartTimer()
	}
	b.StopTimer()

	b.ReportMetric(float64(did.prims), "prims/field")
	b.ReportMetric(float64(did.frags), "frags/field")
	b.ReportMetric(float64(did.zRej), "zrej/field")
}
