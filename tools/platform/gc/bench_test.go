package gc

// bench_test.go is the harness every performance change reports against, and the gate that
// stops a "3× faster" claim from quietly being a "3× faster and subtly wrong" one.
//
// The oracle's output is ground truth: other code in this repository is verified against it.
// So an optimisation that moves one pixel has not sped the oracle up, it has broken it.
// Every claim is therefore made twice — once in milliseconds, once in hashes — and a change
// that improves the first while moving the second is a regression, not a win.
//
// WHAT IS HASHED, AND WHY EACH ONE. This needed more care here than on the 3DS, because of
// one fact about a GameCube: THE EMBEDDED FRAMEBUFFER IS ERASED AS THE FRAME IS DELIVERED.
// The pixel engine's copy out to the external framebuffer clears the EFB behind it (see
// copyDisplay), so a run that ends after a copy finds the EFB holding flat clear-colour. A
// hash of it would pass no matter what the pipe drew. So:
//
//	ram    all 24 MiB of main memory. THE XFB IS IN HERE — the copy writes the delivered
//	       picture into RAM, so this is the durable frame, and it catches every CPU-side
//	       trajectory change on the way as well. The most valuable pin of the four.
//	xfb    the decoded picture VI is scanning out. Redundant with ram by construction, and
//	       kept separate anyway so a failure can say "the picture moved" rather than only
//	       "a byte moved" — and so the thing a human looks at has a hash of its own.
//	aram   the 16 MiB the DSP addresses. It is exactly what diverged when one lost DSP
//	       interrupt killed every savestate in the tree — but measure before trusting it:
//	       it is 92.66% non-zero and BYTE-IDENTICAL between both states pinned below, i.e.
//	       the sound bank is DMA'd in once and never touched again. So it is a weak
//	       defender in practice, kept because it costs nothing and would catch an ARAM DMA
//	       regression that nothing else here looks at.
//	cpu    the Gekko's register file and clocks: the trajectory no picture can see. Both
//	       halves of every FPR are in it — PS1 is carried by instructions that have never
//	       heard of it, and a hash that dropped it would miss precisely that bug class.
//
// The EFB and the depth buffer are hashed too, because they are free, but they are WEAK
// DEFENDERS for the reason above: they only say anything when a run ends mid-frame.
//
// The pins are a gate against *change*, not a proof of correctness — the picture they defend
// was judged by eye first, and TestFrameHashes writes the frame out on failure so the next
// person judges it by eye too rather than trusting the hash ([[pinned-hash-guards-change]]).
//
// They are also machine-specific: Go permits FMA contraction, so a float result can differ
// between arm64 and amd64. These were pinned on darwin/arm64. A mismatch on another host is
// not necessarily a bug — but it must be explained before it is re-pinned.

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The recipe. Neither the disc image (copyright) nor the savestates (work/ is ignored) are
// committed, so everything here skips when they are absent.
const (
	lmImage = "../../../games/luigis-mansion-gc/image/Luigi's Mansion (USA).iso"

	// The intro cutscene: Luigi and his flashlight in the twisted forest. It is the field
	// the profiler's own header counts, so every number this harness reports is comparable
	// to the ones written down there — ~12,770 GX commands, ~9,774 draws, ~1.4M fragments.
	stateCutscene = "../../../games/luigis-mansion-gc/work/states/intro-cutscene.state"

	// The shadow state earns its place as a SECOND pin rather than a spare: it is the only
	// state in the set that exercises the TEV's swap tables and its compare mode. Those are
	// precisely the machinery a "decode the TEV once per draw" refactor is most likely to
	// break, and a gate that watched only the cutscene would let that land green.
	stateShadow = "../../../games/luigis-mansion-gc/work/states/shadow.state"

	// benchFields is what the pins below were taken over. It is part of the recipe: change
	// it and every hash changes with it. Fields, not flips — see RunFields.
	benchFields = 4

	// benchBudget caps a field that never ends, so a hang is a failed gate rather than a
	// wedged test run. One field is fieldInstructions; this is generous.
	benchBudget = 200_000_000
)

// gatePin is one state's pinned output after benchFields fields.
type gatePin struct {
	name  string
	state string
	ram   string
	xfb   string
	aram  string
	cpu   string
}

// The pins, taken on 2026-07-16 (darwin/arm64) from the frames this recipe reproduces —
// which were LOOKED AT, both of them, before these were written down:
//
//	intro-cutscene  Luigi and his flashlight in the twisted forest: the trunks lit and
//	                textured, the cone of the torch on the ground, his cap and overalls
//	                the right colours. 9,774 draws, 737,611 fragments.
//	shadow          the YOUR MANSION map flyer held up in Luigi's gloved hands, the
//	                mansion and the rainbow on it, the HERE arrow over the house — and
//	                the shadow the map casts, which is the whole reason this state is
//	                here. 18,350 draws, 599,857 fragments, 0 alpha-rejected.
//
// If one of these moves, open the PNG the failure prints and JUDGE THE PICTURE AGAIN
// before touching this block. A pinned hash guards change, not correctness.
var gatePins = []gatePin{
	{
		name:  "intro-cutscene",
		state: stateCutscene,
		ram:   "cfe74635ea25a21e8c8217be3fe3ff7bae90a2ae9f1d638e770284d45046b36d",
		xfb:   "8cc397e1a4fea02be5ce793a6b91b4ffac4ace715b2bc25772727f0be755488c",
		aram:  "85b317d31fbf4c465840d9ff3803a522253387d95af6c5a4511e5094ade6e558",
		cpu:   "a3858ea576743ae00c5fd1219de166fa195a15b23a67068bfb03d462dfd71c50",
	},
	{
		name:  "shadow",
		state: stateShadow,
		ram:   "f158bd9b06de10367c648dfc1b46fd309f19dd07d44a3f3702ea5f7935b66804",
		xfb:   "34525c0d2dcdab1bd2b85bdba8ca9ee09f69d131d5a0700198a3508993e7ed6b",
		aram:  "85b317d31fbf4c465840d9ff3803a522253387d95af6c5a4511e5094ade6e558",
		cpu:   "a236e2d4f92a2818ce769d79b92f0027d300c84cebdff8d51a58b38404ec9cd0",
	},
}

// atState restores the machine to one of the recipe's savestates, ready to draw.
func atState(tb testing.TB, state string) *Machine {
	tb.Helper()
	disc, err := Open(lmImage)
	if err != nil {
		tb.Skip("Luigi's Mansion image not present (game images are not committed)")
	}
	tb.Cleanup(func() { disc.Close() })

	m, err := NewMachine(disc)
	if err != nil {
		tb.Fatal(err)
	}
	if err := m.LoadStateFile(state); err != nil {
		tb.Skipf("no savestate at %s (work/ is not committed): %v", filepath.Base(state), err)
	}
	return m
}

// gateHashes fingerprints the machine's output and its CPU trajectory.
func gateHashes(m *Machine) (ram, xfb, aram, cpu string) {
	sum := func(f func(h *hasher)) string {
		h := &hasher{Hash: sha256.New()}
		f(h)
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	ram = sum(func(h *hasher) { h.Write(m.RAM) })
	aram = sum(func(h *hasher) { h.Write(m.ARAM) })

	xfb = sum(func(h *hasher) {
		img, err := m.RenderXFB()
		if err != nil {
			// Not a failure: a run that never programmed VI has no picture, and saying so
			// is a stable fingerprint of exactly that.
			h.Write([]byte("no xfb: " + err.Error()))
			return
		}
		h.Write(img.Pix)
	})

	cpu = sum(func(h *hasher) {
		c := m.CPU
		for _, r := range c.GPR {
			h.u64(uint64(r))
		}
		// BOTH halves of every FPR. PS1 is the one a scalar-only hash would miss, and it
		// is the one that draws wrong geometry when an instruction forgets to carry it.
		for _, f := range c.FPR {
			h.f64(f.PS0)
			h.f64(f.PS1)
		}
		for _, v := range []uint32{c.PC, c.LR, c.CTR, c.CR, c.XER, c.MSR, c.FPSCR,
			c.HID0, c.HID1, c.HID2, c.HID4, c.WPAR, c.DMAU, c.DMAL, c.L2CR,
			c.SRR0, c.SRR1, c.DSISR, c.DAR, c.SDR1, c.PVR, c.DEC, c.ReserveAddr} {
			h.u64(uint64(v))
		}
		for _, v := range c.GQR {
			h.u64(uint64(v))
		}
		for _, v := range c.SPRG {
			h.u64(uint64(v))
		}
		for _, v := range c.SR {
			h.u64(uint64(v))
		}
		for _, b := range [][4][2]uint32{c.IBAT, c.DBAT} {
			for _, e := range b {
				h.u64(uint64(e[0]))
				h.u64(uint64(e[1]))
			}
		}
		h.u64(c.TB)
		h.bool(c.Reserved)
		h.bool(c.Halted)
		h.u64(c.Steps)
		// The locked cache is real data, not a cache: it is 16 KiB of scratchpad the game
		// addresses directly, and its contents are part of the machine.
		h.Write(c.LC.Data[:])

		// The machine's own clocks. The video field counter is what paces every device
		// here, so a run that drifted by one field would show up in nothing else.
		h.u64(m.Instrs)
		h.u64(m.vi.Field)
		h.u64(uint64(m.vi.Counter))
	})
	return ram, xfb, aram, cpu
}

// hasher is sha256 with the fixed-width writers the fingerprint needs. Little-endian and
// eight bytes wide throughout, so a value's width cannot change what it hashes to.
type hasher struct{ hash.Hash }

func (h *hasher) u64(v uint64) {
	var w [8]byte
	binary.LittleEndian.PutUint64(w[:], v)
	h.Write(w[:])
}

// f64 hashes a float by its BITS, not its value: two NaNs that compare unequal must still
// fingerprint the same way, and a negative zero must not fingerprint as a positive one.
func (h *hasher) f64(v float64) { h.u64(math.Float64bits(v)) }

func (h *hasher) bool(b bool) {
	if b {
		h.u64(1)
		return
	}
	h.u64(0)
}

// gateCounters snapshots the GPU's per-run tallies, so a benchmark can report the work done
// as well as the time taken. FRAGMENTS PER MILLISECOND is the number that says whether a
// change worked; wall time alone cannot distinguish a faster rasteriser from a field that
// drew less.
type gateCounters struct {
	cmds, draws, culled, frags, zRej, aRej int
}

func snapGate(m *Machine) gateCounters {
	c := m.profCounters()
	return gateCounters{c.cmds, c.draws, c.culled, c.frags, c.zRejected, c.aRejected}
}

func (c gateCounters) sub(o gateCounters) gateCounters {
	return gateCounters{c.cmds - o.cmds, c.draws - o.draws, c.culled - o.culled,
		c.frags - o.frags, c.zRej - o.zRej, c.aRej - o.aRej}
}

// TestFrameHashes is the gate. It runs each recipe and compares the machine's output
// against the pins.
//
// On a mismatch it writes the frame it actually produced to a PNG and names the path,
// because a hash says only "different" — and the last time this repository trusted a hash
// without looking at what it defended, the hash was defending a cropped frame.
func TestFrameHashes(t *testing.T) {
	if testing.Short() {
		t.Skip("the frame gate runs the GPU for several fields; -short skips it")
	}
	if raceBuild {
		// Not a bug, and not a race: the race build inhibits FMA contraction, so the TEV's
		// blend rounds differently and the frame differs in its last bits from the one
		// these hashes were pinned on — with the machine running single-threaded, where a
		// data race cannot exist. What -race is for here is TestFrameDeterminism, which
		// compares a build against ITSELF.
		t.Skip("-race changes the floating-point result (FMA contraction); the pins are for the ordinary build")
	}

	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			m := atState(t, p.state)

			before := snapGate(m)
			start := time.Now()
			res := m.RunFields(benchFields, benchBudget)
			elapsed := time.Since(start)
			did := snapGate(m).sub(before)

			if m.CPU.Halted {
				t.Fatalf("machine halted during the gate at 0x%08X: %s", m.CPU.PC, m.CPU.HaltReason)
			}
			t.Logf("%d fields: %s (%s/field), %d instructions — %s",
				benchFields, elapsed.Round(time.Millisecond),
				(elapsed / benchFields).Round(time.Millisecond), res.Steps, res.Reason)
			t.Logf("  %d gx commands, %d draws, %d tris culled, %d fragments drawn, %d depth-rejected, %d alpha-rejected",
				did.cmds, did.draws, did.culled, did.frags, did.zRej, did.aRej)
			if ms := elapsed.Seconds() * 1000; ms > 0 {
				t.Logf("  %.0f fragments/ms", float64(did.frags)/ms)
			}

			ram, xfb, aram, cpu := gateHashes(m)
			for _, c := range []struct{ name, got, want string }{
				{"ram", ram, p.ram},
				{"xfb", xfb, p.xfb},
				{"aram", aram, p.aram},
				{"cpu", cpu, p.cpu},
			} {
				if c.want == "PIN_ME" {
					t.Errorf("%s hash is unpinned; it is %s", c.name, c.got)
					continue
				}
				if c.got != c.want {
					t.Errorf("%s hash = %s, want %s", c.name, c.got, c.want)
				}
			}
			if t.Failed() {
				writeGateFrame(t, m, filepath.Join(os.TempDir(), "gc-gate-fail-"+p.name+".png"))
			}
		})
	}
}

// writeGateFrame puts the picture the machine actually produced on disk and says so loudly.
func writeGateFrame(t *testing.T, m *Machine, path string) {
	t.Helper()
	img, err := m.RenderXFB()
	if err != nil {
		t.Logf("(could not render the frame out: %v)", err)
		return
	}
	f, err := os.Create(path)
	if err != nil {
		t.Logf("(could not write the frame out: %v)", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Logf("(could not encode the frame out: %v)", err)
		return
	}
	t.Logf("LOOK AT THE FRAME before re-pinning: %s", path)
}

// TestFrameDeterminism checks the property the gate rests on: the same state, run the same
// way, twice, produces the same machine. If this fails, no pinned hash above means anything
// — the machine has a nondeterminism (a map iterated somewhere, a wall clock consulted) and
// that is the bug to fix first.
func TestFrameDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the GPU for several fields twice; -short skips it")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			var first [4]string
			for run := 0; run < 2; run++ {
				m := atState(t, p.state)
				m.RunFields(benchFields, benchBudget)
				ram, xfb, aram, cpu := gateHashes(m)
				got := [4]string{ram, xfb, aram, cpu}
				if run == 0 {
					first = got
					continue
				}
				for i := range got {
					if got[i] != first[i] {
						t.Errorf("%s differs between two identical runs: %s vs %s",
							[4]string{"ram", "xfb", "aram", "cpu"}[i], got[i], first[i])
					}
				}
			}
		})
	}
}

// BenchmarkField reports ns/field and — because a field is only comparable to another field
// that drew the same thing — the work it did alongside.
//
// It re-restores the savestate for each measured batch: b.N fields of a game that is
// animating would otherwise drift into a different scene, and the number would stop meaning
// what it says.
func BenchmarkField(b *testing.B) {
	m := atState(b, stateCutscene)
	snap := m.SaveState()

	var did gateCounters
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := m.LoadState(snap); err != nil {
			b.Fatal(err)
		}
		before := snapGate(m)
		b.StartTimer()

		m.RunFields(1, benchBudget)

		b.StopTimer()
		did = snapGate(m).sub(before)
		b.StartTimer()
	}
	b.StopTimer()

	b.ReportMetric(float64(did.draws), "draws/field")
	b.ReportMetric(float64(did.frags), "frags/field")
	b.ReportMetric(float64(did.zRej), "zrej/field")
}

// TestWriteGateFrame writes out what the recipe currently draws, so the picture the pins
// defend can be looked at on purpose rather than only on failure. Off by default: set
// GC_GATE_PNG to a directory to run it.
func TestWriteGateFrame(t *testing.T) {
	dir := os.Getenv("GC_GATE_PNG")
	if dir == "" {
		t.Skip("set GC_GATE_PNG=<dir> to write the gate's frames out")
	}
	for _, p := range gatePins {
		m := atState(t, p.state)
		m.RunFields(benchFields, benchBudget)
		writeGateFrame(t, m, filepath.Join(dir, "gc-gate-"+p.name+".png"))
	}
}
