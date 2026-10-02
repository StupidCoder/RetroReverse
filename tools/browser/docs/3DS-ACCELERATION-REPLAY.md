# M2: isolated graphics replay

The prototype selects **integer compute with reference fallbacks**. Conventional
rasterization disagrees with the browser reference's inclusive edge coverage on
41 pixels in three public triangles. The compute coverage probe agrees, but that
does not establish equivalent interpolation, sampling, lighting or depth rounding.
WGSL's [floating-point evaluation rules](https://www.w3.org/TR/WGSL/#floating-point-evaluation)
permit transformations that require care when matching the reference's ordered
float32 operations. General PICA draws therefore remain reference operations.

The implemented replay subset is:

| Operation | Support |
| --- | --- |
| 16/24/32-bit memory fill | Exact, including partial trailing units |
| Texture copy | Exact contiguous/gapped rows; overlapping physical backing rejected |
| Display transfer | Existing reference input format, five output formats, flip and clipped dimensions |
| Stencil-rejected triangles | Half-pixel coordinates in a bounded exact integer range; no sampling, lighting or alpha test; all eight stencil operations and write masks |
| Textured/lit draws and other depth/blend combinations | Reference fallback; not claimed as GPU-equivalent |

This is a deliberate narrowing at M2's decision gate. M3 integrates these proven
operations. General draw input capture, textured replay and wider PICA coverage
remain work for M4; they are not implied by the transfer results. A private scene
is **not** a complete GPU replay: all its transfers are compared, while its
remaining draws are explicitly counted as reference work.

## Input ownership and bounds

`core/graphics-stream.h` records inputs before each transfer and expected output
after reference execution. The stream is independent of Render's write evidence.
Each version-1 record owns its source, initial destination, parameters and expected
destination. A later guest write cannot alter an earlier packet. Aliases are
checked using physical backing pointers, not virtual address equality. Recording
is capped at 256 MiB and reports omissions instead of silently truncating.

Packets use little-endian `u32`: file magic `0x50475252`, version `1`; then records
with byte size, operation, parameter count, input/before/expected byte lengths,
source and destination guest addresses. Parameters precede the three byte arrays,
each padded to four bytes. The final parameter is the physical-overlap flag.
Stencil packets contain framebuffer dimensions, stencil registers, triangle count,
then each bounding box and six doubled vertex coordinates. No private stream is
committed. The JS decoder rejects malformed lengths and unknown stream versions.

The compute backend owns one operation at a time and reuses bounded buffers.
Preparation compiles asynchronously; validation errors and device loss reject the
result. Output is an isolated byte array. M2 never installs it into a live machine.
Shader source is maintained in `site/emulators/graphics-3ds.js`, and the browser
port generator retains the recording hooks without changing the Go emulator.

## Evidence

`graphics-3ds.cpp` generates public fixtures using the browser C++ implementation.
`graphics-3ds-browser.mjs` compares every destination byte on Chrome's Apple Metal
adapter. All 51 supported public fixtures match; both overlapping-copy fixtures
are rejected. Fixtures include source overwrite, physical aliases, unaligned
destinations, row gaps, trailing fill bytes and repeated stencil coverage.

Private welcome and title/Mii-dialog runs cover three display intervals each:
13 and 18 transfer packets respectively, all exact. The runs contain 178 and 280
reference draws; those draws are not hidden in the GPU coverage count. The input
recorder also compares complete serialized continuation with recording disabled.
Reports are `../results/3ds-acceleration-m2-*.json`.

Timing reports separate host upload, command submission, GPU timestamp duration
(when available), queue/map wait and readback copy. Queue/map time includes host
scheduling and GPU completion and must not be interpreted as shader time alone.
Presentation is not part of this offscreen harness. M3 measures the live bridge;
these isolated numbers are not a claimed emulator speedup.

## Reproduce

Build `tools/browser/tests/graphics-3ds.cpp` with the normal C++20 reference flags
and pass an output `.bin` path. Serve the repository locally and run
`graphics-3ds-browser.mjs <origin> <stream-URL> <report.json>` with Playwright.
`record-3ds-graphics.mjs <media> <browser-state> <output.bin> [intervals]` creates
private transfer packets using the browser WASM core. Set `CORE_DIR` to a newly
built core when testing before packaging. Private media and states stay local.
