package psp

// profile.go answers "where does a frame's time actually go?" — per subsystem, on the
// machine itself, without a profiler attached.
//
// The measurement design is borrowed wholesale from the 3DS's (n3ds/profile.go), for
// the reason given there: the naive version destroys what it measures. A time.Now()
// pair around each fragment costs more than the fragment does, and the answer it gives
// is then about the clock, not the GE. So nothing here times anything fine-grained. It
// times only boundaries that are already coarse — a display list, a primitive's vertex
// fetch, a primitive's triangle loop, a block transfer, a syscall — and the interpreter
// is a derived remainder rather than a pretended measurement.
//
// The buckets are disjoint leaves, so they add up:
//
//	list      walking a display list out of memory at enqueue      (per list)
//	vertex    decoding and transforming a primitive's vertices     (per primitive)
//	raster    a primitive's triangle/sprite loop, texturing inside (per primitive)
//	xfer      a GE block transfer                                  (per transfer)
//	syscall   the kernel HLE, LESS the GE work done inside it      (per syscall)
//
// That last subtraction is the one thing this file has to get right that the 3DS did
// not. On the 3DS the GPU runs asynchronously and an svc is an svc. On the PSP the GE
// runs SYNCHRONOUSLY inside sceGeListEnQueue — the whole frame's drawing happens under
// one syscall — so timing syscalls naively would count every list, every triangle and
// every fragment twice, the buckets would overlap, and the remainder would go negative.
// So the syscall bucket subtracts the GE time accumulated while it ran.
//
// Time is accumulated only while Run is running, so a debugger sitting paused at a
// frame boundary for ten minutes does not report ten minutes of Allegrex.

import "time"

// The buckets, in report order.
const (
	bucketList = iota
	bucketVertex
	bucketRaster
	bucketXfer
	bucketSyscall
	numBuckets
)

var bucketNames = [numBuckets]string{
	bucketList:    "list capture",
	bucketVertex:  "vertex + transform",
	bucketRaster:  "rasterise",
	bucketXfer:    "block transfer",
	bucketSyscall: "syscall HLE",
}

// geBuckets are the buckets that are GE work, and that a syscall containing the GE
// must therefore not also claim.
var geBuckets = [...]int{bucketList, bucketVertex, bucketRaster, bucketXfer}

// ProfileBucket is one subsystem's share of a frame.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int // how many of the thing was done: lists, primitives, transfers, syscalls
}

// ProfileCounter is one tally for the frame. Fragments per millisecond is the number
// that says whether an optimisation worked; wall time alone cannot tell a faster
// rasteriser from a frame that drew less.
type ProfileCounter struct {
	Name  string
	Value int
}

// FrameProfile is the last completed frame's cost, by subsystem.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter
}

// geCounters is the tally the frame's work is reported in. Kept unconditionally: an
// integer increment beside a fragment that has just been textured and blended is not
// a cost anyone can measure, and having the counts always available means the profile
// does not change the run it profiles.
type geCounters struct {
	lists, prims, verts       int
	frags, zKilled, aKilled   int
	sKilled, scissored, xfers int
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time // when the current Run call started
	inRun    bool
	frameNs  int64 // run time accumulated since the last frame boundary

	base      geCounters // the counters at the last frame boundary
	baseInstr uint64

	// gen counts frames closed. A syscall that was in flight when the frame closed
	// belongs to a frame that has already been reported — see profEndSyscall.
	gen int

	last FrameProfile // the last completed frame
	has  bool
}

// profStart reads the clock, if profiling is on. The zero time means "off", so profEnd
// on it is a no-op — callers do not have to branch themselves.
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

// profGeNs is the GE time booked so far — what a syscall must not claim as its own.
func (m *Machine) profGeNs() int64 {
	var n int64
	for _, b := range geBuckets {
		n += m.prof.ns[b]
	}
	return n
}

// profEndSyscall books a syscall's time, less whatever GE work ran inside it.
//
// The gen check is not decoration. sceDisplaySetFrameBuf is itself a syscall, and it
// closes the frame from inside — profFrame runs, reports the buckets and zeroes them, and
// only then does this deferred call return. Its geBefore baseline now refers to counters
// that no longer exist, so the subtraction below would ADD the whole frame's GE time
// rather than remove it, and dump it into the NEXT frame's syscall bucket. (It hid here:
// a PSP frame's total is large enough that the inflated bucket still fit inside it. The
// 3DO, whose frame is mostly cel drawing, failed the disjointness test outright and gave
// this away.)
func (m *Machine) profEndSyscall(t time.Time, geBefore int64, gen int) {
	if t.IsZero() || gen != m.prof.gen {
		return
	}
	ns := int64(time.Since(t)) - (m.profGeNs() - geBefore)
	if ns < 0 {
		ns = 0 // the clock is not that precise; a tiny negative is noise, not a claim
	}
	m.prof.ns[bucketSyscall] += ns
	m.prof.count[bucketSyscall]++
}

// profRunEnter/profRunExit bracket a Run call, so "the frame's total" is run time and
// not wall time.
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

// profFrame closes the frame: it turns the accumulators into a FrameProfile and resets
// them. Called at the flip, before the debugger's frame hook, so a debugger that stops
// on the hook reads the frame it just watched.
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

	buckets := make([]ProfileBucket, 0, numBuckets+1)
	var summed int64
	for b := 0; b < numBuckets; b++ {
		summed += p.ns[b]
		buckets = append(buckets, ProfileBucket{Name: bucketNames[b], Millis: ms(p.ns[b]), Count: int(p.count[b])})
	}
	// What is left is the Allegrex interpreter, the scheduler and the HLE outside a
	// syscall. Derived, and labelled as derived. A remainder that went negative would
	// mean the buckets overlap — it is clamped, and a zero here is the signal to come
	// back and look at the nesting rather than to trust the picture.
	other := total - summed
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "allegrex + rest (derived)", Millis: ms(other), Count: 0})

	now := m.geCnt
	d := geCounters{
		lists: now.lists - p.base.lists, prims: now.prims - p.base.prims,
		verts: now.verts - p.base.verts, frags: now.frags - p.base.frags,
		zKilled: now.zKilled - p.base.zKilled, aKilled: now.aKilled - p.base.aKilled,
		sKilled: now.sKilled - p.base.sKilled, scissored: now.scissored - p.base.scissored,
		xfers: now.xfers - p.base.xfers,
	}

	p.last = FrameProfile{
		TotalMs: ms(total),
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"display lists", d.lists},
			{"primitives", d.prims},
			{"vertices", d.verts},
			{"fragments drawn", d.frags},
			{"depth-killed", d.zKilled},
			{"alpha-killed", d.aKilled},
			{"stencil-killed", d.sKilled},
			{"scissored", d.scissored},
			{"block transfers", d.xfers},
			{"allegrex instructions", int(m.CPU.Steps - p.baseInstr)},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.base, p.baseInstr = now, m.CPU.Steps
	p.gen++
}

// FrameProfile reports the last completed frame's cost by subsystem. The zero value
// (no frame profiled yet, or profiling off) has an empty bucket list.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the
// accumulators, so the first frame reported is a whole frame and not the difference
// between two unrelated machines.
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base, m.prof.baseInstr = m.geCnt, m.CPU.Steps
	}
}
