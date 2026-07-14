package n3ds

// profile.go answers "where does a frame's time actually go?" — per subsystem, on
// the machine itself, without a profiler attached.
//
// The measurement design is the whole point, because the naive version destroys
// what it measures: a time.Now() pair around each fragment costs more than the
// fragment does, and the answer it gives is then about the clock, not the GPU. So
// nothing here times anything fine-grained. It times only boundaries that are
// already coarse — a command list, a draw, a texture cache miss, an audio frame,
// a supervisor call — which on Captain Toad is a few hundred clock reads against
// a few hundred thousand fragments. The overhead is unmeasurable, and the
// resolution is exactly the one an optimisation is chosen at.
//
// The buckets are disjoint leaves, so they can be added up:
//
//	cmd      reading a command list out of memory and decoding it     (per list)
//	vertex   the vertex fetch + shader loop in a draw                 (per draw)
//	raster   the triangle loop in a draw, less texture decoding       (per draw)
//	texture  decoding a texture on a cache miss                       (per miss)
//	gx       DisplayTransfer / MemoryFill / TextureCopy               (per command)
//	dsp      one audio frame                                          (~3.4/VBlank)
//	svc      one supervisor call                                      (per svc)
//
// and what is left of the frame — ARM11 interpretation, the scheduler, the HLE
// outside an svc — is reported as a derived remainder rather than pretended to be
// measured. A remainder that goes negative would mean the buckets are overlapping;
// it is clamped and would show up as an "other" of zero, which is the signal to
// come back here and look at the nesting.
//
// Time is accumulated only while Run is running, so a debugger that sits paused at
// a frame boundary for ten minutes does not report ten minutes of ARM11.
//
// Live behind Machine.Profile: an oracle run that does not ask for it pays a
// predictable-branch on a bool per boundary and nothing else.

import "time"

// The buckets, in report order.
const (
	bucketCmd = iota
	bucketVertex
	bucketRaster
	bucketTexture
	bucketGX
	bucketDSP
	bucketSVC
	numBuckets
)

var bucketNames = [numBuckets]string{
	bucketCmd:     "command decode",
	bucketVertex:  "vertex + shader",
	bucketRaster:  "rasterise",
	bucketTexture: "texture decode",
	bucketGX:      "gx transfers",
	bucketDSP:     "dsp",
	bucketSVC:     "svc + ipc",
}

// ProfileBucket is one subsystem's share of a frame.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int // how many of the thing was done: lists, draws, misses, svcs
}

// ProfileCounter is one tally for the frame. Fragments per millisecond is the
// number that says whether an optimisation worked; wall time alone cannot tell a
// faster rasteriser from a frame that drew less.
type ProfileCounter struct {
	Name  string
	Value int
}

// FrameProfile is the last completed frame's cost, by subsystem.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter

	// Drew says whether the game drew anything in this field. It renders at 30 Hz on a
	// 60 Hz console, so half the fields are logic alone — and a reader that does not
	// know which is which cannot tell a cheap frame from an empty one.
	Drew bool
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time // when the current Run call started
	inRun    bool
	frameNs  int64 // run time accumulated since the last frame boundary

	idleSkips int
	base      profCounters // the GPU counters at the last frame boundary
	baseInstr uint64

	last FrameProfile // the last completed frame
	has  bool
}

// profCounters is the tally the frame's work is reported in.
type profCounters struct {
	draws, frags, depthKilled, culled, rejected, shadowWrites, listHops int
}

func (m *Machine) profCounters() profCounters {
	g := m.gpu
	return profCounters{g.Draws, g.PixelsDrawn, g.DepthKilled, g.CulledTris, g.RejectedTris, g.ShadowWrites, g.ListHops}
}

// profStart reads the clock, if profiling is on. The zero time means "off", so
// profEnd on it is a no-op — callers do not have to branch themselves.
func (m *Machine) profStart() time.Time {
	if !m.Profile {
		return time.Time{}
	}
	return time.Now()
}

func (m *Machine) profEnd(bucket int, t time.Time) {
	if t.IsZero() {
		return
	}
	m.prof.ns[bucket] += int64(time.Since(t))
	m.prof.count[bucket]++
}

// profRunEnter/profRunExit bracket a Run call, so "the frame's total" is run time
// and not wall time.
func (m *Machine) profRunEnter() {
	if !m.Profile {
		return
	}
	m.prof.runStart, m.prof.inRun = time.Now(), true
}

func (m *Machine) profRunExit() {
	if !m.Profile || !m.prof.inRun {
		return
	}
	m.prof.frameNs += int64(time.Since(m.prof.runStart))
	m.prof.inRun = false
}

// profFrame closes the frame: it turns the accumulators into a FrameProfile and
// resets them. Called from deliverVBlank, before the debugger's frame hook, so a
// debugger that stops on the hook reads the frame it just watched.
func (m *Machine) profFrame() {
	if !m.Profile {
		return
	}
	p := &m.prof
	total := p.frameNs
	if p.inRun {
		total += int64(time.Since(p.runStart))
		p.runStart = time.Now()
	}
	p.frameNs = 0

	ms := func(ns int64) float64 { return float64(ns) / 1e6 }

	// The texture decode sits inside the triangle loop (a cache miss happens under
	// a fragment's sampleTexture), so the raster bucket has already counted it.
	// Subtract it out to keep the buckets disjoint.
	raster := p.ns[bucketRaster] - p.ns[bucketTexture]
	if raster < 0 {
		raster = 0
	}

	buckets := make([]ProfileBucket, 0, numBuckets+1)
	var summed int64
	for b := 0; b < numBuckets; b++ {
		ns := p.ns[b]
		if b == bucketRaster {
			ns = raster
		}
		summed += ns
		buckets = append(buckets, ProfileBucket{Name: bucketNames[b], Millis: ms(ns), Count: int(p.count[b])})
	}
	// What is left is the ARM11 interpreter, the scheduler and the HLE outside an
	// svc. Derived, and labelled as derived.
	other := total - summed
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "arm11 + rest (derived)", Millis: ms(other), Count: 0})

	now := m.profCounters()
	d := profCounters{
		draws: now.draws - p.base.draws, frags: now.frags - p.base.frags,
		depthKilled: now.depthKilled - p.base.depthKilled, culled: now.culled - p.base.culled,
		rejected: now.rejected - p.base.rejected, shadowWrites: now.shadowWrites - p.base.shadowWrites,
		listHops: now.listHops - p.base.listHops,
	}
	instrs := int(m.instrs - p.baseInstr)

	p.last = FrameProfile{
		TotalMs: ms(total),
		Drew:    d.draws > 0,
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"draws", d.draws},
			{"fragments drawn", d.frags},
			{"depth-killed", d.depthKilled},
			{"tris culled", d.culled},
			{"tris w-rejected", d.rejected},
			{"shadow writes", d.shadowWrites},
			{"list hops", d.listHops},
			{"arm11 instructions", instrs},
			{"idle skips", p.idleSkips},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.idleSkips = 0
	p.base, p.baseInstr = now, m.instrs
}

// FrameProfile reports the last completed frame's cost by subsystem. The zero
// value (no frame profiled yet, or profiling off) has an empty bucket list.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the
// accumulators so the first frame reported is a whole frame.
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base, m.prof.baseInstr = m.profCounters(), m.instrs
	}
}
