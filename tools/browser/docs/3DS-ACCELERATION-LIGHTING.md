# D3a: directional fragment lighting on the GPU

Experimental mode now runs the demo's normal-mapped directional lighting with
its existing GPU coverage, texture sampling, depth and texture-combiner stages.
Reference keeps the original renderer and observation hooks. This is the first
D3 feature family; general stencil, shadow textures, additional light sources and
attenuation remain open. The Go emulator is unchanged.

The implementation lives in [graphics-lighting.h](../../platform/n3ds/browser/core/graphics-lighting.h),
[lighting-3ds.js](../../../site/emulators/lighting-3ds.js), and the existing raster,
fragment and transport modules. The worker's renderer identity includes the new
lighting module, so saved execution identities distinguish its semantics.

## Eligibility and exact arithmetic

Kind 8 supports one directional light with depth testing, optional depth writes,
normal maps with optional Z reconstruction, environments 0–7, two-sided diffuse,
highlight clamping, specular distributions, reflection tables and Fresnel alpha.
Each active shared LUT has the same absolute/signed indexing, stored slope,
clamping and scale as Reference. Tangent maps have no observable lighting effect
in these environments; environment 8, which uses the tangent projection, stays
in Reference. Positional/multiple lights, geometric factors, distance/spotlight
attenuation, lighting shadow attenuation, general stencil and shadow sampling
also stay in Reference. Existing observer and bounded-work guards still apply.

All operations use the existing integer binary32 helpers. The new square-root
helper normalizes the radicand, obtains a normal-range hardware estimate, corrects
it using exact 64-bit integer squares, and rounds against the exact midpoint.
No approximate result reaches emulated memory. Quaternion rotation and the sums
and products in lighting preserve Reference's operation order. Normalizing a
constant directional light is hoisted to packet preparation with the Reference
helper; changing its direction still changes the next packet.

Quaternion/view components may have magnitude at most 2^20 at vertices and 2^24
after interpolation. These keep normalization arithmetic finite. Nonfinite or
out-of-range runtime interpolants reject the entire draw before canonical writes.
The initial 1,024 limit excluded the landscape's view coordinates and was too
restrictive; it was replaced after a rejection census. Color/position/spot inputs
are finite and bounded, active LUT values/slopes are finite within [-1,1], and
unused tables are zeroed. Signed zero is retained during LUT clamping.

Enabled textures must already exist in the Reference decoded cache. Cold textures
retain the original lazy warming behavior, including draws rejected by depth and
texture/depth aliases. Color/depth aliasing is refused. Immutable packets and the
existing bridge commit both surfaces and counters only after a complete valid
GPU result; an unsupported or failed result falls back once. The guest remains
suspended through readback. Shader caches never enter canonical saved state.

## Pipeline specialization

The generic exact kernel is immediately available. A per-device cache compiles
up to 64 specializations, with at most two pending, asynchronously. Keys contain
all specialized lighting flags, enabled tables, input selectors, absolute/signed
selection, scales and two-sided mode. Colors, direction, texture contents, LUT
values, offsets and framebuffer state remain runtime inputs. Pending, refused or
failed compilation uses the generic kernel. Disposal prevents pending compilation
from installing a result. No additional guest effects are speculated.

## Packet contract

Graphics-stream schema 1, kind 8:

- Parameters 0–55 retain kind 7, including expected drawn/depth-killed counters.
  Parameters 56–62 contain lighting flags, active-table mask, LUT input/absolute/
  scale words, input LUT offset and two-sided mode. Parameters 63–83 contain
  seven RGB vectors: global ambient, specular 0/1, diffuse, light ambient,
  normalized direction and spot direction. Parameter 84 is the overlap flag.
- Each triangle has 68 words: five header words and three 21-word vertices.
  Each vertex appends quaternion XYZW and view XYZ after kind 7's Z coordinate.
  Ordered coarse bins still contain absolute word offsets.
- Decoded textures are followed by seven shared tables, each holding 256 pairs
  of binary32 value and difference. This adds 14 KiB per lighting packet.
- Color/depth snapshots, paired outputs, early depth rejection, stencil-byte
  preservation, counter validation and transactional fallback match kind 7.

The JavaScript validator checks complete lengths, ranges, ordered bin membership,
active/inactive LUT data and numeric bounds before dispatching an offline replay.

## Validation

The public native fixture emits 597 packets: the previous 405 plus 192 lighting
cases. Lighting cases cover all admitted environments, map modes, selectors,
scales, signed/absolute LUT inputs, negative slopes, zero and tiny normals,
large coordinates, signed zero, alpha/depth rejection and changed runtime colors,
direction and LUTs under the same specialization key. Unsupported combinations
are explicitly checked for fallback.

Hardware replay matches all 595 supported packets byte for byte; the two existing
ordered-overlap cases remain refused. A second replay of the 192 lighting packets
also matches, including 128 uses of ready specialized kernels. All 64 compiled
specializations finish without errors. Draw and depth-killed counters match,
and degenerate interpolation refuses kind 6, 7 and 8 packets before committing.

The arithmetic fixture compares 599,296 operations (add, multiply, divide and
square root) over 149,824 input pairs with zero mismatches. A separate test checks
all 16,777,216 binary32 inputs in [1,4), covering every normalized significand and
exponent parity for square root, with zero mismatches. The ordinary fixture also
covers signs, exponent boundaries and subnormal scaling.

Public reports:
[graphics](../results/3ds-acceleration-d3-graphics.json),
[arithmetic](../results/3ds-acceleration-d3-float.json),
[square root](../results/3ds-acceleration-d3-sqrt.json).

## Running-demo performance

Chrome 154.0.8037.98 on the same Apple M3 Pro / Metal 3 MacBook, AC power and
low-power mode disabled. The private checkpoint starts at display interval 3240
with Mario running along the landscape. Five alternating 300-interval trials use
the production worker, both screens and idle input, with timestamps off. Only
one emulator executes at a time. Compilation and other emulator tests do not run
alongside the timed trials. The baseline is the shipped D2 WASM build.

| Metric | D2 baseline | D3a lighting |
| --- | ---: | ---: |
| Mean elapsed per interval, five trials | 231.61 ms | 194.11 ms |
| Individual trial means | 230.98–231.96 ms | 193.03–194.94 ms |
| Nominal display rate | 7.20% | 8.59% |
| Mean trial p95 | 479.20 ms | 415.80 ms |
| Worst 60-interval window, maximum across trials | 242.40 ms | 210.33 ms |
| Longest execution slice, maximum across trials | 202.70 ms | 195.10 ms |
| PICA commands / vertices | 12.07 ms | 11.72 ms |
| Remaining software rasterization | 131.30 ms | 64.91 ms |
| GPU upload / wait / readback | 30.18 ms | 58.68 ms |
| GPU input preparation | 0.45 ms | 1.10 ms |
| ARM / Horizon | 49.45 ms | 49.54 ms |

Whole-scene elapsed time decreases **16.2%**, or **1.19× throughput**. Moving
66.39 ms of software work adds 28.49 ms to GPU round trips; the net result is
37.49 ms saved per interval. The 25% target remains unmet. This is a Mac result;
no Galaxy S26 Ultra improvement is claimed.

Each 300-interval candidate trial accelerates 4,450 lighting draws, adding to the
previous 6,898 operations. All 16 scene lighting specializations become ready,
with no compilation failures. The first full trial uses 4,440 specialized lighting
draws; later trials use all 4,450. No submitted GPU operation is refused. Paired
framebuffer upload/readback increases from 2.79 to 6.20 GB per trial, in each
direction; immutable input uploads rise from 0.51 to 1.02 GB. Batching and resident
surfaces are consequently more valuable after this coverage change.

All trial, timestamp and diagnostic runs produce identical complete canonical
state (`c44259c6…701b7391`), sampled pixels and guest counters. The midpoint image
shows Mario jumping along the blue brick wall; the sequence continues to the red
platform. Maximum WASM heap capacity is 1,385,889,792 bytes for both builds; this
is reserved capacity, not a measurement of live allocations or GPU driver code.

The separate timestamp run averages 198.91 ms per interval. GPU execution is
33.80 ms within 53.51 ms of queue/map wait; it is not an additional independent
wall-time bucket. Upload is 6.51 ms, submit 0.31 ms and readback copy 1.23 ms per
interval. The diagnostic run averages 199.73 ms. Remaining fallback time is led
by lit stencil/depth draws (21.01 ms), another lit stencil pass (12.82 ms), and
two unlit stencil passes (9.84 and 9.49 ms). These are the next D3 family to cover.

The 30-interval cold window is 239.39 ms/interval for baseline and 197.91 for the
candidate. Backend creation occurs before that window; these are not total
application launch times. Full report: [five comparisons and diagnostics](../results/3ds-acceleration-d3-demo.json).

A separate single generic-only control averages 195.77 ms/interval, with identical
canonical state, pixels and counters and no requested lighting specializations.
This is consistent with a small specialization benefit, but does not establish a
precise independent speedup. The large gain above comes from GPU coverage. See
[generic control](../results/3ds-acceleration-d3-generic.json).

The [handoff run](../results/3ds-acceleration-d3-handoff.json) performs 32 actual
backend switches over 60 intervals with deterministic input. Reference and mixed
runs match checkpoints, pixels, guest counters, memory inspection and save/load.
The mixed run executes kind 8 lighting; Reference executes no compiled vertex
batches. Dense render capture is disabled in this run because of the existing
[capacity limit](3DS-ACCELERATION-DEPTH.md#existing-dense-capture-limit). This is
handoff coverage, not the outstanding D6 520-interval/270-switch/capture gate.
The browser harnesses close their browsers in `finally`, and each emulator test
finishes before the next starts.

Release bundle: `ccb90e2984e7b6ae45b2` (111 pinned assets). Previous bundle `59463ed45f317abd722b` is retained. Only the 3DS core asset hashes change; the other fifteen cores are unchanged.

The complete public browser check suite and packaged-release audit pass. Bridge
fault tests cover specialization rejection and late completion after disposal,
in addition to the existing map, validation, timeout and device-loss cases.
