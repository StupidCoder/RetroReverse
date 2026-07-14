package n3ds

import "testing"

// BenchmarkPoolBarrier is the cost of one fan-out and join with no work in it. The
// GPU does two per draw — the vertex stage and the fill — so Captain Toad pays this
// about 2,000 times a frame, and if it is microseconds then the barrier costs more
// than some of the draws it is splitting.
func BenchmarkPoolBarrier(b *testing.B) {
	for _, n := range []int{2, 4, 8} {
		b.Run(itoa(n), func(b *testing.B) {
			g := &GPU{shEpoch: 1}
			defer func() {
				if g.workers != nil {
					g.workers.close()
				}
			}()
			p := g.pool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.run(n, func(int) {})
			}
		})
	}
}

func itoa(n int) string { return string(rune('0' + n)) }
