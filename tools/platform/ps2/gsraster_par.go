package ps2

// gsraster_par.go parallelises the GS rasteriser — but only WITHIN a single primitive, and
// never across two.
//
// THE PS2 RULE, and why it is not the GameCube's. PS2 drawing is strictly sequential: the
// GS applies primitives in kick order, the alpha blender reads the destination it is about
// to write, and whole passes (Ridge Racer V's blur and pyramid) read back a buffer they
// filled a few primitives earlier. So two different primitives may not run at the same time
// — the second one's inputs are the first one's outputs. The GameCube fans a *batch* of
// triangles out together; here the vertex queue (gsdraw.go) stays serial and only the pixel
// loop of one big sprite or triangle is split.
//
// Splitting one primitive is safe because a primitive writes each pixel it covers exactly
// once (a sprite fills its rectangle, a triangle its half-open bounding box), and the frame
// buffer address is a bijection of (x, y) — so banding the rows hands each worker a disjoint
// set of VRAM words. No two workers touch the same pixel, and within a band the work is
// identical to what the serial loop would do for those rows. DETERMINISM IS IN THE
// PARTITION, NOT THE SCHEDULER: which worker takes which band is a shared counter's whim and
// cannot change the buffer, only the wall time. gsraster_par_test.go pins exactly that.
//
// The one thing that is genuinely shared is the census — the per-pixel tallies plot() keeps
// and the textured/black counts combine() keeps. Those become a per-worker gsStats, merged
// in worker order after the join, so the totals do not depend on who finished first (the
// GameCube's rstats pattern).

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// gsWorkPool is a fixed set of worker goroutines that outlive one primitive's fill. A pool
// (rather than a goroutine per primitive) is the point: a field draws ~1,800 primitives, and
// per-primitive spawning would be the whole cost of the fan-out, forcing the threshold so
// high that nothing would ever split. A channel send is tens of nanoseconds. Copied in shape
// from the GameCube's workPool.
type gsWorkPool struct {
	jobs []chan func(worker int)
	wg   sync.WaitGroup
}

func newGSWorkPool(n int) *gsWorkPool {
	p := &gsWorkPool{jobs: make([]chan func(int), n)}
	for i := range p.jobs {
		ch := make(chan func(int))
		p.jobs[i] = ch
		go func(id int, ch chan func(int)) {
			for f := range ch {
				f(id)
				p.wg.Done()
			}
		}(i, ch)
	}
	return p
}

// run executes f on workers 0..k-1 and returns when all of them have finished. Only the
// goroutine that owns the machine ever calls run — the fan-out is a detail of how one stage
// of a single-threaded machine executes, not a licence for anyone else to drive it.
func (p *gsWorkPool) run(k int, f func(worker int)) {
	if k > len(p.jobs) {
		k = len(p.jobs)
	}
	p.wg.Add(k)
	for i := 0; i < k; i++ {
		p.jobs[i] <- f
	}
	p.wg.Wait()
}

func (p *gsWorkPool) close() {
	for _, ch := range p.jobs {
		close(ch)
	}
	p.jobs = nil
}

// maxRasterWorkers bounds the fan-out. Measured on this repo's twelve-core box; the output
// does not depend on it (the partition decides the answer), only the time.
var maxRasterWorkers = min(12, runtime.GOMAXPROCS(0))

// rasterPoolLive returns the machine's worker pool, starting it on first use. Deliberately
// lazy: a machine restored from a savestate, or one built to answer a question about a
// texture, never starts one.
func (m *Machine) rasterPoolLive() *gsWorkPool {
	if m.rasterPool == nil {
		m.rasterPool = newGSWorkPool(maxRasterWorkers)
	}
	return m.rasterPool
}

// Close releases the machine's rasteriser worker goroutines. A Machine that is simply
// dropped leaks them — the workers hold a reference to the pool and keep it alive forever —
// so anything that makes machines and throws them away (the frame debugger's scratch
// machine, its adapter, the tests) must call this.
func (m *Machine) Close() {
	if m.rasterPool != nil {
		m.rasterPool.close()
		m.rasterPool = nil
	}
}

// gsStats is one worker's tally of where its fragments went. Merged into the GS after the
// join so the census does not race and does not depend on completion order. It carries every
// counter the per-fragment path (plot and the sampler's combine) mutates.
type gsStats struct {
	plotted                  uint64
	rejScissor               uint64
	rejZ                     uint64
	rejAlpha                 uint64
	rejDate                  uint64
	plotNonBlack             [8]uint64
	texBlack, texColor       uint64
	texBlackPSM, texColorPSM [64]uint64
}

func (gs *GS) mergeStats(st *gsStats) {
	gs.plotted += st.plotted
	gs.rejScissor += st.rejScissor
	gs.rejZ += st.rejZ
	gs.rejAlpha += st.rejAlpha
	gs.rejDate += st.rejDate
	for i := range gs.plotNonBlack {
		gs.plotNonBlack[i] += st.plotNonBlack[i]
	}
	gs.texBlack += st.texBlack
	gs.texColor += st.texColor
	for i := range gs.texBlackPSM {
		gs.texBlackPSM[i] += st.texBlackPSM[i]
		gs.texColorPSM[i] += st.texColorPSM[i]
	}
}

// bandRows is how many screen rows one worker fills at a time. The GS frame buffer is
// swizzled, not a flat scanline array, so there is no cache-line argument for a particular
// height (a "row" of pixels is scattered across pages either way); the value is chosen by
// measurement. Finer bands balance the savage unevenness of a field — a handful of
// full-screen blur sprites next to hundreds of tiny ones — at the cost of a few more channel
// sends.
const bandRows = 8

// rasterFanWorkers decides how many workers may fill this primitive. One means "here, on the
// machine's own goroutine".
//
// The first group is FAITHFULNESS, not profit — each is something that observes the fill as
// it happens or that the fill does to the shared machine, and a partition would race on it or
// reorder what it reports:
//
//   - OnGSPixel is the frame debugger's per-pixel provenance hook; it fires per fragment in
//     raster order and a capture must be exactly reproducible.
//   - GSPixelN > 0 is the -gspixel drill-down, which prints per matching write (and arms the
//     sampler's probe) — order is its content.
//   - a droppable target PSM sends plot() down its gs.count() map-write path; keeping such a
//     primitive serial keeps that map write on the machine's own goroutine. These are rare
//     (a 16-bit frame target the machine does not model yet).
//   - SingleThreaded is the caller saying so.
//
// Then the work threshold, which is about profit: a primitive whose bounding box is small
// costs less to fill than the fan-out costs to set up, and one no taller than a single band
// cannot split usefully anyway.
func (gs *GS) rasterFanWorkers(t *gsTarget, x0, x1, y0, y1 int32) int {
	if gs.m == nil || gs.m.SingleThreaded || gs.m.OnGSPixel != nil || gs.m.GSPixelN > 0 {
		return 1
	}
	frameZ := t.psm == psmZ32 || t.psm == psmZ24
	if t.psm != psmCT32 && t.psm != psmCT24 && !frameZ {
		return 1
	}
	w, h := x1-x0, y1-y0
	if w <= 0 || h <= int32(bandRows) {
		return 1
	}
	// The bounding box is what the fill will walk. A sprite covers all of it; a triangle
	// about half — either way this is the cost the fan-out is trying to spread, and small is
	// not worth splitting.
	const minPixels = 8192
	if int(w)*int(h) < minPixels {
		return 1
	}
	return maxRasterWorkers
}

// rasterFill runs body over the row range [y0, y1) — in parallel bands when rasterFanWorkers
// says it is faithful and worth it, otherwise straight through on this goroutine — and merges
// the per-worker stats into the GS afterwards. body fills exactly the rows [yLo, yHi) it is
// handed, accumulating its census into st. x0..x1 is the primitive's horizontal extent, used
// only to size the work for the threshold.
//
// When it fans out it marks the sampler parallel, so the sampler's one shared side effect (a
// one-shot black-texel diagnostic that touches gs.t8Dumped and the note log) stays on the
// serial path. Everything else the sampler does is a read of texture memory.
func (gs *GS) rasterFill(t *gsTarget, smp *gsSampler, x0, x1, y0, y1 int32, body func(yLo, yHi int32, st *gsStats)) {
	workers := gs.rasterFanWorkers(t, x0, x1, y0, y1)
	if workers <= 1 {
		gs.serFills++
		var st gsStats
		body(y0, y1, &st)
		gs.mergeStats(&st)
		return
	}
	if smp != nil {
		smp.parallel = true
	}
	gs.parFills++

	// Never wake more workers than there are bands to hand out: a fill a few bands tall
	// would otherwise pay a channel send + condvar wake for workers that can only find the
	// queue empty and return, and zero a bigger stats slice for nothing. Capping keeps every
	// worker busy and shrinks the per-fill overhead the profile shows on small primitives.
	bands := int(y1-y0+int32(bandRows)-1) / bandRows
	if workers > bands {
		workers = bands
	}

	stats := make([]gsStats, workers)
	var next int32
	gs.m.rasterPoolLive().run(workers, func(w int) {
		st := &stats[w]
		for {
			b := int(atomic.AddInt32(&next, 1)) - 1
			if b >= bands {
				return
			}
			yLo := y0 + int32(b*bandRows)
			yHi := yLo + int32(bandRows)
			if yHi > y1 {
				yHi = y1
			}
			body(yLo, yHi, st)
		}
	})
	// Merged in worker order, so the totals do not depend on who finished first.
	for i := range stats {
		gs.mergeStats(&stats[i])
	}
}
