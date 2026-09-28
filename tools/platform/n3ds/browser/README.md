# Nintendo 3DS browser core

C++20/WASM port of the repository's process-level ARM11 / Horizon HLE and PICA200
software renderer. It accepts decrypted NCSD/CCI images (`.cci` or `.3ds`), up to
1 GiB, supplied locally. It does not emulate the ARM9/system boot firmware, install
CIA packages or decrypt commercial images. Game compatibility follows the existing
experimental Go core.

```
go run ./tools/platform/n3ds/browser/portgen
python3 tools/browser/state/generate.py
python3 tools/platform/n3ds/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
```

The checked-in generated C++ implements the CPU, cooperative scheduler, kernel and
service HLE, filesystem, DSP mixer, PICA command processor, shaders, lighting,
texture decoding, rasterization and GX transfers. Go's work pool becomes a serial
worker; this build requires no SharedArrayBuffer or cross-origin isolation. Offline
asset/container utilities are omitted; ETC1/ETC1A4 decoding remains included. The
browser exposes both screens, touch, buttons and circle-pad injection. The mixer
runs for correct guest progress but browser audio is not exposed.

The portable state format serializes mutable machine state and reconnects CPU
callbacks, mapped pages and read-only RomFS sessions. Game bytes, profiling, shader
and texture caches are excluded. Cached texture objects have explicit ownership;
temporary decoded images are freed after use. Uploads still reside in WASM memory,
so this first port is intended for desktop browsers.

Capture observes PICA register writes, actual raster writes/rejected fragments and
GX fill/copy/display transfers. Captured graphics-memory regions use packed local
offsets, while command metadata retains guest addresses. Display transfers link to
historical source pixels. Texture-sampler ancestry is not yet recorded. The final
LCD mapping (including rotation and the narrower lower panel) is used for replay.
Replay neither executes guest code nor changes the paused machine. Detailed
captures are bounded at eight million writes and explicitly report overflow.

Private-media checks:

```
go run ./tools/platform/n3ds/browser/tests/oracle /path/game.cci 180 > /tmp/3ds-go.jsonl
CORE_DIR="$PWD/tools/platform/n3ds/browser/web" node tools/platform/n3ds/browser/tests/validate.mjs /path/game.cci
```

The test writes temporary checkpoints under `/private/tmp`. `--resume` reuses its
own preceding checkpoint for capture iteration. It is a developer option, not a
runtime game workaround. See `tools/browser/docs/HANDHELD-PORTS.md` for evidence.

The renderer now keeps local helper functions as concrete C++ lambdas and borrows
non-escaping render-target views. See `tools/browser/docs/HANDHELD-PERFORMANCE.md`
for before/after measurements and validation.
