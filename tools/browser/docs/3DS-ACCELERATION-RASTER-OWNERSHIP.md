# D4c: skip hidden fragments and preserve sparse targets in place

This milestone changes browser WebGPU rendering only. Reference execution,
binary32 raster arithmetic, the WASM core and the Go emulator are unchanged.

**Prototype only: all three new/experimental options remain disabled by default.**
Five full-scene comparisons did not establish a speedup. The shipped D4b material
renderer remains active. The validated in-place path is retained for experiments
that can reduce submission boundaries, rather than only empty workgroups.

## Early stencil and depth rejection

Reference evaluates alpha before stencil. Moving stencil ahead of shading is
valid only when alpha cannot discard: alpha testing is disabled, or its compare
function is ALWAYS. Other alpha comparisons retain their original ordering.

For those eligible draws, coverage, barycentrics, reciprocal-W validation and
depth calculation still run first. The shader then performs the same stencil
comparison, masked fail/depth-fail/pass update and depth-rejection count. A rejected
fragment skips color/UV/quaternion/view interpolation, textures, lighting and TEV.
A passing fragment shades normally and never applies stencil a second time.

This does not skip guest-visible shader side effects. GPU eligibility already
excludes observers and shadow sampling, and paired raster textures must be warm
before packet creation. Texture and lighting inputs are immutable snapshots.
Arithmetic refusal remains transactional: no partial GPU output reaches guest RAM.

## In-place raster surfaces

Every raster invocation owns one color word and, for paired surfaces, its depth
word. It never reads another pixel's framebuffer value. All textures and LUTs are
separate immutable inputs. Ordered compute passes can therefore update a shared
target directly while retaining exact primitive order within each pixel.

Only raster kinds 6–10 use this binding. The unused read-only framebuffer binding
points at the other buffer, avoiding storage aliasing. Transfers and the older
fragment/stencil paths retain separate source/destination bindings and their own
cached bind group. Allocation replacement discards both cached binding variants.

Occupied-bin dispatch now leaves all unvisited pixels in the initialized target.
It no longer copies the complete color/depth pair before every sparse draw.
Partial edge tiles and empty draws retain their original bytes. Dense dispatch
remains available when fewer than 25% of bins are empty. Both in-place and sparse
options are immutable per backend, and remain separately configurable for tests.

Batch ownership is unchanged: guest CPU execution is suspended; incompatible
targets, dependent reads and Reference work still force a flush. One initial
surface upload and final readback surround the ordered passes. Every counter and
output length validates before canonical writes. A refusal discards the output;
the next request uploads its canonical starting surface again.

## Validation

The native stream adds 392 draws covering disabled alpha tests, all stencil
comparisons/operations, optional depth, write masks, lighting and overdraw.
Quarter-pixel geometry keeps stencil-NEVER cases on general raster instead of
the older integer stencil shortcut. The hardware replay matches all 1,379
supported operations byte-for-byte; two overlapping transfers correctly refuse.
All drawn/depth-rejection counters agree with Reference.

The material replay additionally checks ready compiled shaders, sparse/dense
ordered batches, partial edge tiles, empty draws and arithmetic refusal.
An ownership test alternates transfer/fragment and raster requests on one
backend, then injects arithmetic failure after a prior batch pass may have
modified the GPU target. Retrying from canonical input must produce exactly the
Reference bytes and counters. It tests sparse in-place, dense in-place and sparse
ping-pong configurations sequentially.

Reports: [public graphics](../results/3ds-acceleration-d4c-graphics.json),
[1,360 material/sparse records and 71 batches](../results/3ds-acceleration-d4c-materials.json),
[960 alternating requests and 12 failed-batch recoveries](../results/3ds-acceleration-d4c-ownership.json).

## Running-demo experiment and decision

Chrome 154.0.8037.98, Apple M3 Pro / Metal 3, Mac15,6, 12 CPU cores, 36 GiB RAM,
AC power and low-power mode disabled. Three configurations use identical D5a
WASM, the same private landscape checkpoint and the production worker with idle
input and both screens. Five 300-interval trials alternate forward/reverse order;
only one emulator executes at a time. Ordinary trials have timestamps disabled.

| Metric | Released D4b | Early rejection only | Early + sparse in-place |
| --- | ---: | ---: | ---: |
| Mean elapsed / interval | 119.85 ms | 127.70 ms | 121.65 ms |
| Individual trial means | 118.71, 119.07, 118.34, 121.55, 121.59 | 119.53, 118.12, 129.56, 136.28, 135.03 | 119.16, 117.65, 120.57, 120.47, 130.41 |
| Nominal display rate | 13.91% | 13.05% | 13.70% |
| Mean trial p95 | 267.16 ms | 298.88 ms | 285.68 ms |
| Worst 60-interval window, maximum | 138.51 ms | 153.58 ms | 149.14 ms |
| Longest execution slice, maximum | 168.90 ms | 168.10 ms | 169.00 ms |
| ARM / Horizon | 34.97 ms | 35.69 ms | 35.40 ms |
| PICA commands / vertices | 11.73 ms | 12.04 ms | 11.98 ms |
| PICA software bucket | 12.57 ms | 12.70 ms | 12.65 ms |
| GPU round trips | 50.29 ms | 56.89 ms | 51.39 ms |

Early rejection alone is **6.6% slower** in this campaign; the combined prototype
is **1.5% slower**, and wins only two of five comparisons. Later trials vary more,
but all are retained. Neither prototype earns a default change. The fresh control
is faster than D4b's earlier campaign; that cross-session difference is not a
new optimization gain. The 25% target still requires about 44% less elapsed time.

The combined prototype dispatches 7,876,556 workgroups instead of 30,555,000
(74.2% fewer) and performs zero preservation copies, removing the 11.25 GB of copy
payload from D4b's sparse experiment. Yet all configurations still perform 20,610
operations, 8,508 submissions and 7,158 raster batches per 300 intervals. They
upload 1,211,779,408 packet-input bytes and transfer 5,232,384,000 framebuffer bytes
in each direction. No GPU refusals occur; all 33 material variants compile.

Separate timestamp runs average 125.52 / 125.52 / 125.56 ms. GPU compute is
28.20 / 27.28 / 27.25 ms inside 44.42 / 43.69 / 43.54 ms queue/map wait. Uploads
are 6.23 / 6.25 / 6.32 ms and readback copies 1.09 / 1.13 / 1.12 ms. These single
diagnostic runs show only about 1 ms less GPU compute; do not add compute to its
containing wait or substitute these runs for the five ordinary trials. The combined
diagnostic run averages 125.73 ms.

Cold 30-interval windows average 123.10 / 122.65 / 121.76 ms. Backend creation
precedes the window; these are not launch times. Driver caches may be shared.
Maximum WASM capacity is 1,385,889,792 bytes and tracked GPU buffer/presentation
allocation is 6,488,320 bytes in every configuration; opaque driver shader-cache
memory and live heap use are not measured.

All full-window state hashes (`c44259c6…701b7391`), sampled images and guest
counts agree, including timestamp and diagnostic runs. The report records the
actual served JavaScript hashes before defaults were gated:
[complete comparison](../results/3ds-acceleration-d4c-demo.json).

The final default WGSL has identical tokens to released D4b; enabling the prototype
emits identical tokens to the tested early/in-place shaders. Only comments and
whitespace differ. There is no new rendering behavior in the default path.
For future reproduction, set `RASTER_BUILDS=early,candidate`, `DENSE_BUILD=early`
and `PINGPONG_BUILD=early` with the three-build benchmark. `BASELINE_SITE` selects
the first build's released graphics modules. Direct backend tests use
`earlyRejection`, `inPlaceRaster` and `sparseRaster` explicitly.

## Reference handoff and next work

The enabled combined prototype passes 32 mode switches over 60 intervals with
deterministic input, memory inspection and save/load. An injected invalid final
counter follows a completed multi-draw GPU batch; no partial bytes/counters reach
canonical RAM, Reference replay recovers, and continuation matches.
[Handoff and fault recovery](../results/3ds-acceleration-d4c-handoff.json).
Dense capture remains excluded because of the existing capacity limit; this does
not complete the D6 long-run/capture gate. No phone performance claim is made.

The next D4 experiment should target submission boundaries. Small Reference draws
account for roughly 8 ms/interval in D4b's diagnostic trace and often terminate
otherwise compatible batches. Trial a bounded small-draw admission policy only
inside the suspended-CPU ownership scope, using sparse in-place dispatch and the
same immutable recovery inputs. Compare total compute, maps, input copies and
software savings together. Raising the eight-draw cap alone is unlikely to suffice:
only 548 of 7,158 current batches reach it. Larger batches also need explicit
query/counter/input limits and cancellation checks.

ARM/Horizon still costs roughly 35 ms/interval against its 18 ms budget. Keep D5's
CPU/scheduler profiling and bounded decode/recompiler work on the plan; graphics
work alone has not demonstrated the remaining reduction.

## Release validation

The complete browser check suite passes, including native fixtures, JavaScript
checks, the C64 sanitizer checks and release audit. The final default backend also
passes the expanded 1,381-record hardware replay (1,379 exact operations and two
expected refusals): [default replay](../results/3ds-acceleration-d4c-default.json).
The ownership test was rerun with explicit checks that each selected binding and
dispatch path executed. All browsers close in `finally`, with no simultaneous
emulator execution between benchmark or test runs.

Release `e5e7c9477fc5a154b869` contains 112 pinned assets and retains D4b for
rollback. All sixteen core binaries and loaders are unchanged.
