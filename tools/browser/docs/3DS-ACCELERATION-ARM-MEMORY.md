# D5a: guarded ARM word and halfword accesses

The browser Experimental backend now reads and writes ordinary RAM words with a
single host access. Instruction fetches and ARM data accesses previously called
the byte bus two or four times, repeating page lookup and observation checks for
each byte. Reference retains those original accesses. This change does not add a
CPU recompiler or change the Go emulator.

## Memory and debugger behavior

Every fast access checks the current indexed page and the backing region's
bounds. There is no retained pointer, decode cache or executable-memory epoch.
Aliases share their original backing bytes, an overriding page mapping takes
effect immediately, and a store to instruction memory is visible on the next
fetch. Clearing/reconstructing the page table takes the normal bus path.

Unindexed or partial pages, accesses crossing a page, unmapped addresses and
address wraparound use the original helpers. Read/write observers and HID tracing
also force the byte path, preserving byte events, their PCs and their order.
The fast path is controlled by the existing Experimental enable flag, so mode
handoff adds no state to materialize or invalidate. Graphics batches already
flush before guest ARM execution resumes.

ARMv6 unaligned data words retain their unaligned semantics. Older ARM variants
align then rotate loads and align stores. Aligned instruction fetch helpers and
halfword helpers keep their existing semantics. A distinct optional wide bus
retains its original routing; the browser CPU's wide-bus pointer remains null.
Unaligned host access uses `memcpy`, avoiding C++ alignment and aliasing undefined
behavior. Host byte order is handled explicitly.

## Validation

The public native fixture compares fast and Reference reads and complete memory
after writes across three ARM variants, every alignment, page and region edges,
the zero/high address boundary, physical aliases, replacement mappings, partial
regions and distinct wide buses. Observed cases compare every byte event and PC.
HID histogram counts remain identical. Actual ARM and Thumb sequences compare
registers, flags, PC, instruction counts and memory after loads/stores, including
an instruction overwritten and executed immediately. Reference mode and an
explicitly disabled fast path both refuse direct access.

## Running-demo results

Chrome 154.0.8037.98 on the Apple M3 Pro / Metal 3 MacBook, AC power and low-power
mode disabled. Five alternating 300-interval comparisons use the private running
landscape checkpoint at interval 3240, idle input, both screens and the production
worker. The baseline is the shipped D4a core; JavaScript is identical. Only one
emulator executes at a time, after compilation has finished. Ordinary comparisons
disable GPU timestamps. Private cartridges, checkpoints and scene images stay local.

| Metric | D4a baseline | D5a memory accesses |
| --- | ---: | ---: |
| Mean elapsed per interval, five trials | 138.53 ms | 124.07 ms |
| Individual trial means | 136.60–145.74 ms | 122.51–126.50 ms |
| Nominal display rate | 12.03% | 13.43% |
| Mean trial p95 | 304.54 ms | 270.96 ms |
| Worst 60-interval window, maximum across trials | 157.32 ms | 146.74 ms |
| Longest execution slice, maximum across trials | 169.10 ms | 167.90 ms |
| ARM / Horizon | 48.85 ms | 34.15 ms |
| PICA commands / vertices | 11.54 ms | 11.56 ms |
| PICA software bucket | 12.19 ms | 12.20 ms |
| GPU upload / wait / readback | 56.59 ms | 55.88 ms |
| GPU packet preparation | 1.64 ms | 1.64 ms |

Whole-scene elapsed time falls **10.4%**, or **1.12× throughput**. ARM/Horizon
time falls **30.1%**. The first baseline trial is slower than later trials; all
five are retained. The independent short comparison also improved, from 133.28
to 121.84 ms/interval. The cold 30-interval windows in the full campaign measure
141.08 and 125.92 ms/interval; backend creation precedes these windows.

Both builds execute the same 20,610 GPU operations and 8,508 submissions per
300-interval trial. Packet traffic and guest work are unchanged. GPU allocation
is 6,488,320 bytes in both builds; maximum WASM heap capacity is 1,385,889,792
bytes in both (capacity, not live use). The WASM binary grows by 3,202 bytes.
All ordinary, timestamp and diagnostic runs match complete canonical state
(`c44259c6…701b7391`), sampled images and guest counters. The inspected midpoint
still shows Mario jumping along the blue brick wall.

The separate timestamp run measures 126.61 ms/interval. GPU computation is
32.50 ms inside 50.02 ms queue/map wait, with 5.75 ms upload, 0.20 ms submission
and 0.97 ms CPU readback. The diagnostic run is 127.94 ms/interval, with 32.66 ms
GPU computation inside 50.54 ms wait. GPU-only time is nested inside the wait;
it must not be added again to the exclusive profile.

Report: [five comparisons and diagnostics](../results/3ds-acceleration-d5a-demo.json).

The **25% gate remains unmet**. Reaching 66.7 ms/interval still requires about
46% less elapsed time. GPU round trips now dominate at 55.88 ms, followed by
ARM/Horizon at 34.15 ms. Further GPU shader/dispatch work and CPU execution work
remain necessary. No phone speedup is claimed from this Mac benchmark.

## Reference handoff

A 60-interval run makes 32 actual mode switches with deterministic input and
compares against a separate Reference continuation. Canonical state, pixels,
guest counters, save/load and memory inspection match. Compiled PICA vertex and
kind-10 stencil work still execute in Experimental; Reference uses no compiled
vertex batches. Report: [handoff validation](../results/3ds-acceleration-d5a-handoff.json).

Dense render capture remains disabled for this private run because of the existing
[capture capacity limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit).
The public memory fixture covers watched byte semantics. This run does not claim
the outstanding D6 520-interval/270-switch/capture acceptance gate.

The full `tools/browser/check.py` suite passes, including native sanitizer cases
and release audit. Packaged release `1a35d31b741005e821a2` pins 111 assets; only
the 3DS WASM core changes among the sixteen cores. All owned test browsers were
closed after validation, with no remaining emulator benchmark processes.
