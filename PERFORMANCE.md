# Making the oracles faster

**Status (2026-07-13): done for the 3DS. The frame went from 256.8 ms to 77.6 ms — 3.3× — and
it is byte-identical. The frame debugger's scrubber is 2.3× on top of that.**

**Status (2026-07-16): done for the GameCube. 2.8× on the intro cutscene, 3.1× on the shadow
scene, and 4.4× on a boot stretch — byte-identical. See *The GameCube* below.**

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

### The four things this taught, none of which the 3DS's write-up could have told us

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
