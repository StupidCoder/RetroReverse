package gc

// profile.go answers "where does a frame's time actually go?" — per subsystem, on the
// machine itself, without a profiler attached.
//
// The measurement design is the whole point, because the naive version destroys what it
// measures: a time.Now() pair around each fragment costs more than the fragment does, and
// the answer it then gives is about the clock, not the GPU. So nothing here times anything
// fine-grained. It times only boundaries that are already coarse, and which ones those are
// was settled by counting them rather than by taste. One field of the intro cutscene:
//
//	8,785 FIFO bursts        the write-gather pipe handing over a cache line
//	12,770 GX commands       of which 9,774 are draws
//	3,644 DSP batches        the audio core running, 64 instructions at a time
//	~3,400,000 fragments     the rasteriser deciding a pixel
//
// The first three are timed — ~64k clock reads a field, measured at 0.9% of the field's run
// time, against a field that takes ~270 ms. The last is three orders of magnitude too many
// and is not, so THERE IS NO TEXTURE BUCKET: texture decoding happens under a fragment's
// sample, and timing it would cost more than the whole rest of the profile. It is inside
// "rasterise" and stays there. (The 3DS can afford a texture bucket because it has a texture
// cache and times only the misses; this machine reads main memory on every sample, so there
// is no miss to time.)
//
// The buckets are disjoint leaves, so they can be added up:
//
//	command decode   FIFO opcode parsing + CP/XF/BP register loads   (derived, see below)
//	vertex + xf      vertex fetch, transform, lighting               (per draw)
//	rasterise        the triangle loop: TEV, texturing, blending     (per draw)
//	pe copy          EFB->XFB, EFB->texture, and the EFB clear       (per copy)
//	dsp              the audio core                                  (per batch)
//
// and what is left of the field — Gekko interpretation, the devices, the VI/DI/AID ticks —
// is reported as a derived remainder rather than pretended to be measured.
//
// "Command decode" is itself derived: the whole FIFO drain is timed at its one entry point
// (feed), and the draws and copies nested inside it are subtracted out. Timing the decode
// directly would mean a clock pair around each of 12,770 commands to measure the cheapest
// thing in the frame.
//
// A remainder that goes negative would mean the buckets overlap; it is clamped, and a
// remainder pinned at zero is the signal to come back here and look at the nesting rather
// than a measurement. That is not hypothetical — see profFrame.
//
// Time is accumulated only while Run is running, so a debugger sitting paused at a frame
// boundary for ten minutes does not report ten minutes of Gekko.
//
// Live behind Machine.Profile: a run that does not ask for it pays a predictable branch per
// boundary and nothing else.

import "time"

// The measured buckets, in report order. "command decode" and the remainder are derived at
// report time and are not accumulators.
const (
	bucketFIFO   = iota // the whole FIFO drain; the draws/copies below are nested inside it
	bucketVertex        // vertex fetch + transform + lighting
	bucketRaster        // the triangle loop
	bucketCopy          // the pixel engine's copies
	bucketDSP           // the audio core
	numBuckets
)

// ProfileBucket is one subsystem's share of a field.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int // how many of the thing was done: bursts, draws, copies, batches
}

// ProfileCounter is one tally for the field. Fragments per millisecond is the number that
// says whether an optimisation worked; wall time alone cannot tell a faster rasteriser from
// a field that drew less.
type ProfileCounter struct {
	Name  string
	Value int
}

// FrameProfile is the last completed field's cost, by subsystem.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter

	// Drew says whether the game drew anything in this field. Luigi's Mansion renders at
	// 30 Hz on a 60 Hz console, so half the fields are logic alone — and a reader who does
	// not know which is which cannot tell a cheap field from an empty one.
	Drew bool
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time // when the current Run call started
	inRun    bool
	frameNs  int64 // run time accumulated since the last frame boundary

	// The FIFO drain's timer is held here rather than on the stack, because the frame
	// boundary happens INSIDE it and profFrame has to be able to close and restart it.
	fifoStart time.Time
	inFIFO    bool

	base      profCounters // the GPU counters at the last frame boundary
	baseInstr uint64

	last FrameProfile // the last completed field
	has  bool
}

// profCounters is the tally a field's work is reported in.
type profCounters struct {
	cmds, draws, culled, frags, zRejected, aRejected, fifoBytes int
}

func (m *Machine) profCounters() profCounters {
	g := &m.gpu
	return profCounters{
		cmds: m.gxTotalCmds, draws: g.profDraws, culled: g.profCulled, frags: g.pixWritten,
		zRejected: g.pixZRej, aRejected: g.pixARej, fifoBytes: int(m.wgFIFO.Bytes),
	}
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

// profFIFOEnter/profFIFOExit bracket the FIFO drain. Only the outermost pair counts: a
// display list re-enters the interpreter, and timing the inner one would count its
// contents twice.
func (m *Machine) profFIFOEnter() {
	if !m.Profile || m.prof.inFIFO {
		return
	}
	m.prof.fifoStart, m.prof.inFIFO = time.Now(), true
}

func (m *Machine) profFIFOExit() {
	if !m.Profile || !m.prof.inFIFO {
		return
	}
	m.prof.ns[bucketFIFO] += int64(time.Since(m.prof.fifoStart))
	m.prof.count[bucketFIFO]++
	m.prof.inFIFO = false
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
// It is called from inside the pixel engine's copy, which is inside a BP register load,
// which is inside the FIFO drain, which is inside a CPU store. Two timers are therefore in
// flight across this call and both would otherwise be booked to the wrong field:
//
//   - The run timer. Closed into this field and restarted, so the rest of the Run call
//     belongs to the next one.
//   - The FIFO timer. Same treatment — the burst that carried the copy is split at the
//     boundary rather than dropped or double-counted.
//
// The copy's own conversion work is booked by copyDisplay BEFORE it calls this, for the same
// reason. Getting this wrong is not theoretical: on the PSP and the 3DO a deferred "book this
// call, less the work nested inside it" ran after the reset, against a baseline that no
// longer existed, and ADDED the field's draw time to the next field's bucket. It hid on the
// PSP because the field total was big enough to absorb it, and gave itself away on the 3DO as
// a remainder pinned at zero.
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
	if p.inFIFO {
		p.ns[bucketFIFO] += int64(time.Since(p.fifoStart))
		p.count[bucketFIFO]++
		p.fifoStart = time.Now()
	}

	ms := func(ns int64) float64 { return float64(ns) / 1e6 }

	// The draws and the copies happen inside the FIFO drain, so the FIFO accumulator has
	// already counted them. What is left of it is the opcode decoding and the register
	// loads — the one bucket that is too cheap per unit to time directly.
	nested := p.ns[bucketVertex] + p.ns[bucketRaster] + p.ns[bucketCopy]
	decode := p.ns[bucketFIFO] - nested
	if decode < 0 {
		decode = 0
	}

	buckets := []ProfileBucket{
		{Name: "command decode (derived)", Millis: ms(decode), Count: int(p.count[bucketFIFO])},
		{Name: "vertex + xf", Millis: ms(p.ns[bucketVertex]), Count: int(p.count[bucketVertex])},
		{Name: "rasterise", Millis: ms(p.ns[bucketRaster]), Count: int(p.count[bucketRaster])},
		{Name: "pe copy", Millis: ms(p.ns[bucketCopy]), Count: int(p.count[bucketCopy])},
		{Name: "dsp", Millis: ms(p.ns[bucketDSP]), Count: int(p.count[bucketDSP])},
	}

	// What is left is the Gekko interpreter and the devices it drives. Derived, and
	// labelled as derived.
	summed := decode + nested + p.ns[bucketDSP]
	other := total - summed
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "gekko + rest (derived)", Millis: ms(other), Count: 0})

	now := m.profCounters()
	d := profCounters{
		cmds: now.cmds - p.base.cmds, draws: now.draws - p.base.draws,
		culled: now.culled - p.base.culled,
		frags:  now.frags - p.base.frags, zRejected: now.zRejected - p.base.zRejected,
		aRejected: now.aRejected - p.base.aRejected, fifoBytes: now.fifoBytes - p.base.fifoBytes,
	}
	instrs := int(m.Instrs - p.baseInstr)

	p.last = FrameProfile{
		TotalMs: ms(total),
		Drew:    d.draws > 0,
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"gx commands", d.cmds},
			{"draws", d.draws},
			{"tris culled", d.culled},
			{"fragments drawn", d.frags},
			{"depth-rejected", d.zRejected},
			{"alpha-rejected", d.aRejected},
			{"fifo bytes", d.fifoBytes},
			{"gekko instructions", instrs},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.base, p.baseInstr = now, m.Instrs
}

// FrameProfile reports the last completed field's cost by subsystem. The zero value (no
// field profiled yet, or profiling off) has an empty bucket list.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the accumulators so
// the first field reported is a whole field.
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base, m.prof.baseInstr = m.profCounters(), m.Instrs
	}
}
