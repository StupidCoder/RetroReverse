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
