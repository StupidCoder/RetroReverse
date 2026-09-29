# PS2 browser core

C++20 translation of this project's Go emulator, plus a browser host, portable
states and rendering evidence. No external console emulator is embedded.
See `../../../browser/docs/PS2-GC-PORTS.md` for validation and limits.

```sh
# From repository root. Only needed after changing the Go reference:
go run ./tools/browser/console-portgen ps2
python3 tools/browser/console-portgen/state.py
# Native clang++ executable and Emscripten WASM:
python3 tools/platform/ps2/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
```

`core/generated.cpp` and `core/state-fields.h` are generated; other core files
are hand maintained. Generation includes the measured hot-path specializations,
profiling, and write-capture hooks. `work/` and generated `web/` bundles are ignored.
The published bundle is tracked in `site/emulators/`.
