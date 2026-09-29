# Game Boy raster inspection

Design: keep the existing RetroReverse palette (paper #f4f6fa, ink #283246,
muted #607087, accent #4d578f), with #176f9e reserved for the scanline indicator.
Use the existing Avenir/system family; numbers use tabular alignment. The drawing
itself supplies the visual identity; no new decorative chrome or fonts.

The Play workspace keeps media loading and game controls. Render is a separate,
paused workspace: two source previews on the left, accumulated output in the
center, contextual evidence on the right, one scanline timeline underneath.
At narrow widths the inspector moves below the pictures. Workspace registration
is separate from the renderer, allowing a future memory view without embedding
its navigation into game controls. Only implemented workspaces are shown.

Source previews mean state frozen at the selected scanline. The output preserves
actual previously rendered lines. Neither preview executes the CPU nor changes
PPU state. The Game Boy window's internal line counter must be captured, rather
than guessed from WY after mid-frame changes.

## Using the workspace

Load any supported DMG image in Play, run it, and press Pause. After the next
complete frame has been captured, Render opens automatically. Its three panels
show background/window, sprites before background priority, and the screen built
through the selected line. The blue guide can be hidden. First/last and line-step
buttons accompany the keyboard-operable range slider; small timeline markers
jump to lines where video memory or registers changed.

Play and Render share the same paused machine. Switching views neither resumes
execution nor changes the selected scanline. Resume game returns to Play and
invalidates the capture. Reset, load, save, and advancing the machine also discard
old evidence through the existing capture lifecycle. Workspace registration is
in `site/emulators/workspaces.js`; Game Boy presentation and worker messages are
in `raster.js`, now shared with C64. Other platforms retain their existing inspection UI.

The side panel shows video registers, the internal window row, and changes since
the preceding rendered line. Register writes include the real CPU writer PC.
Click any layer, or use its coordinate inputs, to inspect tile-row, tilemap,
palette, and sprite-attribute sources. Each source query uses the write-history
cutoff for that snapshot. Clicking an earlier completed output line uses that
line's historical state, not the currently selected line's settings.

## Capture and accuracy

The renderer has one shared, pure scanline kernel for normal execution and layer
previews. During capture only, each rendered line records VRAM, OAM, registers,
window row counter, trace position, and its final pixels. The additional fixed
snapshot storage is 1,316,736 bytes (1.26 MiB). CPU execution and state serialization
are unchanged; inspection uses scratch storage, never live machine memory.

The previews represent the selected line's settings held across the whole
screen. Their selected row is exact within this core's rendering model; other
rows are illustrative projections. The window preview anchors its internal row
counter to the captured value, including when the window was disabled earlier.
Completed output rows always retain their actual recorded pixels. Future or
uncaptured rows remain transparent, and querying them gives an explicit message.
A capture with the LCD off has no available rows rather than stale imagery.

This is a DMG, scanline-level visualization. It does not add dot-level LCD timing,
mid-scanline register effects, or Game Boy Color support. The sprite panel obeys
the ten-object limit and object-to-object priority, but shows the winning sprite
even when background priority hides it in the final screen. Source history is
bounded by the existing capture interval and trace budget; earlier writes are
identified as contents already present at capture start.

## Validation (2026-09-29)

- Public native regression executables and shared JavaScript checks passed; the
  final packaged release passes `check-release.py` for all sixteen routes.
- `tests/raster-gb.cpp` passes as native C++ and actual Emscripten WASM. It checks
  a mid-frame scroll split, sprite priority, tilemap/OAM changes, a stopped and
  resumed window counter, forward/backward seeks, unchanged serialized state,
  source addresses, final output equality, and an empty LCD capture.
- Private Super Mario Land boot still matches the native checkpoint. A scrolling
  gameplay frame passes 1,872 pixel reconstructions and 5,688 historical source
  checks over all 144 lines. The accumulated final output equals the captured
  framebuffer; complete serialized state is identical before and after seeks.
  See `results/gb-raster-super-mario-land.json` (no media or checkpoint included).
- That scene changes SCX to zero at line 0 from PC `0x0089`, then to 69 at line 16
  from PC `0x00A5`. The source background shifts while the completed status bar
  stays fixed. WASM reconstruction measured a median 0.118 ms per three-panel
  seek, maximum 0.672 ms in Node.js; these exclude worker transfer and canvas work
  and are not browser presentation-rate measurements.
- Browser testing covered local ROM/state loading, Pause capture, all three
  panels, change markers, keyboard scrubbing, pixel/source queries, workspace
  switching, resume, recapture, and saving. No browser errors were recorded.
  Render and its timeline fit without page scrolling at 1280×720 and 1440×900.
  At 390×844 the inspector flows below the panels without horizontal overflow.

Reproduce the media-free test with:

```sh
clang++ -O2 -std=c++20 -fwrapv -ffp-contract=off -Wno-parentheses-equality \
  tools/browser/tests/raster-gb.cpp -o /tmp/raster-gb -lz
/tmp/raster-gb
```

For a private image and raw core checkpoint:

```sh
node tools/browser/tests/validate-raster-gb.mjs /path/to/game.gb /path/to/raw.state
```
