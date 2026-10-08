# Reaching 25 percent display rate in the 3DS demo

Raise Super Mario 3D Land's running demo from the reported 6% to at least 25%
nominal display rate on the user's MacBook. This requires approximately **4.17×
the throughput**, or **76% less elapsed time for the same guest work**. With the
browser's 60 Hz nominal rate, the target is at least 15 display intervals/s and
at most 66.7 ms average wall time per interval. Aim for 60 ms to leave headroom.
Display intervals and unique presented game frames remain separate metrics.

The recommended sequence is to accelerate PICA vertex execution, cover the
expensive software draws, batch compatible GPU work, and reduce ARM/scheduler
cost. Small submission optimizations alone cannot meet this target. The numbers
below are engineering budgets and decision gates, not promised speedups.

Scope remains the browser C++/WASM core and its JavaScript/WGSL backend. Reference
remains available, with exact state and continuation across mode changes. The Go
emulator is outside scope. This is a new performance campaign after the completed
M0–M5 prototype and unlit depth extension; the original ARM debugger/recompiler
milestone remains relevant to the CPU stage below.

Implementation progress: [D0 running-demo baseline](3DS-ACCELERATION-D0.md),
[D1a vertex reuse/direct fetching](3DS-ACCELERATION-VERTICES.md) and
[D2 compiled vertex batches](3DS-ACCELERATION-SHADERS.md) are validated. The D2 campaign's five
alternating comparisons reduce whole-scene time from 252.95 to 232.65 ms/interval
(6.59% to 7.16% nominal). Commands/vertices fall from 31.96 to 12.14 ms/interval,
close to the 12 ms stage budget. SIMD and the remaining rendering/CPU stages are
still open; the 25% gate remains unmet.

[D3a directional lighting](3DS-ACCELERATION-LIGHTING.md) is now validated. Five
alternating comparisons reduce 231.61 to 194.11 ms/interval (7.20% to 8.59%
nominal), with exact state and sampled images. Software rasterization drops from
131.30 to 64.91 ms, while GPU round trips rise from 30.18 to 58.68 ms. General
stencil was the next D3 family.

[D3b general stencil](3DS-ACCELERATION-STENCIL.md) reduces 194.81 to 163.60
ms/interval in five alternating comparisons (8.56% to 10.19% nominal). Remaining
software rendering falls to 11.71 ms, while GPU round trips rise to 80.43 ms and
ARM/Horizon remains 49.87 ms. These last two buckets now dominate. D4 batching and
CPU work remain necessary; another 59% elapsed-time reduction is needed to meet
the 66.7 ms target. This milestone does not meet the 25% gate.

D3b's separate timestamp run measures 35.42 ms of actual GPU execution per
interval inside 67.88 ms of queue/map wait, plus 13.07 ms upload and 2.59 ms CPU
readback copy. Batching can attack the transfers and synchronization, but the
measured compute time alone exceeds D4's original 22 ms budget. Reprofile after
batching and reduce shader work/dispatch coverage as well, or explicitly rebalance
the budgets using measured CPU gains. Do not count batching as sufficient on its
own to reach 25%.

[D4a coherent draw batches](3DS-ACCELERATION-BATCHING.md) reduce 161.57 to 138.68
ms/interval in five alternating comparisons (10.32% to 12.02% nominal). GPU
submissions fall 58.7% and paired framebuffer traffic falls 60.7%; total GPU round
trips drop from 78.74 to 55.22 ms. ARM/Horizon remains 49.50 ms. Batches stay inside
suspended-CPU command-list execution, with dependency flushes and whole-batch
Reference recovery. Further CPU and GPU compute work is required; another 52%
elapsed-time reduction is needed for the 66.7 ms target.

[D5a guarded ARM memory accesses](3DS-ACCELERATION-ARM-MEMORY.md) reduce 138.53
to 124.07 ms/interval in five alternating comparisons (12.03% to 13.43% nominal).
ARM/Horizon falls 30.1%, from 48.85 to 34.15 ms, with exact canonical state and
images and 32 verified Reference/Experimental switches. Page boundaries, partial
pages and observed accesses keep the byte bus path. GPU round trips now dominate
at 55.88 ms; another 46% elapsed-time reduction is still needed for 25% nominal.

## What the new profile establishes

The October 8 screenshot shows Mario running through the landscape on the top
screen. The seven exclusive buckets total 522.85 ms **over a UI update window**,
not one frame. The reported 6% rate is rounded; use a fixed-work benchmark to
establish the actual baseline before evaluating changes.

| Work | Window time | Share |
| --- | ---: | ---: |
| PICA commands and vertex processing | 195.30 ms | 37.4% |
| Remaining software rasterization | 177.20 ms | 33.9% |
| ARM11 and Horizon scheduler | 87.60 ms | 16.8% |
| WebGPU upload, wait and readback | 60.50 ms | 11.6% |
| DSP, GX and GPU input preparation | 2.25 ms | 0.4% |

PICA command/vertex work and software rasterization together consume 71.3% of
the measured window. GPU timings show 28.64 ms of compute **inside** 52.23 ms of
completion wait, across 46 operations. The 0.99 ms CPU readback copy is about
0.2% of the exclusive total. The remaining wait includes transfer, scheduling
and driver costs; it is not an isolated readback measurement. Hardware timing
collection was active in this screenshot.

Assuming these shares represent the same sustained workload:

- Removing all existing WebGPU wall time would improve 6% only to about 6.8%.
- Removing all software raster time without adding GPU cost would reach about
  9.1%; removing all command/vertex time would reach about 9.6%.
- Even removing both command/vertex and software raster time entirely would
  reach only about 20.9% if everything else stayed constant.
- Unchanged ARM/scheduler time alone would consume about 70% of a total budget
  that is 24% of today's time. CPU optimization is likely necessary.

These are optimistic upper bounds, not predictions. New GPU coverage adds GPU
work, and the screenshot excludes idle and display copies. Acceptance measures
the whole session, including presentation, rather than adding profiler rows.

## Current implementation and opportunities

The [draw and shader code](../../platform/n3ds/browser/core/generated.cpp) fetches
attributes and calls `n3ds_GPU_shaderRun` for every submitted vertex, including
repeated indices. Shader instructions are already decoded and cached with
`shEpoch`; caching their decode again will not remove interpreter dispatch.
The combined PICA bucket also contains command handling, vertex fetching,
clipping and setup, so its entire 37.4% cannot yet be assigned to the shader.

The [GPU eligibility gate](../../platform/n3ds/browser/core/graphics-fragments.h)
now supports one directional light with normal maps and shared lighting LUTs,
plus general stencil with or without depth testing.
It still refuses other lighting combinations, shadow sampling
and several bounded workloads. The earlier [native census](../results/3ds-acceleration-depth-census.json)
identifies plausible families to investigate, but is a different checkpoint and
predates depth acceleration. Select new coverage by **current fallback time**,
not draw count or the old census alone.

The [live bridge](../../../site/emulators/graphics-live-3ds.js) now submits bounded
batches of compatible draws within a command list, flushing before dependent
reads, software fallback and guest CPU execution. The
[compute rasterizer](../../../site/emulators/raster-3ds.js) uses ordered triangle
bins and exact binary32 helpers; the host dispatches across the full surface.
The [build](../../platform/n3ds/browser/build.py) already uses `-O3` and disables
float contraction, but does not enable WASM SIMD. The browser worker-pool adapter
is serial. Each is a concrete optimization opportunity, not evidence of its
eventual speedup.

## D0 Establish the actual running demo benchmark

Save a private checkpoint with Mario actively traversing the landscape, plus a
deterministic input sequence that keeps the demo running. Verify the start,
middle and end visually and record changing positions/images. Do not substitute
the welcome screen, Mii dialog or an animated title merely because those are
the existing fixtures. A transition into a cheaper scene invalidates the run.

Record the Mac model, power mode, browser/version, adapter, release identity and
checkpoint hash. Run five alternating baseline/candidate trials of at least 300
display intervals, extending the workload to cover the demo's heavy section.
Separate shader cold-start results from warmed playback. Keep the app foreground
and include both screens and the actual worker/presentation path. Use timestamps
only in separate diagnostic runs; compare Performance-panel-open results too.

Split PICA into command parsing, attribute/index reads, vertex shader, clipping
and triangle/bin setup. Record total/unique indices, shader hashes and invocation
counts. Split fallbacks by rejection reason, feature combination, pixels covered
and exclusive CPU time. Record transfer bytes, maps, submits and GPU time per
**emulated interval**, plus slice lengths, presentation and unexplained wall time.
Measure instrumentation overhead and keep counters out of canonical state.

Acceptance: reproducible scene and throughput, median/p95 interval times, identical
guest work, and enough attribution to rank the next changes. Instrumentation must
not enable capture/read/write hooks that disable acceleration. Extend
[the live harness](../tests/graphics-live-3ds-browser.mjs) to sample pixels at
multiple boundaries and report interval distributions; its current final hash
and aggregate timing alone do not establish sustained 25% demo performance.

## D1 Remove repeated vertex work and use exact SIMD

Start with a bounded cache of transformed vertices **within one indexed draw**.
Key by the resolved vertex index under that draw's fixed program, uniforms,
attribute bindings and mapping. Reference resets shader output and temporary
state per invocation. Reuse therefore has a plausible exact path, but must be
disabled for observable read/trace hooks and any nonordinary memory access.
Retain submission order, errors and canonical counters. Discard the cache at
the draw boundary; cross-draw reuse has much broader invalidation requirements.

Decode attribute layout once per draw. Add guarded direct RAM reads where the
whole range and mapping are proven safe; retain the existing path for aliases,
boundaries and inspection. Measure this independently of shader execution.

Try ordinary WASM SIMD for component-wise shader operations, vertex fetch and
remaining CPU kernels. Preserve the order of dot-product additions, separate
multiply/add rounding, signed zero, NaN behavior and float-to-int conversions.
Use exact helpers where a SIMD instruction differs. Emscripten enables SIMD with
[`-msimd128`](https://emscripten.org/docs/porting/simd.html); do not enable relaxed
SIMD or fast-math as an implicit correctness change. Retain a compatible scalar
build/path when SIMD is unavailable.

Acceptance: bitwise vertex outputs and continued machine state match, including
duplicate indices, shader errors and memory boundaries. Publish cache hit rate,
vertex-stage and whole-scene speedups separately. If reuse is low or fetch is a
minor cost, stop tuning it and move directly to D2.

## D2 Compile hot PICA vertex shaders to WASM

Translate frequently executed PICA programs into specialized WASM functions that
read the existing canonical uniforms and write the existing output layout. This
removes per-instruction dispatch and swizzle/mask decoding from each invocation,
while keeping clipping and raster packet construction in their current places.
Invoke a compiled function for a **batch of vertices**, not through a JavaScript
callback per vertex. Use shared WASM memory and a validated function-table or
equivalent batch ABI. Compile outside the guest-critical path; interpret until a
bounded, fully validated module is ready.

Cache by shader code, operand descriptors, entry point and compiler semantics.
Uniforms stay runtime inputs unless specialized values are included in the key.
Honor shader epoch changes, state load, masked/relative writes, condition flags,
call/loop semantics and runaway limits. Preserve helper behavior for reciprocal,
reciprocal square root, ordered min/max and conversions. Unsupported programs
stay in the interpreter before committing output; do not replay partial effects.

WASM is the first target because the next stage currently consumes CPU vertices.
A GPU-only vertex stage followed immediately by vertex readback could replace
interpreter overhead with another synchronization point. Reconsider GPU vertex
execution together with GPU clipping/binning if D2 misses its budget.

Acceptance: public instruction/control-flow fixtures, differential randomized
inputs, actual scene shader outputs and full continuation all match. Target a
**combined command/vertex/setup budget of 12 ms per display interval** after D1
and D2. If the remaining cost is clipping or command parsing, optimize that measured
component rather than reporting shader-only success as meeting the budget.

## D3 Move the expensive fallback families onto the GPU

Use D0's ordered list of fallback time. Likely candidates are lighting, general
stencil/depth interactions and projection/shadow textures; implement the largest
measured family first. Each family is its own validated commit. A draw can require
several of these features, so identify the complete combinations needed to remove
the expensive passes rather than adding isolated features that leave them blocked.

Extend immutable input packets with the required quaternion/view attributes,
lighting state and LUTs, stencil operations, shadow configuration and counters.
Preserve the Reference order of texture access, alpha, depth and stencil effects,
including its observable texture-cache behavior and render-target aliases. Retain
the cold-texture protection introduced by the depth milestone.

Specialize compute pipelines for actual TEV, lighting and framebuffer state to
remove unused work. Dispatch only nonempty tiles, preserving primitive order and
untouched pixels; keep numeric operation order when hoisting expressions. Use
stable decoded-texture versions to avoid redundant upload where safe. Compile
asynchronously with a bounded cache and exact fallback during warmup.

Plain WGSL float replacement is not an exact shortcut: the
[WGSL specification](https://www.w3.org/TR/WGSL/#floating-point-evaluation) permits
reassociation/fusion and varying accuracy. Extend exact arithmetic or prove a
bounded exact specialization. Lighting's normalization and LUT edges need their
own tests. A conventional hardware raster pipeline is an isolated research option
if compute proves too expensive; its coverage and numeric equivalence must be
established before enabling it in the same continuity mode.

Acceptance: color, depth, stencil, counters and continuation match for each
enabled feature combination and alias case. Target **at most 4 ms remaining
software raster time per interval**, while measuring the extra GPU time. Moving
177 ms from one profiler row into another is not a speedup. Keep small draws on
the CPU when their total cost is lower and correctness is unchanged.

## D4 Batch compatible GPU draws behind a coherent boundary

Once expensive fallback boundaries are rare enough, replace one wait per draw
with bounded batches. Initially batch within a command-list segment, keep the
guest CPU suspended, and materialize every modified surface before returning.
CPU-side command parsing and vertex preparation can still read prior targets:
declare their reads, writes and physical aliases and flush before a dependency.
GPU-supported producer/consumer passes may remain ordered inside the batch.

Retain color/depth/stencil targets in GPU buffers across compatible draws. Upload
each required version once and read each modified range back once per batch.
Use ordered dispatches and transfers; sharing a target is not permission to reorder
blending, stencil, sampling or guest-visible completion events. Refuse unknown
dependencies and flush before software fallbacks, CPU reads, inspection and save.
Cap batch work and queue depth so Pause and cancellation remain responsive.

A batch must commit coherently. Keep its pre-batch canonical state plus the bounded
input history needed to recover if validation, mapping, timeout or device loss
fails. Restore speculative counters/cache changes and replay exactly once. The
current per-operation fallback cannot simply be reused after partial batch effects.

Only consider keeping dirty GPU data across guest CPU execution after auditing
all bus helpers **and direct RAM views**, including HLE and display consumers.
That larger ownership redesign is optional; it is not needed to trial suspended
command-list batching. Buffer mapping remains an explicit synchronization boundary.

Acceptance: fault injection at every batch phase, read/write aliases, overlapping
transfers, fallback ordering, delayed GPU, save/load and repeated handoff. Target
**at most 22 ms total GPU critical-path cost per interval**, including the new
D3 work, uploads, waits and materialization. Report maps and bytes saved per
interval; the screenshot's 46 operations are per update window, not per frame.

## D5 Reduce ARM and scheduler cost if the budget still requires it

Reprofile after the graphics changes. First separate ARM execution, scheduler/HLE,
memory helpers and Asyncify overhead. Try guarded fetch/data fast paths, cached
instruction decode and measured HLE improvements. Preserve bus behavior, event
boundaries, input ordering and the existing interpreter semantics. Audit Asyncify's
instrumented call graph before narrowing it; graphics can suspend inside execution.

If this does not reach **18 ms per display interval** for ARM/Horizon, implement
the [original M6](3DS-ACCELERATION-PLAN.md#m6-arm-debugger-and-bounded-wasm-recompiler)
bounded ARM/Thumb-to-WASM recompiler. Start with reference register inspection,
stepping and breakpoints; compile the hot instruction subset only. Preserve
flags, VFP state, instruction budgets, scheduler exits, exceptions, memory effects,
executable-write invalidation and reference deoptimization. Do not assume speeding
ARM instructions will remove scheduler/HLE time.

Acceptance: block and instruction differential tests, self-modifying code and
aliases, interpreter stepping followed by compiled resume, and all CPU/renderer
combinations. If the 25% target is already sustained, defer the recompiler rather
than expanding this campaign unnecessarily.

## Combined budget and completion gate

| Component | Target per display interval |
| --- | ---: |
| ARM11 and Horizon | 18 ms |
| PICA commands, vertices, clipping and setup | 12 ms |
| Software raster fallbacks | 4 ms |
| GPU work, upload, synchronization and materialization | 22 ms |
| DSP, input preparation, presentation and remaining overhead | 4 ms |
| Total planning budget | **60 ms** |

The combined budget is intentionally demanding. It includes migrated GPU work
and does not multiply unrelated microbenchmark gains. At the reported baseline,
unchanged ARM cost would be roughly 47 ms/interval if screenshot shares are
representative; the CPU allowance therefore requires a substantial reduction too.
Replace all estimates with D0 measurements and rebalance after every stage. If
exact lighting or vertex compilation cannot meet its allowance, run an isolated
GPU vertex/clipping or specialized raster experiment before claiming 25% is
achievable. Approximate rendering would require a separate explicit decision;
switching to Reference later cannot undo already divergent guest state.

D6 completes when five warmed trials of the verified running demo each sustain
at least 15 display intervals/s, the median interval is at most 66.7 ms, and p95
is at most 100 ms. Report the full-window average and worst 60-interval window,
longest uninterruptible slice, cold-start latency and peak memory. Require exact
periodic state and color/depth/stencil comparisons, matching guest instruction
and draw work, and successful continuation. No frame skipping, disabled effects,
reduced emulated resolution or synthetic frame-counter advancement counts.

Repeat at least 520 intervals and 270 actual backend switches in the running
scene, with memory inspection, save/load, input, cancellation, reset and device
loss. Preserve complete capture/replay checks in the available bounded windows.
The [existing dense capture limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit)
is a separate debugger-capacity issue: overflowing captures must remain explicitly
incomplete and must not corrupt subsequent save/resume. Do not present a
capture-disabled endurance run as complete capture coverage. Phone testing is
a separate result, with no extrapolated Mac speedup claim.

## Delivery

Commit and push this planning milestone, then D0–D6 as separately reviewable
milestones when implemented. Split D3 by feature family and D4 by ownership scope.
After each implementation milestone publish matched throughput, attribution,
correctness evidence and the revised remaining budget. Stop and reconsider an
approach when the whole-scene result misses its gate; do not keep adding features
on the assumption that their local speedups will eventually multiply to 4.17×.

Keep maintained browser acceleration modules separate from generated code and
preserve their hooks in the browser port generator. The generator's Go language
does not make it the Go emulator. Stage only scoped changes, rebuild/package only
the 3DS core or site assets as required, run relevant public and hardware tests
and the release audit, then commit and push `main`. Keep private cartridges,
checkpoints and captured resources local. This plan changes no runtime behavior
and requires no emulator rebuild or release-bundle rotation.
