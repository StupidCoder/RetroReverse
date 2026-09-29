# MS-DOS PC and Xbox browser ports

These ports translate the repository's Go x86 interpreter and DOS/Xbox platforms
to C++20, then compile them with Emscripten. They do not incorporate DOSBox,
86Box, QEMU, xemu or another emulator. Both use the same generated CPU source
and the shared bus, memory fast paths and local-file adapter. Separate WASM
modules keep platform memory and lifecycle independent.

## Scope

The CPU includes the existing 16/32-bit x86, x87, MMX and SSE implementation.
This is a Pentium III-era instruction interpreter, not a cycle-exact or complete
Pentium III PC. DOS runs real-mode MZ and DJGPP go32/COFF executables through
DOS/BIOS/DPMI HLE, with VGA, keyboard and real-mode mouse support. It does not
boot DOS from a disk, run Windows, implement paging or support arbitrary DOS
extenders. It has no audio output. The game's entire folder is selected locally;
an EXE selector accepts unfamiliar executables as well as the tested games.

The existing Quake DJGPP base-address accommodation is restricted to the known
409,600-byte executable, FNV-1a `b155c323`, and the compatibility toggle. Other
executables do not receive this patch. No game data is distributed.

Xbox includes kernel HLE, cooperative threads, USB/XID controller handling,
disc/cache-partition I/O and software NV2A rendering. It accepts XDVDFS ISO/XISO
images, including the existing supported full-disc partition offsets, up to
16 GiB. It does not need a BIOS. The left stick, digital buttons and pressure
buttons are exposed; right-stick controls and audio output are not exposed.
Kernel, GPU and peripheral compatibility remains experimental.

## Browser integration

Both pages use the existing worker shell: run, pause, reset, input, subsystem
timings, local states, automatic pause capture, pixel history and rendering
replay. Generated illustrations were added to the sixteen-system gallery;
prompts are in `ARTWORK.json`.

Selected File objects stay in the worker. DOS reads individual file slices;
Xbox uses the shared 256 KiB disc window. Neither uploads game files or reads
an entire ISO into memory. DOS writes use a virtual overlay, included in states,
without changing the user's files. States include open-file offsets and input
state. Xbox states retain shared thread/object aliases and its virtual HDD
cache. Derived GPU caches are rebuilt. Browser states require identical media,
configuration and core hashes, and failed imports leave the running machine
intact. Initial media hashing for a large Xbox ISO can take tens of seconds.

DOS display counters are instruction-budget intervals, not measured game FPS.
The capture waits for VGA writes and a short quiet period, with a bounded
timeout. It records CPU PCs, VRAM, palette and display-register writes. One REP
instruction can account for most of a software-rendered VGA update, so a capture
can legitimately contain only one writer event. Pixel details also link to the
palette entry. This is not instruction-level ancestry through Quake's software
rasterizer.

Xbox capture ends at the next modeled flip. It records NV2A draws, RAM writes
and Kelvin register snapshots; antialiased pixels expose their individual stored
samples. Replay uses separate memory and does not advance the paused machine.
It does not trace rejected fragments, sampled textures or the CPU instructions
which assembled a push buffer. The modeled flip retains the Go core's scanout
limitations. Detailed race captures can exceed 1 GiB of WASM memory.

## Optimizations and porting corrections

* Direct 16/32-bit fetches and operands translate/check one RAM span rather than
  repeating byte accesses. Segment wrapping, 20-bit wrapping, MMIO, VGA, watches
  and capture retain their slow paths.
* Endian readers borrow temporary byte spans instead of repeatedly allocating
  shared ownership records.
* NV2A combiner mapping/scaling processes RGBA with explicit WASM SIMD. Operation
  order, NaNs and signed-zero behavior are checked against the scalar version;
  fast-math is not enabled.
* Texture entries have explicit ownership and bounded collection at flips.
  Replacing/invalidation no longer leaves decoded textures in the global arena.
* Go shadow declarations such as `ord := ord` must read the outer variable
  before declaring the C++ local. Fixing this restored the Xbox suspend/resume
  kernel handlers. Stored texture-decoder closures also need value captures.
* SHA-1 includes incremental state import/export compatible with the Go kernel's
  representation. The implementation is local code, not a new dependency.

## Validation

The public native suite and the four additional actual-WASM tests cover x86
wrapping, unaligned access, VGA and watches; synthetic MZ execution; virtual
file overlays; combined keyboard/controller input; VGA palette provenance;
NV2A triangles and replay; state corruption/rollback and pointer aliases;
thread-handler closures; SHA-1 continuation; texture ownership; and SIMD/scalar
equivalence for random float bit patterns, infinities and NaNs.

Private local media checks include Ultima Underworld's native intro and Quake's
native/browser cold boot, menu and running demo. Quake's browser workflow was
tested through folder selection, keyboard input, state import, pause, pixel
selection and first/final replay. A gameplay checkpoint also passed deterministic state
round trips and replay equality in Node/WASM.

OutRun's frame 250 cold-boot SEGA image is byte-identical between Go, native C++
and Node/WASM (RGBA FNV `86c4f3a5`). At frame 3,000, native C++ and WASM both show
the same white loading image; the extended loading phase is not a browser-only
display failure. Do not treat the initial logo as evidence of complete menu
acceptance. The separately restored driving checkpoint renders the race in the
browser. Small numerical differences from Go remain in 3D rendering.

The browser Xbox acceptance included local ISO selection, state import, resume,
pause/capture, pixel contributors, and first/intermediate/final render replay.
A 120-frame Node/WASM race soak stayed at 430 MiB, versus the explicit 600 MiB
test budget. Detailed browser capture grew the heap to 1,403 MiB; ordinary running
and detailed inspection have very different memory costs.

## Performance

Sequential fresh-process A/B runs on this Mac, Node 26.3.1 / arm64, three samples
per variant, loading the same checkpoint before each sample. These measure core
execution, not browser canvas delivery. The reference disables the new direct
x86 accesses and explicit NV2A SIMD; both variants include correctness fixes and
bounded texture ownership. CPU, RAM and framebuffer proofs match across all
six samples for each workload.

| Workload | Scalar reference median | Optimized median | Speedup |
| --- | ---: | ---: | ---: |
| Quake, 300 synthetic intervals | 4,755 ms | 4,100 ms | 1.16× |
| OutRun race, 3 flips | 2,023 ms | 1,506 ms | 1.34× |

Quake retires about 40 million interpreted instructions/s in this window. The
73 synthetic intervals/s figure is **not** 73 rendered game frames/s. OutRun
improves from 1.48 to 1.99 rendered flips/s. Its rasterizer median drops from
1,367 to 898 ms per three flips; it remains the largest cost. This is useful for
inspection, not real-time Xbox play. Profiles, hashes and individual samples
are in `../results/dos-xbox-performance.json`.

Cold boot is a different workload: reaching OutRun's frame 250 took 36.5 seconds
in one Node/WASM run. The later white loading phase is expensive and should not
be presented as a completed interactive cold-boot acceptance. Race checkpoints
avoid repeating that loading phase, but are private game-derived files and are
not hosted with the site.

## Reproduction

Run from the repository root. Emscripten 4.0.16, clang++, Go, Python and Node are
needed for development; the deployed site only serves committed static files.

```sh
go run ./tools/browser/console-portgen dos
go run ./tools/browser/console-portgen xbox
python3 tools/browser/console-portgen/state.py
python3 tools/platform/dos/browser/build.py --emcc /path/to/em++
python3 tools/platform/xbox/browser/build.py --emcc /path/to/em++
python3 tools/browser/tests/check-x86-wasm.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
```

Both build scripts accept `--reference`, emitting the scalar/byte-access build
under `browser/work/reference` without replacing the shipping core. For a
private-media A/B test:

```sh
node tools/browser/tests/console-compare.mjs xbox /local/game.iso \
  /local/race.state tools/platform/xbox/browser/work/reference \
  tools/platform/xbox/browser/web 3 3
```

Checkpoints, game files, screenshots and generated disassemblies stay local and
must not be committed. `tests/console-state.mjs` wraps private native checkpoints
for browser import. `tests/console-check.mjs` verifies state/replay determinism.
