package xbox

// bench_test.go is the gate the coming NV2A optimisations report against — the thing that stops
// a "2x faster" claim from quietly being a "2x faster and subtly wrong" one. The Xbox oracle is
// ground truth (the render surface is what other code is verified against), so a change that
// moves one pixel has not sped the machine up, it has broken it. Every claim is made twice: once
// in milliseconds (BenchmarkField), once in hashes (TestFrameHashes).
//
// WHAT IS HASHED, AND WHY. The Xbox's frame boundary is the FLIP, and the gate stops AT the flip
// (OnFlip → StopRequested) so the completed frame is still the bound colour surface — exactly the
// window -stopflip/-surfpng use to photograph a whole presented frame rather than a mid-frame
// slice. Three fingerprints:
//
//	ram      all of physical RAM: every CPU-side and DMA-side trajectory, so it is the pin that
//	         catches a change nothing draws.
//	surface  the resolved render surface (RenderDrawTarget: de-swizzled, AA-resolved) — the
//	         picture. A mis-parallelised fill or a broken combiner lands here and nowhere else.
//	cpu      the x86 register file and its clocks: the trajectory no picture can see. The GPRs,
//	         the segment selectors, the flags, the x87 stack, the SSE/MMX registers as raw bits
//	         (two NaNs that compare unequal must still fingerprint the same), and the step count.
//
// The pins are a gate against CHANGE, not a proof of correctness — the frame they defend was
// judged by eye first (an3-drive, the in-race driving view), and a mismatch writes the frame out
// so the next person judges it again rather than trusting the hash ([[pinned-hash-guards-change]]).
// They are machine-specific: Go permits FMA contraction, so a float result can differ between
// hosts. These were pinned on darwin/arm64; a mismatch elsewhere must be explained, not re-pinned.

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The recipe. Neither the disc image (copyright) nor the savestates (work/ is ignored) are
// committed, so everything here skips when they are absent.
const (
	outrunImage = "../../../games/outrun-2006-xbox/image/OutRun 2006 - Coast 2 Coast (EUR).iso"
	outrunXBE   = "/default.xbe"

	// The driving field: the in-race view an3-drive.state restores into — 523 draws, 1.76M
	// fragments. It is the field the performance analysis profiled, so the numbers this harness
	// reports are comparable to that write-up (PERFORMANCE.md, "The Xbox").
	stateDrive = "../../../games/outrun-2006-xbox/work/states/an3-drive.state"

	// benchFields is what the pins below were taken over. It is part of the recipe: change it
	// and every hash changes with it.
	benchFields = 2

	// benchBudget caps a field that never ends, so a hang is a failed gate rather than a wedged
	// test run. A driving field is ~1.6M x86 instructions; this is generous.
	benchBudget = 300_000_000
)

// gatePin is one state's pinned output after benchFields fields.
type gatePin struct {
	name           string
	state          string
	ram, surf, cpu string
}

// The pins, taken on darwin/arm64 from the frame this recipe reproduces — an3-drive, which was
// LOOKED AT before these were written down. If one moves, open the PNG the failure prints and
// JUDGE THE PICTURE AGAIN before touching this block.
var gatePins = []gatePin{
	{
		// Re-pinned under FlipVSync (on by default): the vsync-accurate clock re-times the
		// two fields this runs, so RAM/surface/CPU moved from the pre-FlipVSync values even
		// though the picture — an3-drive at the start line, 000 km/h — is the same to the eye.
		name:  "an3-drive",
		state: stateDrive,
		ram:   "44a1e8153f58e6319535255c251e6ab684af3622d9e7016dd34cb0aaef7fcc1a",
		surf:  "b16da58bf2088ae17893c1636b2850e9361695c36f8fc4b3467b071f0817af9d",
		cpu:   "3b9e608fd2272e81470b4694822f401800c6d02737bc437566ce483d58261ec9",
	},
}

// atState boots the disc the way the bootoracle does and restores one of the recipe's savestates.
func atState(tb testing.TB, image, xbePath, state string) (*Machine, *Image) {
	tb.Helper()
	disc, err := Open(image)
	if err != nil {
		tb.Skipf("disc image not present (game images are not committed): %v", err)
	}
	xbeBytes, err := disc.ReadFile(xbePath)
	if err != nil {
		disc.Close()
		tb.Fatalf("reading %s: %v", xbePath, err)
	}
	xbe, err := ParseXBE(xbeBytes)
	if err != nil {
		disc.Close()
		tb.Fatalf("parsing XBE: %v", err)
	}
	m, err := NewMachine(xbe, disc)
	if err != nil {
		disc.Close()
		tb.Fatalf("building machine: %v", err)
	}
	m.EnableGPU()
	if err := m.LoadStateFile(state); err != nil {
		disc.Close()
		tb.Skipf("no savestate at %s (work/ is not committed): %v", filepath.Base(state), err)
	}
	if m.CPU.Halted {
		// A frontier state stopped on an unimplemented ordinal, whose trap is retried on resume.
		m.ClearHalt()
	}
	return m, disc
}

// runFields advances the machine exactly n flips — the frame boundary the profiler and the
// debugger both stop on — and stops AT the nth flip so the completed frame is still the bound
// colour surface (the -stopflip window).
func runFields(m *Machine, n int, budget uint64) {
	got := 0
	prev := m.OnFlip
	m.OnFlip = func(mm *Machine) {
		if prev != nil {
			prev(mm)
		}
		got++
		if got >= n {
			mm.StopRequested = true
		}
	}
	m.Run(budget)
	m.OnFlip = prev
}

// gateHashes fingerprints the machine's memory, its presented surface, and its x86 trajectory.
func gateHashes(m *Machine) (ram, surf, cpu string) {
	sum := func(f func(h *hasher)) string {
		h := &hasher{Hash: sha256.New()}
		f(h)
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	ram = sum(func(h *hasher) { h.Write(m.RAM) })
	surf = sum(func(h *hasher) {
		img, err := m.RenderDrawTarget()
		if err != nil || img == nil {
			h.Write([]byte("no surface: " + fmt.Sprint(err)))
			return
		}
		h.Write(img.Pix)
	})
	cpu = sum(func(h *hasher) {
		c := m.CPU
		for _, r := range c.Regs {
			h.u32(r)
		}
		for _, s := range c.Seg {
			h.u32(uint32(s))
		}
		for _, b := range c.SegBase {
			h.u32(b)
		}
		h.u32(c.IP)
		// The flags, one bit each, packed into a word.
		flags := uint32(0)
		for i, f := range []bool{c.CF, c.PF, c.AF, c.ZF, c.SF, c.TF, c.IF, c.DF, c.OF} {
			if f {
				flags |= 1 << uint(i)
			}
		}
		h.u32(flags)
		// The x87 stack as bit patterns, plus its tags and control/status.
		for _, st := range c.FPU.St {
			h.u64(math.Float64bits(st))
		}
		h.Write(c.FPU.Tag[:])
		h.u32(uint32(c.FPU.Top))
		h.u32(uint32(c.FPU.Ctrl))
		h.u32(uint32(c.FPU.Stat))
		// The SSE and MMX registers as raw bytes, and the SSE control word.
		for i := range c.XMM {
			h.Write(c.XMM[i][:])
		}
		for i := range c.MMX {
			h.Write(c.MMX[i][:])
		}
		h.u32(c.MXCSR)
		// The trajectory clock: a run that drifted by a field shows nowhere else.
		h.u64(c.Steps)
	})
	return ram, surf, cpu
}

type hasher struct{ hash.Hash }

func (h *hasher) u32(v uint32) {
	var w [4]byte
	binary.LittleEndian.PutUint32(w[:], v)
	h.Write(w[:])
}

func (h *hasher) u64(v uint64) {
	var w [8]byte
	binary.LittleEndian.PutUint64(w[:], v)
	h.Write(w[:])
}

// gateCounters snapshots the pgraph's per-run tallies, so a benchmark can report the work done
// alongside the time taken. FRAGMENTS PER MILLISECOND is the number that says whether the
// rasteriser got faster; wall time alone cannot tell that from a field that drew less.
type gateCounters struct {
	draws, frags, zRej int
}

func snapGate(m *Machine) gateCounters {
	g := m.pgraph
	if g == nil {
		return gateCounters{}
	}
	return gateCounters{draws: g.Draws, frags: g.pixWritten, zRej: g.pixZRej}
}

func (c gateCounters) sub(o gateCounters) gateCounters {
	return gateCounters{c.draws - o.draws, c.frags - o.frags, c.zRej - o.zRej}
}

// TestFrameHashes is the gate. It runs the recipe and compares the machine's memory against the
// pins. On a mismatch it writes the frame it actually produced to a PNG and names the path,
// because a hash says only "different".
func TestFrameHashes(t *testing.T) {
	if testing.Short() {
		t.Skip("the frame gate runs the NV2A for several fields; -short skips it")
	}
	if raceBuild {
		t.Skip("-race changes the floating-point result (FMA contraction); the pins are for the ordinary build")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			m, disc := atState(t, outrunImage, outrunXBE, p.state)
			defer disc.Close()

			before := snapGate(m)
			start := time.Now()
			runFields(m, benchFields, benchBudget)
			elapsed := time.Since(start)
			did := snapGate(m).sub(before)

			t.Logf("%d fields in %s (%s/field) — %d draws, %d fragments drawn, %d depth-rejected",
				benchFields, elapsed.Round(time.Millisecond),
				(elapsed / benchFields).Round(time.Millisecond), did.draws, did.frags, did.zRej)
			if ms := elapsed.Seconds() * 1000; ms > 0 {
				t.Logf("  %.0f fragments/ms", float64(did.frags)/ms)
			}

			ram, surf, cpu := gateHashes(m)
			for _, c := range []struct{ name, got, want string }{
				{"ram", ram, p.ram},
				{"surface", surf, p.surf},
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
				writeGateFrame(t, m, filepath.Join(os.TempDir(), "xbox-gate-fail-"+p.name+".png"))
			}
		})
	}
}

// writeGateFrame puts the picture the machine actually produced on disk and says so loudly.
func writeGateFrame(t *testing.T, m *Machine, path string) {
	t.Helper()
	img, err := m.RenderDrawTarget()
	if err != nil || img == nil {
		t.Logf("(no displayable frame to write out: %v)", err)
		return
	}
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

// TestFrameDeterminism checks the property the gate rests on: the same state, run the same way,
// twice, produces the same machine. If this fails, no pinned hash means anything.
func TestFrameDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the NV2A for several fields twice; -short skips it")
	}
	var first [3]string
	for run := 0; run < 2; run++ {
		m, disc := atState(t, outrunImage, outrunXBE, stateDrive)
		runFields(m, benchFields, benchBudget)
		ram, surf, cpu := gateHashes(m)
		disc.Close()
		got := [3]string{ram, surf, cpu}
		if run == 0 {
			first = got
			continue
		}
		for i, name := range []string{"ram", "surface", "cpu"} {
			if got[i] != first[i] {
				t.Errorf("%s differs between two identical runs: %s vs %s", name, got[i], first[i])
			}
		}
	}
}

// warmFields is how many consecutive fields BenchmarkWarmFields runs per op: one cold field that
// fills the texture cache plus (warmFields-1) warm ones. The steady state a real run spends its
// life in is the warm field, and some optimisations (PERFORMANCE.md #1, the cross-field texture
// cache) only help fields 2+ — a one-field benchmark is blind to them. Kept small so the
// animating scene does not drift into a different workload.
const warmFields = 8

// BenchmarkWarmFields reports ms/field over a warm run — the number the cross-field texture cache
// (item #1) is meant to move. It reloads the savestate each op so the run always starts from the
// same cold cache and the same scene.
func BenchmarkWarmFields(b *testing.B) {
	m, disc := atState(b, outrunImage, outrunXBE, stateDrive)
	defer disc.Close()
	snap := m.SaveState()

	var did gateCounters
	var total time.Duration
	for i := 0; i < b.N; i++ {
		if err := m.LoadState(snap); err != nil {
			b.Fatal(err)
		}
		if m.CPU.Halted {
			m.ClearHalt()
		}
		before := snapGate(m)
		start := time.Now()
		runFields(m, warmFields, benchBudget)
		total += time.Since(start)
		did = snapGate(m).sub(before)
	}
	b.ReportMetric(float64(total.Nanoseconds())/float64(b.N)/warmFields/1e6, "ms/field")
	b.ReportMetric(float64(did.draws)/warmFields, "draws/field")
	b.ReportMetric(float64(did.frags)/warmFields, "frags/field")
}

// BenchmarkField reports ns/field and — because a field is only comparable to another field that
// drew the same thing — the work it did alongside. It re-restores the savestate for each measured
// batch so an animating scene does not drift into a different frame.
func BenchmarkField(b *testing.B) {
	m, disc := atState(b, outrunImage, outrunXBE, stateDrive)
	defer disc.Close()
	snap := m.SaveState()

	var did gateCounters
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := m.LoadState(snap); err != nil {
			b.Fatal(err)
		}
		if m.CPU.Halted {
			m.ClearHalt()
		}
		before := snapGate(m)
		b.StartTimer()

		runFields(m, 1, benchBudget)

		b.StopTimer()
		did = snapGate(m).sub(before)
		b.StartTimer()
	}
	b.StopTimer()

	b.ReportMetric(float64(did.draws), "draws/field")
	b.ReportMetric(float64(did.frags), "frags/field")
	b.ReportMetric(float64(did.zRej), "zrej/field")
}
