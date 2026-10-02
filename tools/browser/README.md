# RetroReverse browser emulators

Code inspection and game knowledge work is tracked in the
[M0 contracts and baseline](docs/code-m0/README.md) and
[M1 knowledge schema/package acceptance](docs/code-m1/README.md).
The [C64 Code workspace and debugger](docs/code-m2/README.md) implement M2. See [knowledge authoring](../knowledge/README.md)
for validation/export commands. Annotation-only releases can use
`python3 tools/browser/package.py --site-only` to reuse verified packaged cores.

The deployed application is `site/emulators/`; it is part of the existing
no-build Cloudflare Pages site. The homepage links to its console gallery.
All execution is in a Web Worker. Game files remain local, C64 firmware is
served from the site, and disc images use bounded local File reads.

## Preview and rebuild

From the repository root:

```sh
python3 tools/browser/preview.py --port 8772
# http://127.0.0.1:8772/emulators/
```

This preview applies `site/_headers`, including cross-origin isolation for
higher-resolution profiling timers. The headers are scoped to the emulator
area; the existing asset viewers can still use their CDN dependencies.

Rebuild with an installed Emscripten SDK (em++ on PATH), Python 3 and clang++:

```sh
python3 tools/platform/c64/browser/build.py
python3 tools/platform/psx/browser/build.py
python3 tools/platform/n64/browser/build.py
python3 tools/platform/threedo/browser/build.py
python3 tools/browser/package.py
```

Each build also accepts `--emcc /path/to/em++`. The static JS/WASM artifacts are
committed so Cloudflare does not need these tools. Packaging regenerates hashes
and checks the per-asset size limit. Unhashed core filenames are revalidated via
`Cache-Control: no-cache` to avoid mixing versions after deployment.

If regenerating the N64/3DO Go translations, reapply the host profiling and
file-I/O adaptations, then regenerate the resumable scheduler before building:

```sh
python3 tools/browser/instrument.py
python3 tools/browser/capture-instrument.py
python3 tools/platform/threedo/browser/optimize.py
python3 tools/browser/state/generate.py
python3 tools/platform/threedo/browser/slice.py
```

The original Go emulators remain reference implementations. No game-specific
asset extraction or server API is required by the deployed emulator shell.
The older prototype launcher/servers in this directory are development tools.

## Checks and scope

```sh
node tools/browser/tests/media.mjs
node tools/browser/tests/input.mjs
clang++ -O2 -std=c++20 tools/browser/tests/runtime.cpp -o /tmp/rr-runtime
/tmp/rr-runtime
```

The C++ slice drivers in `tests/` take local reference images; they compare
machine-state hashes across execution budgets. `results/` contains the recorded
checkpoints, not game data. See `docs/U0-CONTRACTS.md`, `docs/U1-ACCEPTANCE.md`,
`docs/U2-ACCEPTANCE.md`, and `docs/U3-ACCEPTANCE.md` for the implementation and
remaining compatibility/validation limits. U4–U6 now add portable states, automatic
Pause capture and a shared pixel inspector. See `docs/U4-ACCEPTANCE.md`,
`docs/U5-ACCEPTANCE.md`, and `docs/U6-ACCEPTANCE.md`. Rendering replay is implemented in U7. See `docs/U7-ACCEPTANCE.md`,
`docs/U8-RELEASE.md` and `docs/FRAME-PACING.md` for release checks and pacing.

Use `python3 tools/browser/build-release.py --emcc /path/to/em++` for the
pinned release rebuild, `python3 tools/browser/check.py` for the public CI suite,
and `python3 tools/browser/check-release.py` for a fast artifact audit.

Additional checks: `node tools/browser/tests/inspector.mjs`, the four
`tests/pixel-*.cpp` native drivers, and `tests/capture-wasm.mjs PLATFORM IMAGE`.
Set `NATIVE_STATE` to a local raw native checkpoint and `PIXEL_GRID=1` for
85 sampled pixels per interval. `scene-3do.cpp` produces
reference-game checkpoints from local media; never publish those checkpoints.
Owned native C64 tests run through `tools/platform/c64/owned/check.py`; translated cores can use
`-Wno-parentheses-equality`.

Console illustrations were generated with the built-in image generator. Their
exact prompts are in `docs/ARTWORK.json`; assets are in `site/emulators/art/`.

The 3DO hot-path and movie playback refactor is documented in `docs/3DO-PERFORMANCE.md`.

The [3DS acceleration prototype plan](docs/3DS-ACCELERATION-PLAN.md) covers an
optional execution mode, WebGPU rendering and verified reference-debugger handoff.

Game Boy and Game Gear ports are documented in `docs/GB-GG-PORTS.md`.
The separate Game Boy, Game Gear and C64 Play/Render workspaces and historical
scanline layers are documented in `docs/GB-RASTER-INSPECTION.md`,
`docs/GG-RASTER-INSPECTION.md` and `docs/C64-RASTER-INSPECTION.md`.

The new Amiga 500 chipset model and Marble Madness/Turrican validation are documented in `docs/AMIGA-PORT.md`.

The experimental PS2 and GameCube ports, rendering-capture scope, performance measurements, and reproduction commands are documented in `docs/PS2-GC-PORTS.md`.

Game Boy Advance and Dreamcast ports, optimization measurements, media formats and capture limits are documented in `docs/GBA-DC-PORTS.md`.

MS-DOS PC and Xbox share the x86 C++ port. See `docs/DOS-XBOX-PORTS.md` for executable/disc formats, tests, performance and capture limits.

Amiga’s separate scanline/Copper and blitter/mask views are documented in
[Amiga raster inspection](docs/AMIGA-RASTER-INSPECTION.md).

The common Play/Render shell, adapter API and UI checks are documented in
[Shared emulator UI](docs/SHARED-UI.md).

[M3](docs/code-m3/README.md) adds shared Code/Memory state panels and bounded structured decoding.

## Owned C64 release

The C64 production build now compiles `tools/platform/c64/owned`. See the
[C7 report](../platform/c64/owned/C7-RELEASE.md). The old vendor-specific native
fixtures (`debug-c64`, `memory-c64`, `pixel-c64`, `raster-c64`, `state-c64`)
are superseded by the owned `browser`, `render`, `debugger` and portable `state`
suites, each run natively, with UBSan and in WASM. Shared JS debugger, tour and
experiment tests now run against the shipped owned WASM. `browser-fort.mjs`
replaces the old private `scene-c64` fixture with authentic boot and 272-line
rendering/provenance checks. The legacy state-field generator is retired.

The pinned rollback page is `site/emulators/c64/legacy.html`; packaging retains
its complete executable bundle and validates all hashes from `c64/rollback.json`.
Historical C64 screenshots/reports describe the old core where noted.
