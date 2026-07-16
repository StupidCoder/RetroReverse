package gc

// workpool.go is the graphics unit's fan-out primitive: a fixed set of worker goroutines
// that outlive a draw.
//
// It is a PRECONDITION here rather than a refinement, and the arithmetic says why. The 3DS
// spawned its workers per draw across 143 draws a frame — about a thousand goroutine
// creations, cheap each and not free in aggregate, and expensive enough that two thirds of
// its draws refused to split at all for want of a few microseconds. THIS MACHINE ISSUES
// 9,774 DRAWS IN A SINGLE FIELD: sixty-eight times as many. Per-draw spawning would not be
// a tax here, it would be the whole cost, and the threshold needed to hide it would reject
// nearly every draw in the frame.
//
// A pool costs a channel send per worker instead — tens of nanoseconds — which is what lets
// the threshold come down far enough for an ordinary draw to be worth splitting.
//
// It changes nothing about determinism. The workers are anonymous: what each one does is
// decided by the partition the caller hands it, not by which goroutine picks it up. See
// gpu.fill's row bands.

import (
	"runtime"
	"sync"
)

// workPool runs the same function on k workers and waits for all of them.
//
// Only the goroutine that owns the machine ever calls run. The machine is single-threaded by
// contract, and the pool is a detail of how one of its stages executes — not a licence for
// anyone else to drive it.
type workPool struct {
	jobs []chan func(worker int)
	wg   sync.WaitGroup
}

func newWorkPool(n int) *workPool {
	p := &workPool{jobs: make([]chan func(int), n)}
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

// run executes f on workers 0..k-1 and returns when all of them have finished.
func (p *workPool) run(k int, f func(worker int)) {
	if k > len(p.jobs) {
		k = len(p.jobs)
	}
	p.wg.Add(k)
	for i := 0; i < k; i++ {
		p.jobs[i] <- f
	}
	p.wg.Wait()
}

// close stops the workers.
func (p *workPool) close() {
	for _, ch := range p.jobs {
		close(ch)
	}
	p.jobs = nil
}

// pool returns the graphics unit's worker pool, starting it on first use.
//
// It is deliberately lazy: the pool is not part of the machine's state, it is a detail of
// how a stage executes. A machine restored from a savestate, or one built to answer a
// question about a texture, never starts one.
func (g *gpu) pool() *workPool {
	if g.workers == nil {
		g.workers = newWorkPool(maxWorkers)
	}
	return g.workers
}

// Close releases the machine's worker goroutines.
//
// A Machine that is simply dropped leaks them — the goroutines hold a reference to the pool
// and keep it alive forever — so anything that makes machines and throws them away must call
// this. In this repository that is the frame debugger's scratch machine and its adapter's
// own Close, and the tests.
func (m *Machine) Close() {
	if m.gpu.workers != nil {
		m.gpu.workers.close()
		m.gpu.workers = nil
	}
}

// maxWorkers bounds the fan-out.
//
// The 3DS caps this at eight for two reasons, and only one of them survives measurement here.
// The first — past about eight these stages are memory-bound and the extra workers mostly
// contend — is simply not what this machine does: on twelve cores, 4/8/10/12 workers measured
// 1.335/1.276/1.26/1.254 s over the shadow scene, still improving at twelve.
//
// The second reason was that a fixed bound keeps a frame time comparable between a laptop and
// a build box. That one is DELIBERATELY TRADED AWAY for the ~2%, because it was never worth
// much: the rule that actually protects a measurement is to A/B on one machine in one sitting,
// and a frame time is not comparable across hosts anyway. What is comparable across hosts is
// the OUTPUT, and that does not depend on the worker count at all — the partition decides the
// answer, not the scheduler, which is exactly what TestParallelMatchesSerial pins.
var maxWorkers = min(12, runtime.GOMAXPROCS(0))
