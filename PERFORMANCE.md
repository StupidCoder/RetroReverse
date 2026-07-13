# Making the oracles faster

**Status (2026-07-13): done for the 3DS. The frame went from 256.8 ms to 105.7 ms — 2.4× — and
it is byte-identical.** Read *Phase 0 — the results* and *What was actually done* below; the
original plan (kept, below the line) guessed the ordering and got most of it wrong, which is
exactly what Phase 0 existed to find out.

A plan, ordered cheapest-and-safest first.

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

**256.8 → 105.7 ms/frame: 2.43× faster, with the frame byte-identical.** Both screens were
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

### What is left

- The bucket shares now: rasterise 47%, vertex + shader 27%, the derived remainder 16% (which
  now includes goroutine scheduling), GX transfers 4%, texture decode 3.5%.
- **Most draws are still serial.** 68 of Captain Toad's 143 draws per frame are too small to be
  worth goroutines. A persistent worker pool instead of spawning per draw would take the
  threshold down; ~1,100 goroutine spawns a frame is not free.
- **B3 (free parallelism outside the machine)** is not done: framedbg's `RenderAfter(k)` replays
  are independent and would make a scrub drag instant.
- **The ARM11 is 0.6% of the frame.** It stays that way. Do not build the instruction-decode
  cache or the word-granular bus.

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
