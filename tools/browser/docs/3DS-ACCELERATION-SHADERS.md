# D2: compiled PICA vertex batches

Experimental mode specializes PICA vertex programs into scalar WebAssembly. The
compiled module imports the emulator's existing memory and receives a whole batch
of unique vertices in one call. Uniforms, integer registers and the boolean mask
remain live inputs. Reference keeps its generated interpreter, including every
observation and debugger hook. No Go emulator code changes.

The maintained implementation is [shader-3ds.js](../../../site/emulators/shader-3ds.js),
[shader-wasm.h](../../platform/n3ds/browser/core/shader-wasm.h), and the existing
[draw generator](../../platform/n3ds/browser/portgen/vertices.go). The worker gives
each core its own compiler manager and includes compiler bytes in execution
identity. Browser release packaging pins that identity with the WASM core.

## Exactness and lifetime

Programs are keyed by all 4,096 instruction words, 128 operand descriptors and
the entry point. The native cache checks complete bytes after the hash; it never
relies on a hash alone. Compiler semantics are fixed by the owning module and its
release identity. Epoch changes invalidate the quick lookup, state restoration
invalidates the machine lookup, and uniforms are never baked into code.

Binary construction and WebAssembly compilation run after the guest call yields.
An immutable copy protects pending compilation from subsequent shader uploads.
Execution uses the interpreter until the compiled module is ready. Unsupported
instructions, backward jumps, recursive calls, excessive flow expansion and
resource limits all retain the interpreter. Compilation failure also leaves it
available. A compiler manager is disposed when its core is replaced; a pending
promise cannot install a module into a disposed manager.

Arithmetic preserves operation order and separate binary32 rounding. Dot products
and multiply/add instructions never contract into FMA. Ordered min/max retain the
Reference's tie and signed-zero behavior. Reciprocal square root uses binary64
sqrt/division followed by binary32 conversion, matching the existing helper.
Relative uniform accesses return positive zero outside the bank before applying
swizzle or negation. Calls, conditional ranges and inclusive loops preserve the
Reference's scope of END and address-register updates. Conversions outside the
supported integer range and arithmetic NaNs decline the batch instead of assuming
payload or conversion equivalence. A bounded loop-work guard can also decline it.

A draw may contain at most 32,768 submitted vertices and 8,192 unique inputs on
this path. Larger draws retain D1's interpreter/reuse path. Two scratch arrays
hold at most 4 MiB of vertex data together; index maps hold at most 160 KiB.
Their retained vector capacities can be larger than these live payloads. Up to 128 immutable
program images occupy about 2.1 MiB, plus browser-managed compiled code and its
immutable compilation inputs. These bounds describe the new caches, not the
emulator's total live or reserved memory.

No mapped draw output changes until the entire batch succeeds. Failure discards
scratch outputs and advances the vertex-cache generation before interpreting the
draw. It cannot reuse an output from an incomplete compiled attempt. Successful
batches map each unique output once and copy repeated indices in submission order.
Ordinary RAM and observation eligibility remain the same as D1. None of this host
state is serialized into a canonical save state.

## Validation method

The public [WASM differential fixture](../tests/shader-3ds.mjs) runs generated
modules against the actual C++/WASM interpreter, sharing identical input bits and
uniforms. Its 396 cases cover all implemented arithmetic layouts, partial writes,
swizzles, aliased registers, relative uniforms, comparisons, nested conditionals,
calls, loops, END scope, forward jumps, signed zero, subnormals, infinities, NaNs,
conversions, randomized programs, memory growth and manager disposal. It checks
all 64 output components bit for bit for each accepted vertex batch. Declined
numerical cases are counted separately, not called exact compiled results.

Run it with Emscripten using:

```sh
python3 tools/browser/tests/shader-3ds.py --emcc /path/to/em++
```

The separate [native batch fixture](../tests/shader-batch-3ds.cpp) supplies a test
bridge to the actual interpreter. It verifies transaction rollback after a
partially written scratch batch, output mapping/reuse, live uniform changes,
code/descriptor/entry invalidation, hash collisions, observation and size bounds.
It is included in the public browser suite and also runs under
UndefinedBehaviorSanitizer. The WASM oracle has only test exports; they are not
added to the shipped emulator.

A separate `build.py --shader-audit` build adds a raw-output comparison for every
accepted real-scene vertex against the WASM interpreter before mapping it into a
draw. It writes to `work/shader-audit`, leaves the production build untouched, and
is never packaged. This supplements state/pixel checks with intermediate shader
output equality; its slower timings are not used as performance results.

Performance acceptance uses the same private running-landscape checkpoint and
production worker as D0/D1: five alternating 300-interval trials, separate cold and
instrumented runs, full canonical end-state equality and periodic image equality.
The two comparison instances stay loaded, with only one executing at a time.
Compilation and other emulator tests finish before the timed campaign. Each
browser harness closes its browser in a finally block.

The compiler's `compileLatencyMs` is elapsed request-to-ready latency, summed
across programs; it includes asynchronous scheduling and is not exclusive compiler
CPU time. Shader counts and timings are disposable diagnostics. Batch fetch and
shader timings are measured for whole successful batches, while interpreter
vertex timings remain sampled estimates. Full command/vertex timing remains the
budget metric.

## Measured performance

The [production-worker comparison](../results/3ds-acceleration-d2-demo.json) uses
five warmed, alternating 300-interval trials per build, on the same MacBook and
running-demo checkpoint as D0/D1. The baseline is the shipped D1a core. It measures
all elapsed session time, including presentation, rather than summing selected
profiler rows. Private cartridge bytes, states, shader programs and images are
not included in the report.

| Ordinary trials | D1a baseline | D2 |
| --- | ---: | ---: |
| Mean wall time per interval | 252.95 ms | 232.65 ms |
| Nominal display rate | 6.59% | 7.16% |
| Range of trial means | 251.58–255.20 ms | 231.63–233.52 ms |
| Average of trial p95 values | 522.60 ms | 481.32 ms |
| Slowest 60-interval mean | 267.70 ms | 242.85 ms |
| Longest measured execution slice | 248.60 ms | 201.60 ms |
| PICA command/vertex time per interval | 31.96 ms | 12.14 ms |

This removes **8.03% of whole-scene elapsed time**, or **1.087× throughput**.
The command/vertex stage is **2.63× faster**, close to its 12 ms budget but still
0.14 ms above it. This does not meet the overall 25% target. Reserved WASM heap
capacity reaches 1,385,889,792 bytes in both builds during repeated restores;
that is not a measurement of live allocations or phone memory use.

Every ordinary candidate trial accepts 6,580,078 compiled vertex invocations in
54,188 batches. This is 97.18% of the 6,771,178 unique vertices; 300 numerical batch
refusals retain the interpreter. Seven observed program images compile, with no
unsupported programs or compilation failures. The cold run remains separate;
compilation latency is not hidden in a warmed-performance claim.

The remaining ordinary-trial costs average 131.65 ms/interval in software
rasterization, 49.93 ms in ARM/Horizon and 30.26 ms in GPU round trips. These three
now account for about 91% of whole-session time. Reaching 25% still needs roughly
**3.49×** the current throughput. Lighting and stencil/shadow coverage, GPU
synchronization and CPU work remain necessary stages of the campaign. No phone
performance result or SIMD benefit is claimed by this milestone.

All 14 performance runs pass canonical continuation and periodic image/guest
counter equality. The separate timing-only and diagnostic runs average 233.71
and 235.73 ms/interval. Their GPU timing also differs, so the complete difference
is not isolated instrumentation overhead. The diagnostic run retains D1's exact
counts: 15,340,476 submitted vertices, 6,771,178 shader invocations and 8,569,298
reuse hits. Batch/interpreter diagnostic totals are about 2.52 ms/interval for
fetch/input preparation, 2.80 ms for shader work and 1.46 ms for clipping/setup;
these have different scopes from the ordinary command/vertex budget.

The production core was rebuilt after adding the test-only audit switch and
retains the measured SHA-256
`c02c893ba8c4812d01e143ea7d5bd87e17c024d0db6dce3c87552bb03c99eebc`.

The [raw vertex audit](../results/3ds-acceleration-d2-audit.json) passes the same
300-interval scene, comparing all 6,580,078 accepted vertices with the WASM
interpreter before mapping. Its final canonical state and all sampled image/status
records equal the production campaign. This audit build is not shipped and its
timings are not included in the speedup above.

The [handoff test](../results/3ds-acceleration-d2-handoff.json) passes 60 intervals
with 32 actual mode switches, exact checkpoints/images/counters, save/load and
non-invasive memory inspection. Reference executes zero compiled vertices; the
mixed trial executes 320,363. Dense capture is disabled in this check, which does
not replace D6's longer 520-interval/270-switch and capture campaign.

All media-free public checks pass, including the native transaction fixture and
its separate UndefinedBehaviorSanitizer run. The shader WASM differential suite
passes 283 bit-exact cases and 113 deliberate numerical refusals. Final release
validation passes for **59463ed45f317abd722b** and its 110 pinned assets. Only the
3DS core changed; all other packaged core hashes are retained. The final process
check finds no leftover emulator test browsers.
