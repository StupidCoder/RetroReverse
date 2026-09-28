# PSP browser port

The seventh static emulator is PlayStation Portable. LocoRoco boots from a
locally selected CSO, passes language selection and New game, and runs the
first stage with L/R control. It shares the existing landing page, generated
console art, Run/Pause/Reset, gamepad handling, portable save states, subsystem
timings, pixel inspection and rendering replay. No server emulator, system
firmware, uploaded game or game-specific allowlist is required.

## Implementation

The bounded Go-to-C++ source generator covers Allegrex (integer, FPU and VFPU),
the existing PSP kernel/scheduler/filesystem HLE and software GE. The host loop
is rewritten: its synthetic display clock survives each worker execution slice,
it stops at actual display-buffer presentations, and it does not build the Go
oracle's per-instruction spin-count map. Natural-width RAM access and per-draw
texture views avoid repeated byte dispatch and ownership work. Texture views
borrow live storage instead of caching decoded pixels, preserving feedback
rendering and aliases. Reference-sampler comparisons cover all six supported
texture formats, swizzles and mip levels. Floating-point contraction remains off.

The UMD adapter retains the user's `File` and reads bounded ranges in the worker.
CSO uses its index, a 256 KiB compressed window, a 1 MiB sector cache and zlib raw
DEFLATE. The full decompressed image is never allocated. Portable SHA-1 and AES
support the existing PRX/NID contracts; public vectors verify both. Unknown PRX
variants fail explicitly, following existing compatibility limits.

The raw state stores machine fields and 32 MiB RAM separately, rebinds the file
source and HLE callbacks, and restores thread contexts, VFPU and the persistent
clock. Invalid loads leave the current machine intact. The common container
pins media and core identity; hashing a large disc during first save/load can
still take noticeable time.

Capture records GE commands, register context, memory writes and rejected
fragments. Replays apply those historical effects in scratch memory and show
the captured display buffer being built. Native GPU addresses are in command
metadata; the capture memory map compacts VRAM/RAM/scratchpad. Full texture-fetch
ancestry and an offscreen-target selector remain unimplemented and are stated
in the page. The live state, final pixels and continuation survive arbitrary
backward/forward replay.

## Measurements on the development Mac

The same dense LocoRoco stage checkpoint was advanced by 120 display updates.
Loading, state import and proof hashing are outside the timed region. The browser
probe also transfers each frame to a canvas. These are observed samples, not a
realtime guarantee or a cycle-accurate hardware benchmark.

| Execution | 120 updates | Updates/s |
|---|---:|---:|
| Native C++ | 6.21 s | 19.3 |
| WASM, Node/V8 | 9.22 s | 13.0 |
| WASM, isolated Chrome without automation attached | 9.30 s | 12.9 |
| WASM, Chrome controlled through browser automation | 62.05 s | 1.9 |

The isolated browser uses a fresh worker, bounded local range reads and real
wall-clock time, with no virtual-time acceleration. All four end proofs match:
CPU PC, instruction count, RAM, VRAM and framebuffer. The early first-stage
checkpoint is lighter: 120 updates in 4.89 s, about 24.6/s. The language screen
exceeded 100 updates/s even in the automated browser. Software GE rasterization
dominates gameplay; this is useful educational performance, not full-speed PSP.

Do not infer ordinary browser throughput from a debugging session. The direct
controlled-browser result is retained above rather than hidden. V8 documents
that attaching a debugger can change WebAssembly compilation tiers:
[WebAssembly compilation architecture](https://github.com/v8/v8/blob/main/docs/wasm/architecture.md).
The clean Chrome result establishes that browser delivery itself does not
explain the sixfold slowdown seen during this automation session; the exact
instrumentation contribution was not isolated.

## Validation and limitations

- Native and WASM cold boot agree after 12,000,000 instructions: identical CPU
  PC, 32 MiB RAM, 2 MiB VRAM and framebuffer hashes.
- Native state loads into WASM exactly. A dense gameplay continuation agrees at
  instruction 2,116,837,106 / presentation 4,480 across native, Node and Chrome.
- The dense capture has 8,852 GE events and 1,209,452 writes, no overflow.
  Replay final equals the live display, arbitrary seeks are repeatable, and
  replay leaves serialized state and subsequent execution unchanged.
- Public checks cover crypto, raw/compressed CSO random reads, malformed offsets,
  texture aliasing/format equivalence, capture/replay, scanout color decoding,
  timing across different execution slice sizes and transactional state load.
- Browser UI checks cover local image load, cold boot, restore, Run/Pause,
  capture, pixel evidence, and replay controls. The WASM binary is about 1 MiB.

MPEG/AVC and audio decoding are not part of the existing core and are not added
here. Video intervals can remain black until the stream completes or the game
accepts a skip input. Burnout's PRX decrypts and its boot executes, but its
browser gameplay is not validated. The existing HLE, texture and graphics
coverage limits apply to arbitrary UMD images. Game images and private
checkpoints are intentionally excluded from the repository and deployment.
