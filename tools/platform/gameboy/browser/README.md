# Game Boy C++ / WebAssembly core

Build with `python3 tools/platform/gameboy/browser/build.py --emcc /path/to/em++`.
Regenerate the bounded Go translation with `go run ./tools/browser/handheld-portgen gameboy`.

See [handheld port documentation](../../../browser/docs/GB-GG-PORTS.md) for
rendering, capture addresses, state format, validation and compatibility limits.
The native executable is `work/handheld-native ROM [FRAMES] [PPM]`; generated browser
artifacts are in `web/` and copied into the static site by `tools/browser/package.py`.
