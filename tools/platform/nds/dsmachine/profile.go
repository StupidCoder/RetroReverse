package dsmachine

import "time"

// Per-subsystem frame timing.
//
// The rule this follows is the one PERFORMANCE.md arrived at the hard way: time only
// the boundaries that are already coarse. A clock read costs more than a fragment, so
// timing every fragment measures the clock. What is timed here is a polygon, a whole
// rasterised frame, a whole 2D compose, and a DMA transfer — each of which is large
// enough that reading the clock either side of it is noise.
//
// The CPU's share is a REMAINDER, not a measurement: whatever the frame took, minus the
// subsystems that were timed. That is honest — the two ARM cores are the frame's
// baseline and everything else is stolen from them — and it is the only way to get
// buckets that are disjoint and sum to the total without pretending to a precision the
// instrumentation does not have.
//
// Counters ride alongside the times, and they are not decoration. Milliseconds alone
// cannot tell a faster rasteriser from a frame that drew less, and that is precisely
// the mistake a profile panel exists to prevent.

// Profile is where one frame's time went.
type Profile struct {
	TotalMs float64

	GeometryMs float64 // the geometry engine: command decode, matrices, lighting, clipping
	RasterMs   float64 // the 3D rasteriser
	ComposeMs  float64 // the two 2D engines
	DMAMs      float64
	CPUMs      float64 // the remainder: both ARM cores

	Commands  int // geometry commands executed
	Polygons  int // polygons the rasteriser was handed
	Fragments int // fragments it produced, kept or killed
	DMAXfers  int
	Frames    uint64
}

// profiler accumulates a frame's timings. It is off unless SetProfile(true).
type profiler struct {
	on bool

	frameStart time.Time
	geom       time.Duration
	raster     time.Duration
	compose    time.Duration
	dma        time.Duration

	polys, frags, xfers int
	cmdStart            int
}

// SetProfile turns per-subsystem timing on or off. It costs a few hundred clock reads a
// frame — nothing beside the per-pixel hook a debugger already installs — so the frame
// debugger simply leaves it on.
func (m *Machine) SetProfile(on bool) {
	m.prof.on = on
	m.prof.reset(m)
}

func (p *profiler) reset(m *Machine) {
	p.frameStart = time.Now()
	p.geom, p.raster, p.compose, p.dma = 0, 0, 0, 0
	p.polys, p.frags, p.xfers = 0, 0, 0
	p.cmdStart = m.gpu3d.count
}

// FrameProfile reports the last completed frame. It is meaningful only with SetProfile
// on; without it the times are zero and the counters still are not, which is deliberate
// — the work a frame did is worth knowing even when nobody timed it.
func (m *Machine) FrameProfile() Profile {
	p := &m.prof
	total := time.Since(p.frameStart)
	ms := func(d time.Duration) float64 { return float64(d.Nanoseconds()) / 1e6 }

	pr := Profile{
		TotalMs:    ms(total),
		GeometryMs: ms(p.geom),
		RasterMs:   ms(p.raster),
		ComposeMs:  ms(p.compose),
		DMAMs:      ms(p.dma),
		Commands:   m.gpu3d.count - p.cmdStart,
		Polygons:   p.polys,
		Fragments:  p.frags,
		DMAXfers:   p.xfers,
		Frames:     m.vid.frames,
	}
	// The CPU is what is left. Clamp at zero rather than report a negative bucket: the
	// clock is not monotonic to the nanosecond across a frame's worth of calls, and a
	// bar chart with a negative segment is a bug report about the profiler, not the
	// machine.
	pr.CPUMs = pr.TotalMs - pr.GeometryMs - pr.RasterMs - pr.ComposeMs - pr.DMAMs
	if pr.CPUMs < 0 {
		pr.CPUMs = 0
	}
	return pr
}
