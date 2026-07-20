package ps2

// profile.go answers "where does a PS2 field's time actually go?" — per subsystem, on the
// machine itself, without a sampling profiler attached. It is the sibling of gc/profile.go
// and n3ds/profile.go, and it exists to make one change measurable: parallelising the GS
// rasteriser is only a win if the rasterise bucket shrinks and the picture does not move.
//
// THE MEASUREMENT DESIGN IS THE WHOLE POINT, because the naive version destroys what it
// measures: a time.Now() pair around each of the millions of fragments a field plots costs
// far more than the fragment does, and the answer it then gives is about the clock, not the
// GS. So nothing here times anything fine-grained. It times only boundaries that are already
// coarse, and which ones those are was settled by COUNTING them rather than by taste. A Jak
// title field is ~5,430 GS primitives; an RRV flyover field ~34,000 — thousands, timeable.
// The fragments those primitives plot are in the millions, and are NOT timed: they live
// inside the rasterise bucket and are reported as a counter, exactly as the GameCube's
// fragments are. (No texture bucket either, for the same reason the GC has none: the PS2
// samples GS local memory on every texel with no cache to miss, so there is nothing to time
// that is not already inside a fragment.)
//
// The timeable leaves, then:
//
//	drain       the whole EE-DMA transport (dmacStart): GIF/VIF/SPR. Everything the frame
//	            draws is nested inside it — VIF unpack -> MSCAL -> VU1 -> XGKICK -> GS, and
//	            PATH2 DIRECT -> GS. Thousands of kicks a field.
//	vu1         one VU1 microprogram run (a MSCAL), the geometry transform, nested in drain —
//	            but EXCLUDING the rasterisation its own XGKICK triggers. A PATH1 kick draws to
//	            the GS synchronously inside vu.Run, so that pixel work lands in the raster
//	            bucket and is subtracted back out of vu1, or it would be counted twice.
//	rasterise   one primitive's pixel loop: texturing, blending, the plot. Nested in drain
//	            (and, for a PATH1 primitive, inside a vu.Run — see vu1 above).
//
// And the reported buckets, three of which are derived from those:
//
//	gif/vif/dma decode (derived)  = drain - vu1 - raster   (the GIFtag/VIF parse + register
//	                                writes + SPR memcpy, minus the geometry and pixels nested
//	                                in the same transport — the GC "command decode" shape)
//	vu1 (geometry)                = vu1
//	rasterise                     = raster
//	ee + iop + rest (derived)     = total - drain
//
// WHY THE IOP IS A COUNTER AND NOT A TIMED LEAF. The IOP steps once every eight EE
// instructions (iopStepRatio), i.e. ~125,000 times per field. Timing each would be a quarter
// of a million clock reads a field — the same three-orders-of-magnitude-too-many mistake the
// fragment would be, and it would distort the very number it set out to measure. So the IOP's
// instruction count is reported as a counter and its time sits, unmeasured, in the "ee + iop +
// rest" remainder alongside the EE interpreter and the devices. The EE is there for the same
// reason: a million CPU.Step()s a field cannot each carry a clock. The remainder is derived
// and labelled derived; a remainder clamped to zero is the signal to come back and look at the
// nesting, not a measurement (see profFrame).
//
// THE FRAME BOUNDARY IS THE VBLANK (deliverVBlank), and it is cleaner than the GameCube's: it
// fires between EE instructions in the run loop's own pacing, NOT nested inside a GS burst, so
// no drain/vu1/raster timer is ever in flight across it. Only the run timer is, and profFrame
// splits it (close into this field, restart for the next) exactly as gc/profile.go does. That
// split is the one thing a direct unit test guards, because a derived remainder absorbs a
// nesting error silently — its bar still adds up to the total and the chart still looks right.
//
// Counters are DELTAS against a baseline captured at each frame boundary, which sidesteps a
// PS2-specific trap: the GS census and the reject tallies are savestated, so a loadstate
// inherits the boot's totals. A per-frame delta never sees those inherited totals, because the
// baseline is taken after the load, when the profiler is (re-)armed.
//
// Time is accumulated only while Run is running, so a debugger paused at a frame boundary for
// ten minutes does not report ten minutes of EE. Live behind Machine.Profile: a run that does
// not ask for it pays a predictable branch per boundary and nothing else.

import "time"

// The timed leaves. "decode" and the remainder are derived at report time and are not
// accumulators of their own.
const (
	bucketDrain  = iota // the whole DMA transport; vu1 and raster below are nested inside it
	bucketVU1           // a VU1 microprogram run (MSCAL)
	bucketRaster        // one primitive's pixel loop
	numBuckets
)

// ProfileBucket is one subsystem's share of a field. Count is how many of the thing was done:
// DMA drains, VU1 kicks, primitives rasterised.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int
}

// ProfileCounter is one tally for the field. FRAGMENTS PER MILLISECOND is the number that
// says whether the rasteriser got faster; wall time alone cannot tell a faster rasteriser
// from a field that drew less.
type ProfileCounter struct {
	Name  string
	Value int
}

// FrameProfile is the last completed field's cost, by subsystem.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter

	// Drew says whether the field put a primitive on the GS. A field that only runs logic
	// (or waits) is a cheap field, and a reader who cannot tell it from an empty one cannot
	// read the numbers.
	Drew bool
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time // when the current Run call started
	inRun    bool
	frameNs  int64 // run time accumulated since the last frame boundary

	// The DMA drain's timer, held here rather than on the stack so profFrame could split it
	// if a vblank ever landed mid-drain. dmacStart is not re-entrant today, but the depth
	// counter makes "only the outermost drain counts" true regardless.
	drainStart time.Time
	drainDepth int

	base profCounters // the counters at the last frame boundary
	last FrameProfile // the last completed field
	has  bool
}

// profCounters is the running totals a field's work is measured as deltas of.
type profCounters struct {
	eeSteps, iopSteps          uint64
	prims, frags               uint64
	rejZ, rejScissor, rejAlpha uint64
	rejDate                    uint64
}

func (m *Machine) profCounters() profCounters {
	c := profCounters{eeSteps: m.steps}
	if m.IOP != nil {
		c.iopSteps = m.IOP.Steps()
	}
	if g := m.gs; g != nil {
		c.prims = uint64(g.prims)
		c.frags = g.plotted
		c.rejZ, c.rejScissor, c.rejAlpha, c.rejDate = g.rejZ, g.rejScissor, g.rejAlpha, g.rejDate
	}
	return c
}

// profStart reads the clock, if profiling is on. The zero time means "off", so profEnd on it
// is a no-op — callers do not have to branch themselves.
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

// profDrainEnter/profDrainExit bracket a DMA transport (dmacStart). Only the outermost pair
// counts, so a drain that starts another is timed once.
func (m *Machine) profDrainEnter() {
	if !m.Profile {
		return
	}
	if m.prof.drainDepth == 0 {
		m.prof.drainStart = time.Now()
	}
	m.prof.drainDepth++
}

func (m *Machine) profDrainExit() {
	if !m.Profile || m.prof.drainDepth == 0 {
		return
	}
	m.prof.drainDepth--
	if m.prof.drainDepth == 0 {
		m.prof.ns[bucketDrain] += int64(time.Since(m.prof.drainStart))
		m.prof.count[bucketDrain]++
	}
}

// profVU1Start/profVU1End bracket one VU1 microprogram run and book it into the vu1 bucket
// EXCLUDING the rasterisation nested inside it. A PATH1 XGKICK runs the GS synchronously from
// within vu.Run, so the raster funcs' own timers fire during the interval this one spans; the
// raster time is captured before and after and subtracted, so vu1 is the geometry transform
// alone and vu1 + raster stay disjoint (which is what lets decode = drain - vu1 - raster be a
// non-negative measurement rather than a double-count). No-ops when profiling off.
func (m *Machine) profVU1Start() (time.Time, int64) {
	if !m.Profile {
		return time.Time{}, 0
	}
	return time.Now(), m.prof.ns[bucketRaster]
}

func (m *Machine) profVU1End(t time.Time, rasterBefore int64) {
	if t.IsZero() {
		return
	}
	nested := m.prof.ns[bucketRaster] - rasterBefore // the XGKICK's own pixels, already booked to raster
	m.prof.ns[bucketVU1] += int64(time.Since(t)) - nested
	m.prof.count[bucketVU1]++
}

// profRunEnter/profRunExit bracket a Run call, so "the field's total" is run time and not
// wall time.
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

// profFrame closes the field: it turns the accumulators into a FrameProfile and resets them.
//
// It is called from deliverVBlank, which is called from the run loop between EE instructions
// — NOT from inside a DMA drain — so only the run timer is in flight here. That timer is
// closed into this field and restarted for the next, exactly as the GameCube splits its run
// timer; getting the restart wrong adds this field's time to the next one's total, and a
// derived remainder would absorb it silently, which is why TestProfFrameSplitsRunTimer tests
// the split directly. (The drain timer is split too, defensively, in case a future path ever
// delivers a vblank mid-drain.)
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
	if p.drainDepth > 0 {
		p.ns[bucketDrain] += int64(time.Since(p.drainStart))
		p.drainStart = time.Now()
	}

	ms := func(ns int64) float64 { return float64(ns) / 1e6 }

	drain, vu1, raster := p.ns[bucketDrain], p.ns[bucketVU1], p.ns[bucketRaster]

	// The geometry and the pixels happen inside the transport, so the drain accumulator has
	// already counted them. What is left of it is the GIFtag/VIF parse, the register writes
	// and the SPR memcpy — too cheap per unit to time directly.
	decode := drain - vu1 - raster
	if decode < 0 {
		decode = 0
	}

	buckets := []ProfileBucket{
		{Name: "gif/vif/dma decode (derived)", Millis: ms(decode), Count: int(p.count[bucketDrain])},
		{Name: "vu1 (geometry)", Millis: ms(vu1), Count: int(p.count[bucketVU1])},
		{Name: "rasterise", Millis: ms(raster), Count: int(p.count[bucketRaster])},
	}

	// What is left of the field is the EE interpreter, the IOP, and the devices. Derived,
	// and labelled derived. Clamped: a negative remainder means the buckets overlap and is a
	// bug report, not a measurement.
	other := total - drain
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "ee + iop + rest (derived)", Millis: ms(other)})

	now := m.profCounters()
	d := profCounters{
		eeSteps: now.eeSteps - p.base.eeSteps, iopSteps: now.iopSteps - p.base.iopSteps,
		prims: now.prims - p.base.prims, frags: now.frags - p.base.frags,
		rejZ: now.rejZ - p.base.rejZ, rejScissor: now.rejScissor - p.base.rejScissor,
		rejAlpha: now.rejAlpha - p.base.rejAlpha, rejDate: now.rejDate - p.base.rejDate,
	}

	p.last = FrameProfile{
		TotalMs: ms(total),
		Drew:    d.prims > 0,
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"gs primitives", int(d.prims)},
			{"fragments drawn", int(d.frags)},
			{"depth-rejected", int(d.rejZ)},
			{"scissor-rejected", int(d.rejScissor)},
			{"alpha-rejected", int(d.rejAlpha)},
			{"date-rejected", int(d.rejDate)},
			{"vu1 kicks", int(p.count[bucketVU1])},
			{"ee instructions", int(d.eeSteps)},
			{"iop instructions", int(d.iopSteps)},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.base = now
}

// FrameProfile reports the last completed field's cost by subsystem. The zero value (no field
// profiled yet, or profiling off) has an empty bucket list.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the accumulators and
// takes the counter baseline, so the first field reported is a whole field measured against
// the state as it stands now (which is what makes the delta immune to a loadstate's inherited
// census totals).
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base = m.profCounters()
	}
}
