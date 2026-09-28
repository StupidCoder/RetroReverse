# Nintendo DS browser core

See [color and rendering-replay corrections](../../../browser/docs/DS-RENDERING.md)
for the 3D/compositor format fix and active-target replay behavior.

The original ARM9/ARM7, cartridge/SPI, DMA/timers, GX rasterizer and two 2D engines
are translated into C++20. `portgen` is a bounded AST translator for these source
files, not a general Go compiler. `generated.cpp` is checked in; browser consumers
need neither Go nor a server. Firmware is synthesized; BIOS calls use the existing
HLE. No game image is bundled.

From the repository root:

```
go run ./tools/platform/nds/browser/portgen
python3 tools/browser/state/generate.py
python3 tools/platform/nds/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
```

`core/api.cpp` supplies the shared worker API, bounded instruction slices, stylus,
controls, subsystem profiling, field-wise states and capture/replay. State restore
rebinds CPU buses, BIOS callbacks and VRAM mappings, preserving the cartridge
outside the snapshot. Transient bus objects use stack storage in hot paths;
returned BIOS closures retain their bus.

The recorder uses explicit capture-local RGBA surfaces for engine A, engine B and
3D. It records polygon fragments/rejections, 2D layer priority, blend decisions,
master brightness and final writes. A winning 3D layer links to its historical
source surface. It does not yet trace individual BG/OBJ/3D texture fetches to
physical VRAM. Replay operates on scratch surfaces with the final screen mapping.

Private-media validation:

```
go run ./tools/platform/nds/browser/tests/oracle /path/game.nds 700 touch > /tmp/ds-go.jsonl
CORE_DIR="$PWD/tools/platform/nds/browser/web" node tools/platform/nds/browser/tests/validate.mjs /path/game.nds /tmp/ds-go.jsonl 700 --touch
```

The optional `touch` sequence is for the tested European Super Mario 64 DS image.
The core and UI contain no title-specific boot script. An unfamiliar NDS image is
accepted, subject to the original emulator's device coverage. See
`tools/browser/docs/HANDHELD-PORTS.md` for measured results and limitations.

Performance work and the ARM9 CP15 WFI correction are documented in
`tools/browser/docs/HANDHELD-PERFORMANCE.md`. `core/fast.h` contains bounded bus
fast paths; the generator retains reference bus functions for regression checks.
Raw states are version 2 (WFI included), with development version-1 import.
