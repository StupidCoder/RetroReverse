# PSP C++ / WebAssembly core

The existing Go Allegrex/VFPU, PSP kernel HLE and software GE interpreter,
compiled as C++20 with browser-specific memory, UMD, timing, state and capture
adapters. The unified page is `site/emulators/psp/`.

## Build

From the repository root, with Go, Python, Clang, Emscripten and native zlib:

```sh
go run ./tools/platform/psp/browser/portgen
python3 tools/browser/state/generate.py
python3 tools/platform/psp/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
```

Emscripten fetches its pinned zlib port on the first build. Generated C++ is
checked in; Go is not needed for ordinary rebuilds. `runtime.h` shares the
reviewed Go-value compatibility types with the DS and 3DO ports. The generator
rejects unsupported source constructs instead of silently omitting them.

## Browser behavior

- ISO or CSO v1 UMD images, 2048-byte sectors, up to 4 GiB. No game allowlist.
- Retained local `File`, `FileReaderSync` in the worker, bounded random reads.
  CSO index is in memory, with a 256 KiB compressed window and 512 decoded
  sectors (1 MiB). Neither the complete disc nor an expanded CSO is uploaded.
- Allegrex integer/FPU/VFPU execution, kernel imports, thread scheduling,
  asynchronous filesystem HLE, GE display lists and software rasterization.
- One persistent synthetic VBlank clock across browser execution slices;
  display updates occur on `sceDisplaySetFrameBuf`, matching the Go debugger.
- Digital buttons and independent analog stick; keyboard and standard gamepad.
- Portable raw state format 7/1, wrapped by the common hashed `.rrstate`
  container. RAM, VRAM, CPU/VFPU, thread contexts, file positions, kernel objects,
  GE registers, buttons and clocks persist. Immutable disc contents and host
  callbacks do not; callbacks are rebound after a validated transactional load.
- Exclusive wall-time profile: CPU/scheduler, kernel HLE, GE commands,
  vertices/rasterization, UMD reads and display conversion.

## Capture and replay

Pause captures the next presentation interval. GE command words, relevant
register state, CPU/GPU framebuffer writes and rejection tests are recorded.
Rejected fragments distinguish depth, alpha, stencil, scissor and write masks.
Backward and forward scrubbing replays historical writes in separate memory;
it cannot change guest CPU state or re-read a live texture.

Capture addresses use a compact map: VRAM at 0, main RAM at `0x200000`,
scratchpad at `0x2200000`. The physical GPU addresses are included in command
metadata. The replay screen follows the final captured scanout buffer. Offscreen
writes are retained, but selecting an offscreen target and per-texel ancestry
remain future work. A pixel may legitimately have no writes in this interval.
The recorder caps writes at four million and metadata at 16 MiB, reports overflow,
and never claims complete evidence after reaching a cap. Tested stage captures
use about 190 MiB of evidence, plus state checkpoints and replay memory.

## Compatibility

LocoRoco is exercised from cold boot, through language selection and New game,
into the player-controlled first stage. Cross confirms, left/right changes the
New/Continue choice, and L/R tilt the stage. Burnout Legends PRX decryption and
boot execute, but gameplay has not yet been validated in the browser port.

This inherits the experimental Go emulator's HLE and graphics limitations.
Supported encrypted PRX tags are the existing LocoRoco and Burnout variants;
other tags fail explicitly. MPEG/AVC and audio codecs are not implemented.
Video streams advance without pictures and may remain black until completion
or a game-supported skip input. The memory-stick utility is HLE; browser save
states are the supported way to preserve progress.

## Validation

`tools/browser/tests/pixel-psp.cpp` checks public SHA-1/AES vectors, CSO random
reads and malformed offsets, all six texture formats against the scalar sampler,
live texture aliasing, frame reconstruction, rejected-fragment provenance,
backward seeks, persistent timing across slice sizes and transactional states.
It runs in the normal public CI without copyrighted files.

Private media/checkpoints can be supplied to these reproducible harnesses:

```sh
node tools/browser/tests/benchmark-psp.mjs /path/to/game.cso /path/to/raw.state
node tools/browser/tests/validate-psp.mjs /path/to/game.cso /path/to/raw.state
PSP_FRAMES=500 PSP_SAVE=/tmp/psp.state tools/platform/psp/browser/work/psp-native /path/to/game.cso 100000000 /tmp/frame.ppm
```

The native harness accepts `PSP_LOAD`, `PSP_SAVE`, `PSP_FRAMES`, `PSP_CROSS`
(periodic Cross input), `PSP_TRACE`, `PSP_CAPTURE` and `PSP_LOG`. The WASM
harnesses accept `CORE_DIR` to compare an un-packaged build. Results and browser
measurement caveats are in `tools/browser/docs/PSP-PORT.md`.
