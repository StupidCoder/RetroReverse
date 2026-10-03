# Unlit depth rasterization

The browser Experimental backend adds exact unlit depth testing and writes.
Release bundle: `b603d833217206c8a8ef`; previous bundle retained:
`680f5367cab4ac6c7c94`. Final source and packaged-release checks pass.

## Why this subset

The next Galaxy S26 Ultra screenshot was taken after the demo level became
visible on the top screen. Its exclusive window totals 1265.24 ms: 462.50 ms
software rasterization, 441.00 ms GPU round trips, 190.40 ms PICA commands/vertices,
164.30 ms ARM/Horizon, 4.43 ms DSP, 2.53 ms GPU preparation and 0.08 ms GX.
These are update-window totals, not frame times.

For 50 GPU operations, timestamps report 289.34 ms compute inside 412.88 ms
completion wait. Compute is about 22.9% of the total exclusive window and 70.1%
of completion wait. The 9.97 ms CPU readback copy is under 1% of the window.
The remaining 123.54 ms of completion wait also includes transfers, scheduling
and browser/driver overhead; it is not an isolated synchronization measurement.

A native build of the **browser C++ core**, instrumented inside `n3ds_GPU_fill`
after packet preparation, sampled 120 display intervals of the animated title
with the level visible behind the logo. This is a comparable scene, not proof
of the phone's exact game state. Existing eligible kind-6 draws were separated
from software fallbacks. The latter consumed 8986.38 ms, of which 2649.05 ms
(29.5%) belonged to two large unlit, depth-tested passes per interval. Other
fallback groups include lit meshes, stencil/shadow work and small draws. Native
CPU times select a useful extension; they do not predict WASM/GPU performance.
The [census](../results/3ds-acceleration-depth-census.json) contains feature flags,
counts and checkpoint hashes, with no media, pixel data or saved state.

## Contract

Kind 7 extends kind 6's ordered, per-pixel compute rasterizer with unlit depth
testing and optional depth writes. It implements all eight comparison functions,
Z-buffer and W-buffer depth, scale/offset, clamping, and the Reference sequence of
binary32 rounding followed by truncation to a 24-bit integer. Alpha rejection
leaves depth untouched; successful writes preserve the existing high stencil
byte. Depth rejects update the canonical `DepthKilled` counter.

Color and depth are uploaded as one paired snapshot and returned by one mapped
readback. The WASM bridge validates the entire result and both counters before
copying either surface. Guest execution stays suspended until both copies and
counter updates finish. A rejected result falls back to the original draw once.
Overlapping color/depth ranges, including virtual aliases, remain in Reference.
Lighting, stencil-enabled draws, shadow sampling and existing bounded-input
fallbacks remain in Reference. A depth-write flag with depth testing disabled
has no effect in Reference and can use the existing color-only path.

Enabled textures must already be decoded in the Reference texture cache. A cold
texture keeps the draw in Reference, which warms it only if a fragment reaches
sampling. This avoids speculative cache population for fully depth-rejected
draws, including depth/texture aliases observed by later draws. Subsequent
cache hits use the GPU. The matched 30-interval scene accelerates 58 depth draws;
the two initial cold draws stay in Reference.

Packet layout, still under graphics-stream schema 1:

- Parameters 0–50 retain kind 6's meanings. Parameter 51 has enabled bit 0,
  depth-write bit 1, Z-buffer bit 2 and comparison bits 4–6. Parameters 52–53 are
  binary32 depth scale/offset; 54 is reserved zero, 55 is expected `DepthKilled`
  or `UINT32_MAX` for live execution, and 56 is the overlap flag.
- Each triangle has 47 words: the existing five-word header and three 14-word
  vertices. Each vertex appends screen-space Z after its existing 13 attributes.
  Bin entries still hold absolute input-word offsets.
- `before` and `expected` concatenate color then depth, each `width*height*4`
  bytes. The two surfaces need not be adjacent in guest memory.
- Readback counters are two words: drawn, then depth rejected. The existing
  high-bit arithmetic failure signal remains in the drawn word. Their sum must
  not exceed the packet's bounding-box workload.

## Performance

Chrome 154 / Apple Metal 3, three warmed and alternating Reference/Experimental
trials per build, 30 display intervals per trial including presentation, with
idle input so the animated title remains visible. Timestamp queries are off.
The previous release is rerun with the same checkpoint and harness.

| Build | Experimental median per display interval |
| --- | ---: |
| Previous release | 254.94 ms |
| Unlit depth GPU path | 219.97 ms |

This is a **13.7% reduction in elapsed time (1.16× speedup)** on the Mac.
Reference medians are 362.53 ms and 364.50 ms respectively. Every trial and the
additional delayed-GPU run produce the same full saved state and exact final
scanout. This is not a phone measurement or a full-speed claim. Lighting,
stencil/shadow passes and PICA vertex processing remain substantial costs.

Reports: [previous build](../results/3ds-acceleration-depth-baseline-demo.json) and
[updated build with delayed continuation](../results/3ds-acceleration-depth-live-demo.json).

## Validation

The C++ fixture generates 405 packets, including 96 new depth cases. Those cover
all comparison functions, depth write on/off, Z/W buffering, alpha rejection,
ordered overdraw, negative/out-of-range/subnormal depth, signed zero, preserved
stencil bytes, and texture/depth aliasing. Each depth fixture also verifies that
a cold-cache preparation falls back without inserting a texture. Color/depth
overlap falls back before packet construction. Hardware replay matches all 403
supported packets, including both counters and both surfaces; two overlapping
copy cases retain their expected fallback. Invalid interpolation causes whole
operation refusal for both color-only and depth packets.
[Public GPU report](../results/3ds-acceleration-depth-public.json).

The [WASM bridge report](../results/3ds-acceleration-depth-bridge.json) compares
full saved state after two display intervals, including accelerated depth draws.
It uses recorded Reference output, **not GPU execution**. Valid delayed responses
and nine failure modes all reproduce the same Reference state: truncated paired
output; missing, negative, excessive, fractional or nonfinite depth counters;
a combined counter overflow; explicit rejection; and an exception. Failed
responses contain deliberately corrupted bytes to detect partial commits.

Hardware device loss and timeout injected into a kind-7 operation recover to
Reference with unchanged state and pixels. The UI verifies a complete animated
scene capture and pixel evidence, resume to Experimental, memory handoff,
save/load, cancellation, reset, and absent-WebGPU recovery.
[Recovery report](../results/3ds-acceleration-depth-recovery.json) and
[UI report](../results/3ds-acceleration-depth-ui.json).

Both sustained handoff tests pass 520 display intervals per sequence, with 270
backend switches and exact full-state and display-pixel matches at all eleven
checkpoints. The demo sequence executes 50 paired depth operations on the GPU;
the welcome sequence additionally verifies five complete captures, replay pixels
and capture evidence against Reference. Memory inspection and state round trips
pass in both sequences. The complete `tools/browser/check.py` run passes,
including the release's 109 pinned assets and sixteen direct routes.
[Demo handoff report](../results/3ds-acceleration-depth-handoff-demo.json) and
[welcome handoff report](../results/3ds-acceleration-depth-handoff-welcome.json).

## Existing dense-capture limit

At a denser point in this scene, a three-interval Reference trace reaches the
8,388,608-write cap and reports 2,420,022 dropped entries. The previous release
and the initial depth prototype produce identical counts. A further state-save
stress after retaining and inspecting that large trace also returns failure in
both builds. Its precise allocation cause has not been isolated. These limits
are not solved by moving depth draws to the GPU: inspection uses Reference.
The passing UI capture covers a different, complete window in the same scene.
[Matched capture-limit evidence](../results/3ds-acceleration-depth-capture-limit.json).

The strict welcome endurance test continues to require complete captures. The
long demo endurance test uses `CAPTURE_MODE=none` and verifies repeated engine
switching, full state, display pixels and memory/state round trips; it does not
claim complete dense-scene capture histories. No production capture budget or
capture duration is changed, and the default test still rejects incomplete
captures.

## Reproduction

Use an owned cartridge and a browser checkpoint showing the animated title.
Keep those inputs and generated binary streams local. Build with the existing
3DS browser build command, then run:

```sh
CORE_DIR=tools/platform/n3ds/browser/web node tools/browser/tests/graphics-depth-bridge-3ds.mjs "$MEDIA" "$STATE" /tmp/depth-bridge.json 2
clang++ -std=c++20 -O1 -fwrapv -ffp-contract=off -Wno-parentheses-equality tools/browser/tests/graphics-3ds.cpp -o /tmp/graphics-3ds-depth
/tmp/graphics-3ds-depth tools/platform/n3ds/browser/work/graphics-public-depth.bin
node tools/browser/tests/graphics-3ds-browser.mjs http://127.0.0.1:8790 /tools/platform/n3ds/browser/work/graphics-public-depth.bin /tmp/depth-public-gpu.json
IDLE_INPUT=1 RASTER_KIND=7 node tools/browser/tests/graphics-live-3ds-browser.mjs "$MEDIA" "$STATE" /tmp/depth-live.json 30 3 3
IDLE_INPUT=1 FAULT_KIND=7 node tools/browser/tests/3ds-recovery-browser.mjs "$MEDIA" "$STATE" /tmp/depth-recovery.json
CAPTURE_MODE=none RASTER_KIND=7 node tools/browser/tests/3ds-handoff-browser.mjs "$MEDIA" "$STATE" /tmp/depth-handoff-demo.json 520
node tools/browser/tests/3ds-handoff-browser.mjs "$MEDIA" "$WELCOME_STATE" /tmp/depth-handoff-welcome.json 520
```

Set `PLAYWRIGHT_MODULE` if Playwright is outside the default module search path.
`IDLE_INPUT=1` prevents benchmark/recovery input from dismissing the title scene;
`FAULT_KIND=7` injects recovery failures into a paired depth operation. Rerun the
same idle benchmark against the previous release for a matched baseline. Hardware
public replay must match both color/depth bytes and both counters. Packaging uses
`tools/browser/package.py --core 3ds`; only the 3DS core is refreshed. The other
fifteen packaged cores remain byte-identical. Run the
packaged UI acceptance and complete browser checks before publishing.
