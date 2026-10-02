# GPU coverage, interpolation and texture sampling

For the subsequent phone profile, submission changes and on-demand GPU timings,
see [GPU submission costs](3DS-ACCELERATION-SUBMISSION.md).

Experimental now uploads triangles and decoded textures instead of constructing
per-pixel fragment lists in WASM. For the existing unlit draw subset, WebGPU
performs inclusive edge coverage, perspective-correct vertex-color/UV
interpolation, texture wrapping/sampling, all six TEV stages, alpha testing,
blend/logic operations and color masks. Reference remains the default, and the
existing Play selector enables this path. No Go emulator code is involved.

The Galaxy S26 Ultra report motivating this change attributed 64.7% of its
measured update window to CPU coverage/sampling/input preparation and 26.2% to
WebGPU upload/wait/readback. This implementation removes the per-pixel CPU work
from supported draws. It retains eager readback for coherent guest RAM and
inspection. Lighting, depth, general stencil, projection/shadow textures and
unsupported draws still use the existing fallbacks. Texture **decoding**, vertex
programs, clipping and ARM execution remain on the CPU.

## Matching the browser reference

Ordinary WGSL floats do not guarantee the browser reference's results: the
[WGSL rules](https://www.w3.org/TR/WGSL/#floating-point-evaluation) permit
reassociation/fusion and bounded division error. `raster-3ds.js` implements
binary32 addition, multiplication and division with integer significands,
explicit round-to-nearest/ties-to-even, gradual underflow and signed zero.
Every reference intermediate is rounded separately. Division starts with a
normal-range floating estimate, corrects the quotient using integer products
and remainders, then rounds the exact result. The estimate never becomes a guest
result directly. Floor and color conversion also use integer float bits.

A 16×16 coarse bin holds ordered triangle references. Each invocation owns one
framebuffer pixel, tests its bin's triangles in reference order and uses the
existing integer fragment tail. There are no unordered framebuffer atomics;
only the drawn-fragment counter is atomic. One invocation cannot race another
pixel's blend or color-mask update. The tiled framebuffer and flipped texture
coordinates follow the browser Reference.

The C++ producer checks the same eligibility conditions as the prior hybrid
path, plus bounded screen coordinates and triangle area. Larger numeric inputs
retain the hybrid fallback. Decoded textures are copied before submission and
sampled from that immutable snapshot, preserving the reference draw's texture
cache behavior even when texture storage aliases a render target. No guest
execution can run while the asynchronous operation is pending.

The host validates bin membership, primitive order, bounds, workload, float
inputs and texture ranges. A zero/invalid interpolated reciprocal or unexpected
UV magnitude marks the entire GPU result unusable. The bridge refuses that
result without copying any bytes; Reference executes the original operation.
Successful results update canonical RAM and drawn counters before WASM resumes.
The renderer source is included in portable-state core identity and the static
release manifest.

## Packet kind 6

The existing immutable graphics stream remains version 1; kind 6 adds a new
operation without changing kinds 1–5. All words are little-endian u32.

- Parameters 0–34 retain kind 5's framebuffer/TEV/blend fields. Field 9 is the
  sum of triangle bounding-box areas, an upper bound on drawn fragments; 34 is
  the recorded expected count (UINT32_MAX for live execution).
- Parameters 35–38 are triangle count, bin columns, bin rows and the three
  texture-enable bits. Parameters 39–50 are three descriptors containing input
  word offset, width, height and wrap register. Parameter 51 is zero overlap.
- Input starts with two words per bin (list word offset, list length), then 44
  words per triangle: four bounding-box coordinates, area float bits, and three
  vertices of 13 words (x, y, reciprocal W, RGBA, three UV pairs).
- Ordered bin lists follow the triangles. Decoded RGBA texture words follow the
  lists. Empty enabled textures retain Reference's white sample; disabled units
  retain zero. Buffers remain capped at 16 MiB, with at most 1,024 triangles and
  the prior hybrid bounding-box work budget.

The shared GPU allocation now scales with geometry/textures and framebuffer
size, instead of the much larger per-fragment lists. Textures are still uploaded
per draw; this change does not introduce a persistent GPU texture cache.

## Validation and measurement

The arithmetic test compares 396,288 GPU results against independently rounded
JavaScript binary32 arithmetic, including random operands, signed zeros,
subnormals, overflow, cancellation and adjacent/halfway cases: zero mismatches.
The public browser-C++ fixture produces 309 packets: 307 exact results and two
intentional overlapping-copy fallbacks. This includes 96 retained hybrid draws,
160 GPU raster draws, 14 texture formats, three texture units, eight wrap modes,
empty/partial bins, shared and negative edges, extreme reciprocal W, subnormal
attributes, constant-color rounding, framebuffer/texture aliases, TEV, alpha,
blend, logic and channel masks. Drawn counters match every packet. A deliberately
invalid interpolation packet is refused without committing output.
The private welcome replay matches every byte of all 25 recorded packets; input
recording also preserves full serialized continuation.

Three warmed 30-interval trials per scene compare Reference with Experimental,
including presentation of both screens and alternating execution order. A fresh
run of the prior release `a02250b22ef758e75f86` uses the same harness/checkpoints.
Reported values below are medians; old/new saved state hashes also match.

| Browser checkpoint | Reference | Previous Experimental | GPU raster | Improvement over previous |
| --- | ---: | ---: | ---: | ---: |
| Welcome | 264.1 ms | 145.3 ms | 48.7 ms | 2.99× |
| Title/Mii dialog | 303.5 ms | 172.7 ms | 57.3 ms | 3.02× |

Times are per display interval, not unique game frames. Relative to Reference,
the new backend measures 5.3–5.6× across these trials. CPU preparation drops from
82.5/95.6 ms to 0.82/0.89 ms per interval. Total input uploads drop from
28.44/33.77 MiB to 1.58/2.03 MiB per interval. GPU scratch plus presentation drops
from 17,367,108 bytes to 2,687,044 bytes. This is **GPU allocation**, not total
browser or WASM memory. The GPU upload/wait/readback bucket is now the dominant
cost; this change does not remove per-operation synchronization.

All live trials match full saved machine state and final scanout exactly. The
six-interval delayed-GPU test also matches. Device loss during readback and an
injected timeout recover to Reference with exact state/pixels. The packaged UI
passes mode selection, pause/capture/replay/pixel inspection, Resume, Memory,
portable save/load with Play preference, cancellation, stepping, reset, and
loading an Experimental preference when WebGPU is absent. The sustained
welcome check completes 520 intervals in each engine sequence, 270 actual mode
switches and 5 complete Reference captures. Periodic saved states, pixel histories,
replay output and Memory inspection match the uninterrupted Reference sequence.
GPU draws continue throughout the run. The complete public browser CI suite and
release audit pass for `9909cce5d24c89960501`.

These are Chrome 154 / Apple Metal 3 measurements on the development Mac. They
are menu/welcome checkpoints, not verified interactive gameplay, and remain
below full speed. No Galaxy S26 Ultra or other GPU/browser performance is claimed.
The new arithmetic is designed for portable WGSL behavior, but this hardware
validation is limited to the available adapter. No private media, checkpoints,
texture data or screenshots are committed.

Evidence: [arithmetic](../results/3ds-acceleration-raster-float.json),
[public replay](../results/3ds-acceleration-raster-public.json),
[private replay](../results/3ds-acceleration-raster-replay-welcome.json),
[new welcome](../results/3ds-acceleration-raster-live-welcome.json),
[new title](../results/3ds-acceleration-raster-live-title.json),
[previous welcome](../results/3ds-acceleration-raster-baseline-welcome.json),
[previous title](../results/3ds-acceleration-raster-baseline-title.json),
[recovery](../results/3ds-acceleration-raster-recovery.json), and
[UI](../results/3ds-acceleration-raster-ui.json), and
[sustained handoff](../results/3ds-acceleration-raster-handoff-welcome.json).

## Reproduce

Build the browser core with the usual `build.py --emcc /path/to/em++` command.
Build `tests/graphics-3ds.cpp` with C++20, `-fwrapv -ffp-contract=off`, then run it
with a local packet-output filename. With the repository served locally and
Playwright/Chrome installed:

```sh
node tools/browser/tests/raster-float-3ds-browser.mjs /tmp/float-report.json
node tools/browser/tests/graphics-3ds-browser.mjs http://127.0.0.1:8790 /packet-URL /tmp/replay-report.json
node tools/browser/tests/graphics-live-3ds-browser.mjs /path/game.cci /path/browser.state /tmp/live-report.json 30 3 0
```

`PLAYWRIGHT_MODULE` can identify an installed Playwright entry point. The live
harness accepts `CORE_BASE` and `GRAPHICS_BASE` URL prefixes and `RASTER_KIND=5`
for reproducing the previous hybrid release. By default it requires actual
kind-6 GPU execution and rejects unexpected packet fallbacks. Public CI remains
`python3 tools/browser/check.py`; browser GPU tests require a real adapter.
