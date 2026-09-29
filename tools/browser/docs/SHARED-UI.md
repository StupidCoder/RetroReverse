# Shared emulator UI

Design: retain RetroReverse's paper (#f4f6fa), artwork grey (#e9edf3), ink
(#283246), muted text (#607087), indigo (#4d578f), and scanline blue (#176f9e).
Avenir/system type, left-aligned controls and explanations, centered pixel images.
The hardware imagery is the focus; no visual rebranding is needed for this refactor.

All systems use the same Play / Render navigation. Play owns media, transport,
input, states and performance. Render owns the output buffer, auxiliary-panel
slots, inspection sidebar and one common timeline. Systems provide content and
interpretation, not copies of the page layout.

```
Play:    [game image / system-specific media options]
         [running screen            ] [states, controls, performance]
         [transport / input         ]
Render:  [view-specific controls     ]                 [Resume game]
         [auxiliary buffers] [output buffer] [pixel / hardware details]
         [-------------- shared scrub controls ------------]
```

The plan deliberately preserves the existing educational presentation, while
making the output, timeline and Play controls consistent across all sixteen
systems. Desktop panels have bounded internal scrolling; narrow screens stack.
Running and inspection canvases are separate, so scrubbing cannot replace the
last actual display shown in Play. Adapters retain their existing evidence limits.

## Ownership

- `ui-shell.js` mounts the one Play/Render page template. Individual route HTML
  files contain only the title, platform ID, fallback text and release-pinned
  entry point. Media, transport, states, gamepad buttons and performance stay in
  the shared application; `ui-platforms.js` describes dimensions, render adapter,
  firmware options and existing compatibility explanations.
- `render-workspace.js` owns workspace registration, capture lifetime, the
  separate output canvas, auxiliary/toolbar/sidebar slots and adapter selection.
  Adapters expose `reset`, `setCapture` and `result`; they do not register tabs.
- `buffer-view.js` creates every inspection buffer. It owns captions, image
  dimensions, pixel presentation and coordinate mapping. Normal display buffers
  honor the system's display aspect; one-bit blitter panels contain their image
  with letterboxing and use the corresponding pixel mapping.
- `render-timeline.js` owns First/Previous/Next/Last, the range control, markers,
  options and seek callbacks. It supports continuous command ranges and sparse
  sets of recorded scanlines or filtered blits.
- `replay.js` and `inspector.js` provide the command-based adapter used by twelve
  systems. `raster.js` supplies GB/GG/C64 layer and register details.
  `amiga-raster.js` supplies Copper/scanline and blitter lenses. The latter uses
  the same primary output panel for the post-blit playfield preview, with A/B/C/D
  and the pre-blit preview in its auxiliary slot.
- `style.css` owns shared layout and styling. System dimensions are CSS variables
  from metadata, not a list of repeated platform selectors. Only auxiliary
  content such as Amiga palette swatches and channel arrangements needs specific
  styling.

The command inspector still explains the **final captured pixel**, even when the
output is scrubbed to an earlier command; the timeline says so. The shared shell does not change those evidence limits. Source previews retain
existing frozen-state/scanline and blit-buffer limitations.

## Adding a view

Use `ui.auxiliary`, `ui.toolbar` and `ui.sidebar` for system-specific content.
Create auxiliary images with `ui.buffer(parent, options)` and paint the common
output with `ui.output.draw(pixels, width, height)`. Configure `ui.timeline.range`,
`value`, `setMarkers`, `toggle` and `onSeek` instead of creating another slider.
Select the adapter in `ui-platforms.js` and register its factory in
`render-workspace.js`. No per-emulator HTML, shared input code or worker timing
changes are needed for a new presentation.

Shared look-and-feel changes belong in `ui-shell.js`, `buffer-view.js`,
`render-timeline.js` or `style.css`; adapters should only explain their hardware.

## Verification

`node tools/browser/tests/ui-shell.mjs` covers every platform's required shared
controls, unique IDs, firmware inputs, compatibility toggles and DOS directory
selection. It runs in the normal public CI suite.

Serve the repository (not just `site/`) and open
`tools/browser/tests/ui-workspace.html` for media-free browser checks. Synthetic
captures exercise all sixteen adapters, command/scanline navigation, sparse
lines, Amiga masks and filters, cancelled/stale replies, corrected-aspect pixel
coordinates, resume/reset, and preservation of Play pixels while scrubbing.
The harness uses the actual DOM, canvas and production components.

Private game images and checkpoints are not included in the tests or the published site.

## Historical character and tile memory

`tileset.js` is a shared auxiliary view, with pure decoders in
`tileset-decode.js`. C64, GB and GG switch between their layer previews and atlas;
GBA exposes an atlas beside command replay. Atlas pixels keep their square-pixel
aspect independently of the output display. Palette and memory explanations live
in the shared inspector. Clicking a tile identifies its code and memory address.

The worker attaches a binary snapshot to the same completed seek response as the
output image. Existing request/capture guards reject stale atlas responses too.
C64/GB/GG exports copy the selected scanline's frozen graphics memory and palette;
GBA exports scratch replay memory, now including captured display-register writes.
No decoder reads live emulated memory. Palette selections only reinterpret the
captured bytes; they never write to the guest.

- C64: 256 codes from the current VIC-visible character address, including ROM
  overlay, hires/per-cell multicolor and extended-background mode. The default
  color comes from each code's first screen cell. An explicit cell selector is
  available because the same glyph can have different colors in different cells.
  Unused glyphs use the stated fallback cell. Bitmap/invalid modes are labelled as
  raw character memory rather than pretending this atlas generated the screen.
- Game Boy: all 384 tiles, BGP/OBP0/OBP1, signed background IDs and object-zero
  transparency. The existing core is DMG; this adds no CGB VRAM banks.
- Game Gear: all 512 four-plane tiles and both 16-color palettes. Palette zero is
  shown as its actual swatch; object transparency is explained separately.
- GBA: the selected background's current 16 KiB character block, current bit depth
  (including affine backgrounds), and selectable 4-bit palette bank. Object views
  cover the 32 KiB object region in either depth. Color zero is transparent; raw
  data in disabled/bitmap layers and the displayed block extent are labelled.
  The atlas is a memory arrangement, not a reconstructed tilemap or object layout.

These inspection exports do not alter save-state payload formats. Rebuilding the
four cores changes their binary identities, so the browser's existing strict
build check still rejects states made by a different release. No identity checks
are bypassed. Private fixtures are loaded and re-saved by the new core for QA.

`tileset.mjs` checks mode, palette and address decoding. The four corresponding
native raster/pixel tests check historical memory/bank changes, backward seeks,
capture invalidation and unchanged machine state. The browser harness exercises
atlas switching, tile address inspection and layout alongside all sixteen adapters.

## PlayStation VRAM and sampler state

`ps1-vram.js` plugs into the command adapter's auxiliary and inspector slots;
`ps1-vram-decode.js` decodes scratch VRAM after the same completed seek that
updates the output. Automatic mode follows the current packet's 4/8-bit indexed
or 16-bit direct texture setting and the last captured CLUT. Explicit mode
controls reinterpret the bytes without changing the guest. Unknown initial CLUTs
use labelled grayscale indices. Texture color zero is transparent only when its
resolved RGB555/STP word is zero; palette index zero itself may be opaque.

The auxiliary selector can enlarge either buffer on its own.
The overview keeps the 1024-word-wide physical VRAM layout (packed texels are
compressed horizontally); the 256×256 page preview has square texels. Yellow
marks the page and current primitive's UV outline, blue the CLUT. The page applies
the captured texture window by default. Clicking either image shows the packed
word, texel index, palette address and resolved color at the completed step.

The core's capture hook runs before command execution. `gpu-inspection.h`
projects that packet's page/depth/CLUT or environment setting into a capture-only
snapshot, fixing the previous command metadata's stale texture depth. Both
flat/Gouraud polygon layouts and rectangles are covered. Empty drawing commands
and setting commands retain their place in replay. No guest GPU fields, rendering
algorithms or portable state payloads change.

The inspector shows drawing area/offset, texture window, shading and raw-texture
requests, plus observed blend/dither/mask settings. Unsupported settings are
unknown until observed in this capture. The UI labels the existing core's ignored
blend/dither/mask effects and raw-polygon approximation. It also explains that
post-command VRAM can differ from a self-modifying primitive's earlier samples;
the existing pixel provenance retains the sampled values.

Register/packet reference: [psx-spx rendering attributes](https://psx-spx.consoledev.net/ps1/gpu/rendering-attributes/)
and [VRAM formats](https://psx-spx.consoledev.net/ps1/gpu/video-memory-vram/).
Tests: `vram-ps1.cpp`, `ps1-vram.mjs`, and the shared browser harness. A private
Ridge Racer WASM capture verified 93 pixel source/CLUT address chains across a
2,140-step frame history, with 13 texture pages and unchanged full machine state.
