# Game Boy Advance and Dreamcast browser ports

Both ports translate RetroReverse's existing Go emulators to C++20 and WASM.
They use the shared browser shell: local images, Run/Pause/Reset/Next frame,
keyboard/on-screen/standard-gamepad mappings, subsystem timing, portable states,
automatic pause capture, pixel history and rendering replay. Neither embeds a
third-party GBA or Dreamcast emulator. Emscripten/LLVM and their runtime libraries
provide compilation and browser support. The gallery now contains fourteen systems.

## Scope and media

GBA includes ARM7TDMI, BIOS HLE, PPU modes 0–5, tile/bitmap and affine
backgrounds, objects, windows, priority and color effects, DMA, timers, audio
synthesis and the original EEPROM model. Select a `.gba` cartridge between 192
bytes and 32 MiB; loading is not restricted by title/hash. Mosaic, SRAM/Flash
save chips and link cable retain the Go core's limitations. Timing retains its
instruction-budget model. This is not a claim of cycle-exact emulation.

Dreamcast includes SH-4/FPU, AICA ARM7 and synthesis, Maple controller, the
original BIOS/GD-ROM HLE and software PowerVR renderer. Select the CUE sheet
and its **single combined raw BIN together**. The parser handles the existing
cdrdao-style track sheet and standard raw Mode 1 CUE indexes, anchoring data
tracks from their sector headers. GDI, CDI, CHD, multiple separate track files
and implicit gaps are unsupported. The worker retains the local File and reads
bounded slices through the shared 256 KiB disc cache. The whole disc is not
copied into WASM or uploaded. Save/load identity verification hashes the local
files; the reference disc took about 32 seconds in the browser.

No GBA or Dreamcast firmware is distributed or requested. The inherited HLE
and compatibility limits are explicit in the UI. Audio output, external battery
saves/VMU files and disc audio are not exposed; portable browser states preserve
the session. Dreamcast translucency ordering and some PowerVR features remain
incomplete. Arbitrary images may be selected; broad commercial-game compatibility
has not been established.

## Captured evidence

GBA records actual PPU layer samples in capture-local surfaces and final pixel
composition (top/second layers and colors, window mask and blend control).
Layer samples retain video-memory/palette sources. The inspector can follow a
final pixel to its PPU sample and then the recorded source writes. BIOS RAM
clears, CPU and DMA video-memory writes are observed. Display-register snapshots
accompany layer events. These are actual modeled PPU effects, not a fabricated
wipe over the completed screen. The replay displays the final composition as it
is produced; a separate layer viewer is not provided. Second-layer inputs are
recorded as composition metadata, without a second automatic source chain.

Dreamcast records actual VRAM writes, framebuffer clears, TA parameter words
and deduplicated PowerVR register snapshots. Three fields include double-buffer
producer context. Pixel queries handle interleaved VRAM even when a packed
RGB888 pixel crosses a bank boundary. The replay uses the final scanout mapping.
Rejected fragments, sampled-texture histories and CPU command-buffer ancestry
are not captured. PCs identify observed writes or submission, as labeled.

Both replays use separate memory and do not change the paused guest. Capture
limits are explicit: 1,048,576 writes for GBA, 8,388,608 for Dreamcast, 131,072
events and 16 MiB of command metadata each. Overflow and pre-existing data are
reported. The tested GBA capture used roughly 16 MiB of evidence; the Dreamcast
checkpoint used 89 MiB before state copies and replay caches. A later browser
Dreamcast capture grew the heap to 427 MiB. Normal emulation uses much less.

## Optimizations and measured throughput

- GBA halfword memory reads use checked direct backing-memory paths, falling
  back for device/exceptional/boundary/watched reads. Background priorities are
  read once per scanline instead of looking up the register map inside the
  per-pixel layer loop.
- Dreamcast instruction/data and AICA ARM reads have checked direct-memory
  paths. Texture Morton addressing interleaves bits with masks/shifts instead
  of a sixteen-iteration loop. The field timer caches its division until the
  live scanline-count register changes; it still checks that register each
  instruction and preserves field/interrupt timing.
- Temporary GBA bus objects are stack-owned, and the BIOS callback owns its
  machine capture safely. Dreamcast display conversion avoids a persistent
  arena allocation for every presented frame. These lifetime fixes preceded
  the timing baseline.
- Both compile with `-O3 -msimd128 -fwasm-exceptions -fwrapv
  -ffp-contract=off`. There is no fast-math, resolution reduction, instruction
  skipping, JIT or new game-specific speed shortcut. SIMD is enabled for LLVM;
  these changes do not add a hand-written vector rasterizer.

Results are recorded in [gba-dc-performance.json](../results/gba-dc-performance.json).
Three fresh Node processes per variant ran sequentially in alternating order,
starting from identical checkpoints with 10,000-instruction host slices,
profiling on and capture off. Values are medians on the user's macOS arm64
machine, Node 24.19.0 / Emscripten 4.0.16.

| Scene | Initial C++/WASM | Optimized C++/WASM | Throughput gain |
|---|---:|---:|---:|
| Minish Cap opening sequence, 300 frames | 909.4 ms / 329.9 frames/s | 576.8 ms / 520.1 frames/s | 57.7% |
| Crazy Taxi driving checkpoint, 60 fields | 6963.6 ms / 8.62 fields/s | 5292.3 ms / 11.34 fields/s | 31.6% |

These measure core execution, not unique rendered game frames or browser canvas
FPS. GBA has ample headroom in this scene. Dreamcast remains below real time,
roughly 19% of its declared 60-field rate at the tested driving checkpoint;
reaching scenes still takes patience. In the browser Minish Cap presented about
59.5 frames/s in normal mode; Crazy Taxi driving was observed around 11–12
updates/s. Those browser observations were not controlled benchmarks.

## Validation and limitations

- The public suite includes synthetic PPU blending/blanking and source-history
  checks, TA triangles/clear, all Dreamcast scanout formats (including bank
  boundaries), intermediate/final replay, transactional state restoration and
  malformed-state rejection. The new pixel tests pass native and WASM.
- Randomized direct-read tests and 100,000 changing-register field-timer steps
  compare the fast paths against retained translated reference functions.
- Before/after benchmark machine and framebuffer proofs match exactly for
  every sample. Private native/WASM captures verify unchanged emulation with
  capture enabled, final replay equality and paused-machine isolation.
- A 600-field uncaptured soak held the WASM heap at 64 MiB for GBA and 128 MiB
  for Dreamcast, at every 60-field sample.
- GBA cold boot at 300 frames matches the Go core's CPU, RAM and display hashes.
  Dreamcast's imported driving checkpoint matches Go after one field. After
  thirty fields, CPU steps/PC, RAM and AICA RAM still match, but rendered bytes
  differ: 37,933 of 307,200 pixels, all but ten by at most one RGB565 quantum per
  channel. Native FMA contraction alone did not eliminate the difference. The
  precise raster arithmetic cause remains unresolved; do not claim bit-exact
  Go renderer equivalence. Native C++ and WASM agree, and the optimization pass
  preserves the initial C++ results.
- Chromium in-app browser validation reached the GBA title and save-slot menu
  through Start input, verified normal/turbo speed, pause capture, source-chain
  queries, replay controls, state creation and imported-state restoration.
  Dreamcast cold boot reached the VMU warning/title screens; an imported driving
  checkpoint ran, accepted trigger input and captured its rendering. Pixel
  evidence matched the displayed color, register snapshots were readable, and
  the rendering slider showed partially drawn geometry. Download-link
  activation did not yield an observable file in this
  browser test, so save-file download/reimport remains an acceptance gap.
  Physical gamepad, Firefox and Safari validation are outstanding.

## Reproduction

```sh
go run ./tools/browser/console-portgen gba
go run ./tools/browser/console-portgen dc
python3 tools/browser/console-portgen/state.py
python3 tools/platform/gba/browser/build.py --emcc /path/to/em++
python3 tools/platform/dc/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py

# Private media and checkpoints only:
CORE_DIR=tools/platform/gba/browser/web node tools/browser/tests/console-check.mjs gba /path/game.gba /path/raw.state
CORE_DIR=tools/platform/dc/browser/web node tools/browser/tests/console-check.mjs dc /path/game.cue /path/raw.state
node tools/browser/tests/console-compare.mjs dc /path/game.cue /path/raw.state /path/baseline-core tools/platform/dc/browser/web 60 3
node tools/browser/tests/console-state.mjs dc /path/game.cue /path/raw.state /path/private.rrstate
```

Raw core states use platform IDs 13/14 and version 1. The outer browser container
checks media, core and configuration identities. Regeneration preserves the Go
reference sources. Private cartridge/disc images, checkpoints and screenshots
are not committed or deployed. The source working tree was based on `b7bc38bc`;
pre-existing unrelated edits were left untouched.

Generated artwork is saved at `site/emulators/art/gba.png` and
`site/emulators/art/dc.png`. Exact image-generation prompts and mode are in
[ARTWORK.json](ARTWORK.json).
