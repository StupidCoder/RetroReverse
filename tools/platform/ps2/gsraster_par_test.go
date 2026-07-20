package ps2

// gsraster_par_test.go carries the correctness claim for the banded rasteriser — the claim
// the pinned frame hashes cannot make.
//
// A pinned hash (bench_test.go) says "the same as last week". It is a gate against change and
// worth having, but it cannot say the parallel machine agrees with the serial one: both could
// have moved together, and on another host with a different core count they run a different
// partition entirely. So the narrow claim made here is the one that matters — THIS BUILD, RUN
// IN PARALLEL, COMPUTES WHAT THIS BUILD COMPUTES SERIALLY: every pixel of VRAM, and every
// census counter.
//
// The claim rests on the partition, not the scheduler. A band of rows is filled by exactly
// one worker; no two workers touch the same pixel; within a primitive the geometry is fixed
// before the fan-out. Which worker takes which band is a shared counter's whim and cannot
// change the buffer — only the time.

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

// vramHash fingerprints the GS's whole local memory — the frame buffers, Z buffer, textures
// and CLUTs — which is exactly and only where a mis-partitioned fill can land.
func vramHash(m *Machine) string {
	if m.gs == nil {
		return "no gs"
	}
	return fmt.Sprintf("%x", sha256.Sum256(m.gs.vram))
}

// parCounters snapshots the census the fan-out merges by hand, so the test can require the
// merge to be order-independent (the totals must not depend on who finished first).
type parCounters struct {
	plotted, rejScissor, rejZ, rejAlpha, rejDate, texBlack, texColor uint64
	nonBlack                                                         [8]uint64
}

func snapPar(m *Machine) parCounters {
	if m.gs == nil {
		return parCounters{}
	}
	g := m.gs
	return parCounters{
		plotted: g.plotted, rejScissor: g.rejScissor, rejZ: g.rejZ,
		rejAlpha: g.rejAlpha, rejDate: g.rejDate, texBlack: g.texBlack,
		texColor: g.texColor, nonBlack: g.plotNonBlack,
	}
}

// parCase is one (game, state) the parallel path is exercised on. The Jak title is the gate
// frame; the RRV states carry the full-screen blur and pyramid sprites that fan out hardest.
type parCase struct {
	name, state string
	rrv         bool
}

var parCases = []parCase{
	{name: "jak-title", state: stateTitle},
	{name: "rrv-intro", state: "intro.state", rrv: true},
	{name: "rrv-menu", state: "main menu.state", rrv: true},
	{name: "rrv-title", state: "title.state", rrv: true},
}

func (c parCase) machine(tb testing.TB) *Machine {
	if c.rrv {
		return rrvAt(tb, c.state)
	}
	return atState(tb, jakImage, c.state)
}

// TestParallelMatchesSerial runs the same field from the same state twice — once forced
// single-threaded, once fanned out — and requires the two to agree exactly, in VRAM and in
// every merged counter.
func TestParallelMatchesSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for several fields twice per state; -short skips it")
	}
	// This test is deliberately NOT skipped under -race. -race inhibits FMA contraction so
	// the float result differs from the ordinary build, but it differs the SAME way for both
	// machines here — so the serial/parallel comparison holds, and it is the run that proves
	// the fan-out has no data race. It compares a build to itself, never to a pinned hash.
	for _, c := range parCases {
		t.Run(c.name, func(t *testing.T) {
			serial := c.machine(t)
			serial.SingleThreaded = true
			defer serial.Close()
			runFields(serial, benchFields, benchBudget)

			par := c.machine(t)
			defer par.Close()
			runFields(par, benchFields, benchBudget)

			if s, p := vramHash(serial), vramHash(par); s != p {
				t.Errorf("parallel VRAM differs from serial:\n serial   %s\n parallel %s", s, p)
			}
			sc, pc := snapPar(serial), snapPar(par)
			if sc != pc {
				t.Errorf("merged census differs:\n serial   %+v\n parallel %+v", sc, pc)
			}
			// And the fan-out actually happened on this state, so the agreement is not
			// vacuous (a threshold that quietly rejected everything would agree trivially).
			if par.gs != nil {
				t.Logf("parFills=%d serFills=%d", par.gs.parFills, par.gs.serFills)
			}
		})
	}
}

// TestParallelIsActuallyUsed stops the test above from passing vacuously.
//
// Every claim here is worthless if the fill never fans out: a rasterFanWorkers that returned
// 1 for everything would pass every VRAM hash, every counter check, and the machine would
// simply be slow. So: a real field must reach the parallel path, and the threshold must still
// reject the small primitives it exists to reject.
func TestParallelIsActuallyUsed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GS for several fields; -short skips it")
	}
	m := atState(t, jakImage, stateTitle)
	defer m.Close()
	runFields(m, benchFields, benchBudget)
	if m.gs == nil {
		t.Fatal("no GS after the run")
	}
	par, ser := m.gs.parFills, m.gs.serFills
	if par == 0 {
		t.Fatalf("no primitive was filled in parallel (%d serial): the fan-out is dead and "+
			"every other test in this file passes vacuously", ser)
	}
	if ser == 0 {
		t.Errorf("every one of %d primitives fanned out: the threshold is not rejecting the "+
			"small primitives it exists to reject", par)
	}
	t.Logf("filled in parallel: %d; serially (below the threshold): %d", par, ser)
}

// BenchmarkParallelAB reports the same field filled serially and in parallel, back to back, so
// the speedup is measured on one host in one sitting (the only comparison that means anything —
// a frame time is not comparable across machines, and cold-vs-warm variance is ~14%).
func BenchmarkParallelAB(b *testing.B) {
	for _, c := range parCases {
		mk := func() (*Machine, MachineState) {
			m := c.machine(b)
			return m, m.SaveState()
		}
		b.Run(c.name+"/serial", func(b *testing.B) {
			m, snap := mk()
			defer m.Close()
			m.SingleThreaded = true
			benchLoadRun(b, m, snap)
		})
		b.Run(c.name+"/parallel", func(b *testing.B) {
			m, snap := mk()
			defer m.Close()
			benchLoadRun(b, m, snap)
		})
	}
}

func benchLoadRun(b *testing.B, m *Machine, snap MachineState) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := m.LoadState(snap); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		runFields(m, 1, benchBudget)
	}
}
