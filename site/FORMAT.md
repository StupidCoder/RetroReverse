# The Studio's common level format (format 1)

Every 2-D tilemap game (Sonic GG, Fort Apocalypse, Turrican, Marble Madness' map view,
Super Mario Land) exports the same JSON shapes into `site/public/<game>/`, consumed by
the shared viewer core in `site/src/shared/`. The format is a **superset**: a game uses
the sections its hardware/engine has, and omits the rest. All data is produced by each
game's Go exporter (`<Game>/extract/cmd/webexport`) straight from the reverse-engineered
formats — nothing is hand-authored.

Conventions: **all coordinates are world pixels** unless a field says otherwise; all
times are **engine frames** at the game's `tickHz` (50 = PAL C64/Amiga, 60 = GG/GB);
colours are `#rrggbb` strings; grids are row-major.

## meta.json

```jsonc
{
  "format": 1,
  "game": "sonic",
  "native": { "w": 160, "h": 144 },     // the machine's visible resolution
  "tickHz": 60,
  "wrap": "x",                          // optional: the map is a horizontal cylinder (Fort)
  "levels": [
    { "name": "Green Hills Act 1",
      "section": "Green Hills",         // optional accordion grouping in the Studio menu
      "file": "act01.json",
      "atlas": "atlas_0.png" }
  ]
}
```

## Level file

```jsonc
{
  "format": 1,
  "name": "Green Hills Act 1",

  // --- the tilemap -------------------------------------------------------------
  "grid": {
    "tileSize": 8,                      // px per tile (8 or 32)
    "atlas": "atlas_0.png",             // tile sheet PNG
    "atlasCols": 16,                    // tiles per atlas row
    "atlasGutter": 0,                   // 1 = each tile extruded by 1px (bleed guard)
    "width": 203, "height": 16,         // map size in CELLS
    "cells": [ ... ],                   // width*height cell values
    "hflipMask": 32768                  // optional: cell & mask = draw h-flipped (Turrican)
  },

  // Optional block indirection (Sonic): cells are indices into `tiles`, each block a
  // size x size arrangement of atlas tiles. Cell px size = tileSize * size.
  "blocks": { "size": 4, "tiles": [[16 tile ids], ...], "shapes": [shape id, ...] },

  // --- framing / player ----------------------------------------------------------
  "view":  { "x": 0, "y": 256, "w": 160, "h": 144 },  // initial camera framing rect
  "spawn": { "x": 160, "y": 368, "sprite": "0/00", "tint": "#b8c76f" },  // sprite/tint optional

  // --- objects --------------------------------------------------------------------
  "objects": [
    { "type": 8, "name": "crab", "x": 1344, "y": 401,   // engine REST position
      "sprite": "0/08", "tint": "#ffffff", "hard": true }  // sprite/tint/hard optional
  ],
  // Randomized placements (Fort): the viewer picks `count` candidates per pool each
  // time the objects layer is enabled (Math.random, or mulberry32(?seed=N)).
  "objectPools": [
    { "count": 8,
      "candidates": [[x, y], ...],        // [x, y, minDx, maxDx] when the pool patrols
      // Optional back-and-forth patrol (Fort's tanks/prisoners/mines): every
      // `stepFrames` engine frames one update fires; each update advances the
      // dirStamps phase, and every `updatesPerStep`-th update (default 1) moves the
      // placement `stepPx` in its current direction. A move that would leave the
      // candidate's inclusive [minDx, maxDx] span flips the direction (and the
      // facing) and moves the other way in the same update — mirroring the game
      // engines' probe-reverse-retry loops. minDx == maxDx == 0 stands still.
      // With `slide` the phase is the object's SUB-CELL POSITION instead: it
      // walks up with the direction and down again after a reversal, the stepPx
      // move commits when it wraps, and a span edge turns the walk around in
      // place (Fort's mines creep 2px per update this way; their dirStamps are
      // the same art at four sub-cell offsets). `startPhase` is the rest phase.
      // Spans are precomputed against the static map from the engine's own
      // turn-around rule; mover-vs-mover reversal (tank meets tank) isn't modelled.
      "patrol": { "stepPx": 8, "stepFrames": 8, "updatesPerStep": 1,
                  "slide": false, "startPhase": 0,
                  "start": "random" },    // initial direction: "random"|"right"|"left"
      "variants": [                       // one variant per placement, picked at random
        { "stamps": [{ "dx": 0, "dy": -8, "tile": 73 }, ...],   // atlas tiles stamped at offsets
          "sprite": "chopper-side", "tint": "#352879" },        // or/and a sprite
        // Patrolling art may be direction-dependent: one stamps array per animation
        // phase per facing (right.length == left.length; phases may differ in stamp
        // count). Facing swaps when the patrol reverses; anim-off shows right[0]
        // (or left[0] if the placement started leftward).
        { "dirStamps": { "right": [[{ "dx": 0, "dy": 0, "tile": 91 }, ...], ...],
                         "left":  [[...], ...] } }
      ] }
  ],

  // --- animation -------------------------------------------------------------------
  // Tile animation: any map cell showing tiles[i] cycles through frames[step][i].
  "tileAnims": [
    { "tiles": [252, 253, 254, 255], "frames": [[...4 ids], ...], "periodFrames": 10 }
  ],
  // Placement-anchored cell animators: a tw x th tile strip at (tx,ty) [8px tile
  // coords] repainted through phases with per-phase hold times. tw/th default to
  // 2x4 (Sonic's $50 objects); Marble's Ultimate screen-swap uses 36x30 (the
  // engine repaints its final screen from hidden variant rows stored past the
  // playable map).
  "cellAnims": [
    { "tx": 36, "ty": 44, "tw": 2, "th": 4,
      "phases": [{ "tiles": [tw*th ids], "frames": 240 }, ...] }
  ],

  // --- collision --------------------------------------------------------------------
  // kind "grid": per-TILE solidity, sub x sub cells each (sub 1 = whole tile).
  //   solid[tile*sub*sub + r*sub + c] = 0 (empty) or a class byte; `legend` maps class
  //   bytes to overlay colours (default red).
  // kind "profiles": per-column height profiles shared via shapesFile (Sonic):
  //   shapes.json = { count, profiles[count][32] (signed; -128 = none), angles[count] },
  //   selected per block by blocks.shapes[].
  "collision": { "kind": "grid", "sub": 4, "solid": [...], "legend": { "1": "#ff3030" } },

  // --- palette effects (Sonic only) --------------------------------------------------
  "paletteFx": {
    "palette": ["#...", ...],                          // the 16 BG colours as exported
    "cycle": { "steps": [["#..",..], ...], "periodFrames": 32, "tiles": [12, 13] },
    "waterLine": { "y": 416, "palette": ["#...", ...] } // raster split: below y use this
  },

  "music": "greenhills"                  // optional Studio track name
}
```

## sprites/index.json

One entry per sprite key. Sonic scopes keys by zone (`"0/08"`); other games use flat
names (`"chopper-fwd"`, `"tank"`, SML type ids).

```jsonc
{
  "0/08": {
    "src": "sprites/0/08.png",
    "frames": [[0, 0, 48, 48], [48, 0, 48, 48], ...],  // rects into src
    "anchor": [0, 0],                                   // draw at (x-ax, y-ay); default 0,0
    "durations": [13, 13, 13, 13],                      // per-frame hold (engine frames)
    "steps": [[0, 300], [1, 16], ...],                  // optional explicit play program
    "path": [[dx, dy], ...]                             // optional per-frame movement offsets
  }
}
```

- A sprite with one frame is static. With `durations` the frames play in order; with
  `steps` the explicit program plays (frame index + hold); both loop.
- `path` loops too: the object's draw position each engine frame is
  `(x, y) + path[t % len]` (moving platforms; sampled from the engines' own tables).

## What each game uses

| Section | Sonic | Fort | Turrican | Marble | SML |
|---|---|---|---|---|---|
| grid | ✓ (via blocks) | ✓ | ✓ + hflipMask | ✓ | ✓ |
| blocks | ✓ | – | – | – | – |
| tileAnims | ✓ | ✓ | – | ✓ (gold glitter) | – |
| cellAnims | ✓ | – | – | ✓ (36x30 swap) | – |
| objects | ✓ | – | ✓ | ✓ (overlays) | ✓ |
| objectPools | – | ✓ (patrol) | – | – | – |
| sprites index | ✓ (anims, paths) | ✓ (tinted) | ✓ | ✓ (flag anims) | ✓ (anchored) |
| collision | profiles | – | grid sub 4 | – | grid sub 1 |
| paletteFx | ✓ | – | – | – | – |
| wrap | – | ✓ | – | – | – |

Marble's 3-D "Slopes" view (`<key>.slope.json`: height field + patrol-route markers) is
outside this format — it stays a per-game three.js view, like Elite and Stunt Car Racer.
