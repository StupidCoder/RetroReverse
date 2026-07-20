package xbox

// profile.go answers "where does one presented frame's time actually go?" — per
// subsystem, on the machine itself, without a host sampling profiler attached. It is the
// GameCube's gc/profile.go transposed onto the NV2A, and the transposition is the whole
// content of this file, because the boundaries a console affords for timing are decided by
// COUNTING them, not by taste. One presented frame of OutRun's race (race-leg1):
//
//	1 flip                   the FLIP_STALL that ends the frame
//	112,881 nv methods       (subchannel, method, argument) writes the pusher decodes
//	573 draws                BEGIN/END batches
//	14,166,941 fragments     the rasteriser deciding a pixel (6.4M drawn, 6.1M z-killed,
//	                         1.7M alpha-killed)
//
// The draws are timed per draw (~64 clock reads/frame each side — free). The methods are
// NOT: 112k clock pairs would measure the clock, not the decode, so the whole push-buffer
// drain is timed at ONE coarse boundary (runPusher, the analogue of the GameCube's FIFO
// burst) and the decode is derived from it. The fragments are three orders of magnitude
// too many to time at all — so THERE IS NO TEXTURE BUCKET, exactly as on the GameCube:
// texture sampling happens under a fragment and lives inside "rasterise". (The NV2A reads
// RAM on every sample; there is no texture-cache miss to time, which is the only thing the
// 3DS can afford to bucket separately.) And THERE IS NO AUDIO BUCKET: the APU (apu.go) is
// HLE MMIO latches and executes no DSP56k microcode, so there is no audio work to measure.
//
// The buckets are disjoint leaves plus two derived remainders:
//
//	vertex + xf      vertex fetch, transform, shader        (per draw, runDraw)
//	rasterise        the triangle fill loop + texturing     (per draw, assemble)
//	clear            CLEAR_SURFACE fills                     (per clear, clearSurface)
//	command decode   push-buffer opcode/register decoding   (derived: pusher − the above)
//	x86 + rest       the interpreter and the devices        (derived: run − pusher)
//
// The nesting is: Run (CPU steps) ⊃ a store to DMA_PUT → runPusher ⊃ { method decode +
// vertex + raster + clear }. So the pusher bucket is a CONTAINER that has already counted
// the three leaves nested in it; "command decode" is what is left of it, and "x86 + rest"
// is what is left of the run once the pusher is subtracted. A remainder that would go
// negative means the buckets overlap; it is clamped, and a remainder pinned at zero is the
// signal to come back here and look at the nesting rather than a measurement.
//
// Time accrues only while Run is running, so a debugger paused at a flip for ten minutes
// does not report ten minutes of x86. Live behind Machine.Profile: a run that does not ask
// for it pays a predictable branch per boundary (profStart returns the zero time, so
// profEnd is a no-op) and nothing else. The per-fragment counters are the one thing timed
// unconditionally — but they are three integer increments threaded through a per-lane
// rstats (nv2a_raster.go), never a shared counter, so they cost the same whether or not
// anyone is profiling and never race the parallel fill.

import "time"

// The measured buckets, in accumulator order. "command decode" and "x86 + rest" are
// derived at report time and are not accumulators of their own.
const (
	bucketVertex = iota // vertex fetch + transform + shader
	bucketRaster        // the triangle fill loop
	bucketClear         // CLEAR_SURFACE fills
	bucketPusher        // the whole push-buffer drain; the three above nest inside it
	numBuckets
)

// ProfileBucket is one subsystem's share of a presented frame. Count is how many of the
// thing was done: draws, clears, pusher drains.
type ProfileBucket struct {
	Name   string
	Millis float64
	Count  int
}

// ProfileCounter is one tally for the frame. Fragments per millisecond is the number that
// says whether an optimisation worked; wall time alone cannot tell a faster rasteriser
// from a frame that drew less.
type ProfileCounter struct {
	Name  string
	Value int
}

// FrameProfile is the last presented frame's cost, by subsystem.
type FrameProfile struct {
	TotalMs  float64
	Buckets  []ProfileBucket
	Counters []ProfileCounter

	// Drew says whether the frame put geometry on the screen. The Xbox's frame boundary is
	// the FLIP, and a flip normally follows a drawn frame, so this is usually true — but a
	// flip that presents an unchanged buffer (a paused menu) draws nothing, and a reader
	// who does not know which is which cannot tell a cheap frame from an empty one.
	Drew bool
}

// profCounters is the tally a frame's work is reported in, read straight off the pgraph.
type profCounters struct {
	draws, methods, frags, zRej, aRej int
}

func (m *Machine) profCounters() profCounters {
	g := m.pgraph
	return profCounters{
		draws: g.Draws, methods: g.Methods,
		frags: g.pixWritten, zRej: g.pixZRej, aRej: g.pixARej,
	}
}

type profState struct {
	ns    [numBuckets]int64
	count [numBuckets]int64

	runStart time.Time // when the current Run call started
	inRun    bool
	frameNs  int64 // run time accumulated since the last frame boundary

	// The pusher-drain timer is held here rather than on the stack, because the frame
	// boundary (FLIP_STALL) happens INSIDE a drain and profFrame has to be able to close
	// and restart it.
	pushStart time.Time
	inPush    bool

	base      profCounters // the pgraph counters at the last frame boundary
	baseInstr uint64       // CPU.Steps at the last frame boundary

	last FrameProfile // the last completed frame
	has  bool
}

// profStart reads the clock, if profiling is on. The zero time means "off", so profEnd on
// it is a no-op — callers do not have to branch themselves.
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

// profPusherEnter/profPusherExit bracket the push-buffer drain (runPusher). runPusher's own
// re-entrancy guard means a call that gets this far is always the outermost, but the inPush
// flag is kept anyway so profFrame — which fires mid-drain at the flip — can tell whether a
// timer is in flight.
func (m *Machine) profPusherEnter() {
	if !m.Profile || m.prof.inPush {
		return
	}
	m.prof.pushStart, m.prof.inPush = time.Now(), true
}

func (m *Machine) profPusherExit() {
	if !m.Profile || !m.prof.inPush {
		return
	}
	m.prof.ns[bucketPusher] += int64(time.Since(m.prof.pushStart))
	m.prof.count[bucketPusher]++
	m.prof.inPush = false
}

// profRunEnter/profRunExit bracket a Run call, so "the frame's total" is run time and not
// wall time (a debugger paused at a flip must not bank the pause as x86).
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
// them. It is called at FLIP_STALL — inside a method dispatch, inside runPusher, inside the
// CPU store that kicked the pusher, inside Run. Two timers are therefore in flight across
// this call and both would otherwise be booked to the wrong frame:
//
//   - The run timer. Closed into this frame and restarted, so the rest of the Run call
//     belongs to the next one.
//   - The pusher timer. Same treatment — the drain that carried the flip is split at the
//     boundary rather than dropped or double-counted.
//
// Getting this wrong is not theoretical: on the PSP, the 3DO and the GameCube a deferred
// "book this call, less the work nested inside it" ran after the reset, against a baseline
// that no longer existed, and ADDED the frame's draw time to the next frame's bucket. It
// hid wherever the frame total was big enough to absorb it and gave itself away as a
// remainder pinned at zero. TestProfFrameClosesInFlightTimers guards both halves.
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
	if p.inPush {
		p.ns[bucketPusher] += int64(time.Since(p.pushStart))
		p.count[bucketPusher]++
		p.pushStart = time.Now()
	}

	ms := func(ns int64) float64 { return float64(ns) / 1e6 }

	// The vertex, raster and clear work all happens inside the pusher drain, so the pusher
	// accumulator has already counted them. What is left of it is the opcode decoding and
	// the register loads — the one bucket too cheap per unit to time directly.
	nested := p.ns[bucketVertex] + p.ns[bucketRaster] + p.ns[bucketClear]
	decode := p.ns[bucketPusher] - nested
	if decode < 0 {
		decode = 0
	}

	buckets := []ProfileBucket{
		{Name: "vertex + xf", Millis: ms(p.ns[bucketVertex]), Count: int(p.count[bucketVertex])},
		{Name: "rasterise", Millis: ms(p.ns[bucketRaster]), Count: int(p.count[bucketRaster])},
		{Name: "clear", Millis: ms(p.ns[bucketClear]), Count: int(p.count[bucketClear])},
		{Name: "command decode (derived)", Millis: ms(decode), Count: int(p.count[bucketPusher])},
	}

	// What is left of the run once the pusher is subtracted is the x86 interpreter and the
	// devices it drives. Derived, and labelled as derived.
	other := total - p.ns[bucketPusher]
	if other < 0 {
		other = 0
	}
	buckets = append(buckets, ProfileBucket{Name: "x86 + rest (derived)", Millis: ms(other), Count: 0})

	now := m.profCounters()
	d := profCounters{
		draws: now.draws - p.base.draws, methods: now.methods - p.base.methods,
		frags: now.frags - p.base.frags, zRej: now.zRej - p.base.zRej, aRej: now.aRej - p.base.aRej,
	}
	instrs := int(m.CPU.Steps - p.baseInstr)

	p.last = FrameProfile{
		TotalMs: ms(total),
		Drew:    d.draws > 0,
		Buckets: buckets,
		Counters: []ProfileCounter{
			{"draws", d.draws},
			{"nv methods", d.methods},
			{"fragments drawn", d.frags},
			{"depth-killed", d.zRej},
			{"alpha-killed", d.aRej},
			{"x86 instructions", instrs},
		},
	}
	p.has = true

	p.ns, p.count = [numBuckets]int64{}, [numBuckets]int64{}
	p.base, p.baseInstr = now, m.CPU.Steps
}

// FrameProfile reports the last presented frame's cost by subsystem. The zero value (no
// frame profiled yet, or profiling off) has an empty bucket list.
func (m *Machine) FrameProfile() FrameProfile {
	if !m.prof.has {
		return FrameProfile{}
	}
	return m.prof.last
}

// SetProfile turns per-subsystem timing on or off. Turning it on resets the accumulators so
// the first frame reported is a whole frame (its counters rebased on where the machine now
// stands, which is what a fresh savestate load needs).
func (m *Machine) SetProfile(on bool) {
	m.Profile = on
	m.prof = profState{}
	if on {
		m.prof.base, m.prof.baseInstr = m.profCounters(), m.CPU.Steps
	}
}
