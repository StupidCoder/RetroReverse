# GPU submission costs and on-demand timing

The subsequent Galaxy S26 Ultra profile reports 3.92 ms / 0.8% CPU graphics
preparation and 438.90 ms / 88.3% WebGPU upload/wait/readback. The CPU bottleneck
has moved, but this aggregate includes shader execution as well as host overhead.
It does not establish that readback alone is responsible. These values cover a
UI update window and cannot be converted into a frame rate or compared directly
with a different update window.

## Changes

- Compile six operation-specific pipeline variants using a WGSL override and a
  common explicit bind-group layout. The compiler can remove the rasterizer from
  fills/copies/display conversion. All variants finish initialization before Play;
  no guest-program shader cache or compile backlog is introduced.
- Request error-scope completion and readback mapping together after submission.
  Both must succeed before returning any output to WASM. A mapping failure cancels
  or releases the mapping before buffers can be reused. This removes a serialized
  validation wait without weakening the commit boundary.
- Collect optional GPU timestamps in the same mapped readback allocation as
  pixels and the drawn counter. There is one map per operation, including when
  measuring GPU execution.
- Handle zero/one and power-of-two binary32 arithmetic with exact identities and
  exponent adjustment. Underflow still uses explicit ties-to-even rounding.
  Skip identity TEV stages while preserving delayed buffer updates. Across 21 captured
  welcome draws, 42 of the 126 declared TEV stages are identities. This is a
  stage census, not a claimed percentage speedup.

Every operation still returns its output to canonical RAM before WASM resumes.
No GPU-resident dirty memory, deferred guest writes, overlapping guest execution,
frame skipping or approximate interpolation is introduced. This is a bounded
reduction in submission and shader work, not the larger batching redesign.

The API contracts used here are documented in
[asynchronous error scopes](https://developer.mozilla.org/en-US/docs/Web/API/GPUDevice/popErrorScope)
and [compute pipeline constants](https://developer.mozilla.org/en-US/docs/Web/API/GPUDevice/createComputePipelineAsync).

## Reading the phone's next profile

Open **Session / controls → Performance** while running Experimental. The added
WebGPU breakdown shows upload/submit, completion wait and CPU readback-copy time.
When the adapter supports timestamp queries, it also shows GPU compute time and
how many operations were sampled. Compute time is **inside** completion wait; do
not add the two or interpret them as exclusive percentages. The timestamp
interval excludes GPU buffer copies outside the compute pass.

GPU timestamp recording starts when Performance opens and stops when it closes.
A partial first update can contain fewer timestamp samples than completed
operations; both counts are shown. The remaining host timings remain available
without timestamp support. Timestamp collection does not pause guest execution,
change machine state, or enable Reference instrumentation hooks. Closing the
panel avoids its query/resolve overhead during normal play.

A large compute component points toward shader work. A much smaller compute
component points toward submission, copies, scheduling and synchronization.
Completion wait includes those costs and cannot isolate individual driver or
browser components on its own. Larger batching or GPU-resident render targets
would still require auditing CPU and fallback memory accesses before relaxing
per-operation readback.

## Evidence

Matched Chrome 154 / Apple Metal 3 measurements use three warmed, alternating
30-interval Reference/Experimental trials per scene, including presentation.
The previous release `9909cce5d24c89960501` is rerun with the same harness and
checkpoints. GPU timestamp recording is off in these performance runs.

| Browser checkpoint | Previous Experimental | Updated Experimental | Time reduction |
| --- | ---: | ---: | ---: |
| Welcome | 48.44 ms | 45.59 ms | 5.9% |
| Title/Mii dialog | 57.68 ms | 55.37 ms | 4.0% |

These are median milliseconds per display interval. This is a modest Mac gain,
not a claimed fix for the phone or a full-speed/gameplay result. Full saved state
and final pixels match Reference and the previous release in every trial.

A separate six-interval diagnostic run measures 157 operations: 18.60 ms
upload/submit, 145.10 ms completion wait, 3.10 ms readback copy and 63.83 ms GPU
compute within that wait. Both measured and deliberately delayed runs preserve
exact state and scanout. This diagnostic is not part of the performance table;
GPU queries and changed host scheduling can affect execution time.

The arithmetic suite passes 449,472 comparisons, now including every normal and
subnormal power-of-two exponent with mixed signs, underflow and overflow. Public
replay matches all 307 supported packets and retains both expected overlapping
copy fallbacks. The updated fixture exercises identity TEV stages followed by
buffer reads, including independent delayed RGB/alpha updates. Host unit tests
exercise validation and mapping rejections, cleanup/reuse and overlapping promise
completion, plus immutable timing snapshots and zero-duration query samples.
Timestamp-enabled device loss and timeout recover to Reference with identical
state and pixels. The packaged UI verifies opening/closing Performance changes
hardware measurement without pausing Play, alongside capture, memory inspection,
portable state, cancellation and no-WebGPU fallback. The sustained welcome test
passes 520 intervals per sequence, 270 engine switches and five complete captures
with matching saved states, pixel histories and replay output. The complete public
browser suite and release audit pass for `680f5367cab4ac6c7c94`; all sixteen
packaged WASM cores are unchanged.

Reports: [previous welcome](../results/3ds-acceleration-submission-baseline-welcome.json),
[updated welcome](../results/3ds-acceleration-submission-live-welcome.json),
[previous title](../results/3ds-acceleration-submission-baseline-title.json),
[updated title](../results/3ds-acceleration-submission-live-title.json),
[timing and delayed continuation](../results/3ds-acceleration-submission-timing.json),
[arithmetic](../results/3ds-acceleration-submission-float.json), and
[public replay](../results/3ds-acceleration-submission-public.json),
[recovery](../results/3ds-acceleration-submission-recovery.json),
[UI](../results/3ds-acceleration-submission-ui.json), and
[sustained handoff](../results/3ds-acceleration-submission-handoff.json).

The existing reproduction commands in [GPU rasterization](3DS-ACCELERATION-RASTER.md)
apply. Set `MEASURE_GPU=1` on `graphics-live-3ds-browser.mjs` for hardware timestamp
diagnostics. Omit it for performance comparisons. The CPU/WASM build is unchanged;
packaging uses `python3 tools/browser/package.py --site-only`.
