# Game Gear C++ / WebAssembly core

Build with `python3 tools/platform/gamegear/browser/build.py --emcc /path/to/em++`.
Regenerate the bounded Go translation with `go run ./tools/browser/handheld-portgen gamegear`.

See [handheld port documentation](../../../browser/docs/GB-GG-PORTS.md) for
rendering, capture addresses, state format, validation and compatibility limits.
The [raster workspace and timing update](../../../browser/docs/GG-RASTER-INSPECTION.md)
documents the C++ Z80/VDP timing, source previews and version 2 states.
The native executable is `work/handheld-native ROM [FRAMES] [PPM]`; generated browser
artifacts are in `web/` and copied into the static site by `tools/browser/package.py`.
