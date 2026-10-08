# D4a: coherent PICA draw batches

The browser Experimental backend now groups compatible PICA draws while the ARM
CPU is suspended inside a command-list call. Reference keeps its original draw
order and software renderer. No Go emulator behavior changes.

## Ownership and boundaries

A batch holds at most eight raster packets and 8 MiB of immutable packet inputs.
All draws must have the same backing color/depth pointers, dimensions and paired
surface layout. Kinds 6–10 can participate; transfers and the older fragment-only
paths keep individual submissions. Two GPU buffers alternate input/output roles
across ordered compute passes. The host uploads the initial surface pair once,
then reads the final pair and every draw's counters with one map.

Batching is confined to `n3ds_GPU_Execute`. It flushes at command-list return,
surface changes, queue limits and before software fallback. PICA byte/word reads,
command-buffer copies (including chained lists) and direct vertex/index reads
check pending writes before reading. The guard checks each covered page and
compares backing pointers, so virtual aliases and overriding mappings inside a
larger region cannot bypass it. Unknown or wrapping ranges flush conservatively.
Framebuffer view construction itself does not read pixels and does not flush.

Decoded texture inputs retain Reference cache semantics. Cold textures still
follow the existing fallback rules. A decode that reads a pending surface flushes
first; a warm decoded texture remains its existing version, including the
Reference behavior of depth aliases. Per-draw cache invalidation still happens
in its original place. Neither guest CPU execution nor debugger observation can
see a pending batch. Read/write/pixel/PICA observers, command limits, capture,
tracing and profiling disable batching before it starts.

## Atomic completion and failure

Each draw has separate parameter, input and counter storage. A single submission
contains ordered compute passes; sharing a target never permits draw reordering.
All mapped output lengths, arithmetic refusal flags and per-draw drawn/depth-killed
counts validate before either canonical surface or any guest counter changes.
One failed draw rejects the entire batch.

The core retains immutable triangles, framebuffer/TEV/lighting state, registers,
lighting LUTs and the original decoded texture versions for each queued draw.
On failure, RAM still holds the pre-batch version. Reference replays those draws
once, in order, and merges their counters into the original GPU. It does not retry
the command list, rewind vertex processing, undo later register writes or restore
an obsolete live texture cache. Temporary texture objects remain owned until the
synchronous command-list scope ends. Snapshot state is bounded and excluded from
portable machine state.

The live bridge retains its exclusive-operation, timeout and quiescence handling
for both individual draws and batches. The Performance panel now reports logical
operations and actual submissions separately. Timestamp samples cover each
compute pass; their summed GPU time remains inside completion wait.

## Public validation

The new native fixture produces 32 continuous Reference draws across kinds 6–10,
changing stencil, depth, alpha, blend, color masks, lighting LUTs and decoded
texture versions. It forces every batch submission to fail and compares complete
RAM and drawn/depth-killed counts after ordered Reference replay. Dependency
variants use aliased byte/word reads, direct vertex access and command-buffer
range reads. Further cases cover an overriding alias page, unknown ranges,
software fallback, observers, capture recording and command limits.

Hardware replay tests 71 groupings of those draws with sizes 1, 2, 3, 4 and 8.
Final color/depth/stencil bytes and every draw's counters match Reference; input
snapshots remain unchanged. An invalid draw in the middle rejects the entire
batch and a subsequent valid batch recovers. Malformed sizes, layouts and packet
metadata are rejected before submission.

The existing hardware suite also matches all 987 supported packets among 989
records; its two ordered-overlap cases remain expected refusals. Mock device
tests cover batch mapping and validation failures, an invalid later counter and
an arithmetic refusal, checking cleanup and reuse after each failure.

Public reports:
[batch replay](../results/3ds-acceleration-d4-batch-graphics.json),
[existing packet regression](../results/3ds-acceleration-d4-graphics.json).

## Running-demo results

Chrome 154.0.8037.98 on the Apple M3 Pro / Metal 3 MacBook, AC power and low-power
mode disabled. Five alternating 300-interval comparisons start at the private
running-landscape checkpoint at interval 3240. The baseline uses the shipped
D3b WASM core; both builds use the updated JavaScript submission backend,
production worker, both screens and idle input. Timestamps are disabled in the
ordinary comparisons. Only one emulator executes at a time, with compilation
and other emulator tests finished before the timed runs.

| Metric | D3b baseline | D4a batching |
| --- | ---: | ---: |
| Mean elapsed per interval, five trials | 161.57 ms | 138.68 ms |
| Individual trial means | 160.50–162.32 ms | 138.02–139.34 ms |
| Nominal display rate | 10.32% | 12.02% |
| Mean trial p95 | 363.76 ms | 316.32 ms |
| Worst 60-interval window, maximum across trials | 182.69 ms | 156.61 ms |
| Longest execution slice, maximum across trials | 192.90 ms | 169.80 ms |
| ARM / Horizon | 49.50 ms | 49.50 ms |
| PICA commands / vertices | 11.81 ms | 11.67 ms |
| PICA software bucket, including recovery snapshot preparation | 11.67 ms | 12.46 ms |
| GPU upload / wait / readback | 78.74 ms | 55.22 ms |
| GPU packet preparation | 1.78 ms | 1.71 ms |

Whole-scene elapsed time decreases **14.2%**, or **1.17× throughput**. GPU round
trips save 23.52 ms per interval; retaining immutable Reference recovery inputs
adds CPU work to the software bucket. The 25% gate remains unmet: another 52%
elapsed-time reduction is needed to reach 66.7 ms per interval. This is a Mac
result; no phone speedup is claimed.

Both builds execute the same 20,610 GPU operations per 300-interval trial, with
the same distribution of draw kinds and no submitted GPU refusals. Candidate
submissions fall from 20,610 to 8,508 (**58.7% fewer**). Of those, 7,158 are raster
batch submissions, including one-draw batches at incompatible boundaries; the
remaining 1,350 are fills and display transfers. Framebuffer uploads and readbacks
fall from 13.32 to 5.23 GB in each direction (**60.7% less**). Immutable packet
input traffic remains 1.21 GB. GPU scratch/presentation allocation is about
6.49 MB versus 4.78 MB; both builds reach the same 1,385,889,792-byte WASM heap
capacity, which is not a measurement of live memory use.

All ordinary, timestamp and diagnostic runs match complete canonical state
(`c44259c6…701b7391`), sampled images and guest counters. The inspected midpoint
image shows Mario jumping along the blue brick wall; the endpoint continues to
the red platform. The 30-interval cold window measures 153.67 ms/interval for the
baseline and 139.86 ms for the candidate. Backend creation precedes that window;
these are not application launch times.

The separate timestamp run averages 153.37 ms/interval: 35.61 ms GPU computation
inside 58.68 ms queue/map wait, plus 6.15 ms upload, 0.33 ms submit and 1.15 ms
CPU readback copy. The diagnostic run averages 143.34 ms, with 32.51 ms GPU
computation inside 48.79 ms queue/map wait. Both runs are retained; this variation
does not change the five ordinary trial results. GPU-only time is nested inside
completion wait and must not be added again to the exclusive profile.

Full report: [five comparisons and diagnostics](../results/3ds-acceleration-d4-demo.json).

Batching addresses a major synchronization cost, but actual GPU computation
still exceeds the original 22 ms GPU budget. The next large opportunities are
ARM/Horizon execution and reducing shader/dispatch work. Small software draws
also split otherwise compatible batches; accelerating them should be judged
with batching enabled, since adding isolated submissions was previously costly.
Keeping surfaces dirty while ARM code executes remains outside this milestone.

## Handoff and injected failures

Four independent 60-interval runs each switch modes 32 times with deterministic
input, for 128 actual switches in total. Reference and mixed runs match checkpoint
state, pixels, guest counters, memory inspection and save/load. Each fault run
injects exactly one failure after a multi-draw GPU batch has completed: an explicit
rejection, an invalid later draw counter, or a truncated output buffer. The core
rejects the complete result and Reference recovery preserves continuation. All
runs still execute kind-10 stencil and compiled vertex work in Experimental;
Reference executes no compiled vertex batches.

Reports: [control](../results/3ds-acceleration-d4-handoff-none.json),
[rejection](../results/3ds-acceleration-d4-handoff-reject.json),
[invalid counters](../results/3ds-acceleration-d4-handoff-counter.json),
[truncated output](../results/3ds-acceleration-d4-handoff-length.json).

Dense render capture is disabled in these private runs because of the existing
[capture capacity limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit).
The D6 520-interval/270-switch/capture gate remains outstanding. Public fixtures
exercise observation guards separately. Browser harnesses close their browsers
in `finally`; the final process audit finds no emulator test browser still running.

Release bundle: `3fe586e1f73b5fc4bb83` (111 pinned assets), retaining previous bundle
`c7828c4179a8c172075b`. Only the 3DS core asset hashes change; the other fifteen
cores are unchanged.

The complete public browser check suite and the packaged-release audit pass.
