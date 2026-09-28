# DS color and rendering replay corrections

The 3D rasterizer stores pixels as `AABBGGRR`, but the 2D compositor read them as
`RRGGBBAA`. This moved alpha into red, shifted the remaining color channels, and
used red as alpha. `threeDLine` now decodes the rasterizer's existing format
explicitly, preserving the internal surface/state format. Go and generated C++
have the fix. Tests exercise primary colors and alpha through the actual
raster-publish/compositor boundary.

Super Mario 64 DS was cold-booted with ordinary stylus input, then Adventure and
the first new save slot were selected. Peach's letter and the Lakitu/castle
sequence now show pink dress, blonde hair, blue sky and yellow Lakitu. Existing
renderer limitations such as fog and edge marking are separate from this bug.

DS capture records engine A, engine B and a separate 3D render target. Replaying
polygon writes changed that third surface while the UI kept showing completed
LCD buffers. Replay now follows the actual 3D surface during clear/polygon steps
and labels it **3D render target before 2D composition**. Compositor steps show
LCD outputs; Last always shows the exact captured display. The other LCD remains
visible. The preview does not execute the game again or replace final-frame pixel
evidence.

Capture spans two display intervals to include the delayed 3D producer before
2D composition. Source history can therefore include the recorded clear and
polygon writes instead of only pre-existing pixels. Capture-local surface
addresses remain labeled as such.

Validation covers both LCD routings, primary colors/transparency, unchanged LCD
buffers with changing 3D draws, backward seeks, exact final pixels, historical
sources, state continuation and unchanged live state. Browser scrubbing at steps
1, 86, 177 and 2,177 in the Lakitu scene showed clear, partial geometry and final
output. A selected yellow pixel reconstructed as `#f7ad00` and linked back to
matching historical 3D source bytes.

Reproduce the private-media boot sequence and replay checks:

```sh
node tools/platform/nds/browser/tests/intro.mjs /path/to/SM64DS.nds /private/tmp/ds-intro
node tools/platform/nds/browser/tests/render-replay.mjs /path/to/SM64DS.nds /private/tmp/ds-intro/1850.state
node tools/platform/nds/browser/tests/render-replay.mjs /path/to/SM64DS.nds /private/tmp/ds-intro/2600.state
```

The first command writes game-derived images/states outside the published site.
Only [measurement summaries](../results/ds-rendering-fixes.json) are committed.
Public portable states still require the matching core hash.

## Texture transparency and polygon order

The follow-up corrected a second, independent set of rasterizer problems:

- Fragment opacity now comes from the combined texture/polygon alpha. A3I5 and
  A5I3 textures on polygons with alpha 31 blend correctly, and only their opaque
  texels write depth unless POLYGON_ATTR explicitly enables translucent writes.
- Opaque polygons render first. Stable bottom-Y/top-Y ordering applies to opaque
  polygons and to translucent polygons in automatic mode; manual mode preserves
  translucent submission order. SWAP_BUFFERS latches the mode for the next list.
- Overlapping translucent fragments with the same polygon ID do not blend twice.
  Pixel inspection reports that rejection separately from depth and alpha tests.
- The DS integer modulation/decal/blend equations and nonzero RGB5-to-RGB6
  expansion replace normalized arithmetic. Transparent decal texels keep the
  vertex color rather than discarding the fragment.
- Byte/halfword writes reach rendering registers immediately; command FIFO ports
  still wait for a complete word.

These rules were checked against [GBATEK](https://mgba-emu.github.io/gbatek/#ds-3d-polygon-attributes)
and the ordering/fragment behavior in the
[melonDS renderer](https://github.com/melonDS-emu/melonDS/blob/master/src/GPU3D_Soft.cpp).
No reference emulator code is incorporated.

The private-media cold boot was repeated through Adventure and save slot A.
The title wallpaper has filled, shaded logos; clouds blend softly into the blue
sky, and translucent layers are depth-tested against the already-drawn portrait.
The new native/WASM tests cover opacity, depth writes, duplicate polygon IDs,
decals, stable automatic/manual ordering, register widths, and replay isolation.
Portable core-state format 3 adds the sort latch and translucent ID buffer; raw
formats 1/2 still load. Public state containers retain the core-hash compatibility
check. Fog, edge marking, shadow volumes and coverage antialiasing remain separate
renderer limitations.
