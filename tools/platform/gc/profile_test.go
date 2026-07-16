package gc

import (
	"os"
	"testing"
	"time"
)

const cutscenePath = "../../../games/luigis-mansion-gc/work/states/intro-cutscene.state"

// profiledField runs one field of the intro cutscene with profiling on and returns its
// profile. It needs the disc and a savestate at a drawn frame — the boot to one is ~1.5
// billion instructions, far too long for a test.
func profiledField(t *testing.T) FrameProfile {
	t.Helper()
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the disc image is not present")
	}
	if _, err := os.Stat(cutscenePath); err != nil {
		t.Skip("the cutscene savestate is not present (work/ is regenerable scratch)")
	}
	d, err := Open(discPath)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	m, err := NewMachine(d)
	if err != nil {
		t.Fatal(err)
	}
	m.SetSpinDetect(false)
	if err := m.LoadStateFile(cutscenePath); err != nil {
		t.Fatal(err)
	}
	m.SetProfile(true)
	// Run to the second flip: the first closes a field that began before profiling was on,
	// so its buckets would be a fraction of its total.
	flips := 0
	m.OnFlip = func(mm *Machine) {
		flips++
		if flips >= 2 {
			mm.StopRequested = true
		}
	}
	m.Run(400_000_000)
	p := m.FrameProfile()
	if len(p.Buckets) == 0 {
		t.Fatal("no field was profiled")
	}
	return p
}

// TestProfileBucketsAreDisjointAndSumToTotal is the test the whole design turns on.
//
// The buckets are reported as a stacked bar, which is a claim: that they are disjoint leaves
// which add up to the field. If they overlap, the sum exceeds the total, and the derived
// remainder — the Gekko's share — gets clamped to zero to hide it. That clamp is why this
// has to be asserted rather than eyeballed: a profile with overlapping buckets still renders
// as a perfectly plausible bar chart, with the CPU's slice merely looking small.
//
// It is also the exact bug the PSP and the 3DO shipped, where work nested inside a frame's
// closing call was booked after the accumulators had been reset and landed in the next
// field's bucket. On this machine the copy that closes a field runs inside the FIFO burst
// that is being timed, so the same shape was one obvious refactor away.
func TestProfileBucketsAreDisjointAndSumToTotal(t *testing.T) {
	p := profiledField(t)

	var summed float64
	for _, b := range p.Buckets {
		if b.Millis < 0 {
			t.Errorf("bucket %q is negative (%.3f ms)", b.Name, b.Millis)
		}
		summed += b.Millis
	}
	if p.TotalMs <= 0 {
		t.Fatalf("the field's total is %.3f ms", p.TotalMs)
	}
	// The buckets are derived from the same clock as the total, so they should sum to it
	// almost exactly; a percent of slack covers the clock reads that happen between a
	// bucket closing and the total being taken.
	if diff := summed - p.TotalMs; diff > p.TotalMs*0.01 || diff < -p.TotalMs*0.01 {
		t.Errorf("the buckets sum to %.3f ms but the field is %.3f ms (off by %.3f):\n%s",
			summed, p.TotalMs, diff, formatProfile(p))
	}
}

// TestProfileRemainderIsNotClamped: the Gekko's derived share must be a real measurement,
// not a zero produced by clamping an overlap away.
//
// A remainder pinned at zero means the measured buckets have eaten the whole field, which on
// an interpreted machine cannot be true — the Gekko has to have run the code that issued the
// draws. So a zero here is a bug report about the nesting, and it is what gave the 3DO's
// version away.
func TestProfileRemainderIsNotClamped(t *testing.T) {
	p := profiledField(t)
	var rest *ProfileBucket
	for i := range p.Buckets {
		if p.Buckets[i].Name == "gekko + rest (derived)" {
			rest = &p.Buckets[i]
		}
	}
	if rest == nil {
		t.Fatal("no derived remainder bucket in the profile")
	}
	if rest.Millis <= 0 {
		t.Errorf("the Gekko's derived share is %.3f ms — clamped, so the measured buckets overlap:\n%s",
			rest.Millis, formatProfile(p))
	}
}

// TestProfileCountsTheFieldsWork: the counters must describe THIS field, not the run so far.
// They are lifetime tallies underneath, reported as deltas, so a base that failed to reset
// would show the whole cutscene's work in one field.
func TestProfileCountsTheFieldsWork(t *testing.T) {
	p := profiledField(t)
	got := map[string]int{}
	for _, c := range p.Counters {
		got[c.Name] = c.Value
	}
	if !p.Drew {
		t.Error("the cutscene field reports that it drew nothing")
	}
	// One field of the forest: ~12.7k commands, ~9.8k draws. An order of magnitude either
	// way means the delta is wrong, not that the scene changed.
	if d := got["draws"]; d < 1000 || d > 100_000 {
		t.Errorf("draws = %d, want a field's worth (~9,800)", d)
	}
	if c := got["gx commands"]; c < 1000 || c > 200_000 {
		t.Errorf("gx commands = %d, want a field's worth (~12,700)", c)
	}
	if f := got["fragments drawn"]; f <= 0 {
		t.Errorf("fragments drawn = %d in a field that drew", f)
	}
	if i := got["gekko instructions"]; i <= 0 {
		t.Errorf("gekko instructions = %d in a field that ran", i)
	}
}

// TestProfFrameClosesInFlightTimers pins the mechanism that makes cross-boundary attribution
// right, and it exists because the end-to-end sum test above CANNOT catch this.
//
// Why not: "command decode" is derived as fifo − (vertex + raster + copy). So work that is
// booked to the wrong side of the boundary is silently absorbed by the derived bucket, the
// buckets still add up to the field, and the bar chart still looks right — the copy's cost
// simply moves into "command decode" and stays there. (The 3DO's profiler gave its version of
// this bug away only because its buckets were independent leaves, so an overlap broke the
// sum. Deriving a bucket buys an honest measurement of something too cheap to time and pays
// for it with a weaker end-to-end check, so the check has to be made here instead.)
//
// The invariant: the frame boundary happens INSIDE the FIFO burst that is being timed, so
// profFrame must close that timer into the field that is ending and restart it for the field
// that is beginning. Neither dropping it nor letting it run on is acceptable.
func TestProfFrameClosesInFlightTimers(t *testing.T) {
	m := testMachine()
	m.SetProfile(true)
	m.profRunEnter()

	// A burst is in flight when the field closes — the copy that ends a frame is a command
	// inside it.
	m.profFIFOEnter()
	spin(2 * time.Millisecond)
	m.profFrame() // the boundary, mid-burst
	p1 := m.FrameProfile()

	// The burst's time so far must be in the field that just closed. It reaches the report
	// through the derived decode bucket, since nothing was nested inside it.
	if got := bucketMs(p1, "command decode (derived)"); got < 1.0 {
		t.Errorf("the in-flight burst contributed %.3f ms to the field it spanned; it was dropped", got)
	}

	// And the timer must still be running: the rest of the burst belongs to the new field.
	spin(2 * time.Millisecond)
	m.profFIFOExit()
	m.profFrame()
	p2 := m.FrameProfile()
	if got := bucketMs(p2, "command decode (derived)"); got < 1.0 {
		t.Errorf("the burst's remainder contributed %.3f ms to the next field; the timer was not restarted", got)
	}
	// Neither field may claim the whole burst: that would be double counting.
	if total := bucketMs(p1, "command decode (derived)") + bucketMs(p2, "command decode (derived)"); total > 6.0 {
		t.Errorf("the two fields claim %.3f ms of a ~4 ms burst between them — it was counted twice", total)
	}
}

// spin burns wall time without sleeping, so the profiler's clock sees work rather than a
// scheduler gap.
func spin(d time.Duration) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
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

// TestProfileOffCostsNothing: with profiling off nothing is recorded, so an oracle run pays
// only the branch.
func TestProfileOffIsSilent(t *testing.T) {
	m := testMachine()
	if p := m.FrameProfile(); len(p.Buckets) != 0 || p.TotalMs != 0 {
		t.Errorf("a machine that was never profiled reports %+v", p)
	}
	// The boundary helpers must be no-ops rather than panicking on a zero clock.
	m.profFIFOEnter()
	m.profEnd(bucketRaster, m.profStart())
	m.profFIFOExit()
	m.profFrame()
	if p := m.FrameProfile(); len(p.Buckets) != 0 {
		t.Errorf("profiling is off but a field was reported: %+v", p)
	}
}

func formatProfile(p FrameProfile) string {
	s := ""
	for _, b := range p.Buckets {
		s += "    " + b.Name + ": " + msString(b.Millis) + "\n"
	}
	return s
}

func msString(f float64) string {
	const digits = "0123456789"
	i := int(f*1000 + 0.5)
	whole, frac := i/1000, i%1000
	out := ""
	if whole == 0 {
		out = "0"
	}
	for w := whole; w > 0; w /= 10 {
		out = string(digits[w%10]) + out
	}
	return out + "." + string(digits[frac/100]) + string(digits[frac/10%10]) + string(digits[frac%10]) + " ms"
}
