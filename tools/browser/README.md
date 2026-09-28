# RetroReverse browser emulators

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
remaining compatibility/validation limits. Save-state files, automatic provenance
capture, pixel selection and rendering replay belong to U4 and later.

Console illustrations were generated with the built-in image generator. Their
four exact prompts are in `docs/ARTWORK.json`; assets are in `site/emulators/art/`.
