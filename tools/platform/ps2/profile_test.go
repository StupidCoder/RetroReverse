package ps2

// profile_test.go guards the two things about the field profiler that a chart cannot show you
// are wrong: that the buckets ADD UP to the field total, and that the run timer is SPLIT
// correctly at the frame boundary.
//
// The second is the one that needs a direct test. "gif/vif/dma decode" and "ee + iop + rest"
// are DERIVED — decode = drain - vu1 - raster, remainder = total - drain — so a nesting error
// is absorbed into a derived bucket, its bar still adds up to the total, and the disjointness
// test below still passes. The only way to catch a mis-split run timer is to test the split
// mechanism itself, mutation-tested both ways, which is what TestProfFrameSplitsRunTimer does.

import (
	"testing"
	"time"
)

// TestProfileBucketsSumToTotal: the reported buckets partition the field. Because two of them
// are derived, this is a check on the arithmetic (a sign error, a wrong subtrahend), not proof
// the attribution is right — see the file comment.
func TestProfileBucketsSumToTotal(t *testing.T) {
	m := NewMachine()
	m.SetProfile(true)

	// A synthetic field: 10 ms total, of which the DMA transport was 6 ms, and inside that
	// 2.5 ms was VU1 and 2 ms was the rasteriser. decode should be 6-2.5-2 = 1.5 ms, and the
	// remainder 10-6 = 4 ms.
	p := &m.prof
	p.frameNs = int64(10 * time.Millisecond)
	p.ns[bucketDrain] = int64(6 * time.Millisecond)
	p.ns[bucketVU1] = int64(2500 * time.Microsecond)
	p.ns[bucketRaster] = int64(2 * time.Millisecond)
	p.count[bucketDrain], p.count[bucketVU1], p.count[bucketRaster] = 3, 4, 5

	m.profFrame()
	fp := m.FrameProfile()

	if fp.TotalMs < 9.9 || fp.TotalMs > 10.1 {
		t.Fatalf("total = %.3f ms, want ~10", fp.TotalMs)
	}
	byName := map[string]ProfileBucket{}
	var sum float64
	for _, b := range fp.Buckets {
		byName[b.Name] = b
		sum += b.Millis
		if b.Millis < 0 {
			t.Errorf("bucket %q is negative: %.3f ms", b.Name, b.Millis)
		}
	}
	if d := sum - fp.TotalMs; d < -0.05 || d > 0.05 {
		t.Errorf("buckets sum to %.3f ms, total is %.3f ms (diff %.3f)", sum, fp.TotalMs, d)
	}
	if got := byName["gif/vif/dma decode (derived)"].Millis; got < 1.45 || got > 1.55 {
		t.Errorf("decode = %.3f ms, want ~1.5 (drain 6 - vu1 2.5 - raster 2)", got)
	}
	if got := byName["ee + iop + rest (derived)"].Millis; got < 3.95 || got > 4.05 {
		t.Errorf("remainder = %.3f ms, want ~4 (total 10 - drain 6)", got)
	}
	if got := byName["gif/vif/dma decode (derived)"].Count; got != 3 {
		t.Errorf("decode count = %d, want 3 (the drain count)", got)
	}
}

// TestProfileDecodeClampsNotNegative: if the nested leaves ever exceed the drain that is
// supposed to contain them, decode would go negative — a bug report, not a measurement. It is
// clamped to zero (the "come back and look at the nesting" signal), never negative.
func TestProfileDecodeClampsNotNegative(t *testing.T) {
	m := NewMachine()
	m.SetProfile(true)
	p := &m.prof
	p.frameNs = int64(10 * time.Millisecond)
	p.ns[bucketDrain] = int64(3 * time.Millisecond)
	p.ns[bucketVU1] = int64(2 * time.Millisecond)
	p.ns[bucketRaster] = int64(2 * time.Millisecond) // vu1+raster (4) > drain (3)

	m.profFrame()
	for _, b := range m.FrameProfile().Buckets {
		if b.Millis < 0 {
			t.Errorf("bucket %q went negative (%.3f ms) instead of clamping", b.Name, b.Millis)
		}
	}
}

// TestProfFrameSplitsRunTimer is the test the derived buckets structurally cannot be: it
// checks that profFrame closes the run timer into THIS field and restarts it for the next, so
// one field's time is not booked to the following field's total.
//
// It is mutation-tested both ways against profile.go's profFrame:
//   - drop `p.runStart = time.Now()` (the restart): the second field's total balloons to
//     include the first field's already-counted run, and the assertion on f2 fails.
//   - drop `p.frameNs = 0` (the close/reset): frameNs accumulates across fields and the second
//     field's total is inflated by the first's, and the assertion on f2 fails.
func TestProfFrameSplitsRunTimer(t *testing.T) {
	if testing.Short() {
		t.Skip("uses short real sleeps to exercise the timer; -short skips it")
	}
	m := NewMachine()
	m.SetProfile(true)

	// Field 1: pretend ~12 ms of run has elapsed since the run started.
	m.prof.inRun = true
	m.prof.runStart = time.Now().Add(-12 * time.Millisecond)
	m.profFrame()
	f1 := m.FrameProfile().TotalMs
	if f1 < 10 || f1 > 20 {
		t.Fatalf("field 1 total = %.1f ms, want ~12", f1)
	}

	// The gap between fields, which the run timer must NOT have counted after the split.
	time.Sleep(4 * time.Millisecond)

	// Field 2: close again. If the restart happened, field 2 is only the ~4 ms since the
	// split; if it did not, it is the whole ~16 ms since field 1's phantom start.
	m.profFrame()
	f2 := m.FrameProfile().TotalMs
	if f2 > 10 {
		t.Errorf("field 2 total = %.1f ms — the run timer was not split at the boundary (field 1's time leaked in)", f2)
	}
}

// TestSetProfileResetsAndArms: turning the profiler on clears the accumulators and takes the
// counter baseline, so the first field is measured against the state as it stands — which is
// what makes the counter deltas immune to a loadstate's inherited census totals.
func TestSetProfileResetsAndArms(t *testing.T) {
	m := NewMachine()
	m.SetProfile(true)
	m.prof.ns[bucketRaster] = 999
	m.prof.frameNs = 999
	m.SetProfile(true) // re-arm
	if m.prof.ns[bucketRaster] != 0 || m.prof.frameNs != 0 {
		t.Error("re-arming did not reset the accumulators")
	}
	if !m.Profile {
		t.Error("re-arming left Profile off")
	}
	m.SetProfile(false)
	if m.Profile {
		t.Error("SetProfile(false) left Profile on")
	}
	// With profiling off, FrameProfile is the zero value (no field ever completed).
	if fp := m.FrameProfile(); len(fp.Buckets) != 0 {
		t.Errorf("profiler off but FrameProfile has %d buckets", len(fp.Buckets))
	}
}
