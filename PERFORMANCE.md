# Making the oracles faster

**Status (2026-07-13): done for the 3DS. The frame went from 256.8 ms to 77.6 ms — 3.3× — and
it is byte-identical. The frame debugger's scrubber is 2.3× on top of that.**

**Status (2026-07-16): done for the GameCube. 2.8× on the intro cutscene, 3.1× on the shadow
scene, and 4.4× on a boot stretch — byte-identical. See *The GameCube* below.**

**Status (2026-07-20): Phase 0 + item #1 done for the Xbox. OutRun's driving field is
serial-bound (the banded parallel fill already won; 71% of the field is one goroutine). Item #1
(a content-addressed texture cache that persists across fields AND finally caches the shadow map)
took the warm field from 583.9 → 397.4 ms — 1.47×, byte-identical. Item #2 (register-combiner array
register file + in-place map/op) another −8.4%, byte-identical (guarded by a 20k-config differential
test). Item #3 (raster worker pool) was a measured no-op and reverted. Item #4 (vertex-shader decode
cache) another −8.3%, byte-identical. Separately, FlipVSync (a clock fidelity fix, now on by default) removed the ~6× duplicate
presents the under-paced RDTSC clock was rendering — a 10 FPS sim became 60, the bigger
checkpoint/framedbg lever. See *The Xbox* below.**

Read *Phase 0 — the results* and *What was actually done* below; the original plan (kept, below
the line) guessed the ordering and got most of it wrong, which is exactly what Phase 0 existed to
find out.

A plan, ordered cheapest-and-safest first.

---

## The GameCube (2026-07-16) — and why the 3DS's ordering did not transfer

**Luigi's Mansion, M3 Pro / darwin-arm64. Every number A/B'd back to back in one sitting.**

| workload | before | after | |
|---|---|---|---|
| intro-cutscene, 20 fields | 2.420 s | 0.873 s | **2.77×** |
| shadow, 16 fields | 2.820 s | 0.893 s | **3.16×** |
| **boot stretch, 200M emulated instructions** | 9.763 s | 2.257 s | **4.33×** |

The boot stretch is the one the work existed for — a savestate is a cache, and a Gekko or
Flipper fix invalidates it — and it is now more than four times cheaper.

**The single biggest item was not in the plan, and it is not a graphics optimisation at all:
four fifths of this machine's instructions were three addresses.** See *The idle loop* below.

### What each item was worth, against what it was predicted to be worth

| item | predicted | measured | |
|---|---|---|---|
| 1.1 spin map → O(1) array | 20-40% | **12.7%** boot, **9.7%** draw | the heuristic now costs 0.8%, was 12.7% |
| 2.2 `binary.BigEndian` on the bus | 3-6% (raised to ~14% by Phase 0) | **0.5%** | *both* estimates wrong; kept for the test, not the speed |
| 3.1 TEV decoded once per draw | 10-20% | **4.5%** / 6.5% | |
| 3.1b register file by pointer | — | **3.8%** | not in the plan |
| 3.2 texture state once per draw | 5-10% | **1.5%** | landed for the Phase 4 hazard, not the speed |
| **4 banded parallel fill + pool** | 1.7-2.2× | **23.8%** / **35.8%** | the biggest *graphics* win |
| **5 the idle-loop skip** | — | **2.31×** boot, **1.71×** draw | **not in the plan; the biggest win of all** |
| 6 bandRows 8→4, maxWorkers 8→12 | — | **4.5%** | two knobs set by analogy and never measured |
| 4.7 parallel vertex stage | "doubtful" | **not built** | the bucket is 3.0%; there is nothing there |
| hot-path allocations | ~1% | **not built** | `GOGC=off` moves nothing |

### The idle loop — the biggest win, found by asking a question nobody had asked

A PC histogram over one field. That is all it took, and it should have been the first thing
measured rather than the last:

	801E0864  lwz    r0,5784(r13)     26.57%
	801E0868  cmplwi cr0,r0,0         26.57%
	801E086C  beq    0x801E0864       26.57%

**79.7% of every instruction retired, at three addresses** (76.9% in the shadow scene). It is
the OS idle loop, and it is not a gap in the model — the game is correctly waiting for an
interrupt handler to set a flag. `fieldInstructions` is a *modelling choice*: hardware retires
~8.1M instructions a field and this interpreter hands the game 2M, so the game finishes its
work early and waits. The emulator was spending four fifths of its life proving that zero is
still zero.

**The skip is not a heuristic, and that distinction is the whole of it.** If the machine
returns to a state it has already been in — same PC, same registers, same everything — having
stored nothing, then it is a function with no inputs that has repeated itself, and it will
repeat identically forever until something outside the processor changes. The skip computes
the loop's limit rather than approximating it.

The risk is entirely "what can change while the CPU is not looking", and that list is closed
here **only because every clock is instruction-paced**: the video field, the audio DMA's next
block, an in-flight disc transfer (which writes RAM), and the decrementer underflowing — the
least obvious, being the only clock that raises an exception with no device involved. Those
are deadlines. A running DSP is not a deadline but a veto, since tickDSP steps it off the
Gekko's clock; it is asleep 99.8% of the time.

**Skip a whole number of loop periods, not to the raw deadline.** That is the difference
between equivalent and exact, and it cost two bytes of RAM to learn: jumping to the deadline
made the retrace always interrupt the PC the detector had snapshotted rather than wherever the
loop had got to, so SRR0 differed, the loop took a couple of instructions longer to leave, and
**the game's own frame-time counter noticed** — it reads the time base at 0x801A14E0 and rings
the delta into a buffer, and it came out one tick different. The picture was identical; the
machine was arguably just as correct; it simply was not the *same* machine. The state repeat
tells you the period, so round to it.

**Two things it broke, both caught by tests rather than luck.** `-steps` counted loop
iterations, so the flag quietly ran five times further into the game; the budget counts
emulated instructions now. And the spin detector's window counted iterations too — a machine
wedged in `b .` is *exactly* what the skip fast-forwards, so a genuine hang would have been
reported as an exhausted budget instead of the spin it is.

### Does the idle skip transfer to the other oracles?

**It is one problem with two solutions, and which one an oracle needs is decided by a single
question: does it have an operating system to ask?**

**If there is an OS HLE, the guest tells you it is idle, and you already do this.** The 3DS has
had an idle fast-forward all along — `n3ds/run.go`: *"When every thread is blocked it advances
idle time (waking sleepers) or, if truly deadlocked, halts."* It computes the same shape of
deadline the GameCube's does (gx / dsp / sleep / vblank, whichever is soonest) and jumps
`m.instrs` to it. **That is why its ARM11 is 0.6% of a frame and the GameCube's Gekko was 30%
of one** — not because an ARM11 is cheap, but because the 3DS was never interpreting its idle.
A game blocking in `svcWaitSynchronization` hands the emulator the answer for free. Nothing to
port to the 3DS, the PSP, the 3DO or the Xbox — check they do it, do not rebuild it.

**If there is no OS, there is no signal, and the idle loop IS the operating system.** A
GameCube game's scheduler busy-waits in its own code; to the emulator that loop is
indistinguishable from work — it is `lwz`/`cmplwi`/`beq`, three ordinary instructions. Nothing
in the machine knows they are pointless. **`gc/idle.go` is the LLE answer to the problem the
HLE oracles get for free**, and it recovers the signal the only way an LLE machine can: by
proving the loop cannot progress.

**And the LLE oracles have this problem today — one of them has had it written down for a
year.** `n64/vi.go`, on libultra: *"with no VI interrupt every thread eventually sleeps and only
the idle thread runs — **a `b .` loop** that a boot trace mistakes for a crash."* That is the
same finding as the histogram above, recorded before anyone thought to count how much of a run
it was. The N64 would be the easiest port (its idle thread is literally one instruction, so the
state repeat is found on the first try) if it were still being driven rather than kept as a
verification harness. **The PSX and the DS are the live candidates** — both are LLE, both spin
on VBlank/GPU status, and neither has any idle handling at all today (checked: the only "idle"
in `psx/` is `idLen` and a pad comment). **PS2 once it runs** — with a caution, since its
current frontier *is* a stuck poll loop, and a fast-forward makes a hang cheaper to reach
rather than easier to see.

**The precondition holds repo-wide, and it is not a coincidence.** The deadline is computable
only because every clock is paced by the instruction count — and every core here does that
already, for determinism, not for speed. `n3ds/run.go` calls it *"the N64/DOS discipline"*;
`gekko/timer.go` says the timers are instruction-paced *"for the reason every core in this
repository paces its timers that way: a run must be reproducible"*. **The convention that makes
savestates replayable is the same one that makes the idle skip exact.** An oracle that paced a
device by wall clock could not do this, and could not be replayed either.

**Check for the free version first.** If the guest uses a halt/wait-for-interrupt instruction
rather than spinning, model *that* and fast-forward from it — no detector needed, and no
proof obligation. `tools/cpu/sm83` already has `halt bool // executed HALT, idling until an
interrupt is pending`; `tools/cpu/x86` decodes `HLT`. Whether those run loops actually
fast-forward or merely spin through a halted core is unmeasured, and it is the first thing to
look at on those two.

**What ports and what does not.** The detector is generic — a state repeat with no stores is
not a GameCube fact. Per platform you write exactly two things: a snapshot type (the whole
architectural state; **hash floats by bits or a NaN makes every snapshot differ and the skip
silently never fires**), and a deadline function — *the closed list of everything that can
change while the CPU is not looking*. Getting that list wrong is the only way this breaks, and
the entries are not all obvious: the GameCube's least obvious was **the decrementer**, the only
clock that raises an exception with no device involved at all. A coprocessor stepped off the
CPU's clock (the GameCube's DSP, the N64's RSP) is a **veto**, not a deadline.

**Three gotchas that will recur anywhere.** A budget counted in loop iterations silently
changes meaning — `-steps` must count *emulated* instructions, and a skip must be clamped to
it. A spin/hang detector whose window counts iterations goes quiet exactly when it is most
wanted, because a wedged machine is precisely what the skip fast-forwards. And skipping to the
raw deadline is *equivalent but not exact* — round to the loop period.

**A second-order consequence, measured.** `fieldInstructions` is a fiction tuned partly for
speed (2M against hardware's ~8.1M), and the skip makes faithfulness much cheaper: cold-booting
to field 100 costs 2.08 s at 2M and 2.95 s at 8.1M — **4× the emulated instructions for 1.42×
the time**, where before it would have been 4×. Cheaper, *not* free, and it would move every
trajectory and invalidate every savestate. A real decision, not a freebie — but no longer an
unaffordable one.

### The five things this taught, none of which the 3DS's write-up could have told us

**1. The 3DS's biggest non-parallel win had nothing to bite on, and its "do not touch the
interpreter" verdict was wrong here.** Its −13% was page-tabling a linear region scan; this
machine's bus is already `if a+3 < RAMSize { m.RAM[a] }`. Meanwhile "gekko + rest" is **29.9%**
of a *drawing* field against the 3DS's 6.7% ARM11, so the interpreter *is* a target — but still
not via a decode cache, because `exec.go` is a switch on the raw word and **the decode IS the
shift**. The answer was the run loop, not the CPU.

**2. A cumulative or flat share is an upper bound on what a change inside it can be worth, and
this bit three times in one day, from three directions.** `Write32` was 14% flat and the endian
change bought 0.5% — the 14% is the store's own memory traffic. `shade` was 27% cumulative and
hoisting its whole decode bought 4.5% — the rest is the arithmetic. `sampleTexmap` was 7.2% and
the hoist bought 1.5%. The 3DS learned this about *other threads* (pprof samples the GC's
workers); the GameCube learned it about *the same goroutine*. Measure the change, never the
function.

**3. The workload had to be measured before the parallel design could be chosen, and the answer
was not the 3DS's.** 9,774 draws a field — 68× the 3DS's 143 — of which **5,264 draw no triangles
at all**, and the median survivor covers **84 bounding-box pixels**. Per-draw fan-out is
hopeless at that size. But **half of every bounding-box pixel in the field lives in 45 draws**,
1% of them. So a 256-pixel threshold rejects 65% of draws and still captures **97% of the work**.
That concentration is why the fill pays and why the vertex stage cannot.

**4. `-nospin` measured the largest single CPU-side win before a line was written.** The flag
already existed and disabled exactly one thing. Ten minutes, no code, and it put a number on
Phase 1.1 before anyone argued about it. Look for the flag the machine already has.

**5. ASK WHAT THE MACHINE IS ACTUALLY EXECUTING BEFORE OPTIMISING HOW IT EXECUTES IT.** Every
item in the plan took the workload as given and made it cheaper. A PC histogram — twenty lines
of test code, ten minutes, available from day one — showed that four fifths of the workload
should not have been running at all, and beat the entire rest of the list combined. pprof
answers "where is the time going"; it does not answer "should this code be running". The
buckets in `profile.go` could not have found it either: the idle loop is Gekko time, and the
Gekko bucket is *derived as a remainder*, so it looked like an honest 30% of a drawing field
right up until someone asked what those instructions were.

### What is not done, and honestly why

- **1.2, folding the four per-instruction device ticks.** pprof puts them at ~6.6% together, but
  most of that is `tickDSP`'s *actual DSP work*, which folding cannot remove — by lesson 2 above
  the real prize is probably ~2%. And the idle skip has now removed four fifths of the
  instructions that were paying them at all. Not attempted.
- **2.1, the BAT translation cache.** Predicted 8-15%; `Translate` measures **4% cumulative** and
  `batMatch` 1.5%. It carries a genuinely nasty invalidation hazard (a missed BAT/MSR/HID2/restore
  drop reads the right block and the wrong word, silently). Not worth 4%.
- **3.3, the CMPR block cache**, and **3.4, a full texture cache.** `decodeCMPR` is 4.5%; the
  block cache is the cheap safe half and is the obvious next thing.
- **The framedbg parallel scrubber.** Free of the determinism question, worth ~2-3× on a drag.
  Note a GameCube is ~43 MB against the 3DS's 1.3 GB, so `maxReplayers` should be bounded by CPU,
  not memory — do not inherit the 3DS's `4` or its reasoning.

---

## The Xbox (2026-07-20) — OutRun is serial-bound, and the parallel fill already won

**OutRun 2006: Coast 2 Coast, `an3-drive.state` (an in-race driving field), M-series darwin/arm64.
Phase 0 only so far: the measurements and the gate, no optimisation landed yet.**

The field costs **651 ms**, measured two ways that agree — the machine's own `-profile` buckets
and a host pprof (`-cpuprofile`) over 20 fields:

| bucket (machine, wall) | ms | % | parallel? | what pprof says is inside it |
|---|---|---|---|---|
| command decode (derived) | 260.6 | **40.0%** | serial | ~half is `texDecode`; the rest is `combDecode`, near-plane clip, 94k method dispatches |
| rasterise | 184.5 | 28.3% | parallel (~1.84× of 10 cores) | the per-fragment register combiner — `combine`+`combFetch`+`combMap`+`combOp` ≈ **33% of all samples** |
| vertex + xf | 139.4 | 21.4% | serial | transform + the vertex-shader interpreter |
| x86 + rest | 65.2 | 10.0% | serial | the CPU |
| clear | 1.8 | 0.3% | | |

Counters: 523 draws, 94,310 methods, 1.76M fragments drawn, 419k depth-killed, 627k alpha-killed,
1.62M x86 instructions. (pprof also shows `compress/flate` at ~5% — that is the `-flipshots` PNG
encoder writing the sample frames, an instrument artefact a real field does not pay. Excluded.)

**The strategic finding is the opposite of the GameCube's.** On Luigi's Mansion the prize was the
CPU — four fifths of every instruction was the idle loop. Here the CPU is **1.62M instructions and
10% of the field**: there is nothing to skip, and the idle-loop question is answered before it is
asked. **The field is GPU-bound and, within the GPU, SERIAL-bound — 71% of it (vertex +
command-decode + x86 = 465 ms) runs on one goroutine.** The banded parallel fill has already done
its job: rasterise is only 28% and spread across cores, so *more* raster parallelism is
Amdahl-capped at about 1.4×. The wins are in the serial buckets, and they are the same two moves
the 3DS and GameCube already proved — **decode-once-per-draw** and a **persistent worker pool** —
pointed at different code.

### The ranked plan (predictions, not results — nothing here is measured yet)

| # | item | predicted | why, from the profile | prior art |
|---|---|---|---|---|
| 1 | **persist the texture cache across fields** (write-invalidated) | 10-18% | `texCache` is dropped every pusher run (`nv2a_pfifo.go:72`), so every field re-decodes every texture; `texDecode` is 11% of samples, **serial**, and ~half of the 40% command-decode bucket — it hits wall time with no Amdahl discount | 3DS precise tex-cache invalidation: 2,216 → 2.2 decodes/frame |
| 2 | **compile the register combiner once per draw** | 8-14% | `combine()` re-interprets fully-general state **per fragment** (~33% of samples) for OutRun's single MODULATE stage; `combDecode` already extracts the state per draw — extend it to a resolved plan/closure so the per-fragment path is a few float ops. Also kills the `[4]float32`/`duffcopy` traffic | 3DS TEV decoded once/draw: −14.2%, the biggest non-parallel win |
| 3 | **persistent raster worker pool + lower the threshold** | 5-10% | `rasterParallel` spawns 10 goroutines **per draw** across 523 draws → `pthread_cond_signal`+`notewakeup` ≈ 13% of samples, and only 1.84× of 10 cores; `rasterParallelMinArea=16384` leaves many draws serial. Amdahl-capped (raster is 28%) but removes the scheduler waste | 3DS `workpool.go`: −11.5%, thresholds 8192 → 256 px |
| 4 | **attack the vertex stage** (decode cache or parallelise) | 10-15% | 139 ms / 21% is fully **serial** and the 2nd-largest bucket; a vertex is a pure function of its index, so it parallelises deterministically. GC dismissed this ("nothing there", 3% bucket) — but Xbox's is 21%, a real target | 3DS vsh decode cache −9.3%, parallel vertex −25.7% |

Per the doc's own lessons: a CPU-sample share is an **upper bound** on the wall win (the combiner's
33% includes irreducible float arithmetic); A/B every change back-to-back in one sitting and never
compare across sessions; and the frame is serial-bound, so do **not** over-invest in raster
parallelism — spend the effort on #1 and #4.

### What was actually done, and what it was worth

**The running tally, measured on the warm field** (`BenchmarkWarmFields`: one cold field to fill
the cache plus seven warm ones, the steady state a real run lives in), A/B'd back to back:

Each row is A/B'd against its OWN in-sitting baseline (the doc's caveat: never compare a warm-field
number to one from an earlier session — the variance between machine states is bigger than several
of these wins, so the absolute ms drift between rows and the delta within a row is the honest read).

| phase | warm ms/field | in-sitting Δ | byte-identical? |
|---|---|---|---|
| Phase 0 (baseline) | 583.9 | — | (gate pinned) |
| #1 texture cache (colour + depth) | **397.4** | **1.47× / −32.0%** | yes |
| #2 combiner: array reg file + in-place map/op | **413.3 → 378.8** | **−8.4%** | yes |
| #3 raster worker pool + worker scaling | 384.6 → 380.2 / 421.8 | **~0 / worse — reverted** | (not shipped) |
| #4 vertex-shader decode cache | **384.6 → 352.7** | **−8.3%** | yes |

**#1, predicted 10-18%, measured −32% — and the surplus is the whole lesson.** The plan aimed the
texture cache at cross-field *colour* re-decode. That part landed exactly as predicted and no more:
persisting the colour cache across fields (validating each entry by an FNV-1a of its source bytes,
so a reflection RTT the game re-renders every field is re-decoded and a static road texture is
kept) was worth **3.75%** on its own — `decodeDXT`/`decodeSwizzled` fell off the profile, `hashRAM`
never appeared, and that was that.

**Then Phase 0's lesson 5 paid out: ask what the code is executing.** With colour decode gone,
`texDecode`'s residual 1.81 s of samples was one thing — the **shadow-map depth texture**, a 512×512
buffer that the old cache *bypassed on purpose* (the caster pass writes it mid-run) and therefore
re-decoded on **every receiver draw** that sampled it. The content hash makes that bypass
unnecessary: a mid-run caster write changes the bytes, the hash sees it, and the map is decoded
**once per run after the caster** instead of once per receiver. That is where the other ~28 points
came from. The subsystem *command-decode* bucket fell from **260.6 → 62.9 ms** and the field is now
cleanly raster (41%) + vertex (30%) bound — the shape the ranked plan predicted, reached a phase
early.

**Correctness.** Byte-identical through both changes, gate and determinism green — and the gate
*earns* it here: an3-drive casts shadows, so a stale depth decode would move the surface hash. The
one residual assumption (≤ one caster pass per run, so the within-run "already validated" fast path
cannot serve a stale map) is guarded by that same frame. The FNV-1a span is computed by `texSource`
to match exactly what each decoder reads, so any byte a decode would read changing forces a
re-decode; a 64-bit content hash can in principle collide (~2⁻⁶⁴), the standard texture-cache bet,
with the gate as backstop.

Files: `nv2a_texture.go` (the cache is now content-addressed — `texEntry`, `hashRAM`, `texSpan`,
`texSource`, `cacheTex`), `nv2a_pfifo.go` (bump a run sequence instead of dropping the cache),
`nv2a_pgraph.go` + `state.go` (the map's value type), `bench_test.go` (`BenchmarkWarmFields`).

### #2 — the register combiner, predicted 8-14%, measured −8.4%

The combiner (`combine`, the per-fragment fragment pipeline) was **31.5% of samples** and the top
target after #1 — but the plan's "compile it per draw" framing overreached: `combDecode` already
decodes the control words once per draw, so what remained per fragment was the combiner *math*,
which is irreducible (a CPU-sample share is an upper bound on the wall win — lesson 2). What was
NOT irreducible was the dispatch around it: a switch-per-access register file, and a `[4]float32`
copied out of every `combMap`/`combOp` call (~2.4% of the frame was `runtime.duffcopy`). Two safe,
byte-identical changes: the register file became a `[16][4]float32` **array** (read is an index,
write an index plus a writable guard — the old switch was pure dispatch over the same values), and
`combMap`/`combOp` **mutate in place** instead of returning a fresh array. Array file alone −3.0%;
adding the in-place helpers −8.4% total. `combine`'s cumulative share fell 31.5% → 25.5%.

The correctness bar is high here — the combiner is fragment output, and the an3-drive gate exercises
only OutRun's single MODULATE stage, so a config the game never programs but the hardware allows
would slip past it ([[reference-match-hides-bugs]]). Guarded by `TestCombinerMatchesRef`: a FROZEN
copy of the pre-rewrite combiner (its own types, byte-for-byte the old arithmetic) run against the
live one over **20,000 random configs × random fragment inputs**, asserting bit-identical float
output. It passes, the gate did not move (no re-pin), and the arithmetic was preserved verbatim so
FMA contracts both paths identically (the differential test also passes under `-race`, FMA off).
Not done: computing the colour side in rgb and the alpha side as a scalar (both discard components
the writes never read) would shave more, but it reorganises the float ops and earns real FMA/order
risk for a raster path that is parallel (wall win compressed) — deferred, not free.

Files: `nv2a_combiner.go` (array `combRegs`, in-place `combMap`/`combOp`), `combiner_diff_test.go`.

### #3 — the raster worker pool, predicted 5-10%, measured ~0 and reverted

The 3DS won −11.5% from a persistent worker pool because *there* goroutine creation was the cost
(its profile showed `newproc`/`malg`/`stackalloc`). On the Xbox the profile said otherwise —
`pthread_cond_signal` at ~15% and **no creation functions at all** (Go's free lists already make a
per-draw `go func()` cheap). So the cost is the wake/join itself: ~100 parallel draws × 10 lanes ≈
1000 contended futex ops a field on the shared `WaitGroup`, and a machine pinned under two cores of
twelve. A pool sends to a channel instead of spawning — but that still *wakes* each lane, so it does
not touch the actual cost. Measured, in one sitting: goroutine-spawn 384.6, pool (same 10 lanes)
380.2 — **within noise**. Ported the pool for nothing.

The follow-on idea — size the fan-out to the draw (`rasterWorkersFor`, 2 lanes for a threshold draw
up to 10 for a full-screen pass) to cut the wake count — was **worse: 421.8 (+11%)**. Fewer lanes
meant the medium draws that were the parallel win now filled more serially, and the fill parallelism
it gave up outweighed the wakes it saved. So the wakes were not free, but the lanes doing them were
earning more than they cost; trading them away is a loss.

Both reverted. The real lesson is Phase 0's, re-confirmed by three measurements: **this frame is
serial-bound** (71% one goroutine), so the raster fan-out is already past the point where more
parallelism pays — and neither cheaper dispatch nor a different lane count changes that. The next
lever is the serial work, i.e. item #4 (the vertex stage), not the fill. A 3DS win does not transfer
just because the code looks the same; the bottleneck has to be the same, and here it was not.

### #4 — the vertex-shader decode cache, predicted 10-15%, measured −8.3% (and this time it DID transfer)

The vertex stage is **serial** (one goroutine) and the largest wall bucket at 30.9%, so unlike the
fill its CPU cost is its wall cost — and pprof, which counts CPU across the parallel raster's ten
workers too, undercounts it (10% of samples against 31% of wall). `transform` → `vshRun` ran the
transform program per vertex by calling `vshDecode(pc)` for **every instruction of every vertex** —
re-parsing ~20 bit-fields per instruction over a program that is constant for the whole draw (the
output path writes o registers and constant memory, never the program words). This is precisely the
per-vertex-re-decode the 3DS's shader cache killed for −9.3%, and here the bottleneck matches — so
the same move works: `vshCompile` decodes start→FINAL into a reused `g.vshProg` once per draw, and
`vshRun` iterates the decoded slice. Byte-identical by construction (same `vshDecode`, same
execution, just hoisted); the vertex bucket fell **134 → 104 ms**, the warm field 384.6 → 352.7
(−8.3%), gate and determinism unmoved (no re-pin), `-race` clean. `TestVSHInterpreter` had to learn
the new shape — it drives the interpreter directly, so it now `vshCompile`s before `vshRun`.

Contrast with #3: this is the same "a 3DS win might transfer" bet, and it paid — because the
bottleneck was the same (a per-vertex interpreter re-deriving a per-draw constant), not merely the
code shape. Not done: `fetchArrayVertex` and the ~120-byte `kelvinVtx` append-copy are the vertex
bucket's remaining cost; parallelising the transform is off the table while a program can write
constant memory a later vertex reads (the vertices are not independent), which is also why the
decode cache — not a fan-out — was the right tool.

Files: `nv2a_vsh.go` (`vshCompile`, `vshRun` over `g.vshProg`), `nv2a_vertex.go` (compile once in
`runDraw`), `nv2a_pgraph.go` (the `vshProg` buffer), `nv2a_kelvin_test.go` (test compiles first).

### The bigger checkpoint lever was not a per-frame optimisation — it was 6× fewer frames

A per-frame speedup is not the only way to make run-to-checkpoint cheap; presenting fewer frames
to reach the same game state is the other. OutRun's race loop (`0x20AFA`) is an RDTSC
fixed-timestep catch-up, and the guest TSC is instruction-paced — so a present that retires only
~2.1M instructions accrues ~0.18 of a 1/60 s field, and the loop **duplicated each simulation step
across ~6 presents** (a 10 FPS sim on a 60 FPS engine; the presented picture updated only every
~6th flip, which made OutRun in framedbg unusably slow to scrub). `Machine.FlipVSync` (on by
default, `nv2a_kelvin.go`/`interrupt.go`) models `FLIP_STALL`'s vsync wait by advancing the tick a
whole field per present — verified at exactly 1.000 vblank/flip — so the sim steps every frame and
those five duplicate presents per step **stop being rendered at all**. Reaching a given game state
now costs ~6× fewer flips of rasteriser work, which for framedbg scrubbing and checkpoint building
dwarfs the 32% the texture cache bought per frame. It is a fidelity change (it re-times the
trajectory), so the gate was re-pinned with it on; the picture at `an3-drive` is identical to the
eye. Guarded by `TestFlipVSyncCadence`; `-flipvsync=false` / `RR_FLIP_VSYNC=0` is the A/B control.

### Phase 0 (2026-07-20) — the gate is built

| | where | how to run it |
|---|---|---|
| 0.1 pprof | `bootoracle -cpuprofile F` (already existed) | `bootoracle -image OutRun….iso -loadstate …/an3-drive.state -gpu -profile -cpuprofile x.prof -flipshots 1:1:20:x` |
| 0.2 gate + bench | `tools/platform/xbox/bench_test.go` | `go test ./tools/platform/xbox/ -run TestFrameHashes` / `-bench BenchmarkField` |
| 0.3 per-subsystem timing | `tools/platform/xbox/profile.go`, `bootoracle -profile` (already existed) | `bootoracle … -gpu -profile` |

`TestFrameHashes` pins RAM + the resolved render surface + the x86 register file after two fields
from `an3-drive.state`; `TestFrameDeterminism` proves the property the pins rest on — the same
state run twice is the same machine; `BenchmarkField` reports ns/field with draws and fragments
alongside, because *fragments per millisecond* is the number that says whether a change worked. The
disc image and the savestate are not committed (`work/` and game images are ignored), so all three
**skip** when absent. `-race` skips the pins: it inhibits FMA contraction and moves the last float
bits ([[reference-match-hides-bugs]] is not this, but the FMA caveat from the 3DS is). The frame the
pins defend was looked at first ([[pinned-hash-guards-change]]); a mismatch writes the PNG it
actually produced and names the path.

---

## Why this is worth doing

A savestate is not an asset, it is a **cache** — and the cache is invalidated by exactly the
work we do most: fixing a CPU or GPU bug changes the trajectory that produced the state, so the
state has to be rebuilt. Captain Toad's was rebuilt twice in one day after two ARMv6K fixes (40
minutes each), and now that the PICA200 actually rasterises geometry, reaching the same point
costs **over two hours**. That cost is not stable — it grows every time the model gets more
complete. It is a tax on the exact activity the repo exists to do.

So the goal is not "playable". It is: **make the run-to-checkpoint cheap enough that a CPU fix
costs minutes, not an afternoon.**

## The rule that governs every step below

Every optimisation here must be **bit-exact**. These machines are oracles; their output is the
ground truth other code is verified against ([[decode-reimplement-rule]]). An optimisation that
changes one pixel has not sped up the oracle, it has broken it. So Phase 0 builds the gate
before Phase 1 touches anything, and no optimisation lands without it passing.

---

## Phase 0 — Measure. Do not optimise anything yet.

Three deliverables, in this order. The first two are the honest answers; the third is the one
that is nice to look at.

### 0.1 A pprof profile of one Captain Toad frame — half a day, zero risk

Add `-cpuprofile FILE` to the 3DS bootoracle (`pprof.StartCPUProfile`, four lines). Load
`toadstereo4.state`, run 60 frames, profile it.

This answers "is it 90% rasterising?" *today*, with no instrumentation, no bias and no risk of
the measurement changing the thing measured. **Nothing else in this document should be started
until this has been run**, because the entire ordering below is a guess that pprof can overturn
in ten minutes. My guess is that `writePixel` + `depthCompare` + `Machine.Read`/`Write`
dominate, and that the ARM interpreter is a minority of the frame — but it is a guess.

### 0.2 A benchmark with a bit-exactness gate — the thing that makes the rest safe

`tools/platform/n3ds/bench_test.go`:

- Load a committed-recipe state, run a fixed number of frames, report ns/frame.
- **Hash the colour buffer, the depth buffer and the CPU register file at the end.** Pin the
  hashes. Any change that moves them fails.

This is the harness every later phase reports against, and the thing that stops a "3× faster"
claim from quietly being a "3× faster and subtly wrong" claim. It is worth more than any single
optimisation in this file.

### 0.3 Per-subsystem timing, surfaced in framedbg

**The measurement design matters, because the naive version destroys what it measures.** A
`time.Now()` pair around each fragment costs more than the fragment does. So time only at
boundaries that are *already coarse*, and derive the rest:

| bucket | where it is timed | granularity |
|---|---|---|
| GPU: command decode | around `DecodePICA` + the `g.write` loop in `GPU.Execute` | per command list (hundreds/frame) |
| GPU: vertex fetch + shader | around the vertex loop in `GPU.draw` | per draw (~143/frame in Captain Toad) |
| GPU: rasterise | around the triangle loop in `GPU.draw` | per draw |
| GPU: texture decode | around `GPU.texture` cache misses | per miss |
| DSP | around `dspTick` | per audio frame (~3.4/VBlank) |
| HLE (svc + IPC) | around `handleSVC` | per svc |
| ARM interpretation | **derived**: wall time − everything above | — |
| idle skip | around the fast-forward branch in `Run` | per event |

That is a few hundred `time.Now()` calls per frame against millions of fragments — unmeasurable
overhead. Carry the counters that already exist alongside it (`Draws`, `PixelsDrawn`,
`DepthKilled`, `RejectedTris`, `CulledTris`, `ListHops`) as per-frame deltas, because
*fragments per millisecond* is the number that tells you whether a change worked.

Live behind a `Machine.Profile bool` so a normal run pays nothing.

**In framedbg**, this is a new optional capability — the design already accommodates it:

```go
// debug.go
Profiler interface {
    FrameProfile() []ProfileBucket   // Name, Millis, Count
}
const CapProfile = "profile"
```

`frameMsg` already carries `StepMs`; add the bucket list next to it, and add a capability-gated
`profile` panel (a bar per bucket + the counters). Step a frame, read:

```
frame 12 — 812 ms
  rasterise        640 ms  79%   3.4M fragments (520k drawn, 2.9M depth-killed)
  vertex + shader   90 ms  11%   143 draws, 41k vertices
  ARM11             41 ms   5%   1.2M instructions
  command decode    30 ms   4%   98,669 writes
  DSP               12 ms   1%
```

That is the artefact you asked for, and it is also the thing that will keep every later phase
honest.

---

## Phase 0 — the results (what the measurements say)

All three deliverables are built and in the tree:

| | where | how to run it |
|---|---|---|
| 0.1 pprof | `bootoracle -cpuprofile F` (+ `-frames N`, a frame-bounded run) | `bootoracle -image toad.cci -loadstate …/toadstereo4.state -frames 60 -cpuprofile toad.prof` |
| 0.2 gate + bench | `tools/platform/n3ds/bench_test.go` | `go test ./tools/platform/n3ds/ -run TestFrameHashes` / `-bench BenchmarkFrame` |
| 0.3 per-subsystem timing | `tools/platform/n3ds/profile.go`, `bootoracle -profile`, and a `profile` panel in framedbg | `bootoracle … -frames 60 -profile` |

The gate pins three hashes (VRAM, the two presented screens, every thread's register file)
after four frames from `toadstereo4.state`. `TestFrameDeterminism` proves the property the
gate rests on — the same state, run twice, is the same machine — and it passes, so the
oracle is genuinely replayable and the pins mean something. The frame the pins defend was
looked at, on both screens, before they were written down; on a mismatch the test writes the
frame it actually produced and says so, because [[pinned-hash-guards-change]].

### The frame, measured (Captain Toad, opening stage, M3 Pro, darwin/arm64)

**234 ms/frame** over 60 frames — and only **30 of the 60 frames draw anything** (the game
renders at 30 Hz), so a *drawing* frame costs about **460 ms**. Per drawing frame: 143 draws,
187,899 fragments written, 39,214 depth-killed, 2,507 command-buffer hops, 4.46M ARM11
instructions.

pprof and the machine's own buckets were built independently and agree:

| bucket (machine, wall) | | pprof (CPU samples) | |
|---|---|---|---|
| rasterise | 44.9% | `GPU.triangle` | 41.9% |
| vertex + shader | 41.1% | `GPU.shaderRun` 23.1% + fetch/copies | ~39% |
| gx transfers | 3.9% | `gxDisplayTransfer`/`MemoryFill`/`TextureCopy` | ~2% |
| texture decode | 2.0% | `GPU.texture` miss | ~2% |
| command decode | 1.3% | `DecodePICA` + the buffer read | ~3% |
| svc + ipc | 0.1% | `handleSVC` | <1% |
| dsp | 0.0% | `dspTick` | <1% |
| arm11 + rest (derived) | 6.7% | `CPU.execARMv6` | **0.6%** |

(The two disagree in detail because pprof samples every OS thread — the garbage collector's
workers included — while the buckets time the goroutine the machine runs on. They agree on
everything that matters.)

### What this changes — read before doing Phase 1

1. **The ARM interpreter is ~1% of the frame.** `execARMv6` is 0.6% of samples. **Phase 1.3
   (word-granular bus) and Phase 1.4 (instruction-decode cache) are capped at about 1%
   between them and should not be done** — 1.4 in particular is described below as "worth a
   lot", and against this workload it is worth almost nothing. This is the single biggest
   correction Phase 0 makes.

2. **The biggest cost in a frame is not in this document at all: the PICA vertex shader.**
   41% of the frame is the vertex path, and 23% is the shader *interpreter* — `readSrc`
   (9.3% on its own), `arith`, `exec` — re-decoding the same shader instruction for every
   vertex of every draw. The instruction-decode-cache argument of 1.4 is the right argument
   pointed at the wrong CPU: it belongs here, where it pays an order of magnitude more.

3. **The memory bus is real but it is not the ballgame.** `Machine.Read` + `Write` are 14% of
   samples, of which the linear `regionOf` scan is ~10%. Phase 1.1 (page table) is worth
   roughly that, and it is cheap and safe — do it. But Phase 1.2's premise is weaker than it
   reads: the framebuffer traffic specifically (`writePixel` + `depthCompare` + `depthWrite`
   + `shadowMapWrite`) is only **~7%** of the frame, not the "whole ballgame" the section
   claims.

4. **Phase 2.3 is stale — its premise is gone.** It says the workload is 85% garbage (2.9M
   depth-killed against 520k drawn). That was the *broken-geometry* frame. Since the vertex
   permutation and framebuffer-origin fixes (Parts XVII/XVIII), the frame is **187,899 drawn
   against 39,214 killed — 17% killed**, which is an ordinary overdraw figure. The rasteriser
   is now a legitimate optimisation target and does not need to wait for a geometry fix.

5. **New, cheap, and not in the plan: the allocator.** ~10-15% of samples are in
   `mallocgc`/`madvise`/`mspan.init`/`memclr`, from three allocations on the hot path —
   `buf := make([]byte, size)` per command list (2,500 of them a frame), `DecodePICA`'s
   `append`, and `outs := make([]vsOut, …)` per draw. All three can be reused buffers on the
   GPU struct. Bit-exact by construction, an afternoon's work.

6. **Also new: 6% of the frame is `runtime.duffcopy`, almost all of it in `GPU.draw`** —
   `shaderRun` returns `[16][4]float32` (256 bytes) by value, per vertex, and `mapOutputs`
   copies it again. Write into a caller-owned buffer instead.

### Two caveats the measuring itself taught

**The benchmark's variance between machine states is bigger than several of the wins on this
list.** The same commit, the same frame, the same 143 draws, measured 225 ms/frame on a cold
laptop and 257 ms/frame on a warm one — 14%, which is more than Phase 1.2 is worth in total.
An A/B run back-to-back is stable to well under 1%. So: **never compare a number to one from
an earlier session.** Measure before and after, in the same sitting, and quote the delta.

**The per-subsystem timing costs nothing when it is off, and that was checked rather than
asserted.** Stripping every hook out of the machine (and then the struct fields too, and then
`profile.go` entirely) left the benchmark at 256 ms — unchanged. The `Machine.Profile` bool
is what makes that true: the clocks are only read at boundaries that already cost thousands
of times more than a clock read.

---

## What was actually done, and what it was actually worth (2026-07-13)

**256.8 → 77.6 ms/frame: 3.3× faster, with the frame byte-identical.** Both screens were
compared image against image, not merely hashed, and Super Mario 3D Land's pinned md5 did not
move. A drawing frame went from 460 ms to about 200 ms; fragments per millisecond went from
989 to 2,235.

Every item was measured A/B, back to back, in one sitting. **Predictions are in the table
because most of them were wrong, and the pattern in how they were wrong is the most useful
thing here.**

| item | predicted | measured | |
|---|---|---|---|
| A1 hot-path allocations + 256-byte value copies | 10-15% | **0.8%** | the estimate was nonsense; see below |
| A2 shader decoded-instruction cache | 12-18% | **9.3%** | |
| A3 page-table the address space | 6-9% | **13%** | the one that beat its estimate |
| A4+A5 direct target access, one tiled offset per fragment | 7-14% | **2.3%** | A3 had already taken most of it |
| A4 8×8 tile rejection | included above | **−1.4%** | *slower*. Reverted. |
| C1 one reciprocal instead of 11-18 divides per fragment | 10-20% | **0%** | *and it changed the output.* Reverted. |
| B1 parallel vertex shading | ~30% | **25.7%** | |
| B2 parallel tiled rasteriser | 1.7-2.2× | **27.6%** | |
| a persistent worker pool + lower thresholds | — | **11.5%** | not in the plan; see below |
| B3 parallel scrubber (framedbg) | "instant" | **2.3×** on a drag | 8 positions: 1.60 s → 0.70 s |
| **TEV decoded once per draw, not per fragment** | — | **14.2%** | the biggest single win after the parallel stages |
| command lists copied, not fetched byte-by-byte | — | **3.1%** | |
| precise texture-cache invalidation | — | 1% wall, but 2,216 → 2 decodes/frame | *and it exposed a real bug* |
| per-draw vertex-fetch plan | — | **−1.7%** | *slower*. Reverted. |

### The four things this taught, which are worth more than the 2.4×

**1. A CPU-sample share is an upper bound on a wall-clock win, and often a wild one.** A1 was
predicted at 10-15% because pprof put that many samples in `mallocgc`/`madvise`/`mspan`. But
pprof samples *every OS thread*, and Go's garbage collector runs on its own: `GOGC=off` changes
the frame time by **nothing**, and `gctrace` shows the GC running twice in a ten-frame run.
Those samples were real CPU and zero wall clock. Only what runs on the goroutine the machine
runs on can be saved by making it cheaper.

**2. Two of the plan's most confident items were worth nothing, and one of them was worse than
nothing.** The 8×8 tile rejection and the reciprocal-instead-of-divide are both textbook. Both
are *right for a rasteriser that is not this one*: tile rejection pays against long spiky
triangles with huge bounding boxes and slivers of coverage — the workload the *broken* geometry
produced — and the divides are pipelined and overlapped on an M3, so removing them saved zero.
The reciprocal would also have moved the oracle's output. It was reverted for a 0% gain that
cost a pixel.

**3. `go test -race` changes the floating-point result.** The race build inhibits FMA
contraction, which arm64 applies to the shader's MAD, so it computes a frame that differs in
its last bits — *with the machine forced single-threaded, where a data race cannot exist*. The
pinned hashes are a property of the ordinary build; the gate skips under `-race` and says why.
What `-race` is for here is proving there is no data race (there is none), and the correctness
claim for the parallel stages is not a hash from last week but
`TestParallelVertexMatchesSerial`: the parallel machine agrees, bit for bit, with what *this
build* computes serially.

**4. Parallelism was the biggest win, and it cost nothing in determinism — because the
partition, not the scheduler, decides the answer.** A vertex is a pure function of its index. A
band of 16 rows is filled by exactly one worker, and within a band the triangles are applied in
submission order. Bands are handed out from a shared counter, so *which* worker takes *which*
band varies run to run — and cannot change the result, only the time. The three things that had
to be fixed first were all writes hiding in read-shaped code: the lazily-filled shader decode
cache, the GPU's counters, and **a texture cache miss, which is a write**.

### The worker pool — the item that was not in the plan

The parallel stages spawned their workers **per draw**: 143 draws a frame, so ~1,100 goroutine
creations a frame. Cheap each, not free in aggregate — and expensive enough that both stages had
to refuse to split a small draw at all, so **68 of the 143 draws ran serially for want of a few
microseconds**. A pool (`workpool.go`) makes a fan-out a channel send instead, which let the
thresholds come down (vertex: 64 → 16 vertices per worker; fill: 8,192 → 256 bounding-box
pixels). **−11.5%.**

Bands went from 16 rows to 8 (worth ~3%: more bands balance better). **Four-row bands measured a
hair faster still and are deliberately not used** — an 8×8 Morton tile is eight rows tall, so a
four-row band lets two workers write into the same tile and false-share its cache lines. Same
speed, worse reason.

The pool means a `Machine` now owns goroutines, and they keep it alive after the last reference
is dropped. `Machine.Close()` stops them; the frame debugger's adapter calls it.

### The fragment stage — the same argument, three times

The one optimisation that keeps paying on this machine is **decode the thing once, where it stops
being per-anything**: the vertex shader's instructions (−9.3%), the TEV's six stages (−14.2%), and
the render-target/lighting state (folded into the draw). All three were re-deriving, per vertex or
per fragment, something that cannot change while a draw runs — a register write only happens in
the command processor, and the command processor is not running.

**It does not generalise, and that is the point.** The same trick applied to the vertex *attribute
fetch* was 1.7% SLOWER: the format decode there is a few shifts, and the indirection of a
precomputed plan cost more than it saved. Measured, reverted.

### The texture cache, and a bug the gate caught

Every GX DMA and TextureCopy dropped the **whole** texture cache, and Captain Toad issues those
constantly — so it was rebuilt from scratch several times a frame: 2,216 whole-texture decodes a
frame for a scene with a few dozen textures. Invalidating only what a write overlapped brings that
to **2.2 a frame**.

And the gate immediately failed — because the blanket clears were **hiding a real bug**. The shadow
map is written by the *rasteriser*, not by the GX engine, so nothing invalidated it; the scene draws
sample it back through a Shadow2D unit, and a stale decode simply never survived long enough to be
seen. The moment the cache stopped being thrown away, it was. A draw that writes its colour target
now invalidates what overlaps it. **The gate moved, the frame was written out, the cause was a bug —
not a re-pin.**

### What is left

- The bucket shares now: rasterise 43%, vertex + shader 29%, the derived remainder 20% (mostly the
  parallel stages' scheduling), GX transfers 4.7%, texture decode 2.6%, command decode 0.5%.
- **The ARM11 is 0.6% of the frame.** It stays that way. Do not build the instruction-decode
  cache or the word-granular bus.
- The remaining single-threaded work is the command decode and the GX transfers; neither is
  large enough to be worth the next increment of complexity. **Stop here unless the frame
  changes shape.**

---

Everything below this line is the original plan, written before any of it was measured. It is
left as it was, and wrong where the results say so.

---

## Phase 1 — The memory bus. Cheap, safe, single-threaded, probably the biggest win.

`Machine.Read`/`Write` are **byte-granular** and call `regionOf(addr)`, which is a **linear scan
over every mapped region**, for *every single byte*. Then:

- `writePixel` does 4 reads + up to 4 writes = **8 region scans per fragment**.
- `depthCompare` does 3 reads, `depthWrite` 3 writes = **6 more**.
- `ReadWord`/`WriteWord` are 4 byte calls each.
- The ARM core fetches instructions through the same byte bus: **4 region scans per instruction
  fetch**, before it does any work.

So a fragment costs ~14 linear region scans plus ~14 function calls. This is very likely the
whole ballgame, and none of it needs threads.

### 1.1 Page-table the address space

Replace the linear `regionOf` scan with a direct lookup: `pages [1<<20]*memRegion` indexed by
`addr >> 12` (4 KiB pages, 8 MiB of pointers — fine). Rebuilt on `mapRegion` and on snapshot
restore. One indexed load instead of a scan, and it speeds up **CPU and GPU alike**.

Cheaper alternative if the 8 MiB bothers: a one-entry "last region used" cache in front of the
scan. Temporal locality on this workload is enormous. Measure both; the page table is more
predictable.

### 1.2 Give the rasteriser direct access to its target

`fbstate()` already resolves the colour and depth addresses **once per triangle**. Resolve the
*region* there too, and hand `writePixel`/`depthCompare`/`depthWrite` a `[]byte` + base offset
so they index a slice directly instead of going through the bus at all.

Fall back to the bus path if the target does not sit inside one contiguous region (it always
will — it is VRAM — but do not assume it; fall back, do not panic).

This removes ~14 region lookups and ~14 dynamic calls per fragment, and it is a purely local
change to four functions in `gpu_raster.go`.

### 1.3 A word-granular bus for the CPU

`arm.Bus` is byte-granular by construction, which costs 4 dispatches per fetch and per
load/store. Add an optional interface the core uses when the bus implements it:

```go
type Bus32 interface { ReadWord32(a uint32) uint32; WriteWord32(a, v uint32) }
```

`n3ds.Machine` implements it over the page table. Same for `Bus16`. The core keeps working
unchanged against any bus that does not offer it (N64/PSX/etc. are untouched).

### 1.4 An instruction-decode cache

The interpreter re-decodes every instruction **every time it executes it** — inner loops decode
the same word millions of times. A direct-mapped cache keyed by `(addr, thumb)` holding the
decoded `Inst` turns that into one tag compare. Code regions are effectively static; invalidate
the cache on a write into a code region, on `mapRegion`, and on snapshot restore.

This is the standard interpreter win and it is worth a lot — but it is **second**, because 1.1
and 1.2 are simpler and (per my guess) larger. pprof decides.

---

## Phase 2 — The rasteriser's inner loop. Still single-threaded.

Only if Phase 0 says fill is dominant, and only *after* Phase 1, since Phase 1 changes what the
inner loop costs.

### 2.1 Incremental edge functions

`triangle()` recomputes three `edge()` calls in float per pixel over the bounding box. The
textbook form steps them incrementally (`w += dx` per pixel, per row), turning 3 multiply-adds
per edge into one add. Same arithmetic, same results — but verify bit-exactness, because float
accumulation is **not** associative and stepping can differ from recomputing in the last bit.
If it does differ, use fixed-point integers (which the comment already claims it does), which
step exactly.

### 2.2 Reject empty tiles before touching pixels

Test the edge functions at 8×8 tile corners; skip the whole tile when it is entirely outside.
Big win on the long thin triangles — which have enormous bounding boxes and tiny coverage, and
which Captain Toad is currently *full* of.

### 2.3 The elephant: the workload itself may be pathological

Captain Toad's frame reports **2.9M depth-killed fragments against 520k drawn** — 85% of all
fill is thrown away — plus tens of millions of culled triangles. That is exactly what the
*known-broken* geometry would produce: long spiky triangles spanning the screen, massively
overdrawing each other.

**Fixing the vertex/transform bug may itself be the single largest performance win available,**
and it is already the top priority for other reasons. Do not spend a week micro-optimising a
rasteriser against a workload that is 85% garbage — re-measure once the geometry is coherent.
This is the most important sentence in this document.

---

## Phase 3 — Parallelism, where it is free or provably deterministic.

### 3.1 Free, no determinism risk (do these regardless)

- **framedbg's scrubber**: `RenderAfter(k)` for different `k` are independent replays on
  disposable scratch machines. Fan them out; a scrub drag becomes instant.
- **Verification harnesses**: the geomoracle/coloracle pattern is thousands of independent
  comparisons. Trivially parallel.
- **Checkpoint builds**: if a long run must happen, it can at least happen unattended (see
  Phase 4).

### 3.2 A tiled parallel rasteriser — the one place threading belongs

Bin triangles by screen tile; N workers, each owning a disjoint set of tiles; within a tile,
draws are applied in submission order. **This is deterministic** — tiles do not interact, and
order is preserved where it matters (blending reads the destination, but only its own tile's).
Join before anything can observe the buffer: the end of `GPU.Execute`, a `DisplayTransfer`, or
a CPU read of VRAM.

Expect near-linear scaling on a fill-bound frame. Do this **last**, because Phase 1 may make
each fragment cheap enough that ~4 cores of a fixed budget beats 8 cores of a slow one, and
because parallelising code you are still changing is a way to debug two things at once.

### 3.3 What NOT to do: concurrent CPU cores

Do **not** run the ARM11's threads, or the PS2's EE and IOP, on real OS threads concurrently.
Determinism is load-bearing across the whole repo — savestates as a regression instrument, the
frame debugger's replay, every oracle-verifies-Go-reimplementation check — and it rests on a
single instruction-count clock (`m.instrs` on the 3DS; the GX and DSP deadlines are compared
against it in one time base). Concurrent cores have no shared clock; making them reproducible
means lockstepping them, which serialises them again *and* adds synchronisation cost. The
emulated "threads" are green threads on one core, and that is the correct design here.

---

## Phase 4 — Attack the rebuild itself, not just its speed

Orthogonal to everything above, and possibly the highest value per hour:

- **Commit the recipe, not just the state.** A small file per checkpoint recording exactly how
  to reach it — image hash, `-keys` script, frame counts — plus a `make-state` command that
  replays it. The 2-hour run does not get shorter, but it becomes **unattended and
  reproducible**: kick it off, go and do something else, and never again reconstruct by hand
  what the sequence of button presses was.
- **Chain checkpoints.** Save every N frames during the long run. A fix whose effect begins
  late can resume from the last checkpoint that predates it. (Honest caveat: a *CPU* fix
  usually matters from the very first instruction, so this helps less than it sounds. A GPU or
  audio fix, though, often invalidates nothing before the first draw.)
- **Ask whether the state is actually invalid.** A CPU fix invalidates a state only if the bug
  it fixed had already corrupted the state that got captured. Sometimes it demonstrably has not
  — and the cheapest 2-hour run is the one you establish you do not need. The bit-exactness gate
  from 0.2 is what lets that question be answered instead of guessed.

---

## Summary of the ordering

1. **pprof one frame.** Ten minutes. It may reorder everything below.
2. **Benchmark + bit-exactness hashes.** The gate every later step reports against.
3. **Per-subsystem timing in framedbg.** The panel that answers "90% rasterising?" at a glance.
4. **Page-table the bus; give the rasteriser direct target access.** Cheap, safe, likely huge.
5. **Word-granular bus; instruction-decode cache.**
6. **Fix Captain Toad's geometry — then re-measure before optimising the rasteriser further.**
7. **Rasteriser inner loop** (incremental edges, tile rejection).
8. **Parallel tiled rasteriser; parallel scrubber and verification harnesses.**
9. **Never: concurrent emulated CPU cores.**
