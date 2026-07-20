package xbox

import (
	"fmt"
	"os"
	"testing"
	"time"

	"retroreverse.com/tools/cpu/x86"
)

// A drawn savestate: the title screen, already past the boot the debugger is meant to skip.
// It is scratch (work/ is regenerable), so the state tests skip when it is absent.
const profStatePath = "../../../games/outrun-2006-xbox/work/title.state"

// profiledMachine loads the drawn savestate into a fresh machine with the profiler armed and
// runs one presented frame (stopping at the first flip). It skips when the disc or the state
// is not present.
func profiledMachine(t *testing.T) *Machine {
	t.Helper()
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the OutRun disc image is not present")
	}
	if _, err := os.Stat(profStatePath); err != nil {
		t.Skip("the title savestate is not present (work/ is regenerable scratch)")
	}
	m := bootMachine(t)
	m.EnableGPU()
	if err := m.LoadStateFile(profStatePath); err != nil {
		t.Fatalf("loadstate: %v", err)
	}
	m.ClearHalt()
	m.SetProfile(true)
	m.OnFlip = func(mm *Machine) { mm.StopRequested = true }
	m.Run(400_000_000)
	return m
}

// TestProfileMeasuresAPresentedFrame is the end-to-end platform check: a profiled frame
// reports disjoint buckets that sum to the total, names the rasteriser as the cost centre,
// and carries the work counters wall time alone cannot interpret.
func TestProfileMeasuresAPresentedFrame(t *testing.T) {
	p := profiledMachine(t).FrameProfile()
	if p.TotalMs == 0 || len(p.Buckets) == 0 {
		t.Fatalf("no frame profiled: %+v", p)
	}

	// Buckets are disjoint leaves plus two derived remainders, so they must sum to the total.
	var sum float64
	for _, b := range p.Buckets {
		sum += b.Millis
	}
	if d := sum - p.TotalMs; d > 0.5 || d < -0.5 {
		t.Errorf("buckets sum to %.3f ms but total is %.3f ms (Δ %.3f) — not disjoint\n%s",
			sum, p.TotalMs, d, formatProfile(p))
	}

	// The rasteriser is this machine's cost centre (14M fragments/frame); a profile that does
	// not say so is measuring the wrong thing.
	if bucketMs(p, "rasterise") <= 0 {
		t.Errorf("rasterise bucket is empty on a drawing frame:\n%s", formatProfile(p))
	}
	if !p.Drew {
		t.Errorf("a frame with draws reported Drew=false:\n%s", formatProfile(p))
	}
	// Counters carry the work; ms alone cannot tell a faster rasteriser from a smaller frame.
	if counterVal(p, "fragments drawn") <= 0 || counterVal(p, "draws") <= 0 {
		t.Errorf("the profile carries no work counters:\n%s", formatProfile(p))
	}
	t.Logf("profiled frame:\n%s", formatProfile(p))
}

// TestProfiledFragmentCountsMatchAcrossParallelAndSerial pins the parallel counter design:
// the per-lane rstats summed after the join must equal the serial path's tallies exactly.
// Run under -race it is also the race check — the fragment counters are the one thing timed
// on the worker goroutines, and a shared increment would both race and miscount.
func TestProfiledFragmentCountsMatchAcrossParallelAndSerial(t *testing.T) {
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the OutRun disc image is not present")
	}
	if _, err := os.Stat(profStatePath); err != nil {
		t.Skip("the title savestate is not present (work/ is regenerable scratch)")
	}
	run := func(serial bool) (w, z, a int) {
		save := rasterSerial
		rasterSerial = serial
		defer func() { rasterSerial = save }()
		m := bootMachine(t)
		m.EnableGPU()
		if err := m.LoadStateFile(profStatePath); err != nil {
			t.Fatalf("loadstate: %v", err)
		}
		m.ClearHalt()
		m.OnFlip = func(mm *Machine) { mm.StopRequested = true }
		m.Run(400_000_000)
		g := m.pgraph
		return g.pixWritten, g.pixZRej, g.pixARej
	}
	pw, pz, pa := run(false) // parallel
	sw, sz, sa := run(true)  // serial
	if pw != sw || pz != sz || pa != sa {
		t.Fatalf("parallel counters (drawn=%d z=%d a=%d) differ from serial (drawn=%d z=%d a=%d)",
			pw, pz, pa, sw, sz, sa)
	}
	t.Logf("fragment counts agree across paths: drawn=%d depth-killed=%d alpha-killed=%d", pw, pz, pa)
}

// newProfMachine builds the bare minimum a profiler test needs — a pgraph to read counters
// from and a CPU whose Steps the profiler baselines — without a disc, so the timer test runs
// on any checkout.
func newProfMachine() *Machine {
	m := &Machine{}
	m.pgraph = newPgraph(m)
	m.CPU = x86.NewCPU(m)
	return m
}

// TestProfFrameClosesInFlightTimers is the guard the GameCube memory demands: the sum test
// above CANNOT see a cross-boundary misattribution, because a mis-split still sums to the
// total. This drives the mechanism directly. profFrame fires mid-drain with the run and
// pusher timers in flight; it must close BOTH into the ending frame and restart them for the
// next. It is mutation-tested both ways — drop the close in profFrame and frame 1 loses its
// in-flight pusher drain (count 0); drop the restart and frame 2 double-counts frame 1's.
func TestProfFrameClosesInFlightTimers(t *testing.T) {
	m := newProfMachine()
	m.SetProfile(true)

	// Frame 1: a run and a pusher drain are in flight when the flip fires.
	m.profRunEnter()
	m.profPusherEnter()
	time.Sleep(5 * time.Millisecond) // pusher work before the flip
	m.profFrame()                    // the boundary, mid-drain
	p1 := m.FrameProfile()

	// The close must have booked the in-flight drain into frame 1 (as command decode — nothing
	// is nested) and counted it. A dropped close leaves the count at zero.
	if c := bucketCount(p1, "command decode (derived)"); c != 1 {
		t.Fatalf("frame 1 did not close the in-flight pusher drain (close dropped): count=%d\n%s", c, formatProfile(p1))
	}
	if d := bucketMs(p1, "command decode (derived)"); d < 3 {
		t.Fatalf("frame 1 lost the in-flight pusher time (close dropped): %.2f ms\n%s", d, formatProfile(p1))
	}

	// Frame 2: the SAME drain continues after the flip, then ends, then the next flip closes.
	time.Sleep(5 * time.Millisecond) // pusher work after the flip, in the restarted timer
	m.profPusherExit()
	m.profRunExit()
	m.profFrame()
	p2 := m.FrameProfile()

	d2 := bucketMs(p2, "command decode (derived)")
	if d2 < 3 {
		t.Fatalf("frame 2 lost the restarted pusher tail (restart dropped → timer never rearmed): %.2f ms\n%s", d2, formatProfile(p2))
	}
	if d2 > 8 {
		t.Fatalf("frame 2 also counted frame 1's pusher time (restart dropped → stale start): %.2f ms\n%s", d2, formatProfile(p2))
	}
}

// TestProfileOffIsFree confirms the promise in the file header: with profiling off, no frame
// is ever recorded and the accumulators stay empty, so a run pays only the branch.
func TestProfileOffIsFree(t *testing.T) {
	m := newProfMachine()
	if p := m.FrameProfile(); len(p.Buckets) != 0 || p.TotalMs != 0 {
		t.Fatalf("a machine that never profiled reported a frame: %+v", p)
	}
	// Drive the brackets with the profiler off: nothing must accumulate.
	m.profRunEnter()
	m.profPusherEnter()
	m.profEnd(bucketRaster, m.profStart())
	m.profPusherExit()
	m.profRunExit()
	m.profFrame()
	if p := m.FrameProfile(); len(p.Buckets) != 0 {
		t.Fatalf("profiling off still produced a frame: %+v", p)
	}
}

func bucketMs(p FrameProfile, name string) float64 {
	for _, b := range p.Buckets {
		if b.Name == name {
			return b.Millis
		}
	}
	return -1
}

func bucketCount(p FrameProfile, name string) int {
	for _, b := range p.Buckets {
		if b.Name == name {
			return b.Count
		}
	}
	return -1
}

func counterVal(p FrameProfile, name string) int {
	for _, c := range p.Counters {
		if c.Name == name {
			return c.Value
		}
	}
	return -1
}

func formatProfile(p FrameProfile) string {
	s := fmt.Sprintf("total %.2f ms (drew=%v)\n", p.TotalMs, p.Drew)
	for _, b := range p.Buckets {
		s += fmt.Sprintf("  %-26s %8.2f ms  x%d\n", b.Name, b.Millis, b.Count)
	}
	for _, c := range p.Counters {
		s += fmt.Sprintf("  %-26s %d\n", c.Name, c.Value)
	}
	return s
}
