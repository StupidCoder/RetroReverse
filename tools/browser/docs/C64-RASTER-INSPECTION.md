# C64 raster inspection

Design: extend the accepted Play/Render workspace pattern. Keep paper #f4f6fa,
ink #283246, muted #607087, accent #4d578f, and scanline blue #176f9e, with the
existing Avenir/system typography. Align labels and evidence left; center the
three image panels. Use C64's wider display proportions rather than Game Boy's.

The left column shows graphics/border and sprites; the center accumulates the
actual output; the right explains the selected line's VIC settings and source
writes. A scanline timeline spans the image area. On narrow screens evidence
flows below. Reuse workspace navigation and the raster UI, with platform-specific
evidence formatting. Keep tape and firmware controls in Play.

The source previews hold the line's memory and registers fixed. A separate VIC
instance renders those previews using the existing chip implementation, without
executing guest CPU code or touching the paused machine. The accumulated screen
and its pixel evidence retain the original cycle-level fetches and priority
decisions, including register changes partway through a line. Clearly distinguish
frozen-state previews from actual recorded output.

Fort Apocalypse is validation material, not a condition in the renderer. Its HUD
character definitions at $5000, playfield definitions at $5800, and rewritten
scanner characters at $52E0–$53FF should be discoverable from ordinary captured
addresses and writer PCs. No game-specific overlay or hardcoded detection.

## Using the workspace

Load a supported local C64 image, run it, and press Pause. Once the next complete
display interval is captured, Render opens with three panels: graphics/border,
sprites before background priority, and the actual screen drawn through the
selected line. Tape, firmware, input, state and execution controls remain in Play.
Switching workspaces preserves the paused machine and scanline selection.
Resuming execution discards the old capture and returns to Play.

The slider covers all 272 visible rows, including borders. The side panel also
shows the physical VIC raster number, active screen and character/bitmap bases,
video mode, and register values. Timeline markers jump to value-changing video
writes; repeated writes are still listed in the side panel. Select a write to see
its actual CPU PC, cycle, old/new values, and code bytes.

Click a pixel in any panel, or enter coordinates, to inspect its screen-memory
byte, character/bitmap byte, color RAM, sprite data and control registers.
Selecting a source reveals the last recorded CPU writer. Clicking an earlier
completed output row uses that row's original fetch evidence, even after
scrubbing past a character-set switch. Future rows are transparent and cannot
be queried as if already drawn.

The shared UI and worker protocol live in `site/emulators/raster.js` and
`worker.js`; workspace registration remains in `workspaces.js`. This extends the
Game Boy layout without adding game-specific conditions or changing the other
emulators' inspection layouts.

## Capture and accuracy

During capture only, each visible line snapshots RAM, color RAM, VIC registers,
bank selection, cycle and write-journal cutoff at its first visible pixel.
Captured fetches identify RAM used for screen, graphics and sprite data, allowing
the timeline to find CPU changes to those sources. The CPU-write journal is
bounded to 65,536 entries; overflow is reported as incomplete evidence.

Layer previews use a separate instance of the existing VIC implementation. It
runs through two PAL frames with the selected snapshot's memory and registers
held fixed, without running the CPU. These are projections of the configuration,
not historical replays of other scanlines. They do not reconstruct hidden VIC
state accumulated by earlier mid-frame changes, so even their selected row may
differ from actual output in raster-effect scenes. The accumulated screen and
its provenance instead retain the original cycle-level pixels, fetches, bank
selection and priority decisions, including changes within a scanline.

The sprite panel obeys sprite-to-sprite priority but exposes the winning sprite
before graphics/border priority. Preview queries resolve writers against the
snapshot's history cutoff. Neither seeking nor querying mutates the live board,
paused state, or original capture. Character ROM sources have no CPU writer.

Fort Apocalypse's test frame adds 21,678,768 bytes (20.7 MiB) of snapshot and
writer evidence to the existing 10,371,672-byte pixel capture. This work happens
only when capturing or inspecting; the normal running path does not copy RAM
every line. The production WASM initial memory remains 128 MiB.

## Fort Apocalypse findings

The scanner/minimap is a small software bitmap presented through redefined
characters. Its 36 displayed characters occupy `$52E0–$53FF`, within the HUD
character set at `$5000`. The copy routine starting at `$ADD3` writes their rows;
the captured pixel history identifies the actual store at **`$ADDD`**.

During the frame, the game changes D018 from `$14` to `$16`, switching character
definitions from `$5000` to `$5800` for the playfield while screen RAM stays at
`$4400`. In the validation capture, the new setting is visible at output line
105 (VIC raster 120), written by PC `$AE35`. Scrubbing across that change replaces
the projected graphics layer, while the already drawn minimap remains intact in
the accumulated output. These addresses come from the generic capture; the UI
does not recognize Fort Apocalypse by name or hash.

## Validation (2026-09-29)

- The full public native regression suite passed, including both raster
  inspectors and the other fourteen emulator routes. The final pinned release
  passes the artifact and static-network audit.
- `tests/raster-c64.cpp` passes as native C++ and Emscripten WASM with assertions
  and stack-overflow checks at the production 4 MiB stack size. It exercises text
  and bitmap modes, character ROM overlays, sprite priority, actual 6502 glyph
  writes and D018 switching, historical source queries, transparent future rows,
  empty captures, and unchanged serialized machine state.
- A private Fort Apocalypse gameplay checkpoint passes 2,720 actual-pixel and
  3,088 RAM-source checks across all 272 rows. Forty-eight sampled minimap
  graphics references resolve to the store at `$ADDD`. The accumulated last
  frame matches actual output; serialized state is unchanged by inspection.
  See `results/c64-raster-fort-apocalypse.json`; no game image or state is included.
- Three-panel WASM seeks in that scene measured a median 2.739 ms and maximum
  4.640 ms in Node.js. These exclude worker transfer and canvas presentation.
- Browser checks cover hosted firmware, local tape/state loading, pause capture,
  charset-switch scrubbing, minimap source/writer selection, workspace switching,
  resume and recapture. The layout fits 1280×720 without page scrolling; at
  390×844 it flows vertically without horizontal overflow. The shared Game Boy
  inspector also passes a browser regression with Super Mario Land.

Reproduce the media-free test with:

```sh
clang++ -O1 -std=c++20 -Wno-address-of-temporary \
  tools/browser/tests/raster-c64.cpp -o /tmp/raster-c64
/tmp/raster-c64
```

For a private tape image and raw core checkpoint:

```sh
node tools/browser/tests/validate-raster-c64.mjs /path/to/game.tap /path/to/raw.state
```
