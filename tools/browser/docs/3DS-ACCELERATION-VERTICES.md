# D1a: reuse transformed vertices and decode attribute layouts once

This implements the scalar vertex work in
[D1 of the 25 percent campaign](3DS-ACCELERATION-25-PERCENT.md#d1-remove-repeated-vertex-work-and-use-exact-simd).
Experimental mode now reuses an earlier transformed vertex when the same index
appears again in the same draw. It also decodes attribute formats and padding
once per draw and reads verified RAM spans directly. Reference retains the
original byte-wise fetch and shader loop. The Go emulator is unchanged.

## Correctness boundary

The cache holds output indices, not another copy of every transformed vertex.
Two fixed tables of 65,536 32-bit entries consume 512 KiB. A generation changes
at every indexed draw, including after mode changes or state restores. Generation
wrap clears old entries. No program, uniform, memory or mapping version can be
reused across draws. Primitive order and the existing clipping/rasterization
paths remain unchanged. A vertex is remembered only after successful shader
execution and output mapping.

Source spans must fit an ordinary RAM region and use the same mapped region on
every covered page. A page overridden by another mapping, partial/unmapped
regions, address wrap or unsupported layouts retain the original path. Aliases
of ordinary RAM are safe within a draw because guest execution cannot modify
memory during vertex processing. Float attributes preserve their raw bits;
there is no arithmetic reassociation, float contraction or reduced precision.

The path is disabled for read/write/pixel observation, HID tracing, machine
profiling, verbose output, shader/draw tracing, census output and nonserial
execution. No trace event or read callback is skipped. Both fast paths are
bounded, disposable host state and absent from canonical save states.

The maintained implementation is [vertex-fast.h](../../platform/n3ds/browser/core/vertex-fast.h).
The [port-generator hooks](../../platform/n3ds/browser/portgen/vertices.go)
retain the original fetch loop for fallback. Optional D0 diagnostics now report
actual shader invocations and cache hits separately from submitted vertices;
sampled vertex times scale only over actual shader invocations.

## Measured result

The [full running-demo comparison](../results/3ds-acceleration-d1-demo.json)
passes all 14 runs: five alternating 300-interval trials per core, separate cold
runs, and separate timing/diagnostic runs. Every full run has the same canonical
state hash as D0 and the same sampled images and guest counters. The scene and
measurement method are unchanged. Only one worker executes at a time; the other
comparison instance remains paused. The browser is closed after the test suite.

| Ordinary trials | D0 | D1a |
| --- | ---: | ---: |
| Mean wall time per interval | 321.12 ms | 256.67 ms |
| Nominal display rate from total elapsed time | 5.19% | 6.49% |
| Range of trial means | 318.69–324.01 ms | 252.18–272.87 ms |
| Average of trial p95 values | 661.58 ms | 537.38 ms |
| Slowest 60-interval mean | 338.42 ms | 298.63 ms |
| Longest measured execution slice | 286.20 ms | 222.40 ms |
| PICA command/vertex time per interval | 99.05 ms | 31.92 ms |

Whole-scene elapsed time falls **20.1%**, equivalent to **1.25× throughput**.
The command/vertex stage is **3.10× faster**. All five candidate trials are kept:
the first spends 46.36 ms/interval in GPU round trips, compared with about 30 ms
in the later runs. That difference is measured; its cause is not established.

The diagnostic run shades 6,771,178 unique vertices and reuses 8,569,298 outputs
out of 15,340,476 submissions. It removes all 55.9% repeated shader invocations
in this workload. Sampled fetch/input time is about 6.24 ms/interval, shader time
25.18 ms, and clipping/setup 1.31 ms. These are diagnostic estimates and cannot
be substituted for the exact ordinary-trial stage timing. The separate timing
and diagnostic runs average 254.44 and 265.96 ms/interval, respectively; their
GPU round-trip time also differs, so that entire difference is not isolated
instrumentation overhead.

The fixed cache adds 512 KiB. Both cores' WASM heap capacities grow during the
repeated-restore workload; the largest observed capacities are 1,385,365,504 bytes
for D0 and 1,385,889,792 bytes for D1a. This records reserved heap capacity, not a
claim about live allocations or phone memory use.

The remaining ordinary-trial costs average 132.81 ms/interval in software
rasterization, 49.91 ms in ARM/Horizon and 33.50 ms in GPU round trips. Reaching
25% still requires approximately **3.85×** the current whole-scene throughput.
Shader compilation and the measured lighting/stencil/shadow families remain
necessary next steps; this result does not meet the final campaign target.

## Validation

The native differential fixture compares reuse alone, direct fetching alone and
both together against the generated Reference path. It covers 8/16-bit indices,
maximum indices, every scalar attribute format and component count, padding,
repeated attributes, fixed inputs, input/output mappings, signed zero, subnormals,
NaNs, infinity, shader and uniform edits, source writes, physical aliases,
overridden pages, generation wrap, observation hooks and shader errors.

The complete public browser suite and the vertex fixture under UndefinedBehaviorSanitizer
pass. AddressSanitizer could not execute this fixture on this host: both sandboxed
and unsandboxed runs spun in `__asan::InitializeShadowMemory` / dyld startup before
`main`. Those processes were stopped before timing; this is not an ASan pass.

The [60-interval handoff check](../results/3ds-acceleration-d1-handoff.json)
passes 32 actual Reference/Experimental switches, with matching canonical state,
pixels and guest counters at intervals 50 and 60. State reload and memory
inspection preserve continuation. It uses the existing deterministic button/stick
sequence and disables dense capture. It does not replace D6's 520-interval,
270-switch campaign or claim complete capture coverage.

The short real-scene preflight matches canonical state and sampled images.
Timing during that preflight is not acceptance evidence because other checks
were compiling concurrently. Sustained measurements use the same checkpoint and
production-worker harness as [D0](3DS-ACCELERATION-D0.md), after compilation and
other emulator tests finish.

This milestone does not yet enable SIMD or compile PICA shaders. Those remain
separate experiments, with exactness and whole-scene measurements required before
shipping. The 25% display-rate target remains open.
