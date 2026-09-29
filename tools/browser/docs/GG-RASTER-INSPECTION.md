# Game Gear raster inspection

Design: extend the existing Play/Render workspace, with paper #f4f6fa, ink
#283246, muted #607087, accent #4d578f and scanline blue #176f9e. Keep the
Avenir/system type, left-aligned evidence and centered pixel panels. The 160×144
LCD uses the same proportions as Game Boy. Background and sprites sit on the
left, accumulated output in the center, VDP settings and source history on the
right. The timeline spans the image area; evidence flows below on narrow screens.
This deliberately reuses the accepted layout and makes Game Gear's distinct
scroll latches, palette words and split sprite table the content of the view.

Accuracy work must precede the UI: replace the instruction-budget scheduler with
Z80 T-state timing, model line/frame IRQ persistence and acknowledgement, latch
horizontal scroll per line and vertical scroll per frame, and commit Game Gear
12-bit palette pairs together. A shared pure scanline renderer must power both
normal execution and frozen-state previews without modifying paused state.

Hardware references:

- [Sega Game Gear hardware manual](https://www.smspower.org/Development/GGOfficialDocs)
- [Zilog Z80 CPU manual](https://www.zilog.com/docs/z80/um0080.pdf)

The view will describe scanline-level sampling explicitly; it does not imply a
dot-level VDP fetch pipeline or accurate effects from writes within a scanline.

## Using the workspace

Load any supported local cartridge in Play. Pause captures the next complete
display interval and opens Render. The left panels show the background and
winning sprites before background priority, with the selected line's settings
held fixed across the LCD. The center accumulates the actual recorded lines;
future rows remain transparent. Switching between Play and Render keeps the
paused machine and selected row. Resume returns to Play and discards old evidence.

The timeline covers LCD rows 0–143, corresponding to VDP lines 24–167 and columns
48–207. Markers identify video writes since the preceding snapshot. The first
row includes writes during the preceding blanking/off-screen interval. The side
panel shows register values, effective scroll, line-counter reload/current value,
pending line/frame interrupts, and table locations. It also counts palette and
VRAM writes; tilemap/sprite counts are subsets of VRAM writes.

Pixels lead to four-byte planar tile rows, two-byte tilemap entries, two-byte
palette colors, and the sprite table's separate Y and X/tile fields. Their source
queries use the selected snapshot's write-history cutoff. Earlier output rows
use their own snapshots, not the scrub position's current state. The UI labels
physical VRAM/CRAM offsets; the existing capture ABI retains separate address
spaces at `0x10000` (VRAM), `0x14000` (CRAM) and `0x14040` (VDP registers).

The renderer is shared between running emulation and inspection. Previews use
scratch machine data and do not set collision/overflow flags on the live VDP.
Each row captures VRAM, CRAM, registers, scroll latches, interrupt state and output
pixels. Extra snapshot storage is 2,466,432 bytes (2.35 MiB), allocated alongside
the existing approximately 4 MiB frame trace. Normal execution does not copy
these snapshots. The WASM heap still starts at 32 MiB.

## Accuracy changes

The browser host now schedules a frame using 228 Z80 T-states per line and 262
lines, at 3.579545 MHz (about 59.92 Hz). Instruction durations account for taken
branches, index/CB/ED prefixes, block repetition, interrupts and HALT. I/O advances
the VDP before the port operation; block instructions then finish their remaining
cycles. Opcode fetches update R correctly, consecutive EI instructions retain
their delay, and repeated/ignored index prefixes use the proper register bank.

Line and frame interrupts stay pending until a status read. Their enable bits
gate the CPU IRQ line without deleting the pending conditions. The V-counter
follows the NTSC discontinuity after `$DA`. R8 is latched at each F4 line boundary;
R9 is latched on the pre-display line. Game Gear CRAM buffers the even byte and
commits a masked 12-bit color only on the odd write, including odd-address writes
that reuse the buffered low byte. VDP port mirrors and 14-bit address wrapping
are modeled.

The translated Go routines remain available as reference implementations. The
generator explicitly routes Z80 stepping and Game Gear I/O to the maintained C++
implementations; regenerating does not erase the fixes. No additional emulator
library was imported. Go's original extraction/oracle scheduler is unchanged.

The raw Game Gear state format is now version 2, preserving cycle phase, pending
interrupts, scroll latches and the unfinished CRAM word. Earlier Game Gear states
are incompatible with this core and must be recreated. Loads remain transactional.

This is **scanline-level rendering with instruction-cycle scheduling**. VRAM,
palette and sprite evaluation are sampled once per line; mid-line fetches, pixel
output timing, VDP access waits and exact collision-flag timing are not modeled.
The external H-counter latch, alternate mappers, cartridge SRAM, link cable,
non-Mode-4 video and audio output remain outside this port's current scope.

## Validation (2026-09-29)

- The full public emulator suite and final release audit pass. The new
  `timing-gg.cpp` and `raster-gg.cpp` tests also pass in Emscripten WASM with
  assertions and stack checking at the production 1 MiB stack size.
- Timing fixtures exercise conditional branches, indexed and block operations,
  refresh, HALT/EI/IRQ handling, interrupt acknowledgement and enable changes,
  the V-counter discontinuity, scroll latches, paired palette writes, port
  mirrors, address wrapping and timing state restore.
- Raster fixtures include an actual Z80 program whose line IRQ handler writes
  R8 every sixteen lines. Captures identify its OUT at `$0046`. Other fixtures
  cover tilemap/palette/sprite changes, rejected sprite candidates, partial
  output, old source snapshots, empty captures and byte-identical saved state
  after arbitrary seeks.
- Private Sonic gameplay and Labyrinth captures each pass 1,872 pixel
  reconstructions and 6,192 historical-source checks over all 144 rows.
  Accumulated output equals the final framebuffer; inspection leaves complete
  serialized state unchanged. The Labyrinth fixture uses a private level/position
  setup, followed by the game's own execution; there are no Sonic conditions in
  the core. No cartridge or checkpoint is published.
- In the Labyrinth capture, palette changes occur on LCD rows 117–125. Pixel
  history identifies the underwater stores at `$01F7` and `$0201`, and the
  surface-palette loader at `$05EA`. Scrubbing changes the frozen background's
  colors while preserving previously completed rows in the center panel.
- See `results/gg-raster-sonic.json` and `results/gg-raster-sonic-labyrinth.json`.
  Median three-panel WASM seek time is approximately 0.3 ms in Node.js, excluding
  browser transfer and drawing. A separate 600-frame gameplay run took 340 ms
  (about 1,764 emulated frames/s); native and WASM end-state proofs match. This
  measures core execution, not browser presentation rate.
- Browser checks cover local cartridge/state loading, pause capture, scrubbing,
  source queries, workspace preservation, resume and recapture. The workspace
  fits 1280×720 and 1440×900 without page scrolling; its inspector can scroll
  internally. At 390×844 it flows vertically without horizontal overflow.

Reproduce the public tests with:

```sh
clang++ -O1 -std=c++20 -Wno-parentheses-equality \
  tools/browser/tests/timing-gg.cpp -o /tmp/timing-gg
/tmp/timing-gg
clang++ -O1 -std=c++20 -Wno-parentheses-equality \
  tools/browser/tests/raster-gg.cpp -o /tmp/raster-gg
/tmp/raster-gg
```

For private media and an optional raw checkpoint:

```sh
node tools/browser/tests/validate-raster-gg.mjs /path/to/game.gg /path/to/raw.state
```
