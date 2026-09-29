# Game Boy and Game Gear browser ports

Both handhelds now use the existing static emulator UI: local cartridge selection,
Run/Pause/Reset, keyboard/on-screen/standard-gamepad controls, performance counters,
portable `.rrstate` files, automatic frame capture and rendering replay. No firmware
or server process is required. There is no game allowlist.

Both handhelds now have separate Play/Render workspaces. See
[Game Boy raster inspection](GB-RASTER-INSPECTION.md) and
[Game Gear raster inspection](GG-RASTER-INSPECTION.md). The latter also replaces
the original Game Gear instruction-budget timing described by the initial port.

The bounded source generator in `tools/browser/handheld-portgen/` translates the
existing Go SM83 / Z80 CPU and cartridge/device models. It selects the execution
sources and fails on unsupported syntax. The C++ hosts add stable completed LCD
frames and explicitly serialize CPU/device/mapper/timing state. ROM data stays in
the selected local file and is identified by the shared state container's hash.
State loads validate into a separate machine and only replace the running machine
after validation. Host pointers are rebound, never serialized.

## Rendering and inspection

The Game Boy renderer samples VRAM, OAM and registers for each scanline. It supports
signed/unsigned background tile addressing, scrolling, the window's own line counter,
8×8 / 8×16 objects, flips, the ten-object line limit and DMG X/OAM priority. Object
priority is resolved before the winning object's background-priority flag. The LCD
mode and STAT interrupt model now supports scanline changes such as Super Mario
Land's status bar. Boot uses the DMG post-boot register state.

The Game Gear renderer implements Mode 4 background tiles, palette selection,
flips, scrolling locks, sprite size/zoom, eight sprites per line, object/background
priority, collision/overflow flags and the central 160×144 LCD view of the 256×192
VDP image. Sega memory mapping and the VDP read buffer/control latch are retained.

Pause records the next whole display interval. Rendering steps represent hardware
scanline/layer operations, rather than pretending that these machines submit GPU
commands. Each pixel records its sampled tile row, palette bytes, coordinates,
transparency/priority rejection and the CPU PC at rendering time. Following a tile
or palette shows the actual CPU writes within the capture; unchanged data is
explicitly described as predating the capture. The scanline PC is not claimed to be
the instruction that produced the graphic.

A modeled LCD framebuffer occupies capture address `0x20000`. Game Boy source
addresses retain the DMG VRAM/OAM/register addresses; cartridge RAM is stored at
`0x10000`. Game Gear capture addresses are `0x10000` for VRAM, `0x14000` for CRAM,
`0x14040` for VDP registers and `0xC000` for work RAM. These capture addresses are
explained in each page's Compatibility panel. Replay rebuilds the display from
recorded effects in scratch memory and does not run or mutate the guest machine.

## Validation and performance

Super Mario Land reaches playable World 1-1; Sonic the Hedgehog reaches playable
Green Hill. Native and WASM runs agree on instruction count, PC, captured memory
and RGBA hashes through these boot/input sequences. Public fixtures check CPU
execution, mapper behavior, window and object priority, sprite limits, VDP writes,
source history, reconstruction, seek reversibility and transactional state restore.
The private WASM harness also verifies historical tile/palette bytes and that
replay leaves state and continuation untouched. No commercial ROM or checkpoint is
committed or deployed.

| WASM workload | Display updates | Core time | Updates/s |
|---|---:|---:|---:|
| Super Mario Land boot through gameplay | 600 | 0.303 s | 1,977 |
| Sonic boot through gameplay | 1,200 | 1.363 s | 880 |

These Node/V8 measurements exclude canvas delivery and browser pacing. They show
ample headroom, not an assertion that browsers present thousands of frames per
second. Both handheld browser pages were observed at approximately 60 presented
frames/s at normal speed. Each core starts with a 32 MiB WASM heap. Tested frame
captures contain about 4 MiB of evidence, and raw states are approximately 229 KiB
(Game Boy) / 208 KiB (Game Gear). Exact proofs and timings are recorded in
`tools/browser/results/handheld-ports.json`.

## Compatibility limits

- Game Boy: original monochrome DMG, ROM-only / MBC1 cartridges of 32 KiB–2 MiB.
  MBC1 banking mode 1 includes the banked low ROM window. Color-only cartridges
  and other mappers are rejected explicitly. Pixel FIFO timing, DMA duration,
  VRAM/OAM access restrictions and timer-edge quirks are not cycle accurate.
- Game Gear: Sega mapper, 16 KiB–4 MiB, optional 512-byte copier header. Z80
  instruction-cycle scheduling, persistent line/frame interrupts, scroll latches
  and paired CRAM writes are modeled. Rendering remains scanline-level; VDP fetch
  timing within a line, the external H-counter latch, alternate mappers,
  cartridge SRAM and link cable are not modeled.
- Neither browser port connects audio output. These are educational ports of the
  project's existing experimental cores, not replacements for mature emulators.

Rendering was checked against the hardware notes in
[Pan Docs rendering](https://raw.githubusercontent.com/gbdev/pandocs/master/src/Rendering.md),
[Pan Docs OAM](https://raw.githubusercontent.com/gbdev/pandocs/master/src/OAM.md),
[SMS Power VDP registers](https://www.smspower.org/Development/VDPRegisters) and
[SMS Power sprites](https://www.smspower.org/Development/Sprites).

## Rebuild

```sh
go run ./tools/browser/handheld-portgen gameboy
go run ./tools/browser/handheld-portgen gamegear
python3 tools/browser/state/generate.py
python3 tools/platform/gameboy/browser/build.py --emcc /path/to/em++
python3 tools/platform/gamegear/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
node tools/browser/tests/validate-handheld.mjs gb /path/to/game.gb /path/to/native.state
node tools/browser/tests/validate-handheld.mjs gg /path/to/game.gg /path/to/native.state
```

The native harness accepts `HH_START` (frame to press Start for eight frames),
`HH_LOAD` and `HH_SAVE`. Raw native states can be supplied to the WASM validation
harness. Browser `.rrstate` files wrap these bytes with media/core identities and
the pending input state. Generated console artwork and prompts are in
`site/emulators/art/` and `tools/browser/docs/ARTWORK.json`.
