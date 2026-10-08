# D4b: compiled material programs and occupied-tile dispatch

This milestone changes the browser JavaScript/WGSL renderer. The WASM interpreter,
Reference renderer and Go emulator are unchanged. It reduces repeated fragment
program interpretation and evaluates dispatching only occupied parts of a surface.

## Compiled materials

Each validated fragment packet describes six TEV stages. The material compiler
emits straight-line WGSL for their source selection, operands, integer arithmetic,
scales and clamping. Delayed combiner-buffer updates retain their original order.
Fixed color-write, blend, alpha/depth/stencil and lighting modes are specialized
as well. The binary32 raster, interpolation and lighting helpers keep their exact
integer implementations; shader specialization does not introduce approximate math.

The key includes the packet kind and every compiled mode/stage word. Constant
colors, alpha/stencil references and masks, texture contents, lighting colors and
directions, LUT contents, geometry and addresses remain runtime inputs. Compiling
a key snapshots its parameters before yielding, so later mutation cannot install
a shader under an obsolete identity.

At most 64 material variants and 64 existing lighting variants are retained.
They share a maximum of two pending compilations. A pending, failed or uncached
material uses the existing generic/lighting kernel. Failed keys do not recompile
on every draw, and completion after disposal or device loss cannot install a
pipeline. No guest state or framebuffer ownership changes during compilation.

## Occupied tiles

The existing raster packet contains validated 16-by-16 triangle bins. The host
can upload the indices of nonempty bins and dispatch four 64-thread groups per
bin, mapping each group to an 8-by-8 Morton tile. Bounds guards cover surfaces
whose dimensions are multiples of eight but not sixteen. An empty list produces
one guarded group and no pixel writes.

Before a sparse dispatch, a GPU buffer copy preserves the complete previous
color/depth pair. The shader updates only occupied bins; no uninitialized output
or previous draw's scratch data can become visible. Dense draws keep the full
surface dispatch. The cutoff is fewer than 75% occupied bins. This path is disabled by default;
`create3DSGraphics({sparseRaster:true})` remains available for controlled tests.
The additional copies stay on the GPU and add no CPU upload/readback or map.

Batch ping-pong ordering, per-draw counters, all-or-nothing result validation and
Reference recovery remain unchanged. GPU timestamps measure compute passes;
the new buffer copies are included in completion wait but outside those compute
timestamps. End-to-end time determines whether sparse dispatch is worthwhile.

## Validation

The hardware material test executes every one of the 936 existing fragment/raster
Reference fixtures with a ready material shader, recreating bounded caches in
chunks so cache saturation cannot silently substitute generic coverage. It checks
complete surface bytes and drawn/depth-killed counts. Cases include six stages,
delayed buffers, operand transforms, blend equations/factors, alpha, depth,
stencil, lighting, changing uniforms, texture inputs and binary32 edge cases.

A new native fixture creates 32 continuous draws on a 72-by-88 surface, with 24
sparse draws interleaved with dense ones. It covers partial edge tiles and ordered
overdraw. Reference recovery compares full RAM after forced batch failures and
dependency flushes. Hardware tests replay those packets individually and in 71
batch groupings of sizes 1, 2, 3, 4 and 8. Surface bytes, per-draw counters and
immutable input snapshots agree. Two empty-draw cases preserve color/depth;
arithmetic failures still reject output across raster kinds 6–10.

Mock-device tests cover key identity, runtime uniforms, immutable compilation
snapshots, concurrent/cache limits, failed compilation and late disposal.
Report: [material and tile parity](../results/3ds-acceleration-d4b-materials.json).

## Performance experiment

The running-demo comparison uses identical D5a WASM cores with three JavaScript
configurations: the released renderer, material compilation with full-surface
dispatch, and material compilation with occupied-tile dispatch. Five 300-interval
trials alternate forward/reverse configuration order. Only one emulator executes
at a time. Every trial starts at the same private landscape checkpoint, with
idle input, both screens and the production worker. Runtime source hashes are
recorded per configuration. Ordinary trials disable GPU timestamps; separate
timestamp runs cover all configurations.

Chrome 154.0.8037.98, Apple M3 Pro / Metal 3, AC power, low-power mode disabled.
The selected default is **compiled materials with full-surface dispatch**. Its
shipping JavaScript hashes exactly match the measured `dense` configuration.

| Metric | Released D5a | Materials (`dense`) | Materials + tile skipping |
| --- | ---: | ---: | ---: |
| Mean elapsed / interval | 134.39 ms | 125.12 ms | 130.00 ms |
| Individual trial means | 126.57–150.65 ms | 122.49–131.64 ms | 124.33–137.21 ms |
| Nominal display rate | 12.40% | 13.32% | 12.82% |
| Mean trial p95 | 336.22 ms | 285.04 ms | 307.46 ms |
| Worst 60-interval window, maximum | 217.67 ms | 146.41 ms | 159.29 ms |
| Longest execution slice, maximum | 274.00 ms | 182.30 ms | 175.40 ms |
| ARM / Horizon | 35.64 ms | 36.14 ms | 40.00 ms |
| PICA commands / vertices | 12.18 ms | 12.22 ms | 12.46 ms |
| PICA software bucket | 12.85 ms | 12.86 ms | 13.03 ms |
| GPU round trips | 63.15 ms | 53.42 ms | 54.10 ms |

Materials alone reduce measured elapsed time **6.9%**, or **1.07× throughput**.
They are faster in four of five comparisons. Baseline GPU waiting varies widely,
including a 150.65 ms fourth trial; all trials are retained. The new baseline is
slower than the prior campaign's 124.07 ms result. Compare configurations within
this campaign rather than treating cross-session absolute rates as a speedup.
The 25% gate remains unmet, with roughly 47% less elapsed time still required.

All configurations execute 20,610 GPU operations and 8,508 submissions, with the
same per-kind counts, 7,158 raster batches and no GPU refusals. The selected
configuration compiles 33 material variants successfully; later trials execute
all 19,260 raster draws through them. CPU uploads/readbacks and packet inputs
are unchanged. Maximum WASM heap capacity is 1,385,889,792 bytes and tracked GPU
buffer/presentation allocation is 6,488,320 bytes in all configurations. These
figures exclude opaque driver shader-cache memory and do not measure live heap use.

Tile skipping reduces total workgroups from 30,555,000 to 7,876,556 (**74.2%**)
but adds 11.25 GB of GPU buffer-copy payload per 300 intervals. Its higher overall
time and ARM bucket show why workgroup count alone is insufficient evidence.
It is not enabled in the default renderer. No phone speedup is claimed.

The 30-interval cold windows measure 127.16 / 141.98 / 127.00 ms for released,
materials and sparse respectively. Backend creation precedes these windows;
these are not application launch times, and driver compilation caches may be
shared between configurations. All material compilations have finished by each
recorded window's end. Runtime compilation during the first scene traversal is
included in its timing.

Separate timestamp runs average 138.68 / 152.50 / 133.86 ms. Compute time is
38.98 / 38.15 / 29.48 ms inside 55.06 / 66.77 / 46.39 ms queue/map wait. The
material-only timestamp run is slower, so these individual diagnostic runs do
not establish a consistent GPU-compute improvement. The sparse diagnostic run
averages 133.28 ms with 28.27 ms compute inside 45.42 ms wait. All are retained;
none replaces the five ordinary trials in the performance claim.

Complete canonical state (`c44259c6…701b7391`), sampled images and guest counts
match across all configurations, trials and diagnostic runs. Report:
[three configurations and GPU timing](../results/3ds-acceleration-d4b-demo.json).

## Reference handoff

The selected default passes 32 mode switches over 60 intervals with deterministic
input. State, pixels, guest counters, memory inspection and save/load agree with
Reference. Report: [handoff](../results/3ds-acceleration-d4b-handoff.json).
Dense capture remains excluded because of the existing
[capture capacity limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit);
the D6 long-run/capture gate is still outstanding.

## Release validation

The complete `tools/browser/check.py` suite passes, including the native public
fixtures, JavaScript checks, C64 sanitizer checks and release audit. Release
`4fc3820cf0a153ff0834` contains 112 pinned assets; all sixteen core binaries and
their loaders are unchanged. The previous release remains available for rollback.
All owned browser test instances were closed after testing.
