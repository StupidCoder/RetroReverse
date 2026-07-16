package gc

// parallel_test.go defends the banded rasteriser, and it is the file that carries the
// correctness claim for it — not the pinned hashes.
//
// A pinned hash says "the same as it was last week". That is a gate against change, and it is
// worth having, but it cannot say the parallel machine agrees with the serial one: both could
// have moved together, and on another host with a different core count they would not even be
// running the same partition. So the claim made here is the narrow one that matters:
// THIS BUILD, RUN IN PARALLEL, COMPUTES WHAT THIS BUILD COMPUTES SERIALLY — every pixel, every
// depth value, every counter.
//
// The claim rests on the partition and not on the scheduler. A band of rows is filled by
// exactly one worker; within a band the triangles are applied in submission order. Which
// worker takes which band is decided by a shared counter and varies run to run, and cannot
// change the answer — only the time. This test is what says so out loud.

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

// framebufferHash fingerprints what the rasteriser actually left behind: the embedded
// framebuffer and its depth buffer, unmediated by a copy to the XFB.
//
// This is the one place the EFB is the RIGHT thing to hash — it is hashed here mid-frame,
// before the pixel engine's copy erases it, which is exactly the case bench_test.go says its
// EFB hash is too weak to catch.
func framebufferHash(g *gpu) string {
	h := sha256.New()
	var w [4]byte
	put := func(v uint32) {
		w[0], w[1], w[2], w[3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
		h.Write(w[:])
	}
	for _, v := range g.EFB {
		put(v)
	}
	for _, v := range g.ZBuf {
		put(v)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// TestParallelMatchesSerial runs the same field twice on two machines from the same state,
// one forced single-threaded, and requires the two to agree exactly.
func TestParallelMatchesSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GPU for several fields twice; -short skips it")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			serial := atState(t, p.state)
			serial.SingleThreaded = true
			defer serial.Close()
			serial.RunFields(benchFields, benchBudget)

			par := atState(t, p.state)
			defer par.Close()
			par.RunFields(benchFields, benchBudget)

			// The picture and the depth buffer, byte for byte.
			if s, pz := framebufferHash(&serial.gpu), framebufferHash(&par.gpu); s != pz {
				t.Errorf("the parallel fill left a different framebuffer than the serial one:\n serial   %s\n parallel %s", s, pz)
			}
			// And the tallies, which are merged after the join and must not depend on who
			// finished first.
			for _, c := range []struct {
				name      string
				got, want int
			}{
				{"pixels written", par.gpu.pixWritten, serial.gpu.pixWritten},
				{"depth-rejected", par.gpu.pixZRej, serial.gpu.pixZRej},
				{"alpha-rejected", par.gpu.pixARej, serial.gpu.pixARej},
				{"tris culled", par.gpu.profCulled, serial.gpu.profCulled},
				{"draws", par.gpu.profDraws, serial.gpu.profDraws},
			} {
				if c.got != c.want {
					t.Errorf("%s: parallel %d, serial %d", c.name, c.got, c.want)
				}
			}
		})
	}
}

// TestParallelIsActuallyUsed is the test that stops the one above from being a tautology.
//
// Every claim in this file is worthless if the fill never fans out — a rasterWorkers that
// returned 1 for everything would pass every hash, every determinism check and every
// serial-agreement check in the package, and the machine would simply be slow. It nearly
// happened: the first probe of this run reported zero parallel draws, because the state needs
// four fields to reach a draw batch and it had been given two.
//
// So: assert that a real field actually reaches the parallel path, and that the threshold
// still rejects the draws it is there to reject.
func TestParallelIsActuallyUsed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GPU for several fields; -short skips it")
	}
	m := atState(t, stateCutscene)
	defer m.Close()
	m.RunFields(benchFields, benchBudget)
	parallel, serial := m.gpu.profParFills, m.gpu.profSerFills

	if parallel == 0 {
		t.Fatalf("no draw was filled in parallel (%d serial): the fan-out is dead and every "+
			"other test in this file is passing vacuously", serial)
	}
	if serial == 0 {
		t.Errorf("every one of %d draws fanned out: the threshold is not rejecting the small "+
			"draws it exists to reject", parallel)
	}
	t.Logf("draws filled in parallel: %d; serially (below the threshold): %d", parallel, serial)
}

// TestTexCanHaltAgreesWithDecoder pins the duplication texCanHalt is built on.
//
// texCanHalt lists the formats decodeTexel knows, so that a draw which could halt is never
// fanned out — a worker goroutine must not halt the machine. Two lists of the same fact drift,
// and the drift is silent in the dangerous direction: if texCanHalt says "safe" about a format
// decodeTexel halts on, a worker halts and races the machine.
//
// So this does not check texCanHalt against a second copy of the list. It asks THE DECODER,
// for every format and every palette format, and requires the two to agree.
func TestTexCanHaltAgreesWithDecoder(t *testing.T) {
	for format := 0; format < 16; format++ {
		for tlutFmt := 0; tlutFmt < 4; tlutFmt++ {
			tx := texState{format: format, width: 8, height: 8, base: 0x1000, tlutFmt: tlutFmt}

			m := testMachine()
			m.gpu.Tlut = make([]byte, 0x80000)
			m.gpu.decodeTexel(m, tx, 0, 0)
			decoderHalts := m.CPU.Halted

			if got := texCanHalt(&tx); got != decoderHalts {
				t.Errorf("format 0x%X, tlutFmt %d: texCanHalt says %v, but the decoder %s",
					format, tlutFmt, got,
					map[bool]string{true: "halts", false: "does not halt"}[decoderHalts])
			}
		}
	}
}
