# D3b: exact stencil rendering on WebGPU

The browser Experimental backend now accelerates general stencil operations,
including the directional-lighting subset introduced in D3a. Reference retains
the interpreter and software renderer. The Go emulator is unchanged.

## Supported operations and ordering

Stencil uses the existing paired RGBA8/D24S8 surface transaction. All eight
comparison functions and all eight operations are supported: keep, zero,
replace, saturating increment/decrement, invert and wrapping increment/decrement.
The comparison applies the read mask to both the reference and stored value;
updates apply the write mask and honor the stencil write gate. Color masks and
depth test/write settings remain independent. Stencil draws with depth testing
disabled still preserve and update the paired D24S8 surface.

Shading and alpha testing precede stencil comparison. A stencil failure selects
the stencil-fail operation; a subsequent depth failure selects the depth-fail
operation; success selects the pass operation. Rejection may therefore update
stencil without writing color or depth. Only a depth failure increments the
depth-killed counter. Ordered triangle bins retain primitive order for repeated
updates of the same pixel. Existing nonstencil draws retain early depth rejection.

Enabled textures must already be in the Reference decoded cache. Cold textures,
color/depth aliases, unsupported depth formats, shadow sampling and the remaining
lighting restrictions retain the Reference path. A refused stencil draw cannot
fall through to the older color-only fragment path. Existing observation and
numeric bounds remain in force.

Both surfaces and their counters commit only after a complete valid GPU result.
Validation, arithmetic rejection or device failure leaves canonical memory intact
and permits one Reference execution. The guest remains suspended during readback;
this change does not introduce deferred surface ownership or batching.

## Packet and pipeline contract

Graphics-stream schema 1 adds kind 9 for unlit stencil and kind 10 for lit stencil.
Kind 9 retains kind 7's 47-word triangles; kind 10 retains kind 8's 68-word
triangles and shared lighting LUTs. Each packet appends three stencil words before
the trailing overlap flag: the masked stencil-test register, the masked operation
register and a Boolean write gate. Parameter counts are 60 and 88 respectively.
The paired depth flags permit depth testing to be disabled for these two kinds,
but reject depth writes when testing is disabled.

Lighting specialization keys include the packet kind. Identical lighting settings
on kinds 8 and 10 must select different kernels because only kind 10 performs
stencil operations. The existing 64-entry cache and two pending-compilation limit
remain unchanged; generic kernels handle warmup and compilation failure.

## Validation

The public native fixture emits 989 packets, including 192 unlit stencil cases,
192 lit stencil cases and interspersed ordinary lit-depth draws. Hardware replay
matches all 987 supported packets byte for byte; the two existing ordered-overlap
cases remain refused. Coverage includes every compare and update operation,
read/write masks, zero/255 boundaries, alpha rejection, disabled depth testing,
Z/W depth, color masks and repeated overlapping triangles. Color, depth, stencil,
drawn counts and depth-killed counts all match Reference.

A fresh-device replay exercises both lit packet kinds after specialization,
including 193 kind-10 and 132 kind-8 specialized executions. Invalid interpolation
rejects kinds 6 through 10 transactionally. Native checks cover cold textures,
wrong depth format and color/depth aliasing even with depth testing disabled.
JavaScript checks reject malformed stencil metadata and invalid paired surfaces.

Public graphics report: [stencil replay](../results/3ds-acceleration-d3-stencil-graphics.json).

## Running-demo results

Chrome 154.0.8037.98 on the Apple M3 Pro / Metal 3 MacBook, AC power and low-power
mode disabled. Five alternating 300-interval comparisons start at the private
running-landscape checkpoint at interval 3240. The baseline is the shipped D3a
WASM core; both builds use the current JavaScript renderer, production worker,
both screens and idle input. Timestamps are disabled for the timed comparisons.
Only one emulator executes at a time, with no concurrent compilation or tests.

| Metric | D3a baseline | D3b stencil |
| --- | ---: | ---: |
| Mean elapsed per interval, five trials | 194.81 ms | 163.60 ms |
| Individual trial means | 193.90–195.13 ms | 160.47–173.97 ms |
| Nominal display rate | 8.56% | 10.19% |
| Mean trial p95 | 418.28 ms | 367.96 ms |
| Worst 60-interval window, maximum across trials | 210.75 ms | 195.63 ms |
| Longest execution slice, maximum across trials | 196.30 ms | 178.50 ms |
| ARM / Horizon | 49.61 ms | 49.87 ms |
| PICA commands / vertices | 11.82 ms | 11.77 ms |
| Remaining software rasterization | 65.31 ms | 11.71 ms |
| GPU upload / wait / readback | 58.80 ms | 80.43 ms |
| GPU input preparation | 1.11 ms | 1.79 ms |

Whole-scene elapsed time decreases **16.0%**, or **1.19× throughput**. The slower
third candidate trial is retained: its extra time is primarily in GPU round trips
(90.06 ms versus 77–79 ms in the other candidate trials). This is measured run
variation, not a discarded sample. The 25% target remains unmet and requires a
further 59% reduction from 163.60 to at most 66.7 ms per interval. No phone result
is inferred from these Mac measurements.

Each 300-interval trial adds 600 unlit and 8,662 lit stencil operations, bringing
the total to 20,610 GPU operations. Paired framebuffer uploads and readbacks rise
from 6.20 to 13.32 GB per trial in each direction; immutable input uploads rise
from 1.02 to 1.21 GB. All 17 lighting specializations become ready without errors.
The first candidate trial uses 13,107 specialized draws; later trials use 13,112.
No submitted GPU operation is refused. Maximum WASM heap capacity remains
1,385,889,792 bytes for both builds; this is capacity, not live memory usage.

All five candidate trials match baseline canonical state (`c44259c6…701b7391`),
sampled pixels and guest counters. The midpoint image shows Mario jumping along
the blue brick wall, continuing to the red platform at the end. The 30-interval
cold window measures 192.21 ms/interval for baseline and 158.54 for candidate;
backend initialization precedes that window, so these are not launch times.

Full report: [five comparisons and diagnostics](../results/3ds-acceleration-d3-stencil-demo.json).

The separate timestamp run averages 171.05 ms/interval. GPU execution is 35.42 ms
inside 67.88 ms of queue/map wait, with 13.07 ms upload, 0.49 ms submit and 2.59 ms
CPU readback copy. These are nested measurements, not independent totals to add
to the exclusive profile. The diagnostic run averages 171.63 ms. Small unlit draws
cost 3.52 ms/interval, lighting vertex-range fallbacks 3.01 ms, and other small lit
and stencil draws account for most of the remaining software work.

GPU compute is only slightly higher than D3a's 33.80 ms diagnostic result, while
transfers and synchronization grow substantially. The next rendering priority is
coherent batching with fewer uploads and maps; moving every small CPU draw to the
GPU individually would add more of the newly dominant overhead. ARM/Horizon at
about 50 ms also needs attention to fit the complete 66.7 ms budget.

The [handoff run](../results/3ds-acceleration-d3-stencil-handoff.json) makes 32
actual mode switches over 60 intervals with deterministic input. Reference and
mixed runs match checkpoint state, pixels, counters, memory inspection and
save/load. The mixed run executes 32 kind-9 and 244 kind-10 draws; Reference runs
no compiled vertex batches. Dense render capture is disabled because of the
existing [capacity limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit).
The D6 520-interval/270-switch/capture gate remains outstanding. All browser
harnesses close their browsers in `finally`; the post-run process audit finds
no remaining emulator test browser.

Release bundle: `c7828c4179a8c172075b` (111 pinned assets). Previous bundle
`ccb90e2984e7b6ae45b2` is retained. Only the 3DS core asset hashes change; the other
fifteen cores are unchanged. After benchmarking, the asynchronous specialization
closure was tightened to snapshot the packet kind along with its lighting flags;
this does not change execution of the immutable benchmark packets. Bridge fault
tests and the handoff test cover the final packaged source.

The complete public browser check suite and packaged-release audit pass,
including native graphics fixtures and the bridge's map, validation, timeout,
device-loss, specialization-failure and late-disposal cases.
