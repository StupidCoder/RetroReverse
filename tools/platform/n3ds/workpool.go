package n3ds

// workpool.go is the GPU's fan-out primitive: a fixed set of worker goroutines that
// outlive a draw.
//
// The parallel vertex and raster stages used to spawn their workers per draw. Captain
// Toad runs 143 draws a frame, so that was on the order of a thousand goroutine
// creations a frame — cheap each, not free in aggregate, and expensive enough that
// both stages had to refuse to parallelise a small draw at all. Two thirds of the
// draws in a frame fell back to running serially for want of a few microseconds.
//
// A pool costs a channel send per worker instead (tens of nanoseconds), which is what
// lets the thresholds come down far enough for those draws to be worth splitting.
//
// It changes nothing about determinism. The workers are anonymous: what each one does
// is decided by the partition the caller hands it, not by which goroutine picks it up
// (see GPU.fill's row bands, and the vertex stage's index ranges).

import (
	"runtime"
	"sync"
)

// workPool runs the same function on k workers and waits for all of them.
//
// Only the goroutine that owns the machine ever calls run — the machine is
// single-threaded by contract, and the pool is a detail of how one of its stages
// executes, not a licence for anyone else to drive it.
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

// close stops the workers. A machine that is finished with must call it, or its
// goroutines keep it alive forever — which matters in the frame debugger, where
// scratch machines are made and dropped.
func (p *workPool) close() {
	for _, ch := range p.jobs {
		close(ch)
	}
	p.jobs = nil
}

// pool returns the GPU's worker pool, starting it on first use. maxWorkers is the
// only place the machine's parallelism is bounded.
func (g *GPU) pool() *workPool {
	if g.workers == nil {
		g.workers = newWorkPool(maxWorkers)
	}
	return g.workers
}

// Close releases the machine's worker goroutines. A Machine that is simply dropped
// leaks them, so anything that creates machines and throws them away — the frame
// debugger's scratch machine, a test — should call this.
func (m *Machine) Close() {
	if m.gpu != nil && m.gpu.workers != nil {
		m.gpu.workers.close()
		m.gpu.workers = nil
	}
}

// maxWorkers bounds the fan-out. Past about eight the GPU's stages are memory-bound
// and the extra workers mostly contend; and a fixed bound keeps the machine's
// behaviour the same on a laptop and on a build box, which matters when the thing
// being compared between them is a frame time.
var maxWorkers = min(8, runtime.GOMAXPROCS(0))
