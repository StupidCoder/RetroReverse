# Amiga render workspace

Design plan: extend the accepted Play/Render layout. Paper #f4f6fa, ink #283246,
muted #607087, indigo #4d578f, scanline blue #176f9e, mask teal #287d77.
Use the existing Avenir/system typography. Keep explanations left aligned and
pixel images centered, with a bounded scrolling inspector on desktop.

Two lenses inside Render serve different parts of the hardware:

```
Scanlines:  [Playfield / selected plane] [Screen so far] [Copper, palette, sources]
            [Hardware sprites       ]
            [---------------- scanline timeline ------]
Blitter:    [A mask] [B image]          [Operation, shifts, masks, memory sources]
            [C input] [D result]       [Destination in displayed playfield]
            [---------------- operation timeline -----]
```

The visual distinction is the real Amiga data flow: BOBs are writes into planar
memory, not a hardware sprite layer. Channel panels show the actual one-bit
inputs after masking/shifting and the output of the selected Boolean operation.
A pixel can follow a playfield word to its producing blit. Copper events retain
the Copper-list address separately from the concurrent 68000 PC. A three-frame
capture gives buffered renderers more producer history; earlier history may
still be absent. No game-specific rendering or inferred BOB identification.

Accuracy scope: preserve horizontal palette changes instead of coloring an
entire line with its last palette, timestamp chipset work at its actual modeled
clock, and preserve the pending palette state in saves. Bitplane fetching and
blitter bus arbitration remain approximate; the UI must state this explicitly.

## Implemented workflow

Pause finishes the current interval and captures three PAL display intervals.
Scanlines & Copper shows the final display, one row at a time. The two source
panels reconstruct the selected line's memory and settings in scratch space;
they are illustrative previews, while completed output rows are the recorded
scanout. Select Combined or an individual bitplane. Hardware sprites remain a
separate layer, including attached pairs. Pixels on earlier output rows use
those rows' source snapshots, irrespective of the current slider position.

The timeline marks display-register writes. Row zero also includes setup during
vertical blanking. The sidebar shows the Copper instruction address (distinct
from the concurrent CPU PC), PAL line, horizontal position, palette, display
registers and bitplane pointers. A register or pixel-source button opens its
historical writes. Blitter-produced words offer an **Inspect blit** link.

Blitter & masks has an operation selector, filter and timeline. Its A/B/C panels
show the actual Boolean inputs after shifts and A's first/last-word masks; D
shows the computed result. A is teal, with bright pixels representing set bits.
The $CA operation explains mask selection as `D = (A & B) | (~A & C)`. Other
operations retain their own minterm and truth table; they are not classified as
BOBs merely because they write image data. Copies, clears, area fill, descending
operations and line operations are inspectable too. For line mode, displayed
rows are successive computed words, not a rectangular destination footprint.

Clicking any input/output bit exposes its fetched source words, previous words
that contribute shift carry, data-register constants for disabled DMA channels,
and destination before/after. History cutoffs are recorded per word, so overlap
and in-place blits do not read future source values. The operation's submission
PC identifies the instruction writing BLTSIZE; Copper-started operations retain
the Copper address instead.

Playfield before/after decodes memory on either side of the operation using the
final captured frame's per-line bitplane layout and palette. Changed regions can
be outlined. These are memory previews: a write to another buffer may produce no
change, and the images are not claimed to have been visible at the blit time.
Many BOBs require one operation per bitplane; the inspector deliberately exposes
those individual operations rather than guessing game-specific object groups.

## Core changes and accuracy

Palette writes now retain their modeled horizontal position. Previously the
scanline renderer used the last palette for the entire row. The chipset clock
also advances within each two-cycle tick, rather than stamping a whole CPU slice
with its ending clock. Copper evaluation no longer observes the previous line
with an out-of-range horizontal position at the line boundary. Sprite DMA's frame
wrap now processes line zero, not a fictitious line 312. Sprite data history uses
the DMA fetch cutoff, including each word of attached sprites.

The existing blitter implementation still executes memory effects atomically,
then models a busy interval. This view does **not** model incremental blitter DMA
bus slots. Bitplanes and non-palette display controls are still sampled per
scanline. CPU register writes are placed at the core's scheduling granularity;
Copper/display pipeline latency and bus contention are approximate. HAM component
pixels are identified, but their dependency chain through preceding pixels is
not expanded. ECS/AGA, interlace, exact collision timing and audio remain outside
this work. No additional emulator library was imported.

The raw state format is version 2, adding the initial line palette and pending
horizontal changes. Raw version-1 research checkpoints can be read and normalized;
the browser wrapper continues to enforce executable identity as before.
Inspection changes neither portable machine state nor execution timing counters,
and preview rendering is excluded from emulation performance buckets.

## Capture implementation

`inspection.h` holds bounded line snapshots, register events and blitter-word
records. It records data only during capture. Historical chip RAM is reconstructed
by applying or reversing captured writes; no 512 KiB RAM copy is stored for every
scanline. Scratch previews call the same scanline renderer as normal execution.
`inspection-api.h` supplies scanline, pixel, blit and source metadata to the worker.
The UI lives in `site/emulators/amiga-raster.js`, alongside the shared workspace
navigation and the existing local-media/state infrastructure.

Additional blitter words are capped at 16 MiB; the operation and register-event
lists are capped at 4,096 and 32,768 respectively. Existing trace limits also
apply. Truncation is reported as incomplete evidence. Source history cannot infer
writers before the captured intervals. The redundant full-frame clear writes
were removed: every output row already overwrites the display, so the three-frame
capture remains below the shared write cap in both reference scenes.

## Validation (2026-09-29)

- The full public suite for all sixteen browser cores passes.
- `raster-amiga.cpp` passes natively and in WASM with assertions and stack checking
  at the production 8 MiB stack size. It covers horizontal palette boundaries,
  a real Copper WAIT/MOVE, exact event clocks, shifted/edge-masked cookie cuts,
  historical source values, separate hardware-sprite previews, line-mode register
  inputs, before/after previews, unchanged paused state and restoration partway
  through a palette-changing scanline. Existing Amiga tests
  retain all 256 minterms, descending/shifted copies, CIA/disk/CPU and replay tests.
- Turrican and Marble Madness each match 120 frames of native/WASM continuation
  byte-for-byte. Both pass scanline reconstruction, historical sources, complete
  output/replay equality and unchanged state after seeking. Private checkpoints
  and disk images are not published.
- The Turrican fixture contains 117 blits, including 84 cookie cuts. Its test
  checks 144 screen pixels, more than 1,000 historical sources and 90 blitter
  words, including links from plane memory to producing blits. Another fresh
  boot passes the original loader, intro and title and reaches gameplay.
- Browser checks cover Play/Render switching, scanlines, individual planes,
  Copper history, blit selection and links from destination memory to its blit.
  Both lenses fit 1440×900 and 1280×720 desktop viewports; the mobile layout
  stacks without horizontal overflow at 390×844. No browser errors were logged.
- Core WASM execution remains roughly 250–285 emulated frames/s on this Mac.
  Three-frame capture takes roughly 60 ms. Median scanline and blit seeks are
  about 1.6 ms and 3 ms, excluding browser transfer/drawing. Captured evidence is
  about 41–43 MiB, plus roughly 1 MiB of inspection records. These are local
  measurements, not a claim about all browsers or hardware.

Run public Amiga tests with:

```sh
python3 tools/platform/amiga/browser/build.py --native-only --test
```

Private-media validation (optional native end state):

```sh
node tools/browser/tests/validate-raster-amiga.mjs game.adf checkpoint.state native-end.state
```

The hardware interpretation follows Commodore's
[Blitter Hardware chapter](https://www.amigarealm.com/computing/knowledge/hardref/ch6.htm)
and the existing port's
[Amiga Hardware Reference Manual](https://www.ikod.se/wp-content/uploads/2020/08/Amiga_Hardware_Reference_Manual_3rd_Edition.pdf).
Turrican-specific examples were checked against the repository's
`games/turrican-amiga/turrican-amiga.md`; no game recognition is in the core.
