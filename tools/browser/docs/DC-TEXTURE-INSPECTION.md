# Dreamcast texture inspection

Render uses the shared output, timeline and pixel inspector. The Dreamcast
adapter adds a texture preview and GPU state that follow the accepted seek.
Previous/Next texture skip directly to textured draws.
It does not change Play, CPU/GPU execution or the portable core-state format.

## Views

- Decoded texture: the bound image, its current format and palette, and an
  optional UV outline for the current strip triangle or sprite. Coordinates
  outside the texture are clipped in the guide, not wrapped into invented
  triangles. The sampler itself wraps. The guide is not exact pixel coverage.
- Memory order: consecutive stored texels laid out in rows. Twiddled images
  demonstrate the address rearrangement. For VQ, grayscale bytes select
  dictionary entries, one byte per 2×2 block.
- Active palette: 16/256 colors from the selected command's PVR register
  snapshot. VQ instead shows 256 dictionary entries, each a 2×2 texel block.
  A texel click highlights its palette entry or compression dictionary block.
  The lookup panel can be hidden to enlarge the texture.
- Source details: byte address, nibble, index, palette register or VQ index
  byte/dictionary entry, stored color and decoded RGBA. Addresses refer to the
  core's 64-bit VRAM layout; framebuffer registers use the CPU 32-bit aperture.
- GPU state: dimensions, binding, format, storage, mipmapping, shading/list,
  depth compare/write, framebuffer address and requested sampler controls.

## Capture and fidelity

`texture-inspection.h` projects each pending TA header/vertex because the
existing callback precedes execution. It maintains strip UVs and handles
sprite UV16 with the inferred fourth corner. A clear resets the binding;
initial/CPU-only steps have no active binding. Each command retains the id of
its existing deduplicated PVR register snapshot, including palette and format.
There is no read from live palette RAM during a seek.

The three `rr_vram_*` exports expose scratch replay VRAM and selected metadata.
The worker copies the 8 MiB snapshot only for captures and completed seeks;
the existing request/capture guards reject stale and cancelled responses.
Decoded image buffers are lazy and cached until the next snapshot. State
metadata is bounded by the recorder's event cap; its small per-event records
are additional to the existing capture byte estimate.

`dc-texture-decode.js` matches the current core sampler for ARGB1555, RGB565,
ARGB4444, PAL4/PAL8, square/rectangular Morton layout, row order, VQ and the
largest level of mipmapped textures. Palette colors also support ARGB8888.
Missing/out-of-range source data is transparent and identified in inspection;
unsupported formats are magenta, matching the core.

Existing limitations are stated in the panel: point sampling, wrapping rather
than requested clamp/flip, ignored requested strides, no mip LOD selection,
unsupported YUV/bump formats, and direct polygon rendering rather than hardware
tile/deferred processing and translucent auto-sort. Preview memory is from
**after** the selected step; reads within self-modifying primitives can differ.
This adds texel-address inspection, not captured per-fragment texture-read
provenance. The existing pixel inspector still traces framebuffer writes.

Hardware field reference: [KallistiOS PVR headers](https://kos-docs.dreamcast.wiki/pvr__header_8h_source.html).

## Checks

`texture-dc.cpp` tests historical bindings/palettes/VRAM, header-only steps,
UV16 strips and end markers, sprites, capture reset, and paused-state isolation.
It can emit 34 synthetic sampler fixtures. `dc-texture.mjs` tests format/alpha,
Morton addressing, palettes, compression and bounds, and compares 2,023 texels
against those actual native sampler results. Both are included in `check.py`.

`ui-workspace.html` exercises accepted/stale/cancelled seek updates, texture
source clicks, format changes, memory-order view and lookup-panel controls.
The private Crazy Taxi driving capture traversed 18,570 steps, found 131 texture
bindings in four formats (RGB565, ARGB4444, ARGB1555, PAL8), and checked 8,384
texel lookups. Full serialized machine state remained identical after all seeks;
execution with capture matched execution without it. No game data is committed.
