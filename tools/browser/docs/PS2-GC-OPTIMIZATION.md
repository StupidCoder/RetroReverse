# PS2 and GameCube optimization pass — 2026-09-29

The new browser cores reach the same emulated checkpoints and produce the same
RAM/framebuffer hashes as the previously published `d0c82546` build, with less
host execution time. There are no changes to instruction budgets, rendering
resolution, floating-point contraction, game compatibility workarounds or the
capture/replay contract.

## Measured improvement

These are median times for **60 emulated video fields**, not 60 unique rendered
game frames. Each variant ran three times in a fresh Node 24.19.0 process on the
same macOS arm64 machine. Before/after runs were sequential and their order
alternated. Capture was disabled and subsystem profiling remained enabled.
Boot, state import, final proof calculation and browser canvas delivery are
outside the timed interval. Each run restored an identical private checkpoint.

| Scene | Previous WASM | New WASM | Throughput gain | New fields/s |
|---|---:|---:|---:|---:|
| Luigi's Mansion, foyer | 16.47 s | 13.12 s | 25.6% | 4.57 |
| Jak and Daxter, 3D opening scene | 15.61 s | 10.55 s | 48.0% | 5.69 |
| Jak and Daxter, early title/logo | 4.47 s | 3.98 s | 12.5% | 15.09 |

Full samples, core hashes and output proofs are in
[`ps2-gc-optimization.json`](../results/ps2-gc-optimization.json).
The older checkpoint filename `gameplay-go.state` refers to the 3D opening
scene; it does not establish free-roaming gameplay compatibility.

## Changes

### GameCube

- The TEV interpreter resolves color/alpha operand selectors once per draw.
  Fragment evaluation reads the resolved inputs directly instead of dispatching
  through eight selector switches for every stage. Packed comparison modes
  compute their shared comparison once. Stage order, register write timing,
  swaps, comparison widths, clamping and floating-point formulas are preserved.
- Instruction fetch uses one checked big-endian word load. Unexpected addresses
  retain the reference diagnostics/bounds path.
- BAT translation has a small ordinary path separate from formatted failure
  reporting. All four BAT entries, current privilege, MSR and locked-cache state
  remain checked on every access; there is no stale translation cache.

### PS2

- Repeated draw configurations reuse the original diagnostic census entries.
  Previously the emulator formatted and hashed long texture/target descriptions
  for each primitive. Two bounded cache entries, one per GS context, now retain
  the resolved counters. Counts still update immediately and are serialized as
  before. Map replacement and state restoration invalidate the host cache.
- PSMT8, PSMT4, PSMCT32 and PSMZ32 use precomputed within-page address layouts.
  Texture and CLUT contents are always read live, preserving render-to-texture
  aliases and read-after-write behavior. The tables occupy 52 KiB total.
- Common texture sampling paths are separated from diagnostic formatting and
  use checked word loads. Probes, unsupported formats, exceptional bounds and
  the original one-shot black-texel diagnostic retain their reference paths.

### Profiling

The WASM profiler reads `emscripten_get_now()` directly, avoiding the WASI
64-bit-nanosecond/JavaScript-BigInt conversion in the previous `chrono` path.
Native builds retain `steady_clock`. Profiling remains enabled and exclusive
parent/child accounting is unchanged. Only the PS2 and GameCube artifacts were
rebuilt; the other ten published cores have identical hashes.

PS2 now times complete GS draws, including target/sampler setup and census work.
Previously only the pixel loop was inside the GS scope, so graphics setup was
charged to the enclosing VU/GIF call. In particular, the previously reported
40% VU bucket was not 40% VU instruction interpretation. After this pass the
3D checkpoint spends approximately 53% in the GS, 30% in EE/IOP/scheduling and
13% in vector-unit work. Old and new bucket percentages should not be compared
without accounting for this scope correction.

## SIMD experiment

Both builds still use `-O3 -msimd128`, with compiler-generated SIMD in the
existing code. A separate experiment processed four GameCube coverage/edge
calculations together, preserving arithmetic order and pixel write order.
Output matched, but alternating A/B runs measured 13.39 s versus 13.22 s for the
scalar coverage loop, about 1.3% slower. That experiment was removed. The final
renderer keeps the original coverage walk; it benefits from the TEV and CPU
changes above. SIMD instruction presence alone does not establish a speedup.

## Validation and limits

The public regression suite passes, including the twelve platform checks,
native pixel provenance, render replay and state round trips. New differential
tests also run as standalone WASM executables:

- GameCube: unaligned/boundary instruction fetches, 4,000 varied BAT/MSR/cache
  translations and faults, and 12,000 randomized complete TEV programs with
  swaps, texture sampling, arithmetic/compare modes and alpha tests.
- PS2: repeated/mutating/context-switched census configurations, census map
  replacement/rehashing, exhaustive small coordinate/width swizzle comparisons
  plus large wrapping addresses, and all nine supported sampling formats with
  changing texture/CLUT contents and wrapping modes.

The translated reference functions remain available for these differential
tests. Their handwritten replacements live in each core's `fast.h`; regeneration
preserves them. Host caches/derived TEV plans are not serialized, so the raw
state layouts remain version 2. The existing browser container still checks
the exact core hash and does not automatically accept another build's saves.

All benchmark output proofs match between old and new WASM. Private-media
capture/state checks verify deterministic restore, truncated-state rollback,
capture-on/off equivalence, changing intermediate replay, final-frame equality
and pixel contributors. Native and WASM game checkpoint proofs also match.
Each core completed a 600-field uncaptured run; the WASM heap remained at
193,331,200 bytes (184.4 MiB) at every 60-field sample. The complete results are
in [`ps2-gc-optimization-validation.json`](../results/ps2-gc-optimization-validation.json).

Interactive browser acceptance was not completed in
this pass: the in-app browser file chooser timed out without selecting the
disc, and the native picker fallback was unavailable. The UI/protocol is unchanged.

Existing rendering and compatibility limitations remain. These are measured
improvements for the named scenes, not a guarantee for all games or a claim of
realtime performance. Detailed captures still require substantial memory.

## Reproduction

```sh
go run ./tools/browser/console-portgen gc
go run ./tools/browser/console-portgen ps2
python3 tools/browser/console-portgen/state.py
python3 tools/platform/gc/browser/build.py --emcc /path/to/em++
python3 tools/platform/ps2/browser/build.py --emcc /path/to/em++
python3 tools/browser/check.py

# Keep both builds in separate directories containing core.js and core.wasm.
node tools/browser/tests/console-compare.mjs gc /path/game.iso /path/raw.state /path/before /path/after 60 3
CORE_DIR=/path/after node tools/browser/tests/console-check.mjs gc /path/game.iso /path/raw.state
CORE_DIR=/path/after node tools/browser/tests/console-soak.mjs gc /path/game.iso /path/raw.state 600 256
```

Use `ps2` for the corresponding PS2 checks. Private discs, firmware, states and
screenshots are not included in the repository or deployment.
