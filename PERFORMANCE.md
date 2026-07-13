# Making the oracles faster

A plan, ordered cheapest-and-safest first. Nothing here is started.

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
