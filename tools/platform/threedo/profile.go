package threedo

// profile.go answers "where does a frame's time actually go?" — per subsystem, on the
// machine itself, without a profiler attached.
//
// The measurement design follows the 3DS's (n3ds/profile.go) and the PSP's, for the
// reason given there: the naive version destroys what it measures. Nothing here times a
// pixel. It times only boundaries that are already coarse — a cel, a screen clear, a
// folio call, an SWI — and what is left of the frame is reported as a derived remainder
// rather than pretended to be measured.
//
// The buckets are disjoint leaves, so they add up:
//
//	cel        decoding and drawing one cel                     (per cel)
//	clear      a FlashClear over a buffer                       (per clear)
//	folio      a Portfolio folio call, LESS the cels inside it  (per call)
//	swi        one supervisor call                              (per SWI)
//
// The folio bucket subtracts the cel time inside it for the same reason the PSP's syscall
// bucket subtracts the GE: DrawCels is a folio call that draws the entire chain before it
// returns, so a folio bucket that did not subtract would count every cel twice and drive
// the remainder negative.

import "time"

// The buckets, in report order.
const (
	bucketCel = iota
	bucketClear
	bucketFolio
	bucketSWI
	numBuckets
)

var bucketNames = [numBuckets]string{
	bucketCel:   "cel draw",
	bucketClear: "flash clear",
	bucketFolio: "folio HLE",
	bucketSWI:   "swi",
}

// gfxBuckets are the buckets that are cel-engine work, and that a folio call containing
// the cel engine must therefore not also claim.
var gfxBuckets = [...]int{bucketCel, bucketClear}

// ProfileBucket is one subsystem's share of a frame.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int
}

// ProfileCounter is one tally for the frame.
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

// celCounters is the tally the frame's work is reported in.
type celCounters struct {
	cels   int
	pixels int
	chains int
	clears int
	folios int
	swis   int
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time
	inRun    bool
	frameNs  int64

	base      celCounters
	baseInstr uint64

	// gen counts frames closed. A folio call that was in flight when the frame closed
	// belongs to a frame that has already been reported, and must not be booked into
	// the next one — see profEndFolio.
	gen int

	last FrameProfile
	has  bool
}

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

// profGfxNs is the cel-engine time booked so far — what a folio call must not claim.
func (m *Machine) profGfxNs() int64 {
	var n int64
	for _, b := range gfxBuckets {
		n += m.prof.ns[b]
	}
	return n
}

// profEndFolio books a folio call's time, less whatever cel work ran inside it.
//
// The gen check is not decoration. DisplayScreen is itself a folio call, and it closes
// the frame from inside — profFrame runs, reports the buckets and zeroes them, and only
// then does this deferred call return. Its gfxBefore baseline now refers to a set of
// counters that no longer exists, so the subtraction below would ADD the whole frame's
// cel time instead of removing it, and dump it into the NEXT frame's folio bucket. The
// frame that call belongs to has already been reported; the honest thing is to drop it.
func (m *Machine) profEndFolio(t time.Time, gfxBefore int64, gen int) {
	if t.IsZero() || gen != m.prof.gen {
		return
	}
	ns := int64(time.Since(t)) - (m.profGfxNs() - gfxBefore)
	if ns < 0 {
		ns = 0 // the clock is not that precise; a tiny negative is noise, not a claim
	}
	m.prof.ns[bucketFolio] += ns
	m.prof.count[bucketFolio]++
}

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

// profFrame closes the frame at DisplayScreen: it turns the accumulators into a
// FrameProfile and resets them.
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
	// What is left is the ARM60 interpreter, the scheduler and the HLE outside a folio
	// call. Derived, and labelled as derived. A remainder that went negative would mean
	// the buckets overlap; it is clamped, and a zero here is the signal to come back and
	// look at the nesting rather than to trust the picture.
	other := total - summed
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "arm60 + rest (derived)", Millis: ms(other), Count: 0})

	now := m.celCnt
	d := celCounters{
		cels: now.cels - p.base.cels, pixels: now.pixels - p.base.pixels,
		chains: now.chains - p.base.chains, clears: now.clears - p.base.clears,
		folios: now.folios - p.base.folios, swis: now.swis - p.base.swis,
	}

	p.last = FrameProfile{
		TotalMs: ms(total),
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"cels drawn", d.cels},
			{"pixels written", d.pixels},
			{"DrawCels chains", d.chains},
			{"flash clears", d.clears},
			{"folio calls", d.folios},
			{"SWIs", d.swis},
			{"arm60 instructions", int(m.CPU.Instrs - p.baseInstr)},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.base, p.baseInstr = now, m.CPU.Instrs
	p.gen++
}

// FrameProfile reports the last completed frame's cost by subsystem.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the accumulators
// so the first frame reported is a whole frame.
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base, m.prof.baseInstr = m.celCnt, m.CPU.Instrs
	}
}
