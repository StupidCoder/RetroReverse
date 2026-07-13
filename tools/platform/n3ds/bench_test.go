package n3ds

// bench_test.go is the harness every performance change reports against, and the
// gate that stops a "3× faster" claim from quietly being a "3× faster and subtly
// wrong" one.
//
// The oracle's output is ground truth: other code in this repository is verified
// against it. So an optimisation that moves one pixel has not sped the oracle up,
// it has broken it. Every claim made here is therefore made twice — once in
// nanoseconds, once in hashes — and a change that improves the first while moving
// the second is a regression, not a win.
//
// What is hashed, and why each one:
//
//	vram    the colour buffer, the depth buffer and the shadow map all live in
//	        VRAM. Hashing the whole region catches a moved pixel wherever the
//	        registers happened to point at the end of the run.
//	screens the two framebuffers the GSP last presented — the picture a human
//	        would look at, which lives in the linear heap, not in VRAM, and is
//	        reached only through a DisplayTransfer.
//	cpu     every thread's register file plus the machine's instruction and tick
//	        clocks: the CPU-side trajectory, which a GPU-only hash cannot see.
//
// The pins below are a gate against *change*, not a proof of correctness — the
// picture they defend was judged by eye first, and TestFrameHashes writes the
// frame out on failure so the next person judges it by eye too rather than
// trusting the hash ([[pinned-hash-guards-change]]).
//
// They are also machine-specific: Go permits FMA contraction, so a float result
// can differ between arm64 and amd64. These were pinned on darwin/arm64. A
// mismatch on another host is not necessarily a bug — but it must be explained
// before it is re-pinned.

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The Captain Toad opening stage: the only scene in the repository that exercises
// the whole PICA pipeline (vertex shader, depth test, fragment lighting, TEV,
// shadow map). Neither the image (copyright) nor the savestate (work/ is ignored)
// is committed, so everything here skips when they are absent.
const (
	toadImage = "../../../games/captain-toad-treasure-tracker-3ds/image/Captain Toad - Treasure Tracker (Europe) (En,Fr,De,Es,It,Nl).cci"
	toadState = "../../../games/captain-toad-treasure-tracker-3ds/work/states/toadstereo4.state"

	// benchFrames is what the pinned hashes below were taken over. It is part of
	// the recipe: change it and every hash changes with it.
	benchFrames = 4

	// benchBudget caps a frame that never ends (a hang shows up as a failed gate,
	// not a wedged test run). ~4.5M instructions is one frame's VBlank period.
	benchBudget = 200_000_000
)

// The pinned state after benchFrames frames from toadstereo4.state. Taken on
// 2026-07-13 (darwin/arm64) from the frame reproduced below — which was looked at,
// on both screens, before these were written down: the opening-stage diorama, lit,
// with Toad on the steps, the star on the roof and the ivy on the tower. That is
// what these three lines defend. If one of them moves, open the PNG the failure
// prints and judge the picture again before touching this block.
const (
	wantVRAM    = "082cc9b4ef4d7b2b739b33554119c1168e9ebda3164c36d1476caf466374025f"
	wantScreens = "5152b1e41df25df83a3753da21fc19f879bd0427fbb43ee51b74639cd17fb03c"
	wantCPU     = "55b5cf4f4101c6b25f577a964bff75f379eec38871c13695cc1443f148eef8ac"
)

// toadAtScene restores the machine to the opening stage, ready to draw.
func toadAtScene(tb testing.TB) *Machine {
	tb.Helper()
	img, err := os.ReadFile(toadImage)
	if err != nil {
		tb.Skip("Captain Toad image not present (game images are not committed)")
	}
	m, err := NewMachine(img)
	if err != nil {
		tb.Fatal(err)
	}
	if err := m.LoadState(toadState); err != nil {
		tb.Skip("no savestate at the opening stage (work/ is not committed): " + err.Error())
	}
	return m
}

// hashes fingerprints the machine's output and its CPU trajectory.
func hashes(m *Machine) (vram, screens, cpu string) {
	h := sha256.New()
	for _, r := range m.regions {
		if r.name == "vram" {
			h.Write(r.data)
		}
	}
	vram = fmt.Sprintf("%x", h.Sum(nil))

	h.Reset()
	for _, screen := range []string{"top", "bottom"} {
		img := m.Framebuffer(screen)
		if img == nil {
			h.Write([]byte("absent:" + screen))
			continue
		}
		h.Write(img.Pix)
	}
	screens = fmt.Sprintf("%x", h.Sum(nil))

	h.Reset()
	var w [8]byte
	put := func(v uint64) {
		binary.LittleEndian.PutUint64(w[:], v)
		h.Write(w[:])
	}
	putb := func(b bool) {
		if b {
			put(1)
		} else {
			put(0)
		}
	}
	put(m.instrs)
	put(m.tick)
	put(m.vblankCount)
	// The live thread's registers are in CPU; every other thread's are in its
	// saved context. Threads are held in creation order, which is stable.
	for _, t := range m.threads {
		ctx := &t.ctx
		if t == m.curThread {
			ctx = m.CPU
		}
		for _, r := range ctx.R {
			put(uint64(r))
		}
		putb(ctx.N)
		putb(ctx.Z)
		putb(ctx.C)
		putb(ctx.V)
		putb(ctx.Thumb)
		put(uint64(ctx.Mode))
		put(ctx.Instrs)
		put(uint64(t.state))
	}
	cpu = fmt.Sprintf("%x", h.Sum(nil))
	return vram, screens, cpu
}

// counters snapshots the GPU's per-run tallies, so a benchmark can report the
// work done as well as the time taken. Fragments per millisecond is the number
// that says whether a change worked; wall time alone cannot distinguish a faster
// rasteriser from a frame that drew less.
type counters struct {
	draws, pixels, depthKilled, culled, rejected int
	shadowWrites, shadowSamples                  int
}

func snapCounters(g *GPU) counters {
	return counters{g.Draws, g.PixelsDrawn, g.DepthKilled, g.CulledTris, g.RejectedTris, g.ShadowWrites, g.ShadowSamples}
}

func (c counters) sub(o counters) counters {
	return counters{c.draws - o.draws, c.pixels - o.pixels, c.depthKilled - o.depthKilled,
		c.culled - o.culled, c.rejected - o.rejected, c.shadowWrites - o.shadowWrites, c.shadowSamples - o.shadowSamples}
}

// TestFrameHashes is the gate. It runs the recipe and compares the machine's
// output against the pins.
//
// On a mismatch it writes the frame it actually produced to a PNG and names the
// path, because a hash says only "different" — and the last time this repository
// trusted a hash without looking at what it defended, the hash was defending a
// cropped frame.
func TestFrameHashes(t *testing.T) {
	if testing.Short() {
		t.Skip("the frame gate runs the GPU for several frames; -short skips it")
	}
	m := toadAtScene(t)

	before := snapCounters(m.gpu)
	start := time.Now()
	ran := m.RunFrames(benchFrames, benchBudget)
	elapsed := time.Since(start)
	did := snapCounters(m.gpu).sub(before)

	if r := m.HaltReason(); r != "" {
		t.Fatalf("machine halted during the gate: %s", r)
	}
	t.Logf("%d frames: %s (%s/frame), %d instructions", benchFrames, elapsed.Round(time.Millisecond),
		(elapsed / benchFrames).Round(time.Millisecond), ran)
	t.Logf("per frame: %d draws, %d fragments drawn, %d depth-killed, %d tris culled, %d shadow writes",
		did.draws/benchFrames, did.pixels/benchFrames, did.depthKilled/benchFrames,
		did.culled/benchFrames, did.shadowWrites/benchFrames)

	vram, screens, cpu := hashes(m)
	for _, c := range []struct{ name, got, want string }{
		{"vram", vram, wantVRAM},
		{"screens", screens, wantScreens},
		{"cpu", cpu, wantCPU},
	} {
		if c.got != c.want {
			t.Errorf("%s hash = %s, want %s", c.name, c.got, c.want)
		}
	}
	if t.Failed() {
		base := filepath.Join(os.TempDir(), "n3ds-gate-fail")
		if err := m.Screenshot(base); err != nil {
			t.Logf("(could not write the frame out: %v)", err)
		} else {
			t.Logf("LOOK AT THE FRAME before re-pinning: %s_top.png / %s_bottom.png", base, base)
		}
	}
}

// TestFrameDeterminism checks the property the gate rests on: the same state, run
// the same way, twice, produces the same machine. If this fails, no pinned hash
// above means anything — the machine has a nondeterminism (a map iterated in the
// HLE, a wall clock consulted) and that is the bug to fix first.
func TestFrameDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GPU for several frames twice; -short skips it")
	}
	var first [3]string
	for run := 0; run < 2; run++ {
		m := toadAtScene(t)
		v, s, c := "", "", ""
		m.RunFrames(benchFrames, benchBudget)
		v, s, c = hashes(m)
		if run == 0 {
			first = [3]string{v, s, c}
			continue
		}
		for i, got := range []string{v, s, c} {
			if got != first[i] {
				t.Errorf("%s differs between two identical runs: %s vs %s",
					[3]string{"vram", "screens", "cpu"}[i], got, first[i])
			}
		}
	}
}

// BenchmarkFrame reports ns/frame, and — because a frame is only comparable to
// another frame that drew the same thing — the work it did alongside it.
//
// It re-restores the savestate for each measured batch: b.N frames of a game that
// is animating would otherwise drift into a different scene, and the number would
// stop meaning what it says.
func BenchmarkFrame(b *testing.B) {
	m := toadAtScene(b)
	snap := m.SnapshotState()

	var did counters
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := m.RestoreState(snap); err != nil {
			b.Fatal(err)
		}
		before := snapCounters(m.gpu)
		b.StartTimer()

		m.RunFrames(1, benchBudget)

		b.StopTimer()
		did = snapCounters(m.gpu).sub(before)
		b.StartTimer()
	}
	b.StopTimer()

	b.ReportMetric(float64(did.draws), "draws/frame")
	b.ReportMetric(float64(did.pixels), "frags/frame")
	b.ReportMetric(float64(did.depthKilled), "zkilled/frame")
}

// TestWriteGateFrame writes out what the recipe currently draws, so the picture
// the pins defend can be looked at on purpose rather than only on failure. Off by
// default: set N3DS_GATE_PNG to a path base to run it.
func TestWriteGateFrame(t *testing.T) {
	out := os.Getenv("N3DS_GATE_PNG")
	if out == "" {
		t.Skip("set N3DS_GATE_PNG=<base> to write the gate's frame out")
	}
	m := toadAtScene(t)
	m.RunFrames(benchFrames, benchBudget)
	if err := m.Screenshot(out); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s_top.png / %s_bottom.png", out, out)
}
