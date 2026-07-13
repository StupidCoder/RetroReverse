package n3ds

// The parallel vertex stage's correctness claim is not "it matches a hash written
// down last week" — it is "it matches what this same build computes serially". That
// is the property this checks, and it is the one that can actually break.
//
// It has to be said this way because of a discovery made here: `go test -race`
// CHANGES THE FLOATING-POINT RESULT. The race build inhibits FMA contraction, which
// arm64 otherwise applies to the shader's MAD (s1*s2 + s3), so a race build and an
// ordinary build produce frames that differ in the last bit — with the machine
// running single-threaded, where a data race is not even possible. The pinned hashes
// in bench_test.go are therefore a property of the ordinary build, and the gate skips
// under -race rather than reporting a bug that is not there.
//
// What -race IS good for is the thing it is for: proving there is no data race. It
// reports none.

import "testing"

func TestParallelVertexMatchesSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GPU for several frames twice")
	}

	ms := toadAtScene(t)
	ms.SingleThreaded = true
	ms.RunFrames(benchFrames, benchBudget)
	sv, ss, sc := hashes(ms)

	mp := toadAtScene(t)
	mp.SingleThreaded = false
	mp.RunFrames(benchFrames, benchBudget)
	pv, ps, pc := hashes(mp)

	for _, c := range []struct{ name, serial, parallel string }{
		{"vram", sv, pv},
		{"screens", ss, ps},
		{"cpu", sc, pc},
	} {
		if c.serial != c.parallel {
			t.Errorf("%s: the parallel vertex stage does not agree with the serial one\n serial   %s\n parallel %s",
				c.name, c.serial, c.parallel)
		}
	}
}
