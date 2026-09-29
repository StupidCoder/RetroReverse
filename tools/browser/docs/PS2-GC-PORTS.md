# PS2 and GameCube browser ports

Both ports translate RetroReverse's existing Go emulators to C++20 and WASM.
They do not embed PCSX2, Dolphin, or another third-party console emulator.
The Go emulators remain the reference and retain their own compatibility limits.
Emscripten/LLVM and their standard runtime libraries provide compilation and
browser support, as with the other C++ ports.

## Implementation

- GameCube: Gekko interpreter, paired-single/FPU instructions, GX command
  processor, vertex transforms, TEV software rasterizer, pixel-engine copies,
  VI, DSP and platform devices. The disc's real apploader executes after the
  existing IPL setup substitute.
- PS2: Emotion Engine/R5900, IOP/MIPS, VU0/VU1, VIF/GIF/DMA, GS software
  rasterizer, kernel HLE and disc/IOP module handling. A bounded ELF32/IRX reader
  replaces the Go `debug/elf` host dependency.
- Workers retain a local File and read slices on demand. The disc cache is
  256 KiB; neither a 1.4 GiB GameCube disc nor a multi-GiB PS2 disc is copied
  wholesale into WASM. Media is not uploaded.
- PS2 accepts ISO and single data-track BIN/CUE images. GameCube accepts raw
  ISO/GCM. File selection is not restricted to known game hashes.
- Jak and Daxter contains its own IOP image. Games needing additional IOP
  firmware can use a local 4 or 8 MiB PS2 BIOS through the optional firmware
  control. No PS2 or GameCube firmware is distributed.
- Run, Pause, Reset, Next frame, keyboard/standard gamepad input, subsystem
  timings, save/load, automatic capture, pixel inspection and render replay
  use the shared browser shell. Both consoles have generated gallery artwork.
- Portable states preserve nil versus allocated-empty Go slices, hardware
  state, CPU/device callbacks after restoration, and PS2 scheduler phase
  across execution slices. State headers have independent platform IDs and
  version 2; the outer container checks media, firmware and core identities.

Audio playback, external memory-card files, PS2 right-stick and GameCube
C-stick input are not exposed. Browser save states provide session persistence.
This remains experimental emulation, not a claim of general commercial-game
compatibility or cycle-exact hardware behavior. The original Go core's boot
substitutes and game compatibility workarounds are retained, not newly hidden
shortcuts in the browser port.

## Rendering evidence

PS2 captures actual GS VRAM writes from primitives, image transfers and local
copies, including depth writes. Draw events retain vertex information and all
128 GS registers. Four video fields include double-buffer producer context.
The replay rebuilds the final display buffer in command order.

GameCube captures GX commands, EFB color writes, clears, texture/display-copy
writes and 256 BP registers at draws. Three fields account for VI display delay
and double buffering. The pixel query follows the EFB value before the copy
that supplied the selected VI display buffer. The replay shows EFB construction
and then the YUY2 display copy. RGB-to-YUY2 conversion is lossy and shares chroma
between neighboring pixels; the UI distinguishes EFB byte reconstruction from
scanout-color equality.

These are recorded rendering writes, not a scanline wipe over a finished frame.
Rejected fragments, sampled-texel histories, VU instruction ancestry and the CPU
instructions that originally constructed command buffers are not captured.
Submission PCs are labeled as submission PCs. Register snapshots are inspectable
as bytes; a full structured GPU-state UI remains future work.

Capture is bounded (GC: 8,388,608 writes/131,072 events; PS2: 8,388,608
writes/131,072 events). PS2 register/command metadata is capped at 128 MiB. Overflow is reported. Private captures used hundreds of
MiB, so these two inspectors are aimed at desktop browsers. Normal uncaptured
execution starts with a 128 MiB WASM heap. Render replay operates on separate
memory and does not advance or change the paused guest.

## Optimizations

1. Compile with native WebAssembly exceptions (`-fwasm-exceptions`). The first
   translation used Emscripten's JavaScript exception machinery. Profiling
   showed frequent WASM-to-JavaScript invoke/stack-handling calls even without
   guest exceptions. Native WASM exceptions remove those crossings.
   See [Emscripten's exception documentation](https://emscripten.org/docs/porting/exceptions.html).
   A browser supporting WASM exception handling and SIMD is required.
2. Borrow short-lived endian-helper spans instead of incrementing shared slice
   ownership for every instruction/texel access. Large read-only GC vertex and
   texture records are passed by reference.
3. PS2 instruction fetch has a checked direct-RAM path. Other address spaces
   continue through the original memory/device decoding.
4. PS2's spin detector needs to distinguish at most six instruction addresses;
   a saturating seven-entry set replaces per-instruction hash-map work.
5. PS2's unhandled-exception check exits immediately unless the PC is an
   exception vector, avoiding a map/string lookup on every guest instruction.

6. PS2 texture samplers are stack-owned for the duration of each synchronous
   draw. The first translation placed these temporary Go pointers in a
   machine-lifetime arena, accumulating a sampler for every textured primitive.
   A longer browser run caught growth to roughly 1.5 GiB before capture; the
   stack-owned form removes that leak and its repeated heap allocations.

The builds use `-O3 -fwrapv -ffp-contract=off`, without fast-math or changes to
texture/render resolution. The existing serial Go raster paths are used;
thread pools and shared-memory browser workers are not introduced.

## Validation and performance

See the recorded measurements in `../results/ps2-gc-performance.json`.
The baseline is the initial C++/WASM translation from this porting session,
before the above optimizations. Timed runs use identical restored checkpoints,
60 emulated video fields, 10,000-instruction host slices, profiling enabled and
capture disabled. Each WASM case runs twice in a fresh Node process, sequentially.
These are core throughput measurements, not browser canvas FPS or unique game
frames; a game may render at half its video-field rate.

Measured on the user's macOS arm64 machine, Node 24.19.0 / Emscripten 4.0.16:

| Test scene, 60 fields | Initial WASM | Optimized WASM | Speedup | Optimized fields/s |
|---|---:|---:|---:|---:|
| Luigi's Mansion, foyer checkpoint | 38.69 s | 16.43 s | 2.36× | 3.65 |
| Jak and Daxter, early title/logo checkpoint | 10.51 s | 4.35 s | 2.41× | 13.78 |
| Jak and Daxter, later 3D opening scene | 31.12 s | 15.82 s | 1.97× | 3.79 |

Optimized native C++ took 12.09 s and 3.28 s in the first two cases, versus
16.43 s and 4.35 s in WASM (about 1.36× and 1.33× the native time). These are
single native comparison runs; native timing includes the final proof hash.
The file named `gameplay-go.state` in the results is an older project's named
checkpoint showing a later opening scene, not evidence of free-roaming gameplay.

At the foyer checkpoint GC now splits most time between Gekko/device execution
and TEV rasterization. The later Jak scene spends roughly 40% in vector units
and 36% in GS rasterization. Further substantial gains would require deeper CPU/VU
interpreter specialization and renderer work. The browser UI has additional
scheduling and display costs, so the table should not be read as a promise of
real-time play or constant browser FPS.


Public tests use synthetic GX/GS primitives, overlapping/depth-tested pixels,
clear/display copy, historical pixel reconstruction, intermediate replay,
final-frame equality, timer-phase preservation and deterministic state restore.
Private-media checks additionally test truncated-state rejection, rollback on
failure, capture/replay isolation and capture-on/off equality.

Native and WASM restored-state runs agree on the recorded output proofs before
and after optimization. Early GC cold boot also matched Go exactly. Longer
cold Go/C++ runs are not yet bit-identical; PS2 also changes scheduler-phase
handling to remain stable across browser slices. Do not treat the current
checks as proof of complete Go/hardware equivalence.

Ridge Racer V also passed a native 60-field boot smoke test from raw BIN using
a local SCPH-10000 BIOS. That is not a gameplay compatibility claim.
The port was generated from the existing working tree based on `be9b1ffb`;
pre-existing Go-source edits were left untouched.

## Browser acceptance

Tested in the in-app desktop browser with local game files:

- GameCube cold boot through title/menu into the intro; Start/A controls;
  restored foyer gameplay checkpoint; Next frame and Reset; automatic Pause
  capture, intermediate replay, pixel histories and register snapshots.
- PS2 cold boot through the memory-card prompt to the title; Cross input and
  Reset; restored later 3D scene; automatic Pause capture, render slider,
  matching pixel-color reconstruction and GPU register bytes.
- Both platforms: download a state, reload it against the local disc and return
  to the same paused video field. Initial full-disc identity checks took about
  24–30 seconds; this is local hashing, not an upload.
- PS2's sampler fix passed a 600-field uncaptured soak from the later 3D scene.
  WASM memory stayed at 193,331,200 bytes (184.4 MiB) at every 60-field sample.
  A public synthetic regression creates 100,000 samplers and verifies that the
  machine-lifetime allocation arena does not grow.

The GameCube foyer capture used about 499 MiB of evidence/checkpoints and a
1,390 MiB peak WASM heap; the PS2 scene used about 575 MiB and 1,338 MiB.
These memory peaks apply to detailed capture, not ordinary execution after the
sampler fix. Capture storage is retained for replay; resetting creates a fresh
worker. Cross-browser and physical-controller testing remains outstanding.

## Reproduction

```sh
# Regenerate only when changing the Go reference or translator:
go run ./tools/browser/console-portgen gc
go run ./tools/browser/console-portgen ps2
python3 tools/browser/console-portgen/state.py

python3 tools/platform/gc/browser/build.py --emcc /path/to/em++
python3 tools/platform/ps2/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py

# Private media; checkpoints and game images must remain local:
CORE_DIR="$PWD/tools/platform/gc/browser/web" node tools/browser/tests/console-check.mjs gc /path/game.iso /path/raw.state
CORE_DIR="$PWD/tools/platform/ps2/browser/web" node tools/browser/tests/console-bench.mjs ps2 /path/game.iso /path/raw.state 60
```

`tests/oracle-console` exports private Go checkpoints to the portable C++ state
format for differential testing. `console-play.mjs` applies bounded button/stick
input and writes private checkpoints and screenshots. No game data, screenshots,
checkpoints or BIOS files are added to the public deployment.

Artwork generation mode, exact prompts and output paths are recorded in
`ARTWORK.json`; assets are `site/emulators/art/{ps2,gc}.png`.
