package dc

// bench_test.go is the pinned-hash gate, in the shape gc/bench_test.go
// established: every claim about this machine is made twice, once in
// milliseconds and once in hashes, and a change that improves the first
// while moving the second is a regression until the frame has been LOOKED
// AT again. The pins guard change, not correctness — the pictures they
// defend were judged by eye first, and a failure writes the frame out so
// the next person judges it by eye too.
//
// What is hashed, and why each one:
//
//	ram   all 16 MiB of main memory: the durable record of the CPU-side
//	      trajectory, and where every parsed input and game state lives.
//	fb    the picture the CRTC would scan out (RenderFB), decoded — the
//	      thing a human judges. A machine with no configured video hashes
//	      the error string, so "no picture" is a stable fingerprint too.
//	vram  the full 8 MiB in the 64-bit-path layout: textures, both
//	      framebuffers, the TA's working set. Catches a renderer change
//	      that fb alone would miss (the back buffer, an intermediate).
//	cpu   the SH-4's whole programmer's model plus the machine's clocks,
//	      and the raw on-chip register file in sorted order.
//	snd   the AICA side: sound RAM, the ARM7's registers and instruction
//	      count, the timers. The sound processor is a real CPU here and
//	      its divergence would otherwise be invisible until it mattered.
//
// The pins are machine-specific (Go permits FMA contraction, so float
// results can differ between arm64 and amd64); these were pinned on
// darwin/arm64. A mismatch on another host must be explained before it is
// re-pinned.

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// The recipe. Neither the disc image (copyright) nor the savestates
// (work/ is ignored) are committed, so everything here skips when they
// are absent.
const (
	ctImage = "../../../games/crazy-taxi-dc/image/Crazy Taxi (US).cue"

	// The no-VMU warning screen: the first picture this machine ever drew.
	// Sprites, twiddled textures, the 2D submission path by ch2 DMA.
	stateWarning = "../../../games/crazy-taxi-dc/work/states/warn-fresh.st"

	// The attract-mode city: the black convertible on the boulevard,
	// storefronts and billboards. The 3D path — store-queue FIFO
	// submission, intensity colours, VQ and mipmapped textures, the
	// z-buffer, perspective-correct UVs, both VRAM windows.
	stateCity = "../../../games/crazy-taxi-dc/work/states/city-fresh.st"

	// The drive: in-game behind the yellow cab — sky, palms, traffic, the
	// HUD's rolling counters. The deepest state the machine reaches and the
	// widest coverage of the pipeline in one frame.
	stateDrive = "../../../games/crazy-taxi-dc/work/states/drive.st"

	// benchFields is what the pins were taken over; change it and every
	// hash changes with it.
	benchFields = 30

	// benchBudget caps a field that never ends, so a hang is a failed
	// gate rather than a wedged test run.
	benchBudget = 200_000_000
)

// gatePin is one state's pinned output after benchFields fields.
type gatePin struct {
	name  string
	state string
	ram   string
	fb    string
	vram  string
	cpu   string
	snd   string
}

// The pins, taken 2026-07-30 (darwin/arm64) from frames that were looked
// at first:
//
//	warning  the no-VMU warning screen: the CRAZY TAXI logo with its flame
//	         swoosh and taxi-ball emblem in chrome lettering, the Japanese
//	         save-file text readable around it.
//	city     the attract's PRESS START BUTTON scene: the yellow cab riding
//	         a green car-transporter past the park, a purple pickup on the
//	         road, the prompt overlaid.
//	drive    IN THE GAME, behind Axel's cab on the boulevard: blue sky and
//	         clouds, palms, traffic, "38 game time", the TOTAL EARNED
//	         counters, the D gear indicator — reached from cold boot by the
//	         game's own inputs (start at the warning and title, A through
//	         MODE SELECTION / ARCADE RULES / driver select).
//
// If one of these moves, open the PNG the failure prints and JUDGE THE
// PICTURE AGAIN before touching this block.
//
// RE-PINNED the same day, every hash, and the reason is the repository's own
// lesson written down in [[savestate-lineage-carries-old-model]]: the first
// pins were cut from savestates descended from an old build's boot whose
// texture-path DMA never wrote the 1AAxxx-1CFxxx region — the logo screen
// had no logo and the attract's cars were untextured black. The states were
// re-cut from a clean boot (which also surfaced the DEVINFO contract: the
// game only accepts a pad whose FT word and device-status layout are the
// hardware's own — see maple.go), and both frames were judged again before
// these hashes were written.
var gatePins = []gatePin{
	{
		name:  "warning",
		state: stateWarning,
		ram:   "2e1e2ff681a1fd3077d87ebecc1a8fa2db26dc0c4bd16ca0e85f3e9d4134f412",
		fb:    "9934acc8748b26a9574e12bb40de7b45600513d3f0b08a8071161d555026f398",
		vram:  "87501550438d5c6916a66639963ecd311b51912bcf2ef35180e7f28850b32510",
		cpu:   "6518dc3b4f380bc465091f3a8029847c8577bf41168658259222102e78116320",
		snd:   "650949223fe7fc70d4691f3a3a3944dfcd273c75daee67be9e7a31d1d274fdf1",
	},
	{
		name:  "city",
		state: stateCity,
		ram:   "a4fc508e28b9d353e2dcb931301f0fb88801dd18537c45d4c65b9d5c52dfc837",
		fb:    "79ce036423dcd52dae7572bf38e758cf20c91e675c5a3c9906e9cf4266283050",
		vram:  "eedc291f6b8e73f63cdddfc2cbcd6bcca533a0565784edfef1b78142087eaeff",
		cpu:   "db8c5c6af567c4dc3f0e13933c323f93ad5b2b5ea80d1ec60c0a2c0b7a73c08e",
		snd:   "8fe828e3a0539c1a5fb060802be2cbb4c11d78bcdf575749c64691e4ed3e8cd7",
	},
	{
		name:  "drive",
		state: stateDrive,
		ram:   "663440b6c865c7c0740e9eaea33090f3f8995880da344f62ed6f300d4036fd1b",
		fb:    "ee283e37fc90035a408fc28493480098d31fd2e3fea708ffea034543fc5a703f",
		vram:  "1669d7aec113f652d2a484b76f422d880037cba7d2f05b023edf41314b573a2c",
		cpu:   "45b4d31ee4288e71f58ac7b86e4a9ac24974500fa790cca25a85d280f0953467",
		snd:   "21848e99567b26cef2f27d6cd8131c989cb8cff0254c52edf69ca21d0bcebb30",
	},
}

// atState restores the machine to one of the recipe's savestates.
func atState(tb testing.TB, state string) *Machine {
	tb.Helper()
	disc, err := OpenDisc(ctImage)
	if err != nil {
		tb.Skip("Crazy Taxi image not present (game images are not committed)")
	}
	m := NewMachine(disc)
	if err := m.LoadStateFile(state); err != nil {
		tb.Skipf("no savestate at %s (work/ is not committed): %v", filepath.Base(state), err)
	}
	return m
}

// gateHashes fingerprints the machine's output and both CPUs' trajectories.
func gateHashes(m *Machine) (ram, fb, vram, cpu, snd string) {
	sum := func(f func(h *hasher)) string {
		h := &hasher{Hash: sha256.New()}
		f(h)
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	ram = sum(func(h *hasher) { h.Write(m.RAM) })
	vram = sum(func(h *hasher) { h.Write(m.VRAM) })

	fb = sum(func(h *hasher) {
		img, err := m.RenderFB()
		if err != nil {
			h.Write([]byte("no fb: " + err.Error()))
			return
		}
		h.Write(img.Pix)
	})

	cpu = sum(func(h *hasher) {
		s := m.CPU.Snapshot()
		for _, r := range s.R {
			h.u64(uint64(r))
		}
		for _, r := range s.Rbank {
			h.u64(uint64(r))
		}
		for _, v := range []uint32{s.SR, s.GBR, s.VBR, s.SSR, s.SPC, s.SGR, s.DBR,
			s.MACH, s.MACL, s.PR, s.PC, s.NextPC, s.FPSCR, s.FPUL,
			s.MMUCR, s.CCR, s.TRA, s.EXPEVT, s.INTEVT,
			s.PTEH, s.PTEL, s.TTB, s.TEA, s.QACR0, s.QACR1,
			s.ICR, s.IPRA, s.IPRB, s.IPRC,
			s.IRLLevel, s.IRLCode, s.CurPC} {
			h.u64(uint64(v))
		}
		for _, bank := range s.FPR {
			for _, f := range bank {
				h.u64(uint64(f))
			}
		}
		for _, q := range s.SQ {
			for _, w := range q {
				h.u64(uint64(w))
			}
		}
		h.bool(s.DelaySlot)
		h.bool(s.PendingDelay)
		h.bool(s.Halted)
		h.u64(s.Steps)
		// The raw on-chip register file, in sorted order — maps must not
		// hash in iteration order.
		keys := make([]uint32, 0, len(s.OnchipRaw))
		for k := range s.OnchipRaw {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			h.u64(uint64(k))
			h.u64(uint64(s.OnchipRaw[k]))
		}
		// The machine's own clocks: a run that drifted a field shows up
		// here and nowhere else.
		h.u64(m.Instrs)
		h.u64(m.Fields)
		h.u64(uint64(m.CurLine))
		h.u64(m.TAWrites)
	})

	snd = sum(func(h *hasher) {
		h.Write(m.AICARAM)
		a := saveARM(m.ARM)
		for _, r := range a.R {
			h.u64(uint64(r))
		}
		h.u64(uint64(a.CPSR))
		h.bool(a.Halted)
		h.u64(a.Instrs)
		h.bool(m.ARMRunning)
		keys := make([]uint32, 0, len(m.AICARegs))
		for k := range m.AICARegs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			h.u64(uint64(k))
			h.u64(uint64(m.AICARegs[k]))
		}
	})
	return ram, fb, vram, cpu, snd
}

// hasher is sha256 with fixed-width writers, so a value's width cannot
// change what it hashes to.
type hasher struct{ hash.Hash }

func (h *hasher) u64(v uint64) {
	var w [8]byte
	binary.LittleEndian.PutUint64(w[:], v)
	h.Write(w[:])
}

func (h *hasher) bool(b bool) {
	if b {
		h.u64(1)
		return
	}
	h.u64(0)
}

// TestFrameHashes is the gate: each recipe runs benchFields fields and the
// machine's output is compared against the pins. On a mismatch the frame
// is written out — a hash says only "different".
func TestFrameHashes(t *testing.T) {
	if testing.Short() {
		t.Skip("the frame gate runs the rasteriser for several fields; -short skips it")
	}

	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			m := atState(t, p.state)

			start := time.Now()
			res := m.RunFields(benchFields, benchBudget)
			elapsed := time.Since(start)

			if m.CPU.Halted {
				t.Fatalf("machine halted during the gate at 0x%08X: %s", m.CPU.CurPC(), m.CPU.HaltReason)
			}
			t.Logf("%d fields: %s (%s/field), %d instructions — %s",
				benchFields, elapsed.Round(time.Millisecond),
				(elapsed / benchFields).Round(time.Millisecond), res.Steps, res.Reason)

			ram, fb, vram, cpu, snd := gateHashes(m)
			for _, c := range []struct{ name, got, want string }{
				{"ram", ram, p.ram},
				{"fb", fb, p.fb},
				{"vram", vram, p.vram},
				{"cpu", cpu, p.cpu},
				{"snd", snd, p.snd},
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
				writeGateFrame(t, m, filepath.Join(os.TempDir(), "dc-gate-fail-"+p.name+".png"))
			}
		})
	}
}

// writeGateFrame puts the picture the machine actually produced on disk.
func writeGateFrame(t *testing.T, m *Machine, path string) {
	t.Helper()
	img, err := m.RenderFB()
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

// TestFrameDeterminism checks the property the gate rests on: the same
// state, run the same way, twice, produces the same machine. If this
// fails, no pinned hash above means anything.
func TestFrameDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the rasteriser for several fields twice; -short skips it")
	}
	for _, p := range gatePins {
		t.Run(p.name, func(t *testing.T) {
			var first [5]string
			for run := 0; run < 2; run++ {
				m := atState(t, p.state)
				m.RunFields(benchFields, benchBudget)
				ram, fb, vram, cpu, snd := gateHashes(m)
				got := [5]string{ram, fb, vram, cpu, snd}
				if run == 0 {
					first = got
					continue
				}
				for i, name := range []string{"ram", "fb", "vram", "cpu", "snd"} {
					if got[i] != first[i] {
						t.Errorf("%s differs between two identical runs: %s vs %s", name, first[i], got[i])
					}
				}
			}
		})
	}
}
